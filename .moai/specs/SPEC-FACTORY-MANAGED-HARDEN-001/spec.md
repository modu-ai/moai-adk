---
id: SPEC-FACTORY-MANAGED-HARDEN-001
title: "Factory 관리 세션 소유자 강건화 — 서버 요청 응답과 턴 단위 실패 격리 (F3·F4, SPEC-FACTORY-MANAGED-SESSION-001 후속)"
version: "0.5.1"
status: in-progress
created: 2026-10-03
updated: 2026-10-03
author: GOOS (manager-spec)
priority: P1
phase: "v3.2.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tier: M
related_specs: [SPEC-FACTORY-MANAGED-SESSION-001]
tags: "factory, managed-session, codex, app-server, hardening, failure-isolation"
---

# SPEC-FACTORY-MANAGED-HARDEN-001 — Factory 관리 세션 소유자 강건화 (F3·F4)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.5.1 | 2026-10-03 | manager-spec | 축소 범위 plan-audit 2차(FAIL 0.78, 감사 커밋 `d43a52dac`) 개정 — 운영자가 감사 상한을 한 번 더 열었고 범위는 N1·N2 로 한정됐다. `acceptance.md` 만 고쳤다: AC-MH-007 의 `below_ceiling_continues` 시나리오 정의와 mu13 지목, mu18 지목 축소, #11 채택용 변이 mu19·mu20, 채택 회계 문장 정정. 변이 18→20개. REQ·AC 개수(12·13)와 요구 문면은 불변. |
| 0.5.0 | 2026-10-03 | manager-spec | 축소 범위 plan-audit 1차(FAIL 0.75, 감사 커밋 `bca1e0629`) 개정. 검증층(오라클·변이표·시험 규약)만 고쳤다: 변이 17→18개와 변이 미채택 불변 가드 G1–G8 분리, 로그 관찰 규약(로그 선기록·폴링·단일 sink), `-race` 안전 주장 정정, 재현 시험 5→6개(`TestManagedCodexNonCompletedTurnIsolated`). REQ 12·AC 13 불변. REQ-MH-006/008 의 관계 문구("until REQ-MH-008 applies", "notwithstanding REQ-MH-006")와 귀속 상세의 design.md 이관만 요구 본문을 건드렸다. |
| 0.4.0 | 2026-10-03 | manager-spec | **범위 분할 개정.** 운영자 결정(리더 중계, 2026-10-03)으로 F5(시그널 처리와 `Start`/`Close` 수명주기)가 새 카드 t1459 로 나가고 이 SPEC은 F3·F4 로 줄었다. F5 요구·AC·변이·설계(구 REQ-MH-010·011·012·013, 구 AC-MH-009·010·011·016, 수명 순서표)를 모두 제거하고 REQ·AC 를 연속 번호로 다시 매겼다(REQ 16→12, AC 16→13). plan-audit 3차(FAIL 0.81, `951f2bfb6`)의 N12(귀속 규칙 구멍)·N10(변이 m14·m17·m19)·N15(이음새)를 이 범위 안에서 해소했다. 이전 텍스트는 git 이력(`95dfd85c8`, `820eff47f`, `951f2bfb6`)에서만 볼 수 있다. |
| 0.3.0 | 2026-10-03 | manager-spec | plan-audit 2차(FAIL 0.81, 감사 커밋 `820eff47f`) 개정(F5 포함 범위). |
| 0.2.0 | 2026-10-03 | manager-spec | plan-audit 1차(FAIL 0.75, 감사 커밋 `95dfd85c8`) 개정(F5 포함 범위). |
| 0.1.0 | 2026-10-03 | manager-spec | Initial SPEC (card t1409). 완료된 SPEC-FACTORY-MANAGED-SESSION-001의 독립 sync 감사 지적 F3·F4·F5를 닫는 후속 SPEC이었다. 부모 SPEC은 본문·frontmatter 모두 수정하지 않는다. |

## §A. 개요

부모 SPEC-FACTORY-MANAGED-SESSION-001(상태 `completed`)이 만든 관리 세션 계층의 Codex/Claude 소유자에는 독립 sync 감사(`.moai/reports/t1375/sync-audit.md`, 원문 인용: `.moai/reports/t1409/t1375-f3-f5-verbatim.md`)가 남긴 공백이 있다. 운영자는 2026-10-02 결정으로 F3·F4·F5 를 후속 카드 t1409 로 닫기로 했고, 2026-10-03 에 **F5 를 새 카드 t1459 로 분리**했다(리더 중계). 이 SPEC은 F3·F4 만 다룬다.

