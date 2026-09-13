# MoAI Gateway 진행 상태 보고

기준 시각: 2026-09-12 00:42:52 KST (+0900)  
대상 worktree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`  
브랜치: `WT-unified-gateway` · HEAD: `81c1d58f9`  
SPEC: `SPEC-MOAI-GATEWAY-001` · version `0.9.0` · status `draft`

## Claim

현재 트리에는 다음 구현과 회귀 보강이 반영되어 있다.

- Claude Code 2.1.268의 `thinking: {"type":"adaptive","display":"omitted"}`를 검증하고 GPT Responses wire에서는 presentation-only `display`를 생략한다.
- tool의 `defer_loading: true`를 입력에서 검증하고 초기 Responses tool 목록에서 제외한다. `tool_reference` 후속 로딩은 아직 지원하지 않는다.
- 고정 GPT 구독(AuthPKCE) 경로는 `max_output_tokens`를 보내지 않는다. client가 비스트림을 요청하면 upstream에만 `stream:true`를 설정하고 엄격한 SSE를 모아 Messages JSON으로 반환한다.
- Responses 출력 message의 `phase`는 `commentary`, `final_answer`, `null`만 수용한다.
- Responses `reasoning` output item의 opaque 왕복·receipt 연결은 활성화 게이트 밖이다. 해당 응답은 502 명시 실패로 남으므로 실제 GPT 생성 성공이나 resume 호환성 PASS를 주장하지 않는다.
- SPEC 문서의 lint와 변경 영역 시험은 통과했지만, worktree에는 다른 세션의 변경을 포함한 dirty 상태가 있고 전체 제품 수용 완료는 아니다.

## Evidence

### 원본 debug log

실행 명령:

`rg -n 'unrecognized_model|Bad Request|API error|dispatching to firstParty|API REQUEST' '/Users/goos/.moai/state/gateway-conversations/families/15f670fb-7e15-4c00-802d-9246ef267ed1/native/debug/15f670fb-7e15-4c00-802d-9246ef267ed1.txt' | head -40`

관찰 출력:

```text
616:2026-09-11T14:10:16.906Z [WARN] [claude-code:unrecognized_model] {"model":"gpt-5.6-sol","query_source":"repl_main_thread:outputStyle:custom"}
622:2026-09-11T14:10:16.911Z [DEBUG] [API:timing] dispatching to firstParty model=gpt-5.6-sol
623:2026-09-11T14:10:16.913Z [DEBUG] [API REQUEST] /v1/messages source=repl_main_thread:outputStyle:custom
625:2026-09-11T14:10:16.932Z [ERROR] API error (attempt 1/11): 400 400 {"error":{"message":"Bad Request","type":"invalid_request_error"},"type":"error"}
626:2026-09-11T14:10:16.945Z [ERROR] Error in API request: 400 {"error":{"message":"Bad Request","type":"invalid_request_error"},"type":"error"}
629:2026-09-11T14:10:16.947Z [ERROR] [engine] turn ended in error: API Error: 400 Bad Request
```

### 변경된 gateway 시험

실행 명령:

`go test ./internal/gateway -run 'TestOpenAISubscription(NonStreamCollectsSSE|OutputPolicyWireAndValidation)' -count=1 -v -timeout=120s`

관찰 출력:

```text
=== RUN   TestOpenAISubscriptionNonStreamCollectsSSE
--- PASS: TestOpenAISubscriptionNonStreamCollectsSSE (0.06s)
=== RUN   TestOpenAISubscriptionOutputPolicyWireAndValidation
--- PASS: TestOpenAISubscriptionOutputPolicyWireAndValidation (0.03s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.387s
```

실행 명령:

`go test ./internal/gateway/translate -run 'TestNativePolicyObservedClaudeRequest|TestNonStreamAcceptsObservedOutputMessagePhase' -count=1 -v -timeout=120s`

관찰 출력:

```text
=== RUN   TestNativePolicyObservedClaudeRequest
--- PASS: TestNativePolicyObservedClaudeRequest (0.00s)
=== RUN   TestNonStreamAcceptsObservedOutputMessagePhase
--- PASS: TestNonStreamAcceptsObservedOutputMessagePhase (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	0.168s
```

### 선행 연결·경계 시험 재확인

실행 명령:

```text
go test -race ./internal/cli -run 'TestGPTBinding|TestGatewayGPTModels|TestGatewayChildEnvironment|TestProductionGatewayFactory|TestGPTClosedVerbs' -count=1 -timeout=120s
```

관찰 출력:

```text
ok   	github.com/modu-ai/moai-adk/internal/cli	2.495s
```

실행 명령:

```text
go test ./internal/cli -run 'TestRunCC_(NoProjectRoot|FindProjectRootFails|WithTeamModeMessage)|^TestCGEntryGuardRunsBeforeLaunchAndSpawn$' -count=1 -v -timeout=120s
```

관찰 출력:

```text
=== RUN   TestRunCC_NoProjectRoot
--- PASS: TestRunCC_NoProjectRoot (0.00s)
=== RUN   TestRunCC_WithTeamModeMessage
Team mode disabled (was: glm)
Launching Claude Code...
--- PASS: TestRunCC_WithTeamModeMessage (0.00s)
=== RUN   TestCGEntryGuardRunsBeforeLaunchAndSpawn
--- PASS: TestCGEntryGuardRunsBeforeLaunchAndSpawn (0.00s)
=== RUN   TestRunCC_FindProjectRootFails
--- PASS: TestRunCC_FindProjectRootFails (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.988s
```

실행 명령:

```text
go test -race ./internal/gateway -run '^TestNativeEventRawByteBoundaries$' -count=1 -v -timeout=120s
```

관찰 출력:

```text
=== RUN   TestNativeEventRawByteBoundaries
CRLF=false fragmented=false largest_raw_event=145 event_limit=144 success=false err=native response stream failed
CRLF=false fragmented=false largest_raw_event=145 event_limit=145 success=true err=<nil>
CRLF=false fragmented=false largest_raw_event=145 event_limit=146 success=true err=<nil>
CRLF=false fragmented=true largest_raw_event=145 event_limit=144 success=false err=native response stream failed
CRLF=false fragmented=true largest_raw_event=145 event_limit=145 success=true err=<nil>
CRLF=false fragmented=true largest_raw_event=145 event_limit=146 success=true err=<nil>
CRLF=true fragmented=false largest_raw_event=148 event_limit=147 success=false err=native response stream failed
CRLF=true fragmented=false largest_raw_event=148 event_limit=148 success=true err=<nil>
CRLF=true fragmented=false largest_raw_event=148 event_limit=149 success=true err=<nil>
CRLF=true fragmented=true largest_raw_event=148 event_limit=147 success=false err=native response stream failed
CRLF=true fragmented=true largest_raw_event=148 event_limit=148 success=true err=<nil>
CRLF=true fragmented=true largest_raw_event=148 event_limit=149 success=true err=<nil>
--- PASS: TestNativeEventRawByteBoundaries (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway	1.312s
```

### 변경 패키지·동시성·정적 검증

실행 명령:

`go test ./internal/gateway/... -count=1 -timeout=180s`

관찰 출력:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway	6.757s
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	3.536s
ok  	github.com/modu-ai/moai-adk/internal/gateway/conversation	1.549s
ok  	github.com/modu-ai/moai-adk/internal/gateway/opaque	0.756s
ok  	github.com/modu-ai/moai-adk/internal/gateway/receipt	2.900s
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	1.817s
```

실행 명령:

`go test ./internal/cli -run 'TestGateway|TestGPT|TestNativeReceipt|TestNative|TestFamily|TestCRLF|TestOpaque|TestReceipt' -count=1 -timeout=180s`

관찰 출력:

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	1.328s
```

실행 명령:

`go test -race ./internal/gateway/... -count=1 -timeout=180s`

관찰 출력:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway	9.278s
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	6.034s
ok  	github.com/modu-ai/moai-adk/internal/gateway/conversation	2.971s
ok  	github.com/modu-ai/moai-adk/internal/gateway/opaque	3.206s
ok  	github.com/modu-ai/moai-adk/internal/gateway/receipt	4.014s
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	2.676s
```

실행 명령:

`go vet ./internal/gateway/... ./internal/cli && echo 'go vet: PASS'`

관찰 출력:

```text
go vet: PASS
```

### SPEC·Windows·로그인 상태

실행 명령:

`go run ./cmd/moai spec lint .moai/specs/SPEC-MOAI-GATEWAY-001/spec.md`

관찰 출력:

`✓ No findings — all SPEC documents are valid`

실행 명령:

`GOOS=windows GOARCH=amd64 go test -c ./internal/gateway/auth -o /tmp/gateway-auth-windows.test.exe`  
`GOOS=windows GOARCH=amd64 go test -c ./internal/cli -o /tmp/gateway-cli-windows.test.exe`

관찰 출력:

```text
windows test binaries: present
```

이는 macOS에서의 교차 컴파일 확인이며 Windows API 실행 결과가 아니다.

실행 명령:

`bash scripts/ci-census/windows-gateway-evidence-test.sh`

관찰 출력:

```text
"PASS Windows gateway native evidence: 2 required tests"
REJECTED missing
REJECTED skipped
REJECTED failed
REJECTED pass_without_run
REJECTED truncated
REJECTED malformed
REJECTED absent_stream
```

실행 명령:

`/tmp/moai-gateway-clean gpt status`

관찰 출력:

`GPT: logged in`

### Worktree 기준선

실행 명령:

`git rev-parse --short HEAD`  
`git branch --show-current`  
`git status --short | wc -l`  
`git fetch origin main 2>&1 && git rev-list --count --left-right origin/main...HEAD`

관찰 출력:

```text
81c1d58f9
WT-unified-gateway
     268
From https://github.com/modu-ai/moai-adk
 * branch                main       -> FETCH_HEAD
0	2879
```

### 실제 구독 SSE shape 계측

임시 shape 계측은 본문·토큰을 기록하지 않고 다음 형태만 기록했다.

```text
native fields thinking_present=true thinking_keys=display,type ... stream=true
thinking scalar type=adaptive display=omitted
tools count=12 shapes=defer_loading,description,input_schema,name|description,input_schema,name defer_true=1
sse ... item_type=reasoning
sse ... item_keys=content,encrypted_content,id,summary,type item_type=reasoning
sse ... item_keys=content,id,phase,role,status,type item_status=in_progress item_phase=final_answer item_role=assistant
subscription response conversion error: item must start in progress
subscription response conversion error: unsupported field "phase"
```

계측은 진단 후 제거했으며 현재 `rg -n 'gatewayDebug|streamDebugEvent|MOAI_GATEWAY_DIAGNOSTIC' internal/gateway` 출력은 비어 있고 `internal/gateway/debug_tmp.go`는 존재하지 않는다.

## Baseline-attribution

모든 위 명령은 이 worktree와 현재 소스 트리에 대해 2026-09-12 KST에 다시 실행했다. 원본 debug log는 사용자가 제공한 절대 경로에서 line 616–629를 읽었다. test와 lint 출력은 현재 HEAD `81c1d58f9`에서 측정했으며, uncommitted 변경은 HEAD에 포함되지 않은 현재 작업 트리 상태다. Windows 항목은 교차 컴파일과 synthetic CI parser 시험의 범위만 측정했다. 실제 Windows runner·전원 차단·Win32 ACL 실행은 측정하지 않았다.

## Gaps

- `reasoning` output item의 encrypted opaque를 Messages `redacted_thinking` 또는 Responses 후속 input으로 안전하게 왕복시키는 receipt 통합이 아직 활성화되지 않았다. 따라서 실제 `gpt-5.6-sol` 요청은 현재 정상 답변까지 도달했다고 말할 수 없다.
- `tool_reference`를 이용한 deferred tool 후속 로딩은 지원하지 않는다.
- 구독 nonstream은 endpoint가 직접 지원하는 것이 아니라 upstream SSE 수집으로 흉내 낸 경로다. 서버 출력 정책과 MoAI byte/cancel/deadline 한도는 동일하지 않다.
- Windows 시험은 GitHub Actions에서 실제 권한·잠금·강제 종료 복구를 실행해야 한다. 현재 로컬 교차 컴파일은 실행 증거가 아니다.
- Claude OAuth, GPT PKCE 실제 inference, tool 왕복, receipt/resume/fork, TEAMMATE 및 전체 acceptance matrix의 제품 통합은 미완료다.
- 6차 delta plan audit는 Opus API rate limit(HTTP 429)로 실행되지 않아 독립 PASS 판정이 없다.
- worktree status 268건에는 이번 카드와 무관한 공유 세션 변경도 포함되어 있으므로 이 보고서는 commit·merge·push 상태를 뜻하지 않는다.

## Residual-risk

- Claude Code가 새 native 필드·event 종류를 추가하면 현재 exact validator가 400/502로 fail closed한다. 이는 조용한 오라우팅을 막지만 새 버전 호환성을 자동 보장하지 않는다.
- 공식 Responses 문서는 reasoning item을 후속 입력에 포함하라고 요구하며, 현재 gateway는 이를 보존하지 않는다. 같은 provider resume에서 문맥 손실 또는 upstream 거절이 남는다. [OpenAI Responses streaming reference](https://platform.openai.com/docs/api-reference/responses-streaming/response/refusal?lang=python)
- deferred tool을 초기 목록에서 제외하면 최초 prompt 가시성은 보존되지만 후속 `tool_reference` 요청은 실패할 수 있다. [Anthropic tool reference](https://platform.claude.com/docs/en/agents-and-tools/tool-use/tool-reference)
- `thinking.display`의 `omitted`는 GPT wire에서 표현 필드로 생략한다. `summarized`를 reasoning 응답으로 변환하는 계약은 확인되지 않아 GPT profile에서 거절한다. [Anthropic Messages API](https://platform.claude.com/docs/en/api/http/messages)

## Next actions

1. opaque/receipt 활성화 게이트를 별도 계획 감사와 실제 유실·resume 음성 시험으로 통과시킨 뒤 reasoning item 왕복을 구현한다.
2. GitHub Actions Windows runner에서 ACL·sharing·flush·원자 교체·강제 종료 복구를 실행하고 artifact/run ID를 기록한다.
3. 실제 GPT 네 모델의 text·tool·title·resume 왕복을 수행하되 reasoning 실패와 정상 결과를 각각 별도 판정한다.
4. PICKER·AUTH·TEAMMATE 형제 SPEC이 착지하기 전에는 코어 전체 출시 또는 제품 PASS를 선언하지 않는다.

## Modified surface

- `internal/gateway/openai.go` — subscription output policy, upstream SSE force/collector.
- `internal/gateway/translate/request.go` — `defer_loading` validation and omission.
- `internal/gateway/translate/native_policy.go` — adaptive display validation.
- `internal/gateway/translate/response.go` / `stream.go` — output message phase support and subscription collector.
- `internal/gateway/openai_subscription_test.go` / `openai_test.go` / `internal/cli/gateway_factory_test.go` — SSE and output-policy regressions.
- `.moai/specs/SPEC-MOAI-GATEWAY-001/{spec,plan,design,acceptance,research}.md` — measured contract and gaps.
