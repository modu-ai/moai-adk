---
id: SPEC-PRECOMMIT-AFFECTED-SCOPE-001
title: "Pre-commit gate tests only the affected Go packages; ast-grep suppression check stops flagging test fixtures"
version: "0.1.0"
status: draft
created: 2026-10-18
updated: 2026-10-18
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/hook/quality, internal/cli, internal/config"
lifecycle: spec-anchored
tags: "pre-commit, quality-gate, affected-packages, go-test, ast-grep, hooks"
tier: M
related_specs: [SPEC-PRECOMMIT-GATE-SCOPE-001, SPEC-PRETOOL-GATE-MOVE-001, SPEC-GATE-ASTGREP-REPAIR-001, SPEC-UTIL-002]
---

# SPEC-PRECOMMIT-AFFECTED-SCOPE-001

## HISTORY

| Date | Author | Change |
|------|--------|--------|
| 2026-10-18 | manager-spec | Plan-phase first draft (card t1673, Class C, Tier M). Findings re-observed on tree 3642a3a9b; see § A. |

## A. Background

### A.1 What this SPEC changes (WHAT)

When the git pre-commit hook runs the heavy gate (`gate.pre_commit.enabled: true`), the Go test step must test the packages the commit can affect — the packages that own a changed file plus their direct dependents — instead of every package in the module. A second, independent defect that blocks any `moai gate` run on this repository (the ast-grep suppression-policy check flagging a test fixture) is fixed in the same SPEC because it is the base defect the card names as blocking hook re-enablement.

### A.2 Why (WHY)

`AGENTS.md` §4 says verification is scoped to the change and the full suite is CI's job. The distributed pre-commit gate does the opposite: its Go test step is `go test ./...` on every commit (observed: `internal/hook/quality/gate.go`, the `toolchains` table, Go entry, `testStep` with args `test`, `./...`). On this repository the `internal/cli` package alone carries 1022 `_test.go` files (observed from `go list` on tree 3642a3a9b), and board ruling 35e9 bars full `internal/cli` suite runs. The unchanged 120 s test budget therefore cannot hold the full-suite step, which is one reason the shared git configuration currently has every hook switched off.

### A.3 Observed facts (tree 3642a3a9b9cf101863656153d2d99b69fcb66a05)

| # | Fact | How observed |
|---|------|--------------|
| F1 | Go test step is `go test ./...`, run once per detected Go module root, with the 120 s test budget. | read `gate.go`; `gate.yaml` `timeouts.test: 120` |
| F2 | The step has a staged-file query (`git diff --cached --name-only`, scrubbed of `GIT_*` location variables) used today only to skip steps by extension. It returns nil for both "nothing staged" and "query failed". | read `gate.go` `stagedFiles` |
| F3 | The pre-commit hook exports `MOAI_PRECOMMIT=1` and runs `moai gate`; the runner branches on that marker and on `gate.pre_commit.enabled`. The hook script is byte-identical to its template twin (`diff` printed no difference). | read `internal/cli/gate.go`; `.git/hooks/pre-commit` vs `internal/template/templates/.git_hooks/pre-commit` |
| F4 | `gate.yaml` in this repository has `pre_commit.enabled: true`; the template copy has `false`. | `diff` of the two files |
| F5 | `git config --show-origin --get core.hooksPath` prints `file:/Users/goos/MoAI/moai-adk-go/.git/config	/dev/null` (exit 0): the switch lives in the shared repository config, so it affects every worktree and session. | executed |
| F6 | The module has 176 packages. `internal/config` has 48 direct dependents (47 with tests); `internal/hook/quality` has 2 (`internal/cli`: 1022 test files, `internal/hook`: 296). | `go list -e -json=...` on the tree, `jq` |
| F7 | The base defect is real, not hypothetical: `moai gate` built from this tree, run against a project holding only the `suppressed.go` fixture, with every other step disabled and the ast-grep gate configured advisory (`block_on_error: false`, `warn_only_mode: true`), exits 1 with `ast-grep suppression at .../internal/astgrep/testdata/fixtures/go/suppressed.go:15 requires adjacent '// @MX:REASON <rationale>' on next line`. Verbatim output: `acceptance.md` LEDGER-PAS-013-A. | executed |
| F8 | Cause of F7: the suppression-pairing check is a pure-Go step that ignores `block_on_error` / `warn_only_mode`, requires `@MX:REASON` on the line AFTER `// ast-grep-ignore`, and its file walk skips `_test.go` but not `testdata`. The Go fixture, added by 59f201e85, puts `@MX:REASON` BEFORE the marker because ast-grep suppresses the line immediately following the marker. | read `astgrep_gate.go`; `sg scan` on the fixture prints `[]`; on a variant with the marker first it reports rule `go-interface-empty-not-any` |
| F9 | The five other `suppressed.*` fixtures (kotlin, csharp, ruby, elixir, php) already follow marker-then-REASON; php has no comment prefix in the checker. The Go fixture is the only offender found by grep outside `_test.go`. | grep over the tree (a full-tree run of the real checker was not performed) |

