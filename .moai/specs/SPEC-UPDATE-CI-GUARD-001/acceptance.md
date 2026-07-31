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
   `plan/epic-update-config-audit`. That branch's merge-base with `origin/main` was `76d9a8f3b`; the
   local branch `main` at `1d4e4f7da` was a stale divergent branch, not this branch's base, and the
   original wording naming it as such was factually wrong. The branch has since been merged with
   `origin/main`, so the baseline for every criterion below is **`d5336214e`**, and each figure was
   re-observed against that tree. Each AC carries its observed pre-change baseline so a reviewer can
   distinguish a real change from a no-op.
5. **`git stash` is prohibited.** This checkout is shared with concurrent sessions and `git stash` is
   repository-global. Falsification uses `go test -overlay` or a scratch `git worktree` driven by
   `go -C`.
6. **All fixtures use `t.TempDir()`** and touch no path outside it (NFR-UCG-001).
7. **Classification.** Each AC is tagged `[behavioural]` or `[presence]`. §D records the totals.
8. **A command that cannot observe its own expectation is not an AC, and a mutation that mutates
   nothing is not a falsification.** Clause 1 requires a command; this clause requires the command to
   be *capable of seeing* what the criterion claims. The plan audit found this SPEC failing the
   clause in five places, each rewritten in place with the observed before/after so the failure mode
   stays documented rather than merely removed:
   - **all five §C procedures were inert.** Three (§C-3, §C-4, §C-5) expressed their mutation step as
     a shell comment, so the scratch tree equalled `HEAD` and the test passed against unmodified
     code. §C-2's `sed '/IsZero()/,+2d'` matched zero lines in its target file (`grep -c 'IsZero()'
     internal/cli/update/backup/merge.go` → `0`), copying the file unchanged. §C-1 relied on
     `go test -overlay` to replace a file read at *run time*, which overlay does not do (§A clause 9).
     Every §C procedure now carries a `diff -q` no-op guard that aborts when the edit matched nothing.
   - **AC-UCG-010's extraction ran to end-of-file.** Its `sed` range never matched a closing pattern,
     so it emitted 232 of `ci.yml`'s 373 lines and the `grep` searched the whole file rather than the
     filter block. Replaced with a bounded extractor carrying its own positive and negative control.
   - **AC-UCG-002's falsification could not discriminate the open tolerance decision** (see §B).
   - **AC-UCG-017 passed vacuously on one branch** of an undecided design choice (see §B).
   - **AC-UCG-020 scanned class *definition* lines**, requiring a definition to name the file it is
     written in — unsatisfiable by construction (see §B).
9. **`go test -overlay` replaces build inputs, not runtime reads.** Measured: a test calling
   `os.ReadFile` on an overlaid non-Go file still reads the original bytes, under both relative and
   absolute overlay keys. Overlay is therefore used **only** to substitute `.go` sources (§C-2,
   verified effective); any falsification that must change a file the test opens at run time uses a
   scratch worktree under `.claude/worktrees/` with a real edit, driven by `go -C`.

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

#### AC-UCG-002 `[behavioural]` — the gate passes on the current tree — **decision-gated (tolerance)**

```bash
go test -run 'TestCoverageBaselineNoRegression' -count=1 -v ./internal/quality/
```

Expected: a `--- PASS: TestCoverageBaselineNoRegression` line. The test reads
`.moai/config/coverage-baselines.yaml`, reads a `coverage.out` profile, computes per-package
statement coverage, and asserts no package is below its recorded baseline.

Falsification: §C-1 — raise `internal/cli`'s baseline by **0.1pp** (75.7 → 75.8) in a scratch
worktree and the test MUST FAIL naming that package with both figures. A PASS means the comparison
is not load-bearing and the gate is decorative.

**Why 0.1pp, and why this criterion is decision-gated.** The tolerance value is an open decision
(spec.md §B.3). The original wording used a **1.0pp** bump, which cannot discriminate any candidate:
a 1.0pp drop fails under zero tolerance *and* under a 0.5pp epsilon *and* under a 1.0pp epsilon, so
the criterion passed identically whichever value was later chosen — it tested the gate's existence,
never its threshold. A 0.1pp bump fails **only** under zero tolerance, so it tests the proposal
rather than assuming it.

Until the tolerance is chosen this criterion is **decision-gated**: it asserts the zero-tolerance
behaviour, which is the proposal. If the Implementation Kickoff Approval gate selects an epsilon
instead, replace the single 0.1pp bump with that row's bracketing pair from spec.md §B.3 — one bump
below the threshold that MUST PASS and one above that MUST FAIL — and drop the decision-gated
marker. A single bump can never verify an epsilon; only the pair brackets it.

Falsification uses a scratch worktree rather than `go test -overlay`: the test reads
`.moai/config/coverage-baselines.yaml` at run time, and overlay does not reach runtime reads
(§A clause 9).

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

Falsification: flip `internal/cli`'s `accepted_debt` to `false` while leaving `baseline: 75.7` and
`policy_target: 90` — the test MUST FAIL naming `internal/cli`.

Baseline observed while authoring, establishing that this criterion has real subjects —
`go test -cover ./internal/cli/... ./internal/config/...`:

```
internal/cli                     75.7%   (policy 90)  -> debt
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

