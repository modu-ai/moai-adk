//go:build windows

// migrate_agency_signal_windows.go: Windows stub for signal handling.
// Windows does not support SIGINT/SIGTERM in the same way as POSIX.
// @SPEC:SPEC-AGENCY-ABSORB-001:REQ-MIGRATE-013
package cli

import (
	"os"
)

// installSignalHandler is a no-op on Windows.
// Returns a no-op cancel function.
func installSignalHandler(_ *migrationCheckpoint, _ string, _ func(os.Signal)) func() {
	return func() {}
}
