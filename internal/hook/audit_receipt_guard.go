package hook

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

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
// A config-orphaned linked worktree (its repository keeps .moai untracked, so
// the worktree has no workflow config) reads the gate from its primary
// checkout and keeps its records in the primary checkout's store under its own
// tree identity (SPEC-WORKTREE-STATE-ROOT-001 REQ-WSR-003/009). When that
// primary cannot be identified the gate is assumed `required` and, with no
// store to write to, the guard refuses and denies without recording anything
// (REQ-WSR-010).
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

// guardTree is the resolved scope of the guard for one hook input.
type guardTree struct {
	tree  string // canonical tree identity the auditor runs in
	store string // store root holding the tree's records; "" when unresolved
}

// assumed reports the fail-closed case: a config-orphaned tree whose primary
// checkout could not be identified, so the gate is assumed `required` and no
// store exists.
func (g guardTree) assumed() bool { return g.store == "" }

// auditReceiptScope resolves the guarded tree for a hook input, returning false
// when the tree did not opt in. Every entry point below starts here, so a
// project that never wrote the gate value pays one config read and gets three
// no-ops; a tree that is not a config-orphaned worktree runs no git beyond the
// toplevel lookup it always ran.
func auditReceiptScope(input *HookInput) (guardTree, bool) {
	if input == nil {
		return guardTree{}, false
	}
	tree := auditreceipt.TreeRootFromCWD(input.CWD)
	if tree == "" {
		return guardTree{}, false
	}
	store, err := auditreceipt.StoreRoot(tree)
	if err != nil {
		return guardTree{tree: tree}, true // fail-closed: gate assumed required, no store
	}
	if !auditreceipt.CodexGateRequired(store) {
		return guardTree{}, false
	}
	return guardTree{tree: tree, store: store}, true
}

// recordAuditorStart writes the start marker for an auditor subagent
// (REQ-CAG-010). Failures are logged, never surfaced: SubagentStart has no
// blocking channel, and a marker that could not be written shows up later as
// the "start marker missing" refusal cause rather than as a silent pass.
//
// A background Agent() spawn delivers no agent_id, so the marker is keyed by
// the identity the payload does carry (session_id + agent_type, card t1544):
// keyed by agent id alone the write was a structural no-op and the auditor's
// later PASS permanently unprovable.
func recordAuditorStart(input *HookInput) {
	if input == nil || !auditreceipt.IsAuditorAgent(input.AgentType) {
		return
	}
	key := auditreceipt.StartMarkerKey(input.AgentID, input.SessionID, input.AgentType)
	if key == "" {
		return // no identity at all: no key, no marker
	}
	g, ok := auditReceiptScope(input)
	if !ok || g.assumed() {
		return // no store: nothing may be written under any root (REQ-WSR-010)
	}
	if auditreceipt.IsDerivedMarkerKey(key) {
		// A derived key is a session-era anchor shared by every same-role
		// background auditor of the session: an existing marker keeps the
		// earliest start, so a second concurrent spawn does not move
		// StartedAt forward past receipts the first instance will cite.
		// Every start still counts on the instance ledger
		// (SPEC-RECEIPT-REUSE-001): the outstanding count is what tells a
		// single-live end from an ambiguous one.
		if err := auditreceipt.RecordInstanceStart(g.store, key, auditreceipt.Now()); err != nil {
			slog.Warn("auditor start not counted", "agent_id", key, "error", err)
		}
		if _, err := auditreceipt.ReadStartMarker(g.store, key); err == nil {
			return
		}
	}
	m := auditreceipt.StartMarker{
		AgentID:   key,
		AgentType: input.AgentType,
		SessionID: input.SessionID,
		TreeRoot:  g.tree,
	}
	if err := auditreceipt.WriteStartMarker(g.store, &m); err != nil {
		slog.Warn("auditor start marker not recorded", "agent_id", key, "tree_root", g.tree, "error", err)
	}
}

