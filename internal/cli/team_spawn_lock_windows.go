//go:build windows

package cli

import (
	"os"
)

// lockFile is a no-op on Windows; flock(2) is not available.
// Windows file locking requires LockFileEx via syscall/windows, but ClaimTask
// is only exercised from tmux-based team workflows that do not run on Windows.
// A no-op is intentional here to allow the binary to compile on Windows
// while preserving correct concurrent behaviour on Unix.
func lockFile(_ *os.File) error {
	return nil
}

// unlockFile is a no-op on Windows; see lockFile.
func unlockFile(_ *os.File) error {
	return nil
}
