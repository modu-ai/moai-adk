# plan.md — SPEC-PRECOMMIT-AFFECTED-SCOPE-001

Tier M. Milestones are ordered by decision-reversibility: user-visible configuration and semantics first (most likely to change on review), mechanical fixes last. Priority labels only; no time estimates.

## A. Context

- Work tree: card t1673, branch `WT-pre-commit-gate`, planned against `3642a3a9b9cf101863656153d2d99b69fcb66a05`.
- Development mode per `.moai/config/sections/quality.yaml` (read at run entry; this plan is mode-neutral: every AC names a test whose RED output is captured before GREEN).
- Verification constraints that bind the run phase (leader dispatch, not negotiable here): no test skips, no timeout enlargement (gate budgets stay 30/60/120/300), no full `internal/cli` suite (ruling 35e9; use `-run`), an importer with zero `*_test.go` files is dropped from remeasure with the reason `no test files (ls)` (ruling 148c).
- Decisions: see `decision-index.md` (Q1 ruled B and Q5 ruled hooks-off, both by the leader; Q2 and Q3 carry applied defaults; Q4 is a run-phase measurement and gates nothing in the code).

## B. Technical approach

### B.1 Where the scope decision lives

The marker branch already exists in the CLI entry (`MOAI_PRECOMMIT=1` plus `gate.pre_commit.enabled`). The scope decision is added one level down: the CLI entry tells the gate it is in the pre-commit context, and the Go test step, only for the Go toolchain and only in that context, replaces its `./...` argument with the computed package list. Every other path constructs the gate without that flag and keeps `./...` (REQ-PAS-010). Build-tag injection continues to apply to the replaced argument list.

### B.2 Affected-set computation

Inputs: (1) changed files, repo-relative; (2) the package graph of the toolchain's module root, from one `go list -e -json` query over `./...` (fields: import path, directory, Go/ignored/test file lists, the three import lists, the resolved embedded-file lists `EmbedFiles`/`TestEmbedFiles`/`XTestEmbedFiles` — use these already-resolved lists rather than re-implementing pattern resolution; M3 confirms the fields on this toolchain before relying on them — and the package error/incomplete markers); (3) the module root's offset inside the repository.

