# SPEC-LSP-CORE-002: Acceptance Criteria

## Functional

### AC1: powernap Integration
- [ ] `go.mod` includes `github.com/charmbracelet/powernap` at pinned version
- [ ] `go mod verify` passes
- [ ] No breaking import cycles in `internal/lsp/core/`

### AC2: Client Interface
- [ ] `Client` interface defined with all 6 required methods (Start, OpenFile, GetDiagnostics, FindReferences, GotoDefinition, Shutdown)
- [ ] Concrete `powernapClient` implements `Client`
- [ ] All methods accept `context.Context` as first parameter

### AC3: Manager Coordinator
- [ ] `Manager.Start(ctx)` does NOT eagerly spawn servers
- [ ] First `Manager.GetDiagnostics(ctx, "foo.go")` spawns gopls lazily
- [ ] First `Manager.GetDiagnostics(ctx, "bar.py")` spawns pylsp lazily (separate subprocess)
- [ ] `Manager.Close(ctx)` shuts down all active clients

### AC4: Graceful Degradation
- [ ] Missing server binary → warn log + return unavailable result, not error
- [ ] Server crash mid-session → log + return cached or empty results, not panic
- [ ] Config disabled → `Manager` is nil, callers handle gracefully

### AC5: Integration Tests
- [ ] `TestManager_Go` passes with real gopls (skip-if-missing)
- [ ] `TestManager_Python` passes with real pylsp (skip-if-missing)
- [ ] `TestManager_TypeScript` passes with real typescript-language-server (skip-if-missing)
- [ ] All 3 servers can coexist in a single test run without conflict

### AC6: Feature Flag
- [ ] `lsp.client_impl: gopls_bridge` (default) routes to SPEC-GOPLS-BRIDGE-001 implementation
- [ ] `lsp.client_impl: powernap_core` routes to SPEC-LSP-CORE-002 implementation
- [ ] Runtime flag toggle is respected without restart

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
