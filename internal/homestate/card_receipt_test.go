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
}
