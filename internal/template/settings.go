package template

import (
	"os"
	"path/filepath"
	"strings"
)

// BuildSmartPATH captures the current terminal PATH and ensures essential directories are included.
// Unlike hardcoded approaches, this preserves all user-installed tool paths (nvm, pyenv, cargo, etc.)
// while ensuring moai-essential directories are present.
// Used by TemplateContext.SmartPATH for settings.json.tmpl rendering.
func BuildSmartPATH() string {
	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		homeDir = os.Getenv("HOME")
	}

	currentPATH := os.Getenv("PATH")
	sep := string(os.PathListSeparator)

	// Essential directories that must be in PATH for moai to function
	essentialDirs := []string{
		filepath.Join(homeDir, ".local", "bin"),
		filepath.Join(homeDir, "go", "bin"),
	}

	// Prepend essential dirs if not already present
	for i := len(essentialDirs) - 1; i >= 0; i-- {
		dir := essentialDirs[i]
		if !PathContainsDir(currentPATH, dir, sep) {
			currentPATH = dir + sep + currentPATH
		}
	}

	return currentPATH
}

// PathContainsDir checks if a PATH string contains a specific directory entry.
// Handles trailing slashes and exact segment matching to avoid false positives
// (e.g., "/usr/local/bin" should not match "/usr/local/bin2").
func PathContainsDir(pathStr, dir, sep string) bool {
	dir = strings.TrimRight(dir, "/\\")

	for entry := range strings.SplitSeq(pathStr, sep) {
		entry = strings.TrimRight(entry, "/\\")
		if entry == dir {
			return true
		}
	}
	return false
}
