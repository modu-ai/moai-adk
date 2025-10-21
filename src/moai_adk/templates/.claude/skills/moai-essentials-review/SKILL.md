---

name: moai-essentials-review
description: Automated code review with SOLID principles, code smells, and language-specific best practices. Use when preparing concise review checklists for code changes.
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
---

# Alfred Code Reviewer

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | On demand during Sync stage (review gate) |
| Trigger cues | Code review requests, quality checklist preparation, merge readiness checks. |

## What it does

Automated code review with language-specific best practices, SOLID principles verification, and code smell detection.

## When to use

- Loads when someone asks for a code review or a pre-merge quality assessment.
- “Please review the code”, “How can this code be improved?”, “Check the code quality”
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
- IEEE. "Recommended Practice for Software Reviews." IEEE 1028-2008.
- Cisco. "Peer Review Best Practices." https://www.cisco.com/c/en/us/support/docs/optical/ons-15454-esc/15114-peer-review.html (accessed 2025-03-29).

## Changelog
- 2025-03-29: Essentials 스킬의 입력/출력 정의를 정비했습니다.

## Works well with

- moai-foundation-specs
- moai-essentials-refactor

## Best Practices
- 간단한 개선이라도 결과를 기록해 추적 가능성을 높입니다.
- 사람 검토가 필요한 항목을 명확히 표시하여 자동화와 구분합니다.
