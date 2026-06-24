package v4manifest

import (
	"strings"
	"testing"
)

// TestRunnerTemplate_DispatchesAllFivePrimitivesVerbatim verifies the Runner
// template consumes specialist.primitive verbatim and dispatches all 5
// primitives (AC-HV4-005b — no heuristic re-derivation).
func TestRunnerTemplate_DispatchesAllFivePrimitivesVerbatim(t *testing.T) {
	cases := []struct {
		primitive string
		marker    string // substring the dispatch branch carries
	}{
		{PrimitiveSubAgent, `case "sub-agent"`},
		{PrimitiveDynamicWorkflow, `case "dynamic-workflow"`},
		{PrimitiveWorktree, `case "worktree"`},
		{PrimitiveGoal, `case "/goal"`},
		{PrimitiveAdversarialFanOut, `case "adversarial-fan-out"`},
	}
	for _, tc := range cases {
		t.Run(tc.primitive, func(t *testing.T) {
			if !strings.Contains(RunnerTemplate, tc.marker) {
				t.Fatalf("RunnerTemplate missing dispatch branch %s for primitive %q", tc.marker, tc.primitive)
			}
			// The primitive literal must appear as the switch case value
			// (verbatim consumption, not a derived/normalized form).
			if !strings.Contains(RunnerTemplate, tc.primitive) {
				t.Fatalf("RunnerTemplate does not carry primitive literal %q verbatim", tc.primitive)
			}
		})
	}
}

// TestRunnerTemplate_NoHeuristicReDerivation verifies the Runner does NOT
// re-derive the primitive from heuristics (AC-HV4-005b). The dispatch must be
// a direct switch on specialist.primitive, with no heuristic inference.
func TestRunnerTemplate_NoHeuristicReDerivation(t *testing.T) {
	// Heuristic-re-derivation markers that would indicate the Runner is
	// inferring the primitive from task signals rather than reading it
	// verbatim. None of these should appear.
	heuristics := []string{
		"derivePrimitive",
		"inferPrimitive",
		"guessPrimitive",
		"selectPrimitiveHeuristic",
	}
	for _, h := range heuristics {
		if strings.Contains(RunnerTemplate, h) {
			t.Fatalf("RunnerTemplate contains heuristic-re-derivation marker %q (AC-HV4-005b violation)", h)
		}
	}
}

// TestRunnerTemplate_SingleConfigReadPath verifies the Runner reads ONLY
// manifest.json for its dispatch logic (AC-HV4-006b — single source of truth).
// There must be exactly one config-read entry point (readManifest), and it
// must read manifest.json (not a separate config file).
func TestRunnerTemplate_SingleConfigReadPath(t *testing.T) {
	// Exactly one readJson call site, and it targets manifest.json.
	if !strings.Contains(RunnerTemplate, "manifest.json") {
		t.Fatal("RunnerTemplate does not read manifest.json (AC-HV4-006b violation)")
	}
	// The SINGLE config-read path is readManifest(). Count its definition +
	// call sites — the contract is that ALL dispatch logic flows through it.
	if !strings.Contains(RunnerTemplate, "function readManifest()") {
		t.Fatal("RunnerTemplate missing readManifest() single config-read path")
	}
	// No alternative config-read patterns (hard-coded specialist info, separate
	// config files). These would indicate a second source of truth.
	forbidden := []string{
		"readConfig(",       // separate config file
		"specialists.json",  // a second specialists file
		"hardcodedSpecialists",
	}
	for _, f := range forbidden {
		if strings.Contains(RunnerTemplate, f) {
			t.Fatalf("RunnerTemplate contains forbidden second-config-read pattern %q (AC-HV4-006b violation)", f)
		}
	}
}

// TestRunnerTemplate_DeterministicScriptBody verifies the Runner script body
// is deterministic (C-HV4-003): no Date.now() or Math.random() calls. These
// would break resume caching (per dynamic-workflows.md determinism constraint).
func TestRunnerTemplate_DeterministicScriptBody(t *testing.T) {
	// Date.now() and Math.random() calls in the script body are prohibited.
	// (Mentioning them in a comment is fine per dynamic-workflows.md v2.1.172
	// fix, but an actual CALL breaks resume caching.)
	forbidden := []string{
		"Date.now()",
		"Math.random()",
	}
	for _, f := range forbidden {
		if strings.Contains(RunnerTemplate, f) {
			t.Fatalf("RunnerTemplate script body contains non-deterministic call %q (C-HV4-003 violation)", f)
		}
	}
}

// TestRunnerTemplate_AppliesSprintContractConditionally verifies the Runner
// applies the Sprint Contract (Generator-Evaluator separation) and that the
// evaluator is conditional (AC-HV4-008b — skipped when task within solo range).
func TestRunnerTemplate_AppliesSprintContractConditionally(t *testing.T) {
	if !strings.Contains(RunnerTemplate, "applySprintContract") {
		t.Fatal("RunnerTemplate missing applySprintContract (REQ-HV4-008)")
	}
	if !strings.Contains(RunnerTemplate, "withinSoloRange") {
		t.Fatal("RunnerTemplate missing conditional-evaluator skip logic (AC-HV4-008b)")
	}
	if !strings.Contains(RunnerTemplate, "evaluator: \"skipped\"") {
		t.Fatal("RunnerTemplate does not record evaluator=skipped with rationale (AC-HV4-008b)")
	}
	if !strings.Contains(RunnerTemplate, "evaluator: \"invoked\"") {
		t.Fatal("RunnerTemplate does not record evaluator=invoked for complex tasks (AC-HV4-008b)")
	}
}

// TestRunnerTemplate_EmitsWorktreeCleanupDirective verifies the Runner emits
// a cleanup directive when any specialist declared isolation:worktree
// (REQ-HV4-007 Runner end-of-run cleanup, R-HV4-002 mitigation).
func TestRunnerTemplate_EmitsWorktreeCleanupDirective(t *testing.T) {
	if !strings.Contains(RunnerTemplate, "emitCleanupDirective") {
		t.Fatal("RunnerTemplate missing cleanup directive emission")
	}
	if !strings.Contains(RunnerTemplate, `s.isolation === "worktree"`) {
		t.Fatal("RunnerTemplate does not gate cleanup on isolation:worktree specialists")
	}
}

// TestRunnerTemplate_TemplateNeutrality verifies the Runner template carries
// NO internal-state markers (C-HV4-005): no SPEC IDs, REQ tokens, AC tokens,
// or commit SHAs. The template is moai-distributable generic content.
func TestRunnerTemplate_TemplateNeutrality(t *testing.T) {
	forbidden := []string{
		"SPEC-V3R6-HARNESS-V4-001",
		"REQ-HV4-",
		"AC-HV4-",
	}
	for _, f := range forbidden {
		if strings.Contains(RunnerTemplate, f) {
			t.Fatalf("RunnerTemplate leaks internal-state marker %q (C-HV4-005 neutrality violation)", f)
		}
	}
}
