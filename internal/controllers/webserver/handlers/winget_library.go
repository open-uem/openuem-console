package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/go-co-op/gocron/v2"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	openuem_nats "github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/views/deploy_views"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/openuem-console/internal/wingetrepo"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	// WinGetSourceName is the name of the console's repository in winget on the agents
	WinGetSourceName = "OpenUEM"

	// The repository is registered on the agents again every 30 minutes, so agents that
	// were admitted automatically, reinstalled or offline get it too. Agents only run
	// winget when the source changed.
	winGetSourceRegistrationInterval = 30 * time.Minute
	winGetSourceRegistrationTimeout  = 10 * time.Second
)

var identifierPartPattern = regexp.MustCompile(`[^0-9A-Za-z]+`)

// DeployLibrary shows the packages of the private winget repository
func (h *Handler) DeployLibrary(c echo.Context) error {
	return h.renderLibrary(c, "")
}

func (h *Handler) renderLibrary(c echo.Context, successMessage string) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	if h.WinGetLibrary == nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "library.not_available"), false))
	}

	packages, err := h.libraryPackages(commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return RenderView(c, deploy_views.DeployIndex("| Deploy", deploy_views.Library(c, packages, successMessage, commonInfo), commonInfo))
}

// libraryPackages returns the packages of the current organization and those available to
// every organization
func (h *Handler) libraryPackages(commonInfo *partials.CommonInfo) ([]wingetrepo.Package, error) {
	all, err := h.WinGetLibrary.List()
	if err != nil {
		return nil, err
	}

	tenantID, _ := strconv.Atoi(commonInfo.TenantID)
	packages := []wingetrepo.Package{}
	for _, p := range all {
		if tenantID == -1 || p.VisibleTo([]int{tenantID}) {
			packages = append(packages, p)
		}
	}
	return packages, nil
}

// hiddenLibraryPackageIDs returns the identifiers of the library packages the current
// organization can't install. The package catalog is shared by all organizations, so
// the package searches leave these out. Nothing is hidden when all organizations are
// selected (-1).
func hiddenLibraryPackageIDs(all []wingetrepo.Package, tenantID int) []string {
	if tenantID == -1 {
		return nil
	}
	visible := map[string]bool{}
	for _, p := range all {
		if p.VisibleTo([]int{tenantID}) {
			visible[p.PackageIdentifier] = true
		}
	}
	hidden := []string{}
	seen := map[string]bool{}
	for _, p := range all {
		if !visible[p.PackageIdentifier] && !seen[p.PackageIdentifier] {
			seen[p.PackageIdentifier] = true
			hidden = append(hidden, p.PackageIdentifier)
		}
	}
	return hidden
}

// hiddenLibraryPackageIDsForProfile returns the identifiers of the library packages a
// profile's task can't install: those of other organizations. A global profile, or one of
// several organizations, applies to computers of organizations that can't install an
// organization's package, so it only gets the packages available to all of them.
func hiddenLibraryPackageIDsForProfile(all []wingetrepo.Package, profileTenantIDs []int) []string {
	if len(profileTenantIDs) == 1 {
		return hiddenLibraryPackageIDs(all, profileTenantIDs[0])
	}
	return hiddenLibraryPackageIDs(all, 0)
}

// hiddenLibraryPackagesForProfile returns the library packages a profile's tasks can't
// install, or those hidden from the current organization if the profile isn't known
func (h *Handler) hiddenLibraryPackagesForProfile(profileID string, commonInfo *partials.CommonInfo) []string {
	id, err := strconv.Atoi(profileID)
	if err != nil || h.WinGetLibrary == nil {
		return h.hiddenLibraryPackages(commonInfo)
	}
	tenantIDs, err := h.Model.ProfileTenantIDs(id)
	if err != nil {
		return h.hiddenLibraryPackages(commonInfo)
	}
	all, err := h.WinGetLibrary.List()
	if err != nil {
		log.Printf("[ERROR]: could not list the packages of the winget repository, reason: %v", err)
		return nil
	}
	return hiddenLibraryPackageIDsForProfile(all, tenantIDs)
}

// hiddenLibraryPackages returns the library packages hidden from the current organization
func (h *Handler) hiddenLibraryPackages(commonInfo *partials.CommonInfo) []string {
	if h.WinGetLibrary == nil {
		return nil
	}
	all, err := h.WinGetLibrary.List()
	if err != nil {
		log.Printf("[ERROR]: could not list the packages of the winget repository, reason: %v", err)
		return nil
	}
	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return nil
	}
	return hiddenLibraryPackageIDs(all, tenantID)
}

