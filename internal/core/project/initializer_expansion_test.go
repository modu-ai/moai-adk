package project

// initializer_expansion_test.go — Tests for Phase 1 yaml write helpers.
// Table-driven, bytewise diff against expected fixtures (plan §M5 LEAN approach).
// All temp dirs use t.TempDir() for auto-cleanup.

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// setupSectionsDir creates the .moai/config/sections/ directory hierarchy.
func setupSectionsDir(t *testing.T) (root, sectionsDir string) {
	t.Helper()
	root = t.TempDir()
	sectionsDir = filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(sectionsDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	return root, sectionsDir
}

// TestWritePhase1Configs_NoOpWhenNotStandard verifies that Quick mode produces no files.
func TestWritePhase1Configs_NoOpWhenNotStandard(t *testing.T) {
	t.Parallel()
	root, sectionsDir := setupSectionsDir(t)

	opts := InitOptions{
		ProjectRoot:  root,
		StandardMode: false,
	}
	result := &InitResult{}

	if err := WritePhase1Configs(opts, result); err != nil {
		t.Fatalf("WritePhase1Configs: %v", err)
	}
	if len(result.CreatedFiles) != 0 {
		t.Errorf("no-op: got CreatedFiles = %v, want empty", result.CreatedFiles)
	}

	// Verify no Phase 1 files created
	for _, name := range []string{defs.HarnessYAML, defs.LSPYAML, defs.DesignYAML} {
		if _, err := os.Stat(filepath.Join(sectionsDir, name)); err == nil {
			t.Errorf("Quick mode: %s should not exist", name)
		}
	}
}

// TestWriteHarnessProfileYAML verifies harness.yaml content.
func TestWriteHarnessProfileYAML(t *testing.T) {
	t.Parallel()
	root, sectionsDir := setupSectionsDir(t)

	opts := InitOptions{
		ProjectRoot:    root,
		StandardMode:   true,
		HarnessProfile: "strict",
	}
	result := &InitResult{}
	if err := writeHarnessProfileYAML(sectionsDir, opts, result); err != nil {
		t.Fatalf("writeHarnessProfileYAML: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(sectionsDir, defs.HarnessYAML))
	if err != nil {
		t.Fatalf("ReadFile harness.yaml: %v", err)
	}
	want := []byte("harness:\n  default_profile: strict\n")
	if !bytes.Equal(got, want) {
		t.Errorf("harness.yaml content mismatch:\ngot:  %q\nwant: %q", got, want)
	}
}

// TestWriteHarnessProfileYAML_DefaultProfile verifies empty HarnessProfile defaults to "default".
func TestWriteHarnessProfileYAML_DefaultProfile(t *testing.T) {
	t.Parallel()
	root, sectionsDir := setupSectionsDir(t)

	opts := InitOptions{
		ProjectRoot:    root,
		StandardMode:   true,
		HarnessProfile: "", // empty → default
	}
	result := &InitResult{}
	if err := writeHarnessProfileYAML(sectionsDir, opts, result); err != nil {
		t.Fatalf("writeHarnessProfileYAML: %v", err)
	}

	got, _ := os.ReadFile(filepath.Join(sectionsDir, defs.HarnessYAML))
	want := []byte("harness:\n  default_profile: default\n")
	if !bytes.Equal(got, want) {
		t.Errorf("default profile mismatch:\ngot:  %q\nwant: %q", got, want)
	}
}

// TestWriteLSPYAML verifies lsp.yaml content for both enabled and disabled states.
func TestWriteLSPYAML(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		enabled bool
		want    string
	}{
		{"disabled (default)", false, "lsp:\n  enabled: false\n"},
		{"enabled", true, "lsp:\n  enabled: true\n"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			root, sectionsDir := setupSectionsDir(t)
			opts := InitOptions{ProjectRoot: root, StandardMode: true, LSPEnabled: c.enabled}
			result := &InitResult{}
			if err := writeLSPYAML(sectionsDir, opts, result); err != nil {
				t.Fatalf("writeLSPYAML: %v", err)
			}
			got, _ := os.ReadFile(filepath.Join(sectionsDir, defs.LSPYAML))
			if !bytes.Equal(got, []byte(c.want)) {
				t.Errorf("lsp.yaml mismatch:\ngot:  %q\nwant: %q", got, c.want)
			}
		})
	}
}

// TestWriteDesignYAML verifies design.yaml content for all combinations.
func TestWriteDesignYAML(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name              string
		designEnabled     bool
		claudeDesignEnabled bool
		want              string
	}{
		{
			"both enabled (default)",
			true, true,
			"design:\n  enabled: true\n  claude_design:\n    enabled: true\n",
		},
		{
			"design disabled",
			false, false,
			"design:\n  enabled: false\n  claude_design:\n    enabled: false\n",
		},
		{
			"design enabled, claude_design disabled",
			true, false,
			"design:\n  enabled: true\n  claude_design:\n    enabled: false\n",
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			root, sectionsDir := setupSectionsDir(t)
			opts := InitOptions{
				ProjectRoot:         root,
				StandardMode:        true,
				DesignEnabled:       c.designEnabled,
				ClaudeDesignEnabled: c.claudeDesignEnabled,
			}
			result := &InitResult{}
			if err := writeDesignYAML(sectionsDir, opts, result); err != nil {
				t.Fatalf("writeDesignYAML: %v", err)
			}
			got, _ := os.ReadFile(filepath.Join(sectionsDir, defs.DesignYAML))
			if !bytes.Equal(got, []byte(c.want)) {
				t.Errorf("design.yaml mismatch:\ngot:  %q\nwant: %q", got, c.want)
			}
		})
	}
}

