package jevmeasure

import (
	"context"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/jev"
)

// edges_test.go — the paths a happy-path measurement never reaches. Each one
// is a way the harness could fail QUIETLY, which is the only failure mode that
// matters here: a gate that errors loudly gets fixed, and a gate that returns
// a plausible number does not.

// TestConstantBaseline_RejectsUnlabelledSample: an unlabelled sample cannot be
// scored, and silently excluding it would shift the base rate the whole gate
// is read against.
func TestConstantBaseline_RejectsUnlabelledSample(t *testing.T) {
	set := Set{Consumer: "trial", Shape: jev.KindNoul, Samples: []Sample{
		{ID: "a", Label: "alive"},
		{ID: "b", Label: "   "},
	}}
	_, err := ConstantBaseline(set)
	if err == nil {
		t.Fatal("ConstantBaseline accepted an unlabelled sample")
	}
	if !strings.Contains(err.Error(), "b") {
		t.Errorf("the error does not name the offending sample: %v", err)
	}
}

// TestConstantBaseline_TieIsDeterministic: two labels at equal count must
// resolve the same way on every run, or two measurements of the same set are
// not comparable.
func TestConstantBaseline_TieIsDeterministic(t *testing.T) {
	set := Set{Consumer: "trial", Shape: jev.KindNoul, Samples: []Sample{
		{ID: "a", Label: "zebra"}, {ID: "b", Label: "alpha"},
	}}
	first, err := ConstantBaseline(set)
	if err != nil {
		t.Fatalf("ConstantBaseline: %v", err)
	}
	for i := 0; i < 20; i++ {
		again, err := ConstantBaseline(set)
		if err != nil {
			t.Fatalf("ConstantBaseline: %v", err)
		}
		if again.Label != first.Label {
			t.Fatalf("tie resolved to %q then %q — the baseline is not deterministic", first.Label, again.Label)
		}
	}
	if first.Accuracy != 0.5 {
		t.Errorf("Accuracy = %v, want 0.5", first.Accuracy)
	}
}

// TestVerdict_NoBaselineWithholds: a report carrying no baseline has no number
// to beat, and must withhold rather than treat the absence as a clear pass.
func TestVerdict_NoBaselineWithholds(t *testing.T) {
	rep := Report{Consumer: "trial", ModelID: jev.ModelID, Source: SourceLive}
	v, reason := rep.Verdict()
	if v != VerdictWithhold {
		t.Errorf("Verdict = %q with no baseline, want %q", v, VerdictWithhold)
	}
	if !strings.Contains(reason, "baseline") {
		t.Errorf("the reason does not name the missing baseline: %q", reason)
	}
}

// TestValidate_RejectsMissingModelAndMismatchedThreshold.
func TestValidate_RejectsMissingModelAndMismatchedThreshold(t *testing.T) {
	if err := (Report{Consumer: "trial"}).Validate(); err == nil {
		t.Error("Validate accepted a report citing no model id")
	}

	th, err := NewThreshold("other-consumer", jev.KindNoul, 0.62, 20, "fitted here")
	if err != nil {
		t.Fatalf("NewThreshold: %v", err)
	}
	rep := Report{Consumer: "trial", ModelID: jev.ModelID, Shape: jev.KindNoul, Threshold: &th}
	if err := rep.Validate(); err == nil {
		t.Error("Validate accepted a threshold fitted for a different consumer")
	}

	// Positive control: the same report with a matching threshold validates,
	// so the rejection above is about the mismatch and not about thresholds.
	matching, err := NewThreshold("trial", jev.KindNoul, 0.62, 20, "fitted here")
	if err != nil {
		t.Fatalf("NewThreshold: %v", err)
	}
	rep.Threshold = &matching
	if err := rep.Validate(); err != nil {
		t.Errorf("positive control failed: a matching threshold was rejected: %v", err)
	}
}

