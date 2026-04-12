package hook

import (
	"os"
	"path/filepath"
	"testing"
)

// TestPhaseTypeConstants verifies PhaseType constant values (REQ-QG-001).
func TestPhaseTypeConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		phase PhaseType
		want  string
	}{
		{"plan phase", PhasePlan, "plan"},
		{"run phase", PhaseRun, "run"},
		{"sync phase", PhaseSync, "sync"},
		{"auto phase", PhaseAuto, "auto"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if string(tt.phase) != tt.want {
				t.Errorf("PhaseType(%q) string = %q, want %q", tt.name, string(tt.phase), tt.want)
			}
		})
	}
}

// TestQualityGatePhaseField verifies QualityGate accepts a Phase field (REQ-QG-001).
func TestQualityGatePhaseField(t *testing.T) {
	t.Parallel()

	gate := QualityGate{
		Phase:       PhaseRun,
		MaxErrors:   0,
		MaxWarnings: 10,
		BlockOnError: true,
	}

	if gate.Phase != PhaseRun {
		t.Errorf("Phase = %q, want %q", gate.Phase, PhaseRun)
	}
}

// TestPhaseAwareEnforcer_PlanPhase verifies plan phase captures baseline when require_baseline is true (REQ-QG-002).
func TestPhaseAwareEnforcer_PlanPhase(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupQualityConfig(t, tmpDir, `constitution:
  lsp_quality_gates:
    enabled: true
    plan:
      require_baseline: true
    run:
      max_errors: 0
      max_type_errors: 0
      max_lint_errors: 0
      allow_regression: false
    sync:
      max_errors: 0
      max_warnings: 10
      require_clean_lsp: true
`)

	enforcer := NewQualityGateEnforcer(tmpDir)
	result, err := enforcer.EnforcePhase(PhasePlan, SeverityCounts{Errors: 0, Warnings: 5}, nil)
	if err != nil {
		t.Fatalf("EnforcePhase(plan) unexpected error: %v", err)
	}

	// Plan phase with require_baseline: does not block on diagnostics, captures baseline
	if result.ShouldBlock {
		t.Error("plan phase should not block on diagnostics")
	}
	if !result.BaselineCaptured {
		t.Error("plan phase with require_baseline=true should set BaselineCaptured")
	}
}

// TestPhaseAwareEnforcer_PlanPhase_NoBaseline verifies plan phase without require_baseline (REQ-QG-002).
func TestPhaseAwareEnforcer_PlanPhase_NoBaseline(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupQualityConfig(t, tmpDir, `constitution:
  lsp_quality_gates:
    enabled: true
    plan:
      require_baseline: false
`)

	enforcer := NewQualityGateEnforcer(tmpDir)
	result, err := enforcer.EnforcePhase(PhasePlan, SeverityCounts{Errors: 0}, nil)
	if err != nil {
		t.Fatalf("EnforcePhase(plan) unexpected error: %v", err)
	}

	if result.ShouldBlock {
		t.Error("plan phase should not block")
	}
	if result.BaselineCaptured {
		t.Error("plan phase with require_baseline=false should not set BaselineCaptured")
	}
}

// TestPhaseAwareEnforcer_RunPhase_MaxErrors verifies run phase enforces max_errors (REQ-QG-003).
func TestPhaseAwareEnforcer_RunPhase_MaxErrors(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupQualityConfig(t, tmpDir, `constitution:
  lsp_quality_gates:
    enabled: true
    run:
      max_errors: 0
      max_type_errors: 0
      max_lint_errors: 0
      allow_regression: false
`)

	enforcer := NewQualityGateEnforcer(tmpDir)
	counts := SeverityCounts{Errors: 1}
	result, err := enforcer.EnforcePhase(PhaseRun, counts, nil)
	if err != nil {
		t.Fatalf("EnforcePhase(run) unexpected error: %v", err)
	}

	if !result.ShouldBlock {
		t.Error("run phase should block when errors exceed max_errors=0")
	}
}