Baseline at `d5336214e` — the profile is produced and uploaded, and nothing gates on it:

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
grep -n -A15 'test-update-windows' .github/workflows/ci.yml | \
  grep -E 'runs-on: windows-latest|internal/cli/update|if:'
```

Expected: at least three matches — `runs-on: windows-latest`, a run step whose package argument
begins with `internal/cli/update`, and an `if:` gating the job on a path-filter output. A job that
runs `./...` on Windows at PR time fails this criterion (REQ-UCG-005).

**Scope-agnostic by construction (D6).** The match is on the `internal/cli/update` **prefix**, so
this criterion is satisfied by the full `./internal/cli/update/...` form *and* by the narrowed
`./internal/cli/update/backup/... ./internal/cli/update/merge/...` form that REQ-UCG-005 permits
when the measured wall-time requires it. The earlier wording demanded the full spelling in prose,
which meant taking plan.md §F M2's own narrowing option would have failed this criterion — the
requirement, the plan, and the criterion disagreed three ways. What this AC does *not* accept is a
Windows job scoped to `./...`, which is the actual hazard REQ-UCG-005 guards.

Baseline at `d5336214e` — no such job. The PR-scope test matrix is ubuntu-only
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
awk '/^            behavioral:$/{f=1;next} f{ if ($0 !~ /^              - /) exit; print }' \
  .github/workflows/ci.yml | grep -cE "internal/template/templates|\.moai/config"
```

Expected: a count `>= 2` — the filter that gates the test job lists both
`internal/template/templates/**` and `.moai/config/**` (REQ-UCG-001).

**Why an `awk` extractor and not a `sed` range (D8).** The original form was
`sed -n '/^            behavioral:/,/^            [a-z_]*:/p'`. A `sed` range whose closing pattern
never matches runs to end-of-file, and the gating filter is the **last** key in the `filters:` block,
so no closing key follows it. Measured against `go_code`, which sits in that same last position
today:

```
$ sed -n '/^            go_code:/,/^            [a-z_]*:/p' .github/workflows/ci.yml | wc -l
     232          # of ci.yml's 373 lines — the whole remainder of the file
$ awk '/^            go_code:$/{f=1;next} f{ if ($0 !~ /^              - /) exit; print }' \
    .github/workflows/ci.yml | wc -l
       6          # exactly the filter's six entries
```

The consequence was not cosmetic: the `grep` searched 232 lines of unrelated job definitions, so any
occurrence of `internal/template/templates` **anywhere below the filter block** would satisfy the
criterion while the filter itself stayed empty — the precise failure this AC exists to catch. The
`awk` form terminates at the first line that is not a 14-space list item, so it cannot leak past the
block regardless of which key is last.

