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
	return buildSmartPATHFor(runtime.GOOS, homeDir, os.Getenv, isExistingDir, runtime.GOOS == "linux" && IsWSL2())
}

// isExistingDir reports whether path exists and is a directory. A regular
// file at a candidate location does not count: the probe semantics require a
// directory that can hold executables.
func isExistingDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// @MX:ANCHOR: [AUTO] buildSmartPATHFor is the GOOS-injected, deterministic core of BuildSmartPATH
// @MX:REASON: [AUTO] fan_in 5 through the BuildSmartPATH wrapper (initializer.go:412, update.go:1019, update_template_sync.go:275+335, update_clean_install.go:449) — every settings.json write surface routes through this one generator, so its per-platform output shape is an invariant contract
// @MX:SPEC: SPEC-WIN-SMARTPATH-001
//
// buildSmartPATHFor assembles the platform-appropriate PATH from fully
// injected inputs (goos, home, envLookup, stat, wsl2), which makes its output
// a pure function of its arguments — table tests can pin exact output per
// platform without touching the host filesystem or runtime.GOOS.
func buildSmartPATHFor(goos, home string, envLookup func(string) string, stat func(string) bool, wsl2 bool) string {
	// The list separator is a property of the TARGET platform, not the build
	// host: os.PathListSeparator is pinned to the build GOOS at compile time,
	// so it is derived from the injected goos instead. unix builds join with
	// ":" exactly as os.PathListSeparator does there; windows joins with ";".
	sep := ":"
	if goos == "windows" {
		sep = ";"
	}

	// User-specific directories (always included, cross-platform)
	candidates := []string{
		filepath.Join(home, ".local", "bin"), // XDG user-local binaries
		filepath.Join(home, "go", "bin"),     // Go workspace binaries
	}

	// Platform-specific package manager and system paths
	switch goos {
	case "windows":
		// @MX:NOTE: [AUTO] Windows assembles Windows-shaped entries only — the POSIX system tail below is never appended under windows (GH #1690: the GOOS-blind generator fell into the default/linux branch and shipped a ";-joined" mix of POSIX dirs that broke exec-form hook bash resolution).
		if systemRoot := envLookup("SystemRoot"); systemRoot != "" {
			candidates = append(candidates, filepath.Join(systemRoot, "System32"))
		}
		// Git Bash candidates resolve from standard install-location
		// environment variables (no hardcoded C:\ literals in source) and are
		// appended only when the probe reports an existing directory. Git for
		// Windows registers only Git\cmd on the system PATH; bash.exe lives
		// in Git\bin, which exec-form hooks must be able to resolve.
		if pf := envLookup("ProgramFiles"); pf != "" && stat(filepath.Join(pf, "Git", "bin")) {
			candidates = append(candidates, filepath.Join(pf, "Git", "bin"))
		}
		if pf86 := envLookup("ProgramFiles(x86)"); pf86 != "" && stat(filepath.Join(pf86, "Git", "bin")) {
			candidates = append(candidates, filepath.Join(pf86, "Git", "bin"))
		}
		if lad := envLookup("LOCALAPPDATA"); lad != "" && stat(filepath.Join(lad, "Programs", "Git", "bin")) {
			candidates = append(candidates, filepath.Join(lad, "Programs", "Git", "bin"))
		}
	case "darwin":
		if brewPrefix := envLookup("HOMEBREW_PREFIX"); brewPrefix != "" {
			candidates = append(candidates,
				filepath.Join(brewPrefix, "bin"),
				filepath.Join(brewPrefix, "sbin"),
			)
		}
		// Always include standard Homebrew paths as fallback (ensures binaries are accessible even if HOMEBREW_PREFIX is not set)
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

	// Standard POSIX system paths (always required — except under windows,
	// where POSIX dirs must never appear; see the windows case above)
	if goos != "windows" {
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
	if goos == "linux" && wsl2 {
		seen := make(map[string]bool, len(candidates))
		for _, c := range candidates {
			seen[strings.TrimRight(c, "/\\")] = true
		}
		for _, entry := range strings.Split(envLookup("PATH"), sep) {
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
