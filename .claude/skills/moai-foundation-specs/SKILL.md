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
- INCOSE. "Guide for Writing Requirements." INCOSE-TP-2010-006-02 (accessed 2025-03-29).
- IEEE. "Software Requirements Specification Standard." IEEE 830-1998.

## Changelog
- 2025-03-29: Foundation skill templates have been enhanced to align with best practice structures.

## Works well with

- moai-foundation-ears
- moai-foundation-tags

## Best Practices
- When changing standards, the reason for change and supporting documents are recorded.
- Follow the single source principle and avoid modifying the same item in multiple places.
