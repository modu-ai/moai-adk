# 진입점

> `/moai codemaps`로 생성됐습니다.

**최초 측정 트리**: worktree `.claude/worktrees/t592`, 브랜치 `WT-home-state-rollout`, HEAD `e7bd89ee3`, 2026-09-10
**재측정 트리**: worktree `.claude/worktrees/t869`, 브랜치 `WT-codemaps-refresh`, HEAD `a851b205c`, 2026-09-18 — § `main()`, § Cobra 명령 트리의 모든 수치와 등록 목록, § 훅의 개수 네 가지(설정 엔트리 38, 셸 래퍼 48, 이벤트 서브커맨드 26, `Register` 30), § MCP 서버 표면 전체. 훅 절의 부가 `RunE` 목록은 다시 대조하지 않았습니다. § HOME 상태·웹 콘솔·CI 종료 코드 절은 이번 변경과 무관해 앞 판을 이어받았습니다.

---

## `main()`

배포되는 바이너리는 하나입니다.

- **`cmd/moai/main.go`** — `cli.Execute()`를 호출하고, 에러가 `ExitCoder`를 실으면
  `cli.ResolveExitCode`로 코드를 꺼내 종료합니다. 래핑된 `*exec.ExitError`는 **의도적으로
  거부**하므로 서브프로세스의 종료 코드를 그대로 채택하지 않습니다 — 이것이 rc=128 무성 실패를
  막습니다.

나머지 main 패키지 5개는 배포 대상이 아닌 도구입니다:
`cmd/t657-merge/main.go`(카드 t657의 일회성 큐 병합 도구 — 사용자 verb가 아니며 실제 저장소를
명시적 절대 경로 플래그로만 받는다), `internal/template/scripts/gen-catalog-hashes.go`,
`scripts/i18n-validator/main.go`, `scripts/docs-version-snapshot/main.go`,
`scripts/convert-nextra-to-hextra/main.go`. 산출: `grep -rl '^package main' cmd scripts internal/template/scripts`
(i18n-validator는 한 패키지에 파일 셋).

**빌드 타깃으로만 진입하는 방출기 2개**는 `main()`이 아니라 Makefile 타깃과 골든 테스트를
통해 실행됩니다 — `internal/template/agentemit`(`make agents-emit`, `.md` × 매니페스트 →
`.codex/agents/*.toml`)와 `internal/template/commandemit`(`make commands-emit`,
`.claude/commands/moai/*` → `.agents/skills/moai-<command>/SKILL.md`). 둘 다 비테스트
fan-in이 0인 것은 고아라서가 아니라 이 진입 형태 때문입니다.

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

**lazy-init 패스**: `root.go`의 `trivialCommands` 맵에 10개가 있습니다 —
`--version` · `version` · `-v` · `help` · `--help` · `-h` · `completion` · `cc` · `cg` · `glm`.
이들은 의존성 그래프 조립을 건너뜁니다. `cc` / `glm`이 포함된 이유는 `syscall.Exec`로
프로세스를 통째 교체하기 때문입니다. **`cg`는 명령이 아니라 은퇴 토큰**입니다 — 맵의 주석이
"retired token: never initialize launch dependencies"라고 적고, `moai --help`의 LAUNCH COMMANDS
그룹에는 `cc` · `glm` · `codex`만 렌더됩니다.

**등록 사이트가 두 갈래**입니다.

1. **`root.go`의 `init()`** — 명시적 `rootCmd.AddCommand(...)` **30회**.
   worktree, agentlint(agent/workflow 2종), statusline, ast-grep, ast-edit, telemetry,
   constitution, state, tokens, clean, **skills**, navigator 5종(enrich/sync/tiers/route/fix),
   migration, **chain**, harness-router, tool-policy, tool, mcp-server, mcp, inventory, preference,
   model, plan, feedback, inbox.
   - `skills`(`newSkillsCmd()`, `root.go:187`) — `moai skills disable <name> --codex` 형태로
     **계층을 플래그로 명명**하는 스킬 노출 제어 트리. `--codex`가 필수인 것이 opt-in의
     기계적 형태이며, 어떤 프로젝트 설정 키도 이 verb를 구동하지 않습니다(사용자 HOME에
     쓰는 일을 프로젝트 설정이 요청하게 두지 않는다).
   - `chain`(`newChainCmd()`, `root.go:214`) — 워크트리 세션 origin-trail 원장 조회·정리.
2. **자기 파일의 `init()`에서 스스로 등록** — `AddCommand`를 호출하는 파일이 **70개**입니다
   (`grep -rl "AddCommand" internal/cli --include='*.go' | grep -v _test`).
   `hook.go`, `todo.go`, `kanban.go`, `glm.go`, `cc.go`, `update.go`, `doctor.go`, `spec.go`,
   `gate.go`, `graph.go`, `goal.go`, `integration.go` 등이 이 방식이고, 이 판에서
   `gtd.go`(`NewGTDCommand()` — todo 명령 트리를 감싸 `Use`만 `gtd`로 바꾼 두 번째 이름)와
   `slot.go`(`moai slot` — 무거운 실행용 세션 간 자원 임대)가 더해졌습니다.
   비테스트 `AddCommand(` 호출은 모두 **219회**, 그중 `rootCmd.AddCommand(`는 **65회**입니다.

