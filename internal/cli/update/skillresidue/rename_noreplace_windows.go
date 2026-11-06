//go:build windows

package skillresidue

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/windows"
)

// renameIsDescriptorBound is false here: the rename is path-based, so every move, a put-back
// included, is preceded by the root-bound look at its source path.
const renameIsDescriptorBound = false

// noReplaceRenameAt renames src in srcDir to dst in dstDir with MoveFileEx and no flags:
// without MOVEFILE_REPLACE_EXISTING the call fails when dst exists. A destination that exists
// is reported with the operating system's own error (errors.Is(err, fs.ErrExist)); a volume
// that cannot do the call is reported as errNoReplaceUnsupported.
//
// Unlike the unix builds this rename is path-based: Windows has no folder-relative no-replace
// rename in golang.org/x/sys, so the open folders only supply the paths. The root-bound look
// that moveNoReplace makes before each rename is the containment, and a microsecond window
// remains between the look and the call.
func noReplaceRenameAt(srcDir *os.File, src string, dstDir *os.File, dst string) error {
	from16, err := windows.UTF16PtrFromString(filepath.Join(srcDir.Name(), src))
	if err != nil {
		return err
	}
	to16, err := windows.UTF16PtrFromString(filepath.Join(dstDir.Name(), dst))
	if err != nil {
		return err
	}
	err = windows.MoveFileEx(from16, to16, 0)
	if errors.Is(err, windows.ERROR_NOT_SUPPORTED) {
		return fmt.Errorf("%w: %v", errNoReplaceUnsupported, err)
	}
	return err
}

// removeDirAt removes the empty folder name inside dir with rmdir (path-based here).
func removeDirAt(dir *os.File, name string) error {
	return syscall.Rmdir(filepath.Join(dir.Name(), name))
}
