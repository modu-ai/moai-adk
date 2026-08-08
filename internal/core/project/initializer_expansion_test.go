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

// TestWritePhase1Configs_NoOpWhenNotStandard was DELETED by C31: its subject
// was the standard-mode early return that C31 removes, so it asserts behaviour
// this SPEC deliberately eliminates and cannot be reconciled. It is named on
// the plan.md §G carve-out delete-list.

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
			opts := InitOptions{ProjectRoot: root, LSPEnabled: c.enabled}
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
		name                string
		designEnabled       bool
		claudeDesignEnabled bool
		want                string
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

// TestWritePhase1Configs_AllFiles pins the CREATE-IF-ABSENT fallback: when no
// lsp.yaml / design.yaml exists (the no-deployer path), WritePhase1Configs
// creates a minimal block. C35 deliberately preserved this branch, so it stays
// asserted here. The PATCH branch — which is what the deployer path actually
// takes — is pinned by TestWritePhase1Configs_PatchesExistingFiles below.
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
		ProjectMode:               "team",
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

	// C36: harness.yaml is NOT part of the Page-3 write set any more. On the
	// real (deployer) path the file is already deployed with the correct
	// default_profile, so WritePhase1Configs must leave it entirely alone —
	// here that means never creating it.
	if _, err := os.Stat(filepath.Join(sectionsDir, defs.HarnessYAML)); err == nil {
		t.Error("harness.yaml was written; C36 removed it from the Page-3 write set")
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

// TestWritePhase1Configs_PatchesExistingFiles pins the C35 read-patch semantics
// at the WritePhase1Configs (aggregate) level: when lsp.yaml and design.yaml
// ALREADY exist — which is the real deployer-path state — the Page-3 answers
// are patched into them in place, and every unrelated line survives.
//
// Non-vacuity: reverting writeLSPYAML / writeDesignYAML to their pre-C35
// wholesale os.WriteFile form fails this test, because the surrounding keys
// (`servers:`, `delegate_to_astgrep.enabled`, `gan_loop.enabled`, `figma`)
// would be erased rather than preserved.
func TestWritePhase1Configs_PatchesExistingFiles(t *testing.T) {
	t.Parallel()
	root, sectionsDir := setupSectionsDir(t)

	// Deployed-shaped fixtures: a nested same-named `enabled:` key at a deeper
	// indent in each file, mirroring the real lsp.yaml / design.yaml shapes.
	lspBefore := `lsp:
  enabled: false
  servers:
    go: gopls
  delegate_to_astgrep:
    enabled: true
`
	designBefore := `design:
  enabled: false
  gan_loop:
    enabled: true
  claude_design:
    enabled: false
  figma:
    enabled: true
`
	if err := os.WriteFile(filepath.Join(sectionsDir, defs.LSPYAML), []byte(lspBefore), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sectionsDir, defs.DesignYAML), []byte(designBefore), 0644); err != nil {
		t.Fatal(err)
	}

	opts := InitOptions{
		ProjectRoot:         root,
		LSPEnabled:          true,
		DesignEnabled:       true,
		ClaudeDesignEnabled: true,
	}
	if err := WritePhase1Configs(opts, &InitResult{}); err != nil {
		t.Fatalf("WritePhase1Configs: %v", err)
	}

	lspAfter, err := os.ReadFile(filepath.Join(sectionsDir, defs.LSPYAML))
	if err != nil {
		t.Fatalf("ReadFile lsp.yaml: %v", err)
	}
	wantLSP := `lsp:
  enabled: true
  servers:
    go: gopls
  delegate_to_astgrep:
    enabled: true
`
	if string(lspAfter) != wantLSP {
		t.Errorf("lsp.yaml not patched in place:\ngot:\n%s\nwant:\n%s", lspAfter, wantLSP)
	}

	designAfter, err := os.ReadFile(filepath.Join(sectionsDir, defs.DesignYAML))
	if err != nil {
		t.Fatalf("ReadFile design.yaml: %v", err)
	}
	wantDesign := `design:
  enabled: true
  gan_loop:
    enabled: true
  claude_design:
    enabled: true
  figma:
    enabled: true
`
	if string(designAfter) != wantDesign {
		t.Errorf("design.yaml not patched in place:\ngot:\n%s\nwant:\n%s", designAfter, wantDesign)
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

	opts := InitOptions{ProjectRoot: root, ProjectMode: "team"}
	result := &InitResult{}
	if err := writeProjectModeYAML(sectionsDir, opts, result); err != nil {
		t.Fatalf("writeProjectModeYAML: %v", err)
	}

	got, _ := os.ReadFile(filepath.Join(sectionsDir, defs.ProjectYAML))
	if !bytes.Contains(got, []byte("mode: team")) {
		t.Errorf("fresh project.yaml missing mode: team; got: %q", got)
	}
}
