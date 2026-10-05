# 카드 t669 판정서 — gateway 계열 red 3건 수리

## Claim (주장)

develop 기반 트리(base `5ddccacc9`, 브랜치 `WT-gateway-red-repair`)에서 t649/t652 gateway 착지가 남긴 red 테스트 3건이 모두 수리됐고, 카드 범위 밖의 나머지 검증 대상 테스트는 전부 green을 유지한다.

- `TestNoBareGLMEnvVarLiteralsInCLIProduction` — `gateway_prepare.go:26`의 bare 리터럴 3개를 `config.EnvClaudeCode*` 상수로 교체해 수리.
- `TestCodexCommand_RegisteredInLaunchGroup` — `moai cg` 은퇴에 뒤처진 stale 테스트를 현행 Launchers 섹션(cc, glm, codex, gpt)에 맞게 갱신해 수리. 코드가 아니라 테스트가 stale이었다.
- `TestCharacterize_GLM_WarningPrintedToStderr` — 동일하게 stale 단언(`"moai cg"` 언급)을 현행 의도 텍스트(`"Mixed Claude/GLM teammate roles"`)로 갱신해 수리.

## Evidence (증거)

### 1. 수리 전 red 재현 (수정 전, 본 실행에서 직접 관측)

```
$ unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED \
  && go test ./internal/cli/ -run 'TestCodexCommand_RegisteredInLaunchGroup|TestCharacterize_GLM_WarningPrintedToStderr|TestNoBareGLMEnvVarLiteralsInCLIProduction' -count=1 -v

    codex_launcher_test.go:236: launcher "cg" missing from the launchers section block (block commands: [cc gpt glm codex web statusline])
--- FAIL: TestCodexCommand_RegisteredInLaunchGroup (0.00s)
    glm_new_test.go:954: stderr should mention 'moai cg' as alternative, got: "WARNING: moai glm uses GLM models for the MAIN SESSION. ... Mixed Claude/GLM teammate roles require verified teammate routing support.\n"
--- FAIL: TestCharacterize_GLM_WarningPrintedToStderr (0.00s)
    glm_env_parity_test.go:115: bare GLM env-var string literal(s) found in production CLI source ...
          internal/cli/gateway_prepare.go:26:254
          internal/cli/gateway_prepare.go:26:383
          internal/cli/gateway_prepare.go:26:427
--- FAIL: TestNoBareGLMEnvVarLiteralsInCLIProduction (0.18s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.535s
```

오케스트레이터가 제시한 3건의 재현 출력과 동일한 실패가 동일 트리에서 재현됐다.

### 2. 수리 내용 (커밋 diff, 3파일 / +5 -5)

- `internal/cli/gateway_prepare.go:26` — `"CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS"` → `config.EnvClaudeCodeDisableExperimentalBetas`, `"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC"` → `config.EnvClaudeCodeDisableNonessentialTraffic`, `"CLAUDE_CODE_TEAMMATE_DISPLAY"` → `config.EnvClaudeCodeTeammateDisplay`. 상수명은 `internal/config/envkeys.go:318/322/327`에서 확인. 같은 줄의 `MOAI_BACKUP_AUTH_TOKEN`, `API_TIMEOUT_MS`, `CLAUDE_CODE_AUTO_COMPACT_WINDOW`, `CLAUDE_CODE_MAX_CONTEXT_TOKENS`, `MOAI_STATUSLINE_CONTEXT_SIZE`는 게이트 금지 목록 밖이므로 손대지 않았다(범위 규율).
- `internal/cli/codex_launcher_test.go:234` — want 목록 `[]string{"cc", "glm", "cg", "codex"}` → `[]string{"cc", "glm", "codex", "gpt"}. 함수 doc comment의 열거(`(cc, glm, cg, codex)`)도 `(cc, glm, codex, gpt)`로 갱신해 주석 정확성을 유지.
- `internal/cli/glm_new_test.go:953-955` — `strings.Contains(got, "moai cg")` → `strings.Contains(got, "Mixed Claude/GLM teammate roles")`, 에러 메시지도 `stderr should mention the mixed Claude/GLM teammate-routing constraint, got: %q`로 갱신.

### 3. 수리 후 재측정 (수정 후, 본 실행에서 직접 관측)

대상 3건:

```
$ unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED \
  && go test ./internal/cli/ -run 'TestCodexCommand_RegisteredInLaunchGroup|TestCharacterize_GLM_WarningPrintedToStderr|TestNoBareGLMEnvVarLiteralsInCLIProduction' -count=1 -v

