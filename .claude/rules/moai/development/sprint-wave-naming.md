# Sprint vs Wave Naming Convention — SSOT

> **Single Source of Truth** for the distinction between "Sprint" and "Wave" terminology.
> Cross-referenced by: `.claude/rules/moai/workflow/spec-workflow.md`, all SPEC plan-phase templates.

---

## Core Rule

[ZONE:Frozen] [HARD] **Sprint** and **Wave** carry distinct meanings and MUST NOT be used interchangeably.

| Term | Meaning | Scope |
|------|---------|-------|
| **Sprint** | Time-unit grouping of multiple SPECs bundled by schedule, release, or thematic focus | Multi-SPEC, project-wide |
| **Wave** | Single-SPEC internal phase grouping; sub-division of a large SPEC to mitigate Anthropic SSE stream stall on large prompts | Within one SPEC |
| **Milestone** | Single-SPEC ordered work steps within a Tier S/M/L lifecycle (M1, M2, ... M6) | Within one SPEC, finer than Wave |

---

## Sprint — Multi-SPEC Time-Unit

A Sprint is a **time-unit container** for one or more SPECs being worked on together. Sprints have:

- A sequence number (Sprint 1, Sprint 2, ...)
- An optional lane subdivision for parallel tracks (e.g., "Sprint N Lane A" = SPECs working in series; "Sprint N Lane B" = SPECs in a parallel track)
- A scheduling rationale (release cut, dependency batch, thematic refactor)
- One or more SPEC IDs as members

**Use Sprint when**:
- Referring to a group of SPECs (≥2 SPECs)
- The grouping driver is schedule/release/theme (NOT internal phase split)
- The grouping is project-wide visible (release-tracker, MEMORY.md headers, paste-ready resume)

---

## Wave — Single-SPEC Internal Phase

A Wave is a **sub-division of a single SPEC's implementation** to mitigate Anthropic SSE `stream_idle_partial` stall for large prompts. Waves have:

- A sequence number within the SPEC (Wave 1, Wave 2, ...)
- A task subset (typically 5-10 tasks per Wave when SPEC has 30+ tasks total)
- A stall-avoidance rationale (Wave-by-Wave Agent delegation instead of monolithic SPEC prompt)
- A SINGLE SPEC ID as parent

**Use Wave when**:
- Referring to a phase WITHIN a single SPEC
- The split driver is SSE stall mitigation (≥30 tasks or ≥5 phases)
- The Wave is invisible at project level (only inside the SPEC plan.md / progress.md)

---

## Milestone — Single-SPEC Lifecycle Step

A Milestone is a **finer-grained single-SPEC work step** (M1, M2, ... M6) within Tier M/L SPEC lifecycle. Milestones are the standard manager-develop delegation unit.

**Milestone vs Wave**:
- A SPEC may have 5-6 Milestones (M1~M6) by default — this is the standard Tier M/L workflow structure
- A SPEC has Waves ONLY if it crosses the SSE stall threshold (≥30 tasks). Most SPECs do NOT have Waves
- When a SPEC has both Milestones and Waves, Waves contain Milestones (Wave 1 covers M1~M3, Wave 2 covers M4~M6, etc.)

---

## Anti-Patterns

### AP-SWN-001 — Calling a multi-SPEC group "Wave"

Incorrect: "Wave 1 Lane A 4/4 SPEC complete" when referring to multiple SPECs bundled by schedule.
Correct: "Sprint 1 Lane A 4/4 SPEC complete".

### AP-SWN-002 — Calling a single-SPEC Milestone "Wave"

Incorrect: "SPEC-X Wave 1 (skill body) → Wave 2 (cross-ref) → ..." when these are sequential milestones.
Correct: "SPEC-X M1 (skill body) → M2 (cross-ref) → ... → M6".

Milestones are the standard lifecycle unit; Wave is reserved for SSE-stall-driven 30+ task splits.

### AP-SWN-003 — Mixing Sprint and Wave in the same context

Incorrect: "Sprint 1 Wave 2 implements ..." with no scope distinction.
Correct: "Sprint 1 contains 4 SPECs; SPEC-X within Sprint 1 has 3 Waves due to 47-task scope".

Always state Sprint AND Wave separately when both apply to the same context.

---

## Migration Checklist (when correcting drifted naming)

When updating documents or memory entries that use "Wave" incorrectly:

- [ ] Determine the intended meaning by checking the scope (multi-SPEC = Sprint; within-SPEC = Wave / Milestone)
- [ ] Rename files: `project_<scope>_wave<N>_*` → `project_<scope>_sprint<N>_*` (when multi-SPEC bundle)
- [ ] Update body text: "Wave N Lane X" → "Sprint N Lane X"
- [ ] Update MEMORY.md index entries
- [ ] Mark old entries with `[SUPERSEDED by <new-file>]` prefix (Lessons Protocol)
- [ ] Cross-reference the rule: `Per .claude/rules/moai/development/sprint-wave-naming.md, Sprint = multi-SPEC, Wave = within-SPEC.`

---

## Cross-References

- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — Milestone definitions
- `.claude/rules/moai/workflow/session-handoff.md` — paste-ready resume format

---

Version: 1.0.0
Status: Active — applies to all new SPEC bodies, memory entries, paste-ready resume messages, and orchestrator session output
