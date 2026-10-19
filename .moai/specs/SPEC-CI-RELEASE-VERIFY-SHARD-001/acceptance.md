# acceptance.md — SPEC-CI-RELEASE-VERIFY-SHARD-001

Tier M verification layer. Document-level tree pin for every RED cell: **`74ad197380`** (binds every criterion carrying no pin of its own). The four-element RED cells (command, verbatim stdout, exit code, tree) live in the evidence ledger in `research.md` §B and are cited here by id — the ledger is the carrier (verification-completeness §2.1). Criteria are the verification layer: Given-When-Then, binary-testable. The GEARS requirements they verify are the REQ-RVS entries in `spec.md`.

Classification: **RB** release-blocking (RED re-executable on this tree and flipped by this SPEC's run phase); **RG** regression-guard (an invariant that is green now and must stay green; a mutant that breaks it is named); **PL** post-landing observation (made by the leader after landing; not a lane gate).

## §B. Acceptance criteria

### AC-RVS-001 — Six shards per OS run as one matrix job (RB; maps REQ-RVS-001, REQ-RVS-008)

- **Given** the release workflow at the run-phase tree,
- **When** its `full-matrix-test` job is read with `yq`,
- **Then** `strategy.matrix.os` has the three runner labels, `strategy.matrix.shard` has exactly six ids (`cli-1`..`cli-4`, `rosterguard`, `remainder`), `strategy.fail-fast` is `false`, and the job id is still `full-matrix-test`.
- RED-now: L-E14a (`yq '.jobs.full-matrix-test.strategy.matrix.shard | length'` -> `0`, exit 0) and L-E14b (matrix keys: `os` only), tree 74ad197380 — red because no shard dimension exists.
- Green path: M1 — the matrix gains the shard dimension and per-shard `include` entries; the length query returns `6`.
- Mutant probe: a five-shard matrix (dropping `remainder`) satisfies "a shard dimension exists" but not the `== 6` and id-set assertion; AC-RVS-002/003 independently catch a dropped shard's lost tests.

### AC-RVS-002 — Selector identity with ci.yml is enforced by a failing test (RB; maps REQ-RVS-003)

- **Given** `ci.yml` and the release workflow,
- **When** the parity test runs and a shard selector or the remainder exclusion list is altered in either file,
- **Then** the test fails and names the shard (or the exclusion list) that diverged; with both files untouched it passes.
- RED-now: L-E14d (`ls internal/template/release_verify_shard_parity_test.go` -> not found, exit 1), tree 74ad197380 — the guard does not exist.
- Green path: M3 — the test exists and passes on the run-phase tree; the run-phase evidence records the mutant runs (selector `[J-R]` -> `[J-Q]` in a copy of either file fails with the shard named; one exclusion line removed fails).
- Mutant probe: a test that only checks "the file parses" passes the all-different mutant; the criterion requires the divergence-names-the-shard failure to be observed (verification-completeness §1.1).

### AC-RVS-003 — The four internal/cli selectors are an exact cover of the real test list (RB; maps REQ-RVS-002)

- **Given** the release workflow's four internal/cli selectors,
- **When** the exact-cover listing is run (`go test -list` per selector against the full `internal/cli` list),
- **Then** the full list and the union of the four selector listings have the same line count and no diff, with no name in two selectors.
- RED-now: L-E14f (`shard-cover.sh ... release-pr-multi-os.yml` -> `FAIL: expected 4 internal/cli selectors ... found 0`, exit 9), tree 74ad197380 — red because the release workflow carries no selectors. Baseline reference on `ci.yml`: L-E9 (5981 = 2111+862+1851+1157, exact cover, exit 0).
- Green path: M1 + M3 — the same script run against the release workflow (matrix `include` form) exits 0 with `EXACT COVER`; the in-repo parity test asserts the same property over parsed test-function names.
- Mutant probe: L-E10 (shard 3 narrowed by one letter) exits 1 with 809 names dropped — the checker is not vacuous. The one-sided "no gap" mutant and the "no overlap" mutant (shard 2 widened to `[G-J]`) are both required to fail in the run-phase evidence.

### AC-RVS-004 — The gate job is unchanged in text and semantics (RG; maps REQ-RVS-005)

- **Given** base `74ad197380` and the run-phase tree,
- **When** `.jobs.release-pr-gate` is extracted with `yq` from both and compared, and the verifier suite runs,
- **Then** the two extractions are identical (including `needs: [full-matrix-test, detect-release]`), and `python3 scripts/ci-mirror/test/verifier-fail-closed_test.py` passes all 5 tests including the 12-case release-gate matrix.
- Baseline (green now): L-E8 (5 tests OK, exit 0), L-E1 (job keys) and `needs` read `["full-matrix-test","detect-release"]`, tree 74ad197380.
- Mutant probe: a gate that accepts `cancelled` is rejected by the existing case `("success","cancelled","false", False)` of that suite; a gate whose `needs` drops the matrix job fails the extraction identity check.

### AC-RVS-005 — No required-check name changes (RG; maps REQ-RVS-006)

- **Given** the live protection of `main` and the run-phase tree,
- **When** the required contexts are read (read-only) and the required-checks validator runs,
- **Then** the five contexts are unchanged (`Test (ubuntu-latest)`, `Lint`, `Build (linux/amd64)`, `Analyze (Go) (go)`, `Release PR Multi-OS Gate`), none contains `Release Verify`, and `sh scripts/ci-mirror/validate-required-checks.sh` exits 0 with the gate listed as publishable.
- Baseline (green now): L-E6 (contexts list, read-only GET) and L-E13 (validator `All validations passed`), tree 74ad197380.
- Mutant probe: renaming the gate job's `name:` makes the validator report a phantom context and exit non-zero (the validator's dimension D, SPEC-CI-VERDICT-INTEGRITY-001 AC-CI-011).

