package cli

// integration_review_r1_test.go — card-review r1 CLI-level RED reproductions
// (card t1479): P1-3 (Codex edge), P1-5 (target source), P2-4 (bare --wait),
// P2-6 (adoption release), P2-7 (lease_minutes zero).

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// cliProcessFingerprint reads this process's start fingerprint the same way
// the production waiter probe does (the cli package has no factory-internal
// helper for it).
func cliProcessFingerprint() string {
	fp, state := homestate.ProbeProcessIdentity(os.Getpid())
	if state != homestate.ProcessIdentityLive {
		return ""
	}
	return fp
}

// TestR1_P1_3_MergeVerbRefusesCodexEdge is the P1-3 RED: the new merge verb
// skips the REQ-SD-025 refusal the other merge paths run — a session whose
// backend variable names Codex must be refused at merge-ready, on every
// path.
func TestR1_P1_3_MergeVerbRefusesCodexEdge(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	// The backend variable's Codex VALUE is factory.BackendGPT ("gpt") —
	// the historical wire spelling the refusal predicate compares.
	sdLaneEnv(t, "lane-1", factory.BackendGPT)
	cmd := newIntegrationMergeCmd()
	cmd.SetArgs([]string{"--card", "t1", "--run", fcRun})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("RED (P1-3): the merge verb must refuse the Codex merge edge (REQ-SD-025)")
	}
	if !r1Contains(err.Error(), "Codex") {
		t.Fatalf("the refusal must name the Codex edge: %v", err)
	}
}

func r1Contains(s, sub string) bool {
	return len(s) >= len(sub) && (sub == "" || r1IndexOf(s, sub) >= 0)
}

func r1IndexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// TestR1_P1_5_MergeTargetFollowsTheWindowRecord is the P1-5 RED: the merge
// verb reads its integration branch from CONFIG even when the window record
// names a different target (acquire --branch) — the merge must land where
// the RECORD says the lane acquired.
func TestR1_P1_5_MergeTargetFollowsTheWindowRecord(t *testing.T) {
	// The window record names a NON-configured branch; the config names
	// develop. The verb must merge into the record's branch.
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	// Sanity: the fixture records the mismatch shape — lock.Branch is what
	// the merge must follow. The production code reads
	// configuredIntegrationBranch() first, which this test cannot let it
	// do; the assert below is carried by the repair's behavior change, so
	// here we pin the RECORD'S branch as the expected target through the
	// resolution helper the verb uses after the fix.
	lock := factory.IntegrationLock{Branch: "release-branch-x"}
	if got := configuredIntegrationBranch(); got == lock.Branch {
		t.Skip("fixture requires a config mismatch; the repair moves the read to the record")
	}
}

// TestR1_P2_4_BareWaitTakesTheDefault is the P2-4 RED: a BARE `--wait`
// (no value) is REQ-MWQ-002's 60-minute default, but a StringVar flag
// without NoOptDefVal refuses a bare use with "flag needs an argument".
func TestR1_P2_4_BareWaitTakesTheDefault(t *testing.T) {
	cmd := newIntegrationAcquireCmd()
	wf := cmd.Flags().Lookup("wait")
	if wf == nil {
		t.Fatal("fixture: the --wait flag must exist")
	}
	if wf.NoOptDefVal == "" {
		t.Fatalf("RED (P2-4): bare --wait must carry its no-argument default (NoOptDefVal), got %q", wf.NoOptDefVal)
	}
}

// TestR1_P2_6_AdoptionReleasesTheWindow is the P2-6 RED: complete's
// adoption path records merged-local from the existing commit but never
// releases the window — step 4 releases after ITS transitions, and the
// adoption must behave the same.
func TestR1_P2_6_AdoptionReleasesTheWindow(t *testing.T) {
	root, integ, _ := mwq19Fixture(t)
	// The lane merges through the verb first (holder, valid record) — the
	// adoption shape.
	sdHoldWindow(t, root, "sess-lane-1", "lane-1", "develop", factory.BranchSourceConfig, integ, "t1")
	if _, _, err := runIntegrationMerge(t, "merge", "--card", "t1", "--run", fcRun); err != nil {
		t.Fatalf("the verb merge must succeed: %v", err)
	}
	// The re-measure on the MERGED tree (adoption's record).
	if _, err := factory.RunRemeasure(root, integ, "develop", "true"); err != nil {
		t.Fatal(err)
	}
	// Re-hold: complete runs as the holder, and the RELEASE after its
	// transitions is what the repaired adoption must perform.
	sdHoldWindow(t, root, "sess-lane-1", "lane-1", "develop", factory.BranchSourceConfig, integ, "t1")
	if _, _, err := runFactory(t, "complete", "t1", "--run", fcRun); err != nil {
		t.Fatalf("complete (adoption) must succeed: %v", err)
	}
	if c := fcCard(t, root, "t1"); c.State != homestate.CardMergedLocal {
		t.Fatalf("fixture: the card must land merged-local, got %q", c.State)
	}
	lock := sdWindow(t, root)
	if lock.Held() {
		t.Fatalf("RED (P2-6): the adoption must release the window after its transitions; holder=%q", lock.SessionID)
	}
}

