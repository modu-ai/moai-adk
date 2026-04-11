# SPEC-LSP-MULTI-006: Acceptance Criteria

## Functional

### AC1: Discovery Engine
- [ ] `discovery.Detect(projectRoot)` returns list of detected languages (REQ-LM-001)
- [ ] Handles multi-language projects (e.g., TypeScript + Python monorepo) (REQ-LM-001, REQ-LM-002)
- [ ] Returns empty slice for empty directory (REQ-LM-001)
- [ ] Respects `project_markers` from lsp.yaml (REQ-LM-001)

### AC2: Binary Resolver
- [ ] Prefers `node_modules/.bin` for JavaScript/TypeScript (REQ-LM-005)
- [ ] Prefers `.venv/bin` for Python (REQ-LM-005)
- [ ] Falls back to PATH when project-local missing (REQ-LM-005)
- [ ] Falls back to `fallback_binaries` in declared order (REQ-LM-008)
- [ ] Returns clear error with install hint when all options exhausted (REQ-LM-004)

### AC3: All 16 Languages Covered
- [ ] Integration test exists for: cpp, csharp, elixir, flutter, go, java, javascript, kotlin, php, python, r, ruby, rust, scala, swift, typescript (REQ-LM-003, REQ-LM-006)
- [ ] Each test skips gracefully when binary missing (`exec.LookPath` check) (REQ-LM-003)
- [ ] Each passing test produces at least 1 diagnostic from fixture error file (REQ-LM-003)
- [ ] CI runs with at least 5 languages installed (Go, Python, TypeScript, Rust, plus 1 rotating) (REQ-LM-003)

### AC4: Doctor Subcommand
- [ ] `moai lsp doctor` prints table of detected languages and server status (REQ-LM-007)
- [ ] `moai lsp doctor --json` produces machine-readable output (REQ-LM-007)
- [ ] Exit code 0 when all detected languages have installed servers (REQ-LM-007)
- [ ] Exit code 1 when missing servers exist (with install hints) (REQ-LM-004, REQ-LM-007)

### AC5: Capability Tracking
- [ ] `Client.Capabilities()` returns parsed ServerCapabilities (REQ-LM-009)
- [ ] Capability-dependent features check before invoking (e.g., rename) (REQ-LM-009)
- [ ] Unsupported operations return structured error, not panic (REQ-LM-009)

### AC6: Language Neutrality (Section 22)
- [ ] All 16 integration test files exist with equivalent structure (REQ-LM-003, REQ-LM-006)
- [ ] No language has priority in the test runner (REQ-LM-006)
- [ ] `lsp doctor` output lists languages alphabetically (REQ-LM-006, REQ-LM-007)
- [ ] Documentation treats all 16 as equal first-class citizens (REQ-LM-006)
- [ ] Aggregator skips unavailable languages without error (REQ-LM-010)

## Quality (TRUST 5)

### Tested
- [ ] 16 per-language integration tests
- [ ] Discovery engine unit tests (≥ 10 scenarios)
- [ ] Doctor subcommand tests (all detected/some missing/all missing)

### Readable
- [ ] `lsp-multi-language.md` architecture doc
- [ ] Each integration test has fixture file comments

### Unified
- [ ] Single `discovery.Engine` used by Manager and Doctor
- [ ] Capability struct shared with `go.lsp.dev/protocol` (if adopted)

### Secured
- [ ] Resolver does not execute user-provided commands
- [ ] Binary paths validated before exec

### Trackable
- [ ] `@MX:ANCHOR` on `discovery.Detect` (fan_in ≥ 3)
- [ ] `@MX:ANCHOR` on `Manager.Spawn`
- [ ] SPEC-LSP-MULTI-006 in commit scopes

## Deliverables

- [ ] 16 integration test files (one per language)
- [ ] Discovery engine (detector + resolver)
- [ ] `moai lsp doctor` subcommand
- [ ] Capability tracking layer
- [ ] Architecture doc
- [ ] README language installation section
- [ ] CHANGELOG entry
