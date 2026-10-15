// todo_auto_mode_split_test.go — card t1814: the `--auto` two-mode split.
//
// `moai codex -l` boots its lane with a todo cycle. The unified `--auto` lane
// path prints a dispatch directive and then polls up to `--auto-wait` for the
// worker evidence, but a Codex session acts only after a command exits — the
// lane and the command wait on each other until the deadline. `--auto-lane`
// prints the directive and exits 0 without waiting (completion stays with the
// existing factory stage/complete path); `--auto-leader` is the leader's
// serial cycle, refused in a lane session. `--auto` is unchanged.
//
// The Jev display line of the leader cycle is built from the ranking stage's
// own Jev result: no local script is executed, and the line reads the same in a
// root with or without a scripts/ directory.
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/jev"
)

// autoModeRun runs the todo command with a hard time bound: a cycle that waits
// for worker evidence would block for the production --auto-wait deadline, so
// returning inside the bound is the observable "no wait".
func autoModeRun(t *testing.T, bound time.Duration, args ...string) (string, string, error) {
	t.Helper()
	type result struct {
		out, errOut string
		err         error
	}
	done := make(chan result, 1)
	go func() {
		out, errOut, err := runTodo(t, args...)
		done <- result{out, errOut, err}
	}()
	select {
	case r := <-done:
		return r.out, r.errOut, r.err
	case <-time.After(bound):
		t.Fatalf("moai todo %s did not return within %s — the command waited for worker evidence", strings.Join(args, " "), bound)
		return "", "", nil
	}
}

