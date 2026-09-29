---
id: SPEC-WORKFLOW-TASKS-001
title: "카드 수행 레인의 tasks 도구 진행 표시 의무화 (TaskCreate/TaskUpdate 레인 규율)"
version: "1.0.0"
status: in-progress
created: 2026-09-29
updated: 2026-09-29
author: manager-spec
priority: P2
tier: S
phase: "v3.2.x target"
module: ".claude/rules/moai/workflow/kanban-dispatch.md, internal/hook/session_start_factory_i18n.go, internal/hook/session_start_kanban_i18n.go, internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md"
lifecycle: spec-anchored
tags: "kanban, factory, lane-discipline, task-tracking, doctrine"
---

# SPEC-WORKFLOW-TASKS-001 — 카드 수행 tasks 도구 진행 표시 의무화

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-09-29 | 1.0.0 | 초판 작성 — 카드 t1334 (운영자 직접 배차, lane 규율 성문화) |
| 2026-09-29 | 1.0.0 (rev1) | review-1 (PASS 0.91) 결함 정리 — D1: plan.md 골든 테스트 포인터를 실존 파일로 교정 + §E4 셀렉터 확장 / D2: `tier: S` 추가 / D3: 비목표 불릿 이중 대시 수리 / D5: REQ-TASKS-005의 Go 심볼을 §6 참조로 이관 / D6: `module:`에 kanban i18n 표면 추가 / D7(일부): progress.md 모드 선택 헤딩 `§F` 접두어. D4는 REQ/AC 매핑 재번호 없이는 분해 불가하여 복합 REQ 존치(추적성 보존 결정). |

---

## 1. 배경 (Background)

칸반/팩토리 레인 세션이 카드를 받아 수행할 때, 진행 상태를 세션의 tasks 도구(`TaskCreate`/`TaskUpdate`)에 등록·갱신하는 관행이 이미 측정된 모범 사례로 존재한다 — lane-1 카드 t1330이 수행 중 7개 태스크 목록을 `TaskCreate`/`TaskUpdate`로 유지한 것이 그 실측 사례다. 그러나 이 규율은 어디에도 성문화되어 있지 않다.

- 배차 교리(`.claude/rules/moai/workflow/kanban-dispatch.md`)는 레인에게 tasks 도구 사용을 요구하지 않는다.
- 런처 부트스트랩(`moai cc -k` / `moai cc -f` / `moai glm -k` 레인 활성화 시 SessionStart 훅이 주입하는 레인 규칙 텍스트, `internal/hook/session_start_factory_i18n.go`의 `laneNextCardRule`/`laneOwnedCardRule`)에도 같은 요구가 없다.

본 SPEC은 이 관행을 (a) 교리 조항, (b) 런처 부트스트랩 주입 텍스트, (c) 템플릿 미러 3면에 걸쳐 의무 규율로 성문화한다. 훅 기반 기계적 강제는 plan-phase 판단으로 권고만 담는다(§6, plan.md).

## 2. 요구사항 (Requirements — GEARS)

- **REQ-TASKS-001**: When a lane or worker session receives a card, the session shall register that card's execution stages as tasks via `TaskCreate` before beginning the first stage. (카드 인계 시 단계 태스크 등록 의무)
- **REQ-TASKS-002**: When the session completes a stage or moves to the next stage, the session shall update the corresponding task's status via `TaskUpdate` so the task list reflects the current stage at all times. (단계 전이 시 진행 상태 갱신 의무)
- **REQ-TASKS-003**: When the session reports card completion, the session shall report it only while the task list reflects the end state, or When the task list diverges from the end state, the session shall carry an explicit annotation explaining the divergence in its completion report. (완료 보고는 태스크 목록 종결 상태와 일치하거나, 불일치 사유 명시를 전제로 한다)
- **REQ-TASKS-004**: The kanban dispatch doctrine (`.claude/rules/moai/workflow/kanban-dispatch.md`) shall carry the task-registration and status-currency discipline as lane rules, phrased as a rule with its measured example (lane-1 card t1330, 7-task list) rather than as an enumeration of instances. (교리 표면 성문화)
- **REQ-TASKS-005**: The lane bootstrap injection text across all 4 locales — the factory lane rules and the kanban companion lane rule (surfaces and Go symbol pointers: spec.md §6) — shall include the TaskCreate/TaskUpdate discipline alongside the existing next-card/owned-card instructions. (런처 부트스트랩 주입 텍스트 반영)
- **REQ-TASKS-006**: The template mirror copy (`internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md`) shall stay byte-identical in content to the doctrine source; When the doctrine source changes, the mirror shall be regenerated and verified via `make build` parity before the change lands. (템플릿 원본-미러 패리티)

