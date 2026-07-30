# SPEC-UPDATE-CI-GUARD-001 — Acceptance Criteria

## §A Discipline

1. **Every AC states a command runnable as written from the repository root, and its expected
   observable output** — not merely its expected exit code. A criterion phrased as a property with
   no command is not an AC.
2. **`go test -run <pattern>` exits 0 on zero matches.** Every `-run` AC therefore additionally
   requires the literal `--- PASS: <exact test name>` line in the output. Vacuity baseline recorded
   in this tree while authoring:
   ```
   $ go test -run 'TestCIFilterCoversBehavioralPaths' ./internal/template/ ; echo "exit=$?"
   ok  	github.com/modu-ai/moai-adk/internal/template	(cached) [no tests to run]
   exit=0
   ```
   An AC whose only assertion is `exit 0` would pass against a tree containing no such test at all,
   and is rejected. Sibling `SPEC-UPDATE-YAML-PRESERVE-001` failed its plan audit at 0.71 because 14
   of its 22 criteria passed this way.
3. **Presence is not reachability.** A grep proving a token exists in a YAML file proves the token
   exists, not that the gate it configures fires. Every presence-only criterion below is paired with
   a behavioural criterion, and each behavioural criterion states its **falsification**: what would
   have to be broken for it to fail, and confirmation that it would in fact fail then.
4. **Baselines were observed in this tree while authoring** — worktree `9426bf49b` on
   `plan/epic-update-config-audit`, whose `main` base is `1d4e4f7da`. Each AC carries its observed
   pre-change baseline so a reviewer can distinguish a real change from a no-op.
5. **`git stash` is prohibited.** This checkout is shared with concurrent sessions and `git stash` is
   repository-global. Falsification uses `go test -overlay` or a scratch `git worktree` driven by
   `go -C`.
6. **All fixtures use `t.TempDir()`** and touch no path outside it (NFR-UCG-001).
7. **Classification.** Each AC is tagged `[behavioural]` or `[presence]`. §D records the totals.

## §B Acceptance criteria

### M1 — Coverage gate contract

#### AC-UCG-001 `[presence]` — the baseline file exists and records attributable values

```bash
test -f .moai/config/coverage-baselines.yaml && \
  grep -c 'baseline:' .moai/config/coverage-baselines.yaml && \
  grep -c 'policy_target:' .moai/config/coverage-baselines.yaml && \
  grep -c 'accepted_debt:' .moai/config/coverage-baselines.yaml
```

Expected: `test -f` succeeds, and the three `grep -c` invocations each print an identical count
`>= 9` — one entry per package measured in spec.md §A.3, each carrying all three fields
(REQ-UCG-008, REQ-UCG-010).

Baseline: the file does not exist —
`ls .moai/config/coverage-baselines.yaml` prints
`ls: .moai/config/coverage-baselines.yaml: No such file or directory` and exits 1.

Paired behavioural criterion: AC-UCG-002.

#### AC-UCG-002 `[behavioural]` — the gate passes on the current tree

```bash
go test -run 'TestCoverageBaselineNoRegression' -count=1 -v ./internal/quality/
```

Expected: a `--- PASS: TestCoverageBaselineNoRegression` line. The test reads
`.moai/config/coverage-baselines.yaml`, reads a `coverage.out` profile, computes per-package
statement coverage, and asserts no package is below its recorded baseline.

Falsification: edit any baseline upward by 1.0pp (via `go test -overlay`, §C-1) and the test MUST
FAIL naming that package with both figures. A PASS under the falsification means the comparison is
not load-bearing and the gate is decorative.

Baseline: `[no tests to run]`, exit 0 (§A clause 2, verbatim above).

#### AC-UCG-003 `[behavioural]` — the failure message states both figures

```bash
go test -run 'TestCoverageBaselineNoRegression/reports_both_figures' -count=1 -v ./internal/quality/
```

Expected: a `--- PASS: TestCoverageBaselineNoRegression/reports_both_figures` line. The subtest
feeds a synthetic below-baseline profile and asserts the produced message contains both the literal
baseline value and the literal measured value for the offending package (REQ-UCG-009).

