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

---

## v1.6.0 Amendment Audit (Phase 1 run-gate, 2026-07-28)

**Scope**: v1.6.0 DELTA ONLY — Amendment 2 (B2-2/B2-3 hardening of AC-GDR-010 components b/c). v1.0 content was already audit-PASSed at iteration 2 (0.86) and closed; v1.5.0 D2 amendment was closed at PR #1181 (62df55fab). Reasoning context ignored per M1 Context Isolation. Commands executed in worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/retire001-b22b23` at `HEAD = bf711ec80`, branch `worktree-retire001-b22b23`.

### Verdict: **FAIL**

Tier S threshold = 0.75. Harmonic-mean score = **0.80** (passes threshold) but one MUST-FIX finding (ownership-routing contradiction) is a plan-phase authorization of a forbidden ownership crossing that the run-phase cannot self-heal without a blocker cycle. FAIL on structural grounds pending the one-line-per-site fix.

### Must-Pass Results (delta scope)

| | Result | Basis |
|---|---|---|
| MP-1 REQ consistency | N/A | No REQs added in amendment; v1.0 REQ-GDR-001..011 unchanged. |
| MP-2 GEARS compliance | N/A | No new ACs in amendment; AC-GDR-010 body unchanged at plan-phase (refactor deferred to run-phase). |
| MP-3 frontmatter validity | **PASS** | All 12 canonical fields present (`id`, `title`, `version: "1.6.0"` quoted, `status: in-progress`, `created`, `updated`, `author`, `priority: MEDIUM`, `phase`, `module`, `lifecycle: spec-anchored`, `tags`). `tier: S`, `amendment_of`, `run_commit_sha`, `sync_commit_sha` optional fields consistent. No snake_case aliases. |
| MP-4 language neutrality | N/A | Docs-site i18n SPEC; single-surface (4 locales of 3 docs pages). No tool/language primacy claim. |
| MP-5 D7 cross-SPEC | **PASS** | Only cross-SPEC reference is `SPEC-GOAL-SURFACE-UNIFY-001` (unchanged from v1.0); no new references introduced by the amendment. |
| MP-6 D8 cross-platform | **PASS** | `grep -c syscall` over the v1.6.0 delta (HISTORY row + Amendments §2 + plan.md §I) → 0. |
| MP-7 clarification gate | **PASS** | `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → no markers (only match was this file's own prior-iteration report text describing the check). |

### Category Scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|-----------|-------|------|----------|
| Clarity | 0.75 | 0.75 — minor ambiguity | plan.md §I.1–§I.10 mechanisms are clearly explained (chosen option + rejected alternatives + rationale) for both B2-2 and B2-3. However, the ownership-routing statement ("run-phase scope (manager-develop)") at plan.md:256 and spec.md:62 directly contradicts the v1.5.0 routing within the same plan.md file (plan.md:182 → `manager-spec`) and the canonical matrix — a reader cannot determine the correct owner without consulting the SSOT. |
| Completeness | 0.85 | 0.75 — one non-critical gap | HISTORY v1.6.0 row, Amendments §2 (with prior_completed_sha `62df55fab` verified), plan.md §I.1–§I.10 (Context, B2-2/B2-3 mechanisms, Baseline invariant, Integration boundary, Pre-flight, Constraints, Self-Verification, Residual-risk, Milestone M1') all present. Tier S artifact set is 2 files (spec + plan); acceptance.md correctly deferred to run-phase. |
| Testability | 0.80 | 0.75 — minor interpretation | AC-GDR-010 components are binary-testable. The new B2-2/B2-3 checks are described at mechanism level with implementation shape deferred to run-phase ("run-phase agent finalizes the exact shell"). The recorded baseline is asserted verbatim and I verified it holds today. The deferral is acceptable for Tier S plan-phase, but the exact shell form will only be testable at run-phase. |
| Traceability | 0.80 | 0.75 — indirect mapping | Amendment §2 of spec.md traces to plan.md §I; the B2-2/B2-3 findings trace to progress.md B2-2/B2-3 (deferred from iteration 2). The amendment hardens AC-GDR-010's interpretation of REQ-GDR-009 (component c) and REQ-GDR-010 (component b) without adding new REQs — the mapping is indirect but traceable. |

**Harmonic mean** = 4 / (1/0.75 + 1/0.85 + 1/0.80 + 1/0.80) = 4 / (1.333 + 1.176 + 1.250 + 1.250) = 4 / 5.009 = **0.799** (rounds to 0.80).

### Defects Found

**D1. Ownership-routing contradiction (MUST-FIX).**
`plan.md:256` (§I header) AND `spec.md:62` (Amendments §2, "Plan-phase artifacts" bullet) — Severity: **critical** — both sites state: *"The `acceptance.md` body refactor is run-phase scope (manager-develop)"* / *"a separate run-phase delegation (manager-develop)"*. The Status Transition Ownership Matrix (`.claude/rules/moai/development/spec-frontmatter-schema.md` § Forbidden ownership crossings) explicitly forbids this: *"manager-develop MUST NOT modify spec.md / plan.md / acceptance.md body content … When run-phase reveals a need to modify SPEC body content, manager-develop MUST return a blocker report and the orchestrator re-delegates to manager-spec for the scope-doc update before re-delegating back."* The `completed → in-progress (amendment)` matrix row assigns amendment work to **manager-spec** via the D-NEW-1 inline-fix pattern. The contradiction is internally demonstrable: the SAME plan.md file at `plan.md:182` (v1.5.0 §F.2 milestone M1, for the AC-GDR-012 refactor — the structurally identical case) correctly says *"Run-phase owner: `manager-spec` (per the D-NEW-1 inline-fix pattern … the `completed → in-progress (amendment)` row of the Status Transition Ownership Matrix names `manager-spec` as the owner via re-delegation). When run-phase reveals a need to modify the AC body, `manager-develop` MUST return a blocker report and the orchestrator re-delegates to `manager-spec`."* The v1.6.0 §I sites forgot the v1.5.0 §F.2 routing that sits 74 lines above them in the same file. **Required fix:** change both sites to route to `manager-spec (re-delegation per D-NEW-1 inline-fix pattern)`, mirroring `plan.md:182` verbatim. Concretely:
- `plan.md:256` — replace `run-phase scope (manager-develop)` with `run-phase scope (manager-spec, re-delegation per D-NEW-1 inline-fix pattern — same routing as §F.2 v1.5.0 M1)`.
- `spec.md:62` — replace `a separate run-phase delegation (manager-develop)` with `a separate run-phase delegation (manager-spec, per the D-NEW-1 inline-fix pattern)`.
- `plan.md §I.10` M1' deliverable should additionally state the run-phase owner explicitly (mirroring §F.2's *"Run-phase owner: `manager-spec`"* sentence), so the routing is unambiguous from the milestone heading alone.

**D2. B2-3 mitigation reasoning gap (SHOULD-FIX).**
`plan.md:372` (§I.9 Residual-risk, "exemption-list structured-field format drift" bullet) — Severity: **minor** — the residual-risk note correctly identifies the rename hazard (`exempt_detectors:` → `exemptions:`) and asserts the "exactly once" predicate catches it. That reasoning IS correct for the pure-rename case (count goes 1 → 0, predicate trips). However, it does NOT cover the "both fields coexist" case: a future contributor could keep the empty `exempt_detectors: []` (count = 1, passes "exactly once") AND add a parallel `exemptions:` field carrying actual exempted detectors — the new field would silently bypass component (a) symmetry while the old empty field keeps the check passing. **Required fix:** extend the residual-risk mitigation to note that the run-phase implementation SHOULD assert the `exempt_detectors:` declaration is the ONLY exemption-list-shaped field in component (c)'s block (e.g., assert absence of `exemptions:`, `exempt:`, `excluded_detectors:` sibling fields by name), or document this as accepted forward-looking debt. Tier S defense-in-depth is acceptable; the gap is the mitigation reasoning, not the mechanism itself.

### Baseline Verification (verbatim command output)

The following commands were executed against the current worktree (`HEAD = bf711ec80`) to verify the §I.4 baseline-preservation invariant holds pre-refactor. All recorded baselines reproduce.

**§I.6 step 2 — 12 scoped pages exist at immutable base `e306e21a9` with non-zero heading content:**
```
en/advanced/autonomous-loops.md: headings=9
en/advanced/self-evolving.md:    headings=11
en/cli-reference/handoff.md:    headings=5
ja/advanced/autonomous-loops.md: headings=9
ja/advanced/self-evolving.md:    headings=11
ja/cli-reference/handoff.md:    headings=5
ko/advanced/autonomous-loops.md: headings=9
ko/advanced/self-evolving.md:    headings=11
ko/cli-reference/handoff.md:    headings=5
zh/advanced/autonomous-loops.md: headings=9
zh/advanced/self-evolving.md:    headings=11
zh/cli-reference/handoff.md:    headings=5
```
All 12 pages non-zero.

**§I.6 step 3 — heading-set equality base `e306e21a9` ↔ working tree:**
```
ALL 12 PAGES: heading sets byte-identical between e306e21a9 and working tree
```
B2-2 mechanism's pre-refactor precondition holds: the frozen base is structurally current.

**§I.6 step 4 — exemption list prose invariant:**
```
$ grep -c 'exemption list is empty' .moai/specs/SPEC-GOAL-DOCS-RETIRE-001/acceptance.md
1
```
B2-3 mechanism's pre-refactor precondition holds: the list IS empty today (prose form, single occurrence).

**AC-GDR-010 recorded baseline (working tree, pre-refactor):**
```
paired_al:distinct=1,live_min=1,apt=1 auto_mode:distinct=1,live_min=1,apt=1 l7:distinct=1,live_min=1,apt=1 paired_se:distinct=1,live_min=2,apt=1 handoff:distinct=1,live_min=1,apt=1
```
Byte-identical to the recorded baseline at acceptance.md:214 and to plan.md §I.4 line 302.

**AC-GDR-012 (Amendment 1 D2 block, untouched) — aggregate:**
```
tree_total=0   (post-sweep, working tree — recorded target)
base_total=24  (pre-sweep, against immutable base e306e21a9 — recorded baseline)
```
Both reproduce verbatim. The v1.5.0 D2 refactor is preserved; the integration boundary with Amendment 1 is sound.

### EVALUATE verdicts (the 5 questions posed)

1. **OWNERSHIP ROUTING — MUST-FIX defect (D1 above).** The plan's manager-develop assignment is a forbidden ownership crossing per the canonical matrix. The correct agent routing for the `acceptance.md` AC-GDR-010 (b)/(c) refactor is **`manager-spec`** (re-delegation per the D-NEW-1 inline-fix pattern), identical to the v1.5.0 §F.2 M1 routing for the structurally identical AC-GDR-012 refactor. The plan.md §F.2 routing at line 182 is the authoritative in-file precedent; the v1.6.0 §I sites at line 256 (and spec.md:62) contradict it and must be corrected.

2. **B2-2 MECHANISM — SOUND.** The frozen-base structural-currency assertion via `^#` heading-set equality correctly catches the stale-true hazard. The chosen granularity (heading lines, sorted, byte-compared via `diff -q`/`cmp -s`) is the right discriminator because the sweep edits BODY content (emission references inline), never heading structure — so any heading divergence is necessarily a restructuring event. The rejected alternatives are correctly characterized: (b) `git cat-file -e` catches only path-existence (too weak — a renamed-but-still-extant base path passes); (c) content-similarity floor requires choosing N and admits gradual drift below threshold. The "legitimate future section-add trips the check = desired behavior" reasoning is sound — it forces the contributor to either re-anchor the frozen base or document the restructuring, which is exactly the discipline a frozen-base discipline should enforce.

