// Package version provides version information for MoAI-ADK
//
// These variables are set via ldflags during the build process:
//
//	-X github.com/anthropics/moai-adk-go/pkg/version.Version={{.Version}}
//	-X github.com/anthropics/moai-adk-go/pkg/version.Commit={{.Commit}}
//	-X github.com/anthropics/moai-adk-go/pkg/version.Date={{.Date}}
package version

// Version is the current version of MoAI-ADK
// Set by goreleaser during build
var Version = "dev"

// Commit is the git commit hash
// Set by goreleaser during build
var Commit = "unknown"

// Date is the build date in ISO 8601 format
// Set by goreleaser during build
var Date = "unknown"

// GetVersion returns the full version string
func GetVersion() string {
	if Version == "dev" {
		return "dev"
	}
	return Version
}

// GetVersionInfo returns detailed version information
func GetVersionInfo() map[string]string {
	return map[string]string{
		"version": GetVersion(),
		"commit":  Commit,
		"date":    Date,
	}
}
