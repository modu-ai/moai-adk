# acceptance.md — SPEC-PRECOMMIT-AFFECTED-SCOPE-001

This file is the verification layer: every entry is Given / When / Then and binary. Requirements (GEARS) live in `spec.md` § B.

Conventions. "Run" commands are plain single invocations, scoped per `plan.md` § D (no full `internal/cli` suite). A criterion counts as met only when its named tests print `--- PASS` and the swept count is at least the number stated; a `[no tests to run]` line is a failure of the instrument, not a pass (see LEDGER-PAS-VACUOUS). Tree pin for the starting observations in the ledger: `888ce5c035a2527d81507441d71e69c929042ad0`; the `moai` binary used was built from `3642a3a9b`, and `git diff --stat 3642a3a9b HEAD -- internal cmd go.mod go.sum` prints nothing, so the Go code is identical (tool-provenance: judging build and tree HEAD cited together).

## D. AC Matrix

| AC | REQ | Class |
|----|-----|-------|
| AC-PAS-001 | REQ-PAS-002 | release-blocking (RED-now: LEDGER-PAS-016-A) |
| AC-PAS-002 | REQ-PAS-002 | adoption-pending (policy P1) |
| AC-PAS-003 | REQ-PAS-003, REQ-PAS-004 | adoption-pending (policy P1) |
| AC-PAS-004 | REQ-PAS-004 | release-blocking (RED-now: LEDGER-PAS-016-A, test-only scenario) |
| AC-PAS-005 | REQ-PAS-005 | release-blocking (RED-now: LEDGER-PAS-009-A row iv) |
| AC-PAS-006 | REQ-PAS-006 | release-blocking (RED-now: LEDGER-PAS-016-A, docs-only scenario) |
| AC-PAS-007 | REQ-PAS-007 | regression-guard (green today by construction) |
| AC-PAS-008 | REQ-PAS-008 | regression-guard (green today by construction) |
| AC-PAS-009 | REQ-PAS-009, REQ-PAS-005 | release-blocking (RED-now: LEDGER-PAS-009-A) |
| AC-PAS-010 | REQ-PAS-010 | regression-guard |
| AC-PAS-011 | REQ-PAS-011 | release-blocking (RED-now: LEDGER-PAS-016-A, invalid-mode scenario) |
| AC-PAS-012 | REQ-PAS-012 | withdrawn (Q1) |
| AC-PAS-013 | REQ-PAS-013 | release-blocking (RED-now: LEDGER-PAS-013-A) |
| AC-PAS-014 | REQ-PAS-014 | regression-guard |
| AC-PAS-015 | REQ-PAS-015 | regression-guard |
| AC-PAS-016 | REQ-PAS-001, REQ-PAS-010 | release-blocking (RED-now: LEDGER-PAS-016-A) |

Policy P1 (adoption of unit-level criteria). Criteria marked adoption-pending are unadopted in the sense of `verification-completeness.md` §2 until their RED-now cell exists: the run phase writes each such test before its implementation, and records in `progress.md` §E.2 one ledger row per AC holding the command, the verbatim failing stdout, the exit code, and the tree SHA, before the milestone's GREEN commit. An adoption-pending AC reported as passed without its row is a Gap, not a pass. Only AC-PAS-002 and AC-PAS-003 remain under P1: they judge the pure selection core, which has no observable end-to-end behaviour today, and their tests do not exist, so the only available "red" for them is an empty-swept-set green (LEDGER-PAS-VACUOUS). AC-PAS-004, -006 and -011 have behaviour-level RED-now cells in LEDGER-PAS-016-A.

### AC-PAS-001 — Direct dependents only (REQ-PAS-002)
**Given** an in-memory graph with `a`; `b` importing `a`; `c` importing `b`; `d` standalone, all with test files
**When** the changed list is `a/x.go`
**Then** the selection is exactly `{a, b}`; `c` and `d` are absent.
Run: `go test ./internal/hook/quality -run '^TestAffectedScope_DirectDependents$' -count=1 -v` — expect 1 PASS line, swept ≥ 1.
Mutant probe: a transitive-closure implementation selects `c` and fails; a package-only implementation omits `b` and fails.

