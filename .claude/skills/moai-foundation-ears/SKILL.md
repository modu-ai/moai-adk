---

name: moai-foundation-ears
description: EARS requirement authoring guide (Ubiquitous/Event/State/Optional/Constraints). Use when teams need guidance on EARS requirements structure.
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
---

# Alfred EARS Authoring Guide

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | SessionStart (foundation bootstrap) |
| Trigger cues | Requests to draft or refine EARS-style requirements, “write spec”, or “requirements format” cues. |

## What it does

EARS (Easy Approach to Requirements Syntax) authoring guide for writing clear, testable requirements using 5 statement patterns.

## When to use

- Activates whenever the user asks to draft structured requirements or mentions EARS syntax.
- “Writing SPEC”, “Requirements summary”, “EARS syntax”
- Automatically invoked by `/alfred:1-plan`
- When writing or refining SPEC documents

## How it works

EARS provides 5 statement patterns for structured requirements:

### 1. Ubiquitous (Basic Requirements)
**Format**: The system must provide [function]
**Example**: The system must provide user authentication function

### 2. Event-driven (event-based)
**Format**: WHEN If [condition], the system must [operate]
**Example**: WHEN When the user logs in, the system must issue a JWT token

### 3. State-driven
**Format**: WHILE When in [state], the system must [operate]
**Example**: WHILE When the user is authenticated, the system must allow access to protected resources

### 4. Optional (Optional function)
**Format**: If WHERE [condition], the system can [operate]
**Example**: If WHERE refresh token is provided, the system can issue a new access token

### 5. Constraints
**Format**: IF [condition], the system SHOULD [constrain]
**Example**: IF an invalid token is provided, the system SHOULD deny access

## Writing Tips

✅ Be specific and measurable
✅ Avoid vague terms (“adequate”, “sufficient”, “fast”)
✅ One requirement per statement
✅ Make it testable

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
- Mavin, A., et al. "Easy Approach to Requirements Syntax (EARS)." IEEE RE, 2009.
- INCOSE. "Guide for Writing Requirements." INCOSE-TP-2010-006-02 (accessed 2025-03-29).

## Changelog
- 2025-03-29: Foundation 스킬 템플릿을 베스트 프랙티스 구조에 맞게 보강했습니다.

## Works well with

- moai-foundation-specs

## Best Practices
- 표준 변경 시 변경 사유와 근거 문서를 함께 기록합니다.
- 단일 소스 원칙을 지켜 동일 항목을 여러 곳에서 수정하지 않도록 합니다.
