//go:build windows

// Package cli — update_cleanup_windows.go
//
// Windows-specific process-liveness probe for SPEC-V3R3-UPDATE-CLEANUP-001
// stale-lock detection.
//
// SPEC-V2.20.0-RC1 hotfix: extracted from update_cleanup.go to fix
// `syscall.Kill undefined` compile error on the windows/amd64 target.
// Conservative implementation: always returns true to avoid false-positive
// stale-lock detection on Windows. Proper Windows-native PID validation
// (OpenProcess + GetExitCodeProcess) is deferred — see follow-up SPEC.

package cli

// isProcessAlive is conservative on Windows: always returns true to avoid
// false-positive stale-lock detection. A subsequent SPEC will introduce
// proper Windows-native PID validation via OpenProcess/GetExitCodeProcess.
func isProcessAlive(pid int) bool {
	_ = pid
	return true
}
