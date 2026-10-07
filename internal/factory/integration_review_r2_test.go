package factory

// integration_review_r2_test.go — card-review ROUND 2 RED reproductions
// (card t1479): class-based sweeps A–G. One RED per class where
// reproducible; call-site inventories live in the class disposition
// comments.

import (
	"os"
	"strings"
	"testing"
	"time"
)

// TestR2_A_LiteralPathspecBracketCollision is the NEW P1 RED: the candidate
// adds the leaf `[a]`; the worktree holds an ignored regular file `[a]` and
// TRACKS `a`. git ls-files/clean interpret `[a]` as a PATHSPEC GLOB
// (character class matching `a`), so the tracked-file probe answers
// "tracked" for a path that is NOT the tracked one — the check misses the
// collision and the merge destroys `[a]`. The fix: :(literal) pathspecs on
// every ls/clean probe in the check.
func TestR2_A_LiteralPathspecBracketCollisionDetected(t *testing.T) {
	r := newR1Repo(t)
	r.git("commit", "-q", "--allow-empty", "-m", "base")
	// Track the file `a` — the glob-lookalike that `[a]` resolves onto.
	r.write("a", "tracked-a")
	r.git("add", "a")
	r.git("commit", "-q", "-m", "track a")
	r.write(".gitignore", "[a]\n")
	r.git("add", ".gitignore")
	r.git("commit", "-q", "-m", "ignore bracket")
	r.fromBranch("cand")
	r.write("[a]/payload", "p")
	r.git("add", "-f", "[a]/payload")
	r.git("commit", "-q", "-m", "cand adds bracket leaf")
	r.git("checkout", "-q", "main")
	ignoredPath := r.write("[a]", "secretbytes")
	before := checksumOf(t, ignoredPath)

	// Sanity: the bracket path IS glob-interpreted by ls-files (the
	// precondition the fix answers) — the tracked file `a` answers for it.
	probe := r.git("ls-files", "--", "[a]")
	if probe != "a" {
		t.Fatalf("fixture: ls-files -- [a] must resolve onto the tracked `a` under glob interpretation, got %q", probe)
	}

	got, err := FindAddedPathCollisions(r.dir, r.git("rev-parse", "main"), r.git("rev-parse", "cand"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatalf("RED (P1 NEW, class A): the bracket-path collision must be detected under literal pathspecs; got %v", got)
	}
	if after := checksumOf(t, ignoredPath); after != before {
		t.Fatalf("the ignored bytes must be untouched: before=%s after=%s", before, after)
	}
}

// TestR2_A_LiteralPathspecBeneathProbe pins the beneath-probe leg of class
// A: the beneath probe (`git clean -nd -x -- <dir>`) must probe the LITERAL
// directory name — a bracket-named ignored directory (`[a]`, whose name is
// a character class) must answer through its own bytes, never through the
// glob resolving onto a tracked `a`.
func TestR2_A_LiteralPathspecBeneathProbe(t *testing.T) {
	r := newR1Repo(t)
	r.git("commit", "-q", "--allow-empty", "-m", "base")
	// Track `a` (the glob target) and hold an ignored DIRECTORY `[a]` with
	// bytes inside.
	r.write("a", "tracked-a")
	r.git("add", "a")
	r.git("commit", "-q", "-m", "track a")
	r.write(".gitignore", "[a]/secret\n")
	r.git("add", ".gitignore")
	r.git("commit", "-q", "-m", "ignore bracket dir")
	secretPath := r.write("[a]/secret", "secretbytes")
	before := checksumOf(t, secretPath)

	// The literal beneath probe must see the bracket directory's bytes.
	beneath, err := untrackedBytesBeneath(r.dir, "[a]")
	if err != nil {
		t.Fatal(err)
	}
	if !beneath {
		t.Fatalf("RED (class A beneath leg): the literal beneath probe must see the ignored bytes inside [a]")
	}
	if after := checksumOf(t, secretPath); after != before {
		t.Fatalf("the ignored bytes must be untouched: before=%s after=%s", before, after)
	}

	// The glob under a literal probe answers NOTHING for a name that only
	// exists as the bracket directory: probing the tracked-file name `a`
	// (which the bracket would resolve onto under glob interpretation) must
	// not be how the beneath probe behaves.
	if res := r.git("ls-files", "--", ":(literal)[a]"); res != "" {
		t.Fatalf("the literal pathspec must not match the tracked `a` for the bracket path, got %q", res)
	}
}

// TestR2_B_WaitLoopRereadsPolicyInMutation pins class B at the factory
// layer the wait loop composes: EnqueueTicket and PromotedAfterBound run
// INSIDE their mutations with the policy read INSIDE the mutation — a hold
// written between the loop's entry and a mutation governs that mutation.
// The CLI-level red (policy captured at loop entry) is repaired by moving
// the read into each UpdateIntegrationWindow callback.
func TestR2_B_MutationsRereadPolicyInside(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)
	prevClock := WindowClock
	WindowClock = func() time.Time { return now }
	t.Cleanup(func() { WindowClock = prevClock })
	pid := os.Getpid()

	// The wait path's enqueue callback reads the policy INSIDE the
	// mutation: write hold BEFORE the mutation runs and the enqueue must
	// not promote even though the caller read open earlier.
	if err := WriteIntegrationWindowPolicy(root, IntegrationWindowPolicy{Policy: PolicyHold, Reason: "late-hold"}); err != nil {
		t.Fatal(err)
	}
	err := UpdateIntegrationWindow(root, func(w *IntegrationLock) error {
		policy, policyErr := ReadIntegrationWindowPolicy(root)
		if policyErr != nil {
			return policyErr
		}
		return EnqueueTicket(w, IntegrationTicket{SessionID: "sess-b", OwnerPID: pid, WaiterPID: pid, WaiterStart: currentProcessFingerprint()},
			DefaultWindowProcProbe(), now, policy)
	})
	if err != nil {
		t.Fatal(err)
	}
	lock, _ := ReadIntegrationLock(root)
	if lock.SessionID == "sess-b" {
		t.Fatalf("RED (class B): a hold written before the mutation must govern it; sess-b holds the window")
	}
}

