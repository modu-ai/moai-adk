package doctor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// --- NewDoctor ---

func TestNewDoctor(t *testing.T) {
	d := NewDoctor("/tmp/testdir")
	if d == nil {
		t.Fatal("NewDoctor returned nil")
	}
	if d.projectDir != "/tmp/testdir" {
		t.Errorf("projectDir = %q, want %q", d.projectDir, "/tmp/testdir")
	}
	if d.results == nil {
		t.Error("results slice is nil")
	}
	if len(d.results) != 0 {
		t.Errorf("results length = %d, want 0", len(d.results))
	}
}

// --- CheckResult ---

func TestCheckResultFields(t *testing.T) {
	cr := &CheckResult{
		Name:    "test-check",
		Status:  "success",
		Message: "all good",
		Value:   "/some/path",
	}
	if cr.Name != "test-check" {
		t.Errorf("Name = %q, want %q", cr.Name, "test-check")
	}
	if cr.Status != "success" {
		t.Errorf("Status = %q, want %q", cr.Status, "success")
	}
	if cr.Message != "all good" {
		t.Errorf("Message = %q, want %q", cr.Message, "all good")
	}
	if cr.Value != "/some/path" {
		t.Errorf("Value = %q, want %q", cr.Value, "/some/path")
	}
}

// --- checkSettingsJSON ---

func TestCheckSettingsJSON_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	d := NewDoctor(tmpDir)

	d.checkSettingsJSON()

	if len(d.results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(d.results))
	}
	r := d.results[0]
	if r.Name != "settings.json" {
		t.Errorf("Name = %q, want %q", r.Name, "settings.json")
	}
	if r.Status != "warning" {
		t.Errorf("Status = %q, want %q", r.Status, "warning")
	}
	if r.Message != "not found (run 'moai init')" {
		t.Errorf("Message = %q", r.Message)
	}
}

func TestCheckSettingsJSON_ValidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	settingsData := map[string]interface{}{
		"hooks": map[string]interface{}{
			"SessionStart": map[string]string{"type": "command", "command": "echo hello"},
		},
	}
	data, _ := json.MarshalIndent(settingsData, "", "  ")
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	d := NewDoctor(tmpDir)
	d.checkSettingsJSON()

	if len(d.results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(d.results))
	}
	r := d.results[0]
	if r.Status != "success" {
		t.Errorf("Status = %q, want %q", r.Status, "success")
	}
	if r.Message != "valid" {
		t.Errorf("Message = %q, want %q", r.Message, "valid")
	}
}

func TestCheckSettingsJSON_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte("{invalid json"), 0644); err != nil {
		t.Fatal(err)
	}

	d := NewDoctor(tmpDir)
	d.checkSettingsJSON()

	if len(d.results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(d.results))
	}
	r := d.results[0]
	if r.Status != "error" {
		t.Errorf("Status = %q, want %q", r.Status, "error")
	}
	if r.Message != "invalid JSON" {
		t.Errorf("Message = %q, want %q", r.Message, "invalid JSON")
	}
}

// --- checkMoaiConfig ---

func TestCheckMoaiConfig_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	d := NewDoctor(tmpDir)

	d.checkMoaiConfig()

	if len(d.results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(d.results))
	}
	r := d.results[0]
	if r.Status != "warning" {
		t.Errorf("Status = %q, want %q", r.Status, "warning")
	}
}

func TestCheckMoaiConfig_Incomplete(t *testing.T) {
	tmpDir := t.TempDir()
	// Create config dir but not sections
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai", "config"), 0755); err != nil {
		t.Fatal(err)
	}

	d := NewDoctor(tmpDir)
	d.checkMoaiConfig()

	if len(d.results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(d.results))
	}
	r := d.results[0]
	if r.Status != "warning" {
		t.Errorf("Status = %q, want %q", r.Status, "warning")
	}
	if r.Message != "incomplete" {
		t.Errorf("Message = %q, want %q", r.Message, "incomplete")
	}
}

func TestCheckMoaiConfig_Complete(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai", "config", "sections"), 0755); err != nil {
		t.Fatal(err)
	}

	d := NewDoctor(tmpDir)
	d.checkMoaiConfig()

	if len(d.results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(d.results))
	}
	r := d.results[0]
	if r.Status != "success" {
		t.Errorf("Status = %q, want %q", r.Status, "success")
	}
	if r.Message != "complete" {
		t.Errorf("Message = %q, want %q", r.Message, "complete")
	}
}

// --- checkLanguageConfig ---

func TestCheckLanguageConfig_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	d := NewDoctor(tmpDir)

	d.checkLanguageConfig()

	if len(d.results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(d.results))
	}
	r := d.results[0]
	if r.Status != "warning" {
		t.Errorf("Status = %q, want %q", r.Status, "warning")
	}
}

