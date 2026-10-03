# red-baseline.md — M1 RED 기준선 (SPEC-FACTORY-MANAGED-HARDEN-001, 카드 t1409)

> run 단계 증거 파일이다(SPEC 본문이 아니다). 추적 경로이고 `.moai/reports/` 가 아니라서 커밋 그래프에 남는다. 이 파일과 재현 시험 6개는 **한 커밋**에 있고, 그 커밋은 수리 커밋(M2·M3)의 조상이어야 한다(REQ-MH-011, AC-MH-011).

## 측정 조건

- 측정 트리 SHA: `fe79bfa0e5e28e8a91e53c8e067fadfe30c632b6` (기준 HEAD, 브랜치 `WT-managed-session-hardening`, 이 run 에서 측정). 시험 파일은 그 HEAD 위의 작업 트리에 있었고 프로덕션 파일은 하나도 바뀌지 않았다.
- 도구: `go version go1.26.8 darwin/arm64`. 판정 주체는 `go test` 이며 프로젝트 자체 빌드(`moai`)가 아니므로 VCI §2.2 의 두 번째 좌표는 해당 없음.
- 모든 `go test` 는 환경 정리 접두를 붙인 단일 호출이다: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND && go test …` (아래 표의 명령은 접두를 줄여 적는다).
- 선택 수(`-list '^.*(Managed|managed).*$'` 의 `^Test` 줄 수): **이전 66 → 이후 72** (최상위 시험 6개 추가). 시험 6개만 고른 `-list` 는 이름 6개를 모두 냈다.
- 시험 파일: `internal/cli/managed_hardening_test.go`(신규: 시험 6개, 가짜 서버 `newHardenPair`, 가짜 스트림 세션), `internal/cli/managed_codex_factory_test.go`(기존 가짜 App Server 에 턴별 상태 순서 환경변수 `T1409_FAKE_APP_SERVER_TURN_STATUSES` 와 `turn-prompt` 로그 줄 추가 — 시험 파일 안에서만).

## 요약

| # | 시험 | AC | 명령 | exit | 붉은 이유(한 줄) | 옳은 이유? |
|---|---|---|---|---|---|---|
| 1 | `TestManagedCodexServerRequestPolicy` | AC-MH-001 | `go test -race ./internal/cli -run '^TestManagedCodexServerRequestPolicy$' -count=1 -v` | 1 | 소유자가 서버 요청 11건 어느 것에도 답장을 쓰지 않아 가짜 서버가 2 s 뒤 "응답 없음"을 기록 | 예 — 단언 대상 관측(응답 부재) |
| 2 | `TestManagedCodexTurnSurvivesServerRequest` | AC-MH-002 | `go test ./internal/cli -run '^TestManagedCodexTurnSurvivesServerRequest$' -count=1 -v` | 1 | 답이 없어 서버가 턴을 끝내지 않으므로 시험 자신의 5 s 상한에서 `DeliverTurn` 이 막힌 채 남음(`during_turn`), 턴 사이 요청은 5 s 안에 답을 받지 못함(`between_turns`) | 예 — 시험 소유 5 s 상한(10분 상수·패키지 타임아웃 아님) |
| 3 | `TestManagedCodexServerRequestIDCollision` | AC-MH-003 | `go test ./internal/cli -run '^TestManagedCodexServerRequestIDCollision$' -count=1 -v` | 1 | 같은 정수 id 서버 요청이 응답으로 오인돼 진짜 응답이 소실되고(`result ""`), 문자열 id 는 읽기 루프를 끝내 연결 닫힘 오류가 남 | 예 |
| 4 | `TestManagedCodexDeclinedBrokerElicitationFailsTurn` | AC-MH-006 #13·#14 | `go test -race ./internal/cli -run '^TestManagedCodexDeclinedBrokerElicitationFailsTurn$' -count=1 -v` | 1 | 거부된 `moai` 브로커 elicitation 이 있었는데도 App Server 가 턴을 `completed` 로 끝내면 `DeliverTurn` 이 nil 을 반환 | 예 |
| 5 | `TestManagedDriverIsolatesTurnFailure` | AC-MH-005 | `go test -race ./internal/cli -run '^TestManagedDriverIsolatesTurnFailure$' -count=1 -v` | 1 | 두 번째 턴의 `is_error` 결과로 드라이버가 반환해 세 번째 턴이 전달되지 않음 | 예 |
| 6 | `TestManagedCodexNonCompletedTurnIsolated` | AC-MH-006 #2 | `go test -race ./internal/cli -run '^TestManagedCodexNonCompletedTurnIsolated$' -count=1 -v` | 1 | 첫 비완료 상태(`failed`)에서 드라이버가 반환해 이후 운영자 턴 2개가 App Server 에 전달되지 않음 | 예 |

여섯 시험 모두 컴파일은 통과했고(`go vet ./internal/cli` exit 0), 패키지 타임아웃·하네스 인프라 실패로 붉은 것은 없다. 시험 2 의 `between_turns` 는 시험이 정한 5 s 대기가 만료되어 붉은 것이며 그 대기가 곧 단언 대상 관측이다.

## 1. TestManagedCodexServerRequestPolicy

명령: `go test -race ./internal/cli -run '^TestManagedCodexServerRequestPolicy$' -count=1 -v` (정리 접두 포함). exit 1. 트리 `fe79bfa0e`.

결정 출력(원문, 중간의 `=== RUN` 줄 생략):

```
    managed_hardening_test.go:383: no answer to server request 101 (item/commandExecution/requestApproval) within 2s: the owner does not answer server-originated requests
    managed_hardening_test.go:383: no answer to server request 102 (item/fileChange/requestApproval) within 2s: the owner does not answer server-originated requests
    managed_hardening_test.go:383: no answer to server request 103 (item/permissions/requestApproval) within 2s: the owner does not answer server-originated requests
    managed_hardening_test.go:383: no answer to server request 104 (mcpServer/elicitation/request) within 2s: the owner does not answer server-originated requests
    managed_hardening_test.go:383: no answer to server request 105 (item/tool/requestUserInput) within 2s: the owner does not answer server-originated requests
    managed_hardening_test.go:383: no answer to server request 106 (item/tool/call) within 2s: the owner does not answer server-originated requests
    managed_hardening_test.go:383: no answer to server request 107 (account/chatgptAuthTokens/refresh) within 2s: the owner does not answer server-originated requests
    managed_hardening_test.go:383: no answer to server request 108 (attestation/generate) within 2s: the owner does not answer server-originated requests
    managed_hardening_test.go:383: no answer to server request 109 (applyPatchApproval) within 2s: the owner does not answer server-originated requests
    managed_hardening_test.go:383: no answer to server request 110 (execCommandApproval) within 2s: the owner does not answer server-originated requests
    managed_hardening_test.go:383: no answer to server request 111 (item/unlisted/needsAnswer) within 2s: the owner does not answer server-originated requests
