package atomicfile

// owner.go — the owner-verified stale-lock break's identity primitives
// (SPEC-FEEDBACK-PARTICIPATION-001 D37/D40, design.md section 6).
//
// A lock is broken ONLY when its recorded owner is verifiably dead. Two
// verification paths, both conservative:
//
//   - boot identity (currentBootID, platform files): a lock recorded on a
//     DIFFERENT boot than the current one cannot have a live owner — the
//     break is unconditional (even if a live process now carries the
//     recorded pid, it is not the owner);
//   - pid liveness (PidAlive): signal 0 on unix — ESRCH proves death; any
//     other result (alive, or unverifiable) reads as ALIVE.
//
// Where a platform cannot verify (no boot id, no liveness probe), the owner
// reads as live and the lock blocks through the retry budget: the wedge is
// recoverable by the operator, an incorrect break discards a live writer's
// committed mutation, and only the first direction is acceptable under the
// D40 invariant. The platform-specific symbol surface lives in
// bootid_*.go — build-tagged, because a runtime GOOS branch would
// still compile the macOS-only syscall symbol on every other platform (the
// B1 defect this split closes).
//
// The machinery moved here from internal/feedback (which still owns the
// queue lock) so the capture spool's section lock — whose package must not
// import internal/feedback — can share the SAME owner-verification logic
// instead of a drifting copy.

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"runtime"
	"syscall"
)

// LockOwner is what a section lock file records: the owner's process id,
// the boot the owner ran on, and when the lock was taken. The boot-unique
// id is what makes a stale-break SAFE: a pid alone can be recycled across a
// reboot, so a previous boot's lock breaks even when some live process now
// carries that pid (it is not the owner), and a same-boot lock breaks only
// when the pid is verifiably dead.
type LockOwner struct {
	PID       int    `json:"pid"`
	BootID    string `json:"boot_id"`
	CreatedAt string `json:"created_at"`
}

// PidAlive probes a pid with signal 0 — a liveness check that changes
// nothing. It returns false when the process verifiably does not exist:
// ESRCH from the kernel, or os.ErrProcessDone — a process this machine's
// runtime already observed exiting is finished, and Signal surfaces that as
// ErrProcessDone rather than ESRCH. Reading ErrProcessDone as alive judged
// every finished same-boot child alive, so an orphan lock whose owner
// exited was never reclaimed. Every OTHER outcome (EPERM from a live
// process owned by someone else, a platform without a liveness probe)
// reads as alive — the conservative direction: an incorrect "dead" breaks a
// lock and discards a live writer's committed mutation, an incorrect
// "alive" only delays.
func PidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false // no such process on this platform's FindProcess
	}
	if err := p.Signal(syscall.Signal(0)); err != nil {
		if errors.Is(err, os.ErrProcessDone) || err == syscall.ESRCH {
			return false
		}
		// EPERM (a live process owned by someone else) and any other error
		// read as alive — the conservative direction.
		return true
	}
	return true
}

// OwnerIsDead applies the two verification paths.
func OwnerIsDead(owner LockOwner) bool {
	if cur := BootIDIdentity(); cur != "" && owner.BootID != "" && owner.BootID != cur {
		// The owner ran on a previous boot: it cannot be alive, whatever
		// lives at that pid today.
		return true
	}
	return !PidAlive(owner.PID)
}

// BootIDIdentity is the lock record's wire form of the boot identity: hex
// of currentBootID's raw bytes. The raw form is BINARY on the BSD family
// (kern.boottime is a struct timeval), and JSON serialization of binary
// bytes mangles invalid UTF-8 into replacement runes — a mangled record
// then compares unequal to the live machine's identity and a LIVE
// same-boot owner is misjudged as a previous boot, firing the break
// mid-claim (the lost-update the D37 repair exists to close). Hex is
// lossless over any byte string, so the record round-trips and the
// comparison is exact.
func BootIDIdentity() string {
	raw := currentBootID()
	if raw == "" {
		return ""
	}
	return hex.EncodeToString([]byte(raw))
}

// OwnerFromBytes decodes a lock file's owner record.
func OwnerFromBytes(raw []byte) (LockOwner, bool) {
	var owner LockOwner
	if err := json.Unmarshal(raw, &owner); err != nil || owner.PID == 0 {
		return LockOwner{}, false
	}
	return owner, true
}

var _ = runtime.GOOS
