---
id: SPEC-GRAPH-FRESHNESS-CADENCE-001
title: "Graph-freshness cadence: described-worthy metric predicate, threshold re-derivation, and non-inheriting failure attribution"
version: "0.1.0"
status: draft
created: 2026-08-27
updated: 2026-08-27
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/graph, internal/mx, internal/config, .moai/config/sections/gate.yaml"
lifecycle: spec-anchored
tags: "graph, codemaps, freshness, metric, threshold, attribution, cadence"
era: V3R6
tier: M
related_specs: [SPEC-V3R6-GRAPH-FRESHNESS-001, SPEC-V3R6-GRAPH-FRESHNESS-002, SPEC-STAMP-REACHABILITY-001]
---

# SPEC-GRAPH-FRESHNESS-CADENCE-001 — Graph-freshness cadence

## HISTORY

| Version | Date | Change | Author |
|---------|------|--------|--------|
| 0.1.0 | 2026-08-27 | Initial plan-phase authoring for card t322. All baseline figures in §B re-measured in worktree `.claude/worktrees/t322` at HEAD `d2cba5e21`; the orchestrator's dispatch figures (55 / 0 / 5, 21 / 198) reproduced exactly. Three judgments adjudicated in §D with rejected alternatives recorded. Ordering basis against t311 / t304 established in §F. | manager-spec |

## §A. Problem Statement

The `graph-freshness` CI job reds on integration after integration into `develop`, and the only
remedy in use is a manual restamp. A restamp moves the measurement anchor; it does not correct a
single word of the six curated codemaps documents. The red therefore returns on the next
integration, and the documents stay as wrong as they were.

Three consecutive integrations failed the job, and the fourth — a restamp — passed:

| Integration | Card | `graph-freshness` verdict |
|---|---|---|
| `3809d1d36` | t228 | failure |
| `50de25e5a` | t308 | failure |
| `48eb945df` | t301 | failure |
| `d2cba5e21` | t228 restamp | success |

This SPEC addresses **why the red keeps returning** and **what change stops it**. The inaccuracy of
the codemaps documents themselves is deliberately out of scope — see §E.

## §B. Measured Baseline

Every figure below was measured in this worktree (`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t322`,
branch `WT-graph-freshness-cadence`, HEAD `d2cba5e21`) during plan-phase authoring on 2026-08-27.
Each row names the command that produced it.

### B.1 — The failures are inherited, not individually caused

`git diff --name-only <merge>^1..<merge> -- internal cmd pkg | grep -c .`

| Integration | Files under described roots | Alone exceeds threshold 40? |
|---|---|---|
| `3809d1d36` (t228) | **55** | yes |
| `50de25e5a` (t308) | **0** | no — contributed nothing, still failed |
| `48eb945df` (t301) | **5** | no — still failed |

The metric is cumulative since the stamped commit, so once one integration pushes the running count
past the threshold, every later integration inherits a red that is not its own until a human
notices and restamps.

Cumulative confirmation. The stamp in force before the streak was `9326b5478d0f…`
(`git show 3809d1d36^1:.moai/project/codemaps/provenance.json`). At the third failure:

```
git diff --name-only 9326b5478d0f51979dfb498527458dcea5e0370b 48eb945df -- internal cmd pkg | grep -c .
→ 65
```

### B.2 — The metric counts files no curated document could describe

`gitDiffNameCount` (`internal/graph/check.go`, the function immediately preceding `gitOutput`)
unions `git diff --name-only <stamp> -- <roots>` with `git ls-files --others --exclude-standard --
<roots>` and returns the set size. It applies no extension filter, no `testdata` exclusion, and no
predicate asking whether a path is architectural surface at all.

Composition of t228's 55 files:

| Category | Count |
|---|---|
| `internal/astgrep/testdata/rule-tests/**` fixtures | 42 |
| `internal/template/templates/**` rule YAMLs | 7 |
| `_test.go` | 4 |
| production `.go` (`coverage_matrix.go`, `rule_severity.go`) | **2** |

