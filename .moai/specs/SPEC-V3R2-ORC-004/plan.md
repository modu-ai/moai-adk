# SPEC-V3R2-ORC-004 Implementation Plan

> Implementation plan for **Worktree MUST Rule for write-heavy role profiles**.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `acceptance.md` v0.1.0, `tasks.md` v0.1.0.
> Authored on branch `plan/SPEC-V3R2-ORC-004` (Step 1 plan-in-main; base `origin/main` HEAD `dca57b14d`, rebased mid-session from `3356aa9a9`).
> Run phase will execute on a fresh worktree `feat/SPEC-V3R2-ORC-004` per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline Step 2.

## HISTORY

| Version | Date       | Author        | Description                                                                                                                                                                                                                                  |
|---------|------------|---------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec  | Initial implementation plan per `.claude/skills/moai/workflows/plan.md`. Scope: convert `spec.md` §5 EARS REQs into milestones. Acknowledges **partial-pre-completion** discovered in `research.md` §3 (LR-05 + LR-09 already in agent_lint.go). |

---

## 1. Plan Overview

### 1.1 Goal Restatement

`spec.md` §1 의 핵심 목표를 milestone 분해:

> Upgrade worktree isolation from SHOULD to MUST for write-heavy v3r2 agents (manager-cycle, expert-backend, expert-frontend, expert-refactoring, researcher) and team-mode role profiles (implementer, tester, designer). Promote LR-05 from warning to error severity, add LR-09 read-only protection, document `ORC_WORKTREE_*` sentinel keys, and add `moai workflow lint` for static workflow.yaml enforcement.

`research.md` §3 의 inventory 결과 다음 partial-pre-completion 발견:

- **LR-05 lint rule** (`agent_lint.go:444-491`): 이미 error severity로 구현됨. write-heavy agent 5개 정확히 매칭. 추가 코드 변경 불필요.
- **LR-09 lint rule** (`agent_lint.go:520-535`): 이미 error severity로 구현됨. permissionMode=plan + isolation=worktree 거부. 추가 코드 변경 불필요.
- **`workflow.yaml` role_profiles**: 이미 spec.md REQ-004 와 100% 정합 (implementer/tester/designer = worktree, 그 외 = none).
- **0 agents** 이 현재 `isolation: worktree` 를 frontmatter에 선언함 — 4개 기존 에이전트 (expert-backend, expert-frontend, expert-refactoring, researcher) + ORC-001 dependency manager-cycle 모두 미반영.
- **`worktree-integration.md` line 135**: 여전히 SHOULD verbiage. 표준 SHOULD → MUST 업그레이드 필요.

이로 인한 **Plan-phase delta** (research.md §3 from baseline → 100% spec coverage):

1. `internal/template/templates/.claude/agents/moai/expert-backend.md` frontmatter 에 `isolation: worktree` 추가.
2. `internal/template/templates/.claude/agents/moai/expert-frontend.md` frontmatter 에 `isolation: worktree` 추가.
3. `internal/template/templates/.claude/agents/moai/expert-refactoring.md` frontmatter 에 `isolation: worktree` 추가.
4. `internal/template/templates/.claude/agents/moai/researcher.md` frontmatter 에 `isolation: worktree` 추가 + body line "All experiments in worktree isolation when possible" → "All experiments in worktree isolation (mandatory per SPEC-V3R2-ORC-004)" 변경.
5. `internal/template/templates/.claude/agents/moai/manager-cycle.md` frontmatter 에 `isolation: worktree` 추가 (조건부 — ORC-001 머지 후).
6. `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` line 135 SHOULD → MUST 업그레이드 + `manager-cycle, expert-backend, expert-frontend, expert-refactoring, researcher` 명시.
7. 신규 CLI: `moai workflow lint` (~150 LOC) — `workflow.yaml` role_profiles 검증, `ORC_WORKTREE_REQUIRED` sentinel.
8. `agent_lint.go` 의 LR-05/LR-09 violation message에 sentinel key (`ORC_WORKTREE_MISSING`, `ORC_WORKTREE_ON_READONLY`) 추가.
9. `agent_lint_test.go` 에 AC-06/AC-07 fixture 추가.
10. `make build && make install` 후 `moai update` 로 .claude/ 미러링.
11. CHANGELOG entry + MX tag annotations.

핵심 deltas (research.md §3/§4/§6 cross-reference):

