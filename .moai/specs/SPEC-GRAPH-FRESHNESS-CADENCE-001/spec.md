---
id: SPEC-GRAPH-FRESHNESS-CADENCE-001
title: "Graph-freshness cadence: described-worthy metric predicate, threshold re-derivation, and non-inheriting failure attribution"
version: "0.2.3"
status: completed
created: 2026-08-27
updated: 2026-09-02
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/graph, internal/mx"
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
| 0.2.3 | 2026-09-02 | Un-reversal of the v0.2.1 `treeDirty` deferral (§D.1 and the §E residual entry), closed by card t327: `treeDirty` (`internal/mx/provenance.go`) now applies the same REQ-GFC-002 described-worthy predicate as the codemaps fingerprint (`IsDescribedWorthy`) to `git status --porcelain -z --untracked-files=all` output, so a testdata-only-dirty tree is no longer refused the `--commit` merge-base anchor delivered by `SPEC-STAMP-REACHABILITY-001`. Repair is test-covered (`TestStampCodemaps_ExplicitCommitAllowsTestdataOnlyDirty`, `TestStampCodemaps_DefaultPathTestdataOnlyDirtyRecordsCommit`, `TestStampCodemaps_ExplicitCommitRejectsUntrackedDescribedSource`), with RED observed before the fix; the stale `:225` line reference is corrected to `:241`. | manager-spec |
| 0.2.2 | 2026-08-28 | Deferred citation refresh executed at run-phase close, against the delivered tree at `8b11bbba1` — the one-shot the M1 drift record scheduled (`progress.md` §E.2, `#### Coordinate drift caused by M1`) after M1 and M3 each moved source coordinates. Line numbers only: no requirement, judgment, scope decision or acceptance criterion is altered. Every live citation in `spec.md`, `plan.md` and `progress.md` §E.1 was resolved against the construct its own sentence names, and the resolution output is recorded in `progress.md` §E.1. Two citations also carried a **shape** change and their prose was corrected with them, because renumbering alone would have produced a citation that resolves to a valid line and describes something else: `baseProvenance` gained a `fingerprint fingerprintFunc` parameter, so its former `aggregateFingerprint` call is now `fingerprint(projectRoot, describedRoots)`; and `aggregateFingerprint`'s walk moved into `aggregateFingerprintPred`, where the `admit` predicate is threaded. Deliberately **not** refreshed: the `provenance.go:196` quotation in the v0.2.1 row below, together with the `:208` / `:219` values it names as the correction. Those three coordinates are the *subject* of audit finding N2 rather than addresses into the tree, and refreshing them would erase the correction N2 records. The `provenance.go:196` figure therefore still reads against the plan-phase tree by design. `acceptance.md` was checked and carries no line-number citation at all. | manager-spec |
| 0.2.1 | 2026-08-27 | Remediation of plan audit iter-2 PASS-WITH-DEBT 0.895 (`.moai/reports/t322/plan-audit-iter2.md`) — the Tier M iteration ceiling, so these are applied directly rather than as a plan-phase re-entry. **N1**: §D.2's cadence figure rested on `6786c3fa4`, the self-referential SPEC-V3R6-GRAPH-FRESHNESS-001 delivery that §B.5 already disowns and that contributes 29 of the union's 49 at the crossing. The cumulative walk was re-measured here both ways — 40-crossing at integration **10** with it, **16** without (15-crossing unaffected at 5, it precedes the outlier) — both recorded in §B.6, and §D.2 / §G Q4 restated on the outlier-excluded ≈1.6 days. AC-GFC-006 now requires the run-phase Axis 2 to carry the same counterfactual. The §D.2 reversal is untouched: it rests on §B.6 consequence 1 (the streak's corrected cumulative is 2, so the threshold is not load-bearing), which the figure does not affect. **N2**: `baseProvenance` cited at `provenance.go:196` in three places — `:196` is inside `ResolveCommit`; corrected to `:208` (declaration) and `:219` (the `aggregateFingerprint` call) per site. **N3**: counts left stale by the D8 withdrawal recounted — twelve live acceptance criteria, four §G questions, 11 live requirements. **O2** accepted: `treeDirty` (`provenance.go:225`) given an explicit disposition — **examined and deferred**, with the accepted residual stated (a tree dirty only under `testdata` is still refused the `--commit` merge-base anchor) so a later reader can tell the consumer was judged rather than missed (§D.1, §E). **O4** accepted: the premise that makes the unpaired mx-scan / graph-build producers safe — `check.go:222` is the only non-test `ContentFingerprint` comparator, so nothing compares what those producers write — is stated as a §D.1 body paragraph carrying its own obligation (adding a comparing consumer requires re-checking REQ-GFC-003's pairing), because that failure mode is silent and would otherwise take this SPEC's safety argument with it unrecorded. Deliberately **not** promoted to a requirement or an AC. §G Q4 additionally carries a disclaimer that the cadence figure is not the reason 40 is retained — the retention rests on the corrected cumulative of 2 — so the operator cannot mistake the number for the argument. | manager-spec |
| 0.2.0 | 2026-08-27 | Remediation of plan audit PASS-WITH-DEBT 0.82 (`.moai/reports/t322/plan-audit.md`), four blocking defects. **D1**: the predicate is no longer applied inside `aggregateFingerprint` — re-measured, `dirFingerprint` (`internal/graph/meta.go:67`) hashes `.moai/project/codemaps`, `.moai/specs`, `.moai/reports`, of which the first two contain **zero** `.go` files, so a `.go`-only filter there collapses both to the same empty-entry constant `e3b0c442…` and permanently greens the edges layer. Predicate relocated to the codemaps call sites; a second collateral consumer the audit did not name (`baseProvenance`, `internal/mx/provenance.go:240`, shared by the codemaps-gen / mx-scan / graph-build stamp writers) is handled by REQ-GFC-003's producer/consumer pairing. **D2**: cumulative-crossing cadence measured and added to §B.5; §D.2 reversed — **40 is retained**, re-justified on the integration axis. **D3**: AC-GFC-003 now decided through the built artifact. **D4**: AC-GFC-005 extended from four absent branches to all seven. **D5-D7** applied (p90 corrected to 9 with the convention named; `DefaultDescribedRoots` consumer citation corrected to five call sites; AC-GFC-007's search space bounded). **D8** accepted: `internal/template/templates` holds 0 tracked `.go` files and the prefix rule removes nothing the `.go` rule does not — the clause, REQ-GFC-004, REQ-GFC-012, milestone M4 and their criteria are withdrawn. | manager-spec |
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
**3,628** tracked files, of which **998** are production Go outside `testdata/`. 72% of the surface
the metric counts is not architectural surface.