The same shape holds for the described roots as a whole (`git ls-files internal cmd pkg`):
**3,628** tracked files, of which **998** are production Go outside `testdata/` and outside
`internal/template/templates/`. 72% of the surface the metric counts is not architectural surface.

And the curated documents corroborate it. Of the **80** distinct `internal|pkg|cmd` paths cited
across `.moai/project/codemaps/*.md`, **0** contain a `testdata` segment, and the template payload
tree is cited exactly once — as the bare directory `internal/template/templates/`, never per file.

The counterfactual is decisive. Applying a described-worthy predicate (production `.go`; excluding
`_test.go`, any `/testdata/` segment, and `internal/template/templates/**`) to the same cumulative
window that produced 65:

```
git diff --name-only 9326b5478d… 48eb945df -- internal cmd pkg \
  | grep '\.go$' | grep -v '_test\.go$' | grep -v '/testdata/' \
  | grep -v '^internal/template/templates/' | grep -c .
→ 2
```

**65 → 2.** The entire three-integration red streak was produced by files the codemaps documents do
not describe and never have. Per integration under the same predicate: t228 → 2, t308 → 0,
t301 → 0.

The tree's live state says the same thing. The five files currently differing from the stamp
`d2fcecc8b40d1cb…` are `internal/template/catalog.yaml` and four files under
`internal/template/templates/` — zero production Go.

### B.3 — The threshold encodes an implicit refresh cadence that no longer holds

`SPEC-V3R6-GRAPH-FRESHNESS-001/progress.md` records the calibration verbatim: `-10` → 137, `-50` →
233, decision *"threshold 40 retained — ≈2-3 commits of typical described-source churn"*.

Re-running the same command on this tree:

| Window | Then | Now |
|---|---|---|
| `git log -10 --name-only --pretty=format: -- internal cmd pkg \| sort -u \| wc -l` | 137 | **21** |
| same, `-50` | 233 | **198** |

Per-commit churn is **lower** now, not higher. The hypothesis "the factory merges more, so churn
grew" is measurably false and is not a premise of this SPEC.

### B.4 — Refresh has no attachment point in any pipeline

`grep -rln "graph stamp" .claude/` returns nothing. No hook, skill, workflow, or CI step refreshes
the stamp. The refresh trigger is a human noticing the red and running
`moai graph stamp codemaps`.

### B.5 — Described-worthy contribution per integration

Distribution over the last 30 first-parent integrations
(`git log --first-parent -30 --name-only --pretty=format:"===%h" -- internal cmd pkg`, filtered by
the §B.2 predicate):

| Statistic | Value |
|---|---|
| integrations contributing 0 | 12 of 30 |
| median | 2 |
| p90 | 11 |
| maximum | 29 (`6786c3fa4` — the SPEC-V3R6-GRAPH-FRESHNESS-001 delivery itself) |
| mean | 3.9 |

## §C. Requirements (GEARS)

**REQ-GFC-001 (Ubiquitous).** The codemaps staleness metric shall count only files that a curated
architecture document could describe.

**REQ-GFC-002 (Ubiquitous).** The described-worthy predicate shall admit a path only when it ends
in `.go`, and shall reject it when the name ends in `_test.go`, when any path segment equals
`testdata`, or when the path lies under a configured payload-exclusion prefix whose default set is
`internal/template/templates`.

**REQ-GFC-003 (Ubiquitous).** The clean-stamp endpoint diff and the dirty-stamp content
fingerprint shall admit files through one shared predicate function, so that a file excluded from
the count is also excluded from the fingerprint.

**REQ-GFC-004 (Where the project configures `gate.graph_freshness`).** Where a project supplies
`gate.graph_freshness.described_exclude_prefixes`, the checker shall use that list in place of the
default payload-exclusion set, and where the key is absent the default set shall apply.

**REQ-GFC-005 (Ubiquitous).** The `CodemapsChangedFiles` threshold shall be re-derived from a
measurement taken at implementation time against the corrected predicate, and the run-phase
evidence shall record the command, its output, and the derivation.

