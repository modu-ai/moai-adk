//go:build !windows

package atomicfile

import "os"

// readFile is a direct os.ReadFile passthrough. POSIX rename(2) swaps the
// directory entry atomically and never leaves the destination momentarily
// unopenable, so a reader racing a replace sees either the old or the new file
// and never needs a retry.
func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
