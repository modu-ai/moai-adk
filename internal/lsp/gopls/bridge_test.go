package gopls

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"
)

// ─── Mock gopls helpers ───────────────────────────────────────────────────────

// mockGopls simulates a gopls server using paired pipes.
// It sends pre-prepared JSON-RPC responses in order.
type mockGopls struct {
	// clientReader: the side the bridge (client) reads from → the mock server writes to it
	clientReader *io.PipeReader
	serverWriter *io.PipeWriter
	// serverReader: the mock server reads messages written by the bridge (client)
	serverReader *io.PipeReader
	clientWriter *io.PipeWriter

	mu       sync.Mutex
	received []json.RawMessage // record of messages sent by the bridge
}

// newMockGopls creates a mockGopls connected by pipe pairs.
func newMockGopls() *mockGopls {
	cr, sw := io.Pipe() // server → client direction
	sr, cw := io.Pipe() // client → server direction
	m := &mockGopls{
		clientReader: cr,
		serverWriter: sw,
		serverReader: sr,
		clientWriter: cw,
	}
	// Asynchronously read messages sent by the server and store them in received.
	go m.readLoop()
	return m
}

// readLoop reads messages in the client → server direction and stores them.
func (m *mockGopls) readLoop() {
	r := NewReader(m.serverReader)
	for {
		msg, err := r.Read()
		if err != nil {
			return
		}
		m.mu.Lock()
		m.received = append(m.received, msg)
		m.mu.Unlock()
	}
}

// sendResponse sends a result response for the given id to the bridge.
func (m *mockGopls) sendResponse(id int64, result any) error {
	w := NewWriter(m.serverWriter)
	type response struct {
		JSONRPC string `json:"jsonrpc"`
		ID      int64  `json:"id"`
		Result  any    `json:"result"`
	}
	return w.Write(response{JSONRPC: "2.0", ID: id, Result: result})
}

// sendNotification sends a notification with the given method to the bridge.
func (m *mockGopls) sendNotification(method string, params any) error {
	w := NewWriter(m.serverWriter)
	type notification struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  any    `json:"params"`
	}
	return w.Write(notification{JSONRPC: "2.0", Method: method, Params: params})
}

// close closes the mock server pipes.
func (m *mockGopls) close() {
	_ = m.serverWriter.Close()
	_ = m.clientWriter.Close()
}

// receivedCount returns the number of received messages.
func (m *mockGopls) receivedCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.received)
}

// ─── Bridge tests ─────────────────────────────────────────────────────────────

// newTestBridge creates a Bridge for testing by injecting mock I/O.
// It does not spawn an actual gopls process.
func newTestBridge(mock *mockGopls, cfg *Config) *Bridge {
	if cfg == nil {
		cfg = DefaultConfig()
		cfg.Timeout = 2 * time.Second
		cfg.InitTimeout = 2 * time.Second
		cfg.ShutdownTimeout = 2 * time.Second
		cfg.DebounceWindow = 50 * time.Millisecond
	}
	b := &Bridge{
		writer:        NewWriter(mock.clientWriter),
		reader:        NewReader(mock.clientReader),
		diagnosticsCh: make(chan DiagnosticEvent, 16),
		pendingDiag:   make(map[string][]Diagnostic),
		shutdownCh:    make(chan struct{}),
		config:        cfg,
	}
	b.dispatcher = NewNotificationDispatcher()
	b.dispatcher.Register("textDocument/publishDiagnostics", b.handlePublishDiagnostics)
	return b
}

