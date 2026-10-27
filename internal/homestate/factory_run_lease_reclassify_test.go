package homestate

import (
	"context"
	"testing"
	"time"
)

// factory_run_lease_reclassify_test.go — AC-FCR-023's c10 sub-case (REQ-FCR-019,
// SPEC v0.3.5, Stop-hook codex P2): the lease path abandons another run's row
// only while that run's owner is dead. A run proved dead through the lead peer
// or the boot proof has no row-level stamp for the lease transaction to compare,
// so such a source is classified AGAIN right before the transaction opens; if it
// is no longer dead the abandon is skipped — the new row is still recorded and
// the old row is left exactly as it was.
//
// The seam is the one the other lease tests use (leaseReboundBeforeTx, between
// the classification and the transaction) plus an injected classifier / boot
// clock whose answer flips after that point.

func rbLeaseReclassifyCase(t *testing.T, basis ProofBasis) (*FactoryDB, ReconcileOptions, *rbFlip) {
	t.Helper()
	db := openSandboxFactory(t)
	flip := &rbFlip{}
	opts := rbOpts()
	opts.Classify = func(pid int, _ string) OwnerClassification {
		switch pid {
		case rbPIDLive:
			return OwnerLive
		case 777: // the legacy peer's pid
			if flip.flipped {
				return OwnerLive
			}
			return OwnerDead
		}
		return OwnerDead
	}
	mustRecordRun(t, db, FactoryRun{RunID: "run-now", ManifestJSON: "{}", LeadPID: rbPIDLive, LeadProcessStart: "fp"})
	mustRecordRun(t, db, FactoryRun{RunID: "run-peer", ManifestJSON: "{}", LeadPID: 0, LeadProcessStart: ""})
	rbCard(t, db, "run-peer", "c10", CardRun) // no lease, no worker: abandonable while its owner is dead
	switch basis {
	case BasisPeer:
		opts.Fallback = func(runID string) (int, string, bool) {
			if runID == "run-peer" {
				return 777, "peer-fp", true
			}
			return 0, "", false
		}
	case BasisBoot:
		opts.BootTime = func() (time.Time, bool) {
			if flip.flipped {
				return time.Unix(0, 0), true
			}
			return time.Now().Add(time.Hour), true
		}
		opts.LegacyPeerRecordAbsent = func(string) bool { return true }
	}
	return db, opts, flip
}

func TestRecordPickedReclassifiesPeerAndBootSources(t *testing.T) {
	for _, basis := range []ProofBasis{BasisPeer, BasisBoot} {
		t.Run(string(basis)+"/control: the owner stays dead, the old row is abandoned", func(t *testing.T) {
			db, opts, _ := rbLeaseReclassifyCase(t, basis)
			if _, err := db.RecordPickedReconciled(context.Background(), "run-now", "c10", CardFields{}, "lane-1", rbLeaseNow, opts, nil); err != nil {
				t.Fatalf("record: %v", err)
			}
			if got := rbRowState(t, db, "run-peer", "c10"); got != CardAbandoned {
				t.Fatalf("run-peer c10 = %s, want abandoned", got)
			}
		})
		t.Run(string(basis)+"/the owner comes back to life before the transaction opens", func(t *testing.T) {
			db, opts, flip := rbLeaseReclassifyCase(t, basis)
			leaseReboundBeforeTx = func() { flip.flipped = true }
			t.Cleanup(func() { leaseReboundBeforeTx = nil })
			oldRow, err := db.LoadCard(context.Background(), "run-peer", "c10")
			if err != nil {
				t.Fatal(err)
			}
			got, err := db.RecordPickedReconciled(context.Background(), "run-now", "c10", CardFields{}, "lane-1", rbLeaseNow, opts, nil)
			if err != nil {
				t.Fatalf("record: %v", err)
			}
			if got.RunID != "run-now" || got.State != CardPicked {
				t.Fatalf("the new row = %+v, want a picked row in run-now (the record still goes ahead)", got)
			}
			if after, err := db.LoadCard(context.Background(), "run-peer", "c10"); err != nil || after != oldRow {
				t.Fatalf("run-peer c10 = (%+v, %v), want it untouched (%+v): the owner is alive again", after, err, oldRow)
			}
			if n := len(frEvents(t, db, "card.rebound")); n != 0 {
				t.Fatalf("card.rebound events = %d, want 0", n)
			}
		})
	}
}
