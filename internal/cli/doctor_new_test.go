package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// --- Tests for GitInstallHint (additional coverage) ---

func TestGitInstallHint_OSSpecific(t *testing.T) {
	hint := GitInstallHint()
	switch runtime.GOOS {
	case "darwin":
		if !strings.Contains(hint, "xcode-select") && !strings.Contains(hint, "brew") {
			t.Errorf("darwin hint should mention xcode-select or brew, got %q", hint)
		}
	case "windows":
		if !strings.Contains(hint, "winget") && !strings.Contains(hint, "git-scm.com") {
			t.Errorf("windows hint should mention winget or git-scm.com, got %q", hint)
		}
	default:
		if !strings.Contains(hint, "apt") && !strings.Contains(hint, "yum") {
			t.Errorf("linux hint should mention apt or yum, got %q", hint)
		}
	}
}

func TestGitInstallHint_NotEmpty(t *testing.T) {
	hint := GitInstallHint()
	if hint == "" {
		t.Error("GitInstallHint() should never return empty string")
	}
}

func TestGitInstallHint_ContainsInstallKeyword(t *testing.T) {
	hint := GitInstallHint()
	if !strings.Contains(hint, "Install") {
		t.Errorf("GitInstallHint() = %q, should contain 'Install'", hint)
	}
}

// --- Tests for checkGit ---

func TestCheckGit_Name(t *testing.T) {
	check := checkGit(false)
	if check.Name != "Git" {
		t.Errorf("checkGit Name = %q, want %q", check.Name, "Git")
	}
}

func TestCheckGit_VerboseShowsPath(t *testing.T) {
	check := checkGit(true)
	if check.Status == CheckOK && check.Detail == "" {
		t.Error("verbose checkGit should have Detail when git is available")
	}
	if check.Status == CheckOK && !strings.Contains(check.Detail, "path:") {
		t.Errorf("verbose Detail should contain 'path:', got %q", check.Detail)
	}
}

func TestCheckGit_NonVerboseNoPath(t *testing.T) {
	check := checkGit(false)
	if check.Status == CheckOK && check.Detail != "" {
		t.Errorf("non-verbose checkGit should not have Detail when OK, got %q", check.Detail)
	}
}

func TestCheckGit_FailIncludesInstallHint(t *testing.T) {
	check := checkGit(false)
	if check.Status == CheckFail {
		if check.Detail == "" {
			t.Error("failed checkGit should include install hint in Detail")
		}
		if !strings.Contains(check.Detail, "Install git") {
			t.Errorf("Detail = %q, should contain 'Install git'", check.Detail)
		}
	}
}

func TestCheckGit_StatusIsValid(t *testing.T) {
	check := checkGit(false)
	validStatuses := map[CheckStatus]bool{
		CheckOK:   true,
		CheckWarn: true,
		CheckFail: true,
	}
	if !validStatuses[check.Status] {
		t.Errorf("checkGit returned invalid status %q", check.Status)
	}
}

func TestCheckGit_MessageNotEmpty(t *testing.T) {
	check := checkGit(false)
	if check.Message == "" {
		t.Error("checkGit Message should not be empty")
	}
}

// --- Tests for exportDiagnostics ---

func TestExportDiagnostics_JSONFormat(t *testing.T) {
	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "diag.json")

	checks := []DiagnosticCheck{
		{Name: "Check1", Status: CheckOK, Message: "all good"},
		{Name: "Check2", Status: CheckWarn, Message: "warning", Detail: "detail info"},
		{Name: "Check3", Status: CheckFail, Message: "failed"},
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

	if len(loaded) != 3 {
		t.Fatalf("expected 3 checks, got %d", len(loaded))
	}

	// Verify each check round-trips correctly.
	for i, check := range loaded {
		if check.Name != checks[i].Name {
			t.Errorf("check[%d].Name = %q, want %q", i, check.Name, checks[i].Name)
		}
		if check.Status != checks[i].Status {
			t.Errorf("check[%d].Status = %q, want %q", i, check.Status, checks[i].Status)
		}
		if check.Message != checks[i].Message {
			t.Errorf("check[%d].Message = %q, want %q", i, check.Message, checks[i].Message)
		}
		if check.Detail != checks[i].Detail {
			t.Errorf("check[%d].Detail = %q, want %q", i, check.Detail, checks[i].Detail)
		}
	}
}

func TestExportDiagnostics_EmptyChecks(t *testing.T) {
	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "empty.json")

	if err := exportDiagnostics(exportPath, []DiagnosticCheck{}); err != nil {
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

	if len(loaded) != 0 {
		t.Errorf("expected 0 checks, got %d", len(loaded))
	}
}

