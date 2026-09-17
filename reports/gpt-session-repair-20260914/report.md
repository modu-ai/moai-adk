# moai gpt 502 원인 분석 및 수정 검증 보고서

> **후속 검증·배포 완료 안내:** 이 문서는 프로필 락 해제 전의 기록이다. 사용자가 기존 세션을 종료한 뒤 실제 구독 검증과 추가 수정, rc.11 로컬 설치를 수행했다. 현재 상태는 [실제 구독 검증 및 로컬 배포 보고서](live-verification.md)를 기준으로 한다. 아래의 이전 미완료 기록은 당시 증거로 보존한다.

## Claim

**502를 유발할 수 있는 대화 식별·취소·도구 이벤트 처리 문제를 재현하여 수정했으며, 영향 패키지의 race 검사와 실제 Claude Code를 사용하는 로컬 통합 검사를 통과했다. 실제 OpenAI 구독을 연결한 최종 운영 검증과 설치본 교체는 아직 완료하지 않았다.**

기존 rc.11 세션이 인증 프로필의 독점 락을 계속 점유하고 있다. 이 세션을 강제로 종료하거나 인증 파일·실패한 대화 상태를 삭제하지 않았다. 따라서 “Claude Code의 모든 기능을 GPT로 빠짐없이 사용할 수 있다” 또는 “GPT 생성 속도까지 검증했다”는 완료 판정은 내리지 않는다.

### 로그와 코드의 상관관계

분석한 원본:

`/Users/goos/.moai/state/gateway-conversations/families/48a351b8-4a08-48b5-887d-e5ceae71d74b/native/debug/48a351b8-4a08-48b5-887d-e5ceae71d74b.txt`

| 관측 | 수정 전 처리 | 영향 및 수정 |
|---|---|---|
| 13:40:24.947 UTC에 custom agent 요청, 13:40:32.579에 첫 502 | `server.go`가 실제 Claude의 agent/session 헤더를 버리고, prepare는 title 외 요청에 같은 owner 사용 | 실제 클라이언트에서 캡처한 session/agent 헤더를 검증해 main·child를 분리 |
| 13:40:54.833에 `agent_summary` 요청 | summary도 child와 같은 owner 사용 | 실제 summary는 child와 agent ID·system prompt가 같음을 PTY로 확인. 관측한 native 지시문을 별도 snapshot으로 분리 |
| 13:42:20.150에 로컬 취소, 이후 502 반복 | 취소 후 `failed`가 저장되고, 일반 오류가 모두 502로 변환 | 정확한 turn의 종료 확인과 미응답 도구 부재를 조건으로 복구. 불확실한 실행은 자동 재전송하지 않음 |
| 영속 상태가 `Phase: failed`, 모델 `gpt-5.6-sol` | 실패한 owner의 후속 요청도 반복 거부 | 원인별 고정 오류 코드와 비재시도 HTTP 응답, 개인 오류 로그 추가 |

최초 502의 내부 예외 원문은 기존 구현에서 폐기되었다. 따라서 최초 요청이 정확히 어느 예외 분기에서 실패했는지는 확정할 수 없다. 대화 owner 충돌과 취소 경합은 이번 회귀 테스트에서 별도로 재현했다.

### 구현 범위