Both controls were run to prove the extractor discriminates rather than merely returning a
convenient number:

```
$ awk '...go_code...'   .github/workflows/ci.yml | grep -cE "internal/template/templates|\.moai/config"
0     # negative control: the un-widened filter correctly scores 0
$ awk '...docs_only...' .github/workflows/ci.yml | grep -cE "\.moai/"
4     # positive control: the extractor does find matches when they are present
```

A negative control alone would be ambiguous — a broken extractor also returns 0. The positive
control is what shows the 0 is a real absence.

Baseline at `d5336214e` — the `go_code` filter (`ci.yml:65-71`) is six entries and contains
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
subtests from plan.md §F M4, split as follows (REQ-UCG-014):

| subtest | expected marker | why |
|---|---|---|
| `three_way_divergence` | `--- PASS` | hard assertion — measured correct at `d5336214e` |
| `template_only_key` | `--- PASS` | hard assertion — measured correct |
| `zero_value_false` | `--- PASS` | hard assertion — measured correct |
| `zero_value_empty_string` | `--- PASS` | hard assertion — measured correct |
| `zero_value_int` | `--- PASS` | hard assertion — measured correct |
| `user_only_key` | `--- SKIP` naming `SPEC-UPDATE-YAML-PRESERVE-001` | sibling-owned known defect — key dropped |
| `nested_old_only` | `--- SKIP` naming `SPEC-UPDATE-YAML-PRESERVE-001` | sibling-owned known defect — same cause, one level down |

**The SKIP set is exactly two, and neither is a zero-value case.** This corrects the criterion twice
over. The original wording expected the three **zero-value** subtests to be the skipped ones; the
plan audit rejected that premise correctly, but proposed `nested_old_only` as the single replacement.
Re-running the matrix shows **two** rows fail, because `DeepMerge3Way` iterates only `newMap`
(`merge.go:53`) and so never visits `user_only_key` either:

```
user_only_key    base{}       old{k:"v"}  new{}     -> map[]          key dropped
nested_old_only  base{a.b:1}  old{a.b:1}  new{a:{}} -> map[a:map[]]   a.b dropped
zero_value_false / _empty_string / _int                                all three correct
```

Had the audit's single-case fix been adopted verbatim, `user_only_key` would have been declared a
hard assertion and failed unexpectedly at run time — the same shape of error, one case narrower.