--- PASS: TestCodexCommand_RegisteredInLaunchGroup (0.00s)
--- PASS: TestCharacterize_GLM_WarningPrintedToStderr (0.00s)
--- PASS: TestNoBareGLMEnvVarLiteralsInCLIProduction (0.08s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.274s
```

영향 가족 1 — `TestCodexCommand_` (9 RUN = 6 최상위 + 3 서브테스트, 최상위 6건 전부 PASS, FAIL 0):

```
$ go test ./internal/cli/ -run 'TestCodexCommand_' -count=1 -v

--- PASS: TestCodexCommand_NeutralityScan (0.00s)
--- PASS: TestCodexCommand_GuardFileLiteralsNeutral (0.02s)
--- PASS: TestCodexCommand_RegisteredInLaunchGroup (0.00s)
--- PASS: TestCodexCommand_HelpExitsZero (0.00s)
--- PASS: TestCodexCommand_HelpCopyGuidance (0.00s)
--- PASS: TestCodexCommand_HelpDescribesReversedDefault (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.184s
```

영향 가족 2 — `TestCharacterize_GLM_` (최상위 10건 전부 PASS, FAIL 0. RUN 12건 중 나머지 2건은 `TestCharacterize_GLM_HelpFlagShortCircuits`의 서브테스트 `--help`, `-h`이며 들여쓰기 된 `--- PASS` 2줄과 정확히 대응해 불일치 없음):

```
$ go test ./internal/cli/ -run 'TestCharacterize_GLM_' -count=1 -v

--- PASS: TestCharacterize_GLM_ModeIsAlwaysGLM (0.00s)
--- PASS: TestCharacterize_GLM_AutoModeRejected (0.00s)
--- PASS: TestCharacterize_GLM_AutoModeEqualsSyntaxRejected (0.00s)
--- PASS: TestCharacterize_GLM_HelpFlagShortCircuits (0.00s)
--- PASS: TestCharacterize_GLM_SetupSubcommandRouted (0.43s)
--- PASS: TestCharacterize_GLM_WarningPrintedToStderr (0.00s)
--- PASS: TestCharacterize_GLM_ContainsPermissionMode_SpaceSyntax (0.00s)
--- PASS: TestCharacterize_GLM_ContainsPermissionMode_EqualSyntax (0.00s)
--- PASS: TestCharacterize_GLM_ContainsPermissionMode_OtherModeNotMatched (0.00s)
--- PASS: TestCharacterize_GLM_ContainsPermissionMode_Empty (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.609s
```

빌드·정적 분석:

```
$ go build ./...        → BUILD_OK  (exit 0)
$ go vet ./internal/cli/ → VET_OK    (exit 0)
```

`golangci-lint run ./internal/cli/`는 31건의 기존 지적(errcheck 30, staticcheck 1)을 보고했다. 전부 `gateway_child.go`, `gateway_factory_test.go`, `gateway_ui_state*.go`, `migrate_cg.go`, `gpt_auth.go` 등 t649/t652 착지 표면의 파일이며, 본 카드가 고친 3파일·3위치에는 한 건도 없다. 즉 카드 변경으로 인한 신규 lint 지적은 0건이다(기존 지적의 귀속은 Gaps 절 참조).

## Baseline-attribution (baseline 귀속)

모든 측정은 본 실행에서, worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t669`, base `5ddccacc9`, 브랜치 `WT-gateway-red-repair`, clean tree에서 수행했다.

- 수리 전 red 재현: 커밋 `5ddccacc9` 작업 트리, `-count=1`, 위 Evidence §1의 명령과 출력.
- 수리 후 재측정: 동일 트리에 위 Evidence §2의 diff만 적용한 상태. `--stat` 확인: `internal/cli/codex_launcher_test.go`, `internal/cli/gateway_prepare.go`, `internal/cli/glm_new_test.go` 3파일 +5 -5.
- 테스트 변경 정당성 근거(카드 [HARD] 요건):
  - Fix 2(`codex_launcher_test.go`): `internal/cli/cg.go:9`에 `var errCGRetired = errors.New("moai cg is retired; run moai migrate cg to preview an explicit teammate-role migration")`가 존재하고, 이는 `ce79ef7ca`(t649, "feat(gateway): integrate GPT launcher, authentication and runtime repairs")가 도입했다. 같은 커밋이 루트 help의 Launchers 섹션에서 cg를 빼고 gpt/web/statusline을 넣었다 — 섹션 등록 제거는 의도적 은퇴이므로 stale인 쪽은 테스트다.
  - Fix 3(`glm_new_test.go`): `git show ce79ef7ca -- internal/cli/glm.go`에서 `"For hybrid mode (Claude lead + GLM teammates), use 'moai cg' instead."` 및 `"If you want Claude as leader and GLM for teammates, use 'moai cg' instead."`가 제거되고 `+ "Mixed Claude/GLM teammate roles require verified teammate routing support."`가 추가된 것을 직접 관측했다. 마이그레이션 안내는 glm 경고가 아니라 cg 자체의 은퇴 에러(`moai migrate cg`)가 담당한다.
  - Fix 1(`gateway_prepare.go`): 상수 존재는 `internal/config/envkeys.go:318`(`EnvClaudeCodeDisableExperimentalBetas`), `:322`(`EnvClaudeCodeDisableNonessentialTraffic`), `:327`(`EnvClaudeCodeTeammateDisplay`)에서 grep으로 확인.

## Gaps (미검증)

- `./internal/cli/` 전체 스위트(~26분)를 로컬에서 돌리지 않았다 — 카드 재측정 범위는 영향 가족 2개 + build + vet + lint이며, 전체 판정은 develop push 시 CI의 몫이다.
- `go test -race`를 돌리지 않았다.
- golangci-lint의 31건 기존 지적에 대해 base 트리에서의 대조 실행(base에서 동일 명령 재실행)은 하지 않았다. 다만 31건 전부가 카드가 건드리지 않은 파일·라인에 있다는 것은 diff(위 Evidence §2)와 lint 출력의 라인 대조로 직접 확인했고, `gateway_prepare.go`·`codex_launcher_test.go`·`glm_new_test.go`에는 지적이 0건이다.
- 영향 가족 외 패키지(`internal/gateway` 등 gateway 하위 패키지)의 테스트는 이 카드에서 돌리지 않았다 — 3파일 모두 `internal/cli` 내부 변경이고 `go build ./...`가 전 트리 컴파일을 보증한다.

## Residual-risk (잔여 위험)

- lint 31건은 t649/t652 착지가 남긴 기존 기술부채로 판단되지만, CI의 lint 게이트가 이를 error로 승격시키면 develop push가 막힐 수 있다. 이 경우 수리 주체는 본 카드가 아니라 gateway 착지 카드다.
- `-race` 미측정이므로 변경 라인 주변의 경합 문제는 이 실행으로는 관측되지 않는다(변경이 상수 치환과 테스트 단언 갱신뿐이라 경합 표면은 없다).
- 본 카드 커밋이 develop에 병합되기 전까지 다른 레인이 같은 3파일을 건드리면 재측정 귀속이 무효화된다 — 병합 전 develop 흡수 후 영향 가족 재실행을 권한다.
