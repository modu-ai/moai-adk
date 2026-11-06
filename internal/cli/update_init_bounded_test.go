package cli

// update_init_bounded_test.go — card t2053: `moai update` and `moai init` must
// not go silent while they wait.
//
//	lock wait   after a short threshold the user-lock wait prints, once, that
//	            another process holds the moai user lock and how long the wait
//	            can last (update and init share the helper).
//	git probe   init's `git remote` detection is bounded; a hang falls back to
//	            the manual default with a reported reason.
//	steps       the binary update check, the pre-migration housekeeping and the
//	            backup prune each carry a progress label.
//
// All fixtures live under t.TempDir(); the blocking `git` is a throwaway
// `exec sleep` stand-in that the timeout under test kills.

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/update"
	"github.com/modu-ai/moai-adk/internal/userassets"
	"github.com/modu-ai/moai-adk/pkg/version"
)

func TestAcquireUserLockNotified_PrintsNoticeOncePastThreshold(t *testing.T) {
	home := t.TempDir()
	held, err := userassets.AcquireUserLock(home, time.Second)
	if err != nil {
		t.Fatalf("hold the lock: %v", err)
	}
	t.Cleanup(func() { _ = held.Release() })

	orig := userLockNoticeDelay
	userLockNoticeDelay = 50 * time.Millisecond
	t.Cleanup(func() { userLockNoticeDelay = orig })

	var out bytes.Buffer
	lock, err := acquireUserLockNotified(home, 400*time.Millisecond, &out)
	if err == nil {
		_ = lock.Release()
		t.Fatalf("acquired a lock that another holder keeps")
	}
	if !errors.Is(err, userassets.ErrLocked) {
		t.Errorf("error %v does not wrap ErrLocked", err)
	}
	got := out.String()
	if n := strings.Count(got, "moai user lock"); n != 1 {
		t.Fatalf("notice printed %d times, want once:\n%s", n, got)
	}
	if !strings.Contains(got, "another process") || !strings.Contains(got, "400ms") {
		t.Errorf("notice names neither the holder nor the maximum wait:\n%s", got)
	}
}

func TestAcquireUserLockNotified_QuietWhenLockIsFree(t *testing.T) {
	orig := userLockNoticeDelay
	userLockNoticeDelay = 20 * time.Millisecond
	t.Cleanup(func() { userLockNoticeDelay = orig })

	home := t.TempDir()
	var out bytes.Buffer
	lock, err := acquireUserLockNotified(home, time.Second, &out)
	if err != nil {
		t.Fatalf("acquire a free lock: %v", err)
	}
	defer func() { _ = lock.Release() }()
	// Wait well past the threshold: the stopped timer must not fire late.
	time.Sleep(100 * time.Millisecond)
	if out.Len() != 0 {
		t.Errorf("a free lock produced output: %q", out.String())
	}
}

// t2053BlockingGit puts a `git` at the front of PATH that hangs for any
// invocation (until the caller's timeout kills it).
func t2053BlockingGit(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "git"), []byte("#!/bin/sh\nexec sleep 30\n"), 0o755); err != nil {
		t.Fatalf("write fake git: %v", err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestDetectGitConfig_GitTimeoutFallsBackToManualWithReason(t *testing.T) {
	t2053BlockingGit(t)
	orig := gitDetectTimeout
	gitDetectTimeout = 300 * time.Millisecond
	t.Cleanup(func() { gitDetectTimeout = orig })

	start := time.Now()
	mode, provider, reason := detectGitConfigReport(t.TempDir())
	if mode != gitModeManual || provider != gitProviderGitHub {
		t.Errorf("fallback = (%q, %q), want (%q, %q)", mode, provider, gitModeManual, gitProviderGitHub)
	}
	if !strings.Contains(reason, "timed out") {
		t.Errorf("reason = %q, want a timed-out reason", reason)
	}
	if elapsed := time.Since(start); elapsed > 10*time.Second {
		t.Errorf("detection took %s; the bound did not hold", elapsed)
	}
	// The two-value wrapper keeps its contract.
	if m, p := detectGitConfig(t.TempDir()); m != gitModeManual || p != gitProviderGitHub {
		t.Errorf("detectGitConfig = (%q, %q)", m, p)
	}
}

func TestDetectGitConfigReport_NoReasonOnPlainAbsence(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skipf("git unavailable: %v", err)
	}
	mode, _, reason := detectGitConfigReport(t.TempDir())
	if mode != gitModeManual || reason != "" {
		t.Errorf("non-git dir = (%q, reason %q), want manual with no reason", mode, reason)
	}
}

func TestRunBinaryUpdateStep_ShowsCheckProgress(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()
	deps = &Dependencies{
		UpdateChecker: &mockUpdateChecker{
			isUpdateAvailFunc: func(string) (bool, *update.VersionInfo, error) { return false, nil, nil },
		},
	}
	origVersion := version.Version
	defer func() { version.Version = origVersion }()
	version.Version = "v2.0.0"

	var buf bytes.Buffer
	cmd := updateCmd
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})
	cmd.SetContext(context.Background())
	if _, err := runBinaryUpdateStep(cmd); err != nil {
		t.Fatalf("runBinaryUpdateStep: %v", err)
	}
	if !strings.Contains(buf.String(), "Checking for a moai binary update") {
		t.Errorf("the binary update check shows no progress label:\n%s", buf.String())
	}
}

func TestRunUpdateProgressStep_LabelsThenCompletes(t *testing.T) {
	var buf bytes.Buffer
	ran := false
	runUpdateProgressStep(&buf, "Tidying leftovers...", "Tidy complete", func() { ran = true })
	if !ran {
		t.Fatalf("the step body did not run")
	}
	got := buf.String()
	open, done := strings.Index(got, "Tidying leftovers..."), strings.Index(got, "Tidy complete")
	if open < 0 || done < 0 || open > done {
		t.Errorf("want the label before the completion line:\n%s", got)
	}
}

func TestPruneBackupsWithProgress_LabelsAndReports(t *testing.T) {
	root := t.TempDir()
	var buf bytes.Buffer
	rep := pruneBackupsWithProgress(&buf, root)
	if rep.Removed != 0 {
		t.Errorf("an empty project removed %d backups", rep.Removed)
	}
	if !strings.Contains(buf.String(), "Pruning old backups") {
		t.Errorf("the backup prune shows no progress label:\n%s", buf.String())
	}
}
