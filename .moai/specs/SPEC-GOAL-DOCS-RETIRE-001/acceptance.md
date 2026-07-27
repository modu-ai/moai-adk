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
- **Occurrence counts vs presence checks.** `grep -o … | wc -l` (occurrences) is the default. `grep -c` (matching-line count) is used **deliberately** by AC-GDR-004, AC-GDR-005, and AC-GDR-006, where a per-line **presence** check is the intent: each targets a single marker or structure on one line, and presence is the property being asserted. Marked here rather than per-criterion (finding B-5). Consequence accepted: two markers on one line would count as one — verified not to occur for these three, since each anchor appears at most once per line in the measured corpus.
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

The meta-guard is **AC-GDR-010**: it asserts every emission detector yields exactly ONE distinct value across the four locales, **and** that each detector still matches non-zero content against the immutable recorded base, **and** that each detector's pattern actually carries a literal `/goal` token. The second component distinguishes a swept surface from a broken regex (finding B-1); the third distinguishes a detector aimed at the emission reference from one aimed at a token that merely sits beside it (finding B2-1).

### §A.3 Asymmetry carve-out — why symmetry is not an unconditional rule

Locale asymmetry has **two** possible causes, and only one is a defect:

| Cause | Example | Disposition |
|---|---|---|
| **Prose-anchored detector** (the N2 defect class) | `per-turn` reads `en:2 ja:0 ko:0 zh:0` against content present in all four locales | Disqualifying — re-anchor the detector |
| **Genuinely locale-asymmetric content** | `advanced/hooks-reference.md` carries the reference at `en:170` and `ko:170` only; `ja`/`zh` have no such line, measured `en:1 ja:0 ko:1 zh:0` → `distinct=2` | **Not** disqualifying — REQ-GDR-008 forbids closing this pre-existing gap |

A blanket symmetry rule conflates them, and this SPEC contains a live example of the second: a *correct* detector aimed at `hooks-reference.md` would read `distinct=2` and be rejected, creating pressure to manufacture symmetry — exactly what REQ-GDR-008 prohibits (finding B-2).

Resolution: AC-GDR-010 component (a) applies only where the underlying content is symmetric. A detector aimed at genuinely asymmetric content is **exempted by name**, with the measured content asymmetry cited as justification.

