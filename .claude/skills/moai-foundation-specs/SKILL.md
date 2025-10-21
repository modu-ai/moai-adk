---

name: moai-foundation-specs
description: Validates SPEC YAML frontmatter (7 required fields) and HISTORY section. Use when enforcing SPEC documentation standards.
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
---

# Alfred SPEC Metadata Validation

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | SessionStart (foundation bootstrap) |
| Trigger cues | SPEC metadata validation, frontmatter completeness, specification readiness checks. |

## What it does

Validates SPEC document structure including YAML frontmatter (7 required fields) and HISTORY section compliance.

## When to use

- Activates when verifying SPEC frontmatter or preparing new specification templates.
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

## Examples
```markdown
- 표준 문서를 스캔하여 누락 섹션을 보고합니다.
- 변경된 규약을 CLAUDE.md에 반영합니다.
```

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
- INCOSE. "Guide for Writing Requirements." INCOSE-TP-2010-006-02 (accessed 2025-03-29).
- IEEE. "Software Requirements Specification Standard." IEEE 830-1998.

## Changelog
- 2025-03-29: Foundation 스킬 템플릿을 베스트 프랙티스 구조에 맞게 보강했습니다.

## Works well with

- moai-foundation-ears
- moai-foundation-tags

## Best Practices
- 표준 변경 시 변경 사유와 근거 문서를 함께 기록합니다.
- 단일 소스 원칙을 지켜 동일 항목을 여러 곳에서 수정하지 않도록 합니다.
