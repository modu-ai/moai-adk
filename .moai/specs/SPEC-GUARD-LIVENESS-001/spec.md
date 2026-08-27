---
id: SPEC-GUARD-LIVENESS-001
title: "Guard firing-liveness: declare when each CI guard should have fired, and make a guard that silently stopped visible (card t333)"
version: "0.4.0"
status: draft
created: 2026-08-28
updated: 2026-08-28
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: ".moai/guards, internal/cli, internal/guard, .claude/rules/moai/development"
lifecycle: spec-anchored
tags: "guard, liveness, cadence, ci-workflows, silent-absence, event-history, t333"
tier: M
era: V3R6
---

# SPEC-GUARD-LIVENESS-001 — Guard Firing-Liveness

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-08-28 | manager-spec | Initial plan-phase authoring (card t333). Scope narrowed to the event-history axis; the binary-lag state comparison stays with card t326. |
| 0.2.0 | 2026-08-28 | manager-spec | Amended after the evidence artifact was extended. §A.3 split into Claim A / Claim B (the two observations are not of equal strength). §A.4 rewritten to the complete third instance — the gap closed because one observer handed the other the question, and nothing would have told the lead its picture was wrong. the always-red subsection (renumbered to §A.8 in v0.3.0) extended with the always-red obligation. §D restructured into two independent binding constraints (self-observation, unprompted discoverability) plus what the design does not close. REQ-GDL-013 strengthened to require zero operator input; REQ-GDL-016 added (change-leading advisory). AC-GDL-013 and AC-GDL-014 added. Counts 15→16 REQ, 12→14 AC. |
| 0.3.0 | 2026-08-28 | manager-spec | Three inputs from the t241 lane's prediction-ledger verdict (all six rows `false`). §A.7 added — instance 6, a `-run` selector matching two tests that did not exist, recurring inside the landed rule's own `paths:` two days after it landed; the always-red subsection renumbered §A.7→§A.8 and its inbound references updated. §B.1 added — a policy rule landing is not a policy rule working. §C.1.1 added — the one-number-two-events trap with the lane's measured occurred/survived table; REQ-GDL-003 strengthened to forbid it. REQ-GDL-001 extended to require the watched set be held as data; §D.4 added — subject-agnostic shape without a subject-agnostic deliverable. §E gains two entries: C5 (policy-rule firing) as a named follow-up candidate with its grounding, and C2 (unpinned invariant assertions) explicitly declined on the lane's exemption-discriminant warning. AC-GDL-015 added. Counts 16 REQ (unchanged, at budget), 14→15 AC. |
| 0.4.0 | 2026-08-28 | manager-spec | Precision correction to the card's shape. §A.0 added — absent execution is not suppressed failure; an exit code answers a question about the selected set, and nothing compares the selected set against the intended one. §A.7's instance restated: the value is that the absence never reached the exit code, and the green was true about what was selected rather than false. §B's boundary against the landed rule's `(c)` rewritten on the same distinction — reachability presupposes something to reach. REQ-GDL-004 restated as the set comparison it always was; §C.2 gains a preamble naming the three requirements that force it. plan.md gains B-0 and two anti-patterns (outcome-only evaluator, hidden-failure framing). No requirement or criterion added — counts unchanged at 16 REQ / 15 AC. |

## §A Context and Problem

A guard fired when it was built, then something changed — deployment, trigger, branch model — and it silently stopped, and nothing announced the stop. Because it is an **absence** rather than a failure, it reads as green.

### A.0 Absent execution is not suppressed failure

Two shapes are easy to blur, and this card is entirely the second. Getting them the wrong way round misdescribes every instance below.

- **Suppressed failure** — something ran, went red, and the red did not reach anyone. This is the landed rule's §1.2 **(c)**, and it is *not* this card's subject (§B).
- **Absent execution** — nothing ran, so nothing could go red, and the absence is reported as success by a mechanism that is answering a narrower question than the reader thinks it is.

The distinction is not pedantic, because it rules out the obvious framing. It is wrong to say the check *should have failed and did not*: there was no failure to suppress. The accurate sentence is **"nothing failed, and there was nothing there to fail."**

