package web

import (
	"os"
	"testing"
	"time"
)

// TestSessionStateJoinsLivenessSignalsWithOr is the regression guard for
// GH #1711 defect 3: sessionState joined the PID probe and the heartbeat floor
// with AND, so a live session whose heartbeat had frozen rendered as stale.
// session.LiveAnchoredSessions (anchor.go) already joined the same two signals
// with OR; this pins the two judgements to the same shape.
//
// The mutant this catches: restoring the AND turns the first row Live→stale.
func TestSessionStateJoinsLivenessSignalsWithOr(t *testing.T) {
	t.Parallel()

	now := time.Now()
	fresh := now.Add(-time.Minute)
	frozen := now.Add(-24 * time.Hour)
	livePID := os.Getpid() // this test process is alive by construction

	cases := []struct {
		name          string
		pid           int
		lastHeartbeat time.Time
		want          string
	}{
		{
			// The defect. A live PID is direct evidence the session is running;
			// a frozen heartbeat must not overrule it.
			name: "live pid with frozen heartbeat is live",
			pid:  livePID, lastHeartbeat: frozen, want: StateLive,
		},
		{
			// The conservative fallback for platforms where the PID probe
			// cannot prove death.
			name: "fresh heartbeat without a live pid is live",
			pid:  0, lastHeartbeat: fresh, want: StateLive,
		},
		{
			// Negative control: the OR must not make every row live. Without
			// this row a resolver that always returned StateLive would pass.
			name: "no live pid and frozen heartbeat is stale",
			pid:  0, lastHeartbeat: frozen, want: StateStale,
		},
		{
			name: "live pid with fresh heartbeat is live",
			pid:  livePID, lastHeartbeat: fresh, want: StateLive,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := sessionState(tc.lastHeartbeat, tc.pid, now); got != tc.want {
				t.Errorf("sessionState(pid=%d, heartbeat age=%s) = %q, want %q",
					tc.pid, now.Sub(tc.lastHeartbeat), got, tc.want)
			}
		})
	}
}
