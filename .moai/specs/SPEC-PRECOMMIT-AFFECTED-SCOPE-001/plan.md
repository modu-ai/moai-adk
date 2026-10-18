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

Inputs: (1) changed files, repo-relative; (2) the package graph of the toolchain's module root, from one `go list -e -json` query over `./...` (fields: import path, directory, Go/ignored/test file lists, the three import lists, embed patterns); (3) the module root's offset inside the repository.

Steps:
1. Drop changed files outside the module root. If a widening trigger file is present (REQ-PAS-007) return the full scope and the trigger name.
2. Owner of each file: a `.go` file (or a file listed in the package's Go, ignored-Go, cgo or test file lists) belongs to its directory's package, including files excluded by build constraints on this platform; any other file follows REQ-PAS-003 — embedded-set membership first (pattern roots resolved relative to the package directory), then nearest ancestor package directory; `testdata` resolves through the same ancestor walk.
3. For deleted or renamed paths use the former directory's import path directly (REQ-PAS-005); no on-disk package is needed to look up importers.
4. Direct dependents: for each package whose files were changed other than test-only (REQ-PAS-004), add every package whose regular, test or external-test import list contains its import path.
5. Drop packages with no test files (in-package or external); record the counts (REQ-PAS-006).
6. Hand the remaining import paths to the existing step runner as the argument list, preserving the 120 s budget, the process-group termination and the stdout/stderr merging unchanged.

The core (steps 2–5) takes the changed list and a plain in-memory graph and returns the selection plus the report lines; nothing in it spawns a process or reads time or environment. Two thin adapters feed it: the git adapter (changed files, see B.3) and the `go list` adapter. Tests inject both (B.5).

### B.3 Commit-accurate changed set (REQ-PAS-009) — measure first

The existing staged-file helper scrubs `GIT_INDEX_FILE` and `GIT_DIR` from its environment on purpose (card t560: steps run inside a hook must not address the caller's repository). For a plain `git commit` the real index equals the commit's content, so the scrubbed query is right. For `git commit -a` and `git commit <paths>`, git runs the hook against a temporary index; a query against the real index can then miss the files the commit will record, and an under-scoped set would pass a commit whose tests were never run. This has not been measured here (Gap G1). Milestone M3 starts by measuring both commit shapes in a temporary repository with a real hook, then picks the narrowest mechanism that makes AC-PAS-009 pass — candidate: read the hook-provided index location for this one query only, leaving every step's environment scrubbed. The helper must also distinguish "nothing staged" (empty selection, explicit skip) from "query failed" (REQ-PAS-008 fallback); today both return nil.

### B.4 Configuration

New keys under `gate.pre_commit` in `gate.yaml`: `test_scope.mode` (`affected` default, `full`). No package-exclusion key exists (Q1 ruled B). Touch points: config types, defaults, validation (reject unknown mode), the CLI mapping into the gate configuration, the local `gate.yaml`, and its template twin (`internal/template/templates/.moai/config/sections/gate.yaml`; the two files differ today only in `pre_commit.enabled` and `rules_dir`, F4 — keep those differences). `moai update` already preserves user edits to `gate.yaml` (covered by `update_gate_yaml_preserve_test.go`); the new keys must not break that behaviour. The `moai web` gate panel is not extended (Out of Scope: nothing requires it).

### B.5 Deterministic test strategy

- Core: table-driven tests over hand-written in-memory graphs and changed lists. No `go list`, no git, no clock. Covers REQ-PAS-002/-003/-004/-005/-006/-007.
- Adapters: one test builds a throwaway Go module in `t.TempDir()` (packages `a`, `b` importing `a`, `c` importing `b`, `d` standalone, one with an external test, one with `testdata`, one embedding a file), runs the real `go list` adapter, and compares to the expected import-path set. `go` is necessarily present whenever `go test` runs, so no skip guard is added (REQ-PAS-014).
- Git: tests create a temporary repository, install a real `pre-commit` script that records its view of the changed set, and commit with `git commit`, `-a`, and explicit paths (REQ-PAS-009). They call git with a scrubbed environment as the existing `gate_staged_files_git_env_test.go` pattern does.
- Gate wiring: a fake `go` binary on PATH (pattern already used in `gate_test.go`, `writeFakeBinary`) records its argv, proving the exact argument list for each scenario without running real tests.
- Nothing asserts elapsed time; the existing step timeouts are untouched.

### B.6 Suppression-policy fix (REQ-PAS-013)

The file walk that feeds the suppression check skips any path with a `testdata` segment, alongside the segments it already skips (`vendor`, `.git`, `node_modules`, `__pycache__`) and `_test.go` files. This matches the stated rationale of that walk ("suppression policy checks apply to production code") and the sibling path filter used for ast-grep findings, which already excludes `/testdata/`. Alternatives rejected with evidence: reordering the Go fixture (observed: a marker-first variant makes `sg scan` report `go-interface-empty-not-any`, so `TestRuleSeed` `fixtures-scan` would fail); letting the checker accept REASON before the marker (redefines SPEC-UTIL-002's policy, wider than the defect — decision-index Q3).

## C. Milestones

### M1 — Configuration surface and template twin (Priority High; most likely to change)
RED: config tests for default (`affected`), valid values, invalid-value rejection, and byte-level agreement of the two `gate.yaml` copies on the new block. GREEN: types, defaults, validation, loader mapping, both YAML files. Covers REQ-PAS-011; AC-PAS-011.

### M2 — Affected-set core (Priority High)
RED→GREEN over the injected-graph tables. Covers REQ-PAS-002/-003/-004/-005/-006/-007 at the core level; AC-PAS-001..-007.

### M3 — Adapters, gate wiring, fallbacks, summary (Priority High)
Start with the B.3 measurement and record its output in `progress.md` §E.2. Then: git adapter, `go list` adapter, wiring behind the pre-commit flag, fallback notices, summary lines (selected / dropped / widened). Covers REQ-PAS-001/-008/-009/-010; AC-PAS-008, -009, -010, -016.

### M4 — WITHDRAWN (slot retained so M5/M6 numbering stays stable)
The deferral mechanism was withdrawn by the Q1 ruling (board d-20261018T005931Z-f264); REQ-PAS-012 and AC-PAS-012 are withdrawn with it. Nothing is implemented under this milestone.

### M5 — Suppression-check fix (Priority Medium; mechanical)
RED: a test that builds a temporary project holding a copy of the Go fixture under `.../testdata/...` plus a production file with a bare marker; expects only the production violation. GREEN: walk change (B.6). Covers REQ-PAS-013; AC-PAS-013.

### M6 — Verification sweep and dogfood (Priority High)
Run the scoped test set (§D), the preservation guards (AC-PAS-014, -015), and the dogfood step: feed this SPEC's own diff to the new resolver and compare against a hand `go list` computation. Record all verbatim output in `progress.md` §E.2.

## D. Verification plan under the leader's constraints

Affected packages of this SPEC's own change (expected, to be re-derived by the resolver at M6): `internal/hook/quality`, `internal/config`, `internal/cli` (if the CLI mapping is edited), plus direct dependents. Observed on tree 3642a3a9b: `internal/hook/quality` is imported by `internal/cli` and `internal/hook`; `internal/config` has 48 direct dependents, 47 with tests, the one without (`internal/closure/closuretest`) is dropped with reason `no test files (ls)` (ruling 148c).

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
3. Q4 measured: run the heavy gate's non-test steps once, outside any hook, under a resource slot lease (`moai slot acquire --resource <name> --max-duration <limit>`), and compare each step's duration to its budget (30 / 60 / 300 s). If any exceeds its budget, stop: re-enabling would block commits on timeouts and the budget question goes back to the operator (this SPEC forbids enlarging them).
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
| R2 | `internal/cli` appears in the affected set of many commits (direct dependent of `internal/hook/quality`; dependent of most packages through its own imports) and its suite alone exceeds the unchanged 120 s budget. A commit that selects it fails by budget overrun. | Accepted by ruling Q1=(B): no deferral or skip mechanism exists and the 120 s budget is unchanged. A pre-commit `-run` selection for `internal/cli` (as in ruling 35e9) is a follow-up card recorded by the leader, outside this SPEC. |
| R3 | Tests that read repository files by path (not embed, not testdata) are invisible to package analysis. | Accepted; CI is the net (AGENTS.md §4). Recorded as residual risk. |
| R4 | A package whose only change is a build-constraint-excluded file on this platform. | Ownership counts ignored-Go files (B.2 step 2). |
| R5 | `go list` cost on a 176-package module is unmeasured as wall time (Gap G2). | Executed once per gate run, not per package; its failure falls back (REQ-PAS-008). Measure at M3 and record; no timing assertion in tests. |
| R6 | A widening trigger (REQ-PAS-007) re-selects the full module, which on this repository exceeds the budget. | Q2 default keeps today's behaviour for that rare commit; the summary says why. |
| R7 | Template twin drift between the two `gate.yaml` copies. | AC-PAS-011 compares the new block byte-for-byte. |
