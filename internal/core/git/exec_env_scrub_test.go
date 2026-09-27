package git

import (
	"context"
	"path/filepath"
	"testing"
)

// TestExecHelpersIgnoreInheritedRepoEnv pins that both git executors act on
// the directory they are given, not on the repository an inherited GIT_DIR /
// GIT_WORK_TREE names. git exports those into every hook it runs, and they
// outrank cmd.Dir, so an executor that only appends to os.Environ() writes a
// hook-launched caller's commits into the caller's repository instead.
//
// The inherited variables are set inside a subtest and the result is read
// after it returns: t.Setenv restores them at subtest cleanup, and the fixture
// helpers read os.Environ() too, so reading earlier would ask the decoy.
// Not parallel: t.Setenv forbids it.
func TestExecHelpersIgnoreInheritedRepoEnv(t *testing.T) {
	helpers := []struct {
		name string
		run  func(dir string, args ...string) error
	}{
		{"execGit", func(dir string, args ...string) error {
			_, err := execGit(context.Background(), dir, args...)
			return err
		}},
		{"execGitExit", func(dir string, args ...string) error {
			_, err := execGitExit(context.Background(), dir, nil, args...)
			return err
		}},
	}
	for _, h := range helpers {
		t.Run(h.name, func(t *testing.T) {
			target := initTestRepo(t)
			decoy := initTestRepo(t)
			targetBefore := runGit(t, target, "rev-parse", "HEAD")
			decoyBefore := runGit(t, decoy, "rev-parse", "HEAD")

			t.Run("under inherited GIT_DIR", func(t *testing.T) {
				t.Setenv("GIT_DIR", filepath.Join(decoy, ".git"))
				t.Setenv("GIT_WORK_TREE", decoy)
				if err := h.run(target, "commit", "--allow-empty", "-m", "t1220 probe"); err != nil {
					t.Fatalf("commit in target: %v", err)
				}
			})

			if got := runGit(t, decoy, "rev-parse", "HEAD"); got != decoyBefore {
				t.Errorf("%s wrote into the repository named by the inherited GIT_DIR: decoy HEAD moved %s -> %s", h.name, decoyBefore, got)
			}
			if got := runGit(t, target, "rev-parse", "HEAD"); got == targetBefore {
				t.Errorf("%s did not commit in the directory it was given: target HEAD still %s", h.name, got)
			}
			if got := runGit(t, target, "log", "-1", "--format=%s"); got != "t1220 probe" {
				t.Errorf("%s: target's last commit subject = %q, want %q", h.name, got, "t1220 probe")
			}
		})
	}
}
