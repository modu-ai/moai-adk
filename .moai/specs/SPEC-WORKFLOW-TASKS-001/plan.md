---
id: SPEC-WORKFLOW-TASKS-001
title: "구현 계획 — 카드 수행 tasks 도구 진행 표시 의무화"
version: "1.0.0"
created: 2026-09-29
updated: 2026-09-29
author: manager-spec
---

# plan.md — SPEC-WORKFLOW-TASKS-001

## A. 컨텍스트

카드 t1334: 레인/워커 세션이 카드를 받으면 `TaskCreate`로 단계 태스크를 등록하고 `TaskUpdate`로 현재 상태를 유지하는 규율을 성문화한다. 실측 근거는 lane-1 카드 t1330의 7-태스크 운영이다. 변경은 텍스트·부트스트랩·미러 3면이며 신규 로직은 없다(Tier S).

리서치에서 확인한 주입 지점:

| 표면 | 경로 | 성격 |
|------|------|------|
| 교리 | `.claude/rules/moai/workflow/kanban-dispatch.md` (§ The dispatch cycle / § Dispatch format 부근) | always-loaded, [MODIFY] |
| 팩토리 레인 규칙 | `internal/hook/session_start_factory_i18n.go` `laneNextCardRule`/`laneOwnedCardRule` — en/ko/ja/zh 4 로케일 + 골든: `session_start_factory_rule_test.go` (`TestSD_AC019_NextCardRuleInjection`, 로케일별 토큰 단정), `session_start_factory_worker_test.go` (`TestFactoryGuideTeachesLaneFormsInEveryLocale`) | Go 소스(미러 없음), [MODIFY] |
| 칸반 동형 레인 규칙 | `internal/hook/session_start_kanban_i18n.go` (컴패니언 레인 안내 텍스트) | Go 소스, [MODIFY] |
| 템플릿 미러 | `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` | make build 재생산물, [MODIFY-파생] |

## B. 기존 이슈

- 교리와 미러는 현재 완전 일치 상태이며, 본 변경은 양면을 같은 커밋에서 움직여야 한다(한 면만 고치는 누수 방지 — `feedback_a_fix_can_land_in_one_mirror_and_miss_the_other`).
- 레인 규칙 텍스트에는 REQ-SD-019 부담이 있다: MCP 도구 토큰은 로케일 불문 verbatim. `TaskCreate`/`TaskUpdate`도 같은 규약으로 취급한다.

## C. 사전 점검 (Pre-flight)

- [ ] `.moai/specs/SPEC-WORKFLOW-TASKS-001/` 하위 외 쓰기 금지 확인 (plan-phase 제약)
- [ ] `moai spec lint .moai/specs/SPEC-WORKFLOW-TASKS-001/spec.md` PASS
- [ ] 훅 i18n 골든 테스트 위치 확인: `internal/hook/session_start_factory_rule_test.go` (`TestSD_AC019_NextCardRuleInjection`), `internal/hook/session_start_factory_worker_test.go` (`TestFactoryGuideTeachesLaneFormsInEveryLocale`), `internal/hook/session_start_kanban_i18n_test.go`

## D. 제약

- always-loaded 파일(`kanban-dispatch.md`) 편집은 세션 경계 직전 배치(cache-aware-execution.md directive 3) — run-phase에서 편집 커밋을 마지막으로 몰아서 수행.
- 요구사항 서술은 규칙형으로, 사례 열거 금지.
- 시간 추정 금지 — 우선순위 라벨만 사용.

## E. 자기 검증

- [ ] `moai spec lint` PASS 출력 인용
- [ ] `grep -c "REQ-TASKS" spec.md` ≥ 6 (목록 마커 행에서 수집됨)
- [ ] `make build` 후 미러 parity 통과 출력 인용
- [ ] 훅 i18n 테스트(`go test ./internal/hook/ -run '^(TestSD_AC019_NextCardRuleInjection|TestFactoryGuideTeachesLaneFormsInEveryLocale|(TestFactory.*Lane|TestKanban.*Lane))$'`) PASS 출력 인용 — `TestSD_AC019_NextCardRuleInjection`을 셀렉터에 명시하는 이유: 레인 규칙 문구가 바뀔 때 갱신해야 하는 골든 단정(로케일별 토큰 핀)이 이 테스트에 있어, 기존 `TestFactory.*Lane`/`TestKanban.*Lane` 패턴으로는 건너뛰어지기 때문이다.