### AC-PAS-002 — Test-import edges count (REQ-PAS-002)
**Given** package `e` that imports `a` only from its in-package `_test.go`, and package `f` that imports `a` only from an external `_test` package
**When** the changed list is `a/x.go`
**Then** both `e` and `f` are selected.
Run: `... -run '^TestAffectedScope_TestImportEdges$'`.
Mutant probe: an implementation reading only regular imports selects neither.

### AC-PAS-003 — Ownership and widening per file kind (REQ-PAS-003, REQ-PAS-004)
**Given** the graph of AC-PAS-001 plus: package `p` with `p/testdata/in.json` and `p/testdata/fix.go`, and package `pp` importing `p`; package `q` whose resolved embedded-file list contains `q/tpl/t.txt`, and package `r` importing `q`; a file `docs/note.md` under no package directory
**When** the changed list is `p/testdata/in.json`, `p/testdata/fix.go`, `q/tpl/t.txt`, `docs/note.md`
**Then** the selection is exactly `{p, q, r}`: `p` alone (test-only changes, `pp` absent, and the `.go` fixture is owned by `p`, not ownerless); `q` and its dependent `r` (an embedded file is a source change); and the summary names `docs/note.md` as owned by no package.
Run: `... -run '^TestAffectedScope_Ownership$'`.
Mutant probe: a directory-of-the-file rule leaves `p/testdata/fix.go` ownerless and fails; treating embedded files as test-only omits `r` and fails; treating `testdata` as source selects `pp` and fails.

### AC-PAS-004 — Test-only changes do not widen (REQ-PAS-004)
**Given** the AC-PAS-001 graph
**When** the changed list is `a/x_test.go`, `a/testdata/y.json` and `a/testdata/f.go`
**Then** the selection is exactly `{a}`.
Run: `... -run '^TestAffectedScope_TestOnlyNarrows$'`.
Mutant probe: treating any change as a source change selects `b`.

### AC-PAS-005 — Deleted and renamed files, including through the real git adapter (REQ-PAS-005) — release-blocking
**Given** (core) the AC-PAS-001 graph with `a/x.go` deleted and `a` retaining no files, and separately `a/x.go` renamed to `a/y.go`; and (adapter) a temporary git repository in which `git mv a/x.go e/x.go` (cross-package rename) is staged, and in which the last file of a package is deleted
**When** the core receives the old path (core cases), and the REAL git adapter produces the changed list from the real staged state (adapter cases)
**Then** core: `b` is selected by import-path match in the deleted case, and the rename case equals the AC-PAS-001 result; adapter: the produced list contains both `a/x.go` and `e/x.go`, and the importers of the old package are selected end to end.
Run: `... -run '^TestAffectedScope_DeletedAndRenamed$'` and `... -run '^TestAffectedScope_RealGitRenameDelete$'`.
Mutant probe: an adapter using the default name list drops `a/x.go` and fails the adapter case (the core case alone, which injects both paths, cannot catch it).

### AC-PAS-006 — Dropped packages are counted; empty selection is an explicit skip (REQ-PAS-006)
**Given** a selection containing one package with test files and one without; and, separately, a changed list containing only `docs/note.md`
**When** the gate summarizes the test step
**Then** the first case reports `selected 2, dropped 1` and the dropped package with the reason `no test files`; the second case records the Go test step as skipped with a stated reason and the gate does not report a pass for that step.
Run: `... -run '^TestAffectedScope_DropAndEmpty$'` and `... -run '^TestQualityGate_AffectedScope_EmptySelectionSkips$'`.

