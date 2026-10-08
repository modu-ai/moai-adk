package report

import (
	"fmt"
)

// OutcomeKind represents the type of update operation result
type OutcomeKind int

const (
	// OutcomeAlreadyUpToDate indicates no updates were needed
	OutcomeAlreadyUpToDate OutcomeKind = iota
	// OutcomeUpdatedFiles indicates files were updated
	OutcomeUpdatedFiles
	// OutcomeDryRun indicates a dry run was performed
	OutcomeDryRun
)

// String returns the string representation of OutcomeKind
func (k OutcomeKind) String() string {
	switch k {
	case OutcomeAlreadyUpToDate:
		return "AlreadyUpToDate"
	case OutcomeUpdatedFiles:
		return "UpdatedFiles"
	case OutcomeDryRun:
		return "DryRun"
	default:
		return "Unknown"
	}
}

// RenderOutcome renders an outcome card as plain text
// AC-TUX3-012: Single renderer for all 3 outcome types (parameterized by kind)
// AC-TUX3-013: Shows backup path and recovery command when backupPath is non-empty
func RenderOutcome(kind OutcomeKind, fileCount int, backupPath string) string {
	var result string

	switch kind {
	case OutcomeAlreadyUpToDate:
		result = "✓ Up to date · Skipping sync"
		if backupPath != "" {
			result += "\n\nBackup: " + backupPath
			result += "\nRecover: moai update --restore " + backupPath
		}

	case OutcomeUpdatedFiles:
		result = "✓ Updated "
		if fileCount == 1 {
			result += "1 file"
		} else {
			result += fmt.Sprintf("%d files", fileCount)
		}
		if backupPath != "" {
			result += "\n\nBackup: " + backupPath
			result += "\nRecover: moai update --restore " + backupPath
		}

	case OutcomeDryRun:
		result = "Dry run: "
		if fileCount == 1 {
			result += "1 file"
		} else {
			result += fmt.Sprintf("%d files", fileCount)
		}
		result += " would be updated"
		if backupPath != "" {
			result += "\n\nBackup: " + backupPath
			result += "\nRecover: moai update --restore " + backupPath
		}
	}

	return result
}

// ReconciliationCounts carries the five reconciliation outcome totals
// (SPEC-UPDATE-MIGRATION-001 REQ-UPM-030). Plain numbers — the t1527 UX
// overhaul owns any richer presentation (REQ-UPM-032).
type ReconciliationCounts struct {
	// Refreshed counts template-owned files the deploy rewrote in place.
	Refreshed int
	// Merged counts user-modified files written back as a clean 3-way merge.
	Merged int
	// Conflicts counts conflict dispositions (preserved file + .moai-new
	// sidecar), collisions included.
	Conflicts int
	// Preserved counts user-owned files left byte-for-byte untouched.
	Preserved int
	// ArchivedRemoved counts stale files copied to the migration archive and
	// then removed from place.
	ArchivedRemoved int
}

// RenderReconciliation renders the reconciliation outcome as plain counts and
// per-path lists (SPEC-UPDATE-MIGRATION-001 REQ-UPM-030/031/032): every
// category is named with its count, deletions are always visible (each
// archived removal listed by path), conflicts name their sidecar paths, and
// preserved files are listed so the summary can never claim a deletion-free
// run over what it actually kept. No banners, no severity prefixes — the
// existing plain-text outcome structure only.
func RenderReconciliation(counts ReconciliationCounts, conflicted, preserved, archivedRemoved []string) string {
	total := counts.Refreshed + counts.Merged + counts.Conflicts + counts.Preserved + counts.ArchivedRemoved
	if total == 0 {
		return ""
	}
	result := fmt.Sprintf("Reconciliation: %d refreshed, %d merged, %d conflict(s), %d preserved, %d archived-removed",
		counts.Refreshed, counts.Merged, counts.Conflicts, counts.Preserved, counts.ArchivedRemoved)
	for _, p := range conflicted {
		result += "\n  conflict: " + p
	}
	for _, p := range archivedRemoved {
		result += "\n  archived-removed: " + p
	}
	for _, p := range preserved {
		result += "\n  preserved: " + p
	}
	return result
}