// TestBridge_InitializeHandshake verifies the LSP initialization handshake.
// REQ-GB-010, REQ-GB-011: initialize request → response → initialized notification order.
func TestBridge_InitializeHandshake(t *testing.T) {
	mock := newMockGopls()
	defer mock.close()

	bridge := newTestBridge(mock, nil)
	go bridge.readLoop()

	// Asynchronously send the initialize response.
	go func() {
		// Wait briefly until the bridge sends the initialize request.
		time.Sleep(20 * time.Millisecond)
		result := InitializeResult{Capabilities: ServerCapabilities{}}
		if err := mock.sendResponse(1, result); err != nil {
			t.Errorf("initialize response send failed: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := bridge.initialize(ctx, "/tmp/test"); err != nil {
		t.Fatalf("initialize error: %v", err)
	}

	// Verify that the bridge sent two messages (initialize + initialized).
	time.Sleep(50 * time.Millisecond)
	if mock.receivedCount() < 2 {
		t.Errorf("received message count = %d, expected at least 2 (initialize + initialized)", mock.receivedCount())
	}
}

// TestBridge_InitializeTimeout verifies the initialize timeout.
// REQ-GB-012: 30-second timeout (the test uses a shorter setting).
func TestBridge_InitializeTimeout(t *testing.T) {
	mock := newMockGopls()
	defer mock.close()

	bridge := newTestBridge(mock, nil)
	go bridge.readLoop()

	// Set a short timeout on ctx. If no response is sent, a timeout must occur.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := bridge.initialize(ctx, "/tmp/test")
	if err == nil {
		t.Error("expected error on timeout, got nil")
	}
}

// TestBridge_GetDiagnostics verifies diagnostics collection.
// REQ-GB-020, REQ-GB-021: receive publishDiagnostics notification after sending didOpen.
func TestBridge_GetDiagnostics(t *testing.T) {
	mock := newMockGopls()
	defer mock.close()

	cfg := DefaultConfig()
	cfg.Timeout = 2 * time.Second
	cfg.InitTimeout = 2 * time.Second
	cfg.ShutdownTimeout = 2 * time.Second
	cfg.DebounceWindow = 30 * time.Millisecond

	bridge := newTestBridge(mock, cfg)
	bridge.initialized.Store(true) // mark as already initialized
	go bridge.readLoop()

	// Platform-independent path/URI construction: on Windows the drive prefix is added,
	// so use the result of pathToURI to match the mock URI.
	filePath := filepath.Join(t.TempDir(), "main.go")
	expectedURI, err := pathToURI(filePath)
	if err != nil {
		t.Fatalf("pathToURI: %v", err)
	}

	// Asynchronously send the publishDiagnostics notification.
	go func() {
		time.Sleep(20 * time.Millisecond)
		params := PublishDiagnosticsParams{
			URI: expectedURI,
			Diagnostics: []Diagnostic{
				{
					Severity: SeverityError,
					Message:  "undefined: foo",
					Source:   "compiler",
					Range: Range{
						Start: Position{Line: 5, Character: 2},
						End:   Position{Line: 5, Character: 5},
					},
				},
				{
					Severity: SeverityWarning,
					Message:  "SA1001: unused variable",
					Source:   "staticcheck",
					Code:     "SA1001",
				},
			},
		}
		if err := mock.sendNotification("textDocument/publishDiagnostics", params); err != nil {
			t.Errorf("publishDiagnostics send failed: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	diags, err := bridge.GetDiagnostics(ctx, filePath)
	if err != nil {
		t.Fatalf("GetDiagnostics error: %v", err)
	}
	if len(diags) != 2 {
		t.Fatalf("diagnostic count = %d, expected 2", len(diags))
	}
	if diags[0].Severity != SeverityError {
		t.Errorf("diags[0].Severity = %v, expected SeverityError", diags[0].Severity)
	}
	if diags[1].Source != "staticcheck" {
		t.Errorf("diags[1].Source = %q, expected staticcheck", diags[1].Source)
	}
}

// TestBridge_GetDiagnostics_Empty verifies that an empty slice is returned for a file with no diagnostics.
func TestBridge_GetDiagnostics_Empty(t *testing.T) {
	mock := newMockGopls()
	defer mock.close()

	cfg := DefaultConfig()
	cfg.Timeout = 500 * time.Millisecond
	cfg.DebounceWindow = 30 * time.Millisecond

	bridge := newTestBridge(mock, cfg)
	bridge.initialized.Store(true)
	go bridge.readLoop()

	// Platform-independent path/URI construction.
	filePath := filepath.Join(t.TempDir(), "clean.go")
	expectedURI, err := pathToURI(filePath)
	if err != nil {
		t.Fatalf("pathToURI: %v", err)
	}

	// Send an empty diagnostics notification.
	go func() {
		time.Sleep(20 * time.Millisecond)
		params := PublishDiagnosticsParams{
			URI:         expectedURI,
			Diagnostics: []Diagnostic{},
		}
		if err := mock.sendNotification("textDocument/publishDiagnostics", params); err != nil {
			t.Errorf("publishDiagnostics send failed: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	diags, err := bridge.GetDiagnostics(ctx, filePath)
	if err != nil {
		t.Fatalf("GetDiagnostics error: %v", err)
	}
	if len(diags) != 0 {
		t.Errorf("diagnostic count = %d, expected 0", len(diags))
	}
}

// TestBridge_CircuitBreaker verifies that the circuit breaker opens after consecutive failures.
// Related to REQ-GB-005: 3 consecutive failures → 30-second open state.
func TestBridge_CircuitBreaker(t *testing.T) {
	mock := newMockGopls()
	defer mock.close()

	cfg := DefaultConfig()
	cfg.Timeout = 30 * time.Millisecond // fast timeout
	cfg.DebounceWindow = 5 * time.Millisecond

	bridge := newTestBridge(mock, cfg)
	bridge.initialized.Store(true)
	go bridge.readLoop()

	ctx := context.Background()
	// Timeout without response → after 3 iterations the circuit breaker opens.
	for i := 0; i < circuitBreakerThreshold; i++ {
		_, _ = bridge.GetDiagnostics(ctx, fmt.Sprintf("/tmp/test/file%d.go", i))
	}

	// If the circuit breaker is open, it should return immediately.
	start := time.Now()
	_, err := bridge.GetDiagnostics(ctx, "/tmp/test/new.go")
	elapsed := time.Since(start)
	if err == nil {
		t.Error("expected error while circuit breaker is open, got nil")
	}
	// The circuit breaker should return immediately (within 20ms).
	if elapsed > 20*time.Millisecond {
		t.Errorf("took too long while circuit breaker is open: %v", elapsed)
	}
}

// TestBridge_GracefulShutdown verifies that Close sends the shutdown/exit sequence.
// REQ-GB-004: shutdown request + exit notification.
func TestBridge_GracefulShutdown(t *testing.T) {
	mock := newMockGopls()
	defer mock.close()

	cfg := DefaultConfig()
	cfg.ShutdownTimeout = 500 * time.Millisecond

	bridge := newTestBridge(mock, cfg)
	bridge.initialized.Store(true)
	go bridge.readLoop()

	// Asynchronously send the shutdown response.
	go func() {
		time.Sleep(20 * time.Millisecond)
		// The shutdown request ID is managed atomically by the bridge, so it may not be 1.
		// Look up the first ID in bridge.pending and respond.
		// Simpler: try response ID 1.
		_ = mock.sendResponse(1, nil)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := bridge.Close(ctx); err != nil {
		t.Fatalf("Close error: %v", err)
	}

	// Verify that shutdown + exit messages were sent.
	time.Sleep(50 * time.Millisecond)
	if mock.receivedCount() < 2 {
		t.Errorf("message count on shutdown = %d, expected at least 2 (shutdown + exit)", mock.receivedCount())
	}
}

// TestBridge_SendRequest_Concurrent verifies that concurrent requests are handled correctly.
func TestBridge_SendRequest_Concurrent(t *testing.T) {
	mock := newMockGopls()
	defer mock.close()

	bridge := newTestBridge(mock, nil)
	go bridge.readLoop()

	const numRequests = 5
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(idx int) {
			// Register the request and immediately send a response.
			id := bridge.nextID.Add(1)
			ch := bridge.pendingReg.Register(id)

			// Async response transmission.
			go func() {
				time.Sleep(10 * time.Millisecond)
				if err := mock.sendResponse(id, map[string]any{"ok": true}); err != nil {
					results <- fmt.Errorf("response send failed: %w", err)
					return
				}
			}()

			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()
			select {
			case <-ch:
				results <- nil
			case <-ctx.Done():
				results <- fmt.Errorf("request %d timed out", idx)
			}
		}(i)
	}

	for i := 0; i < numRequests; i++ {
		if err := <-results; err != nil {
			t.Errorf("concurrent request error: %v", err)
		}
	}
}

// TestNewBridge_MissingGopls verifies that (nil, nil) is returned when the gopls binary is missing.
// REQ-GB-002: gopls missing → return (nil, nil).
func TestNewBridge_MissingGopls(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.Binary = "gopls-nonexistent-binary-xyz"

	ctx := context.Background()
	b, err := NewBridge(ctx, t.TempDir(), cfg)
	if err != nil {
		t.Fatalf("error returned when gopls is missing: %v (expected (nil,nil))", err)
	}
	if b != nil {
		t.Error("Bridge returned when gopls is missing (expected nil)")
	}
}

// TestNewBridge_Disabled verifies that a nil bridge is returned when cfg.Enabled=false.
// REQ-GB-051: when disabled, return nil.
func TestNewBridge_Disabled(t *testing.T) {
	cfg := DefaultConfig() // Enabled = false
	ctx := context.Background()
	b, err := NewBridge(ctx, t.TempDir(), cfg)
	if err != nil {
		t.Fatalf("error returned when disabled: %v", err)
	}
	if b != nil {
		t.Error("non-nil Bridge returned when disabled")
	}
}

// TestBridge_DiagnosticsChannelOverflow verifies that when the diagnostics channel is full,
// older events are discarded and new events are accepted.
func TestBridge_DiagnosticsChannelOverflow(t *testing.T) {
	mock := newMockGopls()
	defer mock.close()

	bridge := newTestBridge(mock, nil)
	// Fill the channel (size 16).
	for i := 0; i < 16; i++ {
		bridge.diagnosticsCh <- DiagnosticEvent{
			URI:         fmt.Sprintf("file:///tmp/other%d.go", i),
			Diagnostics: []Diagnostic{{Message: fmt.Sprintf("err%d", i)}},
		}
	}

	// Send a new event while the channel is full — must be handled without panic.
	newEvent := DiagnosticEvent{
		URI:         "file:///tmp/new.go",
		Diagnostics: []Diagnostic{{Message: "new error"}},
	}
	bridge.handlePublishDiagnostics(mustMarshal(t, PublishDiagnosticsParams(newEvent)))

	// Drain events from the channel and verify that the new event is included.
	found := false
	for {
		select {
		case evt := <-bridge.diagnosticsCh:
			if evt.URI == newEvent.URI {
				found = true
			}
		default:
			goto done
		}
	}
done:
	if !found {
		t.Error("new event missing from channel after overflow handling")
	}
}

// mustMarshal is a test helper that calls Fatal if JSON marshalling fails.
func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}
	return b
}

// TestReadLoop_ExitsPromptlyOnShutdown verifies that readLoop terminates promptly when Close is called.
// F3 defect reproduction: while readLoop is blocked on reader.Read(), even if shutdownCh is closed,
// if stdout is not closed, the goroutine lingers for up to 5 seconds (ShutdownTimeout).
func TestReadLoop_ExitsPromptlyOnShutdown(t *testing.T) {
	mock := newMockGopls()
	// mock.close() is invoked explicitly later, so do not use defer.

	cfg := DefaultConfig()
	cfg.ShutdownTimeout = 5 * time.Second // without F3, this much time elapses

	bridge := newTestBridge(mock, cfg)

	// Start readLoop and track goroutine termination.
	done := make(chan struct{})
	go func() {
		bridge.readLoop()
		close(done)
	}()

	// Give readLoop time to block on reader.Read().
	time.Sleep(20 * time.Millisecond)

	// Close shutdownCh (the action Close() performs).
	bridge.closeOnce.Do(func() {
		close(bridge.shutdownCh)
	})
	// Close stdout to unblock reader.Read() (the core of the F3 fix).
	_ = mock.serverWriter.Close() // close the write side of the pipe the client reads from

	// readLoop must terminate within 100ms (much shorter than the 5s timeout).
	select {
	case <-done:
		// Normal: readLoop terminated.
	case <-time.After(200 * time.Millisecond):
		t.Error("readLoop did not terminate within 200ms (stdout.Close should unblock readLoop)")
	}

	mock.close()
}

// TestClose_DoesNotLeakTimer verifies that the time.NewTimer in Close() does not leak.
// F4 defect reproduction: when using time.After, even if the done branch executes quickly,
// the goroutine lingers for 5s.
func TestClose_DoesNotLeakTimer(t *testing.T) {
	// Create Bridge multiple times and call Close to verify the goroutine count does not blow up.
	// Measure indirectly via runtime.NumGoroutine().
	before := countGoroutines()

	const iterations = 20
	for i := 0; i < iterations; i++ {
		mock := newMockGopls()
		cfg := DefaultConfig()
		cfg.ShutdownTimeout = 100 * time.Millisecond

		bridge := newTestBridge(mock, cfg)
		go bridge.readLoop()

		// Run only closeOnce without shutdown (initialized=false so sendShutdown is not invoked).
		ctx := context.Background()
		_ = bridge.Close(ctx)
		mock.close()
	}

	// Give time for the goroutine count to stabilize.
	time.Sleep(200 * time.Millisecond)
	after := countGoroutines()

	// Even with 20 iterations, goroutine leakage must be within a small bound.
	// time.After keeps the timer goroutine alive for ShutdownTimeout (5s),
	// so without F4 there could be 20+ residual goroutines at this point.
	// After the F4 fix, timer.Stop() releases immediately, so there should be no leak.
	delta := after - before
	if delta > 5 {
		t.Errorf("goroutine leak detected after repeated Close: before=%d, after=%d, delta=%d (expected <= 5)",
			before, after, delta)
	}
}

// countGoroutines returns the current goroutine count.
func countGoroutines() int {
	// runtime.NumGoroutine() returns the number of currently running goroutines.
	return runtime.NumGoroutine()
}

// TestCollectDiagnostics_PreservesOtherURIEvents verifies that other-URI events are not lost during collection.
// F2 defect reproduction: while collectDiagnostics is waiting for file A, if diagnostics for file B arrive,
// then when B is requested afterward the diagnostics must still be returned.
func TestCollectDiagnostics_PreservesOtherURIEvents(t *testing.T) {
	mock := newMockGopls()
	defer mock.close()

	cfg := DefaultConfig()
	cfg.Timeout = 500 * time.Millisecond
	cfg.DebounceWindow = 30 * time.Millisecond

	bridge := newTestBridge(mock, cfg)
	bridge.initialized.Store(true)
	go bridge.readLoop()

	uriA := "file:///tmp/test/a.go"
	uriB := "file:///tmp/test/b.go"

	// Inject the diagnostics for file B directly into the channel first (before waiting for A).
	bridge.diagnosticsCh <- DiagnosticEvent{
		URI:         uriB,
		Diagnostics: []Diagnostic{{Message: "B file error", Severity: SeverityError}},
	}

	// Send the diagnostics for file A after a slight delay.
	go func() {
		time.Sleep(20 * time.Millisecond)
		if err := mock.sendNotification("textDocument/publishDiagnostics", PublishDiagnosticsParams{
			URI:         uriA,
			Diagnostics: []Diagnostic{{Message: "A file error"}},
		}); err != nil {
			t.Errorf("A notification send failed: %v", err)
		}
	}()

	// Collect file A: must succeed.
	ctx := context.Background()
	diagsA, err := bridge.collectDiagnostics(ctx, uriA)
	if err != nil {
		t.Fatalf("collectDiagnostics(A) error: %v", err)
	}
	if len(diagsA) != 1 || diagsA[0].Message != "A file error" {
		t.Errorf("collectDiagnostics(A) = %v, expected one 'A file error' entry", diagsA)
	}

	// Collect file B: should return the event stored in pendingDiag.
	diagsB, err := bridge.collectDiagnostics(ctx, uriB)
	if err != nil {
		t.Fatalf("collectDiagnostics(B) error: %v", err)
	}
	if len(diagsB) != 1 || diagsB[0].Message != "B file error" {
		t.Errorf("collectDiagnostics(B) = %v, expected one 'B file error' entry (other-URI event lost)", diagsB)
	}
}

// TestNewBridge_RealGopls creates and shuts down the bridge if a real gopls binary is available.
// Skips the test if gopls is not on PATH.
func TestNewBridge_RealGopls(t *testing.T) {
	if _, err := exec.LookPath("gopls"); err != nil {
		t.Skip("gopls binary not found, skipping test")
	}

	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.Timeout = 10 * time.Second
	cfg.InitTimeout = 15 * time.Second
	cfg.ShutdownTimeout = 5 * time.Second

	// Use a temporary directory as the project root.
	projectRoot := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	b, err := NewBridge(ctx, projectRoot, cfg)
	if err != nil {
		t.Fatalf("NewBridge error: %v", err)
	}
	if b == nil {
		t.Fatal("NewBridge returned nil")
	}
	defer func() {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer closeCancel()
		if err := b.Close(closeCtx); err != nil {
			t.Logf("Close error (ignored): %v", err)
		}
	}()

	t.Log("real gopls bridge created and initialized successfully")
}
