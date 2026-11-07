package cli

// factory_close_landed_test.go — `moai factory close-landed`: the leader's one
// verb for the af-flow landed-card close chain (todo landed → rebind-evidence →
// approve → todo done), over the same fixture the formal-verbs test uses.

import (
	"strings"
	"testing"
)

// cardArchived reports whether the queue has moved card id out of the live set.
func cardArchived(t *testing.T, id string) bool {
	t.Helper()
	rec, err := newTodoStore().LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	for i := range rec.Items {
		if rec.Items[i].ID == id {
			return false
		}
	}
	return rec.ArchivedIndex(id) >= 0
}

func TestFactoryCloseLanded(t *testing.T) {
	t.Run("happy_path_runs_the_whole_chain", func(t *testing.T) {
		root, _, _, _, _ := squashLandedFixture(t, "")
		emptyStageRow(t, root)
		merge := afFlowLocalMainRoot(t, root)

		out, stderr, err := runFactory(t, "close-landed", "t1", "--sha", merge, "--run", fcRun, "--ref", "main")
		if err != nil {
			t.Fatalf("close-landed: %v\nstdout: %s\nstderr: %s", err, out, stderr)
		}
		for _, want := range []string{"is on main", "landed t1 ref=main", "t1 evidence_sha", "approved t1", "done t1"} {
			if !strings.Contains(out, want) {
				t.Errorf("output lacks %q:\n%s", want, out)
			}
		}
		if got := fcCard(t, root, "t1").EvidenceSHA; got != merge {
			t.Errorf("evidence_sha = %q, want %s", got, merge)
		}
		if !cardArchived(t, "t1") {
			t.Error("the card is still live after close-landed")
		}
	})

	t.Run("a_re_run_is_idempotent", func(t *testing.T) {
		root, _, _, _, _ := squashLandedFixture(t, "")
		emptyStageRow(t, root)
		merge := afFlowLocalMainRoot(t, root)
		if _, _, err := runFactory(t, "close-landed", "t1", "--sha", merge, "--run", fcRun, "--ref", "main"); err != nil {
			t.Fatalf("first close-landed: %v", err)
		}
		out, _, err := runFactory(t, "close-landed", "t1", "--sha", merge, "--run", fcRun, "--ref", "main")
		if err != nil {
			t.Fatalf("second close-landed: %v, want exit 0", err)
		}
		if !strings.Contains(out, "already done; nothing to do") {
			t.Errorf("output %q lacks the already-done note", out)
		}
	})

	t.Run("a_partly_run_chain_skips_the_satisfied_steps", func(t *testing.T) {
		root, _, _, _, _ := squashLandedFixture(t, "")
		emptyStageRow(t, root)
		merge := afFlowLocalMainRoot(t, root)
		if _, _, err := runTodo(t, "landed", "t1", "--sha", merge, "--ref", "main"); err != nil {
			t.Fatalf("todo landed: %v", err)
		}
		if out, err := rebindRun(t, "partial chain"); err != nil {
			t.Fatalf("rebind: %v\n%s", err, out)
		}
		out, _, err := runFactory(t, "close-landed", "t1", "--sha", merge, "--run", fcRun, "--ref", "main")
		if err != nil {
			t.Fatalf("close-landed: %v", err)
		}
		if !strings.Contains(out, "step 2 (todo landed) skipped") || !strings.Contains(out, "already bound to") {
			t.Errorf("output does not report the satisfied steps:\n%s", out)
		}
		if !cardArchived(t, "t1") {
			t.Error("the card is still live")
		}
	})

	t.Run("a_commit_not_on_the_target_is_refused_before_any_write", func(t *testing.T) {
		root, _, _, _, _ := squashLandedFixture(t, "")
		emptyStageRow(t, root)
		afFlowLocalMainRoot(t, root)
		// A commit no branch reaches: it is not an ancestor of main.
		stray := fcGit(t, root, "commit-tree", "HEAD^{tree}", "-p", "HEAD~1", "-m", "never landed")

		_, _, err := runFactory(t, "close-landed", "t1", "--sha", stray, "--run", fcRun, "--ref", "main")
		if err == nil || !strings.Contains(err.Error(), "step 1 (ancestry) refused") {
			t.Fatalf("err = %v, want the step-1 ancestry refusal", err)
		}
		if cardArchived(t, "t1") {
			t.Error("a refused close archived the card")
		}
		if got := fcCard(t, root, "t1").EvidenceSHA; got != "" {
			t.Errorf("evidence_sha = %q after a step-1 refusal, want it untouched", got)
		}
	})

	t.Run("a_later_refusal_names_its_step", func(t *testing.T) {
		root, _, _, _, _ := squashLandedFixture(t, "")
		emptyStageRow(t, root)
		merge := afFlowLocalMainRoot(t, root)
		// An unknown run: the rebind step has no row to bind.
		_, _, err := runFactory(t, "close-landed", "t1", "--sha", merge, "--run", "no-such-run", "--ref", "main")
		if err == nil || !strings.Contains(err.Error(), "step 3 (rebind-evidence) refused") {
			t.Fatalf("err = %v, want the step-3 refusal", err)
		}
		if cardArchived(t, "t1") {
			t.Error("a refused close archived the card")
		}
	})
}
