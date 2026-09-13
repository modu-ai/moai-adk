package codexapp

import (
	"os"
	"path/filepath"
)

// ValidatePrivatePath checks the same owner/ACL boundary as an App Server
// profile. The caller must serialize mutations; this is not a filesystem lease.
func ValidatePrivatePath(path string, directory bool) error {
	if !filepath.IsAbs(path) {
		return os.ErrPermission
	}
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !privateOwner(path, info) || (directory && !info.IsDir()) || (!directory && !info.Mode().IsRegular()) {
		return os.ErrPermission
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return err
	}
	if resolved != filepath.Clean(path) {
		return os.ErrPermission
	}
	return nil
}
