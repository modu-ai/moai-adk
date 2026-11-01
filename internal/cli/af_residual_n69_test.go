package cli

// N69 (audit a36ea8c368, the bullets after the first; residual classification): refusals whose text does not lead to the way
// out that exists. Both ways out work today (the tests assert that), and neither is in the message:
//   - a lease recorded before the repository moved names the old path; evidence without --repo then says "cannot tell which
//     repository <old path> belongs to (not in a git repository)", and `--repo <new path>` is the way on;
//   - a link left at the old path makes move-repository say the repository "still exists on disk, so it did not move"; the
//     way on is to remove the link, move, and make it again.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAFResidual_N69TheRefusalsOfAMovedRepositoryNameTheWayOut(t *testing.T) {
	afSkipWithoutSh(t)
	root := t.TempDir()

	t.Run("a lease recorded before the move", func(t *testing.T) {
		p, c := afResRepo(t, filepath.Join(root, "P"))
		db := filepath.Join(root, "central", "ledger.db")
		afResLedger(t, db)
		t.Setenv("CLAUDE_PROJECT_DIR", p)
		if text, err := afResVerb(db, "claim", "--session", "s", "--worktree", p, "--bind-repo", p); err != nil {
			t.Fatalf("claim: %v (%.200s)", err, text)
		}
		moved := filepath.Join(root, "Q")
		if err := os.Rename(p, moved); err != nil {
			t.Fatal(err)
		}
		t.Setenv("CLAUDE_PROJECT_DIR", moved)
		if text, err := afResVerb(db, "move-repository", "--reason", "moved", "--bind-repo", moved); err != nil {
			t.Fatalf("move-repository: %v (%.200s)", err, text)
		}
		args := []string{"evidence", "t1", "--epoch", "1", "--kind", "gate", "--subject-sha", c, "--result", "pass"}
		text, err := afResVerb(db, args...)
		if err == nil {
			t.Fatalf("setup: evidence without --repo went through, so the lease does not name the old path (%.200s)", text)
		}
		if !strings.Contains(text, "--repo") {
			t.Errorf("the refusal does not say that --repo gets on: %.300s", strings.Join(strings.Fields(err.Error()), " "))
		}
		if text, err := afResVerb(db, append(args, "--repo", moved)...); err != nil {
			t.Errorf("the way out (--repo %s) is refused: %v (%.200s)", moved, err, text)
		}
	})

	t.Run("a link left at the old path", func(t *testing.T) {
		p, _ := afResRepo(t, filepath.Join(root, "P2"))
		db := filepath.Join(root, "central", "ledger2.db")
		afResLedger(t, db)
		t.Setenv("CLAUDE_PROJECT_DIR", p)
		if text, err := afResVerb(db, "reap", "--bind-repo", p); err != nil {
			t.Fatalf("bind: %v (%.200s)", err, text)
		}
		moved := filepath.Join(root, "Q2")
		if err := os.Rename(p, moved); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(moved, p); err != nil {
			t.Fatal(err)
		}
		t.Setenv("CLAUDE_PROJECT_DIR", moved)
		text, err := afResVerb(db, "move-repository", "--reason", "moved", "--bind-repo", moved)
		if err == nil {
			t.Fatalf("setup: the move went through with a link at the old path (%.200s)", text)
		}
		if flat := strings.ToLower(err.Error()); !strings.Contains(flat, "link") {
			t.Errorf("the refusal does not say that what stands at the old path is a link: %.300s", strings.Join(strings.Fields(err.Error()), " "))
		}
		// the way on: remove the link, record the move, make the link again
		if err := os.Remove(p); err != nil {
			t.Fatal(err)
		}
		if text, err := afResVerb(db, "move-repository", "--reason", "moved", "--bind-repo", moved); err != nil {
			t.Errorf("the move is refused after the link is removed: %v (%.200s)", err, text)
		}
	})
}
