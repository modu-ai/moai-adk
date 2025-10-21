---

name: moai-alfred-code-reviewer
description: Automated code review with language-specific best practices, SOLID principles, and actionable improvement suggestions. Use when reviewing code changes for issues and strengths.
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
---

# Alfred Code Reviewer

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | /alfred:3-sync review phase |
| Trigger cues | Alfred-driven review summaries, diff inspection, merge gate decisions. |

## What it does

Automated code review with language-specific best practices, SOLID principles verification, and code smell detection.

## When to use

- Activates when Alfred needs to summarize diffs or prepare merge feedback.
- "코드 리뷰해줘", "이 코드 개선점은?", "코드 품질 확인"
- Optionally invoked after `/alfred:3-sync`
- Before merging PR

## How it works

**Code Constraints Check**:
- File ≤300 LOC
- Function ≤50 LOC
- Parameters ≤5
- Cyclomatic complexity ≤10

**SOLID Principles**:
- Single Responsibility
- Open/Closed
- Liskov Substitution
- Interface Segregation
- Dependency Inversion

**Code Smell Detection**:
- Long Method
- Large Class
- Duplicate Code
- Dead Code
- Magic Numbers

**Language-specific Best Practices**:
- Python: List comprehension, type hints, PEP 8
- TypeScript: Strict typing, async/await, error handling
- Java: Streams API, Optional, Design patterns

**Review Report**:
```markdown
## Code Review Report

### 🔴 Critical Issues (3)
1. **src/auth/service.py:45** - Function too long (85 > 50 LOC)
2. **src/api/handler.ts:120** - Missing error handling
3. **src/db/repository.java:200** - Magic number

### ⚠️ Warnings (5)
1. **src/utils/helper.py:30** - Unused import

### ✅ Good Practices Found
- Test coverage: 92%
- Consistent naming
```

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
- Google. "Code Review Developer Guide." https://google.github.io/eng-practices/review/ (accessed 2025-03-29).
- IEEE. "Code Review Best Practices." IEEE Software, 2022.

## Changelog
- 2025-03-29: Alfred 전용 스킬에 입력/출력/실패 대응을 추가했습니다.

## Works well with

- alfred-trust-validation
- alfred-refactoring-coach
