package jevmeasure

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/jev"
)

// measure_test.go — the measurement harness (SPEC-JEV-OPTIN-MEASURE-001
// REQ-JEVO-008..015; AC-JEVO-010..014).
//
// [HARD] No test in this package contacts api.typesafe.ai. Every measurement
// below runs against an injected answerer, and every Report it produces
// carries SourceInjected — which the verdict refuses to ship on. That refusal
// is the point: a harness that could render a shipping verdict from a fake
// client would be a harness that makes a fabricated number actionable.

// stubAnswerer answers from a fixed table keyed by the sample id carried in
// the state. Absent entries return an unavailable result, so the harness's
// handling of a missing answer is exercised rather than assumed.
type stubAnswerer struct {
	answers map[string]bool
	probs   map[string]float64
	missing map[string]bool
}

func (s stubAnswerer) Ask(_ context.Context, req jev.Request) jev.Result {
	id := extractStubID(req.State)
	if s.missing[id] {
		return jev.Result{Availability: jev.Unreachable, Condition: "stub: no answer for " + id}
	}
	p := 0.5
	if v, ok := s.probs[id]; ok {
		p = v
	}
	// A Noul answer is ONE probability, so the fixture's decision and its
	// probability are folded into that number: the probability is mirrored
	// across 0.5 when it disagrees with the decision the fixture wants read.
	if s.answers[id] && p <= noulTrueAbove {
		p = 1 - p
		if p <= noulTrueAbove {
			p = 0.75
		}
	} else if !s.answers[id] && p > noulTrueAbove {
		p = 1 - p
	}
	return jev.Result{
		Availability: jev.Available,
		Answers: []jev.Answer{{
			QuestionID: req.Questions[0].ID, Kind: jev.KindNoul,
			Probability: p,
		}},
		Usage: jev.Usage{Model: "jev-9.8.7", RequestModel: jev.ModelID},
	}
}

// extractStubID pulls the sample id out of the serialized state without a JSON
// dependency in the stub: the builder always writes it as a plain string field.
func extractStubID(state string) string {
	const key = `"sample_id":"`
	i := strings.Index(state, key)
	if i < 0 {
		return ""
	}
	rest := state[i+len(key):]
	j := strings.Index(rest, `"`)
	if j < 0 {
		return ""
	}
	return rest[:j]
}

// modelAnswerer answers exactly as its inner stub does, then sets the interpreted
// model the response reports, chosen per request. It is the seam the straddle
// tests use: the same answers, carrying the model ids the test names.
type modelAnswerer struct {
	inner   stubAnswerer
	modelOf func(req jev.Request) string
}

func (m modelAnswerer) Ask(ctx context.Context, req jev.Request) jev.Result {
	res := m.inner.Ask(ctx, req)
	if res.OK() {
		res.Usage.Model = m.modelOf(req)
	}
	return res
}

// trialSet builds a labelled set whose majority label is "alive" at 75% — the
// same base rate the rejected premise-death task carried, so the harness is
// exercised on the shape that already misled this repository once.
func trialSet() Set {
	set := Set{Consumer: "trial", Shape: jev.KindNoul, Labels: []string{"dead", "alive"}}
	for i := 0; i < 20; i++ {
		label := "alive"
		if i < 5 {
			label = "dead"
		}
		id := string(rune('a' + i))
		set.Samples = append(set.Samples, Sample{
			ID:    id,
			Label: label,
			Text: map[Arm]string{
				ArmKoreanOriginal:  "카드 본문 " + id,
				ArmEnglishTranslat: "card body " + id,
			},
		})
	}
	return set
}

// trialBuild is the request builder the trial consumer uses. It obeys the
// question-design constraints: the state carries only fields the question
// reads, and nothing is computed by the model.
func trialBuild(s Sample, arm Arm) (jev.Request, error) {
	st := NewState().
		Set("sample_id", s.ID).
		Set("card_text", s.Text[arm])
	return BuildRequest(st, []Question{{
		ID:    "premise_dead",
		Text:  "Reading card_text only, does the card state a premise that no longer holds?",
		Kind:  jev.KindNoul,
		Reads: []string{"sample_id", "card_text"},
	}})
}

