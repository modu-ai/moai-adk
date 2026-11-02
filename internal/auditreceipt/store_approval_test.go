package auditreceipt

// SPEC-AUDIT-ENGINE-CONVERGENCE-001 M2 — ClearRejectionOnApproval, the store half
// of the operator approval path (REQ-AEC-013; AC-AEC-004a, AC-AEC-004b; card
// t2017).
//
// Written together with the function (the symbol is not observable before it
// exists); each assertion was checked by mutating the function and watching it
// fail. Every test works in t.TempDir() directories: none touches a real
// .moai/state/audit-receipts.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func approvalLogPath(store string) string { return filepath.Join(StateDir(store), approvalsFile) }

func readApprovalLines(t *testing.T, store string) []string {
	t.Helper()
	b, err := os.ReadFile(approvalLogPath(store))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read approval log: %v", err)
	}
	return strings.Split(strings.TrimRight(string(b), "\n"), "\n")
}

func markFile(store, treeRoot, kind, agent, spec string) string {
	return rejectionPathInKind(store, treeRoot, kind, agent, spec)
}

func approvalFor(agent, spec string) *Approval {
	a := &Approval{
		Operator: "Op Erator", Email: "op@example.test", Confirmation: agent + " " + spec,
		EvidencePath: ".moai/reports/c/verdict.md", EvidenceSHA256: strings.Repeat("ab", 32), Head: strings.Repeat("c", 40),
	}
	a.Mark.AgentType, a.Mark.SpecID = agent, spec
	return a
}

func writeMark(t *testing.T, store, tree, agent, spec, cause, kind string) {
	t.Helper()
	r := Rejection{AgentType: agent, SpecID: spec, Cause: cause, TreeRoot: tree, Kind: kind}
	if err := WriteRejectionIn(store, &r); err != nil {
		t.Fatalf("WriteRejectionIn: %v", err)
	}
}

func exists(p string) bool { _, err := os.Stat(p); return err == nil }

// The named mark is removed, exactly one record is appended carrying the mark as
// it was read, and every other mark (another SPEC, another agent, another kind,
// another tree) is byte-identical.
func TestClearRejectionOnApproval_RemovesExactlyTheNamedMark_AndRecordsIt(t *testing.T) {
	store := t.TempDir()
	other := filepath.Join(store, "elsewhere")
	writeMark(t, store, store, AgentPlanAuditor, "SPEC-A-001", CauseNoReceiptCited, "")
	writeMark(t, store, store, AgentPlanAuditor, "SPEC-B-002", CauseReceiptAuditNotRun, "")
	writeMark(t, store, store, AgentSyncAuditor, "SPEC-A-001", CauseReceiptVerdictFail, "")
	writeMark(t, store, store, AgentPlanAuditor, "SPEC-A-001", CauseServedModelDrift, KindServed)
	writeMark(t, store, other, AgentPlanAuditor, "SPEC-A-001", CauseReceiptUnknown, "")

	keep := []string{
		markFile(store, store, KindReceipt, AgentPlanAuditor, "SPEC-B-002"),
		markFile(store, store, KindReceipt, AgentSyncAuditor, "SPEC-A-001"),
		markFile(store, store, KindServed, AgentPlanAuditor, "SPEC-A-001"),
		markFile(store, other, KindReceipt, AgentPlanAuditor, "SPEC-A-001"),
	}
	before := map[string][]byte{}
	for _, p := range keep {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("setup: %s: %v", p, err)
		}
		before[p] = b
	}
	target := markFile(store, store, KindReceipt, AgentPlanAuditor, "SPEC-A-001")

	a := approvalFor(AgentPlanAuditor, "SPEC-A-001")
	if err := ClearRejectionOnApproval(store, store, a); err != nil {
		t.Fatalf("ClearRejectionOnApproval: %v", err)
	}

	if exists(target) {
		t.Errorf("the named mark must be removed")
	}
	for p, b := range before {
		if got, err := os.ReadFile(p); err != nil || string(got) != string(b) {
			t.Errorf("%s must be untouched (read err %v)", p, err)
		}
	}
	lines := readApprovalLines(t, store)
	if len(lines) != 1 {
		t.Fatalf("exactly one approval record, got %d", len(lines))
	}
	for what, want := range map[string]string{
		"the operator":       "Op Erator",
		"the git identity":   "op@example.test",
		"the mark's cause":   CauseNoReceiptCited, // read from the store, not supplied
		"the mark's SPEC ID": "SPEC-A-001",
		"the evidence path":  ".moai/reports/c/verdict.md",
		"the evidence hash":  strings.Repeat("ab", 32),
		"the HEAD":           strings.Repeat("c", 40),
		"the tree root":      store,
	} {
		if !strings.Contains(lines[0], want) {
			t.Errorf("the record does not carry %s (%q): %s", what, want, lines[0])
		}
	}
	if strings.Contains(lines[0], CauseReceiptAuditNotRun) || strings.Contains(lines[0], CauseServedModelDrift) {
		t.Errorf("the record carries another mark's cause: %s", lines[0])
	}
}

