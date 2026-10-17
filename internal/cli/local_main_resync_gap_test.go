package cli

// local_main_resync_gap_test.go — card t1619, codex review P1 (SPEC-LOCAL-MAIN-FLOW-001).
// The re-sync re-checks window ownership and then runs the fast-forward it guards. Both
// must sit in one mutation section (the RunMergeStep shape). A rival's --force takeover
// that lands between the two would leave the previous owner moving primary's main under
// the new owner's merge. The rival here takes the window from inside the gap.

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/factorylane"
)

func TestLocalMainResyncRivalTakeoverBeforeFastForwardCannotLand(t *testing.T) {
	root := lmfResyncFixture(t, "")
	base := lmfHead(t, root)
	target := lmfOriginCommit(t, root, base, "origin-ahead", map[string]string{"ahead.txt": "ahead\n"}, false)

	var rivalErr error
	prev := localMainResyncBeforeFastForward
	localMainResyncBeforeFastForward = func(repoRoot string) {
		_, rivalErr = factory.AcquireIntegrationLock(repoRoot, factory.IntegrationLock{
			SessionID:    lmfRival,
			SessionName:  "lane-rival",
			PID:          os.Getpid(),
			PIDSource:    factory.PIDSourceSessionOwner,
			Branch:       localMainResyncBranch,
			BranchSource: factory.BranchSourceConfig,
			Worktree:     repoRoot,
		}, true)
	}
	t.Cleanup(func() { localMainResyncBeforeFastForward = prev })

	_, err := runLocalMainResync(root)
	if rivalErr == nil {
		t.Fatalf("a rival --force takeover landed between the ownership re-check and the fast-forward (re-sync err = %v): the move is not serialized with the window's mutation lock", err)
	}
	if err != nil {
		t.Fatalf("with the rival's takeover refused, the re-sync must complete: %v", err)
	}
	if head := lmfHead(t, root); head != target {
		t.Fatalf("local main must move to origin/main %s, got HEAD %s", target, head)
	}
}

