# progress.md — SPEC-CI-RELEASE-VERIFY-SHARD-001

Card: t1892 · Lane: lane-40 (run tn1p4e) · Branch: WT-release-verify-os · Plan-phase tree: 74ad197380

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready (iter1 FAIL 0.78 on f6f502a84c fixed in the iter1-fix commit; pending re-audit)
- iter1 fix record: D1-D11 addressed per `.moai/reports/t1892/plan-audit-iter1.md`; spec 0.1.1 (14 REQ), acceptance 16 AC, decision-index 8 rows; stale hash b42f05fe…c6b3 superseded by the value below.
- plan_complete_at: 2026-10-19

plan_status: audit-ready
plan_complete_at: 2026-10-19
audit_ready: true
audited_sha: c31c261141ff5aa007c9895f710972db4a6b5dd6
audit_verdict: PASS
audit_score: 0.93
audit_verdict_path: .moai/reports/t1892/plan-audit-iter7.md
audit_receipt: rcpt-cb4d28b10d0523a1aa212c47
plan_artifact_hash: 744a64e834de8b964f3d2e676039aac8ac12d1f74d2a63623effd235d4c6912a

- tier: M (spec.md + plan.md + acceptance.md; research.md and decision-index.md alongside; repro/ holds two committed probe scripts)
- iter2 fix record: plan-audit iter2 (FAIL 0.84, codex fail) D1-D7 addressed per `.moai/reports/t1892/plan-audit-iter2.md`; spec 0.1.2 (16 REQ: 003 split into 003/015/016).
- artifacts: spec.md (16 REQ), plan.md (4 milestones), acceptance.md (16 AC), research.md (ledger E1-E16), decision-index.md (8 rows, all FOUNDER / implementation-level / DEFAULT-APPLIED), repro/ (shard-cover.sh fixed to fail on a non-zero `go test`; leg-match-probe.sh extracts the real lookup)
- iter3 fix record: plan-audit iter3 (FAIL 0.84, codex fail) D1-D8 addressed per `.moai/reports/t1892/plan-audit-iter3.md`; spec 0.1.3 (still 16 REQ / 16 AC). Cover and discovery checks now run over `go test -list` names; no heavy Go re-run was made after the leader's quiet-window notice (run-phase re-check of the listing and the scratch behaviours L-E20 is an open run-phase item).
- iter4 fix record: plan-audit iter4 (FAIL 0.88, codex fail) D1-D5 addressed per `.moai/reports/t1892/plan-audit-iter4.md`; spec 0.1.4 (still 16 REQ / 16 AC). Name listing taken under `-race` (race-only-name mutant (x) specified, not run); nested-list compile weight recorded as residual risk (spec §G, research G7/G8); `repro/listing-fixture/` preserves the scratch package. No Go execution during the quiet window.
- iter5 fix record: plan-audit iter5 (label-only FAIL 0.90, codex fail) D1-D3 addressed per `.moai/reports/t1892/plan-audit-iter5.md`; spec 0.1.5 (still 16 REQ / 16 AC). Mutant (x) now needs a paired plain/`-race` run on one clean-base fixture (`repro/listing-fixture/race-pair/`, race-only `FuzzRaceOnly`); expected outcomes written down, not observed. No Go execution.
- iter6 fix record: plan-audit iter6 (FAIL 0.91, codex fail) D1 addressed per `.moai/reports/t1892/plan-audit-iter6.md`; spec 0.1.6 (still 16 REQ / 16 AC). `ExampleAlpha` removed from `race-pair/` (vet `tests` analyzer: "refers to unknown identifier"); no `-vet=off` on race-pair commands; run 1 must exit 0. Vet-analyzer scan of the fixtures is from reading `cmd/vendor/golang.org/x/tools/go/analysis/passes/tests/tests.go` (go1.26.8, the toolchain `go.mod` requires; the source was read from the 1.26.0 install present on this machine, not from a 1.26.8 tree), not execution. No Go execution.
- plan_artifact_hash: 744a64e834de8b964f3d2e676039aac8ac12d1f74d2a63623effd235d4c6912a (iter6-fix; supersedes 40e8f1ed…; SHA-256 per the `internal/runtime/audit_cache.go` `ComputeHash` rule over acceptance.md, decision-index.md, plan.md, research.md, spec.md, computed by `python3 -I /private/tmp/landing-record-lane40/computehash.py <SPEC dir>` after the final edit)
- spec lint: `moai spec lint .moai/specs/SPEC-CI-RELEASE-VERIFY-SHARD-001/spec.md` and `--strict` -> at iter3 the actual output of `--strict` is one INFO finding (`OwnershipTransitionUnmeasured`, the first plan commit has no `Authored-By-Agent` trailer) followed by `0 error(s), 0 warning(s)`, exit 0 (the first pass had flagged two unanchored `-run` patterns, since anchored)
- spec id check: `ID="SPEC-CI-RELEASE-VERIFY-SHARD-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` -> PASS; uniqueness: `ls .moai/specs | grep -c "SPEC-CI-RELEASE-VERIFY"` -> 0 before authoring
- AC-RVS-015 baseline (observed before the plan commit, staged by explicit pathspec): `git diff --cached --name-only` listed only paths under `.moai/specs/SPEC-CI-RELEASE-VERIFY-SHARD-001/` — acceptance.md, decision-index.md, plan.md, progress.md, repro/leg-match-probe.sh, repro/shard-cover.sh, research.md, spec.md; nothing under `.github/`, `scripts/`, `internal/`, or `.moai/reports/`.
- open items for the leader (not blockers for run entry, no keep-set operator action required by this SPEC):
  1. PR #1830 runs the workflow file at its merge ref; the change must reach that ref (landing sequence is the leader's) before AC-RVS-014 can be observed.
  2. Repository is public (hosted standard-runner minutes free) versus the stated 10 USD/month budget; the real limit is macOS concurrency (5 on the Team plan, 6 macos shards).
  3. macOS/windows per-shard durations are unmeasured (gap G2); the first live run is the measurement. A shard over the unchanged 35 min ceiling becomes a rebalance card, never a timeout raise.
  4. `scripts/release.sh --require-matrix-run` must change in the same landing as the workflow (decision-index Q2).

## §E.2 Run-phase Evidence

### M1 — workflow shard matrix (spawn 1 of the run phase; M2, M3, M4 NOT started)

Commits (branch WT-release-verify-os, no push): `ea831e5ec4` status draft to in-progress (spec.md `status:` only); `e95e795219` M1 workflow change (`.github/workflows/release-pr-multi-os.yml` only). Cycle: TDD/characterization — the RED-now cells below were observed on the unmodified workflow at HEAD `658c8e6cae` before any edit; the green observations were taken on the edited workflow, content identical to `e95e795219`. Resource rule in force this spawn: no `go test`, `go build`, `go vet`, `go list`; the Go toolchain was not invoked. Every command below was run by this agent; outputs are verbatim unless marked `(abridged)`. The tool harness showed output but not exit codes for successful commands, so an exit code is stated only where the harness reported one (`Exit code 1`).

Pre-flight baseline (HEAD `658c8e6cae`, workflow unmodified):
- `actionlint .github/workflows/release-pr-multi-os.yml` -> no output (clean; actionlint v1.7.10, `actionlint -version` first line)
- `sh scripts/ci-mirror/validate-required-checks.sh` -> ends `✅ All validations passed`; includes `✓ [main] Release PR Multi-OS Gate is publishable`
- `python3 scripts/ci-mirror/test/verifier-fail-closed_test.py` -> `Ran 5 tests in 3.536s` / `OK`

RED-now (HEAD `658c8e6cae`, unmodified workflow) vs green (edited workflow):

