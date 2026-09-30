package config

import "sort"

// closed_sets.go — exported accessors for the closed value sets that the web
// console and the TUI wizard render as select / radio widgets.
//
// These exist so no consumer re-declares an option list. A literal option set
// written a second time drifts silently: the widget keeps offering a value the
// validator no longer accepts (or refuses one it does), and nothing fails until
// a user hits it. Where a set already had a canonical home (the AuditModel* /
// AuditGate* constants, the validHarnessLevels map) these accessors derive from
// it rather than restating it.

// EvaluatorMemoryScopePerIteration is the FROZEN evaluator.memory_scope value
// (design-constitution §11.4.1). The loader rejects anything else with
// ErrEvalMemoryFrozen.
const EvaluatorMemoryScopePerIteration = "per_iteration"

// ExecutionModeAuto is the workflow.execution_mode value that defers the choice
// to harness auto-selection rather than pinning a mode.
const ExecutionModeAuto = "auto"

// ValidExecutionModePins returns the execution modes a user may pin, in the
// order the harness config lists them.
//
// These are exactly the keys of harness.mode_defaults: that map assigns a
// harness level to each execution mode, and execution_mode selects which of
// those modes is in force. The two are one concept seen from two sides, so this
// accessor is the shared declaration and the harness mode_defaults field list is
// built from it — a set written twice would drift, and the drift is silent
// (the console would refuse a mode the harness knows, or offer one it cannot
// resolve a level for).
//
// `cg` is a genuine member. Nothing in the Go tree reads ExecutionMode, so the
// value's meaning is carried by prose; the harness router contract names
// solo|team|cg as the modes consulted when execution_mode is `auto`, and the
// shipped harness.yaml declares all three. An earlier {auto, solo, team}
// declaration omitted `cg`, which mattered once the console became a closed-set
// widget: an unlisted value stops being savable.
func ValidExecutionModePins() []string {
	return []string{"solo", "team", "cg"}
}

// ValidExecutionModes returns the closed set for workflow.execution_mode:
// `auto` defers to harness auto-selection; the remaining members pin the shape.
func ValidExecutionModes() []string {
	return append([]string{ExecutionModeAuto}, ValidExecutionModePins()...)
}

// ValidWorkflowDefaultModes returns the closed set for workflow.default_mode —
// the `/moai run` mode dispatch values. An empty value is also legal and means
// "harness-based auto-selection"; it is carried by the widget's empty option
// rather than by this set.
func ValidWorkflowDefaultModes() []string {
	return []string{"autopilot", "loop", "team"}
}

// ValidGLMModels returns the closed set offered for the llm.glm.models.* tier
// slots, default-first: glm-5.3-flash (the default) leads, and no capability
// ordering between flash and glm-5.3 is claimed beyond that placement.
//
// DELETION + WITHDRAWAL RECORD (SPEC-MODEL-MATRIX-UPDATE-001 REQ-MMU-004,
// DR-2 — operator override 2026-09-30): the set is exactly
// {glm-5.3-flash, glm-5.3}. The former extras glm-5.1, glm-4.7, and
// glm-4.5-air are withdrawn from the offered set, and ALL old-model id
// surfaces — the named "glm-4.5"-through-"glm-5-turbo" constants, the
// statusline context-window entries, the legacy opus/sonnet/haiku alias
// fields — are deleted outright, superseding the glm-5.2 precedent of
// withdrawing the offered value while keeping its constant loadable. A
// stored tier-slot value naming a removed id resolves to the tier default
// with a one-line warning (fail-open); the llm.glm.context_windows user
// override key stays the path that still maps a custom window for any id.
//
// The members are DERIVED from the DefaultGLM* constants rather than restated:
// a second literal list would drift from the defaults the launcher actually
// injects, and the widget would keep offering a model id the runtime no longer
// maps. glm-5.3 is listed EXPLICITLY (DefaultGLM53) even though no tier slot
// defaults to it anymore: the set derives from constants, so a default
// retarget without the explicit member would silently drop glm-5.3 from the
// offered set and break an existing explicit selection.
func ValidGLMModels() []string {
	return []string{DefaultGLM53Flash, DefaultGLM53}
}

// ValidAuditModels returns the closed set for workflow.audit.model, derived
// from the AuditModel* constants that activeAuditBackend validates against.
func ValidAuditModels() []string {
	return []string{AuditModelClaude, AuditModelCodex, AuditModelGLM, AuditModelMulti}
}

// ValidCodexAuditModels returns the closed set for workflow.audit.codex.model
// (card t1278). The repo owns no codex model-id constants — the id rides the
// codex CLI — so the set is the operator-visible pair: the stored value from
// the 2026-09-30 settings screenshot (gpt-5.6-sol) and the schema example
// (gpt-5.6). Widening is a one-line change here; a stored foreign id keeps
// round-tripping through the RC2 passthrough-preserve in parseSchemaForm.
func ValidCodexAuditModels() []string {
	return []string{"gpt-5.6-sol", "gpt-5.6"}
}

// ValidProjectContinuations returns the closed set for
// workflow.project.continuation, derived from the ProjectContinuation*
// constants the resolver validates against
// (SPEC-PROJECT-CONTINUATION-KEY-001 REQ-PCK-001). Order is domain order, not
// alphabetical: none does least, pipeline does most.
func ValidProjectContinuations() []string {
	return []string{ProjectContinuationNone, ProjectContinuationCard, ProjectContinuationPipeline}
}

// ValidAuditGates returns the closed set for workflow.audit.gates.*, derived
// from the AuditGate* constants.
func ValidAuditGates() []string {
	return []string{AuditGateOff, AuditGateAdvisory, AuditGateRequired}
}

// ValidHarnessLevels returns the FROZEN harness level enum in sorted order,
// derived from the validHarnessLevels map the loader validates against. Sorting
// makes the render order deterministic (map iteration order is not).
func ValidHarnessLevels() []string {
	out := make([]string, 0, len(validHarnessLevels))
	for level := range validHarnessLevels {
		out = append(out, level)
	}
	sort.Strings(out)
	return out
}

// ValidModeDefaultLevels returns the closed set for harness.mode_defaults.*:
// the harness levels plus `auto`, which defers to the Complexity Estimator.
func ValidModeDefaultLevels() []string {
	return append([]string{"auto"}, ValidHarnessLevels()...)
}

// ValidEvaluatorMemoryScopes returns the closed set for
// harness.evaluator.memory_scope. It has exactly one member because the value
// is FROZEN; rendering it as a one-option radio is what makes that visible in
// the console instead of inviting a text edit the loader will reject.
func ValidEvaluatorMemoryScopes() []string {
	return []string{EvaluatorMemoryScopePerIteration}
}
