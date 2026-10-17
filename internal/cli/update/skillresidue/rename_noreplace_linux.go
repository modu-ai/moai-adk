//go:build linux

package skillresidue

import "golang.org/x/sys/unix"

// renameatNoReplace is renameat2(RENAME_NOREPLACE), relative to the two folder descriptors.
func renameatNoReplace(sfd int, src string, dfd int, dst string) error {
	return unix.Renameat2(sfd, src, dfd, dst, unix.RENAME_NOREPLACE)
}
