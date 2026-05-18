package router

import (
	"log/slog"
	"sync"
)

// hardEscalationCeiling은 MaxEscalations의 절대 상한입니다.
// REQ-HRN-001-013: max_escalations hard ceiling = 3.
//
// @MX:WARN: [AUTO] FROZEN 상한 — 이 값을 변경하면 REQ-HRN-001-013 위반
// @MX:REASON: design-constitution §5 + SPEC-V3R2-HRN-001 REQ-013: hard ceiling = 3
const hardEscalationCeiling = 3

// EscalationContext는 에스컬레이션 판단에 필요한 컨텍스트입니다.
// REQ-HRN-001-004/009.
type EscalationContext struct {
	// CurrentLevel은 에스컬레이션 발동 전 현재 harness 레벨입니다.
	CurrentLevel Level
	// TriggerType은 에스컬레이션을 발동한 이벤트 유형입니다.
	// 예: "quality_gate_fail", "review_critical", "test_coverage_low".
	TriggerType string
}

// EscalationManager는 harness 레벨 에스컬레이션을 관리합니다.
// REQ-HRN-001-004: CheckTriggers(ctx) → (newLevel, escalated).
// REQ-HRN-001-013: MaxEscalations 상한 적용.
// REQ-HRN-001-018: 상한 초과 시 HRN_ESCALATION_CAP_REACHED 로그.
//
// @MX:ANCHOR: [AUTO] 에스컬레이션 카운터 + 레벨 범프 엔트리 포인트
// @MX:REASON: fan_in >= 3: CLI validate, unit tests, phase-result processor에서 사용
type EscalationManager struct {
	mu             sync.Mutex
	maxEscalations int
	count          int
}

// NewEscalationManager는 새 EscalationManager를 반환합니다.
// maxEscalations가 hardEscalationCeiling(3)을 초과하면 ceiling으로 조정됩니다.
func NewEscalationManager(maxEscalations int) *EscalationManager {
	if maxEscalations > hardEscalationCeiling {
		maxEscalations = hardEscalationCeiling
	}
	if maxEscalations < 0 {
		maxEscalations = 0
	}
	return &EscalationManager{
		maxEscalations: maxEscalations,
	}
}

// CheckTriggers는 에스컬레이션 컨텍스트를 확인하여 레벨을 올립니다.
// REQ-HRN-001-009: 트리거 발동 시 레벨 한 단계 올림 (minimal→standard→thorough).
// REQ-HRN-001-013: 카운터가 MaxEscalations에 도달하면 에스컬레이션 차단.
// REQ-HRN-001-018: 상한 초과 시 HRN_ESCALATION_CAP_REACHED 로그 출력.
func (m *EscalationManager) CheckTriggers(ctx EscalationContext) (newLevel Level, escalated bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// thorough에서는 이미 최고 레벨 — 에스컬레이션 불필요
	if ctx.CurrentLevel == LevelThorough {
		return LevelThorough, false
	}

	// 카운터 상한 확인
	if m.count >= m.maxEscalations {
		slog.Warn("HRN_ESCALATION_CAP_REACHED: maximum escalation count exceeded",
			"max_escalations", m.maxEscalations,
			"current_count", m.count,
			"trigger", ctx.TriggerType,
			"current_level", ctx.CurrentLevel,
		)
		return ctx.CurrentLevel, false
	}

	// 레벨 한 단계 올림
	m.count++
	switch ctx.CurrentLevel {
	case LevelMinimal:
		return LevelStandard, true
	case LevelStandard:
		return LevelThorough, true
	default:
		return ctx.CurrentLevel, false
	}
}

// Count는 현재까지 에스컬레이션된 횟수를 반환합니다.
func (m *EscalationManager) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.count
}

// MaxEscalations는 설정된 최대 에스컬레이션 횟수를 반환합니다.
func (m *EscalationManager) MaxEscalations() int {
	return m.maxEscalations
}
