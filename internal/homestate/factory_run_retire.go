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
}

// Reconciliation reports both halves of a reconciliation pass: what was
// retired, and what was left with each survivor's classification. The second
// half is what the caller renders into the AMBIGUOUS_FACTORY message.
type Reconciliation struct {
	Retired   []RunOwner
	Remaining []RunOwner
}

// ReconcileOptions carries the two seams every retirement path shares: the
// legacy-row identity fallback and the classifier. A zero value uses the
// default classifier and no fallback.
type ReconcileOptions struct {
	Fallback LeadIdentityLookup
	Classify OwnerClassifier
}

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
	classify := opts.classifier()
	var owners []RunOwner
	for rows.Next() {
		var o RunOwner
		if err := rows.Scan(&o.RunID, &o.Status, &o.LeadPID, &o.LeadProcessStart, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		pid, start := o.LeadPID, o.LeadProcessStart
		// lead_pid = 0 is the legacy-row sentinel: consult the run's
		// registered role='lead' peer instead (REQ-006).
		if pid < 1 && opts.Fallback != nil {
			if fpid, fstart, ok := opts.Fallback(o.RunID); ok {
				pid, start = fpid, fstart
			}
		}
		if pid < 1 || strings.TrimSpace(start) == "" {
			// Neither source yields an identity.
			o.Classification = OwnerIndeterminate
		} else {
			o.Classification = classify(pid, start)
		}
		owners = append(owners, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return owners, nil
}

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
		if err := f.retireRun(ctx, o.RunID, o.Classification); err != nil {
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
		if err := f.retireRun(ctx, runID, o.Classification); err != nil {
			return o.Classification, err
		}
		return o.Classification, nil
	}
	return "", fmt.Errorf("factory run %s not found", runID)
}

// retireRun transitions one active run to 'retired' and appends a run.retired
// event. The row is preserved: deleting it would make the retirement itself
// unobservable.
func (f *FactoryDB) retireRun(ctx context.Context, runID string, c OwnerClassification) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	tx, err := f.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `UPDATE runs SET status='retired', updated_at=? WHERE run_id=? AND status='active'`, now, runID); err != nil {
		return err
	}
	payload := fmt.Sprintf(`{"classification":%q}`, string(c))
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
