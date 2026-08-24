# t197 — Codex 전용 런처: plan-phase 실측 기록

각 항목은 **복사해 그대로 실행할 수 있는 명령 + 그 호출의 rc + 명령이 실제 낸 출력 전문** 으로 기록한다. 출력에 손으로 넣은 생략표시(`…`)는 쓰지 않는다. 범위를 좁힐 때는 필터 명령 자체를 온전히 적어 필터가 증거의 일부가 되게 한다.

| 항목 | 값 |
|---|---|
| 트리 | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t197` (worktree) |
| 브랜치 | `WT-codex-launcher` |
| 측정 기준 커밋 | `746177017` (분기 지점 `9280c96b3`) |
| 작업 트리 | 깨끗함 |
| 측정 일자 | 2026-08-24 |
| codex | `codex-cli 0.149.0`, `/Users/goos/.local/bin/codex` |

```
$ git rev-parse --short HEAD; echo "rc=$?"
746177017
rc=0

$ git status --short | wc -l; echo "rc=$?"
       0
rc=0

$ git merge-base --is-ancestor 7b217da7c HEAD; echo "rc=$?"
rc=0
```

마지막 명령의 rc 0 은 t88(M4) 산출물 `7b217da7c` 가 이 트리의 조상임을 뜻한다.

---

## M-1. `moai codex` 부재 (카드 전제)

```
$ go build -o /tmp/moai-t197 ./cmd/moai; echo "rc=$?"
rc=0

$ /tmp/moai-t197 --help 2>&1 | grep -ci 'codex'; echo "rc=${PIPESTATUS[1]}"
0
rc=1
```

도움말 **전문** 을 대상으로 한 대소문자 무시 검색이 0건이다 (`grep -c` 의 rc 1 = 매치 없음 — 파이프 앞이 죽어 0이 나온 경우와 구분된다). 런처 그룹만 좁혀 보면:

```
$ /tmp/moai-t197 --help 2>&1 | sed -n '/LAUNCH COMMANDS/,/PROJECT COMMANDS/p' | sed 's/ *$//'; echo "rc=${PIPESTATUS[0]}"
  LAUNCH COMMANDS:

    cc [-p profile] [-k [SPEC-ID] | -k --name <role> | -f [N] | -f lane-<n>] [-- claude-args...]             Launch Claude Code with Claude backend
    glm [command] [-p profile] [-k [SPEC-ID] | -k --name <role> | -f [N] | -f lane-<n>] [-- claude-args...]  Launch Claude Code with GLM backend
    cg [-p profile]                                                                                          Launch Claude Code with Claude + GLM hybrid mode

  PROJECT COMMANDS:
