package wingetrepo

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"io"
	"strings"
	"unicode/utf16"
)

// MSIXInfo is what can be read from an MSIX package's manifest
type MSIXInfo struct {
	Name         string
	DisplayName  string
	Publisher    string
	Version      string
	Architecture string
	// PackageFamilyName is how winget recognizes the installed package
	PackageFamilyName string
}

type appxManifest struct {
	Identity struct {
		Name                  string `xml:"Name,attr"`
		Publisher             string `xml:"Publisher,attr"`
		Version               string `xml:"Version,attr"`
		ProcessorArchitecture string `xml:"ProcessorArchitecture,attr"`
	} `xml:"Identity"`
	Properties struct {
		DisplayName          string `xml:"DisplayName"`
		PublisherDisplayName string `xml:"PublisherDisplayName"`
	} `xml:"Properties"`
}

// maxManifestSize limits how much of AppxManifest.xml is read
const maxManifestSize = 1 << 20

// ReadMSIXInfo reads AppxManifest.xml of an MSIX package
func ReadMSIXInfo(path string) (MSIXInfo, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return MSIXInfo{}, errors.New("not an MSIX package")
	}
	defer r.Close()

	var file *zip.File
	for _, f := range r.File {
		if f.Name == "AppxManifest.xml" {
			file = f
			break
		}
	}
	if file == nil {
		return MSIXInfo{}, errors.New("AppxManifest.xml not found, MSIX bundles aren't supported")
	}

	rc, err := file.Open()
	if err != nil {
		return MSIXInfo{}, err
	}
	defer rc.Close()

	m := appxManifest{}
	if err := xml.NewDecoder(io.LimitReader(rc, maxManifestSize)).Decode(&m); err != nil {
		return MSIXInfo{}, err
	}
	if m.Identity.Name == "" || m.Identity.Publisher == "" || m.Identity.Version == "" {
		return MSIXInfo{}, errors.New("the package identity is incomplete")
	}

	info := MSIXInfo{
		Name:              m.Identity.Name,
		DisplayName:       m.Identity.Name,
		Publisher:         publisherName(m.Identity.Publisher),
		Version:           m.Identity.Version,
		Architecture:      msixArchitecture(m.Identity.ProcessorArchitecture),
		PackageFamilyName: m.Identity.Name + "_" + PublisherID(m.Identity.Publisher),
	}
	// Names starting with ms-resource: are looked up in the package's resources
	if d := strings.TrimSpace(m.Properties.DisplayName); d != "" && !strings.HasPrefix(d, "ms-resource:") {
		info.DisplayName = d
	}
	if d := strings.TrimSpace(m.Properties.PublisherDisplayName); d != "" && !strings.HasPrefix(d, "ms-resource:") {
		info.Publisher = d
	}
	return info, nil
}

// publisherName returns the common name of a publisher's distinguished name, e.g.
// "Example Corp" for "CN=Example Corp, O=Example Corp, C=US"
func publisherName(dn string) string {
	for _, part := range strings.Split(dn, ",") {
		if k, v, ok := strings.Cut(strings.TrimSpace(part), "="); ok && strings.EqualFold(k, "CN") {
			return strings.Trim(strings.TrimSpace(v), `"`)
		}
	}
	return dn
}

func msixArchitecture(a string) string {
	switch strings.ToLower(a) {
	case "x64", "x86", "arm64":
		return strings.ToLower(a)
	}
	return "neutral"
}

// PublisherID is the publisher part of a package family name: the first 8 bytes of the
// SHA-256 of the UTF-16 publisher, followed by a 0 bit, in Crockford's base32
func PublisherID(publisher string) string {
	const alphabet = "0123456789abcdefghjkmnpqrstvwxyz"

	units := utf16.Encode([]rune(publisher))
	data := make([]byte, len(units)*2)
	for i, u := range units {
		binary.LittleEndian.PutUint16(data[i*2:], u)
	}
	sum := sha256.Sum256(data)
	bits := binary.BigEndian.Uint64(sum[:8])

	id := make([]byte, 13)
	for i := 0; i < 12; i++ {
		id[i] = alphabet[(bits>>(59-5*i))&31]
	}
	// The last character has the last 4 bits and the 0 bit
	id[12] = alphabet[(bits&15)<<1]
	return string(id)
}
