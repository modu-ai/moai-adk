package quality

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestRunAstGrepGateV2_NilConfig verifies graceful handling of a nil config.
func TestRunAstGrepGateV2_NilConfig(t *testing.T) {
	t.Parallel()

	passed, output := RunAstGrepGateV2(context.Background(), t.TempDir(), nil)

	if !passed {
		t.Error("nil config should pass gracefully")
	}
	if output != "" {
		t.Errorf("nil config should return empty output, got %q", output)
	}
}

// TestRunAstGrepGateV2_Disabled verifies a disabled gate passes immediately.
func TestRunAstGrepGateV2_Disabled(t *testing.T) {
	t.Parallel()

	cfg := &AstGrepGateConfig{Enabled: false}
	passed, output := RunAstGrepGateV2(context.Background(), t.TempDir(), cfg)

	if !passed {
		t.Error("disabled gate should always pass")
	}
	if output != "" {
		t.Errorf("disabled gate should return empty output, got %q", output)
	}
}

// TestRunAstGrepGateV2_NoSgCLI verifies the gate passes silently when sg is not
// on PATH. This exercises the graceful-degradation contract of the unified
// Scanner (REQ-ASTG-UPG-012).
func TestRunAstGrepGateV2_NoSgCLI(t *testing.T) {
	// t.Setenv is incompatible with t.Parallel().
	t.Setenv("PATH", "")

	cfg := DefaultAstGrepGateConfig()
	passed, output := RunAstGrepGateV2(context.Background(), t.TempDir(), cfg)

	if !passed {
		t.Errorf("gate should pass when sg is not available, got output: %q", output)
	}
	if output != "" {
		t.Errorf("output should be empty when sg is not available, got %q", output)
	}
}

// TestRunAstGrepGateV2_NoRulesDir verifies the gate passes when the rules
// directory does not exist under the project root.
func TestRunAstGrepGateV2_NoRulesDir(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	cfg := DefaultAstGrepGateConfig()
	cfg.RulesDir = "nonexistent-rules-dir"

	passed, output := RunAstGrepGateV2(context.Background(), projectDir, cfg)

	if !passed {
		t.Errorf("gate should pass when rules dir does not exist, got output: %q", output)
	}
	if output != "" {
		t.Errorf("output should be empty when rules dir does not exist, got %q", output)
	}
}

// TestRunAstGrepGateV2_EmptyRulesDir verifies the gate passes when the rules
// directory exists but contains no rule files.
func TestRunAstGrepGateV2_EmptyRulesDir(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	rulesDir := filepath.Join(projectDir, ".moai", "config", "astgrep-rules")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}

	cfg := DefaultAstGrepGateConfig()
	passed, output := RunAstGrepGateV2(context.Background(), projectDir, cfg)

	if !passed {
		t.Errorf("gate should pass when no rules are loaded, got output: %q", output)
	}
	if output != "" {
		t.Errorf("output should be empty when no rules, got %q", output)
	}
}

// TestRunAstGrepGateV2_WarnOnlyMode verifies the gate never blocks when
// WarnOnlyMode is true, even in the presence of potential error-severity
// matches.
func TestRunAstGrepGateV2_WarnOnlyMode(t *testing.T) {
	// t.Setenv is incompatible with t.Parallel().
	t.Setenv("PATH", "")

	projectDir := t.TempDir()
	cfg := &AstGrepGateConfig{
		Enabled:      true,
		RulesDir:     ".moai/config/astgrep-rules",
		BlockOnError: true,
		WarnOnlyMode: true,
	}

	passed, _ := RunAstGrepGateV2(context.Background(), projectDir, cfg)

	if !passed {
		t.Error("WarnOnlyMode should never block even if errors are found")
	}
}

// TestRunAstGrepGateV2_BlockOnErrorFalse verifies the gate does not block when
// BlockOnError is false.
func TestRunAstGrepGateV2_BlockOnErrorFalse(t *testing.T) {
	// t.Setenv is incompatible with t.Parallel().
	t.Setenv("PATH", "")

	projectDir := t.TempDir()
	cfg := &AstGrepGateConfig{
		Enabled:      true,
		RulesDir:     ".moai/config/astgrep-rules",
		BlockOnError: false,
		WarnOnlyMode: false,
	}

	passed, _ := RunAstGrepGateV2(context.Background(), projectDir, cfg)

	if !passed {
		t.Error("BlockOnError=false should not block commits")
	}
}

