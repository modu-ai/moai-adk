# t668 — 게이트웨이 런치가 effort를 `CLAUDE_CODE_EFFORT_LEVEL`로 고정한다

카드 t668. 브랜치 `WT-gateway-effort-pin`, 기준 `8eade2f8910c3c046539db198433f31e8b8cd740` (origin/develop). 선행 카드 t595(평문 `moai cc` 경로 수리)의 잔여분이다.

**상태: 구현 완료, 검증 완료 (2026-09-13).** 영향 패키지 전량 실행의 FAIL 4건은 전부 이 변경과 무관함을 측정으로 귀속했다(§ 검증 실행 결과). 실물 3모드 기동은 미관측(§ Gaps).

---

## Claim

1. 게이트웨이 바인딩(claude / gpt / glm 모드)으로 뜨는 런치는 `launcher.go`의 `binding != nil` 갈래에서 `buildEnvForLaunch(effectiveEffort, os.Environ())`를 호출해 프로필 effort를 `CLAUDE_CODE_EFFORT_LEVEL`에 심었다. 이 변수는 Claude Code에서 override라서, 세션 중 `/effort`·`/model`의 effort 변경을 세션 내내 거부하게 만든다(t595가 실측으로 확인한 기전).
2. 그 갈래의 주석이 적은 근거("the adapter reads the effort off the environment it is handed")는 **거짓**이다. 게이트웨이의 어떤 구성요소도 이 변수를 읽지 않는다.
3. 게이트웨이 런치는 이미 t595가 배선한 주입 `--settings` payload(`effortLevel`)를 받고 있으며, 그 키는 게이트웨이 설정 병합(`prepareGatewayOverlay`)을 거쳐 자식 설정 파일까지 살아남는다. 따라서 수리는 게이트웨이 갈래를 평문 Claude 갈래와 같은 `buildEnvForClaudeLaunch(os.Environ())`로 합치는 것뿐이다.

## Evidence — 가설 판정 (착수 전, 이 트리)

### (a) 게이트웨이 코드에 이 변수를 읽는 곳이 없다

```
$ grep -rn 'CLAUDE_CODE_EFFORT_LEVEL\|EffortLevel' internal/gateway | grep -v '_test.go'
(출력 없음, rc=1)

$ grep -rn 'EnvClaudeCodeEffortLevel' internal pkg cmd | grep -v '_test.go'
internal/config/envkeys.go:402:	EnvClaudeCodeEffortLevel = "CLAUDE_CODE_EFFORT_LEVEL"
internal/cli/launcher.go:1207:	key := config.EnvClaudeCodeEffortLevel
internal/cli/launcher.go:1262:		if strings.HasPrefix(e, config.EnvClaudeCodeEffortLevel+"=") {
```

`:1207`은 `buildEnvForLaunch`(쓰기), `:1262`는 `buildEnvForGLMLaunch`의 제거(쓰기)다. 문자열 리터럴 전수 grep(`internal pkg cmd`, 테스트 제외)도 주석·템플릿 문서 외에는 이 두 함수만 나왔다. 읽는 곳은 0곳이다.

### (b) 게이트웨이가 import하는 패키지의 환경 읽기

`go list -deps ./internal/gateway/...` 중 저장소 내부 패키지: `codexapp`, `atomicfile`, `codextools`, `codexbridge`, `defs`, `paths`, `glmcred`, `gateway/{auth,opaque,receipt,translate,conversation}`, `execerr`, `foundation`, `core/git`, `homestate`.

```
$ grep -rln 'os.Getenv\|os.Environ\|LookupEnv' internal/gateway | grep -v _test.go
internal/gateway/auth/broker.go
```

- `internal/gateway/auth/broker.go:75` — `os.Getenv("SystemRoot")` 뿐. 자식 환경을 목록으로 새로 구성한다(`:72` 주석 "constructed here, rather than subtracting keys from Environ").
- `internal/codexapp/client.go:140` — `PATH, LANG, LC_ALL, SYSTEMROOT, SystemRoot, WINDIR, TEMP, TMP, TMPDIR` 허용 목록만 전달.
- `internal/glmcred/glmcred.go:48,101` — `paths.EnvHome`, 테스트 키만.

### (c) effort가 어댑터에 도달하는 실제 경로는 요청 본문이다

