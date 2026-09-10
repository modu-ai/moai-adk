---
id: SPEC-MOVING-REF-GUARD-001
title: "Moving-ref invariant guard: warn on unpinned invariant claims, with the anchor-or-subject exemption predicate shipped alongside"
version: "0.9.0"
status: completed
created: 2026-08-28
updated: 2026-09-02
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/spec, .claude/rules/moai/core"
lifecycle: spec-anchored
tags: "lint, verification, evidence, moving-ref, invariant, preserve, baseline-attribution"
era: V3R6
tier: M
related_specs: [SPEC-GRAPH-FRESHNESS-CADENCE-001, SPEC-VERIFICATION-COMPLETENESS-001]
---

# SPEC-MOVING-REF-GUARD-001 — Moving-ref invariant guard

## HISTORY

| Version | Date | Change | Author |
|---------|------|--------|--------|
| 0.9.0 | 2026-09-02 | **Resumption — the option C deferral of §H Q0 is superseded by operator follow-up card t353.** REQ-MRG-010 (the R4-form lint exclusion), AC-MRG-013, and both counter-mutations CM-1/CM-2 were delivered at commit `0026dc7b9` (branch `WT-r4-imperative-exempt`): the exclusion in `internal/spec/lint_movingref.go` keys on **imperative structure** — both conjuncts, an imperative measuring directive plus a syntactically demoted dated reference — and never on a command token, per the [HARD] Q0 constraint the deferral recorded, implemented as written. The resume condition was re-measured **MET** on 2026-09-02: live external R4-form occupants exist (`SPEC-IGNORED-EVIDENCE-CITATION-001/progress.md:480`, self-declared R4; `SPEC-SPECLINT-GITBLIND-001/progress.md:233`, Korean-form — both previously silenced via R3 markers), so the exclusion acquired something to under-exempt and the trade-off that decided option C inverted. Lane evidence: RED observed pre-fix → GREEN after (13 MovingRef tests); CM-1's positional mutant and CM-2's command-token mutant both observed blocked — **the unrun counter-mutation obligation the deferral carried is discharged, not merely travelling**; corpus lint 115→113 findings with exactly the two genuinely-R4 lines newly exempted, zero over-exemption. Verdict `.moai/reports/t353/verdict.md` (commit `d191d28b4`); implementation record `.moai/reports/t353/impl-record.md`. This row and the §H Q0 resumption record are the only body edits of this revision: the REQ-MRG-010 deferral annotation in §E and the L6 prospective qualifier in §F are superseded by this row and left verbatim as history. `status:` deliberately unchanged — recorded in HISTORY rather than by reopening the lifecycle. | manager-spec |
| 0.8.0 | 2026-08-29 | **Retraction of REQ-MRG-009's `--strict` gating clause; the finding ships advisory.** Post-close body amendment on a `completed` SPEC, forced by measurement rather than by review. The run-phase merge `84b3b7949` turned `develop`'s SPEC Lint job from success to failure and it stayed red across `1a635aea8`, `91ec14dbe` and `a6bbbf82b`; the preceding head `c48e8aaa2`, which does not contain this card, was green, so the attribution is unambiguous. At `a6bbbf82b`: `0 error(s), 155 warning(s)` → exit 1, **110 of them `MovingRefUnpinned`** (the other 45 carry `[grandfathered era — downgraded to warning]` and are advisory, so they do not gate). `.github/workflows/spec-lint.yml` runs the linter with `--strict`. Reproduced in this worktree at `a7763666d`: `moai spec lint --strict` → rc 1, `0 error(s), 174 warning(s)`. **The defect is a composition nobody ran**: §D.5 reasoned to `warning` precisely so the corpus would not red on day one, and REQ-MRG-009 then required the `--strict` escalation that reds it — two clauses written against each other, with AC-MRG-005 asserting the same `--strict` non-zero and therefore agreeing with the defect instead of catching it. **The mechanism is emission-site advisory, not the era map**: `eraDemotableCodes` is consulted only for `SeverityError` findings, so a warning never reaches it and adding the code there would do nothing; the gating findings are on modern-era (V3R6) SPECs, which no era path can demote. The rule sets `Advisory: true` on its own findings, as `StatusGitConsistency` already does. **Applied**: REQ-MRG-009 rewritten (exit 0 with and without `--strict`, findings still reported, retraction and mechanism recorded inline); §D.5 extended to name the hole in its own argument, where it was found (develop CI, not this SPEC's tests), and advisory emission as the completion of §D.5's intent rather than a retreat from it; **new §D.7** carrying the three-part disclosure — (a) why advisory now (110 findings make bulk suppression the rational response to a day-one gate, the outcome §A exists to prevent), (b) the promotion condition (the `.moai/reports/t342/corpus-triage.md` findings remediated or R3-exempted by their owning SPECs, at which point the retracted clause returns verbatim), and (c) the failure mode of forgetting, stated without softening: an advisory-forever guard is **the same shape this SPEC exists to detect** — a claim true when written, quietly emptied afterwards, with nothing signalling the change, and by `acceptance.md` §A's own words indistinguishable from a guard that is switched off. AC-MRG-005 rewritten to decide the advisory behaviour and made falsifiable in **both** directions — exit 0 under `--strict` **and** a non-zero reported finding count — with a separate mutation per half (drop `Advisory: true` → strict exit reds; unregister the rule → count drops to 0 while both exits stay 0), because exit 0 alone passes vacuously against a rule that emits nothing. Also swept: `acceptance.md` §A's fixture-era precondition, whose stated reason was the retracted `--strict` half and which now holds for the opposite reason (an era-demoted fixture would be advisory twice over and could not observe the rule's own flag); `plan.md` §B, which claimed `warning` alone was sufficient and mis-attributed the severity fix to REQ-MRG-009. `status:` deliberately unchanged — recorded in HISTORY rather than by reopening the lifecycle. No Go source, no test, no doctrine file, no template mirror, no `progress.md` touched; the implementation is a separate delegation. | manager-spec |
| 0.7.0 | 2026-08-28 | Acceptance-criteria repair round, on six defects the run-phase agents raised and deliberately did not patch themselves (`acceptance.md` body is manager-spec's). Every one is a criterion that was true when written and became false or vacuous later; none is a criterion reshaped to fit the artifact it decides. **(a) AC-MRG-011's second half replaced.** `grep -c 'ANCHOR' ≥ 1` guarded nothing — `ANCHOR` is one of the predicate's two class names, so with L7 deleted the count half correctly drops 7 → 6 while the ANCHOR half returns **9**. Re-keyed on L7's distinguishing phrase (`the ANCHOR branch of the predicate is unvalidated`), deliberately not on `L7 —`, which would re-measure the count half's axis. Mutation re-planted: count `7 → 6`, phrase `1 → 0`, retired key steady at 9. **(b) AC-MRG-011's stale figure** corrected 5 → **6**: the limit set grew 5 → 6 → 7 across revisions and the mutation text was not swept with it. **(c) AC-MRG-002's fixture wording** corrected from "`origin/main` **replaced by** a SHA" — which deletes the token, fails conjuncts 1+2 rather than being exempted, and makes the criterion's own mutation unable to redden it — to the **retaining** form the implementation already used and REQ-MRG-008 is worded for. Mutation re-planted (delete the hex-SHA branch): RED. **(d) AC-MRG-009's separation clause KEPT, by re-placing the fixture rather than deleting the clause.** At M3 the body-only mutant took seven criteria red together, AC-MRG-001 among them, because its fixture row sat in `acceptance.md` while `SPECDoc.Body` carries `spec.md` alone. The row was moved into the fixture's `spec.md` and the mutant re-run: **AC-MRG-009 red, AC-MRG-001 green** (with -002, -005, -008, -014 also green; -003/-004/-006 collateral-red on the same sibling property). The separation is now measured rather than asserted. **(e) §H Q0 answered — option C, operator decision**: REQ-MRG-010, AC-MRG-013, CM-1 and CM-2 deferred out of scope, with grounds (R4's reachable class measured empty twice — 0 of 42 by §B.7's probes, external occupancy 0 against M4's 97 findings — so the exclusion can only over-exempt, and the iter-1 audit measured a token-keyed exclusion passing all thirteen criteria while silencing 76 of 117 real divergence lines), and with a resume condition M4 sharpened: 0 external S2 lines but **2 S2 lines in this SPEC's own directory**, both R4-form, the class filling prospectively from this card's own output exactly as §B.7 predicted. R4 stays in the doctrine and in the finding message; only its lint exclusion defers. The unrun counter-mutation obligation travels with the requirement rather than being discharged. **(f) AC-MRG-010's decider re-anchored** from `$BASELINE_SHA` to a second frozen literal `MERGE_BASELINE_SHA`, recorded at `plan.md` §C step 5. The old range starts before this branch diverged, so it spanned mainline and returned four paths from cards t322/t303 at every measurement — a wrong-reason red that also could not distinguish its own stated mutation from a routine merge. The new anchor stays a *recorded* value, never computed at read time, and carries a [HARD] re-recording obligation on any further mainline absorption (limit L6 acting on this SPEC's own evidence, handled by R2 + R4). Also swept: AC-MRG-001's fixture location, §B's baseline anchor (two literals, two questions), §D DoD, `plan.md` M3 and §C, and §F L6 marked prospective while the exclusion is deferred. `verification-claim-integrity.md`, its template mirror, and `lint_movingref.go` are byte-identical to their pre-round state; `lint_movingref_test.go` carries one assertion change (`acceptance.md` → `spec.md`) that the (d) re-placement makes necessary. | manager-spec |
| 0.6.0 | 2026-08-28 | Run-phase M1-M2. **M1**: the doctrine section landed as `### 2.1 Moving-ref attribution — the anchor-or-subject predicate` inside `.claude/rules/moai/core/verification-claim-integrity.md`, placed directly under §2 as a specialization of the baseline-attribution invariant it enforces. It carries the four predicate tests with Test 1's read-time evaluation and the tie-break routing a Test 1 / Test 3 disagreement to Test 4 (never to ANCHOR), the two-classes-four-remedies statement, R1-R4 with the §D.4 pricing table, the five grounded instances with the classification skew in both readings, the marker surface, all seven limits L1-L7, and the divergence-figure carrier split. Deciding greps recorded verbatim in `progress.md` §E.2. **M2**: the marker surface is fixed — syntax, mandatory non-empty reason, line scope (flagged line or the line immediately above), and the incomplete-marker finding of REQ-MRG-003. **Q1 resolved**: the invisible HTML-comment form is adopted on the grounds that the marker is an author-to-linter annotation rather than rendered content, with the reviewer-visibility concern met by the reason's visibility in the source diff. **Finding raised for the operator, not silently patched**: AC-MRG-011's third mutation (delete L7 → the limit count drops **and** `grep -c 'ANCHOR'` returns 0) is only half-falsifying against this doctrine. `ANCHOR` is a class name used throughout the predicate, so deleting L7 moves the count 10 → 9, never to 0; the `≥ 1` half of that criterion is satisfied by the predicate regardless of L7 and therefore guards nothing. The count half (`grep -c 'L[1-7] —'` 7 → 6) does go red as specified. M3 (detector), M4 (corpus triage) and M5 (template mirror) are untouched; Q0 remains with the operator and M3 cannot be implemented without it. | manager-develop |
| 0.5.0 | 2026-08-28 | D13 closed on its second condition — R4's scope characterized by measurement instead of asserted. **The reachable scope is empty: 0 of the 42 candidate lines** (§B.7, two independent probes). R4 is one day old, so no corpus line is written in its form, and the class is populated only by remediations the doctrine itself will produce. REQ-MRG-010's "does real work on exactly one class" is therefore corrected to the measured truth — the class is nameable but currently unoccupied, and AC-MRG-013 is a synthetic-only criterion with no live counterpart. Per the lead's instruction not to manufacture a scope to keep a clause alive, **§H Q0 is answered on that basis and now carries a scope decision for the operator**: option C (ship R4 as doctrine, defer the exclusion, silence early R4 lines with the R3 marker) eliminates the D11 bypass entirely and converts Q0's unsolved shape-recognition problem into the already-solved marker problem, at the cost of one marker per line until volume justifies the exclusion. Recommended, not taken — it changes scope. §B.7 also records that the earlier `81` re-cited as current had drifted to `86`, this card's own subject recurring a fourth time. | manager-spec |
| 0.4.0 | 2026-08-28 | Plan audit iter-2 (PASS-WITH-DEBT 0.86, +0.06 monotonic, `.moai/reports/t342/plan-audit-iter2.md`) closed all eight iter-1 blocking defects and raised two new ones. Applied here as a **targeted pre-run edit**, not a plan-phase re-entry — iter-2 was the Tier M ceiling. **D13** (critical): AC-MRG-013's fixture was the lead's dispatch line, which cannot fail — re-measured independently here, it carries **0** `origin/<branch>` slash tokens (`git fetch origin develop` is two arguments, an L1 blind-spot line by this SPEC's own §F), **1** hex SHA (so REQ-MRG-008 already exempts it) and **0** claim markers, and the full REQ-MRG-001 pipeline returns **0** against it. Removing the R4 exclusion entirely therefore did not make the finding reappear, so the card's [HARD] deliverable had no criterion proving the exemption exists at all — the suite guarded thoroughly against an exclusion that is too *wide* and not at all against one that does nothing. Fixture replaced with a REQ-MRG-001-residual-class line whose six properties are measured in `progress.md` §E.1, and REQ-MRG-010 now carries the residual-scope sentence the old fixture was hiding: the exclusion does real work on exactly one class, because a SHA-valued R4 reference is already exempt via REQ-MRG-008 and a divergence-count reference via REQ-MRG-006's date conjunct — both verified by measurement. **D14** (major): §D.3's ANCHOR-branch limitation was publishable-optional — `grep -n 'ANCHOR' acceptance.md` returned nothing, so M1 could drop the one limitation this audit round added while passing every criterion. Promoted to **L7**, which pulls it into AC-MRG-011 with the count raised 6→7, the same handling L6 received at v0.2.0. Optional **D15** (Test 4's gate named Tests 1-3 though Test 2 is conditional and never returns a class — gate reworded to Tests 1 and 3, with Test 2's non-classifying role stated) and **D16** (§B.3's "three dispositions the predicate produces" was the triage taxonomy, not the predicate's — the same residue swept from plan.md at v0.3.0) both applied. | manager-spec |
| 0.3.0 | 2026-08-28 | Remediation of plan audit iter-1 PASS-WITH-DEBT 0.82, revised 0.80 after the targeted re-audit (`.moai/reports/t342/plan-audit.md`, written in this worktree because the isolation guard refused the primary path). Eight blocking defects fixed, four optional accepted. **D1** (critical): Test 1 carried no evaluation time, and the "Test 1 governs" tie-break routed the predicate's own founding case back to ANCHOR — so R4 was unreachable for the case it was invented for. Test 1 now specifies read-time evaluation and the tie-break routes a Test 1 / Test 3 disagreement to Test 4 instead of resolving to ANCHOR. The v0.2.0 correction had fixed the worked example without fixing the procedure; that is now recorded as the distinction it is. **D12** (major): §D.4's incentive argument was still the two-remedy one, leaving R4 an unpriced exemption — an author could rephrase into R4 shape and be done, which is the bulk-suppression outcome by another route. §D.4 restated for four remedies and R4 given an explicit cost; L6 extended with the incentive residual alongside the freshness one. **D11** (critical): an R4 exclusion keyed on the fetch verb passes all thirteen criteria including AC-MRG-013's counter-mutation. Second counter-mutation fixture added, and Q0 now requires the R4 signature to key on imperative structure rather than any command token. **D2** (critical): no negative control on the invariant-claim conjunct — AC-MRG-014 added. **D3**: fixture era precondition stated so AC-MRG-005's `--strict` half is satisfiable. **D4**: AC-MRG-010 now decided on both the committed diff and the working tree. **D6**: `version` was left at 0.1.0 while HISTORY carried a 0.2.0 row — the documented-vs-actual divergence this SPEC prohibits, in its own frontmatter; bumped here to 0.3.0 with the miss recorded rather than silently corrected. **D7**: two pre-addendum citations in plan.md swept. Optional **D5** (DoD now Q0-Q4, Q4 being newly opened by D1's tie-break change), **D8** (`Where` → `When`/`While` on five requirements), **D9** (REQ-MRG-011 added; AC-MRG-010 is no longer an orphan) and **D10** (the three grep alternations recorded verbatim in `progress.md` §E.1) also applied. §D.3 now states the ANCHOR-branch validation gap as a limitation, not only the skew as a strength. Every figure re-measured in this tree at `43329ec8b` rather than carried from the audit — two of the auditor's D11 figures did not reproduce under the now-recorded alternations, and the re-measurement is worse than their estimate; see `progress.md` §E.1. | manager-spec |
| 0.2.0 | 2026-08-28 | Lead addendum absorbed. **(1)** Grounded instance 3 re-attributed: the stale-base occurrence is the **lead's own dispatch block**, named as such rather than as an anonymous occurrence, and §B.4 now states that it sits on the same axis as the card's side-discipline — relaying a once-measured value without re-measuring is the form this card prohibits. The attribution was re-verified in this turn before being written (`git rev-parse ec15ec2cd^` → `44095ddc2…`; `git reflog show WT-moving-ref-guard` → created from `origin/develop` at `ec15ec2cd`), because citing the dispatched figure unverified would have made this SPEC an instance of its own defect. **(2)** Fifth grounded instance added (§B.5, §D.3): the lead's change to the dispatch base line, from `base: origin/develop <sha>` to `base: measure at entry … (dispatch-time reference value: <sha>)`. **(3)** The predicate is **kept two-way** at the top level but gains **Test 4**, which routes within the SUBJECT class, and the remedy space expands from three branches to **four** with the addition of R4 (state the measuring command as the criterion, demote the value to a dated reference). §D.1 now states explicitly that classification and remedy selection are two steps, because conflating them is a second, subtler route to the card's dominant failure mode. Why the new case does not become a third top-level branch, and why it does not fold into ANCHOR, is argued in §D.1.1. **(4)** New REQ-MRG-010 / AC-MRG-013: the R4 form must not be flagged. **(5)** New detection limit L6 — the detector cannot tell a live reference value from a rotted one — with AC-MRG-011's grep count raised from 5 to 6. | manager-spec |
| 0.1.0 | 2026-08-28 | Initial plan-phase authoring for card t342, in worktree `.claude/worktrees/t342` at HEAD `ec15ec2cd`. Every figure in §B was measured in this tree; the corpus survey (§B.3) and the two-dot/three-dot discriminator experiment (§B.2) were run here rather than carried over. The card's own dispatch supplied a third live instance of its subject (§B.4), verified against `git reflog show WT-moving-ref-guard`. The exemption predicate (§D.1) is specified as the primary deliverable per the card's [HARD] clause, and milestone ordering (plan.md §F) puts it first for that reason. | manager-spec |

## §A. Problem Statement

An invariant claim — "byte-unchanged", "PRESERVE", "this path is absent from the diff" — is
asserted by naming a comparison anchor and reporting what the comparison returned. When that anchor
is a **moving ref** (`origin/main`, `origin/develop`, `origin/HEAD`), the anchor's value changes
every time upstream commits. The claim's *text* does not change. Its *truth* does.

The failure is silent in the direction that matters. Nothing errors; the claim simply stops meaning
what it meant when it was written, and the next reader who re-runs it gets a different answer with
no signal that the anchor moved underneath them.

The card's measured instance: a `plan.md` recorded its PRESERVE proof as
`git diff --stat origin/main -- <path>` and the command returned `18 files changed / 2825 deletions`
**before any work had been done in that tree** — `origin/main` had advanced ten commits in the
interval between authoring and reading. The evidence offered for "every existing file is
byte-unchanged" was false on arrival, and false for a reason that had nothing to do with the work
it was describing.

This is a `verification-claim-integrity.md` §2 defect specifically: the claim is not attributed to
an actually-measured baseline, because the thing it names as its baseline is not a fixed value.

### What this SPEC does, and what it deliberately does not

It adds a **warning** that surfaces the shape — an invariant claim decided by a moving ref with no
pin — and it ships, in the same delivery, the **predicate that decides whether the warning should
be acted on by pinning at all**.

The predicate is not an accompaniment. It is the primary deliverable, and §D.1 exists because a
warning shipped without it has a well-understood failure mode, stated by the card: the next person
mechanically pins every occurrence. Pinning is right for most of them and **wrong** for the class
where the moving ref is the claim's *subject* rather than its anchor — a provenance narrative about
what mainline currently carries is weakened, not strengthened, by being frozen to one commit.

## §B. Measured Baseline

Every figure below was measured in `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t342`, branch
`WT-moving-ref-guard`, HEAD `ec15ec2cd`, on 2026-08-28. Each row names the command that produced it.

### B.1 — The tree this was measured in

```
git rev-parse --show-toplevel  → /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t342
git branch --show-current      → WT-moving-ref-guard
git rev-parse --short HEAD     → ec15ec2cd
git status --short             → (empty)
```

No work had been performed in this tree at measurement time. That matters for §B.4: every non-empty
diff reported there is upstream drift and nothing else.

### B.2 — Three-dot is **not** a safe form (a tempting mitigation, disproved)

A plausible first reading is that `git diff A...B` (merge-base relative) is immune to upstream
advance, and that the guard could therefore exempt the three-dot form. Measured here, it is not.

Same claim ("`internal/hook` is unchanged in my work"), same tree, two anchor values — `44095ddc2`
(one integration stale) and `ec15ec2cd` (current, `== HEAD`):

| Command | Output | Claim reads |
|---|---|---|
| `git diff --stat ec15ec2cd -- internal/hook` | *(empty)* | true |
| `git diff --stat 44095ddc2 -- internal/hook` | `3 files changed, 288 insertions(+)` | **false** |
| `git diff --stat 44095ddc2...HEAD -- internal/hook` | `3 files changed, 288 insertions(+)` | **false** |

The three-dot form gave the identical wrong answer. The reason is that merge-base is stable only
while the upstream advance **diverges** from the branch; once the branch absorbs that advance — by
merge, rebase, or by being created from the newer tip — `merge-base(stale, HEAD)` moves to the stale
ref itself and three-dot degenerates to two-dot. Confirmed:
`git merge-base ec15ec2cd 44095ddc2 → 44095ddc2cc1c9fed2b3bd5ac946f48017988aba`.

Consequence for the detector: `...` MUST NOT be treated as an exemption (AC-MRG-004). Recording the
disproof here is deliberate — it is the mitigation a later reader is most likely to propose.

### B.3 — Corpus survey: how much of this shape exists today

Over `.moai/specs/**/*.md` in this tree:

| # | Filter | Count | Command |
|---|---|---|---|
| 1 | Lines naming a moving ref at all | **1477** | `grep -rnE 'origin/(main\|develop\|HEAD)' .moai/specs --include='*.md' \| wc -l` |
| 2 | (1) inside a `git <verb>` command context | **527** | (1) piped through `grep -E 'git [a-z-]+[^\`]*origin/…'` |
| 3 | (2) also carrying an invariant-claim marker | **53** | (2) piped through the claim-marker alternation |
| 4 | (3) with no 7-40 hex SHA on the line | **42** | (3) piped through `grep -v -E '\b[0-9a-f]{7,40}\b'` |

**42 is the raw candidate set, not 42 defects.** The set was enumerated and read; it contains all
three **triage** dispositions — anchor-unpinned, subject-correctly-unpinned, and already-frozen —
which is the useful result. These are the categories the M4 corpus triage sorts into, not the
predicate's own classes (two, with the S1/S2 split — §D.1):

- **Anchor, unpinned — genuine finding.** `AC-TST-010` (`SPEC-V3R6-TEST-REFACTOR-001/acceptance.md:35`,
  `git diff origin/main -- .moai/specs/…` deciding "no predecessor SPEC body modified"); `AC-AFS-012`
  (`SPEC-V3R6-AGENT-FOLDER-SPLIT-001/progress.md:126`, `git diff origin/main -- <9 files> | wc -l`
  deciding "0 (byte-identical)"); `AC-MSD-010`; `AC-WC-009`; `AC-CFS-010`; `AC-NS2-005`.
- **Subject, correctly unpinned — must not be flagged.** `AC-COORD-016`
  (`SPEC-V3R6-MULTI-SESSION-COORD-001/progress.md:296`) asserts that the *literal string*
  `git rev-list --count --left-right origin/main...HEAD` is preserved verbatim inside a document.
  The ref token is quoted subject matter, not a measurement anchor. `REQ-LB-006`
  (`SPEC-V3R5-LATE-BRANCH-001/spec.md:168`) documents `git reset --hard origin/main` as a
  doctrine step — again the command is the subject.
- **Already remediated by freezing, not pinning.** `SPEC-DESIGN-MOAIWEBV2-002/plan.md:36` records
  `BASELINE_SHA=$(git rev-parse origin/main)` captured *before* the first run-phase commit, with
  `acceptance.md:14` deciding its criteria against `$BASELINE_SHA`. This is a third remediation
  branch that pre-dates this SPEC and is promoted to the recommended default in §D.2 (R2).

### B.4 — The lead's dispatch reproduced this card's own subject (live, this tree)

**Attributed to the lead**, who owns the instance and asked for it to be recorded as theirs rather
than as an anonymous occurrence. The dispatch block named the base as `origin/develop 44095ddc2`.
The tree was created from a later tip:

```
git reflog show WT-moving-ref-guard
  ec15ec2cd WT-moving-ref-guard@{0}: Branch: renamed refs/heads/worktree-t342 to refs/heads/WT-moving-ref-guard
  ec15ec2cd WT-moving-ref-guard@{1}: branch: Created from origin/develop
```

`origin/develop` was `ec15ec2cd` at creation; `44095ddc2` is its parent
(`git log --oneline -3 origin/develop`). The dispatch's ref value went stale between being written
and being acted on — one integration's worth — and the §B.2 table is the consequence: a
`44095ddc2`-anchored PRESERVE check in this tree reports `3 files changed, 288 insertions(+)`
against work that does not exist.

This is the card's subject occurring in the act of dispatching the card, and it is used as
grounding instance 3 in §D.3.

**It is the same axis as the card's side-discipline, not a neighbouring one.** The side-discipline
says a once-measured divergence value must not keep being cited without re-measurement; the stale
base line is that rule violated on a base SHA instead of a divergence count. Both are a measurement
whose validity expired being re-served as current. §D.6 therefore treats them as one defect on two
carriers rather than as two rules.

The attribution above was **re-verified in this turn before being written** rather than copied from
the dispatch — `git rev-parse ec15ec2cd^` → `44095ddc2cc1c9fed2b3bd5ac946f48017988aba`, and the
reflog line quoted above. Citing the dispatched figure on the dispatcher's word would have been the
prohibited form, in the document that prohibits it.

### B.5 — The lead's dispatch-format change (attributed decision, 2026-08-28)

Recorded on the lead's authority as a design decision, **not** as an observed practice — see the
Gap note in `progress.md` §E.1. The base line changes from:

```
base: origin/develop <sha>
```

to:

```
base: measure at entry with git fetch origin develop (dispatch-time reference value: <sha>)
```

The lead's stated reasoning, quoted because it names the mechanism more precisely than a paraphrase
would: *writing the value first makes it read as the criterion, and demotes re-measurement to a
confirmation step. Inverting the order is the fix.*

This is the observation that produces remedy **R4** (§D.2) and Test 4 (§D.1). Its significance is
that the correct remedy here is neither pinning nor keeping a bare moving ref: it is **stating the
measuring command as the criterion and demoting the value to a dated reference**. A reader who
knows only "pin it" has no way to reach this form.

### B.6 — The fetch-verb bypass, re-measured (audit D11)

The iter-1 audit found that an R4 exclusion keyed on the **fetch verb** — the most salient token in
R4's only exemplar, and the natural thing to reach for while Q0 leaves the signature ungeneralized —
passes all thirteen acceptance criteria including AC-MRG-013's counter-mutation, because no other
criterion's fixture carries a fetch verb.

**Re-measured here at `43329ec8b` rather than carried from the audit**, with every alternation
recorded (`progress.md` §E.1). Two of the auditor's figures did not reproduce under this SPEC's own
alternations, and one did:

| Figure | Audit | This re-measurement | Command |
|---|---|---|---|
| Candidate lines, diff-shaped class | 100 | **52** | filters 2-4 of §B.3, `grep -v SPEC-MOVING-REF-GUARD-001` |
| Of those, carrying the fetch verb | 36 | **6** | same, `grep -c 'git fetch'` |
| Corpus lines pairing fetch with `rev-list --count --left-right` | 81 | **81** | `grep -rn 'rev-list --count --left-right' .moai/specs --include='*.md' \| grep -c 'git fetch'` |

The divergence on the first two is D10's consequence, not a dispute: the v0.2.0 artifacts described
the claim-marker alternation instead of quoting it, so the auditor reconstructed a different one.
That is precisely why the alternations are now recorded verbatim.

**The re-measurement makes the finding worse, not weaker.** The diff-shaped class is the wrong place
to size this bypass; the class it actually lands on is the divergence form of REQ-MRG-006, and
measured directly:

```
grep -rn 'rev-list --count --left-right' .moai/specs --include='*.md' \
  | grep -v SPEC-MOVING-REF-GUARD-001 | grep -cvE '\b[0-9a-f]{7,40}\b'   → 117
… | grep -c 'git fetch'                                                   → 76
```

**76 of 117** unpinned divergence lines — 65% — would be silenced. The pairing is structural rather
than incidental: the Pre-Spawn / Pre-Edit Sync Check block in `agent-common-protocol.md` is a fetch
followed by `git rev-list --count --left-right origin/main...HEAD`, so every progress record quoting
that block carries both tokens on one line. A sample of the silenced set is in `progress.md` §E.1.

The consequence for the design is in §H Q0: the R4 signature keys on imperative structure, never on
a command token.

### B.7 — R4's reachable scope, measured: empty today

D13's second condition. An exemption whose scope is empty is not an exemption but a dead clause, so
REQ-MRG-010's reach is measured here rather than asserted.

**Probe 1 — narrow.** Of the 42 REQ-MRG-001 candidate lines (§B.3 filters 2-4, this SPEC's own
directory excluded), how many are additionally in an imperative measure-instruction shape?

```
… filters 2-4 …  | grep -ciE 'run `git|by running|measure at entry|reference (value|reading)|at read time'
→ 0        (denominator re-confirmed in the same pipeline: 42)
```

**Probe 2 — deliberately broader**, in case the first alternation was too specific:

```
… filters 2-4 …  | grep -ciE 'reference|as of|at entry|re-measure|remeasure|재측정'
→ 0
```

**Both return zero. R4's reachable scope in the live corpus is empty.**

The reason is not subtle and is not a defect: R4 was invented one day before this measurement, so no
existing line is written in a form that did not exist when it was written. The class is **nameable
but currently unoccupied** — a REQ-MRG-001-class line in R4 form whose demoted reference value is a
measurement result rather than a commit SHA. Its future occupants are created by the doctrine that
ships alongside it: every R4 remediation M1 recommends produces exactly one.

Two consequences, both recorded rather than smoothed over:

- **AC-MRG-013 is a synthetic-only criterion.** Its fixture has no live counterpart and cannot
  acquire one until the doctrine is in use. It proves the exclusion works against a constructed
  line; it cannot demonstrate the exclusion is needed.
- **The exclusion can currently only over-exempt.** With no occupants, REQ-MRG-010 has nothing to
  under-exempt, so every error it can commit today is a D11-shaped one. This is what makes the Q0
  option in §H worth the operator's attention.

**Count-drift note.** The `81` figure for fetch ∧ divergence lines (§B.6, measured at `43329ec8b`)
re-measures to **86** at this revision, and the unexcluded divergence total to `130` against the
self-excluded `117`. Neither is a correction: the corpus grew because *this SPEC* added lines
quoting those commands, and the self-directory exclusion is the whole difference. It is recorded
because re-citing `81` as current would be this card's own subject occurring a fourth time in its
own lifecycle — after §B.4, §B.5, and the audit-window recurrence at `progress.md` §E.1 P1.

## §C. Scope

In scope: SPEC artifacts under `.moai/specs/**` (`spec.md`, `plan.md`, `acceptance.md`,
`progress.md`), the `internal/spec` lint engine, one always-loaded doctrine section, and the
template mirror of that doctrine.

## §D. Judgments

### D.1 — The exemption predicate: **is the ref the ANCHOR or the SUBJECT?**

This is the deliverable the card marks [HARD]. It is a three-test decision procedure, applied in
order, and every test is answerable by reading the sentence the ref appears in.

**Test 1 — Substitution.** Replace the ref token with the SHA it resolves to *right now*, then
re-read the sentence **as a later reader will act on it — not as you read it at the moment of
substitution.**

- It still says what it meant, *and still will when acted on later* → the ref is an **ANCHOR** (an
  address at which a measurement was taken). Anchors are pinned.
- It now says something different, narrower, or weaker — **including a meaning that is correct at
  the instant of substitution and decays afterwards** → the ref is the **SUBJECT** (the claim is
  *about* mainline as a living thing). Subjects keep the moving ref.

The evaluation time is load-bearing and is stated because omitting it broke this predicate once.
Write-time and read-time are exactly what separate an ANCHOR from an S2 claim: substituting today's
tip into "the base you will start from" reads correctly *today* and is wrong for every reader after.
A Test 1 applied at substitution time returns ANCHOR for every S2 claim there is.

**Test 2 — Falsification source.** Applies when the claim currently reads false. Attribute the flip:
were the commits that caused it authored by this SPEC's work, or not?

- Authored by this work → **true signal**. The claim is genuinely broken. Fix the work; do not
  touch the ref. Pinning here would hide a real defect, and that is the worse error of the two.
- Not authored by this work → **spurious red** from upstream drift. Remediate per §D.2.

**Test 3 — Re-measurement expectation.** Re-run this claim next week with no work done in between.
Is the same answer expected?

- Yes → anchor. Pin.
- No, *and that variance is the point of the claim* → subject. Keep the moving ref.

**Tie-break.** Tests 1 and 3 will normally agree. Where they disagree, **run Test 4** — do not
resolve to ANCHOR.

A disagreement between them is not noise to be settled by precedence; it is the **signature of an
S2 claim**. Test 1 (read at the substitution instant) says "the value fits" while Test 3 says "the
value must not be fixed", and that combination is what a live-state claim looks like from the
inside. The v0.1.0 tie-break said "Test 1 governs" and thereby routed every S2 claim to ANCHOR —
which is how instance 3 was misclassified, and why R4 was unreachable for the case that motivated
it. The correction is in the *procedure*, not only in the worked example (§D.3).

**Test 4 — Read-time action.** Runs when **Tests 1 and 3** return SUBJECT, and also whenever they
disagree (per the tie-break above). Test 2 is deliberately not part of this gate: it is conditional
— it runs only when the claim currently reads false — and it returns a *falsification attribution*
(true signal vs spurious red), never a class. A gate naming "Tests 1-3" would therefore be
unsatisfiable for a claim that reads true, which is most of them. Ask: must a later reader *act* on
this claim by measuring something?

- **No — the claim is narrative.** It describes what mainline carries, quotes a command as text, or
  records a coordinate as the subject of a correction. Nothing is measured at read time. → **S1**,
  remedied by R3.
- **Yes — the claim asserts the current state of a moving thing, and a reader will act on it.** →
  **S2**, remedied by **R4**: state the measuring command as the criterion, demote any value to a
  dated reference.

Test 4 exists because the two SUBJECT sub-shapes take structurally different remedies, and because
S2 written in anchor form is exactly how the lead's dispatch failed (§B.4, §B.5).

**Classification and remedy are two steps, not one.** The tests return a class; the class does not
name the remedy by itself — ANCHOR selects between R1 and R2, SUBJECT between R3 and R4. This is
stated because collapsing the two steps is a second, subtler route to the card's dominant failure
mode: a reader who believes the class *is* the remedy has only as many remedies as there are
classes, and reaches for the first one that fits.

#### D.1.1 — Why the predicate stays two-way at the top level

The lead asked whether S2 should be a third top-level branch. It should not, and the reason is that
the top-level tests answer one question — *is this ref an address, or the thing being talked
about?* — and S1 and S2 give the same answer to it. Separating them requires a different kind of
question (what will a reader **do** with this claim?), which is naturally a follow-on rather than a
peer; promoting it would mean the top-level branch is decided by two unrelated criteria at once,
which is how decision procedures become unusable.

It also must not fold into ANCHOR, and the reason is sharper: an anchor's entire purpose is that its
value is **fixed**, while an S2 claim's entire purpose is that its value **must not be**. They are
opposite requirements on the same property. Routing S2 to ANCHOR is precisely the error that
produced the stale base line — treating "the tip you will start from" as though it were an address.

What the addendum genuinely changes is therefore the **remedy space**, which goes from three
branches to four (§D.2), plus the routing test above. Recording this as "two classes, four remedies"
rather than "three classes" keeps the classification honest and puts the new case where a reader
will actually find it.

#### D.1.2 — Generality

The predicate is not specific to git refs. It governs any *moving coordinate* — a source line
number is the same hazard on a moving tree, and §D.3 instance 2 is that case. What is detected here
is the git-ref form only; what is *published* is the predicate.

### D.2 — Four remediation branches

The failure mode the card names is a reader who learns one branch. The finding message therefore
names all four (AC-MRG-008), and the count matters more than any single entry: with one remedy on
offer the reader pins; with four, choosing requires reading the predicate.

| | Branch | Class | When | Form |
|---|---|---|---|---|
| **R1** | Pin the literal SHA | ANCHOR | the anchor value is already known at authoring time | replace `origin/main` with the resolved 40-hex SHA, recorded with the tree and date it was resolved in |
| **R2** | Freeze at pre-flight *(recommended default for anchors)* | ANCHOR | the value is not knowable when the criterion is written — the usual case for a run-phase PRESERVE criterion | `BASELINE_SHA=$(git rev-parse origin/main)` captured before the first run-phase commit; criteria decided against `$BASELINE_SHA`, resolved value recorded in `progress.md` |
| **R3** | Keep the moving ref, declare the exemption | SUBJECT / S1 | narrative — nothing is measured at read time | leave the ref; add the inline marker with a stated reason |
| **R4** | State the measuring command; demote the value to a dated reference | SUBJECT / S2 | the claim asserts the current state of a moving thing and a reader will act on it | lead with the command that must be run at read time; any value follows it, parenthesized and dated, explicitly labelled a reference |

R2 is recommended over R1 for run-phase criteria because it removes the authoring-time knowledge
requirement that makes R1 awkward, while giving the same fixed-value guarantee. It is not novel —
§B.3 records it already in use.

**R4's ordering is load-bearing, not stylistic.** Per §B.5: a value written first reads as the
criterion and demotes re-measurement to a confirmation step. The remedy is the inversion — command
first, value second and marked as a reference — so a reader who skims the line still sees an
instruction to measure rather than a number to trust.

### D.3 — The five grounded instances, mapped onto the predicate

**Instance 1 — t292 (the card's own citation): provenance narrative.** The claim's subject is what
`origin/main` *currently carries*. Test 1: substituting a SHA converts "what mainline carries" into
"what this one commit carried", which is a different and weaker statement. → **SUBJECT → R3.**

**Instance 2 — t322, SPEC-GRAPH-FRESHNESS-CADENCE-001 v0.2.2 (landed `44095ddc2`).** Its citation
refresh deliberately left `internal/mx/provenance.go:196` and the `:208` / `:219` values
unrefreshed. The SPEC's own HISTORY row states the reason: those three coordinates are "the
*subject* of audit finding N2 rather than addresses into the tree, and refreshing them would erase
the correction N2 records." Test 1, applied to a line number rather than a ref: substituting the
current coordinate destroys the record of the miscitation. → **SUBJECT → R3.** This instance is
what establishes the predicate's generality (§D.1).

**Instance 3 — the lead's dispatch base line (§B.4).** Attributed to the lead at their request; the
attribution was verified in this turn rather than taken on the dispatcher's word. The line named
`44095ddc2` meaning "the tip of develop you will start from". Test 1: substituting the *current*
value corrects it, so the ref is not narrative — but the substitution decays immediately, because
the sentence means "whatever the tip is when you enter". Test 3: re-measuring next week gives a
different tip **and that variance is the point** — the reader is supposed to get today's tip. →
**SUBJECT / S2 → R4.**

This instance is the reason Test 4 exists. Under the v0.1.0 predicate it was classified ANCHOR → R2,
and that classification was **wrong in a way worth recording**: R2 would have had the *dispatcher*
freeze a value, which is exactly the shape that failed. The defect is not that the wrong SHA was
chosen; it is that a value was stated at all where a command belonged. The measured cost of the
v0.1.0 reading is in §B.2 — a `44095ddc2`-anchored PRESERVE check reports
`3 files changed, 288 insertions(+)` against work that does not exist.

**Trace of the v0.3.0 procedure over this instance, step by step**, because v0.2.0 corrected the
label here while leaving the tests that produced the wrong label untouched — and a reader is
entitled to check that the procedure now reaches the answer on its own:

| Step | Result |
|---|---|
| Test 1, read **as a later reader will act on it** | substituting today's tip is correct today and wrong for every reader after → decaying meaning → **SUBJECT** |
| Test 3 | re-measuring next week gives a different tip and that variance is the point → **SUBJECT** |
| Tie-break | not invoked — the tests agree. (Had Test 1 been read at substitution time it would have returned ANCHOR; the tie-break would then have routed to Test 4 rather than to ANCHOR, reaching the same place by the other road.) |
| Test 4 | a reader must measure to obtain their base → **S2** |
| Remedy | **R4** |

Both routes now terminate at R4. Under v0.1.0 neither did.

**Instance 4 (found by the §B.3 survey) — AC-COORD-016.** Asserts a literal command string is
preserved verbatim in a document. The ref token sits inside quoted subject matter, and no reader
measures anything on the strength of it. Test 4: no read-time action. → **SUBJECT / S1 → R3.**
Retained because it is the class the detector will most often flag wrongly, and it is the fixture
for AC-MRG-003.

**Instance 5 — the lead's dispatch-format change (§B.5).** The remedy of instance 3, adopted as a
standing format: `base: measure at entry with git fetch origin develop (dispatch-time reference
value: <sha>)`. It is instance 3 correctly classified — S2, remedied by R4 — and it is the only one
of the five that shows the remedy rather than the defect. It supplies the detector's R4-recognition
fixture (AC-MRG-013), so the guard does not flag the very form it recommends.

**Classification tally, and a skew worth naming.** All five grounded instances are SUBJECT-class —
S1 ×3 (instances 1, 2, 4), S2 ×2 (instances 3 and 5, the same case as defect and as remedy). **No
grounded instance is ANCHOR.** The ANCHOR class is evidenced instead by the seven corpus lines named
in §B.3 (`AC-TST-010`, `AC-AFS-012`, `AC-MSD-010`, `AC-WC-009`, `AC-CFS-010`, `AC-NS2-005`, and the
already-frozen `SPEC-DESIGN-MOAIWEBV2-002` pair).

The skew is not an accident of collection and is the strongest available argument for shipping the
predicate: the instances that got *noticed* and escalated into cards are overwhelmingly the ones
where pinning would have been wrong. Anchor-class defects are quietly correct to pin and generate no
incident; subject-class ones destroy information when pinned, which is what makes them memorable.
A guard tuned only on the noticed cases would therefore be tuned entirely on the exemption class.

Both SUBJECT sub-shapes and both of their remedies are represented, which is the minimum needed to
validate Test 4 at all. Five instances remains a small validation set (`progress.md` §E.1 residual
risk).

**The skew is also a limitation, and the doctrine must publish it as one.** The paragraph above
gives the skew's positive reading; the negative reading follows from the same fact and is not
optional. The SUBJECT branch has five adjudicated instances. **The ANCHOR branch has zero** — it
rests entirely on seven corpus lines classified by reading, none of which was independently
escalated, disputed, or adjudicated by anyone but the author. The ANCHOR branch is therefore the
**unvalidated** half of this predicate.

That is not an abstract worry. D1 — Test 1 lacking an evaluation time and thereby over-returning
ANCHOR — is exactly the failure an unvalidated branch would be expected to have, and it was found
by an auditor rather than by the author. The skew and D1 are one fact seen from two directions: the
ANCHOR branch is both the untested one and the one that demonstrably misclassified. A reader
applying this predicate should weight an ANCHOR verdict less confidently than a SUBJECT one, and
should treat a Test 1 / Test 3 disagreement as evidence against ANCHOR rather than for it.

### D.4 — The exemption is author-declared, and it costs a reason

The predicate's tests are judgments about meaning. No regex decides them, and pretending otherwise
would produce a detector that is confidently wrong on the subject class.

The exemption is therefore an inline marker the author writes **after** applying the predicate:

```
<!-- moving-ref-ok: <reason> -->
```

on the flagged line or the line immediately above it. An HTML comment is chosen so it is invisible
in rendered markdown.

**The reason is mandatory and non-empty (AC-MRG-003).** A bare marker would make "silence the
warning" cheaper than "pin the SHA", which inverts the incentive this SPEC exists to set. With a
reason required, declaring and pinning cost about the same, and the author picks on the merits.

**Pricing across all four remedies (restated for v0.3.0).** The paragraph above was computed against
a two-door choice and is no longer sufficient on its own: R4 opened a third door and arrived
carrying no reason requirement at all. An author who simply wants a line to stop being flagged
could rephrase it into R4 shape and be done — and satisfying a shape is always cheaper than writing
a justification. Left unpriced, R4 is the cheapest available silencer, which is the bulk-suppression
outcome §D.5 exists to prevent, reached by a different road.

**R4's cost is that it must name the command a reader is to run, and that command must be the one
that actually decides the claim.** This is not an added tax; it is R4's definition made binding. Its
cost is real for the same reason the marker's reason is: a wrong or vague command is visible to the
next reader who runs it, and a plausible one cannot be produced without knowing what would decide
the claim. Rephrasing to escape a warning does not survive it.

Pricing table — all four remedies carry a cost that is not merely shape-satisfying:

| Remedy | What it costs the author |
|---|---|
| R1 | resolving the SHA and recording the tree and date it was resolved in |
| R2 | capturing the baseline before the first run-phase commit, and recording the resolved value |
| R3 | writing a non-empty reason a reviewer can disagree with |
| R4 | naming the deciding command, which a later reader will run |

The **count** is what does the work here, not any single entry: with one remedy on offer the author
pins, and with four, none of which is free, choosing requires applying the predicate.

The existing per-SPEC `lint.skip` mechanism (`applylintSkip`, `internal/spec/lint.go`) remains
available but is deliberately **not** the exemption path: it silences the code for a whole document,
which is the wrong granularity for a per-claim judgment.

### D.5 — Severity is `warning`, never `error`

42 candidate lines exist today (§B.3). Shipping this at `SeverityError` would red the corpus on the
first run and the rational response would be a bulk suppression — the outcome this SPEC is designed
to prevent. `SeverityWarning` is non-blocking by default in the existing engine (`internal/spec/lint.go`,
the `Strict` branch of the exit-code computation) and escalates only under `--strict`.

**The argument above is correct and was incomplete (recorded at v0.8.0).** Its last clause names
the hole: `--strict` escalates any non-advisory warning, so choosing `SeverityWarning` did not
achieve what this section wanted. REQ-MRG-009 then went further and *required* the `--strict`
escalation, so the two clauses were written against each other — §D.5 reasoning that the corpus
must not red on day one, REQ-MRG-009 specifying the invocation under which it would. Nobody ran the
composition, and no criterion in this SPEC could have caught it: AC-MRG-005 asserted the same
`--strict` non-zero as REQ-MRG-009, so the suite agreed with the defect.

**Where it was found: `develop`'s CI, not this SPEC's own tests.** The run-phase merge `84b3b7949`
turned `develop`'s SPEC Lint job from success to failure, and it stayed red across the three
following heads. At `a6bbbf82b` the run reported `0 error(s), 155 warning(s)` → exit 1, of which
**110** are `MovingRefUnpinned`; the other 45 carry `[grandfathered era — downgraded to warning]`
and are advisory, so they do not gate. `.github/workflows/spec-lint.yml` invokes the linter with
`--strict`. The rule behaved exactly as designed; the design was wrong.

**Advisory emission completes this section's intent rather than retreating from it.** What §D.5
wanted was a finding that is seen and does not block. `SeverityWarning` delivers half of that and
`--strict` takes it back; emitting the finding advisory delivers the whole of it, under every
invocation, without touching severity. The finding stays a `warning`, stays in the report, and
stops deciding an exit code. §D.7 records why that state is provisional and what must be true
before the gate returns.

### D.6 — Disposition of the side-discipline (divergence figures)

The card asks whether "a divergence value measured once (`0 0`) must not keep being cited without
re-measurement" is a requirement, guidance, or out of scope.

**It is the same defect as §A, on a different carrier — not a neighbouring rule.** §B.4 establishes
this: the lead's stale base line is the side-discipline violated on a base SHA rather than on a
divergence count, and both are a measurement whose validity expired being re-served as current. The
remedy is correspondingly the same: **R4**, not R1. One does not pin `0 0`; one writes the
re-measuring command as the criterion and demotes `0 0` to a dated reference — which is exactly the
transformation the lead applied to the dispatch base line (§B.5).

**Split by carrier, and recorded as such:**

- **In scope as a requirement (REQ-MRG-006) for the SPEC-document carrier.** A `progress.md` or
  `plan.md` line citing a `rev-list --count --left-right` figure with no accompanying SHA or
  timestamp is the same defect on a different carrier — a measurement whose validity has expired
  being re-served as current — and it is detectable in exactly the same place.
- **Guidance only for the dispatch carrier.** A dispatch message is not a file in the tree; the
  detector cannot see it, and claiming otherwise would be an acceptance criterion that cannot fail.
  It is stated as doctrine (REQ-MRG-005) and nothing more. §B.4 is that doctrine's worked example.

The split is recorded rather than resolved silently because the two carriers look identical in
prose and are not identical to any mechanism.

### D.7 — The finding ships advisory, and that is a deferral with a stated cost

**(a) Why advisory now.** The corpus carries **110 `MovingRefUnpinned` findings** that the rule
correctly identifies — it is not over-firing, and the count is what the shape actually costs the
corpus today. A gate on day one leaves an author two choices: remediate 110 lines before anything
else can merge, or suppress in bulk. Bulk suppression is the rational one, and it is the exact
outcome §A names as this SPEC's reason for existing. Shipping the finding advisory buys the corpus
the interval in which remediation is cheaper than silencing.

**(b) The promotion condition.** The gate returns when the corpus is dealt with: every finding
triaged in `.moai/reports/t342/corpus-triage.md` either remediated by its owning SPEC through one
of the four §D.2 branches, or exempted through the R3 marker with a reason. At that point the
retracted clause returns verbatim — `MovingRefUnpinned` stops being advisory, and `--strict` exits
non-zero on it. The condition is a state of the corpus, not a date and not a judgment call: it is
readable by re-running the linter and comparing against the triage record.

**(c) The failure mode of forgetting.** A guard that declares a gate, ships advisory, and never
promotes has declared a rule and enforced nothing. Say it plainly: **that state is the same shape
this SPEC exists to detect.** REQ-MRG-009's gating clause was true when it was written, was
quietly emptied by what happened afterwards — a corpus that turned out to carry 110 violations —
and nothing in the requirement's text signals the change. A reader meets a requirement that reads
like a gate and a linter that gates on nothing, with no marker between them. In `acceptance.md`
§A's own words: **an advisory-forever guard is indistinguishable from a guard that is switched
off.** The report reads green and nothing is being detected.

This clause is the marker the retracted requirement would otherwise lack. It is not a note about
future work; it is the record that the enforcement claim is currently unbacked, placed where the
enforcement claim is read.

## §E. Requirements

- **REQ-MRG-001**: The SPEC lint engine shall emit a `MovingRefUnpinned` finding of severity
  `warning` for each line in a SPEC artifact that names a moving ref (`origin/main`,
  `origin/develop`, `origin/HEAD`) inside a git-command context, carries an invariant-claim marker,
  and carries neither a 7-40 character hexadecimal SHA nor a frozen-baseline variable reference.
- **REQ-MRG-002**: When a flagged line, or the line immediately preceding it, carries a
  `<!-- moving-ref-ok: <reason> -->` marker with a non-empty reason, the linter shall suppress the
  finding for that line.
- **REQ-MRG-003**: When such a marker carries an empty or whitespace-only reason, the linter shall
  emit a finding reporting the marker as incomplete rather than suppressing.
- **REQ-MRG-004**: The `MovingRefUnpinned` finding message shall name all four remediation branches
  of §D.2 — pin, freeze, declare, and state-the-command — and shall not present pinning as the sole
  remedy.
- **REQ-MRG-005**: The doctrine section shall state the anchor-or-subject predicate and its four
  tests (§D.1), the five grounded instances (§D.3), the four remediation branches (§D.2), and the
  seven detection limits (§F) **including L7's ANCHOR-branch validation gap**, and shall be mirrored
  into the distributed template.
- **REQ-MRG-006**: The linter shall emit a `MovingRefUnpinned` finding for a line citing a
  branch-divergence count (`rev-list --count --left-right` against a moving ref) that carries
  neither a SHA nor a date.
- **REQ-MRG-007**: While the three-dot (`A...B`) form is present, the linter shall treat it as
  equivalent to the two-dot form and shall not suppress the finding.
- **REQ-MRG-008**: While a line's moving ref carries a resolved SHA pin or a frozen-baseline
  variable reference (`$BASELINE_SHA` or equivalent), the linter shall not emit
  `MovingRefUnpinned` for that line.
- **REQ-MRG-009** *(amended at v0.8.0 — the `--strict` gating half is retracted; §D.7)*: The
  `MovingRefUnpinned` finding shall be emitted **advisory** — that is, excluded from exit-code
  gating while remaining present in the report. While only `MovingRefUnpinned` findings are
  present, `moai spec lint` shall exit 0 **both with and without `--strict`**, and the findings
  shall remain visible in the report under both invocations.

  **What was retracted, and why it is recorded rather than edited away.** The original text read
  "While only `MovingRefUnpinned` findings are present, `moai spec lint` shall exit 0; under
  `--strict` it shall exit non-zero." That second clause shipped and broke `develop`'s SPEC Lint
  job — `.github/workflows/spec-lint.yml` runs the linter with `--strict`, and the corpus carried
  110 `MovingRefUnpinned` findings on the day the rule landed. The requirement declared a gate
  without asking what the gate would do to the corpus it was pointed at.

  **The mechanism is emission-site advisory, not era demotion.** `internal/spec/lint.go` consults
  `eraDemotableCodes` only for findings already at `SeverityError`; a warning never reaches it, so
  adding the code there would change nothing. Era demotion (`applyEraDemotion`) already marks
  warnings advisory on grandfathered SPECs — the findings that gate are precisely the ones on
  modern-era (V3R6) SPECs, which no era path can demote. The rule therefore sets `Advisory: true`
  on its own findings at emission, the mechanism `StatusGitConsistency` already uses for the same
  purpose. `Report.HasErrors` escalates a warning under `--strict` only when it is **not**
  advisory, so this is sufficient and is the only available route.
- **REQ-MRG-010** *(DEFERRED out of this SPEC at v0.7.0 — §H Q0, option C, operator decision)*:
  While a line is written in the R4 form — a measuring command stated as the criterion in
  imperative structure, with any value syntactically demoted to a labelled reference — the linter
  shall not emit `MovingRefUnpinned` for that line.

  **Not implemented here.** M3 ships no R4 branch; AC-MRG-013, CM-1 and CM-2 defer with this
  requirement and were never run. The requirement text is retained verbatim as the specification a
  follow-up starts from, along with the [HARD] signature constraint (§H Q0: imperative structure,
  never a command token) and the unrun counter-mutation obligation. Early R4-form lines are
  silenced with the R3 marker in the meantime; R4 remains a doctrine remedy and stays in the
  finding message — only its lint exclusion is deferred. Grounds and resume condition: §H Q0.

  **Residual scope — nameable, and empty today (measured).** The class this exclusion could act on
  is exactly one: a REQ-MRG-001-class line — moving ref in a git-command context, carrying a claim
  marker, with no SHA and no frozen-baseline variable — written in R4 form **whose demoted reference
  value is a measurement result rather than a commit SHA**. The other two R4 shapes need no
  exclusion, and both were verified by measurement rather than reasoning (`progress.md` §E.1): an R4
  line whose reference value is a SHA is already exempt under REQ-MRG-008, and one in the divergence
  class under REQ-MRG-006's date conjunct, since a dated reference is precisely what that
  requirement asks for.

  **That class currently has no occupants: 0 of 42 candidate lines, on two independent probes
  (§B.7).** R4 is one day old, so the emptiness is expected rather than anomalous — the class is
  populated prospectively, by the R4 remediations M1's doctrine will produce. The consequence is
  stated plainly because it bears on Q0: with no occupants, this exclusion can only over-exempt
  today, never under-exempt, so every error available to it is a D11-shaped one. §H Q0 carries the
  resulting scope decision for the operator.
- **REQ-MRG-011**: The corpus triage of milestone M4 shall classify every finding without modifying
  any SPEC artifact outside this SPEC's own directory, measured against both the committed diff and
  the working tree.

## §F. Detection limits (stated, not discovered later)

Required by the card. Each limit names what the mechanism does not see, and no acceptance criterion
in this SPEC implies coverage of any of them.

- **L1 — Refs expressed without an `origin/` token are invisible.** `git diff --stat main`,
  `git diff @{u}`, `git diff HEAD~10`, and the prose form "compared against mainline" all carry the
  same hazard and none of them match. This is the lane-14 recorded Gap, carried forward unresolved.
- **L2 — The detector reads shape, never subject.** It cannot apply the §D.1 predicate. Every
  finding is a question put to a human, never a verdict, and the message says so.
- **L3 — Detection is line-scoped.** A claim whose command and whose invariant word sit on different
  lines — a wrapped table row, a fenced block with its assertion in prose above — is missed.
- **L4 — A documented command is indistinguishable from an asserted claim.** `REQ-LB-006` (§B.3)
  documents `git reset --hard origin/main` as doctrine and will be flagged. This is why the
  exemption is a marker rather than a cleverer regex.
- **L5 — No carrier outside the SPEC tree is covered.** Dispatch messages, commit messages, PR
  bodies, and reports under `.moai/reports/` carry the same defect and are not scanned (§D.6). The
  lead's dispatch base line — grounded instance 3, the occurrence that motivated R4 — sits on
  exactly such a carrier and would not have been caught by this guard.
- **L6 — A rotted reference value is indistinguishable from a live one.** *(Prospective as of
  v0.7.0: REQ-MRG-010 is deferred per §H Q0 option C, so the shipped detector carries no R4
  exclusion and this limit does not yet act on it. The limit is retained rather than withdrawn
  because it belongs to the R4 remedy, which ships as doctrine — a reader applying R4 today faces
  exactly this residual, enforced by review instead of by the detector. The doctrine's own L6
  states it without this qualifier; the qualifier is about the detector's current reach, not about
  the remedy.)* REQ-MRG-010 exempts the R4 form, and the detector cannot tell whether the
  parenthesized reference value is current or years stale; it reads the *shape* of the demotion,
  not the freshness of the number. This limit is
  **created by** the R4 exemption rather than pre-existing it, and it is accepted deliberately: the
  alternative is flagging the recommended form, which would teach readers to avoid it. The residual
  is bounded by R4's own design — the command is stated first, so a reader who follows the line
  re-measures regardless of what the stale value says.

  **L6 carries a second residual, on incentives rather than freshness.** The same shape-blindness
  that prevents a freshness judgment also means the detector cannot distinguish an author who
  applied the predicate and reached R4 from one who rephrased into R4 shape to stop being flagged.
  §D.4 prices R4 against this — the form must name the deciding command — but that price is enforced
  by **review, not by the detector**, which reads shape only. An author determined to silence a line
  can still do it; what the pricing buys is that doing so requires stating a command a later reader
  will run, which is harder to fake convincingly than an empty gesture. D11 is this gap exploited by
  an implementer writing the exclusion; this is the same gap exploited by an author writing the
  claim.
- **L7 — the ANCHOR branch of the predicate is unvalidated.** Not a limit of the detector but of the
  doctrine it enforces, and stated among the limits because a reader applying the predicate needs it
  at the same moment they need the others. The SUBJECT branch has five adjudicated instances
  (§D.3); the ANCHOR branch has **zero** — it rests on seven corpus lines classified by the author
  alone, none independently escalated or disputed. D1 of the iter-1 audit, in which Test 1
  over-returned ANCHOR for want of a stated evaluation time, is exactly the failure an unvalidated
  branch would be expected to have, and it was found by an auditor rather than by the author.
  Consequences for a reader: weight an ANCHOR verdict less confidently than a SUBJECT one, and treat
  a Test 1 / Test 3 disagreement as evidence *against* ANCHOR rather than for it (§D.1 tie-break).

## §G. Out of Scope

### Out of Scope — corpus remediation

- Bulk-pinning or bulk-marking the 42 candidate lines of §B.3. The run-phase deliverable is a
  triage record that classifies them; editing them is separate work under their owning SPECs.
- Retrofitting closed or grandfather-era SPECs.

### Out of Scope — mechanism reach

- Detecting the L1 unpinned forms that carry no `origin/` token.
- Scanning any carrier outside `.moai/specs/**` — reports, dispatches, commit messages, PR bodies.
- Automatic remediation. The guard never rewrites a claim; a fix-it mode would apply R1 blindly and
  is precisely the failure mode §D.1 exists to prevent.

### Out of Scope — adjacent concerns

- The `graph-freshness` stamp cadence (SPEC-GRAPH-FRESHNESS-CADENCE-001) and stamp reachability
  (SPEC-STAMP-REACHABILITY-001). Both involve moving anchors; neither is this defect.
- Any change to `lint.skip`, to the era-demotion table, or to existing rule severities.

## §H. Open questions for the operator

- **Q0 — R4 form recognition (the least settled, and now partly constrained).** REQ-MRG-010 exempts
  the R4 form, but the form is prose and its recognizable signature is not fixed — the lead's
  dispatch line is one instantiation, not a grammar. Too loose and the exclusion is a bypass; too
  tight and it misses legitimate variations.

  **One constraint is now settled and is not open:** the signature MUST key on **imperative
  structure** — an instruction to measure, with the value syntactically demoted — and MUST NOT key
  on any command token. A token key is forgeable by construction, and the audit demonstrated it: an
  exclusion keyed on the fetch verb passes all thirteen criteria, counter-mutation included, while
  silencing the largest real class of the defect (§B.6). What remains open is how imperative
  structure is recognized, not whether it is the key.

  **Answered in part by §B.7's measurement, which reframes the question.** R4's reachable scope is
  empty today — 0 of 42 candidate lines — so the exclusion has nothing to under-exempt and every
  error available to it is a D11-shaped over-exemption. Three options follow, and the choice is the
  operator's because it changes scope:

  | | Option | Consequence |
  |---|---|---|
  | **A** | Build the exclusion now, keyed on imperative structure | Q0's recognition problem must be solved before M3, unvalidatable against the corpus (synthetic fixtures only), and carries the full D11 bypass risk for zero measured present benefit |
  | **B** | Drop R4 from the doctrine entirely | Eliminates the problem and loses the remedy — S2 claims would have no correct form. Rejected: R4 is the right answer for live-state claims (§D.3 instances 3 and 5) |
  | **C** | **Ship R4 as doctrine, defer the exclusion; silence early R4 lines with the R3 marker** *(recommended)* | REQ-MRG-010 and AC-MRG-013 defer to a follow-up. The D11 bypass disappears because no exclusion exists to key wrongly, and Q0's unsolved shape-recognition problem is replaced by the already-solved marker problem. Cost: one marker-with-reason per R4 line until volume justifies the exclusion — and that reason is itself useful, since it records *why* the line is S2 |

  **Recommended: C**, on the grounds that it trades an unsolved and unvalidatable recognition
  problem for a solved one at a cost that scales with actual usage rather than being paid upfront.
  It is not taken here — deferring a requirement is a scope decision, and this SPEC declines to make
  it silently.

  ---

  **RESOLVED — option C, by operator decision (2026-08-28, v0.7.0).** The scope decision was put to
  the operator rather than taken by an agent, and the operator took the recommendation.

  **What is deferred, exactly:** REQ-MRG-010 (the R4-form lint exclusion), AC-MRG-013, and both of
  AC-MRG-013's counter-mutations CM-1 and CM-2. None was implemented; CM-1 and CM-2 were never run.
  M3 shipped no R4 branch in the rule, so there was nothing for those criteria to probe — and the
  DoD's counter-mutation obligation is recorded as *deferred*, travelling with the requirement,
  rather than quietly discharged.

  **Why, on the measurement rather than on preference:** §B.7 measured R4's reachable class as
  **0 of 42** candidate lines on two independent probes, and M4 re-measured it against the rule's
  own 97 external findings with the same answer — external R4 occupancy **0**. An exclusion with no
  occupants cannot under-exempt; every error available to it is therefore a D11-shaped
  over-exemption. That failure is not hypothetical: the iter-1 audit measured a fetch-verb-keyed
  exclusion passing all thirteen criteria then in force, counter-mutation included, while silencing
  **76 of 117** real unpinned divergence lines — the largest live class of this defect, shipping
  green. Deferring eliminates the bypass by construction, because no exclusion exists to key
  wrongly, and converts Q0's unsolved shape-recognition problem into the already-solved marker
  problem.

  **What is NOT deferred:** R4 itself. It remains a doctrine remedy in §D.2, keeps its §D.4 pricing,
  and stays named in the finding message (REQ-MRG-004 / AC-MRG-008) — the branch a reader cannot
  reach by intuition is exactly the one that must not vanish from the message. Only the lint
  exclusion is deferred. Meanwhile early R4-form lines are silenced with the R3 marker, whose
  mandatory reason records *why* the line is S2 — the cost option C names, paid one line at a time.

  **Resume condition — sharpened by M4 beyond what §B.7 could see.** Reconsider when the R4 form is
  actually observed in the corpus. M4's rule output found **0 external S2 lines** but **2 S2 lines
  inside this SPEC's own directory**, both R4-form: precisely what §B.7 predicted in writing that
  the class is "populated prospectively, by the R4 remediations M1's doctrine will produce". The
  class has begun to fill, and it is filling from this card's own output. When it holds live
  external occupants, the exclusion acquires something to under-exempt and the trade-off that
  decided option C inverts. The [HARD] constraint above survives the deferral unchanged: whatever
  eventually implements it MUST key on imperative structure, never on a command token.

  No follow-up card is issued from this SPEC — card issuance is the operator's act.

  ---

  **RESUMED — the option C deferral is superseded by operator card t353 (2026-09-02, v0.9.0).**
  The resume condition above was re-measured **MET**: live external R4-form occupants exist —
  `SPEC-IGNORED-EVIDENCE-CITATION-001/progress.md:480` (self-declared R4) and
  `SPEC-SPECLINT-GITBLIND-001/progress.md:233` (Korean-form), both previously silenced via R3
  markers — so the exclusion acquired something to under-exempt and the trade-off that decided
  option C inverted. Card t353 delivered REQ-MRG-010 + AC-MRG-013 + CM-1/CM-2 at commit
  `0026dc7b9` (branch `WT-r4-imperative-exempt`): the exclusion in `internal/spec/lint_movingref.go`
  keys on **imperative structure** — both conjuncts, an imperative measuring directive plus a
  syntactically demoted dated reference — and never on a command token, per the [HARD] constraint
  above, which survived the deferral unchanged and was implemented as written. **The unrun
  counter-mutation obligation the deferral carried is discharged**: RED observed pre-fix and GREEN
  after (13 MovingRef tests), CM-1's positional mutant and CM-2's command-token mutant both
  observed blocking their bypasses, and the corpus lint moved 115→113 findings with exactly the
  two genuinely-R4 lines newly exempted and zero over-exemption. Lane verdict:
  `.moai/reports/t353/verdict.md` (commit `d191d28b4`); implementation record:
  `.moai/reports/t353/impl-record.md`.
- **Q4 — the Test 1 / Test 3 undecidable region.** The v0.3.0 tie-break routes a disagreement to
  Test 4 rather than resolving it, which is correct for every case examined so far but is a
  *deferral*, not a decision procedure: it presumes a disagreement always signals S2. A claim that
  is genuinely an anchor and *also* trips Test 3 would be misrouted, and no such case has been
  found — which is not evidence that none exists, given the ANCHOR branch's validation gap (§D.3).
  Recorded as an open question in its own right; `progress.md` §E.1's "a fifth disposition may
  exist" covers coverage of the classes, not this ambiguity between two tests.
- **Q1 — Marker syntax. RESOLVED at M2 (2026-08-28): the invisible HTML-comment form
  `<!-- moving-ref-ok: <reason> -->` is adopted.** The proposal stood against a visible inline tag
  readable in the rendered document. Grounds for the resolution: the marker is an **author-to-linter
  annotation**, not content for a reader of the rendered artifact — its audience is the detector and
  the reviewer reading the source diff, both of which see the comment. A visible tag would put
  lint-suppression bookkeeping into the rendered prose of every artifact carrying a subject-class
  claim, which taxes the reader who has no use for it. The reviewer-visibility concern that kept Q1
  open is met without visibility in the render: the marker's mandatory reason is visible in the
  source diff, which is where a reviewer disagrees with it (§D.4). Recorded rather than assumed
  because the alternative was a legitimate taste call, and M2 fixes the syntax authors will type
  into artifacts — changing it later means editing every marker written in the meantime.
- **Q2 — REQ-MRG-006's scope.** The divergence-figure rule is the least certain requirement here:
  it may over-fire on progress records that legitimately narrate a sequence of measurements. It is
  specified narrowly (SHOULD-tier criterion) and can be withdrawn at run-phase without disturbing
  the rest; recorded rather than guessed.
- **Q3 — L1.** Closing L1 requires either a parser over fenced command blocks or a ref-resolution
  step, both materially larger than this card. Left explicitly open rather than implied away.