// TestPhaseAwareEnforcer_RunPhase_TypeErrors verifies run phase enforces max_type_errors (REQ-QG-003).
func TestPhaseAwareEnforcer_RunPhase_TypeErrors(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupQualityConfig(t, tmpDir, `constitution:
  lsp_quality_gates:
    enabled: true
    run:
      max_errors: 5
      max_type_errors: 0
      max_lint_errors: 5
      allow_regression: false
`)

	enforcer := NewQualityGateEnforcer(tmpDir)
	counts := SeverityCounts{Errors: 0, TypeErrors: 1}
	result, err := enforcer.EnforcePhase(PhaseRun, counts, nil)
	if err != nil {
		t.Fatalf("EnforcePhase(run) unexpected error: %v", err)
	}

	if !result.ShouldBlock {
		t.Error("run phase should block when type_errors exceed max_type_errors=0")
	}
}

// TestPhaseAwareEnforcer_RunPhase_LintErrors verifies run phase enforces max_lint_errors (REQ-QG-003).
func TestPhaseAwareEnforcer_RunPhase_LintErrors(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupQualityConfig(t, tmpDir, `constitution:
  lsp_quality_gates:
    enabled: true
    run:
      max_errors: 5
      max_type_errors: 5
      max_lint_errors: 0
      allow_regression: false
`)

	enforcer := NewQualityGateEnforcer(tmpDir)
	counts := SeverityCounts{LintErrors: 1}
	result, err := enforcer.EnforcePhase(PhaseRun, counts, nil)
	if err != nil {
		t.Fatalf("EnforcePhase(run) unexpected error: %v", err)
	}

	if !result.ShouldBlock {
		t.Error("run phase should block when lint_errors exceed max_lint_errors=0")
	}
}

// TestPhaseAwareEnforcer_RunPhase_AllowRegression verifies allow_regression flag (REQ-QG-003, REQ-QG-007).
func TestPhaseAwareEnforcer_RunPhase_AllowRegression(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupQualityConfig(t, tmpDir, `constitution:
  lsp_quality_gates:
    enabled: true
    run:
      max_errors: 0
      max_type_errors: 0
      max_lint_errors: 0
      allow_regression: false
  lsp_integration:
    regression_detection:
      error_increase_threshold: 0
      warning_increase_threshold: 10
`)

	enforcer := NewQualityGateEnforcer(tmpDir)

	// Provide a baseline snapshot with 0 errors; current has 1 => regression
	baseline := &SeverityCounts{Errors: 0}
	counts := SeverityCounts{Errors: 1}

	result, err := enforcer.EnforcePhase(PhaseRun, counts, baseline)
	if err != nil {
		t.Fatalf("EnforcePhase(run) unexpected error: %v", err)
	}

	// With allow_regression=false and regression detected, should block
	if !result.ShouldBlock {
		t.Error("run phase should block when regression detected and allow_regression=false")
	}
	if !result.HasRegression {
		t.Error("result should report regression")
	}
}

// TestPhaseAwareEnforcer_RunPhase_AllowRegressionTrue verifies allow_regression=true does not block (REQ-QG-003).
func TestPhaseAwareEnforcer_RunPhase_AllowRegressionTrue(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupQualityConfig(t, tmpDir, `constitution:
  lsp_quality_gates:
    enabled: true
    run:
      max_errors: 5
      max_type_errors: 5
      max_lint_errors: 5
      allow_regression: true
  lsp_integration:
    regression_detection:
      error_increase_threshold: 0
      warning_increase_threshold: 10
`)

	enforcer := NewQualityGateEnforcer(tmpDir)
	baseline := &SeverityCounts{Errors: 0}
	counts := SeverityCounts{Errors: 1}

	result, err := enforcer.EnforcePhase(PhaseRun, counts, baseline)
	if err != nil {
		t.Fatalf("EnforcePhase(run) unexpected error: %v", err)
	}

	// allow_regression=true and errors within max_errors=5 => should not block
	if result.ShouldBlock {
		t.Error("run phase should not block when allow_regression=true and within thresholds")
	}
}

