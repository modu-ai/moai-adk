package mission

import (
	"io/fs"
	"runtime"
)

// POSIX permission bits do not exist on Windows: os.Stat reports 0666 (or 0444
// for a read-only file) whatever access control list the file carries, so a
// check for exactly 0600 or for "no group or other bits" rejects every file
// there ("receipt_unsafe"). The operator policy of 2026-11-06 is that on Windows
// only the portable properties are checked (the file is regular, is not a
// symlink, can be read); POSIX keeps the exact bit checks (card t2130, #1969).

// privateModeOK reports whether perm satisfies the "exactly want" requirement.
func privateModeOK(goos string, perm, want fs.FileMode) bool {
	if goos == "windows" {
		return true
	}
	return perm == want
}

// noGroupOrOtherAccess reports whether perm grants nothing to group or others.
func noGroupOrOtherAccess(goos string, perm fs.FileMode) bool {
	if goos == "windows" {
		return true
	}
	return perm&0o077 == 0
}

// hostPrivateModeOK and hostNoGroupOrOtherAccess apply the rule to the running platform.
func hostPrivateModeOK(perm, want fs.FileMode) bool { return privateModeOK(runtime.GOOS, perm, want) }

func hostNoGroupOrOtherAccess(perm fs.FileMode) bool {
	return noGroupOrOtherAccess(runtime.GOOS, perm)
}
