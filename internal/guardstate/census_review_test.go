package guardstate

import (
	"os"
	"path/filepath"
	"testing"
)

// AC-CCR-018 (REQ-CCR-017, SPEC-CI-CODEX-REVIEW-001) — the guard-liveness manifest lists exactly the
// workflows on disk, so after the three review workflows are retired none keeps an entry. The set equality
// in both directions is the existing TestCensus_SetDifferenceEmptyBothDirections; these rows name the three
// retired workflows so that a half-done retirement fails with a message that says which one.
//
// Mutant: leaving any manifest entry in place (or any workflow on disk) turns the matching row red.
func TestCensusOmitsRetiredReviewWorkflows(t *testing.T) {
	m, root := loadRepoManifest(t)
	if len(m.Entries) == 0 {
		t.Fatal("the manifest is empty; an absence check over it would pass vacuously")
	}
	for _, tc := range []struct{ name, locator string }{
		{"claude_entry_absent", ".github/workflows/claude.yml"},
		{"claude_code_review_entry_absent", ".github/workflows/claude-code-review.yml"},
		{"review_quality_gate_entry_absent", ".github/workflows/review-quality-gate.yml"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, e := range m.Entries {
				if e.Locator == tc.locator {
					t.Fatalf("the manifest still lists the retired workflow %s", tc.locator)
				}
			}
			if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(tc.locator))); err == nil {
				t.Fatalf("the retired workflow %s is still on disk", tc.locator)
			}
		})
	}
}
