# GPT 내부 에이전트 경계 수정 — rc.11

## Claim

Codex 내부 에이전트 생성 차단 설정을 보완했다. Claude Code의 동적 `Agent` 도구와 서브 에이전트는 유지한다. 등록되지 않은 thread의 요청을 거부하는 보안 검사는 변경하지 않았다.

## Evidence

### 장애와 원인

장애 family: `e8e3e15b-adb0-41dd-851c-73434717ad51`.
App Server 부모 thread: `01a0a08b-e817-7d41-9d01-dcbd1daddfcd`.
내부 자식 thread: `01a0a08c-7a5b-7a92-ba15-3cfec82155ba`.

실제 debug 로그와 읽기 전용 SQLite 조회에서 다음 순서를 관측했다(UTC).

```text
15:32:36 parent: collaboration spawn_agent
15:32:39 parent: collaboration wait_agent
15:32:43 child: tools.Bash
2026-09-14T15:32:43.388Z [ERROR] API error (attempt 1/11): 400 400 {"error":{"message":"appserver_protocol_error","type":"invalid_request_error"},"type":"error"}
```

`internal/codexbridge/engine.go`의 reader는 등록되지 않은 thread의 ID가 있는 RPC 요청을 받으면 `ErrProtocol`로 엔진을 종료한다. 원래 RPC wire envelope는 확보하지 않았으므로 이 결론은 실제 로그와 코드 경로의 상관관계이며 wire 재생 검증과 구분한다.

로컬 공식 Codex 소스 `/tmp/openai-codex-research`의 기준 커밋은 `b966240bea4210711884d488375bfed8c0297f54`이다. `codex-rs/core/src/config/mod.rs:1555`의 우선순위는 V2 feature 강제 설정, `agents.enabled=false`, 모델 메타데이터, feature fallback 순이다. 관리 프로필의 실제 모델 캐시에는 astra·sol·terra가 v2, luna가 v1로 기록되어 있었다.

따라서 feature 두 개를 false로 설정하는 것만으로는 충분하지 않았다.

### 수정

- `internal/codexapp/client.go`: 시작 옵션 `agents.enabled=false` 추가.
- `internal/codexbridge/engine.go`: thread/start 및 thread/resume의 config에 `agents.enabled=false`, `features.multi_agent=false`, `features.multi_agent_v2=false` 명시.
- 새 단위 테스트 2개 파일과 실제 구독 검증 파일 추가.
- 기존 큐 수정, Claude dynamicTools, unknown-thread 보안 검사 유지.

### 수정 전 실패와 수정 후 검사

```sh
go test ./internal/codexapp ./internal/codexbridge -run 'TestAppServerArgsDisableNativeAgents|TestThreadConfigDisablesNativeAgentsOnStartAndResume' -count=1
```

수정 전 출력:

```text
missing native agent override "agents.enabled=false"
thread/start missing false config override "agents.enabled": map[]
thread/resume missing false config override "agents.enabled": map[]
FAIL
```

수정 후 동일 필터 count=3:

```text
ok github.com/modu-ai/moai-adk/internal/codexapp 0.367s
ok github.com/modu-ai/moai-adk/internal/codexbridge 0.953s
```

영향 패키지 race 검사:

```sh
unset MOAI_GPT_LIVE MOAI_GPT_LONG_WAIT_LIVE MOAI_CODEX_TRANSPORT_LIVE CODEXAPP_LIVE_BINARY && go test -race ./internal/codexapp ./internal/codexbridge -count=1
```

```text
ok github.com/modu-ai/moai-adk/internal/codexapp 3.662s
ok github.com/modu-ai/moai-adk/internal/codexbridge 15.982s
```

`go vet ./internal/codexapp ./internal/codexbridge` 및 `git diff --check`: exit 0. 위 RED/GREEN/race는 구현 담당자가 동일 작업 트리에서 실행했다.

실제 구독 검증:

```sh
MOAI_GPT_LIVE=1 go test ./internal/cli -run '^TestSharedGPTLiveNativeAgentIsolationPreservesDynamicAgent$' -count=1 -v -timeout=120s
```

