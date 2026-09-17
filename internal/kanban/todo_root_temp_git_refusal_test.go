package kanban

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// t705 regression: a git repository physically under a temporary root is still
// a temporary origin, and the two layers that describe that decision must
// agree. StateDirForRoot resolves the queue project-local (the guard's whole
// purpose — a /tmp fixture that git-inits must not reach the home store), so
// TempOriginRefusal — the command path's read of the same decision — must
// report the refusal, not claim the home queue. Before the repair the refusal
// function consulted git first and reported refused=false while the resolver
// kept the queue local: guidance contradicted resolution.
func TestTempOriginRefusalMatchesStateDirForRootForTempGit(t *testing.T) {
	t.Setenv("MOAI_HOME", "") // no absolute override; the guard branch must decide

	dir := t.TempDir()
	gitDir := filepath.Join(dir, "project")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = gitDir
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")
	run("config", "user.email", "t@t")
	run("config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(gitDir, "README.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-q", "-m", "init")

	resolved := StateDirForRoot(gitDir)
	substitute, matchedRoot, refused := TempOriginRefusal(gitDir)
	reason, isTemp := TempOriginReason(gitDir)

	if !isTemp {
		t.Fatalf("fixture is not inside a temp root; the temp anchors regressed (matched %q)", reason)
	}
	if !pathWithin(resolved, dir) {
		t.Errorf("resolution side: temp-origin git repo resolved OUT of the temp root: %s", resolved)
	}
	if !refused {
		t.Errorf("guidance side: TempOriginRefusal reported refused=false while StateDirForRoot kept the queue project-local (%s)", resolved)
	}
	if substitute != gitDir {
		t.Errorf("TempOriginRefusal substitute = %q, want the repo dir %q", substitute, gitDir)
	}
	if matchedRoot == "" {
		t.Error("TempOriginRefusal named no matched temp root; guidance could not name what it matched")
	}
}
