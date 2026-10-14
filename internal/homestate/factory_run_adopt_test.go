package homestate

import (
	"context"
	"strings"
	"testing"
)

// mustInsertCard seeds one card row for a run. Only the NOT NULL columns the
// schema requires are named; every other column carries its default.
func mustInsertCard(t *testing.T, db *FactoryDB, runID, cardID, state string) {
	t.Helper()
	if _, err := db.DB.Exec(`INSERT INTO cards(run_id,card_id,state,updated_at) VALUES(?,?,?,?)`,
		runID, cardID, state, "2026-10-14T00:00:00Z"); err != nil {
		t.Fatalf("insert card %s: %v", cardID, err)
	}
}

// mustInsertWorker seeds one registered worker row for a run.
func mustInsertWorker(t *testing.T, db *FactoryDB, runID, label string) {
	t.Helper()
	if _, err := db.DB.Exec(`INSERT INTO workers(label,pid,run_id,registered_at,heartbeat_at) VALUES(?,?,?,?,?)`,
		label, 1, runID, "2026-10-14T00:00:00Z", "2026-10-14T00:00:00Z"); err != nil {
		t.Fatalf("insert worker %s: %v", label, err)
	}
}

// AC-adopt-1 — the incident shape: one active run whose owner classifies dead
// and which still carries live records is the single adoption candidate.
func TestFindResumableRunAdoptsSingleDeadOwnerWithLiveRecords(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-old", ManifestJSON: "{}", LeadPID: 20, LeadProcessStart: "gone"})
	mustInsertCard(t, db, "run-old", "t1", "run")

	adopt, candidates, err := db.FindResumableRun(context.Background(), ReconcileOptions{Classify: fixedClassifier(OwnerDead)})
	if err != nil {
		t.Fatalf("find resumable run: %v", err)
	}
	if adopt.RunID != "run-old" {
		t.Fatalf("adopt = %+v, want run-old", adopt)
	}
	if len(candidates) != 1 || candidates[0].RunID != "run-old" {
		t.Fatalf("candidates = %+v, want exactly run-old", candidates)
	}
}

// A live owner is never a candidate: the run's own leader is still running.
func TestFindResumableRunSkipsLiveOwner(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-live", ManifestJSON: "{}", LeadPID: 20, LeadProcessStart: "gone"})
	mustInsertCard(t, db, "run-live", "t1", "run")

	adopt, candidates, err := db.FindResumableRun(context.Background(), ReconcileOptions{Classify: fixedClassifier(OwnerLive)})
	if err != nil {
		t.Fatalf("find resumable run: %v", err)
	}
	if adopt.RunID != "" || len(candidates) != 0 {
		t.Fatalf("adopt = %+v candidates = %+v, want none (a live owner's run is not adoptable)", adopt, candidates)
	}
}

// A dead owner with no live records is a finished run, not a resumable one.
func TestFindResumableRunSkipsRunWithoutLiveRecords(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-done", ManifestJSON: "{}", LeadPID: 20, LeadProcessStart: "gone"})

	adopt, candidates, err := db.FindResumableRun(context.Background(), ReconcileOptions{Classify: fixedClassifier(OwnerDead)})
	if err != nil {
		t.Fatalf("find resumable run: %v", err)
	}
	if adopt.RunID != "" || len(candidates) != 0 {
		t.Fatalf("adopt = %+v candidates = %+v, want none (no live records)", adopt, candidates)
	}
}

// Two live-records runs with dead owners are ambiguous — nothing is adopted
// and the caller warns with the full candidate list.
func TestFindResumableRunReportsMultipleCandidates(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-a", ManifestJSON: "{}", LeadPID: 20, LeadProcessStart: "gone"})
	mustRecordRun(t, db, FactoryRun{RunID: "run-b", ManifestJSON: "{}", LeadPID: 21, LeadProcessStart: "gone"})
	mustInsertCard(t, db, "run-a", "c1", "run")
	mustInsertCard(t, db, "run-b", "c2", "sync")

	adopt, candidates, err := db.FindResumableRun(context.Background(), ReconcileOptions{Classify: fixedClassifier(OwnerDead)})
	if err != nil {
		t.Fatalf("find resumable run: %v", err)
	}
	if adopt.RunID != "" {
		t.Fatalf("adopt = %+v, want none (ambiguous)", adopt)
	}
	if len(candidates) != 2 || candidates[0].RunID != "run-a" || candidates[1].RunID != "run-b" {
		t.Fatalf("candidates = %+v, want run-a and run-b in id order", candidates)
	}
}

// An indeterminate owner still admits adoption: the incident's resuming
// leader cannot prove the prior process dead — absence of knowledge must not
// strand the run's records (REQ-003b's residual falls on the adoptable side
// here, because the alternative is minting over a live run).
func TestFindResumableRunAdmitsIndeterminateOwner(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-unk", ManifestJSON: "{}", LeadPID: 20, LeadProcessStart: "gone"})
	mustInsertWorker(t, db, "run-unk", "lane-3")

	adopt, candidates, err := db.FindResumableRun(context.Background(), ReconcileOptions{Classify: fixedClassifier(OwnerIndeterminate)})
	if err != nil {
		t.Fatalf("find resumable run: %v", err)
	}
	if adopt.RunID != "run-unk" || len(candidates) != 1 {
		t.Fatalf("adopt = %+v candidates = %+v, want run-unk", adopt, candidates)
	}
}

