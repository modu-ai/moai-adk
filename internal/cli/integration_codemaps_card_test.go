package cli

// integration_codemaps_card_test.go — guards for the codemaps-debt standing
// source (card t1018).
//
// Two of these tests exist because of a measurement rather than a hunch. On
// 2026-09-20, against an isolated queue at /tmp/t1018-dupq/repo, two adds of
// an identical text were refused (`todo add: t1 already holds this card`)
// while two texts differing only in the measured value were both admitted
// (t1 and t2). The design the card inherited put the value INSIDE the text,
// so the suppression it relied on would never have fired. The stability guard
// below is that measurement turned into a test; the discrimination guard next
// to it is the positive control, without which a constant-string
// implementation would pass the stability guard.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/graph"
)

// codemapsLayer builds a codemaps layer report with the fields the trigger
// reads. Other layers are supplied separately by the tests that need them.
func codemapsLayer(value, threshold int, anchor string) graph.LayerReport {
	return graph.LayerReport{
		Layer:         graph.LayerCodemaps,
		Metric:        graph.MetricDescribedSourceDiff,
		Value:         value,
		Threshold:     threshold,
		Verdict:       graph.VerdictFresh,
		ContentAnchor: anchor,
	}
}

func TestCodemapsDebtCardIssuesAtThreshold(t *testing.T) {
	t.Parallel()
	res := graph.CheckResult{Layers: []graph.LayerReport{codemapsLayer(40, 40, "1593d26233255c3623d24727fe747ae95e4447da")}}

	text, issue, reason := codemapsDebtCardText(res)
	if !issue {
		t.Fatalf("value == threshold must issue; declined with %q", reason)
	}
	want := "[GRAPH] codemaps described-source-diff >= 40 (anchor 1593d26233255c3623d24727fe747ae95e4447da)"
	if text != want {
		t.Fatalf("card text\n got: %q\nwant: %q", text, want)
	}
}

func TestCodemapsDebtCardSilentBelowThreshold(t *testing.T) {
	t.Parallel()
	res := graph.CheckResult{Layers: []graph.LayerReport{codemapsLayer(39, 40, "1593d262")}}

	text, issue, reason := codemapsDebtCardText(res)
	if issue {
		t.Fatalf("value < threshold must not issue; got %q", text)
	}
	if !strings.Contains(reason, "below threshold") {
		t.Fatalf("reason must name the threshold comparison; got %q", reason)
	}
}

// TestCodemapsDebtCardTextIsStableAcrossValues is the stability guard: the
// queue suppresses an exact text repeat and nothing weaker, so two landings
// in the same debt period MUST produce byte-identical text.
func TestCodemapsDebtCardTextIsStableAcrossValues(t *testing.T) {
	t.Parallel()
	anchor := "1593d26233255c3623d24727fe747ae95e4447da"
	first := graph.CheckResult{Layers: []graph.LayerReport{codemapsLayer(45, 40, anchor)}}
	second := graph.CheckResult{Layers: []graph.LayerReport{codemapsLayer(46, 40, anchor)}}

	a, issueA, _ := codemapsDebtCardText(first)
	b, issueB, _ := codemapsDebtCardText(second)
	if !issueA || !issueB {
		t.Fatalf("both must issue; got %v and %v", issueA, issueB)
	}
	if a != b {
		t.Fatalf("same anchor must yield identical text, or the queue admits one card per landing\n first: %q\nsecond: %q", a, b)
	}
}

// TestCodemapsDebtCardTextDiscriminatesAnchors is the positive control for the
// guard above: an implementation returning a constant would satisfy stability
// and never let a genuinely new debt period reach the queue.
func TestCodemapsDebtCardTextDiscriminatesAnchors(t *testing.T) {
	t.Parallel()
	first := graph.CheckResult{Layers: []graph.LayerReport{codemapsLayer(45, 40, "1593d26233255c3623d24727fe747ae95e4447da")}}
	second := graph.CheckResult{Layers: []graph.LayerReport{codemapsLayer(45, 40, "8213ca48c0e1f2a3b4c5d6e7f8091a2b3c4d5e6f")}}

	a, _, _ := codemapsDebtCardText(first)
	b, _, _ := codemapsDebtCardText(second)
	if a == b {
		t.Fatalf("different anchors must yield different text; both were %q", a)
	}
}

func TestCodemapsDebtCardDeclinesWithoutAnchor(t *testing.T) {
	t.Parallel()
	res := graph.CheckResult{Layers: []graph.LayerReport{codemapsLayer(45, 40, "")}}

	_, issue, reason := codemapsDebtCardText(res)
	if issue {
		t.Fatal("no content anchor means no stable key; the trigger must decline")
	}
	if !strings.Contains(reason, "anchor") {
		t.Fatalf("reason must name the missing anchor; got %q", reason)
	}
}

// TestCodemapsDebtCardIgnoresOtherLayers is the exit-code hazard as a unit
// test. `moai graph check` exits non-zero when any layer is not fresh, and a
// fresh worktree reports mx-index and edges absent while codemaps is fresh —
// so a trigger reading the process exit code would issue a codemaps card from
// every new worktree. Reading the codemaps layer alone is what prevents it.
func TestCodemapsDebtCardIgnoresOtherLayers(t *testing.T) {
	t.Parallel()
	res := graph.CheckResult{Layers: []graph.LayerReport{
		codemapsLayer(9, 40, "1593d26233255c3623d24727fe747ae95e4447da"),
		{Layer: graph.LayerMXIndex, Verdict: graph.VerdictAbsent, Value: 0, Threshold: 1},
		{Layer: graph.LayerEdges, Verdict: graph.VerdictAbsent, Value: 0, Threshold: 0},
	}}

	if _, issue, _ := codemapsDebtCardText(res); issue {
		t.Fatal("absent sibling layers must not issue a codemaps-debt card")
	}
}

// TestCodemapsDebtCardTreatsAbsentAsUnjudgeable guards the direction that
// fails quietly. An absent layer reports Value 0, and 0 is below every
// threshold — so a trigger comparing value against threshold alone reads an
// unmeasurable tree as a debt-free one, forever, without a word.
func TestCodemapsDebtCardTreatsAbsentAsUnjudgeable(t *testing.T) {
	t.Parallel()
	layer := codemapsLayer(0, 40, "")
	layer.Verdict = graph.VerdictAbsent
	layer.Reason = "codemaps directory missing"
	res := graph.CheckResult{Layers: []graph.LayerReport{layer}}

	_, issue, reason := codemapsDebtCardText(res)
	if issue {
		t.Fatal("an unjudgeable layer must not issue a card")
	}
	if strings.Contains(reason, codemapsBelowThresholdReason) {
		t.Fatalf("absent must not be reported as below-threshold — that is the silent read: %q", reason)
	}
	if !strings.Contains(reason, "unjudgeable") {
		t.Fatalf("reason must name the unjudgeability; got %q", reason)
	}
}

func TestCodemapsDebtCardDeclinesWithoutCodemapsLayer(t *testing.T) {
	t.Parallel()
	res := graph.CheckResult{Layers: []graph.LayerReport{
		{Layer: graph.LayerEdges, Verdict: graph.VerdictAbsent},
	}}

	_, issue, reason := codemapsDebtCardText(res)
	if issue {
		t.Fatal("a report without a codemaps layer must not issue")
	}
	if !strings.Contains(reason, "codemaps layer") {
		t.Fatalf("reason must name the missing layer; got %q", reason)
	}
}
