package version

import "fmt"

// Build-time variables injected via -ldflags.
// Default version for RC/test builds (overridden by -ldflags in production)
var (
	Version = "v3.1.3"
	Commit  = "none"
	Date    = "unknown"
	// BuildID is the monotone build identity — the tag plus the commit
	// distance and hash, so two builds in an ancestor relation never read as
	// the same string. Version cannot serve this purpose: it derives with
	// --abbrev=0 and so reports the same tag floor for every commit since that
	// tag, and an explicit release-candidate Version reads higher than a later
	// default build. Empty means the build carried no ldflags.
	BuildID = ""
)

// GetVersion returns the current version string.
func GetVersion() string {
	return Version
}

// GetCommit returns the build commit hash.
func GetCommit() string {
	return Commit
}

// GetBuildID returns the monotone build identity, falling back to the commit
// hash when the build carried no ldflags. It never falls back to Version:
// Version is the string that cannot order two builds, which is the whole
// reason this identity exists.
//
// @MX:NOTE: never falls back to Version — Version cannot order two builds
// @MX:SPEC: SPEC-BINARY-LAG-VISIBILITY-001
func GetBuildID() string {
	if BuildID != "" {
		return BuildID
	}
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
