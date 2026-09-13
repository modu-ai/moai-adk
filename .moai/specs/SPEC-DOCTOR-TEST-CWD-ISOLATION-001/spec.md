---
id: SPEC-DOCTOR-TEST-CWD-ISOLATION-001
title: "Isolate full doctor command tests from the repository working directory"
version: "0.2.0"
status: draft
created: 2026-09-13
updated: 2026-09-13
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "doctor, tests, cwd-isolation, determinism, card-t675"
tier: S
---

# SPEC-DOCTOR-TEST-CWD-ISOLATION-001 — Doctor test CWD isolation

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-13 | manager-spec | Initial Tier S plan-phase draft for card t675; operator-directed CWD-isolation scope supersedes the earlier embedded-C1 hypothesis and Tier M Class B classification. |

## §1 Problem Statement

Nine existing `internal/cli` tests invoke the complete doctor diagnostic set while inheriting the
Go test process's repository working directory. `runGroupedChecksObserved` reads `os.Getwd()` and
passes that directory to repository-sensitive checks. In particular, the Agent Emit Embed check
walks upward from that directory, recognizes the moai-adk committed emission set, and judges the
repository's current `bin/moai`.

This makes the nine tests depend on ambient repository state that their assertions do not own. The
failure was observed on the card's baseline commit
`74d872aafbd90235e67163a5bc233f7c8a934491`: with `MOAI_EMBED_CHECK_BIN=/usr/bin/false`,
`TestDoctorCmd_Execution` exits 1 because the complete doctor run reports one failed check. The same
input makes all nine scoped tests fail. Without that injected failing target, the same nine-test
selection passed on the same commit, demonstrating that the test result changes with ambient doctor
inputs rather than with the behavior each test intends to verify.

The scope is limited to these nine tests across three existing test files:

| File | Scoped tests |
|------|--------------|
| `internal/cli/coverage_improvement_test.go` | `TestRunDoctor_WithExport`, `TestRunDoctor_WithFix`, `TestRunDoctor_Verbose`, `TestRunDoctor_AllFlags`, `TestRunDoctor_VerboseAndDetail`, `TestRunDoctor_ExportMode` |
| `internal/cli/doctor_test.go` | `TestDoctorCmd_Execution` |
| `internal/cli/integration_test.go` | `TestDoctorCmd_ExportFlag`, `TestDoctorCmd_VerboseExecution` |

## §2 Requirements (GEARS)

### REQ-DTC-001 — Isolated execution context

**When** any scoped test invokes an unfiltered complete doctor run, the test shall execute that run
from its own fresh temporary working directory outside the repository tree.

### REQ-DTC-002 — Independence from repository-only checks

**While** a scoped test is executing, repository-only doctor checks discovered through the package
source directory shall not determine the test result. A failing Agent Emit Embed judgment target in
the repository environment shall therefore be inapplicable to the scoped test's isolated run.

### REQ-DTC-003 — Preserve test intent and production behavior

The change shall preserve each scoped test's existing flags, output assertions, export assertions,
and error expectations. The change shall not modify non-test Go code or alter doctor check
registration, diagnostic output, applicability rules, or exit-status behavior.

### REQ-DTC-004 — Automatic isolation cleanup

**When** a scoped test completes or fails, its temporary directory and working-directory change
shall be restored by the Go test framework so that no later test inherits its fixture state.

## §3 Acceptance Criteria (inline, Tier S)

### AC-DTC-001 — Representative complete doctor run ignores the poisoned repository target

**Covers:** REQ-DTC-001, REQ-DTC-002, REQ-DTC-004

**Given** commit `74d872aafbd90235e67163a5bc233f7c8a934491`, the scoped
`TestDoctorCmd_Execution` test, and `MOAI_EMBED_CHECK_BIN=/usr/bin/false`, **When** the test runs,
**Then** it shall exit 0 and report the test as passed because its temporary working directory does
not make the repository-only Agent Emit Embed check applicable.

### AC-DTC-002 — All nine scoped tests remain non-empty, pass under the poisoned target, and stay test-only

**Covers:** REQ-DTC-001, REQ-DTC-002, REQ-DTC-003, REQ-DTC-004

**Given** the exact nine-test selection listed in §1 and
`MOAI_EMBED_CHECK_BIN=/usr/bin/false`, **When** the scoped selection runs after the isolation change,
**Then** the command shall exit 0, the selection listing shall contain all nine named tests, no
selected test shall fail, and the implementation diff under `internal/cli` shall contain only the
three test files listed in §1.

