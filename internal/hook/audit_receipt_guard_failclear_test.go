package hook

// Card t1855 — an auditor that stopped without a verdict line leaves a
// verdict-line-missing refusal, and the same role's later parsed FAIL must be
// able to clear exactly that refusal, or the fix-agent spawn the FAIL calls for
// stays denied. Non-parallel (t.Setenv, package clock).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
)

const failClearNoVerdict = "I reviewed the change and found problems but forgot the verdict line."

const failClearFail = "AUDIT-VERDICT: FAIL spec=SPEC-Y-001 receipts=none"

// seedFailClearRejection writes one refusal into the store with the given
// identity; kind "" is a receipt-kind record.
func seedFailClearRejection(t *testing.T, store, tree, kind, role, spec, cause string) {
	t.Helper()
	r := auditreceipt.Rejection{AgentType: role, SpecID: spec, Cause: cause, TreeRoot: tree, Kind: kind}
	if err := auditreceipt.WriteRejectionIn(store, &r); err != nil {
		t.Fatalf("WriteRejectionIn: %v", err)
	}
}

func failClearSpawnDenied(t *testing.T, root string) bool {
	t.Helper()
	isDenied, reason := denied(runSpawn(t, receiptSpawnInput(root, "manager-develop", "Agent")))
	return isDenied && strings.HasPrefix(reason, auditReceiptViolation)
}

func failClearHas(t *testing.T, store, tree, kind, role, spec string) bool {
	t.Helper()
	if kind == "" {
		kind = auditreceipt.KindReceipt
	}
	_, err := auditreceipt.ReadRejectionInKind(store, tree, kind, role, spec)
	return err == nil
}

// Kills the mutant that leaves the FAIL branch inert (never clears), which is
// the deadlock itself.
func TestAuditorFail_ClearsOwnVerdictLineMissingRejectionAndUnblocksFixSpawn(t *testing.T) {
	root := newGateTree(t, "required")
	role := auditreceipt.AgentPlanAuditor

	if out := runStop(t, stopInput(root, role, "a1", failClearNoVerdict, false)); out.Decision != "block" {
		t.Fatalf("setup: no-verdict stop decision = %q, want block", out.Decision)
	}
	if !failClearSpawnDenied(t, root) {
		t.Fatal("setup: manager-develop spawn must be denied while the verdict-line-missing refusal stands")
	}

	if out := runStop(t, stopInput(root, role, "a1", failClearFail, false)); out.Decision != "" {
		t.Fatalf("FAIL stop decision = %q, want none", out.Decision)
	}
	if failClearSpawnDenied(t, root) {
		t.Error("manager-develop spawn is still denied after the same role's parsed FAIL")
	}
	if left, err := auditreceipt.ListRejections(root); err != nil || len(left) != 0 {
		t.Errorf("rejections after FAIL = %+v (err %v), want none", left, err)
	}
}

// Kills the mutant that clears every receipt-kind refusal of the role on a
// FAIL (cause not consulted).
func TestAuditorFail_KeepsDifferentCauseRejection(t *testing.T) {
	root := newGateTree(t, "required")
	role := auditreceipt.AgentPlanAuditor
	seedFailClearRejection(t, root, root, "", role, auditreceipt.UnknownSpec, auditreceipt.CauseVerdictLineMissing)
	seedFailClearRejection(t, root, root, "", role, "SPEC-X-001", auditreceipt.CauseNoReceiptCited)

	runStop(t, stopInput(root, role, "a1", failClearFail, false))

	if failClearHas(t, root, root, "", role, auditreceipt.UnknownSpec) {
		t.Error("the verdict-line-missing refusal of the role must be cleared by its FAIL")
	}
	if !failClearHas(t, root, root, "", role, "SPEC-X-001") {
		t.Error("a no-receipt-cited refusal must survive the FAIL")
	}
	if !failClearSpawnDenied(t, root) {
		t.Error("manager-develop spawn must stay denied by the surviving no-receipt-cited refusal")
	}
}

