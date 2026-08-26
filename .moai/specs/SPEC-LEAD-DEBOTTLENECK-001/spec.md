---
id: SPEC-LEAD-DEBOTTLENECK-001
title: "리드 병목 해소 — manager-lead 백그라운드 병렬 위임 (deputy dispatch surface)"
version: "0.1.0"
status: in-progress
created: 2026-08-26
updated: 2026-08-26
author: manager-spec (card t283)
priority: P1
phase: "v3.1.4 target"
module: ".claude/agents/moai/manager-lead.md, .claude/rules/moai/workflow/kanban-dispatch.md"
lifecycle: spec-anchored
tags: "kanban, factory, manager-lead, deputy, cross-session-messaging, template-mirror"
related_specs: [SPEC-TEAMMATE-REVIVAL-SOLE-WRITER-001]
tier: M
---

# SPEC: 리드 병목 해소 — manager-lead 백그라운드 병렬 위임

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-08-26 | manager-spec | 최초 작성 (card t283, 운영자 지시 2026-08-26 우선순위 1위). RED-now 기준값 전부 2026-08-26 t283 워크트리(HEAD `175d63f3f`)에서 실측 — acceptance.md §D.0 |

## 1. 문제 — 측정된 형태

### 1.1 병목 메커니즘

`-k`/`-f` 리드 세션의 조율 작업은 전부 manager-lead 서브에이전트를 경유하지만, 그 서브에이전트의 `tools:` 허용목록에 `SendMessage`/`ListAgents`가 없다 (`.claude/agents/moai/manager-lead.md:10`, 실측). 세션 간 메시지는 리드 세션 자신의 턴 루프를 통해서만 나갈 수 있고, manager-lead가 초안을 쓰면 리드 세션이 발송한다 — 직렬. 역설적으로 agent 본문(line ~36 "Lead-session posture")은 이미 "handles cross-session messaging to companions and lanes"라는 목표 상태를 선언하지만, 도구가 없어 실행할 수 없다. 런타임 제한이 아니다: 백그라운드 서브에이전트 스키마에 `SendMessage`가 존재함은 이미 실측되어 기록되어 있다 (`.claude/rules/moai/workflow/orchestration-mode-selection.md:124` — "SendMessage present"). 누락은 허용목록 선택이며, 본 SPEC이 뒤집는 대상이다.

### 1.2 리드 세션 실측 병목 (2026-08-25~26, 카드 t283 조사 전수)

