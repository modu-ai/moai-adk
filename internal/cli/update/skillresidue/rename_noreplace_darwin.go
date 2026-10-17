//go:build darwin

package skillresidue

import (
	"errors"
	"fmt"

	"golang.org/x/sys/unix"
)

// noReplaceRename renames from to to with renamex_np(RENAME_EXCL): the kernel refuses, in
// one atomic step, when to exists. A destination that exists is reported with the operating
// system's own error (it satisfies errors.Is(err, fs.ErrExist)); a filesystem that does not
// implement the flag is reported as errNoReplaceUnsupported.
func noReplaceRename(from, to string) error {
	err := unix.RenamexNp(from, to, unix.RENAME_EXCL)
	if errors.Is(err, unix.ENOTSUP) || errors.Is(err, unix.EINVAL) || errors.Is(err, unix.ENOSYS) {
		return fmt.Errorf("%w: %v", errNoReplaceUnsupported, err)
	}
	return err
}
