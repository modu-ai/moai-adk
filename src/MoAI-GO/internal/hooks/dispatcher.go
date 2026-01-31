package hooks

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/anthropics/moai-adk-go/internal/hooks/handlers"
	"github.com/anthropics/moai-adk-go/internal/hooks/protocol"
)

// HookDispatcher routes hook events to appropriate handlers
type HookDispatcher struct {
	sessionStartHandler *handlers.SessionStartHandler
	sessionEndHandler   *handlers.SessionEndHandler
	preToolHandler      *handlers.PreToolHandler
	postToolHandler     *handlers.PostToolHandler
}

// NewHookDispatcher creates a new hook dispatcher
func NewHookDispatcher() *HookDispatcher {
	return &HookDispatcher{
		sessionStartHandler: handlers.NewSessionStartHandler(),
		sessionEndHandler:   handlers.NewSessionEndHandler(),
		preToolHandler:      handlers.NewPreToolHandler(),
		postToolHandler:     handlers.NewPostToolHandler(),
	}
}

// Dispatch routes the hook event to the appropriate handler
func (d *HookDispatcher) Dispatch(ctx context.Context, input *protocol.HookInput) error {
	// Create timeout context
	timeoutSeconds := protocol.GetTimeout(protocol.HookEvent(input.Event))
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	// Route to appropriate handler
	switch protocol.HookEvent(input.Event) {
	case protocol.EventSessionStart:
		return d.handleSessionStart(ctx, input)

	case protocol.EventPreTool:
		return d.handlePreTool(ctx, input)

	case protocol.EventPostTool:
		return d.handlePostTool(ctx, input)

	case protocol.EventSessionEnd:
		return d.handleSessionEnd(ctx, input)

	case protocol.EventPreCompact:
		return d.handleSimple(ctx, "Pre-compact")

	case protocol.EventStop:
		return d.handleSimple(ctx, "Stop")

	case protocol.EventNotification:
		return d.handleSimple(ctx, "Notification")

	case protocol.EventQualityGate:
		return d.handleSimple(ctx, "Quality gate")

	case protocol.EventCommit:
		return d.handleSimple(ctx, "Commit")

	case protocol.EventPush:
		return d.handleSimple(ctx, "Push")

	case protocol.EventCompact:
		return d.handleSimple(ctx, "Compact")

	default:
		return d.handleError(ctx, fmt.Errorf("unknown event: %s", input.Event))
	}
}

// handleSessionStart handles session-start events
func (d *HookDispatcher) handleSessionStart(ctx context.Context, input *protocol.HookInput) error {
	response, err := d.sessionStartHandler.Handle(input)
	if err != nil {
		return d.handleError(ctx, err)
	}
	return response.WriteToStdout()
}

// handleSessionEnd handles session-end events
func (d *HookDispatcher) handleSessionEnd(ctx context.Context, input *protocol.HookInput) error {
	response, err := d.sessionEndHandler.Handle(input)
	if err != nil {
		return d.handleError(ctx, err)
	}
	return response.WriteToStdout()
}

// handlePreTool handles pre-tool events
func (d *HookDispatcher) handlePreTool(ctx context.Context, input *protocol.HookInput) error {
	response, err := d.preToolHandler.Handle(input)
	if err != nil {
		return d.handleError(ctx, err)
	}
	return response.WriteToStdout()
}

// handlePostTool handles post-tool events
func (d *HookDispatcher) handlePostTool(ctx context.Context, input *protocol.HookInput) error {
	response, err := d.postToolHandler.Handle(ctx, input)
	if err != nil {
		return d.handleError(ctx, err)
	}
	return response.WriteToStdout()
}

// handleSimple handles events that don't require special handling
func (d *HookDispatcher) handleSimple(ctx context.Context, eventType string) error {
	// For now, just suppress output
	response := protocol.NewMessageResponse("", true)
	return response.WriteToStdout()
}

// handleError handles errors and writes error response
func (d *HookDispatcher) handleError(ctx context.Context, err error) error {
	response := protocol.NewErrorResponse(err.Error())
	if writeErr := response.WriteToStdout(); writeErr != nil {
		// If we can't write JSON, at least log to stderr
		fmt.Fprintf(os.Stderr, "Hook error: %v\n", err)
		return writeErr
	}
	return err
}

// Run is the main entry point for hook execution
func Run() error {
	// Read input from stdin
	input, err := protocol.ReadFromStdin()
	if err != nil {
		response := protocol.NewErrorResponse(err.Error())
		_ = response.WriteToStdout()
		os.Exit(1)
	}

	// Create dispatcher and dispatch
	dispatcher := NewHookDispatcher()
	if err := dispatcher.Dispatch(context.Background(), input); err != nil {
		os.Exit(1)
	}

	return nil
}