| 병목 지점 | 실측 | 결과 |
|---|---|---|
| 디스패치 발송 직렬화 | 8건 | 발송마다 리드 턴 점유 |
| CI watch 점유 | 6건 | 대기 중 리드가 다른 판정 불가 (#1648 3라운드) |
| CodeRabbit 슬롯 폴링 | 5회 | #1651 슬롯 대기 중 조율 정지 |
| 레인 응답 처리 지연 | 12건 | 판정 대기 큐 적체 |

### 1.3 설계 방향

레인 지시·응답 처리를 리드 세션의 직렬 턴에서 **manager-lead 백그라운드 병렬(deputy)로 이관**한다. 단, 판정 권한 위임은 "실행자 자기 판정 방지" 원칙(`kanban-dispatch.md` § Completion is read, never trusted / § The verdict's home)과 충돌하지 않아야 한다: manager-lead는 실행자가 아니므로 **1차 판정(증거 읽기 + 권고)은 위임 가능**하나, **최종 머지 승인은 리드 게이트가 유지**된다. Go 계층(`internal/cli/**`) 변경은 불필요하다 — deputy는 에이전트 정의 + 교리(doctrine) 계층 구성물이며, 런칭 코드(`internal/cli/{kanban,factory,launcher}.go`, `internal/hook/session_start_{kanban,factory}.go`)는 에이전트 도구 허용목록을 제어하지 않는다 (조사 실측, §5 제약).

## 2. 용어 — deputy

**deputy** = 리드 세션이 UNNAMED 백그라운드 `Agent()`로 spawn한 manager-lead 인스턴스 중 디스패치·감시·1차 검증을 수행하는 역할. 신규 에이전트가 아니라 manager-lead의 역할 확장이며(12-에이전트 카탈로그 불변), `SendMessage`/`ListAgents` 도구 추가로 리드 세션의 턴을 점유하지 않고 레인·컴패니언 세션과 직접 메시지를 주고받는다. GLM 위험(UNNAMED spawn 규율, named spawn의 in-process teammate 전환)과 정지 티메이트 부활 위험(SPEC-TEAMMATE-REVIVAL-SOLE-WRITER-001)을 상속받는다.

## 3. 요구사항 (GEARS)

### 3.1 에이전트 정의 계층

**REQ-LDB-001** (Ubiquitous) The manager-lead agent definition shall include `SendMessage` and `ListAgents` in its `tools:` allowlist.

**REQ-LDB-002** (Ubiquitous) The manager-lead agent body shall carry a "Deputy dispatch surface" section that codifies the §4 permission matrix — delegable duties, retained powers, and the delivery-shape verification protocol.

**REQ-LDB-003** (Unwanted) The deputy shall not perform any of the LEAD-RETAINED acts: final merge approval, operator gates (the orchestrator-exclusive user-question tool), queue mutations (`moai todo` add/pick/done), final PASS/FAIL verdicts, CodeRabbit discipline adjudication (the deputy reads and reports the two-condition result; it does not decide slot-wait outcomes), or cross-session dispute resolution. The prohibition shall be expressed as a grep-able `DEPUTY-RETAINED-BY-LEAD` marker in the agent body.

**REQ-LDB-004** (Event-driven) **When** the deputy sends a dispatch address block via `SendMessage`, the deputy shall read the send result and treat a `routing` object on the result as a lost delivery, re-sending to the `name [ref]` form per `kanban-dispatch-detail.md` § The dispatch cycle.

**REQ-LDB-005** (State-driven) **While** a CI/CodeRabbit watch or a first-pass evidence read is delegated to the deputy, the deputy shall return terminal states, the two-condition CodeRabbit read (combined-status description `Review completed` + `Merge Risk:` line matching headRefOid), and a recommendation — as a report to the lead, never as an adjudication.

**REQ-LDB-006** (Event-detected) **When** the deputy observes a forbidden-act request in its delegation (a merge, a final verdict, a queue mutation), the deputy shall refuse it and return a blocker report naming the `DEPUTY-RETAINED-BY-LEAD` clause violated.

### 3.2 교리 계층

**REQ-LDB-007** (Ubiquitous) The kanban-dispatch rule set (`.claude/rules/moai/workflow/kanban-dispatch.md` stub + `kanban-dispatch-detail.md` companion) shall gain the deputy surface as an **extension**: `kanban-dispatch-detail.md` § The lead works through manager-lead gains the deputy mode, and the stub carries the deputy's [HARD] boundary clauses — with every existing [HARD] clause preserved verbatim (extension, not rewrite).

**REQ-LDB-008** (State-driven) **While** the deputy holds delegated dispatch duty, the queue on disk shall remain the delegation channel — the deputy's `SendMessage` is a nudge, never the delegation itself, and card advancement continues to require evidence the lead read (`kanban-dispatch.md` § The delegation channel is the queue, § Completion is read, never trusted).

**REQ-LDB-009** (Capability gate) **Where** a distributed surface is edited (`.claude/agents/moai/manager-lead.md`, `.claude/rules/moai/workflow/kanban-dispatch*.md`), the implementation shall carry the `internal/template/templates/` mirror, a `make build` regeneration, and template-neutrality compliance — the template copies carry no SPEC IDs, REQ tokens, audit citations, or internal dates (C1–C8 catalogue, `.moai/docs/template-internal-isolation-doctrine.md` §25.1).

### 3.3 불변식

**REQ-LDB-010** (Ubiquitous) The tools addition shall leave the depth seal intact: manager-lead remains the sole retained agent carrying `Agent` in `tools:`, and `internal/template/manager_lead_depth_test.go` continues to pass. The messaging tools are additive and MUST NOT introduce a second `Agent` carrier or a nesting path.

**REQ-LDB-011** (Unwanted) The implementation shall not modify any Go source under `internal/`, `pkg/`, or `cmd/` — the deputy is an agent-layer + doctrine-layer construct. A discovered blocker that requires a Go change shall be returned as a scoped finding for a follow-up card, not absorbed silently.

**REQ-LDB-012** (State-driven) **While** the deputy runs, no second write-capable agent shall run concurrently with it — the deputy's write surface is limited to report files, and its lane-facing activity is messaging + read.

## 4. 권한 분리 매트릭스 (초안 — plan 심사 + 운영자 kickoff 확인 대상)

| 조율 작업 | deputy로 위임 | 리드 세션 유지 | 근거 |
|---|---|---|---|
| 디스패치 주소블록 발송 (+ 발송 형상 검증) | ✓ | — | 리드 턴 점유 8건 중 최대 빈도 |
| CI watch 배치 (종착 상태 수집) | ✓ | — | 점유 6건, #1648 3라운드 |
| CodeRabbit 2조건 **판독·보고** | ✓ | — | 폴링 5회, #1651 |
| CodeRabbit 규율 **판정** (슬롯 대기 결론) | — | ✓ | 판정 권한은 리드 |
| 1차 증거 읽기 + 권고 | ✓ | — | manager-lead는 실행자 아님 → 자기판정 아님 |
| 최종 PASS/FAIL 판정 | — | ✓ | `kanban-dispatch.md` § The verdict's home |
| 최종 머지 승인 | — | ✓ | 카드 지시: 리드 게이트 유지 권장 |
| 운영자 게이트 (AskUserQuestion) | — | ✓ | 이미 orchestrator 전유 (agent NOT-for) |
| 카드 발행 / `done` (`moai todo` 변이) | — | ✓ | 입장·종료는 운영자·리드 행위 |
| 크로스세션 분쟁 조율 | — | ✓ | 사실 전달은 위임, 결정은 리드 |
| 요약 보고 (리드 세션 반환) | ✓ | — | deputy의 산출물 형태 |

## 5. 제약 — PRESERVE

- `internal/cli/**`, `internal/hook/**` 등 Go 전체 (REQ-LDB-011)
- 12-에이전트 카탈로그 수 (신규 에이전트 없음 — manager-lead **확장**)
- `kanban-dispatch.md` 기존 [HARD] 규칙 전부 (확장 전용)
- `CLAUDE.md`, 타 에이전트 파일 전부
- UNNAMED spawn 규율, queue-on-disk 채널, 동시 write-capable 에이전트 금지
- `manager_lead_depth_test.go` CI 가드 (baseline: `go test ./internal/template/ -run 'TestManagerLeadIsSoleAgentCarrier|TestManagerLeadCarriesAgent|TestNoNestedLeafWorkerCarrier' -count=1` → `ok github.com/modu-ai/moai-adk/internal/template 3.305s`, 2026-08-26 t283 워크트리)

## 6. Out of Scope

### Out of Scope — Go 런칭 코드 변경
- `internal/cli/{kanban,factory,launcher,cc}.go` 및 `internal/hook/session_start_{kanban,factory}.go`의 어떤 수정도 하지 않는다 (deputy는 에이전트+교리 계층 구성물). Go 변경이 필요한 blocker 발견 시 scoped finding으로 반환.

### Out of Scope — 판정 권한 이관
- 최종 머지 승인, 최종 PASS/FAIL, 운영자 게이트, 카드 발행/done의 deputy 위임은 설계하지 않는다 — "실행자 자기 판정 방지" 및 "판정의 소재" 원칙 유지.

### Out of Scope — 신규 에이전트/신규 세션
- deputy 전용 에이전트 파일, 신규 컴패니언 세션, Agent Teams 정적 계층 도입 없음. 세션 생성은 여전히 운영자 행위.

### Out of Scope — 메시징 채널 재설계
- 리드-레인 메시지를 dispatch의 원천으로 승격하지 않는다 (queue-on-disk 불변). Codex broker(`session_msg_*`) 경로 확장도 본 카드 범위 밖.

## 7. 교차참조

- `SPEC-TEAMMATE-REVIVAL-SOLE-WRITER-001` (t269, completed) — 정지 티메이트 SendMessage 부활 금지 + 감사 중 단독-작성자 규율. deputy 메시징 위험과 상호작용: deputy는 정지된 세션에 메시지를 보내 부활시켜서는 안 된다 (해당 SPEC의 교리를 상속받아 agent 본문에 인용).
- `.claude/rules/moai/workflow/kanban-dispatch.md` + `kanban-dispatch-detail.md` — 확장 대상 교리의 SSOT
- `.claude/rules/moai/workflow/cross-session-messaging.md` — role-boundary dispatch 3조건 (선언된 위상 ✓ / SSOT 포인터 ✓ / 격리 트리 ✓ — deputy dispatch는 3조건 충족 형태로 설계됨)
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` §C.1 — 백그라운드 서브에이전트 SendMessage 존재 실측 기록
- `.moai/docs/template-internal-isolation-doctrine.md` §25.1 — C1–C8 중립성 카탈로그
