---
id: SPEC-GOAL-DOCS-RETIRE-001
title: Retire native /goal emission references from public and internal documentation across four locales
version: 1.0.0
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: MEDIUM
phase: "v3.1.0"
module: docs
lifecycle: spec-anchored
tags: "goal, docs-site, i18n, locale-parity, split-surface"
tier: M
depends_on: [SPEC-GOAL-SURFACE-UNIFY-001]
---

> Thin by design. This is a documentation refactor with no new architecture. It covers the two structural properties that shape the criteria: how a split surface is judged, and why detector locale-invariance is a design constraint rather than a review checklist item.

## §A Split-Surface Judgment

A file is a **split surface** when the retain/sweep line runs *inside* it. `autonomous-loops.md` is the case: its purpose is to distinguish native `/goal`, `/moai goal`, and `/moai loop`, so it necessarily carries native-`/goal` content that must survive — while also carrying MoAI-primitive listings that must go.

A file-level affected/retain classification cannot express this, and that is exactly how the parent SPEC produced a destructive criterion: classifying the file "affected" and targeting `0` would have deleted two sections, a comparison-table row, and the Axis-B justification, in four locales.

The judgment form used here is **positive on both sides**:

| Half | Form | Why not the alternative |
|---|---|---|
| Sweep | Name each emission marker and target `0` | A file-level `0` cannot distinguish emission from retained content |
| Retain | Pin each retained structure at its exact count | Absence of a sweep is not evidence of preservation |

Neither half is falsifiable alone: the sweep half fails at baseline but says nothing about what survived; the pin half holds at baseline and says nothing about what was removed. They are therefore **compounded** — every retention criterion here references the aggregate emission criterion (AC-GDR-012), so "swept the emission AND kept the structure" is a single verifiable claim.

## §B Locale-Invariance as a Design Constraint

The membership test for retention is inherited from the parent: *does the sentence become false when MoAI stops emitting native `/goal`?* What is **new** at this layer is that the same content exists four times, and a detector can see some copies and not others.

### §B.1 The failure mode

A detector anchored on translatable prose reads a false zero. Measured on this corpus: an anchor on `per-turn` reads `en:2 ja:0 ko:0 zh:0` against content present in all four locales, because `per-turn` becomes `ターンごとの` / `턴별` / `每回合`. An acceptance criterion built on it goes green while three locales ship the retired content — and it does so **silently**, since the aggregate is `0`.

This is not a measurement error. The measurement was correct; the *detector* was incomplete. That distinction matters, because it means review of the recorded number cannot catch it — only running the detector per locale can.

### §B.2 The two-layer guard

Because per-detector diligence is not self-verifying, the criteria carry two layers:

1. **Per-locale targets** on every emission criterion (AC-GDR-001..005). A `0` in one locale is not a pass.
2. **A meta-guard** (AC-GDR-010) whose subject is the detectors themselves: each must yield exactly ONE distinct value across the four locales. A prose-anchored detector produces `{2, 0}` — two distinct values — and fails here even when its own aggregate reads `0`.

Layer 2 is what makes the N2 defect class *mechanically* impossible rather than merely documented. It is the only criterion in either SPEC whose subject is a detector rather than the corpus.

### §B.3 Anchor selection rule

An anchor is admissible when it survives translation. On this corpus:

- **Admissible** — backticked code spans (`` `/goal` ``, `` `/moai loop` ``), code identifiers (`ac_converge`), CLI flags (`--goal`), structural markers (`^## `, `^| `, `^### `), and strings already untranslated in all four locales (`auto mode`, verified `1 1 1 1`).
- **Inadmissible** — any translatable prose. `per-turn`, "continuation", "per-tool" are all localized.

Where an anchor's untranslated status is asserted, it is measured rather than assumed: `research.md` §C.2 carries the per-locale count for each.

## §C Dependency Boundary with the Parent

One file couples the two SPECs: `cli-reference/handoff.md` documents the `--goal` flag whose help string the parent's M7 rewrites. The coupling is one-directional and narrow:

- The parent decides the **replacement wording** (`internal/cli/handoff.go:104`).
- This SPEC mirrors it into four locale pages.
- The flag **name** changes in neither — it is a CLI contract invoked from the session-handoff doctrine.

Hence `depends_on: [SPEC-GOAL-SURFACE-UNIFY-001]` and the N3 milestone's blocked status, with the dependency expressed as a held-out gate (`acceptance.md` §C) rather than an acceptance criterion: it is a precondition on *when* N3 can be specified, not a property of this SPEC's output.

Nothing else couples them. The doctrine, template-mirror, wrapper, and Go layers are entirely the parent's; the retention registers are separate documents with no shared rows.