// TestConstantBaseline_IsTheMajorityLabel is REQ-JEVO-008's core: the number
// to beat is the accuracy of always answering the majority label, not zero.
func TestConstantBaseline_IsTheMajorityLabel(t *testing.T) {
	b, err := ConstantBaseline(trialSet())
	if err != nil {
		t.Fatalf("ConstantBaseline: %v", err)
	}
	if b.Label != "alive" {
		t.Errorf("Label = %q, want \"alive\"", b.Label)
	}
	if b.N != 20 || b.Correct != 15 {
		t.Errorf("Correct/N = %d/%d, want 15/20", b.Correct, b.N)
	}
	if b.Accuracy != 0.75 {
		t.Errorf("Accuracy = %v, want 0.75", b.Accuracy)
	}

	// An empty set has no baseline, and must say so rather than returning a
	// zero that reads as "0% to beat" — the most dangerous possible default.
	if _, err := ConstantBaseline(Set{Consumer: "empty", Shape: jev.KindNoul}); err == nil {
		t.Error("ConstantBaseline returned a baseline for an empty set")
	}
}

// TestRun_ProducesBothArmsAndDelta is REQ-JEVO-010 / AC-JEVO-010: every
// measurement runs both language arms and records the per-arm result and the
// delta, under the interpreted model id (REQ-JRM-006).
func TestRun_ProducesBothArmsAndDelta(t *testing.T) {
	set := trialSet()
	// Answer every sample "alive" (noul=false for "dead"), which reproduces
	// the constant answer exactly — so the model scores the baseline and
	// beats nothing.
	stub := stubAnswerer{answers: map[string]bool{}, probs: map[string]float64{}}

	rep, err := Run(context.Background(), stub, set, trialBuild)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if rep.ModelID != "jev-9.8.7" {
		t.Errorf("ModelID = %q, want the interpreted %q", rep.ModelID, "jev-9.8.7")
	}
	if rep.RequestModel != jev.ModelID {
		t.Errorf("RequestModel = %q, want the request alias %q", rep.RequestModel, jev.ModelID)
	}
	if len(rep.Arms) != 2 {
		t.Fatalf("arms = %d, want 2", len(rep.Arms))
	}
	if rep.Arms[0].Arm != ArmKoreanOriginal || rep.Arms[1].Arm != ArmEnglishTranslat {
		t.Errorf("arms = %v / %v, want the Korean original then the English translation",
			rep.Arms[0].Arm, rep.Arms[1].Arm)
	}
	if rep.Delta != rep.Arms[1].Accuracy-rep.Arms[0].Accuracy {
		t.Errorf("Delta = %v, want the English-minus-Korean difference", rep.Delta)
	}
	if rep.Baseline.Accuracy != 0.75 {
		t.Errorf("Baseline.Accuracy = %v, want 0.75", rep.Baseline.Accuracy)
	}
	if rep.Source != SourceInjected {
		t.Errorf("Source = %q, want %q — the answerer was not the live client", rep.Source, SourceInjected)
	}
}

