package core

import (
	"os"
	"runtime"
)

// TempDir returns the appropriate temp directory.
// On Windows, allows override to avoid 8.3 short filename issues with non-ASCII usernames.
func TempDir() string {
	if runtime.GOOS == "windows" {
		// Check if user has configured custom temp dir
		if custom := os.Getenv("MOAI_TEMP_DIR"); custom != "" {
			return custom
		}
	}
	return os.TempDir()
}

// CreateTempFile wraps os.CreateTemp with path normalization for Windows
func CreateTempFile(dir, pattern string) (*os.File, error) {
	if dir == "" {
		dir = TempDir()
	}
	return os.CreateTemp(dir, pattern)
}
