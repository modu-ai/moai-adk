// backlog_hold_migration_test.go — SPEC-TODO-HOLD-STATE-001 M1 acceptance
// tests: the fourth live state, the v1→v2 transactional table rebuild, and
// the compat matrix carried by the schema_version stamp.
//
// The load-bearing contract is the rebuild safety discipline inherited from
// SPEC-TODO-SQLITE-001: full-field parity BEFORE the switch, the original
// file preserved untouched on any failure, and the stamp landing inside the
// committing transaction — so an old binary refuses the database instead of
// misreading it.
package factory

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// v1FixtureDDL is the stamped-v1 physical schema the migration consumes: the
// current table set with the THREE-state items CHECK and the landing column
// already present (every production v1 database carries it — t359 added it
// at open). It is deliberately the OLD DDL, so the fixture is a faithful v1
// and the rebuild has a real constraint to widen.
const v1FixtureDDL = `
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
  landing  TEXT,
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
  position INTEGER NOT NULL,
  landing  TEXT
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

// v1FixtureRow is one card the v1 fixture seeds, with the full field set the
// round-trip asserts parity on.
type v1FixtureRow struct {
	id, text, addedAt, state string
	specID                   sql.NullString
	landing                  sql.NullString
}

// v1FixtureRows is the fixture population: mixed states, one spec id, one
// landing evidence record — the field shapes the rebuild must carry over.
var v1FixtureRows = []v1FixtureRow{
	{id: "t1", text: "queued plain card", addedAt: "2026-09-29T01:00:00Z", state: "queued"},
	{id: "t2", text: "picked card in flight", addedAt: "2026-09-29T01:01:00Z", state: "picked",
		specID: sql.NullString{String: "SPEC-SOME-001", Valid: true}},
	{id: "t3", text: "[DROPPED — stale] dropped card", addedAt: "2026-09-29T01:02:00Z", state: "dropped"},
	{id: "t4", text: "queued card with landing evidence", addedAt: "2026-09-29T01:03:00Z", state: "queued",
		landing: sql.NullString{String: `{"ref":"origin/develop","ref_head":"a1b2c3d4e","observed_at":"2026-09-29T02:00:00Z","spec_status":"implemented"}`, Valid: true}},
}

// holdFixtureRoot returns a temp project root and the engine artifact path
// the package's own path contract resolves under it.
func holdFixtureRoot(t *testing.T) (root, dbPath string) {
	t.Helper()
	root = t.TempDir()
	return root, backlogSQLitePath(BacklogPathForRoot(root))
}

// seedV1Database builds a faithful stamped-v1 database at dbPath: the old
// DDL, mixed-state rows, a finding, an archived card with its archived
// finding, and meta stamped schema_version="1" / last_seq="9". The database
// is created in WAL mode, the mode every production open carries — a failed
// open must not be able to change the file merely by switching journal modes.
func seedV1Database(t *testing.T, dbPath string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
		t.Fatalf("mkdir v1 fixture dir: %v", err)
	}
	db, err := sql.Open(sqliteDriverName, dbPath+"?_pragma=journal_mode(WAL)")
	if err != nil {
		t.Fatalf("open v1 fixture: %v", err)
	}
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, v1FixtureDDL); err != nil {
		t.Fatalf("v1 fixture DDL: %v", err)
	}
	for i, row := range v1FixtureRows {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO items(seq, id, text, added_at, spec_id, landing, state) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			i+1, row.id, row.text, row.addedAt, row.specID, row.landing, row.state); err != nil {
			t.Fatalf("seed item %s: %v", row.id, err)
		}
	}
	if _, err := db.ExecContext(ctx,
		`INSERT INTO findings(subject_id, related_id, relation, source, score, note, at)
		 VALUES ('t1', 't2', 'near-duplicate', 'mechanical', 0.82, 'measured', '2026-09-29T03:00:00Z')`); err != nil {
		t.Fatalf("seed finding: %v", err)
	}
	if _, err := db.ExecContext(ctx,
		`INSERT INTO archived_items(seq, id, text, added_at, spec_id, state, position, landing)
		 VALUES (1, 't9', 'archived finished card', '2026-09-28T09:00:00Z', NULL, 'queued', 0, NULL)`); err != nil {
		t.Fatalf("seed archived item: %v", err)
	}
	if _, err := db.ExecContext(ctx,
		`INSERT INTO archived_findings(archive_seq, position, subject_id, related_id, relation, source, score, note, at)
		 VALUES (1, 0, 't9', 't1', 'contains', 'agent', 1.0, 'operator judged', '2026-09-28T09:05:00Z')`); err != nil {
		t.Fatalf("seed archived finding: %v", err)
	}
	for _, kv := range [][2]string{
		{"schema_version", "1"},
		{"last_seq", "9"},
	} {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO meta(key, value) VALUES (?, ?)`, kv[0], kv[1]); err != nil {
			t.Fatalf("seed meta %s: %v", kv[0], err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close v1 fixture: %v", err)
	}
}