// TestLocalMainResyncRivalTakeoverDuringRevalidateCannotLand pins that the ownership read,
// the rest of the re-validation, and the fast-forward share ONE unbroken mutation section
// (card t1670, sync-audit F1 of iterations 1 to 3). With its pair below, it must fail for
// every regression the audits built: the re-validation moved out of the section (mutant
// A), the re-validation and the move split into two sections (mutant B), the re-validation
// result computed before the section and only checked inside it (mutant C), only the
// ownership read hoisted out of the section (mutant D), the move run after the section
// (mutant E), the mutation lock dropped altogether (mutant F), and the ownership read run
// in a section of its own with the cached record checked and moved on in a second section
// (mutant H, sync-audit F1 of iteration 4).
//
// It is one half of a pair: TestLocalMainResyncRivalTakeoverBeforeFastForwardCannotLand
// attempts the same synchronous takeover right before the fast-forward and is the
// deterministic catch for mutant E. This test is the deterministic catch for A to D and H.
// TestLocalMainResyncRivalTakeoverSeenByOwnershipReadRefuses pins that the holder decision
// rests on the record this read returns.
//
// Three signals, two of them deterministic.
//
//  1. The read probe. From the localMainResyncReadWindow seam, which wraps the real
//     ownership read, a rival --force takeover attempted synchronously right after the read
//     must be refused as lock-busy (factory.IsIntegrationLockBusy). The probe is tied to the
//     read itself, so it follows the read wherever a regression moves it: a read outside the
//     section lets the takeover land (mutants A, C and D). Inside the section the probe waits
//     out the mutation lock's whole retry budget before it comes back busy.
//  2. A rival in a goroutine, started from the end of a successful re-validation, keeps
//     attempting the --force takeover and records primary's HEAD when an attempt lands.
//     A takeover that lands while HEAD is still the pre-fast-forward commit is the defect.
//     This signal is timing-dependent for mutant B: the mutation lock is a polled
//     flock(LOCK_NB) with a 5-50ms jittered retry, so when section 1 releases, the re-sync
//     re-takes the lock for section 2 microseconds later while the rival is usually
//     asleep between polls. It is kept because it is the real-world failure, not because
//     it is reliable.
//  3. The deterministic tripwire. Every acquisition of the mutation lock rewrites the lock
//     artifact with a fresh owner record (ftruncate + write on Unix, create-new on
//     Windows); a contender that loses the lock writes nothing. The read seam plants a
//     record no acquisition can produce (created_at 1970-01-01) into the artifact right
//     after the read probe (which lost the lock and wrote nothing), and the
//     pre-fast-forward seam reads it back. Still intact means nobody acquired the lock
//     between the ownership read and the move: one section from the read to the move.
//     Gone means the lock was taken in between, which is exactly mutant A (the section is
//     entered after the read), mutant B (section 2 is entered after section 1), and mutant
//     H (the move's section is entered after the read's own section). The tripwire is
//     planted at the read, not at the end of the re-validation, so a section boundary
//     between the read and the rest of the re-validation falls inside the observed span.
//     The positive control is the end of the re-sync: the window release acquires the lock
//     again, so the tripwire must be gone by then, or the artifact no longer records
//     acquisitions and the tripwire cannot see a split.
//
// A rival that lands after the move but before the re-sync releases the window is the
// legitimate post-merge race: the re-sync then reports that releasing the window failed
// because the rival holds it. That outcome is accepted only with HEAD already at
// origin/main when the rival landed, and the rival as the recorded holder.
func TestLocalMainResyncRivalTakeoverDuringRevalidateCannotLand(t *testing.T) {
	root := lmfResyncFixture(t, "")
	base := lmfHead(t, root)
	target := lmfOriginCommit(t, root, base, "origin-ahead", map[string]string{"ahead.txt": "ahead\n"}, false)

	// The mutation lock artifact (factory's integrationMutationLockFileName) beside the
	// window record, and an owner record no real acquisition writes: same pid, epoch time.
	lockPath := filepath.Join(root, ".moai", "state", "integration-mutation.lock")
	tripwire := []byte(fmt.Sprintf("{\n  \"pid\": %d,\n  \"created_at\": \"1970-01-01T00:00:00Z\"\n}\n", os.Getpid()))
	tripwireIntact := func() bool {
		got, err := os.ReadFile(lockPath)
		return err == nil && bytes.Equal(got, tripwire)
	}
	rivalRecord := func(repoRoot string) factory.IntegrationLock {
		return factory.IntegrationLock{
			SessionID:    lmfRival,
			SessionName:  "lane-rival",
			PID:          os.Getpid(),
			PIDSource:    factory.PIDSourceSessionOwner,
			Branch:       localMainResyncBranch,
			BranchSource: factory.BranchSourceConfig,
			Worktree:     repoRoot,
		}
	}
	// probeSection attempts one synchronous rival takeover. Inside the mutation section it
	// waits out the lock's retry budget and comes back lock-busy; outside it, it lands.
	probeSection := func(repoRoot string) error {
		_, err := factory.AcquireIntegrationLock(repoRoot, rivalRecord(repoRoot), true)
		return err
	}

	type landing struct {
		head    string
		headErr error
	}
	var (
		mu       sync.Mutex
		landings []landing
		wg       sync.WaitGroup
		stopOnce sync.Once
	)
	stop := make(chan struct{})
	halt := func() {
		stopOnce.Do(func() { close(stop) })
		wg.Wait()
	}
	t.Cleanup(halt)

	attempted, moveSeen, intactAtMove := false, false, false
	var plantErr error
	readProbed := false
	var readProbeErr error
	prevRead := localMainResyncReadWindow
	localMainResyncReadWindow = func(repoRoot string) (*factory.IntegrationLock, error) {
		lock, err := prevRead(repoRoot)
		if !readProbed {
			readProbed = true
			readProbeErr = probeSection(repoRoot)
			// Planted after the probe: inside the section the probe lost the lock and
			// wrote nothing, so from here only a later acquisition can remove it.
			plantErr = os.WriteFile(lockPath, tripwire, 0o644)
		}
		return lock, err
	}
	prevAfter := localMainResyncAfterRevalidate
	localMainResyncAfterRevalidate = func(repoRoot string) {
		attempted = true
		wg.Add(1)
		go func() {
			defer wg.Done()
			rival := rivalRecord(repoRoot)
			// Bounded: the deadline and the stop channel both end the loop, and halt joins it.
			deadline := time.Now().Add(30 * time.Second)
			for time.Now().Before(deadline) {
				if _, err := factory.AcquireIntegrationLock(repoRoot, rival, true); err == nil {
					head, headErr := factorylane.ExecGitRunner{Dir: repoRoot}.Git("rev-parse", "HEAD")
					mu.Lock()
					landings = append(landings, landing{head: strings.TrimSpace(head), headErr: headErr})
					mu.Unlock()
					return
				}
				select {
				case <-stop:
					return
				case <-time.After(time.Millisecond):
				}
			}
		}()
	}
	prevBefore := localMainResyncBeforeFastForward
	localMainResyncBeforeFastForward = func(string) {
		moveSeen = true
		intactAtMove = tripwireIntact()
	}
	t.Cleanup(func() {
		localMainResyncReadWindow = prevRead
		localMainResyncAfterRevalidate = prevAfter
		localMainResyncBeforeFastForward = prevBefore
	})

	_, err := runLocalMainResync(root)
	halt()

	if !attempted || !moveSeen {
		t.Fatalf("the re-validation seam ran = %v and the pre-fast-forward seam ran = %v (re-sync err = %v): both must run for the section to be observed", attempted, moveSeen, err)
	}
	if plantErr != nil {
		t.Fatalf("plant the tripwire in %s: %v", lockPath, plantErr)
	}
	if tripwireIntact() {
		t.Fatalf("positive control: the tripwire survived the window release, so an acquisition of the mutation lock no longer rewrites %s and this test cannot see a split section", lockPath)
	}
	if !readProbed {
		t.Fatalf("the ownership read seam localMainResyncReadWindow never ran (re-sync err = %v): the re-validation no longer reads the window through it, so the read's placement is unobserved", err)
	}
	if !factory.IsIntegrationLockBusy(readProbeErr) {
		t.Errorf("a rival --force takeover right after the ownership read was not refused as lock-busy (got %v; re-sync err = %v): the ownership read ran outside the mutation section", readProbeErr, err)
	}
	if !intactAtMove {
		t.Errorf("the mutation lock was acquired between the ownership read and the fast-forward (re-sync err = %v): the read, the rest of the re-validation, and the move do not share one unbroken mutation section", err)
	}
	mu.Lock()
	got := append([]landing(nil), landings...)
	mu.Unlock()
	for _, l := range got {
		if l.headErr != nil {
			t.Errorf("read HEAD after the rival takeover: %v", l.headErr)
			continue
		}
		if l.head == base {
			t.Errorf("a rival --force takeover landed while primary HEAD was still %s, before the fast-forward to %s (re-sync err = %v): the ownership re-check and the move are not serialized", base, target, err)
		}
	}
	if err != nil {
		lock, readErr := factory.ReadIntegrationLock(root)
		postMergeRace := len(got) > 0 && got[0].head == target && strings.Contains(err.Error(), "releasing the window failed") &&
			readErr == nil && lock.Held() && lock.SessionID == lmfRival
		if !postMergeRace {
			t.Fatalf("the re-sync must complete, or lose only its window release to a rival that landed after the move: %v", err)
		}
	}
	if head := lmfHead(t, root); head != target {
		t.Fatalf("local main must move to origin/main %s, got HEAD %s", target, head)
	}
}

