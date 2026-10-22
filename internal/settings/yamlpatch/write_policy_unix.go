//go:build !windows

package yamlpatch

import (
	"os"
	"path/filepath"
)

func createModeTemp(path string, _, _ os.FileMode) (*os.File, error) {
	return os.CreateTemp(filepath.Dir(path), ".yamlpatch-*.tmp")
}

func chmodModeTemp(path string, mode os.FileMode) error { return os.Chmod(path, mode) }

func missingWriteTarget(_ string, err error) bool { return os.IsNotExist(err) }
