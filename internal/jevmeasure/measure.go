package jevmeasure

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/jev"
)

// Arm names a language arm. Every measurement runs both: the text in its
// Korean original, and the same text in English translation, with the delta
// recorded (REQ-JEVO-010).
//
// The arms MEASURE and decide nothing. The model card notes reduced non-English
// and CJK accuracy and this repository's cards are Korean, so the effect is
// real and worth sizing — but sizing it is not the same as acting on it, and
// nothing in this package changes a card, a schema, or an issuance path
// (REQ-JEVO-011).
type Arm string

const (
	// ArmKoreanOriginal is the text as this repository actually stores it.
	ArmKoreanOriginal Arm = "ko-original"
	// ArmEnglishTranslat is the same text in English translation.
	ArmEnglishTranslat Arm = "en-translated"
)

// Arms is the fixed arm order. Korean first, because it is the arm that
// describes the repository as it is; English second, because it is the
// counterfactual.
func Arms() []Arm { return []Arm{ArmKoreanOriginal, ArmEnglishTranslat} }

// Sample is one labelled item drawn from this repository's own data.
type Sample struct {
	ID    string
	Text  map[Arm]string
	Label string
}

// Set is a labelled answer set for one consumer.
type Set struct {
	Consumer string
	Shape    jev.AnswerKind
	Labels   []string
	Samples  []Sample
}

// Baseline is the constant-answer baseline: the accuracy of always answering
// the majority label, without measuring anything at all.
//
// This is the number to beat, and it is often high. On the rejected
// premise-death task it sat at 75% while the measured headline moved from
// 31.5% to 47.6% — a sixteen-point gain that read as success and was not.
type Baseline struct {
	Label    string
	Correct  int
	N        int
	Accuracy float64
}

// ConstantBaseline computes the majority-label baseline for a set.
//
// An empty set returns an error rather than a zero baseline. A zero would read
// as "0% to beat", which is the most dangerous possible default: every
// measurement would clear it.
func ConstantBaseline(set Set) (Baseline, error) {
	if len(set.Samples) == 0 {
		return Baseline{}, fmt.Errorf("jevmeasure: consumer %q has no labelled samples — there is no baseline to beat", set.Consumer)
	}
	counts := map[string]int{}
	for _, s := range set.Samples {
		if strings.TrimSpace(s.Label) == "" {
			return Baseline{}, fmt.Errorf("jevmeasure: sample %q carries no label", s.ID)
		}
		counts[s.Label]++
	}
	// Deterministic winner: highest count, ties broken by label order.
	labels := make([]string, 0, len(counts))
	for l := range counts {
		labels = append(labels, l)
	}
	sort.Strings(labels)
	best := labels[0]
	for _, l := range labels {
		if counts[l] > counts[best] {
			best = l
		}
	}
	n := len(set.Samples)
	return Baseline{
		Label:    best,
		Correct:  counts[best],
		N:        n,
		Accuracy: float64(counts[best]) / float64(n),
	}, nil
}

// ArmResult is one arm's measured outcome.
type ArmResult struct {
	Arm      Arm
	N        int
	Answered int
	// Unavailable counts samples whose call produced no answer. They are NEVER
	// counted correct: an absent answer is not a negative answer.
	Unavailable   int
	Correct       int
	Accuracy      float64
	Probabilities []float64
	// Conditions records the distinct unavailability conditions observed, so a
	// run that degraded can say HOW rather than only how often.
	Conditions []string
}

// Answerer is the call seam. *jev.Client satisfies it; every test supplies a
// stub. The interface exists so this package never imports a transport and no
// test can reach the vendor endpoint by accident.
type Answerer interface {
	Ask(ctx context.Context, req jev.Request) jev.Result
}

// BuildFunc constructs the request for one sample in one arm. The consumer
// owns it, because the question is the consumer's — the harness owns only the
// scoring and the gate.
type BuildFunc func(s Sample, arm Arm) (jev.Request, error)

// Source names what produced a report's numbers.
type Source string

const (
	// SourceLive — the answers came from the live client against the vendor
	// endpoint under a real credential.
	SourceLive Source = "live"
	// SourceInjected — the answers came from an injected answerer (a stub, a
	// fake, a replay). Numbers from this source are shape demonstrations, not
	// measurements, and Verdict refuses to ship on them.
	SourceInjected Source = "injected"
)