// TestRun_RefusesAStraddledMeasurement is AC-JRM-006 (REQ-JRM-006): when the
// answered calls of one run report two interpreted model ids, the run returns
// no artifact, and its error names both ids.
func TestRun_RefusesAStraddledMeasurement(t *testing.T) {
	set := trialSet()
	// Sample "b" is answered by jev-9.8.8; every other call by jev-9.8.7.
	straddled := modelAnswerer{
		inner: stubAnswerer{answers: map[string]bool{}, probs: map[string]float64{}},
		modelOf: func(req jev.Request) string {
			if extractStubID(req.State) == "b" {
				return "jev-9.8.8"
			}
			return "jev-9.8.7"
		},
	}
	rep, err := Run(context.Background(), straddled, set, trialBuild)
	if err == nil {
		t.Fatal("Run returned an artifact for a run whose answered calls report two model versions")
	}
	if !reflect.DeepEqual(rep, Report{}) {
		t.Errorf("a refused run returned a non-empty artifact: %+v", rep)
	}
	first := strings.Index(err.Error(), "jev-9.8.7")
	second := strings.Index(err.Error(), "jev-9.8.8")
	if first < 0 || second < 0 {
		t.Fatalf("the error does not name both interpreted model ids: %v", err)
	}
	if first > second {
		t.Errorf("the error names the model ids out of sorted order: %v", err)
	}

	// Positive control: the same run with one id throughout is measured, so the
	// refusal above is about the straddle and not about the fixture.
	uniform := modelAnswerer{
		inner:   stubAnswerer{answers: map[string]bool{}, probs: map[string]float64{}},
		modelOf: func(jev.Request) string { return "jev-9.8.7" },
	}
	if _, err := Run(context.Background(), uniform, set, trialBuild); err != nil {
		t.Errorf("positive control failed: a run with one model id was refused: %v", err)
	}

	// An OK call with zero answers still reports an interpreted model, and that
	// model counts toward the straddle (REQ-JRM-006): the run mixes jev-9.8.8
	// (answered) with jev-9.8.7 (OK, no answers) and must be refused.
	t.Run("ok call with no answers counts toward the straddle", func(t *testing.T) {
		mixed := answerlessAnswerer{
			inner: modelAnswerer{
				inner: stubAnswerer{answers: map[string]bool{}, probs: map[string]float64{}},
				modelOf: func(req jev.Request) string {
					if extractStubID(req.State) == "b" {
						return "jev-9.8.7"
					}
					return "jev-9.8.8"
				},
			},
			empty: func(req jev.Request) bool { return extractStubID(req.State) == "b" },
		}
		rep, err := Run(context.Background(), mixed, set, trialBuild)
		if err == nil {
			t.Fatalf("Run returned an artifact citing %q for a run that also saw an OK call on another model", rep.ModelID)
		}
		if !reflect.DeepEqual(rep, Report{}) {
			t.Errorf("a refused run returned a non-empty artifact: %+v", rep)
		}
		first := strings.Index(err.Error(), "jev-9.8.7")
		second := strings.Index(err.Error(), "jev-9.8.8")
		if first < 0 || second < 0 {
			t.Fatalf("the error does not name both interpreted model ids: %v", err)
		}
		if first > second {
			t.Errorf("the error names the model ids out of sorted order: %v", err)
		}
	})

	// Control: an OK call with no answers whose interpreted id is not valid
	// (empty) never enters the model set, so the run is measured on one id.
	t.Run("ok call with no answers and an empty id stays out of the model set", func(t *testing.T) {
		control := answerlessAnswerer{
			inner: modelAnswerer{
				inner: stubAnswerer{answers: map[string]bool{}, probs: map[string]float64{}},
				modelOf: func(req jev.Request) string {
					if extractStubID(req.State) == "b" {
						return ""
					}
					return "jev-9.8.7"
				},
			},
			empty: func(req jev.Request) bool { return extractStubID(req.State) == "b" },
		}
		rep, err := Run(context.Background(), control, set, trialBuild)
		if err != nil {
			t.Fatalf("a run with answered calls on one id was refused: %v", err)
		}
		if rep.ModelID != "jev-9.8.7" {
			t.Errorf("ModelID = %q, want %q", rep.ModelID, "jev-9.8.7")
		}
	})
}

// answerlessAnswerer wraps an answerer and strips the answers from the OK
// results of the requests selected by empty: an OK call with zero answers.
type answerlessAnswerer struct {
	inner modelAnswerer
	empty func(req jev.Request) bool
}

func (a answerlessAnswerer) Ask(ctx context.Context, req jev.Request) jev.Result {
	res := a.inner.Ask(ctx, req)
	if res.OK() && a.empty(req) {
		res.Answers = nil
	}
	return res
}

