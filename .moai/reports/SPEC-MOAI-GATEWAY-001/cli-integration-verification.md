# CLI·환경·훅 연결 검증 — SPEC-MOAI-GATEWAY-001

## Claim

SPEC 0.8.0의 launcher 연결을 검증 가능한 함수와 실제 private child 경로로 구현했다. 이 보고서는 M3/M4/M7 전체 또는 제품 출시 PASS가 아니다. 인증 운반 키와 PICKER의 실제 설정 우선순위가 검증되기 전까지 production gateway launch는 활성화하지 않았다. 기존 cc/glm은 기존 경로를 사용하며, gpt는 login/logout/status만 제공하고 실행은 진입 부수 작업 전에 명시적으로 거절한다.

구현 범위는 다음과 같다.

- 공통 `runClaudeEntry`를 통해 gpt가 cc의 profile·spawn·worktree·kanban/factory 진입 처리를 재사용한다. `gg` 별칭은 없다.
- `unifiedLaunchWithGateway`는 기존 profile 처리를 유지하면서 gateway 경로에서 옛 applyCC/applyGLM 설정·tmux 변경을 건너뛴다. 프로젝트 settings.local.json에 남은 14키는 기존 잠금·원자적 쓰기 함수로 정리한다. 비어 있지 않은 기존 backup token만 복원하고 backup 키 자체는 항상 지운다. 사용자 권한 설정은 보존한다.
- 초기 모델은 명시적 `--model`, `--model=`, `-m`을 우선한다. `--` 이후 값은 prompt로 보존한다. cc에만 Claude profile/DO_CLAUDE 기본값을 적용하고, gpt는 gpt-5.6-sol, glm은 GLM.High가 기본이다. GLM의 짧은 tier 별칭은 해당 설정 ID로 풀이하며 전체 Claude/GPT ID는 명시적 provider 전환으로 처리한다.
- 세 gateway 명령은 항상 일반 Claude Code effort 계산을 사용한다. 옛 team_mode=glm에 따른 effort 제거·변환을 건너뛰며 기존 비-gateway GLM 경로는 유지한다.
- child env의 14키를 먼저 지운 뒤 launcher 값만 추가한다. Z_AI_API_KEY는 cc/gpt에서 지우고 glm에서 resolver가 공급한 값으로 바꾼다. GLM 네 슬롯은 모두 Z.AI로 등록된 설정 ID인지 확인한다. MOAI_LAUNCH_PROVIDER는 command mode가 아니라 초기 모델의 실제 provider를 나타낸다.
- 사용자 `--settings`의 env는 문자열 map으로 검사해 메모리의 child env 입력으로 옮기고, 파일 overlay에서는 제거한다. permissions 등 나머지 설정을 보존하고 gateway overlay를 적용한다. 중복 JSON key·잘못된 UTF-8·복수 settings 입력·잘못된 env 값은 명시적으로 거절한다. 이 이동은 실제 Claude source precedence가 확인되기 전의 호환성 후보다.
- 선택된 named profile은 settings.env의 CLAUDE_CONFIG_DIR가 다른 profile로 바꾸지 못한다. gateway settings 인수는 기존 인수를 교체하며 `--` 앞에 둔다.
- `newGatewaySessionBinding`은 실제 gateway.StartChild를 호출하고 child가 만든 overlay와 실제 bound address를 받는다. 같은 준비 결과를 normal/continue/fallback에 사용한다. 오류나 continue 반환에서는 child를 Stop한다. POSIX exec와 Windows spawn-and-exit seam은 기존 함수를 그대로 사용한다. 신규 spawn telemetry는 성공한 continue에 기록하지 않는다.
- 숨은 `internal-gateway`는 private stdin 설정만 읽고 handler factory를 통해 RunChildWithControl로 진입한다. root dependency graph를 초기화하지 않으며 factory가 없으면 시작하지 않는다. 부모의 Claude OAuth secret을 payload에 스냅샷으로 담는 코드는 없다.
- gateway SessionStart/End는 기존 GLM credential 주입·tmux 정리·settings GLM 정리를 건너뛴다. teammateMode만 in-process로 바꾸고 나머지 env와 permissions를 보존한다. session record는 초기 provider 신호를 우선하며 기존 record는 덮어쓰지 않는다.
- GPT의 private store는 정규화한 MoAI home/gateway-auth를 사용한다. installed Codex broker는 login 요청 때만 찾아 실행하며 사용자 CODEX_HOME을 읽지 않는다. status는 로그인 여부만 표시하고 generation은 노출하지 않는다. logout 문구는 로컬 로그아웃임을 밝히며 원격 revoke 완료를 주장하지 않는다.

## Evidence

