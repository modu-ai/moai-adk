---
id: SPEC-V3R2-WF-003
title: Multi-Mode Router (--mode flag, loop/run/design)
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: Wave 2 SPEC writer (Layer 6/7/Cleanup)
priority: P1 High
phase: "v3.0.0 — Phase 6 — Multi-Mode Workflow"
module: ".claude/skills/moai/workflows/, internal/cli/, .claude/commands/moai/"
dependencies:
  - SPEC-V3R2-WF-001
  - SPEC-V3R2-WF-004
related_gap:
  - pattern-library-O-4
  - r2-opensource-tools-omc
related_theme: "Theme 6 — Multi-Mode Execution Styles"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "multi-mode, router, loop, ralph, design, unified, pattern-O-4, workflow, v3"
---

# SPEC-V3R2-WF-003: Multi-Mode Router

## HISTORY

| Version | Date       | Author | Description                                                             |
|---------|------------|--------|-------------------------------------------------------------------------|
| 0.1.0   | 2026-04-23 | Wave 2 | Initial SPEC — unify /moai loop (Ralph) + /moai run + /moai design      |

---

## 1. Goal (목적)

`/moai run`, `/moai loop` (Ralph), `/moai design` 세 subcommand가 본질적으로 "implementation style" 축을 공유하면서도 현재는 별개 명령으로 파편화되어 있다. 본 SPEC은 pattern-library §O-4 (Multi-Mode Router, R2 §3 OMC 6-mode design의 축소판)를 MoAI에 도입하여, `--mode` flag를 통한 execution style 선택을 1차 시민화한다: **autopilot(단일 lead), loop(Ralph 반복), team(병렬 오케스트레이션), pipeline(Agentless 고정 파이프라인)**. 기본 모드는 harness level에 따라 자동 선택된다(minimal→autopilot, standard→autopilot, thorough→team). Subcommand는 유지되며, `--mode`는 subcommand의 실행 방식을 결정한다.

### 1.1 배경

pattern-library §O-4: "Explicit top-level mode surface beyond subcommands: `/team`, `/autopilot`, `/ultrawork`, `/ralph`, `/ralplan`, `omc team`. v3 disposition: CONSIDER — don't adopt all 6 OMC modes; instead add 2-3 execution styles as explicit `--mode` flags on `/moai run` (e.g., `--mode loop` = Ralph, `--mode team` = team orchestration, `--mode autopilot` = single-lead). Default mode auto-selected by harness."
problem-catalog: Ralph 반복(`/moai loop`)과 단일 패스 (`/moai run`)가 별개 command로 분리되어 user mental model에 "같은 SPEC을 다른 command로 다시 실행" 부담을 유발. `/moai design` path A(Claude Design)와 path B(code-based)도 서브커맨드 분기 없이 `--mode` 축에 통합 가능.

### 1.2 비목표 (Non-Goals)

- OMC 6-mode 전체 도입 (축소형 4-mode만 채택)
- Mode 간 context sharing 프로토콜 신설 (각 mode는 독립 실행)
- `/moai run --mode team` 시의 TeamCreate 자동화 확장 (기존 workflow.yaml role_profiles 재사용)
- 신규 subcommand 추가 (`/moai ralph`, `/moai autopilot` 등 OMC 대응 표면 도입 금지)
- Harness level 자동 산정 로직 변경 (기존 complexity estimator 재사용)
- `--mode pipeline` 구현 세부 (SPEC-V3R2-WF-004가 Agentless 분류를 담당, 본 SPEC은 mode 이름만 정의)

---

## 2. Scope (범위)

### 2.1 In Scope

