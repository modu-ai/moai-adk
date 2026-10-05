package auditreceipt

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRejectionKind_StoredApartAndClearedByKind(t *testing.T) {
	root := t.TempDir()
	legacy := Rejection{AgentType: AgentPlanAuditor, SpecID: "SPEC-X-001", Cause: CauseNoReceiptCited}
	if err := WriteRejection(root, &legacy); err != nil {
		t.Fatalf("WriteRejection: %v", err)
	}
	served := Rejection{AgentType: AgentPlanAuditor, SpecID: "SPEC-X-001", Kind: KindServed, Cause: CauseServedModelDrift, TreeRoot: root}
	if err := WriteRejectionIn(root, &served); err != nil {
		t.Fatalf("WriteRejectionIn: %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(StateDir(root), rejectionsRel))
	if err != nil || len(entries) != 2 {
		t.Fatalf("record files = %d (err %v), want 2 — the kinds overwrote each other", len(entries), err)
	}

	got, err := ReadRejectionInKind(root, root, KindServed, AgentPlanAuditor, "SPEC-X-001")
	if err != nil || got.Kind != KindServed {
		t.Fatalf("ReadRejectionInKind(served) = %+v, %v", got, err)
	}
	legacyRead, err := ReadRejectionInKind(root, root, KindReceipt, AgentPlanAuditor, "SPEC-X-001")
	if err != nil || RejectionKind(legacyRead) != KindReceipt {
		t.Fatalf("a kind-less record must read as receipt kind: %+v, %v", legacyRead, err)
	}

	if err := ClearRejectionsForRoleInTreeKind(root, root, AgentPlanAuditor, KindServed); err != nil {
		t.Fatalf("clear served: %v", err)
	}
	left, err := ListRejectionsForTree(root, root)
	if err != nil || len(left) != 1 || RejectionKind(left[0]) != KindReceipt {
		t.Fatalf("after clearing served: %+v (err %v), want only the receipt record", left, err)
	}

	if err := WriteRejectionIn(root, &served); err != nil {
		t.Fatalf("WriteRejectionIn: %v", err)
	}
	if err := ClearRejectionsForRoleInTreeKind(root, root, AgentPlanAuditor, KindReceipt); err != nil {
		t.Fatalf("clear receipt: %v", err)
	}
	left, err = ListRejectionsForTree(root, root)
	if err != nil || len(left) != 1 || RejectionKind(left[0]) != KindServed {
		t.Fatalf("after clearing receipt: %+v (err %v), want only the served record", left, err)
	}
}

func TestServedGateEnabled_ReadsOnlyAnExplicitTrue(t *testing.T) {
	cases := []struct {
		name string
		body string
		want bool
	}{
		{"explicit_true", "workflow:\n  served_model_gate:\n    enabled: true\n", true},
		{"explicit_false", "workflow:\n  served_model_gate:\n    enabled: false\n", false},
		{"absent_key", "workflow:\n  audit:\n    gates:\n      codex: required\n", false},
		{"malformed_yaml", "workflow: [\n", false},
		{"no_file", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if tc.body != "" {
				dir := filepath.Join(root, ".moai", "config", "sections")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(tc.body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if got := ServedGateEnabled(root); got != tc.want {
				t.Fatalf("ServedGateEnabled = %v, want %v", got, tc.want)
			}
		})
	}
	if ServedGateEnabled("") {
		t.Fatal("an empty root must read as disabled")
	}
}
