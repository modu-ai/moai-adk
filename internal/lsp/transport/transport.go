package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	pntr "github.com/charmbracelet/x/powernap/pkg/transport"
)

// Transport is the JSON-RPC interface used by core.Client to communicate with
// a language server subprocess (REQ-LC-001).
//
// Implementations MUST be safe for concurrent use.
// @MX:ANCHOR: [AUTO] Transport interface — central contract between core.Client and the wire layer
// @MX:REASON: fan_in >= 3 — core.Client, document sync, notification handler, and tests all consume this interface
type Transport interface {
	// Call sends a JSON-RPC request and blocks until the response is received
	// or ctx is cancelled. The response is unmarshalled into result.
	Call(ctx context.Context, method string, params, result any) error

	// Notify sends a JSON-RPC notification (no response expected).
	Notify(ctx context.Context, method string, params any) error

	// OnNotification registers a handler for server-initiated notifications.
	// Only one handler per method is supported; subsequent calls replace the previous handler.
	OnNotification(method string, handler func(params json.RawMessage))

	// Close releases all resources associated with this transport.
	// After Close returns, all pending and future calls return errors.
	Close() error
}

// rpcConn abstracts powernap's *transport.Connection for testability.
type rpcConn interface {
	Call(ctx context.Context, method string, params, result any) error
	Notify(ctx context.Context, method string, params any) error
	RegisterNotificationHandler(method string, handler pntr.NotificationHandler)
	Close() error
}

// pnTransport wraps powernap's transport.Connection to implement Transport.
// At Sprint 1 this is a skeleton: real JSON-RPC correlation is implemented in T-006.
//
// @MX:WARN: [AUTO] pnTransport wraps a live subprocess stdio connection
// @MX:REASON: subprocess stdio is not safe to close from multiple goroutines; Close must be called exactly once
type pnTransport struct {
	conn rpcConn
}

// NewPowernapTransport creates a Transport backed by powernap's Connection.
//
// stream must be an io.ReadWriteCloser that wraps the language server's stdio pipes.
// Passing nil is permitted for testing purposes; all methods will return errors.
//
// REQ-LC-001: delegates to github.com/charmbracelet/x/powernap for JSON-RPC.
func NewPowernapTransport(stream io.ReadWriteCloser) Transport {
	if stream == nil {
		// nil-stream sentinel: conn is nil, methods return descriptive errors
		return &pnTransport{conn: nil}
	}

	ctx := context.Background()
	conn, err := pntr.NewConnection(ctx, stream, slog.Default())
	if err != nil {
		// Connection creation failed: return a transport that reports the error
		return &errorTransport{err: fmt.Errorf("powernap connection failed: %w", err)}
	}

	return &pnTransport{conn: conn}
}

// Call implements Transport.Call.
func (t *pnTransport) Call(ctx context.Context, method string, params, result any) error {
	if t.conn == nil {
		return fmt.Errorf("transport: Call %q: no active connection (nil stream)", method)
	}
	return t.conn.Call(ctx, method, params, result)
}

// Notify implements Transport.Notify.
func (t *pnTransport) Notify(ctx context.Context, method string, params any) error {
	if t.conn == nil {
		return fmt.Errorf("transport: Notify %q: no active connection (nil stream)", method)
	}
	return t.conn.Notify(ctx, method, params)
}

// OnNotification implements Transport.OnNotification.
func (t *pnTransport) OnNotification(method string, handler func(params json.RawMessage)) {
	if t.conn == nil {
		// no-op: no connection means no notifications will arrive
		return
	}
	t.conn.RegisterNotificationHandler(method, func(_ context.Context, _ string, raw json.RawMessage) {
		handler(raw)
	})
}

// Close implements Transport.Close.
func (t *pnTransport) Close() error {
	if t.conn == nil {
		return nil
	}
	return t.conn.Close()
}

// errorTransport is a Transport that returns a fixed error on every operation.
// Used when NewPowernapTransport cannot establish a connection.
type errorTransport struct {
	err error
}

func (e *errorTransport) Call(_ context.Context, _ string, _, _ any) error {
	return e.err
}

func (e *errorTransport) Notify(_ context.Context, _ string, _ any) error {
	return e.err
}

func (e *errorTransport) OnNotification(_ string, _ func(json.RawMessage)) {}

func (e *errorTransport) Close() error { return nil }