- **Owns**: `--mode {autopilot|loop|team|pipeline}` flag 도입, mode 기본값 자동 선택 로직, subcommand × mode matrix.
- `/moai run`: autopilot (default), loop (Ralph), team (다중 agent 병렬) 지원.
- `/moai loop`: alias for `/moai run --mode loop` (기존 command는 thin wrapper로 전환).
- `/moai design`: autopilot (path B code-based) / import (path A Claude Design handoff, 기존 `moai-workflow-design-import` 재사용) / team (large design brief with parallel copywriting + brand-design).
- Mode 선택 decision tree (harness level → default mode).
- `--mode` 값 검증 (unknown mode → error), subcommand × mode 조합 유효성 검증.
- `/moai plan`, `/moai sync`, `/moai project`, `/moai db`는 **mode 축 비적용** (기본 autopilot 고정).

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- OMC 6개 mode 전체 이식
- Mode 간 마이그레이션 도우미 (`/moai run --mode autopilot` 도중 loop로 전환)
- Team mode 새 role_profile 추가 (기존 workflow.yaml role_profiles 재사용)
- Pipeline mode 구현 세부 (SPEC-V3R2-WF-004 담당)
- `/moai fix`, `/moai coverage`, `/moai mx`, `/moai codemaps`, `/moai clean`는 **mode 축 비적용** (SPEC-V3R2-WF-004에서 pipeline 고정)
- 기존 subcommand의 rename
- `/moai loop`와 `/moai run --mode loop` 간 behavior 차이 허용 (후자가 정식, 전자는 alias)

---

## 3. Environment (환경)

- 런타임: Claude Code slash command resolver, moai-adk-go CLI(Go 1.26+)
- 영향 디렉터리:
  - 수정: `.claude/skills/moai/SKILL.md`, `.claude/skills/moai/workflows/run.md`, `.claude/skills/moai/workflows/loop.md`, `.claude/skills/moai/workflows/design.md`
  - 수정: `.claude/commands/moai/run.md`, `.claude/commands/moai/loop.md`, `.claude/commands/moai/design.md`
  - 확장: mode 선택 decision tree 문서화
- 외부 레퍼런스: pattern-library §O-4, R2 §3 OMC, `.claude/rules/moai/workflow/workflow-modes.md`

---

## 4. Assumptions (가정)

- 기존 `/moai loop` command는 이미 Ralph Engine을 호출하는 thin wrapper이며, alias 전환 시 사용자 경험이 동일하다.
- Harness level(minimal/standard/thorough)은 `.moai/config/sections/harness.yaml`에서 결정되며 mode 자동 선택에 활용된다(MIG-003가 loader 추가 시 runtime 활용 가능).
- `--mode team` 실행 시 workflow.yaml의 `team.enabled: true` + `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` env 프리-조건이 이미 충족되어야 한다 (본 SPEC은 사전 검증만 수행).
- `/moai design` path A는 `moai-workflow-design-import`가 이미 bundle handoff를 파싱하며, path B는 `moai-domain-copywriting` + `moai-domain-brand-design` 파이프라인을 사용한다 (agency FROZEN 유지).
- Mode 기본값 규칙은 user override보다 낮은 우선순위이다: `--mode` 명시 > `.moai/config/sections/workflow.yaml default_mode` > harness 자동.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-WF003-001**
The system **shall** support a `--mode {autopilot|loop|team|pipeline}` flag on `/moai run` and `/moai design`.

**REQ-WF003-002**
The system **shall** define `autopilot` as the default mode when harness level is `minimal` or `standard`.

**REQ-WF003-003**
The system **shall** define `team` as the default mode when harness level is `thorough` AND `workflow.team.enabled: true`.

**REQ-WF003-004**
`/moai loop` **shall** be an alias resolving to `/moai run --mode loop` with identical arguments.

**REQ-WF003-005**
For subcommands {`plan`, `sync`, `project`, `db`}, the system **shall** ignore `--mode` flag and run the default subcommand workflow.

**REQ-WF003-006**
For subcommands {`fix`, `coverage`, `mx`, `codemaps`, `clean`}, the system **shall** operate in pipeline mode per SPEC-V3R2-WF-004 (Agentless fixed pipeline).

**REQ-WF003-007**
The skill documentation **shall** publish a subcommand × mode matrix showing which combinations are valid.

### 5.2 Event-Driven Requirements

**REQ-WF003-008**
**When** a user runs `/moai run SPEC-XXX --mode loop`, the system **shall** invoke the Ralph Engine loop (existing `moai-workflow-loop` skill) with the given SPEC context.

**REQ-WF003-009**
**When** a user runs `/moai design --mode team`, the system **shall** spawn `moai-domain-copywriting` and `moai-domain-brand-design` teammates in parallel per the GAN Loop contract.

