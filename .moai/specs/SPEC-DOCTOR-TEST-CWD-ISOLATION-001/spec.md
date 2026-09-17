---
id: SPEC-DOCTOR-TEST-CWD-ISOLATION-001
title: "Isolate full doctor command tests from the repository working directory"
version: "0.3.1"
status: in-progress
created: 2026-09-13
updated: 2026-09-18
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
| 0.2.0 | 2026-09-13 | manager-spec | Committed draft audited by plan-audit iteration 1 (FAIL 0.67): four requirements, two inline ACs, RED evidence pinned to baseline `74d872aafbd90235e67163a5bc233f7c8a934491` (recovered from commit `49bf74a82`). |
| 0.3.0 | 2026-09-18 | manager-spec | Plan-audit FAIL 0.67 remediation (D1-D5 + MP-9 GAP). One `shall` per requirement (REQ-DTC-002/003 split, renumbered to six); cleanup rewritten as two observable outcomes with the framework mechanism moved to `plan.md §D C3`; RED ledger re-measured on `dd235a66b` with raw output persisted and quoted verbatim; ACs restated against the implementation descendant; assertion-deletion and bare-`os.Chdir` mutant criteria added; plan milestones bind ACs through `Exit:` lines. Scope unchanged. |
| 0.3.1 | 2026-09-18 | manager-spec | Wording-only corrections from plan-audit iteration 2 optional defects (O1-O3); no requirement or AC semantics change. O1: added the missing 0.2.0 row. O2: post-merge behavior of AC-DTC-003..005 restated as failure on the empty range, not a vacuous pass (`plan.md §B`, §3 preamble). O3: the no-injection pass citation in §1 re-measured with the exact anchored nine-test selector (previous selector `'TestRunDoctor_|TestDoctorCmd_'` matched 27 tests). |

## §1 Problem Statement

Nine existing `internal/cli` tests invoke the complete doctor diagnostic set while inheriting the
Go test process's repository working directory. `runGroupedChecksObserved` reads `os.Getwd()` and
passes that directory to repository-sensitive checks. In particular, the Agent Emit Embed check
walks upward from that directory, recognizes the moai-adk committed emission set, and judges the
repository's `bin/moai` or the executable named by `MOAI_EMBED_CHECK_BIN`.

The nine tests therefore depend on ambient repository state that their assertions do not own. On
the RED baseline `dd235a66b` (§3.1), injecting `MOAI_EMBED_CHECK_BIN=/usr/bin/false` makes the
representative test and all nine scoped tests fail with `doctor: 1 check(s) failed`. Without that
injection the same nine tests pass only because no `bin/moai` exists, so the check skips
(`progress.md §E.1a`): on HEAD `8831e42972d779bb8b2b0efee911e33d9943b3d0` (no Go change relative
to `dd235a66b`), the exact anchored selection
`go test -count=1 -v -run '^(TestRunDoctor_WithExport|TestRunDoctor_WithFix|TestRunDoctor_Verbose|TestRunDoctor_AllFlags|TestRunDoctor_VerboseAndDetail|TestRunDoctor_ExportMode|TestDoctorCmd_Execution|TestDoctorCmd_ExportFlag|TestDoctorCmd_VerboseExecution)$' ./internal/cli/`
exited 0 with exactly nine `--- PASS:` lines naming the nine tests of §1, zero `--- FAIL` lines, and
`ok  	github.com/modu-ai/moai-adk/internal/cli	106.622s` (raw output
`.moai/reports/t675/red/natural-9-anchored.txt`, sha256
`97c975d464b2780799f9bc3ee1b444508b8536b21f6b2f9fce9ed4cb1ed0a8c0`). The test verdict changes with
ambient doctor inputs rather than with the behavior each test intends to verify.

The scope is limited to these nine tests across three existing test files:

| File | Scoped tests |
|------|--------------|
| `internal/cli/coverage_improvement_test.go` | `TestRunDoctor_WithExport`, `TestRunDoctor_WithFix`, `TestRunDoctor_Verbose`, `TestRunDoctor_AllFlags`, `TestRunDoctor_VerboseAndDetail`, `TestRunDoctor_ExportMode` |
| `internal/cli/doctor_test.go` | `TestDoctorCmd_Execution` |
| `internal/cli/integration_test.go` | `TestDoctorCmd_ExportFlag`, `TestDoctorCmd_VerboseExecution` |

## §2 Requirements (GEARS)

### REQ-DTC-001 — Isolated execution context

**When** a scoped test invokes an unfiltered complete doctor run, the test shall execute that run
from a fresh temporary working directory outside the repository tree.

### REQ-DTC-002 — Independence from repository-only checks

**While** a scoped test is executing, the test result shall be independent of any repository-only
doctor check whose judgment target in the environment fails.

