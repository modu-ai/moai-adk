package homestate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/auditverdict"
)

// auditKickoffRefusal decides the audit decider's Kickoff conditions (T8a)
// and returns "" when every one holds. The card's own state (not blocked,
// not needs-decision) is guaranteed by the edge, which starts at kickoff.
func auditKickoffRefusal(cur Card, queueHold string) string {
	switch queueHold {
	case QueueHoldClear:
	case QueueHoldHeld:
		return "the queue item is on hold"
	default:
		return "the queue item's hold state is unreadable (fail closed)"
	}
	v, err := readAuditVerdict(cur.WorktreePath, cur.CardID, "plan-audit", cur.EvidenceSHA)
	if err != nil {
		return err.Error()
	}
	if ok, reason := admitVerdictFile(cur, v.Path, auditverdict.PhasePlan); !ok {
		return reason
	}
	specDir := filepath.Join(cur.WorktreePath, ".moai", "specs", cur.SpecID)
	if !auditReadyRecorded(filepath.Join(specDir, "progress.md")) {
		return "the plan phase records no audit-ready status (progress.md §E.1)"
	}
	return founderRowRefusal(filepath.Join(specDir, "decision-index.md"))
}

// auditReadyRecorded reports whether progress.md's §E.1 section carries the
// explicit `audit_ready: true` signal and no conflicting `audit_ready` value.
func auditReadyRecorded(path string) bool {
	raw, err := readBoundedFile(path)
	if err != nil {
		return false
	}
	_, after, found := strings.Cut(string(raw), "## §E.1")
	if !found {
		return false
	}
	section := after
	if i := strings.Index(after, "\n## "); i >= 0 {
		section = after[:i]
	}
	if _, rest, ok := strings.Cut(section, "\n"); ok {
		section = rest
	}
	// Only the explicit signal counts; a non-empty section is not readiness.
	ready, conflicting := false, false
	for _, line := range strings.Split(section, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok || strings.TrimSpace(key) != "audit_ready" {
			continue
		}
		if strings.TrimSpace(value) == "true" {
			ready = true
		} else {
			conflicting = true
		}
	}
	return ready && !conflicting
}

// founderRowRefusal reads decision-index.md (absent means no rows) and
// refuses on any row — whatever its label — holding DEFAULT-APPLIED outside an
// implementation-level row with a Default, and on any FOUNDER row whose
// verdict is empty.
func founderRowRefusal(path string) string {
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ""
		}
		return "decision-index unreadable: " + err.Error()
	}
	for _, block := range strings.Split("\n"+string(raw), "\n### ") {
		var label, class, verdict string
		hasDefault, hasVerdictLine := false, false
		for _, line := range strings.Split(block, "\n") {
			line = strings.TrimSpace(line)
			switch {
			case strings.HasPrefix(line, "Label:"):
				label = strings.TrimSpace(strings.TrimPrefix(line, "Label:"))
			case strings.HasPrefix(line, "Class:"):
				class = strings.TrimSpace(strings.TrimPrefix(line, "Class:"))
			case strings.HasPrefix(line, "Default:"):
				hasDefault = true
			case strings.HasPrefix(line, "Operator verdict:"):
				hasVerdictLine = true
				verdict = strings.TrimSpace(strings.TrimPrefix(line, "Operator verdict:"))
			}
		}
		head, _, _ := strings.Cut(strings.TrimSpace(block), "\n")
		// The DEFAULT-APPLIED restriction binds every row whatever its label:
		// a relabelled row must not carry a default the rule never allowed.
		if strings.HasPrefix(verdict, "DEFAULT-APPLIED") && (class != "implementation-level" || !hasDefault) {
			return fmt.Sprintf("row %q holds DEFAULT-APPLIED outside an implementation-level row with a Default", head)
		}
		if label != "FOUNDER" {
			continue
		}
		if !hasVerdictLine || verdict == "" {
			return fmt.Sprintf("FOUNDER row %q has an empty verdict", head)
		}
	}
	return ""
}