func TestCheckLanguageConfig_ValidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	langDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(langDir, 0755); err != nil {
		t.Fatal(err)
	}

	langConfig := map[string]interface{}{
		"language": map[string]string{
			"conversation_language": "en",
		},
	}
	data, _ := yaml.Marshal(langConfig)
	if err := os.WriteFile(filepath.Join(langDir, "language.yaml"), data, 0644); err != nil {
		t.Fatal(err)
	}

	d := NewDoctor(tmpDir)
	d.checkLanguageConfig()

	if len(d.results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(d.results))
	}
	r := d.results[0]
	if r.Status != "success" {
		t.Errorf("Status = %q, want %q", r.Status, "success")
	}
}

func TestCheckLanguageConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	langDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(langDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(langDir, "language.yaml"), []byte(":\t\tinvalid\nyaml: [unclosed"), 0644); err != nil {
		t.Fatal(err)
	}

	d := NewDoctor(tmpDir)
	d.checkLanguageConfig()

	if len(d.results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(d.results))
	}
	r := d.results[0]
	if r.Status != "error" {
		t.Errorf("Status = %q, want %q", r.Status, "error")
	}
}

// --- checkUserConfig ---

func TestCheckUserConfig_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	d := NewDoctor(tmpDir)

	d.checkUserConfig()

	if len(d.results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(d.results))
	}
	r := d.results[0]
	if r.Status != "warning" {
		t.Errorf("Status = %q, want %q", r.Status, "warning")
	}
}

func TestCheckUserConfig_ValidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	userDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(userDir, 0755); err != nil {
		t.Fatal(err)
	}

	userConfig := map[string]interface{}{
		"user": map[string]string{
			"name": "TestUser",
		},
	}
	data, _ := yaml.Marshal(userConfig)
	if err := os.WriteFile(filepath.Join(userDir, "user.yaml"), data, 0644); err != nil {
		t.Fatal(err)
	}

	d := NewDoctor(tmpDir)
	d.checkUserConfig()

	if len(d.results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(d.results))
	}
	r := d.results[0]
	if r.Status != "success" {
		t.Errorf("Status = %q, want %q", r.Status, "success")
	}
}

func TestCheckUserConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	userDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(userDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(userDir, "user.yaml"), []byte(":\t\tinvalid\nyaml: [unclosed"), 0644); err != nil {
		t.Fatal(err)
	}

	d := NewDoctor(tmpDir)
	d.checkUserConfig()

	if len(d.results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(d.results))
	}
	r := d.results[0]
	if r.Status != "error" {
		t.Errorf("Status = %q, want %q", r.Status, "error")
	}
}

// --- checkHookRegistration ---

func TestCheckHookRegistration_NoSettingsFile(t *testing.T) {
	tmpDir := t.TempDir()
	d := NewDoctor(tmpDir)

	d.checkHookRegistration()

	if len(d.results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(d.results))
	}
	r := d.results[0]
	if r.Status != "warning" {
		t.Errorf("Status = %q, want %q", r.Status, "warning")
	}
}

func TestCheckHookRegistration_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte("{bad json"), 0644); err != nil {
		t.Fatal(err)
	}

	d := NewDoctor(tmpDir)
	d.checkHookRegistration()

	if len(d.results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(d.results))
	}
	r := d.results[0]
	if r.Status != "warning" {
		t.Errorf("Status = %q, want %q", r.Status, "warning")
	}
}

func TestCheckHookRegistration_NoHooksKey(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	data, _ := json.Marshal(map[string]interface{}{"someKey": "value"})
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	d := NewDoctor(tmpDir)
	d.checkHookRegistration()

	if len(d.results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(d.results))
	}
	r := d.results[0]
	if r.Status != "error" {
		t.Errorf("Status = %q, want %q", r.Status, "error")
	}
	if r.Message != "no hooks registered" {
		t.Errorf("Message = %q, want %q", r.Message, "no hooks registered")
	}
}

func TestCheckHookRegistration_MissingHooks(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	settings := map[string]interface{}{
		"hooks": map[string]interface{}{
			"SessionStart": map[string]string{"type": "command"},
			// Missing PreToolUse and PostToolUse
		},
	}
	data, _ := json.Marshal(settings)
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	d := NewDoctor(tmpDir)
	d.checkHookRegistration()

	if len(d.results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(d.results))
	}
	r := d.results[0]
	if r.Status != "error" {
		t.Errorf("Status = %q, want %q", r.Status, "error")
	}
}

