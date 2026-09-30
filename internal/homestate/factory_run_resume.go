package homestate

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ResumeRun is the discovery-path resume writer (SPEC-FACTORY-LANE-JOIN-SOCKET-001
// REQ-004): the ONLY write the lane-join leader discovery performs, and the
// write that makes a verified live leader's run joinable again after its
// record was retired or never written.
//
// It is deliberately a sibling of RecordRun and NOT a reuse of it. RecordRun
// and recordFactoryRunStart stamp the CALLING process; called from a joining
// lane's launcher they would name the lane as the run's owner, and the lane's
// exit would then classify — and retire — a live lead's run. ResumeRun stamps
// the SUPPLIED owner identity, which is the identity the discovery probe
// measured, so the stamp is retire-grade correct at write time: if the leader
// dies afterwards, ReconcileActiveRuns re-retires the row by the same proof
// standard. Self-healing; no new enforcement surface.
//
// In ONE transaction it:
//   - creates the row 'active' when absent (the never-recorded reading of
//     run-record absence);
//   - reactivates the row when retired (or otherwise non-active), overwriting
//     the owner stamp with the verified one;
//   - leaves an already-active row untouched (discovery would not have fired;
//     defensive branch against a TOCTOU race between the discovery read and
//     this write);
//   - appends one run.resumed event recording the verification basis and the
//     outcome, so an operator can audit why the row came back (REQ-004). The
//     kind is distinct from run.started by construction.
//
// basis is the probe-fact record the discovery layer composed (liveness
// fingerprint, leader label, membership evidence, run-id source).
func (f *FactoryDB) ResumeRun(ctx context.Context, runID string, verifiedPID int, verifiedProcessStart, basis string) error {
	if strings.TrimSpace(runID) == "" {
		return errors.New("factory run id is empty")
	}
	if verifiedPID < 1 || strings.TrimSpace(verifiedProcessStart) == "" {
		return errors.New("factory run resume requires a verified owner identity")
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	tx, err := f.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	// The WHERE on the conflict branch is the already-active no-op: a row
	// that resolved active between the discovery read and this write keeps
	// its owner stamp, and RowsAffected reports the outcome for the event.
	res, err := tx.ExecContext(ctx, `INSERT INTO runs(run_id,lead_session_id,lead_backend,status,manifest_json,lead_pid,lead_process_start,created_at,updated_at)
VALUES(?,?,?,'active','{}',?,?,?,?) ON CONFLICT(run_id) DO UPDATE SET
status='active',lead_pid=excluded.lead_pid,lead_process_start=excluded.lead_process_start,updated_at=excluded.updated_at WHERE runs.status!='active'`,
		runID, "", "", verifiedPID, verifiedProcessStart, now, now)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	outcome := "resumed"
	if affected == 0 {
		outcome = "already-active"
	}
	payload := fmt.Sprintf(`{"basis":%q,"outcome":%q}`, basis, outcome)
	if _, err := tx.ExecContext(ctx, `INSERT INTO events(run_id,kind,payload_json,created_at) VALUES(?,'run.resumed',?,?)`, runID, payload, now); err != nil {
		return err
	}
	return tx.Commit()
}
