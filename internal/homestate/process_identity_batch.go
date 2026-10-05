package homestate

// process_identity_batch.go — the batched process-identity probe
// (SPEC-CODEX-LANE-SLOTS-001 REQ-013): a listing's owner-classification
// resolves every identity-bearing run through ONE probe invocation instead
// of one probe per run row. The batch reads the SAME platform seam the
// per-pid probe reads — every build-tagged platformProcessFingerprint
// variant is covered by its own platformBatchProcessIdentity — and preserves
// the per-pid classification semantics exactly: the liveness probe runs
// unchanged, and a live pid whose fingerprint cannot be read stays
// indeterminate.

// ProcessIdentity is one pid's probe outcome: the fingerprint and the state
// ProbeProcessIdentity reports for it.
type ProcessIdentity struct {
	Fingerprint string
	State       ProcessIdentityState
}

// BatchProcessIdentityProbe probes many pids through one invocation. The
// result map holds every requested pid; a pid the platform cannot resolve
// still appears with its Dead or Indeterminate state rather than going
// silently missing.
type BatchProcessIdentityProbe func(pids []int) map[int]ProcessIdentity

// BatchProbeProcessIdentity is the production batch probe, built on the
// same platform seam as ProbeProcessIdentity. Duplicate and non-positive
// pids are collapsed before the platform variant sees them.
func BatchProbeProcessIdentity(pids []int) map[int]ProcessIdentity {
	unique := make([]int, 0, len(pids))
	seen := make(map[int]bool, len(pids))
	for _, pid := range pids {
		if pid < 1 || seen[pid] {
			continue
		}
		seen[pid] = true
		unique = append(unique, pid)
	}
	return platformBatchProcessIdentity(unique)
}

// ClassifyOwnerFromResult classifies an owner identity from an
// already-probed result — the batch path's per-row half of ClassifyOwnerWith,
// with the identical decision table: a live pid whose fingerprint matches
// the recorded stamp is live, a drifted fingerprint is dead, a dead pid is
// dead, and every absence of knowledge is indeterminate (REQ-003b).
func ClassifyOwnerFromResult(fingerprint string, state ProcessIdentityState, processStart string) OwnerClassification {
	switch state {
	case ProcessIdentityLive:
		if fingerprint == processStart {
			return OwnerLive
		}
		return OwnerDead
	case ProcessIdentityDead:
		return OwnerDead
	default:
		return OwnerIndeterminate
	}
}
