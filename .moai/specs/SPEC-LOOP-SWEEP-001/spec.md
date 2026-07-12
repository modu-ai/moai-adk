---
id: SPEC-LOOP-SWEEP-001
title: "Loop Sweep — /moai loop redefined as project-wide improvement sweep built on the goal engine"
version: "0.1.0"
status: draft
created: 2026-07-12
updated: 2026-07-12
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/skills/moai/workflows, .claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "agentic-core, loop-sweep, goal-preset, project-wide, scan-lens, diagnostics"
era: V3R6
tier: M
depends_on: [SPEC-GOAL-ENGINE-001]
---

# SPEC-LOOP-SWEEP-001 — Loop Sweep (`/moai loop` as a goal preset)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-12 | manager-spec | Initial plan-phase authoring. Epic AGENTIC-CORE, SPEC 3 of 3. depends_on SPEC-GOAL-ENGINE-001. Shared findings in SPEC-ANALYZE-FIRST-ROUTING-001/research.md (§C loop/fix). |

> **Epic**: AGENTIC-CORE (schema has no `epic:` field; recorded in body). SPEC 3 of 3.
> Docs-heavy (no new Go engine — builds on `SPEC-GOAL-ENGINE-001`'s engine).

## §A — Context and Motivation (WHY)

Today `/moai loop` is a prose iteration over EXACTLY four diagnostic tools
(LSP / AST-grep / test / coverage) with a mechanical exit "zero errors + tests +
coverage" (`loop.md:54`, `:79`, `:122-123`, `:173-198`, `:231`; research.md §C.1a).
It occupies the **goal-based** quadrant of the loop taxonomy (`loop.md:54`) but
predates the goal engine. With `SPEC-GOAL-ENGINE-001` delivering a universal
condition-declared loop, `/moai loop` becomes redundant machinery: its
"iterate-until-clean" behavior is exactly a goal preset whose condition is
"issue queue drained + diagnostics clean."

This SPEC REDEFINES `/moai loop` as a **project-wide improvement sweep** built ON
the goal engine as a preset. A scan stage supplies a FINITE issue queue; the goal
condition is "queue drained + diagnostics clean(zero errors + tests[+coverage when
enabled])" bounded by a ceiling. It also reconciles the two unrelated "loop"
surfaces (skill vs the Go `moai loop` SPEC-lifecycle controller; research.md §C.1)
and the four-quadrant taxonomy. `/moai fix` and `/moai review` are UNCHANGED;
their relationship to the redefined loop is documented.

### §A.1 The user's "loop vs review overlap" concern (resolved in §B)

The redefinition MUST explicitly resolve the overlap between `/moai loop` and
`/moai review`: review is a read-only lens (report-only, unchanged); loop CONSUMES
review lenses as queue suppliers. Both surfaces are documented as layered, not
competing (spec.md §B.2 + a 1-2 paragraph loop.md/review.md cross-ref).

## §B — Scope (WHAT this SPEC delivers)

- **D1 — loop.md rewrite (loop = goal preset)**: `/moai loop` is a goal preset with
  condition "issue queue drained + diagnostics clean(zero errors + tests[+coverage
  when enabled])" + ceiling. The scan stage supplies a FINITE queue: default lenses
  = LSP + lint + test failures + review lenses (security, `@MX`); opt-in
  `--lens clean|simplify|coverage`. HARD boundary: no work outside the queue
  ("no invented improvements"); empty queue → exit. PRESERVE: ceiling precedence
  (CLI `--max` > ralph.yaml `loop.max_iterations` > workflow.yaml
  `loop_prevention.max_iterations`), the 5-section ceiling-exit verdict,
  memory-pressure guard, `loop-verdict-<id>.json` (extend `exit_kind`), and the
  Step 1.5 independent final pass.
- **D2 — review relationship**: review = read-only lens; standalone review
  report-only UNCHANGED; loop consumes review lenses as queue suppliers. Document
  the layering in loop.md + a review.md cross-ref (1-2 paragraphs, NO review.md
  behavior change). Resolve the user's overlap concern in spec.md prose (§A.1/§B.2).
