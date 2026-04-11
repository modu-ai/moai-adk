# SPEC-LSP-MULTI-006: Acceptance Criteria

## Functional

### AC1: Discovery Engine
- [ ] `discovery.Detect(projectRoot)` returns list of detected languages
- [ ] Handles multi-language projects (e.g., TypeScript + Python monorepo)
- [ ] Returns empty slice for empty directory
- [ ] Respects `project_markers` from lsp.yaml

### AC2: Binary Resolver
- [ ] Prefers `node_modules/.bin` for JavaScript/TypeScript
- [ ] Prefers `.venv/bin` for Python
- [ ] Falls back to PATH when project-local missing
- [ ] Falls back to `fallback_binaries` in declared order
- [ ] Returns clear error with install hint when all options exhausted

### AC3: All 16 Languages Covered
- [ ] Integration test exists for: cpp, csharp, elixir, flutter, go, java, javascript, kotlin, php, python, r, ruby, rust, scala, swift, typescript
- [ ] Each test skips gracefully when binary missing (`exec.LookPath` check)
- [ ] Each passing test produces at least 1 diagnostic from fixture error file
- [ ] CI runs with at least 5 languages installed (Go, Python, TypeScript, Rust, plus 1 rotating)

### AC4: Doctor Subcommand
- [ ] `moai lsp doctor` prints table of detected languages and server status
- [ ] `moai lsp doctor --json` produces machine-readable output
- [ ] Exit code 0 when all detected languages have installed servers
- [ ] Exit code 1 when missing servers exist (with install hints)

### AC5: Capability Tracking
- [ ] `Client.Capabilities()` returns parsed ServerCapabilities
- [ ] Capability-dependent features check before invoking (e.g., rename)
- [ ] Unsupported operations return structured error, not panic

### AC6: Language Neutrality (Section 22)
- [ ] All 16 integration test files exist with equivalent structure
- [ ] No language has priority in the test runner
- [ ] `lsp doctor` output lists languages alphabetically
- [ ] Documentation treats all 16 as equal first-class citizens

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
