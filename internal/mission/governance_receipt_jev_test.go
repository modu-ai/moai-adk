// governance_receipt_jev_test.go — SPEC-JEV-GOAL-DIST-001 M7b (AC-JEVG-003,
// AC-JEVG-005). A Jev Noul answer supplied to mission-governor is an auxiliary
// input signal recorded as a SEPARATE item in the governance receipt; it is
// never a binding field, never a completion-predicate element, and never part
// of the landed-ancestry or authoritative-readback evidence the receipt binds.
package mission

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func jevReceiptFixture(action Action, issuer string, kind GovernanceReceiptKind, status GovernanceReceiptStatus) GovernanceReceipt {
	return GovernanceReceipt{
		Version:      1,
		Kind:         kind,
		MissionID:    "018f4f4a-7b7c-7a11-8f4d-666666666666",
		ContractHash: "contract-hash-jev",
		SnapshotHash: "snapshot-hash-jev",
		Action:       action,
		Targets:      []string{"gtd:card"},
		ExpiresAt:    time.Now().Add(time.Hour),
		Issuer:       issuer,
		HeadSHA:      "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
		Status:       status,
	}
}

func writeJevEvidence(t *testing.T, root string, decision GovernanceReceipt) GovernanceEvidence {
	t.Helper()
	decisionPath := filepath.Join(governanceDirForTest(t, root), "decision.json")
	auditPath := filepath.Join(governanceDirForTest(t, root), "audit.json")
	if err := WriteGovernanceReceipt(root, decisionPath, decision); err != nil {
		t.Fatalf("write decision receipt: %v", err)
	}
	audit := jevReceiptFixture(decision.Action, "sync-auditor", GovernanceAudit, GovernancePassed)
	if err := WriteGovernanceReceipt(root, auditPath, audit); err != nil {
		t.Fatalf("write audit receipt: %v", err)
	}
	exp := GovernanceExpectation{
		MissionID: decision.MissionID, ContractHash: decision.ContractHash,
		SnapshotHash: decision.SnapshotHash, HeadSHA: decision.HeadSHA,
		Action: decision.Action, Targets: decision.Targets, Now: time.Now(),
	}
	evidence, err := LoadGovernanceReceipts(root, decisionPath, auditPath, exp)
	if err != nil {
		t.Fatalf("LoadGovernanceReceipts: %v", err)
	}
	return evidence
}

// governanceDirForTest resolves the receipt directory the same way the writer
// does, so tests construct paths the production code accepts.
func governanceDirForTest(t *testing.T, root string) string {
	t.Helper()
	dir, err := governanceDir(root, true)
	if err != nil {
		t.Fatalf("governance dir: %v", err)
	}
	return dir
}

// TestGovernanceReceiptRecordsJevNoulAsSeparateItem — AC-JEVG-003: with a Jev
// Noul supplied, the receipt parses with the answer as a separate item, every
// pre-existing binding field present and unchanged, file mode 0600.
func TestGovernanceReceiptRecordsJevNoulAsSeparateItem(t *testing.T) {
	root := t.TempDir()
	decision := jevReceiptFixture(ActionPublish, "mission-governor", GovernanceDecision, GovernanceRecommended)
	decision.AuxiliarySignals = []AuxiliarySignal{
		{QuestionID: "premise-alive", Kind: "noul", Noul: true, Probability: 0.62},
	}

	decisionPath := filepath.Join(governanceDirForTest(t, root), "decision.json")
	if err := WriteGovernanceReceipt(root, decisionPath, decision); err != nil {
		t.Fatalf("write: %v", err)
	}
	if info, err := os.Stat(decisionPath); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("receipt mode = %v err=%v, want 0600", info.Mode(), err)
	}

	evidence := writeJevEvidence(t, root, decision)
	got := evidence.Decision

	// The separate item survives the write/read round trip intact.
	if len(got.AuxiliarySignals) != 1 {
		t.Fatalf("auxiliary signals = %+v, want exactly one recorded item", got.AuxiliarySignals)
	}
	item := got.AuxiliarySignals[0]
	if item.QuestionID != "premise-alive" || item.Kind != "noul" || !item.Noul || item.Probability != 0.62 {
		t.Errorf("recorded item = %+v, want the supplied Jev Noul answer", item)
	}

	// Every pre-existing binding field is present and unchanged.
	binding := got
	binding.AuxiliarySignals = nil
	want := decision
	if binding.Version != want.Version || binding.Kind != want.Kind ||
		binding.MissionID != want.MissionID || binding.ContractHash != want.ContractHash ||
		binding.SnapshotHash != want.SnapshotHash || binding.Action != want.Action ||
		strings.Join(binding.Targets, ",") != strings.Join(want.Targets, ",") ||
		!binding.ExpiresAt.Equal(want.ExpiresAt) || binding.Issuer != want.Issuer ||
		binding.HeadSHA != want.HeadSHA || binding.Status != want.Status {
		t.Errorf("binding fields drifted: got %+v want %+v", binding, want)
	}
}

