# SPEC-LSP-CORE-002: Implementation Plan

## Phase Breakdown

### Phase 1: Dependency + Package Scaffolding

**Goal**: Introduce powernap dependency and create the `internal/lsp/core/` package skeleton.

**Files**:
- `go.mod` (modify): add `github.com/charmbracelet/powernap` at pinned version
- `go.sum` (regenerated): checksums for new dependency
- `internal/lsp/core/doc.go` (new, ~30 LOC): package godoc with architecture overview
- `internal/lsp/core/client.go` (new, ~120 LOC): `Client` interface definition + `ClientState` enum
- `internal/lsp/core/manager.go` (new, ~80 LOC): `Manager` type skeleton with placeholder methods
- `internal/lsp/core/config.go` (new, ~150 LOC): config loader reusing `pkg/models.LSPConfig`
- `internal/lsp/core/errors.go` (new, ~60 LOC): `ErrCapabilityUnsupported`, `ErrServerUnavailable`, wrappers

**Estimated LOC**: +440 new

### Phase 2: powernap Client Wrapper

**Goal**: Adapt powernap's API surface to the internal `Client` interface.

**Files**:
- `internal/lsp/core/powernap_client.go` (new, ~300 LOC):
  - `type powernapClient struct { lang, bin, conn, state, capabilities, openFiles }`
  - `Start(ctx)`: spawn subprocess via powernap, perform `initialize`/`initialized` handshake, store capabilities
  - `Shutdown(ctx)`: LSP `shutdown` + `exit` with timeout fallback to SIGKILL
  - `OpenFile(ctx, path, content)`: send `didOpen` if not cached, `didChange` otherwise
  - `GetDiagnostics(ctx, path)`: ensure open, wait for `publishDiagnostics` within debounce window
  - `FindReferences(ctx, path, position)`: send `textDocument/references` request with capability check
  - `GotoDefinition(ctx, path, position)`: send `textDocument/definition` with capability check
  - State machine transitions logged at slog.Debug
- `internal/lsp/core/capabilities.go` (new, ~150 LOC):
  - Parse `ServerCapabilities` from initialize response
  - `SupportsRename(caps)`, `SupportsReferences(caps)`, etc. helpers
- `internal/lsp/core/powernap_client_test.go` (new, ~250 LOC):
  - Table-driven tests with mock powernap connection
  - State transition coverage
  - Capability-gated query tests

**Estimated LOC**: +700 new

### Phase 3: Manager (Multi-Server Coordinator)

**Goal**: Route requests to the correct `Client` based on file extension and project markers; coordinate lazy spawning and idle cleanup.

**Files**:
- `internal/lsp/core/manager.go` (extend, ~250 LOC):
  - `Start(ctx)` (does NOT spawn servers, reserves lazy init)
  - `GetDiagnostics(ctx, path)`, `FindReferences`, `GotoDefinition`: language detection + client lookup
  - `resolveLanguage(path) (string, error)`: extension + marker file matching
  - `getOrSpawnClient(ctx, lang) (Client, error)`: lazy client spawn with mutex
  - `Close(ctx)`: cascade shutdown to all active clients
- `internal/lsp/core/idle_sweeper.go` (new, ~120 LOC):
  - Background goroutine monitoring `lastActivity` timestamps
  - Gracefully shuts down clients idle beyond `idle_shutdown_seconds`
  - Safe for concurrent `Manager.GetDiagnostics` interleaving
- `internal/lsp/core/manager_test.go` (new, ~300 LOC):
  - Lazy spawn tests with mock client factory
  - Concurrent spawn race tests (`go test -race`)
  - Idle cleanup timing tests
  - Cascade shutdown tests

**Estimated LOC**: +670 new

### Phase 4: Initial Language Validation (Go, Python, TypeScript)

**Goal**: Integration tests with real language servers to validate cross-language capability negotiation and file sync.

