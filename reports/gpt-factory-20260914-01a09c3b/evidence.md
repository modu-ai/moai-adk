# 실행 근거 — moai gpt 팩토리 조사

## Claim

기존 develop 코드의 선택 테스트와 synthetic 이력 검사 결과를 보존한다. 실계정 Codex 또는 실제 Claude 팩토리 성공을 주장하지 않는다.

## Baseline-attribution

- 날짜: 2026-09-14
- 워크트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/develop`
- source_session_id: `01a09c3b-3734-7c30-b65d-650e3c63d0aa`
- 브랜치/HEAD 명령: `git branch --show-current && git rev-parse --short HEAD`

```text
develop
c9ceff175
```

설치 MoAI binary와 이 source는 동일 baseline이 아니다. 보고서의 런타임 로그는 별도로 식별한다.

## Evidence

### 1. CLI 및 모델 슬롯

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN OPENAI_API_KEY Z_AI_API_KEY && GOCACHE=/tmp/moai-gpt-factory-research-cache go test ./internal/cli -run 'TestGPTLaunchPreservesCommonEntry|TestParseFactoryFlag|TestResolveFactoryBranch|TestEnterFactory.*Env|TestGatewayProviderContractPickerAndSlots' -count=1 -v -timeout 90s
```

원출력, exit 0:

