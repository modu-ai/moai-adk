//go:build windows

package auth

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"golang.org/x/sys/windows"
)

// A failed post-replacement verification does not imply rollback. This handle
// becomes unusable; reopening requires all ownership and content checks again.
var ErrAuthCommitUncertain = errors.New("authentication state replaced but final verification failed")

type platformStore struct {
	pins      []*os.File
	identity  os.FileInfo
	uncertain atomic.Bool
}

func openPlatformStore(dir string) (*Store, error) {
	if !filepath.IsAbs(dir) {
		return nil, ErrAuthState
	}
	dir = filepath.Clean(dir)
	// Pin each existing ancestor without FILE_SHARE_DELETE. Reparse points are
	// rejected by handle, so a rename cannot exchange any path component later.
	var paths []string
	for p := filepath.Dir(dir); ; p = filepath.Dir(p) {
		paths = append(paths, p)
		if filepath.Dir(p) == p {
			break
		}
	}
	ps := &platformStore{}
	fail := func() (*Store, error) {
		for _, f := range ps.pins {
			f.Close()
		}
		return nil, ErrAuthState
	}
	for i := len(paths) - 1; i >= 0; i-- {
		f, e := pinWindowsDirectory(paths[i], false)
		if e != nil {
			return fail()
		}
		ps.pins = append(ps.pins, f)
	}
	if _, e := os.Lstat(dir); os.IsNotExist(e) {
		if createWindowsPrivateDirectory(dir) != nil {
			return fail()
		}
	} else if e != nil {
		return fail()
	}
	f, e := pinWindowsDirectory(dir, true)
	if e != nil {
		return fail()
	}
	ps.pins = append(ps.pins, f)
	ps.identity, e = f.Stat()
	if e != nil {
		return fail()
	}
	root, e := os.OpenRoot(dir)
	if e != nil {
		return fail()
	}
	s := &Store{dir: dir, root: root, platform: ps}
	if s.validatePlatformRoot() != nil {
		root.Close()
		return fail()
	}
	return s, nil
}

