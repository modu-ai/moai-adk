# t595 — `moai cc` launcher pins effort into `CLAUDE_CODE_EFFORT_LEVEL`

카드 t595. 브랜치 `WT-effort-env-pin`, 기준 `eabce74448e094dd1a4393044138a04b2016af99` (origin/develop).

**상태: 구현 완료, 검증 완료 (2026-09-12 18:33~19:05, load 38.41→23.26 구간).** 앞 세션의 부하 홀드로 미뤄졌던 검증을 이 런에서 전부 실행했습니다. 실행 결과는 아래 § 검증 실행 결과, 남은 미관측은 § Gaps 입니다.

---

## Claim

`moai cc`(및 `glm` / `cg`) 런처가 프로필의 effort를 `CLAUDE_CODE_EFFORT_LEVEL` 환경변수로 심어, Claude Code가 세션 도중의 `/effort`·`/model` effort 변경을 거부하게 만든다. 수정은 환경변수 주입을 없애고, 런처가 이미 주입하는 transient `--settings` 파일에 `effortLevel` 키로 실어 보내는 것이다.

## Evidence — 재현 (2026-09-12, Claude Code v2.1.269)

격리 환경에서만 측정했다. 임시 `CLAUDE_CONFIG_DIR` + 임시 프로젝트를 쓰고, 실제 세션 설정(`.claude/settings.local.json`, `~/.moai`)은 읽기만 했다.

### 코드 경로

`git show origin/develop:internal/cli/launcher.go`

- `:820` — `launchEnv = buildEnvForLaunch(effectiveEffort, os.Environ())` (Claude 백엔드 분기)
- `:1138` — `buildEnvForLaunch`가 `CLAUDE_CODE_EFFORT_LEVEL` 항목을 심거나 교체
- `:1166` — `resolveLaunchEffort(prefs.EffortLevel, prefs.ModelPolicy)`

### 비대화형 대조 3조

판독 지점: 세션 transcript의 `assistant` 행 `effort` 필드. 공통 조건 `--model sonnet`, 프로젝트 `.claude/settings.json`에 `effortLevel` 지정.

| # | settings `effortLevel` | env `CLAUDE_CODE_EFFORT_LEVEL` | 관측된 `effort` |
|---|---|---|---|
| A | `high` | (unset) | `high` |
| B | `high` | `medium` | `medium` |
| C | `xhigh` | `medium` | `medium` |

A가 대조군이다 — 환경변수가 없을 때는 settings 값이 그대로 적용된다. B·C에서 환경변수가 settings를 덮는다.

### 대화형 관측 — Claude Code 자신의 진술

tmux 세션, env `medium` / settings `xhigh`. 기동 배너:

```
Sonnet 5 with medium effort · Claude Max
```

`/effort`에서 xhigh를 선택했을 때의 응답 (verbatim):

```
CLAUDE_CODE_EFFORT_LEVEL=medium overrides this session — clear it and xhigh takes over
```

즉 이 결함은 추론이 아니라 런타임이 문장으로 진술한다.

### 수정 방향 사전 검증

환경변수를 완전히 scrub하고 `--settings <임시파일>`에 `{"effortLevel":"medium"}`만 주입:

- 기동 배너: `Sonnet 5 with medium effort` — 기동 기본값은 유지된다
- `/effort` → xhigh: `Set effort level to xhigh (saved as your default for new sessions)` — 세션 중 덮어쓰기가 **허용된다**

`--settings` 경로가 카드의 목적(기동 기본값 유지 + 세션 중 변경 가능)을 만족함을 착수 전에 확인했다.

### 실사용 확인 (읽기 전용)

- `~/.moai/claude-profiles/moai-adk/settings.json` → `effortLevel: high`
- 리드 세션 env → `CLAUDE_CODE_EFFORT_LEVEL=medium`, `$CLAUDE_EFFORT=medium`

## Baseline-attribution

- `git rev-parse origin/develop` → `eabce74448e094dd1a4393044138a04b2016af99`
- `claude --version` → `2.1.269 (Claude Code)`
- 위 측정은 전부 이 런에서, 이 트리 기준으로 수행했다.

---

