// backlog_sqlite.go — the SQLite storage engine beneath the backlog queue
// (SPEC-TODO-SQLITE-001 REQ-TOSQ-001..006, M1).
//
// The queue's persistence moves from one JSON document to one SQLite
// database file sitting BESIDE where the JSON used to be: the store still
// resolves through BacklogPathForRoot (the queue-file join, unchanged in
// shape), and the engine derives its own artifact as the sibling
// `<queue-dir>/backlog.db`. Keeping BacklogPathForRoot on the json name is
// deliberate: every caller contract around it is preserved verbatim, the
// downgrade story stays exactly "an old binary reads only backlog.json", and
// the characterization suites seed legacy JSON the lazy migration consumes.
//
// Engine posture (REQ-TOSQ-003): WAL journal mode so readers never block the
// writer, busy_timeout >= 5000ms so a contending writer waits instead of
// failing spuriously, and transactions opened IMMEDIATE so a write never
// upgrades mid-transaction. One pooled connection per store serializes
// same-process access deterministically; cross-process serialization remains
// the outer backlog.lock file lock (REQ-TOSQ-008).
//
// No SQL statement ever interpolates user text: every value travels through
// a parameter placeholder (acceptance.md D.6 Secured gate).
package kanban

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite" // registers the pure-Go "sqlite" database/sql driver (CGO_ENABLED=0 safe)
	sqlite3 "modernc.org/sqlite/lib"
)

const (
	// sqliteDriverName is the database/sql name modernc.org/sqlite registers.
	sqliteDriverName = "sqlite"

	// backlogDBFileNameSuffix turns a queue-file base name into the engine's
	// sibling artifact name (backlog.json -> backlog.db).
	backlogDBFileNameSuffix = ".db"

	// backlogSchemaVersion stamps the physical layout in meta. Bumped ONLY by
	// a future SPEC redefining the DDL; the runtime reads it and refuses an
	// unrecognized version rather than guessing.
	backlogSchemaVersion = "1"

	// backlogBusyTimeoutMS is the per-connection busy timeout (REQ-TOSQ-003
	// mandates >= 5000): how long a writer blocked on another process's lock
	// waits before surfacing ErrBacklogBusy.
	backlogBusyTimeoutMS = 5000

	// backlogOpenTimeout bounds the connection-establishment health ping, so
	// a wedged filesystem fails fast into the named taxonomy instead of
	// hanging a verb indefinitely.
	backlogOpenTimeout = 10 * time.Second
)

// meta-table keys. Values are TEXT; last_seq stores a decimal integer.
const (
	backlogMetaKeyLastSeq       = "last_seq"
	backlogMetaKeySchemaVersion = "schema_version"

	// backlogMetaKeyQuarantinePending records that a migration committed its
	// data but has not yet quarantined the legacy file. It is written inside
	// the migration transaction and cleared once the rename lands.
	//
	// It exists to answer a question the filesystem cannot: a database sitting
	// beside a backlog.json is ambiguous. That json may be pre-cutover legacy
	// left by a crash between the commit and the quarantine — which must be
	// quarantined — or it may be an export this binary just wrote for a
	// downgrade, which must be left exactly where the operator put it.
	// Quarantining an export would silently delete the only artifact a
	// downgraded release can read, and the operator would not find out until
	// the older binary came up to an empty queue.
	backlogMetaKeyQuarantinePending = "legacy_quarantine_pending"
)

