# t209 — worktree reaper: plan-phase investigation

Read-only survey run in `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t209`
(branch `WT-worktree-reaper`, base `cd0cee1b8` = `origin/main`), 2026-08-24.
Every number below was measured in this run with the command shown.

## The card's premise is half-wrong, and the correct half is already built

The dispatch asks me to design three axes — safe-disposal condition, L1/L2
disposal path, session-anchor detection. **All three already exist in the
codebase**, in `internal/cli/session_worktree_prmerge.go` (`prMergeCleanup`,
SPEC-SESSION-WORKTREE-001 M8), which runs automatically at `moai session
register` and `moai session list` and is **already enabled in this repo**
(`.moai/config/sections/workflow.yaml:124` `auto_cleanup: true`).

I watched it act during this very investigation — running `moai session list
--json` emitted:

```
moai: PR-merge cleanup skipped (uncommitted changes): worktree …/t192 preserved
moai: PR-merge cleanup skipped (uncommitted changes): worktree …/t206 preserved
moai: removed by PR-merge cleanup: [WT] worktree …/t208 (branch WT-profile-test-isolation merged)
moai: PR-merge cleanup failed (fatal: cannot remove a locked working tree, lock reason: claude session t212 (pid 31329 …)): worktree …/t212 left on disk
```

So the deliverable is **not a new reaper**. It is (a) closing the defect that
makes the existing one preserve almost everything, and (b) closing the safety
hole the same run exposed.

## Measured state

| Measurement | Command | Value |
|---|---|---|
| worktrees registered | `git worktree list \| wc -l` | **155** (154 + primary) |
| disk footprint | `du -sh .claude/worktrees` | **30G** |
| under `.claude/worktrees/` (L1) | `git worktree list --porcelain` parent-dir tally | **154** |
| under `~/.moai/worktrees/` (L2) | same | **0** |
| L2 registry contents | `cat ~/.moai/worktrees/MoAI-ADK/.moai-worktree-registry.json` | `{}` (2 bytes) |
| prunable (stale admin dirs) | `git worktree prune --dry-run` | **0** |
| locked worktrees | `grep -c '^locked'` on porcelain | **5** |
| `WT-*` branches on worktrees | porcelain branch tally | **111** |
| non-`WT-*` worktree branches | 154 − 111 | **43** |
| `WT-*` branches merged into `origin/main` | `git branch --merged origin/main \| grep -c '^WT-'` | **99** |

**L2 is empty.** Every worktree in this repo is L1. `moai worktree done` is
therefore not merely the wrong verb for these trees — it has *nothing to
operate on at all* in this repository. The only disposal path that applies is
`git worktree remove` (with `git worktree unlock` first where locked), which
is exactly what `prMergeCleanup` already calls.

## Defect 1 — `gh pr view` goes blind the moment the remote branch is deleted

`branchMergedForCleanup` (`session_worktree_prmerge.go:186`) takes the `gh`
path whenever `gh` is on PATH, and only falls back to `git branch --merged`
when `gh` is **absent**:

```go
if ghAvailable {
    return sessionWorktreeGhPRViewState(branch) == "MERGED"
}
```

`ghPRViewStateReal` returns `""` on any gh error — including the ordinary,
expected case of a merged PR whose head branch was deleted on the remote:

```
$ gh pr view WT-forge-counts --json state
no pull requests found for branch "WT-forge-counts"

$ git branch --merged origin/main --format='%(refname:short)' | grep -x 'WT-forge-counts'
WT-forge-counts
```

git says merged; gh says no PR; the reaper reads `""`, concludes "not merged",
and preserves the tree — permanently, on every future sweep. The fallback that
would have caught it is unreachable because `gh` IS installed.

This is why **99 merged `WT-*` worktrees are still on disk** while the sweep
runs on every `moai session list`. The mechanism is not a missing feature; it
is one branch of an if-statement that treats "gh could not tell me" and "gh
told me it is not merged" as the same answer.

Note the asymmetry that makes this safe to fix: the two sources fail in
opposite directions. `gh` sees squash merges that `git branch --merged` cannot
(the documented reason gh is primary); `git branch --merged` sees deleted-branch
merges that `gh` cannot. Consulting the second when the first returns *no
answer* — as distinct from a negative answer — loses neither property.

## Defect 2 — the anchor guard is 1-of-5 blind, and git's lock is what is actually protecting live sessions

`prMergeCleanup` calls `session.LiveAnchoredSessions(path, now)`
(`internal/session/anchor.go:49`), which reads the session registry
(`.moai/state/active-sessions.json`) and matches entry `cwd` against the tree.

