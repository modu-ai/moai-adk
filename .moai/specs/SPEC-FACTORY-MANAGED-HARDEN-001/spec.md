---
id: SPEC-FACTORY-MANAGED-HARDEN-001
title: "Factory 관리 세션 소유자 강건화 — 서버 요청 응답, 턴 단위 실패 격리, 시그널 정리 (SPEC-FACTORY-MANAGED-SESSION-001 후속)"
version: "0.1.0"
status: draft
created: 2026-10-03
updated: 2026-10-03
author: GOOS (manager-spec)
priority: P1
phase: "v3.2.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tier: M
related_specs: [SPEC-FACTORY-MANAGED-SESSION-001]
tags: "factory, managed-session, codex, app-server, hardening, failure-isolation, signal"
---

# SPEC-FACTORY-MANAGED-HARDEN-001 — Factory 관리 세션 소유자 강건화

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-10-03 | manager-spec | Initial SPEC (card t1409). 완료된 SPEC-FACTORY-MANAGED-SESSION-001의 독립 sync 감사 지적 F3·F4·F5를 닫는 후속 SPEC이다. 부모 SPEC은 본문·frontmatter 모두 수정하지 않는다. |

## §A. 개요

부모 SPEC-FACTORY-MANAGED-SESSION-001(상태 `completed`)이 만든 관리 세션 계층의 Codex/Claude 소유자에는 독립 sync 감사(`.moai/reports/t1375/sync-audit.md`, 원문 인용: `.moai/reports/t1409/t1375-f3-f5-verbatim.md`)가 남긴 세 가지 공백이 있다. 운영자는 2026-10-02 결정으로 이 셋을 후속 카드 t1409로 닫기로 했다. 이 SPEC은 그 셋만 다룬다.

| 지적 | 현상 | 관측 등급 |
|---|---|---|
| F3 | 서버가 먼저 보내는 요청(`id`와 `method`를 함께 가진 프레임)에 응답하지 않는다. | 아래 세 항목은 기준 트리 `7109e0900` 에서 임시 테스트로 **실측**했다(테스트는 커밋하지 않고 삭제함): (1) 서버가 보낸 요청 id=7에 클라이언트가 1초 안에 어떤 프레임도 돌려보내지 않았다. (2) 서버 요청 id=1이 대기 중인 클라이언트 id=1과 같을 때 `call()` 이 서버 요청을 응답으로 오인해 빈 결과를 오류 없이 돌려주고, 진짜 응답 본문은 버려졌다. (3) 서버 요청 id가 문자열(`"srv-1"`)이면 읽기 루프가 디코딩 오류로 끝나 연결이 닫힌 것으로 처리됐다. |
| F4 | 턴 하나가 실패하면 `driveManagedFactorySession` 이 반환해 세션 전체가 끝난다. | 소스 판독(`internal/cli/managed_factory_session.go:345`)과 기존 테스트 `TestManagedDriverFailureBranches`. 분류 기준은 SPEC에 없었다(감사서 §F4). |
| F5 | 관리 소유자에 시그널 처리가 없다. | 소스 판독: `internal/cli/managed_*.go` 비테스트 파일에서 `signal.Notify`·`NotifyContext` 0행. 메커니즘은 별도 작은 프로그램으로 실측: `defer` 정리와 자식·임시 디렉터리를 가진 프로그램이 SIGTERM(종료 코드 143)·SIGHUP(129)을 받자 `defer` 가 돌지 않고 자식이 살아 있고 임시 디렉터리가 남았다. 실제 런처에 대한 kill 실험은 하지 않았다. SIGINT는 해당 측정 하네스(비대화형 셸의 백그라운드 잡은 SIGINT를 무시함)의 한계로 관측하지 못했다. |

보조 사실(전부 기준 트리 `7109e0900` 에서 관측):