- **Frontmatter 추가**: 4 existing agents + 1 conditional (manager-cycle). Template-first → make build → moai update 순서 강제.
- **Researcher body 정정**: P-A22 self-contradiction 해결 (frontmatter↔body 일관성).
- **Rule text 업그레이드**: SHOULD → MUST + 정확한 5개 v3r2 agent 명시.
- **Sentinel keys 명시**: `ORC_WORKTREE_REQUIRED` (workflow lint), `ORC_WORKTREE_MISSING` (LR-05), `ORC_WORKTREE_ON_READONLY` (LR-09).
- **`moai workflow lint`**: 신규 CLI, `workflow.yaml` role_profiles 의 isolation 키를 검증.

### 1.2 Implementation Methodology

`.moai/config/sections/quality.yaml`: `development_mode: tdd`. Run phase MUST follow RED → GREEN → REFACTOR per `.claude/rules/moai/workflow/spec-workflow.md` § Run Phase.

- **RED**: 신규 test 작성. (a) `agent_lint_test.go` 에 AC-06 fixture (expert-backend frontmatter에서 isolation 제거 시 LR-05 + ORC_WORKTREE_MISSING sentinel 메시지 검증), AC-07 fixture (evaluator-active 에 isolation: worktree 추가 시 LR-09 + ORC_WORKTREE_ON_READONLY sentinel 메시지 검증). (b) 신규 `internal/cli/workflow_lint_test.go` 생성: workflow.yaml 의 implementer.isolation을 none으로 변경 시 ORC_WORKTREE_REQUIRED 에러 검증, designer/tester 동일. (c) RED 단계에서는 sentinel 메시지가 아직 없으므로 모든 신규 test FAIL.
- **GREEN**: §1.1 의 11 deltas 구현. 4 frontmatter additions, 1 rule text edit, 1 researcher body edit, 1 새 CLI command, 2 sentinel message updates. 모든 RED test PASS.
- **REFACTOR**: lint engine helper 정리 (sentinel key를 const로 추출), workflow_lint.go 의 yaml parse 로직을 internal/config 의 기존 SystemConfig 패턴과 정합.

### 1.2.1 Acknowledged Discrepancies

본 plan 이 spec.md 와 의도적으로 다르게 처리하는 부분:

- **spec.md §1 "6 cross-file-write agents"** (R5 raw count) vs **plan §1.1 "5 v3r2 agents"** (post-ORC-001):
  R5 audit 의 6 agents = {manager-ddd, manager-tdd, expert-backend, expert-frontend, expert-refactoring, researcher}. ORC-001 retire mechanism 으로 manager-ddd + manager-tdd → manager-cycle 통합. 결과 v3r2 roster 의 write-heavy = 5 agents. spec.md §2.1 도 5개를 enumerate (manager-cycle 포함). plan은 v3r2 reality 의 5개를 binding count 로 사용.

- **REQ-008 "spawn wrapper" 의 enforcement layer**:
  spec.md REQ-008 은 "spawn wrapper shall verify ... a mismatch shall result in a structured blocker report". research.md §2.6 은 Go-side 의 spawn wrapper 가 존재하지 않음을 확인. 이 plan은 enforcement layer를 **static CI gate** (`moai workflow lint`)로 구현. 런타임 spawn-time 가로채기는 Claude Code hook architecture 변경이 필요한 future work이며 ORC-004 scope 밖. spec.md AC-09 는 "Attempting to spawn ... fails with blocker report `ORC_WORKTREE_REQUIRED`" 라고 표현되었지만, 본 plan 의 AC-09 는 "Modifying workflow.yaml to remove implementer.isolation=worktree causes `moai workflow lint` to fail with `ORC_WORKTREE_REQUIRED`" 으로 협소화 (acceptance.md 참조).

- **`spec.md` §2.1 "Implement a write-heavy classifier in agent_lint.go"** vs **research.md §2.2 already implemented**:
  classifier 는 이미 구현되어 있음. plan 은 message-text 정정 (sentinel key 추가) 만 수행.

### 1.3 Deliverables

