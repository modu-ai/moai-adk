---
paths: "**/sprint-round-naming.md"
---

# Epic Naming Taxonomy — SSOT

> **Single Source of Truth** for the Epic-based naming taxonomy. Defines exactly four canonical terms (`Epic`, `SPEC`, `Milestone`, `Constitution`) and documents the retired legacy aliases (`Sprint`, `cohort`, `Round`, `Wave`).
> Cross-referenced by: `.claude/rules/moai/workflow/spec-workflow.md`, all SPEC plan-phase templates, `.claude/output-styles/moai/moai.md` § Banner Localization.

---

## Core Rule

[ZONE:Frozen] [HARD] The MoAI naming taxonomy defines **exactly four canonical terms**. They carry distinct meanings and MUST NOT be used interchangeably. The legacy aliases (`Sprint`, `cohort`, `Round`, `Wave`) are RETIRED and MUST NOT appear as canonical terms in new content (they survive only in the Legacy Aliases section below and in historical artifacts).

| Term | Meaning | Scope | Korean Equivalent |
|------|---------|-------|-------------------|
| **Epic** | A time-unit or thematic container for one or more SPECs, grouped by schedule, release, or theme (formerly `Sprint`) | Multi-SPEC, project-wide | 에픽 |
| **SPEC** | A single work unit (feature, refactor, bugfix) with a unique `SPEC-{DOMAIN}-{NUM}` identifier | Single work unit | (English required) |
| **Milestone** | An ordered within-SPEC work step (M1, M2, ... M6) within a Tier S/M/L lifecycle — the standard manager-develop delegation unit | Within one SPEC | 마일스톤 |
| **Constitution** | Project-level governing principles and development guidelines (referenced for SDD vocabulary alignment) | Project-wide | 컨스티튜션 / 프로젝트 헌장 |

> **Semantics note (AC-LR-009)**: `Epic` is a pure rename of the former `Sprint`. Only the label changes — the grouping concept (multi-SPEC bundle by schedule, release, or thematic focus) is identical. An Epic MAY carry a sequence number (`Epic 1`, `Epic 2`) for time-unit grouping OR a thematic name (`Epic Docs-v3`) for release-cut grouping; both forms are valid, matching the former Sprint flexibility.

### What `Milestone` absorbs (formerly `Round`)

