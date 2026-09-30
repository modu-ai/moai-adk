//go:build windows

package homestate

import (
	"fmt"
	"golang.org/x/sys/windows"
)

func platformProcessFingerprint(pid int) (string, bool) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return "", false
	}
	defer windows.CloseHandle(h)
	var created, exited, kernel, user windows.Filetime
	if err := windows.GetProcessTimes(h, &created, &exited, &kernel, &user); err != nil {
		return "", false
	}
	return fmt.Sprintf("%d", created.Nanoseconds()), true
}

// platformBatchProcessIdentity probes many pids through one invocation
// (SPEC-CODEX-LANE-SLOTS-001 REQ-013). Each pid costs ONE OpenProcess call
// instead of the per-pid path's two — the liveness state and the
// creation-time fingerprint come from the same handle — with the identical
// decision table: ERROR_INVALID_PARAMETER is dead, any other open failure
// indeterminate, a failed times read indeterminate.
func platformBatchProcessIdentity(pids []int) map[int]ProcessIdentity {
	results := make(map[int]ProcessIdentity, len(pids))
	for _, pid := range pids {
		h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
		if err == windows.ERROR_INVALID_PARAMETER {
			results[pid] = ProcessIdentity{State: ProcessIdentityDead}
			continue
		}
		if err != nil {
			results[pid] = ProcessIdentity{State: ProcessIdentityIndeterminate}
			continue
		}
		var created, exited, kernel, user windows.Filetime
		if timesErr := windows.GetProcessTimes(h, &created, &exited, &kernel, &user); timesErr != nil {
			windows.CloseHandle(h)
			results[pid] = ProcessIdentity{State: ProcessIdentityIndeterminate}
			continue
		}
		windows.CloseHandle(h)
		results[pid] = ProcessIdentity{Fingerprint: fmt.Sprintf("%d", created.Nanoseconds()), State: ProcessIdentityLive}
	}
	return results
}
