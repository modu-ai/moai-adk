package homestate

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

const factorySchemaVersion = 5

const factoryDDL = `
CREATE TABLE IF NOT EXISTS meta (key TEXT PRIMARY KEY, value TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS workers (
  label TEXT PRIMARY KEY,
  pid INTEGER NOT NULL,
  backend TEXT NOT NULL DEFAULT '',
  session_id TEXT NOT NULL DEFAULT '',
  run_id TEXT NOT NULL DEFAULT '',
  registered_at TEXT NOT NULL,
  heartbeat_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS runs (
  run_id TEXT PRIMARY KEY,
  lead_session_id TEXT NOT NULL DEFAULT '',
  lead_backend TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL,
  manifest_json TEXT NOT NULL DEFAULT '{}',
  lead_pid INTEGER NOT NULL DEFAULT 0,
  lead_process_start TEXT NOT NULL DEFAULT '',
  lane_capacity INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS cards (
  run_id TEXT NOT NULL,
  card_id TEXT NOT NULL,
  owner_label TEXT NOT NULL DEFAULT '',
  state TEXT NOT NULL,
  version INTEGER NOT NULL DEFAULT 1,
  evidence_path TEXT NOT NULL DEFAULT '',
  updated_at TEXT NOT NULL,
  stage TEXT NOT NULL DEFAULT '',
  lease_holder TEXT NOT NULL DEFAULT '',
  lease_expires_at TEXT NOT NULL DEFAULT '',
  heartbeat_at TEXT NOT NULL DEFAULT '',
  decision_gate TEXT NOT NULL DEFAULT '',
  decision_question TEXT NOT NULL DEFAULT '',
  decision_resume TEXT NOT NULL DEFAULT '',
  decider TEXT NOT NULL DEFAULT '',
  decided_at TEXT NOT NULL DEFAULT '',
  failure_reason TEXT NOT NULL DEFAULT '',
  hint_prefer TEXT NOT NULL DEFAULT '',
  hint_after TEXT NOT NULL DEFAULT '',
  spec_id TEXT NOT NULL DEFAULT '',
  worktree_path TEXT NOT NULL DEFAULT '',
  evidence_sha TEXT NOT NULL DEFAULT '',
  merge_sha TEXT NOT NULL DEFAULT '',
  merge_tree TEXT NOT NULL DEFAULT '',
  remeasure_path TEXT NOT NULL DEFAULT '',
  contract_spec_id TEXT NOT NULL DEFAULT '',
  contract_sha256 TEXT NOT NULL DEFAULT '',
  contract_signed_at TEXT NOT NULL DEFAULT '',
  contract_event TEXT NOT NULL DEFAULT '',
  PRIMARY KEY(run_id, card_id)
);
CREATE TABLE IF NOT EXISTS events (
  seq INTEGER PRIMARY KEY AUTOINCREMENT,
  run_id TEXT NOT NULL DEFAULT '',
  kind TEXT NOT NULL,
  payload_json TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS dead_letters (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  run_id TEXT NOT NULL DEFAULT '',
  envelope_json TEXT NOT NULL,
  error TEXT NOT NULL,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS resume_handoffs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  status TEXT NOT NULL CHECK(status IN ('pending','claimed','consumed','failed','expired','cleared')),
  schema_version INTEGER NOT NULL,
  spec_id TEXT NOT NULL DEFAULT '',
  phase TEXT NOT NULL DEFAULT '',
  saved_at TEXT NOT NULL,
  saved_by_session TEXT NOT NULL DEFAULT '',
  conversation_language TEXT NOT NULL DEFAULT '',
  directives_json TEXT NOT NULL DEFAULT '{}',
  embedded_goal_json TEXT,
  body TEXT NOT NULL,
  body_sha256 TEXT NOT NULL,
  claim_token TEXT NOT NULL DEFAULT '',
  claimed_at TEXT,
  claim_expires_at TEXT,
  claim_owner_pid INTEGER,
  claim_owner_session TEXT NOT NULL DEFAULT '',
  claim_owner_fingerprint TEXT NOT NULL DEFAULT '',
  legacy_recovery INTEGER NOT NULL DEFAULT 0,
  legacy_recovery_reason TEXT NOT NULL DEFAULT '',
  consumed_at TEXT,
  error TEXT NOT NULL DEFAULT ''
);
CREATE UNIQUE INDEX IF NOT EXISTS one_pending_resume_handoff
ON resume_handoffs(status) WHERE status='pending';
CREATE TABLE IF NOT EXISTS memory_handoffs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  status TEXT NOT NULL CHECK(status IN ('pending','persisted','failed','expired')),
  sprint TEXT NOT NULL,
  spec TEXT NOT NULL,
  result_status TEXT NOT NULL,
  body TEXT NOT NULL,
  index_line TEXT NOT NULL,
  supersedes TEXT NOT NULL DEFAULT '',
  body_sha256 TEXT NOT NULL,
  created_at TEXT NOT NULL,
  persisted_at TEXT,
  error TEXT NOT NULL DEFAULT ''
);
CREATE UNIQUE INDEX IF NOT EXISTS one_pending_memory_handoff
ON memory_handoffs(status) WHERE status='pending';
CREATE TABLE IF NOT EXISTS handoff_events (
  seq INTEGER PRIMARY KEY AUTOINCREMENT,
  flow TEXT NOT NULL,
  handoff_id INTEGER NOT NULL,
  from_status TEXT NOT NULL DEFAULT '',
  to_status TEXT NOT NULL,
  detail TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS leader_approvals (
  card_uuid TEXT NOT NULL,
  run_id TEXT NOT NULL,
  card_id TEXT NOT NULL,
  factory_version INTEGER NOT NULL,
  evidence_hash TEXT NOT NULL,
  issuer TEXT NOT NULL,
  issuer_role TEXT NOT NULL,
  issued_at TEXT NOT NULL,
  PRIMARY KEY(card_uuid, run_id)
);
`

