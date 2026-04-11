# SPEC-GOPLS-BRIDGE-001: Implementation Plan

## Phase Breakdown

### Phase 1: JSON-RPC Protocol Layer (`protocol.go`)

**Goal**: Hand-rolled Content-Length framing reader/writer.

**Files**:
- `internal/lsp/gopls/protocol.go` (new, ~150 LOC)
  - `type Writer struct { w io.Writer; buf bytes.Buffer }`
  - `func (w *Writer) Write(msg any) error` — marshal + frame
  - `type Reader struct { r *bufio.Reader }`
  - `func (r *Reader) Read() (json.RawMessage, error)` — parse header + read N bytes
  - Header parsing helper: `parseHeaders(r *bufio.Reader) (int, error)`
- `internal/lsp/gopls/protocol_test.go` (new, ~200 LOC)
  - Round-trip tests: write → read → unmarshal
  - Edge cases: empty body, malformed header, truncated stream, Unicode in body

**Estimated LOC**: +350

### Phase 2: LSP Message Types (`messages.go`)

**Goal**: Minimal LSP 3.17 type definitions needed for the bridge.

**Files**:
- `internal/lsp/gopls/messages.go` (new, ~200 LOC)
  - Request/Notification/Response envelopes
  - `InitializeParams`, `InitializeResult`, `InitializedParams`
  - `DidOpenTextDocumentParams`, `TextDocumentItem`
  - `PublishDiagnosticsParams`, `Diagnostic`, `Range`, `Position`
  - Exit/Shutdown messages
  - Severity enum

**Estimated LOC**: +200

### Phase 3: Handler + Pending Requests (`handler.go`)

**Goal**: Route incoming messages to callers (response) or handlers (notification).

**Files**:
- `internal/lsp/gopls/handler.go` (new, ~200 LOC)
  - `type PendingRegistry struct { sync.Map }`
  - `func (r *PendingRegistry) Register(id int64) <-chan json.RawMessage`
  - `func (r *PendingRegistry) Dispatch(id int64, payload json.RawMessage) bool`
  - `type NotificationDispatcher struct { handlers map[string]NotificationHandler }`
  - `func (d *NotificationDispatcher) Dispatch(method string, payload json.RawMessage)`
- `internal/lsp/gopls/handler_test.go` (new, ~150 LOC)
  - Concurrent Register/Dispatch tests with race detector
  - Timeout handling

**Estimated LOC**: +350

### Phase 4: Bridge Lifecycle (`bridge.go`)

**Goal**: Subprocess spawn, initialize, shutdown, diagnostic collection.

**Concrete Types** (moved here from spec.md per scope boundary rule):

```go
// internal/lsp/gopls/bridge.go
type Bridge struct {
    cmd          *exec.Cmd
    stdin        io.WriteCloser
    stdout       io.ReadCloser
    writer       *Writer          // framed JSON-RPC writer
    reader       *Reader          // framed JSON-RPC reader
    nextID       atomic.Int64
    pending      sync.Map         // id → chan *Response
    diagnostics  chan DiagnosticEvent
    shutdown     chan struct{}
    config       *Config
}

func NewBridge(ctx context.Context, projectRoot string, cfg *Config) (*Bridge, error)
func (b *Bridge) GetDiagnostics(ctx context.Context, filePath string) ([]Diagnostic, error)
func (b *Bridge) Close(ctx context.Context) error
```

**Files**:
- `internal/lsp/gopls/bridge.go` (new, ~350 LOC)
  - `type Bridge struct { cmd, writer, reader, pending, dispatcher, config }`
  - `func NewBridge(ctx, projectRoot, cfg) (*Bridge, error)`
  - `func (b *Bridge) initialize(ctx) error` — LSP handshake
  - `func (b *Bridge) GetDiagnostics(ctx, filePath) ([]Diagnostic, error)`
  - `func (b *Bridge) Close(ctx) error` — graceful shutdown
  - `func (b *Bridge) readLoop()` — goroutine dispatching reader results
  - Circuit breaker for repeated failures (3 strikes, 30s open)
- `internal/lsp/gopls/bridge_test.go` (new, ~300 LOC)
  - Mock gopls via pipe pair (inject scripted responses)
  - Real gopls integration test (skipped if gopls missing)
  - Initialization timeout test
  - Graceful shutdown test
  - Diagnostic collection with debounce

**Estimated LOC**: +650

### Phase 5: Config Loader (`config.go`)

**Goal**: Read settings from `.moai/config/sections/lsp.yaml`.

**Files**:
- `internal/lsp/gopls/config.go` (new, ~100 LOC)
  - `type Config struct { Enabled, Binary, Args, InitOptions, Timeouts }`
  - `func LoadConfig(configPath string) (*Config, error)` — YAML unmarshal
- `internal/lsp/gopls/config_test.go` (new, ~100 LOC)

**Estimated LOC**: +200

### Phase 6: GoFeedbackGenerator Integration

**Goal**: Wire the bridge into Ralph's feedback loop.

