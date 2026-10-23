//go:build !(linux || darwin || freebsd || netbsd || openbsd || dragonfly)

package runtime

import (
	"fmt"
	"os"
)

// openRegularFile is the variant for platforms without O_NOFOLLOW/O_NONBLOCK
// (windows): it takes an Lstat before the open, opens, then requires the
// OPENED file to be regular AND to be the very file the Lstat saw
// (os.SameFile), so a link swapped in between the two is refused.
func openRegularFile(path string) (*os.File, error) {
	before, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !before.Mode().IsRegular() {
		return nil, notRegularError(path, before.Mode())
	}
	beforeRegularOpen(path)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if !info.Mode().IsRegular() {
		_ = f.Close()
		return nil, notRegularError(path, info.Mode())
	}
	if !os.SameFile(before, info) {
		_ = f.Close()
		return nil, fmt.Errorf("%s is not a regular file (the path was replaced between the check and the open)", path)
	}
	return f, nil
}
