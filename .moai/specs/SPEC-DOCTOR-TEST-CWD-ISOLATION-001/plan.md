# Plan — SPEC-DOCTOR-TEST-CWD-ISOLATION-001

Card t675 · worktree `.claude/worktrees/t675` · branch `WT-doctor-red` · RED baseline
`dd235a66b` (full SHA recorded in `spec.md §3.1`) · Tier S · plan version 0.3.1 (2026-09-18)

## §A Context

Nine `internal/cli` tests run the complete doctor check set without first leaving the Go package's
repository working directory. `runGroupedChecksObserved` reads `os.Getwd()` and the Agent Emit Embed
check walks upward from that value to the repository's committed emission set. The tests therefore
judge ambient repository state that is unrelated to their assertions.

The dependency is mechanically observable on the RED baseline `dd235a66b`: with
`MOAI_EMBED_CHECK_BIN=/usr/bin/false` the representative test and the nine-test selection both exit 1
(`spec.md §3.1` E-RED-001 / E-RED-002, raw output under `.moai/reports/t675/red/`). Without the
override the nine tests pass, because this tree has no `bin/moai` and the Agent Emit Embed check
skips (`progress.md §E.1a`). The production doctor behavior is correct: in an applicable repository
it must report an extraction failure. The defect is test fixture isolation.

The operator fixed this as a Tier S, test-only scope. That decision supersedes the card's earlier
embedded-C1 hypothesis and Tier M Class B classification.

## §B Known Issues

- Process CWD is global state. The scoped tests must not use `t.Parallel()`, and restoration must hold
  even on `Fatal` or panic paths.
- `doctorCmd` is a package-global Cobra command. Existing flag reset behavior in
  `integration_test.go` must remain intact; this card does not redesign command construction.
- `TestRunDoctor_AllFlags` routes live progress through the command's default error writer, so a
  failing full check produces a larger terminal trace. Evidence is persisted to files and cited.
- A normal green run alone does not prove isolation: this tree has no `bin/moai`, so the Agent Emit
  Embed check skips. The poisoned `MOAI_EMBED_CHECK_BIN=/usr/bin/false` input is required to show that
  repository applicability no longer reaches the scoped tests.
- `internal/cli/coverage_improvement_test.go` is large. Changes are restricted to the six named test
  bodies; no adjacent coverage cleanup belongs in this card.
- The diff-based criteria (AC-DTC-003..005) use `develop...HEAD`, whose left end is the merge-base
  with local `develop`. They are pre-merge evaluations only: once the card merges into `develop` the
  range is empty, so the "exactly three paths" (AC-DTC-003) and "exactly nine added lines/statements"
  (AC-DTC-004 condition (2), AC-DTC-005) conditions are unmet and the checks report failure — not a
  vacuous pass. Post-merge evidence is the merge tree's identity with the card branch tree instead.

## §C Pre-flight

1. Confirm location: `git branch --show-current` prints `WT-doctor-red`, and
   `git rev-parse --show-toplevel` prints the t675 worktree root.
2. Confirm lineage: `git merge-base --is-ancestor dd235a66b HEAD` exits 0.
3. Re-list the exact nine tests with `go test ./internal/cli -list
   '^(TestRunDoctor_(WithExport|WithFix|Verbose|AllFlags|VerboseAndDetail|ExportMode)|TestDoctorCmd_(Execution|ExportFlag|VerboseExecution))$'`
   and require all nine names.
4. Reproduce RED with the exact E-RED-002 command from `spec.md §3.1` on the pre-change tree; require
   exit 1 and nine `--- FAIL` lines before any test edit.
5. Confirm the Go-file scope is clean with `git status --short --
   internal/cli/coverage_improvement_test.go internal/cli/doctor_test.go internal/cli/integration_test.go`.

## §D Constraints

- C1 — TEST-ONLY: modify only the three scoped `*_test.go` files. No production Go file, generated
  artifact, template, binary, or unrelated test changes.
- C2 — EXACT TEST SET: isolate exactly the nine tests listed in `spec.md §1`; no package-wide sweep.
- C3 — IMPLEMENTATION APPROACH (moved from the requirement layer in v0.3.0): each scoped test begins
  with the Go testing framework's managed working-directory change into a framework-managed temporary
  directory, as one added statement: `t.Chdir(t.TempDir())`. `t.TempDir()` removes the directory
  through test cleanup; `t.Chdir` restores the prior working directory through test cleanup, which
  runs after `Fatal`/`FailNow` as well as after a pass. The module targets `go 1.26.8`, which provides
  `testing.T.Chdir`. No helper, no bare `os.Chdir`, no deferred best-effort restoration.
- C4 — PRESERVE ASSERTIONS: no existing line in the three files is removed or edited; the isolation
  statement is purely additive. AC-DTC-004 checks this mechanically.
