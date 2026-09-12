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

따라서 진단은 **바이너리 지연**이다: `~/go/bin/moai` 가 t607 착지 이전 빌드라 명령을 모른다. 첫 관측의 실수는 결론이 아니라 측정 대상이었다 — PATH 의 설치본 하나만 재고 "없다"로 적었고, 갓 빌드한 트리 사본을 대조하지 않았다. `CLAUDE.local.md` §11 이 경고하는 바로 그 축이다.

교정: 무거운 실행 직렬화가 필요하면 `~/go/bin/moai` 를 `make install` 로 갱신하면 된다. 이 런은 레인 단독이라 슬롯이 필요 없었으므로 t595 절차에는 영향이 없다.

## Residual-risk

- 운영자가 직접 `--settings`를 넘긴 경우 effort 주입이 빠진다(`operatorSuppliedSettings` 분기). 의도된 동작이며 운영자 의사 우선이지만, 그 경로에서는 프로필 effort가 적용되지 않는다.
- `--settings`는 Claude Code 우선순위에서 사용자 settings보다 위다. 따라서 `/effort`가 "새 세션 기본값"으로 저장한 값은 다음 기동에서 다시 주입 payload에 눌린다. 카드가 목표한 **세션 중** 변경은 성립하지만, "저장한 기본값이 다음 세션까지 간다"는 기대는 성립하지 않는다. 리드 판정(2026-09-12): 기재는 승인, t595 범위 밖 — 별도 카드로 발행.
- `AC-FM-023c`(칸반 env가 자식 환경에 도달)는 이제 **양쪽 반쪽으로 덮인다.** GLM 반쪽은 기존 `TestACFM023c_KanbanEnvReachesChildEnvironment`의 겨냥을 `buildEnvForGLMLaunch`로 옮긴 것(래퍼가 남아 있는 유일한 경로), Claude 반쪽은 신규 `TestClaudeLaunchEnvPreservesInheritedEffort`다. 후자는 동일성이 아니라 **행동**을 단정한다 — 심어 둔 `CLAUDE_CODE_EFFORT_LEVEL=xhigh`가 Claude 기동 env 빌드를 통과해 값까지 살아남는지. 뮤턴트 확인: scrub하는 래퍼를 넣으면 항목이 사라져 RED, pin하는 래퍼를 넣으면 값이 바뀌어 RED. 리드 지침(2026-09-12)에 따른 보강이며, `os.Environ()` 동어반복이 아니다.
- **`buildEnvForClaudeLaunch` seam의 한계 — 알려진 우회로.** 이 단정은 seam **안에서** 일어나는 변경만 잡는다. 누군가 호출부 자체를 다른 래퍼로 바꿔치우면(`launchEnv = <다른함수>(os.Environ())`) 테스트는 초록으로 남는다. 소스 스캔형 가드를 덧대는 선택지는 리드가 반려했다(2026-09-12): 범위 밖이고, grep은 env 키가 상수나 간접 참조로 바뀌면 조용히 통과하므로 약한 가드가 초록을 근거로 읽히는 편이 더 나쁘다는 판단이다. 따라서 이 우회로는 **막지 않고 기록한다** — 같은 모양(계약 지점을 호출부 교체로 우회)이 다른 카드에서 재발하면 그때 별도 카드로 다룬다.

## 적용 규칙

- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 1, §2 — 미실행 검증을 Claim이 아니라 Gaps로 기재
- `.claude/rules/moai/development/verification-completeness.md` §1.1 — 신규 테스트는 red 관측 전까지 미완성. 위 Gaps가 그 미완성 상태를 명시한다