### AC-RVS-006 — No timeout increase, no exclusion, no relaxation (RG; maps REQ-RVS-007)

- **Given** base and run-phase trees,
- **When** the release workflow is scanned,
- **Then** the job `timeout-minutes` is 40, every `go test` carries `-timeout 35m` and no larger value, there is no `-skip`, no `continue-on-error`, and no reference to `retry-flaky.sh`.
- Baseline (green now): L-E2 (`timeout-minutes: 40`, `-timeout 35m`), tree 74ad197380.
- Mutant probe: a diff raising either number (or adding `-skip`) is a failing change; the assertion compares numerically against the base values, not by string presence.

### AC-RVS-007 — Artifact names are unique for every shard x OS (RB; maps REQ-RVS-010)

- **Given** the 18 legs of one run,
- **When** the `Upload test event stream` step's `with.name` is evaluated per leg,
- **Then** the 18 names are pairwise distinct and each contains the OS and the shard id.
- RED-now: L-E14c (`test-stream-release-verify-${{ matrix.os }}` — no shard component, exit 0), tree 74ad197380; with six shards per OS the names would collide.
- Green path: M1 — the name template carries both `matrix.os` and `matrix.shard`; evaluated over the matrix it yields 18 distinct strings.
- Mutant probe: a name carrying only `matrix.shard` collides across OS; one carrying only `matrix.os` collides across shards — both fail the pairwise-distinct check.

### AC-RVS-008 — Existing steps are preserved (RG; maps REQ-RVS-010, REQ-RVS-012)

- **Given** base and run-phase trees,
- **When** the matrix job's step names are listed,
- **Then** every base step name in L-E1b is present (the `Diagnose ancestor .moai` pre/post steps still gated on `runner.os == 'Windows'` with `shell: pwsh`, the `sg` install and identity check, `go vet`, census, compress, upload) and the windows `if-no-files-found: error` expression is intact.
- Baseline (green now): L-E1b, tree 74ad197380.
- Mutant probe: removing the post-test diagnostic step, or its `always()`, is detected by the name-and-condition comparison.

### AC-RVS-009 — Static tooling stays clean (RG; maps REQ-RVS-005..007)