모든 Go 명령의 작업 디렉터리는 이 WT다. 실제 Claude, 실제 Codex 계정 로그인, provider endpoint 요청은 실행하지 않았다. broker 시험은 임시 auth.json fixture 또는 실패만 반환하는 임시 shell 실행 파일을 사용했다. supervisor 시험의 HTTP handler는 localhost에서 204만 반환한다.

### TDD RED 기록

신규 함수와 타입을 작성하기 전에 실행한 시험의 컴파일 RED에는 다음 출력이 있었다(각 실행 exit 1).

```text
internal/hook/gateway_guard_test.go:15:19: undefined: config.EnvMoaiLaunchProvider
internal/cli/gateway_prepare_test.go:17:12: undefined: prepareGatewayLaunch
internal/cli/gateway_prepare_test.go:17:33: undefined: gatewayPrepareInput
internal/cli/gateway_child_test.go:14:8: undefined: newGatewayChildCommand
internal/cli/gateway_launcher_test.go:19:13: undefined: gatewayLaunchBinding
internal/cli/gateway_launcher_test.go:19:61: undefined: gatewayLaunchRequest
internal/cli/gateway_launcher_test.go:24:11: undefined: launchClaudeWithGateway
internal/cli/gateway_settings_test.go:15:11: undefined: cleanupGatewaySettings
internal/cli/gateway_launcher_test.go:65:10: undefined: unifiedLaunchWithGateway
internal/cli/gpt_test.go:11:19: undefined: newGPTCommand
internal/cli/gpt_test.go:11:33: undefined: gptCommandServices
internal/cli/gateway_settings_test.go:46:24: undefined: prepareGatewayOverlay
internal/cli/gpt_test.go:58:12: undefined: newGPTAuthServices
internal/gateway/json_object_test.go:6:161: undefined: ValidateJSONObject
internal/cli/gateway_session_test.go:15:11: undefined: newGatewaySessionBinding
internal/cli/gateway_session_test.go:15:36: undefined: gatewaySessionOptions
internal/cli/gateway_session_test.go:23:12: undefined: replaceGatewaySettingsArgs
```

각 RED 명령은 아래 형식에서 해당 시험 이름만 좁혀 실행했다.

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test ./internal/cli -run <해당_TestGateway_또는_TestGPT_이름> -timeout 30s
```

hook은 `./internal/hook -run TestGatewayHooksPreserveRoutingSettings`, JSON helper는 `./internal/gateway -run TestValidateJSONObject`로 실행했다. 함수 생성 이후의 동작 RED도 남았다.

```text
--- FAIL: TestGatewayPreparationRemovesExactFourteenInheritedKeys (0.00s)
    gateway_prepare_test.go:55: inherited key survived: CLAUDE_CODE_AUTO_COMPACT_WINDOW=polluted
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.930s
FAIL

--- FAIL: TestGatewayRecordUsesInitialProviderAndKeepsExistingRecord (0.00s)
    gateway_guard_test.go:62: backend "claude"
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.669s
FAIL

--- FAIL: TestGatewaySessionSettingsCannotRedirectSelectedProfile (0.00s)
    gateway_session_test.go:158: profile override "CLAUDE_CONFIG_DIR=wrong-profile"
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.043s
FAIL

--- FAIL: TestGatewayGLMExplicitTierAliasesKeepConfiguredRouting (0.00s)
    gateway_prepare_test.go:77: opus resolved claude-opus-5, want glm-high
--- FAIL: TestGatewayCommandsAvoidDependencyGraphInitialization (0.00s)
    gpt_test.go:158: gateway command initializes unrelated dependencies: [gpt status]
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.997s
FAIL

--- FAIL: TestGPTStatusShowsLoginStateWithoutInternalGeneration (0.00s)
    gpt_test.go:215: unexpected status "GPT: not logged in (generation 0)\n"
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.005s
FAIL

