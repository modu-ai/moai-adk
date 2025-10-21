---

name: moai-alfred-trust-validation
description: Validates TRUST 5-principles compliance (Test coverage 85%+, Code constraints, Architecture unity, Security, TAG trackability). Use when enforcing TRUST checkpoints before progression.
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
---

# Alfred TRUST Validation

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | /alfred:3-sync quality gate |
| Trigger cues | TRUST checklist enforcement, release readiness scoring, risk gating. |

## What it does

Validates MoAI-ADK's TRUST 5-principles compliance to ensure code quality, testability, security, and traceability.

## When to use

- Activates when Alfred evaluates TRUST compliance before handoff.
- "Check the TRUST principle", "Quality verification", "Check code quality"
- Automatically invoked by `/alfred:3-sync`
- Before merging PR or releasing

## How it works

**T - Test First**:
- Checks test coverage ≥85% (pytest, vitest, go test, cargo test, etc.)
- Verifies TDD cycle compliance (RED → GREEN → REFACTOR)

**R - Readable**:
- File ≤300 LOC
- Function ≤50 LOC
- Parameters ≤5
- Cyclomatic complexity ≤10

**U - Unified**:
- SPEC-driven architecture consistency
- Clear module boundaries
- Language-specific standard structures

**S - Secured**:
- Input validation implementation
- No hardcoded secrets
- Access control applied

**T - Trackable**:
- TAG chain integrity (@SPEC → @TEST → @CODE → @DOC)
- No orphaned TAGs
- No duplicate SPEC IDs

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
- SonarSource. "Quality Gate: Developer's Guide." https://www.sonarsource.com/company/newsroom/white-papers/quality-gate/ (accessed 2025-03-29).
- ISO/IEC 25010. "Systems and software quality models." (accessed 2025-03-29).

## Changelog
- 2025-03-29: Alfred 전용 스킬에 입력/출력/실패 대응을 추가했습니다.

## Works well with

- alfred-tag-scanning (TAG traceability)
- alfred-code-reviewer (code quality analysis)

## Files included

- templates/trust-report-template.md
