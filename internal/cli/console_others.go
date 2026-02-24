//go:build !windows

package cli

// initConsole is a no-op on non-Windows platforms where UTF-8 is the default.
func initConsole() {}