| 항목 | 구현 및 관측 범위 | 최종 상태 |
|---|---|---|
| 메인·서브에이전트 분리 | 실제 Claude header 캡처, 인증된 family 내 ID 검증, native Agent 왕복 | 로컬 통합 PASS |
| 진행 요약 | 실제 PTY 요청 형태 캡처, child와 다른 snapshot owner | 형태 관측·회귀 PASS, 구독 통합 대기 |
| `/compact` | 요약 생성과 요약 이후 문맥 분리, 요약 블록 digest 유지 | 형태 관측·회귀 PASS, 구독 통합 대기 |
| 새 child의 문맥 | 이전 public assistant/tool 기록을 실행 지시가 아닌 데이터로 전달, 이미지 유지 | 회귀 PASS |
| 스트리밍 | 텍스트 도착 즉시 전달, 도구 상태 저장 및 완료 확인 후 성공 종료 | 회귀·native 통합 PASS |
| 취소와 계속하기 | 대기 요청 취소, 늦은 완료 이벤트, 큐 안의 완료 이벤트 보존 | race PASS |
| 병렬 도구 | 도구 호출 사이의 정보 알림 수용, 후속 텍스트 순서 보존 | race PASS |
| 이미지 | PNG/JPEG/GIF/WebP base64 및 HTTPS source 검증, 이미지 첨부·도구 반환·상속 문맥 전달 | 회귀 PASS, 실제 이미지 해석 대기 |
| JSON schema | 검증된 출력 schema를 native `turn/start`로 전달 | 회귀 PASS |
| 모델 변경 | idle 상태에서만 모델 변경 허용 | 회귀 PASS |
| 지시문 변경 | digest 비교, idle 상태에서 갱신, 같은 지시문은 추가 RPC 없음 | 회귀 PASS, 실제 completed thread 효과 대기 |
| 사용량 | 관측된 누적 증가분을 입력·출력·캐시 표준 필드로 전달 | 회귀·native 수신 PASS, 실제 이벤트 시점 검증 대기 |
| 여러 세션 | 단일 공식 App Server와 private Unix WebSocket 연결, 연결별 수명 분리 | 실제 로컬 전송 PASS, 동시 모델 호출 대기 |
| 인증 상태 | 공식 `codex login status`, 실행 중 락과 충돌하지 않는 조회 | 실제 사용자 프로필 PASS |
| 오류 진단 | 고정 cause/route/digest/시각만 기록, 0600, 1 MiB 한도 | 회귀 PASS |

모델은 기존 네 개만 유지한다: `gpt-6-astra`, `gpt-5.6-sol`, `gpt-5.6-terra`, `gpt-5.6-luna`. 새 plan/run 역할별 모델 표나 사용자가 요청하지 않은 agent preset은 추가하지 않았다. 위의 대화 분리는 Claude가 실제로 보내는 독립 요청을 처리하기 위한 내부 구분이다.

## Evidence

### 최종 영향 패키지 검사

```sh
unset MOAI_GPT_LIVE MOAI_CLAUDE_INTEGRATION MOAI_CODEX_TRANSPORT_LIVE MOAI_GPT_TEST_BINARY && go test -race ./internal/codexapp ./internal/codexbridge ./internal/gateway ./internal/gateway/translate -count=1
```

```text
ok  	github.com/modu-ai/moai-adk/internal/codexapp	3.827s
ok  	github.com/modu-ai/moai-adk/internal/codexbridge	16.565s
ok  	github.com/modu-ai/moai-adk/internal/gateway	13.757s
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	5.241s
```

전체 출력: [race 검사](evidence/moai-gpt-final-race.log).

### 실제 Claude Code 및 공식 App Server 로컬 검사

```sh
unset MOAI_GPT_LIVE && MOAI_CLAUDE_INTEGRATION=1 MOAI_CODEX_TRANSPORT_LIVE=1 MOAI_GPT_TEST_BINARY=/tmp/moai-gpt-session-repair go test ./internal/cli ./internal/codexapp -run '^(TestManagedGPT.*|TestGPTProductionAppServerWiring|TestShared.*|TestClaudeCodeProductionGatewayAgentRoundTrip|TestOfficialIdleDeveloperInstructionOverride)$' -count=1 -v
```

핵심 출력:

