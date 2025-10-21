---

name: moai-alfred-refactoring-coach
description: Refactoring guidance with design patterns, code smells detection, and step-by-step improvement plans. Use when outlining refactor steps that preserve functionality.
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
---

# Alfred Refactoring Coach

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | /alfred:2-run refactor lane |
| Trigger cues | Refactoring retros, duplication cleanup, code health follow-ups inside Alfred. |

## What it does

Refactoring guidance with design pattern recommendations, code smell detection, and step-by-step improvement plans.

## When to use

- Activates when Alfred is asked to plan or stage refactoring work.
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
- Fowler, Martin. "Refactoring: Improving the Design of Existing Code." Addison-Wesley, 2018.
- IEEE Software. "Managing Technical Debt." IEEE Software, 2021.

## Changelog
- 2025-03-29: Alfred 전용 스킬에 입력/출력/실패 대응을 추가했습니다.

## Works well with

- alfred-code-reviewer
- alfred-trust-validation