func TestCheckHookRegistration_AllPresent(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	settings := map[string]interface{}{
		"hooks": map[string]interface{}{
			"SessionStart": map[string]string{"type": "command"},
			"PreToolUse":   map[string]string{"type": "command"},
			"PostToolUse":  map[string]string{"type": "command"},
		},
	}
	data, _ := json.Marshal(settings)
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	d := NewDoctor(tmpDir)
	d.checkHookRegistration()

	if len(d.results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(d.results))
	}
	r := d.results[0]
	if r.Status != "success" {
		t.Errorf("Status = %q, want %q", r.Status, "success")
	}
	if r.Message != "all registered" {
		t.Errorf("Message = %q, want %q", r.Message, "all registered")
	}
}

// --- getExitCode ---

func TestGetExitCode_NoIssues(t *testing.T) {
	d := NewDoctor("/tmp")
	d.results = []*CheckResult{
		{Name: "test1", Status: "success", Message: "ok"},
		{Name: "test2", Status: "success", Message: "ok"},
	}

	err := d.getExitCode()
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestGetExitCode_WithErrors(t *testing.T) {
	d := NewDoctor("/tmp")
	d.results = []*CheckResult{
		{Name: "test1", Status: "success", Message: "ok"},
		{Name: "test2", Status: "error", Message: "failed"},
		{Name: "test3", Status: "error", Message: "also failed"},
	}

	err := d.getExitCode()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "2 critical failures found" {
		t.Errorf("error = %q", err.Error())
	}
}

func TestGetExitCode_WithWarnings(t *testing.T) {
	d := NewDoctor("/tmp")
	d.results = []*CheckResult{
		{Name: "test1", Status: "success", Message: "ok"},
		{Name: "test2", Status: "warning", Message: "optional"},
	}

	err := d.getExitCode()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "1 optional tools missing" {
		t.Errorf("error = %q", err.Error())
	}
}

func TestGetExitCode_ErrorsTakePrecedence(t *testing.T) {
	d := NewDoctor("/tmp")
	d.results = []*CheckResult{
		{Name: "test1", Status: "warning", Message: "optional"},
		{Name: "test2", Status: "error", Message: "critical"},
	}

	err := d.getExitCode()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Errors take precedence over warnings
	if err.Error() != "1 critical failures found" {
		t.Errorf("error = %q", err.Error())
	}
}

// --- checkTool ---

func TestCheckTool_RequiredNotFound(t *testing.T) {
	d := NewDoctor("/tmp")
	// Use a command that certainly does not exist
	d.checkTool("nonexistent-tool", "this-command-does-not-exist-12345", []string{"--version"}, true)

	if len(d.results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(d.results))
	}
	r := d.results[0]
	if r.Status != "error" {
		t.Errorf("Status = %q, want %q for required missing tool", r.Status, "error")
	}
	if r.Message != "not found" {
		t.Errorf("Message = %q, want %q", r.Message, "not found")
	}
}

func TestCheckTool_OptionalNotFound(t *testing.T) {
	d := NewDoctor("/tmp")
	d.checkTool("nonexistent-tool", "this-command-does-not-exist-12345", []string{"--version"}, false)

	if len(d.results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(d.results))
	}
	r := d.results[0]
	if r.Status != "warning" {
		t.Errorf("Status = %q, want %q for optional missing tool", r.Status, "warning")
	}
}

// --- checkBinaryVersion ---

func TestCheckBinaryVersion(t *testing.T) {
	d := NewDoctor("/tmp")
	d.checkBinaryVersion()

	if len(d.results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(d.results))
	}
	r := d.results[0]
	if r.Name != "Binary Version" {
		t.Errorf("Name = %q, want %q", r.Name, "Binary Version")
	}
	// Should always succeed when running tests
	if r.Status != "success" {
		t.Errorf("Status = %q, want %q", r.Status, "success")
	}
}

// --- checkConfigFiles ---

func TestCheckConfigFiles_AllMissing(t *testing.T) {
	tmpDir := t.TempDir()
	d := NewDoctor(tmpDir)

	d.checkConfigFiles()

	// Should have 4 results: settings.json, config structure, language.yaml, user.yaml
	if len(d.results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(d.results))
	}

	for _, r := range d.results {
		if r.Status == "error" {
			t.Errorf("check %q should be warning, not error (for missing files)", r.Name)
		}
	}
}

// --- printResults ---

func TestPrintResults_DoesNotPanic(t *testing.T) {
	d := NewDoctor("/tmp")
	d.results = []*CheckResult{
		{Name: "success-check", Status: "success", Message: "all good"},
		{Name: "warning-check", Status: "warning", Message: "optional", Value: "/some/path"},
		{Name: "error-check", Status: "error", Message: "critical failure", Value: "/error/path"},
	}

	// Should not panic
	d.printResults()
}

func TestPrintResults_AllPassing(t *testing.T) {
	d := NewDoctor("/tmp")
	d.results = []*CheckResult{
		{Name: "check1", Status: "success", Message: "ok"},
		{Name: "check2", Status: "success", Message: "ok"},
	}

	// Should not panic and print "All checks passed!"
	d.printResults()
}