| Deliverable | Path | REQ Coverage |
|-------------|------|--------------|
| Add `isolation: worktree` to expert-backend | `internal/template/templates/.claude/agents/moai/expert-backend.md` (1 line) | REQ-002, AC-02 |
| Add `isolation: worktree` to expert-frontend | `internal/template/templates/.claude/agents/moai/expert-frontend.md` (1 line) | REQ-002, AC-02 |
| Add `isolation: worktree` to expert-refactoring | `internal/template/templates/.claude/agents/moai/expert-refactoring.md` (1 line) | REQ-002, AC-02 |
| Add `isolation: worktree` to researcher | `internal/template/templates/.claude/agents/moai/researcher.md` (1 line frontmatter + 1 line body) | REQ-002, REQ-011, AC-02, AC-08 |
| Add `isolation: worktree` to manager-cycle | `internal/template/templates/.claude/agents/moai/manager-cycle.md` (1 line, post-ORC-001) | REQ-002, AC-02 |
| Update SHOULD → MUST in worktree-integration.md | `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` line 135 (1 line replacement) | REQ-001, AC-01 |
| Document `ORC_WORKTREE_*` sentinel keys | `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` (new section ~20 LOC) | REQ-006, REQ-008, REQ-013, REQ-014 |
| Sentinel key on LR-05 message | `internal/cli/agent_lint.go` line ~485 (extend message string) | REQ-013, AC-06 |
| Sentinel key on LR-09 message | `internal/cli/agent_lint.go` line ~532 (extend message string) | REQ-014, AC-07 |
| New `moai workflow lint` CLI | `internal/cli/workflow_lint.go` (new ~150 LOC) | REQ-008, AC-09 |
| New `moai workflow lint` tests | `internal/cli/workflow_lint_test.go` (new ~120 LOC) | REQ-008, AC-09 |
| AC-06 fixture in agent_lint_test.go | `internal/cli/agent_lint_test.go` (extend ~40 LOC) | REQ-013, AC-06 |
| AC-07 fixture in agent_lint_test.go | `internal/cli/agent_lint_test.go` (extend ~40 LOC) | REQ-014, AC-07 |
| MX tags per §6 | 4 files (per §6 below) | mx_plan |
| CHANGELOG entry | `CHANGELOG.md` Unreleased section | Trackable (TRUST 5) |
| Template parity verified | `diff -r .claude/agents/moai/ internal/template/templates/.claude/agents/moai/` returns 0 diff | REQ-005 |
| Workflow lint integrated into agent group | `internal/cli/agent_lint.go` rootCmd registration extends to `workflow lint` subcommand | REQ-008 |

Embedded-template parity is **applicable** because all `.claude/agents/moai/*.md` and `.claude/rules/moai/workflow/worktree-integration.md` changes flow through `internal/template/templates/`. `make build && make install && moai update` is mandatory.

### 1.4 Traceability Matrix (REQ → AC → Task)

Per plan-auditor PASS criterion #2 (every REQ maps to at least one AC and at least one Task). Built **after** tasks.md was finalized; each row references actual T-ORC004-NN IDs.

| REQ | Description | Maps to AC | Implementation Task |
|-----|-------------|-----------|--------------------|
| REQ-V3R2-ORC-004-001 | worktree-integration.md SHOULD → MUST | AC-01 | T-ORC004-04 |
| REQ-V3R2-ORC-004-002 | 5 agents declare `isolation: worktree` | AC-02 | T-ORC004-05, T-ORC004-06, T-ORC004-07, T-ORC004-08, T-ORC004-09 |
| REQ-V3R2-ORC-004-003 | Read-only agents NOT declare `isolation: worktree` | AC-03 | T-ORC004-10 (verification only) |
| REQ-V3R2-ORC-004-004 | workflow.yaml role_profiles aligned | AC-04 | T-ORC004-11 (verification only) |
| REQ-V3R2-ORC-004-005 | Template-first mirror | (CI gate) | T-ORC004-19 |
| REQ-V3R2-ORC-004-006 | LR-05 promoted from warning to error | AC-05, AC-06 | T-ORC004-13, T-ORC004-14 (sentinel msg) |
| REQ-V3R2-ORC-004-007 | Write-heavy classifier | AC-06 | T-ORC004-13 (verification only — already implemented) |
| REQ-V3R2-ORC-004-008 | Spawn-time validation | AC-09 | T-ORC004-16, T-ORC004-17 (workflow lint) |
| REQ-V3R2-ORC-004-009 | Lint check active | AC-05, AC-06 | T-ORC004-13, T-ORC004-14 |
| REQ-V3R2-ORC-004-010 | Path rules retained | AC-10 | T-ORC004-22 (skill body verification) |
| REQ-V3R2-ORC-004-011 | Researcher body line update | AC-08 | T-ORC004-08 |
| REQ-V3R2-ORC-004-012 | manager-cycle test isolation | (informational) | T-ORC004-09 |
| REQ-V3R2-ORC-004-013 | LR-05 fail with `ORC_WORKTREE_MISSING` | AC-06 | T-ORC004-13, T-ORC004-14 |
| REQ-V3R2-ORC-004-014 | LR-09 fail with `ORC_WORKTREE_ON_READONLY` | AC-07 | T-ORC004-15 |
| REQ-V3R2-ORC-004-015 | Absolute path prompt rejection | AC-10 | T-ORC004-22 (skill body verification) |

### 1.5 Interface Signatures

#### 1.5.1 New CLI subcommand `moai workflow lint`

