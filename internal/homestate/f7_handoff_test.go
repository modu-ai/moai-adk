package homestate

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// The memory handoff is a single-slot queue: saving a new one supersedes the
// pending one, the newest is what a session reads back, and only a pending
// handoff can change status.
func TestMemoryHandoffLifecycle(t *testing.T) {
	db := frOpen(t)
	ctx := context.Background()

	if _, ok, err := db.ReadPendingMemory(ctx); err != nil || ok {
		t.Fatalf("read empty memory = ok=%v err=%v, want ok=false err=nil", ok, err)
	}
	if err := db.SaveMemory(ctx, MemoryHandoff{Sprint: "e1", Spec: "S-1", Status: "complete", Body: "first"}); err != nil {
		t.Fatal(err)
	}
	if err := db.SaveMemory(ctx, MemoryHandoff{Sprint: "e1", Spec: "S-2", Status: "complete", Body: "second"}); err != nil {
		t.Fatal(err)
	}
	got, ok, err := db.ReadPendingMemory(ctx)
	if err != nil || !ok || got.Spec != "S-2" || got.Body != "second" {
		t.Fatalf("pending memory = %+v ok=%v err=%v, want the newest (S-2/second)", got, ok, err)
	}
	if err := db.SetMemoryStatus(ctx, got.ID, "persisted", "written"); err != nil {
		t.Fatalf("persist: %v", err)
	}
	if err := db.SetMemoryStatus(ctx, got.ID, "persisted", "again"); err == nil || !strings.Contains(err.Error(), "not pending") {
		t.Fatalf("persist twice: err = %v, want not-pending refusal", err)
	}
	if _, ok, _ := db.ReadPendingMemory(ctx); ok {
		t.Fatal("pending memory survives its own persistence")
	}
}

// Clear and expire sweep the pending resume and record its event. The
// schema's unique-status index makes pending a single row, so the sweep
// operates on that one row.
func TestPendingResumeBulkTransitions(t *testing.T) {
	db := frOpen(t)
	ctx := context.Background()
	if err := db.SaveResume(ctx, ResumeHandoff{SpecID: "S-1", Body: "b-1"}); err != nil {
		t.Fatal(err)
	}
	if err := db.ClearPendingResume(ctx); err != nil {
		t.Fatalf("clear: %v", err)
	}
	var pending int
	if err := db.DB.QueryRowContext(ctx, `SELECT count(*) FROM resume_handoffs WHERE status='pending'`).Scan(&pending); err != nil || pending != 0 {
		t.Fatalf("pending after clear = %d err=%v, want 0", pending, err)
	}
	var events int
	if err := db.DB.QueryRowContext(ctx, `SELECT count(*) FROM handoff_events WHERE flow='resume' AND to_status='cleared' AND detail='explicit clear'`).Scan(&events); err != nil || events != 1 {
		t.Fatalf("clear events = %d err=%v, want 1", events, err)
	}

	if err := db.SaveResume(ctx, ResumeHandoff{SpecID: "S-3", Body: "b3"}); err != nil {
		t.Fatal(err)
	}
	if err := db.ExpirePendingResume(ctx); err != nil {
		t.Fatalf("expire: %v", err)
	}
	if err := db.DB.QueryRowContext(ctx, `SELECT count(*) FROM resume_handoffs WHERE status='expired'`).Scan(&pending); err != nil || pending != 1 {
		t.Fatalf("expired after expire = %d err=%v, want 1", pending, err)
	}
	if err := db.DB.QueryRowContext(ctx, `SELECT count(*) FROM handoff_events WHERE flow='resume' AND to_status='expired' AND detail='stale TTL'`).Scan(&events); err != nil || events != 1 {
		t.Fatalf("expire events = %d err=%v, want 1", events, err)
	}
}

// ExpireResumeIfPending expires exactly one pending row and reports whether
// it did.
func TestExpireResumeIfPending(t *testing.T) {
	db := frOpen(t)
	ctx := context.Background()
	if err := db.SaveResume(ctx, ResumeHandoff{SpecID: "S-1", Body: "b"}); err != nil {
		t.Fatal(err)
	}
	row, ok, err := db.ReadPendingResume(ctx)
	if err != nil || !ok {
		t.Fatalf("read pending: ok=%v err=%v", ok, err)
	}
	did, err := db.ExpireResumeIfPending(ctx, row.ID)
	if err != nil || !did {
		t.Fatalf("expire pending = did=%v err=%v, want true nil", did, err)
	}
	did, err = db.ExpireResumeIfPending(ctx, row.ID)
	if err != nil || did {
		t.Fatalf("expire again = did=%v err=%v, want false nil", did, err)
	}
}

