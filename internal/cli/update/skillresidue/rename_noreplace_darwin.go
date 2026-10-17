//go:build darwin

package skillresidue

import "golang.org/x/sys/unix"

// renameatNoReplace is renameatx_np(RENAME_EXCL), relative to the two folder descriptors.
func renameatNoReplace(sfd int, src string, dfd int, dst string) error {
	return unix.RenameatxNp(sfd, src, dfd, dst, unix.RENAME_EXCL)
}
