package cli

// Defect answer (review finding P2): runWeb built web.Config without Ledger, so /board and /api/board always said
// "ledger reader not provided" even when the project's AF ledger existed. These tests pin that runWeb hands web.Run the
// SAME ledger afLedgerPath resolves, and that a project without a ledger file keeps a nil reader and gets no file made.

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
	"github.com/modu-ai/moai-adk/internal/web"
)

func pinWebLedgerEnv(t *testing.T) string {
	t.Helper()
	t.Setenv("MOAI_HOME", t.TempDir())
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	path, err := afLedgerPath("")
	if err != nil {
		t.Fatalf("afLedgerPath: %v", err)
	}
	return path
}

func TestRunWeb_ReadsTheProjectLedgerThatAFUses(t *testing.T) {
	root := t.TempDir()
	seams := materializeSeams(t, root) // sets its own MOAI_HOME; pin ours after it
	path := pinWebLedgerEnv(t)
	ctx := context.Background()
	st, err := ledger.Open(ctx, path)
	if err != nil {
		t.Fatalf("open the ledger: %v", err)
	}
	if err := st.AddCard(ctx, ledger.Card{
		ID: "w1", Title: "card w1", Position: core.Position{Stage: core.StageReady}, Priority: 1,
		Spec: ledger.Spec{Acceptance: []string{"TestX"}, Class: "A"},
	}); err != nil {
		t.Fatalf("add card: %v", err)
	}
	if err := st.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	var got web.Config
	runWebHarness(t, seams, withRootDir(root), func(c *runWebHarnessCfg) {
		c.captured = &got
		c.onRun = func(cfg web.Config) {
			if cfg.Ledger == nil {
				t.Error("runWeb gave web.Run no Ledger: the board would always read \"ledger reader not provided\"")
				return
			}
			cards, err := cfg.Ledger.Cards(ctx)
			if err != nil {
				t.Errorf("read cards through Config.Ledger while the server runs: %v", err)
				return
			}
			if len(cards) != 1 || cards[0].ID != "w1" {
				t.Errorf("Config.Ledger cards = %+v, want the one card w1", cards)
			}
		}
	})
	// The server has stopped: the ledger runWeb opened must be closed again.
	if got.Ledger != nil {
		if _, err := got.Ledger.Cards(ctx); err == nil {
			t.Error("the ledger is still readable after the server stopped: runWeb did not close it")
		}
	}
}

// webLedgerFixture creates a current-schema ledger with one card at the path `moai web` resolves and closes it.
func webLedgerFixture(t *testing.T) string {
	t.Helper()
	path := pinWebLedgerEnv(t)
	ctx := context.Background()
	st, err := ledger.Open(ctx, path)
	if err != nil {
		t.Fatalf("open the ledger: %v", err)
	}
	if err := st.AddCard(ctx, ledger.Card{
		ID: "w1", Title: "card w1", Position: core.Position{Stage: core.StageReady}, Priority: 1,
		Spec: ledger.Spec{Acceptance: []string{"TestX"}, Class: "A"},
	}); err != nil {
		t.Fatalf("add card: %v", err)
	}
	if err := st.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	return path
}

// setLedgerUserVersion rewrites PRAGMA user_version of the ledger file and leaves no -wal/-shm behind.
func setLedgerUserVersion(t *testing.T, path string, v int) {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open raw: %v", err)
	}
	if _, err := db.Exec(fmt.Sprintf("PRAGMA user_version=%d", v)); err != nil {
		t.Fatalf("set user_version: %v", err)
	}
	if _, err := db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		t.Fatalf("checkpoint: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close raw: %v", err)
	}
}

// ledgerFingerprint is the sha256 of the db file plus the size of its -wal (absent counts as 0). A WAL reader may leave
// empty -wal/-shm sidecars behind; what must never change is the db content and that no frame was written to the log.
func ledgerFingerprint(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	sum := sha256.Sum256(b)
	var walSize int64
	if info, err := os.Stat(path + "-wal"); err == nil {
		walSize = info.Size()
	}
	return fmt.Sprintf("%x|wal:%d", sum, walSize)
}

