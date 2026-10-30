package hook

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
	"github.com/modu-ai/moai-adk/internal/config"
)

type beatCall struct {
	db, card string
	epoch    int64
}

// ledgerBeatFixture points the seam at a temp project and records the DB beats.
func ledgerBeatFixture(t *testing.T) (root string, calls *[]beatCall, now *time.Time, fail *error) {
	t.Helper()
	root = t.TempDir()
	t.Setenv(config.EnvClaudeProjectDir, root)
	if err := os.MkdirAll(filepath.Join(root, ".moai", "state"), 0o755); err != nil {
		t.Fatal(err)
	}
	clock := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	var rec []beatCall
	var failure error
	origBeat, origNow := ledgerBeat, ledgerHeartbeatNow
	ledgerHeartbeatNow = func() time.Time { return clock }
	ledgerBeat = func(db, card string, epoch int64) error {
		rec = append(rec, beatCall{db, card, epoch})
		return failure
	}
	t.Cleanup(func() { ledgerBeat, ledgerHeartbeatNow = origBeat, origNow })
	return root, &rec, &clock, &failure
}

func bind(t *testing.T, root, sess string) {
	t.Helper()
	if err := BindLedgerHeartbeat(root, sess, LedgerBinding{DB: "/db/ledger.db", CardID: "c1", Epoch: 7}); err != nil {
		t.Fatal(err)
	}
}

func TestLedgerHeartbeat_NoBindingNoWork(t *testing.T) {
	root, calls, _, _ := ledgerBeatFixture(t)
	LedgerHeartbeatSeam(&HookInput{SessionID: "s1", CWD: root})
	if len(*calls) != 0 {
		t.Fatalf("a session without a binding must not touch the ledger: %+v", *calls)
	}
	if _, err := os.Stat(filepath.Join(root, ".moai", "state", "ledger-heartbeat")); err == nil {
		t.Fatal("the seam must not create state for an unbound session")
	}
}

func TestLedgerHeartbeat_ThrottlesToTheStampFile(t *testing.T) {
	root, calls, now, _ := ledgerBeatFixture(t)
	bind(t, root, "s1")
	in := &HookInput{SessionID: "s1", CWD: root}

	LedgerHeartbeatSeam(in)
	if len(*calls) != 1 || (*calls)[0] != (beatCall{"/db/ledger.db", "c1", 7}) {
		t.Fatalf("first call = %+v", *calls)
	}
	*now = now.Add(LedgerHeartbeatEvery - time.Second)
	LedgerHeartbeatSeam(in)
	if len(*calls) != 1 {
		t.Fatalf("a stamp younger than %v must skip the DB write: %d calls", LedgerHeartbeatEvery, len(*calls))
	}
	*now = now.Add(2 * time.Second)
	LedgerHeartbeatSeam(in)
	if len(*calls) != 2 {
		t.Fatalf("a stamp older than %v must beat again: %d calls", LedgerHeartbeatEvery, len(*calls))
	}
}

func TestLedgerHeartbeat_StaleEpochDropsTheBinding(t *testing.T) {
	root, calls, now, fail := ledgerBeatFixture(t)
	bind(t, root, "s1")
	*fail = ledger.ErrStaleEpoch
	LedgerHeartbeatSeam(&HookInput{SessionID: "s1", CWD: root})
	*now = now.Add(time.Hour)
	LedgerHeartbeatSeam(&HookInput{SessionID: "s1", CWD: root})
	if len(*calls) != 1 {
		t.Fatalf("after ErrStaleEpoch the old epoch must not beat again: %d calls", len(*calls))
	}
}

func TestLedgerHeartbeat_OtherErrorsFailOpenAndStayThrottled(t *testing.T) {
	root, calls, _, fail := ledgerBeatFixture(t)
	bind(t, root, "s1")
	*fail = errors.New("locked")
	in := &HookInput{SessionID: "s1", CWD: root}
	LedgerHeartbeatSeam(in)
	LedgerHeartbeatSeam(in)
	if len(*calls) != 1 {
		t.Fatalf("a failed beat must not be retried inside the window: %d calls", len(*calls))
	}
}

func TestLedgerHeartbeat_SessionsAreIndependent(t *testing.T) {
	root, calls, _, _ := ledgerBeatFixture(t)
	bind(t, root, "s1")
	bind(t, root, "s2")
	LedgerHeartbeatSeam(&HookInput{SessionID: "s1", CWD: root})
	LedgerHeartbeatSeam(&HookInput{SessionID: "s2", CWD: root})
	if len(*calls) != 2 {
		t.Fatalf("each session keeps its own stamp: %d calls", len(*calls))
	}
}

func TestLedgerHeartbeat_SessionIDCannotEscapeTheStateDir(t *testing.T) {
	root, calls, _, _ := ledgerBeatFixture(t)
	if err := BindLedgerHeartbeat(root, "../evil", LedgerBinding{DB: "/db", CardID: "c", Epoch: 1}); err == nil {
		t.Fatal("a session id with a path separator must be refused")
	}
	LedgerHeartbeatSeam(&HookInput{SessionID: "../evil", CWD: root})
	if len(*calls) != 0 {
		t.Fatal("must not read a binding outside the state dir")
	}
}

// Wiring: both a user turn and a tool completion carry the beat.
func TestLedgerHeartbeat_WiredIntoPostToolUse(t *testing.T) {
	root, calls, _, _ := ledgerBeatFixture(t)
	bind(t, root, "s1")
	h := NewPostToolHandler()
	if _, err := h.Handle(t.Context(), &HookInput{SessionID: "s1", CWD: root, ToolName: "Read"}); err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 1 {
		t.Fatalf("PostToolUse beats = %d, want 1", len(*calls))
	}
}
