# t652 App Server AS3 구현 증거

## 5초 상한 이후 late allocation 재수리

부모의 실제 subprocess 재현에서 6초 뒤 turn/started가 도착하면 첫 수리의 5초 상관관계 상한 후 interrupt가 누락됨을 다시 확인했다. 이 경계를 단순 한계 고지로 닫지 않고 수정했다. `late-start-red.log`는 같은 반례를 이 트리에서 다시 실행한 실패 증거다.

- known failed/canceled thread의 늦은 turn/started에서 소유 turn ID를 얻어 정확한 turn/interrupt를 보낸다. 대화마다 늦은 정리 worker는 한 번만 배정하며, 중복 notification과 다른 turn ID로 작업 수가 늘어나지 않는다. 기존 MaxConversations/큐 상한과 process context를 유지한다.
- stop 전에 이미 queue에 들어온 turn/started도 버리지 않고 ID를 보존하여 기존 종료 경로가 interrupt할 수 있게 했다. `queued-late-start-red.log`에 수리 전 누락을 보존했다.
- RPC reply 상관관계는 계속 최대 5초지만, 그 상한 이후의 소유 allocation notification도 별도로 처리한다. 공유 App Server 프로세스를 종료하는 방식은 사용하지 않는다.

`TestCanceledStartBeyondReplyDeadlineStillInterrupts`는 실제 subprocess에서 6초 지연 후 같은 turn/started를 반복하고 다른 turn ID도 보낸다. fixture는 interrupt의 threadId/turnId를 검사하고 기록이 정확히 한 번인지 확인하며 shared transport가 살아 있음을 검사한다. 기존 250ms late reply와 QueueSize2의 A/B 취소 후 C 진행 테스트도 유지했다.

최종 검증:

```sh
go test -race -p 1 ./internal/codexapp ./internal/codexbridge ./internal/gateway -coverprofile=.moai/reports/t652/late-repair-coverage.out -count=1 -timeout=120s
```

```text
ok  	github.com/modu-ai/moai-adk/internal/codexapp	2.326s	coverage: 86.6% of statements
ok  	github.com/modu-ai/moai-adk/internal/codexbridge	11.466s	coverage: 86.2% of statements
ok  	github.com/modu-ai/moai-adk/internal/gateway	10.504s	coverage: 91.7% of statements
```

`go vet`, Windows bridge 교차 컴파일, git diff --check 모두 exit 0이다. 함수별 수치는 `late-repair-function-coverage.txt`에 보존했다. 소스 16파일 manifest SHA-256은 `bc012a4b8d0768ee533ffcd678803d0aaf8479889b3aecd3c8f7030a2e45f8df`이다. 이 재수리는 engine.go, engine_test.go, lifecycle_subprocess_test.go에 한정했고 commit/push/install은 수행하지 않았다.

정확한 판정 범위: 5초를 넘겨 소유 turn/started가 실제 전달되는 재현은 PASS다. 프로세스 종료 또는 upstream이 RPC reply와 allocation notification을 모두 전달하지 않아 ID 자체가 없는 상황까지 exact interrupt 성공을 주장하지 않는다. thread/start 도중 취소된 빈 thread 회수, 실제 모델/CLI/E2E 및 Windows 실행은 기존 gap으로 남는다. 독립 재감사 판정은 PENDING이다.

## 첫 F1/F2 수리 기록

독립 감사에서 최초 구현의 취소 경계 두 가지가 실제 subprocess로 실패했다. F1은 HTTP context 취소로 turn/start 상관관계가 먼저 사라져 늦게 배정된 turn을 interrupt하지 못했다. F2는 취소된 server RPC가 Client.requests에 남아 QueueSize 2에서 세 번째 대화가 실패했다. 최초 fakeRPC가 context 취소를 따르지 않아 F1을 검증하지 못한 점도 확인했다.

수리 내용:

- turn/start는 HTTP context와 분리된 프로세스 context 아래에서 최대 5초 동안 응답 상관관계를 유지한다. HTTP 요청은 취소로 끝나더라도 늦은 응답의 해당 turn ID를 받아 interrupt한다. 다른 대화의 프로세스를 종료하지 않는다. 작업은 Engine.Close에서 정리한다.
- Client.DiscardRequest는 소유자가 취소·실패시킨 RPC ID만 응답 없이 폐기한다. Engine은 pending 및 중단 대화에 늦게 들어온 RPC 등록을 제거한다. 큐 상한은 유지한다. unknown/invalid ID가 기존 다른 요청을 소비하지 않으며 폐기 후 같은 ID의 응답·폐기 재시도를 거절함을 검증했다.
- 늦은 이벤트와 stop의 큐 정리를 같은 대화 잠금으로 직렬화했다. 이전 turn ID가 새 turn의 취소 대상으로 쓰이지 않도록 새 시작 전에 비운다.

