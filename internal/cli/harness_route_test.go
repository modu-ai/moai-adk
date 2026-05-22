package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness/router"
)

// TestHarnessRouterCmd — basic check of the newHarnessRouterCmd() factory.
func TestHarnessRouterCmd(t *testing.T) {
	t.Parallel()

	cmd := newHarnessRouterCmd()
	if cmd == nil {
		t.Fatal("newHarnessRouterCmd() returned nil")
	}
	useFirst := strings.SplitN(cmd.Use, " ", 2)[0]
	if useFirst != "harness" {
		t.Errorf("Use first token: got %q, want %q", useFirst, "harness")
	}

	// route + validate subcommands must be present
	subCmds := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		subCmds[strings.SplitN(sub.Use, " ", 2)[0]] = true
	}
	if !subCmds["route"] {
		t.Error("route subcommand not found")
	}
	if !subCmds["validate"] {
		t.Error("validate subcommand not found")
	}
}

// TestHarnessRouteJSONSchema — AC-HRN-001-06: JSON output schema compliance.
func TestHarnessRouteJSONSchema(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create harness.yaml for the test
	harnessYAML := `harness:
    default_profile: default
    effort_mapping:
        minimal: medium
        standard: high
        thorough: xhigh
    escalation:
        enabled: true
        max_escalations: 2
        triggers:
            - quality_gate_fail
    evaluator:
        memory_scope: per_iteration
    levels:
        minimal:
            description: minimal
            evaluator: false
            plan_audit:
                enabled: true
                max_iterations: 1
                require_must_pass: false
            skip_phases: []
            sprint_contract: false
        standard:
            description: standard
            evaluator: true
            evaluator_mode: final-pass
            plan_audit:
                enabled: true
                max_iterations: 3
                require_must_pass: true
            skip_phases: []
            sprint_contract: false
        thorough:
            description: thorough
            evaluator: true
            evaluator_mode: per-sprint
            evaluator_profile: strict
            plan_audit:
                enabled: true
                max_iterations: 3
                require_must_pass: true
            skip_phases: []
            sprint_contract: true
    mode_defaults:
        cg: thorough
        solo: auto
        team: auto
`
	harnessPath := filepath.Join(tmpDir, "harness.yaml")
	if err := os.WriteFile(harnessPath, []byte(harnessYAML), 0o644); err != nil {
		t.Fatalf("write harness.yaml: %v", err)
	}

	// Create a SPEC file for the test
	specYAML := `---
id: SPEC-TST-CLI-001
title: "CLI Route Test SPEC"
version: "0.1.0"
status: draft
created: 2026-05-18
updated: 2026-05-18
author: Test
priority: P2
phase: "v3.0.0"
module: "internal/test"
lifecycle: spec-anchored
tags: "feature, cli"
---

# SPEC-TST-CLI-001

## 5. Requirements

- REQ-TST-CLI-001-001 (Ubiquitous) — 기능 구현.
`
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-TST-CLI-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("mkdir spec dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specYAML), 0o644); err != nil {
		t.Fatalf("write spec.md: %v", err)
	}

	// Run the CLI (--json mode)
	var buf bytes.Buffer
	cmd := newHarnessRouterCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"route", "--spec", "SPEC-TST-CLI-001", "--json", "--path", harnessPath, "--base-dir", tmpDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("route command error: %v", err)
	}

	// Parse JSON
	output := buf.String()
	if output == "" {
		t.Fatal("empty JSON output")
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); err != nil {
		t.Fatalf("JSON parse error: %v (output: %q)", err, output)
	}

	// Verify required fields (AC-HRN-001-06)
	requiredFields := []string{"level", "rationale", "effort", "evaluator_profile", "sprint_contract", "plan_audit"}
	for _, field := range requiredFields {
		if _, ok := result[field]; !ok {
			t.Errorf("JSON output missing field: %q", field)
		}
	}

	// Verify rationale subfields
	rationale, ok := result["rationale"].(map[string]any)
	if !ok {
		t.Fatal("rationale field is not an object")
	}

	rationaleFields := []string{"matched_rule", "file_count", "domain_count", "spec_type", "spec_priority", "keywords"}
	for _, field := range rationaleFields {
		if _, ok := rationale[field]; !ok {
			t.Errorf("rationale missing field: %q", field)
		}
	}

	// Verify the level value is valid
	level, _ := result["level"].(string)
	switch router.Level(level) {
	case router.LevelMinimal, router.LevelStandard, router.LevelThorough:
		// valid
	default:
		t.Errorf("invalid level: %q", level)
	}
}

