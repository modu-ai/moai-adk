package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// archive_test.go — RED-phase tests for the SPEC auto-archive capability
// (SPEC-SESSIONSTART-PERF-001 M2, REQ-SSP-007 .. REQ-SSP-013).
//
// The central risk this file guards is the grandfather false-positive documented in
// .claude/rules/moai/core/verification-claim-integrity.md §5: an inference drawn from
// frontmatter text alone once nearly batch-touched 29 grandfather-protected SPECs.
// The eligibility predicate here is therefore deliberately status+date ONLY — era
// classification is REPORTED but is never a gate in either direction.

// specFixture describes one synthetic SPEC directory.
type specFixture struct {
	id string
	// status is the frontmatter status value.
	status string
	// updated is the frontmatter `updated:` date (YYYY-MM-DD). Empty means omitted.
	updated string
	// grandfather, when true, omits progress.md so ClassifyEra resolves to V2.x
	// (era-final / grandfather-protected) via heuristic H-1. When false, a V3R6
	// progress.md is written so the SPEC is modern-era (NOT grandfather-protected).
	grandfather bool
}

// writeSpecFixtures materialises the fixtures under <baseDir>/.moai/specs/.
func writeSpecFixtures(t *testing.T, baseDir string, fixtures ...specFixture) {
	t.Helper()

	for _, f := range fixtures {
		dir := filepath.Join(baseDir, ".moai", "specs", f.id)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}

		updated := f.updated
		if updated == "" {
			updated = "2020-01-01"
		}

		specMD := fmt.Sprintf(`---
id: %s
title: "Fixture"
version: "0.1.0"
status: %s
created: 2020-01-01
updated: %s
author: test
priority: P1
phase: "v3.0.0"
module: "internal/spec"
lifecycle: spec-anchored
tags: "fixture"
---

# %s
`, f.id, f.status, updated, f.id)

		if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(specMD), 0o644); err != nil {
			t.Fatalf("write spec.md: %v", err)
		}

		if f.grandfather {
			// No progress.md → ClassifyEra heuristic H-1 → V2.x → EraFinal() == true.
			continue
		}

		// A V3R6-shaped progress.md (H-4: §E.2 + §E.4 + sync_commit_sha) → NOT era-final.
		progressMD := `## §E.2 Run-phase Evidence

## §E.4 Sync-phase Audit-Ready Signal
sync_commit_sha: "abc123def456"
`
		if err := os.WriteFile(filepath.Join(dir, "progress.md"), []byte(progressMD), 0o644); err != nil {
			t.Fatalf("write progress.md: %v", err)
		}
	}
}

// staticActivity builds an archiveDeps whose last-activity map is fixed, so the
// eligibility logic is tested without a git repository.
func staticActivity(activity map[string]time.Time) archiveDeps {
	return archiveDeps{
		lastActivity: func(string) map[string]time.Time { return activity },
		move:         osRenameMove,
	}
}

// noGitActivity is the deps variant where git yields nothing, forcing the
// frontmatter `updated:` fallback.
func noGitActivity() archiveDeps {
	return staticActivity(map[string]time.Time{})
}

func candidateIDs(plan *ArchivePlan) []string {
	ids := make([]string, 0, len(plan.Candidates))
	for _, c := range plan.Candidates {
		ids = append(ids, c.SPECID)
	}
	return ids
}

func hasCandidate(plan *ArchivePlan, specID string) bool {
	for _, c := range plan.Candidates {
		if c.SPECID == specID {
			return true
		}
	}
	return false
}

func TestIsArchiveTerminalStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status string
		want   bool
	}{
		// The 4 terminal statuses (spec.md REQ-SSP-007 / design.md §M2.1).
		{"completed", true},
		{"superseded", true},
		{"archived", true},
		{"rejected", true},
		// Non-terminal — never archive-eligible regardless of age.
		{"draft", false},
		{"planned", false},
		{"in-progress", false},
		{"implemented", false},
		{"", false},
		{"nonsense", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			t.Parallel()
			if got := IsArchiveTerminalStatus(tt.status); got != tt.want {
				t.Errorf("IsArchiveTerminalStatus(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

// TestIsArchiveTerminalStatus_DivergesFromDriftTerminal pins the deliberate
// difference between the two "terminal" notions in this package. drift.go's
// isTerminalStatus deliberately EXCLUDES "completed" (git can positively infer a
// close, so frontmatter is not authoritative there); the archive predicate
// INCLUDES it. Collapsing the two would silently stop archiving every completed
// SPEC — the single largest slice of the corpus.
func TestIsArchiveTerminalStatus_DivergesFromDriftTerminal(t *testing.T) {
	t.Parallel()

	if isTerminalStatus("completed") {
		t.Fatal("drift isTerminalStatus should NOT treat completed as terminal")
	}
	if !IsArchiveTerminalStatus("completed") {
		t.Fatal("archive terminal set MUST include completed")
	}
}

func TestPlanArchive_TerminalPastGraceIsEligible(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	writeSpecFixtures(t, base, specFixture{id: "SPEC-OLD-001", status: "completed"})

	now := time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)
	deps := staticActivity(map[string]time.Time{
		"SPEC-OLD-001": now.AddDate(0, 0, -120), // 120d old, grace is 90d
	})

	plan, err := planArchive(base, ArchiveOptions{GraceDays: 90, Now: now}, deps)
	if err != nil {
		t.Fatalf("planArchive: %v", err)
	}

	if !hasCandidate(plan, "SPEC-OLD-001") {
		t.Fatalf("expected SPEC-OLD-001 eligible, got %v", candidateIDs(plan))
	}
}

func TestPlanArchive_TerminalWithinGraceIsNotEligible(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	writeSpecFixtures(t, base, specFixture{id: "SPEC-FRESH-001", status: "completed"})

	now := time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)
	deps := staticActivity(map[string]time.Time{
		"SPEC-FRESH-001": now.AddDate(0, 0, -30), // well inside the 90d window
	})

	plan, err := planArchive(base, ArchiveOptions{GraceDays: 90, Now: now}, deps)
	if err != nil {
		t.Fatalf("planArchive: %v", err)
	}

	if hasCandidate(plan, "SPEC-FRESH-001") {
		t.Fatal("a SPEC inside the grace window must NOT be archive-eligible")
	}
}

func TestPlanArchive_NonTerminalPastGraceIsNotEligible(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	writeSpecFixtures(t,
		base,
		specFixture{id: "SPEC-DRAFT-001", status: "draft"},
		specFixture{id: "SPEC-INPROG-001", status: "in-progress"},
		specFixture{id: "SPEC-IMPL-001", status: "implemented"},
	)

	now := time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)
	ancient := now.AddDate(-3, 0, 0)
	deps := staticActivity(map[string]time.Time{
		"SPEC-DRAFT-001":  ancient,
		"SPEC-INPROG-001": ancient,
		"SPEC-IMPL-001":   ancient,
	})

	plan, err := planArchive(base, ArchiveOptions{GraceDays: 90, Now: now}, deps)
	if err != nil {
		t.Fatalf("planArchive: %v", err)
	}

	if len(plan.Candidates) != 0 {
		t.Fatalf("age alone must never make a non-terminal SPEC eligible; got %v", candidateIDs(plan))
	}
}

// TestPlanArchive_GraceBoundary pins the boundary as STRICTLY before the cutoff:
// a SPEC whose last activity lands exactly on the cutoff instant stays.
func TestPlanArchive_GraceBoundary(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	writeSpecFixtures(t,
		base,
		specFixture{id: "SPEC-EXACT-001", status: "completed"},
		specFixture{id: "SPEC-ONEDAYOLDER-001", status: "completed"},
	)

	now := time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)
	cutoff := now.AddDate(0, 0, -90)

	deps := staticActivity(map[string]time.Time{
		"SPEC-EXACT-001":       cutoff,                     // exactly on the cutoff
		"SPEC-ONEDAYOLDER-001": cutoff.Add(-1 * time.Hour), // one hour past it
	})

	plan, err := planArchive(base, ArchiveOptions{GraceDays: 90, Now: now}, deps)
	if err != nil {
		t.Fatalf("planArchive: %v", err)
	}

	if hasCandidate(plan, "SPEC-EXACT-001") {
		t.Error("a SPEC exactly on the cutoff must NOT be eligible (strictly-before boundary)")
	}
	if !hasCandidate(plan, "SPEC-ONEDAYOLDER-001") {
		t.Error("a SPEC past the cutoff must be eligible")
	}
}

