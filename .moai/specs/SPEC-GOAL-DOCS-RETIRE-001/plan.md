---
id: SPEC-GOAL-DOCS-RETIRE-001
title: Retire native /goal emission references from public and internal documentation across four locales
version: 1.5.0
status: in-progress
created: 2026-07-25
updated: 2026-07-27
author: manager-spec
priority: MEDIUM
phase: "v3.1.0"
module: docs
lifecycle: spec-anchored
tags: "goal, docs-site, i18n, locale-parity, split-surface"
tier: S
depends_on: [SPEC-GOAL-SURFACE-UNIFY-001]
amendment_of: SPEC-GOAL-DOCS-RETIRE-001
run_commit_sha: 24c84c56e
sync_commit_sha: 2a12e2b7d9aee1b5cdfbfba31d6b28ab5d7312b8
---

## §A Context

This is the **plan-phase artifact** for the **in-place amendment** of `SPEC-GOAL-DOCS-RETIRE-001` (D2 debt — aggregate liveness/aptness guard). The amendment is declared in `spec.md` § Amendments (Amendment 1, v1.5.0). This plan describes HOW the AC-GDR-012 judgment-command block will be rewritten in run-phase. It is NOT the edit itself — run-phase owns `acceptance.md`.

The original SPEC (v1.0.0 → v1.4.0) is `completed`: the 12 sweep-target locale files have been swept, the four retention surfaces are pinned, and `total=0` holds against the working tree. The amendment does NOT re-open any of that work. It touches ONLY the `acceptance.md` AC-GDR-012 judgment-command block — refactoring its 5 inline-literal `t=$((t + ...))` lines into the same single-`p=`-source structure AC-GDR-010 uses, and adding liveness + aptness guards to the aggregate path so a future weakening of an inline pattern cannot disguise a fake `total=0` as "sweep complete".

### §A.1 Why this is an amendment, not a new SPEC

Three properties make this an in-place amendment (per `spec-frontmatter-schema.md` § Status Transition Ownership Matrix row `completed → in-progress (amendment)`):

1. **The defect is in the SPEC's own acceptance criterion**, not in the implementation. The 12 sweep-target files are correctly swept; `total=0` is true. The defect is that AC-GDR-012's *judgment command* is structurally weaker than AC-GDR-010's, so a future edit could weaken it further without tripping the existing guards.
2. **The closure is local to one AC body**. Unlike a successor SPEC that carries new scope, this amendment touches ONE acceptance criterion's shell block plus the spec.md amendment declaration. No new requirements, no new files, no new detector anchors.
3. **The recorded baseline and target are unchanged**. The refactor preserves `total=24` (pre-sweep baseline, against immutable base `e306e21a9`) and `total=0` (post-sweep target, against working tree) verbatim. Only the *structure* of the judgment command changes — not the measured result.

### §A.2 The D2 defect (verified this session)

AC-GDR-010 (`acceptance.md:199-211`) adopted the single-`p=`-source discipline at audit iteration 2 (finding B2-1). Each detector declares its pattern ONCE in a `p=` variable that both the counting function `w()` and the aptness assertion read:

```bash
for name in paired_al auto_mode l7 paired_se handoff; do
  case $name in
    paired_al) p='`/moai loop`[ ·/]+`/goal`'; w() { grep -ohE "$p" "$1" | wc -l | tr -d ' '; }; f=advanced/autonomous-loops.md;;
    auto_mode) p='auto mode.*`/goal`';        w() { grep -c "$p" "$1"; };                      f=advanced/autonomous-loops.md;;
    l7)        p='`/goal`';                   w() { sed -n '7p' "$1" | grep -c "$p"; };        f=advanced/autonomous-loops.md;;
    paired_se) p='`/moai loop`[ ·/]+`/goal`'; w() { grep -ohE "$p" "$1" | wc -l | tr -d ' '; }; f=advanced/self-evolving.md;;
    handoff)   p='`/goal';                    w() { grep -ohF "$p" "$1" | wc -l | tr -d ' '; }; f=cli-reference/handoff.md;;
  esac
  ...
  case "$p" in *'/goal'*) a=1;; *) a=0;; esac   # aptness reads the SAME $p
