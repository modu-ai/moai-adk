---
id: SPEC-LEAD-DEPUTY-001
title: "리드 세션 직렬 병목 해소 — 상주 deputy 채택 (채택·보고 위임·idle 통지)"
version: "0.1.0"
status: completed
created: 2026-09-03
updated: 2026-09-06
author: manager-spec (card t471)
priority: P1
phase: "v3.1.5 target"
module: ".claude/agents/moai/manager-lead.md, .claude/rules/moai/workflow/kanban-dispatch.md"
lifecycle: spec-anchored
tags: "kanban, factory, manager-lead, deputy, resident-deputy, idle-notification, report-split, template-mirror"
related_specs: [SPEC-LEAD-DEBOTTLENECK-001]
depends_on: [SPEC-LEAD-DEBOTTLENECK-001]
tier: M
---

# SPEC: 리드 세션 직렬 병목 해소 — 상주 deputy 채택

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-03 | manager-spec | 최초 작성 (card t471, 리드 발행 2026-09-03, 운영자 지시). RED-now 기준값 전부 리드 세션 자체 계수 `[리드 자체 계수]` (lead-1 session 2026-09-03 self-count) — acceptance.md §D.0 |

## 1. 문제 — 측정된 형태

### 1.1 병목의 형상은 직렬성(seriality)이지 물량이 아니다

`-k`/`-f` 리드 세션에서 모든 위임 가능한 작업이 리드의 **직렬 턴 루프**를 통과한다. 레인 ~10개가 병렬로 도는 동안 모든 산출물(완료 보고 → 증거 재측정 → 판정 → 디스패치 → 회차 보고)이 리드의 단일 턴 루프에서 대기한다. 관측 증상: 통합 창이 비어 있는데도 준비 완료 레인 여럿이 수 회차 대기 `[리드 자체 계수]` (lead-1 session 2026-09-03, 당일 05:19–05:49 창 관측).

### 1.2 리드 세션 실측 부하 `[리드 자체 계수]` (lead-1 session 2026-09-03 self-count)

아래 수치는 리드 세션의 **자체 계수**다 — 본 SPEC을 실행하는 에이전트가 재유도할 수 있는 측정이 아니며, 인용 시 반드시 `[리드 자체 계수]` 귀속을 동반한다. 재측정 레시피는 acceptance.md AC-LDP-001이 소유한다.

| 부하 지점 | 실측 | 성격 |
|---|---|---|
| 인바운드 레인 완료 보고 | 13건 — 각 건마다 리드가 raw 트리를 직접 재측정 | 리드 턴 점유의 최대 원인 |
| 아웃바운드 디스패치·판정 메시지 | 12건 | 발송마다 리드 턴 1회 |
| 회차 보고 갱신 | 4회차 — 회차당 본문 113–175줄 저작 + 툴 배치 5–6회 | 기계적 저작의 리드 점유 |
| 회차 보고 파일 | 단일 파일 5,550줄 / 458KB — 회차마다 헤더부터 전체 재작성 | 파일 구조 결함 |
| 리드가 spawn한 서브에이전트 | 2건 (Explore — 운영자 지시 조사뿐) | 위임 거의 없음 |
| manager-lead deputy spawn | **0건** | 기구는 있는데 쓰지 않음 |

### 1.3 기구는 이미 있다 — 미사용이 결함이다

