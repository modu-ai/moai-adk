package cli

// @MX:NOTE: [AUTO] Cross-platform home directory resolution with test override support
// @MX:NOTE: [AUTO] Checks HOME env first (Windows: os.UserHomeDir ignores HOME)

import "github.com/modu-ai/moai-adk/internal/paths"

// userHomeDir returns the current user's home directory.
// Thin delegate to internal/paths.Home since SPEC-V3R6-MOAI-HOME-PATHS-001
// (REQ-MHP-010): exactly one implementation of the HOME-first contract
// remains, and this shim keeps the package-local call sites and the
// userHomeDirFn test seam stable.
// It checks the HOME environment variable first so that tests can override
// the home directory via t.Setenv("HOME", tmpDir) on all platforms,
// including Windows where os.UserHomeDir() ignores HOME in favour of
// USERPROFILE/HOMEPATH/HOMEDRIVE.
// If HOME is not set, it falls back to os.UserHomeDir().
func userHomeDir() (string, error) {
	return paths.Home()
}
