package mergeq

// Card t2027 finding 1: the tree the gate passed must be the commit that lands. A gate that leaves a tracked file
// of the check worktree modified, deleted or staged has judged a tree that differs from the merge result, so the
// card fails and the trunk stays. Untracked and git-ignored files are not part of the commit and are no reason to
// refuse.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const trackedChangeGate = "unit-gate"

func TestLand_GateThatModifiesATrackedFile_IsRefused(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(t *testing.T, dir string)
	}{
		{"modified", func(t *testing.T, dir string) { write(t, dir, "base.txt", "changed by the gate\n") }},
		{"deleted", func(t *testing.T, dir string) {
			if err := os.Remove(filepath.Join(dir, "base.txt")); err != nil {
				t.Fatal(err)
			}
		}},
		{"staged", func(t *testing.T, dir string) {
			write(t, dir, "base.txt", "staged by the gate\n")
			git(t, dir, "add", "base.txt")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := repo(t)
			cand := candidate(t, dir, "c1", "a.txt", "a\n")
			before := git(t, dir, "rev-parse", "refs/heads/trunk")
			q := &fakeQueue{}
			m := newMerger(t, dir, q, func(_ context.Context, d string) error {
				tc.mutate(t, d)
				return nil
			})
			m.Gate = trackedChangeGate

			res, err := m.Land(context.Background(), "t1", cand)
			if err != nil {
				t.Fatal(err)
			}
			if res.Outcome != OutcomeFailed {
				t.Fatalf("outcome %q (reason %q, trunkAfter %q), want %q: the gate changed a tracked file and the card still landed",
					res.Outcome, res.Reason, res.TrunkAfter, OutcomeFailed)
			}
			if !strings.Contains(res.Reason, "base.txt") || !strings.Contains(res.Reason, trackedChangeGate) {
				t.Errorf("reason %q names neither the changed path nor the gate %q", res.Reason, trackedChangeGate)
			}
			if res.TrunkAfter != "" {
				t.Errorf("TrunkAfter = %q, want empty for a card that did not land", res.TrunkAfter)
			}
			if after := git(t, dir, "rev-parse", "refs/heads/trunk"); after != before {
				t.Errorf("the trunk moved %s -> %s although the landing was refused", before, after)
			}
			// it failed through the existing failure path: the row is written failed with the same reason
			if last := q.states[len(q.states)-1]; last != StateFailed || q.failReason != res.Reason {
				t.Errorf("final state %v, recorded reason %q, want failed with %q", last, q.failReason, res.Reason)
			}
			if q.locks != 0 {
				t.Errorf("the landing lock was taken %d times for a refused card", q.locks)
			}
			if out := git(t, dir, "worktree", "list", "--porcelain"); strings.Count(out, "worktree ") != 1 {
				t.Errorf("the check worktree was left behind:\n%s", out)
			}
		})
	}
}

func TestLand_GateThatLeavesOnlyAnUntrackedFile_StillLands(t *testing.T) {
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	q := &fakeQueue{}
	m := newMerger(t, dir, q, func(_ context.Context, d string) error {
		write(t, d, "scratch-output.txt", "build output\n")
		return nil
	})
	m.Gate = trackedChangeGate

	res, err := m.Land(context.Background(), "t1", cand)
	if err != nil {
		t.Fatal(err)
	}
	if res.Outcome != OutcomeLanded {
		t.Fatalf("outcome %q (reason %q), want %q: an untracked file is not part of the commit", res.Outcome, res.Reason, OutcomeLanded)
	}
	if got := git(t, dir, "rev-parse", "refs/heads/trunk"); got != res.TrunkAfter {
		t.Errorf("trunk %s, want the merge result %s", got, res.TrunkAfter)
	}
}

func TestLand_GateThatTouchesOnlyAnIgnoredFile_StillLands(t *testing.T) {
	dir := repo(t)
	// the ignore rule is part of the candidate, so the check worktree has it as a tracked file
	cand := candidate(t, dir, "c1", ".gitignore", "gate-cache.log\n")
	q := &fakeQueue{}
	m := newMerger(t, dir, q, func(_ context.Context, d string) error {
		write(t, d, "gate-cache.log", "cache\n")
		if got := git(t, d, "status", "--porcelain", "--ignored"); !strings.Contains(got, "!! gate-cache.log") {
			t.Errorf("control: gate-cache.log is not ignored in the check worktree (status %q)", got)
		}
		return nil
	})
	m.Gate = trackedChangeGate

	res, err := m.Land(context.Background(), "t1", cand)
	if err != nil {
		t.Fatal(err)
	}
	if res.Outcome != OutcomeLanded {
		t.Fatalf("outcome %q (reason %q), want %q: an ignored file is not part of the commit", res.Outcome, res.Reason, OutcomeLanded)
	}
}