Falsification: replace the message with a bare `coverage regression detected` and the subtest MUST
FAIL. A PASS means the message content is unasserted and an operator would have to re-measure
locally to act on the failure.

Baseline: no such test; `-run` matches nothing.

#### AC-UCG-004 `[behavioural]` — sub-policy baselines are recorded as debt, not normalised

```bash
go test -run 'TestCoverageBaselineDebtDeclared' -count=1 -v ./internal/quality/
```

Expected: a `--- PASS: TestCoverageBaselineDebtDeclared` line. For every entry whose `baseline` is
below its `policy_target`, the test asserts `accepted_debt: true` is set (REQ-UCG-010), so a
sub-policy figure cannot be recorded as if it were compliant.

Falsification: flip `internal/cli`'s `accepted_debt` to `false` while leaving `baseline: 75.6` and
`policy_target: 90` — the test MUST FAIL naming `internal/cli`.

Baseline observed while authoring, establishing that this criterion has real subjects —
`go test -cover ./internal/cli/... ./internal/config/...`:

```
internal/cli                     75.6%   (policy 90)  -> debt
internal/config                  80.5%   (policy 85)  -> debt
internal/cli/specid              58.3%   (policy 85)  -> debt
internal/cli/update              88.9%   (policy 85)  -> compliant
internal/cli/update/backup       88.6%   (policy 85)  -> compliant
internal/cli/update/deploy       97.5%   (policy 85)  -> compliant
internal/cli/update/merge        90.3%   (policy 85)  -> compliant
internal/cli/update/plan         95.0%   (policy 85)  -> compliant
internal/cli/update/report       92.9%   (policy 85)  -> compliant
```

#### AC-UCG-005 `[presence]` — CI invokes the gate

```bash
grep -n 'TestCoverageBaselineNoRegression\|coverage-baselines' .github/workflows/ci.yml
```

Expected: at least one line, positioned after the existing `Run tests with race detector and
coverage` step, so the gate consumes the `coverage.out` that step already produces.

Baseline at `main` `1d4e4f7da` — the profile is produced and uploaded, and nothing gates on it:

```
$ grep -n 'coverprofile\|codecov' .github/workflows/ci.yml
165:        run: go test -race -coverprofile=coverage.out -covermode=atomic ./...
168:        uses: codecov/codecov-action@v7
$ ls -a | grep -i codecov
(no output — no codecov.yml, no .codecov.yml)
```

Paired behavioural criterion: AC-UCG-002.

### M2 — Windows leg for the update path

#### AC-UCG-006 `[presence]` — a Windows job exists and is scoped to the update packages

```bash
grep -n -A12 'test-update-windows' .github/workflows/ci.yml | \
  grep -E 'runs-on: windows-latest|internal/cli/update|if:'
```

Expected: three matches — `runs-on: windows-latest`, a run step naming
`./internal/cli/update/...`, and an `if:` gating the job on a path-filter output. A job that runs
`./...` on Windows at PR time fails this criterion (REQ-UCG-005).

Baseline at `main` `1d4e4f7da` — no such job. The PR-scope test matrix is ubuntu-only
(`ci.yml:95: os: [ubuntu-latest]`), and the only Windows runner at PR time is `test-integration`
(`ci.yml:214`, matrix at `:222`) whose single run step is
`go test -tags=integration -race -timeout 180s ./test/integration/harness/...` — no update package.

Paired behavioural criterion: AC-UCG-007.

#### AC-UCG-007 `[behavioural]` — the update packages actually pass on Windows

```bash
GOOS=windows GOARCH=amd64 go vet ./internal/cli/update/... && \
  gh run list --workflow=ci.yml --limit 1 --json databaseId --jq '.[0].databaseId' | \
  xargs -I{} gh run view {} --json jobs --jq '.jobs[] | select(.name | test("Test Update \\(windows")) | .conclusion'
```

Expected: `go vet` exits 0 for the Windows target, and the `gh run view` query prints `success` for
the most recent CI run of a pull request touching `internal/cli/update/**`. The second half is the
load-bearing assertion — a job that exists but never runs, or runs and is skipped, does not satisfy
REQ-UCG-004.

