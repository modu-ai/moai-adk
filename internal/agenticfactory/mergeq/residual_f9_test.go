package mergeq

// F9 of the final audit of a36ea8c368 (residual classification): the context ends while the gates run. The verifying write that
// would record the gate failure then fails with "context canceled", so the gate's real reason is lost from what the caller is told
// and the row stays verifying. A retry under a fresh context lands. The expectation: the real reason is not lost (the retry
// that lands is asserted as well, and holds today).

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestResidualF9_ALandingWhoseContextEndsDuringTheGateDoesNotLoseTheGateReason(t *testing.T) {
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	s := queuedCard(t, "t1", cand)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := realMerger(t, dir, s)
	m.Verify = func(_ context.Context, _ string) error {
		cancel()
		return errors.New("gate failed: the real reason")
	}
	res, err := m.Land(ctx, "t1", cand)
	t.Logf("land: outcome=%q reason=%q err=%v row=%s", res.Outcome, res.Reason, err, rowState(t, s, "t1"))
	res2, err2 := realMerger(t, dir, s).Land(context.Background(), "t1", cand)
	t.Logf("land retry: outcome=%q err=%v row=%s", res2.Outcome, err2, rowState(t, s, "t1"))
	if err2 != nil || res2.Outcome != OutcomeLanded {
		t.Errorf("the retry after a landing whose context ended did not land: outcome=%q err=%v", res2.Outcome, err2)
	}
	told := res.Reason
	if err != nil {
		told += " " + err.Error()
	}
	if !strings.Contains(told, "the real reason") {
		t.Errorf("the gate's failure reason is lost: outcome=%q reason=%q err=%v", res.Outcome, res.Reason, err)
	}
}
