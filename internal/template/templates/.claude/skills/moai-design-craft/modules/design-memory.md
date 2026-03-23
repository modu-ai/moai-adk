# Design Memory: system.md Protocol

`.moai/design/system.md` is the persistent design memory for a project. It accumulates design intent, domain vocabulary, and craft decisions across sessions and SPECs.

## File Location

```
.moai/design/system.md
```

Tracked in git. Shared across the team. Not ignored by `.gitignore`.

## When to Read

Read `.moai/design/system.md` before:
- Starting any new UI-adjacent SPEC
- Reviewing existing UI components for design compliance
- Making copy or label decisions
- Naming new components

## When to Write

Write to `.moai/design/system.md` after:
- Completing a Design Direction phase (new domain vocabulary, design intent)
- Discovering a persistent design principle during a critique
- Extracting patterns from existing code that should be preserved

## File Structure

```markdown
# Design System

## Design Intent
[1–3 sentence statement of the overall product design intent]

## Domain Vocabulary
[Table or list of canonical domain terms with brief definitions]

## Craft Principles
[Numbered list of non-negotiable quality constraints]

## Per-Feature Direction
### [SPEC-ID or Feature Name]
[Design Direction statement specific to this feature]

## Anti-Patterns
[What NOT to do — patterns that were tried and failed]
```

## Write Protocol

1. Read the current `.moai/design/system.md` first
2. Identify the section to update (do not create duplicate sections)
3. If the file already contains a Design Direction for the same SPEC or feature:
   - Compare new direction against existing direction
   - If conflicting: Present both versions to the user and ask "Overwrite existing direction or merge with it?"
   - If additive: Append new sections without modifying existing entries
4. Append or update — do not silently overwrite existing entries
5. Keep entries concise: design memory is a reference, not documentation

## Stub Template

When `.moai/design/system.md` does not exist, create it with the stub from `internal/template/templates/.moai/design/system.md`. The stub provides structure without prescribing decisions.

## Design Memory vs SPEC Documents

| | `.moai/design/system.md` | `.moai/specs/SPEC-*/spec.md` |
|---|---|---|
| Scope | Project-wide, persistent | SPEC-specific, time-bounded |
| Content | Vocabulary, intent, principles | Requirements, acceptance criteria |
| Authors | Any design-aware agent | manager-spec |
| Lifecycle | Grows over time | Completed when SPEC closes |
