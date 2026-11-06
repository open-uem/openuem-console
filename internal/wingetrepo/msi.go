package wingetrepo

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/richardlehane/mscfb"
	"golang.org/x/text/encoding/charmap"
)

// maxMSIStreamSize limits the size of the MSI streams read to get the properties
const maxMSIStreamSize = 64 << 20

// ErrNotMSI is returned for files that aren't MSI packages
var ErrNotMSI = errors.New("the file is not an MSI package")

// ReadMSIProperties returns the Property table of an MSI package, e.g. ProductName,
// Manufacturer, ProductVersion and ProductCode.
//
// An MSI package is a compound file whose stream names are encoded. The Property table is
// stored in the !Property stream as two columns of references to the string pool, kept in
// the !_StringPool (lengths) and !_StringData (bytes) streams.
func ReadMSIProperties(r io.ReaderAt) (map[string]string, error) {
	doc, err := mscfb.New(r)
	if err != nil {
		return nil, ErrNotMSI
	}

	streams := map[string][]byte{}
	for entry, err := doc.Next(); err == nil; entry, err = doc.Next() {
		name := decodeMSIStreamName(entry.Name)
		if name != "!_StringPool" && name != "!_StringData" && name != "!Property" {
			continue
		}
		if entry.Size > maxMSIStreamSize {
			return nil, fmt.Errorf("the %s stream of the MSI package is too big", name)
		}
		data := make([]byte, entry.Size)
		if _, err := io.ReadFull(entry, data); err != nil {
			return nil, err
		}
		streams[name] = data
	}

	pool, refSize, err := readMSIStringPool(streams["!_StringPool"], streams["!_StringData"])
	if err != nil {
		return nil, err
	}

	table := streams["!Property"]
	rowSize := 2 * refSize
	if len(table) == 0 || len(table)%rowSize != 0 {
		return nil, errors.New("the MSI package has no valid Property table")
	}

	// Tables are stored column by column
	rows := len(table) / rowSize
	properties := map[string]string{}
	for row := 0; row < rows; row++ {
		key := pool.get(readMSIStringRef(table[row*refSize:], refSize))
		value := pool.get(readMSIStringRef(table[(rows+row)*refSize:], refSize))
		if key != "" {
			properties[key] = value
		}
	}
	return properties, nil
}

// decodeMSIStreamName decodes the names of the streams of an MSI package. Two characters
// of [0-9A-Za-z._] are packed into a single rune and tables are prefixed with !.
func decodeMSIStreamName(name string) string {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz._"
	decoded := []rune{}
	for _, r := range name {
		switch {
		case r >= 0x3800 && r < 0x4800:
			r -= 0x3800
			decoded = append(decoded, rune(charset[r&0x3f]), rune(charset[(r>>6)&0x3f]))
		case r >= 0x4800 && r < 0x4840:
			decoded = append(decoded, rune(charset[r-0x4800]))
		case r == 0x4840:
			decoded = append(decoded, '!')
		default:
			decoded = append(decoded, r)
		}
	}
	return string(decoded)
}

type msiStringPool struct {
	strings  map[int]string
	codepage uint32
}

func (p msiStringPool) get(id int) string {
	s := p.strings[id]
	if p.codepage == 65001 || utf8.ValidString(s) {
		return s
	}
	decoded, err := charmap.Windows1252.NewDecoder().String(s)
	if err != nil {
		return s
	}
	return decoded
}

// readMSIStringPool reads the strings of the package. The pool starts with the codepage
// and a flag for 3 byte string references, followed by (length, reference count) pairs.
// Strings longer than 64 KiB use an empty entry followed by the length in two words.
func readMSIStringPool(pool, data []byte) (msiStringPool, int, error) {
	if len(pool) < 4 || len(pool)%4 != 0 {
		return msiStringPool{}, 0, errors.New("the MSI package has no valid string pool")
	}

	words := func(i int) (uint16, uint16) {
		return binary.LittleEndian.Uint16(pool[i*4:]), binary.LittleEndian.Uint16(pool[i*4+2:])
	}

	low, high := words(0)
	refSize := 2
	if high&0x8000 != 0 {
		refSize = 3
	}
	p := msiStringPool{strings: map[int]string{}, codepage: uint32(low) | uint32(high&^0x8000)<<16}

	count := len(pool) / 4
	offset := 0
	for i, id := 1, 1; i < count; id++ {
		length, refs := words(i)
		size := int(length)
		switch {
		case length == 0 && refs == 0:
			i++
			continue
		case length == 0:
			if i+1 >= count {
				return msiStringPool{}, 0, errors.New("the MSI string pool is truncated")
			}
			lengthLow, lengthHigh := words(i + 1)
			size = int(lengthHigh)<<16 | int(lengthLow)
			i += 2
		default:
			i++
		}

		if offset+size > len(data) {
			return msiStringPool{}, 0, errors.New("the MSI string data is truncated")
		}
		p.strings[id] = string(data[offset : offset+size])
		offset += size
	}
	return p, refSize, nil
}

func readMSIStringRef(b []byte, size int) int {
	if size == 3 {
		return int(b[0]) | int(b[1])<<8 | int(b[2])<<16
	}
	return int(binary.LittleEndian.Uint16(b))
}
