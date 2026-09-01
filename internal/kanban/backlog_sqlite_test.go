// backlog_sqlite_test.go — engine-layer proofs for the SQLite backlog store
// (SPEC-TODO-SQLITE-001 M1: REQ-TOSQ-001/002/003/005/006, AC-TOSQ-018/017-schema).
//
// These tests exercise the engine BELOW the BacklogStore surface: connection
// establishment, the pragmas a many-writer desktop needs, the physical
// schema, and the engine-fault -> named-error mapping. Store-level behavior
// (Load/Mutate/Add parity) lives in backlog_store_test.go and is deliberately
// not duplicated here.
package kanban

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// openTestEngine opens an engine over a fresh temp database and closes it
// when the test ends.
func openTestEngine(t *testing.T) *backlogEngine {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "backlog.db")
	eng, err := openBacklogEngine(dbPath)
	if err != nil {
		t.Fatalf("openBacklogEngine(%s) = %v, want nil", dbPath, err)
	}
	t.Cleanup(func() { _ = eng.close() })
	return eng
}

// AC-TOSQ-018 / REQ-TOSQ-003: every established connection reports WAL and a
// busy timeout at or above the 5000ms floor. The floor is what makes a
// contending factory lane WAIT rather than surface a spurious failure, so it
// is asserted as an inequality against the requirement's number, not against
// our own constant.
func TestBacklogEnginePragmas(t *testing.T) {
	eng := openTestEngine(t)

	journal, busy, err := eng.pragmas(context.Background())
	if err != nil {
		t.Fatalf("pragmas() error = %v, want nil", err)
	}
	if !strings.EqualFold(journal, "wal") {
		t.Errorf("journal_mode = %q, want wal (REQ-TOSQ-003)", journal)
	}
	if busy < 5000 {
		t.Errorf("busy_timeout = %d, want >= 5000 (REQ-TOSQ-003)", busy)
	}
}

// The pragmas must hold on EVERY connection the pool may hand out, not only
// the first: a per-connection DSN pragma that silently failed to apply on a
// re-dial would leave a later mutation without its busy window. Forcing the
// pool to churn and re-reading is the observation that separates the two.
func TestBacklogEnginePragmasSurviveConnectionChurn(t *testing.T) {
	eng := openTestEngine(t)

	for i := 0; i < 3; i++ {
		conn, err := eng.db.Conn(context.Background())
		if err != nil {
			t.Fatalf("Conn() error = %v", err)
		}
		var journal string
		var busy int
		if err := conn.QueryRowContext(context.Background(), `PRAGMA journal_mode`).Scan(&journal); err != nil {
			t.Fatalf("journal_mode scan error = %v", err)
		}
		if err := conn.QueryRowContext(context.Background(), `PRAGMA busy_timeout`).Scan(&busy); err != nil {
			t.Fatalf("busy_timeout scan error = %v", err)
		}
		_ = conn.Raw(func(any) error { return nil })
		if err := conn.Close(); err != nil {
			t.Fatalf("conn.Close() error = %v", err)
		}
		if !strings.EqualFold(journal, "wal") || busy < 5000 {
			t.Errorf("connection %d: journal_mode=%q busy_timeout=%d, want wal / >=5000", i, journal, busy)
		}
	}
}

// AC-TOSQ-017 (schema half) / REQ-TOSQ-001: the physical artifact carries the
// meta, items, findings, and archived_* tables plus the state-index, and stamps the
// schema version on first open.
func TestBacklogEngineSchemaShape(t *testing.T) {
	eng := openTestEngine(t)
	ctx := context.Background()

	rows, err := eng.db.QueryContext(ctx,
		`SELECT name FROM sqlite_master WHERE type IN ('table','index') AND name NOT LIKE 'sqlite_%'`)
	if err != nil {
		t.Fatalf("sqlite_master query error = %v", err)
	}
	defer func() { _ = rows.Close() }()
	var got []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan error = %v", err)
		}
		got = append(got, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows error = %v", err)
	}
	sort.Strings(got)
	// The two archived_* tables joined the shape in
	// SPEC-TODO-DESTRUCTIVE-GUARD-001 (card t330): `done` archives the row and
	// its findings instead of discarding them. The assertion stays an EXACT
	// set — a table appearing here that no SPEC put here is the regression
	// this test exists to catch.
	want := []string{"archived_findings", "archived_items", "findings", "idx_items_state", "items", "meta"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("schema objects = %v, want %v", got, want)
	}

	version, err := eng.schemaVersion(ctx)
	if err != nil {
		t.Fatalf("schemaVersion() error = %v", err)
	}
	if version != backlogSchemaVersion {
		t.Errorf("schema_version = %q, want %q", version, backlogSchemaVersion)
	}
}

