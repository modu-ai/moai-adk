// integration_lock_cross_test.go — AC-ILA-001 / AC-ILA-002 (REQ-ILA-001/002/003,
// card t336): the integration-lock record's read-modify-write is serialized
// across OS PROCESSES.
//
// ONE test function, run in two tree states. It asserts the GREEN invariant
// (exactly one holder, the other refused). Run against the unrepaired mutation
// path — before M2 exists, or after M2 with the critical section disabled per
// the AC-ILA-006 one-line revert — it FAILS, and its failure report carries the
// attributed double-hold that IS the RED observation. There is no second,
// phantom RED-named function.
//
// The interleaving is CONSTRUCTED, not waited for. The unserialized window is
// one read, a branch, and one write — tens of microseconds — so a
// barrier-released pair hits it a few percent of the time at best, and a
// criterion that waits for luck has no stop rule: widening until RED appears is
// tuning until luck arrives, and stopping early proceeds with no observation.
// Child A is therefore HELD inside the nil-by-default interleaving hook between
// its decision and its write while child B runs its entire acquire.
//
// Attribution is what makes the observation mean the race rather than something
// adjacent. Three things are asserted about a double-hold before it counts:
//   - REPLACED=none on BOTH children — a takeover is a stale reclaim, not the
//     read-modify-write race (the iteration-1 D1 hazard);
//   - the two children's SESSION ids DIFFER — the same-session path is a LEGAL
//     re-acquire (integration_lock.go's "re-acquiring a window the caller
//     already holds succeeds"), and it produces successes=2 with REPLACED=none
//     on both and a live winner, passing every other check while never
//     exercising the race (audit finding N2);
//   - the record reads non-Stale() — the recorded holder is alive, so this is a
//     live double-hold and not two callers each reclaiming a dead record.
package kanban

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"
	"time"
)

const (
	// integrationStallReleaseTimeout releases child A when B has not finished.
	// The timeout is a LIVENESS bound, not a race window: under the repair B
	// blocks on the mutation lock until A releases it, so waiting only for B
	// would deadlock.
	//
	// Its size is load-bearing and is stated with margin rather than as a bare
	// strict inequality (audit finding N3). The mutation-lock wait budget is
	// boardLockWaitBudget = boardLockSupportedWriters × boardLockCIMutationCost
	// × boardLockHeadroom = 10 × 33ms × 5 = 1.65s (board_store.go:96-117).
	// 500ms is 30.3% of that budget — inside the "at most a third" headroom the
	// audit asked for — so after A is released B still has ~1.15s (69.7%) of
	// budget left to win the lock. A timeout merely SHORTER than the budget
	// (1.6s, say) satisfies the inequality while leaving B retrying under
	// jitter for essentially its whole budget, and one scheduling hiccup then
	// yields RESULT=busy — which AC-ILA-002 fails as a misconfigured harness.
	// That is flakiness re-entering a MUST-PASS criterion by a different door.
	integrationStallReleaseTimeout = 500 * time.Millisecond

	// integrationStallMarkerWait bounds how long the parent waits for A to
	// report STALLED. Generous: it covers a cold `go test` child start, and
	// exceeding it means the hook never fired, which is a harness fault to
	// report rather than a race to retry.
	integrationStallMarkerWait = 20 * time.Second

	// integrationHelperDeadline is the external deadline on every child, paired
	// with a t.Cleanup-registered kill. A trailing kill at the end of a test
	// body is NOT cleanup: every early-return path skips it.
	integrationHelperDeadline = 60 * time.Second
)

// integrationAcquireOutcome classifies an acquire error into the helper's
// RESULT vocabulary. It lives here rather than in the helper so the busy
// sentinel can join it when the mutation lock lands, without re-editing the
// child dispatch.
func integrationAcquireOutcome(err error) string {
	switch {
	case err == nil:
		return "acquired"
	case IsIntegrationLockHeld(err):
		return "held"
	default:
		return "error"
	}
}

// integrationOutcome is one child's parsed report line.
type integrationOutcome struct {
	result   string
	replaced string
	session  string
	raw      string
}

