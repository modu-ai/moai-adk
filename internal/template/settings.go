package template

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// procVersionPath is the path to the kernel version file used by IsWSL2.
// It is a package-level variable so tests can override it with a temp file
// containing a synthetic kernel string without touching real /proc/version.
var procVersionPath = "/proc/version"

// IsWSL2 reports whether the current process is running inside WSL2
// (Windows Subsystem for Linux). It checks the WSL_DISTRO_NAME environment
// variable first (fastest), then falls back to procVersionPath for robustness.
func IsWSL2() bool {
	// WSL_DISTRO_NAME is set by the WSL runtime (e.g., "Ubuntu", "Debian")
	if os.Getenv("WSL_DISTRO_NAME") != "" {
		return true
	}
	// Fallback: /proc/version contains "microsoft-standard-WSL" on WSL2 kernels
	// (e.g., "6.6.87.2-microsoft-standard-WSL2"). Using the full prefix avoids
	// false positives on Azure VMs that may contain "microsoft" in kernel strings.
	data, err := os.ReadFile(procVersionPath)
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(data)), "microsoft-standard-wsl")
}

// BuildSmartPATH constructs a portable, platform-appropriate PATH for use in settings.json.
// Unlike the previous approach that captured the terminal PATH at init/update time, this
// function builds a stable PATH from well-known locations. This prevents issue #467 where
// machine-specific paths (e.g., Linux paths from CI) were baked into settings.json and
// broke MCP servers on macOS.
//
// Exception: In WSL2 environments, Windows interop paths (entries starting with /mnt/)
// are extracted from the current terminal PATH and appended. This ensures that Windows
// executables like powershell.exe remain accessible after settings.json is written (issue #495).
//
// Used by TemplateContext.SmartPATH for settings.json.tmpl rendering.
func BuildSmartPATH() string {
	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		homeDir = os.Getenv("HOME")
	}

	sep := string(os.PathListSeparator)

	// User-specific directories (always included, cross-platform)
	candidates := []string{
		filepath.Join(homeDir, ".local", "bin"), // XDG user-local binaries
		filepath.Join(homeDir, "go", "bin"),     // Go workspace binaries
	}

	// Platform-specific package manager and system paths
	switch runtime.GOOS {
	case "darwin":
		candidates = append(candidates,
			"/opt/homebrew/bin",  // Apple Silicon Homebrew
			"/opt/homebrew/sbin", // Apple Silicon Homebrew system
			"/usr/local/bin",     // Intel Homebrew / system
			"/usr/local/sbin",    // Intel Homebrew system
		)
	default: // linux, etc.
		candidates = append(candidates,
			"/usr/local/bin",
			"/usr/local/sbin",
		)
	}

	// Standard POSIX system paths (always required)
	candidates = append(candidates, "/usr/bin", "/bin", "/usr/sbin", "/sbin")

	// WSL2: append Windows interop paths from the current terminal PATH.
	// WSL2 automatically adds /mnt/<drive>/... entries to PATH, which allows
	// running Windows executables (powershell.exe, cmd.exe, etc.).
	// Without this, writing a static PATH to settings.json would remove those
	// entries, causing "command not found: powershell.exe" (issue #495).
	//
	// We intentionally capture ALL /mnt/ entries rather than filtering to a
	// known allow-list (e.g., only /mnt/c/Windows/System32) because:
	//   1. Users frequently install custom Windows tools in non-standard locations
	//      such as /mnt/d/tools or /mnt/c/Program Files/... and expect them to
	//      remain accessible inside WSL2 after settings.json is written.
	//   2. An allow-list would be fragile: it would silently drop valid entries
	//      whenever a user's setup deviates from the expected pattern.
	//   3. Security concern is low: these paths originate from the user's own
	//      WSL2 session PATH, which is already trusted.
	if runtime.GOOS == "linux" && IsWSL2() {
		seen := make(map[string]bool, len(candidates))
		for _, c := range candidates {
			seen[strings.TrimRight(c, "/\\")] = true
		}
		for _, entry := range strings.Split(os.Getenv("PATH"), sep) {
			if strings.HasPrefix(entry, "/mnt/") {
				normalized := strings.TrimRight(entry, "/\\")
				if !seen[normalized] {
					candidates = append(candidates, entry)
					seen[normalized] = true
				}
			}
		}
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