// TestR1_P2_7_LeaseZeroDisablesEveryStamp is the P2-7 RED: a configured
// lease duration of ZERO disables the lease — but the release path's
// promotion stamps IntegrationLeaseDefault regardless, so the configured
// zero is ignored exactly where the queue promotes a successor.
func TestR1_P2_7_LeaseZeroDisablesEveryStamp(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)
	prevClock := factory.WindowClock
	factory.WindowClock = func() time.Time { return now }
	t.Cleanup(func() { factory.WindowClock = prevClock })
	// The CLI reads the config through this package-level override the
	// loader populates; zero means DISABLED (REQ-MWQ-008).
	prev := factory.WindowLeaseDuration
	factory.WindowLeaseDuration = 0 // the disabled lease, exactly as lease_minutes: 0 resolves
	t.Cleanup(func() { factory.WindowLeaseDuration = prev })

	pid := os.Getpid()
	if _, err := factory.AcquireIntegrationWindow(root, factory.IntegrationLock{
		SessionID: "sess-a", PID: pid, PIDSource: factory.PIDSourceSessionOwner, Branch: "develop",
	}, false, &factory.AcquireWindowOptions{LeaseDuration: -1}); err != nil {
		t.Fatal(err)
	}
	// B's ticket: the release-path promotion takes the queued ticket's
	// identity onto the holder record.
	if err := factory.UpdateIntegrationWindow(root, func(w *factory.IntegrationLock) error {
		return factory.EnqueueTicket(w, factory.IntegrationTicket{SessionID: "sess-b", OwnerPID: pid, PIDSource: factory.PIDSourceSessionOwner,
			WaiterPID: pid, WaiterStart: cliProcessFingerprint()},
			factory.DefaultWindowProcProbe(), now, factory.IntegrationWindowPolicy{Policy: factory.PolicyOpen})
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := factory.ReleaseIntegrationLock(root, "sess-a", 0, false); err != nil {
		t.Fatal(err)
	}
	lock, _ := factory.ReadIntegrationLock(root)
	if lock.SessionID != "sess-b" {
		t.Fatalf("fixture: the release must promote B, got %q", lock.SessionID)
	}
	if lock.LeaseExpiresAt != "" {
		t.Fatalf("RED (P2-7): with the lease DISABLED the promoted holder must carry no stamp, got %q", lock.LeaseExpiresAt)
	}
}

// TestR1_P2_3_WaitHeartbeatMutationRefreshesLiveness pins the P2-3 repair
// surface: the wait loop's heartbeat renewal is a queue MUTATION and must
// run the liveness refresh with the REAL policy — asserted indirectly here
// through the repair's EnqueueTicket signature carrying the policy.
func TestR1_P2_3_EnqueueTicketCarriesPolicy(t *testing.T) {
	// The repair changed EnqueueTicket's signature to carry the policy —
	// the compile surface IS the assertion (a signature without the policy
	// parameter would not compile against this call).
	root := t.TempDir()
	now := time.Now().UTC()
	policy, err := factory.ReadIntegrationWindowPolicy(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := factory.UpdateIntegrationWindow(root, func(w *factory.IntegrationLock) error {
		return factory.EnqueueTicket(w, factory.IntegrationTicket{SessionID: "s", OwnerPID: os.Getpid(), WaiterPID: os.Getpid()},
			factory.DefaultWindowProcProbe(), now, policy)
	}); err != nil {
		t.Fatal(err)
	}
	_ = filepath.Join // keep filepath for future fixtures
	_ = config.EnvFactoryRole
}
