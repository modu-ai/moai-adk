# moai gpt 실제 구독 검증 및 rc.11 로컬 배포

> **후속 장애·수정 안내:** 이후 실제 사용에서 `appserver_limit_exceeded`와 120초 이상의 세션 조회 지연이 발견되었다. 이 문서의 검증은 아래에 명시된 당시 범위에 한정된다. 현재 설치본과 후속 수정은 [9월 15일 이벤트 큐·조회 지연 수정 보고서](../gpt-queue-repair-20260915/report.md)를 기준으로 한다.

## Claim

**실제 ChatGPT 구독을 사용하는 GPT 메인 세션, 서브에이전트, 도구 왕복, 재개, 압축 후 재개, 지시문 변경, 모델 변경, 이미지 입력, 취소 후속 요청을 검증하고 rc.11 수정본을 로컬 설치했다.** 모든 Claude Code 기능의 완전한 동등성이나 일반적인 속도 보장까지 입증한 것은 아니다.

사용자가 기존 세션을 종료한 뒤 `profile_lock=free`를 확인했다. 새 공유 App Server를 제품 경로로 시작했으며 인증 파일을 삭제하거나 별도 로그인 소유자를 강제로 추가하지 않았다.

### 실제 검증 결과

| 검증 대상 | 관측 결과 | 판정 |
|---|---|---|
| 네 모델 공유 연결 | sol·terra, luna·astra 두 쌍의 동시 echo 도구 인자·결과 분리; 한 client 종료 후 다른 client 계속 응답 | PASS |
| 실제 Claude Code 메인 | `MOAI_GPT_LIVE_OK.` 반환 | PASS |
| 실제 Claude Code Agent | 자식 1개 시작·완료, 실패 0; 부모에 자식 결과 반환 | PASS |
| 실제 Claude Code 파일 도구 | synthetic 임시 파일 Write → Read, transcript에서 두 도구 이름 확인, 파일 내용 `TOOL_FILE_OK` | PASS |
| 새 프로세스의 대화 재개 | 이전 자식 응답 `CHILD_LIVE_OK` 회상 | PASS |
| 수동 `/compact` 후 재개 | native `compact_boundary manual` 확인, 재개하여 압축 전 자식 응답 회상 | PASS |
| 완료된 turn 이후 지시문 변경 | 실제 결과 `PHASE_ALPHA` → `PHASE_BETA` | PASS |
| idle 모델 변경 | sol → terra 이후 이전 synthetic 단어 유지 | PASS |
| engine·client 재생성 | 새 연결에서 durable resume 후 이전 단어 유지 | PASS |
| 이미지 입력 | 실제 native inline 64×64 빨간 PNG를 `red`로 식별 | PASS |
| JSON schema | Claude `structured_output.status=SCHEMA_OK` | PASS |
| 실제 스트림 취소 | 텍스트 수신 후 취소, idle 복구 후 같은 owner의 새 요청 `CANCEL_FOLLOWUP_OK` | PASS |
| 최종 설치본 | 실제 구독 Agent 왕복 `INSTALLED_GATEWAY_OK INSTALLED_CHILD_OK`, `is_error=false` | PASS |

### 이번 실사용 검사에서 추가 발견·수정한 문제

1. **지시문 갱신 무시:** 기존 `thread/resume.developerInstructions`는 이미 로드된 native thread에서 무시됐다. 실제 동일 실패를 세 차례 관측했다. 현재 지시문을 `turn/start.collaborationMode.settings.developer_instructions`로 전달하고 모델·effort도 함께 명시하도록 수정했다. 지시문 변경용 추가 RPC는 없다. 새 thread에는 변경 가능한 지시문을 정적으로 고정하지 않는다.
2. **launcher `--` 전달 오류:** 첫 `--`를 Claude까지 전달해 뒤의 `--print` 등이 옵션으로 처리되지 않았다. MoAI의 첫 구분자만 소비하고 나머지 인자·빈 문자열·두 번째 구분자는 그대로 보존한다. 수정 전 실제 명령 실패, 회귀 RED, 수정 후 실제 명령 성공을 확인했다.
3. **압축된 대화의 재개 거부:** native compact summary의 user row를 미완료 사용자 요청으로 오인했다. 직전 완료 응답이 검증되고 boundary UUID·parent·native summary flags·관측한 prefix가 일치할 때만 완료를 보존한다. 미완료 도구·누락된 summary·foreign identity·나중 사용자 요청을 느슨하게 허용하지 않는다.
4. **테스트 서버의 이전 규격 의존:** 지시문 전달 수정 후 테스트용 서버가 thread/start에서만 자식 지시문을 찾아 75초 제한에 걸렸다. 실제 규격인 turn/start에서 읽도록 fixture를 수정했고 기존 검증 조건을 유지한 채 통과했다. 이 중간 실패도 증거로 보존한다.

이 작업은 `moai-fix`의 재현·수정·재검증 절차를 적용했다. 구현은 전용 작업 트리에서 담당 파일을 분리해 수행했다.