- `internal/gateway/translate/native_policy.go:47-55` — 요청 본문 `m["effort"]` → `p.Effort`
- `:138-139` — `out["reasoning"] = map[string]any{"effort": p.Effort}`

### (d) 게이트웨이 자식 프로세스는 런치 환경을 받지 않는다

- `internal/cli/gateway_product_binding.go:148` — `Env: gatewayChildEnvironment(os.Environ())`. `launchEnv`가 아니라 moai 프로세스 자신의 환경을 scrub한 것이다. 즉 런처가 `launchEnv`에 심은 값은 게이트웨이 자식에 **구조적으로 도달하지 않는다**. 도달하는 곳은 호스트되는 Claude Code 하나뿐이다.
- `internal/cli/gateway_child.go` — 자식 명령은 설정을 stdin(`gateway.ReadChildConfig`)으로만 받는다.

### (e) 주입 payload는 게이트웨이 경로에 이미 닿는다

- `internal/cli/launcher.go:233` — `extraArgs = appendCrossSessionSettings(root, profileName, extraArgs)`가 `if binding != nil { return launchClaudeWithGateway(...) }` **앞**에서 실행된다. 즉 게이트웨이 런치도 `effortLevel`을 담은 `--settings`를 받는다(`crosssession_settings.go:88` `applyLaunchEffort`).
- `internal/cli/gateway_session.go:120` — 실제 바인딩이 `prepareGatewayOverlay(in.Args, in.Inherited, overlay)`로 그 파일을 읽어 병합한다. `gateway_settings.go:93-125`는 `env`만 떼어내고 나머지 키는 보존하며, overlay가 쓰는 키는 `teammateMode`·`model`·`modelPicker`·`availableModels`·`fallbackModel`·`permissions` 등으로 `effortLevel`과 겹치지 않는다.
- `internal/gateway/supervisor_runner.go:134-153` `newOwnedOverlay` — 병합 결과 바이트를 `settings.json`에 그대로 쓴다. `launcher.go:824-825` `replaceGatewaySettingsArgs`가 원래 `--settings`를 그 경로로 교체한다.
- 칸반/팩토리 레인은 `kanban_settings.go:81`에서 같은 병합을 하므로 동일하다.

**판정: 가설 성립.** 배선 추가는 필요 없고, env 고정만 제거하면 된다.

## Baseline-attribution

- `git rev-parse origin/develop HEAD` → 둘 다 `8eade2f8910c3c046539db198433f31e8b8cd740` (작업 전 추적 파일 수정 0)
- `go version` → `go1.26.8 darwin/arm64`, `claude --version` → `2.1.270 (Claude Code)`
- 위 grep·판독은 전부 이 런에서 이 트리에 대해 수행했다.

---

## 구현

| 파일 | 변경 |
|---|---|
| `internal/cli/launcher.go` | `else if binding != nil { launchEnv = buildEnvForLaunch(...) }` 갈래 삭제 → 게이트웨이도 `launchEnv = buildEnvForClaudeLaunch(os.Environ())`. 거짓 주석을 근거가 있는 설명(읽는 주체는 호스트 Claude Code뿐, 어댑터는 요청 본문, 자식은 별도 scrub 환경, payload는 병합을 통과)으로 교체. `buildEnvForLaunch`의 godoc·`@MX:NOTE`를 "t668 이후 호출부 없음"으로 정정 |
| `internal/cli/gateway_launcher_test.go` | `TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode` 재겨냥 + `TestGatewayLaunchEnvPreservesInheritedEffort` 신규. 공용 픽스처 `gatewayEffortFixture`, `--settings` 추출기 `gatewaySettingsSource` |
| `internal/cli/launch_effort_settings.go` | Scope 주석에 게이트웨이 바인딩 포함 명시 (주석만) |
| `internal/cli/launcher_blockcap_infinite_test.go` | `TestACFM023c_KanbanEnvReachesChildEnvironment` 근거 주석 정정 — "게이트웨이 갈래의 `buildEnvForLaunch`도 필터 래퍼"라는 서술이 거짓이 됨. 단정은 불변 |
| `internal/cli/launcher_test.go`, `internal/cli/mcp_doctor_coverage_test.go` | `buildEnvForLaunch` 테스트 구획 머리주석 정정 (단정 불변) |

GLM 비게이트웨이 갈래(`buildEnvForGLMLaunch`)는 건드리지 않았다.

### 테스트 설계

