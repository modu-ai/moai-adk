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

- `mx.DefaultDescribedRoots` is consumed by three call sites (`internal/graph/check.go`,
  `internal/graph/symbol/symbol.go`, `internal/mx/provenance.go`). This SPEC must not change that
  list — the predicate is a separate axis (spec.md §D.1).
- `AggregateDescribedFingerprint` currently hashes every regular file under the roots. Applying the
  predicate to it (REQ-GFC-003) changes the fingerprint value, which invalidates any **dirty** stamp
  in existence. Dirty stamps are transient by construction (they exist only on trees with
  uncommitted described-source changes), so this is accepted, and M1 records it rather than working
  around it.
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
- **No threshold raise to pass a check** (REQ-GFC-006). A raise is admissible only as the recorded
  output of M2's derivation.
- **No full local suite.** Run the affected packages; CI renders the full verdict.
- **No `e2e/` execution.** The operator has a live console and that suite mutates real profiles.
- **Template-First.** Any `.moai/config` key change lands in `internal/template/templates/` in the
  same commit, neutral of this repository's internal state (REQ-GFC-012).

## §E. Milestones

### M1 — Described-worthy predicate (highest reversibility risk)

The judgment most likely to be revised on review: *which files count*.

- Introduce one exported predicate in `internal/mx` — a single function deciding whether a
  repo-relative path is described-worthy, per REQ-GFC-002: `.go` suffix, not `_test.go`, no path
  segment equal to `testdata`, not under a supplied exclusion prefix.
- Apply it in `gitDiffNameCount` (`internal/graph/check.go`) to both the `git diff --name-only`
  branch and the `git ls-files --others` branch.
- Apply the same function in `aggregateFingerprint` (`internal/mx/provenance.go`) so the clean and
  dirty paths agree (REQ-GFC-003).
- Preserve every absent-verdict path unchanged (REQ-GFC-011).

Mutant to kill: a predicate wired into the diff branch only, leaving the untracked branch or the
fingerprint unfiltered. The acceptance fixture must fail if either is missed.

### M2 — Threshold re-derivation

Depends on M1: the distribution cannot be measured until the predicate exists.

- Re-measure the per-integration described-worthy distribution on the tree at implementation time,
  using the `git log --first-parent` form recorded in `spec.md` §B.5.
- Derive the value from the two stated intents — above the single-integration p90, low enough that a
  small number of integrations accumulate past it. `spec.md` §D.2 proposes **15** from the
  plan-phase measurement (p90 = 11, mean = 3.9); M2 confirms or corrects it against its own
  measurement.
- Record the command, its verbatim output, and the derivation in `progress.md` §E.2. A value adopted
  without that record fails REQ-GFC-005.

### M3 — Failure attribution

- Compute the change's own described-worthy contribution alongside the cumulative count
  (REQ-GFC-007). The comparison base is the checkout's first-parent predecessor where one exists;
  where it does not, the field is reported absent rather than fabricated.
- Extend the stale-verdict stderr output to name the driving paths, bounded to a readable maximum
  with an explicit overflow indicator (REQ-GFC-008).
- Carry both as fields on the `--json` report (REQ-GFC-010).

### M4 — Configuration surface and template mirror (mechanical)

- Add `gate.graph_freshness.described_exclude_prefixes` to the config struct and the gate loader,
  defaulting to the REQ-GFC-002 set when absent (REQ-GFC-004). The existing
  `graphCheckThresholds` malformed-gate.yaml contract (present-but-unparseable → exit 2) is
  preserved.
- Mirror the key into `internal/template/templates/.moai/config/sections/gate.yaml`, then
  `make build`, then verify neutrality.

## §F. Technical Approach

The predicate is a pure function over a repo-relative slash path plus an exclusion-prefix list. It
takes no filesystem access, which keeps it testable in isolation and keeps the two call sites
(`git`-driven and `filepath.Walk`-driven) honest about supplying comparable path forms — the walk
already produces `filepath.ToSlash(rel)`, and `git` already emits slash paths, so no normalization
asymmetry is introduced.

Attribution needs a comparison base the checker does not currently hold. Deriving it from `HEAD^1`
is correct for a merge commit and wrong for a linear one; the implementation must therefore treat
"no first parent" as an absent field rather than fall back to a base that silently means something
else.

## §G. Anti-Patterns

- **Restamp-as-remedy**, in any automated form. Rejected with reasoning in `spec.md` §D.3.
- **Rescaling the threshold** by the metric's shrink factor. `spec.md` §D.2 shows the scaled value
  (≈3) fails the calibration's own second intent.
- **Predicate in `described_roots`.** Cannot express an interleaved `testdata` exclusion, and
  silently changes two unrelated consumers.
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
