package cli

// factory_checkpoint_rows_t2013_test.go — the checkpoint reads the card's
// CURRENT row (not a rebind-abandoned older one) and never reports another
// checkout's git state as the card's when the card worktree is absent
// (card t2013, sync-audit-2 F2 and F3).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// TestFactoryCheckpointPrefersLiveRowOverNewerAbandoned: a rebind leaves the
// source run's row `abandoned` with a NEWER updated_at than the surviving
// target row. The checkpoint must describe the live row.
func TestFactoryCheckpointPrefersLiveRowOverNewerAbandoned(t *testing.T) {
	f := newCheckpointFixture(t, "leased", "run") // run-cp, updated_at 2026-11-02
	db, err := homestate.OpenFactory(factoryRootOf(f.root))
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	_, err = db.DB.Exec(`INSERT INTO cards(run_id,card_id,state,version,updated_at,stage,worktree_path) VALUES('run-old',?,?,1,'2026-11-03T09:00:00Z',?,?)`,
		cpCard, "abandoned", "plan", "")
	if err != nil {
		t.Fatalf("seed abandoned row: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	out, err := runCheckpointCmd(t, cpCard)
	if err != nil {
		t.Fatalf("checkpoint: %v\n%s", err, out)
	}
	cp := parseCheckpoint(out)
	if cp.State != "leased" || cp.Stage != "run" {
		t.Errorf("checkpoint state/stage = %q/%q, want leased/run (the live row, not the newer abandoned one):\n%s", cp.State, cp.Stage, out)
	}
	if _, err := os.Stat(f.path()); err != nil {
		t.Errorf("checkpoint must be written in the live row's worktree: %v", err)
	}
}

// TestFactoryCheckpointAbsentWorktreeReportsNoForeignGitState: the recorded
// worktree is gone while the project root is a different git checkout with its
// own branch, head, dirty file, and upstream. None of that is the card's.
func TestFactoryCheckpointAbsentWorktreeReportsNoForeignGitState(t *testing.T) {
	f := newCheckpointFixture(t, "merge-ready", "merge-ready")
	// The project root becomes a distinct, dirty git checkout.
	cpGit(t, f.root, "init", "-b", "primary-branch")
	cpWrite(t, filepath.Join(f.root, "p.txt"), "p\n")
	cpGit(t, f.root, "add", "p.txt")
	cpGit(t, f.root, "commit", "-m", "primary base")
	cpWrite(t, filepath.Join(f.root, "dirty.txt"), "dirty\n")
	primaryHead := cpGit(t, f.root, "rev-parse", "HEAD")
	if err := os.RemoveAll(f.wt); err != nil {
		t.Fatal(err)
	}

	out, err := runCheckpointCmd(t, cpCard)
	if err != nil {
		t.Fatalf("checkpoint: %v\n%s", err, out)
	}
	cp := parseCheckpoint(out)
	if cp.Branch != "unknown" || cp.Head != "unknown" || cp.Dirty != -1 || cp.Unpushed != -1 {
		t.Errorf("absent worktree must leave git fields unknown, got branch=%q head=%q dirty=%d unpushed=%d:\n%s",
			cp.Branch, cp.Head, cp.Dirty, cp.Unpushed, out)
	}
	if strings.Contains(out, primaryHead) || strings.Contains(out, "primary-branch") {
		t.Errorf("the project root's git state leaked into the card checkpoint:\n%s", out)
	}
	if !strings.Contains(out, "worktree: absent\n") {
		t.Errorf("the checkpoint must say the worktree is absent:\n%s", out)
	}
	if cp.Dirty > 0 || checkpointNextAction(cp) == "run moai factory complete" {
		t.Errorf("next action %q must not follow from foreign git state", checkpointNextAction(cp))
	}
}
