// retired_handler.go: Backward compatibility stubs for retired agent handlers.
// SPEC-V3R3-RETIRED-AGENT-001 + SPEC-V3R3-RETIRED-DDD-001
//
// manager-ddd, manager-tdd, expert-debug, and expert-testing were retired.
// Their handler functions are preserved as stubs for backward compatibility
// with legacy user projects that have not run `moai update`.
package agents

import (
	"context"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// @MX:NOTE: [AUTO] Stub handlers for retired agents (manager-ddd, manager-tdd, expert-debug, expert-testing).
// These return no-op handlers that always allow. Preserved for backward compatibility.

// NewDDDHandler creates a stub handler for the retired manager-ddd agent.
// Use manager-develop (develop-* actions) instead.
func NewDDDHandler(action string) hook.Handler {
	event := hook.EventPreToolUse // default
	switch action {
	case "post-transformation":
		event = hook.EventPostToolUse
	case "completion":
		event = hook.EventSubagentStop
	}
	return &stubHandler{
		baseHandler: baseHandler{
			action: action,
			event:  event,
			agent:  "ddd",
		},
	}
}

// NewTDDHandler creates a stub handler for the retired manager-tdd agent.
// Use manager-develop (develop-* actions) instead.
func NewTDDHandler(action string) hook.Handler {
	event := hook.EventPreToolUse // default
	switch action {
	case "post-implementation":
		event = hook.EventPostToolUse
	case "completion":
		event = hook.EventSubagentStop
	}
	return &stubHandler{
		baseHandler: baseHandler{
			action: action,
			event:  event,
			agent:  "tdd",
		},
	}
}

// NewDebugHandler creates a stub handler for the retired expert-debug agent.
// Use manager-quality instead.
func NewDebugHandler(action string) hook.Handler {
	event := hook.EventPostToolUse // default
	switch action {
	case "completion":
		event = hook.EventSubagentStop
	}
	return &stubHandler{
		baseHandler: baseHandler{
			action: action,
			event:  event,
			agent:  "debug",
		},
	}
}

// NewTestingHandler creates a stub handler for the retired expert-testing agent.
// Use manager-develop instead.
func NewTestingHandler(action string) hook.Handler {
	event := hook.EventPostToolUse // default
	switch action {
	case "completion":
		event = hook.EventSubagentStop
	}
	return &stubHandler{
		baseHandler: baseHandler{
			action: action,
			event:  event,
			agent:  "testing",
		},
	}
}

// stubHandler is a no-op handler for retired agents.
type stubHandler struct {
	baseHandler
}

// Handle always allows for retired agent stubs.
func (h *stubHandler) Handle(ctx context.Context, input *hook.HookInput) (*hook.HookOutput, error) {
	return hook.NewAllowOutput(), nil
}

func (h *stubHandler) EventType() hook.EventType {
	return h.event
}
