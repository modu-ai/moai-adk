---
id: SPEC-TRUNK-BACKUP-SYNC-001
title: "Acceptance criteria — trunk backup sync verb + workflow-exclusion regression guard"
version: "0.1.0"
created: 2026-10-17
updated: 2026-10-17
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, .github/workflows, .moai/docs"
lifecycle: spec-anchored
tags: "trunk-backup, sync-ref, git-push, workflow-triggers, regression-guard"
---

# acceptance.md — SPEC-TRUNK-BACKUP-SYNC-001

> Verification layer. Every entry is Given-When-Then and binary-testable. GEARS obligations live in `spec.md` §D; they are not restated here. Stateless on the status axis.

## §A — Traceability matrix

| AC | Requirement | Milestone | Severity | Class |
|---|---|---|---|---|
| AC-TBS-001 | REQ-TBS-005 | M1 | Must | regression-guard (live sweep) |
| AC-TBS-002 | REQ-TBS-005 | M1 | Must | release-blocking (mutant RED) |
| AC-TBS-003 | REQ-TBS-001 | M2 | Must | release-blocking |
| AC-TBS-004 | REQ-TBS-001, REQ-TBS-002 | M2 | Must | release-blocking |
| AC-TBS-005 | REQ-TBS-003 | M2 | Must | release-blocking |
| AC-TBS-006 | REQ-TBS-006, REQ-TBS-008 | M2/M3 | Must | release-blocking |
| AC-TBS-007 | REQ-TBS-001 | M2 | Must | release-blocking |
| AC-TBS-008 | REQ-TBS-007, REQ-TBS-009 | M3 | Should | release-blocking |
| AC-TBS-009 | REQ-TBS-004 | M2 | Must | release-blocking |

## §B — Acceptance criteria

- **AC-TBS-001 — live workflow sweep**. Given the repository's `.github/workflows/` at the tested commit, when the regression test enumerates every `*.yml`/`*.yaml` file and evaluates each `on.push` trigger's branch filter against the ref name `sync/main`, then every push trigger excludes `sync/main`, and the test asserts it found at least one workflow file carrying a push trigger (empty-sweep guard: a zero-count sweep fails the test, never passes it).
- **AC-TBS-002 — mutant detection (observed-failure evidence)**. Given the in-test mutant fixtures — a bare `on: push` block with no branch filter, and a `branches: ['sync/**']` list — when the same matcher evaluates them, then each fixture is flagged as admitting `sync/main`. These cases are the guard's red-on-known-failing-input record (verification-completeness §1.1); a matcher that passes the live sweep but misses a mutant fails this criterion.
- **AC-TBS-003 — first backup push against a temp remote**. Given a temporary repository whose `origin` is a bare repository under `t.TempDir()`, with `refs/heads/main` at commit SHA `X` and no `sync/main` on the remote, when the verb runs, then the remote gains `refs/heads/sync/main` at `X`, no other remote ref changes, and the verb's output names `X` and the target ref.
- **AC-TBS-004 — force scoping and the non-FF case**. Given the temp remote's `refs/heads/sync/main` at a commit that is not an ancestor of local `main` (the rewound-main shape), when the verb runs, then the push succeeds and updates `sync/main`; and given the test's captured record of every git command the verb issued across all runs, then `--force` (or `-f`) appears only on the exact constant refspec `refs/heads/main:refs/heads/sync/main`, and no captured command updates any other remote ref.
- **AC-TBS-005 — refusal on missing main**. Given a temporary repository in which `refs/heads/main` does not resolve, when the verb runs, then it exits non-zero, prints guidance naming the missing ref and the constant target, and its captured command record shows no push command was executed.
- **AC-TBS-006 — no-real-push guard**. Given the test suite's captured verb commands and the run-phase evidence, when they are inspected, then every remote operand resolves under a `t.TempDir()` root, no test names the development repository's configured remotes or any network URL, and the run phase as a whole contacted no real remote and opened no pull request (REQ-TBS-008).
- **AC-TBS-007 — environment scrub (t1478 shape)**. Given a process environment carrying inherited `GIT_DIR`/`GIT_WORK_TREE` pointing at a foreign temporary repository, when the verb runs, then the push targets the caller's repository (the temp clone the test constructed), proving repository-location variables were removed before exec.
- **AC-TBS-008 — runbook and template isolation**. Given the card's final state, then `.moai/docs/trunk-backup-runbook.md` exists and carries the four REQ-TBS-007 statements (leader-round step; operator-approved origin-push scope; redundancy-copy-only status; unchanged batch discipline and main protection), and no file under `internal/template/templates/` and no branch-protection or guard configuration changed.
- **AC-TBS-009 — push-failure reporting**. Given a temporary remote configured to reject the ref update (a failing pre-receive hook), when the verb runs, then it exits non-zero, prints both the stdout and the stderr of the failed push command, and its captured command record shows no git command was issued after the failed push.

## §C — Edge cases (covered by the tests or explicitly excluded)

- `release.yml` carries `tags: ['v*']` with no `branches` filter — the matcher must classify it as not firing on branch pushes (AC-TBS-001 covers via the live sweep; a `tags`-only fixture pins the semantics).
- A workflow file with `on: [push, pull_request]` (list form) — the parser must read the push trigger from the list form.
- A push trigger with `branches-ignore` only — under GitHub semantics it fires on every branch not ignored; the matcher must treat `sync/main` as admitted unless an ignore pattern matches it. (No current workflow uses this form; a mutant fixture pins the semantics.)
- Local `main` is a symbolic or ambiguous ref — REQ-TBS-003's refusal path.
- Verb invoked from a card worktree rather than the primary checkout — REQ-TBS-001 resolves the shared repository; the temp-remote tests cover the shape.

## §D — Quality gate criteria

- `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- `go test` for the new test files passes with the scoped run recorded verbatim.
- `golangci-lint` reports no NEW findings on the added files (baseline split recorded).
- `moai spec lint SPEC-TRUNK-BACKUP-SYNC-001` reports no error-severity findings on the SPEC artifacts.

## §E — Adoption classification (two-cell discipline)

- **Red-at-arrival (release-blocking RED-now)**: AC-TBS-003..005, AC-TBS-007, AC-TBS-009 — the verb does not exist at plan time; the first RED run of each milestone-2 test is the starting observation, captured verbatim per E8.
- **Regression-guard (undecidable-disposition class)**: AC-TBS-001's live sweep — the current tree passes it by design; it guards a currently-true invariant. Its failure evidence is AC-TBS-002's mutant fixtures, which are executable now and at every future run. The pair (001, 002) is the two-cell adoption for the guard.
- **Green path**: M1 flips AC-TBS-001/002; M2 flips AC-TBS-003..005, 007, and 009; M3 flips AC-TBS-006/008.

## §F — Definition of Done

- All nine criteria PASS with attributable evidence (command + verbatim output + tree SHA per run).
- No push to any real remote occurred at any point in the run phase (§ AC-TBS-006 evidence).
- `progress.md` §E.2 carries the run-phase evidence; §E.3 the audit-ready signal.
- The four decision-index rows remain untouched unless the operator records verdicts mid-run; a verdict landing mid-run is applied as the one-commit amendment described in `spec.md` §E.
