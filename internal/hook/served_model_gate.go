// served_model_gate.go — the opt-in adoption-refusal layer on top of the
// served-model observation.
//
// A gate auditor (plan-auditor / sync-auditor) whose responses came from a
// model other than the expected one — or whose served model cannot be
// determined — has produced a verdict nobody can vouch for. With
// workflow.served_model_gate.enabled: true in the tree the auditor ran in, the
// SubagentStop hook records a served-kind refusal for that role and SPEC, and
// the PreToolUse hook denies the phase-entry spawns while it stands. The audit
// run itself is never blocked, re-run, or rewritten: blocking the same
// instance cannot change which model serves it, so a block would only loop.
//
// The gate has its own scope predicate. It never widens auditReceiptScope,
// which three receipt-guard entry points share: a tree that turns on only the
// served gate must not have the receipt guard armed behind its back.
package hook

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
)

// servedModelViolation is the deny sentinel of the served-model gate, matched
// by the orchestrator without parsing the reason prose.
const servedModelViolation = "SERVED_MODEL_VIOLATION"

// servedGateScope resolves the tree and store the served gate acts on,
// returning false when the gate is not enabled there. The tree comes from the
// hook input's own cwd, as for the receipt guard; a config-orphaned worktree
// reads the gate from, and records into, its primary checkout's store. When
// that store cannot be identified the gate records nothing — the receipt
// guard's fail-closed assumption belongs to the codex gate, not to this one.
func servedGateScope(input *HookInput) (guardTree, bool) {
	if input == nil {
		return guardTree{}, false
	}
	tree := auditreceipt.TreeRootFromCWD(input.CWD)
	if tree == "" {
		return guardTree{}, false
	}
	store, err := auditreceipt.StoreRoot(tree)
	if err != nil {
		slog.Warn("served-model gate: store unresolved, nothing recorded", "tree_root", tree, "error", err)
		return guardTree{}, false
	}
	if !auditreceipt.ServedGateEnabled(store) {
		return guardTree{}, false
	}
	return guardTree{tree: tree, store: store}, true
}

// servedRefusalCause renders the stored cause of a served-kind refusal.
func servedRefusalCause(obs ServedObservation) string {
	base := auditreceipt.CauseServedModelDrift
	if obs.Verdict == ServedVerdictUnknown {
		base = auditreceipt.CauseServedModelUnknown + " (" + obs.Cause + ")"
	}
	expected := obs.ExpectedModel
	if expected == "" {
		expected = "-"
	}
	return fmt.Sprintf("%s: expected %s, served [%s]", base, expected, strings.Join(obs.ServedModels, ", "))
}

// checkServedModelStop applies the gate to one observed stop. It persists a
// served-kind refusal for a gate auditor whose verdict is served_drift or
// unknown, clears that role's served-kind refusals in the tree on `ok`, and
// returns the notice to surface. It never returns a decision.
func checkServedModelStop(input *HookInput, obs ServedObservation) string {
	if input == nil || obs.Verdict == "" || !auditreceipt.IsAuditorAgent(input.AgentType) {
		return ""
	}
	g, ok := servedGateScope(input)
	if !ok {
		return ""
	}
	switch obs.Verdict {
	case ServedVerdictOK:
		if err := auditreceipt.ClearRejectionsForRoleInTreeKind(g.store, g.tree, input.AgentType, auditreceipt.KindServed); err != nil {
			slog.Warn("served-model refusals not cleared", "agent_type", input.AgentType, "tree_root", g.tree, "error", err)
		}
		return ""
	case ServedVerdictDrift, ServedVerdictUnknown:
	default:
		return ""
	}

	specID := auditreceipt.UnknownSpec
	if line, parsed := auditreceipt.ParseVerdictLine(input.LastAssistantMessage); parsed {
		specID = line.SpecID
	}
	cause := servedRefusalCause(obs)
	rj, err := auditreceipt.ReadRejectionInKind(g.store, g.tree, auditreceipt.KindServed, input.AgentType, specID)
	if err != nil {
		rj = auditreceipt.Rejection{AgentType: input.AgentType, SpecID: specID, RejectedAt: auditreceipt.Now()}
	}
	rj.AgentID, rj.Cause, rj.TreeRoot, rj.Kind = input.AgentID, cause, g.tree, auditreceipt.KindServed
	if err := auditreceipt.WriteRejectionIn(g.store, &rj); err != nil {
		slog.Warn("served-model refusal not recorded", "agent_type", input.AgentType, "spec_id", specID, "tree_root", g.tree, "error", err)
		return ""
	}
	return fmt.Sprintf(
		"%s: the %s verdict for %s is not adopted — %s. Phase-entry spawns (manager-develop / manager-docs / manager-git) stay denied in %s until a later %s run is observed on the expected model.",
		servedModelViolation, input.AgentType, specID, cause, g.tree, input.AgentType)
}

// checkServedModelSpawn is the PreToolUse consumer of served-kind refusals. It
// denies a phase-entry spawn while any served-kind refusal of the spawning
// tree is outstanding and the gate is enabled there; an unreadable refusal
// record denies too.
func checkServedModelSpawn(input *HookInput) (decision, reason string) {
	sp, ok := extractAgentSpawn(input.ToolInput)
	if !ok || !phaseEntryAgents[sp.Agent] {
		return "", ""
	}
	g, ok := servedGateScope(input)
	if !ok {
		return "", ""
	}
	all, err := auditreceipt.ListRejectionsForTree(g.store, g.tree)
	if err != nil {
		return DecisionDeny, fmt.Sprintf(
			"%s: the refusal records in %s cannot be read, so %s cannot be cleared to spawn: %v",
			servedModelViolation, g.store, sp.Agent, err)
	}
	var parts []string
	for _, r := range all {
		if auditreceipt.RejectionKind(r) == auditreceipt.KindServed {
			parts = append(parts, fmt.Sprintf("%s / %s / %s", r.AgentType, r.SpecID, r.Cause))
		}
	}
	if len(parts) == 0 {
		return "", ""
	}
	return DecisionDeny, fmt.Sprintf(
		"%s: %s cannot be spawned while an auditor verdict stands unadopted in %s because of the model that served it. Outstanding: %s. Re-run the auditor on the expected model.",
		servedModelViolation, sp.Agent, g.tree, strings.Join(parts, "; "))
}
