package wingetrepo

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testdata/test.msi was created with msitools:
//
//	msibuild test.msi -s "Test App" "Example Corp" ";1033" "{11111111-2222-3333-4444-555555555555}"
//	msibuild test.msi -q "CREATE TABLE `Property` (...)" and one INSERT per property
func TestReadMSIProperties(t *testing.T) {
	f, err := os.Open("testdata/test.msi")
	require.NoError(t, err)
	defer f.Close()

	properties, err := ReadMSIProperties(f)
	require.NoError(t, err)
	assert.Equal(t, "Test App ünicode", properties["ProductName"])
	assert.Equal(t, "Example Corp", properties["Manufacturer"])
	assert.Equal(t, "1.2.3", properties["ProductVersion"])
	assert.Equal(t, "{0D8B6A0E-1F2A-4C3B-9D4E-5F6A7B8C9D0E}", properties["ProductCode"])
}

func TestReadMSIPropertiesNotMSI(t *testing.T) {
	_, err := ReadMSIProperties(bytes.NewReader([]byte("MZ this is not an MSI package")))
	assert.ErrorIs(t, err, ErrNotMSI)
}

func TestDecodeMSIStreamName(t *testing.T) {
	// "!Property": table prefix, then pairs of characters packed into one rune
	encode := func(a, b byte) rune {
		const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz._"
		return rune(0x3800 + bytes.IndexByte([]byte(charset), a) + bytes.IndexByte([]byte(charset), b)<<6)
	}
	name := string([]rune{0x4840, encode('P', 'r'), encode('o', 'p'), encode('e', 'r'), encode('t', 'y')})
	assert.Equal(t, "!Property", decodeMSIStreamName(name))
}

func TestReadMSIStringPoolLongString(t *testing.T) {
	// codepage 65001, one short string, one string longer than 64 KiB
	long := bytes.Repeat([]byte("x"), 70000)
	pool := []byte{0xe9, 0xfd, 0x00, 0x00, // codepage 65001, 2 byte references
		3, 0, 1, 0, // "abc"
		0, 0, 1, 0, // long string marker
		byte(70000 & 0xff), byte(70000 >> 8 & 0xff), byte(70000 >> 16), 0}
	data := append([]byte("abc"), long...)

	p, refSize, err := readMSIStringPool(pool, data)
	require.NoError(t, err)
	assert.Equal(t, 2, refSize)
	assert.Equal(t, "abc", p.get(1))
	assert.Len(t, p.get(2), 70000)
}
