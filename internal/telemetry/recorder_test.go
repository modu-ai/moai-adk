package telemetry

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestRecordSkillUsage_AppendsJSONLLine verifies that RecordSkillUsage appends
// a valid JSONL line to the daily telemetry file.
func TestRecordSkillUsage_AppendsJSONLLine(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Ensure .moai directory exists to pass the project root guard
	if err := os.MkdirAll(filepath.Join(dir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}

	ts := time.Date(2026, 4, 11, 10, 30, 0, 0, time.UTC)
	r := UsageRecord{
		Timestamp:   ts,
		SessionID:   "sess-abc",
		SkillID:     "moai-workflow-tdd",
		Trigger:     TriggerExplicit,
		ContextHash: "a1b2c3d4",
		AgentType:   "manager-tdd",
		Phase:       "run",
		DurationMs:  45000,
		Outcome:     OutcomeUnknown,
	}

	if err := RecordSkillUsage(dir, r); err != nil {
		t.Fatalf("RecordSkillUsage() error = %v", err)
	}

	// Verify file exists
	expectedPath := filepath.Join(dir, ".moai", "evolution", "telemetry", "usage-2026-04-11.jsonl")
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("expected telemetry file not found: %v", err)
	}

	// Parse and verify record
	var got UsageRecord
	if err := json.Unmarshal(data[:len(data)-1], &got); err != nil { // strip trailing newline
		t.Fatalf("failed to parse JSONL line: %v", err)
	}

	if got.SkillID != r.SkillID {
		t.Errorf("SkillID = %q, want %q", got.SkillID, r.SkillID)
	}
	if got.SessionID != r.SessionID {
		t.Errorf("SessionID = %q, want %q", got.SessionID, r.SessionID)
	}
	if got.Outcome != r.Outcome {
		t.Errorf("Outcome = %q, want %q", got.Outcome, r.Outcome)
	}
	if got.Trigger != r.Trigger {
		t.Errorf("Trigger = %q, want %q", got.Trigger, r.Trigger)
	}
}

// TestRecordSkillUsage_AppendMultipleLines verifies that multiple records are
// appended as separate JSONL lines.
func TestRecordSkillUsage_AppendMultipleLines(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}

	ts := time.Date(2026, 4, 11, 10, 30, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		r := UsageRecord{
			Timestamp: ts,
			SessionID: "sess-abc",
			SkillID:   "moai-workflow-tdd",
			Outcome:   OutcomeUnknown,
		}
		if err := RecordSkillUsage(dir, r); err != nil {
			t.Fatalf("RecordSkillUsage() error = %v on iteration %d", err, i)
		}
	}

	expectedPath := filepath.Join(dir, ".moai", "evolution", "telemetry", "usage-2026-04-11.jsonl")
	f, err := os.Open(expectedPath)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		count++
	}
	if count != 3 {
		t.Errorf("got %d JSONL lines, want 3", count)
	}
}

// TestHashContext_Returns8CharHex verifies that HashContext returns exactly
// 8 hexadecimal characters (SHA-256 truncation).
func TestHashContext_Returns8CharHex(t *testing.T) {
	t.Parallel()

	inputs := []string{
		"implement authentication feature",
		"",
		"fix bug in parser",
		"a very long task description that contains lots of detail about what needs to be done",
	}

	for _, input := range inputs {
		got := HashContext(input)
		if len(got) != 8 {
			t.Errorf("HashContext(%q) = %q (len=%d), want 8 chars", input, got, len(got))
		}
		// Verify it's hex
		for _, c := range got {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				t.Errorf("HashContext(%q) = %q contains non-hex char %q", input, got, c)
			}
		}
	}
}

// TestHashContext_IsDeterministic verifies that the same input always produces
// the same hash (no PII capture, just identity check).
func TestHashContext_IsDeterministic(t *testing.T) {
	t.Parallel()

	input := "implement user authentication"
	h1 := HashContext(input)
	h2 := HashContext(input)

	if h1 != h2 {
		t.Errorf("HashContext is not deterministic: got %q then %q", h1, h2)
	}
}

// TestRecordSkillUsage_DailyRotation verifies that records with different dates
// go into separate files.
func TestRecordSkillUsage_DailyRotation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}

	day1 := time.Date(2026, 4, 10, 10, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC)

	r1 := UsageRecord{Timestamp: day1, SkillID: "skill-a", Outcome: OutcomeUnknown}
	r2 := UsageRecord{Timestamp: day2, SkillID: "skill-b", Outcome: OutcomeUnknown}

	if err := RecordSkillUsage(dir, r1); err != nil {
		t.Fatal(err)
	}
	if err := RecordSkillUsage(dir, r2); err != nil {
		t.Fatal(err)
	}

	telDir := filepath.Join(dir, ".moai", "evolution", "telemetry")
	entries, err := os.ReadDir(telDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 daily files, got %d", len(entries))
	}

	// Verify filenames
	names := make(map[string]bool)
	for _, e := range entries {
		names[e.Name()] = true
	}
	if !names["usage-2026-04-10.jsonl"] {
		t.Error("missing usage-2026-04-10.jsonl")
	}
	if !names["usage-2026-04-11.jsonl"] {
		t.Error("missing usage-2026-04-11.jsonl")
	}
}

