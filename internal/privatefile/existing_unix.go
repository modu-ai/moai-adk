//go:build !windows

package privatefile

import (
	"fmt"
	"os"
)

func checkDirectoryNative(_ string, info os.FileInfo) error {
	if info.Mode().Perm() != 0o700 {
		return fmt.Errorf("private directory: mode %04o, want 0700", info.Mode().Perm())
	}
	return nil
}

func secureExistingNative(path string, _ os.FileInfo, directory, _ bool) error {
	mode := os.FileMode(0o600)
	if directory {
		mode = 0o700
	}
	// Preserve chmod's ability to repair an owner directory whose current mode
	// denies opening it. Lstat above refuses symlinks as the prior caller did.
	return os.Chmod(path, mode)
}

func restrictHandleNative(f *os.File, directory bool) error {
	mode := os.FileMode(0o600)
	if directory {
		mode = 0o700
	}
	return f.Chmod(mode)
}

func mkdirNative(path string) error { return os.Mkdir(path, 0o700) }
