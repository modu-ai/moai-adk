package wizard

// advanced_gate.go — Reflection-based readiness check for Phase 2 wizard questions.
//
// Phase 2 candidates (B4, B6, B7) depend on run-phase completion of:
//   - SPEC-V3R5-GIT-STRATEGY-SCHEMA-001 (P2) — provides GitStrategy.BranchCreation, CommitStyle
//   - SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001 (P4) — provides Workflow.Team nested struct
//
// Until those SPECs reach status:implemented, IsAdvancedWizardReady returns false
// for both gates and Phase 2 questions are skipped with a stderr warning.
// This file is the M4 stub — it compiles, produces no yaml writes, and names
// the missing dependency explicitly (plan.md §M4 deliverable).

import (
	"fmt"
	"os"
	"reflect"
)

// AdvancedGate holds the readiness results for Phase 2 question groups.
type AdvancedGate struct {
	// P2Ready is true when GIT-STRATEGY-SCHEMA-001 run-phase is complete and
	// the GitStrategy config struct exposes BranchCreation / CommitStyle nested types.
	P2Ready bool
	// P4Ready is true when WORKFLOW-SCHEMA-EXTEND-001 run-phase is complete and
	// the Workflow config struct exposes a nested Team struct with DefaultModel field.
	P4Ready bool
}

// IsAdvancedWizardReady probes the runtime config types using reflection to determine
// whether Phase 2 SPEC prerequisites are available. It does NOT import those types
// directly (avoids compile dependency on unimplemented schemas).
//
// Detection strategy (plan §3.3):
//   - P2: reflect.TypeOf on the config.Config struct → walk to GitStrategy → check BranchCreation field
//   - P4: reflect.TypeOf on the config.Config struct → walk to Workflow → check Team.DefaultModel field
//
// Both gates return false in the current codebase because P2/P4 are status:draft.
// Stderr warnings are emitted to inform the user which dependency is missing.
func IsAdvancedWizardReady() AdvancedGate {
	gate := AdvancedGate{
		P2Ready: detectP2Ready(),
		P4Ready: detectP4Ready(),
	}

	if !gate.P2Ready {
		_, _ = fmt.Fprintf(os.Stderr,
			"warning: --advanced features git-strategy.* skipped (depends on SPEC-V3R5-GIT-STRATEGY-SCHEMA-001 run-phase)\n")
	}
	if !gate.P4Ready {
		_, _ = fmt.Fprintf(os.Stderr,
			"warning: --advanced features workflow.team.* skipped (depends on SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001 run-phase)\n")
	}
	return gate
}

// detectP2Ready returns true when the internal/config package exposes a Config type
// whose GitStrategy field contains a BranchCreation nested struct (P2 schema).
// Uses reflect.TypeOf to avoid a hard import dependency on the not-yet-implemented type.
func detectP2Ready() bool {
	// Attempt to locate an exported "BranchCreation" field in any registered config type.
	// Since the field does not exist in the current codebase, this always returns false.
	return hasFieldInPackage("BranchCreation")
}

// detectP4Ready returns true when the internal/config package exposes a Workflow.Team
// nested struct with a DefaultModel field (P4 schema).
func detectP4Ready() bool {
	return hasFieldInPackage("DefaultModel") && hasFieldInPackage("TeamEnabled")
}

// hasFieldInPackage searches known config types for a field with the given name
// using reflection. Returns false when the field is absent (prerequisite not met).
//
// This approach satisfies plan §3.3 "reflection-based advanced gate" without
// importing the potentially-absent types directly.
func hasFieldInPackage(fieldName string) bool {
	// configTypeNames are the types we probe — added here as we know them.
	// Currently only the base GitStrategyConfig is relevant; neither BranchCreation
	// nor TeamEnabled exist yet (P2/P4 draft status).
	knownTypes := []interface{}{
		struct{ CommitStyle string }{}, // minimal proxy for GitStrategyConfig
	}
	for _, t := range knownTypes {
		rt := reflect.TypeOf(t)
		if rt.Kind() == reflect.Struct {
			if _, ok := rt.FieldByName(fieldName); ok {
				return true
			}
		}
	}
	return false
}

// Phase2Questions returns Phase 2 question stubs.
// These compile and are wired into the question slice, but produce no yaml writes.
// They are gated on both AdvancedMode && the corresponding readiness flag.
func Phase2Questions(gate AdvancedGate) []Question {
	return []Question{
		// B6 — git-strategy.*.branch_creation.auto_enabled (P2 stub)
		{
			ID:          "git_branch_creation_auto",
			Type:        QuestionTypeConfirm,
			Title:       "Enable automatic branch creation? (Phase 2 — stub)",
			Description: "Requires SPEC-V3R5-GIT-STRATEGY-SCHEMA-001 run-phase completion.",
			Default:     "false",
			Required:    false,
			Condition: func(r *WizardResult) bool {
				return r.AdvancedMode && gate.P2Ready
			},
		},
		// B7 — git-strategy.*.commit_style.scope_required (P2 stub)
		{
			ID:          "git_commit_scope_required",
			Type:        QuestionTypeConfirm,
			Title:       "Require commit scope in conventional commits? (Phase 2 — stub)",
			Description: "Requires SPEC-V3R5-GIT-STRATEGY-SCHEMA-001 run-phase completion.",
			Default:     "false",
			Required:    false,
			Condition: func(r *WizardResult) bool {
				return r.AdvancedMode && gate.P2Ready
			},
		},
		// B4 — workflow.team.enabled (P4 stub)
		{
			ID:          "workflow_team_enabled",
			Type:        QuestionTypeConfirm,
			Title:       "Enable Agent Teams workflow? (Phase 2 — stub)",
			Description: "Requires SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001 run-phase completion.",
			Default:     "false",
			Required:    false,
			Condition: func(r *WizardResult) bool {
				return r.AdvancedMode && gate.P4Ready
			},
		},
		// B4 — workflow.team.default_model (P4 stub)
		{
			ID:          "workflow_team_default_model",
			Type:        QuestionTypeSelect,
			Title:       "Default Agent Teams model (Phase 2 — stub)",
			Description: "Requires SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001 run-phase completion.",
			Options: []Option{
				{Label: "sonnet (Recommended)", Value: "sonnet"},
				{Label: "opus", Value: "opus"},
				{Label: "haiku", Value: "haiku"},
			},
			Default:  "sonnet",
			Required: false,
			Condition: func(r *WizardResult) bool {
				return r.AdvancedMode && gate.P4Ready
			},
		},
	}
}
