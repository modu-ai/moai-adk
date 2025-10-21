---

name: moai-alfred-trust-validation
description: Validates TRUST 5-principles compliance (Test coverage 85%+, Code constraints, Architecture unity, Security, TAG trackability). Use when enforcing TRUST checkpoints before progression.
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
---

# Alfred TRUST Validation

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | /alfred:3-sync quality gate |
| Trigger cues | TRUST checklist enforcement, release readiness scoring, risk gating. |

## What it does

Validates MoAI-ADK's TRUST 5-principles compliance to ensure code quality, testability, security, and traceability.

## When to use

- Activates when Alfred evaluates TRUST compliance before handoff.
- "Check the TRUST principle", "Quality verification", "Check code quality"
- Automatically invoked by `/alfred:3-sync`
- Before merging PR or releasing

## How it works

**T - Test First**:
- Checks test coverage ≥85% (pytest, vitest, go test, cargo test, etc.)
- Verifies TDD cycle compliance (RED → GREEN → REFACTOR)

**R - Readable**:
- File ≤300 LOC
- Function ≤50 LOC
- Parameters ≤5
- Cyclomatic complexity ≤10

**U - Unified**:
- SPEC-driven architecture consistency
- Clear module boundaries
- Language-specific standard structures

**S - Secured**:
- Input validation implementation
- No hardcoded secrets
- Access control applied

**T - Trackable**:
- TAG chain integrity (@SPEC → @TEST → @CODE → @DOC)
- No orphaned TAGs
- No duplicate SPEC IDs

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
- SonarSource. "Quality Gate: Developer's Guide." https://www.sonarsource.com/company/newsroom/white-papers/quality-gate/ (accessed 2025-03-29).
- ISO/IEC 25010. "Systems and software quality models." (accessed 2025-03-29).

## Changelog
- 2025-03-29: Added input/output/failure response to Alfred-specific skills.

## Works well with

- alfred-tag-scanning (TAG traceability)
- alfred-code-reviewer (code quality analysis)

## Files included

- templates/trust-report-template.md