// TestPlanArchive_GrandfatherIsNotAGate is the [HARD] REQ-SSP-010 guard and the
// direct codification of the verification-claim-integrity.md §5 incident.
//
// Grandfather (era-final) status is ORTHOGONAL to archive eligibility:
//   - it must NOT force archival (an era-final SPEC that is draft, or still inside
//     the grace window, stays put)
//   - it must NOT forbid archival (an era-final SPEC that independently satisfies
//     terminal + grace IS eligible)
func TestPlanArchive_GrandfatherIsNotAGate(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	writeSpecFixtures(t,
		base,
		// era-final + terminal + past grace  → ELIGIBLE (grandfather does not forbid)
		specFixture{id: "SPEC-GFELIGIBLE-001", status: "completed", grandfather: true},
		// era-final + NOT terminal + past grace → NOT eligible (grandfather does not force)
		specFixture{id: "SPEC-GFDRAFT-001", status: "draft", grandfather: true},
		// era-final + terminal + INSIDE grace  → NOT eligible (grandfather does not force)
		specFixture{id: "SPEC-GFFRESH-001", status: "completed", grandfather: true},
		// modern era + terminal + past grace   → ELIGIBLE (control)
		specFixture{id: "SPEC-MODERN-001", status: "completed", grandfather: false},
	)

	now := time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)
	old := now.AddDate(0, 0, -200)
	fresh := now.AddDate(0, 0, -10)

	deps := staticActivity(map[string]time.Time{
		"SPEC-GFELIGIBLE-001": old,
		"SPEC-GFDRAFT-001":    old,
		"SPEC-GFFRESH-001":    fresh,
		"SPEC-MODERN-001":     old,
	})

	plan, err := planArchive(base, ArchiveOptions{GraceDays: 90, Now: now}, deps)
	if err != nil {
		t.Fatalf("planArchive: %v", err)
	}

	if !hasCandidate(plan, "SPEC-GFELIGIBLE-001") {
		t.Error("grandfather status must NOT forbid archival when terminal+grace hold independently")
	}
	if hasCandidate(plan, "SPEC-GFDRAFT-001") {
		t.Error("grandfather status must NOT force archival of a non-terminal SPEC")
	}
	if hasCandidate(plan, "SPEC-GFFRESH-001") {
		t.Error("grandfather status must NOT force archival of a SPEC inside the grace window")
	}
	if !hasCandidate(plan, "SPEC-MODERN-001") {
		t.Error("modern-era terminal SPEC past grace must be eligible")
	}
}

// TestPlanArchive_ReportsEraFinal verifies era classification is still SURFACED on
// each candidate (for operator review) even though it does not gate the decision.
func TestPlanArchive_ReportsEraFinal(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	writeSpecFixtures(t,
		base,
		specFixture{id: "SPEC-GF-001", status: "completed", grandfather: true},
		specFixture{id: "SPEC-MOD-001", status: "completed", grandfather: false},
	)

	now := time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)
	old := now.AddDate(0, 0, -200)
	deps := staticActivity(map[string]time.Time{"SPEC-GF-001": old, "SPEC-MOD-001": old})

	plan, err := planArchive(base, ArchiveOptions{GraceDays: 90, Now: now}, deps)
	if err != nil {
		t.Fatalf("planArchive: %v", err)
	}

	got := map[string]bool{}
	for _, c := range plan.Candidates {
		got[c.SPECID] = c.EraFinal
	}

	if !got["SPEC-GF-001"] {
		t.Error("SPEC-GF-001 (no progress.md → V2.x) should report EraFinal=true")
	}
	if got["SPEC-MOD-001"] {
		t.Error("SPEC-MOD-001 (V3R6 progress.md) should report EraFinal=false")
	}
}