### AC-PAS-007 — Widening triggers, evaluated before module filtering (REQ-PAS-007) — regression-guard
**Given** a changed list containing, one case each, `go.mod`, `go.sum`, `go.work`, `go.work.sum`, `.moai/config/build-tags`; and a nested-module layout whose Go module root is `m/`, with the trigger file at the repository root (`go.work`, and separately `.moai/config/build-tags`) and every other changed file inside `m/`
**When** the Go test step's argument list is computed
**Then** it is the full-module scope (`./...`) and the summary names the trigger file, in every case including the nested layout.
Run: `... -run '^TestAffectedScope_WideningTriggers$'`.
Mutant probe: filtering to the module root before checking triggers drops the root-level `go.work` and narrows the scope; the nested case fails it.

### AC-PAS-008 — Failed or incomplete queries fall back, never pass (REQ-PAS-008) — regression-guard
**Given** an injected changed-file query that fails; an injected package-graph query that fails; and, through the REAL `go list` adapter, a temporary module in which one package imports a nonexistent dependency (`go list -e` exits 0 and marks the package incomplete)
**When** the gate computes the Go test step
**Then** in every case the argument list is `./...`, the summary names the failed query or the incomplete package, and no result is reported as a pass for the computation itself.
Run: `... -run '^TestQualityGate_AffectedScope_QueryFailureFallsBack$'` and `... -run '^TestAffectedScope_IncompleteGraphFallsBack$'`.
Mutant probe: an adapter that trusts a zero exit status selects a normal set from the incomplete graph and fails the third case.

### AC-PAS-009 — Changed set equals what the commit records (REQ-PAS-009, REQ-PAS-005) — release-blocking
**Given** a temporary git repository with a real `pre-commit` script that records the changed set the gate computes, containing tracked files `a/x.go` and `d/z.go`, both modified, only `a/x.go` staged; the same repository also as a linked worktree
**When** (i) `git commit`, (ii) `git commit -a`, (iii) `git commit d/z.go`, (iv) a staged rename `old/o.go` → `new/o.go`, (v) `git commit --amend` after staging one further file are each run
**Then** the recorded sets are (i) `{a/x.go}`, (ii) `{a/x.go, d/z.go}`, (iii) `{d/z.go}`, (iv) `{old/o.go, new/o.go}`, (v) exactly the newly staged file; identical in the plain repository and in the linked worktree.
Run (the only green path for this AC): `... -run '^TestAffectedScope_ChangedSetMatchesCommit$' -count=1 -v`. `evidence/probe009.sh` is the immutable RED-now witness of the current helper and is never a green condition.
Mutant probe: reading the real index under `-a` yields `{a/x.go}` for (ii) and fails; an implementation that includes unstaged files yields `{a/x.go, d/z.go}` for (i) and fails; the default name list yields `{new/o.go}` for (iv) and fails.

### AC-PAS-010 — Context isolation (REQ-PAS-010) — regression-guard
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

### AC-PAS-014 — Step budgets unchanged at every definition and call site (REQ-PAS-014) — regression-guard
**Given** the gate's budget definition sites — the `DefaultGateConfig` values (30 s vet, 60 s lint, 120 s test, 300 s typecheck, 30 s lock wait) in `internal/hook/quality/gate.go`, and the `timeouts:` block of both `gate.yaml` copies (30 / 60 / 120 / 300 / 30) — and the call sites that pass them to the step runner (vet, typecheck, lint and test steps in `gate.go`)
**When** the budget test drives the gate with the default configuration and a recording runner, and the diff of this SPEC's commits is read
**Then** the runner receives exactly 30 s, 60 s, 120 s and 300 s for the vet, lint, test and typecheck steps respectively; the `timeouts:` hunk of both `gate.yaml` files is absent from the diff; the added `_test.go` lines contain no `t.Skip`; the removed lines contain no `func Test`.
Run: `... -run '^TestQualityGate_StepBudgetsUnchanged$' -count=1 -v`; `git diff 3642a3a9b -- .moai/config/sections/gate.yaml internal/template/templates/.moai/config/sections/gate.yaml`; `git diff -U0 3642a3a9b -- '*_test.go'` read for added `t.Skip` and removed `func Test` lines (expect none).
Mutant probe: replacing the test step's call-site argument with a multiple of the configured test budget (for example `2*g.config.TestTimeout` at the test-step call in `gate.go`) leaves every value in the diff checks unchanged but makes the runner receive 240 s — the budget test fails it; changing a default or a YAML value fails both.

