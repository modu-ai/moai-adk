// cycle_handler.go: manager-cycle unified DDD/TDD lifecycle hook handling.
// SPEC-V3R3-RETIRED-AGENT-001 D-EVAL-01 fix: added `case "cycle":` to factory dispatch.
// REQ-RA-009 acceptance criterion (factory dispatch).
package agents

import (
"context"

"github.com/modu-ai/moai-adk/internal/hook"
)

// @MX:NOTE: [AUTO] cycleHandler is the unified DDD/TDD agent (manager-cycle) lifecycle hook dispatcher.
// Maps cycle-pre-implementation / cycle-post-implementation / cycle-completion actions to
// PreToolUse / PostToolUse / SubagentStop events. Currently pass-through (default allow);
// subsequent SPEC can add cycle-specific validation (e.g., RED phase test existence enforcement).
//
// cycleHandler handles unified DDD/TDD workflow hooks for the manager-cycle agent.
// SPEC-V3R2-ORC-001 retirement decision integrated manager-tdd / manager-ddd with manager-cycle,
// and this handler processes cycle-* actions.
type cycleHandler struct {
baseHandler
}

// NewCycleHandler creates a manager-cycle handler for the given action.
// Actions: pre-implementation, post-implementation, completion
//
// SubagentStart is also handled by agentStartHandler in internal/hook/subagent_start.go
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
event: event,
agent: "cycle",
},
}
}

// Handle handles cycle workflow hooks.
// Currently pass-through (default allow). Subsequent SPEC can add cycle-specific validation.
func (h *cycleHandler) Handle(ctx context.Context, input *hook.HookInput) (*hook.HookOutput, error) {
// pre-implementation: RED/ANALYZE phase pre-validation
// post-implementation: GREEN/PRESERVE/IMPROVE/REFACTOR phase verification
// completion: cycle workflow completion report
return hook.NewAllowOutput(), nil
}

func (h *cycleHandler) EventType() hook.EventType {
return h.event
}