// noulTrueAbove is the probability above which a Noul answer reads as its TRUE
// label when scoring accuracy. The vendor returns a Noul as a probability, not
// a boolean; "more likely true than not" is the neutral cut, and it is only the
// scoring rule — a consumer's decision gate is a fitted Threshold, not this.
const noulTrueAbove = 0.5

// Verdict is the gate's binding decision (REQ-JEVO-009).
type Verdict string

const (
	// VerdictShip — the consumer beat its constant baseline under a live
	// measurement.
	VerdictShip Verdict = "ship"
	// VerdictWithhold — it did not, or the measurement was not live.
	VerdictWithhold Verdict = "withhold"
)

// PositiveControl demonstrates that the measuring apparatus fires on the path
// an absence was claimed about.
type PositiveControl struct {
	Description string
	Observed    string
}

// Absence is a zero-hit / no-difference / null result, together with the
// control that makes it a finding rather than a gap (REQ-JEVO-015).
type Absence struct {
	Claim   string
	Control PositiveControl
}

// Report is the measurement artifact.
type Report struct {
	Consumer  string
	ModelID   string
	Shape     jev.AnswerKind
	Source    Source
	Baseline  Baseline
	Arms      []ArmResult
	Delta     float64
	Threshold *Threshold
	Absence   *Absence
}

// BestArm returns the arm with the highest accuracy. Ties resolve to the
// Korean original — the arm that describes the repository as it actually is,
// so a tie never reads as an argument for translating anything.
func (r Report) BestArm() ArmResult {
	best := ArmResult{}
	for i, a := range r.Arms {
		if i == 0 || a.Accuracy > best.Accuracy {
			best = a
		}
	}
	return best
}

// Verdict applies the gate. It is binding, not advisory.
func (r Report) Verdict() (Verdict, string) {
	if r.Source != SourceLive {
		return VerdictWithhold, fmt.Sprintf(
			"the measurement source is %q, not %q — an injected answerer demonstrates the harness's shape and measures nothing",
			r.Source, SourceLive)
	}
	if r.Baseline.N == 0 {
		return VerdictWithhold, "no constant-answer baseline was computed"
	}
	best := r.BestArm()
	if best.Accuracy <= r.Baseline.Accuracy {
		return VerdictWithhold, fmt.Sprintf(
			"best arm %s scored %.3f against a constant-%q baseline of %.3f — the margin over the constant is what changes a decision, and there is none",
			best.Arm, best.Accuracy, r.Baseline.Label, r.Baseline.Accuracy)
	}
	return VerdictShip, fmt.Sprintf(
		"best arm %s scored %.3f against a constant-%q baseline of %.3f",
		best.Arm, best.Accuracy, r.Baseline.Label, r.Baseline.Accuracy)
}

// Validate checks the artifact's internal obligations.
func (r Report) Validate() error {
	if r.ModelID == "" {
		return fmt.Errorf("jevmeasure: report cites no model id (REQ-JEVO-014)")
	}
	if r.Absence != nil {
		if strings.TrimSpace(r.Absence.Control.Description) == "" {
			return fmt.Errorf(
				"jevmeasure: the report claims an absence (%q) with no positive control — a zero-hit and a broken apparatus are indistinguishable without one",
				r.Absence.Claim)
		}
		if strings.TrimSpace(r.Absence.Control.Observed) == "" {
			return fmt.Errorf(
				"jevmeasure: the positive control for %q records no observation — a control that does not say what it produced asserts nothing",
				r.Absence.Claim)
		}
	}
	if r.Threshold != nil {
		if err := r.Threshold.ApplicableTo(r.Consumer, r.Shape); err != nil {
			return err
		}
	}
	return nil
}