// Kills the mutant that clears by role and cause across trees (tree not
// consulted).
func TestAuditorFail_KeepsOtherTreeVerdictLineMissingRejection(t *testing.T) {
	fx := newWSRHookFixture(t, wsrGateRequired)
	other := fx.addWorktree(t, "W2")
	role := auditreceipt.AgentPlanAuditor
	store, err := auditreceipt.StoreRoot(fx.W)
	if err != nil {
		t.Fatalf("StoreRoot: %v", err)
	}
	own := auditreceipt.TreeRootFromCWD(fx.W)
	otherTree := auditreceipt.TreeRootFromCWD(other)
	seedFailClearRejection(t, store, own, "", role, auditreceipt.UnknownSpec, auditreceipt.CauseVerdictLineMissing)
	seedFailClearRejection(t, store, otherTree, "", role, auditreceipt.UnknownSpec, auditreceipt.CauseVerdictLineMissing)

	wsrStop(t, fx.W, "a1", failClearFail, false)

	if failClearHas(t, store, own, "", role, auditreceipt.UnknownSpec) {
		t.Error("the FAIL's own-tree verdict-line-missing refusal must be cleared")
	}
	if !failClearHas(t, store, otherTree, "", role, auditreceipt.UnknownSpec) {
		t.Error("another tree's verdict-line-missing refusal must survive the FAIL")
	}
	if failClearSpawnDenied(t, fx.W) {
		t.Error("own-tree spawn must be allowed once its only refusal is cleared")
	}
	if !failClearSpawnDenied(t, other) {
		t.Error("the other tree's spawn must stay denied")
	}
}

// Kills the mutant that clears by tree and cause across roles (role not
// consulted).
func TestAuditorFail_KeepsOtherRoleVerdictLineMissingRejection(t *testing.T) {
	root := newGateTree(t, "required")
	role := auditreceipt.AgentPlanAuditor
	seedFailClearRejection(t, root, root, "", role, auditreceipt.UnknownSpec, auditreceipt.CauseVerdictLineMissing)
	seedFailClearRejection(t, root, root, "", auditreceipt.AgentSyncAuditor, auditreceipt.UnknownSpec, auditreceipt.CauseVerdictLineMissing)

	runStop(t, stopInput(root, role, "a1", failClearFail, false))

	if failClearHas(t, root, root, "", role, auditreceipt.UnknownSpec) {
		t.Error("the FAIL role's own refusal must be cleared")
	}
	if !failClearHas(t, root, root, "", auditreceipt.AgentSyncAuditor, auditreceipt.UnknownSpec) {
		t.Error("the other auditor role's refusal must survive the FAIL")
	}
	if !failClearSpawnDenied(t, root) {
		t.Error("manager-develop spawn must stay denied by the other role's refusal")
	}
}

// Kills the mutant that clears by role, tree and cause across kinds (kind not
// consulted): the served gate owns served-kind refusals.
func TestAuditorFail_KeepsServedKindRejectionEvenWithVerdictLineMissingCause(t *testing.T) {
	root := newGateTree(t, "required")
	role := auditreceipt.AgentPlanAuditor
	seedFailClearRejection(t, root, root, "", role, auditreceipt.UnknownSpec, auditreceipt.CauseVerdictLineMissing)
	seedFailClearRejection(t, root, root, auditreceipt.KindServed, role, auditreceipt.UnknownSpec, auditreceipt.CauseVerdictLineMissing)

	runStop(t, stopInput(root, role, "a1", failClearFail, false))

	if failClearHas(t, root, root, "", role, auditreceipt.UnknownSpec) {
		t.Error("the receipt-kind refusal must be cleared by the FAIL")
	}
	if !failClearHas(t, root, root, auditreceipt.KindServed, role, auditreceipt.UnknownSpec) {
		t.Error("the served-kind refusal must survive the FAIL")
	}
}

// Kills the mutant that reaches the store when the gate is only assumed: with
// no store there is nothing to read, clear, or write.
func TestAuditorFail_AssumedGateWritesAndClearsNothing(t *testing.T) {
	fx := newWSRHookFixture(t, wsrNoGate)
	if err := os.MkdirAll(filepath.Join(fx.W, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(fx.P, ".git", "worktrees", "W", "HEAD")); err != nil {
		t.Fatal(err)
	}

	out := wsrStop(t, fx.W, "a1", failClearFail, false)

	if out != nil && (out.Decision != "" || out.SystemMessage != "") {
		t.Errorf("assumed-gate FAIL output = %+v, want no decision and no message", out)
	}
	if _, err := os.Stat(filepath.Join(fx.P, ".moai", "state", "audit-receipts")); !os.IsNotExist(err) {
		t.Errorf("no audit-receipts directory may exist under P/.moai/state (stat err=%v)", err)
	}
	if entries, err := os.ReadDir(filepath.Join(fx.W, ".moai")); err != nil || len(entries) != 0 {
		t.Errorf("W/.moai must stay empty (entries=%d err=%v)", len(entries), err)
	}
}
