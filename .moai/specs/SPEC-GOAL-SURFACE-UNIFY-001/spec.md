---
id: SPEC-GOAL-SURFACE-UNIFY-001
title: Unify the goal surface on /moai goal and relocate goal presentation to the Implementation Kickoff Approval gate
version: 1.0.0
status: draft
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: HIGH
phase: plan
module: doctrine
lifecycle: spec-anchored
tags: [goal, doctrine, session-handoff, slash-command, template-mirror]
tier: L
---

## HISTORY

| Version | Date | Change | Author |
|---------|------|--------|--------|
| 1.0.0 | 2026-07-25 | Initial plan-phase authoring (Tier L, 6 milestones) | manager-spec |
| 1.1.0 | 2026-07-25 | Scope expansion per approved decisions D4 (Go emission paths → M7, `cycle_type: tdd`) and D5 (public docs → sync phase). Scope now spans three layers: doctrine, Go code, public docs. | manager-spec |

---

## §A Context

Two distinct commands are conflated across the MoAI doctrine surface today:

- **native `/goal`** — a Claude Code built-in TUI command. HUMAN-ONLY: the model cannot invoke it on the user's behalf, and a slash line pasted mid-body is inert plain text (slash commands parse only at input start).
- **`/moai goal`** — MoAI's own programmatic reimplementation. The orchestrator can register and arm it directly; the `stop-goal` Stop-hook evaluator judges the condition at each turn-end.

Because native `/goal` is HUMAN-ONLY, the doctrine grew an elaborate two-step delivery mechanism (`§ Post-Paste /goal Follow-up Block`) whose entire reason for existing is that the model cannot type the command. `/moai goal` is orchestrator-armed, so that mechanism is obsolete the moment the emission surface switches.

A third, independent defect: the `goal` subcommand is the only `/moai` subcommand with no `/moai:<sub>` slash-command wrapper, so it is absent from the `/`-prefix command list users browse.

Measured baseline at `origin/main` = `e306e21a9` (commands and observed outputs recorded in `research.md` §A):

- **14 local files, 175 native-`/goal` occurrences** (13 under `.claude/`, plus the root `CLAUDE.md`).
- **14 template mirrors, 176 occurrences**; the 13 `.claude/**` pairs are byte-identical (`diff -q` → 13 `same`, 0 `DIFF`), while the `CLAUDE.md` pair is intentionally divergent.
- Slash-command wrappers: **14 present**, `goal` **absent** in both trees.
- **8 native-`/goal` emission literals in non-test Go code** (`research.md` §F) — 4 of them inside the auto-injected-resume renderer that the user actually reads.
- **13 public/internal doc files** carrying native-`/goal` emission references (`research.md` §G).

### §A.1 Three-layer scope

| Layer | Phase | Scope |
|---|---|---|
| Doctrine (rules / skills / output-style / root instruction) | run — M1..M4, M6 | 28 existing files + 14 template mirrors |
| Slash-command surface | run — M5 | 2 new files |
| Go emission paths | run — M7 (`cycle_type: tdd`) | 4 files, 8 literals |
| Public + internal docs | **sync** (owner `manager-docs`) | 13 files |

**Tier L is re-confirmed.** The tier rests on three independent factors, any two of which alone would already exceed Tier M: (a) file count — 47 paths across 4 layers; (b) layer heterogeneity — doctrine prose, Go code with a TDD cycle, and 4-locale public docs are three different verification regimes; (c) an irreversibility surface — the Go renderer at `internal/hook/handoff_inject_render.go` is user-visible output with **no existing test**, so M7 must author its own regression guard before changing behaviour.

---

## §B Requirements (GEARS notation)

### §B.1 Emission-surface unification (W1)

- **REQ-GSU-001** (Ubiquitous) — The MoAI doctrine surface shall present `/moai goal` as the single goal-arming surface for every orchestrator emission path.

- **REQ-GSU-002** (Event-driven) — **When** a doctrine surface instructs the orchestrator to emit a goal-arming directive, the doctrine shall specify `/moai goal` rather than native `/goal`.

- **REQ-GSU-003** (Unwanted) — The MoAI pipeline shall not emit a native `/goal` line on any surface, because native `/goal` is HUMAN-ONLY and no tool call can trigger it.

- **REQ-GSU-004** (Capability gate) — **Where** a surface's native-`/goal` reference is a *classification, prohibition rationale, or runtime-interoperation invariant* rather than an emission instruction, that reference shall be retained. Exactly **four** such surfaces exist:

  | Retention surface | Retained content | Why it is not an emission path |
  |---|---|---|
  | `goal-directive.md` § Native `/goal` Prohibition | The prohibition rationale | A prohibition needs its subject |
  | `native-invocation-model.md` § Classification Matrix + Axis B | The HUMAN-ONLY classification | A factual statement about Claude Code; it is the justification for `/moai goal` existing |
  | `docs-site/content/*/claude-code/**` (28 pages) | Documentation of Claude Code's own feature | Native `/goal` genuinely is a Claude Code command; the statements stay true |
  | `internal/goal/evaluate.go` (native-`/goal` yield invariant) | `NativeGoalActive` field, the step-4 yield branch, and the verdict reason string | Implements interoperation *with* native `/goal` (`stop-goal` yields to avoid double-block). Deleting it would remove a safety invariant, not an emission |

