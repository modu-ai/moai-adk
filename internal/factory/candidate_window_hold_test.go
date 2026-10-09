package factory

// candidate_window_hold_test.go — card t1478 Finding 4 (SPEC-CANDIDATE-CI-001
// REQ-CCI-012, design.md D11): the per-card red hold holds at every point the
// window is granted to a card — the direct grant, the queue promotion, and the
// refresh paths that promote. Only a RED candidate holds. Pending, green, and
// no-candidate cards are granted exactly as before, and the gate is off unless
// workflow.candidate_ci.enabled is true.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// windowHoldSeed writes one card's candidate record with the given verdict.
func windowHoldSeed(t *testing.T, root, card, verdict string) {
	t.Helper()
	if err := WriteCandidateRecord(root, CandidateRecord{
		CardID: card, PinnedSHA: "pin-" + card, CandidateSHA: "cand-" + card,
		IntegrationBranch: "develop", IntegrationTip: "tip-" + card,
		CandidateBranch: "ci/" + card, Verdict: verdict, PushedAt: "2026-10-10T08:00:00Z",
	}); err != nil {
		t.Fatal(err)
	}
}

// windowHoldTicket is one queued waiter for card, its owner and waiter alive.
func windowHoldTicket(sessionID, card string) IntegrationTicket {
	return IntegrationTicket{
		SessionID: sessionID, SessionName: sessionID, Card: card, Branch: "develop",
		OwnerPID: os.Getpid(), PIDSource: PIDSourceSessionOwner, WaiterPID: os.Getpid(),
	}
}

// windowHoldAlwaysLive reports every owner and waiter alive.
var windowHoldAlwaysLive = WindowProcProbe{
	OwnerAlive:  func(int) bool { return true },
	WaiterAlive: func(int, string) bool { return true },
}

func TestWindowHoldGate(t *testing.T) {
	t.Run("the gate is off unless workflow.candidate_ci.enabled is true", func(t *testing.T) {
		root := t.TempDir()
		windowHoldSeed(t, root, "tA", CandidateVerdictRed)
		if CandidateGrantGate(root) != nil {
			t.Error("a gate with no candidate-CI config: the path must never enable itself")
		}
	})

	t.Run("with the key on, a red card is refused and every other card is admitted", func(t *testing.T) {
		root := t.TempDir()
		writeWorkflowCandidateCI(t, root, "true")
		windowHoldSeed(t, root, "tRed", CandidateVerdictRed)
		windowHoldSeed(t, root, "tPending", CandidateVerdictPending)
		windowHoldSeed(t, root, "tGreen", CandidateVerdictGreen)
		gate := CandidateGrantGate(root)
		if gate == nil {
			t.Fatal("no gate under the enabled key")
		}
		if err := gate("tRed"); !IsCandidateHold(err) {
			t.Errorf("red card: err %v, want a candidate-hold refusal", err)
		}
		for _, card := range []string{"tPending", "tGreen", "tNoCandidate", ""} {
			if err := gate(card); err != nil {
				t.Errorf("card %q: refused (%v), want admitted — only RED holds", card, err)
			}
		}
	})

	t.Run("an unreadable candidate record refuses fail-closed", func(t *testing.T) {
		// A corrupt entry may be the card's newest verdict, so the hold cannot
		// read past it: the grant is refused until the record is repaired.
		root := t.TempDir()
		writeWorkflowCandidateCI(t, root, "true")
		windowHoldSeed(t, root, "tBroken", CandidateVerdictGreen)
		dir := filepath.Join(candidateDir(root), "tBroken")
		if err := os.WriteFile(filepath.Join(dir, "garbage.json"), []byte("{not json"), 0o600); err != nil {
			t.Fatal(err)
		}
		gate := CandidateGrantGate(root)
		if gate == nil {
			t.Fatal("no gate under the enabled key")
		}
		if err := gate("tBroken"); !IsCandidateHold(err) {
			t.Errorf("unreadable record: err %v, want a fail-closed candidate-hold refusal", err)
		}
	})
}

