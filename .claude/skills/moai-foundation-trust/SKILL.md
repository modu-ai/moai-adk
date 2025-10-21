---

name: moai-foundation-trust
description: Validates TRUST 5-principles (Test 85%+, Readable, Unified, Secured, Trackable). Use when aligning with TRUST governance.
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
---

# Foundation: TRUST Validation

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | SessionStart (foundation bootstrap) |
| Trigger cues | TRUST compliance checks, release readiness reviews, quality gate enforcement. |

## What it does

Validates MoAI-ADK's TRUST 5-principles compliance to ensure code quality, testability, security, and traceability.

## When to use

- Activates when TRUST compliance or release readiness needs to be evaluated.
- "Check the TRUST principle", "Quality verification", "Check code quality"
- Automatically invoked by `/alfred:3-sync`
- Before merging PR or releasing

## How it works

**T - Test First**:
- Checks test coverage ≥85% (pytest, vitest, go test, cargo test, etc.)
- Verifies TDD cycle compliance (RED → GREEN → REFACTOR)

**R - Readable**:
- File ≤300 LOC, Function ≤50 LOC, Parameters ≤5, Complexity ≤10

**U - Unified**:
- SPEC-driven architecture consistency, Clear module boundaries

**S - Secured**:
- Input validation, No hardcoded secrets, Access control

**T - Trackable**:
- TAG chain integrity (@SPEC → @TEST → @CODE → @DOC)

## Inputs
- 프로젝트 표준 문서(예: `CLAUDE.md`, `.moai/config.json`).
- 관련 서브 에이전트의 최신 출력물.

## Outputs
- MoAI-ADK 표준에 맞는 템플릿 또는 정책 요약.
- 재사용 가능한 규칙/체크리스트.

## Failure Modes
- 필수 표준 파일이 없거나 접근 권한이 제한된 경우.
- 상충하는 정책이 감지되어 조정이 필요할 때.

## Dependencies
- cc-manager와 함께 호출될 때 시너지가 큽니다.

## References
- SonarSource. "Quality Gate: Developer's Guide." https://www.sonarsource.com/company/newsroom/white-papers/quality-gate/ (accessed 2025-03-29).
- ISO/IEC 25010. "Systems and software quality models." (accessed 2025-03-29).

## Changelog
- 2025-03-29: Foundation 스킬 템플릿을 베스트 프랙티스 구조에 맞게 보강했습니다.

## Works well with

- moai-foundation-tags (TAG traceability)
- moai-foundation-specs (SPEC validation)

## Examples
```markdown
- 표준 문서를 스캔하여 누락 섹션을 보고합니다.
- 변경된 규약을 CLAUDE.md에 반영합니다.
```

## Best Practices
- 표준 변경 시 변경 사유와 근거 문서를 함께 기록합니다.
- 단일 소스 원칙을 지켜 동일 항목을 여러 곳에서 수정하지 않도록 합니다.