Falsification: §C-2, in two halves — the two SKIP rows must fail against the **unmodified** tree (no
mutation required, since the defect is real today), and the three zero-value rows must flip to FAIL
under the §C-2 mutation, proving their assertions are load-bearing rather than incidentally true.

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
git diff --stat d5336214e -- internal/cli/update/backup/merge.go
```

Expected: **no output** — REQ-UCG-014 and plan.md §D3 forbid this SPEC from editing the merge
implementation. Any diff means the SPEC crossed into `SPEC-UPDATE-YAML-PRESERVE-001`'s scope.

**Diff base corrected from `main` to `d5336214e` (D7).** The local branch `main` is not an ancestor
of this work and cannot serve as a diff base:

```
$ git branch -v --list main
+ main 1d4e4f7da [ahead 1, behind 12] ...
$ git merge-base --is-ancestor main HEAD ; echo "exit=$?"
exit=1                       # not an ancestor
```

`git diff --stat main -- <path>` therefore compared against a divergent tip rather than this SPEC's
baseline. It happened to print nothing, so the criterion appeared to pass — but it was measuring the
wrong thing, which §A clause 8 rejects regardless of the verdict it produced. §A clause 4 fixes the
baseline at `d5336214e`, and this criterion now names that commit, as do AC-UCG-019 and AC-UCG-023.

Baseline: no output at `d5336214e` (the file is unmodified relative to the code baseline).

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

Baseline at `d5336214e` — each shape escapes for a distinct structural reason, verified
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

#### AC-UCG-017 `[behavioural]` — the pattern extension is correct for the branch actually chosen — **decision-gated (expansion method)**

```bash
go test -run 'TestLeakPatternRejectsPedagogicalPlaceholders' -count=1 -v ./internal/template/
```

Expected: a `--- PASS: TestLeakPatternRejectsPedagogicalPlaceholders` line. **What that test must
assert depends on which branch of REQ-UCG-016's open decision is taken**, and the test carries the
subtests for the chosen branch only:

| Branch chosen (REQ-UCG-016) | Required subtests | What makes it fail |
|---|---|---|
| **generic wildcard + allowlist** | `rejects_placeholders` — the extended pattern does **not** match `SPEC-BUG-042`, `SPEC-X-001`, `SPEC-PAY-001` | removing or emptying the allowlist: all three placeholders match and the subtest names each one |
| **enumeration extension** | `matches_target_family` — the extended enumeration **does** match `SPEC-AGENT-ARCH-V2-001`; plus `rejects_placeholders` as a negative fixture | reverting `AGENT-ARCH-V2` out of the enumeration: `matches_target_family` fails naming the undetected shape |

**Why the criterion had to be branch-conditioned (D10).** The original wording asserted only the
allowlist half. That is load-bearing under the wildcard branch, where over-matching is the risk — but
under the enumeration branch it is **vacuous**: a narrow enumeration that names only
`AGENT-ARCH-V2` cannot match `SPEC-BUG-042` no matter how the allowlist is written, so the assertion
holds without testing anything, and its stated falsification ("a PASS means the allowlist is absent
or inert") could not distinguish a correct enumeration from a missing one. The enumeration branch's
real risk is the opposite failure — an extension too narrow to match the shape it was added for —
which `matches_target_family` is what catches.

Both branches keep `rejects_placeholders`; only the enumeration branch adds the positive subtest.
The three placeholder shapes are named verbatim in the existing rationale comment at
`internal/template/internal_content_leak_test.go:271-281` (REQ-UCG-017).

Falsification, per branch, is the "what makes it fail" column above; each MUST FAIL naming the
offending shape. The decision-gated marker is dropped once AC-UCG-018's measurement selects the
branch and the unused row is deleted from this criterion.

Baseline: no such test; the current narrow enumeration trivially does not match the placeholders,
and equally trivially does not match `SPEC-AGENT-ARCH-V2-001` — which is the gap REQ-UCG-015 exists
to close. Nothing asserts either fact today.

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
git diff --stat d5336214e -- internal/template/templates/.moai/config/
```

Expected: **no output**. Removing the three leaks is `SPEC-CONFIG-KEY-HONESTY-001` REQ-CKH-012's
work (spec.md §C); this SPEC only makes the shapes detectable. Diff base corrected from the
divergent `main` to `d5336214e` per AC-UCG-015 (D7).

Baseline: no output at `d5336214e`. The three leaks are present and unmodified:

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

Expected: a `--- PASS: TestLeakClassCitationsNameOwningFile` line. The test scans the **prose
citations** of `\bC[0-9][a-z]?\b` in four active files — the two guard test files, the neutrality
workflow, and `.moai/docs/template-internal-isolation-doctrine.md` — and asserts each appears within
a line that also names its owning file, so `C1` and `C6` are never ambiguous between the two
independently-numbered schemes (REQ-UCG-018, spec.md §A.6 drift 4).

**Two exclusions are mandatory, not optional (D11).**

1. **Class *definition* lines are excluded** — any line matching `name:\s*"C[0-9]`. A definition
   such as `name: "C1-spec-id-prefix",` inside `internal_content_leak_test.go` cannot be required to
   also name `internal_content_leak_test.go`: the requirement is self-contradictory, since the
   definition *is* the file's own declaration of the class.
2. **Historical SPEC and report artifacts are excluded** — `.moai/specs/**` and `.moai/reports/**`.
   Roughly twenty such files carry class citations from completed SPECs whose bodies are immutable.

