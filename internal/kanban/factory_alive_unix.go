//go:build !windows

package kanban

// factory_alive_unix.go — the Unix liveness probe behind the factory
// worker-name registry (SPEC-FACTORY-WORKER-FANOUT-001). Signal 0 performs
// the permission check without delivering anything: success (or EPERM, which
// means the process exists but belongs to someone else) reads as live.
// Indeterminate never arises here, so there is no conservative default to
// pick — unlike the Windows twin.
//
// Moved from internal/cli (t85 lead loop) unchanged in behavior; the cli
// package keeps a package var seam over it.

import (
	"errors"
	"syscall"
)

// FactoryProcessAlive reports whether pid names a live process on Unix.
// pid <= 0 is never live.
func FactoryProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}
