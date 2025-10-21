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
- IEEE. "Software Requirements Specification Standard." IEEE 830-1998.
- NASA. "Systems Engineering Handbook." https://www.nasa.gov/seh/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Added input/output/failure response to Alfred-specific skills.

## Works well with

- alfred-ears-authoring (SPEC writing guide)
- alfred-tag-scanning (SPEC ID duplication check)

## Reference

SSOT (Single Source of Truth): `.moai/memory/spec-metadata.md`
