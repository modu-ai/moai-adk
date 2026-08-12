// Package harness — doctor learning-axis tests.
// SPEC-HARNESS-EVO-RUN-REPORT-001 REQ-HRR-005 (AC-HRR-005, AC-HRR-010(c)).
//
// These tests cover the M3 learning axis added to checkHarness: the 3
// contract-validity checks (tier vocabulary, confidence_floor range,
// enabled+findings-declaration coherence) plus the legacy-absent case
// (learning block absent is NOT an ERROR — REQ-HRR-010 backward compat).
package harness

import (
	"path/filepath"
	"testing"
)

// doctorLearningManifestJSON returns a schema-valid v4 manifest whose learning
// block JSON body is supplied verbatim. An empty learningJSON omits the block
// entirely (the legacy 8-field shape).
func doctorLearningManifestJSON(name, learningJSON string) string {
	learningField := ""
	if learningJSON != "" {
		learningField = `,"learning": ` + learningJSON
	}
	return `{
  "name": "` + name + `",
  "domain": "test domain",
  "source_request": "test request",
  "patterns": ["Pipeline"],
  "specialists": [{"role":"tester","primitive":"sub-agent","isolation":"none","effort":"low","model":"inherit"}],
  "sprint_contract": {"dimensions":["correctness"],"thresholds":{"correctness":0.8}},
  "entry_command": "/harness:` + name + `",
  "runner_workflow": "harness-` + name + `-run.js"` + learningField + `
}`
}

// learningRunnerJS returns a Runner JS that resolves MANIFEST_PATH and carries
// (or omits) the `findings` return-schema key. No specialist agent reference is
// embedded so the agent-axis stays quiet, isolating the learning-axis signal.
// The findings-line comment is intentionally ABSENT so the regex heuristic is
// tested against the real array-key form, not prose.
func learningRunnerJS(manifestPath string, withFindings bool) string {
	findingsField := ""
	if withFindings {
		findingsField = ", findings: []"
	}
	return `const MANIFEST_PATH = "` + manifestPath + `";
async function run() {
  return { manifest: MANIFEST_PATH` + findingsField + ` };
}
module.exports = { run, MANIFEST_PATH };
`
}

// buildLearningHarnessFixture writes a minimal v4 harness (command + manifest +
// Runner) for learning-axis tests. The manifest and Runner JS are supplied
// verbatim so each test injects its own defect variant.
func buildLearningHarnessFixture(t *testing.T, root, name, manifestJSON, runnerJS string) {
	t.Helper()
	writeFile(t, filepath.Join(root, ".claude", "commands", "harness", name+".md"), "# /harness:"+name+"\nThin wrapper.\n")
	writeFile(t, filepath.Join(root, ".claude", "commands", "harness", name, "manifest.json"), manifestJSON)
	writeFile(t, filepath.Join(root, ".claude", "workflows", "harness-"+name+"-run.js"), runnerJS)
}

// hasLearningError reports whether the report carries at least one ERROR
// finding on the learning axis for the named harness.
func hasLearningError(report DoctorReport, harnessName string) bool {
	for _, f := range report.Findings {
		if f.Severity == SeverityError && f.Axis == "learning" && (harnessName == "" || f.Harness == harnessName) {
			return true
		}
	}
	return false
}