**Exemption list: empty.** No current detector needs it — `hooks-reference.md` is a retention surface (guarded by AC-GDR-008's `hooks=2` pin) and appears in no emission detector, so the conflict is latent rather than live. The carve-out is stated so the rule is correct as written, not because a detector currently relies on it.

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
- **Line-number anchored, with a locale-invariant fallback (finding B-7).** `sed -n '7p'` is brittle if the line moves. The stated fallback must NOT be `provides` + `` `/goal` `` — `provides` is English prose and would reintroduce the N2 class. The durable fallback is the **triple co-occurrence** `` `/goal` `` + `` `/moai goal` `` + `` `/moai loop` `` on one line, all three code literals and therefore locale-invariant: `grep -c '`/goal`.*`/moai goal`.*`/moai loop`\|`/goal`.*`/moai loop`.*`/moai goal`' <file>`. Measured at the base: `en:1 ja:1 ko:1 zh:1` — symmetric, so the fallback is admissible under §A.3.

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
grep -vF 'native `/goal` emission is retired' .moai/docs/autonomous-workflow-strategy.md | grep -ohF '`/goal' | wc -l | tr -d ' '
```

- Recorded baseline: `0` / `25`
- Target: `>= 1` / `25`
- The second detector excludes the sentinel line before counting because the phrase component 1 mandates — ``native `/goal` emission is retired`` — itself carries a backticked `` `/goal ``; counting it would drive the pin to `26` the moment component 1 is satisfied, making the compound self-contradictory. Excluding it keeps the pin asserting exactly what it means: all 25 *historical* occurrences survive.
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

**AC-GDR-010** — **The locale-invariance meta-guard, with a liveness floor and an aptness assertion.** Compound, four components:

**(a) Symmetry** — every emission detector yields exactly ONE distinct value across the four locales. A prose-anchored detector produces ≥ 2 distinct values (e.g. `{2, 0}`) and fails here even while its own aggregate reads `0`.

**(b) Liveness (added at audit iteration 1, finding B-1)** — every emission detector, run against the **immutable recorded base** `e306e21a9`, matches non-zero content in **all four** locales. This is what separates "content swept" from "regex broken": at judgment time both produce `0/0/0/0` against the working tree, and component (a) alone is satisfied by a dead detector (`distinct=1` over four zeros). Anchoring liveness to a fixed commit rather than the working tree means the check keeps its discriminating power *after* the sweep lands.

**(c) Asymmetry carve-out (added at audit iteration 1, finding B-2)** — component (a) is disqualifying only where the underlying **content** is locale-symmetric. Where content is genuinely locale-asymmetric, the asymmetry is recorded in the baseline with a justification and the detector is exempted from (a) **by name**. No detector is currently exempt; the exemption list is empty and any addition must cite the measured content asymmetry. See §A.3.

**(d) Aptness (added at audit iteration 2, finding B2-1)** — every emission detector's match pattern contains a literal `/goal` token. Components (a) and (b) constrain a detector's *behaviour* (symmetric, non-empty against the base) but never tie it to the token the sweep must remove, so a detector that is live and symmetric yet semantically aimed elsewhere passes both. Each detector therefore declares its pattern **once** in a `p=` variable that the counting function `w()` and the aptness test both read, so a weakening of the pattern cannot leave a stale assertion behind. The mechanism is one `case` expansion, not a second command.

> **Rejected alternative — base-match-set subset.** Audit iteration 2 proposed a stronger form: assert each detector's base-match line set is a subset of the base lines carrying `` `/goal` ``. It was **executed and refuted**. In the base `e306e21a9`, `ac_converge` and `auto mode` both occur on the same line as `` `/goal` `` (`advanced/autonomous-loops.md` line 7 in `en`), so both attack detectors yield `leak=0` and pass the subset test. Line granularity is exactly the granularity the "adjacent token" attack exploits, so the subset form cannot discriminate here; the pattern-literal form rejects all three controls.

```bash
# (a) symmetry — distinct values per detector, working tree
# (b) liveness — same detector against the immutable base, minimum across locales
# (d) aptness  — the single pattern source $p carries a literal /goal token
for name in paired_al auto_mode l7 paired_se handoff; do
  case $name in
    paired_al) p='`/moai loop`[ ·/]+`/goal`'; w() { grep -ohE "$p" "$1" | wc -l | tr -d ' '; }; f=advanced/autonomous-loops.md;;
    auto_mode) p='auto mode.*`/goal`';        w() { grep -c "$p" "$1"; };                      f=advanced/autonomous-loops.md;;
    l7)        p='`/goal`';                   w() { sed -n '7p' "$1" | grep -c "$p"; };        f=advanced/autonomous-loops.md;;
    paired_se) p='`/moai loop`[ ·/]+`/goal`'; w() { grep -ohE "$p" "$1" | wc -l | tr -d ' '; }; f=advanced/self-evolving.md;;
    handoff)   p='`/goal';                    w() { grep -ohF "$p" "$1" | wc -l | tr -d ' '; }; f=cli-reference/handoff.md;;
  esac
  d=$(for l in en ja ko zh; do w docs-site/content/$l/$f; done | sort -u | wc -l | tr -d ' ')
  m=$(for l in en ja ko zh; do git show e306e21a9:docs-site/content/$l/$f > /tmp/gdr-live.md; w /tmp/gdr-live.md; done | sort -n | head -1)
  case "$p" in *'/goal'*) a=1;; *) a=0;; esac
  printf "%s:distinct=%s,live_min=%s,apt=%s " "$name" "$d" "$m" "$a"