## 구현

| 파일 | 변경 |
|---|---|
| `internal/cli/launch_effort_settings.go` (신규) | `applyLaunchEffort(payload, profileName)` — 프로필 effort를 주입 payload에 `effortLevel` 키로 병합. 해석은 기존 `resolveLaunchEffort` 그대로(explicit `effort_level` 우선, `model_policy` 폴백, 둘 다 비면 주입 없음). 프로필 read 실패는 fail-open. 여기에 `buildEnvForClaudeLaunch(base)` seam도 둔다 — 반환값은 `base` 그대로이고, 그 **불변식**(Claude 경로는 `CLAUDE_CODE_EFFORT_LEVEL`을 심지도 지우지도 않는다)을 단정 가능한 지점으로 만드는 것이 존재 이유다 |
| `internal/cli/launcher.go` | Claude 백엔드 분기에서 `buildEnvForLaunch` 호출 제거 → `launchEnv = buildEnvForClaudeLaunch(os.Environ())`. 죽은 `buildEnvForLaunch` 삭제. `appendCrossSessionSettings`에 `profileName` 전달 |
| `internal/cli/crosssession_settings.go` | 일반 런치 주입 payload = crosssession 번역 ∪ 프로필 effort. 주입 지점은 하나 유지 |
| `internal/cli/kanban_settings.go` | 칸반/팩토리 레인 payload에도 동일 병합 (`prepareKanbanSettings(profileName, args)`) |
| `internal/cli/cc.go`, `glm.go` | `prepareKanbanSettings` 호출부 8곳에 `profileName` 전달 |
| `.claude/rules/.../settings-management.md` + 템플릿 미러 | `effortLevel` 행 정정 — 환경변수는 override라 세션 중 변경을 막는다는 사실 명시 |

### 설계 결정 — 물려받은 환경변수는 건드리지 않는다

런처는 더 이상 `CLAUDE_CODE_EFFORT_LEVEL`을 **심지 않지만**, 바깥에서 물려받은 값은 제거하지 않는다. 운영자 확인 후 채택(2026-09-12).

- 근거: `CLAUDE_CODE_EFFORT_LEVEL=max moai cc`는 `settings-management.md`가 안내하는 per-session override다. scrub하면 문서화된 탈출구가 사라진다.
- 대가: 이미 pin이 박힌 세션의 터미널에서 `moai cc`를 띄우면 자식 세션이 환경변수를 상속해 여전히 막힌다. 다만 **새 pin은 만들어지지 않으므로** 기존 세션이 종료되면 연쇄가 끊긴다 — 과도기 비용이며 영구 결함이 아니다.

### 카드 전제 정정

카드 본문은 원인을 "preferences.yaml `effort_level`"이라고 적었으나, 실제 경로는 두 갈래다: `resolveLaunchEffort`는 explicit `effort_level`이 없으면 `model_policy`에서 파생한다. 실사용 프로필(`~/.moai/claude-profiles/moai-adk/preferences.yaml`)에는 `effort_level` 키가 **없고** `model_policy: high`만 있다. 수정은 두 갈래 모두를 덮는다.

## 검증 실행 결과 (2026-09-12, 이 런)

전부 워크트리 `.claude/worktrees/t595` 안에서, 브랜치 `WT-effort-env-pin` 기준으로 실행했다.

| 명령 | exit | 관측 |
|---|---|---|
| `go build ./internal/cli/...` | 0 | 출력 없음 |
| `go vet ./internal/cli/...` | 0 | 출력 없음 |
| `golangci-lint run ./internal/cli/...` | 0 | `0 issues.` |
| `make build` | 0 | `catalog.yaml updated successfully (12899 bytes)` + `-o bin/moai`. 직후 작업 트리 상태 조회에 추가 변경 파일 없음 |
| `go test ./internal/template/... -count=1` | 0 | `internal/template 46.953s` / `agentemit 6.624s` / `commandemit 6.759s` 전부 `ok` |
| `go test ./internal/cli/... -count=1 -timeout 40m` | 0 | 17개 패키지 전부 `ok` (`internal/cli` 1307.734s). raw 로그에서 `^--- FAIL` 및 `^FAIL` 매치 0건 |

