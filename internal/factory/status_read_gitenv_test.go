package factory

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// t1710 — the two status reads below pin LC_ALL=C on top of os.Environ, so an
// inherited GIT_DIR / GIT_WORK_TREE still outranks the directory each command
// was given: reportedBranch answers about the caller's repository, and
// readBranchStatus errors against it. Both must read the tree they were
// handed, whatever the calling process exported (t1589's scrub, applied
// here).
func TestReportedBranchIgnoresInheritedGitEnv(t *testing.T) {
	root := t1589FactoryRepo(t)
	wt := filepath.Join(t.TempDir(), "wt")
	if out, err := exec.Command("git", "-C", root, "worktree", "add", "-q", "-b", "wt-branch", wt).CombinedOutput(); err != nil {
		t.Fatalf("git worktree add: %v\n%s", err, out)
	}

	other := t1589FactoryRepo(t)
	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	t.Setenv("GIT_WORK_TREE", other)

	got, err := reportedBranch(wt)
	if err != nil {
		t.Fatalf("reportedBranch(%s) error = %v", wt, err)
	}
	if got != "wt-branch" {
		t.Errorf("reportedBranch(%s) = %q, want %q (the inherited GIT_DIR named %s)", wt, got, "wt-branch", other)
	}
}

func TestReadBranchStatusIgnoresInheritedGitEnv(t *testing.T) {
	root := t1589FactoryRepo(t)
	if out, err := exec.Command("git", "-C", root, "checkout", "-q", "-b", "wt-branch").CombinedOutput(); err != nil {
		t.Fatalf("git checkout -b: %v\n%s", err, out)
	}
	specDir := filepath.Join(root, ".moai", "specs", "SPEC-T1710-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	spec := "---\nstatus: draft\n---\n# SPEC-T1710-001\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(spec), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	for _, args := range [][]string{
		{"add", "-A"},
		{"-c", "user.name=test", "-c", "user.email=test@example.invalid", "commit", "-q", "-m", "spec"},
	} {
		if out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	other := t1589FactoryRepo(t)
	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	t.Setenv("GIT_WORK_TREE", other)

	cs, err := readBranchStatus(root, "SPEC-T1710-001", "wt-branch", root)
	if err != nil {
		t.Fatalf("readBranchStatus error = %v (the inherited GIT_DIR named %s)", err, other)
	}
	if !cs.SpecFilePresent || cs.Status != "draft" {
		t.Errorf("readBranchStatus = %+v, want the spec read from %s's wt-branch (status draft)", cs, root)
	}
}
