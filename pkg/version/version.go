package version

import (
	"fmt"
	"strings"
)

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

// IsDevBuild reports whether v identifies a non-release build and therefore
// must never take part in binary self-update (the SessionStart auto-update
// handler and the moai update binary step both gate on this). A version is
// a release build only when it parses as a three-part numeric version with
// an optional leading "v" and prerelease suffix; anything else — build
// codenames like "moai_cp/20260910_130400", "dev", VCS-local markers like
// "-dirty", or empty — is a dev build. The dirty/none substring checks are
// kept because "-dirty"/"-none" suffixes themselves parse as valid
// prereleases.
func IsDevBuild(v string) bool {
	if v == "" || strings.Contains(v, "dirty") || strings.Contains(v, "none") {
		return true
	}
	s := strings.TrimPrefix(v, "v")
	if i := strings.Index(s, "-"); i >= 0 {
		s = s[:i]
	}
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return true
	}
	for _, p := range parts {
		if p == "" {
			return true
		}
		for _, r := range p {
			if r < '0' || r > '9' {
				return true
			}
		}
	}
	return false
}

// GetFullVersion returns a formatted full version string.
func GetFullVersion() string {
	return fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, Date)
}
