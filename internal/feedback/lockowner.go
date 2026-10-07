package feedback

// lockowner.go — the owner-verified stale-lock break's identity primitives
// (SPEC-FEEDBACK-PARTICIPATION-001 D37/D40, design.md section 6).
//
// A lock is broken ONLY when its recorded owner is verifiably dead. Two
// verification paths, both conservative:
//
//   - boot identity (currentBootID, platform files): a lock recorded on a
//     DIFFERENT boot than the current one cannot have a live owner — the
//     break is unconditional (even if a live process now carries the
//     recorded pid, it is not the owner);
//   - pid liveness (pidAlive): signal 0 on unix — ESRCH proves death; any
//     other result (alive, or unverifiable) reads as ALIVE.
//
// Where a platform cannot verify (no boot id, no liveness probe), the owner
// reads as live and the lock blocks through the retry budget: the wedge is
// recoverable by the operator, an incorrect break discards a live writer's
// committed mutation, and only the first direction is acceptable under the
// D40 invariant. The platform-specific symbol surface lives in
// lockowner_bootid_*.go — build-tagged, because a runtime GOOS branch would
// still compile the macOS-only syscall symbol on every other platform (the
// B1 defect this split closes).

import (
	"encoding/json"
	"os"
	"runtime"
	"syscall"
)

// pidAlive probes a pid with signal 0 — a liveness check that changes
// nothing. It returns false ONLY when the kernel answers ESRCH (no such
// process); every other outcome reads as alive. On Windows Signal is
// unsupported and always errors, which reads as alive — the conservative
// direction.
func pidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false // no such process on this platform's FindProcess
	}
	if err := p.Signal(syscall.Signal(0)); err != nil {
		if err == syscall.ESRCH {
			return false
		}
		// EPERM (a live process owned by someone else) and any other error
		// read as alive — the conservative direction.
		return true
	}
	return true
}

// lockOwnerIsDead applies the two verification paths.
func lockOwnerIsDead(owner lockOwner) bool {
	if cur := currentBootID(); cur != "" && owner.BootID != "" && owner.BootID != cur {
		// The owner ran on a previous boot: it cannot be alive, whatever
		// lives at that pid today.
		return true
	}
	return !pidAlive(owner.PID)
}

// lockOwnerFromBytes decodes a lock file's owner record.
func lockOwnerFromBytes(raw []byte) (lockOwner, bool) {
	var owner lockOwner
	if err := json.Unmarshal(raw, &owner); err != nil || owner.PID == 0 {
		return lockOwner{}, false
	}
	return owner, true
}

var _ = runtime.GOOS