- **Given** the changed workflow and the template test package,
- **When** `actionlint .github/workflows/release-pr-multi-os.yml`, `sh scripts/ci-mirror/validate-required-checks.sh`, and `go test -count=1 -run '^(TestReleaseScriptMatrixGate|TestTestInputFiltersSingleSource|TestTestInputFiltersCoversGoTestInputs|TestReleaseVerifyShardParity)$' ./internal/template/` run,
- **Then** each exits 0; actionlint prints nothing.
- Baseline (green now): L-E13 (actionlint v1.7.10 installed; verified clean on base), tree 74ad197380. If `actionlint` is absent on the run machine, install it with `go install github.com/rhysd/actionlint/cmd/actionlint@v1.7.10` before use and record the version.
- Mutant probe: an undefined `matrix.*` reference (typo) is an actionlint error; the criterion requires that mutant to be shown failing once.

### AC-RVS-010 — A shard that swept nothing fails; test results are not cached (RB; maps REQ-RVS-004, REQ-RVS-009)

- **Given** the shard step body extracted from the workflow,
- **When** it is run with a selector that matches no test, and separately with a selector that matches tests,
- **Then** the first exits non-zero with the empty-sweep message and the second exits 0, and every `go test` invocation in the body carries `-count=1`.
- RED-now: L-E14e (`grep -c 'empty sweep'` -> `0`, exit 1; `grep -c -e '-count=1'` -> `0`, exit 1), tree 74ad197380.
- Green path: M1 — the body carries the empty-sweep refusal and `-count=1`; M4 records both local runs (a tiny package, `-race`, one matching and one non-matching selector).
- Mutant probe: a body whose refusal is spelled but never reached (e.g. after `exit $rc`) passes the grep and fails the non-matching-selector run — the run, not the grep, is the evidence.

### AC-RVS-011 — release.sh accepts sharded legs and rejects partial ones (RB; maps REQ-RVS-011)

- **Given** a dispatched run whose jobs are `Release Verify (<os>, <shard>)` for 3 OS x 6 shards,
- **When** `scripts/release.sh --require-matrix-run` evaluates it against the extended gh-double scenario table,
- **Then** an all-success run passes; and each of these withholds the gate with a reason naming the OS and shard: one shard `failure`, one shard `skipped`, one shard absent on a single OS, an OS with zero shard jobs.
- RED-now: L-E7 (`leg-match-probe.sh` -> `shard-suffixed names -> <no leg>`, exit 0 = RED state), tree 74ad197380 — the current exact-name lookup finds no leg among sharded names, so even an all-green sharded run fails closed.
- Green path: M2 — the lookup and the Go scenario table change together; `go test -count=1 -run '^TestReleaseScriptMatrixGate$' ./internal/template/` passes with the added scenarios.
- Mutant probe: a lookup that accepts "any one shard success" per OS passes the all-green scenario but fails the one-shard-failed scenario; a lookup that does not compare shard sets across OS fails the shard-absent-on-one-OS scenario.

### AC-RVS-012 — The shard step runs under macOS bash 3.2 and Git Bash conventions (RB local, PL windows; maps REQ-RVS-012)

- **Given** the shard step body extracted from the workflow and run on this macOS machine under `/bin/bash` (3.2.57),
- **When** run with `SHARD_PKG=<a small package>` and a matching selector,
- **Then** it exits 0, prints the census, and writes a non-empty `test-stream.json`; the selector arrives only through the `SHARD_RUN` environment variable (no `${{ matrix.run }}` interpolated into the script text).
- RED-now: none claimed as a flip on a different OS — the body does not exist yet; L-E14e shows the empty-sweep refusal is absent. The windows half is a PL observation (AC-RVS-014): Git Bash cannot be exercised from this machine.
- Green path: M1 + M4 — local run under `/bin/bash` recorded verbatim; `grep -n '\${{ matrix\.' ` over the `run:` block returns nothing.
- Mutant probe: a body that interpolates the selector into shell text passes the matching run but fails the grep; a body using a bash-4-only construct fails the `/bin/bash` 3.2 run.