Without these carve-outs the criterion is unsatisfiable. Measured at `d5336214e`:

```
$ grep -cE '\bC[0-9][a-z]?\b' internal_content_leak_test.go        →  60
$ grep -cE '\bC[0-9][a-z]?\b' template_neutrality_audit_test.go    →  34
$ grep -cE '\bC[0-9][a-z]?\b' template-neutrality-check.yaml       →   9      (103 total)
$ ... of which lines already naming an owning file                 →   2
$ ... of which are class definition lines (name: "C<N>-…")         →  18
```

2 of 103 compliant is not a milestone, it is an impossible target, and 18 of the shortfall could
never be closed at all. The scoped surface — comment and prose citations, definitions excluded — is
**77** across the four active files (35 + 25 + 9 + 8), which is what plan.md §F M6 now sizes.

Falsification: strip the filename qualifier from one **prose** citation and the test MUST FAIL naming
the file and line. Additionally, re-adding a class definition line to the scan MUST make the test
fail on that definition — demonstrating the exclusion is implemented rather than assumed. A PASS in
either case means the scan is not reaching the citations it claims to cover.

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
git diff d5336214e -- .github/workflows/ci.yml | grep -E '^[-+] *name: (Test|Lint|Build|Integration)' | sort
```

Expected: every `-` line has a matching `+` line with identical text, i.e. no required-check `name:`
value was renamed or removed (NFR-UCG-004). New jobs may add new `+` lines with no `-` counterpart;
a `-` line with no `+` counterpart fails this criterion. Diff base corrected from the divergent
`main` to `d5336214e` per AC-UCG-015 (D7).

Baseline: no diff today. The current required-check names are
`Test (${{ matrix.os }})` (emitted by both `test` at `ci.yml:87` and `test-skip-marker` at `:185`),
`Integration Tests (${{ matrix.os }})` (`:215`), `Lint` (`:247`), and
`Build (${{ matrix.goos }}/${{ matrix.goarch }})` (`:288`).

## §C Falsification procedures

Each new guard must be shown to FAIL against unfixed or deliberately-broken code. `git stash` is
prohibited (§A clause 5).

**All five procedures were inert as originally written, and all five are rebuilt here.** Three
(§C-3, §C-4, §C-5) expressed their mutation as a `#` shell comment, so the scratch tree equalled
`HEAD` and the test ran against unmodified code; §C-2's `sed` pattern matched zero lines in its
target file; §C-1 used `go test -overlay` to replace a file that is read at run time, which overlay
does not do. Each rewrite below states the executable mutation and carries a `diff -q` no-op guard.

### Choosing the mutation vehicle

| What must be mutated | Vehicle | Why |
|---|---|---|
| a `.go` source | `go test -overlay` | verified effective — mutating `merge.go` under overlay changes observed behaviour |
| a file read at **run time** (`coverage-baselines.yaml`, `ci.yml`) | scratch worktree + real edit + `go -C` | overlay replaces build inputs only (§A clause 9) |

### Shared preamble

Every worktree-based procedure below uses this setup and teardown. The worktree lives **inside the
project** so edits stay within the writable tree, and each edit is written as
`sed … > tmp && mv tmp target` rather than `sed -i`, whose in-place flag is spelled differently on
BSD and GNU.

```bash
WT=.claude/worktrees/ucg-falsify
git worktree add "$WT" HEAD          # detached checkout of the current commit
# … per-procedure mutation + diff -q guard + `go -C "$WT" test …` …
git worktree remove --force "$WT"
```

The no-op guard used after every mutation, which is what makes an inert procedure report itself
instead of printing a false PASS:

```bash
diff -q <original> <mutated> >/dev/null && { echo "MUTATION NO-OP — pattern matched nothing"; exit 1; }
```

### C-1 — the coverage gate actually catches a regression

