package hook

// shutdown_test.go — SPEC-HOOK-TRACE-FLUSH-001 registry teardown coverage.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// lastTraceHandler is a distinct handler type so the entry it produces is
// identifiable by the %T value the registry records in the trace.
type lastTraceHandler struct{ event EventType }

func (h *lastTraceHandler) EventType() EventType { return h.event }

func (h *lastTraceHandler) Handle(_ context.Context, _ *HookInput) (*HookOutput, error) {
	return &HookOutput{}, nil
}

// readTraceHandlers returns the handler field of every entry in a trace file.
func readTraceHandlers(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read trace file %s: %v", path, err)
	}
	var handlers []string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var entry struct {
			Handler string `json:"handler"`
		}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("parse trace line %q: %v", line, err)
		}
		handlers = append(handlers, entry.Handler)
	}
	return handlers
}

// TestRegistryShutdownFlushesLastHandlerEntry asserts that the entry produced
// by the last handler of a dispatch actually reaches disk once teardown
// returns (AC-HTF-007 — REQ-HTF-001, REQ-HTF-008).
//
// This is the direct refutation of the measured defect: the last handler's
// entry is the one that loses the race against process exit, and it is also
// the one carrying the blocking decision. Asserting only that a trace file was
// created would prove nothing — the observed production symptom was a file
// that existed and was zero bytes.
//
// The handler count is large on purpose. Dispatch enqueues entries far faster
// than the writer drains them, so without the flush barrier most entries are
// still queued when teardown returns, and the assertion below fails decisively
// rather than racing.
func TestRegistryShutdownFlushesLastHandlerEntry(t *testing.T) {
	t.Parallel()

	logDir := t.TempDir()
	reg := NewRegistry(config.NewConfigManager())
	reg.EnableObservability(logDir)

	const fillerHandlers = 59
	for range fillerHandlers {
		reg.Register(&mockHandler{event: EventSessionStart, output: &HookOutput{}})
	}
	reg.Register(&lastTraceHandler{event: EventSessionStart})
	wantEntries := fillerHandlers + 1

	const sessionID = "shutdown-flush-last"
	if _, err := reg.Dispatch(context.Background(), EventSessionStart, &HookInput{
		SessionID: sessionID,
	}); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}

	reg.Shutdown()

	handlers := readTraceHandlers(t, filepath.Join(logDir, "trace-"+sessionID+".jsonl"))
	if len(handlers) != wantEntries {
		t.Errorf("trace holds %d entries after teardown, want %d", len(handlers), wantEntries)
	}

	const wantLast = "*hook.lastTraceHandler"
	sawLast := false
	for _, h := range handlers {
		if h == wantLast {
			sawLast = true
			break
		}
	}
	if !sawLast {
		t.Errorf("no entry for the last registered handler %s; got handlers %v", wantLast, handlers)
	}
}

// TestRegistryShutdownNoopWithoutObservability asserts that tearing down a
// registry whose observability was never enabled is a harmless no-op, so the
// CLI can always defer it unconditionally (AC-HTF-005 — REQ-HTF-005).
func TestRegistryShutdownNoopWithoutObservability(t *testing.T) {
	t.Parallel()

	reg := NewRegistry(config.NewConfigManager())

	done := make(chan struct{})
	go func() {
		defer close(done)
		reg.Shutdown()
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Shutdown blocked with no trace writer present, want an immediate no-op")
	}
}

// TestRegistryShutdownIsIdempotent asserts that repeated teardown after a real
// trace writer exists stays safe and prompt (AC-HTF-005).
func TestRegistryShutdownIsIdempotent(t *testing.T) {
	t.Parallel()

	reg := NewRegistry(config.NewConfigManager())
	reg.EnableObservability(t.TempDir())
	reg.Register(&mockHandler{event: EventSessionStart, output: &HookOutput{}})

	if _, err := reg.Dispatch(context.Background(), EventSessionStart, &HookInput{
		SessionID: "shutdown-idempotent",
	}); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		reg.Shutdown()
		reg.Shutdown()
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("repeated Shutdown blocked, want both calls to return")
	}
}
