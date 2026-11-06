package hook

// integration_lock_guard_common_root_test.go — card t1680, sync-audit iter5 finding F1.
//
// The CLI verbs and the MCP factory_complete handler resolve the integration window to the
// repository common root, so a CLAUDE_PROJECT_DIR naming a LINKED worktree stores the lock record
// at the primary checkout. The PreToolUse guard read the record at the variable as written, so a
// live foreign holder was invisible to it and `git merge` was allowed. These tests go from the
// record the CLI writes (factory.AcquireIntegrationLock at the primary checkout) to the deny the
// real PreToolUse handler returns, through the real project-root resolution
// (CLAUDE_PROJECT_DIR), and pin the controls that must not move.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// commonRootGuardRepo builds a repository with one commit and one linked worktree and returns
// (primary, linked). The base is symlink-resolved so every path is git's own spelling.
func commonRootGuardRepo(t *testing.T) (primary, linked string) {
	t.Helper()
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("fixture: resolve temp dir: %v", err)
	}
	primary = filepath.Join(base, "primary")
	if err := os.MkdirAll(primary, 0o755); err != nil {
		t.Fatalf("fixture: mkdir: %v", err)
	}
	linked = filepath.Join(base, "linked")
	for _, step := range []struct {
		dir  string
		args []string
	}{
		{primary, []string{"init", "-q", "-b", "main"}},
		{primary, []string{"config", "user.name", "t1680 fixture"}},
		{primary, []string{"config", "user.email", "t1680@example.invalid"}},
		{primary, []string{"commit", "-q", "--allow-empty", "--no-gpg-sign", "-m", "base"}},
		{primary, []string{"worktree", "add", "-q", "-b", "WT-guard-common-root", linked}},
	} {
		cmd := exec.Command("git", step.args...)
		cmd.Dir = step.dir
		if out, runErr := cmd.CombinedOutput(); runErr != nil {
			t.Fatalf("fixture: git %s: %v\n%s", strings.Join(step.args, " "), runErr, out)
		}
	}
	return primary, linked
}

// acquireAsTheCLIDoes records a live holder at root through the same accessor the CLI acquire
// verb uses, so the seed is the CLI's write and not a hand-built file.
func acquireAsTheCLIDoes(t *testing.T, root, sessionID string) {
	t.Helper()
	if _, err := factory.AcquireIntegrationLock(root, factory.IntegrationLock{
		SessionID:   sessionID,
		SessionName: "lane-5",
		PID:         os.Getpid(),
		Branch:      "main",
		AcquiredAt:  "2026-01-01T00:00:00Z",
	}, false); err != nil {
		t.Fatalf("seed: acquire at %s: %v", root, err)
	}
}

// guardHandlerFromEnv builds the real PreToolUse handler whose project root is resolved from
// CLAUDE_PROJECT_DIR (the production path), with the integration-lock guard on or off.
func guardHandlerFromEnv(t *testing.T, projectDirEnv string, guardEnabled bool) *preToolHandler {
	t.Helper()
	hmpIsolateHome(t)
	t.Setenv("CLAUDE_PROJECT_DIR", projectDirEnv)
	return &preToolHandler{
		cfg:                 &mockConfigProvider{cfg: hmpCfg(false, guardEnabled, config.SlotLeaseConfig{})},
		policy:              DefaultSecurityPolicy(),
		projectRootResolver: projectRootResolver{caller: "t1680"},
	}
}

func mergeAs(t *testing.T, h *preToolHandler, sessionID, cwd string) (decision, reason string) {
	t.Helper()
	return hmpHandle(t, h, hmpInput(t, "Bash", sessionID, cwd, "git merge --no-ff WT-thing"))
}

// The finding: the record sits at the primary checkout (what the CLI writes), the variable names a
// linked worktree, and a different session's merge must be denied.
func TestIntegrationLockGuard_LinkedWorktreeVariableSeesTheRecordTheCLIWrote(t *testing.T) {
	primary, linked := commonRootGuardRepo(t)
	acquireAsTheCLIDoes(t, primary, "sess-holder")
	h := guardHandlerFromEnv(t, linked, true)

	decision, reason := mergeAs(t, h, "sess-other", linked)
	if decision != DecisionDeny {
		t.Fatalf("decision = %q (reason %q), want deny: a live foreign holder recorded at the primary checkout is invisible to the guard when CLAUDE_PROJECT_DIR names a linked worktree", decision, reason)
	}
	if !strings.HasPrefix(reason, integrationLockViolationPrefix) || !strings.Contains(reason, "lane-5") {
		t.Errorf("deny reason lacks the sentinel or the holder: %q", reason)
	}
}