And the curated documents corroborate it. Of the **80** distinct `internal|pkg|cmd` paths cited
across `.moai/project/codemaps/*.md`, **0** contain a `testdata` segment, and the template payload
tree is cited exactly once — as the bare directory `internal/template/templates/`, never per file.

The counterfactual is decisive. Applying a described-worthy predicate (`.go`, excluding `_test.go`
and any `testdata` path segment) to the same cumulative window that produced 65:

```
git diff --name-only 9326b5478d… 48eb945df -- internal cmd pkg \
  | grep '\.go$' | grep -v '_test\.go$' | grep -v '/testdata/' | grep -c .
→ 2
```

The template payload tree needs no rule of its own. `git ls-files internal/template/templates |
grep -c '\.go$'` → **0**: it holds no tracked Go file, so the `.go` rule already removes all of it,
t228's 7 rule YAMLs included. Adding `internal/template/templates` as an exclusion prefix changes
this window's result from 2 to 2 — it excludes nothing measured, and is therefore not part of the
predicate (v0.2.0, audit D8).

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
| p90 (nearest-rank, `ceil(0.9·30)` = rank 27) | **9** |
| maximum | 29 (`6786c3fa4` — the SPEC-V3R6-GRAPH-FRESHNESS-001 delivery itself) |
| mean | 3.9 |

v0.1.0 reported p90 = 11; that value sits at rank 28, the 93rd percentile. Corrected here, with the
convention named (audit D5).

### B.6 — Cumulative-crossing cadence: the axis the gate actually runs on

Per-integration contribution is not the axis that decides how often the gate is red. The metric is
cumulative since the stamp, so the deciding question is how many integrations elapse before the
running union crosses the threshold. Measured over the same 30-integration window, oldest first,
accumulating the distinct described-worthy union
(`git log --first-parent -30 --reverse --name-only --pretty=format:"===%h" -- internal cmd pkg`,
filtered by the predicate, unioned):