// TestPhaseAwareEnforcer_SyncPhase_MaxErrors verifies sync phase enforces max_errors (REQ-QG-004).
func TestPhaseAwareEnforcer_SyncPhase_MaxErrors(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupQualityConfig(t, tmpDir, `constitution:
  lsp_quality_gates:
    enabled: true
    sync:
      max_errors: 0
      max_warnings: 10
      require_clean_lsp: false
`)

	enforcer := NewQualityGateEnforcer(tmpDir)
	counts := SeverityCounts{Errors: 1}
	result, err := enforcer.EnforcePhase(PhaseSync, counts, nil)
	if err != nil {
		t.Fatalf("EnforcePhase(sync) unexpected error: %v", err)
	}

	if !result.ShouldBlock {
		t.Error("sync phase should block when errors exceed max_errors=0")
	}
}

// TestPhaseAwareEnforcer_SyncPhase_MaxWarnings verifies sync phase enforces max_warnings (REQ-QG-004).
func TestPhaseAwareEnforcer_SyncPhase_MaxWarnings(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupQualityConfig(t, tmpDir, `constitution:
  lsp_quality_gates:
    enabled: true
    sync:
      max_errors: 0
      max_warnings: 5
      require_clean_lsp: false
`)

	enforcer := NewQualityGateEnforcer(tmpDir)
	counts := SeverityCounts{Errors: 0, Warnings: 6}
	result, err := enforcer.EnforcePhase(PhaseSync, counts, nil)
	if err != nil {
		t.Fatalf("EnforcePhase(sync) unexpected error: %v", err)
	}

	if !result.ShouldBlock {
		t.Error("sync phase should block when warnings exceed max_warnings=5")
	}
}

// TestPhaseAwareEnforcer_SyncPhase_RequireCleanLSP verifies require_clean_lsp enforcement (REQ-QG-004).
func TestPhaseAwareEnforcer_SyncPhase_RequireCleanLSP(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupQualityConfig(t, tmpDir, `constitution:
  lsp_quality_gates:
    enabled: true
    sync:
      max_errors: 0
      max_warnings: 100
      require_clean_lsp: true
`)

	enforcer := NewQualityGateEnforcer(tmpDir)

	// Any diagnostics with require_clean_lsp=true should block
	counts := SeverityCounts{Errors: 0, Warnings: 1}
	result, err := enforcer.EnforcePhase(PhaseSync, counts, nil)
	if err != nil {
		t.Fatalf("EnforcePhase(sync) unexpected error: %v", err)
	}

	if !result.ShouldBlock {
		t.Error("sync phase should block when require_clean_lsp=true and any diagnostics present")
	}
}

// TestPhaseAwareEnforcer_SyncPhase_CleanLSP verifies clean LSP passes require_clean_lsp (REQ-QG-004).
func TestPhaseAwareEnforcer_SyncPhase_CleanLSP(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupQualityConfig(t, tmpDir, `constitution:
  lsp_quality_gates:
    enabled: true
    sync:
      max_errors: 0
      max_warnings: 10
      require_clean_lsp: true
`)

	enforcer := NewQualityGateEnforcer(tmpDir)
	counts := SeverityCounts{Errors: 0, Warnings: 0}
	result, err := enforcer.EnforcePhase(PhaseSync, counts, nil)
	if err != nil {
		t.Fatalf("EnforcePhase(sync) unexpected error: %v", err)
	}

	if result.ShouldBlock {
		t.Error("sync phase should not block when no diagnostics and require_clean_lsp=true")
	}
}

// TestPhaseAwareEnforcer_AutoPhase_FallsToRun verifies auto phase falls back to run settings.
func TestPhaseAwareEnforcer_AutoPhase_FallsToRun(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupQualityConfig(t, tmpDir, `constitution:
  lsp_quality_gates:
    enabled: true
    run:
      max_errors: 0
      max_type_errors: 0
      max_lint_errors: 0
      allow_regression: false
    sync:
      max_errors: 5
      max_warnings: 10
      require_clean_lsp: false
`)

	enforcer := NewQualityGateEnforcer(tmpDir)
	counts := SeverityCounts{Errors: 1}
	result, err := enforcer.EnforcePhase(PhaseAuto, counts, nil)
	if err != nil {
		t.Fatalf("EnforcePhase(auto) unexpected error: %v", err)
	}

	// Auto falls back to run: max_errors=0, should block on 1 error
	if !result.ShouldBlock {
		t.Error("auto phase (falling back to run) should block when errors > max_errors=0")
	}
}

