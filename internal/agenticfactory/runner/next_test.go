package runner

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func handedOffOwner(t *testing.T) (*Runner, core.IssueRef, Ticket) {
	t.Helper()
	r, ref := newRunner(t)
	tk := landed(t, r, ref)
	if _, err := r.HandOff(tk, goodHandoff()); err != nil {
		t.Fatal(err)
	}
	return r, ref, tk
}

// After a PR the owner may take another Issue, but the repair of the first one
// comes first: a repair signal on any of its open PRs holds new claims.
func TestRepairSignalHoldsNewClaimsUntilTheOwnerReopens(t *testing.T) {
	r, ref, tk := handedOffOwner(t)
	next, _ := core.NewIssueRef(2000)
	if _, err := r.Claim(next, "sess-a", "disp-2", "run-1"); err != nil {
		t.Fatalf("no repair is waiting, so a second Issue is allowed: %v", err)
	}
	if err := r.Store.Release(next, "sess-a"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Store.MarkRepairNeeded(ref, ""); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Errorf("a signal needs its reason: %v", err)
	}
	if rec, err := r.Store.MarkRepairNeeded(ref, "ci failed on latest head"); err != nil || rec.RepairNeeded == "" || rec.Holder != "sess-a" || rec.Epoch != tk.Epoch {
		t.Fatalf("signal: %+v %v", rec, err)
	}
	if _, err := r.Claim(next, "sess-a", "disp-3", "run-1"); !errors.Is(err, core.ErrRepairFirst) {
		t.Errorf("new work before the repair: %v", err)
	}
	// another agent's claim of another Issue is not held by this owner's repair
	if _, err := r.Claim(next, "sess-b", "disp-4", "run-1"); err != nil {
		t.Errorf("a different agent is not blocked: %v", err)
	}
	_ = r.Store.Release(next, "sess-b")
	rec, err := r.ReopenOwn(tk, "ci failed on latest head", RepairEvidence{ScopeReserved: true, OldWriterEnded: true})
	if err != nil || rec.RepairNeeded != "" {
		t.Fatalf("reopen takes the repair up and clears the signal: %+v %v", rec, err)
	}
	// the repair is under way and not handed off again: it is still the first job
	if _, err := r.Claim(next, "sess-a", "disp-5", "run-1"); !errors.Is(err, core.ErrRepairFirst) {
		t.Errorf("claim while the repair runs: %v", err)
	}
}

func TestSignalNeedsAnOpenPR(t *testing.T) {
	r, ref := newRunner(t)
	if _, err := r.Store.MarkRepairNeeded(ref, "x"); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("unknown Issue: %v", err)
	}
	tk := landed(t, r, ref)
	if _, err := r.Store.MarkRepairNeeded(ref, "x"); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("an Issue without a receipt has no PR to repair: %v", err)
	}
	_ = tk
}

