---

name: moai-alfred-tag-scanning
description: Scans all @TAG markers directly from code and generates TAG inventory (CODE-FIRST principle - no intermediate cache). Use when rebuilding the TAG inventory from live code.
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
---

# Alfred TAG Scanning

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | /alfred:3-sync traceability gate |
| Trigger cues | Traceability scans, orphan TAG cleanup, TAG chain verification in Alfred. |

## What it does

Scans all @TAG markers (SPEC/TEST/CODE/DOC) directly from codebase and generates TAG inventory without intermediate caching (CODE-FIRST principle).

## When to use

- Activates when Alfred needs TAG inventories or chain verification.
- "TAG Scan", "TAG List", "TAG Inventory"
- Automatically invoked by `/alfred:3-sync`
- “Find orphan TAG”, “Check TAG chain”

## How it works

**CODE-FIRST Scanning**:
```bash
# Direct code scan without intermediate cache
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/
```

**TAG Inventory Generation**:
- Lists all TAGs with file locations
- Detects orphaned TAGs (no corresponding SPEC/TEST/CODE)
- Identifies broken links in TAG chain
- Reports duplicate IDs

**TAG Chain Verification**:
- @SPEC → @TEST → @CODE → @DOC connection check
- Ensures traceability across all artifacts

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
- BurntSushi. "ripgrep User Guide." https://github.com/BurntSushi/ripgrep/blob/master/GUIDE.md (accessed 2025-03-29).
- ReqView. "Requirements Traceability Matrix Guide." https://www.reqview.com/doc/requirements-traceability-matrix/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Alfred 전용 스킬에 입력/출력/실패 대응을 추가했습니다.

## Works well with

- alfred-trust-validation (TAG traceability verification)
- alfred-spec-metadata-validation (SPEC ID validation)

## Files included

- templates/tag-inventory-template.md
