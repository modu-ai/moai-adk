package graph

import "testing"

func intp(n int) *int { return &n }

func boolp(b bool) *bool { return &b }

// policyResult builds a result whose codemaps stamp and body are stable against
// the first parent (the normal case); policyResultWithStability overrides it.
func policyResult(value int, codemaps string, contribution *int, others ...string) CheckResult {
	return policyResultWithStability(boolp(true), value, codemaps, contribution, others...)
}

func policyResultWithStability(stable *bool, value int, codemaps string, contribution *int, others ...string) CheckResult {
	// These hand-built rows model a non-overlapping change: base = value - contribution.
	var base *int
	if contribution != nil {
		base = intp(value - *contribution)
	}
	layers := []LayerReport{{Layer: LayerCodemaps, Verdict: codemaps, Value: value, Contribution: contribution, BaseCount: base, CodemapsInputsStable: stable}}
	names := []string{LayerMXIndex, LayerEdges, LayerCitations}
	for i, v := range others {
		layers = append(layers, LayerReport{Layer: names[i], Verdict: v})
	}
	return CheckResult{Layers: layers}
}

// The red line itself is unchanged by the policy (R208: no raised numbers).
func TestDefaultThresholdsUnchangedByPolicy(t *testing.T) {
	th := DefaultThresholds()
	if th.CodemapsChangedFiles != 40 || th.MXIndexChangedFiles != 1 {
		t.Fatalf("thresholds moved: codemaps=%d mx=%d, want 40/1", th.CodemapsChangedFiles, th.MXIndexChangedFiles)
	}
}

func TestClassifyVerdicts(t *testing.T) {
	th := DefaultThresholds()
	f := VerdictFresh
	cases := []struct {
		name         string
		res          CheckResult
		event        string
		policy       string
		wantBlocking bool
		wantWarnings int
		wantOrigin   string
	}{
		// inherited: base alone (Value - Contribution) is already at the threshold.
		{"base 40 + pr 1 is inherited", policyResult(41, VerdictStale, intp(1), f, f, f), EventPullRequest, InheritedDriftWarn, false, 1, DriftInherited},
		{"base 41 + pr 0 is inherited", policyResult(41, VerdictStale, intp(0), f, f, f), EventPullRequest, InheritedDriftWarn, false, 1, DriftInherited},
		{"base 40 + pr 0 is inherited", policyResult(40, VerdictStale, intp(0), f, f, f), EventPullRequest, InheritedDriftWarn, false, 1, DriftInherited},
		// crossed by this change: base under the line, the change took it over.
		{"base 25 + pr 16 crossed", policyResult(41, VerdictStale, intp(16), f, f, f), EventPullRequest, InheritedDriftWarn, true, 0, DriftCrossedByChange},
		{"base 39 + pr 1 crossed", policyResult(40, VerdictStale, intp(1), f, f, f), EventPullRequest, InheritedDriftWarn, true, 0, DriftCrossedByChange},
		// originated: the change alone is at or over the line.
		{"base 0 + pr 41 originated", policyResult(41, VerdictStale, intp(41), f, f, f), EventPullRequest, InheritedDriftWarn, true, 0, DriftOriginated},
		{"base 0 + pr 40 originated", policyResult(40, VerdictStale, intp(40), f, f, f), EventPullRequest, InheritedDriftWarn, true, 0, DriftOriginated},
		// unmeasured: no origin can be stated.
		{"unmeasured contribution blocks", policyResult(41, VerdictStale, nil, f, f, f), EventPullRequest, InheritedDriftWarn, true, 0, ""},
		// event and policy gates.
		{"push blocks", policyResult(41, VerdictStale, intp(1), f, f, f), EventPush, InheritedDriftWarn, true, 0, DriftInherited},
		{"unknown event blocks", policyResult(41, VerdictStale, intp(1), f, f, f), "", InheritedDriftWarn, true, 0, DriftInherited},
		{"policy block blocks", policyResult(41, VerdictStale, intp(1), f, f, f), EventPullRequest, InheritedDriftBlock, true, 0, DriftInherited},
		// other layers.
		{"codemaps absent blocks", policyResult(41, VerdictAbsent, intp(1), f, f, f), EventPullRequest, InheritedDriftWarn, true, 0, ""},
		{"edges stale alongside blocks", policyResult(41, VerdictStale, intp(1), f, VerdictStale, f), EventPullRequest, InheritedDriftWarn, true, 0, DriftInherited},
		{"edges absent alongside blocks", policyResult(41, VerdictStale, intp(1), f, VerdictAbsent, f), EventPullRequest, InheritedDriftWarn, true, 0, DriftInherited},
		{"citations stale alongside blocks", policyResult(41, VerdictStale, intp(1), f, f, VerdictStale), EventPullRequest, InheritedDriftWarn, true, 0, DriftInherited},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res := c.res
			res.ClassifyVerdicts(c.event, c.policy, th)
			if got := res.Blocking(); got != c.wantBlocking {
				t.Errorf("Blocking() = %v, want %v", got, c.wantBlocking)
			}
			if got := len(res.Warnings()); got != c.wantWarnings {
				t.Errorf("len(Warnings()) = %d, want %d", got, c.wantWarnings)
			}
			if got := res.Layers[0].DriftOrigin; got != c.wantOrigin {
				t.Errorf("DriftOrigin = %q, want %q", got, c.wantOrigin)
			}
			// The verdict is a fact and is never rewritten by classification.
			if res.Layers[0].Verdict != c.res.Layers[0].Verdict {
				t.Errorf("verdict rewritten: %q -> %q", c.res.Layers[0].Verdict, res.Layers[0].Verdict)
			}
		})
	}
}