### 겨냥 실행 — 선택자가 실제로 무엇을 골랐는지

`go test ./internal/cli/ -count=1 -v -run '<12개 패턴>'` → exit 0, `=== RUN Test` 50행, 최상위 `--- PASS` 27건, `--- FAIL` 0건.

선택자가 없는 이름을 조용히 버리는 함정을 피하려고, 개별 지정한 9개 이름이 실제로 실행됐는지 `--- PASS: <name> (` 형태로 각각 대조했다 — 9/9 전부 1건씩 적중:

`TestApplyLaunchEffort` · `TestResolveLaunchEffort` · `TestClaudeLaunchEnvPreservesInheritedEffort` · `TestLaunchEffortNeverInjectedWhenOperatorSuppliedSettings` · `TestLaunchEffortReachesGeneralInjection` · `TestLaunchEffortReachesKanbanInjection` · `TestACFM023c_KanbanEnvReachesChildEnvironment` · `TestAppendCrossSessionSettingsGeneralLaunch` · `TestOperatorSuppliedSettingsDetectsAllForms`

### Mutant 확인 — 이번엔 실제로 실행했다

앞 판본은 mutant 결과를 서술했으나 그 시점에 테스트를 한 번도 돌리지 않았으므로 그것은 미관측 주장이었다. 이 런에서 실제로 주입하고 관측했다. 매 회차마다 `launch_effort_settings.go` 원본을 복원했고, 마지막 복원 뒤 md5 동일성을 확인했다 (`1875d88ce974f301417aa5c522c660bc`).

| # | 주입한 mutant | 대상 테스트 | 관측 |
|---|---|---|---|
| M1 | `buildEnvForClaudeLaunch` 가 `CLAUDE_CODE_EFFORT_LEVEL=` 항목을 걸러냄 (scrub) | `TestClaudeLaunchEnvPreservesInheritedEffort` | exit 1, `--- FAIL` |
| M2 | 같은 함수가 `CLAUDE_CODE_EFFORT_LEVEL=medium` 을 덧붙임 (pin) | 같음 | exit 1, `--- FAIL` |
| M3 | `applyLaunchEffort` 가 `payload[effortSettingsKey]` 대입을 생략 | `TestApplyLaunchEffort` / `...ReachesGeneralInjection` / `...ReachesKanbanInjection` | exit 1, 3건 전부 `--- FAIL` |

M1·M2 가 양방향으로 빨간불이므로 이 단정은 `os.Environ()` 동어반복이 아니다 — 값이 사라져도, 값이 바뀌어도 잡는다.

## Gaps — 아직 관측하지 않은 것

| 항목 | 사유 |
|---|---|
| 수정 후 `moai cc` 실물 기동에서 `/effort` 변경이 먹히는지 | 재빌드된 `bin/moai` 는 있으나 실물 기동은 이 세션의 런처를 갈아치우므로 실행하지 않았다. 착수 전 대조 실험(§ 수정 방향 사전 검증)이 `--settings` 경로의 동작을 이미 보였고 이 카드의 변경은 그 경로로 값을 옮기는 것이지만, **런처 경유 실물 확인 자체는 미관측**이다 |
| GLM 백엔드 경로(`ANTHROPIC_REASONING_EFFORT`) | 리드가 범위 밖으로 지정 |
| 워크트리 LSP 진단 | 워크트리가 go workspace에 없어 전량 위양성 — 판정 근거로 쓰지 않았다 |
| 크로스 플랫폼 빌드(`GOOS=windows`) | 이 런에서 실행하지 않았다. 변경은 플랫폼 분기를 건드리지 않으나, 판정은 CI 몫 |

### 커밋에서 제외한 것 — 카드와 무관한 로컬 설정 드리프트

작업 트리에 t595 와 무관한 수정 2본이 함께 있었다. 손대지 않고 **커밋에서 제외**했다 (범위 규율):

- `.moai/config/sections/gate.yaml` — `gate.pre_commit.enabled: true` 추가
- `.moai/config/sections/workflow.yaml` — audit codex 모델 `gpt-5.6-sol`/`high` → `gpt-6-astra`/`low`, `audit.gates` + `audit.model: codex` 추가, worktree `auto_cleanup`/`auto_create` false → true