### AC-PAS-015 — No code path writes git configuration (REQ-PAS-015) — regression-guard
**Given** the diff of this SPEC's deliverables against `3642a3a9b`
**When** the added non-test Go lines and the added lines of the shipped YAML and hook files are scanned
**Then** no added line contains `hooksPath`, and no added line invokes `git` with a `config` (or `-c`) argument that writes configuration; the hook script and its Go raw-string twin (`internal/template/templates/.git_hooks/pre-commit`, `internal/cli/hook_install_precommit.go`) are unmodified.
Run: `git diff -U0 3642a3a9b -- '*.go' '*.yaml'` read for added lines containing `hooksPath` or a git-config write (expect none, test files excluded); `git diff --stat 3642a3a9b -- internal/template/templates/.git_hooks internal/cli/hook_install_precommit.go` (expect empty).
This AC does not assert the live value of the shared `core.hooksPath`, which is mutable operator state: the sanctioned post-landing switch (plan.md § E) must not turn a code guard red. The value is recorded in the runbook's step 1 instead.
Mutant probe: a helper that runs `git config core.hooksPath <path>` fails the scan; a no-op change does not.

### AC-PAS-016 — End-to-end rehearsal on a throwaway module (REQ-PAS-001, REQ-PAS-010) — release-blocking
**Given** a temporary repository holding a three-package Go module (`a`; `b` importing `a`; `c` standalone, each with a test) with a real `pre-commit` script that exports `MOAI_PRECOMMIT=1` and runs `moai gate` (built from this tree), `pre_commit.enabled: true`, every ast-grep and graph step disabled, and a hook activated by `core.hooksPath` of that temporary repository only (no shared configuration touched)
**When** (i) a change to `a/a.go` is committed with a wrapper `go` on PATH that delegates every verb except `test` to the real `go` (so the gate's `go list` graph query works) and records only `test` invocations; (ii) a change to `a` is committed with the real `go`, `b`'s test failing and `c`'s test writing a marker file when it runs
**Then** (i) the commit succeeds (exit 0, HEAD moved) AND the recorded `go test` argv is exactly `test ./a ./b` — module-root-relative directory patterns, the one argument form used by REQ-PAS-001, plan.md B.2 and this ledger; (ii) the commit is refused, the output names `b` as failing, and the marker file written by `c`'s test is absent. A refused commit never counts as GREEN.
Run: `go test ./internal/cli -run '^TestPreCommitAffectedScopeE2E$' -count=1 -v`.
Mutant probe: any implementation that leaves `./...` fails (i); one that omits `b` fails (i) and (ii).

## Evidence ledger (starting observations)

### LEDGER-PAS-016-A — RED-now for AC-PAS-016 and AC-PAS-001 (observed red, for the right stated reason)
Why red: REQ-PAS-001/-002 are not implemented, so the pre-commit gate hands the Go test step `./...` even though only `a` and its direct dependent `b` are affected.
Also the RED-now cell for AC-PAS-004 (test-only scenario), AC-PAS-006 (docs-only scenario) and AC-PAS-011 (invalid-mode scenario), one line each in the output below.
Command: `env PROBE_DIR=/private/tmp/pas-probe sh .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/evidence/probe016.sh <moai built from the tree under measurement>` — the script (committed beside this file) builds the three-package module, a wrapper `go` on PATH that delegates every verb except `test` to the real `go` and records only `test` invocations, and a hook that runs `MOAI_PRECOMMIT=1 CLAUDE_PROJECT_DIR=<repo> moai gate`. Per scenario it requires the commit to succeed (exit 0, HEAD moved) and the WHOLE recorded `go test` log to equal the wanted one — exactly one call for a source change, exactly one for a test-only change, zero for a docs-only change, argv matching, so an extra `go test` call after the right one is RED; the last scenario runs the gate with `test_scope.mode: bogus` and requires a non-zero exit whose message names `gate.pre_commit.test_scope.mode`. It exits 1 if any scenario is RED.
Tree: `576da7f1933e44c5c144c0bcbeebef3d232f12ec` (clean apart from the corrected `probe016.sh` itself); judging build: `moai` built from `3642a3a9b` (Go code identical, see header).
Exit code: 1
Output (the gate's own banner lines omitted):
```
AC-016/001 change a: RED  (commit exit=0, HEAD moved=yes, go test calls=1 (wanted 1), wanted go argv 'test ./a ./b', got 'test ./...')
AC-004 test-only change: RED  (commit exit=0, HEAD moved=yes, go test calls=1 (wanted 1), wanted go argv 'test ./a', got 'test ./...')
AC-006 docs-only change: RED  (commit exit=0, HEAD moved=yes, go test calls=1 (wanted 0), wanted go argv 'none', got 'test ./...')
AC-011 mode: bogus: RED  (gate exit=0, message names the key=no)
exit=1
```
Green path: M3 flips it; the same script then prints GREEN for all four scenarios and exits 0. The reachable green path is real: the wrapper lets a `go list` query through and the comparison covers the whole recorded `go test` log (call count and every line) while `go list` calls are not recorded, so an implementation that queries the graph first still passes and one that adds an extra `go test` call does not. Negative controls (stub gates under `/private/tmp`, never committed): a stub recording only the correct calls reads GREEN for all four scenarios; the same stub followed by an extra `go test ./...` reads RED for the three commit scenarios.

### LEDGER-PAS-009-A — RED-now for AC-PAS-009 and for the adapter half of AC-PAS-005 (immutable witness of the CURRENT helper)
This row records the present behaviour of the existing staged-file helper. `probe009.sh` hard-codes that helper's query, so it can never turn green and is not a green condition anywhere in this SPEC; the only green path for AC-PAS-009 and for the adapter half of AC-PAS-005 is `TestAffectedScope_ChangedSetMatchesCommit` and `TestAffectedScope_RealGitRenameDelete`, which call the real adapter.
Why red: the gate's existing staged-file helper reads the real index with the repository-location variables removed, so it does not report what the commit records under `commit -a` and path-limited commits, and the default name list reports only a rename's new path.
Command: `env PROBE_DIR=/private/tmp/pas-probe sh .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/evidence/probe009.sh` — a temporary repository whose hook records (a) the commit-accurate list (`git diff --cached --name-only --no-renames`, environment as git gave it) and (b) the helper's form (`git diff --cached --name-only` with the `gitenv.RepoScopingVars` entries removed from the environment: `GIT_DIR`, `GIT_COMMON_DIR`, `GIT_WORK_TREE`, `GIT_INDEX_FILE`, `GIT_OBJECT_DIRECTORY`, `GIT_ALTERNATE_OBJECT_DIRECTORIES`, `GIT_NAMESPACE`, `GIT_PREFIX`, `GIT_SUPER_PREFIX`, `GIT_CEILING_DIRECTORIES`, `GIT_QUARANTINE_PATH`); it exits 1 when (b) differs from (a). The probe replicates the helper's query and scrub; it does not call the helper (unexported).
Tree: `888ce5c035a2527d81507441d71e69c929042ad0` (Go code identical to `3642a3a9b`).
Exit code: 1
Output:
```
(ii)  git commit -a            (a/x.go staged, d/z.go modified unstaged)
  commit records : a/x.go d/z.go 
  helper form    : a/x.go    [DIFFERENT]
(iii) git commit d/z.go        (path-limited)
  commit records : d/z.go 
  helper form    : a/x.go    [DIFFERENT]
(iv)  git mv old/o.go new/o.go (rename)
  commit records : new/o.go old/o.go 
  helper form    : new/o.go    [DIFFERENT]
exit=1
```
Row (iv) is the rename: the helper's default name list omits `old/o.go` (REQ-PAS-005 requires both paths). The plan-audit iter1 probe (`.moai/reports/t1673/plan-audit-iter1.md` E2) measured the same three divergences independently.
Green path: none in this script. M3 flips AC-PAS-009 through `TestAffectedScope_ChangedSetMatchesCommit` (real adapter, real git, plain repository and linked worktree).
Not covered by this cell: `--amend` and the linked-worktree form (`GIT_DIR` = `.git/worktrees/<name>`); M3 measures them and records the rows in `progress.md` §E.2.

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
Tree: `3642a3a9b9cf101863656153d2d99b69fcb66a05` (measured by the author), reproduced on `2febc5c2f5036059724819044b67131f64e9a596` by the plan-audit iter1 (E1).
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

### LEDGER-PAS-VACUOUS — why AC-PAS-002 and AC-PAS-003 carry no RED-now command (policy P1 applies)
Command: `go test ./internal/hook/quality -run '^TestAffectedScope$' -count=1`
Exit code: 0
Output: `ok  	github.com/modu-ai/moai-adk/internal/hook/quality	0.510s [no tests to run]`
This is NOT a red. The named tests do not exist yet, and the runner reports success with an empty swept set. AC-PAS-002 and AC-PAS-003, which judge the selection core alone, are therefore adopted by policy P1 (the failing output captured before GREEN, recorded as a ledger row); every other AC that can be shown red today carries its cell above (001, 004, 006, 011 and 016 via LEDGER-PAS-016-A, 005 and 009 via LEDGER-PAS-009-A, 013 via LEDGER-PAS-013-A). An AC whose run output shows `[no tests to run]` or a swept count of 0 is failed, never passed.

### Regression-guard classification
AC-PAS-007, -008, -010, -014 and -015 assert preservation or fallback behavior that equals today's by construction (the full scope, the unchanged budgets, no configuration writes). They are regression guards, not RED→GREEN criteria; their mutants are named in the entries.

## Edge cases the tests must cover
- A changed file inside a nested module: handled relative to that module's root; files outside it contribute nothing to that module's step, but a trigger file anywhere in the changed list widens first (AC-PAS-007).
- A commit renaming a package directory: both former and new directories attributed (AC-PAS-005).
- `git commit --amend` with nothing newly staged: empty selection, explicit skip (AC-PAS-006, AC-PAS-009 row v).
- A build-constraint-excluded Go file: owner is its directory's package (plan.md R4).

## Definition of Done
1. Release-blocking ACs (001, 004, 005, 006, 009, 011, 013, 016) pass with swept count ≥ 1, each flipped from its ledger RED-now cell; adoption-pending ACs (002, 003) pass only with their P1 ledger row in `progress.md` §E.2; regression guards (007, 008, 010, 014, 015) pass; AC-PAS-012 is withdrawn.
2. LEDGER-PAS-013-A command exits 0 on the final tree, from a binary built from that tree (tool-provenance: the judging build's commit is cited beside the tree HEAD); likewise `probe016.sh`, which judges the whole recorded `go test` log and call count, prints GREEN for all four scenarios and exits 0. `probe009.sh` is the immutable witness of the pre-change helper and is not a green condition; AC-PAS-009 is green only through `TestAffectedScope_ChangedSetMatchesCommit`.
3. `internal/hook/quality`, `internal/config` pass in full; `internal/cli` and the other dependents pass under the § D selection with the `ls` evidence for dropped importers recorded.
4. Both `gate.yaml` copies updated; no timeout value changed; no skip added; no code path writes git configuration.
5. `progress.md` §E.2 carries the verbatim evidence, including the M3 measurement of `--amend` and the linked-worktree form.
