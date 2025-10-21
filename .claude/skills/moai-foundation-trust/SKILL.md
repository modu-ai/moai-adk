---

name: moai-foundation-trust
description: Validates TRUST 5-principles (Test 85%+, Readable, Unified, Secured, Trackable). Use when aligning with TRUST governance.
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
---

# Foundation: TRUST Validation

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | SessionStart (foundation bootstrap) |
| Trigger cues | TRUST compliance checks, release readiness reviews, quality gate enforcement. |

## What it does

Validates MoAI-ADK's TRUST 5-principles compliance to ensure code quality, testability, security, and traceability.

## When to use

- Activates when TRUST compliance or release readiness needs to be evaluated.
- "Check the TRUST principle", "Quality verification", "Check code quality"
- Automatically invoked by `/alfred:3-sync`
- Before merging PR or releasing

## How it works

**T - Test First**:
- Checks test coverage ≥85% (pytest, vitest, go test, cargo test, etc.)
- Verifies TDD cycle compliance (RED → GREEN → REFACTOR)

**R - Readable**:
- File ≤300 LOC, Function ≤50 LOC, Parameters ≤5, Complexity ≤10

**U - Unified**:
- SPEC-driven architecture consistency, Clear module boundaries

**S - Secured**:
- Input validation, No hardcoded secrets, Access control

**T - Trackable**:
- TAG chain integrity (@SPEC → @TEST → @CODE → @DOC)

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
- SonarSource. "Quality Gate: Developer's Guide." https://www.sonarsource.com/company/newsroom/white-papers/quality-gate/ (accessed 2025-03-29).
- ISO/IEC 25010. "Systems and software quality models." (accessed 2025-03-29).

## Changelog
- 2025-03-29: Foundation skill templates have been enhanced to align with best practice structures.

## Works well with

- moai-foundation-tags (TAG traceability)
- moai-foundation-specs (SPEC validation)

## Examples
```markdown
- Scan standard documents and report missing sections.
- Changed regulations are reflected in CLAUDE.md.
```

## Best Practices
- When changing standards, the reason for change and supporting documents are recorded.
- Follow the single source principle and avoid modifying the same item in multiple places.
