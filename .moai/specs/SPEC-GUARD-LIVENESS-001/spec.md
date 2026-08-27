---
id: SPEC-GUARD-LIVENESS-001
title: "Guard firing-liveness: declare when each CI guard should have fired, and make a guard that silently stopped visible (card t333)"
version: "0.1.0"
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

## §A Context and Problem

A guard fired when it was built, then something changed — deployment, trigger, branch model — and it silently stopped, and nothing announced the stop. Because it is an **absence** rather than a failure, it reads as green.

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

This lane closed that pending observation by measurement. Both guards have now actually fired on `develop` push and both succeeded. Full evidence, with commands and verbatim output, is at `.moai/reports/t333/trigger-axis-observation.md`.

The load-bearing point is not that they fired. It is that **they fired, they were correct, and none of that reached anyone**. A firing and a non-firing were indistinguishable from outside until someone went looking. Had they *not* fired, the same plan would have produced the same silence.

### A.4 Instance 3 — the two observations were not of equal strength

Measured while documenting A.3. `spec-lint`'s firing is visible in an unfiltered `gh run list --branch develop`; `docs-i18n-check`'s is **not**, and recovering it required a query targeted at that workflow by name. Absence from the default listing has at least two causes — a `paths:` filter that did not match, or a trigger that cannot fire at all — and the listing does not separate them. The lead session, querying the default way, reproduced the first and could not reproduce the second: two competent readers of the same repository, minutes apart, reached different pictures of which guards were alive.

This is why REQ-GDL-006 forbids the repository-global listing as an evidence source. It is not a preference about query ergonomics; the global listing is measurably incapable of answering the question for a low-frequency guard.

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

### A.7 The always-red variant (recorded, not this card's subject)

`Graph Freshness` failed on every `develop` push in the measured window (3/3). It is card t322's subject. It is recorded because it is a second route to the same end state: a guard that is red on every run stops being read just as thoroughly as one that never runs. Silence and constant noise are different mechanisms arriving at the same place — nobody looks. This card addresses only the silence mechanism.

## §B Relationship to the landed verification-completeness rule

`.claude/rules/moai/development/verification-completeness.md` landed at `7f5b6a947` (card t261) and already carries the observed-failure discipline, the three-part check spec, the two-cell adoption pair, and the mutant probe. **None of it is re-authored here.**

The extension point is measured, and it is one line:

```
$ git show origin/develop:.claude/rules/moai/development/verification-completeness.md \
    | grep -n "WHEN|주기|cadence|지속|periodic"
55:- **(a) WHEN** it must run to be meaningful. A check scheduled at a structurally always-green
```

That `(a)` speaks about **authoring time** — a check scheduled at a structurally always-green moment. Nothing in the rule speaks about a check that was correctly scheduled and later stopped. **The rule watches a check being born; this card watches whether it stays alive.**

Boundary against the rule's `(c)`, stated explicitly so the two clauses are not read as overlapping: the rule's `(c)` covers a failure that **happened and nobody saw** — a red at debug level, absent from traces. This card covers a failure that **never happens at all**. Absence looks green, so `(c)` cannot reach it.

The t241 lane's out-of-scope statement ("규칙 파일 본문 개정 없음") confirms the rule body will not be edited by its own work, so the extension point is stable. This card's rule work is **additive only** (REQ-GDL-015).

## §C Requirements (GEARS)

Budget: Tier M ≤ 16 requirements. **Count: 15.**

### C.1 The expectation record — where firing expectations are written (question (a))

- **REQ-GDL-001** — The system shall carry a guard-liveness manifest declaring one expectation entry per workflow file under `.github/workflows/`. The manifest shall live outside `.moai/config/`, because `moai update` deletes that root wholesale (`CleanMoaiManagedPaths`) and a manifest lost on update is a guard-liveness record that itself silently stops.
- **REQ-GDL-002** — Each manifest entry shall carry the workflow file path, the triggering event or events under which firing is expected, an expectation window, and exactly one measured quantity.
- **REQ-GDL-003** — The measured-quantity vocabulary shall be exactly `fired-at-all`, `fired-with-effect`, and `verdict-rendered`, where `fired-with-effect` excludes runs whose conclusion is `skipped` or `cancelled`, and `verdict-rendered` additionally requires a terminal `success` or `failure`. Each entry names exactly one, so a reader can tell which of two different things a number measures.
- **REQ-GDL-004** — Where a workflow file has no manifest entry, the evaluator shall classify it `UNDECLARED` and report it. It shall not be skipped, ignored, or counted toward a clean result.
- **REQ-GDL-005** — Where a guard is legitimately expected to be quiet outside a release cycle, its entry shall declare that condition explicitly rather than being omitted, so "correctly quiet" is a recorded expectation rather than an absence a reader must infer.

