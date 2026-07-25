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

## §A Baseline Provenance

Every judgment command below was **executed** in the worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/goal-retire` (branch `feat/SPEC-GOAL-SURFACE-UNIFY-001`, base `origin/main` = `e306e21a9`). Baselines are verbatim observed output. Every criterion **FAILS** at baseline.

### §A.1 Detector conventions

- **Backticked detector.** `` `/goal `` (backtick + `/goal`). An unbackticked search matches link paths such as `/en/cli-reference/goal`; using it inflated a first pass to 58 files / 32 `claude-code` pages against the correct **50 / 28**.
- **Occurrence counts.** `grep -o … | wc -l`, never bare `grep -c`, except where a per-line presence check is the intent (marked).
- **Per-locale baselines are mandatory** for every emission criterion (REQ-GDR-004). A single aggregate number cannot distinguish "removed everywhere" from "removed in English only".

### §A.2 The locale-invariance rule and why every detector here obeys it

The parent SPEC's plan-audit iteration 2 found an emission detector anchored on the English string `per-turn`, reading `en:2 ja:0 ko:0 zh:0` while the same content sat in all four locales at line 94 — only the trigger word was translated (`ターンごとの` / `턴별` / `每回合`). Sweeping English alone would have driven it to `0` with three locales' emission content intact.

Every detector below is therefore anchored on a token that **survives translation**, and each is proven symmetric by measurement rather than inspection:

| Anchor | Translated? | Used by |
|---|---|---|
| `` `/goal` `` (backticked code span) | No | all |
| `` `/moai loop` `` | No | AC-GDR-001, 002 |
| `auto mode` | No — untranslated in all four locales | AC-GDR-005 |
| `ac_converge` | No — code identifier | AC-GDR-005 (alternate) |
| `--goal` | No — CLI flag | AC-GDR-003 |
| structural markers (`^## `, `^\| `, `^### `) | No | AC-GDR-006 |
| `per-turn`, "continuation", "per-tool" | **Yes — PROHIBITED as anchors** | none |

The meta-guard is **AC-GDR-010**: it asserts every emission detector yields exactly ONE distinct value across the four locales, making the N2 defect class mechanically impossible to ship.

---

## §B AC Matrix — 12 criteria

Grouped by milestone rather than by identifier, so each milestone's sweep and retention halves sit together. Identifiers are contiguous `001`..`012`.

Shell prelude:

```bash
AL=docs-site/content/*/advanced/autonomous-loops.md
SE=docs-site/content/*/advanced/self-evolving.md
HO=docs-site/content/*/cli-reference/handoff.md
```

### N1 — `autonomous-loops.md` split surface

**AC-GDR-001** — The MoAI-primitive paired listing is removed in **every** locale.

```bash
for l in en ja ko zh; do printf "%s:%s " $l "$(grep -ohE '`/moai loop`[ ·/]+`/goal`' docs-site/content/$l/advanced/autonomous-loops.md | wc -l | tr -d ' ')"; done
```

- Recorded baseline: `en:1 ja:1 ko:1 zh:1` (symmetric)
- Target: `en:0 ja:0 ko:0 zh:0`

**AC-GDR-004** — The L7 mis-attribution is removed in **every** locale. L7 currently frames native `/goal` as one of *MoAI-ADK's* primitives — the exact mis-attribution the retirement thesis rejects. Closes the parent's finding S-new-3, where this line sat in a coverage hole: neither swept nor pinned.

```bash
for l in en ja ko zh; do printf "%s:%s " $l "$(sed -n '7p' docs-site/content/$l/advanced/autonomous-loops.md | grep -c '`/goal`')"; done
```

- Recorded baseline: `en:1 ja:1 ko:1 zh:1` (symmetric)
- Target: `en:0 ja:0 ko:0 zh:0`
- Line-number anchored by design: the mis-attribution is the page's opening framing sentence. Should the line move, the criterion is re-anchored on `provides` + `` `/goal` `` co-occurrence rather than silently passing.

**AC-GDR-005** — **The N2 fix.** The auto-mode pairing is removed in **every** locale, detected via the untranslated string `auto mode` co-occurring with the code span `` `/goal` `` on one line — replacing the parent's `per-turn` prose anchor.

```bash
for l in en ja ko zh; do printf "%s:%s " $l "$(grep -c 'auto mode.*`/goal`' docs-site/content/$l/advanced/autonomous-loops.md)"; done
```

