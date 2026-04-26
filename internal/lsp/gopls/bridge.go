package gopls

// @MX:ANCHOR: [AUTO] Bridge — core type for gopls subprocess lifecycle management and diagnostic collection
// @MX:REASON: fan_in >= 3 (NewBridge, GetDiagnostics, Close, readLoop, initialize)
// @MX:WARN: [AUTO] readLoop runs as a goroutine and must be terminated via shutdownCh
// @MX:REASON: goroutine lifetime is bound to Bridge.Close() and risks leaking

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"
)

// circuitBreakerThreshold is the number of consecutive failures required to open the circuit breaker.
// Related to REQ-GB-005: open state after 3 failures.
const circuitBreakerThreshold = 3

// circuitBreakerOpenDuration is the duration the circuit breaker remains in the open state.
const circuitBreakerOpenDuration = 30 * time.Second

// DiagnosticEvent is the event that delivers publishDiagnostics notifications to the internal channel.
type DiagnosticEvent struct {
	URI         string
	Diagnostics []Diagnostic
}

// Bridge manages communication with the gopls subprocess.
// Responsible for subprocess lifecycle, LSP handshake, and diagnostic collection.
//
// All public methods are concurrency-safe.
type Bridge struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser

	writer *Writer
	reader *Reader

	// nextID atomically generates JSON-RPC request IDs.
	nextID atomic.Int64
	// pendingReg manages response channels for in-flight requests.
	pendingReg PendingRegistry
	dispatcher *NotificationDispatcher

	// diagnosticsCh delivers publishDiagnostics notifications to GetDiagnostics.
	// Buffer size 16: older events are discarded on overflow (risk mitigation from plan.md).
	diagnosticsCh chan DiagnosticEvent

	// pendingMu protects access to pendingDiag.
	pendingMu sync.Mutex
	// pendingDiag stores diagnostics that arrive for URIs other than the one collectDiagnostics is waiting on.
	// Design choice: handlePublishDiagnostics stores the latest diagnostics for every URI in pendingDiag.
	// On entry, collectDiagnostics checks pendingDiag[uri] first; non-matching events are also stored in pendingDiag.
	// This resolves the F2 defect where diagnostics for other URIs were permanently lost.
	pendingDiag map[string][]Diagnostic

	// shutdownCh is the signal channel used to terminate readLoop.
	shutdownCh chan struct{}
	// closeOnce ensures shutdownCh is closed exactly once.
	closeOnce sync.Once

	// initialized indicates whether the LSP handshake has completed.
	initialized atomic.Bool

	// Circuit breaker state.
	cbMu          sync.Mutex
	cbFailures    int
	cbOpenUntil   time.Time

	config *Config
}