// TestRender_CarriesThresholdAbsenceAndConditions: the rendered artifact shows
// every part a reader needs to judge it, including the degradation conditions
// an arm observed. A run that mostly failed must not render like a clean one.
func TestRender_CarriesThresholdAbsenceAndConditions(t *testing.T) {
	th, err := NewThreshold("trial", jev.KindNoul, 0.62, 20, "fitted on this repository's data")
	if err != nil {
		t.Fatalf("NewThreshold: %v", err)
	}
	rep := Report{
		Consumer: "trial", ModelID: jev.ModelID, Shape: jev.KindNoul, Source: SourceLive,
		Baseline: Baseline{Label: "alive", Correct: 15, N: 20, Accuracy: 0.75},
		Arms: []ArmResult{
			{Arm: ArmKoreanOriginal, N: 20, Answered: 18, Unavailable: 2, Correct: 16, Accuracy: 0.8,
				Conditions: []string{"jev unavailable (rate-limited)"}},
			{Arm: ArmEnglishTranslat, N: 20, Answered: 20, Correct: 17, Accuracy: 0.85},
		},
		Delta:     0.05,
		Threshold: &th,
		Absence: &Absence{
			Claim:   "no sample scored above the gate",
			Control: PositiveControl{Description: "a fixture known to exceed it", Observed: "1 sample above"},
		},
	}
	out := rep.Render()
	for _, want := range []string{
		"fitted on this repository's data",
		"no sample scored above the gate",
		"a fixture known to exceed it",
		"rate-limited",
		"unavailable",
		"+0.050",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the rendered report lacks %q:\n%s", want, out)
		}
	}
}

// TestLabelOf_ChoiceAndUnderspecifiedNoul: a Choice answer maps to its chosen
// option; a Noul over a label set too small to interpret maps to nothing,
// which matches no sample and is therefore scored wrong rather than right.
func TestLabelOf_ChoiceAndUnderspecifiedNoul(t *testing.T) {
	if got := labelOf(jev.Answer{Kind: jev.KindChoice, Choice: "lane-2"}, nil); got != "lane-2" {
		t.Errorf("Choice label = %q, want lane-2", got)
	}
	if got := labelOf(jev.Answer{Kind: jev.KindNoul, Noul: true}, []string{"only-one"}); got != "" {
		t.Errorf("an under-specified Noul mapped to %q, want the empty string", got)
	}
	if got := labelOf(jev.Answer{Kind: jev.KindScore, Score: 0.4}, []string{"a", "b"}); got != "" {
		t.Errorf("a Score mapped to %q, want the empty string", got)
	}
}

// TestSourceOf_LiveClientIsTheOnlyLiveSource: the live source is inferred from
// the concrete client type, never declared. A declared source would be the one
// field a hurried change could set to "live" to make a stub's numbers ship.
func TestSourceOf_LiveClientIsTheOnlyLiveSource(t *testing.T) {
	if got := sourceOf(jev.New(false)); got != SourceLive {
		t.Errorf("sourceOf(*jev.Client) = %q, want %q", got, SourceLive)
	}
	if got := sourceOf(stubAnswerer{}); got != SourceInjected {
		t.Errorf("sourceOf(stub) = %q, want %q", got, SourceInjected)
	}
}

// TestNewThreshold_RejectsAbsentSampleSize: a threshold with no sample size is
// a number with no error bar presented as a decision rule — the exact shape the
// rejected task's two high-gate cells had, at 5 and 4 samples.
func TestNewThreshold_RejectsAbsentSampleSize(t *testing.T) {
	if _, err := NewThreshold("trial", jev.KindNoul, 0.62, 0, "fitted"); err == nil {
		t.Error("NewThreshold accepted a threshold with no sample size")
	}
	if _, err := NewThreshold("trial", jev.KindNoul, 0.62, -1, "fitted"); err == nil {
		t.Error("NewThreshold accepted a negative sample size")
	}
}