```go
// internal/cli/workflow_lint.go (new)

// WorkflowLintViolation represents a workflow.yaml lint rule violation.
type WorkflowLintViolation struct {
    Rule        string `json:"rule"`         // e.g. "ORC_WORKTREE_REQUIRED"
    Severity    string `json:"severity"`     // "error" | "warning"
    Path        string `json:"path"`         // YAML path, e.g. "workflow.team.role_profiles.implementer.isolation"
    Expected    string `json:"expected"`     // expected value
    Actual      string `json:"actual"`       // actual value
    Message     string `json:"message"`
}

// runWorkflowLint validates .moai/config/sections/workflow.yaml.
// Returns errLintViolations (cobra-friendly) when violations are found.
func runWorkflowLint(cmd *cobra.Command, _ []string) error

// validateRoleProfiles checks role_profiles.{implementer,tester,designer}.isolation == "worktree".
func validateRoleProfiles(cfg *workflowConfig) []WorkflowLintViolation

// loadWorkflowYAML reads and parses .moai/config/sections/workflow.yaml into a typed struct.
// Reuses internal/config patterns where possible.
func loadWorkflowYAML(path string) (*workflowConfig, error)
```

CLI shape:

```
moai workflow lint [--format text|json] [--strict]
  Exit codes:
    0: No violations
    1: Violations found
    2: Malformed workflow.yaml
    3: IO error
```

The command is registered as a sibling of `moai agent lint` under a new `moai workflow` subcommand group. The group is added to `rootCmd` with `GroupID: "tools"`.

#### 1.5.2 Sentinel key constants

```go
// internal/cli/agent_lint.go (extend)

const (
    // Sentinel keys per SPEC-V3R2-ORC-004 REQ-013/014.
    SentinelWorktreeMissing      = "ORC_WORKTREE_MISSING"
    SentinelWorktreeOnReadonly   = "ORC_WORKTREE_ON_READONLY"
)
```

```go
// internal/cli/workflow_lint.go (new)

const (
    SentinelWorktreeRequired = "ORC_WORKTREE_REQUIRED"
)
```

LR-05 violation message becomes:

> `Write-heavy agent 'expert-backend' must have 'isolation: worktree' (SPEC-V3R2-ORC-004 ORC_WORKTREE_MISSING)`

LR-09 violation message becomes:

> `Read-only agent (permissionMode: plan) MUST NOT have 'isolation: worktree' — plan mode already prevents writes (SPEC-V3R2-ORC-004 ORC_WORKTREE_ON_READONLY)`

`moai workflow lint` violation message:

> `role_profiles.implementer.isolation must be 'worktree' (got 'none') (SPEC-V3R2-ORC-004 ORC_WORKTREE_REQUIRED)`

#### 1.5.3 workflowConfig struct

Internal type matching the relevant subset of `workflow.yaml`:

```go
type workflowConfig struct {
    Workflow struct {
        Team struct {
            RoleProfiles map[string]struct {
                Mode        string `yaml:"mode"`
                Model       string `yaml:"model"`
                Isolation   string `yaml:"isolation"`
                Description string `yaml:"description"`
            } `yaml:"role_profiles"`
        } `yaml:"team"`
    } `yaml:"workflow"`
}
```

This type is local to `internal/cli/workflow_lint.go` to avoid coupling with the broader internal/config.SystemConfig.

---

## 2. Milestones (Priority-Ordered)

Milestones are ordered by priority (P0 first). No time estimates per CLAUDE.md §Time Estimation HARD rule.

### M1 — Frontmatter Additions (P0)

Add `isolation: worktree` to 4 existing v3r2 agents (template-first).

- Tasks: T-ORC004-05 .. T-ORC004-08
- Acceptance: AC-02 (4 of 5 agents declare worktree)
- Rationale: Direct delivery on REQ-002. Mechanical edits, low risk.
- Dependencies: None.

### M2 — manager-cycle Frontmatter (P0, conditional)

Add `isolation: worktree` to manager-cycle (created by ORC-001).

- Tasks: T-ORC004-09
- Acceptance: AC-02 (5/5 agents declare worktree)
- Rationale: Completes REQ-002 coverage.
- Dependencies: SPEC-V3R2-ORC-001 must be merged. If not merged at run-phase start, T-ORC004-09 emits a structured blocker via `manager-spec` → orchestrator. Run phase pauses; orchestrator runs AskUserQuestion {wait, proceed without manager-cycle, abort}.

### M3 — Rule Text Update (P0)

Update `worktree-integration.md` §HARD Rules line 135 SHOULD → MUST and add sentinel key documentation section.

