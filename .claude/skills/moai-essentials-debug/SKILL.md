---

name: moai-essentials-debug
description: Advanced debugging with stack trace analysis, error pattern detection, and fix suggestions. Use when delivering quick diagnostic support for everyday issues.
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
---

# Alfred Debugger Pro

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | On demand during Run stage (debug-helper) |
| Trigger cues | Runtime error triage, stack trace analysis, root cause investigation requests. |

## What it does

Advanced debugging support with stack trace analysis, common error pattern detection, and actionable fix suggestions.

## When to use

- Loads when users share stack traces or ask why a failure occurred.
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

## Examples
```markdown
- 현재 diff를 점검하고 즉시 수정 가능한 항목을 나열합니다.
- 후속 작업은 TodoWrite로 예약합니다.
```

## Inputs
- 현재 작업 중인 코드/테스트/문서 스냅샷.
- 진행 중인 에이전트 상태 정보.

## Outputs
- 즉시 실행 가능한 체크리스트 또는 개선 제안.
- 다음 단계 실행 여부에 대한 권장 사항.

## Failure Modes
- 필요한 파일이나 테스트 결과를 찾지 못한 경우.
- 작업 범위가 과도하게 넓어 간단한 지원만으로 해결할 수 없을 때.

## Dependencies
- 주로 `tdd-implementer`, `quality-gate` 등과 연계해 사용합니다.

## References
- Microsoft. "Debugging Techniques." https://learn.microsoft.com/visualstudio/debugger/ (accessed 2025-03-29).
- JetBrains. "Debugging Code." https://www.jetbrains.com/help/idea/debugging-code.html (accessed 2025-03-29).

## Changelog
- 2025-03-29: Essentials 스킬의 입력/출력 정의를 정비했습니다.

## Works well with

- moai-essentials-refactor

## Best Practices
- 간단한 개선이라도 결과를 기록해 추적 가능성을 높입니다.
- 사람 검토가 필요한 항목을 명확히 표시하여 자동화와 구분합니다.