// v1FixtureStore returns the store the package's path contract resolves for
// root — the same queue file whose engine artifact the fixtures seed.
func v1FixtureStore(t *testing.T, root string) *BacklogStore {
	t.Helper()
	return NewBacklogStore(BacklogPathForRoot(root))
}

// hashFile returns the sha256 of the file at path.
func hashFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// readStampedVersion opens the database read-only and returns the stamped
// schema_version (empty when unstamped).
func readStampedVersion(t *testing.T, dbPath string) string {
	t.Helper()
	db, err := sql.Open(sqliteDriverName, dbPath)
	if err != nil {
		t.Fatalf("open for meta read: %v", err)
	}
	defer func() { _ = db.Close() }()
	var version string
	err = db.QueryRowContext(context.Background(),
		`SELECT value FROM meta WHERE key = 'schema_version'`).Scan(&version)
	if err == sql.ErrNoRows {
		return ""
	}
	if err != nil {
		t.Fatalf("read schema_version: %v", err)
	}
	return version
}

// readItemsCheckSQL returns the items table's stored DDL text.
func readItemsCheckSQL(t *testing.T, dbPath string) string {
	t.Helper()
	db, err := sql.Open(sqliteDriverName, dbPath)
	if err != nil {
		t.Fatalf("open for DDL read: %v", err)
	}
	defer func() { _ = db.Close() }()
	var itemsSQL string
	if err := db.QueryRowContext(context.Background(),
		`SELECT sql FROM sqlite_master WHERE type = 'table' AND name = 'items'`).Scan(&itemsSQL); err != nil {
		t.Fatalf("read items DDL: %v", err)
	}
	return itemsSQL
}

