---
id: SPEC-TODO-LANDING-ATTRIBUTION-001
title: "The landed verdict: an attribution-position predicate, and a ref chain that asks the branch this repository actually integrates on"
version: "0.4.0"
status: completed
created: 2026-09-03
updated: 2026-09-06
author: manager-spec (card t472)
priority: P1
phase: "v3.2.0 target"
module: "internal/kanban, internal/cli"
lifecycle: spec-anchored
tags: "kanban, backlog-queue, landing-state, attribution, git-flow, integration-branch, false-positive"
tier: M
related_specs:
  - SPEC-TODO-LANDING-STATE-001
  - SPEC-KANBAN-QUEUE-PR-SYNC-001
  - SPEC-WORKTREE-BASEREF-001
---

# SPEC: The landed verdict — attribution position, then the right ref

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.4.0 | 2026-09-04 | Plan-audit iteration-3 remediation, **D3-1 only** — an operator decision taken after iteration 3 returned PASS-WITH-DEBT 0.8375 (down 0.0375, firing `spec-workflow.md:160`'s STOP condition alongside the three-iteration ceiling). **There will be no fourth audit**; this change is verified by the two commands quoted in §A.4.2 rather than by another round. Measured in this tree at HEAD `012d9680a` against the pinned corpus `7835148d3`, with a binary built from this tree. **Form 2b added** — form 2's own position with a pull-request reference group appended, which form 2's `)$` anchor silently excluded: 43 subjects, 40 distinct ids, **39 attributed by no other form**. Attributed set **270 → 309**; subject-present-but-unattributed **77 → 38**; the named-shape under-count is restated as a **floor of at least 10, explicitly not a total** — the residual has never been exhaustively classified by author, auditor, or lane. `plan.md` §D's tolerance ground ("all seven are archived-era, impact nil") is **withdrawn as false** and re-argued on the failure direction alone. AC-TLA-002 gains clause 3 (fixture `t210`, form 2b's uniquely-attributed set = 39); the AC-TLA-005 map gains form 2b's row and form 1's column is corrected 42 → 41 (`t230` is now shared). Counts unchanged: **13 REQ / 13 AC** (Tier M ceiling 16/16). D3-2 through D3-5 are recorded as accepted debt in `progress.md` and are deliberately NOT fixed here. |
| 0.3.0 | 2026-09-04 | Plan-audit iteration-2 remediation, measured in this tree at HEAD `75e63d6f2` against the **pinned corpus commit** `7835148d3` (5,837 subjects, 414 merges). iter-2 D1: form 3b gains a single-token restriction and a falsifying criterion built on `t412` (AC-TLA-003 clauses 4-5). iter-2 D2: the recorded under-count "exactly one card" is **withdrawn** — measured `>= 19` under four named missed shapes; **form 3c** (card-led local merge, `merge: <card>` — 31 subjects, 100% precision) is added, taking the residual to **7**, and `plan.md` §D's tolerance is re-argued at 7 rather than at 1. iter-2 D3: form 3b's target is stated as **derived from the resolved landed ref**, with AC-TLA-003 clause 6 falsifying a hardcoded `develop`; the audit's M1-window claim is **refuted by measurement** (§A.7). iter-2 D4: the non-attribution rule is restated **contraposed** on target mismatch; the `WT-` prefix becomes an illustration. iter-2 D5 handled in `progress.md`. Counts: **13 REQ / 12 AC** (REQ-TLA-013 added for the derived target; covered by AC-TLA-003 clause 6 — Tier M ceiling 16/16, so both stay in budget). |
| 0.2.0 | 2026-09-03 | Plan-audit iteration-1 remediation, measured in this tree at HEAD `e227871b4` against `origin/develop` `7835148d3` (5,837 subjects). D1 (BLOCKING) closed: §A.4 form 3 was an occurrence test and is replaced by two positional shapes (3a, 3b) plus an explicit non-attribution rule for absorb-direction merges; `MUT-MERGE-ANY-TOKEN` added to the mutant set. D2 citation corrected and its residual re-stated as measured-disjoint. D4/D5/D6/D7/D8/D9 dispositions recorded in `acceptance.md`, `plan.md`, and `progress.md`. Requirement and criterion counts unchanged at 12/12. |
| 0.1.0 | 2026-09-03 | Initial plan-phase authoring (card t472), measured in worktree `.claude/worktrees/t472` at HEAD `4bcac7079` (branch `WT-landed-drift-detect`). Two axes, merged by lead verdict into one SPEC: the attribution predicate (axis F) and the ref chain plus its disclosure (axes A+B). Every figure below is carried from this lane's own committed measurements at `.moai/reports/t472/premise-recheck.md` and `.moai/reports/t472/axis-bf-measurement.md`, or re-run in this tree and cited beside the command. |

> **Provenance discipline.** Every `file:line` citation is measured at tree `4bcac7079`. Every figure
> **about** a moving ref (`origin/main`, `origin/develop`, `refs/remotes/origin/HEAD`) travels with the
> command that produced it and is re-measured rather than re-cited by a later reader
> (`verification-claim-integrity.md` §2.1, remedy R4).
>
> **The corpus is pinned to a commit, not to a branch name (VCI §2.1 remedy R1).** Every count below
> is measured over the commit `7835148d3` and is cited as such. That pin is not decoration: at
> version 0.2.0 `origin/develop` resolved to `7835148d3`; re-read at version 0.3.0
> (`git rev-parse origin/develop`, 2026-09-04) it resolves to `25a3212a9`, and `7835148d3` remains
> an ancestor (`git merge-base --is-ancestor 7835148d3 25a3212a9`, rc=0). A reader re-running these
> commands against the branch name will measure a different corpus and should; a reader checking
> **these** figures runs them against `7835148d3`.

---

## §A Context

### A.1 What is wrong, in one sentence per axis

**Axis F — the predicate answers a question nobody asked.** `LandedGrepArgs`
(`internal/kanban/prlink_landed.go:96-109`, tree `4bcac7079`) builds

    git log <ref> --perl-regexp --grep=\btNNN\b --oneline

and `--grep` matches the **whole commit message**. A commit that merely *mentions* a card in its
body therefore reads as that card having landed. The file's own comment already knows the hazard —
"a card's first matching commit may be another card's report commit that merely mentions it"
(`prlink_landed.go:139-140`) — but that knowledge was spent suppressing SHA output, not on the
match itself.

**Axes A+B — the verdict asks the wrong branch, and never says which.** `LandedRefFor`
(`prlink_landed.go:74-80`) reads `git_strategy.worktree_base_branch` from the root the queue hangs
from, which `todoLandedRef` (`internal/cli/todo.go:81-90` — doc comment `:81-87`, func `:88-90`) deliberately resolves to the **primary
checkout** — "the queue and the integration branch are properties of one repository, not of
whichever worktree the command happens to run in". That primary checkout is parked on `main`, whose
copy of the key is empty, so resolution falls to the hardcoded `DefaultLandedRef = "origin/main"`
(`prlink_landed.go:41`). The fallback is taken with no run-time notice, and the verdict line
`done <id> landing=<verdict>` (`internal/cli/todo.go:508`) does not name the ref that answered.

### A.2 The measured size of axis F

Measured in this tree with a binary built from it (`go build ./cmd/moai`), not the installed build
— VCI §2.2.

| Population | Landed verdicts | True positives | False positives |
|---|---|---|---|
| Live queue against `origin/main` (today's ref), 57 rows | 2 (t237, t312) | 0 | **2** |
| Live cards against `origin/develop` (the ref axis A would select), 31 cards | 9 | 2 (t401, t440) | **7** |

Precision of the shipped predicate on the population it currently marks landed: **0/2**.

Reproduction of one false positive, re-run in this tree at `4bcac7079`:

    git log origin/main --perl-regexp --grep='\bt237\b' --oneline
      539349c5b docs(t230): t230 sync-audit evidence ... (#1649)
      32d2221fa feat(cli): back up and disclose user-modified pre-commit hooks (t230) (#1647)

Both commits are **t230's**. `32d2221fa`'s body reads "the card about to change the hook body is
t237/#1641. Re-measured: the issue is OPEN" — a statement that t237 has *not* landed, read as
evidence that it has. Commits whose subject attributes t237: zero.

Control (the comparison is not vacuous), re-measured at `origin/develop` `7835148d3` under the
**repaired** §A.4 enumeration (forms 1, 2, 3a, 3b):

    # attributed ids — forms 1/2/3a
    git log origin/develop --format=%s | grep -ohE '^[a-z]+\(t[0-9]+\)!?:|\((card )?t[0-9]+\)$|^Merge card t[0-9]+' \
      | grep -oE 't[0-9]+' | sort -u | wc -l           → 257
    # form 3b adds exactly one id beyond those three   → t412
    # ids mentioned anywhere in a message
    git log origin/develop --format=%B | grep -oE '\bt[0-9]+\b' | sort -u | wc -l   → 379

Attributed = **258**, mentioned = **379**, and `attributed ⊆ mentioned` holds (`comm -23` → 0 lines).
Both operands are non-empty, so the comparison asserts something.

*Correction (plan-audit D9a).* Version 0.1.0 recorded this control as **291**. That figure was
produced by a broader proxy (paren convention + bare token) than §A.4 defines, and does not
reproduce under the SPEC's own enumeration. The figure above is the reproducible one. The control's
purpose — both operands non-empty — held under either.

### A.3 Why matching the subject is not enough

Two of the seven `develop` false positives survive a "match the subject only" repair, because they
are **other cards' subjects mentioning this card**. Re-run in this tree:

    git log origin/develop --perl-regexp --grep='\bt216\b' --oneline    # 3 lines, shown in full
      48c35a4d4 Merge branch 'WT-incremental-rebuild' into develop (card t263)
      673d3d8a0 docs(t263): die-at-exit reproduced (0/5) — remedy sequenced behind t216
      2f170549b fix(hooks): re-wire the navigator SessionStart hook (t243)

    git log origin/develop --perl-regexp --grep='\bt443\b' --format='%h %s'   # 14 lines; 1 shown
      0d26f8a00 chore(catalog): revert sync-auditor hash to develop value — t443 jurisdiction (t461)
      [13 further lines elided — body mentions and other cards' subjects; selection rule: the single
       line whose *subject* contains the queried token is shown, the rest are body-only matches]

None of the three `t216` lines attributes t216: `48c35a4d4` and `673d3d8a0` are **t263**'s and
`2f170549b` is **t243**'s (its body mentions t216). `0d26f8a00` is attributed to **t461**. In the two
subject-occurrence cases (`673d3d8a0`, `0d26f8a00`) the queried card appears in the subject and is
not the card the commit belongs to. The discriminator is therefore not *occurrence*, nor
*occurrence in the subject*, but **attribution position**.

### A.4 The attributing positions — six positional shapes

Every shape below is a **position**, not an occurrence. A card token that appears anywhere else in a
subject attributes nothing. Three non-attributing positions are named explicitly, because each is a
place a loosened discriminator would immediately misread:

- **mid-sentence** — `docs(t263): … remedy sequenced behind t216` attributes t263, not t216;
- **inside a branch name** — `merge: t79 — glm_task delegation family (branch WT-t80)` attributes
  t79, not t80, and `Merge branch 'WT-t250-followup' into develop — card t279 …` attributes t279,
  not t250. A branch name lives inside the subject, so an occurrence test reads it as an
  attribution; it is a **name for a place work happened**, never a claim about which card landed;
- **inside a dependency or absorb note** — `(t387 depends on t386 convention doc)`,
  `(card t36, absorbs t2)`.

Counts measured over the pinned commit `7835148d3`, 5,837 subjects, 414 of them merges.

| # | Form | Positional rule | Observed | Example |
|---|---|---|---|---|
| 1 | Conventional-commit scope | `<type>(<card>):` **at subject start** | 290 | `docs(t440): record develop-absorb re-measure evidence` |
| 2 | Trailing parenthetical | `(<card>)` or `(card <card>)` **closing the subject**, the group carrying nothing else | 869 | `... refresh catalog moai whole-tree hash (t447)` |
| 2b | Trailing parenthetical **before a reference group** | a card-bearing parenthetical group carrying **exactly one** card token, followed by a **reference group** `(#NNNN)` that closes the subject — form 2's own position with a pull-request reference appended | 43 | `feat(kanban): moai todo pr — read-only card-to-PR and landed link (t210) (#1628)` |
| 3a | Merge, card-led (`Merge card`) | subject **begins** `Merge card <card>` | 5 | `Merge card t440 (WT-delivery-notice-docs) into develop: ...` |
| 3b | Merge, integration-targeted | the merge's **named target is the branch the resolved landed ref names** (§A.4.1), AND the subject's trailing parenthetical group carries **exactly one** card token | 77 | `Merge branch 'WT-mx-tag-edges' into develop (card t412 — SPEC-MX-TAG-EDGES-001)` |
| 3c | Merge, card-led (`merge:`) | subject **begins** `merge: <card>` — the local-merge spelling of 3a | 31 | `merge: t106 — todo queue resolves to the primary checkout from worktrees — review-PASS` |

Forms 3a and 3c are one shape in two spellings: the card is the **first token after the merge
verb**. They are listed apart only because the two spellings anchor differently.

Form 2b is likewise form 2 in a second spelling, and is listed apart for the same reason: form 2
anchors the card-bearing group at **end of subject**, and that anchor is what form 2b's subjects
fail. Form 2b relaxes the anchor by **exactly one** trailing group, and that group must be a
reference group `(#NNNN)` carrying no card token — it does not relax "the group carrying nothing
else" into "anything may follow". The single-card-token restriction is form 3b's, applied here for
the same reason: without it the group `(branch WT-t80)` would attribute a branch name. Measured over
the pinned corpus, the restriction costs nothing on this form's own subject set — **0** of the 43
groups names two distinct cards:

    <pinned-corpus subjects> | grep -oE '\([^()]*t[0-9]+[^()]*\) \(#[0-9]+\)$' \
      | grep -cE '\bt[0-9]+\b.*\bt[0-9]+\b'                          → 0

**[HARD] The non-attribution rule, stated as a target mismatch (plan-audit iter-2 D4).** A merge
subject that **names a target** attributes no card **unless that target is the branch the resolved
landed ref names**. A merge into anything else — a card worktree branch, an agent worktree, a
release branch — absorbs work rather than landing it: the subject of the sentence is a branch, and
its parenthetical is a dependency note or an absorb record.

Version 0.2.0 wrote this rule as a `WT-…` **prefix** test. That keyed it on a branch-naming
convention this SPEC does not own (it belongs to `kanban-dispatch.md` § Isolation is entered, never
provisioned) and which **this repository's own history violates**. Merge targets over the pinned
corpus that are card or agent worktree branches yet carry no `WT-` prefix — extracted from the
subject stream with `grep -E '^[Mm]erge' | grep -ohE ' into [A-Za-z0-9/_.-]+' | sed -E 's/ into //' | sort -u`:

    worktree-t166, worktree-t176,
    worktree-agent-{a205e7a01ec2e0f27, a350b7a40faaf39c6, a66650c94c57df3f8,
                    a6a00e00dfdb5f0ee, ab81b7087bc743f02, ad55b5fbe632611a7},
    t403, t78, t86

Nine targets, and **three carry no prefix at all** (`t403`, `t78`, `t86`) — no prefix rule of any
spelling reaches those. The contraposed form is strictly stronger, costs nothing, and is form 3b's
own target test read the other way; `WT-…` survives only as an illustration.

No false attribution results from the prefix form **today** — none of the nine subjects carries a
card-bearing trailing group — so this was soundness rather than a live defect. It is repaired
because the shipped surface is `moai todo`, which reaches repositories whose branch names MoAI does
not prescribe. Measured exclusions under the repaired rule, unchanged for the `WT-`-shaped majority:

    <pinned-corpus subjects> | grep -E '^Merge ' | grep -E 'into WT-' \
      | grep -oE '\([^()]*\)$' | grep -cE 't[0-9]+'        → 21

#### A.4.1 Form 3b's target is derived, never spelled

[HARD] Form 3b's target is **the branch the resolved landed ref names** — derived at evaluation time
from the ref the resolver returns (the §B.2 chain once M2 lands; the un-repaired resolver before
that). It is **not** the literal string `develop`. An implementation that spells `develop` into the
form is wrong in every repository whose integration branch is not `develop`, permanently and
silently: it invents attributions from `into develop` merges the repository never made, and misses
every merge into the branch it actually integrates on. **AC-TLA-003 clause 6 is the falsifier**, and
§A.7 records what the form yields during the M1-only window.

**Why 3b, and not "a merge subject naming the card" (plan-audit D1).** Version 0.1.0 wrote form 3 as
*"a merge subject naming the card"* — an **occurrence** test, and therefore the exact reading §A.3
rejects, reproduced inside this SPEC's own enumeration. Measured, the occurrence reading admits
**146** merge subjects against the 5 that form 3a describes — a 29× widening — of which **50** are
caught by no attributing form at all. Two of them state the case:

    Merge branch 'WT-audit-evidence-store' into WT-audit-advice-integrity (t387 depends on t386 convention doc)
    Merge branch 'develop' into WT-inbox-drain-gap (absorb t280, lane-15 window; includes t239 merge e79c010b8)

The first names **t387 and t386**, as a dependency note; the second names **t280 and t239**, as an
absorb record. Neither is an attribution, and both are excluded by the non-attribution rule above.

**Why "exactly one" token in form 3b's group (plan-audit iter-2 D1).** Version 0.2.0 admitted **any**
token lying inside the trailing group, which contradicted this section's own preamble: a group
reading `(card t500 — absorb t280, includes t239)` would attribute all three — the occurrence
reading relocated inside a parenthesis. The single-token restriction removes the contradiction at
**zero measured cost**: of the 77 `into develop` merges whose trailing group carries a card token,
**none** names two distinct cards (per-line distinct-token count over the extracted groups → 0
lines). The shape is nevertheless real in the corpus at large — **8** trailing groups name two or
more distinct cards (`(t46/t73/t74)`, `(card t36, absorbs t2)`, `(t333/t347)`, `(t387 depends on
t386 convention doc)`, …) — so the restriction guards a population that exists, merely not yet on
this form's own subject set.

**The merge verb's casing is latitude, not prescription (card t486, from t482 §2.4).** Form 3b's
definition — named integration target, exactly one card token in the trailing group — does not
prescribe the merge verb's casing, and an implementation shall not add a case restriction the
corpus refutes: `^[Mm]erge … into develop` counts **77** where `^Merge …` counts **76**, and the one
diverging subject is `merge: WT-ci-test-observability into develop (t358)`
(`.moai/reports/t482/form3b-delta.txt`). That subject's trailing group `(t358)` carries exactly one
card token and the subject is attributed by form 2 regardless, so the residual is invariant under
the 76 → 77 correction — 347 / 309 / 38 and the 28-floor of §A.4.2 do not move (verdict.md §2.4).

#### A.4.2 What the enumeration costs — the under-count, measured

**What form 3b buys.** Under forms 1/2/3a the attributed-id set is **257**; form 3b adds exactly
**one** further id — `t412`, whose only landing evidence is the merge `b6231290d ... into develop
(card t412 — SPEC-MX-TAG-EDGES-001)`, whose trailing group carries text after the card id and so
escapes form 2's `)$` anchor. Set total **258**. Its three sibling subjects (`d8c91d907`,
`63435427c`, `57d2f3ae3`) all merge **into `WT-mx-tag-edges`** and attribute nothing.

**The claim version 0.2.0 made about the cost was wrong and is withdrawn (plan-audit iter-2 D2).**
It recorded *"one under-count survives — t250"*. Re-derived over the pinned corpus at HEAD
`75e63d6f2`, with a binary built from this tree:

    ids appearing anywhere in a subject                              → 347
    ids attributed under forms 1/2/3a/3b                             → 258
    subject-present but attributed by no form                        →  89

Of those 89, the ids whose subject evidence sits in an **attributing-shaped position the four forms
miss** — classified by naming the shape, not by eyeballing the list — number **at least 19**:

| Missed shape | Ids | Example subject |
|---|---|---|
| card-led local merge, `merge: <card>` | t106 t110 t113 t114 t32 t36 t56 t59 t69 t79 t98 t99 (12) | `merge: t114 — always-loaded budget trim 76680→75297 …` |
| bare `<card>:` prefix | t225 t250 (2) | `t250: graph freshness, symbol layer, and MCP code queries (#1648)` |
| non-exact trailing group | t46 t68 t73 t74 (4) | `merge: anchor-session guards for worktree disposal + registry CWD relocation (t46/t73/t74)` |
| legacy `worktree-` merge, card-led after the colon | t40 (1) | `Merge branch 'worktree-t40': t40 — moai update observability (3 quiet failures)` |

Nineteen, not one — a **19×** error, on a figure `plan.md` §D used as its ground for accepting the
loud-failure direction. The audit that found the defect measured **7**; the two sets differ because
it did not name the `merge: <card>` family, which is the largest of the four.

**The repair: form 3c, adopted on measurement.** The `merge: <card>` family is not a stray — it is a
convention used **31** times, and in **31 of 31** subjects the first token after `merge: ` is the
card the merge delivers (every subject read). It is a position (subject start), so admitting it
readmits no occurrence reading. Adopting it moves the attributed set **258 → 270** and takes the
named-shape under-count **19 → 7**:

    ids attributed under forms 1/2/3a/3b/3c                          → 270
    residual named-shape under-count                                 →   7
      t225 t250 (bare prefix) · t40 (legacy worktree- merge) · t46 t68 t73 t74 (non-exact group)

Form 3c under-counts in one measured way, and loudly: a multi-card merge (`merge: t85+t94 …`,
`merge: t92+t93 …`, `merge: t77+t64 …`) attributes only its first card. Three such subjects exist;
the second card in each is attributed elsewhere in the corpus.

**The widening this SPEC declines, and the measured ground for declining it.** The remaining 7 would
mostly be recovered by widening form 2 to accept a card token as the **first** token of a trailing
group. That widening is rejected on a corpus instance: `merge: t79 — glm_task delegation family
(branch WT-t80)` closes with the group `(branch WT-t80)`, whose only card token is **t80** — a
**branch name**. A widening that reads it attributes a card that did not land, on a commit belonging
to t79. The under-count is loud; that false positive would be silent, and silence is the failure
direction this SPEC exists to remove.

**The second repair: form 2b, and the third correction of this figure (plan-audit iter-3 D3-1).**
The 7 above was wrong too, and wrong in the same way — the classification of the residual has never
been exhaustive on any iteration, this one included. The audit named a shape neither the author nor
the lane had classified: form 2's own position with a pull-request reference appended, which form
2's `)$` anchor silently excludes. Re-measured in this tree at HEAD `012d9680a` over the pinned
corpus `7835148d3`, with a binary built from this tree (`go build ./cmd/moai`, VCI §2.2):

    grep -cE '\([^()]*t[0-9]+[^()]*\) \(#[0-9]+\)$' <pinned-corpus subjects>          →  43
    ... the same, ids extracted, sort -u                                              →  40
    ... of those, ids attributed by none of forms 1/2/3a/3b/3c (comm -23)             →  39

Adopting form 2b moves the attributed set **270 → 309** and the subject-present-but-unattributed
population **77 → 38**:

    ids appearing anywhere in a subject                                               → 347
    ids attributed under forms 1/2/2b/3a/3b/3c                                        → 309
    subject-present but attributed by no form                                         →  38

**[HARD] The residual is a FLOOR, not a total — since card t482, a measured floor.** The 38 have
since been classified exhaustively — every id, by shape — by card t482's audit
(`.moai/reports/t482/verdict.md` §4, verdict line "합계 38 = C 5 + M 28 + A 5"; raw 38-id dump in
`.moai/reports/t482/residual-evidence.txt`; measured 2026-09-04 in tree `.claude/worktrees/t482` at
HEAD `25a3212a9`, against this SPEC's pinned corpus `7835148d3`): **5 correct exclusions (C) + 28
clear omissions (M) + 5 judgment-deferred (A) = 38**. The clear-omission floor this section carried
as "at least 10" therefore stands at **28 measured** — the v0.4.0 floor understated the measured
clear omissions by 2.8×. 28 is still a floor, and this is not a closure: the 5 judgment-deferred ids
(`t311`, `t158`, `t460`, `t155`, `t157`) were decided on subject evidence alone, were never opened
to diff level, and opening them can only raise M — to at most 33 — never lower it. The round history
now reads 1 → 19 → 7 → ≥10 → 28 measured; every round raised the figure because each round
classified the part of the residual it happened to notice. The residual's live-queue status,
unmeasured at v0.4.0, is likewise measured by the same audit — 34 of the 38 absent from the queue
store, 4 archived, 0 live (verdict.md §2.5) — which bounds the residual's *current* operational
impact near zero and closes nothing.

The floor of 28 therefore stands as an accepted, measured cost, re-argued in `plan.md` §D on the
failure **direction** rather than on any dating or impact claim about the residual.

**The residual's largest single shape, named: S1 — the release-integrate merge scope (card t486,
from t482 §4.1).** Subject **begins** `merge(WT-…)` or `merge(worktree-…)`, the merge integrates
into a **release branch** (e.g. `merge(WT-t143): integrate into release/v3.1.1`, the other spelling
`merge(worktree-t132): …`), and the card token appears ONLY inside the merge scope — the branch
name — never as an attribution. Two figures name it and they measure different things
(`.moai/reports/t482/s1-reconcile.txt`; verdict.md §4.1): **shape size 18 subjects / 18 distinct
ids** (spellings: 10 `WT-` / 8 `worktree-`) and **residual contribution 14** — the other 4 of the 18
(`t119 t130 t145 t146`) are attributed by other forms elsewhere in the corpus and so never entered
the residual. S1 passed unnamed through this SPEC's three author versions and three plan-audit
iterations; it is named here so the next reader does not rediscover it. **[HARD] NAMING ONLY — the
operator explicitly rejected adopting S1 as a form** (verdict.md §5.1 cause 2; §9 판정 1): a rule
that reads `merge(WT-t131):` as attributing t131 attributes a **branch name** — a positional
exception the §A.4 non-attribution rule must not pay for. No attribution form is added.

**[HARD] The enumeration does not close — a property, not a count (card t486, from t482 §5.1).**
Three measured grounds: **(a) the residual population is open** — `t409` entered 2026-09-01 and
`t460` 2026-09-03, so new shapes were entering up to two days before the corpus pin; an enumeration
can close for the past, never for the future. **(b) Expanding the largest residual shape — S1, the
release-integrate merge scope (`merge(WT-…)` / `merge(worktree-…)` integrating into a release
branch; named in full above) — into a form collides with §A.4's own [HARD] branch-name
non-attribution rule**, the same silent over-count direction this SPEC already rejects at
`(branch WT-t80)`. The collision is technically escapable — the merge scope and a trailing
parenthetical are different positions, so a seventh form reading the scope token would not import
the t80 false positive — but the escape makes the rule read *"a branch name is never an attribution,
except in this position"*, and a rule with positional exceptions is no longer a ground of judgment
but a post-hoc list of judgment outcomes. **(c) The structural bypass — attributing a card from
inside a merge scope via ancestor propagation — is refuted by the corpus**: `git log d2ad26c90^2
--not d2ad26c90^1` yields one commit carrying no card token, so the t143 merge has no inheritable
attribution to propagate (verdict.md §2.6). The method's bias is itself measured: S1's true size is
18, of which 4 ids accidentally matched other forms and were therefore excluded from the residual —
**searching for a shape by reading only the residual list systematically underestimates that
shape's true size**, and the three rounds that re-counted the residual each time (1 → 19 → 7 → ≥10)
were exactly that under-counting procedure. The residual count is consequently not a number a
future round could be asked to shrink: a further round that re-counts the residual repeats the
defect this paragraph records, and the enumeration debt's termination path is card t359, not
another round (§D).

### A.5 The ref: the repository already knows the answer

    git symbolic-ref refs/remotes/origin/HEAD   → refs/remotes/origin/develop
    primary .git/HEAD                            → ref: refs/heads/main
    primary .moai/config/sections/git-strategy.yaml:7 → worktree_base_branch: ""

The git-level answer is already `develop`. The landed check ignores it and uses a compiled-in
constant, because the only level it consults — a config file in a checkout parked on `main` — is
empty there. The key's value **is** `develop` on the integration branch (`a9c61cf56`), but that
commit is not an ancestor of `origin/main`, so the value cannot take effect until a release carries
it across.

### A.6 What the four consumers of the key want

All four consumers of `git_strategy.worktree_base_branch` want the **integration** meaning; none
wants the release meaning. One value can therefore answer all four.

| Consumer | Location (`4bcac7079`) | Wanted meaning |
|---|---|---|
| Card-worktree base | `internal/cli/session_worktree.go:215` | integration |
| doctor check | `internal/cli/doctor_worktree_base.go:43` | integration |
| Landed ref | `internal/kanban/prlink_landed.go:75` | integration |
| SessionStart alignment | `internal/hook/worktree_base_branch.go:125` (the write; see below) | integration |

The fourth **writes** `refs/remotes/origin/HEAD` to match the setting. The key is therefore not
inert, and no requirement below may be justified on the premise that changing its value is
mechanism-free.

*Correction (plan-audit D2).* Version 0.1.0 cited `worktree_base_branch.go:156` as the write. Measured
at HEAD `e227871b4`: `:155` is `worktreeBaseBranchReadConfigReal`, a config **read**; the write is
`WorktreeBaseBranchSetHead(configured)` at **`:125`**, backed by `worktreeBaseBranchSetHeadReal`
(`git remote set-head`) at **`:170`**. The corrected citation is `:125` / `:170`.

*And the write and chain level 2 are measurably disjoint.* `RunWorktreeBaseAlignment` gates on the
primary checkout at `:92` and returns at `:97-100` when the configured key is **empty**:

    // internal/hook/worktree_base_branch.go:92-100, HEAD e227871b4
    if !WorktreeBaseBranchInPrimaryCheckout() { return data }
    configured := worktreeBaseBranchReadConfig(projectRoot)
    if configured == "" {
        // REQ-WBR-005: the neutral value performs no git-metadata read at all.
        return data
    }

Chain level 2 fires only when level 1 is **empty**; the writer fires only when level 1 is
**non-empty**, from the same primary-checkout root `LandedRefFor` reads. The two are mutually
exclusive by construction, so **no cycle exists**. §D records what remains.

### A.7 The ordering decision — [HARD]

**Axis F lands before axis A.** Correcting the ref first does not fix a symptom; it **grows the
false-positive population from 2 to 9** (§A.2). Nine cards would then read landed on a ref the
operator trusts, seven of them wrongly. The predicate is repaired first, and the ref chain is laid
on top of a predicate that can bear it. This decision binds the milestone order in `plan.md`.

**What form 3b yields in the M1-only window (plan-audit iter-2 D3), measured.** The ordering creates
a window in which M1's repaired predicate runs against the **un-repaired** resolver, so the landed
ref is still `DefaultLandedRef = "origin/main"` (`prlink_landed.go:41,74-80`). The audit inferred
from the rule — and recorded the inference as a Gap, having measured no `origin/main` corpus — that
a `develop`-hardcoding implementation would attribute all 76 `into develop` merges in that window
while a correct one attributes none. **That inference does not hold on this repository, and the
refutation is a measurement rather than an argument.** The predicate walks `git log <ref>`, so in
the window it walks `origin/main`'s history, not `develop`'s:

    <origin/main 7ad9f8534 subjects>                                  → 4,457
      of which merge subjects                                         →   101
      'into develop' merges with a card-bearing trailing group        →     0
      'into main'    merges with a card-bearing trailing group        →     0

Form 3b contributes **exactly zero attributions either way** in the window: a hardcoded-`develop`
implementation and a ref-derived one are behaviourally identical here. The window is therefore not
where the target-derivation defect bites, and the [HARD] ordering — which rests on the measured 2→9
growth — is untouched by it.

**The defect the audit found is real, and it is permanent rather than windowed.** A repository whose
integration branch is `main` — the default this tool ships against — gets nothing from a
`develop`-spelled form 3b, and a repository using some third name gets false attributions from
`into develop` merges it never made. That is why §A.4.1 states the derivation as [HARD] and
AC-TLA-003 clause 6 falsifies a spelled target by varying the resolved ref, rather than by relying
on a window in which the two implementations cannot be told apart.

---

## §B Requirements

Notation: GEARS. Requirement IDs are stable; milestone assignment is in `plan.md`.

### B.1 Milestone 1 — the attribution predicate (axis F)

- **REQ-TLA-001** (Ubiquitous) — The landed predicate shall decide `landed` on **attribution
  position** within a commit's subject line, and shall not decide it on the presence of the card
  token anywhere in the commit message.

- **REQ-TLA-002** (Ubiquitous) — The set of attributing positions shall be exactly the positional
  shapes enumerated in §A.4 — conventional-commit scope (form 1), trailing parenthetical, bare and
  `card`-prefixed, carrying nothing else in the group (form 2), card-led merge subject in its
  `Merge card` spelling (form 3a) and its `merge:` spelling (form 3c), and integration-targeted
  merge subject whose trailing parenthetical group carries **exactly one** card token (form 3b) —
  and shall include §A.4's non-attribution rule in its contraposed form: a merge subject that names
  a target shall attribute no card unless that target is the branch the resolved landed ref names.
  No shape shall be expressed as a bare occurrence test — "the subject contains the token" is not a
  position, and admitting it reintroduces REQ-TLA-001's defect through the enumeration. The
  enumeration and its non-attribution rule shall live in one named place in the implementation, so a
  further form is a reviewable one-place diff.

- **REQ-TLA-003** (unwanted) — The landed predicate shall not report `landed` for a commit whose
  only occurrence of the queried card token lies in the commit body.

- **REQ-TLA-004** (unwanted) — The landed predicate shall not report `landed` for a commit whose
  subject contains the queried card token in a **non-attributing** position, that is, a commit
  attributed to a different card.

- **REQ-TLA-005** (Ubiquitous) — The argv the landed check runs shall be constructed by a single
  exported builder, so a tripwire test asserts against the implementation's own construction rather
  than a transcription of it.

- **When** the landed query cannot be evaluated — no git, no such ref, a query error — the querier
  shall answer `unknown` rather than `not-landed` (**REQ-TLA-006**, event-detected; preserves the
  three-valued contract of `SPEC-TODO-LANDING-STATE-001`).

- **REQ-TLA-013** (Ubiquitous) — Form 3b's target, and the non-attribution rule's target comparison,
  shall be **derived from the resolved landed ref** at evaluation time, and shall not be a
  compiled-in branch name (§A.4.1). This is stated as its own requirement rather than folded into
  REQ-TLA-002 because it is the one clause of the enumeration whose violation is invisible in this
  repository's own corpus during the M1-only window (§A.7) and permanent in any repository whose
  integration branch is not `develop`. Numbered last, and listed here at the end of the M1 block,
  because the twelve preceding ids are stable across versions (`spec.md` §B notation).

### B.2 Milestone 2 — the ref chain and its disclosure (axes A+B)

- **REQ-TLA-007** (Ubiquitous) — The landed ref resolver shall resolve through an ordered
  three-level chain: (1) the configured `git_strategy.worktree_base_branch`, (2)
  `refs/remotes/origin/HEAD`, (3) `DefaultLandedRef`.

- **Where** level 1 yields an empty value, the resolver shall consult `refs/remotes/origin/HEAD`
  before reaching the constant (**REQ-TLA-008**, capability gate).

- **When** `refs/remotes/origin/HEAD` cannot be resolved, the resolver shall yield
  `DefaultLandedRef` and shall not fail the invocation (**REQ-TLA-009**, event-detected).

- **REQ-TLA-010** (Ubiquitous) — The `todo done` landing verdict line shall name the ref that
  answered, appended so that the existing `done <id> ` prefix every reader keys off is preserved.

- **While** the answering ref came from a level below the configured key, **when** a landing verdict
  is emitted, the command shall disclose on stderr which chain level supplied the ref
  (**REQ-TLA-011**, compound).

- **REQ-TLA-012** (unwanted) — The landed ref resolution shall not write any repository ref, and in
  particular shall not set `refs/remotes/origin/HEAD`.

---

## §C Constraints

- **C-1** — The three-valued `LandingAnswer` contract (`landed` / `not-landed` / `unknown`) is
  inherited from `SPEC-TODO-LANDING-STATE-001` and is not renegotiated here.
- **C-2** — `--require-landed` remains opt-in, and remains asymmetric: it refuses on `not-landed`
  only, and proceeds on `unknown`. The measured behaviour at `internal/cli/todo.go:555-578` is the
  baseline.
- **C-3** — The landed query performs no network I/O. Level 2 of the chain reads a local ref.
- **C-4** — A project that configures the key keeps its current behaviour exactly: level 1 still
  answers first.
- **C-5** — The queue root resolution (`resolveTodoQueueRoot`, one repository one queue) is not
  changed. The chain is added below the config read, not moved to a different tree.

---

## §D Exclusions

This section states what is **out of scope** for this SPEC and why, so no successor re-derives the
boundary.

### Out of Scope — axis C, landing evidence storage

- `archived_items` carries no landing column. Its measured schema is
  `seq, id, text, added_at, spec_id, state, position` — no landing SHA, no answering ref, no verdict
  instant. The `landing=` line `todo done` prints is stdout-only and is never persisted; all 126
  archived rows are closed with no stored evidence.
- This is recorded here as **measured context only**. The surface belongs to card **t359**, which is
  picked and carries plan-audit iteration-1 redesign items. Authoring requirements against it here
  would duplicate a live card.
- **The enumeration of §A.4 is a transitional instrument, and card t359 is its sole termination
  path** (card t486, from t482 §5.2(1) and §6 recommendation 4). Landing-time recording of landing
  evidence — the merging side writing what it observed into the store, and the predicate reading
  that — is what ends the enumeration debt; expanding the enumeration further is mitigation, not
  resolution. Further enumeration rounds are **not the plan**; §A.4.2's non-closure property is
  why.

### Out of Scope — axis D, the blank card-to-SPEC link

- Re-measured in this tree: `items` 57/57 rows with a blank `spec_id`; `archived_items` 126/126
  blank. The attaching path (`todo next <n> --spec <SPEC-ID>`) exists and is used zero times.
- The card excludes this axis explicitly. It is an operational gap, not a schema defect, and it is
  recorded as context only.

### Out of Scope — axis E, a new detection subcommand

- The detection surface already ships: `moai todo pr` reads the whole live queue (57 rows), emits a
  five-value outcome column, and writes nothing.
- What is missing is that column's **precision**, which REQ-TLA-001..004 supply, and an operational
  routine for running it, which is not code.
- Representative mutant, declared a non-goal: building a new query subcommand. An implementation
  that adds one has not delivered this axis; it has re-delivered `todo pr` under a second name.

### Out of Scope — the level-2 / SessionStart-writer interaction

- Measured disjoint (§A.6): chain level 2 fires only on an empty configured key, and
  `RunWorktreeBaseAlignment`'s write fires only on a non-empty one, from the same primary-checkout
  root. No cycle exists, so nothing here is a requirement.
- What genuinely remains is the converse and it is **not** a coupling: a project that **configures**
  the key never reaches level 2 at all, so the two surfaces never interact on this path. Should a
  future change make the writer fire on an empty key, that disjointness ends — but repairing it then
  is that change's obligation, not this SPEC's.

### Out of Scope — release-wait as the axis-A remedy

- Waiting for `develop` to reach `main` was rejected by lead verdict: its timing is bound to the
  t204 deploy gate, and the closed-but-picked backlog recurs on every release cycle regardless.

### Out of Scope — semantic landing ("has this card's last step landed")

- The predicate answers "has a commit attributed to this card landed on the ref", not "has this
  card's sync step landed". That limit is inherited verbatim from `SPEC-TODO-LANDING-STATE-001` and
  is not narrowed here.

---

## §E Cross-references

- `SPEC-TODO-LANDING-STATE-001` — the ancestor: the ref became resolvable, and the answer became
  three-valued. This SPEC corrects the predicate that ref is asked with, and completes the
  resolution chain.
- `.moai/reports/t472/premise-recheck.md` — the card's original axis-A premise, re-checked in this
  tree and found false.
- `.moai/reports/t472/axis-bf-measurement.md` — the axis B–F measurements and the merged axis A+B
  design section.
- `internal/kanban/prlink_landed.go`, `internal/cli/todo.go`, `internal/cli/todo_pr.go` — the
  implementation surface.
