---
id: SPEC-WORKTREE-SWEEP-001
title: "Acceptance criteria — worktree post-landing auto-sweep"
version: "0.1.0"
created: 2026-09-30
updated: 2026-09-30
author: manager-spec
---

# acceptance.md — SPEC-WORKTREE-SWEEP-001

Every criterion is binary-testable against a fixture repository built with the package's existing test helpers. The dry-run/apply distinction is exercised through the command's RunE via the cobra test harness used in `worktree/*_test.go`.

## §D — AC Matrix

| AC | Requirement(s) | Scenario | Machine-verifiable command (shape) |
|----|----------------|----------|------------------------------------|
| AC-WS-001 | REQ-WS-001, 002 | Landed branch | `go test -run TestSweepLandedBranchDisposes ./internal/cli/worktree/` |
| AC-WS-002 | REQ-WS-001, 003 | Not landed (exit 1) | `go test -run TestSweepNotLandedPreserves ./internal/cli/worktree/` |
| AC-WS-003 | REQ-WS-003 | Unanswerable ancestry | `go test -run TestSweepAncestryUnanswerable ./internal/cli/worktree/` |
| AC-WS-004 | REQ-WS-005 | Dirty tree on landed branch | `go test -run TestSweepDirtyPreserves ./internal/cli/worktree/` |
| AC-WS-005 | REQ-WS-005, 007 | Unique unpushed commits | `go test -run TestSweepUnpushedPreserves ./internal/cli/worktree/` |
| AC-WS-006 | REQ-WS-006 | Locked tree / anchor decision | `go test -run TestSweepLockedAndAnchored ./internal/cli/worktree/` |
| AC-WS-007 | REQ-WS-006 | Live cwd probe hit + unanswerable | `go test -run TestSweepProcessCWDPredicate ./internal/cli/worktree/` |
| AC-WS-008 | REQ-WS-012 | Dry-run default / --yes / --json parity | `go test -run TestSweepDryRunDefaultAndYes ./internal/cli/worktree/` |
| AC-WS-009 | REQ-WS-009 | L1 routing: hoist before remove; hoist failure preserves | `go test -run TestSweepL1HoistThenRemove ./internal/cli/worktree/` |
| AC-WS-010 | REQ-WS-008 | L2 routing via done core | `go test -run TestSweepL2DonePath ./internal/cli/worktree/` |
| AC-WS-011 | REQ-WS-007 | Never-dispose list | `go test -run TestSweepNeverDispose ./internal/cli/worktree/` |
| AC-WS-012 | REQ-WS-004 | Base default and --base override | `go test -run TestSweepBaseFlag ./internal/cli/worktree/` |
| AC-WS-013 | REQ-WS-010, 011 | Verdict record completeness; branch survives apply | `go test -run TestSweepVerdictRecord ./internal/cli/worktree/` |
| AC-WS-014 | REQ-WS-013, 014 | Removal-failure resilience; doctrine step present; budget delta vs the progress.md §E.1 run-start measurement | `go test -run TestSweepRemovalFailure ./internal/cli/worktree/` + `wc -c` + `grep` below |

## §D.1 — Scenario detail (Given-When-Then)

### AC-WS-001 — Landed branch disposes
- **Given** a fixture repo with base `origin/develop` and a worktree whose branch tip is an ancestor of the (fetched) base ref, tree clean, unlocked, unanchored, probe-negative.
- **When** the sweep runs with `--yes`.
- **Then** the verdict record reads `verdict=DISPOSE`, the tree directory is gone, the branch ref still resolves (`git rev-parse --verify <branch>` exit 0 — REQ-WS-010), and every predicate field reads affirmatively (no `undetermined`).

### AC-WS-002 — Not landed preserves
- **Given** a worktree whose branch tip is NOT an ancestor of the fetched base ref (exit 1).
- **When** the sweep runs (either mode).
- **Then** the record reads `verdict=PRESERVE`, `reason` contains `cause=not-landed`, and the tree still exists.

### AC-WS-003 — Unanswerable ancestry preserves (three-way contract)
- **Given** (a) a fetch that fails (seam-injected), (b) an ancestry exit of 2, and (c) an unresolvable base ref — one fixture cell each.
- **When** the sweep runs.
- **Then** each cell reads `verdict=PRESERVE` with `cause=fetch-failed` / `cause=landed-check-failed` / `cause=landed-check-failed` respectively; no cell reads DISPOSE; no cell reads PRESERVE-for-not-landed (the "no answer is not a no" distinction).

### AC-WS-004 — Dirty tree preserves
- **Given** a landed, unanchored worktree with one untracked file.
- **When** the sweep runs with `--yes`.
- **Then** `verdict=PRESERVE`, `dirty=yes`, tree survives, file intact.

