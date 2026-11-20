package wingetrepo

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

// SourceIdentifier identifies the console's REST source to winget
const SourceIdentifier = "OpenUEM"

// ServerSupportedVersions are the versions of the WinGet REST source API implemented here
var ServerSupportedVersions = []string{"1.0.0", "1.1.0"}

// Information is the answer to GET /information
type Information struct {
	Data InformationData `json:"Data"`
}

type InformationData struct {
	SourceIdentifier        string   `json:"SourceIdentifier"`
	ServerSupportedVersions []string `json:"ServerSupportedVersions"`
}

// NewInformation describes the source
func NewInformation() Information {
	return Information{Data: InformationData{SourceIdentifier: SourceIdentifier, ServerSupportedVersions: ServerSupportedVersions}}
}

// ErrorResponse is the body of an error answer of the REST source
type ErrorResponse struct {
	ErrorCode    int    `json:"ErrorCode"`
	ErrorMessage string `json:"ErrorMessage"`
}

// SearchRequest is the body of POST /manifestSearch
type SearchRequest struct {
	MaximumResults    int           `json:"MaximumResults,omitempty"`
	FetchAllManifests bool          `json:"FetchAllManifests,omitempty"`
	Query             *RequestMatch `json:"Query,omitempty"`
	Inclusions        []MatchFilter `json:"Inclusions,omitempty"`
	Filters           []MatchFilter `json:"Filters,omitempty"`
}

type RequestMatch struct {
	KeyWord   string `json:"KeyWord"`
	MatchType string `json:"MatchType"`
}

type MatchFilter struct {
	PackageMatchField string       `json:"PackageMatchField"`
	RequestMatch      RequestMatch `json:"RequestMatch"`
}

// SearchResponse is the answer to POST /manifestSearch
type SearchResponse struct {
	Data []SearchResult `json:"Data"`
}

type SearchResult struct {
	PackageIdentifier string          `json:"PackageIdentifier"`
	PackageName       string          `json:"PackageName"`
	Publisher         string          `json:"Publisher"`
	Versions          []SearchVersion `json:"Versions"`
}

type SearchVersion struct {
	PackageVersion string   `json:"PackageVersion"`
	ProductCodes   []string `json:"ProductCodes,omitempty"`
}

// Search finds the packages matching a request. As in the reference implementation, the
// query and the inclusions are alternatives and every filter must match. A request
// without query and inclusions returns every package.
//
// A package matches when one of its versions does, and the response lists all of its
// versions: winget finds an installed program by the product code of the version that is
// installed, and offers the newest version of the response as its update.
func Search(packages []Package, req SearchRequest) SearchResponse {
	matched := map[string]bool{}
	for _, p := range packages {
		if searchMatches(p, req) {
			matched[strings.ToLower(p.PackageIdentifier)] = true
		}
	}

	results := []SearchResult{}
	index := map[string]int{}
	for _, p := range packages {
		key := strings.ToLower(p.PackageIdentifier)
		if !matched[key] {
			continue
		}

		version := SearchVersion{PackageVersion: p.PackageVersion}
		if p.ProductCode != "" {
			version.ProductCodes = []string{p.ProductCode}
		}

		if i, ok := index[key]; ok {
			results[i].Versions = append(results[i].Versions, version)
			continue
		}
		if req.MaximumResults > 0 && len(results) >= req.MaximumResults {
			continue
		}
		index[key] = len(results)
		results = append(results, SearchResult{
			PackageIdentifier: p.PackageIdentifier,
			PackageName:       p.PackageName,
			Publisher:         p.Publisher,
			Versions:          []SearchVersion{version},
		})
	}
	return SearchResponse{Data: results}
}

// searchMatches reports whether a version of a package matches a search request
func searchMatches(p Package, req SearchRequest) bool {
	matches := req.Query == nil && len(req.Inclusions) == 0
	if req.Query != nil && (matchValue(p.PackageIdentifier, *req.Query) || matchValue(p.PackageName, *req.Query)) {
		matches = true
	}
	for _, inclusion := range req.Inclusions {
		if matchField(p, inclusion) {
			matches = true
		}
	}
	for _, filter := range req.Filters {
		if !matchField(p, filter) {
			matches = false
		}
	}
	return matches
}

// matchField matches the package field a filter is about. Fields the library doesn't
// have (monikers, commands, tags, package family names, markets) never match.
func matchField(p Package, filter MatchFilter) bool {
	switch filter.PackageMatchField {
	case "PackageIdentifier":
		return matchValue(p.PackageIdentifier, filter.RequestMatch)
	case "PackageName":
		return matchValue(p.PackageName, filter.RequestMatch)
	case "ProductCode":
		return p.ProductCode != "" && strings.EqualFold(p.ProductCode, filter.RequestMatch.KeyWord)
	case "NormalizedPackageNameAndPublisher":
		return normalize(filter.RequestMatch.KeyWord) == normalize(p.PackageName+p.Publisher)
	}
	return false
}

func matchValue(value string, m RequestMatch) bool {
	keyword := m.KeyWord
	switch m.MatchType {
	case "Exact":
		return value == keyword
	case "CaseInsensitive":
		return strings.EqualFold(value, keyword)
	case "StartsWith":
		return strings.HasPrefix(strings.ToLower(value), strings.ToLower(keyword))
	default:
		// Substring, Wildcard, Fuzzy and FuzzySubstring
		return strings.Contains(strings.ToLower(value), strings.ToLower(keyword))
	}
}

