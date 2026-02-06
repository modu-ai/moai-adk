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

// CreateHandler creates a handler for the given agent action.
// Action format: {agent}-{action}
// Examples: ddd-pre-transformation, backend-validation, docs-completion
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
