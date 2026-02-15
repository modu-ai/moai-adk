package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- DDD PRESERVE: Characterization tests for doctor command behavior ---

func TestDoctorCmd_Exists(t *testing.T) {
	if doctorCmd == nil {
		t.Fatal("doctorCmd should not be nil")
	}
}

func TestDoctorCmd_Use(t *testing.T) {
	if doctorCmd.Use != "doctor" {
		t.Errorf("doctorCmd.Use = %q, want %q", doctorCmd.Use, "doctor")
	}
}

func TestDoctorCmd_Short(t *testing.T) {
	if doctorCmd.Short == "" {
		t.Error("doctorCmd.Short should not be empty")
	}
}

func TestDoctorCmd_Long(t *testing.T) {
	if doctorCmd.Long == "" {
		t.Error("doctorCmd.Long should not be empty")
	}
}

func TestDoctorCmd_IsSubcommandOfRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "doctor" {
			found = true
			break
		}
	}
	if !found {
		t.Error("doctor should be registered as a subcommand of root")
	}
}

func TestDoctorCmd_HasFlags(t *testing.T) {
	flags := []string{"verbose", "fix", "export", "check"}
	for _, name := range flags {
		if doctorCmd.Flags().Lookup(name) == nil {
			t.Errorf("doctor command should have --%s flag", name)
		}
	}
}

func TestDoctorCmd_VerboseShortFlag(t *testing.T) {
	f := doctorCmd.Flags().ShorthandLookup("v")
	if f == nil {
		t.Error("doctor command should have -v shorthand for --verbose")
	}
}

func TestDoctorCmd_Execution(t *testing.T) {
	buf := new(bytes.Buffer)
	doctorCmd.SetOut(buf)
	doctorCmd.SetErr(buf)

	err := doctorCmd.RunE(doctorCmd, []string{})
	if err != nil {
		t.Fatalf("doctor command RunE error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "System Diagnostics") {
		t.Errorf("doctor output should contain 'System Diagnostics', got %q", output)
	}
	if !strings.Contains(output, "passed") {
		t.Errorf("doctor output should contain 'passed' in summary, got %q", output)
	}
}

func TestDoctorCmd_HelpOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	doctorCmd.SetOut(buf)
	doctorCmd.SetErr(buf)

	// Use UsageString() instead of Execute() to avoid global state issues
	usage := doctorCmd.UsageString()
	if !strings.Contains(usage, "doctor") {
		t.Error("doctor usage should contain 'doctor'")
	}

	// Verify Long description covers diagnostics/health
	if !strings.Contains(doctorCmd.Long, "diagnostic") || !strings.Contains(doctorCmd.Long, "health") {
		t.Error("doctor Long description should mention diagnostics/health checks")
	}
}

// --- TDD: Tests for GitInstallHint ---

func TestGitInstallHint(t *testing.T) {
	hint := GitInstallHint()
	if hint == "" {
		t.Error("GitInstallHint() returned empty string")
	}
	if !strings.Contains(hint, "git") {
		t.Errorf("GitInstallHint() = %q, expected to contain 'git'", hint)
	}
	if !strings.Contains(hint, "Install") {
		t.Errorf("GitInstallHint() = %q, expected to contain 'Install'", hint)
	}
}

func TestCheckGit_DetailWhenMissing(t *testing.T) {
	// We can only test the Detail field indirectly:
	// when git IS available, Detail should be empty (non-verbose)
	// when git is NOT available, Detail should contain install hint
	check := checkGit(false)
	if check.Status == CheckFail {
		if check.Detail == "" {
			t.Error("checkGit should set Detail with install hint when git is not found")
		}
		if !strings.Contains(check.Detail, "Install git") {
			t.Errorf("checkGit Detail = %q, expected to contain 'Install git'", check.Detail)
		}
	}
	// If git is available, Detail should be empty in non-verbose mode
	if check.Status == CheckOK && check.Detail != "" {
		t.Errorf("checkGit Detail should be empty in non-verbose mode when git is found, got %q", check.Detail)
	}
}

// --- TDD: Tests for diagnostic check functions ---

func TestCheckGoRuntime(t *testing.T) {
	check := checkGoRuntime(false)
	if check.Name != "Go Runtime" {
		t.Errorf("check.Name = %q, want 'Go Runtime'", check.Name)
	}
	if check.Status != CheckOK {
		t.Errorf("check.Status = %q, want %q", check.Status, CheckOK)
	}
	if check.Message == "" {
		t.Error("check.Message should not be empty")
	}
}

