# SPEC Review Report: SPEC-GOAL-DOCS-RETIRE-001 (SPEC-B)

Iteration: 1/3 (first audit)
Verdict: **FAIL** (no must-pass failure; two MUST-FIX in the structural centrepiece)
Overall Score: **0.75** (harmonic mean; Tier **M** threshold 0.80)

Reasoning context ignored per M1 Context Isolation. Commands executed in `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/goal-retire` at `HEAD = cd7e3c759`.

---

## Must-Pass Results

| | Result | Basis |
|---|---|---|
| MP-1 REQ number consistency | **PASS** | `REQ-GDR-001 … 008` contiguous (8), `AC-GDR-001 … 012` contiguous (12); `sort \| uniq -d \| wc -l` → `0` for both. Fully contiguous, unlike the parent |
| MP-2 GEARS compliance | **PASS** | `grep -cE '^- \*\*REQ-GDR-[0-9]{3}\*\*.*shall' spec.md` → `8` of `8`. Patterns: Event-driven (001, 004), State-driven (002), Ubiquitous (003), Where/capability-gate (005, 007), Unwanted (006, 008). No legacy IF/THEN |
| MP-3 frontmatter validity | **PASS** | `spec lint --strict spec.md` → `✓ No findings — all SPEC documents are valid`; repo-wide `spec lint` cites this SPEC `0` times. `tags` is the string form; all 12 canonical fields present; `tier: M`, `depends_on: [SPEC-GOAL-SURFACE-UNIFY-001]` |
| MP-4 language neutrality | N/A | `grep -ciE 'gopls\|pylsp\|rust-analyzer' spec.md` → `0`. Docs-scoped, no tooling claim |
| MP-5 D7 cross-SPEC | **PASS** | Referenced ID: `SPEC-GOAL-SURFACE-UNIFY-001`. `.moai/specs/SPEC-GOAL-SURFACE-UNIFY-001/spec.md` exists; `grep '^status:'` → `status: draft`, not in {retired, superseded, archived} → no reconciliation obligation, no BLOCKING finding |
| MP-6 D8 cross-platform | **PASS** | `grep -c syscall` → `0` across all six artifacts |
| MP-7 clarification gate | **PASS** | `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → none |

**Operational note on `depends_on` (not a defect).** Per `spec-workflow.md` § Depends_on Pre-flight Check, fulfilment is defined strictly as the dependency's `status: completed`; `draft` is unfulfilled. SPEC-B therefore cannot enter run/sync phase until the parent completes, and a `/moai run` attempt will correctly surface the 3-option blocker (wait / `--ignore-deps` with logged rationale / abort). This is correct sequencing — the dependency is concrete, not organizational (`spec.md` §A.2: `docs-site/content/en/cli-reference/handoff.md:34` mirrors `internal/cli/handoff.go:104`, which the parent's M7 rewrites) — and SPEC-B's own §C gate records it (`grep -c 'a /moai goal condition' internal/cli/handoff.go` → `0` at baseline, reproduced).

---

## Tier M assessment — the classification is CORRECT

The author's reasoning was tested against `spec-workflow.md` § SPEC Complexity Tier rather than accepted:

| Tier M criterion | Measured | Verdict |
|---|---|---|
| Files affected 5-15 | `sed -n '/§F.1/,/^---$/p' plan.md` → `4+4+4+1+0 = 13`, matching the stated total | Inside the ceiling |
| Scope 300-1000 LOC | 24 emission markers across 13 doc files (AC-GDR-012 baseline `total=24`, reproduced) | Well under |
| Not constitutional | Documentation only; no rule, agent, hook, or Go change | Not constitutional |

**"One verification regime" — I agree, and the author's phrasing is the right test.** "A grep repeated per locale is the same regime four times, not four regimes" holds: 11 of the 12 criteria are grep/`ls`, and the twelfth is a single `hugo` build. Compare the parent's three genuinely distinct regimes — grep on doctrine prose, a Go RED-GREEN TDD cycle that must author its own missing test, and byte-identical mirror parity across 14 pairs. **None of those shapes exists here**, which is the substantive difference, not the file count.

**Does the four-locale split-surface problem warrant Tier L? No.** It is *conceptually* hard — it defeated the parent's acceptance criteria — but Tier is determined by scope, file count, and constitutional status, not by difficulty. Conceptual difficulty is answered by criterion *quality*, and this SPEC answers it well (per-locale baselines on every emission criterion plus a meta-guard). Inflating the tier would buy nothing: the only mechanical consequences are the artifact set and the PASS threshold, and SPEC-B already delivers the Tier L artifact set voluntarily (see B-6). **Reversibility** also checks out: doc edits are git-revertible, and there is no untested user-visible code path — the parent's third tier driver is genuinely absent here.

Tier M sets the PASS threshold at **0.80**.

---

## AC-GDR-010 stress test — the structural centrepiece has two holes

This is the fix for the parent's N2 and the highest-value thing to break, so I attacked it directly. **The N2 fix itself works** — I ran both anchors:

```
$ for l in en ja ko zh; do grep -c 'auto mode.*`/goal`' docs-site/content/$l/advanced/autonomous-loops.md; done
en:1 ja:1 ko:1 zh:1          # accepted anchor — symmetric
$ for l in en ja ko zh; do grep -ohE '`/goal`[^.]*per-turn|per-turn[^.]*`/goal`' … | wc -l; done
en:2 ja:0 ko:0 zh:0          # disqualified anchor — asymmetric, exactly as documented
```

The re-anchoring is real and the disqualifying signal reproduces. But `distinct=1` alone does not carry the weight the SPEC places on it. Both attack vectors land:

**Attack (b) — a dead detector passes.** Executed with a deliberately typo'd anchor (`auto-mode`, hyphenated, which co-occurs with `` `/goal` `` nowhere):

```
$ for l in en ja ko zh; do grep -c 'auto-mode.*`/goal`' docs-site/content/$l/advanced/autonomous-loops.md; done
en:0 ja:0 ko:0 zh:0
$ … | sort -u | wc -l
1                             # distinct=1 → PASSES AC-GDR-010
```

A detector matching nothing anywhere yields `distinct=1` and its own emission target `0` — both green. At judgment time (post-sweep) AC-GDR-010 cannot distinguish "content swept" from "regex broken", because both produce `0/0/0/0`. The only thing separating them today is the recorded **non-zero baseline** (`en:1 ja:1 ko:1 zh:1` and `en:2 ja:2 ko:2 zh:2`), which is a plan-phase artifact that AC-GDR-010 does not itself assert. → **B-1**.

**Attack (a) — legitimately asymmetric content would fail a correct detector.** `hooks-reference.md` carries the reference in only two locales:

```
$ for l in en ja ko zh; do grep -ohF '`/goal' docs-site/content/$l/advanced/hooks-reference.md | wc -l; done
en:1 ja:0 ko:1 zh:0
$ … | sort -u | wc -l
2                             # distinct=2 → would FAIL AC-GDR-010
```

That asymmetry is the pre-existing content gap REQ-GDR-008 explicitly forbids closing. No live conflict exists today (the page is a retention surface, absent from AC-GDR-010's five-detector list), but REQ-GDR-003 and REQ-GDR-004 are stated **generally** — "Every emission detector … the per-locale values shall be symmetric" — so the rule as written would reject a correct detector aimed at genuinely locale-asymmetric content, and the SPEC contains a live example of such content. → **B-2**. `grep -ciE 'non-zero|floor|liveness|legitimately asymmetric' acceptance.md` → `0`: neither hole is acknowledged.

---

## Category Scores

| Dimension | Score | Evidence |
|-----------|-------|----------|
| Clarity | 0.75 | §A.1's articulation of *why* this half is structurally different is the strongest prose in either SPEC. Deducted for the `18` vs measured `24` emission-marker contradiction (B-4) and the provenance table's four wrong rows (B-3) |
| Completeness | 0.75 | All sections present; four-row retention register each row guarded; exclusions thorough and reasoned; REQ↔AC matrix 8/8; §C gates include the parent-dependency gate. Deducted because the centrepiece criterion's own specification omits a liveness floor and an asymmetry carve-out (B-1, B-2) |
| Testability | 0.75 | **12 of 12 baselines reproduce verbatim**, every emission criterion carries a per-locale baseline, and the N2 re-anchoring is demonstrably effective (both anchors run above). Deducted for the two demonstrated holes in AC-GDR-010, which is the criterion the whole design rests on |
| Traceability | 0.75 | §E matrix is genuine on spot-check — REQ-GDR-007 (split surface pinned, not zeroed) → AC-GDR-006, whose target is literally "unchanged `h3=1,h2=1,row=1`" per locale, so it would fail on violation. Deducted for the §D provenance table (B-3), which is the artifact binding this SPEC's criteria to the parent's audited baselines |

```
harmonic mean (0.75, 0.75, 0.75, 0.75) = 0.7500 → 0.75
```

Tier M threshold 0.80 → FAIL, narrowly.

### Baseline reproduction — 12 of 12, zero divergence

AC-GDR-001 `en:1 ja:1 ko:1 zh:1` · AC-GDR-002 `en:2 ja:2 ko:2 zh:2` · AC-GDR-003 `en:1 ja:1 ko:1 zh:1` · AC-GDR-004 `en:1 ja:1 ko:1 zh:1` · AC-GDR-005 `en:1 ja:1 ko:1 zh:1` · AC-GDR-006 `en:h3=1,h2=1,row=1` ×4 identical · AC-GDR-007 `0` / `25` · AC-GDR-008 `cc=80 goal.md=12 moai-goal=4 hooks=2 research=4` · AC-GDR-009 `autonomous-loops.md=4 handoff.md=4 self-evolving.md=4` · AC-GDR-010 all five `distinct=1` · AC-GDR-011 `exit=0` · AC-GDR-012 `total=24`.

Every emission criterion is symmetric across locales exactly as claimed — verified per locale, not by inspection. Every criterion FAILS at baseline (the four compounds are paired with AC-GDR-012, which reads `24`, so none is a no-op).

---

## MUST-FIX

**B-1. AC-GDR-010's `distinct=1` admits a dead detector — the meta-guard cannot distinguish "swept" from "broken" at judgment time** — `acceptance.md:167-184` — Severity: **critical** — Command and observed output:

```
$ for l in en ja ko zh; do printf "%s:%s " $l "$(grep -c 'auto-mode.*`/goal`' docs-site/content/$l/advanced/autonomous-loops.md)"; done
en:0 ja:0 ko:0 zh:0
$ for l in en ja ko zh; do grep -c 'auto-mode.*`/goal`' docs-site/content/$l/advanced/autonomous-loops.md; done | sort -u | wc -l
1
```

A typo'd or otherwise dead anchor yields `distinct=1` (passing AC-GDR-010) and `0` aggregate (passing its own emission AC). Since the target state of every emission detector is `0/0/0/0`, `distinct=1` is satisfied by both success and detector death, and AC-GDR-010 asserts no property that separates them. The recorded non-zero baselines are the only thing that currently does, and they are plan-phase artifacts the criterion does not reference. This matters precisely because AC-GDR-010 is billed as making "the N2 defect class mechanically impossible to ship" — a broken detector is a *different* way to ship the same outcome. Required fix: add a **liveness component** asserting each detector still matches the pre-sweep content, e.g. per detector `git show cd7e3c759:<path> | <detector>` → non-zero for at least one locale, so a post-sweep `0` proves "swept" rather than "dead regex". State the non-zero floor explicitly in the criterion text.

**B-2. AC-GDR-010's blanket symmetry rule contradicts REQ-GDR-008's protected pre-existing asymmetry, with a live example in scope** — `acceptance.md:167`, `spec.md:71,73,85` — Severity: **major** — Command and observed output:

```
$ for l in en ja ko zh; do printf "%s:%s " $l "$(grep -ohF '`/goal' docs-site/content/$l/advanced/hooks-reference.md | wc -l | tr -d ' ')"; done
en:1 ja:0 ko:1 zh:0
$ … | sort -u | wc -l
2
$ grep -ciE 'non-zero|floor|liveness|legitimately asymmetric|genuinely asymmetric' acceptance.md
0
```

REQ-GDR-003 ("Every emission detector … shall anchor on a token that survives translation") and REQ-GDR-004 ("the per-locale values shall be symmetric; an asymmetric baseline is evidence the detector is prose-anchored") are stated without exception, and AC-GDR-010 enforces them as "exactly ONE distinct value". But asymmetry has **two** possible causes: a prose-anchored detector (the N2 defect) or genuinely locale-asymmetric content (the `hooks-reference.md` gap that REQ-GDR-008 forbids closing, and that `spec.md:106` names explicitly). The rule conflates them. No live conflict today because `hooks-reference.md` is a retention surface outside AC-GDR-010's detector list — but should any emission target ever carry legitimately asymmetric content, a correct detector would be rejected and the pressure would be to manufacture symmetry, which is exactly what REQ-GDR-008 prohibits. Required fix: add a carve-out — asymmetry is disqualifying only when the underlying *content* is symmetric; where content is genuinely locale-asymmetric, the baseline records the asymmetry with a justification and AC-GDR-010 exempts that detector by name.

---

## SHOULD-FIX

**B-3. The §D provenance table maps 4 of its 7 rows to the wrong criterion** — `spec.md:120-128` — Severity: **major** — Commands: `sed -n '/| This SPEC | Came from | Change |/,/^A \*\*REQ/p' spec.md` for the claims; per-AC first lines from `acceptance.md`; `grep -n 'cc=80 goal.md=12' acceptance.md` → `153`; `grep -n 'hugo --quiet' acceptance.md` → `189`. Observed:

| §D claims | Actual subject of that AC | Verdict |
|---|---|---|
| AC-GDR-001..003 ← AC-GSU-028 emission half | 001/002 paired listings, 003 handoff — all emission | ✓ |
| AC-GDR-004 ← new | L7 mis-attribution | ✓ |
| **AC-GDR-005 ← AC-GSU-028 retention half, "Per-locale pins"** | AC-GDR-005 is **"The N2 fix"** — the auto-mode **emission** detector. The retention half is AC-GDR-006 | ✗ |
| **AC-GDR-006 ← AC-GSU-029, "Unchanged baselines"** | AC-GDR-006 is the split-surface `h3/h2/row` retention pins. AC-GSU-029's baseline (`cc=80 …`) is at line 153, inside **AC-GDR-008** | ✗ |
| AC-GDR-007 ← AC-GSU-032 | superseding note + `25` pin | ✓ |
| **AC-GDR-008..011 ← new** | AC-GDR-008 is **not** new — it carries AC-GSU-029 verbatim | ✗ |
| **AC-GDR-012 ← new, "Held-out docs build"** | AC-GDR-012 is the **aggregate emission** integration criterion; the docs build is **AC-GDR-011** | ✗ |

The table's purpose is to keep this SPEC's criteria bound to the parent's audited baselines ("so the two audits' evidence stays connected"), and four rows break that binding. This is the mirror image of the parent's §B.7 defect (finding A-1 in `SPEC-GOAL-SURFACE-UNIFY-001/plan-audit-3.md`); fix both together so the two registers agree. Required fix: `AC-GDR-005 ← AC-GSU-028 (emission half, a3 re-anchored)`; `AC-GDR-006 ← AC-GSU-028 (retention half)`; `AC-GDR-008 ← AC-GSU-029`; `AC-GDR-009..011 ← new`; `AC-GDR-012 ← new (aggregate emission integration)`.

**B-4. `spec.md:58` states 18 emission markers; AC-GDR-012's measured baseline is 24** — Severity: **major** — Commands: `grep -n 'emission mar' spec.md` → line 58 "**8 are sweep targets** carrying **18 emission markers**"; AC-GDR-012 reproduced → `total=24` (`paired_al 4 + paired_se 8 + auto_mode 4 + l7 4 + handoff 4`). Arithmetic: `4+8+4+4+4 = 24`; the stale `18 = 4+8+2+4` is the **pre-fix** count — it uses the disqualified `per-turn` detector's `2` instead of the corrected `4`, and omits the L7 marker (`4`) entirely. This under-states scope by exactly the two things this SPEC exists to fix. An implementer working to "18 markers" would stop before sweeping the L7 mis-attribution, which is AC-GDR-004's whole subject. Required fix: state `24` in `spec.md` §A.3 and check `plan.md` for the same figure (`grep -nE '18|24' plan.md | grep -iE 'marker|emission'` → no hits, so `spec.md:58` is the only site).

**B-5. AC-GDR-005 uses `grep -c` (line count) without the marking §A.1 requires** — `acceptance.md:84` — Severity: minor — Command: `sed -n '/^\*\*AC-GDR-005\*\*/,/^- Target/p' acceptance.md | grep -ciE 'per-line|line count|presence check'` → `0`. §A.1 states "Occurrence counts. `grep -o … | wc -l`, never bare `grep -c`, except where a per-line presence check is the intent (marked)". AC-GDR-005 (and AC-GDR-004, AC-GDR-006) use `grep -c` without that marking. The choice is defensible — a per-line presence check is exactly the intent for a one-line pairing — but the convention requires it be said, and an unmarked `grep -c` would undercount two markers on one line. Required fix: mark them, or note in §A.1 that the per-locale criteria use presence semantics by design.

**B-6. `tier: M` but the Tier L artifact set is delivered** — Severity: minor — Command: `ls -1 *.md | wc -l` → `6` (`spec`, `plan`, `acceptance`, `design`, `research`, `progress`). `spec-workflow.md` § SPEC Complexity Tier assigns Tier M a 3-file set (spec + plan + acceptance). Delivering `design.md` and `research.md` as well is harmless over-delivery — and given B-1/B-2 concern design reasoning, the extra artifacts earn their place — but it means the tier declaration understates the evidence base. No fix strictly required; note it so a future reader does not read the extra artifacts as a tier misclassification.

**B-7. AC-GDR-004 is line-number anchored** — `acceptance.md:74` — Severity: minor — the detector is `sed -n '7p' … | grep -c '`/goal`'`. The criterion acknowledges the brittleness and states a re-anchor plan ("re-anchored on `provides` + `` `/goal` `` co-occurrence rather than silently passing"), which is the right disclosure. Recorded because the stated fallback is prose-adjacent (`provides` is English) and would itself risk the N2 class; prefer a structural anchor (first non-front-matter paragraph) or the `` `/goal` ``+`` `/moai goal` ``+`` `/moai loop` `` triple co-occurrence, which is locale-invariant.

---

## Clean Dimensions

- **Retention register independence**: clean and complete. `plan.md` §A.2 holds **4** rows — `claude-code/**` (28 pages), `autonomous-loops.md` as a split surface, `cli-reference/goal.md` + `utility-commands/moai-goal.md` (the factual-contrast pair), and `.moai/docs/autonomous-workflow-strategy.md` — and `spec.md:59` states "Four retention surfaces", so the count is internally consistent (unlike the parent's history). Each row is guarded: row 1 by AC-GDR-008's `cc=80` pin, row 2 by AC-GDR-006's per-locale `h3/h2/row` pins, row 3 by AC-GDR-008's `goal.md=12` and `moai-goal=4`, row 4 by AC-GDR-007's `25` content pin. **No sweep AC can satisfy itself by deleting a registered surface**: every one of the four compounds (006, 008, 009, 010) pairs its pin with AC-GDR-012's aggregate rather than with a target of its own, so the pins cannot be traded away for the emission target. REQ-GDR-005 correctly binds by reference and does **not** restate the register — better discipline than the parent's REQ-GSU-004.
- **Per-locale detector symmetry, verified independently**: every emission detector re-run per locale. `paired_al` 1/1/1/1 · `paired_se` 2/2/2/2 · `auto_mode` 1/1/1/1 · `l7` 1/1/1/1 · `handoff` 1/1/1/1. All symmetric as reported. Anchors checked against the §A.1 anchor table: `` `/moai loop` ``, `` `/goal` ``, `--goal` are code literals; `auto mode` is untranslated in all four locales (verified by reading L94 in each — `auto mode` appears verbatim in ja/ko/zh alongside translated surroundings); the structural anchors (`^## `, `^\| `, `^### `) are markdown syntax. **None is anchored on translatable prose**, and the one that was (`per-turn`) is recorded in the table as PROHIBITED and demonstrably disqualified.
- **Split path ownership**: SPEC-B owns zero `.claude/` or `internal/` paths (`grep -cE '^\| \`\.claude/|^\| \`internal/' plan.md` → `0`); the parent owns zero `docs-site` paths as scope. `§F.1` arithmetic verified: `4+4+4+1+0 = 13`, matching §A.3's "13 files" and §D's "13 paths".
- **Held-out gates**: `hugo --quiet` in `docs-site/` → `exit=0` (AC-GDR-011's first half). Four-locale inventory `ls docs-site/content/*/advanced/hooks-reference.md | wc -l` → `4`. Parent-dependency gate `grep -c 'a /moai goal condition' internal/cli/handoff.go` → `0`, correctly marking N3 blocked.
- **Traceability §E matrix**: 8/8 REQs cited, reverse direction complete. Spot-checked for padding — REQ-GDR-007 → AC-GDR-006 is genuine (AC-006's target is literally the unchanged per-locale pins, so violation fails it); REQ-GDR-008 → AC-GDR-008 (`hooks=2`) + AC-GDR-009 (inventory) are two independent guards that would each fail on violation, as the matrix claims.
- **Time estimates**: none.

---

## Recommendation

FAIL at 0.75 against Tier M's 0.80. The distance is small and concentrated: **B-1 and B-2 are both defects in AC-GDR-010's specification, not in its measurement**, and B-3/B-4 are bookkeeping. Fix order:

1. **B-1** — add the liveness/non-zero-floor component to AC-GDR-010. This is the one finding I would call blocking on its own: without it, the criterion billed as making the N2 class "mechanically impossible" is satisfied by a dead regex, and the run phase has no way to tell.
2. **B-2** — add the content-symmetry carve-out so REQ-GDR-003/004 stop contradicting REQ-GDR-008.
3. **B-4** — correct `18` → `24` in `spec.md:58`; the understatement would cause the L7 sweep to be skipped.
4. **B-3** — correct the four provenance rows, jointly with the parent's A-1 so the two registers agree.
5. **B-5 … B-7** — one-line clarifications.

What this SPEC gets right is worth stating plainly, because it is the reason the split was the correct call. The parent's N2 was a single prose-anchored detector that read zero in three locales; SPEC-B does not merely fix that instance — it names the failure class, prohibits the anchor category in a table, requires per-locale baselines as a recording rule (REQ-GDR-004), and adds a meta-guard whose subject is the detectors themselves. All 12 baselines reproduce, every emission detector is symmetric, and the retention register is consistent and fully guarded on the first attempt. The two MUST-FIX are gaps in how completely the meta-guard was specified, not evidence that the approach is wrong — and both are closable with components in the same style the SPEC already uses.
