package report

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderOutcome_AlreadyUpToDate(t *testing.T) {
	// AC-TUX3-012: OutcomeKind parameterized through single renderer
	// AC-TUX3-013: No backup path means no recovery command shown
	result := RenderOutcome(OutcomeAlreadyUpToDate, 0, "")

	assert.Contains(t, result, "Up to date")
	assert.NotContains(t, result, "Backup")
	assert.NotContains(t, result, "restore")
	assert.NotContains(t, result, "--restore")
}

// TestRenderReconciliation — SPEC-UPDATE-MIGRATION-001 (card t1547): the
// reconciliation outcome renderer. Zero total renders nothing; a populated
// run names every category count and lists each conflict, archived removal,
// and preserved path (deletions always visible, REQ-UPM-031); plain text
// only (REQ-UPM-032).
func TestRenderReconciliation(t *testing.T) {
	// Zero boundary: nothing reconciled → empty (the caller prints no rows).
	assert.Empty(t, RenderReconciliation(ReconciliationCounts{}, nil, nil, nil))

	counts := ReconciliationCounts{
		Refreshed:       3,
		Merged:          2,
		Conflicts:       1,
		Preserved:       4,
		ArchivedRemoved: 2,
	}
	result := RenderReconciliation(counts,
		[]string{".claude/rules/moai/policy.json (sidecar: .claude/rules/moai/policy.json.moai-new)"},
		[]string{".claude/rules/moai/local-note.md"},
		[]string{".claude/rules/moai/old-rule.md", ".claude/rules/moai/older-rule.md"})

	assert.Contains(t, result, "3 refreshed")
	assert.Contains(t, result, "2 merged")
	assert.Contains(t, result, "1 conflict(s)")
	assert.Contains(t, result, "4 preserved")
	assert.Contains(t, result, "2 archived-removed")
	assert.Contains(t, result, "conflict: .claude/rules/moai/policy.json")
	assert.Contains(t, result, "archived-removed: .claude/rules/moai/old-rule.md")
	assert.Contains(t, result, "archived-removed: .claude/rules/moai/older-rule.md")
	assert.Contains(t, result, "preserved: .claude/rules/moai/local-note.md")
}

func TestRenderOutcome_UpdatedFiles(t *testing.T) {
	// AC-TUX3-012: UpdatedFiles outcome through same renderer
	// AC-TUX3-013: With backup path shows recovery command
	result := RenderOutcome(OutcomeUpdatedFiles, 3, "/tmp/backup-20260714.tar.gz")

	assert.Contains(t, result, "3")
	assert.Contains(t, result, "file")
	assert.Contains(t, result, "/tmp/backup-20260714.tar.gz")
	assert.Contains(t, result, "restore")
	assert.Contains(t, result, "--restore")
}

func TestRenderOutcome_DryRun(t *testing.T) {
	// AC-TUX3-012: DryRun outcome through same renderer
	// AC-TUX3-013: With backup path shows recovery command
	result := RenderOutcome(OutcomeDryRun, 5, "/tmp/backup-dryrun.tar.gz")

	assert.Contains(t, result, "Dry run")
	assert.Contains(t, result, "5")
	assert.Contains(t, result, "file")
	assert.Contains(t, result, "/tmp/backup-dryrun.tar.gz")
	assert.Contains(t, result, "restore")
}

func TestRenderOutcome_NoBackupPath(t *testing.T) {
	// AC-TUX3-013: Empty backup path omits recovery command entirely
	testCases := []struct {
		name       string
		kind       OutcomeKind
		fileCount  int
		backupPath string
	}{
		{"AlreadyUpToDate no backup", OutcomeAlreadyUpToDate, 0, ""},
		{"UpdatedFiles no backup", OutcomeUpdatedFiles, 2, ""},
		{"DryRun no backup", OutcomeDryRun, 1, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := RenderOutcome(tc.kind, tc.fileCount, tc.backupPath)

			// Should NOT contain recovery command when backupPath is empty
			assert.NotContains(t, result, "restore")
			assert.NotContains(t, result, "--restore")
			assert.NotContains(t, result, "Backup")
		})
	}
}

func TestRenderOutcome_SingleRendererPath(t *testing.T) {
	// AC-TUX3-012: All outcomes route through the same RenderOutcome function
	// This test validates the contract - single entry point for all outcomes
	outcomes := []OutcomeKind{
		OutcomeAlreadyUpToDate,
		OutcomeUpdatedFiles,
		OutcomeDryRun,
	}

	for _, kind := range outcomes {
		// Just verify the function accepts all OutcomeKind values
		result := RenderOutcome(kind, 1, "/tmp/backup.tar.gz")
		assert.NotEmpty(t, result, "RenderOutcome should return non-empty for all outcome kinds")
	}
}

func TestOutcomeKindString(t *testing.T) {
	// Validate String() method for OutcomeKind enum
	tests := []struct {
		kind     OutcomeKind
		expected string
	}{
		{OutcomeAlreadyUpToDate, "AlreadyUpToDate"},
		{OutcomeUpdatedFiles, "UpdatedFiles"},
		{OutcomeDryRun, "DryRun"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.kind.String())
		})
	}
}
