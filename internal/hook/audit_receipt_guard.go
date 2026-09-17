package hook

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
)

// audit_receipt_guard.go — SPEC-CODEX-AUDIT-GATE-AXES-001 axis (b), hook half.
//
// An auditor's PASS is text the auditor wrote, so it cannot show that the audit
// tool was ever called. The observed failure this closes: an auditor returned
// PASS 0.92 three times without invoking the codex gate, and the same review
// forced through the gate came back FAIL with six defects.
//
// Three additions, all inert unless the tree EXPLICITLY declares
// workflow.audit.gates.codex: required — the value that opts in:
//
//  1. SubagentStart records that an auditor instance began, so a receipt minted
//     before it cannot be recycled as evidence for it.
//  2. SubagentStop reads the auditor's verdict line and refuses a PASS whose
//     cited receipts the store does not corroborate, persisting the refusal.
//  3. PreToolUse denies the phase-entry spawns (manager-develop / manager-docs /
//     manager-git) while any refusal is outstanding, so an unproven PASS cannot
//     simply be walked past.
//
// The tree is taken from the hook input's own cwd, never CLAUDE_PROJECT_DIR:
// in a worktree session that variable names the primary checkout, and a guard
// keyed on it would compare a worktree auditor against another tree's store.
//
// @MX:ANCHOR: [AUTO] auditor receipt guard; entry points on SubagentStart, SubagentStop and the Agent/Task spawn path
// @MX:REASON: fan_in = 3 hook handlers, and a wrong answer here either blocks every spawn or silently accepts an unproven audit

// auditReceiptViolation is the deny sentinel the orchestrator matches without
// parsing the reason prose.
const auditReceiptViolation = "AUDIT_RECEIPT_VIOLATION"

// phaseEntryAgents are the spawns a rejected audit must not be walked past:
// run entry, sync entry, and the PR that follows sync.
var phaseEntryAgents = map[string]bool{
	"manager-develop": true,
	"manager-docs":    true,
	"manager-git":     true,
}

// auditReceiptTree resolves the guarded tree for a hook input, returning ""
// when the tree did not opt in. Every entry point below starts here, so a
// project that never wrote the gate value pays one config read and gets three
// no-ops.
func auditReceiptTree(input *HookInput) string {
	if input == nil {
		return ""
	}
	tree := auditreceipt.TreeRootFromCWD(input.CWD)
	if tree == "" || !auditreceipt.CodexGateRequired(tree) {
		return ""
	}
	return tree
}

// recordAuditorStart writes the start marker for an auditor subagent
// (REQ-CAG-010). Failures are logged, never surfaced: SubagentStart has no
// blocking channel, and a marker that could not be written shows up later as
// the "start marker missing" refusal cause rather than as a silent pass.
func recordAuditorStart(input *HookInput) {
	if input == nil || !auditreceipt.IsAuditorAgent(input.AgentType) || input.AgentID == "" {
		return
	}
	tree := auditReceiptTree(input)
	if tree == "" {
		return
	}
	m := auditreceipt.StartMarker{
		AgentID:   input.AgentID,
		AgentType: input.AgentType,
		SessionID: input.SessionID,
		TreeRoot:  tree,
	}
	if err := auditreceipt.WriteStartMarker(tree, &m); err != nil {
		slog.Warn("auditor start marker not recorded", "agent_id", input.AgentID, "tree_root", tree, "error", err)
	}
}

// checkAuditorStop decides an auditor's SubagentStop (REQ-CAG-011 to 013,
// REQ-CAG-016). It returns nil when the guard has no opinion — a non-auditor, a
// tree that did not opt in, or a verdict this guard does not judge.
func checkAuditorStop(input *HookInput) *HookOutput {
	if input == nil || !auditreceipt.IsAuditorAgent(input.AgentType) {
		return nil
	}
	tree := auditReceiptTree(input)
	if tree == "" {
		return nil
	}

	line, parsed := auditreceipt.ParseVerdictLine(input.LastAssistantMessage)
	if parsed && !line.IsPass() {
		// A FAIL needs no receipt: it is not claiming an audit approved anything.
		clearStartMarker(tree, input.AgentID)
		return nil
	}

	specID := auditreceipt.UnknownSpec
	cause := auditreceipt.CauseVerdictLineMissing
	var cited []string
	if parsed {
		specID = line.SpecID
		cited = line.Receipts
		start := readStartMarker(tree, input.AgentID)
		ok, failure := auditreceipt.CheckCitedReceipts(tree, start, cited)
		if ok {
			// A proven PASS clears this role's outstanding refusals — the other
			// SPECs' and the unknown-spec one included (operator decision K4).
			if err := auditreceipt.ClearRejectionsForRole(tree, input.AgentType); err != nil {
				slog.Warn("audit rejections not cleared", "agent_type", input.AgentType, "tree_root", tree, "error", err)
			}
			clearStartMarker(tree, input.AgentID)
			return nil
		}
		cause = failure
	}

	persistAuditRejection(tree, input, specID, cause, cited)

	if input.StopHookActive {
		// Re-entry: blocking again would loop the subagent against its own stop
		// hook. The refusal stays on disk, so the phase-entry spawns stay denied.
		clearStartMarker(tree, input.AgentID)
		return &HookOutput{SystemMessage: fmt.Sprintf(
			"%s: this %s PASS is not accepted — %s. Phase-entry spawns (manager-develop / manager-docs / manager-git) stay denied in %s until a PASS citing a valid audit receipt is recorded.",
			auditReceiptViolation, input.AgentType, cause, tree)}
	}
	// The start marker is deliberately KEPT on a first-stop block: the auditor
	// continues after the block, and the receipt it mints then must be able to
	// prove itself against the marker of the instance that minted it.
	return &HookOutput{
		Decision: "block",
		Reason: fmt.Sprintf(
			"%s: %s reported %s but the audit could not be corroborated — %s. Call the codex audit (codex_audit or audit_multi) for this tree, then end with a verdict line citing the receipt id it returns.",
			auditReceiptViolation, input.AgentType, passLabel(parsed, line), cause),
	}
}

