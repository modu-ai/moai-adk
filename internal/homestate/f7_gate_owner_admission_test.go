package homestate

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The push gate asks the card's repository: with a remote the decision
// requests `pushed`, without one it requests `done`, and a repository git
// cannot inspect is an error rather than a guess.
func TestPushGateTargetByRemote(t *testing.T) {
	withRemote := frNewRepo(t, true)
	got, err := PushGateTarget(context.Background(), Card{WorktreePath: withRemote.Dir})
	if err != nil || got != CardPushed {
		t.Fatalf("remote repo target = %q err=%v, want pushed", got, err)
	}
	noRemote := frNewRepo(t, false)
	got, err = PushGateTarget(context.Background(), Card{WorktreePath: noRemote.Dir})
	if err != nil || got != CardDone {
		t.Fatalf("remote-less repo target = %q err=%v, want done", got, err)
	}
	if _, err := PushGateTarget(context.Background(), Card{WorktreePath: filepath.Join(t.TempDir(), "missing")}); err == nil {
		t.Fatal("missing worktree: err = nil, want git failure")
	}
}

// StampRunOwner refuses an empty run id and an empty owner identity, and a
// valid restamp rewrites the recorded identity.
func TestStampRunOwnerValidationAndRestamp(t *testing.T) {
	db := frOpen(t)
	ctx := context.Background()
	mustRecordRun(t, db, FactoryRun{RunID: "run-a", ManifestJSON: "{}", LeadPID: 111, LeadProcessStart: "old"})
	if err := db.StampRunOwner(ctx, "", 222, "new"); err == nil || !strings.Contains(err.Error(), "run id is empty") {
		t.Fatalf("empty run id: err = %v", err)
	}
	if err := db.StampRunOwner(ctx, "run-a", 0, "new"); err == nil || !strings.Contains(err.Error(), "owner identity is empty") {
		t.Fatalf("zero pid: err = %v", err)
	}
	if err := db.StampRunOwner(ctx, "run-a", 222, "  "); err == nil {
		t.Fatal("blank process start: err = nil, want refusal")
	}
	if err := db.StampRunOwner(ctx, "run-a", 222, "new"); err != nil {
		t.Fatalf("restamp: %v", err)
	}
	var pid int
	var start string
	if err := db.DB.QueryRowContext(ctx, `SELECT lead_pid, lead_process_start FROM runs WHERE run_id=?`, "run-a").Scan(&pid, &start); err != nil || pid != 222 || start != "new" {
		t.Fatalf("restamped owner = %d/%q err=%v, want 222/new", pid, start, err)
	}
}

// WithRuntimeAdmission runs the registration under the admission lock on a
// quiet project, and refuses to run it while a migration marker stands.
func TestWithRuntimeAdmissionMarkerGate(t *testing.T) {
	root := factorySandbox(t)
	called := false
	if err := WithRuntimeAdmission(root, func() error { called = true; return nil }); err != nil || !called {
		t.Fatalf("quiet project: err=%v called=%v, want registration to run", err, called)
	}
	done, err := InstallMigrationMarkerLocked(root, "m-1")
	if err != nil {
		t.Fatalf("install marker: %v", err)
	}
	called = false
	if err := WithRuntimeAdmission(root, func() error { called = true; return nil }); err == nil || called {
		t.Fatalf("marker present: err=%v called=%v, want refusal without registration", err, called)
	}
	if err := done(true); err != nil {
		t.Fatalf("clear marker: %v", err)
	}
	called = false
	if err := WithRuntimeAdmission(root, func() error { called = true; return nil }); err != nil || !called {
		t.Fatalf("after clear: err=%v called=%v, want registration to run", err, called)
	}
}

// AcquireAdmissionLock fails when the barrier directory cannot exist.
func TestAcquireAdmissionLockParentIsFile(t *testing.T) {
	root := factorySandbox(t)
	barrier, err := MigrationBarrierPath(root)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(barrier)
	if err := os.MkdirAll(filepath.Dir(dir), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dir, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := AcquireAdmissionLock(root); err == nil || !strings.Contains(err.Error(), "mkdir") {
		t.Fatalf("lock under file parent: err = %v, want mkdir failure", err)
	}
}
