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
- Mavin, A., et al. "Easy Approach to Requirements Syntax (EARS)." IEEE RE, 2009.
- INCOSE. "Guide for Writing Requirements." INCOSE-TP-2010-006-02 (accessed 2025-03-29).

## Changelog
- 2025-03-29: Added input/output/failure response to Alfred-specific skills.

## Works well with

- alfred-spec-metadata-validation
- alfred-trust-validation

## Reference

`.moai/memory/development-guide.md#ears-requirements-how-to`