- **REQ-GSU-005** (Ubiquitous) — The `goal-directive.md` rule shall be rewritten in place under its existing filename, so that no cross-reference path anywhere in the repository requires updating.

### §B.2 Two-step-mechanism removal (W1 structural)

- **REQ-GSU-006** (Event-driven) — **When** the goal surface becomes orchestrator-armed, the doctrine shall remove the `§ Post-Paste /goal Follow-up Block` structure in full: the section body, its localization-table instruction-line row, its pre-emit self-check item, its cross-references, and the corresponding block in the `moai.md` §8 render surface.

- **REQ-GSU-007** (State-driven) — **While** a pre-emit self-check list states an item count, that stated count shall equal the actual number of checklist items in the list.

### §B.3 Goal-presentation relocation (W2)

- **REQ-GSU-008** (Ubiquitous) — The doctrine shall state that `/moai goal` is **arm-only**: it records goal state and the `stop-goal` hook blocks turn-end until the condition holds, and it starts no work of its own.

- **REQ-GSU-009** (Event-driven) — **When** the orchestrator composes a paste-ready resume message, Block 5 (`Run:`) shall carry the work-starting action (`/moai run SPEC-X`) and shall not carry a bare goal-arming directive as its single primary action, because an arm-only directive alone would spin idle turns until the ceiling with no work to do.

- **REQ-GSU-010** (Event-driven) — **When** the orchestrator reaches the Implementation Kickoff Approval gate (the plan→run human gate), the goal shall be offered as the autonomous-vs-semi-autonomous progression-mode axis inside that gate's `AskUserQuestion`, and the orchestrator shall arm it only after the gate passes.

- **REQ-GSU-011** (Unwanted) — Arming a goal shall not authorize autonomous run-phase entry; the Implementation Kickoff Approval human gate remains required in both progression modes.

- **REQ-GSU-012** (Ubiquitous) — The doctrine shall record the rejected alternative `/moai goal --run SPEC-X "<cond>"` (a composite arm-and-run argument) together with its rejection reason: it would place arming before approval, inverting the gate order, and would require amending the Block 5 single-primary-action constraint.

### §B.4 Slash-command wrapper (W3)

- **REQ-GSU-013** (Ubiquitous) — The `goal` subcommand shall have a `/moai:goal` thin slash-command wrapper, matching the frontmatter shape and body form of the existing sibling wrappers.

- **REQ-GSU-014** (Event-driven) — **When** the wrapper is added, it shall be authored template-source-first (`internal/template/templates/.claude/commands/moai/goal.md.tmpl`) and then embedded via `make build`, so the distributed binary carries it.

- **REQ-GSU-015** (Capability gate) — **Where** the `argument-hint` frontmatter field names the verb surface, it shall match the verbs actually delivered by `workflows/goal.md` § Verbs — the condition form, `status [--all]`, and `clear` (the `resume` verb is documented there as deferred and not delivered).

### §B.5 Template mirror parity and neutrality (W1 propagation)

- **REQ-GSU-016** (State-driven) — **While** a `.claude/**` doctrine file has a template mirror under `internal/template/templates/`, the two shall remain byte-identical after the work, exactly as they are before it.

- **REQ-GSU-017** (Unwanted) — No file under `internal/template/templates/` shall contain this SPEC's ID, its REQ or AC tokens, an internal work date, a commit SHA, or an audit citation.

- **REQ-GSU-018** (Capability gate) — **Where** a local/template pair is intentionally divergent (the root `CLAUDE.md` pair), byte-identity shall not be asserted; instead the specific edited clause shall be verified present and identical on both sides.

### §B.6 Go emission paths (D4 — M7)

- **REQ-GSU-019** (Ubiquitous) — The Go emission paths shall emit `/moai goal`, so the retirement is not inert at the point of user contact.

- **REQ-GSU-020** (Event-driven) — **When** the auto-injected resume renderer builds its restoration-guidance line, it shall name `/moai goal` in every one of its four locale blocks (ko / ja / zh / en).

- **REQ-GSU-021** (State-driven) — **While** the renderer has no regression test, a behaviour change to it shall not be made; M7 shall first author a test asserting the rendered output for all four locales, observe it fail, and only then change the renderer.

- **REQ-GSU-022** (Capability gate) — **Where** the `PrimitiveGoal` manifest token value is changed, the run phase shall first verify that no harness manifest or workflow script declares the old token; **when** any declaration is found, back-compat accepting both tokens shall be required instead of a hard rename.

