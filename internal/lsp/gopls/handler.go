package gopls

import (
	"encoding/json"
	"sync"
)

// ─── PendingRegistry ─────────────────────────────────────────────────────────
//
// REQ-GB-033: correlate requests and responses by managing response channels keyed by request id.

// NotificationHandler is the function type that processes a notification payload.
type NotificationHandler func(payload json.RawMessage)

// PendingRegistry is a registry that tracks in-flight requests.
// Uses sync.Map to manage id → response channel mappings in a concurrency-safe manner.
//
// @MX:ANCHOR: [AUTO] Core type for request-response correlation. Used by readLoop and sendRequest in bridge.go.
// @MX:REASON: fan_in >= 3 (Register, Dispatch, Unregister call sites)
type PendingRegistry struct {
		// sync.Map manages map[int64]chan json.RawMessage in a concurrency-safe manner.
	m sync.Map
}

// Register creates and registers a response channel for the given id.
// The returned channel receives the payload when Dispatch is called.
func (r *PendingRegistry) Register(id int64) <-chan json.RawMessage {
	// Buffer size 1: prevents Dispatch from blocking the sender.
	ch := make(chan json.RawMessage, 1)
	r.m.Store(id, ch)
	return ch
}

// Dispatch delivers the payload to the channel for the given id and removes the channel from the registry.
// Returns false if no channel is registered for the id.
func (r *PendingRegistry) Dispatch(id int64, payload json.RawMessage) bool {
	v, ok := r.m.LoadAndDelete(id)
	if !ok {
		return false
	}
	ch := v.(chan json.RawMessage)
	ch <- payload
	return true
}

// Unregister removes the channel for the given id from the registry.
// Called when no longer waiting for a response, e.g. on timeout.
func (r *PendingRegistry) Unregister(id int64) {
	r.m.Delete(id)
}

// ─── NotificationDispatcher ──────────────────────────────────────────────────
//
// REQ-GB-034: route messages without an id (notifications) to registered handlers.

// NotificationDispatcher routes LSP notifications to handlers registered by method.
// The handler map is protected by a mutex.
type NotificationDispatcher struct {
	mu       sync.RWMutex
	handlers map[string]NotificationHandler
}

// NewNotificationDispatcher creates a new NotificationDispatcher.
func NewNotificationDispatcher() *NotificationDispatcher {
	return &NotificationDispatcher{
		handlers: make(map[string]NotificationHandler),
	}
}

// Register registers a handler for the given method.
// Registering the same method twice overwrites the previous handler.
func (d *NotificationDispatcher) Register(method string, handler NotificationHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[method] = handler
}

// Dispatch synchronously calls the handler for the given method.
// Silently ignored (no panic) when no handler is registered.
func (d *NotificationDispatcher) Dispatch(method string, payload json.RawMessage) {
	d.mu.RLock()
	handler, ok := d.handlers[method]
	d.mu.RUnlock()

	if !ok {
		// Unknown notifications are silently ignored.
		return
	}
	handler(payload)
}
