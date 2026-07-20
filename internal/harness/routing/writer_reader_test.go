package routing

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestWriterAppendRoundTrip verifies a written row reads back identically.
func TestWriterAppendRoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, LedgerFileName)
	w := NewWriter(path)

	row := Row{
		SchemaVersion:     SchemaVersion,
		TS:                time.Now().UTC().Format(time.RFC3339),
		SessionID:         "s1",
		ModelClass:        "opus",
		RequestDigest:     "sha256:0123456789ab",
		RequestClass:      "feature",
		MatchedSubcommand: "run",
		Outcome:           OutcomeSuccess,
		Delegations:       []Delegation{},
		EvidenceRefs:      []EvidenceRef{},
	}
	if err := w.Append(row); err != nil {
		t.Fatalf("append: %v", err)
	}
	rows, skipped, err := NewReader(path).Read(Filter{})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if skipped != 0 || len(rows) != 1 {
		t.Fatalf("got %d rows, %d skipped; want 1,0", len(rows), skipped)
	}
	if rows[0].SessionID != "s1" || rows[0].Outcome != OutcomeSuccess {
		t.Fatalf("round-trip mismatch: %+v", rows[0])
	}
}

// TestConcurrentAppend is the AC-HEV-003 race-safety guard: N goroutines append
// concurrently; under -race there must be no data race, and every line must be
// present and valid JSON (O_APPEND single-write semantics, REQ-HEV-007).
func TestConcurrentAppend(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, LedgerFileName)
	w := NewWriter(path)

	const n = 40
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			_ = w.Append(Row{
				SchemaVersion:     SchemaVersion,
				TS:                time.Now().UTC().Format(time.RFC3339),
				SessionID:         "concurrent",
				MatchedSubcommand: "run",
				RequestDigest:     "sha256:aaaaaaaaaaaa",
				Outcome:           OutcomeSuccess,
			})
		}(i)
	}
	wg.Wait()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	lines := 0
	for _, ln := range splitNonEmptyLines(data) {
		var r Row
		if err := json.Unmarshal([]byte(ln), &r); err != nil {
			t.Fatalf("concurrent append produced a corrupt/interleaved line: %q (%v)", ln, err)
		}
		lines++
	}
	if lines != n {
		t.Fatalf("got %d ledger lines, want %d (lost or interleaved appends)", lines, n)
	}
}

// TestReaderFilters covers composable subcommand/outcome/time-window filters
// (REQ-HEV-008 traceability for AC-HEV-001).
func TestReaderFilters(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, LedgerFileName)
	w := NewWriter(path)

	base := time.Date(2026, 7, 12, 3, 0, 0, 0, time.UTC)
	mk := func(sub string, oc Outcome, ts time.Time) Row {
		return Row{SchemaVersion: SchemaVersion, TS: ts.Format(time.RFC3339), MatchedSubcommand: sub, Outcome: oc}
	}
	rows := []Row{
		mk("run", OutcomeSuccess, base),
		mk("run", OutcomeFail, base.Add(time.Hour)),
		mk("plan", OutcomeSuccess, base.Add(2*time.Hour)),
		mk("sync", OutcomeAbort, base.Add(3*time.Hour)),
	}
	for _, r := range rows {
		if err := w.Append(r); err != nil {
			t.Fatal(err)
		}
	}
	rd := NewReader(path)

	got, _, _ := rd.Read(Filter{Subcommand: "run"})
	if len(got) != 2 {
		t.Errorf("subcommand=run -> %d rows, want 2", len(got))
	}
	got, _, _ = rd.Read(Filter{Outcome: "success"})
	if len(got) != 2 {
		t.Errorf("outcome=success -> %d rows, want 2", len(got))
	}
	got, _, _ = rd.Read(Filter{Subcommand: "run", Outcome: "fail"})
	if len(got) != 1 {
		t.Errorf("run+fail -> %d rows, want 1", len(got))
	}
	since := base.Add(90 * time.Minute)
	until := base.Add(150 * time.Minute)
	got, _, _ = rd.Read(Filter{Since: &since, Until: &until})
	if len(got) != 1 || got[0].MatchedSubcommand != "plan" {
		t.Errorf("time window -> %d rows (%v), want 1 plan", len(got), got)
	}
}

// TestReaderSkipsMalformed asserts malformed lines are skipped fail-open with a
// reported count and no panic (REQ-HEV-008).
func TestReaderSkipsMalformed(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, LedgerFileName)

	content := "" +
		`{"schema_version":1,"matched_subcommand":"run","outcome":"success"}` + "\n" +
		"this is not json\n" +
		"\n" + // blank line ignored, not counted
		`{"schema_version":1,"matched_subcommand":"plan","outcome":"fail"}` + "\n" +
		`{"broken":`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	rows, skipped, err := NewReader(path).Read(Filter{})
	if err != nil {
		t.Fatalf("reader returned error on malformed input (should fail-open): %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("got %d valid rows, want 2", len(rows))
	}
	if skipped != 2 {
		t.Errorf("got %d skipped, want 2 (the non-json + truncated lines)", skipped)
	}
}

// TestReaderAbsentLedger: an absent ledger yields empty result, not an error.
func TestReaderAbsentLedger(t *testing.T) {
	t.Parallel()
	rows, skipped, err := NewReader(filepath.Join(t.TempDir(), "nope.jsonl")).Read(Filter{})
	if err != nil || len(rows) != 0 || skipped != 0 {
		t.Fatalf("absent ledger: got rows=%d skipped=%d err=%v; want 0,0,nil", len(rows), skipped, err)
	}
}

// splitNonEmptyLines splits data on newlines, dropping empty trailing lines.
func splitNonEmptyLines(data []byte) []string {
	var out []string
	for _, ln := range splitLines(string(data)) {
		if ln != "" {
			out = append(out, ln)
		}
	}
	return out
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}
