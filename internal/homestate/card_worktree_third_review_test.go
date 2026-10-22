package homestate

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func thirdRebindGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := scrubbedGit(dir, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func thirdRebindWorld(t *testing.T) (repo, old, dest string, identity WorktreeRebindIdentity) {
	t.Helper()
	repo, old, dest = t.TempDir(), filepath.Join(t.TempDir(), "old"), filepath.Join(t.TempDir(), "new")
	thirdRebindGit(t, repo, "init", "-q", "--template=", "-b", "main")
	thirdRebindGit(t, repo, "-c", "user.name=Review", "-c", "user.email=rebind@example.invalid", "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-qm", "seed")
	thirdRebindGit(t, repo, "worktree", "add", "-qb", "WT-review-rebind", old)
	var err error
	identity, err = CaptureWorktreeRebindIdentity(old)
	if err != nil {
		t.Fatal(err)
	}
	thirdRebindGit(t, repo, "worktree", "move", old, dest)
	return
}

func TestReviewRebindRequiresActualRegisteredTarget(t *testing.T) {
	for _, shape := range []string{"copied-pointer", "foreign-repository", "changed-branch"} {
		t.Run(shape, func(t *testing.T) {
			repo, old, dest, identity := thirdRebindWorld(t)
			db := rebindOpen(t)
			rebindPick(t, db, "review-run", "t7", old)
			target := dest
			switch shape {
			case "copied-pointer":
				target = t.TempDir()
				pointer, err := os.ReadFile(filepath.Join(dest, ".git"))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(target, ".git"), pointer, 0600); err != nil {
					t.Fatal(err)
				}
			case "foreign-repository":
				_, _, target, _ = thirdRebindWorld(t)
			case "changed-branch":
				thirdRebindGit(t, dest, "checkout", "-q", "--detach", "HEAD")
			}
			if _, err := db.RebindCardWorktree(context.Background(), "t7", old, target, "review", rebindNow, identity); err == nil {
				t.Errorf("%s was accepted as the moved tree's identity", shape)
			}
			if card := rebindCard(t, db, "review-run", "t7"); card.WorktreePath != old {
				t.Errorf("refusal changed card binding to %s", card.WorktreePath)
			}
			if shape == "changed-branch" {
				thirdRebindGit(t, dest, "checkout", "-q", "WT-review-rebind")
			}
			t.Setenv("GIT_DIR", filepath.Join(repo, ".git"))
			if ok, err := defaultCardWorktreeTargetProbe(dest, identity); err != nil || !ok {
				t.Errorf("real moved target refused: %v, %v", ok, err)
			}
			if _, err := exec.LookPath("git"); err != nil {
				t.Fatal(err)
			}
		})
	}
}