| Integrations since a hypothetical restamp | Distinct described-worthy union |
|---|---|
| 3 | 11 |
| 5 | **17** ← crosses 15 |
| 8 | 21 |
| 10 | **49** ← crosses 40 |
| 30 | 86 |

**The 40-crossing above is produced by a single integration, and that integration is the one §B.5
disowns.** The union goes `…8 → 21, 9 → 21, 10 → 49`: the jump from 21 to 49 is `6786c3fa4` alone,
the SPEC-V3R6-GRAPH-FRESHNESS-001 delivery that introduced this gate — a self-referential outlier
contributing 29 described-worthy files
(`git diff --name-only 6786c3fa4^1..6786c3fa4 -- internal cmd pkg | grep '\.go$' | grep -v
'_test\.go$' | grep -v '/testdata/' | grep -c .` → **29**). A cadence figure that rests on it
describes how often the gate reds *when the gate itself is being built*, which is not the cadence
the operator is being asked about. So the walk was re-run with that integration removed from the
window (29 integrations, same predicate, same union accumulation):

| Threshold | Crossing with `6786c3fa4` | Crossing without it |
|---|---|---|
| 15 | integration 5 (union 17) | integration 5 (union 17) |
| **40** | **integration 10** (union 49) | **integration 16** (union 43) |

The 15-crossing is unaffected — it happens before the outlier. The 40-crossing moves from 10 to 16,
a 60% difference in the red rate, in the direction that makes the gate *less* red. (Convention note:
"without" removes the outlier from the window entirely. Retaining it as a zero-contribution slot
instead puts the same crossing at slot 17; it is one event under two counting conventions, not two
measurements.)

The window `39c677f47 … 48eb945df` spans 2026-08-25 to 2026-08-27 — 30 integrations in three days,
≈10/day (≈9.7/day with the outlier removed). **The operative figures are the outlier-excluded ones**:
the corrected metric crosses **40 in about 1.6 days** and **15 in about half a day**, and it stays
red for every later integration until the six documents are regenerated by hand. The
with-outlier row is retained above only so the counterfactual is auditable.

Two consequences follow, and §D.2 is decided on them:

1. **The threshold is not load-bearing for the defect this SPEC exists to fix.** The streak's
   corrected cumulative is 2, which passes under 15 and under 40 alike. §D.1's predicate alone
   stops the reported streak.
2. **"40 became ~13× laxer" is true only on the commit axis.** On the integration axis — the one CI
   runs on — corrected-40 reds within roughly a day and a half even after the self-referential
   outlier is removed. It is not lax there.

## §C. Requirements (GEARS)

**REQ-GFC-001 (Ubiquitous).** The codemaps staleness metric shall count only files that a curated
architecture document could describe.

**REQ-GFC-002 (Ubiquitous).** The described-worthy predicate shall admit a path only when it ends
in `.go`, and shall reject it when the name ends in `_test.go` or when any path segment equals
`testdata`.

**REQ-GFC-003 (Ubiquitous).** The predicate shall be applied at the codemaps call sites only, and
the producer and consumer of the codemaps content fingerprint shall apply the same predicate, so
that a dirty-stamped tree is judged against a fingerprint computed the same way it was stamped.

**REQ-GFC-003a (Unwanted).** The predicate shall not be applied inside `aggregateFingerprint`, and
the fingerprint that `dirFingerprint` computes for the edges layer's four source sets shall be
byte-identical before and after this change.

**REQ-GFC-004** — *withdrawn at v0.2.0 (audit D8).* A configurable payload-exclusion list was
specified before any file was shown to need it; `internal/template/templates` holds 0 tracked `.go`
files (§B.2), so the list would have excluded nothing. Recorded rather than renumbered, to keep the
surviving REQ identifiers stable.

**REQ-GFC-005 (Ubiquitous).** The `CodemapsChangedFiles` threshold shall be confirmed or corrected
from measurements taken at implementation time on **both** axes — the per-integration contribution
distribution and the cumulative-crossing cadence of §B.6 — and the run-phase evidence shall record
each command, its output, the percentile convention used, and the derivation.

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

