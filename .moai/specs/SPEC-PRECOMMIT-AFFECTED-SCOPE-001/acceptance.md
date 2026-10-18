# acceptance.md — SPEC-PRECOMMIT-AFFECTED-SCOPE-001

This file is the verification layer: every entry is Given / When / Then and binary. Requirements (GEARS) live in `spec.md` § B.

Conventions. "Run" commands are plain single invocations, scoped per `plan.md` § D (no full `internal/cli` suite). A criterion counts as met only when its named tests print `--- PASS` and the swept count is at least the number stated; a `[no tests to run]` line is a failure of the instrument, not a pass (see LEDGER-PAS-VACUOUS below). Tree pin for all starting observations: `3642a3a9b9cf101863656153d2d99b69fcb66a05`.

## D. AC Matrix

| AC | REQ | Class |
|----|-----|-------|
| AC-PAS-001 | REQ-PAS-002 | release-blocking |
| AC-PAS-002 | REQ-PAS-002 | release-blocking |
| AC-PAS-003 | REQ-PAS-003 | release-blocking |
| AC-PAS-004 | REQ-PAS-004 | release-blocking |
| AC-PAS-005 | REQ-PAS-005 | release-blocking |
| AC-PAS-006 | REQ-PAS-006 | release-blocking |
| AC-PAS-007 | REQ-PAS-007 | release-blocking |
| AC-PAS-008 | REQ-PAS-008 | release-blocking |
| AC-PAS-009 | REQ-PAS-009 | release-blocking |
| AC-PAS-010 | REQ-PAS-010 | regression-guard |
| AC-PAS-011 | REQ-PAS-011 | release-blocking |
| AC-PAS-012 | REQ-PAS-012 | withdrawn (Q1) |
| AC-PAS-013 | REQ-PAS-013 | release-blocking |
| AC-PAS-014 | REQ-PAS-014 | regression-guard |
| AC-PAS-015 | REQ-PAS-015 | regression-guard |
| AC-PAS-016 | REQ-PAS-001, REQ-PAS-010 | release-blocking |

### AC-PAS-001 — Direct dependents only (REQ-PAS-002)
**Given** an in-memory graph with `a`; `b` importing `a`; `c` importing `b`; `d` standalone, all with test files
**When** the changed list is `a/x.go`
**Then** the selection is exactly `{a, b}`; `c` and `d` are absent.
Run: `go test ./internal/hook/quality -run '^TestAffectedScope_DirectDependents$' -count=1 -v` — expect 1 PASS line, swept ≥ 1.
Mutant probe: a transitive-closure implementation selects `c` and fails this AC; a package-only implementation omits `b` and fails it.

### AC-PAS-002 — Test-import edges count (REQ-PAS-002)
**Given** package `e` that imports `a` only from its in-package `_test.go`, and package `f` that imports `a` only from an external `_test` package
**When** the changed list is `a/x.go`
**Then** both `e` and `f` are selected.
Run: `... -run '^TestAffectedScope_TestImportEdges$'`.
Mutant probe: an implementation reading only regular imports selects neither.

### AC-PAS-003 — Ownership of non-Go files (REQ-PAS-003)
**Given** a package `p` with `p/testdata/in.json`, a package `q` embedding `q/tpl/t.txt`, and a file `docs/note.md` under no package
**When** the changed list is those three files
**Then** the selection is `{p, q}` (with `p`/`q` dependents per AC-PAS-001 rules), and the summary line names `docs/note.md` as owned by no package.
Run: `... -run '^TestAffectedScope_NonGoOwnership$'`.

### AC-PAS-004 — Test-only changes do not widen (REQ-PAS-004)
**Given** the AC-PAS-001 graph
**When** the changed list is `a/x_test.go` and `a/testdata/y.json`
**Then** the selection is exactly `{a}`.
Run: `... -run '^TestAffectedScope_TestOnlyNarrows$'`.
Mutant probe: treating any change as a source change selects `b`.

