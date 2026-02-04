// Package version provides build-time version information for MoAI-ADK.
//
// The version information is injected at build time via ldflags:
//
//	go build -ldflags "-X github.com/modu-ai/moai-adk/pkg/version.Version=1.0.0 \
//	  -X github.com/modu-ai/moai-adk/pkg/version.Commit=abc123 \
//	  -X github.com/modu-ai/moai-adk/pkg/version.Date=2026-01-01"
//
// # Variables
//
// The package exports three variables that contain version metadata:
//   - Version: The semantic version string (e.g., "1.14.0")
//   - Commit: The git commit hash (e.g., "abc123def")
//   - Date: The build date in ISO format (e.g., "2026-01-01")
//
// # Functions
//
// Helper functions are provided to safely retrieve version information:
//   - GetVersion(): Returns the version string or "dev" if not set
//   - GetCommit(): Returns the commit hash or "unknown" if not set
//   - GetDate(): Returns the build date or "unknown" if not set
//   - GetFullVersion(): Returns a formatted string with all version info
//
// # Usage
//
//	import "github.com/modu-ai/moai-adk/pkg/version"
//
//	func main() {
//	    fmt.Println(version.GetFullVersion())
//	    // Output: moai v1.14.0 (abc123def) built on 2026-01-01
//	}
package version
