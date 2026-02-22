package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// JSONRPCError represents a JSON-RPC 2.0 error object.
type JSONRPCError struct {
	// Code is the error code.
	Code int `json:"code"`

	// Message is a short description of the error.
	Message string `json:"message"`

	// Data contains additional information about the error.
	Data json.RawMessage `json:"data,omitempty"`
}

// Error implements the error interface for JSONRPCError.
func (e *JSONRPCError) Error() string {
	return fmt.Sprintf("jsonrpc error %d: %s", e.Code, e.Message)
}

// Standard JSON-RPC 2.0 error codes.
const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
)

// jsonrpcRequest is the outgoing JSON-RPC 2.0 request message.
type jsonrpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// jsonrpcNotification is the outgoing JSON-RPC 2.0 notification (no ID, no response expected).
type jsonrpcNotification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// incomingMessage represents any incoming JSON-RPC 2.0 message from the server.
type incomingMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// isResponse returns true if the message is a response (has ID, no method).
func (m *incomingMessage) isResponse() bool {
	return len(m.ID) > 0 && m.Method == ""
}

// MessageTransport handles reading and writing LSP base protocol messages
// (Content-Length headers + JSON body) over a byte stream.
type MessageTransport interface {
	// ReadMessage reads a complete LSP message and returns the JSON body.
	ReadMessage(ctx context.Context) (json.RawMessage, error)

	// WriteMessage writes a JSON body as a complete LSP message with headers.
	WriteMessage(ctx context.Context, data json.RawMessage) error

	// Close closes the transport connection.
	Close() error
}

// StreamTransport implements MessageTransport over an io.ReadWriteCloser
// using the LSP base protocol (Content-Length headers).
type StreamTransport struct {
	reader  *bufio.Reader
	writer  io.Writer
	closer  io.Closer
	writeMu sync.Mutex
}

// NewStreamTransport creates a new StreamTransport from separate reader, writer, and closer.
// For stdio transport, reader is stdout pipe, writer is stdin pipe, closer closes both.
func NewStreamTransport(reader io.Reader, writer io.Writer, closer io.Closer) *StreamTransport {
	return &StreamTransport{
		reader: bufio.NewReaderSize(reader, 64*1024),
		writer: writer,
		closer: closer,
	}
}

// ReadMessage reads a complete LSP message from the stream.
// It parses the Content-Length header and reads exactly that many bytes.
func (t *StreamTransport) ReadMessage(_ context.Context) (json.RawMessage, error) {
	var contentLength int

	for {
		line, err := t.reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("read header: %w", err)
		}
		line = strings.TrimRight(line, "\r\n")

		if line == "" {
			break
		}

		if after, ok := strings.CutPrefix(line, "Content-Length: "); ok {
			val := after
			n, err := strconv.Atoi(val)
			if err != nil {
				return nil, fmt.Errorf("invalid Content-Length %q: %w", val, err)
			}
			contentLength = n
		}
		// Ignore other headers (e.g., Content-Type)
	}

	if contentLength <= 0 {
		return nil, fmt.Errorf("missing or invalid Content-Length")
	}

	body := make([]byte, contentLength)
	if _, err := io.ReadFull(t.reader, body); err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return json.RawMessage(body), nil
}

// WriteMessage writes a JSON body with Content-Length header to the stream.
func (t *StreamTransport) WriteMessage(_ context.Context, data json.RawMessage) error {
	t.writeMu.Lock()
	defer t.writeMu.Unlock()

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	if _, err := io.WriteString(t.writer, header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	if _, err := t.writer.Write(data); err != nil {
		return fmt.Errorf("write body: %w", err)
	}
	return nil
}

// Close closes the underlying stream.
func (t *StreamTransport) Close() error {
	if t.closer != nil {
		return t.closer.Close()
	}
	return nil
}

// TCPTransport creates a MessageTransport over a TCP connection.
func TCPTransport(addr string) (*StreamTransport, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("tcp dial %s: %w", addr, err)
	}
	return NewStreamTransport(conn, conn, conn), nil
}

// Conn represents a JSON-RPC 2.0 connection that can send requests and notifications.
type Conn interface {
	// Call sends a JSON-RPC request and unmarshals the response result into result.
	// result must be a pointer to the expected response type, or nil.
	Call(ctx context.Context, method string, params any, result any) error

	// Notify sends a JSON-RPC notification (no response expected).
	Notify(ctx context.Context, method string, params any) error

	// Close closes the connection and releases resources.
	Close() error
}