- Recorded baseline: `en:1 ja:1 ko:1 zh:1` (symmetric)
- Target: `en:0 ja:0 ko:0 zh:0`
- **Contrast with the superseded detector**, which is why this AC exists: `grep -ohE '\`/goal\`[^.]*per-turn|per-turn[^.]*\`/goal\`'` reads `en:2 ja:0 ko:0 zh:0` on the same content. Both were run; the asymmetry is the disqualifying signal.

**AC-GDR-006** — Retention pins: the split surface's native structure survives in **every** locale (register row 2). Per-locale, so an edit that flattened one locale's page cannot hide behind three intact ones.

```bash
for l in en ja ko zh; do printf "%s:h3=%s,h2=%s,row=%s " $l \
  "$(grep -cF '### `/goal` — native Claude Code (HUMAN-ONLY)' docs-site/content/$l/advanced/autonomous-loops.md)" \
  "$(grep -cE '^## .*/goal' docs-site/content/$l/advanced/autonomous-loops.md)" \
  "$(grep -cE '^\| `/goal` \|' docs-site/content/$l/advanced/autonomous-loops.md)"; done
```

- Recorded baseline: `en:h3=1,h2=1,row=1 ja:… ko:… zh:…` — all four identical
- Target: **unchanged** — all four remain `h3=1,h2=1,row=1`
- The pins hold at baseline, so this criterion is compounded with AC-GDR-012 (aggregate emission at `0`) for falsifiability. Alone it would be a no-op; together they express "swept the emission, kept the structure".

### N2 — `self-evolving.md`

**AC-GDR-002** — The paired listing is removed in **every** locale.

```bash
for l in en ja ko zh; do printf "%s:%s " $l "$(grep -ohE '`/moai loop`[ ·/]+`/goal`' docs-site/content/$l/advanced/self-evolving.md | wc -l | tr -d ' ')"; done
```

- Recorded baseline: `en:2 ja:2 ko:2 zh:2` (symmetric — L40 and L97 in each)
- Target: `en:0 ja:0 ko:0 zh:0`

### N3 — `handoff.md`

**AC-GDR-003** — The `--goal` row's native reference is removed in **every** locale. Blocked on the parent's M7, which decides the replacement wording (`plan.md` B4).

