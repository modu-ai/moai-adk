// cycle_handler.go: manager-develop unified DDD/TDD lifecycle hook handling.
// Originally manager-cycle (SPEC-V3R3-RETIRED-AGENT-001); renamed to manager-develop
// per ORC-001 follow-up rename.
// REQ-RA-009 acceptance criterion (factory dispatch).
package agents

import (
"context"

"github.com/modu-ai/moai-adk/internal/hook"
)

// @MX:NOTE: [AUTO] developHandler is the unified DDD/TDD agent (manager-develop) lifecycle hook dispatcher.
// Maps develop-pre-implementation / develop-post-implementation / develop-completion actions to
// PreToolUse / PostToolUse / SubagentStop events. Currently pass-through (default allow);
// subsequent SPEC can add develop-specific validation (e.g., RED phase test existence enforcement).
//
// developHandler handles unified DDD/TDD workflow hooks for the manager-develop agent.
// Originally manager-cycle (SPEC-V3R2-ORC-001); ORC-001 follow-up rename changed canonical
// name to manager-develop. This handler processes develop-* actions.
type developHandler struct {
baseHandler
}

// NewDevelopHandler creates a manager-develop handler for the given action.
// Actions: pre-implementation, post-implementation, completion
//
// SubagentStart is also handled by agentStartHandler in internal/hook/subagent_start.go
// (REQ-RA-007 retired-rejection guard).
func NewDevelopHandler(action string) hook.Handler {
event := hook.EventPreToolUse
switch action {
case "post-implementation":
event = hook.EventPostToolUse
case "completion":
event = hook.EventSubagentStop
}

return &developHandler{
baseHandler: baseHandler{
action: action,
event: event,
agent: "develop",
},
}
}

// NewCycleHandler is a backward-compatibility alias for NewDevelopHandler.
// Preserved for factory dispatch backward compat (legacy "cycle" case in factory.go).
func NewCycleHandler(action string) hook.Handler {
return NewDevelopHandler(action)
}

// Handle handles develop workflow hooks.
// Currently pass-through (default allow). Subsequent SPEC can add develop-specific validation.
func (h *developHandler) Handle(ctx context.Context, input *hook.HookInput) (*hook.HookOutput, error) {
// pre-implementation: RED/ANALYZE phase pre-validation
// post-implementation: GREEN/PRESERVE/IMPROVE/REFACTOR phase verification
// completion: develop workflow completion report
return hook.NewAllowOutput(), nil
}

func (h *developHandler) EventType() hook.EventType {
return h.event
}
