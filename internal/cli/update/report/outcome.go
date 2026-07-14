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
		result = "✓ Up to date"
		if backupPath != "" {
			result += "\n\nBackup: " + backupPath
			result += "\nRecover: moai update --restore-config " + backupPath
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
			result += "\nRecover: moai update --restore-config " + backupPath
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
			result += "\nRecover: moai update --restore-config " + backupPath
		}
	}

	return result
}