- **D3 — fix relationship**: `/moai fix` UNCHANGED; the residue handoff text
  (recommends `/moai loop`) is updated so residue enters the loop queue.
- **D4 — doctrine reconciliation**: re-express the four-quadrant taxonomy
  (goal engine + presets) in loop.md/fix.md siblings; add a `goal-directive.md`
  table row; `cadence-bridge.md` eligibility (loop still NOT cadence-eligible);
  `spec-workflow.md` § Subcommand Classification loop row (Multi-Agent; keep/retire
  the `run --mode loop` alias — decide + justify); `CLAUDE.md:46` wording;
  `moai-workflow-loop` SKILL.md (preset architecture); Go CLI `moai loop`
  disambiguation — minimal: rename help text to clarify it is the SPEC-lifecycle
  controller, document the two surfaces, defer engine unification (plan § Deferred).
- **D5 — Template mirrors** for changed skill/rule files + `make build`.

### §B.2 loop vs review layering (resolves §A.1)

- **`/moai review`**: read-only. Produces findings/report. Does NOT modify code,
  does NOT iterate. Behavior UNCHANGED by this SPEC.
- **`/moai loop`**: a mutation sweep. Its scan stage may INVOKE review lenses
  (security, `@MX`) as queue SUPPLIERS — the review lens produces findings, the
  loop enqueues them as fixable issues, and iterates until the queue drains. Loop
  never re-defines what review reports; it consumes review output as one of
  several queue sources (LSP, lint, test failures, review lenses).
- **Non-overlap**: run a review to SEE findings without changing anything; run a
  loop to FIX the finite set of issues the scan (including review lenses) found.

## §C — GEARS Requirements

> Notation: GEARS (current). `<subject>` generalized. Each REQ maps to an AC.

### §C.1 D1 — loop = goal preset

- **REQ-LSW-001** (Ubiquitous): The `loop.md` workflow shall define `/moai loop`
  as a goal preset whose completion condition is "issue queue drained AND
  diagnostics clean (zero errors + tests passing [+ coverage when enabled])".
- **REQ-LSW-002** (Ubiquitous): The scan stage shall build a FINITE issue queue
  from default lenses (LSP, lint, test failures, review lenses: security + `@MX`)
  and shall accept opt-in `--lens clean|simplify|coverage`.
- **REQ-LSW-003** (Unwanted behavior): The loop shall not perform work outside the
  scanned queue ("no invented improvements"); an empty queue shall cause an
  immediate exit.
- **REQ-LSW-004** (Ubiquitous): The rewrite shall PRESERVE the iteration-ceiling
  precedence (CLI `--max` > ralph.yaml `loop.max_iterations` > workflow.yaml
  `loop_prevention.max_iterations`), the 5-section ceiling-exit verdict, the
  memory-pressure guard, and the Step 1.5 independent final pass.
- **REQ-LSW-005** (Event-driven): **When** the loop exits at the ceiling, it shall
  persist residue to `.moai/state/loop-verdict-<id>.json` with `exit_kind:
  "sweep-residue"` (settled value) — added ADDITIVELY as a fourth value alongside
  the base `ceiling | manual-residue | one-shot-residue` enum, without reassigning
  the base enum owner.

### §C.2 D2 — review relationship

- **REQ-LSW-006** (Ubiquitous): `loop.md` shall document that `/moai loop` consumes
  review lenses (security, `@MX`) as queue suppliers, and that standalone
  `/moai review` remains read-only and report-only.
- **REQ-LSW-007** (Ubiquitous): A 1-2 paragraph cross-reference shall be added
  linking `review.md` and `loop.md` (the layering), with NO behavior change to
  `/moai review`.

### §C.3 D3 — fix relationship

