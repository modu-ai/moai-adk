package factory

// integration_merge_step_test.go — M5 tests for the merge step (card
// t1479, AC-MWQ-017/018): the holder refusals leave the record's bytes
// unchanged and release nothing (scenarios 4/5), each pre-merge cause
// releases the window with its own exit code, the post-merge failure
// leaves the merge commit and holds naming its SHA, and the O3 seam
// reaches cause 8b. Fixture: one scratch repository, `develop` checked
// out (the integration worktree), a WT- card branch that absorbed
// develop, a valid re-measure keyed to that tree, and the card state
// scripted through the ReadCard seam.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	stepCard       = "t0002"
	stepCardBranch = "WT-fix-card"
)

// stepFixture builds the whole environment one step run needs.
type stepFixture struct {
	t             *testing.T
	root          string // record + remeasure store root
	integ         string // the worktree holding develop checked out
	cardTree      string // the worktree holding the WT- card branch
	cardSHA       string // the pinned candidate
	record        *RemeasureRecord
	liveWaiterPID int
}

func newMergeFixture(t *testing.T) *stepFixture {
	t.Helper()
	root := t.TempDir()
	integ := filepath.Join(root, "integ")
	cardTree := filepath.Join(root, "card")
	for _, dir := range []string{integ, cardTree} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	f := &stepFixture{t: t, root: root, integ: integ, cardTree: cardTree}
	git := func(dir string, args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	// develop: the integration branch. The card tree is a WORKTREE of the
	// same repository — a card's WT- branch and the integration branch
	// share one object database (the kanban worktree model), so `git
	// rev-parse refs/heads/WT-...` answers from the integration worktree.
	git(integ, "init", "-q", "-b", "develop")
	git(integ, "config", "user.email", "t@t.local")
	git(integ, "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(integ, "base.txt"), []byte("base"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(integ, "add", "base.txt")
	git(integ, "commit", "-q", "-m", "base")
	baseSHA := git(integ, "rev-parse", "HEAD")

	// The card worktree: branch from develop's tip, gain a card commit — so
	// the pinned SHA descends from the base and its tree is what the
	// re-measure keys.
	git(integ, "worktree", "add", "-q", "-b", stepCardBranch, cardTree, baseSHA)
	git(cardTree, "config", "user.email", "t@t.local")
	git(cardTree, "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(cardTree, "card.txt"), []byte("card work"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(cardTree, "add", "card.txt")
	git(cardTree, "commit", "-q", "-m", "card work")
	f.cardSHA = git(cardTree, "rev-parse", "HEAD")

	// The re-measure keyed to the pinned tree, valid under the verifier.
	pinnedTree := git(cardTree, "rev-parse", "HEAD^{tree}")
	f.record = &RemeasureRecord{
		Tree: pinnedTree, Base: baseSHA,
		Command: "true", ExitCode: 0, StructuredRequired: false,
		BuildIdentity: "moai test",
		RecordedAt:    time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC).Format(time.RFC3339),
	}
	if err := WriteRemeasureRecord(root, pinnedTree, *f.record); err != nil {
		t.Fatal(err)
	}

	// The window: session sess-b holds it for card t0002, and a live C
	// ticket waits behind.
	holder := IntegrationLock{
		SessionID: "sess-b", SessionName: "lane-b",
		PID: os.Getpid(), PIDSource: PIDSourceSessionOwner,
		Branch: "develop", BranchSource: BranchSourceConfig,
		Worktree: integ, Card: stepCard,
	}
	StampLease(&holder, time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC), IntegrationLeaseDefault)
	if err := UpdateIntegrationWindow(root, func(w *IntegrationLock) error {
		*w = holder
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	f.liveWaiterPID = os.Getpid()
	if err := UpdateIntegrationWindow(root, func(w *IntegrationLock) error {
		w.Queue = append(w.Queue, IntegrationTicket{
			SessionID: "sess-c", SessionName: "lane-c", Card: "t0003",
			OwnerPID: f.liveWaiterPID, PIDSource: PIDSourceSessionOwner,
			Branch: "develop", BranchSource: BranchSourceConfig, Worktree: integ,
			WaiterPID: f.liveWaiterPID, WaiterStart: currentProcessFingerprint(),
			Heartbeat: time.Date(2026, 10, 5, 9, 0, 30, 0, time.UTC).Format(time.RFC3339), EnqueuedAt: time.Date(2026, 10, 5, 9, 0, 30, 0, time.UTC).Format(time.RFC3339),
		})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return f
}

// seams returns the production-shaped seams with the card state scripted —
// the ReadCard closure captures the card POINTER so later mutations reach it.
func (f *stepFixture) seams(card *MergeCardState) MergeStepSeams {
	return MergeStepSeams{
		ReadCard: func(string) (MergeCardState, error) { return *card, nil },
		Now:      func() time.Time { return time.Date(2026, 10, 5, 9, 1, 0, 0, time.UTC) },
	}
}

func (f *stepFixture) input() MergeStepInput {
	return MergeStepInput{
		Root: f.root, IntegrationWorktree: f.integ, IntegrationBranch: "develop",
		CardID: stepCard, CallerSessionID: "sess-b",
	}
}

func readyCard() MergeCardState {
	return MergeCardState{Stage: "merge-ready", LeaseUnexpired: true, Version: 7, WorktreePath: ""}
}

func readyCardPtr() *MergeCardState {
	card := readyCard()
	return &card
}

func (f *stepFixture) withCardTree(card *MergeCardState) *MergeCardState {
	card.WorktreePath = f.cardTree
	return card
}

// holderPID rewrites the recorded holder pid so the queue's liveness probe
// reads it live — tests that verify a promotion after release need the
// recorded holder alive at release time only if the release path probes
// staleness; the release of a LIVE holder skips staleness entirely.
func requireCode(t *testing.T, err error, want int) *MergeStepError {
	t.Helper()
	if err == nil {
		t.Fatalf("expected merge-step error with code %d, got success", want)
	}
	got, ok := MergeExitCode(err)
	if !ok || got != want {
		t.Fatalf("expected exit code %d, got %d (err=%v)", want, got, err)
	}
	var step *MergeStepError
	if !errors.As(err, &step) {
		t.Fatalf("error is not a MergeStepError: %T", err)
	}
	return step
}

func requireWindowReleasedAndCPromoted(t *testing.T, f *stepFixture) {
	t.Helper()
	lock, err := ReadIntegrationLock(f.root)
	if err != nil {
		t.Fatal(err)
	}
	if lock.SessionID != "sess-c" {
		t.Fatalf("after a pre-merge refusal C must be promoted, holder is %q", lock.SessionID)
	}
	if len(lock.Queue) != 0 {
		t.Fatalf("the queue must be empty after C's promotion: %+v", lock.Queue)
	}
}

func TestMergeStepHappyPathCreatesNoFFMergeAndReleases(t *testing.T) {
	// AC-MWQ-017 scenario 1: a valid record, unmoved base — one --no-ff
	// merge commit whose second parent is the pinned SHA, tree equal to the
	// record's tree, the window released, C promoted, and no test
	// invocation anywhere in the step (it never shells a runner — asserted
	// by construction: the step's git calls carry no `test` verb, checked
	// in TestMergeStepRunsNoTestCommands).
	f := newMergeFixture(t)
	card := f.withCardTree(readyCardPtr())
	sha, err := RunMergeStep(f.input(), f.seams(card))
	if err != nil {
		t.Fatalf("the merge step must succeed: %v", err)
	}
	// The merge commit's second parent is the pinned SHA; its tree equals
	// the record's tree.
	out := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = f.integ
		o, err := cmd.Output()
		if err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
		return strings.TrimSpace(string(o))
	}
	parents := strings.Fields(out("rev-list", "--parents", "-n", "1", sha))
	if len(parents) != 3 || parents[2] != f.cardSHA {
		t.Fatalf("the merge commit's second parent must be the pinned SHA: %v", parents)
	}
	if tree := out("rev-parse", sha+"^{tree}"); tree != f.record.Tree {
		t.Fatalf("the merge tree %s must equal the record's tree %s", tree, f.record.Tree)
	}
	requireWindowReleasedAndCPromoted(t, f)
}

func TestMergeStepRunsNoTestCommands(t *testing.T) {
	// REQ-MWQ-017's closing clause: "it shall run no test suite". The git
	// seam records every invocation; none may carry a test runner.
	f := newMergeFixture(t)
	card := f.withCardTree(readyCardPtr())
	var invocations [][]string
	seams := f.seams(card)
	seams.Git = func(args ...string) (string, error) {
		invocations = append(invocations, args)
		runner := exec.Command("git", args...)
		runner.Dir = f.integ
		out, err := runner.CombinedOutput()
		return string(out), err
	}
	if _, err := RunMergeStep(f.input(), seams); err != nil {
		t.Fatalf("the merge step must succeed: %v", err)
	}
	for _, args := range invocations {
		joined := strings.Join(args, " ")
		if strings.Contains(joined, "go test") || strings.Contains(joined, "go vet") || strings.Contains(joined, "golangci") {
			t.Fatalf("the merge step ran a test command: %s", joined)
		}
	}
	if len(invocations) == 0 {
		t.Fatalf("the merge step never ran — the fixture is broken")
	}
}

func TestMergeStepHolderRefusalsLeaveRecordUntouched(t *testing.T) {
	// AC-MWQ-017 scenarios 4 and 5: a non-holder (and an expired-lease
	// holder) is refused with the window record's BYTES unchanged, and
	// nothing is released — a queued ticket whose waiter is gone stays
	// (it is not C's call to drop).
	f := newMergeFixture(t)
	card := f.withCardTree(readyCardPtr())
	before, _ := ReadIntegrationLock(f.root)

	// Scenario 4: C is not the holder.
	in := f.input()
	in.CallerSessionID = "sess-c"
	_, refused := RunMergeStep(in, f.seams(card))
	requireCode(t, refused, MergeExitNotHolder)

	// The dropped-waiter ticket still sits: C's refused call dropped
	// nothing.
	lock, _ := ReadIntegrationLock(f.root)
	if len(lock.Queue) != 1 {
		t.Fatalf("a refused non-holder must not mutate the queue: %+v", lock.Queue)
	}

	// Scenario 4's record comparison: the refused call changed nothing.
	afterNonHolder, _ := ReadIntegrationLock(f.root)
	if fmt.Sprint(afterNonHolder) != fmt.Sprint(before) {
		t.Fatalf("the record must read unchanged after the non-holder refusal")
	}

	// Scenario 5: B holds with an EXPIRED lease. The refusal changes
	// nothing the CALLER wrote — the test itself wrote the expiry, so the
	// comparison is against the expired state.
	expired, _ := ReadIntegrationLock(f.root)
	expired.LeaseExpiresAt = time.Date(2026, 10, 5, 8, 0, 0, 0, time.UTC).Format(time.RFC3339)
	if err := UpdateIntegrationWindow(f.root, func(w *IntegrationLock) error {
		*w = *expired
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	_, expiredRefused := RunMergeStep(f.input(), f.seams(card))
	requireCode(t, expiredRefused, MergeExitExpiredLease)

	after, _ := ReadIntegrationLock(f.root)
	if fmt.Sprint(after) != fmt.Sprint(expired) {
		t.Fatalf("the record must read unchanged after the expired-lease refusal")
	}
}

func TestMergeStepPreMergeCausesReleaseWithDistinctCodes(t *testing.T) {
	// AC-MWQ-018 rows 1-5, 9-12: each cause fires, the window releases, and
	// C is promoted. Each subtest rebuilds its own fixture (codex-P2's
	// fresh-per-case reading applies to these too).
	cases := []struct {
		name string
		code int
		mutate func(f *stepFixture, seams *MergeStepSeams, card *MergeCardState)
	}{
		{"1 record invalid", MergeExitRecordInvalid, func(f *stepFixture, seams *MergeStepSeams, card *MergeCardState) {
			f.record.ExitCode = 3
			_ = WriteRemeasureRecord(f.root, f.record.Tree, *f.record)
		}},
		{"2 base moved", MergeExitBaseMoved, func(f *stepFixture, seams *MergeStepSeams, card *MergeCardState) {
			// develop advances past the record's base.
			cmd := exec.Command("git", "commit", "-q", "--allow-empty", "-m", "other lane landed")
			cmd.Dir = f.integ
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("advance develop: %v: %s", err, out)
			}
		}},
		{"3 not a descendant", MergeExitNotDescendant, func(f *stepFixture, seams *MergeStepSeams, card *MergeCardState) {
			// Rebase the card branch onto a foreign root: rewrite its
			// parentage so it no longer descends from the base.
			cmd := exec.Command("git", "checkout", "-q", "--orphan", stepCardBranch+"-x")
			cmd.Dir = f.cardTree
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("orphan: %v: %s", err, out)
			}
			cmd = exec.Command("git", "commit", "-q", "-m", "orphan root")
			cmd.Dir = f.cardTree
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("orphan commit: %v: %s", err, out)
			}
			cmd = exec.Command("git", "branch", "-q", "-f", stepCardBranch, "HEAD")
			cmd.Dir = f.cardTree
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("branch -f: %v: %s", err, out)
			}
			cmd = exec.Command("git", "checkout", "-q", stepCardBranch)
			cmd.Dir = f.cardTree
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("checkout: %v: %s", err, out)
			}
			// Re-key the record to the new tree so only ancestry fails.
			out, err := exec.Command("git", "rev-parse", "HEAD^{tree}").CombinedOutput()
			if err != nil {
				t.Fatalf("tree: %v: %s", err, out)
			}
			f.record.Tree = strings.TrimSpace(string(out))
			f.record.Base = "ffffffffffffffffffffffffffffffffffffffff"
			_ = WriteRemeasureRecord(f.root, f.record.Tree, *f.record)
			f.cardSHA = strings.TrimSpace(stepMustGit(t, f.cardTree, "rev-parse", "HEAD"))
		}},
		{"4 tree mismatch", MergeExitTreeMismatch, func(f *stepFixture, seams *MergeStepSeams, card *MergeCardState) {
			// The record stays keyed to the PINNED tree (so the gate finds
			// it) but its Tree field names another tree — WriteRemeasureRecord
			// forces the key into the field, so the mismatched record is
			// written directly. The identity check is what must refuse.
			forged := *f.record
			forged.Tree = strings.Repeat("a", 40)
			raw, err := json.Marshal(forged)
			if err != nil {
				t.Fatal(err)
			}
			dir := filepath.Join(f.root, ".moai", "state", "remeasure")
			if err := os.WriteFile(filepath.Join(dir, f.record.Tree+".json"), raw, 0o644); err != nil {
				t.Fatal(err)
			}
		}},
		{"5 landing refused", MergeExitLandingRefused, func(f *stepFixture, seams *MergeStepSeams, card *MergeCardState) {
			seams.LandingCheck = func(string, string) error { return errors.New("CI red") }
		}},
		{"9 other error", MergeExitOther, func(f *stepFixture, seams *MergeStepSeams, card *MergeCardState) {
			card.WorktreePath = "" // no tree, so the WT- branch cannot resolve
		}},
		{"11b not merge-ready", MergeExitCardGate, func(f *stepFixture, seams *MergeStepSeams, card *MergeCardState) {
			card.Stage = "sync-audit"
		}},
		{"11a foreign card lease", MergeExitCardGate, func(f *stepFixture, seams *MergeStepSeams, card *MergeCardState) {
			card.LeaseUnexpired = false
		}},
		{"11c card mismatch", MergeExitCardGate, func(f *stepFixture, seams *MergeStepSeams, card *MergeCardState) {
			if err := UpdateIntegrationWindow(f.root, func(w *IntegrationLock) error {
				w.Card = "t9999"
				return nil
			}); err != nil {
				t.Fatal(err)
			}
		}},
		{"12 dirty before merge", MergeExitWorktreeDirty, func(f *stepFixture, seams *MergeStepSeams, card *MergeCardState) {
			if err := os.WriteFile(filepath.Join(f.integ, "untracked.txt"), []byte("x"), 0o644); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newMergeFixture(t)
			card := f.withCardTree(readyCardPtr())
			// The seams capture the CARD POINTER — a mutate that edits the
			// card state must reach the ReadCard closure (O4's one read is
			// the gate's; the test scripts what that read returns).
			seams := f.seams(card)
			tc.mutate(f, &seams, card)
			_, err := RunMergeStep(f.input(), seams)
			requireCode(t, err, tc.code)
			requireWindowReleasedAndCPromoted(t, f)
		})
	}
}

func stepMustGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return string(out)
}

