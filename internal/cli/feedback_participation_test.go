package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/bugreport"
	"github.com/modu-ai/moai-adk/internal/feedback"
	"github.com/modu-ai/moai-adk/internal/feedback/outbox"
	"github.com/spf13/cobra"
)

// participationCliFixture seeds a temporary MOAI_HOME, consent on, and two
// queued reports whose render is deterministic; it returns the new command.
func participationCliFixture(t *testing.T) *cobra.Command {
	t.Helper()
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	t.Setenv("CI", "")

	dir := filepath.Join(home, filepath.FromSlash(bugreport.BugreportStoreDir))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir store: %v", err)
	}
	configDir := filepath.Join(home, "config")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "participation.yaml"),
		[]byte("participation:\n  enabled: true\n  asked: true\n"), 0o600); err != nil {
		t.Fatalf("seed consent: %v", err)
	}

	store := outbox.BugreportQueueStore()
	err := store.Mutate(func(rec *feedback.QueueRecord) error {
		rec.Items = []feedback.QueueItem{
			{ID: "f1", Title: "[auto-report] panic abcdef1234567890", Body: "<!-- moai-bugreport:v1 schema=v1 fingerprint=abcdef1234567890 kind=panic -->", QueuedAt: "2026-10-07T00:00:00Z"},
			{ID: "f2", Title: "[auto-report] internal_error 0123456789abcdef", Body: "<!-- moai-bugreport:v1 schema=v1 fingerprint=0123456789abcdef kind=internal_error -->", QueuedAt: "2026-10-07T00:00:00Z"},
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed queue: %v", err)
	}

	cmd := &cobra.Command{Use: "preview"}
	cmd.SetOut(&strings.Builder{})
	cmd.SetErr(&strings.Builder{})
	return cmd
}

// TestParticipationPreview prints, per queued item, exactly the rendered
// title and body bytes — the one render the create path reuses — and makes
// no network request (the outbox package has no network code at all).
func TestParticipationPreview(t *testing.T) {
	participationCliFixture(t)

	var out strings.Builder
	cmd := &cobra.Command{Use: "preview"}
	cmd.SetOut(&out)
	if err := runParticipationPreview(cmd, nil); err != nil {
		t.Fatalf("runParticipationPreview: %v", err)
	}

	rendered := out.String()
	for _, want := range []string{
		"[auto-report] panic abcdef1234567890",
		"<!-- moai-bugreport:v1 schema=v1 fingerprint=abcdef1234567890 kind=panic -->",
		"[auto-report] internal_error 0123456789abcdef",
		"2 queued report(s)",
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("preview output missing %q", want)
		}
	}
}

// TestOutboxLogAppendOnly0600 covers AC-014's log arm: the outbox log is
// JSONL, append-only across runs, mode 0600, and holds the exact payload
// for queued rows.
func TestOutboxLogAppendOnly0600(t *testing.T) {
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	t.Setenv("CI", "")

	store := outbox.BugreportQueueStore()
	_ = store.Mutate(func(rec *feedback.QueueRecord) error {
		rec.Items = []feedback.QueueItem{{ID: "f1", Title: "t", Body: "b"}}
		return nil
	})

	// Two append runs: the second must ADD, not truncate.
	if err := outbox.AppendOutbox(outbox.OutboxRow{Outcome: "queued", Title: "t", Body: "b"}); err != nil {
		t.Fatalf("append 1: %v", err)
	}
	if err := outbox.AppendOutbox(outbox.OutboxRow{Outcome: "sent", Title: "t", Body: "b"}); err != nil {
		t.Fatalf("append 2: %v", err)
	}

	logPath, err := outbox.StorePath(outbox.OutboxFileName)
	if err != nil {
		t.Fatalf("StorePath: %v", err)
	}
	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("stat log: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("log mode = %v, want 0600", info.Mode().Perm())
	}
	raw, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	lines := strings.Count(strings.TrimSpace(string(raw)), "\n") + 1
	if lines != 2 {
		t.Fatalf("log carries %d lines, want 2 (append-only)", lines)
	}
	if !strings.Contains(string(raw), `"outcome":"queued"`) || !strings.Contains(string(raw), `"outcome":"sent"`) {
		t.Fatalf("log lost a row:\n%s", raw)
	}
}

// TestParticipationPurge removes the queue, the spool, the ledger, and the
// outbox log; absent stores are fine; no network, no model.
func TestParticipationPurge(t *testing.T) {
	participationCliFixture(t)

	// Stores present: queue (from the fixture), log (append one), ledger,
	// spool (write one).
	if err := outbox.AppendOutbox(outbox.OutboxRow{Outcome: "queued", Title: "t", Body: "b"}); err != nil {
		t.Fatalf("append: %v", err)
	}
	spoolPath, err := bugreport.SpoolPath()
	if err != nil {
		t.Fatalf("SpoolPath: %v", err)
	}
	if err := os.WriteFile(spoolPath, []byte("{}\n"), 0o600); err != nil {
		t.Fatalf("write spool: %v", err)
	}

	cmd := &cobra.Command{Use: "purge"}
	if err := runParticipationPurge(cmd, nil); err != nil {
		t.Fatalf("runParticipationPurge: %v", err)
	}

	home := os.Getenv("MOAI_HOME")
	dir := filepath.Join(home, filepath.FromSlash(bugreport.BugreportStoreDir))
	for _, name := range []string{"queue.json", "ledger.json", "outbox.log", filepath.Base(spoolPath)} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Errorf("%s survived the purge", name)
		}
	}

	// Idempotent: a second purge is not an error.
	if err := runParticipationPurge(cmd, nil); err != nil {
		t.Fatalf("second purge: %v", err)
	}
}
