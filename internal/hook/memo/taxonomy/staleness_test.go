package taxonomy_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/hook/memo/taxonomy"
)

// fixedNow is a stable reference time used across staleness tests.
var fixedNow = time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)

// writeMemoryFile creates a memory file at path and sets its mtime to age before now.
func writeMemoryFile(t *testing.T, path string, age time.Duration) {
	t.Helper()
	content := "---\nname: test\ndescription: test desc\ntype: user\n---\nbody\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile %s: %v", path, err)
	}
	modTime := fixedNow.Add(-age)
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatalf("Chtimes %s: %v", path, err)
	}
}

// TestDetectStale_Boundary verifies the 24h inclusive threshold.
// mtime exactly 24h ago → stale; 23h 59m ago → not stale; 25h ago → stale.
func TestDetectStale_Boundary(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		age       time.Duration
		wantStale bool
	}{
		{"23h59m not stale", 23*time.Hour + 59*time.Minute, false},
		{"24h exactly stale (inclusive)", 24 * time.Hour, true},
		{"25h stale", 25 * time.Hour, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			writeMemoryFile(t, filepath.Join(dir, "mem.md"), tt.age)

			reports, err := taxonomy.DetectStale(dir, 24, fixedNow)
			if err != nil {
				t.Fatalf("DetectStale error = %v", err)
			}
			gotStale := len(reports) > 0
			if gotStale != tt.wantStale {
				t.Errorf("stale = %v, want %v (age=%v)", gotStale, tt.wantStale, tt.age)
			}
		})
	}
}

// TestDetectStale_EmptyDir verifies that an empty or non-existent directory
// returns no error and no stale reports.
func TestDetectStale_EmptyDir(t *testing.T) {
	t.Parallel()
	// Non-existent directory
	reports, err := taxonomy.DetectStale("/nonexistent/path/that/does/not/exist", 24, fixedNow)
	if err != nil {
		t.Fatalf("DetectStale(non-existent) error = %v, want nil", err)
	}
	if len(reports) != 0 {
		t.Errorf("reports = %d, want 0", len(reports))
	}

	// Empty directory
	dir := t.TempDir()
	reports, err = taxonomy.DetectStale(dir, 24, fixedNow)
	if err != nil {
		t.Fatalf("DetectStale(empty dir) error = %v, want nil", err)
	}
	if len(reports) != 0 {
		t.Errorf("reports = %d, want 0", len(reports))
	}
}

// TestDetectStale_InjectedNow verifies time injection ensures test stability
// (no dependency on real wall clock).
func TestDetectStale_InjectedNow(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// File written at a specific time
	fileTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	path := filepath.Join(dir, "mem.md")
	content := "---\nname: t\ndescription: d\ntype: user\n---\nbody\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, fileTime, fileTime); err != nil {
		t.Fatal(err)
	}

	// now = fileTime + 30h → file is 30h old → stale at 24h threshold
	now := fileTime.Add(30 * time.Hour)
	reports, err := taxonomy.DetectStale(dir, 24, now)
	if err != nil {
		t.Fatalf("DetectStale error = %v", err)
	}
	if len(reports) != 1 {
		t.Errorf("len(reports) = %d, want 1", len(reports))
	}

	// now = fileTime + 10h → not stale
	now2 := fileTime.Add(10 * time.Hour)
	reports2, err := taxonomy.DetectStale(dir, 24, now2)
	if err != nil {
		t.Fatal(err)
	}
	if len(reports2) != 0 {
		t.Errorf("len(reports2) = %d, want 0", len(reports2))
	}
}

// TestDetectStale_WrappedContent verifies that stale report content is wrapped
// in <system-reminder> with the staleness caveat (REQ-EXT001-006).
func TestDetectStale_WrappedContent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeMemoryFile(t, filepath.Join(dir, "stale.md"), 25*time.Hour)

	reports, err := taxonomy.DetectStale(dir, 24, fixedNow)
	if err != nil {
		t.Fatalf("DetectStale error = %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("want 1 report, got %d", len(reports))
	}

	wrapped := reports[0].Wrapped
	if !strings.Contains(wrapped, "<system-reminder>") {
		t.Error("Wrapped does not contain <system-reminder>")
	}
	if !strings.Contains(wrapped, "verify against current state before acting on it") {
		t.Error("Wrapped does not contain staleness caveat")
	}
	if !strings.Contains(wrapped, "</system-reminder>") {
		t.Error("Wrapped does not contain </system-reminder>")
	}
}

// TestAggregateWarning_Counts verifies REQ-EXT001-017 aggregation behavior.
// < 10 files → per-file warnings; >= 10 files → single aggregated warning.
func TestAggregateWarning_Counts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		count       int
		wantAggr    bool // true = single aggregated line expected
		wantContain string
	}{
		{"zero", 0, false, ""},
		{"one", 1, false, ""},
		{"nine", 9, false, ""},
		{"ten exactly", 10, true, "10"},
		{"eleven", 11, true, "11"},
		{"twelve", 12, true, "12"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reports := makeStaleReports(tt.count)
			msg := taxonomy.AggregateWarning(reports)

			if tt.count == 0 {
				if msg != "" {
					t.Errorf("AggregateWarning(0) = %q, want empty", msg)
				}
				return
			}

			if tt.wantAggr {
				// Single aggregated line: should not contain per-file path listing
				// and should contain the count
				if !strings.Contains(msg, tt.wantContain) {
					t.Errorf("aggregated message %q does not contain count %q", msg, tt.wantContain)
				}
				lines := strings.Split(strings.TrimSpace(msg), "\n")
				if len(lines) > 2 {
					// Aggregated should be short (1-2 lines), not per-file
					t.Errorf("aggregated message has %d lines, want <= 2 lines:\n%s", len(lines), msg)
				}
			} else {
				// Per-file: should contain each file path
				for _, r := range reports {
					if !strings.Contains(msg, r.Path) {
						t.Errorf("per-file message does not contain path %q", r.Path)
					}
				}
			}
		})
	}
}

// TestAggregateWarning_ExactlyTen verifies the exact threshold boundary (10 = aggregation).
func TestAggregateWarning_ExactlyTen(t *testing.T) {
	t.Parallel()
	reports := makeStaleReports(10)
	msg := taxonomy.AggregateWarning(reports)
	if !strings.Contains(msg, "10") {
		t.Errorf("AggregateWarning(10) = %q, want message containing '10'", msg)
	}
}

// TestAggregateWarning_Nine verifies that 9 stale files keep per-file warnings.
func TestAggregateWarning_Nine(t *testing.T) {
	t.Parallel()
	reports := makeStaleReports(9)
	msg := taxonomy.AggregateWarning(reports)
	// Each path should appear in the message
	for _, r := range reports {
		if !strings.Contains(msg, r.Path) {
			t.Errorf("per-file message does not contain path %q", r.Path)
		}
	}
}

// makeStaleReports creates n synthetic StaleReport values for testing.
func makeStaleReports(n int) []taxonomy.StaleReport {
	reports := make([]taxonomy.StaleReport, n)
	for i := range reports {
		reports[i] = taxonomy.StaleReport{
			Path:    filepath.Join("/agent-memory", "expert-backend", fmt.Sprintf("note%02d.md", i)),
			Age:     25 * time.Hour,
			Wrapped: "<system-reminder>stale content</system-reminder>",
		}
	}
	return reports
}

