package homestate

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// OwnerClassification is the answer to "is this run's owner still running?".
//
// The three values below are the ones the probe produces today. The set is
// deliberately NOT closed: retirement is gated on a positive OwnerDead
// (see retirable), so a value added later declines retirement by default
// rather than falling through to it.
type OwnerClassification string

const (
	OwnerLive          OwnerClassification = "live"
	OwnerDead          OwnerClassification = "dead"
	OwnerIndeterminate OwnerClassification = "indeterminate"
)

// ProofBasis names which proof established a run's classification (REQ-010b).
// It is written into every run.retired event so an operator can tell a probed
// death from the boot proof, which infers death from time alone.
type ProofBasis string

const (
	// BasisStamp — the run row's own owner stamp was probed.
	BasisStamp ProofBasis = "stamp"
	// BasisPeer — the REQ-006 role='lead' peer identity was probed.
	BasisPeer ProofBasis = "peer"
	// BasisBoot — the REQ-006b boot proof held; no process was probed.
	BasisBoot ProofBasis = "boot"
)

// ErrRunOwnerNotDead reports a refused retirement: the named run's owner did
// not classify dead, so no retirement path may touch it.
var ErrRunOwnerNotDead = errors.New("factory run owner is not dead")

// OwnerClassifier maps a recorded owner identity to a classification.
type OwnerClassifier func(pid int, processStart string) OwnerClassification

// LeadIdentityLookup resolves a run's registered role='lead' peer identity.
// It is a function parameter rather than an import because the peer lives in a
// factorymsg broker database and factorymsg imports homestate — the direction
// is fixed and Go will not permit it to be closed.
type LeadIdentityLookup func(runID string) (pid int, processStart string, ok bool)

// RunOwner is one run row together with its owner classification.
type RunOwner struct {
	RunID            string
	Status           string
	LeadPID          int
	LeadProcessStart string
	CreatedAt        string
	UpdatedAt        string
	Classification   OwnerClassification
	// Basis names the proof behind Classification; it is empty when no
	// identity was found and the boot proof did not hold.
	Basis ProofBasis
}

// Reconciliation reports both halves of a reconciliation pass: what was
// retired, and what was left with each survivor's classification. The second
// half is what the caller renders into the AMBIGUOUS_FACTORY message.
type Reconciliation struct {
	Retired   []RunOwner
	Remaining []RunOwner
}

// ReconcileOptions carries the seams every retirement path shares: the
// legacy-row identity fallback, the classifier, and the two premises of the
// boot proof. A zero value uses the default classifier, no fallback, and no
// boot proof.
type ReconcileOptions struct {
	Fallback LeadIdentityLookup
	Classify OwnerClassifier
	// BootTime reports when the host last booted. Nil, or a false second
	// result, disables the boot proof.
	BootTime func() (time.Time, bool)
	// LeadRecordAbsent reports that the run has no lead-peer record at all —
	// not merely one the Fallback failed to read. Nil disables the boot proof.
	LeadRecordAbsent func(runID string) bool
}

// SystemBootTime reports when this host last booted, or false where the
// platform cannot say.
func SystemBootTime() (time.Time, bool) { return platformBootTime() }

func (o ReconcileOptions) classifier() OwnerClassifier {
	if o.Classify != nil {
		return o.Classify
	}
	return DefaultOwnerClassifier
}

// DefaultOwnerClassifier classifies an owner identity with the process probe
// already in service for this purpose.
func DefaultOwnerClassifier(pid int, processStart string) OwnerClassification {
	return ClassifyOwnerWith(ProbeProcessIdentity, pid, processStart)
}

