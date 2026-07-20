package config

import (
	"os"
	"path/filepath"
	"testing"
)

// This file houses the distinctness-invariant test suite for
// workflow.agentic_loop.max_iterations (SPEC-V3R6-AGENTIC-LOOP-CONFIG-001).
//
// The agentic_loop key (pipeline-level completion-loop ceiling, default 10) is
// DISTINCT from loop_prevention.max_iterations (per-operation diagnostic fix-loop
// bound, default 100). These tests mechanically enforce that the two keys parse
// to separate Go fields with separate defaults and never alias one another.
//
// AC-DISTINCT-001 (spec-level invariant): two distinct Go fields, separate defaults.
// AC-DISTINCT-002 (dedicated anti-aliasing test): parse-time distinctness under
// concurrent custom values.
// AC-002/003a/003b: explicit-value parse, absent-block fallback, absent-subkey fallback.

// writeWorkflowYAML writes a workflow.yaml fixture into a temp .moai/config/sections
// dir and returns the .moai root path suitable for Loader.Load().
func writeWorkflowYAML(t *testing.T, content string) string {
	t.Helper()
	tempDir := t.TempDir()
	sectionsDir := filepath.Join(tempDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sections dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sectionsDir, "workflow.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write workflow.yaml: %v", err)
	}
	return filepath.Join(tempDir, ".moai")
}

// TestAgenticLoopMaxIterations_ExplicitValue verifies that a user-supplied
// agentic_loop.max_iterations is parsed verbatim.
// AC-002: explicit custom value parses correctly.
func TestAgenticLoopMaxIterations_ExplicitValue(t *testing.T) {
	t.Parallel()

	yamlContent := "workflow:\n    agentic_loop:\n        max_iterations: 42\n"
	moaiRoot := writeWorkflowYAML(t, yamlContent)

	loader := NewLoader()
	cfg, err := loader.Load(moaiRoot)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Workflow.AgenticLoop.MaxIterations != 42 {
		t.Errorf("AgenticLoop.MaxIterations: got %d, want 42", cfg.Workflow.AgenticLoop.MaxIterations)
	}
}

// TestAgenticLoopMaxIterations_AbsentBlock_Default verifies that an absent
// agentic_loop block falls back to the default (NEVER zero, NEVER 100).
// AC-003a: absent block -> DefaultAgenticLoopMaxIterations.
func TestAgenticLoopMaxIterations_AbsentBlock_Default(t *testing.T) {
	t.Parallel()

	// workflow.yaml with no agentic_loop block -- only loop_prevention present.
	yamlContent := "workflow:\n    loop_prevention:\n        max_iterations: 100\n"
	moaiRoot := writeWorkflowYAML(t, yamlContent)

	loader := NewLoader()
	cfg, err := loader.Load(moaiRoot)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	got := cfg.Workflow.AgenticLoop.MaxIterations
	if got != DefaultAgenticLoopMaxIterations {
		t.Errorf("AgenticLoop.MaxIterations (absent block): got %d, want default %d",
			got, DefaultAgenticLoopMaxIterations)
	}
	// NEVER zero (the Go zero-value hazard -- REQ-ALC-002).
	if got == 0 {
		t.Error("AgenticLoop.MaxIterations (absent block): got 0, must NEVER be zero")
	}
	// NEVER collide with loop_prevention's 100 (distinctness invariant -- REQ-ALC-002).
	if got == cfg.Workflow.LoopPrevention.MaxIterations {
		t.Errorf("AgenticLoop.MaxIterations (%d) must NOT equal LoopPrevention.MaxIterations (%d) -- distinctness collision",
			got, cfg.Workflow.LoopPrevention.MaxIterations)
	}
}

// TestAgenticLoopMaxIterations_AbsentSubKey_Default verifies that a present
// agentic_loop parent block with the max_iterations sub-key omitted falls back
// to the default.
// AC-003b: present parent, absent sub-key -> DefaultAgenticLoopMaxIterations.
func TestAgenticLoopMaxIterations_AbsentSubKey_Default(t *testing.T) {
	t.Parallel()

	// agentic_loop block present but max_iterations omitted (YAML comment only).
	yamlContent := "workflow:\n    agentic_loop:\n        # max_iterations omitted\n"
	moaiRoot := writeWorkflowYAML(t, yamlContent)

	loader := NewLoader()
	cfg, err := loader.Load(moaiRoot)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	got := cfg.Workflow.AgenticLoop.MaxIterations
	if got != DefaultAgenticLoopMaxIterations {
		t.Errorf("AgenticLoop.MaxIterations (absent sub-key): got %d, want default %d",
			got, DefaultAgenticLoopMaxIterations)
	}
}

