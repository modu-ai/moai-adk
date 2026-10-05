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
package factory

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
	// a SPEC that redefines the DDL; the runtime reads it and refuses an
	// unrecognized version rather than guessing. The "2" stamp (SPEC-TODO-
	// HOLD-STATE-001) marks the four-state items CHECK — a layout change
	// SQLite cannot ALTER into place, which is why it rides the rebuild
	// migration rather than an ADD COLUMN.
	backlogSchemaVersion = "2"

	// backlogSchemaVersionV1 is the previous stamp. It is a KNOWN version,
	// not a current one: the pure reader accepts it (an un-migrated queue is
	// readable), and the adopting open migrates it to the current stamp by
	// rebuilding the items table (backlog_rebuild.go). A stamp the binary
	// knows as neither current nor previous still refuses at open.
	backlogSchemaVersionV1 = "1"

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

// backlogSchemaVersionOverride, when non-empty, replaces the version the
// reader and the schema switch treat as current (test-only seam for
// REQ-THS-005's old-binary simulation — the AC sanctions reader-path version
// injection; always empty in production).
var backlogSchemaVersionOverride string

// backlogCurrentSchemaVersion returns the schema stamp this binary treats as
// current, honoring the test-only override.
func backlogCurrentSchemaVersion() string {
	if backlogSchemaVersionOverride != "" {
		return backlogSchemaVersionOverride
	}
	return backlogSchemaVersion
}

// backlogLandingColumn is the landing-evidence column carried by both
// card-bearing tables. It is added by ensureLandingColumn rather than by
// backlogDDL — see that function for why.
const backlogLandingColumn = "landing"