```text
=== RUN   TestClaudeCodeProductionGatewayAgentRoundTrip
    gpt_appserver_claude_integration_test.go:112: actual Claude Code -> production gateway -> App Server fixture: main=4 agent=1 threads=2 toolResults=1 children=1 result=MAIN_GATEWAY_OK
--- PASS: TestClaudeCodeProductionGatewayAgentRoundTrip (0.50s)
--- PASS: TestSharedGPTSupervisorTerminatesOfficialOwner (1.58s)
ok  	github.com/modu-ai/moai-adk/internal/cli	3.689s
--- PASS: TestSharedOfficialTransport (0.16s)
--- PASS: TestSharedOfficialLeaseInheritance (0.05s)
ok  	github.com/modu-ai/moai-adk/internal/codexapp	0.669s
```

첫 검사는 실제 설치된 Claude Code가 production gateway·bridge·SSE·공유 전송 경로를 사용한다. 마지막 App Server peer만 fixture이다. 두 번째 종류의 검사는 실제 설치 Codex의 로컬 연결·수명·락을 사용하며 모델 생성은 하지 않는다. **0.50초와 0.16초는 GPT 추론 속도가 아니다.**

동일 실행의 명시적 SKIP:

```text
--- SKIP: TestManagedGPTLiveSimpleTurn (0.00s)
--- SKIP: TestManagedGPTLiveResume (0.00s)
--- SKIP: TestSharedGPTLiveConcurrentToolRouting (0.00s)
--- SKIP: TestOfficialIdleDeveloperInstructionOverride (0.09s)
```

마지막 진단은 생성 이력이 없는 빈 thread에 대해 baseline resume와 override resume가 모두 RPC `-32600`으로 거부되었다. 실제 생성 이후 thread에서 override가 동작하는지는 이 진단으로 증명하지 못했다.

전체 출력: [native 검사 및 SKIP](evidence/moai-gpt-final-native.log).

### 빌드·정적 검사·실제 상태 조회

```sh
go build -o /tmp/moai-gpt-session-repair ./cmd/moai
go vet ./internal/codexapp ./internal/codexbridge ./internal/gateway ./internal/gateway/translate ./internal/cli
git diff --check
node --check internal/cli/testdata/claude_identity_probe.mjs
```

모두 출력 없이 exit 0. [vet 출력 파일](evidence/moai-gpt-final-vet.log)은 빈 파일이다.

```text
$ /usr/bin/time -p /tmp/moai-gpt-session-repair gpt status
GPT: logged in using ChatGPT (managed profile)
real 0.12
user 0.02
sys 0.02
```

빌드 SHA-256:

```text
3cd85c9743dc616d4e4a160e05d5366eb9fb909ef4aae5b08e0f85c87b4dd7f6  /tmp/moai-gpt-session-repair
5295ffc469ea8e75130b3b6caac702d14b9641eae0509cafc7925dd6be706131  /Users/goos/go/bin/moai
```

두 파일은 다르며 설치 경로의 rc.11을 교체하지 않았다.

### 현재 운영 검증을 막는 상태

기존 PID 확인:

```text
  PID  PPID COMM
94098 94079 /Users/goos/go/bin/moai
94100 94098 /Users/goos/.local/bin/codex
94079 90353 claude
```

`.moai-appserver.lock`를 읽기 전용으로 연 뒤 nonblocking flock을 시도한 결과:

```text
profile_lock=busy
```

락을 삭제하거나 기존 프로세스를 강제 종료하지 않았다.

## Baseline-attribution

