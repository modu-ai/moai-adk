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

## CLI 명령 (root 등록 61개, non-test AddCommand 호출 201개 — 2026-08-31 재검증, 값 불변)

**프로젝트**: init, update, doctor, config, version, web  
**SPEC**: plan, run, sync, spec (audit/lint/close)  
**개발**: loop, clean, mx, fix, goal  
**인프라**: hook, migration, worktree, session  
**다중LLM**: cc, glm, cg, codex  
**기타**: research, constitution, design, project, codemaps, feedback, review, coverage, e2e

---

## `moai codex` — Codex 런처 (런처 그룹, cc/glm/cg 형제)

```bash
moai codex            # 준비 상태 리드아웃 6행 (기동 없음, rc 0)
moai codex status     # 동일
moai codex cli [--spawn] [-- <codex-args...>]   # Codex CLI를 프로젝트 루트에서 기동
moai codex app [--spawn]                        # Codex 데스크톱 앱 (codex app) 기동
```

- 동사는 폐쇄 집합 {bare, status} × {cli, app} — 모르는 토큰은 기동으로 라우팅되지 않고 1행 사용법 진단 후 rc 1.
- 두 기동 동사는 기동 직전 **init-offer 게이트**(`codexInitOfferGate`, `internal/cli/codex_init.go`)를 통과한다 — 배선 불완전 시 상태·처방을 보고하고, 대화형 세션에서는 배선 생성을 제안한다(수락 시 `codexwiring` 생성기 1회 호출 후 AGENTS.md ↔ CLAUDE.md 지시 계약 확보). 게이트는 `--spawn` 인자를 받지 않아 직접/spawn 양쪽 기동 경로가 같은 함수를 지난다.
- 리드아웃 6행: codex 바이너리, CODEX_HOME, auth 제공자, 프로젝트 배선, 생성된 agent TOML, 하네스 진입.

---

## 훅 진입점 (30 이벤트 · 35개 handle-*.sh 스크립트)

```bash
moai hook <event>  # JSON stdin → Handler dispatch → exit 0/2
```

**주요 이벤트**:
- SessionStart, PostToolUse, Stop, SubagentStop, TaskCompleted
- PreCompact, PostCompact, WorktreeCreate, WorktreeRemove
- UserPromptSubmit, Notification, TeammateIdle, TaskCompleted, ... (총 30개 EventType — `internal/hook/types.go`)

---

## HTTP 서버

```bash
moai web  # 127.0.0.1:3041 (loopback only, 기본 포트)
```

---

**생성**: `/moai codemaps --force`로 자동 생성
