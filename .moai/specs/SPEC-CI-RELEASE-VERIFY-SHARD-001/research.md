# research.md — SPEC-CI-RELEASE-VERIFY-SHARD-001

Card t1892 · plan-phase tree pin: **`74ad197380`** (`git rev-parse --short HEAD`, branch `WT-release-verify-os`, clean at authoring; = PR #1830 head). Every entry below was measured in this plan run on that tree unless it names a remote run id (those are remote observations read through `gh`, not re-executable offline, and are labelled so). Verbatim outputs are quoted; long listings are bounded and the bound is stated. Where the tool surface hid an exit code (a passing command prints none) the exit is recorded as 0 only for commands that produced output or no error; negative controls name their exit explicitly.

## §A Findings (what the evidence establishes)

1. **Root cause is a single serial step against a fixed budget, not a flaky test.** One step runs `go test -json -race -timeout 35m ./...` on each OS inside a 40-minute job (E1, E2). Run 39180942620 cancelled all three legs in that step at 40 minutes (E4); two earlier dispatch runs show the same shape (E5) and the one ubuntu leg that passed did so in 38.2 minutes (E5). The budget is exhausted by workload growth, so raising the timeout (forbidden) would only move the cliff.
2. **The required check is the gate, not the legs.** Branch protection requires five contexts; `Release Verify (...)` is not among them (E6). Renaming the matrix job therefore changes no required-check name. `Release PR Multi-OS Gate` is required and must not change (E6).
3. **Two non-protection consumers read the leg names** and constrain the design: `scripts/release.sh --require-matrix-run` looks up the exact strings `Release Verify (ubuntu-latest)` etc. (E7), and a Python verifier extracts the gate's step text and substitutes `needs.full-matrix-test.result` (E8). Keeping the job id `full-matrix-test` and the gate byte-identical satisfies the second; the first must be updated for sharded names (REQ-RVS-011).
4. **ci.yml's selectors are an exact cover of internal/cli today** (E9): 2111 + 862 + 1851 + 1157 = 5981 = the full listing, no overlap, no gap; and the checker is not vacuous — a one-letter mutant is caught (E10).
5. **The rest of the portability surface is mostly already exercised.** `jq` is on all three runner images (E11); the census and `retry-flaky.sh` have no bash-4-only constructs (E12); macOS `bash` is 3.2.57 (E11). What is unexercised on macOS/windows is `retry-flaky.sh` itself — a reason the plan does not adopt it (decision-index Q5).
6. **Cost premise differs from the stated budget.** The repository is public (E3), so standard hosted-runner minutes are free regardless of the stated 10 USD/month; the real constraint on 18 legs is concurrency — the org is on the Team plan, macOS jobs cap at 5 concurrent (E3), so the 6 macos shards queue one deep. Recorded for the leader; the SPEC keeps six shards and no duplicate runs either way.

## §B Evidence ledger

| ID | Claim | Command (single invocation) | Exit | Tree / source |
|----|-------|------------------------------|------|---------------|
| E1 | Release workflow jobs and the matrix job's steps | `yq '.jobs \| keys' .github/workflows/release-pr-multi-os.yml` and `yq '.jobs.full-matrix-test.steps[].name' .github/workflows/release-pr-multi-os.yml` | 0 | 74ad197380 |
| E2 | Timeout values in the release workflow | `grep -n -e 'timeout' .github/workflows/release-pr-multi-os.yml` | 0 | 74ad197380 |
| E3 | Repo visibility, org plan, concurrency limits | `gh api repos/modu-ai/moai-adk --jq '{private, visibility}'`; `gh api orgs/modu-ai --jq '{plan: .plan.name, public: .public_repos}'`; GitHub docs fetch (limits table) | 0 | live read 2026-10-19 |
| E4 | PR #1830 run 39180942620: all legs cancelled in the test step | `gh run view 39180942620 --repo modu-ai/moai-adk --json conclusion,createdAt,updatedAt,jobs --jq '...'` | 0 | remote run |
| E5 | Earlier release-verify runs with real matrix execution | `gh run list --repo modu-ai/moai-adk --workflow release-pr-multi-os.yml --limit 300 --json ... --jq '<duration > 300s>'` and `gh run view <id> --json jobs --jq '...'` for 38363720321 and 37861821414 | 0 | remote runs |
| E6 | Branch protection required contexts (read-only GET succeeded) | `gh api repos/modu-ai/moai-adk/branches/main/protection --jq '{contexts: .required_status_checks.contexts, checks: .required_status_checks.checks, strict: .required_status_checks.strict}'` | 0 | live read 2026-10-19 |
| E7 | release.sh leg lookup is exact-name | `bash .moai/specs/SPEC-CI-RELEASE-VERIFY-SHARD-001/repro/leg-match-probe.sh <worktree-root>` | 0 | 74ad197380 |
| E8 | Python verifier couples to job ids and gate step 0 | `python3 scripts/ci-mirror/test/verifier-fail-closed_test.py` | 0 | 74ad197380 |
| E9 | ci.yml internal/cli selectors are an exact cover | `bash .moai/specs/SPEC-CI-RELEASE-VERIFY-SHARD-001/repro/shard-cover.sh <worktree-root> <outdir> .github/workflows/ci.yml` | 0 | 74ad197380 |
| E10 | Mutant: shard 3 `[J-R]` -> `[J-Q]` is caught | same script on a mutated copy of ci.yml | 1 | 74ad197380 + mutation |
| E11 | Runner image tooling (jq, yq, bash, Git Bash) | WebFetch of runner-images Windows2025 / macos-15-arm64 readmes | n/a | remote docs 2026-10-19 |
| E12 | No bash-4-only constructs in the helper scripts | `grep -nE 'declare -A\|mapfile\|readarray\|<<<\|...' scripts/ci/retry-flaky.sh scripts/ci-census/test-census.sh` | 1 (no match) | 74ad197380 |
| E13 | Baseline: actionlint clean, validator green, template tests green | `actionlint .github/workflows/release-pr-multi-os.yml`; `sh scripts/ci-mirror/validate-required-checks.sh`; `go test -count=1 -run 'TestReleaseScriptMatrixGate\|TestTestInput' ./internal/template/` | 0 / 0 / 0 | 74ad197380 |
| E14 | RED-now shape cells (matrix, artifact, parity test, empty-sweep, -count) | see ledger blocks L-E14a..f, L-E17, L-E18 | L-E14a 0; L-E14b 0; L-E14c 0; L-E14d 1; L-E14e 1 and 1; L-E14f 9; L-E17 0 and 0; L-E18 1 | 74ad197380 |
| E15 | `shard-cover.sh` falsifiers (iter1 fix D5): print-then-fail stub `go` (old script reads it as success, fixed script fails), overlap mutant | see L-E15 | 0 old / 1 fixed / 1 | 74ad197380 + stub/mutation |
| E16 | RED-now for the local-bash step-body criterion: no step passes the selector through `SHARD_RUN` | `grep -c SHARD_RUN .github/workflows/release-pr-multi-os.yml` | 1 | 74ad197380 |

### Verbatim ledger

```text
L-E1  | yq '.jobs | keys' .github/workflows/release-pr-multi-os.yml   (comment lines elided; 3 keys)
- detect-release
- full-matrix-test
- release-pr-gate

L-E1b | yq '.jobs.full-matrix-test.steps[].name' .github/workflows/release-pr-multi-os.yml   (11 lines, full)
Checkout code
Setup Go
Download dependencies
Install ast-grep (provides `sg`)
Verify `sg` resolves to ast-grep
Run go vet
Diagnose ancestor .moai (pre-test)
Run tests with race detector
Compress test event stream
Upload test event stream
Diagnose ancestor .moai (post-test)

L-E2  | grep -n -e 'timeout' .github/workflows/release-pr-multi-os.yml   (9 lines, full)
41:    timeout-minutes: 5
94:    timeout-minutes: 40
155:      # -timeout raised above Go's 10m per-package default. internal/cli spends
159:      # the job's own timeout-minutes; this value exists to stop a package from
210:          # -timeout is per test binary, not per job. internal/cli under -race
215:          # inside this job's timeout-minutes: 40 budget. Applies to every
217:          go test -json -race -timeout 35m ./... > test-stream.json || rc=$?
279:    timeout-minutes: 5

L-E3  | gh api repos/modu-ai/moai-adk --jq '{private, visibility}'
{"private":false,"visibility":"public"}
      | gh api orgs/modu-ai --jq '{plan: .plan.name, public: .public_repos}'
{"plan":"team","public":9}
      | GitHub docs "Actions limits" table, standard runners (WebFetch 2026-10-19):
Free 20 total / 5 macOS; Pro 40 / 5; Team 60 / 5; Enterprise 500 / 50. Matrix cap: 256 jobs per workflow run.

L-E4  | run 39180942620 (jobs; steps filtered to non-success/non-skipped), createdAt 2026-10-19T00:28:41Z, updatedAt 2026-10-19T01:09:45Z, conclusion cancelled
Detect Release PR              success   00:28:44Z -> 00:29:04Z
Release Verify (ubuntu-latest)  cancelled 00:29:07Z -> 01:09:24Z  step "Run tests with race detector" cancelled
Release Verify (macos-latest)   cancelled 00:29:12Z -> 01:09:38Z  step "Run tests with race detector" cancelled
Release Verify (windows-latest) cancelled 00:29:08Z -> 01:09:24Z  step "Run tests with race detector" cancelled
Release PR Multi-OS Gate        failure   01:09:41Z -> 01:09:44Z  step "Verify all OS legs passed" failure

L-E5  | runs whose duration exceeded 300 s (excerpt of the first 4 of the 20 printed; the remaining are 2026-10-07 successes of 6-18 min that predate the workload growth)
39180942620 cancelled pull_request 2026-10-19T00:28:41Z 2026-10-19T01:09:45Z
38363720321 failure workflow_dispatch 2026-10-13T07:36:25Z 2026-10-13T08:17:37Z
37861821414 failure workflow_dispatch 2026-10-08T23:52:34Z 2026-10-09T00:33:27Z
37658342973 success pull_request 2026-10-07T17:20:49Z 2026-10-07T17:27:35Z
      | run 38363720321 jobs:
Release Verify (windows-latest) cancelled 07:36:51Z -> 08:17:16Z failedsteps=Run tests with race detector
Release Verify (ubuntu-latest)  failure   07:36:52Z -> 08:16:13Z failedsteps=Run tests with race detector
Release Verify (macos-latest)   cancelled 07:36:59Z -> 08:17:28Z failedsteps=Run tests with race detector
      | run 37861821414 jobs:
Release Verify (ubuntu-latest)  success   23:53:03Z -> 00:31:13Z   (38 min 10 s of a 40 min budget)
Release Verify (windows-latest) cancelled 23:53:03Z -> 00:33:20Z failedsteps=Run tests with race detector
Release Verify (macos-latest)   failure   23:53:08Z -> 00:32:57Z failedsteps=Run tests with race detector

L-E6  | gh api repos/modu-ai/moai-adk/branches/main/protection --jq '{contexts: ..., checks: ..., strict: ...}'
{"checks":[{"app_id":15368,"context":"Test (ubuntu-latest)"},{"app_id":15368,"context":"Lint"},{"app_id":15368,"context":"Build (linux/amd64)"},{"app_id":15368,"context":"Analyze (Go) (go)"},{"app_id":15368,"context":"Release PR Multi-OS Gate"}],"contexts":["Test (ubuntu-latest)","Lint","Build (linux/amd64)","Analyze (Go) (go)","Release PR Multi-OS Gate"],"strict":false}

L-E7  | bash .moai/specs/SPEC-CI-RELEASE-VERIFY-SHARD-001/repro/leg-match-probe.sh /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892   (exit 0 = RED state as designed; as first committed this probe ran a hard-coded copy of the lookup — the version in the L-E16 block extracts the program text from the real file; it still probes the extracted expression, not the whole script)
exact job name        -> found|success
shard-suffixed names  -> <no leg>
RED: the lookup finds no leg among shard-suffixed job names
      | consumer references to leg names / job id (grep -I over the repo, excluding .git/.moai/node_modules): scripts/release.sh:360 (`for leg in "Release Verify (ubuntu-latest)" "Release Verify (macos-latest)" "Release Verify (windows-latest)"`), internal/template/release_script_test.go:223-225,248 (stub job names), scripts/ci-mirror/test/verifier-fail-closed_test.py:21 (job id `full-matrix-test`), internal/template/test_input_filter_parity_test.go:83 (string `needs.detect-release.outputs.go_code != 'false'`).

L-E8  | python3 scripts/ci-mirror/test/verifier-fail-closed_test.py
test_auxiliary_overlap_still_fails ... ok
test_platform_gate ... ok
test_release_gate ... ok
test_required_context_extraction_failure ... ok
test_required_context_structure ... ok
Ran 5 tests in 6.406s
OK

L-E9  | bash .moai/specs/SPEC-CI-RELEASE-VERIFY-SHARD-001/repro/shard-cover.sh <worktree-root> <scratch> .github/workflows/ci.yml   (go test -list per selector vs the full list; load average 23 on 16 cores, sequential, no background load)
selector 1  ^Test([A-E]|Fa)  -> 2111 names
selector 2  ^Test(F([^a]|$)|[G-I])  -> 862 names
selector 3  ^Test[J-R]  -> 1851 names
selector 4  ^Test([S-Z]|[^A-Za-z])  -> 1157 names
full=5981 union=5981 union_unique=5981
EXACT COVER: full list == union of the four selectors (no gap, no overlap)
      | side observation: no bare `Test` name and no non-letter first character exists on this tree (grep -c '^Test$' = 0; grep -c -E '^Test[^A-Za-z]' = 0), so the `[^A-Za-z]` arm of selector 4 is empty today; it covers future names by construction.
      | package count: `go list ./...` -> 176 packages on this tree (`go list ./... | grep -c .`), of which `internal/cli` and `internal/harness/rosterguard` are two, so the remainder shard covers 174.

L-E10 | same script, ci.yml copy with `'^Test[J-R]'` rewritten to `'^Test[J-Q]'`   (exit 1 = the checker detects the mutant)
selector 3  ^Test[J-Q]  -> 1042 names
full=5981 union=5172 union_unique=5172
NOT AN EXACT COVER (first differences):
4016,4824d4015
< TestR1_P1_3_MergeVerbRefusesCodexEdge
< TestR1_P1_5_MergeTargetFollowsTheWindowRecord
(809 names dropped; first 2 shown, output bounded)

L-E11 | runner images (docs fetch): Windows2025 readme -> "jq 1.8.2", "Git 2.56.0.windows.2", "Bash 5.3.15(2)-release"; macos-15-arm64 readme -> "jq 1.8.2", "yq 4.53.6", "Bash 3.2.57(1)-release". (Image versions drift; these are 2026-10-19 reads, not a guarantee for the run date.)

L-E12 | grep -nE 'declare -A|mapfile|readarray|<<<|\$\{[A-Za-z_]+(,,|\^\^)\}|&>>|coproc' scripts/ci/retry-flaky.sh scripts/ci-census/test-census.sh
(no output)        exit 1  -> no bash-4-only construct found by pattern (static check; not an execution on macOS bash 3.2 — see Gaps)

L-E13 | actionlint -version -> v1.7.10 (built from source, go1.25.6); actionlint .github/workflows/release-pr-multi-os.yml -> (no output); sh scripts/ci-mirror/validate-required-checks.sh -> "✅ All validations passed" (incl. "[main] Release PR Multi-OS Gate is publishable"); go test -list 'TestReleaseScriptMatrixGate|TestTestInput' ./internal/template/ -> TestReleaseScriptMatrixGate, TestTestInputFiltersSingleSource, TestTestInputFiltersCoversGoTestInputs (3 names: the sweep is not empty); go test -count=1 -run ... ./internal/template/ -> "ok  github.com/modu-ai/moai-adk/internal/template  17.983s"

L-E14a | yq '.jobs.full-matrix-test.strategy.matrix.shard | length' .github/workflows/release-pr-multi-os.yml
0        (exit 0)
L-E14b | yq '.jobs.full-matrix-test.strategy.matrix | keys' .github/workflows/release-pr-multi-os.yml
- os        (exit 0)
L-E14c | yq '.jobs.full-matrix-test.steps[] | select(.name == "Upload test event stream") | .with.name' .github/workflows/release-pr-multi-os.yml
test-stream-release-verify-${{ matrix.os }}        (exit 0)
L-E14d | ls internal/template/release_verify_shard_parity_test.go
ls: internal/template/release_verify_shard_parity_test.go: No such file or directory        (exit 1)
L-E14e | grep -c 'empty sweep' .github/workflows/release-pr-multi-os.yml  -> 0  (exit 1);  grep -c -e '-count=1' .github/workflows/release-pr-multi-os.yml -> 0  (exit 1)
L-E14f | bash .moai/specs/SPEC-CI-RELEASE-VERIFY-SHARD-001/repro/shard-cover.sh <worktree-root> <scratch> .github/workflows/release-pr-multi-os.yml
FAIL: expected 4 internal/cli selectors in .github/workflows/release-pr-multi-os.yml, found 0        (exit 9)
```

```text
L-E15 | stub go (scratch bin/go: answers `go test -list REGEX pkg` from names TestAlpha TestFoo TestJk TestSz, prints an ok line, then exits 42) put first on PATH
  OLD shard-cover.sh (as committed in f6f502a84c; pipelines return grep's status):
selector 1  ^Test([A-E]|Fa)  -> 1 names
selector 2  ^Test(F([^a]|$)|[G-I])  -> 1 names
selector 3  ^Test[J-R]  -> 1 names
selector 4  ^Test([S-Z]|[^A-Za-z])  -> 1 names
full=4 union=4 union_unique=4
EXACT COVER: full list == union of the four selectors (no gap, no overlap)        exit 0   <- report-not-verdict defect (D5)
  FIXED shard-cover.sh (status of each go test -list captured before filtering):
FAIL: go test -list exited 42 on the full listing (a failed listing is not evidence)        exit 1
  FIXED script, real go, ci.yml: selectors 2111 / 862 / 1851 / 1157, full=5981 union=5981 union_unique=5981, EXACT COVER, exit 0
  FIXED script, overlap mutant (ci.yml copy with shard 2 `|[G-I])` -> `|[F-I])`):
selector 2  ^Test(F([^a]|$)|[F-I])  -> 1130 names
full=5981 union=6249 union_unique=5981
NOT AN EXACT COVER (first differences): ...        exit 1   (268 names selected twice; output bounded)

L-E16 | grep -c SHARD_RUN .github/workflows/release-pr-multi-os.yml
0        (exit 1)
      | leg-match-probe.sh (iter1 fix D6) now extracts the awk program text from the real scripts/release.sh and runs it:
extracted legs line : 360:            for leg in "Release Verify (ubuntu-latest)" "Release Verify (macos-latest)" "Release Verify (windows-latest)"; do
extracted awk prog  : $1 == leg {print "found|" $2; exit}
exact job name        -> found|success
shard-suffixed names  -> <no leg>
RED: the lookup finds no leg among shard-suffixed job names        exit 0
      | Example/Fuzz presence: grep -lE '^func (Example|Fuzz)' internal/cli/*_test.go internal/harness/rosterguard/*_test.go -> (no output); none exist today. The run-phase guard (REQ-RVS-016) follows Go's discovery rule, not a name-prefix ban: a fuzz target (`*testing.F`) or an example with an output comment is run by go test; a lowercase-suffixed `Fuzz...` helper or an output-less example is not (plan-audit-iter2.md D2, codex fixture; not re-run here).

L-E17 | yq '.jobs.full-matrix-test.strategy.matrix.include | length' .github/workflows/release-pr-multi-os.yml
0        (exit 0)
      | yq '.jobs.full-matrix-test.strategy.matrix | has("exclude")' .github/workflows/release-pr-multi-os.yml
false    (exit 0)

L-E18 | grep -c 'internal/harness/rosterguard' .github/workflows/release-pr-multi-os.yml
0        (exit 1)

L-E19 | duplicate-include mutant (scratch matrix with two `shard: cli-1` include entries, six shard ids, three OS)
      | yq '.jobs.full-matrix-test.strategy.matrix.include | length' <scratch mutant>
7        (exit 0; expected 6 -> the structural assertion fails)
      | yq '[.jobs.full-matrix-test.strategy.matrix.include[].shard] | unique | length' <scratch mutant>
6        (exit 0; 7 entries over 6 distinct shard ids -> a repeat)
      | leg count is NOT changed by the repeat: a shard-keyed include entry overwrites no original matrix value, so the second copy merges into the same combinations (GitHub include rule; `scripts/ci-mirror/validate-required-checks.sh` GATE-6/GATE-7 comment block, lines ~389-397) — cited from the plan-audit iter3 finding and the script's comment, not re-run through the expander.

L-E20 | what `go test -list` enumerates (scratch package in the session scratchpad, 8 functions + 1 helper; run once before the quiet window, not repeated): `go -C <scratch> test -vet=off -list '.*' ./...`
TestReal
Test
BenchmarkB
Fuzz
FuzzReal
Example
ExampleLower
ExampleBlock
ExampleUnordered
ok  	gl	0.291s        (exit 0)
      | source had: bare `Test`, bare `Fuzz`, bare `Example` (with an empty `// Output:`), `Fuzzhelper(s string) string`, `ExampleAlpha()` with no output comment, `ExampleLower` with `// output: x`, `ExampleBlock` with `/* Output: */`, `ExampleUnordered` with `// Unordered output:`. Not listed: `Fuzzhelper`, `ExampleAlpha`. Listed: every other one, including bare names and the lower-case and block-comment output markers. (Without `-vet=off`, `go test` fails the build on the vet "tests" analyzer's unknown-identifier complaints for `ExampleAlpha` etc.; the run-phase guard must call `go test -list` the way ci.yml's race legs do and not depend on this scratch flag.)
      | bare `Test` regexp semantics: measured by the plan auditor with Go regexp (plan-audit-iter1.md evidence row 11): `Test` matches no ci.yml selector; `TestF` matches 2; `TestFa` 1; `Test_x`/`Test1`/`TestÀ` shard 4 — cited, not re-run here.
```

## §C Required analyses

### C.1 Job and step structure; what branch protection requires

Current structure (L-E1, L-E1b): `detect-release` -> `full-matrix-test` (matrix `os` only, one 11-step job per OS) -> `release-pr-gate`, whose `needs` is `["full-matrix-test","detect-release"]`. The gate evaluates `needs.full-matrix-test.result`, which GitHub aggregates over every matrix leg: any `failure` yields `failure`, otherwise any `cancelled` yields `cancelled`, and only all-`success` yields `success`. Because a matrix job's result already aggregates its legs, widening the matrix from 3 to 18 legs inside the same job id leaves the gate's text and semantics untouched.

Required contexts of `main` (L-E6, read-only GET, permitted in this session): `Test (ubuntu-latest)`, `Lint`, `Build (linux/amd64)`, `Analyze (Go) (go)`, `Release PR Multi-OS Gate`. **`Release Verify (...)` is not required**, so adding a shard component to the matrix job name changes no required-check name. The validator (`scripts/ci-mirror/validate-required-checks.sh`) expands `${{ matrix.* }}` including `include` entries and confirmed the gate name publishable at baseline (L-E13); it is re-run after the change (AC-RVS-005).

### C.2 Selector coverage

`ci.yml` internal/cli selectors verified as an exact cover of the 5981 top-level tests (L-E9), and the checker is falsifiable (L-E10). Fresh-tree caveat: the count is a measurement of tree `74ad197380`; the check is re-run, not re-quoted, whenever a selector or the test set changes (verification-completeness §4).

### C.3 Per-OS timing expectations

Evidence available:
- Ubuntu, single serial leg: 38 min 10 s (success) on 2026-10-09, cancelled at 40 min on 2026-10-13 and 2026-10-19 (L-E5, L-E4). The leg had no slack.
- Ubuntu per-shard proxies (from `ci.yml`'s comment, run 37861953651, measured before several later test additions): cli-1 605 s, cli-2 408 s, cli-3 532 s, cli-4 496 s, rosterguard 782 s, remainder's largest package 383 s. Each is far under 35 min.
- macos and windows: **no per-shard measurement exists** (gap G2). Both legs were cancelled at 40 minutes in every recent full run (L-E4, L-E5), so only a lower bound is known. A 2x-3x slowdown on the heaviest shard (rosterguard, 782 s on ubuntu) would still land at roughly 26-39 minutes — inside the unchanged 35 min test ceiling only at the low end. This is a real residual risk, owned by the first live run (REQ-RVS-013), and the response if it materialises is a rebalance card, never a timeout raise.

### C.4 Options

**A — one matrix job, shard x os via `include` (job id `full-matrix-test` kept).** 18 legs; the gate, its `needs`, the job id, and the `go_code != 'false'` condition are untouched; the Python verifier and the filter-parity test keep passing without edits. Cost: job names gain a shard suffix, so `scripts/release.sh --require-matrix-run` (exact-name lookup, L-E7) and its Go test need an update (REQ-RVS-011). Selectors are duplicated from `ci.yml`, guarded by a parity test (REQ-RVS-003).

**B — shared selector source (script or composite action) consumed by both workflows.** Removes selector duplication at the source. But the full benefit needs `ci.yml` edited (it carries required-check jobs, and the task scope is the release workflow), the composite/script must run under macOS bash 3.2 and Git Bash, and it adds a new shared surface that every future CI change must understand. A shared source is also a larger blast radius than a parity test for the same drift risk.

**C (examined, rejected) — three per-OS shard jobs plus three aggregator jobs named exactly `Release Verify (<os>)`.** Would keep `release.sh` untouched, but `needs.<job>.result` cannot be split per OS from a matrix, so per-OS aggregation forces three separate job definitions (or a composite action), and the gate's `needs` and text must change — the one surface the operator forbade relaxing. Rejected.

**Recommendation: A with a parity test** (decision-index Q1, Q3). Reasons: (1) the required gate stays byte-identical, which is the operator's hardest constraint; (2) one job definition, one place to read; (3) drift between the two workflows is converted from a silent hazard into a failing repository test that also checks exact cover over the real `internal/cli` test names, using the same regexp engine `go test -run` uses; (4) it adds no shared runtime surface to the Windows/macOS legs; (5) B remains available later as a pure refactor behind the same parity test.

### C.5 Windows and macOS specifics

- **Shell.** The test step already declares `shell: bash` (Git Bash on windows, system bash 3.2.57 on macos — L-E11). The shard step keeps `shell: bash` and passes the selector as an environment variable (`SHARD_RUN`), so regex metacharacters (`|`, `$`, `^`, `[^…]`) are never re-parsed by a shell on any OS. MSYS path conversion applies to arguments that look like POSIX paths; none of the selectors starts with `/`, and the post-landing run is the observation (AC-RVS-014).
- **retry-flaky.sh** is bash and has no bash-4-only construct (L-E12), but it has only ever run on ubuntu (every call site is a `ci.yml` ubuntu job). Its registry is empty today ("Entries: none"), so it would be an inert wrapper that adds an unexercised buffering layer and, once an entry is added, a retry on a required gate — a relaxation the operator forbade. Not adopted (Q5).
- **`go list ./... | grep -v -x`** (remainder shard) uses only POSIX utilities available in Git Bash; the leg refuses an empty package list, and the test census refuses an empty stream, so a mis-listing fails instead of passing.
- **Windows artifact rule** (`if-no-files-found: error` on windows) is preserved (REQ-RVS-012).
- **Concurrency.** 6 macos shards vs a 5-job macOS cap on the Team plan: one shard queues behind the others (L-E3). Wall time rises by up to one shard's runtime on macos; no correctness effect, and `fail-fast: false` keeps every shard reporting.

## §D Gaps (explicitly not observed)

- **G1** — macOS/windows execution of the new step body: not run (this machine is darwin but the runner images differ; the step body is exercised locally under `/bin/bash` 3.2.57 as a run-phase AC, AC-RVS-012, with a tiny package, not a runner execution).
- **G2** — per-shard macOS/windows duration: no measurement exists (C.3).
- **G3** — runner image tool versions are 2026-10-19 doc reads; they can move before the run date.
- **G4** — whether `release.sh --require-matrix-run` is exercised by any currently open release flow: not queried; the update is required regardless because the flag fails closed on sharded names (L-E7).
- **G5** — the `Test` bare-name / non-letter-first edge: confirmed empty on this tree (L-E9 side observation); not a guarantee for future names — the parity test (not this count) is the continued-firing answer (verification-completeness §1.3).
- **G6** — PR #1830's `pull_request` run uses the workflow file at its merge ref; whether the change reaches that ref (and how) is a leader/landing matter, outside the lane.

## §E Residual risk

- A shard on windows or macos exceeding the unchanged 35 min test ceiling would fail visibly (no longer a silent whole-leg cancel) but still block the release. Mitigation is a follow-up rebalance card; the SPEC refuses to pre-empt it with a timeout raise.
- Selector duplication is guarded, not removed: a drift fails the parity test only on CI runs that execute `internal/template` tests, which `ci.yml`'s go-code filter triggers for both workflow files.
