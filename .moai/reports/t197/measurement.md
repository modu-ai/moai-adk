# t197 — Codex 전용 런처: plan-phase 실측 기록

- 트리: worktree `.claude/worktrees/t197`, 브랜치 `WT-codex-launcher`, HEAD `9280c96b3`
- t88(M4) 생성물 포함 확인: `git merge-base --is-ancestor 7b217da7c HEAD` → rc 0
- 측정 일자: 2026-08-24
- 로컬 codex: `codex-cli 0.149.0` (`/Users/goos/.local/bin/codex`)

## M-1. `moai codex` 부재 (카드 전제 확인)

```
$ go build -o /tmp/moai-t197 ./cmd/moai && /tmp/moai-t197 --help
LAUNCH COMMANDS:
  cc …   Launch Claude Code with Claude backend
  glm …  Launch Claude Code with GLM backend
  cg …   Launch Claude Code with Claude + GLM hybrid mode
```

`codex` 를 `Use:` 로 갖는 최상위 커맨드는 없다 (`grep -rn 'Use: *"codex' internal/cli/*.go` → `hook.go:221 codex-review-gate` 하위 커맨드 1건뿐). 카드 전제 성립.

## M-2. auth 상태가 항상 `unknown` — 원인 확정

재현 (MCP 도구, 같은 트리):

```
mcp__moai__codex_setup →
{"installed":true,"binary":"/Users/goos/.local/bin/codex",
 "version":"codex-cli 0.149.0","auth_provider":"unknown", …}
```

그런데 실제 로그인은 되어 있다:

```
$ codex login status
Logged in using ChatGPT
$ codex login status >/dev/null 2>&1; echo $?
0
```

스트림 측정 — 이 출력은 **stdout 이 아니라 stderr** 로 나간다:

```
$ codex login status 2>/dev/null | wc -c       # stdout
0
$ codex login status 2>&1 1>/dev/null | wc -c  # stderr
24
```

기전: `internal/cli/mcp_codex.go:273 realCodexRunner.run` 이 `cmd.Stderr = &bytes.Buffer{}` 로 stderr 를 **버리고** stdout 만 반환한다 → `classifyCodexAuth` (:1325) 가 빈 문자열을 받아 `codexAuthUnknown` 으로 fail-open. 분류기의 패턴 문제가 아니라 **읽는 스트림이 틀렸다**.

영향 표면 3곳: `codex_setup` MCP 도구, `moai web` 콘솔의 Codex 상태 카드 (`internal/web/codex_state.go` — `unknown` 일 때 안내문 분기), 신설될 런처의 auth 진단.

## M-3. codex CLI 가 이미 제공하는 표면 (재구현 금지 근거)

```
$ codex --help
  app       Launch the Desktop app (opens the app installer if missing)
  doctor    Diagnose local Codex installation, config, auth, and runtime health
  login     Manage login
  mcp       Manage external MCP servers for Codex
```

`codex doctor` 는 설치·런타임·디스크·git·CODEX_HOME 여유 공간까지 이미 진단한다 (실측 출력 확인). 런처가 진단을 재구현할 이유가 없고, 담당은 **moai 쪽 프로젝트 배선 상태** 로 한정된다.

## M-4. moai 쪽 Codex 배선 표면 (t88 M4 산출물)

- 생성기: `moai init --agent claude|codex|both` (`internal/cli/init.go:393-400`)
- 산출물: `.codex/hooks.json`, `.codex/config.toml` (`internal/codexwiring/codexwiring.go:29,31`), 신뢰 사이드카 (`LoadSidecar`)
- 에이전트 TOML 12종: `internal/template/templates/.codex/agents/moai/*.toml`
- 진단: `moai doctor` 의 `Codex Wiring` 체크 (`internal/cli/doctor_codex.go`) — 배선 없으면 정보성 skip
- 런타임: `moai hook --harness codex` (`internal/cli/hook_harness_codex.go`)
- 이 저장소 자체에는 `.codex/` 가 없다 (`ls .codex` → 없음) — 배선 미적용 프로젝트가 기본값

## M-5. CODEX_HOME 을 읽는 런타임 코드는 없다

```
$ grep -rn "CODEX_HOME" internal/ --include="*.go"   → 0건
```

문서·SPEC·`settings.local.json` 허용 목록에만 등장한다. 런처가 도입해야 할 신규 표면.

## M-6. 기존 런처 3종의 공통 경로는 Claude 전용

`runCC` / `runCG` / `runGLM` 은 모두 `unifiedLaunch(profile, mode, args)` (`internal/cli/launcher.go:56`) 로 수렴하고, 그 본체는 `.claude/settings.local.json` 변형 + Claude 프로필 해석 + `claude` exec 이다. codex 는 **다른 바이너리** 라 이 경로를 재사용할 수 없다 — 별도 실행 경로가 필요하며, 공유 가능한 것은 `--spawn` (tmux) 정도다.
