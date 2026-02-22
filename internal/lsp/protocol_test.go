package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

func TestJSONRPCErrorString(t *testing.T) {
	t.Parallel()

	e := &JSONRPCError{Code: CodeMethodNotFound, Message: "method not found"}
	got := e.Error()
	want := "jsonrpc error -32601: method not found"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestJSONRPCErrorCodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code int
		want int
	}{
		{name: "ParseError", code: CodeParseError, want: -32700},
		{name: "InvalidRequest", code: CodeInvalidRequest, want: -32600},
		{name: "MethodNotFound", code: CodeMethodNotFound, want: -32601},
		{name: "InvalidParams", code: CodeInvalidParams, want: -32602},
		{name: "InternalError", code: CodeInternalError, want: -32603},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.code != tt.want {
				t.Errorf("%s = %d, want %d", tt.name, tt.code, tt.want)
			}
		})
	}
}

func TestEncodeMessage(t *testing.T) {
	t.Parallel()

	body := []byte(`{"jsonrpc":"2.0","id":1,"method":"test"}`)
	encoded := EncodeMessage(body)
	want := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(body), body)

	if string(encoded) != want {
		t.Errorf("EncodeMessage() = %q, want %q", encoded, want)
	}
}

func TestStreamTransportReadWrite(t *testing.T) {
	t.Parallel()

	clientReader, serverWriter := io.Pipe()
	serverReader, clientWriter := io.Pipe()

	clientCloser := &multiCloser{clientReader, clientWriter}
	serverCloser := &multiCloser{serverReader, serverWriter}

	clientTransport := NewStreamTransport(clientReader, clientWriter, clientCloser)
	serverTransport := NewStreamTransport(serverReader, serverWriter, serverCloser)

	body := json.RawMessage(`{"jsonrpc":"2.0","id":1,"result":"ok"}`)

	// Server writes a message.
	go func() {
		err := serverTransport.WriteMessage(context.Background(), body)
		if err != nil {
			t.Errorf("server WriteMessage error: %v", err)
		}
	}()

	// Client reads the message.
	got, err := clientTransport.ReadMessage(context.Background())
	if err != nil {
		t.Fatalf("client ReadMessage error: %v", err)
	}

	if string(got) != string(body) {
		t.Errorf("ReadMessage() = %q, want %q", got, body)
	}

	_ = clientTransport.Close()
	_ = serverTransport.Close()
}

func TestStreamTransportMultipleMessages(t *testing.T) {
	t.Parallel()

	reader, writer := io.Pipe()
	closer := &multiCloser{reader, writer}
	transport := NewStreamTransport(reader, writer, closer)
	defer func() { _ = transport.Close() }()

	messages := []string{
		`{"jsonrpc":"2.0","id":1,"result":"first"}`,
		`{"jsonrpc":"2.0","id":2,"result":"second"}`,
		`{"jsonrpc":"2.0","id":3,"result":"third"}`,
	}

	go func() {
		for _, msg := range messages {
			if err := transport.WriteMessage(context.Background(), json.RawMessage(msg)); err != nil {
				return
			}
		}
	}()

	for i, want := range messages {
		got, err := transport.ReadMessage(context.Background())
		if err != nil {
			t.Fatalf("message %d ReadMessage error: %v", i, err)
		}
		if string(got) != want {
			t.Errorf("message %d = %q, want %q", i, got, want)
		}
	}
}

func TestStreamTransportReadClosedPipe(t *testing.T) {
	t.Parallel()

	reader, writer := io.Pipe()
	transport := NewStreamTransport(reader, writer, reader)

	_ = writer.Close()

	_, err := transport.ReadMessage(context.Background())
	if err == nil {
		t.Error("expected error reading from closed pipe, got nil")
	}
}