// NewBridge creates a gopls subprocess, performs the LSP handshake, and returns the Bridge.
// Returns (nil, nil) if the gopls binary is absent or cfg.Enabled=false.
//
// REQ-GB-002: gopls not found → log slog.Warn and return (nil, nil).
// REQ-GB-003: Initialize on explicit call, not on the first GetDiagnostics call.
//             (This implementation initializes at NewBridge call time; lazy init is left as an option.)
func NewBridge(ctx context.Context, projectRoot string, cfg *Config) (*Bridge, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	// REQ-GB-051: return nil if the master switch is false.
	if !cfg.Enabled {
		return nil, nil
	}

	// F5: deep defensive validation — handles direct Config injection that bypasses LoadConfig.
	if err := validateBinary(cfg.Binary); err != nil {
		return nil, fmt.Errorf("gopls: NewBridge: binary validation failed: %w", err)
	}
	if err := validateArgs(cfg.Args); err != nil {
		return nil, fmt.Errorf("gopls: NewBridge: argument validation failed: %w", err)
	}

	// REQ-GB-002: check whether the gopls binary exists.
	goplsPath, err := exec.LookPath(cfg.Binary)
	if err != nil {
		slog.Warn("gopls bridge disabled: binary not found",
			"binary", cfg.Binary,
			"hint", "go install golang.org/x/tools/gopls@latest",
		)
		return nil, nil
	}

	// Start the gopls subprocess.
	args := append([]string{"serve"}, cfg.Args...)
	cmd := exec.CommandContext(ctx, goplsPath, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("gopls: NewBridge: failed to create stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("gopls: NewBridge: failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("gopls: NewBridge: failed to start subprocess: %w", err)
	}

	b := &Bridge{
		cmd:           cmd,
		stdin:         stdin,
		stdout:        stdout,
		writer:        NewWriter(stdin),
		reader:        NewReader(stdout),
		diagnosticsCh: make(chan DiagnosticEvent, 16),
		pendingDiag:   make(map[string][]Diagnostic),
		shutdownCh:    make(chan struct{}),
		config:        cfg,
	}
	b.dispatcher = NewNotificationDispatcher()
	b.dispatcher.Register("textDocument/publishDiagnostics", b.handlePublishDiagnostics)

	// Start the read loop.
	go b.readLoop()

	// Perform the LSP handshake.
	initCtx, cancel := context.WithTimeout(ctx, cfg.InitTimeout)
	defer cancel()
	if err := b.initialize(initCtx, projectRoot); err != nil {
		// Force-kill the process on initialization failure.
		b.forceKill()
		return nil, fmt.Errorf("gopls: NewBridge: initialization failed: %w", err)
	}

	return b, nil
}

// initialize performs the LSP initialization handshake.
// Implements REQ-GB-010, REQ-GB-011, REQ-GB-013.
func (b *Bridge) initialize(ctx context.Context, projectRoot string) error {
	id := b.nextID.Add(1)
	ch := b.pendingReg.Register(id)

	// Convert rootUri to an RFC 3986-compliant file:// URI.
	// Direct string concatenation violates the LSP specification for paths with spaces, Unicode, or Windows drive letters.
	rootURI, err := pathToURI(projectRoot)
	if err != nil {
		b.pendingReg.Unregister(id)
		return fmt.Errorf("gopls: initialize: rootURI conversion failed: %w", err)
	}

	// Configure InitOptions.
	initOpts := map[string]any{"staticcheck": true}
	if b.config != nil && len(b.config.InitOptions) > 0 {
		initOpts = b.config.InitOptions
	}

	params := InitializeParams{
		RootURI: rootURI,
		ClientCapabilities: ClientCapabilities{
			TextDocument: TextDocumentClientCapabilities{
				PublishDiagnostics: PublishDiagnosticsClientCapabilities{
					RelatedInformation: true,
				},
			},
		},
		InitializationOptions: initOpts,
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		b.pendingReg.Unregister(id)
		return fmt.Errorf("gopls: initialize: params serialization failed: %w", err)
	}

	req := Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  "initialize",
		Params:  paramsJSON,
	}
	if err := b.writer.Write(req); err != nil {
		b.pendingReg.Unregister(id)
		return fmt.Errorf("gopls: initialize: request send failed: %w", err)
	}

	// Wait for the initialize response.
	select {
	case raw, ok := <-ch:
		if !ok {
			return fmt.Errorf("gopls: initialize: response channel closed")
		}
		var resp Response
		if err := json.Unmarshal(raw, &resp); err != nil {
			return fmt.Errorf("gopls: initialize: response deserialization failed: %w", err)
		}
		if resp.Error != nil {
			return fmt.Errorf("gopls: initialize: server error %d: %s", resp.Error.Code, resp.Error.Message)
		}
	case <-ctx.Done():
		b.pendingReg.Unregister(id)
		return fmt.Errorf("gopls: initialize: timeout: %w", ctx.Err())
	}

	// Send the initialized notification.
	// REQ-GB-011: send initialized notification after receiving the initialize response.
	notif := Notification{
		JSONRPC: "2.0",
		Method:  "initialized",
		Params:  json.RawMessage(`{}`),
	}
	if err := b.writer.Write(notif); err != nil {
		return fmt.Errorf("gopls: initialize: initialized notification send failed: %w", err)
	}

	b.initialized.Store(true)
	return nil
}

