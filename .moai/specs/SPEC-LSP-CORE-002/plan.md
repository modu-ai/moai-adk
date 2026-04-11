# SPEC-LSP-CORE-002: Implementation Plan

## Phase Breakdown

### Phase 1: Dependency + Package Scaffolding

- Add `github.com/charmbracelet/powernap` to `go.mod` (pinned version)
- Create `internal/lsp/core/` directory
- Define `Client` interface in `client.go`
- Define `Manager` coordinator in `manager.go`
- Config loader in `config.go` (reuse lsp.yaml schema)

**Files**: ~5 new Go files, ~400 LOC

### Phase 2: powernap Client Wrapper

- `internal/lsp/core/client.go`: adapter from powernap API to our `Client` interface
- Handle capability negotiation per language
- Lifecycle: `Start`, `Shutdown`, crash recovery

**Files**: ~2 new + 1 modified, ~600 LOC

### Phase 3: Manager (Multi-Server Coordinator)

- `internal/lsp/core/manager.go`: routes calls by file extension + project markers
- Lazy spawn: server started on first file-of-language request
- Shutdown cascade: `Manager.Close(ctx)` closes all clients

**Files**: ~2 new, ~500 LOC

### Phase 4: Initial Language Validation (Go, Python, TypeScript)

- Integration tests with real gopls, pylsp, typescript-language-server
- Skip-when-missing pattern for each
- Validate capability negotiation across languages

**Files**: ~3 integration test files, ~500 LOC

### Phase 5: deps.go Wiring + Feature Flag

- `internal/cli/deps.go`: read `lsp.enabled` from config; instantiate Manager or keep nil
- Feature flag: `lsp.client_impl: gopls_bridge|powernap_core` (default `gopls_bridge` for safety)

**Files**: ~2 modified, ~100 LOC

### Phase 6: Documentation

- `.claude/rules/moai/core/lsp-client.md`: architecture diagram, usage patterns
- CHANGELOG entry

**Files**: ~2 new docs

## Dependencies

- **Hard prerequisite**: SPEC-GOPLS-BRIDGE-001 complete (to validate the general approach before swapping)
- **Blocks**: SPEC-LSP-AGG-003, SPEC-LSP-QGATE-004, SPEC-LSP-LOOP-005, SPEC-LSP-MULTI-006

## Risks

- powernap API changes: pin version strictly
- Multi-language server config complexity: Phase 4 validates with 3 real servers
- Backwards compatibility: feature flag allows rollback to gopls_bridge