// backlogDDL is the physical schema (design.md §2). Everything is IF NOT
// EXISTS so opening any existing database is idempotent. seq carries the
// insertion-order duty of the legacy array (ORDER BY seq reproduces it);
// findings keep NO uniqueness constraint because lossless doctrine forbids
// letting a storage constraint reject legacy duplicate tuples at migration —
// tuple-once behavior stays application-level (AppendFindingOnce).
//
// The two archived_* tables are the reversal storage (SPEC-TODO-DESTRUCTIVE-
// GUARD-001 REQ-TDG-003/004). They are ADDITIVE and cost nothing on an
// existing database: this whole DDL runs on every open and every statement is
// IF NOT EXISTS, so a queue created by an earlier binary gains the tables the
// first time a newer one opens it — which is precisely why the archive is a
// pair of tables rather than a fourth `state` value. SQLite cannot ALTER a
// CHECK constraint, so admitting a fourth state would need a table rebuild on
// every operator queue in the field.
//
// archived_items deliberately carries NO state CHECK: an archived row is
// history rather than a live lifecycle position, and leaving the constraint
// off keeps the live three-value enum the single constrained surface.
const backlogDDL = `
CREATE TABLE IF NOT EXISTS meta (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS items (
  seq      INTEGER PRIMARY KEY,
  id       TEXT    NOT NULL UNIQUE,
  text     TEXT    NOT NULL,
  added_at TEXT    NOT NULL,
  spec_id  TEXT,
  state    TEXT    NOT NULL CHECK (state IN ('queued','picked','dropped'))
);
CREATE TABLE IF NOT EXISTS findings (
  subject_id TEXT  NOT NULL,
  related_id TEXT  NOT NULL,
  relation   TEXT  NOT NULL,
  source     TEXT  NOT NULL,
  score      REAL  NOT NULL,
  note       TEXT  NOT NULL DEFAULT '',
  at         TEXT  NOT NULL
);
CREATE TABLE IF NOT EXISTS archived_items (
  seq      INTEGER PRIMARY KEY,
  id       TEXT    NOT NULL UNIQUE,
  text     TEXT    NOT NULL,
  added_at TEXT    NOT NULL,
  spec_id  TEXT,
  state    TEXT    NOT NULL,
  position INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS archived_findings (
  archive_seq INTEGER NOT NULL,
  position    INTEGER NOT NULL,
  subject_id  TEXT  NOT NULL,
  related_id  TEXT  NOT NULL,
  relation    TEXT  NOT NULL,
  source      TEXT  NOT NULL,
  score       REAL  NOT NULL,
  note        TEXT  NOT NULL DEFAULT '',
  at          TEXT  NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_items_state ON items(state);
`

// Named domain errors mapped from engine faults (REQ-TOSQ-006, design.md §6).
// Under NONE of them may recovery delete or overwrite either the database or
// any quarantined legacy artifact.
var (
	// ErrBacklogBusy maps SQLITE_BUSY surviving the busy window (another
	// process held the database beyond 5s) or the outer lock timing out.
	ErrBacklogBusy = errors.New("kanban backlog store busy")

	// ErrBacklogCorrupt maps an unreadable/malformed database image. The file
	// is NEVER deleted or rewritten by this store; operator action is
	// documented in the downgrade procedure.
	ErrBacklogCorrupt = errors.New("kanban backlog store corrupt")

	// ErrBacklogIDConflict maps a UNIQUE(id) constraint violation: the
	// transaction aborts and prior state stands intact.
	ErrBacklogIDConflict = errors.New("kanban backlog id conflict")
)

// IsBacklogBusy reports whether err is the busy sentinel.
func IsBacklogBusy(err error) bool { return errors.Is(err, ErrBacklogBusy) }

// IsBacklogCorrupt reports whether err is the corruption sentinel.
func IsBacklogCorrupt(err error) bool { return errors.Is(err, ErrBacklogCorrupt) }

// IsBacklogIDConflict reports whether err is the id-conflict sentinel.
func IsBacklogIDConflict(err error) bool { return errors.Is(err, ErrBacklogIDConflict) }

// classifyBacklogEngineCode maps a raw sqlite result code onto the named
// domain sentinels; nil-mapped codes fall through wrapped genericly. Primary
// codes are matched through the low byte so extended codes
// (SQLITE_BUSY_SNAPSHOT and friends) classify identically.
func classifyBacklogEngineCode(code int) error {
	switch code & 0xff {
	case sqlite3.SQLITE_BUSY:
		return ErrBacklogBusy
	case sqlite3.SQLITE_CORRUPT:
		return ErrBacklogCorrupt
	case sqlite3.SQLITE_NOTADB:
		return ErrBacklogCorrupt
	default:
		return nil
	}
}

// mapBacklogEngineError wraps err with op context and, when the driver
// surfaces a recognized result code, the matching named sentinel. The
// wrapping preserves the original chain for errors.Is callers.
func mapBacklogEngineError(op string, err error) error {
	if err == nil {
		return nil
	}
	var sqlErr interface{ Code() int }
	if errors.As(err, &sqlErr) {
		if named := classifyBacklogEngineCode(sqlErr.Code()); named != nil {
			return fmt.Errorf("%s: %w: %w", op, named, err)
		}
	}
	return fmt.Errorf("%s: %w", op, err)
}

// backlogSQLitePath derives the engine artifact path from a queue-file path:
// same directory, base name minus a trailing .json, plus .db. Deterministic
// and injectable-free — tests construct stores over temp paths and locate
// the database the same way production does.
func backlogSQLitePath(queuePath string) string {
	base := filepath.Base(queuePath)
	return filepath.Join(filepath.Dir(queuePath), strings.TrimSuffix(base, ".json")+backlogDBFileNameSuffix)
}