**Concrete Types** (moved here from spec.md per scope boundary rule):

```go
// internal/loop/go_feedback.go (updated)
type GoFeedbackGenerator struct {
    projectRoot string
    bridge      *gopls.Bridge  // nil when disabled
}

func (g *GoFeedbackGenerator) Collect(ctx context.Context) (*Feedback, error) {
    fb := &Feedback{
        TestsFailed: g.runGoTest(ctx),
        LintErrors:  g.runGoVet(ctx),
    }
    if g.bridge != nil {
        diags, _ := g.bridge.GetDiagnostics(ctx, g.projectRoot)
        fb.Diagnostics = diags  // new field
    }
    return fb, nil
}
```

**Files**:
- `internal/loop/feedback.go` (modify): add `Diagnostics []gopls.Diagnostic` to `Feedback` struct
- `internal/loop/go_feedback.go` (modify, ~50 LOC):
  - Accept optional `*gopls.Bridge` via constructor
  - Call `bridge.GetDiagnostics` in `Collect()`
  - Populate `Feedback.Diagnostics`
- `internal/loop/go_feedback_test.go` (modify, ~100 LOC): add tests with mock bridge

**Estimated LOC**: +150, modify existing 50

### Phase 7: Ralph ClassifyFeedback Update

**Goal**: Use diagnostic severity/source in classification.

**Files**:
- `internal/ralph/engine.go` (modify, ~50 LOC):
  - Update `ClassifyFeedback` signature (or add overload) to accept diagnostics
  - Add classification branches for severity (error/warning/hint) and source (compiler/staticcheck/vet)
- `internal/ralph/engine_test.go` (modify, ~100 LOC): add diagnostic-based classification tests

**Estimated LOC**: +150 modify

### Phase 8: deps.go Wiring

**Goal**: Inject bridge into dependency graph.

**Files**:
- `internal/cli/deps.go` (modify, ~30 LOC):
  - Load lsp.yaml config
  - If `lsp.enabled && lsp.servers.go.enabled`, create bridge: `gopls.NewBridge(ctx, cwd, cfg)`
  - Pass bridge to `GoFeedbackGenerator` constructor
  - Update the existing `slog.Warn` to report bridge status (enabled/disabled/missing)

**Estimated LOC**: +30 modify

### Phase 9: Integration Test with Real gopls

**Goal**: End-to-end verification when gopls is installed.

**Files**:
- `internal/lsp/gopls/integration_test.go` (new, ~300 LOC)
  - Skipped if `gopls` binary missing (use `exec.LookPath`)
  - Create fixture Go project with intentional errors
  - Spawn bridge, call GetDiagnostics
  - Assert on severity, source, message fields
  - Test graceful shutdown

**Estimated LOC**: +300

---

## Total Estimate

| Phase | Added LOC | Modified LOC | Notes |
|-------|-----------|--------------|-------|
| 1 | 350 | 0 | Protocol framing |
| 2 | 200 | 0 | LSP message types |
| 3 | 350 | 0 | Handler registry |
| 4 | 650 | 0 | Bridge lifecycle |
| 5 | 200 | 0 | Config loader |
| 6 | 150 | 50 | Loop integration |
| 7 | 0 | 150 | Ralph classification |
| 8 | 0 | 30 | deps.go wiring |
| 9 | 300 | 0 | Integration test |
| **Total** | **~2200** | **~230** | — |

**Note**: Original estimate in ultraplan was 200-400 LOC. Revised up to ~2400 total because hand-rolled JSON-RPC framing, handler registry, and tests are more substantial than initially scoped. Still within P1 budget.

---

## Dependencies

- **Hard prerequisite**: Section 22 compliant `lsp.yaml.tmpl` exists (✅ PR #625)
- **Soft prerequisite**: gopls ≥ v0.20.0 installed locally for integration tests
- **Blocks**: None directly; SPEC-LSP-LOOP-005 depends on this for Ralph integration

## Risks

| Risk | Mitigation |
|------|------------|
| Hand-rolled JSON-RPC bugs | Extensive unit tests + real gopls integration test in Phase 9 |
| gopls startup time (200ms-5s) | Lazy initialization; user perception: "first feedback slower" is acceptable |
| Bridge process leak on session crash | Signal handler in deps.go calls `Bridge.Close(ctx)` on shutdown |
| Memory accumulation in debounce buffer | Bounded channel (default 16 slots); overflow drops oldest diagnostics |
| Ralph classification regression | Phase 7 tests lock in existing integer-based behavior as fallback |

---

## Rollout Plan

1. Phases 1-3 (protocol + handler): self-contained, low risk, can merge independently
2. Phases 4-5 (bridge + config): requires Phase 1-3 complete
3. Phases 6-8 (integration): requires Phase 4-5 + review to ensure Ralph doesn't regress
4. Phase 9 (integration test): CI-opt-in, not blocking

**Default state**: `lsp.enabled: false` in config until all phases land and stabilize. Flip to `true` in a follow-up commit.