```bash
WT=.claude/worktrees/ucg-falsify
git worktree add "$WT" HEAD
B="$WT/.moai/config/coverage-baselines.yaml"
sed 's/baseline: 75\.7/baseline: 75.8/' "$B" > "$B.tmp" && mv "$B.tmp" "$B"
git -C "$WT" diff --quiet -- .moai/config/coverage-baselines.yaml \
  && { echo "MUTATION NO-OP — pattern matched nothing"; git worktree remove --force "$WT"; exit 1; }
go -C "$WT" test -run 'TestCoverageBaselineNoRegression' -count=1 -v ./internal/quality/
git worktree remove --force "$WT"
```

Expected: **FAIL**, naming `internal/cli` with `baseline 75.8%, measured 75.7%`. A PASS means the
per-package comparison is not load-bearing and the gate would not catch real erosion either.

**Why not `go test -overlay` (D1).** The original procedure overlaid
`.moai/config/coverage-baselines.yaml` and expected the test to read the substituted file. Overlay
does not work that way — it replaces the *build*'s view of source files, not the bytes a running
test gets from `os.ReadFile`. Measured directly:

```
$ printf 'baseline: 75.7\n' > data.yaml ; printf 'baseline: 76.7\n' > fake.yaml
$ go test -overlay=overlay.json -run TestRuntimeRead -v .
    ov_test.go:13: runtime os.ReadFile saw: "baseline: 75.7\n"
--- PASS
```

The overlaid `76.7` never reached the test, under either a relative or an absolute overlay key. The
original procedure would therefore have run the gate against the **real** baseline, passed, and been
recorded as a successful falsification — the single most misleading outcome available, since a
passing falsification is indistinguishable from a decorative gate. The worktree form performs a real
edit, so there is nothing for the runtime read to miss.

**Bump size.** 0.1pp, not the original 1.0pp, so the procedure discriminates the open tolerance
decision instead of passing under every candidate — see AC-UCG-002 and spec.md §B.3.

### C-2 — the semantic merge guard actually catches a broken merge

Two halves, because the guard's seven rows divide into rows that already fail and rows that must be
shown capable of failing.

**C-2a — the two known-defective rows fail with no mutation at all.**

```bash
go test -run 'TestMergeSemanticOutcomes/user_only_key|TestMergeSemanticOutcomes/nested_old_only' \
  -count=1 -v ./internal/cli/update/backup/
```

Expected: both subtests report the sibling-owned defect — `--- SKIP` carrying
`SPEC-UPDATE-YAML-PRESERVE-001` once committed per REQ-UCG-014, and `--- FAIL` naming the dropped
key if the skip marker is removed. Nothing is mutated here because nothing needs to be: the defect
is live in the current tree, measured at `d5336214e`:

```
user_only_key    base{}       old{k:"v"}  new{}     -> map[]          key dropped
nested_old_only  base{a.b:1}  old{a.b:1}  new{a:{}} -> map[a:map[]]   a.b dropped
```

**C-2b — the five passing rows are shown to be load-bearing.**

```bash
D=/tmp/ucg-falsify ; mkdir -p "$D"
# neutralise the branch that decides the zero-value outcome: merge.go:87 `result[k] = newV`.
# Anchored on four leading tabs so the same-text lines at 56 and 65 are untouched.
sed 's/^\t\t\t\tresult\[k\] = newV$/\t\t\t\tresult[k] = oldV/' \
  internal/cli/update/backup/merge.go > "$D/merge.go"
diff -q internal/cli/update/backup/merge.go "$D/merge.go" >/dev/null \
  && { echo "MUTATION NO-OP — pattern matched nothing"; exit 1; }
printf '{"Replace":{"%s/internal/cli/update/backup/merge.go":"%s/merge.go"}}' "$PWD" "$D" > "$D/ov.json"
go test -overlay="$D/ov.json" -run 'TestMergeSemanticOutcomes' -count=1 -v ./internal/cli/update/backup/
```

