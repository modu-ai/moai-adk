# SPEC-MOAI-GATEWAY-001 — 코드베이스 조사

## 0. 귀속 (baseline attribution)

| 항목 | 값 |
|---|---|
| 트리 | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified` (세션 고정 경로라 디렉터리 이름은 바뀌지 않았다) |
| 브랜치 | `WT-unified-gateway` (0.1.0 작성 시점 이름은 `WT-moai-proxy-unified`) |
| HEAD | `81c1d58f9` (0.7.0 재기준. 0.6.0은 `ed71054d3`, 0.1.0~0.5.0은 `d060e0d13`) |
| divergence | `git rev-list --count --left-right origin/develop...HEAD` → `0 0` (0.7.0, 저자가 fetch 없이 측정, 로컬 `origin/develop` = `81c1d58f9`) | <!-- moving-ref-ok: origin/develop is the SUBJECT of this identity reading, not its anchor; the anchor is the pinned HEAD 81c1d58f9 in the row above, and the criterion is the measuring command with a dated 2026-09-11 reference taken by the author without fetch during the 0.7.0 revision -->
| 작업 트리 | 0.1.0 작성 시점 `git status --short` → 빈 출력 |
| 조사 방식 | read-only 렌즈 4개(launcher-surface / settings-env-auth / cg-gg-removal-scope / http-proxy-precedent) + 합성기 1개의 fan-out, 같은 트리·같은 HEAD에서 실행. 0.2.0, 0.3.0, 0.3.1, 0.3.2, 0.4.0에서 저자가 같은 트리·같은 HEAD로 추가 판독. 0.5.0에서 세션 프로브 두 건의 클라이언트 실측 요약을 §15에 더했다(정적 판독이 아니다). 0.6.0에서 HEAD `ed71054d3`로 재기준하고 프로브 조건의 한계(§15.6)와 teammate 표시 판독(§16)을 더했다. 0.7.0에서 HEAD `81c1d58f9`로 재기준하고 바뀐 두 파일의 행 인용을 옮겼다(아래 재기준 측정 0.7.0) |
| 재기준 측정 (0.6.0) | `git diff --stat d060e0d13 ed71054d3 -- internal/cli/launcher.go internal/cli/glm.go internal/cli/spawn.go internal/cli/launch_exec_posix.go internal/cli/settings.go internal/hook/session_start.go internal/hook/session_end.go internal/hook/glm_tmux.go internal/glmcred internal/paths internal/config/envkeys.go .github/workflows/ci.yml .github/workflows/release-pr-multi-os.yml` → 빈 출력. 같은 범위의 대조군 `git diff --shortstat d060e0d13 ed71054d3` → `1158 files changed, 101328 insertions(+), 1924 deletions(-)`. 따라서 위 경로의 행 인용은 두 커밋에서 같다. 위 목록 밖 경로의 행 인용은 이 측정이 덮지 않는다 |
| 재기준 측정 (0.7.0) | `git diff --stat ed71054d3 81c1d58f9 -- <위 목록> internal/cli/glm_tools.go` → `.github/workflows/ci.yml \| 2 ++`, `internal/hook/session_end.go \| 5 +++++`, `2 files changed, 7 insertions(+)`. 대조군 `git diff --shortstat ed71054d3 81c1d58f9` → `261 files changed, 39406 insertions(+), 578 deletions(-)`. 두 파일의 행 인용은 저자가 `81c1d58f9`에서 열어 옮겼다 — `session_end.go`는 옛 16~261행이 +1, 262행 이후가 +5, `ci.yml`은 옛 68~90행이 +1, 91행 이후가 +2다. 측정 시점에 기록한 명령 출력(§1.5, §6.3의 코드 블록)은 고치지 않고 블록 뒤에 옮긴 행을 적었다 |

아래 `file:line` 인용은 그 fan-out 또는 저자의 추가 판독이 이 트리·이 HEAD에서 직접 읽은 것이다.
설계 보고서와 감사 보고서에서 옮겨 온 사실은 보고서 이름을 붙여 구분한다 — 이번 실행의 실측으로
제시하지 않는다. 요구사항 본문은 심볼 이름만 쓰고, 행 번호는 이 문서에만 둔다.

### 0.1 명칭 대응

설계 원문(`reports/moai-proxy-three-provider-redesign-20260910.md`), 핸드오프
(`reports/moai-proxy-next-session-handoff-20260910.md`), iter1 감사
(`.moai/reports/SPEC-MOAI-PROXY-001/plan-audit-iter1.md`)는 이 구성 요소를 "proxy"라 부르고
`REQ-MP` / `AC-MP` 식별자를 쓴다. iter2 감사(`.moai/reports/SPEC-MOAI-GATEWAY-001/plan-audit-iter2.md`)부터는
"gateway" / `REQ-MG` / `AC-MG`를 쓴다. 이 SPEC도 후자를 쓴다. 보고서 파일명과 감사 경로는 실제 경로이므로
바꾸지 않는다. 렌즈 이름 `http-proxy-precedent`와 렌즈가 원문으로 쓴 표현("in-binary proxy")은 인용이므로
그대로 둔다.

## 1. 저자가 직접 측정한 것

### 1.1 `moai gg` 부재 — 두 시점

**SPEC 산출물 작성 전 (0.1.0).** 대조군 없는 0은 "부재"와 "도구가 아무것도 읽지 않았다"를
구분하지 못하므로 동일 범위 양성 대조군을 함께 쟀다.

```
$ /usr/bin/grep -rn "moai gg" . | wc -l
       0
$ /usr/bin/grep -rn "moai cg" . | wc -l
     745
```

이 `0`은 **작성 전** 값이다. 그 뒤 이 SPEC 문서와 iter1 감사 보고서가 그 토큰을 담게 되어 같은
명령으로는 재현되지 않는다(iter1 감사가 `9`를 관측했다).

**0.2.0 재측정.** 이 SPEC 디렉터리와 `.moai/reports/`를 제외하고, 같은 파이프라인으로 양성
대조군을 쟀다.

```
$ /usr/bin/grep -rn "moai gg" . | /usr/bin/grep -v '^\./\.moai/specs/SPEC-MOAI-GATEWAY-001/' \
    | /usr/bin/grep -v '^\./\.moai/reports/' | wc -l
       0
$ /usr/bin/grep -rn "moai cc" . | /usr/bin/grep -v '^\./\.moai/specs/SPEC-MOAI-GATEWAY-001/' \
    | /usr/bin/grep -v '^\./\.moai/reports/' | wc -l
    1230
$ /usr/bin/grep -rln "moai gg" .
./.moai/specs/SPEC-MOAI-GATEWAY-001/research.md
./.moai/specs/SPEC-MOAI-GATEWAY-001/spec.md
./.moai/specs/SPEC-MOAI-GATEWAY-001/acceptance.md
./.moai/reports/SPEC-MOAI-PROXY-001/plan-audit-iter1.md
```

제외 없이 잡히는 파일 넷이 전부 이 SPEC 자신과 감사 보고서다. `AC-MG-015`가 문자열 스윕 대신
help 출력과 command tree를 판정 계기로 삼는 이유다.

`/usr/bin/grep`을 명시한 이유: 이 셸의 `grep`은 ugrep 래퍼이며 gitignore 규칙과 바이너리 추정
필터를 조용히 적용한다. 부재 판정에서 그 필터는 결과를 왜곡한다.

### 1.2 exec 경로

```
$ sed -n '18,30p' internal/cli/launch_exec_posix.go
...
// REQ-CGH-001: syscall.Exec is POSIX-only. The Windows companion
// (launch_exec_windows.go) spawns a child and propagates its exit code instead,
// mirroring the reexecNewBinary pattern in update.go.
func execOrSpawnClaude(claudeBin string, args, env []string) error {
	return syscall.Exec(claudeBin, args, withSessionPID(env, os.Getpid()))
}
```

Windows 구현(`internal/cli/launch_exec_windows.go:31-60`)은 `child.Start()` →
`homestate.ProbeProcessIdentity` → `transferProfileLeaseToChild` → `child.Wait()`로 종료 코드를
전파한다. 종료 코드 전파 행은 §6.4.

### 1.3 SPEC ID 정규식

0.1.0 시점에 옛 식별자로 실행해 `PASS`를 관측했다. 이후 식별자 `SPEC-MOAI-GATEWAY-001`의
검사 결과는 이 문서가 아니라 각 개정 보고에 인용한다.

### 1.4 새 환경변수 이름의 미사용 확인 (0.3.0 재측정)

이 SPEC 디렉터리와 `.moai/reports/`를 **둘 다** 제외했다. iter2 감사 보고서가 이 토큰을 담게 되어,
보고서를 빼지 않은 0.2.0의 측정은 더 이상 재현되지 않는다.

```
$ /usr/bin/grep -rn 'MOAI_LAUNCH_PROVIDER\|EnvMoaiLaunchProvider' . --include='*.go' \
    --include='*.md' --include='*.yaml' | /usr/bin/grep -v '^\./\.moai/specs/SPEC-MOAI-GATEWAY-001/' \
    | /usr/bin/grep -v '^\./\.moai/reports/' | wc -l
       0
$ /usr/bin/grep -rn 'MOAI_KANBAN_BACKEND\|EnvMoaiKanbanBackend' . --include='*.go' \
    --include='*.md' --include='*.yaml' | /usr/bin/grep -v '^\./\.moai/specs/SPEC-MOAI-GATEWAY-001/' \
    | /usr/bin/grep -v '^\./\.moai/reports/' | wc -l
      20