// TestNewThreshold_VendorValueIsAllowedWithMeasuredProvenance: the prohibition
// is on ADOPTING a vendor figure, not on the number itself. A fit that
// genuinely lands there says so and passes.
func TestNewThreshold_VendorValueIsAllowedWithMeasuredProvenance(t *testing.T) {
	for _, v := range VendorIllustrativeThresholds() {
		if _, err := NewThreshold("trial", jev.KindNoul, v, 40, "fitted on 40 labelled samples in this repository"); err != nil {
			t.Errorf("a measured fit landing on %v was rejected: %v", v, err)
		}
	}
}

// TestFitThreshold_RejectsEmptyArm: an arm that produced no probabilities
// cannot be fitted on, and must say so rather than returning zero — a zero
// gate admits everything.
func TestFitThreshold_RejectsEmptyArm(t *testing.T) {
	if _, err := FitThreshold("trial", jev.KindNoul, ArmResult{Arm: ArmKoreanOriginal, N: 20}); err == nil {
		t.Error("FitThreshold fitted a gate on an arm with no probabilities")
	}
}

// TestReadNoulPair_BothMissing: neither polarity present is its own error, and
// must not be reported as one of them being derivable from the other.
func TestReadNoulPair_BothMissing(t *testing.T) {
	_, err := ReadNoulPair(nil, "a", "b")
	if err == nil {
		t.Fatal("ReadNoulPair succeeded with no answers at all")
	}
	if strings.Contains(err.Error(), "derivable") {
		t.Errorf("the both-missing error talks about derivation: %v", err)
	}

	// The single-missing case names derivation explicitly, so a reader learns
	// WHY the absent value was not filled in.
	_, err = ReadNoulPair([]jev.Answer{{QuestionID: "a", Kind: jev.KindNoul, Probability: 0.6}}, "a", "b")
	if err == nil || !strings.Contains(err.Error(), "derivable") {
		t.Errorf("the single-missing error does not state non-derivability: %v", err)
	}
}

// TestState_NamesAndReSetPreserveOrder: a builder that conditionally overrides
// a field must not reorder the payload, or two runs over the same data
// serialize differently and stop being comparable.
func TestState_NamesAndReSetPreserveOrder(t *testing.T) {
	st := NewState().Set("a", 1).Set("b", 2).Set("a", 3)
	got := strings.Join(st.Names(), ",")
	if got != "a,b" {
		t.Errorf("Names = %q, want \"a,b\"", got)
	}
	payload, err := st.JSON()
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	if payload != `{"a":3,"b":2}` {
		t.Errorf("JSON = %s, want {\"a\":3,\"b\":2}", payload)
	}
}

// TestBuildRequest_RejectsMalformedQuestions covers the shapes that would
// otherwise reach the wire as an unanswerable request.
func TestBuildRequest_RejectsMalformedQuestions(t *testing.T) {
	st := NewState().Set("card_id", "t1020")
	cases := []struct {
		name string
		qs   []Question
	}{
		{"no questions", nil},
		{"empty id", []Question{{ID: "", Text: "Does it hold?", Kind: jev.KindNoul, Reads: []string{"card_id"}}}},
		{"empty text", []Question{{ID: "q", Text: "  ", Kind: jev.KindNoul, Reads: []string{"card_id"}}}},
		{"reads nothing", []Question{{ID: "q", Text: "Does it hold?", Kind: jev.KindNoul}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := BuildRequest(st, tc.qs); err == nil {
				t.Error("BuildRequest accepted a malformed question set")
			}
		})
	}
	if _, err := BuildRequest(nil, []Question{{ID: "q", Text: "x", Kind: jev.KindNoul, Reads: []string{"y"}}}); err == nil {
		t.Error("BuildRequest accepted a nil state")
	}
}

// TestRun_RejectsAbsentBuilder: a nil builder would otherwise panic partway
// through a measurement, leaving a half-run whose numbers look real.
func TestRun_RejectsAbsentBuilder(t *testing.T) {
	if _, err := Run(context.Background(), stubAnswerer{}, trialSet(), nil); err == nil {
		t.Error("Run proceeded with no request builder")
	}
}