### §3.1 RED-now and green-path evidence

The acceptance criteria use the following baseline observations. Commands are single invocations;
stdout and exit codes were observed in this worktree on the pinned baseline.

#### E-RED-001 — AC-DTC-001

- **Tree SHA:** `74d872aafbd90235e67163a5bc233f7c8a934491`
- **Command:** `MOAI_EMBED_CHECK_BIN=/usr/bin/false go test ./internal/cli -count=1 -run '^TestDoctorCmd_Execution$'`
- **Exit code:** `1`
- **Verbatim stdout:**

```text
--- FAIL: TestDoctorCmd_Execution (18.09s)
    doctor_test.go:76: doctor command RunE error: doctor: 1 check(s) failed
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	19.040s
FAIL
```

- **Why RED:** the test inherits the package/repository CWD, so the complete doctor run reaches the
  repository-only Agent Emit Embed check and judges the injected failing executable.
- **Green path:** Milestone M1 gives the test a framework-managed temporary CWD; rerunning the same
  command exits 0.
- **Mutant probe:** changing only the environment variable, suppressing the doctor error, or
  weakening the assertion does not satisfy the criterion; the same poisoned command must pass with
  the original assertion intact.

#### E-RED-002 — AC-DTC-002

- **Tree SHA:** `74d872aafbd90235e67163a5bc233f7c8a934491`
- **Command:** `MOAI_EMBED_CHECK_BIN=/usr/bin/false go test ./internal/cli -count=1 -run '^(TestRunDoctor_(WithExport|WithFix|Verbose|AllFlags|VerboseAndDetail|ExportMode)|TestDoctorCmd_(Execution|ExportFlag|VerboseExecution))$'`
- **Exit code:** `1`
- **Observed terminal verdict:** all nine named tests emitted `--- FAIL`; the package ended with
  `FAIL\nFAIL\tgithub.com/modu-ai/moai-adk/internal/cli\t169.247s\nFAIL`.
- **Why RED:** every selected test inherits the repository CWD and reaches the same injected
  repository-only failure.
- **Green path:** Milestone M1 isolates all nine tests; Milestone M2 reruns the exact command and
  separately lists the selector to prove the swept set contains all nine tests.
- **Mutant probe:** isolating only the representative test leaves the other eight failures visible;
  renaming or deleting tests is rejected by the exact nine-name listing; changing production doctor
  logic is rejected by the three-test-file diff boundary.

## §4 Constraints

- The configured run methodology is TDD. The observed poisoned-target failure is the RED baseline;
  the isolation-only test edits provide GREEN.
- Scoped tests shall remain serial because process working directory and `doctorCmd` flags are
  process/package-global state.
- Local verification shall remain scoped to `internal/cli`; the full local repository suite is not
  run. CI on the integrated `develop` head owns the full-suite verdict.
- Temporary directories shall be framework-managed and shall not be created under the repository.

## §5 Exclusions

### Out of Scope — doctor production behavior

- No production doctor check, root-discovery algorithm, Agent Emit Embed applicability rule,
  diagnostic text, command flag, or exit-status behavior is changed.

### Out of Scope — unrelated doctor tests

- No doctor test outside the exact nine-test list in §1 is modified or reclassified.

### Out of Scope — embedded artifact repair

- Rebuilding `bin/moai`, regenerating agent-emit artifacts, or changing committed/template artifact
  bytes is not part of this SPEC. Those states are deliberate hostile inputs for the isolation
  check, not defects this card repairs.

### Out of Scope — broad test refactoring

- No shared doctor-test harness, package-wide CWD abstraction, parallelization change, or cleanup of
  adjacent tests is introduced.

## §6 Cross-References

- Doctor CWD acquisition: `internal/cli/doctor.go` (`runGroupedChecksObserved`)
- Repository-sensitive applicability walk: `internal/cli/doctor_agentemit_embed.go`
- Scoped tests: `internal/cli/coverage_improvement_test.go`, `internal/cli/doctor_test.go`,
  `internal/cli/integration_test.go`
- Prior observed failure set: `.moai/reports/t662/verdict.md` and
  `.moai/reports/t662/preattrib-develop-55b757ff2.log`
- Superseded gateway failure record: `.moai/reports/t675/gateway-502-20260913.md`
- Card: t675; worktree `.claude/worktrees/t675`; branch `WT-doctor-red`; baseline
  `74d872aafbd90235e67163a5bc233f7c8a934491`
