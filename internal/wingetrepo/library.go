// Package wingetrepo is the console's private winget repository: a library of packages
// uploaded by the administrators, published to the agents as a WinGet REST source.
//
// Each package version is stored in <dir>/<package id>/<version>/ with its installer and
// a package.json file with its metadata.
package wingetrepo

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

const metadataFile = "package.json"

var (
	ErrInvalidPackage  = errors.New("invalid package")
	ErrPackageNotFound = errors.New("package not found")

	// The same rules as winget's manifest schema: 2 to 8 parts separated by dots
	packageIdentifierPattern = regexp.MustCompile(`^[^\.\s\\/:\*\?"<>\|\x01-\x1f]{1,32}(\.[^\.\s\\/:\*\?"<>\|\x01-\x1f]{1,32}){1,7}$`)
	packageVersionPattern    = regexp.MustCompile(`^[0-9A-Za-z][0-9A-Za-z.+_-]{0,63}$`)
	installerFilePattern     = regexp.MustCompile(`^[0-9A-Za-z][0-9A-Za-z._+-]{0,127}\.msi$`)
)

// Package is a version of a package of the library
type Package struct {
	PackageIdentifier string    `json:"packageIdentifier"`
	PackageVersion    string    `json:"packageVersion"`
	PackageName       string    `json:"packageName"`
	Publisher         string    `json:"publisher"`
	ProductCode       string    `json:"productCode,omitempty"`
	Architecture      string    `json:"architecture"`
	InstallerFile     string    `json:"installerFile"`
	InstallerSha256   string    `json:"installerSha256"`
	Size              int64     `json:"size"`
	TenantID          int       `json:"tenantId,omitempty"`
	Uploaded          time.Time `json:"uploaded"`
}

// Validate checks the fields that are used in paths, URLs and winget manifests
func (p Package) Validate() error {
	switch {
	case !packageIdentifierPattern.MatchString(p.PackageIdentifier),
		!packageVersionPattern.MatchString(p.PackageVersion),
		!installerFilePattern.MatchString(p.InstallerFile),
		strings.TrimSpace(p.PackageName) == "",
		strings.TrimSpace(p.Publisher) == "",
		!validArchitectures[p.Architecture]:
		return ErrInvalidPackage
	}
	return nil
}

var validArchitectures = map[string]bool{"neutral": true, "x86": true, "x64": true, "arm64": true}

// VisibleTo reports if the package can be installed by agents of the given organizations.
// Packages without an organization are available to all of them.
func (p Package) VisibleTo(tenantIDs []int) bool {
	if p.TenantID == 0 {
		return true
	}
	for _, id := range tenantIDs {
		if id == p.TenantID {
			return true
		}
	}
	return false
}

// Library stores the uploaded packages
type Library struct {
	dir string
	mu  sync.RWMutex
}

// NewLibrary creates the library, creating its directory if needed
func NewLibrary(dir string) (*Library, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return &Library{dir: dir}, nil
}

// Stage copies an upload to a temporary file of the library, so it can be inspected
// before it's added. The caller removes the file if it isn't added.
func (l *Library) Stage(src io.Reader) (string, error) {
	tmp, err := os.CreateTemp(l.dir, ".upload-*")
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(tmp, src); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return "", err
	}
	return tmp.Name(), nil
}

// Add moves a staged installer into the library. The installer's hash and size are set
// from the file. An existing package with the same identifier and version is replaced.
func (l *Library) Add(p Package, stagedPath string) (Package, error) {
	f, err := os.Open(stagedPath)
	if err != nil {
		return Package{}, err
	}
	hash := sha256.New()
	size, err := io.Copy(hash, f)
	f.Close()
	if err != nil {
		return Package{}, err
	}

	p.InstallerSha256 = strings.ToUpper(hex.EncodeToString(hash.Sum(nil)))
	p.Size = size
	if p.Uploaded.IsZero() {
		p.Uploaded = time.Now()
	}
	if err := p.Validate(); err != nil {
		return Package{}, err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	dir := l.versionDir(p.PackageIdentifier, p.PackageVersion)
	if err := os.RemoveAll(dir); err != nil {
		return Package{}, err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return Package{}, err
	}
	if err := os.Rename(stagedPath, filepath.Join(dir, p.InstallerFile)); err != nil {
		return Package{}, err
	}

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return Package{}, err
	}
	if err := os.WriteFile(filepath.Join(dir, metadataFile), data, 0o600); err != nil {
		return Package{}, err
	}
	return p, nil
}