What every instance below shares is a mechanism correctly answering a question about a set that had silently become the wrong set. An exit code of zero means *everything selected passed*; it does not mean *everything that should have passed, passed*. Nothing in the mechanism compares the selected set against the intended one. **Green is accurate and uninformative at the same time.**

That is also the sharpest form of design question (b). A liveness check that can only read the outcomes of things that executed inherits exactly this blindness. It must ask not "did this run pass?" but **"was this in the set that ran, and should it have been?"** — which is why REQ-GDL-004 and REQ-GDL-007 are load-bearing rather than housekeeping.

### A.1 The axis this card owns, and the axis it does not

This card owns the **event-history axis**: *when did this guard last fire, and was it supposed to have fired by now?*

The **state axis** — whether the installed binary's build commit is a strict ancestor of the tree HEAD, and the session-start advisory that surfaces it — is card t326's (`SPEC-BINARY-LAG-VISIBILITY-001`, REQ-BLV-001..009, in flight). No part of it is re-specified here.

The boundary is **state vs history**, and its sharpest statement came from the t326 lane:

> t326's yardstick measures the `pull_request`-triggered guards as perfectly current: the files sit in the tree and the binary is not behind. The fact that they do not fire does not register on that yardstick.

The deployment axis is not dropped here. The t298 instance below happens to also be expressible as state, which is why t326 catches *that instance* — but the general form ("fired at authoring time, then silently stopped") is untouched by t326 and belongs to this card.

### A.2 Instance 1 — deployment axis (t298, 2026-08-27)

`SPEC-INTEGRATION-LOCK-LIVENESS-001` fixed integration-lock liveness and landed, but the installed `~/go/bin/moai` was built from a commit roughly six hours older, so the fix never ran. Three lanes each independently observed the `reclaimable` misread; one real eviction happened without `--force`; the lead issued a card for a defect that was already fixed. Nothing in any signal said "deployment lag". It was resolved only when one lane ran an isolated A/B of the old and new binaries side by side.

### A.3 Instance 2 — trigger axis (git-flow transition, 2026-08-27)

The git-flow transition removed card PRs, so two guards whose trigger was `pull_request` stopped running on `develop` entirely. Card t314 rewired both — `spec-lint.yml` and `docs-i18n-check.yml` gained `push` on `develop` — and closed with the first firing left as a **pending opportunistic observation**; the operator's recorded decision was literally "go with opportunistic observation".

This lane closed that pending observation by measurement. Full evidence, with commands and verbatim output, is at `.moai/reports/t333/trigger-axis-observation.md`. **The two observations are not of equal strength, and the difference is itself this card's subject — they are stated separately and must not be folded into one sentence.**

- **Claim A — `spec-lint`.** Fired on `develop` push, success, on three consecutive pushes. Visible in an **unfiltered** `gh run list --branch develop`. Independently reproduced by the lead session from its own query.
- **Claim B — `docs-i18n-check`.** Fired once on `develop` push, success. **Not visible in the unfiltered listing.** Recovering it required a query targeted at that workflow by name.

The load-bearing point is not that they fired. It is that **they fired, they were correct, and none of that reached anyone**. A firing and a non-firing were indistinguishable from outside until someone went looking. Had they *not* fired, the same plan would have produced the same silence.

### A.4 Instance 3 — checking is not sufficient, because the default view is lossy from inside

Produced by measuring the other two, which is what makes it the strongest of the three.

`docs-i18n-check`'s absence from the default listing has at least two causes, and the listing does not separate them: its `paths:` filter did not match this round, or its trigger is broken and it cannot fire at all. Both render as *not in the list*. The path-filter reading is an inference from opening the workflow file — and **a coherent explanation is not an observation**.

The lead session, querying the default way, reproduced Claim A and could not reproduce Claim B. Two competent readers of the same repository, minutes apart, held different pictures of which guards were alive — not because either measured badly, but because the surface does not carry the answer.

**Then the lead did reproduce Claim B — but only after this lane handed it the query.** It ran the targeted query, got matching output, and recorded why it could: the workflow's name had reached it in a message from this lane. It did not suspect the answer; it was handed the question.

