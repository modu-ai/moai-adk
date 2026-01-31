package handlers

import (
	"os"
	"path/filepath"
)

// findProjectRoot finds the project root directory by searching up
// for .git or .moai directory markers.
func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}

	// Search up for .git or .moai directory
	maxDepth := 5
	for i := 0; i < maxDepth; i++ {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		if _, err := os.Stat(filepath.Join(dir, ".moai")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "."
}