// TestDoctor_LearningAxis_TierTypo verifies AC-HRR-005 check 1: a learning.tier
// value outside the Tier.String() vocabulary ("recommendation" — the parallel
// vocabulary PIPE-REPAIR removed) yields an ERROR-severity finding and the gate
// exits non-zero.
//
// Note: the tier-vocabulary rejection is owned by v4manifest.Validate (M1,
// LearningBlock.Tier godoc + AP-1), which emits a manifest-axis ERROR when the
// schema is validated. The doctor does NOT re-check tier vocabulary (that would
// double-report). This test asserts the defect is caught as an ERROR at any
// axis — the schema path is the correct catch site for this defect class.
func TestDoctor_LearningAxis_TierTypo(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	manifest := doctorLearningManifestJSON("badtier", `{"enabled":true,"tier":"recommendation","confidence_floor":0.7,"max_findings_per_run":5}`)
	buildLearningHarnessFixture(t, root, "badtier", manifest,
		learningRunnerJS(".claude/commands/harness/badtier/manifest.json", true))

	report, err := Doctor(root)
	if err != nil {
		t.Fatalf("Doctor: %v", err)
	}
	if report.ErrorCount == 0 {
		t.Fatalf("error_count = 0, want >= 1 (invalid tier must fail the gate); findings=%+v", report.Findings)
	}
	// Confirm the tier defect was reported (manifest-axis via schema Validate).
	sawTierError := false
	for _, f := range report.Findings {
		if f.Severity == SeverityError && f.Harness == "badtier" {
			sawTierError = true
			break
		}
	}
	if !sawTierError {
		t.Errorf("expected an ERROR finding for the invalid tier; findings=%+v", report.Findings)
	}
}

// TestDoctor_LearningAxis_ConfidenceFloorOutOfRange verifies AC-HRR-005 check
// 2: a learning.confidence_floor outside [0,1] yields an ERROR-severity finding.
func TestDoctor_LearningAxis_ConfidenceFloorOutOfRange(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	manifest := doctorLearningManifestJSON("badfloor", `{"enabled":true,"tier":"rule","confidence_floor":1.5,"max_findings_per_run":5}`)
	buildLearningHarnessFixture(t, root, "badfloor", manifest,
		learningRunnerJS(".claude/commands/harness/badfloor/manifest.json", true))

	report, err := Doctor(root)
	if err != nil {
		t.Fatalf("Doctor: %v", err)
	}
	if !hasLearningError(report, "badfloor") {
		t.Errorf("expected a learning-axis ERROR for confidence_floor > 1; findings=%+v", report.Findings)
	}
	if report.ErrorCount == 0 {
		t.Errorf("error_count = 0, want >= 1 (out-of-range floor must fail the gate)")
	}
}

// TestDoctor_LearningAxis_EnabledButNoFindingsDeclaration verifies AC-HRR-005
// check 3: learning.enabled is true but the Runner does not declare a `findings`
// return-schema key → ERROR-severity finding.
func TestDoctor_LearningAxis_EnabledButNoFindingsDeclaration(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	manifest := doctorLearningManifestJSON("nofindings", `{"enabled":true,"tier":"rule","confidence_floor":0.7,"max_findings_per_run":5}`)
	// Runner WITHOUT a findings return key.
	buildLearningHarnessFixture(t, root, "nofindings", manifest,
		learningRunnerJS(".claude/commands/harness/nofindings/manifest.json", false))

	report, err := Doctor(root)
	if err != nil {
		t.Fatalf("Doctor: %v", err)
	}
	if !hasLearningError(report, "nofindings") {
		t.Errorf("expected a learning-axis ERROR for enabled+no-findings-declaration; findings=%+v", report.Findings)
	}
	if report.ErrorCount == 0 {
		t.Errorf("error_count = 0, want >= 1 (enabled harness must declare findings)")
	}
}

// TestDoctor_LearningAxis_AbsentNotError verifies AC-HRR-005 + REQ-HRR-010: a
// harness with NO learning block (legacy 8-field manifest) produces NO
// learning-axis ERROR (무보과 or INFO only), and the gate exits 0.
func TestDoctor_LearningAxis_AbsentNotError(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	manifest := doctorLearningManifestJSON("legacy", "") // no learning block
	buildLearningHarnessFixture(t, root, "legacy", manifest,
		learningRunnerJS(".claude/commands/harness/legacy/manifest.json", false))

	report, err := Doctor(root)
	if err != nil {
		t.Fatalf("Doctor: %v", err)
	}
	if hasLearningError(report, "legacy") {
		t.Errorf("legacy harness (no learning block) MUST NOT produce a learning ERROR; findings=%+v", report.Findings)
	}
	if report.ErrorCount != 0 {
		t.Errorf("error_count = %d, want 0 (learning-absent is backward-compat clean); findings=%+v", report.ErrorCount, report.Findings)
	}
}