done
```

AC-GDR-012 (`acceptance.md:238-246`) — the aggregate — still carries its OWN inline copies of the same 5 patterns as literals, NOT sourcing from any shared `p=`:

```bash
t=0
t=$((t + $(grep -rhoE '`/moai loop`[ ·/]+`/goal`' docs-site/content/*/advanced/autonomous-loops.md | wc -l | tr -d ' ')))
t=$((t + $(grep -rhoE '`/moai loop`[ ·/]+`/goal`' docs-site/content/*/advanced/self-evolving.md | wc -l | tr -d ' ')))
t=$((t + $(grep -rhc 'auto mode.*`/goal`' docs-site/content/*/advanced/autonomous-loops.md | paste -sd+ - | bc)))
t=$((t + $(for l in en ja ko zh; do sed -n '7p' docs-site/content/$l/advanced/autonomous-loops.md | grep -c '`/goal`'; done | paste -sd+ - | bc)))
t=$((t + $(grep -rhoF '`/goal' docs-site/content/*/cli-reference/handoff.md | wc -l | tr -d ' ')))
echo "total=$t"
```

Recorded baseline: `total=24` (paired_al 4 + paired_se 8 + auto_mode 4 + l7 4 + handoff 4) against the immutable base `e306e21a9`. Target: `total=0` against the working tree. Confirmed this session: AC-GDR-012's command exists ONLY inside `acceptance.md` — no external verify script under `scripts/`, `.claude/`, or `.moai/` runs it.

**Regression hazard (the D2 defect)**: AC-GDR-010's aptness guard inspects ONLY its own `p=`. If one of AC-GDR-012's inline patterns is weakened (e.g. the `` `/goal` `` token dropped during a future edit), the aggregate `total` drops and a fake `total=0` ("sweep complete" disguise) becomes possible — and AC-GDR-010 would NOT catch it because its own pattern is intact. The single-source discipline must be extended to the aggregate path.

---

## §B Known Issues / Risks

- **B1 — AC-GDR-010 and AC-GDR-012 compute different shapes from the same 5 detectors.** AC-GDR-010 emits per-detector `distinct=X,live_min=Y,apt=Z` (the meta-guard components: symmetry, liveness, aptness). AC-GDR-012 emits a single aggregate `total=N`. The refactor MUST preserve both shapes — the case block iterates the same 5 detectors, but each AC computes its own output from the iteration. Do NOT collapse AC-GDR-012 into a copy of AC-GDR-010; the aggregate output is what the four compound ACs (AC-GDR-006/008/009/011) reference via `AC-GDR-012 at 0`.
- **B2 — `paired_al` and `paired_se` share a pattern but target different files.** Both detectors use `p='\`/moai loop\`[ ·/]+\`/goal\`'`, but `paired_al` targets `autonomous-loops.md` and `paired_se` targets `self-evolving.md`. The case block MUST preserve this file mapping (`f=advanced/autonomous-loops.md` vs `f=advanced/self-evolving.md`); conflating them would either double-count or zero-count one of the two surfaces.
- **B3 — AC-GDR-012's current block aggregates over ALL locales at once (`docs-site/content/*/...`), whereas AC-GDR-010's liveness check iterates per-locale.** The aggregate's new liveness check must mirror AC-GDR-010's per-locale discipline (run the detector against the immutable base `e306e21a9` in each of the four locales and take the minimum), not collapse to a single locale-agnostic count — otherwise a pattern that breaks in one locale but not the other three would pass the aggregate liveness check while the per-locale AC-GDR-010 still flags it.
- **B4 — The aggregate output line `echo "total=$t"` is referenced by AC-GDR-006/008/009/011 as the compound condition `AC-GDR-012 at 0`.** The refactor MUST preserve the literal output format `total=$t` (with `t` carrying the 5-detector sum). Renaming the variable or changing the output format silently breaks the four compound ACs.
- **B5 — The refactor changes STRUCTURE, not the measured result.** After the refactor, re-running AC-GDR-012 against the immutable base `e306e21a9` MUST still observe `total=24`; against the current working tree MUST still observe `total=0`. Any refactor that produces a different number against either surface is wrong.
- **B6 — Prose-only constraint.** The refactor touches ONLY the AC-GDR-012 judgment-command block in `acceptance.md`. It MUST NOT modify spec.md §B requirements, AC-GDR-001..011 body content, plan.md milestones N1-N5, or any of the 12 sweep-target locale files.
- **B7 — `/tmp/gdr-live.md` shared-path overwrite (NIT, accepted).** The §F.2 liveness loop writes `git show e306e21a9:... > /tmp/gdr-live.md` per-locale inside a single `$()`; sequential invocation is safe (each iteration's `w` reads its own just-written file). The same path is used by AC-GDR-010 (`acceptance.md:208`) — concurrent execution of AC-GDR-010 and AC-GDR-012 (or any parallel verifier) would collide on `/tmp/gdr-live.md`. The ACs are designed to run independently at judgment time (no concurrent execution is a use case the SPEC exercises); the shared path is accepted as defense-in-depth adequate for Tier S. A future hardening MAY switch to per-AC unique paths (`/tmp/gdr-live-010.md` vs `/tmp/gdr-live-012.md`) or `mktemp`; the §F.2 illustration and the M1 acceptance.md code keep `/tmp/gdr-live.md` to preserve structural parallelism with AC-GDR-010.
- **B8 — Glob → explicit-locale-list semantic shift (NIT, bounded).** The v1.4.0 AC-GDR-012 used a shell glob `docs-site/content/*/advanced/autonomous-loops.md` (would include any future locale directory); the v1.5.0 §F.2 refactored form uses an explicit `for l in en ja ko zh` list (excludes any future locale). For the current 4-locale state both forms produce identical counts (verified `total=24` base / `total=0` tree at plan-audit iteration 1). If a 5th locale directory is added in the future, the glob form would include it (raising `total` above 24); the explicit-list form would not. REQ-GDR-008 ("shall not create new locale pages") bounds this risk — the future-locale divergence is latent, not live, and the explicit list is the safer of the two forms for forward-compatibility.

---

## §C Pre-flight

The run-phase implementer runs these BEFORE editing AC-GDR-012:

```bash
# 1. Current branch + HEAD (Route B PR — repo-local policy, enforce_admins: true)
git branch --show-current
git rev-parse HEAD

