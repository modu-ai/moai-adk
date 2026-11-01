//go:build !(linux || darwin || freebsd || netbsd || openbsd || dragonfly)

package jevledger

import (
	"io/fs"
	"os"
)

// openLedger is the variant for every GOOS the Unix set excludes (windows
// included), so no platform is left without an implementation. These platforms
// have no O_NOFOLLOW, so a link is refused by Lstat before the open; the
// opened handle is then checked exactly as on the Unix set. The Lstat and the
// open are two path resolutions: the window between them is the residual risk
// of this variant (plan.md section H), and the handle check still refuses a
// non-regular object that was swapped in.
func openLedger(path string) (*os.File, error) {
	if info, err := os.Lstat(path); err == nil && info.Mode()&fs.ModeSymlink != 0 {
		return nil, errNotRegular
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, fileMode)
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
		return nil, errNotRegular
	}
	return f, nil
}
