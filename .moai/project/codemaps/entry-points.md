# 진입점

> `/moai codemaps`로 생성됐습니다.

**측정 트리**: worktree `.claude/worktrees/t476`, 브랜치 `WT-codemaps-progress`, HEAD `25a3212a9`
**측정**: 2026-09-04

---

## `main()`

배포되는 바이너리는 하나입니다.

- **`cmd/moai/main.go`** — `cli.Execute()`를 호출하고, 에러가 `ExitCoder`를 실으면
  `cli.ResolveExitCode`로 코드를 꺼내 종료합니다. 래핑된 `*exec.ExitError`는 **의도적으로
  거부**하므로 서브프로세스의 종료 코드를 그대로 채택하지 않습니다 — 이것이 rc=128 무성 실패를
  막습니다.

나머지 `main()` 4개는 배포 대상이 아닌 도구입니다:
`internal/template/scripts/gen-catalog-hashes.go`, `scripts/i18n-validator/main.go`,
`scripts/docs-version-snapshot/main.go`, `scripts/convert-nextra-to-hextra/main.go`.

---

## Cobra 명령 트리

**루트**: `internal/cli/root.go` — `rootCmd = &cobra.Command{Use: "moai", ...}`

**실행 심**:

```
root.go Execute()
  → initConsole()
  → configureLogging(args)
  → (비-trivial 명령일 때만) InitDependencies()
  → reorderRootHelpCommands
  → executeRoot
  → fang.go runFang(ctx, cmd)      # charm.land/fang/v2 가 help·에러·--version·completion 렌더링
```

**lazy-init 패스**: `root.go`의 `trivialCommands` 맵에 9개가 있습니다 —
`--version` · `version` · `-v` · `help` · `--help` · `-h` · `completion` · `cc` · `cg` · `glm`.
이들은 의존성 그래프 조립을 건너뜁니다. `cc` / `cg` / `glm`이 포함된 이유는 `syscall.Exec`로
프로세스를 통째 교체하기 때문입니다.

**등록 사이트가 두 갈래**입니다.

1. **`root.go`의 `init()`** — 명시적 `rootCmd.AddCommand(...)` 26회.
   worktree, agentlint, statusline, ast-grep, ast-edit, telemetry, constitution, state, tokens,
   clean, navigator 5종(enrich/sync/tiers/route/fix), migration, chain, harness-router,
   tool-policy, mcp-server, mcp, inventory, preference, model, plan, feedback, inbox.
2. **자기 파일의 `init()`에서 스스로 등록** — `AddCommand`를 호출하는 파일이 **64개**입니다
   (`grep -rl "AddCommand" internal/cli --include='*.go' | grep -v _test`).
   `hook.go`, `todo.go`, `kanban.go`, `glm.go`, `cc.go`, `update.go`, `doctor.go`, `spec.go`,
   `gate.go`, `graph.go`, `goal.go` 등이 이 방식입니다.

**합성 루트**: `internal/cli/deps.go` — `type Dependencies` + `InitDependencies()`.
Config · Git(Repository/Branch/Worktree) · HookRegistry · HookProtocol · UpdateChecker/Orchestrator ·
LoopController · Logger · PerfTiming을 조립하고 전역 변수 `deps *Dependencies`로 노출합니다.

**의도적 미등록 1건**: `root.go`의 주석이 밝히듯 `newHarnessCmd()`는 폐기된 팩토리로
**의도적으로 트리에 등록되지 않고** 컴파일 가능 상태로만 남아 있습니다. 라이브 등록은
`newHarnessRouterCmd()` 하나입니다.

---

## 훅

네 층으로 내려갑니다.

**1. 바깥쪽 배선** — `internal/template/templates/.claude/settings.json.tmpl`.
훅 엔트리 38개가 모두 같은 모양입니다:

```
"command": "bash"
args: ["-c", "[ -f \"$0\" ] && exec bash \"$0\"; ...missing 로그 후 exit 0",
       "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/handle-<event>.sh"]
```

래퍼 스크립트가 없으면 `.moai/logs/hook-missing.log`에 남기고 **exit 0으로 fail-open** 합니다.

**2. 셸 래퍼** — `internal/template/templates/.claude/hooks/moai/` 아래 47개 `.sh` / `.sh.tmpl`.
예: `handle-pre-tool.sh.tmpl`이 `printf '%s' "$payload" | moai hook pre-tool`.

**3. CLI 디스패처** — `internal/cli/hook.go`의 `hookCmd`. `init()`에서 26개 이벤트 서브커맨드를
`hook.EventType`과 함께 일괄 등록하고, `--harness` persistent flag(claude / codex)가 붙습니다.
그 밖에 `hook list`, `agent-hook`, `harness-observe`(3종), `spec-status`,
`session-start-compact`, `security-{scan,turn,commit}`, `harness-classify`, `codex-review-gate`,
`multi-review-gate`가 별도 `RunE`로 붙습니다.

**4. 핸들러 등록과 디스패치** — `internal/cli/deps.go`에서 `deps.HookRegistry.Register(...)`가
**30회** 호출되고, `internal/hook/registry.go`의 `(*registry).Dispatch`가 체인을 돌립니다.
부가로 `runAlwaysRunTail`, `defaultOutputForEvent`, 비동기 trace writer 플러시 배리어 `Shutdown`이
있습니다.

---

## MCP 서버 표면

- **명령**: `internal/cli/mcp_server.go`의 `newMCPServerCmd()` — `root.go`에서 등록. stdio
  JSON-RPC이고 `mark3labs/mcp-go` SDK는 전송만 담당합니다. **기본 off**이며 `.mcp.json`
  프로비저닝은 opt-in입니다.
- **도구 수**: `mcp_server.go` 안 `add("...")` 호출 **28회**. 카탈로그
  `internal/mcp/catalog.go`는 **29개**를 선언합니다 — 차이 1개는 `audit_multi`로,
  `mcp_audit_multi.go`에서 별도 등록됩니다.
- **도구 목록**: `session_list`, `goal_status`, `goal_arm`, `spec_progress`, `verify_snapshot`,
  `verify_trend`, `spec_audit`, `spec_drift`, `audit_cache`,
  `codex_{audit,setup,task,job_status,job_result,job_cancel}`,
  `glm_{task,job_status,job_result,job_cancel,audit}`, `audit_multi`,
  `session_msg_{register,list,send,poll}`,
  `graph_{file_api,find_code,trace_calls,shortest_path}`.
- **가드**: `mcp.yaml`에서 도구별 활성화를 읽고, `add()` 헬퍼의 첫 인자가 `mcp.NewTool` 이름 및
  카탈로그와 일치해야 한다는 계약을 가드 테스트가 강제합니다
  (`mcp_annotation_guard_test.go`, `mcp_boundary_test.go`).
- **엔트리 관리와 서버 실행은 별개 서브트리**입니다 — `newMCPCmd()`(`mcp.go`)가 `.mcp.json`의
  add/remove/list를, `mcp-server`가 실행을 담당합니다.
