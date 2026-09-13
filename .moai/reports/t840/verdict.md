# t840 판정서 — 게이트웨이(gpt) 백엔드 모델 상속 표시 (웹 매트릭스 1단계)

- 카드: t840 (Class B, Tier S)
- 워크트리: `.claude/worktrees/t840`, 브랜치 `WT-web-inherit-row`
- 기준 커밋: `4da5d1c4e` (develop)
- 작성 세션: t841 워크트리에 고정된 세션. 워크트리 세션 가드가 t840 대상 git 명령을 모두 거부하여 **이 세션은 커밋을 수행하지 못했다** (§ 미검증 참조).

## 1. 주장 (Claim)

1. 게이트웨이(gpt) 백엔드 신호가 있을 때 웹 콘솔의 Agents 패널은 에이전트별 모델 칸을 편집 가능한 select 대신 고정 `inherit` 표시로 렌더링하고, effort select만 편집 가능하게 유지한다. Claude·GLM 백엔드에서는 기존 렌더링이 그대로다.
2. 모델 칸이 제출되지 않아도 저장 경로는 손상되지 않는다. `parseAgentFMForm`이 미제출 모델을 프로필 매트릭스 해석값으로 채우므로(기존 코드, 496~498행), effort만 바꾼 저장은 `{해석된 모델, 새 effort}`를 고정하고 effort가 기본값이면 오버라이드를 남기지 않는다.
3. `moai model profile --json`은 게이트웨이 백엔드에서 `backend: "gpt"`를 보고하고, 매트릭스에 속한 각 에이전트 행에 `gateway_model: "inherit"` 필드를 싣는다. GLM 열의 형식 관례(백엔드 한정 populate, `omitempty`)를 그대로 따르며, 사람용 표에는 `GATEWAY_MODEL` 열이 추가된다.
4. 백엔드 신호는 `template.IsGatewayBackend(llm)` 하나로 읽는다. `llm.team_mode: gpt`(신설 상수 `config.TeamModeGPT`) 또는 휴면 필드 `llm.mode: gpt`가 참이면 게이트웨이다. 런처는 llm.yaml에 아무것도 쓰지 않으므로(기존 테스트 `TestUnifiedGatewayLaunchSkipsLegacyModeMutation`이 이를 고정), 실행 중인 gpt 세션 안에서는 런처가 자식 환경에 심는 `MOAI_LAUNCH_PROVIDER=gpt`를 `LLMConfig.WithLaunchProvider`가 접어 넣는다. 웹 콘솔(GET 뷰 시딩·POST 파싱)과 `moai model profile`이 같은 접기를 호출한다.
5. `internal/web` `internal/template` `internal/config` 패키지 전체 테스트, 변경 패키지 `go vet`, `golangci-lint`가 모두 통과했다.

## 2. 증거 (Evidence)

### 2.1 RED — 구현 전 실패 확인

```
$ go test -C <t840> ./internal/config/ -run TestWithLaunchProvider -count=1
internal/config/team_mode_test.go:17:45: undefined: TeamModeGPT
internal/config/team_mode_test.go:26:44: LLMConfig{…}.WithLaunchProvider undefined
FAIL	github.com/modu-ai/moai-adk/internal/config [build failed]

$ go test -C <t840> ./internal/template/ -run TestIsGatewayBackend -count=1
internal/template/gateway_backend_test.go:29:14: undefined: IsGatewayBackend
FAIL	github.com/modu-ai/moai-adk/internal/template [build failed]

$ go test -C <t840> ./internal/web/ -run 'TestAgentFMGateway' -count=1
internal/web/agentfm_gateway_inherit_test.go:113:62: undefined: config.TeamModeGPT
FAIL	github.com/modu-ai/moai-adk/internal/web [build failed]

$ go test -C <t840> ./internal/cli/ -run 'TestResolveModelProfileReport_Gateway|TestResolveModelProfileReport_ClaudeHasNoGatewayField' -count=1
internal/cli/model_gateway_test.go:54:8: e.GatewayModel undefined (type modelProfileEntry has no field or method GatewayModel)
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
```

### 2.2 GREEN — 선택 테스트

