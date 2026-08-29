# SPEC-CI-TEST-OBSERVABILITY-001 — Progress

Card: **t358** · Branch: `WT-ci-test-observability` · Base: `origin/develop c6aa61346`

## §E.1 Plan-phase Audit-Ready Signal

- Artifact set authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M).
- SPEC ID regex self-check executed as Bash: `[[ "SPEC-CI-TEST-OBSERVABILITY-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS`.
- ID uniqueness confirmed against `.moai/specs/` (no prior `SPEC-CI-TEST-OBSERVABILITY-*`).
- Frontmatter carries all 12 canonical fields; `phase: "v3.1.4 target"` is a release target, not a lifecycle stage.
- Terminal AC is the positive control (`acceptance.md` AC-CTO-007), not a YAML-content assertion.
- Addendum folded in (coordinator, post-authoring): single-predicate design gain recorded as the deciding reason (`spec.md` §C.1); failure-body survival and `-json`+coverage coexistence recorded as observed, not assumed; scope stated at three call sites [**SUPERSEDED** — a fourth, `ci.yml:329`, was found in plan audit iter-1 and brought in scope by lead ruling. Preserved rather than rewritten per `verification-claim-integrity.md` §2: silently rewriting a measured-false claim would erase the record that this SPEC once asserted a completeness it did not have]; ONE-run dispatch constraint written into AC-CTO-007; house conventions (`upload-artifact@v7`, `retention-days: 7`, `jq` established) recorded.
- **Plan audit round 1: FAIL 0.81.** Nine blocking defects repaired (D1-D9) plus two citation corrections; D10/D11 addressed. Repair detail in the return report. AC-CTO-007 survived the adversarial pass unchanged and was NOT weakened — the auditor independently verified its two premises (`internal/statusline/usage_test.go:186` unconditional `t.Skip`; `profile_bench_test.go:305-307` env-gated with neither workflow setting the var), so the single dispatch is guaranteed to contain a real skip.
- **Scope corrected to FOUR call sites.** `ci.yml:329` (`test-integration`, 3-OS matrix, `-tags=integration`) was missed in the first draft and is in scope by lead ruling.
- **Named debt, carried openly (never reported as passes):**
  1. AC-CTO-003's CI-level red-path confirmation — verified locally instead; the CI red path is not exercised pre-merge.
  2. AC-CTO-005b (artifact present on a FAILED run) — `if: always()` is a declaration of intent, not observed behaviour.
  Both discharge on the first genuinely red CI run after this lands. Closing either by manufacturing a second dispatch is prohibited (lead's ONE-run constraint is [HARD]).
- If the fallback observation path is taken, AC-CTO-007 closes post-merge and that too is debt.
- Full-suite artifact size remains an EXTRAPOLATION from two packages; the real figure is recorded at AC-CTO-005a from the approved dispatch, which also supersedes the extrapolation label.
- **Plan audit iter-2: PASS-WITH-DEBT 0.895** (+0.085 monotonic; Tier M threshold 0.80 — also clears Tier L's 0.85). Final text-only pass applied (N1-N9): stale 3-site count corrected, REQ order fixed, AC-CTO-003 split so each criterion is executable exactly as written, DoD reconciled with §I, remote-ref cleanup added with its ordering constraint, plus four precision fixes.
- **Pattern-label correction**: iter-1's D11 was **withdrawn** — `Unwanted` is an EARS-legacy name absent from the canonical GEARS five (Ubiquitous, Event-driven, State-driven, Capability gate, Event-detected), so iter-1 had pushed away from the canonical name toward a non-canonical one. REQ-CTO-004 is restored to `(Event-detected)`; the three always-active prohibitions (REQ-CTO-003, -009, -011) carry `(Ubiquitous)`.
- **Prose-accuracy correction after M1-M3 (2026-08-29).** The SPEC asserted, in `spec.md` §F constraint 2 and `plan.md` §B.2, that build errors travel on **stderr** and that a stdout-only redirect therefore keeps them console-visible. **Measurement falsified this**: `go test -json` on a package with a syntax error yields stdout 666 B / **stderr 0 B**, with diagnostics carried as `build-output` events terminated by `build-fail`. Measured by the coordinator in a throwaway module, then reproduced independently by this agent (darwin/arm64, go 1.26.4, identical action sequence). The earlier claim is **corrected in place with the correction recorded here** rather than silently rewritten, per `verification-claim-integrity.md` §2 — the SPEC did assert something false, and that record is worth keeping.
- Consequence now stated in the text: the census `BUILD FAILED` row is **load-bearing**, because it — not stderr — is what restores console visibility of a build failure. REQ-CTO-011's requirement is unchanged; only its rationale was wrong (stream integrity, not console visibility). AC-CTO-003b's pass criteria are unchanged; only the stated mechanism was corrected.
- **Two items recorded in the new `spec.md` §J** — residual risk: `test-stream.json` is visible in the repo root mid-run to any test walking the filesystem (gitignored, unmeasured, not a defect). Named gap: the census has run on **darwin/arm64 only**, and all four `jq`-using workflows in this repo are `ubuntu-latest`, so the dispatch is the census's first exposure to windows and macOS.
- Status: `draft`. Awaiting Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

Baseline-attribution for every row below: measured in this run, in worktree
`.claude/worktrees/t358`, against tree `1a635aea8` (the run-phase base; the
implementation commits are `13ec99545` M1 and `d89998b13` M2+M3). Milestones
M1-M3 only — **M4 (push, dispatch, positive-control observation) is the lead's
and was NOT performed here.**

### AC matrix

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-CTO-001 | PASS | `bash --noprofile --norc -e -o pipefail .moai/reports/t358/evidence/step-body-correct.sh` then `wc -c -l` and `grep -c '^{"Time":'` | console **2 lines / 105 bytes**, `grep -c '^{"Time":'` → **0**. Pre-change console for the same package: 1 line / 68 bytes. Ratio **1.54×**, not the 913× `-v` was measured to cause. 44,757 bytes / 221 lines went to the stream FILE. |
| AC-CTO-002 | PASS | `bash scripts/ci-census/census-check.sh` | `census-check: PASS (census output matches …/testdata/expected.txt)`, rc=0. Mutant probe below. |
| AC-CTO-003 | PASS (CI-level red path = DEBT, see §E.2 debt) | exact step body under `bash --noprofile --norc -e -o pipefail`, tree carrying one planted failing test | step **rc=1**; census **printed**; named the test and its captured output — `FAILED  …/internal/hook/mx/complexity  TestT358DeliberateFailure` / `t358_probe_test.go:7: t358 probe: expected 7, got 3`. Naive `; rc=$?` form on the SAME tree: rc=1 and console **completely empty**. |
| AC-CTO-003b | PASS | same body, tree with a planted syntax error | step **rc=1**; build error console-visible via `BUILD FAILED` rows carrying `expected ')', found '{'`; **stderr = 0 bytes**; stream = 7 events, `jq -s` parses, `grep -vc '^{'` → 0; census re-parses it (rc=0). |
| AC-CTO-004 | PASS | `head -1 coverage.out; wc -l < coverage.out` after the converted step | `mode: atomic`, **47 lines** — identical to the `spec.md` §A.1 baseline. `-coverprofile`/`-covermode` unchanged in the step; Codecov step untouched. |
| AC-CTO-005a | PARTIAL — declaration half only | `yaml.safe_load` of both workflows | all four sites: `if: always()`, `actions/upload-artifact@v7`, `retention-days: 7`. Round-trip measured locally: 44,773 → 2,625 B gzip (**17.1×**), gunzip → census rc=0. **The "present in the dispatch run's artifact list" half is M4 and is NOT claimed here.** |
| AC-CTO-005b | DEBT — not asserted | — | needs a genuinely-red CI run; the ONE-run constraint forbids manufacturing one. |
| AC-CTO-006 | PASS | job `name:` sets extracted from `git show HEAD:<file>` vs working tree, both parsed with `yaml.safe_load` | **IDENTICAL** for both files. `test` `if: needs.detect.outputs.go_code == 'true'`; `test-skip-marker` `if: needs.detect.outputs.go_code != 'true'` — strict complements, unchanged. |
| AC-CTO-007 | NOT CLOSED — M4, lead-owned | — | no CI run was dispatched by run-phase. |
| AC-CTO-008 | PASS | step bodies extracted from both workflows via `yaml.safe_load`; `git diff HEAD --stat` on the excluded files | all four sites write `-json` to a file, capture rc with `\|\| rc=$?`, run the census, `exit $rc`. Eight artifact names, no collision. `git diff HEAD --stat -- .github/workflows/lsel-leak-guard.yaml .github/workflows/template-neutrality-check.yaml` → **empty** (untouched). |

### Mutant probe (AC-CTO-002)

Planted `select(.Action=="skip" and .Test != null)` in `test-census.sh`, exactly the
mutant `acceptance.md` names. `census-check.sh` returned **rc=1** with the diff:

```
-NOTHING RAN   example.com/censusfix/beta
```

— the mutant keeps the `SKIPPED TEST` row (so it passes a naive "names the
skipped test" check) and silently loses every zero-test-file package. Reverted;
check returns rc=0. A weaker RED was also observed first: with `test-census.sh`
absent the check exits rc=2.

### Eight artifact names (no collision)

`test-stream-ci-test-ubuntu-latest`, `test-stream-ci-race-ubuntu-latest`,
`test-stream-ci-integration-{ubuntu,macos,windows}-latest`,
`test-stream-release-verify-{ubuntu,macos,windows}-latest`.

### Finding — the SPEC's own stderr assumption is false

`spec.md` §F.2 and `plan.md` §B.2 state that build and toolchain errors "never
appear as JSON events" and that redirecting only stdout "keeps them
console-visible". **Measured false.** Under `-json`, `go test` writes compiler
diagnostics into the stream as `build-output` / `build-fail` events on
**stdout**, and stderr is **0 bytes**. A plain stdout redirect therefore
*swallows* the build error. AC-CTO-003b would have failed as written. The census
prints `BUILD FAILED` rows for exactly this reason, which is what makes the
criterion pass; the underlying prose in `spec.md`/`plan.md` is still wrong and is
manager-spec's to correct. Raw capture: `.moai/reports/t358/censusfix/`.

### Deviations from the plan, stated

1. **M2 and M3 landed in one commit.** Three of the four sites are in `ci.yml`;
   splitting would have left an intermediate state where the paths filter names
   a script only some steps use.
2. **The AC-CTO-003 / -003b step body substitutes the package selector.** The
   body was extracted verbatim except `./...` → `./internal/hook/mx/complexity/...`,
   because a full-suite local run is prohibited (load hazard, CLAUDE.local §4).
   The selector is not part of the shell semantics under test (`-e`, `|| rc=$?`,
   the redirect, the census, `exit $rc`), all of which are unchanged.
3. **Census row order and totals format are mine** (the plan left them open):
   failures print before the non-run census, because a red run is what the
   census exists for; totals are `packages=N passed=N skipped=N nothing-ran=N
   failed=N build-failed=N`, avoiding plural grammar and staying greppable.
4. **Two rows the plan did not name** — `BUILD FAILED` (see the finding above)
   and `FAILED PKG` (a package that failed with no failing test and no build
   failure, e.g. a `TestMain` exit).
5. **Three additions inside the SPEC's envelope**: `scripts/ci-census/**` in the
   `detect` paths filter (without it a census-only PR gets the skip-marker stub
   and the edit is never exercised — the stale-guard hazard of
   `verification-completeness.md` §1.3); `test-stream.json{,.gz}` gitignored;
   `if-no-files-found: warn` on the uploads so an `always()` upload after a step
   that never wrote the file does not add a spurious red.

### Named debt (carried, never reported as passes)

1. **AC-CTO-003 CI-level red path** — shell semantics verified locally against
   the exact converted body; the CI-level red path is not exercised pre-merge.
2. **AC-CTO-005b** — `if: always()` is a declaration of intent; the artifact was
   not observed on a FAILED run.

Both discharge on the first genuinely red CI run after this lands.

### Gaps — explicitly NOT observed in run-phase

- **No CI run of any kind.** M4 was not performed: nothing pushed, nothing
  dispatched, no run URL, no job id. AC-CTO-007 is open.
- **No windows or macOS execution.** The census was run only on darwin/arm64.
  `jq` on `windows-latest`/`macos-latest`, and `gzip`/`sed`/`sort` under the
  windows git-bash `shell: bash`, are **assumed available, not measured** — the
  four workflows that already use `jq` all run on ubuntu, so this repo carries
  no prior evidence for the windows leg. First real exposure is the 3-OS matrix.
- **No full-suite run.** Every measurement is on one package
  (`internal/hook/mx/complexity`) or on the fixture. The full-suite artifact size
  remains the `spec.md` §A.1 extrapolation.
- **`shellcheck` was not run** — not installed on this machine, and CI carries no
  shell lint.

### Residual risk

- A later editor "simplifies" `|| rc=$?` to `; rc=$?`; no green run contradicts
  them. Mitigation is comment-only, at all four sites.
- `test-stream.json` is written at the repo root *while* `go test ./...` runs. It
  is gitignored, so git-based checks stay clean, but a test that walks the repo
  root by filesystem rather than by git would see it. Not measured.
- The census shells out to `jq` once per failing test (capped at 50 bodies) and a
  handful of times for totals. On a large red run that is several passes over a
  multi-megabyte file — bounded, but not measured at full-suite scale.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-29
run_commit_sha: [13ec99545, d89998b13]
run_status: partial-milestones-m1-m3
milestones_delivered: [M1, M2, M3]
milestones_not_delivered: [M4, M4.1, M4.2]   # lead-owned; not attempted
ac_pass_count: 6        # 001, 002, 003, 003b, 004, 006, 008 -> 7 rows, of which
                        # 003 passes with named debt; counted strictly: 001, 002,
                        # 003b, 004, 006, 008 = 6 clean + 003 pass-with-debt
ac_partial_count: 1     # 005a (declaration half only)
ac_open_count: 1        # 007 (M4)
ac_debt_count: 2        # 005b, and 003's CI-level red path
ac_fail_count: 0
preserve_list_post_run_count: 0   # ci.yml:30-32 concurrency untouched; the four
                                  # -v-carrying steps untouched (diff empty)
l44_pre_commit_fetch: not-performed   # no push in run-phase scope
l44_post_push_fetch: not-performed
new_warnings_or_lints_introduced: none-observed  # no Go code changed; no shell
                                                 # lint exists in CI or locally
cross_platform_build:
  status: not-applicable   # no Go source changed in this card
  windows_census_execution: NOT-MEASURED   # named gap
total_run_phase_files: 31   # measured: git diff 1a635aea8 --name-only | wc -l
m1_to_mN_commit_strategy: "M1 separate (census + fixture check + draft->in-progress); M2+M3 combined (three of four sites share ci.yml)"
deliberate_breakages_reverted:
  - internal/hook/mx/complexity/t358_probe_test.go   # AC-CTO-003, removed; git status clean
  - internal/hook/mx/complexity/t358_broken.go       # AC-CTO-003b, removed; go build rc=0 after
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