### AC-WS-005 — Unpushed commits preserve
- **Given** a worktree whose branch carries commits unreachable from any remote (`gitHasUnpushedReal` semantics).
- **When** the sweep runs.
- **Then** `verdict=PRESERVE` (via the landing predicate's exit-1 path), tree survives.

### AC-WS-006 — Locked and anchored preserve
- **Given** (a) a tree with a git worktree lock, (b) a tree named by the session registry anchor, (c) a run where the lock source is unreadable (seam).
- **When** the sweep runs.
- **Then** (a) and (b) preserve with the anchor source named; (c) preserves every tree with `anchored=undetermined` and the process exits 2 after the full report (the `clean --stale` degraded-run contract).

### AC-WS-007 — Process-cwd predicate
- **Given** (a) a seam cwd list containing the tree path, (b) a seam returning an error, (c) the Windows stub.
- **When** the sweep runs.
- **Then** all three preserve; (a) names the probe in the reason; (b)/(c) record `undetermined`-style unanswerable, never a negative.

### AC-WS-008 — Dry-run default, --yes, --json parity
- **Given** a fixture with one disposable and one preserved tree.
- **When** the sweep runs (a) bare, (b) with `--yes`, (c) with `--json`.
- **Then** (a) removes nothing and prints the Would-remove preview; (b) removes exactly the disposable tree; (c) removes nothing and the JSON records carry the same predicate values the text report showed (one evaluation, two renderings).

### AC-WS-009 — L1 hoist-then-remove
- **Given** an L1 fixture tree (under the fixture main root's `.claude/worktrees/`) holding `.moai/reports/` evidence, fully landed and otherwise clean.
- **When** the sweep runs with `--yes`.
- **Then** the hoist seam executed before the removal seam (order asserted), evidence exists under the project root after, the tree is gone; and in a second cell where the hoist fails, the tree survives with `cause=hoist-failed`.

### AC-WS-010 — L2 done-path routing
- **Given** a disposable tree outside both L1 roots of the fixture main root.
- **When** the sweep runs with `--yes`.
- **Then** disposal went through the done core (hoist performed, no `--force`, no branch deletion), `tier=L2` in the record; and an L1-tree cell confirms routing never sends L1 trees to the done core.

### AC-WS-011 — Never-dispose list
- **Given** the fixture main checkout, the process's own tree, a base-branch checkout, a locked tree.
- **When** the sweep runs with `--yes`.
- **Then** none appears as DISPOSE; protected trees are absent from the records entirely (outside the universe, per the `clean --stale` convention).

### AC-WS-012 — Base default and override
- **Given** a fixture whose landing truth differs between `origin/develop` and `origin/main`.
- **When** the sweep runs bare and then with `--base origin/main`.
- **Then** bare uses `origin/develop` (help text states the divergence from `clean --stale`), the override reclassifies accordingly.

### AC-WS-013 — Verdict record completeness
- **Given** the AC-WS-001 fixture.
- **When** `--json` runs.
- **Then** every record carries `path`, `branch`, `tier`, `verdict`, `reason`, and every predicate field with a value from {yes, no, undetermined, not-checked}; an unobserved predicate is never rendered as a negative.

### AC-WS-014 — Removal-failure resilience + doctrine step
- **Given** (a) a disposable tree whose removal fails (seam), (b) the two doctrine files.
- **When** the sweep runs with `--yes`; then the doctrine checks execute.
- **Then** (a) the failure is a non-blocking notice, the remaining trees still process, exit reflects the degraded-run contract; (b) both doctrine files contain a `sweep`-naming mandatory-lifecycle wording (`grep -n 'worktree sweep' <file>` ≥ 1 hit in each), `worktree-integration.md` added NO new headings (`grep -c '^#'` unchanged) and its byte delta — the `wc -c` value at judgment time minus the run-start `wc -c` baseline recorded in `progress.md` §E.1 (measured 41194 on 2026-09-30 in this tree; the §E.1 record, not this citation, is the instrument) — is ≤ +1,200 chars, and `kanban-dispatch.md` contains a companion-pointer reference (`grep -n 'kanban-dispatch-detail\|worktree-integration' ` at the addition site).

## §D.2 — Edge cases

- Detached-HEAD tree → landing predicate unanswerable (no branch tip) → PRESERVE.
- Tree vanished between enumeration and removal → non-fatal notice, count reported honestly.
- Empty population → "nothing to sweep" exit 0 with the swept count stated (an empty sweep asserts nothing).

## §D.3 — Quality gates

- Coverage ≥85% on new sweep files (E3). Lint: no NEW findings vs baseline (E5). GOOS=windows build green (E2). Subagent-boundary grep 0 (E4).

## §D.4 — Definition of Done

- All 14 ACs PASS with verbatim evidence in run-phase §E; M5 doctrine deltas recorded (byte counts) in progress.md; no PRESERVE-target file modified beyond the one `root.go` registration line.
