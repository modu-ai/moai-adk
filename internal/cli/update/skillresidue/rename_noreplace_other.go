//go:build !linux && !darwin && !windows

package skillresidue

import (
	"os"
	"path/filepath"
	"syscall"
)

// renameIsDescriptorBound is false here (no rename primitive at all, the link fallback runs).
const renameIsDescriptorBound = false

// noReplaceRenameAt reports that this platform has no rename that refuses to replace, so the
// held-back rename uses the link-based path.
func noReplaceRenameAt(*os.File, string, *os.File, string) error { return errNoReplaceUnsupported }

// removeDirAt removes the empty folder name inside dir with rmdir (path-based here).
func removeDirAt(dir *os.File, name string) error {
	return syscall.Rmdir(filepath.Join(dir.Name(), name))
}