// Controls: each moves nothing and must pass both before and after the repair.
func TestIntegrationLockGuard_CommonRootControls(t *testing.T) {
	t.Run("variable names the primary checkout still denies", func(t *testing.T) {
		primary, _ := commonRootGuardRepo(t)
		acquireAsTheCLIDoes(t, primary, "sess-holder")
		h := guardHandlerFromEnv(t, primary, true)
		if d, r := mergeAs(t, h, "sess-other", primary); d != DecisionDeny {
			t.Errorf("decision = %q (%q), want deny", d, r)
		}
	})

	t.Run("holder is the caller's own session allows", func(t *testing.T) {
		primary, linked := commonRootGuardRepo(t)
		acquireAsTheCLIDoes(t, primary, "sess-holder")
		h := guardHandlerFromEnv(t, linked, true)
		if d, r := mergeAs(t, h, "sess-holder", linked); d == DecisionDeny {
			t.Errorf("the holder was denied its own window: %q", r)
		}
	})

	t.Run("no lock held allows", func(t *testing.T) {
		_, linked := commonRootGuardRepo(t)
		h := guardHandlerFromEnv(t, linked, true)
		if d, r := mergeAs(t, h, "sess-other", linked); d == DecisionDeny {
			t.Errorf("denied with no lock held: %q", r)
		}
	})

	t.Run("guard disabled reads nothing and allows", func(t *testing.T) {
		primary, linked := commonRootGuardRepo(t)
		acquireAsTheCLIDoes(t, primary, "sess-holder")
		acquireAsTheCLIDoes(t, linked, "sess-holder")
		h := guardHandlerFromEnv(t, linked, false)
		if d, r := mergeAs(t, h, "sess-other", linked); d == DecisionDeny {
			t.Errorf("a disabled guard denied: %q", r)
		}
	})

	t.Run("non-repo root keeps the verbatim path", func(t *testing.T) {
		plain, err := filepath.EvalSymlinks(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		acquireAsTheCLIDoes(t, plain, "sess-holder")
		h := guardHandlerFromEnv(t, plain, true)
		if d, r := mergeAs(t, h, "sess-other", plain); d != DecisionDeny {
			t.Errorf("non-repo root with a record at that root: decision = %q (%q), want deny (verbatim path unchanged)", d, r)
		}
	})

	t.Run("missing root fails open", func(t *testing.T) {
		plain, err := filepath.EvalSymlinks(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		missing := filepath.Join(plain, "T1680MISSING")
		h := guardHandlerFromEnv(t, missing, true)
		if d, r := mergeAs(t, h, "sess-other", missing); d == DecisionDeny {
			t.Errorf("a missing root denied: %q", r)
		}
	})

	t.Run("a command that is not a merge is unchanged", func(t *testing.T) {
		primary, linked := commonRootGuardRepo(t)
		acquireAsTheCLIDoes(t, primary, "sess-holder")
		h := guardHandlerFromEnv(t, linked, true)
		if d, r := hmpHandle(t, h, hmpInput(t, "Bash", "sess-other", linked, "git status")); d == DecisionDeny {
			t.Errorf("git status denied: %q", r)
		}
	})
}

// An indirect PowerShell construct is never denied; with the record visible it is audited, and the
// audit line lands where the hook has always written it — the project root the hook resolved.
func TestIntegrationLockGuard_UnclassifiedAuditStaysAtTheHookProjectRoot(t *testing.T) {
	primary, linked := commonRootGuardRepo(t)
	acquireAsTheCLIDoes(t, primary, "sess-holder")
	h := guardHandlerFromEnv(t, linked, true)

	if d, r := hmpHandle(t, h, hmpInput(t, "PowerShell", "sess-other", linked, `iex "git merge --no-ff WT-x"`)); d == DecisionDeny {
		t.Fatalf("unclassifiable construct denied: %q", r)
	}
	if lines := hmpAuditLines(t, filepath.Join(linked, integrationLockAuditRelPath), "powershell-unclassified"); len(lines) != 1 {
		t.Errorf("want exactly 1 powershell-unclassified line at the hook project root, got %d: %q", len(lines), lines)
	}
	if lines := hmpAuditLines(t, filepath.Join(primary, integrationLockAuditRelPath), "powershell-unclassified"); len(lines) != 0 {
		t.Errorf("the audit line must not follow the lock to the common root, got %d at the primary checkout", len(lines))
	}
}
