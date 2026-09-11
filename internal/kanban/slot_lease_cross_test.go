package kanban

// slot_lease_cross_test.go — AC-RSL-001a (REQ-RSL-003, card t607): the control
// group. Two separate OS processes, two different session ids, one resource.
//
//   - probe_then_start: without the surface, each child checks "is anyone using
//     the resource?" and then records a start. Child A is held between its
//     check and its write while child B runs whole. Both start (starts=2). This
//     is the POSITIVE control: it proves the harness can observe a double start
//     at all, so the lease branch's refusal is interpretable. It must pass in
//     every tree state.
//   - lease: the same constructed interleaving, now inside AcquireSlotLease
//     between the decision and the write (inside the per-resource mutation
//     lock). Under a correct lock B waits on the lock, A is released by the
//     stall timeout, and B then reads A's record: acquired=1 refused=1.
//
// The interleaving is CONSTRUCTED, never waited for (same reasoning and the
// same stall-release arithmetic as integration_lock_cross_test.go:46-63), and
// both children record the PARENT's pid as owner so neither record reads
// stale — a child recording its own pid would exit, go stale, and be taken
// over legitimately, producing a double success under a correct lock.

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	// slotStallReleaseTimeout releases child A when B has not finished. It
	// must be at most a third of the mutation-lock wait budget so B keeps at
	// least two thirds of the budget after A is released (asserted below).
	slotStallReleaseTimeout = 500 * time.Millisecond

	// slotStallMarkerWait bounds the wait for A to report STALLED; exceeding
	// it is a harness fault, not a race to retry.
	slotStallMarkerWait = 20 * time.Second

	// slotHelperDeadline is the external deadline on every child, paired with
	// a t.Cleanup-registered kill.
	slotHelperDeadline = 60 * time.Second

	// slotHelperStallCeiling bounds the child's own wait inside the hook.
	slotHelperStallCeiling = 30 * time.Second
)

var slotOutcomeRE = regexp.MustCompile(`RESULT=(\S+) SESSION=(\S+)`)

type slotOutcome struct {
	result  string
	session string
	raw     string
	err     error
}

// slotChildEnv is os.Environ() with the lane identity variables removed, so a
// child never inherits the real session id or owner pid of the runtime.
func slotChildEnv(extra ...string) []string {
	var env []string
	for _, kv := range os.Environ() {
		switch {
		case strings.HasPrefix(kv, "CLAUDE_PROJECT_DIR="),
			strings.HasPrefix(kv, "CLAUDE_CODE_SESSION_ID="),
			strings.HasPrefix(kv, "MOAI_SESSION_PID="):
			continue
		}
		env = append(env, kv)
	}
	return append(env, extra...)
}

// runSlotHelper runs one child op and returns its parsed outcome. It never
// calls t.Fatalf (it runs in goroutines); the caller asserts.
func runSlotHelper(t *testing.T, env ...string) slotOutcome {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), slotHelperDeadline)
	t.Cleanup(cancel)
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestSlotLeaseHelperProcess$", "--")
	cmd.Env = slotChildEnv(env...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Start(); err != nil {
		return slotOutcome{err: err}
	}
	t.Cleanup(func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	})
	waitErr := cmd.Wait()
	out := slotOutcome{raw: buf.String(), err: waitErr}
	if m := slotOutcomeRE.FindStringSubmatch(out.raw); m != nil {
		out.result, out.session = m[1], m[2]
	}
	return out
}

