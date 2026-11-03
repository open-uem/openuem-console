package taskfiles

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeName(t *testing.T) {
	assert.Equal(t, "setup.msi", SanitizeName("setup.msi"))
	assert.Equal(t, "setup.msi", SanitizeName(`C:\Users\me\Downloads\setup.msi`))
	assert.Equal(t, "setup.msi", SanitizeName("../../setup.msi"))
	assert.Equal(t, "My_App_1.2_x64_.msi", SanitizeName("My App 1.2 (x64).msi"))
	assert.Equal(t, "a_b.msi", SanitizeName("a'\"b.msi"))
	assert.Equal(t, "", SanitizeName(".."))
}

func TestSaveAndPath(t *testing.T) {
	store, err := New(t.TempDir())
	require.NoError(t, err)

	f, err := store.Save(3, "../evil name.msi", strings.NewReader("content"))
	require.NoError(t, err)
	assert.Equal(t, 3, f.ProfileID)
	assert.Equal(t, "evil_name.msi", f.Name)
	assert.Equal(t, "ed7002b439e9ac845f22357d822bac1444730fbdb6016d3ec9432297b9ec9f73", f.Hash)

	p, err := store.Path(f)
	require.NoError(t, err)
	data, err := os.ReadFile(p)
	require.NoError(t, err)
	assert.Equal(t, "content", string(data))

	// Another profile can't reach the file
	_, err = store.Path(File{ProfileID: 4, Hash: f.Hash, Name: f.Name})
	assert.ErrorIs(t, err, ErrNotFound)

	// Path traversal is rejected
	_, err = store.Path(File{ProfileID: 3, Hash: f.Hash, Name: "../../etc/passwd"})
	assert.ErrorIs(t, err, ErrInvalidFile)

	_, err = store.Save(0, "a.msi", strings.NewReader("x"))
	assert.ErrorIs(t, err, ErrInvalidFile)
}

func TestURLRoundTrip(t *testing.T) {
	f := File{ProfileID: 12, Hash: strings.Repeat("a", 64), Name: "App_1.0.msi"}
	parsed, ok := ParseURL("https://console.example.com:1324" + f.URLPath())
	assert.True(t, ok)
	assert.Equal(t, f, parsed)

	for _, u := range []string{
		"https://example.com/app.msi",
		"https://console.example.com:1324/agent-files/12/notahash/App.msi",
		"https://console.example.com:1324/agent-files/x/" + strings.Repeat("a", 64) + "/App.msi",
		"https://console.example.com:1324/agent-files/12/" + strings.Repeat("a", 64) + "/..%2F..%2Fpasswd",
		`\\server\share\app.msi`,
	} {
		_, ok := ParseURL(u)
		assert.False(t, ok, u)
	}
}

func TestRemoveUnused(t *testing.T) {
	store, err := New(t.TempDir())
	require.NoError(t, err)

	used, err := store.Save(1, "used.msi", strings.NewReader("used"))
	require.NoError(t, err)
	unused, err := store.Save(1, "unused.msi", strings.NewReader("unused"))
	require.NoError(t, err)

	// Recently uploaded files are kept during the grace period
	require.NoError(t, store.RemoveUnused(map[File]bool{used: true}, time.Hour))
	_, err = store.Path(unused)
	assert.NoError(t, err)

	// Temporary files of interrupted uploads are removed too
	stale := filepath.Join(store.dir, "1", ".upload-123")
	require.NoError(t, os.WriteFile(stale, []byte("partial"), 0o600))

	require.NoError(t, store.RemoveUnused(map[File]bool{used: true}, 0))
	_, err = store.Path(unused)
	assert.ErrorIs(t, err, ErrNotFound)
	_, err = store.Path(used)
	assert.NoError(t, err)
	_, err = os.Stat(stale)
	assert.True(t, os.IsNotExist(err))
}
