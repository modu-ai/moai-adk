//go:build !windows

package harness

import (
	"os"
	"syscall"
)

// entryOwnedByCurrentUser reports whether the entry at path is owned by the effective user. It reads
// the entry's own record without following a symbolic link, so a link counts as owned when the link
// itself is, whatever its target's owner. An entry that cannot be read counts as not owned.
func entryOwnedByCurrentUser(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return false
	}
	return int64(st.Uid) == int64(os.Geteuid())
}
