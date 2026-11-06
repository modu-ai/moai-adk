package cli

import (
	"log/slog"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
)

// auditProvenanceFor records the auditor identity on a codex/glm audit
// verdict (SPEC-MCP-NEUTRALITY-001 REQ-MN-020): the same provenance surface
// the claude_audit verdict fills, so every backend's verdict names the model
// that produced it. Only the fields these transports actually know are set —
// the token counters stay nil (a pointer nil is the provenance contract's
// "not measured"), and the unknown strings stay empty rather than being
// invented.
func auditProvenanceFor(backend, requestedModel, resolvedModel, resolvedEffort string) *AuditProvenance {
	var transport string
	switch backend {
	case "codex":
		transport = "codex app-server"
	case "glm":
		transport = "z.ai http"
	default:
		transport = backend + " transport"
	}
	return &AuditProvenance{
		Backend:         backend,
		Transport:       transport,
		Source:          "moai mcp-server",
		RequestedModel:  requestedModel,
		ResolvedModel:   resolvedModel,
		RequestedEffort: resolvedEffort,
		ToolSurface:     backend + "_audit",
	}
}

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

// recordAuditReceiptBound writes one receipt for an audit call and returns the id
// to EXPOSE on the result — empty unless the audited tree explicitly declared the
// codex gate required. A write failure is logged and returns an empty id: the
// receipt is evidence, and failing the audit because the evidence could not be
// filed would turn a disk problem into a verdict. The receipt carries the audited
// tree's identity and is kept in that tree's store root — the primary checkout
// for a config-orphaned worktree (SPEC-WORKTREE-STATE-ROOT-001 REQ-WSR-002/003);
// where that primary cannot be identified there is no store: the write is
// skipped, never redirected, and notice says so (REQ-WSR-004). It is the plain
// receipt writer with the receipt's state
// binding SUPPLIED by the caller (card t1632 sync-audit F5): the caller
// captured the tree state at AUDIT START, and the receipt must attest THAT
// state — a post-audit bind would let a tree edit during the audit launder a
// PASS through the freshness check. Card t1632 sync-audit r3 F14: a capture
// FAILURE (bindFailed) records the explicit bind_capture_failed marker
// instead — the write-time fallback is forbidden, because a binding written
// after the audit judged nothing.
func recordAuditReceiptBound(tool, projectRoot, codexVerdict, codexError, gateUnmet, bindHead, bindDigest string, bindFailed bool) (id, notice string) {
	return recordAuditReceiptAtBound(tool, projectRoot, codexVerdict, codexError, gateUnmet, nil, bindHead, bindDigest, bindFailed)
}

// recordAuditReceiptAtBound records one receipt with the codex gate the
// exposure follows supplied by the caller: a non-nil codexRequired is the gate
// fixed at call start (the one the unmet-gate enforcement used), so the id
// exposed and the verdict persisted cannot disagree under an edit of the
// tree's configuration made while the backends ran
// (SPEC-AUDIT-MODEL-CONVERGE-001 REQ-ACV-009). A nil codexRequired keeps the
// read of the tree's configuration that every caller carrying no call-start
// plan has always had. The receipt is recorded either way; only its exposure
// follows the gate.
// Card t1632 sync-audit F5: the caller-supplied state binding — non-empty
// bindHead/bindDigest are written into the receipt verbatim, so the receipt
// attests the state the audit STARTED with. Card t1632 sync-audit r3 F14:
// bindFailed records the explicit capture-failed marker — the receipt then
// carries NO binding and the reuse check refuses it; the write-time fallback
// is forbidden, because a binding written after the audit judged nothing.
func recordAuditReceiptAtBound(tool, projectRoot, codexVerdict, codexError, gateUnmet string, codexRequired *bool, bindHead, bindDigest string, bindFailed bool) (id, notice string) {
	root, source := receiptTarget(projectRoot)
	if root == "" {
		return "", ""
	}
	store, err := auditreceipt.StoreRoot(root)
	if err != nil {
		return "", "audit receipt not recorded: " + err.Error()
	}
	r := auditreceipt.Receipt{
		Tool:              tool,
		TreeRoot:          root,
		RootSource:        source,
		CodexVerdict:      codexVerdict,
		CodexError:        codexError,
		GateUnmet:         gateUnmet,
		HeadCommit:        bindHead,
		TreeDigest:        bindDigest,
		BindCaptureFailed: bindFailed,
	}
	id, err = auditreceipt.WriteReceipt(store, &r)
	if err != nil {
		slog.Warn("audit receipt not recorded", "tool", tool, "tree_root", root, "error", err)
		return "", ""
	}
	var required bool
	if codexRequired != nil {
		required = *codexRequired
	} else {
		required = receiptCodexGateRequired(root)
	}
	if !required {
		return "", ""
	}
	return id, ""
}