# 2. Divergence from origin/main (must be 0 0 or local-ahead-only — parallel-race guard)
git fetch origin main
git rev-list --count --left-right origin/main...HEAD

# 3. AC-GDR-012 baseline re-measurement against the immutable base (pre-sweep state)
# Expected: total=24 (paired_al 4 + paired_se 8 + auto_mode 4 + l7 4 + handoff 4)
for l in en ja ko zh; do
  git show e306e21a9:docs-site/content/$l/advanced/autonomous-loops.md > /tmp/gdr-base-al-$l.md
  git show e306e21a9:docs-site/content/$l/advanced/self-evolving.md    > /tmp/gdr-base-se-$l.md
  git show e306e21a9:docs-site/content/$l/cli-reference/handoff.md     > /tmp/gdr-base-ho-$l.md
done
t=0
t=$((t + $(grep -rhoE '`/moai loop`[ ·/]+`/goal`' /tmp/gdr-base-al-*.md | wc -l | tr -d ' ')))
t=$((t + $(grep -rhoE '`/moai loop`[ ·/]+`/goal`' /tmp/gdr-base-se-*.md | wc -l | tr -d ' ')))
t=$((t + $(for l in en ja ko zh; do grep -c 'auto mode.*`/goal`' /tmp/gdr-base-al-$l.md; done | paste -sd+ - | bc)))
t=$((t + $(for l in en ja ko zh; do sed -n '7p' /tmp/gdr-base-al-$l.md | grep -c '`/goal`'; done | paste -sd+ - | bc)))
t=$((t + $(grep -rhoF '`/goal' /tmp/gdr-base-ho-*.md | wc -l | tr -d ' ')))
echo "base_total=$t"  # MUST be 24

# 4. AC-GDR-012 current-tree measurement (post-sweep state landed at v1.4.0)
# Expected: total=0
t=0
t=$((t + $(grep -rhoE '`/moai loop`[ ·/]+`/goal`' docs-site/content/*/advanced/autonomous-loops.md | wc -l | tr -d ' ')))
t=$((t + $(grep -rhoE '`/moai loop`[ ·/]+`/goal`' docs-site/content/*/advanced/self-evolving.md | wc -l | tr -d ' ')))
t=$((t + $(grep -rhc 'auto mode.*`/goal`' docs-site/content/*/advanced/autonomous-loops.md | paste -sd+ - | bc)))
t=$((t + $(for l in en ja ko zh; do sed -n '7p' docs-site/content/$l/advanced/autonomous-loops.md | grep -c '`/goal`'; done | paste -sd+ - | bc)))
t=$((t + $(grep -rhoF '`/goal' docs-site/content/*/cli-reference/handoff.md | wc -l | tr -d ' ')))
echo "tree_total=$t"  # MUST be 0

