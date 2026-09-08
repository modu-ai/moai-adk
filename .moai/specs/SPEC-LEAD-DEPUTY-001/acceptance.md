---
id: SPEC-LEAD-DEPUTY-001
title: "acceptance — 리드 세션 직렬 병목 해소 (상주 deputy 채택)"
version: "0.1.0"
created: 2026-09-03
updated: 2026-09-06
author: manager-spec (card t471)
tier: M
---

# ACCEPTANCE: SPEC-LEAD-DEPUTY-001

## §D.0 RED-now 기준값 (귀속 명시)

**모든 수치는 `[리드 자체 계수]` (lead-1 session 2026-09-03 self-count)다.** 이는 본 SPEC 작성 에이전트가 재유도한 측정이 아니며, 재측정 레시피는 AC-LDP-001이 규정한다. 측정 트리: 리드 세션 산출물 (당일 회차 보고 파일) — 본 워크트리 HEAD가 아니므로 트리 SHA 대신 세션 좌표로 귀속한다.

| 항목 | RED-now 값 | 성격 |
|---|---|---|
| deputy spawn 수 (배치당) | 0 | 기구 존재, 미사용 |
| 인바운드 레인 보고 → 리드 raw 재측정 | 13건, 전부 리드 턴에서 | 위임 0 |
| 아웃바운드 디스패치·판정 발송 | 12건, 전부 리드 턴에서 | 위임 0 |
| 회차 보고 갱신 | 4회차 × (113–175줄 저작 + 툴 배치 5–6회) | 위임 0 |
| 회차 보고 파일 | 단일 파일 5,550줄 / 458KB, 회차마다 전체 재작성 | 파일 구조 결함 |
| 리드 툴 배치 (위임 3계열 합산) | 위의 원천 이벤트 수 = 배치 수의 상한 근사 | AC-LDP-001 baseline |

**GREEN 경로**: M1이 deputy spawn+보고 경로를, M2가 보고 위임+분할을, M3가 idle 통지를 뒤집는다. 각 AC의 green cell은 뒤집을 마일스톤을 명기했다.

## §D AC Matrix (Given-When-Then)

### AC-LDP-001 — 성공 지표: 리드 턴 툴 배치 수 감소 (기계 판정형) [M1-M3]
- **Given** 채택 후 리드 세션 트랜스크립트가 활성 레인 ≥8개 창(관측 창 = 배치 1일, RED-now 창과 레인 구성을 함께 기록)을 덮고,
- **When** 계수 레시피 — "리드 턴의 툴 사용 배치 중 위임 3계열 {SendMessage 디스패치 발송, raw 트리 증거 판독, 보고 측정·초안 저작}에 귀속되는 배치를 세션 로그에서 집계" — 을 실행하면,
- **Then** 집계치가 §D.0 baseline의 **50% 이하**다 (측정 명령·원문 출력·창 구성을 progress.md §E.2에 귀속). 턴 수는 지표가 아니다 (교락 요인 명기).
- RED-now: 위임 3계열 배치 100%가 리드 루프에서 발생 (§D.0). GREEN: M1-M3 착지 후 동일 레시피 재측정.

### AC-LDP-002 — idle 통지는 일정 힌트다 (경계 AC) [M3]
- **Given** 어느 레인의 `notify_when_idle` 통지가 도착하고,
- **When** 리드 또는 deputy가 그 통지를 처리하면,
- **Then** 카드는 통지만으로 전진하지 않고 — progress.md 또는 위임된 증거 경로를 **읽은 뒤에만** 전진하며 — 교리 텍스트가 경계 절("scheduling hint, not completion evidence" 취지)을 `cross-session-messaging.md` 해당 절의 인용으로 운반한다.
- 기계 검증: `grep -c 'scheduling hint' .claude/rules/moai/workflow/kanban-dispatch.md .claude/agents/moai/manager-lead.md` — M3가 `cross-session-messaging.md` § An idle notice is a scheduling hint 상호참조를 양쪽에 새기면 GREEN으로 뒤집는다. RED-now (트리 `109a4615d` 실측): `.claude/rules/moai/workflow/kanban-dispatch.md:0`, `.claude/agents/moai/manager-lead.md:0` — 상호참조가 아직 양쪽 모두 없다.

### AC-LDP-003 — 상주 deputy spawn 의무 [M1]
- **Given** 새 `-k`/`-f` 배치가 시작되면,
- **When** 첫 레인 디스패치가 발생하기 전 세션 로그를 읽으면,
- **Then** UNNAMED 배경 `Agent()` manager-lead deputy spawn 기록이 정확히 1개 존재하고 `name` 파라미터가 없다.
- RED-now: §D.0 — spawn 0. GREEN: M1.

