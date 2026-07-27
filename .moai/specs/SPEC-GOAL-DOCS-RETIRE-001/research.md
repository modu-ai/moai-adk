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

> Thin by design: the documentation measurement, the split-surface classification, and the per-detector locale-symmetry proof. Every figure cites the command that produced it.

## §A Corpus Measurement

Executed in the worktree at base `origin/main` = `e306e21a9`.

```bash
grep -rlF '`/goal' docs-site/content/ | sort
```

Observed: **50 files** — **28** under `claude-code/`, **22** on the MoAI surface.

The MoAI-surface 22 decompose as: `advanced/autonomous-loops.md` ×4, `cli-reference/handoff.md` ×4, `advanced/self-evolving.md` ×4, `cli-reference/goal.md` ×4, `utility-commands/moai-goal.md` ×4, `advanced/hooks-reference.md` ×2 (en, ko only).

### §A.1 URL-path false positives — why the detector is backticked

An unbackticked `/goal` search returns **58** files and **32** `claude-code` pages. The delta is link paths such as `/en/cli-reference/goal`:

```bash
grep -n '/goal' docs-site/content/en/cli-reference/loop.md
# 41:- [moai goal](/en/cli-reference/goal)
```

`cli-reference/loop.md` (×4) carries **no** command reference and is excluded from every criterion. All figures in this SPEC use the backticked form.

## §B Split-Surface Classification — `autonomous-loops.md`

```bash
grep -nE '/goal' docs-site/content/en/advanced/autonomous-loops.md
grep -nE '^###* .*/goal' docs-site/content/{en,ja,ko,zh}/advanced/autonomous-loops.md
```

Observed structure, present in all four locales:

| Line | Content | Class |
|---|---|---|
| 7 | "MoAI-ADK provides three continuation-loop primitives … `/goal`" | **sweep** — mis-attributes a Claude Code command to MoAI-ADK |
| 21 | `\| \`/goal\` \| User TUI (HUMAN-ONLY) \| …` comparison row | retain |
| 35 | `### \`/goal\` — native Claude Code (HUMAN-ONLY)` | retain |
| 37 | "a native Claude Code TUI command … the model cannot invoke" | retain |
| 45 | "Bare `/goal` checks status; `/goal clear` terminates early" | retain |
| 49 | "Since native `/goal` is HUMAN-ONLY, this is the only path for the orchestrator to …" | retain — Axis-B justification |
| 67 | `## Native /goal Details` (localized: `ネイティブ` / `네이티브` / `原生`) | retain |
| 83 | "`/goal` (native) — implemented in Claude Code runtime (requires v2.1.139+)" | retain |
| 92 | "Even with `/goal` active, user approval … is mandatory" | retain |
| 94 | "auto mode … with `/goal` (per-turn continuation) … `ac_converge`" | **sweep** |
| 99 | "`/moai loop` / `/goal` convergence trajectories" | **sweep** |

Applying the membership test — *does the sentence become false when MoAI stops emitting native `/goal`?* — every "retain" row stays true. Line 49 is verbatim the same justification for which `cli-reference/goal.md` is retained. **This is why a blanket `0` target on this file is destructive**: it would delete two sections, a comparison row, and the Axis-B justification, four times over.

## §C Locale-Symmetry Proof, Per Detector

This is the section that closes the parent SPEC's audit finding N2. Each detector was run against all four locale files; a symmetric result is the acceptance condition for using it.

### §C.1 The disqualified detector (recorded so it is not re-introduced)

```bash
for l in en ja ko zh; do grep -ohE '`/goal`[^.]*per-turn|per-turn[^.]*`/goal`' \
  docs-site/content/$l/advanced/autonomous-loops.md | wc -l; done
```

Observed: **`en:2 ja:0 ko:0 zh:0`** — asymmetric.

The content is present in all four locales; only the trigger word is translated:

```
en:94  … auto mode (per-tool auto-approval) with `/goal` (per-turn continuation) …
ja:94  … auto mode（ツールごとの自動承認）と`/goal`（ターンごとの連続）を組合せると…
ko:94  … auto mode(도구별 자동 승인)와 `/goal`(턴별 연속)을 조합하면…
zh:94  … auto mode（每工具自动批准）与 `/goal`（每回合连续）组合可实现…
```

`per-turn` → `ターンごとの` / `턴별` / `每回合`. An AC built on this detector would go green with three locales' emission content intact.

### §C.2 The accepted detectors — all symmetric

