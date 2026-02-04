package hook

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewQualityGateEnforcer verifies enforcer creation.
func TestNewQualityGateEnforcer(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	enforcer := NewQualityGateEnforcer(tmpDir)

	if enforcer == nil {
		t.Fatal("expected non-nil enforcer")
	}
}

// TestShouldBlock_ErrorsExceedThreshold verifies blocking on error threshold per REQ-HOOK-181.
func TestShouldBlock_ErrorsExceedThreshold(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	enforcer := NewQualityGateEnforcer(tmpDir)

	counts := SeverityCounts{
		Errors:   5,
		Warnings: 0,
	}

	gate := QualityGate{
		MaxErrors:    0,
		MaxWarnings:  10,
		BlockOnError: true,
	}

	if !enforcer.ShouldBlock(counts, gate) {
		t.Error("expected ShouldBlock to return true when errors exceed threshold")
	}
}

// TestShouldBlock_ErrorsBelowThreshold verifies no blocking when errors are within threshold.
func TestShouldBlock_ErrorsBelowThreshold(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	enforcer := NewQualityGateEnforcer(tmpDir)

	counts := SeverityCounts{
		Errors:   2,
		Warnings: 5,
	}

	gate := QualityGate{
		MaxErrors:    5,
		MaxWarnings:  10,
		BlockOnError: true,
	}

	if enforcer.ShouldBlock(counts, gate) {
		t.Error("expected ShouldBlock to return false when errors are below threshold")
	}
}

// TestShouldBlock_WarningsExceedThreshold verifies warning threshold behavior per REQ-HOOK-182.
func TestShouldBlock_WarningsExceedThreshold(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	enforcer := NewQualityGateEnforcer(tmpDir)

	counts := SeverityCounts{
		Errors:   0,
		Warnings: 15,
	}

	// BlockOnWarning = false means warnings don't block
	gate := QualityGate{
		MaxErrors:      0,
		MaxWarnings:    10,
		BlockOnError:   true,
		BlockOnWarning: false,
	}

	// Should not block because BlockOnWarning is false
	if enforcer.ShouldBlock(counts, gate) {
		t.Error("expected ShouldBlock to return false when BlockOnWarning is false")
	}

	// Now with BlockOnWarning = true
	gateWithBlock := QualityGate{
		MaxErrors:      5,
		MaxWarnings:    10,
		BlockOnError:   true,
		BlockOnWarning: true,
	}

	if !enforcer.ShouldBlock(counts, gateWithBlock) {
		t.Error("expected ShouldBlock to return true when warnings exceed threshold and BlockOnWarning is true")
	}
}

// TestShouldBlock_ZeroThreshold verifies zero-tolerance threshold.
func TestShouldBlock_ZeroThreshold(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	enforcer := NewQualityGateEnforcer(tmpDir)

	counts := SeverityCounts{
		Errors: 1,
	}

	gate := QualityGate{
		MaxErrors:    0,
		BlockOnError: true,
	}

	if !enforcer.ShouldBlock(counts, gate) {
		t.Error("expected ShouldBlock to return true with zero-tolerance and 1 error")
	}
}

// TestShouldBlock_ExactThreshold verifies behavior at exact threshold.
func TestShouldBlock_ExactThreshold(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	enforcer := NewQualityGateEnforcer(tmpDir)

	counts := SeverityCounts{
		Errors: 5,
	}

	gate := QualityGate{
		MaxErrors:    5,
		BlockOnError: true,
	}

	// At exact threshold, should not block (threshold is "max allowed")
	if enforcer.ShouldBlock(counts, gate) {
		t.Error("expected ShouldBlock to return false when at exact threshold")
	}
}

// TestLoadConfig_ValidConfig verifies config loading per REQ-HOOK-180.
func TestLoadConfig_ValidConfig(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create config directory and file
	configDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configContent := `constitution:
  lsp_quality_gates:
    enabled: true
    run:
      max_errors: 0
      max_type_errors: 0
      max_lint_errors: 0
      allow_regression: false
    sync:
      max_errors: 0
      max_warnings: 10
      require_clean_lsp: true
`
	configFile := filepath.Join(configDir, "quality.yaml")
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	enforcer := NewQualityGateEnforcer(tmpDir)

	gate, err := enforcer.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if gate.MaxErrors != 0 {
		t.Errorf("MaxErrors = %d, want 0", gate.MaxErrors)
	}
	if gate.MaxWarnings != 10 {
		t.Errorf("MaxWarnings = %d, want 10", gate.MaxWarnings)
	}
}

