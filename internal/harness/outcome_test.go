// Package harness — Apply OUTCOME capture tests (SPEC-HARNESS-OUTCOME-CAPTURE-001).
//
// These tests cover the additive Apply-OUTCOME record: the new
// EventTypeApplyOutcome event type, the additive omitempty Event fields, the
// OutcomeRecord composition value, the RecordOutcome observer method, and the
// in-Apply seam emit (kept / rolled-back / gate-inactive). They assert that the
// capture is purely additive — it never alters the regression gate's keep/rollback
// decision, the returned error, the baseline-store update, or the M6 lineage write.
package harness

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// ─────────────────────────────────────────────
// M1 — Event type + additive omitempty fields (AC-OC-001, AC-OC-002)
// ─────────────────────────────────────────────

// TestApplyOutcomeEvent_RoundTrip verifies an apply_outcome Event round-trips
// the verdict / decision / proposal_id / baseline+candidate triples / regressed
// list through JSON marshal+unmarshal (AC-OC-001, REQ-OC-001, REQ-OC-004).
func TestApplyOutcomeEvent_RoundTrip(t *testing.T) {
	t.Parallel()

	evt := Event{
		EventType:           EventTypeApplyOutcome,
		Subject:             "apply:reg-001",
		ContextHash:         "ctx-abc",
		OutcomeVerdict:      "rolled-back",
		OutcomeDecision:     "regression-blocked",
		OutcomeProposalID:   "reg-001",
		OutcomeBaseTests:    100,
		OutcomeBaseCoverage: 87.0,
		OutcomeBaseLint:     0,
		OutcomeCandTests:    95,
		OutcomeCandCoverage: 80.0,
		OutcomeCandLint:     3,
		OutcomeRegressed:    []string{"tests_passed", "coverage", "lint_count"},
	}

	data, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Event
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.EventType != EventTypeApplyOutcome {
		t.Errorf("EventType = %q, want %q", got.EventType, EventTypeApplyOutcome)
	}
	if got.OutcomeVerdict != "rolled-back" {
		t.Errorf("OutcomeVerdict = %q, want rolled-back", got.OutcomeVerdict)
	}
	if got.OutcomeDecision != "regression-blocked" {
		t.Errorf("OutcomeDecision = %q, want regression-blocked", got.OutcomeDecision)
	}
	if got.OutcomeProposalID != "reg-001" {
		t.Errorf("OutcomeProposalID = %q, want reg-001", got.OutcomeProposalID)
	}
	if got.OutcomeBaseTests != 100 || got.OutcomeBaseCoverage != 87.0 || got.OutcomeBaseLint != 0 {
		t.Errorf("baseline triple = {%d %.1f %d}, want {100 87.0 0}", got.OutcomeBaseTests, got.OutcomeBaseCoverage, got.OutcomeBaseLint)
	}
	if got.OutcomeCandTests != 95 || got.OutcomeCandCoverage != 80.0 || got.OutcomeCandLint != 3 {
		t.Errorf("candidate triple = {%d %.1f %d}, want {95 80.0 3}", got.OutcomeCandTests, got.OutcomeCandCoverage, got.OutcomeCandLint)
	}
	if len(got.OutcomeRegressed) != 3 {
		t.Errorf("OutcomeRegressed = %v, want 3 dimensions", got.OutcomeRegressed)
	}
}

// TestApplyOutcomeEvent_OmitemptyOnOtherEvents verifies a non-outcome Event
// (e.g. a moai_subcommand event) marshals WITHOUT any outcome_* keys, proving
// the additive fields are omitempty (AC-OC-001, C9).
func TestApplyOutcomeEvent_OmitemptyOnOtherEvents(t *testing.T) {
	t.Parallel()

	evt := Event{
		EventType:   EventTypeMoaiSubcommand,
		Subject:     "/moai plan",
		ContextHash: "ctx-1",
	}
	data, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	for _, key := range []string{
		"outcome_verdict", "outcome_decision", "outcome_proposal_id",
		"outcome_baseline_tests", "outcome_baseline_coverage", "outcome_baseline_lint",
		"outcome_candidate_tests", "outcome_candidate_coverage", "outcome_candidate_lint",
		"outcome_regressed",
	} {
		if bytes.Contains(data, []byte(key)) {
			t.Errorf("non-outcome event JSON unexpectedly contains %q: %s", key, data)
		}
	}
}

