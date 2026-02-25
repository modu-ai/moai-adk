//go:build !windows

package cli

import "testing"

// TestInitConsoleNoop verifies initConsole does not panic on non-Windows platforms.
func TestInitConsoleNoop(t *testing.T) {
	t.Parallel()
	// On non-Windows platforms initConsole is a no-op.
	// This test ensures the function exists and is callable without side effects.
	initConsole()
}