- **REQ-LSW-008** (Ubiquitous): `/moai fix` behavior shall remain UNCHANGED
  (single-pass Agentless pipeline); only the residue-handoff TEXT recommending
  `/moai loop` shall be updated to state that residue enters the loop queue.

### §C.4 D4 — doctrine reconciliation

- **REQ-LSW-009** (Ubiquitous): The four-quadrant taxonomy shall be re-expressed
  as "goal engine + presets" consistently across the `loop.md`/`fix.md` sibling
  quadrant notes.
- **REQ-LSW-010** (Ubiquitous): `goal-directive.md` shall carry a table row for
  `/moai loop` as a goal preset (distinct from native `/goal` and `/moai goal`).
- **REQ-LSW-011** (State-driven): **While** documenting cadence eligibility,
  `cadence-bridge.md` shall continue to state that `/moai loop` is NOT
  cadence-eligible (it mutates / may enter run-phase; cadence recipes are read-only).
- **REQ-LSW-012** (Ubiquitous): The `spec-workflow.md` § Subcommand Classification
  loop row shall be updated (loop = Multi-Agent) and shall state the settled
  `run --mode loop` alias disposition — **KEEP** (both `/moai run --mode loop` and
  `/moai loop` resolve to the goal-preset sweep) — with its backward-compat
  justification.
- **REQ-LSW-013** (Ubiquitous): The Go CLI `moai loop` help text shall be renamed
  to clarify it is the SPEC-lifecycle controller (distinct from the `/moai loop`
  sweep skill), and the two surfaces shall be documented; engine unification is
  deferred.

### §C.5 D5 — Template mirrors

- **REQ-LSW-014** (Event-driven): **When** the changed skill/rule files are
  mirrored, `make build` shall succeed and the template bodies shall carry no
  internal SPEC ID (§25 neutrality).

## §D — Exclusions (What NOT to Build)

[HARD] This SPEC explicitly does NOT deliver the following.

### §D.1 Out of Scope — the goal engine itself

- The condition-declared goal engine (`internal/goal/`, `moai hook stop-goal`,
  per-session state) is owned by `SPEC-GOAL-ENGINE-001`. This SPEC only CONFIGURES
  a loop PRESET on top of it; it builds no new engine code.

### §D.2 Out of Scope — Go `moai loop` engine unification

- Unifying the Go CLI `moai loop` SPEC-lifecycle controller (`internal/cli/loop.go`,
  `internal/loop`, `internal/ralph`) with the goal/skill loop is DEFERRED
  (research.md §C.1). This SPEC only renames help text + documents the two
  surfaces (REQ-LSW-013).

### §D.3 Out of Scope — /moai fix and /moai review behavior

- `/moai fix` and `/moai review` behavior is UNCHANGED. Only cross-reference TEXT
  is touched (REQ-LSW-007, REQ-LSW-008). No pipeline-contract or lens change.

### §D.4 Out of Scope — cadence eligibility change

- `/moai loop` stays NOT cadence-eligible (REQ-LSW-011). This SPEC does NOT add
  loop to the cadence-bridge recipe catalog.

### §D.5 Out of Scope — docs-site 4-locale translation

- Translating the redefined loop documentation into docs-site locales is a
  DEFERRED follow-up (plan § Deferred), NOT run-phase scope.

## §E — Dependencies and Follow-ups

- **depends_on**: `SPEC-GOAL-ENGINE-001` (frontmatter) — loop is a preset ON the
  goal engine. Run-phase entry follows the Depends_on pre-flight (GOAL-ENGINE must
  be `completed`, or `--ignore-deps` + logged rationale).
- **Coordination**: `SPEC-CADENCE-BRIDGE-001` (completed) owns the time-based
  quadrant; REQ-LSW-011 keeps loop cadence-ineligible (no conflict; research.md §D.3).
- **Doctrine anchors**: shared `research.md` §C (loop/fix), §D.3 (cadence);
  `loop.md`; `fix.md`; `goal-directive.md`; `cadence-bridge.md`;
  `spec-workflow.md` § Subcommand Classification.