**REQ-WF003-010**
**When** an invalid `--mode` value is supplied, the system **shall** reject with `MODE_UNKNOWN` error and suggest the 4 valid values.

**REQ-WF003-011**
**When** `--mode team` is requested but `workflow.team.enabled: false` or `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` is unset, the system **shall** emit `MODE_TEAM_UNAVAILABLE` error and suggest `--mode autopilot` fallback.

### 5.3 State-Driven Requirements

**REQ-WF003-012**
**While** mode auto-selection decides `team` but team prerequisites are not satisfied, the system **shall** automatically downgrade to `autopilot` and print an info-level message explaining the downgrade.

**REQ-WF003-013**
**While** `/moai design` is executing with `--mode import`, the system **shall** skip path B (copywriting + brand-design) and only run `moai-workflow-design-import` for handoff bundle parsing.

### 5.4 Optional Requirements

**REQ-WF003-014**
**Where** the user sets `.moai/config/sections/workflow.yaml: default_mode: <mode>`, the system **shall** use that value in lieu of the harness-based default.

**REQ-WF003-015**
**Where** future v3.x versions extend the mode axis (e.g., adding `ultrawork`), the schema **shall** allow additive extension without breaking existing `--mode` consumers.

### 5.5 Complex Requirements (Unwanted Behavior / Composite)

**REQ-WF003-016 (Unwanted Behavior)**
**If** `--mode pipeline` is specified on `/moai run` (a multi-agent subcommand per SPEC-V3R2-WF-004), **then** the system **shall** reject with `MODE_PIPELINE_ONLY_UTILITY` pointing to the utility subcommand set.

**REQ-WF003-017 (Complex: State + Event)**
**While** `/moai run --mode loop` is executing, **when** the Ralph loop reaches convergence, the system **shall** terminate the loop and report the final iteration's output (not continue infinitely).

**REQ-WF003-018 (Unwanted Behavior)**
**If** two conflicting mode selection sources disagree (e.g., `--mode autopilot` + `default_mode: team` in workflow.yaml), **then** the CLI-provided `--mode` **shall** win (precedence: CLI > config > harness auto).

---

## 6. Acceptance Criteria (수용 기준 요약)

- **AC-WF003-01**: Given `/moai run SPEC-001 --mode autopilot` When executed Then single lead agent executes the SPEC (maps REQ-WF003-001, REQ-WF003-002).
- **AC-WF003-02**: Given `/moai run SPEC-001 --mode loop` When executed Then `moai-workflow-loop` (Ralph) is invoked (maps REQ-WF003-008).
- **AC-WF003-03**: Given `/moai loop SPEC-001` When executed Then it resolves to `/moai run --mode loop SPEC-001` with identical behavior (maps REQ-WF003-004).
- **AC-WF003-04**: Given harness level = `thorough` AND `team.enabled: true` When `/moai run SPEC-001` is executed without `--mode` Then `team` mode is auto-selected (maps REQ-WF003-003).
- **AC-WF003-05**: Given `/moai design --mode team` When executed Then `moai-domain-copywriting` + `moai-domain-brand-design` teammates run in parallel per GAN Loop (maps REQ-WF003-009).
- **AC-WF003-06**: Given `/moai run --mode banana` When executed Then `MODE_UNKNOWN` error with 4 valid values is emitted (maps REQ-WF003-010).
- **AC-WF003-07**: Given `--mode team` requested but `team.enabled: false` When executed Then `MODE_TEAM_UNAVAILABLE` error suggests `--mode autopilot` (maps REQ-WF003-011).
- **AC-WF003-08**: Given harness auto-select = team but prerequisites not met When executed Then downgrade to autopilot with info log (maps REQ-WF003-012).
- **AC-WF003-09**: Given `/moai plan --mode loop` When executed Then `--mode` is ignored and plan runs normally (maps REQ-WF003-005).
- **AC-WF003-10**: Given `/moai fix --mode autopilot` When executed Then fix operates in pipeline mode per SPEC-V3R2-WF-004 (not autopilot) (maps REQ-WF003-006, REQ-WF003-016).
- **AC-WF003-11**: Given `/moai run --mode pipeline` When executed Then `MODE_PIPELINE_ONLY_UTILITY` error is emitted (maps REQ-WF003-016).
- **AC-WF003-12**: Given `workflow.yaml: default_mode: loop` When `/moai run` without `--mode` is executed Then `loop` mode activates (maps REQ-WF003-014).
- **AC-WF003-13**: Given CLI `--mode autopilot` AND `workflow.yaml default_mode: team` When executed Then CLI wins and autopilot runs (maps REQ-WF003-018).
- **AC-WF003-14**: Given Ralph loop reaches convergence When loop ends Then final iteration output is reported and loop terminates (maps REQ-WF003-017).
- **AC-WF003-15**: Given `/moai design --mode import` When executed Then only design-import runs, skipping copy + brand (maps REQ-WF003-013).

