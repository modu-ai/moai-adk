---
id: SPEC-RECEIPT-REUSE-001
title: "감사 영수증 인스턴스 간 재사용 차단 — 순차 감사 인스턴스의 영수증 경계"
version: "0.1.0"
status: completed
created: 2026-10-07
updated: 2026-10-07
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: internal/hook
lifecycle: spec-anchored
tier: M
tags: "audit-gate, receipt-reuse, start-marker, subagent-stop, security, t1562, issue-1783"
issue_number: 1783
related_specs: [SPEC-CODEX-AUDIT-GATE-AXES-001]
---

# SPEC-RECEIPT-REUSE-001 — 감사 영수증 인스턴스 간 재사용 차단

카드: **t1562** · 외부 이슈: **#1783** · 카드 클래스: **B (결함 — 원인 코드 특정 완료)** · 실행: **RED 재현 우선 (cycle_type=tdd)** · 동기화 게이트 렌즈: **--security --deep**

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-10-07 | manager-spec | 초안 — 카드 t1562 디스패치. 재사용 경로를 코드에서 검증(spec §1.1), RED 존재론적 기록 3건 + 보존 대상 기저 GREEN 1건 수렴(acceptance.md §B) |

## 1. 문제 정의

### 1.1 관측된 결함 — 검증된 재사용 경로