- **REQ-GSU-023** (Unwanted) — The `moai handoff save --goal` flag **name** shall not be renamed; it is a CLI contract invoked from the session-handoff doctrine. Only its help string changes.

- **REQ-GSU-024** (Unwanted) — The run phase shall not remove the native-`/goal` yield invariant in `internal/goal/evaluate.go`, nor any `/moai goal` implementation identifier (the `internal/goal/` package, the `.moai/state/goal/` path constants, the `stop-goal` hook name, `internal/cli/goal.go`).

### §B.7 Public and internal documentation (D5 — sync phase)

- **REQ-GSU-025** (Event-driven) — **When** the sync phase runs, `manager-docs` shall update the 13 documentation files whose native-`/goal` references are emission-surface renderings of the doctrine this SPEC rewrites.

- **REQ-GSU-026** (Unwanted) — The sync phase shall not modify the 28 `docs-site/content/*/claude-code/**` pages, the `/moai goal`-vs-native factual-contrast pages, or the `.moai/research/` archives.

- **REQ-GSU-027** (Capability gate) — **Where** a page carrying a native-`/goal` reference exists in only some locales, the sync phase shall not create new locale pages to force symmetry; the four-locale obligation binds pages that exist, and a content gap predating this SPEC is not this SPEC's drift to close.

---

## §C Out of Scope

### Out of Scope — native `/goal` behaviour and the Claude Code runtime

- Changing, wrapping, or intercepting how native `/goal` itself behaves. Native `/goal` is a Claude Code built-in; this SPEC only stops the MoAI pipeline from emitting it.
- Adding a `resume` verb to `/moai goal`. `workflows/goal.md` § Verbs already documents it as deferred with a stated reason (`clear` deletes rather than tombstones); reversing that is a separate change.

### Out of Scope — goal-engine implementation

- Any change to the `stop-goal` Stop-hook evaluator's decision logic, the goal state schema at `.moai/state/goal/<session-id>.json`, the turn ceiling, or the stagnation guard. M7 changes emitted *strings* and one manifest *token*, not engine behaviour.
- The native-`/goal` yield invariant in `internal/goal/evaluate.go`. It implements interoperation with the native command and is a retention surface (REQ-GSU-024), not an emission path.
- Introducing a new progression-mode mechanism. The autonomous / semi-autonomous axis and the persisted `progression_mode` field already exist; W2 is codification of when the axis is presented, not new mechanism.

### Out of Scope — sibling command surfaces

- Adding a `design` slash-command wrapper. `design` is deliberately wrapper-less (its workflow file exists but it is not registered in the SKILL.md router — it runs via the `manager-design` agent). The missing-wrapper count is exactly one.
- Retiring or re-pointing any other `/moai` subcommand onto a native bundled skill (the `native-invocation-model.md` Axis A follow-up remains deferred).

### Out of Scope — unrelated doctrine drift

- Any edit to the 106 dirty files on the stale branch in the main checkout. This SPEC executes entirely in the isolated worktree at `origin/main`.
- Renaming `goal-directive.md`. The rewrite is in place under the existing filename by explicit decision (D3).
- Closing the pre-existing four-locale content gap in `docs-site/content/*/advanced/hooks-reference.md` (the page exists in all four locales; only `en` and `ko` carry the reference). Flagged, not fixed (REQ-GSU-027).
- Reconciling the pre-existing verb-surface inconsistency where `SKILL.md` advertises a `resume` verb that `workflows/goal.md` documents as deferred and not delivered. M3 aligns the wrapper's `argument-hint` to the delivered surface; it does not adjudicate the `resume` verb itself.

### Out of Scope — URL-path false positives

- `docs-site/content/*/cli-reference/loop.md` (4 files). A bare `/goal` search matches these, but the match is the link path `/en/cli-reference/goal`, not a command reference. They carry no native-`/goal` command mention and are excluded from every count in this SPEC.

---

## §D Acceptance Criteria

Enumerated in `acceptance.md` (**29** criteria, AC-GSU-001 through AC-GSU-029), each with its judgment command and the verbatim baseline output observed at `origin/main` = `e306e21a9`. AC-GSU-022..027 cover M7 (Go); AC-GSU-028..029 cover the sync-phase doc set.

---

## §E Cross-References

- `plan.md` — milestone decomposition M1..M6 with per-milestone single-owner file assignment.
- `acceptance.md` — AC matrix with recorded baselines.
- `design.md` — doctrine-surface boundary (rules / skills / output-style render surface) and the SSOT-to-render-surface parity relationship.
- `research.md` — baseline measurement commands and observed outputs; the two-command distinction.
- `CLAUDE.local.md` §2 (Template-First), §25 (template internal-content isolation), §27.3 (`/moai:<sub>` wrapper retention).