func ledgerUserVersion(t *testing.T, path string) int {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		t.Fatalf("open raw ro: %v", err)
	}
	defer func() { _ = db.Close() }()
	var v int
	if err := db.QueryRow("PRAGMA user_version").Scan(&v); err != nil {
		t.Fatalf("read user_version: %v", err)
	}
	return v
}

// (a) A ledger written by older code is never migrated by `moai web`: the board reads "not provided" instead.
func TestOpenWebLedger_OldSchemaIsNotMigrated(t *testing.T) {
	path := webLedgerFixture(t)
	setLedgerUserVersion(t, path, ledger.SchemaVersion-1)
	before := ledgerFingerprint(t, path)
	if got := ledgerUserVersion(t, path); got != ledger.SchemaVersion-1 {
		t.Fatalf("fixture user_version = %d, want %d", got, ledger.SchemaVersion-1)
	}

	r, closeFn := openWebLedger(context.Background())
	closeFn()
	if r != nil {
		t.Errorf("reader = %T, want nil for an old-schema ledger", r)
	}
	if got := ledgerUserVersion(t, path); got != ledger.SchemaVersion-1 {
		t.Errorf("user_version changed to %d: moai web migrated the ledger", got)
	}
	if after := ledgerFingerprint(t, path); after != before {
		t.Errorf("ledger files changed by moai web:\nbefore %s\nafter  %s", before, after)
	}
}

// (b) A ledger with a newer schema yields a nil reader and is left untouched.
func TestOpenWebLedger_NewerSchemaYieldsNilReaderAndIsUntouched(t *testing.T) {
	path := webLedgerFixture(t)
	setLedgerUserVersion(t, path, ledger.SchemaVersion+7)
	before := ledgerFingerprint(t, path)

	r, closeFn := openWebLedger(context.Background())
	closeFn()
	if r != nil {
		t.Errorf("reader = %T, want nil for a newer-schema ledger", r)
	}
	if after := ledgerFingerprint(t, path); after != before {
		t.Errorf("ledger files changed by moai web:\nbefore %s\nafter  %s", before, after)
	}
}

// (c) A current-schema ledger is read, and reading leaves the files unchanged.
func TestOpenWebLedger_CurrentSchemaReadsCardsWithoutChangingFiles(t *testing.T) {
	path := webLedgerFixture(t)
	before := ledgerFingerprint(t, path)

	r, closeFn := openWebLedger(context.Background())
	if r == nil {
		t.Fatal("reader is nil for a current-schema ledger")
	}
	cards, err := r.Cards(context.Background())
	if err != nil || len(cards) != 1 || cards[0].ID != "w1" {
		t.Errorf("Cards = %+v, err %v; want the one card w1", cards, err)
	}
	if _, err := r.Notes(context.Background(), 0); err != nil {
		t.Errorf("Notes: %v", err)
	}
	closeFn()
	if after := ledgerFingerprint(t, path); after != before {
		t.Errorf("ledger files changed by reading:\nbefore %s\nafter  %s", before, after)
	}
}

// (d) The reader moai web holds cannot write, even through the concrete store behind the interface.
func TestOpenWebLedger_ReaderCannotWrite(t *testing.T) {
	webLedgerFixture(t)
	r, closeFn := openWebLedger(context.Background())
	defer closeFn()
	if r == nil {
		t.Fatal("reader is nil for a current-schema ledger")
	}
	w, ok := r.(interface {
		AddNote(ctx context.Context, n ledger.Note) error
	})
	if !ok {
		return // read-only by construction
	}
	if err := w.AddNote(context.Background(), ledger.Note{Kind: ledger.NoteDecision, Text: "x"}); err == nil {
		t.Error("a write through the reader succeeded: the ledger is not opened read-only")
	}
}

func TestRunWeb_NoLedgerFileKeepsNilReaderAndCreatesNothing(t *testing.T) {
	root := t.TempDir()
	seams := materializeSeams(t, root)
	path := pinWebLedgerEnv(t)
	var got web.Config
	runWebHarness(t, seams, withRootDir(root), func(c *runWebHarnessCfg) { c.captured = &got })
	if got.Ledger != nil {
		t.Errorf("Config.Ledger = %T, want nil when the project has no ledger file", got.Ledger)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("moai web created or touched the ledger file %s (stat err: %v)", path, err)
	}
}
