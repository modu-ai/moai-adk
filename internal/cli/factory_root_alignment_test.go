package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// Card t1862 (t1827 sync-audit F1 follow-up): `moai factory runs` and
// `moai factory handoff recover-resume` open the factory at the raw working
// directory, while every other factory surface opens it at the moai-paths
// root — factoryCardRoot(): CLAUDE_PROJECT_DIR first, then the launch
// directory, mapped to the queue root (SPEC-FACTORY-LEASE-RENEW-001, which
// binds every surface that opens the factory). A cwd naming a different
// project than moai paths therefore reads another repository's factory —
// and, because OpenFactory reads the lease keys from the root it is opened
// at, misjudges lease state on the same wrong root. The boundary pinned
// here: cwd names project B while CLAUDE_PROJECT_DIR names project A; both
// verbs must read A's factory, never B's.

// rootAlignmentSandbox scrubs home-level state once and returns two distinct
// non-git project roots: pathsRoot, the root moai paths resolve to, and cwd,
// the directory the verbs are launched from.
func rootAlignmentSandbox(t *testing.T) (pathsRoot, cwd string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("MOAI_HOME", filepath.Join(home, ".moai"))
	t.Setenv("MOAI_CLAUDE_BIN", filepath.Join(home, "bin", "claude"))
	pathsRoot = filepath.Join(t.TempDir(), "project-a")
	cwd = filepath.Join(t.TempDir(), "project-b")
	for _, root := range []string{pathsRoot, cwd} {
		if err := os.MkdirAll(root, 0o700); err != nil {
			t.Fatalf("create sandbox root %s: %v", root, err)
		}
	}
	return pathsRoot, cwd
}

func TestFactoryRunsOpensMoaiPathsRootNotCwd(t *testing.T) {
	pathsRoot, cwd := rootAlignmentSandbox(t)
	seedRun(t, pathsRoot, "run-at-paths-root", 0, "")
	t.Setenv("CLAUDE_PROJECT_DIR", pathsRoot)
	t.Chdir(cwd)
	out, err := runFactoryCommand(t, "runs")
	if err != nil {
		t.Fatalf("factory runs: %v (output %q)", err, out)
	}
	if !strings.Contains(out, "run-at-paths-root") {
		t.Fatalf("factory runs read the cwd factory, not the moai-paths root (card t1862): output %q", out)
	}
}

func TestRecoverResumeOpensMoaiPathsRootNotCwd(t *testing.T) {
	pathsRoot, cwd := rootAlignmentSandbox(t)
	f, err := homestate.OpenFactory(pathsRoot)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := f.SaveResume(ctx, homestate.ResumeHandoff{SchemaVersion: 1, SavedAt: time.Now(), Body: "legacy", DirectivesJSON: "{}"}); err != nil {
		t.Fatal(err)
	}
	row, _, err := f.ClaimResume(ctx, homestate.ResumeClaim{Token: "old", OwnerPID: -1, TTL: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.DB.Exec(`UPDATE resume_handoffs SET legacy_recovery=1,legacy_recovery_reason='unknown',claim_expires_at=NULL WHERE id=?`, row.ID); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", pathsRoot)
	t.Chdir(cwd)
	cmd := newFactoryCommand()
	cmd.SetArgs([]string{"handoff", "recover-resume", "--id", fmt.Sprint(row.ID), "--expected-token", "old", "--decision", "fail"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("recover-resume read the cwd factory, not the moai-paths root (card t1862): %v", err)
	}
}

// Codex PR #1833 review P2-1: with CLAUDE_PROJECT_DIR unset and the working
// directory deleted, resolveProjectDir() returns "" and factoryCardRoot() fed
// OpenFactory("") — which creates a stray factory DB under the scrubbed home
// instead of failing. Both verbs must return the working-directory error and
// create nothing.
func TestFactoryRunsAndRecoverRefuseUnresolvableRoot(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("MOAI_HOME", filepath.Join(home, ".moai"))
	t.Setenv("MOAI_CLAUDE_BIN", filepath.Join(home, "bin", "claude"))
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	gone := filepath.Join(t.TempDir(), "gone")
	if err := os.MkdirAll(gone, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Chdir(gone)
	if err := os.Remove(gone); err != nil {
		t.Fatal(err)
	}
	out, err := runFactoryCommand(t, "runs")
	if err == nil {
		t.Fatalf("factory runs succeeded with an unresolvable working directory: output %q", out)
	}
	out, err = runFactoryCommand(t, "handoff", "recover-resume", "--id", "1", "--expected-token", "x", "--decision", "fail")
	if err == nil {
		t.Fatalf("recover-resume succeeded with an unresolvable working directory: output %q", out)
	}
	var stray []string
	_ = filepath.Walk(home, func(p string, info os.FileInfo, walkErr error) error {
		if walkErr == nil && !info.IsDir() && strings.HasSuffix(p, ".db") {
			stray = append(stray, p)
		}
		return nil
	})
	if len(stray) != 0 {
		t.Fatalf("a factory DB appeared under the scrubbed home: %v", stray)
	}
}

// Codex PR #1833 review P2-2: relaunchRetireFromRun still rooted at the raw
// working directory, so with CLAUDE_PROJECT_DIR naming project A and the cwd
// naming project B it probed B's factory and reported A's run unknown. It must
// apply the same factory-root rule and retire A's dead-owner run.
func TestRelaunchRetireFromRunOpensMoaiPathsRoot(t *testing.T) {
	pathsRoot, cwd := rootAlignmentSandbox(t)
	deadPID, deadStart := deadOwnerIdentity(t)
	seedRun(t, pathsRoot, "run-relaunch-dead", deadPID, deadStart)
	t.Setenv("CLAUDE_PROJECT_DIR", pathsRoot)
	t.Chdir(cwd)
	outcome := relaunchRetireFromRun(context.Background(), "run-relaunch-dead")
	if strings.Contains(outcome, "unknown (no such run)") {
		t.Fatalf("relaunch probed the cwd factory, not the moai-paths root: %q", outcome)
	}
	if !strings.Contains(outcome, "retired (owner dead)") {
		t.Fatalf("the dead-owner run at the moai-paths root was not retired: %q", outcome)
	}
}
