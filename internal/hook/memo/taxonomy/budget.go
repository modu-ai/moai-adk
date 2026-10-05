// budget.go — MEMORY.md size measurement and the index budget audit
// (SPEC-MEMORY-FOLD-BUDGET-001 REQ-MFB-008 / REQ-MFB-009, plan.md M1).
//
// The doctor reports four measures — raw bytes, Unicode code points,
// loaded-content code points, and lines — and warns on configured budget
// boundaries. Raw bytes is the axis the warning keys on, because bytes is
// never smaller than characters, so it warns earliest and errs toward a false
// alarm on CJK-heavy indexes (plan.md OD-3); every finding says in its own
// text that bytes is a conservative proxy, because the loader's actual cut is
// unconfirmed (spec.md §1.4). The audit is advisory: it emits findings and
// never enforces.
package taxonomy

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Budget finding codes (REQ-MFB-009).
const (
	WarnIndexBudgetWarn  AuditCode = "MEMORY_INDEX_BUDGET_WARN"
	WarnIndexBudgetAtCap AuditCode = "MEMORY_INDEX_BUDGET_AT_CAP"
)

// budgetBasis is the basis sentence every budget finding carries (REQ-MFB-009).
const budgetBasis = "raw bytes: conservative proxy; the loader's cut is unconfirmed"

// htmlCommentPattern matches one HTML comment span; (?s) lets it run across
// lines.
var htmlCommentPattern = regexp.MustCompile(`(?s)<!--.*?-->`)

// IndexMeasurements holds the four measures the doctor reports for one
// MEMORY.md (REQ-MFB-008).
type IndexMeasurements struct {
	// Bytes is the raw file size.
	Bytes int
	// Chars is the Unicode code-point count.
	Chars int
	// LoadedChars is the code-point count after removing the leading YAML
	// frontmatter block and HTML comments — a reconstruction of the host's
	// documented exclusion, informational only and unverified against the
	// loader (spec.md §1.4).
	LoadedChars int
	// Lines is the line count, counted the way the doctor counts index_lines
	// so the two figures never diverge (AC-MFB-009: index_lines unchanged).
	Lines int
}

// MeasureIndex measures one MEMORY.md content.
func MeasureIndex(data []byte) IndexMeasurements {
	content := string(data)
	return IndexMeasurements{
		Bytes:       len(data),
		Chars:       utf8.RuneCountInString(content),
		LoadedChars: utf8.RuneCountInString(loadedContent(content)),
		Lines:       len(strings.Split(strings.TrimRight(content, "\n"), "\n")),
	}
}

// loadedContent removes the documented exclusions from content: the leading
// frontmatter block and HTML comments.
func loadedContent(content string) string {
	return htmlCommentPattern.ReplaceAllString(stripLeadingFrontmatter(content), "")
}

// stripLeadingFrontmatter removes the leading YAML frontmatter block — the
// same shape ParseFile parses: a first line of --- through a closing ---
// line, both compared trimmed. An unclosed block removes nothing.
func stripLeadingFrontmatter(content string) string {
	lines := strings.SplitAfter(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return content
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return strings.Join(lines[i+1:], "")
		}
	}
	return content
}

// AuditIndexBudget evaluates both budget axes against the caps and returns
// the findings, byte axis first. A non-positive byte or line cap, or a warn
// percent outside 1..100, falls back to the configured default, so no
// constant appears as a literal in the check. On the byte axis
// MEMORY_INDEX_BUDGET_AT_CAP replaces the warning once the cap is reached;
// on the line axis a count above the cap belongs to MEMORY_INDEX_OVERFLOW
// (AuditIndex), so the budget audit emits nothing there. The line cap fed
// here is the same one the caller gives AuditIndex, so one line cap is in
// force per invocation.
func AuditIndexBudget(indexPath string, m IndexMeasurements, byteCap, warnPercent, lineCap int) []AuditFinding {
	if byteCap <= 0 {
		byteCap = config.DefaultMemoryIndexByteCap
	}
	if warnPercent <= 0 || warnPercent > 100 {
		warnPercent = config.DefaultMemoryIndexWarnPercent
	}
	if lineCap <= 0 {
		lineCap = config.DefaultMemoryIndexLineCap
	}

	var findings []AuditFinding
	switch {
	case m.Bytes >= byteCap:
		findings = append(findings, budgetFinding(indexPath, WarnIndexBudgetAtCap, "bytes", m.Bytes, byteCap))
	case m.Bytes*100 >= warnPercent*byteCap:
		findings = append(findings, budgetFinding(indexPath, WarnIndexBudgetWarn, "bytes", m.Bytes, byteCap))
	}
	if m.Lines <= lineCap && m.Lines*100 >= warnPercent*lineCap {
		findings = append(findings, budgetFinding(indexPath, WarnIndexBudgetWarn, "lines", m.Lines, lineCap))
	}
	return findings
}

// budgetFinding builds one finding naming the axis, the measured value, the
// cap, the percentage and the basis (REQ-MFB-009).
func budgetFinding(indexPath string, code AuditCode, axis string, value, cap int) AuditFinding {
	pct := 0
	if cap > 0 {
		pct = value * 100 / cap
	}
	return AuditFinding{
		Code:   code,
		Path:   indexPath,
		Detail: fmt.Sprintf("index %s: %d of a %d-%s cap (%d%%) — %s", axis, value, cap, axis, pct, budgetBasis),
	}
}
