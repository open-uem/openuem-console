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
		"identifier without dot":        testPackage("App", "1.0"),
		"identifier with traversal":     testPackage("../Example.App", "1.0"),
		"identifier with separator":     testPackage("Example/App.X", "1.0"),
		"version with traversal":        testPackage("Example.App", ".."),
		"version with separator":        testPackage("Example.App", "1/0"),
		"installer that's not an msi":   func() Package { p := testPackage("Example.App", "1.0"); p.InstallerFile = "a.exe"; return p }(),
		"installer with traversal":      func() Package { p := testPackage("Example.App", "1.0"); p.InstallerFile = "../a.msi"; return p }(),
		"unknown architecture":          func() Package { p := testPackage("Example.App", "1.0"); p.Architecture = "mips"; return p }(),
		"no name":                       func() Package { p := testPackage("Example.App", "1.0"); p.PackageName = " "; return p }(),
		"exe without silent switches":   exePackage(InstallerTypeEXE, ""),
		"switches on several lines":     exePackage(InstallerTypeEXE, "/S\r\n/D=C:\\"),
		"msi installer for an exe type": func() Package { p := exePackage(InstallerTypeInno, ""); p.InstallerFile = "setup.msi"; return p }(),
		"unknown installer type":        func() Package { p := testPackage("Example.App", "1.0"); p.InstallerType = "appx"; return p }(),
		"msix without family name":      func() Package { p := msixPackage(); p.PackageFamilyName = ""; return p }(),
	} {
		assert.ErrorIs(t, p.Validate(), ErrInvalidPackage, name)
	}

	assert.NoError(t, exePackage(InstallerTypeEXE, "/S").Validate(), "generic exe with switches")
	for _, installerType := range []string{InstallerTypeInno, InstallerTypeNullsoft, InstallerTypeBurn} {
		assert.NoError(t, exePackage(installerType, "").Validate(), installerType+" without switches")
	}
	assert.NoError(t, msixPackage().Validate())
	assert.Equal(t, InstallerTypeMSI, testPackage("Example.App", "1.0").Type(), "packages without a type are msi")
}

func exePackage(installerType, switches string) Package {
	p := testPackage("Example.Tool", "3.1.4")
	p.InstallerType = installerType
	p.InstallerFile = "setup.exe"
	p.ProductCode = ""
	p.SilentSwitches = switches
	return p
}

