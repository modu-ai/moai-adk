# Plan — SPEC-DOCTOR-TEST-CWD-ISOLATION-001

Card t675 · worktree `.claude/worktrees/t675` · branch `WT-doctor-red` · base
`74d872aafbd90235e67163a5bc233f7c8a934491` · Tier S · operator-directed CWD-isolation scope

## §A Context

Nine `internal/cli` tests run the complete doctor check set without first leaving the Go package's
repository working directory. `runGroupedChecksObserved` reads `os.Getwd()` and the Agent Emit Embed
check walks upward from that value to the repository's committed emission set. The tests therefore
judge ambient repository state that is unrelated to their assertions.

The dependency is mechanically observable on the pinned baseline. The nine-test selection passes
normally (`ok github.com/modu-ai/moai-adk/internal/cli 192.291s`), but the same selection with
`MOAI_EMBED_CHECK_BIN=/usr/bin/false` fails all nine tests and exits 1. The production doctor behavior
is correct: in an applicable repository it must report an extraction failure. The defect is in test
fixture isolation, not in the check, its applicability walk, or its exit status.

The operator fixed this as a Tier S, test-only scope. That decision supersedes the card's earlier
embedded-C1 hypothesis and Tier M Class B classification.

## §B Known Issues

- Process CWD is global state. The scoped tests must not use `t.Parallel()`, and isolation must be
  restored through the test framework even on `Fatal` or panic paths.
- `doctorCmd` is a package-global Cobra command. Existing flag reset behavior in
  `integration_test.go` must remain intact; this card does not redesign command construction.
- `TestRunDoctor_AllFlags` routes live progress through the command's default error writer, so a
  failing full check can produce a larger terminal trace. Verification stays on the exact scoped
  selector and uses bounded output in reports.
- A normal green run alone does not prove isolation: the current repository binary happens to pass
  the Agent Emit Embed check. The poisoned `MOAI_EMBED_CHECK_BIN=/usr/bin/false` input is required to
  demonstrate that repository applicability no longer reaches the scoped tests.
- `internal/cli/coverage_improvement_test.go` is large. Changes are restricted to the six named test
  bodies; no adjacent coverage cleanup belongs in this card.

## §C Pre-flight

1. Confirm lineage and location: `git merge-base --is-ancestor 74d872aafbd90235e67163a5bc233f7c8a934491 HEAD`
   exits 0, `git branch --show-current` prints `WT-doctor-red`, and
   `git rev-parse --show-toplevel` prints the t675 worktree root.
2. Re-list the exact nine tests with `go test ./internal/cli -list
   '^(TestRunDoctor_(WithExport|WithFix|Verbose|AllFlags|VerboseAndDetail|ExportMode)|TestDoctorCmd_(Execution|ExportFlag|VerboseExecution))$'`
   and require all nine names plus a non-empty package result.
3. Reproduce RED with the exact E-RED-002 command from `spec.md §3.1`; require exit 1 and all nine
   named failures before any test edit. Do not reinterpret a normal green run as the RED baseline.
4. Confirm the planned Go-file scope is clean with `git status --short --
   internal/cli/coverage_improvement_test.go internal/cli/doctor_test.go internal/cli/integration_test.go`.

## §D Constraints

- C1 — TEST-ONLY: modify only the three scoped `*_test.go` files during run implementation. No
  production Go file, generated artifact, template, binary, or unrelated test changes.
- C2 — EXACT TEST SET: isolate exactly the nine tests listed in `spec.md §1`; no package-wide sweep.
- C3 — SIMPLE FIX: use the repository's direct framework-managed temporary-CWD idiom in each scoped
  test. Do not add a helper or abstraction unless the direct form cannot satisfy an observed case.
- C4 — PRESERVE ASSERTIONS: existing flags, output/export assertions, error handling, and flag-reset
  cleanup remain byte-for-byte or semantically unchanged apart from the isolation setup.
- C5 — SERIAL EXECUTION: do not add `t.Parallel()` to tests that mutate CWD or package-global command
  flags.
- C6 — SCOPED VERIFICATION: do not run `go test ./...` locally. Run only the selected tests and
  affected-package static checks; the integrated `origin/develop` CI run owns the full-suite verdict.