**REQ-GFC-006 (Unwanted).** The threshold shall not be raised in order to make a failing check
pass; a raise is admissible only as the recorded outcome of REQ-GFC-005's derivation.

**REQ-GFC-007 (When the codemaps layer is measured).** When the checker measures the codemaps
layer, it shall report both the cumulative count since the stamp and the contribution of the
change under judgment, so that a reader can tell an inherited red from an originated one.

**REQ-GFC-008 (When the codemaps layer is stale).** When the codemaps layer reports a stale
verdict, the failure output shall name the described-worthy paths driving the count, bounded to a
readable maximum, rather than the count alone.

**REQ-GFC-009 (Unwanted).** No unconditional automatic restamp shall be introduced into any merge,
push, or release pipeline; the stamp refresh obligation remains attached to the act of
regenerating the codemaps documents.

**REQ-GFC-010 (Ubiquitous).** The `--json` report shall carry the per-change contribution and the
driving-path list as machine-readable fields alongside the existing per-layer row.

**REQ-GFC-011 (Ubiquitous).** The absent-verdict semantics of the existing checker — missing
directory, missing or unparseable provenance, invalid described root, not-comparable stamped
commit — shall be preserved unchanged; an unjudgeable layer stays unjudgeable and never becomes
fresh.

**REQ-GFC-012 (Where the project distributes templates).** Where a configuration key or default is
added under `.moai/config/sections/gate.yaml`, the corresponding template source under
`internal/template/templates/` shall carry the same key, neutral of this repository's internal
state.

## §D. Judgments

The card asks for three judgments. Each is decided here with its evidence and its rejected
alternatives.

### D.1 — (a) Should the metric distinguish described-worthy files? **Yes.**

Ground: §B.2. Applying the predicate to the streak's own cumulative window turns 65 into 2. The
three-integration red streak was not a signal about the architecture documents at all; it was 42
rule-test fixtures and 7 template YAMLs crossing a threshold calibrated on the assumption that
counted files are described files.

**Where the predicate belongs: in the metric, not in `described_roots`.** `described_roots` is a
list of directory prefixes (`internal`, `cmd`, `pkg` — `mx.DefaultDescribedRoots`), and the noise
is *interleaved inside* those roots: 397 tracked files sit under a `testdata` segment scattered
across arbitrary packages, and `internal/template/templates` is a payload subtree of a described
package. A prefix-inclusion list cannot express "everything under `internal` except any
`testdata` segment". Narrowing the roots would also silently change
`mx.AggregateDescribedFingerprint` and `internal/graph/symbol`, which consume the same list for
unrelated purposes.

**Consistency obligation (REQ-GFC-003).** Applying the predicate only to the endpoint diff would
leave the dirty-stamp path hashing every file, so a `testdata` edit would flip a dirty-stamped
tree stale while leaving a clean-stamped tree untouched — the same tree judged two ways depending
on how it was stamped. One shared predicate closes that.

**Rejected — narrow `described_roots` instead.** Rejected on the expressiveness measurement above,
and on the shared-consumer coupling.

**Rejected — add a citation-existence axis instead** (the shape t304 raises). That axis checks
whether documents cite real packages; it is orthogonal to whether the *count* is inflated, and it
would not have prevented any of the three failures. It belongs to t304 — see §F.

### D.2 — (b) Is 40 still the right threshold? **No, and it cannot be carried over.**

Ground: §B.3 and §B.5. Two facts settle it.

First, 40 was calibrated against the inflated metric, at 13.7 files/commit (137 over 10 commits).
Under the corrected predicate the same tree yields ~1.0 described-worthy files/commit (51 over 50
commits). The metric shrinks by roughly 13×; a threshold carried across that change is not the
same threshold.

Second, the original calibration stated two intents at once — *"routine small PRs pass"* and
*"accumulated drift reds"* — and under the old metric the noise floor was doing the work of the
first. Scaling 40 by the shrink factor gives ≈3, which fails the first intent outright: §B.5 shows
a single integration contributing 11 described-worthy files at p90 and 29 at maximum. A threshold
of 3 would convert the drift gate into a per-change regeneration mandate.

