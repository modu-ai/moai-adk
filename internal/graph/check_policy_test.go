package graph

import "testing"

func intp(n int) *int { return &n }

func policyResult(codemaps string, contribution *int, others ...string) CheckResult {
	layers := []LayerReport{{Layer: LayerCodemaps, Verdict: codemaps, Contribution: contribution}}
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
	cases := []struct {
		name         string
		res          CheckResult
		event        string
		policy       string
		wantBlocking bool
		wantWarnings int
	}{
		{"inherited pull_request warns", policyResult(VerdictStale, intp(1), VerdictFresh, VerdictFresh, VerdictFresh), EventPullRequest, InheritedDriftWarn, false, 1},
		{"contribution just under threshold warns", policyResult(VerdictStale, intp(39), VerdictFresh, VerdictFresh, VerdictFresh), EventPullRequest, InheritedDriftWarn, false, 1},
		{"contribution at threshold blocks", policyResult(VerdictStale, intp(40), VerdictFresh, VerdictFresh, VerdictFresh), EventPullRequest, InheritedDriftWarn, true, 0},
		{"self contribution 41 blocks", policyResult(VerdictStale, intp(41), VerdictFresh, VerdictFresh, VerdictFresh), EventPullRequest, InheritedDriftWarn, true, 0},
		{"unmeasured contribution blocks", policyResult(VerdictStale, nil, VerdictFresh, VerdictFresh, VerdictFresh), EventPullRequest, InheritedDriftWarn, true, 0},
		{"push blocks", policyResult(VerdictStale, intp(1), VerdictFresh, VerdictFresh, VerdictFresh), EventPush, InheritedDriftWarn, true, 0},
		{"unknown event blocks", policyResult(VerdictStale, intp(1), VerdictFresh, VerdictFresh, VerdictFresh), "", InheritedDriftWarn, true, 0},
		{"policy block blocks", policyResult(VerdictStale, intp(1), VerdictFresh, VerdictFresh, VerdictFresh), EventPullRequest, InheritedDriftBlock, true, 0},
		{"codemaps absent blocks", policyResult(VerdictAbsent, intp(1), VerdictFresh, VerdictFresh, VerdictFresh), EventPullRequest, InheritedDriftWarn, true, 0},
		{"edges stale alongside blocks", policyResult(VerdictStale, intp(1), VerdictFresh, VerdictStale, VerdictFresh), EventPullRequest, InheritedDriftWarn, true, 0},
		{"edges absent alongside blocks", policyResult(VerdictStale, intp(1), VerdictFresh, VerdictAbsent, VerdictFresh), EventPullRequest, InheritedDriftWarn, true, 0},
		{"citations stale alongside blocks", policyResult(VerdictStale, intp(1), VerdictFresh, VerdictFresh, VerdictStale), EventPullRequest, InheritedDriftWarn, true, 0},
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
			// The verdict is a fact and is never rewritten by classification.
			if res.Layers[0].Verdict != c.res.Layers[0].Verdict {
				t.Errorf("verdict rewritten: %q -> %q", c.res.Layers[0].Verdict, res.Layers[0].Verdict)
			}
		})
	}
}

// A caller that never classifies keeps the fail-closed behavior.
func TestBlockingWithoutClassificationFailsClosed(t *testing.T) {
	res := policyResult(VerdictStale, intp(1), VerdictFresh, VerdictFresh, VerdictFresh)
	if !res.Blocking() {
		t.Fatal("an unclassified stale layer must block")
	}
}