func passLabel(parsed bool, line auditreceipt.VerdictLine) string {
	if parsed {
		return line.Verdict
	}
	return "no parseable verdict line"
}

// persistAuditRejection writes (or updates) the refusal record. An existing
// record keeps its original timestamp and cause and gains the re-entry mark, so
// the record reads as one refusal that was warned about rather than two.
func persistAuditRejection(tree string, input *HookInput, specID, cause string, cited []string) {
	rj, err := auditreceipt.ReadRejection(tree, input.AgentType, specID)
	if err != nil {
		rj = auditreceipt.Rejection{AgentType: input.AgentType, SpecID: specID, RejectedAt: auditreceipt.Now()}
	}
	// The cause is always the CURRENT refusal's cause: a record left showing a
	// stale reason would send the reader to a condition that has since changed.
	// The original timestamp survives, so one unresolved refusal stays one
	// refusal rather than resetting its age on every stop.
	rj.AgentID, rj.Cause, rj.CitedReceipts = input.AgentID, cause, cited
	if input.StopHookActive {
		rj.ReentryWarned = true
	}
	if err := auditreceipt.WriteRejection(tree, &rj); err != nil {
		slog.Warn("audit rejection not recorded", "agent_type", input.AgentType, "spec_id", specID, "tree_root", tree, "error", err)
	}
}

func readStartMarker(tree, agentID string) *auditreceipt.StartMarker {
	if agentID == "" {
		return nil
	}
	m, err := auditreceipt.ReadStartMarker(tree, agentID)
	if err != nil {
		return nil // absent or unreadable: both mean this instance cannot be corroborated
	}
	return &m
}

func clearStartMarker(tree, agentID string) {
	if agentID == "" {
		return
	}
	if err := auditreceipt.RemoveStartMarker(tree, agentID); err != nil {
		slog.Debug("start marker not removed", "agent_id", agentID, "error", err)
	}
}

// checkAuditReceiptSpawn is the PreToolUse consumer (REQ-CAG-014). It returns
// (DecisionDeny, reason) while any refusal is outstanding in the spawning
// tree, and ("", "") otherwise. An unreadable refusal record denies too: a
// record nobody can read is not a record that went away.
func checkAuditReceiptSpawn(input *HookInput) (decision, reason string) {
	sp, ok := extractAgentSpawn(input.ToolInput)
	if !ok || !phaseEntryAgents[sp.Agent] {
		return "", ""
	}
	tree := auditReceiptTree(input)
	if tree == "" {
		return "", ""
	}
	rejections, err := auditreceipt.ListRejections(tree)
	if err != nil {
		return DecisionDeny, fmt.Sprintf(
			"%s: the audit rejection records in %s cannot be read, so %s cannot be cleared to spawn: %v",
			auditReceiptViolation, tree, sp.Agent, err)
	}
	if len(rejections) == 0 {
		return "", ""
	}
	parts := make([]string, 0, len(rejections))
	for _, r := range rejections {
		parts = append(parts, fmt.Sprintf("%s / %s / %s", r.AgentType, r.SpecID, r.Cause))
	}
	return DecisionDeny, fmt.Sprintf(
		"%s: %s cannot be spawned while an audit PASS stands unproven in %s. Outstanding: %s. Re-run the audit through codex_audit or audit_multi and let the auditor end with a verdict line citing the receipt.",
		auditReceiptViolation, sp.Agent, tree, strings.Join(parts, "; "))
}
