package cli

// factory_aff_s2_dirty_t2013_test.go — card t2013, sync-audit finding F1
// (BLOCKING): REQ-FCR-026 (b) says the fast required check is of the card's HEAD
// tree, but the check runs `go build/vet/test` in the card's WORKING COPY while
// the merge step merges the card's COMMITTED HEAD. A card tree with uncommitted
// or untracked changes could therefore pass the check and land a broken commit
// on the release branch. The release route refuses such a tree before the check
// runs: a card hold with reason dirty-card-tree, everything else unchanged.
// Real git, the real merge step and the real `moai verify run` path throughout.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/verify"
)

// affs2DirtyNoCheckRecord asserts the fast check never ran against the tree in
// its current (dirty) state: no `moai verify run` record exists under its key.
func affs2DirtyNoCheckRecord(t *testing.T, f affs2Fix) {
	t.Helper()
	key, err := verify.Key(context.Background(), f.cardWT)
	if err != nil {
		t.Fatal(err)
	}
	store, err := auditreceipt.StoreRoot(f.cardWT)
	if err != nil {
		t.Fatal(err)
	}
	snap, err := verify.Load(store, key)
	if err == nil && snap != nil && len(snap.Checks) > 0 {
		t.Errorf("a moai verify run record exists for the dirty card tree (key %s): %+v — the check must not run", key, snap.Checks)
	}
}

