package quality

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// wantCleanPassOutput returns the pass-path output RunAstGrepGateV2 produces
// for a project whose scan has nothing to report: the empty string when sg
// resolves (a scan really ran and found nothing) and the scanner-unavailable
// reason when it does not.
//
// Branching keeps the tests below deterministic on hosts with and without sg.
// It is load-bearing rather than defensive: those two outcomes used to collapse
// into the same empty string, which is exactly the false all-clear this package
// now refuses to emit.
func wantCleanPassOutput() string {
	if _, err := exec.LookPath("sg"); err != nil {
		return astGrepReasonScannerUnavailable
	}
	return ""
}

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

// TestRunAstGrepGateV2_NoSgCLI verifies the gate passes when sg is not on PATH
// but says why. Graceful degradation still means "never block a commit over an
// optional tool" (REQ-ASTG-UPG-012); it no longer means passing silently,
// because a silent pass is byte-identical to a clean scan.
//
// This assertion was previously `output == ""`. It is inverted rather than
// dropped, so it fails again if the gate ever goes back to reporting an absent
// scanner as nothing to say.
func TestRunAstGrepGateV2_NoSgCLI(t *testing.T) {
	// t.Setenv is incompatible with t.Parallel().
	t.Setenv("PATH", "")

	cfg := DefaultAstGrepGateConfig()
	// DefaultAstGrepGateConfig leaves RulesDir empty (t50: gate.yaml is the
	// SSOT), which now short-circuits to the unconfigured reason before the
	// scanner is consulted. This test pins the scanner-unavailable branch, so
	// it configures a rules dir explicitly — the value Default carried before
	// t50 removed the code-level default.
	cfg.RulesDir = ".moai/config/astgrep-rules"
	passed, output := RunAstGrepGateV2(context.Background(), t.TempDir(), cfg)

	if !passed {
		t.Errorf("gate should pass when sg is not available, got output: %q", output)
	}
	if output != astGrepReasonScannerUnavailable {
		t.Errorf("output should name the scanner-unavailable skip, want %q, got %q",
			astGrepReasonScannerUnavailable, output)
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
	// This test does not strip PATH, so the expected pass-path output depends
	// on whether the host has sg — see wantCleanPassOutput.
	if want := wantCleanPassOutput(); output != want {
		t.Errorf("output when rules dir does not exist: want %q, got %q", want, output)
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
	// Point at the (existing, empty) directory explicitly: Default now leaves
	// RulesDir empty, which means "not configured" rather than "the default
	// path exists but holds no rules" — the scenario this test pins.
	cfg.RulesDir = ".moai/config/astgrep-rules"
	passed, output := RunAstGrepGateV2(context.Background(), projectDir, cfg)

	if !passed {
		t.Errorf("gate should pass when no rules are loaded, got output: %q", output)
	}
	// This test does not strip PATH, so the expected pass-path output depends
	// on whether the host has sg — see wantCleanPassOutput.
	if want := wantCleanPassOutput(); output != want {
		t.Errorf("output when no rules are loaded: want %q, got %q", want, output)
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

	// PATH is stripped below, so every enabled case reaches the scanner and
	// finds it absent. wantOutput distinguishes the two pass shapes that used
	// to be one: an early skip that never consulted the scanner returns "",
	// while a case that did consult it names the skip.
	tests := []struct {
		name       string
		cfg        *AstGrepGateConfig
		wantPassed bool
		wantOutput string
		// expectation: "pass" means gate returns true and non-blocking.
		// "skip" means the gate returned before reaching the scanner at all.
		expectation string
	}{
		{
			name:        "nil config skips",
			cfg:         nil,
			wantPassed:  true,
			wantOutput:  "",
			expectation: "skip",
		},
		{
			name:        "disabled skips",
			cfg:         &AstGrepGateConfig{Enabled: false},
			wantPassed:  true,
			wantOutput:  "",
			expectation: "skip",
		},
		{
			name: "enabled with missing rules dir passes with the skip reason",
			cfg: &AstGrepGateConfig{
				Enabled:      true,
				RulesDir:     "path/does/not/exist",
				BlockOnError: true,
			},
			wantPassed: true,
			// The scanner is consulted before the rules directory is, so an
			// absent sg is reported even though the rules dir is also missing.
			wantOutput:  astGrepReasonScannerUnavailable,
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
			wantOutput:  astGrepReasonScannerUnavailable,
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
			wantOutput:  astGrepReasonScannerUnavailable,
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
			if output != tt.wantOutput {
				t.Errorf("%s: output = %q, want %q", tt.expectation, output, tt.wantOutput)
			}
		})
	}
}

// TestRunAstGrepGateV2_ProjectDirPathVariants verifies the gate tolerates a
// variety of projectDir inputs without panicking: absolute paths, relative
// paths, empty strings, and path-traversal attempts. In all cases the gate
// must return passed=true, and — because PATH is stripped below — must
// accompany that pass with the scanner-unavailable reason rather than the
// empty string a genuinely clean scan returns.
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
	// Explicit rules dir so these variants exercise the scanner step (and its
	// unavailable reason) rather than short-circuiting on the now-empty
	// default RulesDir (t50: empty means "not configured").
	cfg.RulesDir = ".moai/config/astgrep-rules"

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			passed, output := RunAstGrepGateV2(context.Background(), tt.projectDir, cfg)
			if !passed {
				t.Errorf("gate should pass for projectDir=%q, got blocked with output: %q",
					tt.projectDir, output)
			}
			if output != astGrepReasonScannerUnavailable {
				t.Errorf("gate should name the scanner-unavailable skip for projectDir=%q, want %q, got: %q",
					tt.projectDir, astGrepReasonScannerUnavailable, output)
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
	cfg.RulesDir = ".moai/config/astgrep-rules"
	passed, _ := RunAstGrepGateV2(ctx, t.TempDir(), cfg)

	if !passed {
		t.Error("cancelled context should still allow gate to pass gracefully")
	}
}

// TestRunAstGrepGateV2_ScanDegradedReason covers the second pass-path reason:
// the scanner was present but the scan itself failed, so its empty result is
// not a clean bill of health either.
//
// Reaching this branch end-to-end needs a resolvable sg whose output the
// scanner cannot parse — RunAstGrepGateV2 hard-codes SGBinary and offers no
// injection seam, so a fake sg emitting malformed JSON is the only lever. It
// reuses the fixture shape of TestRunAstGrepGateV2_AdvisoryOutputWithFindings.
//
// Without this the degraded branch would be reachable only in principle, and
// the two reason classes could quietly collapse into one in production while
// the constants stayed distinct in the unit test.
func TestRunAstGrepGateV2_ScanDegradedReason(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake sg shell script requires a POSIX shell")
	}

	binDir := t.TempDir()
	// Emits text that is not valid sg JSON, so parseSGFindings fails and Scan
	// returns a non-sentinel error.
	script := "#!/bin/sh\necho 'not json at all'\n"
	if err := os.WriteFile(filepath.Join(binDir, "sg"), []byte(script), 0o755); err != nil {
		t.Fatalf("write fake sg: %v", err)
	}
	// t.Setenv is incompatible with t.Parallel().
	t.Setenv("PATH", binDir)

	projectDir := t.TempDir()
	rulesDir := filepath.Join(projectDir, ".moai", "config", "astgrep-rules")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatalf("mkdir rules dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rulesDir, "sgconfig.yml"), []byte("ruleDirs:\n  - .\n"), 0o644); err != nil {
		t.Fatalf("write sgconfig.yml: %v", err)
	}

	passed, output := RunAstGrepGateV2(context.Background(), projectDir, &AstGrepGateConfig{
		Enabled:      true,
		RulesDir:     ".moai/config/astgrep-rules",
		BlockOnError: true,
	})

	if !passed {
		t.Errorf("a failed scan must never block a commit, got blocked with output: %q", output)
	}
	if output != astGrepReasonScanDegraded {
		t.Errorf("output should name the degraded scan, want %q, got %q", astGrepReasonScanDegraded, output)
	}
	if output == astGrepReasonScannerUnavailable {
		t.Error("a present-but-failing scanner must not be reported as an absent one")
	}
}

// TestRunAstGrepGateV2_AdvisoryOutputWithFindings characterizes advisory mode:
// with error-severity findings present and WarnOnlyMode enabled, the gate
// returns passed==true together with non-empty advisory output (the findings
// are reported but the commit is never blocked).
func TestRunAstGrepGateV2_AdvisoryOutputWithFindings(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake sg shell script requires a POSIX shell")
	}

	// Fake sg binary on PATH emitting one error-severity finding as sg JSON.
	binDir := t.TempDir()
	script := `#!/bin/sh
echo '[{"ruleId":"demo-rule","severity":"error","message":"demo finding","file":"main.go","range":{"start":{"line":0,"column":0},"end":{"line":0,"column":1}}}]'
`
	if err := os.WriteFile(filepath.Join(binDir, "sg"), []byte(script), 0o755); err != nil {
		t.Fatalf("write fake sg: %v", err)
	}
	// t.Setenv is incompatible with t.Parallel().
	t.Setenv("PATH", binDir)

	// Project with an sgconfig.yml so the scanner takes the single-invocation
	// config-based path.
	projectDir := t.TempDir()
	rulesDir := filepath.Join(projectDir, ".moai", "config", "astgrep-rules")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatalf("mkdir rules dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rulesDir, "sgconfig.yml"), []byte("ruleDirs:\n  - .\n"), 0o644); err != nil {
		t.Fatalf("write sgconfig.yml: %v", err)
	}

	cfg := &AstGrepGateConfig{
		Enabled:      true,
		RulesDir:     ".moai/config/astgrep-rules",
		BlockOnError: false,
		WarnOnlyMode: true,
	}

	passed, output := RunAstGrepGateV2(context.Background(), projectDir, cfg)

	if !passed {
		t.Errorf("advisory mode must never block, got blocked with output: %q", output)
	}
	if output == "" {
		t.Error("advisory mode with findings should return non-empty advisory output")
	}
	if !strings.Contains(output, "demo-rule") {
		t.Errorf("advisory output should contain the finding rule id, got: %q", output)
	}
}

