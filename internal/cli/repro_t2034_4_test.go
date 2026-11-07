//go:build repro_t2034

package cli

// repro_t2034_4_test.go — card t2034 claim 4 (measure only, never default-run).
//
// Claim (lane-6 card-review table, hook_stop_goal_wait.go:79, P2): "operator wait has no explicit close path".
// Interpretation under test: stopGoalValidWait treats an open wait whose reason names a `next=` event as a
// valid reason to let a turn end; nothing gives an operator wait an explicit CLOSE, so it would stay open and
// keep ending turns. The measurement seeds a card worktree and real waits (through the real `factory wait`
// verb, so the `next=` spelling is the writer's own), calls stopGoalValidWait, then performs each close path
// found in the code and observes the answer flip to false. Every fixture lives under t.TempDir().

import (
	"context"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func reproValidWait(t *testing.T, e stopWaitEnv) bool {
	t.Helper()
	ok, diag := stopGoalValidWait(e.cwd, e.root)
	t.Logf("stopGoalValidWait -> %t (diag %q)", ok, diag)
	return ok
}

func TestRepro4OperatorWaitClosePath(t *testing.T) {
	t.Run("control: no wait", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		e := newStopWaitEnv(t)
		renewMoveAssertIsolated(t, factoryRootOf(e.root))
		e.seedCard(t)
		if reproValidWait(t, e) {
			t.Fatal("a card with no wait must not be a valid wait")
		}
	})

	t.Run("operator wait: open -> valid; a landing does not close it; --resolve closes it", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		e := newStopWaitEnv(t)
		renewMoveAssertIsolated(t, factoryRootOf(e.root))
		e.seedCard(t)
		recordOperatorWait(t, e.root) // writer spelling: "next=operator ruling arrives needs the ruling"
		if !reproValidWait(t, e) {
			t.Fatal("an open operator wait with a next= event must read as valid")
		}
		// close path candidate 1: a landing (factoryPrintWakes -> SettleLandedWaits) of the card itself
		db := openWaitRoot(t, e.root)
		factoryPrintWakes(context.Background(), discardWriter{}, db, homestate.Card{CardID: "t2020", State: homestate.CardDone})
		if !reproValidWait(t, e) {
			t.Fatal("MEASURED: a landing closed an operator wait (docs say it never does)")
		}
		// close path candidate 2: the explicit verb
		out, err := runFactoryWaitCmd(t, e.root, "t2020", "--resolve")
		t.Logf("factory wait t2020 --resolve -> err=%v out=%q", err, out)
		if err != nil {
			t.Fatalf("--resolve failed: %v", err)
		}
		if reproValidWait(t, e) {
			t.Fatal("REPRODUCED: --resolve did not close the operator wait")
		}
	})

	t.Run("predecessor-landed wait with next=: open -> valid; landing of the predecessor closes it", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		e := newStopWaitEnv(t)
		renewMoveAssertIsolated(t, factoryRootOf(e.root))
		e.seedCard(t)
		out, err := runFactoryWaitCmd(t, e.root, "t2020", "--on", "t2010", "--waker", "lane-3",
			"--reason", "needs t2010", "--next", "t2010 lands", "--actor", "lane-3")
		t.Logf("factory wait t2020 --on t2010 -> err=%v out=%q", err, out)
		if err != nil {
			t.Fatal(err)
		}
		if !reproValidWait(t, e) {
			t.Fatal("an open predecessor wait with a next= event must read as valid")
		}
		db := openWaitRoot(t, e.root)
		factoryPrintWakes(context.Background(), discardWriter{}, db, homestate.Card{CardID: "t2010", State: homestate.CardMergedLocal})
		if reproValidWait(t, e) {
			t.Fatal("the predecessor landing did not close the wait")
		}
	})

	t.Run("wait without a next= event is not a valid wait (open but ignored)", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		e := newStopWaitEnv(t)
		renewMoveAssertIsolated(t, factoryRootOf(e.root))
		e.seedCard(t, operatorWait("needs the ruling"))
		if reproValidWait(t, e) {
			t.Fatal("an open wait with no next= event must not be valid")
		}
	})
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }
