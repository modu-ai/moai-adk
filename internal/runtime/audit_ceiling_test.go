package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeRoundFile writes one plan-audit round file into dir.
func writeRoundFile(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("verdict: PASS\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}

// evidenceDir makes a fresh directory holding the named round files.
func evidenceDir(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range names {
		writeRoundFile(t, dir, name)
	}
	return dir
}

// TestCountPlanAuditRounds — AC-ACR-001: one recorded file one round, counted
// from durable evidence on disk across the SPEC-scoped directory (absent → 0)
// and the explicitly listed directories (absent → error, tested by
// TestCountPlanAuditRoundsListedDirMissing); identical listed paths count once.
func TestCountPlanAuditRounds(t *testing.T) {
	t.Parallel()

	t.Run("plan-audit.md plus iter1..3 counts 4", func(t *testing.T) {
		t.Parallel()
		dir := evidenceDir(t, "plan-audit.md", "plan-audit-iter1.md", "plan-audit-iter2.md", "plan-audit-iter3.md")
		got, err := CountPlanAuditRounds(dir, nil)
		if err != nil {
			t.Fatalf("CountPlanAuditRounds: %v", err)
		}
		if got != 4 {
			t.Fatalf("count = %d, want 4", got)
		}
	})

	t.Run("SPEC-scoped directory absent contributes 0", func(t *testing.T) {
		t.Parallel()
		absent := filepath.Join(t.TempDir(), "SPEC-UNAUDITED-001")
		got, err := CountPlanAuditRounds(absent, nil)
		if err != nil {
			t.Fatalf("absent SPEC-scoped directory must not error: %v", err)
		}
		if got != 0 {
			t.Fatalf("count = %d, want 0", got)
		}
	})

	t.Run("identical listed paths are deduplicated", func(t *testing.T) {
		t.Parallel()
		dir := evidenceDir(t, "plan-audit-iter1.md")
		dup := filepath.Join(dir, "nested")
		if err := os.MkdirAll(dup, 0o755); err != nil {
			t.Fatal(err)
		}
		// The same directory listed twice contributes once; a path equal to the
		// SPEC-scoped one does not double it either.
		got, err := CountPlanAuditRounds(dir, []string{dir, filepath.Clean(dir) + "/", dir})
		if err != nil {
			t.Fatalf("CountPlanAuditRounds: %v", err)
		}
		if got != 1 {
			t.Fatalf("count = %d, want 1 (identical listings deduplicate)", got)
		}
	})

	t.Run("non round-family files are ignored", func(t *testing.T) {
		t.Parallel()
		dir := evidenceDir(t, "plan-audit-iter1.md", "sync-audit.md", "notes.md", "plan-audit.md.bak")
		got, err := CountPlanAuditRounds(dir, nil)
		if err != nil {
			t.Fatalf("CountPlanAuditRounds: %v", err)
		}
		if got != 1 {
			t.Fatalf("count = %d, want 1 (only the iter family counts)", got)
		}
	})
}

// TestCountPlanAuditRoundsUnparseable — AC-ACR-002: a round file whose
// iteration suffix does not parse as a positive integer makes the count an
// error naming the file, never a silent skip.
func TestCountPlanAuditRoundsUnparseable(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"plan-audit-iterX.md", "plan-audit-iter0.md", "plan-audit-iter-1.md", "plan-audit-iter.md"} {
		dir := evidenceDir(t, name)
		_, err := CountPlanAuditRounds(dir, nil)
		if err == nil {
			t.Errorf("%s: expected an error, got nil", name)
			continue
		}
		if !strings.Contains(err.Error(), name) {
			t.Errorf("%s: error does not name the file: %v", name, err)
		}
	}
}

// TestCountPlanAuditRoundsListedDirMissing — AC-ACR-013: an EXPLICITLY LISTED
// evidence directory that does not exist (a typo) is an error naming the
// missing path, while the SPEC-scoped directory being absent still contributes
// 0 (an unaudited SPEC is not an error).
func TestCountPlanAuditRoundsListedDirMissing(t *testing.T) {
	t.Parallel()

	missing := filepath.Join(t.TempDir(), "typo-evidence")
	_, err := CountPlanAuditRounds(filepath.Join(t.TempDir(), "SPEC-UNAUDITED-002"), []string{missing})
	if err == nil {
		t.Fatal("a missing explicitly listed directory must error, got nil")
	}
	if !strings.Contains(err.Error(), missing) {
		t.Fatalf("error does not name the missing path: %v", err)
	}
}

// TestLatestVerdictTieIsError — AC-ACR-001's tie arm: the same highest
// iteration number in two evidence directories is a selection error rather
// than a silent pick.
func TestLatestVerdictTieIsError(t *testing.T) {
	t.Parallel()

	dirA := evidenceDir(t, "plan-audit-iter2.md")
	dirB := evidenceDir(t, "plan-audit-iter2.md")
	_, err := SelectLatestVerdict(dirA, []string{dirB})
	if err == nil {
		t.Fatal("a cross-directory iteration tie must error, got nil")
	}

	// A strict ordering across directories selects the highest.
	dirC := evidenceDir(t, "plan-audit.md", "plan-audit-iter5.md")
	got, err := SelectLatestVerdict(dirA, []string{dirC})
	if err != nil {
		t.Fatalf("SelectLatestVerdict: %v", err)
	}
	if !strings.HasSuffix(got, filepath.Join("plan-audit-iter5.md")) || !strings.Contains(got, filepath.Base(dirC)) {
		t.Fatalf("latest = %q, want dirC's plan-audit-iter5.md", got)
	}

	// plan-audit.md alone ranks as iteration 0 and is selected.
	dirD := evidenceDir(t, "plan-audit.md")
	got, err = SelectLatestVerdict(dirD, nil)
	if err != nil {
		t.Fatalf("SelectLatestVerdict: %v", err)
	}
	if !strings.HasSuffix(got, "plan-audit.md") {
		t.Fatalf("latest = %q, want plan-audit.md", got)
	}

	// No evidence at all selects nothing and does not error.
	got, err = SelectLatestVerdict(filepath.Join(t.TempDir(), "SPEC-UNAUDITED-003"), nil)
	if err != nil {
		t.Fatalf("empty evidence must not error: %v", err)
	}
	if got != "" {
		t.Fatalf("latest = %q, want empty", got)
	}
}

// TestResolvePlanAuditCeiling — AC-ACR-003's resolution arms: the shipped map
// resolves S:1 M:2 L:3, and an absent or unknown tier resolves to L
// (auditverdict.SpecTier's rule).
func TestResolvePlanAuditCeiling(t *testing.T) {
	t.Parallel()

	shipped := map[string]int{"S": 1, "M": 2, "L": 3}
	cases := []struct {
		tier string
		want int
	}{
		{"S", 1},
		{"M", 2},
		{"L", 3},
		{"", 3},     // absent tier → L
		{"bogus", 3}, // unknown tier → L
	}
	for _, c := range cases {
		got, err := ResolvePlanAuditCeiling(c.tier, shipped)
		if err != nil {
			t.Fatalf("tier %q: %v", c.tier, err)
		}
		if got != c.want {
			t.Errorf("tier %q: ceiling = %d, want %d", c.tier, got, c.want)
		}
	}
}

// TestResolvePlanAuditCeilingInvalid — AC-ACR-003's configuration-error arms:
// a ceilings map missing the resolved tier's key, or resolving ≤ 0, is a
// configuration error, never a ceiling of zero.
func TestResolvePlanAuditCeilingInvalid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		tier     string
		ceilings map[string]int
	}{
		{"missing tier key", "M", map[string]int{"S": 1, "L": 3}},
		{"missing resolved L key", "", map[string]int{"S": 1, "M": 2}},
		{"zero ceiling", "M", map[string]int{"S": 1, "M": 0, "L": 3}},
		{"negative ceiling", "L", map[string]int{"S": 1, "M": 2, "L": -2}},
		{"nil map", "S", nil},
	}
	for _, c := range cases {
		got, err := ResolvePlanAuditCeiling(c.tier, c.ceilings)
		if err == nil {
			t.Errorf("%s: expected a configuration error, got ceiling %d", c.name, got)
		}
	}
}
