//go:build !linux && !darwin && !freebsd && !netbsd && !openbsd && !dragonfly

// proc_info_other.go — ancestry lookup is unavailable on this platform
// (Windows, and anything else without a supported process table reader).
//
// Reporting "unsupported" rather than guessing keeps the resolver on its
// documented fallback: it records os.Getpid(), the pre-existing behavior. The
// twin conservatism already lives in anchor_pid_windows.go, which reports every
// PID alive so the guard fails toward protecting a possibly-live session.
package session

// platformProcInfo always reports unsupported.
func platformProcInfo(pid int) (int, string, bool) {
	_ = pid
	return 0, "", false
}