# 5. Confirm no external verify script runs the AC-GDR-012 patterns (gap-closure assumption still holds)
grep -rn 'grep -rhoE .\`/moai loop\`' scripts/ .claude/ .moai/ 2>/dev/null | grep -v 'RETIRE-001' || echo "no external callers"
```

---

## §D Constraints

- **D1 [HARD] Single-`p=`-source discipline within AC-GDR-012.** Each detector declares its pattern ONCE in a `p=` variable that BOTH the counting function `w()` AND the aptness assertion read. A second inline copy of the pattern anywhere in the aggregate block is a defect — it is exactly the drift surface the amendment closes.
- **D2 [HARD] Byte-identical pattern literals across AC-GDR-010 and AC-GDR-012.** The 5 pattern values in AC-GDR-012's `case` block MUST be byte-identical to AC-GDR-010's (`acceptance.md:199-211` is the authoritative source): `paired_al` = `` `/moai loop`[ ·/]+`/goal` `` (targeting `autonomous-loops.md`); `paired_se` = `` `/moai loop`[ ·/]+`/goal` `` (targeting `self-evolving.md` — same pattern, different file); `auto_mode` = `auto mode.*`/goal``; `l7` = `` `/goal` ``; `handoff` = `` `/goal`` (opening backtick only, no closing — matches the original `-ohF` fixed-string detector).
- **D3 [HARD] Liveness assertion (aggregate mirror of AC-GDR-010 component b).** Each detector, run against the immutable base `e306e21a9`, MUST match non-zero content in ALL FOUR locales. The minimum across locales MUST be `>= 1`. A detector that reads zero against the base is broken (regex drift), not "swept" — the aggregate liveness floor is what separates the two at the aggregate level.
- **D4 [HARD] Aptness assertion (aggregate mirror of AC-GDR-010 component d).** Each detector's pattern MUST carry a literal `/goal` token, read from the SAME `p=` source as the counting function. A pattern that drops the `` `/goal` `` half (e.g. weakening `` `/moai loop`[ ·/]+`/goal` `` to `` `/moai loop` ``) MUST be rejected at `apt=0`. This is the specific hole the amendment closes — a weakened inline pattern can no longer hide behind a stale assertion.
- **D5 [HARD] Baseline-preservation invariant (load-bearing).** After the refactor, re-running AC-GDR-012 against the immutable base `e306e21a9` MUST still observe `total=24` (recorded baseline); against the current working tree MUST still observe `total=0` (recorded target). The refactor changes STRUCTURE only, not the measured result. Re-measurement command: see §C pre-flight steps 3-4 (run identically post-refactor).
- **D6 [HARD] Output-format preservation.** The aggregate's final output line MUST remain `echo "total=$t"` — the four compound ACs (AC-GDR-006/008/009/011) reference `AC-GDR-012 at 0` and parse `total=0` from this literal output. Renaming the variable or changing the format silently breaks the compounds.
- **D7 [HARD] Prose-only amendment scope.** Touch ONLY `acceptance.md` AC-GDR-012 (judgment-command block, lines ~238-246). Do NOT modify: spec.md §B requirements; AC-GDR-001 through AC-GDR-011 body content; plan.md milestones N1-N5; the 12 sweep-target locale files; `progress.md` (run-phase owns §E.2/§E.3, sync-phase owns §E.4); `run_commit_sha` / `sync_commit_sha` provenance fields; `tier:` / `amendment_of:` / `version:` / `status:` frontmatter fields (the plan-phase already set them).
- **D8 [HARD] Repo-local PR policy + Late-Branch closure.** This repo has `enforce_admins: true` — ALL tiers use Route B (PR). Do NOT push directly to `main`. Commit the run-phase change on the current `main` checkout; manager-git creates the feature branch + PR at PR-creation time. Conventional commit subject per the Status Transition Ownership Matrix: `feat(SPEC-GOAL-DOCS-RETIRE-001): in-place amendment D2 aggregate liveness/aptness guard`. Pathspec-stage the commit (only the SPEC files touched); never `git add -A` (shared main checkout discipline).
- **D9 [HARD] Cross-AC verdict consistency.** The aggregate's new liveness + aptness guards MUST be consistent with AC-GDR-010's verdicts, not contradictory. The 3 control detectors AC-GDR-010 validates — (i) hyphenated `auto-mode` dead detector (rejected by liveness at `live_min=0`); (ii) `ac_converge` adjacent token (rejected by aptness at `apt=0`); (iii) weakened `auto mode` half (rejected by aptness at `apt=0`) — MUST also be rejected by AC-GDR-012's guards if substituted for a real detector. The aggregate's guards do not get to contradict the meta-guard.
- **D10** No time estimates.

