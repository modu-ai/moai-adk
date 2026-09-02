# 진입점 및 명령 참고

> 이 문서는 `/moai codemaps --force`로 자동 생성된 진입점 목록입니다.

**모듈**: `github.com/modu-ai/moai-adk`  
**Go 버전**: go 1.26.4

---

## 바이너리 진입점

### cmd/moai
```go
main() → cli.Execute() → cobra rootCmd.Execute()
```

---

## Composition Root

### internal/cli/deps.go
```go
InitDependencies() // 모든 서브시스템 와이어링
```

---

## CLI 명령 (root 등록 60건, non-test AddCommand 호출 202개 — 2026-09-02 실측)

`moai --help` 노출 기준 명령 그룹 (help 미노출: hidden `statusline`, cobra 자동 `help`/`completion`):

**Project**: init, status, doctor, update, migrate, pr  
**Launchers**: cc, glm, cg, codex  
**Autonomous/Dev**: loop, spec (audit/lint/close), plan, goal, gate  
**Governance**: constitution, mx, telemetry  
**Tools/Infra**: hook, session, worktree, migration (run/status/rollback), integration, graph, chain, handoff, verify, todo, epic, memory, model, tokens, clean, inventory, ast-grep, ast-edit, mcp, mcp-server, config, tool-policy, preference, github, github 워크플로우, lsp, research, agent, workflow, web, version

> 참고: 터미널 CLI에는 독립 `moai run`/`moai sync` root 명령이 **없다** — plan/run/sync 워크플로우는 `/moai` Claude Code 스킬(`.claude/commands/moai/`)로 제공되며, `run` 동사는 `moai migration`의 서브커맨드다. `moai sync`와 유사한 이름의 `navigator-sync`는 별개 명령이다.

---

## `moai codex` — Codex 런처 (런처 그룹, cc/glm/cg 형제)

```bash
moai codex [--spawn] [-w <worktree>] [-- <codex-args...>]   # Codex CLI를 프로젝트 루트에서 기동
moai codex cli [--spawn] [-w <worktree>] [-- <codex-args...>]   # 같은 기동, 명시 별칭
moai codex status     # 준비 상태 리드아웃 6행 (기동 없음, rc 0)
moai codex app [--spawn]                        # Codex 데스크톱 앱 (codex app) 기동
```

- 동사는 폐쇄 집합 {bare, cli, app} 기동 × {status} 리드아웃 — 모르는 토큰은 기동으로 라우팅되지 않고 1행 사용법 진단 후 rc 1.
- 라우팅 하류의 argv 번역 단계가 **실재하는 codex 서브커맨드만** 자식에게 넘긴다 — `app` 은 전달되고, moai 쪽 동사인 맨몸·`cli` 는 전달되지 않아 자식은 운영자가 친 tail 만 받는다. `-w` 는 moai 가 소화해 자식의 작업 디렉터리를 **기존** 워크트리로 잡으며(생성하지 않는다) 자식 argv 에 남지 않는다.
- 세 기동 형태는 기동 직전 **init-offer 게이트**(`codexInitOfferGate`, `internal/cli/codex_init.go`)를 통과한다 — 배선 불완전 시 상태·처방을 보고하고, 대화형 세션에서는 배선 생성을 제안한다(수락 시 `codexwiring` 생성기 1회 호출 후 AGENTS.md ↔ CLAUDE.md 지시 계약 확보). 게이트는 `--spawn` 인자를 받지 않아 직접/spawn 양쪽 기동 경로가 같은 함수를 지난다.
- 리드아웃 6행: codex 바이너리, CODEX_HOME, auth 제공자, 프로젝트 배선, 생성된 agent TOML, 하네스 진입.

---

## 훅 진입점 (30 이벤트 · 39개 고유 handle-* 래퍼 — 템플릿 기준 43개 파일, .sh/.sh.tmpl 쌍 포함)

```bash
moai hook <event>  # JSON stdin → Handler dispatch → exit 0/2
```

**주요 이벤트**:
- SessionStart, PostToolUse, Stop, SubagentStop, TaskCompleted
- PreCompact, PostCompact, WorktreeCreate, WorktreeRemove
- UserPromptSubmit, Notification, TeammateIdle, TaskCreated, ... (총 30개 EventType — `internal/hook/types.go`)

---

## HTTP 서버

```bash
moai web  # 127.0.0.1:3041 (loopback only, 기본 포트)
```

---

**생성**: `/moai codemaps --force`로 자동 생성
