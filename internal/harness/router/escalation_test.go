package router_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness/router"
)

// TestEscalationCapEnforcement — AC-HRN-001-08, REQ-HRN-001-009/013/018.
// max_escalations=2 설정 시 3번째 에스컬레이션 시도는 false를 반환하고
// HRN_ESCALATION_CAP_REACHED 로그를 출력합니다.
func TestEscalationCapEnforcement(t *testing.T) {
	t.Parallel()

	mgr := router.NewEscalationManager(2) // max_escalations: 2

	// 1번째 에스컬레이션: minimal → standard
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

	// 2번째 에스컬레이션: standard → thorough
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

	// 3번째 에스컬레이션 시도: cap 초과 → escalated: false, level 유지
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

// TestEscalationTriggers — REQ-HRN-001-009: 다양한 trigger 타입 검증.
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

// TestEscalationAtThorough — thorough에서 에스컬레이션 시도 시 cap 없어도 thorough 유지.
func TestEscalationAtThorough(t *testing.T) {
	t.Parallel()

	mgr := router.NewEscalationManager(5)
	newLevel, escalated := mgr.CheckTriggers(router.EscalationContext{
		CurrentLevel: router.LevelThorough,
		TriggerType:  "quality_gate_fail",
	})
	// thorough에서 에스컬레이션 — 이미 최고 레벨
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

	// max_escalations > 3을 시도해도 hard ceiling 3이 적용되어야 합니다
	mgr := router.NewEscalationManager(10)

	// 3번 에스컬레이션
	for i := 0; i < 3; i++ {
		mgr.CheckTriggers(router.EscalationContext{
			CurrentLevel: router.LevelMinimal,
			TriggerType:  "quality_gate_fail",
		})
	}

	// 4번째 시도: hard ceiling으로 차단
	_, escalated := mgr.CheckTriggers(router.EscalationContext{
		CurrentLevel: router.LevelMinimal,
		TriggerType:  "quality_gate_fail",
	})
	if escalated {
		t.Error("hard ceiling 3 exceeded: escalated should be false after 3rd escalation")
	}
}
