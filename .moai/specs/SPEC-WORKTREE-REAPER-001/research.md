# research.md — SPEC-WORKTREE-REAPER-001

Prior-art and existing-code survey. Everything here was read or measured in
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t209` (branch
`WT-worktree-reaper`, base `cd0cee1b8`) on 2026-08-24. The narrative survey that
established the problem lives in `.moai/reports/t209/investigation.md`; this file
carries the code-level findings the design decisions rest on.

## §A — There are TWO worktree sweeps, and they share one blind dependency

The single most consequential finding, and the one v0.1.0 of this SPEC missed.

| | `prMergeCleanup` | `cleanStaleWorktrees` |
|---|---|---|
| File | `internal/cli/session_worktree_prmerge.go` | `internal/cli/worktree/clean.go` |
| Entry | automatic, on `moai session register` / `moai session list` | explicit, `moai worktree clean --stale` |
| Gate | `workflow.worktree.auto_cleanup` (ships `false`; `true` here) | `--yes` (preview otherwise) |
| Population | branches prefixed `WT-` only | **every registered worktree** |
| Merge source | `gh pr view` primary, `git branch --merged origin/main` fallback | `IsBranchMerged(branch, base)`, base defaults to local `main` |
| Dirty guard | `worktreeIsDirty` → `git status --porcelain` | `worktreeHasLocalChanges` → `git status --porcelain` |
| Anchor guard | `session.LiveAnchoredSessions` | `session.LiveAnchoredSessions` |
| Branch deletion | never | never |

The two rows that matter: they sweep **different populations** but consume the
**same anchor source**, and the one with the wider population is the one gated
behind an explicit flag. Repairing only the automatic sweep leaves the blind
guard on the surface that can remove more.

This is why REQ-WR-019 places the repaired decision in `internal/session` rather
than in a new `internal/cli` file.

## §B — Existing seam and test conventions to extend

The package already has the injection mechanism this SPEC needs; nothing new is
warranted.

- **Seam style**: package-level function variables with a `Real` counterpart
  (`sessionWorktreeGhPRViewState = ghPRViewStateReal`), swapped by a
  `swap*Seams(t, …)` helper that registers restoration via `t.Cleanup`.
  Two helpers exist: `swapPRMergeSeams` (M8 seams) and `swapSessionWorktreeSeams`
  (M2/M4/M7 seams).
- **Existing `prMergeCleanup` coverage**: 24 test functions across
  `session_worktree_prmerge_test.go` and `session_worktree_prmerge_anchor_test.go`.
  The ones this SPEC leans on as regression guards, with their **exact** names:
  `TestPRMergeCleanup_GhPresentMergedRemoves`,
  `TestPRMergeCleanup_GhPresentOpenDoesNotRemove`,
  `TestPRMergeCleanup_GhPresentSeesSquashMerge`,
  `TestPRMergeCleanup_GhAbsentBranchMergedRemoves`,
  `TestPRMergeCleanup_GhAbsentEmitsBlindnessNoticeOnce`,
  `TestPRMergeCleanup_ToggleOffNoOp`, `TestPRMergeCleanup_NilCfgNoOp`,
  `TestPRMergeCleanup_WorktreeListErrorFailOpen`,
  `TestPRMergeCleanup_AnchoredSessionSkipsRemoval`.
  (v0.1.0 of this SPEC cited two names — `…_GhMergedRemoves`,
  `…_GhAbsentFallback` — that do not exist. Both are corrected in
  `acceptance.md`.)
- **Existing `clean --stale` coverage**: `TestCleanStale_KeepsAnchoredWorktree`,
  `…_KeepsDirtyWorktree`, `…_KeepsUnmergedWorktree`, `…_PreviewsByDefault`,
  `…_RemovesWithYes`, `…_SkipsProtectedAndDetached`,
  `…_RejectsMergedOnlyCombination`.
- **No-prompt guards**: 31 `Test*NoAskUserQuestion` functions exist across
  `internal/cli`. An unanchored `-run 'NoAskUserQuestion'` matches all of them
  and binds nothing about a new change — hence the exact-name requirement in
  `acceptance.md` AC-WR-020.

## §C — Merge-detection prior art

### C.1 — `branchMergedForCleanup` (the defect site)

```go
if ghAvailable {
    return sessionWorktreeGhPRViewState(branch) == "MERGED"
}
```

The git fallback below this line is reachable only when `gh` is **absent**.
`ghPRViewStateReal` returns `""` on any `exec` error and on unparseable JSON, so
"gh could not tell me" and "gh told me it is not merged" are the same value.

### C.2 — `IsBranchMerged` is a richer predicate than `git branch --merged`

`internal/core/git/worktree.go` implements a staged check: **S1** is
`git branch --merged <base>` (reachability — the only signal that recognises a
true merge commit and a strictly-behind branch), followed by patch-id stages
that recognise squash and rebase merges. It has a dedicated 10-case test suite
(`worktree_squash_merge_test.go`) covering squash, rebase, true-merge,
strictly-behind, empty-diff, partially-applied, revert, superset, and
rename-re-add cases.

Two consequences for this SPEC:

1. `clean --stale` is **not** squash-blind, unlike `prMergeCleanup`'s fallback.
   The two sweeps have different merge-detection strength, which is a latent
   inconsistency this SPEC does not attempt to unify (out of scope) but does
   record.
2. `IsBranchMerged`'s S1 stage *is* `git branch --merged`, so the
   "unique-commit check" that `clean --stale`'s doc comment advertises is not a
   separate predicate — it is what reachability already means. This is the
   measurement behind `design.md` §A.4's decision to accept the
   zero-unique-commit class rather than add a redundant guard.

### C.3 — Neither sweep fetches

`grep` over both files finds no `git fetch`. `prMergeCleanup` compares against
whatever `origin/main` the local ref currently points at. A stale ref yields
fewer ancestors and therefore fewer removals — the safe direction — with the
single exception of a force-pushed / rewritten `origin/main`. Recorded as EC-10;
no requirement asserts safety here.

## §D — Anchor-detection prior art

### D.1 — `LiveAnchoredSessions` consults two registries, and still misses 4 of 5

Its doc comment already documents the failure mode this SPEC routes around:
sessions that enter a tree via `EnterWorktree` "keep their launch-time CWD and
stay invisible here" until the `CwdChanged` hook calls `RelocateSession`.
Measured, that relocation is not running for 4 of the 5 live lanes.

A second measured property compounds it: run from *inside* a worktree,
`moai session list --json` returns `[]`. Any anchor test built on that query is
a false-negative machine.

### D.2 — The platform probes cannot express uncertainty

- `anchor_pid_unix.go` — `isProcessAlive(pid) bool`: `nil` and `EPERM` → true;
  **every other errno → false**, i.e. reported as dead.
- `anchor_pid_windows.go` — returns `true` unconditionally, by design, with an
  in-file note deferring native `OpenProcess` validation.

So REQ-WR-008's "probe undetermined" case cannot be expressed by the existing
signature. `design.md` §B.5 specifies the two-valued replacement and its
per-platform mapping. The Windows behaviour is safe for this SPEC's purposes —
a platform that can never assert "dead" can never widen removal — but it does
mean `GOOS=windows go vet` is compilation evidence only.

### D.3 — `moai` never locks, and never unlocks

`grep` over the repository finds no `git worktree lock` invocation and exactly
one `unlock` reference — a hint string in `internal/cli/worktree/done.go`. The
locks on the five live trees are written by Claude Code at `EnterWorktree`.

`materializeSessionWorktree` (`internal/cli/session_worktree.go`) creates
`WT-<session>-<subcommand>` trees with a plain `git worktree add` — inside the
swept prefix set, and unlocked. That is the residual REQ-WR-020 records.

### D.4a — git's removal check is stricter than `git status --porcelain`

Measured in a scratch repository (`.moai/reports/t209/ec9-measurement.md` §Q1):
with a committed `.gitignore` and one ignored file present in a linked worktree,
`git status --porcelain` returns **0** lines while `git worktree remove` exits
**128** with `contains modified or untracked files, use --force to delete it`,
and the ignored file survives.

So the two checks disagree about what counts as content: `git status
--porcelain` omits gitignored entries, `git worktree remove` counts them. Every
guard in this SPEC built on the former is therefore backstopped by the latter —
which is why EC-9 closes in the safe direction and why `design.md` §A.4's
accept-the-class decision holds at all three layers (committed work, tracked or
untracked work, ignored work).

The second-order finding is what drove REQ-WR-021's generalisation: the
ignored-only tree is selected on every sweep and refused every time, with
nothing clearing the condition — the identical permanent-recurring shape as the
locked tree in §D.4. Two causes, one symptom, one of them not detectable by the
guard that would have to catch it.

### D.4 — git refuses to remove a locked tree

`gitWorktreeRemoveReal` runs `git worktree remove <path>` with no `--force`. git
exits 128 on a locked tree regardless of the locking process's liveness. The
resulting notice is already visible in the investigation transcript for t212.
This is the measurement behind the dead-lock-inert decision (`design.md` §B.6,
REQ-WR-021).

## §E — Verification-tooling finding

The falsifiability property of the criterion set is itself a research finding,
because it is a property of the Go toolchain rather than of this codebase:

```
go test <pkg> -run <PatternMatchingNothing>   → exit 0, "ok … [no tests to run]"
```

Any acceptance criterion of the form "run this test, expect exit 0" is therefore
satisfied by a tree in which the test does not exist. The falsifiable form, and
its verification in both directions, is recorded in `acceptance.md` §0. This is
a general hazard for SPEC authoring in this repository, not specific to t209.

## §F — What was surveyed and found NOT to apply

- **L2 worktree tooling** (`moai worktree done`, `~/.moai/worktrees/`, the L2
  registry) — the registry file is `{}` (2 bytes) and no worktree lives under
  the L2 root. Every disposal path in scope is L1.
- **`git worktree prune`** — `--dry-run` reports 0 prunable entries; there are no
  stale administrative directories. The 155 entries are all real trees on disk.
- **`--merged-only`** (the other `clean` mode) — removes merged worktrees without
  the dirty/anchor pairing that `--stale` applies. Not extended by this SPEC and
  not used as prior art; noted so a future reader does not confuse the two modes.
