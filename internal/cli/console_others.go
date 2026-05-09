//go:build !windows

package cli

// @MX:NOTE: [AUTO] Platform-specific UTF-8 console initialization (non-Windows)
// @MX:NOTE: [AUTO] No-op function - UTF-8 is default on Unix-like systems

// initConsole is a no-op on non-Windows platforms where UTF-8 is the default.
func initConsole() {}