Rationale (non-normative): the Agent Emit Embed check is applicable only when the working directory
lies inside the moai-adk repository, so an isolated working directory makes it inapplicable.

### REQ-DTC-003 — Preserve test intent

The change to the scoped tests shall keep each scoped test's existing flags, output assertions,
export assertions, and error expectations.

### REQ-DTC-004 — Production behavior unchanged

The change shall not modify any non-test Go file.

### REQ-DTC-005 — Temporary directory removed

**When** a scoped test finishes, whether it passed or failed, the temporary working directory
created for that test shall no longer exist.

### REQ-DTC-006 — Original working directory restored

**When** a scoped test finishes, whether it passed or failed, the process working directory shall
equal the working directory that was in effect when that test started.

## §3 Acceptance Criteria (inline, Tier S)

All criteria are evaluated on the **implementation descendant**: a commit on `WT-doctor-red` that
descends from the RED baseline `dd235a66b1145922565841d33acafef0d1ded6a8` and contains the isolation
change. The RED baseline itself is recorded only in the §3.1 ledger and is expected to fail
AC-DTC-001, AC-DTC-002, and AC-DTC-005. AC-DTC-003..005 read the range `develop...HEAD` (left end =
merge-base with local `develop`) and are valid only while the card is unmerged; after merge the range
is empty and their exact-count conditions report failure rather than a vacuous pass.

### AC-DTC-001 — Representative complete doctor run ignores the poisoned repository target

**Covers:** REQ-DTC-001, REQ-DTC-002

**Given** the implementation descendant and `MOAI_EMBED_CHECK_BIN=/usr/bin/false`, **When**
`go test -count=1 -v -run '^TestDoctorCmd_Execution$' ./internal/cli/` runs, **Then** it exits 0,
the output contains `--- PASS: TestDoctorCmd_Execution`, and the output contains no `--- FAIL` line.

### AC-DTC-002 — All nine scoped tests pass under the poisoned target and the sweep is non-empty

**Covers:** REQ-DTC-001, REQ-DTC-002, REQ-DTC-003

**Given** the implementation descendant and `MOAI_EMBED_CHECK_BIN=/usr/bin/false`, **When** the
E-RED-002 command runs unchanged, **Then** it exits 0, the output contains exactly nine
`--- PASS:` lines naming the nine tests of §1, and the output contains zero `--- FAIL` lines; and
**When** `go test ./internal/cli -list '^(TestRunDoctor_(WithExport|WithFix|Verbose|AllFlags|VerboseAndDetail|ExportMode)|TestDoctorCmd_(Execution|ExportFlag|VerboseExecution))$'`
runs, **Then** it exits 0 and lists all nine names.

### AC-DTC-003 — Test-only change boundary

**Covers:** REQ-DTC-004

**Given** the implementation descendant, **When** `git diff --name-only develop...HEAD -- '*.go'`
runs, **Then** it exits 0 and its output is exactly the three file paths of §1, one per line, with
no other path.

### AC-DTC-004 — Existing assertions survive (rejects an assertion-deletion mutant)

**Covers:** REQ-DTC-003

**Given** the implementation descendant, **When**
`git diff -U0 --no-color develop...HEAD -- internal/cli/coverage_improvement_test.go internal/cli/doctor_test.go internal/cli/integration_test.go`
runs, **Then** it exits 0 and its output satisfies both conditions:

- (1) zero removed content lines (lines beginning with `-` other than the `--- a/` file headers);
- (2) exactly nine non-blank added content lines (lines beginning with `+` other than the `+++ b/`
  file headers), each consisting of whitespace followed by the isolation statement
  `t.Chdir(t.TempDir())` and nothing else.

A mutant that deletes or edits any existing assertion produces a removed content line and fails
condition (1); a mutant that inserts an early `return`, `t.Skip`, or any other statement fails
condition (2). Probe: `.moai/reports/t675/red/mutant-probe.txt` (variant `mutA`).

### AC-DTC-005 — Working-directory change is restored by construction (rejects a bare `os.Chdir` mutant)

**Covers:** REQ-DTC-001, REQ-DTC-005, REQ-DTC-006

**Given** the implementation descendant, **When** the AC-DTC-004 diff command runs, **Then** zero
added content lines contain `os.Chdir(`, and the hunk headers (`@@ … @@ func …`) of the nine added
`t.Chdir(t.TempDir())` statements name the nine tests of §1, each exactly once.

A mutant that changes directory with bare `os.Chdir` (with or without a deferred restore) produces
an added `os.Chdir(` line and fails. On the RED baseline the command yields no added line, so the
nine-statement condition is unmet. Probe: `.moai/reports/t675/red/mutant-probe.txt` (variant `mutB`).