### A.4 Existing coverage this SPEC does not repeat

`SPEC-PRECOMMIT-GATE-SCOPE-001` (completed) made the heavy gate opt-in under `MOAI_PRECOMMIT=1` and wired `gate.pre_commit.enabled`. This SPEC narrows what the opted-in gate TESTS; it keeps that marker branch and that key as they are.

## B. Requirements (GEARS)

### REQ-PAS-001 — Affected-only test scope (Event-driven)

When the Go toolchain's test step runs in the pre-commit context with the heavy gate opted in and `gate.pre_commit.test_scope.mode` resolving to `affected`, the quality gate shall test only the affected package set rather than every package in the module.

### REQ-PAS-002 — Definition of the affected set (Ubiquitous)

The affected package set shall be the union of (a) every package that owns a changed file and (b) every package that directly imports a package in (a) through its regular, in-package-test, or external-test imports; packages that depend on (a) only transitively shall not be included. The set shall be a pure function of the changed-file list, the package import graph, and the module root — it shall not read the wall clock or the process environment.

### REQ-PAS-003 — Ownership of non-Go files (Ubiquitous)

The owner of a changed non-`.go` file shall be the package whose embedded-file set contains it, and otherwise the nearest ancestor directory that is a Go package directory (a `testdata` file belongs to the package whose directory contains that `testdata`). A file with no owner shall contribute no package and shall be named in the gate's summary as owned by none.

### REQ-PAS-004 — Test-only changes do not widen (Event-driven)

When every changed file owned by a package is a `_test.go` file or lies under that package's `testdata`, the gate shall include that package and shall not include its dependents on account of those files.

### REQ-PAS-005 — Deleted and renamed files (Event-driven)

When a commit deletes or renames a Go file, the gate shall attribute the file to the package of its former directory and shall select that package's importers by import path, whether or not the former package still exists on disk.

### REQ-PAS-006 — No-test packages are dropped and counted (Ubiquitous)

The gate shall drop selected packages that have no test files, shall report in its summary how many packages were selected, how many were dropped, and the drop reason `no test files` for each dropped package, and, when the selection is empty after dropping, shall record the test step as skipped with an explicit reason rather than as a pass.

### REQ-PAS-007 — Dependency-graph changes widen to the full scope (Event-driven)

When the commit changes `go.mod`, `go.sum`, `go.work`, `go.work.sum`, or `.moai/config/build-tags`, the gate shall run the full-module test scope and shall state that widening and its trigger in the summary.

### REQ-PAS-008 — Failure falls back, never to a pass (Event-driven)

When the changed-file query or the package-graph query fails, the gate shall fall back to the full-module test scope with a notice naming the failed query, and shall not report a pass derived from the failed computation.

### REQ-PAS-009 — Changed set is what the commit records (Ubiquitous)

