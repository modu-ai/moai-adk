//go:build !windows

package privatefile

import (
	"fmt"
	"os"
)

func checkPublicationHandleNative(f *os.File) error {
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("private file: not a regular file")
	}
	if err := checkSingleLinkNative(f); err != nil {
		return err
	}
	return checkNative(f.Name(), info)
}
