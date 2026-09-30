---
id: SPEC-WORKTREE-SWEEP-001
title: "Design decisions — worktree post-landing auto-sweep"
version: "0.1.0"
created: 2026-09-30
updated: 2026-09-30
author: manager-spec
---

# design.md — SPEC-WORKTREE-SWEEP-001

## §A — Decision: new `sweep` verb vs strengthening `clean --stale` in place

**Decision: a new `moai worktree sweep` verb that composes the hardened predicates and both tier disposal paths, reusing `clean --stale` internals.** The card asked for "신설" (new command); the simplicity ladder demanded the reuse-first alternative be evaluated honestly. It was, and it loses — but narrowly, and the losses are stated below.

### A.1 Why extend-in-place was considered

`clean --stale` already enumerates every registered tree, previews by default, gates removal behind `--yes`, carries the ignored-content guard, the shared anchor decision, the exit-2 degraded contract, and a `--json` inventory (SPEC-WORKTREE-REAPER-001 M3). On paper, "sweep" is `clean --stale` plus a fetch-hardened landing predicate and L2 routing. Reusing that shell is exactly what the ladder's rung 2 asks for.

### A.2 Why it loses

1. **Predicate divergence is the point, not an accident.** `clean --stale`'s merge check is local `IsBranchMerged` against `origin/main` — deliberately no fetch (the comment pins that choice: the default is a remote ref read without mutation). The sweep's core repair is that this is not a landing observation: it needs a fresh `git fetch` + a three-way `--is-ancestor` contract with a different default base (`origin/develop`). Folding that into `clean --stale` means either changing the existing command's behavior (its default base, its no-fetch property — both are pinned by REQ-WR-022 and its own comments, i.e. a breaking change to a documented contract) or carrying two divergent predicate sets under one verb — the worst of both.
2. **Flag-matrix saturation.** `clean` already multiplexes `--merged-only` × `--stale` × `--json` × `--base` × `--yes` with explicit mutual-exclusion checks. Adding tier routing (L1 hoist+remove vs L2 done-core) and the fetch semantics grows that matrix past the point where its help text or its tests can hold it.
3. **Tier routing is a different disposal graph.** `clean --stale` only knows `WorktreeProvider.Remove`. The sweep must route L2 trees through the done core (built-in hoist, anchor refusal, L1 refusal) and L1 trees through hoist+remove. That is a second disposal path `clean` structurally lacks.
4. **Doctrine names the verb.** Deliverable 3 mandates doctrine say `moai worktree sweep`. A doctrine that names a verb no binary ships is worse than either alternative.

### A.3 What the new verb costs (stated honestly)

- **Discoverability**: two adjacent verbs (`clean`, `sweep`) invite "which one?" confusion. Mitigation: `sweep`'s help cross-references `clean --stale` ("stale-references and no-unique-commits sweeping lives in clean --stale; sweep disposes only remote-landing-confirmed trees"), and both share the `cause=` output vocabulary so one grep finds both.
- **Duplicated shell**: enumeration, protected-tree exclusion, and preview/apply scaffolding are re-implemented (~150 lines) rather than shared. Mitigation: the *predicates* are shared functions, not copies; only the orchestration shell is new.
- **Naming drift risk**: `clean --stale`'s `--yes` vs a hypothetical `--apply`. Resolved in favor of `--yes` for consistency with the existing verb.

### A.4 Reuse ledger (what is NOT re-implemented)

`worktreeLockStates`, `ignoredContentVerdict`, `protectedWorktreePaths`, `isBaseBranch`, `worktreeHasLocalChanges`, `staleState*` vocabulary (`clean.go`); `isL1WorktreePath`, `hoistBeforeDisposal`, `runDoneWorktreeCleanupWithOptions`, `formatAnchored` (`done.go`); `session.AnchorDecision` (`internal/session`); and the enumeration itself — `WorktreeProvider.List()` (the GitWorktree-backed manager wired in `internal/cli/root.go:149`), the L1/L2 tier classification's input. New code: the fetch+ancestry seam, the cwd-probe seam (+ platform split), the verdict record, the orchestration loop.

## §B — Positioning against the existing sweeps

| Surface | Signal | Branch scope | Tier | Default |
|---|---|---|---|---|
| `prMergeCleanup` (AutoCleanup) | gh MERGED or local `--merged` (squash-blind), no fetch | `WT-` only | L1 | off (distributed) |
| `clean --stale` | local `IsBranchMerged` vs `origin/main`, no fetch | all registered | L1 (remove only) | preview |
| `worktree done` | manual, per-tree | one | L2 only (refuses L1) | apply |
| **`sweep` (this SPEC)** | fresh fetch + three-way `--is-ancestor`, base-configurable | all registered | L1 + L2 | preview |

The sweep is stricter than all three where it disposes: no disposition without affirmative remote-landing evidence, and every unanswerable predicate preserves.

## §C — The cwd probe

Follows `internal/cli/update_worktree_processes.go` (`lsof -a -d cwd -F0n`, NUL-field parse) and the `leader_readers_darwin.go` precedent. It cannot import the existing helper: `internal/cli` is the parent package (imports `worktree` for DI wiring), so a child→parent import cycles. The probe is duplicated small behind a seam var (`processCWDInside func(tree string) (bool, error)`) in `sweep_cwd_posix.go` / `sweep_cwd_windows.go`, citing the original. Windows: return an unanswerable error (no supported cwd API — research.md §4). Scale note: one whole-system lsof invocation cached per sweep run (not per tree) keeps 992-tree runs off the `+D` slow path (research.md §4).

## §D — Fetch semantics

The fetch runs once per sweep invocation (`git fetch origin <base-branch>` derived from `--base`; on fetch failure every tree's landing predicate is unanswerable → all preserve with `cause=fetch-failed`). Per REQ-WS-001 a stale remote-tracking ref never satisfies the predicate alone; this follows SPEC-WORKTREE-GC-001 REQ-WGC-004 (t1059 lesson). Dry-run and `--json` fetch too — an inventory that lies about landing is worse than a mutated remote-tracking ref.

## §E — Removal-time re-read

The `clean --stale` removal loop re-reads the ignored-content predicate immediately before each removal because classification of the whole population precedes per-tree action. The sweep replicates that pattern (check→act window narrowed, not closed) and additionally re-checks nothing else: the landing/anchor predicates are read-only, and git independently refuses dirty tracked content at removal time.
