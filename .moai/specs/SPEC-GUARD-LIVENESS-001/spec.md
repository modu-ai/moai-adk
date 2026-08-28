---
id: SPEC-GUARD-LIVENESS-001
title: "Guard firing-liveness — the surfacing model: make a guard that silently stopped reach the operator unbidden (card t333)"
version: "1.4.0"
status: draft
created: 2026-08-28
updated: 2026-08-28
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: ".claude/hooks, internal/cli, .claude/rules/moai/development"
lifecycle: spec-anchored
tags: "guard, liveness, advisory, surfacing, unprompted-discoverability, silent-absence, t333"
tier: M
era: V3R6
related_specs: [SPEC-GUARD-STATE-MODEL-001]
---

# SPEC-GUARD-LIVENESS-001 — Guard Firing-Liveness: the Surfacing Model

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-08-28 | manager-spec | Initial plan-phase authoring (card t333). |
| 0.2.0 | 2026-08-28 | manager-spec | Claim A / Claim B split; the complete third instance; the unprompted-discoverability constraint. |
| 0.3.0 | 2026-08-28 | manager-spec | t241 ledger inputs: instance 6, the one-number-two-events trap, C5/C2 scope decisions. |
| 0.4.0 | 2026-08-28 | manager-spec | §A.0 — absent execution is not suppressed failure. |
| 0.5.0 | 2026-08-28 | manager-spec | plan-audit iter-1 repair: all 7 blocking defects (D1-D7) closed. |
| 0.6.0 | 2026-08-28 | manager-spec | plan-audit iter-2 repair: N1, N2, N5 closed. |
| 1.0.0 | 2026-08-28 | manager-spec | **Scope reduction after the iter-3 FAIL + STOP (0.800 → 0.800 → 0.667).** No fourth repair round — the regression clause forbids one without an override and none was granted. The SPEC is split along the seam its own defects kept landing on. This SPEC keeps the **surfacing model**, which converged: the auditor re-ran both prior N1 mutants and could not revive either. The **state model** (former REQ-GDL-001..010 + 012, carrying D2, D5, N2, T2, T4 as one never-converging family) moves to `SPEC-GUARD-STATE-MODEL-001`. Seam resolved as **contract, not dependency** (§B.1). Requirements renumbered contiguously — mapping in §B.2. **T9 closed** — the unconditional-invocation clause is promoted to its own requirement (REQ-GDL-003) and gains AC-GDL-003, the criterion the audit named the single most consequential missing one. **T3 dissolved structurally** rather than restated: the trigger is now expressed in contract terms, so no criterion can encode a classification list. Optional T7, T10, T11 folded into §D.3; N4 taken in AC-GDL-006. Counts 16→9 REQ, 16→9 AC. |
| 1.1.0 | 2026-08-28 | manager-spec | **§A.9 added — instance 7, measured on `091966c55`.** Every completed `ci.yml` run on `develop` in the observed window failed (5/5; two cancelled, one in progress), failing jobs `Test (ubuntu-latest)` and `Race Test` on run `33117635467`; the cause is recorded as the lead's attribution (a doctor check structurally red in CI's bin-absent environment since `812ee01fc`, card t346) and not this lane's measurement. The instance is the roles reversed: instances 1-6 are absence read as success, while this is **a verdict that was produced, was correct, was red, and reached no one** — the orchestrator on card t333 named `origin/develop` CI as the judge when reporting t298's landing and never returned to read the verdict. It is the live case for §A.8, which was hypothetical until this card produced one, and it is the strongest available evidence that the discipline is not sufficient without a mechanism, because it happened to the party that spent the session policing this exact failure in others. The sibling state-model SPEC's card id `t347` is filled throughout. No requirement or criterion changed — counts unchanged at 9 REQ / 9 AC. |
| 1.2.0 | 2026-08-28 | manager-spec | **Card t326's landed session-start advisory disposed of as IN SCOPE, composing (§B.3) — not excluded.** §A.10 added: the t326 citations are pinned to `origin/develop` at `ec15ec2cd`, a tree **diverged** from this SPEC's baseline (diverged, `merge-base --is-ancestor` false, `merge-base --is-ancestor` false), where reading the baseline for a t326 surface reports a landed feature as absent. **REQ-GDL-010** requires joining the existing session-start additional-context block rather than opening a second surface — that block is already a multi-contributor surface and t326's helper records itself as the fifth joiner. **REQ-GDL-011** records a constraint the design did not previously state and which the read discovered: the host surface is latency-bounded (t326 bounds its comparable path at 250 ms and affords an inline comparison only because it is two short local `git` calls), while this evaluator issues one forge query per subject — so **render and refresh are separate acts** and the advisory reads a persisted result. That split was already implied by REQ-GDL-006's persisted timestamp and never stated. The divergence from t326's silence policy (it speaks on one outcome; REQ-GDL-004 speaks on any non-clean) is deliberate and its cost is recorded in §D.3. AC-GDL-010 and AC-GDL-011 adopt both. Counts 9→11 REQ, 9→11 AC. |
| 1.3.0 | 2026-08-28 | manager-spec | **plan-audit iter-4 repair (FAIL 0.75 vs threshold 0.80; six blocking, both optional taken).** **D1** — the seam contract was minimal enough to pass its falsification test and too minimal to specify against: it asserted one value means *nothing to report* and gave the consumer no way to identify which, leaving only a forbidden route (hardcode the literal, violating AC-GDL-001(b)) and a wrong one (trigger on the surface fold, which under-fires because more than one classification folds to the clean surface value). REQ-GDL-001 gains clauses (iii)/(iv) requiring a carried machine-readable clean-value designator; **the producing half landed in the same commit** as `SPEC-GUARD-STATE-MODEL-001` REQ-GSM-012, because a seam repaired on one side is how the two SPECs drift. **D2** — REQ-GDL-003 bound "the evaluator" to activation while REQ-GDL-011 split render from refresh and §D.3 called the refresh trigger undecided; all three could not hold. Resolved by deciding it: REQ-GDL-003 now binds **both** acts to activation, and new REQ-GDL-012 states how that survives the latency bound — the refresh is initiated and **never awaited**, its result persisted for a later activation. §D.1's entailment is restated against the refresh, and the price (an advisory can be one activation stale) is recorded rather than hidden. **D3** — §E over-swung: t326's *code* is landed on `origin/develop` at `ec15ec2cd`, its *SPEC* reads `status: in-progress`; §A.1 and §E now state both halves, and composition-target instability is a recorded risk. **D4** — §B.2's mapping table now version-qualifies every left-column id. **D5** — AC-GDL-010(a) was satisfied by the host runtime, which concatenates every contributor's context unconditionally; both clauses restated against artefacts the deliverable controls, (b) as a handler-file count against a measured baseline of 7. **D6** — REQ-GDL-011 now requires a declared join bound capped at 250 ms, since no shared bound exists and an undeclared one is unbounded. **T9's second layer** — AC-GDL-003 measured invocations and read the call site, so the "invoked unconditionally, evaluates conditionally" mutant survived; clauses now count at the **query layer** and require two opposite-diff fixtures to agree. **D7** and **D8** taken. Counts 11→12 REQ, 11→12 AC. |
| 1.4.0 | 2026-08-28 | manager-spec | **plan-audit iter-2 repair (PASS-WITH-DEBT 0.8625, up from 0.75; fit to implement after three localized repairs, none touching the design or the milestone map).** **D9** — AC-GDL-010(b)'s cited baseline of 7 reproduced under no definition. Re-derived from the stated one (*a type whose `EventType()` returns `EventSessionStart`*) on `origin/develop` at `ec15ec2cd`: **4 handlers**, in `auto_update.go`, `handoff_inject.go`, `session_start.go`, `session_start_compact.go`, with the command and its output quoted. The 7 came from counting files *mentioning* the token — including the dispatch, the enum, and the binary-lag file, **which registers no handler**. The definition was not adjusted to preserve the figure. The same measurement confirmed the composition pattern the criterion mandates: `session_start.go:479` calls the shared helper from inside the existing handler. **D10** — `acceptance.md` §D.7 still carried the "refresh trigger is undecided … no criterion covers it" bullet that D2 had falsified in every clause, in the register a reader consults to learn what is open. Replaced with its correct successor, and the replacement is annotated rather than silently swapped. **D11** — REQ-GDL-001(iv) required reading the designation and REQ-GDL-004 fired on non-clean, and neither said what happens when the designation is **absent, null, or multi-valued**: the partition has no referent, nothing is identified as non-clean, and nothing renders — §A.0 at the consumer's layer. New REQ-GDL-013 and AC-GDL-013 refuse an all-clear on a contract-violating result, mirroring the producing SPEC's own zero-count guard; the producing side gained the matching clause in the same commit. **Twin sweep, prompted by D10 being the second such case:** two further stale twins found by sweeping rather than by audit — the composition-target risk was in `spec.md` §D.3 and missing from `acceptance.md` §D.7, and the sibling SPEC's reuse-verdict count had gone stale under its own row-8 addition. Counts 12→13 REQ, 12→13 AC. |

