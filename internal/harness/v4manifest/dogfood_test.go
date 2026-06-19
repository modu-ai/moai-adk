// Package v4manifest — end-to-end dogfooding validation (AC-HV4-012a/012b).
//
// This test validates the v4 harness pipeline end-to-end at the COMPONENT level
// on a sample "moai-adk-dev" manifest: it constructs a realistic manifest,
// exercises the FULL v4 path (Validate → DecideIsolation per specialist →
// GenerateCommand → DecideEvaluator → RunnerTemplate stamp → dispatch-switch
// verification → round-trip write/re-read/re-Validate to a temp dir), and
// confirms every M3/M4/M5 component composes correctly.
//
// Scope note (AC-HV4-012b verification-claim integrity): this is a COMPONENT-
// LEVEL end-to-end dry-run, NOT a live orchestrator-driven harness build. The
// full live orchestrator-driven /moai:harness build (Context-First Discovery →
// ANALYZE fan-out → PLAN opus-xhigh → AskUserQuestion approval → GENERATE
// fan-out → ACTIVATE /goal+A/B) and the real-task with/without A/B comparison
// are NOT executed here — a subagent cannot drive the orchestrator-direct
// Builder. That gap is disclosed in the dogfooding-report.md 5-Section
// Evidence-Bearing Format per AC-HV4-012b.
//
// @MX:ANCHOR: [AUTO] TestDogfood_MoaiAdkDevHarnessEndToEnd is the v4 system composition gate.
// @MX:REASON: [AUTO] fan_in >= 3 candidate: M3/M4/M5 component integration, release gate, regression detection.
package v4manifest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sampleMoaiAdkDevManifest constructs a realistic "moai-adk-dev" harness
// manifest exercising the full v4 schema: 2 specialists (a read-only
// template-neutrality-auditor + a conflict-prone cli-codemaps-extractor), 2
// patterns from the 6-catalog, and a Sprint Contract with >= 1 dimension.
//
// The manifest is the kind the live Builder would produce for the natural-
// language request "moai-adk 개발을 위한 하네스를 구축해줘".
func sampleMoaiAdkDevManifest() Manifest {
	return Manifest{
		Name:          "moai-adk-dev",
		Domain:        "moai-adk CLI template development",
		SourceRequest: "moai-adk 개발을 위한 하네스를 구축해줘",
		Patterns:      []string{PatternFanOutFanIn, PatternExpertPool},
		Specialists: []Specialist{
			{
				Role:      "template-neutrality-auditor",
				Primitive: PrimitiveSubAgent,
				Isolation: IsolationNone,
				Effort:    EffortLow,
				Model:     ModelHaiku,
			},
			{
				Role:      "cli-codemaps-extractor",
				Primitive: PrimitiveWorktree,
				Isolation: IsolationWorktree,
				Effort:    EffortMedium,
				Model:     ModelSonnet,
			},
		},
		SprintContract: SprintContract{
			Dimensions: []string{"correctness", "template-neutrality", "coverage"},
			Thresholds: map[string]interface{}{
				"correctness":         0.95,
				"template-neutrality": 1.0,
				"coverage":            0.85,
			},
		},
		EntryCommand:   "/harness:moai-adk-dev",
		RunnerWorkflow: "harness-moai-adk-dev-run.js",
	}
}

