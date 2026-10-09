package factory

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TodoRuntime is the latest execution report, not card completion authority.
type TodoRuntime struct {
	Runs        []TodoRuntimeRun        `json:"runs"`
	Assignments []TodoRuntimeAssignment `json:"assignments"`
	// DispatchCurrent is the queue's own record of each card's CURRENT
	// dispatch engagement — the (run, owner) identity the last dispatch
	// assignment carried (review round-23, card t1538). Written only by the
	// hook-bearing assignment path (RecordFactoryCardAssignment), in the
	// same transaction as the assignment and the factory binding hook, so
	// the reconciliation adjudicates supersession by IDENTITY — what the
	// queue names — instead of deriving engagement order from row history.
	// omitempty: queues written before the record existed carry no key and
	// marshal byte-identically to before.
	DispatchCurrent []TodoDispatchCurrent `json:"dispatch_current,omitempty"`
}

type TodoRuntimeRun struct {
	ProjectUUID  *string `json:"project_uuid"`
	RunID        string  `json:"run_id"`
	Backend      string  `json:"backend"`
	ManifestJSON string  `json:"manifest_json"`
}

type TodoRuntimeAssignment struct {
	ProjectUUID    *string `json:"project_uuid"`
	CardUUID       *string `json:"card_uuid"`
	RunID          string  `json:"run_id"`
	CardID         string  `json:"card_id"`
	OwnerLabel     string  `json:"owner_label"`
	ReportedState  string  `json:"reported_state"`
	EventKind      string  `json:"event_kind"`
	ProvenanceJSON string  `json:"provenance_json"`
}

// TodoDispatchCurrent is one card's current dispatch engagement per the
// queue's own record (review round-23, card t1538).
type TodoDispatchCurrent struct {
	CardID     string `json:"card_id"`
	RunID      string `json:"run_id"`
	OwnerLabel string `json:"owner_label"`
}

// MarshalJSON keeps old/absent queues compatible with the additive array contract.
func (r TodoRuntime) MarshalJSON() ([]byte, error) {
	type wire TodoRuntime
	if r.Runs == nil {
		r.Runs = []TodoRuntimeRun{}
	}
	if r.Assignments == nil {
		r.Assignments = []TodoRuntimeAssignment{}
	}
	return json.Marshal(wire(r))
}

const todoRuntimeDDL = `
CREATE TABLE IF NOT EXISTS todo_runtime_runs (
 run_id TEXT PRIMARY KEY, backend TEXT NOT NULL, manifest_json TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS todo_runtime_assignments (
 run_id TEXT NOT NULL, card_id TEXT NOT NULL, owner_label TEXT NOT NULL,
 reported_state TEXT NOT NULL, event_kind TEXT NOT NULL, provenance_json TEXT NOT NULL,
 PRIMARY KEY(run_id, card_id)
);
CREATE TABLE IF NOT EXISTS todo_dispatch_current (
 card_id TEXT PRIMARY KEY, run_id TEXT NOT NULL, owner_label TEXT NOT NULL
);
INSERT INTO meta(key,value) VALUES('runtime_schema_version','1')
 ON CONFLICT(key) DO NOTHING;
`

func (e *backlogEngine) runtimeVersion(ctx context.Context) (bool, error) {
	var version string
	err := e.queryDB().QueryRowContext(ctx, `SELECT value FROM meta WHERE key='runtime_schema_version'`).Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		var tables int
		if err := e.queryDB().QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE name IN ('todo_runtime_runs','todo_runtime_assignments')`).Scan(&tables); err != nil {
			return false, err
		}
		if tables != 0 {
			return false, fmt.Errorf("unstamped partial runtime schema: %w", ErrBacklogCorrupt)
		}
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if version != "1" {
		return false, fmt.Errorf("unsupported runtime_schema_version %q: %w", version, ErrBacklogCorrupt)
	}
	var tables int
	if err := e.queryDB().QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type='table' AND name IN ('todo_runtime_runs','todo_runtime_assignments')`).Scan(&tables); err != nil {
		return false, err
	}
	if tables != 2 {
		return false, fmt.Errorf("incomplete stamped runtime schema: %w", ErrBacklogCorrupt)
	}
	return true, nil
}

