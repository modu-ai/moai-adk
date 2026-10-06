package homestate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/runtime"
)

// TestCardTransitionRequiredBackendRefusal is AC-ACE-022's card-transition
// seam arm (SPEC-AUDIT-CEILING-001 REQ-ACE-009): admitVerdictFile resolves the
// card worktree's configured gate set and refuses a verdict carrying a
// required-backend fail receipt although the auditor's own label, score,
// must-pass, blocking, and hash all pass. The control arm — the same verdict
// with a pass receipt — admits, proving the refusal comes from the receipt
// check and not from the mere presence of the configuration.
func TestCardTransitionRequiredBackendRefusal(t *testing.T) {
	const specID = "SPEC-ACE-TEST-001"
	root := t.TempDir()
	specDir := filepath.Join(root, ".moai", "specs", specID)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("---\ntier: L\n---\n# spec\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfgDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "workflow.yaml"), []byte("workflow:\n  audit:\n    gates:\n      claude: required\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	hash, err := runtime.NewInMemoryCache().ComputeHash(specDir)
	if err != nil {
		t.Fatal(err)
	}
	writeVerdict := func(receiptLine string) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), "plan-audit.md")
		raw := "# plan-audit\nverdict: PASS\nOverall Score: 0.90\nmust_pass_failed: 0\nblocking_count: 0\nplan_artifact_hash: " + hash + "\nconvergence_overall: pass\n" + receiptLine + "\n"
		if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	card := Card{State: CardPlanAudit, SpecID: specID, WorktreePath: root}

	t.Run("fail_receipt_refuses", func(t *testing.T) {
		path := writeVerdict("required_backend: claude fail")
		ok, reason := admitCardVerdict(card, path)
		if ok {
			t.Fatal("card transition admitted a required-backend fail receipt")
		}
		if !strings.Contains(reason, "claude") {
			t.Fatalf("reason %q does not name the backend", reason)
		}
	})

	t.Run("pass_receipt_admits", func(t *testing.T) {
		path := writeVerdict("required_backend: claude pass")
		if ok, reason := admitCardVerdict(card, path); !ok {
			t.Fatalf("control arm refused: %s", reason)
		}
	})

	// Card-review F7: a BELOW-ceiling required-backend refusal is still
	// recorded to the audit trail (REQ-ACE-007/012 — recording is
	// independent of the ceiling state).
	t.Run("below_ceiling_refusal_recorded", func(t *testing.T) {
		root2 := t.TempDir()
		specDir2 := filepath.Join(root2, ".moai", "specs", specID)
		if err := os.MkdirAll(specDir2, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(specDir2, "spec.md"), []byte("---\ntier: L\n---\n# spec\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(specDir2, "plan.md"), []byte("# plan\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		cfgDir2 := filepath.Join(root2, ".moai", "config", "sections")
		if err := os.MkdirAll(cfgDir2, 0o755); err != nil {
			t.Fatal(err)
		}
		harness := "harness:\n  evaluator:\n    memory_scope: per_iteration\n  plan_audit_tier_ceilings:\n    S: 1\n    M: 2\n    L: 3\n  plan_audit_ceiling_policy:\n    auto_delta_rounds: 0\n    on_final_hit: hold-and-split\n"
		if err := os.WriteFile(filepath.Join(cfgDir2, "harness.yaml"), []byte(harness), 0o600); err != nil {
			t.Fatal(err)
		}
		wf := "workflow:\n  audit:\n    gates:\n      claude: required\n"
		if err := os.WriteFile(filepath.Join(cfgDir2, "workflow.yaml"), []byte(wf), 0o600); err != nil {
			t.Fatal(err)
		}
		reportsDir := filepath.Join(root2, ".moai", "reports", "t9003")
		if err := os.MkdirAll(reportsDir, 0o755); err != nil {
			t.Fatal(err)
		}
		hash2, err := runtime.NewInMemoryCache().ComputeHash(specDir2)
		if err != nil {
			t.Fatal(err)
		}
		body := "# SPEC Review Report: " + specID + "\nverdict: PASS\nOverall Score: 0.90\nmust_pass_failed: 0\nblocking_count: 0\nplan_artifact_hash: " + hash2 + "\nconvergence_overall: pass\nrequired_backend: claude fail\n"
		path := filepath.Join(reportsDir, "plan-audit-iter1.md")
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		card2 := Card{State: CardPlanAudit, SpecID: specID, CardID: "t9003", WorktreePath: root2}
		ok, reason := admitCardVerdict(card2, path)
		if ok {
			t.Fatal("below-ceiling receipt refusal admitted")
		}
		if !strings.Contains(reason, "claude") {
			t.Fatalf("reason %q does not name the backend", reason)
		}
		trail, err := os.ReadFile(filepath.Join(root2, ".moai", "state", "audit-enforcement.log"))
		if err != nil {
			t.Fatalf("required-backend refusal not recorded: %v", err)
		}
		if !strings.Contains(string(trail), "required-backend-refusal") || !strings.Contains(string(trail), specID) {
			t.Fatalf("trail line incomplete: %s", trail)
		}
	})
}
