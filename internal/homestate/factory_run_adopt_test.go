package homestate

import (
	"context"
	"os"
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

// mustInsertWorker seeds one registered worker row for a run. The worker pid is
// this test process: a pid that is live on every platform (REQ-FCR-022 probes
// worker pids, and pid 1 reads dead on Windows).
func mustInsertWorker(t *testing.T, db *FactoryDB, runID, label string) {
	t.Helper()
	if _, err := db.DB.Exec(`INSERT INTO workers(label,pid,run_id,registered_at,heartbeat_at) VALUES(?,?,?,?,?)`,
		label, os.Getpid(), runID, "2026-10-14T00:00:00Z", "2026-10-14T00:00:00Z"); err != nil {
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

// Sync-audit t1707 F2 — a single candidate whose owner is only INDETERMINATE
// is NOT auto-adopted: live records prove unfinished work, not the owner's
// death, and the probe-miss path classifies a live owner indeterminate. The
// candidate is still listed, so the caller's warn-and-offer path names it.
func TestFindResumableRunWarnsSingleIndeterminateOwner(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-unk", ManifestJSON: "{}", LeadPID: 20, LeadProcessStart: "gone"})
	mustInsertWorker(t, db, "run-unk", "lane-3")

	adopt, candidates, err := db.FindResumableRun(context.Background(), ReconcileOptions{Classify: fixedClassifier(OwnerIndeterminate)})
	if err != nil {
		t.Fatalf("find resumable run: %v", err)
	}
	if adopt.RunID != "" {
		t.Fatalf("adopt = %+v, want none (an indeterminate owner is never auto-adopted)", adopt)
	}
	if len(candidates) != 1 || candidates[0].RunID != "run-unk" {
		t.Fatalf("candidates = %+v, want run-unk listed for the warn path", candidates)
	}
}

// Card-review round 2 — an unstamped legacy row classified through the
// leader-peer fallback carries no row-level identity to compare-and-set
// against, so it is listed for the warn path but never auto-adopted.
func TestFindResumableRunWarnsUnstampedPeerClassifiedRun(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-legacy", ManifestJSON: "{}"})
	mustInsertCard(t, db, "run-legacy", "c1", "run")

	adopt, candidates, err := db.FindResumableRun(context.Background(), ReconcileOptions{
		Classify: fixedClassifier(OwnerDead),
		Fallback: func(string) (int, string, bool) { return 42, "peer-start", true },
	})
	if err != nil {
		t.Fatalf("find resumable run: %v", err)
	}
	if adopt.RunID != "" {
		t.Fatalf("adopt = %+v, want none (a peer-classified unstamped row has no row identity to CAS on)", adopt)
	}
	if len(candidates) != 1 || candidates[0].RunID != "run-legacy" || candidates[0].Basis != BasisPeer {
		t.Fatalf("candidates = %+v, want run-legacy classified on the peer basis", candidates)
	}
}

// Adoption restamps the owner to the resuming process and appends one
// run.resumed event carrying the adoption basis, with the row left active.
func TestAdoptRunRestampsOwnerAndAppendsResumedEvent(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-old", ManifestJSON: "{}", LeadPID: 20, LeadProcessStart: "gone"})

	adopt := RunOwner{RunID: "run-old", LeadPID: 20, LeadProcessStart: "gone"}
	if err := db.AdoptRun(context.Background(), adopt, 4242, "new-start", "leader-resume-adoption"); err != nil {
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

// Sync-audit t1707 F1 — the restamp is a compare-and-set on the classified
// owner stamp: once one resume launch has adopted (the stamp now names ITS
// process), a second launch still carrying the ORIGINAL stamp is refused
// instead of overwriting the first adopter's live identity.
func TestAdoptRunRefusesChangedStamp(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-old", ManifestJSON: "{}", LeadPID: 20, LeadProcessStart: "gone"})

	first := RunOwner{RunID: "run-old", LeadPID: 20, LeadProcessStart: "gone"}
	if err := db.AdoptRun(context.Background(), first, 4242, "first-resume", "leader-resume-adoption"); err != nil {
		t.Fatalf("first adopt: %v", err)
	}
	second := RunOwner{RunID: "run-old", LeadPID: 20, LeadProcessStart: "gone"}
	if err := db.AdoptRun(context.Background(), second, 5151, "second-resume", "leader-resume-adoption"); err == nil {
		t.Fatalf("second adopt with the stale stamp succeeded; it must be refused")
	}
	var pid int
	var start string
	if err := db.DB.QueryRow(`SELECT lead_pid, lead_process_start FROM runs WHERE run_id='run-old'`).Scan(&pid, &start); err != nil {
		t.Fatalf("read row after refused adopt: %v", err)
	}
	if pid != 4242 || start != "first-resume" {
		t.Fatalf("stamp = (%d, %q), want the FIRST adopter's (4242, \"first-resume\")", pid, start)
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
	adopt := RunOwner{RunID: "run-gone", LeadPID: 20, LeadProcessStart: "gone"}
	if err := db.AdoptRun(context.Background(), adopt, 4242, "new-start", "leader-resume-adoption"); err == nil {
		t.Fatalf("AdoptRun on a retired run succeeded; it must refuse")
	}
	if got := runStatus(t, db, "run-gone"); got != "retired" {
		t.Fatalf("status = %q, want \"retired\"", got)
	}
}

// The live-records predicate's terminal set is spelled as a SQL literal
// (SQLite cannot bind a list); this test pins the literal to the shared
// terminalStates definition so the two cannot drift apart.
func TestLiveRecordsTerminalSetMatchesSharedDefinition(t *testing.T) {
	var want strings.Builder
	want.WriteString("(")
	for i, s := range terminalStates {
		if i > 0 {
			want.WriteString(",")
		}
		want.WriteString("'" + s + "'")
	}
	want.WriteString(")")
	if sqlTerminalStates != want.String() {
		t.Fatalf("sqlTerminalStates = %s, want the shared terminalStates set rendered as %s", sqlTerminalStates, want.String())
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

// A failed card is a terminal record (shared terminalStates set): a
// dead-owner run whose cards are all terminal — done, failed, or abandoned —
// retires exactly as before the guard existed.
func TestReconcileRetiresDeadOwnerWithTerminalRecordsOnly(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-fin", ManifestJSON: "{}", LeadPID: 20, LeadProcessStart: "gone"})
	mustInsertCard(t, db, "run-fin", "c-done", "done")
	mustInsertCard(t, db, "run-fin", "c-failed", "failed")
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