// A caller that never classifies keeps the fail-closed behavior.
func TestBlockingWithoutClassificationFailsClosed(t *testing.T) {
	res := policyResult(41, VerdictStale, intp(1), VerdictFresh, VerdictFresh, VerdictFresh)
	if !res.Blocking() {
		t.Fatal("an unclassified stale layer must block")
	}
}

// F2: Value - Contribution is a statement about the base only while the change
// left the codemaps stamp and body alone. A moved or unmeasured window never
// reads as inherited, however large the count.
func TestClassifyVerdicts_MovedOrUnknownAnchorBlocks(t *testing.T) {
	th := DefaultThresholds()
	f := VerdictFresh
	for _, c := range []struct {
		name   string
		stable *bool
	}{
		{"stamp or body moved", boolp(false)},
		{"stability not measured", nil},
	} {
		t.Run(c.name, func(t *testing.T) {
			res := policyResultWithStability(c.stable, 41, VerdictStale, intp(0), f, f, f)
			res.ClassifyVerdicts(EventPullRequest, InheritedDriftWarn, th)
			if !res.Blocking() || len(res.Warnings()) != 0 {
				t.Fatalf("must block, got blocking=%v warnings=%d", res.Blocking(), len(res.Warnings()))
			}
			if got := res.Layers[0].DriftOrigin; got != DriftAnchorMoved {
				t.Errorf("DriftOrigin = %q, want %q", got, DriftAnchorMoved)
			}
		})
	}
}

// An unmeasured base count never reads as inherited, however large Value is.
func TestClassifyVerdicts_UnmeasuredBaseCountBlocks(t *testing.T) {
	res := policyResult(41, VerdictStale, intp(1), VerdictFresh, VerdictFresh, VerdictFresh)
	res.Layers[0].BaseCount = nil
	res.ClassifyVerdicts(EventPullRequest, InheritedDriftWarn, DefaultThresholds())
	if !res.Blocking() || res.Layers[0].DriftOrigin != DriftAnchorMoved {
		t.Fatalf("blocking=%v origin=%q, want blocking anchor-moved", res.Blocking(), res.Layers[0].DriftOrigin)
	}
}

// A change that is itself over the line stays "originated" even when the
// window also moved: the stronger fact wins.
func TestClassifyVerdicts_OriginatedBeatsMovedAnchor(t *testing.T) {
	res := policyResultWithStability(boolp(false), 41, VerdictStale, intp(41), VerdictFresh, VerdictFresh, VerdictFresh)
	res.ClassifyVerdicts(EventPullRequest, InheritedDriftWarn, DefaultThresholds())
	if got := res.Layers[0].DriftOrigin; got != DriftOriginated {
		t.Errorf("DriftOrigin = %q, want %q", got, DriftOriginated)
	}
}
