// review_observation_test.go — the t1595-owned subset of the t1498 review
// overlay (SPEC-DISPATCH-INTEGRITY-001), committed per decision-index Q2.
//
// Provenance: gofmt-normalized from the overlay original (the tracked
// measurement mirror `owned-tests/owned_red_tests.go.txt`); function-body
// logic is verbatim from the bodies the evidence-ledger cells executed.
// The foreign-card tests of the overlay suite (t1596 ×3, t1562 ×1, t1561
// ×1) are deliberately absent — each belongs to its own card.
//
// AC-DI-010's TestReviewFindingFoldInterleavedArchiveLoss is the RE-AUTHORED
// serialized shape (plan-audit gate-P2): fold B is a separate OS process
// with its own lock acquisition; it never shares fold A's in-process seam.
// TestReviewFindingFoldArchiveConcurrentWrite and
// TestReviewFindingFoldGuardArchiveChange are the arch-coverage instruments
// (plan-audit iter-7 debt, dispose_in=run): the non-cooperating archive
// change cases the MEMORY.md-only instrument left unexercised.
package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestReviewFindingMergedPRPredecessor(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStateHold, factory.BacklogStatePicked)
	for _, id := range []string{"t1", "t2"} {
		fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	}
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardMergedPR}, homestate.Card{CardID: "t2", State: homestate.CardPicked, HintAfter: "t1"})
	sdRegisterLane(t, root, "lane-1")
	t.Chdir(root)
	got := fbLeasedCard(t, root, "lane-1")
	t.Logf("predecessor=merged-pr; factory next leased=%q", got)
	if got != "t2" {
		t.Errorf("merged-pr predecessor still blocks t2")
	}
}

func TestReviewFindingNominatedOverwritesDependency(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStateHold, factory.BacklogStateHold, factory.BacklogStatePicked)
	for _, id := range []string{"t1", "t2", "t3"} {
		fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	}
	fbSeedFiles(t, store, "t2", "internal/template/catalog.yaml")
	fbSeedFiles(t, store, "t3", "internal/template/catalog.yaml")
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardPicked}, homestate.Card{CardID: "t2", State: homestate.CardMergedPR}, homestate.Card{CardID: "t3", State: homestate.CardPicked, HintAfter: "t1"})
	sdRegisterLane(t, root, "lane-1")
	t.Chdir(root)
	sdLaneEnv(t, "lane-1", "")
	_, _, err := runFactory(t, "next", "--card", "t3", "--run", fcRun)
	c := fcCard(t, root, "t3")
	t.Logf("err=%v t3 state=%s after=%q; original t1 state=%s", err, c.State, c.HintAfter, fcCard(t, root, "t1").State)
	// AC-DI-005 negative arm (plan-audit iter-3 D13; micro-followup D15):
	// in this fixture t1 is UNMERGED, so a nil error is itself the failure —
	// the lease was granted past a live dependency check, i.e. the guard was
	// removed. It fails unconditionally; the only lease outcome this test
	// family tolerates is asserted by the positive control below
	// (TestReviewFindingNominatedLeasesAfterPredecessorMerges).
	if err == nil {
		t.Fatalf("leased past an unmerged predecessor — dependency check did not run: state=%s after=%q", c.State, c.HintAfter)
	}
	if !strings.Contains(err.Error(), "has not reached") || !strings.Contains(err.Error(), "t1") {
		t.Errorf("refusal does not name the unmerged predecessor t1: %v", err)
	}
	if c.State != homestate.CardPicked || c.HintAfter != "t1" {
		t.Errorf("t3 moved on refusal: state=%s after=%q", c.State, c.HintAfter)
	}
}

// TestReviewFindingNominatedLeasesAfterPredecessorMerges is AC-DI-005's
// positive control (plan-audit iter-3 D13): the same nomination with the
// predecessor at merged-pr must actually lease t3 with the stored hint —
// proving the refusal in the sibling test comes from the dependency check,
// not from a dead path.
func TestReviewFindingNominatedLeasesAfterPredecessorMerges(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStateHold, factory.BacklogStateHold, factory.BacklogStatePicked)
	for _, id := range []string{"t1", "t2", "t3"} {
		fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	}
	fbSeedFiles(t, store, "t2", "internal/template/catalog.yaml")
	fbSeedFiles(t, store, "t3", "internal/template/catalog.yaml")
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardMergedPR}, homestate.Card{CardID: "t2", State: homestate.CardMergedPR}, homestate.Card{CardID: "t3", State: homestate.CardPicked, HintAfter: "t1"})
	sdRegisterLane(t, root, "lane-1")
	t.Chdir(root)
	sdLaneEnv(t, "lane-1", "")
	_, _, err := runFactory(t, "next", "--card", "t3", "--run", fcRun)
	c := fcCard(t, root, "t3")
	t.Logf("positive control: err=%v t3 state=%s after=%q", err, c.State, c.HintAfter)
	if err != nil || c.State != homestate.CardLeased || c.HintAfter != "t1" {
		t.Errorf("merged predecessor did not release the nomination: err=%v state=%s after=%q", err, c.State, c.HintAfter)
	}
}

