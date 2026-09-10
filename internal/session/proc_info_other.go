//go:build !linux && !darwin && !freebsd && !netbsd && !openbsd && !dragonfly && !windows

// proc_info_other.go — ancestry lookup is unavailable on platforms without a
// supported process table reader (Windows reads its ancestry via Toolhelp32
// in proc_info_windows.go).
//
// Reporting "unsupported" rather than guessing keeps the resolver on its
// documented fallback: it records os.Getpid(), the pre-existing behavior. The
// twin conservatism still lives in anchor_pid_windows.go, which reports every
// PID alive so the guard fails toward protecting a possibly-live session.
package session

// platformProcInfo always reports unsupported.
func platformProcInfo(pid int) (int, string, bool) {
	_ = pid
	return 0, "", false
}
