package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"
)

// --- Integration tests covering multiple commands and DI wiring ---

func TestExecute_InitsDeps(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = nil

	// Execute with --help to avoid side effects
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})

	err := Execute()
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	if deps == nil {
		t.Error("Execute should initialize deps")
	}

	// Reset args for other tests
	rootCmd.SetArgs(nil)
}

func TestRootCmd_AllCommandsRegistered(t *testing.T) {
	expected := []string{
		"init", "doctor", "status", "version",
		"update", "hook", "cc", "glm", "rank",
		"worktree", "statusline",
	}

	registered := make(map[string]bool)
	for _, cmd := range rootCmd.Commands() {
		registered[cmd.Name()] = true
	}

	for _, name := range expected {
		if !registered[name] {
			t.Errorf("command %q should be registered on root", name)
		}
	}
}

func TestUpdateCmd_DefaultIsTemplateSync(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	// Run in a temp directory to avoid polluting the source tree with deployed templates.
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Default moai update should run template sync, not binary update.
	// Even with nil deps, the command should proceed to template sync.
	deps = nil

	buf := new(bytes.Buffer)
	updateCmd.SetOut(buf)
	updateCmd.SetErr(buf)

	// Ensure check flag is false (default flow)
	if err := updateCmd.Flags().Set("check", "false"); err != nil {
		t.Fatal(err)
	}

	err = updateCmd.RunE(updateCmd, []string{})

	// Template sync may fail in test environment (no project root, etc.)
	// but the error should NOT be about orchestrator or "not initialized".
	if err != nil {
		if strings.Contains(err.Error(), "not initialized") {
			t.Errorf("default update should not require orchestrator, got: %v", err)
		}
		if strings.Contains(err.Error(), "orchestrator") {
			t.Errorf("default update should not reference orchestrator, got: %v", err)
		}
	}

	output := buf.String()
	if !strings.Contains(output, "Current version") {
		t.Errorf("output should contain version info, got %q", output)
	}
}

func TestHookEventCmd_NilDeps(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = nil

	// Find session-start subcommand
	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "session-start" {
			err := cmd.RunE(cmd, []string{})
			if err == nil {
				t.Error("hook session-start with nil deps should error")
			}
			if !strings.Contains(err.Error(), "not initialized") {
				t.Errorf("error should mention not initialized, got: %v", err)
			}
			return
		}
	}
	t.Error("session-start subcommand not found")
}

func TestRankLogin_NilDeps(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = nil

	for _, cmd := range rankCmd.Commands() {
		if cmd.Name() == "login" {
			err := cmd.RunE(cmd, []string{})
			if err == nil {
				t.Error("rank login with nil deps should error")
			}
			return
		}
	}
	t.Error("login subcommand not found")
}

func TestRankLogout_NilDeps(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = nil

	for _, cmd := range rankCmd.Commands() {
		if cmd.Name() == "logout" {
			err := cmd.RunE(cmd, []string{})
			if err == nil {
				t.Error("rank logout with nil deps should error")
			}
			return
		}
	}
	t.Error("logout subcommand not found")
}

func TestRankStatus_NilRankClient(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = &Dependencies{}

	for _, cmd := range rankCmd.Commands() {
		if cmd.Name() == "status" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("rank status with nil client should not error, got: %v", err)
			}
			if !strings.Contains(buf.String(), "not configured") {
				t.Errorf("output should say not configured, got %q", buf.String())
			}
			return
		}
	}
	t.Error("status subcommand not found")
}

func TestRankSync_Output(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	// Set up minimal deps for sync command (no auth = not logged in message)
	deps = &Dependencies{}

	for _, cmd := range rankCmd.Commands() {
		if cmd.Name() == "sync" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("rank sync error: %v", err)
			}
			output := buf.String()
			// Should show not logged in message when no credentials
			if !strings.Contains(output, "Not logged in") && !strings.Contains(output, "Sync complete") {
				t.Errorf("output should contain login or sync message, got %q", output)
			}
			return
		}
	}
	t.Error("sync subcommand not found")
}

