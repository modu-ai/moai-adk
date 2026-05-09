package bodp

import "time"

// auditTrailDir is the canonical relative path under repoRoot where BODP
// decisions are persisted as markdown files.
const auditTrailDir = ".moai/branches/decisions"

// AuditEntry is the structured payload written for one BODP decision.
type AuditEntry struct {
	Timestamp     time.Time
	EntryPoint    EntryPoint
	CurrentBranch string
	NewBranch     string
	Decision      BODPDecision
	UserChoice    Choice
	ExecutedCmd   string
}

// WriteDecision persists an AuditEntry as a markdown file under
// repoRoot/.moai/branches/decisions/<normalized-branch-name>.md.
//
// Stub implementation for the W7-T04 RED phase. The GREEN phase wires
// frontmatter rendering, body sections, and directory auto-creation.
func WriteDecision(repoRoot string, entry AuditEntry) error {
	_ = repoRoot
	_ = entry
	return nil
}

// HasAuditTrail returns true when an audit trail file exists under
// repoRoot/.moai/branches/decisions/ for the given branch name.
//
// Stub implementation for the W7-T04 RED phase.
func HasAuditTrail(repoRoot, branchName string) bool {
	_ = repoRoot
	_ = branchName
	return false
}