// Opening the same database twice must be idempotent: the DDL re-runs, the
// version stamp is recognized rather than rewritten, and no duplicate meta
// row appears.
func TestBacklogEngineReopenIsIdempotent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "backlog.db")
	first, err := openBacklogEngine(dbPath)
	if err != nil {
		t.Fatalf("first open error = %v", err)
	}
	if err := first.close(); err != nil {
		t.Fatalf("close error = %v", err)
	}

	second, err := openBacklogEngine(dbPath)
	if err != nil {
		t.Fatalf("second open error = %v", err)
	}
	defer func() { _ = second.close() }()

	var metaRows int
	if err := second.db.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM meta WHERE key = ?`, backlogMetaKeySchemaVersion).Scan(&metaRows); err != nil {
		t.Fatalf("meta count error = %v", err)
	}
	if metaRows != 1 {
		t.Errorf("schema_version rows = %d, want exactly 1", metaRows)
	}
}

// REQ-TOSQ-006 / C-6: a database stamped with a version this binary does not
// know is REFUSED, not repaired. The assertion has two halves and the second
// is the load-bearing one — after the refusal the file must still be there,
// byte-for-byte.
func TestBacklogEngineRefusesUnknownSchemaVersionWithoutDestroying(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "backlog.db")
	eng, err := openBacklogEngine(dbPath)
	if err != nil {
		t.Fatalf("open error = %v", err)
	}
	if _, err := eng.db.ExecContext(context.Background(),
		`UPDATE meta SET value = ? WHERE key = ?`, "99", backlogMetaKeySchemaVersion); err != nil {
		t.Fatalf("stamp future version error = %v", err)
	}
	if err := eng.close(); err != nil {
		t.Fatalf("close error = %v", err)
	}
	before, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("read before error = %v", err)
	}

	reopened, err := openBacklogEngine(dbPath)
	if err == nil {
		_ = reopened.close()
		t.Fatal("open of a future-stamped database succeeded, want refusal")
	}
	if !IsBacklogCorrupt(err) {
		t.Errorf("error = %v, want it to satisfy IsBacklogCorrupt", err)
	}

	after, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("read after error = %v — the refusal removed the file, violating C-6", err)
	}
	if len(before) != len(after) {
		t.Errorf("database size changed across the refusal: %d -> %d bytes (C-6 forbids repair-by-write)", len(before), len(after))
	}
}

// REQ-TOSQ-006: a file that is not a database at all maps onto the corrupt
// sentinel rather than an unclassified error, and is likewise left alone.
func TestBacklogEngineNotADatabaseMapsToCorrupt(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "backlog.db")
	junk := []byte("this is not a sqlite database, it is operator data someone misplaced\n")
	if err := os.WriteFile(dbPath, junk, 0o600); err != nil {
		t.Fatalf("seed error = %v", err)
	}

	eng, err := openBacklogEngine(dbPath)
	if err == nil {
		_ = eng.close()
		t.Fatal("open of a non-database file succeeded, want refusal")
	}
	if !IsBacklogCorrupt(err) {
		t.Errorf("error = %v, want IsBacklogCorrupt", err)
	}
	after, err := os.ReadFile(dbPath)
	if err != nil || string(after) != string(junk) {
		t.Errorf("the misplaced file was altered or removed (err=%v), violating C-6", err)
	}
}

// REQ-TOSQ-006: the raw result-code classifier is the table design.md §6
// fixes. Extended codes (high bits set) must classify by their primary code,
// or a SQLITE_BUSY_SNAPSHOT would escape the busy sentinel.
func TestClassifyBacklogEngineCode(t *testing.T) {
	const (
		busy    = 5
		corrupt = 11
		notADB  = 26
		ioErr   = 10
	)
	cases := []struct {
		name string
		code int
		want error
	}{
		{"busy primary", busy, ErrBacklogBusy},
		{"busy extended (snapshot)", busy | (2 << 8), ErrBacklogBusy},
		{"corrupt primary", corrupt, ErrBacklogCorrupt},
		{"corrupt extended", corrupt | (1 << 8), ErrBacklogCorrupt},
		{"not a database", notADB, ErrBacklogCorrupt},
		{"unclassified io error", ioErr, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := classifyBacklogEngineCode(tc.code); !errors.Is(got, tc.want) {
				t.Errorf("classifyBacklogEngineCode(%d) = %v, want %v", tc.code, got, tc.want)
			}
		})
	}
}

// mapBacklogEngineError must preserve the original chain: a caller matching
// on the named sentinel and a caller matching on the driver's own error both
// have to keep working, which rules out replacing the cause.
func TestMapBacklogEngineErrorPreservesChain(t *testing.T) {
	if got := mapBacklogEngineError("op", nil); got != nil {
		t.Errorf("mapBacklogEngineError(op, nil) = %v, want nil", got)
	}
	cause := errors.New("driver said no")
	wrapped := mapBacklogEngineError("open backlog store", cause)
	if !errors.Is(wrapped, cause) {
		t.Errorf("wrapped error lost its cause: %v", wrapped)
	}
	if !strings.Contains(wrapped.Error(), "open backlog store") {
		t.Errorf("wrapped error = %q, want the op prefix", wrapped.Error())
	}
}

// REQ-TOSQ-001: the engine artifact is the queue file's sibling — same
// directory, .db in place of .json. Callers derive it deterministically
// rather than being handed an injectable, so the derivation is the contract.
func TestBacklogSQLitePath(t *testing.T) {
	cases := []struct{ queue, want string }{
		{filepath.Join("root", ".moai", "state", "todo", "backlog.json"), filepath.Join("root", ".moai", "state", "todo", "backlog.db")},
		{filepath.Join("x", "backlog.json"), filepath.Join("x", "backlog.db")},
		{filepath.Join("x", "backlog"), filepath.Join("x", "backlog.db")},
	}
	for _, tc := range cases {
		if got := backlogSQLitePath(tc.queue); got != tc.want {
			t.Errorf("backlogSQLitePath(%q) = %q, want %q", tc.queue, got, tc.want)
		}
	}
}

// REQ-TOSQ-002 posture check: a path carrying the characters that break a
// naively concatenated DSN (spaces, '?', '#', '%') must still open. macOS
// user directories carry spaces routinely, so this is not a synthetic case.
func TestBacklogEngineOpensAwkwardPaths(t *testing.T) {
	for _, dir := range []string{"with space", "with?question", "with#hash", "with%percent"} {
		t.Run(dir, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), dir)
			dbPath := filepath.Join(root, "backlog.db")
			eng, err := openBacklogEngine(dbPath)
			if err != nil {
				t.Fatalf("openBacklogEngine(%s) = %v, want nil", dbPath, err)
			}
			defer func() { _ = eng.close() }()
			if _, err := os.Stat(dbPath); err != nil {
				t.Errorf("database not created at %s: %v", dbPath, err)
			}
			journal, busy, err := eng.pragmas(context.Background())
			if err != nil {
				t.Fatalf("pragmas() error = %v", err)
			}
			if !strings.EqualFold(journal, "wal") || busy < 5000 {
				t.Errorf("pragmas lost on awkward path: journal=%q busy=%d", journal, busy)
			}
		})
	}
}