func TestCheckGoRuntime_Verbose(t *testing.T) {
	check := checkGoRuntime(true)
	if check.Detail == "" {
		t.Error("check.Detail should not be empty in verbose mode")
	}
	if !strings.Contains(check.Detail, "GOPATH") {
		t.Error("verbose detail should contain GOPATH")
	}
}

func TestCheckGit(t *testing.T) {
	check := checkGit(false)
	if check.Name != "Git" {
		t.Errorf("check.Name = %q, want 'Git'", check.Name)
	}
	// Git should be available in test environments
	if check.Status != CheckOK {
		t.Skipf("git not available: %s", check.Message)
	}
	if !strings.Contains(check.Message, "git version") {
		t.Errorf("check.Message should contain 'git version', got %q", check.Message)
	}
}

func TestCheckMoAIConfig_Missing(t *testing.T) {
	// Use a temp directory without .moai/
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if chErr := os.Chdir(origDir); chErr != nil {
			t.Logf("failed to restore working directory: %v", chErr)
		}
	}()

	check := checkMoAIConfig(false)
	if check.Status != CheckWarn {
		t.Errorf("check.Status = %q, want %q for missing .moai/", check.Status, CheckWarn)
	}
}

func TestCheckMoAIConfig_Present(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai", "config", "sections"), 0o755); err != nil {
		t.Fatal(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if chErr := os.Chdir(origDir); chErr != nil {
			t.Logf("failed to restore working directory: %v", chErr)
		}
	}()

	check := checkMoAIConfig(false)
	if check.Status != CheckOK {
		t.Errorf("check.Status = %q, want %q for present .moai/", check.Status, CheckOK)
	}
}

func TestCheckClaudeConfig_Missing(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if chErr := os.Chdir(origDir); chErr != nil {
			t.Logf("failed to restore working directory: %v", chErr)
		}
	}()

	check := checkClaudeConfig(false)
	if check.Status != CheckWarn {
		t.Errorf("check.Status = %q, want %q for missing .claude/", check.Status, CheckWarn)
	}
}

func TestCheckMoAIVersion(t *testing.T) {
	check := checkMoAIVersion(false)
	if check.Name != "MoAI Version" {
		t.Errorf("check.Name = %q, want 'MoAI Version'", check.Name)
	}
	if check.Status != CheckOK {
		t.Errorf("check.Status = %q, want %q", check.Status, CheckOK)
	}
	if !strings.Contains(check.Message, "moai-adk") {
		t.Errorf("check.Message should contain 'moai-adk', got %q", check.Message)
	}
}

func TestStatusIcon(t *testing.T) {
	tests := []struct {
		status   CheckStatus
		contains string
	}{
		{CheckOK, "\u2713"},       // ✓
		{CheckWarn, "\u26A0"},     // ⚠
		{CheckFail, "\u2717"},     // ✗
		{CheckStatus("unknown"), "?"},
	}
	for _, tt := range tests {
		got := statusIcon(tt.status)
		if !strings.Contains(got, tt.contains) {
			t.Errorf("statusIcon(%q) = %q, want string containing %q", tt.status, got, tt.contains)
		}
	}
}

func TestRunDiagnosticChecks_All(t *testing.T) {
	checks := runDiagnosticChecks(false, "")
	if len(checks) < 5 {
		t.Errorf("expected at least 5 checks, got %d", len(checks))
	}
}

func TestRunDiagnosticChecks_Filtered(t *testing.T) {
	checks := runDiagnosticChecks(false, "Go Runtime")
	if len(checks) != 1 {
		t.Errorf("expected 1 check when filtered, got %d", len(checks))
	}
	if len(checks) > 0 && checks[0].Name != "Go Runtime" {
		t.Errorf("filtered check name = %q, want 'Go Runtime'", checks[0].Name)
	}
}

func TestExportDiagnostics(t *testing.T) {
	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "diagnostics.json")

	checks := []DiagnosticCheck{
		{Name: "Test", Status: CheckOK, Message: "passed"},
	}

	if err := exportDiagnostics(exportPath, checks); err != nil {
		t.Fatalf("exportDiagnostics error: %v", err)
	}

	data, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("read exported file: %v", err)
	}

	var loaded []DiagnosticCheck
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("unmarshal exported JSON: %v", err)
	}

	if len(loaded) != 1 {
		t.Fatalf("expected 1 check, got %d", len(loaded))
	}
	if loaded[0].Name != "Test" {
		t.Errorf("loaded[0].Name = %q, want 'Test'", loaded[0].Name)
	}
}
