package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
)

// loadPrePushTestDeps builds a *Dependencies whose Config has the active
// git_strategy.mode set to mode and that mode's hooks.pre_push set to prePush.
//
// Injection note: the Loader.Load() chain does NOT read a git-strategy.yaml file
// into cfg.GitStrategy (the whole git_strategy section is itself unwired — see
// internal/config/loader.go, which has no loadGitStrategySection). The only
// production-real override path is ConfigManager.SetSection("git_strategy", ...),
// which is exactly what this helper uses to deterministically place a known
// GitStrategyConfig into the injected deps.Config. When mode is "" the resulting
// ActiveModeProfile() returns (nil, false).
func loadPrePushTestDeps(t *testing.T, mode, prePush string) *Dependencies {
	t.Helper()

	tmpDir := t.TempDir()
	sectionsDir := filepath.Join(tmpDir, defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cfgMgr := config.NewConfigManager()
	if _, err := cfgMgr.Load(tmpDir); err != nil {
		t.Fatalf("load config: %v", err)
	}

	// Override the git_strategy section with a known mode + pre_push value.
	gs := config.GitStrategyConfig{Mode: mode}
	switch mode {
	case "manual":
		gs.Manual.Hooks.PrePush = prePush
	case "personal":
		gs.Personal.Hooks.PrePush = prePush
	case "team":
		gs.Team.Hooks.PrePush = prePush
	}
	if err := cfgMgr.SetSection("git_strategy", gs); err != nil {
		t.Fatalf("set git_strategy section: %v", err)
	}

	return &Dependencies{Config: cfgMgr}
}

// withPrePushTestDeps temporarily installs d as the package-level deps and
// restores the original on cleanup.
func withPrePushTestDeps(t *testing.T, d *Dependencies) {
	t.Helper()
	origDeps := deps
	t.Cleanup(func() { deps = origDeps })
	deps = d
}

// -----------------------------------------------------------------------------
// SPEC-PREPUSH-MODE-WIRING-001 — resolvePrePushAction() pure-helper decision
// tests (REQ-PMW-001/008/010/011, AC-PMW-002/003/006/007/009). These assert the
// resolver's action enum, NOT process exit or stdin (REQ-PMW-002a seam).
// -----------------------------------------------------------------------------

// AC-PMW-002: gate ON + active-mode pre_push=enforce ⇒ resolver returns enforce.
func TestResolvePrePushAction_GateOnEnforce(t *testing.T) {
	t.Setenv("MOAI_ENFORCE_ON_PUSH", "")
	t.Setenv("MOAI_PRE_PUSH", "")
	withPrePushTestDeps(t, loadPrePushTestDeps(t, "team", "enforce"))

	got := resolvePrePushAction()
	if got != prePushEnforce {
		t.Errorf("resolvePrePushAction() = %v, want prePushEnforce", got)
	}
}

// AC-PMW-003: gate ON + active-mode pre_push=warn ⇒ resolver returns warn.
func TestResolvePrePushAction_GateOnWarn(t *testing.T) {
	t.Setenv("MOAI_ENFORCE_ON_PUSH", "")
	t.Setenv("MOAI_PRE_PUSH", "")
	withPrePushTestDeps(t, loadPrePushTestDeps(t, "team", "warn"))

	got := resolvePrePushAction()
	if got != prePushWarn {
		t.Errorf("resolvePrePushAction() = %v, want prePushWarn", got)
	}
}

// AC-PMW-004 (decision side): gate ON + active-mode pre_push=skip ⇒ resolver
// returns skip.
func TestResolvePrePushAction_GateOnSkip(t *testing.T) {
	t.Setenv("MOAI_ENFORCE_ON_PUSH", "")
	t.Setenv("MOAI_PRE_PUSH", "")
	withPrePushTestDeps(t, loadPrePushTestDeps(t, "personal", "skip"))

	got := resolvePrePushAction()
	if got != prePushSkip {
		t.Errorf("resolvePrePushAction() = %v, want prePushSkip", got)
	}
}

// AC-PMW-006: gate ON + Mode empty (ActiveModeProfile false) ⇒ default enforce.
func TestResolvePrePushAction_NilProfile(t *testing.T) {
	t.Setenv("MOAI_ENFORCE_ON_PUSH", "")
	t.Setenv("MOAI_PRE_PUSH", "")
	withPrePushTestDeps(t, loadPrePushTestDeps(t, "", ""))

	got := resolvePrePushAction()
	if got != prePushEnforce {
		t.Errorf("resolvePrePushAction() with nil ModeProfile = %v, want prePushEnforce", got)
	}
}

// AC-PMW-007: gate ON + unknown pre_push value ⇒ normalized to enforce.
func TestResolvePrePushAction_UnknownValue(t *testing.T) {
	t.Setenv("MOAI_ENFORCE_ON_PUSH", "")
	t.Setenv("MOAI_PRE_PUSH", "")
	withPrePushTestDeps(t, loadPrePushTestDeps(t, "team", "garbage"))

	got := resolvePrePushAction()
	if got != prePushEnforce {
		t.Errorf("resolvePrePushAction() with unknown value = %v, want prePushEnforce", got)
	}
}

// -----------------------------------------------------------------------------
// SPEC-PREPUSH-MODE-WIRING-001 — decideExit() pure-helper decision tests
// (REQ-PMW-002a, AC-PMW-002/003/005). Asserts the intended exit code WITHOUT
// calling os.Exit.
// -----------------------------------------------------------------------------

// AC-PMW-002 (decision side): decideExit(enforce, ≥1) ⇒ 2 (block).
func TestDecideExit_EnforceViolation(t *testing.T) {
	if got := decideExit(prePushEnforce, 1); got != 2 {
		t.Errorf("decideExit(enforce, 1) = %d, want 2", got)
	}
	if got := decideExit(prePushEnforce, 5); got != 2 {
		t.Errorf("decideExit(enforce, 5) = %d, want 2", got)
	}
}

// AC-PMW-003 (decision side): decideExit(warn, ≥1) ⇒ 0 (non-blocking).
func TestDecideExit_WarnViolation(t *testing.T) {
	if got := decideExit(prePushWarn, 1); got != 0 {
		t.Errorf("decideExit(warn, 1) = %d, want 0", got)
	}
	if got := decideExit(prePushWarn, 5); got != 0 {
		t.Errorf("decideExit(warn, 5) = %d, want 0", got)
	}
}

// AC-PMW-005: decideExit(*, 0) ⇒ 0 for both enforce and warn (clean path).
func TestDecideExit_CleanCommits(t *testing.T) {
	if got := decideExit(prePushEnforce, 0); got != 0 {
		t.Errorf("decideExit(enforce, 0) = %d, want 0", got)
	}
	if got := decideExit(prePushWarn, 0); got != 0 {
		t.Errorf("decideExit(warn, 0) = %d, want 0", got)
	}
}

// skip is always non-blocking regardless of violation count.
func TestDecideExit_SkipNoOp(t *testing.T) {
	if got := decideExit(prePushSkip, 0); got != 0 {
		t.Errorf("decideExit(skip, 0) = %d, want 0", got)
	}
	if got := decideExit(prePushSkip, 3); got != 0 {
		t.Errorf("decideExit(skip, 3) = %d, want 0", got)
	}
}

// -----------------------------------------------------------------------------
// SPEC-PREPUSH-MODE-WIRING-001 — gate-OFF regression (AC-PMW-001/013). The
// gate-OFF row IS reachable in-process (runPrePush short-circuits before any
// stdin read or os.Exit).
// -----------------------------------------------------------------------------

// AC-PMW-013: gate OFF ⇒ runPrePush returns nil immediately and pre_push has
// zero effect even when set to skip/garbage (resolver never reached).
func TestRunPrePush_GateOff_PrePushNotConsulted(t *testing.T) {
	t.Setenv("MOAI_ENFORCE_ON_PUSH", "false")
	t.Setenv("MOAI_PRE_PUSH", "")

	for _, prePush := range []string{"skip", "garbage", "warn", "enforce"} {
		// Even with a mode profile that has a non-default pre_push value, the
		// gate-OFF short-circuit means runPrePush returns nil with no effect.
		withPrePushTestDeps(t, loadPrePushTestDeps(t, "team", prePush))

		cmd := &cobra.Command{Use: "pre-push-test"}
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		if err := runPrePush(cmd, nil); err != nil {
			t.Errorf("runPrePush gate-OFF with pre_push=%q should return nil, got: %v", prePush, err)
		}
		if buf.Len() != 0 {
			t.Errorf("runPrePush gate-OFF with pre_push=%q should produce no output, got: %q", prePush, buf.String())
		}
	}
}

// AC-PMW-004 (runtime side): gate ON + pre_push=skip ⇒ runPrePush returns nil
// without loading the convention or producing output (the skip short-circuit is
// reachable in-process because it precedes the convention load + stdin read).
func TestRunPrePush_Skip_GateOnNoOp(t *testing.T) {
	t.Setenv("MOAI_ENFORCE_ON_PUSH", "true")
	t.Setenv("MOAI_PRE_PUSH", "")
	withPrePushTestDeps(t, loadPrePushTestDeps(t, "team", "skip"))

	cmd := &cobra.Command{Use: "pre-push-test"}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := runPrePush(cmd, nil); err != nil {
		t.Errorf("runPrePush gate-ON + skip should return nil, got: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("runPrePush gate-ON + skip should produce no output, got: %q", buf.String())
	}
}

// AC-PMW-003 (warn print-loop side-effect, named test): gate ON + warn resolves
// past the skip short-circuit into the convention-load + validate-and-print path.
// With empty stdin this exits 0 via the "No commit messages" branch, proving the
// warn branch does NOT os.Exit (it reaches the non-blocking completion path). The
// blocking decision for warn+violations is asserted at the pure level by
// TestDecideExit_WarnViolation; this test names the runtime print-path seam.
func TestRunPrePush_WarnBranch_GateOnEmptyStdin(t *testing.T) {
	t.Setenv("MOAI_ENFORCE_ON_PUSH", "true")
	t.Setenv("MOAI_PRE_PUSH", "")
	// Point repo root at an empty tmpDir so convention load is deterministic.
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	withPrePushTestDeps(t, loadPrePushTestDeps(t, "team", "warn"))

	cmd := &cobra.Command{Use: "pre-push-test"}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// runPrePush reaches the convention loader + stdin read; with empty stdin it
	// returns nil (non-blocking). It MUST NOT os.Exit on the warn branch.
	if err := runPrePush(cmd, nil); err != nil {
		t.Logf("runPrePush warn-branch returned error (acceptable — convention load in empty tmpDir): %v", err)
	}
}

// -----------------------------------------------------------------------------
// SPEC-PREPUSH-MODE-WIRING-001 — MOAI_PRE_PUSH env severity override
// (REQ-PMW-012, AC-PMW-012). The env severity sits BELOW the gate.
// -----------------------------------------------------------------------------

// AC-PMW-012: env MOAI_PRE_PUSH=warn wins over config pre_push=enforce (gate ON).
func TestResolvePrePushAction_EnvOverride(t *testing.T) {
	t.Setenv("MOAI_ENFORCE_ON_PUSH", "")
	t.Setenv("MOAI_PRE_PUSH", "warn")
	withPrePushTestDeps(t, loadPrePushTestDeps(t, "team", "enforce"))

	got := resolvePrePushAction()
	if got != prePushWarn {
		t.Errorf("resolvePrePushAction() with MOAI_PRE_PUSH=warn = %v, want prePushWarn (env wins over config enforce)", got)
	}
}

// An unrecognized MOAI_PRE_PUSH value is ignored (falls through to the config
// pre_push, then REQ-PMW-011 normalization). Config=skip should win here.
func TestResolvePrePushAction_EnvOverrideUnknownFallsThrough(t *testing.T) {
	t.Setenv("MOAI_ENFORCE_ON_PUSH", "")
	t.Setenv("MOAI_PRE_PUSH", "garbage")
	withPrePushTestDeps(t, loadPrePushTestDeps(t, "personal", "skip"))

	got := resolvePrePushAction()
	if got != prePushSkip {
		t.Errorf("resolvePrePushAction() with unknown MOAI_PRE_PUSH = %v, want prePushSkip (falls through to config)", got)
	}
}

// AC-PMW-012 (gate boundary): MOAI_PRE_PUSH set while enforce_on_push=false does
// NOT turn the gate on — runPrePush is still a no-op (env severity ≠ gate).
func TestRunPrePush_EnvSeverity_GateOff(t *testing.T) {
	t.Setenv("MOAI_ENFORCE_ON_PUSH", "false")
	t.Setenv("MOAI_PRE_PUSH", "enforce")
	withPrePushTestDeps(t, loadPrePushTestDeps(t, "team", "enforce"))

	cmd := &cobra.Command{Use: "pre-push-test"}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := runPrePush(cmd, nil); err != nil {
		t.Errorf("runPrePush with MOAI_PRE_PUSH=enforce but gate OFF should return nil, got: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("runPrePush with MOAI_PRE_PUSH=enforce but gate OFF should produce no output, got: %q", buf.String())
	}
}

func TestPrePushCmd_Exists(t *testing.T) {
	if prePushCmd == nil {
		t.Fatal("prePushCmd should not be nil")
	}
}

func TestPrePushCmd_Use(t *testing.T) {
	if prePushCmd.Use != "pre-push" {
		t.Errorf("prePushCmd.Use = %q, want %q", prePushCmd.Use, "pre-push")
	}
}

func TestPrePushCmd_Short(t *testing.T) {
	if prePushCmd.Short == "" {
		t.Error("prePushCmd.Short should not be empty")
	}
}

func TestPrePushCmd_IsSubcommandOfHook(t *testing.T) {
	found := false
	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "pre-push" {
			found = true
			break
		}
	}
	if !found {
		t.Error("pre-push should be registered as a subcommand of hook")
	}
}

func TestResolveConventionName_Default(t *testing.T) {
	// With no env and no deps, should return "auto".
	origDeps := deps
	defer func() { deps = origDeps }()
	deps = nil

	t.Setenv("MOAI_GIT_CONVENTION", "")

	name := resolveConventionName()
	if name != "auto" {
		t.Errorf("resolveConventionName() = %q, want %q", name, "auto")
	}
}

func TestResolveConventionName_EnvOverride(t *testing.T) {
	t.Setenv("MOAI_GIT_CONVENTION", "angular")

	name := resolveConventionName()
	if name != "angular" {
		t.Errorf("resolveConventionName() = %q, want %q", name, "angular")
	}
}

func TestReadStdinLines_Empty(t *testing.T) {
	// readStdinLines reads from /dev/stdin which may not be available in test.
	// The function handles this gracefully by returning nil.
	// We test the line parsing logic indirectly through the command.
	lines, err := readStdinLines()
	if err != nil {
		t.Logf("readStdinLines returned error (expected in test): %v", err)
	}
	_ = lines // may be nil or empty, both acceptable
}

func TestHookCmd_PrePushSubcommandCount(t *testing.T) {
	// 36 previous - 1 "setup" (removed by SPEC-V3R2-MIG-002 M2.1) = 35.
	// +1 "harness-classify" (added by SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001) = 36.
	count := len(hookCmd.Commands())
	if count != 36 {
		names := make([]string, 0, count)
		for _, cmd := range hookCmd.Commands() {
			names = append(names, cmd.Name())
		}
		t.Errorf("hook should have 36 subcommands, got %d: %v", count, names)
	}
}

func TestHookCmd_HasPrePushSubcommand(t *testing.T) {
	expected := []string{
		"session-start", "pre-tool", "post-tool", "session-end",
		"stop", "compact", "list", "agent", "pre-push",
	}
	for _, name := range expected {
		found := false
		for _, cmd := range hookCmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("hook should have %q subcommand", name)
		}
	}
}

func TestPrePushCmd_OutputFormat(t *testing.T) {
	// Verify the command produces output to the designated writer.
	buf := new(bytes.Buffer)
	prePushCmd.SetOut(buf)
	prePushCmd.SetErr(buf)

	// We cannot easily test the full RunE because it reads stdin,
	// but we can verify the command is properly configured.
	if prePushCmd.RunE == nil {
		t.Error("prePushCmd.RunE should not be nil")
	}
}

func TestIsEnforceOnPushEnabled_Default(t *testing.T) {
	// With no env and no deps, should return false.
	origDeps := deps
	defer func() { deps = origDeps }()
	deps = nil

	t.Setenv("MOAI_ENFORCE_ON_PUSH", "")

	if isEnforceOnPushEnabled() {
		t.Error("isEnforceOnPushEnabled() should return false by default")
	}
}

func TestIsEnforceOnPushEnabled_EnvTrue(t *testing.T) {
	t.Setenv("MOAI_ENFORCE_ON_PUSH", "true")

	if !isEnforceOnPushEnabled() {
		t.Error("isEnforceOnPushEnabled() should return true when env is 'true'")
	}
}

func TestIsEnforceOnPushEnabled_EnvOne(t *testing.T) {
	t.Setenv("MOAI_ENFORCE_ON_PUSH", "1")

	if !isEnforceOnPushEnabled() {
		t.Error("isEnforceOnPushEnabled() should return true when env is '1'")
	}
}

func TestIsEnforceOnPushEnabled_EnvFalse(t *testing.T) {
	t.Setenv("MOAI_ENFORCE_ON_PUSH", "false")

	if isEnforceOnPushEnabled() {
		t.Error("isEnforceOnPushEnabled() should return false when env is 'false'")
	}
}