// TestDogfood_MoaiAdkDevHarnessEndToEnd validates the v4 harness pipeline
// end-to-end on the sample "moai-adk-dev" manifest (AC-HV4-012a — component
// level). It exercises the FULL path a live Builder would walk:
//
//  1. Validate(manifest) passes (M3)
//  2. DecideIsolation per specialist returns none|worktree with rationale (M5)
//  3. GenerateCommand emits the thin-wrapper referencing harness-<name>-run.js (M4)
//  4. DecideEvaluator returns conditional skip/invoke (M3 Sprint Contract)
//  5. RunnerTemplate stamp + dispatch switch fires for each declared primitive
//  6. Round-trip: write manifest.json + command + runner to t.TempDir(),
//     re-read, re-Validate (plumbing integrity)
//
// What this test does NOT prove (AC-HV4-012b Gap): the full live orchestrator-
// driven /moai:harness build and the real-task with/without A/B. See
// dogfoading-report.md.
func TestDogfood_MoaiAdkDevHarnessEndToEnd(t *testing.T) {
	manifest := sampleMoaiAdkDevManifest()

	// (1) Validate — the manifest must satisfy the canonical schema (M3).
	if err := Validate(manifest); err != nil {
		t.Fatalf("step 1 Validate: expected the sample moai-adk-dev manifest to "+
			"pass the canonical schema, got error: %v", err)
	}

	// (2) DecideIsolation per specialist — the PLAN phase would consult this
	// helper per specialist before stamping the manifest's isolation field (M5).
	// The sample manifest's two specialists exercise both branches: read-only
	// (→ none) and conflict-prone-worktree (→ worktree).
	isolationResults := make([]IsolationDecision, 0, len(manifest.Specialists))
	for _, s := range manifest.Specialists {
		// Reconstruct the IsolationInput the PLAN phase would have built for
		// this specialist. The manifest's isolation field is the RESULT; this
		// test cross-validates that the DecideIsolation helper agrees.
		in := IsolationInput{
			Role:         s.Role,
			ReadOnly:     s.Primitive == PrimitiveSubAgent && s.Isolation == IsolationNone,
			Risky:        s.Isolation == IsolationWorktree,
			Parallel:     manifest.usesParallelPattern(),
			OverlapsPeer: false, // the 2 specialists target disjoint paths
			TargetPaths:  nil,
		}
		dec := DecideIsolation(in)
		if dec.Isolation != IsolationNone && dec.Isolation != IsolationWorktree {
			t.Errorf("step 2 DecideIsolation(%s): returned %q, want none|worktree",
				s.Role, dec.Isolation)
		}
		if dec.Rationale == "" {
			t.Errorf("step 2 DecideIsolation(%s): rationale is empty (NFR-HV4-002 "+
				"requires every isolation decision to be auditable)", s.Role)
		}
		isolationResults = append(isolationResults, dec)
	}
	// The template-neutrality-auditor (read-only) MUST get isolation=none.
	if isolationResults[0].Isolation != IsolationNone {
		t.Errorf("step 2: template-neutrality-auditor (read-only) got isolation=%q, "+
			"want none (AC-HV4-007a)", isolationResults[0].Isolation)
	}
	// The cli-codemaps-extractor (worktree-declared) MUST get isolation=worktree.
	if isolationResults[1].Isolation != IsolationWorktree {
		t.Errorf("step 2: cli-codemaps-extractor (risky) got isolation=%q, "+
			"want worktree (REQ-HV4-007 risky-changes carve-out)",
			isolationResults[1].Isolation)
	}

	// (3) GenerateCommand — emits the thin-wrapper command referencing the
	// Runner Workflow (M4). The command file dispatches /harness:<name>.
	commandMD, err := GenerateCommand(manifest)
	if err != nil {
		t.Fatalf("step 3 GenerateCommand: expected success, got error: %v", err)
	}
	if !strings.Contains(commandMD, manifest.Name) {
		t.Errorf("step 3: generated command does not reference harness name %q",
			manifest.Name)
	}
	if !strings.Contains(commandMD, manifest.RunnerWorkflow) {
		t.Errorf("step 3: generated command does not reference Runner Workflow %q",
			manifest.RunnerWorkflow)
	}
	if !strings.Contains(commandMD, "manifest.json") {
		t.Error("step 3: generated command does not reference manifest.json " +
			"(AC-HV4-006b single source of truth)")
	}

	// (4) DecideEvaluator — conditional skip/invoke per the Sprint Contract (M3).
	// A simple task within the model's solo range MUST skip the evaluator;
	// a complex task MUST invoke it with the dimensions echoed.
	simpleProfile := TaskProfile{
		WithinSoloRange:   true,
		ComplexitySignals: "single-skill generation, no adversarial-verification need",
	}
	simpleDecision := DecideEvaluator(simpleProfile, manifest.SprintContract)
	if simpleDecision.Invoked {
		t.Error("step 4: simple task (within solo range) invoked the evaluator; " +
			"want SKIPPED (AC-HV4-008b / C-HV4-001 simplest-solution-first)")
	}
	if simpleDecision.Rationale == "" {
		t.Error("step 4: simple-task evaluator skip has empty rationale")
	}
	complexProfile := TaskProfile{WithinSoloRange: false}
	complexDecision := DecideEvaluator(complexProfile, manifest.SprintContract)
	if !complexDecision.Invoked {
		t.Error("step 4: complex task (exceeds solo range) skipped the evaluator; " +
			"want INVOKED")
	}
	if len(complexDecision.Dimensions) != len(manifest.SprintContract.Dimensions) {
		t.Errorf("step 4: complex-task evaluator did not echo the Sprint Contract "+
			"dimensions (got %d, want %d) — NFR-HV4-002 observability",
			len(complexDecision.Dimensions), len(manifest.SprintContract.Dimensions))
	}

	// (5) RunnerTemplate stamp + dispatch switch fires for each declared
	// primitive. The Runner consumes specialist.primitive VERBATIM — no
	// heuristic re-derivation (AC-HV4-005b).
	runnerJS := stampRunnerForHarness(manifest.Name)
	for _, s := range manifest.Specialists {
		if !containsCaseDispatch(runnerJS, s.Primitive) {
			t.Errorf("step 5: Runner template dispatch switch does not cover "+
				"primitive %q (declared by specialist %q) — AC-HV4-005b verbatim "+
				"dispatch", s.Primitive, s.Role)
		}
	}
	// The Runner reads manifest.json as its SINGLE config-read path (AC-HV4-006b).
	if !strings.Contains(runnerJS, "manifest.json") {
		t.Error("step 5: Runner template does not read manifest.json " +
			"(AC-HV4-006b single config-read path)")
	}

	// (6) Round-trip: write manifest.json + command + runner to t.TempDir(),
	// re-read, re-Validate (plumbing integrity). This proves the generated
	// artifacts compose: a manifest written to disk is re-readable and still
	// schema-valid.
	dir := t.TempDir()
	harnessDir := filepath.Join(dir, ".claude", "commands", "harness", manifest.Name)
	if err := os.MkdirAll(harnessDir, 0o755); err != nil {
		t.Fatalf("step 6: mkdir %s: %v", harnessDir, err)
	}
	// manifest.json (the Runner's single config-read path).
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("step 6: marshal manifest: %v", err)
	}
	manifestPath := filepath.Join(harnessDir, "manifest.json")
	if err := os.WriteFile(manifestPath, manifestBytes, 0o644); err != nil {
		t.Fatalf("step 6: write manifest.json: %v", err)
	}
	// The thin-wrapper command file.
	commandPath := filepath.Join(dir, ".claude", "commands", "harness",
		manifest.Name+".md")
	if err := os.MkdirAll(filepath.Dir(commandPath), 0o755); err != nil {
		t.Fatalf("step 6: mkdir command dir: %v", err)
	}
	if err := os.WriteFile(commandPath, []byte(commandMD), 0o644); err != nil {
		t.Fatalf("step 6: write command: %v", err)
	}
	// The Runner Workflow script.
	runnerPath := filepath.Join(dir, ".claude", "workflows", manifest.RunnerWorkflow)
	if err := os.MkdirAll(filepath.Dir(runnerPath), 0o755); err != nil {
		t.Fatalf("step 6: mkdir workflow dir: %v", err)
	}
	if err := os.WriteFile(runnerPath, []byte(runnerJS), 0o644); err != nil {
		t.Fatalf("step 6: write runner: %v", err)
	}
	// Re-read the manifest and re-Validate.
	reread, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("step 6: re-read manifest.json: %v", err)
	}
	var roundTripped Manifest
	if err := json.Unmarshal(reread, &roundTripped); err != nil {
		t.Fatalf("step 6: unmarshal round-tripped manifest: %v", err)
	}
	if err := Validate(roundTripped); err != nil {
		t.Fatalf("step 6: round-tripped manifest failed re-Validate: %v", err)
	}
	// The round-tripped manifest MUST be identical to the original on the
	// load-bearing fields.
	if roundTripped.Name != manifest.Name {
		t.Errorf("step 6: round-trip Name drift: got %q, want %q",
			roundTripped.Name, manifest.Name)
	}
	if roundTripped.RunnerWorkflow != manifest.RunnerWorkflow {
		t.Errorf("step 6: round-trip RunnerWorkflow drift: got %q, want %q",
			roundTripped.RunnerWorkflow, manifest.RunnerWorkflow)
	}
	if len(roundTripped.Specialists) != len(manifest.Specialists) {
		t.Errorf("step 6: round-trip specialists count drift: got %d, want %d",
			len(roundTripped.Specialists), len(manifest.Specialists))
	}
	// manifest.json is present in the temp dir after write (construction-time
	// invariant).
	if _, err := os.Stat(filepath.Join(harnessDir, "manifest.json")); err != nil {
		t.Fatalf("step 6: manifest.json not present in temp dir after write: %v", err)
	}
}