// The mark is taken from the store, not from the caller: a caller-supplied cause
// is overwritten with what was read.
func TestClearRejectionOnApproval_RecordsTheMarkAsRead_NotAsSupplied(t *testing.T) {
	store := t.TempDir()
	writeMark(t, store, store, AgentPlanAuditor, "SPEC-A-001", CauseNoReceiptCited, "")
	a := approvalFor(AgentPlanAuditor, "SPEC-A-001")
	a.Mark.Cause = "a cause the caller made up"
	if err := ClearRejectionOnApproval(store, store, a); err != nil {
		t.Fatal(err)
	}
	line := readApprovalLines(t, store)[0]
	if strings.Contains(line, "made up") || !strings.Contains(line, CauseNoReceiptCited) {
		t.Errorf("the record must carry the stored cause: %s", line)
	}
}

// The log is append-only: a second approval adds a second line and leaves the
// first byte-identical.
func TestClearRejectionOnApproval_LogIsAppendOnly(t *testing.T) {
	store := t.TempDir()
	writeMark(t, store, store, AgentPlanAuditor, "SPEC-A-001", CauseNoReceiptCited, "")
	writeMark(t, store, store, AgentPlanAuditor, "SPEC-B-002", CauseNoReceiptCited, "")
	if err := ClearRejectionOnApproval(store, store, approvalFor(AgentPlanAuditor, "SPEC-A-001")); err != nil {
		t.Fatal(err)
	}
	first := readApprovalLines(t, store)[0]
	if err := ClearRejectionOnApproval(store, store, approvalFor(AgentPlanAuditor, "SPEC-B-002")); err != nil {
		t.Fatal(err)
	}
	lines := readApprovalLines(t, store)
	if len(lines) != 2 || lines[0] != first {
		t.Errorf("want two lines with the first unchanged, got %d lines (first changed: %v)", len(lines), len(lines) > 0 && lines[0] != first)
	}
}

// An absent mark, and a mark that exists only for another tree, are refused with
// ErrMarkNotFound and nothing is written or removed.
func TestClearRejectionOnApproval_AbsentOrOtherTreeMark_IsRefused_NothingChanges(t *testing.T) {
	store := t.TempDir()
	other := filepath.Join(store, "elsewhere")
	writeMark(t, store, other, AgentPlanAuditor, "SPEC-A-001", CauseNoReceiptCited, "") // another tree's mark only
	otherFile := markFile(store, other, KindReceipt, AgentPlanAuditor, "SPEC-A-001")
	before, err := os.ReadFile(otherFile)
	if err != nil {
		t.Fatal(err)
	}

	for name, spec := range map[string]string{"absent": "SPEC-NOPE-999", "recorded for another tree": "SPEC-A-001"} {
		err := ClearRejectionOnApproval(store, store, approvalFor(AgentPlanAuditor, spec))
		if !errors.Is(err, ErrMarkNotFound) {
			t.Errorf("%s: want ErrMarkNotFound, got %v", name, err)
		}
	}
	if exists(approvalLogPath(store)) {
		t.Errorf("a refused call must write no approval record")
	}
	if got, err := os.ReadFile(otherFile); err != nil || string(got) != string(before) {
		t.Errorf("the other tree's mark must be untouched (read err %v)", err)
	}
}

// A file at this tree's path whose own tree_root names another tree is refused.
func TestClearRejectionOnApproval_MarkWhoseTreeRootDisagrees_IsRefused(t *testing.T) {
	store := t.TempDir()
	r := Rejection{AgentType: AgentPlanAuditor, SpecID: "SPEC-A-001", Cause: CauseNoReceiptCited, TreeRoot: filepath.Join(store, "elsewhere")}
	// Written at the store's own path (what a copied file looks like).
	if err := writeJSON(rejectionPathInKind(store, store, KindReceipt, r.AgentType, r.SpecID), &r); err != nil {
		t.Fatal(err)
	}
	if err := ClearRejectionOnApproval(store, store, approvalFor(AgentPlanAuditor, "SPEC-A-001")); !errors.Is(err, ErrMarkNotFound) {
		t.Errorf("want ErrMarkNotFound, got %v", err)
	}
	if !exists(markFile(store, store, KindReceipt, AgentPlanAuditor, "SPEC-A-001")) || exists(approvalLogPath(store)) {
		t.Errorf("a refused call must remove nothing and write nothing")
	}
}

