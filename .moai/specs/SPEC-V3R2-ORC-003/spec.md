---
id: SPEC-V3R2-ORC-003
title: "Effort-Level Calibration Matrix for 17 agents"
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: GOOS
priority: P1 High
phase: "v3.0.0 — Phase 3 — Agent Cleanup"
module: ".claude/agents/moai/, .claude/rules/moai/development/agent-authoring.md, .claude/rules/moai/core/moai-constitution.md"
dependencies:
  - SPEC-V3R2-CON-001
  - SPEC-V3R2-ORC-001
  - SPEC-V3R2-ORC-002
related_problem:
  - P-A02
  - P-A03
related_theme: "Layer 4 — Orchestration, Master §4.4, §8 BC-V3R2-002"
breaking: true
bc_id: [BC-V3R2-002]
lifecycle: spec-anchored
tags: "agent, effort, calibration, opus-4-7, adaptive-thinking, v3r2"
---

# SPEC-V3R2-ORC-003: Effort-Level Calibration Matrix for 17 agents

## HISTORY

| Version | Date       | Author | Description                          |
|---------|------------|--------|--------------------------------------|
| 0.1.0   | 2026-04-23 | GOOS   | Initial draft (Wave 4 SPEC writer, round 2) |

---

## 1. Goal (목적)

`.claude/rules/moai/core/moai-constitution.md` §Opus 4.7 Prompt Philosophy prescribes `effort` level selection per agent role: reasoning-intensive agents (manager-spec, manager-strategy, plan-auditor, evaluator-active, expert-security, expert-refactoring) → `xhigh` or `high`; implementation agents → `high`; speed-critical agents (manager-git) → `medium`. R5 audit §Effort-level calibration matrix finds 7 of 22 agents miscalibrated: three agents have explicit wrong values (expert-security, evaluator-active, plan-auditor are declared `high` but the constitution names them for `xhigh`), and 19 of 22 agents omit the field entirely, causing the session default to apply silently — under-invoking Opus 4.7 Adaptive Thinking for implementation-heavy work.

This SPEC publishes the canonical effort-level matrix covering the 17 v3r2 agents (post-ORC-001 consolidation), populates every agent's frontmatter with the correct `effort` value, and promotes lint rule LR-03 from warning to error so that future agents without a declared `effort` fail CI.

### 1.1 Background

Opus 4.7 Adaptive Thinking does not accept a fixed `budget_tokens`; instead, effort levels dynamically allocate reasoning based on task complexity (Anthropic "what's new in claude-4-7" Sep 2025 guidance). Missing `effort:` means the session-level default applies, which per `settings.json` in v2.1.111+ is `medium` — appropriate for some agents but under-invokes reasoning for xhigh candidates.

R5 audit effort matrix:

| Category                     | Agents                                                    | Recommended |
|------------------------------|-----------------------------------------------------------|-------------|
| Reasoning-intensive (xhigh) | manager-spec, manager-strategy, expert-security, expert-refactoring, evaluator-active, plan-auditor, researcher | xhigh       |
| Implementation (high)       | manager-cycle, manager-quality, expert-backend, expert-frontend, expert-performance | high        |
| Template-driven (medium)    | manager-docs, manager-project, expert-devops, builder-platform | medium      |
| Speed-critical (medium)     | manager-git                                               | medium      |

Only `xhigh` and `high` trigger meaningful Opus 4.7 Adaptive Thinking budget allocation; `medium` preserves speed for template or mechanical tasks.

*Source: r5-agent-audit.md §Effort-level calibration matrix; moai-constitution.md §Opus 4.7 Prompt Philosophy; Master §7.2 Agent inventory.*

### 1.2 Non-Goals