**합성 루트**: `internal/cli/deps.go` — `type Dependencies` + `InitDependencies()`.
Config · Git(Repository/Branch/Worktree) · HookRegistry · HookProtocol · UpdateChecker/Orchestrator ·
LoopController · Logger · PerfTiming을 조립하고 전역 변수 `deps *Dependencies`로 노출합니다.

**의도적 미등록 1건**: `root.go`의 주석이 밝히듯 `newHarnessCmd()`는 폐기된 팩토리로
**의도적으로 트리에 등록되지 않고** 컴파일 가능 상태로만 남아 있습니다. 라이브 등록은
`newHarnessRouterCmd()` 하나입니다.

### 창(window)을 잡기 전에 도는 선행 조건

`moai integration acquire`는 창을 기록하기 **전에** 호출자 트리를 단정합니다 —
`internal/cli/integration_settings_drift.go`가 tracked `.claude/settings.json`의 워킹 사본
드리프트를 재고, 적중이면 사본을 보존한 뒤 원장 한 줄을 남깁니다. 같은 술어를 창 없이
물을 수 있는 독립 verb가 `moai integration preflight [경로]`이며, **두 표면 중 어느 쪽도
뺄 수 없습니다** — 선행 조건이 없으면 검사가 사회적 약속이 되고, 독립 verb가 없으면 창을
잡지 않고는 물을 방법이 없습니다.

---

## HOME 상태와 Factory 복구 진입점

- `moai migrate home-state` — 기본은 읽기 전용 점검입니다. `--apply`만으로는 쓸 수 없고
  `--verified-live`를 함께 줘야 하며, 내부에서 현재 HEAD에 대한 검증 증거와 두 번의
  zero-active runtime census를 다시 확인합니다.
- `moai migrate home-state recover` — 소유 프로세스가 죽은 migration marker만 복구합니다.
  소유자 상태가 불명확하거나 살아 있으면 fail-closed입니다.
- `moai migrate home-state rollback` — 백업 manifest, SHA-256, 프로젝트 키·루트가 일치하는
  검증된 백업만 복원합니다.
- `moai factory handoff recover-resume --id <id> --expected-token <token> --decision <fail|requeue>`
  — v1 레거시 claim이 자동 판정 불가능할 때 쓰는 명시적 운영 복구 표면입니다. 현재 소유자가
  살아 있거나 상태가 불명확하면 재점유하지 않습니다.

런타임 쪽 진입점은 별도 명령이 아니라 공통 gate입니다. SessionStart, MCP 서버, Factory가
`internal/homestate` admission lock을 잡고 migration marker를 검사한 뒤에만 상태를 엽니다.
프로필 lease는 런처에서 provisional 생성, 자식 PID로 transfer, SessionStart에서 session ID를
enrich하고 SessionEnd에서 release하는 흐름입니다.

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

**2. 셸 래퍼** — `internal/template/templates/.claude/hooks/moai/` 아래 48개 `.sh` / `.sh.tmpl`.
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

### 감사 영수증 가드 — 세 이벤트에 걸친 하나의 판정

이 판에서 더해진 `internal/hook/audit_receipt_guard.go`는 훅 표면 중 드물게 **세 이벤트를 하나의
판정으로 엮습니다**. 읽고 쓰는 기록은 전부 `internal/auditreceipt`가 소유합니다.

| 이벤트 | 하는 일 |
|---|---|
| `SubagentStart` | 감사자 서브에이전트(`plan-auditor` · `sync-auditor`) 1건당 시작 마커를 쓴다 |
| `SubagentStop` | 판정 줄을 파싱하고 `CheckCitedReceipts`로 인용된 영수증이 실재하는지 보며, 실패하면 거부 기록을 쓴다 |
| `PreToolUse` | 거부가 미해소인 동안 페이즈 진입 에이전트(`manager-develop` · `manager-docs` · `manager-git`)의 spawn을 거절한다 |

세 이벤트를 나눠 읽으면 각각 멀쩡해 보이므로 함께 적습니다 — 첫째가 없으면 둘째가 비교할
기준이 없고, 셋째가 없으면 거부가 아무것도 막지 않습니다. 게이트 자체는 opt-in이며
`.moai/config/sections/workflow.yaml`이 **문자열 `required`일 때만** 켜집니다(`CodexGateRequired`).

---

## MCP 서버 표면

- **명령**: `internal/cli/mcp_server.go`의 `newMCPServerCmd()` — `root.go`에서 등록. stdio
  JSON-RPC이고 `mark3labs/mcp-go` SDK는 전송만 담당합니다. **기본 off**이며 `.mcp.json`
  프로비저닝은 opt-in입니다.