func TestRankExclude_Output(t *testing.T) {
	for _, cmd := range rankCmd.Commands() {
		if cmd.Name() == "exclude" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			// Use unique pattern to avoid conflicts with previous test runs
			pattern := "*.test-" + time.Now().Format("20060102-150405")
			err := cmd.RunE(cmd, []string{pattern})
			if err != nil {
				t.Fatalf("rank exclude error: %v", err)
			}
			if !strings.Contains(buf.String(), pattern) {
				t.Errorf("output should contain pattern, got %q", buf.String())
			}
			return
		}
	}
	t.Error("exclude subcommand not found")
}

func TestRankInclude_Output(t *testing.T) {
	for _, cmd := range rankCmd.Commands() {
		if cmd.Name() == "include" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			// Use unique pattern to avoid conflicts with previous test runs
			pattern := "*.go-" + time.Now().Format("20060102-150405")
			err := cmd.RunE(cmd, []string{pattern})
			if err != nil {
				t.Fatalf("rank include error: %v", err)
			}
			if !strings.Contains(buf.String(), pattern) {
				t.Errorf("output should contain pattern, got %q", buf.String())
			}
			return
		}
	}
	t.Error("include subcommand not found")
}

func TestRankRegister_Output(t *testing.T) {
	for _, cmd := range rankCmd.Commands() {
		if cmd.Name() == "register" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{"my-org"})
			if err != nil {
				t.Fatalf("rank register error: %v", err)
			}
			if !strings.Contains(buf.String(), "my-org") {
				t.Errorf("output should contain org name, got %q", buf.String())
			}
			return
		}
	}
	t.Error("register subcommand not found")
}

func TestStatuslineCmd_WithDeps(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	InitDependencies()

	buf := new(bytes.Buffer)
	StatuslineCmd.SetOut(buf)
	StatuslineCmd.SetErr(buf)

	err := StatuslineCmd.RunE(StatuslineCmd, []string{})
	if err != nil {
		t.Fatalf("statusline with deps error: %v", err)
	}

	output := buf.String()
	// Statusline should produce some output (git status, version, branch, or fallback)
	output = strings.TrimSpace(output)
	if output == "" {
		t.Errorf("output should not be empty")
	}
	// The new independent collection shows git status, version, and branch
	// Check for any of the statusline indicators (emoji or content)
	if !strings.Contains(output, "📊") && !strings.Contains(output, "🔅") && !strings.Contains(output, "🔀") {
		// If no indicators, at least check for some content
		if len(output) < 3 {
			t.Errorf("output should have meaningful content, got %q", output)
		}
	}
}

func TestDoctorCmd_ExportFlag(t *testing.T) {
	tmpDir := t.TempDir()
	exportPath := tmpDir + "/diag.json"

	buf := new(bytes.Buffer)
	doctorCmd.SetOut(buf)
	doctorCmd.SetErr(buf)

	// Set export flag
	if err := doctorCmd.Flags().Set("export", exportPath); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := doctorCmd.Flags().Set("export", ""); err != nil {
			t.Logf("reset flag: %v", err)
		}
	}()

	err := doctorCmd.RunE(doctorCmd, []string{})
	if err != nil {
		t.Fatalf("doctor --export error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "exported") {
		t.Errorf("output should mention export, got %q", output)
	}
}

func TestDoctorCmd_VerboseExecution(t *testing.T) {
	buf := new(bytes.Buffer)
	doctorCmd.SetOut(buf)
	doctorCmd.SetErr(buf)

	// Set verbose flag
	if err := doctorCmd.Flags().Set("verbose", "true"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := doctorCmd.Flags().Set("verbose", "false"); err != nil {
			t.Logf("reset flag: %v", err)
		}
	}()

	err := doctorCmd.RunE(doctorCmd, []string{})
	if err != nil {
		t.Fatalf("doctor --verbose error: %v", err)
	}

	// Verbose should include details like GOPATH
	output := buf.String()
	if !strings.Contains(output, "GOPATH") {
		t.Errorf("verbose output should contain GOPATH detail, got %q", output)
	}
}