두 테스트 모두 `launchClaudeWithGateway`가 아니라 **`unifiedLaunchWithGateway`**를 통과시킨다. 기존 테스트는 `launchClaudeWithGateway`를 직접 불러 `appendCrossSessionSettings`(`:233`)를 우회했기 때문에 payload를 관측할 수 없었다. 저장된 GLM 모드 픽스처(`llm.yaml team_mode: glm`)는 원래 테스트의 의도(저장된 GLM 모드가 게이트웨이 런치를 GLM effort 번역으로 끌고 가지 않음)를 보존하려고 유지했다.

- `TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode` (3모드) — 부모 환경에서 `CLAUDE_CODE_EFFORT_LEVEL`을 unset한 채로: (a) 바인딩이 받는 `Inherited`에 그 키가 **없다**, (b) `ANTHROPIC_REASONING_EFFORT`는 픽스처 값 그대로, (c) `--settings` 파일의 `effortLevel == "high"`, (d) 그 인자를 실제 병합 함수 `prepareGatewayOverlay`에 넣은 결과에도 `effortLevel == "high"`.
- `TestGatewayLaunchEnvPreservesInheritedEffort` (3모드) — 부모 환경 `CLAUDE_CODE_EFFORT_LEVEL=xhigh`, 프로필 `high`: 키가 정확히 1개, 값 `xhigh`. t595의 "물려받은 값은 건드리지 않는다" 불변식의 게이트웨이판.

## 검증 실행 결과 (2026-09-13, 이 런)

증거 로그 경로: `.moai/state/verify/t668/` (gitignore 대상 — 로컬 보존).

### RED — 수정 전 코드 (`red.log`)

`go test ./internal/cli/ -count=1 -v -run 'TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode|TestGatewayLaunchEnvPreservesInheritedEffort'` → **exit 1** (load 19.09 시점)

```
=== RUN   TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode/claude
    gateway_launcher_test.go:210: gateway launch env gained CLAUDE_CODE_EFFORT_LEVEL=high; the override freezes in-session effort changes
=== RUN   TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode/gpt
    gateway_launcher_test.go:210: gateway launch env gained CLAUDE_CODE_EFFORT_LEVEL=high; the override freezes in-session effort changes
=== RUN   TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode/glm
    gateway_launcher_test.go:210: gateway launch env gained CLAUDE_CODE_EFFORT_LEVEL=high; the override freezes in-session effort changes
--- FAIL: TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode (0.01s)
=== RUN   TestGatewayLaunchEnvPreservesInheritedEffort/claude
    gateway_launcher_test.go:266: CLAUDE_CODE_EFFORT_LEVEL = "high", want the inherited xhigh
=== RUN   TestGatewayLaunchEnvPreservesInheritedEffort/gpt
    gateway_launcher_test.go:266: CLAUDE_CODE_EFFORT_LEVEL = "high", want the inherited xhigh
=== RUN   TestGatewayLaunchEnvPreservesInheritedEffort/glm
    gateway_launcher_test.go:266: CLAUDE_CODE_EFFORT_LEVEL = "high", want the inherited xhigh
--- FAIL: TestGatewayLaunchEnvPreservesInheritedEffort (0.01s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.932s
```

주의: RED에서 payload 단정(c)(d)는 **이미 통과**했다 — 실패 줄이 env 단정뿐이다. 이는 § Evidence (e)의 판독(payload는 이미 게이트웨이에 닿는다)과 일치한다. payload 단정은 이 카드가 새로 만든 동작이 아니라 **회귀 가드**이며, 살아 있음은 § Mutant M3·M4가 보인다.

### GREEN — 겨냥 실행 (`green.log`)

`go test ./internal/cli/ -count=1 -v -run '<12개 패턴>'` → **exit 0**, `ok github.com/modu-ai/moai-adk/internal/cli 1.095s`

선택자가 없는 이름을 조용히 버리는 함정 때문에 `--- PASS: <name> (` 줄을 이름별로 대조했다. 최상위 PASS 16줄(`TestBuildEnvForLaunch` 접두사가 `_ReplacesExisting`·`_PreservesOtherVars`·`_EmptyEffort`·`_AddsNewEntry` 4개를 추가로 집음), FAIL 0.