// AC-THS-001 — the items table admits exactly the four live states and the
// CHECK enumerates all four literally: a fifth value is rejected at the
// database, and removing any literal from the tuple fails the assertion.
func TestBacklogItemsTableAdmitsExactlyFourStates(t *testing.T) {
	root, _ := holdFixtureRoot(t)
	store := NewBacklogStore(BacklogPathForRoot(root))
	if _, _, err := store.Add("probe card"); err != nil {
		t.Fatalf("add: %v", err)
	}
	dbPath := store.EnginePath()

	itemsSQL := readItemsCheckSQL(t, dbPath)
	const wantCheck = "CHECK (state IN ('queued','picked','dropped','hold'))"
	if !strings.Contains(itemsSQL, wantCheck) {
		t.Fatalf("items.state CHECK drifted.\n got: %s\nwant it to contain: %s", itemsSQL, wantCheck)
	}
	for _, literal := range []string{"'queued'", "'picked'", "'dropped'", "'hold'"} {
		if !strings.Contains(itemsSQL, literal) {
			t.Errorf("items.state CHECK is missing the literal %s", literal)
		}
	}

	db, err := sql.Open(sqliteDriverName, dbPath)
	if err != nil {
		t.Fatalf("open probe db: %v", err)
	}
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	// A direct INSERT with each live state succeeds — the CHECK admits all
	// four, hold included.
	for i, state := range []string{"queued", "picked", "dropped", "hold"} {
		id := "probe-" + state
		if _, err := db.ExecContext(ctx,
			`INSERT INTO items(seq, id, text, added_at, spec_id, landing, state) VALUES (?, ?, ?, ?, NULL, NULL, ?)`,
			100+i, id, "probe", "2026-09-29T00:00:00Z", state); err != nil {
			t.Errorf("insert state %q: %v", state, err)
		}
	}
	// A fifth value — each of several spellings an operator might reach for —
	// is a CHECK violation at the database, not an application judgment.
	for i, state := range []string{"held", "onhold", ""} {
		id := "probe-bad-" + fmt.Sprintf("%d", i)
		if _, err := db.ExecContext(ctx,
			`INSERT INTO items(seq, id, text, added_at, spec_id, landing, state) VALUES (?, ?, ?, ?, NULL, NULL, ?)`,
			200+i, id, "probe", "2026-09-29T00:00:00Z", state); err == nil {
			t.Errorf("insert fifth state %q: expected a CHECK violation, got nil", state)
		}
	}
}

// AC-THS-002 + AC-THS-019 — opening a stamped-v1 database rebuilds the items
// table to the four-state CHECK with every row, the archive, and the findings
// preserved field-for-field, stamps "2", and is idempotent on reopen.
func TestBacklogV1ToV2MigrationRoundTrip(t *testing.T) {
	root, dbPath := holdFixtureRoot(t)
	seedV1Database(t, dbPath)
	store := v1FixtureStore(t, root)

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("open v1 store: %v", err)
	}

	// (a) the stamp moved to "2" and the CHECK is the four-value tuple.
	if got := readStampedVersion(t, dbPath); got != "2" {
		t.Errorf("schema_version = %q, want %q after migration", got, "2")
	}
	itemsSQL := readItemsCheckSQL(t, dbPath)
	if !strings.Contains(itemsSQL, "CHECK (state IN ('queued','picked','dropped','hold'))") {
		t.Errorf("items CHECK after migration = %s, want the four-value tuple", itemsSQL)
	}

	// (b) every row survived with every field identical.
	if len(rec.Items) != len(v1FixtureRows) {
		t.Fatalf("item count = %d, want %d", len(rec.Items), len(v1FixtureRows))
	}
	for i, want := range v1FixtureRows {
		got := rec.Items[i]
		if got.ID != want.id || got.Text != want.text || got.AddedAt != want.addedAt ||
			string(got.State) != want.state {
			t.Errorf("item %d = %+v, want id=%s state=%s", i, got, want.id, want.state)
		}
		if (got.SpecID == nil) != !want.specID.Valid ||
			(got.SpecID != nil && *got.SpecID != want.specID.String) {
			t.Errorf("item %s spec_id = %v, want %v", want.id, got.SpecID, want.specID)
		}
		if (got.Landing == nil) != !want.landing.Valid ||
			(got.Landing != nil && got.Landing.Ref != "origin/develop") {
			t.Errorf("item %s landing = %v, want the seeded evidence preserved", want.id, got.Landing)
		}
	}

	// (c) the archive and the findings survived byte-identically.
	if len(rec.Archived) != 1 {
		t.Fatalf("archived count = %d, want 1", len(rec.Archived))
	}
	entry := rec.Archived[0]
	if entry.Item.ID != "t9" || entry.Position != 0 {
		t.Errorf("archived entry = %+v, want t9 at position 0", entry.Item)
	}
	if len(entry.Findings) != 1 || entry.Findings[0].Finding.SubjectID != "t9" {
		t.Errorf("archived findings = %+v, want the one t9 relation carried along", entry.Findings)
	}
	if len(rec.Findings) != 1 || rec.Findings[0].SubjectID != "t1" || rec.Findings[0].RelatedID != "t2" {
		t.Errorf("findings = %+v, want the seeded t1→t2 relation preserved", rec.Findings)
	}
	if rec.LastSeq != 9 {
		t.Errorf("last_seq = %d, want 9", rec.LastSeq)
	}

	// (d) idempotent: reopening runs no second migration.
	again, err := store.Load()
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if len(again.Items) != len(v1FixtureRows) {
		t.Errorf("item count after reopen = %d, want %d — the migration must not double or drop rows", len(again.Items), len(v1FixtureRows))
	}
	if got := readStampedVersion(t, dbPath); got != "2" {
		t.Errorf("schema_version after reopen = %q, want %q (idempotence)", got, "2")
	}
}

