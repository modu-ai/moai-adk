---
id: SPEC-GOAL-SURFACE-UNIFY-001
title: Unify the goal surface on /moai goal and relocate goal presentation to the Implementation Kickoff Approval gate
version: 1.3.0
status: draft
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: HIGH
phase: "v3.1.0"
module: doctrine
lifecycle: spec-anchored
tags: "goal, doctrine, session-handoff, slash-command, template-mirror"
tier: L
---

## HISTORY

| Version | Date | Change | Author |
|---------|------|--------|--------|
| 1.0.0 | 2026-07-25 | Initial plan-phase authoring (Tier L, 6 milestones) | manager-spec |
| 1.1.0 | 2026-07-25 | Scope expansion per approved decisions D4 (Go emission paths → M7, `cycle_type: tdd`) and D5 (public docs → sync phase). Scope now spans three layers: doctrine, Go code, public docs. | manager-spec |
| 1.2.0 | 2026-07-25 | Plan-audit iteration 1 remediation (D1-D6 + material SHOULD-FIX): frontmatter `tags` string form, retention register, M7 test-fixture ownership, split-surface reclassification, union detector, REQ↔AC matrix. | manager-spec |
| 1.3.0 | 2026-07-25 | Plan-audit iteration 2 returned FAIL 0.64 with STOP (score regression). **Scope reduction executed**: public-documentation scope split to `SPEC-GOAL-DOCS-RETIRE-001`. N1-N5 closed; retention register re-derived to three surfaces; identifiers deliberately not renumbered (§B.7). | manager-spec |

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
- **13 public/internal doc files** carrying native-`/goal` emission references — **split out** to `SPEC-GOAL-DOCS-RETIRE-001` (§B.7).

### §A.1 Four-layer scope (post-split)

| Layer | Phase | Paths |
|---|---|---|
| Doctrine — local (rules / skills / output-style / root instruction) | run — M1..M4 | 15 |
| Doctrine — template mirrors | run — M6 | 15 |
| Slash-command surface | run — M5 | 2 new |
| Go emission paths | run — M7 (`cycle_type: tdd`) | 5 files, 8 literals |
| **Canonical total** | | **37** |

The per-milestone breakdown is `plan.md` §F.1, which is the arithmetic SSOT for this table.

The public-documentation layer (13 paths) was split out to `SPEC-GOAL-DOCS-RETIRE-001` — see §B.7.

**Tier L is re-confirmed after the split.** The tier rests on three independent factors, any two of which alone would already exceed Tier M: (a) file count — 37 paths across four layers, against Tier M's 15-file ceiling; (b) layer heterogeneity — doctrine prose, Go code with a TDD cycle, and byte-identical template mirrors are three different verification regimes; (c) an irreversibility surface — the Go renderer at `internal/hook/handoff_inject_render.go` is user-visible output with **no existing test**, so M7 must author its own regression guard before changing behaviour.

---

## §B Requirements (GEARS notation)

### §B.1 Emission-surface unification (W1)

- **REQ-GSU-001** (Ubiquitous) — The MoAI doctrine surface shall present `/moai goal` as the single goal-arming surface for every orchestrator emission path.

- **REQ-GSU-002** (Event-driven) — **When** a doctrine surface instructs the orchestrator to emit a goal-arming directive, the doctrine shall specify `/moai goal` rather than native `/goal`.

- **REQ-GSU-003** (Unwanted) — The MoAI pipeline shall not emit a native `/goal` line on any surface, because native `/goal` is HUMAN-ONLY and no tool call can trigger it.

- **REQ-GSU-004** (Capability gate) — **Where** a surface's native-`/goal` reference is a *classification, prohibition rationale, or runtime-interoperation invariant* rather than an emission instruction, that reference shall be retained. The authoritative membership list is the retention register at `plan.md` §A.2 — this requirement binds to that register rather than restating it, so the two cannot drift apart. Within this SPEC's post-split scope (doctrine + Go) the register holds **three** surfaces:

  | # | Retention surface | Layer | Why it is not an emission path |
  |---|---|---|---|
  | 1 | `goal-directive.md` § Native `/goal` Prohibition | doctrine | A prohibition needs its subject |
  | 2 | `native-invocation-model.md` § Classification Matrix + Axis B | doctrine | A factual statement about Claude Code; it is the justification for `/moai goal` existing |
  | 3 | `internal/goal/evaluate.go` (native-`/goal` yield invariant) | Go | Implements interoperation *with* native `/goal` (`stop-goal` yields to avoid double-block). Deleting it would remove a safety invariant, not an emission |

  The three documentation-layer retention surfaces — `docs-site/content/*/claude-code/**`, the `autonomous-loops.md` native sections, and `.moai/docs/autonomous-workflow-strategy.md` — moved to `SPEC-GOAL-DOCS-RETIRE-001` with the sync-phase scope. They are registered in that SPEC's own retention register, not here.

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

