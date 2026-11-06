package cli

// factory_af_flow_close_test.go — card t2071: the af-flow landing path. A card
// the leader integrated into LOCAL main (the merge never reached origin/main,
// the lane never walked `factory stage`) must close through the formal verbs
// only: `todo landed` → `factory rebind-evidence` → `factory approve` →
// `todo done`. No store is edited directly and the receipt gate stays binding.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// afFlowLocalMainRoot merges a card commit into the queue root's LOCAL main
// with the local-merge subject form and leaves origin/main at the base commit,
// so the delivering commit is reachable from `main` and not from origin/main.
func afFlowLocalMainRoot(t *testing.T, root string) (merge string) {
	t.Helper()
	// The af-flow primary checkout holds `main`; initGitRepo may have named the
	// branch otherwise, so name it explicitly.
	fcGit(t, root, "branch", "-M", "main")
	fcGit(t, root, "update-ref", "refs/remotes/origin/main", fcGit(t, root, "rev-parse", "HEAD"))
	if err := os.WriteFile(filepath.Join(root, "afflow.txt"), []byte("af-flow\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fcGit(t, root, "add", "afflow.txt")
	fcGit(t, root, "commit", "-q", "-m", "[t1] merge: card work into local main (card t1)")
	return fcGit(t, root, "rev-parse", "HEAD")
}

func TestAFFlowLocalMainCardClosesThroughFormalVerbs(t *testing.T) {
	root, uuid, _, _, _ := squashLandedFixture(t, "")
	emptyStageRow(t, root)
	merge := afFlowLocalMainRoot(t, root)

	if _, _, err := runTodo(t, "landed", "t1", "--sha", merge, "--ref", "main"); err != nil {
		t.Fatalf("todo landed --sha <local-main merge> --ref main: %v", err)
	}
	if out, err := rebindRun(t, "af-flow: card merged into local main, no stage walked"); err != nil {
		t.Fatalf("rebind-evidence from the local-main landing record: %v\nout: %s", err, out)
	}
	if got := fcCard(t, root, "t1").EvidenceSHA; got != merge {
		t.Fatalf("evidence_sha = %q, want the local-main merge commit %s", got, merge)
	}
	if _, _, err := runFactory(t, "approve", "t1", "--run", fcRun, "--issuer", "lead"); err != nil {
		t.Fatalf("approve: %v", err)
	}
	db := fcOpen(t, root)
	if err := db.VerifyApprovalReadonly(context.Background(), "t1", uuid); err != nil {
		t.Fatalf("done gate refused the formal chain: %v", err)
	}
	_ = db.Close()
	out, stderr, err := runTodo(t, "done", "t1")
	if err != nil {
		t.Fatalf("todo done: %v\nstdout: %s\nstderr: %s", err, out, stderr)
	}
	if !strings.Contains(out, "t1") {
		t.Errorf("todo done stdout = %q, want it to name the card", out)
	}
	_ = homestate.CardLeased
}

// The refusals an af-flow card meets before its landing is recorded must name
// the missing step, so the operator reaches the formal chain without reading
// the source (card t2071).
func TestAFFlowRefusalsNameTheLandingStep(t *testing.T) {
	t.Run("rebind_without_a_landing_record_names_todo_landed_with_ref", func(t *testing.T) {
		root, _, _, _, _ := squashLandedFixture(t, "")
		emptyStageRow(t, root)
		_, err := rebindRun(t, "no landing record yet")
		if err == nil || !strings.Contains(err.Error(), "cannot infer the audit phase") {
			t.Fatalf("err = %v, want the audit-phase refusal kept", err)
		}
		for _, want := range []string{"moai todo landed t1 --sha", "--ref"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("err = %q, want it to name %q", err, want)
			}
		}
	})

	t.Run("landed_on_the_default_ref_names_the_local_branch_form", func(t *testing.T) {
		root, _, _, _, _ := squashLandedFixture(t, "")
		merge := afFlowLocalMainRoot(t, root)
		_, _, err := runTodo(t, "landed", "t1", "--sha", merge)
		if err == nil || !strings.Contains(err.Error(), "is not reachable from origin/main") {
			t.Fatalf("err = %v, want the reachability refusal kept", err)
		}
		if !strings.Contains(err.Error(), "--ref main") {
			t.Errorf("err = %q, want it to name `--ref main` for a card merged into local main", err)
		}
	})
}