// TestRun_CitesTheInterpretedModelAcrossArms: the interpreted model is taken from
// the answers of both language arms, not from the first arm alone. When both
// arms report one version the artifact cites it, with the request alias in its
// own field. When only the English arm reports another version, the run is
// refused: a citation that read the Korean arm alone would drop that version.
func TestRun_CitesTheInterpretedModelAcrossArms(t *testing.T) {
	set := trialSet()
	agreeing := stubAnswerer{answers: map[string]bool{}, probs: map[string]float64{}}

	rep, err := Run(context.Background(), agreeing, set, trialBuild)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, a := range rep.Arms {
		if a.Answered != len(set.Samples) {
			t.Errorf("arm %s answered %d of %d: the citation must rest on both arms",
				a.Arm, a.Answered, len(set.Samples))
		}
	}
	if rep.ModelID != "jev-9.8.7" {
		t.Errorf("ModelID = %q, want the interpreted %q cited across both arms", rep.ModelID, "jev-9.8.7")
	}
	if rep.RequestModel != jev.ModelID {
		t.Errorf("RequestModel = %q, want the request alias %q", rep.RequestModel, jev.ModelID)
	}

	englishDrift := modelAnswerer{
		inner: stubAnswerer{answers: map[string]bool{}, probs: map[string]float64{}},
		modelOf: func(req jev.Request) string {
			if strings.Contains(req.State, "card body") {
				return "jev-9.8.8"
			}
			return "jev-9.8.7"
		},
	}
	if _, err := Run(context.Background(), englishDrift, set, trialBuild); err == nil {
		t.Error("Run cited one model although the English arm reported another")
	}
}

// TestVerdict_WithholdsWhenBaselineNotBeaten is REQ-JEVO-009: the verdict is
// binding, not advisory. A consumer that only matches its constant baseline
// is not shipped.
//
// The report is constructed with Source: SourceLive deliberately. Running the
// stub and asserting on the result would NOT exercise this gate: the
// injected-source guard returns first, so the baseline comparison is never
// reached. That was measured — mutating the baseline comparison to `if false`
// left the stub-driven version of this test passing, which is exactly the
// vacuous green this file exists to prevent. Constructing the report is what
// puts the gate under the assertion.
func TestVerdict_WithholdsWhenBaselineNotBeaten(t *testing.T) {
	cases := []struct {
		name     string
		bestArm  float64
		baseline float64
		want     Verdict
	}{
		{"below the constant", 0.60, 0.75, VerdictWithhold},
		{"equal to the constant", 0.75, 0.75, VerdictWithhold},
		{"above the constant", 0.80, 0.75, VerdictShip},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rep := Report{
				Consumer: "trial",
				ModelID:  "jev-9.8.7",
				Shape:    jev.KindNoul,
				Source:   SourceLive,
				Baseline: Baseline{Label: "alive", Correct: 15, N: 20, Accuracy: tc.baseline},
				Arms: []ArmResult{
					{Arm: ArmKoreanOriginal, N: 20, Accuracy: tc.bestArm},
					{Arm: ArmEnglishTranslat, N: 20, Accuracy: tc.bestArm - 0.05},
				},
			}
			v, reason := rep.Verdict()
			if v != tc.want {
				t.Errorf("Verdict = %q, want %q (best arm %v vs baseline %v)",
					v, tc.want, tc.bestArm, tc.baseline)
			}
			if reason == "" {
				t.Error("the verdict carries no reason")
			}
			if tc.want == VerdictWithhold && !strings.Contains(reason, "baseline") {
				t.Errorf("a withholding verdict does not name the baseline: %q", reason)
			}
		})
	}
}

// TestVerdict_EqualAccuracyDoesNotShip pins the boundary explicitly: REQ-JEVO-009
// says a consumer whose accuracy does not EXCEED its baseline is not shipped,
// so equality withholds. This is the case a ">=" typo would silently flip.
func TestVerdict_EqualAccuracyDoesNotShip(t *testing.T) {
	rep := Report{
		Consumer: "trial", ModelID: "jev-9.8.7", Shape: jev.KindNoul, Source: SourceLive,
		Baseline: Baseline{Label: "alive", Correct: 15, N: 20, Accuracy: 0.75},
		Arms:     []ArmResult{{Arm: ArmKoreanOriginal, N: 20, Accuracy: 0.75}},
	}
	if v, _ := rep.Verdict(); v != VerdictWithhold {
		t.Errorf("an arm exactly matching its baseline shipped (%q)", v)
	}
}

