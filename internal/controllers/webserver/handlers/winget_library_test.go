package handlers

import (
	"path/filepath"
	"testing"

	"github.com/open-uem/openuem-console/internal/wingetrepo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultPackageIdentifier(t *testing.T) {
	assert.Equal(t, "ExampleCorp.TestApp", defaultPackageIdentifier("Example Corp.", "Test App", "1.0"))
	assert.Equal(t, "MullerSohne.TestAppUnicode", defaultPackageIdentifier("Müller & Söhne", "Test App ünicode", "1.0"))
	assert.Equal(t, "Mozilla.FirefoxESRX64DeDE", defaultPackageIdentifier("Mozilla", "Firefox ESR (x64 de-DE)", "128.0"))

	// The version in a product name is left out, so every version of a program gets the
	// same identifier and winget offers the newer one as its update
	assert.Equal(t, "IgorPavlov.7ZipX64Edition", defaultPackageIdentifier("Igor Pavlov", "7-Zip 24.08 (x64 edition)", "24.08.00.0"))
	assert.Equal(t, "IgorPavlov.7ZipX64Edition", defaultPackageIdentifier("Igor Pavlov", "7-Zip 24.09 (x64 edition)", "24.09.00.0"))
	assert.Equal(t, "Mozilla.FirefoxESRX64DeDE", defaultPackageIdentifier("Mozilla", "Firefox 128.5.1 ESR (x64 de-DE)", "128.5.1"))
	assert.Equal(t, "ExampleCorp.Tool", defaultPackageIdentifier("Example Corp", "Tool v3.1.4", "3.1.4.0"))
	assert.Equal(t, "ExampleCorp.Tool2", defaultPackageIdentifier("Example Corp", "Tool 2", "2.0"), "a single number isn't a version")
	assert.Equal(t, "ExampleCorp.Tool2410", defaultPackageIdentifier("Example Corp", "Tool 24.10", "24.08"), "another version stays")

	// Contact details in the publisher of an EXE's version resource are left out
	assert.Equal(t, "DonHO.Notepad", defaultPackageIdentifier("Don HO don.h@free.fr", "Notepad++", "8.7"))
	assert.Equal(t, "ExampleCorp.Tool", defaultPackageIdentifier("Example Corp (https://example.com)", "Tool", "1.0"))
	assert.Equal(t, "ExampleCorp.Tool", defaultPackageIdentifier("Example Corp www.example.com", "Tool", "1.0"))

	p := wingetrepo.Package{PackageIdentifier: defaultPackageIdentifier("7-Zip", "7-Zip 24.08 (x64)", "24.08"), PackageVersion: "24.08", PackageName: "7-Zip", Publisher: "7-Zip", Architecture: "neutral", InstallerFile: "7z.msi"}
	assert.Equal(t, "7Zip.7ZipX64", p.PackageIdentifier)
	assert.NoError(t, p.Validate(), "derived identifiers are valid")
}

func TestInspectInstaller(t *testing.T) {
	testdata := filepath.Join("..", "..", "..", "wingetrepo", "testdata")

	msi, err := inspectInstaller(filepath.Join(testdata, "test.msi"), ".msi")
	require.NoError(t, err)
	assert.Equal(t, wingetrepo.InstallerTypeMSI, msi.InstallerType)
	assert.NotEmpty(t, msi.ProductCode)

	exe, err := inspectInstaller(filepath.Join(testdata, "test-inno.exe"), ".exe")
	require.NoError(t, err)
	assert.Equal(t, installerDetails{InstallerType: wingetrepo.InstallerTypeInno, Name: "Example Tool", Publisher: "Example Corp", Version: "3.1.4"}, exe)

	_, err = inspectInstaller(filepath.Join(testdata, "test.exe"), ".msi")
	assert.Error(t, err, "an exe renamed to .msi")
	_, err = inspectInstaller(filepath.Join(testdata, "test.msi"), ".exe")
	assert.Error(t, err, "an msi renamed to .exe")
	_, err = inspectInstaller(filepath.Join(testdata, "test.msi"), ".zip")
	assert.Error(t, err)
}

func TestHiddenLibraryPackageIDs(t *testing.T) {
	pkg := func(id string, tenantID int) wingetrepo.Package {
		return wingetrepo.Package{PackageIdentifier: id, TenantID: tenantID}
	}
	all := []wingetrepo.Package{pkg("Org1.App", 1), pkg("Org2.App", 2), pkg("Global.App", 0), pkg("Shared.App", 2), pkg("Shared.App", 1)}

	assert.Equal(t, []string{"Org2.App"}, hiddenLibraryPackageIDs(all, 1), "another organization's packages are hidden, unless a version is visible")
	assert.Equal(t, []string{"Org1.App"}, hiddenLibraryPackageIDs(all, 2))
	assert.Empty(t, hiddenLibraryPackageIDs(all, -1), "all organizations")
}

func TestHiddenLibraryPackageIDsForProfile(t *testing.T) {
	pkg := func(id string, tenantID int) wingetrepo.Package {
		return wingetrepo.Package{PackageIdentifier: id, TenantID: tenantID}
	}
	all := []wingetrepo.Package{pkg("Org1.App", 1), pkg("Org2.App", 2), pkg("Global.App", 0)}

	assert.Equal(t, []string{"Org2.App"}, hiddenLibraryPackageIDsForProfile(all, []int{1}), "profile of an organization")
	assert.Equal(t, []string{"Org1.App", "Org2.App"}, hiddenLibraryPackageIDsForProfile(all, nil), "global profile: only packages of all organizations")
	assert.Equal(t, []string{"Org1.App", "Org2.App"}, hiddenLibraryPackageIDsForProfile(all, []int{1, 2}), "profile of several organizations")
}