func TestPlanArchive_DestinationIsYearPartitioned(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	writeSpecFixtures(t, base, specFixture{id: "SPEC-DEST-001", status: "completed"})

	now := time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)
	deps := staticActivity(map[string]time.Time{
		"SPEC-DEST-001": time.Date(2024, 3, 9, 0, 0, 0, 0, time.UTC),
	})

	plan, err := planArchive(base, ArchiveOptions{GraceDays: 90, Now: now}, deps)
	if err != nil {
		t.Fatalf("planArchive: %v", err)
	}
	if len(plan.Candidates) != 1 {
		t.Fatalf("want 1 candidate, got %v", candidateIDs(plan))
	}

	c := plan.Candidates[0]
	wantSrc := filepath.Join(".moai", "specs", "SPEC-DEST-001")
	wantDst := filepath.Join(".moai", "archive", "specs", "2024", "SPEC-DEST-001")

	if c.SourceDir != wantSrc {
		t.Errorf("SourceDir = %q, want %q", c.SourceDir, wantSrc)
	}
	if c.DestDir != wantDst {
		t.Errorf("DestDir = %q, want %q", c.DestDir, wantDst)
	}
}

// TestPlanArchive_NeverMoves is the dry-run safety invariant: planning is
// observation-only. Nothing under .moai/ may change.
func TestPlanArchive_NeverMoves(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	writeSpecFixtures(t, base, specFixture{id: "SPEC-STAY-001", status: "completed"})

	now := time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)
	deps := staticActivity(map[string]time.Time{"SPEC-STAY-001": now.AddDate(0, 0, -300)})

	plan, err := planArchive(base, ArchiveOptions{GraceDays: 90, Now: now}, deps)
	if err != nil {
		t.Fatalf("planArchive: %v", err)
	}
	if len(plan.Candidates) != 1 {
		t.Fatalf("fixture should be eligible; got %v", candidateIDs(plan))
	}

	// Source must still be there, destination must not exist.
	if _, err := os.Stat(filepath.Join(base, ".moai", "specs", "SPEC-STAY-001", "spec.md")); err != nil {
		t.Errorf("planArchive moved the source: %v", err)
	}
	if _, err := os.Stat(filepath.Join(base, ".moai", "archive")); !os.IsNotExist(err) {
		t.Error("planArchive created the archive tree; planning must be observation-only")
	}
}

func TestExecuteArchive_MovesOnlyEligible(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	writeSpecFixtures(t,
		base,
		specFixture{id: "SPEC-GO-001", status: "completed"},   // eligible
		specFixture{id: "SPEC-KEEP-001", status: "draft"},     // non-terminal
		specFixture{id: "SPEC-YOUNG-001", status: "rejected"}, // inside grace
	)

	now := time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)
	deps := staticActivity(map[string]time.Time{
		"SPEC-GO-001":    time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
		"SPEC-KEEP-001":  time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
		"SPEC-YOUNG-001": now.AddDate(0, 0, -5),
	})

	plan, err := executeArchive(base, ArchiveOptions{GraceDays: 90, Now: now}, deps)
	if err != nil {
		t.Fatalf("executeArchive: %v", err)
	}
	if len(plan.Candidates) != 1 || plan.Candidates[0].SPECID != "SPEC-GO-001" {
		t.Fatalf("want only SPEC-GO-001 moved, got %v", candidateIDs(plan))
	}

	// Moved: gone from specs, present under archive/<year>/.
	if _, err := os.Stat(filepath.Join(base, ".moai", "specs", "SPEC-GO-001")); !os.IsNotExist(err) {
		t.Error("SPEC-GO-001 should no longer be under .moai/specs")
	}
	moved := filepath.Join(base, ".moai", "archive", "specs", "2023", "SPEC-GO-001", "spec.md")
	if _, err := os.Stat(moved); err != nil {
		t.Errorf("SPEC-GO-001 not found at archive destination: %v", err)
	}

	// Untouched.
	for _, id := range []string{"SPEC-KEEP-001", "SPEC-YOUNG-001"} {
		if _, err := os.Stat(filepath.Join(base, ".moai", "specs", id, "spec.md")); err != nil {
			t.Errorf("%s must not be moved: %v", id, err)
		}
	}
}

