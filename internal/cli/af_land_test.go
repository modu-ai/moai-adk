package cli

// The operational entry points of the merge queue (AC-AFL-014): `moai af land` lands an enqueued card and
// `moai af reconcile` settles the rows a restart left behind. Both run on a real ledger and a real git repository;
// the gate is a command run in the candidate worktree (`true` or `false` here, stand-ins: no build or test runs).

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// afQueuedCard is a claimed class A card with passing evidence, queued with a candidate commit that is a change
// on a branch off the trunk; the repository's trunk ref is checked out nowhere.
func afQueuedCard(t *testing.T) (db, repo, cand string, st *ledger.Store, epoch int64) {
	t.Helper()
	ctx := context.Background()
	db = filepath.Join(t.TempDir(), "af", "ledger.db")
	var err error
	st, err = ledger.Open(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })

	repo = t.TempDir()
	// The project directory is where the verbs look for the repository.
	t.Setenv("CLAUDE_PROJECT_DIR", repo)
	afGit(t, repo, "init", "-q", "-b", "work")
	if err := os.WriteFile(filepath.Join(repo, "base.txt"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	afGit(t, repo, "add", "-A")
	afGit(t, repo, "commit", "-q", "-m", "base")
	afGit(t, repo, "branch", "af/trunk")
	afGit(t, repo, "checkout", "-q", "-b", "c1", "af/trunk")
	if err := os.WriteFile(filepath.Join(repo, "a.txt"), []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	afGit(t, repo, "add", "-A")
	afGit(t, repo, "commit", "-q", "-m", "card t1")
	cand = afGit(t, repo, "rev-parse", "HEAD")
	afGit(t, repo, "checkout", "-q", "work")

	if err := st.AddCard(ctx, ledger.Card{ID: "t1", Title: "t1", Position: core.Position{Stage: core.StageReady}, Priority: 1,
		Spec: ledger.Spec{Class: "A", Paths: []string{"a.txt"}}}); err != nil {
		t.Fatal(err)
	}
	c, err := st.Claim(ctx, ledger.ClaimRequest{SessionID: "s-t1", Worktree: repo, PID: os.Getpid(), PIDStart: "start-t1"})
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{ledger.KindGreen, ledger.KindGate} {
		if err := st.AddEvidence(ctx, "t1", c.Epoch, ledger.Evidence{Kind: kind, TreeSHA: "tree-" + cand, SubjectSHA: cand, Command: "go test ./...", Result: ledger.ResultPass}); err != nil {
			t.Fatal(err)
		}
	}
	if err := st.Enqueue(ctx, "t1", c.Epoch, cand); err != nil {
		t.Fatal(err)
	}
	return db, repo, cand, st, c.Epoch
}

func afRowState(t *testing.T, st *ledger.Store) ledger.QueueState {
	t.Helper()
	rows, err := st.Queue(context.Background())
	if err != nil || len(rows) != 1 {
		t.Fatalf("queue: %+v %v", rows, err)
	}
	return rows[0].State
}

func TestAFLand_AGateThatFailsFailsTheCardAndLeavesTheTrunk(t *testing.T) {
	db, repo, _, st, _ := afQueuedCard(t)
	before := afGit(t, repo, "rev-parse", "refs/heads/af/trunk")
	var res struct {
		Outcome string `json:"outcome"`
		Reason  string `json:"reason"`
	}
	out := runAF(t, "land", "t1", "--db", db, "--gate", "false")
	if err := json.Unmarshal([]byte(out), &res); err != nil || res.Outcome != "failed" || res.Reason == "" {
		t.Fatalf("output %q: %+v %v, want outcome failed with a reason", out, res, err)
	}
	if got := afGit(t, repo, "rev-parse", "refs/heads/af/trunk"); got != before {
		t.Errorf("the trunk moved although the gate failed: %s -> %s", before, got)
	}
	if got := afRowState(t, st); got != ledger.QueueFailed {
		t.Errorf("the row is %s, want failed", got)
	}
}

func TestAFLand_RefusesACardThatIsNotQueued(t *testing.T) {
	db, _, _, _, _ := afQueuedCard(t)
	cmd := newAFCmd()
	cmd.SetArgs([]string{"land", "ghost", "--db", db, "--gate", "true"})
	cmd.SetOut(&strings.Builder{})
	cmd.SetErr(&strings.Builder{})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "not in the queue") {
		t.Errorf("landing a card that is not queued: %v, want an error saying it is not in the queue", err)
	}
}

func TestAFReconcile_SettlesALandingRowWhoseCandidateIsInTheTrunk(t *testing.T) {
	ctx := context.Background()
	db, repo, cand, st, epoch := afQueuedCard(t)
	// the dead lander had taken the row to landing and had moved the trunk to the candidate
	lock, err := st.AcquireLock(ctx, ledger.LockMerge, ledger.LockHolder{SessionID: "dead", PID: os.Getpid(), PIDStart: "start-dead"}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.QueueUpdate(ctx, "t1", ledger.CardEpoch(epoch, cand), ledger.QueueVerifying, "t0", ""); err != nil {
		t.Fatal(err)
	}
	if err := st.QueueUpdate(ctx, "t1", ledger.LockEpoch(lock, cand), ledger.QueueLanding, "t0", ""); err != nil {
		t.Fatal(err)
	}
	if err := st.ReleaseLock(ctx, ledger.LockMerge, lock); err != nil {
		t.Fatal(err)
	}
	afGit(t, repo, "update-ref", "refs/heads/af/trunk", cand)

	var res struct {
		Actions []struct {
			CardID string `json:"card_id"`
			From   string `json:"from"`
			To     string `json:"to"`
		} `json:"actions"`
	}
	out := runAF(t, "reconcile", "--db", db)
	if err := json.Unmarshal([]byte(out), &res); err != nil || len(res.Actions) != 1 || res.Actions[0].CardID != "t1" || res.Actions[0].To != "landed" {
		t.Fatalf("output %q: %+v %v, want one action settling t1 as landed", out, res, err)
	}
	if got := afRowState(t, st); got != ledger.QueueLanded {
		t.Errorf("the row is %s, want landed", got)
	}
}
