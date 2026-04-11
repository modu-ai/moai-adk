# SPEC-LSP-CORE-002: Acceptance Criteria

## Functional

### AC1: powernap Integration
- [ ] `go.mod` includes `github.com/charmbracelet/powernap` at pinned version (REQ-LC-001)
- [ ] `go mod verify` passes (REQ-LC-001)
- [ ] No breaking import cycles in `internal/lsp/core/` (REQ-LC-001)
- [ ] powernap version pin rationale documented in `.claude/rules/moai/core/lsp-client.md` (REQ-LC-001a)

### AC2: Client Interface
- [ ] `Client` interface defines lifecycle methods `Start(ctx)` and `Shutdown(ctx)` (REQ-LC-002)
- [ ] `Client` interface defines file operation method `OpenFile(ctx, path, content)` (REQ-LC-002a)
- [ ] `Client` interface defines query methods `GetDiagnostics`, `FindReferences`, `GotoDefinition` (REQ-LC-002b)
- [ ] Concrete `powernapClient` implements `Client` (REQ-LC-002, REQ-LC-002a, REQ-LC-002b)
- [ ] All methods accept `context.Context` as first parameter (REQ-LC-002)

### AC3: Manager Coordinator
- [ ] `Manager.Start(ctx)` does NOT eagerly spawn servers (REQ-LC-009)
- [ ] First `Manager.GetDiagnostics(ctx, "foo.go")` spawns gopls lazily (REQ-LC-009)
- [ ] First `Manager.GetDiagnostics(ctx, "bar.py")` spawns pylsp lazily (separate subprocess) (REQ-LC-005, REQ-LC-008)
- [ ] `Manager.Close(ctx)` shuts down all active clients (REQ-LC-008)
- [ ] Idle clients are shut down after `idle_shutdown_seconds` (REQ-LC-050)

### AC4: Document Synchronization
- [ ] First query for an unopened file sends `textDocument/didOpen` with version 1 (REQ-LC-020)
- [ ] Changed content triggers `textDocument/didChange` with incremented version (REQ-LC-021)
- [ ] Idle files are closed via `textDocument/didClose` (REQ-LC-022)
- [ ] `didSave` is sent when caller explicitly requests it (REQ-LC-023)
- [ ] Redundant `didOpen` for cached files is suppressed (REQ-LC-006)

### AC5: Lifecycle State Machine
- [ ] Client transitions through `spawning → initializing → ready` during `Start` (REQ-LC-030)
- [ ] State transitions emit `slog.Debug` log entries (REQ-LC-030)
- [ ] Crashed server enters `degraded` state; queries return empty results (REQ-LC-031)
- [ ] Graceful `Shutdown` completes within configured timeout (REQ-LC-007)
- [ ] Timeout on `Shutdown` escalates to SIGKILL (REQ-LC-007)

### AC6: Capability Negotiation
- [ ] `initialize` request includes ClientCapabilities for publishDiagnostics, references, definition (REQ-LC-032)
- [ ] `ServerCapabilities` stored on client after initialize response (REQ-LC-033)
- [ ] Unsupported operations return `ErrCapabilityUnsupported` (REQ-LC-033)
- [ ] Per-language `init_options` from lsp.yaml are merged into initialize payload (REQ-LC-003a)

### AC7: Graceful Degradation
- [ ] Missing server binary → warn log + return unavailable result, not error (REQ-LC-004)
- [ ] Server crash mid-session → log + return cached or empty results, not panic (REQ-LC-031)
- [ ] Config disabled → `Manager` is nil, callers handle gracefully (REQ-LC-010)

### AC8: Error Handling
- [ ] LSP method errors wrapped with method name, file URI, server language (REQ-LC-040)
- [ ] Request context timeout removes pending request from correlation map (REQ-LC-041)
- [ ] No goroutine leaks on timeout (verified with `goleak` or equivalent) (REQ-LC-041)

### AC9: Integration Tests
- [ ] `TestManager_Go` passes with real gopls (skip-if-missing) (REQ-LC-005, REQ-LC-008)
- [ ] `TestManager_Python` passes with real pylsp (skip-if-missing) (REQ-LC-005, REQ-LC-008)
- [ ] `TestManager_TypeScript` passes with real typescript-language-server (skip-if-missing) (REQ-LC-005, REQ-LC-008)
- [ ] All 3 servers can coexist in a single test run without conflict (REQ-LC-005)

### AC10: Feature Flag
- [ ] `lsp.client_impl: gopls_bridge` (default) routes to SPEC-GOPLS-BRIDGE-001 implementation (REQ-LC-010)
- [ ] `lsp.client_impl: powernap_core` routes to SPEC-LSP-CORE-002 implementation (REQ-LC-010)
- [ ] Runtime flag toggle is respected without restart (REQ-LC-010)

## Quality (TRUST 5)

### Tested
- [ ] ≥ 85% coverage for `internal/lsp/core/`
- [ ] `go test -race ./internal/lsp/core/` passes
- [ ] Mock-based tests + real-server integration tests

### Readable
- [ ] Architecture diagram in `.claude/rules/moai/core/lsp-client.md`
- [ ] Exported types have godoc
- [ ] `Manager.GetDiagnostics` has routing logic explanation

### Unified
- [ ] Config schema matches existing `lsp.yaml` structure (no new schema)
- [ ] Shares `Diagnostic` type with `gopls.Diagnostic` via interface alias

### Secured
- [ ] powernap library reviewed for known CVEs
- [ ] Subprocess spawning uses absolute paths (no PATH injection)

### Trackable
- [ ] `@MX:ANCHOR` on `Manager.GetDiagnostics` (expected fan_in ≥ 4)
- [ ] Commits reference SPEC-LSP-CORE-002

## Deliverables

- [ ] 7+ new Go files in `internal/lsp/core/`
- [ ] 3 integration test files
- [ ] 1 architecture rule doc
- [ ] CHANGELOG entry
- [ ] Zero regressions in existing gopls_bridge path