// TestVerdict_InjectedSourceNeverShips: even a perfect score from an injected
// answerer withholds. A fabricated figure must not become actionable.
func TestVerdict_InjectedSourceNeverShips(t *testing.T) {
	set := trialSet()
	// Answer every sample correctly.
	answers := map[string]bool{}
	for _, s := range set.Samples {
		answers[s.ID] = s.Label == "dead"
	}
	rep, err := Run(context.Background(), stubAnswerer{answers: answers, probs: map[string]float64{}}, set, trialBuild)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if rep.BestArm().Accuracy != 1.0 {
		t.Fatalf("positive control failed: the stub did not score 100%% (got %v)", rep.BestArm().Accuracy)
	}
	v, reason := rep.Verdict()
	if v != VerdictWithhold {
		t.Errorf("Verdict = %q on a perfect INJECTED run, want %q", v, VerdictWithhold)
	}
	if !strings.Contains(strings.ToLower(reason), "injected") {
		t.Errorf("the reason does not name the injected source: %q", reason)
	}
}

// TestRun_UnavailableAnswersAreRecordedNotCountedCorrect: an absent answer is
// never read as a correct one, and the arm records how many were missing.
func TestRun_UnavailableAnswersAreRecordedNotCountedCorrect(t *testing.T) {
	set := trialSet()
	answers := map[string]bool{}
	for _, s := range set.Samples {
		answers[s.ID] = s.Label == "dead"
	}
	missing := map[string]bool{set.Samples[0].ID: true}
	rep, err := Run(context.Background(),
		stubAnswerer{answers: answers, probs: map[string]float64{}, missing: missing}, set, trialBuild)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	arm := rep.Arms[0]
	if arm.Unavailable != 1 {
		t.Errorf("Unavailable = %d, want 1", arm.Unavailable)
	}
	if arm.Answered != len(set.Samples)-1 {
		t.Errorf("Answered = %d, want %d", arm.Answered, len(set.Samples)-1)
	}
	if arm.Correct != len(set.Samples)-1 {
		t.Errorf("Correct = %d, want %d (the missing sample must not count correct)",
			arm.Correct, len(set.Samples)-1)
	}
}

// TestThreshold_FittedCarriesSampleSizeAndProvenance is REQ-JEVO-012: a fitted
// threshold records the sample size it was fitted on. A number with no error
// bar presented as a decision rule is the failure the rejected task recorded,
// where two high-gate cells rested on 5 and 4 samples.
func TestThreshold_FittedCarriesSampleSizeAndProvenance(t *testing.T) {
	set := trialSet()
	answers := map[string]bool{}
	probs := map[string]float64{}
	for i, s := range set.Samples {
		answers[s.ID] = s.Label == "dead"
		probs[s.ID] = 0.4 + float64(i)*0.02
	}
	rep, err := Run(context.Background(), stubAnswerer{answers: answers, probs: probs}, set, trialBuild)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	th, err := FitThreshold(set.Consumer, set.Shape, rep.BestArm())
	if err != nil {
		t.Fatalf("FitThreshold: %v", err)
	}
	if th.SampleSize != 20 {
		t.Errorf("SampleSize = %d, want 20", th.SampleSize)
	}
	if th.Provenance == "" {
		t.Error("a fitted threshold carries no provenance")
	}
	if th.Consumer != "trial" || th.Shape != jev.KindNoul {
		t.Errorf("threshold identity = %q/%v, want trial/noul", th.Consumer, th.Shape)
	}
}

// TestThreshold_RejectsVendorCopyAndCrossShapeTransfer is REQ-JEVO-013: a
// threshold is not copied from vendor documentation, and is not transferred
// between a Noul and a Choice.
func TestThreshold_RejectsVendorCopyAndCrossShapeTransfer(t *testing.T) {
	for _, v := range VendorIllustrativeThresholds() {
		if _, err := NewThreshold("trial", jev.KindNoul, v, 20, ""); err == nil {
			t.Errorf("NewThreshold accepted the vendor figure %v with no measured provenance", v)
		}
	}

	th, err := NewThreshold("trial", jev.KindNoul, 0.62, 20, "fitted on 20 labelled cards in this repository")
	if err != nil {
		t.Fatalf("NewThreshold: %v", err)
	}
	if err := th.ApplicableTo("trial", jev.KindNoul); err != nil {
		t.Errorf("a threshold rejected its own consumer and shape: %v", err)
	}
	if err := th.ApplicableTo("trial", jev.KindChoice); err == nil {
		t.Error("a Noul threshold was accepted for a Choice question")
	}
	if err := th.ApplicableTo("other", jev.KindNoul); err == nil {
		t.Error("a threshold was accepted for a different consumer")
	}
}

