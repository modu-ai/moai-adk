---
id: SPEC-WORKTREE-SWEEP-001
title: "Worktree post-landing auto-sweep: remote-landing-confirmed disposal of merged worktrees across L1 and L2 tiers"
version: "0.1.0"
status: draft
created: 2026-09-30
updated: 2026-09-30
author: manager-spec
priority: P1
phase: "v3.3.0 target"
module: "internal/cli/worktree"
lifecycle: spec-anchored
tags: "worktree,sweep,disposal,remote-landing,lifecycle,doctrine"
tier: M
related_specs: [SPEC-WORKTREE-REAPER-001, SPEC-WORKTREE-GC-001]
---

# SPEC-WORKTREE-SWEEP-001

## §A — History

- **2026-09-30** — plan-phase v0.1.0 authored from card t1369 (operator directive 2026-09-30). Operator disk scan measured 2.83 TB scanned, 1.46 TB recoverable, 992 workspaces; the mo.ai.kr workspace's `.moai` directories alone hold 1.21 TB (2–4 GB per tree). Trees whose work is merged and remotely landed keep accumulating without disposal. SPEC-WORKTREE-GC-001 (completed 09-23) proved a one-time disposal without a recurring mechanism re-accumulates. Prior art SPEC-WORKTREE-REAPER-001 (v0.4.1, completed) supplies the shared anchor decision and `clean --stale` sweep internals; both are reused, not re-implemented. Tier M, cycle tdd. External best-practices research (delivered with the card) persisted as `research.md`.

## §B — Problem

Card worktrees whose branch is fully merged into the integration branch **and** confirmed landed on the remote keep accumulating on disk. The operator's 2026-09-30 scan measured 2.83 TB across 992 workspaces with 1.46 TB recoverable; per-tree `.moai/` state alone weighs 2–4 GB. Three mechanisms already exist and none closes the gap:

1. `moai worktree clean --stale` (SPEC-WORKTREE-REAPER-001 M3) sweeps every registered tree, but its merge predicate is a **local** `IsBranchMerged` against `--base origin/main` (the wrong integration branch for this git-flow repo) with **no fetch** — a stale remote-tracking ref is not a landing observation, and the command cannot distinguish "not landed" from "no answer".
2. The `Workflow.Worktree.AutoCleanup` PR-merge sweep covers only `WT-` branches on a gh-`MERGED` signal (squash-merge blind) and is off by default.
3. `moai worktree done` disposes L2 trees correctly but is per-tree, manual, and refuses L1 trees by code.

SPEC-WORKTREE-GC-001 disposed ~530 trees in a one-time operation; accumulation resumed immediately because no **mandatory lifecycle step** re-runs the disposal. The gap is therefore threefold: no remote-landing observation (fresh fetch + three-way ancestry), no tier-agnostic single-pass sweep covering both L1 and L2 trees with evidence hoisting, and no doctrine step that makes the sweep actually run.

## §C — Goal

A `moai worktree sweep` command that disposes — after per-tree, auditable predicate evaluation — only trees whose branch is confirmed landed on the remote integration base, across both L1 (`.claude/worktrees/`) and L2 (registry) tiers, hoisting evidence before every disposal, preserving everything it cannot affirmatively judge, and a doctrine amendment that makes "remote-landing confirmed → tree disposal" a mandatory step of the card lifecycle.

## §D — Requirements (GEARS)

14 requirements. `<subject>` is "the sweep" (`moai worktree sweep`) unless named otherwise. All safety predicates are AND-ed; every predicate that cannot be answered means PRESERVE (REQ-WS-005).

### D.1 — M1: Remote-landing observation

- **REQ-WS-001** (Ubiquitous) — The remote-landing predicate shall observe ancestry with a fresh `git fetch` of the integration base's remote, followed by `git merge-base --is-ancestor <branch-tip> <base>`, and shall represent three distinct outcomes: *landed* (ancestry exit 0), *not landed* (ancestry exit 1), and *unanswerable* (fetch failure, unresolvable base ref, or any ancestry exit other than 0 or 1). A stale remote-tracking ref shall never satisfy the predicate on its own.
- **REQ-WS-002** (State-driven) — While the ancestry check exits 0 for a tree's branch tip, the sweep shall classify that branch as remotely landed.
- **REQ-WS-003** (Event-driven) — When the fetch fails, the base ref cannot be resolved, or the ancestry check exits with any code other than 0 or 1, the sweep shall classify the tree as PRESERVED with a `cause=`-prefixed reason naming the failure — never as disposable.
- **REQ-WS-004** (Ubiquitous) — The integration base shall default to `origin/develop` and shall be overridable per invocation via `--base <ref>`; the default's divergence from `clean --stale`'s `origin/main` default shall be stated in the command help.

### D.2 — M2: Safety predicate set (AND-ed, fail-safe)

- **REQ-WS-005** (Unwanted) — The sweep shall not dispose a tree unless every predicate is affirmatively observed: not a protected tree, not locked, not anchored by a live session, no live process cwd inside the tree, clean working tree, no irreplaceable ignored content, and remotely landed per REQ-WS-001. Any predicate that cannot be answered — error, unsupported platform, unreadable source — shall preserve the tree. Absence of a failure signal is never a dispose signal.
- **REQ-WS-006** (Ubiquitous) — The live-session predicate shall combine the shared lock-∪-registry anchor decision (`session.AnchorDecision`, the SPEC-WORKTREE-REAPER-001 repair) with a process-cwd probe (lsof on darwin/Linux, following the repository's existing platform-split pattern); the probe shall run with a test-injectable seam, and an unanswerable probe (lsof absent or failing, or any Windows host) shall report the tree as anchored.
- **REQ-WS-007** (Unwanted) — The sweep shall not dispose the main/integration checkout, the tree the process runs in, a locked tree, a tree checked out on the base branch, or a tree with unique unpushed commits; and the sweep shall never call `git worktree unlock` nor pass `--force` to `git worktree remove`.

### D.3 — M3: Tier routing and evidence hoisting

- **REQ-WS-008** (State-driven) — While a fully-predicated tree is an L2 registry tree (outside both L1 roots), the sweep shall dispose it through the `done` removal core (`runDoneWorktreeCleanupWithOptions` semantics: anchor refusal, built-in evidence hoist, non-forced removal, no branch deletion).
- **REQ-WS-009** (State-driven) — While a fully-predicated tree is an L1 tree (under `<mainRoot>/.claude/worktrees/` or `<mainRoot>/.moai/worktrees/`, resolved from the target path per the SPEC-WORKTREE-DONE-TIER-001 predicate), the sweep shall hoist the tree's `.moai/reports/` evidence into the project root before removal, and a hoist failure shall preserve the tree.
- **REQ-WS-010** (Unwanted) — The sweep shall never delete a branch; branch deletion remains out of scope.

### D.4 — M4: Command surface and output

- **REQ-WS-011** (Ubiquitous) — The sweep shall emit one per-tree verdict record carrying the tree path, branch, tier (L1/L2), each predicate's observed value (using `clean --stale`'s four-valued vocabulary — `yes`/`no`/`undetermined`/`not-checked` — with unobserved predicates never reported as negatives), and a DISPOSE/PRESERVE verdict with a cause-naming reason — so an operator can audit every decision without inspecting trees.
- **REQ-WS-012** (Ubiquitous) — The sweep shall preview by default and remove nothing without `--yes`; `--json` shall emit the same evaluation as machine-readable records, so the inventory and the sweep can never disagree (one evaluation, rendered twice).
- **REQ-WS-013** (Event-driven) — When a `--yes` removal fails for a tree, the sweep shall report the failure as a non-blocking notice and continue with the remaining trees; a degraded run (unreadable anchor source) shall end with the distinguished exit-2 preservation signal, matching the `clean --stale` contract.

### D.5 — M5: Doctrine — a mandatory lifecycle step

- **REQ-WS-014** (Ubiquitous) — The worktree lifecycle doctrine shall name "remote-landing confirmed → tree disposal" as a mandatory step of the card lifecycle, via minimal precise additions to `.claude/rules/moai/workflow/kanban-dispatch.md` and `.claude/rules/moai/workflow/worktree-integration.md` that name `moai worktree sweep`. The `worktree-integration.md` addition shall extend existing [HARD] clauses or cross-reference sentences rather than adding new sections, and shall not grow the file materially past its over-budget state — the file already exceeds the 40,000-char instruction-file budget, and the concrete bound is a byte delta of at most +1,200 from the run-start `wc -c` measurement recorded in `progress.md` §E.1 (a point-in-time figure; each run re-measures before judging); the `kanban-dispatch.md` addition may use companion-pointer wording to keep the stub small.

## §E — Constraints

- **All predicates AND-ed; fail-safe everywhere.** The dispose decision requires affirmative evidence on every predicate. The three-way ancestry contract (0/1/other) follows `git merge-base --is-ancestor`'s documented exit semantics.
- **Fetch is an accepted mutation.** `git fetch` updates remote-tracking refs; this is the price of a genuine landing observation (SPEC-WORKTREE-GC-001 REQ-WGC-004 established the fetch-first rule: a remote ref read without fetch measures its own staleness). Dry-run still fetches; `--json` fetches; only the pure predicate-unit tests run without git.
- **Reuse over reimplementation.** The sweep composes existing internals: `worktreeLockStates`, `ignoredContentVerdict`, `protectedWorktreePaths`, `isBaseBranch`, `worktreeHasLocalChanges` (`clean.go`), `isL1WorktreePath` / `hoistBeforeDisposal` (`done.go`), `session.AnchorDecision`, and the done removal core. The process-cwd probe is new code following the `activeProcessCWDs` lsof pattern, and must live in the `worktree` package (importing `internal/cli` would create an import cycle) behind a seam.
- **L1 refusal is not bypassable.** An L1 tree routed to the `done` core is refused by `isL1WorktreePath` with or without force; routing is therefore determined before disposal, not discovered by the refusal.
- **Windows behavior is preservation, not feature parity.** There is no supported cwd API on Windows (research.md §4); the probe reports unanswerable there, and cross-platform verification is `GOOS=windows` build evidence only.
- **Tests follow the package idioms**: table-driven, fixture repos via `gitInitFixture`-style helpers, `t.TempDir()` isolation, injection through package-level function-variable seams extended — never replaced.
- **`make build`** verifies after template-affecting changes; not expected in plan phase.

