package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
	if runtime.GOOS == "windows" {
		t.Skip("Windows cannot remove the current working directory (PR #1833)")
	}
	for name, args := range map[string][]string{
		"runs":    {"runs"},
		"recover": {"handoff", "recover-resume", "--id", "1", "--expected-token", "x", "--decision", "fail"},
	} {
		t.Run(name, func(t *testing.T) {
			_, cwd := rootAlignmentSandbox(t)
			t.Setenv("CLAUDE_PROJECT_DIR", "")
			t.Chdir(cwd)
			if err := os.Remove(cwd); err != nil {
				t.Fatal(err)
			}
			out, err := runFactoryCommand(t, args...)
			if err == nil {
				t.Fatalf("command succeeded with deleted cwd: output %q", out)
			}
			if !errors.Is(err, os.ErrNotExist) {
				t.Errorf("error lost deleted cwd cause: %v", err)
			}
			assertRootAlignmentNoDB(t)
		})
	}
}

func assertRootAlignmentNoDB(t *testing.T) {
	t.Helper()
	err := filepath.Walk(os.Getenv("HOME"), func(p string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !info.IsDir() && strings.HasSuffix(p, ".db") {
			t.Errorf("unexpected factory DB: %s", p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk sandbox home: %v", err)
	}
}

// Preserve the pre-existing unresolved-directory refusal even when cwd is valid.
func TestVerifyWorkingDirRejectsUnresolvedDirectory(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := verifyWorkingDir(""); err == nil || !strings.Contains(err.Error(), "project directory unresolved") {
		t.Fatalf("verifyWorkingDir(empty) = %v, want unresolved-directory refusal", err)
	}
}

func TestVerifyWorkingDirPreservesStatError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")
	err := verifyWorkingDir(missing)
	var pathErr *os.PathError
	if !errors.Is(err, os.ErrNotExist) || !errors.As(err, &pathErr) || pathErr.Path != missing {
		t.Fatalf("verifyWorkingDir(%q) = %v, want wrapped stat PathError", missing, err)
	}
}

func TestVerifyWorkingDirRejectsFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := verifyWorkingDir(file); err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("verifyWorkingDir(file) = %v, want non-directory error", err)
	}
}

func TestFactoryRootCommandsRejectInvalidProjectDir(t *testing.T) {
	for _, kind := range []string{"missing", "file"} {
		for name, args := range map[string][]string{
			"runs":    {"runs"},
			"recover": {"handoff", "recover-resume", "--id", "1", "--expected-token", "x", "--decision", "fail"},
		} {
			t.Run(kind+"/"+name, func(t *testing.T) {
				_, cwd := rootAlignmentSandbox(t)
				dir := filepath.Join(t.TempDir(), kind)
				if kind == "file" {
					if err := os.WriteFile(dir, nil, 0o600); err != nil {
						t.Fatal(err)
					}
				}
				t.Setenv("CLAUDE_PROJECT_DIR", dir)
				t.Chdir(cwd)
				_, err := runFactoryCommand(t, args...)
				if err == nil {
					t.Fatal("invalid project directory accepted")
				}
				if kind == "missing" && !errors.Is(err, os.ErrNotExist) {
					t.Errorf("missing path cause lost: %v", err)
				}
				if kind == "file" && !strings.Contains(err.Error(), "not a directory") {
					t.Errorf("non-directory cause lost: %v", err)
				}
				assertRootAlignmentNoDB(t)
			})
		}
	}
}

func TestRelaunchRetireFromRunRefusesDeletedCwd(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows cannot remove the current working directory (PR #1833)")
	}
	_, cwd := rootAlignmentSandbox(t)
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	t.Chdir(cwd)
	if err := os.Remove(cwd); err != nil {
		t.Fatal(err)
	}
	out := relaunchRetireFromRun(context.Background(), "run-deleted")
	if !strings.Contains(out, "working directory unavailable") || !strings.Contains(out, "left untouched") {
		t.Fatalf("unexpected relaunch outcome: %q", out)
	}
	assertRootAlignmentNoDB(t)
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