---

## §E Self-Verification

The run-phase implementer returns the E1-E7 deliverables in the 5-section Evidence-Bearing format per `verification-claim-integrity.md` §3 (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk):

- **E1 — AC-GDR-012 baseline-preservation re-measurement (load-bearing).** Run the refactored AC-GDR-012 block against BOTH the immutable base `e306e21a9` AND the current working tree. Report the verbatim command + output:
  - Against base `e306e21a9`: `total=24` (MUST equal the recorded baseline; decomposition `paired_al 4 + paired_se 8 + auto_mode 4 + l7 4 + handoff 4 = 24` MUST hold).
  - Against current tree: `total=0` (MUST equal the recorded target).
- **E2 — Liveness floor (per-detector, against base).** Each of the 5 detectors MUST yield `live_min >= 1` against `e306e21a9` in all four locales. Report per-detector: `paired_al:N paired_se:N auto_mode:N l7:N handoff:N` (N is the minimum count across the four locales).
- **E3 — Aptness (per-detector, from `p=`).** Each of the 5 detectors MUST yield `apt=1`. Report per-detector: `paired_al:apt=1 paired_se:apt=1 auto_mode:apt=1 l7:apt=1 handoff:apt=1`. The aptness check reads the SAME `p=` variable that the counting function `w()` reads — single source.
- **E4 — Cross-AC consistency with AC-GDR-010 (the meta-guard).** Re-run AC-GDR-010 against the current tree and confirm the 5 detectors still yield `distinct=1,live_min>=1,apt=1` (the meta-guard components). AC-GDR-010 was NOT modified by this amendment — its verdict MUST be unchanged. If it shifts, the amendment has a side effect.
- **E5 — Output-format preservation.** The aggregate's final `echo "total=$t"` line is byte-identical to the v1.4.0 form. Confirm via `grep -c 'echo "total=$t"' .moai/specs/SPEC-GOAL-DOCS-RETIRE-001/acceptance.md` (exactly one match, in the AC-GDR-012 block).
- **E6 — Scope discipline.** `git diff --name-only origin/main...HEAD` returns exactly one file: `.moai/specs/SPEC-GOAL-DOCS-RETIRE-001/acceptance.md`. If the diff includes any other file, the run-phase made an out-of-scope edit — return a blocker (do NOT silently expand scope).
- **E7 — Blocker report (if any).** If the refactor cannot preserve `total=24` against base or `total=0` against tree, OR if AC-GDR-010's verdict shifts, return a structured blocker report (NEVER silently adjust the baseline or weaken the invariant).

---

## §F Milestones

Tier S — single milestone. Ordered by decision-reversibility (the highest-change-likelihood decision is the shared-pattern mechanism choice, surfaced first in §F.1 below).

### §F.1 Shared-pattern mechanism — single `case` block with byte-identical literals across ACs (decision)

The brief offered two mechanisms: (i) one `case` expansion shared across both ACs, or (ii) a single sourced variable both ACs reference. **Choice: (i), realized as two structurally-identical `case` blocks (one per AC) whose 5 pattern literals are byte-identical.** Each AC carries its own `case` block; the cross-AC parity is a defense-in-depth property (constraint D2), not a runtime mechanism.

**Why not a single sourced variable (option ii)?** A sourced variable would require either:
- (a) A shared shell function defined once and sourced by both ACs — but AC commands are self-contained shell snippets embedded in markdown code blocks, and the gap-closure check confirmed AC-GDR-012's command exists ONLY inside `acceptance.md`. Introducing an external sourced file (under `scripts/` or `.claude/`) would change that property and add a new template-mirror + drift surface. Rejected.
- (b) A shell function defined at the top of `acceptance.md` and re-invoked by both AC code blocks — but AC code blocks are independently runnable (each must be pasteable into a shell and produce its recorded output). The function would have to be re-defined inside every AC's code block, defeating the "single source" goal. Rejected.

**Why is the duplicated-case-block approach safe?** The "single source" discipline operates at TWO levels, both of which close the D2 defect:

1. **Within each AC** — the `case $name in ...; p=...; w() {...; }` block is the single source for THAT AC. The counting function `w()` and the aptness check both read the SAME `$p`. A weakening of `$p` cannot leave a stale assertion behind because the assertion reads the same variable. (This is the closure AC-GDR-010 already adopted at iteration 2; the amendment extends it to AC-GDR-012.)
2. **Across the two ACs** — the 5 pattern literals are byte-identical (constraint D2). A future edit that weakens ONE AC's `case` block while leaving the other intact is now detectable at BOTH levels: AC-GDR-010's verdict would shift if its own `case` block is touched (the meta-guard catches it there), and AC-GDR-012's NEW aptness guard catches a weakened `$p` in its own `case` block. Either edit is caught; the cross-AC parity is defense-in-depth.

The mechanism choice is the most reversible decision in this plan — switching to option (ii) later is a local refactor that does not change any recorded baseline or target.

**Defense-in-depth is acceptable for Tier S; a parity lint may be a follow-up SPEC.** The within-AC single-source (`p=` read by both `w()` and the aptness check) is the primary D2 closure — a weakened `$p` cannot leave a stale assertion behind because the assertion reads the same variable. The cross-AC byte-identity (constraint D2) is the secondary defense: it ensures the two ACs cannot drift apart, but a drift would only matter if BOTH ACs were independently weakened in the same way (a two-step regression). A future SPEC MAY propose a mechanical parity lint that extracts the 5 pattern literals from both AC case blocks (acceptance.md:199-211 and the AC-GDR-012 case block) and diffs them at CI time; until then, the parity is author-discipline + audit-time verification (verified `diff exit=0` at plan-audit iteration 1 against commit `449c7cb28`).

### §F.2 Milestone M1 — AC-GDR-012 judgment-command refactor

**Priority High.** Owns `acceptance.md` AC-GDR-012 block (lines ~238-246). Single edit; single commit. **Run-phase owner: `manager-spec`** (per the D-NEW-1 inline-fix pattern — `.claude/rules/moai/development/spec-frontmatter-schema.md` § Forbidden ownership crossings forbids `manager-develop` from modifying `acceptance.md` body content; the `completed → in-progress (amendment)` row of the Status Transition Ownership Matrix names `manager-spec` as the owner via re-delegation). When run-phase reveals a need to modify the AC body, `manager-develop` MUST return a blocker report and the orchestrator re-delegates to `manager-spec`.

The 5 inline-literal `t=$((t + ...))` lines are refactored into the form below. `manager-spec` applies this edit to `acceptance.md`; the literal pattern values come from AC-GDR-010's `case` block at `acceptance.md:199-211` (the authoritative source, byte-identical per constraint D2):

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

**Cross-check against AC-GDR-010's 3 control detectors (constraint D9).** The aggregate's new guards MUST reject all three controls if substituted for a real detector (the run-phase implementer verifies this by temporarily substituting each control into the `case` block and confirming the guard fires; the substitution is reverted before commit):

| Control detector | `live_min` | `apt` | Rejected by |
|---|---|---|---|
| `auto-mode` (hyphenated typo) | **`0`** | — | (b) liveness — `live_min=0` against base |
| `ac_converge` (adjacent token, co-occurs with `` `/goal` `` on the same base line) | `1` | **`0`** | (d) aptness — pattern lacks `/goal` token |
| `auto mode` (compound regex with `` `/goal` `` half dropped) | `1` | **`0`** | (d) aptness — pattern lacks `/goal` token |

