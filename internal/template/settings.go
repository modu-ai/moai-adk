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

	// Cross-platform Node.js runtime paths (volta, bun)
	// These are stable well-known locations that don't change per Node version,
	// unlike nvm which uses version-specific paths (handled by bash -l on POSIX).
	candidates = append(candidates,
		filepath.Join(homeDir, ".volta", "bin"), // Volta (cross-platform)
		filepath.Join(homeDir, ".bun", "bin"),   // Bun (cross-platform)
		filepath.Join(homeDir, ".fnm"),          // fnm shims (cross-platform)
	)

	// Platform-specific package manager and system paths
	switch runtime.GOOS {
	case "windows":
		// Node.js and npm paths
		programFiles := os.Getenv("ProgramFiles")
		appData := os.Getenv("APPDATA")
		localAppData := os.Getenv("LOCALAPPDATA")
		if programFiles != "" {
			candidates = append(candidates, filepath.Join(programFiles, "nodejs"))
		}
		if appData != "" {
			candidates = append(candidates, filepath.Join(appData, "npm"))
		}
		if localAppData != "" {
			candidates = append(candidates, filepath.Join(localAppData, "fnm_multishells"))
		}
		// Windows system paths
		systemRoot := os.Getenv("SystemRoot")
		if systemRoot == "" {
			systemRoot = `C:\Windows`
		}
		candidates = append(candidates,
			filepath.Join(systemRoot, "System32"),
			systemRoot,
		)
	case "darwin":
		candidates = append(candidates,
			"/opt/homebrew/bin",  // Apple Silicon Homebrew
			"/opt/homebrew/sbin", // Apple Silicon Homebrew system
			"/usr/local/bin",     // Intel Homebrew / system
			"/usr/local/sbin",    // Intel Homebrew system
		)
		// Standard POSIX system paths
		candidates = append(candidates, "/usr/bin", "/bin", "/usr/sbin", "/sbin")
	default: // linux, etc.
		candidates = append(candidates,
			"/usr/local/bin",
			"/usr/local/sbin",
		)
		// Standard POSIX system paths
		candidates = append(candidates, "/usr/bin", "/bin", "/usr/sbin", "/sbin")
	}

	// WSL2: append Windows drive-mount paths from the current terminal PATH.
	// WSL2 maps Windows drives as /mnt/<letter>/ (e.g., /mnt/c/, /mnt/d/),
	// which allows running Windows executables (powershell.exe, cmd.exe, etc.).
	// Without this, writing a static PATH to settings.json would remove those
	// entries, causing "command not found: powershell.exe" (issue #495).
	//
	// Only paths matching the WSL2 drive-mount pattern (/mnt/<single-letter>/...)
	// are included. This filters out non-drive mounts (e.g., /mnt/wslg, /mnt/foo)
	// while preserving all legitimate Windows drive paths regardless of depth
	// (e.g., /mnt/c/Windows/System32, /mnt/d/tools/bin).
	if runtime.GOOS == "linux" && IsWSL2() {
		seen := make(map[string]bool, len(candidates))
		for _, c := range candidates {
			seen[strings.TrimRight(c, "/\\")] = true
		}
		for _, entry := range strings.Split(os.Getenv("PATH"), sep) {
			if isWSL2DrivePath(entry) && !isUserScopedWindowsPath(entry) {
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

// isWSL2DrivePath reports whether entry looks like a WSL2 Windows drive mount.
// WSL2 maps Windows drives as /mnt/<letter>/ (e.g., /mnt/c/, /mnt/d/).
// Only single lowercase-letter mounts are accepted to exclude non-drive mounts
// such as /mnt/wslg or /mnt/foo.
func isWSL2DrivePath(entry string) bool {
	if len(entry) < 6 || entry[:5] != "/mnt/" {
		return false
	}
	letter := entry[5]
	return letter >= 'a' && letter <= 'z' && (len(entry) == 6 || entry[6] == '/')
}

// isUserScopedWindowsPath reports whether a WSL2 drive-mount path points to a
// per-user Windows directory. Such paths (e.g., /mnt/c/Users/alice/AppData/...)
// are machine-specific and must not be persisted in settings.json, as they would
// break portability and partially undo the fix from issue #467.
//
// Rejected prefixes (case-insensitive):
//   - /users/           — Windows user home directories
//   - /appdata/         — per-user application data
//   - /documents and settings/ — legacy Windows XP user profiles
func isUserScopedWindowsPath(entry string) bool {
	lower := strings.ToLower(entry)
	// Strip the /mnt/<letter> prefix to check the Windows-relative path.
	if len(lower) < 7 {
		return false
	}
	// Find the path after /mnt/<letter>, e.g. "/mnt/c/Users/..." → "/users/..."
	rest := lower[6:] // skip "/mnt/X"
	for _, segment := range []string{"/users/", "/appdata/", "/documents and settings/"} {
		if strings.Contains(rest, segment) {
			return true
		}
	}
	return false
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
