package template

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// BuildSmartPATH constructs a portable, platform-appropriate PATH for use in settings.json.
// Unlike the previous approach that captured the terminal PATH at init/update time, this
// function builds a stable PATH from well-known locations. This prevents issue #467 where
// machine-specific paths (e.g., Linux paths from CI) were baked into settings.json and
// broke MCP servers on macOS.
// Used by TemplateContext.SmartPATH for settings.json.tmpl rendering.
func BuildSmartPATH() string {
	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		homeDir = os.Getenv("HOME")
	}

	sep := string(os.PathListSeparator)

	// User-specific directories (only if homeDir resolved)
	var candidates []string
	if homeDir != "" {
		candidates = append(candidates,
			filepath.Join(homeDir, ".local", "bin"), // XDG user-local binaries
			filepath.Join(homeDir, "go", "bin"),     // Go workspace binaries
		)
	}

	// Platform-specific paths
	switch runtime.GOOS {
	case "darwin":
		candidates = append(candidates,
			"/opt/homebrew/bin",  // Apple Silicon Homebrew
			"/opt/homebrew/sbin", // Apple Silicon Homebrew system
			"/usr/local/bin",     // Intel Homebrew / system
			"/usr/local/sbin",    // Intel Homebrew system
		)
	case "windows":
		// Windows uses %PATH% natively; only add Go bin if homeDir available
		// System paths are managed by Windows itself
	default: // linux, etc.
		candidates = append(candidates,
			"/usr/local/bin",
			"/usr/local/sbin",
		)
	}

	// Standard POSIX system paths (skip on Windows)
	if runtime.GOOS != "windows" {
		candidates = append(candidates, "/usr/bin", "/bin", "/usr/sbin", "/sbin")
	}

	return strings.Join(candidates, sep)
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