// constructSlotRound runs the constructed interleaving for one op: A stalls at
// the interleaving point, B runs, A is released on B's completion or the stall
// timeout, whichever comes first.
func constructSlotRound(t *testing.T, op string, common ...string) (a, b slotOutcome) {
	t.Helper()
	flags := t.TempDir()
	marker := filepath.Join(flags, "stalled")
	proceed := filepath.Join(flags, "proceed")

	aDone := make(chan slotOutcome, 1)
	bDone := make(chan slotOutcome, 1)
	go func() {
		aDone <- runSlotHelper(t, append([]string{
			"MOAI_SLOT_HELPER=" + op, "HELPER_SESSION=lane-a",
			"HELPER_STALL_MARKER=" + marker, "HELPER_PROCEED_FLAG=" + proceed,
		}, common...)...)
	}()

	// Wait for STALLED, but also notice A finishing without ever reaching the
	// interleaving point — that is a harness fault with its own report.
	deadline := time.Now().Add(slotStallMarkerWait)
	for {
		if _, err := os.Stat(marker); err == nil {
			break
		}
		select {
		case early := <-aDone:
			t.Fatalf("[%s] child A finished without reaching the interleaving point — nothing was constructed. A: %q (err %v)", op, firstLine(early.raw), early.err)
		default:
		}
		if time.Now().After(deadline) {
			_ = os.WriteFile(proceed, []byte("go\n"), 0o644)
			late := <-aDone
			t.Fatalf("[%s] child A never reported STALLED within %s — harness fault. A: %q (err %v)", op, slotStallMarkerWait, late.raw, late.err)
		}
		time.Sleep(5 * time.Millisecond)
	}

	go func() {
		bDone <- runSlotHelper(t, append([]string{"MOAI_SLOT_HELPER=" + op, "HELPER_SESSION=lane-b"}, common...)...)
	}()
	bReceived := false
	select {
	case b = <-bDone:
		bReceived = true
	case <-time.After(slotStallReleaseTimeout):
	}
	if err := os.WriteFile(proceed, []byte("go\n"), 0o644); err != nil {
		t.Fatalf("releasing child A: %v", err)
	}
	if !bReceived {
		b = <-bDone
	}
	a = <-aDone
	for label, r := range map[string]slotOutcome{"A": a, "B": b} {
		if r.result == "" {
			t.Fatalf("[%s] child %s produced no outcome line: %q (err %v)", op, label, r.raw, r.err)
		}
	}
	if a.session == b.session {
		t.Fatalf("[%s] both children reported SESSION=%q — the round measured the harness, not the race", op, a.session)
	}
	return a, b
}

// AC-RSL-001a — probe-then-start admits both sessions; the lease admits one.
func TestSlotLease_ControlGroupTwoSessions(t *testing.T) {
	if runtimeIsWindows() {
		t.Skip("helper re-exec plumbing exercised on unix; windows substrate covered by GOOS=windows build")
	}
	if slotStallReleaseTimeout*3 > boardLockWaitBudget {
		t.Fatalf("stall-release timeout %s exceeds a third of the %s wait budget — B would be left retrying for most of its budget and busy would re-enter as flakiness", slotStallReleaseTimeout, boardLockWaitBudget)
	}
	ownerPID := strconv.Itoa(os.Getpid())

	t.Run("probe_then_start", func(t *testing.T) {
		starts := t.TempDir()
		a, b := constructSlotRound(t, "probe-start", "HELPER_STARTS_DIR="+starts)
		started := 0
		for _, r := range []slotOutcome{a, b} {
			if r.result == "started" {
				started++
			}
		}
		t.Logf("control: starts=%d (A: %s | B: %s)", started, firstLine(a.raw), firstLine(b.raw))
		if started != 2 {
			t.Fatalf("control: starts=%d, want 2 — the harness cannot observe a double start, so the lease branch's refusal cannot be interpreted", started)
		}
	})

	t.Run("lease", func(t *testing.T) {
		root := t.TempDir()
		a, b := constructSlotRound(t, "slot-acquire", "HELPER_ROOT="+root, "HELPER_OWNER_PID="+ownerPID)
		acquired, refused, busy, other := 0, 0, 0, 0
		winner := ""
		for _, r := range []slotOutcome{a, b} {
			switch r.result {
			case "acquired":
				acquired++
				winner = r.session
			case "held":
				refused++
			case "busy":
				busy++
			default:
				other++
			}
		}
		t.Logf("lease: acquired=%d refused=%d busy=%d other=%d (A: %s | B: %s)", acquired, refused, busy, other, firstLine(a.raw), firstLine(b.raw))
		if busy > 0 {
			t.Fatalf("lease: busy=%d — a harness configuration fault, not a verdict: the stall-release timeout left B without budget", busy)
		}
		if other > 0 {
			t.Fatalf("lease: other=%d — a child reported neither acquired nor held. A: %q B: %q", other, a.raw, b.raw)
		}
		if acquired != 1 || refused != 1 {
			t.Fatalf("lease: acquired=%d refused=%d, want 1 and 1 — the record's read-modify-write is not serialized across processes", acquired, refused)
		}
		final, err := ReadSlotLease(root, slotTestResource)
		if err != nil {
			t.Fatalf("reading the final record: %v", err)
		}
		if final.SessionID != winner {
			t.Errorf("final holder = %q, want the acquiring child %q", final.SessionID, winner)
		}
		if final.Stale() {
			t.Errorf("final record reads stale — the round exercised the stale-reclaim path, not serialization")
		}
	})
}

