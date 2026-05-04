---
id: SPEC-V3R3-RETIRED-DDD-001
title: manager-ddd Retired Stub Standardization (Predecessor Pattern Replication)
version: "0.3.0"
status: draft
created_at: 2026-05-04
updated_at: 2026-05-04
author: MoAI Plan Workflow
priority: P1
labels: [agent-runtime, templates, retired-stub, manager-cycle, manager-ddd, follow-up, v3r3]
issue_number: 778
phase: "v3.0.0 — Phase 7 — Agent Runtime Robustness (Follow-up)"
module: "internal/template/templates/.claude/agents/moai/, internal/template/templates/.claude/rules/moai/, internal/template/templates/.claude/skills/, internal/template/templates/CLAUDE.md"
dependencies:
  - SPEC-V3R3-RETIRED-AGENT-001
related_specs:
  - SPEC-V3R2-ORC-001
  - SPEC-V3R3-RETIRED-AGENT-001
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "agent-runtime, retired-stub, manager-cycle, manager-ddd, template-parity, predecessor-pattern, audit-test, v3r3"
---

# SPEC-V3R3-RETIRED-DDD-001: manager-ddd Retired Stub Standardization

## HISTORY

| Version | Date       | Author              | Description                                                                                                                                |
|---------|------------|---------------------|--------------------------------------------------------------------------------------------------------------------------------------------|
| 0.3.0   | 2026-05-04 | MoAI Plan Workflow  | Force-accept after iter 3 (score 0.78, must-pass MP-1~4 PASS). Manual textual sync — 4 fixes at spec.md:L170 + plan.md §3 + Risk row + C18. |
| 0.2.0   | 2026-05-04 | MoAI Plan Workflow  | iter 2 regression fix — D-NEW-1~5 cross-artifact partial sync resolved.                                                                     |
| 0.1.0   | 2026-05-04 | MoAI Plan Workflow  | 최초 작성. 이슈 #778 base. Predecessor SPEC-V3R3-RETIRED-AGENT-001 패턴을 manager-ddd에 동일 적용. iter 1 plan-audit FAIL 0.78 (D1-D8) → iter 2/3 진행. |

---

## 1. Goal (목적)

