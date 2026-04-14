## Task Decomposition
SPEC: SPEC-LSP-CORE-002
Methodology: TDD (RED-GREEN-REFACTOR)
Team Mode: --team (implementer-backend + implementer-transport + tester)

### Approved Decisions (Decision Point 1, 2026-04-12)
- powernap version pin: Semver tag
- Integration tests: local-first, CI skip allowed (`//go:build integration`)
- Default `lsp.client`: LSP-CORE (powernap); GOPLS-BRIDGE opt-in retained (REQ-LC-010)

### Task Table

| Task ID | Description | Requirement | Dependencies | Planned Files | Status |
|---------|-------------|-------------|--------------|---------------|--------|
| T-001 | Pin powernap semver tag + create lsp-client.md rationale | REQ-LC-001, REQ-LC-001a | — | go.mod, go.sum, .claude/rules/moai/core/lsp-client.md | pending |
| T-002 | ServerConfig types + YAML loader (lsp.servers.<language>) | REQ-LC-003, REQ-LC-003a | T-001 | internal/lsp/config/types.go, internal/lsp/config/loader.go, internal/lsp/config/types_test.go, internal/lsp/config/loader_test.go | pending |
| T-003 | subprocess.Launcher with PATH probe + warn_and_skip | REQ-LC-004, REQ-LC-005 | T-002 | internal/lsp/subprocess/launcher.go, internal/lsp/subprocess/launcher_test.go | pending |
| T-004 | subprocess.Supervisor crash detection | REQ-LC-005, REQ-LC-031 | T-003 | internal/lsp/subprocess/supervisor.go, internal/lsp/subprocess/supervisor_test.go | pending |
| T-005 | transport.Transport powernap adapter skeleton | REQ-LC-001 | T-001 | internal/lsp/transport/transport.go, internal/lsp/transport/transport_test.go | pending |
| T-006 | Request correlation + context timeout | REQ-LC-040, REQ-LC-041 | T-005 | internal/lsp/transport/request.go, internal/lsp/transport/request_test.go | pending |
| T-007 | Notification handler (publishDiagnostics) | REQ-LC-002b | T-005 | internal/lsp/transport/notification.go, internal/lsp/transport/notification_test.go | pending |
| T-008 | core.ClientState machine + transitions | REQ-LC-030, REQ-LC-031 | T-004 | internal/lsp/core/state.go, internal/lsp/core/state_test.go | pending |
| T-009 | core.Client interface + Start/Shutdown | REQ-LC-002, REQ-LC-007 | T-005, T-008 | internal/lsp/core/client.go, internal/lsp/core/client_test.go | pending |
| T-010 | ClientCapabilities/ServerCapabilities + ErrCapabilityUnsupported | REQ-LC-032, REQ-LC-033 | T-009 | internal/lsp/core/capabilities.go, internal/lsp/core/errors.go, internal/lsp/core/capabilities_test.go, internal/lsp/core/errors_test.go | pending |
| T-011 | Document cache + didOpen/didChange | REQ-LC-002a, REQ-LC-006, REQ-LC-020, REQ-LC-021 | T-009 | internal/lsp/core/document.go (partial), internal/lsp/core/document_test.go | pending |
| T-012 | didClose (idle 5m) + didSave | REQ-LC-022, REQ-LC-023 | T-011 | internal/lsp/core/document.go (complete), internal/lsp/core/document_test.go | pending |
| T-013 | GetDiagnostics query + error wrapping | REQ-LC-002b, REQ-LC-040 | T-010, T-011 | internal/lsp/core/queries.go (partial), internal/lsp/core/queries_test.go | pending |
| T-014 | FindReferences/GotoDefinition + capability precheck | REQ-LC-002b, REQ-LC-033 | T-013 | internal/lsp/core/queries.go (complete), internal/lsp/core/queries_test.go | pending |
| T-015 | Manager routing (extension + project marker) | REQ-LC-008 | T-014 | internal/lsp/core/manager.go (partial), internal/lsp/core/manager_test.go | pending |
| T-016 | Manager lazy spawn | REQ-LC-009 | T-015 | internal/lsp/core/manager.go (increment) | pending |
| T-017 | Manager idle shutdown (600s default) | REQ-LC-050 | T-016 | internal/lsp/core/manager.go (increment) | pending |
| T-018 | gopls integration test (real subprocess) | REQ-LC-010, validation | T-017 | internal/lsp/core/integration_test.go | pending |
| T-019 | pyright/pylsp integration test | validation | T-018 | internal/lsp/core/integration_python_test.go | pending |
| T-020 | typescript-language-server integration test | validation | T-019 | internal/lsp/core/integration_typescript_test.go | pending |

### Sprint Execution Plan

- **Sprint 1** (Foundation): T-001, T-002, T-005 (partial parallel)
- **Sprint 2** (Subprocess + Transport): T-003, T-004, T-006, T-007 (parallel across roles)
- **Sprint 3** (Core Lifecycle): T-008, T-009, T-010 (sequential)
- **Sprint 4** (Document + Queries): T-011, T-012, T-013, T-014 (sequential within files)
- **Sprint 5** (Manager): T-015, T-016, T-017 (sequential)
- **Sprint 6** (Integration Validation): T-018, T-019, T-020 (sequential)

### Team File Ownership [HARD]

- **implementer-backend**: `internal/lsp/core/**`, `internal/lsp/config/**`, `.claude/rules/moai/core/lsp-client.md`, `go.mod`, `go.sum`
- **implementer-transport**: `internal/lsp/transport/**`, `internal/lsp/subprocess/**`, `internal/cli/deps.go` (opt-in injection only)
- **tester**: all `*_test.go` (unit + integration)
- **Rule**: tester owns all tests exclusively. Production writers must not touch `_test.go`.

### Success Criteria (TRUST 5 verification at Phase 2.5)

- [ ] `go test -race ./internal/lsp/...` passes, zero race warnings
- [ ] Per-package coverage >= 85% (core/transport/subprocess/config)
- [ ] `go test -tags=integration ./internal/lsp/core/...` passes when binaries present
- [ ] `go vet ./internal/lsp/...` clean
- [ ] `golangci-lint run ./internal/lsp/...` clean
- [ ] All REQ-LC-001 through REQ-LC-050 marked implemented
- [ ] powernap pinned via semver tag in go.mod
- [ ] `.claude/rules/moai/core/lsp-client.md` created (rationale + upgrade policy)
- [ ] SPEC-GOPLS-BRIDGE-001 test suite: zero regression
- [ ] MX tags: `@MX:ANCHOR` for fan_in>=3, `@MX:WARN` for goroutines/crash supervisor
- [ ] `internal/lsp/models.go` reused (no duplicate Diagnostic/Position/Range)
- [ ] Go-only boundary: no frontend/DB leakage into `internal/lsp/**`
