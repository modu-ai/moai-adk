package router_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness/router"
)

// TestEscalationCapEnforcement — AC-HRN-001-08, REQ-HRN-001-009/013/018.
// When max_escalations=2, the 3rd escalation attempt returns false and emits the
// HRN_ESCALATION_CAP_REACHED log.
func TestEscalationCapEnforcement(t *testing.T) {
	t.Parallel()

	mgr := router.NewEscalationManager(2) // max_escalations: 2

	// 1st escalation: minimal → standard
	newLevel, escalated := mgr.CheckTriggers(router.EscalationContext{
		CurrentLevel: router.LevelMinimal,
		TriggerType:  "quality_gate_fail",
	})
	if !escalated {
		t.Error("1st escalation: escalated should be true")
	}
	if newLevel != router.LevelStandard {
		t.Errorf("1st escalation: newLevel got %q, want %q", newLevel, router.LevelStandard)
	}

	// 2nd escalation: standard → thorough
	newLevel, escalated = mgr.CheckTriggers(router.EscalationContext{
		CurrentLevel: router.LevelStandard,
		TriggerType:  "quality_gate_fail",
	})
	if !escalated {
		t.Error("2nd escalation: escalated should be true")
	}
	if newLevel != router.LevelThorough {
		t.Errorf("2nd escalation: newLevel got %q, want %q", newLevel, router.LevelThorough)
	}

	// 3rd escalation attempt: cap exceeded → escalated: false, level unchanged
	newLevel, escalated = mgr.CheckTriggers(router.EscalationContext{
		CurrentLevel: router.LevelThorough,
		TriggerType:  "quality_gate_fail",
	})
	if escalated {
		t.Error("3rd escalation (cap exceeded): escalated should be false")
	}
	if newLevel != router.LevelThorough {
		t.Errorf("3rd escalation (cap exceeded): newLevel got %q, want %q (should stay at thorough)", newLevel, router.LevelThorough)
	}
}

// TestEscalationTriggers — REQ-HRN-001-009: verify various trigger types.
func TestEscalationTriggers(t *testing.T) {
	t.Parallel()

	triggers := []string{"quality_gate_fail", "review_critical", "test_coverage_low"}

	for _, trigger := range triggers {
		t.Run(trigger, func(t *testing.T) {
			t.Parallel()
			mgr := router.NewEscalationManager(3)
			newLevel, escalated := mgr.CheckTriggers(router.EscalationContext{
				CurrentLevel: router.LevelMinimal,
				TriggerType:  trigger,
			})
			if !escalated {
				t.Errorf("trigger %q: escalated should be true", trigger)
			}
			if newLevel != router.LevelStandard {
				t.Errorf("trigger %q: newLevel got %q, want %q", trigger, newLevel, router.LevelStandard)
			}
		})
	}
}

// TestEscalationAtThorough — attempting to escalate at thorough keeps thorough even without a cap.
func TestEscalationAtThorough(t *testing.T) {
	t.Parallel()

	mgr := router.NewEscalationManager(5)
	newLevel, escalated := mgr.CheckTriggers(router.EscalationContext{
		CurrentLevel: router.LevelThorough,
		TriggerType:  "quality_gate_fail",
	})
	// Escalation at thorough — already at the max level
	if escalated {
		t.Error("escalation at thorough: already at max level, escalated should be false")
	}
	if newLevel != router.LevelThorough {
		t.Errorf("escalation at thorough: newLevel got %q, want thorough", newLevel)
	}
}

// TestEscalationMaxHardCeiling — REQ-HRN-001-013: hard ceiling = 3.
func TestEscalationMaxHardCeiling(t *testing.T) {
	t.Parallel()

	// Even with max_escalations > 3, the hard ceiling of 3 must apply
	mgr := router.NewEscalationManager(10)

	// 3 escalations
	for i := 0; i < 3; i++ {
		mgr.CheckTriggers(router.EscalationContext{
			CurrentLevel: router.LevelMinimal,
			TriggerType:  "quality_gate_fail",
		})
	}

	// 4th attempt: blocked by the hard ceiling
	_, escalated := mgr.CheckTriggers(router.EscalationContext{
		CurrentLevel: router.LevelMinimal,
		TriggerType:  "quality_gate_fail",
	})
	if escalated {
		t.Error("hard ceiling 3 exceeded: escalated should be false after 3rd escalation")
	}
}
