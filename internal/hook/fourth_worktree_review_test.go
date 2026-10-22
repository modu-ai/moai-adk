package hook

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/paths"
	"github.com/modu-ai/moai-adk/internal/worktreeroot"
)

func TestCurrentReviewRegisteredCardRequiresOwnedHome(t *testing.T) {
	for _, mode := range []string{"lookalike", "unregistered", "corrupt-owner", "registered"} {
		t.Run(mode, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv(paths.EnvHome, home)
			var tree string
			if mode == "lookalike" {
				tree = filepath.Join(t.TempDir(), "worktrees", "projects", "ordinary", "card")
				if err := os.MkdirAll(tree, 0o755); err != nil {
					t.Fatal(err)
				}
			} else {
				repo := initWorktreeTestRepo(t)
				base, err := worktreeroot.Root(repo)
				if err != nil {
					t.Fatal(err)
				}
				tree = filepath.Join(base, "card")
				if mode == "unregistered" {
					if err := os.Mkdir(tree, 0o755); err != nil {
						t.Fatal(err)
					}
				} else {
					runWorktreeTestGit(t, repo, "worktree", "add", "-q", "--detach", tree, "HEAD")
				}
				if mode == "corrupt-owner" {
					if err := os.WriteFile(filepath.Join(base, ".moai-project.json"), []byte("{"), 0o600); err != nil {
						t.Fatal(err)
					}
				}
			}
			want := ""
			if mode == "registered" {
				want = "card"
			}
			if got := cardIDFromRegisteredWorktree(tree); got != want {
				t.Fatalf("%s card ID = %q, want %q", mode, got, want)
			}
		})
	}
}
