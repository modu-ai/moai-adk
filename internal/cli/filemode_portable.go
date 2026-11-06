package cli

import (
	"io/fs"
	"runtime"
)

// POSIX permission bits do not exist on Windows: os.Stat reports 0666 whatever
// access control list a file carries, so a "no group or other bits" check
// rejects every file there. The operator policy of 2026-11-06 is that Windows
// checks only the portable properties (regular file, not a symlink); POSIX
// keeps the exact bit checks (card t2131, #1969).

// noGroupOrOtherAccessOn reports whether perm grants nothing to group or others
// on the named platform; on Windows the bits carry no meaning and it is true.
func noGroupOrOtherAccessOn(goos string, perm fs.FileMode) bool {
	if goos == "windows" {
		return true
	}
	return perm&0o077 == 0
}

// hostNoGroupOrOtherAccess applies noGroupOrOtherAccessOn to the running platform.
func hostNoGroupOrOtherAccess(perm fs.FileMode) bool {
	return noGroupOrOtherAccessOn(runtime.GOOS, perm)
}