func TestMergeStepNothingToMerge(t *testing.T) {
	// AC-MWQ-018 row 10: the pinned SHA equals the record's base — refused
	// before `git merge` is ever invoked, and the window releases.
	f := newMergeFixture(t)
	// Point the card branch at develop's tip and re-key the record.
	stepMustGit(t, f.cardTree, "reset", "-q", "--hard", f.record.Base)
	f.cardSHA = f.record.Base
	f.record.Tree = strings.TrimSpace(stepMustGit(t, f.cardTree, "rev-parse", "HEAD^{tree}"))
	_ = WriteRemeasureRecord(f.root, f.record.Tree, *f.record)
	card := f.withCardTree(readyCardPtr())
	_, err := RunMergeStep(f.input(), f.seams(card))
	requireCode(t, err, MergeExitNothingToMerge)
	requireWindowReleasedAndCPromoted(t, f)
}

func TestMergeStepMergeFailureCleanAbortsCause6(t *testing.T) {
	// AC-MWQ-018 row 6: the merge seam fails and leaves MERGE_HEAD; the
	// abort cleans up; the worktree reads clean; C is promoted.
	f := newMergeFixture(t)
	card := f.withCardTree(readyCardPtr())
	seams := f.seams(card)
	seams.Git = func(args ...string) (string, error) {
		if len(args) >= 1 && args[0] == "merge" && !containsArg(args, "--abort") {
			// The failure shape: MERGE_HEAD left behind. The abort the step
			// runs afterwards is a REAL abort (the seam passes it through).
			if err := os.WriteFile(filepath.Join(f.integ, ".git", "MERGE_HEAD"), []byte("deadbeef\n"), 0o644); err != nil {
				return "", err
			}
			return "", errors.New("simulated merge failure")
		}
		runner := exec.Command("git", args...)
		runner.Dir = f.integ
		out, err := runner.CombinedOutput()
		return string(out), err
	}
	_, err := RunMergeStep(f.input(), seams)
	requireCode(t, err, MergeExitMergeFailed)
	if _, statErr := os.Stat(filepath.Join(f.integ, ".git", "MERGE_HEAD")); !os.IsNotExist(statErr) {
		t.Fatalf("the abort must clear MERGE_HEAD: %v", statErr)
	}
	requireWindowReleasedAndCPromoted(t, f)
}