// TestHarnessValidateCmd — AC-HRN-001-04: validation of a well-formed harness.yaml.
func TestHarnessValidateCmd(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	harnessYAML := `harness:
    default_profile: default
    effort_mapping:
        minimal: medium
        standard: high
        thorough: xhigh
    escalation:
        enabled: true
        max_escalations: 2
        triggers:
            - quality_gate_fail
    evaluator:
        memory_scope: per_iteration
    levels:
        minimal:
            description: minimal
            evaluator: false
            plan_audit:
                enabled: true
                max_iterations: 1
                require_must_pass: false
            skip_phases: []
            sprint_contract: false
        standard:
            description: standard
            evaluator: true
            evaluator_mode: final-pass
            plan_audit:
                enabled: true
                max_iterations: 3
                require_must_pass: true
            skip_phases: []
            sprint_contract: false
        thorough:
            description: thorough
            evaluator: true
            evaluator_mode: per-sprint
            plan_audit:
                enabled: true
                max_iterations: 3
                require_must_pass: true
            skip_phases: []
            sprint_contract: true
    mode_defaults:
        cg: thorough
        solo: auto
        team: auto
`
	harnessPath := filepath.Join(tmpDir, "harness.yaml")
	if err := os.WriteFile(harnessPath, []byte(harnessYAML), 0o644); err != nil {
		t.Fatalf("write harness.yaml: %v", err)
	}

	var buf bytes.Buffer
	cmd := newHarnessRouterCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"validate", "--path", harnessPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("validate command error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "OK") {
		t.Errorf("validate output should contain 'OK', got: %q", output)
	}
}

// TestHarnessValidate_UnknownLevel — AC-HRN-001 level enum validation.
func TestHarnessValidate_UnknownLevel(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	harnessYAML := `harness:
    default_profile: default
    effort_mapping:
        minimal: medium
        standard: high
        thorough: xhigh
    escalation:
        enabled: true
        max_escalations: 2
        triggers:
            - quality_gate_fail
    evaluator:
        memory_scope: per_iteration
    levels:
        minimal:
            description: minimal
            evaluator: false
            plan_audit:
                enabled: true
                max_iterations: 1
                require_must_pass: false
            skip_phases: []
            sprint_contract: false
        expert:
            description: invalid level
            evaluator: true
            plan_audit:
                enabled: true
                max_iterations: 3
                require_must_pass: true
            skip_phases: []
            sprint_contract: false
    mode_defaults:
        cg: thorough
        solo: auto
        team: auto
`
	harnessPath := filepath.Join(tmpDir, "harness.yaml")
	if err := os.WriteFile(harnessPath, []byte(harnessYAML), 0o644); err != nil {
		t.Fatalf("write harness.yaml: %v", err)
	}

	var errBuf bytes.Buffer
	cmd := newHarnessRouterCmd()
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"validate", "--path", harnessPath})

	err := cmd.Execute()
	if err == nil {
		t.Error("validate with unknown level 'expert' should fail")
	}
}

// TestHarnessRouteSpecOverride — AC-HRN-001-09: spec_override matched_rule.
func TestHarnessRouteSpecOverride(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	harnessYAML := `harness:
    default_profile: default
    effort_mapping:
        minimal: medium
        standard: high
        thorough: xhigh
    escalation:
        enabled: true
        max_escalations: 2
        triggers:
            - quality_gate_fail
    evaluator:
        memory_scope: per_iteration
    levels:
        minimal:
            description: minimal
            evaluator: false
            plan_audit:
                enabled: true
                max_iterations: 1
                require_must_pass: false
            skip_phases: []
            sprint_contract: false
        standard:
            description: standard
            evaluator: true
            evaluator_mode: final-pass
            plan_audit:
                enabled: true
                max_iterations: 3
                require_must_pass: true
            skip_phases: []
            sprint_contract: false
        thorough:
            description: thorough
            evaluator: true
            evaluator_mode: per-sprint
            plan_audit:
                enabled: true
                max_iterations: 3
                require_must_pass: true
            skip_phases: []
            sprint_contract: true
    mode_defaults:
        cg: thorough
        solo: auto
        team: auto
`
	harnessPath := filepath.Join(tmpDir, "harness.yaml")
	if err := os.WriteFile(harnessPath, []byte(harnessYAML), 0o644); err != nil {
		t.Fatalf("write harness.yaml: %v", err)
	}

	// SPEC with harness_level: thorough (override)
	specYAML := `---
id: SPEC-TST-OVR-001
title: "Override Test SPEC"
version: "0.1.0"
status: draft
created: 2026-05-18
updated: 2026-05-18
author: Test
priority: P3
phase: "v3.0.0"
module: "internal/test"
lifecycle: spec-anchored
tags: "test"
harness_level: thorough
---

# SPEC-TST-OVR-001

## 5. Requirements

- REQ-TST-OVR-001-001 (Ubiquitous) — 단순 테스트.
`
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-TST-OVR-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specYAML), 0o644); err != nil {
		t.Fatalf("write spec.md: %v", err)
	}

	var buf bytes.Buffer
	cmd := newHarnessRouterCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"route", "--spec", "SPEC-TST-OVR-001", "--json", "--path", harnessPath, "--base-dir", tmpDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("route command error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &result); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}

	// AC-HRN-001-09: level == "thorough" AND matched_rule == "spec_override"
	if level, _ := result["level"].(string); level != "thorough" {
		t.Errorf("level: got %q, want %q", level, "thorough")
	}
	rationale, _ := result["rationale"].(map[string]any)
	if matchedRule, _ := rationale["matched_rule"].(string); matchedRule != "spec_override" {
		t.Errorf("matched_rule: got %q, want %q", matchedRule, "spec_override")
	}
}