**Re-derivation, not rescaling.** The threshold must sit above the single-integration p90 so no
routine integration reds alone, and low enough that a handful of integrations accumulate past it.
With p90 = 11 and mean = 3.9 per integration, **15** satisfies both: it clears the p90 by a
margin, and it fires after roughly four integrations of mean accumulation. The maximum observed
single integration (29) would red on its own — correctly, since 29 described-worthy files is an
architectural change that should carry regeneration.

15 is this SPEC's proposed value, and REQ-GFC-005 requires the run phase to re-measure the
distribution and record the derivation rather than adopt the number on this SPEC's authority.

**A threshold change alone does not address inheritance** — see D.3.

### D.3 — (c) Where should stamp refresh attach in the merge pipeline? **Nowhere. The framing is the trap.**

Ground: §B.4, plus two hazards already recorded in the queue.

The question presupposes that the defect is a missing refresh. It is not. Restamping is what is
already being done by hand, and the card's own constraint names why it is not the remedy: it
clears the red and leaves the inaccuracy. **An automatic refresh is that same act, performed
faster and without a human ever seeing the signal it erases.** It converts a visible, intermittent
red into a permanently green gate that reports nothing. Rejected.

The second hazard is orphaning. t291 / t292 record that a stamp naming a branch-local HEAD is
stranded by a squash merge, leaving the layer not-comparable. `SPEC-STAMP-REACHABILITY-001`
already delivered the mitigation (`moai graph stamp codemaps --commit "$(git merge-base HEAD
origin/main)"`) and records the tension explicitly: a merge-base stamp is behind the branch tip by
construction, so it *starts* with a nonzero diff. Any auto-refresh attached in a pipeline must
pick one of these poisons — a HEAD stamp that a squash may strand, or a merge-base stamp that is
stale on arrival. (This repository's `develop` integrations are true merge commits — verified,
`git rev-list --parents -n1` returns three fields for all three failing integrations — so the
orphan hazard bites at the `develop` → `main` release boundary rather than at lane integration.)

**What actually stops the repetition** is D.1. Once the metric counts only described-worthy files,
a threshold crossing means real undescribed architectural surface exists — and at that point
inheritance stops being a defect and becomes the intended behaviour. The documents *are* stale,
and they stay stale for every later integration until they are regenerated. Inheritance is
pathological only while the red is spurious.

What is genuinely missing is not automation but **attribution**: a lane that inherits a red today
cannot tell from the output whether the drift is its own. REQ-GFC-007 and REQ-GFC-008 supply that
— the cumulative count, the change's own contribution, and the paths driving it.

So the refresh obligation stays where `moai graph stamp codemaps` already documents it: the last
step of a codemaps regeneration. This SPEC adds no pipeline step and removes none.

**Rejected — unconditional post-merge restamp on `develop`.** Erases the drift signal entirely;
restamp-as-a-service.

**Rejected — conditional auto-restamp when the merged change touched
`.moai/project/codemaps/`.** This is already the contract (`graph stamp` is documented as the last
step of regeneration); automating it adds a pipeline dependency for no behavioural change, and it
would silently stamp regenerations that a human abandoned mid-way.

**Rejected — replace the cumulative metric with a purely per-change one.** It would eliminate
inheritance structurally, but it also eliminates the accumulated-drift signal that is the gate's
entire purpose: N integrations of 3 undescribed files each would each pass, and the documents
would rot without a single red. Per-change contribution is therefore *reported* (REQ-GFC-007), not
*gated on*.

## §E. Exclusions

### Out of Scope — codemaps document accuracy

- Filling in described surface that t197 and t228 added and the six documents do not mention. Owned by **t311**.
- Removing descriptions of six packages that do not exist in the tree. Owned by **t304**.
- Any regeneration of `.moai/project/codemaps/*.md`. This SPEC changes the metric that judges those documents; it does not edit them.

### Out of Scope — stamp reachability

- The squash-orphan structural fix and its live instance repair. Owned by **t291** / **t292** and delivered by `SPEC-STAMP-REACHABILITY-001`. §D.3 consumes that SPEC's recorded tension as evidence and adds nothing to it.

