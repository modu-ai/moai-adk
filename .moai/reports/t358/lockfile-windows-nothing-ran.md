# `internal/lockfile` runs zero tests on windows-latest

Card t358 · SPEC-CI-TEST-OBSERVABILITY-001 · found by the census on its first
real CI run, 2026-08-30.

**Not investigated. Not fixed.** This file records an observation and stops there.
Card issuance is the operator's call; the lead carries it to that list.

---

## Claim

On `windows-latest`, the package `github.com/modu-ai/moai-adk/internal/lockfile`
executes **no tests at all**. On `ubuntu-latest` and `macos-latest`, in the same
run, it executes tests normally.

The job is not red because of it. Under rc-only reporting this state is
**indistinguishable from a package that ran and passed** — `go test` prints an
`ok`-shaped line either way, and the exit status is unaffected.

## Evidence

One run, three OS legs, dispatched once:
`https://github.com/modu-ai/moai-adk/actions/runs/33308057570`
(event `workflow_dispatch`, branch `WT-ci-test-observability`, head `09dd1bee9`).

Census totals, verbatim from each leg's `Run tests with race detector` step:

| OS | job id | conclusion | packages | passed | skipped | **nothing-ran** | failed | build-failed |
|---|---|---|---|---|---|---|---|---|
| ubuntu-latest | 99248050305 | success | 136 | 19513 | 100 | **3** | 0 | 0 |
| macos-latest | 99248050297 | failure | 136 | 19512 | 91 | **3** | 1 | 0 |
| windows-latest | 99248050298 | failure | 136 | 18854 | 332 | **4** | 146 | 0 |

The difference is one package. The `NOTHING RAN` rows, sorted:

ubuntu and macOS — identical, three rows:

```
NOTHING RAN github.com/modu-ai/moai-adk/cmd/moai
NOTHING RAN github.com/modu-ai/moai-adk/internal/template/scripts
NOTHING RAN github.com/modu-ai/moai-adk/scripts/convert-nextra-to-hextra
```

windows — the same three, plus one:

```
NOTHING RAN github.com/modu-ai/moai-adk/cmd/moai
NOTHING RAN github.com/modu-ai/moai-adk/internal/lockfile
NOTHING RAN github.com/modu-ai/moai-adk/internal/template/scripts
NOTHING RAN github.com/modu-ai/moai-adk/scripts/convert-nextra-to-hextra
```

The other three are expected and were predicted before the run: `plan.md` §D M1
names exactly those three as the zero-test-file baseline for this tree, derived
from `go list -f '{{len .TestGoFiles}} {{len .XTestGoFiles}}'`. A census reporting
fewer than three has a bug; one reporting more has found something. It reported
four on one OS.

## Baseline-attribution

Measured in this run, against head `09dd1bee9`, by the census implementation at
`scripts/ci-census/test-census.sh` as committed on branch
`WT-ci-test-observability`.

The `NOTHING RAN` label is emitted for a `-json` event `{"Action":"skip"}` carrying
**no** `Test` field — a package-level skip, which `go test` emits when a package
contributes no test to the run. It is the same single `Action=="skip"` pass that
produces `SKIPPED TEST` rows, split on the presence of `.Test`
(`spec.md` §C.1, REQ-CTO-006).

No prior CI run of this repository could have produced this observation: before
this card, all four `go test` call sites ran without `-v` or `-json`, so no
per-package or per-test event reached any artifact.

## Gaps — what was NOT observed

- **The cause was not investigated.** Whether this is a build constraint
  (`//go:build !windows` on the test files), a windows-specific `t.Skip` in a
  `TestMain`, a compilation exclusion, or something else — not determined. No
  source file in `internal/lockfile` was read for this report.
- **Whether it is intentional** — not determined. A package deliberately excluded
  from windows would produce exactly this signal, and would be correct behaviour
  reported honestly rather than a defect.
- **Whether coverage was ever claimed for it on windows** — not checked.
- **Single run.** Observed once. Not established as stable across runs, though a
  build-constraint cause would be deterministic rather than intermittent.
- **The other three call sites.** This run exercised
  `release-pr-multi-os.yml` only; `ci.yml`'s three call sites have not yet run in
  CI, so this is a one-workflow observation.

## Residual risk

If the cause is a build constraint that is intended, the correct outcome is a
recorded exclusion, not a fix — and the census will keep reporting it on every
windows run, which is the point rather than noise.

If it is unintended, the package has been shipping without windows test coverage
for an unknown length of time, and nothing in CI would have said so. The interval
cannot be bounded from this run: the signal did not exist before this card.
