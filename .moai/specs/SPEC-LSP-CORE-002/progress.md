## SPEC-LSP-CORE-002 Progress

- Started: 2026-04-12
- Methodology: TDD (development_mode: tdd)
- Execution mode: --team (implementer-backend + implementer-transport + tester)
- Harness level: standard (auto)

### Phase 0.9 (Language Detection)
- Detected: Go project (go.mod present)
- Language skill: moai-lang-go
- Status: complete

### Phase 0.95 (Scale-Based Mode Selection)
- Files: ~18 new files
- Domains: 4 packages (core, transport, subprocess, config)
- Mode: Team Mode (user --team flag + complexity threshold met)
- Status: complete

### Phase 1 (Strategy Analysis)
- Agent: manager-strategy
- Output: 20 tasks decomposed, file plan confirmed, YAGNI check clean
- Status: complete

### Decision Point 1 (User Approval)
- powernap pin: Semver tag (approved)
- Integration tests: local-first, CI skip allowed (approved)
- Default lsp.client: LSP-CORE (approved; GOPLS-BRIDGE opt-in via REQ-LC-010)
- Team mode: confirmed
- Status: complete

### Phase 1.5 (Task Decomposition)
- Artifact: tasks.md generated
- 20 tasks, 6 sprints
- Status: complete

### Phase 1.6 (Acceptance Criteria)
- Status: pending

### Phase 1.7 (File Scaffolding)
- Status: pending

### Phase 2B (TDD Implementation)
- Status: Sprint 2 complete (2026-04-12)
- T-001: DONE — powernap v0.1.3 pinned, lsp-client.md created (local + template)
- T-002: DONE — config package 100% coverage, Load + MergeInitOptions
- T-005: DONE — transport package 95.7% coverage, Transport interface + pnTransport skeleton
- T-003: DONE — subprocess.Launcher 87.9% package coverage, ErrBinaryNotFound sentinel, isolated stdio pipes
- T-004: DONE — subprocess.Supervisor race-free, Watch goroutine owns cmd.Wait, multiple watcher support
- T-006: DONE — WrapCallError + CallWithTimeout + ErrRequestTimeout, 100% function coverage
- T-007: DONE — NotificationRouter + RegisterPublishDiagnostics, reuses lsp.Diagnostic, transport 97.0% coverage

### Sprint 4 (Document Sync + Queries): 2026-04-13
- T-011: DONE — documentCache (openOrChange/touch/snapshot/remove) + OpenFile impl + pathToURI/resolveLanguageID helpers
- T-012: DONE — reapIdle + didSave + Client.DidSave + WithIdleTimeout option
- T-013: DONE — GetDiagnostics (push-model cache) + ErrFileNotOpen sentinel + publishDiagnostics handler in Start
- T-014: DONE — FindReferences + GotoDefinition + capability precheck + parseLocations (tolerant decoder)
- Coverage: 94.0% (internal/lsp/core)
- Race detector: PASS

### Sprint 5 (Manager): 2026-04-13
- T-015: DONE — Manager.detectLanguage (extension → language mapping, project marker disambiguation), ErrNoLanguageDetected sentinel, Manager.routeFor
- T-016: DONE — Manager.getOrSpawn (lazy spawn, concurrent-safe: factory called exactly once per language), Start error cleanup
- T-017: DONE — Manager.Start (ctx + cancel + reaper goroutine), Manager.Shutdown (parallel client shutdown + errors.Join aggregation), Manager.reaper (@MX:WARN), Manager.reapIdleClients, WithReaperInterval option
- Coverage: 94.0% (internal/lsp/core — unchanged from Sprint 4 aggregate)
- Race detector: PASS (go test -race ./internal/lsp/core/...)
- go vet: PASS
- Full LSP suite: PASS (go test -race ./internal/lsp/...)
- MX tags: @MX:ANCHOR on ErrNoLanguageDetected, Manager, NewManager, Manager.routeFor; @MX:WARN on reaper goroutine

### Sprint 6 (Integration Validation): 2026-04-13
- T-018: DONE — gopls integration test (4/4 PASS)
  - TestIntegration_Gopls_InitializeAndShutdown: PASS (0.05s)
  - TestIntegration_Gopls_OpenFileAndGetDiagnostics: PASS (0.15s)
  - TestIntegration_Gopls_FindReferences: PASS (0.56s)
  - TestIntegration_Gopls_GotoDefinition: PASS (0.57s)
- T-019: DONE — pyright integration test (2/2 PASS)
  - TestIntegration_Pyright_InitializeAndShutdown: PASS (0.24s)
  - TestIntegration_Pyright_OpenFileAndGetDiagnostics: PASS (0.44s)
- T-020: DONE — typescript integration test (2/2 SKIP, binary absent — expected)
  - TestIntegration_TypeScript_InitializeAndShutdown: SKIP
  - TestIntegration_TypeScript_OpenFileAndDiagnostics: SKIP
- Total integration runtime: ~3.3s
- Race detector: PASS (go test -tags=integration -race ./internal/lsp/core/...)
- go vet: PASS
- Unit test regression: ZERO (go test -race ./internal/lsp/... all PASS)

### Production Code Bug Fixes (revealed by integration tests)
- config.ServerConfig: RootDir 필드 추가 (yaml:"-") — initialize 요청의 rootUri/workspaceFolders 전달
- client.initialize(): rootUri + workspaceFolders 동시 전달 (gopls workspace 활성화)
- client.initialize(): LSP initialized 알림 추가 — 누락 시 gopls가 workspace를 활성화하지 않음
- pathToURI(): filepath.EvalSymlinks 적용 — macOS /var/folders → /private/var/folders 일관성 유지

### Wave 1 Task Summary (All 20 Tasks)
- T-001 through T-007: Sprint 1-2 (Foundation + Transport) — DONE
- T-008 through T-014: Sprint 3-4 (Core Lifecycle + Document Sync) — DONE
- T-015 through T-017: Sprint 5 (Manager) — DONE
- T-018 through T-020: Sprint 6 (Integration Validation) — DONE

### SPEC-LSP-CORE-002 Status: COMPLETED 2026-04-13
