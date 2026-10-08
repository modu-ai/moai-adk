package outbox

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/bugreport"
)

// TestDrainUsesCaptureTimeIdentity pins the review-gate P2 finding: the
// build identity must be captured WITH the signal. The spool entry here
// was captured by v3.2.0/abcdef1234567; the binary that flushes it
// identifies as v9.9.9/fffffff9999999. The queued fingerprint must derive
// from the CAPTURE-TIME identity — one defect on two builds yields two
// separate, correctly-versioned issues, and a report flushed by a later
// binary never masquerades as that binary's build.
func TestDrainUsesCaptureTimeIdentity(t *testing.T) {
	home := freshTestHome(t)
	t.Setenv("CI", "")
	consentOn(t)

	dir := filepath.Join(home, filepath.FromSlash(bugreport.BugreportStoreDir))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir store: %v", err)
	}
	frames := []string{"internal/cli.Execute"}
	entry := map[string]any{
		"kind":    "panic",
		"verdict": "moai",
		"frames":  frames,
		// The capture-time identity, stamped by the v3.2.0 binary.
		"version": "v3.2.0",
		"commit":  "abcdef1234567",
	}
	line, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "spool.jsonl"), append(line, '\n'), 0o600); err != nil {
		t.Fatalf("write spool: %v", err)
	}

	// The flushing binary identifies as v9.9.9/fffffff9999999 — the drain
	// must not let it near this signal's fingerprint.
	prev := buildIdentityForTest
	buildIdentityForTest = func() (string, string) { return "v9.9.9", "fffffff9999999" }
	t.Cleanup(func() { buildIdentityForTest = prev })

	if err := Drain(); err != nil {
		t.Fatalf("drain: %v", err)
	}

	rec, err := BugreportQueueStore().Load()
	if err != nil {
		t.Fatalf("load queue: %v", err)
	}
	if len(rec.Items) != 1 {
		t.Fatalf("queue holds %d items, want the one captured signal", len(rec.Items))
	}
	want := captureTimeFingerprint("v3.2.0", "abcdef1234567", "panic", frames)
	if rec.Items[0].Fingerprint != want {
		t.Fatalf("fingerprint = %s, want the capture-time identity's %s (the flushing binary's identity leaked in)", rec.Items[0].Fingerprint, want)
	}
	if !strings.Contains(rec.Items[0].Title, want) {
		t.Fatalf("title %q does not carry the capture-time fingerprint", rec.Items[0].Title)
	}
}

// captureTimeFingerprint is the canonical-input recipe (design section 4)
// evaluated for an explicit build identity — the value the queued
// fingerprint must carry when the spool entry preserves its own.
func captureTimeFingerprint(version, commit, kind string, frames []string) string {
	lines := []string{
		"v1",
		version,
		commit,
		runtime.GOOS + "/" + runtime.GOARCH,
		kind,
	}
	lines = append(lines, frames...)
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return hex.EncodeToString(sum[:])[:16]
}
