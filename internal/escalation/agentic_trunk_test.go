package escalation

// agentic_trunk_test.go — SPEC-GIT-STRATEGY-AGENTIC-TRUNK-001 M2 (card
// t1831). AC-AGT-025 Judge 3: cardBase resolves the merge-base with main
// through the loader projection, even when origin/HEAD points at a branch
// other than main — the fixture separates the projection from the fallback.

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// agenticTrunkGit runs one git command in dir with a scrubbed identity.
func agenticTrunkGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_NAME=f", "GIT_AUTHOR_EMAIL=f@example.com", "GIT_COMMITTER_NAME=f", "GIT_COMMITTER_EMAIL=f@example.com")
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, errb.String())
	}
	return strings.TrimSpace(out.String())
}

func TestAgenticTrunkCardBase(t *testing.T) {
	root := t.TempDir()
	git := func(args ...string) string { return agenticTrunkGit(t, root, args...) }

	// A remote whose HEAD points at a branch other than main, so the
	// fallback chain (origin/HEAD first) would answer the wrong branch if
	// the projection failed.
	remote := filepath.Join(t.TempDir(), "origin.git")
	agenticTrunkGit(t, t.TempDir(), "init", "-q", "--bare", "-b", "notmain", remote)

	git("init", "-q", "-b", "main")
	git("config", "user.name", "t1831")
	git("config", "user.email", "t1831@test.local")
	if err := os.WriteFile(filepath.Join(root, "base.txt"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", ".")
	git("commit", "-q", "-m", "base")
	git("remote", "add", "origin", remote)
	git("push", "-q", "origin", "main")
	// Point the remote's HEAD at notmain and mirror the local view: a local
	// notmain branch that advanced beyond main, plus refs/remotes/origin/HEAD
	// resolving to notmain.
	git("branch", "notmain", "main")
	git("checkout", "-q", "notmain")
	if err := os.WriteFile(filepath.Join(root, "diverge.txt"), []byte("diverge\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", ".")
	git("commit", "-q", "-m", "notmain advances")
	git("checkout", "-q", "main")
	git("push", "-q", "origin", "notmain")
	git("remote", "set-head", "origin", "notmain")
	if head := git("symbolic-ref", "-q", "--short", "refs/remotes/origin/HEAD"); head != "origin/notmain" {
		t.Fatalf("fixture self-check: refs/remotes/origin/HEAD = %q, want origin/notmain", head)
	}

	// The agentic-trunk config, no develop_branch key: the projection names
	// main as the integration branch.
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"),
		[]byte("git_strategy:\n    mode: manual\n    manual:\n        workflow: agentic-trunk\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	base, err := cardBase(root)
	if err != nil {
		t.Fatalf("cardBase: %v", err)
	}
	want, err := gitOut(root, "merge-base", "HEAD", "main")
	if err != nil || want == "" {
		t.Fatalf("merge-base with main: %v %q", err, want)
	}
	if base != want {
		t.Errorf("cardBase = %s, want the merge-base with main %s (not the origin/HEAD fallback)", base, want)
	}
	if strings.TrimSpace(agenticTrunkGit(t, root, "rev-parse", "notmain")) == base {
		t.Error("cardBase resolved the origin/HEAD branch notmain — the projection did not win")
	}
}