// backlogDSN builds the driver DSN applying the per-connection pragmas
// (REQ-TOSQ-003) and immediate write transactions. Values ride url.Values so
// paths carrying spaces or '?' survive encoding intact.
//
// A Windows drive path must enter the URI as a root-absolute slash form —
// C:\dir\db.db becomes /C:/dir/db.db, rendering as file:///C:/dir/db.db.
// Handed the raw path, url.URL.String() emits file://C:/… and the driver
// parses the drive colon as the URI authority, refusing every open
// ("invalid uri authority"). filepath.ToSlash is a no-op on POSIX, so unix
// DSNs keep their exact historical bytes.
func backlogDSN(dbPath string) string {
	v := url.Values{}
	v.Add("_pragma", fmt.Sprintf("busy_timeout(%d)", backlogBusyTimeoutMS))
	v.Add("_pragma", "journal_mode(WAL)")
	v.Add("_txlock", "immediate")
	p := filepath.ToSlash(dbPath)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	u := url.URL{Scheme: "file", Path: p, RawQuery: v.Encode()}
	return u.String()
}

// backlogEngine owns one database/sql handle over the queue database.
type backlogEngine struct {
	db     *sql.DB
	dbPath string
}

// openBacklogEngine opens (creating when absent) the database at dbPath,
// applies health checks, materializes the schema idempotently, and stamps
// schema_version when absent. An unrecognized stamped version refuses the
// open rather than operating against unknown bytes.
func openBacklogEngine(dbPath string) (*backlogEngine, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("open backlog store %s: creating dir: %w", dbPath, err)
	}
	db, err := sql.Open(sqliteDriverName, backlogDSN(dbPath))
	if err != nil {
		return nil, mapBacklogEngineError(fmt.Sprintf("open backlog store %s", dbPath), err)
	}
	db.SetMaxOpenConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), backlogOpenTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, mapBacklogEngineError(fmt.Sprintf("open backlog store %s", dbPath), err)
	}
	eng := &backlogEngine{db: db, dbPath: dbPath}
	if err := eng.ensureSchema(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return eng, nil
}

// ensureSchema executes the idempotent DDL and reconciles the
// schema_version marker. A stamped FUTURE or alien version aborts the open
// with ErrBacklogCorrupt semantics: refuse-to-operate, never repair-by-delete.
func (e *backlogEngine) ensureSchema(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, backlogOpenTimeout)
	defer cancel()
	if _, err := e.db.ExecContext(ctx, backlogDDL); err != nil {
		return mapBacklogEngineError(fmt.Sprintf("schema %s", e.dbPath), err)
	}
	version, err := e.schemaVersion(ctx)
	if err != nil {
		return err
	}
	switch version {
	case "":
		if _, err := e.db.ExecContext(ctx,
			`INSERT INTO meta(key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`,
			backlogMetaKeySchemaVersion, backlogSchemaVersion); err != nil {
			return mapBacklogEngineError(fmt.Sprintf("stamp schema_version %s", e.dbPath), err)
		}
	case backlogSchemaVersion:
		// current layout
	default:
		return fmt.Errorf("schema %s: unsupported schema_version %q (want %q): %w",
			e.dbPath, version, backlogSchemaVersion, ErrBacklogCorrupt)
	}
	return nil
}

// schemaVersion reads the stamped version from meta, empty when unstamped.
func (e *backlogEngine) schemaVersion(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, backlogOpenTimeout)
	defer cancel()
	var version string
	err := e.db.QueryRowContext(ctx,
		`SELECT value FROM meta WHERE key = ?`, backlogMetaKeySchemaVersion).Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", mapBacklogEngineError(fmt.Sprintf("read schema_version %s", e.dbPath), err)
	}
	return version, nil
}

// pragmas reads the effective journal_mode and busy_timeout off the live
// connection pool — the assertion surface for REQ-TOSQ-003.
func (e *backlogEngine) pragmas(ctx context.Context) (journalMode string, busyTimeout int, err error) {
	ctx, cancel := context.WithTimeout(ctx, backlogOpenTimeout)
	defer cancel()
	if err := e.db.QueryRowContext(ctx, `PRAGMA journal_mode`).Scan(&journalMode); err != nil {
		return "", 0, mapBacklogEngineError(fmt.Sprintf("read journal_mode %s", e.dbPath), err)
	}
	if err := e.db.QueryRowContext(ctx, `PRAGMA busy_timeout`).Scan(&busyTimeout); err != nil {
		return "", 0, mapBacklogEngineError(fmt.Sprintf("read busy_timeout %s", e.dbPath), err)
	}
	return journalMode, busyTimeout, nil
}

// close releases the pooled handle.
func (e *backlogEngine) close() error {
	if e == nil || e.db == nil {
		return nil
	}
	return e.db.Close()
}
