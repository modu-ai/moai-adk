package template

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// BuildSmartPATH captures the current terminal PATH and ensures essential directories
// for the current platform are included.
// Unlike hardcoded approaches, this preserves all user-installed tool paths (nvm, pyenv, cargo, etc.)
// while ensuring moai-essential and platform-specific directories are present.
// Used by TemplateContext.SmartPATH for settings.json.tmpl rendering.
func BuildSmartPATH() string {
	return BuildSmartPATHForPlatform(runtime.GOOS)
}

// BuildSmartPATHForPlatform captures the current terminal PATH and ensures essential
// directories for the given platform are included.
//
// Essential directories are platform-specific (issue #467):
//   - darwin: ~/go/bin, /opt/homebrew/bin, /opt/homebrew/sbin
//     (Homebrew paths are critical for MCP servers installed via brew on macOS)
//   - linux:  ~/go/bin, ~/.local/bin
//   - windows: ~/go/bin
//
// An empty platform string defaults to runtime.GOOS.
func BuildSmartPATHForPlatform(platform string) string {
	if platform == "" {
		platform = runtime.GOOS
	}

	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		homeDir = os.Getenv("HOME")
	}

	currentPATH := os.Getenv("PATH")
	sep := string(os.PathListSeparator)

	// Essential directories vary by platform.
	// Order matters: dirs are prepended in reverse order, so the first entry in
	// essentialDirs ends up first in the resulting PATH.
	var essentialDirs []string
	switch platform {
	case "darwin":
		// macOS: Homebrew paths are critical for MCP servers (npx, uvx, etc.)
		// installed via brew. ~/.local/bin is a Linux convention, not used on macOS.
		essentialDirs = []string{
			filepath.Join(homeDir, "go", "bin"),
			"/opt/homebrew/sbin",
			"/opt/homebrew/bin",
		}
	case "windows":
		essentialDirs = []string{
			filepath.Join(homeDir, "go", "bin"),
		}
	default: // linux and other Unix-like systems
		essentialDirs = []string{
			filepath.Join(homeDir, ".local", "bin"),
			filepath.Join(homeDir, "go", "bin"),
		}
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
