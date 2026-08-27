# SPEC-GRAPH-FRESHNESS-CADENCE-001 — Implementation Plan

Milestones are ordered by decision-reversibility: the decisions most likely to change on review
lead, and the mechanical work trails.

## §A. Context

The `graph-freshness` job reds on integration after integration because its codemaps metric counts
every changed file under `internal`, `cmd`, `pkg` — fixtures and template payload included — against
a threshold calibrated as though those files were architectural surface. See `spec.md` §B for the
measured baseline and §D for the three adjudicated judgments this plan implements.

Cycle type: **TDD**. Every milestone below has a mechanically observable red state before its green
state, and the metric change is exactly the kind that a reporting-only implementation would silently
fake.

## §B. Known Issues Inherited

- `mx.DefaultDescribedRoots` is consumed by five call sites (`internal/graph/check.go:168`,
  `internal/graph/symbol/symbol.go:99`, and `internal/mx/provenance.go:237,253,264` — the
  codemaps-gen, mx-scan and graph-build stamp writers). This SPEC must not change that list — the
  predicate is a separate axis (spec.md §D.1).
- **`aggregateFingerprint` has four call sites and only one of them may be filtered.** Measured:
  `internal/graph/check.go:181` (codemaps dirty checker — filter), `internal/graph/meta.go:67`
  (`dirFingerprint`, the edges layer's source sets — **must not** filter, REQ-GFC-003a) and
  `internal/mx/provenance.go:219` (the call inside `baseProvenance`, whose declaration is at `:208`;
  shared by all three stamp writers — filter for the codemaps stamp only). spec.md §D.1 records what
  a filter inside the function would do to the edges layer.
- **The codemaps content fingerprint has a producer and a consumer, and they must move together.**
  `baseProvenance` writes it; `check.go:181` computes the current value and `:187` compares it.
  Filtering one without the other leaves every dirty codemaps stamp permanently mismatched. Existing
  dirty codemaps stamps are invalidated by this change; they are transient by construction (only on
  trees with uncommitted described-source changes), so this is accepted rather than migrated.
- **Leaving the mx-scan and graph-build stamps unfiltered rests on a measured premise**: `check.go:187`
  is the only non-test comparator of `ContentFingerprint` in the tree (`provenance.go:280` reads it
  for display only). AC-GFC-004's sole-comparator clause pins it; if the run phase finds a second
  comparator, the pairing in REQ-GFC-003 is no longer sufficient and the milestone stops for a
  judgment rather than proceeding.
- **`treeDirty` (`provenance.go:201`) is deliberately not touched.** It selects `baseProvenance`'s
  anchor branch without consulting the predicate, so a tree dirty only under `testdata` is still
  refused the `--commit` anchor. Recorded as a known residual (spec.md §D.1 / §E), not a work item —
  do not "fix" it in passing.
- The tracked `provenance.json` in this tree names `d2fcecc8b40d1cb…`. It is an ancestor of this
  branch's HEAD (verified `git merge-base --is-ancestor` → rc 0), so the checker is comparable here.
  Nothing in this plan restamps it.

## §C. Pre-flight

```bash
git rev-parse --show-toplevel        # → the t322 worktree, not the primary checkout
git branch --show-current            # → WT-graph-freshness-cadence
go build ./internal/graph/... ./internal/mx/... ./internal/config/...
./bin/moai graph check --json        # baseline verdict, recorded before any edit
```

## §D. Constraints

- **No restamp.** No milestone may run `moai graph stamp codemaps`, and no deliverable may add a
  restamp to a pipeline (REQ-GFC-009).
- **No threshold raise to pass a check** (REQ-GFC-006). The threshold is retained at 40 (`spec.md`
  §D.2); any change is admissible only as the recorded output of M2's two-axis measurement.
- **No full local suite.** Run the affected packages; CI renders the full verdict.
- **No `e2e/` execution.** The operator has a live console and that suite mutates real profiles.
- **No configuration key.** M4 is withdrawn, so no `.moai/config` key is added and Template-First
  has nothing to mirror. Adding one back requires a measured need first (`spec.md` §G Q2).

## §E. Milestones

### M1 — Described-worthy predicate (highest reversibility risk)

The judgment most likely to be revised on review: *which files count*.

- Introduce one exported predicate in `internal/mx` — a single pure function deciding whether a
  repo-relative slash path is described-worthy, per REQ-GFC-002: `.go` suffix, not `_test.go`, no
  path segment equal to `testdata`. No configuration parameter (REQ-GFC-004 withdrawn).
- Apply it in `gitDiffNameCount` (`internal/graph/check.go`) to both the `git diff --name-only`
  branch and the `git ls-files --others` branch.
- Add a predicate-bearing fingerprint variant (e.g. `AggregateDescribedFingerprintFiltered`) and use
  it at `check.go:181` **and** in the codemaps stamp path, so producer and consumer agree
  (REQ-GFC-003). Two mechanisms are open: give `baseProvenance` a fingerprint-function parameter and
  have `StampCodemaps` pass the filtered one, or have `StampCodemaps` recompute `ContentFingerprint`
  after `baseProvenance` returns. Either is acceptable; the first is cleaner, the second smaller.
- **Leave `AggregateDescribedFingerprint` and `aggregateFingerprint` unfiltered** so `meta.go:67`
  and the mx-scan / graph-build stamps are untouched (REQ-GFC-003a).
- Preserve every absent-verdict path unchanged (REQ-GFC-011).

Mutants to kill: (i) a predicate wired into the diff branch only, leaving the untracked branch
unfiltered; (ii) a predicate pushed down into `aggregateFingerprint`, which greens the edges layer
permanently; (iii) a filtered checker with an unfiltered codemaps stamp writer, which reds every
dirty tree. AC-GFC-002, AC-GFC-014 and AC-GFC-004 respectively must fail under each.

### M2 — Threshold confirmation (the value is held fixed unless measurement says otherwise)

`spec.md` §D.2 **retains 40**. M2 confirms that decision against the tree at implementation time
rather than deriving a new value; a correction is admissible only if the measurement supports one.

- Re-measure both axes: the per-integration contribution distribution (`spec.md` §B.5) and the
  cumulative-crossing cadence (`spec.md` §B.6, the `--reverse` union walk). Name the percentile
  convention used — nearest-rank (audit D5).
- Run the cumulative walk **twice** — once over the whole window and once with any self-referential
  or otherwise outlier integration removed — and report both. §B.6's 40-crossing moves from 10 to 16
  on that single exclusion, so a single-walk figure is not a cadence measurement (audit N1).
- State the expected red frequency at the measured integration rate, on the outlier-excluded walk,
  and check it against §D.2's stated intent of roughly one red every day and a half of factory
  activity.
- Record every command, its verbatim output, the convention, and the conclusion in `progress.md`
  §E.2. A threshold left at 40 without that record fails REQ-GFC-005 exactly as a changed one would.

M2 is not load-bearing for the reported defect: the streak's corrected cumulative is 2, which passes
at 15 and at 40 alike. M1 alone stops it.

### M3 — Failure attribution

- Compute the change's own described-worthy contribution alongside the cumulative count
  (REQ-GFC-007). The comparison base is the checkout's first-parent predecessor where one exists;
  where it does not, the field is reported absent rather than fabricated.
- Extend the stale-verdict stderr output to name the driving paths, bounded to a readable maximum
  with an explicit overflow indicator (REQ-GFC-008).
- Carry both as fields on the `--json` report (REQ-GFC-010).

### M4 — *withdrawn at v0.2.0 (audit D8)*

M4 added `gate.graph_freshness.described_exclude_prefixes` and its template mirror. Measured,
`git ls-files internal/template/templates | grep -c '\.go$'` → 0, and the streak counterfactual is 2
with the prefix rule and 2 without it — the configuration surface was justified by no measured file.
Withdrawn with REQ-GFC-004 and REQ-GFC-012. No configuration key is added, so `internal/config`,
`gate.yaml`, and the template tree are untouched by this SPEC; `graphCheckThresholds` keeps its
existing behaviour unchanged.

## §F. Technical Approach

The predicate is a pure function over a repo-relative slash path. It takes no configuration and no
filesystem access, which keeps it testable in isolation and keeps the two call sites
(`git`-driven and `filepath.Walk`-driven) honest about supplying comparable path forms — the walk
already produces `filepath.ToSlash(rel)`, and `git` already emits slash paths, so no normalization
asymmetry is introduced.

Attribution needs a comparison base the checker does not currently hold. Deriving it from `HEAD^1`
is correct for a merge commit and wrong for a linear one; the implementation must therefore treat
"no first parent" as an absent field rather than fall back to a base that silently means something
else.

## §G. Anti-Patterns

- **Restamp-as-remedy**, in any automated form. Rejected with reasoning in `spec.md` §D.3.
- **Moving the threshold in this SPEC.** Two variables changed at once make the outcome
  unattributable; `spec.md` §D.2 holds 40 as the control. Rescaling it by the metric's shrink factor
  is doubly wrong — the shrink is a commit-axis figure, and the scaled value (≈3) fails the
  calibration's own second intent.
- **Predicate in `described_roots`.** Cannot express an interleaved `testdata` exclusion, and
  changes five unrelated call sites, three of them stamp writers.
- **Predicate inside `aggregateFingerprint`.** Permanently greens the edges layer (`spec.md` §D.1).
- **A configuration surface for the predicate.** Withdrawn at v0.2.0; it excluded no measured file.
- **Reporting-only implementation.** A per-change contribution that is computed and displayed but
  never influences the verdict is correct here *by design* (REQ-GFC-007 reports, it does not gate) —
  but a *cumulative* count that is displayed without gating is the mutant.
- **Gating on the per-change count.** Eliminates inheritance and the drift signal together
  (`spec.md` §D.3, third rejection).

## §H. Cross-References

- `spec.md` §B (baseline), §D (judgments), §F (ordering basis), §G (open questions)
- `acceptance.md` (the AC matrix each milestone closes)
- `SPEC-V3R6-GRAPH-FRESHNESS-001`, `SPEC-STAMP-REACHABILITY-001`
- Evidence path: `.moai/reports/t322/`
