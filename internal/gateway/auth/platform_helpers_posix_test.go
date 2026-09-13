//go:build !windows

package auth

import "os"

func setTestFilePermissions(path string, mode os.FileMode) error { return os.Chmod(path, mode) }
func blockTestCleanup(path string) (func(), error) {
	e := os.Chmod(path, 0000)
	return func() { os.Chmod(path, 0700) }, e
}
