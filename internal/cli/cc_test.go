package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestCCCmd_Exists(t *testing.T) {
	if ccCmd == nil {
		t.Fatal("ccCmd should not be nil")
	}
}

func TestCCCmd_Use(t *testing.T) {
	if !strings.HasPrefix(ccCmd.Use, "cc") {
		t.Errorf("ccCmd.Use should start with 'cc', got %q", ccCmd.Use)
	}
}

func TestCCCmd_Short(t *testing.T) {
	if ccCmd.Short == "" {
		t.Error("ccCmd.Short should not be empty")
	}
}

func TestCCCmd_IsSubcommandOfRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "cc" {
			found = true
			break
		}
	}
	if !found {
		t.Error("cc should be registered as a subcommand of root")
	}
}

func TestCCCmd_Execution_NoDeps(t *testing.T) {
	// Use a temporary project root to prevent any mutation of real project files.
	// The project root finder is overridden via findProjectRootFn.
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}

	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return tmpDir, nil }
	defer func() { findProjectRootFn = origFn }()

	origDeps := deps
	defer func() { deps = origDeps }()
	deps = nil

	// Override launchClaude to skip actual exec
	origLaunch := launchClaudeFunc
	defer func() { launchClaudeFunc = origLaunch }()

	var launchedProfile string
	launchClaudeFunc = func(profile string, args []string) error {
		launchedProfile = profile
		return nil
	}

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)

	err := ccCmd.RunE(ccCmd, []string{})
	if err != nil {
		t.Fatalf("cc command should not error with nil deps, got: %v", err)
	}

	if launchedProfile != "" {
		t.Errorf("default profile should be empty, got %q", launchedProfile)
	}
}

func TestCCCmd_WithProfile(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()
	deps = nil

	origLaunch := launchClaudeFunc
	defer func() { launchClaudeFunc = origLaunch }()

	var launchedProfile string
	launchClaudeFunc = func(profile string, args []string) error {
		launchedProfile = profile
		return nil
	}

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)

	err := ccCmd.RunE(ccCmd, []string{"-p", "work"})
	if err != nil {
		t.Fatalf("cc -p work should not error, got: %v", err)
	}

	if launchedProfile != "work" {
		t.Errorf("profile should be 'work', got %q", launchedProfile)
	}
}

// ── Characterization tests: capture existing behavior of cc.go (M6-S1 DDD) ──

// TestCharacterize_CC_HelpFlag verifies that -h / --help flags are intercepted
// by runCC before profile parsing and trigger cobra's Help() output.
func TestCharacterize_CC_HelpFlag(t *testing.T) {
	for _, flag := range []string{"--help", "-h"} {
		t.Run(flag, func(t *testing.T) {
			buf := new(bytes.Buffer)
			ccCmd.SetOut(buf)
			ccCmd.SetErr(buf)

			// runCC returns nil after printing help; launchClaudeFunc must NOT be called.
			origLaunch := launchClaudeFunc
			defer func() { launchClaudeFunc = origLaunch }()
			called := false
			launchClaudeFunc = func(_ string, _ []string) error {
				called = true
				return nil
			}

			err := runCC(ccCmd, []string{flag})
			if err != nil {
				t.Errorf("runCC(%s) should not error, got: %v", flag, err)
			}
			if called {
				t.Errorf("launchClaudeFunc must not be called when %s is present", flag)
			}
		})
	}
}

// TestCharacterize_CC_BypassFlag verifies that -b / --bypass pass through to
// unifiedLaunch as extra args (not consumed by parseProfileFlag).
func TestCharacterize_CC_BypassFlag(t *testing.T) {
	origLaunch := unifiedLaunchFunc
	defer func() { unifiedLaunchFunc = origLaunch }()

	var capturedArgs []string
	unifiedLaunchFunc = func(_ string, _ string, args []string) error {
		capturedArgs = args
		return nil
	}

	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return t.TempDir(), nil }
	defer func() { findProjectRootFn = origFn }()

	origDeps := deps
	defer func() { deps = origDeps }()
	deps = nil

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)

	err := runCC(ccCmd, []string{"-b"})
	if err != nil {
		t.Fatalf("runCC(-b) should not error, got: %v", err)
	}
	// -b is not -p/--profile so parseProfileFlag passes it through unchanged.
	if len(capturedArgs) == 0 || capturedArgs[0] != "-b" {
		t.Errorf("expected -b in extra args, got: %v", capturedArgs)
	}
}

