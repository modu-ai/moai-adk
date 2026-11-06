// launch_view_abs_worktree_t2013_test.go — card t2013: an absolute `-w` value is
// entered (the launcher starts inside the tree and hands the backend no worktree
// selection) only when it equals a REGISTERED worktree of this repository; an
// absolute path that is not registered is refused with a message naming the path,
// before the launch writes or execs anything. The root restriction of
// validateWorktreeL2Value stays in force and runs first; short names behave as
// before.
//
// Every fixture is a scratch repository under t.TempDir() with its own git
// config and MOAI_HOME. The linked trees made with `git worktree add` are
// scratch fixtures, the sanctioned exception to AGENTS.md section 3.
package cli

import (
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/paths"
)

// absWtFixture is a scratch repository with two registered linked worktrees: one
// outside the repository (registered) and one under <repo>/.moai/worktrees
// (legacy), and a plain directory that is not a worktree (unregistered).
type absWtFixture struct {
	repo         string
	registered   string
	legacy       string
	unregistered string
}

func absWtSetup(t *testing.T) absWtFixture {
	t.Helper()
	cfg := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(cfg, []byte("[user]\n\tname = Fixture\n\temail = fixture@example.invalid\n[init]\n\tdefaultBranch = main\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", cfg)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv(paths.EnvHome, t.TempDir())

	repo := t.TempDir()
	fcGit(t, repo, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(repo, "a.txt"), []byte("a\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fcGit(t, repo, "add", "a.txt")
	fcGit(t, repo, "commit", "-q", "-m", "base")

	registered := filepath.Join(t.TempDir(), "wt-registered")
	fcGit(t, repo, "worktree", "add", "-q", "-b", "wt-registered", registered)
	legacy := filepath.Join(repo, ".moai", "worktrees", "legacy-one")
	if err := os.MkdirAll(filepath.Dir(legacy), 0o755); err != nil {
		t.Fatal(err)
	}
	fcGit(t, repo, "worktree", "add", "-q", "-b", "wt-legacy", legacy)
	unregistered := filepath.Join(t.TempDir(), "not-a-worktree")
	if err := os.MkdirAll(unregistered, 0o755); err != nil {
		t.Fatal(err)
	}

	prev := findProjectRootFn
	findProjectRootFn = func() (string, error) { return repo, nil }
	t.Cleanup(func() { findProjectRootFn = prev })
	return absWtFixture{repo: repo, registered: registered, legacy: legacy, unregistered: unregistered}
}

// absWtListing is the sorted entry names of dir, the "nothing was written" probe.
func absWtListing(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	slices.Sort(names)
	return names
}

// (e) An absolute path equal to a registered worktree is entered: the entry
// directory is that tree, the selection is stripped from the argv, and nothing
// is written into the tree.
func TestT2013AbsoluteRegisteredWorktreeIsEntered(t *testing.T) {
	fx := absWtSetup(t)
	listing := absWtListing(t, fx.registered)
	status := fcGit(t, fx.registered, "status", "--porcelain=v1")

	for _, args := range [][]string{
		{"--worktree", fx.registered},
		{"-b", "--worktree", fx.registered, "--model", "opus"},
	} {
		dir := existingWorktreeEntryDir(args)
		if dir == "" || resolveSymlinks(dir) != resolveSymlinks(fx.registered) {
			t.Errorf("existingWorktreeEntryDir(%v) = %q, want the registered tree %q", args, dir, fx.registered)
		}
		if err := validateAbsoluteWorktreeRegistered(readFirstWorktree(args)); err != nil {
			t.Errorf("validateAbsoluteWorktreeRegistered(%v) = %v, want nil for a registered tree", args, err)
		}
	}

	stripped := stripWorktreeSelection([]string{"-b", "--worktree", fx.registered, "--model", "opus"})
	if want := []string{"-b", "--model", "opus"}; !slices.Equal(stripped, want) {
		t.Errorf("stripWorktreeSelection = %v, want %v (the backend gets no worktree selection)", stripped, want)
	}
	if got := absWtListing(t, fx.registered); !slices.Equal(got, listing) {
		t.Errorf("entry wrote into the tree: %v → %v", listing, got)
	}
	if got := fcGit(t, fx.registered, "status", "--porcelain=v1"); got != status {
		t.Errorf("entry changed the tree's status: %q → %q", status, got)
	}
}

// (f) An absolute path that is not registered is refused, the message names the
// path, and nothing is created or written.
func TestT2013AbsoluteUnregisteredWorktreeIsRefused(t *testing.T) {
	fx := absWtSetup(t)
	missing := filepath.Join(t.TempDir(), "never-created")
	inside := filepath.Join(fx.registered, "a-subdirectory")
	if err := os.MkdirAll(inside, 0o755); err != nil {
		t.Fatal(err)
	}
	listing := absWtListing(t, fx.unregistered)
	repoListing := absWtListing(t, fx.repo)

	for _, path := range []string{fx.unregistered, missing, inside, fx.repo + "-sibling"} {
		args := []string{"--worktree", path}
		err := validateAbsoluteWorktreeRegistered(readFirstWorktree(args))
		if err == nil {
			t.Errorf("validateAbsoluteWorktreeRegistered(%q) = nil, want a refusal", path)
			continue
		}
		if !strings.Contains(err.Error(), path) || !strings.Contains(err.Error(), "not a registered worktree") {
			t.Errorf("refusal for %q = %q, want it to name the path and that it is not a registered worktree", path, err)
		}
		if dir := existingWorktreeEntryDir(args); dir != "" {
			t.Errorf("existingWorktreeEntryDir(%q) = %q, want no entry directory", path, dir)
		}
	}
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Errorf("the refused path %q was created (stat err: %v)", missing, err)
	}
	if got := absWtListing(t, fx.unregistered); !slices.Equal(got, listing) {
		t.Errorf("the refused directory was written to: %v → %v", listing, got)
	}
	if got := absWtListing(t, fx.repo); !slices.Equal(got, repoListing) {
		t.Errorf("the repository was written to: %v → %v", repoListing, got)
	}
}

// (g) Short names and a bare -w behave exactly as before: the registered check is
// a no-op for them and the entry directory stays empty for a name that names no
// existing tree.
func TestT2013ShortNameBehaviorUnchanged(t *testing.T) {
	absWtSetup(t)
	for _, args := range [][]string{
		{"--worktree", "feat-login"},
		{"-w", "feat-login"},
		{"-w"},
		{"--model", "opus"},
		{"--worktree", "."},
		{},
	} {
		if err := validateAbsoluteWorktreeRegistered(readFirstWorktree(args)); err != nil {
			t.Errorf("validateAbsoluteWorktreeRegistered(%v) = %v, want nil for a non-absolute -w", args, err)
		}
		if dir := existingWorktreeEntryDir(args); dir != "" {
			t.Errorf("existingWorktreeEntryDir(%v) = %q, want \"\" (short name without an existing home-root tree)", args, dir)
		}
	}
}

// (h) A registered tree under <repo>/.moai/worktrees/<name> (a legacy-root
// tree) is accepted when named by its absolute path; the root restriction of
// validateWorktreeL2Value still runs first and still refuses a registered tree
// that sits outside every accepted prefix.
func TestT2013AbsoluteRegisteredLegacyRootTreeAccepted(t *testing.T) {
	fx := absWtSetup(t)
	args := []string{"--worktree", fx.legacy}

	if err := validateWorktreeL2Value(readFirstWorktree(args), io.Discard); err != nil {
		t.Fatalf("validateWorktreeL2Value(legacy tree) = %v, want the accepted prefix to pass", err)
	}
	if err := validateAbsoluteWorktreeRegistered(readFirstWorktree(args)); err != nil {
		t.Errorf("validateAbsoluteWorktreeRegistered(legacy tree) = %v, want nil", err)
	}
	if dir := existingWorktreeEntryDir(args); dir == "" || resolveSymlinks(dir) != resolveSymlinks(fx.legacy) {
		t.Errorf("existingWorktreeEntryDir(legacy tree) = %q, want %q", dir, fx.legacy)
	}

	outside := []string{"--worktree", fx.registered}
	if err := validateWorktreeL2Value(readFirstWorktree(outside), io.Discard); err == nil {
		t.Errorf("validateWorktreeL2Value(registered tree outside every accepted prefix) = nil, want the root restriction to refuse it first")
	}
}
