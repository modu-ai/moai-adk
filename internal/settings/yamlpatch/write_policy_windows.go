package yamlpatch

import (
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
)

func createModeTemp(path string, mode, override os.FileMode) (*os.File, error) {
	return atomicfile.CreatePolicyTemp(path, ".yamlpatch-*.tmp", mode, override)
}

// Native permissions were established before the stage was exposed to data.
func chmodModeTemp(_ string, _ os.FileMode) error { return nil }

func missingWriteTarget(path string, err error) bool {
	if !os.IsNotExist(err) {
		return false
	}
	// Windows ERROR_PATH_NOT_FOUND also describes a file used as an ancestor.
	// Only a genuinely absent suffix below an existing directory is greenfield.
	for parent := filepath.Dir(path); ; parent = filepath.Dir(parent) {
		info, parentErr := os.Stat(parent)
		if parentErr == nil {
			return info.IsDir()
		}
		if !os.IsNotExist(parentErr) || filepath.Dir(parent) == parent {
			return false
		}
	}
}
