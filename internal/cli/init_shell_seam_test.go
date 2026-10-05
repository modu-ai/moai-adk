package cli

// Primary observation for the shell-config step seam (SPEC-INIT-QUIET-WIZARD-001
// M1, AC-IQW-004): the real runInit, on the interactive path, carries the
// default SkipShellConfig=false gate all the way to Step 6.

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
)

// TestRunInit_ShellConfigStepReachedViaSeam runs the real runInit with the
// home-safety helper in place and asserts the shell-config seam spy was called
// exactly once. runInit never sets SkipShellConfig, so one call proves the
// default gate reached Step 6; zero means the step was skipped or bypassed.
func TestRunInit_ShellConfigStepReachedViaSeam(t *testing.T) {
	_, spy := prepareSafeInitHome(t)

	// Interactive path without a TTY: swap the stdin and wizard seams.
	origInteractive := isInteractiveStdin
	isInteractiveStdin = func() bool { return true }
	t.Cleanup(func() { isInteractiveStdin = origInteractive })

	origDeps := deps
	deps = nil
	t.Cleanup(func() { deps = origDeps })

	// design.md §6.1 injected result: the four kept answers plus the fixed
	// defaults RunWithDefaults seeds, so opts match a production interactive run.
	wizResult := &wizard.WizardResult{
		ConversationLang:          "en",
		UserName:                  "tester",
		AgentWiring:               "claude",
		AutonomyTier:              "semi-auto",
		LSPEnabled:                true,
		EnforceQuality:            true,
		CoverageExemptionsEnabled: false,
		DesignEnabled:             true,
		ClaudeDesignEnabled:       true,
	}
	origWizard := runWizardFn
	runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) { return wizResult, nil }
	t.Cleanup(func() { runWizardFn = origWizard })

	projectDir := filepath.Join(t.TempDir(), "shell-seam-proj")
	cmd := newInitTestCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)

	if err := runInit(cmd, []string{projectDir}); err != nil {
		t.Fatalf("runInit: %v (stderr: %s)", err, errBuf.String())
	}

	if got := spy.Calls(); got != 1 {
		t.Errorf("shell-config seam calls = %d, want exactly 1 (runInit must carry the default SkipShellConfig=false gate to Step 6)", got)
	}
}