출처 불명이며 이 런에서 만든 것이 아니다. 처분은 리드 몫.

### 도구 관측 (별건) — 설치본 지연, 미구현 아님

첫 관측은 `moai slot status` → `Unknown command "slot" for "moai"` 였고, 그때 나는 이를 "독트린이 지시하는 명령이 설치본에 없다"로만 적었다. 리드는 이를 **미구현**으로 판정했다(2026-09-12). 그 판정은 틀렸고, 아래가 실측이다.

| 대상 | `slot status` | 관측 |
|---|---|---|
| `./bin/moai` (이 워크트리, `make build` 직후) | exit 0 | `no slot leases recorded` |
| `~/go/bin/moai` (설치본) | 비정상 | `Unknown command "slot" for "moai".` |

소스에도 있다 — `internal/cli/slot.go:84` `newSlotCmd()`, `:328` `rootCmd.AddCommand(newSlotCmd())`. 독트린 참조도 있다 — `.claude/rules/moai/workflow/kanban-dispatch.md:209`, 전용 룰 `.claude/rules/moai/workflow/resource-slot-lease.md`, `.claude/rules/local/gitflow-lane-protocol.md:100`. 기능 자체는 `SPEC-RESOURCE-SLOT-LEASE-001`(status: completed, 카드 t607)로 이미 착지해 있다.

진단은 단순한 "낡은 빌드"보다 한 칸 더 구체적이다 — **브랜치 계열 차이**다. 객체 존재 확인(`cat-file -e`)을 이 런에서 양쪽 ref 에 직접 돌렸다:

- `develop:internal/cli/slot.go` → exit 0 (존재)
- `main:internal/cli/slot.go` → exit 128, `exists on disk, but not in 'main'`

즉 t607 은 develop 에 착지했고 main 에는 아직 없다. `~/go/bin/moai` 를 포함한 main 계열 빌드는 **릴리스 전까지 이 명령을 갖지 않는 것이 정상**이며, 고장이 아니다. 무거운 실행 직렬화가 필요하면 develop 계열 사본(`./bin/moai`)을 쓰거나 `make install` 로 갱신한다.

첫 관측의 실수는 결론이 아니라 측정 대상이었다 — PATH 의 설치본 하나만 재고 "없다"로 적었고, 갓 빌드한 트리 사본을 대조하지 않았다. 같은 이름의 바이너리가 트리마다 다른 시점·다른 계열의 빌드라는 것이 이 축의 함정이며, `CLAUDE.local.md` §11 이 경고하는 바로 그 지점이다.

교정: 무거운 실행 직렬화가 필요하면 `~/go/bin/moai` 를 `make install` 로 갱신하면 된다. 이 런은 레인 단독이라 슬롯이 필요 없었으므로 t595 절차에는 영향이 없다.

## 병합 (2026-09-12, 통합 창 lane-7)

리드 지명으로 창을 받아 로컬 develop `5ddccacc9` 를 흡수했다. 기준은 `eabce7444` 였으므로 35커밋을 흡수한다.

### 리드 판정 — A안 (2026-09-12)

흡수 전 정적 정찰에서 t649(`ce79ef7ca`, GPT 게이트웨이)와 **의미 충돌**을 발견해 창 진입 전에 보고했다. t649 는 백엔드 분기를 `if glmBackend && binding == nil / else` 로 바꿔, t595 가 겨냥한 else 갈래가 이제 평문 Claude 와 **모든 게이트웨이 바인딩**을 함께 덮는다. 그리고 t649 는 그 갈래의 env 주입을 단정하는 테스트를 함께 실었다 — `TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode`(claude/gpt/glm 3모드).

세 선택지를 올렸고 리드가 **A안**을 택했다(2026-09-12): 게이트웨이에는 env 주입을 남기고 평문 Claude 경로에서만 제거한다.