- Populating `effort:` on agents outside the 17-agent v3r2 roster (retired agents do not receive a value).
- Dynamic per-spawn effort override (harness-level effort mapping is owned by SPEC-V3R2-HRN-001 `effort_mapping` in harness.yaml).
- Changing Opus 4.7 model selection per agent (model selection is a separate frontmatter field; no changes here).
- Telemetry-driven auto-tuning of effort (Master §12 Open Question #5 defers to post-v3.0).
- Adding new effort levels beyond the 5-value enum `low|medium|high|xhigh|max`.
- Modifying `internal/config/schema/agent.go` (schema is SPEC-V3-AGT-001 territory; this SPEC writes frontmatter values only).

---

## 2. Scope (범위)

### 2.1 In Scope

- Publish the 17-agent effort matrix as a Required Field table in `.claude/rules/moai/development/agent-authoring.md`.
- Populate the `effort:` frontmatter field on each of the 17 v3r2 agents (template tree + local tree byte-identical).
- Promote SPEC-V3R2-ORC-002 lint rule LR-03 from warning to error for every agent in the v3r2 roster.
- Preserve the constitution-level effort guidance in `moai-constitution.md` §Opus 4.7 Prompt Philosophy (no new rule text — only cross-reference to the matrix table).
- Correct the 3 explicit drift cases: expert-security (`high` → `xhigh`), evaluator-active (`high` → `xhigh`), plan-auditor (`high` → `xhigh`).
- Add `effort:` to the 19 agents missing the field (minus the 5 retired in ORC-001), yielding complete coverage on the 17-agent roster.
- Template-first: matrix updates land in `internal/template/templates/.claude/` first; `make build` regenerates; local tree stays byte-identical.

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- Per-SPEC effort overrides at `/moai run` invocation time (future work, Master §12 question #5).
- Telemetry collection of `(effort_level, duration, success)` triples (deferred).
- Auto-recalibration after model upgrades (harness.yaml §model_upgrade_review checklist handles this manually).
- Changing `model:` field on any agent (model vs effort are orthogonal; this SPEC touches only effort).
- Adding effort to output style files, skill files, or commands (agent frontmatter only).
- Retroactive effort assignment to retired v2 agents — they carry `status: retired` per ORC-001.
- Experimental `effort: max` use cases (reserved per constitution guidance until evals show headroom).

---

## 3. Environment (환경)

- Runtime: moai-adk-go v3.0.0-alpha.3+ (Phase 3)
- Claude Code v2.1.110+ required for `effortLevel` setting (coding-standards.md compatibility table)
- Affected frontmatter field: `effort: low|medium|high|xhigh|max` (enum)
- Affected files: 17 agents × 2 trees (template + local) = 34 file edits; 1 rule file (agent-authoring.md); 1 constitution cross-reference
- Lint promotion: ORC-002 LR-03 changes severity from warning to error upon this SPEC's merge
- CI impact: any agent missing `effort:` after this SPEC ships fails `moai agent lint`

---

## 4. Assumptions (가정)

- SPEC-V3R2-ORC-001 has merged; the 17-agent roster is stable.
- SPEC-V3R2-ORC-002 has merged; LR-03 exists as a warning rule.
- Opus 4.7 Adaptive Thinking behavior is consistent with `claude-opus-4-7` model as of Master §Version 14.0.0 (2026-04-03 last updated in CLAUDE.md).
- `effort` values outside the 5-value enum are rejected by existing frontmatter validator (SPEC-V3-AGT-001 enforces `Effort ∈ {"", "low", "medium", "high", "xhigh", "max"}`).
- The 3 explicit-drift corrections (expert-security, evaluator-active, plan-auditor moving from `high` to `xhigh`) represent breaking changes (BC-V3R2-002) since downstream callers may relied on specific reasoning depth; deprecation window is one v3.x cycle with both old and new values acceptable during migration.
- Session default `effort: medium` in `settings.json` continues to apply for agents that legitimately want medium depth.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous (항시)

**REQ-ORC-003-001 (Ubiquitous) — 표준 매트릭스 공표**
The file `.claude/rules/moai/development/agent-authoring.md` **shall** contain a Required Fields table listing the effort level for each of the 17 v3r2 agents using the enumeration `{low, medium, high, xhigh, max}`.

**REQ-ORC-003-002 (Ubiquitous) — 매트릭스 내용**
The effort matrix **shall** prescribe the following values (canonical matrix, this SPEC's deliverable):

| Agent | Effort |
|---|---|
| manager-spec | xhigh |
| manager-strategy | xhigh |
| manager-cycle | high |
| manager-quality | high |
| manager-docs | medium |
| manager-git | medium |
| manager-project | medium |
| expert-backend | high |
| expert-frontend | high |
| expert-security | xhigh |
| expert-devops | medium |
| expert-performance | high |
| expert-refactoring | xhigh |
| builder-platform | medium |
| evaluator-active | xhigh |
| plan-auditor | xhigh |
| researcher | xhigh |

**REQ-ORC-003-003 (Ubiquitous) — 프론트매터 반영**
Each of the 17 v3r2 agent files under `.claude/agents/moai/` **shall** declare an `effort:` frontmatter field whose value matches the canonical matrix in REQ-002.

**REQ-ORC-003-004 (Ubiquitous) — 템플릿 동기화**
The files under `internal/template/templates/.claude/agents/moai/` **shall** carry the same `effort:` values as the local `.claude/agents/moai/` tree (enforced by the existing `diff -r` CI gate per CLAUDE.local.md §2).

**REQ-ORC-003-005 (Ubiquitous) — 상수 참조**
The file `.claude/rules/moai/core/moai-constitution.md` §Opus 4.7 Prompt Philosophy **shall** reference the matrix table in agent-authoring.md (cross-link) without duplicating the table contents.

### 5.2 Event-Driven (이벤트 기반)

**REQ-ORC-003-006 (Event-Driven) — 린트 승격**
**When** this SPEC merges, SPEC-V3R2-ORC-002 lint rule LR-03 ("missing effort: field") **shall** be promoted from warning severity to error severity, causing `moai agent lint` exit code 1 on any agent missing the field.

**REQ-ORC-003-007 (Event-Driven) — 드리프트 보정**
**When** an existing agent frontmatter contains an `effort:` value that disagrees with the canonical matrix, the migrator (SPEC-V3R2-MIG-001) **shall** rewrite the value to the canonical matrix entry and emit a migration log line noting the drift.

### 5.3 State-Driven (상태 기반)

**REQ-ORC-003-008 (State-Driven) — 매트릭스 유일 원천**
**While** the v3.0.0 minor cycle is active, the matrix in `.claude/rules/moai/development/agent-authoring.md` **shall** be the single source of truth for effort-level recommendations; no other document **shall** duplicate the full table (cross-references only).

**REQ-ORC-003-009 (State-Driven) — 새 에이전트 요구**
**While** the lint promotion (REQ-006) is active, every NEW agent added under `.claude/agents/moai/` **shall** declare an `effort:` field; otherwise `moai agent lint` fails and blocks PR merge.

### 5.4 Optional (선택)

**REQ-ORC-003-010 (Optional) — 하니스 오버라이드**
**Where** the harness configuration (SPEC-V3R2-HRN-001) specifies `effort_mapping.thorough: xhigh`, agents spawned at harness level `thorough` **may** override frontmatter effort via the session-level `effortLevel` setting; the agent frontmatter value acts as the default when no harness override applies.

**REQ-ORC-003-011 (Optional) — 드리프트 리포트**
**Where** a contributor runs `moai agent lint --format=json`, the JSON output **may** include an `effort_drift` summary field listing agents whose declared effort disagrees with the canonical matrix (advisory, non-blocking warning).

### 5.5 Unwanted Behavior

**REQ-ORC-003-012 (Unwanted Behavior) — 범위 외 효과 값 금지**
**If** an agent frontmatter declares `effort:` with a value outside the 5-value enum `{low, medium, high, xhigh, max}`, **then** the frontmatter validator (per SPEC-V3-AGT-001 REQ-002) **shall** reject the agent with error `AGT_INVALID_FRONTMATTER` naming the `effort` field and its illegal value.

**REQ-ORC-003-013 (Unwanted Behavior) — 드리프트 재도입 금지**
**If** a PR introduces an `effort:` value different from the canonical matrix for any of the 17 v3r2 agents, **then** CI (via lint rule LR-03 strict mode) **shall** fail with error `ORC_EFFORT_MATRIX_DRIFT` naming the file and the mismatching value.

**REQ-ORC-003-014 (Unwanted Behavior) — fixed budget_tokens 금지**
**If** an agent frontmatter or agent body attempts to set a fixed `budget_tokens` value for Opus 4.7, **then** CI (via ORC-002 LR extension) **shall** fail with error `ORC_FIXED_BUDGET_PROHIBITED` (Opus 4.7 Adaptive Thinking rejects fixed budgets per constitution).

---

## 6. Acceptance Criteria (수용 기준 요약)

Detailed Given-When-Then scenarios are in `acceptance.md`.

Core criteria:

- **AC-ORC-003-01**: `.claude/rules/moai/development/agent-authoring.md` contains the effort matrix table with all 17 agents.
- **AC-ORC-003-02**: Each of the 17 agent files has an `effort:` field matching the canonical matrix.
- **AC-ORC-003-03**: `moai agent lint` on the v3r2 roster exits 0 with no LR-03 warnings.
- **AC-ORC-003-04**: Introducing an agent without `effort:` causes CI to fail with LR-03 error.
- **AC-ORC-003-05**: Changing `effort:` on expert-security from `xhigh` to `high` fails CI with `ORC_EFFORT_MATRIX_DRIFT`.
- **AC-ORC-003-06**: Template and local trees diff equivalent for `effort:` fields.
- **AC-ORC-003-07**: `moai-constitution.md` §Opus 4.7 Prompt Philosophy contains a cross-reference to the matrix but not a duplicate table.
- **AC-ORC-003-08**: Attempting `effort: ultra` (invalid enum) fails with `AGT_INVALID_FRONTMATTER`.
- **AC-ORC-003-09**: Running SPEC-V3R2-MIG-001 migrator on a v2 agent tree rewrites declared-but-drifted effort values and logs each rewrite.
- **AC-ORC-003-10**: The three explicit-drift corrections are visible in a git diff: expert-security, evaluator-active, plan-auditor frontmatter now show `effort: xhigh`.

---

## 7. Constraints (제약)

- [HARD] Breaking change declared (BC-V3R2-002): three agents' effort values change from `high` to `xhigh`; migration window is one v3.x minor cycle.
- [HARD] FROZEN constitution rule preservation: `moai-constitution.md` §Opus 4.7 Prompt Philosophy remains FROZEN; this SPEC only adds a cross-reference; no FROZEN clause text changes.
- [HARD] Template-First: matrix edits in template tree first; `make build` regenerates; local tree byte-identical.
- [HARD] 5-value enum preservation: only `{low, medium, high, xhigh, max}` accepted (REQ-012).
- [HARD] No fixed `budget_tokens` allowed (REQ-014).
- [HARD] Matrix is authoritative (REQ-008): other documents cross-reference.
- [HARD] Lint promotion (REQ-006) timed with SPEC merge to avoid stale-warning regression.

---

## 8. Risks & Mitigations (리스크 및 완화)

| Risk                                                             | Impact | Mitigation                                                                                                   |
|------------------------------------------------------------------|--------|--------------------------------------------------------------------------------------------------------------|
| Three `high` → `xhigh` upgrades cause latency regression on some workloads | MEDIUM | BC-V3R2-002 declared; telemetry at beta.1; revert per-agent if latency regression >20%                       |
| `xhigh` over-invokes Adaptive Thinking on trivial plan-auditor tasks | MEDIUM | Harness routing (HRN-001) can override per level; `minimal` harness still maps to `medium`                   |
| Contributors add new agent without `effort:` field              | MEDIUM | LR-03 strict mode blocks PR merge (REQ-006, REQ-009)                                                         |
| Matrix drift between constitution and agent-authoring           | MEDIUM | Cross-reference only (REQ-005, REQ-008); lint fails if matrix drift detected                                 |
| Opus 4.7 guidance evolves (new effort level added)              | LOW    | Harness review cycle (harness.yaml §model_upgrade_review) captures this                                      |
| Fixed `budget_tokens` reintroduced by well-meaning contributor  | LOW    | REQ-014 lint rejection + constitution HARD rule                                                              |
| Non-matrix agents (e.g., dev-local 98/99 commands) confuse lint | LOW    | Scope scan to `.claude/agents/moai/*.md` only; 98/99 commands are not agents                                 |
| Effort matrix table format drift breaks automated parsing        | LOW    | Canonical format declared in agent-authoring.md with regression test                                         |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3R2-CON-001** (FROZEN/EVOLVABLE codification)
- **SPEC-V3R2-ORC-001** (17-agent roster baseline)
- **SPEC-V3R2-ORC-002** (LR-03 lint rule exists as warning)

### 9.2 Blocks

- **SPEC-V3R2-HRN-001** (Harness routing) — uses effort_mapping aligned to this matrix.
- **SPEC-V3R2-MIG-001** (migrator) — references this SPEC's matrix when rewriting v2 agent effort values.

### 9.3 Related

- **SPEC-V3-AGT-001** (Agent Frontmatter Expansion) — Effort enum definition (v3-legacy inherited by v3r2).
- **SPEC-V3R2-CON-002** (Amendment protocol) — future matrix changes pass through graduation.

---

## 10. Traceability (추적성)

- REQ-to-AC mapping: REQ-001 → AC-01; REQ-002 → AC-01, AC-02, AC-10; REQ-003 → AC-02, AC-10; REQ-004 → AC-06; REQ-005 → AC-07; REQ-006 → AC-03, AC-04; REQ-007 → AC-09; REQ-008 → AC-07; REQ-009 → AC-04; REQ-010 → HRN-001 cross-test; REQ-011 → JSON field regression test; REQ-012 → AC-08; REQ-013 → AC-05; REQ-014 → CI fixture test.
- Total REQ count: 14 (Ubiquitous 5, Event-Driven 2, State-Driven 2, Optional 2, Unwanted 3)
- Expected AC count: 10
- Wave 1/2 sources:
  - `r5-agent-audit.md` §Effort-level calibration matrix (drift rate 32%)
  - `moai-constitution.md` §Opus 4.7 Prompt Philosophy (enum + role-to-effort mapping)
  - `problem-catalog.md` P-A02 (HIGH, 19/22 missing field), P-A03 (HIGH, 3 explicit drifts)
  - `major-v3-master.md` §8 BC-V3R2-002, §1.2 changes table
  - Anthropic "what's new in claude-4-7" Sep 2025 guidance (Adaptive Thinking behavior)
- Code-side paths:
  - `.claude/rules/moai/development/agent-authoring.md` (modified — matrix table, REQ-001, REQ-002)
  - `.claude/rules/moai/core/moai-constitution.md` (modified — cross-reference, REQ-005)
  - `.claude/agents/moai/manager-spec.md` (modified, REQ-003 — declare xhigh)
  - `.claude/agents/moai/manager-strategy.md` (modified, REQ-003 — declare xhigh)
  - `.claude/agents/moai/manager-cycle.md` (created by ORC-001, populated by REQ-003 — high)
  - `.claude/agents/moai/manager-quality.md` (modified, REQ-003 — high)
  - `.claude/agents/moai/manager-docs.md` (modified, REQ-003 — medium)
  - `.claude/agents/moai/manager-git.md` (modified, REQ-003 — medium)
  - `.claude/agents/moai/manager-project.md` (modified, REQ-003 — medium)
  - `.claude/agents/moai/expert-backend.md` (modified, REQ-003 — high)
  - `.claude/agents/moai/expert-frontend.md` (modified, REQ-003 — high)
  - `.claude/agents/moai/expert-security.md` (modified, REQ-003 — xhigh, upgrades from high)
  - `.claude/agents/moai/expert-devops.md` (modified, REQ-003 — medium)
  - `.claude/agents/moai/expert-performance.md` (modified, REQ-003 — high)
  - `.claude/agents/moai/expert-refactoring.md` (modified, REQ-003 — xhigh)
  - `.claude/agents/moai/builder-platform.md` (created by ORC-001, populated by REQ-003 — medium)
  - `.claude/agents/moai/evaluator-active.md` (modified, REQ-003 — xhigh, upgrades from high)
  - `.claude/agents/moai/plan-auditor.md` (modified, REQ-003 — xhigh, upgrades from high)
  - `.claude/agents/moai/researcher.md` (modified, REQ-003 — xhigh)
  - `internal/template/templates/.claude/...` (template-first mirrors, REQ-004)
- **File:line anchors** (per D5 traceability requirement):
  - `docs/design/major-v3-master.md:L1054` (§11.4 ORC-003 definition)
  - `docs/design/major-v3-master.md:L961` (§8 BC-V3R2-002 — effort field)
  - `.moai/design/v3-redesign/synthesis/problem-catalog.md` (P-A02, P-A03)

---

End of SPEC.