- Tasks: T-ORC004-04, T-ORC004-12
- Acceptance: AC-01
- Rationale: REQ-001 normative text.
- Dependencies: M1 (so the rule text matches the actual frontmatter state).

### M4 — Researcher Body Reconciliation (P1)

Update researcher.md body line "All experiments in worktree isolation when possible" → "All experiments in worktree isolation (mandatory per SPEC-V3R2-ORC-004)".

- Tasks: T-ORC004-08 (combined with frontmatter)
- Acceptance: AC-08
- Rationale: P-A22 self-contradiction resolution.
- Dependencies: M1.

### M5 — Lint Sentinel Messages (P1)

Extend LR-05 / LR-09 violation messages with sentinel keys.

- Tasks: T-ORC004-13, T-ORC004-14, T-ORC004-15
- Acceptance: AC-06, AC-07
- Rationale: REQ-013, REQ-014.
- Dependencies: None (independent code path).

### M6 — `moai workflow lint` Implementation (P1)

Implement new CLI subcommand. RED-GREEN-REFACTOR cycle.

- Tasks: T-ORC004-16, T-ORC004-17, T-ORC004-18
- Acceptance: AC-09
- Rationale: REQ-008 static enforcement layer.
- Dependencies: None.

### M7 — Template Mirror & Build (P0)

`make build && make install && moai update` to mirror template changes to local tree.

- Tasks: T-ORC004-19
- Acceptance: AC-02 verification on `.claude/agents/moai/*.md` (not just templates)
- Rationale: REQ-005 template-first.
- Dependencies: M1, M2, M3, M4 complete.

### M8 — Verification & Documentation (P2)

Run all lint commands, full test suite, update CHANGELOG, add MX tags.

- Tasks: T-ORC004-20, T-ORC004-21, T-ORC004-22
- Acceptance: AC-05, AC-10
- Rationale: TRUST 5 quality gates, MX tag protocol.
- Dependencies: M1-M7 complete.

---

## 3. Migration Strategy

### 3.1 Run-Time Backwards Compatibility

`spec.md` declares `breaking: false` and `bc_id: []`. The change adds an *enforcement* mechanism rather than altering an interface. Backward compatibility analysis:

- **Existing agent invocations**: Agents currently invoked without `isolation: worktree` still execute correctly — Claude Code's runtime defaults the field to `none`. Adding the field to frontmatter does not break existing flows; it changes the *behavior* (CWD becomes ephemeral worktree) but the interface remains identical.
- **Lint engine**: Pre-existing behavior of `moai agent lint` did not return errors for these 4 agents (no `isolation: worktree` declared, but the agents were not in the LR-05 hardcoded list at the time? — research.md §2.2 confirms LR-05 was already hardcoded with the 5 v3r2 agents). Therefore, contributors who added these agents to PR queues *before* ORC-004 merges already see LR-05 errors today. ORC-004 does not introduce a new failure mode for currently-merged code; it mandates the fix that contributors should already have applied.
- **CI failure for new-PR contributors**: Any PR that introduces or modifies one of the 5 v3r2 write-heavy agents without declaring `isolation: worktree` will now fail CI. This is a **lint failure**, not a runtime breakage — fix is to add 1 line to the frontmatter. No code rewrite required.

### 3.2 No `bc_id` declaration justification

The lint promotion is not a breaking change of any of the following types: API surface change, CLI flag rename, exit-code semantics change, public schema rename. It is a *quality-gate enforcement* over a previously SHOULD-style rule. CI failures are the migration mechanism; the migration cost is `<5 minutes per affected PR` (1-line frontmatter addition).

### 3.3 Migration Path for External Forks

If an external project has forked one of the 5 v3r2 agents and customized it without the isolation flag, the next sync from upstream will surface an LR-05 error. Migration:

```bash
# In external fork
git pull --rebase upstream main
moai agent lint  # discovers LR-05 errors
# Fix: add `isolation: worktree` to forked agent frontmatter
moai agent lint  # exit 0
```

The `manager-docs` sync phase will note this in the SPEC-V3R2-ORC-004 release notes section.

---

## 4. Test Strategy

### 4.1 RED Phase (Test Authoring)

The TDD methodology requires RED tests before any implementation. Test additions:

#### 4.1.1 `internal/cli/agent_lint_test.go` extensions (~80 LOC)

```go
// AC-06 fixture: removing isolation from expert-backend triggers LR-05 with ORC_WORKTREE_MISSING
func TestLintLR05_OrcWorktreeMissingSentinel(t *testing.T) {
    tmpDir := t.TempDir()
    agentFile := filepath.Join(tmpDir, "expert-backend.md")
    err := os.WriteFile(agentFile, []byte(`---