- 채택 근거: 운영자의 긴급 대상은 일반 `moai cc` 세션의 `/effort` 무력화이며, A가 그것을 즉시 던진다.
- B안(t649 단정을 settings payload 로 겨냥 이동)은 **타 카드의 테스트**를 창 안에서 고치는 사안이고 어댑터 실측이 선행돼야 해 창 밖 소관으로 분류됐다.
- C안(어댑터가 주입된 settings 를 읽게 개조)은 t649 소관의 큰 변경으로 범위 밖이다.
- 게이트웨이 반쪽은 후속 카드 **t668** 로 발행됐다.

### 충돌 해결

| 파일 | 해결 |
|---|---|
| `internal/cli/launcher.go` | 분기를 **3갈래**로: GLM(`glmBackend && binding == nil`) / 게이트웨이(`binding != nil`, `buildEnvForLaunch` 유지) / 평문 Claude(`buildEnvForClaudeLaunch`). `effectiveEffort` 는 앞의 두 갈래가 모두 쓰므로 develop 처럼 분기 바깥에서 선언 |
| `internal/cli/cc.go` (2곳) | develop 의 `backend` 변수(t649) + 내 `profileName` 인자 **조합** |
| `internal/cli/glm.go`, `launcher_test.go`, 템플릿 룰 미러 | 자동 병합 |

### [중요] 자동 병합이 삼킨 것 — 복원 항목

**`launcher_test.go` 는 충돌 없이 자동 병합됐고, 그 과정에서 내 삭제가 조용히 적용됐다.** A안은 `buildEnvForLaunch` 를 게이트웨이 갈래에서 계속 쓰므로 이는 **살아 있는 함수의 테스트가 사라진 상태**다. 빌드도 테스트도 초록이라 diff 를 읽지 않으면 드러나지 않는다.

흡수 **전** 정적 정찰에서 이 손실을 예측해 두었기에 잡았다. 복원 목록:

| 복원 대상 | 위치 | 성격 |
|---|---|---|
| `buildEnvForLaunch()` | `internal/cli/launcher.go` | b0bf31c4e 가 삭제했던 함수. A안에서 게이트웨이 갈래가 사용 |
| `TestBuildEnvForLaunch` | `internal/cli/launcher_test.go` | 상동 |
| `TestBuildEnvForLaunch_EmptyEffort` | `internal/cli/mcp_doctor_coverage_test.go` | 상동 |
| `TestBuildEnvForLaunch_AddsNewEntry` | 같음 | 상동 |
| `TestBuildEnvForLaunch_ReplacesExisting` | 같음 | 상동 |
| `TestBuildEnvForLaunch_PreservesOtherVars` | 같음 | 상동 |

복원하며 **주석 하나를 정정했다**: develop 판 `buildEnvForLaunch` 의 `@MX:NOTE` 는 자신을 "Effort injection point (Claude backend)" 라고 적는데, A안에서 그 함수는 게이트웨이 갈래 전용이 되므로 거짓이다. `(gateway backend)` 로 고치고, 평문 Claude 는 settings payload 로 옮겨갔다는 설명을 함께 달았다. 스테일한 주장을 그대로 되살리지 않기 위한 것이다.

### blockcap 주석 재작성 (before / after)

`internal/cli/launcher_blockcap_infinite_test.go` `TestACFM023c_KanbanEnvReachesChildEnvironment` 의 근거 주석은 A안에서 거짓이 된다. 단정 자체는 그대로 두고 근거만 고쳤다.

- **before**: "The Claude path no longer wraps os.Environ() at all — ... the drop hazard this AC guards therefore survives only in the GLM wrapper"
- **after**: "The PLAIN Claude path no longer wraps os.Environ() at all — ... Two wrappers still filter the inherited environment and so still carry the drop hazard: buildEnvForGLMLaunch, and buildEnvForLaunch on the gateway branch"

## 병합 트리 재측정 (2026-09-12)

전부 병합 트리(`WT-effort-env-pin`, MERGE_HEAD `5ddccacc9`)에서 실행했다. **이전 런(기준 `eabce7444`)의 수치와 비교하지 않는다** — 트리가 35커밋 다르고 아래 사전 귀속 red 12건이 섞이므로 두 수치는 같은 것을 재고 있지 않다.