### AC-PAS-005 — Deleted and renamed files (REQ-PAS-005)
**Given** the AC-PAS-001 graph with `a/x.go` deleted and `a` retaining no files, and separately `a/x.go` renamed to `a/y.go`
**When** the changed list is the deleted path (and, for the rename, both paths)
**Then** `b` is still selected by import-path match in the deleted case; in the rename case the selection equals the AC-PAS-001 result.
Run: `... -run '^TestAffectedScope_DeletedAndRenamed$'`.

### AC-PAS-006 — Dropped packages are counted; empty selection is an explicit skip (REQ-PAS-006)
**Given** a selection containing one package with test files and one without; and, separately, a changed list containing only `docs/note.md`
**When** the gate summarizes the test step
**Then** the first case reports `selected 2, dropped 1` and the dropped package with the reason `no test files`; the second case records the Go test step as skipped with a stated reason and the gate does not report a pass for that step.
Run: `... -run '^TestAffectedScope_DropAndEmpty$'` and `... -run '^TestQualityGate_AffectedScope_EmptySelectionSkips$'`.

### AC-PAS-007 — Widening triggers (REQ-PAS-007; decision-index Q2 default applied)
**Given** a changed list containing, one case each, `go.mod`, `go.sum`, `go.work`, `go.work.sum`, `.moai/config/build-tags`
**When** the Go test step's argument list is computed
**Then** it is the full-module scope (`./...`) and the summary names the trigger file.
Run: `... -run '^TestAffectedScope_WideningTriggers$'`.

### AC-PAS-008 — Failed queries fall back, never pass (REQ-PAS-008)
**Given** an injected changed-file query that fails, and separately an injected package-graph query that fails
**When** the gate computes the Go test step
**Then** the argument list is `./...`, the summary names which query failed, and no result is reported as a pass for the computation itself.
Run: `... -run '^TestQualityGate_AffectedScope_QueryFailureFallsBack$'`.

### AC-PAS-009 — Changed set equals what the commit records (REQ-PAS-009) — release-blocking
**Given** a temporary git repository with a real `pre-commit` script that records the changed set the gate computes, containing tracked files `a/x.go` and `d/z.go`, both modified, only `a/x.go` staged
**When** (i) `git commit`, (ii) `git commit -a`, (iii) `git commit d/z.go` are each run
**Then** the recorded sets are (i) `{a/x.go}`, (ii) `{a/x.go, d/z.go}`, (iii) `{d/z.go}`.
Run: `... -run '^TestAffectedScope_ChangedSetMatchesCommit$' -count=1 -v`.
Mutant probe: reading the real index under `-a` yields `{a/x.go}` for (ii) and fails; an implementation that includes unstaged files yields `{a/x.go, d/z.go}` for (i) and fails.

### AC-PAS-010 — Context isolation (REQ-PAS-010)
**Given** the gate constructed without the pre-commit flag (standalone and PreToolUse shapes), and separately constructed for the node and python toolchains with the flag
**When** each toolchain's step list is built
**Then** the Go test step argument list is `test ./...` without the flag; the node and python test steps are identical to today's; vet, lint, typecheck, graph-freshness and ast-grep steps are identical in all shapes.
Run: `... -run '^TestQualityGate_AffectedScope_ContextIsolation$'` (fake `go` binary records argv).

### AC-PAS-011 — Scope setting (REQ-PAS-011)
**Given** `gate.yaml` text with `test_scope.mode` absent, `affected`, `full`, and `bogus`
**When** the configuration is loaded and validated
**Then** absent → `affected`; `affected` and `full` accepted; `bogus` rejected with a message containing `gate.pre_commit.test_scope.mode`, `affected` and `full`; and the new block is byte-identical in `.moai/config/sections/gate.yaml` and `internal/template/templates/.moai/config/sections/gate.yaml`.
Run: `go test ./internal/config -run '^TestGatePreCommitTestScope$' -count=1 -v`; `go test ./internal/cli -run '^TestUpdateGateYAMLPreservesUserValues$' -count=1 -v` and `go test ./internal/cli -run '^TestUpdateGateYAMLDeliversNewPreCommitKey$' -count=1 -v` (existing tests; both must still pass, and the second is the pattern for delivering the new keys).

