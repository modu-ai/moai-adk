package harness

// Regression tests for card t1433: a log line that fails JSON parsing must
// survive a prune verbatim. They use only the public Retention API, so they
// compile against the pre-fix code and fail there (RED) instead of failing to build.

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// unparsedShapes are lines json.Unmarshal into Event rejects, in the ways a
// damaged log realistically produces them.
var unparsedShapes = map[string]string{
	"truncated-json":   `{"timestamp":"2026-09-01T00:00:00Z","event_type":"moai_subcommand","subj`,
	"plain-text":       `not json at all`,
	"wrong-field-type": `{"timestamp":123,"subject":"typed-wrong"}`,
	"json-array":       `[1,2,3]`,
}

func marshalEvent(t *testing.T, evt Event) string {
	t.Helper()
	raw, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}
	return string(raw)
}

func readLogLines(t *testing.T, logPath string) []string {
	t.Helper()
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	return splitNonEmptyLines(string(data))
}

// TestPruneKeepsUnparsedLinesVerbatim: a stale event is archived and removed, while
// the lines around it that do not parse stay in the log byte for byte and in file order.
func TestPruneKeepsUnparsedLinesVerbatim(t *testing.T) {
	t.Parallel()
	for name, bad := range unparsedShapes {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			logPath := filepath.Join(dir, "usage-log.jsonl")
			archiveDir := filepath.Join(dir, "archive")
			now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)

			stale := marshalEvent(t, Event{Timestamp: now.AddDate(0, 0, -40), EventType: EventTypeMoaiSubcommand, Subject: "stale", ContextHash: "h", SchemaVersion: LogSchemaVersion})
			fresh := marshalEvent(t, Event{Timestamp: now.AddDate(0, 0, -1), EventType: EventTypeFeedback, Subject: "fresh", ContextHash: "h", SchemaVersion: LogSchemaVersion})
			trailing := `{"half-written`
			content := strings.Join([]string{stale, bad, fresh, trailing}, "\n") + "\n"
			if err := os.WriteFile(logPath, []byte(content), 0o644); err != nil {
				t.Fatalf("write log: %v", err)
			}

			if err := NewRetention(logPath, archiveDir, func() time.Time { return now }).PruneStaleEntries(30); err != nil {
				t.Fatalf("prune: %v", err)
			}

			got := readLogLines(t, logPath)
			if len(got) != 3 {
				t.Fatalf("log has %d lines after prune, want 3 (unparsed, fresh, unparsed): %q", len(got), got)
			}
			if got[0] != bad {
				t.Errorf("line 0 = %q, want the unparsed line verbatim %q", got[0], bad)
			}
			var evt Event
			if err := json.Unmarshal([]byte(got[1]), &evt); err != nil || evt.Subject != "fresh" {
				t.Errorf("line 1 = %q, want the fresh event (err=%v)", got[1], err)
			}
			if got[2] != trailing {
				t.Errorf("line 2 = %q, want the unparsed line verbatim %q", got[2], trailing)
			}

			// The unparsed lines are kept in the log, never copied into the archive.
			archived := readArchive(t, filepath.Join(archiveDir, "2026-08.jsonl.gz"))
			if !strings.Contains(archived, `"subject":"stale"`) {
				t.Errorf("stale event missing from archive: %q", archived)
			}
			if strings.Contains(archived, bad) || strings.Contains(archived, trailing) {
				t.Errorf("an unparsed line leaked into the archive: %q", archived)
			}
		})
	}
}

// TestPruneNothingStaleLeavesLogUntouched: unparsed lines alone never trigger a rewrite.
func TestPruneNothingStaleLeavesLogUntouched(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)

	fresh := marshalEvent(t, Event{Timestamp: now.AddDate(0, 0, -1), EventType: EventTypeFeedback, Subject: "fresh", ContextHash: "h", SchemaVersion: LogSchemaVersion})
	content := "garbage line\n" + fresh + "\n  \n[1]\n"
	if err := os.WriteFile(logPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write log: %v", err)
	}

	if err := NewRetention(logPath, filepath.Join(dir, "archive"), func() time.Time { return now }).PruneStaleEntries(30); err != nil {
		t.Fatalf("prune: %v", err)
	}

	got, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if string(got) != content {
		t.Fatalf("log changed although nothing was stale:\n got %q\nwant %q", got, content)
	}
}

// readArchive returns the decompressed content of a (possibly multi-stream) gzip archive.
func readArchive(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read archive: %v", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("open archive: %v", err)
	}
	out, err := io.ReadAll(zr)
	if err != nil {
		t.Fatalf("decompress archive: %v", err)
	}
	return string(out)
}
