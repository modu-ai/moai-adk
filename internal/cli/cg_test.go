package cli

// ── Characterization tests for cg.go (M6-S2 DDD PRESERVE) ──
//
// cg.go is a thin delegate: parseProfileFlag → unifiedLaunch(_, "claude_glm", _).
// Tests mirror cc_test.go pattern: mock unifiedLaunchFunc, verify mode constant,
// profile extraction, flag pass-through, and error propagation.

import (
	"bytes"
	"errors"
	"slices"
	"strings"
	"testing"
)

// TestCharacterize_CG_CmdExists verifies cgCmd is registered and non-nil.
func TestCharacterize_CG_CmdExists(t *testing.T) {
	if cgCmd == nil {
		t.Fatal("cgCmd should not be nil")
	}
}

// TestCharacterize_CG_UsePrefix verifies the Use field starts with "cg".
func TestCharacterize_CG_UsePrefix(t *testing.T) {
	if !strings.HasPrefix(cgCmd.Use, "cg") {
		t.Errorf("cgCmd.Use should start with %q, got %q", "cg", cgCmd.Use)
	}
}

// TestCharacterize_CG_IsSubcommandOfRoot verifies cg is registered under rootCmd.
func TestCharacterize_CG_IsSubcommandOfRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "cg" {
			found = true
			break
		}
	}
	if !found {
		t.Error("cg should be registered as a subcommand of root")
	}
}

// TestCharacterize_CG_ModeIsAlwaysClaudeGLM verifies runCG always passes
// modeOverride="claude_glm" to unifiedLaunch regardless of extra flags.
func TestCharacterize_CG_ModeIsAlwaysClaudeGLM(t *testing.T) {
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

	err := runCG(cgCmd, []string{})
	if err != nil {
		t.Fatalf("runCG should not error, got: %v", err)
	}
	const wantMode = "claude_glm"
	if capturedMode != wantMode {
		t.Errorf("modeOverride must always be %q, got %q", wantMode, capturedMode)
	}
}

// TestCharacterize_CG_ProfileFlag verifies -p <name> is consumed by parseProfileFlag
// and forwarded as profileName, with flag pair removed from extra args.
func TestCharacterize_CG_ProfileFlag(t *testing.T) {
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

	err := runCG(cgCmd, []string{"-p", "team", "--print"})
	if err != nil {
		t.Fatalf("runCG(-p team) should not error, got: %v", err)
	}
	if capturedProfile != "team" {
		t.Errorf("expected profile %q, got %q", "team", capturedProfile)
	}
	// -p and value must be stripped; --print passes through
	for _, a := range capturedArgs {
		if a == "-p" || a == "team" {
			t.Errorf("profile flag/value must be stripped from extra args, got: %v", capturedArgs)
		}
	}
	if !slices.Contains(capturedArgs, "--print") {
		t.Errorf("--print should be preserved in extra args, got: %v", capturedArgs)
	}
}

// TestCharacterize_CG_UnknownFlagPassThrough verifies unrecognised flags are
// forwarded verbatim to unifiedLaunch as extra args.
func TestCharacterize_CG_UnknownFlagPassThrough(t *testing.T) {
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

	err := runCG(cgCmd, []string{"--some-claude-flag"})
	if err != nil {
		t.Fatalf("runCG(--some-claude-flag) should not error, got: %v", err)
	}
	if !slices.Contains(capturedArgs, "--some-claude-flag") {
		t.Errorf("unknown flag should be forwarded; got args: %v", capturedArgs)
	}
}

// TestCharacterize_CG_LaunchErrorPropagated verifies errors from unifiedLaunch
// are propagated unchanged by runCG.
func TestCharacterize_CG_LaunchErrorPropagated(t *testing.T) {
	origLaunch := unifiedLaunchFunc
	defer func() { unifiedLaunchFunc = origLaunch }()

	sentinel := errors.New("tmux not found")
	unifiedLaunchFunc = func(_ string, _ string, _ []string) error {
		return sentinel
	}

	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return t.TempDir(), nil }
	defer func() { findProjectRootFn = origFn }()

	err := runCG(cgCmd, []string{})
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error to propagate, got: %v", err)
	}
}

// TestCharacterize_CG_HelpFlag verifies that -h / --help flags are intercepted
// by runCG before profile parsing and the tmux precondition check, returning
// nil without invoking unifiedLaunch. Mirrors TestCharacterize_CC_HelpFlag.
// Regression guard for the cg --help exit-1 fix (help must short-circuit before
// the tmux precondition error).
func TestCharacterize_CG_HelpFlag(t *testing.T) {
	for _, flag := range []string{"--help", "-h"} {
		t.Run(flag, func(t *testing.T) {
			buf := new(bytes.Buffer)
			cgCmd.SetOut(buf)
			cgCmd.SetErr(buf)

			// runCG returns nil after printing help; unifiedLaunchFunc must NOT be called.
			origLaunch := unifiedLaunchFunc
			defer func() { unifiedLaunchFunc = origLaunch }()
			called := false
			unifiedLaunchFunc = func(_ string, _ string, _ []string) error {
				called = true
				return nil
			}

			err := runCG(cgCmd, []string{flag})
			if err != nil {
				t.Errorf("runCG(%s) should not error, got: %v", flag, err)
			}
			if called {
				t.Errorf("unifiedLaunchFunc must not be called when %s is present", flag)
			}
		})
	}
}
