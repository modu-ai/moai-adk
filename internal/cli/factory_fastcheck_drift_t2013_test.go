package cli

// factory_fastcheck_drift_t2013_test.go — card t2013, card-review round 2 P2 #1:
// a fast-check step that exits 0 after it changed the card tree must not pass
// the check. The check speaks for the tree it started on; a tree (or HEAD) that
// moved while it ran is not that tree.

import (
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// TestT2013FastCheckTreeEditedByAStepHoldsAndMergesNothing commits a passing
// test into the card tree that rewrites the tracked base.txt while it runs, then
// completes the card: the check exits 0 on a tree that no longer matches the one
// it started on, so the card must be held and the release branch left alone.
func TestT2013FastCheckTreeEditedByAStepHoldsAndMergesNothing(t *testing.T) {
	f := affs2New(t, affs2Opts{})
	affs2Write(t, filepath.Join(f.cardWT, "calc", "mutate_test.go"),
		"package calc\n\nimport (\n\t\"os\"\n\t\"testing\"\n)\n\nfunc TestMutate(t *testing.T) {\n\tif err := os.WriteFile(\"../base.txt\", []byte(\"mutated while the check ran\\n\"), 0o644); err != nil {\n\t\tt.Fatal(err)\n\t}\n}\n")
	affs2Git(t, f.cardWT, "add", "calc")
	affs2Git(t, f.cardWT, "commit", "-q", "-m", "card change beta")
	if _, err := factory.RunRemeasure(f.root, f.cardWT, f.relBranch, "true"); err != nil {
		t.Fatalf("place re-measure record: %v", err)
	}

	_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
	t.Logf("complete: err=%v", err)
	if err == nil {
		t.Fatal("complete succeeded although a check step changed the card tree while it ran")
	}
	if got := affs2Git(t, f.relWT, "rev-parse", "HEAD"); got != f.relBefore {
		t.Errorf("release branch moved %s -> %s although the check ran on a tree that changed", f.relBefore, got)
	}
	if c := fcCard(t, f.root, "t1"); c.State != homestate.CardMergeReady {
		t.Errorf("card = %s, want merge-ready (held, not merged)", c.State)
	}
	affs2WindowReleased(t, f, "after the drifted check")
}
