---

name: moai-foundation-tags
description: Scans @TAG markers directly from code and generates inventory (CODE-FIRST). Use when establishing or auditing TAG conventions.
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
---

# Alfred TAG Scanning

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | SessionStart (foundation bootstrap) |
| Trigger cues | TAG scanning, traceability audits, orphan TAG remediation requests. |

## What it does

Scans all @TAG markers (SPEC/TEST/CODE/DOC) directly from codebase and generates TAG inventory without intermediate caching (CODE-FIRST principle).

## When to use

- Activates when scanning or auditing TAG chains or locating orphaned tags.
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
- BurntSushi. "ripgrep User Guide." https://github.com/BurntSushi/ripgrep/blob/master/GUIDE.md (accessed 2025-03-29).
- ReqView. "Requirements Traceability Matrix Guide." https://www.reqview.com/doc/requirements-traceability-matrix/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Foundation 스킬 템플릿을 베스트 프랙티스 구조에 맞게 보강했습니다.

## Works well with

- moai-foundation-trust
- moai-foundation-specs

## Best Practices
- 표준 변경 시 변경 사유와 근거 문서를 함께 기록합니다.
- 단일 소스 원칙을 지켜 동일 항목을 여러 곳에서 수정하지 않도록 합니다.
