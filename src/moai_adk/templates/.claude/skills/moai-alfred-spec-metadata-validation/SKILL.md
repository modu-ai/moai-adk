---

name: moai-alfred-spec-metadata-validation
description: Validates SPEC YAML frontmatter (7 required fields) and HISTORY section compliance. Use when validating SPEC metadata for consistency.
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
---

# Alfred SPEC Metadata Validation

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | /alfred:1-plan spec validation |
| Trigger cues | SPEC frontmatter checks, history table enforcement, metadata guardrails. |

## What it does

Validates SPEC document structure including YAML frontmatter (7 required fields) and HISTORY section compliance.

## When to use

- Activates when Alfred validates SPEC templates or enforces metadata standards.
- "SPEC verification", "Metadata check", "SPEC structure check"
- Automatically invoked by `/alfred:1-plan`
- Before creating SPEC document

## How it works

**YAML Frontmatter Validation (7 required fields)**:
- `id`: SPEC ID (e.g., AUTH-001)
- `version`: Semantic Version (e.g., 0.0.1)
- `status`: draft|active|completed|deprecated
- `created`: YYYY-MM-DD format
- `updated`: YYYY-MM-DD format
- `author`: @{GitHub ID} format
- `priority`: low|medium|high|critical

**HISTORY Section Validation**:
- Checks existence of HISTORY section
- Verifies version history (INITIAL/ADDED/CHANGED/FIXED tags)
- Validates author and date consistency

**Format Validation**:
```bash
# Check required fields
rg "^(id|version|status|created|updated|author|priority):" .moai/specs/SPEC-*/spec.md

# Verify HISTORY section
rg "^## HISTORY" .moai/specs/SPEC-*/spec.md
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
- IEEE. "Software Requirements Specification Standard." IEEE 830-1998.
- NASA. "Systems Engineering Handbook." https://www.nasa.gov/seh/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Alfred 전용 스킬에 입력/출력/실패 대응을 추가했습니다.

## Works well with

- alfred-ears-authoring (SPEC writing guide)
- alfred-tag-scanning (SPEC ID duplication check)

## Reference

SSOT (Single Source of Truth): `.moai/memory/spec-metadata.md`
