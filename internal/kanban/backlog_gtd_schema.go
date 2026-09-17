package kanban

import (
	"context"
	"database/sql"
	"fmt"
)

type gtdExecer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func bumpGTDRevision(ctx context.Context, execer gtdExecer) error {
	_, err := execer.ExecContext(ctx, `UPDATE gtd_meta SET value=CAST(CAST(value AS INTEGER)+1 AS TEXT) WHERE key='logical_revision'`)
	return err
}

const gtdSchemaVersion = "1"

// gtdDDL is additive to backlogDDL. It intentionally owns a separate
// gtd_meta version marker: the frozen queue schema remains version 1 and its
// five original tables, identifiers, and state checks are not rebuilt.
const gtdDDL = `
CREATE TABLE IF NOT EXISTS gtd_meta (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS gtd_items (
  item_id          TEXT PRIMARY KEY,
  content          TEXT NOT NULL,
  source           TEXT NOT NULL,
  sensitivity      TEXT NOT NULL,
  event_id         TEXT NOT NULL UNIQUE,
  outcome          TEXT NOT NULL DEFAULT '',
  completion_evidence TEXT NOT NULL DEFAULT '',
  disposition      TEXT NOT NULL DEFAULT '',
  class            TEXT NOT NULL DEFAULT '',
  action_context   TEXT NOT NULL DEFAULT '',
  review_at        TEXT NOT NULL DEFAULT '',
  authority        TEXT NOT NULL DEFAULT '',
  source_trust     TEXT NOT NULL DEFAULT '',
  status           TEXT NOT NULL,
  source_revision  INTEGER NOT NULL,
  card_id          TEXT,
  cancelled        INTEGER NOT NULL DEFAULT 0 CHECK (cancelled IN (0,1))
);
CREATE TABLE IF NOT EXISTS gtd_relations (
  subject_id       TEXT NOT NULL,
  object_id        TEXT NOT NULL,
  kind             TEXT NOT NULL,
  note             BLOB NOT NULL DEFAULT X'',
  source           TEXT NOT NULL,
  assertion_status TEXT NOT NULL,
  source_revision  INTEGER NOT NULL,
  policy_version   TEXT NOT NULL,
  PRIMARY KEY(subject_id, object_id, kind)
);
CREATE TABLE IF NOT EXISTS gtd_contracts (
  mission_id       TEXT PRIMARY KEY,
  policy_version   TEXT NOT NULL,
  contract_hash    TEXT NOT NULL,
  contract_json    BLOB NOT NULL,
  approved         INTEGER NOT NULL DEFAULT 0 CHECK (approved IN (0,1))
);
CREATE TABLE IF NOT EXISTS gtd_missions (
  mission_id       TEXT PRIMARY KEY,
  mission_mode     TEXT NOT NULL,
  state            TEXT NOT NULL,
  contract_hash    TEXT NOT NULL,
  snapshot_hash    TEXT NOT NULL,
  owner_id         TEXT NOT NULL DEFAULT '',
  lease_version    INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS gtd_events (
  event_id         TEXT PRIMARY KEY,
  mission_id       TEXT NOT NULL DEFAULT '',
  kind             TEXT NOT NULL,
  payload_hash     TEXT NOT NULL,
  created_at       TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS gtd_operations (
  operation_id     TEXT PRIMARY KEY,
  mission_id       TEXT NOT NULL,
  action           TEXT NOT NULL,
  target           TEXT NOT NULL,
  state            TEXT NOT NULL,
  receipt_json     BLOB NOT NULL,
  snapshot_hash    TEXT NOT NULL,
  updated_at       TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_gtd_items_card ON gtd_items(card_id);
CREATE INDEX IF NOT EXISTS idx_gtd_relations_object ON gtd_relations(object_id);
CREATE INDEX IF NOT EXISTS idx_gtd_operations_mission ON gtd_operations(mission_id);
`

// MigrateGTDSchema adds the GTD management tables to the same SQLite file as
// the existing queue. The transaction and IF NOT EXISTS statements make the
// migration repeatable without changing logical revision on a second run.
func MigrateGTDSchema(store *BacklogStore) error {
	if store == nil {
		return fmt.Errorf("migrate gtd schema: nil backlog store")
	}
	engine, err := store.openEngine(false)
	if err != nil {
		return fmt.Errorf("migrate gtd schema: %w", err)
	}
	defer func() { _ = engine.close() }()

	ctx, cancel := context.WithTimeout(context.Background(), backlogOpenTimeout)
	defer cancel()
	tx, err := engine.db.BeginTx(ctx, nil)
	if err != nil {
		return mapBacklogEngineError("begin gtd migration", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, gtdDDL); err != nil {
		return mapBacklogEngineError("apply gtd migration", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO gtd_meta(key,value) VALUES('schema_version',?),('logical_revision','0')`, gtdSchemaVersion); err != nil {
		return mapBacklogEngineError("stamp gtd schema", err)
	}
	if err := tx.Commit(); err != nil {
		return mapBacklogEngineError("commit gtd migration", err)
	}
	return nil
}