--- FAIL: TestLegacyContinueDoesNotRecordNewWorktreeSpawn (0.13s)
    gateway_launcher_test.go:136: continued session recorded a new spawn: before 0 after 1
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.143s
FAIL
```

마지막 effort 변경 전 명령은 `go test ./internal/cli -run TestGatewayEffortUsesClaudeSettings -timeout 30s`였으며, 동일한 unset/GOCACHE 환경에서 다음 RED를 얻었다.

```text
--- FAIL: TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode (0.01s)
    --- FAIL: TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode/claude (0.00s)
        gateway_launcher_test.go:168: gateway Claude effort "", want high
    --- FAIL: TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode/gpt (0.00s)
        gateway_launcher_test.go:168: gateway Claude effort "", want high
    --- FAIL: TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode/glm (0.00s)
        gateway_launcher_test.go:168: gateway Claude effort "", want high
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.991s
FAIL
```

이들 시험은 순서대로 해당 구현을 수정한 뒤 GREEN을 확인했다. task-list 도구가 없는 이 실행에서는 이 RED/GREEN 기록이 시험 상태 기록을 대신한다.

### 환경 때문에 달라진 관측

처음 기존 cc/factory 시험은 MOAI_HOME을 따로 지정하지 않아 `mkdir /Users/goos/.moai/run/001-f745b447: operation not permitted`로 실패했다. 실제 사용자 홈을 허용해 해결하지 않고 임시 MOAI_HOME에서 다시 실행했다. 그 값을 hook의 기존 credential fixture에도 공통 적용하면 fixture의 `.moai/.env.glm` 대신 지정된 home을 보게 되어 기존 ensureGLMCredentials 시험이 실패했다. 따라서 CLI 기존 회귀와 hook 기존 회귀의 환경을 분리해 측정했다.

실제 supervisor 시험은 sandbox 안에서 parent fingerprint를 얻지 못해 `gateway child configuration invalid`로 실패했다. 제한을 풀고 임시 helper·localhost만 사용하는 동일 시험을 다시 실행해 PASS를 확인했다. pgrep도 sandbox에서는 `Cannot get process list`였으므로 이 결과를 프로세스 부재 증거로 사용하지 않았다. 권한을 얻어 아래 읽기 전용 명령을 실행했을 때 exit 1, 출력 없음이었다.

```sh
pgrep -f '^/.*cli[.]test -test[.]run=.*TestGatewayCLIChildHelper'
```

### 최종 측정

실제 실행 명령:

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/gateway-cli-verification-home GOCACHE=/tmp/gateway-foundation-cache go test ./internal/cli -run 'TestGPT|TestGateway|TestUnifiedGateway|TestLaunchClaudeDefault|TestRunCC|TestCC|TestCharacterize_CC|TestContinueLaunch|TestLegacyContinueDoesNotRecord|TestBuildEnvForGLMLaunch|TestResolveGLMBackendForLaunch|TestResolveGLMMainSessionEffort' -coverprofile=/tmp/gateway-cli-final-cover.out -timeout 90s
```

exit 0:

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	3.703s	coverage: 10.7% of statements
```

10.7%는 큰 기존 CLI 패키지 전체를 분모로 한 값이다. 같은 coverage profile에서 신규 실행 파일의 statement 수를 합산한 결과는 아래와 같다. gateway_launcher.go는 타입 선언만 있어서 실행 statement가 없다.

| 신규 실행 파일 | 실행 statement / 전체 | coverage |
|---|---:|---:|
| gateway_child.go | 14 / 14 | 100.0% |
| gateway_prepare.go | 43 / 47 | 91.5% |
| gateway_session.go | 50 / 56 | 89.3% |
| gateway_settings.go | 69 / 75 | 92.0% |
| gpt.go | 21 / 24 | 87.5% |
| gpt_auth.go | 40 / 46 | 87.0% |

독립적인 읽기 전용 검증을 한 batch로 요청했으며, 모든 결과를 확인했다.

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY MOAI_HOME MOAI_LAUNCH_PROVIDER && GOCACHE=/tmp/gateway-foundation-cache go test ./internal/hook -run 'TestGateway|TestEnsureGLMCredentials|TestEnsureTeammateMode|TestEnsureTmuxGLMEnv|TestCleanupGLMSettingsLocal|TestSession.*Record|TestFactoryLaneRecords' -timeout 60s
```

exit 0:

```text
ok  	github.com/modu-ai/moai-adk/internal/hook	(cached)
```

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/gateway-cli-verification-home GOCACHE=/tmp/gateway-foundation-cache go test -race ./internal/cli ./internal/hook -run 'TestGateway|TestGPT' -timeout 90s
```

exit 0:

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	3.717s
ok  	github.com/modu-ai/moai-adk/internal/hook	(cached)
```

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test ./internal/gateway -run 'TestValidateJSONObject|TestValidation' -timeout 30s
```

exit 0:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.457s
```

