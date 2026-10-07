//go:build !darwin && !linux

package runtime

import "os"

// seedFileMetadata makes the temp file carry the original's metadata on
// platforms with no explicit-file ACL/xattr axis this package copies (the
// mode, via a plain stat/chmod — on Windows the temp file also inherits
// the directory's ACL at creation, and ownership preservation has no
// portable axis here: both postures are flagged for leader review).
// Round-4 class closure; a failure is an error and the caller aborts the
// replace.
func seedFileMetadata(tmp *os.File, tmpPath, original string) error {
	info, err := os.Stat(original)
	if err != nil {
		return err
	}
	// The mode rides the HELD descriptor (F13).
	return tmp.Chmod(info.Mode().Perm())
}
