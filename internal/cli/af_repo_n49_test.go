package cli

// N49 (audit of 1c05083f71) through the real `moai af` verbs, a real ledger and real git.
//
// N49: the commit-looking verbs (evidence, enqueue) bound the ledger without telling afBindRepository where the ledger
// file sits, so with the ledger file inside repository P, a project directory that is no repository and a lease worktree
// in repository W, the first bind took W and the card ran to DONE with its commit absent from P.

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// afN49Ledger opens the ledger file db with card t1 ready and nothing else: unbound, no lease. The returned store is
// closed with the test.
func afN49Ledger(t *testing.T, db string) *ledger.Store {
	t.Helper()
	st, err := ledger.Open(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if err := st.AddCard(context.Background(), ledger.Card{ID: "t1", Title: "t1", Position: core.Position{Stage: core.StageReady}, Priority: 1,
		Spec: ledger.Spec{Class: "A", Paths: []string{"t1.txt"}}}); err != nil {
		t.Fatal(err)
	}
	return st
}

// afN49Other is afN38Other with a commit of its own: two repositories made in the same second hold the very same commit
// otherwise, and a commit that is in both is no test of a repository that lacks it.
func afN49Other(t *testing.T) (repo, x string) {
	t.Helper()
	repo, _ = afN38Other(t)
	afGit(t, repo, "checkout", "-q", "x")
	if err := os.WriteFile(filepath.Join(repo, "only-here.txt"), []byte(repo+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	afGit(t, repo, "add", "-A")
	afGit(t, repo, "commit", "-q", "-m", "only here")
	x = afGit(t, repo, "rev-parse", "HEAD")
	afGit(t, repo, "checkout", "-q", "work")
	return repo, x
}

// afN49Claim claims the ready card through the verb and returns its epoch as the flag value.
func afN49Claim(t *testing.T, db, worktree string) string {
	t.Helper()
	out, errOut, err := afRunBoth("claim", "--db", db, "--session", "s1", "--worktree", worktree)
	if err != nil {
		t.Fatalf("claim: %v\nstdout %.300s\nstderr %.300s", err, out, errOut)
	}
	var got struct {
		Epoch json.Number `json:"epoch"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil || got.Epoch == "" {
		t.Fatalf("claim printed %.300s: %v", out, err)
	}
	return got.Epoch.String()
}

func afN49Bound(t *testing.T, st *ledger.Store) string {
	t.Helper()
	got, err := st.Repository(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return got
}

// afN49Refused runs the verb, which must be refused with a message that contains every want, leaving the ledger as it was
// and unbound.
func afN49Refused(t *testing.T, st *ledger.Store, what string, wants []string, args ...string) {
	t.Helper()
	before := afSnapshot(t, st)
	out, errOut, err := afRunBoth(args...)
	if err == nil {
		t.Errorf("%s: accepted, output %.200s %.200s", what, out, errOut)
		return
	}
	for _, want := range wants {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("%s: the refusal does not name %q: %v", what, want, err)
		}
	}
	if after := afSnapshot(t, st); !reflect.DeepEqual(before, after) {
		t.Errorf("%s: the refusal changed the ledger:\nbefore %+v\nafter  %+v", what, before, after)
	}
	if got := afN49Bound(t, st); got != "" {
		t.Errorf("%s: the refused call bound the ledger to %q", what, got)
	}
}

// The audit's reproduction (TestA4Bind_CommitCheckSkipsTheLedgerFileClue): the ledger file is inside repository P, the
// project directory is no repository (so the first verb binds nothing), the lease worktree is W. Evidence from W used to
// bind W and the card to run on to DONE with x absent from P.
func TestAFVerbs_N49CommitCheckAsksTheLedgerFileBeforeTheFirstBind(t *testing.T) {
	afSkipWithoutSh(t)
	p, px := afN38Other(t)
	w, x := afN49Other(t)
	idP, idW := afN38GitDir(t, p), afN38GitDir(t, w)
	db := filepath.Join(p, ".moai", "af", "ledger.db")
	st := afN49Ledger(t, db)
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	e := afN49Claim(t, db, w)
	if got := afN49Bound(t, st); got != "" {
		t.Fatalf("setup: claim from a project directory that is no repository bound the ledger to %q", got)
	}
	wants := []string{"not bound yet", idP, idW}
	afN49Refused(t, st, "evidence green from W", wants, "evidence", "t1", "--db", db, "--epoch", e, "--kind", "green", "--subject-sha", x, "--result", "pass")
	afN49Refused(t, st, "evidence gate from W", wants, "evidence", "t1", "--db", db, "--epoch", e, "--kind", "gate", "--subject-sha", x, "--result", "pass")
	afN49Refused(t, st, "enqueue from W", wants, "enqueue", "t1", "--db", db, "--epoch", e, "--candidate", x)
	cards, err := st.Cards(context.Background())
	if err != nil || len(cards) != 1 || cards[0].Position.Stage == core.StageDone {
		t.Errorf("cards %+v %v: the card must not be finished", cards, err)
	}
	// the project (P) is never required to hold a commit it lacks, and nothing was queued
	if exec.Command("git", "-C", p, "cat-file", "-e", x).Run() == nil {
		t.Error("repository P holds W's commit")
	}
	if rows, qerr := st.Queue(context.Background()); qerr != nil || len(rows) != 0 {
		t.Errorf("the queue must be empty: %+v %v", rows, qerr)
	}

	// the control: a lease worktree in the repository the ledger file sits in is not a disagreement
	db2 := filepath.Join(p, ".moai", "af2", "ledger.db")
	st2 := afN49Ledger(t, db2)
	e2 := afN49Claim(t, db2, p)
	if out, errOut, err := afRunBoth("evidence", "t1", "--db", db2, "--epoch", e2, "--kind", "green", "--subject-sha", px, "--result", "pass"); err != nil {
		t.Fatalf("evidence in the repository the ledger file sits in: %v\n%.300s %.300s", err, out, errOut)
	}
	if got, want := afN49Bound(t, st2), idP; got != want {
		t.Errorf("bound %q, want %q", got, want)
	}
}
