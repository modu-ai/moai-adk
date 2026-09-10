package guardstate

import (
	"errors"
	"testing"
)

// AC-GSM-002 (a) — every entry carries its five fields.
// AC-GSM-002 (b) — an entry without a measured quantity, and separately one
// without a kind, is REJECTED WITH A NAMED ERROR rather than defaulted.
//
// Mutant this kills: a reader that silently defaults a missing measured
// quantity to `fired-at-all`, which converts a skipped-only subject into a
// green. Clause (a) alone does not kill it; the named-error assertion does.
func TestEntryValidate_RejectsMissingFieldsByName(t *testing.T) {
	complete := func() Entry {
		return Entry{
			Kind:    "github-workflow",
			Locator: ".github/workflows/ci.yml",
			Events:  []string{"push", "pull_request"},
			Window:  "7d",
			Measure: MeasureVerdictRendered,
		}
	}

	if err := complete().Validate(); err != nil {
		t.Fatalf("complete entry rejected: %v", err)
	}

	cases := []struct {
		name    string
		mutate  func(*Entry)
		wantErr error
	}{
		{"missing kind", func(e *Entry) { e.Kind = "" }, ErrMissingKind},
		{"missing locator", func(e *Entry) { e.Locator = "" }, ErrMissingLocator},
		{"missing events", func(e *Entry) { e.Events = nil }, ErrMissingEvents},
		{"missing window", func(e *Entry) { e.Window = "" }, ErrMissingWindow},
		{"missing measure", func(e *Entry) { e.Measure = "" }, ErrMissingMeasure},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := complete()
			tc.mutate(&e)
			err := e.Validate()
			if err == nil {
				t.Fatalf("%s was accepted; a defaulting reader is the mutant this kills", tc.name)
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("%s rejected with %v, want a named %v", tc.name, err, tc.wantErr)
			}
		})
	}
}

// REQ-GSM-002 — the window is a field the rows READ, so a window that cannot be
// interpreted is not a shipped field. It is refused for the same reason a
// missing one is: rows 3 and 4 cannot decide on it either way.
func TestEntryValidate_RejectsUninterpretableWindow(t *testing.T) {
	for _, w := range []string{"soon", "0d", "-7d", "7", "7w", "d7", "7 d"} {
		e := Entry{Kind: "github-workflow", Locator: "x", Events: []string{"push"}, Window: w, Measure: MeasureFiredAtAll}
		if err := e.Validate(); !errors.Is(err, ErrInvalidWindow) {
			t.Errorf("window %q validated with %v, want a named %v", w, err, ErrInvalidWindow)
		}
	}
	for _, w := range []string{"3d", "90d", "12h"} {
		e := Entry{Kind: "github-workflow", Locator: "x", Events: []string{"push"}, Window: w, Measure: MeasureFiredAtAll}
		if err := e.Validate(); err != nil {
			t.Errorf("window %q rejected: %v", w, err)
		}
	}
}

// AC-GSM-003 (a) — the three named measured-quantity values are accepted.
// AC-GSM-003 (b) — any fourth value is rejected with a named error.
//
// Mutant this kills: the rule-pairing corollary — an accept-only assertion
// passes an all-accepting reader, a reject-only assertion passes a
// nothing-accepting one. Both directions are asserted here.
func TestMeasureVocabulary_ClosedAtThreeValues(t *testing.T) {
	accepted := []Measure{MeasureFiredAtAll, MeasureFiredWithEffect, MeasureVerdictRendered}
	if len(accepted) != 3 {
		t.Fatalf("vocabulary is %d values, REQ-GSM-003 fixes it at 3", len(accepted))
	}
	for _, m := range accepted {
		e := Entry{Kind: "github-workflow", Locator: "x", Events: []string{"push"}, Window: "7d", Measure: m}
		if err := e.Validate(); err != nil {
			t.Fatalf("named value %q rejected: %v", m, err)
		}
	}
	for _, fourth := range []Measure{"fired-and-caught-something", "fired-at-All", "ok", "green"} {
		e := Entry{Kind: "github-workflow", Locator: "x", Events: []string{"push"}, Window: "7d", Measure: fourth}
		err := e.Validate()
		if err == nil {
			t.Fatalf("fourth value %q accepted; the vocabulary is not closed", fourth)
		}
		if !errors.Is(err, ErrUnknownMeasure) {
			t.Fatalf("fourth value %q rejected with %v, want a named %v", fourth, err, ErrUnknownMeasure)
		}
	}
}