// connection implements Conn using a MessageTransport.
type connection struct {
	transport MessageTransport
	nextID    atomic.Int64
	pending   map[int64]chan *incomingMessage
	mu        sync.Mutex
	done      chan struct{}
	closeOnce sync.Once
	closeErr  error
}

// NewConn creates a new JSON-RPC connection over the given transport.
// It starts a background goroutine to read and dispatch responses.
func NewConn(transport MessageTransport) Conn {
	c := &connection{
		transport: transport,
		pending:   make(map[int64]chan *incomingMessage),
		done:      make(chan struct{}),
	}
	go c.readLoop()
	return c
}

// Call sends a JSON-RPC request and waits for the matching response.
func (c *connection) Call(ctx context.Context, method string, params any, result any) error {
	select {
	case <-c.done:
		return ErrConnectionClosed
	default:
	}

	id := c.nextID.Add(1)
	ch := make(chan *incomingMessage, 1)

	func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.pending[id] = ch
	}()

	req := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
	}
	if params != nil {
		p, err := json.Marshal(params)
		if err != nil {
			c.removePending(id)
			return fmt.Errorf("marshal params for %s: %w", method, err)
		}
		req.Params = p
	}

	data, err := json.Marshal(req)
	if err != nil {
		c.removePending(id)
		return fmt.Errorf("marshal request for %s: %w", method, err)
	}

	if err := c.transport.WriteMessage(ctx, json.RawMessage(data)); err != nil {
		c.removePending(id)
		return fmt.Errorf("write %s request: %w", method, err)
	}

	select {
	case <-ctx.Done():
		c.removePending(id)
		return ctx.Err()
	case msg, ok := <-ch:
		if !ok || msg == nil {
			return ErrConnectionClosed
		}
		if msg.Error != nil {
			return msg.Error
		}
		if result != nil && len(msg.Result) > 0 {
			if err := json.Unmarshal(msg.Result, result); err != nil {
				return fmt.Errorf("unmarshal %s result: %w", method, err)
			}
		}
		return nil
	}
}

// Notify sends a JSON-RPC notification (no response expected).
func (c *connection) Notify(ctx context.Context, method string, params any) error {
	select {
	case <-c.done:
		return ErrConnectionClosed
	default:
	}

	notif := jsonrpcNotification{
		JSONRPC: "2.0",
		Method:  method,
	}
	if params != nil {
		p, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("marshal params for %s: %w", method, err)
		}
		notif.Params = p
	}

	data, err := json.Marshal(notif)
	if err != nil {
		return fmt.Errorf("marshal notification %s: %w", method, err)
	}

	return c.transport.WriteMessage(ctx, json.RawMessage(data))
}

// Close stops the connection and releases all pending requests.
func (c *connection) Close() error {
	c.closeOnce.Do(func() {
		close(c.done)
		c.closeErr = c.transport.Close()
		c.drainPending()
	})
	return c.closeErr
}

// @MX:WARN: [AUTO] 무한 루프에서 메시지를 읽습니다. 종료 조건이 불분명합니다.
// @MX:REASON: [AUTO] 고루틴이 오류 시에도 종료되지 않고 계속 실행될 수 있습니다
// readLoop reads messages from the transport and dispatches responses.
func (c *connection) readLoop() {
	for {
		msg, err := c.transport.ReadMessage(context.Background())
		if err != nil {
			c.drainPending()
			return
		}

		var incoming incomingMessage
		if err := json.Unmarshal(msg, &incoming); err != nil {
			continue
		}

		if incoming.isResponse() {
			var id int64
			if err := json.Unmarshal(incoming.ID, &id); err != nil {
				continue
			}

			ch := func() chan *incomingMessage {
				c.mu.Lock()
				defer c.mu.Unlock()
				ch, ok := c.pending[id]
				if ok {
					delete(c.pending, id)
					return ch
				}
				return nil
			}()

			if ch != nil {
				ch <- &incoming
			}
		}
		// Notifications from the server are currently ignored.
		// Future: dispatch to registered notification handlers.
	}
}

// removePending removes a pending request channel by ID.
func (c *connection) removePending(id int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.pending, id)
}

// drainPending closes all pending request channels.
func (c *connection) drainPending() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for id, ch := range c.pending {
		close(ch)
		delete(c.pending, id)
	}
}

// EncodeMessage encodes a JSON body with LSP Content-Length headers.
// This is a utility function for testing and low-level protocol work.
func EncodeMessage(body []byte) []byte {
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	result := make([]byte, len(header)+len(body))
	copy(result, header)
	copy(result[len(header):], body)
	return result
}
