//go:build darwin

package atomicfile

import "golang.org/x/sys/unix"

func renameNoReplace(oldpath, newpath string) error {
	return unix.RenameatxNp(unix.AT_FDCWD, oldpath, unix.AT_FDCWD, newpath, unix.RENAME_EXCL)
}
