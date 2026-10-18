package cli

// integration_common_root_test.go — card t1680, sync-audit round 1 finding F1.
//
// integrationLockRoot() answered CLAUDE_PROJECT_DIR (spelled by git) whenever the
// variable was set, so a variable naming a LINKED worktree made the verb's
// integration-window and re-measure root that worktree, while the MCP handler
// derived the parent checkout from project_root. The two surfaces then used
// different windows. factory.IntegrationCommonRoot is the one rule both surfaces
// now share (and the PreToolUse guard, card t1680 iter5): the repository common
// root (the primary checkout), never the caller's own worktree. Its unit cases
// live beside it in internal/factory; this file keeps the CLI-level pins.

import (
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
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
		window := factory.IntegrationCommonRoot(dir)
		if record := factoryRootOf(dir); window != record {
			t.Errorf("%s: window root %q != record root %q", name, window, record)
		}
		t.Setenv("CLAUDE_PROJECT_DIR", dir)
		if verb := integrationLockRoot(); verb != window {
			t.Errorf("%s: verb window root %q != handler window root %q", name, verb, window)
		}
	}
}
