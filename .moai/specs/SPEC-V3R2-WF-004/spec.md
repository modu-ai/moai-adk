---
id: SPEC-V3R2-WF-004
title: Agentless Fixed-Pipeline Classification for Utility Subcommands
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: Wave 2 SPEC writer (Layer 6/7/Cleanup)
priority: P1 High
phase: "v3.0.0 — Phase 6 — Multi-Mode Workflow"
module: ".claude/skills/moai/workflows/, .claude/rules/moai/workflow/"
dependencies:
  - SPEC-V3R2-WF-001
related_gap:
  - pattern-library-O-6
  - r1-ai-harness-papers-agentless
related_theme: "Theme 6 — Workflow Consolidation"
breaking: true
bc_id: [BC-V3R2-007]
lifecycle: spec-anchored
tags: "agentless, pipeline, utility, fix, coverage, mx, codemaps, clean, v3"
---

# SPEC-V3R2-WF-004: Agentless Fixed-Pipeline Classification

## HISTORY

| Version | Date       | Author | Description                                                    |
|---------|------------|--------|----------------------------------------------------------------|
| 0.1.0   | 2026-04-23 | Wave 2 | Initial SPEC — classify utility subcommands as fixed pipeline  |

---

## 1. Goal (목적)

R1 §25 Agentless (Xia et al. 2024) 패턴을 MoAI의 utility subcommand 집합에 적용한다. Agentless 파이프라인은 **localize → repair → validate** 3-phase 고정 플로우로, LLM이 control flow를 결정하지 않기에 (a) 비용 예측 가능, (b) 재현성 높음, (c) 무한 루프 리스크 제거. 본 SPEC은 **fix, coverage, mx, codemaps, clean** 다섯 subcommand를 Agentless pipeline으로 공식 분류하고 multi-agent 라우팅을 금지한다. 반면 **plan, run, sync, design**은 개방형 태스크로 기존 multi-agent 오케스트레이션을 유지한다.

### 1.1 배경

pattern-library §O-6: "Three-phase non-agentic pipeline (localize → repair → validate) without LLM-driven control flow. 27.3% SWE-bench Lite, outperforming open-source agentic competitors at lower cost. When to apply: Utility subcommands where task structure is known: `/moai fix`, `/moai coverage`, `/moai codemaps`, `/moai mx`, `/moai clean`. Currently some over-use multi-agent. v3 disposition: ADOPT." R1 §25: "Agentless eliminates LLM-as-orchestrator for well-structured tasks." R6 §2.1-§2.4 audit는 pre/post tool hook에서 MX validation, LSP, metrics가 고정 순서로 실행됨을 보여주며 이들은 이미 준-Agentless다. 본 SPEC은 이 관찰을 규약화한다.

### 1.2 비목표 (Non-Goals)

- Agentless pipeline 구현 세부 (기존 Go handler + hook 재사용)
- `plan`, `run`, `sync`, `design`의 multi-agent 라우팅 변경
- 각 subcommand의 내부 로직 rewrite (분류만 규정)
- 새로운 subcommand 창설
- Agentless pipeline 외부 언어로 이식 (기존 moai-adk-go 범위 내)
- `/moai feedback`, `/moai review`, `/moai e2e`의 classification (3개는 본 SPEC scope 외; 차기 v3.x에서 개별 판정)

---

## 2. Scope (범위)

### 2.1 In Scope

- **Owns**: `fix`, `coverage`, `mx`, `codemaps`, `clean` 5개 subcommand의 pipeline mode 선언.
- Fixed-pipeline 3-phase 계약 정의: localize → repair → validate (또는 analogous 3-phase per subcommand).
- `plan`, `run`, `sync`, `design` 4개 subcommand의 multi-agent 오케스트레이션 유지 선언.
- Subcommand × mode 매트릭스 (SPEC-V3R2-WF-003와 통합): utility 5개는 `--mode` 플래그 무시(pipeline 고정), implementation 4개는 SPEC-V3R2-WF-003의 `--mode` 지원.
- `.claude/rules/moai/workflow/workflow-modes.md` (또는 spec-workflow.md merge 후)에 pipeline subcommand 표 업데이트.

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- `plan`, `run`, `sync`, `design` 의 Agentless 전환 (multi-agent 유지)
- `feedback`, `review`, `e2e` 의 mode classification (추후 결정)
- 신규 pipeline phase 도입 (3-phase 계약 유지)
- subcommand 간 context sharing 신설
- 구현: Agentless Go framework 신설 (기존 구조 재활용)
- Pipeline 실패 시 multi-agent fallback (Agentless는 fail-fast)

---

## 3. Environment (환경)

- 런타임: Claude Code slash command resolver, moai-adk-go CLI
- 영향 디렉터리:
  - 수정: `.claude/rules/moai/workflow/workflow-modes.md` (subcommand classification 표)
  - 수정: `.claude/skills/moai/workflows/{fix,coverage,mx,codemaps,clean}.md` (pipeline header 추가)
  - 참조: `.claude/skills/moai/workflows/{plan,run,sync,design}.md` (multi-agent 표 업데이트)