type FactoryDB struct {
	DB   *sql.DB
	Path string
}

func OpenFactory(projectRoot string) (*FactoryDB, error) {
	if err := EnsureProjectLayout(projectRoot); err != nil {
		return nil, err
	}
	path, err := FactoryDBPath(projectRoot)
	if err != nil {
		return nil, err
	}
	return OpenFactoryPath(path)
}

// factoryBusyTimeoutDefault is the busy timeout every connection opened
// through OpenFactory and OpenFactoryPath carries in its DSN.
const factoryBusyTimeoutDefault = 5 * time.Second

// OpenFactoryBounded is OpenFactory with a busy timeout of busy carried in the
// connection's DSN instead of the default 5 s (SPEC-FACTORY-ATOMIC-LEASE-001
// plan D2). A runtime PRAGMA does not substitute: it did not survive a
// context-cancelled call, so the value rides the DSN. The lease path opens its
// record connection through it so that a claim stalled behind another writer
// overshoots its deadline by at most the busy timeout.
func OpenFactoryBounded(projectRoot string, busy time.Duration) (*FactoryDB, error) {
	if err := EnsureProjectLayout(projectRoot); err != nil {
		return nil, err
	}
	path, err := FactoryDBPath(projectRoot)
	if err != nil {
		return nil, err
	}
	return openFactoryPathBusy(path, busy)
}

// OpenFactoryPath opens a factory database at an already-resolved path. It is
// used by compatibility adapters whose public API historically accepted a
// registry path rather than a project root.
func OpenFactoryPath(path string) (*FactoryDB, error) {
	return openFactoryPathBusy(path, factoryBusyTimeoutDefault)
}

// OpenFactoryReadonly opens a factory database strictly for reading: no DDL
// runs, no migration fires, nothing is created. A reader that cannot write
// must never migrate a store as a side effect (SPEC-FACTORY-COMPLETION-RECOVERY-001
// review P2-2): a database on an older schema is read as it stands, and
// tables a later schema added are simply absent — callers treat a missing
// table as empty data, never as a reason to migrate. The path must already
// exist; a missing file is an error here, so callers stat before they call.
func OpenFactoryReadonly(path string) (*FactoryDB, error) {
	// A missing file is refused here rather than created: the read-only
	// surface observes a store that exists, it never bootstraps one.
	if _, statErr := os.Stat(path); statErr != nil {
		return nil, fmt.Errorf("open factory database read-only: %w", statErr)
	}
	values := url.Values{}
	// query_only refuses every SQL write — DDL included — at the pragma
	// layer, so no schema statement can ever run through this handle.
	// mode=ro is deliberately NOT set: a WAL database cannot be opened
	// through a plain read-only connection when its -shm is absent, and the
	// scan must read a live store. The no-migration guarantee comes from
	// never executing the schema DDL on this handle at all.
	values.Add("_pragma", "query_only(ON)")
	values.Add("_pragma", fmt.Sprintf("busy_timeout(%d)", factoryBusyTimeoutDefault.Milliseconds()))
	p := filepath.ToSlash(path)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	dsn := (&url.URL{Scheme: "file", Path: p, RawQuery: values.Encode()}).String()
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("open factory database read-only: %w", err)
	}
	return &FactoryDB{DB: db, Path: path}, nil
}

