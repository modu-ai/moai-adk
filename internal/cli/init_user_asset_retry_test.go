// init_user_asset_retry_test.go — card t1587 defect (1): the user-asset
// ensure's failure advice ("Fix the cause and re-run 'moai init' — the
// ensure is idempotent and completes the shortfall") must be a REAL recovery
// path. The ensure used to run AFTER executor.Execute had recorded the
// project as initialized, so a systemic ensure failure (~/.codex/agents as a
// regular file — the twice-observed t1561 shape) left the project marked
// initialized with the shortfall uncompleted, and the advised re-run was
// refused with "project already initialized".
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInitUserAssetFailureThenRetryCompletesInit reproduces the blocked
// recovery end to end: make $HOME/.codex/agents a regular file so the
// user-asset ensure fails systemically, run init (expect the documented
// failure), remove the cause, then re-run init exactly as the advice
// instructs. The re-run must complete the initialization — deploy, ensure,
// and the post-deploy wiring — instead of refusing with "project already
// initialized".
func TestInitUserAssetFailureThenRetryCompletesInit(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")

	projectDir := filepath.Join(t.TempDir(), "retry-proj")

	// Block the user-asset ensure: ~/.codex/agents as a REGULAR FILE makes
	// the codex-agents root unresolvable (resolve root: not a directory —
	// the same systemic class as the observed field failures).
	blocker := filepath.Join(homeDir, ".codex", "agents")
	if err := os.MkdirAll(filepath.Dir(blocker), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(blocker, []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Run 1: init fails at the user-asset ensure — expected, and the failure
	// carries the recovery advice under test.
	cmd := newInitTestCmd()
	if err := cmd.Flags().Set("llm", "claude"); err != nil {
		t.Fatal(err)
	}
	var out1, errBuf1 bytes.Buffer
	cmd.SetOut(&out1)
	cmd.SetErr(&errBuf1)
	runErr := runInit(cmd, []string{projectDir})
	if runErr == nil {
		t.Fatal("run 1: expected the blocked user-asset ensure to fail init")
	}
	if !strings.Contains(runErr.Error(), "user-asset install failed") {
		t.Fatalf("run 1: expected a user-asset install failure, got: %v", runErr)
	}
	if !strings.Contains(runErr.Error(), "re-run 'moai init'") {
		t.Fatalf("run 1: failure does not carry the re-run advice: %v", runErr)
	}

	// Remove the cause, exactly as the advice instructs.
	if err := os.Remove(blocker); err != nil {
		t.Fatal(err)
	}

	// Run 2: the promised retry. It must not hit the already-initialized
	// refusal — the ensure failure must have aborted before the project was
	// recorded initialized.
	cmd2 := newInitTestCmd()
	if err := cmd2.Flags().Set("llm", "claude"); err != nil {
		t.Fatal(err)
	}
	var out2, errBuf2 bytes.Buffer
	cmd2.SetOut(&out2)
	cmd2.SetErr(&errBuf2)
	if err := runInit(cmd2, []string{projectDir}); err != nil {
		t.Fatalf("run 2 (retry after fixing the cause): %v (stderr: %s)", err, errBuf2.String())
	}

	// The retry completed the shortfall: L0 assets under the user folder
	// (Claude + Codex roots) and the project initialized.
	for _, skill := range []string{"moai", "moai-workflow-tdd", "moai-workflow-spec"} {
		if _, err := os.Stat(filepath.Join(homeDir, ".claude", "skills", skill, "SKILL.md")); err != nil {
			t.Errorf("L0 skill %s missing after the retry: %v", skill, err)
		}
	}
	if _, err := os.Stat(filepath.Join(homeDir, ".codex", "agents", "manager-spec.toml")); err != nil {
		t.Errorf("L0 agent missing under ~/.codex/agents after the retry: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".moai")); err != nil {
		t.Errorf("project .moai missing after the retry: %v", err)
	}
}
