---

name: moai-alfred-debugger-pro
description: Advanced debugging support with stack trace analysis, error pattern detection, and fix suggestions. Use when unraveling complex runtime errors or stack traces.
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
---

# Alfred Debugger Pro

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | Triggered by Alfred debug-helper |
| Trigger cues | Runtime failures surfaced in Alfred runs, stack trace walkthroughs, hotfix triage. |

## What it does

Advanced debugging support with stack trace analysis, common error pattern detection, and actionable fix suggestions.

## When to use

- Activates when Alfred encounters runtime errors and needs guided debugging steps.
- “Resolve the error”, “What is the cause of this error?”, “Stack trace analysis”
- Automatically invoked on runtime errors (via debug-helper sub-agent)
- "Why not?", "Solving NullPointerException"

## How it works

**Stack Trace Analysis**:
```python
# Error example
jwt.exceptions.ExpiredSignatureError: Signature has expired

# Alfred Analysis
📍 Error Location: src/auth/service.py:142
🔍 Root Cause: JWT token has expired
💡 Fix Suggestion:
   1. Implement token refresh logic
   2. Check expiration before validation
   3. Handle ExpiredSignatureError gracefully
```

**Common Error Patterns**:
- `NullPointerException` → Optional usage, guard clauses
- `IndexError` → Boundary checks
- `KeyError` → `.get()` with defaults
- `TypeError` → Type hints, input validation
- `ConnectionError` → Retry logic, timeouts

**Debugging Checklist**:
- [ ] Reproducible?
- [ ] Log messages?
- [ ] Input data?
- [ ] Recent changes?
- [ ] Dependency versions?

**Language-specific Tips**:
- **Python**: Logging, type guards
- **TypeScript**: Type guards, null checks
- **Java**: Optional, try-with-resources

## Best Practices
- 사용자에게 보여주는 문구는 TUI/보고서용 표현으로 작성합니다.
- 도구 실행 시 명령과 결과 요약을 함께 기록합니다.

## Examples
```markdown
- /alfred 커맨드 내부에서 이 스킬을 호출해 보고서를 생성합니다.
- Completion Report에 요약을 추가합니다.
```

## Inputs
- MoAI-ADK 프로젝트 맥락 (`.moai/project/`, `.claude/` 템플릿 등).
- 사용자 명령 또는 상위 커맨드에서 전달한 파라미터.

## Outputs
- Alfred 워크플로우에 필요한 보고서, 체크리스트 또는 추천 항목.
- 후속 서브 에이전트 호출을 위한 구조화된 데이터.

## Failure Modes
- 필수 입력 문서가 없거나 권한이 제한된 경우.
- 사용자 승인 없이 파괴적인 변경이 요구될 때.

## Dependencies
- cc-manager, project-manager 등 상위 에이전트와 협력이 필요합니다.

## References
- Microsoft. "Debugging Techniques." https://learn.microsoft.com/visualstudio/debugger/ (accessed 2025-03-29).
- JetBrains. "Debugging Code." https://www.jetbrains.com/help/idea/debugging-code.html (accessed 2025-03-29).

## Changelog
- 2025-03-29: Alfred 전용 스킬에 입력/출력/실패 대응을 추가했습니다.

## Works well with

- alfred-code-reviewer
- alfred-trust-validation