- codex 0.160.0 `app-server generate-json-schema` 로 서버→클라이언트 요청이 10종(`item/commandExecution/requestApproval`, `item/fileChange/requestApproval`, `item/permissions/requestApproval`, `mcpServer/elicitation/request`, `item/tool/requestUserInput`, `item/tool/call`, `account/chatgptAuthTokens/refresh`, `attestation/generate`, `applyPatchApproval`, `execCommandApproval`)이고, 요청 `id` 는 문자열 또는 int64라는 것을 확인했다(`RequestId` = `anyOf [string, int64]`). 현재 클라이언트의 프레임 구조체는 `id` 를 `int` 로 선언한다.
- 브로커 `internal/factorymsg/store.go` 의 `messages` 테이블에는 시도 횟수 열이 없고(`:389`), 클라이언트 쪽 API는 `Claim`·`ReadBody`·`RecordDisposition`·`Receipt` 뿐이라 claim을 되돌리는 호출이 없다. `Claim` 은 `state='claimed'` 이고 `claim_expires_at` 이 지난 행을 다시 후보로 삼아 새 claim token으로 재배달한다(`:878`, `:901`, `:916`; 기존 테스트 `TestDispatchResultExactlyOnce/lost_receipt_redelivery` 가 기준 트리에서 PASS). 메시지 수명은 TTL(`MaxTTL` 7일, `:40`)이고 TTL이 지난 행은 `Claim` 이 dead-letter로 보낸다(`:892-898`).

- **미관측 전제**: elicitation 요청을 거부(`decline`)하는 선택은 "부모 SPEC의 승인 스코핑이 MoAI 브로커 도구의 승인창을 없앤다"는 전제에 기댄다. 이것은 **관측된 사실이 아니다** — 부모 research.md(`:87`)가 실제 Factory 승인 대화상자 유무를 라이브 재확인 대상으로 남겼고, 부모의 라이브 테스트 `TestManagedCodexFactoryBrokerLive` 는 SKIP이다. 이 SPEC은 그 전제가 틀렸을 때 조용한 재배달 루프가 되지 않도록 탐지 수단(REQ-MH-002 의 로그 줄, REQ-MH-006 의 턴 단위 실패 처리)을 둔다.

이 SPEC은 부모의 **배달 전용 경계**(부모 design.md D-1)를 그대로 이어받는다: 카드 처분·병합·재알림은 건드리지 않고, 승인 권한을 넓히지 않는다.

## §B. 범위

세 가지를 `internal/cli` 의 관리 소유자와 기본값 파일 한 곳에서 닫는다.

1. **F3 — 서버 시작 요청 응답.** Codex 소유자가 서버 요청 10종 전부와 그 밖의 `id` 달린 메서드에 응답한다. 응답 정책은 최소 권한이다(design.md D-1). 서버 요청 id가 클라이언트 id와 겹치거나 문자열이어도 응답으로 오인하거나 읽기 루프를 잃지 않는다.
2. **F4 — 턴 단위 실패 격리.** 턴 하나의 실패(스트림 `result.is_error`, Codex 턴이 `completed` 가 아닌 상태로 종료)는 로그를 남기고 세션을 이어 가며, 연속 실패가 상한(`internal/config/defaults.go`)에 닿으면 세션이 끝난다. 세션 치명 오류의 분류와 claim된 메시지의 처분을 정한다(design.md D-2).
3. **F5 — 시그널 정리.** SIGINT·SIGTERM·SIGHUP이 기존 정리 경로(세션 종료, launch-pending 롤백, 토큰 디렉터리 삭제)에 닿는다. 진행 중인 턴 전달이 블로킹 읽기여도 풀린다(design.md D-3).

## §C. 요구사항 (GEARS)

### C.1 서버가 먼저 보내는 요청 (F3)

