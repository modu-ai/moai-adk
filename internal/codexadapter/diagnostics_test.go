package codexadapter

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// TestRecordDiscardsWritesSinkRecord — AC-REQ-3a.
func TestRecordDiscardsWritesSinkRecord(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	var stderr bytes.Buffer

	d := Discard{Event: hook.EventPostToolUse, Key: "systemMessage", ContentLength: 13, Reason: "no delivery channel on this event"}
	if err := RecordDiscards(dir, []Discard{d}, false, &stderr); err != nil {
		t.Fatalf("RecordDiscards error = %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, DiagnosticSinkRel))
	if err != nil {
		t.Fatalf("read sink: %v", err)
	}

	var got Discard
	if err := json.Unmarshal(bytes.TrimSpace(raw), &got); err != nil {
		t.Fatalf("unmarshal sink line %q: %v", raw, err)
	}
	if got.Event != d.Event || got.Key != d.Key || got.ContentLength != d.ContentLength {
		t.Errorf("sink record = %+v, want %+v", got, d)
	}
}

// TestRecordDiscardsMirrorsToStderrWhenNotBlocking asserts the operator sees
// the discard on an ordinary exit-0 path.
func TestRecordDiscardsMirrorsToStderrWhenNotBlocking(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	d := Discard{Event: hook.EventPostToolUse, Key: "systemMessage", ContentLength: 5}
	if err := RecordDiscards(t.TempDir(), []Discard{d}, false, &stderr); err != nil {
		t.Fatalf("RecordDiscards error = %v", err)
	}
	if stderr.Len() == 0 {
		t.Fatal("nothing written to stderr on the non-blocking path")
	}
	if !strings.Contains(stderr.String(), "systemMessage") {
		t.Errorf("stderr %q does not name the discarded key", stderr.String())
	}
}

// TestRecordDiscardsSilentOnStderrWhenHookBlocked — AC-REQ-3c.
//
// stderr on an exit-2 path carries the blocking reason (PreToolUse) or the
// continuation prompt (Stop), which REQ-4 requires passing through unmodified.
// A diagnostic line appended there would change what the model receives.
func TestRecordDiscardsSilentOnStderrWhenHookBlocked(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	var stderr bytes.Buffer

	d := Discard{Event: hook.EventPreToolUse, Key: "systemMessage", ContentLength: 5}
	if err := RecordDiscards(dir, []Discard{d}, true, &stderr); err != nil {
		t.Fatalf("RecordDiscards error = %v", err)
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty: the hook exited 2 and stderr carries its reason", stderr.String())
	}
	// The sink still records it — suppressing stderr must not suppress the record.
	if _, err := os.Stat(filepath.Join(dir, DiagnosticSinkRel)); err != nil {
		t.Fatalf("sink record missing on the blocking path: %v", err)
	}
}

// TestRecordDiscardsNoOpOnEmpty keeps the sink file from being created when
// there is nothing to record.
func TestRecordDiscardsNoOpOnEmpty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	var stderr bytes.Buffer
	if err := RecordDiscards(dir, nil, false, &stderr); err != nil {
		t.Fatalf("RecordDiscards error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, DiagnosticSinkRel)); !os.IsNotExist(err) {
		t.Fatalf("sink created for an empty discard set (err = %v)", err)
	}
}

// TestRecordDiscardsAppends asserts records accumulate rather than truncate.
func TestRecordDiscardsAppends(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	var stderr bytes.Buffer
	d := Discard{Event: hook.EventPostToolUse, Key: "systemMessage", ContentLength: 1}

	for range 3 {
		if err := RecordDiscards(dir, []Discard{d}, false, &stderr); err != nil {
			t.Fatalf("RecordDiscards error = %v", err)
		}
	}

	raw, err := os.ReadFile(filepath.Join(dir, DiagnosticSinkRel))
	if err != nil {
		t.Fatalf("read sink: %v", err)
	}
	if got := len(bytes.Split(bytes.TrimSpace(raw), []byte("\n"))); got != 3 {
		t.Fatalf("sink lines = %d, want 3", got)
	}
}
