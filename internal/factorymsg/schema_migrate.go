package factorymsg

import (
	"context"
	"database/sql"
	"fmt"
)

// queryer is the read surface shared by *sql.DB and *sql.Tx, so checks that
// run inside a transaction never reach for a second connection (the store
// holds exactly one).
type queryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// messagesLaneScopeDDL is the messages table at schema version 2: idempotency
// is scoped to (project_key, run_id, sender_slot, idem_key), independent of
// the sender's session UUID and generation. It must stay identical to the
// messages statement in the store schema constant.
const messagesLaneScopeDDL = `CREATE TABLE messages_v2(id TEXT PRIMARY KEY, schema_version INTEGER NOT NULL, project_key TEXT NOT NULL, run_id TEXT NOT NULL, sender_session TEXT NOT NULL, sender_generation INTEGER NOT NULL, recipient_session TEXT NOT NULL, recipient_generation INTEGER NOT NULL, kind TEXT NOT NULL, idem_key TEXT NOT NULL, task_ref TEXT NOT NULL, correlation_id TEXT NOT NULL, created_at TEXT NOT NULL, expires_at TEXT NOT NULL, payload BLOB NOT NULL, state TEXT NOT NULL, claim_token TEXT NOT NULL DEFAULT '', claim_expires_at TEXT, disposition TEXT NOT NULL DEFAULT '', acknowledged_at TEXT, sender_slot TEXT NOT NULL DEFAULT '', UNIQUE(project_key,run_id,sender_slot,idem_key))`

const messagesV1Columns = `id,schema_version,project_key,run_id,sender_session,sender_generation,recipient_session,recipient_generation,kind,idem_key,task_ref,correlation_id,created_at,expires_at,payload,state,claim_token,claim_expires_at,disposition,acknowledged_at`

// ensureSchema brings an existing broker database up to the current schema.
// It is read-only when nothing is missing, so the hook hot path pays one
// catalogue query rather than a write lock.
func ensureSchema(ctx context.Context, db *sql.DB) error {
	var dispatches, laneScoped int
	if err := db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM sqlite_master WHERE type='table' AND name='dispatches'), (SELECT count(*) FROM pragma_table_info('messages') WHERE name='sender_slot')`).Scan(&dispatches, &laneScoped); err != nil {
		return err
	}
	if dispatches == 0 {
		if _, err := db.ExecContext(ctx, dispatchSchema); err != nil {
			return err
		}
	}
	if laneScoped == 0 {
		return migrateMessagesLaneScope(ctx, db)
	}
	return nil
}

// migrateMessagesLaneScope rebuilds a schema-version-1 messages table
// (UNIQUE(sender_session, idem_key)) under the lane-slot scope, keeping every
// row, its ID, and every column value. A row's sender slot is the peers slot
// its sender session currently holds; a session no longer in peers maps to
// "legacy:<session>", so it can never collide with a live lane's scope.
// Rows mapping to one real slot all share that slot's single current session,
// so the old per-session uniqueness already guarantees no collision.
//
// @MX:WARN: [AUTO] rewrites the messages table in place (create, copy, drop, rename)
// @MX:REASON: runs inside one immediate transaction and re-checks the column there, so a concurrent opener either migrates first or sees the migrated table; the row-count check aborts before the drop on any loss
func migrateMessagesLaneScope(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var laneScoped int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM pragma_table_info('messages') WHERE name='sender_slot'`).Scan(&laneScoped); err != nil {
		return err
	}
	if laneScoped != 0 {
		return tx.Commit()
	}
	if _, err := tx.ExecContext(ctx, messagesLaneScopeDDL); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO messages_v2(`+messagesV1Columns+`,sender_slot) SELECT `+messagesV1Columns+`,COALESCE((SELECT p.slot FROM peers p WHERE p.session_uuid=messages.sender_session),'legacy:'||sender_session) FROM messages`); err != nil {
		return err
	}
	var before, after int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM messages`).Scan(&before); err != nil {
		return err
	}
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM messages_v2`).Scan(&after); err != nil {
		return err
	}
	if before != after {
		return fmt.Errorf("messages migration would lose rows: %d -> %d", before, after)
	}
	for _, stmt := range []string{
		`DROP TABLE messages`,
		`ALTER TABLE messages_v2 RENAME TO messages`,
		`CREATE INDEX IF NOT EXISTS messages_recipient_state ON messages(recipient_session,recipient_generation,state,created_at)`,
	} {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return tx.Commit()
}