// TestApplyOutcomeEvent_SchemaVersionV2 verifies a newly recorded event carries
// the current schema_version via the observer default-fill path (AC-OC-002, REQ-OC-010).
// The constant was bumped "v2" → "v2.1" by SPEC-V3R6-CONTEXT-GOV-AXIS-001 (REQ-CGA-002);
// this test guards that the constant is stamped on recorded events regardless of its value.
func TestApplyOutcomeEvent_SchemaVersionV2(t *testing.T) {
	t.Parallel()

	if LogSchemaVersion != "v2.1" {
		t.Fatalf("LogSchemaVersion = %q, want v2.1", LogSchemaVersion)
	}

	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	obs := NewObserver(logPath)

	if err := obs.RecordExtendedEvent(Event{EventType: EventTypeApplyOutcome, Subject: "s"}); err != nil {
		t.Fatalf("RecordExtendedEvent: %v", err)
	}
	line := readSingleJSONL(t, logPath)
	var evt Event
	if err := json.Unmarshal(line, &evt); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if evt.SchemaVersion != "v2.1" {
		t.Errorf("recorded schema_version = %q, want v2.1", evt.SchemaVersion)
	}
}

// ─────────────────────────────────────────────
// M2 — OutcomeRecord + RecordOutcome (AC-OC-003)
// ─────────────────────────────────────────────

// TestRecordOutcome_AppendsOneLine verifies RecordOutcome appends exactly one
// apply_outcome JSONL line carrying the verdict + proposal_id + triples, proving
// delegation to the shared RecordExtendedEvent append machinery (AC-OC-003, DD-6).
func TestRecordOutcome_AppendsOneLine(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	obs := NewObserver(logPath)

	rec := OutcomeRecord{
		Verdict:    "kept",
		Decision:   "approved",
		ProposalID: "p-001",
		Baseline:   MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0},
		Candidate:  MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0},
		Regressed:  nil,
	}
	if err := obs.RecordOutcome(rec); err != nil {
		t.Fatalf("RecordOutcome: %v", err)
	}

	lines := readAllJSONL(t, logPath)
	if len(lines) != 1 {
		t.Fatalf("line count = %d, want 1", len(lines))
	}
	var evt Event
	if err := json.Unmarshal(lines[0], &evt); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if evt.EventType != EventTypeApplyOutcome {
		t.Errorf("EventType = %q, want apply_outcome", evt.EventType)
	}
	if evt.OutcomeVerdict != "kept" {
		t.Errorf("OutcomeVerdict = %q, want kept", evt.OutcomeVerdict)
	}
	if evt.OutcomeProposalID != "p-001" {
		t.Errorf("OutcomeProposalID = %q, want p-001", evt.OutcomeProposalID)
	}
	if evt.OutcomeBaseTests != 100 || evt.OutcomeCandTests != 100 {
		t.Errorf("triples not recorded: base=%d cand=%d", evt.OutcomeBaseTests, evt.OutcomeCandTests)
	}
}

// TestRecordOutcome_AutoCreatesDir verifies RecordOutcome auto-creates the parent
// directory (delegation to RecordExtendedEvent's mkdir machinery) (AC-OC-003).
func TestRecordOutcome_AutoCreatesDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logPath := filepath.Join(dir, "nested", "deeper", "usage-log.jsonl")
	obs := NewObserver(logPath)

	if err := obs.RecordOutcome(OutcomeRecord{Verdict: "kept", Decision: "approved", ProposalID: "p-002"}); err != nil {
		t.Fatalf("RecordOutcome: %v", err)
	}
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("log file not created in nested dir: %v", err)
	}
}

// ─────────────────────────────────────────────
// M3 — In-Apply seam emit (AC-OC-004, AC-OC-005, AC-OC-007, AC-OC-008)
// ─────────────────────────────────────────────