- **REQ-MH-001** — **When** the Codex App Server sends a server-to-client request (a frame carrying both an `id` and a `method`), the Codex session owner shall send exactly one reply carrying that request's `id` verbatim (string or integer form) from the connection reader, without waiting for a turn to be in flight.
- **REQ-MH-002** — The Codex session owner shall answer each of the ten server-request kinds defined by the codex 0.160.0 app-server schema with the reply recorded in design.md D-1 (a refusal or a grant-nothing result for every approval kind, a failed-tool result for a dynamic tool call, a JSON-RPC error where the schema offers no refusal), and shall answer any other method that carries an `id` with a JSON-RPC method-not-found error; for every request it answers, the owner shall also write one stderr line naming the method (an `mcpServer/elicitation/request` line also names its `serverName`), so an operator can see what was declined.
- **REQ-MH-003** — The Codex session owner shall not answer any server request with an accepting decision or a non-empty permission grant; the authority the owned App Server holds beyond its defaults shall remain the MoAI broker tool pre-approval already carried on its command line.
- **REQ-MH-004** — **When** a server request carries an `id` equal to an outstanding client request `id`, or an `id` in string form, the Codex session owner shall classify the frame as a server request, shall not take it as the client request's response, and shall keep its connection reader running.
- **REQ-MH-005** — The Codex session owner shall serialize every write to the App Server connection so that no two goroutines write at the same time, replies sent from the connection reader included.

### C.2 턴 단위 실패 격리 (F4)

- **REQ-MH-006** — **When** a turn after the priming turn fails in a turn-scoped way — a Claude or GLM stream turn ending with an error result, a Codex turn ending in any state other than completed, or a Codex turn during which the owner declined an `mcpServer/elicitation/request` whose `serverName` is the MoAI broker MCP server (`moai`), even if the App Server marks that turn completed — the delivery driver shall write one operator-visible line to stderr naming the consecutive-failure count and the cause, and shall continue the delivery loop.
- **REQ-MH-007** — **When** a failure is session-fatal — the priming turn failing, the stream or connection closing, a write failure, a Codex per-turn timeout, or any error not classified turn-scoped — the delivery driver shall return that error as it does today.
- **REQ-MH-008** — **When** the count of consecutive turn-scoped failures reaches `DefaultManagedSessionMaxConsecutiveTurnFailures` (defined in `internal/config/defaults.go`), the delivery driver shall return the last failure's error stating the count, and a turn that completes successfully shall reset the count to zero.
- **REQ-MH-009** — The delivery driver shall not release, acknowledge, or re-address a broker message claimed by a turn that failed; redelivery follows the broker's claim-lease policy, and this layer shall add no attempt cap of its own.

### C.3 시그널 정리 (F5)

- **REQ-MH-010** — **When** the launcher receives SIGINT, SIGTERM, or SIGHUP while it owns a managed session, the launcher shall run the existing teardown — session close, launch-pending rollback, token directory removal — and return an error naming the interruption, instead of ending through the runtime's default signal action.
- **REQ-MH-011** — **When** a signal arrives while a turn delivery is blocked on the backend, the launcher shall unblock that delivery by tearing the child process down, so the owner returns within the bound the acceptance test states (5 seconds).
- **REQ-MH-012** — The managed session `Start` and `Close` shall be safe to call concurrently and in either order, and `Close` shall be safe to call repeatedly and shall perform the teardown exactly once: a session closed before `Start`, during `Start`, or right after `Start` published its connection shall leave no child process, no token directory, and no open connection behind, and a `Start` called after `Close` shall be refused.
- **REQ-MH-013** — The signal handling shall build for `GOOS=windows GOARCH=amd64`, and the `managed_*` source files shall keep their zero-`syscall` property.

**Explicit cross-platform exemption (EXCL-syscall, signal constants only).** The signal helper `internal/cli/launch_signals.go` references `syscall.SIGTERM` and `syscall.SIGHUP` as signal constants only. Both constants exist on Windows, the helper makes no system call, and the Windows cross build is gated by AC-MH-011. This is an explicit, reasoned exception to the wording of parent REQ-MS-012 ("zero `syscall` usage in newly added files"): the parent's purpose is the child-process ownership model with one code path on Windows (no process-replacement exec, no platform-split files), and a constants-only reference leaves that purpose intact. The same package already carries the precedent `internal/cli/mcp_server.go:123` (`signal.NotifyContext(…, os.Interrupt, syscall.SIGTERM)` with no build tag). The `managed_*` files themselves keep zero `syscall.` references, so parent AC-MS-014's grep stays at 0 lines.

### C.4 공시와 검증 자세

