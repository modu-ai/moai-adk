package adapt

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor"
)

func readFixture(t *testing.T, prList string) monitor.PRInfo {
	t.Helper()
	g := &fakeGH{out: map[string]string{"pr list": prList, "issue view": `{"state":"OPEN"}`}}
	info, found, err := (GHReader{PRListLimit: 100, Run: g.run}).ReadPR(t.Context(), 7)
	if err != nil || !found {
		t.Fatalf("found=%v err=%v", found, err)
	}
	return info
}

// A commit status (StatusContext) has context/state and no status/conclusion. It must read by its
// state, not as an unfinished check run.
func TestReadPRStatusContextReadsByState(t *testing.T) {
	cases := []struct {
		state string
		want  string
		phase monitor.PRPhase
	}{
		{"SUCCESS", "success", monitor.PhaseReviewPending},
		{"FAILURE", "failure", monitor.PhaseChecksFailed},
		{"ERROR", "failure", monitor.PhaseChecksFailed},
		{"PENDING", "pending", monitor.PhaseChecksPending},
	}
	for _, tc := range cases {
		t.Run(tc.state, func(t *testing.T) {
			info := readFixture(t, `[{"number":55,"state":"OPEN","headRefOid":"`+head+`","statusCheckRollup":[{"__typename":"StatusContext","context":"ci/legacy","state":"`+tc.state+`"}],"closingIssuesReferences":[{"number":7}]}]`)
			if len(info.Checks) != 1 {
				t.Fatalf("checks = %+v", info.Checks)
			}
			c := info.Checks[0]
			if c.Name != "ci/legacy" || c.Conclusion != tc.want {
				t.Errorf("check = %+v, want name ci/legacy conclusion %s", c, tc.want)
			}
			if got := monitor.Phase(info); got != tc.phase {
				t.Errorf("phase = %s, want %s", got, tc.phase)
			}
		})
	}
}

// A running check run reads as checks-pending, end to end through the reader and Phase.
func TestReadPRRunningCheckRunPhaseIsPending(t *testing.T) {
	info := readFixture(t, `[{"number":55,"state":"OPEN","headRefOid":"`+head+`","statusCheckRollup":[{"name":"build","conclusion":"","status":"IN_PROGRESS"}],"closingIssuesReferences":[{"number":7}]}]`)
	if got := monitor.Phase(info); got != monitor.PhaseChecksPending {
		t.Errorf("phase = %s, want %s", got, monitor.PhaseChecksPending)
	}
}

// One Issue with an open PR and a merged PR: the open one wins whatever the numbers are.
func TestReadPROpenPRBeatsHigherNumberedMergedPR(t *testing.T) {
	info := readFixture(t, `[
	 {"number":5,"state":"OPEN","headRefOid":"`+head+`","statusCheckRollup":[],"closingIssuesReferences":[{"number":7}]},
	 {"number":9,"state":"MERGED","headRefOid":"`+head+`","statusCheckRollup":[],"closingIssuesReferences":[{"number":7}]}
	]`)
	if info.Number != 5 || !info.Open || info.Merged {
		t.Errorf("info = %+v, want open PR 5", info)
	}
}

// With only merged PRs, the newest merged one is read.
func TestReadPRNewestMergedWhenNoneOpen(t *testing.T) {
	info := readFixture(t, `[
	 {"number":5,"state":"MERGED","headRefOid":"`+head+`","statusCheckRollup":[],"closingIssuesReferences":[{"number":7}]},
	 {"number":9,"state":"MERGED","headRefOid":"`+head+`","statusCheckRollup":[],"closingIssuesReferences":[{"number":7}]}
	]`)
	if info.Number != 9 || !info.Merged {
		t.Errorf("info = %+v, want merged PR 9", info)
	}
}