`SPEC-LEAD-DEBOTTLENECK-001` (t283, completed, PR #1664)이 이미 착지시킨 것: manager-lead `tools:`의 `SendMessage`/`ListAgents` (`.claude/agents/moai/manager-lead.md:10`), § Deputy dispatch surface (Role B extension, 같은 파일 line 198+)의 위임 5종 vs 리드 보유 6종 매트릭스, `kanban-dispatch.md` § Deputy dispatch surface의 UNNAMED 배경 스폰 패턴 문서. 결함은 기구의 부재가 아니라 **채택의 부재**다 — 리드 세션 시작 시 deputy를 세우는 의무도, 회차 보고를 deputy에 돌리는 규율도, 폴링을 idle 통지로 바꾸는 규율도 어디에도 없다.

### 1.4 축 분해 (3개 독립축) — 분할 판정

| 축 | 내용 | 의존성 |
|---|---|---|
| **A — 상주 deputy** | 배치 시작 시 UNNAMED 배경 `Agent()` manager-lead deputy 1개 spawn. 레인 완료 보고의 raw 트리 판독을 deputy가 수행하고 `RECOMMEND:` 요약만 리드 턴에 도달 | t283 기구 위에 얹는 채택 의무 — 단독 가치 |
| **B — 회차 보고 위임 + 파일 분할** | 측정 배치·표 초안은 deputy가 작성, 리드는 직접 단언할 수치만 재저작. deputy 측정치는 보고 안에 deputy 귀속 명기. 동시에 458KB 단일 파일을 회차별 파일 + 인덱스로 분할 | deputy가 상주해야 위임이 성립 (A 전제) |
| **C — idle 통지** | 반복 `ListAgents` 폴링을 `SendMessage` `notify_when_idle` 1회 요청으로 대체. 경계: idle 통지는 **일정 힌트**지 완료 증거가 아니다 | deputy가 상속하는 감시 임무의 변경 (A의 표면) |

**분할 판정: 단일 SPEC (Tier M), 마일스톤 A→B→C 순.** 세 축이 같은 두 표면을 편집하므로 3분할은 같은 [HARD] 절 블록을 순차 재편집하게 되고, 성공 지표(리드 턴 툴 배치 수)의 관측 창과 기준값을 축마다 재수립해야 한다. A가 먼저고 단독 출하 가능하므로 "Tier L 번들 금지 / A 단독 가치" 요구를 만족한다. B는 A 없이는 t283의 미사용 형태를 재현하고, C는 A가 정의하는 감시 임무를 고친다.

### 1.5 사라질 수 없는 것 [HARD]

이 세 가지를 먼저 못박지 않으면 제안이 과장된다 (카드 본문 지적):

1. **판정은 리드의 것이다.** deputy의 `RECOMMEND:`는 판정이 아니다. 실행자가 자기 산출을 판정하는 구조를 막는 장치는 병목의 원인이자 이 시스템이 보호하는 가치다 — 당일 t395·t461의 자체 계수 실수가 감사에서 잡혔다.
2. **운영자 게이트는 위임 불가.** Kickoff 승인·창 순번·카드 발행은 인간의 결정이다.
3. **증거 재측정 자체는 사라지지 않는다.** 리드가 인용하는 모든 수치는 리드 귀속이어야 한다 (`verification-claim-integrity.md` §2).

줄어드는 것은 리드의 **턴 점유**지, 읽는 횟수가 아니다.

## 2. 용어 — 상주 deputy (resident deputy)

**상주 deputy** = 배치 시작 시 리드 세션이 UNNAMED 배경 `Agent()`로 spawn하고 배치 종료까지 유지하는 manager-lead 인스턴스. t283의 deputy와 같은 역할 확장이며, 달라진 것은 하나다: **기회적(optional)이 아니라 배치 시작 의무**가 된다. UNNAMED spawn 규율(named spawn → in-process teammate 전환으로 결과 반환 두절)과 정지 티메이트 부활 금지, delivery-shape 검증 의무를 t283 사양 그대로 상속한다.

## 3. 요구사항 (GEARS)

### 3.1 축 A — 상주 deputy

**REQ-LDP-001** (Event-driven) **When** a `-k`/`-f` batch begins (first card dispatched to a lane), the lead session shall spawn exactly one UNNAMED background `Agent()` running manager-lead as the resident deputy, before the first lane dispatch of the batch.

**REQ-LDP-002** (State-driven) **While** a lane completion report is pending, the resident deputy shall perform the raw-tree evidence reading for that report and return to the lead only a `RECOMMEND:`-prefixed summary that names the evidence paths it read — the lead's turn receives the summary, not the raw reading batches.

**REQ-LDP-003** (Event-detected) **When** a `SendMessage` send result issued by the resident deputy carries a `routing` object, the deputy shall treat the dispatch as lost and re-send it to the `name [ref]` form (delivery-shape verification, inherited verbatim from `manager-lead.md` § Deputy dispatch surface).

**REQ-LDP-004** (Unwanted) The resident deputy shall not perform any `DEPUTY-RETAINED-BY-LEAD` act — final merge approval, a `FINAL VERDICT:` token, operator gates, queue mutations, CodeRabbit discipline adjudication, or cross-session dispute coordination. A `RECOMMEND:` is not a verdict, and the verdict's home remains the lead.

### 3.2 축 B — 회차 보고 위임 + 파일 분할

**REQ-LDP-005** (State-driven) **While** a round report is being produced, the measurement batch and table drafting shall be delegated to the resident deputy, and the lead shall re-author only the figures it will personally assert — deputy-measured values carry explicit deputy attribution in the report, and lead-asserted values carry lead attribution.

**REQ-LDP-006** (Ubiquitous) The round report shall be maintained as per-round files plus an index file, and each round shall update only its own per-round file and the index — a single monolithic report file rewritten every round shall not be used.

### 3.3 축 C — idle 통지

**REQ-LDP-007** (Event-driven) **When** the lead or the resident deputy needs to learn when a lane session next goes idle, the lead shall request a one-shot `SendMessage` `notify_when_idle` notice instead of issuing repeated `ListAgents` polling rounds.

**REQ-LDP-008** (Unwanted) An idle notice shall not be treated as completion evidence — the notice establishes only when to read the evidence, never success or failure (a session goes idle when it finishes, when it stops at a permission prompt, and when it dies), and advancing a card on the notice alone is prohibited. The boundary clause of `cross-session-messaging.md` § An idle notice is a scheduling hint is inherited by citation, not restated.

### 3.4 성공 지표 — 리드 턴 툴 배치 수 [HARD]

**REQ-LDP-009** (Ubiquitous) The success metric shall be the reduction of the lead session's **tool-batch count** attributable to the three delegable classes (dispatch sends, raw-tree evidence reading, report measurement/drafting) over a fixed observation window, judged by a counted-batch protocol with a stated baseline and target measured from session logs. Turn count is not the metric — it is confounded by operator instruction frequency.

### 3.5 표면·불변식

**REQ-LDP-010** (Capability gate) **Where** a distributed surface is edited (`.claude/agents/moai/manager-lead.md`, `.claude/rules/moai/workflow/kanban-dispatch.md`, `kanban-dispatch-detail.md`), the implementation shall edit the `internal/template/templates/` source first, run `make build` (and `make agents-emit` when the agent definition changes — C2→C3, never hand-edit C3), and keep every template copy free of internal-state content — SPEC IDs, self-counted figures, incident dates, and card ids MUST NOT enter `internal/template/templates/**`; doctrine text stays generic ("the lead", "the deputy") with no project-internal incidents.

**REQ-LDP-011** (Ubiquitous) The deputy adoption shall leave the depth seal intact — manager-lead remains the sole retained agent carrying `Agent` in `tools:`, no Go source under `internal/`, `pkg/`, or `cmd/` is modified, and `internal/template/manager_lead_depth_test.go` passes unchanged.

## 4. 권한 분리 — 변화 없음 확인

t283의 매트릭스(spec.md §4, `manager-lead.md` § Deputy dispatch surface)를 **그대로 상속**한다. 본 SPEC이 추가하는 것은 위임의 **범위**가 아니라 위임의 **발동 시점과 경로**: 배치 시작 의무(REQ-LDP-001), 완료 보고 경로(REQ-LDP-002), 보고 저작 경로(REQ-LDP-005), 감시 방식(REQ-LDP-007). 리드 보유 6종은 한 줄도 늘어나거나 줄어들지 않는다.

## 5. 제약 — PRESERVE

- `manager-lead.md` 기존 [HARD] 절 전부 — 확장 전용, 특히 § Deputy dispatch surface의 위임 5종/보유 6종 표와 delivery-shape 검증 문단 (원문 보존)
- `kanban-dispatch.md` 기존 [HARD] 절 전부 — § Completion is read, never trusted, § Entry into the board is an operator act, § The delegation channel is the queue
- `internal/`, `pkg/`, `cmd/` Go 소스 전체 (REQ-LDP-011) — 런칭 코드는 에이전트 도구를 제어하지 않음 (t283 §5 실측 상속)
- 12-에이전트 카탈로그 수, depth-2 seal, `manager_lead_depth_test.go`
- UNNAMED spawn 규율, queue-on-disk 채널, 동시 write-capable 에이전트 금지
- `cross-session-messaging.md`의 idle-통지 경계 절 — 인용으로 상속, 재서술로 갈라놓지 않음

## 6. Out of Scope

### Out of Scope — 판정·게이트 권한 이관
- 최종 머지 승인, `FINAL VERDICT:`, 운영자 게이트(Kickoff/창 순번/카드 발행), CodeRabbit 규율 판정의 deputy 이관은 설계하지 않는다 — "실행자 자기 판정 방지" 원칙 유지 (§1.5).

### Out of Scope — 새 에이전트·새 세션
- deputy 전용 에이전트 파일, 신규 컴패니언 세션, Agent Teams 정적 계층 도입 없음. deputy는 manager-lead의 역할 확장이며 세션 생성은 운영자 행위다.

### Out of Scope — Go 런칭 코드 변경
- `internal/cli/{kanban,factory,launcher,cc}.go`, `internal/hook/session_start_{kanban,factory}.go`의 수정 없음. Go 변경이 필요한 blocker는 scoped finding으로 후속 카드에 반환.

### Out of Scope — 메시징 채널 재설계
- `SendMessage`을 dispatch의 원천으로 승격하지 않는다 (queue-on-disk 불변). Codex broker(`session_msg_*`) 경로 확장도 범위 밖.

### Out of Scope — 리드 턴 수 감소
- 성공 지표는 툴 배치 수다(REQ-LDP-009). 리드 턴 수는 운영자 지시 빈도에 의해 교락되므로 지표로 쓰지 않는다.

## 7. 교차참조

- `SPEC-LEAD-DEBOTTLENECK-001` (t283, completed) — 본 SPEC이 채택하는 deputy 메커니즘의 착지 SPEC. depends_on: fulfilled (status: completed).
- `.claude/agents/moai/manager-lead.md` § Deputy dispatch surface (Role B extension) — 위임 5종/보유 6종 매트릭스의 SSOT
- `.claude/rules/moai/workflow/kanban-dispatch.md` § Deputy dispatch surface, § Completion is read, never trusted — UNNAMED 배경 deputy 패턴 + 판정 소재
- `.claude/rules/moai/workflow/cross-session-messaging.md` § An idle notice is a scheduling hint — idle 통지 경계의 정본 (인용 상속)
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 1, §2 — 통지만으로 카드 전진 금지 / 리드 귀속 의무
- `.moai/docs/template-internal-isolation-doctrine.md` §25.1 — C1–C8 중립성 카탈로그
- `CLAUDE.local.md` §2.0 — C1/C2/C3 3사본 + `make agents-emit` 규율 / §2.3 — Template-First (로컬 직접 편집은 `moai update`에 소멸)