**Files**:
- `internal/lsp/core/integration/go_test.go` (new, ~180 LOC):
  - Skip if `gopls` missing via `exec.LookPath`
  - Fixture: `testdata/go_project/main.go` with intentional compiler error
  - Assert `GetDiagnostics` returns >= 1 diagnostic with severity=Error
- `internal/lsp/core/integration/python_test.go` (new, ~180 LOC):
  - Skip if `pylsp` missing
  - Fixture: `testdata/python_project/broken.py` with syntax error
  - Assert diagnostic with severity=Error
- `internal/lsp/core/integration/typescript_test.go` (new, ~180 LOC):
  - Skip if `typescript-language-server` missing
  - Fixture: `testdata/ts_project/bad.ts` with type error
  - Assert diagnostic reflecting TypeScript error
- `internal/lsp/core/integration/coexistence_test.go` (new, ~140 LOC):
  - Spawn all 3 servers in one test run
  - Assert no port/pipe conflicts
  - Assert independent shutdown ordering

**Estimated LOC**: +680 new

### Phase 5: deps.go Wiring + Feature Flag

**Goal**: Expose the Manager to CLI/hook callers behind a feature flag that defaults to the safer gopls_bridge path.

**Files**:
- `internal/cli/deps.go` (modify, ~60 LOC):
  - Read `lsp.enabled` and `lsp.client_impl` from config
  - Branch: `gopls_bridge` (default) → SPEC-GOPLS-BRIDGE-001 bridge; `powernap_core` → new Manager
  - Wire Manager into feedback/hook consumers behind the flag
  - Update existing slog.Warn to describe active client impl
- `internal/lsp/core/factory.go` (new, ~80 LOC):
  - `NewManager(ctx, cfg, projectRoot) (*Manager, error)` public factory
  - Reads per-language `init_options` and wires into each client on spawn
- `internal/cli/deps_test.go` (modify, ~80 LOC):
  - Table-driven tests for feature flag routing
  - Fallback behavior when config missing

**Estimated LOC**: +60 modified + 160 new

### Phase 6: Documentation

**Goal**: Architecture reference for future SPEC authors and onboarding.

**Files**:
- `.claude/rules/moai/core/lsp-client.md` (new, ~200 LOC):
  - Architecture diagram (Manager → Clients → powernap → subprocess)
  - State machine diagram
  - powernap version pinning rationale (REQ-LC-001a)
  - Feature flag semantics (gopls_bridge vs powernap_core)
- `CHANGELOG.md` (modify): add SPEC-LSP-CORE-002 entry

**Estimated LOC**: +200 new + 10 modified

## Total Estimate

| Phase | New LOC | Modified LOC | Notes |
|-------|---------|--------------|-------|
| 1 | 440 | 0 | Scaffolding + interface |
| 2 | 700 | 0 | powernap adapter |
| 3 | 670 | 0 | Manager + idle sweeper |
| 4 | 680 | 0 | Language integration |
| 5 | 160 | 140 | Feature flag wiring |
| 6 | 200 | 10 | Docs + CHANGELOG |
| **Total** | **~2850** | **~150** | — |

Note: frontmatter `estimated_loc: 2500` is a conservative ship target; the breakdown above lists ceiling estimates with test code included.

## Dependencies

- **Hard prerequisite**: SPEC-GOPLS-BRIDGE-001 complete (to validate the general approach before swapping)
- **Blocks**: SPEC-LSP-AGG-003, SPEC-LSP-QGATE-004, SPEC-LSP-LOOP-005, SPEC-LSP-MULTI-006

## Risks

| Risk | Mitigation |
|------|------------|
| powernap API changes | Pin version strictly; REQ-LC-001a enforces re-validation on upgrade |
| Multi-language server config complexity | Phase 4 validates with 3 real servers |
| Backwards compatibility | Feature flag allows rollback to gopls_bridge |
| Race conditions on lazy spawn | Phase 3 has explicit race tests (`go test -race`) |
| Idle sweep killing active clients | Activity timestamp updated on every method call, sweep check is atomic |
