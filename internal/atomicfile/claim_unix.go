//go:build !windows

package atomicfile

import "os"

func createExclusive(path string, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, perm)
}
func removeFile(path string) error { return os.Remove(path) }

func openRead(path string) (*os.File, error) { return os.Open(path) }