name: expert-backend
tools: Read, Write, Edit
permissionMode: bypassPermissions
---

Body.
`), 0o644)
    require.NoError(t, err)

    violations, _, _ := lintAgentFile(agentFile, false)
    require.NotEmpty(t, violations)
    require.Equal(t, "LR-05", violations[0].Rule)
    require.Contains(t, violations[0].Message, "ORC_WORKTREE_MISSING")
    require.Equal(t, SeverityError, violations[0].Severity)
}

// AC-07 fixture: adding isolation to evaluator-active triggers LR-09 with ORC_WORKTREE_ON_READONLY
func TestLintLR09_OrcWorktreeOnReadonlySentinel(t *testing.T) {
    tmpDir := t.TempDir()
    agentFile := filepath.Join(tmpDir, "evaluator-active.md")
    err := os.WriteFile(agentFile, []byte(`---
name: evaluator-active
tools: Read, Grep, Glob
permissionMode: plan
isolation: worktree
---

Body.
`), 0o644)
    require.NoError(t, err)

    violations, _, _ := lintAgentFile(agentFile, false)
    require.NotEmpty(t, violations)
    require.Equal(t, "LR-09", violations[0].Rule)
    require.Contains(t, violations[0].Message, "ORC_WORKTREE_ON_READONLY")
}
```

#### 4.1.2 `internal/cli/workflow_lint_test.go` (new, ~120 LOC)

```go
// AC-09 fixture: workflow.yaml implementer.isolation=none triggers ORC_WORKTREE_REQUIRED
func TestWorkflowLint_OrcWorktreeRequiredOnImplementer(t *testing.T) {
    tmpDir := t.TempDir()
    workflowFile := filepath.Join(tmpDir, "workflow.yaml")
    err := os.WriteFile(workflowFile, []byte(`workflow:
  team:
    role_profiles:
      implementer:
        mode: acceptEdits
        isolation: none
      tester:
        mode: acceptEdits
        isolation: worktree
      designer:
        mode: acceptEdits
        isolation: worktree
`), 0o644)
    require.NoError(t, err)

    cfg, err := loadWorkflowYAML(workflowFile)
    require.NoError(t, err)
    violations := validateRoleProfiles(cfg)
    require.Len(t, violations, 1)
    require.Equal(t, SentinelWorktreeRequired, violations[0].Rule)
    require.Equal(t, "implementer", strings.Split(violations[0].Path, ".")[3])
}

// Symmetric tests for tester and designer
func TestWorkflowLint_OrcWorktreeRequiredOnTester(t *testing.T) { /* ... */ }
func TestWorkflowLint_OrcWorktreeRequiredOnDesigner(t *testing.T) { /* ... */ }

// Exit code 0 when correctly configured
func TestWorkflowLint_NoViolationsOnCorrectConfig(t *testing.T) { /* ... */ }
```

### 4.2 GREEN Phase

Implement deltas per §1.3. Each test from §4.1 must transition RED → GREEN.

### 4.3 REFACTOR Phase

Extract sentinel keys into shared `internal/cli/sentinels.go` (or const block in agent_lint.go). Re-run all tests to confirm GREEN preserved.

### 4.4 Integration Tests

Beyond unit tests:

- `make test ./internal/cli/...` — full lint package green.
- `moai agent lint` against post-edit `.claude/agents/moai/*.md` → exit 0.
- `moai workflow lint` against current `.moai/config/sections/workflow.yaml` → exit 0.
- Manual injection scenarios (AC-06, AC-07, AC-09) with `git stash` rollback.

### 4.5 Drift-Guard Coverage

Compare `tasks.md` planned files against actual modifications post-implementation. Expected ~10 modified files (4 agents + 1 rule + 1 lint.go + 1 lint_test.go + 1 workflow_lint.go + 1 workflow_lint_test.go + 1 CHANGELOG.md + 1 manager-cycle conditional + template mirrors). Drift > 30% triggers re-planning per `spec-workflow.md` § Drift Guard.

---

## 5. Pre-Implementation Validation

### 5.1 Constitution Cross-Check

Per `.claude/rules/moai/design/constitution.md` §3.1 brand context, ORC-004 does not modify brand-related files; OK.

Per `.claude/rules/moai/core/zone-registry.md` ID `CONST-V3R2-021`, the rule "Implementation teammates in team mode (role_profiles: implementer, tester, designer) MUST use isolation: worktree when spawned via Agent()" is classified Evolvable (canary_gate: false). The plan strengthens this rule; no constitutional violation.

### 5.2 Constraint Compliance

Per `spec.md` §7:

- ✓ FROZEN constitution preserved: `worktree-integration.md` line 134 (read-only MUST NOT) is FROZEN; this plan does not edit it. Only line 135 (write-heavy SHOULD → MUST) is edited; that line is in EVOLVABLE zone per CON-001 §3.
- ✓ Template-first: all source-tree edits flow through `internal/template/templates/` first.
- ✓ No absolute paths in worktree-isolated agent prompts: cross-reference only, no new rule text.
- ✓ Minimum Claude Code version 2.1.97: documented in spec.md §3 environment, no CI gate change.
- ✓ Lint runtime budget: new `moai workflow lint` is a separate command with its own runtime budget (workflow.yaml is ~5 KB, parse + validate <50 ms). `moai agent lint` runtime unchanged.
- ✓ No new isolation modes: enum stays at `none | worktree`.
- ✓ No bypass via agent body `--isolation=none`: agent body content is informational; spawn parameter overrides happen at orchestrator skill body level (out of agent scope).

### 5.3 Dependency Verification

- ✓ SPEC-V3R2-CON-001 merged (FROZEN/EVOLVABLE codified).
- ✓ SPEC-V3R2-ORC-002 merged (LR-05 introduced as warning).
- ◊ SPEC-V3R2-ORC-001 status TBD at run-phase start. T-ORC004-09 conditional on its merge.

### 5.4 Cross-SPEC Conflict Check

- SPEC-V3R2-ORC-005 (workflow.yaml schema) is in plan-phase. The new `moai workflow lint` reads role_profiles. If ORC-005 changes the role_profiles schema before ORC-004 ships, workflow_lint.go's struct definition must update. Mitigation: `manager-spec` for ORC-005 imports ORC-004's `workflowConfig` struct as a backward-compat reference.
- SPEC-V3R2-MIG-001 (migrator) — out of phase; no direct conflict.
- SPEC-V3R2-RT-003 (sandbox) — complementary, no conflict.

---

## 6. MX Tag Plan

Per `.claude/rules/moai/workflow/mx-tag-protocol.md`, write-heavy code requires MX annotations. Files receiving tags during run phase:

| File | Tags | Reason |
|------|------|--------|
| `.claude/rules/moai/workflow/worktree-integration.md` | `@MX:ANCHOR(WorktreeMUSTRule)` near line 135 | Rule is invariant contract for all v3r2 write-heavy agents |
| `internal/cli/workflow_lint.go` | `@MX:NOTE(WorkflowLintIntent)` at top of file | Intent-delivery for new CLI |
| `internal/cli/workflow_lint.go` | `@MX:ANCHOR(SentinelWorktreeRequired)` near sentinel const | High fan_in invariant |
| `internal/cli/agent_lint.go` | `@MX:NOTE(SentinelKeysAdded)` near sentinel const block (M5) | Intent for sentinel addition |
| `.claude/agents/moai/researcher.md` | `@MX:NOTE` body comment near body line update | P-A22 reconciliation context |

Total: ~5 MX annotations. Validated during M8 (T-ORC004-21).

---

## 7. plan-auditor Self-Check Mirror

Per the plan-auditor checklist (12 dimensions):

| # | Dimension | Self-Assessment |
|---|-----------|-----------------|
| 1 | EARS coverage | spec.md §5 has 15 REQs, all mapped in §1.4 |
| 2 | Traceability matrix | §1.4 maps every REQ to ≥1 AC and ≥1 Task |
| 3 | Interface signatures | §1.5 has 3 interfaces (CLI, sentinel const, struct) |
| 4 | Risk register | §8 below |
| 5 | Test plan | §4 RED/GREEN/REFACTOR + integration |
| 6 | Migration strategy | §3 with no `bc_id` justification |
| 7 | Pre-implementation validation | §5 constitution + constraint cross-check |
| 8 | MX tag plan | §6 with 5 annotation targets |
| 9 | Drift guard | §4.5 with 30% threshold |
| 10 | Acknowledged discrepancies | §1.2.1 |
| 11 | Build & validation cost estimate | research.md §8 |
| 12 | Cross-SPEC conflict check | §5.4 |

All 12 dimensions present. Risk areas front-loaded:

- **D1 (frozen-zone violation)**: §5.1 demonstrates line 135 is in EVOLVABLE zone.
- **D2 (template-first violation)**: M7 explicitly sequences make build before mirror.
- **D5 (REQ → AC drift)**: REQ-008 enforcement-layer narrowing acknowledged in §1.2.1.
- **D7 (manager-cycle dependency race)**: M2 explicitly conditional, blocker handling defined.
- **D9 (LR-05 false positive risk)**: research.md §4.1 documents 0 current matches.
- **D11 (sentinel key documentation completeness)**: M3 includes documentation section.
- **D12 (CLI integration with rootCmd)**: §1.5.1 specifies registration as subcommand group sibling.