func (e *backlogEngine) readRuntime(ctx context.Context, result *TodoRuntime) error {
	present, err := e.runtimeVersion(ctx)
	if err != nil || !present {
		return err
	}
	rows, err := e.queryDB().QueryContext(ctx, `SELECT run_id,backend,manifest_json FROM todo_runtime_runs ORDER BY run_id`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var run TodoRuntimeRun
		if err := rows.Scan(&run.RunID, &run.Backend, &run.ManifestJSON); err != nil {
			_ = rows.Close()
			return err
		}
		result.Runs = append(result.Runs, run)
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return err
	}
	rows, err = e.queryDB().QueryContext(ctx, `SELECT run_id,card_id,owner_label,reported_state,event_kind,provenance_json FROM todo_runtime_assignments ORDER BY run_id,card_id`)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var a TodoRuntimeAssignment
		if err := rows.Scan(&a.RunID, &a.CardID, &a.OwnerLabel, &a.ReportedState, &a.EventKind, &a.ProvenanceJSON); err != nil {
			return err
		}
		result.Assignments = append(result.Assignments, a)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	// The dispatch-identity table is additive (review round-23): a queue
	// stamped at runtime schema version 1 carries the two runtime tables
	// only until its next dispatch write creates this one, so its absence
	// reads as "no identity recorded", never as corruption.
	var currentTables int
	if err := e.queryDB().QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type='table' AND name='todo_dispatch_current'`).Scan(&currentTables); err != nil {
		return err
	}
	if currentTables == 0 {
		return nil
	}
	curRows, err := e.queryDB().QueryContext(ctx, `SELECT card_id,run_id,owner_label FROM todo_dispatch_current ORDER BY card_id`)
	if err != nil {
		return err
	}
	defer func() { _ = curRows.Close() }()
	for curRows.Next() {
		var c TodoDispatchCurrent
		if err := curRows.Scan(&c.CardID, &c.RunID, &c.OwnerLabel); err != nil {
			return err
		}
		result.DispatchCurrent = append(result.DispatchCurrent, c)
	}
	return curRows.Err()
}

// copyRuntime is only for an unpublished migration database, never a card edit.
// @MX:NOTE: [AUTO] Publication verifies runtime parity before the staging DB becomes visible.
func (e *backlogEngine) copyRuntime(ctx context.Context, runtime TodoRuntime) error {
	if len(runtime.Runs) == 0 && len(runtime.Assignments) == 0 && len(runtime.DispatchCurrent) == 0 {
		return nil
	}
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, todoRuntimeDDL); err != nil {
		return err
	}
	for _, r := range runtime.Runs {
		if _, err := tx.ExecContext(ctx, `INSERT INTO todo_runtime_runs(run_id,backend,manifest_json) VALUES(?,?,?)`, r.RunID, r.Backend, r.ManifestJSON); err != nil {
			return err
		}
	}
	for _, a := range runtime.Assignments {
		if _, err := tx.ExecContext(ctx, `INSERT INTO todo_runtime_assignments(run_id,card_id,owner_label,reported_state,event_kind,provenance_json) VALUES(?,?,?,?,?,?)`, a.RunID, a.CardID, a.OwnerLabel, a.ReportedState, a.EventKind, a.ProvenanceJSON); err != nil {
			return err
		}
	}
	// The dispatch-identity record is engagement state, not a cache: a
	// migration that dropped it would leave every later reconciliation on
	// this queue unadjudicable until each card re-dispatches (review
	// round-23).
	for _, c := range runtime.DispatchCurrent {
		if _, err := tx.ExecContext(ctx, `INSERT INTO todo_dispatch_current(card_id,run_id,owner_label) VALUES(?,?,?)`, c.CardID, c.RunID, c.OwnerLabel); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ArchiveOnRuntimeCompletion is the runtime completion's archive authority
// (card t1542, the audit P1 repair): a card whose completion a mechanical
// edge just established — the merged pull request of the github-flow
// delivery, for one — is closed by the session that completed it, in one
// locked write that archives the live card and rides the landing verdict the
// edge answered. Before this method the report row landed and the card sat
// live forever, so a leader ran `todo done` for every lane completion.
//
// Idempotent by reconciliation: a card already in the archive — with any
// verdict — reads as closed and returns nil, so the delivery edge can run it
// on every observation. A card in neither the queue nor the archive refuses,
// and a refusal writes nothing (Mutate's byte-identity contract).
func (s *BacklogStore) ArchiveOnRuntimeCompletion(cardID string, verdict LandingVerdict) error {
	err := s.Mutate(func(rec *BacklogRecord) error {
		for i := range rec.Items {
			if rec.Items[i].ID != cardID {
				continue
			}
			if err := rec.ArchiveCard(cardID); err != nil {
				return err
			}
			// REQ-TST-008 shape: the answering edge persists what its query
			// said — verdict, answering ref, verdict time — onto the entry
			// ArchiveCard just appended, alongside (never instead of) any
			// operator-recorded evidence the row already carried.
			rec.Archived[len(rec.Archived)-1].LandingVerdict = &verdict
			return nil
		}
		// Reconciliation: a card that is already archived reads as closed.
		for i := range rec.Archived {
			if rec.Archived[i].Item.ID == cardID {
				return nil
			}
		}
		return fmt.Errorf("no backlog item %s", cardID)
	})
	if err != nil {
		return fmt.Errorf("runtime completion archive: %w", err)
	}
	return nil
}

// recordRuntime shares the card mutation lock and commits seed/assignment together.
// @MX:NOTE: [AUTO] External provenance is captured before acquiring this lock; a reported completion never archives a card.
func (s *BacklogStore) recordRuntime(run TodoRuntimeRun, assignment *TodoRuntimeAssignment) (err error) {
	return s.recordRuntimeHook(run, assignment, nil)
}

// recordRuntimeHook is recordRuntime with a follow-up write that must land
// in the same critical section: hook runs after the assignment transaction
// commits while the queue lock is still held, so a completion path cannot
// interleave between the assignment save and the hook's write (review
// round-20 P1, SPEC-FACTORY-COMPLETION-RECOVERY-001). A hook error
// propagates after the lock releases.
func (s *BacklogStore) recordRuntimeHook(run TodoRuntimeRun, assignment *TodoRuntimeAssignment, hook func() error) (err error) {
	// @MX:NOTE: [TID:LINK] Runtime UUIDs are projected from the one live/archive card identity snapshot; owner and provenance remain report data.
	if strings.TrimSpace(run.RunID) == "" {
		return errors.New("runtime run_id is required")
	}
	lock, err := s.acquireLock()
	if err != nil {
		return err
	}
	defer func() { err = joinBacklogReleaseErr(err, lock.Release(), s.path) }()
	if target, readErr := os.ReadFile(filepath.Join(filepath.Dir(s.path), backlogRetiredFileName)); readErr == nil {
		return fmt.Errorf("%w: %s", ErrBacklogRelocated, target)
	} else if !os.IsNotExist(readErr) {
		return readErr
	}
	eng, err := s.openEngine(true)
	if err != nil {
		return err
	}
	defer func() { _ = eng.close() }()
	ctx, cancel := context.WithTimeout(context.Background(), backlogOpTimeout)
	defer cancel()
	tx, err := eng.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	// Recheck on the write snapshot, as another opener may have updated metadata.
	snapshot := &backlogEngine{reader: tx}
	if _, err := snapshot.runtimeVersion(ctx); err != nil {
		return err
	}
	if assignment != nil {
		var candidates int
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM (SELECT id FROM items WHERE id=? UNION ALL SELECT id FROM archived_items WHERE id=?)`, assignment.CardID, assignment.CardID).Scan(&candidates); err != nil {
			return err
		}
		if candidates == 0 {
			return fmt.Errorf("runtime card %q does not exist", assignment.CardID)
		}
		if candidates != 1 {
			return fmt.Errorf("runtime card %q is ambiguous (%d candidates)", assignment.CardID, candidates)
		}
	}
	if err := ensureStoredIdentities(ctx, tx, eng.dbPath); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, todoRuntimeDDL); err != nil {
		return err
	}
	conflict := ` ON CONFLICT(run_id) DO UPDATE SET backend=excluded.backend,manifest_json=excluded.manifest_json`
	if assignment != nil {
		conflict = ` ON CONFLICT(run_id) DO NOTHING`
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO todo_runtime_runs(run_id,backend,manifest_json) VALUES(?,?,?)`+conflict, run.RunID, run.Backend, run.ManifestJSON); err != nil {
		return err
	}
	if assignment != nil {
		// REQ-TSP-052: the write path records the canonical vocabulary
		// only — a legacy spelling carried in (an old launcher's env, a
		// stale dispatch surface) is relabeled at the choke point, so no
		// new legacy row is ever written.
		a := *assignment
		a.OwnerLabel = canonicalOwnerLabel(a.OwnerLabel)
		if _, err := tx.ExecContext(ctx, `INSERT INTO todo_runtime_assignments(run_id,card_id,owner_label,reported_state,event_kind,provenance_json) VALUES(?,?,?,?,?,?) ON CONFLICT(run_id,card_id) DO UPDATE SET owner_label=excluded.owner_label,reported_state=excluded.reported_state,event_kind=excluded.event_kind,provenance_json=excluded.provenance_json`, a.RunID, a.CardID, a.OwnerLabel, a.ReportedState, a.EventKind, a.ProvenanceJSON); err != nil {
			return err
		}
		if hook != nil {
			// The hook-bearing call is the dispatch assignment path
			// (RecordFactoryCardAssignment is its only caller); a state
			// report carries no hook. A dispatch records the queue's own
			// current-engagement identity in the SAME transaction as the
			// assignment and the binding hook (review round-23, card
			// t1538): the reconciliation then adjudicates supersession by
			// what the queue NAMES — an identity — instead of deriving
			// engagement order from row history, which three review rounds
			// measured unable to tell a mid-flight dispatch from
			// legitimate after-history. Last dispatch wins under the same
			// lock every binding writer holds; a hook failure rolls this
			// back with the assignment.
			if _, err := tx.ExecContext(ctx, `INSERT INTO todo_dispatch_current(card_id,run_id,owner_label) VALUES(?,?,?) ON CONFLICT(card_id) DO UPDATE SET run_id=excluded.run_id,owner_label=excluded.owner_label`, a.CardID, a.RunID, a.OwnerLabel); err != nil {
				return err
			}
		}
	}
	// REQ-TSP-050: the one-time relabel rides the same locked transaction —
	// idempotent (canonical rows pass through unchanged), so it costs one
	// DISTINCT scan on a migrated store.
	if _, err := migrateOwnerLabelVocabularyTx(ctx, tx); err != nil {
		return err
	}
	if hook != nil {
		// The hook runs INSIDE the assignment transaction (review round-22
		// P1, SPEC-FACTORY-COMPLETION-RECOVERY-001): a hook failure rolls
		// the assignment back with it — the dispatch fails cleanly instead
		// of leaving the record pointing at a run whose binding never
		// landed, which would let the completion gate close on the
		// previous run's approval.
		if err := hook(); err != nil {
			return err
		}
	}
	return tx.Commit()
}