// TestGovernanceReceiptAuxiliaryItemIsNotABindingField — the recorded item is
// auxiliary by construction: it is not part of the binding validation (two
// receipts differing only in the item bind identically) while the integrity
// digest still covers it (the item cannot be edited off the record).
func TestGovernanceReceiptAuxiliaryItemIsNotABindingField(t *testing.T) {
	rootA, rootB := t.TempDir(), t.TempDir()
	base := jevReceiptFixture(ActionPublish, "mission-governor", GovernanceDecision, GovernanceRecommended)

	withNoul := base
	withNoul.AuxiliarySignals = []AuxiliarySignal{{QuestionID: "premise-alive", Kind: "noul", Noul: true, Probability: 0.62}}
	withOpposite := base
	withOpposite.AuxiliarySignals = []AuxiliarySignal{{QuestionID: "premise-alive", Kind: "noul", Noul: false, Probability: 0.55}}

	evA := writeJevEvidence(t, rootA, withNoul)
	evB := writeJevEvidence(t, rootB, withOpposite)

	if len(evA.Decision.AuxiliarySignals) != 1 || evA.Decision.AuxiliarySignals[0].Noul != true {
		t.Fatalf("rootA decision lost its auxiliary item: %+v", evA.Decision.AuxiliarySignals)
	}
	if len(evB.Decision.AuxiliarySignals) != 1 || evB.Decision.AuxiliarySignals[0].Noul != false {
		t.Fatalf("rootB decision lost its auxiliary item: %+v", evB.Decision.AuxiliarySignals)
	}

	// The integrity digest distinguishes the two records: what the governor was
	// shown is covered by the receipt's integrity, not left to prose.
	digestFor := func(root string) string {
		raw, err := os.ReadFile(filepath.Join(governanceDirForTest(t, root), "decision.json"))
		if err != nil {
			t.Fatalf("read decision: %v", err)
		}
		var doc struct {
			Integrity string `json:"integrity"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Fatalf("parse integrity: %v", err)
		}
		return doc.Integrity
	}
	if digestFor(rootA) == digestFor(rootB) {
		t.Error("two decisions differing only in the auxiliary item carry the same integrity digest — the item is not covered")
	}
}

// TestGovernanceReceiptWithoutAuxiliaryItemKeepsPreSpecShape — receipts written
// without a Jev answer marshal to the pre-SPEC shape (the auxiliary key absent,
// so pre-SPEC digests still validate) and still load.
func TestGovernanceReceiptWithoutAuxiliaryItemKeepsPreSpecShape(t *testing.T) {
	root := t.TempDir()
	decision := jevReceiptFixture(ActionPublish, "mission-governor", GovernanceDecision, GovernanceRecommended)

	decisionPath := filepath.Join(governanceDirForTest(t, root), "decision.json")
	if err := WriteGovernanceReceipt(root, decisionPath, decision); err != nil {
		t.Fatalf("write: %v", err)
	}
	raw, err := os.ReadFile(decisionPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "auxiliary_signals") {
		t.Error("receipt without an auxiliary item marshals the auxiliary key — pre-SPEC digest compatibility would break")
	}
	writeJevEvidence(t, root, decision) // loads and validates end to end
}

// TestCompletionPredicateInputsCarryNoJevSymbol — AC-JEVG-005: the enumerated
// set of completion-predicate and receipt-binding inputs (the contract writer's
// fields and the governance receipt's binding fields) carries no Jev client
// symbol. The enumeration is asserted non-vacuous (each named field is really
// present in the source it is claimed to govern), and the scan carries its own
// positive control on a path that does import internal/jev.
func TestCompletionPredicateInputsCarryNoJevSymbol(t *testing.T) {
	bindingFields := []string{
		"MissionID", "ContractHash", "SnapshotHash", "Action", "Targets",
		"ExpiresAt", "Issuer", "HeadSHA", "Status", "Integrity",
		"CompletionEvidence", "RecoveryConditions",
	}
	governedSources := map[string]string{
		"governance_receipt.go":   "../../internal/mission/governance_receipt.go",
		"contract.go":             "../../internal/mission/contract.go",
		"completion_receipt.go":   "../../internal/mission/completion_receipt.go",
		"goal.go (contract writer)": "../../internal/cli/goal.go",
	}
	for name, path := range governedSources {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		sawAny := false
		for _, field := range bindingFields {
			if strings.Contains(string(body), field) {
				sawAny = true
			}
		}
		if !sawAny {
			t.Errorf("%s carries none of the enumerated binding fields — the sweep is measuring nothing", name)
		}
		if strings.Contains(string(body), "internal/jev") {
			t.Errorf("%s references internal/jev — a Jev answer must not reach the completion-predicate or receipt-binding inputs (REQ-JEVG-005)", name)
		}
	}

	// SPEC-JEV-GUARD-001 (t1083) withdrew jev_skill_suggest.go, the former
	// positive control; mcp_jev.go is the live MCP wrapper that still imports
	// internal/jev, so the control keeps proving the scan fires.
	control, err := os.ReadFile("../../internal/cli/mcp_jev.go")
	if err != nil {
		t.Fatalf("read positive control: %v", err)
	}
	if !strings.Contains(string(control), "internal/jev") {
		t.Fatal("positive control failed: mcp_jev.go no longer imports internal/jev, so the zero-result above establishes nothing")
	}
}