That is the complete form of the instance. The gap did not close because a second observer looked harder. It closed because one observer handed the other the question. Absent that hand-off the lead's picture would have stayed as it was — and **nothing anywhere would have told it the picture was wrong. There is no signal for "your view of which guards are alive is incomplete."** The incompleteness is silent by exactly the mechanism the non-firing is.

Two requirements descend from this instance:

- REQ-GDL-006 forbids the repository-global listing as an evidence source. Not a preference about query ergonomics — the global listing is measurably incapable of answering the question for a low-frequency guard, and its incapacity is invisible from inside it.
- REQ-GDL-013 requires the answer to arrive **unprompted**. A targeted query can only be issued by someone who already suspects the answer, so a design answering only when queried has not solved this problem — it has relocated it into whoever is expected to already know. §D states this as a binding constraint and tests the design against it.

### A.5 Instance 4 — fires, does nothing, reads green (measured in this plan phase)

`spec-status-auto-sync.yml` is triggered by `pull_request: types: [closed]` only — no `push`, no `schedule`, no `workflow_dispatch`. Under git-flow, card PRs are abolished, so its remaining trigger source is release PR merges. Its three most recent runs, measured on `d34a789a4`:

```
$ gh run list --workflow spec-status-auto-sync.yml --limit 3 \
    --json createdAt,event,conclusion,headBranch
08-27T09:35  pull_request  skipped  WT-codex-launcher
08-27T08:45  pull_request  skipped  WT-glm-settings-rename
08-27T08:38  pull_request  skipped  WT-main-stamp-repair
```

All three are `skipped` — the workflow's own `if: github.event.pull_request.merged == true` gate declined. A liveness check asking only "did it fire?" reads these as green. A guard that fires and does nothing reaches the same end state as a guard that does not fire, and it is *harder* to see, because a run record exists. This is why the expectation vocabulary (REQ-GDL-003) separates `fired-at-all` from `fired-with-effect`.

### A.6 Instance 5 — the census gap (measured in this plan phase)

`.github/workflows/` holds **18** workflow files. Grouping the 100 most recent runs by workflow name yields **11** distinct names; seven files never appear. Some of those are legitimately release-only, but the listing does not say which — and the 100-run window it was drawn from spans **about three hours** (`2026-08-27T11:53:56Z` → `14:57:34Z`), because a handful of high-frequency workflows saturate it. Nothing today distinguishes "correctly quiet" from "silently stopped".

### A.7 Instance 6 — a selector that matched nothing, inside the landed rule's own scope

The cleanest of the set, and the one that fixes the mechanism's generality.

`.claude/rules/moai/development/verification-completeness.md` landed at `7f5b6a947` — author and committer date both `Tue Aug 25 13:05:04 2026 +0900`, verified in this plan phase with `git show -s --format='%H%n%ad%n%cd'`. Two days later, on 2026-08-27, its own named defect recurred in `.moai/specs/**/progress.md` — **squarely inside the rule's `paths:` scope**. A `-run` selector named three tests; only one existed under that name. The run printed `ok ... 0.249s`, and nobody saw it until sync close. Source: the t241 lane's prediction-ledger verdict, whose six rows all recorded `false`.

The value of the instance is not that two checks were missing. It is that **their absence never reached the exit code.**

That green was **not false**. It was **true about what was selected** — `rc 0` means everything selected passed. It does not mean everything that should have passed, passed. The exit code answers a question about the selected set, and the selected set had itself become wrong; nothing in the mechanism compares the selected set against the intended one. **Nothing failed, and there was nothing there to fail** (§A.0).

This instance is on a third axis — not deployment (§A.2), not trigger (§A.3), but a **selector matching nothing** — and that is what makes it load-bearing for scope. The defect is not a property of GitHub Actions. **Any check whose non-execution is indistinguishable from its success has it.** That generality is the justification for this SPEC's mechanism being about *firing*, and about a watched set held as data, rather than about workflow files (§D.4).

### A.8 The always-red variant (recorded, not this card's subject)

