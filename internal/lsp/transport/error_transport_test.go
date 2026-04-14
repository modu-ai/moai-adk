// error_transport_test.go uses the same package (white-box test) to access
// pnTransport and errorTransport internals, which are not exported.
// This avoids exporting test-only types while maximising coverage.
package transport

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"

	pntr "github.com/charmbracelet/x/powernap/pkg/transport"
)

// mockRPCConn is a test double for the rpcConn interface.
type mockRPCConn struct {
	callErr   error
	notifyErr error
	closeCalled bool
	handlers  map[string]pntr.NotificationHandler
}

func newMockRPCConn() *mockRPCConn {
	return &mockRPCConn{handlers: make(map[string]pntr.NotificationHandler)}
}

func (m *mockRPCConn) Call(_ context.Context, _ string, _, _ any) error {
	return m.callErr
}

func (m *mockRPCConn) Notify(_ context.Context, _ string, _ any) error {
	return m.notifyErr
}

func (m *mockRPCConn) RegisterNotificationHandler(method string, handler pntr.NotificationHandler) {
	m.handlers[method] = handler
}

func (m *mockRPCConn) Close() error {
	m.closeCalled = true
	return nil
}

// TestPnTransport_WithMockConn_Call verifies Call delegates to rpcConn.
func TestPnTransport_WithMockConn_Call(t *testing.T) {
	t.Parallel()

	mock := newMockRPCConn()
	pt := &pnTransport{conn: mock}

	if err := pt.Call(context.Background(), "initialize", nil, nil); err != nil {
		t.Errorf("Call returned unexpected error: %v", err)
	}
}

// TestPnTransport_WithMockConn_Call_Error verifies errors from rpcConn.Call are propagated.
func TestPnTransport_WithMockConn_Call_Error(t *testing.T) {
	t.Parallel()

	mock := newMockRPCConn()
	wantErr := errors.New("rpc error")
	mock.callErr = wantErr

	pt := &pnTransport{conn: mock}
	if err := pt.Call(context.Background(), "initialize", nil, nil); !errors.Is(err, wantErr) {
		t.Errorf("Call error = %v, want %v", err, wantErr)
	}
}

// TestPnTransport_WithMockConn_Notify verifies Notify delegates to rpcConn.
func TestPnTransport_WithMockConn_Notify(t *testing.T) {
	t.Parallel()

	mock := newMockRPCConn()
	pt := &pnTransport{conn: mock}

	if err := pt.Notify(context.Background(), "initialized", nil); err != nil {
		t.Errorf("Notify returned unexpected error: %v", err)
	}
}

// TestPnTransport_WithMockConn_Notify_Error verifies errors from rpcConn.Notify are propagated.
func TestPnTransport_WithMockConn_Notify_Error(t *testing.T) {
	t.Parallel()

	mock := newMockRPCConn()
	wantErr := errors.New("notify error")
	mock.notifyErr = wantErr

	pt := &pnTransport{conn: mock}
	if err := pt.Notify(context.Background(), "textDocument/didOpen", nil); !errors.Is(err, wantErr) {
		t.Errorf("Notify error = %v, want %v", err, wantErr)
	}
}

// TestPnTransport_WithMockConn_OnNotification verifies handler is registered and fires.
func TestPnTransport_WithMockConn_OnNotification(t *testing.T) {
	t.Parallel()

	mock := newMockRPCConn()
	pt := &pnTransport{conn: mock}

	called := false
	pt.OnNotification("textDocument/publishDiagnostics", func(_ json.RawMessage) {
		called = true
	})

	// 등록된 핸들러가 있어야 함
	h, ok := mock.handlers["textDocument/publishDiagnostics"]
	if !ok {
		t.Fatal("handler not registered in mockRPCConn")
	}

	// 핸들러 호출 — pnTransport 내부 래퍼가 json.RawMessage를 전달해야 함
	h(context.Background(), "textDocument/publishDiagnostics", json.RawMessage(`{}`))
	if !called {
		t.Error("handler was not called after dispatch")
	}
}

// TestPnTransport_WithMockConn_Close verifies Close delegates to rpcConn.
func TestPnTransport_WithMockConn_Close(t *testing.T) {
	t.Parallel()

	mock := newMockRPCConn()
	pt := &pnTransport{conn: mock}

	if err := pt.Close(); err != nil {
		t.Errorf("Close returned unexpected error: %v", err)
	}
	if !mock.closeCalled {
		t.Error("Close did not call rpcConn.Close")
	}
}

// TestPnTransport_NilConn_Close verifies Close returns nil when conn is nil.
func TestPnTransport_NilConn_Close(t *testing.T) {
	t.Parallel()

	pt := &pnTransport{conn: nil}
	if err := pt.Close(); err != nil {
		t.Errorf("Close with nil conn = %v, want nil", err)
	}
}

// TestNewPowernapTransport_WithStream verifies NewPowernapTransport succeeds with a real stream.
// The stream is backed by io.Pipe; we close it immediately after construction to avoid goroutine leak.
func TestNewPowernapTransport_WithStream(t *testing.T) {
	t.Parallel()

	pr, pw := newPipeStream()
	defer func() { _ = pw.Close() }()
	defer func() { _ = pr.Close() }()

	tr := NewPowernapTransport(pr)
	if tr == nil {
		t.Fatal("NewPowernapTransport returned nil")
	}
	// Close releases the underlying connection
	_ = tr.Close()
}

// pipeStream wraps two io.PipeReader+io.PipeWriter as a single io.ReadWriteCloser.
type pipeStream struct {
	r *io.PipeReader
	w *io.PipeWriter
}

func newPipeStream() (*pipeStream, *io.PipeWriter) {
	r, w := io.Pipe()
	return &pipeStream{r: r, w: w}, w
}

func (p *pipeStream) Read(b []byte) (int, error)  { return p.r.Read(b) }
func (p *pipeStream) Write(b []byte) (int, error) { return p.w.Write(b) }
func (p *pipeStream) Close() error                { return p.r.Close() }

// TestErrorTransport_AllMethods verifies errorTransport returns its stored error on every method.
func TestErrorTransport_AllMethods(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("connection failed")
	et := &errorTransport{err: wantErr}

	if err := et.Call(context.Background(), "initialize", nil, nil); !errors.Is(err, wantErr) {
		t.Errorf("Call error = %v, want %v", err, wantErr)
	}
	if err := et.Notify(context.Background(), "initialized", nil); !errors.Is(err, wantErr) {
		t.Errorf("Notify error = %v, want %v", err, wantErr)
	}

	// OnNotification은 empty body — 패닉 없이 호출 가능해야 함
	et.OnNotification("method", nil)

	if err := et.Close(); err != nil {
		t.Errorf("Close error = %v, want nil", err)
	}
}