// FactoryTablePresent reports whether the named table exists in the factory
// database — the missing-table check read-only consumers run instead of
// migrating. Read-only safe.
func (f *FactoryDB) FactoryTablePresent(ctx context.Context, table string) (bool, error) {
	switch table {
	case "meta", "workers", "runs", "cards", "events", "dead_letters",
		"resume_handoffs", "memory_handoffs", "handoff_events", "leader_approvals":
		// SQL: the allowlist pins table to a known identifier; the value never
		// reaches the query from input.
	default:
		return false, fmt.Errorf("%w: unknown factory table %q", ErrInvalidCardInput, table)
	}
	var present int
	err := f.DB.QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&present)
	return present > 0, err
}

func openFactoryPathBusy(path string, busy time.Duration) (*FactoryDB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	if err := os.Chmod(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	values := url.Values{}
	values.Add("_pragma", fmt.Sprintf("busy_timeout(%d)", busy.Milliseconds()))
	values.Add("_pragma", "journal_mode(WAL)")
	values.Add("_pragma", "foreign_keys(ON)")
	values.Add("_txlock", "immediate")
	p := filepath.ToSlash(path)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	dsn := (&url.URL{Scheme: "file", Path: p, RawQuery: values.Encode()}).String()
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := retryFactoryBusy(ctx, func() error {
		_, err := db.ExecContext(ctx, factoryDDL)
		return err
	}); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("initialize factory database: %w", err)
	}
	var version string
	err = retryFactoryBusy(ctx, func() error {
		return db.QueryRowContext(ctx, `SELECT value FROM meta WHERE key='schema_version'`).Scan(&version)
	})
	if errors.Is(err, sql.ErrNoRows) {
		err = retryFactoryBusy(ctx, func() error {
			_, execErr := db.ExecContext(ctx, `INSERT OR IGNORE INTO meta(key,value) VALUES('schema_version',?)`, strconv.Itoa(factorySchemaVersion))
			return execErr
		})
		if err == nil {
			err = retryFactoryBusy(ctx, func() error {
				return db.QueryRowContext(ctx, `SELECT value FROM meta WHERE key='schema_version'`).Scan(&version)
			})
		}
	}
	if err == nil && version == "1" {
		err = migrateFactoryV1ToV2(ctx, db)
		if err == nil {
			version = "2"
		}
	}
	if err == nil && version == "2" {
		err = migrateFactoryV2ToV3(ctx, db)
		if err == nil {
			version = "3"
		}
	}
	if err == nil && version == "3" {
		err = migrateFactoryV3ToV4(ctx, db)
		if err == nil {
			version = "4"
		}
	}
	if err == nil && version == "4" {
		err = migrateFactoryV4ToV5(ctx, db)
		if err == nil {
			version = "5"
		}
	}
	if err == nil && version != strconv.Itoa(factorySchemaVersion) {
		err = fmt.Errorf("unsupported factory schema version %q", version)
	}
	if err == nil {
		_, err = db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS resume_handoff_claim_expiry ON resume_handoffs(status,claim_expires_at,id)`)
	}
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := secureFactoryArtifacts(path); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &FactoryDB{DB: db, Path: path}, nil
}

func migrateFactoryV1ToV2(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	columns := []string{
		`ALTER TABLE resume_handoffs ADD COLUMN claim_expires_at TEXT`,
		`ALTER TABLE resume_handoffs ADD COLUMN claim_owner_pid INTEGER`,
		`ALTER TABLE resume_handoffs ADD COLUMN claim_owner_session TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE resume_handoffs ADD COLUMN claim_owner_fingerprint TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE resume_handoffs ADD COLUMN legacy_recovery INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE resume_handoffs ADD COLUMN legacy_recovery_reason TEXT NOT NULL DEFAULT ''`,
	}
	for _, stmt := range columns {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	rows, err := tx.QueryContext(ctx, `SELECT id,claimed_at FROM resume_handoffs WHERE status='claimed'`)
	if err != nil {
		return err
	}
	type legacy struct {
		id      int64
		claimed sql.NullString
	}
	var claimed []legacy
	for rows.Next() {
		var row legacy
		if err := rows.Scan(&row.id, &row.claimed); err != nil {
			_ = rows.Close()
			return err
		}
		claimed = append(claimed, row)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, row := range claimed {
		if row.claimed.Valid {
			if at, parseErr := time.Parse(time.RFC3339Nano, row.claimed.String); parseErr == nil {
				_, err = tx.ExecContext(ctx, `UPDATE resume_handoffs SET claim_expires_at=? WHERE id=?`, at.Add(5*time.Minute).Format(time.RFC3339Nano), row.id)
				if err != nil {
					return err
				}
				continue
			}
		}
		if _, err = tx.ExecContext(ctx, `UPDATE resume_handoffs SET legacy_recovery=1,legacy_recovery_reason='missing or invalid claimed_at' WHERE id=?`, row.id); err != nil {
			return err
		}
	}
	if _, err = tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS resume_handoff_claim_expiry ON resume_handoffs(status,claim_expires_at,id)`); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE meta SET value='2' WHERE key='schema_version' AND value='1'`); err != nil {
		return err
	}
	return tx.Commit()
}

// migrateFactoryV2ToV3 adds the run-owner identity columns, following the
// shape migrateFactoryV1ToV2 already establishes: the ALTER TABLEs inside one
// transaction, then the meta.schema_version update.
//
// The defaults are what keep every pre-existing row legible rather than
// broken: lead_pid = 0 is the sentinel that routes a legacy row to the
// role='lead' peer fallback, so migration and prevention are one mechanism.
func migrateFactoryV2ToV3(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	// factoryDDL runs before the version check, so on a database whose `runs`
	// table did not exist the DDL has already created it in the v3 shape. Add
	// only the columns that are actually missing.
	existing, err := factoryRunColumns(ctx, tx)
	if err != nil {
		return err
	}
	for column, stmt := range map[string]string{
		"lead_pid":           `ALTER TABLE runs ADD COLUMN lead_pid INTEGER NOT NULL DEFAULT 0`,
		"lead_process_start": `ALTER TABLE runs ADD COLUMN lead_process_start TEXT NOT NULL DEFAULT ''`,
	} {
		if existing[column] {
			continue
		}
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE meta SET value='3' WHERE key='schema_version' AND value='2'`); err != nil {
		return err
	}
	return tx.Commit()
}