`Graph Freshness` failed on every `develop` push in the measured window (`d34a789a4`, `0c7457f8d`, `812ee01fc` — 3/3). Repairing it is card t322's subject and is out of scope here (§E).

It is recorded because it is a second route to the same end state, and because it bears directly on question (c). A guard that is red on every single run stops being read just as thoroughly as one that never runs: the signal is present, carries no information, and is filtered out by every reader. **Silence and constant noise are different mechanisms arriving at the same place — nobody looks.**

The consequence for this card is a design obligation, not a scope expansion: **making a channel and making a channel that gets read are different jobs.** A new advisory rendered next to a permanently-red neighbour inherits that neighbour's filter, so the design must be honest about whether it survives one. REQ-GDL-016 is that obligation, and §D tests the design against it rather than assuming it.

## §B Relationship to the landed verification-completeness rule

`.claude/rules/moai/development/verification-completeness.md` landed at `7f5b6a947` (card t261) and already carries the observed-failure discipline, the three-part check spec, the two-cell adoption pair, and the mutant probe. **None of it is re-authored here.**

The extension point is measured, and it is one line:

```
$ git show origin/develop:.claude/rules/moai/development/verification-completeness.md \
    | grep -n "WHEN|주기|cadence|지속|periodic"
55:- **(a) WHEN** it must run to be meaningful. A check scheduled at a structurally always-green
```

That `(a)` speaks about **authoring time** — a check scheduled at a structurally always-green moment. Nothing in the rule speaks about a check that was correctly scheduled and later stopped. **The rule watches a check being born; this card watches whether it stays alive.**

Boundary against the rule's `(c)`, stated explicitly so the two clauses are not read as overlapping. The rule's `(c)` covers a **suppressed failure** — something ran, went red, and the red did not reach anyone: a red at debug level, absent from traces. This card covers an **absent execution** — nothing ran, so no red exists to be reached (§A.0).

`(c)` is a reachability clause, and reachability presupposes something to reach. When the execution is absent there is no signal to route, no exit code to surface, and no log level to raise: the mechanism reports success accurately, about a set that had silently become the wrong one. That is why `(c)` cannot reach this defect, and why the extension is genuinely additive rather than a restatement of an existing clause at a different volume.

The t241 lane's out-of-scope statement ("규칙 파일 본문 개정 없음") confirms the rule body will not be edited by its own work, so the extension point is stable. This card's rule work is **additive only** (REQ-GDL-015).

### B.1 A policy rule landing is not a policy rule working

§A.7 is evidence about this very rule, and it should be stated plainly rather than left as an awkward implication. The rule was **correct**, **in scope**, and **cited by name in audits** — and the defect it names recurred two days after it landed, inside its own `paths:`. The t241 lane's prediction ledger recorded `false` on all six rows.

Nothing detects at that layer. A rule is a policy artifact: it changes what a careful reader does, and it has no run records, no exit code, and no mechanical firing. That is exactly why §A.7's recurrence went unseen until sync close, and it is the strongest available argument that this card's subject is not a CI-workflow quirk.

It is also why the deliverable here stays mechanical and stays scoped to things that *have* run records (§E). Extending the same three questions to policy rules is a real follow-up, named as C5 in §E, not a widening of this card.

## §C Requirements (GEARS)

Budget: Tier M ≤ 16 requirements. **Count: 16.**

### C.1 The expectation record — where firing expectations are written (question (a))