// GetDiagnostics opens filePath in gopls, collects publishDiagnostics notifications, and returns them.
// Returns an error immediately if the circuit breaker is open.
//
// Implements REQ-GB-020, REQ-GB-021, REQ-GB-023.
func (b *Bridge) GetDiagnostics(ctx context.Context, filePath string) ([]Diagnostic, error) {
	// Check the circuit breaker.
	if err := b.checkCircuitBreaker(); err != nil {
		return nil, err
	}

	// Send the didOpen notification.
	// Convert to an RFC 3986-compliant URI to handle paths with spaces, Unicode, and Windows drive letters.
	uri, err := pathToURI(filePath)
	if err != nil {
		return nil, fmt.Errorf("gopls: GetDiagnostics: URI conversion failed: %w", err)
	}
	params := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        uri,
			LanguageID: "go",
			Version:    1,
			Text:       "",
		},
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("gopls: GetDiagnostics: params serialization failed: %w", err)
	}
	notif := Notification{
		JSONRPC: "2.0",
		Method:  "textDocument/didOpen",
		Params:  paramsJSON,
	}
	if err := b.writer.Write(notif); err != nil {
		b.recordFailure()
		return nil, fmt.Errorf("gopls: GetDiagnostics: didOpen send failed: %w", err)
	}

	// Collect publishDiagnostics notifications during the debounce window.
	// REQ-GB-021: wait for diagnostics_debounce_ms.
	return b.collectDiagnostics(ctx, uri)
}

// collectDiagnostics collects diagnostics for uri from diagnosticsCh.
// Returns after the debounce window elapses with no additional events, within the overall ctx timeout.
//
// F2 fix: on entry, check pendingDiag[uri] first to use diagnostics that have already arrived.
// Non-matching URI events received while waiting are consumed from the channel and stored in pendingDiag for later calls.
// handlePublishDiagnostics always refreshes pendingDiag, so data is preserved even on channel overflow.
//
// @MX:WARN: [AUTO] channel-based debounce logic — watch for race conditions around timer reset
// @MX:REASON: a drain select between timer.Stop() and timer.Reset() prevents the race
func (b *Bridge) collectDiagnostics(ctx context.Context, uri string) ([]Diagnostic, error) {
	debounce := b.config.DebounceWindow

	// Check pendingDiag[uri] on entry.
	// Use diagnostics already stored by handlePublishDiagnostics immediately.
	b.pendingMu.Lock()
	if diags, ok := b.pendingDiag[uri]; ok {
		delete(b.pendingDiag, uri)
		b.pendingMu.Unlock()
		// Use the stored diagnostics as the initial value and start the debounce window.
		// Additional events may arrive on the channel, so debounce proceeds as normal.
		return b.collectWithInitial(ctx, uri, diags, debounce)
	}
	b.pendingMu.Unlock()

	// Create the overall timeout context.
	timeoutCtx, cancel := context.WithTimeout(ctx, b.config.Timeout)
	defer cancel()

	// Wait for the first event up to the overall timeout.
	// When an event arrives, switch to the debounce window.
	debounceTimer := time.NewTimer(debounce)
	debounceTimer.Stop() // Not started yet.
	defer debounceTimer.Stop()

	var result []Diagnostic
	received := false

	for {
		select {
		case event := <-b.diagnosticsCh:
			if event.URI != uri {
				// Different URI event: store in pendingDiag and continue waiting.
				b.pendingMu.Lock()
				b.pendingDiag[event.URI] = event.Diagnostics
				b.pendingMu.Unlock()
				continue
			}
			// Already consumed from pendingDiag — remove duplicate entry.
			b.pendingMu.Lock()
			delete(b.pendingDiag, uri)
			b.pendingMu.Unlock()

			result = event.Diagnostics
			if !received {
				received = true
				debounceTimer.Reset(debounce)
			} else {
				// Additional event: reset the debounce timer.
				if !debounceTimer.Stop() {
					select {
					case <-debounceTimer.C:
					default:
					}
				}
				debounceTimer.Reset(debounce)
			}

		case <-debounceTimer.C:
			// No additional events within the debounce window → collection complete.
			b.resetFailures()
			return result, nil

		case <-timeoutCtx.Done():
			if received {
				// Timed out but diagnostics were already received — return them.
				b.resetFailures()
				return result, nil
			}
			b.recordFailure()
			return nil, fmt.Errorf("gopls: GetDiagnostics: timeout")

		case <-b.shutdownCh:
			return nil, fmt.Errorf("gopls: GetDiagnostics: bridge has been closed")
		}
	}
}