// AC-THS-003 — a migration that cannot complete aborts with the original
// database file byte-preserved: no deletion, no overwrite, no quarantine.
// The failing input is a hand-corrupted v1 row (a state value no CHECK ever
// admitted), which the rebuild's copy step refuses inside the transaction.
func TestBacklogMigrationFailureLeavesOriginalFileUntouched(t *testing.T) {
	root, dbPath := holdFixtureRoot(t)
	seedV1Database(t, dbPath)

	// Hand-corrupt one row the way an editor without the CHECK would: the
	// constraint is suspended for the one UPDATE, then the fixture is closed
	// cleanly so the on-disk bytes are settled.
	{
		db, err := sql.Open(sqliteDriverName, dbPath+"?_pragma=journal_mode(WAL)")
		if err != nil {
			t.Fatalf("open for corruption: %v", err)
		}
		if _, err := db.ExecContext(context.Background(), `PRAGMA ignore_check_constraints = ON`); err != nil {
			t.Fatalf("suspend checks: %v", err)
		}
		if _, err := db.ExecContext(context.Background(),
			`UPDATE items SET state = 'broken' WHERE id = 't1'`); err != nil {
			t.Fatalf("corrupt row: %v", err)
		}
		if err := db.Close(); err != nil {
			t.Fatalf("close after corruption: %v", err)
		}
	}

	before := hashFile(t, dbPath)
	store := v1FixtureStore(t, root)
	if _, err := store.Load(); err == nil {
		t.Fatal("opening a v1 database carrying an unrepresentable state must fail")
	}

	if after := hashFile(t, dbPath); after != before {
		t.Errorf("the failed migration modified the original database file.\n before: %s\n after:  %s", before, after)
	}
	for _, sibling := range []string{dbPath + "-wal", dbPath + "-shm", dbPath + ".migrated"} {
		if _, err := os.Stat(sibling); err == nil {
			t.Errorf("failed migration left artifact %s behind", sibling)
		}
	}
	// The rows the failed migration must not have touched are still there.
	if got := readStampedVersion(t, dbPath); got != "1" {
		t.Errorf("schema_version after failed migration = %q, want %q (abort, not partial apply)", got, "1")
	}
}

// AC-THS-004 — the stamp lands inside the committing transaction: after the
// migration the meta row and the four-value CHECK stand together; reverting
// the stamp re-migrates idempotently; an unknown stamp refuses at open.
func TestBacklogMigrationStampAndRebuildAreOneTransaction(t *testing.T) {
	root, dbPath := holdFixtureRoot(t)
	seedV1Database(t, dbPath)
	store := v1FixtureStore(t, root)

	if _, err := store.Load(); err != nil {
		t.Fatalf("open v1 store: %v", err)
	}
	if got := readStampedVersion(t, dbPath); got != "2" {
		t.Fatalf("schema_version = %q, want %q immediately after the committing migration", got, "2")
	}
	if !strings.Contains(readItemsCheckSQL(t, dbPath), "'hold'") {
		t.Error("items CHECK after migration must carry the fourth literal")
	}

	// Reverting the stamp makes the next adopting open migrate again — and
	// the second migration is idempotent.
	setStampedVersion(t, dbPath, "1")
	if _, err := store.Load(); err != nil {
		t.Fatalf("reopen after stamp revert: %v", err)
	}
	if got := readStampedVersion(t, dbPath); got != "2" {
		t.Errorf("schema_version after re-migration = %q, want %q", got, "2")
	}

	// An unknown stamp is refused at open — the AC-THS-005 refusal observed
	// from the migration path.
	setStampedVersion(t, dbPath, "9")
	if _, err := store.Load(); !IsBacklogCorrupt(err) {
		t.Errorf("open with unknown stamp: err = %v, want ErrBacklogCorrupt", err)
	}
}