**REQ-GFC-012** — *withdrawn at v0.2.0 (audit D8).* It was conditional on adding a configuration
key; with REQ-GFC-004 withdrawn, no key is added and no template mirror is required. Recorded
rather than renumbered.

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
across arbitrary packages. A prefix-inclusion list cannot express "everything under `internal`
except any `testdata` segment". Narrowing the roots would also change five call sites that consume
`mx.DefaultDescribedRoots` for unrelated purposes — `internal/graph/check.go:200`,
`internal/graph/symbol/symbol.go:99`, and the three stamp writers at
`internal/mx/provenance.go:269,285,296` (codemaps-gen, mx-scan, graph-build). Three of the five are
stamp writers, which makes the placement argument stronger than v0.1.0 stated (audit D6).

**Consistency obligation, and its blast radius (REQ-GFC-003 / REQ-GFC-003a).** Applying the
predicate only to the endpoint diff would leave the dirty-stamp path hashing every file, so a
`testdata` edit would flip a dirty-stamped tree stale while leaving a clean-stamped tree untouched.
But applying it *inside* `aggregateFingerprint` — v0.1.0's instruction — is worse, and this is the
audit's critical finding (D1), re-measured here.

`dirFingerprint` (`internal/graph/meta.go:67`) reaches `mx.AggregateDescribedFingerprint` for three
of the edges layer's four source sets: `.moai/project/codemaps`, `.moai/specs`, `.moai/reports`.
Measured with `find` over the three, the first two contain **zero** `.go` files and the third
contains 2 strays in untracked report directories. Under a `.go`-only filter applied there, the
codemaps and specs source sets both collapse to the empty-entry hash
`e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` — the same constant, and a
permanent one — so `compareSourceFingerprints` never reports movement and the edges layer reports
fresh forever. That is precisely the permanently-green gate §D.3 rejects on principle, inflicted on
a layer §E declares out of scope.

A second collateral consumer, which the audit did not name, is `baseProvenance`
(`internal/mx/provenance.go:240`, its fingerprint call at `:251`): it computes the dirty
`ContentFingerprint` for **all three** stamp writers. Filtering only the checker at
`check.go:216` would leave every dirty codemaps stamp permanently mismatched against the checker
reading it.

So the predicate goes to the codemaps call sites — the checker at `check.go:216` and the codemaps
stamp writer — through a predicate-bearing variant, leaving `AggregateDescribedFingerprint`'s
contract intact for `meta.go` and leaving the mx-scan and graph-build stamps unchanged.

**The unpaired producers are safe on a premise, and the premise is stated here because it can
break silently.** `StampMXScan` and `StampEdges` go on writing an unfiltered `ContentFingerprint`
while the codemaps writer is filtered — a deliberate asymmetry, and one this SPEC leaves in place.
It is safe for exactly one reason: **nothing ever compares what those two producers write.**
Measured — `grep -rn --include='*.go' 'ContentFingerprint' . | grep -v _test.go` yields a single
comparing consumer, `checkCodemaps` (`internal/graph/check.go:216` recomputes, `:222` compares),
plus one display-only reader, `Provenance.Describe` (`internal/mx/provenance.go:312`). `checkMXIndex`
and the edges check never read the field at all.

So the producer/consumer pairing REQ-GFC-003 requires is sufficient *given that premise*, and it
stops being sufficient the moment anyone adds a comparing consumer against the mx-index or edges
layer. **Adding such a consumer therefore obliges a re-check of this SPEC's pairing** — the
unfiltered producers would have to be filtered with it, or the new comparison would read a
fingerprint computed over a different file set than the one it is judging. That failure is silent:
nothing in the build or the test suite would flag it, and what breaks is this argument, which is why
the argument is written down here rather than left as a footnote.