func msixPackage() Package {
	p := testPackage("Example.Notes", "2.5.0.0")
	p.InstallerType = InstallerTypeMSIX
	p.InstallerFile = "notes.msix"
	p.ProductCode = ""
	p.PackageFamilyName = "ExampleCorp.Notes_8wekyb3d8bbwe"
	return p
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
	// An installed older version, found by its own product code, is offered the newer
	// version: the response lists every version of the matching package
	packages[1].ProductCode = "{11111111-2222-3333-4444-555555555555}"
	byOldProductCode := Search(packages, SearchRequest{Inclusions: []MatchFilter{{PackageMatchField: "ProductCode", RequestMatch: RequestMatch{KeyWord: "{11111111-2222-3333-4444-555555555555}", MatchType: "Exact"}}}})
	require.Equal(t, []string{"Example.TestApp"}, ids(byOldProductCode))
	assert.Equal(t, []SearchVersion{
		{PackageVersion: "2.0", ProductCodes: []string{"{0D8B6A0E-1F2A-4C3B-9D4E-5F6A7B8C9D0E}"}},
		{PackageVersion: "1.0", ProductCodes: []string{"{11111111-2222-3333-4444-555555555555}"}},
	}, byOldProductCode.Data[0].Versions)
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

func TestBuildManifestInstallerTypes(t *testing.T) {
	installer := func(p Package) Installer {
		return BuildManifest([]Package{p}, "https://console/installers").Data.Versions[0].Installers[0]
	}

	exe := installer(exePackage(InstallerTypeEXE, "/S /norestart"))
	assert.Equal(t, "exe", exe.InstallerType)
	assert.Equal(t, "machine", exe.Scope)
	assert.Equal(t, &InstallerSwitches{Silent: "/S /norestart", SilentWithProgress: "/S /norestart"}, exe.InstallerSwitches)
	assert.Equal(t, []AppsAndFeaturesEntry{{DisplayName: "Test App", Publisher: "Example Corp", DisplayVersion: "3.1.4"}}, exe.AppsAndFeaturesEntries,
		"exe installers are matched with Apps & features by name and publisher")

	inno := installer(exePackage(InstallerTypeInno, ""))
	assert.Equal(t, "inno", inno.InstallerType)
	assert.Nil(t, inno.InstallerSwitches, "winget knows Inno Setup's switches")

	msix := installer(msixPackage())
	assert.Equal(t, "msix", msix.InstallerType)
	assert.Equal(t, "ExampleCorp.Notes_8wekyb3d8bbwe", msix.PackageFamilyName)
	assert.Empty(t, msix.ProductCode)
	assert.Empty(t, msix.AppsAndFeaturesEntries)
	assert.Nil(t, msix.InstallerSwitches)

	data, err := json.Marshal(installer(exePackage(InstallerTypeNullsoft, "")))
	require.NoError(t, err)
	assert.NotContains(t, string(data), "InstallerSwitches", "no empty switches in the manifest")
	assert.NotContains(t, string(data), "PackageFamilyName")
}

func TestSanitizeInstallerName(t *testing.T) {
	assert.Equal(t, "setup.msi", SanitizeInstallerName("setup.msi"))
	assert.Equal(t, "setup.exe", SanitizeInstallerName(`C:\Users\me\Downloads\setup.exe`))
	assert.Equal(t, "setup.msi", SanitizeInstallerName("../../setup.msi"))
	assert.Equal(t, "My_App_1.2_x64_.msix", SanitizeInstallerName("My App 1.2 (x64).msix"))
	assert.Equal(t, "a_b.msi", SanitizeInstallerName("a'\"b.msi"))
	assert.Equal(t, "", SanitizeInstallerName(".."))
	long := SanitizeInstallerName(strings.Repeat("a", 300) + ".msi")
	assert.Len(t, long, 120)
	assert.True(t, installerFilePattern.MatchString(long), "a long name is still a valid installer name")
}

func TestLibraryOwnership(t *testing.T) {
	dir := t.TempDir()
	l, err := NewLibrary(dir)
	require.NoError(t, err)

	orgA := testPackage("Contoso.Tool", "1.2")
	orgA.TenantID = 1
	addPackage(t, l, orgA, "org A")

	stage := func(content string) string {
		staged, err := l.Stage(strings.NewReader(content))
		require.NoError(t, err)
		return staged
	}

	// Another organization, or all organizations, can't take over the identifier
	orgB := testPackage("Contoso.Tool", "1.2")
	orgB.TenantID = 2
	_, err = l.Add(orgB, stage("org B"))
	assert.ErrorIs(t, err, ErrPackageOwned)
	everyone := testPackage("contoso.tool", "2.0")
	_, err = l.Add(everyone, stage("all"))
	assert.ErrorIs(t, err, ErrPackageOwned, "case variant of the identifier")

	path, p, err := l.InstallerPath("Contoso.Tool", "1.2")
	require.NoError(t, err)
	assert.Equal(t, 1, p.TenantID)
	data, _ := os.ReadFile(path)
	assert.Equal(t, "org A", string(data), "org A's installer is kept")

	// The same organization can add versions and replace one, with the stored spelling
	newer := testPackage("CONTOSO.TOOL", "1.3")
	newer.TenantID = 1
	added, err := l.Add(newer, stage("org A 1.3"))
	require.NoError(t, err)
	assert.Equal(t, "Contoso.Tool", added.PackageIdentifier)
	replaced := testPackage("Contoso.Tool", "1.2")
	replaced.TenantID = 1
	_, err = l.Add(replaced, stage("org A again"))
	require.NoError(t, err)
	path, _, err = l.InstallerPath("Contoso.Tool", "1.2")
	require.NoError(t, err)
	data, _ = os.ReadFile(path)
	assert.Equal(t, "org A again", string(data))
	versions, err := l.Versions("contoso.tool")
	require.NoError(t, err)
	assert.Len(t, versions, 2)

	// Staged uploads left by a crash are removed when the library is opened
	leftover := stage("interrupted")
	_, err = NewLibrary(dir)
	require.NoError(t, err)
	_, err = os.Stat(leftover)
	assert.True(t, os.IsNotExist(err), "leftover staged upload removed")
}