done
```

- Recorded baseline: `paired_al:distinct=1,live_min=1,apt=1 auto_mode:distinct=1,live_min=1,apt=1 l7:distinct=1,live_min=1,apt=1 paired_se:distinct=1,live_min=2,apt=1 handoff:distinct=1,live_min=1,apt=1`
- Target: **all five keep `distinct=1` AND `live_min >= 1` AND `apt=1`**, and AC-GDR-012 at `0`
- **Three controls, all executed against this block.** Each defeats a different component, and none is defeated by the others:

  | Control detector | `distinct` | `live_min` | `apt` | Rejected by |
  |---|---|---|---|---|
  | `auto-mode` (hyphenated typo — the iteration-1 dead-detector control) | `1` | **`0`** | — | (b) liveness |
  | `ac_converge` (adjacent token, co-occurs with `` `/goal` `` on the same line) | `1` | `1` | **`0`** | (d) aptness |
  | `auto mode` (AC-GDR-005's compound regex with its `` `/goal` `` half dropped) | `1` | `1` | **`0`** | (d) aptness |

- The second and third controls pass (a) **and** (b) — that is the hole component (d) closes. They are the plausible-mutation path, not a contrived one: simplifying a compound regex to its untranslated-anchor half is exactly the edit a sweep makes while driving a count toward zero, and it leaves a criterion satisfiable by deleting the words "auto mode" while the native-`/goal` reference survives.
- All four components hold at baseline, so this criterion is compounded with AC-GDR-012 for falsifiability.

**AC-GDR-011** — Compound held-out: the docs-site builds **and** the aggregate emission is `0`. A sweep that broke shortcode or front-matter syntax in one locale would fail the build half.

```bash
cd docs-site && hugo --quiet --destination /tmp/gdr-build >/dev/null 2>&1; echo "exit=$?"
```

- Recorded baseline: `exit=0`
- Target: `exit=0`, and AC-GDR-012 at `0`

**AC-GDR-012** — Aggregate emission across all 12 sweep-target locale files reaches `0`. This is the integration criterion the four compounds above reference.

```bash
# AC-GDR-012 (amended v1.5.0) — single-p= source, liveness + aptness guards
t=0
for name in paired_al auto_mode l7 paired_se handoff; do
  case $name in
    paired_al) p='`/moai loop`[ ·/]+`/goal`'; w() { grep -ohE "$p" "$1" | wc -l | tr -d ' '; }; f=advanced/autonomous-loops.md;;
    auto_mode) p='auto mode.*`/goal`';        w() { grep -c "$p" "$1"; };                      f=advanced/autonomous-loops.md;;
    l7)        p='`/goal`';                   w() { sed -n '7p' "$1" | grep -c "$p"; };        f=advanced/autonomous-loops.md;;
    paired_se) p='`/moai loop`[ ·/]+`/goal`'; w() { grep -ohE "$p" "$1" | wc -l | tr -d ' '; }; f=advanced/self-evolving.md;;
    handoff)   p='`/goal';                    w() { grep -ohF "$p" "$1" | wc -l | tr -d ' '; }; f=cli-reference/handoff.md;;
  esac
  # aggregate the working-tree count across all 4 locales
  c=$(for l in en ja ko zh; do w docs-site/content/$l/$f; done | paste -sd+ - | bc)
  t=$((t + c))
  # liveness: each detector matches non-zero content against the immutable base in all 4 locales
  m=$(for l in en ja ko zh; do git show e306e21a9:docs-site/content/$l/$f > /tmp/gdr-live.md; w /tmp/gdr-live.md; done | sort -n | head -1)
  # aptness: the single pattern source $p carries a literal /goal token
  case "$p" in *'/goal'*) a=1;; *) a=0;; esac
  printf "%s:count=%s,live_min=%s,apt=%s " "$name" "$c" "$m" "$a"
done
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
| REQ-GDR-004 (per-locale baselines) | AC-GDR-001..005 | Every emission criterion records per locale |
| REQ-GDR-009 (symmetry, content-conditioned) | AC-GDR-010 (a) + §A.3 | Component (a) plus the named-exemption carve-out |
| REQ-GDR-010 (liveness floor) | AC-GDR-010 (b) | Detector run against the immutable base `e306e21a9` |
| REQ-GDR-011 (aptness) | AC-GDR-010 (d) | Single `p=` pattern source asserted to carry a literal `/goal` token |
| REQ-GDR-005 (retention register) | AC-GDR-006, AC-GDR-007, AC-GDR-008 | One guard per register row |
| REQ-GDR-006 (do not modify CC / contrast / archives) | AC-GDR-008 | Five pinned counts |
| REQ-GDR-007 (split surface pinned, not zeroed) | AC-GDR-006 | Per-locale structural pins |
| REQ-GDR-008 (no locale-symmetry manufacture) | AC-GDR-008 (`hooks=2`), AC-GDR-009 (inventory) | Two independent guards: adding a line fails the first, adding a page fails the second |

Coverage: **11 / 11 REQs** cited by ≥ 1 AC. Reverse: every AC-GDR-001..012 appears in ≥ 1 row. AC-GDR-010's four components are cited separately, since each answers a different requirement.

---

## §F Edge Cases

- **Sweeping `autonomous-loops.md` to zero.** Blocked by AC-GDR-006's per-locale `h3=1,h2=1,row=1` pins.
- **Editing `en` only.** Blocked twice: the per-locale targets in AC-GDR-001..005, and AC-GDR-010's `distinct=1` requirement.
- **Re-introducing a prose-anchored detector.** Blocked by AC-GDR-010, which fails at `distinct=2` even when the detector's own aggregate reads `0`.
- **Substituting a live, symmetric, semantically-wrong detector** — swapping an emission anchor for an adjacent token (`ac_converge`), or dropping the `` `/goal` `` half of a compound regex. Blocked by AC-GDR-010 (d) at `apt=0`; (a) and (b) both pass in this case, which is why (d) exists.
- **Rewriting the strategy record instead of annotating it.** Blocked by AC-GDR-007's `25` content pin.
- **Adding `ja`/`zh` `hooks-reference.md` lines** to force symmetry. Blocked by AC-GDR-008's `hooks=2` pin.
- **Creating a new locale page.** Blocked by AC-GDR-009's inventory pin.
- **Touching `claude-code/**`.** Blocked by AC-GDR-008's `cc=80` pin.
- **Specifying `handoff.md` before the parent's M7 lands.** Blocked by the §C dependency gate at `0`.
- **Breaking a shortcode while sweeping.** Blocked by AC-GDR-011's build half.
