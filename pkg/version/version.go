package version

import "fmt"

// Build-time variables injected via -ldflags.
// Default version for RC/test builds (overridden by -ldflags in production)
var (
	Version = "v2.4.5"
	Commit  = "none"
	Date    = "unknown"
)

// GetVersion returns the current version string.
func GetVersion() string {
	return Version
}

// GetCommit returns the build commit hash.
func GetCommit() string {
	return Commit
}

// GetDate returns the build date.
func GetDate() string {
	return Date
}

// GetFullVersion returns a formatted full version string.
func GetFullVersion() string {
	return fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, Date)
}
