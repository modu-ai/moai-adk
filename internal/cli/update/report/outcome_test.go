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
