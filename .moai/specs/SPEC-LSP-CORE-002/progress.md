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
