package homestate

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/runtime"
)

// TestCardTransitionCeilingRefusal is AC-ACE-015's card-transition seam arm
// (SPEC-AUDIT-CEILING-001 REQ-ACE-003): a ceiling-hit round whose verdict
// fails admission refuses the card transition — the refusal reason names the
// ceiling outcome and the outcome persists to the audit trail, so a mutant
// that skips the engine wiring cannot pass either arm. The control arm — an
// admission-clean verdict at the same ceiling state — admits
// (REQ-ACE-013's pass-through).
func TestCardTransitionCeilingRefusal(t *testing.T) {
	const specID = "SPEC-ACE-CEIL-001"
	const cardID = "t9002"

	// Each arm gets its own fixture: the trail is machine-local state and
	// the control arm asserts its own record, not the absence of a prior
	// subtest's.
	t.Run("failing_verdict_refuses_at_ceiling", func(t *testing.T) {
		f := newCeilingCardFixture(t, specID, cardID)
		path := f.writeVerdict(t, "FAIL", 1)
		ok, reason := admitCardVerdict(f.card, path)
		if ok {
			t.Fatal("card transition admitted a ceiling-hit failing round")
		}
		if !strings.Contains(reason, "ceiling refusal") {
			t.Fatalf("reason %q does not name the ceiling refusal", reason)
		}
		trail, err := os.ReadFile(filepath.Join(f.root, ".moai", "state", "audit-enforcement.log"))
		if err != nil {
			t.Fatalf("ceiling trail not written: %v", err)
		}
		if !strings.Contains(string(trail), "ceiling-refusal") || !strings.Contains(string(trail), specID) {
			t.Fatalf("trail line incomplete: %s", trail)
		}
	})

	t.Run("clean_verdict_passes_through", func(t *testing.T) {
		f := newCeilingCardFixture(t, specID, cardID)
		path := f.writeVerdict(t, "PASS", 1)
		if ok, reason := admitCardVerdict(f.card, path); !ok {
			t.Fatalf("control arm refused: %s", reason)
		}
		if trail, err := os.ReadFile(filepath.Join(f.root, ".moai", "state", "audit-enforcement.log")); err == nil &&
			strings.Contains(string(trail), "ceiling-refusal") {
			t.Fatalf("pass-through wrote a refusal record: %s", trail)
		}
	})
}

// ceilingCardFixture is the card-transition ceiling fixture: a worktree
// whose SPEC directory, harness config (Tier L ceiling 1, no delta rounds),
// and card report directory drive the engine.
type ceilingCardFixture struct {
	root string
	card Card
}

func newCeilingCardFixture(t *testing.T, specID, cardID string) *ceilingCardFixture {
	f := &ceilingCardFixture{root: t.TempDir()}
	specDir := filepath.Join(f.root, ".moai", "specs", specID)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("---\ntier: L\n---\n# spec\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "plan.md"), []byte("# plan\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfgDir := filepath.Join(f.root, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	harness := "harness:\n  evaluator:\n    memory_scope: per_iteration\n  plan_audit_tier_ceilings:\n    S: 1\n    M: 2\n    L: 1\n  plan_audit_ceiling_policy:\n    auto_delta_rounds: 0\n    on_final_hit: hold-and-split\n"
	if err := os.WriteFile(filepath.Join(cfgDir, "harness.yaml"), []byte(harness), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(f.root, ".moai", "reports", cardID), 0o755); err != nil {
		t.Fatal(err)
	}
	f.card = Card{State: CardPlanAudit, SpecID: specID, CardID: cardID, WorktreePath: f.root}
	return f
}

func (f *ceilingCardFixture) writeVerdict(t *testing.T, verdict string, iter int) string {
	t.Helper()
	h, err := runtime.NewInMemoryCache().ComputeHash(filepath.Join(f.root, ".moai", "specs", f.card.SpecID))
	if err != nil {
		t.Fatal(err)
	}
	body := "# SPEC Review Report: " + f.card.SpecID + "\nverdict: " + verdict + "\nOverall Score: 0.90\nmust_pass_failed: 0\nblocking_count: 0\nplan_artifact_hash: " + h + "\n"
	path := filepath.Join(f.root, ".moai", "reports", f.card.CardID, "plan-audit-iter"+strconv.Itoa(iter)+".md")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
