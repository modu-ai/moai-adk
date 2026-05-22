package cli

// TestMergeHistory_* exercise the 3-way merge fallback counter introduced by
// SPEC-V3R6-UPDATE-NOISE-001 (REQ-UN-006 through REQ-UN-011). The tests do not
// invoke the real 3-way merge path inside update.go — they exercise
// recordMergeFallback directly, which is the unit boundary the SPEC's
// acceptance criteria target.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// loadHistory is a small helper that decodes the merge-history ledger from disk.
func loadHistory(t *testing.T, projectRoot string) map[string]mergeHistoryEntry {
	t.Helper()
	path := filepath.Join(projectRoot, ".moai", "cache", "merge-history.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]mergeHistoryEntry{}
	}
	ledger := map[string]mergeHistoryEntry{}
	if jsonErr := json.Unmarshal(data, &ledger); jsonErr != nil {
		t.Fatalf("history JSON parse: %v", jsonErr)
	}
	return ledger
}

// TestMergeHistory_FirstTwoFallbacksSilent covers AC-UN-005 + REQ-UN-007. The
// counter increments on each failure but no output is produced before the
// 3-strike threshold.
func TestMergeHistory_FirstTwoFallbacksSilent(t *testing.T) {
	projectRoot := t.TempDir()

	var buf bytes.Buffer
	recordMergeFallback(projectRoot, "quality.yaml", false /*success*/, false /*verbose*/, &buf)
	recordMergeFallback(projectRoot, "quality.yaml", false, false, &buf)

	if buf.Len() != 0 {
		t.Errorf("first 2 fallbacks must be silent, got: %q", buf.String())
	}

	hist := loadHistory(t, projectRoot)
	if got := hist["quality.yaml"].FallbackCount; got != 2 {
		t.Errorf("FallbackCount after 2 failures = %d, want 2", got)
	}
}

// TestMergeHistory_ThirdFallbackEmitsAdvisory covers AC-UN-006 + REQ-UN-008.
// The advisory MUST be emitted exactly once when the count crosses the
// threshold (== 3) and the wording MUST match the acceptance grep pattern.
func TestMergeHistory_ThirdFallbackEmitsAdvisory(t *testing.T) {
	projectRoot := t.TempDir()

	var buf bytes.Buffer
	for i := 0; i < 3; i++ {
		recordMergeFallback(projectRoot, "quality.yaml", false, false, &buf)
	}

	out := buf.String()
	expected := "hint: 'moai update -c' to resync templates for quality.yaml\n"
	if !strings.Contains(out, expected) {
		t.Errorf("advisory missing or wrong wording.\n got: %q\nwant: contains %q", out, expected)
	}
	// Exactly one advisory line — defensive parse.
	if got := strings.Count(out, "hint: 'moai update -c'"); got != 1 {
		t.Errorf("advisory must appear exactly once across 3 failures, got %d (output=%q)", got, out)
	}

	hist := loadHistory(t, projectRoot)
	if got := hist["quality.yaml"].FallbackCount; got != 3 {
		t.Errorf("FallbackCount after 3 failures = %d, want 3", got)
	}
}

// TestMergeHistory_FourthFallbackSilent covers AC-UN-006 second clause +
// REQ-UN-008 once-per-relPath rule. Counts ≥4 must stay silent until a 3-way
// success resets the counter.
func TestMergeHistory_FourthFallbackSilent(t *testing.T) {
	projectRoot := t.TempDir()

	// Drive counter to 3, consuming the advisory.
	var prime bytes.Buffer
	for i := 0; i < 3; i++ {
		recordMergeFallback(projectRoot, "quality.yaml", false, false, &prime)
	}

	// 4th failure must be silent.
	var fourth bytes.Buffer
	recordMergeFallback(projectRoot, "quality.yaml", false, false, &fourth)
	if fourth.Len() != 0 {
		t.Errorf("4th fallback must be silent, got: %q", fourth.String())
	}

	hist := loadHistory(t, projectRoot)
	if got := hist["quality.yaml"].FallbackCount; got != 4 {
		t.Errorf("FallbackCount after 4 failures = %d, want 4", got)
	}
}