// TestAgenticLoopDistinctness_AntiAliasing is the DoD-blocker anti-aliasing test.
// It parses a YAML fixture setting BOTH keys to distinct custom values and asserts
// both land in the correct separate Go fields with zero cross-contamination.
// AC-DISTINCT-002: dedicated anti-aliasing test (parse-time distinctness).
func TestAgenticLoopDistinctness_AntiAliasing(t *testing.T) {
	t.Parallel()

	// Distinct custom values: 7 for agentic_loop, 99 for loop_prevention.
	// If the loader aliases the two fields, one of these would clobber the other.
	yamlContent := "workflow:\n" +
		"    agentic_loop:\n" +
		"        max_iterations: 7\n" +
		"    loop_prevention:\n" +
		"        max_iterations: 99\n"
	moaiRoot := writeWorkflowYAML(t, yamlContent)

	loader := NewLoader()
	cfg, err := loader.Load(moaiRoot)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Assertion 1: agentic_loop parsed to the correct field.
	if cfg.Workflow.AgenticLoop.MaxIterations != 7 {
		t.Errorf("AgenticLoop.MaxIterations: got %d, want 7", cfg.Workflow.AgenticLoop.MaxIterations)
	}
	// Assertion 2: loop_prevention parsed to the correct field.
	if cfg.Workflow.LoopPrevention.MaxIterations != 99 {
		t.Errorf("LoopPrevention.MaxIterations: got %d, want 99", cfg.Workflow.LoopPrevention.MaxIterations)
	}
	// Assertion 3: setting agentic_loop did NOT propagate to loop_prevention.
	if cfg.Workflow.LoopPrevention.MaxIterations == 7 {
		t.Error("distinctness violation: agentic_loop value 7 propagated to LoopPrevention.MaxIterations")
	}
	// Assertion 4: setting loop_prevention did NOT propagate to agentic_loop.
	if cfg.Workflow.AgenticLoop.MaxIterations == 99 {
		t.Error("distinctness violation: loop_prevention value 99 propagated to AgenticLoop.MaxIterations")
	}
}

// TestAgenticLoopDistinctness_RuntimeMutation verifies that the two fields occupy
// distinct memory addresses -- modifying one at runtime does NOT modify the other.
// AC-DISTINCT-001 evidence: runtime mutation non-propagation (no pointer aliasing).
func TestAgenticLoopDistinctness_RuntimeMutation(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultConfig()

	// Record the original loop_prevention value before mutating agentic_loop.
	originalLP := cfg.Workflow.LoopPrevention.MaxIterations

	// Mutate agentic_loop in-place.
	cfg.Workflow.AgenticLoop.MaxIterations = 999

	// Assert loop_prevention is unchanged.
	if cfg.Workflow.LoopPrevention.MaxIterations == 999 {
		t.Fatal("runtime aliasing: setting AgenticLoop.MaxIterations=999 changed LoopPrevention.MaxIterations to 999")
	}
	if cfg.Workflow.LoopPrevention.MaxIterations != originalLP {
		t.Errorf("LoopPrevention.MaxIterations changed after AgenticLoop mutation: got %d, want %d",
			cfg.Workflow.LoopPrevention.MaxIterations, originalLP)
	}

	// Symmetric check: mutate loop_prevention, assert agentic_loop unchanged.
	originalAL := cfg.Workflow.AgenticLoop.MaxIterations
	cfg.Workflow.LoopPrevention.MaxIterations = 888
	if cfg.Workflow.AgenticLoop.MaxIterations == 888 {
		t.Fatal("runtime aliasing: setting LoopPrevention.MaxIterations=888 changed AgenticLoop.MaxIterations to 888")
	}
	if cfg.Workflow.AgenticLoop.MaxIterations != originalAL {
		t.Errorf("AgenticLoop.MaxIterations changed after LoopPrevention mutation: got %d, want %d",
			cfg.Workflow.AgenticLoop.MaxIterations, originalAL)
	}
}

// TestAgenticLoopDefault_NeverZero_NeverLoopPrevention verifies the default
// config (no YAML overrides) has the correct distinct defaults for both fields.
// AC-DISTINCT-001: default-config distinctness (two fields, separate defaults).
func TestAgenticLoopDefault_NeverZero_NeverLoopPrevention(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultConfig()

	// Condition 1: agentic_loop default is DefaultAgenticLoopMaxIterations.
	if cfg.Workflow.AgenticLoop.MaxIterations != DefaultAgenticLoopMaxIterations {
		t.Errorf("default AgenticLoop.MaxIterations: got %d, want %d",
			cfg.Workflow.AgenticLoop.MaxIterations, DefaultAgenticLoopMaxIterations)
	}
	// Condition 2: loop_prevention default is 100 (the pre-existing default).
	if cfg.Workflow.LoopPrevention.MaxIterations != 100 {
		t.Errorf("default LoopPrevention.MaxIterations: got %d, want 100",
			cfg.Workflow.LoopPrevention.MaxIterations)
	}
	// Condition 3: the two defaults are distinct.
	if cfg.Workflow.AgenticLoop.MaxIterations == cfg.Workflow.LoopPrevention.MaxIterations {
		t.Errorf("default collision: AgenticLoop.MaxIterations == LoopPrevention.MaxIterations == %d",
			cfg.Workflow.AgenticLoop.MaxIterations)
	}
}
