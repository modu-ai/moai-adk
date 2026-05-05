package cli

// @MX:NOTE: [AUTO] Cross-platform home directory resolution with test override support
// @MX:NOTE: [AUTO] Checks HOME env first (Windows: os.UserHomeDir ignores HOME)

import "os"

// userHomeDir returns the current user's home directory.
// It checks the HOME environment variable first so that tests can override
// the home directory via t.Setenv("HOME", tmpDir) on all platforms,
// including Windows where os.UserHomeDir() ignores HOME in favour of
// USERPROFILE/HOMEPATH/HOMEDRIVE.
// If HOME is not set, it falls back to os.UserHomeDir().
func userHomeDir() (string, error) {
	if h := os.Getenv("HOME"); h != "" {
		return h, nil
	}
	return os.UserHomeDir()
}
