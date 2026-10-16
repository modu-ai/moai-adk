package homestate

import (
	"context"
	"testing"
)

// t1772 review P1 — a resume clears the session a previous leader left on a
// non-active row, so the resumed pid is never paired with a stale session.
func TestResumeRunClearsPreviousLeaderSession(t *testing.T) {
	db := openSandboxFactory(t)
	ctx := context.Background()
	if err := db.RecordRun(ctx, FactoryRun{RunID: "run-stale", LeadSessionID: "sess-previous-leader", Backend: "claude", ManifestJSON: "{}", LeadPID: 1111, LeadProcessStart: "1111.000001"}); err != nil {
		t.Fatalf("RecordRun: %v", err)
	}
	if _, err := db.DB.Exec(`UPDATE runs SET status='retired' WHERE run_id=?`, "run-stale"); err != nil {
		t.Fatalf("retire fixture row: %v", err)
	}
	if err := db.ResumeRun(ctx, "run-stale", resumeVerifiedPID, resumeVerifiedStart, resumeBasis); err != nil {
		t.Fatalf("ResumeRun: %v", err)
	}
	var session string
	var pid int
	if err := db.DB.QueryRow(`SELECT lead_session_id,lead_pid FROM runs WHERE run_id=?`, "run-stale").Scan(&session, &pid); err != nil {
		t.Fatalf("read resumed row: %v", err)
	}
	if session != "" || pid != resumeVerifiedPID {
		t.Fatalf("resumed row session=%q pid=%d, want an empty session with the verified pid %d", session, pid, resumeVerifiedPID)
	}
}
