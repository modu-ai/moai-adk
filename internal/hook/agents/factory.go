package agents

import (
	"context"
	"fmt"
	"strings"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// Factory creates agent-specific hook handlers based on action.
type Factory struct{}

// NewFactory creates a new agent handler factory.
func NewFactory() *Factory {
	return &Factory{}
}

// @MX:NOTE: [AUTO] Switch branch that creates one of 12 handler types based on the agent name. Add a new case here when adding a new agent.
// Supported agents: ddd, tdd (legacy retired stub compat), develop, cycle (backward compat alias for develop), backend, frontend, testing, debug, devops, quality, spec, docs
// CreateHandler creates a handler for the given agent action.
// Action format: {agent}-{action}
// Examples: develop-pre-implementation, backend-validation, docs-completion
//
// SPEC-V3R3-RETIRED-AGENT-001 + SPEC-V3R3-RETIRED-DDD-001: develop handler dispatches
// manager-develop's unified DDD/TDD workflow hooks. manager-tdd + manager-ddd retired
// stubs use no hooks (frontmatter cleared) but `case "tdd":` and `case "ddd":` are
// preserved for backward compatibility with legacy user projects that have not run
// `moai update`. `case "cycle":` is preserved for backward compat with pre-rename projects.
func (f *Factory) CreateHandler(action string) (hook.Handler, error) {
	parts := strings.SplitN(action, "-", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid action format: %s (expected {agent}-{action})", action)
	}

	agent := parts[0]
	act := parts[1]

	switch agent {
	case "ddd":
		return NewDDDHandler(act), nil
	case "tdd":
		return NewTDDHandler(act), nil
	case "develop":
		return NewDevelopHandler(act), nil
	case "cycle":
		// Backward compat alias for develop (pre-rename projects)
		return NewDevelopHandler(act), nil
	case "backend":
		return NewBackendHandler(act), nil
	case "frontend":
		return NewFrontendHandler(act), nil
	case "testing":
		return NewTestingHandler(act), nil
	case "debug":
		return NewDebugHandler(act), nil
	case "devops":
		return NewDevOpsHandler(act), nil
	case "quality":
		return NewQualityHandler(act), nil
	case "spec":
		return NewSpecHandler(act), nil
	case "docs":
		return NewDocsHandler(act), nil
	default:
		return NewDefaultHandler(action), nil
	}
}

// baseHandler provides common functionality for all agent handlers.
type baseHandler struct {
	action string
	event  hook.EventType
	agent  string
}

// Handle logs the action and allows it by default.
// Subclasses should override this method to provide specific behavior.
func (h *baseHandler) Handle(ctx context.Context, input *hook.HookInput) (*hook.HookOutput, error) {
	// Log the agent hook action for debugging
	// In production, this would dispatch to the actual handler logic

	// For now, allow all actions by default
	return hook.NewAllowOutput(), nil
}

func (h *baseHandler) EventType() hook.EventType {
	return h.event
}