// TestExecuteArchive_ContentSurvives is the grep-discoverability guard
// (REQ-SSP-007 / AC-SSP-023): archiving is a MOVE, never a delete.
func TestExecuteArchive_ContentSurvives(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	writeSpecFixtures(t, base, specFixture{id: "SPEC-KEEPBODY-001", status: "superseded"})

	now := time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)
	deps := staticActivity(map[string]time.Time{
		"SPEC-KEEPBODY-001": time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC),
	})

	if _, err := executeArchive(base, ArchiveOptions{GraceDays: 90, Now: now}, deps); err != nil {
		t.Fatalf("executeArchive: %v", err)
	}

	body, err := os.ReadFile(filepath.Join(base, ".moai", "archive", "specs", "2022", "SPEC-KEEPBODY-001", "spec.md"))
	if err != nil {
		t.Fatalf("archived spec.md unreadable: %v", err)
	}
	if !containsAll(string(body), "SPEC-KEEPBODY-001", "status: superseded") {
		t.Error("archived SPEC content must survive the move intact (grep-discoverable)")
	}
}

func containsAll(haystack string, needles ...string) bool {
	for _, n := range needles {
		found := false
		for i := 0; i+len(n) <= len(haystack); i++ {
			if haystack[i:i+len(n)] == n {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// TestPlanArchive_FrontmatterFallback covers the no-git-history path: the
// `updated:` frontmatter date stands in for the git last-activity date.
func TestPlanArchive_FrontmatterFallback(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	writeSpecFixtures(t,
		base,
		specFixture{id: "SPEC-FMOLD-001", status: "completed", updated: "2021-06-01"},
		specFixture{id: "SPEC-FMNEW-001", status: "completed", updated: "2026-07-01"},
	)

	now := time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)

	plan, err := planArchive(base, ArchiveOptions{GraceDays: 90, Now: now}, noGitActivity())
	if err != nil {
		t.Fatalf("planArchive: %v", err)
	}

	if !hasCandidate(plan, "SPEC-FMOLD-001") {
		t.Error("old frontmatter updated: date should make the SPEC eligible when git yields nothing")
	}
	if hasCandidate(plan, "SPEC-FMNEW-001") {
		t.Error("recent frontmatter updated: date must keep the SPEC inside the grace window")
	}

	for _, c := range plan.Candidates {
		if c.ActivitySource != ActivitySourceFrontmatter {
			t.Errorf("%s ActivitySource = %q, want %q", c.SPECID, c.ActivitySource, ActivitySourceFrontmatter)
		}
	}
}

// TestPlanArchive_GitActivityWinsOverFrontmatter: a stale `updated:` field must not
// archive a SPEC that git shows was touched recently. Git is the authority.
func TestPlanArchive_GitActivityWinsOverFrontmatter(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	writeSpecFixtures(t, base, specFixture{id: "SPEC-STALEFM-001", status: "completed", updated: "2019-01-01"})

	now := time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)
	deps := staticActivity(map[string]time.Time{
		"SPEC-STALEFM-001": now.AddDate(0, 0, -3), // git says: touched 3 days ago
	})

	plan, err := planArchive(base, ArchiveOptions{GraceDays: 90, Now: now}, deps)
	if err != nil {
		t.Fatalf("planArchive: %v", err)
	}

	if hasCandidate(plan, "SPEC-STALEFM-001") {
		t.Fatal("git last-activity must override a stale frontmatter updated: date")
	}
}

// TestPlanArchive_UnparseableIsSkipped: a directory without a readable status is
// never eligible (fail-safe — we cannot prove it is terminal).
func TestPlanArchive_UnparseableIsSkipped(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	junk := filepath.Join(base, ".moai", "specs", "not-a-spec")
	if err := os.MkdirAll(junk, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	now := time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)
	plan, err := planArchive(base, ArchiveOptions{GraceDays: 90, Now: now}, noGitActivity())
	if err != nil {
		t.Fatalf("planArchive: %v", err)
	}
	if len(plan.Candidates) != 0 {
		t.Fatalf("unparseable dir must never be eligible, got %v", candidateIDs(plan))
	}
}

func TestPlanArchive_NoSpecsDirIsEmptyPlan(t *testing.T) {
	t.Parallel()

	plan, err := planArchive(t.TempDir(), ArchiveOptions{GraceDays: 90}, noGitActivity())
	if err != nil {
		t.Fatalf("missing .moai/specs must not error: %v", err)
	}
	if len(plan.Candidates) != 0 || plan.Scanned != 0 {
		t.Fatalf("want empty plan, got %d candidates / %d scanned", len(plan.Candidates), plan.Scanned)
	}
}

// TestPlanArchive_ZeroGraceDaysUsesDefault: an unset GraceDays resolves to the
// configured default rather than degenerating into "archive everything terminal".
func TestPlanArchive_ZeroGraceDaysUsesDefault(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	writeSpecFixtures(t, base, specFixture{id: "SPEC-RECENT-001", status: "completed"})

	now := time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)
	deps := staticActivity(map[string]time.Time{
		"SPEC-RECENT-001": now.AddDate(0, 0, -30), // inside the 90d default
	})

	plan, err := planArchive(base, ArchiveOptions{GraceDays: 0, Now: now}, deps)
	if err != nil {
		t.Fatalf("planArchive: %v", err)
	}

	if plan.GraceDays != config.DefaultArchiveGraceDays {
		t.Errorf("GraceDays = %d, want the default %d", plan.GraceDays, config.DefaultArchiveGraceDays)
	}
	if len(plan.Candidates) != 0 {
		t.Fatalf("GraceDays=0 must not mean 'no grace'; got %v", candidateIDs(plan))
	}
}

func TestPlanArchive_ScannedCountsAllSpecs(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	writeSpecFixtures(t,
		base,
		specFixture{id: "SPEC-A-001", status: "completed"},
		specFixture{id: "SPEC-B-001", status: "draft"},
		specFixture{id: "SPEC-C-001", status: "in-progress"},
	)

	now := time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)
	plan, err := planArchive(base, ArchiveOptions{GraceDays: 90, Now: now}, noGitActivity())
	if err != nil {
		t.Fatalf("planArchive: %v", err)
	}
	if plan.Scanned != 3 {
		t.Errorf("Scanned = %d, want 3", plan.Scanned)
	}
}

