# SPEC-GRAPH-FRESHNESS-CADENCE-001 — Acceptance Criteria

Every criterion below is binary and names the command that decides it. A criterion whose deciding
command was not run is a Gap, never a pass.

## §D. AC Matrix

| AC | REQ covered | Milestone | Deciding command |
|---|---|---|---|
| AC-GFC-001 | REQ-GFC-002 | M1 | `go test ./internal/mx/ -run TestDescribedWorthy -count=1` |
| AC-GFC-002 | REQ-GFC-001 | M1 | `go test ./internal/graph/ -run TestGitDiffNameCount_Predicate -count=1` |
| AC-GFC-003 | REQ-GFC-001 | M1 | `./bin/moai graph check --json` against a fixture stamped at `9326b5478d…` |
| AC-GFC-004 | REQ-GFC-003 | M1 | `go test ./internal/mx/ -run TestCodemapsFingerprint_ProducerConsumer -count=1` |
| AC-GFC-005 | REQ-GFC-011 | M1 | `go test ./internal/graph/ -run TestCheckCodemaps_Absent -count=1` |
| AC-GFC-006 | REQ-GFC-005 | M2 | `git log --first-parent -30 --name-only …` (derivation record) |
| AC-GFC-007 | REQ-GFC-006 | M2 | `grep -n "CodemapsChangedFiles" internal/graph/check.go` |
| AC-GFC-008 | REQ-GFC-007 | M3 | `go test ./internal/graph/ -run TestCheckCodemaps_Contribution -count=1` |
| AC-GFC-009 | REQ-GFC-008 | M3 | `./bin/moai graph check` on a stale fixture (stderr inspection) |
| AC-GFC-010 | REQ-GFC-010 | M3 | `./bin/moai graph check --json \| jq` |
| AC-GFC-011 | — | — | *withdrawn at v0.2.0 with REQ-GFC-004 (audit D8)* |
| AC-GFC-012 | — | — | *withdrawn at v0.2.0 with REQ-GFC-012 (audit D8)* |
| AC-GFC-013 | REQ-GFC-009 | all | `grep -rn "graph stamp" .github/ .claude/` |
| AC-GFC-014 | REQ-GFC-003a | M1 | `go test ./internal/graph/ -run TestSourceFingerprintsForEdges_Unchanged -count=1` |

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

**AC-GFC-003** — Given a checkout of `48eb945df` whose codemaps provenance is stamped at
`9326b5478d0f51979dfb498527458dcea5e0370b`, When the **built binary** runs `moai graph check --json`
against it, Then the codemaps row reports `value: 2`.
Decided by: `./bin/moai graph check --json | jq '.layers[] | select(.layer=="codemaps") | .value'`
→ `2`, run against that fixture with the binary built from the branch head.
The unfiltered figure `65` is retained only as the **baseline-attribution** row — it records what
the metric did before the change and is not itself a criterion.
*Rewritten at v0.2.0 (audit D3): v0.1.0 decided this by a hand-written shell filter chain over git
output, which re-derived the baseline instead of exercising the delivered code. Both of its figures
reproduce at HEAD with no implementation present, so an implementation that filtered nothing would
have left it green.*

**AC-GFC-004** — Given a tree with uncommitted described-source changes, When
`moai graph stamp codemaps` writes a dirty `ContentFingerprint` and `moai graph check` immediately
reads it back with no intervening edit, Then the codemaps verdict is `fresh`; and When a file under
a `testdata` segment is then modified, Then the verdict remains `fresh`; and When a production `.go`
file is modified, Then the verdict becomes `stale`.
Decided by: `go test ./internal/mx/ -run TestCodemapsFingerprint_ProducerConsumer -count=1`.
Mutant this kills: a filtered checker paired with an unfiltered codemaps stamp writer — the first
assertion fails immediately, because the tree is stale against its own fresh stamp.

**AC-GFC-005** — Given each of the **seven** absent-verdict branches of `checkCodemaps` in turn,
When `checkCodemaps` runs, Then the verdict is `absent` carrying that branch's pre-existing reason
string, and in no case is it `fresh`. The seven, enumerated by `grep -n "VerdictAbsent"
internal/graph/check.go` over the `checkCodemaps` body:

