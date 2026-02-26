package cli

import (
	"bytes"
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
