package wingetrepo

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadExeInfo(t *testing.T) {
	for file, installerType := range map[string]string{
		"test.exe":      InstallerTypeEXE,
		"test-inno.exe": InstallerTypeInno,
		"test-nsis.exe": InstallerTypeNullsoft,
		"test-burn.exe": InstallerTypeBurn,
	} {
		info, err := ReadExeInfo(filepath.Join("testdata", file))
		require.NoError(t, err, file)
		assert.Equal(t, ExeInfo{InstallerType: installerType, ProductName: "Example Tool", CompanyName: "Example Corp", ProductVersion: "3.1.4"}, info, file)
	}

	_, err := ReadExeInfo(filepath.Join("testdata", "test.msi"))
	assert.Error(t, err, "not an executable")
}

func TestScanExeMarkersAcrossBlocks(t *testing.T) {
	// The signature starts just before the end of the first block
	data := make([]byte, 1<<20+100)
	copy(data[1<<20-5:], "NullsoftInst")
	path := filepath.Join(t.TempDir(), "setup.exe")
	require.NoError(t, os.WriteFile(path, data, 0o600))

	installerType, err := scanExeMarkers(path)
	require.NoError(t, err)
	assert.Equal(t, InstallerTypeNullsoft, installerType)
}

func TestNormalizeExeVersion(t *testing.T) {
	assert.Equal(t, "1.2.3.4", normalizeExeVersion("1, 2, 3, 4"))
	assert.Equal(t, "24.08", normalizeExeVersion(" 24.08 "))
	assert.Equal(t, "", normalizeExeVersion("version 2 (beta)"))
}

func TestPublisherID(t *testing.T) {
	// The publisher id of Microsoft's own packages, e.g. Microsoft.DesktopAppInstaller_8wekyb3d8bbwe
	assert.Equal(t, "8wekyb3d8bbwe", PublisherID("CN=Microsoft Corporation, O=Microsoft Corporation, L=Redmond, S=Washington, C=US"))
}

func writeMSIX(t *testing.T, manifest string) string {
	path := filepath.Join(t.TempDir(), "app.msix")
	f, err := os.Create(path)
	require.NoError(t, err)
	w := zip.NewWriter(f)
	if manifest != "" {
		entry, err := w.Create("AppxManifest.xml")
		require.NoError(t, err)
		_, err = entry.Write([]byte(manifest))
		require.NoError(t, err)
	}
	require.NoError(t, w.Close())
	require.NoError(t, f.Close())
	return path
}

func TestReadMSIXInfo(t *testing.T) {
	path := writeMSIX(t, `<?xml version="1.0" encoding="utf-8"?>
<Package xmlns="http://schemas.microsoft.com/appx/manifest/foundation/windows10">
  <Identity Name="ExampleCorp.Notes" Publisher="CN=Microsoft Corporation, O=Microsoft Corporation, L=Redmond, S=Washington, C=US" Version="2.5.0.0" ProcessorArchitecture="x64"/>
  <Properties>
    <DisplayName>Example Notes</DisplayName>
    <PublisherDisplayName>Example Corp</PublisherDisplayName>
  </Properties>
</Package>`)

	info, err := ReadMSIXInfo(path)
	require.NoError(t, err)
	assert.Equal(t, MSIXInfo{
		Name: "ExampleCorp.Notes", DisplayName: "Example Notes", Publisher: "Example Corp", Version: "2.5.0.0",
		Architecture: "x64", PackageFamilyName: "ExampleCorp.Notes_8wekyb3d8bbwe",
	}, info)

	// Names from the package's resources aren't readable here
	path = writeMSIX(t, `<Package><Identity Name="Other.App" Publisher="CN=Other Ltd, C=DE" Version="1.0.0.0"/>
<Properties><DisplayName>ms-resource:AppName</DisplayName><PublisherDisplayName>ms-resource:Publisher</PublisherDisplayName></Properties></Package>`)
	info, err = ReadMSIXInfo(path)
	require.NoError(t, err)
	assert.Equal(t, "Other.App", info.DisplayName)
	assert.Equal(t, "Other Ltd", info.Publisher)
	assert.Equal(t, "neutral", info.Architecture)

	_, err = ReadMSIXInfo(writeMSIX(t, ""))
	assert.Error(t, err, "no manifest, e.g. a bundle")
	_, err = ReadMSIXInfo(filepath.Join("testdata", "test.exe"))
	assert.Error(t, err, "not a zip file")
}