- **REQ-MH-014** — The operator documentation `.moai/docs/factory-managed-session.md` and this SPEC's CHANGELOG entry shall state F3, F4, and F5 as resolved and shall name every limit that remains, without claiming behavior that was not observed; the remaining limits include a poison message redelivered until its TTL, a Codex turn timeout that ends the session, a declined elicitation that blocks the receipt and so leaves redelivery until TTL, the absence of a write deadline on the App Server connection, a second signal during teardown ending the process by the default action, descendant processes of the child surviving its kill, Windows console events other than Ctrl-C, and behavior with a real codex session that was not observed.
- **REQ-MH-015** — The run phase shall land each reproduction test's RED baseline, as the tracked file `.moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/red-baseline.md` together with the reproduction tests, in a commit that precedes the commit carrying the corresponding fix, so the commit graph witnesses the ordering; `.moai/reports/t1409/` holds local copies only.
- **REQ-MH-016** — The delivery shall leave the directory of the completed parent SPEC and `internal/factorymsg/store.go` unchanged.

## §D. 제약

- **배달 전용**: 부모 design.md D-1을 계승한다. 카드 처분·병합 창·재알림·완료 판정 상징을 들이지 않는다(부모 AC-MS-010 grep이 계속 0행).
- **승인 권한 불확대**: 모든 승인류 응답은 거부 또는 아무것도 허가하지 않는 결과다. MoAI 브로커 도구 사전 승인(`factoryMoAIMCPApprovalArgs`)만 유효하다.
- **브로커 무수정**: `internal/factorymsg/store.go` 는 수정하지 않는다. 시도 상한이나 nack 호출이 필요해 보이면 그것은 이 SPEC의 범위 밖이며 후속 과제로 남긴다.
- **하드코딩 금지**: 연속 실패 상한은 `internal/config/defaults.go` 상수다. 구현 코드에 숫자 임계값을 적지 않는다.
- **인터페이스 안정**: `managedSession` 인터페이스의 세 메서드(`Start`, `DeliverTurn`, `Close`)는 바꾸지 않는다(design.md D-3 선택의 귀결).
- **Windows**: `GOOS=windows GOARCH=amd64 go build ./...` 가 통과해야 한다. `managed_*.go` 의 `syscall.` 참조는 0행을 유지한다(부모 AC-MS-014 grep 유지). 새 `launch_signals.go` 의 `syscall` 참조는 시그널 상수뿐이며 시스템 호출이 없다 — 이것은 부모 REQ-MS-012 문면에 대한 명시적 예외이고 cross-platform exemption(EXCL-syscall)의 근거와 선례(`mcp_server.go:123`)는 C.3에 있으며 Windows 크로스 빌드는 AC-MH-011이 게이트한다.
- **검증 부하**: 레인 로컬 검증은 변경 영향 범위만 돌린다. `internal/cli` 전체 스위트는 로컬에서 돌리지 않는다. 분 단위 패키지 스위트는 `moai slot acquire --resource <이름> --max-duration <상한>` 임대 안에서만 돈다(`.claude/rules/local/gitflow-lane-protocol.md` §8).
- **문서 언어**: 문서·안내는 한국어, 코드 주석·커밋은 영어, 로그 문구는 영어(`error_messages: en`).
- **진행 조건**: plan-audit 통과가 run 착수 조건이다(autonomous Kickoff의 증거 기준).

## §E. 성공 기준 요약

`acceptance.md` 의 AC-MH-001..016 전부 통과. 특히: (1) F3·F4·F5 각각의 재현 테스트가 기준 트리에서 RED였고 수리 뒤 GREEN이며, 추적되는 `red-baseline.md` 가 시험과 함께 수리 커밋보다 앞선 별도 커밋에 있다(`git merge-base --is-ancestor` 로 확인). (2) 서버 요청 10종과 id 충돌 사례를 가짜 App Server가 직접 보낸다. (3) 시그널 테스트가 자식 프로세스 부재와 토큰 디렉터리 삭제를 단언한다. (4) Windows 크로스 빌드가 통과한다. (5) 부모 SPEC 디렉터리와 `store.go` 의 diff가 비어 있다.

## §F. Exclusions

