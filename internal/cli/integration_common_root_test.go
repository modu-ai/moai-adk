package cli

// integration_common_root_test.go — card t1680, sync-audit round 1 finding F1.
//
// integrationLockRoot() answered CLAUDE_PROJECT_DIR (spelled by git) whenever the
// variable was set, so a variable naming a LINKED worktree made the verb's
// integration-window and re-measure root that worktree, while the MCP handler
// derived the parent checkout from project_root. The two surfaces then used
// different windows. integrationCommonRootOf is the one rule both surfaces now
// share: the repository common root (the primary checkout), never the caller's
// own worktree.

import (
	"os"
	"path/filepath"
	"testing"
)

// commonRootRepo builds a repository with one commit and one linked worktree, and
// returns (primary, linked). The base is symlink-resolved so every expected value
// is git's own spelling.
func commonRootRepo(t *testing.T) (primary, linked string) {
	t.Helper()
	primary = caseFixtureRepo(t, "CommonRootRepo")
	fcGit(t, primary, "commit", "-q", "--allow-empty", "-m", "base")
	linked = filepath.Join(filepath.Dir(primary), "Common-Root-Linked")
	fcGit(t, primary, "worktree", "add", "-q", "-b", "WT-common-root", linked)
	return primary, linked
}

func TestIntegrationCommonRootOf_LinkedWorktreeResolvesToThePrimaryCheckout(t *testing.T) {
	primary, linked := commonRootRepo(t)
	if got := integrationCommonRootOf(linked); got != primary {
		t.Errorf("integrationCommonRootOf(linked worktree) = %q, want the primary checkout %q", got, primary)
	}
}

func TestIntegrationCommonRootOf_PrimaryCheckoutIsItself(t *testing.T) {
	primary, _ := commonRootRepo(t)
	if got := integrationCommonRootOf(primary); got != primary {
		t.Errorf("integrationCommonRootOf(primary) = %q, want %q", got, primary)
	}
}

func TestIntegrationCommonRootOf_SubdirectoryOfALinkedWorktreeResolvesToThePrimaryCheckout(t *testing.T) {
	primary, linked := commonRootRepo(t)
	sub := filepath.Join(linked, "pkg", "deep")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := integrationCommonRootOf(sub); got != primary {
		t.Errorf("integrationCommonRootOf(subdir of linked worktree) = %q, want the primary checkout %q", got, primary)
	}
}

func TestIntegrationCommonRootOf_NonRepoAndMissingPathsAreReturnedVerbatim(t *testing.T) {
	plain, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(plain, "T1680MISSING")
	for name, dir := range map[string]string{"non-repo": plain, "missing": missing} {
		if got := integrationCommonRootOf(dir); got != dir {
			t.Errorf("%s: integrationCommonRootOf(%q) = %q, want it returned verbatim", name, dir, got)
		}
	}
}

// TestIntegrationCommonRootOf_ReturnsGitsOwnSpellingNotTheCallers is the t766 property
// carried through the new rule: a wrong-case spelling of the primary checkout, and of a
// linked worktree, both answer the primary checkout in git's spelling. On a
// case-sensitive volume the wrong spelling names nothing and is returned verbatim.
func TestIntegrationCommonRootOf_ReturnsGitsOwnSpellingNotTheCallers(t *testing.T) {
	primary, linked := commonRootRepo(t)
	for name, c := range map[string]struct{ in, want string }{
		"primary":  {lowerLeaf(primary), primary},
		"worktree": {lowerLeaf(linked), primary},
	} {
		if c.in == "" || c.in == c.want {
			t.Fatalf("%s: fixture spelling %q is not distinct from %q; the test would be vacuous", name, c.in, c.want)
		}
		want := c.want
		if _, err := os.Stat(c.in); err != nil { // case-sensitive volume: the wrong spelling names nothing
			want = c.in
		}
		if got := integrationCommonRootOf(c.in); got != want {
			t.Errorf("%s: integrationCommonRootOf(%q) = %q, want %q", name, c.in, got, want)
		}
	}
}

// TestIntegrationLockRoot_LinkedWorktreeVariableResolvesToThePrimaryCheckout pins the verb
// side: CLAUDE_PROJECT_DIR naming a linked worktree no longer makes that worktree the
// window root.
func TestIntegrationLockRoot_LinkedWorktreeVariableResolvesToThePrimaryCheckout(t *testing.T) {
	primary, linked := commonRootRepo(t)
	t.Setenv("CLAUDE_PROJECT_DIR", linked)
	if got := integrationLockRoot(); got != primary {
		t.Errorf("integrationLockRoot() with CLAUDE_PROJECT_DIR=linked worktree = %q, want the primary checkout %q", got, primary)
	}
}

// TestIntegrationWindowRootAgreesWithTheFactoryRecordRoot: for the inputs an MCP call
// carries as project_root, the window root the handler derives equals the record root it
// derives (factoryRootOf), and the verb's window root with the variable set to the same
// input is that same directory.
func TestIntegrationWindowRootAgreesWithTheFactoryRecordRoot(t *testing.T) {
	primary, linked := commonRootRepo(t)
	plain, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for name, dir := range map[string]string{"primary": primary, "linked worktree": linked, "non-repo": plain} {
		window := integrationCommonRootOf(dir)
		if record := factoryRootOf(dir); window != record {
			t.Errorf("%s: window root %q != record root %q", name, window, record)
		}
		t.Setenv("CLAUDE_PROJECT_DIR", dir)
		if verb := integrationLockRoot(); verb != window {
			t.Errorf("%s: verb window root %q != handler window root %q", name, verb, window)
		}
	}
}
