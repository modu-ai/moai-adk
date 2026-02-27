package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// @MX:ANCHOR: [AUTO] Core function for project root discovery. All .moai operations are anchored to the project root.
// @MX:REASON: [AUTO] fan_in=10+, used for root path discovery in all project operations
// FindProjectRoot locates the project root directory by searching for .moai directory.
// It starts from the current working directory and traverses upward until it finds .moai.
// Returns the absolute path to the project root, or an error if not in a MoAI project.
//
// This function ensures that all .moai operations (checkpoints, memory, etc.)
// are anchored to the project root, preventing duplicate .moai directories in subfolders.
// ~/.moai/ is treated as global state (credentials, cache), not a project root.
func FindProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	// Normalize to resolve symlinks (macOS /private/var) and Windows 8.3 short paths.
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		dir = resolved
	}

	// Convert to absolute path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve absolute path: %w", err)
	}

	// Resolve home directory to prevent ~/.moai/ being treated as a project root.
	homeDir, _ := os.UserHomeDir()
	if homeDir != "" {
		if resolved, err := filepath.EvalSymlinks(homeDir); err == nil {
			homeDir = resolved
		}
	}

	// Traverse upward to find .moai directory
	for {
		// Skip home directory — ~/.moai/ is global state, not a project root.
		if homeDir != "" && absDir == homeDir {
			return "", fmt.Errorf("not in a MoAI project (no .moai directory found in project directories)")
		}

		moaiPath := filepath.Join(absDir, ".moai")
		if info, err := os.Stat(moaiPath); err == nil && info.IsDir() {
			return absDir, nil
		}

		parent := filepath.Dir(absDir)
		if parent == absDir {
			// Reached root directory without finding .moai
			return "", fmt.Errorf("not in a MoAI project (no .moai directory found in %s or any parent directory)", absDir)
		}
		absDir = parent
	}
}

// FindProjectRootOrCurrent is like FindProjectRoot but returns the current directory
// instead of an error when not in a MoAI project. This is useful for operations
// that can work in non-project contexts.
func FindProjectRootOrCurrent() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	// Try to find project root
	if root, err := FindProjectRoot(); err == nil {
		return root, nil
	}

	// Not in a project, return current directory (normalized to resolve symlinks
	// and Windows 8.3 short paths for consistent path comparison in tests).
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve absolute path: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(absDir); err == nil {
		return resolved, nil
	}
	return absDir, nil
}

// MustFindProjectRoot is like FindProjectRoot but panics on error.
// Use this in contexts where the project root is guaranteed to exist.
func MustFindProjectRoot() string {
	root, err := FindProjectRoot()
	if err != nil {
		panic(fmt.Sprintf("failed to find project root: %v", err))
	}
	return root
}