Expected: **FAIL** on `zero_value_false`, `zero_value_empty_string`, and `zero_value_int` — each
returning the user's old value (`true` / `"x"` / `5`) where the new zero value was expected, named
with expected-versus-actual. `three_way_divergence` MUST still **PASS**: it takes the other branch of
the same `if`, so a mutation that also broke it would prove only that something changed, not which
assertion caught what.

**Why this mutation and not the original (D2).** The original stripped an `IsZero()` branch from
`internal/cli/update/backup/merge.go`. That file contains no such branch, so the `sed` matched
nothing and copied the file byte-for-byte:

```
$ grep -c 'IsZero()' internal/cli/update/backup/merge.go
0
$ sed '/IsZero()/,+2d' internal/cli/update/backup/merge.go > /tmp/merge.go
$ diff -q internal/cli/update/backup/merge.go /tmp/merge.go
(no output — identical, 175 lines both)
```

The overlay then substituted an identical file and the guard ran against the correct implementation.
The zero-value skip the original was reaching for is real, but lives in `internal/config/merge.go`
(`MergeAll:149`, `isZero:200`) — a different file, owned by `SPEC-CONFIG-TIER-PERSIST-001`, and out
of this SPEC's scope (spec.md §C). The mutation above targets a branch that actually exists in the
file under guard, and its effect was verified: under overlay it changes `DeepMerge3Way`'s observed
output.

Overlay is the correct vehicle here — unlike C-1, the mutated file is a `.go` source, which is
exactly what overlay replaces.

### C-3 — the filter-coverage test actually reads the filter

```bash
WT=.claude/worktrees/ucg-falsify
git worktree add "$WT" HEAD
# remove every internal/template/templates entry from the gating filter
sed '/internal\/template\/templates/d' "$WT/.github/workflows/ci.yml" > "$WT/ci.tmp" \
  && mv "$WT/ci.tmp" "$WT/.github/workflows/ci.yml"
git -C "$WT" diff --quiet -- .github/workflows/ci.yml \
  && { echo "MUTATION NO-OP — pattern matched nothing"; git worktree remove --force "$WT"; exit 1; }
go -C "$WT" test -run 'TestCIFilterCoversBehavioralPaths' -count=1 -v ./internal/template/
git worktree remove --force "$WT"
```

Expected: **FAIL**, naming `internal/template/templates/.moai/config/sections/workflow.yaml` as
unmatched. A PASS means the test parses the workflow but asserts nothing about the glob list, and
AC-UCG-010's bounded grep would be the only remaining defence — which §A clause 3 rejects.

### C-4 — the complement-coverage test actually catches a missing skip-marker

```bash
WT=.claude/worktrees/ucg-falsify
git worktree add "$WT" HEAD
# delete the whole test-update-windows-skip job block: from its 2-space job key
# up to the next 2-space job key.
awk '/^  test-update-windows-skip:$/{s=1;next} s && /^  [a-z-]+:$/{s=0} !s' \
  "$WT/.github/workflows/ci.yml" > "$WT/ci.tmp" && mv "$WT/ci.tmp" "$WT/.github/workflows/ci.yml"
git -C "$WT" diff --quiet -- .github/workflows/ci.yml \
  && { echo "MUTATION NO-OP — job key not found"; git worktree remove --force "$WT"; exit 1; }
go -C "$WT" test -run 'TestCIRequiredCheckComplementCoverage' -count=1 -v ./internal/template/
git worktree remove --force "$WT"
```

Expected: **FAIL**, naming the required-check name whose complement path is uncovered. A PASS means
a pull request could be rendered permanently unmergeable without any test noticing — the failure
mode `ci.yml:186-198` already documents from a prior occurrence.

The block-deleter was rehearsed against the existing `test-skip-marker` job, since
`test-update-windows-skip` does not exist until M2 lands: it removed 30 lines, eliminated the target
job, and left the following `test-integration` job intact — so the `s=0` boundary rule stops at the
right place rather than truncating the remainder of the file.

### C-5 — the pattern extension does not silently over-match

