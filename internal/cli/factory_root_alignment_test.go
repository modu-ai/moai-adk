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