| AC | Command | RED-now | After M1 |
|----|---------|---------|----------|
| 001 | `yq '.jobs.full-matrix-test.strategy.matrix.shard \| length' <wf>` | `0` | `yq -o=json -I=0 '.jobs.full-matrix-test.strategy \| {...}'` -> `{"failfast":false,"os":["ubuntu-latest","macos-latest","windows-latest"],"shard":["cli-1","cli-2","cli-3","cli-4","rosterguard","remainder"],"has_exclude":false,"include_len":6,"include_keys":["shard,pkg,run","shard,pkg,run,vet","shard,pkg,run","shard,pkg,run","shard,pkg,run","shard,pkg,run"]}` |
| 001 | `yq '.jobs.full-matrix-test.strategy.matrix.include \| length' <wf>` | `0` | `6` (no `os` key in any include entry; `has_exclude` false) |
| 001 | job id / needs / if / timeout | | `yq -o=json -I=0 '[.jobs \| keys, ...name, ...timeout-minutes, ...needs, ...if]'` -> `[["detect-release","full-matrix-test","release-pr-gate"],"Release Verify (${{ matrix.os }}, ${{ matrix.shard }})",40,"detect-release","needs.detect-release.outputs.proceed == 'true' && (github.event_name == 'workflow_dispatch' \|\| needs.detect-release.outputs.go_code != 'false')"]` |
| 007 | upload step `with.name` | read from the unmodified file at HEAD `658c8e6cae`: `name: test-stream-release-verify-${{ matrix.os }}` (no shard component; read, not a separate command run) | `yq '...steps[] \| select(.name == "Upload test event stream") \| .with'` -> `name: test-stream-release-verify-${{ matrix.os }}-${{ matrix.shard }}` plus `path: test-stream.json.gz`, `retention-days: 7`, `if-no-files-found: ${{ runner.os == 'Windows' && 'error' \|\| 'warn' }}` |
| 010 | `grep -c "empty sweep" <wf>` | `0` | `1` |
| 010 | `grep -c -e "-count=1" <wf>` | `0` | `grep -n -e "-count=1"` -> line 284 (comment) and lines 297, 299 (the two `go test` invocations, both `-json -race -count=1 -timeout 35m`) |
| 012 | `grep -c SHARD_RUN <wf>` | `0` | `3`; `grep -n -F '${{' <extracted run body>` -> no output (no `${{` in the script text) |
| 013 | `grep -c "internal/harness/rosterguard" <wf>` | `0` | present (remainder step + shard entry); `yq '[...include[] \| select(.vet == true)] \| length'` -> `1`; `yq '...steps[] \| select(.name == "Run go vet") \| .if'` -> `matrix.vet == true` |

AC-RVS-004 (gate unchanged): `yq '.jobs.release-pr-gate' <wf>` written to scratch before and after the edit; `diff` of the two -> no output; `shasum -a 256` -> both `068c049bd7a2df013f69c1af8fd3b7b8dc622719652cec04b93a4d711228c1d5`; the post-commit extraction (tree `e95e795219`) has the same digest. `needs: [full-matrix-test, detect-release]` present in the extraction. Verifier suite after the edit and again at the commit: `Ran 5 tests` / `OK`.

AC-RVS-005 (required checks): validator after the edit and again after the last workflow edit -> `✅ All validations passed`, `✓ [main] Release PR Multi-OS Gate is publishable`. The live branch-protection contexts were NOT re-read this spawn (no `gh api` call made): gap G-A below.

AC-RVS-006 (no relaxation): `grep -n -e "-timeout" -e "timeout-minutes" -e "-skip" -e "continue-on-error" -e "retry-flaky" <wf>` -> job `timeout-minutes: 40` (line 121), detect-release `5` and gate `5` unchanged, both `go test` carry `-timeout 35m`, no `-skip` and no `retry-flaky` anywhere; the `continue-on-error` hits are the pre-existing detect-release step (`continue-on-error: true`) and comments. Compared against base by reading, not by an automated numeric diff.

AC-RVS-008 (steps preserved): `yq '.jobs.full-matrix-test.steps[].name' <wf>` after the edit -> the same eleven names as before the edit, in the same order (Checkout code, Setup Go, Download dependencies, Install ast-grep (provides `sg`), Verify `sg` resolves to ast-grep, Run go vet, Diagnose ancestor .moai (pre-test), Run tests with race detector, Compress test event stream, Upload test event stream, Diagnose ancestor .moai (post-test)). The pwsh diagnostic steps and their `if:` conditions are untouched by the diff (`git diff --stat`: 1 workflow file).

AC-RVS-009 (static tooling): `actionlint` after the final workflow edit -> no output. Mutant: a scratch copy with `matrix.shardd` -> `Exit code 1`, `property "shardd" is not defined in object type {os: string; pkg: string; run: string; shard: string; vet: bool} [expression]`. The `go test -list` four-name count and the `internal/template` test run are M2/M3 items (not run).

Selector identity (precursor of AC-RVS-002/003, M3 owns the test): `grep -o -e "-run '[^']*' \./internal[a-z/]*/" .github/workflows/ci.yml` lists `^Test([A-E]|Fa)`, `^Test(F([^a]|$)|[G-I])`, `^Test[J-R]`, `^Test([S-Z]|[^A-Za-z])` for ./internal/cli/ and `^Test` for ./internal/harness/rosterguard/ (plus the unrelated guard-family run); `yq -r '...include[] | .shard + " | " + .pkg + " | " + .run' <wf>` prints the same five selectors and packages plus `remainder | remainder |`. Equal by reading; the exact-cover over real test names (`shard-cover.sh`, needs `go test -list`) was NOT run.

