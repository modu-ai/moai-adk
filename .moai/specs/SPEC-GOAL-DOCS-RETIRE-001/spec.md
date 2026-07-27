---
id: SPEC-GOAL-DOCS-RETIRE-001
title: Retire native /goal emission references from public and internal documentation across four locales
version: 1.3.0
status: in-progress
created: 2026-07-25
updated: 2026-07-27
author: manager-spec
priority: MEDIUM
phase: "v3.1.0"
module: docs
lifecycle: spec-anchored
tags: "goal, docs-site, i18n, locale-parity, split-surface"
tier: M
depends_on: [SPEC-GOAL-SURFACE-UNIFY-001]
---

## HISTORY

| Version | Date | Change | Author |
|---------|------|--------|--------|
| 1.0.0 | 2026-07-25 | Initial authoring. Split from `SPEC-GOAL-SURFACE-UNIFY-001` after its plan-audit iteration 2 emitted STOP (score regression 0.71 → 0.64) and the user chose scope reduction over iteration 3. Carries the public-documentation scope with locale-invariant detectors. | manager-spec |
| 1.1.0 | 2026-07-27 | Plan-audit iteration 2 MUST-FIX **B2-1** (aptness) closed. New `REQ-GDR-011` requires every emission detector's pattern to carry a literal `/goal` token, declared once and shared by the counting step and the assertion; `AC-GDR-010` gains component **(d)** and re-records its baseline with an `apt` field. The auditor's stronger alternative — a base-match-set subset check — was executed and **refuted**: `ac_converge` and `auto mode` co-occur with `` `/goal` `` on the same base line, so a line-level subset test admits both attack detectors. | orchestrator |
| 1.2.0 | 2026-07-27 | Run-gate plan-audit (PASS 0.849) finding **D1** closed. The sweep-target file count is corrected `8 → 12` and the retained-within-scope count `5 → 1` (annotate-only), reconciling `spec.md` §A.3 and `REQ-GDR-001` with `plan.md` §F.1's ownership map (`autonomous-loops.md` ×4 + `self-evolving.md` ×4 + `handoff.md` ×4 + strategy ×1 = 13). The same correction is applied to `acceptance.md` AC-GDR-012's header and to `progress.md` §F's scope row, which had recorded both figures side by side without reconciling them. Iteration 2's finding B-4 corrected the adjacent **marker** count `18 → 24` but did not question the **file** partition. A fourth site the D1 block did not enumerate — `research.md` §C's aggregate sentence — carried the same `8` and was self-contradictory (its own decomposition includes `handoff 4`); it is corrected with the rest, keeping the `24` and the decomposition verbatim. Prose-only: no judgment command, recorded baseline, or target value changed. | manager-spec |
| 1.3.0 | 2026-07-27 | `AC-GDR-007`'s content-pin detector corrected to exclude the mandated sentinel line, closing an off-by-one internal contradiction found during run-phase N5 verification. The criterion is compound: component 1 mandates the literal sentinel ``native `/goal` emission is retired``, and that phrase itself carries a backticked `` `/goal `` that component 2's occurrence count picked up — so satisfying component 1 necessarily drove the pin from `25` to `26`, making the criterion unsatisfiable as written. The second detector now filters the sentinel line out before counting (`grep -vF … \| grep -ohF … \| wc -l`). The pin value (`25`) and the recorded baseline (`0` / `25`) are **unchanged**, as is the criterion's semantics ("all 25 historical occurrences survive"); the raise-to-`26` alternative was declined because `26` is a composite that obscures what the pin asserts. The implementation was already correct — `git diff origin/main...HEAD` on the strategy record is `1 insertion, 0 deletions`. | manager-spec |

---

## §A Context

`SPEC-GOAL-SURFACE-UNIFY-001` retires native `/goal` from every MoAI **emission** path in doctrine and Go code, replacing it with `/moai goal`. Native `/goal` is a Claude Code built-in TUI command and is HUMAN-ONLY — the model cannot invoke it — which is precisely why MoAI reimplemented it programmatically as `/moai goal`.

This SPEC carries the **public and internal documentation** consequence of that retirement, in the sync phase, owned by `manager-docs`.

### §A.1 Why this is its own SPEC

The parent SPEC's plan-audit iteration 2 found that the documentation half carries a structurally different problem from the doctrine half, and that folding it into a Tier L doctrine SPEC produced a defect the parent's own acceptance criteria could not see:

> AC-GSU-028's a3 emission detector anchored on the English string `per-turn`. Per-locale it read `en: 2 · ja: 0 · ko: 0 · zh: 0` — yet the same content exists in all four locales at line 94, with only the trigger word translated (`ターンごとの` / `턴별` / `每回合`). Sweeping English alone would have driven the detector to `0` while three locales shipped the retired emission content.

Two properties make this half hard, and neither appears in the parent:

1. **The retain/sweep line runs *inside* individual pages.** `autonomous-loops.md` exists to *distinguish* the three continuation primitives, so it carries a dedicated native-`/goal` section that must survive alongside MoAI-primitive listings that must go. A file-level affected/retain classification cannot express this.
2. **Every detector must be locale-invariant.** A prose-anchored regex can read zero in three locales while the content is present. Detectors must anchor on code literals or structural markers that survive translation, and each must be proven symmetric by measurement — not by inspection.

