// todo_owner_label.go — SPEC-TODO-SURFACE-POLISH-001 (card t1349)
// REQ-TSP-050/052: the owner_label vocabulary migration and write
// normalization.
//
// The todo_runtime_assignments table accumulated the pre-rename label
// spellings (`lead`, `lead-<suffix>`, `worker-<n>`, `agent-<n>`) while the
// canonical vocabulary moved to `leader` / `leader-<suffix>` (role.go) and
// `lane-<n>` (bootstrap.go). REQ-RNC-009 keeps `lead` a detection-only
// legacy spelling for the ROLE surfaces — this file is the card-scoped
// exception the SPEC resolves (spec.md §B.6/§F.1): the single
// owner_label COLUMN is relabeled once, under the queue lock, to the
// canonical vocabulary, and the write path never writes a legacy spelling
// again.
//
// Scope is exactly one column of one table. No other label store
// (internal/homestate cards, role declarations, workers registries) is
// touched — the SPEC's Out of Scope names them.
package kanban

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// canonicalOwnerLabel maps one owner_label value onto the canonical
// vocabulary, reusing the same detectors the refusal and stale-record
// paths use (REQ-TSP-051): IsLegacyLeaderSpelling for the leader spellings,
// the factory-label discriminators for the lane spellings. A value that is
// already canonical returns unchanged, so the mapper is its own idempotence
// witness — mapping a mapped value is a no-op.
func canonicalOwnerLabel(label string) string {
	switch {
	case IsLegacyLeaderSpelling(label):
		// `lead` → `leader`; `lead-<suffix>` → `leader-<suffix>` — the
		// bumped and run-id forms the leader vocabulary composes.
		return RoleLeader + strings.TrimPrefix(label, legacyLeaderSpelling)
	case IsLegacyFactoryRoleValue(label):
		if n, ok := SplitFactoryLegacyLabel(label); ok {
			// `worker-<n>` / `agent-<n>` → `lane-<n>` — the canonical
			// factory lane label the current launcher issues.
			return FactoryLaneLabel(n)
		}
		// The bare role tokens (`worker`, `agent`) carry no lane number in
		// the legacy regime; the canonical bare role is the answer.
		return RoleLane
	default:
		return label
	}
}

// migrateOwnerLabelVocabularyTx relabels every legacy owner_label row
// inside the caller's transaction — the locked-write discipline: the
// migration commits atomically with whatever write it rides, and a failure
// rolls the whole transaction back, never half a relabel.
//
// It is idempotent by construction: canonical values return unchanged from
// canonicalOwnerLabel, so a second run's UPDATE list is empty.
func migrateOwnerLabelVocabularyTx(ctx context.Context, tx *sql.Tx) (int, error) {
	rows, err := tx.QueryContext(ctx, `SELECT DISTINCT owner_label FROM todo_runtime_assignments`)
	if err != nil {
		return 0, fmt.Errorf("owner label migration: distinct: %w", err)
	}
	var updates [][2]string
	for rows.Next() {
		var label string
		if err := rows.Scan(&label); err != nil {
			_ = rows.Close()
			return 0, fmt.Errorf("owner label migration: scan: %w", err)
		}
		if canonical := canonicalOwnerLabel(label); canonical != label {
			updates = append(updates, [2]string{canonical, label})
		}
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return 0, fmt.Errorf("owner label migration: rows: %w", err)
	}
	_ = rows.Close()

	migrated := 0
	for _, u := range updates {
		res, err := tx.ExecContext(ctx,
			`UPDATE todo_runtime_assignments SET owner_label = ? WHERE owner_label = ?`, u[0], u[1])
		if err != nil {
			return migrated, fmt.Errorf("owner label migration: relabel %q: %w", u[1], err)
		}
		n, err := res.RowsAffected()
		if err != nil {
			return migrated, fmt.Errorf("owner label migration: affected: %w", err)
		}
		migrated += int(n)
	}
	return migrated, nil
}

// ownerLabelHistogram counts owner_label values in the caller's
// transaction — the before/after record REQ-TSP-050 reports.
func ownerLabelHistogram(ctx context.Context, q interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}) (map[string]int, error) {
	rows, err := q.QueryContext(ctx, `SELECT owner_label, count(*) FROM todo_runtime_assignments GROUP BY owner_label`)
	if err != nil {
		return nil, fmt.Errorf("owner label histogram: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := map[string]int{}
	for rows.Next() {
		var label string
		var n int
		if err := rows.Scan(&label, &n); err != nil {
			return nil, fmt.Errorf("owner label histogram: scan: %w", err)
		}
		out[label] = n
	}
	return out, rows.Err()
}

// MigrateOwnerLabelVocabulary migrates the legacy owner_label spellings in
// this store's todo_runtime_assignments table to the canonical vocabulary
// (`lead` → `leader`, `worker-<n>` / `agent-<n>` → `lane-<n>`) in ONE
// locked transaction, and reports the label histogram before and after
// (REQ-TSP-050). Re-running a migrated store migrates zero rows — the
// canonical values pass through the mapper unchanged.
//
// Exactly one column of one table changes: the update list is computed from
// the legacy detectors, and no other table is written by the migration.
func (s *BacklogStore) MigrateOwnerLabelVocabulary() (migrated int, before, after map[string]int, err error) {
	lock, err := s.acquireLock()
	if err != nil {
		return 0, nil, nil, err
	}
	defer func() { err = joinBacklogReleaseErr(err, lock.Release(), s.path) }()
	eng, err := s.openEngine(true)
	if err != nil {
		return 0, nil, nil, err
	}
	defer func() { _ = eng.close() }()
	ctx, cancel := context.WithTimeout(context.Background(), backlogOpTimeout)
	defer cancel()
	tx, err := eng.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("owner label migration: begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// The table may not exist yet on a queue that has never recorded a
	// runtime event — the DDL ensure is the same one recordRuntime runs.
	if _, err := tx.ExecContext(ctx, todoRuntimeDDL); err != nil {
		return 0, nil, nil, fmt.Errorf("owner label migration: ddl: %w", err)
	}
	if before, err = ownerLabelHistogram(ctx, tx); err != nil {
		return 0, nil, nil, err
	}
	if migrated, err = migrateOwnerLabelVocabularyTx(ctx, tx); err != nil {
		return 0, nil, nil, err
	}
	if after, err = ownerLabelHistogram(ctx, tx); err != nil {
		return 0, nil, nil, err
	}
	if err := tx.Commit(); err != nil {
		return 0, nil, nil, fmt.Errorf("owner label migration: commit: %w", err)
	}
	return migrated, before, after, nil
}