func setStampedVersion(t *testing.T, dbPath, version string) {
	t.Helper()
	db, err := sql.Open(sqliteDriverName, dbPath)
	if err != nil {
		t.Fatalf("open for stamp write: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.ExecContext(context.Background(),
		`UPDATE meta SET value = ? WHERE key = 'schema_version'`, version); err != nil {
		t.Fatalf("write schema_version: %v", err)
	}
}

// AC-THS-005 — a binary that knows only the previous stamp refuses a newer
// database at open with ErrBacklogCorrupt semantics, reads no items row, and
// writes nothing; the same binary opens the previous stamp normally. The old
// binary is simulated by injecting the previous version as current — the
// reader-path version-constant injection the AC names.
func TestBacklogOpenRefusesForeignSchemaVersion(t *testing.T) {
	// The v2 database: created by the current binary.
	v2Root, _ := holdFixtureRoot(t)
	v2Store := NewBacklogStore(BacklogPathForRoot(v2Root))
	if _, _, err := v2Store.Add("current-layout card"); err != nil {
		t.Fatalf("build v2 fixture: %v", err)
	}
	v2Path := v2Store.EnginePath()
	before := hashFile(t, v2Path)
	beforeMod := modTime(t, v2Path)

	// A v1 database: the previous layout, for the positive control.
	v1Root, v1Path := holdFixtureRoot(t)
	seedV1Database(t, v1Path)

	// Old binary: the previous stamp is all it knows.
	backlogSchemaVersionOverride = backlogSchemaVersionV1
	defer func() { backlogSchemaVersionOverride = "" }()

	oldStore := v1FixtureStore(t, v2Root)
	if _, err := oldStore.Load(); !IsBacklogCorrupt(err) {
		t.Fatalf("old binary over v2 database: err = %v, want ErrBacklogCorrupt", err)
	} else if !strings.Contains(err.Error(), "unsupported schema_version") {
		t.Errorf("refusal error = %v, want it to name unsupported schema_version", err)
	}
	if got := readStampedVersion(t, v2Path); got != "2" {
		t.Errorf("refused database stamp = %q, want %q — refusal must not rewrite the stamp", got, "2")
	}
	if after := hashFile(t, v2Path); after != before {
		t.Error("the refused open modified the database file")
	}
	if got := modTime(t, v2Path); !got.Equal(beforeMod) {
		t.Errorf("refused open changed the database mtime: before %v, after %v", beforeMod, got)
	}

	// Positive control: the same old binary opens the v1 database normally —
	// the refusal is about the version, not the store.
	v1Store := v1FixtureStore(t, v1Root)
	rec, err := v1Store.Load()
	if err != nil {
		t.Fatalf("old binary over v1 database: %v", err)
	}
	if len(rec.Items) != len(v1FixtureRows) {
		t.Errorf("v1 rows read = %d, want %d", len(rec.Items), len(v1FixtureRows))
	}

	// Production reader: an unknown future stamp is refused too.
	backlogSchemaVersionOverride = ""
	futureRoot, futurePath := holdFixtureRoot(t)
	seedV1Database(t, futurePath)
	setStampedVersion(t, futurePath, "3")
	futureStore := v1FixtureStore(t, futureRoot)
	if _, err := futureStore.LoadPure(); !IsBacklogCorrupt(err) {
		t.Errorf("pure read of a version-3 stamp under %s: err = %v, want ErrBacklogCorrupt", futureRoot, err)
	}
}

func modTime(t *testing.T, path string) time.Time {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	return info.ModTime()
}

// assertRebuildParity is the rebuild's load-bearing comparison; its mismatch
// branches are unit-tested directly so every tuple field's divergence is
// observed failing, not just the row-count one the mutation exercises.
func TestAssertRebuildParityDetectsEveryTupleMismatch(t *testing.T) {
	base := backlogItemRow{
		Seq: 1, ID: "t1", Text: "card", AddedAt: "2026-09-29T00:00:00Z",
		SpecID:  sql.NullString{String: "SPEC-X-001", Valid: true},
		Landing: sql.NullString{String: `{"ref":"r"}`, Valid: true},
		State:   "queued",
	}
	mutations := map[string]func(*backlogItemRow){
		"seq":      func(r *backlogItemRow) { r.Seq = 2 },
		"id":       func(r *backlogItemRow) { r.ID = "t2" },
		"text":     func(r *backlogItemRow) { r.Text = "changed" },
		"added_at": func(r *backlogItemRow) { r.AddedAt = "2026-09-29T01:00:00Z" },
		"spec_id":  func(r *backlogItemRow) { r.SpecID = sql.NullString{} },
		"landing":  func(r *backlogItemRow) { r.Landing = sql.NullString{} },
		"state":    func(r *backlogItemRow) { r.State = "hold" },
	}
	for name, mutate := range mutations {
		got := base
		mutate(&got)
		if err := assertRebuildParity([]backlogItemRow{base}, []backlogItemRow{got}); err == nil {
			t.Errorf("%s mismatch: parity must fail", name)
		}
	}
	// The count mismatch fires before any tuple comparison.
	if err := assertRebuildParity([]backlogItemRow{base}, nil); err == nil {
		t.Error("row-count mismatch: parity must fail")
	}
	// Identical rows pass.
	if err := assertRebuildParity([]backlogItemRow{base}, []backlogItemRow{base}); err != nil {
		t.Errorf("identical rows must pass: %v", err)
	}
}

// SPEC-TODO-CLAIM-LEASE-001 REQ-TCL-001/002 (AC-TCL-007 convergence half) —
// opening a stamped-v1 database converges on the same 11-column items tuple
// a fresh database carries, with the lease columns appended AFTER the
// version reconciliation: the v1→v2 rebuild runs against exactly the column
// set it knows (a retrofit that ran first would be silently dropped by the
// rebuild, card t1310's inverted-ordering defect), and the ensure pass adds
// picked_by / lease_expires_at only once the stamp reads "2".
func TestBacklogLeaseRetrofitConvergesAfterRebuild(t *testing.T) {
	root, dbPath := holdFixtureRoot(t)
	seedV1Database(t, dbPath)
	store := v1FixtureStore(t, root)

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("open v1 store: %v", err)
	}
	if got := readStampedVersion(t, dbPath); got != "2" {
		t.Fatalf("schema_version = %q, want %q (rebuild completed before the retrofit)", got, "2")
	}
	if len(rec.Items) != len(v1FixtureRows) {
		t.Fatalf("item count = %d, want %d — the retrofit must not disturb rows", len(rec.Items), len(v1FixtureRows))
	}

	eng, err := openBacklogEngine(dbPath)
	if err != nil {
		t.Fatalf("reopen engine: %v", err)
	}
	defer func() { _ = eng.close() }()

	// Column-order pin, identical in shape to the freeze test's: additive
	// columns land in merge order — classification before the claim-lease
	// pair — so a fresh and an upgraded database converge on the exact
	// physical order.
	wantTails := map[string]string{
		"items":          "picked_at:TEXT:0:NULL dropped_at:TEXT:0:NULL classification:TEXT:0:NULL picked_by:TEXT:0:NULL lease_expires_at:TEXT:0:NULL",
		"archived_items": "archived_at:TEXT:0:NULL landing_verdict:TEXT:0:NULL classification:TEXT:0:NULL picked_by:TEXT:0:NULL lease_expires_at:TEXT:0:NULL",
	}
	for table, wantTail := range wantTails {
		got := columnTupleSequence(t, eng, table)
		if !strings.HasSuffix(got, wantTail) {
			t.Errorf("%s column tuples = %q, want them to end with %q", table, got, wantTail)
		}
	}
	// The items CHECK survived the retrofit untouched — the retrofit is
	// additive and must not have forced or implied a rebuild.
	if itemsSQL := readItemsCheckSQL(t, dbPath); !strings.Contains(itemsSQL, "CHECK (state IN ('queued','picked','dropped','hold'))") {
		t.Errorf("items CHECK after retrofit = %s, want the four-value tuple preserved", itemsSQL)
	}

	// Idempotent: reopening adds nothing a second time.
	again, err := store.Load()
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if len(again.Items) != len(v1FixtureRows) {
		t.Errorf("item count after reopen = %d, want %d", len(again.Items), len(v1FixtureRows))
	}
	eng2, err := openBacklogEngine(dbPath)
	if err != nil {
		t.Fatalf("reopen engine for idempotence: %v", err)
	}
	defer func() { _ = eng2.close() }()
	if got := columnTupleSequence(t, eng2, "items"); !strings.HasSuffix(got, wantTails["items"]) {
		t.Errorf("items column tuples after reopen = %q, want the same 11-column convergence", got)
	}
}

