# SPEC-GRAPH-FRESHNESS-CADENCE-001 — Acceptance Criteria

Every criterion below is binary and names the command that decides it. A criterion whose deciding
command was not run is a Gap, never a pass.

## §D. AC Matrix

| AC | REQ covered | Milestone | Deciding command |
|---|---|---|---|
| AC-GFC-001 | REQ-GFC-002 | M1 | `go test ./internal/mx/ -run TestDescribedWorthy -count=1` |
| AC-GFC-002 | REQ-GFC-001 | M1 | `go test ./internal/graph/ -run TestGitDiffNameCount_Predicate -count=1` |
| AC-GFC-003 | REQ-GFC-001 | M1 | `git diff --name-only <stamp> <sha> -- internal cmd pkg` + filter |
| AC-GFC-004 | REQ-GFC-003 | M1 | `go test ./internal/mx/ -run TestAggregateFingerprint_Predicate -count=1` |
| AC-GFC-005 | REQ-GFC-011 | M1 | `go test ./internal/graph/ -run TestCheckCodemaps_Absent -count=1` |
| AC-GFC-006 | REQ-GFC-005 | M2 | `git log --first-parent -30 --name-only …` (derivation record) |
| AC-GFC-007 | REQ-GFC-006 | M2 | `grep -n "CodemapsChangedFiles" internal/graph/check.go` |
| AC-GFC-008 | REQ-GFC-007 | M3 | `go test ./internal/graph/ -run TestCheckCodemaps_Contribution -count=1` |
| AC-GFC-009 | REQ-GFC-008 | M3 | `./bin/moai graph check` on a stale fixture (stderr inspection) |
| AC-GFC-010 | REQ-GFC-010 | M3 | `./bin/moai graph check --json \| jq` |
| AC-GFC-011 | REQ-GFC-004 | M4 | `go test ./internal/cli/ -run TestGraphCheckThresholds -count=1` |
| AC-GFC-012 | REQ-GFC-012 | M4 | `git diff --stat internal/template/templates/.moai/config/sections/gate.yaml` |
| AC-GFC-013 | REQ-GFC-009 | all | `grep -rn "graph stamp" .github/ .claude/` |

## §D.1 — Criteria

**AC-GFC-001** — Given a table of repo-relative paths covering each predicate branch, When the
described-worthy predicate is applied, Then `internal/graph/check.go` counts, and
`internal/astgrep/coverage_matrix.go` is admitted while `internal/astgrep/coverage_matrix_test.go`,
`internal/astgrep/testdata/rule-tests/x.yml`, `internal/hook/testdata/nav/a.go`, and
`internal/template/templates/.moai/config/astgrep-rules/go/concurrency.yml` are each rejected, with
the rejecting branch named per case.
Decided by: `go test ./internal/mx/ -run TestDescribedWorthy -count=1`.

**AC-GFC-002** — Given a fixture repository whose working tree differs from a stamped commit by one
production `.go` file and forty `testdata` fixtures, When `gitDiffNameCount` runs, Then it returns
`1`.
Mutant this kills: predicate applied to the `git diff` branch only — the fixture adds one of the
forty fixtures as **untracked**, so an unfiltered `git ls-files --others` branch returns `2` and the
test fails.
Decided by: `go test ./internal/graph/ -run TestGitDiffNameCount_Predicate -count=1`.

**AC-GFC-003** — Given the historical window that produced the failure streak, When the corrected
metric is applied to it, Then the count is 2 rather than 65.
Decided by: `git diff --name-only 9326b5478d0f51979dfb498527458dcea5e0370b 48eb945df -- internal cmd pkg`
piped through the predicate filter, compared against the same command without the filter. Both
figures recorded verbatim; the unfiltered figure must reproduce 65, and a divergence there
invalidates the baseline rather than the implementation.

**AC-GFC-004** — Given a tree stamped dirty, When a file under a `testdata` segment is modified and
`AggregateDescribedFingerprint` is recomputed, Then the fingerprint is unchanged; and When a
production `.go` file is modified, Then the fingerprint changes.
Decided by: `go test ./internal/mx/ -run TestAggregateFingerprint_Predicate -count=1`.

**AC-GFC-005** — Given each of the four unjudgeable conditions in turn (codemaps directory missing;
`provenance.json` missing; `provenance.json` unparseable; described root escaping the project root),
When `checkCodemaps` runs, Then the verdict is `absent` with the pre-existing reason string, and in
no case is it `fresh`.
Decided by: `go test ./internal/graph/ -run TestCheckCodemaps_Absent -count=1`.