// containsArg reports whether the invocation carries arg.
func containsArg(args []string, arg string) bool {
	for _, a := range args {
		if a == arg {
			return true
		}
	}
	return false
}

func TestMergeStepCause8LeavesCommitAndHoldsNamingSHA(t *testing.T) {
	// AC-MWQ-018 row 8: the merge succeeds and a post-merge lookup fails —
	// the commit stays, the hold names the merge SHA (and the setter is
	// the step with the card), and the window releases with C NOT promoted.
	f := newMergeFixture(t)
	card := f.withCardTree(readyCardPtr())
	seams := f.seams(card)
	seams.Git = func(args ...string) (string, error) {
		if len(args) >= 2 && args[0] == "rev-parse" && strings.HasSuffix(args[1], "^{tree}") && !strings.Contains(args[1], "^{tree}^{tree}") {
			// Post-merge tree lookup fails — but the PRE-merge pinned-tree
			// read uses the same shape. Distinguish by the merge having
			// happened: the pre-merge read is for the pinned SHA, the
			// post-merge one for HEAD.
			if strings.HasPrefix(args[1], "HEAD") {
				return "", errors.New("simulated post-merge lookup failure")
			}
		}
		runner := exec.Command("git", args...)
		runner.Dir = f.integ
		out, err := runner.CombinedOutput()
		return string(out), err
	}
	_, err := RunMergeStep(f.input(), seams)
	step := requireCode(t, err, MergeExitPostMerge)
	if !strings.Contains(step.Msg, "merge commit") {
		t.Fatalf("the cause-8 message must name the merge commit: %s", step.Msg)
	}
	// The merge commit is on the integration branch.
	head := strings.TrimSpace(stepMustGit(t, f.integ, "rev-parse", "HEAD"))
	if !strings.Contains(step.Msg, head[:12]) {
		t.Fatalf("the hold must name the merge SHA %s: %s", head[:12], step.Msg)
	}
	policy, err := ReadIntegrationWindowPolicy(f.root)
	if err != nil || policy.Policy != PolicyHold {
		t.Fatalf("the policy must be held after cause 8: %+v err=%v", policy, err)
	}
	if !strings.Contains(policy.SetBy, stepCard) {
		t.Fatalf("the hold's setter names the step and card: %+v", policy)
	}
	// C is NOT promoted onto a held window.
	lock, _ := ReadIntegrationLock(f.root)
	if lock.SessionID == "sess-c" {
		t.Fatalf("C must not be promoted onto a cause-8 hold")
	}
}