## §F — Out of Scope

### Out of Scope — branch deletion

- The sweep never deletes branches (REQ-WS-010). Disposing a tree whose branch is landed loses nothing, but ref management is a separate concern (SPEC-WORKTREE-GC-001 REQ-WGC-011 authorized ref deletion only for its one-time T1 class).

### Out of Scope — the t1338 relation (operator adjudication pending)

- Card t1338 (jev near-duplicate p=0.64) overlaps this SPEC's deliverable 3: its piece 4 proposes "auto-dispose card WT on origin push confirmation, never dispose before remote merge confirmed". This SPEC implements the command-side predicate and the doctrine step; whether t1338's remaining scope is absorbed, narrowed to a config/trigger concern, or closed as delivered-here is the operator's adjudication. Nothing in this SPEC presumes t1338's disposition.

### Out of Scope — scheduling, daemon, and automatic invocation

- No daemon, cron, hook, or session-register trigger is built. REQ-WS-014 makes the step mandatory in doctrine (human/agent-readable obligation); wiring it into an automatic trigger is follow-on work and would need its own gating SPEC (the AutoCleanup toggle precedent shows why default-off matters).

### Out of Scope — bulk disposal of the existing backlog

- Running the sweep against the current 992-tree population is an operator action, taken deliberately after the guards are in place, via the dry-run report. This SPEC ships the mechanism, not the mass removal.

### Out of Scope — Windows process-cwd detection

- The probe preserves on Windows unconditionally. Building a supported cwd API path there (NtQueryInformationProcess territory) is a separate card.

### Out of Scope — unlocking trees

- The sweep never unlocks a locked tree and never force-removes (REQ-WS-007), matching REQ-WR-021. Acquiring the authority to unlock another process's lock is a distinct escalation.

## §G — Known Relations and Non-Goals

- **t1338 (near-duplicate, jev p=0.64)** — recorded for operator adjudication (§F). This SPEC's doctrine step (REQ-WS-014) overlaps t1338 piece 4; the command (deliverable 2) is not claimed by t1338.
- **SPEC-WORKTREE-REAPER-001 (completed)** — supplies the shared anchor decision (REQ-WR-019), the exit-2 degraded-run contract (REQ-WR-016), the ignored-content guard (REQ-WR-024), and `clean --stale`'s internals. This SPEC depends on none of its defects: the remote-landing gap and the tier-agnostic single-pass are precisely what REAPER's M1/M3 did not address.
- **SPEC-WORKTREE-GC-001 (completed)** — one-time proof that disposal works and re-accumulates without a recurring step; this SPEC is the recurring step. Its T1 classification rule (fresh fetch + `--is-ancestor` against `origin/develop`) is the direct ancestor of REQ-WS-001.
- **AutoCleanup PR-merge sweep (`prMergeCleanup`)** — narrower signal (gh MERGED / local `--merged`), `WT-`-prefixed branches only, default-off. The sweep's predicate set is stricter (fresh fetch, explicit base, three-way contract) and tier-agnostic; positioning in design.md §B.

## §H — Cross-References

- Card: t1369 (operator directive 2026-09-30).
- Prior art: SPEC-WORKTREE-REAPER-001, SPEC-WORKTREE-GC-001, SPEC-WORKTREE-DONE-TIER-001 (L1 tier guard), SPEC-REPORTS-LIFECYCLE-001 (hoist obligation), SPEC-SESSION-WORKTREE-001 (unpushed-commit guard, card t673).
- Implementation anchors: `internal/cli/worktree/clean.go`, `done.go`, `remove.go`, `internal/cli/session_worktree.go` (`gitHasUnpushedReal`), `internal/cli/update_worktree_processes.go` (lsof pattern), `internal/discovery/leader_readers_darwin.go` (lsof precedent).
- Doctrine: `.claude/rules/moai/workflow/worktree-integration.md` § Hoist / § branch-naming sweep notes; `.claude/rules/moai/workflow/kanban-dispatch.md` § Integration into the release branch is self-served.
- Research: `research.md` (external best-practices digest with cited URLs); decision record: `design.md`.