The changed-file set the gate uses shall be the set the commit will record, including commits made with `git commit -a` and with explicit paths.

### REQ-PAS-010 — Context isolation (Ubiquitous)

Outside the pre-commit context — the standalone `moai gate`, the PreToolUse path, and the SubagentStop and TeammateIdle paths — and for every non-Go toolchain, the gate's test scope shall remain as it is today; the vet, lint, typecheck, graph-freshness and ast-grep steps shall be unchanged.

### REQ-PAS-011 — Scope setting (Capability gate)

Where `gate.pre_commit.test_scope.mode` is present in `gate.yaml`, the gate shall accept `affected` or `full`, shall default to `affected` when the key is absent, and shall reject any other value with a message naming the key and the accepted values; `full` restores today's `./...` scope in the pre-commit context.

### REQ-PAS-012 — Visible deferral of over-budget packages (Capability gate) — PROVISIONAL, gated by decision-index Q1

Where `gate.pre_commit.test_scope.deferred_packages` lists import-path patterns, the gate shall leave packages matching them out of the executed test set and shall name each one in the summary as deferred to CI; the key shall default to empty so that, absent configuration, nothing is deferred. If decision-index Q1 resolves against the mechanism, this requirement, AC-PAS-012 and plan.md M4 are deleted.

### REQ-PAS-013 — Suppression-policy check ignores fixtures (Unwanted)

The ast-grep suppression-policy check shall not examine files under any `testdata` path segment, and a gate run on this repository shall not report `internal/astgrep/testdata/fixtures/go/suppressed.go`; a suppression marker in production source that lacks an adjacent `@MX:REASON` shall still be reported.

### REQ-PAS-014 — No timeout enlargement, no skips (Unwanted)

The change shall not enlarge any gate timeout, shall not add a test skip (`t.Skip*` or an equivalent), and shall not remove any existing test.

### REQ-PAS-015 — Hooks stay off; no git-configuration writes (Unwanted)

The delivered code and configuration shall not write `core.hooksPath` or any other git configuration key, and the repository's hook switch shall remain as it is when this SPEC lands; re-enabling hooks is the post-landing handoff in § C.

## C. Exclusions

### Out of Scope — Re-enabling git hooks (handoff, not an implementation step)

- Switching hooks back on changes the shared repository configuration (`core.hooksPath` in `/Users/goos/MoAI/moai-adk-go/.git/config`, observed F5) and affects every worktree and every concurrent session. This SPEC delivers code and configuration defaults only; the switch is a post-landing leader/operator decision.
- The runbook for that decision is `plan.md` § Handoff — hooks re-enablement runbook. It is documentation; no milestone of this SPEC executes it.
- Until the switch, lanes keep committing with hooks not running (decision t1663 (A)), without `SKIP_MOAI_PRECOMMIT` and without `--no-verify`, recording one affected-package test run in the card's progress record.

### Out of Scope — Other gate steps and other toolchains

- Project-wide `go vet ./...`, `golangci-lint run`, the typecheck axis, the graph-freshness step and the ast-grep rule scan in the heavy gate are not scoped down here. Their runtime on this repository is unmeasured (decision-index Q4).
- Non-Go toolchains (node, python, rust, java and the rest of the `toolchains` table) keep their current test commands.
- The hook shell script and its Go raw-string twin are not edited; scoping happens inside `moai gate`.

### Out of Scope — Test-selection inside a package

- Choosing individual tests by name (`-run`) is not done by the gate. It remains the operator and lane technique under ruling 35e9 for `internal/cli`.
- Test files that read repository files by relative path are invisible to package-graph analysis (residual risk R3 in plan.md); CI remains the net.

### Out of Scope — Policy semantics of the suppression check

- The marker-then-REASON adjacency rule (SPEC-UTIL-002) is not redefined. Only the set of files the check examines changes.
- The Go fixture's own line ordering is not changed: the order is what makes the ast-grep scan return zero findings.
