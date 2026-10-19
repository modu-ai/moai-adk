package factory

// integration_common_root_test.go — card t1680. IntegrationCommonRoot is the one rule that names
// the integration-window root for the CLI verbs, the MCP handler and the PreToolUse guard: the
// repository common root (the primary checkout), never the caller's own worktree. The cases were
// ported from internal/cli (TestIntegrationCommonRootOf_*) when the rule moved here.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// commonRootGit runs one git command in dir and fails the test on error.
func commonRootGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("fixture: git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

// commonRootRepo builds a repository with one commit and one linked worktree, and returns
// (primary, linked). The base is symlink-resolved so every expected value is git's own spelling.
func commonRootRepo(t *testing.T) (primary, linked string) {
	t.Helper()
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("fixture: resolve temp dir: %v", err)
	}
	primary = filepath.Join(base, "CommonRootRepo")
	if err := os.MkdirAll(primary, 0o755); err != nil {
		t.Fatalf("fixture: mkdir: %v", err)
	}
	linked = filepath.Join(base, "Common-Root-Linked")
	commonRootGit(t, primary, "init", "-q", "-b", "main")
	commonRootGit(t, primary, "config", "user.name", "t1680 fixture")
	commonRootGit(t, primary, "config", "user.email", "t1680@example.invalid")
	commonRootGit(t, primary, "commit", "-q", "--allow-empty", "--no-gpg-sign", "-m", "base")
	commonRootGit(t, primary, "worktree", "add", "-q", "-b", "WT-common-root", linked)
	return primary, linked
}

func commonRootLowerLeaf(path string) string {
	return filepath.Join(filepath.Dir(path), strings.ToLower(filepath.Base(path)))
}

func TestIntegrationCommonRoot_LinkedWorktreeResolvesToThePrimaryCheckout(t *testing.T) {
	primary, linked := commonRootRepo(t)
	if got := IntegrationCommonRoot(linked); got != primary {
		t.Errorf("IntegrationCommonRoot(linked worktree) = %q, want the primary checkout %q", got, primary)
	}
}

func TestIntegrationCommonRoot_PrimaryCheckoutIsItself(t *testing.T) {
	primary, _ := commonRootRepo(t)
	if got := IntegrationCommonRoot(primary); got != primary {
		t.Errorf("IntegrationCommonRoot(primary) = %q, want %q", got, primary)
	}
}

func TestIntegrationCommonRoot_SubdirectoryOfALinkedWorktreeResolvesToThePrimaryCheckout(t *testing.T) {
	primary, linked := commonRootRepo(t)
	sub := filepath.Join(linked, "pkg", "deep")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := IntegrationCommonRoot(sub); got != primary {
		t.Errorf("IntegrationCommonRoot(subdir of linked worktree) = %q, want the primary checkout %q", got, primary)
	}
}

func TestIntegrationCommonRoot_NonRepoAndMissingPathsAreReturnedVerbatim(t *testing.T) {
	plain, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(plain, "T1680MISSING")
	for name, dir := range map[string]string{"non-repo": plain, "missing": missing} {
		if got := IntegrationCommonRoot(dir); got != dir {
			t.Errorf("%s: IntegrationCommonRoot(%q) = %q, want it returned verbatim", name, dir, got)
		}
	}
}

// TestIntegrationCommonRoot_ReturnsGitsOwnSpellingNotTheCallers is the t766 property carried
// through the rule: a wrong-case spelling of the primary checkout, and of a linked worktree, both
// answer the primary checkout in git's spelling. On a case-sensitive volume the wrong spelling
// names nothing and is returned verbatim.
func TestIntegrationCommonRoot_ReturnsGitsOwnSpellingNotTheCallers(t *testing.T) {
	primary, linked := commonRootRepo(t)
	for name, c := range map[string]struct{ in, want string }{
		"primary":  {commonRootLowerLeaf(primary), primary},
		"worktree": {commonRootLowerLeaf(linked), primary},
	} {
		if c.in == "" || c.in == c.want {
			t.Fatalf("%s: fixture spelling %q is not distinct from %q; the test would be vacuous", name, c.in, c.want)
		}
		want := c.want
		if _, err := os.Stat(c.in); err != nil { // case-sensitive volume: the wrong spelling names nothing
			want = c.in
		}
		if got := IntegrationCommonRoot(c.in); got != want {
			t.Errorf("%s: IntegrationCommonRoot(%q) = %q, want %q", name, c.in, got, want)
		}
	}
}

// t1589 — the spelling check reads the repository named by an inherited GIT_DIR instead of the
// directory it was given. Given a symlink to a repository, git answers the real path; under the
// inherited variable it answers the symlink back, and the spelling correction is silently skipped.
// Ported from internal/cli (TestCanonicalTreeSpellingIgnoresInheritedGitEnv) with the function.
func TestCanonicalTreeSpellingIgnoresInheritedGitEnv(t *testing.T) {
	repo, _ := commonRootRepo(t)
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(repo, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	other, _ := commonRootRepo(t)
	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	t.Setenv("GIT_WORK_TREE", other)

	got := canonicalTreeSpelling(link)
	if got == link {
		t.Errorf("canonicalTreeSpelling(link) = %q, want the repository's own spelling %q (the inherited GIT_DIR left the link unresolved)", got, repo)
	}
	if gotInfo, gotErr := os.Stat(got); gotErr != nil {
		t.Errorf("canonicalTreeSpelling(link) = %q does not name a directory: %v", got, gotErr)
	} else if wantInfo, wantErr := os.Stat(repo); wantErr != nil || !os.SameFile(gotInfo, wantInfo) {
		t.Errorf("canonicalTreeSpelling(link) = %q does not name the repository %q", got, repo)
	}
}
