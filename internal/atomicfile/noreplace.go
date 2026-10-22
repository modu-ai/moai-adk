package atomicfile

import "os"

// RenameNoReplace atomically publishes oldpath at an absent newpath. Existing
// destinations are refused by the kernel, rather than an earlier existence
// check. Unsupported kernels/filesystems return an error; there is no replace
// or copy fallback. Callers own private staging, same-volume placement and
// cleanup. Path-based ancestor races are outside this primitive's contract.
func RenameNoReplace(oldpath, newpath string) error {
	if err := renameNoReplace(oldpath, newpath); err != nil {
		return &os.LinkError{Op: "rename-noreplace", Old: oldpath, New: newpath, Err: err}
	}
	return nil
}