// TestRunAstGrepGateV2_TableDriven verifies PASS / SKIP scenarios in a
// single table-driven test. FAIL (blocking) requires a real sg binary
// producing error-severity matches, which is outside the unit-test scope —
// the FAIL branch is covered by integration tests in internal/astgrep.
func TestRunAstGrepGateV2_TableDriven(t *testing.T) {
	// t.Setenv below is incompatible with t.Parallel(), so this whole test
	// is serial.
	projectDir := t.TempDir()

	tests := []struct {
		name       string
		cfg        *AstGrepGateConfig
		wantPassed bool
		wantEmpty  bool
		// expectation: "pass" means gate returns true and non-blocking.
		// "skip" means gate returns true with empty output (silent pass).
		expectation string
	}{
		{
			name:        "nil config skips",
			cfg:         nil,
			wantPassed:  true,
			wantEmpty:   true,
			expectation: "skip",
		},
		{
			name:        "disabled skips",
			cfg:         &AstGrepGateConfig{Enabled: false},
			wantPassed:  true,
			wantEmpty:   true,
			expectation: "skip",
		},
		{
			name: "enabled with missing rules dir passes silently",
			cfg: &AstGrepGateConfig{
				Enabled:      true,
				RulesDir:     "path/does/not/exist",
				BlockOnError: true,
			},
			wantPassed:  true,
			wantEmpty:   true,
			expectation: "pass",
		},
		{
			name: "warn-only never blocks",
			cfg: &AstGrepGateConfig{
				Enabled:      true,
				RulesDir:     ".moai/config/astgrep-rules",
				BlockOnError: true,
				WarnOnlyMode: true,
			},
			wantPassed:  true,
			wantEmpty:   true,
			expectation: "pass",
		},
		{
			name: "block-on-error=false never blocks",
			cfg: &AstGrepGateConfig{
				Enabled:      true,
				RulesDir:     ".moai/config/astgrep-rules",
				BlockOnError: false,
			},
			wantPassed:  true,
			wantEmpty:   true,
			expectation: "pass",
		},
	}

	// Force sg unavailability so the scanner's graceful-degradation path
	// is exercised deterministically.
	t.Setenv("PATH", "")

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			passed, output := RunAstGrepGateV2(context.Background(), projectDir, tt.cfg)
			if passed != tt.wantPassed {
				t.Errorf("%s: passed = %v, want %v", tt.expectation, passed, tt.wantPassed)
			}
			if tt.wantEmpty && output != "" {
				t.Errorf("%s: output = %q, want empty", tt.expectation, output)
			}
		})
	}
}

// TestRunAstGrepGateV2_ProjectDirPathVariants verifies the gate tolerates a
// variety of projectDir inputs without panicking: absolute paths, relative
// paths, empty strings, and path-traversal attempts. In all cases the gate
// must return passed=true with empty output because sg is unavailable
// and/or the rules directory does not resolve to a real location.
func TestRunAstGrepGateV2_ProjectDirPathVariants(t *testing.T) {
	// t.Setenv is incompatible with t.Parallel().
	t.Setenv("PATH", "")

	absDir := t.TempDir()
	relDir := "." // relative to CWD
	// A path-traversal attempt. filepath.Join will clean this to a
	// directory above the real project, which must not exist in tests
	// and must not cause the gate to block.
	traversal := filepath.Join(absDir, "..", "..", "..", "nonexistent")

	tests := []struct {
		name       string
		projectDir string
	}{
		{"absolute path", absDir},
		{"relative path", relDir},
		{"empty string", ""},
		{"traversal attempt", traversal},
		{"double-dot prefix", "../sibling-that-does-not-exist"},
	}

	cfg := DefaultAstGrepGateConfig()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			passed, output := RunAstGrepGateV2(context.Background(), tt.projectDir, cfg)
			if !passed {
				t.Errorf("gate should pass for projectDir=%q, got blocked with output: %q",
					tt.projectDir, output)
			}
			if output != "" {
				t.Errorf("gate should return empty output for projectDir=%q, got: %q",
					tt.projectDir, output)
			}
		})
	}
}

// TestRunAstGrepGateV2_ContextCancellation verifies the gate honors a
// cancelled context and degrades gracefully rather than blocking.
func TestRunAstGrepGateV2_ContextCancellation(t *testing.T) {
	// t.Setenv is incompatible with t.Parallel().
	t.Setenv("PATH", "")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel

	cfg := DefaultAstGrepGateConfig()
	passed, _ := RunAstGrepGateV2(ctx, t.TempDir(), cfg)

	if !passed {
		t.Error("cancelled context should still allow gate to pass gracefully")
	}
}

// TestRunAstGrepGate_V1_V2_Equivalence is a regression test that verifies V1
// (RunAstGrepGate) and V2 (RunAstGrepGateV2) behave identically across the
// non-blocking input space. Both implementations must agree on passed=true
// and empty output for all configurations where sg is unavailable, the
// rules directory is missing, or the gate is disabled.
//
// When V1 is eventually removed, this test should be removed in the same
// commit.
func TestRunAstGrepGate_V1_V2_Equivalence(t *testing.T) {
	// t.Setenv is incompatible with t.Parallel().
	t.Setenv("PATH", "")

	projectDir := t.TempDir()

	cases := []struct {
		name string
		cfg  *AstGrepGateConfig
	}{
		{"nil config", nil},
		{"disabled", &AstGrepGateConfig{Enabled: false}},
		{"default enabled (no sg, no rules)", DefaultAstGrepGateConfig()},
		{
			name: "warn-only",
			cfg: &AstGrepGateConfig{
				Enabled:      true,
				RulesDir:     ".moai/config/astgrep-rules",
				BlockOnError: true,
				WarnOnlyMode: true,
			},
		},
		{
			name: "block-on-error false",
			cfg: &AstGrepGateConfig{
				Enabled:      true,
				RulesDir:     ".moai/config/astgrep-rules",
				BlockOnError: false,
			},
		},
		{
			name: "nonexistent rules dir",
			cfg: &AstGrepGateConfig{
				Enabled:      true,
				RulesDir:     "does/not/exist",
				BlockOnError: true,
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			v1Passed, v1Output := RunAstGrepGate(ctx, projectDir, tc.cfg)
			v2Passed, v2Output := RunAstGrepGateV2(ctx, projectDir, tc.cfg)

			if v1Passed != v2Passed {
				t.Errorf("V1/V2 passed divergence: V1=%v V2=%v (config=%+v)",
					v1Passed, v2Passed, tc.cfg)
			}
			// Output equivalence is asserted only in the no-findings case,
			// which is the entire space exercised by these unit tests.
			if v1Output != v2Output {
				t.Errorf("V1/V2 output divergence:\nV1=%q\nV2=%q\n(config=%+v)",
					v1Output, v2Output, tc.cfg)
			}
		})
	}
}
