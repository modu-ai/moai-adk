//go:build !windows

package auth

import (
	"errors"
	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"os"
	"path/filepath"
)

type platformStore struct{}

func openPlatformStore(dir string) (*Store, error) {
	if !filepath.IsAbs(dir) {
		return nil, ErrAuthState
	}
	dir = filepath.Clean(dir)
	for p := dir; ; p = filepath.Dir(p) {
		if info, e := os.Lstat(p); e == nil && info.Mode()&os.ModeSymlink != 0 {
			return nil, ErrAuthState
		}
		if filepath.Dir(p) == p {
			break
		}
	}
	if e := os.Mkdir(dir, 0700); e != nil && !errors.Is(e, os.ErrExist) {
		return nil, ErrAuthState
	}
	info, e := os.Lstat(dir)
	if e != nil || !info.IsDir() || info.Mode().Perm()&0077 != 0 {
		return nil, ErrAuthState
	}
	root, e := os.OpenRoot(dir)
	if e != nil {
		return nil, ErrAuthState
	}
	return &Store{dir: dir, root: root}, nil
}
func (s *Store) writePlatform(raw []byte) error {
	f, e := os.CreateTemp(s.dir, ".state-")
	if e != nil {
		return ErrAuthState
	}
	name := f.Name()
	defer os.Remove(name)
	if e = f.Chmod(0600); e == nil {
		_, e = f.Write(raw)
	}
	if e == nil {
		e = f.Sync()
	}
	closeErr := f.Close()
	if e != nil || closeErr != nil {
		return ErrAuthState
	}
	if e = atomicfile.Replace(name, filepath.Join(s.dir, "state.json")); e != nil {
		return ErrAuthState
	}
	if e = syncDirectory(s.dir); e != nil {
		return ErrAuthState
	}
	return nil
}

func (s *Store) validatePlatformRoot() error { return nil }
func (s *Store) closePlatform() error        { return nil }
func privateRegular(path string, info os.FileInfo) bool {
	return info.Mode().IsRegular() && info.Mode().Perm()&0077 == 0
}
func privateDirectory(path string) error {
	info, e := os.Lstat(path)
	if e != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm()&0077 != 0 {
		return ErrAuthState
	}
	return nil
}
func makePrivateDirectory(path string) error {
	if e := os.Mkdir(path, 0700); e != nil && !os.IsExist(e) {
		return ErrAuthState
	}
	return privateDirectory(path)
}
func privateScratch(dir, prefix string) (string, error) { return os.MkdirTemp(dir, prefix) }
func seedPrivateFile(path string, raw []byte) error     { return os.WriteFile(path, raw, 0600) }
func (s *Store) openLock(name string) (*os.File, error) {
	return s.root.OpenFile(name, os.O_CREATE|os.O_RDWR, 0600)
}
func validateOpenedPrivateFile(f *os.File) error { return nil }

func holdPrivateDirectory(path string) (func(), error) { return func() {}, nil }

func privatePathInfo(path string, directory bool) (os.FileInfo, error) { return os.Lstat(path) }

func (s *Store) stateInfo() (os.FileInfo, error) { return s.root.Lstat("state.json") }