### §3.1 RED ledger (baseline `dd235a66b1145922565841d33acafef0d1ded6a8`)

Measured 2026-09-18 on a clean worktree at HEAD `dd235a66b1145922565841d33acafef0d1ded6a8`, branch
`WT-doctor-red`, no `bin/moai` present. Each command is a single invocation; the session captured
its combined stdout/stderr to the listed file and its exit status to the sibling `.exit` file. The
blocks below are the file contents verbatim.

#### E-RED-001 — for AC-DTC-001

- **Command:** `MOAI_EMBED_CHECK_BIN=/usr/bin/false go test -count=1 -v -run '^TestDoctorCmd_Execution$' ./internal/cli/`
- **Exit code:** `1` (`.moai/reports/t675/red/e-red-001.exit`)
- **Raw output:** `.moai/reports/t675/red/e-red-001.txt` (sha256 `2ff331b759f1764a778f22559dbd43e15678e3ab3e923f18b0354f56807d3739`)

```text
=== RUN   TestDoctorCmd_Execution
    doctor_test.go:76: doctor command RunE error: doctor: 1 check(s) failed
--- FAIL: TestDoctorCmd_Execution (12.47s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	13.272s
FAIL
```

- **Why RED:** the test inherits the package directory inside the repository, so the complete
  doctor run reaches the repository-only Agent Emit Embed check and judges the injected failing
  executable.
- **Green path:** M1 adds the isolated working directory; the same command then exits 0.

#### E-RED-002 — for AC-DTC-002

- **Command:** `MOAI_EMBED_CHECK_BIN=/usr/bin/false go test -count=1 -v -run '^(TestRunDoctor_(WithExport|WithFix|Verbose|AllFlags|VerboseAndDetail|ExportMode)|TestDoctorCmd_(Execution|ExportFlag|VerboseExecution))$' ./internal/cli/`
- **Exit code:** `1` (`.moai/reports/t675/red/e-red-002.exit`)
- **Raw output:** `.moai/reports/t675/red/e-red-002.txt` (94 lines, sha256 `7d1e24166573f4caec68a378ee7b4e9e8789e93f69c9ac92ece8ccb25a361b7c`)

```text
=== RUN   TestRunDoctor_WithExport
    coverage_improvement_test.go:715: runDoctor error: doctor: 1 check(s) failed
--- FAIL: TestRunDoctor_WithExport (9.24s)
=== RUN   TestRunDoctor_WithFix
    coverage_improvement_test.go:737: runDoctor error: doctor: 1 check(s) failed
--- FAIL: TestRunDoctor_WithFix (14.20s)
=== RUN   TestRunDoctor_Verbose
    coverage_improvement_test.go:777: runDoctor error: doctor: 1 check(s) failed
--- FAIL: TestRunDoctor_Verbose (28.56s)
=== RUN   TestRunDoctor_AllFlags
  ○ Go Runtime
  ✓ Go Runtime
  ○ Git
  ✓ Git
  ○ Claude Code
  ✓ Claude Code
  ○ GitHub CLI
  ✓ GitHub CLI
  ○ ast-grep CLI
  ✓ ast-grep CLI
  ○ Shared Flag Slot
  ✓ Shared Flag Slot
  ○ MoAI Config
  ✓ MoAI Config
  ○ Claude Config
  ✓ Claude Config
  ○ MoAI Version
  ✓ MoAI Version
  ○ Binary Freshness
  ✓ Binary Freshness
  ○ MCP Scope Duplicates
  ✓ MCP Scope Duplicates
  ○ MCP Server Version
  ✓ MCP Server Version
  ○ Agent Emit Embed
  ✗ Error: could not extract embedded artifacts from /usr/bin/false: false init: exit status 1 ()
  ○ Constitution Registry
  ✓ Constitution Registry
  ○ Harness 5-Layer
  ✓ Harness 5-Layer
  ○ Migration
  ✓ Migration
  ○ Plugin Deployment
  ✓ Plugin Deployment
  ○ Home Disk Usage
  ✓ Home Disk Usage
  ○ Hooks Config
  ✓ Hooks Config
  ○ Hook Wiring
  ✓ Hook Wiring
  ○ Hook Delivery
  ✓ Hook Delivery
  ○ Hook opt-in:
  ✓ Hook opt-in:
  ○ Slash Commands
  ✓ Slash Commands
  ○ Skills Allowlist
  ✓ Skills Allowlist
  ○ MX Tag Config
  ✓ MX Tag Config
  ○ Worktree State
  ✓ Worktree State
  ○ Worktree Base Branch
  ✓ Worktree Base Branch
  ○ Git Strategy Workflow
  ✓ Git Strategy Workflow
  ○ BODP Config
  ✓ BODP Config
  ○ Telemetry Config
  ✓ Telemetry Config
  ○ Glamour Cache
  ✓ Glamour Cache
  ○ Codex Wiring
  ✓ Codex Wiring
    coverage_improvement_test.go:4930: unexpected error: doctor: 1 check(s) failed
--- FAIL: TestRunDoctor_AllFlags (20.15s)
=== RUN   TestRunDoctor_VerboseAndDetail
    coverage_improvement_test.go:5754: runDoctor error: doctor: 1 check(s) failed
--- FAIL: TestRunDoctor_VerboseAndDetail (14.08s)
=== RUN   TestRunDoctor_ExportMode
    coverage_improvement_test.go:5804: runDoctor error: doctor: 1 check(s) failed
--- FAIL: TestRunDoctor_ExportMode (13.49s)
=== RUN   TestDoctorCmd_Execution
    doctor_test.go:76: doctor command RunE error: doctor: 1 check(s) failed
--- FAIL: TestDoctorCmd_Execution (16.10s)
=== RUN   TestDoctorCmd_ExportFlag
    integration_test.go:176: doctor --export error: doctor: 1 check(s) failed
--- FAIL: TestDoctorCmd_ExportFlag (11.99s)
=== RUN   TestDoctorCmd_VerboseExecution
    integration_test.go:202: doctor --verbose error: doctor: 1 check(s) failed
--- FAIL: TestDoctorCmd_VerboseExecution (12.50s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	140.953s
FAIL
```