// TestWriteQualityExpansionYAML_Fresh verifies quality.yaml with coverage_exemptions from scratch.
func TestWriteQualityExpansionYAML_Fresh(t *testing.T) {
	t.Parallel()
	root, sectionsDir := setupSectionsDir(t)

	opts := InitOptions{
		ProjectRoot:               root,
		StandardMode:              true,
		EnforceQuality:            false,
		CoverageExemptionsEnabled: true,
	}
	result := &InitResult{}
	if err := writeQualityExpansionYAML(sectionsDir, opts, result); err != nil {
		t.Fatalf("writeQualityExpansionYAML: %v", err)
	}

	got, _ := os.ReadFile(filepath.Join(sectionsDir, defs.QualityYAML))
	// Must contain coverage_exemptions block
	if !bytes.Contains(got, []byte("coverage_exemptions:")) {
		t.Error("quality.yaml missing coverage_exemptions block")
	}
	if !bytes.Contains(got, []byte("enabled: true")) {
		t.Error("quality.yaml coverage_exemptions.enabled should be true")
	}
	if !bytes.Contains(got, []byte("enforce_quality: false")) {
		t.Error("quality.yaml enforce_quality should be false")
	}
}

// TestWriteQualityExpansionYAML_ExistingFile verifies idempotent extension of an existing quality.yaml.
func TestWriteQualityExpansionYAML_ExistingFile(t *testing.T) {
	t.Parallel()
	root, sectionsDir := setupSectionsDir(t)

	// Write a pre-existing quality.yaml (as generateConfigsFallback would)
	existing := "constitution:\n  development_mode: tdd\n  enforce_quality: true\n  test_coverage_target: 85\n"
	if err := os.WriteFile(filepath.Join(sectionsDir, defs.QualityYAML), []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	opts := InitOptions{
		ProjectRoot:               root,
		StandardMode:              true,
		EnforceQuality:            false,
		CoverageExemptionsEnabled: false,
	}
	result := &InitResult{}
	if err := writeQualityExpansionYAML(sectionsDir, opts, result); err != nil {
		t.Fatalf("writeQualityExpansionYAML: %v", err)
	}

	got, _ := os.ReadFile(filepath.Join(sectionsDir, defs.QualityYAML))
	// coverage_exemptions block must be appended
	if !bytes.Contains(got, []byte("coverage_exemptions:")) {
		t.Error("quality.yaml missing coverage_exemptions block after extension")
	}
	// enforce_quality should be updated to false
	if !bytes.Contains(got, []byte("enforce_quality: false")) {
		t.Errorf("quality.yaml enforce_quality not updated; got:\n%s", got)
	}
}

// TestWritePhase1Configs_AllFiles verifies WritePhase1Configs creates all 4 Phase 1 files.
func TestWritePhase1Configs_AllFiles(t *testing.T) {
	t.Parallel()
	root, sectionsDir := setupSectionsDir(t)

	// Pre-create project.yaml (as generateConfigsFallback would)
	projectYAMLContent := `project:
  name: "test"
  description: ""
  mode: personal
  created_at: "2026-05-30T00:00:00Z"
  initialized: true
  optimized: false
  template_version: "v1.0.0"
`
	if err := os.WriteFile(filepath.Join(sectionsDir, defs.ProjectYAML), []byte(projectYAMLContent), 0644); err != nil {
		t.Fatal(err)
	}

	opts := InitOptions{
		ProjectRoot:               root,
		StandardMode:              true,
		ProjectMode:               "team",
		HarnessProfile:            "lenient",
		LSPEnabled:                true,
		EnforceQuality:            true,
		CoverageExemptionsEnabled: false,
		DesignEnabled:             true,
		ClaudeDesignEnabled:       false,
	}
	result := &InitResult{}
	if err := WritePhase1Configs(opts, result); err != nil {
		t.Fatalf("WritePhase1Configs: %v", err)
	}

	// Verify harness.yaml created with correct content
	harness, _ := os.ReadFile(filepath.Join(sectionsDir, defs.HarnessYAML))
	if !bytes.Contains(harness, []byte("default_profile: lenient")) {
		t.Errorf("harness.yaml: %q", harness)
	}

	// Verify lsp.yaml created with enabled=true
	lsp, _ := os.ReadFile(filepath.Join(sectionsDir, defs.LSPYAML))
	if !bytes.Equal(lsp, []byte("lsp:\n  enabled: true\n")) {
		t.Errorf("lsp.yaml: %q", lsp)
	}

	// Verify design.yaml created with correct values
	design, _ := os.ReadFile(filepath.Join(sectionsDir, defs.DesignYAML))
	if !bytes.Contains(design, []byte("enabled: true")) {
		t.Errorf("design.yaml missing enabled: true; got: %q", design)
	}
	if !bytes.Contains(design, []byte("claude_design:")) {
		t.Errorf("design.yaml missing claude_design block; got: %q", design)
	}

	// Verify project.yaml updated with mode=team
	project, _ := os.ReadFile(filepath.Join(sectionsDir, defs.ProjectYAML))
	if !bytes.Contains(project, []byte("mode: team")) {
		t.Errorf("project.yaml: mode not updated to team; got: %q", project)
	}
}

// TestYamlContains verifies the internal helper.
func TestYamlContains(t *testing.T) {
	t.Parallel()
	if !yamlContains("hello world", "world") {
		t.Error("expected true for contained substring")
	}
	if yamlContains("hello world", "xyz") {
		t.Error("expected false for absent substring")
	}
	if yamlContains("", "x") {
		t.Error("expected false for empty string")
	}
}

// TestSplitLines verifies the internal line splitter.
func TestSplitLines(t *testing.T) {
	t.Parallel()
	lines := splitLines("a\nb\nc\n")
	if len(lines) != 3 {
		t.Errorf("splitLines: got %d lines, want 3", len(lines))
	}
}

// TestPatchYAMLKey verifies the patcher correctly replaces a key value.
func TestPatchYAMLKey(t *testing.T) {
	t.Parallel()
	input := "constitution:\n  enforce_quality: true\n  test_coverage_target: 85\n"
	got := patchYAMLKey(input, "constitution", "enforce_quality", "false")
	if !bytes.Contains([]byte(got), []byte("enforce_quality: false")) {
		t.Errorf("patchYAMLKey did not replace value; got: %q", got)
	}
	// Ensure other keys are preserved
	if !bytes.Contains([]byte(got), []byte("test_coverage_target: 85")) {
		t.Errorf("patchYAMLKey removed unrelated key; got: %q", got)
	}
}

// TestWriteProjectModeYAML_FreshFile verifies project.yaml is created when it doesn't exist.
func TestWriteProjectModeYAML_FreshFile(t *testing.T) {
	t.Parallel()
	root, sectionsDir := setupSectionsDir(t)

	opts := InitOptions{ProjectRoot: root, StandardMode: true, ProjectMode: "team"}
	result := &InitResult{}
	if err := writeProjectModeYAML(sectionsDir, opts, result); err != nil {
		t.Fatalf("writeProjectModeYAML: %v", err)
	}

	got, _ := os.ReadFile(filepath.Join(sectionsDir, defs.ProjectYAML))
	if !bytes.Contains(got, []byte("mode: team")) {
		t.Errorf("fresh project.yaml missing mode: team; got: %q", got)
	}
}