// TestDoctor_LearningAxis_ValidBlockPasses verifies AC-HRR-005 happy path: a
// well-formed learning block (valid tier, in-range floor, enabled + findings
// declared) yields 0 ERROR findings.
func TestDoctor_LearningAxis_ValidBlockPasses(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	manifest := doctorLearningManifestJSON("goodlearn", `{"enabled":true,"tier":"auto_update","confidence_floor":0.70,"max_findings_per_run":5}`)
	buildLearningHarnessFixture(t, root, "goodlearn", manifest,
		learningRunnerJS(".claude/commands/harness/goodlearn/manifest.json", true))

	report, err := Doctor(root)
	if err != nil {
		t.Fatalf("Doctor: %v", err)
	}
	if report.ErrorCount != 0 {
		t.Errorf("error_count = %d, want 0 (valid learning block); findings=%+v", report.ErrorCount, report.Findings)
	}
	if hasLearningError(report, "goodlearn") {
		t.Errorf("valid learning block MUST NOT produce a learning ERROR; findings=%+v", report.Findings)
	}
}

// TestDoctor_LearningAxis_DisabledSkipsFindingsCheck verifies that when
// learning.enabled is false, the findings-declaration coherence check is NOT
// applied (the findings path is inert per LearningBlock.Enabled godoc). A
// disabled block with an absent findings declaration is still clean.
func TestDoctor_LearningAxis_DisabledSkipsFindingsCheck(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	manifest := doctorLearningManifestJSON("disabled", `{"enabled":false,"tier":"observation","confidence_floor":0.5,"max_findings_per_run":3}`)
	buildLearningHarnessFixture(t, root, "disabled", manifest,
		learningRunnerJS(".claude/commands/harness/disabled/manifest.json", false))

	report, err := Doctor(root)
	if err != nil {
		t.Fatalf("Doctor: %v", err)
	}
	if hasLearningError(report, "disabled") {
		t.Errorf("disabled learning block MUST NOT trigger the findings-declaration check; findings=%+v", report.Findings)
	}
	if report.ErrorCount != 0 {
		t.Errorf("error_count = %d, want 0 (disabled learning block is clean); findings=%+v", report.ErrorCount, report.Findings)
	}
}

// TestDoctor_LearningAxis_PreActionableTiersValid verifies AC-HRR-002: the
// pre-actionable tier values {observation, heuristic} are VALID (not ERROR) —
// they are non-actionable but within the SSOT vocabulary.
func TestDoctor_LearningAxis_PreActionableTiersValid(t *testing.T) {
	t.Parallel()
	for _, tier := range []string{"observation", "heuristic"} {
		root := t.TempDir()
		manifest := doctorLearningManifestJSON("preact", `{"enabled":true,"tier":"`+tier+`","confidence_floor":0.6,"max_findings_per_run":2}`)
		buildLearningHarnessFixture(t, root, "preact", manifest,
			learningRunnerJS(".claude/commands/harness/preact/manifest.json", true))

		report, err := Doctor(root)
		if err != nil {
			t.Fatalf("Doctor (tier=%s): %v", tier, err)
		}
		if hasLearningError(report, "preact") {
			t.Errorf("pre-actionable tier %q MUST be valid (no learning ERROR); findings=%+v", tier, report.Findings)
		}
	}
}

// TestDoctor_LearningAxis_EmptyTierAccepted verifies EC-1: an empty learning.tier
// is accepted as "unset" (defaulted downstream), not an ERROR at the doctor.
func TestDoctor_LearningAxis_EmptyTierAccepted(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	manifest := doctorLearningManifestJSON("emptytier", `{"enabled":true,"tier":"","confidence_floor":0.7,"max_findings_per_run":5}`)
	buildLearningHarnessFixture(t, root, "emptytier", manifest,
		learningRunnerJS(".claude/commands/harness/emptytier/manifest.json", true))

	report, err := Doctor(root)
	if err != nil {
		t.Fatalf("Doctor: %v", err)
	}
	if hasLearningError(report, "emptytier") {
		t.Errorf("empty learning.tier (unset, defaulted downstream per EC-1) MUST NOT be an ERROR; findings=%+v", report.Findings)
	}
}