The aggregate's verdicts MUST be consistent with AC-GDR-010's verdicts for the same 3 controls (AC-GDR-010's table is at `acceptance.md` lines 218-222).

**Recorded baseline (preserved verbatim)**: `total=24` against base `e306e21a9` (paired_al 4 + paired_se 8 + auto_mode 4 + l7 4 + handoff 4); `total=0` against current working tree. The refactor changes STRUCTURE only; the measured results are unchanged.

**Output-format preservation (constraint D6)**: the final `echo "total=$t"` line is byte-identical to the v1.4.0 form, so AC-GDR-006/008/009/011's compound references to `AC-GDR-012 at 0` continue to parse correctly. **The amended block's output is additive, not replacing** — the per-detector `printf "%s:count=%s,live_min=%s,apt=%s "` line (new) precedes the preserved `echo "total=$t"` line (final). The v1.4.0 form emitted ONLY `total=N`; the v1.5.0 form emits the per-detector lines followed by `total=N`. Three properties hold: (i) AC-GDR-010 uses the label `distinct=` (NOT `count=`) — no collision; (ii) the four compound ACs (006/008/009/011) parse `total=0` unanchored (e.g. `grep -o 'total=0'`) and the final line still carries it; (iii) no compound AC parses the new per-detector fields, so the additive output is forward-compatible.

ACs affected: AC-GDR-012 (refactored); AC-GDR-006/008/009/011 (compound references — unchanged, the output format is preserved).

---

## §G Anti-Patterns

- **Weakening an inline aggregate pattern to drive `total` toward 0.** The whole point of the amendment. Blocked by AC-GDR-012's new aptness guard (`apt=0` rejects a pattern without a literal `/goal` token) and liveness floor (`live_min=0` rejects a broken regex against the base).
- **Conflating `paired_al` and `paired_se` because they share a pattern.** They target different files (`autonomous-loops.md` vs `self-evolving.md`); conflating them double-counts one surface and zeros the other. The `case` block MUST preserve the `f=` mapping.
- **Collapsing AC-GDR-012's per-locale liveness check into a locale-agnostic `grep -rho ... | wc -l` against the base.** A pattern that breaks in one locale but not the other three would pass the locale-agnostic check; the per-locale `for l in en ja ko zh; do ... done | sort -n | head -1` discipline (mirroring AC-GDR-010 component b) MUST be preserved.
- **Changing the output format `echo "total=$t"` to "improve" it.** Silently breaks AC-GDR-006/008/009/011's compound reference to `AC-GDR-012 at 0`. The output literal is load-bearing.
- **Modifying AC-GDR-010 to "share" its `case` block with AC-GDR-012.** AC-GDR-010 is the meta-guard; touching it broadens the amendment scope and risks destabilizing the meta-guard verdict. The amendment touches ONLY AC-GDR-012; AC-GDR-010 remains as-is and its verdict MUST be unchanged (constraint D9, E4).
- **Treating the refactor as "just renaming" without re-measuring against the base.** The baseline-preservation invariant (constraint D5) is the load-bearing check that the refactor is behavior-preserving. Skipping the re-measurement violates `verification-claim-integrity.md` §1.1 surface 2 (manager-agent §E self-verification).
- **Using `git add -A` to stage the run-phase commit.** Shared main checkout discipline (CLAUDE.local.md §23 + `main-checkout-branch-guard.md`); pathspec-stage only the SPEC file touched.

---

## §H Cross-References

- `spec.md` § Amendments — the amendment declaration (Amendment 1, v1.5.0; prior completed version 1.4.0; prior_completed_sha `760f09f73`; rationale; scope).
- `spec.md` §B REQ-GDR-010 / REQ-GDR-011 — the liveness floor and aptness requirements AC-GDR-010 operationalizes and the amendment extends to AC-GDR-012.
- `acceptance.md` §B AC-GDR-010 (lines 199-211) — the meta-guard whose single-`p=`-source discipline (adopted at audit iteration 2, finding B2-1) is being extended to AC-GDR-012; the authoritative source for the 5 pattern literals (byte-identical per constraint D2).
- `acceptance.md` §B AC-GDR-012 (lines 236-249) — the aggregate emission integration criterion (the refactor target).
- `acceptance.md` §A.2 — the locale-invariance rule and the 5 anchor table (the pattern values are byte-identical to this table).
- `acceptance.md` §E — REQ ↔ AC traceability matrix (AC-GDR-012 covers REQ-GDR-001; the amendment does not change the traceability).
- `progress.md` §E.1 — the plan-phase audit-ready signal (re-issued at v1.5.0; §E.2/§E.3 owned by manager-develop run-phase, §E.4 owned by manager-docs sync-phase).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix — row `completed → in-progress (amendment)` (the canonical authority for this transition; commit subject `feat(SPEC-{ID}): in-place amendment <rationale-summary>`).
- `.claude/rules/moai/workflow/repo-local-pr-policy.md` — Route B (PR) is mandatory for ALL tiers in this repo (`enforce_admins: true`).
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 2 — the manager-agent §E self-verification matrix that E1-E7 above must satisfy (every claim row attributable to a directly-observed command whose verbatim output is cited).
