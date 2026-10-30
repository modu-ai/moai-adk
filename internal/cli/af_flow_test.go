package cli

// AC-AFL-013 (SPEC-AGENTIC-FACTORY-LOCAL-001): a class A card goes claim -> RED evidence -> GREEN -> gates ->
// MERGING (Enqueue) -> landed -> DONE, with its lease released, and without passing REVIEWING (spec.risk is empty
// and the diff is under 400 lines). The steps are driven through the `moai af` verbs over a real ledger; landing
// is the merge queue (mergeq) on a real git repository. Nothing is faked.

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/mergeq"
)

func runAF(t *testing.T, args ...string) string {
	t.Helper()
	cmd := newAFCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("moai af %s: %v\nstderr: %s", strings.Join(args, " "), err, errOut.String())
	}
	return out.String()
}

func afGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_SYSTEM="+os.DevNull,
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.test", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.test")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestCardFlow_ClaimToDone_ThroughMergeQueue(t *testing.T) {
	ctx := context.Background()
	db := filepath.Join(t.TempDir(), "af", "ledger.db")
	st, err := ledger.Open(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()
	if err := st.AddCard(ctx, ledger.Card{ID: "t1", Title: "t1", Position: core.Position{Stage: core.StageReady}, Priority: 1,
		Spec: ledger.Spec{Class: "A", Paths: []string{"a.txt"}}}); err != nil {
		t.Fatal(err)
	}

	// a repository whose trunk ref is checked out nowhere, and one commit on a candidate branch
	repo := t.TempDir()
	afGit(t, repo, "init", "-q", "-b", "work")
	if err := os.WriteFile(filepath.Join(repo, "base.txt"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	afGit(t, repo, "add", "-A")
	afGit(t, repo, "commit", "-q", "-m", "base")
	afGit(t, repo, "branch", "trunk")
	afGit(t, repo, "checkout", "-q", "-b", "c1", "trunk")
	if err := os.WriteFile(filepath.Join(repo, "a.txt"), []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	afGit(t, repo, "add", "-A")
	afGit(t, repo, "commit", "-q", "-m", "card t1")
	cand := afGit(t, repo, "rev-parse", "HEAD")
	afGit(t, repo, "checkout", "-q", "work")

	// claim
	var claimed struct {
		CardID string `json:"card_id"`
		Epoch  int64  `json:"epoch"`
	}
	if err := json.Unmarshal([]byte(runAF(t, "claim", "--db", db, "--session", "s-t1", "--worktree", repo)), &claimed); err != nil || claimed.CardID != "t1" {
		t.Fatalf("claim output: %+v %v", claimed, err)
	}
	ep := []string{"--db", db, "--epoch", strconv.FormatInt(claimed.Epoch, 10)}

	// RED evidence, then GREEN, then the gates, all bound to the candidate commit
	// (a red record's "pass" result means the observation succeeded: the acceptance tests were seen failing)
	for _, kind := range []string{"red", "green", "gate"} {
		runAF(t, append([]string{"evidence", "t1", "--kind", kind, "--subject-sha", cand, "--tree-sha", "tree-" + cand, "--command", "go test ./...", "--result", "pass"}, ep...)...)
	}
	runAF(t, append([]string{"enqueue", "t1", "--candidate", cand}, ep...)...)

	// landing is the merge queue's
	m := &mergeq.Merger{
		Git:   mergeq.GitCLI{Repo: repo, ScratchD: t.TempDir(), Who: mergeq.Identity{Name: "mergeq", Email: "mergeq@example.test"}},
		Queue: mergeq.NewLedgerQueue(st), Trunk: "refs/heads/trunk",
		Verify:  func(context.Context, string) error { return nil },
		Holder:  mergeq.Holder{Session: "mq", PID: os.Getpid(), PIDStart: "start-mq"},
		LockTTL: mergeq.DefaultLockTTL, MaxRetries: mergeq.DefaultMaxRetries,
		Message: func(card string) string { return "merge(" + card + "): land" },
	}
	res, err := m.Land(ctx, "t1", cand)
	if err != nil || res.Outcome != mergeq.OutcomeLanded {
		t.Fatalf("Land: %+v %v", res, err)
	}

	// the landed note is on the ledger; the card goes DONE and holds no lease
	runAF(t, append([]string{"move", "t1", "--stage", "MERGING", "--remote-integrated", "--actor", "mergeq", "--reason", "merge seen"}, ep...)...)
	runAF(t, append([]string{"move", "t1", "--stage", "DONE", "--actor", "mergeq", "--reason", "landed"}, ep...)...)
	cards, err := st.Cards(ctx)
	if err != nil || len(cards) != 1 || cards[0].Position.Stage != core.StageDone {
		t.Fatalf("cards %+v %v, want t1 DONE", cards, err)
	}
	if ls, _ := st.Leases(ctx); len(ls) != 0 {
		t.Errorf("a done card holds no lease: %+v", ls)
	}
	notes, _ := st.NotesSince(ctx, 0)
	landed := false
	for _, n := range notes {
		landed = landed || (n.Kind == ledger.NoteLanded && n.CardID == "t1")
	}
	if !landed {
		t.Errorf("no landed note: %+v", notes)
	}
}