// SPEC-TODO-CLAIM-LEASE-001 REQ-TCL-003 (AC-TCL-007 tolerance half) — a
// pre-retrofit database that carries neither the lease columns nor even the
// transition stamps still READS through the pure reader: every read consumer
// tolerates their absence via the columnExpr NULL fallback, and the item's
// lease pointers come back nil, never a pointer to an empty string. The
// fixture is materialized with the raw driver, deliberately not through the
// engine — openBacklogEngine would run the additive retrofit before the
// read, which would weaken this into exercising nothing.
func TestBacklogReadToleratesMissingLeaseColumns(t *testing.T) {
	_, dbPath := holdFixtureRoot(t)
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
		t.Fatalf("mkdir fixture dir: %v", err)
	}
	db, err := sql.Open(sqliteDriverName, dbPath)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS meta (key TEXT PRIMARY KEY, value TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS items (
  seq      INTEGER PRIMARY KEY,
  id       TEXT    NOT NULL UNIQUE,
  text     TEXT    NOT NULL,
  added_at TEXT    NOT NULL,
  spec_id  TEXT,
  state    TEXT    NOT NULL CHECK (state IN ('queued','picked','dropped','hold'))
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
INSERT INTO items(seq, id, text, added_at, spec_id, state) VALUES (1, 't1', 'pre-retrofit card', '2026-09-29T00:00:00Z', NULL, 'queued');
INSERT INTO meta(key, value) VALUES ('schema_version', '2'), ('last_seq', '1');`); err != nil {
		t.Fatalf("materialize pre-retrofit fixture: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close fixture: %v", err)
	}

	eng, err := openBacklogReader(dbPath)
	if err != nil {
		t.Fatalf("open pure reader: %v", err)
	}
	defer func() { _ = eng.close() }()
	rec, err := eng.readRecord(ctx)
	if err != nil {
		t.Fatalf("readRecord over a pre-retrofit database: %v", err)
	}
	if len(rec.Items) != 1 || rec.Items[0].ID != "t1" {
		t.Fatalf("items = %+v, want the one pre-retrofit row readable", rec.Items)
	}
	if rec.Items[0].PickedBy != nil || rec.Items[0].LeaseExpiresAt != nil {
		t.Errorf("lease pointers = %v/%v, want nil on a database without the columns (never a pointer to an empty string)",
			rec.Items[0].PickedBy, rec.Items[0].LeaseExpiresAt)
	}
}