### AC-LDP-004 — RECOMMEND-only 요약 경로 [M1]
- **Given** 레인 완료 보고 1건이 도착하면,
- **When** deputy가 처리를 마치면,
- **Then** 리드 턴에 도달한 것은 읽은 증거 경로를 명명한 `RECOMMEND:` 요약이며, `FINAL VERDICT:` 토큰은 deputy 출력에 존재하지 않는다.
- RED-now: raw 재측정 13건 전부 리드 턴. GREEN: M1.

### AC-LDP-005 — delivery-shape 검증 상속 [M1]
- **Given** deputy의 `SendMessage` 결과에 `routing` 객체가 있으면,
- **When** deputy가 결과를 읽으면,
- **Then** 디스패치를 유실로 판정하고 `name [ref]` 형태로 재발송한다.
- 기존 절 원문 보존 확인: `manager-lead.md` delivery-shape 문단이 편집 전후로 존재.

### AC-LDP-006 — 회차 보고 위임 + deputy 귀속 [M2]
- **Given** 회차 보고를 작성할 때,
- **When** 측정 배치와 표 초안이 만들어지면,
- **Then** 초안은 deputy 소관이고, 보고 안에서 deputy 측정치는 `deputy 측정 (경로)`로, 리드가 직접 단언한 수치는 리드 귀속으로 표기된다 — 무귀속 수치 0건.

### AC-LDP-007 — 회차 보고 파일 분할 [M2]
- **Given** 회차 N이 끝나면,
- **When** 보고 파일을 갱신하면,
- **Then** 회차 N 파일 1개 + 인덱스 1개만 생성·변경되고, 단일 대형 파일의 전체 재작성은 발생하지 않는다 (회차 디렉터리 파일 목록으로 확인: `ls .moai/reports/lead/`에서 회차당 신규 항목 ≤2 — 회차 파일 1 + 인덱스 1. `git diff --stat`은 쓰지 않는다: 회차 파일은 추적되지 않음. RED-now (트리 `109a4615d` 실측): `git ls-files .moai/reports/lead/` = 0행, `ls .moai/reports/lead/` = "No such file or directory").

### AC-LDP-008 — 템플릿 중립성 [M4]
- **Given** `internal/template/templates/**`의 편집이 끝나면,
- **When** `grep -rn "t471\|리드 자체 계수\|SPEC-LEAD-DEPUTY\|SPEC-LEAD-DEBOTTLENECK" internal/template/templates/ | wc -l` 을 실행하면,
- **Then** 출력이 **0**이다. 교리 문구는 "the lead"/"the deputy" 일반 표현이다.

### AC-LDP-009 — depth seal + Go 비접촉 [M4]
- **Given** 모든 편집이 끝나면,
- **When** `go test ./internal/template/ -run 'TestManagerLeadIsSoleAgentCarrier|TestManagerLeadCarriesAgent' -count=1` 과 `git diff --stat internal/ pkg/ cmd/` 을 실행하면,
- **Then** 테스트가 PASS(원문 출력 인용)이고 diff가 비어 있다. `make agents-emit` 후 `.codex` toml이 재생성돼 있다 (`git status` 확인).

### AC-LDP-010 — 위임 불가 3경계 보존 [전체]
- **Given** M1-M3의 모든 편집이 끝나면,
- **When** 편집 전후의 `DEPUTY-RETAINED-BY-LEAD` 6항목과 판정-소재·운영자 게이트 절을 비교하면,
- **Then** 의미 변화가 0건이다 (확장 전용). §1.5의 3경계(판정은 리드 / 운영자 게이트 / 증거 재측정-리드 귀속)가 텍스트로 유지된다.

## §D.1 심각도

- Release-blocking: AC-LDP-001, AC-LDP-002, AC-LDP-003, AC-LDP-008, AC-LDP-009, AC-LDP-010
- Regression-guard: AC-LDP-004, AC-LDP-005, AC-LDP-006, AC-LDP-007 (행위 관찰형 — 세션 로그 좌표 귀속, 단일 명령 재실행형이 아님)

## §D.2 간접 검증

- AC-LDP-001은 채택 후 운용 창이 필요하므로 run-phase에서는 **protocol readiness**(레시피·baseline·target이 교리 텍스트에 존재)만 검증하고, 실제 감소 측정은 첫 채택 배치의 회차 보고에서 수행한다 (2단 검증, 인정된 지연).
- AC-LDP-005/006/007은 run-phase에서는 교리 텍스트 존재로, 채택 후 회차에서 실적으로 확인한다.

## §D.3 Definition of Done

- AC-LDP-001~010 전부 PASS (간접 검증 항목은 protocol-readiness PASS + 채택 후 실적 확인 예정 명기)
- TRUST 5: Trackable(Conventional Commits + t471) / Unified(기존 문서 문체 일치) / Readable(교리 문언 명확) / Secured(내부 상태 템플릿 반입 0) — Tested는 Go 소스 변경 0이므로 기존 테스트 유지 PASS로 충족
- 템플릿 미러 + `make agents-emit` + `make build` 완료, depth-seal 테스트 PASS