```text
=== RUN   TestParseFactoryFlag
=== PAUSE TestParseFactoryFlag
=== RUN   TestParseFactoryFlagPassThroughBoundary
=== PAUSE TestParseFactoryFlagPassThroughBoundary
=== RUN   TestResolveFactoryBranch
=== PAUSE TestResolveFactoryBranch
=== RUN   TestEnterFactoryLeadModeEnv
--- PASS: TestEnterFactoryLeadModeEnv (0.00s)
=== RUN   TestEnterFactoryWorkerModeEnv
--- PASS: TestEnterFactoryWorkerModeEnv (0.00s)
=== RUN   TestGatewayProviderContractPickerAndSlots
--- PASS: TestGatewayProviderContractPickerAndSlots (0.00s)
=== RUN   TestGPTLaunchPreservesCommonEntry
--- PASS: TestGPTLaunchPreservesCommonEntry (0.00s)
=== CONT  TestParseFactoryFlag
=== CONT  TestResolveFactoryBranch
--- PASS: TestResolveFactoryBranch (0.00s)
=== CONT  TestParseFactoryFlagPassThroughBoundary
--- PASS: TestParseFactoryFlagPassThroughBoundary (0.00s)
=== RUN   TestParseFactoryFlag/absent
=== PAUSE TestParseFactoryFlag/absent
=== RUN   TestParseFactoryFlag/bare_-f
=== PAUSE TestParseFactoryFlag/bare_-f
=== RUN   TestParseFactoryFlag/long_form_bare
=== PAUSE TestParseFactoryFlag/long_form_bare
=== RUN   TestParseFactoryFlag/-f_N
=== PAUSE TestParseFactoryFlag/-f_N
=== RUN   TestParseFactoryFlag/-f=N
=== PAUSE TestParseFactoryFlag/-f=N
=== RUN   TestParseFactoryFlag/--factory_N
=== PAUSE TestParseFactoryFlag/--factory_N
=== RUN   TestParseFactoryFlag/--factory=N
=== PAUSE TestParseFactoryFlag/--factory=N
=== RUN   TestParseFactoryFlag/-f_lane-2
=== PAUSE TestParseFactoryFlag/-f_lane-2
=== RUN   TestParseFactoryFlag/-f=lane-3
=== PAUSE TestParseFactoryFlag/-f=lane-3
=== RUN   TestParseFactoryFlag/--factory=lane-7
=== PAUSE TestParseFactoryFlag/--factory=lane-7
=== RUN   TestParseFactoryFlag/positional_flag_is_not_a_value
=== PAUSE TestParseFactoryFlag/positional_flag_is_not_a_value
=== RUN   TestParseFactoryFlag/zero_count_errors
=== PAUSE TestParseFactoryFlag/zero_count_errors
=== RUN   TestParseFactoryFlag/negative_joined_count_errors
=== PAUSE TestParseFactoryFlag/negative_joined_count_errors
=== RUN   TestParseFactoryFlag/non-numeric_non-lane_errors
=== PAUSE TestParseFactoryFlag/non-numeric_non-lane_errors
=== RUN   TestParseFactoryFlag/unnumbered_lane_errors
=== PAUSE TestParseFactoryFlag/unnumbered_lane_errors
=== RUN   TestParseFactoryFlag/worker_zero_errors
=== PAUSE TestParseFactoryFlag/worker_zero_errors
=== RUN   TestParseFactoryFlag/lookalike_long_flag_not_stolen
=== PAUSE TestParseFactoryFlag/lookalike_long_flag_not_stolen
=== CONT  TestParseFactoryFlag/worker_zero_errors
=== CONT  TestParseFactoryFlag/zero_count_errors
=== CONT  TestParseFactoryFlag/lookalike_long_flag_not_stolen
=== CONT  TestParseFactoryFlag/-f_N
=== CONT  TestParseFactoryFlag/--factory=N
=== CONT  TestParseFactoryFlag/--factory_N
=== CONT  TestParseFactoryFlag/unnumbered_lane_errors
=== CONT  TestParseFactoryFlag/non-numeric_non-lane_errors
=== CONT  TestParseFactoryFlag/negative_joined_count_errors
=== CONT  TestParseFactoryFlag/-f=N
=== CONT  TestParseFactoryFlag/absent
=== CONT  TestParseFactoryFlag/--factory=lane-7
=== CONT  TestParseFactoryFlag/long_form_bare
=== CONT  TestParseFactoryFlag/positional_flag_is_not_a_value
=== CONT  TestParseFactoryFlag/bare_-f
=== CONT  TestParseFactoryFlag/-f=lane-3
=== CONT  TestParseFactoryFlag/-f_lane-2
--- PASS: TestParseFactoryFlag (0.00s)
    --- PASS: TestParseFactoryFlag/lookalike_long_flag_not_stolen (0.00s)
    --- PASS: TestParseFactoryFlag/-f_N (0.00s)
    --- PASS: TestParseFactoryFlag/zero_count_errors (0.00s)
    --- PASS: TestParseFactoryFlag/worker_zero_errors (0.00s)
    --- PASS: TestParseFactoryFlag/--factory=N (0.00s)
    --- PASS: TestParseFactoryFlag/--factory_N (0.00s)
    --- PASS: TestParseFactoryFlag/unnumbered_lane_errors (0.00s)
    --- PASS: TestParseFactoryFlag/non-numeric_non-lane_errors (0.00s)
    --- PASS: TestParseFactoryFlag/negative_joined_count_errors (0.00s)
    --- PASS: TestParseFactoryFlag/-f=N (0.00s)
    --- PASS: TestParseFactoryFlag/absent (0.00s)
    --- PASS: TestParseFactoryFlag/--factory=lane-7 (0.00s)
    --- PASS: TestParseFactoryFlag/long_form_bare (0.00s)
    --- PASS: TestParseFactoryFlag/positional_flag_is_not_a_value (0.00s)
    --- PASS: TestParseFactoryFlag/bare_-f (0.00s)
    --- PASS: TestParseFactoryFlag/-f=lane-3 (0.00s)
    --- PASS: TestParseFactoryFlag/-f_lane-2 (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.945s
```

