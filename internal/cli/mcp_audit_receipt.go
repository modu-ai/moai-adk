package cli

import (
	"log/slog"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
)

// mcp_audit_receipt.go — SPEC-CODEX-AUDIT-GATE-AXES-001 axis (b), server half.
//
// Every codex audit this server performs leaves a receipt in the audited
// tree's receipt store, and a tree that declared workflow.audit.gates.codex:
// required also gets the receipt id back on the result, so the auditor can
// cite it. Nothing about the verdict changes: a tree that did not declare the
// gate sees the same bytes it saw before, because the field is omitempty and
// stays unset there.
//
// The receipt is what makes "an audit ran" checkable at all — an auditor's own
// PASS is text, and text cannot show a tool was called.

// auditReceiptFallbackRoot resolves the tree a receipt belongs to when the
// caller named none. It is a seam so a test never writes into the repository
// it is running inside (plan.md §B.5 test isolation).
var auditReceiptFallbackRoot = resolveProjectDir

// receiptTarget names the tree a receipt is written to and how that tree was
// chosen. The distinction is load-bearing downstream: a fallback-rooted
// receipt is recorded against the server's own tree, which a worktree auditor
// correctly refuses as "a different tree".
func receiptTarget(projectRoot string) (root, source string) {
	if projectRoot != "" {
		return auditreceipt.Canonical(projectRoot), auditreceipt.RootSourceArgument
	}
	return auditreceipt.Canonical(auditReceiptFallbackRoot()), auditreceipt.RootSourceFallback
}

// recordAuditReceipt writes one receipt for an audit call and returns the id to
// EXPOSE on the result — empty unless the audited tree explicitly declared the
// codex gate required. A write failure is logged and returns an empty id: the
// receipt is evidence, and failing the audit because the evidence could not be
// filed would turn a disk problem into a verdict.
func recordAuditReceipt(tool, projectRoot, codexVerdict, gateUnmet string) string {
	root, source := receiptTarget(projectRoot)
	if root == "" {
		return ""
	}
	r := auditreceipt.Receipt{
		Tool:         tool,
		TreeRoot:     root,
		RootSource:   source,
		CodexVerdict: codexVerdict,
		GateUnmet:    gateUnmet,
	}
	id, err := auditreceipt.WriteReceipt(root, &r)
	if err != nil {
		slog.Warn("audit receipt not recorded", "tool", tool, "tree_root", root, "error", err)
		return ""
	}
	if !auditreceipt.CodexGateRequired(root) {
		return ""
	}
	return id
}
