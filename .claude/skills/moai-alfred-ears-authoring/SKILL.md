---

name: moai-alfred-ears-authoring
description: EARS (Easy Approach to Requirements Syntax) authoring guide with 5 statement patterns for clear, testable requirements. Use when generating EARS-style requirement sentences.
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
---

# Alfred EARS Authoring Guide

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | /alfred:1-plan requirements phase |
| Trigger cues | Plan board EARS drafting, requirement interviews, structured SPEC authoring. |

## What it does

EARS (Easy Approach to Requirements Syntax) authoring guide for writing clear, testable requirements using 5 statement patterns.

## When to use

- Activates when Alfred is asked to capture requirements using the EARS patterns.
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
- Mavin, A., et al. "Easy Approach to Requirements Syntax (EARS)." IEEE RE, 2009.
- INCOSE. "Guide for Writing Requirements." INCOSE-TP-2010-006-02 (accessed 2025-03-29).

## Changelog
- 2025-03-29: Alfred 전용 스킬에 입력/출력/실패 대응을 추가했습니다.

## Works well with

- alfred-spec-metadata-validation
- alfred-trust-validation

## Reference

`.moai/memory/development-guide.md#ears-requirements-how-to`
