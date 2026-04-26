package transport_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/lsp/transport"
)

// fakeTransport is a test double implementing transport.Transport.
// It records calls and returns configured responses.
type fakeTransport struct {
	callResults  map[string]any   // method → result
	callErrors   map[string]error // method → error
	notifyCalled []string         // methods called via Notify
	callCalled   []string         // methods called via Call
	closed       bool
	handlers     map[string]func(params json.RawMessage)
}

func newFakeTransport() *fakeTransport {
	return &fakeTransport{
		callResults: make(map[string]any),
		callErrors:  make(map[string]error),
		handlers:    make(map[string]func(params json.RawMessage)),
	}
}

func (f *fakeTransport) Call(ctx context.Context, method string, params, result any) error {
	f.callCalled = append(f.callCalled, method)
	if err, ok := f.callErrors[method]; ok {
		return err
	}
	if r, ok := f.callResults[method]; ok && result != nil {
		// Merge the result via JSON: Call's signature is `result any`.
		b, err := json.Marshal(r)
		if err != nil {
			return err
		}
		return json.Unmarshal(b, result)
	}
	return nil
}

func (f *fakeTransport) Notify(ctx context.Context, method string, params any) error {
	f.notifyCalled = append(f.notifyCalled, method)
	if err, ok := f.callErrors[method]; ok {
		return err
	}
	return nil
}

func (f *fakeTransport) OnNotification(method string, handler func(params json.RawMessage)) {
	f.handlers[method] = handler
}

func (f *fakeTransport) Close() error {
	f.closed = true
	return nil
}

// Verify fakeTransport implements transport.Transport at compile time.
var _ transport.Transport = (*fakeTransport)(nil)

// TestTransport_Interface verifies the Transport interface surface is correct.
// This test validates the interface contract, not any concrete implementation.
func TestTransport_Interface(t *testing.T) {
	t.Parallel()

	ft := newFakeTransport()
	ctx := context.Background()

	// Call must accept params and populate result.
	ft.callResults["initialize"] = map[string]any{"serverInfo": "test"}
	var result map[string]any
	if err := ft.Call(ctx, "initialize", nil, &result); err != nil {
		t.Errorf("Call returned unexpected error: %v", err)
	}
	if len(ft.callCalled) != 1 || ft.callCalled[0] != "initialize" {
		t.Errorf("callCalled = %v, want [initialize]", ft.callCalled)
	}

	// Notify sends a notification without a result.
	if err := ft.Notify(ctx, "initialized", nil); err != nil {
		t.Errorf("Notify returned unexpected error: %v", err)
	}
	if len(ft.notifyCalled) != 1 || ft.notifyCalled[0] != "initialized" {
		t.Errorf("notifyCalled = %v, want [initialized]", ft.notifyCalled)
	}

	// Close
	if err := ft.Close(); err != nil {
		t.Errorf("Close returned unexpected error: %v", err)
	}
	if !ft.closed {
		t.Error("Close did not set closed=true")
	}
}

// TestTransport_Call_Error verifies errors from Call are propagated.
func TestTransport_Call_Error(t *testing.T) {
	t.Parallel()

	ft := newFakeTransport()
	wantErr := errors.New("rpc error")
	ft.callErrors["textDocument/definition"] = wantErr

	err := ft.Call(context.Background(), "textDocument/definition", nil, nil)
	if !errors.Is(err, wantErr) {
		t.Errorf("Call error = %v, want %v", err, wantErr)
	}
}

// TestTransport_Notify_Error verifies errors from Notify are propagated.
func TestTransport_Notify_Error(t *testing.T) {
	t.Parallel()

	ft := newFakeTransport()
	wantErr := errors.New("notify error")
	ft.callErrors["textDocument/didOpen"] = wantErr

	err := ft.Notify(context.Background(), "textDocument/didOpen", nil)
	if !errors.Is(err, wantErr) {
		t.Errorf("Notify error = %v, want %v", err, wantErr)
	}
}

// TestTransport_OnNotification_HandlerRegistered verifies handler registration.
func TestTransport_OnNotification_HandlerRegistered(t *testing.T) {
	t.Parallel()

	ft := newFakeTransport()

	called := false
	ft.OnNotification("textDocument/publishDiagnostics", func(params json.RawMessage) {
		called = true
	})

	handler, ok := ft.handlers["textDocument/publishDiagnostics"]
	if !ok {
		t.Fatal("handler not registered for textDocument/publishDiagnostics")
	}

	handler(json.RawMessage(`{}`))
	if !called {
		t.Error("handler was not called")
	}
}

// TestNewPowernapTransport_NotNil verifies NewPowernapTransport returns a non-nil Transport.
// This uses a nil stream to confirm the constructor is callable;
// actual RPC operation requires a real subprocess (deferred to T-006+ integration tests).
func TestNewPowernapTransport_NotNil(t *testing.T) {
	t.Parallel()

	tr := transport.NewPowernapTransport(nil)
	if tr == nil {
		t.Fatal("NewPowernapTransport(nil) = nil, want non-nil Transport")
	}
}

// TestNewPowernapTransport_CloseOnNil verifies Close on a nil-stream transport returns an error or nil (not panic).
func TestNewPowernapTransport_CloseOnNil(t *testing.T) {
	t.Parallel()

	tr := transport.NewPowernapTransport(nil)

	// Close must return either an error or nil when the stream is nil — must not panic.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Close panicked: %v", r)
		}
	}()

	// Must be callable without panicking.
	_ = tr.Close()
}

// TestNewPowernapTransport_NilStream_Call verifies Call returns an error (not panic) when stream is nil.
func TestNewPowernapTransport_NilStream_Call(t *testing.T) {
	t.Parallel()

	tr := transport.NewPowernapTransport(nil)
	err := tr.Call(context.Background(), "initialize", nil, nil)
	if err == nil {
		t.Error("Call on nil-stream transport: expected error, got nil")
	}
}

// TestNewPowernapTransport_NilStream_Notify verifies Notify returns an error (not panic) when stream is nil.
func TestNewPowernapTransport_NilStream_Notify(t *testing.T) {
	t.Parallel()

	tr := transport.NewPowernapTransport(nil)
	err := tr.Notify(context.Background(), "initialized", nil)
	if err == nil {
		t.Error("Notify on nil-stream transport: expected error, got nil")
	}
}

// TestNewPowernapTransport_NilStream_OnNotification verifies OnNotification is a no-op (not panic) when stream is nil.
func TestNewPowernapTransport_NilStream_OnNotification(t *testing.T) {
	t.Parallel()

	tr := transport.NewPowernapTransport(nil)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("OnNotification panicked: %v", r)
		}
	}()

	// nil-stream transport: no-op, no panic
	tr.OnNotification("textDocument/publishDiagnostics", func(_ json.RawMessage) {})
}
