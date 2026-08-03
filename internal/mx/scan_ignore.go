package mx

// DefaultScanIgnore lists directory base-names excluded from a full project
// scan. The scanner matches these against filepath.Base(path) for directories
// (scanner.go ScanDir), so these are plain names — no globbing across path
// separators is needed or supported there.
//
// This is the single source of truth shared by 'moai mx scan'
// (internal/cli/mx_scan.go) and the SessionStart-hook cold-start background
// scan (internal/hook/session_start.go). Both paths MUST agree on the ignore
// set so a hook-built index matches a manually-scanned one byte-for-byte in
// scope. Extracted here to avoid duplicating the list across packages
// (internal/hook cannot import internal/cli — import cycle).
var DefaultScanIgnore = []string{
	".git",
	"node_modules",
	"vendor",
	"worktrees", // harness worktrees (recursive copies of the repo)
	"main-fork", // moai-adk local fork (not distributed source)
	"target",    // rust/cargo build output
	"build",     // java/gradle/kotlin build output
	"dist",      // generic build output
	"out",
	".next", // next.js build cache
	".cache",
	".turbo",
	"coverage", // test coverage reports
	// Harness / state directories hold no user source code.
	".claude",
	".moai",
}
