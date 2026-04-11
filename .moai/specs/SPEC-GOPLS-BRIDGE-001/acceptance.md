# SPEC-GOPLS-BRIDGE-001: Acceptance Criteria

## Functional Acceptance

### AC1: Subprocess Lifecycle
- [ ] `gopls.NewBridge(ctx, projectRoot, cfg)` spawns `gopls serve` as child process
- [ ] When `gopls` binary is missing in PATH, `NewBridge` logs a warning with install hint and returns `(nil, nil)` — not an error
- [ ] Subprocess is spawned lazily (only when `GetDiagnostics` is first called, not at `NewBridge`)
- [ ] `Bridge.Close(ctx)` sends `shutdown` request, waits for response (5s timeout), sends `exit` notification, then waits for process exit
- [ ] On 5s timeout, `Close` sends SIGKILL and cleans up pipes

### AC2: LSP Handshake
- [ ] `initialize` request includes `processId`, `rootUri` (file:// + projectRoot), `capabilities.textDocument.publishDiagnostics`
- [ ] `initialize` params include `initializationOptions.staticcheck: true` when config.InitOptions has it
- [ ] `initialized` notification is sent after successful initialize response
- [ ] Initialization timeout is 30 seconds; on timeout, subprocess is killed and `NewBridge` returns error
- [ ] Server capabilities from initialize response are stored in `Bridge` for later reference

### AC3: JSON-RPC Framing (Phase 1 correctness)
- [ ] `Writer.Write(msg)` produces output matching `Content-Length: <n>\r\n\r\n<json>` pattern
- [ ] `Reader.Read()` correctly parses multi-message stream (back-to-back messages)
- [ ] `Reader.Read()` returns `io.EOF` at stream end cleanly
- [ ] `Reader.Read()` returns error on malformed header (missing Content-Length)
- [ ] `Reader.Read()` returns error on truncated body (fewer bytes than Content-Length)
- [ ] Round-trip test: `msg → Write → Read → unmarshal` preserves all fields including Unicode

### AC4: Handler Registry (Phase 3 correctness)
- [ ] `PendingRegistry.Register(id)` returns a channel that receives the matching response
- [ ] `PendingRegistry.Dispatch(id, payload)` delivers to the registered channel
- [ ] Unregistered `id` dispatch returns `false` and logs a warning
- [ ] Concurrent Register/Dispatch passes `go test -race`
- [ ] Notification dispatch routes `textDocument/publishDiagnostics` to the diagnostics handler

### AC5: Diagnostic Collection
- [ ] `GetDiagnostics(ctx, filePath)` sends `textDocument/didOpen` with file content
- [ ] The call waits up to `request_timeout_seconds` (default 5s) for publishDiagnostics
- [ ] Diagnostics received within `diagnostics_debounce_ms` (default 150ms) are batched together
- [ ] Returned `[]Diagnostic` includes severity, source, code, message, range
- [ ] When no diagnostics arrive within timeout, returns `([], nil)` — not an error
- [ ] `textDocument/didClose` is sent after diagnostics collection to free gopls resources

### AC6: Configuration Integration
- [ ] `LoadConfig` reads `.moai/config/sections/lsp.yaml` `lsp.servers.go` entry
- [ ] When `lsp.enabled: false` in config, `gopls.NewBridge` returns `(nil, nil)` without spawning
- [ ] When `lsp.servers.go.enabled: false`, same no-spawn behavior
- [ ] Config respects `lsp.discovery.on_missing: warn_and_skip`
- [ ] Timeouts are read from config; defaults match spec requirements (30s init, 5s request, 5s shutdown, 150ms debounce)

### AC7: No External Dependencies
- [ ] `go mod tidy` shows no new external imports in `internal/lsp/gopls/` packages
- [ ] Only stdlib imports: `encoding/json`, `bufio`, `os/exec`, `context`, `sync`, `sync/atomic`, `log/slog`, `io`, `bytes`, `time`, `fmt`, `errors`
- [ ] No `go.lsp.dev/*` import
- [ ] No `github.com/charmbracelet/powernap` import
- [ ] No `github.com/sourcegraph/jsonrpc2` import

### AC8: GoFeedbackGenerator Integration
- [ ] `Feedback` struct has new field `Diagnostics []gopls.Diagnostic`
- [ ] `GoFeedbackGenerator.Collect(ctx)` populates `Feedback.Diagnostics` when bridge is non-nil
- [ ] When bridge is nil (disabled), `Diagnostics` is empty slice and existing behavior is preserved
- [ ] Ralph `ClassifyFeedback` uses `Diagnostics` for severity/source classification when non-empty
- [ ] Backwards compatibility: existing tests that didn't use Diagnostics continue passing without changes

### AC9: Ralph Classification Enhancement
- [ ] When `Feedback.Diagnostics` contains a diagnostic with `severity=Error` and `source=compiler`, classified as `ErrorLevelBlocker`
- [ ] When `severity=Warning` and `source=staticcheck`, classified as `ErrorLevelApproval` (SA checks are nuanced)
- [ ] When `severity=Information` or `source=vet`, classified as `ErrorLevelAutoFix`
- [ ] The existing integer-based classification path is preserved as fallback when `Diagnostics` is empty

### AC10: Real gopls Integration Test (Phase 9)
- [ ] `TestBridge_RealGopls` skipped gracefully if `gopls` binary is not found
- [ ] Test creates a fixture Go file with intentional type error
- [ ] Bridge spawns real gopls, initializes, opens file, collects diagnostics
- [ ] At least 1 diagnostic is returned with `severity=Error`
- [ ] Bridge closes cleanly within 5 seconds

---

## Quality Acceptance (TRUST 5)

### Tested
- [ ] Unit test coverage ≥ 85% for `internal/lsp/gopls/`
- [ ] `go test -race ./internal/lsp/gopls/` passes
- [ ] Integration test with real gopls (skip-when-missing) is committed
- [ ] Table-driven tests for JSON-RPC framing edge cases (≥ 10 cases)
- [ ] Mock-based bridge tests exercise initialization, diagnostics, shutdown paths

### Readable
- [ ] Each file has a package doc comment explaining its role
- [ ] Exported types have godoc
- [ ] JSON-RPC framing has inline comments explaining the LSP Content-Length convention
- [ ] Bridge lifecycle state transitions are documented

### Unified
- [ ] All gopls bridge code lives in `internal/lsp/gopls/` (not scattered)
- [ ] Config loading reuses existing `internal/config/` patterns where possible
- [ ] Diagnostic types (`gopls.Diagnostic`) align with LSP 3.17 Diagnostic schema

### Secured
- [ ] No shell command execution beyond `exec.Command("gopls", "serve")`
- [ ] File paths from user are resolved via `filepath.Abs` + cleanup before passing to LSP
- [ ] LSP responses are validated before unmarshal (no arbitrary type injection)

### Trackable
- [ ] `@MX:ANCHOR` on `Bridge.GetDiagnostics` (fan_in ≥ 2: go_feedback, future loop integration)
- [ ] Commits reference SPEC-GOPLS-BRIDGE-001 in scope
- [ ] Error messages include context (e.g., `"gopls initialize timeout after 30s"` not just `"timeout"`)

---

## Deliverable Checklist

- [ ] 6 new Go source files (bridge.go, protocol.go, messages.go, handler.go, config.go, and integration_test.go)
- [ ] 5 new test files (bridge_test.go, protocol_test.go, handler_test.go, config_test.go, integration_test.go)
- [ ] 3 modified files (loop/feedback.go, loop/go_feedback.go, ralph/engine.go)
- [ ] 1 modified wiring file (cli/deps.go)
- [ ] `make build` passes
- [ ] `go test -race ./internal/lsp/gopls/... ./internal/loop/... ./internal/ralph/...` passes
- [ ] `go vet ./...` passes
- [ ] Zero new external dependencies in `go.mod`
- [ ] CHANGELOG.md entry referencing SPEC-GOPLS-BRIDGE-001
