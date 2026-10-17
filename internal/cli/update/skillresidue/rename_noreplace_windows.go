//go:build windows

package skillresidue

import (
	"errors"
	"fmt"

	"golang.org/x/sys/windows"
)

// noReplaceRename renames from to to with MoveFileEx and no flags: without
// MOVEFILE_REPLACE_EXISTING the call fails when to exists. A destination that exists is
// reported with the operating system's own error (it satisfies errors.Is(err, fs.ErrExist));
// a volume that cannot do the call is reported as errNoReplaceUnsupported.
func noReplaceRename(from, to string) error {
	from16, err := windows.UTF16PtrFromString(from)
	if err != nil {
		return err
	}
	to16, err := windows.UTF16PtrFromString(to)
	if err != nil {
		return err
	}
	err = windows.MoveFileEx(from16, to16, 0)
	if errors.Is(err, windows.ERROR_NOT_SUPPORTED) {
		return fmt.Errorf("%w: %v", errNoReplaceUnsupported, err)
	}
	return err
}
