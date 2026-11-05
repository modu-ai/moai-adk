package cli

// factory_wait_resolve_t2013_test.go — card t2013: an operator-ruling wait
// recorded with `moai factory wait --operator` is closed by the same verb with
// `--resolve`. Until closed, an open wait that names a next event keeps the
// stop-goal hook from blocking; closing it restores the block. The stop-goal
// half is driven through the real hook function against a real factory record
// (the fixture of hook_stop_goal_wait_t2013_test.go).

import (
	"context"
	"strings"
	"testing"
)

// recordOperatorWait records the card's operator wait through the verb itself.
func recordOperatorWait(t *testing.T, root string) {
	t.Helper()
	out, err := runFactoryWaitCmd(t, root, "t2020", "--operator", "--waker", "operator",
		"--reason", "needs the ruling", "--next", "operator ruling arrives", "--actor", "lane-3")
	if err != nil {
		t.Fatalf("record operator wait: %v\n%s", err, out)
	}
}

// (1) The existing behaviour: an open operator wait with a next event does not
// block the stop-goal hook.
func TestFactoryWaitResolveOpenOperatorWaitDoesNotBlockStopGoal(t *testing.T) {
	e := newStopWaitEnv(t)
	e.seedCard(t)
	recordOperatorWait(t, e.root)
	e.armModelGoal(t)

	stdout, stderr := driveStopGoalStreams(t, e.hookInput(true))
	if strings.Contains(stdout, "block") {
		t.Fatalf("an open operator wait must not block, stdout=%q stderr=%q", stdout, stderr)
	}
}

// (2) Closing the wait restores the block.
func TestFactoryWaitResolveClosesOperatorWaitAndStopGoalBlocksAgain(t *testing.T) {
	e := newStopWaitEnv(t)
	e.seedCard(t)
	recordOperatorWait(t, e.root)
	e.armModelGoal(t)

	out, err := runFactoryWaitCmd(t, e.root, "t2020", "--resolve")
	if err != nil || !strings.Contains(out, "resolved") {
		t.Fatalf("--resolve = %v / %q, want a resolved line", err, out)
	}
	waits, err := openWaitRoot(t, e.root).LoadCardWaits(context.Background(), "t2020")
	if err != nil || len(waits) != 1 || waits[0].WokenAt == "" {
		t.Fatalf("waits after --resolve = %+v (%v), want one wait with woken_at set", waits, err)
	}
	stdout, _ := driveStopGoalStreams(t, e.hookInput(true))
	assertBlocked(t, stdout)
}

// (3) Closing twice is a no-op: the second call succeeds, says so, and leaves
// the stamp of the first untouched.
func TestFactoryWaitResolveTwiceIsNoOp(t *testing.T) {
	e := newStopWaitEnv(t)
	e.seedCard(t)
	recordOperatorWait(t, e.root)
	if out, err := runFactoryWaitCmd(t, e.root, "t2020", "--resolve"); err != nil {
		t.Fatalf("first --resolve: %v\n%s", err, out)
	}
	first, err := openWaitRoot(t, e.root).LoadCardWaits(context.Background(), "t2020")
	if err != nil || len(first) != 1 {
		t.Fatalf("waits = %+v (%v)", first, err)
	}

	out, err := runFactoryWaitCmd(t, e.root, "t2020", "--resolve")
	if err != nil || !strings.Contains(out, "already resolved") {
		t.Fatalf("second --resolve = %v / %q, want success saying already resolved", err, out)
	}
	second, err := openWaitRoot(t, e.root).LoadCardWaits(context.Background(), "t2020")
	if err != nil || len(second) != 1 || second[0].WokenAt != first[0].WokenAt {
		t.Fatalf("woken_at moved: first=%+v second=%+v (%v)", first, second, err)
	}
}

// (4) With no operator wait to close, --resolve is refused and writes nothing:
// neither a card with no wait at all nor one holding only a predecessor wait.
func TestFactoryWaitResolveRefusesWhenNoOperatorWait(t *testing.T) {
	t.Run("no wait", func(t *testing.T) {
		e := newStopWaitEnv(t)
		e.seedCard(t)
		out, err := runFactoryWaitCmd(t, e.root, "t2020", "--resolve")
		if err == nil || !strings.Contains(err.Error()+out, "no operator wait") {
			t.Fatalf("--resolve with no wait = %v / %q, want a refusal naming no operator wait", err, out)
		}
	})
	t.Run("predecessor wait only", func(t *testing.T) {
		root := t.TempDir()
		if out, err := runFactoryWaitCmd(t, root, "t2020", "--on", "t2010", "--waker", "lane-3", "--reason", "needs t2010"); err != nil {
			t.Fatalf("record wait: %v\n%s", err, out)
		}
		out, err := runFactoryWaitCmd(t, root, "t2020", "--resolve")
		if err == nil || !strings.Contains(err.Error()+out, "no operator wait") {
			t.Fatalf("--resolve with a predecessor wait only = %v / %q, want a refusal", err, out)
		}
		waits, _ := openWaitRoot(t, root).LoadCardWaits(context.Background(), "t2020")
		if len(waits) != 1 || waits[0].WokenAt != "" {
			t.Fatalf("predecessor wait changed: %+v", waits)
		}
	})
}