- 외부 레퍼런스: pattern-library §O-6, R1 §25 Agentless, R6 §2 hooks audit

---

## 4. Assumptions (가정)

- 현재 5개 utility subcommand는 이미 준-Agentless (고정 순서 pre-hook → core → post-hook) 구조를 가진다 (R6 §2.1).
- 3-phase 계약(localize/repair/validate)은 각 subcommand의 자연스러운 해석으로 mapping 가능:
  - `fix`: localize(error) → repair(code) → validate(re-run)
  - `coverage`: localize(untested) → repair(add tests) → validate(re-measure)
  - `mx`: localize(tag drift) → repair(annotate) → validate(re-scan)
  - `codemaps`: localize(stale map) → repair(regenerate) → validate(diff)
  - `clean`: localize(trash) → repair(remove) → validate(tree state)
- SPEC-V3R2-WF-003의 `--mode pipeline` 예약어는 본 SPEC에서 공식화되기 전 오직 문서상 언급만 존재.
- `feedback`, `review`, `e2e`는 본 SPEC scope 외이나 향후 classification 대상으로 명시된다.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-WF004-001**
The following subcommands **shall** be classified as **Agentless fixed-pipeline**: `fix`, `coverage`, `mx`, `codemaps`, `clean`.

**REQ-WF004-002**
The following subcommands **shall** remain **multi-agent**: `plan`, `run`, `sync`, `design`.

**REQ-WF004-003**
Each Agentless subcommand **shall** follow a 3-phase contract: **localize → repair → validate**.

**REQ-WF004-004**
Agentless subcommand execution **shall not** invoke Agent() for control flow decisions; only deterministic tool invocations are permitted.

**REQ-WF004-005**
The classification **shall** be published as a matrix in `.claude/rules/moai/workflow/workflow-modes.md` (or its merge target `spec-workflow.md` per R6 §4.5 recommendation).

### 5.2 Event-Driven Requirements

**REQ-WF004-006**
**When** a user invokes `/moai fix|coverage|mx|codemaps|clean`, the system **shall** execute the 3-phase pipeline without spawning subagents for orchestration.

**REQ-WF004-007**
**When** localize phase finds no targets, the pipeline **shall** exit with status "no-op" and exit code 0, skipping repair and validate.

**REQ-WF004-008**
**When** repair phase encounters an unresolvable error, the pipeline **shall** terminate at that phase and report the error (fail-fast, no multi-agent fallback).

### 5.3 State-Driven Requirements

**REQ-WF004-009**
**While** an Agentless subcommand is running, the SubagentStart hook **shall not** fire from that subcommand's main flow (utility commands use foreground Bash/Read/Write tools only).

**REQ-WF004-010**
**While** `plan` or `run` or `sync` or `design` is executing, SPEC-V3R2-WF-003 `--mode` flag **shall** be honored per WF-003 mode matrix.

### 5.4 Optional Requirements

**REQ-WF004-011**
**Where** a user supplies `--mode` to an Agentless utility subcommand (e.g., `/moai fix --mode team`), the system **shall** ignore the flag and log `MODE_FLAG_IGNORED_FOR_UTILITY`.

**REQ-WF004-012**
**Where** future pipeline enhancements require conditional control flow, the enhancement **shall** be added as a deterministic branch within the 3-phase contract (no LLM dispatcher).

### 5.5 Complex Requirements (Unwanted Behavior / Composite)

**REQ-WF004-013 (Unwanted Behavior)**
**If** an Agentless subcommand implementation starts to spawn Agent() for control flow decisions, **then** a code review check **shall** flag the regression with `AGENTLESS_CONTROL_FLOW_VIOLATION`.

**REQ-WF004-014 (Unwanted Behavior)**
**If** a multi-agent subcommand (e.g., `plan`) is forced into pipeline mode via `--mode pipeline`, **then** the system **shall** emit `MODE_PIPELINE_ONLY_UTILITY` (shared with REQ-WF003-016).

**REQ-WF004-015 (Complex: State + Event)**
**While** an Agentless subcommand is running under `--mode loop` (ignored), **when** the 3-phase completes, the system **shall** not re-enter the pipeline unless the user explicitly re-invokes the command.

---

## 6. Acceptance Criteria (수용 기준 요약)