rc=0
```

소스 쪽 확인:

```
$ grep -rn 'Use: *"codex' internal/cli/*.go; echo "rc=$?"
internal/cli/hook.go:221:		Use:          "codex-review-gate",
rc=0
```

유일한 히트는 `moai hook` 의 하위 커맨드다. 최상위 `codex` 커맨드는 없다.

## M-2. auth 상태가 항상 `unknown` — 원인 확정

프로브 재현. 호출은 MCP 도구 `mcp__moai__codex_setup` (인자 없음) 이며, 반환 JSON 전문:

```json
{"allow_write":false,"auth_provider":"unknown","binary":"/Users/goos/.local/bin/codex","enable_review_gate":false,"installed":true,"node_bridge":false,"version":"codex-cli 0.149.0"}
```

(MCP 도구 호출에는 셸 rc 가 없다 — 도구가 오류 없이 위 결과를 반환한 것이 성공 신호다.)

`installed` 과 `version` 은 맞는데 `auth_provider` 만 `unknown` 이다. 실제 로그인 상태:

```
$ codex login status; echo "rc=$?"
Logged in using ChatGPT
rc=0
```

스트림을 갈라 재면 이 문구는 stdout 이 아니라 stderr 로 나간다:

```
$ codex login status 2>/dev/null | wc -c; echo "rc=${PIPESTATUS[0]}"
       0
rc=0

$ codex login status 2>&1 1>/dev/null | wc -c; echo "rc=${PIPESTATUS[0]}"
      24
rc=0
```

stdout 0바이트 / stderr 24바이트, 두 경우 모두 codex 자체는 rc 0.

기전 — 해당 코드:

```
$ sed -n '273,284p' internal/cli/mcp_codex.go; echo "rc=$?"
func (realCodexRunner) run(ctx context.Context, binaryPath string, args []string, stdin string) (string, error) {
	cmd := exec.CommandContext(ctx, binaryPath, args...)
	cmd.Stdin = strings.NewReader(stdin)
	var out bytes.Buffer
	cmd.Stdout = &out
	// Discard stderr so a noisy codex never corrupts the JSON-RPC stdout parse.
	cmd.Stderr = &bytes.Buffer{}
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return out.String(), nil
}
rc=0
```

```
$ sed -n '1326,1343p' internal/cli/mcp_codex.go; echo "rc=$?"
func classifyCodexAuth(ctx context.Context, binaryPath string) string {
	out, err := codexRunner.run(ctx, binaryPath, []string{"login", "status"}, "")
	if err != nil || out == "" {
		return codexAuthUnknown
	}
	low := strings.ToLower(out)
	switch {
	case strings.Contains(low, "chatgpt"):
		return codexAuthChatGPT
	case strings.Contains(low, "api key"), strings.Contains(low, "apikey"):
		return codexAuthAPIKey
	case strings.Contains(low, "provider"):
		return codexAuthProvider
	default:
		return codexAuthUnknown
	}
}
rc=0
```

stderr 가 버려지므로 `out == ""` 이 되어 `unknown` 으로 fail-open 한다. 위 `switch` 는 **부분 일치** 라, 스트림을 고쳐 오류 문구가 들어오면 그 문구를 성공으로 읽는다 — 이것이 §C.2 설계를 산문 파싱 1순위에서 내린 이유다.

영향 표면 3곳:

```
$ grep -rn 'AuthProvider' internal/web/*.go internal/cli/web.go | grep -v _test; echo "rc=$?"
internal/web/codex_state.go:23:	AuthProvider     string // chatgpt | apiKey | provider | unknown
internal/web/codex_state.go:28:// codexAuthProvider constants mirror the internal/cli classification tokens.
internal/web/codex_state.go:35:	codexAuthProvider = "provider"
internal/web/codex_state.go:45:	return CodexStateView{AuthProvider: codexAuthUnknown}
internal/web/fieldsets_templ.go:3568:			templ_7745c5c3_Var169, templ_7745c5c3_Err = templ.JoinStringErrs(codexAuthProviderLabel(view.CodexState.AuthProvider))
internal/web/fieldsets_templ.go:3580:			if view.CodexState.AuthProvider == "unknown" {
internal/web/schemaform.go:110:// codexAuthProviderLabel maps the auth-provider token the probe emitted to a
internal/web/schemaform.go:114:func codexAuthProviderLabel(provider string) string {
internal/web/schemaform.go:120:	case codexAuthProvider:
internal/cli/web.go:161:				AuthProvider:     s.AuthProvider,
rc=0
```

`codex_setup` MCP 도구 · `moai web` 콘솔 카드 · 신설될 런처가 같은 프로브를 공유한다.

## M-2b. 더 나은 auth 원천 — `auth_mode` 필드

```
$ codex doctor 2>&1 | grep -iE 'auth|login|account'; echo "rc=${PIPESTATUS[1]}"
  ✓ auth         auth is configured
      auth storage mode        File
      auth file                ~/.codex/auth.json
      stored auth mode         chatgpt
      auth mode                chatgpt
      reachability mode        ChatGPT auth
rc=0
```

기계 판독 형태 (`checks["auth.credentials"]` 만 추출하는 필터 명령을 그대로 적는다):

```
$ codex doctor --json > /tmp/t197-doctor.json 2>/dev/null; echo "rc=$?"
rc=0

$ python3 -c 'import json;d=json.load(open("/tmp/t197-doctor.json"));print(json.dumps(d["checks"]["auth.credentials"],indent=1))'; echo "rc=$?"
{
 "id": "auth.credentials",
 "category": "auth",
 "status": "ok",
 "summary": "auth is configured",
 "details": {
  "auth file": "/Users/goos/.codex/auth.json",
  "auth storage mode": "File",
  "stored API key": "false",
  "stored ChatGPT tokens": "true",
  "stored agent identity": "false",
  "stored auth mode": "chatgpt"
 },
 "remediation": null,
 "durationMs": 0
}
rc=0
```

그러나 런처가 doctor 를 부를 수는 없다:

```
$ time codex doctor --json > /dev/null 2>&1
codex doctor --json > /dev/null 2>&1  11.76s user 9.82s system 46% cpu 46.357 total
```

같은 실행의 Notes 가 이유를 말한다:

```
$ codex doctor 2>&1 | sed -n '/^Notes/,/^─/p'; echo "rc=${PIPESTATUS[1]}"
Notes
   ↑ updates      0.149.1 available (current 0.149.0)
   ⚠ rollouts     31,525 active files · 1.44 GB on disk
   ⚠ threads      rollout scan was incomplete or found bad files
   ⚠ desktop      the desktop security assessment was unavailable - check access to macos gatekeeper diagnostics
─────────────────────────────────────────────────────────────
rc=0
```

doctor 가 읽는 원본 파일은 즉시 읽힌다. 아래 스크립트를 파일로 저장해 실행했고, **스크립트 본문 전문** 을 함께 적는다 (값이 아니라 형태만 출력하며, 마지막 두 줄만 값 — 둘 다 비밀값이 아니다):

```
$ cat /tmp/t197-shape.py; echo "rc=$?"
import json, os
p = os.path.expanduser(os.environ.get("CODEX_HOME", "~/.codex") + "/auth.json")
d = json.load(open(p))
def shape(o, depth=0):
    if isinstance(o, dict):
        return {k: (shape(v, depth + 1) if depth < 1 else type(v).__name__) for k, v in o.items()}
    return type(o).__name__
print(json.dumps(shape(d), indent=1))
print("auth_mode =", repr(d.get("auth_mode")))
print("OPENAI_API_KEY populated:", bool(d.get("OPENAI_API_KEY")))
rc=0

$ python3 /tmp/t197-shape.py; echo "rc=$?"
{
 "auth_mode": "str",
 "OPENAI_API_KEY": "NoneType",
 "tokens": {
  "id_token": "str",
  "access_token": "str",
  "refresh_token": "str",
  "account_id": "str"
 },
 "last_refresh": "str"
}
auth_mode = 'chatgpt'
OPENAI_API_KEY populated: False
rc=0
```

`auth_mode` 가 doctor 의 `stored auth mode` 와 같은 값이고, 서브프로세스 없이 즉시 읽힌다.

**이 머신에서 관측한 조합은 하나뿐이다** — `auth_mode="chatgpt"` + 토큰 4종 존재 + `OPENAI_API_KEY` null. `apikey` 모드와 `provider` 모드의 실제 파일 형태는 관측하지 못했다 (로그아웃·재로그인이 필요하고 이 세션의 범위 밖). 따라서 설계는 알려지지 않은 값을 추측하지 않고 `unknown` 으로 내리며, 모드별 최소 구조 조건을 인수 기준으로 고정한다 (AC-CL-008).

## M-3. codex CLI 가 이미 제공하는 표면 (재구현 금지 근거)

`Commands:` 절 **전문** — 손으로 줄인 곳은 없다:

```
$ codex --help 2>&1 | sed -n '/^Commands:/,/^Arguments:/p' | sed 's/ *$//'; echo "rc=${PIPESTATUS[0]}"
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
  completion        Generate shell completion scripts
  update            Update Codex to the latest version
  doctor            Diagnose local Codex installation, config, auth, and runtime health
  sandbox           Run commands within a Codex-provided sandbox
  debug             Debugging tools
  apply             Apply the latest diff produced by Codex agent as a `git apply` to your local
                    working tree [aliases: a]
  resume            Resume a previous interactive session (picker by default; use --last to continue
                    the most recent)
  queue             Queue a message for an existing session
  archive           Archive a saved session by id or session name
  delete            Permanently delete a saved session by id or session name
  migrate-rollouts  Inspect or migrate legacy local sessions to paginated thread history
  unarchive         Unarchive a saved session by id or session name
  fork              Fork a previous interactive session (picker by default; use --last to fork the
                    most recent)
  cloud             [EXPERIMENTAL] Browse tasks from Codex Cloud and apply changes locally
  exec-server       [EXPERIMENTAL] Run the standalone exec-server service
  features          Inspect feature flags
  help              Print this message or the help of the given subcommand(s)

Arguments:
rc=0
```

본 SPEC 이 위임 대상으로 쓰는 것은 `app` 과 (안내 문구로서) `doctor` · `login` 이다.

## M-4. moai 쪽 Codex 배선 표면 (t88 M4 산출물)

```
$ grep -n 'HooksRelPath =\|ConfigRelPath =' internal/codexwiring/codexwiring.go; echo "rc=$?"
29:	HooksRelPath = ".codex/hooks.json"
31:	ConfigRelPath = ".codex/config.toml"
rc=0

$ find internal/template/templates/.codex/agents/moai -name '*.toml' -type f | sort; echo "rc=${PIPESTATUS[0]}"
internal/template/templates/.codex/agents/moai/builder-harness.toml
internal/template/templates/.codex/agents/moai/e2e-tester.toml
internal/template/templates/.codex/agents/moai/manager-design.toml
internal/template/templates/.codex/agents/moai/manager-develop.toml
internal/template/templates/.codex/agents/moai/manager-docs.toml
internal/template/templates/.codex/agents/moai/manager-git.toml
internal/template/templates/.codex/agents/moai/manager-lead.toml
internal/template/templates/.codex/agents/moai/manager-spec.toml
internal/template/templates/.codex/agents/moai/plan-auditor.toml
internal/template/templates/.codex/agents/moai/super-advisor.toml
internal/template/templates/.codex/agents/moai/sync-auditor.toml
rc=0

$ find internal/template/templates/.codex/agents/moai -name '*.toml' -type f | wc -l; echo "rc=${PIPESTATUS[0]}"
      11
rc=0
```

**정정**: 이 문서의 이전 판은 이 값을 12 로 적었다. `ls … | wc -l` 로 셌는데 이 셸의 `ls` 가 긴 형식 별칭이라 `total` 행이 함께 세어졌다. 실제 값은 **11** 이며, 위처럼 `find -type f` 로 세면 별칭에 영향받지 않는다.

생성기·진단·런타임 경로 (각각 명령 출력으로 확인):

```
$ grep -n 'must be one of: claude, codex, both' internal/cli/init.go; echo "rc=$?"
400:			return fmt.Errorf("invalid --agent value %q: must be one of: claude, codex, both", agent)
rc=0

$ grep -n 'func checkCodexWiring' internal/cli/doctor_codex.go; echo "rc=$?"
38:func checkCodexWiring(root string, verbose bool) DiagnosticCheck {
rc=0

$ grep -n 'codexHarnessFlagValue = ' internal/cli/hook_harness_codex.go; echo "rc=$?"
25:const codexHarnessFlagValue = "codex"
rc=0

$ ls -a .codex; echo "rc=$?"
ls: .codex: No such file or directory
rc=1
```

마지막 rc 1 은 이 저장소 자체가 미배선 상태라는 뜻이고, 그것이 기본값이다.

## M-5. CODEX_HOME 을 읽는 런타임 코드는 0건

```
$ grep -rn "CODEX_HOME" internal/ --include="*.go"; echo "rc=$?"
rc=1
```

출력 없음 + rc 1 = 매치 0건. (`| wc -l` 로 세지 않는다 — 파이프 앞이 죽어도 0 이 나오므로 부재의 증거가 되지 못한다.)

## M-6. 기존 런처 3종의 공통 경로는 재사용 불가

```
$ grep -n "func unifiedLaunch\|func unifiedLaunchDefault" internal/cli/launcher.go; echo "rc=$?"
56:func unifiedLaunch(profileName, modeOverride string, extraArgs []string) error {
129:func unifiedLaunchDefault(profileName, modeOverride string, extraArgs []string) error {
rc=0

$ grep -n "return unifiedLaunch(" internal/cli/cc.go internal/cli/cg.go internal/cli/glm.go; echo "rc=$?"
internal/cli/cc.go:225:	return unifiedLaunch(profileName, "claude", filteredArgs)
internal/cli/cg.go:103:	return unifiedLaunch(profileName, "claude_glm", filteredArgs)
internal/cli/glm.go:294:	return unifiedLaunch(profileName, "glm", filteredArgs)
rc=0
```

세 런처가 모두 같은 함수로 수렴한다. 그 본체는 `.claude/settings.local.json` 변형 + Claude 프로필 해석 + `claude` exec 이므로 (M-6 의 `unifiedLaunchDefault`), 다른 바이너리인 codex 는 이 경로에 얹을 수 없다. 공유 가능한 것은 `spawnLaunch` 다:

```
$ grep -n "func spawnLaunch" internal/cli/spawn.go; echo "rc=$?"
117:func spawnLaunch(out io.Writer, subcommand string, args []string) error {
rc=0
```

## M-7. `codexCommandRunner` 구현체는 3개 (감사 지적으로 정정)

최초 기록은 "프로덕션 1 + 시험 1" 이었다. **틀렸다** — Go 인터페이스는 구조적이라 이름을 grep 해서는 암묵 구현을 찾을 수 없다. 메서드 시그니처로 다시 세면:

```
$ grep -rn ") run(" internal/cli/*.go internal/cli/*_test.go | grep -v glm; echo "rc=$?"
internal/cli/mcp_codex.go:273:func (realCodexRunner) run(ctx context.Context, binaryPath string, args []string, stdin string) (string, error) {
internal/cli/codex_rpc_error_test.go:31:func (stubCodexRunner) run(context.Context, string, []string, string) (string, error) {
internal/cli/mcp_codex_test.go:40:func (f *fakeCodexRunner) run(_ context.Context, binaryPath string, args []string, stdin string) (string, error) {
rc=0
```

프로덕션 1 (`realCodexRunner`) + 시험 2 (`stubCodexRunner`, `fakeCodexRunner`). 인터페이스를 넓히면 시험 더블 **둘 다** 고쳐야 한다 — plan 은 이 사실을 반영해 인터페이스를 넓히지 않는 설계로 바꿨다 (§C.2).

## 미관측 항목 (명시)

아래는 이 세션에서 **관측하지 않았다**. 추정으로 채우지 않고 남긴다:

- `auth_mode` 의 `apikey` / `provider` 모드 실제 파일 형태 (로그아웃·재로그인 필요)
- `codex login status` 의 로그아웃 상태 출력 문구와 rc
- 데스크톱 앱 기동 (`codex app`) 의 실제 동작
- 런처 구현이 아직 없으므로 인자 전달·tmux 인용·크로스 플랫폼 동작 일체