Applies to the **generic-wildcard branch** of REQ-UCG-016 (see AC-UCG-017). On the
enumeration-extension branch, substitute the mutation named in that criterion's second row —
reverting `AGENT-ARCH-V2` out of the enumeration — and expect `matches_target_family` to FAIL.

```bash
WT=.claude/worktrees/ucg-falsify
git worktree add "$WT" HEAD
# replace the enumerated SPEC-ID class with a bare wildcard, removing the allowlist's subject
sed 's/SPEC-(V3R\[2-6\]|AGENCY|WORKTREE)-\[A-Z0-9-\]+/SPEC-[A-Z][A-Z0-9-]*-[0-9]{3}/' \
  "$WT/internal/template/internal_content_leak_test.go" > "$WT/leak.tmp" \
  && mv "$WT/leak.tmp" "$WT/internal/template/internal_content_leak_test.go"
git -C "$WT" diff --quiet -- internal/template/internal_content_leak_test.go \
  && { echo "MUTATION NO-OP — pattern matched nothing"; git worktree remove --force "$WT"; exit 1; }
go -C "$WT" test -run 'TestLeakPatternRejectsPedagogicalPlaceholders' -count=1 -v ./internal/template/
git worktree remove --force "$WT"
```

Expected: **FAIL**, naming `SPEC-BUG-042`, `SPEC-X-001`, and `SPEC-PAY-001` as wrongly matched. A
PASS means the false-positive guard is decorative and REQ-UCG-016's measured trade-off is
unenforced.

The `sed` was verified to land on the current tree, rewriting exactly one line:

```
171c171
< 		pattern: regexp.MustCompile(`\bSPEC-(V3R[2-6]|AGENCY|WORKTREE)-[A-Z0-9-]+\b`),
---
> 		pattern: regexp.MustCompile(`\bSPEC-[A-Z][A-Z0-9-]*-[0-9]{3}\b`),
```

**Why §C-3 / §C-4 / §C-5 all needed rebuilding (D3).** In each, the mutation step was a line
beginning with `#` — a shell comment. The scratch worktree was therefore a verbatim checkout of
`HEAD`, the test ran against unmodified code, and the expected FAIL could not occur. All three would
have reported PASS and been recorded as successful falsifications. The scratch tree also lived at
`/tmp/ucg-wt*`, outside the project, where edits can be refused; the worktrees are now created under
`.claude/worktrees/` and removed with `--force` so a mutated tree cannot block teardown.

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

- All of AC-UCG-001 through AC-UCG-023 produce their stated observable output. Two are
  **decision-gated** (AC-UCG-002 on the tolerance value, AC-UCG-017 on the expansion method); each
  is satisfied by the branch selected at the Implementation Kickoff Approval gate, and its
  decision-gated marker is dropped once the unused branch is deleted from the criterion.
- All five falsification procedures C-1 through C-5 produce **FAIL** against broken code, and each
  run reaches its assertion rather than aborting on its `diff -q` no-op guard. A procedure that
  prints `MUTATION NO-OP` has not been run — it has reported that it cannot run, which is the
  failure mode that made all five inert before this revision (§A clause 8).
- C-2 is satisfied in both halves: C-2a shows `user_only_key` and `nested_old_only` failing against
  the **unmodified** tree, and C-2b shows the three zero-value rows flipping to FAIL under mutation
  while `three_way_divergence` stays green.
- `internal/cli/update/backup/merge.go` is unmodified relative to `d5336214e` (AC-UCG-015).
- `internal/template/templates/.moai/config/` is unmodified relative to `d5336214e` (AC-UCG-019).
- No required-check `name:` value was renamed or removed (AC-UCG-023).
- `progress.md` §E.2 cites the observed command output for every claim, per
  `.claude/rules/moai/core/verification-claim-integrity.md`, including the measured Windows-leg
  wall-time (AC-UCG-009) and the measured false-positive counts (AC-UCG-018).