// TestApply_Outcome_Kept verifies a non-regressing Apply through the active gate
// keeps the change (returns nil) AND records exactly one apply_outcome event with
// verdict "kept" / decision "approved" / matching proposal_id (AC-OC-004,
// REQ-OC-002, REQ-OC-005). Non-interference: the gate's nil return is unchanged.
func TestApply_Outcome_Kept(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	skillPath := writeLineageFixture(t)
	snapshotBase := filepath.Join(dir, "snapshots")
	manifestPath := filepath.Join(dir, "learning-history", "manifest.jsonl")
	baselinePath := filepath.Join(dir, "harness", "measurements-baseline.yaml")
	logPath := filepath.Join(dir, "harness", "usage-log.jsonl")

	m := &seqMeasurer{calls: []measureCall{
		{triple: MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0}},
		{triple: MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0}},
	}}
	a := newGateApplier(manifestPath, baselinePath, m)
	a.outcomeObserver = NewObserver(logPath)

	proposal := Proposal{
		ID:               "kept-oc-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "harness frequently triggered",
		PatternKey:       "moai_subcommand:/moai plan:ctx001",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}

	if err := a.Apply(proposal, approvedEvaluator(), snapshotBase, []Session{}); err != nil {
		t.Fatalf("Apply (kept) must return nil (non-interference): %v", err)
	}

	evt := readSingleOutcomeEvent(t, logPath)
	if evt.OutcomeVerdict != "kept" {
		t.Errorf("OutcomeVerdict = %q, want kept", evt.OutcomeVerdict)
	}
	if evt.OutcomeDecision != "approved" {
		t.Errorf("OutcomeDecision = %q, want approved", evt.OutcomeDecision)
	}
	if evt.OutcomeProposalID != "kept-oc-001" {
		t.Errorf("OutcomeProposalID = %q, want kept-oc-001", evt.OutcomeProposalID)
	}
}

// TestApply_Outcome_RolledBack verifies a regressing Apply through the active
// gate rolls back + returns *ApplyRegressionError (unchanged) AND records exactly
// one apply_outcome event with verdict "rolled-back" / decision
// "regression-blocked" / the regressed list / matching proposal_id (AC-OC-005).
func TestApply_Outcome_RolledBack(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	skillPath := writeLineageFixture(t)
	originalContent, _ := os.ReadFile(skillPath)
	snapshotBase := filepath.Join(dir, "snapshots")
	manifestPath := filepath.Join(dir, "learning-history", "manifest.jsonl")
	baselinePath := filepath.Join(dir, "harness", "measurements-baseline.yaml")
	logPath := filepath.Join(dir, "harness", "usage-log.jsonl")

	seedBaseline(t, baselinePath, MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0})
	m := &seqMeasurer{calls: []measureCall{
		{triple: MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0}},
		{triple: MetricTriple{TestsPassed: 100, Coverage: 80.0, LintCount: 3}},
	}}
	a := newGateApplier(manifestPath, baselinePath, m)
	a.outcomeObserver = NewObserver(logPath)

	proposal := Proposal{
		ID:               "rb-oc-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "harness frequently triggered",
		PatternKey:       "moai_subcommand:/moai plan:ctx001",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}

	err := a.Apply(proposal, approvedEvaluator(), snapshotBase, []Session{})
	var regErr *ApplyRegressionError
	if !errors.As(err, &regErr) {
		t.Fatalf("error type = %T, want *ApplyRegressionError (non-interference)", err)
	}

	// File rolled back to original bytes (gate decision unchanged).
	after, _ := os.ReadFile(skillPath)
	if !bytes.Equal(originalContent, after) {
		t.Errorf("file not rolled back (non-interference broken)")
	}

	evt := readSingleOutcomeEvent(t, logPath)
	if evt.OutcomeVerdict != "rolled-back" {
		t.Errorf("OutcomeVerdict = %q, want rolled-back", evt.OutcomeVerdict)
	}
	if evt.OutcomeDecision != "regression-blocked" {
		t.Errorf("OutcomeDecision = %q, want regression-blocked", evt.OutcomeDecision)
	}
	if evt.OutcomeProposalID != "rb-oc-001" {
		t.Errorf("OutcomeProposalID = %q, want rb-oc-001", evt.OutcomeProposalID)
	}
	if len(evt.OutcomeRegressed) == 0 {
		t.Errorf("OutcomeRegressed must list the regressed dimensions, got %v", evt.OutcomeRegressed)
	}
}