- C5 — SERIAL EXECUTION: do not add `t.Parallel()` (the framework also rejects `t.Chdir` in parallel
  tests).
- C6 — SCOPED VERIFICATION: do not run `go test ./...` or `./internal/cli/...` locally. Run only the
  selected tests in `./internal/cli/`; the integrated `origin/develop` CI run owns the full-suite
  verdict.
- C7 — EVIDENCE INTEGRITY: capture command, raw output file, exit code, and HEAD SHA for every GREEN
  check. A zero-test selection is not a pass.
- C8 — CARD PROTOCOL: leave commit, push, integration, and final verdict to the lane/lead
  instructions; every card commit carries t675 in its message.

## §E Self-Verification

Run-phase evidence belongs in `progress.md §E.2` and `§E.3`, owned by manager-develop. Each result
records the exact command, a persisted raw output file, exit code, and HEAD SHA.

1. AC-DTC-001: rerun the E-RED-001 command on the implementation HEAD; exit 0.
2. AC-DTC-002: rerun the E-RED-002 command on the implementation HEAD; exit 0 with nine `--- PASS`
   lines and zero `--- FAIL` lines; plus the `-list` selector from §C step 3.
3. Normal-environment regression: the same nine-test selection without the override; exit 0.
4. AC-DTC-003: `git diff --name-only develop...HEAD -- '*.go'`.
5. AC-DTC-004 and AC-DTC-005: `git diff -U0 --no-color develop...HEAD -- internal/cli/coverage_improvement_test.go internal/cli/doctor_test.go internal/cli/integration_test.go`.
6. Formatting/static checks: `gofmt -l` on the three files prints nothing; `go vet ./internal/cli/`
   exits 0 or any pre-existing baseline is reported separately.

## §F Milestones

Milestones are ordered by change likelihood: first settle the per-test isolation placement across
three files, then perform mechanical verification and evidence capture. The RED ledger in
`spec.md §3.1` is already measured on the pre-change tree and is not a milestone.

### M1 — Isolate the nine complete-doctor tests (Priority: High)

- In `internal/cli/coverage_improvement_test.go`, add the §D C3 isolation statement as the first
  statement of the six scoped `TestRunDoctor_*` tests.
- In `internal/cli/doctor_test.go`, add it to `TestDoctorCmd_Execution`.
- In `internal/cli/integration_test.go`, add it to `TestDoctorCmd_ExportFlag` and
  `TestDoctorCmd_VerboseExecution`, leaving their flag-reset `defer` blocks untouched.
- Keep the change local to each test body; add no helper, production seam, or parallel execution.

Exit: AC-DTC-001, AC-DTC-002

### M2 — Prove scope, assertion preservation, and restoration form (Priority: High)

- Run the three diff-based checks from §E steps 4-5 and persist their raw output.
- Run the normal-environment regression and the scoped static checks from §E steps 3 and 6.
- Persist attributable GREEN evidence in the run-owned progress sections; hand off to sync only after
  all five inline ACs pass.

Exit: AC-DTC-003, AC-DTC-004, AC-DTC-005

## §G Anti-Patterns

- Do not rebuild `bin/moai` or delete one to make the tests green; that hides the ambient dependency.
- Do not change `checkAgentEmitEmbed`, `findEmbedCheckRoot`, `runGroupedChecksObserved`, or doctor exit
  handling. Their failure under the poisoned applicable repository is expected production behavior.
- Do not unset or ignore `MOAI_EMBED_CHECK_BIN` in production or test code, and do not weaken, delete,
  or reorder existing assertions.
- Do not isolate unrelated doctor tests or introduce a package-wide test harness.
- Do not use bare `os.Chdir` with deferred best-effort restoration.
- Do not add `t.Parallel()` around process-CWD or package-global Cobra state.

## §H Cross-References

- Requirements, inline ACs, RED ledger: `.moai/specs/SPEC-DOCTOR-TEST-CWD-ISOLATION-001/spec.md`
- Progress and re-measurement: `.moai/specs/SPEC-DOCTOR-TEST-CWD-ISOLATION-001/progress.md`
- Raw RED and mutant-probe evidence: `.moai/reports/t675/red/`
- Prior audit: `.moai/reports/t675/plan-audit.md` (FAIL 0.67, D1-D5)
- Doctor CWD source: `internal/cli/doctor.go`
- Repository-only check: `internal/cli/doctor_agentemit_embed.go`
- Scoped tests: `internal/cli/coverage_improvement_test.go`, `internal/cli/doctor_test.go`,
  `internal/cli/integration_test.go`
- Gateway record (preserve): `.moai/reports/t675/gateway-502-20260913.md`