3. **B2-3 MECHANISM — SOUND with one residual gap (D2 above).** The structured `exempt_detectors: []` declaration replacing prose, plus the (1) exactly-once (2) zero-entries-today (3) name+asymmetry+justification-on-addition predicate set, correctly enforces the empty-list invariant + justification-on-addition. The "exactly once" predicate IS robust to a pure rename (`exempt_detectors` → `exemptions`): count drops 1 → 0 and trips the predicate. The residual gap is the "both fields coexist" case (see D2) — mitigation reasoning is slightly incomplete, not the mechanism itself.

4. **BASELINE PRESERVATION — VERIFIED.** The integration boundary is sound. I ran §I.6 pre-flight commands 2/3/4 against the current worktree plus the AC-GDR-010 recorded baseline and the AC-GDR-012 aggregate (base + tree); every recorded value reproduces verbatim (see Baseline Verification above). The v1.5.0 D2 refactor (AC-GDR-012 single-`p=`-source + liveness + aptness guards) is preserved byte-identical; AC-GDR-010 components (a) and (d) are untouched at plan-phase; the 12 sweep-target locale files are untouched; the four retention surfaces are untouched.

5. **TIER S APPROPRIATENESS — CORRECT.** Run-phase scope is 1 file (`acceptance.md`), ~10-15 lines of shell+prose within the AC-GDR-010 component (b)/(c) block, non-constitutional. Tier S thresholds: <5 files affected, <300 LOC, non-constitutional. The v1.5.0 amendment already transitioned the SPEC from Tier M → Tier S for the same structural reason (single-AC-body refactor); v1.6.0 retains Tier S correctly.

