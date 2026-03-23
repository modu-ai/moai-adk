package cli

import (
	"context"
	"io"

	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/update"
)

// --- Mock implementations for CLI dependency testing ---

// mockHookProtocol implements hook.Protocol for testing.
type mockHookProtocol struct {
	readInputFunc   func(r io.Reader) (*hook.HookInput, error)
	writeOutputFunc func(w io.Writer, output *hook.HookOutput) error
}

func (m *mockHookProtocol) ReadInput(r io.Reader) (*hook.HookInput, error) {
	if m.readInputFunc != nil {
		return m.readInputFunc(r)
	}
	return &hook.HookInput{}, nil
}

func (m *mockHookProtocol) WriteOutput(w io.Writer, output *hook.HookOutput) error {
	if m.writeOutputFunc != nil {
		return m.writeOutputFunc(w, output)
	}
	return nil
}

// mockHookRegistry implements hook.Registry for testing.
type mockHookRegistry struct {
	registerFunc func(handler hook.Handler)
	dispatchFunc func(ctx context.Context, event hook.EventType, input *hook.HookInput) (*hook.HookOutput, error)
	handlersFunc func(event hook.EventType) []hook.Handler
}

func (m *mockHookRegistry) Register(handler hook.Handler) {
	if m.registerFunc != nil {
		m.registerFunc(handler)
	}
}

func (m *mockHookRegistry) Dispatch(ctx context.Context, event hook.EventType, input *hook.HookInput) (*hook.HookOutput, error) {
	if m.dispatchFunc != nil {
		return m.dispatchFunc(ctx, event, input)
	}
	return hook.NewAllowOutput(), nil
}

func (m *mockHookRegistry) Handlers(event hook.EventType) []hook.Handler {
	if m.handlersFunc != nil {
		return m.handlersFunc(event)
	}
	return nil
}

// mockUpdateChecker implements update.Checker for testing.
type mockUpdateChecker struct {
	checkLatestFunc   func(ctx context.Context) (*update.VersionInfo, error)
	isUpdateAvailFunc func(current string) (bool, *update.VersionInfo, error)
}

func (m *mockUpdateChecker) CheckLatest(ctx context.Context) (*update.VersionInfo, error) {
	if m.checkLatestFunc != nil {
		return m.checkLatestFunc(ctx)
	}
	return &update.VersionInfo{Version: "1.0.0", URL: "https://example.com/moai-binary"}, nil
}

func (m *mockUpdateChecker) IsUpdateAvailable(current string) (bool, *update.VersionInfo, error) {
	if m.isUpdateAvailFunc != nil {
		return m.isUpdateAvailFunc(current)
	}
	return false, nil, nil
}

// mockUpdateOrchestrator implements update.Orchestrator for testing.
type mockUpdateOrchestrator struct {
	updateFunc func(ctx context.Context) (*update.UpdateResult, error)
}

func (m *mockUpdateOrchestrator) Update(ctx context.Context) (*update.UpdateResult, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx)
	}
	return &update.UpdateResult{PreviousVersion: "v0.0.0", NewVersion: "v0.0.1"}, nil
}


// mockHandler implements hook.Handler for testing.
type mockHandler struct {
	eventType hook.EventType
}

func (m *mockHandler) Handle(_ context.Context, _ *hook.HookInput) (*hook.HookOutput, error) {
	return hook.NewAllowOutput(), nil
}

func (m *mockHandler) EventType() hook.EventType {
	return m.eventType
}