| 검증 | exit | 관측 | 비고 |
|---|---|---|---|
| `go build ./internal/cli/...` | 0 | 출력 없음 | |
| `go vet ./internal/cli/...` | 0 | 출력 없음 | |
| **조건1** `-run TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode` | 0 | `--- PASS` + 서브테스트 claude/gpt/glm 3건 전부 PASS | A안 구현 확인. 최우선 측정 |
| **조건2** 뮤턴트 M1 (scrub) | 1 | `--- FAIL: TestClaudeLaunchEnvPreservesInheritedEffort` | 분기 로직 전환 후에도 유효 |
| **조건2** 뮤턴트 M2 (pin) | 1 | 같음 | 복원 후 md5 `1875d88ce974f301417aa5c522c660bc` 동일 |
| `go test ./internal/cli/...` | 1 | `internal/cli` 1435.427s, 하위 16패키지 전부 `ok`. FAIL 은 **사전 귀속 12건뿐** | load 10.31 시점 |
| `go test ./internal/template/...` | 0 | `template 37.358s` / `agentemit 0.904s` / `commandemit 1.214s` | load 8.25 시점 |
| `golangci-lint run ./internal/cli/...` | 1 | 31 issues (errcheck 30 / staticcheck 1) | 사전 귀속 — 아래 |

### 사전 귀속 red 12건 — 실측으로 확정

리드는 gateway 계열 3건을 사전 귀속으로 통보했으나, 병합 트리 전량 측정에서 실제 FAIL 은 **12건**이었다. 통보를 전제로 받지 않고 순수 develop 에서 직접 쟀다.

측정: develop 워크트리, HEAD `5ddccacc9`, 추적 파일 수정 0 확인 후
`go test ./internal/cli/ -count=1 -v -timeout 20m -run '<12개 이름>'` → exit 1, `=== RUN` 14행, `--- PASS` 2, `--- FAIL` 12.

RUN 이 14인 것은 선택자 접두사가 `TestRunDoctor_VerboseMode` 와 `_WithFixFlag` 를 추가로 집었기 때문이며 그 둘은 PASS 다. 겨냥한 12개 이름은 이름별로 대조해 전부 실행·전부 FAIL 임을 확인했다.

| 계열 | 테스트 | 귀속 |
|---|---|---|
| gateway (리드 통보분) | `TestCodexCommand_RegisteredInLaunchGroup`, `TestCharacterize_GLM_WarningPrintedToStderr`, `TestNoBareGLMEnvVarLiteralsInCLIProduction` | develop 소관, 카드 t669 |
| doctor (통보 밖, 내가 발견) | `TestRunDoctor_{WithExport,WithFix,Verbose,AllFlags,VerboseAndDetail,ExportMode}`, `TestDoctorCmd_{Execution,ExportFlag,VerboseExecution}` | develop 소관, 카드 **t675**(이 측정으로 발행) |

doctor 9건은 전부 `runDoctor error: doctor: 1 check(s) failed` 이고, 병합 트리에서 빌드한 바이너리로 doctor 를 직접 돌린 결과 실패 검사는 `Agent Emit Embed`("moai embeds stale agent-emit artifacts (11/11 compared): manager-lead.toml, super-advisor.toml, sync-auditor.toml")와 `Harness 5-Layer` 다. 동시에 병합 트리의 `make agents-emit-check` 는 exit 0 이다 — 즉 어긋난 것은 소스 축이 아니라 **임베드 축**이다. 둘이 동시에 참인 기전은 확정하지 못했다(가설: doctor 의 비교 기준이 로컬 도그푸드 사본 C1 이라 의도된 C1↔C2 분기에서 항상 실패한다 — **미측정 가설**이며 t675 소관).

린트 31건도 같은 방식으로 귀속을 세웠다: 지적 위치가 전부 내가 건드리지 않은 게이트웨이·migrate 파일이었고, 순수 develop 워크트리에서 `golangci-lint run ./internal/cli/...` 를 돌려 **동일하게 31 issues (errcheck 30 / staticcheck 1)** 를 관측했다. 내 병합이 늘린 것이 아니다.

