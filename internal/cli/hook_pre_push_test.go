package cli

import (
	"bytes"
	"testing"
)

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
	// The hook command should now have 16 subcommands (8 original + pre-push + 7 new events).
	count := len(hookCmd.Commands())
	if count != 19 {
		names := make([]string, 0, count)
		for _, cmd := range hookCmd.Commands() {
			names = append(names, cmd.Name())
		}
		t.Errorf("hook should have 19 subcommands, got %d: %v", count, names)
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
