package handlers

import (
	"testing"

	"github.com/open-uem/openuem-console/internal/wingetrepo"
	"github.com/stretchr/testify/assert"
)

func TestDefaultPackageIdentifier(t *testing.T) {
	assert.Equal(t, "ExampleCorp.TestApp", defaultPackageIdentifier("Example Corp.", "Test App"))
	assert.Equal(t, "MullerSohne.TestAppUnicode", defaultPackageIdentifier("Müller & Söhne", "Test App ünicode"))
	assert.Equal(t, "Mozilla.FirefoxESRX64DeDE", defaultPackageIdentifier("Mozilla", "Firefox ESR (x64 de-DE)"))

	p := wingetrepo.Package{PackageIdentifier: defaultPackageIdentifier("7-Zip", "7-Zip 24.08 (x64)"), PackageVersion: "24.08", PackageName: "7-Zip", Publisher: "7-Zip", Architecture: "neutral", InstallerFile: "7z.msi"}
	assert.Equal(t, "7Zip.7Zip2408X64", p.PackageIdentifier)
	assert.NoError(t, p.Validate(), "derived identifiers are valid")
}