```text
model=gpt-5.6-sol phase=idle agents.enabled=false dynamic_Agent_returned=true first_tool=3.736s total=6.75s result="DYNAMIC_AGENT_ISOLATION_OK"
--- PASS: TestSharedGPTLiveNativeAgentIsolationPreservesDynamicAgent (6.75s)
ok github.com/modu-ai/moai-adk/internal/cli 7.652s
```

위 검사는 실제 RPC 설정 전달과 GPT의 동적 Agent 호출을 검증하되 Agent 결과는 합성한다.

별도로 최종 후보 실행 파일로 실제 Claude Code Agent → 자식 Bash → 부모 응답을 실행했다. 자식 transcript에서 다음을 확인했다.

```json
{"type":"tool_use","name":"Bash","input":{"command":"printf NATIVE_BOUNDARY_CHILD_OK","description":"Run required boundary validation"}}
{"type":"tool_result","is_error":false,"content":"NATIVE_BOUNDARY_CHILD_OK"}
```

최종 결과: `NATIVE_BOUNDARY_MAIN_OK NATIVE_BOUNDARY_CHILD_OK`, `is_error=false`, 자식 1개 완료·실패 0, exit 0. 첫 콘텐츠 2.920초, 전체 CLI 14.89초. 로그: `/tmp/moai-native-agent-repair.YgM2te/claude-agent-live.log`. 이 한 번의 측정은 응답 시간 보장이 아니다.

## Baseline-attribution

- 작업 트리: `.claude/worktrees/gpt-session-repair`
- 브랜치: `WT-gpt-session-repair`, 기준 HEAD: `b45c81349`, 미커밋 수정 포함.
- 버전: `v3.2.0-rc.11`
- Build ID: `v3.2.0-rc.11-b45c81349-native-agent-boundary-dirty`
- 후보 SHA-256: `621aa8637bfc6ebeabca37b88395cb74b64f3aa00c73ebecd352537411e7e975`
- 빌드: `2026-09-14T15:40:47Z`
- 배포 상태: `/Users/goos/go/bin/moai` 교체 완료. 설치본 SHA-256은 후보와 동일하며 version 명령으로 새 Build ID를 확인했다.
- 이전 설치본 백업: `/Users/goos/.moai/releases/rc11-native-agent-backup.hrnZue/moai-before-native-agent-repair`.

추가 실제 회귀 검사:

```sh
MOAI_GPT_LIVE=1 go test ./internal/cli -run '^TestSharedGPTLive(ConcurrentToolRouting|InstructionsModelResumeAndImage|CancelThenFollowup)$' -count=1 -v -timeout=240s
```

```text
--- PASS: TestSharedGPTLiveInstructionsModelResumeAndImage (11.47s)
--- PASS: TestSharedGPTLiveConcurrentToolRouting (16.33s)
PASS
ok github.com/modu-ai/moai-adk/internal/cli 33.924s
```

로그: `/tmp/moai-native-agent-repair.YgM2te/live-regression.log`. 실제 네 모델 도구 라우팅, 지시문 변경, 모델 변경, 재개, 이미지 입력을 포함한다. 이는 이미지 생성 검증과 다르다.

## Gaps

전체 저장소 CI와 장시간 Factory 부하는 실행하지 않았다. 기존 실패 thread는 자동 복구하지 않는다. 로드된 thread는 resume config를 무시할 수 있으므로 새 대화로 검증해야 한다. 내부 도구가 호출되지 않은 한 번의 결과만으로 전체 모델 컨텍스트에서 도구가 제거됐다고 주장하지 않는다.

## Residual-risk

지원 모델 또는 Codex 설정 우선순위가 변경되면 재검증이 필요하다. 등록되지 않은 RPC 요청은 계속 거부하므로 다른 원인으로 같은 경계가 침범되면 안전하게 실패할 수 있다. 이번 수정은 해당 검사를 완화하거나 자식 요청을 부모 권한으로 실행하지 않는다.