// Only a receipt-kind mark is clearable: a served-kind mark for the same role and
// SPEC is neither removed nor reported as found.
func TestClearRejectionOnApproval_ServedKindMark_IsNotClearable(t *testing.T) {
	store := t.TempDir()
	writeMark(t, store, store, AgentPlanAuditor, "SPEC-A-001", CauseServedModelDrift, KindServed)
	err := ClearRejectionOnApproval(store, store, approvalFor(AgentPlanAuditor, "SPEC-A-001"))
	if !errors.Is(err, ErrMarkNotFound) {
		t.Errorf("want ErrMarkNotFound, got %v", err)
	}
	if !exists(markFile(store, store, KindServed, AgentPlanAuditor, "SPEC-A-001")) {
		t.Errorf("the served-kind mark must remain")
	}
}

// A tree that is not the store's own (a config-orphaned worktree) clears only its
// own tree-keyed mark, never the store's.
func TestClearRejectionOnApproval_TreeOtherThanTheStore_ClearsOnlyItsOwnMark(t *testing.T) {
	store := t.TempDir()
	tree := filepath.Join(store, "wt")
	writeMark(t, store, store, AgentPlanAuditor, "SPEC-A-001", CauseNoReceiptCited, "")
	writeMark(t, store, tree, AgentPlanAuditor, "SPEC-A-001", CauseReceiptUnknown, "")
	if err := ClearRejectionOnApproval(store, tree, approvalFor(AgentPlanAuditor, "SPEC-A-001")); err != nil {
		t.Fatal(err)
	}
	if exists(markFile(store, tree, KindReceipt, AgentPlanAuditor, "SPEC-A-001")) {
		t.Errorf("the tree's own mark must be removed")
	}
	if !exists(markFile(store, store, KindReceipt, AgentPlanAuditor, "SPEC-A-001")) {
		t.Errorf("the store's own mark must remain")
	}
}

// A mark keyed unknown-spec (an empty SPEC ID) is cleared by the empty SPEC ID.
func TestClearRejectionOnApproval_UnknownSpecMark(t *testing.T) {
	store := t.TempDir()
	writeMark(t, store, store, AgentPlanAuditor, "", CauseVerdictLineMissing, "")
	if err := ClearRejectionOnApproval(store, store, approvalFor(AgentPlanAuditor, "")); err != nil {
		t.Fatal(err)
	}
	if exists(markFile(store, store, KindReceipt, AgentPlanAuditor, "")) {
		t.Errorf("the unknown-spec mark must be removed")
	}
}

// An approval without an agent type or an operator is refused before anything is
// read or written.
func TestClearRejectionOnApproval_IncompleteApproval_IsRefused(t *testing.T) {
	store := t.TempDir()
	writeMark(t, store, store, AgentPlanAuditor, "SPEC-A-001", CauseNoReceiptCited, "")
	noOperator := approvalFor(AgentPlanAuditor, "SPEC-A-001")
	noOperator.Operator = "  "
	noAgent := approvalFor("", "SPEC-A-001")
	for name, a := range map[string]*Approval{"nil": nil, "no operator": noOperator, "no agent type": noAgent} {
		if err := ClearRejectionOnApproval(store, store, a); err == nil || errors.Is(err, ErrMarkNotFound) {
			t.Errorf("%s: want a validation error, got %v", name, err)
		}
	}
	if !exists(markFile(store, store, KindReceipt, AgentPlanAuditor, "SPEC-A-001")) || exists(approvalLogPath(store)) {
		t.Errorf("a refused call must remove nothing and write nothing")
	}
}

// The record carries a UTC time: the Now seam when the caller sets none, and a
// caller-supplied time normalized to UTC.
func TestClearRejectionOnApproval_ApprovedAtIsUTC(t *testing.T) {
	fixed := time.Date(2026, 11, 3, 4, 5, 6, 0, time.FixedZone("KST", 9*3600))
	saved := Now
	Now = func() time.Time { return fixed }
	t.Cleanup(func() { Now = saved })

	store := t.TempDir()
	writeMark(t, store, store, AgentPlanAuditor, "SPEC-A-001", CauseNoReceiptCited, "")
	if err := ClearRejectionOnApproval(store, store, approvalFor(AgentPlanAuditor, "SPEC-A-001")); err != nil {
		t.Fatal(err)
	}
	if line := readApprovalLines(t, store)[0]; !strings.Contains(line, `"approved_at":"2026-11-02T19:05:06Z"`) {
		t.Errorf("approved_at must be the Now seam in UTC: %s", line)
	}
}
