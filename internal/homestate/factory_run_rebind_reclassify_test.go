package homestate

import (
	"context"
	"errors"
	"testing"
	"time"
)

// factory_run_rebind_reclassify_test.go — AC-FCR-021 case (f) (REQ-FCR-020 step
// 1, SPEC v0.3.4, sync-audit finding F-2): a consolidation source that was
// classified dead through the lead peer or the boot proof is classified AGAIN
// right before the transaction opens, and refused if its owner is no longer
// dead. The in-transaction stamp check cannot see this window: such a run has no
// row-level stamp ((0,'')), and the peer identity lives in the broker database.
//
// The classification seam is the same one the other rebind tests use — an
// injected classifier / boot clock whose answer flips between two calls.

// rbFlip drives a classification that changes after the caller has classified.
type rbFlip struct{ flipped bool }

// rbReclassifyCase builds a source run with no row-level stamp and the options
// under which it first classifies dead.
func rbReclassifyCase(t *testing.T, basis ProofBasis) (*FactoryDB, ReconcileOptions, *rbFlip) {
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
	mustRecordRun(t, db, FactoryRun{RunID: "run-new", ManifestJSON: "{}", LeadPID: rbPIDLive, LeadProcessStart: "fp"})
	mustRecordRun(t, db, FactoryRun{RunID: "run-legacy", ManifestJSON: "{}", LeadPID: 0, LeadProcessStart: ""})
	rbCard(t, db, "run-legacy", "c1", CardRun)
	rbBind(t, db, "c1", "run-legacy")
	switch basis {
	case BasisPeer:
		opts.Fallback = func(runID string) (int, string, bool) {
			if runID == "run-legacy" {
				return 777, "peer-fp", true
			}
			return 0, "", false
		}
	case BasisBoot:
		// The host "booted" an hour from now, so every timestamp the run
		// recorded predates the boot: dead by the boot proof. After the flip
		// the boot lies far in the past and the proof no longer holds.
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

func TestRebindReclassifiesPeerAndBootSources(t *testing.T) {
	for _, basis := range []ProofBasis{BasisPeer, BasisBoot} {
		t.Run(string(basis)+"/control: the owner stays dead, the consolidation proceeds", func(t *testing.T) {
			db, opts, _ := rbReclassifyCase(t, basis)
			src := rbOwner(t, db, opts, "run-legacy")
			if src.Classification != OwnerDead || src.Basis != basis {
				t.Fatalf("fixture: source classified %s on %s, want dead on %s", src.Classification, src.Basis, basis)
			}
			res, err := db.RebindRun(context.Background(), RebindRequest{Source: src, Target: rbOwner(t, db, opts, "run-new")}, opts)
			if err != nil {
				t.Fatalf("control: %v", err)
			}
			if len(res.Moved) != 1 || runStatus(t, db, "run-legacy") != "retired" {
				t.Fatalf("control: res=%+v status=%s, want c1 moved and the run retired", res, runStatus(t, db, "run-legacy"))
			}
		})
		t.Run(string(basis)+"/the owner comes back to life between the classification and the call", func(t *testing.T) {
			db, opts, flip := rbReclassifyCase(t, basis)
			req := RebindRequest{Source: rbOwner(t, db, opts, "run-legacy"), Target: rbOwner(t, db, opts, "run-new")}
			if req.Source.Classification != OwnerDead || req.Source.Basis != basis {
				t.Fatalf("fixture: source classified %s on %s, want dead on %s", req.Source.Classification, req.Source.Basis, basis)
			}
			before := rbDump(t, db)
			flip.flipped = true
			res, err := db.RebindRun(context.Background(), req, opts)
			if err == nil {
				t.Fatalf("the consolidation proceeded over an owner that is no longer dead: res=%+v run-legacy status=%s", res, runStatus(t, db, "run-legacy"))
			}
			if !errors.Is(err, ErrRebindRefused) || !errors.Is(err, ErrRunOwnerNotDead) {
				t.Fatalf("err = %v, want ErrRebindRefused wrapping ErrRunOwnerNotDead", err)
			}
			rbRefusedWith(t, err, "run-legacy")
			if after := rbDump(t, db); after != before {
				t.Fatalf("dump changed after the refusal:\n--- before\n%s\n--- after\n%s", before, after)
			}
		})
	}
}