// TestRunAstGrepGateV2_UnconfiguredRulesDir pins the t50 rework F2 guard:
// an empty gate.yaml ast_grep_gate.rules_dir must resolve to "not
// configured" — the gate passes with the unconfigured reason and never
// reaches the scanner. Before the guard, filepath.Join(projectDir, "")
// collapsed to projectDir itself, so the gate recursively loaded every
// .yml/.yaml in the project tree as incidental rules (a whole-tree walk on
// every gate run, and — via the config-load-failure fallback whose
// DefaultAstGrepGateConfig carries BlockOnError:true — a path for incidental
// error-severity rules to block commits).
//
// The decoy YAML files under the project root are loadable rule documents:
// under the old join-to-root behavior the walk would reach them (observable
// as the gate proceeding to the scanner step, reported here as the
// scanner-unavailable reason because PATH is stripped); with the guard the
// unconfigured reason is returned before any of that.
func TestRunAstGrepGateV2_UnconfiguredRulesDir(t *testing.T) {
	// t.Setenv is incompatible with t.Parallel().
	t.Setenv("PATH", "")

	projectDir := t.TempDir()
	// Decoys that a whole-tree walk would pick up as rule documents.
	decoyDirs := []string{
		filepath.Join(projectDir, ".moai", "config", "sections"),
		filepath.Join(projectDir, ".github", "workflows"),
	}
	for _, d := range decoyDirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	decoyRule := "id: decoy-incidental-rule\nlanguage: go\nseverity: error\nmessage: decoy\npattern: fmt.Println($X)\n"
	for _, fp := range []string{
		filepath.Join(decoyDirs[0], "gate.yaml"),
		filepath.Join(decoyDirs[1], "ci.yml"),
	} {
		if err := os.WriteFile(fp, []byte(decoyRule), 0o644); err != nil {
			t.Fatalf("write decoy %s: %v", fp, err)
		}
	}

	cfg := &AstGrepGateConfig{
		Enabled:      true,
		RulesDir:     "", // unconfigured (old template gate.yaml / empty explicit value)
		BlockOnError: true,
	}

	passed, output := RunAstGrepGateV2(context.Background(), projectDir, cfg)

	if !passed {
		t.Errorf("an unconfigured optional gate must pass, got blocked with output: %q", output)
	}
	if output != astGrepReasonRulesDirUnconfigured {
		t.Errorf("output should name the unconfigured rules dir, want %q, got %q",
			astGrepReasonRulesDirUnconfigured, output)
	}
}