// migrateFactoryV3ToV4 adds the F1 card-record columns (card state machine,
// lease, decision, hints, evidence, contract pointer). Every column is TEXT
// NOT NULL with an empty-string default, so no backfill is needed and every v3 row stays
// valid; a row whose state lies outside the F1 state set is classified as
// legacy by the transition API rather than rewritten here.
func migrateFactoryV3ToV4(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	existing, err := factoryTableColumns(ctx, tx, "cards")
	if err != nil {
		return err
	}
	for _, column := range cardF1Columns {
		if existing[column] {
			continue
		}
		// SQL: column comes from the constant cardF1Columns list, never from input.
		if _, err := tx.ExecContext(ctx, `ALTER TABLE cards ADD COLUMN `+column+` TEXT NOT NULL DEFAULT ''`); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE meta SET value='4' WHERE key='schema_version' AND value='3'`); err != nil {
		return err
	}
	return tx.Commit()
}

// cardF1Columns lists, in declaration order, the text columns schema version 4
// adds to `cards`. factoryDDL declares the same set for a fresh database.
var cardF1Columns = []string{
	"stage", "lease_holder", "lease_expires_at", "heartbeat_at",
	"decision_gate", "decision_question", "decision_resume", "decider", "decided_at",
	"failure_reason", "hint_prefer", "hint_after", "spec_id", "worktree_path",
	"evidence_sha", "merge_sha", "merge_tree", "remeasure_path",
	"contract_spec_id", "contract_sha256", "contract_signed_at", "contract_event",
}

// migrateFactoryV4ToV5 adds the runs.lane_capacity column — the run's
// declared lane capacity recorded at leader start (SPEC-CODEX-LANE-SLOTS-001
// REQ-004): the operator-supplied count, or LaneCapacityDerived for a run
// started without one.
//
// The migration default is 1, not the derived marker, on purpose: a row that
// predates the datum has unknown provenance, and the bound the previous
// binary actually applied to its joins was the launcher constant (1) — the
// hard refusal of REQ-006. Defaulting unknown rows to capacity-open would
// let lanes grow into runs whose operator may have declared a bound, so the
// conservative reading keeps every pre-existing row on the behavior it
// already had.
func migrateFactoryV4ToV5(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	// factoryDDL runs before the version check, so on a database whose `runs`
	// table did not exist the DDL has already created it in the v5 shape. Add
	// only the column that is actually missing.
	existing, err := factoryRunColumns(ctx, tx)
	if err != nil {
		return err
	}
	if !existing["lane_capacity"] {
		if _, err := tx.ExecContext(ctx, `ALTER TABLE runs ADD COLUMN lane_capacity INTEGER NOT NULL DEFAULT 1`); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE meta SET value='5' WHERE key='schema_version' AND value='4'`); err != nil {
		return err
	}
	return tx.Commit()
}