- **Why RED:** every selected test inherits the repository working directory and reaches the same
  injected repository-only failure; the nine `--- FAIL` lines name exactly the nine tests of §1.
- **Green path:** M1 isolates all nine tests; the unchanged command then exits 0 with nine
  `--- PASS` lines.

#### E-RED-003 — for AC-DTC-003, AC-DTC-004, AC-DTC-005

- **Measurement context:** local `develop` at `9aee76589935b6a5d255ff0dac6eae84f41868f4`; merge-base
  with HEAD `f67d2193f22cbc80921e42539c3462a4100f31e9`.
- **Command (AC-DTC-003):** `git diff --name-only develop...HEAD -- '*.go'` — **Exit code:** `0` —
  **Verbatim stdout:** empty (zero bytes).
- **Command (AC-DTC-004, AC-DTC-005):** `git diff -U0 --no-color develop...HEAD -- internal/cli/coverage_improvement_test.go internal/cli/doctor_test.go internal/cli/integration_test.go`
  — **Exit code:** `0` — **Verbatim stdout:** empty (zero bytes).
- **Why RED:** the card branch carries no Go change yet, so each criterion is unmet for the right
  reason: AC-DTC-003 lists zero of the three required paths, and AC-DTC-004 condition (2) and
  AC-DTC-005 find zero of the nine required isolation statements. Exit 0 with empty output is the
  RED observation here, not a pass.
- **Mutant evidence:** `.moai/reports/t675/red/mutant-probe.txt` shows the diff form each criterion
  reads for a correct variant, an assertion-deletion mutant, and a bare `os.Chdir` mutant.
- **Classification:** AC-DTC-003..005 are structural gates over the diff; their executed diff output
  on the implementation descendant is captured during run (`progress.md §E.2`).

## §4 Constraints

- The configured run methodology is TDD. The §3.1 ledger is the RED baseline; the isolation-only
  test edits provide GREEN.
- Scoped tests remain serial because process working directory and `doctorCmd` flags are
  process/package-global state.
- Local verification stays scoped to targeted `-run` selections in `./internal/cli/`; the full local
  repository suite is not run. CI on the integrated `develop` head owns the full-suite verdict.
- Temporary directories are not created under the repository.
- The implementation approach (framework-managed temporary directory and working-directory change)
  is fixed in `plan.md §D C3`.

## §5 Exclusions

### Out of Scope — doctor production behavior

- No production doctor check, root-discovery algorithm, Agent Emit Embed applicability rule,
  diagnostic text, command flag, or exit-status behavior is changed.

### Out of Scope — unrelated doctor tests

- No doctor test outside the exact nine-test list in §1 is modified or reclassified, including tests
  that already change directory themselves (for example `TestRunDoctor_FixMode`).

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
- RED raw evidence and mutant probe: `.moai/reports/t675/red/`
- Prior audit: `.moai/reports/t675/plan-audit.md` (FAIL 0.67)
- Prior observed failure set: `.moai/reports/t662/verdict.md` and
  `.moai/reports/t662/preattrib-develop-55b757ff2.log`
- Superseded gateway failure record: `.moai/reports/t675/gateway-502-20260913.md`
- Card: t675; worktree `.claude/worktrees/t675`; branch `WT-doctor-red`; RED baseline
  `dd235a66b1145922565841d33acafef0d1ded6a8`