// TestApply_Outcome_GateInactive_NoEmit verifies the gate-INACTIVE Apply path
// (NewApplier default, no measurer/baseline) records NO apply_outcome event even
// when an observer is wired — DD-4 (no measured triple) (AC-OC-007).
func TestApply_Outcome_GateInactive_NoEmit(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	skillPath := writeLineageFixture(t)
	snapshotBase := filepath.Join(dir, "snapshots")
	logPath := filepath.Join(dir, "harness", "usage-log.jsonl")

	a := NewApplier() // gate inactive (no measurer / baseline store)
	a.outcomeObserver = NewObserver(logPath)

	proposal := Proposal{
		ID:               "inactive-oc-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "harness frequently triggered",
		PatternKey:       "moai_subcommand:/moai plan:ctx001",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}

	if err := a.Apply(proposal, approvedEvaluator(), snapshotBase, []Session{}); err != nil {
		t.Fatalf("gate-inactive Apply must succeed: %v", err)
	}

	if _, err := os.Stat(logPath); err == nil {
		// File exists → it must contain zero apply_outcome events.
		for _, line := range readAllJSONL(t, logPath) {
			var evt Event
			_ = json.Unmarshal(line, &evt)
			if evt.EventType == EventTypeApplyOutcome {
				t.Errorf("gate-inactive path emitted an apply_outcome event (DD-4 violation)")
			}
		}
	}
}

// TestApply_Outcome_NilObserver_NoOp verifies an active gate with a nil outcome
// observer is a safe no-op — no panic, the gate's keep decision is unchanged
// (AC-OC-007, DD-2).
func TestApply_Outcome_NilObserver_NoOp(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	skillPath := writeLineageFixture(t)
	snapshotBase := filepath.Join(dir, "snapshots")
	manifestPath := filepath.Join(dir, "learning-history", "manifest.jsonl")
	baselinePath := filepath.Join(dir, "harness", "measurements-baseline.yaml")

	m := &seqMeasurer{calls: []measureCall{
		{triple: MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0}},
		{triple: MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0}},
	}}
	a := newGateApplier(manifestPath, baselinePath, m)
	// outcomeObserver left nil intentionally.

	proposal := Proposal{
		ID:               "nilobs-oc-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "harness frequently triggered",
		PatternKey:       "moai_subcommand:/moai plan:ctx001",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}

	if err := a.Apply(proposal, approvedEvaluator(), snapshotBase, []Session{}); err != nil {
		t.Fatalf("nil-observer Apply must succeed (no-op recordOutcome): %v", err)
	}
}

// TestApply_Outcome_RecordError_DoesNotFlipVerdict verifies an outcome observer
// write failure does NOT flip the decided verdict — a kept Apply stays kept
// (change not undone) and the observer-write failure surfaces as a wrapped error
// (AC-OC-008, REQ-OC-007).
func TestApply_Outcome_RecordError_DoesNotFlipVerdict(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	skillPath := writeLineageFixture(t)
	originalContent, _ := os.ReadFile(skillPath)
	snapshotBase := filepath.Join(dir, "snapshots")
	manifestPath := filepath.Join(dir, "learning-history", "manifest.jsonl")
	baselinePath := filepath.Join(dir, "harness", "measurements-baseline.yaml")

	m := &seqMeasurer{calls: []measureCall{
		{triple: MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0}},
		{triple: MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0}},
	}}
	a := newGateApplier(manifestPath, baselinePath, m)
	// Point the observer at a path whose parent is a regular FILE, so mkdir +
	// open fail → RecordOutcome returns an I/O error.
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	a.outcomeObserver = NewObserver(filepath.Join(blocker, "usage-log.jsonl"))

	proposal := Proposal{
		ID:               "recerr-oc-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "harness frequently triggered",
		PatternKey:       "moai_subcommand:/moai plan:ctx001",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}

	err := a.Apply(proposal, approvedEvaluator(), snapshotBase, []Session{})
	if err == nil {
		t.Fatal("expected a wrapped outcome-write error to surface")
	}
	// The kept change must NOT be undone (verdict not flipped to rolled-back):
	// the file still contains the enriched description.
	after, _ := os.ReadFile(skillPath)
	if bytes.Equal(originalContent, after) {
		t.Errorf("kept change was undone — verdict flipped (REQ-OC-007 violation)")
	}
	// And it must NOT be an ApplyRegressionError (the decision was kept).
	var regErr *ApplyRegressionError
	if errors.As(err, &regErr) {
		t.Errorf("outcome-write error must not surface as ApplyRegressionError")
	}
}

// ─────────────────────────────────────────────
// M3/M4 — correlation key + reader tolerance (AC-OC-009, AC-OC-006)
// ─────────────────────────────────────────────