// ClassifyOwnerWith classifies an owner identity using the supplied probe.
//
// The pid is consulted together with the process-start fingerprint, because
// process ids are reused: a live pid whose fingerprint differs from the
// recorded one is a different process, and the run's owner is dead.
//
// Every absence of knowledge resolves toward live (REQ-003b). The two errors
// are not symmetric — a stale run that survives has a working operator escape
// (--factory-run), a live run that is retired has none.
func ClassifyOwnerWith(probe func(int) (string, ProcessIdentityState), pid int, processStart string) OwnerClassification {
	if pid < 1 || strings.TrimSpace(processStart) == "" {
		return OwnerIndeterminate
	}
	fingerprint, state := probe(pid)
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

// retirable is the single retirement gate, shared by every path: the
// resolution-time reconciler, the legacy-row migration pass, and the operator
// command alike.
//
// It is stated positively on purpose. A reject-list form — "refuse live,
// refuse indeterminate" — is an enumeration of the classification set alive
// when it was written, and it silently begins retiring any value added later.
// The positive form fails safe by construction.
func retirable(c OwnerClassification) bool { return c == OwnerDead }

// ClassifyRuns reports every run row with its owner classification.
func (f *FactoryDB) ClassifyRuns(ctx context.Context, opts ReconcileOptions) ([]RunOwner, error) {
	return f.classifyRuns(ctx, opts, "")
}

func (f *FactoryDB) classifyRuns(ctx context.Context, opts ReconcileOptions, status string) (_ []RunOwner, err error) {
	query := `SELECT run_id,status,lead_pid,lead_process_start,created_at,updated_at FROM runs ORDER BY run_id`
	args := []any{}
	if status != "" {
		query = `SELECT run_id,status,lead_pid,lead_process_start,created_at,updated_at FROM runs WHERE status=? ORDER BY run_id`
		args = append(args, status)
	}
	rows, err := f.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	var owners []RunOwner
	for rows.Next() {
		var o RunOwner
		if err := rows.Scan(&o.RunID, &o.Status, &o.LeadPID, &o.LeadProcessStart, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		owners = append(owners, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Classification runs after the row cursor is drained: the boot proof
	// issues its own queries, and the pool holds a single connection.
	classify := opts.classifier()
	for i := range owners {
		o := &owners[i]
		pid, start, basis := o.LeadPID, o.LeadProcessStart, BasisStamp
		// lead_pid = 0 is the legacy-row sentinel: consult the run's
		// registered role='lead' peer instead (REQ-006).
		if pid < 1 && opts.Fallback != nil {
			if fpid, fstart, ok := opts.Fallback(o.RunID); ok {
				pid, start, basis = fpid, fstart, BasisPeer
			}
		}
		if pid >= 1 && strings.TrimSpace(start) != "" {
			o.Classification = classify(pid, start)
			o.Basis = basis
			continue
		}
		// Neither source yields an identity.
		o.Classification = OwnerIndeterminate
		dead, err := f.predatesBoot(ctx, opts, o.RunID)
		if err != nil {
			return nil, err
		}
		if dead {
			o.Classification = OwnerDead
			o.Basis = BasisBoot
		}
	}
	return owners, nil
}

// predatesBoot is the boot proof for a run with no owner identity: when the
// run has no lead-peer record and every timestamp recorded for it — the run
// row, its events, cards, workers, and dead letters — is strictly earlier than
// the host's last boot, whatever process owned it existed before that boot
// and cannot be alive now. It is a positive proof of death, not an inference
// from age: a run with any activity after boot, an unreadable timestamp, an
// unknown boot time, or a possible lead record is not proven and stays
// indeterminate (REQ-003b).
//
// @MX:WARN: [AUTO] retires a run with no identity at all; every premise must hold or the row stays indeterminate
// @MX:REASON: a wrong dead verdict retires a live lead's run and has no operator undo (REQ-005)
func (f *FactoryDB) predatesBoot(ctx context.Context, opts ReconcileOptions, runID string) (_ bool, err error) {
	if opts.BootTime == nil || opts.LeadRecordAbsent == nil {
		return false, nil
	}
	boot, ok := opts.BootTime()
	if !ok || boot.IsZero() || !opts.LeadRecordAbsent(runID) {
		return false, nil
	}
	rows, err := f.DB.QueryContext(ctx, runActivityQuery, runID)
	if err != nil {
		return false, err
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	seen := false
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return false, err
		}
		at, perr := time.Parse(time.RFC3339Nano, raw)
		if perr != nil || !at.Before(boot) {
			return false, nil
		}
		seen = true
	}
	if err := rows.Err(); err != nil {
		return false, err
	}
	return seen, nil
}

// runActivityQuery lists every timestamp the factory database records for one
// run. The boot proof requires all of them to predate boot.
const runActivityQuery = `
SELECT created_at FROM runs WHERE run_id=?1
UNION ALL SELECT updated_at FROM runs WHERE run_id=?1
UNION ALL SELECT created_at FROM events WHERE run_id=?1
UNION ALL SELECT updated_at FROM cards WHERE run_id=?1
UNION ALL SELECT registered_at FROM workers WHERE run_id=?1
UNION ALL SELECT heartbeat_at FROM workers WHERE run_id=?1
UNION ALL SELECT created_at FROM dead_letters WHERE run_id=?1`

// ReconcileActiveRuns retires every active run whose owner is classified dead
// and reports what remains. It only ever reduces the active set — it never
// selects among survivors, so fail-closed resolution is preserved.
func (f *FactoryDB) ReconcileActiveRuns(ctx context.Context, opts ReconcileOptions) (Reconciliation, error) {
	owners, err := f.classifyRuns(ctx, opts, "active")
	if err != nil {
		return Reconciliation{}, err
	}
	result := Reconciliation{}
	for _, o := range owners {
		if !retirable(o.Classification) {
			continue
		}
		if err := f.retireRun(ctx, o.RunID, o.Classification, o.Basis); err != nil {
			return Reconciliation{}, err
		}
		o.Status = "retired"
		result.Retired = append(result.Retired, o)
	}
	// Re-query so the survivors reported are the rows the database actually
	// still holds as active, not an in-memory prediction of them.
	remaining, err := f.classifyRuns(ctx, opts, "active")
	if err != nil {
		return Reconciliation{}, err
	}
	result.Remaining = remaining
	return result, nil
}

// RetireRunIfDead is the operator entry to the same mechanism. It shares the
// reconciler's predicate rather than carrying a second, laxer copy of it, and
// returns the classification it observed so the caller can name it as the
// refusal reason.
func (f *FactoryDB) RetireRunIfDead(ctx context.Context, runID string, opts ReconcileOptions) (OwnerClassification, error) {
	owners, err := f.ClassifyRuns(ctx, opts)
	if err != nil {
		return "", err
	}
	for _, o := range owners {
		if o.RunID != runID {
			continue
		}
		if !retirable(o.Classification) {
			return o.Classification, fmt.Errorf("%w: run %s owner classified %s", ErrRunOwnerNotDead, runID, o.Classification)
		}
		if err := f.retireRun(ctx, runID, o.Classification, o.Basis); err != nil {
			return o.Classification, err
		}
		return o.Classification, nil
	}
	return "", fmt.Errorf("factory run %s not found", runID)
}

// retireRun transitions one active run to 'retired' and appends a run.retired
// event carrying the classification and the proof basis that established it
// (REQ-010b). The row is preserved: deleting it would make the retirement
// itself unobservable.
func (f *FactoryDB) retireRun(ctx context.Context, runID string, c OwnerClassification, basis ProofBasis) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	tx, err := f.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `UPDATE runs SET status='retired', updated_at=? WHERE run_id=? AND status='active'`, now, runID); err != nil {
		return err
	}
	payload := fmt.Sprintf(`{"classification":%q,"basis":%q}`, string(c), string(basis))
	if _, err := tx.ExecContext(ctx, `INSERT INTO events(run_id,kind,payload_json,created_at) VALUES(?,'run.retired',?,?)`, runID, payload, now); err != nil {
		return err
	}
	return tx.Commit()
}

// StampRunOwner rewrites one run's owner identity. It is the REQ-002b restamp:
// on a door that does not replace its launching process, the record-time stamp
// names the launcher and this call corrects it to the session process.
func (f *FactoryDB) StampRunOwner(ctx context.Context, runID string, pid int, processStart string) error {
	if strings.TrimSpace(runID) == "" {
		return errors.New("factory run id is empty")
	}
	if pid < 1 || strings.TrimSpace(processStart) == "" {
		return errors.New("factory run owner identity is empty")
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := f.DB.ExecContext(ctx, `UPDATE runs SET lead_pid=?, lead_process_start=?, updated_at=? WHERE run_id=?`, pid, processStart, now, runID)
	return err
}

// ClearRunOwner removes a run's owner identity. It is the run-state half of
// the REQ-002d refusal: when a non-replace door cannot obtain a session
// identity it refuses the launch, and the run must not be left carrying the
// launching process's pid — that identity is known in advance to die.
func (f *FactoryDB) ClearRunOwner(ctx context.Context, runID string) error {
	if strings.TrimSpace(runID) == "" {
		return errors.New("factory run id is empty")
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := f.DB.ExecContext(ctx, `UPDATE runs SET lead_pid=0, lead_process_start='', updated_at=? WHERE run_id=?`, now, runID)
	return err
}
