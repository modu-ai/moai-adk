//go:build !windows && !darwin

package homestate

import (
	"os/exec"
	"strconv"
	"strings"
)

func platformProcessFingerprint(pid int) (string, bool) {
	out, err := exec.Command("ps", "-o", "lstart=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return "", false
	}
	value := strings.TrimSpace(string(out))
	return value, value != ""
}

// platformBatchProcessIdentity probes many pids through one invocation
// (SPEC-CODEX-LANE-SLOTS-001 REQ-013): the liveness probe is unchanged per
// pid, and the live pids' lstart values are read through ONE `ps` subprocess
// instead of one per pid. A pid the batch read cannot resolve falls back to
// the per-pid fingerprint probe, so the classification semantics stay
// exactly the per-pid path's.
func platformBatchProcessIdentity(pids []int) map[int]ProcessIdentity {
	results := make(map[int]ProcessIdentity, len(pids))
	var live []int
	for _, pid := range pids {
		state := platformPIDState(pid)
		if state != ProcessIdentityLive {
			results[pid] = ProcessIdentity{State: state}
			continue
		}
		live = append(live, pid)
	}
	if len(live) == 0 {
		return results
	}
	fingerprints := platformBatchFingerprints(live)
	for _, pid := range live {
		if value := fingerprints[pid]; value != "" {
			results[pid] = ProcessIdentity{Fingerprint: value, State: ProcessIdentityLive}
			continue
		}
		// Missing from the batch read (a process that died mid-listing, or a
		// ps that refused the batch form): the per-pid probe decides.
		fingerprint, ok := platformProcessFingerprint(pid)
		if !ok {
			results[pid] = ProcessIdentity{State: ProcessIdentityIndeterminate}
			continue
		}
		results[pid] = ProcessIdentity{Fingerprint: fingerprint, State: ProcessIdentityLive}
	}
	return results
}

// platformBatchFingerprints reads every listed pid's lstart through one
// `ps -o pid=,lstart=` invocation. The map may hold fewer pids than listed —
// a pid that exited between the liveness probe and this read prints no line.
func platformBatchFingerprints(pids []int) map[int]string {
	nums := make([]string, len(pids))
	for i, pid := range pids {
		nums[i] = strconv.Itoa(pid)
	}
	out, _ := exec.Command("ps", "-o", "pid=,lstart=", "-p", strings.Join(nums, ",")).Output()
	results := make(map[int]string, len(pids))
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.SplitN(strings.TrimSpace(line), " ", 2)
		if len(fields) != 2 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		results[pid] = strings.TrimSpace(fields[1])
	}
	return results
}
