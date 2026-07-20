// section_writer_test.go — TDD coverage for the progress.md §I writer (M3).
// SPEC-TOKEN-ACCOUNTING-001 AC-TA-008, AC-TA-009 (parser-safety companion,
// verified in internal/spec), AC-TA-011 (source guard — era.go untouched).
package tokenusage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sampleAttribution returns a representative Attribution for test assertions.
func sampleAttribution() Attribution {
	return Attribution{
		Usage: Usage{
			TokensInput:         300,
			TokensOutput:        60,
			TokensCacheCreation: 0,
			TokensCacheRead:     1500,
			TokensSpent:         1860,
			CacheHitRatio:       CacheHitRatio(300, 0, 1500),
		},
		AttributionMethod:    AttributionSessionSet,
		Confidence:           ConfidenceHigh,
		SessionCount:         2,
		ContributingSessions: []string{"11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222"},
	}
}

func TestBuildSectionI_ContainsHeading(t *testing.T) {
	t.Parallel()
	got := BuildSectionI(sampleAttribution())
	if !strings.Contains(got, SectionIHeading) {
		t.Fatalf("BuildSectionI output missing canonical heading %q:\n%s", SectionIHeading, got)
	}
	for _, line := range strings.Split(got, "\n") {
		if strings.TrimSpace(line) == SectionIHeading {
			return // found at line start
		}
	}
	t.Fatalf("heading %q not present as a standalone line:\n%s", SectionIHeading, got)
}

func TestBuildSectionI_ContainsAllNineFields(t *testing.T) {
	t.Parallel()
	attr := sampleAttribution()
	got := BuildSectionI(attr)

	required := []string{
		"tokens_spent:",
		"tokens_input:",
		"tokens_output:",
		"tokens_cache_creation:",
		"tokens_cache_read:",
		"cache_hit_ratio:",
		"token_attribution:",
		"token_attribution_confidence:",
		"token_session_count:",
	}
	for _, f := range required {
		if !strings.Contains(got, f) {
			t.Errorf("BuildSectionI missing field %q", f)
		}
	}
	// Spot-check concrete values derived from the Attribution struct.
	if !strings.Contains(got, "tokens_spent: 1860") {
		t.Errorf("tokens_spent value not rendered from attr.TokensSpent; got:\n%s", got)
	}
	if !strings.Contains(got, "token_attribution: session-set") {
		t.Errorf("token_attribution value not rendered from attr.AttributionMethod; got:\n%s", got)
	}
	if !strings.Contains(got, "token_attribution_confidence: high") {
		t.Errorf("token_attribution_confidence value not rendered from attr.Confidence; got:\n%s", got)
	}
	if !strings.Contains(got, "token_session_count: 2") {
		t.Errorf("token_session_count value not rendered from attr.SessionCount; got:\n%s", got)
	}
}