```bash
for l in en ja ko zh; do printf "%s:%s " $l "$(grep -ohF '`/goal' docs-site/content/$l/cli-reference/handoff.md | wc -l | tr -d ' ')"; done
```

- Recorded baseline: `en:1 ja:1 ko:1 zh:1` (symmetric)
- Target: `en:0 ja:0 ko:0 zh:0`
- The `--goal` **flag name** is unchanged — only the description. A criterion asserting the flag's disappearance would break a CLI contract.

### N4 — Strategy record

**AC-GDR-007** — Compound: the superseded strategy record gains a superseding note **and** keeps all 25 historical occurrences. The content pin is what distinguishes annotating a record from rewriting one.

```bash
grep -ciF 'native `/goal` emission is retired' .moai/docs/autonomous-workflow-strategy.md
grep -ohF '`/goal' .moai/docs/autonomous-workflow-strategy.md | wc -l | tr -d ' '
```

- Recorded baseline: `0` / `25`
- Target: `>= 1` / `25`
- A generic `grep -ciE 'superseded|retired'` is disqualified as a detector: it already reads `4` at baseline against unrelated prose. Both were run; hence the specific sentinel.

### N5 — Retention and locale-parity verification

**AC-GDR-008** — Compound retention guard across all five retained groups (register rows 1, 3, 4 plus `hooks-reference.md`), each pinned at its exact count, **and** the aggregate emission at `0`. The pins hold at baseline, so the compound is what makes this falsifiable — and a sweep over-reaching into `claude-code/**` drives a pin off its value and fails.

```bash
printf "cc=%s goal.md=%s moai-goal=%s hooks=%s research=%s\n" \
  "$(grep -rohF '`/goal' docs-site/content/*/claude-code/ | wc -l | tr -d ' ')" \
  "$(grep -rohF '`/goal' docs-site/content/*/cli-reference/goal.md | wc -l | tr -d ' ')" \
  "$(grep -rohF '`/goal' docs-site/content/*/utility-commands/moai-goal.md | wc -l | tr -d ' ')" \
  "$(grep -rohF '`/goal' docs-site/content/*/advanced/hooks-reference.md | wc -l | tr -d ' ')" \
  "$(grep -rohF '`/goal' .moai/research/ | wc -l | tr -d ' ')"
```

- Recorded baseline: `cc=80 goal.md=12 moai-goal=4 hooks=2 research=4`
- Target: **identical**, and AC-GDR-012 at `0`
- The `hooks=2` pin doubles as the REQ-GDR-008 guard: adding `ja`/`zh` lines to force locale symmetry would drive it to `4` and fail.

**AC-GDR-009** — Compound: the four-locale file inventory is unchanged (no new locale pages, REQ-GDR-008) **and** the aggregate emission is `0`.

```bash
for p in advanced/autonomous-loops.md cli-reference/handoff.md advanced/self-evolving.md; do \
  printf "%s=%s " "$(basename $p)" "$(ls docs-site/content/{en,ja,ko,zh}/$p 2>/dev/null | wc -l | tr -d ' ')"; done
```

- Recorded baseline: `autonomous-loops.md=4 handoff.md=4 self-evolving.md=4`
- Target: **identical** (`4` each), and AC-GDR-012 at `0`

**AC-GDR-010** — **The locale-invariance meta-guard.** Every emission detector yields exactly ONE distinct value across the four locales. This is what makes the N2 defect class mechanically impossible: a prose-anchored detector produces ≥ 2 distinct values (e.g. `{2, 0}`) and fails here even while its own aggregate reads `0`.

```bash
for name in paired_al auto_mode l7 paired_se handoff; do
  case $name in
    paired_al) v=$(for l in en ja ko zh; do grep -ohE '`/moai loop`[ ·/]+`/goal`' docs-site/content/$l/advanced/autonomous-loops.md | wc -l | tr -d ' '; done | sort -u | wc -l | tr -d ' ');;
    auto_mode) v=$(for l in en ja ko zh; do grep -c 'auto mode.*`/goal`' docs-site/content/$l/advanced/autonomous-loops.md; done | sort -u | wc -l | tr -d ' ');;
    l7)        v=$(for l in en ja ko zh; do sed -n '7p' docs-site/content/$l/advanced/autonomous-loops.md | grep -c '`/goal`'; done | sort -u | wc -l | tr -d ' ');;
    paired_se) v=$(for l in en ja ko zh; do grep -ohE '`/moai loop`[ ·/]+`/goal`' docs-site/content/$l/advanced/self-evolving.md | wc -l | tr -d ' '; done | sort -u | wc -l | tr -d ' ');;
    handoff)   v=$(for l in en ja ko zh; do grep -ohF '`/goal' docs-site/content/$l/cli-reference/handoff.md | wc -l | tr -d ' '; done | sort -u | wc -l | tr -d ' ');;
  esac
  printf "%s:distinct=%s " "$name" "$v"
done
```

- Recorded baseline: `paired_al:distinct=1 auto_mode:distinct=1 l7:distinct=1 paired_se:distinct=1 handoff:distinct=1`
- Target: **all five remain `distinct=1`**, and AC-GDR-012 at `0`
- Symmetry holds at baseline (all detectors were re-anchored at authoring time precisely so it would), so this is compounded with AC-GDR-012. Its value is that it fails the moment a single-locale edit lands.

**AC-GDR-011** — Compound held-out: the docs-site builds **and** the aggregate emission is `0`. A sweep that broke shortcode or front-matter syntax in one locale would fail the build half.

```bash
cd docs-site && hugo --quiet --destination /tmp/gdr-build >/dev/null 2>&1; echo "exit=$?"
```

- Recorded baseline: `exit=0`
- Target: `exit=0`, and AC-GDR-012 at `0`

**AC-GDR-012** — Aggregate emission across all 8 sweep-target files reaches `0`. This is the integration criterion the four compounds above reference.

```bash
t=0
t=$((t + $(grep -rhoE '`/moai loop`[ ·/]+`/goal`' docs-site/content/*/advanced/autonomous-loops.md | wc -l | tr -d ' ')))
t=$((t + $(grep -rhoE '`/moai loop`[ ·/]+`/goal`' docs-site/content/*/advanced/self-evolving.md | wc -l | tr -d ' ')))
t=$((t + $(grep -rhc 'auto mode.*`/goal`' docs-site/content/*/advanced/autonomous-loops.md | paste -sd+ - | bc)))
t=$((t + $(for l in en ja ko zh; do sed -n '7p' docs-site/content/$l/advanced/autonomous-loops.md | grep -c '`/goal`'; done | paste -sd+ - | bc)))
t=$((t + $(grep -rhoF '`/goal' docs-site/content/*/cli-reference/handoff.md | wc -l | tr -d ' ')))
echo "total=$t"
```

- Recorded baseline: `total=24` (paired_al 4 + paired_se 8 + auto_mode 4 + l7 4 + handoff 4)
- Target: `total=0`

---

## §C Held-Out Gates (NOT acceptance criteria)

| Gate | Command | Baseline observed |
|---|---|---|
| Four-locale inventory beyond the scoped set | `ls docs-site/content/*/advanced/hooks-reference.md \| wc -l` | `4` (all four locales exist; only en/ko carry the reference) |
| Parent dependency landed | `grep -c 'a /moai goal condition' internal/cli/handoff.go` | `0` at baseline — becomes `1` when the parent's M7 lands. **N3 is blocked until then** |

The hugo build appears as AC-GDR-011's first half rather than a pure gate, because compounding it with the emission target makes it falsifiable; as a standalone gate it passes at baseline.

---

## §D Definition of Done

- All 12 criteria transitioned from their recorded baseline to their target.
- All emission detectors still yield `distinct=1` across locales (AC-GDR-010) — no single-locale edits.
- The four retention surfaces in `plan.md` §A.2 are untouched.
- No new locale pages created.
- Working tree contains no changes outside the 13 paths in `plan.md` §F.1.

---

## §E REQ ↔ AC Traceability Matrix

| REQ | Covered by | Note |
|---|---|---|
| REQ-GDR-001 (sweep emission) | AC-GDR-001..005, AC-GDR-012 | One criterion per marker plus the aggregate |
| REQ-GDR-002 (remove in every locale) | AC-GDR-001..005 (per-locale), AC-GDR-010 | Per-locale targets; the meta-guard fails on asymmetry |
| REQ-GDR-003 (detectors survive translation) | AC-GDR-005, AC-GDR-010 | AC-005 is the re-anchored detector; AC-010 enforces the property generally |
| REQ-GDR-004 (per-locale baselines, symmetric) | AC-GDR-010 | The only criterion whose subject is the detectors themselves |
| REQ-GDR-005 (retention register) | AC-GDR-006, AC-GDR-007, AC-GDR-008 | One guard per register row |
| REQ-GDR-006 (do not modify CC / contrast / archives) | AC-GDR-008 | Five pinned counts |
| REQ-GDR-007 (split surface pinned, not zeroed) | AC-GDR-006 | Per-locale structural pins |
| REQ-GDR-008 (no locale-symmetry manufacture) | AC-GDR-008 (`hooks=2`), AC-GDR-009 (inventory) | Two independent guards: adding a line fails the first, adding a page fails the second |

Coverage: **8 / 8 REQs** cited by ≥ 1 AC. Reverse: every AC-GDR-001..012 appears in ≥ 1 row.

---

## §F Edge Cases

- **Sweeping `autonomous-loops.md` to zero.** Blocked by AC-GDR-006's per-locale `h3=1,h2=1,row=1` pins.
- **Editing `en` only.** Blocked twice: the per-locale targets in AC-GDR-001..005, and AC-GDR-010's `distinct=1` requirement.
- **Re-introducing a prose-anchored detector.** Blocked by AC-GDR-010, which fails at `distinct=2` even when the detector's own aggregate reads `0`.
- **Rewriting the strategy record instead of annotating it.** Blocked by AC-GDR-007's `25` content pin.
- **Adding `ja`/`zh` `hooks-reference.md` lines** to force symmetry. Blocked by AC-GDR-008's `hooks=2` pin.
- **Creating a new locale page.** Blocked by AC-GDR-009's inventory pin.
- **Touching `claude-code/**`.** Blocked by AC-GDR-008's `cc=80` pin.
- **Specifying `handoff.md` before the parent's M7 lands.** Blocked by the §C dependency gate at `0`.
- **Breaking a shortcode while sweeping.** Blocked by AC-GDR-011's build half.
