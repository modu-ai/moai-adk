package homestate

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// SPEC-FACTORY-LEASE-RENEW-001 (card t1663), milestone M1, AC-FLR-005: the card
// lease duration is read from workflow.factory_lease.duration on every lease
// write path, with a 15-minute default (REQ-FLR-005). The three clauses are the
// acquisition (assigned to leased), the transition extension (leased to a stage),
// and the renewal (RenewLease).

// writeFactoryWorkflowConfig writes body to <root>/.moai/config/sections/workflow.yaml.
// An empty body writes no file.
func writeFactoryWorkflowConfig(t *testing.T, root, body string) {
	t.Helper()
	if body == "" {
		return
	}
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir config sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}
}

func TestCardLeaseUsesConfiguredDuration(t *testing.T) {
	cases := []struct {
		name string
		body string
		want time.Duration
	}{
		{"default", "", 15 * time.Minute},
		{"configured", "workflow:\n  factory_lease:\n    duration: 20m\n", 20 * time.Minute},
	}
	open := func(t *testing.T, body string) *FactoryDB {
		t.Helper()
		root := factorySandbox(t)
		writeFactoryWorkflowConfig(t, root, body)
		db, err := OpenFactory(root)
		if err != nil {
			t.Fatalf("open factory: %v", err)
		}
		t.Cleanup(func() { _ = db.Close() })
		return db
	}
	wantExpiry := func(t *testing.T, got string, d time.Duration) {
		t.Helper()
		want := frNow.Add(d).Format(time.RFC3339Nano)
		if got != want {
			t.Fatalf("lease expiry = %q, want %q (write instant plus %s)", got, want, d)
		}
	}

	t.Run("acquisition", func(t *testing.T) {
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				db := open(t, c.body)
				repo := frNewRepo(t, true)
				frRegisterWorker(t, db, "worker-1")
				frPlace(t, db, Card{RunID: frRun, CardID: "acq", State: CardAssigned, Version: 1, OwnerLabel: "worker-1", WorktreePath: repo.Dir})
				got, err := db.Transition(context.Background(), TransitionRequest{RunID: frRun, CardID: "acq", To: CardLeased, ExpectedVersion: 1, Actor: "worker-1", Now: frNow})
				if err != nil {
					t.Fatalf("lease acquisition: %v", err)
				}
				wantExpiry(t, got.LeaseExpiresAt, c.want)
			})
		}
	})

	t.Run("transition", func(t *testing.T) {
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				db := open(t, c.body)
				repo := frNewRepo(t, true)
				frRegisterWorker(t, db, "worker-1")
				// The kickoff-audit edge (T8a, guardKickoffAudit) leases the card to its
				// record owner at card_transition.go:601. Its fixture is the one that
				// newFDAKickoff builds: a ready progress record, a decision index with no
				// open founder row, and an admitted PASS plan-audit verdict.
				specDir := filepath.Join(repo.Dir, ".moai", "specs", frSpecID)
				frWrite(t, filepath.Join(specDir, "progress.md"), fdaProgressReady)
				frWrite(t, filepath.Join(specDir, "decision-index.md"),
					"### Q1: x?\nLabel: FOUNDER\nClass: implementation-level\nDefault: keep (rule: preserve current behavior)\nAlternate: change\nOperator verdict: DEFAULT-APPLIED 2026-10-03T00:00:00Z claude+lane\n")
				frWriteVerdict(t, repo.Dir, "kick", "plan-audit.md", "PASS", repo.Commit)
				frPlace(t, db, Card{RunID: frRun, CardID: "kick", State: CardKickoff, Version: 1, OwnerLabel: "worker-1", WorktreePath: repo.Dir,
					EvidenceSHA: repo.Commit, SpecID: frSpecID, DecisionGate: DecisionGateKickoff, DecisionResume: CardRun})
				got, err := db.Transition(context.Background(), TransitionRequest{RunID: frRun, CardID: "kick", To: CardRun, ExpectedVersion: 1, Actor: "worker-1", Decider: DeciderAudit, QueueHold: QueueHoldClear, Now: frNow})
				if err != nil {
					t.Fatalf("stage transition: %v", err)
				}
				wantExpiry(t, got.LeaseExpiresAt, c.want)
			})
		}
	})

	t.Run("renewal", func(t *testing.T) {
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				db := open(t, c.body)
				repo := frNewRepo(t, true)
				frRegisterWorker(t, db, "worker-1")
				frPlace(t, db, Card{RunID: frRun, CardID: "ren", State: CardLeased, Version: 1, OwnerLabel: "worker-1", LeaseHolder: "worker-1", LeaseExpiresAt: frLeaseUntil(time.Minute), Stage: CardRun, WorktreePath: repo.Dir})
				got, err := db.RenewLease(context.Background(), frRun, "ren", "worker-1", frNow)
				if err != nil {
					t.Fatalf("renew lease: %v", err)
				}
				wantExpiry(t, got.LeaseExpiresAt, c.want)
			})
		}
	})
}