// The new Issue is written in its own worktree: never the one holding the open
// PR's branch, never one another open Issue already uses.
func TestNewWorkAfterAPRNeedsItsOwnWorktree(t *testing.T) {
	r, _, _ := handedOffOwner(t)
	next, _ := core.NewIssueRef(2000)
	nt, err := r.Claim(next, "sess-a", "disp-2", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.StartAttempt(nt, "attempt-9", core.DeliveryExecuted); !errors.Is(err, core.ErrWorktreeUnbound) {
		t.Errorf("writing without a bound worktree while another PR is open: %v", err)
	}
	if _, err := r.BindWorktree(nt, ""); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Errorf("empty worktree: %v", err)
	}
	if _, err := r.BindWorktree(nt, goodHandoff().Worktree); !errors.Is(err, core.ErrWorktreeInUse) {
		t.Errorf("the open PR's worktree: %v", err)
	}
	if _, err := r.BindWorktree(nt, wtDir("second")); err != nil {
		t.Fatalf("a separate worktree: %v", err)
	}
	if _, err := r.StartAttempt(nt, "attempt-9", core.DeliveryExecuted); err != nil {
		t.Errorf("start with the bound worktree: %v", err)
	}
	third, _ := core.NewIssueRef(3000)
	t3, err := r.Claim(third, "sess-a", "disp-3", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.BindWorktree(t3, wtDir("second")); !errors.Is(err, core.ErrWorktreeInUse) {
		t.Errorf("a worktree another open Issue uses: %v", err)
	}
}

// With no PR open the first Issue needs no bound worktree: the rule only guards
// the second one.
func TestFirstIssueNeedsNoBoundWorktree(t *testing.T) {
	r, ref := newRunner(t)
	tk, err := r.Claim(ref, "sess-a", "disp-1", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.StartAttempt(tk, "attempt-1", core.DeliveryExecuted); err != nil {
		t.Errorf("first Issue: %v", err)
	}
}

// The open-PR limit is configuration with no default: unset holds every new
// claim after a PR, and a full backlog holds them too.
func TestOpenPRLimitStillGatesNewClaimsAfterAPR(t *testing.T) {
	r, _, _ := handedOffOwner(t)
	_ = r.Store.SetMaxOpenPRs(1)
	next, _ := core.NewIssueRef(2000)
	if _, err := r.Claim(next, "sess-a", "disp-2", "run-1"); !errors.Is(err, core.ErrBacklogFull) {
		t.Errorf("limit reached: %v", err)
	}
}

// The receipt's worktree is held like a bound one, through the same create-only
// reservation: while the PR is open nobody else binds the path, and when the
// Issue is done the reservation is stale and the path is free again.
func TestTheReceiptWorktreeIsReservedUntilTheIssueIsDone(t *testing.T) {
	r, ref, tk := handedOffOwner(t)
	path := goodHandoff().Worktree
	doc, err := r.Store.GetDoc("worktree", claim.WorktreeKey(path))
	if err != nil {
		t.Fatalf("the handoff must reserve the receipt worktree: %v", err)
	}
	if !strings.Contains(string(doc.Data), path) {
		t.Errorf("reservation: %s", doc.Data)
	}
	next, _ := core.NewIssueRef(2000)
	nt, err := r.Claim(next, "sess-b", "disp-2", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.BindWorktree(nt, path); !errors.Is(err, core.ErrWorktreeInUse) {
		t.Errorf("an open PR's worktree: %v", err)
	}
	if _, err := r.Complete(tk, Landing{RemoteMerged: true, LocalSyncRecorded: true, IssueClosed: true, Deploy: "N/A: x"}); err != nil {
		t.Fatal(err)
	}
	_ = ref
	if _, err := r.BindWorktree(nt, path); err != nil {
		t.Errorf("a finished Issue's reservation must be taken over: %v", err)
	}
}

// A caller that does not own the Issue must not be able to park a path on it.
func TestAStrangersHandoffReservesNothing(t *testing.T) {
	r, ref := newRunner(t)
	tk := landed(t, r, ref)
	h := goodHandoff()
	h.Worktree = wtDir("stranger")
	if _, err := r.HandOff(Ticket{Issue: ref, Holder: "someone", Epoch: tk.Epoch}, h); !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Fatalf("a stranger's handoff: %v", err)
	}
	if _, err := r.Store.GetDoc("worktree", claim.WorktreeKey(wtDir("stranger"))); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("a refused handoff left a reservation: %v", err)
	}
}

// The receipt's worktree is compared by what it is, not how it is spelled: a PR
// handed off from a symlink to a tree blocks binding the tree itself. A worktree
// that does not exist, or is not absolute, cannot be handed off from.
func TestReceiptWorktreeSpellingsAreOneWorktree(t *testing.T) {
	r, ref := newRunner(t)
	tk := landed(t, r, ref)
	real := wtDir("spelled")
	link := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	h := goodHandoff()
	for name, p := range map[string]string{"missing": filepath.Join(wtRoot, "does-not-exist"), "relative": "wt/rel"} {
		h.Worktree = p
		if _, err := r.HandOff(tk, h); !errors.Is(err, core.ErrEvidenceIncomplete) {
			t.Errorf("%s worktree: %v", name, err)
		}
		if cur, _ := r.Store.Read(ref); cur.Receipt != nil {
			t.Fatalf("%s: a refused handoff recorded a receipt", name)
		}
	}
	h.Worktree = link
	if _, err := r.HandOff(tk, h); err != nil {
		t.Fatal(err)
	}
	next, _ := core.NewIssueRef(2000)
	nt, err := r.Claim(next, "sess-b", "disp-2", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.BindWorktree(nt, real); !errors.Is(err, core.ErrWorktreeInUse) {
		t.Errorf("the tree behind a handed-off symlink: %v", err)
	}
}
