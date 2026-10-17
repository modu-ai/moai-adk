//go:build !windows

// factory_lease_prefetch_test.go — SPEC-FACTORY-ATOMIC-LEASE-001 (card t1746) tests of
// where the `factory next` verb issues its open-PR query (AC-FAL-015 (vii)): ahead of
// the queue lock of the lease attempt that prints it, and only when that attempt can
// lease a card. They reuse the gh seam (flFirstGhCall) and the lease-wait seam of the
// timing tests. The Windows build compiles test files too, so these stay unix-tagged
// with the helpers they build on.
package cli

import (
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// TestFactoryNextIssuesNoOpenPRQueryWhenNothingIsLeased — a `next` run that leases
// nothing makes zero `gh` calls: an empty queue (once and across --wait polls), a
// permanent nominate refusal, and a quota hold in the bare and the nominated form.
// Each case ends the way it did before the query moved ahead of the lock.
func TestFactoryNextIssuesNoOpenPRQueryWhenNothingIsLeased(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T)
		args  []string
		check func(t *testing.T, out, stderr string, err error)
	}{
		{
			name: "empty queue",
			setup: func(t *testing.T) {
				nmBase(t)
				nmLaneEnv(t, "lane-1", "")
			},
			args:  []string{"--run", fcRun},
			check: func(t *testing.T, _, _ string, err error) { sdExit3(t, "empty queue", err) },
		},
		{
			name: "empty queue across --wait polls",
			setup: func(t *testing.T) {
				nmBase(t)
				nmLaneEnv(t, "lane-1", "")
				qasFakeClock(t, nil, nil)
			},
			args:  []string{"--run", fcRun, "--wait", "--wait-bound", "10s"},
			check: func(t *testing.T, _, _ string, err error) { sdExit3(t, "empty queue under --wait", err) },
		},
		{
			name: "permanent nominate refusal",
			setup: func(t *testing.T) {
				nmBase(t, factory.BacklogStateQueued)
				nmLaneEnv(t, "lane-1", "")
			},
			args: []string{"--run", fcRun, "--card", "t9"},
			check: func(t *testing.T, out, stderr string, err error) {
				nmAssertRefused(t, out, stderr, err, "unknown-card")
			},
		},
		{
			name:  "quota hold, bare",
			setup: func(t *testing.T) { qasFixture(t, qasFixtureOpts{five: qasWin(92, qasReset5)}) },
			args:  []string{"--run", fcRun},
			check: func(t *testing.T, out, stderr string, err error) { qasAssertHeld(t, out, stderr, err) },
		},
		{
			name:  "quota hold, nominated",
			setup: func(t *testing.T) { qasFixture(t, qasFixtureOpts{five: qasWin(92, qasReset5)}) },
			args:  []string{"--run", fcRun, "--card", "t1"},
			check: func(t *testing.T, out, stderr string, err error) {
				nmAssertRefused(t, out, stderr, err, "quota-hold")
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(t)
			gh := flFirstGhCall(t, func() (string, error) { return "[]", nil })
			out, stderr, err := qasRunNext(t, tc.args...)
			tc.check(t, out, stderr, err)
			if at := gh(); !at.IsZero() {
				t.Errorf("a run that leased nothing issued the open-PR query (gh pr list); want zero gh calls (stdout %q, stderr %q)", out, stderr)
			}
		})
	}
}

// TestFactoryNextWaitQueriesOpenPRsForTheSuccessfulAttempt — with --wait, a lease
// that succeeds only after a wait prints the pull requests as they stand at that
// attempt: a PR opened during the wait reads `linked`, not `no-link`. The query
// for that attempt precedes the attempt's queue lock (the lease-wait seam fires
// right before each lease call's lock wait), and the polls that leased nothing
// issued none. The wait is a quota hold that releases after the first sleep, in
// the bare and the nominated form.
func TestFactoryNextWaitQueriesOpenPRsForTheSuccessfulAttempt(t *testing.T) {
	for _, form := range []string{"nominated", "bare"} {
		t.Run(form, func(t *testing.T) {
			root, _ := nmBase(t, factory.BacklogStateQueued)
			qasEnableGate(t, root)
			qasLaneEnv(t, "lane-1", factory.BackendClaude)
			nmIsolatedWorktrees(t, "t1")

			events := &flEventLog{path: filepath.Join(t.TempDir(), "events")}
			var prOpened atomic.Bool
			flFirstGhCall(t, func() (string, error) {
				events.add("gh")
				if prOpened.Load() {
					return `[{"number":42,"title":"t1 delivers the card","body":"","state":"OPEN"}]`, nil
				}
				return "[]", nil
			})
			prevWaitStart := factoryLeaseWaitStart
			factoryLeaseWaitStart = func(time.Time) { events.add("lock") }
			t.Cleanup(func() { factoryLeaseWaitStart = prevWaitStart })

			readings := []float64{92, 84}
			sleeps := qasFakeClock(t, nil, func(n int) {
				if n == 1 {
					prOpened.Store(true)
				}
			})
			qasInjectReadings(t, func() float64 { return readings[min(*sleeps, len(readings)-1)] })

			args := []string{"--run", fcRun, "--wait", "--wait-bound", "1h"}
			if form == "nominated" {
				args = append(args, "--card", "t1")
			}
			out, stderr, err := qasRunNext(t, args...)
			if err != nil {
				t.Fatalf("next --wait: %v (stderr %q)", err, stderr)
			}
			nmAssertLeased(t, root, "t1", "lane-1")
			if *sleeps != 1 {
				t.Errorf("wait slept %d times, want 1 (held at 92, released at 84)", *sleeps)
			}
			if !strings.Contains(out, "t1\tlinked\t#42") {
				t.Errorf("the output does not show the PR opened during the wait as linked (#42):\n%s", out)
			}
			// The held poll takes the queue lock and queries nothing; the successful attempt
			// queries first and takes the lock after.
			if got, want := strings.Join(events.lines(t), ","), "lock,gh,lock"; got != want {
				t.Errorf("event order = %q, want %q (a poll that leases nothing queries nothing; the successful attempt's query precedes its queue lock)", got, want)
			}
		})
	}
}
