package codexwiring

import (
	"encoding/json"
	"testing"
	"time"
)

// renderedStopTimeout reads T_stop from the rendered hooks.json — the value
// Codex actually receives — rather than from the render constant, so the
// budget check follows any later change to moaiHandlerTimeout.
func renderedStopTimeout(t *testing.T) time.Duration {
	t.Helper()
	rendered, err := RenderHooks(nil)
	if err != nil {
		t.Fatalf("RenderHooks(nil): %v", err)
	}
	var doc struct {
		Hooks map[string][]entryJSON `json:"hooks"`
	}
	if err := json.Unmarshal(rendered, &doc); err != nil {
		t.Fatalf("parse rendered hooks.json: %v", err)
	}
	const want = "moai hook stop" + harnessCodexSuffix
	var found []int
	for _, entry := range doc.Hooks["Stop"] {
		for _, h := range entry.Hooks {
			if h.Command == want {
				found = append(found, h.Timeout)
			}
		}
	}
	if len(found) != 1 {
		t.Fatalf("want exactly one %q handler in the rendered Stop entry, got timeouts %v", want, found)
	}
	return time.Duration(found[0]) * time.Second
}

// TestStopChainAggregateBudgetFitsTimeout is AC-HPR-016's sum leg (design
// §D3.5): Σ(internal budgets of in-hook members) + chain_overhead ≤ T_stop,
// where a receipt member contributes only its compare budget, and T_stop does
// not exceed a measured T_codex_max. Until the AC-HPR-021 probe has run,
// T_codex_max is unmeasured and T_stop stays at the render constant.
func TestStopChainAggregateBudgetFitsTimeout(t *testing.T) {
	t.Parallel()
	tStop := renderedStopTimeout(t)

	if len(StopChainMembers) != 8 {
		t.Fatalf("want the 8 Claude Stop members declared, got %d", len(StopChainMembers))
	}
	var sum time.Duration
	for i, m := range StopChainMembers {
		if m.Number != i+1 {
			t.Errorf("member at index %d numbered %d, want %d", i, m.Number, i+1)
		}
		if m.Budget <= 0 {
			t.Errorf("member %d (%s) declares no internal budget", m.Number, m.Name)
		}
		if m.Placement != StopPlacementInHook && m.Placement != StopPlacementReceipt {
			t.Errorf("member %d (%s) has unknown placement %q", m.Number, m.Name, m.Placement)
		}
		if m.Budget+m.UncutBudget > tStop {
			t.Errorf("member %d (%s) budget %v exceeds T_stop %v (REQ-HPR-018)",
				m.Number, m.Name, m.Budget+m.UncutBudget, tStop)
		}
		sum += m.Budget + m.UncutBudget
	}
	if StopChainOverhead <= 0 {
		t.Fatalf("chain_overhead must be declared, got %v", StopChainOverhead)
	}
	if total := sum + StopChainOverhead; total > tStop {
		t.Fatalf("Σ member budgets %v + chain_overhead %v = %v exceeds T_stop %v",
			sum, StopChainOverhead, total, tStop)
	}

	if StopTimeoutCodexMax > 0 {
		if tStop > StopTimeoutCodexMax {
			t.Fatalf("T_stop %v exceeds the measured T_codex_max %v", tStop, StopTimeoutCodexMax)
		}
	} else if tStop != defaultHandlerTimeout*time.Second {
		t.Fatalf("T_codex_max is unmeasured, so T_stop must stay at the render constant %ds, got %v",
			defaultHandlerTimeout, tStop)
	}
}

// TestStopChainReceiptPlacement pins the design §D3.3 placement: the sync gate
// (member 2) and the codex review gate (member 6) run out of hook and are
// compared in-hook by receipt; every other member runs in-hook.
func TestStopChainReceiptPlacement(t *testing.T) {
	t.Parallel()
	for _, m := range StopChainMembers {
		wantReceipt := m.Number == 2 || m.Number == 6
		if got := m.Placement == StopPlacementReceipt; got != wantReceipt {
			t.Errorf("member %d (%s): receipt placement = %v, want %v", m.Number, m.Name, got, wantReceipt)
		}
		if m.UncutBudget > 0 && m.Number != 1 {
			t.Errorf("member %d (%s): only member 1's factory step is exempt from the cut-off", m.Number, m.Name)
		}
	}
}

// TestStopUnmeasuredCapDeclared pins the design §D3.8 cap declaration. The cap
// value is a proposal until M2d, but N < 2 would allow the stop on the first
// unmeasured continuation and make the continuation itself vacuous.
func TestStopUnmeasuredCapDeclared(t *testing.T) {
	t.Parallel()
	if StopUnmeasuredCap < 2 {
		t.Fatalf("StopUnmeasuredCap = %d; the first unmeasured Stop must continue, so N must be >= 2", StopUnmeasuredCap)
	}
	if StopCapStateDir == "" {
		t.Fatal("StopCapStateDir must name the session-scoped counter directory")
	}
}