### §A.2 Dependency on the parent SPEC

`depends_on: [SPEC-GOAL-SURFACE-UNIFY-001]`. The basis is concrete, not organizational: `docs-site/content/en/cli-reference/handoff.md:34` reads

```
| `--goal <condition>` | Record the `/goal` condition (for restore guidance) |
```

which mirrors `internal/cli/handoff.go:104`'s help string `"record a /goal condition (restoration guidance only)"`. The parent's M7 rewrites that help string, and this SPEC's `handoff.md` work cannot be specified correctly until the parent's replacement wording is fixed. The `--goal` **flag name** itself is unchanged in both SPECs — it is a CLI contract.

### §A.3 Measured baseline

Commands and observed outputs are recorded in `research.md`. Backticked detector (`` `/goal ``) throughout, because an unbackticked search matches link paths such as `/en/cli-reference/goal`:

- **50 `docs-site` files** carry a native-`/goal` reference: **28** under `claude-code/`, **22** on the MoAI surface.
- **13 files** are in this SPEC's scope; of those, **12 are sweep targets** carrying **24 emission markers** (3 pages × 4 locales), and **1 is annotate-only** — the strategy record, swept of nothing per AC-GDR-007. The partition matches `plan.md` §F.1's ownership map (`autonomous-loops.md` ×4 + `self-evolving.md` ×4 + `handoff.md` ×4 + strategy ×1 = 13). The marker figure is the measured AC-GDR-012 aggregate (`total=24`), derived by running the command rather than carried from an earlier estimate. A superseded `18` appeared here at authoring: it used the disqualified `per-turn` detector's `2` instead of the corrected `4`, and omitted the L7 marker's `4` entirely — understating scope by exactly the two things this SPEC exists to fix (finding B-4).
- **Four retention surfaces** at this layer — see §B.2 and `plan.md` §A.2.

---

## §B Requirements (GEARS notation)

### §B.1 Emission retirement

- **REQ-GDR-001** (Event-driven) — **When** the sync phase runs, `manager-docs` shall replace every native-`/goal` **emission** reference in the 12 sweep-target locale files (3 pages × 4 locales) with `/moai goal`, or reword it so the MoAI pipeline is no longer described as emitting native `/goal`.

- **REQ-GDR-002** (State-driven) — **While** a documentation page exists in more than one locale, an emission reference removed in one locale shall be removed in every locale in which it appears.

- **REQ-GDR-003** (Ubiquitous) — Every emission detector used to judge this SPEC shall anchor on a token that survives translation — a backticked code literal, a structural marker, or a string already untranslated in all four locales — and shall not anchor on translatable prose.

- **REQ-GDR-004** (Event-driven) — **When** an emission detector is recorded, its baseline shall be stated **per locale**.

### §B.2 Retention

- **REQ-GDR-005** (Capability gate) — **Where** a documentation native-`/goal` reference is a factual statement about Claude Code, a comparison that distinguishes the primitives, or a record of a past proposal, that reference shall be retained. The authoritative membership list is the retention register at `plan.md` §A.2, which this requirement binds to by reference rather than restating.

- **REQ-GDR-006** (Unwanted) — The sync phase shall not modify the 28 `docs-site/content/*/claude-code/**` pages, the `/moai goal`-versus-native factual-contrast pages, or the `.moai/research/` archives.

- **REQ-GDR-007** (Capability gate) — **Where** a page is a **split surface** — carrying both emission and retained references — the criterion judging it shall pin the retained structure alongside the emission target, so the page cannot be swept to zero.

### §B.3 Locale parity

- **REQ-GDR-008** (Unwanted) — The sync phase shall not create new locale pages to force symmetry. The four-locale obligation binds pages that exist; a content gap predating this SPEC is not this SPEC's drift to close.

### §B.4 Detector discipline

- **REQ-GDR-009** (Capability gate) — **Where** the content a detector targets is locale-symmetric, the detector's per-locale baseline shall be symmetric; an asymmetric baseline is then evidence the detector is prose-anchored and shall be re-anchored before the criterion is accepted. **Where** the targeted content is genuinely locale-asymmetric, the asymmetry shall be recorded with its justification and the detector exempted by name, because forcing symmetry there would require creating content REQ-GDR-008 forbids.

- **REQ-GDR-010** (Event-driven) — **When** an emission detector's target state is zero in every locale, the criterion judging it shall additionally assert the detector matches non-zero content against an immutable recorded base, so a detector that matches nothing is distinguishable from a surface that was swept.

- **REQ-GDR-011** (Ubiquitous) — Every emission detector's match pattern shall contain a literal `/goal` token, so that a detector which is live and locale-symmetric but semantically aimed elsewhere is rejected before the criterion is accepted. Each detector's pattern shall be declared once and consumed by both the counting step and the aptness assertion, so the two cannot diverge.

---

## §C Out of Scope

### Out of Scope — the parent SPEC's layers

- Doctrine files under `.claude/`, their `internal/template/templates/` mirrors, the `/moai:goal` slash-command wrapper, and all Go emission paths. Those are `SPEC-GOAL-SURFACE-UNIFY-001` M1-M7.
- The `--goal` CLI flag **name**, which is a contract in both SPECs. Only its help string changes, and that change is the parent's M7.

### Out of Scope — retained documentation

- `docs-site/content/*/claude-code/**` (28 pages). These document Claude Code's own `/goal` feature; the statements stay factually true after MoAI stops emitting it.
- `docs-site/content/*/cli-reference/goal.md` and `docs-site/content/*/utility-commands/moai-goal.md`. These state that `/moai goal` is the programmatic counterpart of the HUMAN-ONLY native command — the justification for `/moai goal` existing.
- `docs-site/content/*/advanced/hooks-reference.md`. Its Stop-hook row's dual `` `/goal`/`/moai goal` `` mention is consistent with the retained native-`/goal` yield invariant in the parent's `internal/goal/evaluate.go`.
- `.moai/research/*.md` (3 files). Historical research archives; retroactively editing a past record would falsify it.
- `.moai/specs/**` historical SPEC artifacts, on the same rationale.

### Out of Scope — locale-symmetry manufacture

- Creating `ja` / `zh` content for `advanced/hooks-reference.md`. The page exists in all four locales; only `en` and `ko` carry a native-`/goal` reference. That is a pre-existing content gap, not drift introduced here (REQ-GDR-008).

### Out of Scope — URL-path false positives

- `docs-site/content/*/cli-reference/loop.md` (4 files). A bare `/goal` search matches the link path `/en/cli-reference/goal`; these carry no command reference and appear in no criterion.

---

## §D Acceptance Criteria

Enumerated in `acceptance.md`: **12 criteria**, AC-GDR-001 through AC-GDR-012 (contiguous), each with its judgment command and the verbatim baseline observed in the worktree. Every emission criterion carries a **per-locale** baseline demonstrating symmetry (REQ-GDR-004).

Provenance of the carried criteria — the parent's identifiers are recorded so the two audits' evidence stays connected:

| This SPEC | Came from `SPEC-GOAL-SURFACE-UNIFY-001` | Change |
|---|---|---|
| AC-GDR-001, 002, 003 | AC-GSU-028 (emission half) | Split per marker; per-locale baselines added |
| AC-GDR-004 | **new** | Closes the parent's finding S-new-3 — the L7 mis-attribution, previously neither swept nor pinned |
| AC-GDR-005 | AC-GSU-028 (emission half) | **The N2 fix** — the a3 detector re-anchored from `per-turn` prose onto the untranslated `auto mode` |
| AC-GDR-006 | AC-GSU-028 (retention half) | Split-surface pins, converted to per-locale |
| AC-GDR-007 | AC-GSU-032 | Baseline unchanged (`0` / `25`) |
| AC-GDR-008 | AC-GSU-029 | Baseline unchanged (`cc=80 goal.md=12 moai-goal=4 hooks=2 research=4`) |
| AC-GDR-009, 010, 011 | **new** | Locale inventory · the symmetry+liveness meta-guard · the docs build |
| AC-GDR-012 | **new** | Aggregate emission integration criterion (`total=24`), which the four compounds reference |

Corrected at this SPEC's audit iteration 1 (finding B-3): four rows mapped to the wrong criterion. Each row was re-derived from the criterion's actual subject rather than carried — AC-GDR-005 is emission (not retention), AC-GDR-006 is the retention half (not AC-GSU-029), AC-GDR-008 carries AC-GSU-029 verbatim (not "new"), and AC-GDR-012 is the aggregate integration criterion (the docs build is AC-GDR-011). The parent's §B.7 register carries the mirror-image table; the two are cross-checked against each other and against the identifiers actually present in each SPEC.

A **REQ ↔ AC traceability matrix** is at `acceptance.md` §E.

**Artifact set — deliberately beyond the Tier M minimum (finding B-6).** Tier M requires three artifacts (`spec` + `plan` + `acceptance`); six are delivered. `design.md` and `research.md` are retained deliberately, not by tier misclassification: the two MUST-FIX findings of this SPEC's first audit (B-1 liveness, B-2 asymmetry carve-out) are both *design* questions about detector semantics, and `research.md` §C carries the per-detector locale-symmetry proof plus the disqualified-detector record that keeps the N2 anchor from being reintroduced. Trimming them would discard the evidence that makes the criteria auditable. The tier remains **M** — confirmed correct against the complexity criteria (13 files, one verification regime, non-constitutional, fully reversible); only the artifact count exceeds the minimum.

---

## §E Cross-References

- `SPEC-GOAL-SURFACE-UNIFY-001` — the parent; `spec.md` §B.7 there carries the moved-identifier register.
- `SPEC-GOAL-SURFACE-UNIFY-001/plan-audit-2.md` — the iteration-2 audit that identified N2 and proposed this split.
- `CLAUDE.local.md` §17 — the four-locale docs-site synchronization obligation.
- `.moai/docs/docs-site-i18n-rules.md` — the docs-site i18n doctrine.
