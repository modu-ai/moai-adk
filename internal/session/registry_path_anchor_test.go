package session

import (
	"os/exec"
	"path/filepath"
	"testing"
)

// TestRegistryPathForAnchorsWorktreeToPrimaryCheckout is the regression guard
// for GH #1711 defect 2: a linked worktree used to get a registry of its own,
// so a lane registering from inside one was invisible to every reader anchored
// to the primary checkout.
//
// The mutant this catches: reverting RegistryPathFor to filepath.Join(dir,
// DefaultRegistryPath) makes the worktree case resolve inside the worktree,
// which is exactly the defect.
func TestRegistryPathForAnchorsWorktreeToPrimaryCheckout(t *testing.T) {
	t.Parallel()
	requireGit(t)

	primary := canonical(t, t.TempDir())
	runGit(t, primary, "init", "-q", "-b", "main")
	runGit(t, primary, "-c", "user.email=t@example.com", "-c", "user.name=t",
		"commit", "-q", "--allow-empty", "-m", "seed")

	worktree := filepath.Join(canonical(t, t.TempDir()), "wt")
	runGit(t, primary, "worktree", "add", "-q", "-b", "wt-branch", worktree)

	want := filepath.Join(primary, DefaultRegistryPath)

	// Positive control: from the primary checkout the answer is the primary's
	// own registry. Without this, a resolver that always returned `want` by
	// accident would pass the load-bearing case below.
	if got := RegistryPathFor(primary); got != want {
		t.Fatalf("RegistryPathFor(primary) = %q, want %q", got, want)
	}

	// The load-bearing case: a directory INSIDE a linked worktree resolves to
	// the primary checkout's registry, not to one under the worktree.
	got := RegistryPathFor(worktree)
	if got != want {
		t.Errorf("RegistryPathFor(worktree) = %q, want %q (worktree must not get its own registry)", got, want)
	}
	if filepath.Dir(filepath.Dir(filepath.Dir(got))) == worktree {
		t.Errorf("RegistryPathFor(worktree) resolved inside the worktree: %q", got)
	}
}

// TestRegistryPathForFallsBackOutsideRepository pins the fallback: a directory
// that is not inside a git repository keeps the pre-anchoring behavior exactly,
// and an empty directory yields the bare relative constant.
func TestRegistryPathForFallsBackOutsideRepository(t *testing.T) {
	t.Parallel()

	dir := canonical(t, t.TempDir())
	if got, want := RegistryPathFor(dir), filepath.Join(dir, DefaultRegistryPath); got != want {
		t.Errorf("RegistryPathFor(non-repo) = %q, want %q", got, want)
	}
	if got := RegistryPathFor(""); got != DefaultRegistryPath {
		t.Errorf("RegistryPathFor(\"\") = %q, want %q", got, DefaultRegistryPath)
	}
}

// canonical resolves symlinks so comparisons hold on macOS, where t.TempDir()
// hands back /var/... while git reports /private/var/....
func canonical(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q): %v", path, err)
	}
	return resolved
}

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}