func TestWindowHoldDirectGrant(t *testing.T) {
	t.Run("a red card's direct acquire is refused before any window record is written", func(t *testing.T) {
		root := t.TempDir()
		writeWorkflowCandidateCI(t, root, "true")
		windowHoldSeed(t, root, "tRed", CandidateVerdictRed)
		_, err := AcquireIntegrationWindow(root, IntegrationLock{
			SessionID: "sess-red", PID: os.Getpid(), PIDSource: PIDSourceSessionOwner, Branch: "develop", Card: "tRed",
		}, false, nil)
		if !IsCandidateHold(err) {
			t.Fatalf("acquire for a red card: err %v, want a candidate-hold refusal", err)
		}
		if _, statErr := os.Stat(integrationLockPath(root)); !errors.Is(statErr, os.ErrNotExist) {
			t.Errorf("the window record exists after a refused acquire (stat: %v): the hold refuses before any write", statErr)
		}
	})

	t.Run("a green card's direct acquire is granted as before", func(t *testing.T) {
		root := t.TempDir()
		writeWorkflowCandidateCI(t, root, "true")
		windowHoldSeed(t, root, "tGreen", CandidateVerdictGreen)
		if _, err := AcquireIntegrationWindow(root, IntegrationLock{
			SessionID: "sess-green", PID: os.Getpid(), PIDSource: PIDSourceSessionOwner, Branch: "develop", Card: "tGreen",
		}, false, nil); err != nil {
			t.Fatalf("acquire for a green card: %v", err)
		}
	})
}

func TestWindowHoldQueuePromotion(t *testing.T) {
	now := time.Date(2026, 10, 10, 9, 0, 0, 0, time.UTC)

	t.Run("a red ticket is withdrawn at promotion and the next admitted ticket takes the window", func(t *testing.T) {
		root := t.TempDir()
		writeWorkflowCandidateCI(t, root, "true")
		windowHoldSeed(t, root, "tRed", CandidateVerdictRed)
		windowHoldSeed(t, root, "tGreen", CandidateVerdictGreen)
		lock := &IntegrationLock{Queue: []IntegrationTicket{
			windowHoldTicket("sess-red", "tRed"), windowHoldTicket("sess-green", "tGreen"),
		}}
		report := RefreshWindowGated(lock, openPolicy(), windowHoldAlwaysLive, now, IntegrationLeaseDefault, CandidateGrantGate(root))
		if !lock.Held() || lock.SessionID != "sess-green" || lock.Card != "tGreen" {
			t.Fatalf("holder %q on card %q after the refresh, want sess-green on tGreen: the red ticket must not take the window", lock.SessionID, lock.Card)
		}
		if len(lock.Queue) != 0 {
			t.Errorf("queue %+v after the refresh: the refused ticket leaves the queue and the admitted one is promoted", lock.Queue)
		}
		if len(report.Refused) != 1 || !strings.Contains(report.Refused[0], "tRed") {
			t.Errorf("refused %v, want the red ticket named", report.Refused)
		}
	})

	t.Run("a stale holder with only refused tickets queued leaves the window free", func(t *testing.T) {
		root := t.TempDir()
		writeWorkflowCandidateCI(t, root, "true")
		windowHoldSeed(t, root, "tRed", CandidateVerdictRed)
		probe := WindowProcProbe{
			OwnerAlive:  func(pid int) bool { return pid != 999 },
			WaiterAlive: func(int, string) bool { return true },
		}
		lock := &IntegrationLock{
			SessionID: "sess-old", PID: 999, PIDSource: PIDSourceSessionOwner, Card: "tOld", Branch: "develop",
			AcquiredAt: "2026-10-10T07:00:00Z", LeaseExpiresAt: "2026-10-10T07:30:00Z",
			Queue: []IntegrationTicket{windowHoldTicket("sess-red", "tRed")},
		}
		report := RefreshWindowGated(lock, openPolicy(), probe, now, IntegrationLeaseDefault, CandidateGrantGate(root))
		if lock.Held() {
			t.Fatalf("holder %q after the refresh, want the window free: the only queued ticket is refused by the red hold", lock.SessionID)
		}
		if len(lock.Queue) != 0 {
			t.Errorf("queue %+v after the refresh, want the refused ticket withdrawn", lock.Queue)
		}
		if len(report.Refused) != 1 || !strings.Contains(report.Refused[0], "tRed") {
			t.Errorf("refused %v, want the red ticket named", report.Refused)
		}
	})
}
