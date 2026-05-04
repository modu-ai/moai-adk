// cycle_handler.go: manager-cycle agent의 unified DDD/TDD lifecycle hook 처리.
// SPEC-V3R3-RETIRED-AGENT-001 D-EVAL-01 fix: factory dispatch에 `case "cycle":` 추가.
// REQ-RA-009 acceptance criterion (factory dispatch).
package agents

import (
	"context"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// @MX:NOTE: [AUTO] cycleHandler는 SPEC-V3R2-ORC-001의 unified DDD/TDD agent
// (manager-cycle) lifecycle hook dispatcher다. cycle-pre-implementation /
// cycle-post-implementation / cycle-completion action을 PreToolUse / PostToolUse /
// SubagentStop event에 매핑한다. 현재는 pass-through (default allow);
// 후속 SPEC에서 cycle-specific 검증 (e.g., RED phase test 존재 강제) 추가 가능.
//
// cycleHandler는 manager-cycle agent의 unified DDD/TDD workflow hook을 처리한다.
// SPEC-V3R2-ORC-001 retirement decision으로 manager-tdd / manager-ddd가 manager-cycle로
// 통합되었으며, 본 핸들러는 cycle-* action을 처리한다.
type cycleHandler struct {
	baseHandler
}

// NewCycleHandler는 주어진 action에 대한 manager-cycle 핸들러를 생성한다.
// Actions: pre-implementation, post-implementation, completion
//
// SubagentStart는 별도 internal/hook/subagent_start.go의 agentStartHandler에서 처리한다
// (REQ-RA-007 retired-rejection guard).
func NewCycleHandler(action string) hook.Handler {
	event := hook.EventPreToolUse
	switch action {
	case "post-implementation":
		event = hook.EventPostToolUse
	case "completion":
		event = hook.EventSubagentStop
	}

	return &cycleHandler{
		baseHandler: baseHandler{
			action: action,
			event:  event,
			agent:  "cycle",
		},
	}
}

// Handle는 cycle workflow hook을 처리한다.
// 현재는 pass-through (default allow). 후속 SPEC에서 cycle-specific 검증 추가 가능.
func (h *cycleHandler) Handle(ctx context.Context, input *hook.HookInput) (*hook.HookOutput, error) {
	// pre-implementation: RED/ANALYZE phase pre-validation
	// post-implementation: GREEN/PRESERVE/IMPROVE/REFACTOR phase verification
	// completion: cycle workflow 완료 보고
	return hook.NewAllowOutput(), nil
}

func (h *cycleHandler) EventType() hook.EventType {
	return h.event
}