- **AC-WF004-01**: Given `/moai fix` invocation When executed Then 3-phase pipeline (localize → repair → validate) runs in order without Agent() control flow (maps REQ-WF004-001, REQ-WF004-003, REQ-WF004-004).
- **AC-WF004-02**: Given `/moai coverage` invocation When executed Then untested files are localized, tests are added, coverage is re-measured (maps REQ-WF004-003).
- **AC-WF004-03**: Given `/moai mx` invocation When executed Then tag-drift is localized, annotations are repaired, scan re-runs as validate (maps REQ-WF004-003).
- **AC-WF004-04**: Given `/moai codemaps` invocation When executed Then stale maps are detected, regenerated, and diff-validated (maps REQ-WF004-003).
- **AC-WF004-05**: Given `/moai clean` invocation When executed Then trash is located, removed, tree state validated (maps REQ-WF004-003).
- **AC-WF004-06**: Given the workflow-modes rule file When inspected post-commit Then a matrix lists 5 utility subcommands as "pipeline" and 4 implementation subcommands as "multi-agent" (maps REQ-WF004-005).
- **AC-WF004-07**: Given localize phase finds 0 targets When fix runs Then exit code is 0 and repair/validate are skipped with status "no-op" (maps REQ-WF004-007).
- **AC-WF004-08**: Given repair phase encounters an unresolvable error When fix runs Then the pipeline terminates with the error reported (maps REQ-WF004-008).
- **AC-WF004-09**: Given an Agentless subcommand main flow When SubagentStart hook is observed Then no fire event originates from that subcommand's orchestration (maps REQ-WF004-009).
- **AC-WF004-10**: Given `/moai fix --mode team` When executed Then `--mode` is ignored and `MODE_FLAG_IGNORED_FOR_UTILITY` info log is emitted (maps REQ-WF004-011).
- **AC-WF004-11**: Given `/moai plan --mode pipeline` When executed Then `MODE_PIPELINE_ONLY_UTILITY` error emerges (maps REQ-WF004-014).
- **AC-WF004-12**: Given a PR that adds Agent() spawn in `fix` implementation When code review runs Then `AGENTLESS_CONTROL_FLOW_VIOLATION` flag appears (maps REQ-WF004-013).
- **AC-WF004-13**: Given `/moai run SPEC-001 --mode loop` When executed Then the implementation subcommand honors `--mode` per WF-003 (maps REQ-WF004-010).

---

## 7. Constraints (제약)

- Agentless pipeline은 LLM 호출을 포함하되 **control flow는 고정**이다 (R1 §25).
- 9-direct-dep 정책 준수.
- SPEC-V3R2-WF-003 `--mode` 플래그와의 일관성: utility 5개는 flag 무시, implementation 4개는 flag 적용.
- `.claude/rules/moai/workflow/workflow-modes.md` (또는 merge target)이 단일 진실의 원천(single source of truth).
- `feedback`, `review`, `e2e`는 본 SPEC scope 외 (classification 연기).

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크 | 영향 | 완화 |
|---|---|---|
| Agentless subcommand에 control flow 추가 유혹 | 점진적 agent화 | REQ-WF004-013의 CI 검증 + PR 체크리스트 |
| 3-phase 계약이 특정 subcommand에 잘 맞지 않음 | 강제 매핑 부자연 | 초기 5개만 classify, 타 subcommand는 후속 SPEC에서 판정 |
| 사용자가 `/moai fix --mode team`을 기대 | 혼란 | REQ-WF004-011의 info log + 문서화 |
| Pipeline 실패 시 recover 경로 부재 | UX 저하 | REQ-WF004-008 fail-fast는 의도된 설계; 로그로 진단 경로 제공 |
| workflow-modes.md와 spec-workflow.md 병합 타이밍 | 문서 충돌 | R6 §4.5 merge가 본 SPEC의 publication 위치만 영향 (최종 merge target에 씀) |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3R2-WF-001: 24-skill 카탈로그의 `moai-workflow-loop`, `moai-workflow-testing` 등은 mode와 pipeline classification 참조.

### 9.2 Blocks

- SPEC-V3R2-WF-003 (Multi-mode router): `--mode pipeline`의 대상 subcommand 집합 정의 완료.

### 9.3 Related

- R1 §25 Agentless, pattern-library §O-6.
- R6 §2 hooks audit (MX / LSP / metrics가 이미 고정 순서).

---

## 10. Traceability (추적성)

- REQ 총 15개: Ubiquitous 5, Event-Driven 3, State-Driven 2, Optional 2, Complex 3.
- AC 총 13개, 모든 REQ에 최소 1개 AC 매핑 (100% 커버리지).
- Wave 2 소스 앵커: pattern-library §O-6, R1 §25, R6 §2.
- BC 영향: 없음 (기존 utility subcommand behavior 보존, 분류만 규정).
- 구현 경로 예상:
  - `.claude/rules/moai/workflow/workflow-modes.md` (matrix 업데이트; 향후 spec-workflow.md merge 시 같이 이동)
  - `.claude/skills/moai/workflows/{fix,coverage,mx,codemaps,clean}.md` (pipeline 3-phase header 추가)
  - 코드 review checklist 확장 (AGENTLESS_CONTROL_FLOW_VIOLATION 검증)
- **File:line anchors** (per D5 traceability requirement):
  - `docs/design/major-v3-master.md:L1069` (§11.6 WF-004 definition)
  - `docs/design/major-v3-master.md:L966` (§8 BC-V3R2-007 — Agentless flip)
  - `docs/design/major-v3-master.md:L993` (§9 Phase 6 Multi-Mode Workflow)
  - `.moai/design/v3-redesign/synthesis/pattern-library.md` §O-6
  - `.moai/design/v3-redesign/research/r1-ai-harness-papers.md` (Agentless-OpenDevin reference)

---

End of SPEC.
