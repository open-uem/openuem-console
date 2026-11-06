package wingetrepo

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testPackage(id, version string) Package {
	return Package{
		PackageIdentifier: id,
		PackageVersion:    version,
		PackageName:       "Test App",
		Publisher:         "Example Corp",
		ProductCode:       "{0D8B6A0E-1F2A-4C3B-9D4E-5F6A7B8C9D0E}",
		Architecture:      "neutral",
		InstallerFile:     "test.msi",
	}
}

func addPackage(t *testing.T, l *Library, p Package, content string) Package {
	staged, err := l.Stage(strings.NewReader(content))
	require.NoError(t, err)
	added, err := l.Add(p, staged)
	require.NoError(t, err)
	return added
}

func TestLibrary(t *testing.T) {
	l, err := NewLibrary(t.TempDir())
	require.NoError(t, err)

	added := addPackage(t, l, testPackage("Example.TestApp", "1.9.0"), "v1.9")
	assert.Len(t, added.InstallerSha256, 64)
	assert.Equal(t, strings.ToUpper(added.InstallerSha256), added.InstallerSha256)
	assert.Equal(t, int64(4), added.Size)

	addPackage(t, l, testPackage("Example.TestApp", "1.10.0"), "v1.10")
	addPackage(t, l, testPackage("Another.App", "2.0"), "other")

	all, err := l.List()
	require.NoError(t, err)
	require.Len(t, all, 3)
	assert.Equal(t, "Another.App", all[0].PackageIdentifier)
	assert.Equal(t, "1.10.0", all[1].PackageVersion, "newest version first")
	assert.Equal(t, "1.9.0", all[2].PackageVersion)

	path, p, err := l.InstallerPath("example.testapp", "1.9.0")
	require.NoError(t, err)
	assert.Equal(t, "1.9.0", p.PackageVersion)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "v1.9", string(data))

	// Replacing a version keeps a single copy
	addPackage(t, l, testPackage("Example.TestApp", "1.9.0"), "v1.9 fixed")
	versions, err := l.Versions("Example.TestApp")
	require.NoError(t, err)
	assert.Len(t, versions, 2)

	require.NoError(t, l.Remove("Example.TestApp", "1.9.0"))
	assert.ErrorIs(t, l.Remove("Example.TestApp", "1.9.0"), ErrPackageNotFound)
	_, _, err = l.InstallerPath("Example.TestApp", "1.9.0")
	assert.ErrorIs(t, err, ErrPackageNotFound)
}

func TestPackageValidate(t *testing.T) {
	assert.NoError(t, testPackage("Example.App", "1.0").Validate())

	for name, p := range map[string]Package{
		"identifier without dot":      testPackage("App", "1.0"),
		"identifier with traversal":   testPackage("../Example.App", "1.0"),
		"identifier with separator":   testPackage("Example/App.X", "1.0"),
		"version with traversal":      testPackage("Example.App", ".."),
		"version with separator":      testPackage("Example.App", "1/0"),
		"installer that's not an msi": func() Package { p := testPackage("Example.App", "1.0"); p.InstallerFile = "a.exe"; return p }(),
		"installer with traversal":    func() Package { p := testPackage("Example.App", "1.0"); p.InstallerFile = "../a.msi"; return p }(),
		"unknown architecture":        func() Package { p := testPackage("Example.App", "1.0"); p.Architecture = "mips"; return p }(),
		"no name":                     func() Package { p := testPackage("Example.App", "1.0"); p.PackageName = " "; return p }(),
	} {
		assert.ErrorIs(t, p.Validate(), ErrInvalidPackage, name)
	}
}

func TestCompareVersions(t *testing.T) {
	assert.Equal(t, 1, CompareVersions("1.10", "1.9"))
	assert.Equal(t, -1, CompareVersions("1.2", "1.2.1"))
	assert.Equal(t, 0, CompareVersions("1.2.0", "1.2"))
	assert.Equal(t, 1, CompareVersions("2.0-beta", "1.99"))
}