Steps:
0. Widening check over ALL changed paths, before any filtering by module root: if any path is `go.mod`, `go.sum`, `go.work`, `go.work.sum` or `.moai/config/build-tags` (at any directory level, so a repository-root `go.work` or build-tags file is seen from a nested module — REQ-PAS-007), return the full scope and the trigger name.
1. Drop changed files outside the module root (the toolchain's step runs only in that root).
2. Owner of each file, in the REQ-PAS-003 order: (a) a path with a `testdata` segment belongs to the package whose directory contains that `testdata` directory, whatever its extension — `go list` never enumerates `testdata` as a package, and this repository tracks `.go` fixtures there, so a directory-of-the-file rule would leave them ownerless; (b) a file in the package's Go, ignored-Go (excluded on this platform), cgo or test lists belongs to its directory's package; (c) any other file belongs to the package whose resolved embedded-file list contains it, else to the nearest ancestor package directory.
3. For deleted or renamed paths use the former directory's import path directly (REQ-PAS-005); no on-disk package is needed to look up importers. The git adapter requests the list with rename detection off, because the default list reports only a rename's new path.
4. Direct dependents: for each package that owns at least one source change (REQ-PAS-004: every changed file that is not a `_test.go` file and not under `testdata`, embedded and non-Go ancestor-fallback files included), add every package whose regular, test or external-test import list contains its import path. A package whose changed files are all test-only is added alone.
4a. If the package graph reports any package as errored or incomplete (missing dependency; `go list -e` still exits 0), stop and fall back to the full scope with a notice naming the package (REQ-PAS-008).
5. Drop packages with no test files (in-package or external); record the counts (REQ-PAS-006).
6. Hand the remaining packages to the existing step runner as module-root-relative directory patterns (`./a ./b`; the import path is mapped back to its directory relative to the module root — the one argument form used by REQ-PAS-001, AC-PAS-016 and the ledger, stable inside a nested module because the step runs at that root) as the argument list, preserving the 120 s budget, the process-group termination and the stdout/stderr merging unchanged.

The core (steps 2–5) takes the changed list and a plain in-memory graph and returns the selection plus the report lines; nothing in it spawns a process or reads time or environment. Two thin adapters feed it: the git adapter (changed files, see B.3) and the `go list` adapter. Tests inject both (B.5).

### B.3 Commit-accurate changed set (REQ-PAS-009) — measure first

The existing staged-file helper scrubs `GIT_INDEX_FILE` and `GIT_DIR` from its environment on purpose (card t560: steps run inside a hook must not address the caller's repository). For a plain `git commit` the real index equals the commit's content, so the scrubbed query is right. For `git commit -a` and `git commit <paths>`, git runs the hook against a temporary index; a query against the real index can then miss the files the commit will record, and an under-scoped set would pass a commit whose tests were never run. This is now observed on the current tree (acceptance.md LEDGER-PAS-009-A: under `commit -a` the scrubbed form misses `d/z.go`; under a path-limited commit it returns `a/x.go`, which the commit does not record and misses `d/z.go`; the plan-audit iter1 probe E2 reports the same). Milestone M3 re-measures the three shapes plus `--amend` and a linked worktree (whose `GIT_DIR` is `.git/worktrees/<name>`, see `internal/gitenv`) in a temporary repository with a real hook, then picks the narrowest mechanism that makes AC-PAS-009 pass — candidate: read the hook-provided index location for this one query only, leaving every step's environment scrubbed. The list is requested with rename detection off (REQ-PAS-005), which the default form lacks (LEDGER-PAS-009-A row iv). `--amend`: per REQ-PAS-009 the changed set is the changes newly staged in the index relative to the HEAD being amended; files already in the amended commit and not re-staged are not re-tested (accepted incremental behaviour). The helper must also distinguish "nothing staged" (empty selection, explicit skip) from "query failed" (REQ-PAS-008 fallback); today both return nil.

### B.4 Configuration

New keys under `gate.pre_commit` in `gate.yaml`: `test_scope.mode` (`affected` default, `full`). No package-exclusion key exists (Q1 ruled B). Touch points: config types, defaults, validation (reject unknown mode), the CLI mapping into the gate configuration, the local `gate.yaml`, and its template twin (`internal/template/templates/.moai/config/sections/gate.yaml`; the two files differ today only in `pre_commit.enabled` and `rules_dir`, F4 — keep those differences). `moai update` already preserves user edits to `gate.yaml` (covered by `update_gate_yaml_preserve_test.go`); the new keys must not break that behaviour. The config struct ↔ YAML symmetry guard (`audit_struct_yaml_symmetry_test.go`, finding `CONFIG_STRUCT_YAML_MISMATCH` per the settings-management rule) is checked at M1 and its case list updated if the new keys require it (not read in this plan phase). The `moai web` gate panel is not extended (Out of Scope: nothing requires it).

### B.5 Deterministic test strategy

- Core: table-driven tests over hand-written in-memory graphs and changed lists. No `go list`, no git, no clock. Covers REQ-PAS-002/-003/-004/-005/-006/-007.
- Adapters: one test builds a throwaway Go module in `t.TempDir()` (packages `a`, `b` importing `a`, `c` importing `b`, `d` standalone, one with an external test, one with `testdata`, one embedding a file), runs the real `go list` adapter, and compares to the expected import-path set. `go` is necessarily present whenever `go test` runs, so no skip guard is added (REQ-PAS-014).
- Git: tests create a temporary repository, install a real `pre-commit` script that records its view of the changed set, and commit with `git commit`, `-a`, explicit paths, `--amend`, a cross-package rename and a last-file deletion, once in a plain repository and once in a linked worktree (REQ-PAS-005/-009). These tests exercise the REAL git adapter — the core is never handed the rename's old path by the test, so an adapter that drops it fails. They call git with a scrubbed environment as the existing `gate_staged_files_git_env_test.go` pattern does.
- `go list` adapter: the throwaway module also holds a package importing a nonexistent dependency (`go list -e` exits 0 and marks it incomplete), a `.go` fixture under `testdata`, and a nested-module layout with a repository-root `go.work` and build-tags file; the real adapter must report the incomplete package (REQ-PAS-008) and the widening trigger must fire before module filtering (REQ-PAS-007).
- Gate wiring: a fake `go` binary on PATH (pattern already used in `gate_test.go`, `writeFakeBinary`) records its argv, proving the exact argument list for each scenario without running real tests.
- Budget values: a test hands the step runner through an injectable boundary and asserts the exact durations it receives for the vet, lint, test and typecheck steps (30 s, 60 s, 120 s, 300 s) — a comparison of argument values, never a measurement of elapsed time — so a call site that multiplies a budget (for example the `g.config.TestTimeout` call in the test step) turns the test red.
- Nothing asserts elapsed time.

### B.6 Suppression-policy fix (REQ-PAS-013)

The file walk that feeds the suppression check skips any path with a `testdata` segment, alongside the segments it already skips (`vendor`, `.git`, `node_modules`, `__pycache__`) and `_test.go` files. This matches the stated rationale of that walk ("suppression policy checks apply to production code") and the sibling path filter used for ast-grep findings, which already excludes `/testdata/`. Alternatives rejected with evidence: reordering the Go fixture (observed: a marker-first variant makes `sg scan` report `go-interface-empty-not-any`, so `TestRuleSeed` `fixtures-scan` would fail); letting the checker accept REASON before the marker (redefines SPEC-UTIL-002's policy, wider than the defect — decision-index Q3).

## C. Milestones

### M1 — Configuration surface and template twin (Priority High; most likely to change)
RED: config tests for default (`affected`), valid values, invalid-value rejection, and byte-level agreement of the two `gate.yaml` copies on the new block. GREEN: types, defaults, validation, loader mapping, both YAML files. Covers REQ-PAS-011; AC-PAS-011.

### M2 — Affected-set core (Priority High)
RED→GREEN over the injected-graph tables. Covers REQ-PAS-002/-003/-004/-005/-006/-007 at the core level; AC-PAS-001..-007 (the core-level halves; AC-PAS-005, -007 and -008 also have real-adapter halves in M3).

### M3 — Adapters, gate wiring, fallbacks, summary (Priority High)
Start with the B.3 measurement and record its output in `progress.md` §E.2. Then: git adapter, `go list` adapter, wiring behind the pre-commit flag, fallback notices, summary lines (selected / dropped / widened). Covers REQ-PAS-001/-005/-008/-009/-010/-014; AC-PAS-005 (adapter half), -007, -008, -009, -010, -014, -016.

### M4 — WITHDRAWN (slot retained so M5/M6 numbering stays stable)
The deferral mechanism was withdrawn by the Q1 ruling (board d-20261018T005931Z-f264); REQ-PAS-012 and AC-PAS-012 are withdrawn with it. Nothing is implemented under this milestone.

### M5 — Suppression-check fix (Priority Medium; mechanical)
RED: a test that builds a temporary project holding a copy of the Go fixture under `.../testdata/...` plus a production file with a bare marker; expects only the production violation. GREEN: walk change (B.6). Covers REQ-PAS-013; AC-PAS-013.

### M6 — Verification sweep and dogfood (Priority High)
Run the scoped test set (§D), the preservation guards (AC-PAS-014, -015), and the dogfood step: feed this SPEC's own diff to the new resolver and compare against a hand `go list` computation. Record all verbatim output in `progress.md` §E.2.

## D. Verification plan under the leader's constraints

Affected packages of this SPEC's own change (expected, to be re-derived by the resolver at M6): `internal/hook/quality`, `internal/config`, `internal/cli` (if the CLI mapping is edited), plus direct dependents. Observed on tree 3642a3a9b: `internal/hook/quality` is imported by `internal/cli` and `internal/hook`; `internal/config` has 47 direct dependents (exact import-path match, itself excluded), 46 with tests, the one without (`internal/closure/closuretest`) is dropped with reason `no test files (ls)` (ruling 148c).

| Package | How it is run |
|---------|---------------|
| `internal/hook/quality` | full package, `go test ./internal/hook/quality -count=1` (not barred) |
| `internal/config` | full package |
| `internal/cli` | `-run` selection only, ruling 35e9: names found with `go test -list` against the gate/pre-commit/update-preserve tests (`TestGate*`, `TestPreCommit*`, `TestPrecommit*`, the `update_gate_yaml_preserve` test) — the exact regex is fixed and recorded at M3 |
| other direct dependents of `internal/config` | the lane runs the dependents list produced by the resolver, one package per `go test` invocation, after `ls <dir>/*_test.go` confirms test files exist; no package is run twice |

The importer drop is a recorded fact (`ls` output), not a skip: a package with no test files has an empty swept set.

## E. Handoff — hooks re-enablement runbook (documentation; no milestone executes it)

Status (Q5 ruling, board d-20261018T005931Z-f264): hooks stay OFF in this card. The actual re-enable is a post-landing operator decision among three options: (1) unset the shared `core.hooksPath`; (2) per-worktree configuration; (3) keep hooks off (rehearsal only through the per-invocation override). The steps below serve options 1 and 2.

Audience: factory leader / operator, after this SPEC has landed on local `main` and the rc binary built from it is installed on PATH (the hook resolves `moai` through PATH; a bare `cp` over the binary produced SIGKILL 137 before — use the clean reinstall in the local lane protocol §9).

Pre-conditions (all observed, none assumed):
1. `moai version` prints the build that contains this SPEC's commits (compare to `git rev-parse --short HEAD` of local `main`).
2. AC-PAS-013's real-tool command exits 0 on the landed tree.
3. Q4 measured: run the heavy gate's non-test steps once, outside any hook, under a resource slot lease (`moai slot acquire --resource <name> --max-duration <limit>`), and compare each step's duration to its budget (30 / 60 / 300 s). If any exceeds its budget, stop: re-enabling would block commits on timeouts and the budget question goes back to the operator (this SPEC forbids enlarging them). Also measure, under the same slot lease, the affected-set test time of a few representative commits (for example a change to `internal/config`, whose 46 tested dependents share one 120 s budget in a single `go test` call, and one to `internal/hook/quality`, which selects `internal/hook` and `internal/cli`). The claim that these fit or overrun the budget is inferred in this plan, never measured; `internal/cli` is measured only by `-run` selection (ruling 35e9).
4. `gate.yaml` `pre_commit.enabled` value for this repository is confirmed by the operator (local copy is `true`, template is `false`, F4).
5. Decision-index Q1 and Q5 carry operator verdicts.

Rehearsal without touching shared configuration: in a card worktree, make a one-package change and commit it with the per-invocation override `git -c core.hooksPath=<absolute path of the shared hooks directory> commit ...`. The override applies to that one command only. Observe the gate summary lists the affected packages and that a failing affected test blocks the commit. Repeat with a docs-only change (expected: explicit skip of the test step, commit passes).

The switch (leader/operator action, affects every session — announce first):
1. Record the current value: `git config --show-origin --get core.hooksPath`.
2. Unset it in the shared repository configuration: `git config --unset core.hooksPath` (run so that it targets `/Users/goos/MoAI/moai-adk-go/.git/config`). Alternative with narrower blast radius, which itself enables an extension in the shared configuration and so needs the same announcement: per-worktree configuration.
3. Side effect to expect: the shared hooks directory also holds a `pre-push` hook that prints a warning and sleeps 5 seconds on pushes to `main`; it becomes live with the switch.
4. Verify: `git config --get core.hooksPath` exits 1 (unset); in a throwaway worktree branch, `git commit --allow-empty -m "hook probe"` prints the `moai gate` pre-commit output.

Rollback (single command, reversible): `git config core.hooksPath /dev/null`.

## F. Risks

| ID | Risk | Mitigation |
|----|------|------------|
| R1 | `git commit -a` / path commits: the changed set read from the real index under-scopes the commit (B.3). | M3 measures both shapes first; AC-PAS-009 is release-blocking. |
| R2 | `internal/cli` appears in the affected set of many commits (direct dependent of `internal/hook/quality`; dependent of most packages through its own imports) and its suite is inferred (1022 test files, ruling 35e9; not measured) to exceed the unchanged 120 s budget, as may `internal/hook` (296) and a wide `internal/config` selection summed into one call. A commit that selects such a set fails by budget overrun. | Accepted by ruling Q1=(B): no deferral or skip mechanism exists and the 120 s budget is unchanged. A pre-commit `-run` selection for `internal/cli` (as in ruling 35e9) is a follow-up card recorded by the leader, outside this SPEC. |
| R3 | Tests that read repository files by path (not embed, not testdata) are invisible to package analysis. | Accepted; CI is the net (AGENTS.md §4). Recorded as residual risk. |
| R4 | A package whose only change is a build-constraint-excluded file on this platform. | Ownership counts ignored-Go files (B.2 step 2). |
| R5 | `go list` cost on a 176-package module is unmeasured as wall time (Gap G2). | Executed once per gate run, not per package; its failure falls back (REQ-PAS-008). Measure at M3 and record; no timing assertion in tests. |
| R6 | A widening trigger (REQ-PAS-007) re-selects the full module, which on this repository exceeds the budget. | Q2 default keeps today's behaviour for that rare commit; the summary says why. |
| R7 | Template twin drift between the two `gate.yaml` copies. | AC-PAS-011 compares the new block byte-for-byte. |
