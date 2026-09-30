---
id: SPEC-WORKTREE-SWEEP-001
title: "Implementation plan — worktree post-landing auto-sweep"
version: "0.1.0"
created: 2026-09-30
updated: 2026-09-30
author: manager-spec
---

# plan.md — SPEC-WORKTREE-SWEEP-001

## §A — Context

- **Worktree**: `.claude/worktrees/t1369`, branch `WT-worktree-sweep`, based on `origin/develop@3dd5adf2f`. All work happens here; lane does not push (gitflow-lane-protocol §4).
- **Cycle**: tdd (RED-GREEN-REFACTOR per milestone). Tier M.
- **Artifacts**: `.moai/specs/SPEC-WORKTREE-SWEEP-001/{spec,plan,acceptance,design,research,progress}.md`.
- **PRESERVE targets** (never modify):
  - `internal/cli/worktree/clean.go`, `done.go`, `remove.go`, `guard.go`, `root.go` bodies except the single registration line in `root.go` `init()` (`newSweepCmd()`).
  - `internal/session/` (anchor decision — consumed, not changed).
  - `internal/cli/update_worktree_processes.go` (lsof pattern source — copied, not edited).
  - Other SPEC directories, `.moai/state/`, `.moai/harness/`, other cards' trees.
- **EXTEND targets**:
  - NEW `internal/cli/worktree/sweep.go` + `sweep_cwd_posix.go` + `sweep_cwd_windows.go` + `sweep_test.go` (+ fixture helpers as needed).
  - `internal/cli/worktree/root.go` `init()` — one registration line.
  - `.claude/rules/moai/workflow/kanban-dispatch.md` — minimal pointer addition (§ Integration into the release branch is self-served or Isolation table vicinity).
  - `.claude/rules/moai/workflow/worktree-integration.md` — minimal extension of an existing [HARD] sentence (hoist clause or L1-disposal clause); NO new section.

## §B — Known Issues (B1–B12, filtered to relevant)

- **B1 Cross-platform build tags**: the cwd probe needs `//go:build !windows` / `//go:build windows` split. Verify `GOOS=windows GOARCH=amd64 go build ./...` (E2). Windows stub returns an unanswerable error (preserve-by-design, REQ-WS-006).
- **B3/B11 Subagent boundary**: CLI code must not call AskUserQuestion. Add the package's canonical static-guard test shape (`TestNew_NoAskUserQuestion` family) for the sweep command.
- **B6 spec-lint headings**: spec.md §F carries `### Out of Scope — <topic>` H3 sub-headings with `-` bullets (done).
- **B8 Working-tree hygiene**: commits stage explicit pathspecs only; no runtime-managed files.
- **Import-cycle hazard (plan-specific)**: the cwd probe CANNOT import `internal/cli`'s `activeProcessCWDs` (parent package imports `internal/cli/worktree` for DI wiring — a child→parent import cycles). Duplicate the small lsof pattern inside the `worktree` package behind a seam var, citing the original in a comment.
- **t1360 lesson**: do not truncate test output through pipes; let failure names and exit codes through.
- **t528 lesson**: worktrees can vanish under you — re-enumerate between batches is not needed here (single pass), but the removal loop must tolerate already-gone trees as a non-fatal notice.
- **t1350 lesson**: lane env falsifies env-reading guard tests locally — the cwd-probe seam keeps tests hermetic; never assert on the developer machine's real lsof output.
- **t1330/t1358 vocabulary guard**: new user-facing strings must pass the vocabulary checks; keep flags/labels consistent with existing `worktree` verbs.

## §C — Pre-flight

```bash
git branch --show-current && git rev-parse --short HEAD   # WT-worktree-sweep
go build ./...                                            # baseline green
GOOS=windows GOARCH=amd64 go build ./...                  # baseline green
golangci-lint run --timeout=2m ./internal/cli/worktree/... | tail -5   # baseline lint
grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/worktree/ | grep -v _test.go | grep -v '// '   # expect 0
ls internal/cli/worktree/                                 # confirm PRESERVE/EXTEND file set
```

## §D — Constraints (DO NOT VIOLATE)

- Every predicate unanswerable ⇒ PRESERVE (REQ-WS-005). No force, no unlock (REQ-WS-007).
- Branch deletion prohibited (REQ-WS-010).
- Dry-run default; nothing removed without `--yes` (REQ-WS-012).
- `--base` default `origin/develop` (REQ-WS-004); divergence from `clean --stale` documented in help text.
- Doctrine edits: `worktree-integration.md` delta must stay within ~1,200 characters added and add no headings (file already over the 40,000-char budget; the judging baseline is the run-start `wc -c` value recorded in progress.md §E.1 — re-measure at M5, never a pinned constant); `kanban-dispatch.md` addition should be pointer-style (stub+companion split). No restructuring of either file.
- Conventional commits, `🗿 MoAI` trailer, card id t1369 in every commit message; `--no-verify` forbidden.
- No AskUserQuestion in CLI code; no free-form user questions from this lane.

