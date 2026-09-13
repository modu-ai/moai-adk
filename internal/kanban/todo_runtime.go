package kanban

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
	return rows.Err()
}

// copyRuntime is only for an unpublished migration database, never a card edit.
// @MX:NOTE: [AUTO] Publication verifies runtime parity before the staging DB becomes visible.
func (e *backlogEngine) copyRuntime(ctx context.Context, runtime TodoRuntime) error {
	if len(runtime.Runs) == 0 && len(runtime.Assignments) == 0 {
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
	return tx.Commit()
}

// recordRuntime shares the card mutation lock and commits seed/assignment together.
// @MX:NOTE: [AUTO] External provenance is captured before acquiring this lock; a reported completion never archives a card.
func (s *BacklogStore) recordRuntime(run TodoRuntimeRun, assignment *TodoRuntimeAssignment) (err error) {
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
		a := assignment
		if _, err := tx.ExecContext(ctx, `INSERT INTO todo_runtime_assignments(run_id,card_id,owner_label,reported_state,event_kind,provenance_json) VALUES(?,?,?,?,?,?) ON CONFLICT(run_id,card_id) DO UPDATE SET owner_label=excluded.owner_label,reported_state=excluded.reported_state,event_kind=excluded.event_kind,provenance_json=excluded.provenance_json`, a.RunID, a.CardID, a.OwnerLabel, a.ReportedState, a.EventKind, a.ProvenanceJSON); err != nil {
			return err
		}
	}
	return tx.Commit()
}