```
--- PASS: TestGatewayLaunchUsesPreparedEnvironmentForNormalAndContinue (0.00s)
--- PASS: TestUnifiedGatewayLaunchSkipsLegacyModeMutation (0.00s)
--- PASS: TestGatewayContinueFallbackRetainsOneSupervisor (0.00s)
--- PASS: TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode (0.01s)
--- PASS: TestGatewayLaunchEnvPreservesInheritedEffort (0.01s)
--- PASS: TestGatewaySessionStartsPrivateChildAndKeepsSecretsOutOfOverlay (0.00s)
--- PASS: TestApplyLaunchEffort (0.00s)
--- PASS: TestLaunchEffortReachesGeneralInjection (0.00s)
--- PASS: TestLaunchEffortReachesKanbanInjection (0.00s)
--- PASS: TestACFM023c_KanbanEnvReachesChildEnvironment (0.00s)
--- PASS: TestBuildEnvForLaunch (0.00s)
--- PASS: TestClaudeLaunchEnvPreservesInheritedEffort (0.00s)
    --- PASS: TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode/claude (0.00s)
    --- PASS: TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode/gpt (0.00s)
    --- PASS: TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode/glm (0.00s)
    --- PASS: TestGatewayLaunchEnvPreservesInheritedEffort/claude (0.00s)
    --- PASS: TestGatewayLaunchEnvPreservesInheritedEffort/gpt (0.00s)
    --- PASS: TestGatewayLaunchEnvPreservesInheritedEffort/glm (0.00s)
```

12개 지정 이름 12/12 적중.

### 정적 검사

| 명령 | exit | 관측 |
|---|---|---|
| `gofmt -l internal/cli/` | 0 | 출력 없음 |
| `go vet ./internal/cli/ ./internal/gateway/` | 0 | 출력 0줄 (`vet.log`) |
| `golangci-lint run ./internal/cli/... ./internal/gateway/...` | 1 | `252 issues: errcheck 239 / staticcheck 11 / unused 2` (`lint.log`). 지적 파일 상위: `gateway/auth/send_test.go` 17, `gateway/anthropic_test.go` 15, `gateway/openai_test.go` 14 … 이 카드가 수정한 6개 파일은 0건 |
| `golangci-lint run --new-from-rev=HEAD ./internal/cli/... ./internal/gateway/...` | 0 | `0 issues.` (`lint-new.log`) — 기준 커밋 대비 이 변경이 새로 들인 지적 0 |

### 영향 패키지 전량 (`full.log`)

`go test ./internal/cli/... ./internal/gateway/... -count=1 -timeout 40m` → **exit 1** (시작 시점 load 20.68)

- `grep -c '^--- FAIL'` → **4**, `grep -c '^FAIL'` → **5** (패키지 줄 2 + 최종 `FAIL` 1 + 단독 `FAIL` 2)
- `internal/cli` 1043.977s FAIL, 하위 16패키지 전부 `ok`. `internal/gateway` 11.805s FAIL, 하위 5패키지 전부 `ok`
- 실패 4건: `TestLiveReadersUnchangedByHistoryVerb`, `TestTodoDone_UndoneRoundTripIsByteIdentical`, `TestTodoUndone_SurvivesMigrationFromLegacyJSON` (이상 `internal/cli`), `TestAppServerSubprocessHTTPToolContinuation` (`internal/gateway`)
- 이 카드의 테스트 2건과 게이트웨이·effort 계열 기존 테스트는 전량 실행에서도 FAIL 줄이 없다

#### 실패 4건 귀속 — 측정

**todo 3건 = 이 세션의 환경 오염.** 실패 diff가 전부 `"runtime":{"runs":[{..."run_id":"tlageh"...}],"assignments":[{..."owner_label":"lead"...}]}`를 추가로 담고 있었다(골든은 `"runs":[],"assignments":[]`). 이 레인 세션의 환경을 조회하니 `MOAI_KANBAN_ID=tlageh`, `MOAI_KANBAN_LEAD_NAME=lead`, `MOAI_KANBAN_LEAD_ADDR`, `MOAI_KANBAN_BACKEND`, `MOAI_KANBAN_SETTINGS_INJECTED=1`, `MOAI_FACTORY_WORKERS=10`이 물려받아져 있었다. 한 번의 복합 호출로 scrub 후 재실행:

