package cli

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestResumeLegacyIndeterminateOperatorRecovery(t *testing.T) {
	f, err := homestate.OpenFactoryPath(filepath.Join(t.TempDir(), "factory.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Errorf("close factory: %v", err)
		}
	})
	ctx := context.Background()
	if err := f.SaveResume(ctx, homestate.ResumeHandoff{SchemaVersion: 1, SavedAt: time.Now(), Body: "legacy", DirectivesJSON: "{}"}); err != nil {
		t.Fatal(err)
	}
	row, _, err := f.ClaimResume(ctx, homestate.ResumeClaim{Token: "old", OwnerPID: 99, TTL: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.DB.Exec(`UPDATE resume_handoffs SET legacy_recovery=1,legacy_recovery_reason='unknown',claim_expires_at=NULL WHERE id=?`, row.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.RecoverLegacyResume(ctx, row.ID, "old", "requeue", func(int) (string, homestate.ProcessIdentityState) { return "", homestate.ProcessIdentityIndeterminate }, nil); err == nil {
		t.Fatal("indeterminate owner was requeued")
	}
	if err := f.RecoverLegacyResume(ctx, row.ID, "old", "requeue", func(int) (string, homestate.ProcessIdentityState) { return "", homestate.ProcessIdentityDead }, nil); err != nil {
		t.Fatal(err)
	}
	var status string
	var expiry sql.NullString
	if err := f.DB.QueryRow(`SELECT status,claim_expires_at FROM resume_handoffs WHERE id=?`, row.ID).Scan(&status, &expiry); err != nil {
		t.Fatal(err)
	}
	if status != "pending" || expiry.Valid {
		t.Fatalf("status=%q expiry=%v", status, expiry)
	}
	claimed, ok, err := f.ClaimResume(ctx, homestate.ResumeClaim{Token: "new", OwnerPID: 100, TTL: time.Minute})
	if err != nil || !ok || claimed.Body != "legacy" {
		t.Fatalf("claim=%v err=%v", ok, err)
	}
	if err := f.FinishResume(ctx, row.ID, "old", "consumed", ""); err == nil {
		t.Fatal("old token accepted")
	}
}

func TestResumeLegacyUnknownOwnerRequiresZeroActiveCensus(t *testing.T) {
	f, err := homestate.OpenFactoryPath(filepath.Join(t.TempDir(), "factory.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Errorf("close factory: %v", err)
		}
	})
	ctx := context.Background()
	if err := f.SaveResume(ctx, homestate.ResumeHandoff{SchemaVersion: 1, SavedAt: time.Now(), Body: "unknown", DirectivesJSON: "{}"}); err != nil {
		t.Fatal(err)
	}
	row, _, err := f.ClaimResume(ctx, homestate.ResumeClaim{Token: "unknown-token", OwnerPID: -1, TTL: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.DB.Exec(`UPDATE resume_handoffs SET legacy_recovery=1,claim_owner_pid=NULL,claim_owner_fingerprint='' WHERE id=?`, row.ID); err != nil {
		t.Fatal(err)
	}
	probe := func(int) (string, homestate.ProcessIdentityState) { return "", homestate.ProcessIdentityIndeterminate }
	if err := f.RecoverLegacyResume(ctx, row.ID, "unknown-token", "requeue", probe, func() error { return fmt.Errorf("active runtime") }); err == nil {
		t.Fatal("unknown owner with active census requeued")
	}
	var status string
	if err := f.DB.QueryRow(`SELECT status FROM resume_handoffs WHERE id=?`, row.ID).Scan(&status); err != nil || status != "claimed" {
		t.Fatalf("status=%s err=%v", status, err)
	}
	if err := f.RecoverLegacyResume(ctx, row.ID, "unknown-token", "requeue", probe, func() error { return nil }); err != nil {
		t.Fatal(err)
	}
}

func TestFactoryRecoverResumeCommandValidatesAndFailsDeadLegacyClaim(t *testing.T) {
	root := t.TempDir()
	t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), "home"))
	t.Chdir(root)
	invalid := newFactoryCommand()
	invalid.SetArgs([]string{"handoff", "recover-resume"})
	if err := invalid.Execute(); err == nil {
		t.Fatal("missing recovery identity accepted")
	}
	f, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := f.SaveResume(ctx, homestate.ResumeHandoff{SchemaVersion: 1, SavedAt: time.Now(), Body: "legacy", DirectivesJSON: "{}"}); err != nil {
		t.Fatal(err)
	}
	row, _, err := f.ClaimResume(ctx, homestate.ResumeClaim{Token: "old", OwnerPID: -1, TTL: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.DB.Exec(`UPDATE resume_handoffs SET legacy_recovery=1,legacy_recovery_reason='unknown',claim_expires_at=NULL WHERE id=?`, row.ID); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	cmd := newFactoryCommand()
	cmd.SetArgs([]string{"handoff", "recover-resume", "--id", "1", "--expected-token", "old", "--decision", "fail"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestFactoryRecoverResumeCommandRequeuesUnknownOwnerAtZeroCensus(t *testing.T) {
	root := t.TempDir()
	t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), "home"))
	t.Chdir(root)
	f, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := f.SaveResume(ctx, homestate.ResumeHandoff{SchemaVersion: 1, SavedAt: time.Now(), Body: "legacy", DirectivesJSON: "{}"}); err != nil {
		t.Fatal(err)
	}
	row, _, err := f.ClaimResume(ctx, homestate.ResumeClaim{Token: "old", OwnerPID: -1, TTL: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.DB.Exec(`UPDATE resume_handoffs SET legacy_recovery=1,claim_owner_pid=NULL,claim_owner_fingerprint='' WHERE id=?`, row.ID); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	cmd := newFactoryCommand()
	cmd.SetArgs([]string{"handoff", "recover-resume", "--id", "1", "--expected-token", "old", "--decision", "requeue"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	check, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := check.Close(); err != nil {
			t.Errorf("close factory: %v", err)
		}
	})
	var status string
	if err := check.DB.QueryRow(`SELECT status FROM resume_handoffs WHERE id=?`, row.ID).Scan(&status); err != nil || status != "pending" {
		t.Fatalf("status=%q err=%v", status, err)
	}
	if err := check.SaveResume(ctx, homestate.ResumeHandoff{SchemaVersion: 1, SavedAt: time.Now(), Body: "blocked", DirectivesJSON: "{}"}); err != nil {
		t.Fatal(err)
	}
	blocked, _, err := check.ClaimResume(ctx, homestate.ResumeClaim{Token: "blocked-token", OwnerPID: -1, TTL: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := check.DB.Exec(`UPDATE resume_handoffs SET legacy_recovery=1,claim_owner_pid=NULL,claim_owner_fingerprint='' WHERE id=?`, blocked.ID); err != nil {
		t.Fatal(err)
	}
	if err := check.Close(); err != nil {
		t.Fatal(err)
	}
	registry := filepath.Join(root, ".moai", "state", "active-sessions.json")
	if err := os.MkdirAll(filepath.Dir(registry), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(registry, []byte(fmt.Sprintf(`[{"pid":%d}]`, os.Getpid())), 0o600); err != nil {
		t.Fatal(err)
	}
	active := newFactoryCommand()
	active.SetArgs([]string{"handoff", "recover-resume", "--id", fmt.Sprint(blocked.ID), "--expected-token", "blocked-token", "--decision", "requeue"})
	if err := active.Execute(); err == nil || !strings.Contains(err.Error(), "active runtimes") {
		t.Fatalf("active census recovery err=%v", err)
	}
}
