package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// @MX:ANCHOR: [AUTO] 프로젝트 루트 탐색의 핵심 함수입니다. 모든 .moai 작업이 프로젝트 루트에 고정됩니다.
// @MX:REASON: [AUTO] fan_in=10+, 모든 프로젝트 작업의 루트 경로 탐색에 사용됩니다
// FindProjectRoot locates the project root directory by searching for .moai directory.
// It starts from the current working directory and traverses upward until it finds .moai.
// Returns the absolute path to the project root, or an error if not in a MoAI project.
//
// This function ensures that all .moai operations (checkpoints, memory, etc.)
// are anchored to the project root, preventing duplicate .moai directories in subfolders.
func FindProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	// Convert to absolute path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve absolute path: %w", err)
	}

	// Traverse upward to find .moai directory
	for {
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

	// Not in a project, return current directory
	return filepath.Abs(dir)
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