독립 재현을 `internal/codexbridge/lifecycle_subprocess_test.go`에 보존했다. setup 프로세스 lifetime은 느린 파일 시스템에서도 시나리오가 도달하도록 5초에서 30초로 조정했다. 250ms 늦은 응답, QueueSize 2, A/B 취소 후 C 계속이라는 재현 조건은 유지했다. 짧은 50ms 추측 대기는 실제 turn/start 진입 신호 후 취소하는 시험으로 교체했다.

```sh
go test -p 1 ./internal/codexbridge ./internal/codexapp -run 'TestAudit|TestDiscardRequest|TestCancelBefore' -v -count=1 -timeout=90s
```

```text
=== RUN   TestCancelBeforeTurnStartReplyInterruptsAllocatedTurn
--- PASS: TestCancelBeforeTurnStartReplyInterruptsAllocatedTurn (0.43s)
=== RUN   TestAuditRealTransportEOFWakesBridge
    lifecycle_subprocess_test.go:58: Step err=Codex App Server EOF transport err=Codex App Server EOF elapsed=218.515167ms done=false
--- PASS: TestAuditRealTransportEOFWakesBridge (1.05s)
=== RUN   TestAuditCancelBeforeRealStartReplyInterruptsTurn
    lifecycle_subprocess_test.go:133: Step err=context canceled transport err=<nil> interrupt observed=true
--- PASS: TestAuditCancelBeforeRealStartReplyInterruptsTurn (2.26s)
=== RUN   TestAuditCanceledRPCDoesNotExhaustSharedTransport
--- PASS: TestAuditCanceledRPCDoesNotExhaustSharedTransport (1.43s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/codexbridge	5.634s
=== RUN   TestDiscardRequestPreservesOtherIDsAndRejectsReplay
--- PASS: TestDiscardRequestPreservesOtherIDsAndRejectsReplay (0.03s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/codexapp	0.451s
```

최종 범위 검증(추가 신뢰 경계 시험 포함):

```sh
go test -race -p 1 ./internal/codexapp ./internal/codexbridge ./internal/gateway -coverprofile=.moai/reports/t652/repair-coverage.out -count=1 -timeout=120s
```

```text
ok  	github.com/modu-ai/moai-adk/internal/codexapp	2.501s	coverage: 87.0% of statements
ok  	github.com/modu-ai/moai-adk/internal/codexbridge	5.656s	coverage: 86.5% of statements
ok  	github.com/modu-ai/moai-adk/internal/gateway	15.628s	coverage: 91.7% of statements
```

`go vet`는 exit 0, Windows bridge 교차 컴파일은 exit 0이었다. 함수별 결과는 `repair-function-coverage.txt`에 있다. private path 검증은 92.3%, AppServerAdapter.Send는 94.4%다. 모든 함수가 85% 이상인 것은 아니다(appServerResponse 81.6%, Authorize 83.3%). 이 수치를 패키지 coverage와 구별한다. 상대·누락·부모 symlink 경로, grant/scope/model 불일치, subscription→API 암묵 전환 거절, account/read 중 generation 변경, 잘못된 stream, text+tool SSE, readonly/비공개성 위반 파일 장벽, 큐 상한과 잘못된 server request를 검증했다.

중간 실행에서는 파일 시스템 Lstat 지연과 짧은 fixture deadline으로 실패가 있었다. 해당 로그 `audit-repair-race.log`, `repair-boundaries.log`를 보존한다. 실패한 배치의 coverage로 PASS를 주장하지 않으며 위 최종 성공 배치만 기준으로 삼는다. 원래 반례 RED는 `audit-red.log`, `audit-f2-red.log`, `discard-red.log`에 보존했다.

현재 소스 16개 파일의 SHA-256은 `as3-repair-source-sha256.json`에 기록했다. 이 수리에서 추가 변경한 파일은 `internal/codexapp/client.go`, `client_test.go`, `internal/codexbridge/lifecycle_subprocess_test.go`, `failure_boundaries_test.go`이며 기존 AS3 파일 일부도 수정했다. commit/push/install은 수행하지 않았다. 마지막 fetch 비교는 origin/main 대비 `0 3305`, branch/HEAD는 기존 값과 같았다.