## §A Context and Problem

A guard fired when it was built, then something changed — deployment, trigger, branch model — and it silently stopped, and nothing announced the stop. Because it is an **absence** rather than a failure, it reads as green.

### A.0 Absent execution is not suppressed failure

Two shapes are easy to blur, and this card is entirely the second. Getting them the wrong way round misdescribes every instance below.

- **Suppressed failure** — something ran, went red, and the red did not reach anyone. This is the landed rule's §1.2 **(c)**, and it is *not* this card's subject (§B).
- **Absent execution** — nothing ran, so nothing could go red, and the absence is reported as success by a mechanism answering a narrower question than the reader thinks it is.

The distinction rules out the obvious framing. It is wrong to say the check *should have failed and did not*: there was no failure to suppress. The accurate sentence is **"nothing failed, and there was nothing there to fail."**

What every instance shares is a mechanism correctly answering a question about a set that had silently become the wrong set. An exit code of zero means *everything selected passed*; it does not mean *everything that should have passed, passed*. Nothing in the mechanism compares the selected set against the intended one. **Green is accurate and uninformative at the same time.**

### A.1 The axis this card owns, and the axes it does not

This card owns the **event-history axis** — *when did this guard last fire, and was it supposed to have fired by now?* — and, after the scope reduction, specifically its **surfacing half**: once that question has an answer, how the answer reaches a human who did not know to ask.

Two neighbouring axes are owned elsewhere:

- The **binary-state axis** — whether the installed binary's build commit is a strict ancestor of the tree HEAD — is card t326's (`SPEC-BINARY-LAG-VISIBILITY-001`). Its **implementation surfaces are landed** on `origin/develop` at `ec15ec2cd`; its **SPEC is still open** (`status: in-progress`, measured there). Both halves matter and §E states them together: the code exists to compose onto, and the card that owns it can still change it.
- The **classification model** — which classifications exist, what each means, how an entry is decided into one, and everything that produces an evaluation result — is `SPEC-GUARD-STATE-MODEL-001` (§B.1). This SPEC consumes a three-clause contract from it and defines none of it.

