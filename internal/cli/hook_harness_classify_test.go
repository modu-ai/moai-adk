// Package cli — hook harness-classify subcommand tests.
//
// SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001 (HCW)
// REQ-HCW-001: Workflow body §2.1 status verb invokes the Go classifier.
// REQ-HCW-002: Classifier reads usage-log.jsonl, aggregates patterns, classifies
//              tiers, and writes promotions to tier-promotions.jsonl.
// REQ-HCW-003: On classifier error, hook surfaces error annotation (stderr) and
//              continues — fail-open with exit code 1 (workflow body interprets
//              exit 1 as the trigger for the error annotation without aborting).
// REQ-HCW-004: When learning.enabled is false, the hook is a complete no-op —
//              promotions file is NOT created and stderr is silent.
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// writeUsageLog writes a usage-log.jsonl file under dir/.moai/harness/.
// Each line in lines is appended verbatim as a JSONL entry.
func writeUsageLog(t *testing.T, dir string, lines []string) {
	t.Helper()
	harnessDir := filepath.Join(dir, ".moai", "harness")
	if err := os.MkdirAll(harnessDir, 0o755); err != nil {
		t.Fatalf("mkdir harness dir: %v", err)
	}
	path := filepath.Join(harnessDir, "usage-log.jsonl")
	body := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write usage-log.jsonl: %v", err)
	}
}

// ─────────────────────────────────────────────
// TestRunHarnessClassify_NoOpWhenLearningDisabled (AC-HCW-004 coverage)
// ─────────────────────────────────────────────

// Verifies REQ-HCW-004: when learning.enabled is false, runHarnessClassify is a
// complete no-op — does NOT create tier-promotions.jsonl and stderr is silent.
func TestRunHarnessClassify_NoOpWhenLearningDisabled(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: false\n")

	// Seed usage-log so the test would otherwise produce promotions.
	writeUsageLog(t, dir, []string{
		`{"timestamp":"2026-05-24T00:00:00Z","event_type":"agent_invocation","subject":"Bash","context_hash":"","schema_version":"v1"}`,
	})

	t.Chdir(dir)

	var stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetErr(&stderr)

	if err := runHarnessClassify(cmd, nil); err != nil {
		t.Fatalf("runHarnessClassify returned error in no-op path: %v", err)
	}

	// Verify tier-promotions.jsonl was NOT created.
	promoPath := filepath.Join(dir, ".moai", "harness", "learning-history", "tier-promotions.jsonl")
	if _, err := os.Stat(promoPath); !os.IsNotExist(err) {
		t.Errorf("tier-promotions.jsonl must not exist when learning.enabled is false; stat err=%v", err)
	}

	// Stderr should be silent on no-op.
	if stderr.Len() != 0 {
		t.Errorf("stderr must be empty on no-op; got: %q", stderr.String())
	}
}

// ─────────────────────────────────────────────
// TestRunHarnessClassify_WritesPromotionsOnHappyPath (AC-HCW-001 + AC-HCW-002)
// ─────────────────────────────────────────────

// Verifies REQ-HCW-001 + REQ-HCW-002: when learning is enabled and the usage
// log contains valid entries, runHarnessClassify writes promotions to
// tier-promotions.jsonl. The promotion count matches the number of unique
// (event_type, subject, context_hash) patterns aggregated from the log.
func TestRunHarnessClassify_WritesPromotionsOnHappyPath(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")

	// Seed three usage-log entries — two share the same (event_type, subject,
	// context_hash) tuple so the aggregator collapses them into a single
	// pattern. Net unique patterns = 2.
	writeUsageLog(t, dir, []string{
		`{"timestamp":"2026-05-24T00:00:00Z","event_type":"agent_invocation","subject":"Bash","context_hash":"","schema_version":"v1"}`,
		`{"timestamp":"2026-05-24T00:01:00Z","event_type":"agent_invocation","subject":"Bash","context_hash":"","schema_version":"v1"}`,
		`{"timestamp":"2026-05-24T00:02:00Z","event_type":"agent_invocation","subject":"Edit","context_hash":"","schema_version":"v1"}`,
	})

	t.Chdir(dir)

	var stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetErr(&stderr)

	if err := runHarnessClassify(cmd, nil); err != nil {
		t.Fatalf("runHarnessClassify returned error on happy path: %v", err)
	}

	// Verify tier-promotions.jsonl exists and is non-empty.
	promoPath := filepath.Join(dir, ".moai", "harness", "learning-history", "tier-promotions.jsonl")
	data, err := os.ReadFile(promoPath)
	if err != nil {
		t.Fatalf("tier-promotions.jsonl read: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("tier-promotions.jsonl must be non-empty after classifier invocation")
	}

	// Count promotion lines — exactly 2 unique patterns expected.
	lineCount := strings.Count(strings.TrimRight(string(data), "\n"), "\n") + 1
	if lineCount != 2 {
		t.Errorf("expected 2 promotion entries (one per unique pattern), got %d\nfile body:\n%s", lineCount, string(data))
	}

	// Verify stderr summary line was emitted.
	stderrOut := stderr.String()
	if !strings.Contains(stderrOut, "harness-classify") {
		t.Errorf("stderr summary line missing; got: %q", stderrOut)
	}
}

// ─────────────────────────────────────────────
// TestRunHarnessClassify_CorruptEntryFailOpen (AC-HCW-003 coverage)
// ─────────────────────────────────────────────

// Verifies REQ-HCW-003: a corrupt JSON entry in usage-log.jsonl does NOT crash
// the classifier. The aggregator skips invalid lines (see harness.AggregatePatterns
// scanner loop which silently skips json.Unmarshal failures) and continues with
// the valid entries. The hook returns nil (workflow body interprets exit 0 as
// success). This validates fail-open behavior at the AggregatePatterns layer;
// any future error path (e.g., disk I/O failure) is exercised separately via
// runtime injection.
func TestRunHarnessClassify_CorruptEntryFailOpen(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")

	// Mix a valid entry with a corrupt one. The valid entry should still
	// produce a promotion; the corrupt entry is silently skipped per
	// AggregatePatterns scanner semantics.
	writeUsageLog(t, dir, []string{
		`{"timestamp":"2026-05-24T00:00:00Z","event_type":"agent_invocation","subject":"Bash","context_hash":"","schema_version":"v1"}`,
		`{this is not valid json`,
	})

	t.Chdir(dir)

	var stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetErr(&stderr)

	if err := runHarnessClassify(cmd, nil); err != nil {
		t.Fatalf("runHarnessClassify returned error despite corrupt entry being skippable: %v", err)
	}

	// Verify the valid entry still produced a promotion (fail-open continues).
	promoPath := filepath.Join(dir, ".moai", "harness", "learning-history", "tier-promotions.jsonl")
	data, err := os.ReadFile(promoPath)
	if err != nil {
		t.Fatalf("tier-promotions.jsonl read after corrupt entry: %v", err)
	}
	if len(data) == 0 {
		t.Error("tier-promotions.jsonl must contain promotion from the valid entry despite the corrupt sibling")
	}
}