| 지적 | 현상 | 관측 등급 |
|---|---|---|
| F3 | 서버가 먼저 보내는 요청(`id`와 `method`를 함께 가진 프레임)에 응답하지 않는다. | 기준 트리 `7109e0900` 에서 임시 테스트로 **실측**했다(테스트는 커밋하지 않고 삭제함): (1) 서버가 보낸 요청 id=7에 클라이언트가 1초 안에 어떤 프레임도 돌려보내지 않았다. (2) 서버 요청 id=1이 대기 중인 클라이언트 id=1과 같을 때 `call()` 이 서버 요청을 응답으로 오인해 빈 결과를 오류 없이 돌려주고, 진짜 응답 본문은 버려졌다. (3) 서버 요청 id가 문자열(`"srv-1"`)이면 읽기 루프가 디코딩 오류로 끝나 연결이 닫힌 것으로 처리됐다. |
| F4 | 턴 하나가 실패하면 `driveManagedFactorySession` 이 반환해 세션 전체가 끝난다. | 소스 판독(`internal/cli/managed_factory_session.go:345`)과 기존 테스트 `TestManagedDriverFailureBranches`. 분류 기준은 SPEC에 없었다(감사서 §F4). |

보조 사실(전부 기준 트리 `7109e0900` 에서 관측):

- codex 0.160.0 `app-server generate-json-schema` 로 서버→클라이언트 요청이 10종(`item/commandExecution/requestApproval`, `item/fileChange/requestApproval`, `item/permissions/requestApproval`, `mcpServer/elicitation/request`, `item/tool/requestUserInput`, `item/tool/call`, `account/chatgptAuthTokens/refresh`, `attestation/generate`, `applyPatchApproval`, `execCommandApproval`)이고, 요청 `id` 는 문자열 또는 int64라는 것을 확인했다(`RequestId` = `anyOf [string, int64]`). 현재 클라이언트의 프레임 구조체는 `id` 를 `int` 로 선언한다.
- 브로커 `internal/factorymsg/store.go` 의 `messages` 테이블에는 시도 횟수 열이 없고(`:389`), 클라이언트 쪽 API는 `Claim`·`ReadBody`·`RecordDisposition`·`Receipt` 뿐이라 claim을 되돌리는 호출이 없다. `Claim` 은 `state='claimed'` 이고 `claim_expires_at` 이 지난 행을 다시 후보로 삼아 새 claim token으로 재배달한다(`:878`, `:901`, `:916`; 기존 테스트 `TestDispatchResultExactlyOnce/lost_receipt_redelivery` 가 기준 트리에서 PASS). 메시지 수명은 TTL(`MaxTTL` 7일, `:40`)이고 TTL이 지난 행은 `Claim` 이 dead-letter로 보낸다(`:892-898`).
- **미관측 전제**: elicitation 요청을 거부(`decline`)하는 선택은 "부모 SPEC의 승인 스코핑이 MoAI 브로커 도구의 승인창을 없앤다"는 전제에 기댄다. 이것은 **관측된 사실이 아니다** — 부모 research.md(`:87`)가 실제 Factory 승인 대화상자 유무를 라이브 재확인 대상으로 남겼고, 부모의 라이브 테스트 `TestManagedCodexFactoryBrokerLive` 는 SKIP이다. 이 SPEC은 그 전제가 틀렸을 때 조용한 재배달 루프가 되지 않도록 탐지 수단(REQ-MH-002 의 로그 줄, REQ-MH-006 의 턴 단위 실패 처리)을 둔다.

이 SPEC은 부모의 **배달 전용 경계**(부모 design.md D-1)를 그대로 이어받는다: 카드 처분·병합·재알림은 건드리지 않고, 승인 권한을 넓히지 않는다.

## §B. 범위

두 가지를 `internal/cli` 의 관리 소유자와 기본값 파일 한 곳에서 닫는다.

1. **F3 — 서버 시작 요청 응답.** Codex 소유자가 서버 요청 10종 전부와 그 밖의 `id` 달린 메서드에 응답한다. 응답 정책은 최소 권한이다(design.md D-1). 서버 요청 id가 클라이언트 id와 겹치거나 문자열이어도 응답으로 오인하거나 읽기 루프를 잃지 않는다. 거부한 MoAI 브로커 elicitation 은 그 턴의 턴 단위 실패로 센다.
2. **F4 — 턴 단위 실패 격리.** 턴 하나의 실패(스트림 `result.is_error`, Codex 턴이 `completed` 가 아닌 상태로 종료, 거부된 브로커 elicitation 이 귀속된 턴)는 로그를 남기고 세션을 이어 가며, 연속 실패가 상한(`internal/config/defaults.go`)에 닿으면 세션이 끝난다. 세션 치명 오류의 분류와 claim된 메시지의 처분을 정한다(design.md D-2).