"내 변경과 무관"은 파일명 추정이 아니라 이 두 측정이 근거다. 12건 중 하나라도 순수 develop 에서 GREEN 이었다면 그것은 내 소관이었고, 그 경우 중단·보고가 사전 합의된 판정식이었다.

## Residual-risk

- **[A안 잔여] 게이트웨이 세션은 여전히 `/effort`·`/model` 중 effort 변경이 막힌다.** A안이 `binding != nil` 갈래에 `CLAUDE_CODE_EFFORT_LEVEL` 주입을 남겼기 때문이며, 그 변수는 Claude Code 가 세션 중 변경을 거부하는 override 다. 즉 t595 가 고친 결함은 **평문 Claude 경로에서만** 사라졌고, `moai gpt` 및 게이트웨이 바인딩을 타는 launch 에서는 그대로 남아 있다. 후속 카드 **t668** 소관. 이것은 "게이트웨이 경로가 이 결함에서 면제된다"는 뜻이 **아니다** — 같은 기전이 같은 증상을 낸다.
- 게이트웨이 effort 경로에 대한 내 정적 관측(어댑터가 env 가 아니라 요청 본문에서 effort 를 읽는다 — `internal/gateway/translate/native_policy.go:47-55`, `:139`)은 **부분 판독이며 측정이 아니다.** t668 에서 확인할 가설로만 취급한다.

- 운영자가 직접 `--settings`를 넘긴 경우 effort 주입이 빠진다(`operatorSuppliedSettings` 분기). 의도된 동작이며 운영자 의사 우선이지만, 그 경로에서는 프로필 effort가 적용되지 않는다.
- `--settings`는 Claude Code 우선순위에서 사용자 settings보다 위다. 따라서 `/effort`가 "새 세션 기본값"으로 저장한 값은 다음 기동에서 다시 주입 payload에 눌린다. 카드가 목표한 **세션 중** 변경은 성립하지만, "저장한 기본값이 다음 세션까지 간다"는 기대는 성립하지 않는다. 리드 판정(2026-09-12): 기재는 승인, t595 범위 밖 — 별도 카드로 발행.
- `AC-FM-023c`(칸반 env가 자식 환경에 도달)는 이제 **양쪽 반쪽으로 덮인다.** GLM 반쪽은 기존 `TestACFM023c_KanbanEnvReachesChildEnvironment`의 겨냥을 `buildEnvForGLMLaunch`로 옮긴 것(래퍼가 남아 있는 유일한 경로), Claude 반쪽은 신규 `TestClaudeLaunchEnvPreservesInheritedEffort`다. 후자는 동일성이 아니라 **행동**을 단정한다 — 심어 둔 `CLAUDE_CODE_EFFORT_LEVEL=xhigh`가 Claude 기동 env 빌드를 통과해 값까지 살아남는지. 뮤턴트 확인: scrub하는 래퍼를 넣으면 항목이 사라져 RED, pin하는 래퍼를 넣으면 값이 바뀌어 RED. 리드 지침(2026-09-12)에 따른 보강이며, `os.Environ()` 동어반복이 아니다.
- **`buildEnvForClaudeLaunch` seam의 한계 — 알려진 우회로.** 이 단정은 seam **안에서** 일어나는 변경만 잡는다. 누군가 호출부 자체를 다른 래퍼로 바꿔치우면(`launchEnv = <다른함수>(os.Environ())`) 테스트는 초록으로 남는다. 소스 스캔형 가드를 덧대는 선택지는 리드가 반려했다(2026-09-12): 범위 밖이고, grep은 env 키가 상수나 간접 참조로 바뀌면 조용히 통과하므로 약한 가드가 초록을 근거로 읽히는 편이 더 나쁘다는 판단이다. 따라서 이 우회로는 **막지 않고 기록한다** — 같은 모양(계약 지점을 호출부 교체로 우회)이 다른 카드에서 재발하면 그때 별도 카드로 다룬다.

## 적용 규칙

- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 1, §2 — 미실행 검증을 Claim이 아니라 Gaps로 기재
- `.claude/rules/moai/development/verification-completeness.md` §1.1 — 신규 테스트는 red 관측 전까지 미완성. 위 Gaps가 그 미완성 상태를 명시한다