**One described-roots consumer was examined and deliberately deferred at v0.2.1 — and that
deferral is now closed by card t327: `treeDirty`.** `treeDirty`
(`internal/mx/provenance.go:241`) decides, from `git status --porcelain` over the
described roots, whether `baseProvenance` takes the dirty-fingerprint branch or the commit-anchor
branch. At v0.2.1 it consulted no predicate, and leaving it unfiltered was an explicit decision,
not an oversight — it was examined during the iter-2 remediation and deferred (a false "dirty"
over-reports rather than under-reports, and an anchor-mode selector was judged to belong to
`SPEC-STAMP-REACHABILITY-001`'s axis, not to the metric this SPEC corrects).

**The residual that decision accepted is now repaired.** As of card t327, `treeDirty` applies the
same REQ-GFC-002 described-worthy predicate as the codemaps fingerprint (`IsDescribedWorthy` —
`.go` files, rejecting `_test.go` names and any `testdata` path segment) to
`git status --porcelain -z --untracked-files=all` output, so the anchor gate and the codemaps
fingerprint judge the tree by the same rule. The predicate-blind reading is gone: a tree dirty
**only** under `testdata` — carrying no described-worthy change at all, by this SPEC's own thesis
— is no longer classified dirty, and is therefore no longer refused the `--commit` merge-base
anchor that `SPEC-STAMP-REACHABILITY-001` delivered and that §D.3 cites as the orphan mitigation.
`--untracked-files=all` keeps a new untracked described source from hiding under a collapsed
`dir/` record, and the `-z` form parses without whitespace splitting. The repair is test-covered:
`TestStampCodemaps_ExplicitCommitAllowsTestdataOnlyDirty` and
`TestStampCodemaps_DefaultPathTestdataOnlyDirtyRecordsCommit` (a testdata-only-dirty tree now
records the commit anchor), plus `TestStampCodemaps_ExplicitCommitRejectsUntrackedDescribedSource`
guarding the `--untracked-files=all` half; the explicit-commit case was observed RED before the
fix (rejected with "described sources carry uncommitted changes"). The §E residual entry is
un-reversed to match this closure.

**Rejected — narrow `described_roots` instead.** Rejected on the expressiveness measurement above,
and on the shared-consumer coupling.

**Rejected — add a citation-existence axis instead** (the shape t304 raises). That axis checks
whether documents cite real packages; it is orthogonal to whether the *count* is inflated, and it
would not have prevented any of the three failures. It belongs to t304 — see §F.

### D.2 — (b) Is 40 still the right threshold? **Yes. Retain it, on a corrected justification.**

*Reversed at v0.2.0. v0.1.0 answered "no, re-derive to 15" from the per-integration axis alone; the
audit's D2 supplied the cumulative axis, which changes the answer.*

Ground: §B.3, §B.5, §B.6.

**Why v0.1.0's argument does not survive.** It rested on "40 was calibrated at 13.7 files/commit,
the corrected metric runs at ~1.0, so the metric shrank ~13× and the threshold cannot carry over".
That is true on the **commit** axis and irrelevant on the **integration** axis, which is the one CI
runs on. §B.6 measures it: corrected-40 crosses at **16 integrations** once the self-referential
`6786c3fa4` is excluded — ≈1.6 days at the observed ≈10 integrations/day. (With that outlier
included the crossing is 10, ≈one day; §B.6 records both and takes the excluded figure as
operative.) 40 is not lax on the axis that decides how often the gate is red — not by the margin
the with-outlier figure suggests, but a gate that reds every day and a half is not a gate nobody
notices.

**Why 15 is worse, not better.** §B.6 puts corrected-15 at 5 integrations — roughly twice per day —
and the red then persists for every subsequent integration until six documents are regenerated by
hand, a regeneration §D.3 deliberately refuses to automate and which t311 and t304 have not been
able to schedule. Against retained-40's ≈1.6 days that is roughly a **threefold** increase in red
rate, and tripling the red rate of a gate whose only exit is manual work nobody has scheduled is a
reliable way to get the gate ignored.

**Why the threshold should not move in this SPEC at all.** §B.6's first consequence is decisive:
the streak's corrected cumulative is **2**, which passes under 15 and under 40 alike. §D.1's
predicate alone stops the reported streak. Changing the threshold in the same SPEC would alter two
variables at once and make the outcome unattributable — if the gate behaved unexpectedly afterwards,
neither the predicate nor the threshold could be cleared. So the threshold is held fixed as the
control.

**Intended red frequency, stated.** At the observed integration rate, retained-40 under the
corrected predicate is expected to red roughly **once every day and a half of factory activity**
(§B.6's outlier-excluded crossing: 16 integrations at ≈10/day). Each red should then be a true
statement that undescribed architectural surface has accumulated. Whether that frequency is
tolerable is a debt-tolerance question, not a metric question — recorded as §G Q4. The figure is a
single-window measurement over three days and is offered as an order of magnitude, not a rate
constant; REQ-GFC-005 requires the run phase to re-measure it.

REQ-GFC-005 requires the run phase to confirm 40 against re-measurements on **both** axes and record
the derivation; a correction is admissible if the measurement supports one, but it must be measured,
not carried from this section.

**No threshold value addresses inheritance** — see D.3.

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

### Out of Scope — threshold sensitivity change

- Moving `CodemapsChangedFiles` away from 40. §D.2 retains it as the control variable so the predicate change is attributable on its own. A sensitivity change belongs to a follow-up taken once a regeneration cadence exists.
- Any configuration surface for the predicate. Withdrawn at v0.2.0 with REQ-GFC-004 / REQ-GFC-012 (§G Q2).

### Out of Scope — the dirty/clean anchor-mode selector

- `treeDirty` (`internal/mx/provenance.go:241`) **no longer stays predicate-blind — the v0.2.1 deferral is closed by card t327.** Provenance retained: it was examined at v0.2.1 and deliberately deferred — not overlooked. It decides `baseProvenance`'s dirty-vs-commit anchor branch, and as of t327 it applies the same REQ-GFC-002 predicate as the codemaps fingerprint (`IsDescribedWorthy`: `.go`, rejecting `_test.go` and any `testdata` path segment) to `git status --porcelain -z --untracked-files=all` output — `--untracked-files=all` so a new untracked described source cannot hide under a collapsed `dir/` record, `-z` for the parse. The former accepted residual, stated in full in §D.1 — a tree dirty only under `testdata` still refused the `--commit` merge-base anchor delivered by `SPEC-STAMP-REACHABILITY-001` — is repaired: a testdata-only-dirty tree now takes the named-commit anchor. §D.1 carries the repair notes and the covering tests.

### Out of Scope — the edges layer's fingerprint

- `mx.AggregateDescribedFingerprint`'s contract and `dirFingerprint`'s four source sets. REQ-GFC-003a pins them byte-identical rather than changing them; §D.1 records why touching them would permanently green a layer this SPEC does not own.

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

2. **Should the predicate be configurable at all?** v0.1.0 specified a configurable
   payload-exclusion list; v0.2.0 withdrew it once measurement showed it excludes nothing
   (§B.2, audit D8). The predicate ships fixed. If a project later needs an exclusion, it should be
   added against a measured file, not in advance — deferred rather than decided.

3. **Does t304's citation-existence axis fold into this SPEC or stay in t304?** §F establishes the
   file-level collision. Folding it here would make one coherent change to `check.go`; keeping it
   in t304 keeps this SPEC's scope to the cadence defect. Not decided.

4. **Is roughly one red every day and a half of factory activity tolerable?** §B.6 measures
   corrected-40 crossing at **16 integrations** ≈ 1.6 days at the observed rate, once the
   self-referential `6786c3fa4` is excluded from the window; §D.2 retains 40 on that basis. (The
   with-outlier figure is 10 integrations ≈ one day; it is recorded in §B.6 but is not the figure
   this question is posed on, because it measures the gate's redness during its own construction.)

   **This number is not the reason 40 is retained — do not read it as one.** The decision to retain
   40 rests on independent ground: §B.6 consequence 1, that the streak this SPEC exists to fix
   carries a corrected cumulative of **2**, which passes under 15 and under 40 alike. The threshold
   is therefore not load-bearing for the defect being fixed, and is held fixed as the control so the
   predicate change is attributable on its own. The cadence figure answers a different question —
   how often the retained gate will be red — and the operator can judge that frequency
   intolerable without any of it disturbing the retention decision. (v0.1.0's own history is the
   caution: a threshold argument built on one axis of one window was reversed once already, and at
   v0.2.0 the replacement figure was itself produced by a single integration §B.5 had already
   declared an outlier — a decision input made out of data the document had excluded from its own
   distribution.)

   Each such red will be a true statement that undescribed surface has accumulated, and its only
   exit is a manual regeneration of six documents. If that frequency is judged intolerable, the
   lever is a regeneration cadence (t311 / t304) or the Q1 advisory-degradation policy — not a
   threshold raise, which would only defer the same red.

## §H. Cross-References

- `SPEC-V3R6-GRAPH-FRESHNESS-001` — the gate this SPEC corrects; source of the threshold calibration record in §B.3.
- `SPEC-V3R6-GRAPH-FRESHNESS-002` — sibling freshness SPEC.
- `SPEC-STAMP-REACHABILITY-001` — the orphan guard and `--commit` mode; source of the merge-base tension in §D.3.
- Card t322 (this SPEC), t311, t304, t291, t292, t294 — queue entries cited in §E and §F.
- Evidence path: `.moai/reports/t322/`.
