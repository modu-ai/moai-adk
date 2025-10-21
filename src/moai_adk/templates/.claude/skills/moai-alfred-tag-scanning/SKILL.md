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
- The text shown to the user is written using TUI/report expressions.
- When running the tool, a summary of commands and results are recorded.

## Examples
```markdown
- Call this skill inside the /alfred command to generate a report.
- Add summary to Completion Report.
```

## Inputs
- MoAI-ADK project context (`.moai/project/`, `.claude/` templates, etc.).
- Parameters passed from user commands or higher commands.

## Outputs
- Reports, checklists or recommendations for your Alfred workflow.
- Structured data for subsequent subagent calls.

## Failure Modes
- When required input documents are missing or permissions are limited.
- When disruptive changes are required without user approval.

## Dependencies
- Cooperation with higher-level agents such as cc-manager and project-manager is required.

## References
- BurntSushi. "ripgrep User Guide." https://github.com/BurntSushi/ripgrep/blob/master/GUIDE.md (accessed 2025-03-29).
- ReqView. "Requirements Traceability Matrix Guide." https://www.reqview.com/doc/requirements-traceability-matrix/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Added input/output/failure response to Alfred-specific skills.

## Works well with

- alfred-trust-validation (TAG traceability verification)
- alfred-spec-metadata-validation (SPEC ID validation)

## Files included

- templates/tag-inventory-template.md