```
$ go test -C <t840> ./internal/config/ -run TestWithLaunchProvider -count=1 -v
--- PASS: TestWithLaunchProvider (0.00s)   (하위 6건 전부 PASS)
ok  	github.com/modu-ai/moai-adk/internal/config	0.546s

$ go test -C <t840> ./internal/template/ -run 'TestIsGatewayBackend|TestIsGLMBackend' -count=1
ok  	github.com/modu-ai/moai-adk/internal/template	0.480s

$ go test -C <t840> ./internal/web/ -run 'TestAgentFMGateway|TestAgentFMGLM|TestM5AgentFM|TestHaiku' -count=1 -v
--- PASS: TestAgentFMGatewayInheritCell (0.01s)
--- PASS: TestAgentFMGatewayCellHiddenUnderClaudeAndGLM (0.02s)
--- PASS: TestAgentFMGatewayNoteKeyInFourLocales (0.00s)
--- PASS: TestAgentFMGatewaySaveBackfillsModel (0.00s)
--- PASS: TestAgentFMGLMReasoningMapRendered (0.01s)
--- PASS: TestAgentFMGLMReasoningMapFlashPinsMax (0.01s)
--- PASS: TestAgentFMGLMReasoningMapHiddenUnderClaude (0.01s)
--- PASS: TestAgentFMGLMNoteKeyInFourLocales (0.00s)
--- PASS: TestM5AgentFMRenderBadgeAndSelects (0.01s)
--- PASS: TestHaikuEffortSelectDisabledOnRender (0.01s)
--- PASS: TestHaikuHintRenderedVisibleForHaiku (0.01s)
--- PASS: TestHaikuHintKeyInFourLocales (0.00s)
--- PASS: TestHaikuSaveWithoutEffortFieldSucceeds (0.01s)
ok  	github.com/modu-ai/moai-adk/internal/web	0.775s

$ go test -C <t840> ./internal/cli/ -run 'TestResolveModelProfileReport|TestModelProfileJSONShapeMatchesDoc' -count=1 -v
--- PASS: TestModelProfileJSONShapeMatchesDoc (0.00s)
--- PASS: TestResolveModelProfileReport_GatewayInherit (0.00s)
--- PASS: TestResolveModelProfileReport_ClaudeHasNoGatewayField (0.00s)
--- PASS: TestResolveModelProfileReport_MaxClaude (0.00s)
--- PASS: TestResolveModelProfileReport_GLMOverlay (0.00s)
--- PASS: TestResolveModelProfileReport_GLMModelExplicitInJSON (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.969s
```

### 2.3 templ 재생성

```
$ go -C <t840> run github.com/a-h/templ/cmd/templ generate -path <t840>/internal/web
(✓) Complete [ updates=1 duration=71.261917ms ]
$ grep -c "data-model-inherit" internal/web/fieldsets_templ.go   → 1
$ grep -c "agentfm.gatewaynote" internal/web/fieldsets_templ.go  → 2
```

### 2.4 패키지 전체 테스트

```
$ go test -C <t840> ./internal/web/ ./internal/template/ ./internal/config/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/web	27.275s
ok  	github.com/modu-ai/moai-adk/internal/template	38.004s
ok  	github.com/modu-ai/moai-adk/internal/config	4.053s
exit=0

$ go test -C <t840> ./internal/cli/ -count=1            (기본 10분 타임아웃)
panic: test timed out after 10m0s
FAIL	github.com/modu-ai/moai-adk/internal/cli	600.782s
  → `--- FAIL` 0건. 타임아웃 패닉은 실패 판정이 아니지만 통과 근거도 아니므로 아래 40분 재실행으로 대체한다.

$ go test -C <t840> ./internal/cli/ -count=1 -timeout 40m
panic: test timed out after 40m0s
	running tests:
		TestTodoVerbGuardStillAddsNonAddresses (11s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	2401.773s
  → `--- FAIL` 0건. 타임아웃 시점에 돌던 테스트는 todo 동사 가드(이번 변경과 무관). 측정 중 머신 load 14~18.
    CLAUDE.local.md §4의 원칙대로 로컬 전량 실행은 머신을 재는 것이므로 재시도하지 않고 § 미검증에 남긴다.

$ go test -C <t840> ./internal/cli/ -run 'Gateway|GPT|Gpt|TeamMode|Launch|ModelProfile' -count=1 -timeout 15m -v
--- PASS 191건, --- FAIL 0건
ok  	github.com/modu-ai/moai-adk/internal/cli	2.462s
  → team_mode 상수·`WithLaunchProvider`·model.go 변경이 닿을 수 있는 런처·게이트웨이·모델 프로필 테스트군.
    선택자 적중 수(191)를 세어 공허 통과가 아님을 확인했다.
```

### 2.5 정적 검사

```
$ go vet -C <t840> ./internal/web/... ./internal/cli/... ./internal/template/ ./internal/config/
vet exit=0

$ (cd <t840>) golangci-lint run ./internal/web/... ./internal/cli/... ./internal/template/ ./internal/config/
0 issues.
lint exit=0

$ gofmt -l <변경 Go 파일 10개>
(없음 — agentfm_gateway_inherit_test.go의 import 순서 1건을 gofmt -w로 정리한 뒤 재확인)
```

## 3. 기준 귀속 (Baseline-attribution)

- 측정 트리: `.claude/worktrees/t840`, HEAD `4da5d1c4e059f526bbdce6a4a7c291df64dfb10b` (`.git/worktrees/t840/HEAD` → `refs/heads/WT-web-inherit-row`를 직접 읽음. git 명령은 가드가 거부).
- 모든 명령은 이 트리에서 `go -C <t840>` 형태로 실행했고, 위 출력은 이 실행에서 관측한 것이다. 다른 트리·다른 시점의 수치를 옮겨 오지 않았다.
- RED 출력(2.1)은 구현 파일을 만들기 전에, GREEN 출력(2.2)은 구현 파일 설치와 templ 재생성 뒤에 각각 관측했다.