첫 수리 당시 경계(위 재수리로 보완): 프로세스가 종료되거나 5초 상관관계 상한을 넘겨 turn ID를 얻지 못한 경우까지 interrupt 성공을 보장하지 않았다. 해당 대화는 복구 오류로 남고 RPC를 재생하지 않는다. thread/start 중 HTTP 취소로 늦게 배정된 빈 thread의 회수·resume는 구현하지 않았으며, 그 뒤 새 turn을 시작하지 않는다. 실제 모델/CLI/E2E/Windows 실행과 기존 제품 연결 gap도 그대로 남는다. 독립 재감사 판정은 아직 PENDING이다.

## 최초 제출 기록

## Claim

공식 App Server 클라이언트와 AS2 도구 레지스트리를 연결하는 AS3 구현을 독립 감사에 제출한다. 기존 gateway 보존 기준 위에 새 엔진·내구 상태 장벽·managed authority·HTTP adapter를 추가했다. 현재 단계는 로컬 구현 검증이며 제품 출시·CLI 연결·실제 GPT 응답 성공을 뜻하지 않는다.

- 한 App Server 이벤트 리더가 thread별 bounded queue로 분배한다. HTTP tool segment 종료는 보류 RPC를 끊지 않으며 다음 요청의 결과가 같은 RPC ID에 한 번 전달된다.
- 서로 다른 대화의 결과 교환, 결과 재전송, 취소 이후 결과를 거절한다. turn/start 응답 전에 발생한 취소도 나중에 배정된 해당 turn을 interrupt한다.
- 도구 공개 전 waiting, RPC 응답 전 responding을 디스크에 기록한다. 새 엔진은 기존 상태가 있으면 명시적 복구 오류로 중단한다. 이전 RPC를 새 프로세스에 재생하지 않는다.
- managed 경로는 공식 account/read 종류와 신뢰된 로컬 profile/account generation을 검사한다. 기존 CredentialRef 저장소를 호출하지 않는다. 이 generation은 upstream account ID가 아니다.
- 사용자 승인에 따라 chatgpt와 apiKey 모두 App Server 출력 정책을 사용한다. Claude max_tokens와 동일한 생성 상한은 보장하지 않으며 별도 응답 바이트 제한과 취소 제한을 적용한다. 자동 구독→API 전환은 없다.
- thread/start와 turn/start 양쪽에 공식 필드 environments: []를 전달한다. 이는 AS1의 프로세스 환경 변수 정리와 별개의 Codex environment access 제한이다.

## Evidence

마지막 실행:

```sh
go test -race ./internal/codexapp ./internal/codexbridge ./internal/gateway -coverprofile=.moai/reports/t652/as3-coverage.out -count=1
```

```text
ok  	github.com/modu-ai/moai-adk/internal/codexapp	2.713s	coverage: 85.2% of statements
ok  	github.com/modu-ai/moai-adk/internal/codexbridge	3.200s	coverage: 77.9% of statements
ok  	github.com/modu-ai/moai-adk/internal/gateway	14.199s	coverage: 90.6% of statements
```

```sh
go vet ./internal/codexapp ./internal/codexbridge ./internal/gateway
```

```text
exit=0
```

```sh
GOOS=windows GOARCH=amd64 go test -c ./internal/codexbridge -o /tmp/t652-codexbridge-windows.test.exe
```

```text
exit=0
```

`TestAppServerSubprocessHTTPToolContinuation/{chatgpt,apiKey}`는 실제 Python subprocess를 codexapp.Start로 실행하고 initialize, account/read, thread/start, turn/start, string ID의 item/tool/call 응답을 통과한다. Gateway 첫 HTTP는 JSON tool_use, 다음 HTTP는 동일 RPC 결과를 받아 SSE 답변과 message_stop을 반환한다. 결과 재전송은 비200이며 성공 terminal이 없다. 이는 실제 모델을 호출하지 않는 프로토콜 fixture다.

엔진 테스트는 A/B 교차 결과 거절·A 취소 후 B 완료, EOF/출력 상한/실패 turn/외부 turn/알 수 없는 RPC·도구/잘못된 인수, 보류 상태 재시작, responding 장벽 이후 응답 불확실성의 무재생을 확인했다. responding fault fixture는 실제 전원 차단 실험이 아니다.

