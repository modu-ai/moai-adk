---

name: moai-essentials-refactor
description: Refactoring guidance with design patterns and code improvement strategies. Use when planning incremental refactors with safety nets.
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
---

# Alfred Refactoring Coach

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | On demand during Run stage (refactor planning) |
| Trigger cues | Refactoring plans, code smell cleanup, design pattern coaching. |

## What it does

Refactoring guidance with design pattern recommendations, code smell detection, and step-by-step improvement plans.

## When to use

- Loads when the user asks how to restructure code or apply design patterns.
- “Help with refactoring”, “How can I improve this code?”, “Apply design patterns” 
- “Code organization”, “Remove duplication”, “Separate functions”

## How it works

**Refactoring Techniques**:
- **Extract Method**: Separate long methods
- **Replace Conditional with Polymorphism**: Remove conditional statements
- **Introduce Parameter Object**: Group parameters
- **Extract Class**: Massive class separation

**Design Pattern Recommendations**:
- Complex object creation → **Builder Pattern**
- Type-specific behavior → **Strategy Pattern**
- Global state → **Singleton Pattern**
- Incompatible interfaces → **Adapter Pattern**
- Delayed object creation → **Factory Pattern**

**3-Strike Rule**:
```
1st occurrence: Just implement
2nd occurrence: Notice similarity (leave as-is)
3rd occurrence: Pattern confirmed → Refactor! 🔧
```

**Refactoring Checklist**:
- [ ] All tests passing before refactoring
- [ ] Code smells identified
- [ ] Refactoring goal clear
- [ ] Change one thing at a time
- [ ] Run tests after each change
- [ ] Commit frequently

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
- Fowler, Martin. "Refactoring: Improving the Design of Existing Code." Addison-Wesley, 2018.
- IEEE Software. "Managing Technical Debt." IEEE Software, 2021.

## Changelog
- 2025-03-29: Essentials 스킬의 입력/출력 정의를 정비했습니다.

## Works well with

- moai-essentials-review

## Best Practices
- 간단한 개선이라도 결과를 기록해 추적 가능성을 높입니다.
- 사람 검토가 필요한 항목을 명확히 표시하여 자동화와 구분합니다.
