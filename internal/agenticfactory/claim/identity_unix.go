//go:build unix

package claim

import (
	"fmt"
	"os"
	"syscall"
)

// fileIdentity names the directory itself (device and inode), so two spellings
// of it, including ones that differ only in case on a case-insensitive
// filesystem, have the same identity.
func fileIdentity(fi os.FileInfo) (string, bool) {
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return "", false
	}
	return fmt.Sprintf("%d:%d", uint64(st.Dev), uint64(st.Ino)), true
}