// A pre-SQLite pending.json imports exactly once: the marker and any existing
// pending row both refuse a second import, and the marker is what
// LegacyResumeRetired reports.
func TestImportLegacyResumeOnce(t *testing.T) {
	db := frOpen(t)
	ctx := context.Background()
	retired, err := db.LegacyResumeRetired(ctx)
	if err != nil || retired {
		t.Fatalf("retired on fresh db = %v err=%v, want false nil", retired, err)
	}
	// A pending row without the marker refuses the import.
	if _, err := db.DB.ExecContext(ctx, `INSERT INTO resume_handoffs(status,schema_version,spec_id,phase,saved_at,saved_by_session,conversation_language,directives_json,body,body_sha256) VALUES('pending',1,'S-existing','','2026-01-01T00:00:00Z','','','','raw',hex('x'))`); err != nil {
		t.Fatal(err)
	}
	if imported, err := db.ImportLegacyResume(ctx, ResumeHandoff{SpecID: "S-legacy", Body: "legacy"}); err != nil || imported {
		t.Fatalf("import with pending row = %v err=%v, want false nil", imported, err)
	}
	if _, err := db.DB.ExecContext(ctx, `UPDATE resume_handoffs SET status='consumed' WHERE status='pending'`); err != nil {
		t.Fatal(err)
	}
	if imported, err := db.ImportLegacyResume(ctx, ResumeHandoff{SpecID: "S-legacy", Body: "legacy"}); err != nil || !imported {
		t.Fatalf("first import = %v err=%v, want true nil", imported, err)
	}
	got, ok, err := db.ReadPendingResume(ctx)
	if err != nil || !ok || got.SpecID != "S-legacy" || got.Body != "legacy" {
		t.Fatalf("pending after import = %+v ok=%v err=%v", got, ok, err)
	}
	if imported, err := db.ImportLegacyResume(ctx, ResumeHandoff{SpecID: "S-again", Body: "again"}); err != nil || imported {
		t.Fatalf("second import = %v err=%v, want false nil (pending exists)", imported, err)
	}
	// With the pending row consumed, the marker alone must still refuse a
	// re-import — this is the arm that makes the import idempotent after the
	// imported row has been claimed and finished.
	if _, err := db.DB.ExecContext(ctx, `UPDATE resume_handoffs SET status='consumed' WHERE status='pending'`); err != nil {
		t.Fatal(err)
	}
	if imported, err := db.ImportLegacyResume(ctx, ResumeHandoff{SpecID: "S-third", Body: "third"}); err != nil || imported {
		t.Fatalf("third import with no pending = %v err=%v, want false nil (marker)", imported, err)
	}
	if retired, err := db.LegacyResumeRetired(ctx); err != nil || !retired {
		t.Fatalf("retired after import = %v err=%v, want true nil", retired, err)
	}
}

func frLegacyClaimedRow(t *testing.T, db *FactoryDB, pid *int, fingerprint string) int64 {
	t.Helper()
	var pidVal any
	if pid != nil {
		pidVal = int64(*pid)
	}
	res, err := db.DB.Exec(`INSERT INTO resume_handoffs(status,schema_version,spec_id,phase,saved_at,saved_by_session,conversation_language,directives_json,body,body_sha256,claim_token,claimed_at,claim_expires_at,claim_owner_pid,claim_owner_session,claim_owner_fingerprint,legacy_recovery,legacy_recovery_reason)
VALUES('claimed',1,'S-legacy','','2026-01-01T00:00:00Z','','','','legacy-body',hex('x'),'tok-1','2026-01-01T00:00:00Z',NULL,?,?,?,1,'missing or invalid claimed_at')`,
		pidVal, "", fingerprint)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := res.LastInsertId()
	return id
}

