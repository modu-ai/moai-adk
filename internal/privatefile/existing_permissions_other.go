//go:build !windows && !linux

package privatefile

import "os"

func secureExistingNative(path string, _ os.FileInfo, directory bool) error {
	mode := os.FileMode(0o600)
	if directory {
		mode = 0o700
	}
	// Preserve chmod's ability to repair an owner directory whose current mode
	// denies opening it. Lstat above refuses symlinks as the prior caller did.
	return os.Chmod(path, mode)
}
