package transport_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/lsp/transport"
)

// TestWrapCallError_ContainsMethod verifies that WrapCallError embeds
// the method name in the returned error (REQ-LC-040).
func TestWrapCallError_ContainsMethod(t *testing.T) {
	t.Parallel()

	inner := errors.New("connection reset")
	err := transport.WrapCallError("textDocument/definition", "file:///a.go", "go", inner)
	if err == nil {
		t.Fatal("WrapCallError returned nil")
	}
	if !errors.Is(err, inner) {
		t.Errorf("WrapCallError: errors.Is(err, inner) = false, want true")
	}
	msg := err.Error()
	if !strings.Contains(msg, "textDocument/definition") {
		t.Errorf("error message %q does not contain method name", msg)
	}
}

// TestWrapCallError_ContainsLanguage verifies that WrapCallError embeds
// the server language string in the returned error (REQ-LC-040).
func TestWrapCallError_ContainsLanguage(t *testing.T) {
	t.Parallel()

	inner := errors.New("timeout")
	err := transport.WrapCallError("textDocument/references", "file:///b.py", "python", inner)
	msg := err.Error()
	if !strings.Contains(msg, "python") {
		t.Errorf("error message %q does not contain language", msg)
	}
}

// TestWrapCallError_ContainsURI verifies that WrapCallError embeds the file URI
// in the returned error when it is non-empty (REQ-LC-040).
func TestWrapCallError_ContainsURI(t *testing.T) {
	t.Parallel()

	inner := errors.New("parse error")
	err := transport.WrapCallError("textDocument/hover", "file:///c.ts", "typescript", inner)
	msg := err.Error()
	if !strings.Contains(msg, "file:///c.ts") {
		t.Errorf("error message %q does not contain URI", msg)
	}
}

// TestWrapCallError_NilErr returns nil when the inner error is nil.
func TestWrapCallError_NilErr(t *testing.T) {
	t.Parallel()

	err := transport.WrapCallError("initialize", "", "go", nil)
	if err != nil {
		t.Errorf("WrapCallError(nil inner) = %v, want nil", err)
	}
}

// TestCallWithTimeout_HappyPath verifies that CallWithTimeout passes the call
// through successfully when the transport returns without error.
func TestCallWithTimeout_HappyPath(t *testing.T) {
	t.Parallel()

	ft := newFakeTransport()
	ft.callResults["initialize"] = map[string]any{"serverInfo": "ok"}

	var result map[string]any
	err := transport.CallWithTimeout(context.Background(), ft, "initialize", nil, &result, "go")
	if err != nil {
		t.Errorf("CallWithTimeout happy path: %v", err)
	}
}

// TestCallWithTimeout_TransportError verifies that errors from the underlying
// transport are wrapped with method + language context (REQ-LC-040).
func TestCallWithTimeout_TransportError(t *testing.T) {
	t.Parallel()

	inner := errors.New("rpc protocol error")
	ft := newFakeTransport()
	ft.callErrors["textDocument/definition"] = inner

	err := transport.CallWithTimeout(context.Background(), ft, "textDocument/definition", nil, nil, "go")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, inner) {
		t.Errorf("errors.Is(err, inner) = false; err = %v", err)
	}
	if !strings.Contains(err.Error(), "textDocument/definition") {
		t.Errorf("error %q does not contain method name", err.Error())
	}
}

// TestCallWithTimeout_DeadlineExceeded verifies that when the context deadline
// is exceeded, ErrRequestTimeout is returned (REQ-LC-041).
func TestCallWithTimeout_DeadlineExceeded(t *testing.T) {
	t.Parallel()

	// fake transport that never completes — channel blocking
	blocking := &blockingTransport{}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := transport.CallWithTimeout(ctx, blocking, "textDocument/hover", nil, nil, "go")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, transport.ErrRequestTimeout) {
		t.Errorf("error = %v, want wrapping ErrRequestTimeout", err)
	}
}

// TestCallWithTimeout_NilCtx verifies that passing a nil context returns
// a typed error rather than panicking.
func TestCallWithTimeout_NilCtx(t *testing.T) {
	t.Parallel()

	ft := newFakeTransport()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("CallWithTimeout panicked on nil ctx: %v", r)
		}
	}()

	//nolint:staticcheck // intentional nil context test
	err := transport.CallWithTimeout(nil, ft, "initialize", nil, nil, "go") //nolint:staticcheck
	if err == nil {
		t.Error("CallWithTimeout(nil ctx): expected error, got nil")
	}
}

// blockingTransport is a Transport that blocks Call indefinitely.
// Used to test context cancellation / timeout behavior.
type blockingTransport struct{}

func (b *blockingTransport) Call(ctx context.Context, _ string, _, _ any) error {
	<-ctx.Done()
	return ctx.Err()
}

func (b *blockingTransport) Notify(_ context.Context, _ string, _ any) error {
	return nil
}

func (b *blockingTransport) OnNotification(_ string, _ func(params json.RawMessage)) {}

func (b *blockingTransport) Close() error { return nil }

// Verify blockingTransport implements Transport at compile time.
var _ transport.Transport = (*blockingTransport)(nil)