// backlogItemsTableColumns is the items table's physical column list — the
// single source shared by backlogDDL (fresh databases) and the v1→v2 rebuild
// (existing ones), so the two shapes cannot drift. The four-value state
// CHECK is the layout the "2" stamp marks; widening it is the change that
// forces the rebuild, because SQLite cannot ALTER a CHECK constraint.
// The landing column sits AFTER state, not before: upgraded databases carry
// it there (ensureLandingColumn's ALTER appends it), and a fresh database
// must converge on the identical physical order — the column-tuple freeze
// test pins the sequence.
const backlogItemsTableColumns = `
  seq      INTEGER PRIMARY KEY,
  id       TEXT    NOT NULL UNIQUE,
  text     TEXT    NOT NULL,
  added_at TEXT    NOT NULL,
  spec_id  TEXT,
  state    TEXT    NOT NULL CHECK (state IN ('queued','picked','dropped','hold')),
  landing  TEXT`

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
// first time a newer one opens it. This DDL once carried a THREE-state items
// CHECK and the comment warned that admitting a fourth state "would need a
// table rebuild on every operator queue in the field" — that rebuild is now
// exactly how the fourth state (`hold`) arrived: a transactional, parity-
// verified rebuild behind the schema_version bump to "2"
// (SPEC-TODO-HOLD-STATE-001, backlog_rebuild.go).
//
// archived_items deliberately carries NO state CHECK: an archived row is
// history rather than a live lifecycle position, and leaving the constraint
// off keeps the live four-value enum the single constrained surface.
const backlogDDL = `
CREATE TABLE IF NOT EXISTS meta (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS items (` + backlogItemsTableColumns + `
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
	ErrBacklogBusy = errors.New("todo queue store busy")

	// ErrBacklogCorrupt maps an unreadable/malformed database image. The file
	// is NEVER deleted or rewritten by this store; operator action is
	// documented in the downgrade procedure.
	ErrBacklogCorrupt = errors.New("todo queue store corrupt")

	// ErrBacklogIDConflict maps a UNIQUE(id) constraint violation: the
	// transaction aborts and prior state stands intact.
	ErrBacklogIDConflict = errors.New("todo queue id conflict")
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
	reader backlogQuery
}

type backlogQuery interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func (e *backlogEngine) queryDB() backlogQuery {
	if e.reader != nil {
		return e.reader
	}
	return e.db
}

// openBacklogReader never adopts a schema or changes permissions/journal mode.
// SQL is query-only. A checkpointed DB uses an existing-file rw connection so
// SQLite can clean up transient WAL coordination files on close; an active WAL
// uses ro so this reader does not checkpoint another connection's committed data.
func openBacklogReader(dbPath string) (*backlogEngine, error) {
	u, err := url.Parse(backlogDSN(dbPath))
	if err != nil {
		return nil, err
	}
	v := url.Values{}
	v.Set("mode", "ro")
	if !fileExists(dbPath+"-wal") && !fileExists(dbPath+"-shm") {
		v.Set("mode", "rw")
	}
	v.Add("_pragma", "query_only(ON)")
	v.Add("_pragma", fmt.Sprintf("busy_timeout(%d)", backlogBusyTimeoutMS))
	u.RawQuery = v.Encode()
	db, err := sql.Open(sqliteDriverName, u.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	e := &backlogEngine{db: db, dbPath: dbPath}
	ctx, cancel := context.WithTimeout(context.Background(), backlogOpenTimeout)
	defer cancel()
	version, err := e.schemaVersion(ctx)
	if err == nil && version != "" && version != backlogCurrentSchemaVersion() && version != backlogSchemaVersionV1 {
		err = fmt.Errorf("unsupported schema_version %q: %w", version, ErrBacklogCorrupt)
	}
	if err == nil {
		_, err = e.runtimeVersion(ctx)
	}
	if err == nil {
		_, err = e.identitySchemaPresent(ctx)
	}
	// Older releases published an empty DB before copying legacy records.
	// Neither a pure read nor an adopting read may call that empty state real.
	if err == nil && fileExists(strings.TrimSuffix(dbPath, ".db")+".json") {
		var seq string
		err = db.QueryRowContext(ctx, `SELECT value FROM meta WHERE key = ?`, backlogMetaKeyLastSeq).Scan(&seq)
		if errors.Is(err, sql.ErrNoRows) {
			err = fmt.Errorf("uninitialized database beside legacy queue; inspect migration before retrying: %w", ErrBacklogCorrupt)
		}
	}
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return e, nil
}

// openBacklogEngine opens (creating when absent) the database at dbPath,
// applies health checks, materializes the schema idempotently, and stamps
// schema_version when absent. An unrecognized stamped version refuses the
// open rather than operating against unknown bytes.
func openBacklogEngine(dbPath string) (*backlogEngine, error) {
	// Inspect existing schema before any writable connection can run pragmas/DDL.
	if info, err := os.Stat(dbPath); err == nil && info.Size() > 0 {
		probe, err := openBacklogReader(dbPath)
		if err != nil {
			return nil, err
		}
		_ = probe.close()
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
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
	if err := os.Chmod(dbPath, 0o600); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("open backlog store %s: secure database: %w", dbPath, err)
	}
	if err := secureBacklogArtifacts(dbPath); err != nil {
		_ = db.Close()
		return nil, err
	}
	return eng, nil
}

func secureBacklogArtifacts(dbPath string) error {
	for _, artifact := range []string{dbPath, dbPath + "-wal", dbPath + "-shm"} {
		if err := os.Chmod(artifact, 0o600); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("open backlog store %s: secure database artifact %s: %w", dbPath, artifact, err)
		}
	}
	return nil
}

// ensureSchema executes the idempotent DDL, adds any column the DDL cannot
// retrofit onto an existing table, and reconciles the schema_version marker.
// A stamped FUTURE or alien version aborts the open with ErrBacklogCorrupt
// semantics: refuse-to-operate, never repair-by-delete.
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
	current := backlogCurrentSchemaVersion()
	switch version {
	case "":
		if _, err := e.db.ExecContext(ctx,
			`INSERT INTO meta(key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`,
			backlogMetaKeySchemaVersion, current); err != nil {
			return mapBacklogEngineError(fmt.Sprintf("stamp schema_version %s", e.dbPath), err)
		}
	case current:
		// current layout
	case backlogSchemaVersionV1:
		// Reached only when the previous stamp is NOT this binary's current
		// one (the case above matches first): rebuild the items table to the
		// current CHECK and stamp — one transaction, parity-verified before
		// the switch (backlog_rebuild.go).
		if err := e.rebuildItemsTable(ctx); err != nil {
			return err
		}
	default:
		return fmt.Errorf("schema %s: unsupported schema_version %q (want %q): %w",
			e.dbPath, version, current, ErrBacklogCorrupt)
	}
	// Additive retrofits run AFTER the version reconciliation, never before
	// it: the v1→v2 rebuild must see exactly the column set it was written
	// against (a rebuild that runs after a retrofit would silently drop the
	// retrofitted columns it does not know), and a forced rebuild failure
	// must leave the database file byte-identical — an ALTER that ran first
	// would touch it before the failure could be observed. Fresh databases
	// carry every column in backlogDDL, so the ensures below are no-ops
	// there. Merge ordering fix, card t1310 (ensure-before-rebuild inverted).
	if err := e.ensureLandingColumn(ctx); err != nil {
		return err
	}
	if err := e.ensureTransitionStampColumns(ctx); err != nil {
		return err
	}
	for _, table := range []string{"items", "archived_items"} {
		if err := e.ensureColumn(ctx, table, backlogClassificationColumn); err != nil {
			return err
		}
	}
	// SPEC-TODO-CLAIM-LEASE-001 REQ-TCL-002: the lease retrofit is the LAST
	// additive pass, strictly after version reconciliation — a v1 database
	// rebuilds first, and only then do the lease columns arrive.
	if err := e.ensureLeaseColumns(ctx); err != nil {
		return err
	}
	// SPEC-TODO-CARD-ISSUANCE-001 REQ-TCI-007/010: the issuance attributes
	// ride the card-bearing tables, the finding disposition rides the
	// finding-bearing tables — all four through the same metadata-gated
	// ADD COLUMN path, AFTER the lease columns so the physical order
	// converges identically on a fresh and an upgraded database.
	if err := e.ensureIssuanceColumns(ctx); err != nil {
		return err
	}
	if err := e.ensureDispositionColumns(ctx); err != nil {
		return err
	}
	return nil
}

// schemaVersion reads the stamped version from meta, empty when unstamped.
func (e *backlogEngine) schemaVersion(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, backlogOpenTimeout)
	defer cancel()
	var version string
	err := e.queryDB().QueryRowContext(ctx,
		`SELECT value FROM meta WHERE key = ?`, backlogMetaKeySchemaVersion).Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", mapBacklogEngineError(fmt.Sprintf("read schema_version %s", e.dbPath), err)
	}
	return version, nil
}

// landingCarryingTables are the two card-bearing tables that carry the landing
// evidence column. archived_items mirrors items because an archived card keeps
// whatever evidence it held when it left the live queue.
var landingCarryingTables = []string{"items", "archived_items"}

// ensureLandingColumn adds the nullable landing TEXT column to both
// card-bearing tables when it is absent.
//
// It runs at open, after backlogDDL and before the schema_version switch.
// Both halves of that ordering matter: the tables do not exist until the DDL
// has run on a brand-new database, and the version switch is where the engine
// declares the database usable, so a column added after it would be missing
// for any caller that short-circuits there.
//
// The column lives here rather than in backlogDDL because CREATE TABLE IF NOT
// EXISTS adds no column to an existing table — the DDL alone cannot serve a
// database already in the field — and keeping it in exactly one place means a
// new and an upgraded database converge on the same statement.
//
// Idempotence is decided by reading the table's own column metadata, not by
// catching SQLite's "duplicate column name": that would turn a normal control
// path into an error path and would swallow a genuine failure sharing the
// message.
func (e *backlogEngine) ensureLandingColumn(ctx context.Context) error {
	for _, table := range landingCarryingTables {
		if err := e.ensureColumn(ctx, table, backlogLandingColumn); err != nil {
			return err
		}
	}
	return nil
}

// backlogTransitionStampColumns lists the transition-stamp columns
// (SPEC-TODO-TRANSITION-STAMPS-001) per table, added by the same additive
// migration discipline as the landing column. items carries picked_at and
// dropped_at; archived_items carries the same two stamps — the archive
// preserves them as they stood at archive time (REQ-TST-007) — plus
// archived_at and the done-time landing verdict record (REQ-TST-008).
// Compile-time constants only: nothing here is ever fed from a runtime value.
var backlogTransitionStampColumns = map[string][]string{
	"items":          {"picked_at", "dropped_at"},
	"archived_items": {"picked_at", "dropped_at", "archived_at", "landing_verdict"},
}

// backlogClassificationColumn is the card-classification column
// (SPEC-TODO-CLASSIFY-DISPATCH-001 REQ-TCD-002), appended to both
// card-bearing tables AFTER the stamp columns by the same
// pragma_table_info-gated ADD COLUMN path — a fresh and an upgraded database
// converge on the identical tuple (the freeze test pins it).
const backlogClassificationColumn = "classification"

// ensureTransitionStampColumns runs the stamp columns through the same
// metadata-gated ADD COLUMN path as the landing column, at the same point in
// the open sequence, so a fresh and an upgraded database converge on the same
// column set (REQ-TST-001..003). Iteration order is deterministic (one table
// at a time, columns in declaration order) so a partially failed open leaves
// an inspectable shape rather than an arbitrary one.
func (e *backlogEngine) ensureTransitionStampColumns(ctx context.Context) error {
	for _, table := range []string{"items", "archived_items"} {
		for _, column := range backlogTransitionStampColumns[table] {
			if err := e.ensureColumn(ctx, table, column); err != nil {
				return err
			}
		}
	}
	return nil
}

// backlogLeaseColumns lists the lease columns (SPEC-TODO-CLAIM-LEASE-001
// REQ-TCL-001) per table, added by the same additive migration discipline as
// the landing column and the transition stamps. items carries picked_by and
// lease_expires_at (RFC 3339); archived_items mirrors the same two — an
// archived card keeps the holder and expiry it held when it left the queue.
// They are deliberately NOT members of backlogItemsTableColumns: the v1→v2
// rebuild must never see them, and the retrofit runs only AFTER version
// reconciliation for the same reason (ensureSchema, REQ-TCL-002). Compile-
// time constants only.
var backlogLeaseColumns = map[string][]string{
	"items":          {"picked_by", "lease_expires_at"},
	"archived_items": {"picked_by", "lease_expires_at"},
}

// ensureLeaseColumns runs the lease columns through the metadata-gated ADD
// COLUMN path at the same point in the open sequence as the stamp columns,
// immediately after them so the physical order converges identically on a
// fresh and an upgraded database: the freeze test pins picked_by and
// lease_expires_at as the final two tuples of both card-bearing tables
// (REQ-TCL-015).
func (e *backlogEngine) ensureLeaseColumns(ctx context.Context) error {
	for _, table := range []string{"items", "archived_items"} {
		for _, column := range backlogLeaseColumns[table] {
			// SECURITY DISPOSITION (sync-phase scan pre-disposition): the
			// ALTER TABLE inside ensureColumn interpolates its table and
			// column identifiers rather than parameterizing them — SQLite
			// cannot parameterize DDL identifiers. Both values here are
			// compile-time constants from backlogLeaseColumns above; no user
			// or runtime input reaches the statement text. This is the
			// established in-repo additive-DDL pattern every retrofit
			// (landing, transition stamps) already runs through.
			if err := e.ensureColumn(ctx, table, column); err != nil {
				return err
			}
		}
	}
	return nil
}

// backlogIssuanceColumns lists the issuance column
// (SPEC-TODO-CARD-ISSUANCE-001 REQ-TCI-007) per card-bearing table, added by
// the same additive migration discipline as the lease columns. Compile-time
// constants only.
var backlogIssuanceColumns = map[string][]string{
	"items":          {"issuance"},
	"archived_items": {"issuance"},
}

// backlogDispositionColumns lists the finding disposition column
// (SPEC-TODO-CARD-ISSUANCE-001 REQ-TCI-010) per finding-bearing table —
// BOTH of them, because an archived finding that lost its disposition would
// break the archive round-trip the archived finding type exists to
// guarantee. Compile-time constants only.
var backlogDispositionColumns = map[string][]string{
	"findings":          {"disposition"},
	"archived_findings": {"disposition"},
}

// ensureIssuanceColumns runs the issuance columns through the
// metadata-gated ADD COLUMN path immediately after the lease columns, at the
// same point in the open sequence, so a fresh and an upgraded database
// converge on the identical tuple (the freeze test pins it).
func (e *backlogEngine) ensureIssuanceColumns(ctx context.Context) error {
	for _, table := range []string{"items", "archived_items"} {
		for _, column := range backlogIssuanceColumns[table] {
			// SECURITY DISPOSITION: compile-time constants only — see the
			// lease-columns call site for the established in-repo additive-DDL
			// pattern rationale.
			if err := e.ensureColumn(ctx, table, column); err != nil {
				return err
			}
		}
	}
	return nil
}

// ensureDispositionColumns runs the finding disposition columns through the
// same metadata-gated ADD COLUMN path, immediately after the issuance
// columns.
func (e *backlogEngine) ensureDispositionColumns(ctx context.Context) error {
	for _, table := range []string{"findings", "archived_findings"} {
		for _, column := range backlogDispositionColumns[table] {
			// SECURITY DISPOSITION: compile-time constants only — see the
			// lease-columns call site.
			if err := e.ensureColumn(ctx, table, column); err != nil {
				return err
			}
		}
	}
	return nil
}

// ensureColumn adds one nullable TEXT column to a table when it is absent —
// the single additive-migration gate both the landing column and the
// transition stamps run through. The interpolated table and column names are
// compile-time constants at every call site and may NEVER be fed from a
// runtime value; idempotence is decided by reading pragma_table_info, never
// by catching SQLite's duplicate-column error text.
func (e *backlogEngine) ensureColumn(ctx context.Context, table, column string) error {
	present, err := e.hasColumn(ctx, table, column)
	if err != nil {
		return err
	}
	if present {
		return nil
	}
	stmt := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s TEXT", table, column)
	if _, err := e.db.ExecContext(ctx, stmt); err != nil {
		return mapBacklogEngineError(
			fmt.Sprintf("add %s.%s %s", table, column, e.dbPath), err)
	}
	return nil
}

// hasColumn reports whether a table already declares the named column.
func (e *backlogEngine) hasColumn(ctx context.Context, table, column string) (bool, error) {
	var found string
	err := e.queryDB().QueryRowContext(ctx,
		`SELECT name FROM pragma_table_info(?) WHERE name = ?`, table, column).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, mapBacklogEngineError(
			fmt.Sprintf("read columns of %s %s", table, e.dbPath), err)
	}
	return true, nil
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