// DeployLibraryUpload adds an MSI package to the repository. Empty fields are filled with
// the properties of the MSI package.
func (h *Handler) DeployLibraryUpload(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()

	if h.WinGetLibrary == nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(ctx, "library.not_available"), false))
	}

	header, err := c.FormFile("library-file")
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(ctx, "library.no_file"), false))
	}
	extension := strings.ToLower(filepath.Ext(header.Filename))
	if extension != ".msi" && extension != ".exe" && extension != ".msix" {
		return RenderError(c, partials.ErrorMessage(i18n.T(ctx, "library.not_supported"), false))
	}

	src, err := header.Open()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	defer src.Close()

	staged, err := h.WinGetLibrary.Stage(src)
	if err != nil {
		log.Printf("[ERROR]: could not save the uploaded package, reason: %v", err)
		return RenderError(c, partials.ErrorMessage(i18n.T(ctx, "library.upload_failed"), false))
	}
	defer os.Remove(staged)

	detected, err := inspectInstaller(staged, extension)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(ctx, "library.not_"+strings.TrimPrefix(extension, ".")), false))
	}

	// Fields filled in the form replace what was read from the installer
	p := wingetrepo.Package{
		PackageName:       firstNotEmpty(c.FormValue("library-name"), detected.Name),
		Publisher:         firstNotEmpty(c.FormValue("library-publisher"), detected.Publisher),
		PackageVersion:    firstNotEmpty(c.FormValue("library-version"), detected.Version),
		ProductCode:       firstNotEmpty(c.FormValue("library-product-code"), detected.ProductCode),
		InstallerType:     detected.InstallerType,
		PackageFamilyName: detected.PackageFamilyName,
		Architecture:      firstNotEmpty(c.FormValue("library-architecture"), detected.Architecture, "neutral"),
		InstallerFile:     wingetrepo.SanitizeInstallerName(header.Filename),
		PackageIdentifier: strings.TrimSpace(c.FormValue("library-id")),
	}
	if extension == ".exe" {
		if t := c.FormValue("library-installer-type"); slices.Contains(wingetrepo.ExeInstallerTypes, t) {
			p.InstallerType = t
		}
		p.SilentSwitches = strings.TrimSpace(c.FormValue("library-silent-switches"))
		if p.InstallerType == wingetrepo.InstallerTypeEXE && p.SilentSwitches == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(ctx, "library.exe_needs_switches"), false))
		}
	}
	if p.PackageIdentifier == "" {
		p.PackageIdentifier = defaultPackageIdentifier(p.Publisher, p.PackageName, p.PackageVersion)
	}
	if !strings.HasSuffix(strings.ToLower(p.InstallerFile), extension) {
		p.InstallerFile += extension
	}
	if c.FormValue("library-scope") != "all" {
		p.TenantID, _ = strconv.Atoi(commonInfo.TenantID)
		if p.TenantID < 0 {
			p.TenantID = 0
		}
	}

	added, err := h.WinGetLibrary.Add(p, staged)
	if errors.Is(err, wingetrepo.ErrInvalidPackage) {
		return RenderError(c, partials.ErrorMessage(i18n.T(ctx, "library.invalid_package", p.PackageIdentifier, p.PackageVersion), false))
	}
	if err != nil {
		log.Printf("[ERROR]: could not add the package %s to the winget repository, reason: %v", p.PackageIdentifier, err)
		return RenderError(c, partials.ErrorMessage(i18n.T(ctx, "library.upload_failed"), false))
	}

	if err := h.Model.SetLibraryPackageInCatalog(added.PackageIdentifier, added.PackageName); err != nil {
		log.Printf("[ERROR]: could not add the package %s to the package catalog, reason: %v", added.PackageIdentifier, err)
	}

	return h.renderLibrary(c, i18n.T(ctx, "library.uploaded", added.PackageName, added.PackageVersion, added.PackageIdentifier))
}

// DeployLibraryDelete removes a package version from the repository
func (h *Handler) DeployLibraryDelete(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()

	if h.WinGetLibrary == nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(ctx, "library.not_available"), false))
	}

	packageID := c.FormValue("library-id")
	version := c.FormValue("library-version")

	// Only packages shown in the current organization can be removed
	packages, err := h.libraryPackages(commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	found := false
	for _, p := range packages {
		if p.PackageIdentifier == packageID && p.PackageVersion == version {
			found = true
		}
	}
	if !found {
		return RenderError(c, partials.ErrorMessage(i18n.T(ctx, "library.not_found"), false))
	}

	if err := h.WinGetLibrary.Remove(packageID, version); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if versions, err := h.WinGetLibrary.Versions(packageID); err == nil && len(versions) == 0 {
		if err := h.Model.RemoveLibraryPackageFromCatalog(packageID); err != nil {
			log.Printf("[ERROR]: could not remove the package %s from the package catalog, reason: %v", packageID, err)
		}
	}

	return h.renderLibrary(c, i18n.T(ctx, "library.removed", packageID, version))
}

