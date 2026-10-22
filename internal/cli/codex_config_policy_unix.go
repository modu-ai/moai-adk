//go:build !windows

package cli

import (
	"io/fs"
	"os"
)

func restoreCodexConfigPolicy(staged, _ string, mode fs.FileMode) error {
	return os.Chmod(staged, mode)
}
