package bodp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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
// @MX:NOTE Audit trail은 BODP 결정의 영구 기록. branch name 정규화
// (slash → dash) 로 filesystem-safe.
func WriteDecision(repoRoot string, entry AuditEntry) error {
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}
	fname := normalizeBranchName(entry.NewBranch) + ".md"
	fullPath := filepath.Join(repoRoot, auditTrailDir, fname)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return fmt.Errorf("create audit trail dir: %w", err)
	}
	if err := os.WriteFile(fullPath, []byte(renderAuditEntry(entry)), 0o644); err != nil {
		return fmt.Errorf("write audit trail %s: %w", fullPath, err)
	}
	return nil
}

// HasAuditTrail returns true when an audit trail file exists under
// repoRoot/.moai/branches/decisions/ for the given branch name.
//
// @MX:NOTE W7-T05 reminder false-positive 방지: 디렉토리 자체 부재 시 false
// 반환 (no error 구분). 신규 프로젝트에서는 reminder 가 발화되지 않는다.
func HasAuditTrail(repoRoot, branchName string) bool {
	if repoRoot == "" || branchName == "" {
		return false
	}
	fname := normalizeBranchName(branchName) + ".md"
	_, err := os.Stat(filepath.Join(repoRoot, auditTrailDir, fname))
	return err == nil
}

// normalizeBranchName replaces path separators with dashes so the result is
// safe to use as a filesystem filename.
func normalizeBranchName(name string) string {
	return strings.ReplaceAll(name, "/", "-")
}

// renderAuditEntry produces the canonical markdown body (frontmatter + 3
// sections: Signals, Decision, Executed) for an AuditEntry.
func renderAuditEntry(entry AuditEntry) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	fmt.Fprintf(&sb, "timestamp: %s\n", entry.Timestamp.UTC().Format(time.RFC3339))
	fmt.Fprintf(&sb, "entry_point: %s\n", entry.EntryPoint)
	fmt.Fprintf(&sb, "current_branch: %s\n", entry.CurrentBranch)
	fmt.Fprintf(&sb, "new_branch: %s\n", entry.NewBranch)
	fmt.Fprintf(&sb, "user_choice: %s\n", entry.UserChoice)
	sb.WriteString("---\n\n")
	fmt.Fprintf(&sb, "# BODP Decision: %s\n\n", entry.NewBranch)

	sb.WriteString("## Signals\n")
	fmt.Fprintf(&sb, "- Signal (a) — Code dependency: %t\n", entry.Decision.SignalA)
	fmt.Fprintf(&sb, "- Signal (b) — Working tree co-location: %t\n", entry.Decision.SignalB)
	fmt.Fprintf(&sb, "- Signal (c) — Open PR head: %t\n", entry.Decision.SignalC)

	sb.WriteString("\n## Decision\n")
	fmt.Fprintf(&sb, "- Recommended: %s\n", entry.Decision.Recommended)
	fmt.Fprintf(&sb, "- User choice: %s\n", entry.UserChoice)
	fmt.Fprintf(&sb, "- Base branch: %s\n", entry.Decision.BaseBranch)
	fmt.Fprintf(&sb, "- Rationale: %s\n", entry.Decision.Rationale)

	sb.WriteString("\n## Executed\n```\n")
	sb.WriteString(entry.ExecutedCmd)
	sb.WriteString("\n```\n")
	return sb.String()
}
