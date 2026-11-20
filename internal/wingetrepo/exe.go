package wingetrepo

import (
	"bytes"
	"debug/pe"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf16"
)

// Installer types of winget's manifests that the library supports
const (
	InstallerTypeMSI      = "msi"
	InstallerTypeEXE      = "exe"
	InstallerTypeInno     = "inno"
	InstallerTypeNullsoft = "nullsoft"
	InstallerTypeBurn     = "burn"
	InstallerTypeMSIX     = "msix"
)

// ExeInstallerTypes are the installer types of .exe files. winget knows the silent switches
// of all of them except the generic exe.
var ExeInstallerTypes = []string{InstallerTypeEXE, InstallerTypeInno, InstallerTypeNullsoft, InstallerTypeBurn}

var errNotPE = errors.New("not a Windows executable")

// exeMarkers identify installers by data they carry after the executable: Inno Setup's
// setup data header and NSIS's first header signature
var exeMarkers = []struct {
	marker        []byte
	installerType string
}{
	{[]byte("Inno Setup Setup Data"), InstallerTypeInno},
	{[]byte("NullsoftInst"), InstallerTypeNullsoft},
}

// ExeInfo is what can be read from an .exe installer
type ExeInfo struct {
	// InstallerType is inno, nullsoft, burn or exe when the technology is unknown
	InstallerType string
	ProductName   string
	CompanyName   string
	// ProductVersion is the product version of the version resource, or its file version
	ProductVersion string
}

// ReadExeInfo reads the version resource of an .exe file and detects the installer
// technology: WiX Burn bundles have a .wixburn section, Inno Setup and NSIS installers
// carry their own signatures.
func ReadExeInfo(path string) (ExeInfo, error) {
	f, err := pe.Open(path)
	if err != nil {
		return ExeInfo{}, errNotPE
	}
	defer f.Close()

	info := ExeInfo{InstallerType: InstallerTypeEXE}
	if values := readVersionStrings(f); values != nil {
		info.ProductName = firstNonEmpty(values["ProductName"], values["FileDescription"])
		info.CompanyName = values["CompanyName"]
		info.ProductVersion = firstNonEmpty(normalizeExeVersion(values["ProductVersion"]), normalizeExeVersion(values["FileVersion"]))
	}

	if f.Section(".wixburn") != nil {
		info.InstallerType = InstallerTypeBurn
		return info, nil
	}

	installerType, err := scanExeMarkers(path)
	if err != nil {
		return ExeInfo{}, err
	}
	if installerType != "" {
		info.InstallerType = installerType
	}
	return info, nil
}

// scanExeMarkers looks for the signatures of installer technologies in the whole file,
// reading it in blocks so large installers aren't loaded in memory
func scanExeMarkers(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	overlap := 0
	for _, m := range exeMarkers {
		overlap = max(overlap, len(m.marker)-1)
	}

	buf := make([]byte, 1<<20+overlap)
	kept := 0
	for {
		n, err := io.ReadFull(f, buf[kept:])
		data := buf[:kept+n]
		for _, m := range exeMarkers {
			if bytes.Contains(data, m.marker) {
				return m.installerType, nil
			}
		}
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		// Keep the end of the block, a signature may continue in the next one
		kept = copy(buf, data[len(data)-overlap:])
	}
}

// normalizeExeVersion turns versions like "1, 2, 3, 4" or " 24.08 " into valid package
// versions. It returns an empty string if the result isn't one.
func normalizeExeVersion(v string) string {
	v = strings.ReplaceAll(v, ",", ".")
	v = strings.ReplaceAll(v, " ", "")
	if !packageVersionPattern.MatchString(v) {
		return ""
	}
	return v
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v = strings.TrimSpace(v); v != "" {
			return v
		}
	}
	return ""
}

// Resource type of version information
const rtVersion = 16

// readVersionStrings returns the strings of the first string table of the version
// resource, or nil if there's none
func readVersionStrings(f *pe.File) map[string]string {
	data, err := versionResource(f)
	if err != nil {
		return nil
	}
	root, err := parseVersionBlock(data, 0)
	if err != nil {
		return nil
	}

	for _, fileInfo := range root.children {
		if fileInfo.key != "StringFileInfo" {
			continue
		}
		for _, table := range fileInfo.children {
			values := map[string]string{}
			for _, s := range table.children {
				values[s.key] = s.text
			}
			if len(values) > 0 {
				return values
			}
		}
	}
	return nil
}