func pinWindowsDirectory(path string, private bool) (*os.File, error) {
	name, e := windows.UTF16PtrFromString(path)
	if e != nil {
		return nil, ErrAuthState
	}
	h, e := windows.CreateFile(name, windows.READ_CONTROL|windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING,
		windows.FILE_FLAG_OPEN_REPARSE_POINT|windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if e != nil {
		return nil, ErrAuthState
	}
	var info windows.ByHandleFileInformation
	if windows.GetFileInformationByHandle(h, &info) != nil || info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 || info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY == 0 || private && validateWindowsPrivateHandle(h, true) != nil {
		windows.CloseHandle(h)
		return nil, ErrAuthState
	}
	return os.NewFile(uintptr(h), path), nil
}

func (s *Store) validatePlatformRoot() error {
	if s.platform == nil || s.platform.uncertain.Load() {
		return ErrAuthState
	}
	info, e := privatePathInfo(s.dir, true)
	if e != nil || !os.SameFile(s.platform.identity, info) {
		return ErrAuthState
	}
	return validateWindowsPrivatePath(s.dir, true)
}
func (s *Store) closePlatform() error {
	var e error
	for _, f := range s.platform.pins {
		e = errors.Join(e, f.Close())
	}
	return e
}
func privateRegular(path string, info os.FileInfo) bool {
	return info.Mode().IsRegular() && validateWindowsPrivatePath(path, false) == nil
}
func privateDirectory(path string) error { return validateWindowsPrivatePath(path, true) }
func makePrivateDirectory(path string) error {
	if _, e := os.Lstat(path); os.IsNotExist(e) {
		return createWindowsPrivateDirectory(path)
	} else if e != nil {
		return ErrAuthState
	}
	return privateDirectory(path)
}
func privateScratch(dir, prefix string) (string, error) {
	path := filepath.Join(dir, prefix+rand.Text())
	if e := createWindowsPrivateDirectory(path); e != nil {
		return "", e
	}
	return path, nil
}
func seedPrivateFile(path string, raw []byte) error { return writeWindowsPrivateCandidate(path, raw) }
func validateOpenedPrivateFile(f *os.File) error {
	return validateWindowsPrivateHandle(windows.Handle(f.Fd()), false)
}
func (s *Store) openLock(name string) (*os.File, error) {
	path := filepath.Join(s.dir, name)
	if _, e := os.Lstat(path); os.IsNotExist(e) {
		if e = writeWindowsPrivateCandidate(path, nil); e != nil {
			// Another Store may have created the immutable lock inode concurrently.
			if validateWindowsPrivatePath(path, false) != nil {
				return nil, ErrAuthState
			}
		}
	} else if e != nil {
		return nil, ErrAuthState
	}
	ptr, e := windows.UTF16PtrFromString(path)
	if e != nil {
		return nil, ErrAuthState
	}
	h, e := windows.CreateFile(ptr, windows.GENERIC_READ|windows.GENERIC_WRITE|windows.READ_CONTROL,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if e != nil {
		return nil, ErrAuthState
	}
	f := os.NewFile(uintptr(h), path)
	if validateOpenedPrivateFile(f) != nil || s.validatePlatformRoot() != nil {
		f.Close()
		return nil, ErrAuthState
	}
	return f, nil
}

// Windows AUTH alone uses same-directory MoveFileEx write-through replacement.
// No cross-volume copy fallback and no POSIX directory-fsync claim is made.
func (s *Store) writePlatform(raw []byte) error {
	if s.validatePlatformRoot() != nil {
		return ErrAuthState
	}
	name := filepath.Join(s.dir, ".state-"+rand.Text())
	defer os.Remove(name)
	if writeWindowsPrivateCandidate(name, raw) != nil {
		return ErrAuthState
	}
	candidate, e := privatePathInfo(name, false)
	if e != nil {
		return ErrAuthState
	}
	if s.validatePlatformRoot() != nil {
		return ErrAuthState
	}
	target := filepath.Join(s.dir, "state.json")
	// Validation precedes the API; any failure after a successful API call is
	// explicitly uncertain and cannot publish or send this candidate.
	if _, e = os.Lstat(target); e == nil {
		if validateWindowsPrivatePath(target, false) != nil {
			return ErrAuthState
		}
	} else if !os.IsNotExist(e) {
		return ErrAuthState
	}
	from, e := windows.UTF16PtrFromString(name)
	if e != nil {
		return ErrAuthState
	}
	to, e := windows.UTF16PtrFromString(target)
	if e != nil {
		return ErrAuthState
	}
	if windows.MoveFileEx(from, to, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH) != nil {
		// An API error is not evidence of rollback. Conservatively require a fresh
		// independently verified Store before any further credential use.
		s.platform.uncertain.Store(true)
		return errors.Join(ErrAuthState, ErrAuthCommitUncertain)
	}
	return s.finishWindowsCommit(target, candidate, raw)
}
func (s *Store) finishWindowsCommit(target string, candidate os.FileInfo, raw []byte) error {
	if e := s.verifyWindowsCommit(target, candidate, raw); e != nil {
		s.platform.uncertain.Store(true)
		return errors.Join(ErrAuthState, ErrAuthCommitUncertain)
	}
	return nil
}
func (s *Store) verifyWindowsCommit(path string, candidate os.FileInfo, raw []byte) error {
	if s.validatePlatformRoot() != nil {
		return ErrAuthState
	}
	f, e := os.Open(path)
	if e != nil {
		return ErrAuthState
	}
	defer f.Close()
	if validateOpenedPrivateFile(f) != nil {
		return ErrAuthState
	}
	info, e := f.Stat()
	if e != nil || !os.SameFile(candidate, info) {
		return ErrAuthState
	}
	actual, e := io.ReadAll(io.LimitReader(f, 1<<20+1))
	if e != nil || !bytes.Equal(actual, raw) {
		return ErrAuthState
	}
	final, e := privatePathInfo(path, false)
	if e != nil || !os.SameFile(info, final) || validateWindowsPrivatePath(path, false) != nil {
		return ErrAuthState
	}
	return s.validatePlatformRoot()
}

func holdPrivateDirectory(path string) (func(), error) {
	f, e := pinWindowsDirectory(path, true)
	if e != nil {
		return nil, e
	}
	var once sync.Once
	return func() { once.Do(func() { f.Close() }) }, nil
}

// privatePathInfo captures identity eagerly from a validated handle. Windows
// path-based os.Stat/FileInfo defers file-ID lookup until SameFile, which is too
// late after rename and can conflict with a directory's pinned sharing mode.
func privatePathInfo(path string, directory bool) (os.FileInfo, error) {
	ptr, e := windows.UTF16PtrFromString(path)
	if e != nil {
		return nil, ErrAuthState
	}
	h, e := windows.CreateFile(ptr, windows.READ_CONTROL|windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil,
		windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT|windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if e != nil {
		return nil, e
	}
	f := os.NewFile(uintptr(h), path)
	defer f.Close()
	if validateWindowsPrivateHandle(h, directory) != nil {
		return nil, ErrAuthState
	}
	return f.Stat()
}

func (s *Store) stateInfo() (os.FileInfo, error) {
	return privatePathInfo(filepath.Join(s.dir, "state.json"), false)
}