// TestCharacterize_CC_ProfileFlag verifies that -p <name> sets profileName
// and is removed from extra args before being forwarded to unifiedLaunch.
func TestCharacterize_CC_ProfileFlag(t *testing.T) {
	origLaunch := unifiedLaunchFunc
	defer func() { unifiedLaunchFunc = origLaunch }()

	var capturedProfile string
	var capturedArgs []string
	unifiedLaunchFunc = func(profile string, _ string, args []string) error {
		capturedProfile = profile
		capturedArgs = args
		return nil
	}

	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return t.TempDir(), nil }
	defer func() { findProjectRootFn = origFn }()

	origDeps := deps
	defer func() { deps = origDeps }()
	deps = nil

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)

	err := runCC(ccCmd, []string{"-p", "myprofile", "--print"})
	if err != nil {
		t.Fatalf("runCC(-p myprofile) should not error, got: %v", err)
	}
	if capturedProfile != "myprofile" {
		t.Errorf("expected profile %q, got %q", "myprofile", capturedProfile)
	}
	// -p + value must be stripped; --print passes through.
	for _, a := range capturedArgs {
		if a == "-p" || a == "myprofile" {
			t.Errorf("profile flag/value must be stripped from extra args, got: %v", capturedArgs)
		}
	}
}

// TestCharacterize_CC_UnknownFlag verifies that unrecognised flags (e.g. --foo)
// are forwarded verbatim to unifiedLaunch as extra args.
func TestCharacterize_CC_UnknownFlag(t *testing.T) {
	origLaunch := unifiedLaunchFunc
	defer func() { unifiedLaunchFunc = origLaunch }()

	var capturedArgs []string
	unifiedLaunchFunc = func(_ string, _ string, args []string) error {
		capturedArgs = args
		return nil
	}

	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return t.TempDir(), nil }
	defer func() { findProjectRootFn = origFn }()

	origDeps := deps
	defer func() { deps = origDeps }()
	deps = nil

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)

	err := runCC(ccCmd, []string{"--unknown-flag", "value"})
	if err != nil {
		t.Fatalf("runCC(--unknown-flag) should not error, got: %v", err)
	}
	if !slices.Contains(capturedArgs, "--unknown-flag") {
		t.Errorf("unknown flag should be forwarded; got args: %v", capturedArgs)
	}
}

// TestCharacterize_CC_LaunchError verifies that errors from unifiedLaunch are
// propagated unchanged by runCC.
func TestCharacterize_CC_LaunchError(t *testing.T) {
	origLaunch := unifiedLaunchFunc
	defer func() { unifiedLaunchFunc = origLaunch }()

	sentinel := errors.New("launch failed: exec not found")
	unifiedLaunchFunc = func(_ string, _ string, _ []string) error {
		return sentinel
	}

	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return t.TempDir(), nil }
	defer func() { findProjectRootFn = origFn }()

	origDeps := deps
	defer func() { deps = origDeps }()
	deps = nil

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)

	err := runCC(ccCmd, []string{})
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error to propagate, got: %v", err)
	}
}

// TestCharacterize_CC_ModeIsAlwaysClaude verifies that runCC always passes
// modeOverride="claude" to unifiedLaunch regardless of extra flags.
func TestCharacterize_CC_ModeIsAlwaysClaude(t *testing.T) {
	origLaunch := unifiedLaunchFunc
	defer func() { unifiedLaunchFunc = origLaunch }()

	var capturedMode string
	unifiedLaunchFunc = func(_ string, mode string, _ []string) error {
		capturedMode = mode
		return nil
	}

	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return t.TempDir(), nil }
	defer func() { findProjectRootFn = origFn }()

	origDeps := deps
	defer func() { deps = origDeps }()
	deps = nil

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)

	err := runCC(ccCmd, []string{})
	if err != nil {
		t.Fatalf("runCC should not error, got: %v", err)
	}
	if capturedMode != "claude" {
		t.Errorf("modeOverride must always be %q, got %q", "claude", capturedMode)
	}
}
