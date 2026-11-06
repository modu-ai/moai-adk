//go:build linux || darwin

package skillresidue

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// renameIsDescriptorBound reports that noReplaceRenameAt resolves both names against the open
// folders' descriptors, so a move between two held folders needs no look at their paths and a
// put-back can still reach a folder that was renamed away.
const renameIsDescriptorBound = true

// withDirFd runs fn with the descriptor of an open folder. The descriptor is only valid
// inside fn, and the file is kept alive for its duration.
func withDirFd(dir *os.File, fn func(fd int) error) error {
	rc, err := dir.SyscallConn()
	if err != nil {
		return err
	}
	var inner error
	if err := rc.Control(func(fd uintptr) { inner = fn(int(fd)) }); err != nil {
		return err
	}
	return inner
}

// noReplaceRenameAt renames src in srcDir to dst in dstDir, both folders held open, with the
// platform's atomic no-replace rename (renameatNoReplace): the kernel refuses, in one atomic
// step, when dst exists. The names are resolved against the open folders' descriptors, so a
// path that is swapped after the folders were opened cannot redirect the rename. A destination
// that exists is reported with the operating system's own error (errors.Is(err, fs.ErrExist));
// a kernel or filesystem that does not know the flag is reported as errNoReplaceUnsupported.
func noReplaceRenameAt(srcDir *os.File, src string, dstDir *os.File, dst string) error {
	err := withDirFd(srcDir, func(sfd int) error {
		return withDirFd(dstDir, func(dfd int) error { return renameatNoReplace(sfd, src, dfd, dst) })
	})
	if errors.Is(err, unix.EINVAL) || errors.Is(err, unix.ENOSYS) || errors.Is(err, unix.ENOTSUP) {
		return fmt.Errorf("%w: %v", errNoReplaceUnsupported, err)
	}
	return err
}

// removeDirAt removes the empty folder name inside dir with unlinkat(AT_REMOVEDIR), relative
// to the open folder's descriptor. It fails on a file, a link, and a folder that holds anything.
func removeDirAt(dir *os.File, name string) error {
	return withDirFd(dir, func(fd int) error { return unix.Unlinkat(fd, name, unix.AT_REMOVEDIR) })
}
