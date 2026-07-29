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

## CLI 명령 (~40 root verbs, 152 non-test AddCommand 호출)

**프로젝트**: init, update, doctor, config, version, web  
**SPEC**: plan, run, sync, spec (audit/lint/close)  
**개발**: loop, clean, mx, fix, goal  
**인프라**: hook, migration, worktree, session  
**다중LLM**: cc, glm, cg  
**기타**: research, constitution, design, project, codemaps, feedback, review, coverage, e2e

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