// TestRunAstGrepGateV2_AbsoluteRulesDirUsedVerbatim pins the t50 rework F1
// guard: an absolute ast_grep_gate.rules_dir must be used verbatim, not
// joined onto the project root. Before the guard the gate computed
// Join("/proj", "/abs/rules") — a polluted nonexistent path — and silently
// scanned 0 rules while the CLI (which passes absolute values through)
// worked, so one config value meant two different directories depending on
// the consumer.
//
// The fake sg emits one error-severity finding, so a scan that actually used
// the absolute rules dir reports it in the advisory output; a joined
// (nonexistent) path yields a clean scan whose output is empty.
func TestRunAstGrepGateV2_AbsoluteRulesDirUsedVerbatim(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake sg shell script requires a POSIX shell")
	}

	// Fake sg binary on PATH emitting one error-severity finding as sg JSON.
	binDir := t.TempDir()
	script := `#!/bin/sh
echo '[{"ruleId":"demo-rule","severity":"error","message":"demo finding","file":"main.go","range":{"start":{"line":0,"column":0},"end":{"line":0,"column":1}}}]'
`
	if err := os.WriteFile(filepath.Join(binDir, "sg"), []byte(script), 0o755); err != nil {
		t.Fatalf("write fake sg: %v", err)
	}
	// t.Setenv is incompatible with t.Parallel().
	t.Setenv("PATH", binDir)

	// The rules dir lives OUTSIDE the project dir, addressed absolutely.
	projectDir := t.TempDir()
	rulesDir := filepath.Join(t.TempDir(), "shared-astgrep-rules")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatalf("mkdir rules dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rulesDir, "sgconfig.yml"), []byte("ruleDirs:\n  - .\n"), 0o644); err != nil {
		t.Fatalf("write sgconfig.yml: %v", err)
	}

	cfg := &AstGrepGateConfig{
		Enabled:      true,
		RulesDir:     rulesDir, // absolute — must pass through unjoined
		BlockOnError: true,
		WarnOnlyMode: true, // advisory: report findings without blocking
	}

	passed, output := RunAstGrepGateV2(context.Background(), projectDir, cfg)

	if !passed {
		t.Errorf("advisory mode must never block, got blocked with output: %q", output)
	}
	if !strings.Contains(output, "demo-rule") {
		t.Errorf("output should contain the finding from the absolute rules dir (used verbatim), got: %q", output)
	}
}
