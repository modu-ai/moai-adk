---
id: SPEC-GOAL-DOCS-RETIRE-001
title: Retire native /goal emission references from public and internal documentation across four locales
version: 1.0.0
status: draft
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

## §A Context

Sync-phase documentation scope, owner `manager-docs`. Split from `SPEC-GOAL-SURFACE-UNIFY-001` after its plan-audit iteration 2 emitted STOP.

### §A.1 Tier M — rationale

**Tier M**, not L. The factors:

| Factor | Assessment |
|---|---|
| File count | 13 — inside Tier M's 15-file ceiling |
| Layers | One (documentation). No code, no tests, no template mirrors |
| Verification regimes | One (locale-symmetric grep). Repeating a grep per locale is the same regime four times, not four regimes |
| Reversibility | Fully reversible. Prose edits, reviewable in a diff, no runtime surface |
| Irreversibility surface | None. The parent's Tier L rested partly on an untested user-visible Go renderer; nothing here has that shape |

The genuinely hard part — the split-surface line running inside pages, and locale-invariance — is handled by **detector design at plan time**, not by added implementation layers. That difficulty justifies its own SPEC and its own audit; it does not by itself lift the tier. Tier M with 12 criteria, four of them dedicated locale-parity guards, is the proportionate envelope.

### §A.2 Retention register — four retained documentation surfaces

This is this SPEC's **own** register, not inherited. `spec.md` REQ-GDR-005 binds to it by reference. The membership test: **does the sentence become false when MoAI stops emitting native `/goal`?** If it stays true, it is retention.

| # | Surface | Files | Retained content | Why | Guard |
|---|---|---:|---|---|---|
| 1 | `docs-site/content/*/claude-code/**` | 28 | All native-`/goal` documentation | Documents Claude Code's own feature; MoAI's emission change does not falsify it | AC-GDR-006 |
| 2 | `docs-site/content/*/advanced/autonomous-loops.md` — **split surface** | 4 | The `### \`/goal\` — native Claude Code (HUMAN-ONLY)` H3, the `Native /goal Details` H2, the primitive-comparison table row, and the factual statements at L37/L45/L49/L83/L92 | The page exists to *distinguish* the three primitives. L49 is verbatim the same Axis-B justification for which surface 3 is retained | AC-GDR-005 (per-locale pins) |
| 3 | `docs-site/content/*/cli-reference/goal.md` + `*/utility-commands/moai-goal.md` | 8 | The programmatic-counterpart contrast | States why `/moai goal` exists — deleting it removes the justification | AC-GDR-006 |
| 4 | `.moai/docs/autonomous-workflow-strategy.md` | 1 | The whole document | A superseded **strategy record**: it proposed a 3-engine model with native `/goal` as one engine. Retirement makes it historically superseded, not factually false. Self-declares `> **상태**: 전략 제안`; zero referrers; not user-distributed | AC-GDR-007 (superseding note, not a sweep) |

`docs-site/content/*/advanced/hooks-reference.md` (2 files, en + ko) is also retained — its Stop-hook dual mention is consistent with the parent's retained yield invariant — and is pinned by AC-GDR-006 rather than given its own row, since it needs no structural protection.

**Why `.moai/docs/autonomous-workflow-strategy.md` is here and not in the parent.** Judged and recorded per the delegation instruction. It is an internal document, so a "public docs" reading would place it in the parent. Three considerations put it here instead: (a) its **treatment** is a documentation-maintenance action — add a superseding note, sweep nothing — which is sync-phase in nature, and the parent has no sync phase after the split; (b) its **content** is the same 3-engine comparison as `autonomous-loops.md`, so grouping them keeps "which documents describe the three engines" in one place and one reviewer's head; (c) its **membership test outcome** is identical to the other retention rows here. The counter-argument — that `.moai/docs/` is not public — is real but weaker: the retain-plus-note treatment does not depend on publication.

---

## §B Known Issues / Risks

- **B1 — A prose-anchored detector can read zero while content survives.** This is the parent's audit finding N2, and it is the single defect class this SPEC exists to prevent. `per-turn` → `ターンごとの` / `턴별` / `每回合`, so an English-anchored regex read `en:2 ja:0 ko:0 zh:0` against content present in all four locales. Every detector here is anchored on a code literal or an already-untranslated string, and every one carries a per-locale baseline.
- **B2 — A split surface cannot be judged by a file-level target.** `autonomous-loops.md` must lose 6 markers and keep 3 structures. A blanket `0` would delete two sections, a table row, and the Axis-B justification, four times over.
- **B3 — Four-locale parity is a HARD project obligation** (`CLAUDE.local.md` §17). An edit landing in `en` only is drift, not partial progress.
- **B4 — The parent's M7 must land first for `handoff.md`.** The `--goal` help-string wording this SPEC mirrors is decided there (§A.2 of `spec.md`).
- **B5 — `hooks-reference.md`'s locale asymmetry is pre-existing.** The page exists in all four locales; only `en` and `ko` carry the reference. Do not create `ja`/`zh` content to force symmetry (REQ-GDR-008); the AC-GDR-006 pin at `2` fails if lines are added.
- **B6 — `.moai/docs/` is not template-distributed**, so no mirror-parity obligation applies to surface 4. Verified: `ls internal/template/templates/.moai/docs/` does not carry it.

---

## §C Pre-flight