func TestReviewFindingFoldConcurrentWrite(t *testing.T) {
	dir := t.TempDir()
	name := "MEMORY.md"
	old := []byte("original\n")
	if err := os.WriteFile(filepath.Join(dir, name), old, 0644); err != nil {
		t.Fatal(err)
	}
	prev := memoryFoldSeam
	t.Cleanup(func() { memoryFoldSeam = prev })
	memoryFoldSeam.orderProbe = func(stage string) {
		if stage == "bytes-done" {
			if err := os.WriteFile(filepath.Join(dir, name), []byte("concurrent author's new memory\n"), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	err := atomicWriteFoldFile(dir, name, []byte("fold output\n"), old, nil, nil, nil)
	got, e := os.ReadFile(filepath.Join(dir, name))
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("err=%v final bytes=%q", err, got)
	// Both clauses of AC-DI-009 (plan-audit iter-3 D12): the write is
	// REFUSED and the concurrent author's bytes are what remains on disk.
	// A mutant that detects the change but overwrites anyway before
	// returning an error fails the byte assertion.
	if err == nil {
		t.Error("concurrent update between recheck and rename lost without refusal")
	}
	if !bytes.Equal(got, []byte("concurrent author's new memory\n")) {
		t.Errorf("concurrent author's bytes were not preserved after refusal; final = %q", got)
	}
	if err != nil && !strings.Contains(err.Error(), "changed since the plan") {
		t.Errorf("refusal is not the change-detection error: %v", err)
	}
}

// TestReviewFindingFoldInterleavedArchiveLoss is AC-DI-010's RE-AUTHORED
// serialized shape (plan-audit gate-P2; the synchronous in-callback original
// pinned the defective concurrency semantics). Fold A runs in-process with
// the order seam; at its archive-write probe — inside A's write transaction —
// the test starts fold B as a SEPARATE OS PROCESS for a different card, and
// A's critical section is held open until B's whole fold has finished or the
// bounded wait expires. Under the cross-process lock (REQ-DISPATCH-008) B
// cannot finish inside A's window: it waits at the lock, and after A's
// transaction completes and releases it, B completes normally against the
// post-A store. Under today's unlocked code B finishes inside A's window,
// A's pre-B snapshot rename overwrites B's archive write, and A refuses the
// MEMORY.md write — B's completed line lands in neither index.
func TestReviewFindingFoldInterleavedArchiveLoss(t *testing.T) {
	line2 := "- [t9003 other card](project_card_t9003_other.md)"
	files := minimalFiles()
	files[fixtureArchive] = minimalArchive()
	files["project_card_t9003_other.md"] = "---\n---\nother\n"
	dir := seedFoldStore(t, minimalMemory(line9001)+line2+"\n", files)
	prev := memoryFoldSeam
	t.Cleanup(func() { memoryFoldSeam = prev })

	var (
		bCmd      *exec.Cmd
		bWaitErr  error
		bExitAt   time.Time
		bInWindow bool // B's process finished while A's transaction was open
		bDone     chan error
	)
	memoryFoldSeam.orderProbe = func(stage string) {
		if stage != "bytes-done" {
			return
		}
		// Disarm: A's second write must not start a second fold B.
		memoryFoldSeam = foldTestSeam{}
		bCmd = foldSubprocess(t, dir, "t9003")
		if err := bCmd.Start(); err != nil {
			t.Errorf("start fold B process: %v", err)
			return
		}
		bDone = make(chan error, 1)
		go func() { bDone <- bCmd.Wait() }()
		// Hold A's write transaction open until B's whole fold has finished
		// (unlocked code) or the bounded wait expires (the serialized wait
		// the cross-process lock produces — B makes no progress inside A's
		// window because it is waiting at the lock).
		select {
		case bWaitErr = <-bDone:
			bExitAt = time.Now()
			bInWindow = true
		case <-time.After(foldSubprocessWait):
			bInWindow = false
		}
	}
	first := runMemoryFold(t, "--card", "t9001", "--yes", "--dir", dir)
	aReturnAt := time.Now()
	if bCmd == nil {
		t.Fatal("fold A's seam never started fold B")
	}
	if !bInWindow {
		// The serialized arrangement: B was still waiting when A's
		// transaction closed. Join B's own process result now — it completes
		// against the post-A store.
		select {
		case bWaitErr = <-bDone:
			bExitAt = time.Now()
		case <-time.After(foldSubprocessWait):
			_ = bCmd.Process.Kill()
			bWaitErr = <-bDone
			t.Error("fold B process did not finish after fold A released the store")
		}
	}
	mem := foldRead(t, dir, "MEMORY.md")
	archive := foldRead(t, dir, fixtureArchive)
	t.Logf("fold A err=%v; fold B err=%v in-window=%v; line t9003 in MEMORY=%v archive=%v; line t9001 in MEMORY=%v archive=%v",
		first.err, bWaitErr, bInWindow, strings.Contains(mem, line2), strings.Contains(archive, line2), strings.Contains(mem, line9001), strings.Contains(archive, line9001))
	// AC-DI-010's Then, asserted on file CONTENT and on B's own process
	// result (D12 strengthened-assertion duty): B completed normally, after
	// A's transaction closed, and every completed fold's index line is
	// present exactly once — no line lost, no duplicate.
	if bWaitErr != nil {
		t.Errorf("the waiting fold did not complete normally: %v", bWaitErr)
	}
	if !bExitAt.After(aReturnAt) {
		t.Errorf("fold B's process finished at %v, before fold A's transaction closed at %v — it did not wait for the store lock", bExitAt, aReturnAt)
	}
	if got := strings.Count(mem, line2) + strings.Count(archive, line2); got != 1 {
		t.Errorf("completed fold B's line is present %d times across the indexes after both folds terminated, want exactly 1", got)
	}
	if got := strings.Count(mem, line9001) + strings.Count(archive, line9001); got != 1 {
		t.Errorf("completed fold A's line is present %d times across the indexes after both folds terminated, want exactly 1", got)
	}
}

// TestReviewFindingFoldArchiveConcurrentWrite is the arch-coverage debt's
// archive-write instrument (plan-audit iter-7 debt, dispose_in=run): the
// fold's archive append runs with no cross-file guard, so its own post-probe
// window — between its last byte comparison and its rename — is where a
// non-cooperating concurrent author's archive change must be detected and
// refused, with the author's bytes preserved on disk.
func TestReviewFindingFoldArchiveConcurrentWrite(t *testing.T) {
	files := minimalFiles()
	files[fixtureArchive] = minimalArchive()
	files["project_card_t9003_other.md"] = "---\n---\nother\n"
	dir := seedFoldStore(t, minimalMemory(line9001), files)
	prev := memoryFoldSeam
	t.Cleanup(func() { memoryFoldSeam = prev })
	probeCount := 0
	memoryFoldSeam.orderProbe = func(stage string) {
		if stage != "bytes-done" {
			return
		}
		probeCount++
		if probeCount != 1 {
			return // probe #2 is the MEMORY.md write's — this instrument acts at the archive write
		}
		// A non-cooperating concurrent author replaces the archive — the
		// file THIS write is about to rename — after its byte comparison
		// and before its rename (plain os.WriteFile: it honors no lock).
		if err := os.WriteFile(filepath.Join(dir, fixtureArchive), []byte("concurrent author's archive bytes\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	first := runMemoryFold(t, "--card", "t9001", "--yes", "--dir", dir)
	archive := foldRead(t, dir, fixtureArchive)
	mem := foldRead(t, dir, "MEMORY.md")
	t.Logf("fold A err=%v archive=%q", first.err, archive)
	if first.err == nil {
		t.Error("the archive rename published over a concurrent author's bytes without refusal")
	}
	if !bytes.Equal([]byte(archive), []byte("concurrent author's archive bytes\n")) {
		t.Errorf("concurrent author's archive bytes were not preserved; archive = %q", archive)
	}
	if first.err != nil && !strings.Contains(first.err.Error(), "changed since the plan") {
		t.Errorf("refusal is not the change-detection error: %v", first.err)
	}
	// A refused fold leaves MEMORY.md untouched — the card's line survives.
	if got := strings.Count(mem, line9001) + strings.Count(archive, line9001); got != 1 {
		t.Errorf("the folded card's line is present %d times across the indexes, want exactly 1", got)
	}
}

// TestReviewFindingFoldGuardArchiveChange is the arch-coverage debt's guard
// instrument (plan-audit iter-7 debt, dispose_in=run; the codex gate's
// data-loss mutant): during the MEMORY.md write — whose guard re-verifies
// the archive — a non-cooperating concurrent author strips the fold's
// appended line from the archive after the guard's byte comparison and
// before the rename. The comparison that FOLLOWS the probe must include the
// guard file; otherwise MEMORY.md is renamed over a change the guard no
// longer sees and the line is lost from both indexes.
func TestReviewFindingFoldGuardArchiveChange(t *testing.T) {
	files := minimalFiles()
	files[fixtureArchive] = minimalArchive()
	dir := seedFoldStore(t, minimalMemory(line9001), files)
	prev := memoryFoldSeam
	t.Cleanup(func() { memoryFoldSeam = prev })
	probeCount := 0
	memoryFoldSeam.orderProbe = func(stage string) {
		if stage != "bytes-done" {
			return
		}
		probeCount++
		if probeCount != 2 {
			return // probe #1 is the archive write's — this instrument acts at the MEMORY.md write
		}
		// A non-cooperating concurrent author removes the fold's appended
		// line from the archive — the guard file — between the guard's byte
		// comparison and the MEMORY.md rename.
		data, err := os.ReadFile(filepath.Join(dir, fixtureArchive))
		if err != nil {
			t.Fatal(err)
		}
		stripped := bytes.Replace(data, []byte(line9001+"\n"), nil, 1)
		if err := os.WriteFile(filepath.Join(dir, fixtureArchive), stripped, 0644); err != nil {
			t.Fatal(err)
		}
	}
	first := runMemoryFold(t, "--card", "t9001", "--yes", "--dir", dir)
	mem := foldRead(t, dir, "MEMORY.md")
	archive := foldRead(t, dir, fixtureArchive)
	t.Logf("fold A err=%v; card line in MEMORY=%v archive=%v", first.err, strings.Contains(mem, line9001), strings.Contains(archive, line9001))
	if first.err == nil {
		t.Error("MEMORY.md renamed over a concurrent archive change the guard no longer sees")
	}
	if first.err != nil && !strings.Contains(first.err.Error(), "changed since the plan") {
		t.Errorf("refusal is not the change-detection error: %v", first.err)
	}
	if got := strings.Count(mem, line9001) + strings.Count(archive, line9001); got != 1 {
		t.Errorf("the folded card's line is present %d times across the indexes after the concurrent archive change, want exactly 1", got)
	}
}

// The fold subprocess used by TestReviewFindingFoldInterleavedArchiveLoss.

const (
	// foldSubprocessDirEnv gates the subprocess entry point: the child of
	// TestReviewFindingFoldInterleavedArchiveLoss re-execs this test binary
	// with it set; ordinary runs leave it unset and skip.
	foldSubprocessDirEnv  = "MOAI_FOLD_SUBPROCESS_DIR"
	foldSubprocessCardEnv = "MOAI_FOLD_SUBPROCESS_CARD"
	// foldSubprocessWait bounds the wait for fold B's process: long enough
	// for the subprocess to boot and fold under -race on a loaded machine
	// (its unlocked whole fold lands well inside), short enough to keep the
	// serialized (post-fix) runs cheap.
	foldSubprocessWait = 60 * time.Second
)

// TestFoldSubprocessHelper runs one memory fold in a SEPARATE OS PROCESS —
// the child of TestReviewFindingFoldInterleavedArchiveLoss. Its fold carries
// the child's own lock acquisition. Ordinary runs (environment gate unset)
// skip.
func TestFoldSubprocessHelper(t *testing.T) {
	dir := os.Getenv(foldSubprocessDirEnv)
	if dir == "" {
		t.Skip("fold subprocess entry — only reached via TestReviewFindingFoldInterleavedArchiveLoss")
	}
	card := os.Getenv(foldSubprocessCardEnv)
	if card == "" {
		t.Fatal("fold subprocess entry without a card")
	}
	runFoldOK(t, "--card", card, "--yes", "--dir", dir)
}

// foldSubprocess builds the re-exec'd test-binary command for one fold in a
// separate process. The parent reads the child's exit status as the fold's
// result.
func foldSubprocess(t *testing.T, dir, card string) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run", "^TestFoldSubprocessHelper$", "-test.timeout", "10m")
	cmd.Env = append(os.Environ(), foldSubprocessDirEnv+"="+dir, foldSubprocessCardEnv+"="+card)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	t.Cleanup(func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	})
	return cmd
}