## §C. 요구사항 (GEARS)

### C.1 서버가 먼저 보내는 요청 (F3)

- **REQ-MH-001** — **When** the Codex App Server sends a server-to-client request (a frame carrying both an `id` and a `method`), the Codex session owner shall send exactly one reply carrying that request's `id` verbatim (string or integer form) from the connection reader, without waiting for a turn to be in flight.
- **REQ-MH-002** — The Codex session owner shall answer each of the ten server-request kinds defined by the codex 0.160.0 app-server schema with the reply recorded in design.md D-1 (a refusal or a grant-nothing result for every approval kind, a failed-tool result for a dynamic tool call, a JSON-RPC error where the schema offers no refusal), and shall answer any other method that carries an `id` with a JSON-RPC method-not-found error; for every request it answers, the owner shall also write one log line naming the method (an `mcpServer/elicitation/request` line also names its `serverName`, the turn it is attributed to or `none` when no turn is in flight, and the number of broker declines counted so far in that turn), so an operator can see what was declined.
- **REQ-MH-003** — The Codex session owner shall not answer any server request with an accepting decision or a non-empty permission grant; the authority the owned App Server holds beyond its defaults shall remain the MoAI broker tool pre-approval already carried on its command line.
- **REQ-MH-004** — **When** a server request carries an `id` equal to an outstanding client request `id`, or an `id` in string form, the Codex session owner shall classify the frame as a server request, shall not take it as the client request's response, and shall keep its connection reader running.
- **REQ-MH-005** — The Codex session owner shall serialize every write to the App Server connection so that no two goroutines write at the same time, replies sent from the connection reader included.

### C.2 턴 단위 실패 격리 (F4)

- **REQ-MH-006** — **When** a turn after the priming turn fails in a turn-scoped way — a Claude or GLM stream turn ending with an error result, a Codex turn ending in any state other than completed, or a Codex turn to which the owner attributed a declined `mcpServer/elicitation/request` whose `serverName` is the MoAI broker MCP server, even if the App Server marks that turn completed — the delivery driver shall, until REQ-MH-008 applies, write one operator-visible log line naming the consecutive-failure count and the cause, and shall continue the delivery loop. A request is attributed to a turn by the order in which the App Server's frames arrive, never by when the consumer reads them, and several attributed requests count as one failure; the attribution rules are recorded in design.md D-1 decision 3.
- **REQ-MH-007** — **When** a failure is session-fatal — the priming turn failing, the stream or connection closing, a write failure, a Codex per-turn timeout, or any error not classified turn-scoped — the delivery driver shall return that error as it does today.
- **REQ-MH-008** — **When** the count of consecutive turn-scoped failures reaches `DefaultManagedSessionMaxConsecutiveTurnFailures` (defined in `internal/config/defaults.go`), the delivery driver shall, notwithstanding REQ-MH-006, write the failure log line and then return the last failure's error stating the count, and a turn that completes successfully shall reset the count to zero.
- **REQ-MH-009** — The delivery driver shall not release, acknowledge, or re-address a broker message claimed by a turn that failed; redelivery follows the broker's claim-lease policy, and this layer shall add no attempt cap of its own.

### C.3 공시와 검증 자세