var integrationOutcomeRE = regexp.MustCompile(`RESULT=(\S+) REPLACED=(\S+) SESSION=(\S+)`)

// runIntegrationAcquireHelper runs one acquire in a separate OS process and
// returns its parsed outcome. It never calls t.Fatalf: it is invoked from
// goroutines, where that is not permitted — the caller asserts.
func runIntegrationAcquireHelper(t *testing.T, root, session string, ownerPID int, stallMarker, proceedFlag string) (integrationOutcome, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), integrationHelperDeadline)
	t.Cleanup(cancel)

	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestKanbanHelperProcess", "--")
	cmd.Env = append(os.Environ(),
		"MOAI_KANBAN_HELPER=integration-acquire",
		"HELPER_ROOT="+root,
		"HELPER_SESSION="+session,
		"HELPER_OWNER_PID="+strconv.Itoa(ownerPID),
	)
	if stallMarker != "" {
		cmd.Env = append(cmd.Env,
			"HELPER_STALL_MARKER="+stallMarker,
			"HELPER_PROCEED_FLAG="+proceedFlag,
		)
	}
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Start(); err != nil {
		return integrationOutcome{}, err
	}
	// Registered BEFORE Wait, so every early return — including a t.Fatalf in
	// the caller while this child is still stalled — reaps the process.
	t.Cleanup(func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	})
	waitErr := cmd.Wait()

	out := integrationOutcome{raw: buf.String()}
	if m := integrationOutcomeRE.FindStringSubmatch(out.raw); m != nil {
		out.result, out.replaced, out.session = m[1], m[2], m[3]
		return out, nil
	}
	return out, waitErr
}

