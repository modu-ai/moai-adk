// Package v4manifest — manifest validation (design §C.2).
//
// Validate checks a Manifest against the canonical schema rules:
//   - all 8 top-level fields required and non-empty
//   - specialists[] non-empty (>= 1 specialist)
//   - each specialist has all 5 sub-fields set to a valid enum value
//   - patterns[] entries are from the 6-pattern catalog
//   - sprint_contract has non-empty dimensions[] and non-nil thresholds
//
// @MX:ANCHOR: [AUTO] Validate is the single entry point for manifest validation.
// @MX:REASON: [AUTO] fan_in >= 3 candidate: GENERATE phase, CLI list/edit, Runner bootstrap
package v4manifest

import (
	"fmt"
	"strings"
)

// Validate returns nil if the manifest satisfies the canonical schema
// (design §C.2), or a non-nil error describing the first validation failure.
// The error wraps the failing field/rule for diagnosis.
func Validate(m Manifest) error {
	// 8 top-level fields required and non-empty (design §C.2 / AC-HV4-006a).
	if strings.TrimSpace(m.Name) == "" {
		return fmt.Errorf("v4manifest: name is required")
	}
	if strings.TrimSpace(m.Domain) == "" {
		return fmt.Errorf("v4manifest: domain is required")
	}
	if strings.TrimSpace(m.SourceRequest) == "" {
		return fmt.Errorf("v4manifest: source_request is required")
	}
	if len(m.Patterns) == 0 {
		return fmt.Errorf("v4manifest: patterns is required (>= 1 from the 6-pattern catalog)")
	}
	if strings.TrimSpace(m.EntryCommand) == "" {
		return fmt.Errorf("v4manifest: entry_command is required")
	}
	if strings.TrimSpace(m.RunnerWorkflow) == "" {
		return fmt.Errorf("v4manifest: runner_workflow is required")
	}

	// patterns[] entries MUST be from the 6-pattern catalog (design §C.2).
	for i, p := range m.Patterns {
		if !validPatterns[p] {
			return fmt.Errorf("v4manifest: patterns[%d] %q is not in the 6-pattern catalog", i, p)
		}
	}

	// specialists[] MUST be non-empty (>= 1 specialist) (design §C.2).
	if len(m.Specialists) == 0 {
		return fmt.Errorf("v4manifest: specialists must be non-empty (>= 1 specialist)")
	}

	// Each specialist has all 5 sub-fields set to a valid enum (AC-HV4-005a).
	for i, s := range m.Specialists {
		if strings.TrimSpace(s.Role) == "" {
			return fmt.Errorf("v4manifest: specialists[%d].role is required", i)
		}
		if !validPrimitives[s.Primitive] {
			return fmt.Errorf("v4manifest: specialists[%d].primitive %q is not one of the 5 primitives", i, s.Primitive)
		}
		if !validIsolations[s.Isolation] {
			return fmt.Errorf("v4manifest: specialists[%d].isolation %q is not none|worktree", i, s.Isolation)
		}
		if !validEfforts[s.Effort] {
			return fmt.Errorf("v4manifest: specialists[%d].effort %q is not low|medium|high|xhigh|max", i, s.Effort)
		}
		if !validModels[s.Model] {
			return fmt.Errorf("v4manifest: specialists[%d].model %q is not inherit|haiku|sonnet|opus", i, s.Model)
		}
	}

	// sprint_contract: dimensions non-empty + thresholds non-nil (REQ-HV4-008a).
	if len(m.SprintContract.Dimensions) == 0 {
		return fmt.Errorf("v4manifest: sprint_contract.dimensions must be non-empty")
	}
	if m.SprintContract.Thresholds == nil {
		return fmt.Errorf("v4manifest: sprint_contract.thresholds must be non-nil")
	}

	return nil
}

// IsValidPrimitive reports whether p is one of the 5 execution primitives.
// Exposed so the Runner engine can double-check a primitive before dispatch
// without re-implementing the set.
func IsValidPrimitive(p string) bool {
	return validPrimitives[p]
}