// TestR2_C_MergeRefusesWhenWindowTakenMidStep is the P1 RED for class C:
// the collision gate passes, then --force takes the window, and the step's
// merge must NOT proceed on a holdership it no longer holds. The
// AfterPrecheck seam is exactly the injection point between the gates and
// the merge.
func TestR2_C_MergeRefusesWhenWindowTakenMidStep(t *testing.T) {
	f := newMergeFixture(t)
	card := f.withCardTree(readyCardPtr())
	seams := f.seams(card)
	takerPID := os.Getpid()
	seams.AfterPrecheck = func() {
		// A competing session force-takes the window right before the
		// merge — recorded, deliberate, and it must govern the step.
		if _, err := AcquireIntegrationWindow(f.root, IntegrationLock{
			SessionID: "sess-taker", SessionName: "lane-taker",
			PID: takerPID, PIDSource: PIDSourceSessionOwner,
			Branch: "develop", BranchSource: BranchSourceConfig,
			Worktree: f.integ, Card: "other",
		}, true, nil); err != nil {
			t.Errorf("fixture: force takeover failed: %v", err)
		}
	}
	_, err := RunMergeStep(f.input(), seams)
	if err == nil {
		t.Fatalf("RED (class C): the merge must refuse once the window was taken mid-step")
	}
	got, ok := MergeExitCode(err)
	if !ok || (got != MergeExitNotHolder && got != MergeExitExpiredLease) {
		t.Fatalf("the mid-step takeover must refuse with a holder code, got %d (err=%v)", got, err)
	}
	// No merge commit: the integration branch stays at develop.
	head := strings.TrimSpace(stepMustGit(t, f.integ, "rev-parse", "HEAD"))
	parents := strings.Fields(stepMustGit(t, f.integ, "rev-list", "--parents", "-n", "1", head))
	if len(parents) == 3 {
		t.Fatalf("no merge commit may exist after a mid-step takeover: %v", parents)
	}
}
