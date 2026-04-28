package gopls

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestPendingRegistry_RegisterAndDispatch verifies the basic flow of registering
// a channel via Register and delivering a payload via Dispatch.
func TestPendingRegistry_RegisterAndDispatch(t *testing.T) {
	t.Parallel()

	r := &PendingRegistry{}
	ch := r.Register(42)

	payload := json.RawMessage(`{"result":"ok"}`)
	dispatched := r.Dispatch(42, payload)
	if !dispatched {
		t.Fatal("Dispatch() = false, want true")
	}

	select {
	case got := <-ch:
		if string(got) != string(payload) {
			t.Errorf("payload = %s, want %s", got, payload)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("did not receive payload from channel")
	}
}

// TestPendingRegistry_DispatchUnknownID verifies that Dispatch with an unregistered
// ID returns false.
func TestPendingRegistry_DispatchUnknownID(t *testing.T) {
	t.Parallel()

	r := &PendingRegistry{}
	dispatched := r.Dispatch(999, json.RawMessage(`{}`))
	if dispatched {
		t.Error("Dispatch() with unknown ID = true, want false")
	}
}

// TestPendingRegistry_UnregisterRemovesEntry verifies that Dispatch returns false
// after Unregister.
func TestPendingRegistry_UnregisterRemovesEntry(t *testing.T) {
	t.Parallel()

	r := &PendingRegistry{}
	r.Register(10)
	r.Unregister(10)

	dispatched := r.Dispatch(10, json.RawMessage(`{}`))
	if dispatched {
		t.Error("Dispatch() after Unregister = true, want false")
	}
}

// TestPendingRegistry_Concurrent verifies that concurrent goroutines calling
// Register/Dispatch operate without data races (verified by the race detector).
func TestPendingRegistry_Concurrent(t *testing.T) {
	t.Parallel()

	r := &PendingRegistry{}
	const numRequests = 50

	var wg sync.WaitGroup
	results := make([]json.RawMessage, numRequests)

	// Register all channels first.
	channels := make([]<-chan json.RawMessage, numRequests)
	for i := int64(0); i < numRequests; i++ {
		channels[i] = r.Register(i)
	}

	// Start receiver goroutines first.
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			select {
			case got := <-channels[idx]:
				results[idx] = got
			case <-time.After(2 * time.Second):
				t.Errorf("goroutine %d: timeout", idx)
			}
		}(i)
	}

	// Run Dispatch goroutines concurrently.
	var dispatchWg sync.WaitGroup
	for i := int64(0); i < numRequests; i++ {
		dispatchWg.Add(1)
		go func(id int64) {
			defer dispatchWg.Done()
			payload := json.RawMessage(fmt.Sprintf(`{"id":%d}`, id))
			r.Dispatch(id, payload)
		}(i)
	}

	dispatchWg.Wait()
	wg.Wait()

	// Verify that all results were received.
	for i, res := range results {
		if res == nil {
			t.Errorf("result %d is nil", i)
		}
	}
}

// TestNotificationDispatcher_Dispatch verifies that the registered handler is
// invoked for the correct method.
func TestNotificationDispatcher_Dispatch(t *testing.T) {
	t.Parallel()

	d := NewNotificationDispatcher()

	called := false
	var receivedPayload json.RawMessage
	d.Register("textDocument/publishDiagnostics", func(payload json.RawMessage) {
		called = true
		receivedPayload = payload
	})

	payload := json.RawMessage(`{"uri":"file:///main.go","diagnostics":[]}`)
	d.Dispatch("textDocument/publishDiagnostics", payload)

	if !called {
		t.Fatal("handler was not invoked")
	}
	if string(receivedPayload) != string(payload) {
		t.Errorf("payload = %s, want %s", receivedPayload, payload)
	}
}

// TestNotificationDispatcher_UnknownMethod verifies that an unregistered method
// is silently ignored without panic.
func TestNotificationDispatcher_UnknownMethod(t *testing.T) {
	t.Parallel()

	d := NewNotificationDispatcher()
	// Must not panic.
	d.Dispatch("window/logMessage", json.RawMessage(`{"message":"hello"}`))
}

// TestNotificationDispatcher_MultipleHandlers verifies that handlers for
// different methods are registered and invoked independently.
func TestNotificationDispatcher_MultipleHandlers(t *testing.T) {
	t.Parallel()

	d := NewNotificationDispatcher()
	called := map[string]bool{}
	var mu sync.Mutex

	for _, method := range []string{"textDocument/publishDiagnostics", "window/logMessage"} {
		m := method
		d.Register(m, func(payload json.RawMessage) {
			mu.Lock()
			called[m] = true
			mu.Unlock()
		})
	}

	d.Dispatch("textDocument/publishDiagnostics", json.RawMessage(`{}`))
	d.Dispatch("window/logMessage", json.RawMessage(`{}`))

	mu.Lock()
	defer mu.Unlock()
	if !called["textDocument/publishDiagnostics"] {
		t.Error("publishDiagnostics handler was not invoked")
	}
	if !called["window/logMessage"] {
		t.Error("logMessage handler was not invoked")
	}
}

// TestPendingRegistry_Timeout verifies that the channel remains blocked when
// Dispatch is not invoked, and that the wait can be terminated by a context
// cancellation pattern.
func TestPendingRegistry_Timeout(t *testing.T) {
	t.Parallel()

	r := &PendingRegistry{}
	ch := r.Register(99)

	// Without Dispatch, verify with a short timeout that the channel stays blocked.
	select {
	case <-ch:
		t.Fatal("received data from channel without Dispatch")
	case <-time.After(50 * time.Millisecond):
		// timed out as expected
	}
}