### AC-RVS-013 — No duplicate runs: vet once per OS, six shards, each test once (RG; maps REQ-RVS-008)

- **Given** the 18-leg matrix,
- **When** the legs that run `go vet ./...` are counted, and the package partition (internal/cli, rosterguard, remainder) is counted against `go list ./...`,
- **Then** exactly 3 legs run vet (one per OS), the shard count is 6, and the three package sets sum to `go list ./...` with no package in two of them (176 packages on tree 74ad197380).
- Baseline (green now): the base runs vet once per OS (L-E1b, one `Run go vet` step per leg of a 3-leg matrix) and the partition holds on `ci.yml` (L-E9; `go list ./...` count 176 measured at this tree).
- Mutant probe: a `go vet` step with no shard condition runs 18 times and fails the count; a remainder exclusion list missing the rosterguard package runs it twice and fails the partition identity.

### AC-RVS-014 — Post-landing: PR #1830 re-run shows every shard green and the gate green (PL; maps REQ-RVS-013)

- **Given** the change is on the head of PR #1830 (or its successor release PR) and the run completes,
- **When** the leader reads the run with `gh run view <id> --json jobs`,
- **Then** all 18 `Release Verify (<os>, <shard>)` jobs are `success` (none `cancelled`, none within 5 minutes of a 40-minute budget), and `Release PR Multi-OS Gate` is `success`.
- Class: post-landing observation, **not closable by the lane** — undecidable offline, so it loses release-blocking eligibility by the verification-completeness §2.1 disposition and is not recorded as a pass until the leader records it. RED reference: L-E4 (run 39180942620: all legs `cancelled`, gate `failure`).
- Green path: leader observation after landing; the per-OS shard durations it reveals close gap G2 and feed any rebalance card.

### AC-RVS-015 — The plan phase wrote only inside the SPEC directory (RG; maps constraint D.5)

- **Given** the plan-phase commit,
- **When** `git diff --name-only 74ad197380 <plan-commit>` is read,
- **Then** every path starts with `.moai/specs/SPEC-CI-RELEASE-VERIFY-SHARD-001/`; nothing under `.github/`, `scripts/`, `internal/`, or `.moai/reports/` appears.
- Baseline: recorded in progress.md §E.1 with the observed command output.
- Mutant probe: a stray edit to a workflow file appears as a path outside the prefix and fails the check.

## §C. Edge cases

- **Docs-only release PR**: `detect-release` reports `go_code == 'false'`; the matrix job is skipped as a whole (all 18 legs), the gate names the `docs-only-release-pr` exclusion exactly as today (L-E1; gate unchanged, AC-RVS-004).
- **`workflow_dispatch`**: always runs the matrix (existing condition, untouched); `scripts/release.sh --require-matrix-run` reads dispatched runs only (AC-RVS-011).
- **One shard cancelled by a runner loss**: `needs.full-matrix-test.result` aggregates to `cancelled`/`failure`; the gate fails; `fail-fast: false` keeps the other shards reporting.
- **macOS concurrency queue**: a queued shard starts later; no correctness effect (research.md §C.5).
- **Future test names**: a new top-level `internal/cli` test is covered by the first-letter selectors without editing them; a name with a first letter outside every selector cannot exist for valid Go test names; the parity test is the check that fires if that ever changes.

## §D. Quality gate criteria and Definition of Done

TRUST 5 (workflow/shell/test changes): **Tested** — the new parity test and the extended release-script scenarios pass; local step-body runs recorded. **Readable** — comments in the workflow state why each guard exists (empty-sweep, `-count=1`, env-var selector). **Unified** — `gofmt` clean on new Go, `actionlint` clean on the workflow. **Secured** — least-privilege `permissions: contents: read` unchanged; no `${{ }}` interpolation into script text; no new third-party action. **Trackable** — commits cite `card t1892` and the SPEC id.

Definition of Done (run phase): AC-RVS-001..013 and 015 recorded green with command and verbatim output in progress.md §E.2; AC-RVS-014 left as the leader's post-landing observation with a pointer; no file outside the declared scope touched; the keep-set boundary respected (no branch-protection write).
