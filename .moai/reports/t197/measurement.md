# t197 — Codex 전용 런처: plan-phase 실측 기록

모든 항목은 명령 · 종료코드 · 출력을 그대로 적는다. 요약으로 대체한 항목은 없다.

| 항목 | 값 |
|---|---|
| 트리 | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t197` (worktree) |
| 브랜치 | `WT-codex-launcher` |
| 기준 커밋 | `a547e3888` (부모 `9280c96b3` = 분기 시점 origin/main) |
| 작업 트리 | 깨끗함 (`git status --short` → 0줄) |
| 측정 일자 | 2026-08-24 |
| codex | `codex-cli 0.149.0`, `/Users/goos/.local/bin/codex` |

t88(M4) 산출물 포함 확인:

```
$ git merge-base --is-ancestor 7b217da7c HEAD; echo "rc=$?"
rc=0
```

---

## M-1. `moai codex` 부재 (카드 전제)

```
$ go build -o /tmp/moai-t197 ./cmd/moai; echo "rc=$?"
rc=0

$ /tmp/moai-t197 --help 2>&1 | sed -n '/LAUNCH COMMANDS/,/PROJECT COMMANDS/p'
  LAUNCH COMMANDS:

    cc [-p profile] [-k [SPEC-ID] | -k --name <role> | -f [N] | -f lane-<n>] [-- claude-args...]             Launch Claude Code with Claude backend
    glm [command] [-p profile] [-k [SPEC-ID] | -k --name <role> | -f [N] | -f lane-<n>] [-- claude-args...]  Launch Claude Code with GLM backend
    cg [-p profile]                                                                                          Launch Claude Code with Claude + GLM hybrid mode

  PROJECT COMMANDS:

