---
id: SPEC-WORKTREE-SWEEP-001
title: "Research — worktree lifecycle best practices and prior art (card t1369 digest)"
version: "0.1.0"
created: 2026-09-30
updated: 2026-09-30
author: manager-spec
---

# research.md — SPEC-WORKTREE-SWEEP-001

Persisted from the operator-commissioned best-practices research delivered with card t1369 (2026-09-30), lightly edited into SPEC-research form and extended with in-repo prior-art positioning. All external URLs were supplied with the research and are cited verbatim.

## §1 — git worktree mechanics

- `git worktree prune` removes only stale administrative metadata — it never inspects content and never deletes branches. `git worktree list --porcelain` exposes `locked` and `prunable` attributes per tree. `git worktree remove` refuses unclean trees without `--force`; a locked tree needs the flag twice; the main worktree cannot be removed. `git worktree lock` prevents deletion. `git merge-base --is-ancestor A B` yields a three-way exit signal: 0 = ancestor, 1 = not an ancestor, any other exit = error.
- **Consequence for the SPEC**: the three-way ancestry contract (REQ-WS-001/003) is git's own documented semantics, not an invented taxonomy. The lock attribute is readable from the porcelain the sweep already parses (`worktreeLockStates`).
- Sources: https://git-scm.com/docs/git-worktree , https://git-scm.com/docs/git-merge-base

## §2 — Claude Code worktree semantics (the platform we live in)

- Claude Code uses conservative, content-aware auto-cleanup: trees it "can't be verified" for prompt the operator rather than being removed. Subagent worktrees are removed automatically when unchanged; trees with changes stay until a periodic sweep can remove them without losing work. The sweep preserves trees with unpushed commits and only touches trees carrying Claude's own metadata marker. While an agent runs, Claude Code holds a `git worktree lock` and never releases user locks.
- **Consequence**: the fail-safe direction (unanswerable ⇒ preserve) matches the platform's own conservative default; MoAI's sweep generalizes it with explicit predicates so "can't be verified" becomes a named cause instead of a silent keep.
- Source: https://code.claude.com/docs/en/worktrees

## §3 — Prior art tools

- **autowt cleanup**: `--dry-run`, `--force` (dirty trees protected by default), cleanup modes (all / remoteless / merged / interactive / github), `-y` confirmation. **git-town**: ships and deletes branches as part of its sync flow.
- **Consequence**: dry-run-first and dirty-tree-protection-by-default are the established community shape; both are adopted (REQ-WS-012, REQ-WS-005). Branch deletion (git-town's model) is explicitly NOT adopted (REQ-WS-010) — this repo's branches carry card-id traceability and deletion is a separate concern.
- Sources: https://steveasleep.com/autowt/cli/autowt_cleanup/ , https://github.com/git-town/git-town

## §4 — Safety patterns

- **Dry-run first**: git itself ships `prune -n`; autowt ships `--dry-run`. Adopted as the default mode.
- **Dirty-tree gate**: `git status --porcelain` emptiness is the canonical gate (matches the repo's `worktreeHasLocalChanges`).
- **Process-cwd detection**: Linux `/proc/<pid>/cwd` readlink is ptrace-governed — an unreadable link means preserve, not "no process". `lsof +D <dir>` is depth-complete but slow and memory-heavy at scale (relevant at 992 trees): the repo's whole-system `lsof -a -d cwd -F0n` pattern (one invocation, path match afterwards) is preferred. Windows has no supported cwd API — preserve on Windows. Honor the porcelain `locked` attribute.
- Sources: https://man7.org/linux/man-pages/man5/proc_pid_cwd.5.html , https://man7.org/linux/man-pages/man8/lsof.8.html , https://learn.microsoft.com/en-us/windows/win32/api/winternl/ns-winternl-rtl_user_process_parameters

## §5 — In-repo prior art (positioning)

- **`internal/cli/worktree/clean.go` (`clean --stale`)** — sweeps clean, no-unique-commit trees; preview default; `--yes` applies; `--json` inventory is the same evaluation rendered twice (REQ-WR-012/013); ignored-content guard (REQ-WR-024) with removal-time re-read; exit-2 degraded contract (REQ-WR-016); `--base` defaults `origin/main` with a comment pinning why. **The sweep reuses its predicates and contracts but not its landing check** — local, unfetched, wrong base for this git-flow repo, and no three-way contract (spec.md §B).
- **`internal/cli/worktree/done.go`** — L2 disposal core with built-in hoist, anchor refusal, and the non-bypassable L1 tier guard (`isL1WorktreePath`, resolved from the target path). The sweep routes L2 trees here and L1 trees through hoist+remove.
- **`internal/cli/session_worktree.go`** — `gitHasUnpushedReal` (card t673): committed-but-unpushed work is invisible to `status --porcelain`; detached HEAD fails closed. Its resolution order (upstream → any-remote) is the preserve-side complement of the sweep's landing predicate.
- **`internal/cli/update_worktree_processes.go`** — the `activeProcessCWDs` lsof seam; duplicated small in the `worktree` package because importing the parent would cycle (design.md §C).
- **SPEC-WORKTREE-REAPER-001** (completed) — supplied the shared anchor decision (lock ∪ registry), the cause-token vocabulary, and the refusal-class requirement (REQ-WR-021: never unlock, never force). The sweep inherits all three by composition.
- **SPEC-WORKTREE-GC-001** (completed 09-23) — one-time disposal of ~530 trees with evidence-first export; accumulation resumed afterward, which is this card's premise. Its T1 rule (fresh fetch + `--is-ancestor` against `origin/develop`; t1059 staleness lesson) is REQ-WS-001's direct ancestor; its exclusion-set and measurement-hygiene disciplines (t1069) carry over operationally.
- **AutoCleanup PR-merge sweep** (`session_worktree_prmerge.go`) — gh-MERGED/local-merged signal, `WT-` branches only, default off, squash-merge blind. Positioning table: design.md §B.
- **Hoist obligation** (`.claude/rules/moai/workflow/worktree-integration.md` § Hoist, [HARD]) — `<tree>/.moai/reports/` is gitignored machine-local evidence; every disposal path must hoist first. REQ-WS-009 encodes this for the L1 path; REQ-WS-008 gets it from the done core.