// RecoverLegacyResume is the operator escape for a legacy claimed row: the
// decision is validated, the CAS precondition must hold, a live or unknown
// owner blocks recovery unless the census says the project is quiet, and a
// dead owner (or PID reuse) allows requeue or fail.
func TestRecoverLegacyResumeDecisions(t *testing.T) {
	deadProbe := func(int) (string, ProcessIdentityState) { return "", ProcessIdentityDead }
	liveProbe := func(int) (string, ProcessIdentityState) { return "fp-recorded", ProcessIdentityLive }
	indeterminateProbe := func(int) (string, ProcessIdentityState) { return "", ProcessIdentityIndeterminate }
	quietCensus := func() error { return nil }
	noisyCensus := func() error { return fmt.Errorf("3 active") }

	t.Run("invalid decision", func(t *testing.T) {
		db := frOpen(t)
		if err := db.RecoverLegacyResume(context.Background(), 1, "tok-1", "bogus", deadProbe, nil); err == nil || !strings.Contains(err.Error(), "invalid recovery decision") {
			t.Fatalf("err = %v, want invalid decision", err)
		}
	})
	t.Run("CAS precondition", func(t *testing.T) {
		db := frOpen(t)
		id := frLegacyClaimedRow(t, db, nil, "")
		if err := db.RecoverLegacyResume(context.Background(), id, "wrong-token", "requeue", deadProbe, nil); err == nil || !strings.Contains(err.Error(), "CAS precondition") {
			t.Fatalf("err = %v, want CAS precondition failure", err)
		}
	})
	t.Run("live owner blocks", func(t *testing.T) {
		db := frOpen(t)
		pid := 4242
		id := frLegacyClaimedRow(t, db, &pid, "fp-recorded")
		if err := db.RecoverLegacyResume(context.Background(), id, "tok-1", "requeue", liveProbe, nil); err == nil || !strings.Contains(err.Error(), "owner is live") {
			t.Fatalf("err = %v, want live-owner refusal", err)
		}
	})
	t.Run("indeterminate owner blocks", func(t *testing.T) {
		db := frOpen(t)
		pid := 4242
		id := frLegacyClaimedRow(t, db, &pid, "fp-recorded")
		if err := db.RecoverLegacyResume(context.Background(), id, "tok-1", "requeue", indeterminateProbe, nil); err == nil || !strings.Contains(err.Error(), "owner is indeterminate") {
			t.Fatalf("err = %v, want indeterminate-owner refusal", err)
		}
	})
	t.Run("pid reuse with recorded fingerprint recovers", func(t *testing.T) {
		db := frOpen(t)
		pid := 4242
		// The probe reports the live process at that pid with a fingerprint
		// different from the one recorded with the claim: pid reuse.
		id := frLegacyClaimedRow(t, db, &pid, "fp-original")
		if err := db.RecoverLegacyResume(context.Background(), id, "tok-1", "fail", liveProbe, nil); err != nil {
			t.Fatalf("pid-reuse recovery: %v", err)
		}
		assertRow(t, db, id, "failed", 0)
	})
	t.Run("dead owner requeue", func(t *testing.T) {
		db := frOpen(t)
		pid := 4242
		id := frLegacyClaimedRow(t, db, &pid, "fp-recorded")
		if err := db.RecoverLegacyResume(context.Background(), id, "tok-1", "requeue", deadProbe, nil); err != nil {
			t.Fatalf("requeue: %v", err)
		}
		assertRow(t, db, id, "pending", 0)
		got, ok, err := db.ReadPendingResume(context.Background())
		if err != nil || !ok || got.ID != id {
			t.Fatalf("requeued row = %+v ok=%v err=%v", got, ok, err)
		}
	})
	t.Run("unknown owner needs census", func(t *testing.T) {
		db := frOpen(t)
		id := frLegacyClaimedRow(t, db, nil, "")
		if err := db.RecoverLegacyResume(context.Background(), id, "tok-1", "requeue", deadProbe, nil); err == nil || !strings.Contains(err.Error(), "zero-active census") {
			t.Fatalf("err = %v, want census requirement", err)
		}
		if err := db.RecoverLegacyResume(context.Background(), id, "tok-1", "requeue", deadProbe, noisyCensus); err == nil || !strings.Contains(err.Error(), "census") {
			t.Fatalf("err = %v, want census refusal", err)
		}
		if err := db.RecoverLegacyResume(context.Background(), id, "tok-1", "fail", deadProbe, quietCensus); err != nil {
			t.Fatalf("quiet-census recovery: %v", err)
		}
		assertRow(t, db, id, "failed", 0)
	})
}

func assertRow(t *testing.T, db *FactoryDB, id int64, wantStatus string, wantLegacy int) {
	t.Helper()
	var status string
	var legacy int
	if err := db.DB.QueryRow(`SELECT status, legacy_recovery FROM resume_handoffs WHERE id=?`, id).Scan(&status, &legacy); err != nil {
		t.Fatal(err)
	}
	if status != wantStatus || legacy != wantLegacy {
		t.Fatalf("row %d = %s legacy=%d, want %s legacy=%d", id, status, legacy, wantStatus, wantLegacy)
	}
}

// ReadPendingResume reports a parse failure for a corrupt saved_at rather
// than returning a zero time.
func TestReadPendingResumeCorruptTimestamp(t *testing.T) {
	db := frOpen(t)
	ctx := context.Background()
	if err := db.SaveResume(ctx, ResumeHandoff{SpecID: "S-1", Body: "b"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB.ExecContext(ctx, `UPDATE resume_handoffs SET saved_at='not-a-time' WHERE status='pending'`); err != nil {
		t.Fatal(err)
	}
	if _, ok, err := db.ReadPendingResume(ctx); err == nil || !strings.Contains(err.Error(), "parse saved_at") || !ok {
		t.Fatalf("read corrupt = ok=%v err=%v, want parse failure with row present", ok, err)
	}
}