이 가드는 감사자 PASS가 실제 감사 실행을 증명하게 하는 반-부정 표면이다(SPEC-CODEX-AUDIT-GATE-AXES-001 축 (b)의 훅 절반). 시작 표식이 세션+역할로 키화된 것(카드 t1544, PR #1763)은 배경 스폰의 PASS를 증명 가능하게 만들었지만, **인스턴스 단위 경계는 여전히 열려 있다**. 검증된 경로:

1. 배경 `Agent()` 스폰은 SubagentStart 페이로드에 `agent_id`가 없어 시작 표식이 파생 키 `bg_<session>_<agent_type>`에 기록된다 (internal/hook/audit_receipt_guard.go:103, internal/auditreceipt/store.go:170-178).
2. 파생 키 표식은 **시대 앵커**다 — 같은 세션+역할의 두 번째 인스턴스가 시작할 때 표식이 이미 있으면 조기 반환으로 최초 `StartedAt`(T0)을 그대로 둔다 (internal/hook/audit_receipt_guard.go:111-119; 카드가 지목한 116행: `if _, err := auditreceipt.ReadStartMarker(g.store, key); err == nil {`).
3. 파생 키 표식은 **어떤 인스턴스의 종료로도 지워지지 않는다** — `consumeStartMarker`가 파생 키를 건너뛴다(동시 인스턴스 보호, 의도된 설계) (internal/hook/audit_receipt_guard.go:276-283).
4. `CheckCitedReceipts`의 화폐성 판정은 단일 경계 `r.CreatedAt.After(start.StartedAt)`뿐이다 (internal/auditreceipt/store.go:658) — 앵커 T0 이후에 주조된 영수증이면 그것이 **어느 인스턴스의 수명 동안** 주조됐는지 구분할 수 없다.

**결과 경로**: 인스턴스 1(T0 시작)이 감사를 실행해 영수증 R(T1 주조, T0 < T1)을 받고, R을 인용한 PASS가 수락된다 — 수락이 표식을 지우지 않는다(3번). 같은 세션+역할의 인스턴스 2가 T3에 시작해도 앵커는 T0로 남는다(2번). 인스턴스 2가 감사 도구를 전혀 호출하지 않고 R을 인용하면 R.CreatedAt(T1) > T0이므로 `CheckCitedReceipts`가 승인한다. 즉 **이전 인스턴스의 영수증이 다음 인스턴스의 PASS를 증명한다**. `markerKeys`의 폴백(guard.go:244-267) 때문에 `agent_id`를 실은 stop조차 시대 앵커로 대변될 수 있어 전경/배경 경계를 넘어 같은 경로가 열린다. 수락 없이 끝난 선행 인스턴스(FAIL·거부·판정 줄 부재)도 파생 표식을 남기므로, 재사용은 선행 인스턴스의 **수락 여부와 무관하게** 성립한다.

기존 테스트 `TestSubagentStop_ConcurrentBackgroundAuditorsShareEraAnchor`(internal/hook/audit_receipt_guard_test.go:468-507)가 이 표면을 못박고 있다: 앵커 시작 이후 주조된 r1·r2를 두 인스턴스가 하나씩 인용해 모두 수락됨을 단언한다 — 동시 인스턴스 보호(카드 t1544 card-review P2)를 지키려다 **순차 재사용까지 용인한** 모양이다. 수리는 이 테스트가 못박은 동시성 의미를 깨지 않으면서(REQ-RR-004) 인스턴스 경계를 세워야 한다.

### 1.2 위협 모델 (--security --deep 렌즈)

- 감사자 PASS는 감사자가 쓴 텍스트일 뿐 도구 호출을 증명하지 못한다 — 이 게이트의 존재 이유(SPEC-CODEX-AUDIT-GATE-AXES-001 서두의 실측: "PASS 0.92 세 번, 게이트 강제 통과 review는 FAIL 6건").
- **공격자 이득**: 결함 있거나 불성실한 감사자(또는 같은 세션 이전 인스턴스의 판정 줄이 컨텍스트에 남은 감사자)가 **감사 도구 호출 0회로** 미입증 PASS를 통과시킨다. 영수증 id는 같은 세션 트랜스크립트의 이전 판정 줄에 그대로 노출되므로 공격 비용은 0이다 — 저장소 쓰기도, 권한도, 타이밍 조작도 필요 없다.
- **파급 경로**: 수락된 PASS는 그 역할의 미결 거부 기록을 지운다(`ClearRejectionsForRoleInTreeKind`, guard.go:174) → 거부가 사라진 트리에서 phase-entry 스폰(manager-develop / manager-docs / manager-git) 금지가 풀린다(`checkAuditReceiptSpawn`, guard.go:289-324) → **감사되지 않은 작업이 run / sync / PR로 진행한다**.
- **판정**: 인스턴스 단위 결속이 없으면 "이 감사가 돌았다"는 증명이 "이 세션의 어느 시점에 감사가 한 번 돌았다"로 퇴화한다. 게이트의 명세는 전자다.

### 1.3 미검증 사항 (열린 질문)

- 배경 스폰의 SubagentStop 페이로드가 `agent_id`를 언제 실는지(항상 / 가끔 / 없음)는 런타임 관측이 없다 — 카드 t1544의 테스트는 두 모양(없음: audit_receipt_guard_test.go:443-462, 있음: :539-560)을 모두 모델링한다. RED 경로(양쪽 모두 agent_id 없음)에는 영향이 없으나, 수리 메커니즘이 stop 페이로드의 `agent_id`에 의쟁할 경우 M2 착수 전 관측으로 확정해야 한다 (plan.md §D.3).

## 2. 요구사항 (GEARS)

#### REQ-RR-001 (이벤트 감지 — 결함 본체, 재범위: 단일-생존 선행 종료)
**When** a same-session same-role auditor instance is the only one outstanding as it terminally ends — whether that end is an accepted PASS, a refusal, a FAIL, or an end without a verdict — and a later instance of the same session and role ends with a PASS citing receipts minted before that end, the audit receipt guard shall refuse the PASS instead of accepting it. 종료 시점에 같은 세션·역할의 다른 인스턴스가 살아 있어(겹침 구간) 종료를 특정 인스턴스에 귀속할 수 없으면 인스턴스 경계는 anonymous 이벤트만으로 판정 불가능하다(plan.md §D.3 반례) — 그 재사용 창은 REQ-RR-004의 경계 동결과 함께 명명된 잔여이며, M2 의미는 AC-RR-009가 못박는다.

#### REQ-RR-002 (이벤트 구동 — 오탐 방지)
**When** an auditor instance cites only receipts minted after the receipt boundary applicable to its own generation — no earlier same-session same-role instance's receipts included — the audit receipt guard shall accept the PASS without blocking. 수리가 인스턴스 자신의 영수증을 증명 못하게 해서는 안 된다.

#### REQ-RR-003 (이벤트 감지 — 거부 지속과 진입 차단)
**When** the guard refuses a PASS for receipt reuse, it shall persist the refusal with a cause distinguishing the reuse from the other refusal causes, and the outstanding refusal shall keep the phase-entry spawns (manager-develop / manager-docs / manager-git) denied until a PASS citing a qualifying receipt is recorded.

#### REQ-RR-004 (상태 구동 — 동시성 보존, 재범위: 모호 종료 경계 동결)
**While** multiple same-role auditor instances of one session are live, the audit receipt guard shall keep every live instance able to prove a receipt minted after the applicable boundary, and a terminal end arriving while more than one such instance is outstanding shall not advance that boundary — 수리가 미증명-PASS 교착(카드 t1544 card-review P2)을 재연해서는 안 되며, 동시성 구간의 판정은 수리 전과 등가로 유지된다.

#### REQ-RR-005 (상태 구동 — 첫 정지 연속성 보존)
**While** an auditor instance is blocked at its first stop, the audit receipt guard shall keep that instance able to prove the receipt it mints after continuing — 첫 정지 차단 시 표식을 의도적으로 남기는 현행 동작(guard.go:199-201)은 유지된다.

#### REQ-RR-006 (이벤트 감지 — 개시 이전 재활용 펜스 보존)
**When** an auditor cites a receipt minted before its instance's start boundary, the audit receipt guard shall refuse the PASS with the before-start cause (`CauseReceiptBeforeStart`) — 이 수리로 변하지 않는다.

#### REQ-RR-007 (능력 게이트 — 불개입)
**Where** the audited tree has not made the codex gate `required`, the audit receipt guard shall perform no store writes and emit no refusals — 이 수리로 변하지 않는다.

## 3. 제약

- 수리는 **현행 훅 페이로드 모양**(SubagentStart 무 `agent_id` / SubagentStop 선택적 `agent_id`) 안에서 동작해야 한다 — Claude Code 런타임 변경에 의쟁하지 않는다.
- 기존 거부 원인 어휘(`CauseReceiptBeforeStart` 등)와 거부 종류(kind) 기계(`KindReceipt`/`KindServed`)는 그대로다 — 재사용 원인은 추가다.
- 영수증은 전역 단일 사용(single-use)이 되지 않는다 — MCP `WriteReceipt` 표면 변화는 범위 밖(§ Out of Scope).
- `development_mode: tdd` + 카드 명시: **첫 run-phase 마일스톤은 실패 재현의 RED 관측이 선행**한다 — 명령 + 원문 출력 + exit code + 트리 SHA가 기록되기 전에 어떤 수리 커밋도 착지하지 않는다.
- gitflow 레인 규율: 검증은 영향 패키지(`./internal/hook/... ./internal/auditreceipt/...`) 한정, 로컬 전체 스위트 금지.

## Out of Scope

### Out of Scope — 수리 기제의 내부 설계
- 경계 기록(종료 이벤트 등)의 저장소 레이아웃·레코드 스키마 구체 설계 — run-phase M2의 결정(plan.md §D.3)
- 파생 키 앵커의 만료·청소 주기 재설계

### Out of Scope — 이 결함과 직접 무관한 가드 표면
- FAIL 판정 처리, served-model 거부(`checkServedModelSpawn`), 트리 루트 해상(`TreeRootFromCWD`), 거부 해제 의미론(운영자 결정 K4) — 재사용 수리가 직접 건드리지 않는 한 변경 없음
- 영수증의 전역 단일 사용화 및 MCP 서버(`WriteReceipt`) 표면 변경
- gate가 `required`가 아닌 트리의 동작
- 비-auditor 에이전트(`plan-auditor`/`sync-auditor` 외)의 표식·거부

## 4. 참조

- **SPEC-CODEX-AUDIT-GATE-AXES-001** (completed) — 이 가드의 모본 SPEC, 축 (b)
- **카드 t1544 / PR #1763** — 세션+역할 키화 선행 수리(SPEC 없이 카드로 착지) — `66bb53ebe`
- **카드 t1538** — receipt 게이트 표면의 봉인; 이 결함은 그 이후 잔여
- `.claude/rules/moai/development/verification-completeness.md` §2 — 2칸 채택 규율
- `.claude/rules/moai/core/verification-claim-integrity.md` — 증거 속성 규율
