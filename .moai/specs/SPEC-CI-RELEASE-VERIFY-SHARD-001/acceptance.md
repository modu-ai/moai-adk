# acceptance.md — SPEC-CI-RELEASE-VERIFY-SHARD-001

Tier M verification layer. Document-level tree pin for every RED cell: **`74ad197380`** (binds every criterion carrying no pin of its own). The four-element RED cells (command, verbatim stdout, exit code, tree) live in the evidence ledger in `research.md` §B and are cited here by id — the ledger is the carrier (verification-completeness §2.1). Criteria are the verification layer: Given-When-Then, binary-testable. The GEARS requirements they verify are the REQ-RVS entries in `spec.md`.

Classification: **RB** release-blocking (RED re-executable on this tree and flipped by this SPEC's run phase); **RG** regression-guard (an invariant that is green now and must stay green; a mutant that breaks it is named); **PL** post-landing observation (made by the leader after landing; not a lane gate).

## §B. Acceptance criteria

### AC-RVS-001 — Six shards per OS run as one matrix job (RB; maps REQ-RVS-001, REQ-RVS-008)

- **Given** the release workflow at the run-phase tree,
- **When** its `full-matrix-test` job is read with `yq`,
- **Then** `strategy.matrix.os` has the three runner labels, `strategy.matrix.shard` has exactly six ids (`cli-1`..`cli-4`, `rosterguard`, `remainder`), `strategy.fail-fast` is `false`, and the job id is still `full-matrix-test`. **The matrix has no `exclude` key and exactly six `include` entries, each keyed by `shard` alone (no `os` key), one per declared shard id.** The leg set computed with `include` and `exclude` applied (the parity test's matrix expander, M3) is exactly 18 legs: 18 distinct `(os, shard)` pairs, each OS x shard pair once.
- RED-now: L-E14a (`yq '.jobs.full-matrix-test.strategy.matrix.shard | length'` -> `0`, exit 0), L-E14b (matrix keys: `os` only, exit 0) and L-E17 (`yq '.jobs.full-matrix-test.strategy.matrix.include | length'` -> `0`, exit 0; `has("exclude")` -> `false`, exit 0), tree 74ad197380 — red because no shard dimension and no shard-keyed `include` entries exist. (`has("exclude") == false` is already true at base and stays true: that half is a guard, not a flip.)
- Green path: M1 defines the matrix (shard dimension, six shard-keyed `include` entries); M3's expander asserts the leg set. The length queries return `6` and `6`.
- Mutant probe: a five-shard matrix (dropping `remainder`) fails the `== 6` and id-set assertion. **An `exclude: [{os: ubuntu-latest, shard: remainder}]` entry** leaves three OS values, six shard ids and a clean cross product, but fails the no-`exclude` assertion and the expander's 18-leg count (17 legs). **A duplicated `include` entry** (two `shard: cli-1` entries) does NOT change the leg count: a shard-keyed entry overwrites no original matrix value, so GitHub's include rule (and the repository expander, `validate-required-checks.sh` GATE-6/GATE-7 merge handling) merges the second copy into the same three combinations and the set stays 18 legs. It is caught structurally instead: the assertion "exactly six `include` entries, one per shard id, each once" fails. Measured on a scratch copy with the duplicate: `include | length` -> `7` (expected 6) and `[.include[].shard] | unique | length` -> `6` (L-E19, exit 0 each; the 7 vs 6 mismatch is the failure signal). Both mutants must be seen failing in the run-phase evidence; AC-RVS-002/003 also catch a dropped shard's lost tests.

### AC-RVS-002 — Selector identity with ci.yml is enforced by a failing test (RB; maps REQ-RVS-003, REQ-RVS-015, REQ-RVS-016)

- **Given** `ci.yml` and the release workflow,
- **When** the parity test runs and a shard selector or the remainder exclusion list is altered in either file,
- **Then** the test fails and names the shard (or the exclusion list) that diverged; with both files untouched it passes.
- **And Given** an in-memory name set for `internal/cli` containing one name that matches no selector (the bare name `Test`, which no ci.yml selector matches) and one that matches two selectors (a `TestFa...` name once shard 2 is widened to `[F-I]`, so it matches shards 1 and 2),
- **When** the parity test's cover check runs over that set,
- **Then** it fails once per offending name, naming the test and the selectors involved (zero selectors for the first, both selector ids for the second); over the real parsed test names it passes.
- **And Given** (REQ-RVS-016) the name list the go tool prints for a package (`go test -list '.*'`) containing a fuzz-target name or an example name (e.g. injected `FuzzX`, `ExampleX`, a bare `Fuzz`, a bare `Example`), **Then** the test fails naming it, because no `^Test` selector selects it; **and Given** a list containing a `Benchmark...` name, **Then** the test passes (benchmarks do not run without `-bench`). The SPEC does not restate Go's discovery rules: what counts as a runnable fuzz target or example is whatever the tool lists. Measured on a scratch package (L-E20): the tool lists a bare `Test`, a bare `Fuzz`, a bare `Example`, and examples with a lower-case `// output:` marker, a block-comment `/* Output: */` marker or `// Unordered output:`; it does not list a lowercase-suffixed `Fuzzhelper` or an example with no output comment.
- RED-now: L-E14d (`ls internal/template/release_verify_shard_parity_test.go` -> not found, exit 1), tree 74ad197380 — the guard does not exist. The zero/two-match failure modes are RED by the same absence; their discriminating power is shown by the checker-level evidence L-E15 (a script-level mutant: the same two failure shapes make `shard-cover.sh` exit 1).
- Green path: M3 — the test exists and passes on the run-phase tree; its in-test mutation table is executed and recorded: (i) selector `[J-R]` -> `[J-Q]` fails naming shard 3; (ii) one remainder exclusion removed fails; (iii) injected name `Test` fails with zero matches; (iv) shard 2 widened to `[F-I]` fails with `TestFa...` matching two selectors; (v) an injected `FuzzX`, `ExampleX`, bare `Fuzz` or bare `Example` name fails; and one must-PASS control, **C1**: an injected `BenchmarkX` name does not fail. The cover check and the discovery guard both run over the names `go test -list '.*'` prints for `internal/cli` and the rosterguard package, with a non-empty-list assertion.
- Mutant probe: a test that only checks "the file parses" passes all of the failing mutants (i)-(v); a test that checks selector string equality but not cover passes (iii) and (iv); a cover check that tests only "at least one match" passes (iv); a source-level name-prefix or comment-text rule fails against the tool's own list (it misses the bare names and lower-case/block-comment output markers, or flags helpers the tool does not list). Each of (i)-(v) must be seen failing and **C1** seen passing (verification-completeness §1.1).

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
- Evaluator (D7): the parity test computes the **final leg list with `include` and `exclude` applied** (not the plain `os` x `shard` product), asserts exactly 18 legs with 18 distinct `(os, shard)` pairs, substitutes each leg's values into the name template, and asserts 18 distinct strings each containing its OS and shard; it fails on any `exclude` key and on any duplicate `include` entry. `yq` alone reads the template, not the expansion, and the required-checks validator expands job display names, not step `with.name`, so neither is the evaluator.
- RED-now: L-E14c (`test-stream-release-verify-${{ matrix.os }}` — no shard component, exit 0), tree 74ad197380; with six shards per OS the names would collide.
- Green path: M1 + M3 — the name template carries both `matrix.os` and `matrix.shard`; the parity test's expansion yields 18 distinct strings.
- Mutant probe: a name carrying only `matrix.shard` collides across OS; one carrying only `matrix.os` collides across shards — both fail the pairwise-distinct check. A matrix with `exclude: [{os: ubuntu-latest, shard: remainder}]` leaves 17 legs and 17 names (the plain-product check would still say 18) and fails the leg-count assertion; a duplicated `include` entry leaves the leg count at 18 (see AC-RVS-001) and is caught by the exactly-six-include-entries assertion, not by the leg count.

### AC-RVS-008 — Existing steps are preserved (RG; maps REQ-RVS-010, REQ-RVS-012)

- **Given** base and run-phase trees,
- **When** the matrix job's step names are listed,
- **Then** every base step name in L-E1b is present (the `Diagnose ancestor .moai` pre/post steps still gated on `runner.os == 'Windows'` with `shell: pwsh`, the `sg` install and identity check, `go vet`, census, compress, upload) and the windows `if-no-files-found: error` expression is intact.
- Baseline (green now): L-E1b, tree 74ad197380.
- Mutant probe: removing the post-test diagnostic step, or its `always()`, is detected by the name-and-condition comparison.

### AC-RVS-009 — Static tooling stays clean (RG; maps REQ-RVS-005..007)

- **Given** the changed workflow and the template test package,
- **When** `actionlint .github/workflows/release-pr-multi-os.yml`, `sh scripts/ci-mirror/validate-required-checks.sh`, and `go test -count=1 -run '^(TestReleaseScriptMatrixGate|TestTestInputFiltersSingleSource|TestTestInputFiltersCoversGoTestInputs|TestReleaseVerifyShardParity)$' ./internal/template/` run,
- **Then** each exits 0; actionlint prints nothing; and the swept set is counted: `go test -list '^(TestReleaseScriptMatrixGate|TestTestInputFiltersSingleSource|TestTestInputFiltersCoversGoTestInputs|TestReleaseVerifyShardParity)$' ./internal/template/` prints exactly four test names (plus the `ok` line), and the `-v` run shows a `--- PASS:` line for each of the four. Three names at base (L-E13) plus the new one is four; fewer than four is a failure of this criterion, whatever the exit code.
- Baseline (green now): L-E13 (actionlint v1.7.10 installed; verified clean on base; three of the four names exist at base), tree 74ad197380. If `actionlint` is absent on the run machine, install it with `go install github.com/rhysd/actionlint/cmd/actionlint@v1.7.10` before use and record the version.
- Mutant probe: an undefined `matrix.*` reference (typo) is an actionlint error; a misnamed or absent `TestReleaseVerifyShardParity` leaves the run green at three names but fails the count of four — both mutants must be shown failing once.

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
- **Then** an all-success run (each OS carrying exactly the six declared shard ids once each) passes; and each of these withholds the gate with a reason naming the OS and shard: one shard `failure`, one shard `skipped`, one shard absent on a single OS, **the same shard absent on all three OS**, **an unknown shard name** (e.g. a seventh `cli-5`), **a duplicated shard job** (the same OS and shard twice), an OS with zero shard jobs.
- RED-now: L-E7 (`leg-match-probe.sh` -> `shard-suffixed names -> <no leg>`, exit 0 = RED state), tree 74ad197380. Fidelity (D6): the probe runs the awk program text extracted from the real `scripts/release.sh`, not the whole script; the whole-script RED is the sharded all-green scenario in `matrixScenarios()`, which the current script refuses and the repaired one allows (M2 adds it first and records the refusal before editing `release.sh`, the two-cell pair).
- Green path: M2 — the lookup and the Go scenario table change together; `go test -count=1 -run '^TestReleaseScriptMatrixGate$' ./internal/template/` passes with the added scenarios, and `go test -list '^TestReleaseScriptMatrixGate$'` is non-empty.
- Mutant probe: a lookup that accepts "any one shard success" per OS passes the all-green scenario but fails the one-shard-failed scenario; a lookup that compares shard sets across OS but not against the declared six passes every single-OS scenario and fails "same shard absent on all three OS" and "unknown shard"; a lookup that tolerates repeats fails the duplicate scenario. Each mutant must be shown failing against a deliberately weakened copy.

### AC-RVS-012 — The shard step runs under macOS bash 3.2 and Git Bash conventions (RB local, PL windows; maps REQ-RVS-012)

- **Given** the shard step body extracted from the workflow and run on this macOS machine under `/bin/bash` (3.2.57),
- **When** run with `SHARD_PKG=<a small package>` and a matching selector,
- **Then** it exits 0, prints the census, and writes a non-empty `test-stream.json`; the selector arrives only through the `SHARD_RUN` environment variable (no `${{ matrix.run }}` interpolated into the script text).
- RED-now (the local-bash half): L-E16 — `grep -c SHARD_RUN .github/workflows/release-pr-multi-os.yml` -> `0`, exit 1, tree 74ad197380; red because no step passes the selector through the `SHARD_RUN` environment variable, so there is no body to run under bash 3.2 at all.
- The windows half is a PL observation (AC-RVS-014): Git Bash cannot be exercised from this machine, and it is not claimed as RED here.
- Green path: M1 + M4 — `grep -c SHARD_RUN` on the workflow is non-zero; the local run under `/bin/bash` is recorded verbatim; `grep -n '\${{ matrix\.' ` over the `run:` block returns nothing.
- Mutant probe: a body that interpolates the selector into shell text passes the matching run but fails the grep; a body using a bash-4-only construct fails the `/bin/bash` 3.2 run.

### AC-RVS-013 — No duplicate runs: vet once per OS, six shards, each test once (RB for the shard-count, partition and remainder-guard parts; RG for the vet-once part; maps REQ-RVS-008, REQ-RVS-014)

- **Given** the 18-leg matrix,
- **When** the legs that run `go vet ./...` are counted, and the package partition (internal/cli, rosterguard, remainder) is counted against `go list ./...`,
- **Then** exactly 3 legs run vet (one per OS), the shard count is 6, and the three package sets sum to `go list ./...` with no package in two of them (176 packages on tree 74ad197380); the remainder step additionally fails when its list still contains `internal/cli` or the rosterguard package (REQ-RVS-014), shown by one local run with a deliberately unfiltered listing.
- Evaluator (D7): the parity test computes the final leg list with `include` and `exclude` applied (as in AC-RVS-007) and counts legs whose merged entry has `vet: true` — it must equal the number of OS values (3), with exactly one `vet: true` include entry; the command-only form is `yq '[.jobs.full-matrix-test.strategy.matrix.include[] | select(.vet == true)] | length' .github/workflows/release-pr-multi-os.yml` -> `1` (times 3 OS = 3 legs). The partition identity is computed from `go list ./...` against the two exclusions and the remainder count (174).
- Baseline (vet-once part, RG, green now): the base runs vet once per OS (L-E1b, one `Run go vet` step per leg of a 3-leg matrix).
- RED-now (shard-count, partition and remainder-guard parts): L-E14a (shard length `0`, exit 0) and L-E18 (`grep -c 'internal/harness/rosterguard' .github/workflows/release-pr-multi-os.yml` -> `0`, exit 1), tree 74ad197380 — red because the release workflow has no shard dimension, no package partition, and no remainder step that names the excluded packages. Reference measurement of the partition on `ci.yml` (not the file under change): L-E9, `go list ./...` count 176.
- Green path: M1 defines the six shards, the partition and the remainder guard; M3's expander and partition identity assert them; M4 records the unfiltered-listing run showing the guard fires.
- Mutant probe: a `go vet` step with no shard condition runs 18 times and fails the count; a remainder exclusion list missing the rosterguard package runs it twice and fails the partition identity; a remainder step without the REQ-RVS-014 guard passes the unfiltered-listing run and is rejected by it.

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

### AC-RVS-016 — Post-landing: per-shard durations on macOS and Windows are recorded and an overrun has a named next step (PL; maps REQ-RVS-013, spec.md §G)

- **Given** the first full 18-leg run of the sharded workflow (AC-RVS-014's run),
- **When** the leader reads each shard job's `startedAt`/`completedAt` with `gh run view <id> --json jobs`,
- **Then** the 18 durations are recorded on the card; each shard finished well inside its 40-minute job budget (none within 5 minutes of it, none `cancelled`); and if any macOS or Windows shard failed on the 35-minute test ceiling or finished within 5 minutes of the job budget, a rebalance card (split that shard by selector) is opened — no timeout is raised.
- Class: post-landing observation, not closable by the lane (undecidable offline; not recorded as a pass until the leader records it). Reference observation: L-E4/L-E5 — no per-OS shard duration exists, and the one passing leg used 38 min 10 s of 40.
- Green path: leader observation after landing; closes gap G2 and feeds decision-index Q8.
- Flagged for the leader (D9): whether to take one measurement before landing (a `workflow_dispatch` of the release workflow on the candidate ref: one added run, not a duplicate matrix) is the leader's call; the SPEC neither requires nor forbids it.

## §C. Edge cases

- **Docs-only release PR**: `detect-release` reports `go_code == 'false'`; the matrix job is skipped as a whole (all 18 legs), the gate names the `docs-only-release-pr` exclusion exactly as today (L-E1; gate unchanged, AC-RVS-004).
- **`workflow_dispatch`**: always runs the matrix (existing condition, untouched); `scripts/release.sh --require-matrix-run` reads dispatched runs only (AC-RVS-011).
- **One shard cancelled by a runner loss**: `needs.full-matrix-test.result` aggregates to `cancelled`/`failure`; the gate fails; `fail-fast: false` keeps the other shards reporting.
- **macOS concurrency queue**: a queued shard starts later; no correctness effect (research.md §C.5).
- **Future test names**: a new top-level `internal/cli` test is covered by the first-letter selectors without editing them, with one exception — the bare function name `Test` (no suffix) is valid Go, matches none of the four ci.yml selectors (measured by the plan auditor with Go `regexp`: `Test` -> no shard; `TestF` -> 2, `TestFa` -> 1, `Test_x`/`Test1` -> shard 4), and would be silently unrun by both workflows. None exists today (`^Test$` count 0, research L-E9). The parity test reads the bare name and fails on it (AC-RVS-002); any fuzz-target or example name the go tool lists for `internal/cli` or rosterguard is run by `go test` but unselected by `^Test`, and fails the parity test (REQ-RVS-016). None exists today (`grep -lE '^func (Example|Fuzz)'` over both packages' test files printed nothing; the run-phase `go test -list` is the authoritative re-check).
- **Name source (parity test; REQ-RVS-015/016)**: the names are those `go test -list '.*'` prints for the package, minus `Benchmark...` names and the trailing `ok` line; no Go discovery rule is re-implemented in the test, so build tags, receivers, signatures, suffix rules and example-output matching are the go tool's, on the platform the test runs on. A name that exists only under another OS's build tags is not seen by that run; the three-OS release matrix and the per-name cover property make this a limit to record, not a gap in the cover of the running OS.

## §D. Quality gate criteria and Definition of Done

TRUST 5 (workflow/shell/test changes): **Tested** — the new parity test and the extended release-script scenarios pass; local step-body runs recorded. **Readable** — comments in the workflow state why each guard exists (empty-sweep, `-count=1`, env-var selector). **Unified** — `gofmt` clean on new Go, `actionlint` clean on the workflow. **Secured** — least-privilege `permissions: contents: read` unchanged; no `${{ }}` interpolation into script text; no new third-party action. **Trackable** — commits cite `card t1892` and the SPEC id.

Definition of Done (run phase): AC-RVS-001..013 and 015 recorded green with command and verbatim output in progress.md §E.2; AC-RVS-014 and AC-RVS-016 (and the windows half of AC-RVS-012) left as the leader's post-landing observations with pointers; no file outside the declared scope touched; the keep-set boundary respected (no branch-protection write).
