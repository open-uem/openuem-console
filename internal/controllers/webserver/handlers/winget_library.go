package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/go-co-op/gocron/v2"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	openuem_nats "github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/taskfiles"
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
		return RenderError(c, partials.ErrorMessage(i18n.T(ctx, "tasks.msi_upload_no_file"), false))
	}
	if !strings.EqualFold(filepath.Ext(header.Filename), ".msi") {
		return RenderError(c, partials.ErrorMessage(i18n.T(ctx, "tasks.msi_upload_not_msi"), false))
	}

	src, err := header.Open()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	defer src.Close()

	staged, err := h.WinGetLibrary.Stage(src)
	if err != nil {
		log.Printf("[ERROR]: could not save the uploaded package, reason: %v", err)
		return RenderError(c, partials.ErrorMessage(i18n.T(ctx, "tasks.msi_upload_failed"), false))
	}
	defer os.Remove(staged)

	properties, err := readStagedMSIProperties(staged)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(ctx, "library.not_msi"), false))
	}

	p := wingetrepo.Package{
		PackageName:       firstNotEmpty(c.FormValue("library-name"), properties["ProductName"]),
		Publisher:         firstNotEmpty(c.FormValue("library-publisher"), properties["Manufacturer"]),
		PackageVersion:    firstNotEmpty(c.FormValue("library-version"), properties["ProductVersion"]),
		ProductCode:       firstNotEmpty(c.FormValue("library-product-code"), properties["ProductCode"]),
		Architecture:      firstNotEmpty(c.FormValue("library-architecture"), "neutral"),
		InstallerFile:     taskfiles.SanitizeName(header.Filename),
		PackageIdentifier: strings.TrimSpace(c.FormValue("library-id")),
	}
	if p.PackageIdentifier == "" {
		p.PackageIdentifier = defaultPackageIdentifier(p.Publisher, p.PackageName)
	}
	if !strings.HasSuffix(strings.ToLower(p.InstallerFile), ".msi") {
		p.InstallerFile += ".msi"
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
		return RenderError(c, partials.ErrorMessage(i18n.T(ctx, "tasks.msi_upload_failed"), false))
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
		URL:     wingetrepo.SourceURL(h.agentFilesBaseURL(), []byte(h.JWTKey), agentID),
		TrustCA: h.ReverseProxyServer == "",
	}
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

func readStagedMSIProperties(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return wingetrepo.ReadMSIProperties(f)
}

// defaultPackageIdentifier builds a winget identifier like Publisher.Name from the MSI's
// manufacturer and product name
func defaultPackageIdentifier(publisher, name string) string {
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

func firstNotEmpty(values ...string) string {
	for _, v := range values {
		if v = strings.TrimSpace(v); v != "" {
			return v
		}
	}
	return ""
}