// TestRecordSkillUsage_HandlesMissingDirectory verifies that RecordSkillUsage
// creates the telemetry directory when it doesn't exist.
func TestRecordSkillUsage_HandlesMissingDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Create .moai dir to pass the project root guard, but NOT evolution/telemetry
	if err := os.MkdirAll(filepath.Join(dir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}

	r := UsageRecord{
		Timestamp: time.Now().UTC(),
		SkillID:   "test-skill",
		Outcome:   OutcomeUnknown,
	}

	if err := RecordSkillUsage(dir, r); err != nil {
		t.Errorf("RecordSkillUsage() with missing dir error = %v, want nil", err)
	}
}

// TestRecordSkillUsage_ConcurrentWrites verifies that concurrent writes do not
// corrupt the JSONL file.
func TestRecordSkillUsage_ConcurrentWrites(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}

	ts := time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC)
	const numGoroutines = 20

	var wg sync.WaitGroup
	errs := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			r := UsageRecord{
				Timestamp: ts,
				SkillID:   "concurrent-skill",
				SessionID: "sess-concurrent",
				Outcome:   OutcomeUnknown,
			}
			errs[idx] = RecordSkillUsage(dir, r)
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: RecordSkillUsage() error = %v", i, err)
		}
	}

	// Verify all lines were written and are valid JSON
	expectedPath := filepath.Join(dir, ".moai", "evolution", "telemetry", "usage-2026-04-11.jsonl")
	f, err := os.Open(expectedPath)
	if err != nil {
		t.Fatalf("failed to open telemetry file: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		line := scanner.Text()
		var rec UsageRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Errorf("invalid JSON at line %d: %v", count+1, err)
		}
		count++
	}

	if count != numGoroutines {
		t.Errorf("got %d lines, want %d", count, numGoroutines)
	}
}

// TestPruneOldFiles_DeletesExpiredFiles verifies that files older than
// retentionDays are deleted.
func TestPruneOldFiles_DeletesExpiredFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	telDir := filepath.Join(dir, ".moai", "evolution", "telemetry")
	if err := os.MkdirAll(telDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create old file (91 days ago)
	oldDate := time.Now().UTC().AddDate(0, 0, -91).Format("2006-01-02")
	oldFile := filepath.Join(telDir, "usage-"+oldDate+".jsonl")
	if err := os.WriteFile(oldFile, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create recent file (yesterday)
	recentDate := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
	recentFile := filepath.Join(telDir, "usage-"+recentDate+".jsonl")
	if err := os.WriteFile(recentFile, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := PruneOldFiles(dir, 90); err != nil {
		t.Fatalf("PruneOldFiles() error = %v", err)
	}

	// Old file should be gone
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Errorf("expected old file to be deleted, but it still exists")
	}

	// Recent file should remain
	if _, err := os.Stat(recentFile); err != nil {
		t.Errorf("expected recent file to remain, but got error: %v", err)
	}
}

// TestPruneOldFiles_HandlesEmptyDirectory verifies that PruneOldFiles handles
// a missing or empty telemetry directory gracefully.
func TestPruneOldFiles_HandlesEmptyDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Don't create the telemetry directory

	// Should not return an error
	if err := PruneOldFiles(dir, 90); err != nil {
		t.Errorf("PruneOldFiles() with missing dir error = %v, want nil", err)
	}
}

// TestPruneOldFiles_IgnoresNonDateFiles verifies that PruneOldFiles ignores
// files that don't match the usage-YYYY-MM-DD.jsonl pattern.
func TestPruneOldFiles_IgnoresNonDateFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	telDir := filepath.Join(dir, ".moai", "evolution", "telemetry")
	if err := os.MkdirAll(telDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a file that doesn't match the pattern
	otherFile := filepath.Join(telDir, "summary.json")
	if err := os.WriteFile(otherFile, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	// PruneOldFiles should not delete non-pattern files
	if err := PruneOldFiles(dir, 90); err != nil {
		t.Fatalf("PruneOldFiles() error = %v", err)
	}

	if _, err := os.Stat(otherFile); err != nil {
		t.Errorf("non-pattern file was unexpectedly deleted: %v", err)
	}
}

// TestRecordSkillUsage_AllOutcomeTypes verifies that all outcome constants
// are valid JSON strings when marshaled.
func TestRecordSkillUsage_AllOutcomeTypes(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}

	outcomes := []string{OutcomeSuccess, OutcomePartial, OutcomeError, OutcomeUnknown}
	triggers := []string{TriggerExplicit, TriggerAuto}

	ts := time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)
	for _, outcome := range outcomes {
		for _, trigger := range triggers {
			r := UsageRecord{
				Timestamp: ts,
				SkillID:   "test-skill",
				Outcome:   outcome,
				Trigger:   trigger,
			}
			if err := RecordSkillUsage(dir, r); err != nil {
				t.Errorf("RecordSkillUsage(outcome=%q, trigger=%q) error = %v", outcome, trigger, err)
			}
		}
	}
}