// TestLocalMainResyncRivalTakeoverSeenByOwnershipReadRefuses pins that the holder decision
// rests on the record the ownership read seam returns (card t1670, sync-audit F2 of
// iteration 4). The seam hands back the real record with the holder swapped to the rival,
// as a takeover that landed just before the read would leave it. The re-sync must refuse
// with the not-holder class and leave primary's HEAD where it was. A holder decision taken
// from any other read (the direct read hoisted out of the section, with the seam kept
// inside only for the lease check: mutant D3) still sees this session as the holder and
// moves HEAD.
func TestLocalMainResyncRivalTakeoverSeenByOwnershipReadRefuses(t *testing.T) {
	root := lmfResyncFixture(t, "")
	base := lmfHead(t, root)
	lmfOriginCommit(t, root, base, "origin-ahead", map[string]string{"ahead.txt": "ahead\n"}, false)

	seamRan := false
	prevRead := localMainResyncReadWindow
	localMainResyncReadWindow = func(repoRoot string) (*factory.IntegrationLock, error) {
		lock, err := prevRead(repoRoot)
		if err != nil || lock == nil {
			return lock, err
		}
		seamRan = true
		rival := *lock
		rival.SessionID = lmfRival
		rival.SessionName = "lane-rival"
		return &rival, nil
	}
	t.Cleanup(func() { localMainResyncReadWindow = prevRead })

	_, err := runLocalMainResync(root)
	if !seamRan {
		t.Fatalf("the ownership read seam localMainResyncReadWindow never returned a record (re-sync err = %v): the holder decision is not read through it", err)
	}
	if code, ok := factory.MergeExitCode(err); !ok || code != factory.MergeExitNotHolder {
		t.Errorf("with the read seam reporting %s as the holder, the re-sync must refuse with the not-holder class %d, got %v", lmfRival, factory.MergeExitNotHolder, err)
	}
	if head := lmfHead(t, root); head != base {
		t.Fatalf("primary HEAD moved to %s although the ownership read returned a record held by %s: the holder decision does not rest on the seam's record", head, lmfRival)
	}
}