```bash
for l in en ja ko zh; do printf "%s " "$(grep -c 'auto mode' docs-site/content/$l/advanced/autonomous-loops.md)"; done
for l in en ja ko zh; do printf "%s " "$(grep -c 'ac_converge' docs-site/content/$l/advanced/autonomous-loops.md)"; done
```

Observed: `auto mode` → **1 1 1 1**; `ac_converge` → **1 1 1 1**. Both untranslated in all four locales, so either is a valid anchor. `auto mode` is used (AC-GDR-005) with `ac_converge` recorded as the alternate.

Full per-detector symmetry table, each row an executed command:

| Detector | en | ja | ko | zh | Symmetric | Used by |
|---|---:|---:|---:|---:|---|---|
| `` `/moai loop` `` + `` `/goal` `` — autonomous-loops | 1 | 1 | 1 | 1 | yes | AC-GDR-001 |
| `` `/moai loop` `` + `` `/goal` `` — self-evolving | 2 | 2 | 2 | 2 | yes | AC-GDR-002 |
| `` `/goal `` — handoff.md | 1 | 1 | 1 | 1 | yes | AC-GDR-003 |
| L7 `` `/goal` `` presence | 1 | 1 | 1 | 1 | yes | AC-GDR-004 |
| `auto mode` + `` `/goal` `` on one line | 1 | 1 | 1 | 1 | yes | AC-GDR-005 |
| native H3 / `## …/goal` H2 / comparison row | 1/1/1 | 1/1/1 | 1/1/1 | 1/1/1 | yes | AC-GDR-006 |
| `` `/goal `` — claude-code/** per locale | 20 | 20 | 20 | 20 | yes | AC-GDR-008 |
| **superseded** `per-turn` anchor | **2** | **0** | **0** | **0** | **NO** | none — disqualified |

Aggregate emission across the 8 sweep files: **24** (paired_al 4 + paired_se 8 + auto_mode 4 + l7 4 + handoff 4).

## §D Retention Rationale, Per Surface

| Surface | Command | Observed | Rationale |
|---|---|---:|---|
| `claude-code/**` | `grep -rohF '\`/goal' docs-site/content/*/claude-code/ \| wc -l` | 80 | Documents Claude Code's own feature |
| `cli-reference/goal.md` | same, scoped | 12 | `ko:9` states `/moai goal` is the programmatic counterpart of the HUMAN-ONLY native command |
| `utility-commands/moai-goal.md` | same, scoped | 4 | `ko:14` same factual contrast |
| `advanced/hooks-reference.md` | same, scoped | 2 | Stop-hook dual mention, consistent with the parent's retained yield invariant |
| `.moai/research/*.md` | same, scoped | 4 | Historical archives |
| `.moai/docs/autonomous-workflow-strategy.md` | `grep -ohF '\`/goal' … \| wc -l` | 25 | Superseded strategy record — see §E |

## §E The Strategy Record — why retain-plus-note, not sweep

```bash
grep -nF '`/goal' .moai/docs/autonomous-workflow-strategy.md
grep -rln 'autonomous-workflow-strategy' .claude/ internal/template/templates/ docs-site/content/
ls internal/template/templates/.moai/docs/
```

Observed: 25 occurrences comprising a 3-engine comparison (L6, 12, 21, 26, 28, 60), Kickoff-invariant statements (L86, 435), an anti-pattern catalogue (L461-467), and a roadmap (L549). **Zero referrers** anywhere in the harness or docs. **Not** present in the template tree, so it is not user-distributed.

The document self-declares its status as a proposal record. Retirement makes it historically superseded, not factually false — the same basis on which `.moai/research/` archives are excluded. Sweeping 25 occurrences would rewrite a past record; a one-line superseding note preserves it while flagging that the proposal's `/goal` engine was retired.

Detector note: a generic `grep -ciE 'superseded|retired'` reads **4** at baseline against unrelated prose, so it cannot serve as the note detector. AC-GDR-007 uses a specific sentinel phrase and pairs it with a `25` content pin.

## §F Locale-Asymmetry Finding — flagged, not fixed

```bash
ls docs-site/content/*/advanced/hooks-reference.md
for l in en ko ja zh; do grep -n '/goal' docs-site/content/$l/advanced/hooks-reference.md; done
```

Observed: the page exists in **all four** locales, but only `en:170` and `ko:170` carry the Stop-hook row mentioning `/goal`; `ja` and `zh` have no such line.

So there is no page to create — this is a pre-existing four-locale **content** gap inside pages that already exist. REQ-GDR-008 forbids closing it here, and AC-GDR-008's `hooks=2` pin fails if `ja`/`zh` lines are added.