- **REQ-GDL-001** — The system shall carry a guard-liveness manifest declaring one expectation entry per workflow file under `.github/workflows/`. The manifest shall live outside `.moai/config/`, because `moai update` deletes that root wholesale (`CleanMoaiManagedPaths`) and a manifest lost on update is a guard-liveness record that itself silently stops. **The manifest shall hold its watched set as data — a list of watched subjects each carrying its own kind, locator, and expected cadence — and shall not be shaped so that only a GitHub workflow can occupy an entry** (§D.4, §E C5).
- **REQ-GDL-002** — Each manifest entry shall carry the workflow file path, the triggering event or events under which firing is expected, an expectation window, and exactly one measured quantity.
- **REQ-GDL-003** — The measured-quantity vocabulary shall be exactly `fired-at-all`, `fired-with-effect`, and `verdict-rendered`, where `fired-with-effect` excludes runs whose conclusion is `skipped` or `cancelled`, and `verdict-rendered` additionally requires a terminal `success` or `failure`. **Each entry shall name exactly one, and no entry shall be written whose single number is asked to measure both whether a guard fired and whether a firing caught anything.** These are different axes: a guard that runs faithfully and catches nothing scores full marks on the first (§C.1.1).
- **REQ-GDL-004** — Where a workflow file has no manifest entry, the evaluator shall classify it `UNDECLARED` and report it. It shall not be skipped, ignored, or counted toward a clean result. This requirement is the set comparison §A.0 names: it asks whether a subject **was in the set that was examined, and should have been**, rather than reading outcomes only for subjects the manifest already knew about. An evaluator without it reproduces the defect at its own layer — accurate about what it looked at, silent about what it never looked at.
- **REQ-GDL-005** — Where a guard is legitimately expected to be quiet outside a release cycle, its entry shall declare that condition explicitly rather than being omitted, so "correctly quiet" is a recorded expectation rather than an absence a reader must infer.

#### C.1.1 The one-number-two-events trap, measured

The trap REQ-GDL-003 forbids is not hypothetical here. The t241 lane's prediction ledger predicted **"0 audit findings"** as its success signal on four rows — a number that moves the *wrong way* when the rule works, because authors still write shallow criteria and the audit now catches them. Its own measurement separates the two events the single number was carrying:

| Ledger row | Defect occurred | Survived to adoption |
|---|---:|---:|
| VC-1 | 3 | 2 |
| VC-2 | 7 | 1 |
| VC-4 | 5 | 1 |
| VC-6 | 2 | 1 |

The same split sits directly in this SPEC's path. An expectation of the form "this guard should fire every N days" carries both **firing count** and **whether a firing caught anything**, and measuring only the first awards full marks to a guard that runs faithfully and catches nothing — the §A.5 shape one layer along. The vocabulary in REQ-GDL-003 exists to make an author choose, and that lane is writing "survived-to-adoption count" into its own next ledger, so the two cards stay consistent.

### C.2 The evaluator — what verifies the expectations (question (b))

The requirements below are shaped by §A.0. An evaluator that reads only the outcomes of runs that happened is answering the same narrower question every instance in §A was undone by. Three of them exist to force the set comparison instead: REQ-GDL-004 asks whether a subject was in the examined set at all, REQ-GDL-007 refuses to read an empty result as an answer, and REQ-GDL-010 refuses a verdict drawn from an empty sweep.

- **REQ-GDL-006** — When the evaluator runs, it shall query run history **per workflow file**, and shall not derive any guard's last-fired time from a repository-global run listing (§A.4, §A.6).
- **REQ-GDL-007** — Where a per-workflow query returns no runs within the retained window, the evaluator shall classify the guard `UNKNOWN` and shall not report it as "never fired". `gh run list` reports the runs the forge retained; a run aged out of retention is indistinguishable from a run that never happened.
- **REQ-GDL-008** — While a guard's most recent qualifying run is older than its declared expectation window, the evaluator shall classify it `STALE` and name the expectation it missed.
- **REQ-GDL-009** — When the evaluator emits a result, it shall carry its own measurement timestamp and its coverage counts: entries declared, entries successfully queried, entries `UNKNOWN`, and workflow files `UNDECLARED`.
- **REQ-GDL-010** — The evaluator shall not report an all-clear while its successfully-queried count is zero. A sweep of nothing asserts nothing.
- **REQ-GDL-011** — The evaluator shall be pull-based and invoked from an attended surface. It shall not be implemented as a scheduled workflow, because a scheduled watcher is itself subject to the defect it watches for and starts an unbounded regress.
- **REQ-GDL-012** — The evaluator shall not write, commit, push, open an issue, or mutate any forge state. It reads and reports.

### C.3 Reachability — who sees the silence (question (c))

