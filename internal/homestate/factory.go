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

const factorySchemaVersion = 1

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

// OpenFactoryPath opens a factory database at an already-resolved path. It is
// used by compatibility adapters whose public API historically accepted a
// registry path rather than a project root.
func OpenFactoryPath(path string) (*FactoryDB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	if err := os.Chmod(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	values := url.Values{}
	values.Add("_pragma", "busy_timeout(5000)")
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
	if err == nil && version != strconv.Itoa(factorySchemaVersion) {
		err = fmt.Errorf("unsupported factory schema version %q", version)
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