// List returns every package version, sorted by identifier and newest version first
func (l *Library) List() ([]Package, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	matches, err := filepath.Glob(filepath.Join(l.dir, "*", "*", metadataFile))
	if err != nil {
		return nil, err
	}

	packages := []Package{}
	for _, m := range matches {
		data, err := os.ReadFile(m)
		if err != nil {
			continue
		}
		p := Package{}
		if json.Unmarshal(data, &p) != nil || p.Validate() != nil {
			continue
		}
		packages = append(packages, p)
	}

	sort.Slice(packages, func(i, j int) bool {
		if packages[i].PackageIdentifier != packages[j].PackageIdentifier {
			return strings.ToLower(packages[i].PackageIdentifier) < strings.ToLower(packages[j].PackageIdentifier)
		}
		return CompareVersions(packages[i].PackageVersion, packages[j].PackageVersion) > 0
	})
	return packages, nil
}

// Versions returns the versions of a package, newest first
func (l *Library) Versions(packageIdentifier string) ([]Package, error) {
	all, err := l.List()
	if err != nil {
		return nil, err
	}
	versions := []Package{}
	for _, p := range all {
		if strings.EqualFold(p.PackageIdentifier, packageIdentifier) {
			versions = append(versions, p)
		}
	}
	return versions, nil
}

// InstallerPath returns the installer of a package version
func (l *Library) InstallerPath(packageIdentifier, packageVersion string) (string, Package, error) {
	versions, err := l.Versions(packageIdentifier)
	if err != nil {
		return "", Package{}, err
	}
	for _, p := range versions {
		if p.PackageVersion == packageVersion {
			return filepath.Join(l.versionDir(p.PackageIdentifier, p.PackageVersion), p.InstallerFile), p, nil
		}
	}
	return "", Package{}, ErrPackageNotFound
}

// Remove deletes a package version
func (l *Library) Remove(packageIdentifier, packageVersion string) error {
	if !packageIdentifierPattern.MatchString(packageIdentifier) || !packageVersionPattern.MatchString(packageVersion) {
		return ErrInvalidPackage
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	dir := l.versionDir(packageIdentifier, packageVersion)
	if _, err := os.Stat(filepath.Join(dir, metadataFile)); err != nil {
		return ErrPackageNotFound
	}
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	// Only removed if no other version is left
	_ = os.Remove(filepath.Dir(dir))
	return nil
}

func (l *Library) versionDir(packageIdentifier, packageVersion string) string {
	return filepath.Join(l.dir, packageIdentifier, packageVersion)
}

// CompareVersions compares two versions part by part, numerically when both parts are
// numbers, e.g. 1.10 > 1.9. It returns -1, 0 or 1.
func CompareVersions(a, b string) int {
	pa, pb := strings.FieldsFunc(a, isVersionSeparator), strings.FieldsFunc(b, isVersionSeparator)
	for i := 0; i < len(pa) || i < len(pb); i++ {
		var x, y string
		if i < len(pa) {
			x = pa[i]
		}
		if i < len(pb) {
			y = pb[i]
		}
		if c := compareVersionPart(x, y); c != 0 {
			return c
		}
	}
	return 0
}

func isVersionSeparator(r rune) bool { return r == '.' || r == '-' || r == '+' || r == '_' }

func compareVersionPart(a, b string) int {
	na, aIsNumber := parseVersionNumber(a)
	nb, bIsNumber := parseVersionNumber(b)
	switch {
	case aIsNumber && bIsNumber:
		switch {
		case na < nb:
			return -1
		case na > nb:
			return 1
		}
		return 0
	case a == b:
		return 0
	case a < b:
		return -1
	default:
		return 1
	}
}

func parseVersionNumber(s string) (uint64, bool) {
	if s == "" {
		return 0, true
	}
	var n uint64
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
		n = n*10 + uint64(r-'0')
	}
	return n, true
}