- C7 — EVIDENCE INTEGRITY: capture command, verbatim output, exit code, and tree SHA for RED and GREEN.
  A zero-test selection is not a pass.
- C8 — CARD PROTOCOL: leave commit, push, integration, and final verdict to the lane/lead instructions;
  every later card commit must carry t675 in its message.

## §E Self-Verification

Run-phase evidence belongs in `progress.md §E.2` and `§E.3`, owned by manager-develop. Each result
records the exact command, verbatim bounded output (or a persistent evidence path), exit code, and
HEAD SHA.

Required checks:

1. AC-DTC-001: rerun E-RED-001 unchanged; exit 0.
2. AC-DTC-002 behavior: rerun E-RED-002 unchanged; exit 0.
3. AC-DTC-002 non-empty sweep: use the `go test -list` selector from §C and observe exactly the nine
   expected names.
4. Normal-environment regression: run the same nine-test selection without the override; exit 0.
5. Scope: diff from the current `develop` merge-base to the implementation HEAD under `internal/cli`
   and observe only the three scoped test files; inspect the diff to confirm no assertion weakening.
6. Formatting/static checks: `gofmt -d` on the three files produces no output;
   `go vet ./internal/cli/...` and `golangci-lint run internal/cli/...` exit 0 or report any measured
   pre-existing baseline separately.

## §F Milestones

Milestones are ordered by change likelihood: first settle the per-test isolation placement across
three files, then perform mechanical verification and evidence capture.

### M1 — Isolate the nine complete-doctor tests (Priority: High)

- In `internal/cli/coverage_improvement_test.go`, give the six scoped `TestRunDoctor_*` tests a fresh,
  framework-managed temporary CWD before they invoke the complete doctor run.
- In `internal/cli/doctor_test.go`, apply the same isolation to `TestDoctorCmd_Execution`.
- In `internal/cli/integration_test.go`, apply the same isolation to
  `TestDoctorCmd_ExportFlag` and `TestDoctorCmd_VerboseExecution` while preserving their existing
  flag-reset cleanup.
- Keep the change local to each test body; add no helper, production seam, or parallel execution.
- Exit: the exact poisoned command in E-RED-002 changes from nine failures/exit 1 to exit 0.

### M2 — Prove the selection, preservation boundary, and package quality (Priority: High)

- List the selector and require all nine test names so an empty or partially renamed sweep cannot
  pass silently.
- Run the nine tests both with and without the poisoned embed-check target. Preserve each existing
  assertion and inspect the diff for assertion weakening.
- Confirm the implementation diff under `internal/cli` names exactly the three scoped test files and
  no production Go file.
- Run scoped formatting, vet, and lint checks from §E. Do not run the repository-wide local suite.
- Persist attributable GREEN evidence in the run-owned progress sections and hand off to sync only
  after both inline ACs pass.

## §G Anti-Patterns

- Do not rebuild `bin/moai` to make the tests green; that hides the ambient dependency and leaves the
  tests nondeterministic.
- Do not change `checkAgentEmitEmbed`, `findEmbedCheckRoot`, `runGroupedChecksObserved`, or doctor exit
  handling. Their failure under the poisoned applicable repository is expected production behavior.
- Do not unset or ignore `MOAI_EMBED_CHECK_BIN` inside production code or weaken test error assertions.
- Do not isolate unrelated doctor tests or introduce a package-wide test harness.
- Do not use manual `os.Chdir` plus deferred best-effort restoration when the test framework owns the
  lifecycle directly.
- Do not add `t.Parallel()` around process-CWD or package-global Cobra state.

## §H Cross-References

- Requirements and inline ACs: `.moai/specs/SPEC-DOCTOR-TEST-CWD-ISOLATION-001/spec.md`
- Progress and prior gateway evidence pointer:
  `.moai/specs/SPEC-DOCTOR-TEST-CWD-ISOLATION-001/progress.md`
- Doctor CWD source: `internal/cli/doctor.go`
- Repository-only check: `internal/cli/doctor_agentemit_embed.go`
- Scoped tests: `internal/cli/coverage_improvement_test.go`, `internal/cli/doctor_test.go`,
  `internal/cli/integration_test.go`
- Prior measured failure: `.moai/reports/t662/preattrib-develop-55b757ff2.log`
- Gateway record (preserve): `.moai/reports/t675/gateway-502-20260913.md`