## §E — Self-Verification (deliverables at run close)

- **E1** AC PASS/FAIL matrix per `acceptance.md` §D (commands verbatim, outputs verbatim).
- **E2** `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` → both exit 0.
- **E3** `go test -cover ./internal/cli/worktree/...` → ≥85% on the new sweep files.
- **E4** subagent-boundary grep → 0 matches.
- **E5** `golangci-lint run --timeout=2m ./internal/cli/worktree/...` → no NEW findings vs the §C baseline.
- **E6** commit SHAs + push state (push is lead-batched; lane reports local SHAs only).
- **E7** blocker report if any user decision is needed.
- **E8** RED failure outputs captured before each GREEN (tdd cycle), verbatim.

## §F — Milestones (ordered by decision-reversibility: data-model and surface first, mechanics last)

### M1 — Verdict record + command skeleton (HIGH change-likelihood: schema + flags)

- RED: tests for the `sweepVerdict` record shape (JSON tags per spec §D.4) and flag surface: `--base` default `origin/develop`, `--yes`, `--json`; dry-run default emits records and removes nothing.
- GREEN: `newSweepCmd()` registered in `root.go`; enumeration + protected-tree exclusion; verdict records emitted for every non-protected tree with all predicates `not-checked` except tier.
- Decision pinned here: record field set, flag names (`--yes` aligns with `clean --stale`), default base.

### M2 — Remote-landing predicate (three-way contract)

- RED: fixture repos proving exit 0 ⇒ landed, exit 1 ⇒ PRESERVE (`cause=not-landed`), fetch failure ⇒ PRESERVE (`cause=fetch-failed`), ancestry exit 2/other ⇒ PRESERVE (`cause=landed-check-failed`), missing base ref ⇒ PRESERVE; `--base` override honored.
- GREEN: `git fetch` + `git merge-base --is-ancestor` behind a seam; `cause=` tokens matching the `clean` sweeps' vocabulary.

### M3 — Safety predicate composition

- RED: dirty ⇒ PRESERVE; irreplaceable ignored content ⇒ PRESERVE; locked ⇒ PRESERVE; anchor decision (lock ∪ registry) ⇒ PRESERVE with source named; cwd-probe hit ⇒ PRESERVE; probe unanswerable (seam returns error / windows) ⇒ PRESERVE; base-branch checkout ⇒ PRESERVE.
- GREEN: compose over `worktreeLockStates`, `ignoredContentVerdict`, `session.AnchorDecision`, `worktreeHasLocalChanges`, new `processCWDInside(tree)` seam (`lsof -a -d cwd -F0n`, `_posix`/`_windows` split).

### M4 — Tier routing + apply path

- RED: L1 tree (fixture under `.claude/worktrees/` of the fixture main root) ⇒ hoist-then-remove order asserted (hoist seam records before remove seam); hoist failure ⇒ PRESERVE; L2 tree ⇒ routed through the done core (hoist inside, no branch deletion); `--yes` removal failure ⇒ non-blocking notice, loop continues; degraded anchor source ⇒ exit 2 after full report.
- GREEN: wire routing via `isL1WorktreePath` + `hoistBeforeDisposal` + `runDoneWorktreeCleanupWithOptions` (or the remove core it wraps).

### M5 — Doctrine additions (both files, minimal delta)

- `worktree-integration.md`: extend the existing hoist [HARD] clause (and/or the L1-disposal clause) with one or two sentences naming `moai worktree sweep` as the recurring execution of the remote-landing-confirmed disposal step. No new headings.
- `kanban-dispatch.md`: one pointer sentence in the integration section citing the sweep and the companion detail. Stub stays small.
- Verify: `wc -c` deltas recorded in progress.md; grep finds `sweep` in both files.

### M6 — Scope verification

- `go test ./internal/cli/worktree/...` (scoped per lane-local verification discipline), E1–E8 batch, `make build` only if templates were touched (not expected).

## §G — Anti-Patterns

- Do NOT re-derive predicates already shipped (`ignoredContentVerdict`, `worktreeLockStates`, `AnchorDecision`) — compose them.
- Do NOT classify the whole population then act tree-by-tree without re-reading the ignored-content predicate immediately before removal (the `clean --stale` removal-time re-read is load-bearing; replicate it).
- Do NOT treat an empty sweep as a pass — assert the swept count in tests (verification-completeness §1.1).
- Do NOT grow `worktree-integration.md` with a new section; the budget hook is already red on that file.
- Do NOT delete branches "while we're in there".

## §H — Cross-References

- spec.md §D (REQ-WS-001..014) · acceptance.md §D (AC matrix) · design.md §A (new-verb decision) · research.md (external digest)
- Prior SPECs: SPEC-WORKTREE-REAPER-001, SPEC-WORKTREE-GC-001, SPEC-WORKTREE-DONE-TIER-001, SPEC-REPORTS-LIFECYCLE-001