공식 설치본에서 생성한 `probe/schema/v2/ThreadStartParams.json`과 `TurnStartParams.json`의 properties.environments를 직접 조회했다. 양쪽 모두 빈 배열이 environment access를 끈다고 정의한다. subprocess fixture는 두 호출에서 이 배열을 요구한다. 기존 미지원 env 필드 제거만으로 같은 계약이 성립한다고 보지 않는다.

### RED 기록

### bridge-red.log

```text
# github.com/modu-ai/moai-adk/internal/codexbridge [github.com/modu-ai/moai-adk/internal/codexbridge.test]
internal/codexbridge/engine_test.go:66:25: undefined: Request
internal/codexbridge/engine_test.go:67:9: undefined: Request
internal/codexbridge/engine_test.go:69:30: undefined: Engine
internal/codexbridge/engine_test.go:69:49: undefined: FileStore
internal/codexbridge/engine_test.go:71:16: undefined: OpenStore
internal/codexbridge/engine_test.go:76:17: undefined: New
internal/codexbridge/engine_test.go:76:48: undefined: Config
internal/codexbridge/engine_test.go:98:16: undefined: ToolResult
internal/codexbridge/engine_test.go:98:58: undefined: Content
internal/codexbridge/engine_test.go:126:16: undefined: ToolResult
internal/codexbridge/engine_test.go:126:16: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/codexbridge [build failed]
FAIL
```

### private-path-red.log

```text
# github.com/modu-ai/moai-adk/internal/codexapp [github.com/modu-ai/moai-adk/internal/codexapp.test]
internal/codexapp/private_test.go:12:12: undefined: ValidatePrivatePath
internal/codexapp/private_test.go:13:12: undefined: ValidatePrivatePath
internal/codexapp/private_test.go:15:10: undefined: ValidatePrivatePath
internal/codexapp/private_test.go:16:77: undefined: ValidatePrivatePath
FAIL	github.com/modu-ai/moai-adk/internal/codexapp [build failed]
FAIL
```

### managed-red.log

```text
# github.com/modu-ai/moai-adk/internal/gateway [github.com/modu-ai/moai-adk/internal/gateway.test]
internal/gateway/appserver_test.go:20:13: undefined: AppServerAuthority
internal/gateway/appserver_test.go:25:30: r.Managed undefined (type RoutedRequest has no field or method Managed)
internal/gateway/appserver_test.go:26:15: undefined: ErrManagedAuthority
internal/gateway/appserver_test.go:28:37: r.Managed undefined (type RoutedRequest has no field or method Managed)
internal/gateway/appserver_test.go:36:20: undefined: NewAppServerAuthority
internal/gateway/appserver_test.go:41:126: undefined: AuthAppServer
internal/gateway/appserver_test.go:46:139: unknown field ManagedAuthority in struct literal of type ServerConfig
internal/gateway/appserver_test.go:70:20: undefined: NewAppServerAuthority
internal/gateway/appserver_test.go:74:107: undefined: AuthAppServer
FAIL	github.com/modu-ai/moai-adk/internal/gateway [build failed]
FAIL
```

### http-red.log

```text
# github.com/modu-ai/moai-adk/internal/gateway [github.com/modu-ai/moai-adk/internal/gateway.test]
internal/gateway/appserver_integration_test.go:90:18: undefined: NewAppServerAdapter
internal/gateway/appserver_integration_test.go:90:38: undefined: AppServerAdapterConfig
FAIL	github.com/modu-ai/moai-adk/internal/gateway [build failed]
FAIL
```

### cancel-start-red.log

```text
--- FAIL: TestCancelBeforeTurnStartReplyInterruptsAllocatedTurn (0.09s)
    engine_test.go:271: allocated turn not interrupted []
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/codexbridge	0.778s
FAIL
```

### scope-delimiter-red.log

```text
--- FAIL: TestScopeDelimiterCannotAliasAnotherConversation (0.22s)
    engine_test.go:395: ambiguous scope admitted <nil>
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/codexbridge	0.738s
FAIL
```

### environments-red.log

```text
--- FAIL: TestAppServerSubprocessHTTPToolContinuation (1.19s)
    appserver_integration_test.go:130: 502 {"error":{"message":"Bad Gateway","type":"api_error"},"type":"error"}
        
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway	1.870s
FAIL
```

### api-policy-red.log