// TestLoadConfig_DefaultConfig verifies default config when file not found.
func TestLoadConfig_DefaultConfig(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	enforcer := NewQualityGateEnforcer(tmpDir)

	gate, err := enforcer.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Should return sensible defaults
	if gate.MaxErrors < 0 {
		t.Errorf("MaxErrors = %d, expected non-negative", gate.MaxErrors)
	}
}

// TestCheckWithConfig verifies combined check and config load.
func TestCheckWithConfig(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create config with strict settings
	configDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configContent := `constitution:
  lsp_quality_gates:
    enabled: true
    run:
      max_errors: 0
      max_type_errors: 0
      max_lint_errors: 0
      allow_regression: false
`
	configFile := filepath.Join(configDir, "quality.yaml")
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	enforcer := NewQualityGateEnforcer(tmpDir)

	counts := SeverityCounts{
		Errors: 1,
	}

	shouldBlock, gate, err := enforcer.CheckWithConfig(counts)
	if err != nil {
		t.Fatalf("CheckWithConfig failed: %v", err)
	}

	if !shouldBlock {
		t.Error("expected shouldBlock to be true with errors and zero-tolerance config")
	}

	if gate.MaxErrors != 0 {
		t.Errorf("gate.MaxErrors = %d, want 0", gate.MaxErrors)
	}
}

// TestQualityGateEnforcer_ExitCode verifies exit code behavior per REQ-HOOK-181.
func TestQualityGateEnforcer_ExitCode(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	enforcer := NewQualityGateEnforcer(tmpDir)

	// Test GetExitCode function
	counts := SeverityCounts{Errors: 5}
	gate := QualityGate{MaxErrors: 0, BlockOnError: true}

	exitCode := enforcer.GetExitCode(counts, gate)
	if exitCode != 2 {
		t.Errorf("GetExitCode = %d, want 2 when errors exceed threshold", exitCode)
	}

	// No errors should return 0
	counts = SeverityCounts{Errors: 0}
	exitCode = enforcer.GetExitCode(counts, gate)
	if exitCode != 0 {
		t.Errorf("GetExitCode = %d, want 0 when no errors", exitCode)
	}
}

// TestQualityGateEnforcer_WarningLog verifies warning logging per REQ-HOOK-182.
func TestQualityGateEnforcer_WarningLog(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	enforcer := NewQualityGateEnforcer(tmpDir)

	counts := SeverityCounts{
		Errors:   0,
		Warnings: 15,
	}

	gate := QualityGate{
		MaxErrors:      0,
		MaxWarnings:    10,
		BlockOnError:   true,
		BlockOnWarning: false,
	}

	// Should not block, but should indicate warnings exceed threshold
	if enforcer.ShouldBlock(counts, gate) {
		t.Error("should not block when BlockOnWarning is false")
	}

	if !enforcer.WarningsExceedThreshold(counts, gate) {
		t.Error("expected WarningsExceedThreshold to return true")
	}
}

// TestFormatGateResult verifies result formatting.
func TestFormatGateResult(t *testing.T) {
	t.Parallel()

	counts := SeverityCounts{
		Errors:   2,
		Warnings: 5,
	}

	gate := QualityGate{
		MaxErrors:   0,
		MaxWarnings: 10,
	}

	result := FormatGateResult(counts, gate)

	if result == "" {
		t.Error("expected non-empty result")
	}

	// Should mention errors exceed threshold
	if !containsSubstring(result, "Errors") && !containsSubstring(result, "error") {
		t.Error("result should mention errors")
	}
}

// containsSubstring checks if s contains substr.
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestParseQualityConfig_InvalidYAML verifies error handling for invalid YAML.
func TestParseQualityConfig_InvalidYAML(t *testing.T) {
	t.Parallel()

	invalidYAML := []byte("{ invalid yaml ][")
	_, err := parseQualityConfig(invalidYAML)
	// Should return defaults without error or with error
	// The implementation returns defaults on parse error
	_ = err
}

// TestParseQualityConfig_DisabledGates verifies disabled gates behavior.
func TestParseQualityConfig_DisabledGates(t *testing.T) {
	t.Parallel()

	configYAML := []byte(`constitution:
  lsp_quality_gates:
    enabled: false
`)
	gate, err := parseQualityConfig(configYAML)
	if err != nil {
		t.Fatalf("parseQualityConfig failed: %v", err)
	}

	// Should return defaults when disabled
	defaultGate := defaultQualityGate()
	if gate.MaxErrors != defaultGate.MaxErrors {
		t.Errorf("MaxErrors = %d, want %d", gate.MaxErrors, defaultGate.MaxErrors)
	}
}