### Out of Scope — TUI attach (card t1408)

- 관리 Codex 세션의 TUI 부착과 화면 출력은 다루지 않는다. 소유자는 이번에도 헤드리스다.

### Out of Scope — Codex lane child-session wiring (card t1440)

- `moai codex -f lane` 이 띄우는 카드 자식 세션에 관리 계층을 연결하는 일은 다루지 않는다.

### Out of Scope — minor items F8, F9, F13 (card t1410)

- 시작 실패 시 토큰 임시 디렉터리 잔존(F8), 10초 핸드셰이크 예산(F9), `/readyz` 리디렉션 루프백 재검사(F13)는 t1410 몫이다(이 카드 뒤에 실행).
- **겹침 경계(`Start()` 영역)**: 이 카드는 `Start`/`Close` 수명주기 규칙(design.md D-3)에 필요한 **닫힌 세션 경로**의 정리만 한다 — `Close` 가 먼저 또는 도중에 불린 `Start` 가 자기가 만든 자식(kill+wait)·토큰 디렉터리·연결을 스스로 치우는 것. 세션이 **열려 있는데** `Start` 가 중간에 실패하는 경로(F8의 토큰 디렉터리 잔존)와 핸드셰이크 예산(F9)은 이 카드가 바꾸지 않는다. 겹치는 코드는 토큰 디렉터리 삭제 호출 한 곳이고, 이 카드는 그것을 닫힌 경로용 작은 도우미로 두어 t1410이 실패 경로에서 재사용할 수 있게 한다. `Start()` 의 그 밖의 변경은 준비 대기 컨텍스트를 세션의 닫힘에 연결해 `Close` 가 그 대기를 취소하게 하는 것이다(핸드셰이크 예산 값 자체는 F9로 t1410 몫이며 바꾸지 않는다).

### Out of Scope — broker attempt cap or nack API

- claim을 되돌리거나 시도 횟수를 세는 브로커 변경은 하지 않는다(`store.go` 무수정). 독 메시지(poison message)가 TTL까지 lease 만료마다 재배달될 수 있다는 한계는 공시하되 이 SPEC이 닫지 않는다.

### Out of Scope — turn interrupt and timeout promotion

- Codex `turn/interrupt` 호출을 새로 쓰지 않는다. 따라서 Codex 턴 타임아웃은 세션 치명으로 남는다(design.md D-2). 턴 단위로 승격하려면 인터럽트 단계를 함께 설계하는 별도 과제다.

### Out of Scope — live Codex gating and Windows native certification

- 실제 codex 세션에서 거부 응답이 모델 동작에 미치는 영향을 관측하는 라이브 게이트는 이 SPEC의 검증 범위가 아니다(`MOAI_FACTORY_LIVE_ROOT` 미설정). Windows 실세션과 콘솔 종료 이벤트(Ctrl-C 외)도 인증하지 않는다.

### Out of Scope — parent SPEC edits

- SPEC-FACTORY-MANAGED-SESSION-001의 본문·frontmatter·HISTORY는 수정하지 않는다. 연결은 이 SPEC의 `related_specs` 와 본문 참조로만 한다.

## §G. 교차 참조

- `plan.md` — 마일스톤 M1..M5, 사전 점검(기준선 실측 원문), 위험.
- `acceptance.md` — AC-MH-001..016, RED-now/green path 쌍(시험 기반 AC는 M1이 채택), 에지 케이스, DoD.
- `design.md` — D-1(서버 요청 응답 정책표와 쓰기 위치), D-2(오류 분류·상한·claim 처분), D-3(시그널 설계 비교).
- `.moai/reports/t1409/intake.md` — 착수 기록과 소스 판독 사실(로컬 사본: `.moai/reports/` 는 gitignore 대상이라 커밋 그래프에 없다).
- `.moai/reports/t1409/t1375-f3-f5-verbatim.md` — 감사 지적 원문(로컬 사본).
- SPEC-FACTORY-MANAGED-SESSION-001 — 부모(완료). design.md D-1(배달 전용 경계), D-4(자식 프로세스 소유), AC-MS-010/014/017.