### 2. receipt 이력과 모델 전환

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN OPENAI_API_KEY Z_AI_API_KEY && GOCACHE=/tmp/moai-gpt-replay-research-cache go test ./internal/gateway/translate -run 'TestModelSwitch|TestSubscriptionReceiptModelSwitch|TestReceiptHistoryRejectsPublicLayoutMutation|TestReceiptHistorySpawnShapedFreshHistoryNeedsNoLineageSeeding' -count=1 -v -timeout 90s
```

원출력, exit 0:

```text
=== RUN   TestReceiptHistorySpawnShapedFreshHistoryNeedsNoLineageSeeding
--- PASS: TestReceiptHistorySpawnShapedFreshHistoryNeedsNoLineageSeeding (0.01s)
=== RUN   TestModelSwitchWithinFamilyAcceptsUnmodifiedHistory
--- PASS: TestModelSwitchWithinFamilyAcceptsUnmodifiedHistory (0.04s)
=== RUN   TestModelSwitchStrippedEnvelopeRejectionIdentifiesReason
--- PASS: TestModelSwitchStrippedEnvelopeRejectionIdentifiesReason (0.04s)
=== RUN   TestReceiptHistoryRejectsPublicLayoutMutation
--- PASS: TestReceiptHistoryRejectsPublicLayoutMutation (0.02s)
=== RUN   TestSubscriptionReceiptModelSwitchPreservesSourceBindings
--- PASS: TestSubscriptionReceiptModelSwitchPreservesSourceBindings (0.04s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	0.558s
```

### 3. bridge와 HTTP adapter

보고서 마무리 시 동일 HEAD에서 다시 실행한 원출력이다. 본문에 제시한 앞선 실행의 소요 시간과 다를 수 있다.

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN OPENAI_API_KEY Z_AI_API_KEY && GOCACHE=/tmp/moai-gpt-factory-research-cache go test ./internal/codexbridge ./internal/gateway -run 'TestToolSegmentRetainsRPCAndCompletesExactlyOnce|TestRestartPendingNeverReplays|TestResumeAttachesOwnedThreadWithoutRebuild|TestForkChildStartsNewThreadAtInheritedPrefix|TestCompactionIsANormalSummaryTurnWithZeroCompactRPCs|TestAppServerSubprocessHTTPToolContinuation|TestUpstreamRetry|TestUpstreamNoRetry' -count=1 -v -timeout 90s
```

원출력, exit 0:

```text
=== RUN   TestCompactionIsANormalSummaryTurnWithZeroCompactRPCs
--- PASS: TestCompactionIsANormalSummaryTurnWithZeroCompactRPCs (0.07s)
=== RUN   TestToolSegmentRetainsRPCAndCompletesExactlyOnce
--- PASS: TestToolSegmentRetainsRPCAndCompletesExactlyOnce (0.07s)
=== RUN   TestRestartPendingNeverReplays
--- PASS: TestRestartPendingNeverReplays (0.04s)
=== RUN   TestForkChildStartsNewThreadAtInheritedPrefix
--- PASS: TestForkChildStartsNewThreadAtInheritedPrefix (0.18s)
=== RUN   TestResumeAttachesOwnedThreadWithoutRebuild
--- PASS: TestResumeAttachesOwnedThreadWithoutRebuild (0.13s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/codexbridge	1.239s
=== RUN   TestAppServerSubprocessHTTPToolContinuation
=== RUN   TestAppServerSubprocessHTTPToolContinuation/chatgpt
=== RUN   TestAppServerSubprocessHTTPToolContinuation/apiKey
--- PASS: TestAppServerSubprocessHTTPToolContinuation (1.31s)
    --- PASS: TestAppServerSubprocessHTTPToolContinuation/chatgpt (0.78s)
    --- PASS: TestAppServerSubprocessHTTPToolContinuation/apiKey (0.53s)
=== RUN   TestUpstreamRetryRecoversSingleFiveHundred
--- PASS: TestUpstreamRetryRecoversSingleFiveHundred (0.00s)
=== RUN   TestUpstreamRetryExhaustionLabelsUpstreamOrigin
--- PASS: TestUpstreamRetryExhaustionLabelsUpstreamOrigin (0.00s)
=== RUN   TestUpstreamRetryDialFailureRecovers
--- PASS: TestUpstreamRetryDialFailureRecovers (0.00s)
=== RUN   TestUpstreamRetryDialExhaustionCarriesReason
--- PASS: TestUpstreamRetryDialExhaustionCarriesReason (0.00s)
=== RUN   TestUpstreamNoRetryAfterResponseStarted
--- PASS: TestUpstreamNoRetryAfterResponseStarted (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway	2.334s
```

### 4. hybrid tool registry

보고서 마무리 시 동일 HEAD에서 다시 실행한 원출력이다. 본문에 제시한 앞선 실행의 소요 시간과 다를 수 있다.

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN OPENAI_API_KEY Z_AI_API_KEY && GOCACHE=/tmp/moai-gpt-factory-research-cache go test ./internal/codextools -run 'TestHybridPositiveAndNoThreadRecreation|TestSixNegativeGroups|TestRevocationBeforeInvocation|TestNativeToolWireDiscriminator' -count=1 -v -timeout 30s
```

원출력, exit 0:

```text
=== RUN   TestHybridPositiveAndNoThreadRecreation
--- PASS: TestHybridPositiveAndNoThreadRecreation (0.00s)
=== RUN   TestSixNegativeGroups
=== RUN   TestSixNegativeGroups/1_collision
=== RUN   TestSixNegativeGroups/2_missing_definition
=== RUN   TestSixNegativeGroups/3_schema_change
=== RUN   TestSixNegativeGroups/4_schema_validation
=== RUN   TestSixNegativeGroups/5_binding_and_replay
=== RUN   TestSixNegativeGroups/6_native_bypass_recursion
--- PASS: TestSixNegativeGroups (0.00s)
    --- PASS: TestSixNegativeGroups/1_collision (0.00s)
    --- PASS: TestSixNegativeGroups/2_missing_definition (0.00s)
    --- PASS: TestSixNegativeGroups/3_schema_change (0.00s)
    --- PASS: TestSixNegativeGroups/4_schema_validation (0.00s)
    --- PASS: TestSixNegativeGroups/5_binding_and_replay (0.00s)
    --- PASS: TestSixNegativeGroups/6_native_bypass_recursion (0.00s)
=== RUN   TestNativeToolWireDiscriminator
--- PASS: TestNativeToolWireDiscriminator (0.00s)
=== RUN   TestRevocationBeforeInvocation
--- PASS: TestRevocationBeforeInvocation (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/codextools	0.426s
```

### 5. 중간 이력 삽입 대조 실험

기존 source 파일을 수정하지 않고 Go overlay로 연구용 테스트를 추가했다. [테스트 원본](research_replay_test.go.txt). `newClassifyFixture`는 develop의 기존 테스트 fixture를 재사용한다. 실제 Claude wire body의 재현은 아니다. 보존본은 `.txt` 확장자로 두어 저장소의 Go 패키지 검색에 포함되지 않게 했다.

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN OPENAI_API_KEY Z_AI_API_KEY && GOCACHE=/tmp/moai-gpt-replay-research-cache go test -overlay /tmp/moai-gpt-factory-schema-5fWMMB/overlay.json ./internal/gateway/translate -run '^TestResearchReminderPlacement$' -count=1 -v -timeout 30s
```

원출력, exit 0:

```text
=== RUN   TestResearchReminderPlacement
    research_replay_test.go:21: unchanged=accepted tail_reminder=accepted
    research_replay_test.go:29: mid_history_reminder=CauseChain; opaque envelopes unchanged
--- PASS: TestResearchReminderPlacement (0.06s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	0.462s
```

이 결과는 동일 reminder도 기존 전체 prefix 뒤에 추가하면 통과하고 과거 중간에 넣으면 거절된다는 것을 보여 준다. 실제 장애가 이 삽입 때문인지 Edit 결과 값 차이 때문인지는 실패 wire가 없어 미확정이다.

### 6. 설치 프로토콜 스키마

```sh
codex app-server generate-json-schema --experimental --out /tmp/moai-gpt-factory-schema-5fWMMB
```

exit 0, 표준 출력 없음. 생성 schema의 `ThreadStartParams.dynamicTools`, `ThreadForkParams.lastTurnId`, `TurnStartParams.model/effort`를 확인했다. 이는 프로토콜 선언 확인이며 실계정 실행 결과가 아니다.

### 7. 현장 로그와 기존 카드 조사

현재 큐: `moai todo list --json`의 필터된 결과에서 last_seq=853, items=147. 이는 반환된 현재 카드 스냅샷이지 전체 보관 카드 수가 아니다.

t851의 `tools/envelope_compare.py`를 읽은 뒤 실행했다. 현재 보존 로그에서 7개 envelope가 일치하는 결과를 재확인했다. 입력 원문·토큰·도구 ID는 이 보고서에 복사하지 않았다. 비교 결과는 transcript에서 추출한 envelope 범위이며 실패 Messages 전체 wire의 동등성을 의미하지 않는다.

family `2e343aa0-5a47-4c1f-bc7e-c3046531713b`의 native debug 파일에서 다음 네 시각의 API 400을 읽었다. 2026-09-13 17:28:14.809Z, 17:30:15.393Z, 17:30:22.145Z, 17:30:30.772Z. 오류는 CauseChain 계열이다. t851 최신 verdict의 두 경쟁 가설을 유지했다.

진행 중 카드의 상태를 변경하거나 코드를 합치지 않았다. t654의 `2113c1e14`와 t851의 `a5ccaca5a`는 `git merge-base --is-ancestor <commit> c9ceff175`에서 exit 1이었다. App Server 구성요소의 develop 포함과 실제 런처 cutover 완료는 서로 다른 주장이다.

### 8. HTML 표시 확인

`pandoc ... --standalone --template report.template.html ... -o report.html`로 생성한 후, 독립 브라우저 세션 `gpt-factory-report`에서 로컬 파일을 열었다. 문서 생성과 제품 동작은 구분한다.

브라우저 DOM 검사에서 관련 필드를 추출한 값:

```json
{"width":1440,"scrollWidth":1440,"sections":12,"tables":6,"mermaidSvg":true,"fallbackVisible":false}
{"width":390,"scrollWidth":390,"sections":12,"mermaidSvg":true}
```

로컬 링크와 파일 검사 출력의 값:

```json
{"htmlBytes":49261,"localLinks":16,"missing":[],"placeholderRemaining":false}
```

`agent-browser ... errors`는 exit 0, 출력 없음이었다. 데스크톱·모바일 스크린샷을 직접 열어 상단 글자·카드·도식 배치를 확인했다. 비교 표와 도식은 좁은 화면에서 개별 영역을 가로로 스크롤하도록 구성했다. CDN 실패 때 표시할 텍스트 대안은 코드에 포함했지만, 네트워크 완전 차단 상태에서 별도 실행하지는 않았다.

## Gaps

- 위 subprocess의 chatgpt/apiKey 케이스는 fake App Server fixture다. OAuth 및 API 인증 성공을 실측한 것이 아니다.
- 전체 suite, 실제 factory N lane, live model switch/resume/compact/native fork는 미실행이다.
- 실제 실패 요청 wire 전체가 없어 정확한 최초 변형 위치·값과 인과관계를 확정하지 않았다.
- 외부 라이브러리는 공식 문서를 조사했으며 설치·성능 비교 실험은 하지 않았다.
- 약관상 양사 보증, 운영 인수, remote 배포는 범위 밖이다.

## Residual-risk

선택 테스트는 해당 fixture와 경로만 검증한다. 동일 기능 이름이 있어도 production 배선·실제 모델의 tool behavior·Claude의 버전별 transcript 변화는 추가 검증이 필요하다. 인증 정보와 비공개 prompt는 근거물에서 제외했다.
