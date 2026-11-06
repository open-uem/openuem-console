// Package taskfiles stores the files uploaded for profile tasks, e.g. MSI packages.
//
// Files are stored per profile as <dir>/<profile id>/<sha256>/<file name> and are served
// by the auth server under /agent-files/<profile id>/<sha256>/<file name> only to agents
// that the profile applies to.
package taskfiles

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// URLPrefix is the path where the auth server serves the uploaded files to agents
const URLPrefix = "/agent-files/"

var (
	ErrInvalidFile = errors.New("invalid file")
	ErrNotFound    = errors.New("file not found")

	hashPattern     = regexp.MustCompile(`^[0-9a-f]{64}$`)
	unsafeNameChars = regexp.MustCompile(`[^A-Za-z0-9._+-]+`)
)

// File identifies an uploaded file
type File struct {
	ProfileID int
	Hash      string
	Name      string
}

// URLPath is the path of the file in the auth server
func (f File) URLPath() string {
	return fmt.Sprintf("%s%d/%s/%s", URLPrefix, f.ProfileID, f.Hash, url.PathEscape(f.Name))
}

// ParseURL returns the file a task path refers to, if it's a file uploaded to the console
func ParseURL(rawURL string) (File, bool) {
	u, err := url.Parse(rawURL)
	if err != nil || !strings.HasPrefix(u.Path, URLPrefix) {
		return File{}, false
	}

	parts := strings.Split(strings.TrimPrefix(u.Path, URLPrefix), "/")
	if len(parts) != 3 {
		return File{}, false
	}

	f, err := NewFile(parts[0], parts[1], parts[2])
	if err != nil {
		return File{}, false
	}
	return f, true
}

// NewFile validates the parts of a file reference, as found in a URL
func NewFile(profileID, hash, name string) (File, error) {
	id, err := strconv.Atoi(profileID)
	if err != nil || id <= 0 || !hashPattern.MatchString(hash) || name == "" || SanitizeName(name) != name {
		return File{}, ErrInvalidFile
	}
	return File{ProfileID: id, Hash: hash, Name: name}, nil
}

// SanitizeName keeps only the base name of an uploaded file and replaces the characters
// that aren't safe in a path or a command line
func SanitizeName(name string) string {
	name = path.Base(strings.ReplaceAll(name, "\\", "/"))
	name = unsafeNameChars.ReplaceAllString(name, "_")
	name = strings.Trim(name, "._")
	if len(name) > 200 {
		name = name[len(name)-200:]
	}
	return name
}

// Store keeps the uploaded files in a directory
type Store struct {
	dir string
	// mu prevents the cleanup from removing directories while a file is saved
	mu sync.Mutex
}

// New creates the store, creating its directory if needed
func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return &Store{dir: dir}, nil
}

// Save stores a file uploaded for a profile. Files with the same content and name are
// only stored once.
func (s *Store) Save(profileID int, name string, src io.Reader) (File, error) {
	name = SanitizeName(name)
	if profileID <= 0 || name == "" {
		return File{}, ErrInvalidFile
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	profileDir := filepath.Join(s.dir, strconv.Itoa(profileID))
	if err := os.MkdirAll(profileDir, 0o700); err != nil {
		return File{}, err
	}

	tmp, err := os.CreateTemp(profileDir, ".upload-*")
	if err != nil {
		return File{}, err
	}
	defer os.Remove(tmp.Name())

	hash := sha256.New()
	if _, err := io.Copy(io.MultiWriter(tmp, hash), src); err != nil {
		tmp.Close()
		return File{}, err
	}
	if err := tmp.Close(); err != nil {
		return File{}, err
	}

	f := File{ProfileID: profileID, Hash: hex.EncodeToString(hash.Sum(nil)), Name: name}
	dst := s.path(f)
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return File{}, err
	}
	if err := os.Rename(tmp.Name(), dst); err != nil {
		return File{}, err
	}
	return f, nil
}

// Path returns the location of a stored file
func (s *Store) Path(f File) (string, error) {
	if _, err := NewFile(strconv.Itoa(f.ProfileID), f.Hash, f.Name); err != nil {
		return "", err
	}
	p := s.path(f)
	if _, err := os.Stat(p); err != nil {
		return "", ErrNotFound
	}
	return p, nil
}

func (s *Store) path(f File) string {
	return filepath.Join(s.dir, strconv.Itoa(f.ProfileID), f.Hash, f.Name)
}

// RemoveUnused deletes the files that no task uses anymore. Files uploaded during the
// grace period are kept, as the task using them may not have been saved yet.
func (s *Store) RemoveUnused(used map[File]bool, grace time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	profiles, err := os.ReadDir(s.dir)
	if err != nil {
		return err
	}

	for _, profile := range profiles {
		profileID, err := strconv.Atoi(profile.Name())
		if err != nil || !profile.IsDir() {
			continue
		}
		profileDir := filepath.Join(s.dir, profile.Name())

		hashes, err := os.ReadDir(profileDir)
		if err != nil {
			return err
		}
		for _, hash := range hashes {
			hashDir := filepath.Join(profileDir, hash.Name())

			// Temporary files of uploads that were interrupted
			if !hash.IsDir() && strings.HasPrefix(hash.Name(), ".upload-") {
				if info, err := hash.Info(); err == nil && time.Since(info.ModTime()) >= grace {
					_ = os.Remove(hashDir)
				}
				continue
			}

			files, err := os.ReadDir(hashDir)
			if err != nil || !hash.IsDir() {
				continue
			}
			for _, file := range files {
				info, err := file.Info()
				if err != nil || time.Since(info.ModTime()) < grace {
					continue
				}
				if !used[File{ProfileID: profileID, Hash: hash.Name(), Name: file.Name()}] {
					if err := os.Remove(filepath.Join(hashDir, file.Name())); err != nil {
						return err
					}
				}
			}
			// Only removed if it's empty
			_ = os.Remove(hashDir)
		}
		_ = os.Remove(profileDir)
	}
	return nil
}
