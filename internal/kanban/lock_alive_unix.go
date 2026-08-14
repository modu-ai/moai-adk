//go:build !windows

// lock_alive_unix.go — the default liveness probe for the stale-lock clear
// (REQ-KB-023): signal 0 performs no signal delivery and reports only
// whether the process exists (or is visible but owned by another user).
package kanban

import "golang.org/x/sys/unix"

// defaultProcessAlive reports whether pid names a live process on this host.
func defaultProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := unix.Kill(pid, 0)
	return err == nil || err == unix.EPERM
}