// REQ-GSM-003 — each value names the conclusion set it admits, and the three
// sets differ. The behavioural check over a recorded run fixture is
// AC-GSM-003 (c) and belongs to M2; what M1 owns is that the three sets are
// DATA and are distinct.
//
// Mutant this kills: a vocabulary of three values that all admit the same
// conclusions, which makes REQ-GSM-003's one-number-two-events prohibition
// unenforceable — an author declares `verdict-rendered` and receives
// `fired-at-all` semantics.
func TestMeasureAdmits_ThreeDistinctConclusionSets(t *testing.T) {
	conclusions := []string{"success", "failure", "skipped", "cancelled"}
	got := map[Measure][]string{}
	for _, m := range []Measure{MeasureFiredAtAll, MeasureFiredWithEffect, MeasureVerdictRendered} {
		for _, c := range conclusions {
			if m.Admits(c) {
				got[m] = append(got[m], c)
			}
		}
	}
	want := map[Measure]int{
		MeasureFiredAtAll:      4,
		MeasureFiredWithEffect: 2,
		MeasureVerdictRendered: 2,
	}
	for m, n := range want {
		if len(got[m]) != n {
			t.Fatalf("%s admits %v (%d), want %d of %v", m, got[m], len(got[m]), n, conclusions)
		}
	}
	// fired-with-effect and verdict-rendered coincide over terminal conclusions;
	// they separate on a run that is neither skipped/cancelled nor terminal
	// (in progress, or a null conclusion).
	if MeasureFiredWithEffect.Admits("") == MeasureVerdictRendered.Admits("") {
		t.Fatalf("fired-with-effect and verdict-rendered admit the same set; REQ-GSM-003 requires verdict-rendered to additionally require a terminal success or failure")
	}
	if !MeasureFiredAtAll.Admits("skipped") || MeasureFiredWithEffect.Admits("skipped") {
		t.Fatalf("fired-at-all and fired-with-effect do not separate on `skipped`")
	}
	if !MeasureFiredAtAll.Admits("cancelled") || MeasureFiredWithEffect.Admits("cancelled") {
		t.Fatalf("fired-at-all and fired-with-effect do not separate on `cancelled`")
	}
}

// AC-GSM-005 (b) — the manifest holds its watched set as DATA: accepting an
// entry of a second, non-workflow kind requires NO change to the manifest
// schema, the classification vocabulary, or the measured-quantity vocabulary.
//
// Mutant this kills: a reader with a hardcoded workflow-path field that merely
// tolerates an unknown value in it. The fixture below carries a kind and a
// locator with no forge workflow behind them and is parsed by the SAME schema
// and the SAME parser the census uses — no second type, no second entry point.
func TestSecondKindEntry_AcceptedWithoutSchemaChange(t *testing.T) {
	const src = `
entries:
  - kind: policy-rule
    locator: .claude/rules/moai/core/moai-constitution.md
    events: [session-start]
    window: 30d
    measure: fired-at-all
`
	m, err := ParseManifest([]byte(src))
	if err != nil {
		t.Fatalf("second-kind entry rejected by the shared schema: %v", err)
	}
	if got := len(m.Entries); got != 1 {
		t.Fatalf("declared count is %d, want 1 — a second-kind entry must be COUNTED, not skipped", got)
	}
	e := m.Entries[0]
	if e.Kind != "policy-rule" {
		t.Fatalf("kind round-tripped as %q, want policy-rule — kind is carried as data", e.Kind)
	}
	if e.Locator != ".claude/rules/moai/core/moai-constitution.md" {
		t.Fatalf("locator round-tripped as %q — the locator field is not workflow-specific", e.Locator)
	}
}

// REQ-GSM-004 — the release-cycle-conditional form is carried as a declared
// field, so "correctly quiet" is a recorded expectation rather than an absence
// a reader must infer. Rows 5 and 6 of the state table are the two branches of
// the test that reads it.
func TestConditionalField_IsReadAndCarried(t *testing.T) {
	const src = `
entries:
  - kind: github-workflow
    locator: .github/workflows/release.yml
    events: [push]
    window: 90d
    measure: verdict-rendered
    expected_when: release-cycle
`
	m, err := ParseManifest([]byte(src))
	if err != nil {
		t.Fatalf("conditional entry rejected: %v", err)
	}
	if got := m.Entries[0].ExpectedWhen; got != "release-cycle" {
		t.Fatalf("expected_when round-tripped as %q, want release-cycle — the field must be READ, not merely declared", got)
	}
	if !m.Entries[0].IsConditional() {
		t.Fatalf("entry with expected_when does not report itself conditional")
	}
	if (Entry{}).IsConditional() {
		t.Fatalf("entry without expected_when reports itself conditional; rows 5 and 6 would not separate")
	}
}

// ParseManifest applies the same validation entry by entry, so a malformed
// entry cannot reach the classifier by arriving through the file rather than
// through a constructor.
func TestParseManifest_RejectsMalformedEntryByName(t *testing.T) {
	const src = `
entries:
  - locator: .github/workflows/ci.yml
    events: [push]
    window: 7d
    measure: fired-at-all
`
	if _, err := ParseManifest([]byte(src)); !errors.Is(err, ErrMissingKind) {
		t.Fatalf("kindless entry parsed with %v, want a named %v", err, ErrMissingKind)
	}
}
