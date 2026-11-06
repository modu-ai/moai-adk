package cli

// integration_session_sidecar_test.go — card t1680, sync-audit iter3 finding R3-1.
//
// integrationSessionID's degraded path reads the SessionStart side-channel file. The hook
// WRITES that file at the project directory exactly as Claude Code names it
// (input.ProjectDir), so the read has to happen at the same place. Once
// integrationLockRoot() began answering the repository common root for a variable that
// names a linked worktree, the read followed it to the primary checkout and adopted
// whichever session last started THERE: a caller with CLAUDE_CODE_SESSION_ID unset could
// release another session's window without --force.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/session"
)

const (
	sidecarHolderID = "HOLDER-SESSION"
	sidecarCallerID = "CALLER-SESSION"
)

// writeSidecar writes the side-channel file the way the SessionStart hook does: at
// <dir>/session.CurrentSideChannelFile.
func writeSidecar(t *testing.T, dir, id string) {
	t.Helper()
	path := filepath.Join(dir, session.CurrentSideChannelFile)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(id+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}

// unstampSessionEnv makes the per-process env var empty so the side-channel file is the
// only remaining source of the caller's id.
func unstampSessionEnv(t *testing.T) {
	t.Helper()
	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
}

// sidecarWorktreeFixture returns a primary checkout and a linked worktree whose
// side-channel slots hold different ids: the primary's names the window holder, the linked
// worktree's names the caller.
func sidecarWorktreeFixture(t *testing.T) (primary, linked string) {
	t.Helper()
	primary, linked = commonRootRepo(t)
	writeSidecar(t, primary, sidecarHolderID)
	writeSidecar(t, linked, sidecarCallerID)
	return primary, linked
}

// The slot at the variable's own location is the caller's; the primary's slot is another
// session's and must not be adopted.
func TestIntegrationSessionID_ReadsTheSlotWhereTheHookWritesIt(t *testing.T) {
	primary, linked := sidecarWorktreeFixture(t)
	unstampSessionEnv(t)

	sub := filepath.Join(linked, "pkg", "deep")
	writeSidecar(t, sub, "SUBDIR-SESSION")

	for name, c := range map[string]struct{ dir, want string }{
		"variable names the linked worktree":    {linked, sidecarCallerID},
		"variable names a subdir of the linked": {sub, "SUBDIR-SESSION"},
		"variable names the primary checkout":   {primary, sidecarHolderID},
	} {
		t.Setenv("CLAUDE_PROJECT_DIR", c.dir)
		if got := integrationSessionID(""); got != c.want {
			t.Errorf("%s: integrationSessionID(\"\") = %q, want %q (the slot at %s)", name, got, c.want, c.dir)
		}
	}
}

// A caller whose own slot names it must NOT be able to release the window another
// session holds at the primary checkout, without --force.
func TestIntegrationRelease_LinkedWorktreeCallerCannotReleaseAnotherSessionsWindow(t *testing.T) {
	primary, linked := sidecarWorktreeFixture(t)
	unstampSessionEnv(t)

	if _, err := factory.AcquireIntegrationLock(primary, factory.IntegrationLock{
		SessionID: sidecarHolderID,
		Branch:    "main",
		Worktree:  primary,
	}, false); err != nil {
		t.Fatalf("seed the holder's window: %v", err)
	}

	out, err := runIntegration(t, linked, "release")
	if err == nil {
		t.Fatalf("release by %s succeeded on a window held by %s without --force: %s", sidecarCallerID, sidecarHolderID, out)
	}
	if !strings.Contains(err.Error(), sidecarHolderID) {
		t.Errorf("the refusal does not name the holder %s: %v", sidecarHolderID, err)
	}
	held, readErr := factory.ReadIntegrationLock(primary)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !held.Held() || held.SessionID != sidecarHolderID {
		t.Errorf("the window was disturbed by the refused release: %+v", held)
	}
}

// The precedence the fix must leave alone: --session, then the env var, then the file; and
// with the variable unset the file is read at the git-resolved root exactly as before.
func TestIntegrationSessionID_ControlCasesKeepTheirPrecedence(t *testing.T) {
	_, linked := sidecarWorktreeFixture(t)
	t.Setenv("CLAUDE_PROJECT_DIR", linked)

	t.Run("explicit --session wins over the file", func(t *testing.T) {
		unstampSessionEnv(t)
		if got := integrationSessionID("EXPLICIT-SESSION"); got != "EXPLICIT-SESSION" {
			t.Errorf("got %q, want EXPLICIT-SESSION", got)
		}
	})
	t.Run("CLAUDE_CODE_SESSION_ID wins over the file", func(t *testing.T) {
		t.Setenv("CLAUDE_CODE_SESSION_ID", "ENV-SESSION")
		if got := integrationSessionID(""); got != "ENV-SESSION" {
			t.Errorf("got %q, want ENV-SESSION", got)
		}
	})
	t.Run("variable unset keeps the git-resolved fallback", func(t *testing.T) {
		unstampSessionEnv(t)
		t.Setenv("CLAUDE_PROJECT_DIR", "")
		t.Chdir(linked)
		// The unchanged fallback is integrationLockRoot(): from inside a linked worktree
		// it answers the primary checkout, whose slot names the holder.
		if got := integrationSessionID(""); got != sidecarHolderID {
			t.Errorf("got %q, want %q (the slot at the git-resolved common root)", got, sidecarHolderID)
		}
	})
}

// For a project that is not a linked worktree the read location is the write location:
// CLAUDE_PROJECT_DIR as given, joined with the shared side-channel path.
func TestIntegrationSessionID_PlainProjectReadsWhereTheHookWrites(t *testing.T) {
	unstampSessionEnv(t)
	plainRepo, _ := commonRootRepo(t)
	nonRepo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for name, dir := range map[string]string{"plain repository": plainRepo, "non-repository directory": nonRepo} {
		writeSidecar(t, dir, "PLAIN-"+name)
		t.Setenv("CLAUDE_PROJECT_DIR", dir)
		raw, readErr := os.ReadFile(filepath.Join(dir, session.CurrentSideChannelFile))
		if readErr != nil {
			t.Fatal(readErr)
		}
		if got, want := integrationSessionID(""), strings.TrimSpace(string(raw)); got != want {
			t.Errorf("%s: integrationSessionID(\"\") = %q, want the hook's slot %q", name, got, want)
		}
	}
}