### AC-PAS-012 — WITHDRAWN (slot retained for traceability; maps the withdrawn REQ-PAS-012)
Withdrawn by the Q1 ruling (board d-20261018T005931Z-f264). It has no Given/When/Then and no test. Its preservation counterpart is AC-PAS-014: no skip path is added.

### AC-PAS-013 — Suppression check ignores fixtures, still catches production (REQ-PAS-013) — release-blocking
**Given** a temporary project with a copy of `suppressed.go` under `internal/astgrep/testdata/fixtures/go/` and a production file with a bare `// ast-grep-ignore`
**When** the ast-grep gate runs
**Then** exactly one violation is reported, for the production file; and the real-tool command in LEDGER-PAS-013-A exits 0 on the fixed tree.
Run: `go test ./internal/hook/quality -run '^TestSuppressionWalkSkipsTestdata$' -count=1 -v`; then the ledger command.
Mutant probe: a walk that skips every directory named `fixtures` would also hide production code, so the production file lives under a path containing a `fixtures` segment but no `testdata` segment and must still be reported; a walk that skips nothing reports the fixture and fails the first half.

### AC-PAS-014 — No timeout enlargement, no skips (REQ-PAS-014) — regression guard
**Given** the diff of this SPEC's commits against `3642a3a9b`
**When** the `timeouts:` block of both `gate.yaml` files and the default timeout values are compared, and the added lines are scanned
**Then** the timeouts are unchanged (30 / 60 / 120 / 300 / 30), the added lines contain no `t.Skip`, and the count of removed `func Test` lines is 0.
Run: `git diff 3642a3a9b -- .moai/config/sections/gate.yaml internal/template/templates/.moai/config/sections/gate.yaml` (the `timeouts:` hunk must be absent) and `git diff -U0 3642a3a9b -- '*_test.go'` read for added `t.Skip` lines and removed `func Test` lines (expect none).

### AC-PAS-015 — Hooks stay off (REQ-PAS-015) — regression guard
**Given** this SPEC's commits landed
**When** `git config --show-origin --get core.hooksPath` is run in the primary checkout, and the added Go lines are scanned for writes to git configuration
**Then** the command still prints `file:/Users/goos/MoAI/moai-adk-go/.git/config	/dev/null`, and no added line invokes `git config` with `hooksPath`.
Run: `git config --show-origin --get core.hooksPath`; `git diff -U0 3642a3a9b -- '*.go'` read for added lines containing `hooksPath` (expect none).

### AC-PAS-016 — End-to-end rehearsal on a throwaway module (REQ-PAS-001)
**Given** a temporary repository holding a three-package Go module with a real `pre-commit` script that exports `MOAI_PRECOMMIT=1` and runs `moai gate` (built from this tree), `pre_commit.enabled: true`, every ast-grep and graph step disabled, and a hook activated with `git -c core.hooksPath=<hooks dir> commit` (no shared configuration touched)
**When** (i) a change to the leaf package with a passing test is committed, (ii) a change to a package whose dependent has a failing test is committed
**Then** (i) the commit succeeds and the gate summary lists only the leaf package; (ii) the commit is refused with the failing dependent named, and the untouched third package's tests did not run.
Run: `go test ./internal/cli -run '^TestPreCommitAffectedScopeE2E$' -count=1 -v`.

## Evidence ledger (starting observations; tree 3642a3a9b9cf101863656153d2d99b69fcb66a05)