// TestSlotLeaseHelperProcess is not a real test: with MOAI_SLOT_HELPER unset it
// returns immediately. When set, the child performs the named op and exits;
// its stdout line is the observation.
func TestSlotLeaseHelperProcess(t *testing.T) {
	op := os.Getenv("MOAI_SLOT_HELPER")
	if op == "" {
		return
	}
	session := os.Getenv("HELPER_SESSION")
	stall := func() {
		marker := os.Getenv("HELPER_STALL_MARKER")
		if marker == "" {
			return
		}
		if err := os.WriteFile(marker, []byte("STALLED\n"), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "stall marker: %v\n", err)
			return
		}
		proceed := os.Getenv("HELPER_PROCEED_FLAG")
		deadline := time.Now().Add(slotHelperStallCeiling)
		for {
			if _, err := os.Stat(proceed); err == nil {
				return
			}
			if time.Now().After(deadline) {
				fmt.Fprintln(os.Stderr, "proceed flag never appeared within the child ceiling")
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
	}

	switch op {
	case "probe-start":
		dir := os.Getenv("HELPER_STARTS_DIR")
		entries, err := os.ReadDir(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "probe: %v\n", err)
			os.Exit(2)
		}
		if len(entries) > 0 {
			_, _ = fmt.Fprintf(os.Stdout, "RESULT=skipped SESSION=%s\n", session)
			os.Exit(0)
		}
		stall() // between the check and the start — the probe-then-start gap
		if err := os.WriteFile(filepath.Join(dir, session), []byte("start\n"), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "start: %v\n", err)
			os.Exit(2)
		}
		_, _ = fmt.Fprintf(os.Stdout, "RESULT=started SESSION=%s\n", session)
		os.Exit(0)
	case "slot-acquire":
		ownerPID, err := strconv.Atoi(os.Getenv("HELPER_OWNER_PID"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "HELPER_OWNER_PID: %v\n", err)
			os.Exit(10)
		}
		if os.Getenv("HELPER_STALL_MARKER") != "" {
			slotLeaseMutationTestHook = stall
		}
		_, acqErr := AcquireSlotLease(os.Getenv("HELPER_ROOT"), SlotLeaseRequest{
			Resource:    slotTestResource,
			SessionID:   session,
			PID:         ownerPID,
			Command:     "heavy-suite",
			MaxDuration: time.Hour,
		})
		result := "error"
		switch {
		case acqErr == nil:
			result = "acquired"
		case IsSlotLeaseHeld(acqErr):
			result = "held"
		case IsSlotLeaseBusy(acqErr):
			result = "busy"
		}
		_, _ = fmt.Fprintf(os.Stdout, "RESULT=%s SESSION=%s\n", result, session)
		if result == "error" {
			fmt.Fprintf(os.Stderr, "slot-acquire: %v\n", acqErr)
		}
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "unknown slot helper op %q\n", op)
		os.Exit(64)
	}
}