// TestReport_AbsenceRequiresPositiveControl is REQ-JEVO-015 / AC-JEVO-014: a
// result reporting an absence is accompanied by a positive control showing the
// apparatus fires on that path. The requirement is built into the artifact's
// shape so a caller cannot report an absence without one.
func TestReport_AbsenceRequiresPositiveControl(t *testing.T) {
	set := trialSet()
	rep, err := Run(context.Background(), stubAnswerer{answers: map[string]bool{}, probs: map[string]float64{}}, set, trialBuild)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	rep.Absence = &Absence{Claim: "no sample scored above the gate"}
	if err := rep.Validate(); err == nil {
		t.Error("Validate accepted an absence claim with no positive control")
	}

	rep.Absence.Control = PositiveControl{
		Description: "the same gate applied to a fixture known to exceed it",
		Observed:    "1 sample above the gate",
	}
	if err := rep.Validate(); err != nil {
		t.Errorf("Validate rejected an absence claim carrying a control: %v", err)
	}

	// A control with an empty observation is not a control: it asserts the
	// apparatus fired without saying what it produced.
	rep.Absence.Control.Observed = ""
	if err := rep.Validate(); err == nil {
		t.Error("Validate accepted a positive control with no observation")
	}
}

// TestReport_RenderCitesInterpretedModelAndAliasOnlyInRequestField is AC-JRM-007
// (REQ-JRM-007): the rendered artifact names the interpreted model on its model
// line and the request alias on its request-model line, and the alias appears on
// no other line. The positive control is a rendering whose model line is the
// alias, which the same check must refuse.
func TestReport_RenderCitesInterpretedModelAndAliasOnlyInRequestField(t *testing.T) {
	set := trialSet()
	rep, err := Run(context.Background(), stubAnswerer{answers: map[string]bool{}, probs: map[string]float64{}}, set, trialBuild)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	out := rep.Render()

	for _, want := range []string{
		"model: jev-9.8.7 (interpreted by the endpoint)",
		"request model: " + jev.ModelID + " (alias sent)",
	} {
		if !hasLine(out, want) {
			t.Errorf("the rendered report lacks the line %q:\n%s", want, out)
		}
	}
	if line := aliasOutsideRequestLine(out); line != "" {
		t.Errorf("the alias appears outside the request-model line: %q", line)
	}

	// Positive control: a rendering whose model line is the alias fails the check.
	aliased := rep
	aliased.ModelID = jev.ModelID
	if aliasOutsideRequestLine(aliased.Render()) == "" {
		t.Error("positive control failed: the check passed a model line that names the alias")
	}

	// The baseline must be rendered beside the accuracy. Raw accuracy without
	// its base rate is unreadable, and has already misled once.
	for _, want := range []string{"baseline", "constant"} {
		if !strings.Contains(strings.ToLower(out), want) {
			t.Errorf("the rendered report does not mention %q:\n%s", want, out)
		}
	}
}

// hasLine reports whether want is one whole line of out.
func hasLine(out, want string) bool {
	for _, line := range strings.Split(out, "\n") {
		if line == want {
			return true
		}
	}
	return false
}

// aliasOutsideRequestLine returns the first line that names the request alias
// anywhere other than the request-model line, or "" when none does.
func aliasOutsideRequestLine(out string) string {
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "request model:") {
			continue
		}
		if strings.Contains(line, jev.ModelID) {
			return line
		}
	}
	return ""
}

// TestRun_RejectsAMissingArm: a set whose samples lack one arm's text cannot
// produce a two-arm measurement, and must fail rather than silently measure
// one arm and report a delta of zero.
func TestRun_RejectsAMissingArm(t *testing.T) {
	set := trialSet()
	delete(set.Samples[3].Text, ArmEnglishTranslat)
	if _, err := Run(context.Background(), stubAnswerer{answers: map[string]bool{}, probs: map[string]float64{}}, set, trialBuild); err == nil {
		t.Error("Run produced a two-arm report from a set missing an arm")
	}
}