func TestTCPTransportRoundTrip(t *testing.T) {
	t.Parallel()

	// Start a TCP listener.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen error: %v", err)
	}
	defer func() { _ = ln.Close() }()

	body := json.RawMessage(`{"jsonrpc":"2.0","id":1,"result":"tcp-ok"}`)

	// Server goroutine: accept one connection, echo back a message.
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		serverTransport := NewStreamTransport(conn, conn, nil)
		msg, err := serverTransport.ReadMessage(context.Background())
		if err != nil {
			return
		}
		_ = serverTransport.WriteMessage(context.Background(), msg)
	}()

	// Client connects via TCP.
	clientTransport, err := TCPTransport(ln.Addr().String())
	if err != nil {
		t.Fatalf("TCPTransport error: %v", err)
	}
	defer func() { _ = clientTransport.Close() }()

	// Client sends and receives.
	if err := clientTransport.WriteMessage(context.Background(), body); err != nil {
		t.Fatalf("WriteMessage error: %v", err)
	}
	got, err := clientTransport.ReadMessage(context.Background())
	if err != nil {
		t.Fatalf("ReadMessage error: %v", err)
	}
	if string(got) != string(body) {
		t.Errorf("round-trip = %q, want %q", got, body)
	}
}

func TestConnectionCall(t *testing.T) {
	t.Parallel()

	clientReader, serverWriter := io.Pipe()
	serverReader, clientWriter := io.Pipe()

	clientTransport := NewStreamTransport(clientReader, clientWriter, &multiCloser{clientReader, clientWriter})
	serverTransport := NewStreamTransport(serverReader, serverWriter, &multiCloser{serverReader, serverWriter})

	// Mock server: reads request, sends response.
	go func() {
		msg, err := serverTransport.ReadMessage(context.Background())
		if err != nil {
			return
		}
		var req struct {
			ID int64 `json:"id"`
		}
		_ = json.Unmarshal(msg, &req)

		resp := fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"result":{"value":"hello"}}`, req.ID)
		_ = serverTransport.WriteMessage(context.Background(), json.RawMessage(resp))
	}()

	conn := NewConn(clientTransport)
	defer func() { _ = conn.Close() }()

	var result struct {
		Value string `json:"value"`
	}
	err := conn.Call(context.Background(), "test/method", nil, &result)
	if err != nil {
		t.Fatalf("Call error: %v", err)
	}
	if result.Value != "hello" {
		t.Errorf("result.Value = %q, want %q", result.Value, "hello")
	}
}

func TestConnectionCallWithParams(t *testing.T) {
	t.Parallel()

	clientReader, serverWriter := io.Pipe()
	serverReader, clientWriter := io.Pipe()

	clientTransport := NewStreamTransport(clientReader, clientWriter, &multiCloser{clientReader, clientWriter})
	serverTransport := NewStreamTransport(serverReader, serverWriter, &multiCloser{serverReader, serverWriter})

	// Mock server: reads request, verifies params, sends response.
	go func() {
		msg, err := serverTransport.ReadMessage(context.Background())
		if err != nil {
			return
		}
		var req struct {
			ID     int64           `json:"id"`
			Params json.RawMessage `json:"params"`
		}
		_ = json.Unmarshal(msg, &req)

		var params struct {
			Name string `json:"name"`
		}
		_ = json.Unmarshal(req.Params, &params)

		resp := fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"result":{"greeting":"hello %s"}}`, req.ID, params.Name)
		_ = serverTransport.WriteMessage(context.Background(), json.RawMessage(resp))
	}()

	conn := NewConn(clientTransport)
	defer func() { _ = conn.Close() }()

	params := map[string]string{"name": "world"}
	var result struct {
		Greeting string `json:"greeting"`
	}
	err := conn.Call(context.Background(), "test/greet", params, &result)
	if err != nil {
		t.Fatalf("Call error: %v", err)
	}
	if result.Greeting != "hello world" {
		t.Errorf("result.Greeting = %q, want %q", result.Greeting, "hello world")
	}
}