The boundary against t326 is **state vs history**, stated most sharply by the t326 lane:

> t326's yardstick measures the `pull_request`-triggered guards as perfectly current: the files sit in the tree and the binary is not behind. The fact that they do not fire does not register on that yardstick.

### A.2 Instance 1 — deployment axis (t298, 2026-08-27)

`SPEC-INTEGRATION-LOCK-LIVENESS-001` fixed integration-lock liveness and landed, but the installed `~/go/bin/moai` was built from a commit roughly six hours older, so the fix never ran. Three lanes each independently observed the `reclaimable` misread; one real eviction happened without `--force`; the lead issued a card for a defect that was already fixed. Nothing in any signal said "deployment lag". It was resolved only when one lane ran an isolated A/B of the old and new binaries side by side.

### A.3 Instance 2 — trigger axis (git-flow transition, 2026-08-27)

The git-flow transition removed card PRs, so two guards whose trigger was `pull_request` stopped running on `develop` entirely. Card t314 rewired both — `spec-lint.yml` and `docs-i18n-check.yml` gained `push` on `develop` — and closed with the first firing left as a **pending opportunistic observation**; the operator's recorded decision was literally "go with opportunistic observation".

Full evidence, with commands and verbatim output, is at `.moai/reports/t333/trigger-axis-observation.md`. **The two observations are not of equal strength, and the difference is itself this card's subject — they are stated separately and must not be folded into one sentence.**

- **Claim A — `spec-lint`.** Fired on `develop` push, success, on three consecutive pushes. Visible in an **unfiltered** `gh run list --branch develop`. Independently reproduced by the lead session from its own query.
- **Claim B — `docs-i18n-check`.** Fired once on `develop` push, success. **Not visible in the unfiltered listing.** Recovering it required a query targeted at that workflow by name.

The load-bearing point is not that they fired. It is that **they fired, they were correct, and none of that reached anyone**. A firing and a non-firing were indistinguishable from outside until someone went looking. Had they *not* fired, the same plan would have produced the same silence.

### A.4 Instance 3 — checking is not sufficient, because the default view is lossy from inside

Produced by measuring the other two, which is what makes it the strongest.

`docs-i18n-check`'s absence from the default listing has at least two causes, and the listing does not separate them: its `paths:` filter did not match this round, or its trigger is broken and it cannot fire at all. Both render as *not in the list*. The path-filter reading is an inference from opening the workflow file — and **a coherent explanation is not an observation**.

The lead session, querying the default way, reproduced Claim A and could not reproduce Claim B. Two competent readers of the same repository, minutes apart, held different pictures of which guards were alive — not because either measured badly, but because the surface does not carry the answer.

**Then the lead did reproduce Claim B — but only after this lane handed it the query.** It ran the targeted query, got matching output, and recorded why it could: the workflow's name had reached it in a message from this lane. It did not suspect the answer; it was handed the question.

That is the complete form of the instance. The gap did not close because a second observer looked harder. It closed because one observer handed the other the question. Absent that hand-off the lead's picture would have stayed as it was — and **nothing anywhere would have told it the picture was wrong. There is no signal for "your view of which guards are alive is incomplete."** The incompleteness is silent by exactly the mechanism the non-firing is.

This instance defines the remaining scope. REQ-GDL-005 requires the answer to arrive **unprompted**: a targeted query can only be issued by someone who already suspects the answer, so a design answering only when queried has not solved this problem — it has relocated it into whoever is expected to already know.

### A.5 Instance 4 — fires, does nothing, reads green

`spec-status-auto-sync.yml` is triggered by `pull_request: types: [closed]` only. Under git-flow, card PRs are abolished, so its remaining trigger source is release PR merges. Its three most recent runs are all `conclusion: skipped` — the workflow's own `if: merged == true` gate declining. A liveness check asking only "did it fire?" reads these as green.

The vocabulary that separates *fired* from *fired with effect* is `SPEC-GUARD-STATE-MODEL-001`'s. The instance is recorded here because it grounds the problem, not because this SPEC decides it.

### A.6 Instance 5 — the census gap

`.github/workflows/` holds **18** workflow files. Grouping the 100 most recent runs by workflow name yields **11** distinct names; seven files never appear. Some are legitimately release-only, but the listing does not say which.

Two premises the subtraction rests on, stated rather than left implicit: the file-to-`name:` mapping is bijective on this tree (measured on `091966c55`: 18 files, 18 distinct `name:` values); and the window is saturated and retention-bounded — the 100-run listing spanned about three hours because a handful of high-frequency workflows fill it. The specific window is not re-measurable and drifts on every run; the durable finding is the *shape*.

### A.7 Instance 6 — a selector that matched nothing, inside the landed rule's own scope

`.claude/rules/moai/development/verification-completeness.md` landed at `7f5b6a947` — author and committer date both `Tue Aug 25 13:05:04 2026 +0900`, verified with `git show -s`. Two days later its own named defect recurred in `.moai/specs/**/progress.md`, **squarely inside the rule's `paths:` scope**. A `-run` selector named three tests; only one existed under that name. The run printed `ok ... 0.249s`, and nobody saw it until sync close.