func factoryRunColumns(ctx context.Context, tx *sql.Tx) (map[string]bool, error) {
	return factoryTableColumns(ctx, tx, "runs")
}

func factoryTableColumns(ctx context.Context, tx *sql.Tx, table string) (_ map[string]bool, err error) {
	// SQL: table is an internal constant ("runs", "cards"), never input.
	rows, err := tx.QueryContext(ctx, `PRAGMA table_info(`+table+`)`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	columns := map[string]bool{}
	for rows.Next() {
		var (
			cid        int
			name, typ  string
			notNull    int
			defaultVal sql.NullString
			pk         int
		)
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultVal, &pk); err != nil {
			return nil, err
		}
		columns[name] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return columns, nil
}

func secureFactoryArtifacts(path string) error {
	for _, artifact := range []string{path, path + "-wal", path + "-shm"} {
		if err := os.Chmod(artifact, 0o600); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("secure factory database artifact %s: %w", artifact, err)
		}
	}
	return nil
}

func retryFactoryBusy(ctx context.Context, fn func() error) error {
	var err error
	for attempt := 0; attempt < 100; attempt++ {
		err = fn()
		if err == nil || !isFactoryBusy(err) {
			return err
		}
		select {
		case <-ctx.Done():
			return errors.Join(ctx.Err(), err)
		case <-time.After(20 * time.Millisecond):
		}
	}
	return err
}

func isFactoryBusy(err error) bool {
	var sqliteErr interface{ Code() int }
	if !errors.As(err, &sqliteErr) {
		return false
	}
	switch sqliteErr.Code() & 0xff {
	case sqlite3.SQLITE_BUSY, sqlite3.SQLITE_LOCKED:
		return true
	default:
		return false
	}
}

func (f *FactoryDB) Close() error { return f.DB.Close() }

// ImportLegacyWorkers imports workers.json only when the SQLite roster is
// empty. The legacy file remains as read-only downgrade evidence.
func (f *FactoryDB) ImportLegacyWorkers(path string) error {
	tx, err := f.DB.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var marker string
	if err := tx.QueryRow(`SELECT value FROM meta WHERE key='legacy_workers_imported'`).Scan(&marker); err == nil {
		return nil
	} else if err != sql.ErrNoRows {
		return err
	}
	var count int
	if err := tx.QueryRow(`SELECT count(*) FROM workers`).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		raw, readErr := os.ReadFile(path)
		if readErr != nil && !os.IsNotExist(readErr) {
			return readErr
		}
		if readErr == nil {
			var rows map[string]struct {
				PID          int    `json:"pid"`
				RegisteredAt string `json:"registered_at"`
			}
			if err := json.Unmarshal(raw, &rows); err != nil {
				return err
			}
			for label, row := range rows {
				at := row.RegisteredAt
				if at == "" {
					at = time.Now().UTC().Format(time.RFC3339Nano)
				}
				if _, err := tx.Exec(`INSERT OR IGNORE INTO workers(label,pid,registered_at,heartbeat_at) VALUES(?,?,?,?)`, label, row.PID, at, at); err != nil {
					return err
				}
			}
		}
	}
	if _, err := tx.Exec(`INSERT INTO meta(key,value) VALUES('legacy_workers_imported','1')`); err != nil {
		return err
	}
	return tx.Commit()
}