Measured coverage:

| Anchor signal | Live anchors it names |
|---|---|
| git worktree lock reason (`locked claude session <name> (pid N …)`) | **5** — t207, t209, t210, t212, t213 |
| session registry `cwd` under a worktree | **1** — t207 only |

```
$ ps -o pid=,comm= -p 36912 -p 34699 -p 51045 -p 31329 -p 15207
51045 claude / 31329 claude / 34699 claude / 15207 claude / 36912 claude    # all 5 alive

$ grep '"cwd"' /Users/goos/MoAI/moai-adk-go/.moai/state/active-sessions.json
  /Users/goos/moai/moai-adk-go        (×2, not a worktree)
  /Users/goos/MoAI/moai-adk-go        (×2, the primary checkout)
  /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t207
```

The cause is documented in `anchor.go`'s own doc comment: a session that enters
a tree via `EnterWorktree` keeps its launch-time CWD until the `CwdChanged`
hook calls `RelocateSession`, and "before that relocation runs they stay
invisible here". Measured, that relocation is not happening for 4 of 5 live
lanes.

Worse, from inside a worktree the registry query returns nothing at all:

```
$ moai session list --json
[]
```

So the anchor test the dispatch proposed — `moai session list --json` cwd
comparison — would have reported **zero anchors** and marked every tree safe.
It is a false-negative machine, and it is the guard the shipped reaper relies on.

What actually saved a live session during this investigation was git, not the
guard: the sweep selected t212 for removal — the registry did not know its
session existed — and `git worktree remove` refused because the tree was
locked. The safety property currently rests on a side effect.

The lock is the better signal for a specific reason: Claude Code writes it at
`EnterWorktree` time, in the same act that anchors the session, and releases it
on `ExitWorktree`. It cannot drift from the thing it describes the way a
separately-maintained registry can. It carries the owning PID inline, so
liveness is a `kill -0` away, and a lock whose PID is dead is exactly the
"judgement impossible → treat as live" case the dispatch asked to preserve.

## Defect 3 — 43 worktrees are outside the reaper's field of view entirely

```go
if !strings.HasPrefix(e.branch, SessionWorktreeBranchPrefix) { continue }   // "WT-"
```

The 43 non-`WT-*` worktrees — `worktree-t*` (18), `worktree-agent-*` (7),
`worktree-<slug>` (~15), plus `feat/…`, `fix/…`, `card-wtiso`, `release/v*` —
are skipped unconditionally. `worktree-agent-*` are the auto-named trees left
by `Agent(isolation: "worktree")`; nothing else ever disposes of them.

The `WT-` prefix filter is not wrong as a *default* — it is what distinguishes a
tree the tooling created from one a human made deliberately — but it means the
long tail can only grow.

## The dispatch's verification sample

`WT-web-live-todo` = worktree `.claude/worktrees/t207`:

- **not merged** — appears in `git branch --no-merged origin/main`
- **anchored** — `locked claude session t207 (pid 36912 …)`, and `ps` confirms 36912 is a live `claude`
- **no PR** — nothing to mark it landed

Two independent conditions each say preserve, so any criterion built on either
axis judges it correctly. It is a good sample for the anchor axis specifically
because t207 is the *one* tree the registry does see — a criterion that passes
this sample on the registry alone would still be 4-of-5 blind on the others.

## Consequence being paid

154 live worktrees × a full checkout each = 30G on disk and 154 directory trees
under filesystem-event watch. Reported alongside: `fseventsd` RSS 25.5G, CPU
165%, swap 32.4/33.8G exhausted, event loop stalled 29s.

## What this makes the SPEC about

Not "build a reaper". Three narrower things, in descending order of value:

1. **Make the existing sweep see merged branches** — distinguish "gh gave no
   answer" from "gh said not merged", and consult `git branch --merged` in the
   first case. Unblocks 99 trees on the next sweep with no new surface.
2. **Make the anchor guard read the signal that is actually authoritative** —
   the worktree lock plus a PID liveness probe, with the registry retained as a
   supplementary source. Undetermined ⇒ anchored.
3. **Decide, explicitly, what happens to the 43 non-`WT-*` trees** — an opt-in
   wider sweep, or a reported-but-not-removed inventory. This is the part that
   genuinely needs a design decision rather than a repair.

Whether (3) ships as a tool or as a documented procedure is the open question
for the plan; (1) and (2) are repairs to shipped code and belong in it either way.