// normalize approximates winget's normalized name and publisher: only letters and digits,
// lower case
func normalize(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, s)
}

// ManifestResponse is the answer to GET /packageManifests/{PackageIdentifier}
type ManifestResponse struct {
	Data Manifest `json:"Data"`
}

type Manifest struct {
	PackageIdentifier string            `json:"PackageIdentifier"`
	Versions          []ManifestVersion `json:"Versions"`
}

type ManifestVersion struct {
	PackageVersion string      `json:"PackageVersion"`
	DefaultLocale  Locale      `json:"DefaultLocale"`
	Installers     []Installer `json:"Installers"`
}

type Locale struct {
	PackageLocale    string `json:"PackageLocale"`
	Publisher        string `json:"Publisher"`
	PackageName      string `json:"PackageName"`
	License          string `json:"License"`
	ShortDescription string `json:"ShortDescription"`
}

type Installer struct {
	InstallerIdentifier    string                 `json:"InstallerIdentifier"`
	InstallerSha256        string                 `json:"InstallerSha256"`
	InstallerUrl           string                 `json:"InstallerUrl"`
	Architecture           string                 `json:"Architecture"`
	InstallerType          string                 `json:"InstallerType"`
	Scope                  string                 `json:"Scope"`
	ProductCode            string                 `json:"ProductCode,omitempty"`
	PackageFamilyName      string                 `json:"PackageFamilyName,omitempty"`
	InstallModes           []string               `json:"InstallModes"`
	InstallerSwitches      *InstallerSwitches     `json:"InstallerSwitches,omitempty"`
	AppsAndFeaturesEntries []AppsAndFeaturesEntry `json:"AppsAndFeaturesEntries,omitempty"`
}

// InstallerSwitches replace winget's default switches of an installer type. Generic exe
// installers have no default switches.
type InstallerSwitches struct {
	Silent             string `json:"Silent,omitempty"`
	SilentWithProgress string `json:"SilentWithProgress,omitempty"`
}

type AppsAndFeaturesEntry struct {
	DisplayName    string `json:"DisplayName,omitempty"`
	Publisher      string `json:"Publisher,omitempty"`
	DisplayVersion string `json:"DisplayVersion,omitempty"`
	ProductCode    string `json:"ProductCode,omitempty"`
}

// BuildManifest describes the versions of a package. installersURL is the URL the
// installers are downloaded from, e.g. https://console:1324/winget/<agent>/<token>/installers
func BuildManifest(versions []Package, installersURL string) ManifestResponse {
	m := Manifest{Versions: []ManifestVersion{}}
	for _, p := range versions {
		m.PackageIdentifier = p.PackageIdentifier
		// The agent installs with --scope machine, as SYSTEM
		installer := Installer{
			InstallerIdentifier: p.PackageIdentifier + "." + p.PackageVersion,
			InstallerSha256:     p.InstallerSha256,
			InstallerUrl:        fmt.Sprintf("%s/%s/%s/%s", installersURL, url.PathEscape(p.PackageIdentifier), url.PathEscape(p.PackageVersion), url.PathEscape(p.InstallerFile)),
			Architecture:        p.Architecture,
			InstallerType:       p.Type(),
			Scope:               "machine",
			ProductCode:         p.ProductCode,
			InstallModes:        []string{"interactive", "silent", "silentWithProgress"},
		}
		switch p.Type() {
		case InstallerTypeMSIX:
			// winget recognizes installed MSIX packages by their family name
			installer.PackageFamilyName = p.PackageFamilyName
			installer.ProductCode = ""
		default:
			if p.SilentSwitches != "" {
				installer.InstallerSwitches = &InstallerSwitches{Silent: p.SilentSwitches, SilentWithProgress: p.SilentSwitches}
			}
			// winget matches installed programs (Apps & features) with these, which also
			// lets it find the installed version of exe installers without a product code
			if p.ProductCode != "" || p.Type() != InstallerTypeMSI {
				installer.AppsAndFeaturesEntries = []AppsAndFeaturesEntry{{
					DisplayName: p.PackageName, Publisher: p.Publisher, DisplayVersion: p.PackageVersion, ProductCode: p.ProductCode,
				}}
			}
		}
		m.Versions = append(m.Versions, ManifestVersion{
			PackageVersion: p.PackageVersion,
			DefaultLocale: Locale{
				PackageLocale:    "en-US",
				Publisher:        p.Publisher,
				PackageName:      p.PackageName,
				License:          "Proprietary",
				ShortDescription: p.PackageName,
			},
			Installers: []Installer{installer},
		})
	}
	return ManifestResponse{Data: m}
}

// AgentToken is the secret part of an agent's source URL. It's derived from the agent id
// with a server key, so nothing has to be stored to check it.
func AgentToken(key []byte, agentID string) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte("openuem-winget-source:" + agentID))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// ValidAgentToken checks an agent's token in constant time
func ValidAgentToken(key []byte, agentID, token string) bool {
	return len(key) > 0 && agentID != "" && hmac.Equal([]byte(AgentToken(key, agentID)), []byte(token))
}

// SourceURL is the URL of the REST source for an agent. baseURL is the auth server's address.
func SourceURL(baseURL string, key []byte, agentID string) string {
	return fmt.Sprintf("%s/winget/%s/%s", baseURL, url.PathEscape(agentID), AgentToken(key, agentID))
}
