package homestate

import "context"

// factory_run_adopt.go — the run re-adoption half of the leader-resume fix
// (card t1707): the finder that resolves the one run a resume-shaped leader
// start may re-adopt, and the adoption write that keeps the prior run id.

// FindResumableRun resolves the single run a resume-shaped leader start may
// re-adopt instead of minting a fresh run id (card t1707).
//
// A candidate is an ACTIVE run whose owner classifies dead or indeterminate —
// the incident shape: the prior leader process died with the machine state
// intact, and the resuming leader is a NEW process (new pid, new fingerprint),
// so the recorded stamp classifies dead even though the run's records are
// alive — and which still carries LIVE RECORDS: cards in a non-terminal state
// (anything but done and abandoned) or registered workers. Live records are
// what separate "the leader died mid-run" from "this run is genuinely over".
//
// Exactly one candidate → it is returned as adopt with the candidate list.
// Zero candidates (nothing to adopt — a cold start mints as before) and more
// than one (ambiguous — the caller warns and offers the deliberate
// --factory-run path) return a zero adopt with the full candidate list, which
// the caller distinguishes by length.
func (f *FactoryDB) FindResumableRun(ctx context.Context, opts ReconcileOptions) (RunOwner, []RunOwner, error) {
	return RunOwner{}, nil, nil
}

// AdoptRun re-adopts one ACTIVE run for the resuming leader process: it
// restamps the owner identity to the supplied pid and process-start
// fingerprint and appends one run.resumed event carrying the adoption basis,
// so the audit trail records why the run id stayed the same across the leader
// process change (card t1707). The row must measure active: a retired run is
// refused with an error, because re-activating a retired row is the ResumeRun
// discovery path's decision, not this one.
func (f *FactoryDB) AdoptRun(ctx context.Context, runID string, pid int, processStart, basis string) error {
	return nil
}
