# moai gpt 제한 오류·세션 조회 지연 수정

## Claim

**재현된 이벤트 폭주·실패 상태 경합과 세션 조회 지연 경로를 수정하고, 최종 코드의 실제 130초 도구 대기 검사 및 rc.11 로컬 설치본 도구 왕복을 통과했다. 설치 경로는 `/Users/goos/go/bin/moai`이다.**

claude-code-proxy의 최신 소스를 확인하고, MoAI의 App Server 구조에 맞춰 이벤트 버퍼 처리를 수정했다. 세션 조회 전에 실행되던 PR 정리 작업도 제거했다. **원래 장애 순간의 정확한 이벤트 폭주는 기록되어 있지 않으므로, 이번 수정이 원래 400의 유일한 원인을 확정했다는 뜻은 아니다.**

### 참고한 원본 구현

조사 기준은 `raine/claude-code-proxy`의 커밋 `ba8cd70dd52b90f2eac893b759efcc36ad96500b`이다. 기존 로컬 사본은 수정된 파일이 있어 사용하지 않고 별도 디렉터리에 새로 복제했다.

| 원본에서 확인한 처리 | MoAI에 적용한 방식 |
|---|---|
| [64개 채널과 비동기 전송](https://github.com/raine/claude-code-proxy/blob/ba8cd70dd52b90f2eac893b759efcc36ad96500b/src/providers/codex/websocket.rs#L1052): 채널이 가득 차면 소비될 때까지 대기 | 공유 RPC reader를 막지 않도록 대화별 바이트 제한 mailbox 사용 |
| [이벤트를 전송하며 종료 이벤트를 구분](https://github.com/raine/claude-code-proxy/blob/ba8cd70dd52b90f2eac893b759efcc36ad96500b/src/providers/codex/websocket.rs#L2315) | 텍스트·도구·사용량·완료 이벤트의 순서를 유지하고 실제 상한 초과는 해당 owner에만 격리 |
| [진행 알림과 의미 있는 출력을 구분](https://github.com/raine/claude-code-proxy/blob/ba8cd70dd52b90f2eac893b759efcc36ad96500b/src/providers/codex/translate/live_stream.rs#L135) | 기존에도 공개 응답에 사용하지 않던 명시적 reasoning/progress 알림만 제외. 미지의 알림과 ID가 있는 RPC 요청은 보존 |

원본은 직접 upstream WebSocket을 처리하고 MoAI는 공식 App Server의 공유 RPC 연결을 처리한다. 따라서 원본의 기다리는 전송을 그대로 이식하면 다른 대화의 RPC 응답까지 막을 수 있다. Rust 코드를 복사하거나 MoAI를 Rust로 전환하지 않았으며, 확인한 처리 원리를 Go 표준 라이브러리로 구현했다.

### 수정 사항

1. 이벤트 개수 64개만으로 전체 engine이 종료되던 처리를 제거했다. mailbox는 JSON 데이터와 이벤트별 메모리 오버헤드를 함께 계산하며, 기본 상한은 8MiB이다. 무제한 버퍼가 아니다.
2. 실제 mailbox 상한을 넘으면 해당 대화의 실패를 영속화하고 해당 turn에만 제한 시간이 있는 interrupt를 요청한다. pending·queued·현재 RPC 요청을 정리하고 다른 owner는 계속 동작한다.
3. 이미 선택된 tool boundary나 완료 이벤트가 비동기 실패를 덮어쓰지 못하도록 실패 상태를 우선한다. 입력이나 도구를 자동 재실행하지 않는다.
4. mailbox 초과는 `appserver_event_queue_bytes_exceeded`로 HTTP와 비공개 오류 로그에서 구분한다. 기존 일반 제한 오류의 호환 처리는 유지한다.
5. `moai session list`의 PR 정리 호출을 제거했다. `session register`의 기존 정리 동작이나 사용자 설정은 바꾸지 않았다. root 초기화의 config cache 등까지 쓰기 없는 명령으로 바꾼 것은 아니다.

`moai-fix` 절차에 따라 수정 전 재현, 파일별 구현 분담, 독립 검토와 최종 재검증을 수행했다.

## Evidence

### 원래 장애 관측

가족 ID: `641a886c-c78c-41f9-8b86-478265cd1a52`.

```text
2026-09-14T15:04:10.591Z ... tool_dispatch_start tool=Bash
2026-09-14T15:06:10.702Z ... tool_dispatch_end tool=Bash ... durationMs=120111
2026-09-14T15:06:10.762Z ... API error ... 400 ... appserver_limit_exceeded
2026-09-14T15:07:17.828Z ... API error ... 400 ... appserver_limit_exceeded
```

도구는 `moai session list --json`이며 120초 후 완료가 아닌 background 전환 메시지를 반환했다. 해당 결과는 332바이트였다. 실제 프로세스의 자식 `gh`, 여러 worktree의 PR 정리 로그, `auto_cleanup: true`를 확인했다.

### 수정 전 실패 재현

도구 결과를 기다리는 동안 256개의 text·usage·reasoning 묶음을 넣는 회귀 테스트:

```text
--- FAIL: TestLongToolWaitRetainsBurstWithoutPoisoningOtherOwners (0.06s)
    event_queue_test.go:30: long tool wait poisoned shared reader: app server bridge limit exceeded
FAIL
FAIL github.com/modu-ai/moai-adk/internal/codexbridge 0.431s
```

실패와 이미 선택된 경계가 만나는 상태를 결정적으로 구성한 회귀 테스트:

```text
--- FAIL: TestOverflowFailureWinsOverAlreadySelectedToolBoundary (0.09s)
    event_queue_test.go:140: terminal overflow overwritten: done=true responses=1 err=<nil>
FAIL
FAIL github.com/modu-ai/moai-adk/internal/codexbridge 0.463s
```

세션 조회의 기존 정리 호출 재현:

```text
--- FAIL: TestSessionListNeverRunsPRCleanup/cleanup-on
    session_worktree_prmerge_test.go:544: read-only list invoked cleanup/external probes 1 times
```

위 재현은 담당자의 동일 작업 트리 실행 결과다. 원래 실사용의 정확한 이벤트 열을 재생한 것은 아니다.

### 최종 영향 범위 race 검사

```sh
unset MOAI_GPT_LIVE MOAI_GPT_LONG_WAIT_LIVE MOAI_CLAUDE_INTEGRATION MOAI_CODEX_TRANSPORT_LIVE MOAI_GPT_TEST_BINARY && go test -race ./internal/codexapp ./internal/codexbridge ./internal/gateway ./internal/gateway/translate ./internal/gateway/conversation -count=1
```

```text
ok  	github.com/modu-ai/moai-adk/internal/codexapp	3.606s
ok  	github.com/modu-ai/moai-adk/internal/codexbridge	17.089s
ok  	github.com/modu-ai/moai-adk/internal/gateway	13.619s
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	5.575s
ok  	github.com/modu-ai/moai-adk/internal/gateway/conversation	5.202s
```

[전체 출력](evidence/race-final.log). 마지막 실패 우선순위 guard까지 포함한 최종 동결 제품 소스에서 실행했다.

### 최종 CLI·공식 로컬 전송 검사

```sh
unset MOAI_GPT_LIVE MOAI_GPT_LONG_WAIT_LIVE && MOAI_CLAUDE_INTEGRATION=1 MOAI_CODEX_TRANSPORT_LIVE=1 MOAI_GPT_TEST_BINARY=/tmp/moai-gpt-queue-repair.HTLZvH/moai go test ./internal/cli ./internal/codexapp -run '^(TestManagedGPT.*|TestGPTProductionAppServerWiring|TestShared.*|TestClaudeCodeProductionGatewayAgentRoundTrip|TestSessionList.*|TestPRMergeCleanup.*|TestParseWorktreeList.*)$' -count=1 -v -timeout=120s
```

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	4.793s
ok  	github.com/modu-ai/moai-adk/internal/codexapp	1.242s
```

[전체 출력과 SKIP](evidence/cli-final.log). 이 실행은 구독 전용 검사를 SKIP하며 아래 별도 실행과 구분한다.

### 실제 구독 회귀 검사

```sh
MOAI_GPT_LIVE=1 go test ./internal/cli -run '^TestSharedGPTLive(ConcurrentToolRouting|InstructionsModelResumeAndImage|CancelThenFollowup)$' -count=1 -v -timeout=240s
```

```text
--- PASS: TestSharedGPTLiveInstructionsModelResumeAndImage (10.64s)
--- PASS: TestSharedGPTLiveConcurrentToolRouting (16.34s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	36.220s
```

[전체 출력](evidence/live-final.log). 실제 네 모델의 도구 결과 분리, 지시문 변경, 모델 변경, 재개, 이미지 입력, 취소 후속 요청을 검증했다.

### 130초 대기 검사의 해석

**수정 전 실제 130초 도구 대기 검사도 통과했다.** 따라서 대기 시간만으로 원래 장애가 재현되었다고 해석할 수 없다.

```text
delayed_result_elapsed=2m10.001s
done=true total=2m16.765s result="LONG_WAIT_COMPLETE"
--- PASS: TestSharedGPTLiveLongToolWait (136.77s)
```

[수정 전 실제 출력](evidence/longwait-before.log). 두 도구를 요청했으나 대기 구간에는 하나만 전달됐다.

최종 동결 코드에서 실행:

```sh
MOAI_GPT_LONG_WAIT_LIVE=1 go test ./internal/cli -run '^TestSharedGPTLiveLongToolWait$' -count=1 -v -timeout=240s
```

```text
delayed_result_elapsed=2m10.001s
done=true calls_returned=3 max_segment_pending=1 total=2m16.601s result="LONG_WAIT_COMPLETE LONG_WAIT_A LONG_WAIT_B LONG_WAIT_C"
--- PASS: TestSharedGPTLiveLongToolWait (136.61s)
PASS
ok github.com/modu-ai/moai-adk/internal/cli 137.490s
```

[최종 실제 출력](evidence/longwait-final.log). 실행 전후 engine와 test 소스 해시가 동일했다. 세 결과는 모두 반환됐으나 한 segment의 최대 pending은 1이었으므로 세 도구의 동시 pending을 검증했다고 주장하지 않는다. histogram은 공유 client가 받은 전체 알림이며 다른 검증 thread의 알림도 포함된다.

### 실제 성능 관측

수정 후보의 실제 프로젝트 세션 조회:

```sh
/usr/bin/time -p /tmp/moai-session-list-pure session list --json
```

```text
첫 실행: JSON array, 29 entries; real 1.03 / user 0.04 / sys 0.03
두 번째: JSON array, 29 entries; real 0.05 / user 0.04 / sys 0.01
```

[첫 측정](evidence/session-list-first.time), [두 번째 측정](evidence/session-list-second.time). 원래 명령은 120초에서 background로 전환되었고 실제 최종 완료 시간은 측정하지 않았다. 따라서 정밀한 배속 비교 수치는 산출하지 않는다.

최종 후보의 실제 Claude Code Agent 왕복은 `QUEUE_GATEWAY_OK QUEUE_CHILD_OK`, 자식 1개 완료·실패 0, exit 0이었다. 첫 콘텐츠 4.056초, 전체 CLI 14.37초였다. 다른 실제 모델 검사를 동시에 수행했으며 통제된 성능 비교가 아니다. [후보 Agent 출력](evidence/agent-final.log)

```sh
go vet ./internal/codexapp ./internal/codexbridge ./internal/gateway ./internal/gateway/translate ./internal/gateway/conversation ./internal/cli
git diff --check
```

exit 0, 출력 없음. [vet 기록](evidence/vet.log)

### 설치본 최종 확인

기존 설치 파일의 SHA-256이 예상 값 그대로인 것을 확인하고 백업한 다음, 설치 디렉터리 안의 임시 파일을 rename하여 교체했다. 설치 파일은 후보와 동일한 SHA-256이다.

```sh
/usr/bin/time -p /Users/goos/go/bin/moai session list --json
```

```text
JSON array, 29 entries
real 0.05
user 0.04
sys 0.01
```

[설치본 세션 조회 시간](evidence/installed-list.time)

```sh
unset CLAUDECODE ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN ANTHROPIC_BASE_URL && /usr/bin/time -p /Users/goos/go/bin/moai gpt --model gpt-5.6-sol -- --print --output-format json --tools Bash --allowedTools Bash --max-turns 3 'Use Bash exactly once to run printf QUEUE_INSTALLED_TOOL_OK. Then reply exactly QUEUE_INSTALLED_OK after seeing the tool output. Do not run any other commands.'
```

실제 native transcript에서 `Bash`의 `printf QUEUE_INSTALLED_TOOL_OK` 호출과 같은 문자열의 성공 tool_result를 확인했다. 최종 JSON의 `result=QUEUE_INSTALLED_OK`, `is_error=false`, `num_turns=2`, 명령 exit 0이었다. 첫 콘텐츠 4.282초, 전체 CLI 6.74초였다. [설치본 실제 GPT 도구 왕복](evidence/installed-tool.log)

## Baseline-attribution

- 작업 공간: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/gpt-session-repair`
- 브랜치/기준 HEAD: `WT-gpt-session-repair` / `b45c813493751106cd6619dce60fba1c61e70583`
- 세션: `ba07dea0-d7ff-427c-b672-fce37365f61c`
- [최종 수정 소스 SHA-256](evidence/source-sha256.txt)
- 설치 버전: `v3.2.0-rc.11`
- 설치 Build ID: `v3.2.0-rc.11-b45c81349-event-queue-repair-dirty`
- 설치 SHA-256: `4ae0134067a105654a27b1f1fee2278f2a5c7652286c5d5d3184718a2ba413c0`
- 빌드 시각: `2026-09-14T15:20:28Z`
- 기존 설치본 백업: `/Users/goos/.moai/releases/rc11-event-queue-backup.CCvwsQ/moai-before-event-queue-repair`
- 기존 설치본 SHA-256: `108487159da3321fee29ca8febe7c770efcba9c91d76ad4bfda069d030bd5f89`

제품 코드는 Go이며 추가 외부 의존성을 넣지 않았다. 이전 수정과 이번 수정은 작업 트리에 있으며 commit·develop merge·push는 하지 않았다.

## Gaps

- 원래 400 순간의 정확한 이벤트 구성은 알 수 없다. 새 진단 코드는 향후 mailbox 초과를 일반 제한 오류와 구별한다.
- synthetic burst는 텍스트·사용량 보존과 다른 owner의 진행을 검증하지만 실제 provider가 그 순서로 이벤트를 보냈다는 증거는 아니다.
- 실제 모델이 세 도구를 요청대로 동시에 보류하지 않으면 해당 실행을 병렬 대기 검증으로 주장하지 않는다.
- 전체 repository CI, 장시간 다중 lane 부하와 메모리 RSS 상한 측정, 모든 Claude Code 기능의 동등성 검증은 포함하지 않았다.
- 기존 실패 대화를 자동 복구하거나 불확실한 도구를 다시 실행하지 않는다. 새 gateway 프로세스에서 수정본을 사용해야 한다.

## Residual-risk

- 상한을 넘는 진짜 이벤트 부하는 계속 안전하게 실패한다. 이번 변경은 무제한 buffering이나 실패 은폐가 아니다.
- interrupt는 제한 시간이 있는 최선의 요청이다. native 연결 자체가 고장 나면 종료를 보장할 수 없다.
- 상한 초과 이벤트가 turn ID를 처음 알려 주는 거대한 `turn/started`이고 RPC 응답도 없으면 turn ID 확보·종료에 추가 한계가 있을 수 있다.
- 명시적 진행 알림 필터는 기존 bridge가 공개 출력으로 사용하지 않던 이벤트에 한정된다. reasoning 표시 기능을 새로 구현한 것은 아니다.
- Claude가 표시하는 첫 바이트와 첫 콘텐츠·전체 완료 시간은 서로 다르며, 단일 측정값은 응답 시간 보장이 아니다.