// TestPhaseEnforceResult_ExitCode verifies EnforceResult.ExitCode (REQ-HOOK-181 compatibility).
func TestPhaseEnforceResult_ExitCode(t *testing.T) {
	t.Parallel()

	blocked := PhaseEnforceResult{ShouldBlock: true}
	if blocked.ExitCode() != ExitCodeQualityGateFailed {
		t.Errorf("ExitCode() = %d, want %d", blocked.ExitCode(), ExitCodeQualityGateFailed)
	}

	passed := PhaseEnforceResult{ShouldBlock: false}
	if passed.ExitCode() != ExitCodeSuccess {
		t.Errorf("ExitCode() = %d, want %d", passed.ExitCode(), ExitCodeSuccess)
	}
}

// TestLoadPhaseAwareConfig_UsesPkgModels verifies config loading uses pkg/models (REQ-QG-005, REQ-QG-006).
func TestLoadPhaseAwareConfig_UsesPkgModels(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupQualityConfig(t, tmpDir, `constitution:
  lsp_quality_gates:
    enabled: true
    plan:
      require_baseline: true
    run:
      max_errors: 3
      max_type_errors: 1
      max_lint_errors: 2
      allow_regression: false
    sync:
      max_errors: 0
      max_warnings: 10
      require_clean_lsp: true
    cache_ttl_seconds: 30
    timeout_seconds: 10
`)

	enforcer := NewQualityGateEnforcer(tmpDir)
	cfg, err := enforcer.LoadPhaseAwareConfig()
	if err != nil {
		t.Fatalf("LoadPhaseAwareConfig() unexpected error: %v", err)
	}

	if !cfg.Enabled {
		t.Error("Enabled should be true")
	}
	if !cfg.Plan.RequireBaseline {
		t.Error("Plan.RequireBaseline should be true")
	}
	if cfg.Run.MaxErrors != 3 {
		t.Errorf("Run.MaxErrors = %d, want 3", cfg.Run.MaxErrors)
	}
	if cfg.Run.MaxTypeErrors != 1 {
		t.Errorf("Run.MaxTypeErrors = %d, want 1", cfg.Run.MaxTypeErrors)
	}
	if cfg.Run.MaxLintErrors != 2 {
		t.Errorf("Run.MaxLintErrors = %d, want 2", cfg.Run.MaxLintErrors)
	}
	if !cfg.Sync.RequireCleanLSP {
		t.Error("Sync.RequireCleanLSP should be true")
	}
	if cfg.CacheTTLSeconds != 30 {
		t.Errorf("CacheTTLSeconds = %d, want 30", cfg.CacheTTLSeconds)
	}
	if cfg.TimeoutSeconds != 10 {
		t.Errorf("TimeoutSeconds = %d, want 10", cfg.TimeoutSeconds)
	}
}

// TestRegressionCheck_ErrorThreshold verifies regression detection per thresholds (REQ-QG-007).
func TestRegressionCheck_ErrorThreshold(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		baseline        SeverityCounts
		current         SeverityCounts
		errorThreshold  int
		warnThreshold   int
		wantRegression  bool
	}{
		{
			name:           "no regression - same counts",
			baseline:       SeverityCounts{Errors: 2, Warnings: 5},
			current:        SeverityCounts{Errors: 2, Warnings: 5},
			errorThreshold: 0,
			warnThreshold:  10,
			wantRegression: false,
		},
		{
			name:           "regression - errors increased",
			baseline:       SeverityCounts{Errors: 0},
			current:        SeverityCounts{Errors: 1},
			errorThreshold: 0,
			warnThreshold:  10,
			wantRegression: true,
		},
		{
			name:           "no regression - errors increased but within threshold",
			baseline:       SeverityCounts{Errors: 0},
			current:        SeverityCounts{Errors: 1},
			errorThreshold: 2,
			warnThreshold:  10,
			wantRegression: false,
		},
		{
			name:           "regression - warnings increased beyond threshold",
			baseline:       SeverityCounts{Warnings: 0},
			current:        SeverityCounts{Warnings: 11},
			errorThreshold: 0,
			warnThreshold:  10,
			wantRegression: true,
		},
		{
			name:           "no regression - warnings within threshold",
			baseline:       SeverityCounts{Warnings: 5},
			current:        SeverityCounts{Warnings: 10},
			errorThreshold: 0,
			warnThreshold:  10,
			wantRegression: false,
		},
		{
			name:           "improvement - errors decreased",
			baseline:       SeverityCounts{Errors: 5},
			current:        SeverityCounts{Errors: 3},
			errorThreshold: 0,
			warnThreshold:  10,
			wantRegression: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := DetectRegression(tt.current, tt.baseline, tt.errorThreshold, tt.warnThreshold)
			if got != tt.wantRegression {
				t.Errorf("DetectRegression() = %v, want %v", got, tt.wantRegression)
			}
		})
	}
}

