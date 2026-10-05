package homestate

import (
	"context"
	"os"
	"strings"
	"testing"
)

// The resume writer is the ONLY write the lane-join discovery path performs
// (SPEC-FACTORY-LANE-JOIN-SOCKET-001 REQ-004): it restores one run row to
// 'active' while stamping a SUPPLIED owner identity — the verified leader's,
// never the calling process's. recordFactoryRunStart stamps the caller, and a
// resume that reused it would name the joining lane as the run's owner, whose
// exit would then retire a live lead's run (AC-007's mutant). Every test here
// runs under the REQ-012 sandbox: a project root outside this repository's
// worktree set with HOME / MOAI_HOME scrubbed.

// resumeFixture is the verified-leader identity a resume stamps. It is chosen
// to be distinguishable from this test process's own identity, so an
// implementation that stamps the caller instead of the supplied owner fails
// the assertion rather than passing by coincidence.
const (
	resumeVerifiedPID   = 424242
	resumeVerifiedStart = "1700000000.004242"
	resumeBasis         = `{"liveness":"pid+fingerprint","label":"leader","membership":"cwd","run_id_source":"env"}`
)

// AC-006 — a run id with no row in runs is created 'active' carrying the
// verified leader's owner stamp.
func TestResumeRunCreatesAbsentRowActiveWithSuppliedOwner(t *testing.T) {
	db := openSandboxFactory(t)

	if err := db.ResumeRun(context.Background(), "run-absent", resumeVerifiedPID, resumeVerifiedStart, resumeBasis); err != nil {
		t.Fatalf("ResumeRun(absent row): %v", err)
	}

	var status string
	var pid int
	var start string
	if err := db.DB.QueryRow(`SELECT status,lead_pid,lead_process_start FROM runs WHERE run_id=?`, "run-absent").
		Scan(&status, &pid, &start); err != nil {
		t.Fatalf("read resumed row: %v", err)
	}
	if status != "active" {
		t.Errorf("resumed absent row status = %q, want active", status)
	}
	if pid != resumeVerifiedPID || start != resumeVerifiedStart {
		t.Errorf("resumed absent row owner = (%d,%q), want the verified leader (%d,%q)", pid, start, resumeVerifiedPID, resumeVerifiedStart)
	}
}

// AC-007 — a retired row is reactivated with the verified leader's owner
// stamp. The retired fixture's stamp names a dead owner (pid 111), so an
// implementation that leaves the old stamp — or stamps the calling process —
// fails here.
func TestResumeRunReactivatesRetiredRowWithLeaderOwnerStamp(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-retired", Backend: "glm", ManifestJSON: "{}", LeadPID: 111, LeadProcessStart: "111.000001"})
	if _, err := db.DB.Exec(`UPDATE runs SET status='retired' WHERE run_id='run-retired'`); err != nil {
		t.Fatalf("retire fixture row: %v", err)
	}

	if err := db.ResumeRun(context.Background(), "run-retired", resumeVerifiedPID, resumeVerifiedStart, resumeBasis); err != nil {
		t.Fatalf("ResumeRun(retired row): %v", err)
	}

	var status string
	var pid int
	var start string
	if err := db.DB.QueryRow(`SELECT status,lead_pid,lead_process_start FROM runs WHERE run_id=?`, "run-retired").
		Scan(&status, &pid, &start); err != nil {
		t.Fatalf("read resumed row: %v", err)
	}
	if status != "active" {
		t.Errorf("resumed retired row status = %q, want active", status)
	}
	if pid != resumeVerifiedPID || start != resumeVerifiedStart {
		t.Errorf("resumed retired row owner = (%d,%q), want the verified leader (%d,%q) — the old (dead) stamp or the caller's identity here is the AC-007 mutant", pid, start, resumeVerifiedPID, resumeVerifiedStart)
	}
	if pid == os.Getpid() {
		t.Errorf("resumed row owner pid = %d, which is the calling test process — recordFactoryRunStart stamped the caller (AC-007 mutant)", pid)
	}
}