```
$ unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_KANBAN_LEAD_NAME MOAI_FACTORY_WORKERS && go test ./internal/cli/ -count=1 -v -run 'TestLiveReadersUnchangedByHistoryVerb|TestTodoDone_UndoneRoundTripIsByteIdentical|TestTodoUndone_SurvivesMigrationFromLegacyJSON|TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode|TestGatewayLaunchEnvPreservesInheritedEffort'
exit=0
--- PASS: TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode (0.01s)
--- PASS: TestGatewayLaunchEnvPreservesInheritedEffort (0.01s)
--- PASS: TestLiveReadersUnchangedByHistoryVerb (4.28s)
--- PASS: TestTodoDone_UndoneRoundTripIsByteIdentical (1.73s)
--- PASS: TestTodoUndone_SurvivesMigrationFromLegacyJSON (0.83s)
ok  	github.com/modu-ai/moai-adk/internal/cli	7.776s
```

같은 트리에서 환경만 바꿔 3건이 PASS로 돌아섰으므로 이 카드의 코드 변경이 원인이 아니다. 이 카드의 테스트 2건은 오염 환경·scrub 환경 양쪽에서 PASS다.

**`TestAppServerSubprocessHTTPToolContinuation` = 이 변경과 구조적으로 무관.** scrub 환경 단독 재실행에서도 결정적으로 실패한다(`appserver_integration_test.go:71: App Server start failed` ×2, exit 1). 귀속 근거는 두 측정이다:

```
$ git diff --stat -- internal/gateway | wc -l
0
$ go list -deps ./internal/gateway/ | grep -c 'internal/cli$'
0
```

`internal/gateway` 테스트 바이너리는 이 카드가 바꾼 파일(전부 `internal/cli`)을 하나도 컴파일하지 않는다. 원인(파이썬 fake-codex 서브프로세스 기동이 이 호스트에서 실패) 자체는 **미측정 가설**이며 이 카드 소관이 아니다. 순수 기준 트리에서의 재현은 돌리지 않았다 — Gaps 참조.

## Mutant — 실제 주입·실행

매 회차 `.go` 원본을 scratchpad 사본에서 복원했고, 마지막 복원 뒤 md5 동일성을 확인했다: `launcher.go 79926ed8bf8613caf86b6365475949d7` / `launch_effort_settings.go 15f03df2959c2c30a8d73e522a5c2aca` / `gateway_settings.go 7668aa1ec3eac9ebb8c40dc1f0e4c630`. 실행 명령은 모두 위 RED와 같은 2개 선택자다.

| # | 주입한 mutant | exit | 관측 |
|---|---|---|---|
| M1 | 게이트웨이일 때 `launchEnv = buildEnvForLaunch(effectiveEffort, launchEnv)` 재고정 (env pin) | 1 | `--- FAIL: TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode` (`:210 gained CLAUDE_CODE_EFFORT_LEVEL=high` ×3) + `--- FAIL: TestGatewayLaunchEnvPreservesInheritedEffort` (`:266 = "high", want the inherited xhigh` ×3) |
| M2 | 게이트웨이일 때 물려받은 `CLAUDE_CODE_EFFORT_LEVEL=` 항목 제거 (scrub) | 1 | `--- FAIL: TestGatewayLaunchEnvPreservesInheritedEffort` (`:263 appears 0 times` ×3, `:266 = ""` ×3). 다른 테스트는 PASS — 의도대로 역할이 나뉜다 |
| M3 | `applyLaunchEffort`가 `payload[effortSettingsKey]` 대입 생략 (payload key drop) | 1 | `--- FAIL: TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode` (`:218 gateway launch carries no --settings payload: []` ×3 — effort만 있던 payload가 비어 주입 자체가 사라짐) |
| M4 | `prepareGatewayOverlay`가 병합 전 `delete(doc, "effortLevel")` (게이트웨이 병합에서 키 유실) | 1 | `--- FAIL: TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode` (`:232 gateway child settings effortLevel = <nil>, want high` ×3) |

M1·M2가 양방향(심기/지우기)으로 빨간불이므로 env 단정은 동어반복이 아니다. M3·M4는 payload 단정이 주입 지점과 게이트웨이 병합 지점을 각각 잡는다는 것을 보인다.

## Gaps — 관측하지 않은 것

