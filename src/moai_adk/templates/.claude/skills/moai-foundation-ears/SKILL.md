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
- Mavin, A., et al. "Easy Approach to Requirements Syntax (EARS)." IEEE RE, 2009.
- INCOSE. "Guide for Writing Requirements." INCOSE-TP-2010-006-02 (accessed 2025-03-29).

## Changelog
- 2025-03-29: Foundation skill templates have been enhanced to align with best practice structures.

## Works well with

- moai-foundation-specs

## Best Practices
- When changing standards, the reason for change and supporting documents are recorded.
- Follow the single source principle and avoid modifying the same item in multiple places.
