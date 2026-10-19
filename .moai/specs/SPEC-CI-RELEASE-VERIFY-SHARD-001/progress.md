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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §G Override and Refusal Record

- 2026-10-19T03:41:18Z SPEC-CI-RELEASE-VERIFY-SHARD-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 7 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter7.md
- 2026-10-19T03:41:24Z SPEC-CI-RELEASE-VERIFY-SHARD-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 7 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter7.md
- 2026-10-19T05:12:30Z SPEC-CI-RELEASE-VERIFY-SHARD-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 7 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1892/.moai/reports/t1892/plan-audit-iter7.md