---

## 7. Constraints (제약)

- SPEC-V3R2-WF-001 24-skill 카탈로그 준수: 신규 skill 추가 금지 (기존 moai-workflow-loop 재사용).
- `.claude/rules/moai/workflow/workflow-modes.md` (향후 spec-workflow.md와 merge per R6 §4.5)와의 일관성 유지.
- 9-direct-dep 정책 준수.
- Mode precedence: CLI `--mode` > config > harness auto (REQ-WF003-018).
- Team mode는 `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` 프리조건 없이 activation되지 않는다.
- `/moai loop` wrapper는 20 LOC thin-wrapper 규격 유지.

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크 | 영향 | 완화 |
|---|---|---|
| 사용자가 `--mode` 명시 없이 기대와 다른 mode 자동 선택됨 | UX 혼란 | 기본 mode 선택 이유를 info log로 표시 + `--mode` 명시적 사용 권장 |
| `/moai loop` alias와 `/moai run --mode loop` 간 edge-case 불일치 | 회귀 | alias는 단순 rewrite, behavior 동일성 CI 테스트 |
| Team mode 프리조건 누락으로 silent fallback | 사용자 의도 파괴 | REQ-WF003-012의 info log 필수 + fallback 명시 |
| `--mode pipeline`을 multi-agent subcommand에 허용 요청 | 의미 파괴 | REQ-WF003-016의 hard error로 거부 |
| OMC 6-mode 전체 채택 압력 | scope creep | Non-Goal에 명시, 4-mode 한정 |
| harness.yaml 부재 시 mode 기본값 결정 실패 | 런타임 오류 | SPEC-V3R2-MIG-003의 loader 도입 전까지 hardcoded default(autopilot) 사용 |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3R2-WF-001: 24-skill 카탈로그의 `moai-workflow-loop`, `moai-workflow-design-import`, `moai-domain-copywriting`, `moai-domain-brand-design`이 존속해야 mode 라우팅 가능.
- SPEC-V3R2-WF-004: pipeline mode의 대상 subcommand 집합 정의.

### 9.2 Blocks

- SPEC-V3R2-MIG-003: harness.yaml 로더가 mode auto-select 근거.

### 9.3 Related

- SPEC-V3R2-WF-002 (Commands thin-wrapper): `/moai run`, `/moai design` command wrapper 레이어 동일.
- pattern-library §O-4, §O-6 (Agentless).

---

## 10. Traceability (추적성)

- REQ 총 18개: Ubiquitous 7, Event-Driven 4, State-Driven 2, Optional 2, Complex 3.
- AC 총 15개, 모든 REQ에 최소 1개 AC 매핑 (100% 커버리지).
- Wave 2 소스 앵커: pattern-library §O-4, R2 §3 OMC 6-mode, `.claude/rules/moai/workflow/workflow-modes.md`.
- BC 영향: 없음 (additive flag, 기존 command behavior 보존).
- 구현 경로 예상:
  - `.claude/skills/moai/SKILL.md` (router 문서 업데이트)
  - `.claude/skills/moai/workflows/run.md` (mode dispatch)
  - `.claude/skills/moai/workflows/design.md` (path A/B + team mode 통합)
  - `.claude/commands/moai/loop.md` (thin alias wrapper 업데이트)
  - `.moai/config/sections/workflow.yaml` schema 확장(optional `default_mode`)

---

End of SPEC.