Falsification: introduce a `filepath.Join(cwd, absPath)` path bug in an update package (the
Windows-divergent shape recorded in `CLAUDE.local.md` §6) and the Windows job MUST report `failure`
while the Ubuntu job reports `success`. If both report identically, the leg is not exercising
platform-divergent behaviour and its wall-time cost buys nothing.

Baseline: the job name does not exist, so the `gh run view` filter prints nothing (empty output).

#### AC-UCG-008 `[behavioural]` — the required check is reported on the complement path

```bash
go test -run 'TestCIRequiredCheckComplementCoverage' -count=1 -v ./internal/template/
```

Expected: a `--- PASS: TestCIRequiredCheckComplementCoverage` line. The test parses `ci.yml`, and for
every job whose `name:` is a required-check name, asserts that the union of its `if:` condition and
its skip-marker sibling's `if:` condition is a tautology over the filter outputs — so exactly one of
the pair always runs (REQ-UCG-002, REQ-UCG-006, NFR-UCG-004).

Falsification: delete the `test-update-windows` skip-marker sibling and the test MUST FAIL naming
the uncovered complement. A PASS means a pull request not touching `internal/cli/update/**` could
wait forever for a status that will not arrive — the exact failure already documented at
`ci.yml:186-198`.

Baseline: no such test; `-run` matches nothing.

#### AC-UCG-009 `[presence]` — the added wall-time is measured, not assumed

```bash
grep -cE 'wall-time|wall time|duration' .moai/specs/SPEC-UPDATE-CI-GUARD-001/progress.md
```