- **REQ-GSU-028** (Event-driven) — **When** M7 changes the `PrimitiveGoal` token value or the runner-template dispatch arm, it shall update `internal/harness/v4manifest/runner_template_test.go` in the same change, so the `v4manifest` package suite continues to exit 0.

### §B.7 Moved-identifier register (scope reduction, plan-audit iteration 2)

The public-documentation scope was split out to **`SPEC-GOAL-DOCS-RETIRE-001`** after plan-audit iteration 2 emitted STOP. Identifiers are **NOT renumbered** here: 62 judgment-command baselines (29 at iteration 1 + 33 at iteration 2) reproduced verbatim against these identifiers across two independent audits, and renumbering would discard that audit trail. The gaps below are deliberate.

| Vacated here | Destination | Subject |
|---|---|---|
| REQ-GSU-025 | `SPEC-GOAL-DOCS-RETIRE-001` REQ-GDR-001 | sync phase updates the affected doc set |
| REQ-GSU-026 | `SPEC-GOAL-DOCS-RETIRE-001` REQ-GDR-002 | sync phase retains CC pages / contrast pages / archives |
| REQ-GSU-027 | `SPEC-GOAL-DOCS-RETIRE-001` REQ-GDR-003 | no locale-symmetry manufacture |
| AC-GSU-028 | `SPEC-GOAL-DOCS-RETIRE-001` AC-GDR-001..005 | split-surface emission markers + retention pins (re-anchored locale-invariantly) |
| AC-GSU-029 | `SPEC-GOAL-DOCS-RETIRE-001` AC-GDR-006 | sync retention pins |
| AC-GSU-032 | `SPEC-GOAL-DOCS-RETIRE-001` AC-GDR-007 | strategy-record superseding note |

Retained identifier set in this SPEC: **REQ-GSU-001..024 + 028** (25 requirements) and **AC-GSU-001..027 + 030, 031, 033** (30 criteria). Numbering is non-contiguous by design.

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

### Out of Scope — historical records

- `.moai/specs/**` historical SPEC artifacts (~60 files carrying native-`/goal`). A completed SPEC body is an immutable record of what was decided at the time; retroactively rewriting it would falsify the record. Same rationale as the `.moai/research/` archives — stated explicitly here rather than left to silence.
### Out of Scope — public and internal documentation (split to SPEC-GOAL-DOCS-RETIRE-001)

- `docs-site/content/{en,ja,ko,zh}/**` in full, and `.moai/docs/autonomous-workflow-strategy.md`. The 13-file affected set and the documentation-layer retention surfaces are owned by `SPEC-GOAL-DOCS-RETIRE-001`, which depends on this SPEC. See §B.7.

---

## §D Acceptance Criteria

Enumerated in `acceptance.md` (**30** criteria — AC-GSU-001..027 plus 030, 031, 033; the gaps at 028/029/032 are the moved identifiers registered in §B.7), each with its judgment command and the verbatim baseline output observed in the worktree. AC-GSU-022..027 and AC-GSU-030 cover M7 (Go); AC-GSU-031 is the union-detector sweep; AC-GSU-033 makes the M7 RED ordering auditable.

A **REQ ↔ AC traceability matrix** is at `acceptance.md` §F: all 25 REQs are cited by at least one AC, and every AC appears in at least one row.

---

## §E Cross-References

- `plan.md` — milestone decomposition M1..M6 with per-milestone single-owner file assignment.
- `acceptance.md` — AC matrix with recorded baselines.
- `design.md` — doctrine-surface boundary (rules / skills / output-style render surface) and the SSOT-to-render-surface parity relationship.
- `research.md` — baseline measurement commands and observed outputs; the two-command distinction.
- `CLAUDE.local.md` §2 (Template-First), §25 (template internal-content isolation), §27.3 (`/moai:<sub>` wrapper retention).