```text
--- FAIL: TestAppServerSubprocessHTTPToolContinuation (2.19s)
    --- FAIL: TestAppServerSubprocessHTTPToolContinuation/apiKey (0.96s)
        appserver_integration_test.go:133: 502 {"error":{"message":"Bad Gateway","type":"api_error"},"type":"error"}
            
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway	2.770s
FAIL
```

## Baseline-attribution

- 작업 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t652`
- 브랜치: `WT-gateway-appserver-turns`
- 기준 HEAD: `342efdcfd44ada42dad495682455c3e9019510c6`
- 부모가 지정한 source_session_id: `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`
- 이번 도구에서 관측한 `moai session current`: `27710212-8852-4379-973d-c17f979ca5ca`. source_session_id를 이 값으로 바꾸지 않았다.
- 이전 gateway 112개 파일과 homestate 보정은 기준 커밋에 포함된 별도 보존 작업이다. 위 검증은 현재 AS3 변경이 더해진 이 트리에서 실행했다.
- 변경 파일: `internal/codexapp/private.go`, `private_test.go`; `internal/codexbridge/engine.go`, `engine_test.go`, `store.go`; `internal/gateway/appserver.go`, `appserver_authority.go`, `appserver_test.go`, `appserver_integration_test.go`, `catalog.go`, `router.go`, `server.go`.
- 독립 코드 감사 전이며 새 AS3 커밋·push·PR·설치는 수행하지 않았다. 기존 `.moai/reports/t652/verdict.md`는 보존했다.

## Gaps

1. 최초 제출의 codexbridge coverage는 77.9%였다. 위 수리 후 최종 배치에서 86.5%로 바뀌었다. 함수별 미측정 분기는 `repair-function-coverage.txt`에 명시하며 모든 함수의 85% 충족을 주장하지 않는다.
2. 실제 GPT subscription/API 모델 호출, 실제 Claude Code 화면의 대화·병렬 에이전트 실행은 수행하지 않았다. 프로토콜 subprocess fixture만 통과했다.
3. 신뢰된 Prepare 경계의 실제 HTTP identity/public-history-prefix/reference 결합 구현은 AS4/AS5 소관이다. 현재 테스트는 신뢰된 fixture resolver를 주입한다. 모델 metadata를 인증 근거로 사용하지 않는다.
4. Windows는 교차 컴파일만 했다. ACL·원자 교체·강제 종료 복구의 실제 실행은 승인된 GitHub CI에서 검증해야 한다. 갑작스러운 전원 차단 내구성은 검증하지 않았다.
5. FileStore는 호출자가 AS1 profile lifetime lease 아래의 private directory를 제공해야 한다. 별도 프로세스가 공유하는 저장소 broker를 구현하지 않았다. 같은 profile로 여러 launcher를 동시에 띄우는 기능은 별개다.
6. 현재 HTTP adapter는 segment를 모아 반환한다. 토큰이 도착할 때마다 실시간 SSE로 전달하는 성능은 검증하지 않았다. usage의 0은 호환용 미계측 값이며 실제 토큰 사용량이라고 주장하지 않는다.
7. 미디어 도구 결과와 이미지 입력의 전체 제품 계약, idle 상태의 resume/fork/모델 전환, 계정 generation을 설정·회전시키는 실제 CLI 연결은 미완료다. API/server 정책 선택의 UX 표시는 후속 CLI 단계에 필요하다.
8. repository-wide 테스트 판정의 소유자는 integration branch의 GitHub CI다. 해당 CI 실행은 아직 없으므로 PENDING이다. 로컬 변경 범위 테스트로 이를 대체하지 않는다.

## Residual-risk

- 프로세스 kill 대신 오류를 주입한 복구 테스트이므로 모든 실제 장애 시점을 입증하지 않는다. 재시작 시 기존 상태를 전부 중단하는 보수적 방식은 자동 복구보다 사용성 제약이 크다.
- private path 검사와 atomic file replace는 같은 사용자 또는 다른 프로세스의 무단 동시 파일 변경을 허용하는 broker가 아니다. 호출자의 profile lease 준수가 필요하다.
- account/read 자체에는 안정된 upstream account ID가 없다. 실제 계정 변경의 generation 관리와 pending 취소는 후속 소유자 구현에서 검증해야 한다.
- environments 빈 배열과 AS1 process 환경 정리는 서로 다른 제한이며, native 기능 전체의 실행 불가 여부를 이 fixture만으로 입증하지 않는다.
- 독립 재감사와 실제 제품 연결 검증이 남아 있으므로 AS3 전체 완료 또는 운영 사용 가능 판정을 내리지 않는다.