Expected: a count `>= 1`, and the surrounding `progress.md` §E.2 text cites the observed job duration
(a `gh run view --json jobs --jq '.jobs[].completedAt'` delta, or the job's reported duration) rather
than an estimate (REQ-UCG-005).

Baseline: `progress.md` §E.2 currently reads `_<pending run-phase>_`.

### M3 — Behavioural gating filter

#### AC-UCG-010 `[presence]` — the gating filter covers template and config paths

```bash
sed -n '/^            behavioral:/,/^            [a-z_]*:/p' .github/workflows/ci.yml | \
  grep -cE "internal/template/templates|\.moai/config"
```

Expected: a count `>= 2` — the filter that gates the test job lists both
`internal/template/templates/**` and `.moai/config/**` (REQ-UCG-001).

Baseline at `main` `1d4e4f7da` — the `go_code` filter (`ci.yml:65-71`) is six entries and contains
neither:

```
go_code:
  - '**/*.go'
  - 'go.mod'
  - 'go.sum'
  - 'Makefile'
  - '.github/workflows/ci.yml'
  - '.github/workflows/codeql.yml'
```

Paired behavioural criterion: AC-UCG-011.

#### AC-UCG-011 `[behavioural]` — a template-only change now triggers the test job

```bash
go test -run 'TestCIFilterCoversBehavioralPaths' -count=1 -v ./internal/template/
```

Expected: a `--- PASS: TestCIFilterCoversBehavioralPaths` line. The test parses `ci.yml`, extracts
the glob list of the filter named in the test job's `if:` condition, and asserts that each of a
representative path set matches at least one glob. The representative set includes at minimum
`internal/template/templates/.moai/config/sections/workflow.yaml` and
`.moai/config/sections/quality.yaml` (REQ-UCG-003).

Falsification: remove `internal/template/templates/**` from the filter and the test MUST FAIL naming
the unmatched path. A PASS under the falsification means the test asserts nothing about the filter's
contents.

Non-vacuity (NFR-UCG-002): the test asserts `len(representativePaths) >= 2` before iterating, so an
accidentally-emptied fixture list fails rather than passing.

Baseline: `[no tests to run]`, exit 0 (§A clause 2, verbatim above).

#### AC-UCG-012 `[behavioural]` — end-to-end: a config-only pull request runs Go tests

```bash
gh pr list --search 'path:.moai/config/sections' --state merged --limit 1 --json number --jq '.[0].number' | \
  xargs -I{} gh pr checks {} --json name,conclusion --jq '.[] | select(.name | startswith("Test")) | "\(.name) \(.conclusion)"'
```

Expected: for the first pull request merged after this milestone that touches only
`.moai/config/**`, the `Test (ubuntu-latest)` check reports `SUCCESS` from the real `test` job — not
from `test-skip-marker`. Confirm by inspecting the run's job list:
`gh run view <id> --json jobs --jq '.jobs[] | select(.name=="Test (ubuntu-latest)") | .steps[].name'`
must include `Run tests with race detector and coverage`, which the skip-marker job does not have.

Falsification: the same query against a pre-milestone config-only pull request shows the job's only
step as `Skip (no Go changes)` — the observable difference between the two paths.

Baseline: the skip-marker path is what fires today. `ci.yml:200` gates it on
`needs.detect.outputs.go_code != 'true'`, and its single step (`ci.yml:206-207`) is
`Skip (no Go changes)`.

### M4 — Semantic merge-outcome guard

#### AC-UCG-013 `[behavioural]` — the semantic guard exists and asserts outcomes

```bash
go test -run 'TestMergeSemanticOutcomes' -count=1 -v ./internal/cli/update/backup/
```

Expected: a `--- PASS: TestMergeSemanticOutcomes` line, with `-v` output listing the seven named
subtests from plan.md §F M4 (`three_way_divergence`, `user_only_key`, `template_only_key`,
`zero_value_false`, `zero_value_empty_string`, `zero_value_int`, `nested_old_only`). Subtests
recording a sibling-owned known defect appear as `--- SKIP` with `SPEC-UPDATE-YAML-PRESERVE-001` in
the skip reason (REQ-UCG-014).

Falsification: §C-2 — against a `merge.go` whose zero-value branch is removed, the guard MUST FAIL.

Baseline: no such test; `-run` matches nothing. The existing suite passes at 88.6% package coverage
with every `merge.go` function at 100.0%:

```
$ go tool cover -func=/tmp/ckh-backup.out | grep -E 'merge\.go|total'
merge.go:20:	MergeYAML3Way		100.0%
merge.go:43:	DeepMerge3Way		100.0%
merge.go:103:	ValuesEqual		100.0%
merge.go:116:	MergeYAMLDeep		100.0%
merge.go:134:	DeepMergeMaps		100.0%
total:					88.6%
```

#### AC-UCG-014 `[behavioural]` — the guard asserts whole-map equality, not absence of error

```bash
go test -run 'TestMergeSemanticOutcomes/three_way_divergence' -count=1 -v ./internal/cli/update/backup/
```

Expected: a `--- PASS: TestMergeSemanticOutcomes/three_way_divergence` line. Additionally, the test
source must compare the complete output map against a declared expected map:

```bash
grep -cE 'reflect\.DeepEqual|cmp\.Diff|assert\.Equal.*want' \
  internal/cli/update/backup/merge_semantics_test.go
```

Expected: a count `>= 1`. A test whose only assertion is `if err != nil` satisfies neither half.

Falsification: replace the whole-map comparison with an error check and AC-UCG-013's falsification
(§C-2) flips to PASS — which is the demonstration that the comparison, not the invocation, is what
detects the defect.

Baseline: the file does not exist; `ls internal/cli/update/backup/merge_semantics_test.go` prints
`No such file or directory` and exits 1.

#### AC-UCG-015 `[presence]` — the merge implementation is unmodified

```bash
git diff --stat main -- internal/cli/update/backup/merge.go
```

Expected: **no output** — REQ-UCG-014 and plan.md §D3 forbid this SPEC from editing the merge
implementation. Any diff means the SPEC crossed into `SPEC-UPDATE-YAML-PRESERVE-001`'s scope.

Baseline: no output today (the file is unmodified relative to `main`).

### M5 — Neutrality detection-pattern coverage

#### AC-UCG-016 `[behavioural]` — the three leak shapes are now detected

```bash
go test -run 'TestLeakPatternCoversKnownShapes' -count=1 -v ./internal/template/
```

Expected: a `--- PASS: TestLeakPatternCoversKnownShapes` line, with subtests asserting that the
extended pattern set matches each of the three measured shapes (REQ-UCG-015):

```
SPEC-AGENT-ARCH-V2-001            (unregistered SPEC-ID family)
issue #653                        (issue, not PR)
plan.md §D D6                     (internal artifact citation)
```

Falsification: revert any one pattern extension and the corresponding subtest MUST FAIL naming the
undetected shape. A PASS under the revert means the subtest asserts nothing about the pattern set.

Baseline at `main` `1d4e4f7da` — each shape escapes for a distinct structural reason, verified
against the implementation:

```
$ grep -n 'C1-spec-id-prefix' -A1 internal/template/internal_content_leak_test.go
170:		name:    "C1-spec-id-prefix",
171:		pattern: regexp.MustCompile(`\bSPEC-(V3R[2-6]|AGENCY|WORKTREE)-[A-Z0-9-]+\b`),
$ grep -n 'C1c-spec-id-non-v3r-known-families' -A1 internal/template/internal_content_leak_test.go
281:		name:    "C1c-spec-id-non-v3r-known-families",
282:		pattern: regexp.MustCompile(`\bSPEC-(DB-SYNC-RELOC|PROJECT-DB-HINT)-[0-9]{3}\b`),
$ grep -n 'C6-pr-number-ref' -A2 internal/template/template_neutrality_audit_test.go
166:		name:      "C6-pr-number-ref",
168:		pattern:   regexp.MustCompile(`PR #[0-9]+`),
```

`AGENT-ARCH-V2` is in neither SPEC-ID enumeration; `issue #653` is not `PR #N`; no class in either
file matches an artifact citation.

#### AC-UCG-017 `[behavioural]` — pedagogical placeholders are not false-positived

```bash
go test -run 'TestLeakPatternRejectsPedagogicalPlaceholders' -count=1 -v ./internal/template/
```

Expected: a `--- PASS: TestLeakPatternRejectsPedagogicalPlaceholders` line, asserting the extended
patterns do **not** match `SPEC-BUG-042`, `SPEC-X-001`, or `SPEC-PAY-001` — the three shapes named
verbatim in the existing rationale comment at
`internal/template/internal_content_leak_test.go:271-281` (REQ-UCG-017).

Falsification: substitute a bare `SPEC-[A-Z-]+-[0-9]+` wildcard with no allowlist and the test MUST
FAIL naming each placeholder it wrongly matched. A PASS means the allowlist is absent or inert and
REQ-UCG-016's trade-off was made blind.

Baseline: no such test; the current narrow enumeration trivially does not match the placeholders,
but nothing asserts that it continues not to.

#### AC-UCG-018 `[presence]` — the false-positive cost is measured, not assumed

```bash
grep -cE 'false.positive' .moai/specs/SPEC-UPDATE-CI-GUARD-001/progress.md
```

Expected: a count `>= 1`, and the surrounding `progress.md` §E.2 text records the observed
whole-tree match count for each candidate pattern — the output of a
`grep -rEc '<pattern>' internal/template/templates/` run — so REQ-UCG-016's decision between the
generic form and the enumeration extension rests on a measurement.

Baseline: `progress.md` §E.2 currently reads `_<pending run-phase>_`.

#### AC-UCG-019 `[presence]` — the shipped leaks are untouched by this SPEC

```bash
git diff --stat main -- internal/template/templates/.moai/config/
```

Expected: **no output**. Removing the three leaks is `SPEC-CONFIG-KEY-HONESTY-001` REQ-CKH-012's
work (spec.md §C); this SPEC only makes the shapes detectable.

Baseline: no output today. The three leaks are present and unmodified:

```
$ grep -rn 'SPEC-AGENT-ARCH-V2-001\|issue #' internal/template/templates/.moai/config/
.../sections/workflow.yaml:65:    # cycle (plan.md §D D6, SPEC-AGENT-ARCH-V2-001 M3b). Values mirror the
.../sections/workflow.yaml:85:    # model_routing_profiles: No-Haiku 3-tier policy (SPEC-AGENT-ARCH-V2-001
.../sections/llm.yaml:179:    # (issue #653). Claude Code reports context_window_size based on the
```

### M6 — Class-naming hygiene

#### AC-UCG-020 `[behavioural]` — no bare class citation remains

```bash
go test -run 'TestLeakClassCitationsNameOwningFile' -count=1 -v ./internal/template/
```

Expected: a `--- PASS: TestLeakClassCitationsNameOwningFile` line. The test scans the two guard test
files and the neutrality workflow for citations matching `\bC[0-9][a-z]?\b` and asserts each appears
within a line that also names its owning file, so `C1` and `C6` are never ambiguous between the two
independently-numbered schemes (REQ-UCG-018, spec.md §A.6 drift 4).

Falsification: strip the filename qualifier from one citation and the test MUST FAIL naming the
file and line. A PASS means the scan is not reaching the citations.

Baseline: both files define a `C1` and a `C6` with different meanings —
`template_neutrality_audit_test.go:126` `C1-macos-bias-path` (`/Users/`) and `:166` `C6-pr-number-ref`
(`PR #[0-9]+`), versus `internal_content_leak_test.go:170` `C1-spec-id-prefix` and its skill-scoped
`C6-agentless-test-ref` at `:215` — and no test asserts the disambiguation.

### Cross-cutting

#### AC-UCG-021 `[behavioural]` — the full suite is green and the build is clean

```bash
go build ./... && go vet ./... && go test -count=1 ./...
```

Expected: exit 0 from all three, with `ok` lines and no `FAIL` line in the test output.

Baseline: green in this worktree at pre-flight (plan.md §C).

#### AC-UCG-022 `[behavioural]` — no test writes outside `t.TempDir()`

```bash
go test -count=1 ./internal/template/ ./internal/quality/ ./internal/cli/update/... && \
  git status --porcelain
```

Expected: the tests pass and `git status --porcelain` prints **no output** — no file created or
modified by the test run (NFR-UCG-001).

Baseline: `git status --porcelain` prints no output in this worktree before the run.

#### AC-UCG-023 `[presence]` — required-check names are unchanged

```bash
git diff main -- .github/workflows/ci.yml | grep -E '^[-+] *name: (Test|Lint|Build|Integration)' | sort
```

Expected: every `-` line has a matching `+` line with identical text, i.e. no required-check `name:`
value was renamed or removed (NFR-UCG-004). New jobs may add new `+` lines with no `-` counterpart;
a `-` line with no `+` counterpart fails this criterion.

Baseline: no diff today. The current required-check names are
`Test (${{ matrix.os }})` (emitted by both `test` at `ci.yml:87` and `test-skip-marker` at `:185`),
`Integration Tests (${{ matrix.os }})` (`:215`), `Lint` (`:247`), and
`Build (${{ matrix.goos }}/${{ matrix.goarch }})` (`:288`).

## §C Falsification procedures

Each new guard must be shown to FAIL against unfixed or deliberately-broken code. `git stash` is
prohibited (§A clause 5).

### C-1 — the coverage gate actually catches a regression

```bash
mkdir -p /tmp/ucg-falsify
sed 's/baseline: 75.6/baseline: 76.6/' .moai/config/coverage-baselines.yaml \
  > /tmp/ucg-falsify/coverage-baselines.yaml
printf '{"Replace":{".moai/config/coverage-baselines.yaml":"/tmp/ucg-falsify/coverage-baselines.yaml"}}' \
  > /tmp/ucg-falsify/overlay.json
go test -overlay=/tmp/ucg-falsify/overlay.json -run 'TestCoverageBaselineNoRegression' \
  -count=1 -v ./internal/quality/
```

Expected: **FAIL**, naming `internal/cli` with `baseline 76.6%, measured 75.6%`. A PASS means the
per-package comparison is not load-bearing and the gate would not catch real erosion either.

### C-2 — the semantic merge guard actually catches a broken merge

```bash
mkdir -p /tmp/ucg-falsify
# strip the zero-value branch from the merge implementation in a copy
sed '/IsZero()/,+2d' internal/cli/update/backup/merge.go > /tmp/ucg-falsify/merge.go
printf '{"Replace":{"internal/cli/update/backup/merge.go":"/tmp/ucg-falsify/merge.go"}}' \
  > /tmp/ucg-falsify/overlay2.json
go test -overlay=/tmp/ucg-falsify/overlay2.json -run 'TestMergeSemanticOutcomes' \
  -count=1 -v ./internal/cli/update/backup/
```

Expected: **FAIL** on `zero_value_false`, `zero_value_empty_string`, and `zero_value_int`, each
naming the key and the expected-versus-actual value. A PASS here is the exact hazard spec.md §A.4
names — the guard would then be measuring execution, not outcome, which statement coverage already
does at 100%.

### C-3 — the filter-coverage test actually reads the filter

```bash
git worktree add /tmp/ucg-wt HEAD
# in the scratch tree only: remove internal/template/templates/** from the gating filter
go -C /tmp/ucg-wt test -run 'TestCIFilterCoversBehavioralPaths' -count=1 -v ./internal/template/
git worktree remove /tmp/ucg-wt
```

Expected: **FAIL**, naming `internal/template/templates/.moai/config/sections/workflow.yaml` as
unmatched. A PASS means the test parses the workflow but asserts nothing about the glob list, and
AC-UCG-010's presence grep would be the only remaining defence — which §A clause 3 rejects.

### C-4 — the complement-coverage test actually catches a missing skip-marker

```bash
git worktree add /tmp/ucg-wt2 HEAD
# in the scratch tree only: delete the test-update-windows skip-marker job
go -C /tmp/ucg-wt2 test -run 'TestCIRequiredCheckComplementCoverage' -count=1 -v ./internal/template/
git worktree remove /tmp/ucg-wt2
```

Expected: **FAIL**, naming the required-check name whose complement path is uncovered. A PASS means
a pull request could be rendered permanently unmergeable without any test noticing — the failure
mode `ci.yml:186-198` already documents from a prior occurrence.

### C-5 — the pattern extension does not silently over-match

```bash
git worktree add /tmp/ucg-wt3 HEAD
# in the scratch tree only: replace the enumerated SPEC-ID class with a bare wildcard
go -C /tmp/ucg-wt3 test -run 'TestLeakPatternRejectsPedagogicalPlaceholders' \
  -count=1 -v ./internal/template/
git worktree remove /tmp/ucg-wt3
```

Expected: **FAIL**, naming `SPEC-BUG-042`, `SPEC-X-001`, and `SPEC-PAY-001` as wrongly matched. A
PASS means the false-positive guard is decorative and REQ-UCG-016's measured trade-off is
unenforced.

## §D Classification totals

| Milestone | behavioural | presence | total |
|---|---:|---:|---:|
| M1 — coverage gate | 3 (AC-002, -003, -004) | 2 (AC-001, -005) | 5 |
| M2 — Windows leg | 2 (AC-007, -008) | 2 (AC-006, -009) | 4 |
| M3 — gating filter | 2 (AC-011, -012) | 1 (AC-010) | 3 |
| M4 — semantic guard | 2 (AC-013, -014) | 1 (AC-015) | 3 |
| M5 — pattern coverage | 2 (AC-016, -017) | 2 (AC-018, -019) | 4 |
| M6 — naming hygiene | 1 (AC-020) | 0 | 1 |
| Cross-cutting | 2 (AC-021, -022) | 1 (AC-023) | 3 |
| **Total** | **14** | **9** | **23** |

Every presence-only criterion is paired with a behavioural criterion covering the same requirement:
AC-001/-005 → AC-002; AC-006 → AC-007; AC-009 → AC-007; AC-010 → AC-011; AC-015 → AC-013;
AC-018/-019 → AC-016; AC-023 → AC-008.

## §E Definition of Done

- All of AC-UCG-001 through AC-UCG-023 produce their stated observable output.
- All five falsification procedures C-1 through C-5 produce **FAIL** against broken code.
- `internal/cli/update/backup/merge.go` is unmodified relative to `main` (AC-UCG-015).
- `internal/template/templates/.moai/config/` is unmodified relative to `main` (AC-UCG-019).
- No required-check `name:` value was renamed or removed (AC-UCG-023).
- `progress.md` §E.2 cites the observed command output for every claim, per
  `.claude/rules/moai/core/verification-claim-integrity.md`, including the measured Windows-leg
  wall-time (AC-UCG-009) and the measured false-positive counts (AC-UCG-018).
