// factory_lease_class_test.go — SPEC-FACTORY-LEASE-BUSY-001 (card t1916):
// the cause-accurate lease refusal classification. The lease path's claim
// failure classes each carry their own reason token — a race lost to another
// claimant (`raced`), a record write held busy past every retry
// (`record-busy`), and the claim's own deadline expired under a still-running
// write (`claim-deadline`) — and the two non-race classes are wait-retryable
// exactly as `raced` is. The queue-lock-held `raced` refusal is out of scope
// (decision-index Q4) and is pinned unchanged by
// TestFactoryLeaseQueueLockStallBounded.
package cli

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// TestFactoryLeaseClaimClassesWaitRetryable — AC-FLB-011's token table: the
// record-contention and claim-deadline tokens are wait-retryable exactly as
// `raced` is, so `--wait` keeps waiting through transient contention instead
// of stopping on the first refusal (REQ-FLB-008). The permanent refusals stay
// permanent.
func TestFactoryLeaseClaimClassesWaitRetryable(t *testing.T) {
	for _, tc := range []struct {
		token string
		want  bool
	}{
		{"raced", true},
		{"record-busy", true},
		{"claim-deadline", true},
		{"serial-slot", true},
		{"quota-hold", true},
		{"unknown-card", false},
		{"held", false},
	} {
		r := &factoryNominateRefusal{Token: tc.token, Detail: "detail"}
		if got := r.waitable(); got != tc.want {
			t.Errorf("waitable(%q) = %v, want %v", tc.token, got, tc.want)
		}
	}
}

// flRealBusyError holds the factory record open and returns a real SQLite
// busy/locked error from a raw write on a second bounded connection: a raw
// ExecContext gets no retryFactoryBusy wrap, so the busy error returns while
// the caller's deadline is still live — the record-contention shape (the
// deadline is the claim's own context, which the probe does not share).
func flRealBusyError(t *testing.T, root string) error {
	t.Helper()
	db, err := homestate.OpenFactoryBounded(root, factoryLeaseClaimBusyTimeout)
	if err != nil {
		t.Fatalf("open the record for the busy probe: %v", err)
	}
	defer func() { _ = db.Close() }()
	_, execErr := db.DB.ExecContext(context.Background(), `UPDATE cards SET version = version WHERE 0`)
	if execErr == nil {
		t.Fatal("the busy probe write returned nil, want a real SQLite busy error against the held record")
	}
	if !homestate.IsBusy(execErr) {
		t.Fatalf("the busy probe write returned %v, want the SQLite busy/locked class", execErr)
	}
	return execErr
}

// TestFactoryLeaseClaimFailureMarksDistinctClasses — the marking seam
// (factoryClaimFailure) splits the two non-race classes: a busy error under a
// live claim deadline is the record-contention class, and a busy error joined
// with the expired claim deadline is the claim-deadline class (the deadline is
// the outer bound — spec § Constraints) whose detail names the observed
// signals: the expired claim deadline and the presence of the concurrent busy
// error, asserting neither a load-only nor a race-only cause (REQ-FLB-003,
// REQ-FLB-004).
func TestFactoryLeaseClaimFailureMarksDistinctClasses(t *testing.T) {
	root, _ := nmQueuedNominee(t)
	end := flHoldRecord(t, root)
	defer end()
	busyErr := flRealBusyError(t, root)

	// (b) — busy while the claim's own deadline is still live.
	contention := factoryClaimFailure(context.Background(), busyErr)
	if !strings.Contains(contention.Error(), "held the record") {
		t.Errorf("the contention marking = %v, want a detail naming the record lock", contention)
	}
	if strings.Contains(contention.Error(), "claim deadline") {
		t.Errorf("a live-deadline busy failure was marked with the deadline class: %v", contention)
	}

	// (c) — the deadline signal joined with the busy error: the deadline class
	// takes precedence, and its detail names both signals.
	dctx, dcancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer dcancel()
	<-dctx.Done()
	joined := factoryClaimFailure(dctx, busyErr)
	if !strings.Contains(joined.Error(), "the claim deadline") {
		t.Errorf("the joined marking = %v, want a detail naming the expired claim deadline", joined)
	}
	if !strings.Contains(joined.Error(), "busy error was present: true") {
		t.Errorf("the joined marking = %v, want its detail to name the concurrent busy error's presence", joined)
	}
	if strings.Contains(joined.Error(), "another writer held the record past every retry") {
		t.Errorf("the joined marking = %v, want the contention-only wording absent (the deadline class takes precedence)", joined)
	}
	if !strings.Contains(joined.Error(), "neither a host-load-only nor a race-only cause") {
		t.Errorf("the joined marking = %v, want it to assert neither a load-only nor a race-only cause", joined)
	}
}