// waitForFile polls for path until deadline. Used for the STALLED marker only.
func waitForFile(path string, within time.Duration) bool {
	deadline := time.Now().Add(within)
	for {
		if _, err := os.Stat(path); err == nil {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(5 * time.Millisecond)
	}
}

type integrationChildResult struct {
	out integrationOutcome
	err error
}

// TestIntegrationLockAcquire_SerializedAcrossProcesses — two separate OS
// processes acquire the same free window under a CONSTRUCTED interleaving:
// child A is held between its decision and its write while child B runs its
// whole acquire. Exactly one must be recorded as holder; the other must be
// refused with ErrIntegrationLockHeld.
func TestIntegrationLockAcquire_SerializedAcrossProcesses(t *testing.T) {
	if runtimeIsWindows() {
		t.Skip("helper re-exec plumbing exercised on unix; windows substrate covered by GOOS=windows build")
	}
	root := t.TempDir()
	flags := t.TempDir()
	stallMarker := filepath.Join(flags, "stalled")
	proceedFlag := filepath.Join(flags, "proceed")

	// The PARENT's pid, alive for the whole round by construction. A child
	// recording its own pid exits immediately after writing, its record then
	// reads STALE, and the second child takes it over legitimately —
	// successes=2 even under a correct lock.
	ownerPID := os.Getpid()

	aDone := make(chan integrationChildResult, 1)
	bDone := make(chan integrationChildResult, 1)

	go func() {
		o, err := runIntegrationAcquireHelper(t, root, "lane-a", ownerPID, stallMarker, proceedFlag)
		aDone <- integrationChildResult{out: o, err: err}
	}()

	if !waitForFile(stallMarker, integrationStallMarkerWait) {
		_ = os.WriteFile(proceedFlag, []byte("go\n"), 0o644)
		a := <-aDone
		t.Fatalf("child A never reported STALLED inside the interleaving hook within %s — the interleaving was not constructed, so this round observed nothing about the lock. A output: %q (err %v)",
			integrationStallMarkerWait, a.out.raw, a.err)
	}

	go func() {
		o, err := runIntegrationAcquireHelper(t, root, "lane-b", ownerPID, "", "")
		bDone <- integrationChildResult{out: o, err: err}
	}()

	var b integrationChildResult
	bReceived := false
	select {
	case b = <-bDone:
		bReceived = true
	case <-time.After(integrationStallReleaseTimeout):
	}

	// Read the record at the moment the second acquire has had its chance,
	// BEFORE A is released. Under the unrepaired path B has already written,
	// so this observes the first winner's record; under the repair B is still
	// blocked on the mutation lock, so the record is legitimately absent.
	midRecord, midErr := ReadIntegrationLock(root)
	if midErr != nil {
		t.Errorf("reading the record between the two acquires: %v", midErr)
	}
	midHeld := midRecord != nil && midRecord.Held()
	midStale := midHeld && midRecord.Stale()

	if err := os.WriteFile(proceedFlag, []byte("go\n"), 0o644); err != nil {
		t.Fatalf("releasing the stalled child: %v", err)
	}
	if !bReceived {
		b = <-bDone
	}
	a := <-aDone

	for label, r := range map[string]integrationChildResult{"A": a, "B": b} {
		if r.out.result == "" {
			t.Fatalf("child %s produced no outcome line — the round measured nothing. output: %q (err %v)", label, r.out.raw, r.err)
		}
	}

	// Distinctness is ASSERTED from the children's own reported ids, never
	// assumed from this setup: a harness bug passing one id to both children
	// would otherwise yield a false RED that survives every other check.
	if a.out.session == b.out.session {
		t.Fatalf("both children reported SESSION=%q — the same-session path is a LEGAL re-acquire, not the race; this round measured the harness, not the lock", a.out.session)
	}

	successes, refusals, other := 0, 0, 0
	for _, r := range []integrationChildResult{a, b} {
		switch r.out.result {
		case "acquired":
			successes++
		case "held":
			refusals++
		default:
			other++
		}
	}

	final, err := ReadIntegrationLock(root)
	if err != nil {
		t.Fatalf("reading the persisted record: %v", err)
	}
	finalStale := final.Held() && final.Stale()

	t.Logf("A: %s", firstLine(a.out.raw))
	t.Logf("B: %s", firstLine(b.out.raw))
	t.Logf("round: successes=%d refusals=%d other=%d sessions_differ=%v mid_record_held=%v mid_record_stale=%v final_holder=%q final_record_stale=%v",
		successes, refusals, other, a.out.session != b.out.session, midHeld, midStale, final.SessionID, finalStale)

	if other > 0 {
		t.Fatalf("a child reported neither acquired nor held (successes=%d refusals=%d other=%d) — RESULT=busy means the stall-release timeout (%s) was not comfortably shorter than the mutation-lock wait budget, and the harness measured its own configuration rather than the lock. A: %q B: %q",
			successes, refusals, other, integrationStallReleaseTimeout, firstLine(a.out.raw), firstLine(b.out.raw))
	}

	if successes != 1 || refusals != 1 {
		attributed := a.out.replaced == "none" && b.out.replaced == "none" &&
			a.out.session != b.out.session && !finalStale && !midStale
		t.Fatalf("successes=%d refusals=%d, want exactly 1 and 1 — the record's read-modify-write is NOT serialized across processes.\n"+
			"  A: %s\n  B: %s\n"+
			"  attributed_double_hold=%v (REPLACED=none on both: %v; session ids differ: %v; record non-Stale at the second acquire: mid_held=%v mid_stale=%v; final record non-Stale: %v)\n"+
			"  final record: holder=%q pid=%d pid_source=%q",
			successes, refusals,
			firstLine(a.out.raw), firstLine(b.out.raw),
			attributed,
			a.out.replaced == "none" && b.out.replaced == "none",
			a.out.session != b.out.session,
			midHeld, midStale, !finalStale,
			final.SessionID, final.PID, final.PIDSource)
	}

	if !final.Held() {
		t.Fatalf("no record persisted after the round — one child reported acquired, so the window must name it")
	}
	if finalStale {
		t.Fatalf("persisted record reads STALE (holder %q pid %d) — the round exercised the stale-reclaim path, not serialization", final.SessionID, final.PID)
	}
}

// firstLine returns the first non-empty line of s, for legible failure reports.
func firstLine(s string) string {
	for _, line := range bytes.Split([]byte(s), []byte("\n")) {
		if len(bytes.TrimSpace(line)) > 0 {
			return string(bytes.TrimSpace(line))
		}
	}
	return "(no output)"
}