- 기준 HEAD: `b45c813493751106cd6619dce60fba1c61e70583`.
- 작업 브랜치: `WT-gpt-session-repair`.
- 작업 공간: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/gpt-session-repair`.
- 세션: `01a09ebc-fbd0-7683-ac83-b9d315338def`.
- 최종 검사는 담당자들의 파일 수정이 멈춘 뒤 이 작업 트리에서 실행했다.
- 변경 소스 및 테스트 38개 파일의 [SHA-256 목록](evidence/source-sha256.txt)을 보관했다.
- 기준 트리의 bridge coverage는 이번 실행에서 `80.4%`였다. 최종 전체 coverage 수치를 새로 측정하지 않았으므로 최종 85% 달성을 주장하지 않는다.
- 수정 전 gateway 검사에서는 `TestGPTAppServerLifecycleAndHistoryAttribution`의 오류 원인 반환·기록 관련 테스트들이 실패했다. 최종 gateway 전체 race 검사에서 통과했다. [수정 전 출력](evidence/moai-gpt-repair-baseline.log)
- primary checkout과 develop의 다른 작업은 합치거나 덮어쓰지 않았다. 현재 수정은 이 작업 트리에 있으며 commit·merge·push·설치 배포를 하지 않았다.

## Gaps

1. 실제 구독의 네 모델 동시 tool roundtrip, 새 바이너리의 실사용 Agent·summary·이미지·모델 변경·재개·지시문 갱신, 첫 토큰 및 완료 지연 측정이 남아 있다.
2. 기존 rc.11 세션을 종료해 락이 해제된 뒤 새 공유 서버를 시작해야 한다. 새 서버가 실행된 상태에서 아래 준비된 검사를 수행할 수 있다.

   ```sh
   MOAI_GPT_LIVE=1 go test ./internal/cli -run '^TestSharedGPTLiveConcurrentToolRouting$' -count=1 -v
   ```

   이 검사는 두 client·별도 engine/store에서 네 모델을 두 쌍씩 동시에 실행하고, echo 인자·결과가 섞이지 않는지와 한 client 종료 후 다른 client의 후속 응답을 확인한다.
3. main/summary/compact 문구 식별은 실제 관측한 Claude Code 2.1.270 기준이다. 모든 향후 버전 또는 모든 auto-compact tail 형태를 검증한 것은 아니다.
4. Windows는 기존 독점 stdio 경로를 유지하며 병렬 세션 완료 범위에 포함하지 않는다.
5. 이미 실패한 기존 대화의 불확실한 도구 실행을 자동 재전송하거나 강제로 복구하지 않는다. 해당 기록의 삭제·리셋도 하지 않았다.
6. 제품상 폐기된 kanban의 `-k`·legacy backend env 및 factory의 `-p` 처리 계약 차이는 기존 GAP이다. 수정 전 별도 `TestGPTFactoryAndRetiredKanbanContract` 실패를 확인했으며, 이번 502 수정에서 해당 런처 정책 제거까지 수행하지 않았다. [기준 출력](evidence/moai-gpt-cli-baseline.log)
7. 전체 저장소 CI 및 최종 설치 바이너리 검증은 수행하지 않았다. 기존 설치본은 rc.11 그대로다.

## Residual-risk

- upstream이 usage를 보내지 않거나 resume의 누적 기준이 없으면 정확한 사용량을 표시하지 못할 수 있다. API 필수 필드의 0을 실제 관측된 사용량 0으로 해석하면 안 된다.
- 미응답 도구가 남거나 native turn ID 자체가 확인되지 않은 취소는 자동 복구 대상이 아니다. 중복 실행을 방지하기 위한 제한이다.
- 완전히 동일한 compact summary 블록은 같은 idle context를 선택할 수 있다. pending context를 입력만으로 덮어쓰는 요청은 거부한다.
- ephemeral owner는 engine 메모리에서 제거하지만 native `thread/unsubscribe`는 적용하지 않았다. 동일 summary가 다시 연결되는 시점과의 경합을 별도로 검증해야 한다.
- 공유 서버는 첫 gateway 종료 후에도 유지된다. 실제 모델 작업으로 생성되는 추가 하위 프로세스까지 포함한 종료·락 반환은 구독 실행 후 추가 확인이 필요하다.
- 로컬 회귀·프로토콜 통합 성공은 실제 모델의 품질·용량·응답 지연 또는 Claude Code 전체 기능의 운영 동등성을 보증하지 않는다.