// TestFactoryLeaseDeadlineClassRefusalToken — AC-FLB-001 (c) and AC-FLB-002:
// a claim deadline expired under a still-running record write refuses with the
// claim-deadline token (never `raced`), exit 4, and the detail names the
// expired deadline and the concurrent busy error whose presence the run
// observed (short deadline override past the 200 ms busy timeout, so at least
// one SQLITE_BUSY surfaced before the deadline fired).
func TestFactoryLeaseDeadlineClassRefusalToken(t *testing.T) {
	root, _ := nmQueuedNominee(t)
	nmIsolatedWorktrees(t, "t1")
	end := flHoldRecord(t, root)
	defer end()
	prev := factoryLeaseClaimDeadlineOverride
	factoryLeaseClaimDeadlineOverride = 2 * factoryLeaseClaimBusyTimeout
	t.Cleanup(func() { factoryLeaseClaimDeadlineOverride = prev })

	_, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
	nmAssertRefused(t, "", stderr, err, "claim-deadline")
	if strings.Contains(stderr, "refused raced") {
		t.Errorf("the deadline-class refusal still reports raced: %q", stderr)
	}
	if !strings.Contains(stderr, "the claim deadline") || !strings.Contains(stderr, "busy error was present: true") {
		t.Errorf("the refusal detail does not name the observed signals (the expired claim deadline, the concurrent busy error present): %q", stderr)
	}
}

// TestFactoryLeaseRecordContentionClassRefusalToken — AC-FLB-001 (b): a claim
// failure whose cause is a record write held busy past every retry, with the
// claim deadline still live, refuses with the record-busy token (never
// `raced`), exit 4, and a detail naming the record lock. The marked busy error
// reaches the nominated lease through the nomination seam — the same marked
// error the claim path produces (the marking itself is pinned by
// TestFactoryLeaseClaimFailureMarksDistinctClasses).
func TestFactoryLeaseRecordContentionClassRefusalToken(t *testing.T) {
	root, _ := nmQueuedNominee(t)
	nmIsolatedWorktrees(t, "t1")
	end := flHoldRecord(t, root)
	defer end()
	busyErr := flRealBusyError(t, root)
	nmSetSeam(t, func(string) error { return factoryClaimFailure(context.Background(), busyErr) })

	_, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
	nmAssertRefused(t, "", stderr, err, "record-busy")
	if strings.Contains(stderr, "refused raced") {
		t.Errorf("the record-contention refusal still reports raced: %q", stderr)
	}
	if !strings.Contains(stderr, "held the record") {
		t.Errorf("the refusal detail does not name the record lock: %q", stderr)
	}
}

// TestFactoryLeaseRefusalTextCapture — AC-FLB-004's capture vehicle: it
// renders each preserved refusal class's stderr line so the pre-change tree's
// lines can be recorded verbatim in progress.md §E.2 (with their tree SHA) and
// the post-change tree re-run against them (recorded-prior-text comparison;
// REQ-FLB-007's text preservation). It asserts only the refusal contract
// itself; the byte-level comparison lives in the recorded evidence pair.
func TestFactoryLeaseRefusalTextCapture(t *testing.T) {
	for _, name := range []string{"serial-slot", "quota-hold", "backend-skip", "foreign-worktree"} {
		t.Run(name, func(t *testing.T) {
			c := nmCaseByName(t, name)[0]
			root, store, nominee := c.prepare(t)
			before := nmSnapshot(t, root, store)
			_, stderr, err := qasRunNext(t, "--run", fcRun, "--card", nominee)
			nmAssertRefused(t, "", stderr, err, c.token)
			t.Logf("CAPTURED %s: %s", name, strings.TrimSpace(stderr))
			if after := nmSnapshot(t, root, store); after != before {
				t.Errorf("the refused nomination changed the store state (before %q, after %q)", before, after)
			}
		})
	}
}