`manager-tdd` retire 패턴 (SPEC-V3R3-RETIRED-AGENT-001 / PR #776 머지 완료, main `20d77d931`)을 **manager-ddd**에 동일하게 적용하여 `manager-cycle`을 진정한 unified DDD/TDD agent로 정립한다.

### 1.1 배경 (Why this SPEC exists)

predecessor SPEC-V3R3-RETIRED-AGENT-001 (이하 RETIRED-AGENT)이 manager-tdd 단일 케이스만 표준화하면서 다음 두 가지 incomplete state를 남겼다:

1. **manager-ddd retired stub 표준화 미적용**: `internal/template/templates/.claude/agents/moai/manager-ddd.md` (현재 7674 bytes, full agent definition)는 기존 DDD 에이전트 정의를 그대로 유지한다. RETIRED-AGENT spec.md §1.3 / §2.2가 이를 "out-of-scope, 후속 SPEC SPEC-V3R3-RETIRED-DDD-001 (가칭)에서 처리" 로 명시.
2. **30+ documentation cross-references in templates still cite `manager-ddd`**: `manager-cycle`로 substitution이 partial하다 (RETIRED-AGENT에서 manager-tdd → manager-cycle만 substitute). predecessor 머지 후 grep evidence: `internal/template/templates/`에 `manager-ddd` 언급 34 files.

이 SPEC은 manager-ddd retire를 RETIRED-AGENT의 frontmatter standard (REQ-RA-002), audit test pattern (REQ-RA-016), documentation substitution discipline (REQ-RA-013)을 그대로 차용해 종결시킨다.

### 1.2 핵심 시정 조치

- **P0 (Immediate)**:
  1. `manager-ddd.md` Cat B B1 rewrite — 7674-byte full agent definition을 ~1.4KB retired stub으로 전환, predecessor `manager-tdd.md` retired stub와 동일한 5-field 표준 frontmatter 적용 (`retired: true`, `retired_replacement: manager-cycle`, `retired_param_hint: "cycle_type=ddd"`, `tools: []`, `skills: []`).
  2. 30 Cat A SUBSTITUTE files: documentation references `manager-ddd` → `manager-cycle` (또는 `manager-cycle with cycle_type=ddd`) 일괄 치환.
  3. 2 Cat C UPDATE-WITH-ANNOTATION files: `manager-ddd`가 retired 사실로서 explicit 언급되어야 하는 위치 (예: agent-hooks.md 표 row, agent-authoring.md Manager Agents listing)에 retired marker 추가.
- **P1 (Audit + CI)**:
  4. `internal/template/agent_frontmatter_audit_test.go` 확장 — RETIRED-AGENT M1에서 작성된 audit test가 manager-tdd만 검증한다. manager-ddd도 동일 5-field 표준 검증하도록 확장 (테이블 row 추가). REQ-RA-002 / REQ-RA-016 audit pattern 재사용.
  5. `internal/hook/agents/factory.go` — 기존 `ddd_handler.go` action key (`ddd-pre-transformation`, `ddd-post-transformation`, `ddd-completion`) dispatch는 유지 (mo.ai.kr 외부 invocation backwards compat). 단 retired path를 통과한 `manager-ddd` spawn은 SubagentStart guard (RETIRED-AGENT REQ-RA-007)에서 block. factory dispatch 자체는 변경 없음.
- **P2 (Documentation + CHANGELOG)**:
  6. `CHANGELOG.md` `[Unreleased]` 섹션 — manager-ddd retirement entry 추가 + `moai update` 권고 메시지.

### 1.3 비목표 (Non-Goals)

- `manager-cycle.md` 재작성 (predecessor SPEC가 이미 unified DDD/TDD body를 정립; 본 SPEC은 cross-references만 손댐).
- `subagent_start.go` agentStartHandler 본체 수정 (predecessor 구현 그대로; manager-ddd retired stub frontmatter parsing은 동일 코드 path로 처리).
- mo.ai.kr 사이드 프로젝트 직접 수정 (`moai update` 자동 sync로 처리 — predecessor와 동일 정책).
- 다른 agent 추가 retire (manager-tdd / manager-ddd 외 agents의 retirement는 별도 SPEC).
- agency/* (copywriter, designer 등) retired agent 일괄 처리 (SPEC-AGENCY-ABSORB-001 영역).
- worktreePath empty-object validation 등 P1 layer (RETIRED-AGENT REQ-RA-005, REQ-RA-010에서 이미 구현됨).
- text/template path interpolation 마이그레이션 (RETIRED-AGENT REQ-RA-006 영역).

---

## 2. Scope (범위)

### 2.1 In Scope — File Taxonomy (33 + 3 ancillary = 36 grand total)

**Cat A — SUBSTITUTE (30 files)**: documentation reference `manager-ddd` → `manager-cycle` (또는 cycle_type=ddd 명시) 단순 치환.

| # | File | Notes |
|---|------|-------|
| A-01 | `internal/template/templates/CLAUDE.md` | §4 Manager Agents listing, §5 Agent Chain |
| A-02 | `internal/template/templates/.claude/agents/moai/manager-cycle.md` | Inline references (3 occurrences) |
| A-03 | `internal/template/templates/.claude/agents/moai/manager-spec.md` | Cross-reference |
| A-04 | `internal/template/templates/.claude/agents/moai/manager-strategy.md` | Cross-reference |
| A-05 | `internal/template/templates/.claude/agents/moai/manager-quality.md` | Cross-reference |
| A-06 | `internal/template/templates/.claude/agents/moai/expert-backend.md` | Cross-reference |
| A-07 | `internal/template/templates/.claude/agents/moai/expert-frontend.md` | Cross-reference |
| A-08 | `internal/template/templates/.claude/agents/moai/expert-mobile.md` | Cross-reference |
| A-09 | `internal/template/templates/.claude/agents/moai/expert-debug.md` | Cross-reference |
| A-10 | `internal/template/templates/.claude/agents/moai/expert-devops.md` | Cross-reference |
| A-11 | `internal/template/templates/.claude/agents/moai/expert-testing.md` | Cross-reference |
| A-12 | `internal/template/templates/.claude/agents/moai/expert-refactoring.md` | Cross-reference |
| A-13 | `internal/template/templates/.claude/agents/moai/evaluator-active.md` | Cross-reference |
| A-14 | `internal/template/templates/.claude/agents/moai/manager-tdd.md` | Migration table & sibling reference |
| A-15 | `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` | Workflow doc |
| A-16 | `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` | HARD rule example |
| A-17 | `internal/template/templates/.claude/output-styles/moai/moai.md` | Forced Delegation Table |
| A-18 | `internal/template/templates/.claude/skills/moai/SKILL.md` | Workflow router |
| A-19 | `internal/template/templates/.claude/skills/moai/workflows/run.md` | Run workflow |
| A-20 | `internal/template/templates/.claude/skills/moai/workflows/moai.md` | Default workflow |
| A-21 | `internal/template/templates/.claude/skills/moai/references/mx-tag.md` | MX tag reference |
| A-22 | `internal/template/templates/.claude/skills/moai/references/reference.md` | General reference |
| A-23 | `internal/template/templates/.claude/skills/moai-foundation-cc/SKILL.md` | CC foundation |
| A-24 | `internal/template/templates/.claude/skills/moai-foundation-core/SKILL.md` | Core foundation |
| A-25 | `internal/template/templates/.claude/skills/moai-foundation-quality/SKILL.md` | Quality foundation |
| A-26 | `internal/template/templates/.claude/skills/moai-meta-harness/SKILL.md` | Meta-harness |
| A-27 | `internal/template/templates/.claude/skills/moai-workflow-loop/SKILL.md` | Loop workflow |
| A-28 | `internal/template/templates/.claude/skills/moai-workflow-spec/SKILL.md` | Spec workflow |
| A-29 | `internal/template/templates/.claude/skills/moai-workflow-spec/references/reference.md` | Spec workflow reference |
| A-30 | `internal/template/templates/.claude/skills/moai-workflow-spec/references/examples.md` | Spec workflow examples |

**Cat B — REWRITE (1 file)**: full transformation.

| # | File | Notes |
|---|------|-------|
| B-01 | `internal/template/templates/.claude/agents/moai/manager-ddd.md` | 7674 bytes → ~1.4KB retired stub. 5-field frontmatter standard. Migration table body. |

**Cat C — UPDATE-WITH-ANNOTATION (2 files)**: retire fact가 explicit 언급되는 documentation table/listing.

| # | File | Notes |
|---|------|-------|
| C-01 | `internal/template/templates/.claude/rules/moai/core/agent-hooks.md` | Agent Hook Actions table — manager-ddd row 유지 (action key dispatch는 backwards compat 위해 보존), 단 row에 retired marker 추가 |
| C-02 | `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` | Manager Agents (8) listing — manager-ddd 항목 제거 또는 retired marker 부여 |

**Ancillary (3 files)**:

| # | File | Notes |
|---|------|-------|
| ANC-01 | `internal/template/agent_frontmatter_audit_test.go` | RETIRED-AGENT가 작성한 audit test 확장 — manager-ddd row 추가 |
| ANC-02 | `internal/hook/agents/factory.go` | Reference verification only (기존 dispatch는 변경 없음, 단 retired guard와의 interaction 검증) |
| ANC-03 | `CHANGELOG.md` | `[Unreleased]` 섹션 manager-ddd retirement entry |

**SPEC artifacts**:
- `.moai/specs/SPEC-V3R3-RETIRED-DDD-001/{spec.md, plan.md, acceptance.md, research.md, spec-compact.md}`
- Plan-audit reports: `.moai/reports/plan-audit/SPEC-V3R3-RETIRED-DDD-001-review-{1,2,3}.md`

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- 구현 코드 변경 (Go function body 변경) — Cat C ANC-02 (factory.go) 검증만; dispatch logic 변경 없음.
- `manager-cycle.md` 재작성 또는 body 확장 (predecessor가 unified body 정립; 본 SPEC은 cross-references만 substitute).
- `manager-ddd` 부활 또는 cycle 설계 변경 (SPEC-V3R2-ORC-001 retirement decision 유지).
- `subagent_start.go` 본체 수정 (predecessor 구현 그대로 사용 — retired stub frontmatter parsing은 동일 path).
- `worktreePath` validation 또는 `text/template` migration (RETIRED-AGENT P1 영역, 이미 구현됨).
- mo.ai.kr 사이드 프로젝트 직접 수정 (`moai update`로 sync).
- 신규 agent retirement workflow / governance 정책 (별도 SPEC; 본 SPEC은 단일 케이스 fix).
- agency/* retired agent 일괄 표준화 (SPEC-AGENCY-ABSORB-001 영역).

### 2.3 CHANGELOG

본 SPEC은 BREAKING CHANGE 없음 (`breaking: false`, `bc_id: []`). 사용자가 `manager-ddd`로 직접 호출하던 경우 retirement message + migration hint를 받게 되며, 이는 backwards-compatible behavior (predecessor와 동일 패턴).

---

## 3. Environment (환경)

- 런타임: Go 1.23+, Cobra CLI, Claude Code Agent() runtime v2.1.97+, bash hook wrappers.
- 영향 디렉터리: `internal/template/templates/.claude/agents/moai/`, `internal/template/templates/.claude/rules/moai/`, `internal/template/templates/.claude/skills/`, `internal/template/templates/CLAUDE.md`, `internal/template/`, `internal/hook/agents/` (verification only).
- 외부 endpoints: 없음.
- 외부 레퍼런스 (codebase grounded scan):
  - `internal/template/templates/.claude/agents/moai/manager-ddd.md` (현재 7674 bytes, retired transition target)
  - `internal/template/templates/.claude/agents/moai/manager-cycle.md` (predecessor 설계 reference)
  - `internal/template/templates/.claude/agents/moai/manager-tdd.md` (predecessor가 표준화한 reference stub, 1392 bytes)
  - `internal/template/agent_frontmatter_audit_test.go` (predecessor M1 작성 audit, 확장 target)
  - 30 Cat A files (위 §2.1 표) — `manager-ddd` 언급 위치
  - `internal/template/templates/.claude/rules/moai/core/agent-hooks.md` (Cat C C-01)
  - `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` (Cat C C-02)
- 의존: PR #776 머지 완료 (main `20d77d931`); `feature/SPEC-V3R3-RETIRED-DDD-001` 브랜치는 origin/main 기반.

---

## 4. Assumptions (가정)

- predecessor SPEC-V3R3-RETIRED-AGENT-001이 manager-tdd retire stub frontmatter standard (REQ-RA-002), audit test pattern (REQ-RA-016), SubagentStart guard (REQ-RA-007)를 모두 정립했다. 본 SPEC은 동일 standard를 manager-ddd에 적용하는 mechanical replication.
- `internal/template/agent_frontmatter_audit_test.go`는 predecessor M1에서 작성됨 (table-driven test). manager-ddd row 추가는 single-line addition.
- Claude Code Agent runtime이 invalid frontmatter (e.g., `tools: []` empty array)에 대해 graceful 처리하며 0 tool_uses agent를 spawn한다 (RETIRED-AGENT 검증 결과). SubagentStart guard가 spawn 전 차단 — manager-ddd retired stub도 동일 path로 reject.
- `manager-cycle` agent는 `cycle_type=ddd` 파라미터를 받아 ANALYZE-PRESERVE-IMPROVE 실행을 수행한다 (predecessor body에 정의됨). Migration table은 사용자에게 정확한 새 호출 패턴을 제시.
- 30 Cat A SUBSTITUTE 파일 grep evidence는 `grep -rl "manager-ddd" internal/template/templates/`로 enumerate 가능. iter 3 plan-audit 시점 evidence 기반 (34 files identified, 30 SUBSTITUTE + 1 REWRITE + 2 UPDATE-WITH-ANNOTATION + 1 manager-tdd.md cross-ref가 Cat A에 포함된 건수 = 34).
- `moai update` 실행 시 사용자 프로젝트는 본 SPEC의 fix를 자동 sync. 사용자 데이터 (.moai/specs/, .moai/project/) 보존.
- TDD methodology (`.moai/config/sections/quality.yaml development_mode: tdd`): RED (M1 audit test 확장) → GREEN (M2-M4 substitution + REWRITE + Cat C annotation) → REFACTOR (M5 docs + CHANGELOG).
- Solo mode, no worktree (predecessor와 동일 운영 정책).

---

## 5. Requirements (EARS 요구사항)

총 **12개 REQs** — Ubiquitous 4, Event-Driven 3, State-Driven 3, Optional 1, Unwanted 1.

### 5.1 Ubiquitous Requirements (always active)

**REQ-RD-001**
The MoAI template tree **shall** transform `internal/template/templates/.claude/agents/moai/manager-ddd.md` from a full agent definition (currently 7674 bytes) into a retired stub (~1.4KB), preserving the canonical retirement frontmatter pattern established by `manager-tdd.md` (predecessor SPEC-V3R3-RETIRED-AGENT-001 REQ-RA-002): `retired: true` (boolean), `retired_replacement: manager-cycle`, `retired_param_hint: "cycle_type=ddd"`, `tools: []`, `skills: []`.

**REQ-RD-002**
The MoAI template tree **shall** substitute every `manager-ddd` reference within the 30 Cat A SUBSTITUTE files enumerated in §2.1 with `manager-cycle` (or `manager-cycle with cycle_type=ddd` where the parameter is contextually relevant). Substitutions **shall** preserve enclosing prose semantics and **shall not** introduce drive-by edits.

**REQ-RD-003**
The MoAI template tree **shall** annotate the 2 Cat C UPDATE-WITH-ANNOTATION files (`agent-hooks.md` Agent Hook Actions table; `agent-authoring.md` Manager Agents listing) so that `manager-ddd` is explicitly marked as retired (e.g., `manager-ddd (retired — see manager-cycle)`) where its action keys are still dispatched for backwards compatibility, and so that the Manager Agents (8) listing reflects the unified `manager-cycle` replacement.

**REQ-RD-004**
The `internal/template/agent_frontmatter_audit_test.go` audit (extended from predecessor M1 scaffolding) **shall** include a manager-ddd table-driven row that asserts the same 5-field retirement frontmatter standard. The test **shall** fail if any of `retired: true` / `retired_replacement: manager-cycle` / `retired_param_hint: "cycle_type=ddd"` / `tools: []` / `skills: []` are missing or malformed.

### 5.2 Event-Driven Requirements

**REQ-RD-005**
**When** the SubagentStart hook (predecessor REQ-RA-004 / REQ-RA-007) receives a spawn event for `manager-ddd` after this SPEC merges, the existing handler **shall** parse the standardized frontmatter, extract `retired_replacement: manager-cycle` and `retired_param_hint: "cycle_type=ddd"`, and emit a `block` decision with reason text instructing the caller to use `manager-cycle` with `cycle_type=ddd`.

**REQ-RD-006**
**When** `make build` regenerates `internal/template/embedded.go` after Cat A/B/C edits, the regeneration **shall** include the standardized `manager-ddd.md` retired stub plus all 30 SUBSTITUTE updates plus the 2 ANNOTATION updates. CI **shall** detect missing regeneration via `git diff internal/template/embedded.go` non-empty.

**REQ-RD-007**
**When** `internal/hook/agents/factory.go` is dispatched with `event = "agent-hook"` and an action key starting with `ddd-` (e.g., `ddd-pre-transformation`, `ddd-completion`), the existing handler **shall** continue to dispatch to `ddd_handler.go` (backwards compatibility for any existing invocations on user projects pre-update). The factory itself is unchanged; only the upstream guard (REQ-RD-005) prevents new spawns.

### 5.3 State-Driven Requirements

**REQ-RD-008**
**While** `retired: true` is present in `manager-ddd.md` frontmatter, the body content **shall** describe (a) retirement reason (consolidation into `manager-cycle`), (b) replacement agent name (`manager-cycle`), (c) migration command pattern (old → new invocation table). Documentation lookup at retirement time **shall** be self-contained, mirroring the `manager-tdd.md` body precedent.

**REQ-RD-009**
**While** `manager-cycle` is the active unified DDD/TDD implementation agent, the documentation references in CLAUDE.md (§4 Manager Agents, §5 Agent Chain), `agent-authoring.md` (Manager Agents listing), `agent-hooks.md` (Agent Hook Actions table), `spec-workflow.md`, `worktree-integration.md` HARD rule example, and 13+ other Cat A files **shall** all reference `manager-cycle` (or `manager-cycle with cycle_type=ddd`) instead of `manager-ddd`. Inconsistencies **shall** be detected by the extended audit test.

**REQ-RD-010**
**While** the `agent_frontmatter_audit_test.go` table-driven test exists, it **shall** be the single source of truth for retired-agent frontmatter compliance: any future agent retirement (or regression that re-adds full agent body to a retired stub) **shall** cause CI failure. The test row schema is `{name string, expected_replacement string, expected_param_hint string}` and applies uniformly to manager-tdd and manager-ddd.

### 5.4 Optional Requirements

**REQ-RD-011**
**Where** the user wants to verify completion of manager-ddd retirement in their installation, the existing `moai agents list --retired` subcommand (predecessor REQ-RA-014) **shall** surface both `manager-tdd` and `manager-ddd` if implemented; otherwise this requirement is informational only and deferred per predecessor decision (REQ-RA-014 was deferred). No new CLI surface is introduced by this SPEC.

### 5.5 Unwanted Behavior Requirements

**REQ-RD-012 (Unwanted Behavior — Composite)**
**If** the manager-ddd retirement is partially applied — e.g., Cat B B-01 frontmatter standardized but ≥1 Cat A SUBSTITUTE file still cites `manager-ddd` in user-facing prose, or the audit test row is missing for manager-ddd — **then** CI **shall** fail with a `RETIREMENT_INCOMPLETE_manager-ddd` style assertion produced by `agent_frontmatter_audit_test.go`. **And** the SubagentStart guard (REQ-RD-005) **shall** never produce silent acceptance (exit 0) for a `manager-ddd` spawn whose frontmatter contains `retired: true`. **And** documentation references **shall not** mix `manager-ddd` (active) with `manager-cycle` (active) language in the same prose context — the audit pass is binary, no drift permitted.

---

## 6. Acceptance Criteria (수용 기준 요약, 4 ACs covering Positive + Edge + Boundary + Negative)

상세 Given/When/Then은 `acceptance.md` 참조. Summary mapping:

- **AC-RD-01 (Positive scenario)** — REQ-RD-001, REQ-RD-002, REQ-RD-006: After M2/M3 implementation + `make build`, `manager-ddd.md` is retired stub, all 30 Cat A files substituted, embedded.go regenerated. End-to-end golden path.
- **AC-RD-02 (Edge scenario)** — REQ-RD-003, REQ-RD-008, REQ-RD-009: Cat C UPDATE-WITH-ANNOTATION files (`agent-hooks.md` + `agent-authoring.md`) preserve action-key dispatch (REQ-RD-007 invariant) while annotating retired status. Migration table body of new stub matches predecessor format.
- **AC-RD-03 (Boundary scenario)** — REQ-RD-004, REQ-RD-010, REQ-RD-012: `agent_frontmatter_audit_test.go` extended with manager-ddd row passes; tampering with any of the 5 standardized fields causes immediate CI failure with `RETIREMENT_INCOMPLETE_manager-ddd`-style message.
- **AC-RD-04 (Negative scenario)** — REQ-RD-005, REQ-RD-007, REQ-RD-012: Attempting `Agent({subagent_type: "manager-ddd"})` post-merge results in SubagentStart guard `block` decision within ≤500ms (predecessor REQ-RA-012 inheritance); no worktree allocated, no `worktreePath: {}` propagation, no path-interpolation `/{}/{}` literal observed. `factory.go` `ddd_*` action-key dispatch path remains intact for hook events that bypass the SubagentStart guard.

REQ↔AC matrix (12 REQs × 4 ACs, full coverage):

| REQ | Mapped AC(s) |
|---|---|
| REQ-RD-001 | AC-RD-01 |
| REQ-RD-002 | AC-RD-01 |
| REQ-RD-003 | AC-RD-02 |
| REQ-RD-004 | AC-RD-03 |
| REQ-RD-005 | AC-RD-04 |
| REQ-RD-006 | AC-RD-01 |
| REQ-RD-007 | AC-RD-04 |
| REQ-RD-008 | AC-RD-02 |
| REQ-RD-009 | AC-RD-02 |
| REQ-RD-010 | AC-RD-03 |
| REQ-RD-011 | (informational; deferred per predecessor) |
| REQ-RD-012 | AC-RD-03, AC-RD-04 |

Coverage: **11/12 REQs mapped to 4/4 ACs**; REQ-RD-011 deferred per predecessor decision (still listed for completeness and to mirror RETIRED-AGENT REQ-RA-014).

---

## 7. Constraints (제약)

- **TDD methodology** (per `.moai/config/sections/quality.yaml development_mode: tdd`): RED (M1 audit row + regression test) → GREEN (M2-M4 substitution + REWRITE + ANNOTATION) → REFACTOR (M5 CHANGELOG + lessons).
- **16-language neutrality** (CLAUDE.local.md §15): retired stub body는 language-agnostic — migration table은 language-neutral 호출 패턴만 명시.
- **Template-First HARD rule** (CLAUDE.local.md §2): 모든 변경은 `internal/template/templates/` 미러 + `make build` 필수. 직접 `.claude/`나 mo.ai.kr 수정 금지.
- **No drive-by refactor** (CLAUDE.md §7 Rule 2 / Agent Core Behavior #5): Cat A files은 `manager-ddd` → `manager-cycle` substitution scope로 한정. 다른 개선 사항은 별도 SPEC.
- **No flat file** (manager-spec skill HARD rule): `.moai/specs/SPEC-V3R3-RETIRED-DDD-001/` 디렉토리 + 5 files 필수 (spec.md, plan.md, acceptance.md, research.md, spec-compact.md).
- **API-key 보안 무관**: 본 SPEC은 secret 다루지 않음.
- **No flag in branch protection bypass**: `feature/SPEC-V3R3-RETIRED-DDD-001` 브랜치는 admin override 없이 PR + CI green 후 merge.
- **Solo mode, no worktree**: 단일 브랜치 작성. Implementation 단계에서 worktree 사용 여부는 별도 결정.
- **Predecessor-pattern fidelity**: 본 SPEC의 frontmatter standard, audit test pattern, migration table format, SubagentStart guard interaction은 모두 RETIRED-AGENT precedent를 verbatim 차용. 새로운 패턴 도입 금지.

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크 | 영향 | 확률 | 완화 |
|---|---|---|---|
| 30 Cat A SUBSTITUTE 중 일부가 grep evidence에서 누락되어 partial substitution 발생 | M | M | M2 단계에서 `grep -rln "manager-ddd" internal/template/templates/`로 enumeration 재검증; audit test가 audit-time CI assertion으로 catch (REQ-RD-012). |
| Cat B B-01 manager-ddd.md retire stub 전환 시 사용자가 기존 외부 호출 (e.g., 쉘 스크립트, CI step)로 spawn 시 11.4s 0-tool_uses 증상 (predecessor 5-layer chain layer 1) 재발 | H | L | predecessor SubagentStart guard (REQ-RA-007)가 이미 `retired: true` frontmatter detect하여 block; manager-ddd retired stub도 동일 path로 reject ≤500ms. predecessor MERGED main `20d77d931`로 verified. |
| Cat C C-01 (agent-hooks.md) Agent Hook Actions 표에서 `manager-ddd` row를 단순 제거 시, `factory.go` `ddd_handler.go` action key dispatch가 orphan으로 남아 documentation drift 발생 | M | M | C-01에서 row 유지 + retired marker 부착 ("manager-ddd (retired — dispatched only for backward compat; new spawns blocked by SubagentStart)"). REQ-RD-003 + REQ-RD-007 invariant 양립. |
| audit test row 추가 시 기존 manager-tdd row와 schema가 미묘하게 달라 false-pass 발생 | M | L | M1 RED phase에서 동일 table-driven schema로 row 추가 (struct field 동일); table loop 자체는 변경 없음. failing run 확인 후 GREEN 진행. |
| `moai update` 실행 시 사용자 로컬 `manager-ddd.md` (수정본) overwrite로 사용자 작업 손실 | L | M | CLAUDE.local.md §2 protected directories는 `.moai/project/`, `.moai/specs/` 만 보호; `.claude/agents/` 는 template sync 대상. moai update가 backup 생성 후 sync (기존 동작). |
| Cat A 30 files에서 `manager-ddd`가 코드 식별자 또는 파일 경로 (e.g., 디렉토리 이름, hook action key)로 등장하는 경우 substitution이 실제 코드 path를 깨뜨림 | H | L | M2 시작 전 grep evidence 분류 — prose vs identifier. action key (`ddd-pre-transformation` 등)는 변경 금지 (REQ-RD-007 invariant); prose만 substitute. iter 2 plan-audit (force-accept 사유 §3 manual sync)에서 검증됨. |
| 30 SUBSTITUTE 작업이 cross-artifact partial sync 상태로 진행되어 plan-audit iter 2와 동일 regression 재발 (D-NEW-1~5 패턴) | M | M | M2 단계에서 grep evidence 기반 file list를 single commit으로 묶음; M3 (Cat C)와 분리. PR description에 file checklist 포함. |
| 30 files 중 `manager-cycle` and `manager-ddd` co-mention prose가 있어 단순 substitute로 의미가 깨짐 | M | M | M2에서 file-by-file Read + Edit (replace_all 금지); ambiguous 위치는 `manager-cycle with cycle_type=ddd` 명시. |
| predecessor merge 후 main에서 manager-cycle.md 정의가 변경되어 retired stub의 `retired_replacement: manager-cycle` 참조가 stale | L | L | feature 브랜치 base는 `origin/main` `20d77d931` (predecessor merged); 본 SPEC merge 시점에 manager-cycle.md 변경 여부 PR review에서 확인. |
| `agent_frontmatter_audit_test.go`가 predecessor M1에서 작성되지 않았거나 다른 파일명일 경우 M1 RED phase가 시작 자체 불가 | H | L | research.md §2에서 predecessor M1 산출 파일명 + 라인 검증; 없을 경우 M1에서 audit_test.go 파일을 신규 작성 (predecessor 패턴 기반 새로 작성 가능). |
| Cat C C-02 (agent-authoring.md Manager Agents listing) 변경 시 agent count "8"이 일관성을 잃음 (manager-tdd 이미 retired, manager-ddd 추가 retire = effective 6 active managers + manager-cycle = 7 effective; 표기 "8"은 manager-cycle 추가 후) | L | M | predecessor와 동일 정책: "8" 표기는 active agents 기준 — manager-tdd retired (-1) + manager-ddd retired (-1) + manager-cycle (+2 effective: tdd+ddd consolidation) = effective 8 유지. C-02에서 listing 자체는 manager-cycle 포함, manager-tdd / manager-ddd 모두 retired marker. |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3R3-RETIRED-AGENT-001** (PR #776, MERGED main `20d77d931`): manager-cycle.md 정의, retired stub frontmatter standard (REQ-RA-002), audit test scaffolding (REQ-RA-016), SubagentStart guard (REQ-RA-007), wrapper validation (REQ-RA-005)이 모두 정립됨. 본 SPEC은 이 패턴을 manager-ddd에 mechanical replication.

### 9.2 Blocks

- 향후 agent retirement workflow / governance SPEC: 본 SPEC + RETIRED-AGENT 두 케이스가 패턴 generalization의 사례 base.

### 9.3 Related

- **SPEC-V3R2-ORC-001** (Agent Roster Consolidation, completed): manager-tdd / manager-ddd retirement decision의 source.
- mo.ai.kr 2026-05-04 21:14:54 incident: predecessor가 1차 evidence base로 사용; 본 SPEC은 동일 incident class의 manager-ddd variant 차단.

---

## 10. BC Migration

본 SPEC은 BREAKING CHANGE 없음 (`breaking: false`, `bc_id: []`).

이유: 모든 변경은 backward-compatible.
- `manager-ddd.md` retired stub 전환: 기존 호출자는 retirement message + migration hint를 받음 (predecessor와 동일 패턴).
- 30 Cat A SUBSTITUTE: documentation references 변경; runtime behavior 변경 없음.
- Cat C UPDATE-WITH-ANNOTATION: documentation table/listing 변경; `factory.go` `ddd_*` action key dispatch는 그대로 유지 (REQ-RD-007).
- audit test 확장: CI assertion 추가; 기존 통과 빌드는 테스트 row가 manager-ddd 표준화 후에야 통과.

마이그레이션 절차: 사용자는 `moai update` 실행만 하면 자동 sync. 별도 user action 불필요.

CHANGELOG entry (proposed):

```markdown
## [Unreleased]

### Bug Fixes / Cleanup

- **SPEC-V3R3-RETIRED-DDD-001**: Standardized `manager-ddd` retirement (predecessor SPEC-V3R3-RETIRED-AGENT-001 pattern replication).
  - Transformed `internal/template/templates/.claude/agents/moai/manager-ddd.md` from full agent definition (7674 bytes) into retired stub (~1.4KB) with canonical 5-field frontmatter (`retired: true`, `retired_replacement: manager-cycle`, `retired_param_hint: "cycle_type=ddd"`, `tools: []`, `skills: []`).
  - Substituted 30 documentation cross-references (`manager-ddd` → `manager-cycle` or `manager-cycle with cycle_type=ddd`).
  - Annotated 2 reference tables (`agent-hooks.md`, `agent-authoring.md`) to mark `manager-ddd` retirement while preserving `factory.go` action-key dispatch.
  - Extended `agent_frontmatter_audit_test.go` with manager-ddd row.
  - User action: run `moai update` to sync the new template files.
```

---

## 11. Traceability (추적성)

- REQ 총 12개: Ubiquitous 4, Event-Driven 3, State-Driven 3, Optional 1, Unwanted 1.
- AC 총 4개 (Positive + Edge + Boundary + Negative); 11/12 REQs mapped (REQ-RD-011 deferred per predecessor REQ-RA-014).
- Predecessor inheritance: RETIRED-AGENT REQ-RA-002 → REQ-RD-001; REQ-RA-007 → REQ-RD-005; REQ-RA-016 → REQ-RD-004 / REQ-RD-012.
- File taxonomy: 30 Cat A + 1 Cat B + 2 Cat C + 3 Ancillary = 36 grand total (within scope §2.1).
- Plan-audit history: iter 1 FAIL 0.78 (8 defects D1-D8) → iter 2 FAIL 0.74 (regression D-NEW-1~5 cross-artifact partial sync) → iter 3 FAIL 0.78 (recovery, MP-1~4 PASS, traceability dim 0.65) → manual textual sync (force-accept) — 4 fixes at spec.md:L170 + plan.md §3 / Risk row / C18 → grep 4-step verification PASS.
- Force-accept justification: must-pass MP-1 (REQ Number Consistency) / MP-2 (EARS Format) / MP-3 (Frontmatter) / MP-4 (Language Neutrality) 4종 모두 PASS. score 0.78은 traceability dimension weighted average; cross-artifact reference manual sync 완료.

---

End of spec.md
