package homestate

import (
	"context"
	"database/sql"
	"path/filepath"
	"sync"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestFactoryV1ClaimedRowsUpgradeToV2(t *testing.T) {
	path := filepath.Join(t.TempDir(), "factory.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE meta(key TEXT PRIMARY KEY,value TEXT NOT NULL);
INSERT INTO meta VALUES('schema_version','1');
CREATE TABLE resume_handoffs(id INTEGER PRIMARY KEY AUTOINCREMENT,status TEXT NOT NULL,schema_version INTEGER NOT NULL,spec_id TEXT NOT NULL DEFAULT '',phase TEXT NOT NULL DEFAULT '',saved_at TEXT NOT NULL,saved_by_session TEXT NOT NULL DEFAULT '',conversation_language TEXT NOT NULL DEFAULT '',directives_json TEXT NOT NULL DEFAULT '{}',embedded_goal_json TEXT,body TEXT NOT NULL,body_sha256 TEXT NOT NULL,claim_token TEXT NOT NULL DEFAULT '',claimed_at TEXT,consumed_at TEXT,error TEXT NOT NULL DEFAULT '');
INSERT INTO resume_handoffs(status,schema_version,saved_at,body,body_sha256,claim_token,claimed_at) VALUES
('claimed',1,'2026-01-01T00:00:00Z','valid','x','a','2026-01-01T00:00:00Z'),
('claimed',1,'2026-01-01T00:00:00Z','legacy','x','b',NULL);`)
	if closeErr := db.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		t.Fatal(err)
	}
	f, err := OpenFactoryPath(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Errorf("close factory: %v", err)
		}
	})
	var version string
	if err := f.DB.QueryRow(`SELECT value FROM meta WHERE key='schema_version'`).Scan(&version); err != nil || version != "2" {
		t.Fatalf("version=%q err=%v", version, err)
	}
	var expiry sql.NullString
	if err := f.DB.QueryRow(`SELECT claim_expires_at FROM resume_handoffs WHERE body='valid'`).Scan(&expiry); err != nil || !expiry.Valid {
		t.Fatalf("expiry=%v err=%v", expiry, err)
	}
	var legacy int
	if err := f.DB.QueryRow(`SELECT legacy_recovery FROM resume_handoffs WHERE body='legacy'`).Scan(&legacy); err != nil || legacy != 1 {
		t.Fatalf("legacy=%d err=%v", legacy, err)
	}
}

func TestResumeLatestPendingThenExpiredReclaim(t *testing.T) {
	f, err := OpenFactoryPath(filepath.Join(t.TempDir(), "factory.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Errorf("close factory: %v", err)
		}
	})
	ctx := context.Background()
	if err := f.SaveResume(ctx, ResumeHandoff{SchemaVersion: 1, SavedAt: time.Now(), Body: "expired", DirectivesJSON: "{}"}); err != nil {
		t.Fatal(err)
	}
	first, ok, err := f.ClaimResume(ctx, ResumeClaim{Token: "old", OwnerPID: 1, TTL: -time.Second})
	if err != nil || !ok {
		t.Fatalf("claim=%v %v", ok, err)
	}
	if err := f.SaveResume(ctx, ResumeHandoff{SchemaVersion: 1, SavedAt: time.Now(), Body: "pending", DirectivesJSON: "{}"}); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	got := make(chan string, 2)
	for _, token := range []string{"a", "b"} {
		wg.Add(1)
		go func(token string) {
			defer wg.Done()
			row, ok, _ := f.ClaimResume(ctx, ResumeClaim{Token: token, OwnerPID: 2, TTL: time.Minute})
			if ok {
				got <- row.Body
			}
		}(token)
	}
	wg.Wait()
	close(got)
	var bodies []string
	for body := range got {
		bodies = append(bodies, body)
	}
	if len(bodies) != 2 || bodies[0] != "pending" || first.Body != "expired" {
		t.Fatalf("bodies=%v first=%q", bodies, first.Body)
	}
}

func TestResumeFinishRejectsABAToken(t *testing.T) {
	f, err := OpenFactoryPath(filepath.Join(t.TempDir(), "factory.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Errorf("close factory: %v", err)
		}
	})
	ctx := context.Background()
	_ = f.SaveResume(ctx, ResumeHandoff{SchemaVersion: 1, SavedAt: time.Now(), Body: "payload", DirectivesJSON: "{}"})
	row, _, _ := f.ClaimResume(ctx, ResumeClaim{Token: "A", OwnerPID: 1, TTL: -time.Second})
	_, _, _ = f.ClaimResume(ctx, ResumeClaim{Token: "B", OwnerPID: 2, TTL: time.Minute})
	if err := f.FinishResume(ctx, row.ID, "A", "consumed", ""); err == nil {
		t.Fatal("old token completed reclaimed row")
	}
	var token, status string
	_ = f.DB.QueryRow(`SELECT claim_token,status FROM resume_handoffs WHERE id=?`, row.ID).Scan(&token, &status)
	if token != "B" || status != "claimed" {
		t.Fatalf("token=%q status=%q", token, status)
	}
}

func TestResumeInjectionCrashIsAtLeastOnce(t *testing.T) {
	f, err := OpenFactoryPath(filepath.Join(t.TempDir(), "factory.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Errorf("close factory: %v", err)
		}
	})
	ctx := context.Background()
	_ = f.SaveResume(ctx, ResumeHandoff{SchemaVersion: 1, SavedAt: time.Now(), Body: "same", DirectivesJSON: "{}"})
	a, _, _ := f.ClaimResume(ctx, ResumeClaim{Token: "A", OwnerPID: 1, TTL: -time.Second})
	b, ok, err := f.ClaimResume(ctx, ResumeClaim{Token: "B", OwnerPID: 2, TTL: time.Minute})
	if err != nil || !ok || a.Body != b.Body {
		t.Fatalf("reclaim=%v err=%v bodies=%q/%q", ok, err, a.Body, b.Body)
	}
}

func TestResumeLeaseRejectsInvalidAndClosedDatabaseOperations(t *testing.T) {
	f, err := OpenFactoryPath(filepath.Join(t.TempDir(), "factory.db"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, ok, err := f.ClaimResume(ctx, ResumeClaim{}); err == nil || ok {
		t.Fatal("empty token accepted")
	}
	if err := f.FinishResume(ctx, 1, "token", "pending", ""); err == nil {
		t.Fatal("invalid finish accepted")
	}
	if err := f.RecoverLegacyResume(ctx, 1, "token", "other", nil, nil); err == nil {
		t.Fatal("invalid recovery accepted")
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := f.ClaimResume(ctx, ResumeClaim{Token: "token"}); err == nil {
		t.Fatal("claim on closed db accepted")
	}
	if err := f.FinishResume(ctx, 1, "token", "failed", "x"); err == nil {
		t.Fatal("finish on closed db accepted")
	}
	if err := f.RecoverLegacyResume(ctx, 1, "token", "fail", nil, nil); err == nil {
		t.Fatal("recover on closed db accepted")
	}
}

func TestFactorySchemaRejectsUnsupportedAndMalformedV1(t *testing.T) {
	unsupported := filepath.Join(t.TempDir(), "unsupported.db")
	db, err := sql.Open("sqlite", unsupported)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE meta(key TEXT PRIMARY KEY,value TEXT NOT NULL); INSERT INTO meta VALUES('schema_version','99');`)
	_ = db.Close()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := OpenFactoryPath(unsupported); err == nil {
		t.Fatal("unsupported schema accepted")
	}
	malformed := filepath.Join(t.TempDir(), "malformed.db")
	db, err = sql.Open("sqlite", malformed)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE meta(key TEXT PRIMARY KEY,value TEXT NOT NULL); INSERT INTO meta VALUES('schema_version','1'); CREATE TABLE resume_handoffs(id INTEGER PRIMARY KEY, claim_expires_at TEXT);`)
	_ = db.Close()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := OpenFactoryPath(malformed); err == nil {
		t.Fatal("malformed v1 migration accepted")
	}
}