// F1: the release route holds a card whose worktree carries uncommitted changes.
func TestAFFS2CompleteDirtyCardTreeHolds(t *testing.T) {
	cases := []struct {
		name     string
		dirty    func(t *testing.T, f affs2Fix) // makes the card tree dirty
		wantPath string                         // named in the hold detail
		reason   string                         // wanted hold reason; "" = clean/ignored, must merge
	}{
		{
			name: "modified tracked file",
			dirty: func(t *testing.T, f affs2Fix) {
				affs2Write(t, filepath.Join(f.cardWT, "calc", "calc.go"), "package calc\n\nfunc Add(a, b int) int { return a + b + 0 }\n")
			},
			wantPath: "calc/calc.go", reason: "dirty-card-tree",
		},
		{
			name: "untracked new file",
			dirty: func(t *testing.T, f affs2Fix) {
				affs2Write(t, filepath.Join(f.cardWT, "calc", "extra.go"), "package calc\n\nfunc Extra() int { return 1 }\n")
			},
			wantPath: "calc/extra.go", reason: "dirty-card-tree",
		},
		{
			name: "staged-only change",
			dirty: func(t *testing.T, f affs2Fix) {
				affs2Write(t, filepath.Join(f.cardWT, "calc", "staged.go"), "package calc\n\nfunc Staged() int { return 2 }\n")
				affs2Git(t, f.cardWT, "add", "calc/staged.go")
			},
			wantPath: "calc/staged.go", reason: "dirty-card-tree",
		},
		{
			name: "path with a space",
			dirty: func(t *testing.T, f affs2Fix) {
				affs2Write(t, filepath.Join(f.cardWT, "notes with space.txt"), "scratch\n")
			},
			wantPath: "notes with space.txt", reason: "dirty-card-tree",
		},
		{
			name: "status query fails closed",
			dirty: func(t *testing.T, f affs2Fix) {
				// A corrupt index makes `git status` fail while the committed branch
				// stays readable, so everything before the check still works.
				gitDir := affs2Git(t, f.cardWT, "rev-parse", "--absolute-git-dir")
				if err := os.WriteFile(filepath.Join(gitDir, "index"), []byte("not an index"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			reason: "card-tree-unreadable",
		},
		{
			name: "ignored file only still merges",
			dirty: func(t *testing.T, f affs2Fix) {
				exclude := affs2Git(t, f.cardWT, "rev-parse", "--path-format=absolute", "--git-path", "info/exclude")
				if err := os.MkdirAll(filepath.Dir(exclude), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(exclude, []byte("scratch.log\n"), 0o644); err != nil {
					t.Fatal(err)
				}
				affs2Write(t, filepath.Join(f.cardWT, "scratch.log"), "ignored output\n")
			},
		},
		{name: "clean tree still merges", dirty: func(t *testing.T, f affs2Fix) {}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := affs2New(t, affs2Opts{})
			tc.dirty(t, f)

			_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
			if tc.reason == "" {
				if err != nil {
					t.Fatalf("complete refused a card tree with nothing to commit: %v", err)
				}
				c := fcCard(t, f.root, "t1")
				if c.State != homestate.CardMergedLocal || c.MergeSHA == "" {
					t.Errorf("card = %s merge=%q, want merged-local", c.State, c.MergeSHA)
				}
				if holds := affs2Holds(t, f.root); len(holds) != 0 {
					t.Errorf("hold records on a mergeable card: %v", holds)
				}
				affs2WindowReleased(t, f, "after the merge")
				return
			}

			if err == nil {
				t.Fatal("complete succeeded although the card tree is not provably its committed HEAD")
			}
			t.Logf("complete refused: %v", err)
			if got := affs2Git(t, f.relWT, "rev-parse", "HEAD"); got != f.relBefore {
				t.Errorf("release branch moved %s -> %s", f.relBefore, got)
			}
			if got := affs2Git(t, f.root, "rev-parse", affs2ReleaseBranch); got != f.relBefore {
				t.Errorf("release branch ref moved %s -> %s", f.relBefore, got)
			}
			if c := fcCard(t, f.root, "t1"); c.State != homestate.CardMergeReady || c.Version != 1 {
				t.Errorf("card = %s v%d, want merge-ready v1 (a hold record, not a transition)", c.State, c.Version)
			}
			holds := affs2Holds(t, f.root)
			if len(holds) != 1 {
				t.Fatalf("hold records = %d, want 1: %v", len(holds), holds)
			}
			h := holds[0]
			if h["reason"] != tc.reason || h["card_head"] != f.cardHead {
				t.Errorf("hold = %v, want reason %s and card_head %s", h, tc.reason, f.cardHead)
			}
			if detail, _ := h["detail"].(string); tc.wantPath != "" && !strings.Contains(detail, tc.wantPath) {
				t.Errorf("hold detail %q does not name %q", detail, tc.wantPath)
			}
			if h["check_key"] != nil && h["check_key"] != "" {
				t.Errorf("hold carries a check_key %v although the check never ran", h["check_key"])
			}
			if tc.reason == "dirty-card-tree" {
				affs2DirtyNoCheckRecord(t, f)
			}
			affs2WindowReleased(t, f, "after the dirty-tree stop")
			// A second acquire succeeds: this call left no window behind.
			sdHoldWindow(t, f.root, "sess-next", "lane-2", f.relBranch, factory.BranchSourceConfig, f.relWT, "t9")
			affs2Untouched(t, f, "after the dirty-tree stop")
		})
	}
}

// The hold detail is bounded: a tree with many changed files names at most ten.
func TestAFFS2CompleteDirtyCardTreeDetailIsBounded(t *testing.T) {
	f := affs2New(t, affs2Opts{})
	for i := 0; i < 25; i++ {
		affs2Write(t, filepath.Join(f.cardWT, "scratch", "file"+string(rune('a'+i))+".txt"), "x\n")
	}
	if _, _, err := runFactory(t, "complete", "t1", "--run", fcRun); err == nil {
		t.Fatal("complete succeeded with 25 untracked files in the card tree")
	}
	holds := affs2Holds(t, f.root)
	if len(holds) != 1 || holds[0]["reason"] != "dirty-card-tree" {
		t.Fatalf("holds = %v, want one dirty-card-tree hold", holds)
	}
	detail, _ := holds[0]["detail"].(string)
	if n := strings.Count(detail, "scratch/file"); n != 10 {
		t.Errorf("detail names %d paths, want exactly 10 (bounded):\n%s", n, detail)
	}
	if !strings.Contains(detail, "commit or remove the changes in the card worktree, then run moai factory complete again") {
		t.Errorf("detail lacks the next-step text:\n%s", detail)
	}
}
