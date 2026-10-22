//go:build unix

package privatefile

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func checkSingleLinkNative(f *os.File) error {
	var info unix.Stat_t
	if err := unix.Fstat(int(f.Fd()), &info); err != nil {
		return err
	}
	if info.Nlink != 1 {
		return fmt.Errorf("private file: multiple hard links")
	}
	return nil
}
