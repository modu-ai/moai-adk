// progress_section_populated_test.go — card t996.
//
// The §E.4 leg of checkV3R6Drift asked only whether the literal token "§E.4"
// appeared anywhere in progress.md (hasProgressMarker == strings.Contains), so
// a section carrying nothing but the plan-phase placeholder `_<pending
// sync-phase>_` satisfied "sync marker present". Paired with a sync_commit_sha
// that lives in a DIFFERENT section, that reports SyncStatusDrift on a SPEC
// whose sync phase never ran.
//
// The assertions below are a PAIR on purpose. A single case is not enough:
// a predicate mutated to constant-false passes the placeholder case alone, and
// a predicate mutated to constant-true passes the populated case alone. Only
// both together pin the behaviour.
package spec

import (
	"path/filepath"
	"testing"
)

// e4PlaceholderProgress is the plan-phase scaffold shape: §E.4 exists as a
// heading whose whole body is the emphasis-wrapped pending note, while a
// sync_commit_sha field sits in a different section. Sync has NOT happened.
const e4PlaceholderProgress = `# progress.md

## §E.2 Run-phase Evidence

| AC | Status |
|----|--------|
| AC-001 | PASS |

## §E.3 Run-phase Audit-Ready Signal

sync_commit_sha: abc1234def

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
`

// e4PopulatedProgress is the post-sync shape: §E.4 carries the audit-ready
// signal itself. Sync HAS happened.
const e4PopulatedProgress = `# progress.md

## §E.2 Run-phase Evidence

| AC | Status |
|----|--------|
| AC-001 | PASS |

## §E.4 Sync-phase Audit-Ready Signal

sync_complete_at: 2026-09-20
sync_commit_sha: abc1234def
`

// e4BoldFieldProgress is the other populated spelling the corpus actually
// uses — a bold-wrapped markdown list field rather than a bare YAML line.
// It is populated and MUST be treated as such; a "body contains a key: value
// line" rule would misread it as pending.
const e4BoldFieldProgress = `# progress.md

## §E.2 Run-phase Evidence

| AC | Status |
|----|--------|
| AC-001 | PASS |

## §E.4 Sync-phase Audit-Ready Signal

- **sync_complete_at**: 2026-09-20 (KST)
- **sync_commit_sha**: ` + "`abc1234def`" + `
`

// TestPopulatedProgressSectionSeparatesPlaceholderFromEvidence pins the pair.
func TestPopulatedProgressSectionSeparatesPlaceholderFromEvidence(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    bool
	}{
		{"placeholder-only body is not populated", e4PlaceholderProgress, false},
		{"yaml field body is populated", e4PopulatedProgress, true},
		{"bold markdown field body is populated", e4BoldFieldProgress, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := hasPopulatedProgressSection(tc.content, "§E.4"); got != tc.want {
				t.Errorf("hasPopulatedProgressSection(%s, §E.4) = %v, want %v",
					tc.name, got, tc.want)
			}
		})
	}
}

// TestPopulatedProgressSectionAbsentSection keeps the absent case distinct from
// the placeholder case: a §E.4 that does not exist at all is also not
// populated, but for a different reason, and the predicate must not panic.
func TestPopulatedProgressSectionAbsentSection(t *testing.T) {
	const noE4 = `# progress.md

## §E.2 Run-phase Evidence

- run evidence line
`
	if hasPopulatedProgressSection(noE4, "§E.4") {
		t.Error("a progress.md with no §E.4 section must not report it as populated")
	}
	if !hasProgressMarker(e4PlaceholderProgress, "§E.4") {
		t.Error("premise broken: the placeholder fixture must still satisfy the " +
			"EXISTENCE predicate — otherwise this test proves nothing about the " +
			"difference between existence and population")
	}
}

// TestSyncStatusDriftIgnoresPlaceholderSyncSection is the end-to-end pair at
// the finding level: the same spec.md status, the same extractable
// sync_commit_sha, differing only in whether §E.4 carries evidence.
func TestSyncStatusDriftIgnoresPlaceholderSyncSection(t *testing.T) {
	drift := func(progress string) *DriftFinding {
		t.Helper()
		base := t.TempDir()
		const id = "SPEC-T996-PAIR-001"
		writeSpecFixture(t, base, id, "in-progress", "2026-09-20", progress)
		specDir := filepath.Join(base, ".moai", "specs", id)
		return checkV3R6Drift(specDir, id, EraSignals{ProgressMDContent: progress})
	}

	if f := drift(e4PlaceholderProgress); f != nil {
		t.Errorf("placeholder §E.4 must not report sync completion; got %s (%s)",
			f.FindingType, f.Severity)
	}
	f := drift(e4PopulatedProgress)
	if f == nil {
		t.Fatal("populated §E.4 with a non-completed status must still report " +
			"SyncStatusDrift — the repair must not silence the real finding")
	}
	if f.FindingType != FindingSyncStatusDrift || f.Severity != "MUST-FIX" {
		t.Errorf("populated §E.4: got %s/%s, want %s/MUST-FIX",
			f.FindingType, f.Severity, FindingSyncStatusDrift)
	}
}
