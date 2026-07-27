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

> **Scope note.** This is a refactor of an emission surface across three layers — doctrine, Go code, and public docs. There is no new architecture and no change to goal-engine decision logic. This document covers only the structural facts an implementer must respect: which surface owns what, which pairs must stay in lockstep, and where native `/goal` is deliberately retained.

## §A Surface Boundary

The in-scope files sit on **five** distinct surfaces (the public-docs surface left with the scope reduction). The surface determines *what kind of edit* is legitimate, which is why the milestone decomposition follows surface boundaries rather than file counts.

| Surface | Location | Role | Legitimate edit |
|---|---|---|---|
| **Rules** | `.claude/rules/moai/**` | Always-loaded doctrine. Owns definitions, invariants, and prohibitions. | Author or remove normative statements. `goal-directive.md` is the SSOT for goal semantics. |
| **Skills / workflows** | `.claude/skills/moai/**` | Orchestrator-facing procedure. *Consumes* rule definitions; must not restate them. | Repoint citations; switch the emitted directive. Never duplicate a rule's normative text. |
| **Output-style render surface** | `.claude/output-styles/moai/moai.md` §8 | The user-visible rendering of a rule. Carries a compact pointer, not a copy. | Keep it a pointer consistent with its SSOT. |
| **Root instruction** | `CLAUDE.md` | Always-loaded execution directive. One-line summaries only. | Single-clause edits. |
| **Go emission paths** | `internal/hook`, `internal/harness/v4manifest`, `internal/cli` | Code that *materializes* the directive at runtime — a rendered string, a manifest token, a flag help line. | Change the emitted string / token. Behaviour changes require a test first. |
| **Public docs** | `docs-site/content/{en,ja,ko,zh}/**`, `.moai/docs/**` | The externally-published rendering of the doctrine. | Sync-phase only; four-locale parity for pages that exist. |

### §A.1 Why the code layer is where the retirement actually bites

The doctrine layer is *declarative*: it tells the orchestrator what to emit. The Go layer is *imperative*: it emits. `internal/hook/handoff_inject_render.go` builds the auto-injected resume that the user reads at session start, and its four locale blocks each hard-code `  • /goal `. A retirement that stopped at the doctrine layer would leave the pipeline literally printing the retired command in four languages while the rules asserted it never does — the failure mode the D4 expansion exists to prevent.

This is also why M7 carries a different verification regime. Doctrine edits are verified by grep; a user-visible rendered string is verified by asserting the rendered output. Since `handoff_inject_render.go` has no test at all, M7 must author that assertion first (RED), observe it fail, and only then change the renderer — hence `cycle_type: tdd` for M7 alone.

Two consequences bind the implementation:

1. **`goal-directive.md` is the SSOT and must land first.** The skills surface cites it by heading name. M1 creates `## Native \`/goal\` Prohibition` and `### Goal-Presentation Timing`; M2-M4 cite those exact names. Reversing the order produces citations to headings that do not yet exist.

2. **The render surface is a pointer, never a second copy.** `session-handoff.md` explicitly designates itself the SSOT and `moai.md` §8 the render surface, with a bidirectional drift-mitigation sentinel. When M2 removes the two-step mechanism from the SSOT, the render surface's compact clause must be removed in the same milestone — which is why both files have a single owner (M2).

## §B SSOT-to-Render-Surface Parity

Two parity relationships operate on different axes and must not be conflated.

### §B.1 Doctrinal parity (SSOT → render surface)

`session-handoff.md` (SSOT) ↔ `moai.md` §8 (render surface). Parity here is *semantic*, not byte-level: the render surface deliberately compresses. The SSOT's own § Cross-references names the four things that must stay consistent — the Localization Table, the 6-block skeleton, the cut-line marker spec, and the pre-emit self-check labels. Two of the four are touched by this SPEC (the Localization Table's instruction-line row, and the self-check item and its stated count), so both surfaces move together in M2.

