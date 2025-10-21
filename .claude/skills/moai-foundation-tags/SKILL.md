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
- Scan standard documents and report missing sections.
- Changed regulations are reflected in CLAUDE.md.
```

## Inputs
- Project standard documents (e.g. `CLAUDE.md`, `.moai/config.json`).
- Latest printouts from relevant sub-agents.

## Outputs
- Templates or policy summaries conforming to the MoAI-ADK standard.
- Reusable rules/checklists.

## Failure Modes
- When required standard files are missing or have limited access rights.
- When conflicting policies are detected and coordination is required.

## Dependencies
- There is great synergy when called together with cc-manager.

## References
- BurntSushi. "ripgrep User Guide." https://github.com/BurntSushi/ripgrep/blob/master/GUIDE.md (accessed 2025-03-29).
- ReqView. "Requirements Traceability Matrix Guide." https://www.reqview.com/doc/requirements-traceability-matrix/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Foundation skill templates have been enhanced to align with best practice structures.

## Works well with

- moai-foundation-trust
- moai-foundation-specs

## Best Practices
- When changing standards, the reason for change and supporting documents are recorded.
- Follow the single source principle and avoid modifying the same item in multiple places.