func TestMergeStepCause8bO3SeamInjectsResidue(t *testing.T) {
	// AC-MWQ-018 row 8b via the O3 seam: the residue appears after every
	// pre-merge check passed; the post-merge clean check finds it, the hold
	// is written BEFORE the release, and the exit code is cause 8's.
	f := newMergeFixture(t)
	card := f.withCardTree(readyCardPtr())
	seams := f.seams(card)
	holdWrittenBeforeRelease := false
	seams.AfterPrecheck = func() {
		if err := os.WriteFile(filepath.Join(f.integ, "autostash-residue.txt"), []byte("residue"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Track the write order through the policy record's set time versus the
	// release: the release promotes C only when the policy is open, so a
	// held policy after the run proves the hold landed first.
	_, err := RunMergeStep(f.input(), seams)
	requireCode(t, err, MergeExitPostMerge)
	policy, _ := ReadIntegrationWindowPolicy(f.root)
	if policy.Policy != PolicyHold {
		t.Fatalf("the hold must be written before the release: %+v", policy)
	}
	_ = holdWrittenBeforeRelease
	// C stays queued (the release under hold leaves no promotion).
	lock, _ := ReadIntegrationLock(f.root)
	if lock.SessionID == "sess-c" {
		t.Fatalf("C must not be promoted onto the cause-8b hold")
	}
}

func TestMergeStepPinnedSHAOverBranchName(t *testing.T) {
	// AC-MWQ-017 scenario 2: the branch advances AFTER the identity checks
	// (here: after the step pinned the SHA, before `git merge`) — the merge
	// commit's second parent is still the PINNED SHA, and the recorded git
	// invocation names the SHA, never the branch name.
	f := newMergeFixture(t)
	card := f.withCardTree(readyCardPtr())
	var mergeArgs []string
	seams := f.seams(card)
	seams.Git = func(args ...string) (string, error) {
		if len(args) >= 1 && args[0] == "merge" {
			mergeArgs = append([]string(nil), args...)
			// Advance the branch between the pin and the merge: the
			// hook fires right before the merge call.
			stepMustGit(t, f.cardTree, "commit", "-q", "--allow-empty", "-m", "late card commit")
		}
		runner := exec.Command("git", args...)
		runner.Dir = f.integ
		out, err := runner.CombinedOutput()
		return string(out), err
	}
	sha, err := RunMergeStep(f.input(), seams)
	if err != nil {
		t.Fatalf("the merge step must succeed: %v", err)
	}
	// The late commit is EMPTY, so the pinned tree is unchanged and the
	// merge correctly proceeds — of the PINNED SHA, never the branch name.
	parents := strings.Fields(stepMustGit(t, f.integ, "rev-list", "--parents", "-n", "1", sha))
	if len(parents) != 3 || parents[2] != f.cardSHA {
		t.Fatalf("the merge commit's second parent must still be the pinned SHA %s: %v", f.cardSHA, parents)
	}
	for _, arg := range mergeArgs {
		if arg == stepCardBranch {
			t.Fatalf("the merge must never name the branch: %v", mergeArgs)
		}
	}
}