```

같은 범위·같은 필터로 기존 키는 20건이 잡히므로, 새 이름의 0은 탐색 실패가 아니라 미사용이다.

### 1.5 GLM 키를 `settings.local.json`에 쓰는 launch 경로 (0.3.1 측정)

0.3.0은 `moai glm`이 오늘 `injectGLMEnv`로 이 파일에 쓴다고 적었다. 같은 트리에서 다시 쟀다.

```
$ /usr/bin/grep -rnw --include='*.go' --exclude='*_test.go' injectGLMEnv internal/ pkg/ cmd/
internal/cli/glm.go:981:// injectGLMEnv adds GLM environment variables to settings.local.json.
internal/cli/glm.go:987:func injectGLMEnv(settingsPath string, glmConfig *GLMConfigFromYAML) error {
$ /usr/bin/grep -rnw --include='*.go' --exclude='*_test.go' removeGLMEnv internal/ pkg/ cmd/
internal/cli/glm.go:985:// MOAI_BACKUP_AUTH_TOKEN before being overwritten. removeGLMEnv restores it.
internal/cli/glm.go:1001:		// This preserves a Claude OAuth token so that removeGLMEnv can restore it.
internal/cli/launcher.go:229:	if err := removeGLMEnv(settingsPath); err != nil {
internal/cli/launcher.go:376:// removeGLMEnv removes GLM environment variables from settings.local.json.
internal/cli/launcher.go:380:func removeGLMEnv(settingsPath string) error {
internal/cli/settings.go:173:// absent. The credential-stripping logic mirrors removeGLMEnv's env handling, but
internal/hook/session_end.go:671:// Cleanup logic mirrors removeGLMEnv() in internal/cli/cc.go:
internal/hook/session_end.go:696:	// Treat empty file as no-op (same as removeGLMEnv in cc.go).
```

0.7.0 재기준 주: 위 출력은 `ed71054d3` 이전 트리의 기록이다. `81c1d58f9`에서 `internal/hook/session_end.go`의 두 행은
`:676`, `:701`이다(§0 재기준 측정 0.7.0).

같은 명령·같은 범위에서 대조군 `removeGLMEnv`는 호출 지점(`launcher.go:229`)을 포함해 8줄이 잡힌다.
`injectGLMEnv`의 2줄은 주석(`:981`)과 정의(`:987`)뿐이다 — **프로덕션 호출자가 없다.** 쌍둥이
`injectGLMEnvForTeam`은 죽은 호출자와 함께 #1531에서 제거되었다(`glm.go:400-401` 주석).

다른 두 `moai glm` 경로도 이 파일을 쓰지 않는다.

- `applyGLMMode`(`internal/cli/launcher.go:248`)는 `setGLMEnv`로 프로세스 env만 싣는다. `:268-272` 주석이 이
  생략을 의도로 밝힌다 — `settings.local.json`에 쓰면 `moai glm`이 끝난 뒤의 `claude` 실행으로 GLM env가 샌다.
- `moai glm setup`은 `runGLMSetup`(`internal/cli/glm.go:451-474`) → `saveGLMKey`(`:955-957`) →
  `glmcred.Save`로 `~/.moai/.env.glm`만 쓴다.

  ```
  $ /usr/bin/grep -rn 'settings.local' internal/glmcred/
  (출력 없음, exit=1)
  $ /usr/bin/grep -n 'WriteFile\|env.glm\|func Save' internal/glmcred/glmcred.go   # 대조군, 발췌
  69:func Save(key string) error {
  81:	if err := os.WriteFile(envPath, []byte(content), 0o600); err != nil {
  ```

반대 방향의 파일 정리는 살아 있다 — `applyCCMode`가 `removeGLMEnv`를 부른다(`launcher.go:229`).

**결론.** 현재 프로덕션 launch 경로 중 GLM 키를 `settings.local.json`에 쓰는 것은 없다. 살아 있는 쓰기
주체는 SessionStart 훅 `ensureGLMCredentials` 하나다(§6.5). `injectGLMEnv`는 옛 동작의 기록으로 트리에 남은
죽은 코드다. 이 SPEC은 그 함수를 지우지 않고 관측으로만 남긴다.

## 2. launcher 표면

- launch cobra 그룹의 현재 구성원은 `cc`, `cg`, `glm`, `codex`이다(도움말 TUI의
  Launchers 절에는 `web`, `statusline`도 표시된다).
- `moai cc` — `internal/cli/cc.go:23` `Use: "cc [-p profile] …"`, `:115 GroupID: "launch"`,
  `:117 DisableFlagParsing: true`, `:121 rootCmd.AddCommand(ccCmd)`, 마지막 줄
  `return unifiedLaunch(profileName, "claude", filteredArgs)`.
- `moai glm` — `internal/cli/glm.go:40,125`; 하위 명령을 **두 방식으로 동시에** 붙인다.
  `glm.go:155-156 glmCmd.AddCommand(glmSetupCmd, glmStatusCmd)` /
  `glm_tools.go:198 glmCmd.AddCommand(glmToolsCmd)`,
  그리고 `glm.go:186-199`의 수동 라우터
  `switch args[0] { case "setup": … case "status": … case "tools": … }`.
  `DisableFlagParsing: true`가 cobra 자동 라우팅을 막기 때문이다.
- `moai codex` — `internal/cli/codex_launcher.go:317-370`. 닫힌 동사 집합
  `cli | status | app`을 `codexVerbRouting`으로 처리하고 `SilenceErrors`/`SilenceUsage`로
  진단 출력을 바이트 단위로 고정한다. `:359 GroupID: "launch"`. 주석이 스스로를
  "the launcher-family sibling of cc/glm/cg"라 부른다.
- 공통 funnel — `internal/cli/launcher.go:57 func unifiedLaunch(...)`,
  `:130 unifiedLaunchDefault`, `:158-171`의 mode switch
  (`"glm"` → `applyGLMMode`, `"claude_glm"` → `applyCGMode`, default → `applyCCMode`),
  `:213 appendCrossSessionSettings`, `:215 launchClaude`, `:626`, `:634`.
- **launcher 이름이 하드코딩된 곳은 최소 세 곳이다.**
  1. `internal/cli/help_order.go:31` — `"launch": {"cc", "glm", "cg"}` (`helpGroupFrequency`)
  2. `internal/cli/help.go:44-49` — `rootHelpGroups()`의 Launchers 행들
     (`{"moai cc", …}`, `{"moai cg", …}`, `{"moai glm", …}`, `{"moai codex", …}`,
     `{"moai web", …}`, `{"moai statusline", …}`)
  3. 각 launcher 실행 함수 안의 `spawnLaunch` **호출 지점 리터럴** —
     `internal/cli/cc.go:140` `spawnLaunch(cmd.OutOrStdout(), "cc", spawnArgs)`,
     `internal/cli/cg.go:93` `spawnLaunch(cmd.OutOrStdout(), "cg", spawnArgs)`,
     `internal/cli/glm.go:205` `spawnLaunch(cmd.OutOrStdout(), "glm", spawnArgs)`.

  `internal/cli/spawn.go:152`의 `spawnLaunch`는 `subcommand string` 파라미터를 받는 함수 정의일
  뿐이고 리터럴을 담지 않는다. 도움말 행은 `registeredRootSubcommands`(`help.go:76-89`)로
  게이팅되므로, 등록이 사라진 이름은 조용히 숨겨지지만 죽은 코드로 남고, **새로 추가한 `gpt`는
  명시적으로 넣지 않으면 절대 나타나지 않는다.**

## 3. provider 선택은 오늘 exec 시점 환경변수다

`internal/cli/glm.go:363-392 setGLMEnv`가 심는 것: `ANTHROPIC_AUTH_TOKEN`,
`ANTHROPIC_BASE_URL`, `ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU,FABLE}_MODEL`,
`CLAUDE_CODE_AUTO_COMPACT_WINDOW`, `CLAUDE_CODE_MAX_CONTEXT_TOKENS`,
`CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS="1"`(Z.AI가 Anthropic beta 헤더를 거부한다),
`API_TIMEOUT_MS="3000000"`, `Z_AI_API_KEY`.

`applyCCMode`(`internal/cli/launcher.go:223`)는 반대 방향으로 `settings.local.json`의 GLM 키를
`removeGLMEnv`(`:229`)로, tmux 세션 env의 GLM 키를 `clearTmuxSessionEnv`로 걷어낸다. 두 삭제 목록은 서로
같지 않다(§6.5).
`internal/config/defaults.go:136 DefaultGLMBaseURL = "https://api.z.ai/api/anthropic"`.

이것이 이 SPEC이 교체하는 "세션당 backend 하나, 바꾸려면 재실행" 모델이다.

## 4. 환경변수 이름 SSOT와 가드

- `internal/config/envkeys.go` — env 이름 SSOT, 상수 85개. L413-419 주석:
  "This block enumerates the ANTHROPIC_* namespace COMPLETELY — that completeness is
  the package's stated SSOT contract, asserted at runtime by
  TestAnthropicBannedSetCoversAllNames." L424 `EnvAnthropicBaseURL`,
  L427 `EnvAnthropicAuthToken`, L432 `EnvAnthropicAPIKey`, L459-471
  `ANTHROPIC_DEFAULT_{,HAIKU_,FABLE_,SONNET_,OPUS_}MODEL`, L537-549
  `ANTHROPIC_CUSTOM_MODEL_OPTION*`, L572 `EnvAnthropicPrefix`.
- **가드는 빌드가 아니라 시험이다.** `internal/config/anthropic_env_ssot_test.go:83`
  `TestNoBareAnthropicEnvVarLiteralsInProduction` — AST 기반(`go/ast`, `go/parser`)으로
  `anthropicScanRoots = []string{"internal", "pkg", "cmd"}`의 production Go를 훑는다. L146-153
  `TestAnthropicBannedSetCoversAllNames`. 실패하는 테스트는 `go build`를 깨지 않는다.
- `envkeys.go:479` — "points at an LLM gateway, and none of them apply on api.anthropic.com."
  이 SPEC이 이름을 "gateway"로 바꾼 근거 중 하나다.
- **`ANTHROPIC_CUSTOM_MODEL_OPTION*`와 `ANTHROPIC_DEFAULT_*_MODEL{,_NAME,_DESCRIPTION,
  _SUPPORTED_CAPABILITIES}`는 이미 선언되어 있으나 MoAI가 읽지도 쓰지도 않는다**
  ("MoAI neither reads nor writes it"). picker에 GPT·GLM 항목을 노출할 기성 표면이다. 0.5.0 프로브 2가
  `ANTHROPIC_CUSTOM_MODEL_OPTION`과 `ANTHROPIC_DEFAULT_OPUS_MODEL`로 picker 항목이 생기는 것을 관측했다(§15).
- **launch 사실 운반 선례**: `envkeys.go:224-235` `EnvMoaiKanbanBackend = "MOAI_KANBAN_BACKEND"`.
  주석 원문: "It is deliberately carried rather than inferred: ANTHROPIC_BASE_URL is set by the
  GLM path but is settable by anyone, so deriving the backend from it would be a guess dressed as
  a measurement (SPEC-KANBAN-RECORD-SESSION-KEY-001 REQ-KRS-006)." 설정 지점은
  `internal/cli/kanban.go:487-492` `exportKanbanLaunchFacts` 하나이고, kanban·factory 진입 경로에서만
  호출된다. 읽는 지점은 `internal/hook/session_start_record.go:88-91`.
- `MOAI_*` 상수 명명 규약: launcher가 운반하는 launch 사실은 `EnvMoaiKanban*`(`:182-263`),
  `EnvMoaiFactory*`(`:276-283`), `EnvMoaiSessionPID`(`:293`)처럼 `EnvMoai*` 접두를 쓴다.

## 5. GLM 활성 판정 — 네 지점, 세 기구

0.1.0은 세 지점을 모두 "부분 문자열 판정"으로 묶었다. iter1 감사가 존재 술어를 구분했고, iter2
감사가 `internal/tmux`의 두 술어가 토큰 판정을 먼저 한다는 사실을 지적했다. 0.3.0에서 저자가 네
지점을 모두 다시 열어 판정 형태를 확인했다.

### 5.1 `hookProcessEnvHasGLM` — 단일 부분 문자열 술어 (이 SPEC이 이관)

`internal/hook/session_start_glm_guardrail.go`를 전량 읽었다.

- `:27` `const glmProcessEnvSubstring = "z.ai"`
- `:39-40` —

  ```go
  func hookProcessEnvHasGLM() bool {
  	return strings.Contains(os.Getenv(config.EnvAnthropicBaseURL), glmProcessEnvSubstring)
  }
  ```

- 토큰 판정이 **없다.** 프로세스 env의 `ANTHROPIC_BASE_URL`에 `"z.ai"`가 있는지만 본다.
- 같은 파일 `:29-38` 주석은 이 검사를 "cg_detect.go `hasGLMEnv`의 base-URL disjunct와 동일한 검사"라
  부른다. 즉 이 파일 스스로가 `hasGLMEnv`를 여러 disjunct를 가진 판정으로 서술한다.
- 비테스트 호출자는 같은 파일의 `glmGuardrailReminder`(`:63`) 하나다.

gateway 아래에서 자식 env의 base URL은 loopback이므로 이 술어는 GLM 세션을 알아보지 못한다.
**코드 판독에 근거한 미래 변경 추론이며 관측된 실패가 아니다.**

### 5.2 `cleanupGLMSettingsLocal` — 존재 술어와 그 쓰기 부작용 (이 SPEC이 이관)

- `internal/hook/session_end.go:96-105` — SessionEnd는 프로젝트 디렉터리가 비어 있지 않으면
  **조건 없이** `cleanupGLMSettingsLocal(projectDir)`을 호출한다. `:96-99` 주석이 원래 목적을 적는다
  — 사용자가 `moai glm`을 쓴 뒤 `moai cc` 없이 세션을 끝낸 경우의 stale 키 정리.
- `:671-685` 함수 주석 — "ANTHROPIC_BASE_URL is used as the GLM-active indicator: Claude Code's OAuth
  flow never sets this variable, so its presence reliably signals GLM mode."
- `:687` `func cleanupGLMSettingsLocal(projectDir string)` — `projectDir/.claude/settings.local.json`을
  읽는다. **프로세스 env가 아니라 파일**이다.
- `:731-736` 판정 — `if _, glmActive := env[config.EnvAnthropicBaseURL]; !glmActive { return }`.
- `:740` — 백업이 있으면 `env[config.EnvAnthropicAuthToken] = backup`으로 복원한 뒤 파일을 다시 쓴다.

### 5.3 `sessionEnvHasGLM`·`hasGLMEnv` — 토큰이 먼저인 이중 판정 (형제 SPEC에 인계)

`internal/tmux/cg_detect.go`를 다시 열었다.

- `sessionEnvHasGLM` `:180` — tmux 세션 env를 읽어 `:185` `env[config.EnvAnthropicAuthToken] != ""`이면
  `true`를 먼저 반환하고, 그렇지 않을 때만 `:188` `strings.Contains(env[config.EnvAnthropicBaseURL], "z.ai")`를
  본다. `:177-179` 주석도 "a non-empty ANTHROPIC_AUTH_TOKEN, or an ANTHROPIC_BASE_URL containing "z.ai""라 적는다.
- `hasGLMEnv` `:203` — 프로세스 env에 대해 `:204` 토큰 판정이 먼저, `:207` z.ai 부분 문자열이 다음이다.
- **보존 의도.** `:197-201` 주석 원문: "PRESERVED from the original SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001
  detector … It is INTENTIONALLY retained — removing it would break the sibling SPEC's IsCGMode tests that
  set ANTHROPIC_AUTH_TOKEN in the process env (C-7 + AC-CGH-006 Scenario 6b)."
- **도달 가능성.** 두 함수의 비테스트 호출자는 `IsCGMode` 안의 `:99`, `:125` 둘뿐이다. `IsCGMode`
  자신에 대한 비테스트 검색 결과는 정의 `:91`과 주석 줄(`cg_detect.go:8,13,41,67,200`,
  `internal/template/glm_effort_overlay.go:307`)뿐이며 호출 지점은 없다.

  ```
  $ /usr/bin/grep -rn 'IsCGMode' internal cmd pkg --include='*.go' | /usr/bin/grep -v '_test.go'
  internal/template/glm_effort_overlay.go:307:// internal/tmux/cg_detect.go IsCGMode (which reads …
  internal/tmux/cg_detect.go:8:// REQ-CGH-006 reworks IsCGMode into a layered OR …
  internal/tmux/cg_detect.go:13:// a sufficient fallback so the sibling SPEC's IsCGMode tests stay green.
  internal/tmux/cg_detect.go:41:// and IsCGMode degrades gracefully (EC-4).
  internal/tmux/cg_detect.go:67:// IsCGMode reports whether the current session is in CG (Claude + GLM) mode.
  internal/tmux/cg_detect.go:91:func IsCGMode(settingsPath string, stderrSink io.Writer) (bool, error) {
  internal/tmux/cg_detect.go:200:// break the sibling SPEC's IsCGMode tests that set ANTHROPIC_AUTH_TOKEN in the
  ```

- iter2 감사 보고서는 추가로 두 SPEC의 상태를 읽었다 — `SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001`은
  implemented, `SPEC-V3R6-CG-MODE-HARDENING-001`은 completed이며 그 `spec.md:106`에 C-7이 있다.
  이 두 사실은 감사 보고서에서 옮긴 것이며 저자가 이번에 재측정하지 않았다.

**처리.** 두 술어는 프로덕션에서 도달하지 않고, 토큰 판정 쪽은 다른 SPEC의 시험이 고정한 동작이다.
이 SPEC은 두 술어를 이관 목록에서 빼고 `SPEC-MOAI-CG-RETIRE-001`(제안)에 넘긴다. gateway 아래에서 두
술어가 내는 값은 세션 접근 토큰 운반 키에 따라 뒤집히므로 단정하지 않는다 — 운반 키가
`ANTHROPIC_AUTH_TOKEN`이면 모든 gateway 세션에서 참이 된다.

### 5.4 관련 주석

- `internal/cli/glm.go:571-573` — `ANTHROPIC_AUTH_TOKEN`은 tmux clear에서 **의도적으로 제외**한다
  (모드 전환을 넘어 살아남아야 할 OAuth token일 수 있다). `ANTHROPIC_BASE_URL`이 "GLM activation
  indicator" 역할을 한다.

## 6. `settings.local.json`의 쓰기 주체

### 6.1 `removeGLMEnv`의 실제 키 목록 (0.3.0 판독)

`internal/cli/launcher.go:380` `func removeGLMEnv(settingsPath string) error` — 비어 있지 않은 파일에
대해 `mutateSettingsLocal`(flock + atomic write)로 다음을 수행한다.

| 행 | 동작 |
|---|---|
| `:402` | `delete(m, "teammateMode")` — 최상위 키 |
| `:406-408` | `env["MOAI_BACKUP_AUTH_TOKEN"]`이 비어 있지 않은 문자열이면 `env[config.EnvAnthropicAuthToken] = backup` 후 `MOAI_BACKUP_AUTH_TOKEN` 삭제 |
| `:410` | 백업이 없으면 `delete(env, config.EnvAnthropicAuthToken)` |
| `:412` | `delete(env, config.EnvAnthropicBaseURL)` |
| `:413-416` | `EnvAnthropicDefaultHaikuModel`, `EnvAnthropicDefaultSonnetModel`, `EnvAnthropicDefaultOpusModel`, `EnvAnthropicDefaultFableModel` 삭제 |
| `:418-420` | `EnvClaudeCodeDisableExperimentalBetas`, `"API_TIMEOUT_MS"`, `EnvClaudeCodeDisableNonessentialTraffic` 삭제 |
| `:422` | `EnvClaudeCodeTeammateDisplay` 삭제 |
| `:425` | `"MOAI_STATUSLINE_CONTEXT_SIZE"` 삭제 |
| `:428` | `EnvClaudeCodeAutoCompactWindow` 삭제 |
| `:430-431` | `env`가 비면 `env` 키 삭제 |

`CLAUDE_CODE_MAX_CONTEXT_TOKENS`는 이 목록에 **없다**. 그 키의 비테스트 쓰기·지우기 지점:

```
$ /usr/bin/grep -rn 'EnvClaudeCodeMaxContextTokens' internal cmd pkg --include='*.go' | /usr/bin/grep -v '_test.go'
internal/config/envkeys.go:379 / :386   (정의)
internal/cli/glm.go:380                 os.Setenv (프로세스 env)
internal/cli/glm.go:536                 tmux 주입 vars
internal/cli/glm.go:620                 buildTmuxClearVars 목록 (tmux 쪽 정리)
internal/cli/glm.go:1027 / :1029        injectGLMEnv가 settings env에 쓰거나 지움 (프로덕션 호출자 없음, §1.5)
internal/hook/session_start.go:859,862  비교용 읽기
internal/hook/session_start.go:966,970  maybeDeclareGLMContextWindow가 settings env에 씀
```

(행 뒤 설명은 저자 요약이다. 원출력은 각 행의 코드 원문이다.)

### 6.2 `ensureTeammateMode` — `teammateMode`의 정당한 재작성

- `internal/hook/session_start.go:631` — SessionStart 체인이 매 세션 `ensureTeammateMode`를 호출한다.
- `:995` `func ensureTeammateMode(projectDir string) string` — `TMUX` 유무로 원하는 값(`tmux`/`auto`)을
  정하고, 현재 값과 다르면 `:1060`에서 `writeSettingsSecure`로 파일을 다시 쓴다. 파일이 없어도 만든다.
  다시 쓸 때 `env`의 legacy `CLAUDE_CODE_TEAMMATE_DISPLAY`도 지운다(`:1033-1047`). 값이 이미 맞으면
  `:1026-1027`에서 돌아가므로 이 삭제도 일어나지 않는다.
- 이 함수와 `removeGLMEnv`(`launcher.go:402`)가 모두 `teammateMode`를 만진다. 둘 다 GLM 라우팅과
  무관하다. 이것이 `AC-MG-018`이 파일 전체 해시 대신 GLM 정리 키 집합 투영을 쓰고 `teammateMode`를
  명시 제외하는 근거다.

### 6.3 GLM 라우팅 키를 settings에 쓰는 비테스트 경로 전수

`settings.local.json`의 env 맵에 GLM 라우팅 키를 대입하는 비테스트 코드를 검색했다.

```
$ /usr/bin/grep -rn '\[config\.EnvAnthropicBaseURL\] = \|\[config\.EnvAnthropicAuthToken\] = \|\["MOAI_BACKUP_AUTH_TOKEN"\] = \|\[config\.EnvAnthropicDefault[A-Za-z]*Model\] = ' internal cmd pkg --include='*.go' | /usr/bin/grep -v '_test.go'
internal/cli/glm.go:1003:			env["MOAI_BACKUP_AUTH_TOKEN"] = existing
internal/cli/glm.go:1007:		env[config.EnvAnthropicAuthToken] = apiKey
internal/cli/glm.go:1008:		env[config.EnvAnthropicBaseURL] = glmConfig.BaseURL
internal/cli/glm.go:1009:		env[config.EnvAnthropicDefaultOpusModel] = glmConfig.Models.High
internal/cli/glm.go:1010:		env[config.EnvAnthropicDefaultSonnetModel] = glmConfig.Models.Medium
internal/cli/glm.go:1011:		env[config.EnvAnthropicDefaultHaikuModel] = glmConfig.Models.Low
internal/cli/glm.go:1012:		env[config.EnvAnthropicDefaultFableModel] = glmConfig.Models.Fable
internal/cli/launcher.go:407:				env[config.EnvAnthropicAuthToken] = backup
internal/cli/settings.go:183:			env[config.EnvAnthropicAuthToken] = backup
internal/hook/session_start.go:882:	settings.Env[config.EnvAnthropicAuthToken] = apiKey
internal/hook/session_start.go:884:		settings.Env[config.EnvAnthropicBaseURL] = config.DefaultGLMBaseURL
internal/hook/session_end.go:735:		env[config.EnvAnthropicAuthToken] = backup
internal/hook/glm_tmux.go:53:		result[config.EnvAnthropicAuthToken] = token

$ /usr/bin/grep -rn '\[config\.EnvClaudeCodeDisableExperimentalBetas\] = \|\["API_TIMEOUT_MS"\] = \|\[config\.EnvClaudeCodeTeammateDisplay\] = \|\["MOAI_STATUSLINE_CONTEXT_SIZE"\] = \|\[config\.EnvClaudeCodeAutoCompactWindow\] = \|\[config\.EnvClaudeCodeDisableNonessentialTraffic\] = \|\["CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS"\] = ' internal cmd pkg --include='*.go' | /usr/bin/grep -v '_test.go'
internal/cli/glm.go:529:		vars[config.EnvClaudeCodeAutoCompactWindow] = window
internal/cli/glm.go:1014:		env[config.EnvClaudeCodeDisableExperimentalBetas] = "1"
internal/cli/glm.go:1015:		env["API_TIMEOUT_MS"] = "3000000"
internal/cli/glm.go:1019:			env[config.EnvClaudeCodeAutoCompactWindow] = window
internal/hook/session_start.go:888:		settings.Env["CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS"] = "1"
internal/hook/session_start.go:953:		env[config.EnvClaudeCodeAutoCompactWindow] = strconv.Itoa(config.Default1MContextTokens)
```

0.7.0 재기준 주: 위 출력의 `internal/hook/session_end.go:735`는 `81c1d58f9`에서 `:740`이다(§0 재기준 측정 0.7.0).

분류(저자 판독):

- `glm.go:1003-1019` — `injectGLMEnv` 본문. 이 함수에는 프로덕션 호출자가 없다(§1.5). 현재 어떤 launch
  경로도 이 쓰기를 실행하지 않으며, 옛 동작의 기록으로만 남아 있다. gateway launch도 되살려 호출하지
  않는다(`design.md` §6.2).
- `glm.go:529` — tmux 주입용 `vars` 맵. 파일 쓰기가 아니다.
- `launcher.go:407`, `settings.go:183`, `session_end.go:740` — 백업 토큰 복원. 각각 `removeGLMEnv`, CG 리더
  정리, SessionEnd 정리다.
- `session_start.go:882,884,888` — `ensureGLMCredentials`(`:809`) 본문.
- `session_start.go:953` — `maybeSet1MAutoCompactWindow`(`:948`). 호출 지점은 `:860`, `:894`로 둘 다
  `ensureGLMCredentials` 안이다. `maybeDeclareGLMContextWindow`(`:965`)의 호출 지점 `:861`, `:898`도 같다.
- `glm_tmux.go:53` — `buildGLMTmuxEnvVars`가 tmux 주입용으로 만드는 **메모리 맵** `result`. 파일 쓰기가
  아니다(`glm_tmux.go:22-58` 판독).

SessionStart 쓰기 체인 `runSettingsChain`(`session_start.go:617-653`)의 네 주체:
`ensureGLMCredentials`, `ensureTeammateMode`(`teammateMode`를 쓰며, 다시 쓸 때 `env`의 legacy `CLAUDE_CODE_TEAMMATE_DISPLAY`를 지운다 — `:1033-1047`. 값이 이미 맞으면 `:1026-1027`에서 쓰기 전에 돌아간다), `ensureTmuxGLMEnv`(tmux 세션 env),
Windows 전용 `injectCLAUDEEnvFile`(`:646-651`에서 `claudeEnvFileGuard`로 게이트, `:1285` 정의, `:1333`
쓰기 — `env["CLAUDE_ENV_FILE"]`만 씀).

**결론.** 프로덕션 코드에서 GLM 정리 키 집합을 settings에 쓰는 경로는 launch 구간과 세션 구간을 통틀어
`ensureGLMCredentials`(와 그 안의 보조 함수 둘)뿐이다(§1.5, §6.5). 이 결론은 비테스트 Go에 대한 대입 패턴 검색이며, 다른 형태의 쓰기(예: 구조체 필드 대입,
문자열 조립)와 삭제(`delete`)는 이 패턴에 잡히지 않는다 — 그래서 0.3.1 판독은 `ensureTeammateMode`의 legacy
키 삭제를 놓쳤고 0.4.0에서 바로잡았다(iter3 G3-A1). 런타임 관측은 없다.

### 6.4 종료 코드 전파 — 두 플랫폼

- POSIX: `internal/cli/launch_exec_posix.go:26` — `syscall.Exec`가 프로세스를 치환하므로 Claude의 종료
  코드가 곧 launcher의 종료 코드다.
- Windows: `internal/cli/launch_exec_windows.go:57` `if err := child.Wait(); err != nil {`, `:60`
  `os.Exit(ee.ExitCode())`, `:64` 정상 종료 시 `os.Exit(0)`.

`REQ-MG-005`의 플랫폼 무관한 종료 코드 전파 의무는 이 두 경로에 근거한다.

### 6.5 `ensureGLMCredentials`의 쓰기와 네 GLM 삭제 목록 (0.3.1 판독)

**유일하게 살아 있는 쓰기 주체 `ensureGLMCredentials`** (`internal/hook/session_start.go:809`)

- 발동 조건: 파일 `env`의 `ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU}_MODEL` 값 중 하나에 `"glm"`이 들어 있다
  (`:833-848`). FABLE 슬롯은 보지 않는다. 훅 로컬 `isCGMode(projectDir)`가 참이면 건너뛴다(`:829`).
- 토큰 있음 분기(`:851`): `maybeSet1MAutoCompactWindow`(`:860`)와 `maybeDeclareGLMContextWindow`(`:861`)로
  `CLAUDE_CODE_AUTO_COMPACT_WINDOW`·`CLAUDE_CODE_MAX_CONTEXT_TOKENS`를 채우고, 두 값 중 어느 하나라도 바뀌면
  (연결 문자열 비교 `after != before`, `:859`, `:862-863`)
  `persistSettingsEnv`로 파일을 다시 쓴다(`:864`). 두 보조 함수는 값이 이미 있으면 건드리지 않는다.
- 토큰 없음 분기: `~/.moai/.env.glm`에서 읽은 키를 `ANTHROPIC_AUTH_TOKEN`에 넣고(`:882`), 비어 있으면
  `ANTHROPIC_BASE_URL = DefaultGLMBaseURL`(`:884`), 비어 있으면
  `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS="1"`(`:888`), 두 context 창 키(`:894`, `:898`)를 넣은 뒤 파일을
  다시 쓴다(`:900`).
- 키 읽기 `loadGLMKeyFromEnvFile`(`:1344`)는 `glmcred.Load`를 거치지 않고 `paths.GlmEnvFile()`을 직접 연다.
  그래서 `glmcred`의 `MOAI_TEST_GLM_KEY` 시험 seam은 이 분기에 닿지 않고, 경로는 `MOAI_HOME`으로만 바꿀 수
  있다(`GlmEnvFile`, `internal/paths/paths.go:95`). 단 `MoaiHome()`은 비어 있지 않으면서 `filepath.IsAbs`가
  참인 값만 따르고, 상대 경로는 말없이 무시한 채 실제 홈 아래 `.moai`로 돌아간다(`internal/paths/paths.go:69`).
  관측으로만 남긴다.

**귀결.** 현재 프로덕션 경로 중 GLM 모델 슬롯을 이 파일에 넣는 것은 없다(§1.5). 따라서 이 훅이 발동하는
파일은 옛 바이너리가 썼거나 사람이 고친 **남은 상태**다. 업그레이드 사용자가 정확히 그 대상이므로 launch
단계 정리는 여전히 필요하다.

**관측 — 주석 부정확(고치지 않음).** 토큰 있음 분기의 주석(`:853-854`)은 창 키가 없는 설정의 출처로
"settings written by an older binary (or by `moai glm setup`)"를 든다. 현재 코드에서 `moai glm setup`은
`settings.local.json`을 쓰지 않는다(§1.5).

**네 GLM 삭제 목록 — 같은 키 집합이 아니다.** ✓ = 지운다.

| 키 | `removeGLMEnv` `launcher.go:380-435` | `stripGLMCredsAndSetTeammateMode` `settings.go:179-207` | `buildTmuxClearVars` `glm.go:595-622` | `cleanupGLMSettingsLocal` `session_end.go:738-749` |
|---|---|---|---|---|
| `ANTHROPIC_BASE_URL`, `ANTHROPIC_DEFAULT_{HAIKU,SONNET,OPUS}_MODEL` | ✓ | ✓ | ✓ | ✓ |
| `ANTHROPIC_AUTH_TOKEN` (백업 복원) | ✓ | ✓ | ✗ (의도적, `glm.go:571-573`) | ✓ |
| `ANTHROPIC_DEFAULT_FABLE_MODEL` | ✓ | ✗ | ✓ | ✗ |
| `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS`, `API_TIMEOUT_MS`, `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC` | ✓ | ✓ | ✓ | ✗ |
| `MOAI_STATUSLINE_CONTEXT_SIZE` | ✓ | ✓ | ✓ | ✗ |
| `CLAUDE_CODE_AUTO_COMPACT_WINDOW` | ✓ | ✗ | ✓ | ✗ |
| `CLAUDE_CODE_TEAMMATE_DISPLAY` | ✓ | ✓ | ✗ | ✗ |
| `CLAUDE_CODE_MAX_CONTEXT_TOKENS`, `ANTHROPIC_REASONING_EFFORT`, `CLAUDE_CONFIG_DIR`, `DISABLE_PROMPT_CACHING` | ✗ | ✗ | ✓ | ✗ |

- `MOAI_STATUSLINE_CONTEXT_SIZE`는 `removeGLMEnv`에서는 문자열 리터럴로, 나머지 둘에서는 상수
  `config.EnvStatuslineContextSize`로 적혀 있지만 같은 키다(`internal/config/envkeys.go:64`).
- `buildTmuxClearVars`는 `settings.local.json`이 아니라 **tmux 세션 env**를 지운다. 다른 표면이다.
- `cleanupGLMSettingsLocal`은 `MOAI_BACKUP_AUTH_TOKEN`을 조건 없이 지운다(`:740`). 표의
  `ANTHROPIC_AUTH_TOKEN` 행은 복원 동작을 뜻한다.
- 이 표는 차이를 측정한 대로 기록할 뿐 어느 목록도 정답으로 삼지 않는다. cg 경로 사본인
  `stripGLMCredsAndSetTeammateMode`는 형제 `SPEC-MOAI-CG-RETIRE-001`(제안) 소관이다. 이 SPEC의 정리 집합은
  `design.md` §6.1이 쓰기 주체에서 따로 유도한다.

**관측 — 동등성 간극(이 SPEC이 만든 것이 아님).** `CLAUDE_CODE_MAX_CONTEXT_TOKENS`는 훅이 파일에 쓰고(옛
`injectGLMEnv`도 썼다) 세 파일 정리 함수 어느 것도 지우지 않는다. 훅이 이 키를 쓰는 것은 파일에 GLM 모델
슬롯이 남아 있을 때뿐이므로, 이 간극은 그런 파일을 거친 경우에만 드러난다. 코드 판독에 근거한 추론이며
런타임에서 재현하지 않았다.

**기존 teardown 경로 요약.**

- `removeGLMEnv` — §6.1.
- `stripGLMCredsAndSetTeammateMode` — `internal/cli/settings.go:179-207` (CG 리더 정리, `:183` 백업 복원)
- `buildTmuxClearVars` — `internal/cli/glm.go:595-622` (tmux 쪽)
- `cleanupGLMSettingsLocal` — `internal/hook/session_end.go:687` (SessionEnd 호출 `:105`)
- SessionStart 재주입 `ensureGLMCredentials` — 위 판독.

settings 쓰기는 `internal/cli/settings.go:66-101 mutateSettingsLocal`(flock +
`writeFileAtomic`, mode `0o600`)를 지난다. 훅 쪽은 `writeSettingsSecure`를 쓴다.

`ANTHROPIC_AUTH_TOKEN`이 tmux에 닿는 경로는 `mgr.InjectSensitiveEnv` 하나뿐이며
(argv-safe, CWE-214), `internal/cli/glm.go:546-568`의 주석은 "On sensitive-injection
failure we MUST NOT fall back to argv (would re-leak the token)"라고 못 박는다.

### 6.6 tmux 세션 env 표면 (0.4.0 판독)

**0.6.0.** 이 절은 0.4.0 판독 기록으로 남긴다. 이 판독에 기댄 tmux 세션 env 계약은 형제 `SPEC-MOAI-GATEWAY-TEAMMATE-001`
(제안)로 옮겼고, 코어의 tmux·teammate 계약은 `design.md` §6.6과 이 문서 §16이다.

저자가 이 트리·이 HEAD에서 직접 열어 확인한 행이다. 런타임 관측은 없다.

- `internal/cli/launcher.go:278-284` — `applyGLMMode`가 tmux 세션 안이면 `injectTmuxSessionEnv(glmConfig,
  apiKey)`를 부른다. 위 주석(`:272`)은 "Tmux team panes still receive env via injectTmuxSessionEnv below (moai
  cg path)", 실패 경고(`:280-282`)는 "Teammates spawned in new tmux panes may not have GLM credentials"라고
  적는다. `applyCCMode`는 `:224`에서 CLI `clearTmuxSessionEnv`를 부른다.
- `internal/cli/glm.go:507-540` `buildTmuxInjectVars` — `ANTHROPIC_AUTH_TOKEN`(GLM 키), `ANTHROPIC_BASE_URL`,
  네 `ANTHROPIC_DEFAULT_*_MODEL`, `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS`, `API_TIMEOUT_MS`를 넣고, High 슬롯
  모델에 따라 `MOAI_STATUSLINE_CONTEXT_SIZE`·`CLAUDE_CODE_AUTO_COMPACT_WINDOW`·`CLAUDE_CODE_MAX_CONTEXT_TOKENS`를
  조건부로 더한다.
- `glm.go:546-567` `injectTmuxSessionEnvVia(mgr tmux.SessionManager, …)` — 토큰을 `mgr.InjectSensitiveEnv`로
  보내고 맵에서 지운 뒤 나머지를 `mgr.InjectEnv`로 보낸다. 민감 값 주입이 실패하면 argv로 되돌아가지 않고
  오류를 돌려준다.
- `internal/tmux/session.go:50` `SessionManager` 인터페이스 — `InjectEnv`(`:55`), `ClearEnv`(`:58`),
  `InjectSensitiveEnv`(`:69`).
- `glm.go:574-585` CLI `clearTmuxSessionEnv` — 목록 `buildTmuxClearVars`(`:595-622`, 14키)를 `mgr.ClearEnv`로
  지운다. `ANTHROPIC_AUTH_TOKEN`은 목록에 없다(주석 `:569-573`, `:592-594`).
- `internal/hook/session_end.go:639-645` `glmEnvVarsToClean` — `ANTHROPIC_AUTH_TOKEN`, `ANTHROPIC_BASE_URL`,
  OPUS·SONNET·HAIKU 슬롯. 주석(`:631-638`)은 tmux의 토큰을 지우는 것이 "always safe"라고 적는다. 훅
  `clearTmuxSessionEnv`(`:651-670`)는 `TMUX`가 없으면 돌아가고, 있으면 키마다 `tmux set-environment -u`를
  직접 실행한다(`:658`). SessionEnd 핸들러(`:66`)가 `:94`에서 부르며, 그 앞의 조기 반환은 홈 디렉터리를 알
  수 없는 경우뿐이다.
- `internal/hook/glm_tmux.go:78` `ensureTmuxGLMEnv` — `TMUX`가 없거나(`:80`), `teammateMode`가 `tmux`가
  아니거나(`:101`), `buildGLMTmuxEnvVars`가 빈 결과를 내면(`:116-119`, 파일 `env`의 `ANTHROPIC_AUTH_TOKEN`이
  비었을 때) 돌아간다. 그렇지 않으면 함수 안에서 만든 세션 관리자로 토큰은 `InjectSensitiveEnv`(`:132`),
  나머지는 `InjectEnv`(`:143`)로 주입한다. `buildGLMTmuxEnvVars`(`:39-58`)는 토큰 값이 GLM 키인지 보지 않는다.
  호출은 `session_start.go:638`이며, 같은 체인에서 `ensureTeammateMode`(`:631`)가 먼저 돈다.
- 판독 결론 — 두 정리 함수는 `ANTHROPIC_AUTH_TOKEN`에 대해 반대로 행동한다. 이 SPEC의 선택은 `design.md`
  §6.6에 적었다.
- 판독 결론 — gateway launch 정리가 백업 토큰을 복원한 파일에서는 `ensureTmuxGLMEnv`의 발동 조건이 모두 참이
  될 수 있다(`design.md` §5.4). 추론이며 실행하지 않았다.

## 7. kanban backend 상수와 소비자

### 7.1 kanban backend 상수

- `internal/kanban/record.go:20-25` —

  ```go
  // Session backends a kanban chain can run on. A mixed-backend session is not a
  // kanban session, so no third value exists.
  const (
  	BackendClaude = "claude"
  	BackendGLM    = "glm"
  )
  ```

- `internal/kanban/record.go:77` — `Record` 필드 주석 `// Backend is BackendClaude or BackendGLM.`
- `internal/config/envkeys.go:225` — `// on: kanban.BackendClaude or kanban.BackendGLM.`

### 7.2 이름이 같은 다른 개념

- `internal/cli/mcp_convergence.go:57-64` — "Backend name constants — the three audit backends."
  `BackendClaude = "claude"`, `BackendCodex = "codex"`, `BackendGLM = "glm"`. cross-model 감사
  수렴 엔진의 backend 집합이며 kanban backend와 무관하다.
- `internal/cli/model.go:88-96` — `resolveModelProfileReport`가 LLM 설정에서 `rpt.Backend`를
  `"claude"` 또는 `"glm"`으로 정한다. 모델 프로필 보고서 필드이며 역시 별개다.

### 7.3 kanban backend 값의 소비자 (non-test)

`moai cc`의 진입 경로(`internal/cli/cc.go:158-232` 판독)는 네 분기다.

| 분기 | 호출 |
|---|---|
| factory lead | `:172` `kanban.RecordFactoryRunStart(…, kanban.BackendClaude, entry.Spec)`, `:173` `exportKanbanLaunchFacts(entry.Spec, kanban.BackendClaude)` |
| factory worker | `:192` `exportKanbanLaunchFacts(entry.Spec, kanban.BackendClaude)` |
| kanban lead | `:209` `exportKanbanLaunchFacts(entry.Spec, kanban.BackendClaude)` |
| kanban companion | `:227` `exportKanbanLaunchFacts(entry.Spec, kanban.BackendClaude)` |

`moai glm`은 같은 구조로 `internal/cli/glm.go:228`(factory lead의 `RecordFactoryRunStart`), `:229`, `:247`,
`:260`, `:276`에서 `kanban.BackendGLM`을 넘긴다.

| 소비자 | 위치 | 동작 |
|---|---|---|
| export 구현 | `internal/cli/kanban.go:487-492` | 받은 값을 `MOAI_KANBAN_BACKEND`로 `os.Setenv` |
| SessionStart 기록 | `internal/hook/session_start_record.go:88-91` | `kanban.NewRecord(…, os.Getenv(config.EnvMoaiKanbanBackend))` |
| 웹 콘솔 factory lane | `internal/web/factory_lanes.go:150` | `row.Backend = rec.Backend` |
| 웹 콘솔 운영 뷰 | `internal/web/viewmodel_ops.go:316` | `Backend: rec.Backend` |
| 홈 상태 저장 | `internal/homestate/runtime.go:26` | `row.Backend`를 그대로 기록 |
| 웹 콘솔 배지 | `internal/web/widgets.templ:69-82` | 아래 참고 |

`backendBadge`의 분기:

```
templ backendBadge(backend string) {
	if backend == "" {
		@missing()
	} else {
		<span class={ "backend", templ.KV("backend--metered", backend == "glm") }>
			<b>{ backend }</b>
			if backend == "glm" {
				<span data-i18n="backend.metered">metered</span>
			} else {
				<span data-i18n="backend.flat">flat rate</span>
			}
		</span>
	}
}
```

`glm`이 아닌 모든 비어 있지 않은 값은 "flat rate"로 표시된다. `gpt` 값이 추가되면 손대지 않는 한
정액제로 표시된다.

### 7.4 프로세스 신원 지문 선례

`internal/homestate/profile_lease.go:195` `func CurrentProcessFingerprint() string`,
`:202` `func ProbeProcessIdentity(pid int) (string, ProcessIdentityState)`. 파일 첫 줄이
`package homestate`로 빌드 태그가 없어 두 플랫폼에서 모두 컴파일된다. Windows launch 경로가 자식 신원
확인에 이미 쓴다(§1.2). `design.md` §3.4의 PID 재사용 방어 근거다.

## 8. gateway 선례 — loopback 서버는 있고, SSE 중계와 OpenAI 클라이언트는 없다

- **`internal/gateway`는 없다.** 0.1.0 렌즈 측정 `find . -type d -name "*proxy*"` → 결과 없음.
  `httputil.ReverseProxy` / `NewSingleHostReverseProxy` 비-테스트 사용 0.
  남은 "proxy" 히트는 `internal/harness/router/router.go:215`의 `ConfigProxy`,
  `internal/graph/architecture_report.go:18`의 "DIRECTORY PROXY",
  `internal/sandbox/context.go:42`의 `"proxy.golang.org"`뿐이다. `internal/harness/router/router.go:4`는
  `package router`다 — 이 SPEC이 패키지 이름으로 `router`를 쓰지 않는 이유 중 하나다.
- **재사용 가능한 가장 가까운 선례는 `internal/web`이다.**
  `server.go:44 const loopbackHost = "127.0.0.1"` — 주석: "the only interface the
  Console ever binds to (REQ-WC-002). Binding to 0.0.0.0 or any non-loopback address
  is a forbidden anti-pattern." `:47 shutdownDrain = 5 * time.Second`,
  `:56 Port int`("0 selects a random free port"), `:155 Handler()`("Exposed for
  httptest-based integration tests."), `:178-190 bind()`.
- **CSRF 계층은 재사용 불가**: `internal/web/app.go:205-250`의 게이트는
  `POST/PUT/PATCH`에 대해 `Sec-Fetch-Site != "same-origin"`이면 403인데, Claude Code는
  그 헤더를 보내지 않는다.
- 포트 회수 선례: `internal/cli/web_port.go:42-105` — `isPortInUse`,
  `portPollAttempts = 30`, `portPollInterval = 100 * time.Millisecond`, `ensurePortFree`
  ("port %d is held by a non-moai process (PID %d); …").
- **SSE 중계 선례 없음**: `internal/web/events.go`의 Hub는 `http.Flusher`를 쓰지만
  명시 계약이 "계약: 이벤트 본문에 데이터를 싣지 않는다"이다. upstream 토큰 스트림
  중계는 선례 0의 신규 코드다.
- provider HTTP 클라이언트: `internal/cli/mcp_glm.go:64-90,304-314` —
  `glmMessagesPath = "/v1/messages"`, `glmAnthropicVersion = "2023-06-01"`,
  `type glmHTTPDoer interface { Do(req *http.Request) (*http.Response, error) }`,
  헤더 `Content-Type` / `x-api-key` / `anthropic-version`.
  `internal/cli/glm_task.go:85` — `var glmTaskHTTPClient glmHTTPDoer = &http.Client{}`
  (주석 `:22`: "with NO client Timeout").
  `internal/statusline/usage.go:22-26,463-491` — `anthropicMessagesURL`,
  `Authorization: Bearer <oauth>`, `anthropic-version`; `messagesURL`/`oauthURL`는
  시험용 교체 가능 필드(`:70-73`).
  **OpenAI/GPT HTTP 클라이언트는 저장소 어디에도 없다.**

## 9. 인증 — Codex 불변식은 이미 성립, PKCE는 전무

- `internal/cli/mcp_codex.go:1781-1927` — `codexAuthFileName = "auth.json"`,
  `codexHomeEnvVar = "CODEX_HOME"`, `codexHomeDirName = ".codex"`,
  `errCodexAuthUnparseable`("REQ-CL-008 forbids a credential being retained, logged,
  or wrapped"), `nonEmptyString`("the value itself dies on the stack"),
  `codexTokenSet{credentialCount}`, `readCodexAuthFile`(읽기 전용).
  **`~/.codex/auth.json`을 쓰는 코드 경로는 없다.**
- **PKCE / OAuth authorization-code 클라이언트는 production Go에 없다** —
  `PKCE|pkce|code_verifier` 검색 히트 0. Anthropic OAuth **읽기**만 statusline에 있다
  (`readOAuthToken`: macOS 키체인 "Claude Code-credentials" → `~/.claude/.credentials.json`
  → `~/.claude/credentials.json`).
- credential 저장 선례는 `internal/glmcred` — 패키지 주석이 스스로를
  "GLM credential SSOT — exactly one writer implementation"이라 선언한다.
  `~/.moai/.env.glm`, mode 0600 + 명시 `Chmod`, dotenv `GLM_API_KEY="…"`,
  `$`/`"`/`\` 이스케이프, `MOAI_TEST_GLM_KEY` 시험 seam, `MOAI_HOME` 우회 — 비어 있지 않은
  절대 경로 값에만 적용되고 상대 경로 값은 무시된다(`internal/paths/paths.go:69`의 `filepath.IsAbs` 조건;
  `:95 GlmEnvFile()`), `internal/defs/dirs.go:400-401
  GlmEnvFileName = ".env.glm"`.
  **`~/.moai/.env.gpt`, `paths.GptEnvFile`, MoAI 키체인 writer는 모두 없다.**
- 비밀 마스킹: `internal/sandbox/env.go:31-37 defaultDenyList = {GITHUB_TOKEN,
  ANTHROPIC_API_KEY, OPENAI_API_KEY, NPM_TOKEN, GH_TOKEN}` (`internal/feedback`과 공유).
  `ANTHROPIC_AUTH_TOKEN`은 이 목록에 **없다**.

## 10. Windows 검증의 판정 지점 (0.3.1 판독 — 운영자 결정: release PR 게이트)

**`.github/workflows/release-pr-multi-os.yml`** — 저자가 직접 읽었다.

- 트리거: `main` 대상 `pull_request`(`:14`)와 `workflow_dispatch`. `detect-release` 잡이
  `startsWith(github.head_ref, 'release/')`로 head가 `release/*`인 PR만 통과시킨다(`:35`).
- 매트릭스: `os: [ubuntu-latest, macos-latest, windows-latest]`(`:91`). `:93` 주석은 "Every OS leg blocks
  the release gate"라 적는다.
- 실행: `go test -json -race -timeout 25m ./... > test-stream.json || rc=$?`(`:203`). `-tags=integration`이
  없으므로 `integration` 태그를 단 시험은 이 레그에서 실행되지 않는다.
- 업로드: `Upload test event stream` 단계(`:213`)가 `test-stream.json.gz`를
  `test-stream-release-verify-${{ matrix.os }}`(`:217`) 이름으로 올린다. `if-no-files-found: warn`(`:220`)
  — 파일이 없어도 레그는 실패하지 않는다.
- 관측(고치지 않음): 워크플로 상단 주석(`:23`)은 "no artifact upload"라 적지만 업로드 단계가 있다.
  판정과 무관한 기존 주석 부정확이다.

**`.github/workflows/ci.yml`** — 카드·develop CI.

- `:110-112` 주석: cross-platform 런타임 커버리지(macOS + Windows race)를 release 시점의
  `release-pr-multi-os.yml`로 옮겼다. 저장소의 의도된 정책이다.
- 주 `test` 잡 매트릭스는 `[ubuntu-latest]`(`:125`)다.
- 3-OS 매트릭스(`:378`)를 가진 `test-integration`은
  `go test -json -tags=integration -race -timeout 180s ./test/integration/harness/...`만 실행한다(`:402`).

결론:

- supervisor Windows 시험을 소유 패키지의 태그 없는 평범한 Go 시험으로 두면, release PR의 windows-latest
  레그가 그 시험을 실행한다. 판정은 그 레그의 이벤트 스트림에서 이름을 정한 시험이 `"Action":"pass"`로
  끝났는지로 읽는다. 아티팩트 부재는 PASS가 아니다.
- 카드·develop CI는 그 시험을 Windows에서 실행하지 않는다. Windows 회귀는 release PR 시점에야 드러난다.
- `test-install.yml`의 windows 레그(`:100,146,196`)는 이 판정 지점이 아니며, 무엇을 실행하는지 읽지
  않았다(§14).

## 11. `moai cg` 철거 규모 (형제 SPEC 소관, 여기서는 근거만 기록)

비-테스트 Go 표면 약 12개 지점: `internal/cli/cg.go`(111줄) 전체,
`launcher.go:163 case "claude_glm"`, `:290 applyCGMode`, `:365 persistTeamMode(root, "cg")`,
`settings.go`의 CG 리더 정리, `factory.go:385-392 rejectFactoryOnCG`(sentinel
`FACTORY_MODE_UNSUPPORTED_BACKEND`), `kanban.go:685-701 rejectKanbanOnCG`(메시지에
`"use 'moai cc --kanban' or 'moai glm --kanban' instead"`가 하드코딩),
`help_order.go:31`, `help.go:45`, `root.go:27`·`:55`,
`config/team_mode.go:14-16 TeamModeCG`, `config/closed_sets.go:42 {"solo","team","cg"}`,
`internal/tmux/cg_detect.go`(211줄), `internal/hook/session_start.go:976-985`의 별도
`isCGMode`(llm.yaml의 `"team_mode: cg"` 문자열 매칭으로 auth 자동 주입을 차단하는
**실제 행동 소비자**), `template/glm_effort_overlay.go:312`,
`statusline/memory.go`, `web/schemaform.go`, `web/agentfm.go`,
`settings/schema_sections.go`, `config/toolpolicy/types.go:58-65`(`ExceptionWhen: "cg_leader"`).

테스트 18개 파일(그중 `cg_test.go` 244줄, `cg_mode_hardening_test.go` 363줄,
`cg_detect_ssot_test.go` 187줄이 cg 전용).

문서·template이 비용의 대부분이다: `docs-site/content` 114개 파일 303건(4 locale,
같은 PR i18n 의무 적용), 4개 README 각 6건, `internal/template/templates` 73건,
live `.claude/` 미러 26개 파일. 이 수치는 0.1.0 렌즈 측정이며 이후 재측정하지 않았다.
`internal/template/templates/.claude/rules/moai/core/glm-web-tooling.md`는 CG Mode의
선언된 SSOT이지만 `moai glm`도 함께 지배하므로 **삭제가 아니라 재작성** 대상이다.

`team_mode`는 **영속 필드**다(`llm.yaml:3`에 `team_mode: ""`가 배포된다). 마이그레이터는
현재 없다.

0.3.0부터 형제 SPEC은 `internal/tmux`의 `sessionEnvHasGLM`·`hasGLMEnv` 처리도 맡는다(§5.3).

### 11.1 끊기는 교차 SPEC 계약 7건과 상태

상태는 0.2.0에서 각 `spec.md`의 `status:` 필드를 직접 읽었다.

| SPEC | 상태 | 끊기는 지점 |
|---|---|---|
| `SPEC-FACTORY-MODE-001` | completed | AC-FM-004가 `moai cg --factory` → `FACTORY_MODE_UNSUPPORTED_BACKEND`를 단언 |
| `SPEC-FACTORY-WORKER-FANOUT-001` | implemented | `spec.md:117`의 cg 참조 |
| `SPEC-KANBAN-BOOTSTRAP-001` | draft | `plan.md:143`, `research.md:328`의 cg 참조 |
| `SPEC-STEERING-ALIGN-GUARDRAIL-HOOK-001` | completed | REQ-GH-005가 cg-leader pane을 명시 예외로 둔다 |
| `SPEC-V3R6-TOOL-POLICY-SSOT-001` | completed | `exception_when: "cg_leader"` |
| `SPEC-MODEL-ROUTING-WIRE-001` | **superseded** (`superseded_by: SPEC-AGENT-ARCH-V2-001`) | `spec.md:148`의 cg 제외. **이미 대체되었으므로 형제 SPEC의 비용 추정에서 빼야 한다.** 또한 agent 모델 라우팅이지 provider 요청 라우팅이 아니어서 이 SPEC과 도메인이 겹치지 않는다 |
| `SPEC-INFINITE-GOAL-001` | completed | `spec.md:77`이 cc/cg launcher를 block-cap 주입 지점으로 지목 |

살아 있는 계약은 6건이다. `internal/tmux`의 두 판정을 넘기면서, iter2 감사가 지적한
`SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001`(implemented)과 `SPEC-V3R6-CG-MODE-HARDENING-001`(completed)도 형제
SPEC이 조율할 계약에 들어간다(상태는 iter2 감사 보고서 인용).

## 12. 교차 렌즈 모순

### C1 — `internal/tmux/cg_detect.go`가 죽은 코드인가 (0.3.0 해소)

- *cg-gg-removal-scope*: `IsCGMode`는 `internal/tmux` 밖에 비-테스트 호출자가 없다.
  `grep -rn "tmux\.IsCGMode" --include='*.go' .` → 0행. "dead-but-tested API"로 분류.
- *settings-env-auth*: 같은 파일 `:185-207`을 "the CG/GLM detection SSOT"로 인용하고,
  그 `strings.Contains(..., "z.ai")` 검사가 loopback base URL 아래에서 깨진다고 본다.
  즉 살아 있는 기계로 취급.

**해소.** 두 렌즈는 서로 다른 것을 보았고 둘 다 사실이다. §5.3의 재측정이 보여주듯 `IsCGMode`에는
비테스트 호출 지점이 없고, 두 술어는 `IsCGMode` 안에서만 호출되므로 **프로덕션에서 도달하지 않는다**.
동시에 두 술어는 다른 SPEC의 시험이 고정한 동작을 담고 있어 "그냥 지워도 되는 코드"도 아니다. 이
SPEC은 두 술어를 이관 목록에서 빼고 `SPEC-MOAI-CG-RETIRE-001`(제안)에 넘겼다. 그래서 두 SPEC이 같은
파일을 서로 다르게 다루는 충돌은 더 이상 없다.

### C2 — launch 그룹 구성원의 수와 정체

- *launcher-surface*: `cc`, `cg`, `glm`, `codex` 네 개(`codexCmd`가
  `codex_launcher.go:359`에서 `GroupID: "launch"`를 단다).
- *cg-gg-removal-scope*, *http-proxy-precedent*: `cc`, `glm`, `cg` 세 개.

두 렌즈 모두 문자 그대로 옳다. 서로 **다른 목록**을 읽었기 때문이다 — 하나는 cobra
`GroupID`를, 다른 하나는 하드코딩된 `helpGroupFrequency["launch"] = {"cc","glm","cg"}`
(`help_order.go:31`)를. 이 불일치 자체가 발견이다: launch 그룹의 구성원이 단일 원천이
아니므로 **`gpt` 추가는 최소 세 곳에 해야 한다** — 두 도움말 목록과 `spawnLaunch` 호출 지점
리터럴(§2). 0.1.0은 이를 "두 곳"으로 적었고 iter1 감사가 세 번째를 지적했다.

### C3 — "in-binary proxy"의 실현 가능성

*http-proxy-precedent*만 `syscall.Exec` 문제를 제기했고, 나머지 두 렌즈는 이 구성 요소를
mode switch의 env 주입 단계 대체물로만 서술하며 이 문제를 언급하지 않았다. 이는
**해소된 이견이 아니라 표면화되지 않은 간극**이다. 운영자 결정 D2가 이 간극을 닫는다
(`design.md` §3).

### C4 — `moai gg` 검색 범위

*cg-gg-removal-scope*는 1.1 MB CHANGELOG를 포함해 탐색했고, *launcher-surface*는
CHANGELOG를 명시적으로 제외한 뒤 그 사실을 gap으로 표시했다. 결과는 둘 다 0으로
같다. 넓은 쪽이 우선하며, §1.1의 두 시점 측정이 그 판정을 다시 확인한다.

## 13. 정직한 부재 (채워야 할 gap이 아니라 관측된 negative)

각 항목은 탐색 범위를 함께 적는다.

- `moai gg` — 이 SPEC 디렉터리와 `.moai/reports/`를 제외한 이 트리 전체에서 0건(§1.1, 대조군 1230).
- `internal/gateway` package — 없다. 어떤 reverse-proxy 코드도 없다.
- PKCE / OAuth authorization-code 클라이언트 — production Go에 없다.
- OpenAI/GPT HTTP 클라이언트 — 저장소 어디에도 없다.
- `login` / `logout` cobra 명령 — `internal/cli` 어디에도 없다.
- SSE **중계**(신호 전용 SSE와 구분) — 선례 없다.
- `~/.moai/.env.gpt`, `paths.GptEnvFile`, MoAI 키체인 writer — 없다.
- GPT kanban backend 상수 — 없다(`BackendClaude`, `BackendGLM`뿐, §7.1).
- `MOAI_LAUNCH_PROVIDER` / `EnvMoaiLaunchProvider` — 이 SPEC 디렉터리와 `.moai/reports/`를 제외하고
  0건(§1.4, 대조군 20).
- `IsCGMode`의 비테스트 호출 지점 — 없다(§5.3, 같은 검색에서 정의와 주석 줄은 잡힌다).
- `.moai/specs` 아래 이 구성 요소를 다루는 **다른** SPEC — 없다.

  ```
  $ ls -d .moai/specs/SPEC-*/ | wc -l
       828
  $ ls .moai/specs | wc -l
       832
  $ ls -d .moai/specs/SPEC-*/ | /usr/bin/grep -E 'PROXY|GPT|GATEWAY'
  .moai/specs/SPEC-MOAI-GATEWAY-001/
  ```

  이름에 `PROXY`·`GPT`·`GATEWAY`가 들어간 SPEC 디렉터리는 이 SPEC 자신 하나다. 제외하면 0건이다.
  대조군은 SPEC 디렉터리 828개다(`.moai/specs`의 항목 수 832에는 SPEC 디렉터리가 아닌 항목이 섞여
  있다).
- `team_mode` 마이그레이션 코드 — 없다.
- shell completion 산출물(`scripts/completions*`) — 없다. 빌드/릴리스 시점 생성 여부는
  확인하지 못했다.

## 14. 확인하지 못한 것 (Gaps)

- `moai gg`가 primary checkout이나 미병합 브랜치에 있는지 — 이 작업 트리만 쟀다.
- `internal/cli/launcher.go`(55KB) 전량 판독 — gateway 기동에 관련된 추가 env/exec seam이
  더 있을 수 있다. `removeGLMEnv`(`:380-435`), `applyCCMode`, `applyGLMMode`와 funnel 주요 행만 읽었다.
- **settings `env`와 프로세스 env의 우선순위** — Claude Code 동작이며 측정하지 않았다. `design.md` §6.4는
  이 사실을 전제로 삼지 않도록 쓰였다.
- GLM 라우팅 키를 settings에 쓰는 경로의 **런타임** 전수 — §6.3은 비테스트 Go의 대입 패턴 검색이다.
  구조체 필드 대입이나 문자열 조립으로 쓰는 경로는 이 패턴에 잡히지 않는다.
- `SPEC-V3R6-CG-MODE-HARDENING-001`의 C-7 본문과 `SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001` 본문 — 저자는
  읽지 않았다. 상태와 C-7 위치는 iter2 감사 보고서에서 옮겼다.
- GLM MCP 경로가 `"stream": true`를 쓰는지 — `mcp_glm.go`/`glm_task.go`에서 `stream`
  검색 히트가 없어 비스트리밍으로 보이나 요청 본문 빌더를 읽지 않았다.
- `test-install.yml`의 windows 레그가 실행하는 명령 — 읽지 않았다. Windows 판정 지점은
  `release-pr-multi-os.yml`이며, 그 파일의 트리거·매트릭스·실행·업로드 단계는 0.3.1에서 저자가 직접
  읽었다(§10).
- release PR 워크플로의 windows-latest 레그가 실제로 초록으로 끝난 실행 기록 — 조회하지 않았다. §10은
  워크플로 정의의 판독이다.
- Windows 동작 전반 — 로컬에서 잰 것이 없다. 대화형 TTY와 job control 동작은 CI runner에서 재현되지
  않는다.
- `internal/homestate/runtime.go`가 `Backend` 값을 저장한 뒤 그 값에 따라 분기하는 소비자가 더
  있는지 — 저장 지점만 확인했다.
- 다른 도구가 Anthropic 형태 loopback gateway를 어떻게 구현하는지에 대한 외부 조사 —
  이번 조사는 저장소 증거만으로 답했다.
- docs-site 114개 파일의 locale별 번역 비용 — 파일 수만 셌다.
- `design.md`의 Go 타입 스케치가 컴파일되는지, 기존 `internal/cli` 타입과 이름이 충돌하는지 —
  확인하지 않았다.
- in-process teammate가 lead 프로세스의 env(loopback `ANTHROPIC_BASE_URL`, `MOAI_LAUNCH_PROVIDER`)를 물려받는지, lead
  종료 뒤 남는지, `teammateMode: "in-process"`가 2.1.267에서 실제로 split pane을 막는지 — 코드와 문서만 읽었다(§16).
- tmux pane teammate의 env 상속과 생존 판정 — 0.6.0에서 형제 `SPEC-MOAI-GATEWAY-TEAMMATE-001`(제안)로 옮겼다.
- 검증 요청의 원본 본문과 억제 플래그 없는 조건의 요청 목록 — 없다(§15.6). `plan.md` M1 진입 게이트가 잰다.
- gateway와 저장소 코드에 대한 런타임 검증은 하나도 하지 않았다. §1~§13은 정적 판독이다. §15는 Claude Code
  클라이언트를 Python loopback mock에 붙인 런타임 관측이며, gateway 코드나 실제 provider를 잰 것이 아니다.

## 15. Claude Code 클라이언트 실측 (0.5.0 — 세션 프로브 두 건)

### 15.1 귀속

| 항목 | 값 |
|---|---|
| 측정 | 2026-09-10, 세션 `25b43a41-c0ac-4110-9dad-f3983da6a527` |
| 트리 | `WT-unified-gateway` @ `d060e0d13` |
| 클라이언트 | Claude Code `2.1.267` 실제 TUI. tmux 안에서 `env -i`, 격리 `CLAUDE_CONFIG_DIR`, 작업 디렉터리는 작업 트리 밖(저장소 훅 미실행) |
| 상대 | Python loopback mock. 모든 요청을 Anthropic Messages 형식으로 답하고 upstream에 닿지 않는다. 자격 증명 값은 기록하지 않는다 |
| 프로브 1 | `.moai/state/verify/25b43a41-c0ac-4110-9dad-f3983da6a527/gwprobe/README.md` — `/model <id>` 연속 전환 |
| 프로브 2 | `.moai/state/verify/25b43a41-c0ac-4110-9dad-f3983da6a527/gwprobe2/README.md` — 검증 실패와 picker `s` 경로 |
| 조건 (0.6.0 기록) | 두 프로브 모두 `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1`, `DISABLE_AUTOUPDATER=1`, `DISABLE_TELEMETRY=1`, `DISABLE_ERROR_REPORTING=1` 아래 실행(프로브 1 README "What was run", 프로브 2 README "Same isolation as probe 1"). 요청 기록은 원본 본문이 아니라 mock이 파생한 요약이다(§15.6) |

**무엇을 잰 것인가.** 클라이언트 동작이다. Go gateway나 실제 provider를 잰 것이 아니다. 저자는 두 README와 같은
디렉터리의 `requests.jsonl`을 읽었고, 프로브를 직접 실행하지는 않았다. 두 경로는 `.moai/state/` 아래이며 버전
관리에서 제외된다(무시 규칙 확인 결과 `.gitignore:354:.moai/state/`). 다른 기계나 CI에서는 풀리지 않는다.

### 15.2 관측한 사실

| # | 사실 | 근거 |
|---|---|---|
| F1 | 같은 세션 안 전환이 동작한다. 세션 ID 하나에서, 전환 뒤 turn 요청은 새 모델 ID와 이전 기록(메시지 둘 이상)을 함께 실었다 | 프로브 1(claude → gpt-5.6-sol → glm-5.3-flash → claude, `/model <id>`), 프로브 2(claude → gpt-5.6-sol, picker `s`) |
| F2 | `/model <id>`는 선택 시점에 검증 요청을 보낸다 — `POST /v1/messages?beta=true`, `stream` 거짓(mock의 `bool` 변환값, §15.6), `max_tokens: 1`, 도구 0개(`len(tools or [])`), user 메시지 `"Hi"` 하나. Claude로 되돌아갈 때를 포함해 전환마다 보냈다 | 프로브 1 `requests.jsonl`의 세 검증 요청 |
| F3 | 검증 요청이 401 `authentication_error` 또는 404 `not_found_error`를 받으면 오류 줄만 표시된다("Authentication failed. Please check your API credentials." / "Model 'bad-model-probe' not found"). 확인 대화상자·설정 쓰기·기록 추가가 없고 다음 turn은 이전 모델로 간다 | 프로브 2 관찰 1 |
| F4 | 검증이 성공하면 `/model <id>`는 확인 대화상자("Switch model? … full history gets re-read")를 띄운다. 확인은 추가 요청을 보내지 않으며, 선택은 기본값으로 저장된다("saved as your default for new sessions", 격리 설정 디렉터리의 `settings.json`에 `model`이 전환마다 덮어써짐) | 프로브 1 관찰 3·4 |
| F5 | picker 경로(인수 없는 `/model` → 항목 → `s` → Yes)는 대화상자 앞뒤로 검증 요청을 보내지 않고, `settings.json`을 만들지 않으며, "Set model to gpt-5.6-sol for this session only"를 표시한다. 새 모델로 가는 첫 요청은 다음 turn이다. picker Enter(기본값 저장) 경로는 재지 않았다 | 프로브 2 관찰 2·3 |
| F6 | `ANTHROPIC_CUSTOM_MODEL_OPTION=gpt-5.6-sol`(+`_NAME`)과 `ANTHROPIC_DEFAULT_OPUS_MODEL=glm-5.3-flash`(+`_NAME`) 아래 picker 항목은 "Default (recommended) — currently glm-5.3-flash[1m]", "GLM-5.3-Flash-via-gateway — Custom Opus model", Claude 기본 목록(Sonnet, Sonnet 5 (1M), Haiku), "GPT-5.6-Sol-via-gateway — Custom model (gpt-5.6-sol)", 현재 모델 Sonnet 4.5였다. Opus 고정이 "Default" 행의 대상도 바꾼다 | 프로브 2 picker 항목, 관찰 4 |
| F7 | 프로브 1은 기동 시 `GET /v1/models`를 관측했고 프로브 2는 관측하지 않았다. picker를 연 프로브 2의 항목에 `/v1/models` 유래 ID는 없었다. 프로브 1은 picker를 열지 않았다. 자동 검색이 picker에 반영되는지는 미측정이다 | 프로브 1 결과 문단·`requests.jsonl`, 프로브 2 관찰 5 |
| F8 | 제목 생성 요청은 그 시점에 선택된 모델을 쓴다 | 프로브 1 결과 문단 |
| F9 | 모든 요청이 모델과 무관하게 같은 종류의 bearer 인증 헤더(시작 env의 dummy `ANTHROPIC_AUTH_TOKEN`)를 실었다 | 프로브 1 관찰 6 |
| F10 | `messages`에 `role: "system"` 항목이 섞인다. 프로브 2의 GPT turn은 user 3, system 3, assistant 2였고, system 항목에는 클라이언트가 주입한 알림(예: `<total_tokens>…`)이 실렸다 | 프로브 2 관찰 6 |
| F11 | 프로브 1에서 Claude로 되돌아간 turn의 메시지 수가 8에서 7로 줄었다. 원인은 확립되지 않았다(주입 항목 수의 변동이 그럴듯하나 미확인) | 프로브 1 관찰 5, 프로브 2 관찰 6 |

두 프로브의 turn 요청과 제목 생성 요청은 모두 `stream: true`, `max_tokens` 32000이었다(`requests.jsonl`). 두 프로브
내내 실제 설정 파일의 해시는 기준값 그대로였다 — 프로필 `da2807026f4c…6266`, `~/.claude` `86e2d9b63abd…c5eb`.

### 15.3 이 SPEC에 준 영향

| 사실 | 반영 |
|---|---|
| F1 | `spec.md` §A, `plan.md` M1 — 연속 전환 미관측 서술을 클라이언트 수준 관측으로 정정 |
| F2·F3 | 결정 9 — `REQ-MG-023`, `AC-MG-003`, `design.md` §4.1, `plan.md` M1(진입 게이트, §15.6)·M5 |
| F4 | 결정 10 — 0.6.0에서 형제 `SPEC-MOAI-GATEWAY-PICKER-001`(제안)로 이관 |
| F5 | `REQ-MG-023`의 `s` 경로 경계, `AC-MG-003` (d). picker 안내와 전환 뒤 설정 해시 판정은 0.6.0에서 `SPEC-MOAI-GATEWAY-PICKER-001`(제안)로 이관 |
| F6·F7 | 0.6.0에서 `SPEC-MOAI-GATEWAY-PICKER-001`(제안)로 이관, `design.md` §10. 0.7.0부터 gateway `moai glm`의 tier 슬롯 키는 코어 소관이며 `design.md` §6.7이 F6의 조건과 함께 적는다 |
| F10·F11 | `REQ-MG-015`, `AC-MG-009`, `design.md` §4.2 |
| F8·F9 | 설계 변경 없음. F9는 M0 운반 키 결정의 참고 관측일 뿐 구독 OAuth 공존(T09)에 대해서는 아무것도 말하지 않는다 |

### 15.4 오늘 launcher의 `--model` 동작 (판독, 결정 10의 영향)

`internal/cli/launcher.go:704-708`이 사용자 인수의 `--model`/`-m`을 읽고, `:741`에서 `resolveMainSessionModel`로
해석한 뒤, `:768-770`에서 값이 비어 있지 않을 때만 `--model`을 Claude 인수에 붙인다. `:743-751`의 주석은 기본
프로필에서 빈 모델을 "the normal, intentional state: many setups deliberately omit a model pin so Claude Code falls
back to the user-scope last-choice"라고 적는다. 결정 10은 초기 모델을 언제나 넘기므로, 기본 프로필 `moai cc`에서
사용자 범위 마지막 선택이 시작 모델을 정하던 동작이 사라진다. HEAD `d060e0d13`의 코드 판독이며 런타임에서 확인하지
않았다. `--continue`로 재개할 때 명시 `--model`이 재개된 세션의 모델에 어떻게 작용하는지도 측정하지 않았다.

**0.6.0.** 결정 10은 형제 `SPEC-MOAI-GATEWAY-PICKER-001`(제안)로 옮겼다. 위 판독은 그 SPEC의 입력으로 남긴다.

### 15.5 측정하지 않은 것 (Gaps)

- 실제 OpenAI Responses 변환, Z.AI endpoint, 실제 provider 인증 — mock이 모든 모델을 Anthropic 형식으로 받았다.
- 도구 왕복(T04), provider 간 reasoning·서명 처리(T15).
- picker Enter 경로, 그리고 그 경로로 사용자 지정·고정 항목을 고를 때 검증 요청이 가는지.
- `s`로 전환한 뒤 새 세션이 launcher의 모델로 시작하는지.
- 명시 `--model`이 저장된 `model` 설정보다 우선하는지 — 형제 `SPEC-MOAI-GATEWAY-PICKER-001`(제안)의 게이트 측정(0.6.0에
  `plan.md` M1에서 이관).
- `/v1/models` 자동 검색이 picker에 반영되는지.
- F11의 원인. 프로브 1은 메시지 구성을 기록하지 않았다.
- Go gateway가 검증 요청에 답하는 동작 — 프로브에서는 mock이 답했다.
- 기본 설정 디렉터리(`~/.claude`)에서의 `settings.json` 쓰기 — 격리 디렉터리에서만 쟀다.
- 2.1.267 이외 Claude Code 버전.
- 사용자 지정 picker 항목을 여러 개 둘 수 있는지 — 하나만 쟀다.

### 15.6 측정 조건의 한계 (0.6.0 — iter4 G4-B2)

저자가 이 판에서 다시 연 것: 프로브 1 README "What was run"(`.moai/state/verify/25b43a41-c0ac-4110-9dad-f3983da6a527/gwprobe/README.md`),
프로브 2 README 머리말, 프로브 1 mock `.moai/state/gwprobe/mock.py:84-96`.

| 한계 | 판독 | 결과 |
|---|---|---|
| 억제 플래그 | 프로브 1은 `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1`, `DISABLE_AUTOUPDATER=1`, `DISABLE_TELEMETRY=1`, `DISABLE_ERROR_REPORTING=1`을 켰다. 프로브 2 README는 "Same isolation as probe 1"이라 적는다 | 제품 세션은 첫째 플래그가 launch 정리로 지워진 조건에서 돈다(`REQ-MG-021`). 그 조건의 요청 목록은 측정되지 않았다 |
| `stream` 기록 | `stream = bool(req.get("stream"))`(`mock.py:89`) | 키 없음과 `false`가 같은 기록을 남긴다. `stream: false`는 JSON 값으로 관측되지 않았다 |
| 도구 기록 | `"tools": len(req.get("tools") or [])`(`mock.py:93`) | 키 없음과 빈 배열이 같은 기록을 남긴다 |
| 기록 형태 | 기록 호출에는 파생 필드만 있다(`mock.py:91-94`) | `requests.jsonl`은 원본 본문도 질의 문자열 원문도 담지 않는다 |

프로브 2 mock(README가 가리키는 `.moai/state/gwprobe2/mock2.py`)은 열지 않았다. 그 README는 역할과 60자 미리보기를 기록한다고
적는다. 인식 기준의 근거는 `plan.md` M1 진입 게이트의 원본 캡처로 옮겼다(`design.md` §4.1).

## 16. teammate 표시 판독 (0.6.0 — 결정 12)

### 16.1 코드 (HEAD `ed71054d3`, 저자가 이 판에서 직접 연 행)

- `internal/hook/session_start.go:628-631` — 체인 주석 "When inside tmux, teammates spawn in separate panes for visibility.
  When outside tmux, fall back to "auto" (in-process display)."와 `ensureTeammateMode` 호출. `:638`에서 `ensureTmuxGLMEnv`가
  그 뒤에 돈다.
- `:986-994` 함수 주석 — tmux 밖이면 "removes override (project default "auto" applies)"(`:989`). `:995` 정의, `:996`
  `inTmux := os.Getenv("TMUX") != ""`, `:1021` `desired := "auto"`, `:1022-1023` tmux이면 `"tmux"`. 현재 값과 같으면 쓰지
  않고 돌아간다.
- 같은 함수의 조용한 반환 — 파일 읽기 실패 `:1000-1001`, JSON 해석 실패 `:1006-1007`, 쓰기 실패 `:1060-1064`(오류 로그 뒤
  빈 문자열 반환).
- 판독 결론 — 주석은 tmux 밖에서 키를 지운다고 적지만 코드는 `"auto"`를 쓴다. 이 SPEC은 코드를 따른다.
- `internal/cli/launcher.go:398-402` — `removeGLMEnv`가 `teammateMode`를 지운다. 주석은 "so settings.json default ("auto")
  applies"라고 적는다. `:224` — `applyCCMode`가 CLI `clearTmuxSessionEnv`를 부른다.
- `internal/cli/settings.go:206` — CG 경로의 `m["teammateMode"] = "tmux"`. 호출은 `launcher.go:345`(CG 모드 적용 경로)다.
- `internal/cli/spawn.go:78-79` — `defaultTmuxSpawn` 주석 "runs `tmux new-window` in the caller's current session", 호출 `:88`.
- 배포 템플릿 `internal/template/templates/.claude/settings.json.tmpl`에는 `teammateMode`가 없다
  (`/usr/bin/grep -rn 'teammateMode' internal/cli internal/hook internal/template/templates/.claude/settings.json.tmpl`의 출력에
  그 파일 행이 없다).

### 16.2 Claude Code 문서 (2026-09-11 WebFetch, `https://code.claude.com/docs/en/agent-teams`)

- 표시 방식은 둘이다: in-process와 split pane(tmux 또는 iTerm2).
- "The default is "in-process". Before v2.1.179 the default was "auto"". "Set "auto" to enable split panes when you're already
  running inside a tmux session, or when your terminal is iTerm2 with the it2 CLI installed, falling back to in-process otherwise."
  `"tmux"`는 split pane 모드를 켠다.
- 설정 키는 `teammateMode`, 세션 단위 플래그는 `claude --teammate-mode <mode>`(실험적이며 `--help`에 나오지 않는다).
- in-process teammate의 백그라운드 작업은 "can't outlive the lead's process"라고 적는다.
- teammate는 CLAUDE.md·MCP 서버·skills를 로드하고 lead의 대화 기록은 물려받지 않는다고 적는다. **env 상속은 말하지 않는다.**

이 절은 문서 판독이다. 2.1.267에서 `"in-process"`가 split pane을 막는지, in-process teammate가 lead env를 물려받는지는
측정하지 않았다(`plan.md` M7).


## 17. 0.8.0 M1 캡처 불일치 보정 근거 (2026-09-11)

기준선은 `WT-unified-gateway`, HEAD `81c1d58f9`, Claude Code `2.1.268`이다. 실행·조건·출력의 원 기록은
[실제 TUI 캡처 보고서](../../reports/SPEC-MOAI-GATEWAY-001/m1-capture-gate.md)다. 이 보고서의 0.7.0 당시 BLOCKED
판정과 Sonnet 4.5 출력은 역사적 증거로 보존한다. 네 억제 플래그가 없는 실제 TUI의 POST 다섯 건 가운데
`request-003.json`은 `/model gpt-5.6-sol` 검증 요청이고 `stream` 키가 없다. `max_tokens`는 정수 `1`, 메시지는
`user` 하나이며 `tools` 키도 없다. 원본 SHA-256은 `0c65416916579d77028e64d1fbbec322055f07b3464e9b3aa1ee0ebe06c133e9`다.
나머지 네 요청(제목 둘·turn 둘)은 `stream: true`, `max_tokens: 32000`이다. 사용자 결정에 따라 `stream` 키 부재와
명시 JSON `false`를 허용하고 나머지 세 조건을 함께 요구하도록 보정한다. `false`는 이번 원문에서 관측하지 않았으며
기존 계약을 보존하는 허용값이다. `null`·문자열·`true`는 음성 변형으로 유지한다.

원본은 `.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/m1-raw/`에 있다.
추적 픽스처는 개인정보·기계 경로를 확인하고 비식별 변경 이력과 원본 해시를 연결한 뒤 고정해야 한다. 원문 열람은
형태 확인의 근거이며 제품 인식기 PASS나 모든 정상 작업 요청과의 완전한 구분을 증명하지 않는다.

[M0 보고서](../../reports/SPEC-MOAI-GATEWAY-001/m0-auth-gate.md)는 별도 세션 헤더와 Bearer 공존을 관측했지만
upstream 429로 정상 응답·refresh를 확인하지 못했다. 판정은 INCONCLUSIVE이며 인증 방식의 음성도 OAuth PASS도 아니다.
사용자가 후속 Claude 실서비스 시험을 2026-09-11 19:00 Asia/Seoul 이후로 지정했다. Opus 5·Sonnet 5의 실제 모델 ID와
계정 접근 가능 여부를 확인하여 시험하며, Sonnet 4.5 역사 기록으로 이를 대체하지 않는다. T09 양성 게이트는 유지한다.


## 18. M5 변환 결정의 공식 근거와 운반 Gap (2026-09-11)

- [OpenAI reasoning](https://developers.openai.com/api/docs/guides/reasoning): stateless `store: false` 응답의 reasoning item에
  암호화 데이터가 포함되며 후속 호출에 재전달한다. 현재 문서는 `reasoning.encrypted_content` include를 호환 입력으로
  인정하나 필수는 아니라고 설명한다. 이는 OpenAI 프로토콜 설명이며 Claude Code가 임의 envelope를 보존한다는 근거가 아니다.
- [Function calling](https://developers.openai.com/api/docs/guides/function-calling): 도구 호출과 함께 반환된 reasoning item도
  도구 결과와 함께 반환해야 한다. Responses에서 strict를 명시 false로 두면 schema 자동 strict 처리를 피할 수 있다.
- [Anthropic Messages](https://platform.claude.com/docs/en/api/http/messages): redacted_thinking의 data는 공급자가 반환한
  불투명 암호화 값이며 후속 대화에서 그대로 전달해야 한다. 이 설명은 다른 공급자 데이터를 담은 MoAI envelope를
  Anthropic이 수용한다거나 Claude Code TUI가 보존한다고 보증하지 않는다.
- Responses create 스키마의 system role 허용은 오케스트레이터가 공식 스키마를 확인해 전달한 입력이다. 이번 작성자의
  [직접 열기](https://developers.openai.com/api/reference/resources/responses/methods/create)는 응답 크기 제한으로 실패했다.
  system 위치 보존은 M5의 문서 결정이며 실제 수용은 golden 및 upstream 실행으로 각각 구분해 검증한다.

숨겨진 reasoning을 공개 text로 합치지 않는 것이 목표다. 전역 응답 ID 상태를 도입하지 않고 왕복시키는 envelope는
PROBE ONLY다. 전체 envelope 유실은 그 envelope 내부 필드 검사로 탐지할 수 없으므로 공개 tool ID 의존성 또는 세션
lineage 가운데 하나의 부재 탐지 계약을 추가 검토해야 한다. 두 대안은 이번 문서에서 채택하지 않았다. 실제 TUI 보존·
resume·parallel 요청 분리·provider 간 제거 시험 전에는 제품 호환성이나 전체 GPT 구현 완료를 주장하지 않는다.

## 19. 19시 이후 완료 관측과 0.9.0 계약 입력

기준은 WT-unified-gateway / HEAD 81c1d58f9 / Claude Code 2.1.268이다. §17·18의 INCONCLUSIVE와 미선택 후보 서술은
그 작성 시점 기록이며, 아래 후속 근거와 0.9.0의 선택된 설계 후보로 구별한다. 제품 전체 PASS를 선언하지 않는다.

- [M0 정상](../../reports/SPEC-MOAI-GATEWAY-001/m0-after19-observation.md): Opus 5·Sonnet 5 실제 HTTP 200과 exit 0,
  별도 X-MoAI-Session-Token 및 OAuth Bearer/beta 공존. 로컬 토큰 upstream 제거.
- [직접 refresh](../../reports/SPEC-MOAI-GATEWAY-001/m0-refresh-transport-observation.md): 19:19 KST Opus 5의 token POST 200과
  같은 본문 재전송 200의 Bearer hash 일치. 임시 관측 CA는 child 한정이며 제품 구성품이 아니다. Sonnet 직접 refresh는 별도 미실행.
- [구독 실측](../../reports/SPEC-MOAI-GATEWAY-001/auth-responses-runtime-observation.md): 네 GPT 직접 text/함수 후속 200,
  Content-Type 없음, UTF-8 SSE, output_item.done의 출력과 completed.output 빈 배열. Astra max_output_tokens는 HTTP 400이며
  detail 문자열의 알려진 unsupported 분류다. 표준 error.type/code 존재를 주장하지 않는다. 사용자 출력 상한 결정은 아직 대기다.
- 같은 보고서의 Luna no-tool 응답은 reasoning→message, encrypted 1400바이트다. 원래 item을 보존한 후속 요청은 200이며
  request SHA-256은 3d6ae8372d5a671bb2e3d62bce1638229dcaf9afa25a48732c02591edd56be9c이다. 삭제 mutant·필수성은 미판정이다.
- [native resume](../../reports/SPEC-MOAI-GATEWAY-001/picker-resume-runtime-observation.md): 두 프로세스의 정확한 UUID 재개에서
  합성 68바이트 carrier·172바이트 tool pair ID·metadata.user_id 내부 session_id가 유지됐다. 둘 다 exit 0이며 local mock이다.
  사용자 설정/실계정/모든 continue 변형의 검증을 대신하지 않는다. UUID continuity는 요청 인증의 증거가 아니다.
- [정규화 비교](../../reports/SPEC-MOAI-GATEWAY-001/after19-contract-proposal.md): 같은 raw request003/004에서 공개 이전 5개
  message prefix는 content 문자열/단일 text block과 cache_control만 정규화하면 일치한다. SHA-256은
  fa2c3d9ddfa7b58bdc5d6f5b4e89793a7fc12555899cfdff90b00f46c288cbfd다. system 본문은 삭제하지 않았다.

0.9.0은 이 제한된 양성 관측을 근거로 design §4.3의 대화별 hash-only receipt와 가역 tool ID를 선택된 설계 후보로 둔다.
실제 opaque의 native 왕복·유실 음성·동시/분기·foreign 전환·compaction·재개와 독립 계획 감사 전에는 활성화하지 않는다.

### 19.1 Private config와 native 보안 저장소 분리의 제한된 근거

[picker-settings-source-observation.md](../../reports/SPEC-MOAI-GATEWAY-001/picker-settings-source-observation.md)는
2.1.268 설치 source의 CLAUDE_SECURESTORAGE_CONFIG_DIR 경로와 synthetic 네 namespace 조합을 판독·실행했다.
private CLAUDE_CONFIG_DIR와 원래 secure namespace의 실제 Opus 5 한 요청은 200·exit 0·정확한 OK였고 원본 여섯 설정
경로의 existence/hash/symlink 상태가 같았다. 원래 두 override가 없는 실제 조합은 secure override를 명시 빈 문자열로
보존했다. 원래 config의 명시 빈 문자열은 별도 미검증이다. 공개 지원 문서나 모든 버전·플랫폼을 입증하지 않는다.

그 관측의 임시 config는 시험 종료 때 삭제되었으므로 retained config의 재개 증거가 아니다. 별도 exact resume 관측은
두 프로세스 동안 동일 config를 유지했다. design §4.3은 두 조건을 제품에서 함께 유지하도록 retained family 수명을
정했으며 실제 합친 통합·continue/선택/fork·설정 격리는 구현 후 시험해야 한다. receipt 저장에는 hash만 허용하고,
native가 쓰는 retained transcript를 supervisor 종료 시 지우지 않는다.

### 19.2 Native policy와 title 수용 근거

[native-policy-readiness-observation.md](../../reports/SPEC-MOAI-GATEWAY-001/native-policy-readiness-observation.md)는
실제 raw003~008의 현재 검증 경계 거절을 기록한다. 정책별 첫 실패 이후를 전수 실행한 증거는 아니다.
[native-policy-contract-proposal.md](../../reports/SPEC-MOAI-GATEWAY-001/native-policy-contract-proposal.md)의 공식 문서/
Codex source와 [native-policy-plan-review.md](../../reports/SPEC-MOAI-GATEWAY-001/native-policy-plan-review.md)의 bounded
변경 준비 PASS를 근거로 native policy를 보강했다. 구현/제품 활성화의 PASS와 구별한다.

[title-policy-runtime-observation.md](../../reports/SPEC-MOAI-GATEWAY-001/title-policy-runtime-observation.md)의 네 exact
GPT 요청은 high/strict title schema로 각각 HTTP 200·SSE completed·title:string 단일 JSON이었다. format name은
moai_title_preflight였으며 제품 후보 moai_native_output과 다르다. 일반 schema·native adaptive 동일 계산량·keep-all
서버 동등성·제품 변환은 이 관측으로 검증하지 않았다. output cap 생략은 별도 사용자 승인을 따른다.

### 19.3 공식 Codex shipped context 선언과 실제 계정의 구분

[codex-context-catalog-observation.md](../../reports/SPEC-MOAI-GATEWAY-001/codex-context-catalog-observation.md)의 기준은
제품 WT와 별개인 pristine official snapshot `5a9eb145c4c05fcfc7158d7c25b80e1322eccae1`이다. models.json SHA-256은
`897614e513a591b362f76c34e4b4d31af5809a2b52f8c0ab57f2990c5bf4ab3d`다. 이 작성자는 부모의 원문 추출 보고서를 읽었다.

| exact model | context_window 기본 선언 | max_context_window override 최대 선언 | 기본 effective percent 계산 |
|---|---:|---:|---:|
| gpt-6-astra | 272000 | 872000 | 272000×95/100=258400 |
| gpt-5.6-sol | 272000 | 872000 | 258400 |
| gpt-5.6-terra | 272000 | 872000 | 258400 |
| gpt-5.6-luna | 272000 | 872000 | 258400 |

95%는 그 source의 usable context headroom이며 생성 token 상한이 아니다. auto_compact_token_limit null과 source의
기본 계산, truncation_policy.limit를 실제 계정 context 수치로 바꾸지 않는다. app-server Model 응답의 읽은 구조에
context 수치가 없으므로 모델 목록 조회만으로 live context를 측정했다고 주장하지 않는다.

snapshot은 image/parallel 등의 선언도 갖지만 이번 SPEC 보강에서 제품 capability를 켜지 않는다. 872000은 default가
아니며 현재 endpoint/account의 실제 최대 입력량·tokenizer·Claude/GLM native cap는 미확인이다. 공개 API 문서의
다른 context 숫자와 이 shipped snapshot을 합쳐 하나의 실계정 수치로 만들지 않는다. native cap의 근거는 plan의 연구
게이트에서 확보하고 실제 계량·큰 입력 수용을 독립 검증한다.

## 19.4 2026-09-12 구독 endpoint와 Claude Code 2.1.268 후속 실측

제공된 native debug log에서 `gpt-5.6-sol` 요청은 `dispatching to firstParty` 뒤
`POST /v1/messages`에서 HTTP 400 `invalid_request_error`로 끝났다. 후속 진단
빌드의 안전한 shape 계측은 실제 입력에 `thinking.type=adaptive`,
`thinking.display=omitted`, 12개 tool 중 하나의 `defer_loading=true`가 있음을
확인했다. 고정 ChatGPT 구독 endpoint에 client `stream:false`를 그대로 보낸
경우에는 응답 detail이 `Stream must be set to true`였다.

AuthPKCE 비스트림 요청을 upstream `stream:true`로 바꾸고 SSE를 수집하는 경로를
loopback 시험으로 고정했다. 실제 SSE에는 `response.output_item.added`의
`reasoning` item(added 시 status 없음, `encrypted_content`·`summary` 포함)과
message item의 `phase=final_answer`가 나타났다. `phase`는 공식 Responses
스키마의 `commentary`·`final_answer`·`null` 범위로 제한해 검증하며 대상
Messages 형식에는 옮기지 않는다. reasoning item의 opaque 왕복과 receipt 연결은
현재 활성화 게이트 밖이므로 adapter는 이를 502 명시 오류로 남긴다. 따라서 이번
실측은 stream 강제·필드 shape 호환과 실패 경계를 보강한 것이며, 실제 GPT
생성 성공·추론량 동등·same-provider resume·tool_reference 후속 로딩을
입증하지 않는다.