Because the self-check counts are stated inline and differ per surface (10 in the SSOT, 12 in the render surface — the render surface's list is a superset), removing one item requires two independent decrements: 10 → 9 and 12 → 11. These are separate ACs (AC-GSU-008, AC-GSU-011) precisely because a single shared decrement would be wrong.

### §B.2 Distribution parity (local ↔ template mirror)

`.claude/**` ↔ `internal/template/templates/.claude/**`. Parity here **is** byte-level and CI-enforced (`internal/template/rule_template_mirror_test.go`). 13 of the 14 pairs are byte-identical today and must remain so.

The 14th pair is the exception that shapes M6: the root `CLAUDE.md` and its mirror are **intentionally divergent**, because the template copy is neutralized per the template internal-content isolation policy. Asserting byte-identity there would be wrong. M6 therefore runs two different checks — a byte-parity loop over 13 pairs, and a single-clause equality check on the 14th.

This asymmetry is the reason the delegation brief's `.claude/`-scoped file discovery missed the pair: the local half lives at the repository root, outside `.claude/`.

## §B.3 The Three Retention Surfaces

A single test separates an emission path from a retention surface: **does the sentence become false when MoAI stops emitting native `/goal`?** If it stays true, it is a retention surface.

| # | Surface | Layer | Sentence it makes | Still true after retirement? |
|---|---|---|---|---|
| 1 | `goal-directive.md` § Native `/goal` Prohibition | doctrine | "the pipeline does not emit native `/goal`, because it is HUMAN-ONLY" | Yes — it *is* the retirement |
| 2 | `native-invocation-model.md` Classification Matrix | doctrine | "native `/goal` is HUMAN-ONLY, therefore Axis B justifies `/moai goal`" | Yes — a fact about Claude Code |
| 3 | `internal/goal/evaluate.go` yield invariant | Go | "when the runtime signals an active native `/goal`, `stop-goal` yields" | Yes — interoperation, not emission |

Three further retention surfaces exist at the **documentation** layer — `docs-site/content/*/claude-code/**`, the `autonomous-loops.md` native sections, and `.moai/docs/autonomous-workflow-strategy.md`. All three left with the sync-phase scope when it was split to `SPEC-GOAL-DOCS-RETIRE-001` (plan-audit iteration 2), and are registered there. The membership test below is identical in both SPECs; only the layer differs.

Row 3 is the one most at risk from a mechanical sweep, and it was mis-classified in the D4 brief as "implementation, not the native command". It is in fact *about* the native command: `NativeGoalActive` exists precisely so a user who typed native `/goal` does not get double-blocked by MoAI's evaluator as well. Deleting those lines would not tidy a stale reference — it would remove a safety invariant (`workflows/goal.md` § Safety Invariants #4). AC-GSU-027 pins them.

A second failure mode, discovered at the documentation layer and recorded here because it generalizes: **a file can be partly emission and partly retention.** `autonomous-loops.md` is such a *split* surface — its MoAI-primitive listings are emission, its native-`/goal` sections are retention. A file-level affected/retain classification cannot express that. The handling of it moved to `SPEC-GOAL-DOCS-RETIRE-001`, but the lesson binds any future sweep here too.

The practical consequence for the run phase: **no AC may be phrased as "zero native-`/goal` in layer X"** without an explicit retention carve-out, and **no file may be classified affected-or-retain without checking whether it is split**. Every sweep AC in `acceptance.md` names its file list positively rather than negating a layer, and each was tested against all three rows before its baseline was recorded.

## §C Why the Two-Step Mechanism Becomes Dead Code

The `§ Post-Paste /goal Follow-up Block` is a workaround whose entire justification is a property of native `/goal`:

- native `/goal` is HUMAN-ONLY → the model cannot type it;
- slash commands parse only at input start → a line inside a pasted body is inert;
- therefore the user must send it as a **separate standalone message**, and the resumed orchestrator carries a *reminder obligation* to tell them to.

Every clause in that chain is a consequence of the human-only constraint. `/moai goal` is orchestrator-armed, so all three premises evaporate at once. The section is not merely redundant after W1 — it describes a problem that no longer exists, which is why M2 removes the structure rather than rewording it.

The same reasoning retires the dependent surfaces: the Localization Table instruction-line row (it localizes an instruction no longer given), the self-check item (it checks for a block no longer emitted), and the render-surface clause (it renders a removed section).

## §D Why W2 Needs No New Mechanism

The progression-mode axis already exists and is already wired to the right gate:

- `workflows/goal.md` § Progression Mode already places the autonomous / semi-autonomous choice **at Implementation Kickoff Approval**;
- goal state already persists `progression_mode`;
- the `stop-goal` hook already emits the semi-autonomous checkpoint JSON for orchestrator-side confirmation.

W2 is therefore codification, not construction — net new development is zero. What is missing is a *statement of relationship*: `session-handoff.md` says Block 5 MAY carry a `/moai goal` directive but never says what that means next to `/moai run`. Because `/moai goal` is arm-only, a Block 5 line carrying only the goal would arm a condition with no work running and spin idle turns to the ceiling.

The rule that closes the gap therefore has two halves, deliberately placed on the two surfaces that own them:

- the **arm-only property and the Block 5 consequence** go in `session-handoff.md` (M2 — it owns Block 5);
- the **presentation timing** goes in `goal-directive.md` § Goal-Presentation Timing (M1 — it is the goal SSOT).

They are linked bidirectionally (AC-GSU-010 forward, AC-GSU-014 backward) so a reader arriving from either side finds the other half. Bidirectional linkage is verified rather than assumed because a one-way reference leaves the rule undiscoverable from the goal side, which is the failure mode that produced the gap in the first place.

## §E Rejected Alternative

`/moai goal --run SPEC-X "<condition>"` — a composite argument that arms the goal and starts the run in one invocation. Rejected on two independent grounds:

1. **It inverts the gate order.** Arming would occur before Implementation Kickoff Approval, and setting a goal starts a turn immediately — so run-phase work would begin before the human gate. This contradicts the Kickoff-remains-required invariant.
2. **It would require amending a constraint.** `session-handoff.md` § Diet Constraints caps Block 5 at a single primary action; a composite arm-and-run directive either violates that cap or forces it to be rewritten.

Recorded in `goal-directive.md` § Goal-Presentation Timing per REQ-GSU-012 so the reasoning survives and the option is not silently re-proposed.
