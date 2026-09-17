package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestExecutionModeDefaultMatchesTemplate pins the compiled default for
// workflow.execution_mode to the value shipped in the template workflow.yaml
// (SPEC-INIT-UPDATE-CONSISTENCY-001 REQ-ICU-002 / F9).
//
// The loader's partial-override contract means a workflow.yaml whose
// execution_mode key is absent gets the compiled default seeded, so a split
// between the two sources inverts the template-declared meaning. `auto` is the
// aligned value: the template ships it, closed_sets.go defines
// ExecutionModeAuto as "defers the choice to harness auto-selection", and
// execution_modes_test.go names it the defer-to-harness default. "team" is a
// relic from before the auto value existed.
func TestExecutionModeDefaultMatchesTemplate(t *testing.T) {
	repoRoot := findRepoRootFromCaller(t)
	templatePath := filepath.Join(repoRoot,
		"internal", "template", "templates", ".moai", "config", "sections", "workflow.yaml")

	data, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q): %v", templatePath, err)
	}

	var doc struct {
		Workflow struct {
			ExecutionMode string `yaml:"execution_mode"`
		} `yaml:"workflow"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("yaml.Unmarshal(%q): %v", templatePath, err)
	}
	if doc.Workflow.ExecutionMode == "" {
		t.Fatalf("template workflow.yaml carries no workflow.execution_mode key — the parity anchor is gone")
	}

	compiled := NewDefaultWorkflowConfig().ExecutionMode
	if compiled != doc.Workflow.ExecutionMode {
		t.Fatalf("compiled default workflow.execution_mode = %q but template workflow.yaml ships %q (SPEC-INIT-UPDATE-CONSISTENCY-001 REQ-ICU-002)",
			compiled, doc.Workflow.ExecutionMode)
	}
}