// checkAuditorStop decides an auditor's SubagentStop (REQ-CAG-011 to 013,
// REQ-CAG-016). It returns nil when the guard has no opinion — a non-auditor, a
// tree that did not opt in, or a verdict this guard does not judge.
func checkAuditorStop(input *HookInput) *HookOutput {
	if input == nil || !auditreceipt.IsAuditorAgent(input.AgentType) {
		return nil
	}
	g, ok := auditReceiptScope(input)
	if !ok {
		return nil
	}

	line, parsed := auditreceipt.ParseVerdictLine(input.LastAssistantMessage)
	if parsed && !line.IsPass() {
		// A FAIL needs no receipt: it is not claiming an audit approved anything.
		if !g.assumed() {
			_, key := readStartMarker(g.store, input)
			consumeStartMarker(g.store, key)
			recordAuditorEnd(g.store, key)
		}
		return nil
	}

	specID := auditreceipt.UnknownSpec
	cause := auditreceipt.CauseVerdictLineMissing
	var cited []string
	if parsed {
		specID = line.SpecID
		cited = line.Receipts
	}
	switch {
	case g.assumed():
		// No store exists, so no receipt can corroborate anything and no
		// refusal can be recorded: refuse, and say why the gate applied.
		cause = auditreceipt.GateAssumedRequiredNote + ", so no audit receipt can be recorded or checked for this tree"
	case parsed:
		start, foundKey := readStartMarker(g.store, input)
		boundary, err := instanceEndBoundary(g.store, foundKey)
		if err != nil {
			// Fail closed (SPEC-RECEIPT-REUSE-001, --security --deep): a
			// ledger that exists but cannot be read makes the boundary
			// unknowable, and unknowable is not zero — zero would wave the
			// predecessor era through. The refusal names the condition the
			// operator fixes, the same reading checkAuditReceiptSpawn applies
			// to unreadable rejection records.
			slog.Warn("instance ledger unreadable", "agent_id", foundKey, "error", err)
			cause = auditreceipt.CauseInstanceLedgerUnreadable
			break
		}
		ok, failure := auditreceipt.CheckCitedReceiptsSince(g.store, start, boundary, cited)
		if ok {
			// A proven PASS clears this role's outstanding refusals in THIS tree —
			// the other SPECs' and the unknown-spec one included (operator
			// decision K4) — and never another tree's (REQ-WSR-003). Only
			// receipt-kind refusals: a served-model refusal is the served
			// gate's to clear (SPEC-SERVED-MODEL-AUDIT-001 REQ-SMA-011).
			if err := auditreceipt.ClearRejectionsForRoleInTreeKind(g.store, g.tree, input.AgentType, auditreceipt.KindReceipt); err != nil {
				slog.Warn("audit rejections not cleared", "agent_type", input.AgentType, "tree_root", g.tree, "error", err)
			}
			consumeStartMarker(g.store, foundKey)
			recordAuditorEnd(g.store, foundKey)
			return nil
		}
		cause = failure
	}

	if !g.assumed() {
		persistAuditRejection(g, input, specID, cause, cited)
	}

	if input.StopHookActive {
		// Re-entry: blocking again would loop the subagent against its own stop
		// hook. The refusal stays on disk (or, with no store, the spawn check
		// fails closed on its own), so the phase-entry spawns stay denied.
		if !g.assumed() {
			_, key := readStartMarker(g.store, input)
			consumeStartMarker(g.store, key)
			recordAuditorEnd(g.store, key)
		}
		return &HookOutput{SystemMessage: fmt.Sprintf(
			"%s: this %s PASS is not accepted — %s. Phase-entry spawns (manager-develop / manager-docs / manager-git) stay denied in %s until a PASS citing a valid audit receipt is recorded.",
			auditReceiptViolation, input.AgentType, cause, g.tree)}
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

// persistAuditRejection writes (or updates) the refusal record of this tree.
// An existing record keeps its original timestamp and cause and gains the
// re-entry mark, so the record reads as one refusal that was warned about
// rather than two.
func persistAuditRejection(g guardTree, input *HookInput, specID, cause string, cited []string) {
	rj, err := auditreceipt.ReadRejectionIn(g.store, g.tree, input.AgentType, specID)
	if err != nil {
		rj = auditreceipt.Rejection{AgentType: input.AgentType, SpecID: specID, RejectedAt: auditreceipt.Now()}
	}
	// The cause is always the CURRENT refusal's cause: a record left showing a
	// stale reason would send the reader to a condition that has since changed.
	// The original timestamp survives, so one unresolved refusal stays one
	// refusal rather than resetting its age on every stop.
	rj.AgentID, rj.Cause, rj.CitedReceipts, rj.TreeRoot = input.AgentID, cause, cited, g.tree
	if input.StopHookActive {
		rj.ReentryWarned = true
	}
	if err := auditreceipt.WriteRejectionIn(g.store, &rj); err != nil {
		slog.Warn("audit rejection not recorded", "agent_type", input.AgentType, "spec_id", specID, "tree_root", g.tree, "error", err)
	}
}

// markerKeys lists the store keys this instance's start marker may be filed
// under, best candidate first: the agent id the payload carried, then the
// session_id+agent_type key a background spawn's start was recorded under —
// its SubagentStart carries no agent id while its stop payload may still
// carry one, so both spellings are tried (card t1544).
func markerKeys(input *HookInput) []string {
	var keys []string
	if id := strings.TrimSpace(input.AgentID); id != "" {
		keys = append(keys, id)
	}
	if bg := auditreceipt.StartMarkerKey("", input.SessionID, input.AgentType); bg != "" {
		keys = append(keys, bg)
	}
	return keys
}

// readStartMarker loads this instance's start marker, trying every key it may
// be filed under. It returns the marker and the store key it was found under;
// absent or unreadable under all of them returns a nil marker and "" — this
// instance cannot be corroborated.
func readStartMarker(store string, input *HookInput) (*auditreceipt.StartMarker, string) {
	for _, key := range markerKeys(input) {
		m, err := auditreceipt.ReadStartMarker(store, key)
		if err == nil {
			return &m, key
		}
	}
	return nil, ""
}

// consumeStartMarker removes the start marker found under key. A derived
// background key is deliberately NOT removed: it is a session-era anchor
// shared by every same-role background auditor of the session, so one
// instance's stop must not destroy another live instance's marker — that
// deletion re-created the unprovable-PASS deadlock for the second instance
// (card t1544 card-review P2). Derived anchors go stale with their session id
// and are never read again.
func consumeStartMarker(store, key string) {
	if key == "" || auditreceipt.IsDerivedMarkerKey(key) {
		return
	}
	if err := auditreceipt.RemoveStartMarker(store, key); err != nil {
		slog.Debug("start marker not removed", "agent_id", key, "error", err)
	}
}

// recordAuditorEnd counts a terminal instance end on the derived key's
// instance ledger (SPEC-RECEIPT-REUSE-001). Only a derived key carries the
// ledger: an agent-id-keyed marker is the instance's own boundary, consumed at
// its own stop. A first-stop block is deliberately NOT terminal — the
// instance continues and must stay able to prove the receipt it mints then
// (REQ-RR-005) — so the call sites are exactly the accepted PASS, the FAIL
// verdict, and the re-entry refusal.
func recordAuditorEnd(store, key string) {
	if key == "" || !auditreceipt.IsDerivedMarkerKey(key) {
		return
	}
	if err := auditreceipt.RecordInstanceEnd(store, key, auditreceipt.Now()); err != nil {
		slog.Warn("auditor end not counted", "agent_id", key, "error", err)
	}
}

// instanceEndBoundary returns the end-event boundary a citation judged against
// the marker found under key must respect (SPEC-RECEIPT-REUSE-001), or the
// error when the ledger exists but cannot be read — the caller fails closed on
// that. A zero boundary with no error means no single-live predecessor end has
// sealed the era, and the anchor's own StartedAt is the only fence, as before.
// Only a derived key carries a ledger: an agent-id-keyed marker starts with
// its instance, so its before-start fence already covers predecessor receipts.
func instanceEndBoundary(store, key string) (time.Time, error) {
	if key == "" || !auditreceipt.IsDerivedMarkerKey(key) {
		return time.Time{}, nil
	}
	l, err := auditreceipt.ReadInstanceLedger(store, key)
	if err != nil {
		return time.Time{}, err
	}
	return l.EndedAt, nil
}

// checkAuditReceiptSpawn is the PreToolUse consumer (REQ-CAG-014). It returns
// (DecisionDeny, reason) while any refusal of the spawning tree is outstanding,
// and ("", "") otherwise. An unreadable refusal record denies too: a record
// nobody can read is not a record that went away.
func checkAuditReceiptSpawn(input *HookInput) (decision, reason string) {
	sp, ok := extractAgentSpawn(input.ToolInput)
	if !ok || !phaseEntryAgents[sp.Agent] {
		return "", ""
	}
	g, ok := auditReceiptScope(input)
	if !ok {
		return "", ""
	}
	if g.assumed() {
		return DecisionDeny, fmt.Sprintf(
			"%s: %s cannot be spawned from %s: %s, so no audit receipt can be recorded or checked there. Make the primary checkout identifiable, or give this worktree its own .moai/config/sections/workflow.yaml.",
			auditReceiptViolation, sp.Agent, g.tree, auditreceipt.GateAssumedRequiredNote)
	}
	rejections, err := auditreceipt.ListRejectionsForTree(g.store, g.tree)
	if err != nil {
		return DecisionDeny, fmt.Sprintf(
			"%s: the audit rejection records in %s cannot be read, so %s cannot be cleared to spawn: %v",
			auditReceiptViolation, g.store, sp.Agent, err)
	}
	parts := make([]string, 0, len(rejections))
	for _, r := range rejections {
		// Served-kind refusals are listed by the served gate under its own
		// sentinel (checkServedModelSpawn), never here.
		if auditreceipt.RejectionKind(r) != auditreceipt.KindReceipt {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s / %s / %s", r.AgentType, r.SpecID, r.Cause))
	}
	if len(parts) == 0 {
		return "", ""
	}
	return DecisionDeny, fmt.Sprintf(
		"%s: %s cannot be spawned while an audit PASS stands unproven in %s. Outstanding: %s. Re-run the audit through codex_audit or audit_multi and let the auditor end with a verdict line citing the receipt.",
		auditReceiptViolation, sp.Agent, g.tree, strings.Join(parts, "; "))
}
