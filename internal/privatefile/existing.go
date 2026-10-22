package privatefile

import (
	"fmt"
	"os"
)

// CheckDirectory verifies a real directory's native owner-only policy.
func CheckDirectory(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("private directory: not a real directory")
	}
	return checkDirectoryNative(path, info)
}

// SecureDirectory narrows an existing directory, refusing symbolic-link inputs.
// Windows uses a protected owner-only inheritable DACL; Unix keeps mode 0700.
func SecureDirectory(path string) error { return secureExisting(path, true) }

// SecureDirectoryNoPropagation narrows only this directory's access policy.
// Windows holds an exclusive handle so existing descendants are not changed;
// the directory retains the same protected owner-only inheritable DACL.
func SecureDirectoryNoPropagation(path string) error {
	return secureExistingWithPropagation(path, true, false)
}

// SecureFile narrows an existing regular file without changing its contents.
// Symbolic-link inputs are refused.
// A foreign or unknown Windows owner is refused, never taken over.
func SecureFile(path string) error { return secureExisting(path, false) }

func secureExisting(path string, directory bool) error {
	return secureExistingWithPropagation(path, directory, true)
}

func secureExistingWithPropagation(path string, directory, propagate bool) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || (directory && !info.IsDir()) || (!directory && !info.Mode().IsRegular()) {
		return fmt.Errorf("private path: unexpected object type")
	}
	return secureExistingNative(path, info, directory, propagate)
}

// OpenAppend opens an owner-only regular file for append, protecting its native
// access policy before the caller can write bytes. Links and other file types
// are refused. Native append semantics preserve prior records.
func OpenAppend(path string) (*os.File, error) {
	before, err := os.Lstat(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err == nil && !before.Mode().IsRegular() {
		return nil, fmt.Errorf("private append: not a regular file")
	}
	f, err := openAppendNative(path)
	if err != nil {
		return nil, err
	}
	opened, err := f.Stat()
	if err == nil && (!opened.Mode().IsRegular() || (before != nil && !os.SameFile(before, opened))) {
		err = fmt.Errorf("private append: file identity changed")
	}
	if err == nil {
		err = restrictHandleNative(f, false)
	}
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	return f, nil
}

// Mkdir creates an owner-only directory. Windows installs the protected DACL
// at creation, before the directory can receive private payloads.
func Mkdir(path string) error {
	if err := mkdirNative(path); err != nil {
		return err
	}
	return CheckDirectory(path)
}
