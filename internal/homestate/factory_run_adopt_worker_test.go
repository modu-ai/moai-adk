package homestate

import (
	"context"
	"os"
	"slices"
	"testing"
)

// AC-FCR-024 (REQ-FCR-022): a worker row counts as a live record unless its
// pid probes positively dead, and an undecided probe counts as live. The same
// predicate serves the reconciler (a, b, c, e) and the adoption candidate scan
// (d). Worker rows are probed in both orders (debt R11): a predicate that
// looks only at the first or only at the last row fails one of the two.
func TestRunHasLiveRecordsIgnoresDeadWorkers(t *testing.T) {
	db := openSandboxFactory(t)
	opts := rbOpts()
	opts.Classify = fixedClassifier(OwnerDead)
	for _, id := range []string{"run-a", "run-b", "run-c", "run-e1", "run-e2", "run-d"} {
		mustRecordRun(t, db, FactoryRun{RunID: id, ManifestJSON: "{}", LeadPID: rbPIDDead, LeadProcessStart: "gone"})
	}
	rbWorker(t, db, "run-a", "w-a", rbPIDWorkerDead)
	rbWorker(t, db, "run-b", "w-b", rbPIDWorkerLive)
	rbWorker(t, db, "run-c", "w-c", rbPIDWorkerUnknown)
	rbWorker(t, db, "run-e1", "w-e1-dead", rbPIDWorkerDead)
	rbWorker(t, db, "run-e1", "w-e1-live", rbPIDWorkerLive)
	rbWorker(t, db, "run-e2", "w-e2-live", rbPIDWorkerLive)
	rbWorker(t, db, "run-e2", "w-e2-dead", rbPIDWorkerDead)
	rbWorker(t, db, "run-d", "w-d", rbPIDWorkerDead) // no cards: dead worker only

	// (d) the adoption candidate scan applies the same predicate: runs a and d
	// hold only a dead worker, so neither is a candidate.
	_, candidates, err := db.FindResumableRun(context.Background(), opts)
	if err != nil {
		t.Fatalf("find resumable run: %v", err)
	}
	var ids []string
	for _, c := range candidates {
		ids = append(ids, c.RunID)
	}
	slices.Sort(ids)
	if want := []string{"run-b", "run-c", "run-e1", "run-e2"}; !slices.Equal(ids, want) {
		t.Fatalf("candidates = %v, want %v (dead-worker-only runs are not candidates)", ids, want)
	}

	rec, err := db.ReconcileActiveRuns(context.Background(), opts)
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	for _, tc := range []struct{ run, want string }{
		{"run-a", "retired"}, // (a) dead worker only
		{"run-d", "retired"}, // (d) dead worker only
		{"run-b", "active"},  // (b) live worker
		{"run-c", "active"},  // (c) probe cannot decide
		{"run-e1", "active"}, // (e) [dead, live]
		{"run-e2", "active"}, // (e) [live, dead]
	} {
		if got := runStatus(t, db, tc.run); got != tc.want {
			t.Errorf("%s status = %q, want %q (preserved=%d)", tc.run, got, tc.want, len(rec.Preserved))
		}
	}
}

// The production probe is the platform pid probe: this process is live, so a
// worker registered under it keeps a run alive with no injected probe.
func TestRunHasLiveRecordsDefaultProbeCountsLiveProcess(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-live-worker", ManifestJSON: "{}", LeadPID: rbPIDDead, LeadProcessStart: "gone"})
	rbWorker(t, db, "run-live-worker", "w-self", os.Getpid())
	rec, err := db.ReconcileActiveRuns(context.Background(), ReconcileOptions{Classify: fixedClassifier(OwnerDead)})
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if got := runStatus(t, db, "run-live-worker"); got != "active" || len(rec.Preserved) != 1 {
		t.Fatalf("status = %q preserved = %d, want the live-worker run preserved", got, len(rec.Preserved))
	}
}