// The defensive branch: discovery would not have fired on an already-active
// row, so a resume arriving on one must change nothing about the owner stamp.
func TestResumeRunLeavesActiveRowUnchanged(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-live", Backend: "glm", ManifestJSON: "{}", LeadPID: 4242, LeadProcessStart: "1700000000.000042"})

	if err := db.ResumeRun(context.Background(), "run-live", resumeVerifiedPID, resumeVerifiedStart, resumeBasis); err != nil {
		t.Fatalf("ResumeRun(active row): %v", err)
	}

	var pid int
	var start string
	if err := db.DB.QueryRow(`SELECT lead_pid,lead_process_start FROM runs WHERE run_id=?`, "run-live").
		Scan(&pid, &start); err != nil {
		t.Fatalf("read active row: %v", err)
	}
	if pid != 4242 || start != "1700000000.000042" {
		t.Errorf("active row owner changed to (%d,%q), want the original (4242,1700000000.000042)", pid, start)
	}
}

// AC-008 — the resume appends an auditable run.resumed event recording the
// verification basis, distinct from run.started.
func TestResumeRunAppendsAuditableResumedEvent(t *testing.T) {
	db := openSandboxFactory(t)

	if err := db.ResumeRun(context.Background(), "run-audited", resumeVerifiedPID, resumeVerifiedStart, resumeBasis); err != nil {
		t.Fatalf("ResumeRun: %v", err)
	}

	payloads := frEvents(t, db, "run.resumed")
	if len(payloads) != 1 {
		t.Fatalf("run.resumed events = %d, want 1", len(payloads))
	}
	if !strings.Contains(payloads[0], "liveness") || !strings.Contains(payloads[0], "membership") {
		t.Errorf("run.resumed payload %q does not record the verification basis", payloads[0])
	}
	var started int
	if err := db.DB.QueryRow(`SELECT count(*) FROM events WHERE run_id='run-audited' AND kind='run.started'`).Scan(&started); err != nil {
		t.Fatal(err)
	}
	if started != 0 {
		t.Errorf("resume emitted run.started %d time(s); the resume event kind is run.resumed (AC-008)", started)
	}
}

// The no-op branch still records why a resume looked at an already-active row
// — the outcome is in the event payload.
func TestResumeRunEventRecordsAlreadyActiveOutcome(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-live", Backend: "glm", ManifestJSON: "{}", LeadPID: 4242, LeadProcessStart: "1700000000.000042"})

	if err := db.ResumeRun(context.Background(), "run-live", resumeVerifiedPID, resumeVerifiedStart, resumeBasis); err != nil {
		t.Fatalf("ResumeRun(active row): %v", err)
	}

	payloads := frEvents(t, db, "run.resumed")
	if len(payloads) != 1 {
		t.Fatalf("run.resumed events = %d, want 1", len(payloads))
	}
	if !strings.Contains(payloads[0], "already-active") {
		t.Errorf("run.resumed payload %q does not record the already-active outcome", payloads[0])
	}
}

// A resume without a verified identity refuses and writes nothing — the write
// exists to carry a probe-measured stamp, and an unverified one is the exact
// hazard REQ-004's dedicated writer exists to prevent.
func TestResumeRunRequiresVerifiedIdentity(t *testing.T) {
	db := openSandboxFactory(t)

	cases := []struct {
		name    string
		runID   string
		pid     int
		start   string
		wantErr string
	}{
		{"empty run id", "", resumeVerifiedPID, resumeVerifiedStart, "empty"},
		{"zero pid", "run-x", 0, resumeVerifiedStart, "verified"},
		{"empty fingerprint", "run-x", resumeVerifiedPID, "", "verified"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := db.ResumeRun(context.Background(), tc.runID, tc.pid, tc.start, resumeBasis)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("ResumeRun(%s) = %v, want an error naming %q", tc.name, err, tc.wantErr)
			}
			var rows int
			if err := db.DB.QueryRow(`SELECT count(*) FROM runs`).Scan(&rows); err != nil {
				t.Fatal(err)
			}
			if rows != 0 {
				t.Errorf("refused resume left %d run row(s) behind", rows)
			}
			var events int
			if err := db.DB.QueryRow(`SELECT count(*) FROM events`).Scan(&events); err != nil {
				t.Fatal(err)
			}
			if events != 0 {
				t.Errorf("refused resume left %d event(s) behind", events)
			}
		})
	}
}