// versionResource returns the data of the first version resource of a PE file
func versionResource(f *pe.File) ([]byte, error) {
	var dir pe.DataDirectory
	switch h := f.OptionalHeader.(type) {
	case *pe.OptionalHeader32:
		if len(h.DataDirectory) <= pe.IMAGE_DIRECTORY_ENTRY_RESOURCE {
			return nil, errNotPE
		}
		dir = h.DataDirectory[pe.IMAGE_DIRECTORY_ENTRY_RESOURCE]
	case *pe.OptionalHeader64:
		if len(h.DataDirectory) <= pe.IMAGE_DIRECTORY_ENTRY_RESOURCE {
			return nil, errNotPE
		}
		dir = h.DataDirectory[pe.IMAGE_DIRECTORY_ENTRY_RESOURCE]
	default:
		return nil, errNotPE
	}
	if dir.VirtualAddress == 0 {
		return nil, errors.New("no resources")
	}

	// The section that holds the resource directory
	var section *pe.Section
	for _, s := range f.Sections {
		if dir.VirtualAddress >= s.VirtualAddress && dir.VirtualAddress < s.VirtualAddress+max(s.VirtualSize, s.Size) {
			section = s
			break
		}
	}
	if section == nil {
		return nil, errors.New("resource section not found")
	}
	rsrc, err := section.Data()
	if err != nil {
		return nil, err
	}
	base := dir.VirtualAddress - section.VirtualAddress

	// Three levels: type, name, language. The first name and language are used.
	offset, err := resourceEntry(rsrc, base, base, func(id uint32) bool { return id == rtVersion })
	if err != nil {
		return nil, err
	}
	for range 2 {
		if offset, err = resourceEntry(rsrc, base, offset, func(uint32) bool { return true }); err != nil {
			return nil, err
		}
	}

	// A data entry: the data's RVA and size
	if int(offset)+8 > len(rsrc) {
		return nil, errors.New("invalid resource data entry")
	}
	rva := binary.LittleEndian.Uint32(rsrc[offset:])
	size := binary.LittleEndian.Uint32(rsrc[offset+4:])
	start := int64(rva) - int64(section.VirtualAddress)
	if start < 0 || start+int64(size) > int64(len(rsrc)) {
		return nil, errors.New("version resource outside the resource section")
	}
	return rsrc[start : start+int64(size)], nil
}

// resourceEntry finds, in the resource directory at offset, the first entry whose id is
// accepted, and returns the offset of what it points to. Offsets are relative to base,
// the start of the resource directory.
func resourceEntry(rsrc []byte, base, offset uint32, accept func(id uint32) bool) (uint32, error) {
	if int(offset)+16 > len(rsrc) {
		return 0, errors.New("invalid resource directory")
	}
	named := binary.LittleEndian.Uint16(rsrc[offset+12:])
	ids := binary.LittleEndian.Uint16(rsrc[offset+14:])
	for i := uint32(0); i < uint32(named)+uint32(ids); i++ {
		entry := offset + 16 + i*8
		if int(entry)+8 > len(rsrc) {
			break
		}
		name := binary.LittleEndian.Uint32(rsrc[entry:])
		target := binary.LittleEndian.Uint32(rsrc[entry+4:])
		// Named entries have the high bit set and never match an id
		if name&0x80000000 != 0 || !accept(name) {
			continue
		}
		return base + target&0x7fffffff, nil
	}
	return 0, errors.New("resource not found")
}

// versionBlock is a block of a VS_VERSIONINFO structure: a key, a value, and child blocks
type versionBlock struct {
	key      string
	text     string
	children []versionBlock
}

// parseVersionBlock parses the block at offset: its length, value length and type, its
// key, then its value and its children, each aligned on 4 bytes
func parseVersionBlock(data []byte, offset int) (versionBlock, error) {
	if offset+6 > len(data) {
		return versionBlock{}, errors.New("truncated version block")
	}
	length := int(binary.LittleEndian.Uint16(data[offset:]))
	valueLength := int(binary.LittleEndian.Uint16(data[offset+2:]))
	isText := binary.LittleEndian.Uint16(data[offset+4:]) == 1
	end := offset + length
	if length < 6 || end > len(data) {
		return versionBlock{}, fmt.Errorf("invalid version block length %d", length)
	}

	b := versionBlock{}
	var pos int
	b.key, pos = readUTF16String(data[:end], offset+6)
	pos = align4(pos)

	if valueLength > 0 {
		if isText {
			b.text, _ = readUTF16String(data[:end], pos)
			// The length is in characters, but some tools write bytes
			pos += min(valueLength*2, end-pos)
		} else {
			pos += min(valueLength, end-pos)
		}
	}
	pos = align4(pos)

	for pos < end {
		child, err := parseVersionBlock(data, pos)
		if err != nil {
			return b, err
		}
		b.children = append(b.children, child)
		pos = align4(pos + int(binary.LittleEndian.Uint16(data[pos:])))
	}
	return b, nil
}

// readUTF16String reads a NUL-terminated UTF-16 string, returning it and the offset after it
func readUTF16String(data []byte, offset int) (string, int) {
	chars := []uint16{}
	for offset+2 <= len(data) {
		c := binary.LittleEndian.Uint16(data[offset:])
		offset += 2
		if c == 0 {
			break
		}
		chars = append(chars, c)
	}
	return strings.TrimSpace(string(utf16.Decode(chars))), offset
}

func align4(n int) int { return (n + 3) &^ 3 }