- **REQ-GDL-013** — When the evaluator's result carries at least one `STALE` or `UNDECLARED` entry, the harness shall surface it to the operator as a non-blocking advisory at an already-attended surface, **without the operator issuing any query and without the operator needing to know which guard to ask about**. A liveness verdict that answers only when queried has relocated the defect into whoever is expected to already know the question (§A.4).
- **REQ-GDL-014** — The advisory shall carry the age of the measurement it reports, so a stale advisory declares its own staleness rather than reading as a current all-clear.
- **REQ-GDL-016** — The advisory shall lead with the entries whose classification **changed** since the previously rendered result, and shall carry any unchanged `STALE` or `UNDECLARED` entries as a compact standing count rather than as a re-rendered list. A channel that reprints an identical block every session trains the filter that removes it, and a new advisory rendered beside a permanently-red neighbour inherits that filter (§A.8).

### C.4 Doctrine

- **REQ-GDL-015** — `.claude/rules/moai/development/verification-completeness.md` shall gain an additive continued-firing clause stating that a check's completion does not survive a change to its trigger, its deployment, or its branch model. No existing text in that file shall be modified.

## §D How the design answers the two binding constraints

Two constraints bind this design, and they are independent — a design can pass either while failing the other, and failing either leaves the same defect one level up.

### D.1 Constraint 1 — self-observation

*Build a periodic check and that check becomes subject to this very card: a check that catches non-firing catches nothing if it itself stops firing.*

The design answers it by **not being periodic**. The evaluator is pull-based (REQ-GDL-011) and runs at a moment that already happens for other reasons. Its firing is *entailed* by someone working, not scheduled independently of them, so there is no cadence of its own that could be silently missed. A scheduled watcher would have one, and the forge additionally disables scheduled workflows after a period of repository inactivity — a silent stop by design, in the exact shape this card exists to catch.

Two further properties close the shallower openings:

- The evaluator declares its own coverage (REQ-GDL-009) and refuses an all-clear on an empty sweep (REQ-GDL-010), so a degraded run announces itself instead of rendering green.
- The advisory carries its own measurement age (REQ-GDL-014), so a stale verdict is legible as stale.

### D.2 Constraint 2 — unprompted discoverability

*A targeted query can only be issued by someone who already suspects the answer, so a design that relies on a reader knowing which question to ask has relocated the problem into whoever is expected to already know* (§A.4).

The design is tested against three questions rather than assumed to pass them.

**Does it surface without being asked, or answer only when queried?** It surfaces. REQ-GDL-013 makes the advisory an output of an already-attended surface, not a verb the operator invokes. The operator supplies no guard name, no workflow file, and no query — this is precisely the input the lead session did not have and could not have produced.

**If it answers only when queried, what tells a reader that a query is owed, and which one?** Not applicable by construction, and that is the point: this question is the one the current state fails, and every candidate that keeps a query in the loop inherits the failure. It is why "add a `moai guard liveness` verb and document it" was rejected as a complete answer — a documented verb is still a question someone must know to ask.

**Is the thing that would announce a missing guard itself discoverable by someone who does not already know it exists?** Yes, in the only sense available: it is not a thing to be discovered. It arrives in the operator's session unbidden, so the reader needs no prior knowledge of the mechanism's existence. This is the property that a manifest plus a CLI verb, on their own, would not have had.

### D.3 What the design does not close

**The regress is relocated, not eliminated, and the SPEC says so rather than claiming closure.**

- If the attended surface hosting the evaluator is removed, the evaluator stops with it — the same defect class, one layer up. Full closure would need an unattended watcher, which reintroduces the regress D.1 avoids.
- REQ-GDL-016 buys legibility against an always-red neighbour by leading with *changes* rather than reprinting a standing list — but a standing count is still a thing a reader can learn to skip. Change-leading raises the cost of habituation; it does not abolish it. **Making a channel and making a channel that gets read are different jobs**, and this design does the first well and the second only partly.
- REQ-GDL-016 also has its own failure direction: an entry that goes `STALE` and stays `STALE` is announced once and thereafter only counted. That is deliberate — re-announcing it every session is the noise that produces the filter — but it means a long-standing `STALE` is quiet by design, and quiet is what this card is about.

