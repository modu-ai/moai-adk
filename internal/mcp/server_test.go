package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// fakeHandler is a test double for Handler that records method calls.
type fakeHandler struct {
	method string
	result any
	err    error
}

func (f *fakeHandler) HandleMethod(_ context.Context, method string, _ json.RawMessage) (any, error) {
	f.method = method
	return f.result, f.err
}

func TestServer_Serve_DispatchesRequest(t *testing.T) {
	t.Parallel()

	handler := &fakeHandler{result: map[string]string{"status": "ok"}}
	srv := NewServer(handler)

	request := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}` + "\n"
	reader := strings.NewReader(request)
	var buf bytes.Buffer

	// Serve will process the single line then exit due to EOF.
	_ = srv.Serve(context.Background(), reader, &buf)

	if handler.method != "tools/list" {
		t.Errorf("handler.method: got %q want %q", handler.method, "tools/list")
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &resp); err != nil {
		t.Fatalf("parse response: %v (body: %s)", err, buf.String())
	}

	if resp.JSONRPC != "2.0" {
		t.Errorf("JSONRPC: got %q want %q", resp.JSONRPC, "2.0")
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
}

func TestServer_Serve_ReturnsErrorResponse(t *testing.T) {
	t.Parallel()

	handler := &fakeHandler{err: errTest("boom")}
	srv := NewServer(handler)

	request := `{"jsonrpc":"2.0","id":42,"method":"unknown"}` + "\n"
	reader := strings.NewReader(request)
	var buf bytes.Buffer

	_ = srv.Serve(context.Background(), reader, &buf)

	var resp JSONRPCResponse
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &resp); err != nil {
		t.Fatalf("parse response: %v (body: %s)", err, buf.String())
	}

	if resp.Error == nil {
		t.Fatal("expected error response, got nil")
	}
	if resp.Error.Message != "boom" {
		t.Errorf("Error.Message: got %q want %q", resp.Error.Message, "boom")
	}
}

func TestServer_Serve_SkipsNotification(t *testing.T) {
	t.Parallel()

	handler := &fakeHandler{result: "unused"}
	srv := NewServer(handler)

	// Notifications have no "id" field.
	notification := `{"jsonrpc":"2.0","method":"notifications/initialized"}` + "\n"
	reader := strings.NewReader(notification)
	var buf bytes.Buffer

	_ = srv.Serve(context.Background(), reader, &buf)

	// No response should have been written for a notification.
	if buf.Len() != 0 {
		t.Errorf("expected empty output for notification, got: %s", buf.String())
	}
}

func TestServer_Serve_SkipsInvalidJSON(t *testing.T) {
	t.Parallel()

	handler := &fakeHandler{}
	srv := NewServer(handler)

	reader := strings.NewReader("not-json\n")
	var buf bytes.Buffer

	_ = srv.Serve(context.Background(), reader, &buf)

	// Invalid JSON should be silently skipped.
	if buf.Len() != 0 {
		t.Errorf("expected no output for invalid JSON, got: %s", buf.String())
	}
	// Handler should not have been called.
	if handler.method != "" {
		t.Errorf("handler should not be called, got method %q", handler.method)
	}
}

// errTest is a simple error type used in tests.
type errTest string

func (e errTest) Error() string { return string(e) }