```bash
# Parent dependency landed?
grep -c 'a /moai goal condition' internal/cli/handoff.go        # expect 1 once parent M7 lands
# Locale file set intact?
ls docs-site/content/{en,ja,ko,zh}/advanced/autonomous-loops.md | wc -l   # expect 4
ls docs-site/content/{en,ja,ko,zh}/cli-reference/handoff.md | wc -l       # expect 4
ls docs-site/content/{en,ja,ko,zh}/advanced/self-evolving.md | wc -l      # expect 4
```

---

## §D Constraints

- **D1 [HARD]** Every emission detector anchors on a translation-surviving token; every emission baseline is recorded per locale and is symmetric (REQ-GDR-003, REQ-GDR-004).
- **D2 [HARD]** Exactly the four surfaces in §A.2 are retained. Test every sweep-style criterion against each of the four before recording it.
- **D3 [HARD]** No new locale pages (REQ-GDR-008).
- **D4 [HARD]** No edit outside the 13 scoped files. `docs-site/content/*/claude-code/**` and `.moai/research/` are untouched.
- **D5** Match each page's existing register and locale conventions; this is a wording change, not a restructure.
- **D6** No time estimates.

---

## §E Self-Verification

- **Held-in** — every criterion transitions from its recorded baseline to its target.
- **Held-out** — the docs-site build stays green (AC-GDR-012) and the four-locale file inventory is unchanged.

---

## §F Milestones

Ordered by dependency. The split surface leads, because its retain/sweep line is the decision most likely to need review.

### N1 — `autonomous-loops.md` split surface (4 files)

**Priority High.** The hardest judgment in the SPEC and the one the parent got wrong.

Owns 4 files: `docs-site/content/{en,ja,ko,zh}/advanced/autonomous-loops.md`

Sweep (6 markers, 1-2 per locale):
- the `` `/moai loop` / `/goal` `` paired-primitive listing (L99)
- the auto-mode / `ac_converge` pairing (L94) — the marker the parent's detector could not see in 3 locales
- L7's "MoAI-ADK provides three continuation-loop primitives … `/goal`" mis-attribution — reword so Claude Code provides `/goal` and MoAI-ADK provides `/moai goal` and `/moai loop`

Retain (register row 2): the native H3, the `Native /goal Details` H2, the comparison row, and L37/L45/L49/L83/L92.

ACs: AC-GDR-001, AC-GDR-003, AC-GDR-004, AC-GDR-005, AC-GDR-008.

### N2 — `self-evolving.md` (4 files)

**Priority Medium.** Owns `docs-site/content/{en,ja,ko,zh}/advanced/self-evolving.md`. Both references (L40, L97, all four locales) name `/goal` as a MoAI convergence primitive whose trajectories the routing ledger records; they become `/moai goal`.

ACs: AC-GDR-002, AC-GDR-009.

### N3 — `handoff.md` (4 files)

**Priority Medium.** Owns `docs-site/content/{en,ja,ko,zh}/cli-reference/handoff.md`. The `--goal` row's description mirrors the parent's M7 help string. **Blocked on the parent** (B4): the replacement wording is decided there.

ACs: AC-GDR-003, AC-GDR-010.

### N4 — `.moai/docs/autonomous-workflow-strategy.md` (1 file)

**Priority Low.** Owns the strategy record. Adds a one-line superseding note near the existing status banner; sweeps nothing. All 25 occurrences preserved.

ACs: AC-GDR-007.

### N5 — Retention verification (0 files edited)

**Priority High**, runs last. Edits nothing; verifies the four retention surfaces are untouched and the locale-parity criteria hold.

ACs: AC-GDR-006, AC-GDR-011, AC-GDR-012.

### §F.1 Ownership map (no file owned twice)

| Milestone | Paths | Composition |
|---|---:|---|
| N1 | 4 | `advanced/autonomous-loops.md` ×4 locales |
| N2 | 4 | `advanced/self-evolving.md` ×4 locales |
| N3 | 4 | `cli-reference/handoff.md` ×4 locales |
| N4 | 1 | `.moai/docs/autonomous-workflow-strategy.md` |
| N5 | 0 | verification only |
| **Total** | **13** | all sync-phase |

4+4+4+1 = 13. Every path appears in exactly one row. The 41 retained files (28 `claude-code` + 8 contrast + 2 `hooks-reference` + 3 `.moai/research`) are edited by no milestone.

---

## §G Anti-Patterns

- **Anchoring an emission detector on translatable prose.** The N2 defect class. `per-turn`, "continuation", "per-tool" are all translated; `` `/goal` ``, `ac_converge`, `auto mode`, and `--goal` are not.
- **Sweeping `autonomous-loops.md` to zero.** Deletes register row 2 four times over.
- **Editing `en` and deferring the other three locales.** Violates B3; AC-GDR-008..011 fail asymmetrically by construction.
- **Rewriting the strategy record instead of annotating it.** AC-GDR-007's content pin at `25` blocks it.
- **Creating `ja`/`zh` `hooks-reference.md` content** to make the pin symmetric. The pin is at `2` precisely so this fails.
- **Touching `claude-code/**`.** AC-GDR-006 pins it at `80`.

---

## §H Cross-References

- `spec.md` §B — REQ-GDR-001..008.
- `acceptance.md` — 12 criteria with per-locale baselines, plus the REQ↔AC matrix (§E).
- `research.md` — the documentation measurement and per-detector locale-symmetry proof.
- `SPEC-GOAL-SURFACE-UNIFY-001` — the parent (`depends_on`).
- `CLAUDE.local.md` §17 · `.moai/docs/docs-site-i18n-rules.md`.
