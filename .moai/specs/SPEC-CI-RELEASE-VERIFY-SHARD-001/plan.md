# plan.md — SPEC-CI-RELEASE-VERIFY-SHARD-001

Tier: M · Card: t1892 (Class C, P0) · Plan-phase tree: `74ad197380` (branch `WT-release-verify-os`, = PR #1830 head, clean at authoring) · Evidence ledger: `research.md` §B (cited by id).

## §A Context

PR #1830's `Release Verify` legs all hit the 40-minute job timeout inside one `go test -json -race -timeout 35m ./...` step (research L-E4). The fix is structural: run the race suite as the same six shards `ci.yml` already runs, on all three OSes, inside the existing `full-matrix-test` job id so the required gate stays byte-identical.

**Decision surface first (highest change-likelihood).** The decisions most likely to change are the job structure and the leg naming (Q1/Q2), then where the selectors live (Q3), then ceilings/retry/vet placement (Q4-Q7). They are recorded with defaults in `decision-index.md`; the mechanical steps (artifact naming, parity test plumbing) come last in the milestones.

Files in scope (run phase), 5 primary:

| # | File | Milestone |
|---|------|-----------|
| 1 | `.github/workflows/release-pr-multi-os.yml` | M1 |
| 2 | `scripts/release.sh` (leg lookup, line ~360) | M2 |
| 3 | `internal/template/release_script_test.go` (extend `matrixScenarios`) | M2 |
| 4 | `internal/template/release_verify_shard_parity_test.go` (new) | M3 |
| 5 | `.moai/specs/SPEC-CI-RELEASE-VERIFY-SHARD-001/progress.md` (§E.2 evidence, run phase) and `repro/` scripts (already committed at plan phase) | M4 |

Read-only inputs: `.github/workflows/ci.yml` (selector source), `scripts/ci-census/test-census.sh`, `scripts/ci-mirror/validate-required-checks.sh`, `scripts/ci-mirror/test/verifier-fail-closed_test.py`. None has a template mirror (`internal/template/templates/` carries no copy of these), so no Template-First regeneration is triggered — verify with `git status` at close.

## §B Known Issues (verified, carried into the run-phase delegation)

- **B-1 Job id and gate are coupled to tests.** The Python verifier substitutes `needs.full-matrix-test.result` into gate step 0 (`verifier-fail-closed_test.py:21`) and `test_input_filter_parity_test.go:83` looks for the literal `needs.detect-release.outputs.go_code != 'false'`. Keep job id `full-matrix-test`, keep the gate job untouched, keep that `if:` condition string verbatim.
- **B-2 Leg-name consumers.** `scripts/release.sh:360` and `release_script_test.go:223-225,248` use the exact names `Release Verify (<os>)` (L-E7). Renaming to `Release Verify (<os>, <shard>)` requires M2 in the same change; shipping M1 without M2 breaks `--require-matrix-run`.
- **B-3 Not a required context.** `Release Verify (...)` is absent from `main` protection (L-E6); renaming it changes no required name. The gate name `Release PR Multi-OS Gate` must not change.
- **B-4 Matrix `include` merge semantics.** An `include` entry that names only the `shard` key adds its other keys to every combination with that shard value (they are keys not present in the original matrix, so nothing is overwritten). `scripts/ci-mirror/validate-required-checks.sh` expands this form (its GATE-6 merge handling); re-run it after M1.
- **B-5 Selector quoting.** Selectors contain `|`, `$`, `^`, `[^…]`. Pass them through `env:` (`SHARD_RUN`) and reference `"$SHARD_RUN"` in the script; never interpolate `${{ matrix.run }}` into `run:` text (REQ-RVS-012). `-run` splits on `/`; none of the selectors contains one.
- **B-6 Empty remainder risk.** The remainder leg's package list comes from `go list ./...` through `grep -v -x`; refuse an empty list and refuse an empty stream, as `ci.yml` does.
- **B-7 Windows artifact rule.** Keep `if-no-files-found: ${{ runner.os == 'Windows' && 'error' || 'warn' }}`.
- **B-9 Shard weight on macOS/Windows is unmeasured** (spec.md §G, decision-index Q8). Do not change shard count or ceilings in the run phase; the first live run measures (AC-RVS-016).
- **B-8 Worktree guard.** AC verification commands use plain single-command forms (no `cd &&`, no compound `;` chains); use `bash <script>` forms for anything multi-step (the repro scripts already do).

## §C Pre-flight (run-phase entry checks)

```bash
git rev-parse --short HEAD
git branch --show-current
actionlint .github/workflows/release-pr-multi-os.yml
sh scripts/ci-mirror/validate-required-checks.sh
python3 scripts/ci-mirror/test/verifier-fail-closed_test.py
```

Record the base values (clean actionlint, validator pass, 5 verifier tests OK = research L-E13, L-E8). Re-measure the exact-cover listing on the run-phase tree before editing (`bash .moai/specs/SPEC-CI-RELEASE-VERIFY-SHARD-001/repro/shard-cover.sh <worktree-root> <scratch> .github/workflows/ci.yml`); do not quote the plan-phase 5981 as the run-phase baseline. Run it once at a time, outside any resource-contention window (`internal/cli` listing compiles the package test binary).

## §D Constraints

1. No timeout above `-timeout 35m` / `timeout-minutes: 40`; no `-skip`; no `continue-on-error`; no retry; no gate edit; no required-context change; no branch-protection write (REQ-RVS-005..007).
2. `.github/workflows/ci.yml` is not edited.
3. Six shards per OS exactly; vet once per OS; `-count=1` on every race invocation.
4. Stage files by explicit pathspec; conventional commits carrying `card t1892` and ending with the `🗿 MoAI` line.
5. Local verification is scoped: the listing script, the targeted `internal/template` tests, and the two small-package step-body runs. No `go test ./...`; no background load.

## §E Self-Verification (run-phase deliverable shape)

Report per AC-RVS-001..013 and 015 in the 5-section evidence form (Claim / Evidence = command + verbatim output / Baseline-attribution = this run, this tree, HEAD SHA / Gaps / Residual-risk). AC-RVS-014 and AC-RVS-016 are listed as leader post-landing observations with the exact `gh run view` command to run. Gaps to carry: windows step execution (G1), per-OS shard timing (G2).

## §F Milestones (priority order; highest change-likelihood first)

### M1 — Release Verify matrix: shard x os in the existing job (Priority High)

Edit `.github/workflows/release-pr-multi-os.yml`, job `full-matrix-test` only:

- `name:` -> `Release Verify (${{ matrix.os }}, ${{ matrix.shard }})`.
- `strategy.matrix`: `os: [ubuntu-latest, macos-latest, windows-latest]`, `shard: [cli-1, cli-2, cli-3, cli-4, rosterguard, remainder]`, plus exactly six `include` entries keyed by `shard` alone (no `os` key, no `exclude` anywhere in the matrix) that add `pkg` and `run` (and `vet: true` on `cli-2`):

  | shard | pkg | run |
  |-------|-----|-----|
  | cli-1 | `./internal/cli/` | `^Test([A-E]\|Fa)` |
  | cli-2 | `./internal/cli/` | `^Test(F([^a]\|$)\|[G-I])` (+ `vet: true`) |
  | cli-3 | `./internal/cli/` | `^Test[J-R]` |
  | cli-4 | `./internal/cli/` | `^Test([S-Z]\|[^A-Za-z])` |
  | rosterguard | `./internal/harness/rosterguard/` | `^Test` |
  | remainder | `remainder` | (empty) |

  (selectors copied verbatim from `ci.yml` at the run-phase base; the table is the Q3 Default.)
- Keep `fail-fast: false`, `timeout-minutes: 40`, `needs`, and `if` unchanged.
- `Run go vet` gets `if: matrix.vet == true` (Q6). Everything else — checkout, setup-go, `go mod download`, `sg` install and identity check, both windows diagnostic steps — unchanged (Q7).
- Replace the test step with one `shell: bash` step: `env: { SHARD_RUN: ${{ matrix.run }}, SHARD_PKG: ${{ matrix.pkg }} }`; body keeps the `|| rc=$?` form and the `test-census.sh` call, runs `go test -json -race -count=1 -timeout 35m` either as `-run "$SHARD_RUN" "$SHARD_PKG"` (cli/rosterguard) or over the `go list ./... | grep -v -x -e …/internal/cli -e …/internal/harness/rosterguard` list (remainder; refuse an empty list, and refuse a list that still contains either excluded package: `printf '%s\n' "$pkgs" | grep -c -x -e …/internal/cli -e …/internal/harness/rosterguard` must be 0 — REQ-RVS-014), writes `test-stream.json`, refuses an empty stream (`grep -q '"Test":"' test-stream.json || …`), then `exit $rc`. No `retry-flaky.sh` (Q5).
- Upload artifact name: `test-stream-release-verify-${{ matrix.os }}-${{ matrix.shard }}`.
- Update the comment block above the job so the explanation of the split, the 38-minute margin, and the unchanged ceilings is stated once, with the Q-row pointers.

Flip evidence (M1-only halves): AC-RVS-001 (shape queries), 004, 005, 006, 007 (name template), 008, 010, 012, 013 (shard count, remainder guard, vet condition); the halves that need M3 are listed under M3. Reversibility: a single revert restores the base file; M2 must revert with it (B-2).

### M2 — release.sh leg lookup and its test (Priority High; ships with M1)

Replace the exact-name loop at `scripts/release.sh` ~L360 by: for each of the three OS values and each of the six declared shard ids (`cli-1`..`cli-4`, `rosterguard`, `remainder`), require exactly one job named `Release Verify (<os>, <shard>)` and require it to be `success`; zero such jobs is "missing", two is "duplicate"; then reject any job whose name starts with `Release Verify (` but is not one of the 18 expected names ("unknown shard"). Each failure sets `run_problem` naming the OS and shard (existing `run_problem` / reason plumbing). The declared shard list is a constant in the script, mirrored in the workflow matrix and asserted equal by the parity test (M3), so the two cannot drift silently.

Extend `matrixScenarios()` in `release_script_test.go`. **Order matters (two-cell pair):** first add the sharded all-green scenario and run it against the unmodified script, recording the refusal (the whole-script RED; `repro/leg-match-probe.sh` is only an extracted-expression probe, D6); then edit `release.sh`. Scenarios: sharded all-green (passes); one shard `failure`; one shard `skipped`; a shard absent on one OS; **the same shard absent on all three OS**; **an unknown shard name**; **a duplicated shard job**; an OS with no shard job. Keep the seven existing scenarios (rewritten to sharded fixtures where they name legs) so each still discriminates. Keep `bash` 3.2-safe constructs (the script already targets macOS).

Flip evidence: AC-RVS-011. Mutants (any-one-shard; cross-OS compare without the declared six; repeats tolerated) must each be shown failing against a deliberately weakened copy.

### M3 — Drift and exact-cover guard (Priority High)

New `internal/template/release_verify_shard_parity_test.go` holding `TestReleaseVerifyShardParity` (uses `t.TempDir()` only where it writes; reads both workflow files with the module's existing YAML dependency):

- Extract the four internal/cli `-run` selectors, the rosterguard selector, and the remainder exclusion list from `ci.yml`'s race steps; extract the same from the release workflow's `include` entries and remainder step; assert equality shard by shard, naming the shard on mismatch.
- Obtain the names from the go tool, not from source parsing: run `go test -list '.*'` for `./internal/cli/` and for the rosterguard package from the test (`exec.Command("go", ...)` with the repository root as the working directory, fail on a non-zero exit, drop the trailing `ok`/`FAIL` line and `Benchmark...` names, assert the list is non-empty). This reuses Go's own discovery (bare `Test`/`Fuzz`/`Example`, case-insensitive and block-comment example output markers, `*testing.T`/`*testing.F` signatures) so none of it is re-implemented; measured behaviour on a scratch package is research.md L-E20. Assert every `internal/cli` name matches exactly one of the four selectors under `regexp` (the engine `go test -run` uses); a name matching zero or two fails naming the test and the selectors. Every rosterguard name must match `^Test`. Discovery guard (REQ-RVS-016): any `Fuzz...` or `Example...` name in either list fails, because no `^Test` selector selects it. The nested `go test -list` compiles both test binaries; it is a run-phase cost accepted for exactness and must run under the resource-entry rules (one Go call, load below the lane limit). Limit: names that exist only under another OS's build tags are not listed on the running OS.
- Assert the release matrix has no `exclude` key, exactly six `include` entries each keyed by `shard` alone, and exactly the six shard ids, that the shard-id list equals the constant list in `scripts/release.sh`, and compute the **final leg list with `include` and `exclude` applied** (a small expander in the test; not the plain cross product): exactly 18 legs, 18 distinct `(os, shard)` pairs, 18 pairwise-distinct artifact names from the name template (AC-RVS-007), and the legs whose merged entry has `vet: true` number exactly the OS count, 3 (AC-RVS-013).
- Mutation table inside the test (edits applied to in-memory copies) proving each failure mode fires — the observed-failure completion act (verification-completeness §1.1): (i) selector `[J-R]`->`[J-Q]`; (ii) a remainder exclusion removed; (iii) injected name `Test` (zero matches); (iv) shard 2 widened to `[F-I]` (a `TestFa...` name matches two); (v) an injected `FuzzX`, `ExampleX`, bare `Fuzz` or bare `Example` name in the listed names; (vi) artifact-name template without `matrix.shard`; (vii) a second `vet: true` include; (viii) `exclude: [{os: ubuntu-latest, shard: remainder}]` (17 legs); (ix) a duplicated `include` entry — the leg count stays 18 (GitHub merges the copy into the same combinations), so it is caught by the exactly-six-include-entries assertion (measured on a scratch copy: `include | length` 7 vs 6, research.md L-E19). All of (i)-(ix) are must-FAIL. Must-PASS control **C1**: an injected `BenchmarkX` name does not fail. A sanity assertion that the swept set is non-empty (parsed name count > 0, 4 selectors found) guards against a vacuous pass.
- The `ci.yml` go-code filter already lists both workflow files (`requiredGoCodeRoots` in `test_input_filter_parity_test.go`), so the test runs whenever either file changes — the continued-firing answer (§1.3).

Flip evidence: AC-RVS-002, AC-RVS-003 (test half), and the M3-dependent halves of AC-RVS-001 (expander leg set), AC-RVS-007 (18 names), AC-RVS-009 (fourth swept name), AC-RVS-013 (vet count, partition identity). Reversibility: new file; revert deletes it.

### M4 — Evidence and boundary (Priority Medium)

Run and record in progress.md §E.2, verbatim: the AC-RVS-009 swept count (`go test -list` on the anchored regex prints four names); the stub-`go` print-then-fail run of `shard-cover.sh` (must exit 1); the remainder-step guard run with an unfiltered listing (must fail); the pre-flight set; `shard-cover.sh` against both workflows (and the two mutants: narrowed and widened selector); `actionlint`; the validator; the verifier suite; the targeted `internal/template` tests; the step body under `/bin/bash` 3.2 for one tiny package with a matching selector (exit 0, census printed) and a non-matching one (exit non-zero, empty-sweep message); the `yq` assertions for AC-RVS-001/004/006/007/008/013; the plan-boundary diff for AC-RVS-015. Write the leader handoff item for AC-RVS-014 (post-landing).

## §G Anti-Patterns

- **Raising a timeout to make the legs pass.** The margin was already 1m50s on the only green run (research L-E5); a bigger number only moves the cliff and is forbidden.
- **Renaming the matrix job without M2.** Breaks `--require-matrix-run` closed and silently (the flag reports "no leg").
- **Editing the gate "to cover the new structure".** The gate already aggregates every matrix leg through `needs.full-matrix-test.result`; any edit is a relaxation risk and breaks the verifier's extraction.
- **Interpolating `${{ matrix.run }}` into the script.** Regex metacharacters get re-parsed by the shell; use the env var.
- **Quoting 5981 as a standing fact.** It is a measurement of one tree; the parity test and the re-run listing are the check.
- **A guard that only greps.** Spelling `empty sweep` or `-count=1` proves nothing; the non-matching-selector run does.
- **Sweep-staging.** Never `git add -A` in the primary checkout; explicit pathspecs only.

## §H Cross-References

- `.github/workflows/ci.yml` lines 313-654 — the shard configuration and its measurement comment (read-only source).
- SPEC-CI-VERDICT-INTEGRITY-001 — the gate's aggregation repair (AC-CI-001/002) this SPEC must not weaken.
- SPEC-CANDIDATE-CI-001 (M6, card t1478) — origin of the six race legs.
- `.claude/rules/moai/development/verification-completeness.md` §1.1, §1.3, §2, §2.1 — observed-failure completion, continued firing, two-cell adoption, four-element RED cells.
- `.claude/rules/local/gitflow-lane-protocol.md` — landing is the leader's; lanes do not push.
