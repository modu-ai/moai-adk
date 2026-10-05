//go:build darwin

package homestate

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func platformProcessFingerprint(pid int) (string, bool) {
	if pid <= 0 {
		return "", false
	}
	info, err := unix.SysctlKinfoProc("kern.proc.pid", pid)
	if err != nil || info == nil || int(info.Proc.P_pid) != pid {
		return "", false
	}
	started := info.Proc.P_starttime
	return fmt.Sprintf("%d.%06d", started.Sec, started.Usec), true
}

// platformBatchProcessIdentity probes many pids through one invocation
// (SPEC-CODEX-LANE-SLOTS-001 REQ-013). The liveness probe is unchanged —
// deriving liveness from the kinfo read alone would flip a non-signalable
// process (kill(0) EPERM) from indeterminate to a decision, and a wrong dead
// verdict retires a live run — and each live pid's fingerprint comes from
// the same SysctlKinfoProc seam the per-pid probe reads.
func platformBatchProcessIdentity(pids []int) map[int]ProcessIdentity {
	results := make(map[int]ProcessIdentity, len(pids))
	for _, pid := range pids {
		state := platformPIDState(pid)
		if state != ProcessIdentityLive {
			results[pid] = ProcessIdentity{State: state}
			continue
		}
		fingerprint, ok := platformProcessFingerprint(pid)
		if !ok {
			results[pid] = ProcessIdentity{State: ProcessIdentityIndeterminate}
			continue
		}
		results[pid] = ProcessIdentity{Fingerprint: fingerprint, State: ProcessIdentityLive}
	}
	return results
}
