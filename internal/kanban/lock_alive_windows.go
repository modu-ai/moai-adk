//go:build windows

// lock_alive_windows.go — the default liveness probe for the stale-lock clear
// (REQ-KB-023) on Windows. Windows process-existence probing without opening
// a handle is best-effort: os.FindProcess succeeds without validating, and
// Signal(0) reports only permission/existence errors the kernel surfaces.
// A false "absent" here cannot silently remove a live lock, because the
// Windows substrate's lock artifact is itself exclusive — the clear's unlink
// would only follow this probe when the artifact exists, and the pre-removal
// re-read still guards a changed identity.
package kanban

import (
	"os"
	"syscall"
)

// defaultProcessAlive reports whether pid names a live process on this host,
// best-effort on Windows.
func defaultProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}
