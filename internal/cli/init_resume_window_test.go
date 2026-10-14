package cli

// init_resume_window_test.go — card t1619 (re-adjusted placement, leader decision): the
// user-asset ensure runs before executor.Execute (card t1587), while the resume record
// and the init marker open after it, in the post-executor window. These tests pin why
// the record and marker must not move ahead of the executor, and provide the
// post-executor interruption the resume tests inject.

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// injectInitWindowFailure makes the first post-executor window fail. The run stops after
// the resume record and the init marker are written: the state a real interruption after
// the executor leaves (the project is initialized, the post-steps have not run, and the
// rerun finds the marker). Later windows pass. The seam is restored on cleanup.
func injectInitWindowFailure(t *testing.T) {
	t.Helper()
	orig := initWindowFn
	t.Cleanup(func() { initWindowFn = orig })
	calls := 0
	initWindowFn = func(string) error {
		calls++
		if calls == 1 {
			return errors.New("simulated post-executor interruption")
		}
		return nil
	}
}

// TestPreExecutorRecordMakesFreshInitRefused pins the cause of the rejected placement
// (card t1619). validator.go:130 reads a .moai/ directory as "already initialized", and
// the resume record lives under .moai/state/. A record written before the executor
// therefore makes a project that was never initialized read as initialized.
func TestPreExecutorRecordMakesFreshInitRefused(t *testing.T) {
	isolatedInitHome(t)
	projectDir := filepath.Join(t.TempDir(), "pre-executor-record")
	if err := persistResumeWiring(projectDir, agentWiringClaude); err != nil {
		t.Fatal(err)
	}
	err := resumeRun(t, projectDir, "claude")
	if err == nil || !strings.Contains(err.Error(), "already initialized") {
		t.Fatalf("a fresh init after a pre-executor record was not refused as already initialized (err = %v)", err)
	}
}

// TestEnsureFailureBeforeExecutorLeavesNoProjectState pins the invariant the placement
// keeps (card t1587): an ensure that fails before the executor leaves no .moai/ behind,
// so the retry starts from an uninitialized project.
func TestEnsureFailureBeforeExecutorLeavesNoProjectState(t *testing.T) {
	isolatedInitHome(t)
	projectDir := filepath.Join(t.TempDir(), "ensure-first-proj")
	stubEnsure(t, func(homeDir string, selection []string, out io.Writer) error {
		return errors.New("simulated lock acquisition refused")
	})
	if err := resumeRun(t, projectDir, "claude"); err == nil || !strings.Contains(err.Error(), "user-asset install failed") {
		t.Fatalf("the pre-executor ensure failure did not surface (err = %v)", err)
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".moai")); !os.IsNotExist(err) {
		t.Fatalf("the failed pre-executor ensure left .moai/ behind (stat err = %v): the retry would read the project as initialized", err)
	}
}
