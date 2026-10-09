package bugreport_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/bugreport"
	"github.com/modu-ai/moai-adk/internal/bugreport/captureframesprobe"
)

// TestCaptureFramesComeFromTheCallerStack pins the review-gate P1 finding:
// the frames in the spool entry are the CALLER's stack — the recover site's
// frames — not the capture worker goroutine's. Capture runs its work inside
// a time-boxed worker goroutine (the box discipline), and collecting the
// stack INSIDE that goroutine loses the caller's frames entirely: the entry
// then carries capture's own frames or none, which the frame filter
// discards — every real report would arrive at the drain frameless, and the
// payload build would reject it.
//
// The test lives in the EXTERNAL test package because its probe must call
// Capture from a frame OUTSIDE the capture package: the frame filter drops
// every internal/bugreport frame by design, so an in-package probe would be
// filtered out of its own assertion.
func TestCaptureFramesComeFromTheCallerStack(t *testing.T) {
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	configDir := filepath.Join(home, "config")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	body := "participation:\n  enabled: true\n  asked: true\n"
	if err := os.WriteFile(filepath.Join(configDir, "participation.yaml"), []byte(body), 0o600); err != nil {
		t.Fatalf("seed consent: %v", err)
	}

	captureframesprobe.Probe()

	spool, err := bugreport.SpoolPath()
	if err != nil {
		t.Fatalf("SpoolPath: %v", err)
	}
	raw, err := os.ReadFile(spool)
	if err != nil || len(raw) == 0 {
		t.Fatalf("spool not written: %v", err)
	}
	var entry bugreport.SpoolEntry
	if err := json.Unmarshal(raw[:strings.IndexByte(string(raw), '\n')], &entry); err != nil {
		t.Fatalf("decode spool line: %v (%s)", err, raw)
	}

	if len(entry.Frames) == 0 {
		t.Fatalf("spool entry carries no frames: %s", raw)
	}
	joined := strings.Join(entry.Frames, "|")
	// The probe is a NAMED moai function on the caller's goroutine: its
	// frame must be present in the entry (module prefix stripped).
	if !strings.Contains(joined, "internal/bugreport/captureframesprobe.Probe") {
		t.Fatalf("caller frame missing from %q — the stack was taken on the worker goroutine", joined)
	}
	// Capture's own worker body must never appear as a reported frame.
	if strings.Contains(joined, "captureWork") {
		t.Fatalf("capture's worker frame leaked into the entry: %q", joined)
	}
}
