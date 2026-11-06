package hook

import (
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
)

// Receipt freshness is SPEC-MCP-NEUTRALITY-001 REQ-MN-020 and nothing more: any
// commit after the audit leaves the receipt stale. SPEC-AUDIT-ENGINE-CONVERGENCE-001
// excludes the evidence-only relaxation (B3) by operator decision (card t2017), so a
// commit that only files the audit's own evidence is no exception. Each row commits
// the change on top of the judged commit and drives the auditor Stop step, the one
// production caller, with a PASS citing the receipt.
func TestAEC_EvidenceOnlyCommits_LeaveTheReceiptStale(t *testing.T) {
	progress := ".moai/specs/" + rangeSpec + "/progress.md"

	rows := []struct {
		name string
		spec string
		do   func(t *testing.T, r rangeRepo)
	}{
		{"a verdict file and a second-review record", rangeSpec, func(t *testing.T, r rangeRepo) {
			rangeWrite(t, r.root, ".moai/reports/t1/verdict.md", "verdict: PASS\nrevised\n")
			rangeWrite(t, r.root, ".moai/reports/t1/second-review.jsonl", "{}\n")
		}},
		{"every member of the audit-artifact family", rangeSpec, func(t *testing.T, r rangeRepo) {
			rangeWrite(t, r.root, ".moai/reports/t1/plan-audit-iter2.md", "iteration two\n")
			rangeWrite(t, r.root, ".moai/reports/t1/sync-audit.md", "sync\n")
			rangeWrite(t, r.root, ".moai/reports/t1/card-review.md", "review\n")
		}},
		{"the E.1 and E.4 lines of the cited SPEC", rangeSpec, func(t *testing.T, r rangeRepo) {
			body := replaceOnce(t, rangeRead(t, r.root, progress), "plan_status: audit-ready", "plan_status: audit-ready (reviewed)")
			rangeWrite(t, r.root, progress, replaceOnce(t, body, "sync_status: pending", "sync_status: complete"))
		}},
		{"a full rewrite of E.3 with E.2 byte-unchanged", rangeSpec, func(t *testing.T, r rangeRepo) {
			body := rangeRead(t, r.root, progress)
			at, end := strings.Index(body, "## §E.3"), strings.Index(body, "## §E.4")
			rangeWrite(t, r.root, progress, body[:at]+"## §E.3 Run-phase Audit-Ready Signal\n\nrun_status: complete\n\n"+body[end:])
		}},
		{"a reports-only commit while the citation names a card id", rangeCard, func(t *testing.T, r rangeRepo) {
			rangeWrite(t, r.root, ".moai/reports/t1/verdict.md", "verdict: PASS\nfiled\n")
		}},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			r := newRangeRepo(t)
			row.do(t, r)
			rangeCommit(t, r.root, row.name)
			requireStale(t, r, row.spec)
		})
	}

	t.Run("an empty commit", func(t *testing.T) {
		r := newRangeRepo(t)
		rangeGit(t, r.root, "commit", "-q", "--allow-empty", "-m", "empty")
		requireStale(t, r, rangeSpec)
	})

	t.Run("a receipt that judged an uncommitted edit, then an evidence-only commit with the edit untouched", func(t *testing.T) {
		root := newGateTree(t, "required")
		rangeWrite(t, root, ".gitignore", ".moai/state/\n")
		rangeWrite(t, root, ".moai/reports/t1/verdict.md", "verdict: PASS\n")
		rangeWrite(t, root, "internal/a.go", "package internal\n")
		rangeCommit(t, root, "base")
		rangeWrite(t, root, "internal/a.go", "package internal\n\nvar Z = 3\n")
		seedStart(t, root, "a1", auditreceipt.AgentPlanAuditor, time.Now().UTC().Add(-time.Hour))
		rid := seedReceipt(t, root, auditreceipt.Receipt{Tool: auditreceipt.ToolAuditMulti, TreeRoot: root, CodexVerdict: "pass"})
		r := rangeRepo{root: root, rid: rid, base: rangeGit(t, root, "rev-parse", "HEAD")}
		rangeWrite(t, root, ".moai/reports/t1/verdict.md", "verdict: PASS\nfiled\n")
		rangeGit(t, root, "add", ".moai/reports/t1/verdict.md")
		rangeGit(t, root, "commit", "-q", "-m", "file the verdict")
		requireStale(t, r, rangeSpec)
	})
}
