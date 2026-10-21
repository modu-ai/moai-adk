//go:build !windows

package privatefile

import (
	"fmt"
	"os"
)

// CreateTemp creates an empty owner-only temporary file, open for read/write.
func CreateTemp(dir, prefix string) (*os.File, error) { return os.CreateTemp(dir, prefix+"*") }

func checkNative(_ string, info os.FileInfo) error {
	if info.Mode().Perm() != 0o600 {
		return fmt.Errorf("private file: mode %04o, want 0600", info.Mode().Perm())
	}
	return nil
}