// collectWithInitial performs the debounce window starting with an initial set of diagnostics.
// Called when diagnostics are found immediately in pendingDiag.
func (b *Bridge) collectWithInitial(ctx context.Context, uri string, initial []Diagnostic, debounce time.Duration) ([]Diagnostic, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, b.config.Timeout)
	defer cancel()

	debounceTimer := time.NewTimer(debounce)
	defer debounceTimer.Stop()

	result := initial

	for {
		select {
		case event := <-b.diagnosticsCh:
			if event.URI != uri {
				// Different URI event: store in pendingDiag.
				b.pendingMu.Lock()
				b.pendingDiag[event.URI] = event.Diagnostics
				b.pendingMu.Unlock()
				continue
			}
			// Additional event for the same URI: update result and reset the timer.
			result = event.Diagnostics
			if !debounceTimer.Stop() {
				select {
				case <-debounceTimer.C:
				default:
				}
			}
			debounceTimer.Reset(debounce)

		case <-debounceTimer.C:
			b.resetFailures()
			return result, nil

		case <-timeoutCtx.Done():
			b.resetFailures()
			return result, nil

		case <-b.shutdownCh:
			return nil, fmt.Errorf("gopls: GetDiagnostics: bridge has been closed")
		}
	}
}

// handlePublishDiagnostics is the handler for publishDiagnostics notifications.
// Registered with the NotificationDispatcher.
//
// Stores the latest diagnostics for every URI in pendingDiag and delivers them non-blocking to the channel.
// collectDiagnostics checks both the channel and pendingDiag to ensure no diagnostics are lost for any URI.
func (b *Bridge) handlePublishDiagnostics(payload json.RawMessage) {
	var params PublishDiagnosticsParams
	if err := json.Unmarshal(payload, &params); err != nil {
		slog.Warn("gopls: publishDiagnostics deserialization failed", "error", err)
		return
	}

	// Store the latest diagnostics in pendingDiag (overwrite).
	b.pendingMu.Lock()
	b.pendingDiag[params.URI] = params.Diagnostics
	b.pendingMu.Unlock()

	event := DiagnosticEvent(params)
	// Non-blocking send: discard the oldest event if the channel is full.
	// Data is preserved on channel overflow because it was already stored in pendingDiag.
	select {
	case b.diagnosticsCh <- event:
	default:
		// Channel is full: drop one old event and insert the new one.
		select {
		case <-b.diagnosticsCh:
		default:
		}
		b.diagnosticsCh <- event
	}
}

// Close performs graceful gopls shutdown using the LSP shutdown/exit sequence.
// REQ-GB-004: SIGKILL after a 5-second timeout.
//
// F3 fix: after closing shutdownCh, immediately close stdout to unblock reader.Read() in readLoop.
// Closing stdout causes bufio.Reader to return EOF, allowing readLoop to exit instantly without a 5s delay.
//
// F4 fix: use time.NewTimer + defer Stop instead of time.After.
// time.After retains the goroutine for ShutdownTimeout (5s) even if the done branch runs first.
// time.NewTimer + defer Stop releases the timer goroutine immediately when the done branch runs.
func (b *Bridge) Close(ctx context.Context) error {
	var shutdownErr error

	// Only send the shutdown request when initialized.
	if b.initialized.Load() {
		shutdownErr = b.sendShutdown(ctx)
	}

	// Signal readLoop to stop and close stdout to unblock reader.Read().
	// F3: stdout.Close() must be called after close(shutdownCh) for readLoop to exit immediately.
	b.closeOnce.Do(func() {
		close(b.shutdownCh)
		// Close stdout so bufio.Reader.Read() returns EOF.
		// forceKill only closes stdin, so stdout is closed separately here.
		if b.stdout != nil {
			_ = b.stdout.Close()
		}
	})

	// Wait for the process to exit cleanly if it exists.
	if b.cmd != nil {
		done := make(chan error, 1)
		go func() {
			done <- b.cmd.Wait()
		}()

		// F4: time.After → time.NewTimer + defer Stop.
		// When the done branch runs first, the timer goroutine is released immediately.
		shutdownTimeout := b.config.ShutdownTimeout
		timer := time.NewTimer(shutdownTimeout)
		defer timer.Stop()
		select {
		case <-done:
			// Clean exit.
		case <-timer.C:
			// Timeout: send SIGKILL.
			slog.Warn("gopls: shutdown timeout, sending SIGKILL")
			b.forceKill()
		}
	}

	return shutdownErr
}