// TestDogfood_EmitCleanupDirectiveFiresForWorktreeSpecialist confirms the
// Runner emits a worktree-cleanup directive when >= 1 specialist declared
// isolation:worktree (M5 EmitCleanupDirective + design §F). The sample
// moai-adk-dev manifest has exactly one such specialist (cli-codemaps-extractor).
func TestDogfood_EmitCleanupDirectiveFiresForWorktreeSpecialist(t *testing.T) {
	manifest := sampleMoaiAdkDevManifest()
	if !EmitCleanupDirective(manifest.Specialists) {
		t.Error("EmitCleanupDirective: expected true (sample manifest has 1 " +
			"isolation:worktree specialist), got false")
	}
	// A manifest with ZERO worktree specialists must NOT emit the directive.
	allNone := make([]Specialist, len(manifest.Specialists))
	for i, s := range manifest.Specialists {
		allNone[i] = Specialist{
			Role: s.Role, Primitive: s.Primitive, Isolation: IsolationNone,
			Effort: s.Effort, Model: s.Model,
		}
	}
	if EmitCleanupDirective(allNone) {
		t.Error("EmitCleanupDirective: expected false (all-none specialists), " +
			"got true — the directive must fire ONLY when >= 1 worktree specialist")
	}
}

// usesParallelPattern reports whether the manifest declares a pattern that
// implies parallel specialist dispatch (Fan-out/Fan-in, Expert Pool).
// Non-parallel patterns (Pipeline, Producer-Reviewer, Supervisor,
// Hierarchical Delegation) run specialists sequentially.
func (m Manifest) usesParallelPattern() bool {
	for _, p := range m.Patterns {
		if p == PatternFanOutFanIn || p == PatternExpertPool {
			return true
		}
	}
	return false
}

// stampRunnerForHarness simulates the Builder GENERATE phase stamping the
// RunnerTemplate with the harness name. The production stamp substitutes the
// manifest.json read path; for this test we assert the dispatch switch is
// present and primitive-coverage is complete (the stamp itself is a string
// substitution that does not change the switch body).
func stampRunnerForHarness(name string) string {
	out := RunnerTemplate
	out = strings.ReplaceAll(out, "<name>", name)
	return out
}

// containsCaseDispatch reports whether the Runner script body contains a
// dispatch case for the given primitive. The switch uses `case "primitive":`
// so we search for the quoted token.
func containsCaseDispatch(runnerJS, primitive string) bool {
	return strings.Contains(runnerJS, `case "`+primitive+`":`)
}
