package hook

import (
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
)

// Sync-audit 2 defect D10 (card t2017, SPEC-AUDIT-ENGINE-CONVERGENCE-001 REQ-AEC-010,
// AC-AEC-003b): a document shape that makes the section parser miss a real
// "## §E.2" heading left both section lists empty and equal, so an edit under that
// heading kept the audit receipt fresh. Each row commits a progress.md with the shape
// before E.2, records a receipt judged at that commit, rewrites one line under the real
// E.2 heading in a later commit, and drives the auditor Stop step (the production
// caller). Every row must read as stale.
func TestAEC003b_E2FenceShapes_RewriteUnderE2MakesTheReceiptStale(t *testing.T) {
	const e2 = "## §E.2 Run-phase Evidence\n\n| AC-1 | PASS |\n\n## §E.3 Run\n\nrun\n\n## §E.4 Sync\n\nsync_status: pending\n"
	rows := []struct{ name, head string }{
		{"control-no-fence", "# P\n\n## §E.1 Plan\n\nplain\n\n"},
		{"control-plain-fence", "# P\n\n## §E.1 Plan\n```text\nplain\n```\n\n"},
		{"D1-four-backtick-holding-three", "# P\n\n## §E.1 Plan\n````text\nA fence opens with a line like:\n```\n````\n\n"},
		{"D1-four-tilde-holding-three", "# P\n\n## §E.1 Plan\n~~~~\n~~~\n~~~~\n\n"},
		{"D1-tilde-holding-backticks", "# P\n\n## §E.1 Plan\n~~~\n```\n~~~\n\n"},
		{"inline-code-starts-a-line", "# P\n\n## §E.1 Plan\n```foo``` is inline code, not a fence\n\n"},
		{"two-inline-code-lines", "# P\n\n## §E.1 Plan\n```a``` one\n```b``` two\n\n"},
		{"backtick-info-strings", "# P\n\n## §E.1 Plan\n```a`b\n```c`d\n\n"},
		{"list-fence-deep-open-shallow-close", "# P\n\n## §E.1 Plan\n- item\n\n    ```\n    code\n  ```\n\n"},
		{"html-comment-holding-backticks", "# P\n\n## §E.1 Plan\n<!--\n```\n-->\n\n"},
		{"html-pre-holding-tildes", "# P\n\n## §E.1 Plan\n<pre>\n~~~\n</pre>\n\n"},
		{"four-open-three-close-unterminated", "# P\n\n## §E.1 Plan\n````\ncode\n```\n\n"},
		{"info-string-as-closer-unterminated", "# P\n\n## §E.1 Plan\n```text\ncode\n```text\n\n"},
		{"indented-four-open-shallow-close", "# P\n\n## §E.1 Plan\n\n    ```\ncode\n```\n\n"},
		{"plain-open-never-closed", "# P\n\n## §E.1 Plan\n```\ncode\n\n"},
	}
	path := ".moai/specs/" + rangeSpec + "/progress.md"
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			root := newGateTree(t, "required")
			rangeWrite(t, root, ".gitignore", ".moai/state/\n")
			rangeWrite(t, root, path, row.head+e2)
			rangeCommit(t, root, "base")
			seedStart(t, root, "a1", auditreceipt.AgentPlanAuditor, time.Now().UTC().Add(-time.Hour))
			rid := seedReceipt(t, root, auditreceipt.Receipt{Tool: auditreceipt.ToolAuditMulti, TreeRoot: root, CodexVerdict: "pass"})
			r := rangeRepo{root: root, rid: rid, base: rangeGit(t, root, "rev-parse", "HEAD")}
			rangeWrite(t, root, path, replaceOnce(t, rangeRead(t, root, path), "| AC-1 | PASS |", "| AC-1 | FAIL |"))
			rangeCommit(t, root, "rewrite one E.2 line")
			requireStale(t, r, rangeSpec)
		})
	}

	t.Run("control: a well-formed document edited outside E.2 keeps the receipt", func(t *testing.T) {
		r := newRangeRepo(t)
		rangeWrite(t, r.root, path, replaceOnce(t, rangeRead(t, r.root, path), "sync_status: pending", "sync_status: complete"))
		rangeCommit(t, r.root, "close the sync line")
		requireFresh(t, r, rangeSpec)
	})
}
