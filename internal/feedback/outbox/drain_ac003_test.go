package outbox

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/bugreport"
	"github.com/modu-ai/moai-adk/internal/feedback"
)

// consentOff writes a user-scoped consent file that says false.
func consentOff(t *testing.T) {
	t.Helper()
	home := os.Getenv("MOAI_HOME")
	configDir := filepath.Join(home, "config")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	body := "participation:\n  enabled: false\n  asked: true\n"
	if err := os.WriteFile(filepath.Join(configDir, "participation.yaml"), []byte(body), 0o600); err != nil {
		t.Fatalf("write consent: %v", err)
	}
}

// seedQueuedItem writes one item into the user-scoped bugreport queue.
func seedQueuedItem(t *testing.T) feedback.QueueItem {
	t.Helper()
	item := feedback.QueueItem{
		ID:          "f999",
		Title:       "[auto-report] panic 0123456789abcdef",
		Body:        "<!-- moai-bugreport:v1 schema=v1 fingerprint=0123456789abcdef kind=panic version=v3.2.0 commit=abcdef1234567 os_arch=linux/amd64 frames=internal/cli.Execute -->\n",
		QueuedAt:    time.Now().UTC().Format(time.RFC3339),
		Fingerprint: "0123456789abcdef",
		Kind:        "panic",
	}
	if err := BugreportQueueStore().Mutate(func(rec *feedback.QueueRecord) error {
		rec.Items = append(rec.Items, item)
		return nil
	}); err != nil {
		t.Fatalf("seed queue item: %v", err)
	}
	return item
}

func queueItems(t *testing.T) []feedback.QueueItem {
	t.Helper()
	rec, err := BugreportQueueStore().Load()
	if err != nil {
		t.Fatalf("load queue: %v", err)
	}
	return rec.Items
}

// TestDrainNoopWhenParticipationOff pins AC-003's drain arm: with
// participation off (absent consent, or a user file saying false), the
// drain makes NO capture-to-publication progression — nothing is queued,
// no queued/sent row appears. (The withdrawal discard itself is AC-022's
// subject; this criterion's "off means nothing progresses" is the
// publication-noop, and the sender-side zero-gh arm lives in the publish
// package's TestSenderNoopWhenParticipationOff.) Mutant killed: a drain
// that runs the pipeline while off.
func TestDrainNoopWhenParticipationOff(t *testing.T) {
	// (a) consent absent (no user file at all): a seeded queued item gains
	// no queued/sent row — the withdrawal discard removes it, nothing
	// progresses toward publication.
	spoolFixture(t, bugreport.KindPanic)
	seedQueuedItem(t)
	rowsBefore := len(outboxRows(t))
	if err := Drain(); err != nil {
		t.Fatalf("drain: %v", err)
	}
	newRows := outboxRows(t)[rowsBefore:]
	for _, r := range newRows {
		if r["outcome"] == "queued" || r["outcome"] == "sent" {
			t.Fatalf("participation off produced a %v row — the drain progressed toward publication", r["outcome"])
		}
	}
	if items := queueItems(t); len(items) != 0 {
		t.Fatalf("participation off left %d queued item(s) after the withdrawal discard ran", len(items))
	}

	// (b) consent false with everything already empty: the drain writes
	// NOTHING at all.
	consentOff(t)
	rowsBefore = len(outboxRows(t))
	if err := Drain(); err != nil {
		t.Fatalf("drain: %v", err)
	}
	if got := len(outboxRows(t)) - rowsBefore; got != 0 {
		t.Fatalf("an off-consent drain on empty stores wrote %d row(s), want zero", got)
	}
}

// TestDrainIgnoresForgedProjectSpool pins AC-003's arm (d): a cloned
// repository ships a well-formed spool at the PROJECT tree's bugreport
// path whose entries carry verdict: moai. The drain consults ONLY the
// user-scoped spool — the project-path store is never read, so a
// repository cannot drive publication from a consenting user's account
// (D36). Mutant killed: a drain resolving its spool from the project tier.
func TestDrainIgnoresForgedProjectSpool(t *testing.T) {
	// Establish the temporary MOAI_HOME BEFORE any consent work: the
	// consent helpers resolve it from the environment, and an unset value
	// would fall back to a path RELATIVE to the package directory — writing
	// test state into the source tree (the stray outbox/config/ this run
	// already had to remove once).
	spoolFixture(t)
	consentOn(t)

	// The user-scoped spool is empty and stays empty: the drain has
	// nothing of the user's to consume.
	home := os.Getenv("MOAI_HOME")
	userSpool := filepath.Join(home, filepath.FromSlash(bugreport.BugreportStoreDir), "spool.jsonl")
	if err := os.Remove(userSpool); err != nil && !os.IsNotExist(err) {
		t.Fatalf("clear user spool: %v", err)
	}

	// The hostile project tree, as cwd: a forged spool with a
	// verdict: moai entry.
	project := t.TempDir()
	forgedDir := filepath.Join(project, ".moai", "state", "bugreport")
	if err := os.MkdirAll(forgedDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	forged := `{"kind":"panic","verdict":"moai","frames":["internal/cli.Execute"]}` + "\n"
	if err := os.WriteFile(filepath.Join(forgedDir, "spool.jsonl"), []byte(forged), 0o644); err != nil {
		t.Fatalf("write forged spool: %v", err)
	}
	t.Chdir(project)

	rowsBefore := len(outboxRows(t))
	if err := Drain(); err != nil {
		t.Fatalf("drain: %v", err)
	}

	newRows := outboxRows(t)[rowsBefore:]
	for _, r := range newRows {
		if r["outcome"] == "queued" || r["outcome"] == "sent" {
			t.Fatalf("the forged project spool drove publication (%v row) — the drain read a project-path store", r["outcome"])
		}
	}
	if items := queueItems(t); len(items) != 0 {
		t.Fatalf("the forged project spool produced %d queue entries", len(items))
	}
	// And the forged file is untouched: the drain never consumed it.
	if _, err := os.Stat(filepath.Join(forgedDir, "spool.jsonl")); err != nil {
		t.Fatalf("the forged project spool was modified: %v", err)
	}
}
