//go:build !windows

package testguard

import (
	"errors"
	"fmt"
	"os"
)

// CheckFileModePolicy checks the platform's actual file-permission contract.
// Unix retains the exact POSIX bits; Windows checks native owner/DACL/access.
// This helper is only for callers in _test.go files.
func CheckFileModePolicy(path string, perm os.FileMode) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm() != perm {
		return fmt.Errorf("file policy = %v, want regular file with exact %04o", info.Mode(), perm)
	}
	return nil
}

// SeedFileModePolicy creates an exclusive fixture with a verified native
// permission policy. It never substitutes a fabricated portable mode value.
func SeedFileModePolicy(path string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	_, writeErr := f.Write(data)
	chmodErr := f.Chmod(perm)
	closeErr := f.Close()
	if err := errors.Join(writeErr, chmodErr, closeErr); err != nil {
		return err
	}
	return CheckFileModePolicy(path, perm)
}
