//go:build !windows

package privatefile

import "os"

func createExclusiveNative(path string, acquired func(*os.File)) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	acquired(f)
	info, err := f.Stat()
	if err != nil {
		return err
	}
	return checkNative(path, info)
}