// TestApply_Outcome_ProposalIDMatchesLineage verifies the apply_outcome event's
// outcome_proposal_id matches the lineage entry's proposal_id for the same Apply,
// proving correlation-key reuse (AC-OC-009, DD-3).
func TestApply_Outcome_ProposalIDMatchesLineage(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	skillPath := writeLineageFixture(t)
	snapshotBase := filepath.Join(dir, "snapshots")
	manifestPath := filepath.Join(dir, "learning-history", "manifest.jsonl")
	baselinePath := filepath.Join(dir, "harness", "measurements-baseline.yaml")
	logPath := filepath.Join(dir, "harness", "usage-log.jsonl")

	m := &seqMeasurer{calls: []measureCall{
		{triple: MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0}},
		{triple: MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0}},
	}}
	a := newGateApplier(manifestPath, baselinePath, m)
	a.outcomeObserver = NewObserver(logPath)

	proposal := Proposal{
		ID:               "corr-oc-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "harness frequently triggered",
		PatternKey:       "moai_subcommand:/moai plan:ctx001",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}
	if err := a.Apply(proposal, approvedEvaluator(), snapshotBase, []Session{}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	evt := readSingleOutcomeEvent(t, logPath)
	entries, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("lineage entries = %d, want 1", len(entries))
	}
	if evt.OutcomeProposalID != entries[0].ProposalID {
		t.Errorf("outcome_proposal_id = %q, lineage proposal_id = %q; want equal", evt.OutcomeProposalID, entries[0].ProposalID)
	}
}

// TestApplyOutcome_ReaderTolerance verifies the existing AggregatePatterns reader
// processes a usage-log containing apply_outcome events (with the new omitempty
// fields) without error — additive fields + unknown event type tolerated
// (AC-OC-006, REQ-OC-009, EC-5).
func TestApplyOutcome_ReaderTolerance(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	obs := NewObserver(logPath)

	// A legacy event + an apply_outcome event in the same log (mixed log, EC-5).
	if err := obs.RecordEvent(EventTypeMoaiSubcommand, "/moai plan", "ctx-1"); err != nil {
		t.Fatalf("RecordEvent: %v", err)
	}
	if err := obs.RecordOutcome(OutcomeRecord{
		Verdict:    "kept",
		Decision:   "approved",
		ProposalID: "p-tol-001",
		Baseline:   MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0},
		Candidate:  MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0},
	}); err != nil {
		t.Fatalf("RecordOutcome: %v", err)
	}

	patterns, err := AggregatePatterns(logPath)
	if err != nil {
		t.Fatalf("AggregatePatterns must tolerate apply_outcome events: %v", err)
	}
	// Both events are aggregated (the unknown event type is grouped by key, not fatal).
	if len(patterns) != 2 {
		t.Errorf("pattern count = %d, want 2 (legacy + apply_outcome both tolerated)", len(patterns))
	}
}

// ─────────────────────────────────────────────
// test helpers
// ─────────────────────────────────────────────

// readAllJSONL reads every non-empty JSONL line from path.
func readAllJSONL(t *testing.T, path string) [][]byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var lines [][]byte
	for _, l := range bytes.Split(data, []byte("\n")) {
		if len(bytes.TrimSpace(l)) == 0 {
			continue
		}
		lines = append(lines, l)
	}
	return lines
}

// readSingleJSONL reads exactly one JSONL line, failing if the count is not 1.
func readSingleJSONL(t *testing.T, path string) []byte {
	t.Helper()
	lines := readAllJSONL(t, path)
	if len(lines) != 1 {
		t.Fatalf("line count = %d, want 1", len(lines))
	}
	return lines[0]
}

// readSingleOutcomeEvent reads the log and returns the single apply_outcome event,
// failing if there is not exactly one.
func readSingleOutcomeEvent(t *testing.T, path string) Event {
	t.Helper()
	var outcomes []Event
	for _, line := range readAllJSONL(t, path) {
		var evt Event
		if err := json.Unmarshal(line, &evt); err != nil {
			t.Fatalf("unmarshal %q: %v", line, err)
		}
		if evt.EventType == EventTypeApplyOutcome {
			outcomes = append(outcomes, evt)
		}
	}
	if len(outcomes) != 1 {
		t.Fatalf("apply_outcome event count = %d, want 1", len(outcomes))
	}
	return outcomes[0]
}
