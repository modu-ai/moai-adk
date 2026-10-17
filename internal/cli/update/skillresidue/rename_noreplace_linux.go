//go:build linux

package skillresidue

import (
	"errors"
	"fmt"

	"golang.org/x/sys/unix"
)

// noReplaceRename renames from to to with renameat2(RENAME_NOREPLACE): the kernel refuses,
// in one atomic step, when to exists. A destination that exists is reported with the
// operating system's own error (it satisfies errors.Is(err, fs.ErrExist)); a kernel or
// filesystem that does not know the flag is reported as errNoReplaceUnsupported.
func noReplaceRename(from, to string) error {
	err := unix.Renameat2(unix.AT_FDCWD, from, unix.AT_FDCWD, to, unix.RENAME_NOREPLACE)
	if errors.Is(err, unix.EINVAL) || errors.Is(err, unix.ENOSYS) || errors.Is(err, unix.ENOTSUP) {
		return fmt.Errorf("%w: %v", errNoReplaceUnsupported, err)
	}
	return err
}