// TestParseGitActivity covers the single-pass `git log --name-only` parser that
// keeps the archive scan at O(1) subprocesses — the same asymptotic guarantee M1
// established for the drift path.
func TestParseGitActivity(t *testing.T) {
	t.Parallel()

	// Newest-first, as git emits. SPEC-DUP-001 appears twice; the NEWEST wins.
	output := "\x1e2026-07-01T10:00:00+09:00\n" +
		".moai/specs/SPEC-DUP-001/spec.md\n" +
		".moai/specs/SPEC-NEW-001/progress.md\n" +
		"\x1e2024-02-03T08:30:00+09:00\n" +
		".moai/specs/SPEC-DUP-001/plan.md\n" +
		".moai/specs/SPEC-OLD-001/spec.md\n" +
		"README.md\n"

	got := parseGitActivity(output)

	if len(got) != 3 {
		t.Fatalf("want 3 SPECs, got %d: %v", len(got), got)
	}

	// Newest entry wins for a SPEC touched by several commits.
	wantDup := time.Date(2026, 7, 1, 10, 0, 0, 0, time.FixedZone("", 9*3600))
	if !got["SPEC-DUP-001"].Equal(wantDup) {
		t.Errorf("SPEC-DUP-001 = %v, want the NEWEST touch %v", got["SPEC-DUP-001"], wantDup)
	}

	wantOld := time.Date(2024, 2, 3, 8, 30, 0, 0, time.FixedZone("", 9*3600))
	if !got["SPEC-OLD-001"].Equal(wantOld) {
		t.Errorf("SPEC-OLD-001 = %v, want %v", got["SPEC-OLD-001"], wantOld)
	}

	if _, ok := got["README.md"]; ok {
		t.Error("paths outside .moai/specs must not enter the activity map")
	}
}

func TestParseGitActivity_EmptyOutput(t *testing.T) {
	t.Parallel()

	if got := parseGitActivity(""); len(got) != 0 {
		t.Fatalf("empty git output must yield an empty map, got %v", got)
	}
}
