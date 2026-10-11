package outbox

import (
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/bugreport"
)

// TestDrainKeepsEntriesCapturedAfterItsRead pins review-gate finding
// (P2, drain.go): the drain consumed the spool with a whole-file
// ClearSpool, so entries capture appended AFTER the drain's read were
// deleted unprocessed — a lost report. Only the consumed batch may be
// removed; later lines survive for the next drain.
func TestDrainKeepsEntriesCapturedAfterItsRead(t *testing.T) {
	path := spoolFixture(t, bugreport.KindPanic)
	consentOn(t)
	validBuildForTest(t)

	// A capture lands while the drain is between its read and its clear.
	prevAfterRead := spoolAfterReadForTest
	t.Cleanup(func() { spoolAfterReadForTest = prevAfterRead })
	spoolAfterReadForTest = func() {
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			t.Errorf("append late capture: %v", err)
			return
		}
		defer func() { _ = f.Close() }()
		line := `{"kind":"hook_timeout","verdict":"ambiguous","frames":["internal/cli.Execute"]}` + "\n"
		if _, err := f.WriteString(line); err != nil {
			t.Errorf("append late capture: %v", err)
		}
	}

	if err := Drain(); err != nil {
		t.Fatalf("drain: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read spool: %v", err)
	}
	if !strings.Contains(string(raw), "hook_timeout") {
		t.Fatalf("the entry captured during the drain was deleted with the consumed batch — a lost report; spool now: %q", string(raw))
	}

	// And the next drain processes it (no lost entry).
	if err := Drain(); err != nil {
		t.Fatalf("second drain: %v", err)
	}
	rows := outboxRows(t)
	if countOutcome(rows, "ambiguous") == 0 {
		t.Fatalf("the surviving entry was never processed: %v", rows)
	}
}

// TestConsumeSpoolPrefixKeepsUnconsumedTail pins the batch primitive
// directly: consuming a prefix removes exactly that prefix and keeps the
// tail, under the spool's cross-process section.
func TestConsumeSpoolPrefixKeepsUnconsumedTail(t *testing.T) {
	path := spoolFixture(t, bugreport.KindPanic, bugreport.KindHookTimeout)
	consentOn(t)

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read spool: %v", err)
	}
	firstLineEnd := strings.IndexByte(string(raw), '\n') + 1
	consumed := raw[:firstLineEnd]

	if err := bugreport.ConsumeSpoolPrefix(consumed); err != nil {
		t.Fatalf("consume: %v", err)
	}

	rest, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read spool after consume: %v", err)
	}
	if !strings.Contains(string(rest), "hook_timeout") {
		t.Fatalf("the unconsumed tail was lost: %q", string(rest))
	}
	if strings.Contains(string(rest), `"kind":"panic"`) {
		t.Fatalf("the consumed prefix survived: %q", string(rest))
	}
}