$ /tmp/moai-t197 --help 2>&1 | grep -ci 'codex'
0
```

전체 도움말 어디에도 `codex` 문자열이 없다 (0건 — 잘라낸 범위가 아니라 전량 대상). 소스 쪽:

```
$ grep -rn 'Use: *"codex' internal/cli/*.go
internal/cli/hook.go:221:		Use:          "codex-review-gate",
```

유일한 히트는 `moai hook` 의 하위 커맨드다. 최상위 `codex` 커맨드는 없다.

## M-2. auth 상태가 항상 `unknown` — 원인 확정

프로브 재현 (같은 트리, MCP 도구 `mcp__moai__codex_setup`), 출력 전문:

```json
{"allow_write":false,"auth_provider":"unknown","binary":"/Users/goos/.local/bin/codex","enable_review_gate":false,"installed":true,"node_bridge":false,"version":"codex-cli 0.149.0"}
```

`installed` 과 `version` 은 맞는데 `auth_provider` 만 `unknown` 이다. 실제 로그인 상태:

```
$ codex login status
Logged in using ChatGPT

$ codex login status >/dev/null 2>&1; echo "rc=$?"
rc=0
```

스트림을 갈라 재면 이 문구는 stdout 이 아니라 stderr 로 나간다:

```
$ codex login status 2>/dev/null | wc -c        # stdout 바이트
       0
$ codex login status 2>&1 1>/dev/null | wc -c   # stderr 바이트
      24
```

기전: `internal/cli/mcp_codex.go:273` `realCodexRunner.run` 이

```go
cmd.Stderr = &bytes.Buffer{}          // 버려진다
...
return out.String(), nil              // stdout 만 반환
```

이라 `classifyCodexAuth`(:1325)가 빈 문자열을 받고 fail-open 으로 `codexAuthUnknown` 을 낸다. 분류 패턴이 아니라 **읽는 스트림이 틀렸다**.

영향 표면 3곳: `codex_setup` MCP 도구, `moai web` 콘솔의 Codex 카드(`internal/web/codex_state.go` — `AuthProvider == "unknown"` 분기), 신설될 런처.

## M-2b. 더 나은 auth 원천 — `auth_mode` 파일 필드

`codex doctor` 는 auth 를 구조화해서 안다:

```
$ codex doctor 2>&1 | grep -iE 'auth|login|account'
  ✓ auth         auth is configured
      auth storage mode        File
      auth file                ~/.codex/auth.json
      stored auth mode         chatgpt
      auth mode                chatgpt
      reachability mode        ChatGPT auth
```

기계 판독 형태도 있다 (`codex doctor --json` → `checks["auth.credentials"]`):

```json
{
 "auth.credentials": {
  "id": "auth.credentials", "category": "auth", "status": "ok",
  "summary": "auth is configured",
  "details": {
   "auth file": "/Users/goos/.codex/auth.json",
   "auth storage mode": "File",
   "stored API key": "false",
   "stored ChatGPT tokens": "true",
   "stored agent identity": "false",
   "stored auth mode": "chatgpt"
  }
 }
}
```

그러나 **런처가 부를 수 없다** — 46초 걸린다:

```
$ time codex doctor --json > /dev/null 2>&1
codex doctor --json > /dev/null 2>&1  11.76s user 9.82s system 46% cpu 46.357 total
```

(같은 실행의 Notes 가 이유를 말한다: `⚠ rollouts  31,525 active files · 1.44 GB on disk` — 롤아웃 스캔이 든다.)

doctor 가 읽는 원본 파일은 즉시 읽힌다. `~/.codex/auth.json` 의 **키 구조만** (값은 토큰이라 출력하지 않는다):

```
$ python3 -c "…shape only…"
{
 "auth_mode": "str",
 "OPENAI_API_KEY": "NoneType",
 "tokens": {"id_token": "str", "access_token": "str", "refresh_token": "str", "account_id": "str"},
 "last_refresh": "str"
}
```

분류에 필요한 두 값 (비밀값 아님):

```
$ python3 -c "…auth_mode + key presence…"
auth_mode = 'chatgpt'
OPENAI_API_KEY set: False
```

즉 `<CODEX_HOME>/auth.json` 의 `auth_mode` 가 doctor 의 `stored auth mode` 와 같은 값이고, 서브프로세스 없이 즉시 읽힌다.

## M-3. codex CLI 가 이미 제공하는 표면 (재구현 금지 근거)

```
$ codex --help 2>&1 | sed -n '/^Commands:/,/^$/p'
Commands:
  agents            Browse all agent sessions on the shared local app-server daemon
  exec              Run Codex non-interactively [aliases: e]
  review            Run a code review non-interactively
  login             Manage login
  logout            Remove stored authentication credentials
  mcp               Manage external MCP servers for Codex
  plugin            Manage Codex plugins
  mcp-server        Start Codex as an MCP server (stdio)
  app-server        [experimental] Run the app server or related tooling
  remote-control    [experimental] Manage the app-server daemon with remote control enabled
  app               Launch the Desktop app (opens the app installer if missing)
  …
  doctor            Diagnose local Codex installation, config, auth, and runtime health
  …
```

(전체 목록은 24개 항목이며, 위는 본 SPEC 이 참조하는 행만 발췌한 것이다 — 발췌임을 명시한다.)

## M-4. moai 쪽 Codex 배선 표면 (t88 M4 산출물)

```
$ grep -n 'HooksRelPath =\|ConfigRelPath =' internal/codexwiring/codexwiring.go
29:	HooksRelPath = ".codex/hooks.json"
31:	ConfigRelPath = ".codex/config.toml"

$ ls internal/template/templates/.codex/agents/moai/ | wc -l
      12

$ ls -a .codex 2>&1
ls: .codex: No such file or directory
```

- 생성기: `moai init --agent claude|codex|both` (`internal/cli/init.go:393-400`)
- 진단: `moai doctor` 의 `Codex Wiring` 체크 (`internal/cli/doctor_codex.go`) — 배선 없으면 정보성 skip
- 런타임: `moai hook --harness codex` (`internal/cli/hook_harness_codex.go`)
- 이 저장소 자체는 미배선 상태이며, 그것이 기본값이다

## M-5. CODEX_HOME 을 읽는 런타임 코드는 0건

```
$ grep -rn "CODEX_HOME" internal/ --include="*.go" | wc -l
       0
```

문서·SPEC·`settings.local.json` 허용 목록에만 등장한다. 런처가 도입해야 할 신규 표면.

## M-6. 기존 런처 3종의 공통 경로는 재사용 불가

```
$ grep -rn "func unifiedLaunch" internal/cli/launcher.go
56:func unifiedLaunch(profileName, modeOverride string, extraArgs []string) error {
```

`runCC` / `runCG` / `runGLM` 이 모두 이리로 수렴하고, 본체는 `.claude/settings.local.json` 변형 + Claude 프로필 해석 + `claude` exec 이다. codex 는 다른 바이너리라 이 경로에 얹을 수 없다. 공유 가능한 것은 `spawnLaunch`(tmux) 정도다.

## M-7. `codexCommandRunner` 구현체는 3개 (감사 지적으로 정정)

최초 기록은 "프로덕션 1 + 시험 1" 이었다. **틀렸다** — Go 인터페이스는 구조적이라 이름을 grep 해서는 암묵 구현을 찾을 수 없다. 메서드 시그니처로 다시 세면:

```
$ grep -rn ") run(" internal/cli/*.go internal/cli/*_test.go | grep -v glm
internal/cli/mcp_codex.go:273:func (realCodexRunner) run(ctx context.Context, binaryPath string, args []string, stdin string) (string, error) {
internal/cli/mcp_codex_test.go:40:func (f *fakeCodexRunner) run(_ context.Context, binaryPath string, args []string, stdin string) (string, error) {
internal/cli/codex_rpc_error_test.go:31:func (stubCodexRunner) run(context.Context, string, []string, string) (string, error) {
```

프로덕션 1 (`realCodexRunner`) + 시험 2 (`fakeCodexRunner`, `stubCodexRunner`). 인터페이스를 넓히면 시험 더블 **둘 다** 고쳐야 한다 — plan 은 이 사실을 반영해 인터페이스를 넓히지 않는 설계로 바꿨다 (§C.2).