// TestMergeHistory_SuccessResetsCounter covers AC-UN-008 + REQ-UN-009. After a
// 3-way success the counter MUST drop to zero and the next failure starts a
// fresh 3-strike cycle.
func TestMergeHistory_SuccessResetsCounter(t *testing.T) {
	projectRoot := t.TempDir()

	var buf bytes.Buffer
	recordMergeFallback(projectRoot, "config.yaml", false, false, &buf)
	recordMergeFallback(projectRoot, "config.yaml", false, false, &buf)

	hist := loadHistory(t, projectRoot)
	if got := hist["config.yaml"].FallbackCount; got != 2 {
		t.Fatalf("FallbackCount after 2 failures = %d, want 2", got)
	}

	// Simulate a 3-way success.
	recordMergeFallback(projectRoot, "config.yaml", true, false, &buf)

	hist = loadHistory(t, projectRoot)
	if got := hist["config.yaml"].FallbackCount; got != 0 {
		t.Errorf("FallbackCount after success reset = %d, want 0", got)
	}

	// Next 2 failures must again be silent (fresh count == 2 < threshold).
	resetBuf := bytes.Buffer{}
	recordMergeFallback(projectRoot, "config.yaml", false, false, &resetBuf)
	recordMergeFallback(projectRoot, "config.yaml", false, false, &resetBuf)
	if resetBuf.Len() != 0 {
		t.Errorf("counter not actually reset — got output: %q", resetBuf.String())
	}
}

// TestMergeHistory_VerboseBypass covers AC-UN-007 (merge half) + REQ-UN-010.
// Under updateVerboseMode the legacy "Warning: 3-way merge failed for X,
// falling back to 2-way" message MUST be emitted on EVERY failure and the
// advisory threshold logic is bypassed.
func TestMergeHistory_VerboseBypass(t *testing.T) {
	projectRoot := t.TempDir()

	var buf bytes.Buffer
	// Each call passes verbose=true directly — the helper does not consult
	// the package-level flag, mirroring how update.go threads the value.
	for i := 0; i < 4; i++ {
		recordMergeFallback(projectRoot, "quality.yaml", false, true, &buf)
	}

	out := buf.String()
	// Legacy message should appear 4 times — one per failure.
	if got := strings.Count(out, "Warning: 3-way merge failed for quality.yaml, falling back to 2-way"); got != 4 {
		t.Errorf("verbose mode must emit legacy warning per-failure: got %d, want 4 (out=%q)", got, out)
	}
	// Advisory must NOT appear under verbose path.
	if strings.Contains(out, "hint: 'moai update -c'") {
		t.Errorf("advisory must not appear under verbose mode, got: %q", out)
	}
}

// TestMergeHistory_CorruptedLedgerRecovers covers AC-UN-009 (merge half) +
// REQ-UN-011. A malformed ledger MUST be silently treated as empty.
func TestMergeHistory_CorruptedLedgerRecovers(t *testing.T) {
	projectRoot := t.TempDir()

	cacheDir := filepath.Join(projectRoot, ".moai", "cache")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatalf("mkdir cache: %v", err)
	}
	corruptPath := filepath.Join(cacheDir, "merge-history.json")
	if err := os.WriteFile(corruptPath, []byte("}}}not json"), 0o644); err != nil {
		t.Fatalf("plant corrupt ledger: %v", err)
	}

	var buf bytes.Buffer
	// Corrupt ledger → treated as empty → 1st failure silent.
	recordMergeFallback(projectRoot, "x.yaml", false, false, &buf)
	if buf.Len() != 0 {
		t.Errorf("post-corruption first fallback must be silent, got: %q", buf.String())
	}

	// The ledger should now contain a valid record with count=1.
	hist := loadHistory(t, projectRoot)
	if got := hist["x.yaml"].FallbackCount; got != 1 {
		t.Errorf("ledger not recovered properly: count=%d want=1", got)
	}
}
