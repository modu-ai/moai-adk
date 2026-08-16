//go:build !windows

package cli

// factory_alive_unix.go — the Unix liveness probe behind the factory
// worker-name registry (SPEC-FACTORY-WORKER-FANOUT-001). Signal 0 performs
// the permission check without delivering anything: success (or EPERM, which
// means the process exists but belongs to someone else) reads as live.
// Indeterminate never arises here, so there is no conservative default to
// pick — unlike the Windows twin.

import (
	"errors"
	"syscall"
)

// defaultFactoryProcessAlive reports whether pid names a live process on
// Unix. pid <= 0 is never live.
func defaultFactoryProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}