// TestAutoLaneDispatchesAndExitsWithoutWait — (a) a lane session's
// --auto-lane leases the ranked card, prints the dispatch directive, and
// returns at once with exit 0: no evidence exists, and the command does not
// wait for it. It leases exactly one card and records no done.
func TestAutoLaneDispatchesAndExitsWithoutWait(t *testing.T) {
	root, store := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued)
	nmLaneEnv(t, "lane-1", "")
	nmIsolatedWorktrees(t, "t1", "t2")
	laneAutoSeams(t)

	out, errOut, err := autoModeRun(t, 30*time.Second, "--auto-lane")
	if err != nil {
		t.Fatalf("--auto-lane: %v (stderr %q)", err, errOut)
	}
	// The queue root resolves through symlinks (/var → /private/var on
	// darwin), so the evidence line is matched on its root-relative tail.
	for _, want := range []string{"dispatch (", "card: t1", filepath.Join(".moai", "reports", "t1", "evidence.md") + "\n", "handoff t1"} {
		if !strings.Contains(out, want) {
			t.Errorf("--auto-lane output lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "card: t2") {
		t.Errorf("--auto-lane leased a second card; it must exit after the first dispatch:\n%s", out)
	}
	if strings.Contains(out, "non-finding: t1 worker evidence absent at deadline") {
		t.Errorf("--auto-lane waited for the evidence deadline:\n%s", out)
	}
	if s := nmQueueState(t, store, "t1"); s != factory.BacklogStatePicked {
		t.Errorf("t1 queue state = %s, want picked (leased; closed later by the existing completion path)", s)
	}
	if s := nmQueueState(t, store, "t2"); s != factory.BacklogStateQueued {
		t.Errorf("t2 queue state = %s, want queued (untouched)", s)
	}
	nmAssertLeased(t, root, "t1", "lane-1")
}

// TestAutoLaneRefusedOutsideLane — (a) --auto-lane is a lane-only mode.
func TestAutoLaneRefusedOutsideLane(t *testing.T) {
	_, store := todoFixture(t)
	sdClearLaneEnv(t)
	if _, _, err := store.Add("a card"); err != nil {
		t.Fatal(err)
	}
	_, _, err := autoModeRun(t, 30*time.Second, "--auto-lane")
	if err == nil {
		t.Fatal("--auto-lane outside a lane session must be refused")
	}
	if !strings.Contains(err.Error(), "--auto-lane") || !strings.Contains(err.Error(), "not a lane session") {
		t.Errorf("refusal must name the flag and the reason, got %q", err.Error())
	}
}

// TestAutoLeaderRefusedInLane — (a) --auto-leader is the leader's cycle; a lane
// session is refused before any queue read.
func TestAutoLeaderRefusedInLane(t *testing.T) {
	_, store := nmBase(t, factory.BacklogStateQueued)
	nmLaneEnv(t, "lane-1", "")
	laneAutoSeams(t)
	before := sdQueueBytes(t, store)
	_, _, err := autoModeRun(t, 30*time.Second, "--auto-leader")
	if err == nil {
		t.Fatal("--auto-leader in a lane session must be refused")
	}
	if !strings.Contains(err.Error(), "--auto-leader") || !strings.Contains(err.Error(), "--auto-lane") {
		t.Errorf("refusal must name the flag and the lane alternative, got %q", err.Error())
	}
	if after := sdQueueBytes(t, store); after != before {
		t.Errorf("a refused --auto-leader changed the queue")
	}
}

// TestAutoModeFlagsMutuallyExclusive — the three cycle flags select one mode.
func TestAutoModeFlagsMutuallyExclusive(t *testing.T) {
	todoFixture(t)
	for _, pair := range [][]string{
		{"--auto", "--auto-lane"},
		{"--auto", "--auto-leader"},
		{"--auto-lane", "--auto-leader"},
	} {
		if _, _, err := runTodo(t, pair...); err == nil {
			t.Errorf("%v must be refused as mutually exclusive", pair)
		}
	}
}

// jevDisplayRun runs the leader cycle with the production Jev display (opts.jev
// unset) over one queued card, the ranker answering disabled, and returns the
// output lines.
func jevDisplayRun(t *testing.T, withScript bool) (lines []string, marker string) {
	t.Helper()
	root, store := autoFixture(t, map[string]string{"one": "the only card"})
	marker = filepath.Join(root, "route-sh-executed")
	if withScript {
		script := filepath.Join(root, "scripts", "jev", "route.sh")
		if err := os.MkdirAll(filepath.Dir(script), 0o700); err != nil {
			t.Fatal(err)
		}
		body := "#!/bin/sh\ntouch '" + marker + "'\necho LOCAL-SIGNAL\n"
		if err := os.WriteFile(script, []byte(body), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	clock := time.Unix(0, 0)
	opts := autoOptions{
		wait:      time.Minute,
		liveness:  autoTestLiveness(root, "t1", true, true, nil),
		sessionID: "s",
		jevRank:   func(jev.Request) jev.Result { return jev.Result{Availability: jev.Disabled} },
		now:       func() time.Time { return clock },
		sleep:     func(time.Duration) { clock = clock.Add(time.Minute) },
	}
	var out bytes.Buffer
	if err := runAutoCycle(&out, store, root, opts); err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.TrimRight(out.String(), "\n"), "\n"), marker
}

func jevPrefixed(lines []string) []string {
	var got []string
	for _, l := range lines {
		if strings.HasPrefix(l, "jev") {
			got = append(got, l)
		}
	}
	return got
}

// TestTodoAutoJevDisplayNoExec — (c) the display line never executes a local
// script, and it reads identically in a root with and without scripts/.
func TestTodoAutoJevDisplayNoExec(t *testing.T) {
	withLines, marker := jevDisplayRun(t, true)
	if _, err := os.Stat(marker); err == nil {
		t.Errorf("scripts/jev/route.sh was executed; the display line must come from the in-process Jev result")
	}
	withoutLines, _ := jevDisplayRun(t, false)
	a, b := jevPrefixed(withLines), jevPrefixed(withoutLines)
	if strings.Join(a, "\n") != strings.Join(b, "\n") {
		t.Errorf("jev display differs by scripts/ presence:\nwith:    %q\nwithout: %q", a, b)
	}
}

// TestTodoAutoJevDisabledOneLine — (c) with the gate off (the ranker answers
// disabled) the cycle prints exactly one jev display line, and it says
// disabled.
func TestTodoAutoJevDisabledOneLine(t *testing.T) {
	lines, _ := jevDisplayRun(t, false)
	got := jevPrefixed(lines)
	if len(got) != 1 {
		t.Fatalf("want exactly one jev display line, got %d: %q", len(got), got)
	}
	if !strings.Contains(got[0], "disabled") {
		t.Errorf("jev display line = %q, want it to say disabled", got[0])
	}
}
