package cli

// The claim-deadline contract (card t1696). The 2026-10-14 load-32 storm
// (run tmvgvn) refused every lane's lease with
// "the factory record stayed busy for 1s: sql: transaction has already been
// committed or rolled back" — the claim's own deadline killing its write
// transaction, reported as if another writer held the record. The refusal
// must name the deadline class, and the post-rollback tx-done text must never
// leak again.

import (
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/worktree"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestFactoryClaimDeadlineRefusalNamesTheDeadline(t *testing.T) {
	_, _ = nmQueuedNominee(t)
	prev := factoryLeaseClaimDeadlineOverride
	factoryLeaseClaimDeadlineOverride = time.Nanosecond
	t.Cleanup(func() { factoryLeaseClaimDeadlineOverride = prev })

	_, stderr, err := qasRunNext(t, "--run", fcRun)
	if err == nil {
		t.Fatal("the lease succeeded with a 1 ns claim deadline, want a refusal")
	}
	combined := stderr + err.Error()
	if strings.Contains(combined, "transaction has already been committed or rolled back") {
		t.Fatalf("the refusal leaked the post-rollback tx-done text: %v (stderr %q)", err, stderr)
	}
	// The claim-deadline class names its observed signals (the expired claim
	// deadline, and the concurrent busy error's presence or absence —
	// SPEC-FACTORY-LEASE-BUSY-001 REQ-FLB-004); it asserts neither a
	// load-only nor a race-only cause.
	if !strings.Contains(combined, "the claim deadline") || !strings.Contains(combined, "busy error was present: false") {
		t.Fatalf("the refusal does not name the deadline class and its observed signals: %v (stderr %q)", err, stderr)
	}
	if strings.Contains(combined, "another lane") || strings.Contains(combined, "the host is under load") {
		t.Fatalf("the refusal asserts a load-only or race-only cause: %v (stderr %q)", err, stderr)
	}
}

// The nominated surface reports the same class through the claim-deadline
// refusal line, carrying the claim error's own words
// (SPEC-FACTORY-LEASE-BUSY-001: the pre-reclassification pin said `raced`).
func TestFactoryNominatedBusyRefusalCarriesTheClaimError(t *testing.T) {
	_, _ = nmQueuedNominee(t)
	prev := factoryLeaseClaimDeadlineOverride
	factoryLeaseClaimDeadlineOverride = time.Nanosecond
	t.Cleanup(func() { factoryLeaseClaimDeadlineOverride = prev })

	_, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
	if nmExit(err) != factoryNextRefusedExit {
		t.Fatalf("exit = %d (err %v stderr %q), want %d", nmExit(err), err, stderr, factoryNextRefusedExit)
	}
	if !strings.Contains(stderr, "refused claim-deadline:") || !strings.Contains(stderr, "t1 was not leased") {
		t.Fatalf("the refusal line lost its shape: %q", stderr)
	}
	if !strings.Contains(stderr, "the claim deadline") {
		t.Fatalf("the refusal line does not carry the claim error's class: %q", stderr)
	}
}

// A worktree step that fails AFTER the lease committed must say the lease
// stands: the lane-20 shape (card t1696) — the lane read a bare failure and
// retried a card it already held.
func TestFactoryNextPostLeaseWorktreeFailureNamesTheLease(t *testing.T) {
	root, _ := nmQueuedNominee(t)
	prevCreator := worktree.WorktreeCreator
	worktree.WorktreeCreator = func(cardID string, out io.Writer) (string, error) {
		return "", errors.New("git exploded under load")
	}
	t.Cleanup(func() { worktree.WorktreeCreator = prevCreator })

	_, stderr, err := qasRunNext(t, "--run", fcRun)
	if err == nil {
		t.Fatal("the lease reported success while the worktree step failed, want the honest failure")
	}
	combined := stderr + err.Error()
	if !strings.Contains(combined, "was leased") || !strings.Contains(combined, "t1") {
		t.Fatalf("the failure does not name the standing lease: %v (stderr %q)", err, stderr)
	}
	c := fcCard(t, root, "t1")
	if c.State != homestate.CardLeased {
		t.Fatalf("card t1 is %s, want leased (the lease stands)", c.State)
	}
}