### Recommendation

**FAIL → fix D1 (one-line change at two sites) + D2 (extend §I.9 mitigation by one sentence) → re-audit.** The re-audit scope is the enumerated delta only:

1. `plan.md:256` — replace `(manager-develop)` with `(manager-spec, re-delegation per D-NEW-1 inline-fix pattern — same routing as §F.2 v1.5.0 M1)`.
2. `spec.md:62` — replace `(manager-develop)` with `(manager-spec, per the D-NEW-1 inline-fix pattern)`.
3. `plan.md §I.10` M1' — add an explicit "Run-phase owner: `manager-spec`" sentence mirroring §F.2 line 182.
4. `plan.md §I.9` — extend the "exemption-list structured-field format drift" bullet to note the "both fields coexist" gap and the recommended mitigation (assert `exempt_detectors:` is the only exemption-list-shaped field in component (c)'s block).

What this amendment gets right is worth stating plainly. The v1.5.0 D2 work closed the aggregate-pattern-weakening hazard (single-`p=`-source discipline extended to AC-GDR-012); the v1.6.0 B2-2/B2-3 work closes two defense-in-depth gaps the original iteration-2 audit flagged as SHOULD-FIX and the v1.5.0 close did not absorb. Both B2-2 (frozen-base structural currency) and B2-3 (mechanical exemption-list enforcement) are the correct mechanisms at the correct granularity, the recorded baseline is preserved verbatim and reproduces today, and the integration boundary with Amendment 1 is clean. The single MUST-FIX is a routing error, not a mechanism defect — fixable in two one-line edits, after which the re-audit is expected to PASS at the same harmonic mean (~0.85 once the clarity score regains the 1.0 band).

---

## v1.6.0 Amendment Re-Audit (Phase 1 run-gate, 2026-07-28)

**Scope**: v1.6.0 DELTA ONLY — verification that manager-spec's commit `8c9e1173a` resolved D1 (MUST-FIX) and D2 (SHOULD-FIX) from the prior iteration above. v1.0 content remains audit-PASSed at iteration 2 (0.86) and closed; v1.5.0 D2 amendment closed at PR #1181 (62df55fab). Reasoning context ignored per M1 Context Isolation. Read-only verification against the worktree at `.claude/worktrees/retire001-b22b23`.

### Verdict: **PASS**

Tier S threshold = 0.75. Harmonic-mean score = **0.86** (passes threshold). Both D1 (MUST-FIX) and D2 (SHOULD-FIX) from the prior iteration are RESOLVED; no new defects introduced.

### Regression Check — prior-iteration defects

- **D1 (Ownership-routing contradiction, MUST-FIX, critical): RESOLVED.** Verified at all three required sites:
  - `plan.md:256` (§I header) — verbatim: *"The `acceptance.md` body refactor is run-phase scope (manager-spec, re-delegation per D-NEW-1 inline-fix pattern — same routing as §F.2 v1.5.0 M1; `.claude/rules/moai/development/spec-frontmatter-schema.md` § Forbidden ownership crossings forbids `manager-develop` from modifying `acceptance.md` body content, and the `completed → in-progress (amendment)` row of the Status Transition Ownership Matrix names `manager-spec` as the owner via re-delegation)."* — now routes to manager-spec with explicit D-NEW-1 citation and §F.2 cross-reference.
  - `spec.md:62` (Amendments § 2, "Plan-phase artifacts" bullet) — verbatim: *"The `acceptance.md` body refactor is a separate run-phase delegation (manager-spec, per the D-NEW-1 inline-fix pattern)."* — `(manager-develop)` replaced with `(manager-spec, per the D-NEW-1 inline-fix pattern)`.
  - `plan.md:376` (§I.10 M1' deliverable) — verbatim: *"**Run-phase owner: `manager-spec`** (D-NEW-1 re-delegation per the Status Transition Ownership Matrix `completed → in-progress (amendment)` row; manager-develop is forbidden from `acceptance.md` body edits per § Forbidden ownership crossings and would return a blocker report if delegated)."* — mirrors the §F.2 line 182 sentence structure verbatim.

  **Consistency with §F.2 precedent (plan.md:182):** confirmed. Both sites use the same "**Run-phase owner: `manager-spec`** (D-NEW-1 re-delegation per …)" form, cite the same canonical matrix row, and cite the same § Forbidden ownership crossings forbiddance. A reader at any of the three v1.6.0 sites now reaches the same owner determination without consulting the SSOT.

  **Residual bare `manager-develop` mentions in §I:** 3 total, all CORRECT (not mis-routing): (a) L256 forbidden-crossing citation, (b) L354 §I.7 progress.md §E.2/§E.3 ownership statement (canonical matrix assigns progress.md §E.2/§E.3 to manager-develop — this is NOT acceptance.md body), (c) L376 forbidden-crossing citation. spec.md Amendments § 2 carries zero residual `manager-develop` mentions. No forbidden ownership crossing remains.

- **D2 (B2-3 coexist-guard gap, SHOULD-FIX, minor): RESOLVED.** `plan.md:372` (§I.9 Residual-risk, "Exemption-list structured-field format drift" bullet) — verbatim added sentence: *"Additionally, the run-phase check SHOULD assert the `exempt_detectors:` declaration is the ONLY exemption-list-shaped field in AC-GDR-010 component (c)'s block — e.g., assert by name the absence of sibling fields like `exemptions:`, `exempt:`, `excluded_detectors:` — covering the bypass where an empty `exempt_detectors: []` passes the 'exactly once' predicate while a parallel-shaped field silently carries the actual exemption entries."* — directly addresses the "both fields coexist" gap identified in the prior iteration.

### Category Scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|-----------|-------|------|----------|
| Clarity | **1.00** | 1.0 — single, unambiguous interpretation | D1 fix restored routing consistency across all 3 sites (plan.md:256, plan.md:376, spec.md:62) with the §F.2 precedent at plan.md:182. A reader at any site reaches manager-spec unambiguously. The 3 residual `manager-develop` mentions are correct forbidden-crossing citations + progress.md §E.2/§E.3 ownership, not contradictions. |
| Completeness | **0.85** | 0.75 — one non-critical gap (unchanged) | HISTORY v1.6.0 row, Amendments § 2 (prior_completed_sha `62df55fab`), plan.md §I.1–§I.10 all present. Tier S artifact set is 2 files (spec + plan); acceptance.md correctly deferred to run-phase. The B2-2/B2-3 mechanism descriptions + pre-flight + constraints + self-verification + residual-risk + milestone are complete; the only residual interpretation surface is "run-phase agent finalizes the exact shell" for B2-2/B2-3 (acceptable for Tier S plan-phase). |
| Testability | **0.80** | 0.75 — minor interpretation (unchanged) | AC-GDR-010 components remain binary-testable. The B2-2 heading-set equality and B2-3 count check are described at mechanism level with implementation shape deferred to run-phase. D2 fix adds a SHOULD-level sibling-field absence assertion (testable via `grep -E '^(exemptions|exempt|excluded_detectors):'`). |
| Traceability | **0.80** | 0.75 — indirect mapping (unchanged) | Amendment § 2 of spec.md traces to plan.md §I; B2-2/B2-3 findings trace to progress.md B2-2/B2-3 (deferred from iteration 2). The amendment hardens AC-GDR-010's interpretation of REQ-GDR-009 (component c) and REQ-GDR-010 (component b) without adding new REQs — the mapping is indirect but traceable. |

**Harmonic mean** = 4 / (1/1.00 + 1/0.85 + 1/0.80 + 1/0.80) = 4 / (1.000 + 1.176 + 1.250 + 1.250) = 4 / 4.676 = **0.855** (rounds to 0.86).

### No-New-Defects Check

- **B2-2 mechanism (heading-set equality)**: unchanged from prior iteration — remains SOUND. The frozen-base structural-currency assertion via `^#` heading-set equality correctly catches the stale-true hazard; the rejected alternatives (`git cat-file -e`, content-similarity floor) are correctly characterized. The "legitimate future section-add trips the check = desired behavior" reasoning is sound.
- **B2-3 mechanism (structured exemption-list field + count check)**: unchanged; D2 fix extended only the §I.9 residual-risk mitigation, not the mechanism itself. The structured `exempt_detectors: []` declaration + (1) exactly-once (2) zero-entries-today (3) name+asymmetry+justification-on-addition predicate set remains SOUND. The D2 fix additionally closes the "both fields coexist" bypass at SHOULD level.
- **Baseline preservation**: §I.4 invariant intact — recorded values (`distinct=1,live_min>=1,apt=1` for all 5 detectors; `paired_se:live_min=2`; `total=24` base / `total=0` tree) are preserved verbatim in the plan text.
- **Tier S appropriateness**: CORRECT — 1 file (acceptance.md), ~10-15 lines of shell+prose within AC-GDR-010 (b)/(c), non-constitutional. v1.5.0 already transitioned the SPEC from Tier M → Tier S for the same single-AC-body refactor shape; v1.6.0 retains Tier S correctly.

### Recommendation

PASS. Both defects from the prior iteration (D1 MUST-FIX + D2 SHOULD-FIX) are resolved at the cited sites with verbatim evidence. The Clarity dimension regains the 1.0 band as expected, lifting the harmonic mean from 0.80 → 0.86. The amendment is authorized to proceed to run-phase (manager-spec, per the D-NEW-1 inline-fix pattern) for the AC-GDR-010 component (b)/(c) refactor. The Implementation Kickoff Approval human gate (plan→run) remains mandatory and score-independent — this PASS verdict does not auto-bypass it.
