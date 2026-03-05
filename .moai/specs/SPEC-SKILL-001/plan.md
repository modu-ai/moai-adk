# SPEC-SKILL-001: Implementation Plan

| Field       | Value                          |
|-------------|--------------------------------|
| SPEC ID     | SPEC-SKILL-001                 |
| Title       | Open Agent Skills Registry Integration - Implementation Plan |
| Created     | 2026-03-05                     |
| Author      | manager-spec                   |

---

## 1. Implementation Strategy

### 1.1 Approach

This implementation spans two layers:

1. **Go Binary** (`moai-adk-go`): New `moai skill` CLI subcommand with install/publish/list/search/validate functionality. This is the primary deliverable.
2. **Templates** (`internal/template/templates/`): New `moai-skill` workflow skill and default `skill.yaml` configuration template. These are secondary deliverables.

Development methodology: **TDD** (per `quality.yaml: development_mode: tdd`)
- Write failing tests first for each CLI handler function
- Implement until tests pass
- Refactor and validate with `go test -race ./...`

### 1.2 Risk Assessment

| Risk                                          | Probability | Impact  | Mitigation                                          |
|-----------------------------------------------|-------------|---------|-----------------------------------------------------|
| GitHub rate limiting on unauthenticated requests | Medium   | Medium  | Add `GITHUB_TOKEN` env support; cache responses     |
| Skill namespace collisions (owner/name conflicts) | Low      | Medium  | Enforce `owner/skill-name` directory structure      |
| SKILL.md format divergence from open standard | Medium      | Low     | Validator warns on incompatibility; soft errors only |
| Network failures during install               | Medium      | Low     | Clear error messages; no partial file writes        |
| Test isolation: tests writing to project `.claude/` | Low    | High    | Always use `t.TempDir()` for test directories       |

---

## 2. Milestones

### MILE-1: Core CLI (Primary Goal)

**Scope**: `moai skill install`, `moai skill validate`, `moai skill list` (local only)

**Deliverables**:
- `internal/skill/registry.go` — GitHub HTTP client with 3-path URL resolution
- `internal/skill/installer.go` — Download, validate, and write skill files
- `internal/skill/validator.go` — Open standard + MoAI-specific validation
- `internal/cli/skill.go` — CLI handler wiring all subcommands
- `cmd/moai/main.go` — Register `skill` subcommand
- `internal/skill/skill_test.go` — TDD tests for all core functions

**Acceptance Criteria**: AC-INSTALL-001, AC-INSTALL-002, AC-INSTALL-003, AC-VALIDATE-001, AC-VALIDATE-002, AC-LIST-001, AC-COMPAT-003, AC-REGISTRY-001

**Test Coverage Target**: 85% for `internal/skill/` package

### MILE-2: Publish & Search (Secondary Goal)

**Scope**: `moai skill publish`, `moai skill search`, registry configuration

**Deliverables**:
- `internal/skill/manifest.go` — `moai.yaml` parsing and generation
- Updated `internal/cli/skill.go` — Add publish and search subcommands
- `internal/template/templates/.moai/config/sections/skill.yaml` — Default registry config template
- Updated tests for publish and search paths

**Acceptance Criteria**: AC-PUBLISH-001, AC-SEARCH-001, AC-LIST-002, AC-REGISTRY-002, AC-MANIFEST-001, AC-COMPAT-001, AC-COMPAT-002

### MILE-3: Template Integration (Final Goal)

**Scope**: New MoAI skill for the skill workflow; documentation

**Deliverables**:
- `internal/template/templates/.claude/skills/moai/workflows/skill.md` — `moai-skill` user-invocable skill
- Updated `internal/template/templates/.claude/skills/moai/SKILL.md` — Add `skill` to routing table
- `make build` regeneration of embedded templates

**Acceptance Criteria**: All remaining ACs

---

## 3. File Impact Analysis

### New Files

| File                                                                    | Description                       |
|-------------------------------------------------------------------------|-----------------------------------|
| `internal/skill/registry.go`                                           | GitHub registry HTTP client       |
| `internal/skill/installer.go`                                          | Skill download and installation   |
| `internal/skill/validator.go`                                          | Skill validation (open + MoAI)    |
| `internal/skill/manifest.go`                                           | moai.yaml parsing and generation  |
| `internal/skill/skill_test.go`                                         | Tests for skill package           |
| `internal/cli/skill.go`                                                | CLI handler for skill subcommand  |
| `internal/template/templates/.moai/config/sections/skill.yaml`        | Default registry config template  |
| `internal/template/templates/.claude/skills/moai/workflows/skill.md`  | moai-skill workflow skill         |
| `.moai/specs/SPEC-SKILL-001/spec.md`                                   | This SPEC document                |
| `.moai/specs/SPEC-SKILL-001/acceptance.md`                             | Acceptance criteria               |
| `.moai/specs/SPEC-SKILL-001/plan.md`                                   | This implementation plan          |

### Modified Files

| File                                                                    | Change                            |
|-------------------------------------------------------------------------|-----------------------------------|
| `cmd/moai/main.go`                                                     | Register `skill` Cobra subcommand |
| `internal/cli/` (existing file for command registration)               | Wire skill CLI handler            |
| `internal/template/templates/.claude/skills/moai/SKILL.md`            | Add `skill` to intent router      |
| `internal/template/embedded.go`                                        | Auto-regenerated by `make build`  |

---

## 4. Test Strategy

### 4.1 Unit Tests (internal/skill/)

Use table-driven tests for:
- `Validate()`: happy path + each error case
- `Resolve()`: all 3 URL path conventions + 404 handling
- `Install()`: success + overwrite + network failure
- `ParseManifest()`: valid yaml + missing required fields

Use `httptest.NewServer()` for mocking GitHub responses (no real network in unit tests).

### 4.2 Integration Tests

Create a temporary directory with a mock GitHub server to test the full install flow.
Do NOT run integration tests against live GitHub in CI (use build tag `integration`).

### 4.3 Test Isolation Rules (per CLAUDE.local.md)

- All test temp directories via `t.TempDir()` — never write to project `.claude/`
- Use `filepath.Abs()` for user-supplied paths — never `filepath.Join(cwd, absPath)`
- Use `httptest.NewServer()` for HTTP mocking — never real GitHub API in unit tests
- Run `go test -race ./internal/skill/...` to verify concurrency safety

---

## 5. Implementation Order

```
1. Write failing tests for validator.go (RED phase)
2. Implement validator.go (GREEN phase)
3. Write failing tests for registry.go with httptest mock (RED)
4. Implement registry.go (GREEN)
5. Write failing tests for installer.go (RED)
6. Implement installer.go (GREEN)
7. Write failing tests for manifest.go (RED)
8. Implement manifest.go (GREEN)
9. Implement cli/skill.go (wiring)
10. Register in cmd/moai/main.go
11. Refactor (REFACTOR phase) — run go test -race ./...
12. Add template files (MILE-3)
13. Run make build to regenerate embedded.go
14. Final: go test ./... && golangci-lint run
```
