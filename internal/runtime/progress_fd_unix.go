//go:build darwin || linux

package runtime

import (
	"os"
	"syscall"
)

// fdMatchesName reports whether name still refers to the inode the held
// descriptor points at (same device + inode). A mismatch means another
// writer swapped the directory entry — typically replacing the temp with
// a symlink to a project-external file — and every name-based step after
// it must fail closed instead of touching that entry (sync-audit-7 F13).
func fdMatchesName(f *os.File, name string) bool {
	var fst syscall.Stat_t
	if err := syscall.Fstat(int(f.Fd()), &fst); err != nil {
		return false
	}
	info, err := os.Lstat(name)
	if err != nil {
		return false
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	return ok && st.Dev == fst.Dev && st.Ino == fst.Ino
}