func TestExportDiagnostics_NilChecks(t *testing.T) {
	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "nil.json")

	if err := exportDiagnostics(exportPath, nil); err != nil {
		t.Fatalf("exportDiagnostics error: %v", err)
	}

	data, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("read exported file: %v", err)
	}

	// nil slice marshals to "null" in JSON.
	trimmed := strings.TrimSpace(string(data))
	if trimmed != "null" {
		t.Errorf("nil checks should marshal to 'null', got %q", trimmed)
	}
}

func TestExportDiagnostics_InvalidPath(t *testing.T) {
	// Writing to a path inside a nonexistent directory should fail.
	err := exportDiagnostics("/nonexistent/dir/diag.json", []DiagnosticCheck{
		{Name: "Test", Status: CheckOK, Message: "ok"},
	})
	if err == nil {
		t.Fatal("exportDiagnostics should error for invalid path")
	}
}

func TestExportDiagnostics_IndentedJSON(t *testing.T) {
	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "pretty.json")

	checks := []DiagnosticCheck{
		{Name: "Test", Status: CheckOK, Message: "ok"},
	}

	if err := exportDiagnostics(exportPath, checks); err != nil {
		t.Fatalf("exportDiagnostics error: %v", err)
	}

	data, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("read exported file: %v", err)
	}

	// Verify it is indented (contains newlines and spaces).
	content := string(data)
	if !strings.Contains(content, "\n") {
		t.Error("exported JSON should be indented (contain newlines)")
	}
	if !strings.Contains(content, "  ") {
		t.Error("exported JSON should be indented (contain spaces)")
	}
}

func TestExportDiagnostics_DetailOmittedWhenEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "omit.json")

	checks := []DiagnosticCheck{
		{Name: "NoDetail", Status: CheckOK, Message: "ok"},
	}

	if err := exportDiagnostics(exportPath, checks); err != nil {
		t.Fatalf("exportDiagnostics error: %v", err)
	}

	data, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("read exported file: %v", err)
	}

	// The "detail" field should be omitted (omitempty tag).
	if strings.Contains(string(data), `"detail"`) {
		t.Error("empty Detail should be omitted from JSON output (omitempty)")
	}
}

// --- Tests for DiagnosticCheck struct JSON marshaling ---

func TestDiagnosticCheck_JSONMarshal(t *testing.T) {
	check := DiagnosticCheck{
		Name:    "TestCheck",
		Status:  CheckWarn,
		Message: "test message",
		Detail:  "test detail",
	}

	data, err := json.Marshal(check)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	if result["name"] != "TestCheck" {
		t.Errorf("name = %v, want 'TestCheck'", result["name"])
	}
	if result["status"] != "warn" {
		t.Errorf("status = %v, want 'warn'", result["status"])
	}
	if result["message"] != "test message" {
		t.Errorf("message = %v, want 'test message'", result["message"])
	}
	if result["detail"] != "test detail" {
		t.Errorf("detail = %v, want 'test detail'", result["detail"])
	}
}

// --- Tests for CheckStatus constants ---

func TestCheckStatus_Values(t *testing.T) {
	tests := []struct {
		name   string
		status CheckStatus
		want   string
	}{
		{"ok", CheckOK, "ok"},
		{"warn", CheckWarn, "warn"},
		{"fail", CheckFail, "fail"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("CheckStatus %s = %q, want %q", tt.name, tt.status, tt.want)
			}
		})
	}
}

// --- Tests for runDiagnosticChecks ---

func TestRunDiagnosticChecks_FilterNonExistent(t *testing.T) {
	checks := runDiagnosticChecks(false, "NonExistentCheck")
	if len(checks) != 0 {
		t.Errorf("expected 0 checks for non-existent filter, got %d", len(checks))
	}
}

func TestRunDiagnosticChecks_AllChecksHaveNames(t *testing.T) {
	checks := runDiagnosticChecks(false, "")
	for i, check := range checks {
		if check.Name == "" {
			t.Errorf("check[%d] has empty Name", i)
		}
	}
}

func TestRunDiagnosticChecks_AllChecksHaveValidStatus(t *testing.T) {
	checks := runDiagnosticChecks(false, "")
	validStatuses := map[CheckStatus]bool{
		CheckOK:   true,
		CheckWarn: true,
		CheckFail: true,
	}
	for i, check := range checks {
		if !validStatuses[check.Status] {
			t.Errorf("check[%d] (%s) has invalid status %q", i, check.Name, check.Status)
		}
	}
}

func TestRunDiagnosticChecks_VerboseAddsDetail(t *testing.T) {
	checks := runDiagnosticChecks(true, "Go Runtime")
	if len(checks) != 1 {
		t.Fatalf("expected 1 check, got %d", len(checks))
	}
	if checks[0].Detail == "" {
		t.Error("verbose Go Runtime check should have Detail")
	}
}