## Evidence

### 실제 구독 최종 검사

작업 트리에서 실행:

```sh
MOAI_GPT_LIVE=1 go test ./internal/cli -run '^TestSharedGPTLive(InstructionsModelResumeAndImage|CancelThenFollowup|ConcurrentToolRouting)$' -count=1 -v -timeout=300s
```

핵심 원문:

```text
actual text cancellation recovered idle in 12ms
same-owner fresh followup="CANCEL_FOLLOWUP_OK"; no tools exposed
--- PASS: TestSharedGPTLiveCancelThenFollowup (5.97s)
initial instructions: model=gpt-5.6-sol first_text=2.731s completed=2.904s result="PHASE_ALPHA" usage_observed=true
completed-turn instructions override: model=gpt-5.6-sol first_text=1.611s completed=1.78s result="PHASE_BETA" usage_observed=true
idle model switch retains history: model=gpt-5.6-terra first_text=1.41s completed=1.627s result="amber-quokka-731" usage_observed=true
new client durable resume: model=gpt-5.6-terra first_text=1.563s completed=1.737s result="amber-quokka-731" usage_observed=false
native inline image input: model=gpt-5.6-terra first_text=1.404s completed=1.538s result="red" usage_observed=true
--- PASS: TestSharedGPTLiveInstructionsModelResumeAndImage (9.59s)
--- PASS: TestSharedGPTLiveConcurrentToolRouting (18.71s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	35.166s
```

[전체 실제 구독 검사 출력](evidence/live-acceptance-final.log)

### 최종 영향 패키지 race 검사

```sh
unset MOAI_GPT_LIVE MOAI_CLAUDE_INTEGRATION MOAI_CODEX_TRANSPORT_LIVE MOAI_GPT_TEST_BINARY && go test -race ./internal/codexapp ./internal/codexbridge ./internal/gateway ./internal/gateway/translate ./internal/gateway/conversation -count=1
```

```text
ok  	github.com/modu-ai/moai-adk/internal/codexapp	3.725s
ok  	github.com/modu-ai/moai-adk/internal/codexbridge	17.497s
ok  	github.com/modu-ai/moai-adk/internal/gateway	14.295s
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	6.268s
ok  	github.com/modu-ai/moai-adk/internal/gateway/conversation	5.888s
```

[race 원문](evidence/race-live-final.log)

### 최종 CLI·공식 로컬 전송 검사