| # | Branch | Reason string |
|---|---|---|
| 1 | codemaps directory missing | `codemaps directory missing` |
| 2 | provenance file missing | `no provenance block — freshness-unjudgeable, not fresh` |
| 3 | provenance unparseable / `schema_version` 0 | `provenance block unparseable — freshness-unjudgeable` |
| 4 | described root invalid | `described root %q invalid: …` |
| 5 | dirty path, roots unreadable | `described roots unreadable: …` |
| 6 | clean stamp carries no commit sha | `clean stamp carries no commit sha — freshness-unjudgeable` |
| 7 | stamped commit not comparable | `stamped commit not comparable (unmeasured, system error follows)` |

Branch 7 additionally returns a **non-nil error** alongside the report, and the criterion pins that
pair shape — `(report.Verdict == absent, err != nil)` — not the verdict alone. It is the branch
t291 / t292's orphan hazard produces, and REQ-GFC-011 names it explicitly.
Branches 5 and 6 sit on the dirty and clean codemaps paths this SPEC modifies.
Decided by: `go test ./internal/graph/ -run TestCheckCodemaps_Absent -count=1`.
*Extended at v0.2.0 from four branches to seven (audit D4).*

**AC-GFC-006** — Given the tree at implementation time, When **both** axes are measured over the last
30 first-parent integrations, Then `progress.md` §E.2 records each command with its verbatim output,
the percentile convention used, and the conclusion for the threshold:

- **Axis 1, per-integration contribution**: `git log --first-parent -30 --name-only
  --pretty=format:"===%h" -- internal cmd pkg`, filtered by the predicate, counted per integration —
  yielding the zero-count, median, nearest-rank p90 and mean.
- **Axis 2, cumulative-crossing cadence**: the same log with `--reverse`, accumulating the distinct
  described-worthy union, yielding the integration count at which the union first reaches the
  threshold, plus the window's date span so the crossing can be expressed as a red frequency.

The record itself is the observable. A threshold **retained** without this record fails exactly as a
changed one would, and a record carrying only Axis 1 fails.
*Extended at v0.2.0 to require Axis 2 and to name the percentile convention (audit D2, D5).*

**AC-GFC-007** — Given the threshold value in the delivered code, When the branch's own record is
searched over a **bounded** space, Then no justification appeals to a check that was failing.
Decided by three commands, all three required:
1. `grep -n "CodemapsChangedFiles" internal/graph/check.go` — locates the constant and its value.
2. `git log d2cba5e21..HEAD --format=%B` — the branch's own commit messages, searched for a
   justification of the form "raised/lowered so that … passes".
3. `git diff d2cba5e21..HEAD -- internal/graph/check.go` — the delivered diff, its comments read for
   the same.
The negative is asserted only over those three surfaces and is stated as such; it is not a claim
about every artifact in the repository.
*Bounded at v0.2.0 (audit D7): v0.1.0 asserted an unbounded negative decided by a grep that could
not decide it.*

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

**AC-GFC-011** — *withdrawn at v0.2.0 with REQ-GFC-004 (audit D8).* No configuration key is added,
so there is nothing to resolve. `graphCheckThresholds` keeps its existing behaviour, including the
present-but-unparseable `gate.yaml` → exit 2 contract, and AC-GFC-014's untouched-surface assertion
covers the fact that this SPEC does not change it.

**AC-GFC-012** — *withdrawn at v0.2.0 with REQ-GFC-012 (audit D8).* No config key is added, so no
template mirror is required.

**AC-GFC-014** — Given the tree before this change and the tree after it, When
`SourceFingerprintsForEdges` is computed over the same fixture project, Then every one of its four
source-set fingerprints is byte-identical across the two, and in particular the `.moai/project/codemaps`
and `.moai/specs` entries are **not** equal to the empty-entry hash
`e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`.
Decided by: `go test ./internal/graph/ -run TestSourceFingerprintsForEdges_Unchanged -count=1`.
Mutant this kills: the predicate pushed down into `aggregateFingerprint`. Measured premise for the
inequality assertion: `.moai/project/codemaps` and `.moai/specs` contain zero `.go` files
(`find … -name '*.go' | wc -l` → 0 for both), so under a `.go`-only filter both collapse to exactly
that constant — which is why the assertion is stated as an inequality against it rather than as a
general "fingerprints differ".
*Added at v0.2.0 (audit D1).*

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