다음 명령은 각각 exit 0, 출력 없음이었다.

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go vet ./internal/cli ./internal/hook ./internal/kanban ./internal/gateway
GOCACHE=/tmp/gateway-foundation-cache gopls check internal/cli/launcher.go internal/cli/cc.go internal/cli/root.go internal/cli/gateway_child.go internal/cli/gateway_prepare.go internal/cli/gateway_session.go internal/cli/gateway_settings.go internal/cli/gpt.go internal/cli/gpt_auth.go internal/hook/gateway_guard.go internal/hook/session_start_record.go internal/gateway/validation.go
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 GOCACHE=/tmp/gateway-foundation-cache go test -c ./internal/cli -o /tmp/gateway-cli-windows.test.exe
```

Windows 결과는 컴파일 증거다. 실제 Windows 실행이나 auth 지원의 증거가 아니다.

## Baseline-attribution

이번 실행에서 읽은 branch는 WT-unified-gateway, HEAD는 81c1d58f9, spec.md version은 0.8.0이며 status는 draft다. commit·push·PR·merge·워크트리 제거 및 SPEC status 전이를 하지 않았다. auth와 translate worker의 코드·검증 기록은 별개다. auth API는 이 WT의 concurrent worker 구현을 소비했지만 해당 디렉터리를 수정하지 않았다.

CLI 신규 파일은 gateway_child, gateway_launcher, gateway_prepare, gateway_session, gateway_settings, gpt, gpt_auth와 대응 시험이다. 기존 launcher.go/cc.go/root.go, config/envkeys.go, kanban/record.go, hook의 gateway guard·session start/end·record 경로를 수정했다. 부모가 허락한 JSON helper 추출로 gateway/validation.go와 json_object_test.go도 변경했다. frontend backendBadge는 기존 코드가 backend 이름을 그대로 표시하고 GPT가 기존 flat-rate 분기로 들어가는 것을 읽어 확인했으며 UI 파일은 수정하지 않았다. 실제 browser 검증은 하지 않았다.

## Gaps

- M0의 세션 인증 운반 키, 요청별 Claude OAuth 전달, 실제 provider adapter factory 연결, 실제 Claude exec/PTY/signal, installed Codex broker 호환성, 실제 계정 인증, 실제 모델 전환·fallback 동작은 미검증이다. plain gpt launch 차단은 해당 검증 전의 임시 상태이며 제품의 최종 목표가 아니다.
- 실제 GLM effort의 wire 전달은 후속 adapter 시험이 필요하다. 상속 ANTHROPIC_REASONING_EFFORT는 14키 정리 목록을 추측으로 확대하지 않기 위해 그대로 두었으며, 실제 영향은 별도 정책 검증 대상이다.
- PICKER 설정의 user/project/managed 우선순위, modelPicker 지속성, fallbackModel 배열과 명시 fallback 플래그의 계약은 후속 검증·SPEC 조정 대상이다. settings.env를 메모리 env로 옮기는 후보가 실제 Claude에서 동등한 우선순위를 갖는다고 주장하지 않는다.
- Windows는 시험 바이너리 컴파일만 수행했다. AUTH Store는 현재 Windows에서 ErrPlatformUnsupported로 거절한다. Windows runtime release gate는 통과하지 않았다.
- factory launch event의 backend는 design §7.3에 따라 기존 command backend 계약을 유지한다. session record의 실제 초기 provider와 이 event를 같은 의미라고 주장하지 않는다. 부모가 SPEC의 §7.1/§7.3 용어를 조정할 예정이다.
- 전체 M3/M4/M7 AC, 전체 저장소 시험, 독립 sync audit는 미완료다. 저장소 전체 시험 판정의 소유자는 통합 브랜치의 `.github/workflows/ci.yml` (`name: CI`) 실행이며 보고 시점 PENDING이다. commit/push 전이므로 이 작업의 CI run ID/URL은 없다.
- 구현 전 LSP baseline은 확보하지 않았다. 최종 diagnostics가 없다는 측정만 하며 전후 회귀 수치를 주장하지 않는다. 기존 대형 CLI/hook 파일 전체 coverage 85%를 달성했다고 주장하지 않는다.

## Residual-risk

production 연결자는 `newGatewaySessionBinding`에 검증된 SessionEnv와 Child.StartOptions의 private Payload를 공급하고, root의 `newGatewayChildCommand(nil)`을 검증된 handler factory로 바꿔야 한다. cc/glm의 `unifiedLaunchDefault`에는 승인된 binding을 연결하고, gpt의 `gptCommandServices.Launch`도 명시적으로 연결해야 한다. child argv에는 내부 동사만 두며 token이나 credential을 넣지 않는다. CLI의 parent env/overlay 조립은 launcher 역할이고 요청별 OAuth 갱신은 요청을 받는 handler 역할이다. 두 책임을 startup credential snapshot으로 합치면 안 된다.

settings.local.json backup 복원은 SPEC 요구대로 유지한다. 이것이 실제 Claude 인증 선택에 끼치는 영향은 M0/source precedence 검증 없이는 판정할 수 없다. 이번 시험의 local helper는 실제 Claude 설정 로더를 대체하지 않는다. supervisor의 기존 private overlay 독점 소유 및 강제 OS 종료 시 cleanup 한계는 supervisor-verification.md의 경계를 그대로 유지한다.