## 4. 미검증 (Gaps)

1. **커밋 미수행.** 이 세션은 t841 워크트리에 고정돼 있어 `git -C <t840>`·`cd <t840> && git` 모두 워크트리 세션 가드에 거부됐다. 변경은 t840 작업 트리에 미커밋 상태로 남아 있다. 리드가 t840에 앵커된 세션에서 아래 변경 파일 목록을 명시 pathspec으로 커밋해야 한다. push는 어차피 금지된 카드다.
2. **실제 gpt 세션에서의 통합 관측 없음.** `MOAI_LAUNCH_PROVIDER=gpt` 접기는 단위 테스트(`TestWithLaunchProvider`)로만 확인했다. 실제 `moai gpt` 세션 안에서 `moai web`을 띄워 inherit 칸이 보이는지, `moai model profile`이 `backend: gpt`를 찍는지는 관측하지 않았다 (이 세션에서 gpt 런치는 트랜스포트 검증 대기 상태로 차단돼 있음).
3. **브라우저 렌더링 미확인.** 서버 렌더 HTML 문자열만 검증했다. `app.js`의 haiku 잠금 로직은 모델 select가 없는 행에서는 대상 요소를 찾지 못해 자연히 건너뛰지만, 실제 브라우저에서 동작을 보지는 않았다.
4. **`internal/cli` 패키지 전량 통과를 관측하지 못했다.** 기본 10분·연장 40분 두 번 모두 타임아웃 패닉으로 끝났고 `--- FAIL`은 0건이다. 이번 변경이 닿는 191건의 선택 실행은 통과했다. 전 패키지 판정은 CI 몫이다.
5. **문서 미갱신.** `model-policy.md`의 `--json` 예시는 GLM 열(`glm_model`)도 싣지 않으므로 `gateway_model`도 추가하지 않았다. 문서 키 ⊆ 실제 출력 방향의 `TestModelProfileJSONShapeMatchesDoc`는 통과. `agentfm.glmnote`가 `internal/web/assets/i18n.js`에만 존재하므로 i18n 템플릿 미러 의무는 없다고 판단했다.
6. `internal/hook` 등 `MOAI_LAUNCH_PROVIDER`를 읽는 다른 소비자는 건드리지 않았고 재측정하지 않았다.

## 5. 잔여 위험 (Residual-risk)

1. `LLMConfig.WithLaunchProvider`는 llm.yaml에 team_mode가 비어 있을 때만 gpt를 접는다. 사용자가 이전에 `moai glm`을 돌려 `team_mode: glm`이 남은 프로젝트에서 `moai gpt`를 띄우면(`moai cc`를 거치지 않은 경우) 콘솔은 GLM 백엔드로 읽는다. 런처가 gpt 런치 시 team_mode를 초기화하지 않는 기존 설계의 결과이며, 이번 카드 범위 밖이다.
2. `moai web`을 gpt 세션 밖의 별도 터미널에서 띄우면 환경 신호가 없어 Claude 백엔드로 렌더링된다. 이 경우 `llm.yaml`에 `team_mode: gpt`를 직접 적으면 같은 결과를 얻는다.
3. `TeamModeGPT`가 llm.yaml에 파싱 허용 값으로 추가됐다. `moai cc`의 `resetTeamModeForCC`는 비어 있지 않은 team_mode를 모두 비우므로 gpt 값도 정리되지만, 그 밖의 team_mode 소비자가 "glm/cg/claude/hybrid" 네 값만 가정하고 있다면 알 수 없는 값으로 취급될 수 있다. 이번 실행에서 그런 소비자를 전수 조사하지는 않았다.

## 6. 변경 파일

신규:
- `internal/config/team_mode_test.go`
- `internal/template/gateway_backend.go`
- `internal/template/gateway_backend_test.go`
- `internal/web/agentfm_gateway.go`
- `internal/web/agentfm_gateway_inherit_test.go`
- `internal/cli/model_gateway_test.go`
- `.moai/reports/t840/verdict.md` (이 문서)

수정:
- `internal/config/team_mode.go` — `TeamModeGPT`·`LLMModeGPT` 상수, `LLMConfig.WithLaunchProvider`
- `internal/cli/model.go` — `GatewayModel` 필드, `backend: gpt` 분기, `GATEWAY_MODEL` 표 열, 런치 프로바이더 접기
- `internal/web/fieldsets.templ` — 게이트웨이 노트, inherit 칸, haiku 잠금 예외
- `internal/web/fieldsets_templ.go` — templ 재생성 산출물
- `internal/web/handlers.go` — POST 파싱 경로에 접기 적용
- `internal/web/schemaform.go` — GET 뷰 시딩에 접기 적용
- `internal/web/assets/i18n.js` — `agentfm.gatewaynote` 4개 로케일
