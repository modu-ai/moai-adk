package session

import (
	"fmt"
)

// TaskLedgerEntry represents a single entry in the task ledger.
type TaskLedgerEntry struct {
	Timestamp string `json:"timestamp"`
	Action    string `json:"action"` // "phase_start", "phase_complete", "blocker", "blocker_resolved"
	Phase     Phase  `json:"phase"`
	SPECID    string `json:"spec_id"`
	Detail    string `json:"detail,omitempty"`
}

// ToMarkdown converts the ledger entry to a markdown format.
func (e TaskLedgerEntry) ToMarkdown() string {
	detail := ""
	if e.Detail != "" {
		detail = fmt.Sprintf(" - %s", e.Detail)
	}
	return fmt.Sprintf("- **%s** %s: %s%s\n", e.Timestamp, e.Action, e.Phase, detail)
}