// TestTRUST5Mapping verifies trust5_integration config is loaded (REQ-QG-009).
func TestTRUST5Mapping(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupQualityConfig(t, tmpDir, `constitution:
  lsp_quality_gates:
    enabled: true
    run:
      max_errors: 0
  lsp_integration:
    trust5_integration:
      tested:
        - lsp_errors == 0
        - lsp_type_errors == 0
      readable:
        - lsp_lint_errors == 0
`)

	enforcer := NewQualityGateEnforcer(tmpDir)
	trust5, err := enforcer.LoadTRUST5Config()
	if err != nil {
		t.Fatalf("LoadTRUST5Config() unexpected error: %v", err)
	}

	if len(trust5.Tested) == 0 {
		t.Error("TRUST5.Tested should not be empty")
	}
	if len(trust5.Readable) == 0 {
		t.Error("TRUST5.Readable should not be empty")
	}
}

// TestSeverityCounts_TypeErrorsAndLintErrors verifies new fields on SeverityCounts.
func TestSeverityCounts_TypeErrorsAndLintErrors(t *testing.T) {
	t.Parallel()

	counts := SeverityCounts{
		Errors:     2,
		TypeErrors: 1,
		LintErrors: 3,
		Warnings:   5,
	}

	if counts.TypeErrors != 1 {
		t.Errorf("TypeErrors = %d, want 1", counts.TypeErrors)
	}
	if counts.LintErrors != 3 {
		t.Errorf("LintErrors = %d, want 3", counts.LintErrors)
	}
}

// TestPhaseAwareEnforcer_MissingConfig verifies defaults when config file is missing.
func TestPhaseAwareEnforcer_MissingConfig(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// No quality.yaml created — enforcer should use defaults.
	enforcer := NewQualityGateEnforcer(tmpDir)

	// With default config (gates disabled), plan phase should not block.
	result, err := enforcer.EnforcePhase(PhasePlan, SeverityCounts{Errors: 5}, nil)
	if err != nil {
		t.Fatalf("EnforcePhase unexpected error: %v", err)
	}
	if result.ShouldBlock {
		t.Error("should not block when config is missing (defaults to disabled)")
	}

	// LoadPhaseAwareConfig should succeed with defaults.
	cfg, err := enforcer.LoadPhaseAwareConfig()
	if err != nil {
		t.Fatalf("LoadPhaseAwareConfig unexpected error: %v", err)
	}
	if cfg.Enabled {
		t.Error("default config should have Enabled=false")
	}

	// LoadTRUST5Config should succeed with empty config.
	trust5, err := enforcer.LoadTRUST5Config()
	if err != nil {
		t.Fatalf("LoadTRUST5Config unexpected error: %v", err)
	}
	_ = trust5 // empty slices are fine
}

// TestParseFullQualityConfig_InvalidYAML verifies error handling for invalid YAML.
func TestParseFullQualityConfig_InvalidYAML(t *testing.T) {
	t.Parallel()

	_, err := parseFullQualityConfig([]byte("{ invalid: ["))
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

// setupQualityConfig is a test helper that writes a quality.yaml to a temp dir.
func setupQualityConfig(t *testing.T, dir, content string) {
	t.Helper()
	configDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	configFile := filepath.Join(configDir, "quality.yaml")
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write quality.yaml: %v", err)
	}
}