// sendShutdown sends the LSP shutdown request and exit notification.
func (b *Bridge) sendShutdown(ctx context.Context) error {
	id := b.nextID.Add(1)
	ch := b.pendingReg.Register(id)

	req := Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  "shutdown",
		Params:  json.RawMessage(`null`),
	}
	if err := b.writer.Write(req); err != nil {
		b.pendingReg.Unregister(id)
		return fmt.Errorf("gopls: Close: shutdown request send failed: %w", err)
	}

	// Wait for the shutdown response.
	shutdownCtx, cancel := context.WithTimeout(ctx, b.config.ShutdownTimeout)
	defer cancel()
	select {
	case <-ch:
	case <-shutdownCtx.Done():
		b.pendingReg.Unregister(id)
		slog.Warn("gopls: shutdown response timeout")
	}

	// Send the exit notification.
	exit := Notification{
		JSONRPC: "2.0",
		Method:  "exit",
		Params:  json.RawMessage(`null`),
	}
	if err := b.writer.Write(exit); err != nil {
		return fmt.Errorf("gopls: Close: exit notification send failed: %w", err)
	}

	return nil
}

// readLoop reads messages from gopls stdout and routes them to the pending registry or dispatcher.
// Runs until the Bridge is closed.
//
// @MX:WARN: [AUTO] goroutine — terminated only by shutdownCh closing or reader EOF
// @MX:REASON: goroutine lifetime is bound to Bridge.Close()
func (b *Bridge) readLoop() {
	for {
		select {
		case <-b.shutdownCh:
			return
		default:
		}

		raw, err := b.reader.Read()
		if err != nil {
			if err != io.EOF {
				slog.Debug("gopls: readLoop error (exiting)", "error", err)
			}
			return
		}

		// Parse the message as a Response and route based on whether it has an id.
		var envelope struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
			Error  *ResponseError  `json:"error,omitempty"`
			Result json.RawMessage `json:"result,omitempty"`
			Params json.RawMessage `json:"params,omitempty"`
		}
		if err := json.Unmarshal(raw, &envelope); err != nil {
			slog.Warn("gopls: message parsing failed", "error", err)
			continue
		}

		isNotification := len(envelope.ID) == 0 || string(envelope.ID) == "null"

		if isNotification && envelope.Method != "" {
			// Notification: forward to dispatcher by method.
			b.dispatcher.Dispatch(envelope.Method, envelope.Params)
		} else if !isNotification {
			// Response: parse the id and forward to the pending registry.
			var id int64
			if err := json.Unmarshal(envelope.ID, &id); err != nil {
				slog.Warn("gopls: response ID parsing failed", "raw_id", string(envelope.ID))
				continue
			}
			b.pendingReg.Dispatch(id, raw)
		}
	}
}

// ─── Circuit breaker helpers ───────────────────────────────────────────────────

// checkCircuitBreaker returns an error if the circuit breaker is open.
func (b *Bridge) checkCircuitBreaker() error {
	b.cbMu.Lock()
	defer b.cbMu.Unlock()
	if b.cbFailures >= circuitBreakerThreshold {
		if time.Now().Before(b.cbOpenUntil) {
			return fmt.Errorf("gopls: circuit breaker open (%d consecutive failures, retry available in %v)",
				b.cbFailures, time.Until(b.cbOpenUntil).Round(time.Second))
		}
		// Open period has elapsed: transition to half-open.
		b.cbFailures = 0
	}
	return nil
}

// recordFailure records a failure and opens the circuit breaker when the threshold is exceeded.
func (b *Bridge) recordFailure() {
	b.cbMu.Lock()
	defer b.cbMu.Unlock()
	b.cbFailures++
	if b.cbFailures >= circuitBreakerThreshold {
		b.cbOpenUntil = time.Now().Add(circuitBreakerOpenDuration)
	}
}

// resetFailures resets the circuit breaker failure count.
func (b *Bridge) resetFailures() {
	b.cbMu.Lock()
	defer b.cbMu.Unlock()
	b.cbFailures = 0
}

// forceKill force-terminates the gopls process.
func (b *Bridge) forceKill() {
	if b.cmd != nil && b.cmd.Process != nil {
		_ = b.cmd.Process.Kill()
	}
	if b.stdin != nil {
		_ = b.stdin.Close()
	}
}