func TestConnectionCallError(t *testing.T) {
	t.Parallel()

	clientReader, serverWriter := io.Pipe()
	serverReader, clientWriter := io.Pipe()

	clientTransport := NewStreamTransport(clientReader, clientWriter, &multiCloser{clientReader, clientWriter})
	serverTransport := NewStreamTransport(serverReader, serverWriter, &multiCloser{serverReader, serverWriter})

	// Mock server: returns a JSON-RPC error.
	go func() {
		msg, err := serverTransport.ReadMessage(context.Background())
		if err != nil {
			return
		}
		var req struct {
			ID int64 `json:"id"`
		}
		_ = json.Unmarshal(msg, &req)

		resp := fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"error":{"code":-32601,"message":"method not found"}}`, req.ID)
		_ = serverTransport.WriteMessage(context.Background(), json.RawMessage(resp))
	}()

	conn := NewConn(clientTransport)
	defer func() { _ = conn.Close() }()

	err := conn.Call(context.Background(), "unknown/method", nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var rpcErr *JSONRPCError
	if ok := errorAs(err, &rpcErr); !ok {
		t.Fatalf("expected JSONRPCError, got %T: %v", err, err)
	}
	if rpcErr.Code != CodeMethodNotFound {
		t.Errorf("error code = %d, want %d", rpcErr.Code, CodeMethodNotFound)
	}
}

func TestConnectionCallContextTimeout(t *testing.T) {
	t.Parallel()

	clientReader, serverWriter := io.Pipe()
	serverReader, clientWriter := io.Pipe()

	clientTransport := NewStreamTransport(clientReader, clientWriter, &multiCloser{clientReader, clientWriter})

	// Drain serverReader so WriteMessage does not block on pipe.
	go func() { _, _ = io.Copy(io.Discard, serverReader) }()

	// Server never writes a response (keep serverWriter open so readLoop blocks).
	_ = serverWriter

	conn := NewConn(clientTransport)
	defer func() { _ = conn.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := conn.Call(ctx, "slow/method", nil, nil)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("error = %v, want context.DeadlineExceeded", err)
	}
}

func TestConnectionNotify(t *testing.T) {
	t.Parallel()

	clientReader, serverWriter := io.Pipe()
	serverReader, clientWriter := io.Pipe()

	clientTransport := NewStreamTransport(clientReader, clientWriter, &multiCloser{clientReader, clientWriter})
	serverTransport := NewStreamTransport(serverReader, serverWriter, &multiCloser{serverReader, serverWriter})
	_ = serverWriter // keep open so readLoop blocks

	conn := NewConn(clientTransport)
	defer func() { _ = conn.Close() }()

	// Read in goroutine first because io.Pipe blocks until both ends are active.
	type readResult struct {
		msg json.RawMessage
		err error
	}
	ch := make(chan readResult, 1)
	go func() {
		msg, err := serverTransport.ReadMessage(context.Background())
		ch <- readResult{msg, err}
	}()

	err := conn.Notify(context.Background(), "initialized", nil)
	if err != nil {
		t.Fatalf("Notify error: %v", err)
	}

	res := <-ch
	if res.err != nil {
		t.Fatalf("server ReadMessage error: %v", res.err)
	}

	var notif struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
	}
	if err := json.Unmarshal(res.msg, &notif); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if notif.JSONRPC != "2.0" {
		t.Errorf("jsonrpc = %q, want %q", notif.JSONRPC, "2.0")
	}
	if notif.Method != "initialized" {
		t.Errorf("method = %q, want %q", notif.Method, "initialized")
	}

	// Verify no ID field.
	var raw map[string]json.RawMessage
	_ = json.Unmarshal(res.msg, &raw)
	if _, hasID := raw["id"]; hasID {
		t.Error("notification should not have 'id' field")
	}
}

func TestConnectionCallAfterClose(t *testing.T) {
	t.Parallel()

	reader, writer := io.Pipe()
	transport := NewStreamTransport(reader, writer, &multiCloser{reader, writer})

	conn := NewConn(transport)
	_ = conn.Close()

	err := conn.Call(context.Background(), "test", nil, nil)
	if err != ErrConnectionClosed {
		t.Errorf("error = %v, want ErrConnectionClosed", err)
	}
}

func TestConnectionNotifyAfterClose(t *testing.T) {
	t.Parallel()

	reader, writer := io.Pipe()
	transport := NewStreamTransport(reader, writer, &multiCloser{reader, writer})

	conn := NewConn(transport)
	_ = conn.Close()

	err := conn.Notify(context.Background(), "test", nil)
	if err != ErrConnectionClosed {
		t.Errorf("error = %v, want ErrConnectionClosed", err)
	}
}

func TestConnectionMultipleCalls(t *testing.T) {
	t.Parallel()

	clientReader, serverWriter := io.Pipe()
	serverReader, clientWriter := io.Pipe()

	clientTransport := NewStreamTransport(clientReader, clientWriter, &multiCloser{clientReader, clientWriter})
	serverTransport := NewStreamTransport(serverReader, serverWriter, &multiCloser{serverReader, serverWriter})

	// Mock server: echoes the method name as result for each request.
	go func() {
		for range 3 {
			msg, err := serverTransport.ReadMessage(context.Background())
			if err != nil {
				return
			}
			var req struct {
				ID     int64  `json:"id"`
				Method string `json:"method"`
			}
			_ = json.Unmarshal(msg, &req)

			resp := fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"result":"%s"}`, req.ID, req.Method)
			_ = serverTransport.WriteMessage(context.Background(), json.RawMessage(resp))
		}
	}()

	conn := NewConn(clientTransport)
	defer func() { _ = conn.Close() }()

	methods := []string{"method/a", "method/b", "method/c"}
	for _, method := range methods {
		var result string
		err := conn.Call(context.Background(), method, nil, &result)
		if err != nil {
			t.Fatalf("Call(%s) error: %v", method, err)
		}
		if result != method {
			t.Errorf("Call(%s) result = %q, want %q", method, result, method)
		}
	}
}

func TestIncomingMessageIsResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		msg  incomingMessage
		want bool
	}{
		{
			name: "response_with_result",
			msg:  incomingMessage{ID: json.RawMessage(`1`), Result: json.RawMessage(`"ok"`)},
			want: true,
		},
		{
			name: "response_with_error",
			msg:  incomingMessage{ID: json.RawMessage(`1`), Error: &JSONRPCError{Code: -1, Message: "err"}},
			want: true,
		},
		{
			name: "notification",
			msg:  incomingMessage{Method: "textDocument/publishDiagnostics"},
			want: false,
		},
		{
			name: "request_from_server",
			msg:  incomingMessage{ID: json.RawMessage(`1`), Method: "window/showMessage"},
			want: false,
		},
		{
			name: "empty",
			msg:  incomingMessage{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.msg.isResponse(); got != tt.want {
				t.Errorf("isResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStreamTransportInvalidContentLength(t *testing.T) {
	t.Parallel()

	input := "Content-Length: abc\r\n\r\n"
	reader := strings.NewReader(input)
	transport := NewStreamTransport(reader, io.Discard, nil)

	_, err := transport.ReadMessage(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid Content-Length, got nil")
	}
	if !strings.Contains(err.Error(), "invalid Content-Length") {
		t.Errorf("error = %q, want to contain 'invalid Content-Length'", err.Error())
	}
}

func TestStreamTransportMissingContentLength(t *testing.T) {
	t.Parallel()

	input := "Content-Type: application/json\r\n\r\n{}"
	reader := strings.NewReader(input)
	transport := NewStreamTransport(reader, io.Discard, nil)

	_, err := transport.ReadMessage(context.Background())
	if err == nil {
		t.Fatal("expected error for missing Content-Length, got nil")
	}
	if !strings.Contains(err.Error(), "missing or invalid Content-Length") {
		t.Errorf("error = %q, want to contain 'missing or invalid Content-Length'", err.Error())
	}
}

// multiCloser implements io.Closer by closing multiple io.Closers.
type multiCloser struct {
	r io.Closer
	w io.Closer
}

func (mc *multiCloser) Close() error {
	err1 := mc.r.Close()
	err2 := mc.w.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

// errorAs is a helper that wraps errors.As for use in tests.
func errorAs[T any](err error, target *T) bool {
	return errors.As(err, target)
}