// DeployLibraryRegister registers the repository on the computers of the current
// organization right away, instead of waiting for the periodic registration
func (h *Handler) DeployLibraryRegister(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return RenderError(c, partials.ErrorMessage(i18n.T(ctx, "nats.not_connected"), false))
	}

	tenantID, _ := strconv.Atoi(commonInfo.TenantID)
	agentIDs, err := h.Model.EnabledWindowsAgentIDs(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	reached := len(agentIDs) - len(h.registerWinGetSources(agentIDs))
	return h.renderLibrary(c, i18n.T(ctx, "library.registered", reached, len(agentIDs)))
}

// winGetSourceConfig is the winget source an agent must register. Agents trust OpenUEM's
// CA unless the console is behind a reverse proxy, which usually has a public certificate.
func (h *Handler) winGetSourceConfig(agentID string) openuem_nats.WinGetSourceConfig {
	return openuem_nats.WinGetSourceConfig{
		Name:    WinGetSourceName,
		URL:     wingetrepo.SourceURL(h.consoleBaseURL(), []byte(h.JWTKey), agentID),
		TrustCA: h.ReverseProxyServer == "",
	}
}

// wingetPackageSource is the winget source a profile task installs a package from: the
// console's repository for its own packages, winget's community repository otherwise
func (h *Handler) wingetPackageSource(packageID string) string {
	isLibraryPackage, err := h.Model.IsLibraryPackage(packageID)
	if err != nil {
		log.Printf("[ERROR]: could not check if %s is a package of the winget repository, reason: %v", packageID, err)
	}
	if isLibraryPackage {
		return WinGetSourceName
	}
	return "winget"
}

// consoleBaseURL is the address of the console's web server as the computers reach it.
// The winget repository is served there because winget can't connect to the auth
// server, which asks for a client certificate.
func (h *Handler) consoleBaseURL() string {
	if h.ReverseProxyServer != "" {
		return "https://" + h.ReverseProxyServer
	}
	return fmt.Sprintf("https://%s:%s", h.ServerName, h.ConsolePort)
}