## F. 마일스톤 (우선순위 기반)

### M1 — 교리 조항 + 템플릿 미러 (Priority: High)

- [MODIFY] `.claude/rules/moai/workflow/kanban-dispatch.md`: 레인 규율 조항 추가 — 카드 인계 시 `TaskCreate` 단계 등록, 단계 전이 시 `TaskUpdate` 갱신, 완료 보고는 태스크 목록 종결 상태(또는 불일치 명시 주석) 전제. 실측 사례(t1330, 7-태스크)를 근거로 인용하되 규칙 자체는 사례 열거가 아닌 규칙형으로 서술.
- [MODIFY-파생] `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md`: 원본과 동일 내용 반영.
- 패리티 검증: `make build` 후 미러 diff 0 확인.

### M2 — 런처 부트스트랩 주입 텍스트 (Priority: High)

- [MODIFY] `internal/hook/session_start_factory_i18n.go`: `laneNextCardRule`(claude/glm 레인)과 `laneOwnedCardRule`(gpt 레인)에 tasks 규율 문장 추가 — 4 로케일 전부. 프로토콜 토큰 verbatim 유지.
- [MODIFY] `internal/hook/session_start_kanban_i18n.go`: 칸반 컴패니언 레인 안내 텍스트에 동일 규율 반영(해당 표면에 레인 역할 문구가 존재하는 경우에 한함).
- [MODIFY] 대응 골든 테스트 문자열 갱신 — `session_start_factory_rule_test.go` (`TestSD_AC019_NextCardRuleInjection`, `sdRuleMCPTools`/`sdRuleCLITools`/`sdRuleAuthPin` 토큰 핀), `session_start_factory_worker_test.go` (`TestFactoryGuideTeachesLaneFormsInEveryLocale`), `session_start_kanban_i18n_test.go`.

### M3 — 검증·마감 (Priority: Medium)

- `go test ./internal/hook/` PASS
- `moai spec lint` PASS
- 미러 parity 재확인

### 의존성

M1과 M2는 상호 독립(다른 파일 계열), M3는 양쪽에 의존. 단일 세션 직렬 수행 권장(§ Phase 4 Mode Selection 참조는 progress.md).

## G. 판단 기록 — 훅 기반 기계적 강제 (plan-phase judgment)

**권고: defer (권장하지 않음, 후속 SPEC 후보로 유보).**

- 한 줄 근거: 규율의 실패 모드(태스크 미등록·미갱신)는 완료 증거 읽기(`kanban-dispatch.md` § Completion is read, never trusted)가 이미 2차 방어를 담당하므로, 즉시 차단 훅은 비용 대비 검증 이득이 작다 — 교리+부트스트랩 텍스트로 먼저 관행을 고정하고, 미이용률이 관측되면 그때 감지(비차단) 훅을 후속 SPEC으로.
- 조건부 재검토 트리거: 레인 완료 보고에서 태스크 목록 부재가 반복 관측되면(운영자 판단) 비차단 관측 훅(TaskCreated/TaskCompleted 이벤트 탭 활용)을 후속 카드로 발행.

## H. 안티 패턴 (금지)

- 요구사항을 사례 나열로 서술 ("t1330처럼 하라" X — 규칙 + 근거 인용 O)
- 미러만 고치고 원본을 놓치거나 그 역
- `TaskCreate` 토큰을 한국어 로케일 문장에서 번역·변형
- always-loaded 파일 편집을 세션 중간에 산발 수행

## I. 교차 참조

- REQ-SD-019 (팩토리 레인 규칙 텍스트 규약) — `internal/hook/session_start_factory_i18n.go`
- `cache-aware-execution.md` directive 3 (always-loaded 편집 타이밍)
- `kanban-dispatch.md` § Completion is read, never trusted (2차 방어선)