The value of the instance is not that two checks were missing. It is that **their absence never reached the exit code.** That green was **true about what was selected**. **Nothing failed, and there was nothing there to fail** (§A.0).

This instance is on a third axis — not deployment, not trigger, but a **selector matching nothing** — which is what makes it load-bearing for scope. **Any check whose non-execution is indistinguishable from its success has this defect.**

### A.8 The always-red variant (recorded, not this card's subject)

`Graph Freshness` failed on every `develop` push in the measured window (3/3). Repairing it is card t322's subject.

It is recorded because it is a second route to the same end state, and it bears directly on this SPEC. A guard red on every single run stops being read just as thoroughly as one that never runs. **Silence and constant noise are different mechanisms arriving at the same place — nobody looks.** The consequence is a design obligation: **making a channel and making a channel that gets read are different jobs**, and REQ-GDL-007 is where this SPEC pays it.

§A.9 is the live case for this section. It was a hypothetical until this card produced one.

### A.9 Instance 7 — a verdict was produced, was correct, was red, and reached no one

The roles are reversed here. Instances 1 through 6 are all *absence read as success*. This one is a **verdict that existed, was correct, and was never collected** — and it was produced by the party running this card.

Measured on `091966c55`:

```
$ gh run list --branch develop --workflow ci.yml --limit 8 \
    --json createdAt,conclusion,headSha,status
2026-08-27T22:49:25Z  in_progress    44095ddc2
2026-08-27T21:19:51Z  completed  failure    4fdbd55c1
2026-08-27T21:16:38Z  completed  cancelled  8806a8788
2026-08-27T18:28:00Z  completed  failure    8da086fbd
2026-08-27T18:22:22Z  completed  cancelled  9a1831efd
2026-08-27T15:24:50Z  completed  failure    f5a834fef
2026-08-27T15:05:15Z  completed  failure    da03d9188
2026-08-27T14:51:23Z  completed  failure    d34a789a4
```

**Every completed run in the window failed** — five of five; the two `cancelled` runs are superseded pushes rather than counterexamples, and one run was still in progress at measurement. The failing jobs on the most recent completed run (`33117635467`) are `Test (ubuntu-latest)` and `Race Test`.

The **cause** is the lead's attribution and not this lane's measurement: a doctor check that always fails in CI's bin-absent environment, structurally red since `812ee01fc`, now card t346 with another lane. What this lane measured is the redness and its span, not why.

**What happened, stated plainly.** When the orchestrator on card t333 reported that t298 had landed, it wrote that the full suite was not run locally because `origin/develop` CI is the judge. It named CI as the authority, and then never went back to read its verdict — it read the run listing while CI was still in progress, saw an empty conclusion, and moved on. The verdict arrived. Nobody collected it.

Note the last row: `d34a789a4` is this SPEC's own original RED-now baseline, the t298 integration push. The failing run is the one this card's own measurements were taken against.

**Why this is the same defect and not a different one.** Deferring to an authority and never reading its answer is, from the outside, **indistinguishable from the authority never having answered**. The green that follows is not false — no claim was made about the verdict — it is simply uninformative, which is §A.0's shape exactly: a mechanism (here, a human protocol) correctly answering a narrower question than the reader believes it answered. "CI is the judge" is true; "CI judged, and we read it" was never established.

**And it composes with §A.8.** Part of why the return trip did not happen is that a red CI row on `develop` has been unremarkable for a day: the channel carries a standing red, so a new red in it carries no information. That is §A.8's mechanism with a first-person instance behind it rather than a hypothetical neighbour.

**The reason it belongs in the SPEC rather than in a lessons file** is what it demonstrates about discipline versus mechanism. It happened to the party that had spent the entire session policing this exact failure mode in others — auditing unobserved claims, requiring measured baselines, rejecting a green whose swept set was empty. A discipline held that attentively still did not survive one deferral. **That is the strongest available evidence that the discipline is not sufficient without a mechanism**, which is the premise of this whole card: the answer must arrive unbidden (REQ-GDL-005), because the alternative is a competent party intending to go back and look, and not going back.

## §A.10 Tree attribution for the t326 citations