## 3. 인수 기준 (Acceptance Criteria — Tier S inline, Given-When-Then)

- **AC-TASKS-001**: Given a lane session receives a card, When the card intake is acknowledged, Then a task list covering the card's execution stages exists via `TaskCreate` before the first stage's work begins. (maps REQ-TASKS-001)
- **AC-TASKS-002**: Given a registered task list, When the lane completes the plan stage and enters run, Then the plan task's status is `completed` and the run task's status is `in_progress` via `TaskUpdate` at the transition. (maps REQ-TASKS-002)
- **AC-TASKS-003**: Given a lane reports card completion, When the lead (or foreman) reads the completion evidence, Then either the task list shows all tasks terminal, or the completion report carries an explicit divergence annotation naming why the list does not reflect the end state. (maps REQ-TASKS-003)
- **AC-TASKS-004**: Given the doctrine source `kanban-dispatch.md` is edited, When the lead (or foreman) audits the dispatch doctrine, Then the task-discipline rule exists as a rule-shaped clause citing the measured example (t1330) and both `TaskCreate` and `TaskUpdate` appear in it. (maps REQ-TASKS-004)
- **AC-TASKS-005**: Given a freshly launched factory lane in any of the 4 locales (en/ko/ja/zh), When the SessionStart hook injects the lane rule, Then the injected text includes the TaskCreate/TaskUpdate discipline sentence with protocol tokens (`TaskCreate`, `TaskUpdate`) verbatim. (maps REQ-TASKS-005)
- **AC-TASKS-006**: Given the doctrine source `kanban-dispatch.md` is edited, When `make build` runs, Then the mirror at `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` matches the source (parity check passes with zero drift findings). (maps REQ-TASKS-006)

## 4. 제약 (Constraints)

- 교리 파일은 always-loaded 표면이다 — 편집은 캐시 친화적 타이밍(세션 경계 직전 배치, `cache-aware-execution.md` directive 3)으로 수행한다.
- 주입 텍스트의 프로토콜 토큰(`TaskCreate`, `TaskUpdate`)은 모든 로케일에서 번역 없이 verbatim 유지한다(기존 훅 i18n 규약과 동일).
- 본 SPEC은 훅 기반 기계적 강제의 구현을 요구하지 않는다 — 권고는 plan.md §판단 기록에만 둔다(doctrine first, enforcement optional/later).
- 요구사항은 규칙으로 서술하고 사례 열거로 서술하지 않는다(`feedback_dispatch_a_rule_not_an_enumeration_of_its_instances` 교훈).

## 5. 비목표 (Out of Scope)

### Out of Scope — 훅 기반 기계적 강제

- TaskCreate/TaskUpdate 미이용을 감지해 차단하는 PreToolUse/Stop 훅의 구현 — 본 SPEC은 교리·부트스트랩 텍스트만 다루고, 강제는 후속 SPEC 후보로 권고만 수행

### Out of Scope — 칸반 큐 자체의 변경

- `moai todo` 큐 스키마, 카드 상태 모델, `factory_next`/`factory_complete` 프로토콜의 변경 — 기존 카드 라이프사이클은 그대로 유지

### Out of Scope — 리더/오케스트레이터 세션의 tasks 규율

- 리더 세션·본 세션(orchestrator)의 TaskList 사용 규율 — 대상은 카드를 수행하는 레인/워커 세션으로 한정

## 6. 근거 자료 (References)

- 실측 사례: lane-1 카드 t1330 — 수행 중 7개 태스크를 `TaskCreate`/`TaskUpdate`로 유지(본 SPEC이 성문화하는 관행의 관측 근거)
- 배차 교리 SSOT: `.claude/rules/moai/workflow/kanban-dispatch.md` § The dispatch cycle / § Dispatch format
- 레인 부트스트랩 주입: `internal/hook/session_start_factory.go` `factoryLaneRuleForSource`, `internal/hook/session_start_factory_i18n.go` `laneNextCardRule`/`laneOwnedCardRule`, 칸반 동형: `internal/hook/session_start_kanban_i18n.go`
- 템플릿 미러: `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` (make build 재생성)