### C.2 The evaluator — what verifies the expectations (question (b))

- **REQ-GDL-006** — When the evaluator runs, it shall query run history **per workflow file**, and shall not derive any guard's last-fired time from a repository-global run listing (§A.4, §A.6).
- **REQ-GDL-007** — Where a per-workflow query returns no runs within the retained window, the evaluator shall classify the guard `UNKNOWN` and shall not report it as "never fired". `gh run list` reports the runs the forge retained; a run aged out of retention is indistinguishable from a run that never happened.
- **REQ-GDL-008** — While a guard's most recent qualifying run is older than its declared expectation window, the evaluator shall classify it `STALE` and name the expectation it missed.
- **REQ-GDL-009** — When the evaluator emits a result, it shall carry its own measurement timestamp and its coverage counts: entries declared, entries successfully queried, entries `UNKNOWN`, and workflow files `UNDECLARED`.
- **REQ-GDL-010** — The evaluator shall not report an all-clear while its successfully-queried count is zero. A sweep of nothing asserts nothing.
- **REQ-GDL-011** — The evaluator shall be pull-based and invoked from an attended surface. It shall not be implemented as a scheduled workflow, because a scheduled watcher is itself subject to the defect it watches for and starts an unbounded regress.
- **REQ-GDL-012** — The evaluator shall not write, commit, push, open an issue, or mutate any forge state. It reads and reports.

### C.3 Reachability — who sees the silence (question (c))

- **REQ-GDL-013** — When the evaluator's result carries at least one `STALE` or `UNDECLARED` entry, the harness shall surface it to the operator as a non-blocking advisory at an already-attended surface. A liveness verdict visible to nobody is the same defect one layer up.
- **REQ-GDL-014** — The advisory shall carry the age of the measurement it reports, so a stale advisory declares its own staleness rather than reading as a current all-clear.

### C.4 Doctrine

- **REQ-GDL-015** — `.claude/rules/moai/development/verification-completeness.md` shall gain an additive continued-firing clause stating that a check's completion does not survive a change to its trigger, its deployment, or its branch model. No existing text in that file shall be modified.

## §D How the design answers the self-application constraint

The card's hardest constraint: build a periodic check and that check becomes subject to this very card — a check that catches non-firing catches nothing if it itself stops firing.

The design answers it by **not being periodic**. The evaluator is pull-based (REQ-GDL-011) and runs at a moment that already happens for other reasons. Its firing is *entailed* by someone working, not scheduled independently of them, so there is no cadence of its own that could be silently missed. A scheduled watcher would have one, and the forge additionally disables scheduled workflows after a period of repository inactivity — a silent stop by design, in the exact shape this card exists to catch.

Two further properties close the shallower openings:

- The evaluator declares its own coverage (REQ-GDL-009) and refuses an all-clear on an empty sweep (REQ-GDL-010), so a degraded run announces itself instead of rendering green.
- The advisory carries its own measurement age (REQ-GDL-014), so a stale verdict is legible as stale.

**This relocates the regress rather than eliminating it, and the SPEC says so.** If the attended surface hosting the evaluator is itself removed, the evaluator stops with it — the same defect class, one layer up. Full closure would require an unattended watcher, which reintroduces the regress this design avoids. Recorded as residual risk in `acceptance.md` §D.7, not claimed as solved.

## §E Out of Scope

### Out of Scope — the binary-lag state comparison
- Whether the installed binary's build commit is a strict ancestor of the tree HEAD, and the session-start advisory that surfaces it. Owned by card t326 (`SPEC-BINARY-LAG-VISIBILITY-001`, REQ-BLV-001..009, in flight). No part of it is re-specified here.
- The A/B binary-comparison recipe that diagnosed the t298 instance.

### Out of Scope — the procedural correlation layer
- The lead session correlating scattered observations, issuing a card, and dispatching it. Excluded on the same grounds card t326 §7 excluded it: it is not a code artifact, and including it inflates a Tier M card.
- Recorded rather than merely dropped, because the contrast matters: the sole executor of that layer today is the lead session, and it disappears when the lead dies or is cleared. t326's deployment check, by contrast, runs without a lead. See `acceptance.md` §D.7.

### Out of Scope — the always-red variant
- `Graph Freshness` failing on every `develop` push. Card t322's subject (§A.7). This card addresses the silence mechanism only, not the constant-noise mechanism that arrives at the same end state.

### Out of Scope — guard correctness
- Whether a guard that fired would have caught a real defect. This card measures firing, not findings. §A.3's evidence read exit status only, and the SPEC claims nothing beyond that.

### Out of Scope — repairing individual guards
- Rewiring `spec-status-auto-sync.yml`'s trigger (§A.5), or any other guard the evaluator classifies `STALE`. The evaluator's job is to make the classification visible; acting on a particular classification is a separate card.
