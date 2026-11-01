package cli

// AC-AFL-013 (SPEC-AGENTIC-FACTORY-LOCAL-001): a class A card goes claim -> RED evidence -> GREEN -> gates ->
// MERGING (Enqueue) -> landed -> DONE, with its lease released, and without passing REVIEWING (spec.risk is empty
// and the diff is under 400 lines). The steps are driven through the `moai af` verbs over a real ledger; landing
// is the merge queue (mergeq) on a real git repository. The ledger, the verbs, the git repository and the merge
// queue are real; two things are stand-ins: the gate runner (Verify is a function that returns nil, so no build
// or test runs in the candidate worktree) and the lock holder's process fingerprint ("start-mq" is a literal, not
// a probe result). The evidence records are written by hand through the verb, and the red record's result "pass"
// means the observation succeeded (the acceptance tests were seen failing), not that they passed.

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
	// The claim verb binds the session's heartbeat under the project directory: a temporary one keeps that out of the
	// real one. It is the repository itself, because the verbs take --repo only for the project's own repository.
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

	// landing is the merge queue's, reached through `moai af land`; the gate is a command run in the candidate
	// worktree (here `true`, a stand-in: no build or test runs)
	var landRes struct {
		CardID  string `json:"card_id"`
		Outcome string `json:"outcome"`
	}
	out := runAF(t, "land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/af/trunk", "--gate", "true")
	if err := json.Unmarshal([]byte(out), &landRes); err != nil || landRes.CardID != "t1" || landRes.Outcome != string(mergeq.OutcomeLanded) {
		t.Fatalf("af land output %q: %+v %v, want outcome landed", out, landRes, err)
	}
	if got := afGit(t, repo, "rev-parse", "refs/heads/af/trunk"); got == afGit(t, repo, "rev-parse", "refs/heads/work") {
		t.Error("the trunk did not move from the base")
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