// Render produces the human-readable artifact. It always names the pinned
// model id and always prints the baseline beside the accuracy: raw accuracy
// without its base rate is unreadable, and has already misled once.
func (r Report) Render() string {
	var b strings.Builder
	fmt.Fprintf(&b, "consumer: %s\n", r.Consumer)
	fmt.Fprintf(&b, "model: %s (pinned)\n", r.ModelID)
	fmt.Fprintf(&b, "shape: %s\n", r.Shape)
	fmt.Fprintf(&b, "source: %s\n", r.Source)
	fmt.Fprintf(&b, "constant-answer baseline: always %q = %.3f (%d/%d)\n",
		r.Baseline.Label, r.Baseline.Accuracy, r.Baseline.Correct, r.Baseline.N)
	for _, a := range r.Arms {
		fmt.Fprintf(&b, "arm %s: accuracy %.3f (%d/%d answered, %d unavailable)\n",
			a.Arm, a.Accuracy, a.Correct, a.Answered, a.Unavailable)
		if len(a.Conditions) > 0 {
			fmt.Fprintf(&b, "  conditions: %s\n", strings.Join(a.Conditions, "; "))
		}
	}
	fmt.Fprintf(&b, "arm delta (english minus korean): %+.3f\n", r.Delta)
	if r.Threshold != nil {
		fmt.Fprintf(&b, "threshold: %.3f fitted on %d samples — %s\n",
			r.Threshold.Value, r.Threshold.SampleSize, r.Threshold.Provenance)
	}
	if r.Absence != nil {
		fmt.Fprintf(&b, "absence claimed: %s\n", r.Absence.Claim)
		fmt.Fprintf(&b, "  positive control: %s -> %s\n",
			r.Absence.Control.Description, r.Absence.Control.Observed)
	}
	v, reason := r.Verdict()
	fmt.Fprintf(&b, "verdict: %s — %s\n", v, reason)
	return b.String()
}

// Run measures a set against an answerer, in both arms.
//
// The caller's Source is inferred rather than declared: only a *jev.Client is
// treated as live. A declared-by-the-caller source would be the one field a
// hurried change could set to "live" to make a stub's numbers shippable.
func Run(ctx context.Context, answerer Answerer, set Set, build BuildFunc) (Report, error) {
	baseline, err := ConstantBaseline(set)
	if err != nil {
		return Report{}, err
	}
	if build == nil {
		return Report{}, fmt.Errorf("jevmeasure: no request builder")
	}

	// Every sample must carry every arm's text. A set missing an arm would
	// otherwise produce a one-arm measurement reporting a delta of zero, which
	// is indistinguishable from a measured finding of no difference.
	for _, s := range set.Samples {
		for _, arm := range Arms() {
			if strings.TrimSpace(s.Text[arm]) == "" {
				return Report{}, fmt.Errorf(
					"jevmeasure: sample %q carries no %s text — a missing arm would report a delta of zero, which is not a measured no-difference",
					s.ID, arm)
			}
		}
	}

	rep := Report{
		Consumer: set.Consumer,
		ModelID:  jev.ModelID,
		Shape:    set.Shape,
		Source:   sourceOf(answerer),
		Baseline: baseline,
	}

	for _, arm := range Arms() {
		res := ArmResult{Arm: arm, N: len(set.Samples)}
		seen := map[string]bool{}
		for _, s := range set.Samples {
			req, err := build(s, arm)
			if err != nil {
				return Report{}, fmt.Errorf("jevmeasure: build request for sample %q arm %s: %w", s.ID, arm, err)
			}
			result := answerer.Ask(ctx, req)
			if !result.OK() || len(result.Answers) == 0 {
				res.Unavailable++
				cond := result.NoticeLine()
				if cond != "" && !seen[cond] {
					seen[cond] = true
					res.Conditions = append(res.Conditions, cond)
				}
				continue
			}
			res.Answered++
			answer := result.Answers[0]
			res.Probabilities = append(res.Probabilities, answer.Probability)
			if labelOf(answer, set.Labels) == s.Label {
				res.Correct++
			}
		}
		if res.N > 0 {
			// The denominator is the FULL set, not the answered subset. An
			// unavailable answer is a failure to judge, and dividing it away
			// would let a run that mostly failed report a high accuracy.
			res.Accuracy = float64(res.Correct) / float64(res.N)
		}
		rep.Arms = append(rep.Arms, res)
	}

	rep.Delta = rep.Arms[1].Accuracy - rep.Arms[0].Accuracy
	return rep, nil
}

// sourceOf reports whether the answerer is the live client.
func sourceOf(a Answerer) Source {
	if _, ok := a.(*jev.Client); ok {
		return SourceLive
	}
	return SourceInjected
}