// Adoption restamps the owner to the resuming process and appends one
// run.resumed event carrying the adoption basis, with the row left active.
func TestAdoptRunRestampsOwnerAndAppendsResumedEvent(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-old", ManifestJSON: "{}", LeadPID: 20, LeadProcessStart: "gone"})

	if err := db.AdoptRun(context.Background(), "run-old", 4242, "new-start", "leader-resume-adoption"); err != nil {
		t.Fatalf("adopt: %v", err)
	}
	var pid int
	var start string
	var status string
	if err := db.DB.QueryRow(`SELECT lead_pid, lead_process_start, status FROM runs WHERE run_id='run-old'`).Scan(&pid, &start, &status); err != nil {
		t.Fatalf("read adopted row: %v", err)
	}
	if pid != 4242 || start != "new-start" || status != "active" {
		t.Fatalf("adopted row = (%d, %q, %s), want (4242, \"new-start\", active)", pid, start, status)
	}
	var payload string
	if err := db.DB.QueryRow(`SELECT payload_json FROM events WHERE run_id='run-old' AND kind='run.resumed'`).Scan(&payload); err != nil {
		t.Fatalf("read run.resumed event: %v", err)
	}
	if !strings.Contains(payload, "leader-resume-adoption") {
		t.Fatalf("run.resumed payload %q does not name the adoption basis", payload)
	}
}

// Adoption refuses a run that does not measure active: re-activating a
// retired row is the discovery resume path's decision, not this one.
func TestAdoptRunRefusesNonActiveRun(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-gone", ManifestJSON: "{}", LeadPID: 20, LeadProcessStart: "gone"})
	if _, err := db.ReconcileActiveRuns(context.Background(), ReconcileOptions{Classify: fixedClassifier(OwnerDead)}); err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if err := db.AdoptRun(context.Background(), "run-gone", 4242, "new-start", "leader-resume-adoption"); err == nil {
		t.Fatalf("AdoptRun on a retired run succeeded; it must refuse")
	}
	if got := runStatus(t, db, "run-gone"); got != "retired" {
		t.Fatalf("status = %q, want \"retired\"", got)
	}
}

// AC-guard-1 — the reconciler preserves a dead-owner run that holds live
// records: it stays active, is reported in Preserved, and only the
// records-less dead run is retired.
func TestReconcilePreservesDeadOwnerRunWithLiveRecords(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-keep", ManifestJSON: "{}", LeadPID: 20, LeadProcessStart: "gone"})
	mustRecordRun(t, db, FactoryRun{RunID: "run-gone", ManifestJSON: "{}", LeadPID: 21, LeadProcessStart: "gone"})
	mustInsertCard(t, db, "run-keep", "c1", "run")

	rec, err := db.ReconcileActiveRuns(context.Background(), ReconcileOptions{Classify: fixedClassifier(OwnerDead)})
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if got := runStatus(t, db, "run-keep"); got != "active" {
		t.Fatalf("live-records run status = %q, want \"active\" (retiring it would orphan the records)", got)
	}
	if len(rec.Preserved) != 1 || rec.Preserved[0].RunID != "run-keep" {
		t.Fatalf("preserved = %+v, want exactly run-keep", rec.Preserved)
	}
	if len(rec.Retired) != 1 || rec.Retired[0].RunID != "run-gone" {
		t.Fatalf("retired = %+v, want exactly run-gone", rec.Retired)
	}
}

// A registered worker alone is a live record too — lanes were joined to the
// run even if no card ever moved.
func TestReconcilePreservesDeadOwnerWithOnlyWorkers(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-w", ManifestJSON: "{}", LeadPID: 20, LeadProcessStart: "gone"})
	mustInsertWorker(t, db, "run-w", "lane-5")

	rec, err := db.ReconcileActiveRuns(context.Background(), ReconcileOptions{Classify: fixedClassifier(OwnerDead)})
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if got := runStatus(t, db, "run-w"); got != "active" {
		t.Fatalf("workers-only run status = %q, want \"active\"", got)
	}
	if len(rec.Preserved) != 1 || rec.Preserved[0].RunID != "run-w" {
		t.Fatalf("preserved = %+v, want exactly run-w", rec.Preserved)
	}
}

// Terminal records are not live records: a dead-owner run whose cards are all
// done or abandoned retires exactly as before the guard existed.
func TestReconcileRetiresDeadOwnerWithTerminalRecordsOnly(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-fin", ManifestJSON: "{}", LeadPID: 20, LeadProcessStart: "gone"})
	mustInsertCard(t, db, "run-fin", "c-done", "done")
	mustInsertCard(t, db, "run-fin", "c-aband", "abandoned")

	rec, err := db.ReconcileActiveRuns(context.Background(), ReconcileOptions{Classify: fixedClassifier(OwnerDead)})
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if got := runStatus(t, db, "run-fin"); got != "retired" {
		t.Fatalf("terminal-records run status = %q, want \"retired\"", got)
	}
	if len(rec.Preserved) != 0 {
		t.Fatalf("preserved = %+v, want none", rec.Preserved)
	}
	if len(rec.Retired) != 1 || rec.Retired[0].RunID != "run-fin" {
		t.Fatalf("retired = %+v, want exactly run-fin", rec.Retired)
	}
}
