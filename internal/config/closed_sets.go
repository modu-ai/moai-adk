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
// slots, in descending capability order.
//
// The members are DERIVED from the DefaultGLM* constants rather than restated:
// a second literal list would drift from the defaults the launcher actually
// injects, and the widget would keep offering a model id the runtime no longer
// maps. Note that these constants are not the only occurrences of these ids in
// the tree — the statusline context-window table keys on some of them too — but
// they are the SSOT for "which model may a tier slot hold", which is what this
// set answers.
func ValidGLMModels() []string {
	return []string{DefaultGLMHigh, DefaultGLM51, DefaultGLMMedium, DefaultGLMLow}
}

// ValidAuditModels returns the closed set for workflow.audit.model, derived
// from the AuditModel* constants that activeAuditBackend validates against.
func ValidAuditModels() []string {
	return []string{AuditModelClaude, AuditModelCodex, AuditModelGLM, AuditModelMulti}
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