- **도구 수**: `mcp_server.go` 안 `add(...)` 호출 **30회**(그중 28개는 이름 리터럴,
  2개는 상수 경유 — `claudeAuditToolName`과 `auditMultiToolName`). 카탈로그
  `internal/mcp/catalog.go`도 **30개**를 선언하며 두 수가 일치합니다.
- **도구 목록**: `session_list`, `goal_status`, `goal_arm`, `spec_progress`, `verify_snapshot`,
  `verify_trend`, `spec_audit`, `spec_drift`, `audit_cache`,
  `codex_{audit,setup,task,job_status,job_result,job_cancel}`, `claude_audit`,
  `glm_{task,job_status,job_result,job_cancel,audit}`, `audit_multi`,
  `session_msg_{register,list,send,poll}`,
  `graph_{file_api,find_code,trace_calls,shortest_path}`.
- **가드**: `mcp.yaml`에서 도구별 활성화를 읽고, `add()` 헬퍼의 첫 인자가 `mcp.NewTool` 이름 및
  카탈로그와 일치해야 한다는 계약을 가드 테스트가 강제합니다
  (`mcp_annotation_guard_test.go`, `mcp_boundary_test.go`).
- **`claude_audit`은 이 판에서 더해진 도구**입니다(`internal/cli/mcp_claude.go`). 읽기 전용 코드
  리뷰를 `claude` CLI 서브프로세스로 수행하고, 자식 환경에서 `CLAUDE_CODE_*`·`CLAUDECODE`를
  지운 뒤 출력 크기에 상한을 둡니다. `audit_multi` 수렴도 같은 수행 함수를 Claude 백엔드로 씁니다.
- **`codex_audit`과 `audit_multi`는 호출될 때마다 영수증을 남깁니다**(이 판에서 더해진
  `internal/cli/mcp_audit_receipt.go`). 도구 호출 1건이 `.moai/state/audit-receipts/receipts/`
  아래 JSON 파일 1개이며, 같은 기록을 훅 쪽 가드가 읽습니다(§ 훅 → 감사 영수증 가드). 이
  배선이 MCP 표면을 **훅의 판정 근거를 생산하는 자리**로 만듭니다 — 도구 목록만 읽어서는
  보이지 않는 결합입니다.
- **엔트리 관리와 서버 실행은 별개 서브트리**입니다 — `newMCPCmd()`(`mcp.go`)가 `.mcp.json`의
  add/remove/list를, `mcp-server`가 실행을 담당합니다.

---

## 웹 콘솔 표면

`moai web`이 띄우는 루프백 콘솔은 탭 단위 표면이며, 탭 하나는 **편집하지 않습니다**.
codex 탭(행 모델은 `internal/web/codexmirror.go`, 렌더는 그 짝 `.templ` 소스에서 생성된
패널)은 Audit·MCP 탭에 사는 codex 설정의 읽기 전용 미러입니다. 이 패널은 `name` 속성을 가진 폼 요소를 하나도 내지 않으며, 그 금지는 숨은 bool
동반자 `<name>__present`까지 덮습니다 — 모든 패널이 한 폼 안에 살고 탭 전환은 표시 전환일
뿐이라 **비활성 패널도 함께 제출되기** 때문입니다.

---

## CI가 소비하는 종료 코드 표면

`moai graph check`는 사람보다 기계가 먼저 읽는 진입점입니다. `.github/workflows/graph-freshness.yml`의
`graph-freshness` 잡과 `moai gate`가 **종료 코드만** 소비하므로, 보고만 하고 항상 0으로 끝나는
구현은 두 소비자를 조용히 무장해제시킵니다.

| 종료 코드 | 의미 | 대표 원인 |
|---|---|---|
| `0` | 모든 층 fresh | — |
| `1` | 한 층 이상 stale 또는 absent | described-source-diff ≥ 40, codemaps 본문 부재(C1), 인용된 경로 부재 |
| `2` | system error — 측정 자체가 성립하지 않음 | 스탬프 커밋 미해석, 스탬프가 HEAD의 조상이 아님, `gate.yaml` 파싱 실패 |

exit 2 경로는 층 표를 렌더하지 않습니다. 숫자 행이 없는 것이 계약입니다 — 성립한 적 없는
비교 창에 대해 값을 내놓지 않기 위해서입니다. 도달 불가 스탬프일 때만
`internal/cli/graph_check.go`가 복구 안내(본문 재생성 → 도달 가능한 커밋으로 재스탬핑)를 덧붙이고,
해석 불가 스탬프에는 붙이지 않습니다. 둘의 해법이 다르기 때문입니다. 판정 순서는
§ `data-flow.md` I를 참조하십시오.

워크플로 쪽 대상 선택은 이벤트별로 셋입니다: `push`는 `HEAD`, 일반 `pull_request`는
`origin/<base_ref>`, `release/*` head의 `pull_request`는 merge preview `HEAD`입니다.
