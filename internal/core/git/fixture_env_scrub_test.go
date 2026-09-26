package git

import (
	"path/filepath"
	"testing"
)

// TestFixturesIgnoreInheritedRepoEnv pins that this package's git test
// fixtures act on the directory they are given, not on the repository an
// inherited GIT_DIR / GIT_WORK_TREE names. git exports those into every hook,
// so running these tests from inside one (a pre-commit hook, for instance)
// would otherwise have the fixtures commit into the repository being committed
// to — the shape GH #1691 reported.
//
// Same pattern as TestExecHelpersIgnoreInheritedRepoEnv: the inherited
// variables live only inside a subtest, and the result is read after t.Setenv
// has restored them. Not parallel: t.Setenv forbids it.
func TestFixturesIgnoreInheritedRepoEnv(t *testing.T) {
	fixtures := []struct {
		name string
		run  func(t *testing.T, dir string, args ...string)
	}{
		{"runGit", func(t *testing.T, dir string, args ...string) { runGit(t, dir, args...) }},
		{"gitFixture", func(t *testing.T, dir string, args ...string) { gitFixture(t, dir, args...) }},
		{"runGitEnv", func(t *testing.T, dir string, args ...string) { runGitEnv(t, dir, nil, args...) }},
	}
	for _, f := range fixtures {
		t.Run(f.name, func(t *testing.T) {
			target := initTestRepo(t)
			decoy := initTestRepo(t)
			targetBefore := runGit(t, target, "rev-parse", "HEAD")
			decoyBefore := runGit(t, decoy, "rev-parse", "HEAD")

			t.Run("under inherited GIT_DIR", func(t *testing.T) {
				t.Setenv("GIT_DIR", filepath.Join(decoy, ".git"))
				t.Setenv("GIT_WORK_TREE", decoy)
				f.run(t, target, "commit", "--allow-empty", "-m", "t1227 probe")
			})

			if got := runGit(t, decoy, "rev-parse", "HEAD"); got != decoyBefore {
				t.Errorf("%s wrote into the repository named by the inherited GIT_DIR: decoy HEAD moved %s -> %s", f.name, decoyBefore, got)
			}
			if got := runGit(t, target, "rev-parse", "HEAD"); got == targetBefore {
				t.Errorf("%s did not commit in the directory it was given: target HEAD still %s", f.name, got)
			}
		})
	}
}