**AC-GFC-006** — Given the tree at implementation time, When the per-integration described-worthy
distribution is measured over the last 30 first-parent integrations, Then `progress.md` §E.2 records
the command, its verbatim output, the resulting p90 and mean, and the derivation from those two
figures to the adopted threshold.
Decided by: `git log --first-parent -30 --name-only --pretty=format:"===%h" -- internal cmd pkg`
filtered by the predicate; the record itself is the observable. A threshold adopted without this
record fails.

**AC-GFC-007** — Given the adopted threshold value, When the run-phase evidence is read, Then the
value is traceable to AC-GFC-006's derivation and to no other justification; specifically, no commit
message, code comment, or evidence line justifies the value by reference to a check that was failing.
Decided by: `grep -n "CodemapsChangedFiles" internal/graph/check.go` plus inspection of the
`progress.md` §E.2 derivation record.

**AC-GFC-008** — Given a merge commit under judgment whose own contribution is 0 while the
cumulative count exceeds the threshold, When the codemaps layer is measured, Then the report carries
cumulative > threshold and contribution = 0; and Given a checkout with no first parent, Then the
contribution field is reported absent rather than 0.
Decided by: `go test ./internal/graph/ -run TestCheckCodemaps_Contribution -count=1`.
Mutant this kills: an implementation that defaults a missing first parent to 0, which would report
every linear checkout as an inheriting one.

**AC-GFC-009** — Given a fixture tree whose described-worthy count exceeds the threshold, When
`moai graph check` runs, Then stderr names at least one driving path, and When the count exceeds the
display bound, Then an explicit overflow indicator is present and the listing is truncated to the
bound.
Decided by: `./bin/moai graph check` against the fixture, stderr captured and inspected.

**AC-GFC-010** — Given the same fixture, When `moai graph check --json` runs, Then the codemaps row
carries the per-change contribution and the driving-path list as distinct fields, and the
pre-existing fields (`layer`, `metric`, `value`, `threshold`, `verdict`, `reason`) are unchanged in
name and meaning.
Decided by: `./bin/moai graph check --json | jq '.layers[] | select(.layer=="codemaps")'`.

**AC-GFC-011** — Given a `gate.yaml` supplying `graph_freshness.described_exclude_prefixes`, When
thresholds and predicate configuration are resolved, Then the supplied list replaces the default;
Given the key absent, Then the default set applies; and Given a present-but-unparseable `gate.yaml`,
Then the pre-existing exit-2 contract holds.
Decided by: `go test ./internal/cli/ -run TestGraphCheckThresholds -count=1`.

**AC-GFC-012** — Given the config key added under `.moai/config/sections/gate.yaml`, When the
template source is inspected, Then it carries the same key, and the template file contains no SPEC
ID, REQ token, internal date, commit SHA, or absolute host path.
Decided by: `git diff --stat internal/template/templates/.moai/config/sections/gate.yaml` followed by
`grep -nE "SPEC-|REQ-|/Users/" internal/template/templates/.moai/config/sections/gate.yaml`
returning no match, and `make build` succeeding.

**AC-GFC-013** — Given the delivered change set, When the pipeline surfaces are searched, Then no
`graph stamp` invocation exists in `.github/` or `.claude/` that the change introduced, and the
tracked `.moai/project/codemaps/provenance.json` is byte-identical to its state at the branch base.
Decided by: `grep -rn "graph stamp" .github/ .claude/` and
`git diff --stat d2cba5e21 -- .moai/project/codemaps/provenance.json` returning empty.

## §D.2 — Edge Cases

- A described root that does not exist on disk contributes nothing and must not error (existing walk
  behaviour, preserved).
- A non-regular file (FIFO, socket, device, symlink) under a described root must not be read; the
  existing regular-file guard in `aggregateFingerprint` is preserved ahead of the predicate.
- A path whose only `testdata` occurrence is a substring rather than a full segment
  (`internal/foo/testdatax/a.go`) must be **admitted** — segment equality, not substring matching.
- An exclusion prefix supplied without a trailing separator must not match a sibling directory
  sharing the prefix as a name fragment.

## §D.3 — Definition of Done

- All thirteen criteria decided, each with its command's verbatim output recorded in `progress.md`
  §E.2.
- `go vet ./internal/graph/... ./internal/mx/... ./internal/config/... ./internal/cli/...` clean.
- Affected-package tests green locally; the full verdict is CI's.
- `make build` succeeds and the template mirror is present.
- No restamp performed, and no restamp introduced (AC-GFC-013).
- The three open questions in `spec.md` §G remain recorded as open, or carry the operator's decision.