### Out of Scope — restamping as a remedy

- Proposing, scripting, or scheduling a restamp in any form. Explicitly excluded by the card; §D.3 records the rejection with its reasoning.

### Out of Scope — CI trigger topology

- `graph-freshness.yml`'s unfiltered `pull_request` trigger and its double-firing risk under git-flow. Owned by **t294**.

### Out of Scope — other gated layers

- The `mx-index` and `edges` layers, their thresholds, and their fingerprints. Both are untracked runtime artifacts bootstrapped in CI before the check runs; neither participates in the failure mode in §B.1.

## §F. Ordering Basis — t322, t311, t304

The card asks this SPEC to supply the evidence the operator needs to order the three, not to decide
the order.

**Write-surface disjointness (measured).** t322 writes `internal/graph`, `internal/mx`,
`internal/config`, `.moai/config/sections/gate.yaml`, and the corresponding template mirror. t311
and t304 both write the six files under `.moai/project/codemaps/`. The intersection is empty, so
t322 raises no write conflict with either and imposes no ordering requirement of its own.

**One exception, and it is real.** t304's scope item (3) proposes *"adding a citation-existence
axis to `graph check`"*. That lands in `internal/graph/check.go`, the same file REQ-GFC-001 through
REQ-GFC-003 modify. If t304 exercises that item, it collides with t322 and the two must be
ordered — t322 first, since the citation axis is additive to a corrected metric but would have to
be rewritten on top of an uncorrected one.

**t322 benefits from landing first, on signal grounds rather than dependency grounds.** While the
metric remains inflated, `develop` keeps producing reds driven by fixtures. Any genuine red that
t311's or t304's regeneration work produces is indistinguishable from that noise, and the operator
cannot use the gate to verify either card's outcome.

**What t322 does not resolve.** The t311 ↔ t304 conflict is untouched: both regenerate the same six
documents, and the later run overwrites the earlier. That ordering — or the decision to merge them
into a single regeneration — remains exactly as the queue records it, an operator call.

**One caution about verifying t311 / t304.** A completion claim of the form "graph check reports
fresh" is satisfiable by a restamp under both the current metric and the corrected one, because
neither metric reads document content. Those two cards need a content-level acceptance criterion of
their own; t322 does not supply one and does not claim to.

## §G. Open Questions (deliberately left to the operator)

1. **Should an inherited but genuine red block an unrelated lane?** Once D.1 lands, a threshold
   crossing means real undescribed surface exists, and every later integration will inherit it
   until regeneration. Whether that should hard-block unrelated work, or degrade to advisory for
   changes contributing zero described-worthy files, is a policy call about how much documentation
   debt is tolerable. This SPEC gates on the cumulative count in both cases and supplies the
   attribution (REQ-GFC-007) either policy would need.

2. **Should `internal/template/templates/**` be excluded from the predicate?** Decided yes in
   REQ-GFC-002, on the measurement that the curated documents cite the tree once as a bare
   directory and never per file. The residual: a change to the *set of subdirectories* under that
   payload tree arguably is described surface, and a file-count predicate cannot express that
   distinction. The exclusion is configurable (REQ-GFC-004) so the operator can reverse it without
   a code change.

3. **Does t304's citation-existence axis fold into this SPEC or stay in t304?** §F establishes the
   file-level collision. Folding it here would make one coherent change to `check.go`; keeping it
   in t304 keeps this SPEC's scope to the cadence defect. Not decided.

## §H. Cross-References

- `SPEC-V3R6-GRAPH-FRESHNESS-001` — the gate this SPEC corrects; source of the threshold calibration record in §B.3.
- `SPEC-V3R6-GRAPH-FRESHNESS-002` — sibling freshness SPEC.
- `SPEC-STAMP-REACHABILITY-001` — the orphan guard and `--commit` mode; source of the merge-base tension in §D.3.
- Card t322 (this SPEC), t311, t304, t291, t292, t294 — queue entries cited in §E and §F.
- Evidence path: `.moai/reports/t322/`.