| 항목 | 사유 |
|---|---|
| 3모드 실물 기동(`moai gpt` 및 claude/glm 게이트웨이 바인딩)에서 `/effort` 변경이 먹히는지 | 개발 프로젝트에서 실물 런치는 실제 설정 파일을 고쳐 쓰므로 실행 금지. 기동 없이 argv/env/payload를 보는 dry-run seam도 제품 코드에 없다. 대신 **테스트 seam**(`gatewayLaunchBinding.Prepare`)에서 바인딩이 받는 env·args·`--settings` 파일·병합 결과를 3모드 모두 관측했다. 실물 확인 자체는 **미관측** |
| 실제 `newGatewaySessionBinding` + `gateway.StartChild`를 거친 자식 `settings.json`에 `effortLevel`이 있는지 | 병합 함수는 실제 코드로 실행했고(M4가 그 지점을 잡음), 파일 쓰기(`newOwnedOverlay`)는 바이트를 그대로 쓴다는 것을 판독으로만 확인했다. 실자식 기동 경로 end-to-end 단정은 없다 |
| 게이트웨이 세션에서 Claude Code가 `effortLevel`을 요청 본문 `effort`로 싣는지 | 어댑터 쪽(`native_policy.go`)은 판독했으나, 이 Claude Code 버전이 게이트웨이 `ANTHROPIC_BASE_URL` 대상 요청에 effort를 싣는 동작은 이 런에서 캡처하지 않았다. 이 카드 이전에도 env 경로가 어댑터에 닿지 않았으므로 **이 변경으로 달라지는 부분은 아니다** |
| `TestAppServerSubprocessHTTPToolContinuation`의 기준 트리 재현 | 기준 커밋 트리에서 돌리지 않았다. 귀속은 "diff 0 + 의존 없음"의 구조 측정으로 세웠고, 실패 원인은 규명하지 않았다 |
| todo 3건의 기준 트리 오염 환경 재현 | 돌리지 않았다. 환경 scrub만으로 PASS 전환을 관측했다 |
| 크로스 플랫폼 빌드(`GOOS=windows`) | 실행하지 않았다. 변경은 플랫폼 분기를 건드리지 않으며, 판정은 CI 몫 |
| 워크트리 LSP 진단 | 판정 근거로 쓰지 않았다 |

## Residual-risk

- **`buildEnvForLaunch`는 이제 제품 호출부가 없다.** 테스트 5건(`TestBuildEnvForLaunch*`)만 쓴다. 승인 없는 삭제는 범위 규율 위반이라 남기고, godoc·`@MX:NOTE`에 "호출부 없음, 런치 경로에 되살리지 말 것"을 명시했다. 삭제 여부는 리드 판정 사안이다.
- **알려진 우회로(t595와 같은 모양).** env 단정은 바인딩이 받는 `Inherited`를 본다. 누군가 `Prepare` 이후 단계(예: `prepared.ChildEnv`)에서 변수를 다시 심으면 이 테스트는 잡지 못한다. 현재 `prepareGatewayLaunch`의 scrub 목록·추가 목록에는 이 키가 없음을 판독으로 확인했다.
- 운영자가 `--settings`를 직접 넘기면 effort 주입이 빠진다(`operatorSuppliedSettings`). t595와 동일한 의도된 동작이며, 게이트웨이도 이제 같은 규칙을 따른다 — 전에는 그 경우에도 env 경로로 effort가 들어갔으므로 **운영자 `--settings` + 게이트웨이** 조합에서는 프로필 effort가 더 이상 적용되지 않는다는 행동 변화가 있다.
- 부모 셸이 이미 `CLAUDE_CODE_EFFORT_LEVEL`을 갖고 있으면(예: 고정된 세션 안에서 `moai gpt` 실행) 자식은 여전히 막힌다. 새 고정은 만들어지지 않으므로 과도기 비용이다(t595와 동일한 설계 결정).
- 게이트웨이 어댑터는 effort를 `high`(GPT native는 `medium`도)만 허용한다(`native_policy.go:51`). 세션 중 다른 값으로 바꾸면 이제 요청 단계에서 거부될 수 있다 — env 고정이 그 가능성을 가리고 있었을 뿐 이 변경이 만든 제약은 아니다. 미측정.

## 적용 규칙

- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 1·4, §2 — 실물 기동은 Claim이 아니라 Gaps로 기재. "어댑터가 env를 읽는다"는 전제는 전수 grep으로 반증
- `moai-workflow-tdd` Test-First Anti-Cheat Invariant i — RED 출력 원문 첨부
