---
paths: "**/sprint-round-naming.md"
---

# Sprint vs Round Naming Convention — SSOT

> **Single Source of Truth** for the distinction between "Sprint" and "Round" terminology.
> Cross-referenced by: `.claude/rules/moai/workflow/spec-workflow.md`, all SPEC plan-phase templates.

---

## Core Rule

[ZONE:Frozen] [HARD] **Sprint**, **Round**, and **Milestone** carry distinct meanings and MUST NOT be used interchangeably.

| Term | Meaning | Scope | Korean Equivalent |
|------|---------|-------|-------------------|
| **Sprint** | Time-unit grouping of multiple SPECs bundled by schedule, release, or thematic focus | Multi-SPEC, project-wide | 스프린트 |
| **Round** | Single-SPEC internal phase grouping; sub-division of a large SPEC to mitigate Anthropic SSE stream stall on large prompts | Within one SPEC | 라운드 |
| **Milestone** | Single-SPEC ordered work steps within a Tier S/M/L lifecycle (M1, M2, ... M6) | Within one SPEC, finer than Round | 마일스톤 |

## Localization

[ZONE:Evolvable] [HARD] **Technical identifiers (SPEC ID, file name, code, frontmatter field) MUST use English** (`Sprint`, `Round`, `Milestone`). **User-facing output in Korean conversations MAY use the Korean equivalent** (`스프린트`, `라운드`, `마일스톤`) for readability.

| Context | English form | Korean form |
|---------|-------------|-------------|
| SPEC ID, memory file name, rule cross-reference | Sprint 1, Round 3, M2 | (English required) |
| Korean documentation, paste-ready resume, progress board | Sprint 1 / 스프린트 1 | 스프린트 1, 라운드 3, M2 |
| Commit message body (per `git_commit_messages: ko`) | mixed allowed | "스프린트 1 4/4 SPEC 완료" |
| Code comments (per `code_comments` setting) | English when `en` | "// 스프린트 1 단계" when `ko` |

When mixing English and Korean in user-facing output, prefer parenthetical pairing on first mention (`Sprint(스프린트) 1 Lane A`) then either form on subsequent mentions.

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

## Round — Single-SPEC Internal Phase

A Round is a **sub-division of a single SPEC's implementation** to mitigate Anthropic SSE `stream_idle_partial` stall for large prompts. Rounds have:

- A sequence number within the SPEC (Round 1, Round 2, ...)
- A task subset (typically 5-10 tasks per Round when SPEC has 30+ tasks total)
- A stall-avoidance rationale (Round-by-Round Agent delegation instead of monolithic SPEC prompt)
- A SINGLE SPEC ID as parent

**Use Round when**:
- Referring to a phase WITHIN a single SPEC
- The split driver is SSE stall mitigation (≥30 tasks or ≥5 phases)
- The Round is invisible at project level (only inside the SPEC plan.md / progress.md)

---

## Milestone — Single-SPEC Lifecycle Step

A Milestone is a **finer-grained single-SPEC work step** (M1, M2, ... M6) within Tier M/L SPEC lifecycle. Milestones are the standard manager-develop delegation unit.

**Milestone vs Round**:
- A SPEC may have 5-6 Milestones (M1~M6) by default — this is the standard Tier M/L workflow structure
- A SPEC has Rounds ONLY if it crosses the SSE stall threshold (≥30 tasks). Most SPECs do NOT have Rounds
- When a SPEC has both Milestones and Rounds, Rounds contain Milestones (Round 1 covers M1~M3, Round 2 covers M4~M6, etc.)

---

## Anti-Patterns

### AP-SRN-001 — Calling a multi-SPEC group "Round"

Incorrect: "Round 1 Lane A 4/4 SPEC complete" when referring to multiple SPECs bundled by schedule.
Correct: "Sprint 1 Lane A 4/4 SPEC complete".

### AP-SRN-002 — Calling a single-SPEC Milestone "Round"

Incorrect: "SPEC-X Round 1 (skill body) → Round 2 (cross-ref) → ..." when these are sequential milestones.
Correct: "SPEC-X M1 (skill body) → M2 (cross-ref) → ... → M6".

Milestones are the standard lifecycle unit; Round is reserved for SSE-stall-driven 30+ task splits.

### AP-SRN-003 — Mixing Sprint and Round in the same context

Incorrect: "Sprint 1 Round 2 implements ..." with no scope distinction.
Correct: "Sprint 1 contains 4 SPECs; SPEC-X within Sprint 1 has 3 Rounds due to 47-task scope".

Always state Sprint AND Round separately when both apply to the same context.

### AP-SRN-004 — Legacy "Wave" terminology

The earlier "Wave" terminology has been retired in favor of "Round" for clarity. Historical references in archived memory or merged commit messages remain unchanged; new content MUST use "Round".

Incorrect (new content): "Wave 1 split across 7 phases"
Correct: "Round 1 split across 7 phases"

---

## Migration Checklist (when correcting drifted naming)

When updating documents or memory entries that use "Round" incorrectly:

- [ ] Determine the intended meaning by checking the scope (multi-SPEC = Sprint; within-SPEC = Round / Milestone)
- [ ] Rename files: `project_<scope>_round<N>_*` → `project_<scope>_sprint<N>_*` (when multi-SPEC bundle)
- [ ] Update body text: "Round N Lane X" → "Sprint N Lane X"
- [ ] Update MEMORY.md index entries
- [ ] Mark old entries with `[SUPERSEDED by <new-file>]` prefix (Lessons Protocol)
- [ ] Cross-reference the rule: `Per .claude/rules/moai/development/sprint-round-naming.md, Sprint = multi-SPEC, Round = within-SPEC.`

---

## Cross-References

- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — Milestone definitions
- `.claude/rules/moai/workflow/session-handoff.md` — paste-ready resume format

---

Version: 2.0.0 (Wave → Round terminology change; AP-SRN-004 added)
Status: Active — applies to all new SPEC bodies, memory entries, paste-ready resume messages, and orchestrator session output
