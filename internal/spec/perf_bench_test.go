package spec

import (
	"fmt"
	"strings"
	"testing"
)

// BenchmarkExtractProgressField measures the hot-path regex extraction cost
// for the era classification path (REQ-PERF-004-A). Before the fix, each call
// compiled 2 regexes via regexp.MustCompile; after the fix, package-level
// pre-compiled patterns are reused.
func BenchmarkExtractProgressField(b *testing.B) {
	// Synthetic progress.md content with §E markers + commit_sha fields
	content := strings.Repeat("# Some heading\n\nBody text.\n\n", 50) +
		"## §E.4 Sync-phase Audit-Ready Signal\n" +
		"sync_commit_sha: \"a1b2c3d4e5f6789\"\n" +
		"mx_commit_sha: \"f9e8d7c6b5a4321\"\n"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = extractProgressField(content, "sync_commit_sha")
		_ = extractProgressField(content, "mx_commit_sha")
	}
}

// BenchmarkParseStatusFromContent measures the status parsing hot path
// (REQ-PERF-004-A). Before the fix, parseStatusFromTable and parseStatusFromMarkdownList
// compiled regexes on every call; after the fix, package-level patterns are reused.
func BenchmarkParseStatusFromContent(b *testing.B) {
	content := "---\nid: SPEC-TEST-001\nstatus: in-progress\nupdated: 2026-07-08\n---\n\n# Body"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseStatusFromContent(content)
	}
}

// BenchmarkParseAcceptanceCriteria measures the SPEC AC parsing hot path.
// parseSingleACLine compiled 5 fixed regexes per AC line and
// ExtractRequirementMappings compiled 2 more per call, so a SPEC with N
// acceptance lines paid 7N regexp compilations.
func BenchmarkParseAcceptanceCriteria(b *testing.B) {
	var sb strings.Builder
	sb.WriteString("# SPEC-BENCH-001\n\n## Acceptance Criteria\n\n")
	for i := 1; i <= 30; i++ {
		fmt.Fprintf(&sb,
			"- AC-BENCH-001-%02d: Given a parsed SPEC document, When the auditor reads it, "+
				"Then the criteria resolve (maps REQ-BENCH-001-%03d, REQ-BENCH-001-%03d)\n",
			i, i, i+100)
	}
	markdown := sb.String()

	// Guard against a vacuous benchmark: a fixture the parser rejects would
	// measure the early-return path, not the AC parsing hot path.
	if got, _ := ParseAcceptanceCriteria(markdown, true); len(got) != 30 {
		b.Fatalf("fixture parsed %d acceptance criteria, want 30", len(got))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseAcceptanceCriteria(markdown, true)
	}
}

// BenchmarkExtractRequirementMappings isolates the 2-pass REQ extractor, which
// runs once per AC line from parseSingleACLine and again from sibling-table
// coverage collection.
func BenchmarkExtractRequirementMappings(b *testing.B) {
	text := "Given a mapping line, Then it resolves " +
		"(maps REQ-BENCH-001-001, REQ-BENCH-001-002, REQ-BENCH-001-003)"

	if got := ExtractRequirementMappings(text); len(got) != 3 {
		b.Fatalf("fixture extracted %d REQ ids, want 3", len(got))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ExtractRequirementMappings(text)
	}
}