The former within-SPEC term `Round` (a sub-division of a single SPEC's implementation to mitigate Anthropic SSE `stream_idle_partial` stall on large prompts) is **folded into `Milestone`**. The Milestone sequence (M1, M2, ... M6) already provides within-SPEC ordered work steps. A SPEC that previously had "Round 1 covers M1~M3, Round 2 covers M4~M6" now simply has Milestones M1~M6 with no intermediate grouping layer. Most SPECs have 5-6 Milestones and no further sub-grouping; the SSE-stall mitigation is handled by manager-develop's per-Milestone delegation rather than by a separate naming layer.

---

## Localization

[ZONE:Evolvable] [HARD] **Technical identifiers (SPEC ID, file name, code, frontmatter field) MUST use English** (`Epic`, `SPEC`, `Milestone`, `Constitution`). **User-facing output in Korean conversations MAY use the Korean equivalent** (`에픽`, `마일스톤`, `컨스티튜션`) for readability.

| Context | English form | Korean form |
|---------|-------------|-------------|
| SPEC ID, file name, rule cross-reference | Epic 1, Milestone M2 | (English required) |
| Korean documentation, paste-ready resume, progress board | Epic 1 / 에픽 1 | 에픽 1, 마일스톤 M2 |
| Commit message body (per `git_commit_messages: ko`) | mixed allowed | "에픽 1 4/4 SPEC 완료" |
| Code comments (per `code_comments` setting) | English when `en` | "// 에픽 1 단계" when `ko` |

When mixing English and Korean in user-facing output, prefer parenthetical pairing on first mention (`Epic(에픽) 1 Lane A`) then either form on subsequent mentions.

---

## Epic — Multi-SPEC Container

An Epic is a **time-unit or thematic container** for one or more SPECs being worked on together. Epics have:

- A sequence number (Epic 1, Epic 2, ...) OR a thematic name (Epic Docs-v3, Epic Harness-Books) — both forms valid
- An optional lane subdivision for parallel tracks (e.g., "Epic N Lane A" = SPECs working in series; "Epic N Lane B" = SPECs in a parallel track)
- An optional cohort sub-grouping for Epic-internal bundling ("Epic N cohort")
- A scheduling rationale (release cut, dependency batch, thematic refactor)
- One or more SPEC IDs as members

**Use Epic when**:
- Referring to a group of SPECs (≥2 SPECs; a single-SPEC Epic is also valid when the grouping concept applies)
- The grouping driver is schedule/release/theme (NOT internal phase split)
- The grouping is project-wide visible (release-tracker, MEMORY.md headers, paste-ready resume)

---

## SPEC — Single Work Unit

A SPEC is a **single work unit** (feature, refactor, bugfix) with a unique `SPEC-{DOMAIN}-{NUM}` identifier. It is the atomic unit of plan→run→sync lifecycle execution. This aligns directly with Spec Kit's `spec` concept.

---

## Milestone — Within-SPEC Ordered Step

A Milestone is an **ordered within-SPEC work step** (M1, M2, ... M6) within a Tier M/L SPEC lifecycle. Milestones are the standard manager-develop delegation unit. The Milestone sequence also absorbs the former `Round` concept (within-SPEC SSE-stall sub-division) — see "What `Milestone` absorbs" above.

**Milestone granularity**:
- A SPEC has 5-6 Milestones (M1~M6) by default — this is the standard Tier M/L workflow structure
- The former "Round" sub-grouping (Round 1 = M1~M3, Round 2 = M4~M6) is no longer a separate naming layer; manager-develop sequences Milestones directly
- SSE-stall mitigation for ≥30-task SPECs is handled by per-Milestone delegation rather than by a Round naming layer

---

## Constitution — Project Governance

The Constitution is the **project-level governing layer**: the technology stack, naming conventions, forbidden libraries, architectural patterns, security standards, and logging standards that all SPECs must respect. Verified during plan-phase against `.moai/project/tech.md`. This term aligns directly with Spec Kit's `constitution` concept.

> **Deferred (EX-6)**: This taxonomy references `Constitution` for vocabulary alignment but does NOT introduce a `/moai constitution` slash command. A follow-up SPEC MAY introduce such a command if Spec Kit-style governance commands are desired.

---

## Legacy Aliases (RETIRED — migration reference only)

The following terms are **RETIRED** from the canonical vocabulary. They appear here ONLY as a migration reference so readers of historical commits, memory entries, and pre-redesign SPECs can map them to the current taxonomy. **New content MUST NOT use these as canonical terms.**

| Legacy term | Current canonical term | Disposition |
|-------------|------------------------|-------------|
| `Sprint` | `Epic` | Renamed (semantics preserved — AC-LR-009) |
| `Sprint N Lane A/B` | `Epic N Lane A/B` | Renamed (Lane retained) |
| `cohort` | (folded into Epic) | Removed as standalone term; use "Epic N cohort" or "Epic N Lane X" for Epic-internal sub-grouping |
| `Round` | (folded into Milestone) | Removed; use the Milestone sequence (M1~M6) for within-SPEC ordered steps |
| `Wave` | (retired) | Confirmed retired (was already pre-retired pre-redesign per the former AP-SRN-004) |

> **Historical artifacts preserved**: merged commit messages, archived memory entries, and completed SPEC bodies written before the redesign retain their original `Sprint`/`cohort`/`Round`/`Wave` vocabulary unchanged. The `[SUPERSEDED by ...]` marker convention handles supersession in the memory index; SPEC bodies are immutable post-completion.

---

## Anti-Patterns

### AP-SRN-001 — Calling a multi-SPEC group by a legacy alias instead of "Epic"

Incorrect: "Round 1 Lane A 4/4 SPEC complete", "Wave 1 Lane A 4/4 SPEC complete", or "cohort 1 4/4 SPEC complete" when referring to multiple SPECs bundled by schedule.
Correct: "Epic 1 Lane A 4/4 SPEC complete".

### AP-SRN-002 — Calling a within-SPEC Milestone "Round"

Incorrect: "SPEC-X Round 1 (skill body) → Round 2 (cross-ref) → ..." when these are sequential milestones.
Correct: "SPEC-X M1 (skill body) → M2 (cross-ref) → ... → M6".

Milestones are the sole within-SPEC ordered-step unit. The former `Round` sub-grouping layer is folded into the Milestone sequence.

### AP-SRN-003 — Mixing Epic and a legacy alias in the same context

Incorrect: "Epic 1 Round 2 implements ..." or "Epic 1 cohort 2 implements ..." with no scope distinction.
Correct: "Epic 1 contains 4 SPECs; SPEC-X within Epic 1 has 6 Milestones (M1~M6) due to a 47-task scope".

When an Epic-internal sub-grouping is needed, use "Epic N Lane X" or "Epic N cohort" — never a standalone `cohort` or `Round` as if it were a canonical term.

### AP-SRN-004 — Legacy "Wave" terminology (RETIRED)

The "Wave" terminology has been retired. Historical references in archived memory or merged commit messages remain unchanged; new content MUST NOT use "Wave" as a SPEC-grouping term.

Incorrect (new content): "Wave 1 split across 7 phases"
Correct: "Epic 1 split across 7 Milestones"

> **Disambiguation note (NOT the taxonomy `Wave`)**: ci-watch pipeline infrastructure uses "Wave 1/2/3/5" as CI-pipeline wave numbering (a separate legacy concept). Those infra-wave references are unrelated to this naming taxonomy and are exempt from AP-SRN-004.

### AP-SRN-005 — Using `cohort` as a standalone multi-SPEC grouping term

Incorrect: "the Q3 cohort shipped 4 SPECs" (treating `cohort` as a top-level grouping).
Correct: "Epic Q3 shipped 4 SPECs" (or "Epic Q3 cohort" for an Epic-internal sub-grouping).

`cohort` is removed as a standalone canonical term. It survives only as an Epic-internal sub-naming ("Epic N cohort") parallel to "Epic N Lane X".

---

## Migration Checklist (when correcting drifted naming)

When updating documents or memory entries that use a legacy alias:

- [ ] Determine the intended meaning by checking the scope (multi-SPEC = Epic; within-SPEC ordered step = Milestone)
- [ ] Rename files: `project_<scope>_sprint<N>_*` → `project_<scope>_epic<N>_*` (when multi-SPEC bundle) — NOTE: memory-file renames are out of scope for the initial migration (EX-4) and are handled by `[SUPERSEDED by ...]` markers in new entries
- [ ] Update body text: "Sprint N Lane X" → "Epic N Lane X"; "Round N" (within-SPEC) → "Milestone M<N>"
- [ ] Update MEMORY.md index entries (preserve historical entries with `[SUPERSEDED by <new-file>]` prefix per Lessons Protocol)
- [ ] Cross-reference the rule: `Per .claude/rules/moai/development/sprint-round-naming.md, Epic = multi-SPEC, Milestone = within-SPEC ordered step.`

> **Out of scope for the initial Axis B migration (EX-4)**: memory files (`project_sprint*_*.md`), MEMORY.md index entries, docs-site content, and archived/historical (T5) paths. Those are deferred to follow-up SPECs. The binary pass condition for the initial migration is "0 residual matches in T1-T4 scope excluding the Legacy Aliases section of this file and excluding grep false-positives (ci-watch infra Wave, Socratic interview Round, generic DB 'cohort')".

---

## Cross-References

- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — Milestone definitions
- `.claude/rules/moai/workflow/session-handoff.md` — paste-ready resume format
- `.claude/output-styles/moai/moai.md` § Banner Localization — `Epic [N]` token rendering across locales
- the lifecycle-redesign design notes §C — Epic taxonomy design decision and migration mapping

---

Version: 3.0.0 (Epic taxonomy rewrite — Sprint/cohort/Round/Wave retired as canonical terms; Epic / SPEC / Milestone / Constitution are the four canonical terms. Axis B of the lifecycle-redesign SPEC.)
Status: Active — applies to all new SPEC bodies, memory entries, paste-ready resume messages, and orchestrator session output
