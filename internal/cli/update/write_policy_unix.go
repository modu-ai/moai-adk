//go:build !windows

package update

import (
	"os"
	"path/filepath"
)

func createRestoreTemp(path string, _ os.FileMode) (*os.File, error) {
	return os.CreateTemp(filepath.Dir(path), ".moai-merge-*")
}

func chmodRestoreTemp(path string, mode os.FileMode) error { return os.Chmod(path, mode) }