All three are recorded in `acceptance.md` §D.7 as residual risk, not as solved problems.

### D.4 Subject-agnostic shape, without a subject-agnostic deliverable

§A.7 establishes that the defect is not a property of GitHub Actions: any check whose non-execution is indistinguishable from its success has it — a workflow that stopped firing, a binary that was never redeployed, a `-run` selector matching two tests that do not exist. The deliverable nonetheless stays scoped to CI guards (§E), because a workflow has run records and a policy rule has none, and those are different mechanics rather than one mechanism at two scales.

The reconciliation is a shape constraint, not a scope expansion. REQ-GDL-001 requires the watched set to be **data**: each entry carries its kind, its locator, and its expected cadence, so a second kind of subject can be added later by adding entries and a reader for that kind — not by rewriting the manifest schema and the classification vocabulary around it.

This is also a design smell test worth applying during M1, and `plan.md` §F records it as such: **if the schema cannot accommodate a second kind of subject without being rewritten, the schema has hardcoded its subject** and should be reshaped before the census is populated. The test costs nothing at M1 and is expensive to apply after 18 entries exist.

## §E Out of Scope

### Out of Scope — the binary-lag state comparison
- Whether the installed binary's build commit is a strict ancestor of the tree HEAD, and the session-start advisory that surfaces it. Owned by card t326 (`SPEC-BINARY-LAG-VISIBILITY-001`, REQ-BLV-001..009, in flight). No part of it is re-specified here.
- The A/B binary-comparison recipe that diagnosed the t298 instance.

### Out of Scope — C5, policy-rule firing (named follow-up candidate)
- Applying the same three questions — where expectations live, what verifies them, who sees the silence — to **policy-layer rules** rather than CI guards. Named here as a follow-up candidate rather than passed over in silence, because the grounding already exists and a later reader should not have to rediscover it.
- The grounding is §A.7 and §B.1: a rule landed at `7f5b6a947`, sat inside its own `paths:` scope, was cited by name in audits — and its named defect recurred two days later with nothing detecting it. The t241 lane's prediction ledger recorded `false` on all six rows.
- Excluded from the deliverable on mechanics, not on merit: a workflow has run records and a policy rule has none, so verifying a rule keeps firing is a different problem, not a bigger instance of this one. Widening the deliverable would inflate a Tier M card.
- What this card owes C5 is a shape, and it pays it: REQ-GDL-001 requires the watched set to be data, so a second kind of subject can be added without rewriting the manifest (§D.4).

### Out of Scope — C2, warning on unpinned invariant assertions
- Explicitly declined rather than merely unaddressed, on the t241 lane's own warning. C2 requires an exemption discriminant: a provenance statement whose subject **is** the mainline correctly carries a moving ref, and pinning it destroys the claim being made. Mechanized without that discriminant it is a false-positive factory.
- It is also a different axis — assertion pinning, not firing liveness — so importing it here would buy a false-positive source in exchange for nothing this card needs.

### Out of Scope — the procedural correlation layer
- The lead session correlating scattered observations, issuing a card, and dispatching it. Excluded on the same grounds card t326 §7 excluded it: it is not a code artifact, and including it inflates a Tier M card.
- Recorded rather than merely dropped, because the contrast matters: the sole executor of that layer today is the lead session, and it disappears when the lead dies or is cleared. t326's deployment check, by contrast, runs without a lead. See `acceptance.md` §D.7.

### Out of Scope — the always-red variant
- `Graph Freshness` failing on every `develop` push. Card t322's subject (§A.8). This card addresses the silence mechanism only, not the constant-noise mechanism that arrives at the same end state.

### Out of Scope — guard correctness
- Whether a guard that fired would have caught a real defect. This card measures firing, not findings. §A.3's evidence read exit status only, and the SPEC claims nothing beyond that.

### Out of Scope — repairing individual guards
- Rewiring `spec-status-auto-sync.yml`'s trigger (§A.5), or any other guard the evaluator classifies `STALE`. The evaluator's job is to make the classification visible; acting on a particular classification is a separate card.
