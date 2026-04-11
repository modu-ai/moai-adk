package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

const profilesDir = ".moai/claude-profiles"

// BaseDirOverride allows tests to inject a custom base directory.
// When non-empty, GetBaseDir returns this value instead of ~/.moai/claude-profiles.
var BaseDirOverride string

// GetBaseDir returns ~/.moai/claude-profiles/.
func GetBaseDir() string {
	if BaseDirOverride != "" {
		return BaseDirOverride
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot determine home directory: %v\n", err)
		return "."
	}
	return filepath.Join(home, profilesDir)
}

// GetCurrentName returns the current profile name based on CLAUDE_CONFIG_DIR.
func GetCurrentName() string {
	configDir := os.Getenv(config.EnvClaudeConfigDir)
	if configDir == "" {
		return "default"
	}

	baseDir := GetBaseDir()

	rel, err := filepath.Rel(baseDir, configDir)
	if err != nil || strings.HasPrefix(rel, "..") || rel == "." {
		return configDir
	}

	parts := strings.SplitN(rel, string(filepath.Separator), 2)
	return parts[0]
}

// ProfileEntry represents a single profile in the list.
type ProfileEntry struct {
	Name    string
	Current bool
}

// List returns all profile names with an indicator for the current one.
func List() []ProfileEntry {
	currentProfile := GetCurrentName()
	baseDir := GetBaseDir()

	var entries []ProfileEntry
	entries = append(entries, ProfileEntry{
		Name:    "default",
		Current: currentProfile == "default",
	})

	dirEntries, err := os.ReadDir(baseDir)
	if err != nil {
		return entries
	}

	for _, entry := range dirEntries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		entries = append(entries, ProfileEntry{
			Name:    name,
			Current: name == currentProfile,
		})
	}

	return entries
}

// Delete removes a profile directory. Returns error if it's the default profile
// or doesn't exist.
func Delete(name string) error {
	if name == "default" {
		return fmt.Errorf("cannot delete the default profile")
	}

	profileDir := filepath.Join(GetBaseDir(), name)

	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		return fmt.Errorf("profile %q does not exist", name)
	}

	currentProfile := GetCurrentName()
	if currentProfile == name {
		fmt.Fprintf(os.Stderr, "Warning: %q is the currently active profile\n", name)
		fmt.Fprintf(os.Stderr, "Run: moai cc (without -p) to use default\n")
	}

	if err := os.RemoveAll(profileDir); err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Deleted profile: %s\n", name)
	return nil
}

// GetProfileDir returns the directory path for a named profile without creating it.
// Returns an empty string for invalid profile names.
func GetProfileDir(name string) string {
	if name == "" || name == "default" {
		return ""
	}
	if !isValidProfileName(name) {
		return ""
	}
	return filepath.Join(GetBaseDir(), name)
}

// isValidProfileName checks that a profile name does not contain path traversal
// characters. Names must not contain slashes, backslashes, or start with a dot.
func isValidProfileName(name string) bool {
	if strings.Contains(name, "/") || strings.Contains(name, "\\") ||
		strings.HasPrefix(name, ".") || filepath.IsAbs(name) {
		return false
	}
	return true
}

// EnsureDir creates the profile directory if it doesn't exist and sets
// CLAUDE_CONFIG_DIR in the current process.
func EnsureDir(name string) error {
	if name == "" || name == "default" {
		return nil
	}
	// Validate: no path traversal possible
	if !isValidProfileName(name) {
		return fmt.Errorf("invalid profile name %q: must not contain path separators or start with '.'", name)
	}
	profileDir := filepath.Join(GetBaseDir(), name)
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return fmt.Errorf("failed to create profile directory: %w", err)
	}
	if err := os.Setenv("CLAUDE_CONFIG_DIR", profileDir); err != nil {
		return fmt.Errorf("set CLAUDE_CONFIG_DIR: %w", err)
	}
	return nil
}