---

## 8. Risks & Mitigations

| Risk | Impact | Likelihood | Mitigation | Reference |
|------|--------|-----------|-----------|-----------|
| ORC-001 not merged when ORC-004 run starts | HIGH | LOW | M2 conditional task, blocker report path defined | research.md §4.2 |
| Template-first discipline violation by implementer | MEDIUM | MEDIUM | M7 explicitly sequences; CI diff parity catch | research.md §4.6 |
| LR-05 substring match false positive on hypothetical agent | LOW | LOW | research.md §7 OQ3 documents no current matches | research.md §4.1 |
| `moai workflow lint` parsing fails on hand-edited workflow.yaml with non-standard formatting | MEDIUM | LOW | Use yaml v3 with strict mode; exit 2 with clear message | §1.5.3 |
| Sentinel key naming conflicts with other SPECs | LOW | LOW | All sentinels prefixed with `ORC_` per SPEC-V3R2 convention | §1.5.2 |
| Spawn-time enforcement narrowed to static gate (not runtime) | MEDIUM | accepted | Documented in §1.2.1; future SPEC for runtime hook | research.md §2.6 |
| Researcher body line text drift if upstream rewrites researcher.md | LOW | LOW | Use anchored grep on "when possible" exact phrase; if absent, log warning and proceed | research.md §4.5 |
| `make build` fails post-edit due to embedded.go regeneration error | MEDIUM | LOW | Run `make build` early in M7 to fail-fast; if fails, manual `internal/template/embedded.go` inspection | CLAUDE.local.md §2 |
| CHANGELOG entry missing on PR | LOW | LOW | T-ORC004-20 explicitly creates entry | TRUST 5 Trackable |
| `moai update` overwrites local-only files unexpectedly | LOW | accepted | CLAUDE.local.md §2 protected directory list verified pre-update | CLAUDE.local.md §2 |
| Agent Teams workflow regression from worktree overhead | MEDIUM | LOW | spec.md §8 risk row 1: only write-heavy agents flagged; single-file writers stay at none | spec.md §8 |
| Worktree merge-back conflicts on overlapping file writes | HIGH | LOW | spec.md §8 risk row 3: CC runtime handles; team-mode file ownership prevents pre-spawn | spec.md §8 |
| User-side `moai workflow lint` integration in pre-commit hook | LOW | accepted | Out of ORC-004 scope; document in sync-phase release notes | n/a |

---

## 9. Out of ORC-004 Scope (Future Work)

The following items are intentionally deferred (cross-reference spec.md §1.2 Non-Goals):

1. **Runtime spawn-time enforcement hook**: A `SubagentStart` hook that intercepts `Agent()` parameters and rejects mismatch in real-time. Requires Claude Code hook architecture extension. Future SPEC if hook telemetry shows real-world drift.
2. **Per-agent fan-out classifier upgrade**: LR-05 currently uses a hardcoded list. A future enhancement could parse agent description to count Write/Edit fan-out and auto-classify. Not needed at current agent count (5).
3. **Worktree creation/teardown mechanism changes**: Owned by Claude Code runtime; out of ORC scope.
4. **role_profiles schema expansion** (e.g., adding new roles like `documenter`): Owned by SPEC-V3R2-ORC-005.
5. **Cross-worktree communication patterns**: Not applicable at this layer.
6. **Performance benchmarks of worktree vs no-worktree**: Out of scope; spec.md §1.2.
7. **Pre-commit hook integration of `moai workflow lint`**: User-configurable; out of ORC-004 scope.
8. **`isolation: remote` mode**: Rejected in v2 and v3r2 per Master §13 Non-Goals.
9. **Recovery protocol for worktree creation failures**: CC handles at runtime; out of scope.

---

## 10. Summary

ORC-004 promotes worktree isolation from SHOULD to MUST for write-heavy v3r2 agents. The work is small (~10 modified files, ~150 LOC of new Go code) but high-leverage: it eliminates the silent file-write conflict failure mode that surfaces only at git merge time. The partial-pre-completion of the lint engine reduces the run scope to frontmatter additions, rule text upgrade, sentinel key documentation, and one new CLI command (`moai workflow lint`).

The static-CI-gate strategy for REQ-008 enforcement is acknowledged as narrower than the spec.md verbatim text but is the maximally practical approach given the absence of a Go-side spawn wrapper. Future runtime enforcement remains future work.

**Estimated implementation complexity**: MEDIUM. **Estimated PR size**: ~12-15 files, ~400 LOC added (mostly tests + new CLI), ~10 LOC modified.

---

End of plan.
