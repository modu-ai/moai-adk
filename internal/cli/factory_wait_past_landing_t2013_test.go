package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// A wait recorded on a predecessor that is already past the landing (pushed,
// ci-green, done) is settled at record time, exactly like merged-local and
// merged-pr. A predecessor that never landed (still running, blocked, or
// ended without landing: abandoned, failed) leaves the wait open, because no
// landing will ever release it on its own.
func TestFactoryWaitPredecessorPastLandingSettlesAtRecord(t *testing.T) {
	cases := []struct {
		state   string
		settled bool
	}{
		{homestate.CardMergedLocal, true},
		{homestate.CardMergedPR, true},
		{homestate.CardPushed, true},
		{homestate.CardCIGreen, true},
		{homestate.CardDone, true},
		{homestate.CardRun, false},
		{homestate.CardMergeReady, false},
		{homestate.CardBlocked, false},
		{homestate.CardAbandoned, false},
		{homestate.CardFailed, false},
	}
	for _, tc := range cases {
		t.Run(tc.state, func(t *testing.T) {
			root := t.TempDir()
			seedWaitCard(t, root, "run-1", "t2010", tc.state)
			out, err := runFactoryWaitCmd(t, root, "t2020", "--on", "t2010", "--waker", "lane-3", "--reason", "needs t2010", "--next", "t2010 lands")
			if err != nil {
				t.Fatalf("wait: %v\n%s", err, out)
			}
			lines := wakeLines(out)
			w := onlyWait(t, root, "t2020")
			if tc.settled {
				if len(lines) != 1 || w.WokenAt == "" {
					t.Errorf("state %s: want one wake line and woken_at set, got lines=%q wait=%+v", tc.state, lines, w)
				}
				return
			}
			if len(lines) != 0 || w.WokenAt != "" {
				t.Errorf("state %s: want the wait left open, got lines=%q wait=%+v", tc.state, lines, w)
			}
		})
	}
}