- **REQ-MH-010** — The operator documentation `.moai/docs/factory-managed-session.md` and this SPEC's CHANGELOG entry shall state F3 and F4 as resolved, shall state the signal-handling gap as still open and owned by card t1459 rather than as resolved, and shall name every limit that remains, without claiming behavior that was not observed; the remaining limits include a poison message redelivered until its TTL, a Codex turn timeout that ends the session, a declined elicitation that blocks the receipt and so leaves redelivery until TTL, a blocked write on the App Server connection that nothing but connection death or closing the session releases (no write deadline, and the turn timeout does not bound it), and behavior with a real codex session that was not observed.
- **REQ-MH-011** — The run phase shall land each reproduction test's RED baseline, as the tracked file `.moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/red-baseline.md` together with the reproduction tests, in a commit that precedes the commit carrying the corresponding fix, so the commit graph witnesses the ordering; `.moai/reports/t1409/` holds local copies only.
- **REQ-MH-012** — The delivery shall leave the directory of the completed parent SPEC and `internal/factorymsg/store.go` unchanged, shall keep the `managed_*` source files free of the platform-specific system-call references they carry none of today (the parent's AC-MS-014 grep stays at 0 lines), and shall keep the `GOOS=windows GOARCH=amd64` cross build passing.

## §D. 제약

- **배달 전용**: 부모 design.md D-1을 계승한다. 카드 처분·병합 창·재알림·완료 판정 상징을 들이지 않는다(부모 AC-MS-010 grep이 계속 0행).
- **승인 권한 불확대**: 모든 승인류 응답은 거부 또는 아무것도 허가하지 않는 결과다. MoAI 브로커 도구 사전 승인(`factoryMoAIMCPApprovalArgs`)만 유효하다.
- **브로커 무수정**: `internal/factorymsg/store.go` 는 수정하지 않는다. 시도 상한이나 nack 호출이 필요해 보이면 그것은 이 SPEC의 범위 밖이며 후속 과제로 남긴다.
- **하드코딩 금지**: 연속 실패 상한은 `internal/config/defaults.go` 상수다. 구현 코드에 숫자 임계값을 적지 않는다.
- **인터페이스 안정**: `managedSession` 인터페이스의 세 메서드와 `driveManagedFactorySession` 의 시그니처는 바꾸지 않는다(이 범위에는 취소 컨텍스트가 필요 없다).
- **Windows**: `GOOS=windows GOARCH=amd64 go build ./...` 가 통과해야 한다(회귀 가드).
- **검증 부하**: 레인 로컬 검증은 변경 영향 범위만 돌린다. `internal/cli` 전체 스위트는 로컬에서 돌리지 않는다. 분 단위 패키지 스위트는 `moai slot acquire --resource <이름> --max-duration <상한>` 임대 안에서만 돈다(`.claude/rules/local/gitflow-lane-protocol.md` §8).
- **문서 언어**: 문서·안내는 한국어, 코드 주석·커밋은 영어, 로그 문구는 영어(`error_messages: en`).
- **진행 조건**: plan-audit 통과가 run 착수 조건이다(autonomous Kickoff의 증거 기준).

## §E. 성공 기준 요약

`acceptance.md` 의 AC-MH-001..013 전부 통과. 특히: (1) F3·F4 각각의 재현 테스트가 기준 트리에서 RED였고 수리 뒤 GREEN이며, 추적되는 `red-baseline.md` 가 시험과 함께 수리 커밋보다 앞선 별도 커밋에 있다(`git merge-base --is-ancestor` 로 확인). (2) 서버 요청 10종과 id 충돌·문자열 id 사례를 가짜 App Server가 직접 보낸다. (3) elicitation 귀속의 각 갈래(턴 사이, 직전 턴 id, 불일치 id, 한 턴 여럿)가 하위 케이스 하나씩으로 고정된다. (4) 부모 SPEC 디렉터리와 `store.go` 의 diff가 비어 있고 Windows 크로스 빌드가 통과한다.

## §F. Exclusions

### Out of Scope — F5 signal handling and Start/Close lifecycle (card t1459)

- 런처의 SIGINT·SIGTERM·SIGHUP 처리와 `Start`/`Close` 수명주기 설계는 새 카드 t1459 몫이다 — 운영자가 검증 표면(변이 수십 개)을 줄이려고 분리했고, 이 카드는 `Start()` 의 자원 생성·`Close` 를 건드리지 않는다. 운영 문서의 "시그널 처리 공백" 줄은 해결이 아니라 t1459 를 가리키는 정확한 안내로 남는다.

### Out of Scope — TUI attach (card t1408)

- 관리 Codex 세션의 TUI 부착과 화면 출력은 다루지 않는다. 소유자는 이번에도 헤드리스다.

### Out of Scope — Codex lane child-session wiring (card t1440)

- `moai codex -f lane` 이 띄우는 카드 자식 세션에 관리 계층을 연결하는 일은 다루지 않는다.

### Out of Scope — minor items F8, F9, F13 (card t1410)

- 시작 실패 시 토큰 임시 디렉터리 잔존(F8), 10초 핸드셰이크 예산(F9), `/readyz` 리디렉션 루프백 재검사(F13)는 t1410 몫이다(이 카드 뒤에 실행).
- **t1410 과의 의존 관계 — 기준 트리 `7109e0900` 의 소스를 읽은 결과(줄 번호는 `grep`·열람으로 측정, 의존·충돌 결론은 추론)**:
  - **F8**: `Start()` 의 `MkdirTemp`(`managed_codex_factory.go:395`)–`WriteFile`(`:401`)–`cmd.Start`(`:413`) 구간의 실패 분기가 대상이다. 이 카드는 그 줄을 **고치지 않는다**(측정: 이 카드가 `Start()` 에서 건드리는 것은 클라이언트 생성 줄 `:431` 과 `initialized` 쓰기 줄 `:438` 뿐이고 둘 다 준비 대기 이후다). 그러므로 F8 은 이 카드의 F3·F4 변경에 의존하지 않는다. F8 이 의존하는 것은 `Start`/`Close` 를 구조적으로 바꾸는 쪽, 곧 **t1459** 의 수명주기 설계다(추론: 같은 줄 영역). 충돌은 이 카드와 t1410 사이가 아니라 t1410 과 t1459 사이에 생긴다.
  - **F9**: `DefaultManagedCodexReadyTimeout`(`defaults.go:117`)의 값 완화이고 사용처는 다이얼 `HandshakeTimeout`(`managed_codex_factory.go:135`)과 `readyCtx`(`:422`)다. 이 카드의 F3·F4 코드와 무관하다(F4 의 턴 타임아웃 `DefaultManagedCodexTurnTimeout`(`:121`)은 다른 상수이고 세션 치명으로 남는다). 이 카드가 `defaults.go` 에 더하는 상수는 `:121` 뒤에 놓아 텍스트 충돌 면적을 피한다(추론).
  - **F13**: `managedCodexAppReady`(`:100-126`)의 `http.Client{Timeout}`(`:104`)에 `CheckRedirect` 가 없고 `client.Do`(`:113`)가 리디렉션을 따라간다. 이 카드는 그 함수를 건드리지 않는다. 독립이다.
  - **공유 시험 파일**: 이 카드는 가짜 App Server(`serveFakeRPC`)에 서버 요청 주입 모드를 더하고 t1410 의 F8·F9 시험도 같은 가짜를 쓸 가능성이 높다(추론) — `managed_codex_factory_test.go` 에서 텍스트 충돌이 날 수 있으며 t1410 이 이 카드 뒤에 실행되므로 이 카드 위에 얹으면 된다.

### Out of Scope — broker attempt cap or nack API

- claim을 되돌리거나 시도 횟수를 세는 브로커 변경은 하지 않는다(`store.go` 무수정). 독 메시지(poison message)가 TTL까지 lease 만료마다 재배달될 수 있다는 한계는 공시하되 이 SPEC이 닫지 않는다.

### Out of Scope — turn interrupt and timeout promotion

- Codex `turn/interrupt` 호출을 새로 쓰지 않는다. 따라서 Codex 턴 타임아웃은 세션 치명으로 남는다(design.md D-2). 턴 단위로 승격하려면 인터럽트 단계를 함께 설계하는 별도 과제다.

### Out of Scope — live Codex gating

- 실제 codex 세션에서 거부 응답이 모델 동작에 미치는 영향을 관측하는 라이브 게이트는 이 SPEC의 검증 범위가 아니다(`MOAI_FACTORY_LIVE_ROOT` 미설정).

### Out of Scope — parent SPEC edits

- SPEC-FACTORY-MANAGED-SESSION-001의 본문·frontmatter·HISTORY는 수정하지 않는다. 연결은 이 SPEC의 `related_specs` 와 본문 참조로만 한다.

## §G. 교차 참조

- `plan.md` — 마일스톤 M1..M4, 사전 점검(기준선 실측 원문), Tier 판단, 위험.
- `acceptance.md` — AC-MH-001..013, RED-now/green path 쌍(시험 기반 AC는 M1이 채택), 변이 탐침, 에지 케이스, DoD.
- `design.md` — D-1(서버 요청 응답 정책표·쓰기 위치·elicitation 귀속), D-2(오류 분류·상한·claim 처분).
- `.moai/reports/t1409/intake.md` — 착수 기록과 소스 판독 사실(로컬 사본: `.moai/reports/` 는 gitignore 대상이라 커밋 그래프에 없다).
- `.moai/reports/t1409/t1375-f3-f5-verbatim.md` — 감사 지적 원문(로컬 사본).
- SPEC-FACTORY-MANAGED-SESSION-001 — 부모(완료). design.md D-1(배달 전용 경계), AC-MS-010/014/017.