```sh
unset MOAI_GPT_LIVE && MOAI_CLAUDE_INTEGRATION=1 MOAI_CODEX_TRANSPORT_LIVE=1 MOAI_GPT_TEST_BINARY=/tmp/moai-gpt-live.idNGVc/moai-rc11-repair go test ./internal/cli ./internal/codexapp -run '^(TestManagedGPT.*|TestGPTProductionAppServerWiring|TestShared.*|TestClaudeCodeProductionGatewayAgentRoundTrip|TestClaudeLauncherConsumesOnlyLauncherSeparator|TestGatewayLaunch.*|TestUnifiedGatewayLaunch.*|TestCodexVerbRouting_PassthroughTailExact)$' -count=1 -v -timeout=120s
```

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	2.021s
ok  	github.com/modu-ai/moai-adk/internal/codexapp	0.383s
```

이 실행의 실제 구독 전용 테스트는 의도적으로 SKIP이다. 위의 별도 구독 실행으로 세 개의 공유 live test를 검증했다. legacy direct-owner 방식의 `TestManagedGPTLiveSimpleTurn/Resume`는 실행하지 않았다.

[최종 CLI 검사 원문](evidence/native-final-corrected.log), [이전 규격 fixture 실패](evidence/native-fixture-stale-protocol.log)

```sh
go vet ./internal/codexapp ./internal/codexbridge ./internal/gateway ./internal/gateway/translate ./internal/gateway/conversation ./internal/cli
git diff --check
```

exit 0, 출력 없음. vet은 fixture의 최종 테스트 전용 변경 이전 실행이며 제품 코드 변경 이후 실행이다.

### 실제 Claude Code 출력 증거

- [메인 응답](evidence/live-simple.log)
- [Agent 왕복](evidence/live-agent.log)
- [재개](evidence/live-resume.log)
- [Write·Read](evidence/live-tools.log)
- [JSON schema](evidence/live-schema.log)
- [압축 명령](evidence/live-compact-before.log)
- [압축 후 재개 수정 전 실패](evidence/live-postcompact-before.log)
- [압축 후 재개 수정 후 성공](evidence/live-postcompact-final.log)
- [최종 설치본 Agent 왕복](evidence/live-installed-agent.log)

설치본 실행 명령:

```sh
unset CLAUDECODE ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN ANTHROPIC_BASE_URL && /usr/bin/time -p /Users/goos/go/bin/moai gpt --model gpt-5.6-sol -- --print --output-format json --tools Agent --allowedTools Agent --agents '{"probe":{"description":"Fixed marker verification child","prompt":"Reply exactly INSTALLED_CHILD_OK. Do not use tools.","tools":[]}}' --max-turns 5 'Use probe once. Confirm it returned INSTALLED_CHILD_OK, then reply INSTALLED_GATEWAY_OK with that marker.'
```

작업 디렉터리는 synthetic fixture 전용 `/tmp/moai-gpt-live.idNGVc`이다. 결과 JSON의 `result`는 `INSTALLED_GATEWAY_OK INSTALLED_CHILD_OK`, `is_error`는 false, 자식 spawned/completed/failed는 각각 1/1/0이었다. 실제 명령은 exit 0이었다.

### 속도 관측

| 시나리오 | 첫 콘텐츠/텍스트 | 완료 |
|---|---:|---:|
| 초기 후보 메인 단문 | 2.616초 | CLI 3.58초 |
| 최종 bridge 지시문 갱신 | 1.611초 | bridge 1.780초 |
| 최종 bridge 이미지 입력 | 1.404초 | bridge 1.538초 |
| 최종 설치본 메인→Agent→메인 | 3.223초 | Claude 9.541초 / 전체 CLI 10.40초 |

각 값은 해당 실행 한 번의 관측이다. 부하·모델·출력 길이·캐시를 통제한 벤치마크나 p95, 기존 rc.11 대비 성능 향상률이 아니다. bridge 시간과 전체 CLI 시간을 서로 바꿔 비교하면 안 된다.

## Baseline-attribution

- 작업 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/gpt-session-repair`
- 브랜치/기준 HEAD: `WT-gpt-session-repair` / `b45c813493751106cd6619dce60fba1c61e70583`
- 현재 검증 세션: `ba07dea0-d7ff-427c-b672-fce37365f61c`
- 배포는 커밋되지 않은 작업 트리의 로컬 빌드다. primary/develop에 merge·commit·push하지 않았다.
- [최종 수정 소스 SHA-256 목록](evidence/source-sha256-live-final.txt)
- 버전: `v3.2.0-rc.11`
- Build ID: `v3.2.0-rc.11-b45c81349-session-repair-dirty`
- 빌드 시각: `2026-09-14T14:50:20Z`
- 설치: `/Users/goos/go/bin/moai`
- 설치 SHA-256: `108487159da3321fee29ca8febe7c770efcba9c91d76ad4bfda069d030bd5f89`
- 기존 rc.11 백업: `/Users/goos/.moai/releases/rc11-session-repair-backup.U87J2G/moai-v3.2.0-rc.11-before-repair`
- 백업 SHA-256: `5295ffc469ea8e75130b3b6caac702d14b9641eae0509cafc7925dd6be706131`
- 같은 백업 디렉터리에 수정본 `moai-v3.2.0-rc.11-session-repair`도 보관했다.

기존 설치 해시가 이전 값 그대로인 것을 재확인한 뒤 백업하고, 설치 디렉터리 안의 임시 파일을 rename하여 교체했다. 이후 설치 파일 해시·version·실제 구독 Agent 왕복을 확인했다.

## Gaps

1. 전체 repository CI, develop 통합, 원격 배포·릴리스는 수행하지 않았다.
2. factory 장시간 다중 lane 부하·세션 종료 조합 전체, Windows 병렬 경로, 모든 MCP·plugin·hook·permission 모드·interactive UI 기능의 전수 검증은 아니다. 기존 kanban/factory launcher GAP은 별도다.
3. 이미지 **입력**을 검증했으며 `codex-images-subscription`의 `gpt-image-2.5` 이미지 **생성**은 이번 실행에 포함하지 않았다.
4. 기존 구현이 정적 지시문을 고정한 오래된 native thread의 지시문 마이그레이션 동등성은 미검증이다. 정상적인 새 `moai gpt` 세션을 권장한다. 원래 502가 난 실패 기록은 삭제·자동 복구하지 않았다.
5. 사용량이 없는 응답이나 재개 첫 응답은 필수 프로토콜 필드가 0으로 표시될 수 있다. 실제 관측된 사용량 0이나 무료 실행이라는 뜻이 아니다.

## Residual-risk

- Claude 출력에 `unrecognized_model` 진단이 남지만 이번 실행에서는 성공 결과가 반환됐다. 향후 Claude 버전의 모든 모델 취급을 보장하지 않는다.
- `costBasis: unknown`인 Claude 계산 달러 금액은 실제 구독 청구액으로 사용할 수 없다.
- 관측한 Claude 2.1.270 compact/summary 형식이 바뀌면 엄격한 식별이 다시 거부할 수 있다.
- 취소 검증은 도구가 없는 텍스트 turn이다. 실행 여부가 불확실한 부작용 도구는 자동 재시도하지 않는다.
- 공유 서버는 여러 gateway가 사용하는 의도적인 지속 프로세스다. 첫 세션 종료가 전체 서버 종료를 뜻하지 않는다.
- 로컬 배포 성공과 여기 명시한 실사용 경로의 성공을, 모든 기능의 완전한 운영 동등성 또는 보편적인 응답 속도 보장으로 확대 해석하지 않는다.
