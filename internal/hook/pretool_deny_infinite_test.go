package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/goal"
	"github.com/modu-ai/moai-adk/internal/hook/security"
)

// TestAC010_DenyAskHoldsUnderArmedInfiniteGoal asserts the deny/ask rules
// (PreToolUse security policy) bind IDENTICALLY whether or not an infinite goal
// is armed. An armed goal decides only continue/stop; it NEVER weakens the deny
// surface. SPEC-INFINITE-GOAL-001 AC-010 (cross-ref SPEC-STOPCHAIN-TRIM-001
// REQ-007/AC-006). Minimal local deny fixture (the sibling SPEC's fixtures live
// in its plan/acceptance prose, not a reusable Go corpus — per acceptance.md §
// D5 fixture-source decision: minimal-local when no reusable corpus exists).
func TestAC010_DenyAskHoldsUnderArmedInfiniteGoal(t *testing.T) {
	policy := DefaultSecurityPolicy()
	// Merge any extra patterns from security.yaml (mirrors production wiring).
	policy.MergeExtraPatterns(security.LoadExtraSecurityConfig(""))

	denyFixtures := []string{
		"git push --force origin main", // matches git\s+push\s+.*--force\s+origin\s+(main|master)
		"rm -rf /",                     // matches the rm -rf dangerous pattern
		"mkfs.ext4 /dev/sda",           // matches mkfs\. dangerous pattern
	}
	askFixtures := []string{
		"git reset --hard", // matches git\s+reset\s+--hard ask pattern
		"git clean -fd",    // matches git\s+clean\s+-fd ask pattern
	}

	h := &preToolHandler{policy: policy}

	check := func(t *testing.T, cmd, wantDecision string) {
		t.Helper()
		raw, _ := json.Marshal(map[string]any{"command": cmd})
		got, _ := h.checkBashCommand(raw)
		if got != wantDecision {
			t.Errorf("deny/ask regression: cmd=%q want=%q got=%q (an armed-infinite goal must not weaken the deny/ask decision)", cmd, wantDecision, got)
		}
	}

	// Cell 1: no goal armed.
	for _, c := range denyFixtures {
		check(t, c, DecisionDeny)
	}
	for _, c := range askFixtures {
		check(t, c, DecisionAsk)
	}

	// Arm an infinite goal in a project state dir and re-check. The deny/ask
	// decision is a pure function of (command, policy) — the PreToolUse handler
	// does NOT read the goal state — so the decisions must be identical under an
	// armed infinite goal.
	tmp := t.TempDir()
	g := goal.NewGoal("ac010", "infinite exits 0", []goal.Condition{
		{Type: goal.ConditionMechanical, Cmd: "true", ExpectExit: 0},
	})
	g.Ceiling.MaxTurns = 0
	g.Ceiling.MaxDuration = 3600
	g.Status = goal.StatusArmed
	if err := os.MkdirAll(filepath.Join(tmp, goal.StateDir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := goal.SaveGoal(tmp, g); err != nil {
		t.Fatal(err)
	}

	// Cell 2: armed infinite goal — decisions identical.
	for _, c := range denyFixtures {
		check(t, c, DecisionDeny)
	}
	for _, c := range askFixtures {
		check(t, c, DecisionAsk)
	}
}
