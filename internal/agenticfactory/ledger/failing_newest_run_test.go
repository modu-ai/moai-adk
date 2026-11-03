package ledger

import "testing"

// issue-1948, N54: FailingNewestRun is the read-only form of the rule Enqueue and finishing a card apply. It names the
// first kind whose newest run about the commit failed (green, gate, review), and "" when none did; a kind never
// recorded is not a failure, a pass after a failure repairs, a failure after a pass is a failure, and a run about
// another commit says nothing about this one.
func TestFailingNewestRun_NamesTheKindWhoseNewestRunFailed(t *testing.T) {
	cases := []struct {
		name string
		do   func(t *testing.T, s *Store, e int64)
		want string
	}{
		{"only the passing runs of the fixture", func(*testing.T, *Store, int64) {}, ""},
		{"green and gate pass", func(t *testing.T, s *Store, e int64) { evidenceOn(t, s, e, "c1") }, ""},
		{"a green failed", func(t *testing.T, s *Store, e int64) { greenOn(t, s, e, "c1", ResultFail) }, KindGreen},
		{"a gate failed", func(t *testing.T, s *Store, e int64) { gateOn(t, s, e, "c1", ResultFail) }, KindGate},
		{"a review failed", func(t *testing.T, s *Store, e int64) { reviewOn(t, s, e, "c1", ResultFail) }, KindReview},
		{"the gate and the review failed: the gate is named first", func(t *testing.T, s *Store, e int64) {
			reviewOn(t, s, e, "c1", ResultFail)
			gateOn(t, s, e, "c1", ResultFail)
		}, KindGate},
		{"a failure repaired by a pass", func(t *testing.T, s *Store, e int64) {
			gateOn(t, s, e, "c1", ResultFail)
			gateOn(t, s, e, "c1", ResultPass)
		}, ""},
		{"a pass followed by a failure", func(t *testing.T, s *Store, e int64) {
			gateOn(t, s, e, "c1", ResultPass)
			gateOn(t, s, e, "c1", ResultFail)
		}, KindGate},
		{"a failure about another commit", func(t *testing.T, s *Store, e int64) { gateOn(t, s, e, "c2", ResultFail) }, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := openTemp(t)
			c := gated(t, s, "a", "c1") // c1 has a passing green and gate already
			tc.do(t, s, c.Epoch)
			got, err := s.FailingNewestRun(bg, "a", "c1")
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("FailingNewestRun = %q, want %q", got, tc.want)
			}
		})
	}
}
