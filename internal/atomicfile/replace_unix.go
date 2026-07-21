//go:build !windows

package atomicfile

import "os"

// replace is a direct os.Rename passthrough. POSIX rename(2) atomically
// replaces the destination regardless of open handles, so no retry or
// fallback is needed and behavior is byte-identical to a bare os.Rename.
func replace(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}
