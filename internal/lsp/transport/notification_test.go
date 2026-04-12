package transport_test

import (
	"encoding/json"
	"errors"
	"testing"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
	"github.com/modu-ai/moai-adk/internal/lsp/transport"
)

// TestNotificationRouter_Register_Dispatch verifies that a registered handler
// is called when Dispatch is invoked with the matching method.
func TestNotificationRouter_Register_Dispatch(t *testing.T) {
	t.Parallel()

	r := transport.NewNotificationRouter()

	var received json.RawMessage
	err := r.Register("textDocument/publishDiagnostics", func(params json.RawMessage) error {
		received = params
		return nil
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	payload := json.RawMessage(`{"uri":"file:///a.go","diagnostics":[]}`)
	if err := r.Dispatch("textDocument/publishDiagnostics", payload); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}

	if string(received) != string(payload) {
		t.Errorf("received = %s, want %s", received, payload)
	}
}

// TestNotificationRouter_DuplicateRegister verifies that registering the same
// method twice returns an error (idempotency guard).
func TestNotificationRouter_DuplicateRegister(t *testing.T) {
	t.Parallel()

	r := transport.NewNotificationRouter()

	noop := func(json.RawMessage) error { return nil }
	if err := r.Register("textDocument/publishDiagnostics", noop); err != nil {
		t.Fatalf("first Register: %v", err)
	}

	err := r.Register("textDocument/publishDiagnostics", noop)
	if err == nil {
		t.Error("second Register: expected error for duplicate method, got nil")
	}
}

// TestNotificationRouter_UnknownMethod_Ignored verifies that Dispatch silently
// ignores notifications for unregistered methods (server may emit out-of-scope
// notifications that we don't care about).
func TestNotificationRouter_UnknownMethod_Ignored(t *testing.T) {
	t.Parallel()

	r := transport.NewNotificationRouter()

	// 미등록 메서드 — 에러 없이 무시
	err := r.Dispatch("$/progress", json.RawMessage(`{}`))
	if err != nil {
		t.Errorf("Dispatch unknown method: %v, want nil", err)
	}
}

// TestNotificationRouter_HandlerError_Surfaced verifies that errors returned
// by registered handlers are propagated to the Dispatch caller.
func TestNotificationRouter_HandlerError_Surfaced(t *testing.T) {
	t.Parallel()

	r := transport.NewNotificationRouter()

	handlerErr := errors.New("handler failed")
	if err := r.Register("textDocument/publishDiagnostics", func(json.RawMessage) error {
		return handlerErr
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	err := r.Dispatch("textDocument/publishDiagnostics", json.RawMessage(`{}`))
	if !errors.Is(err, handlerErr) {
		t.Errorf("Dispatch error = %v, want %v", err, handlerErr)
	}
}

// TestNotificationRouter_PublishDiagnostics_ParsesModel verifies that the
// built-in publishDiagnostics handler parses the payload into []lsp.Diagnostic
// (REQ-LC-002b, reuses internal/lsp/models.go types).
func TestNotificationRouter_PublishDiagnostics_ParsesModel(t *testing.T) {
	t.Parallel()

	r := transport.NewNotificationRouter()

	var captured []lsp.Diagnostic

	err := r.RegisterPublishDiagnostics(func(uri string, diags []lsp.Diagnostic) error {
		captured = diags
		return nil
	})
	if err != nil {
		t.Fatalf("RegisterPublishDiagnostics: %v", err)
	}

	payload := json.RawMessage(`{
		"uri": "file:///main.go",
		"diagnostics": [
			{
				"range": {
					"start": {"line": 1, "character": 0},
					"end":   {"line": 1, "character": 5}
				},
				"severity": 1,
				"message": "undefined: foo"
			}
		]
	}`)

	if err := r.Dispatch("textDocument/publishDiagnostics", payload); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}

	if len(captured) != 1 {
		t.Fatalf("captured %d diagnostics, want 1", len(captured))
	}
	if captured[0].Message != "undefined: foo" {
		t.Errorf("diagnostic message = %q, want %q", captured[0].Message, "undefined: foo")
	}
	if captured[0].Severity != lsp.SeverityError {
		t.Errorf("severity = %v, want Error", captured[0].Severity)
	}
}

// TestNotificationRouter_Attach verifies that Attach registers the router's
// Dispatch as the notification handler on the given Transport.
func TestNotificationRouter_Attach_Dispatches(t *testing.T) {
	t.Parallel()

	ft := newFakeTransport()
	r := transport.NewNotificationRouter()

	var called bool
	if err := r.Register("textDocument/publishDiagnostics", func(json.RawMessage) error {
		called = true
		return nil
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	r.Attach(ft)

	// fakeTransport에 등록된 핸들러 직접 호출로 Attach 통합 검증
	handler, ok := ft.handlers["textDocument/publishDiagnostics"]
	if !ok {
		t.Fatal("Attach did not register handler on transport")
	}
	handler(json.RawMessage(`{}`))

	if !called {
		t.Error("Attach: handler was not invoked via registered notification")
	}
}
