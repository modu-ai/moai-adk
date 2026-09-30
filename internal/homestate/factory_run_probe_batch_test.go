package homestate

// factory_run_probe_batch_test.go — SPEC-CODEX-LANE-SLOTS-001 AC-007
// (REQ-013): the owner-classification resolves every identity-bearing row
// through ONE batch probe invocation per listing — the counting fake asserts
// the invocation bound — instead of one probe per run row. The
// platform-appropriate variant is pinned by the build-tagged test files
// (process_fingerprint_batch_{darwin,unix,windows}_test.go).

import (
	"context"
	"fmt"
	"slices"
	"testing"
)

func TestClassifyRunsBatchesOwnerProbe(t *testing.T) {
	db := openSandboxFactory(t)
	// Four rows over THREE distinct pids — the batch sees the deduped set.
	mustRecordRun(t, db, FactoryRun{RunID: "run-b1", Backend: "codex", LeadPID: 411, LeadProcessStart: "st-411"})
	mustRecordRun(t, db, FactoryRun{RunID: "run-b2", Backend: "codex", LeadPID: 412, LeadProcessStart: "st-412"})
	mustRecordRun(t, db, FactoryRun{RunID: "run-b3", Backend: "codex", LeadPID: 413, LeadProcessStart: "st-413"})
	mustRecordRun(t, db, FactoryRun{RunID: "run-b4", Backend: "codex", LeadPID: 411, LeadProcessStart: "st-411"})

	var calls [][]int
	opts := ReconcileOptions{BatchProbe: func(pids []int) map[int]ProcessIdentity {
		calls = append(calls, slices.Clone(pids))
		result := make(map[int]ProcessIdentity, len(pids))
		for _, pid := range pids {
			// pid 413's fingerprint drifted from its stamp: that row classifies
			// dead while the others classify live — the fake feeds both arms.
			fingerprint := fmt.Sprintf("st-%d", pid)
			if pid == 413 {
				fingerprint = "st-413-drifted"
			}
			result[pid] = ProcessIdentity{Fingerprint: fingerprint, State: ProcessIdentityLive}
		}
		return result
	}}

	owners, err := db.ClassifyRuns(context.Background(), opts)
	if err != nil {
		t.Fatalf("ClassifyRuns: %v", err)
	}
	if len(owners) != 4 {
		t.Fatalf("owners = %d rows, want 4", len(owners))
	}
	if len(calls) != 1 {
		t.Fatalf("batch probe invoked %d times for 4 rows, want 1 (bounded per listing)", len(calls))
	}
	got := slices.Clone(calls[0])
	slices.Sort(got)
	if !slices.Equal(got, []int{411, 412, 413}) {
		t.Fatalf("batch received pids %v, want the deduped [411 412 413]", got)
	}
	byID := map[string]RunOwner{}
	for _, o := range owners {
		byID[o.RunID] = o
	}
	for _, runID := range []string{"run-b1", "run-b2", "run-b4"} {
		if got := byID[runID].Classification; got != OwnerLive {
			t.Errorf("%s classified %s, want live (fingerprint matches its stamp)", runID, got)
		}
		if got := byID[runID].Basis; got != BasisStamp {
			t.Errorf("%s basis = %q, want %q", runID, got, BasisStamp)
		}
	}
	if got := byID["run-b3"].Classification; got != OwnerDead {
		t.Errorf("run-b3 classified %s, want dead (fingerprint drift)", got)
	}
}

// TestClassifyRunsBatchAbsentPidStaysIndeterminate pins the defensive
// branch: an identity-bearing row the batch result somehow lacks classifies
// indeterminate on its own basis and never falls through to the boot proof,
// which is reserved for rows with no identity at all.
func TestClassifyRunsBatchAbsentPidStaysIndeterminate(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-absent1", Backend: "codex", LeadPID: 421, LeadProcessStart: "st-421"})
	opts := ReconcileOptions{BatchProbe: func(pids []int) map[int]ProcessIdentity {
		// The probe answers every pid EXCEPT 421 — the map-miss contract.
		result := make(map[int]ProcessIdentity, len(pids))
		for _, pid := range pids {
			if pid == 421 {
				continue
			}
			result[pid] = ProcessIdentity{Fingerprint: "st-421", State: ProcessIdentityLive}
		}
		return result
	}}
	owners, err := db.ClassifyRuns(context.Background(), opts)
	if err != nil {
		t.Fatalf("ClassifyRuns: %v", err)
	}
	if got := owners[0].Classification; got != OwnerIndeterminate {
		t.Fatalf("absent-pid row classified %s, want indeterminate", got)
	}
	if got := owners[0].Basis; got != BasisStamp {
		t.Fatalf("absent-pid row basis = %q, want %q", got, BasisStamp)
	}
}

// TestClassifyRunsPerPidClassifierStillHonored pins the seam rule: a caller
// that pins a per-pid classifier gets the per-pid path unchanged — the batch
// path is the default, never an override of an explicit Classify seam.
func TestClassifyRunsPerPidClassifierStillHonored(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-pid1", Backend: "codex", LeadPID: 4242, LeadProcessStart: "st-x"})
	classifyCalls := 0
	opts := ReconcileOptions{Classify: func(int, string) OwnerClassification {
		classifyCalls++
		return OwnerLive
	}}
	owners, err := db.ClassifyRuns(context.Background(), opts)
	if err != nil {
		t.Fatalf("ClassifyRuns: %v", err)
	}
	if classifyCalls != 1 {
		t.Fatalf("per-pid classifier invoked %d times, want 1 (the explicit seam)", classifyCalls)
	}
	if owners[0].Classification != OwnerLive {
		t.Fatalf("classification = %s, want the classifier's live", owners[0].Classification)
	}
}
