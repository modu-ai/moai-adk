package utils

import (
	"os"
	"path/filepath"
)

const (
	// MoAIConfigDir is the default MoAI configuration directory name.
	MoAIConfigDir = ".moai"

	// ClaudeConfigDir is the Claude Code configuration directory name.
	ClaudeConfigDir = ".claude"
)

// FindProjectRoot walks up from the current directory to find the project root
// (directory containing .moai/ or .claude/).
func FindProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, MoAIConfigDir)); err == nil {
			return dir, nil
		}
		if _, err := os.Stat(filepath.Join(dir, ClaudeConfigDir)); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

// GetMoAIConfigPath returns the path to the .moai/ config directory.
func GetMoAIConfigPath() (string, error) {
	if envDir := os.Getenv("MOAI_CONFIG_DIR"); envDir != "" {
		return envDir, nil
	}

	root, err := FindProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, MoAIConfigDir), nil
}
