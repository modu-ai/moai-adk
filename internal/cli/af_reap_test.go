package cli

// `moai af reap` takes back the cards whose holder process is gone. The holder is the session that claimed
// (afHolder), so the tests use a real long-lived process as the session and a real kill as its death; the shell
// wrapper that ran the claim is never recorded and its exit changes nothing.

import (
	"context"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// startSession starts a long-lived process standing for the Claude Code session and guarantees it is gone
// when the test ends.
func startSession(t *testing.T) *exec.Cmd {
	t.Helper()
	cmd := exec.Command("sleep", "120")
	if err := cmd.Start(); err != nil {
		t.Skipf("no sleep binary to stand for a session: %v", err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
	return cmd
}

type reapOut struct {
	Reclaimed []struct {
		CardID       string `json:"CardID"`
		ReclaimCount int    `json:"ReclaimCount"`
	} `json:"reclaimed"`
}

func claimedBySession(t *testing.T, session *exec.Cmd) (db string, st *ledger.Store) {
	t.Helper()
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	ctx := context.Background()
	db = filepath.Join(t.TempDir(), "af", "ledger.db")
	st, err := ledger.Open(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if err := st.AddCard(ctx, ledger.Card{ID: "t1", Title: "t1", Position: core.Position{Stage: core.StageReady}, Priority: 1,
		Spec: ledger.Spec{Class: "A", Paths: []string{"a.txt"}}}); err != nil {
		t.Fatal(err)
	}
	origOwner := afOwnerPID
	t.Cleanup(func() { afOwnerPID = origOwner })
	afOwnerPID = func() (int, bool) { return session.Process.Pid, true } // afProbe stays real: the real fingerprint
	runAF(t, "claim", "--db", db, "--session", "s-reap", "--worktree", t.TempDir())
	return db, st
}

func TestAFReap_LiveSessionKeepsItsCard(t *testing.T) {
	session := startSession(t)
	db, st := claimedBySession(t, session)
	var out reapOut
	if err := json.Unmarshal([]byte(runAF(t, "reap", "--db", db)), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Reclaimed) != 0 {
		t.Fatalf("a live session's card was reclaimed: %+v", out.Reclaimed)
	}
	leases, err := st.Leases(context.Background())
	if err != nil || len(leases) != 1 || leases[0].CardID != "t1" {
		t.Fatalf("the lease must remain: %+v %v", leases, err)
	}
}

func TestAFReap_DeadSessionLosesItsCard(t *testing.T) {
	session := startSession(t)
	db, st := claimedBySession(t, session)
	_ = session.Process.Kill()
	_ = session.Wait()
	var out reapOut
	if err := json.Unmarshal([]byte(runAF(t, "reap", "--db", db)), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Reclaimed) != 1 || out.Reclaimed[0].CardID != "t1" || out.Reclaimed[0].ReclaimCount != 1 {
		t.Fatalf("reclaimed = %+v, want t1 once", out.Reclaimed)
	}
	leases, err := st.Leases(context.Background())
	if err != nil || len(leases) != 0 {
		t.Fatalf("the lease must be gone: %+v %v", leases, err)
	}
}
