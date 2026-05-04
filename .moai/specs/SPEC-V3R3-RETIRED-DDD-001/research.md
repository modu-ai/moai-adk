# SPEC-V3R3-RETIRED-DDD-001 Research (Phase 0.5)

> Deep codebase analysis for manager-ddd retired stub standardization (follow-up to SPEC-V3R3-RETIRED-AGENT-001).
> Companion to `spec.md` v0.3.0 and `plan.md` v0.3.0.
> Solo mode, no worktree. Working directory: `/Users/goos/MoAI/moai-adk-go`. Branch: `feature/SPEC-V3R3-RETIRED-DDD-001`.
> main HEAD: `90b849669` (CHANGELOG backfill PR #777 merged).

## HISTORY

| Version | Date       | Author                        | Description                                                                                                                                              |
|---------|------------|-------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-04 | MoAI Plan Workflow (Phase 0.5) | 최초 작성 — 예측사 SPEC-V3R3-RETIRED-AGENT-001 패턴 reference + manager-ddd 현재 상태 + 19-file substitution scope evidence (iter 1 — 부정확 count, see 0.2.0 fix) |
| 0.2.0   | 2026-05-04 | MoAI Plan Workflow (iter 2)   | Audit defect fixes D2/D8: 정확한 파일 분류 (3 disposition categories: SUBSTITUTE-TO-CYCLE 30 / KEEP-AS-IS 3 / UPDATE-WITH-ANNOTATION 2; total 35 grep-detected files = 30+3+2). output-styles/moai/moai.md 추가 발견 (iter 1 누락). 모든 cross-artifact count는 §6 ground truth 기반으로 propagate. |
| 0.3.0   | 2026-05-04 | manager-spec, iter 3 atomic sync | Audit iter 2 defect D-NEW-1 fix: §5.2 New test 2 skeleton의 out-of-range REQ reference를 sequential REQ-RD-012로 정정 (renumbering map 적용). §8 Risk row legacy file-count text를 canonical Cat A count로 갱신. 5 artifacts version bumped to current. |

---

## 1. Research Objectives

본 research.md는 plan-auditor PASS criterion #5 ("research.md grounds decisions in actual codebase scan AND verifies external behavior")를 충족하기 위해 다음을 수행한다:

1. **predecessor pattern reference 정리** — SPEC-V3R3-RETIRED-AGENT-001 (commit `20d77d931`, PR #776 merged)에서 확립된 retirement frontmatter schema, sentinel format, audit test pattern 추출.
2. **manager-ddd 현재 상태 vs manager-tdd retired 상태 비교** — file size, frontmatter shape, body content delta.
3. **factory.go `case "ddd":` 처리 결정** — keep (backward compat) vs remove (clean cycle-only routing) trade-off.
4. **agent_frontmatter_audit_test.go 확장 범위** — 기존 manager-tdd 단언에 manager-ddd 추가 시 변경 surface.
5. **Documentation reference substitution scope (exhaustive grep)** — `manager-ddd` 34 files (iter 2 corrected; iter 1 missed `output-styles/moai/moai.md`); 3 disposition categories (Cat A SUBSTITUTE 30 / Cat B KEEP 3 / Cat C UPDATE-WITH-ANNOTATION 2 — total 35 with C2 having 0 manager-ddd occurrences); `ddd-*` hook actions 별도 분리 (§6.7).
6. **runtime guard 검증** — agentStartHandler가 모든 retired:true 에이전트에 generic하게 작동함을 confirm (REQ-RA-007 implementation 이미 manager-ddd 커버 가능).

---

## 2. Predecessor Pattern Reference (SPEC-V3R3-RETIRED-AGENT-001)

### 2.1 머지 상태 확인

```bash
$ git log --oneline -1 main
90b849669 (post-PR #777 main)
$ git log --oneline | grep "RETIRED-AGENT" | head -3
20d77d931 feat(agent-runtime): SPEC-V3R3-RETIRED-AGENT-001 — retired stub 호환성 + manager-cycle 템플릿 정합화
```

PR #776 머지 완료, 17 files changed, plan-auditor 0.88 + evaluator iter 2 PASS.

### 2.2 Established Retirement Frontmatter Schema (verbatim from manager-tdd.md)

[VERBATIM] `internal/template/templates/.claude/agents/moai/manager-tdd.md` lines 1–12 (post-merge):

```yaml
---
name: manager-tdd
description: |
  Retired (SPEC-V3R3-RETIRED-AGENT-001) — use manager-cycle with cycle_type=tdd.
  This agent has been consolidated into the unified manager-cycle agent.
  See manager-cycle.md for the active replacement.
retired: true
retired_replacement: manager-cycle
retired_param_hint: "cycle_type=tdd"
tools: []
skills: []
---
```

Body structure:
1. H1 title `# manager-tdd — Retired Agent`
2. `## Replacement` section — replacement agent name + invocation pattern
3. `## Migration Guide` table — Old Invocation | New Invocation 2 rows
4. `## Why This Change` paragraph — consolidation rationale
5. `## Active Agent` pointer — `.claude/agents/moai/manager-cycle.md`

Total file size: ~1392 bytes (vs current manager-ddd.md 7628 bytes → projected delta ~-6200 bytes after retirement).

### 2.3 Established 5 Required Retirement Fields

| Field | Type | Required | Manager-tdd Value | Manager-ddd Target Value |
|-------|------|----------|-------------------|--------------------------|
| `retired` | bool | yes | `true` | `true` |
| `retired_replacement` | string | yes | `manager-cycle` | `manager-cycle` |
| `retired_param_hint` | string | yes | `"cycle_type=tdd"` | `"cycle_type=ddd"` |
| `tools` | array | yes (explicit empty) | `[]` | `[]` |
| `skills` | array | yes (explicit empty) | `[]` | `[]` |

Legacy fields **rejected** (per audit test):
- `status: retired` (custom field) — must use `retired: true` boolean instead.

### 2.4 Established CI Sentinel Pattern

`internal/template/agent_frontmatter_audit_test.go` enforces (verbatim from current code):

| Sentinel String | Trigger | Failed Test | REQ |
|-----------------|---------|-------------|-----|
| `RETIREMENT_INCOMPLETE` (no agent suffix) | `status: retired` legacy field detected | `TestAgentFrontmatterAudit` | REQ-RA-002 |
| `RETIREMENT_INCOMPLETE_<agent-name>` | retired:true 에이전트에 5 fields 중 하나 누락 | `TestAgentFrontmatterAudit/<path>` | REQ-RA-002 |
| `RETIREMENT_INCOMPLETE_manager-tdd` | 명시적 단언: manager-tdd MUST be retired | `TestAgentFrontmatterAudit/manager-tdd must be retired` | REQ-RA-002 |
| `RETIREMENT_INCOMPLETE_manager-tdd` (replacement) | manager-cycle.md 부재 | `TestRetirementCompletenessAssertion/manager-tdd replacement manager-cycle must exist` | REQ-RA-016 |
| `RETIREMENT_INCOMPLETE_<agent>` (generic loop) | retired_replacement 파일 부재 | `TestRetirementCompletenessAssertion/all retired agents have replacement` | REQ-RA-016 |
| `ORPHANED_MANAGER_TDD_REFERENCE` | 명시 6 파일 중 manager-tdd 참조 검출 | `TestNoOrphanedManagerTDDReference/<path>` | REQ-RA-013 |

[CRITICAL EVIDENCE] `findManagerTDDReferences()` (lines 315–338) implements name-bound substring search excluding:
- frontmatter `name: manager-tdd` line (manager-tdd.md 자체의 self-reference)
- `# deprecated` 마크다운 헤더
- HTML 코멘트 `<!--` 시작 라인

Same exclusion logic must apply to manager-ddd substitution check (M3).

### 2.5 Established MX Tag Pattern (predecessor M5)

`internal/hook/subagent_start.go` (post-merge) line 165–168:

```go
// @MX:NOTE: [AUTO] agentStartHandler는 SPEC-V3R3-RETIRED-AGENT-001 5-layer defect chain
// Layer 1 (retired stub frontmatter invalid) 차단의 핵심 진입점이다.
// retired:true frontmatter detect 시 block decision JSON + exit code 2 반환으로
// Claude Code Agent runtime의 worktree allocation 전 spawn을 거부한다 (≤500ms 응답 budget).
```

Pattern: `@MX:NOTE` 4 줄 (1 헤더 + 3 본문). 본 SPEC factory.go `case "ddd":` 보존 시 동일 형식 사용.

`internal/cli/worktree_validation.go` `validateWorktreeReturn` 도 `@MX:ANCHOR` + `@MX:REASON` 사용. Reference for any new helper.

### 2.6 Established Performance Budget

REQ-RA-012: `agentStartHandler.Handle ≤500ms`. 측정 결과: **0.056ms × 9000 invocations** (predecessor evaluator iter 2 PASS, target 여유 8927×). 본 SPEC은 동일 핸들러 재사용 — 추가 perf burden 없음.

---

## 3. manager-ddd Current State Analysis

### 3.1 File Size Delta

| Location | Size | Date | Status |
|----------|------|------|--------|
| `internal/template/templates/.claude/agents/moai/manager-ddd.md` | 7628 bytes (163 lines) | 2026-04-30 12:17 | full active definition |
| `mo.ai.kr/.claude/agents/moai/manager-ddd.md` | 1000 bytes | 2026-05-01 13:51 | retired stub (identical pattern to manager-tdd) |

### 3.2 Current Frontmatter (manager-ddd.md lines 1–39)

```yaml
---
name: manager-ddd
description: |
  DDD (Domain-Driven Development) implementation specialist. Use for ANALYZE-PRESERVE-IMPROVE
  cycle when working with existing codebases that have minimal test coverage.
  MUST INVOKE when ANY of these keywords appear in user request:
  --deepthink flag: Activate Sequential Thinking MCP for deep analysis...
  EN: DDD, refactoring, legacy code, behavior preservation, characterization test, domain-driven refactoring
  KO: DDD, 리팩토링, 레거시코드, 동작보존, 특성테스트, 도메인주도리팩토링
  JA: DDD, リファクタリング, レガシーコード, 動作保存, 特性テスト, ドメイン駆動リファクタリング
  ZH: DDD, 重构, 遗留代码, 行为保存, 特性测试, 领域驱动重构
  NOT for: greenfield development (use TDD), deployment, documentation, git operations, security audits
tools: Read, Write, Edit, MultiEdit, Bash, Grep, Glob, TodoWrite, Skill, mcp__sequential-thinking__sequentialthinking, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
model: sonnet
permissionMode: bypassPermissions
memory: project
skills:
  - moai-foundation-core
  - moai-workflow-ddd
  - moai-workflow-testing
hooks:
  PreToolUse:
    - matcher: "Write|Edit|MultiEdit"
      hooks:
        - type: command
          command: "...ddd-pre-transformation"
          timeout: 5
  PostToolUse:
    - matcher: "...ddd-post-transformation"
  SubagentStop:
    - hooks: [...ddd-completion]
---
```

Delta from retirement target:
- `description`: full DDD spec → terse retirement notice (2-3 lines)
- `tools`: full CSV → `[]`
- `model`: `sonnet` → REMOVED (not needed for retired stub)
- `permissionMode`: `bypassPermissions` → REMOVED
- `memory`: `project` → REMOVED
- `skills`: 3 entries → `[]`
- `hooks`: 3 events → REMOVED entirely (retired stub does no work)
- ADDED: `retired: true`, `retired_replacement: manager-cycle`, `retired_param_hint: "cycle_type=ddd"`

### 3.3 Current Body Sections (manager-ddd.md lines 41–163)

H1 + 11 H2 sections:
- `## Primary Mission`
- `## Behavioral Contract (SEMAP)` (Pre/Post/Invariants/Forbidden)
- `## Scope Boundaries` (IN/OUT)
- `## Delegation Protocol`
- `## Execution Workflow` (STEP 1–5 + STEP 1.5 + STEP 3.5 + STEP 4 sub-steps)
- `## Ralph-Style LSP Integration`
- `## Checkpoint and Resume`
- `## @MX Tag Obligations`
- `## DDD vs TDD Decision Guide`
- `## Common Refactoring Patterns`

After retirement: H1 + 4-5 H2 (per manager-tdd retired stub pattern):
- `# manager-ddd — Retired Agent`
- `## Replacement` (manager-cycle with cycle_type=ddd)
- `## Migration Guide` (table: 2 rows)
- `## Why This Change` (consolidation rationale, citing SPEC-V3R3-RETIRED-AGENT-001 + present SPEC-V3R3-RETIRED-DDD-001)
- `## Active Agent` (pointer to manager-cycle.md)

Estimated body delta: -123 lines / -5800 bytes.

### 3.4 Inline References to manager-tdd in manager-ddd.md

`grep -n "manager-tdd\|manager-cycle" manager-ddd.md`:
- Line 47: `For projects with sufficient coverage, use manager-cycle with cycle_type=tdd.` (post-predecessor M5 substitution; manager-tdd → manager-cycle 이미 완료)
- Line 63: `OUT OF SCOPE: ... use manager-cycle with cycle_type=tdd ...` (post-predecessor M5)

**No outstanding manager-tdd references remain in manager-ddd.md** (predecessor M5 cleanup completed). The body retirement (M2 본 SPEC) replaces all this content with retirement stub regardless.

---

## 4. factory.go `case "ddd":` Decision Analysis

### 4.1 Current State (post-PR #776)

`internal/hook/agents/factory.go` line 38–43 (verbatim):

```go
switch agent {
case "ddd":
    return NewDDDHandler(act), nil
case "tdd":
    return NewTDDHandler(act), nil
case "cycle":
    return NewCycleHandler(act), nil
```

Predecessor decision (line 26–28 comment): "manager-tdd retired stub uses no hooks (frontmatter cleared) but `case "tdd":` is preserved for backward compatibility with legacy user projects that have not run `moai update`."

**Observed**: predecessor kept `case "tdd":` 보존 + added @MX:NOTE rationale at the switch statement level (line 19).

### 4.2 Trade-off Analysis (mirror predecessor decision)

**Option A: Keep `case "ddd":` with @MX:NOTE (recommended, mirrors predecessor)**

Pros:
- Backward compatibility: legacy user projects (incl. mo.ai.kr 사이드 프로젝트) that have not yet run `moai update` may still have full active manager-ddd.md with `ddd-pre-transformation` / `ddd-post-transformation` / `ddd-completion` hook references. Removing the case routes their hook events to `default_handler.go` (no-op pass-through) — silent loss of intended behavior.
- Symmetry with `case "tdd":` decision (predecessor commit `20d77d931`). Two retirements use identical pattern.
- Cost: +0 LOC (just preserve existing case). @MX:NOTE update at switch-level comment only (1-2 line edit).

Cons:
- "Dead" case in the long term (after all users update): code clutter.

**Option B: Remove `case "ddd":` (clean cycle-only routing)**

Pros:
- Clean codebase: factory only routes active agents.

Cons:
- Backward compat break: legacy user projects (esp. mo.ai.kr) still firing `ddd-*` hook events get default_handler instead of DDDHandler — silent feature regression for users mid-migration.
- Asymmetric with `case "tdd":` decision. Mixing kept-tdd + removed-ddd creates inconsistency in the codebase.
- Removes existing `internal/hook/agents/ddd_handler.go` + `factory_test.go` test rows → 5 file change instead of 1.

**Decision (research.md outcome)**: Option A (Keep `case "ddd":` with updated @MX:NOTE). Rationale: mirror predecessor decision; backward compat is the conservative choice; mo.ai.kr is the exact use case the predecessor named. Alternative (Option B) deferred to a separate cleanup SPEC if/when telemetry confirms zero `case "ddd":` invocations across all installed projects (≥6 month observation window).

### 4.3 @MX:NOTE Update Surface

Predecessor switch-level @MX:NOTE (line 19–28, manager-tdd retirement rationale):

```go
// @MX:NOTE: [AUTO] Switch branch that creates one of 11 handler types based on the agent name. Add a new case here when adding a new agent.
// Supported agents: ddd, tdd (legacy retired stub compat), cycle, backend, frontend, testing, debug, devops, quality, spec, docs
// CreateHandler creates a handler for the given agent action.
// Action format: {agent}-{action}
// Examples: cycle-pre-implementation, backend-validation, docs-completion
//
// SPEC-V3R3-RETIRED-AGENT-001: cycle handler dispatches manager-cycle's unified
// DDD/TDD workflow hooks (REQ-RA-009). manager-tdd retired stub uses no hooks
// (frontmatter cleared) but `case "tdd":` is preserved for backward
// compatibility with legacy user projects that have not run `moai update`.
```

M5 update target: append a sentence noting `case "ddd":` 도 retired stub backward compat이 됨 (per SPEC-V3R3-RETIRED-DDD-001).

Updated comment skeleton (M5 implementation):

```go
// @MX:NOTE: [AUTO] ... (existing 5 lines preserved)
//
// SPEC-V3R3-RETIRED-AGENT-001 + SPEC-V3R3-RETIRED-DDD-001: cycle handler dispatches
// manager-cycle's unified DDD/TDD workflow hooks. manager-tdd + manager-ddd retired
// stubs use no hooks (frontmatter cleared) but `case "tdd":` and `case "ddd":` are
// preserved for backward compatibility with legacy user projects that have not run
// `moai update`.
```

LOC delta at factory.go: ~+3 lines (NOTE expansion only). No code change.

---

## 5. agent_frontmatter_audit_test.go Extension Surface

### 5.1 Current Test Functions (post-PR #776)

3 top-level test functions:

1. `TestAgentFrontmatterAudit` (lines 63–162):
   - Generic: walks all `.claude/agents/moai/*.md`, validates retirement fields for every retired:true agent.
   - **Already covers manager-ddd automatically** once frontmatter has `retired: true` (M2 후).
   - Has explicit subtest `"manager-tdd must be retired"` (lines 139–161) — RED trigger before predecessor M2.
   - **Extension needed**: add explicit subtest `"manager-ddd must be retired"` (mirror lines 139–161, substitute manager-tdd → manager-ddd).

2. `TestRetirementCompletenessAssertion` (lines 169–228):
   - Generic loop: every retired:true agent's `retired_replacement` file must exist.
   - **Already covers manager-ddd automatically** once frontmatter is set.
   - Has explicit subtest `"manager-tdd replacement manager-cycle must exist"` (lines 179–188) — generic check.
   - **Extension needed**: add explicit subtest `"manager-ddd replacement manager-cycle must exist"` for symmetric protection.

3. `TestNoOrphanedManagerTDDReference` (lines 235–298):
   - Specific: 6 files validated for absence of `manager-tdd` substring.
   - **Does NOT cover manager-ddd**.
   - **Extension needed**: add new test `TestNoOrphanedManagerDDDReference` (sibling, not modification of TDD test) — dedicated function for clean diff + targeted error reporting.

### 5.2 Extension Pattern (M3 implementation)

**New test 1**: `TestAgentFrontmatterAudit/manager-ddd must be retired` subtest (insert after line 161):

```go
t.Run("manager-ddd must be retired", func(t *testing.T) {
    t.Parallel()
    const dddPath = ".claude/agents/moai/manager-ddd.md"
    data, readErr := fs.ReadFile(fsys, dddPath)
    if readErr != nil {
        t.Fatalf("manager-ddd.md 읽기 실패 (파일 존재해야 함): %v", readErr)
    }
    fm, _, parseErr := parseFrontmatterAndBody(string(data))
    if parseErr != "" {
        t.Fatalf("manager-ddd.md frontmatter 파싱 실패: %s", parseErr)
    }
    rf := parseRetiredFields(fm)
    // SPEC-V3R3-RETIRED-DDD-001 REQ-RD-002: manager-ddd는 retired:true여야 함
    if !rf.retired {
        t.Errorf("RETIREMENT_INCOMPLETE_manager-ddd: manager-ddd.md에 'retired: true' 없음. " +
            "SPEC-V3R3-RETIRED-DDD-001 M2에서 retired stub으로 교체 필요 (REQ-RD-002)")
    }
})
```

**New test 2**: Insert in `TestRetirementCompletenessAssertion` (after line 188):

```go
t.Run("manager-ddd replacement manager-cycle must exist", func(t *testing.T) {
    t.Parallel()
    const replacementPath = ".claude/agents/moai/manager-cycle.md"
    _, statErr := fs.Stat(fsys, replacementPath)
    if statErr != nil {
        t.Errorf("RETIREMENT_INCOMPLETE_manager-ddd: 교체 에이전트 '%s'가 embedded FS에 없음. "+
            "SPEC-V3R3-RETIRED-DDD-001 M2에서 manager-cycle.md 검증 필요 (REQ-RD-012)", replacementPath)
    }
})
```

**New test 3**: New top-level `TestNoOrphanedManagerDDDReference` (sibling of TDD test):

Identical structure to `TestNoOrphanedManagerTDDReference` but:
- Helper renamed: `findManagerDDDReferences(content string) []string`
- checkFiles 슬라이스: **30 files** (Cat A SUBSTITUTE-TO-CYCLE files per §6.2 below; explicitly Cat B + Cat C are EXCLUDED from this slice). Each file path enumerated explicitly in the test file (no glob).
- Sentinel: `ORPHANED_MANAGER_DDD_REFERENCE` (mirror naming)

Allow-list exclusions in helper `findManagerDDDReferences` (mirror predecessor TDD test):
- `# deprecated` 마크다운 헤더 (defensive; not expected in Cat A files)
- HTML 코멘트 `<!--` 시작 라인 (general defensive allow)
- (no `ddd-pre-transformation` / `ddd-post-transformation` / `ddd-completion` allow needed: these are Cat C C2 file `handle-agent-hook.sh.tmpl` which is NOT in the 30-file checkFiles slice)
- (no `name: manager-ddd` allow needed: manager-ddd.md is Cat B1 and NOT in the 30-file checkFiles slice)

**Decision**: NOT to extend `TestNoOrphanedManagerTDDReference` to also check manager-ddd. Reasons:
- Clean diff per SPEC (each retirement gets its own test function).
- Targeted error messages (sentinel pattern is per-agent).
- Future retirements follow the same dedicated-test convention.

### 5.3 Test File LOC Delta

Estimated:
- 2 subtests added (~14 lines each = 28 lines)
- 1 new top-level test function (~70 lines, mirror TDD test)
- 1 helper `findManagerDDDReferences()` (~20 lines)

Total: ~+118 lines to `agent_frontmatter_audit_test.go` (current 339 lines → ~457 lines).

---

## 6. Documentation Reference Substitution Scope (Exhaustive Grep — Single Source of Truth)

[AUTHORITATIVE GROUND TRUTH] `grep -rln "manager-ddd" internal/template/templates/` returned **34 files** (corrected from iter 1 incorrect "19" claim — `output-styles/moai/moai.md` was missed). All cross-artifact counts MUST derive from this section.

### 6.1 Three Disposition Categories (Mutually Exclusive)

Each of the 34 grep-detected files falls into exactly ONE of three disposition categories:

| Category | Definition | Post-SPEC State |
|----------|------------|------------------|
| **Cat A — SUBSTITUTE-TO-CYCLE** | `manager-ddd` substring is replaced with `manager-cycle` (or `manager-cycle with cycle_type=ddd`); file is modified | 0 occurrences of `manager-ddd` |
| **Cat B — KEEP-AS-IS** | File contains intentional `manager-ddd` references (retirement notice / self-reference); no edit applied | Unchanged occurrences |
| **Cat C — UPDATE-WITH-ANNOTATION** | File is modified (add @MX:NOTE) but original `manager-ddd` substring is preserved as legacy backward-compat documentation | Original occurrences preserved + annotation added |

The third category (Cat C) is the resolution for D8 (audit iter 1): `agent-hooks.md` and `handle-agent-hook.sh.tmpl` were ambiguously described as "KEEP" in iter 1 but are actually modified to add @MX:NOTE comments. Cat C captures this hybrid state.

### 6.2 Cat A — SUBSTITUTE-TO-CYCLE (30 files, ~58 substitutions)

These 30 files MUST contain 0 occurrences of `manager-ddd` after this SPEC's M3 substitution. They form the authoritative `TestNoOrphanedManagerDDDReference.checkFiles` slice.

**Cat A sub-section A1 — Rule files (3 files, 3 occurrences)**:

| # | File | Occurrences | Action |
|---|------|-------------|--------|
| A1.1 | `rules/moai/development/agent-authoring.md` | 1 (line 104 list entry: `- manager-ddd: DDD implementation cycle`) | REMOVE list entry (manager-cycle already listed elsewhere) |
| A1.2 | `rules/moai/workflow/spec-workflow.md` | 1 (line 219 team mode table) | substitute → `manager-cycle (sequential, cycle_type per quality.yaml)` |
| A1.3 | `rules/moai/workflow/worktree-integration.md` | 1 (line 135 isolation rule list) | substitute → `expert-backend, expert-frontend, manager-cycle` |

**Cat A sub-section A2 — Agent definition files (11 files, ~22 occurrences)**:

| # | File | Occurrences | Action |
|---|------|-------------|--------|
| A2.1 | `agents/moai/manager-strategy.md` | 1 (line 136) | substitute |
| A2.2 | `agents/moai/manager-quality.md` | 3 (lines 64, 107, 113) | substitute (3) |
| A2.3 | `agents/moai/manager-spec.md` | 1 (line 58) | substitute |
| A2.4 | `agents/moai/expert-backend.md` | 2 (lines 62, 119) | substitute (2) |
| A2.5 | `agents/moai/expert-frontend.md` | 1 (line 119) | substitute |
| A2.6 | `agents/moai/expert-testing.md` | 3 (lines 55, 59, 98) | substitute (3) |
| A2.7 | `agents/moai/expert-debug.md` | 2 (lines 59, 90) | substitute (2) |
| A2.8 | `agents/moai/expert-devops.md` | 2 (lines 59, 114) | substitute (2) |
| A2.9 | `agents/moai/expert-mobile.md` | 1 (line 105) | substitute |
| A2.10 | `agents/moai/expert-refactoring.md` | 1 (line 54) | substitute |
| A2.11 | `agents/moai/evaluator-active.md` | 1 (line 94) | substitute |

**Cat A sub-section A3 — Output-style files (1 file, 1 occurrence)**:

| # | File | Occurrences | Action |
|---|------|-------------|--------|
| A3.1 | `output-styles/moai/moai.md` | 1 (line 127, table cell `manager-ddd / manager-tdd`) | substitute → `manager-cycle` |

> **Note**: A3.1 was discovered during iter 2 grep audit; iter 1 research missed this file. Total Cat A count revised from 27 (iter 1) to 30 (iter 2).

**Cat A sub-section A4 — Skill files (15 files, ~32 occurrences)**:

| # | File | Occurrences | Action |
|---|------|-------------|--------|
| A4.1 | `skills/moai/SKILL.md` | 2 (lines 117, 219) | substitute (2) |
| A4.2 | `skills/moai/references/mx-tag.md` | 2 | substitute (2) |
| A4.3 | `skills/moai/references/reference.md` | 1 | substitute |
| A4.4 | `skills/moai/workflows/moai.md` | 3 (lines 78, 148, 239) | substitute (3) |
| A4.5 | `skills/moai/workflows/run.md` | 7 (lines 5, 24, 540, 599, 603, 619, 983) | substitute (7) |
| A4.6 | `skills/moai-foundation-cc/SKILL.md` | 1 | substitute |
| A4.7 | `skills/moai-foundation-core/SKILL.md` | 1 (line 32) | substitute |
| A4.8 | `skills/moai-foundation-quality/SKILL.md` | 1 (line 31) | substitute |
| A4.9 | `skills/moai-meta-harness/SKILL.md` | 2 (lines 155, 215) | substitute (2) |
| A4.10 | `skills/moai-workflow-ddd/SKILL.md` | 1 (line 20 `agent: "manager-ddd"`) | substitute → `agent: "manager-cycle"` |
| A4.11 | `skills/moai-workflow-loop/SKILL.md` | 1 (line 153) | substitute |
| A4.12 | `skills/moai-workflow-spec/SKILL.md` | 2 (lines 223, 356) | substitute (2) |
| A4.13 | `skills/moai-workflow-spec/references/reference.md` | 4 (lines 19, 70, 272, 533) | substitute (4) |
| A4.14 | `skills/moai-workflow-spec/references/examples.md` | 2 (lines 17, 334) | substitute (2) |
| A4.15 | `skills/moai-workflow-testing/SKILL.md` | 1 (line 20 `agent: "manager-ddd"`) | substitute → `agent: "manager-cycle"` |

**Cat A totals**: A1 (3) + A2 (11) + A3 (1) + A4 (15) = **30 files**.

### 6.3 Cat B — KEEP-AS-IS (3 files)

These 3 files contain `manager-ddd` references that are intentional retirement-related cross-references; the SPEC explicitly does NOT modify them (with one exception for `manager-ddd.md` whose body is rewritten in M2 but whose frontmatter `name: manager-ddd` field is preserved).

| # | File | Occurrences | Reason for KEEP |
|---|------|-------------|------------------|
| B1 | `agents/moai/manager-ddd.md` | 1 (line 2 frontmatter `name:`) | Self-reference; M2 rewrites body but preserves frontmatter `name: manager-ddd` (retired-agent metadata key) |
| B2 | `agents/moai/manager-cycle.md` | 3 (lines 61, 65, 70 migration table) | Active replacement agent intentionally lists retired agents in Migration Notes |
| B3 | `agents/moai/manager-tdd.md` | 1 (body line 31 consolidation cross-reference) | Predecessor retired stub body cites both retired agents in `## Why This Change` |

> **Special case for B1**: `manager-ddd.md` is in Cat B for the `name:` field but Cat A would also apply if the body content were not rewritten. Since M2 rewrites the entire body, the post-SPEC file contains exactly 1 occurrence of `manager-ddd` (the frontmatter `name:` field). The `findManagerDDDReferences` allow-list MUST include this `name: manager-ddd` line.

### 6.4 Cat C — UPDATE-WITH-ANNOTATION (2 files)

These 2 files are modified by this SPEC (M4) but the original `manager-ddd` substring (where present) is preserved as legacy backward-compatibility documentation. An @MX:NOTE comment is added explaining the retirement decision.

| # | File | Occurrences (pre/post) | Action |
|---|------|-----------------------|--------|
| C1 | `rules/moai/core/agent-hooks.md` | 2 / 2 (lines 48, 79 preserved) | Add @MX:NOTE after Agent Hook Actions table noting "manager-ddd retired (SPEC-V3R3-RETIRED-DDD-001), action set preserved for backward compat" |
| C2 | `hooks/moai/handle-agent-hook.sh.tmpl` | 0 / 0 (no `manager-ddd` substring; has `ddd-*` action references in lines 5, 9) | Add @MX:NOTE bash comment after lines 5, 9 noting manager-ddd retirement and ddd-* backward compat |

> **C2 nuance**: This file does NOT contain the `manager-ddd` substring (grep confirmed 0 occurrences). It contains `ddd-pre-transformation` etc. (separate concern: hook action names). It is included in Cat C because the M4 modification is annotation-only and conceptually parallel to C1.

### 6.5 Total File Disposition (Authoritative Counts)

| Category | Files | Sum |
|----------|-------|-----|
| **Cat A — SUBSTITUTE-TO-CYCLE** | A1 (3) + A2 (11) + A3 (1) + A4 (15) | **30** |
| **Cat B — KEEP-AS-IS** | B1 + B2 + B3 | **3** |
| **Cat C — UPDATE-WITH-ANNOTATION** | C1 + C2 | **2** |
| **Total files modified by this SPEC** | Cat A (30) + Cat B (1, only B1 manager-ddd.md is rewritten) + Cat C (2) | **33** |
| **Total grep-detected files containing `manager-ddd` substring** | Cat A (30 contain pre-SPEC) + Cat B (3 contain) + Cat C-with-substring (1 = C1 agent-hooks.md, since C2 has 0 occurrences) | **34** |

**The single authoritative number for `TestNoOrphanedManagerDDDReference.checkFiles` is 30** (= Cat A only, since these are the files that MUST contain 0 occurrences after substitution). Cat B and Cat C-C1 are excluded from this slice via the helper allow-list.

### 6.6 `findManagerDDDReferences` Allow-List (mirrors helper exclusions)

The helper function in `agent_frontmatter_audit_test.go` MUST exclude these patterns from orphan-reference detection:

| Pattern | Reason | Files Affected |
|---------|--------|----------------|
| `name: manager-ddd` (frontmatter line) | Cat B1 self-reference in retired stub | manager-ddd.md (but this file is NOT in Cat A checkFiles slice anyway) |
| `# deprecated` (Markdown headers) | Predecessor TDD test allow-list pattern | (any file using deprecation headers) |
| `<!--` (HTML comment open) | Editor metadata in Markdown | (general allow) |
| (none required for Cat A files) | Cat A files contain no allow-listed patterns; substitution should be exhaustive | — |

**Rationale**: Since Cat B (3 files) and Cat C-C1 (`agent-hooks.md`) are EXCLUDED from `checkFiles` slice entirely, the allow-list need only cover patterns a Cat A file might legitimately contain (none expected). However, the allow-list is retained as defensive symmetry with predecessor `findManagerTDDReferences`.

### 6.7 `ddd-*` Hook Action References (separate concern from `manager-ddd`)

Search: `grep -rn "ddd-pre-transformation\|ddd-post-transformation\|ddd-completion"` returns:

| File | Lines | Disposition |
|------|-------|-------------|
| `agents/moai/manager-ddd.md` | 26, 32, 37 | REMOVE (retirement removes `hooks:` block entirely) |
| `hooks/moai/handle-agent-hook.sh.tmpl` | 5, 9 | KEEP (legacy doc; backward compat for case "ddd" routing) — possible @MX:NOTE update |
| `rules/moai/core/agent-hooks.md` | 48 | KEEP (table row legacy doc; backward compat) |

---

## 7. Files NOT in Substitution Scope (Outstanding Concerns)

### 7.1 mo.ai.kr Side Project

`/Users/goos/MoAI/mo.ai.kr/` is a sibling project that consumes moai-adk-go templates via `moai update`. The fact that mo.ai.kr already has `manager-ddd.md` as a 1000-byte retired stub (per spec-v3r3-retired-agent-001 §1.2 evidence) means:

- mo.ai.kr is the proximate motivation for this SPEC (consistency with already-deployed retirement).
- After this SPEC merges, mo.ai.kr will run `moai update` — moai-adk-go template's standardized retired stub will overwrite mo.ai.kr's existing 1000-byte retired stub, which is the desired sync direction.
- **CHANGELOG entry MUST call out `moai update`** so mo.ai.kr maintainers (and any other downstream users) explicitly run it.
- **mo.ai.kr is NOT modified by this SPEC** (out of scope per spec.md §2.2).

### 7.2 manager-strategy.md (line 136 single substitution)

The agent body at line 136 currently reads:
```
- On approval: hand TAG chain, library versions, key decisions, task list to manager-ddd/tdd
```

After substitution: `... to manager-cycle`. No structural change.

### 7.3 docs-site/ (4-locale ko/en/zh/ja docs)

`grep -rln "manager-ddd" docs-site/` (out-of-scope; would require CLAUDE.local.md §17 4-locale sync). **Deferred to follow-up SPEC** when docs-site is updated for v2.20.0+ release. CHANGELOG note: "docs-site reference sync deferred to v3.0 release window."

### 7.4 internal/hook/agents/ddd_handler.go (current implementation, no template)

This is Go source code under `internal/hook/agents/`, NOT a template file. Per factory.go decision (§4.2 Option A: keep), **ddd_handler.go is preserved as-is** — backward compat dictates. No modification needed.

`internal/hook/agents/factory_test.go` lines 200, 229, 231 reference `NewDDDHandler` — preserved (test rows for the kept case).

---

## 8. Risk Assessment Summary (file-anchored)

| Risk | Probability | Impact | File Anchor | Mitigation |
|------|-------------|--------|-------------|------------|
| factory.go `case "ddd":` removal breaks legacy users | L | H | `internal/hook/agents/factory.go` line 39 | Decision: keep with @MX:NOTE expansion (§4.2 Option A) |
| 30 Cat A file substitution scope creates merge conflicts | M | M | 30 Cat A files per §6.2 | Bulk substitution via Edit tool with `replace_all`; verify with grep post-edit |
| manager-ddd.md body retirement strips needed docs | M | L | `agents/moai/manager-ddd.md` lines 41–163 | Mirror manager-tdd retired stub structure exactly (5 H2 sections); SPEC-V3R3-RETIRED-DDD-001 명시 in `## Why This Change` |
| agent_frontmatter_audit_test.go extension breaks predecessor tests | L | M | `internal/template/agent_frontmatter_audit_test.go` lines 139, 179, 235 | Extend by adding NEW subtests + NEW top-level function; do not modify TDD tests |
| handle-agent-hook.sh.tmpl `ddd-*` references confuse new readers | L | L | `hooks/moai/handle-agent-hook.sh.tmpl` lines 5, 9 | Add `@MX:NOTE: SPEC-V3R3-RETIRED-DDD-001` comment noting backward compat reservation |
| moai-workflow-ddd/SKILL.md `agent: "manager-ddd"` is structural | M | L | `skills/moai-workflow-ddd/SKILL.md` line 20 | Substitute to `agent: "manager-cycle"` (skill metadata, no behavioral change) |
| mo.ai.kr maintainers miss `moai update` directive | M | M | CHANGELOG.md (post-merge) | Call out CHANGELOG entry "Run `moai update` to sync new templates" with explicit user action note |
| Substitution scope discovers false positives (manager-ddd in code comments, not template) | L | L | grep results | Manual review at M3; confirm via diff before commit |

---

## 9. Summary of Critical Decisions

| # | Decision | Outcome | Rationale |
|---|----------|---------|-----------|
| D1 | factory.go `case "ddd":` keep vs remove | KEEP with @MX:NOTE | §4.2 Option A; mirror predecessor manager-tdd decision; backward compat for mid-migration users |
| D2 | agent_frontmatter_audit_test.go extension scope | NEW subtests + NEW top-level function (not modify TDD) | §5.2; clean diff per retirement; targeted sentinels |
| D3 | manager-cycle.md migration table (lines 61, 65, 70) | KEEP "Deprecated agents" entries naming manager-ddd + manager-tdd | §6.1 row "manager-cycle.md" KEEP; intentional retirement notice |
| D4 | handle-agent-hook.sh.tmpl `ddd-*` action references | KEEP (legacy doc) + add @MX:NOTE backward compat reservation | §6.3; symmetric with `tdd-*` references |
| D5 | docs-site 4-locale sync | DEFER to follow-up SPEC | §7.3; CLAUDE.local.md §17 scope; release window v3.0 |
| D6 | mo.ai.kr direct modification | OUT OF SCOPE | §7.1 + spec.md §2.2; `moai update` is the canonical sync mechanism |
| D7 | moai-workflow-ddd/SKILL.md `agent:` field | SUBSTITUTE to `manager-cycle` | §6.1; skill metadata, no functional impact |
| D8 | manager-ddd.md body content for retired stub | EXACT MIRROR of manager-tdd retired stub structure | §3.3; consistency reduces maintenance burden |
| D9 | New REQ namespace prefix | Use `REQ-RD-NNN` (RD = Retired DDD); **sequential numbering 001..012 with NO gaps** (iter 2 fix per audit D1) | Distinguishes from predecessor's `REQ-RA-NNN`; aids traceability; "even one gap = FAIL" per agent identity rule M5 |
| D10 | Implementation methodology | TDD per quality.yaml development_mode | M1 RED tests → M2-M4 GREEN → M5 REFACTOR |
| D11 | output-styles/moai/moai.md discovery | ADD to Cat A A3.1 (iter 2 grep audit) | iter 1 research missed this file; total Cat A revised 27→30 files |
| D12 | 3-category disposition taxonomy (Cat A SUBSTITUTE / Cat B KEEP / Cat C UPDATE-WITH-ANNOTATION) | Replaces iter 1's 2-category KEEP-vs-SUBSTITUTE binary | §6.1; resolves audit D8 — `agent-hooks.md` is Cat C, not pure KEEP |

---

## 10. External Behavior Verification

### 10.1 Verify Predecessor PR #776 Merge

```bash
$ git log --all --oneline | grep "RETIRED-AGENT-001" | head -3
20d77d931 feat(agent-runtime): SPEC-V3R3-RETIRED-AGENT-001 — retired stub 호환성 + manager-cycle 템플릿 정합화
```

✅ Confirmed merged.

### 10.2 Verify manager-cycle.md Active Replacement

```bash
$ ls -la internal/template/templates/.claude/agents/moai/manager-cycle.md
-rw-r--r--  1 goos  staff  11385 May  4 18:20
$ wc -l internal/template/templates/.claude/agents/moai/manager-cycle.md
237
```

✅ manager-cycle.md exists, 11385 bytes, 237 lines. Frontmatter has all required fields (name, description, tools CSV, model, permissionMode, memory, skills array, hooks). Migration table (lines 61–70) lists both retired agents.

### 10.3 Verify agentStartHandler Generic Coverage

`internal/hook/subagent_start.go` `agentFrontmatter` struct (lines 159–163):

```go
type agentFrontmatter struct {
    Retired           bool   `yaml:"retired"`
    RetiredReplacement string `yaml:"retired_replacement"`
    RetiredParamHint  string `yaml:"retired_param_hint"`
}
```

`Handle()` line 228: `if !fm.Retired { return &HookOutput{}, nil }` — no agent-name special-casing.
`Handle()` line 234–240: error message uses `agentName` and `fm.RetiredReplacement` dynamically.

✅ Generic implementation: agentStartHandler will block manager-ddd spawn automatically once `retired: true` is set in frontmatter (M2). **No agentStartHandler modification needed for this SPEC** (out of scope per spec.md §2.2).

### 10.4 Verify Test Sentinel Pattern Reusability

Predecessor test `TestRetirementCompletenessAssertion/all retired agents have replacement in embedded FS` (lines 191–227) iterates ALL retired:true agents. Once manager-ddd has `retired: true` + `retired_replacement: manager-cycle`, this test automatically validates manager-cycle.md presence for the manager-ddd retirement → no new code needed in this generic loop. The explicit subtest (§5.2 New test 2) is for **defensive symmetry** with the manager-tdd subtest, not strict necessity.

✅ Sentinel pattern infrastructure already exists; M3 work is additive (new subtests).

---

End of research.md.