Leg expansion (scratch expander `expand.py`, my own script applying GitHub's include/exclude merge rule — NOT the M3 parity test): `python3 -I expand.py matrix.json 'test-stream-release-verify-${{ matrix.os }}-${{ matrix.shard }}'` -> `legs: 18`, `distinct (os, shard) pairs: 18`, `distinct artifact names: 18`, `every name contains its os and shard: True`, `legs with vet true: 3 ['macos-latest', 'ubuntu-latest', 'windows-latest']`, `legs per shard` 3 each, `every leg has pkg and run keys: True`. Mutants on scratch copies: `exclude: [{os: ubuntu-latest, shard: remainder}]` -> `legs: 17` (remainder 2); duplicated first include -> `legs: 18` still (as the SPEC predicted) and `jq '.include | length'` -> `7`, `jq '[.include[].shard] | unique | length'` -> `6`; artifact template with `matrix.os` only -> `distinct artifact names: 3`, `every name contains its os and shard: False`.

Step body under `/bin/bash` 3.2.57 (`BASH_VERSION=3.2.57(1)-release`, run as `bash --noprofile --norc -e -o pipefail`, the Actions shape) with a Go-free STUB `go` script (stubbin/go; real Go NOT invoked; the stub emulates `go list` and `go test -json` output, so this observes the body's control flow, quoting and guards, not Go behaviour):
- cli-2 selector `^Test(F([^a]|$)|[G-I])`, pkg `./internal/cli/` -> census printed (`=== totals: packages=1 passed=1 ...`), body exit code 0; argv seen by the stub: `go test -json -race -count=1 -timeout 35m -run ^Test(F([^a]|$)|[G-I]) ./internal/cli/` (metacharacters intact through the env var)
- rosterguard `^Test` -> body exit code 0, argv `go test -json -race -count=1 -timeout 35m -run ^Test ./internal/harness/rosterguard/`
- selector `^NoSuchTest$` (stub emits a stream with no test event) -> `Exit code 1`, `FAIL: shard ran no test (empty sweep)`, body exit code 1
- remainder, normal listing -> body exit code 0, argv `go test -json -race -count=1 -timeout 35m github.com/modu-ai/moai-adk/cmd/moai github.com/modu-ai/moai-adk/internal/spec` (the stub's four-package listing minus the two excluded)
- remainder, empty listing -> `Exit code 1`, `FAIL: remainder package listing failed or came back empty (a bad listing is not evidence)` (an earlier draft of the step, which kept ci.yml's `test -n` line, exited 1 SILENTLY here because `-e -o pipefail` aborts the failed assignment first; the guard was moved onto the assignment and the case re-observed)
- remainder with a deliberately unfiltered listing (scratch mutant: the `| grep -v -x ...` filter sed-deleted) -> `Exit code 1`, `FAIL: remainder package list still contains an excluded package (2 found); it would run twice`, go test never invoked
- control: the same unfiltered mutant with the REQ-RVS-014 guard lines also deleted -> body exit code 0 and argv lists `.../internal/cli` and `.../internal/harness/rosterguard` (the double run the guard prevents)

`moai spec lint .moai/specs/SPEC-CI-RELEASE-VERIFY-SHARD-001/spec.md --strict` after the status flip -> one INFO `OwnershipTransitionUnmeasured` (the plan commit has no trailer; same as the plan-phase baseline) then `0 error(s), 0 warning(s)`.

Boundary (M1 scope): `git diff --name-only 658c8e6cae HEAD` -> `.github/workflows/release-pr-multi-os.yml` and `.moai/specs/SPEC-CI-RELEASE-VERIFY-SHARD-001/spec.md`. `.github/workflows/ci.yml`, `scripts/`, `internal/`, `.moai/reports/` untouched. Also committed with this evidence: the machine-appended §G ceiling-outcome line (05:12:30Z) that was already uncommitted in `progress.md` at spawn start.

Baseline attribution: tree = branch WT-release-verify-os; base `658c8e6cae`; M1 workflow content = commit `e95e795219`; this run, this machine. Judging builds: actionlint v1.7.10 (`/Users/goos/go/bin/actionlint`), yq v4.53.2 (`/opt/homebrew/bin/yq`), the installed `moai` binary for `spec lint` (build commit not compared with the tree HEAD; see gap G-F).

Not run / open (Gaps):
- G-A: live `main` protection contexts (AC-RVS-005 read-only GET) — no `gh api` call made.
- G-B: AC-RVS-003 exact cover (`repro/shard-cover.sh` against the release workflow) and AC-RVS-009's four-name `go test -list` count — need the Go toolchain; deferred to the slot.
- G-C: AC-RVS-010/012 against REAL Go (a tiny package with a matching and a non-matching selector, census from a real `-json` stream) — only the stub-`go` control flow was observed; the stub-`go` print-then-fail run of `shard-cover.sh` (M4) not run.
- G-D: AC-RVS-011 (M2: `scripts/release.sh` leg lookup + `matrixScenarios()`), AC-RVS-002 and the M3-dependent halves of 001/007/009/013 (the parity test and its mutation table), AC-RVS-013's `go list ./...` partition identity (176/174) — not started (Go work).
- G-E: AC-RVS-012 windows half, AC-RVS-014, AC-RVS-016 — post-landing leader observations; nothing observed. Git Bash behaviour of the new step is unexercised.
- G-F: the installed `moai` build was not compared against the tree HEAD (verification-claim-integrity §2.2); the only `moai`-judged item is the INFO-only spec lint line.
- G-G: whether GitHub accepts `if: matrix.vet == true` / the include-merge as expanded here is read from actionlint's type model and my own expander, not from a live run; the first live run is the confirmation.

Residual risk:
- M1 alone breaks `scripts/release.sh --require-matrix-run` (leg names now `Release Verify (<os>, <shard>)`): it MUST NOT land without M2 (plan B-2). I did not run `release.sh` or its Go test.
- The status flip changes spec.md bytes, so the plan-artifact hash recorded at §E.1 (`744a64e8...`) no longer matches a recomputation; the plan-audit verdict was already consumed at kickoff (the same effect every prior card's draft to in-progress commit had). Recompute if a re-audit is wanted.
- macOS/Windows shard weight stays unmeasured (spec §G); remainder shard compile weight will grow once M3 adds the nested `go test -race -list` inside `internal/template`.
- The remainder shard on Windows runs `go list ./... | grep ...` under Git Bash; the quoting is unexercised there.

### M2 — release.sh leg lookup and its test (spawn 2 of the run phase; M3, M4 NOT started)

Commits (branch WT-release-verify-os, no push, staged by explicit pathspec): `68f741f807` test-only RED commit (`internal/template/release_script_test.go`); `f0c55db330` the `scripts/release.sh` lookup change. Test commit precedes the implementation commit, so the commit graph witnesses the order (verification-claim-integrity §2.3). Cycle: TDD. Resource rules kept: one Go call at a time, `-p 1`, `-run` narrowing to `^TestReleaseScript...`, no exclusions, no timeout change, no background load; machine load recorded per run. The bash that ran the script in the tests is `/bin/bash` (`command -v bash`), 3.2.57.

Go runs (one at a time; `uptime` taken immediately before; all in package `./internal/template/`, `-count=1 -p 1`):

| # | Command (`go test -count=1 -p 1 ... ./internal/template/`) | Script under test | Load (1m) before | Result |
|---|---|---|---|---|
| 1 | `-run '^TestReleaseScriptMatrixGate$/switch_on'` | unmodified (HEAD `5dbc39a3a0`), test = first draft | 17.35 (uptime 15:00) | exit 1, 9 `switch_on` scenarios FAIL (RED) |
| 2 | `-v -run '^TestReleaseScript'` | edited | 16.66 (15:01) | exit 0; all 9 `TestReleaseScript*` PASS, `ok 45.792s` |
| 3 | `-v -run '^TestReleaseScriptMatrixGate$'` | edited (final) | 17.94 (15:02) | exit 0; `--- PASS: TestReleaseScriptMatrixGate (17.99s)`, 30 subtests PASS (15 scenarios x 2 switch states) |
| 4-6 | `-run '^TestReleaseScriptMatrixGate$/switch_on'` | weakened copies m1, m2, m3 (below) | 17.43 / 16.54 / 19.06 | each exit 1 (see mutants) |
| 7 | `-v -run '^TestReleaseScript'` | final (content identical to the two commits; run before they were made) | 18.84 (15:04) | exit 0; 9 `TestReleaseScript*` PASS, `ok 47.395s`; `grep -c` of `MatrixGate` subtests = 30 |
| 8 | `-run '^TestReleaseScriptMatrixGate$/switch_on'` | unmodified (re-measured with the committed final test, HEAD `68f741f807`) | 16.53 (15:06) | exit 1, 10 `switch_on` scenarios FAIL (RED, exact) |

No run executed under load above 32 (peak seen 19.06), so no result needed re-running for load.

RED-now (AC-RVS-011, whole-script, tree `68f741f807` = test commit on the unmodified `scripts/release.sh`), run 8, verbatim FAIL lines:
`--- FAIL: TestReleaseScriptMatrixGate/switch_on/a_sharded_legs_all_green_on_the_target_sha`, `.../f_run_success_but_one_shard_skipped`, `.../h_run_success_but_one_shard_failed`, `.../i_one_shard_absent_on_a_single_os`, `.../j_the_same_shard_absent_on_all_three_os`, `.../k_an_unknown_shard_name_on_every_os_alongside_all_eighteen`, `.../o_an_unknown_shard_name_on_a_single_os_alongside_all_eighteen`, `.../l_a_duplicated_shard_job`, `.../m_an_os_with_no_shard_job`, `.../n_only_the_pre_shard_unsuffixed_leg_names`. The sharded all-green run is refused by the unmodified script: `[FAIL] Matrix gate: run 101 has no leg 'Release Verify (ubuntu-latest)'.` (exit=1, harness line `switch on, green run: exit code = 1, want 0`). Scenario n (only the pre-shard unsuffixed names) was ACCEPTED by the unmodified script (`switch on: exit code = 0, want 1`), i.e. the old lookup is exactly the exact-name shape. Correction of record: the `68f741f807` commit message says nine scenarios fail; run 1 (first-draft test) had nine, the committed test has ten (the unknown-shard scenario was split in two before the commit); the `f0c55db330` message carries the correction.

Green (run 3 and run 7): all 30 `TestReleaseScriptMatrixGate` subtests PASS, including `switch_on/a_sharded_legs_all_green_on_the_target_sha` with the acceptance line `[OK]   Matrix gate: run 101 of release-pr-multi-os.yml is green on all three OS legs (6 shards each) for 1bcab3f768a9` and the 14 refusal scenarios each refusing with the reason substring asserted: `leg 'Release Verify (windows-latest, cli-3)' is skipped`, `leg 'Release Verify (macos-latest, rosterguard)' is failure`, `has no leg 'Release Verify (macos-latest, remainder)'`, `has no leg 'Release Verify (ubuntu-latest, cli-4)'` (cli-4 missing on all three OS), `unknown Release Verify leg 'Release Verify (ubuntu-latest, cli-5)'` and `... (windows-latest, cli-5)'`, `has 2 jobs named 'Release Verify (macos-latest, cli-2)'`, `has no leg 'Release Verify (windows-latest, cli-1)'` (an OS with no shard job), `has no leg 'Release Verify (ubuntu-latest, cli-1)'` (only the pre-shard names). `switch_off` stays green for all 15 (no `gh run` call logged), and `TestReleaseScriptCharacterization`, `HeadAdmission`, `RcChangelogSkip`, `CIGate*`, `TagsTheVerifiedCommit` stay green.

Mutants (AC-RVS-011 mutant probe), each a deliberately weakened copy of the final `scripts/release.sh` built in scratch by `python3 -I mkmutants.py`, copied over `scripts/release.sh` for one `switch_on` run, then the final script restored (`cmp` against the saved copy -> no difference; `git status` clean afterwards):
- m1 "any one shard success per OS": runs 4; wrongly ACCEPTS (`exit code = 0, want 1`) f_skipped, h_failed, i_absent_on_one_os, j_same_shard_absent_on_all_three, k_unknown_on_every_os, o_unknown_on_one_os, l_duplicate. Correctly refused: m (an OS with no shard job), n (pre-shard names), a passes.
- m2 "compare shard sets across OS (all success) without the declared six": wrongly ACCEPTS j_same_shard_absent_on_all_three, k_unknown_on_every_os, l_duplicate (set semantics). Single-OS scenarios f/h/i/m/n/o stay refused, exactly the shape the AC predicts (passes every single-OS scenario, fails "same shard absent on all three OS" and "unknown shard").
- m3 "repeats tolerated" (the duplicate branch disabled): only `l_a_duplicated_shard_job` fails (accepted, `exit code = 0, want 1`).
The `FAIL` lines for m1-m3 also include reason-text mismatches where a mutant refuses with its own wording; the acceptance counts above are taken from the `exit code = 0, want 1` lines only.

`bash -n scripts/release.sh` and `/bin/bash -n scripts/release.sh` -> no output. `gofmt -l internal/template/release_script_test.go` -> no output. `repro/leg-match-probe.sh` against the edited script -> `Exit code 9`, `FAIL: exact-name lookup shape not found in .../scripts/release.sh (extraction)`: by its own contract exit 9 means the lookup no longer has the exact-name shape (it exited 0 on the base); the probe is intentionally NOT edited (plan-phase artifact, and its job was the RED-now cell of AC-RVS-011, now superseded by the whole-script scenario in run 8).

Boundary (M2 scope): `git diff --name-only 5dbc39a3a0 HEAD` -> `internal/template/release_script_test.go`, `scripts/release.sh` (plus this progress.md edit in the evidence commit). Workflow, `ci.yml`, `.moai/reports/` untouched. No template mirror exists for either file.

Baseline attribution: tree = branch WT-release-verify-os; code HEAD `f0c55db330` (script) on `68f741f807` (test); this run, this machine, Go toolchain as installed (not version-compared with the module's `go.mod` toolchain line; `go version` not recorded).

Gaps (M2):
- G-H: `golangci-lint` and `go vet ./internal/template/` were NOT run (kept to the scoped `-run` tests under the slot rules); `gofmt -l` is the only static check on the Go file, and `bash -n` on the script.
- G-I: the shard-id list in `scripts/release.sh` (`MATRIX_SHARDS=(...)`, one line, plus `MATRIX_OSES`) is asserted equal to the workflow matrix only by reading; the equality assertion is the M3 parity test (not started).
- G-J: the weakened copies are deviations I wrote myself, not an independent mutation tool; three mutants cover the three AC-named weakenings only.
- G-K: `release.sh` was exercised through the stubbed `gh` (the existing double) only; no real `gh run view` output was read, so the job-name format `Release Verify (<os>, <shard>)` is the one the M1 workflow edit produces, not an observed live run (AC-RVS-014 stays a post-landing observation).
- G-A..G-G from M1 are unchanged.

Residual risk (M2):
- The expected-name list is joined with `|` and passed to awk with `-v`; a job name containing `|` would also break the pre-existing `name|conclusion` record format (unchanged assumption).
- A SPEC-conform shard rename must edit `MATRIX_SHARDS`, the workflow matrix and the test tables together; only M3 makes that mechanical.

### M3 — drift and exact-cover guard (spawn 3 of the run phase; M4 NOT started)

Commit (branch WT-release-verify-os, no push, staged by explicit pathspec): `a6fb9edf4f` adds `internal/template/release_verify_shard_parity_test.go` (one new file, 1038 lines, `TestReleaseVerifyShardParity` plus `rvs*` helpers). M3 delivers no production code, so the test commit is the only code commit; it precedes this evidence commit in the graph (verification-claim-integrity §2.3). Base HEAD `0efa543e71`. Slot `go-test-heavy` held by this lane 06:07:44Z to 06:37:44Z; all Go calls below ran inside it, one at a time, `-p 1 -count=1` on every `go test`, no `-skip`, no timeout change, no background load. Every command was run by this agent; harness output is verbatim unless marked `(abridged)`. The harness showed an exit code only for failures (`Exit code 1`); an `ok`/`PASS` line stands in for exit 0 where it is quoted.

RED-now (AC-RVS-002, tree `0efa543e71`): `ls /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/internal/template/release_verify_shard_parity_test.go` -> `ls: ...release_verify_shard_parity_test.go: No such file or directory`, `Exit code 1`. The guard did not exist; the test was written after this observation and committed once it was green and its mutants observed.

Design as built: the names come from the go tool (`go test -race -list '.*'`, `rvsShardBuildRaw`, the single place `-race` is chosen), classified by `rvsClassify` (identifier lines only; `Benchmark...` and the trailing `ok` dropped; `Test...` to the cover check; anything else to the discovery guard). Selectors: ci.yml `test-race-<shard>` jobs through a `-run '<sel>' <pkg>` regex; the release workflow through its `include` entries. The remainder exclusions are the `-e '<pkg>'` arguments of the first `grep -v -x` line of each remainder step, not the whole line (M1 note). The matrix is expanded by `rvsExpand` (GitHub include/exclude rule), not the plain product. The shard and OS lists are read from `MATRIX_SHARDS` / `MATRIX_OSES` in `scripts/release.sh`. The fixture for mutant (x) is inlined (interpreted strings, so no source line begins `func Test...`), content identical to `repro/listing-fixture/race-pair/`.

Go calls (one at a time; `uptime` taken immediately before each; load never above 32, peak 13.77, so nothing needed a re-run for load):

| # | Command | Load (1m) before | Wall / result |
|---|---|---|---|
| 1 | `go test -race -list '.*' ./internal/cli/` (standalone, to file) | 8.46 (15:14) | `real 29.242s` (user 45.39s); 5984 lines = 5981 `Test...` + 2 `Benchmark...` (`BenchmarkIsTrivialCommand`, `BenchmarkTodoAuditAnalyze`) + the `ok` line `ok  	github.com/modu-ai/moai-adk/internal/cli	2.330s` |
| 2 | `go test -race -list '.*' ./internal/harness/rosterguard/` (standalone) | 9.76 (15:14) | `real 2.173s`; 20 `Test...` names + `ok  	github.com/modu-ai/moai-adk/internal/harness/rosterguard	1.289s` |
| 3 | `go test -count=1 -p 1 -v -run '^TestReleaseVerifyShardParity$' ./internal/template/` (first draft) | 9.26 (15:15) | `real 12.168s`; `--- PASS: TestReleaseVerifyShardParity (9.98s)`, 21 `--- PASS`, `ok  	github.com/modu-ai/moai-adk/internal/template	10.326s` |
| 4 | same, after adding the two text-level mutants and inlining the fixture as interpreted strings | 6.73 (15:16) | `real 12.667s`; `--- PASS: TestReleaseVerifyShardParity (10.31s)`, 23 `--- PASS`, 0 `--- FAIL`, `ok  	github.com/modu-ai/moai-adk/internal/template	10.731s` |
| 5 | same file with the collector deliberately weakened (`rvsShardBuildRaw` passes `false`), `-run '^TestReleaseVerifyShardParity$/^x_race_only_name_pair$'` | 7.26 (15:16) | `Exit code 1`; `run 2 (-race listing): want Other=[FuzzRaceOnly], got Test=[TestReal] Other=[] (a collector without -race gives run 1's result)`, `--- FAIL: TestReleaseVerifyShardParity/x_race_only_name_pair`; file restored, `shasum -a 256` before and after both `f1cade505fb081768afb33e4e47b39bfaa45895325fe0abea323254ea4160a83` |
| 6 | `go test -count=1 -p 1 -list '^(TestReleaseScriptMatrixGate\|TestTestInputFiltersSingleSource\|TestTestInputFiltersCoversGoTestInputs\|TestReleaseVerifyShardParity)$' ./internal/template/` | 13.77 (15:17) | prints exactly `TestReleaseScriptMatrixGate`, `TestReleaseVerifyShardParity`, `TestTestInputFiltersSingleSource`, `TestTestInputFiltersCoversGoTestInputs`, then `ok  	github.com/modu-ai/moai-adk/internal/template	0.185s` |
| 7 | `go test -count=1 -p 1 -v -run '<the same four-name regex>' ./internal/template/` (after the commit) | 12.82 (15:17) | `real 31.130s`; `--- PASS: TestReleaseScriptMatrixGate (18.41s)`, `--- PASS: TestReleaseVerifyShardParity (10.86s)`, `--- PASS: TestTestInputFiltersSingleSource (0.00s)`, `--- PASS: TestTestInputFiltersCoversGoTestInputs (0.00s)`, `PASS`, `ok  	github.com/modu-ai/moai-adk/internal/template	29.480s` |
| 8 | `go list ./...` (to file; partition count) | 9.01 (15:18) | `real 1.216s`; `wc -l` -> `176`; `grep -c -x -e .../internal/cli -e .../internal/harness/rosterguard` -> `2`; `grep -c -v -x` with the same two -> `174` (176 = 2 + 174) |
| 9 | `go vet ./internal/template/` | 10.07 (15:18) | `real 1.286s`, no output |
| 10 | `golangci-lint run --timeout=3m --new-from-rev=HEAD~1 ./internal/template/...` (golangci-lint v2.1.6, go1.26.8) | 9.66 (15:18) | `0 issues.` |

Nested `-race -list` cost (the residual-risk G7 item, measured on this machine, warm build cache after row 1): inside the test, `go test -race -list '.*' ./internal/cli/` took `6.111s` (run 3) and `6.002s` (run 4), the rosterguard listing `1.596s` / `1.614s`; the two fixture runs `708ms` (plain) and `1.468s` (-race) in run 3, `880ms` / `1.612s` in run 4. The whole test is `9.98s` / `10.31s`. The cold-cache cost on a hosted runner, and on macOS and Windows, is NOT measured here (gap G-N); the first live remainder-shard duration is the measurement (AC-RVS-016). Row 1 (`real 29.242s`) is the standalone listing in a session whose build cache for `internal/cli` under `-race` was not yet fully warm, so it is an upper-ish bound for a warm-ish machine, not a CI figure.

Flip evidence per AC (all from runs 3, 4, 5, 6, 7, 8 above):
- AC-RVS-002 (selector identity and cover): green, `--- PASS: TestReleaseVerifyShardParity/repository_state`. The in-test mutation table, each `--- PASS` in `mutants_must_fail` meaning the mutant WAS detected (a mutant not detected fails the subtest), first reported failure per mutant (run 4 log lines, verbatim): (i) `shard cli-3: selector diverges: .github/workflows/ci.yml has "^Test[J-R]", .github/workflows/release-pr-multi-os.yml has "^Test[J-Q]"`; (i-text, release workflow text edited before parsing) and (i-text, ci.yml text edited before parsing) both -> `shard cli-3: selector diverges`; (ii) `remainder exclusion list diverges: ... ci.yml has [github.com/modu-ai/moai-adk/internal/cli github.com/modu-ai/moai-adk/internal/harness/rosterguard], ...release-pr-multi-os.yml has [github.com/modu-ai/moai-adk/internal/cli]`; (iii) `internal/cli test "Test" matches 0 selectors [], want exactly 1`; (iv) shard 2 widened to `^Test[F-I]` with the synthetic name `TestFaInjected`: `shard cli-2: selector diverges: ...` and the required `test "TestFaInjected" matches 2 selectors` (the subtest asserts both substrings); (v) `discovery guard: go test -race -list lists "FuzzX" for internal/cli, and no ^Test selector would run it`, likewise `"ExampleX"`, bare `"Fuzz"`, bare `"Example"`, and `"FuzzR"` injected into the rosterguard list; (x) the paired fixture runs, below. Controls: `--- PASS: .../controls_must_pass/C1_injected_BenchmarkX` (a benchmark name does not fail) and `.../C2_injected_TestZ` (a name matching exactly one selector passes).
- Mutant (x), the pair on the clean-base fixture package, no `-vet=off`, run 4 log: run 1 `go test -list .* ./... ... took 880ms` (exit 0, names `TestReal`, `BenchmarkB`; the subtest asserts Test=[TestReal], no other names, both guards pass); run 2 `go test -race -list .* ./... ... took 1.612s` (exit 0) and `mutant x detected only under -race: discovery guard: go test -race -list lists "FuzzRaceOnly" for fixture, and no ^Test selector would run it`. Both runs exited 0 (a non-zero exit is a `t.Fatalf` in `rvsListRaw`), so the plan's "read, not executed" vet-analyzer note for the fixture is now observed: the fixture builds and lists under go vet's default analyzers. The collector-without-`-race` variant was observed rejected (row 5).
- AC-RVS-003 (test half): `repository_state` PASS means every one of the 5981 internal/cli `Test...` names the tool printed under `-race` matches exactly one of the four release selectors (the test logs `internal/cli lists 5981 Test names (+0 other), rosterguard 20 (+0 other)`); the 5981 equals the plan-time `shard-cover.sh` count. The script half (`repro/shard-cover.sh` against the release workflow, exact-cover over plain `go test -list`, and its narrowed and widened selector mutants) is M4 and was NOT run (gap G-L).
- AC-RVS-001 (expander halves): `repository_state` PASS asserts no `exclude` key, 6 include entries each shard-keyed with no `os` key, 18 expanded legs, 18 distinct `(os, shard)` pairs, the shard and OS lists equal to `MATRIX_SHARDS` / `MATRIX_OSES`. Mutants detected: (viii) `exclude: [{os: ubuntu-latest, shard: remainder}]` -> `matrix carries an exclude key; the final leg set must be the full os x shard product` (also `17 legs` and `leg (ubuntu-latest, remainder) is missing`); (ix) duplicated include -> `include has 7 entries, want 6 (exactly one per shard id)`, and the subtest first asserts that `rvsExpand` still returns 18 legs for it (the SPEC's prediction that the count alone cannot catch it held); (xi) the `remainder` shard dropped -> `matrix shard ids [cli-1 cli-2 cli-3 cli-4 rosterguard] differ from MATRIX_SHARDS [cli-1 cli-2 cli-3 cli-4 rosterguard remainder] in scripts/release.sh`.
- AC-RVS-007: 18 artifact names pairwise distinct, each containing its OS and shard (asserted in `repository_state`); mutant (vi) -> `artifact name "test-stream-release-verify-ubuntu-latest" for leg (ubuntu-latest, cli-1) does not carry both its os and its shard` (and `shared by` colliding legs).
- AC-RVS-009: row 6 prints exactly four names (the AC's count: three at base plus this one) and row 7 shows a `--- PASS:` line for each of the four. `actionlint` and the validator were not re-run in this spawn (the workflow is unchanged since M1; their M1 results stand). Note: the AC's literal command names `TestTestInputFiltersSingleSource` and `TestTestInputFiltersCoversGoTestInputs`; these two existing tests only read files (0.00s each) and were included for that reason, although the spawn scoped Go runs to "the new test and the existing release script tests".
- AC-RVS-013: 3 legs with `vet: true` in the expanded matrix (asserted; mutant (vii) second `vet: true` include -> `6 legs run go vet, want 3 (once per os, not once per shard)`). Partition identity: the test asserts the remainder exclusion list equals the packages the other shards name (module path from go.mod + `/internal/cli`, `/internal/harness/rosterguard`); row 8 gives the real counts on this tree, 176 total, 2 excluded, 174 remainder.

Boundary (M3 scope): `git diff --name-only 0efa543e71 a6fb9edf4f` -> `internal/template/release_verify_shard_parity_test.go` only. `.github/workflows/ci.yml`, the release workflow, `scripts/`, `.moai/reports/` untouched; no template mirror exists for the new file.

Baseline attribution: tree = branch WT-release-verify-os; test content = commit `a6fb9edf4f` (shasum of the file at that content `f1cade505fb081768afb33e4e47b39bfaa45895325fe0abea323254ea4160a83`); this run, this machine; Go toolchain `go version go1.26.8 darwin/arm64` (equal to the `go 1.26.8` line of go.mod); golangci-lint v2.1.6. `git status --short` was empty after the commit.

Gaps (M3):
- G-L: `repro/shard-cover.sh` against both workflows and its two selector mutants, the stub-`go` print-then-fail run, and the other M4 items (pre-flight set re-run, the real-Go step-body runs for AC-RVS-010/012, the plan-boundary diff for AC-RVS-015) were NOT run: M4.
- G-M: the cover check runs over the names `go test -race -list` prints for the running OS; a name that exists only under another OS's build tags is not seen on this macOS run (the three-OS release matrix is the cover for that). The test is untested on Windows and Linux in this spawn (nested `go` resolution through `exec.Command("go")` on those runners is unobserved).
- G-N: the cold-cache cost of the two nested race compiles on hosted ubuntu, macOS and Windows runners, and their effect on the `remainder` shard duration, is not measured (only the warm-cache figures above); the AC-RVS-016 first live run measures it.
- G-O: lint scope: `golangci-lint ... --new-from-rev=HEAD~1 ./internal/template/...` (only issues new in the M3 commit) and `go vet ./internal/template/`; no full-repository lint; `make fmt-check` not run, `gofmt -l` on the file printed nothing.
- G-P: the job display name template (`Release Verify (${{ matrix.os }}, ${{ matrix.shard }})`) that `scripts/release.sh` matches whole is not asserted by this test (the plan's M3 list does not prescribe it); the shard and OS lists are. Suggested follow-up for the leader: one more assertion and mutant if the coupling should be guarded.
- G-A..G-K from M1 and M2 are unchanged.

Residual risk (M3):
- The nested `go test -race -list` makes the `internal/template` test binary shell out to the Go toolchain; on a runner without cgo or a race-capable toolchain the test fails rather than skips (deliberate: a skip would be a vacuous pass).
- The `internal/cli` `TestMain` runs on each nested listing (measured effect: the listing takes about 6s warm; whether it prints to stdout is covered by `rvsClassify` ignoring non-identifier lines, and by the discovery guard if such a line looked like a name; run 4 reported `+0 other`).
- Text-level mutants depend on the literal anchor text `run: '^Test[J-R]'` (release workflow) and `-run '^Test[J-R]' ./internal/cli/` (ci.yml); `rvsReplaceOnce` fails loudly (not silently) if either is reformatted.

### M4 — evidence and boundary (spawn 4 of the run phase; evidence-only, no source/workflow/test-file change)

Evidence commit only: this progress.md edit (staged by explicit pathspec). HEAD before it `285f133e91` (the M3 evidence commit), tree clean at spawn start and at commit time. Slot `go-test-heavy` held by lane-55 `13:18:30Z`-`13:58:30Z`; every Go call below ran inside it, ONE AT A TIME, `-p 1` and `-count=1` on every test invocation, no `-skip`, no timeout change, no background load. Before the first `go` invocation: `uptime` -> `load averages: 20.79 22.24 17.82` (1-min 20.79 < 32) and `memory_pressure | tail -1` -> `System-wide memory free percentage: 86%` (> 50%). Total Go wall-time this spawn: about 2.5 minutes (well inside the 35-minute bound). Scratch (two stub `go` binaries, two selector-mutant workflow copies, the filter-deleted remainder body, the tiny fixture) lived under `/tmp/t1892-m4/` and the gitignored `temp/` dir; all deleted after the runs (the fixture dir removed file-by-file after an `rm -rf` was denied by the permission system — the narrow removals succeeded). Every command below was run by this agent; outputs are verbatim. As in M1-M3, the harness reports an exit code only for failures (`Exit code N`); for exit-0 commands the decisive output line stands in.

#### M4 item results

Pre-flight set (plan §C):

| Item | Command | Observed output | Exit | Verdict |
|---|---|---|---|---|
| HEAD/branch | `git rev-parse --short HEAD` / `git branch --show-current` | `285f133e91` / `WT-release-verify-os` | 0/0 | matches the handoff premise |
| actionlint | `actionlint <root>/.github/workflows/release-pr-multi-os.yml` (plain single command, absolute path) | no output (clean) | 0 | PASS |
| validator | `sh scripts/ci-mirror/validate-required-checks.sh` | `✅ All validations passed` incl. `✓ [main] Release PR Multi-OS Gate is publishable` | 0 | PASS |
| verifier suite | `python3 scripts/ci-mirror/test/verifier-fail-closed_test.py` | `Ran 5 tests in 6.609s` / `OK` | 0 | PASS |

Note on actionlint: the worktree guard did NOT refuse the plain single-command form this time (the M1 lane re-check refusal was of a compound construct; the plain form used here passes the guard), so the M1 gap G-actionlint is closed by this observation: actionlint clean on the sharded workflow at `285f133e91`.

AC-RVS-009 swept count (item 1): `go test -count=1 -p 1 -list '^(TestReleaseScriptMatrixGate|TestTestInputFiltersSingleSource|TestTestInputFiltersCoversGoTestInputs|TestReleaseVerifyShardParity)$' ./internal/template/` printed exactly the four names `TestReleaseScriptMatrixGate`, `TestReleaseVerifyShardParity`, `TestTestInputFiltersSingleSource`, `TestTestInputFiltersCoversGoTestInputs` then `ok  	github.com/modu-ai/moai-adk/internal/template	0.383s`. Four names (three at base + the new parity test): PASS.

AC-RVS-009 run half (item 9): `go test -count=1 -p 1 -run '^(TestReleaseScriptMatrixGate|TestTestInputFiltersSingleSource|TestTestInputFiltersCoversGoTestInputs|TestReleaseVerifyShardParity)$' ./internal/template/` -> `ok  	github.com/modu-ai/moai-adk/internal/template	32.874s`, exit 0. PASS.

AC-RVS-010 stub-`go` print-then-fail (item 2): stub `go` at `/tmp/t1892-m4/bin/failgo/go` prints four `TestFake*` names to stdout then exits 42; `PATH=/tmp/t1892-m4/bin/failgo:$PATH` shard-cover.sh against the release workflow -> `Exit code 1`, `FAIL: go test -list exited 42 on the full listing (a failed listing is not evidence)`. The printed names were refused: PASS (FAIL-as-expected). (The env-prefix compound form was refused by the worktree guard — refusal quoted under Refusals — so the run went through a scratch runner wrapper `bash /tmp/t1892-m4/run-failgo.sh`, the B-8-sanctioned script form.)

AC-RVS-013 remainder guard, unfiltered listing (item 3): the step body was extracted verbatim with `yq '.jobs.full-matrix-test.steps[] \| select(.name == "Run tests with race detector") \| .run'` to `/tmp/t1892-m4/step-body.sh`; a scratch copy with ONLY the `| grep -v -x -e 'github.com/modu-ai/moai-adk/internal/cli' -e 'github.com/modu-ai/moai-adk/internal/harness/rosterguard'` filter deleted from the `pkgs=` assignment (`diff` of the two shows exactly that one line); stub `go` (`/tmp/t1892-m4/bin/unfiltered/go`) answers `list ./...` with a four-package listing INCLUDING both excluded packages and exits 99 on any other subcommand. Run under the Actions shell shape (`bash --noprofile --norc -e -o pipefail`) via the wrapper -> `Exit code 1`, `FAIL: remainder package list still contains an excluded package (2 found); it would run twice`; the stub's exit-99 trap was never reached, i.e. `go test` was never invoked. PASS (FAIL-as-expected).

AC-RVS-003 exact cover (item 5), each a single run, five `go test -list` invocations per script run:
- vs `ci.yml`: `bash .moai/specs/SPEC-CI-RELEASE-VERIFY-SHARD-001/repro/shard-cover.sh <root> /tmp/t1892-m4/out/ci .github/workflows/ci.yml` -> selectors `2111 / 862 / 1851 / 1157` names, `full=5981 union=5981 union_unique=5981`, `EXACT COVER: full list == union of the four selectors (no gap, no overlap)`, exit 0. PASS (the L-E9 baseline reproduces on this tree).
- vs the release workflow (matrix `include` extraction path): same four counts, `full=5981 union=5981 union_unique=5981`, `EXACT COVER`, exit 0. PASS.
- narrowed mutant (`^Test[J-R]` -> `^Test[J-Q]`, scratch copy `/tmp/t1892-m4/mutants/release-narrowed.yml`): `Exit code 1`, selector 3 -> `1042 names`, `full=5981 union=5172`, `NOT AN EXACT COVER` with `TestR1_...`/`TestR...` names in the diff — 809 names dropped, exactly the L-E10 figure. PASS (FAIL-as-expected).
- widened mutant (shard 2 `[G-I]` -> `[G-J]`, `/tmp/t1892-m4/mutants/release-widened.yml`): `Exit code 1`, selector 2 -> `898 names`, `full=5981 union=6017 union_unique=5981` — 36 `TestJev...`/`TestJ...` names matched two selectors; `NOT AN EXACT COVER`. The no-overlap mutant fails too: PASS (FAIL-as-expected).

AC-RVS-010/012 step body under `/bin/bash` 3.2.57 with REAL Go (item 10): the same extracted body run verbatim under `bash --noprofile --norc -e -o pipefail`, tiny fixture package `temp/t1892-tiny` (gitignored, in-module so the real go tool resolves it; one test `TestTinyMatches`; deleted after):
- FIRST sited the fixture OUTSIDE the module (`/tmp/t1892-m4/fixture/tiny`): the body correctly refused it — census printed `BUILD FAILED  /tmp/t1892-m4/fixture/tiny` / `directory ... outside main module or its selected dependencies`, `=== totals: packages=1 passed=0 ... build-failed=1 ===`, then `FAIL: shard ran no test (empty sweep)`, exit 1. Recorded because it shows the body refuses a build-failed sweep too; a fixture-siting error on my side, not a body defect. Re-sited in-module; the two listed runs below are each a single run:
- matching selector `SHARD_RUN='^Test'`: exit 0, census `=== totals: packages=1 passed=1 skipped=0 nothing-ran=0 failed=0 build-failed=0 ===`. PASS (AC-RVS-012 local half; `test-stream.json` was written non-empty — the body's own `grep -q '"Test":"'` guard passed).
- non-matching selector `SHARD_RUN='^NoSuchTest$'`: `Exit code 1`, census `packages=1 passed=0 ... `, `FAIL: shard ran no test (empty sweep)`. PASS (FAIL-as-expected; AC-RVS-010).
- no-interpolation grep (AC-RVS-012): `grep -n -F '${{' /tmp/t1892-m4/step-body.sh` -> no output; `grep -c SHARD_RUN <wf>` -> `3`. PASS.

yq assertions (item 11), workflow at `285f133e91`:
- AC-RVS-001: `yq -o=json -I=0 '.jobs.full-matrix-test.strategy.matrix \| {"os": .os, "shard": .shard, "has_exclude": has("exclude"), "include_len": (.include \| length), "include_keys": [.include[] \| keys \| join(",")]}'` -> `{"os":["ubuntu-latest","macos-latest","windows-latest"],"shard":["cli-1","cli-2","cli-3","cli-4","rosterguard","remainder"],"has_exclude":false,"include_len":6,"include_keys":["shard,pkg,run","shard,pkg,run,vet","shard,pkg,run","shard,pkg,run","shard,pkg,run","shard,pkg,run"]}`; `yq '.jobs.full-matrix-test.strategy."fail-fast"'` -> `false` (an earlier compound form quoted `."fail-fast"` inside the matrix context and printed `null` — wrong path, fail-fast sits under `strategy`, corrected above). The 18-leg expansion, 18 distinct pairs and 18 distinct artifact names are the M3 parity test's assertions (recorded there).
- AC-RVS-004: `git show 74ad197380:.github/workflows/release-pr-multi-os.yml` -> scratch; `yq '.jobs.release-pr-gate'` extracted from base and from the working tree; `diff` -> no output; `shasum -a 256` both `068c049bd7a2df013f69c1af8fd3b7b8dc622719652cec04b93a4d711228c1d5` (the same digest M1 recorded); `grep -n "needs:"` on the extraction -> `needs: [full-matrix-test, detect-release]`. PASS.
- AC-RVS-006: `grep -n -e "-timeout" -e "timeout-minutes" -e "-skip" -e "continue-on-error" -e "retry-flaky" <wf>` -> job `timeout-minutes: 40` (line 121), both `go test` at lines 300/302 carry `-timeout 35m`, no `-skip` and no `retry-flaky` anywhere; `continue-on-error` hits are the pre-existing detect-release step (line 64) and comments (57, 103, 156). PASS.
- AC-RVS-007: `yq '.jobs.full-matrix-test.steps[] \| select(.name == "Upload test event stream") \| .with.name'` -> `test-stream-release-verify-${{ matrix.os }}-${{ matrix.shard }}`. PASS (18 distinct names asserted by M3).
- AC-RVS-008: `yq '.jobs.full-matrix-test.steps[].name'` -> the same eleven names as M1, same order (Checkout code, Setup Go, Download dependencies, Install ast-grep (provides `sg`), Verify `sg` resolves to ast-grep, Run go vet, Diagnose ancestor .moai (pre-test), Run tests with race detector, Compress test event stream, Upload test event stream, Diagnose ancestor .moai (post-test)). PASS.
- AC-RVS-013 (vet-once): `yq '[.jobs.full-matrix-test.strategy.matrix.include[] \| select(.vet == true)] \| length'` -> `1` (x 3 OS = 3 legs); `yq '.jobs.full-matrix-test.steps[] \| select(.name == "Run go vet") \| .if'` -> `matrix.vet == true`. PASS.

AC-RVS-015 plan-boundary diff (item 12): `git diff --name-only 74ad197380 658c8e6cae` (74ad197380 = plan-phase base tree; 658c8e6cae = the last plan-phase commit, the audit-ready signal) -> 14 paths, every one starting `.moai/specs/SPEC-CI-RELEASE-VERIFY-SHARD-001/` (acceptance.md, decision-index.md, plan.md, progress.md, repro/leg-match-probe.sh, repro/listing-fixture/README.md, repro/listing-fixture/behaviour/{a_test.go.txt,go.mod.txt}, repro/listing-fixture/race-pair/{base_test.go.txt,go.mod.txt,race_only_test.go.txt}, repro/shard-cover.sh, research.md, spec.md); nothing under `.github/`, `scripts/`, `internal/`, `.moai/reports/`. PASS.

Refusals (named per verification-claim-integrity §3.1): the worktree guard refused the env-prefix compound `PATH=... SHARD_PKG=remainder SHARD_RUN= bash --noprofile --norc -e -o pipefail /tmp/t1892-m4/step-body-unfiltered.sh` with `this command runs bash inside a construct too complex to verify; what it reads or is handed as shell text cannot be shown not to run git`. Remedy: the runner-wrapper form (`bash /tmp/t1892-m4/run-remainder-guard.sh`, env set inside the script), which is the plan's own B-8 prescription for multi-step runs; the measurement is the script's, unchanged. The permission system also denied `rm -rf` of the temp fixture dir; narrow per-file removals succeeded and the tree is clean.

Boundary (M4 scope): evidence-only — `git status --short` clean before this edit; the only file this spawn modifies is this progress.md. No source, workflow, test, or template-mirror file touched.

Baseline attribution: tree = branch WT-release-verify-os, HEAD `285f133e91` for every command above; this run, this machine. Judging builds: actionlint (installed, clean run; version recorded at plan time as v1.7.10), yq v4.x (`/opt/homebrew/bin/yq`), go1.26.8 darwin/arm64 (unchanged since M3; go.mod toolchain equal), python3 (system), the repro scripts as committed at `658c8e6cae` content (unchanged since; `git status` clean).

Gaps (M4):
- G-A (carried): live `main` protection contexts (AC-RVS-005 read-only GET) — still no `gh api` call this spawn; the validator's static dimensions re-passed, the live-context read remains the leader's.
- G-E (carried): AC-RVS-012 windows half, AC-RVS-014, AC-RVS-016 — post-landing leader observations; nothing observed.
- G-G (carried): GitHub's acceptance of `if: matrix.vet == true` and the include-merge as expanded remains read from actionlint's type model + the test expander; the first live run confirms.
- The stub-`go` runs (item 2) and the remainder-guard run (item 3) observe the scripts' control flow, not Go's behaviour; the real-Go cover is item 5 and M3 rows 1-2.
- AC-RVS-013's partition identity (176 = 2 + 174) is the M3 measurement (row 8); not re-run at M4 (same-HEAD single-run discipline; the task list does not re-list it).

Residual risk (M4):
- The narrowed/widened mutants were authored by me (scratch copies, single-line replacements; `diff`-verified) — the same self-authored-mutant caveat as M2's G-J.
- `temp/` (gitignored) held the fixture during item 10 and is now removed; if the directory is reused by another session concurrently, the census output above could in principle have been polluted — the census lines read `packages=1`, consistent with the fixture alone.
- The M4 exit-0 commands have no harness-reported exit code (M1-M3 harness limitation); the decisive output lines are the evidence.

#### AC-RVS-014 / AC-RVS-016 leader handoff (post-landing; not closable by the lane)

AC-RVS-014 (PL): after the change reaches PR #1830's head ref (or its successor release PR) and the run completes, run:

```
gh run view <run-id> --json jobs --jq '.jobs[] | {name, conclusion, startedAt, completedAt}'
```

- PASS condition: all 18 `Release Verify (<os>, <shard>)` jobs `conclusion == "success"` (none `cancelled`), and `Release PR Multi-OS Gate` `success`. Record the verdict in this file (or the card verdict) with the verbatim output.
- AC-RVS-016 (same run): each shard job's `startedAt`/`completedAt` gives the 18 durations; record all 18 on the card. Any shard over 35 minutes of test time or within 5 minutes of the 40-minute job budget (or `cancelled`) opens a rebalance card (split that shard by selector) — never a timeout raise.
- D9 flag (leader's call, unchanged from §E.1): an optional pre-landing measurement via `workflow_dispatch` of the release workflow on the candidate ref; one added run, results attach to the PR.
- Current reference state (not a pass): L-E4 — run 39180942620 all legs `cancelled`, gate `failure`, per-OS shard durations nonexistent (gap G2).

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-10-19
run_commit_sha: 3c52f2d150 (the M4 evidence commit; this §E.3 edit itself lands in a later commit by the self-reference rule)
run_milestone_commits: ea831e5ec4 + e95e795219 (M1 status flip + workflow), 68f741f807 + f0c55db330 (M2 test RED then script, commit-graph ordered), a6fb9edf4f (M3 parity test), 5dbc39a3a0 / 0efa543e71 / 285f133e91 / 3c52f2d150 (M1-M4 evidence records)
ac_pass_count: 14
ac_fail_count: 0
ac_post_landing_pl: 2 (AC-RVS-014, AC-RVS-016 — plus the windows half of AC-RVS-012; leader observations, handoff written at M4)
m4_items: 13/13 PASS (swept count, stub-go print-then-fail, remainder guard unfiltered listing, pre-flight set, shard-cover both workflows + narrowed/widened mutants, actionlint, validator, verifier suite, targeted internal/template run, real-Go step body matching + non-matching under /bin/bash 3.2, yq assertions, AC-RVS-015 plan-boundary diff, AC-RVS-014/016 leader handoff)
m4_refusals_recorded_as_gaps: worktree-guard refusal of the env-prefix compound (remedy: scratch runner wrappers, B-8 script form); permission-system denial of `rm -rf` (remedy: narrow per-file removals)
carried_gaps: G-A (live main-protection contexts not re-read at run phase — validator dimensions pass, `gh api` read remains with the leader), G-E (AC-RVS-012 windows half, AC-RVS-014, AC-RVS-016 — post-landing), G-G (GitHub include-merge + `if: matrix.vet == true` confirmed only by actionlint's type model and the test expander until the first live run), stub-`go` runs observe control flow not Go behaviour, AC-RVS-013 partition identity (176 = 2 + 174) cited from the M3 measurement (not re-run at M4)
go_budget: slot go-test-heavy (lane-55) 13:18:30Z-13:58:30Z; Go work finished ~13:31Z; load 1-min 20.79 / memory free 86% at the first `go` call; every Go call one at a time, `-p 1 -count=1`, no `-skip`, no timeout change

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: complete (3-phase close — plan → run → sync, this commit)
sync_complete_at: 2026-10-19
sync_commit_sha: c9f31e8097
close_summary: "Release Verify sharded 6 shards x 3 OS (18 legs, gate aggregation unchanged), scripts/release.sh --require-matrix-run resolves the sharded legs; the CHANGELOG entry, the spec.md status transition, and this signal land in the single sync commit; sync_commit_sha is backfilled in the next commit (a commit cannot cite its own hash)."
changelog_entry_position: CHANGELOG.md [Unreleased] § Changed, first entry
frontmatter_status_transitions:
  - "spec.md status: in-progress -> completed (this single sync commit; the intermediate implemented step consolidated per the Status Transition Ownership Matrix close row)"
updated_field_refresh: "spec.md updated: already 2026-10-19 — no byte change; plan.md / acceptance.md optional updated: refresh skipped per dispatch; spec.md / plan.md / acceptance.md body content untouched"
b12_self_test_a: >-
  pre-emission grep -c 'SPEC-CI-RELEASE-VERIFY-SHARD-001' CHANGELOG.md -> 0
  (halt condition not triggered)
b12_self_test_b: >-
  AC count match — moai-ac counter over acceptance.md (tier M, ac_source
  .moai/specs/SPEC-CI-RELEASE-VERIFY-SHARD-001/acceptance.md, resolved and
  non-empty): default-AC grammar live=17 = 16 declared criteria
  (AC-RVS-001..016) + 1 cross-namespace citation AC-CI-011 (line 54, cites
  SPEC-CI-VERDICT-INTEGRITY-001; the [REF] exclusion token cannot be written
  by this agent — acceptance.md body is forbidden to manager-docs); scoped to
  the file's declared AC-RVS namespace: live=16 excluded=0 ambiguous=0,
  exit 0; the CHANGELOG entry cites 16 (AC-RVS-001..016), matching the
  declared count
b12_self_test_c: >-
  file path verification — every path cited in the CHANGELOG entry confirmed
  present in this tree by ls (release-pr-multi-os.yml, scripts/release.sh,
  internal/template/release_verify_shard_parity_test.go,
  internal/template/release_script_test.go, SPEC spec.md and progress.md)
mx_tag_scan: "run-phase code surface = 2 Go test files (internal/template/release_script_test.go, internal/template/release_verify_shard_parity_test.go) + scripts/release.sh (shell) + workflow YAML (release-pr-multi-os.yml); no new exported production Go identifier; @MX:NOTE/@MX:ANCHOR obligations: none; @MX:WARN candidates (goroutine, complexity >=15): none; scan result: no MX tag needed (expected none)"
docs_publication: "docs-site / README 4-locale: no update needed — CI-internal change (release pipeline internals + the maintainer-only --require-matrix-run leg resolution); no shard or leg-name surface exists in public docs; judgment: unchanged, nothing edited"
```

## §G Override and Refusal Record

- 2026-10-19T03:41:18Z SPEC-CI-RELEASE-VERIFY-SHARD-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 7 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter7.md
- 2026-10-19T03:41:24Z SPEC-CI-RELEASE-VERIFY-SHARD-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 7 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter7.md
- 2026-10-19T05:12:30Z SPEC-CI-RELEASE-VERIFY-SHARD-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 7 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter7.md