### LEDGER-PAS-013-A — RED-now for AC-PAS-013 (observed red, for the right stated reason)
Why red: the suppression check walks `testdata` and the fixture's REASON line precedes its marker (spec.md F8), so a gate run reports the fixture even in advisory ast-grep configuration.
Recipe (re-executable): a project directory holding `go.mod` (`module example.com/x`, `go 1.23`), the fixture copied to `internal/astgrep/testdata/fixtures/go/suppressed.go`, the repository's `.moai/astgrep-rules` copied to `.moai/astgrep-rules`, and `.moai/config/sections/gate.yaml`:
```
gate:
  enabled: true
  skip_tests: true
  disabled_steps:
    "go vet": false
    "golangci-lint": false
    "go test": false
  typecheck:
    enabled: false
  ast_grep_gate:
    enabled: true
    rules_dir: ".moai/astgrep-rules"
    block_on_error: false
    warn_only_mode: true
  graph_freshness:
    enabled: false
```
Command: `env CLAUDE_PROJECT_DIR=<that directory> <moai built from the tree under measurement> gate`
Exit code: 1
Output as printed (stream-merged; banner trailing whitespace not preserved):
```
ast-grep suppression policy violations:

[SUPPRESSION_WITHOUT_REASON] ast-grep suppression at <project>/internal/astgrep/testdata/fixtures/go/suppressed.go:15 requires adjacent '// @MX:REASON <rationale>' on next line

quality gate steps (3 configured):
  - go vet: disabled — turned off by the gate.disabled_steps setting
  - golangci-lint: disabled — turned off by the gate.disabled_steps setting
  - go test: not reached

   ERROR

  Quality gate failed.
```
Green path: M5 flips it; after the fix the same command exits 0 and prints no suppression line.
Not covered by this cell: whether any other file in the full repository trips the checker (not run; grep found none outside `_test.go`, spec.md F9). M6 runs the command against a copy of the whole tree's source directories.

### LEDGER-PAS-VACUOUS — why behavior ACs 001–012 and 016 carry no RED-now command
Command: `go test ./internal/hook/quality -run '^TestAffectedScope$' -count=1`
Exit code: 0
Output: `ok  	github.com/modu-ai/moai-adk/internal/hook/quality	0.510s [no tests to run]`
This is NOT a red. The named tests do not exist yet, and the runner reports success with an empty swept set. Therefore each of those ACs is adopted with: RED cell = the run phase captures, before any implementation (E8), the verbatim failing output of the named test (compile failure or assertion); green path = the milestone named in `plan.md` § C flips it. An AC whose run output shows `[no tests to run]` or a swept count of 0 is failed, never passed.

### Regression-guard classification
AC-PAS-010, -014 and -015 assert preservation (green today by construction). They are regression guards, not release-blocking RED→GREEN criteria; their mutants are named in the entries.

## Edge cases the tests must cover
- A changed file inside a nested module: handled relative to that module's root; files outside it contribute nothing to that module's step.
- A commit renaming a package directory: both former and new directories attributed (AC-PAS-005).
- `git commit --amend` with nothing newly staged: empty selection, explicit skip (AC-PAS-006).
- A build-constraint-excluded Go file: owner is its directory's package (plan.md R4).

## Definition of Done
1. AC-PAS-001..-011, -013, -016 pass with swept count ≥ 1 and RED output captured before GREEN; AC-PAS-012 is withdrawn; AC-PAS-014 and -015 guards pass.
2. LEDGER-PAS-013-A command exits 0 on the final tree, from a binary built from that tree (tool-provenance: the judging build's commit is cited beside the tree HEAD).
3. `internal/hook/quality`, `internal/config` pass in full; `internal/cli` and the other dependents pass under the § D selection with the `ls` evidence for dropped importers recorded.
4. Both `gate.yaml` copies updated; no timeout value changed; no skip added; `core.hooksPath` unchanged.
5. `progress.md` §E.2 carries the verbatim evidence, including the M3 measurement of the three commit shapes.