--- FAIL: TestManagedCodexServerRequestPolicy (2.02s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	3.801s
```

하위 테스트 11개 전부 `--- FAIL`. 붉은 이유: 기준 `read()` 가 `id != 0` 프레임을 이벤트 채널에 올려 `call()`/`waitTurn()` 이 버리므로 서버는 응답을 받지 못한다. 시험이 기대하는 응답 본문(design.md D-1 정책표)은 M2 에서 채워진다. 로그 줄 단언(로그 이음새)은 이음새가 기준 API 에 없어 M2 에서 더한다.

## 2. TestManagedCodexTurnSurvivesServerRequest

명령: `go test ./internal/cli -run '^TestManagedCodexTurnSurvivesServerRequest$' -count=1 -v` (정리 접두 포함). exit 1. 트리 `fe79bfa0e`.

```
=== RUN   TestManagedCodexTurnSurvivesServerRequest/during_turn
    managed_hardening_test.go:416: DeliverTurn still blocked after 5s: the server request was never answered, so the server never completed the turn
=== RUN   TestManagedCodexTurnSurvivesServerRequest/between_turns
    managed_hardening_test.go:446: a server request arriving between turns got no answer within 5s
--- FAIL: TestManagedCodexTurnSurvivesServerRequest (10.01s)
    --- FAIL: TestManagedCodexTurnSurvivesServerRequest/during_turn (5.01s)
    --- FAIL: TestManagedCodexTurnSurvivesServerRequest/between_turns (5.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	11.298s
```

붉은 이유: 서버가 답을 받아야만 턴을 끝내는데 소유자가 답하지 않으므로 시험의 5 s watchdog 이 먼저 깨진다. 10분 `DefaultManagedCodexTurnTimeout` 이나 패키지 타임아웃이 아니라 시험 소유 상한이다.

## 3. TestManagedCodexServerRequestIDCollision

명령: `go test ./internal/cli -run '^TestManagedCodexServerRequestIDCollision$' -count=1 -v` (정리 접두 포함). exit 1. 트리 `fe79bfa0e`.

```
=== RUN   TestManagedCodexServerRequestIDCollision/integer_collision
    managed_hardening_test.go:499: call result "", want the genuine response {"genuine":true} (the server request was taken for the response)
=== RUN   TestManagedCodexServerRequestIDCollision/string_id
    managed_hardening_test.go:513: call error = managed codex app server connection closed, want the genuine response (a string-id request must not end the read loop)
--- FAIL: TestManagedCodexServerRequestIDCollision (0.01s)
    --- FAIL: TestManagedCodexServerRequestIDCollision/integer_collision (0.00s)
    --- FAIL: TestManagedCodexServerRequestIDCollision/string_id (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.248s
```

붉은 이유: 정수 id 충돌에서 `call()` 이 서버 요청을 응답으로 반환(빈 결과·오류 없음)하고, 문자열 id 는 `ID int` 디코딩 실패로 읽기 고루틴이 끝나 이후 호출이 "connection closed" 가 된다 — acceptance.md §2.1 의 보조 실측과 같은 두 형태.

## 4. TestManagedCodexDeclinedBrokerElicitationFailsTurn

명령: `go test -race ./internal/cli -run '^TestManagedCodexDeclinedBrokerElicitationFailsTurn$' -count=1 -v` (정리 접두 포함). exit 1. 트리 `fe79bfa0e`.

```
=== RUN   TestManagedCodexDeclinedBrokerElicitationFailsTurn/single_request_null_turn_id
    managed_hardening_test.go:578: DeliverTurn = nil for a turn whose MoAI broker elicitation was declined, want a non-nil error
=== RUN   TestManagedCodexDeclinedBrokerElicitationFailsTurn/two_requests_in_one_turn
    managed_hardening_test.go:578: DeliverTurn = nil for a turn whose MoAI broker elicitation was declined, want a non-nil error
=== RUN   TestManagedCodexDeclinedBrokerElicitationFailsTurn/request_before_turn_started
    managed_hardening_test.go:578: DeliverTurn = nil for a turn whose MoAI broker elicitation was declined, want a non-nil error
--- FAIL: TestManagedCodexDeclinedBrokerElicitationFailsTurn (0.01s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.598s
```

세 하위 케이스(§1.3 #13·#14 의 컴파일 가능 형태)가 모두 `--- FAIL`. 붉은 이유: 기준 `DeliverTurn` 은 턴이 `completed` 이면 nil 을 반환한다. 단언은 "nil 이 아닌 오류"뿐이며 표식 심볼은 M3 이 도입한다.

## 5. TestManagedDriverIsolatesTurnFailure

명령: `go test -race ./internal/cli -run '^TestManagedDriverIsolatesTurnFailure$' -count=1 -v` (정리 접두 포함). exit 1. 트리 `fe79bfa0e`.

```
=== RUN   TestManagedDriverIsolatesTurnFailure
    managed_hardening_test.go:638: driver returned managed Factory turn failed: overloaded after one failed turn, want it to keep delivering and end nil on /exit
    managed_hardening_test.go:645: turns delivered = ["MoAI Factory 세션 준비 완료라고 한 줄로 답해. 아직 작업은 시작하지 마." "op-fails"], want ["MoAI Factory 세션 준비 완료라고 한 줄로 답해. 아직 작업은 시작하지 마." "op-fails" "op-succeeds"]
--- FAIL: TestManagedDriverIsolatesTurnFailure (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.451s
```

붉은 이유: 실제 `pumpManagedStreamTurn` 이 두 번째 턴의 `is_error` 결과를 오류로 반환하고, 기준 드라이버가 그 오류를 그대로 반환해 세 번째 턴(`op-succeeds`)이 전달되지 않는다. `Factory turn failed (1/` 로그 줄 단언은 M3 의 이음새에 달린다.

## 6. TestManagedCodexNonCompletedTurnIsolated

명령: `go test -race ./internal/cli -run '^TestManagedCodexNonCompletedTurnIsolated$' -count=1 -v` (정리 접두 포함). exit 1. 트리 `fe79bfa0e`. 가짜 App Server 상태 순서 `completed,failed,interrupted,completed`(우선 턴 + 운영자 턴 3개).

```
=== RUN   TestManagedCodexNonCompletedTurnIsolated
    managed_hardening_test.go:680: driver returned managed codex turn fake-turn-2 ended as failed after a non-completed turn, want it to keep delivering and end nil on /exit
    managed_hardening_test.go:691: turn "op-interrupted" was never delivered to the app server:
        …
        method turn/start
        turn-prompt MoAI Factory 세션 준비 완료라고 한 줄로 답해. 아직 작업은 시작하지 마.
        method turn/start
        turn-prompt op-failed
    managed_hardening_test.go:691: turn "op-normal" was never delivered to the app server:
        …
        turn-prompt op-failed
--- FAIL: TestManagedCodexNonCompletedTurnIsolated (0.94s)
```

(앞쪽 핸드셰이크 로그 줄은 `…` 로 줄였다. 원문은 `token-loaded`·`server-cwd …`·`auth-ok`·`method initialize`·`method initialized`·`method thread/start`·`thread-cwd …`·`method thread/name/set` 이며 시험이 핸드셰이크 성공을 이미 확인했다.) 붉은 이유: `failed` 로 끝난 첫 운영자 턴에서 드라이버가 반환해 `op-interrupted`·`op-normal` 이 서버에 닿지 않는다.

## 회귀 확인 (이 run, 이 트리 `fe79bfa0e`)

기존 managed 시험은 새 시험 6개를 `-skip` 으로 뺀 단일 호출로 돌렸다.

명령: `go test ./internal/cli -run '^.*(Managed|managed).*$' -skip '^(TestManagedCodexServerRequestPolicy|TestManagedCodexTurnSurvivesServerRequest|TestManagedCodexServerRequestIDCollision|TestManagedCodexDeclinedBrokerElicitationFailsTurn|TestManagedDriverIsolatesTurnFailure|TestManagedCodexNonCompletedTurnIsolated)$' -count=1 -v` (정리 접두 포함)

관측: exit 0, 최상위 `--- PASS` 63, `--- SKIP` 3, `--- FAIL` 0(합 66 = 기준 선택 수), 끝 줄 `ok  github.com/modu-ai/moai-adk/internal/cli  41.464s`. 공유 가짜 서버(`serveFakeRPC`)를 건드렸는데 기존 시험은 영향이 없다.

## Gaps (미관측)

- 여섯 시험의 **GREEN 가능성**은 관측하지 못했다. 프로덕션 수리가 없는 이 커밋에서는 각 시험의 통과 경로(정책표 응답 본문 검사, 상한 내 반환, 표식 오류 반환)가 실행된 적 없다. M2·M3 이 관측한다.
- 시험은 클라이언트를 `managedCodexAppClient{conn, events, done}` 구조체 리터럴로 직접 만든다(`newHardenPair`). M2 가 이 구조체에 필수 초기화를 더하면 이 도우미 한 곳을 고쳐야 한다.
- 로그 이음새(`Factory server request answered:`, `Factory turn failed (`) 단언은 기준 API 에 이음새가 없어 M1 시험에 없다. M2/M3 이 추가한다.
- `-race` 를 붙인 시험(1, 4, 5, 6)은 각각 데이터 경합 보고 0건으로 붉어졌다(경합 보고가 붉은 이유가 아님). 경합 단언 자체(AC-MH-004)는 M2 의 `TestManagedCodexConcurrentWrites` 몫이다.
- `gofmt -l internal/cli` 는 `internal/cli/mcp_claude.go` 한 줄을 낸다 — 기준 트리에서 이미 있고(미수정 파일, `git status` 에 없음) 이 카드가 만든 것이 아니다. 이 카드가 건드린 두 시험 파일은 `gofmt -l` 출력이 비어 있다.

## Residual-risk

- 시험 2 의 `between_turns` 는 시험 자신의 5 s 대기에 붉으므로 기계가 몹시 느리면 GREEN 에서도 5 s 안에 답이 안 올 수 있다(로컬 loopback 왕복은 ms 단위라는 가정, 측정하지 않음).
- 시험 6 은 실제 자식 프로세스(재실행한 시험 바이너리)를 쓰므로 Windows 에서는 스킵된다(`fakeAppServerScript`). Windows 는 크로스 빌드·vet 로만 덮인다.
- 스크래치 출력(`/private/tmp/.../scratchpad/red1.txt` 등)은 이 파일에 필요한 줄을 모두 옮겼고 인용 대상이 아니다.