// registerWinGetSources asks the agents, concurrently, to register the repository. It
// returns the agents that couldn't be asked, e.g. because they're offline.
func (h *Handler) registerWinGetSources(agentIDs []string) map[string]error {
	var mu sync.Mutex
	var wg sync.WaitGroup
	failures := map[string]error{}
	limit := make(chan struct{}, updatePackagesConcurrency)

	for _, agentID := range agentIDs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			limit <- struct{}{}
			defer func() { <-limit }()

			if err := h.registerWinGetSource(agentID); err != nil {
				mu.Lock()
				failures[agentID] = err
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	return failures
}

func (h *Handler) registerWinGetSource(agentID string) error {
	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() || h.JWTKey == "" {
		return errors.New("the winget repository can't be registered")
	}

	data, err := json.Marshal(h.winGetSourceConfig(agentID))
	if err != nil {
		return err
	}

	msg, err := h.NATSConnection.Request("agent.wingetsource."+agentID, data, winGetSourceRegistrationTimeout)
	if err != nil {
		return err
	}
	if len(msg.Data) > 0 {
		return errors.New(string(msg.Data))
	}
	return nil
}

// StartWinGetRepositoryJob keeps the package catalog in sync with the repository and
// registers the repository on every enabled Windows agent every 30 minutes
func (h *Handler) StartWinGetRepositoryJob() error {
	if h.WinGetLibrary == nil {
		return nil
	}

	if err := h.syncLibraryCatalog(); err != nil {
		log.Printf("[ERROR]: could not add the winget repository packages to the package catalog, reason: %v", err)
	}

	_, err := h.TaskScheduler.NewJob(
		gocron.DurationJob(winGetSourceRegistrationInterval),
		gocron.NewTask(func() {
			agentIDs, err := h.Model.EnabledWindowsAgentIDs(-1)
			if err != nil {
				log.Printf("[ERROR]: could not get the agents to register the winget repository, reason: %v", err)
				return
			}
			h.registerWinGetSources(agentIDs)
		}),
	)
	return err
}

// syncLibraryCatalog makes the package search find the repository's packages, e.g. after
// the database was restored
func (h *Handler) syncLibraryCatalog() error {
	packages, err := h.WinGetLibrary.List()
	if err != nil {
		return err
	}
	inCatalog, err := h.Model.GetLibraryPackagesInCatalog()
	if err != nil {
		return err
	}

	names := map[string]string{}
	for _, p := range packages {
		if _, ok := names[p.PackageIdentifier]; !ok {
			names[p.PackageIdentifier] = p.PackageName
		}
	}
	for _, id := range inCatalog {
		if _, ok := names[id]; !ok {
			if err := h.Model.RemoveLibraryPackageFromCatalog(id); err != nil {
				return err
			}
		}
	}
	for id, name := range names {
		if err := h.Model.SetLibraryPackageInCatalog(id, name); err != nil {
			return err
		}
	}
	return nil
}

// installerDetails is what was read from an uploaded installer
type installerDetails struct {
	InstallerType     string
	Name              string
	Publisher         string
	Version           string
	ProductCode       string
	PackageFamilyName string
	Architecture      string
}

// inspectInstaller reads the name, publisher and version of an installer: the properties
// of an MSI package, the version resource of an .exe, the manifest of an MSIX package
func inspectInstaller(path, extension string) (installerDetails, error) {
	switch extension {
	case ".msi":
		f, err := os.Open(path)
		if err != nil {
			return installerDetails{}, err
		}
		defer f.Close()
		properties, err := wingetrepo.ReadMSIProperties(f)
		if err != nil {
			return installerDetails{}, err
		}
		return installerDetails{
			InstallerType: wingetrepo.InstallerTypeMSI,
			Name:          properties["ProductName"],
			Publisher:     properties["Manufacturer"],
			Version:       properties["ProductVersion"],
			ProductCode:   properties["ProductCode"],
		}, nil
	case ".exe":
		info, err := wingetrepo.ReadExeInfo(path)
		if err != nil {
			return installerDetails{}, err
		}
		// The architecture of an installer isn't the one of what it installs: 32-bit
		// setup programs often install 64-bit applications
		return installerDetails{
			InstallerType: info.InstallerType,
			Name:          info.ProductName,
			Publisher:     info.CompanyName,
			Version:       info.ProductVersion,
		}, nil
	case ".msix":
		info, err := wingetrepo.ReadMSIXInfo(path)
		if err != nil {
			return installerDetails{}, err
		}
		return installerDetails{
			InstallerType:     wingetrepo.InstallerTypeMSIX,
			Name:              info.DisplayName,
			Publisher:         info.Publisher,
			Version:           info.Version,
			PackageFamilyName: info.PackageFamilyName,
			Architecture:      info.Architecture,
		}, nil
	}
	return installerDetails{}, errors.New("unsupported installer")
}

// defaultPackageIdentifier builds a winget identifier like Publisher.Name from the
// installer's publisher and product name. A version in the product name, as in
// "7-Zip 24.08 (x64 edition)", is left out, so every version of a program gets the same
// identifier and winget offers the newer one as the update of the installed one.
func defaultPackageIdentifier(publisher, name, version string) string {
	name = removeVersion(name, version)
	// The publisher of an EXE's version resource may contain an e-mail address or a URL
	publisher = contactPattern.ReplaceAllString(publisher, " ")

	part := func(s string) string {
		// Keep the base letter of accented characters (Müller -> Muller)
		if plain, _, err := transform.String(transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn))), s); err == nil {
			s = plain
		}
		words := identifierPartPattern.Split(s, -1)
		for i, w := range words {
			if w != "" {
				r := []rune(w)
				r[0] = unicode.ToUpper(r[0])
				words[i] = string(r)
			}
		}
		joined := strings.Join(words, "")
		if len(joined) > 32 {
			joined = joined[:32]
		}
		return joined
	}
	return fmt.Sprintf("%s.%s", part(publisher), part(name))
}

// contactPattern matches e-mail addresses and web addresses, with optional brackets
var contactPattern = regexp.MustCompile(`(?i)[(<\[]?\s*(\S+@\S+\.\S+|https?://\S+|www\.\S+)\s*[)>\]]?`)

// removeVersion removes the version, or its first parts (24.08 of 24.08.00.0), from a
// product name, when it stands as a word of its own, optionally with a leading "v". At
// least two parts must match, so the 7 of "7-Zip" isn't taken for version 7.
func removeVersion(name, version string) string {
	parts := strings.Split(strings.TrimSpace(version), ".")
	for n := len(parts); n >= 2; n-- {
		v := strings.Join(parts[:n], ".")
		pattern := regexp.MustCompile(`(?i)(^|[^\pL\pN.])v?` + regexp.QuoteMeta(v) + `($|[^\pL\pN.])`)
		if pattern.MatchString(name) {
			return pattern.ReplaceAllString(name, "$1 $2")
		}
	}
	return name
}

func firstNotEmpty(values ...string) string {
	for _, v := range values {
		if v = strings.TrimSpace(v); v != "" {
			return v
		}
	}
	return ""
}