// labelOf maps a typed answer onto the set's label vocabulary.
//
// For a Noul, the convention is that labels[0] is the TRUE label and labels[1]
// the false one; a set declaring fewer than two labels for a Noul cannot be
// scored and yields the empty string, which matches nothing. A Noul answer is a
// probability, so the label is read at noulTrueAbove.
func labelOf(a jev.Answer, labels []string) string {
	switch a.Kind {
	case jev.KindChoice:
		return a.Choice
	case jev.KindNoul:
		if len(labels) < 2 {
			return ""
		}
		if a.Probability > noulTrueAbove {
			return labels[0]
		}
		return labels[1]
	default:
		return ""
	}
}

// Threshold is a confidence gate fitted on THIS repository's measured data.
type Threshold struct {
	Consumer   string
	Shape      jev.AnswerKind
	Value      float64
	SampleSize int
	Provenance string
}

// vendorIllustrative are the figures that appear in vendor documentation as
// examples. They are not measurements of anything in this repository, and
// copying one is an unmeasured claim wearing a decision rule's clothes.
var vendorIllustrative = []float64{0.6, 0.85}

// VendorIllustrativeThresholds returns the vendor-documentation figures that
// may not be adopted by copy.
func VendorIllustrativeThresholds() []float64 {
	out := make([]float64, len(vendorIllustrative))
	copy(out, vendorIllustrative)
	return out
}

// NewThreshold constructs a threshold with explicit provenance.
//
// A vendor-documentation figure is refused when it arrives with no provenance.
// The refusal is narrow on purpose: if a fitted threshold genuinely lands on
// 0.6, saying so in the provenance is all that is required — what is forbidden
// is adopting the number BECAUSE the documentation printed it.
func NewThreshold(consumer string, shape jev.AnswerKind, value float64, sampleSize int, provenance string) (Threshold, error) {
	if strings.TrimSpace(provenance) == "" {
		for _, v := range vendorIllustrative {
			if value == v {
				return Threshold{}, fmt.Errorf(
					"jevmeasure: %v is a vendor-documentation illustration and carries no measured provenance — fit the threshold on this repository's data (REQ-JEVO-013)", value)
			}
		}
		return Threshold{}, fmt.Errorf("jevmeasure: a threshold with no provenance is an unmeasured claim")
	}
	if sampleSize <= 0 {
		return Threshold{}, fmt.Errorf("jevmeasure: a threshold carries the sample size it was fitted on; got %d", sampleSize)
	}
	return Threshold{
		Consumer: consumer, Shape: shape, Value: value,
		SampleSize: sampleSize, Provenance: provenance,
	}, nil
}

// ApplicableTo refuses a threshold fitted for a different consumer or a
// different question shape (REQ-JEVO-013). A Noul and a Choice calibrate
// differently; carrying a number across is a transfer, not a reuse.
func (t Threshold) ApplicableTo(consumer string, shape jev.AnswerKind) error {
	if t.Consumer != consumer {
		return fmt.Errorf("jevmeasure: threshold was fitted for consumer %q, not %q", t.Consumer, consumer)
	}
	if t.Shape != shape {
		return fmt.Errorf("jevmeasure: threshold was fitted for a %s question, not a %s — shapes calibrate differently", t.Shape, shape)
	}
	return nil
}

// FitThreshold fits a gate on an arm's measured probability distribution.
//
// It is deliberately a simple statistic — the median observed probability —
// because the rejected task measured the alternative: a sweep across
// 0.30-0.80 was FLAT, and the two cells that looked better rested on 5 and 4
// samples, where a single flip moves 60% to 40%. A more elaborate fit would
// dress that noise up as precision. The sample size travels with the value so
// a reader can see how thin the ground is.
func FitThreshold(consumer string, shape jev.AnswerKind, arm ArmResult) (Threshold, error) {
	if len(arm.Probabilities) == 0 {
		return Threshold{}, fmt.Errorf("jevmeasure: arm %s produced no probabilities to fit on", arm.Arm)
	}
	ps := make([]float64, len(arm.Probabilities))
	copy(ps, arm.Probabilities)
	sort.Float64s(ps)
	median := ps[len(ps)/2]
	return Threshold{
		Consumer:   consumer,
		Shape:      shape,
		Value:      median,
		SampleSize: arm.N,
		Provenance: fmt.Sprintf(
			"median observed probability over %d samples of this repository's own data, arm %s (a cost/coverage lever, NOT a correctness filter: high confidence does not imply correctness)",
			arm.N, arm.Arm),
	}, nil
}