func TestWriteSectionI_AppendsWhenAbsent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "progress.md")
	seed := "# Progress — SPEC-X\n\n## §E.2 Run-phase Evidence\n- existing evidence\n\n## §E.4 Sync-phase Audit-Ready Signal\nsync_commit_sha: abc123\n"
	if err := os.WriteFile(path, []byte(seed), 0o644); err != nil {
		t.Fatalf("seed write: %v", err)
	}

	if err := WriteSectionI(path, sampleAttribution()); err != nil {
		t.Fatalf("WriteSectionI: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	s := string(got)

	// AC-TA-008 binary checks.
	if !strings.Contains(s, SectionIHeading) {
		t.Error("§I heading absent after append")
	}
	if !strings.Contains(s, "tokens_spent:") {
		t.Error("tokens_spent field absent after append")
	}
	// Existing sections preserved verbatim.
	if !strings.Contains(s, "## §E.2 Run-phase Evidence") {
		t.Error("§E.2 heading lost after append")
	}
	if !strings.Contains(s, "sync_commit_sha: abc123") {
		t.Error("sync_commit_sha field lost after append")
	}
}

func TestWriteSectionI_ReplacesPlaceholderWhenPresent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "progress.md")
	seed := "# Progress\n\n## §E.2 Run-phase Evidence\n- evidence\n\n## §I Token Accounting\n\n_<pending placeholder — values unrecorded>_\n\n<!--\n- tokens_spent: <int>\n-->\n"
	if err := os.WriteFile(path, []byte(seed), 0o644); err != nil {
		t.Fatalf("seed write: %v", err)
	}

	if err := WriteSectionI(path, sampleAttribution()); err != nil {
		t.Fatalf("WriteSectionI: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	s := string(got)

	if !strings.Contains(s, "tokens_spent: 1860") {
		t.Errorf("placeholder not replaced with measured value; got:\n%s", s)
	}
	// The placeholder markers MUST be gone (full section span replaced).
	if strings.Contains(s, "pending placeholder") {
		t.Errorf("placeholder prose survived replace; got:\n%s", s)
	}
	// Exactly one §I heading (idempotent replace, not duplicate).
	if c := strings.Count(s, SectionIHeading); c != 1 {
		t.Errorf("want exactly 1 §I heading, got %d; got:\n%s", c, s)
	}
	// §E.2 untouched.
	if !strings.Contains(s, "## §E.2 Run-phase Evidence") {
		t.Error("§E.2 heading lost during replace")
	}
}

func TestWriteSectionI_Idempotent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "progress.md")
	seed := "# Progress\n\n## §E.2 Run-phase Evidence\n- evidence\n"
	if err := os.WriteFile(path, []byte(seed), 0o644); err != nil {
		t.Fatalf("seed write: %v", err)
	}
	attr := sampleAttribution()

	if err := WriteSectionI(path, attr); err != nil {
		t.Fatalf("first WriteSectionI: %v", err)
	}
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read first: %v", err)
	}
	if err := WriteSectionI(path, attr); err != nil {
		t.Fatalf("second WriteSectionI: %v", err)
	}
	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read second: %v", err)
	}
	if string(first) != string(second) {
		t.Errorf("WriteSectionI not idempotent — second call changed output\n--- first ---\n%s\n--- second ---\n%s", first, second)
	}
}

func TestWriteSectionI_PreservesSiblingSections(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "progress.md")
	// §I sits between §E.4 and §F — replace must not bleed into neighbors.
	seed := strings.Join([]string{
		"# Progress",
		"",
		"## §E.2 Run-phase Evidence",
		"- run evidence line A",
		"",
		"## §E.4 Sync-phase Audit-Ready Signal",
		"sync_commit_sha: deadbeef",
		"",
		"## §I Token Accounting",
		"",
		"<!-- old placeholder -->",
		"",
		"## §F Phase 0.95 Mode Selection",
		"",
		"**Decision: sub-agent**",
		"",
	}, "\n")
	if err := os.WriteFile(path, []byte(seed), 0o644); err != nil {
		t.Fatalf("seed write: %v", err)
	}

	if err := WriteSectionI(path, sampleAttribution()); err != nil {
		t.Fatalf("WriteSectionI: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	s := string(got)

	// §E content preserved.
	if !strings.Contains(s, "- run evidence line A") {
		t.Error("§E.2 body content lost")
	}
	if !strings.Contains(s, "sync_commit_sha: deadbeef") {
		t.Error("§E.4 sync_commit_sha field lost")
	}
	// §F content preserved (the section AFTER §I).
	if !strings.Contains(s, "## §F Phase 0.95 Mode Selection") {
		t.Error("§F heading lost — replace bled past §I into following section")
	}
	if !strings.Contains(s, "**Decision: sub-agent**") {
		t.Error("§F body content lost")
	}
	// §I now carries measured values, not the old placeholder.
	if strings.Contains(s, "old placeholder") {
		t.Errorf("old placeholder survived; got:\n%s", s)
	}
	if !strings.Contains(s, "tokens_spent: 1860") {
		t.Error("§I measured value missing")
	}
}

func TestWriteSectionI_AbsentFileReturnsWrappedError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	missing := filepath.Join(dir, "nope.md")
	err := WriteSectionI(missing, sampleAttribution())
	if err == nil {
		t.Fatal("expected error for absent file, got nil")
	}
	if !strings.Contains(err.Error(), "nope.md") {
		t.Errorf("error should name the missing path; got %v", err)
	}
}