func TestSearch(t *testing.T) {
	packages := []Package{testPackage("Example.TestApp", "2.0"), testPackage("Example.TestApp", "1.0"), testPackage("Other.Tool", "1.0")}
	packages[2].PackageName = "Other Tool"
	packages[2].ProductCode = ""

	ids := func(r SearchResponse) []string {
		out := []string{}
		for _, d := range r.Data {
			out = append(out, d.PackageIdentifier)
		}
		return out
	}

	// No query: every package, versions grouped
	all := Search(packages, SearchRequest{})
	assert.Equal(t, []string{"Example.TestApp", "Other.Tool"}, ids(all))
	assert.Len(t, all.Data[0].Versions, 2)

	// winget search test
	assert.Equal(t, []string{"Example.TestApp"}, ids(Search(packages, SearchRequest{Query: &RequestMatch{KeyWord: "test", MatchType: "Substring"}})))
	// winget install --id Other.Tool --exact
	assert.Equal(t, []string{"Other.Tool"}, ids(Search(packages, SearchRequest{Filters: []MatchFilter{{PackageMatchField: "PackageIdentifier", RequestMatch: RequestMatch{KeyWord: "Other.Tool", MatchType: "Exact"}}}})))
	assert.Empty(t, Search(packages, SearchRequest{Filters: []MatchFilter{{PackageMatchField: "PackageIdentifier", RequestMatch: RequestMatch{KeyWord: "other.tool", MatchType: "Exact"}}}}).Data)
	// Correlation of installed packages by product code (winget upgrade/list)
	byProductCode := Search(packages, SearchRequest{Inclusions: []MatchFilter{{PackageMatchField: "ProductCode", RequestMatch: RequestMatch{KeyWord: "{0d8b6a0e-1f2a-4c3b-9d4e-5f6a7b8c9d0e}", MatchType: "Exact"}}}})
	assert.Equal(t, []string{"Example.TestApp"}, ids(byProductCode))
	assert.Equal(t, []string{"{0D8B6A0E-1F2A-4C3B-9D4E-5F6A7B8C9D0E}"}, byProductCode.Data[0].Versions[0].ProductCodes)
	// Correlation by name and publisher for packages without a product code
	assert.Equal(t, []string{"Other.Tool"}, ids(Search(packages, SearchRequest{Inclusions: []MatchFilter{{PackageMatchField: "NormalizedPackageNameAndPublisher", RequestMatch: RequestMatch{KeyWord: "othertoolexamplecorp", MatchType: "Exact"}}}})))
	// Unsupported fields never match
	assert.Empty(t, Search(packages, SearchRequest{Inclusions: []MatchFilter{{PackageMatchField: "Moniker", RequestMatch: RequestMatch{KeyWord: "test", MatchType: "Substring"}}}}).Data)
	// Maximum results
	assert.Len(t, Search(packages, SearchRequest{MaximumResults: 1}).Data, 1)
}

func TestBuildManifest(t *testing.T) {
	p := testPackage("Example.TestApp", "2.0")
	p.InstallerSha256 = strings.Repeat("A", 64)
	m := BuildManifest([]Package{p}, "https://console.example.com:1324/winget/agent/token/installers")

	data, err := json.Marshal(m)
	require.NoError(t, err)
	assert.JSONEq(t, `{"Data":{"PackageIdentifier":"Example.TestApp","Versions":[{
		"PackageVersion":"2.0",
		"DefaultLocale":{"PackageLocale":"en-US","Publisher":"Example Corp","PackageName":"Test App","License":"Proprietary","ShortDescription":"Test App"},
		"Installers":[{
			"InstallerIdentifier":"Example.TestApp.2.0",
			"InstallerSha256":"`+strings.Repeat("A", 64)+`",
			"InstallerUrl":"https://console.example.com:1324/winget/agent/token/installers/Example.TestApp/2.0/test.msi",
			"Architecture":"neutral","InstallerType":"msi","Scope":"machine",
			"ProductCode":"{0D8B6A0E-1F2A-4C3B-9D4E-5F6A7B8C9D0E}",
			"InstallModes":["interactive","silent","silentWithProgress"],
			"AppsAndFeaturesEntries":[{"DisplayName":"Test App","Publisher":"Example Corp","DisplayVersion":"2.0","ProductCode":"{0D8B6A0E-1F2A-4C3B-9D4E-5F6A7B8C9D0E}"}]
		}]}]}}`, string(data))
}

func TestAgentToken(t *testing.T) {
	key := []byte("server key")
	token := AgentToken(key, "agent1")
	assert.True(t, ValidAgentToken(key, "agent1", token))
	assert.False(t, ValidAgentToken(key, "agent2", token), "token of another agent")
	assert.False(t, ValidAgentToken([]byte("other key"), "agent1", token), "token of another server")
	assert.False(t, ValidAgentToken(nil, "agent1", AgentToken(nil, "agent1")), "no key")
	assert.Equal(t, "https://console.example.com:1324/winget/agent1/"+token, SourceURL("https://console.example.com:1324", key, "agent1"))
}