[Every citation in §B.3 and §E of `internal/binlag/`, `internal/hook/session_start_binary_lag.go`, `internal/cli/doctor.go` and `internal/cli/uikit/types.go` is measured on **`origin/develop` at `ec15ec2cd`**, not on this SPEC's baseline tree.]

The two trees are **diverged**, not merely offset. **The durable claim is the divergence, not the count** — the right-hand number advances with every commit on this branch, so a pinned count falsifies itself within the session that wrote it. Measured at `9c543b99a`:

```
$ git merge-base --is-ancestor HEAD origin/develop && echo ancestor || echo "NOT ancestor"
NOT ancestor
$ git rev-list --count --left-right origin/develop...HEAD
67	10
```

An earlier draft of this block labelled its figures "measured on `091966c55`" while citing `37263c222` — a commit that did not exist at the tree it named, and whose count (`67 9`) does not reproduce at `091966c55` (`67 7`). **That is the misattribution class this section exists to prevent, occurring inside it**, and it is recorded rather than quietly corrected: the conclusion was unaffected, which is exactly what makes the error easy to carry forward. The block is now labelled with the tree it was run on, and readers should treat the left-hand count as the stable figure and the right-hand one as valid only at the named SHA.

`origin/develop` carries 67 commits this branch does not, including all of t326's landed work; this branch carries 9 it does not, all of them this card's SPEC authoring. **Reading this tree for a t326 surface returns "No such file or directory" for a feature that exists** — a stale-tree reading that reports a landed feature as absent, which is the D7 class this SPEC already paid for once.

Every RED-now cell in `acceptance.md` remains pinned to `091966c55`, because those measure *this* deliverable's absence. The t326 citations are pinned separately to `ec15ec2cd` and labelled at each use. Mixing the two pins silently is the failure this section exists to prevent.

## §B Relationship to the landed rule, and to the state-model SPEC

`.claude/rules/moai/development/verification-completeness.md` landed at `7f5b6a947` and already carries the observed-failure discipline, the three-part check spec, the two-cell adoption pair, and the mutant probe. **None of it is re-authored here.**

The extension point is measured, and it is one line — the `(a) WHEN` clause of §1.2, which speaks about **authoring time**: a check scheduled at a structurally always-green moment. Nothing in the rule speaks about a check that was correctly scheduled and later stopped. **The rule watches a check being born; this card watches whether it stays alive.**

Boundary against the rule's `(c)`: `(c)` covers a **suppressed failure**. This card covers an **absent execution**. `(c)` is a reachability clause, and reachability presupposes something to reach; when the execution is absent there is no signal to route, no exit code to surface, and no log level to raise. That is what makes the extension additive rather than a louder restatement of an existing clause.

### B.1 The seam — a consumed contract, not a dependency

This SPEC's trigger is defined over classifications, which are `SPEC-GUARD-STATE-MODEL-001`'s vocabulary. That is a real coupling, and papering over it would convert a scope reduction into a blocked card.

**The seam is resolved as a contract, not as `depends_on`.** REQ-GDL-001 declares the whole of what this SPEC consumes:

> every entry in an evaluation result carries exactly one classification; exactly one value in that vocabulary means *nothing to report* (the **clean** value); the advisory fires on any other.

That is the entire contract. This SPEC does **not** name the vocabulary, state how many values it has, name any of them, or decide any entry into any of them.

**Why a contract rather than a dependency.** Under `depends_on`, this SPEC could not enter run phase until the state SPEC closed — a deferral wearing a reduction's clothes. Under the contract the two SPECs are independently implementable against a three-clause interface, which is what makes the split a genuine reduction. The falsifiable test of whether the seam is real is whether this SPEC ever restates the values: it does not, and a grep across this artifact set for any classification name returns only the contract's abstract terms.

**The seam also buys a correctness property, which is the strongest evidence it is genuine.** Three iterations of defects — D1, N1, T3 — were one shape: the trigger enumerated **symptoms**, first two named classifications, then those plus a coverage proxy, and each time a run reached the same silence down an unenumerated branch. Under the contract the trigger cannot enumerate, because there is nothing to enumerate: there is the clean value, and there is everything else. **The defect family that failed this SPEC three times is not merely fixed here — it is unrepresentable here.** T3 is dissolved by that restructuring rather than restated as two corrected sentences.

### B.3 The landed session-start advisory — in scope, and it constrains the design

An unprompted-discoverability path **already exists and is running**. Card t326 landed not only a doctor item but a session-start advisory, measured on `origin/develop` at `ec15ec2cd` (§A.10):

```
$ git ls-tree --full-tree -r --name-only origin/develop | grep -i binlag
internal/binlag/binlag.go
internal/hook/session_start_binary_lag.go
…
```

§D.2 argues for a mechanism that reaches a reader without being asked. One exists, so this SPEC owes an answer rather than silence. **The disposition is: in scope, composing with it — not a second surface.**

**Why composition rather than a parallel channel.** The session-start advisory block is already a multi-contributor surface by construction. `session_start_binary_lag.go` carries a helper for exactly this, and its own comment records that it is the fifth joiner:

> `appendAdditionalContext` adds text to the session-start additionalContext, creating the hook-specific output block when this is the first contributor and separating from earlier contributors otherwise. Callers upstream open-code this same shape; the lag advisory is the last one to join, so it uses a helper rather than adding a fifth copy.

A second surface would split the reader's attention across two channels for one concern, and §A.8's mechanism says a second channel is exactly how both end up filtered. REQ-GDL-010 records the composition obligation.

**The constraint this discovery imposes, which the design did not previously state.** The session-start path is latency-bounded:

```
$ git show origin/develop:internal/hook/session_start_binary_lag.go | grep -n "binaryLagJoinBound ="
const binaryLagJoinBound = 250 * time.Millisecond
```

t326 can afford an inline comparison because it is two short **local** `git` invocations. This SPEC's evaluator issues **one forge query per subject** — 18 on this repository — and no sequence of network round-trips fits inside a 250 ms join bound. **The render and the refresh are therefore separate acts**, and the SPEC now says so: the advisory reads a *persisted* result at the host surface and performs no forge query inline (REQ-GDL-011).

This was already implied and never stated — REQ-GDL-006 requires the reported age to be derived from a persisted result, which presupposes that something persisted it at another time. Making the split explicit is what stops an implementer from wiring the evaluator's queries directly into session start and discovering the bound at run time.

**What this SPEC deliberately does not adopt from t326's advisory.** Its silence policy is the opposite of REQ-GDL-004's trigger, and the divergence is intentional:

> Empty covers every non-lag outcome, and they are not distinguished here on purpose: a session start is not a diagnostic report, so the only outcome worth interrupting a reader for is the one where the binary really is running code the tree has moved past.

t326 speaks on **one** outcome and is silent on all others. REQ-GDL-004 fires on **any** non-clean entry. Both are defensible for their own subject — t326 compares one binary and can name the single interesting verdict, while this SPEC sweeps N subjects where "could not determine" is itself the finding (§A.0). The cost of the divergence is real and is recorded in §D.3: composing a speak-on-any-non-clean advisory into a block whose established policy is speak-rarely changes that block's noise profile, and §A.8 is precisely about what a noisier channel does to its readers.

### B.2 Requirement renumbering across the split

Requirements were renumbered contiguously, because a non-contiguous set fails the numbering must-pass. The mapping from v0.6.0 is recorded once, here:

**Every left-column cell is a `v0.6.0` id and every right-column cell is a current id.** The two id spaces overlap numerically and are not the same tokens: `v0.6.0 REQ-GDL-011` and the current `REQ-GDL-011` are different requirements. The version qualifier is written into each cell rather than left to the column header, because a reader who lands on one row must be able to tell which space it is in without carrying the header.

| Pre-split id (`v0.6.0`) | Current id | Disposition |
|---|---|---|
| `v0.6.0` REQ-GDL-001..010, 012 | — | Moved to `SPEC-GUARD-STATE-MODEL-001` (REQ-GSM-*) |
| `v0.6.0` REQ-GDL-011 (pull-based/attended clause) | REQ-GDL-002 | Split from its second clause |
| `v0.6.0` REQ-GDL-011 (unconditional-invocation clause) | REQ-GDL-003 | **Promoted to its own requirement.** T9 found it verified by nothing; a clause sharing a requirement with another is a clause a criterion can silently skip |
| `v0.6.0` REQ-GDL-013 (trigger) | REQ-GDL-004 | Restated in contract terms (§B.1) |
| `v0.6.0` REQ-GDL-013 (no-operator-input clause) | REQ-GDL-005 | Split out for the same reason as REQ-GDL-003 |
| `v0.6.0` REQ-GDL-014 (age) | REQ-GDL-006 | |
| `v0.6.0` REQ-GDL-015 (change-leading) | REQ-GDL-007 | |
| `v0.6.0` REQ-GDL-016 (doctrine) | REQ-GDL-009 | |
| — | REQ-GDL-001 | New: the consumed classification contract (§B.1), extended at v1.3.0 with the clean-value designator |
| — | REQ-GDL-008 | New: the advisory path mutates nothing (`v0.6.0` REQ-GDL-012 covered this and moved with the evaluator) |
| — | REQ-GDL-010, REQ-GDL-011 | New at v1.2.0: composition with the landed session-start path, and the persisted-read/declared-bound obligation (§B.3) |
| — | REQ-GDL-012 | New at v1.3.0: the refresh is initiated on activation and never awaited (§D.1) |

## §C Requirements (GEARS)

Budget: Tier M ≤ 16 requirements. **Count: 13.**

### C.1 The consumed contract

- **REQ-GDL-001** — The advisory shall consume evaluation results in which (i) every entry carries exactly one classification, (ii) exactly one value of that vocabulary denotes *nothing to report* (the **clean** value), (iii) **the result carries a machine-readable designation of which value that is** — a result-level clean-value designator, or an equivalent per-entry cleanliness flag — and (iv) the advisory identifies clean entries by reading that designation and by no other means. This SPEC shall not define the vocabulary, enumerate its values, or decide any entry into any of them. A change to the number or meaning of classifications shall require no change to this SPEC.

  Clause (iii) is what makes the contract *specifiable against* rather than merely minimal. Without it a conforming consumer has exactly two routes and both are wrong: hardcoding the literal violates AC-GDL-001(b) and AC-GDL-002(b), which require a value-token grep over this deliverable's source to return rc=1; and triggering on the surface fold **under-fires**, because the producing SPEC folds more than one classification to its clean surface value while only one classification is the clean value. A designator is not a re-enumeration — it carries *which*, not *what the set is*, so the seam stays closed under a vocabulary change.

### C.2 The advisory

- **REQ-GDL-002** — The evaluator shall be pull-based and invoked from an already-attended surface. It shall not be implemented as a scheduled workflow, because a scheduled watcher is itself subject to the defect it watches for and starts an unbounded regress.
- **REQ-GDL-003** — When the host surface activates, **both** acts shall be initiated — the **render** (synchronous, reading the persisted result) and the **refresh** (asynchronous, producing the next result) — each **unconditionally on that activation, with no path filter, changed-file test, or subject-matter condition gating either**. The prohibition binds the whole path, not the call site: an evaluator invoked on every activation that then returns early on a subject-matter test is the same defect one frame inward. A condition that stops matching is how §A.3's guard went quiet without being removed.
- **REQ-GDL-012** — The refresh shall not block the render or the host surface: it is initiated on activation and its result is persisted for a **later** activation to read. Its completion within any particular activation is not required and shall not be waited on.
- **REQ-GDL-004** — When an evaluation result carries **any entry whose classification is not the clean value**, the harness shall render a non-blocking advisory to the operator. The condition is the clean/non-clean partition and nothing else; it shall not be restated as a list of particular classifications (§B.1).
- **REQ-GDL-013** — Where the consumed result **violates the contract of REQ-GDL-001** — the clean-value designation is absent, null, or designates more than one value — the harness shall render an advisory reporting the contract violation, and shall **not** report an all-clear. On such a result the clean/non-clean partition has no referent, so REQ-GDL-004's condition cannot be evaluated; the path of least resistance is to identify nothing as non-clean and therefore render nothing, which is green about a set the mechanism never learned (§A.0) at the consumer's layer. This mirrors the producing SPEC's own guard, which refuses an all-clear while its queried count is zero or its enumeration returned nothing, for the same reason and in the same shape.
- **REQ-GDL-005** — The advisory shall arrive **without the operator issuing any query and without the operator supplying any guard identifier**. A liveness verdict that answers only when queried has relocated the defect into whoever is expected to already know the question (§A.4).
- **REQ-GDL-006** — The advisory shall state the age of the measurement it reports, **derived from the persisted result's own recorded timestamp**, so a stale advisory declares its own staleness rather than reading as a current all-clear.
- **REQ-GDL-007** — The advisory shall lead with the entries whose classification **changed** since the previously rendered result, and shall carry unchanged non-clean entries as a compact standing count rather than as a re-rendered list. A channel that reprints an identical block every session trains the filter that removes it, and a new advisory rendered beside a permanently-red neighbour inherits that filter (§A.8).
- **REQ-GDL-008** — The advisory path shall not write to the repository working tree, commit, push, open an issue, or mutate any forge state. The result persistence REQ-GDL-007 requires shall live outside the working tree.

### C.3 Composition with the landed session-start path

- **REQ-GDL-010** — The advisory shall join the **existing** session-start additional-context block through that surface's established contributor mechanism, and shall not create a second advisory surface for the same concern. A second channel splits one concern across two surfaces, which is how both acquire a reader's filter (§A.8, §B.3).
- **REQ-GDL-011** — When rendering at the host surface, the advisory shall read a **persisted** evaluation result and shall perform no forge query inline. The render is a separate act from the refresh (REQ-GDL-012) because the host surface is latency-bounded and one forge query per subject cannot fit inside such a bound. **The deliverable shall declare its own render join bound, and that bound shall not exceed 250 ms** — the value the comparable landed contributor on this surface uses (§B.3). Each contributor on a shared surface carries its own bound, so an undeclared bound is an unbounded one.

### C.4 Doctrine

- **REQ-GDL-009** — `.claude/rules/moai/development/verification-completeness.md` shall gain an additive continued-firing clause stating that a check's completion does not survive a change to its trigger, its deployment, or its branch model. No existing text in that file shall be modified.

## §D How the design answers the two binding constraints

### D.1 Constraint 1 — self-observation

*Build a periodic check and that check becomes subject to this very card.*

The design answers it by **not being periodic**. The evaluator is pull-based (REQ-GDL-002) and runs at a moment that already happens for other reasons, so its firing is *entailed* by someone working rather than scheduled independently of them. A scheduled watcher would have a cadence of its own to miss, and the forge additionally disables scheduled workflows after a period of repository inactivity.

**The entailment is load-bearing, and it binds the refresh, not only the render.** "Entailed by someone working" is true only if the act that *produces* a verdict is unconditional on the host's activation. A render bound to attendance over a refresh triggered by something else would leave the entailment resting on nothing — the reader would reliably see a result, and nothing would reliably produce one.

REQ-GDL-003 therefore binds **both** acts to activation, and REQ-GDL-012 states how that is possible despite the latency bound: the refresh is initiated on every activation and persisted for a **later** one to read. Its completion is never awaited, so it cannot block the render, and the reader always sees the most recent completed refresh with its age disclosed (REQ-GDL-006). This is the pattern the comparable landed contributor already uses on this surface — a bounded background goroutine whose overrun work is abandoned rather than waited on (§B.3, measured on `origin/develop` at `ec15ec2cd`).

The consequence is stated rather than hidden: **an advisory can be one activation stale by construction.** That is the price of keeping the refresh unconditional under a latency bound, and REQ-GDL-006's age disclosure is what keeps it legible instead of silent. The audit found it previously verified by nothing, with the surviving mutant being an evaluator gated on whether the session's diff touched `.github/workflows/` — which is `docs-i18n-check.yml` rebuilt inside the deliverable: §A.3's own guard, wearing this card's name.

### D.2 Constraint 2 — unprompted discoverability

*A targeted query can only be issued by someone who already suspects the answer* (§A.4).

Three questions, tested rather than assumed:

**Does it surface without being asked?** It surfaces. REQ-GDL-005 makes the advisory an output of an already-attended surface, not a verb the operator invokes — and the operator supplies no guard name and no query, which is precisely the input the lead session did not have and could not have produced.

**If it answers only when queried, what tells a reader a query is owed?** Not applicable by construction, which is the point: this is the question the current state fails, and every candidate keeping a query in the loop inherits the failure. It is why a documented `moai guard liveness` verb was rejected as a sole answer — a documented verb is still a question someone must know to ask.

**Is the announcer itself discoverable by someone who does not know it exists?** Yes, in the only available sense: it is not a thing to be discovered. It arrives unbidden, so the reader needs no prior knowledge of the mechanism.

### D.3 What the design does not close

- **The host surface can be removed.** The evaluator stops with it — the same defect class, one layer up. Full closure would need an unattended watcher, which reintroduces the regress D.1 avoids.
- **The host surface can stop invoking without being removed.** The weaker and likelier direction: it still exists and still runs, but its invoking condition stops matching. REQ-GDL-003 forecloses the direct form and AC-GDL-003 verifies it, but that binds this deliverable's own wiring — it cannot stop a later edit from reintroducing a filter, and no criterion measures invocation frequency over the deliverable's lifetime.
- **The single trigger fires on every sweep once any entry is permanently non-clean.** This is the failure direction the trigger's own repair created: collapsing to one contract-level condition removed the symptom-enumeration holes and, in exchange, made a permanently non-clean entry a permanent advisory. REQ-GDL-007 is the mitigation and it is partial.
- **Change-leading makes a long-standing non-clean entry quiet by design.** It is announced once and thereafter only counted. Re-announcing it every session is the noise that produces the filter, so the trade is deliberate — but quiet is the subject of this card, and this is the sharpest unresolved tension in the design.
- **The advisory reaches one observer, and §A.4's subject is two.** An advisory rendered in observer 1's session leaves observer 2 exactly where §A.4 found the lead. The design narrows *who must already know the question* from "someone" to "whoever attends a session", which is an improvement rather than a closure.

- **Composing into the landed session-start block changes that block's noise profile.** t326's advisory speaks on one outcome and is silent on every other; REQ-GDL-004 speaks on any non-clean entry. Joining a speak-rarely channel with a speak-often contributor is a decision with a cost, and §A.8 is exactly about what a noisier channel does to its readers. REQ-GDL-007's change-leading is the mitigation and it is partial. The alternative — a second surface — was rejected as worse (§B.3), not as costless.
- **The advisory can be one activation stale by construction.** The refresh trigger *is* decided — host activation, unconditionally (REQ-GDL-003) — but its completion is never awaited (REQ-GDL-012), so a reader sees the most recent *completed* refresh, not one taken this instant. REQ-GDL-006's age disclosure makes that legible; it does not make it untrue. On a surface a user visits rarely, or where the refresh consistently overruns, the persisted result can be arbitrarily old while the advisory renders normally. No criterion bounds that staleness, and bounding it would require awaiting the refresh, which the latency bound forbids.
- **The composition target is owned by an SPEC that has not closed.** Card t326's implementation surfaces are landed on `origin/develop` at `ec15ec2cd`, but its SPEC reads `status: in-progress` (§E). REQ-GDL-010 binds this deliverable to compose onto a surface whose owning card may still change it — the contributor helper, the merge semantics, or the bound convention could move before t326 closes, and nothing here would detect that until the composition broke.

All seven are recorded in `acceptance.md` §D.7 as residual risk, not as solved problems.

## §E Out of Scope

### Out of Scope — the classification and state model
- Which classifications exist, what each means, how an entry is decided into one, the manifest that declares firing expectations, the per-workflow querying, and the set comparison against disk. All of it is `SPEC-GUARD-STATE-MODEL-001`; its card id is pending and recorded in `plan.md` §A.1, the single place the placeholder sits.
- This SPEC consumes only the three-clause contract in REQ-GDL-001 and restates none of the vocabulary (§B.1).

### Out of Scope — the binary-lag comparison itself (but NOT its surfacing path)
- Whether the installed binary's build commit is a strict ancestor of the tree HEAD — the comparison, its verdict vocabulary, and the doctor item that reports it. Owned by card t326.
- **The session-start advisory path is explicitly NOT excluded.** It is in scope as a composition target: REQ-GDL-010 requires this advisory to join that existing block rather than open a second surface, and REQ-GDL-011 records the latency constraint that composition imposes. See §B.3 for the disposition and its reasoning.
- Status, stated precisely because two earlier drafts each got half of it: **t326's implementation surfaces are landed on `origin/develop` at `ec15ec2cd`; its SPEC remains `status: in-progress`.** Measured — `git show origin/develop:.moai/specs/SPEC-BINARY-LAG-VISIBILITY-001/spec.md` returns `version: "0.4.1"`, `status: in-progress`. An earlier draft read *this* tree, found the surfaces absent, and called them unbuilt (they exist — the trees are diverged, §A.10); the correction then over-swung to "landed, not in flight" and conflated code-landed with card-closed. §A.1's "in flight" was right about the card and wrong about the code.
- The distinction is load-bearing rather than pedantic: REQ-GDL-010 binds this deliverable to compose onto a surface owned by an unclosed SPEC, which can still change before t326 closes. Recorded as a risk in §D.3 and `acceptance.md` §D.7.
- The `REQ-BLV-*` range is not cited here: this SPEC cites t326's **source**, measured at a named tree, rather than its requirement numbering.

### Out of Scope — C5, policy-rule firing (named follow-up candidate)
- Applying the same three questions to **policy-layer rules** rather than CI guards. Grounding is §A.7: a rule landed, sat inside its own `paths:`, was cited by name in audits, and its named defect recurred two days later with nothing detecting it.
- Excluded on mechanics, not merit: a workflow has run records and a policy rule has none.

### Out of Scope — C2, warning on unpinned invariant assertions
- Explicitly declined. C2 requires an exemption discriminant: a provenance statement whose subject **is** the mainline correctly carries a moving ref, and pinning it destroys the claim. Mechanized without that discriminant it is a false-positive factory, and it is a different axis.

### Out of Scope — the procedural correlation layer
- The lead session correlating scattered observations, issuing a card, and dispatching it. Not a code artifact.
- Recorded rather than dropped: the sole executor of that layer today is the lead session, and it disappears when the lead dies or is cleared.

### Out of Scope — the always-red variant
- `Graph Freshness` failing on every `develop` push. Card t322's subject (§A.8). This card addresses the silence mechanism, not the constant-noise mechanism that arrives at the same end state.

### Out of Scope — guard correctness
- Whether a guard that fired would have caught a real defect. This card measures firing, not findings.
