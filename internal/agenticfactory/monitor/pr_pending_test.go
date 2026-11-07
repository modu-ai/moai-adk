package monitor

import "testing"

// A check that is still running (the reader reports it as "pending") is not a failure: the PR is
// waiting on checks, not blocked by them.
func TestPhaseRunningCheckIsPendingNotFailed(t *testing.T) {
	cases := []struct {
		name   string
		checks []CheckRun
		want   PRPhase
	}{
		{"one running", []CheckRun{{Name: "build", Conclusion: "pending", HeadSHA: head}}, PhaseChecksPending},
		{"success plus running", []CheckRun{
			{Name: "build", Conclusion: "success", HeadSHA: head},
			{Name: "lint", Conclusion: "pending", HeadSHA: head},
		}, PhaseChecksPending},
		{"failure beats running", []CheckRun{
			{Name: "build", Conclusion: "failure", HeadSHA: head},
			{Name: "lint", Conclusion: "pending", HeadSHA: head},
		}, PhaseChecksFailed},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Phase(PRInfo{Open: true, HeadSHA: head, Checks: tc.checks, ReviewVerdict: "PASS"})
			if got != tc.want {
				t.Errorf("Phase = %s, want %s", got, tc.want)
			}
		})
	}
}
