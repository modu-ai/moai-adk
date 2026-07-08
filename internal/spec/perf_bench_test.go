package spec

import (
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
