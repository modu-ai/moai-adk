package transport

import (
	"encoding/json"
	"fmt"
	"sync"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
)

// NotificationRouter routes server-initiated JSON-RPC notifications to registered
// handlers by method name.
//
// Each method may have at most one handler (idempotency guard: Register returns
// an error on duplicate registration). Unknown methods are silently ignored, as
// servers may emit out-of-scope notifications.
//
// @MX:ANCHOR: [AUTO] NotificationRouter — central notification dispatcher used by core.Client and document sync
// @MX:REASON: fan_in >= 3 — core.Client.Start, publishDiagnostics consumer, and integration tests all use NotificationRouter
type NotificationRouter struct {
	mu       sync.RWMutex
	handlers map[string]func(params json.RawMessage) error
}

// NewNotificationRouter creates a new, empty NotificationRouter.
func NewNotificationRouter() *NotificationRouter {
	return &NotificationRouter{
		handlers: make(map[string]func(params json.RawMessage) error),
	}
}

// Register adds a handler for the given JSON-RPC notification method.
//
// Returns an error if a handler for method is already registered (idempotency
// guard). To replace a handler, create a new NotificationRouter.
func (r *NotificationRouter) Register(method string, handler func(params json.RawMessage) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.handlers[method]; exists {
		return fmt.Errorf("notification router: handler for %q already registered", method)
	}
	r.handlers[method] = handler
	return nil
}

// Dispatch routes a notification to the registered handler for method.
//
// Unknown methods are silently ignored (return nil) to accommodate servers that
// emit notifications for capabilities beyond our registered set.
//
// Handler errors are propagated to the caller as-is.
func (r *NotificationRouter) Dispatch(method string, params json.RawMessage) error {
	r.mu.RLock()
	h, ok := r.handlers[method]
	r.mu.RUnlock()

	if !ok {
		// Unregistered method — silently ignored (servers may emit notifications beyond our registered set)
		return nil
	}
	return h(params)
}

// Attach registers the router's Dispatch as the notification handler for
// each method in the router on the given Transport.
//
// @MX:NOTE: [AUTO] Attach connects the router to a Transport via Transport.OnNotification
// All notifications are funneled through this router, so the order of Register then Attach matters
func (r *NotificationRouter) Attach(t Transport) {
	r.mu.RLock()
	methods := make([]string, 0, len(r.handlers))
	for m := range r.handlers {
		methods = append(methods, m)
	}
	r.mu.RUnlock()

	for _, method := range methods {
		m := method // loop variable capture
		t.OnNotification(m, func(params json.RawMessage) {
			// Errors are handled by the logging layer; Attach is fire-and-forget
			_ = r.Dispatch(m, params)
		})
	}
}

// publishDiagnosticsParams is the JSON payload for textDocument/publishDiagnostics.
// Reuses the URI field and Diagnostics slice types from internal/lsp/models.go.
type publishDiagnosticsParams struct {
	URI         string           `json:"uri"`
	Diagnostics []lsp.Diagnostic `json:"diagnostics"`
}

// RegisterPublishDiagnostics registers a typed handler for
// textDocument/publishDiagnostics that automatically parses the payload into
// (uri string, diagnostics []lsp.Diagnostic) and invokes callback.
//
// Uses internal/lsp/models.go types — no duplicate type definitions (REQ-LC-002b).
//
// @MX:NOTE: [AUTO] publishDiagnostics parsing — reuses internal/lsp.Diagnostic, no duplicate type definitions
func (r *NotificationRouter) RegisterPublishDiagnostics(
	callback func(uri string, diags []lsp.Diagnostic) error,
) error {
	return r.Register("textDocument/publishDiagnostics", func(params json.RawMessage) error {
		var p publishDiagnosticsParams
		if err := json.Unmarshal(params, &p); err != nil {
			return fmt.Errorf("notification router: publishDiagnostics parse: %w", err)
		}
		return callback(p.URI, p.Diagnostics)
	})
}
