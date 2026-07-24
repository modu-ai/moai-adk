# Main-Checkout Branch Guard

Branch-state isolation rules for the primary project checkout. The checkout is **shared**: several Claude Code sessions, teammates, hooks, and background tools can operate on the same working tree at once. Branch state there is global — a `git switch` in one session changes what every other session sees, mid-operation, with no signal to either side.

> **Loading scope**: Intentionally always-loaded — the guard binds any turn that performs git work, which is not predictable from file paths.

## Why This Matters

Two properties combine badly:

1. **`HEAD` is shared mutable state.** A branch switch, reset, or stash in the primary checkout applies to every concurrent reader of that tree.
2. **A read of `HEAD` goes stale immediately.** The branch and commit observed at the start of a turn describe the tree *at that instant*, not at the moment of the next tool call.

The resulting race is quiet. A commit landing on the current branch from another session moves `HEAD` under a turn that already read it — a later `push` then ships more than the turn believed it was shipping. The same race with a *switch* instead of a commit is worse: another session's working tree changes shape mid-edit, and uncommitted work can end up attributed to the wrong branch.

Neither failure raises an error. Both surface later as "commits I did not make" or "my changes are on the wrong branch".

## Rules

[ZONE:Evolvable] [HARD] The orchestrator MUST NOT change branch state in the primary project checkout. Specifically forbidden there:

| Forbidden | Why |
|-----------|-----|
| `git checkout <branch>` / `git switch` | relocates every concurrent session's tree |
| `git checkout -b` / `git switch -c` / `git branch` | same, plus leaves a branch other sessions did not expect |
| `git reset --hard` / `git checkout -- <path>` | discards work the orchestrator cannot see the provenance of |
| `git stash` | the stash is repository-global; it silently absorbs other sessions' uncommitted changes |
| `git rebase` / `git merge` onto the checked-out branch | rewrites or advances shared history mid-operation |

[ZONE:Evolvable] Permitted in the primary checkout:

- Read-only inspection: `git status`, `git log`, `git diff`, `git rev-parse`, `git show`, `git branch -vv`
- `git fetch` (updates remote-tracking refs only; never touches the working tree)
- Commits **to the branch already checked out**, staged by explicit pathspec rather than `git add -A`
- `git push` of the already-checked-out branch

## Procedure — Isolate With a Worktree

When work needs a different branch, create a worktree instead of switching:

```bash
git worktree add -b <branch> <worktree-path> origin/main
git -C <worktree-path> add <paths>
git -C <worktree-path> commit -m "<message>"
git -C <worktree-path> push -u origin <branch>
```

Drive the worktree with `git -C <path>` rather than `cd`. A `cd` inside a compound command changes the shell's working directory for that invocation only, which makes subsequent commands read the wrong tree if the pattern is copied without the `cd`.

Remove the worktree when the branch is merged:

```bash
git worktree remove <worktree-path>
```

## Staleness Rule

[ZONE:Evolvable] [HARD] Re-read branch and commit state **immediately before** any commit or push — never rely on a value read earlier in the turn, and never on the branch reported in session-start context.

```bash
git rev-parse --short HEAD
git branch --show-current
```

If either differs from what the turn assumed, stop and report the divergence rather than proceeding. A moved `HEAD` means another actor is writing to the same tree, and the turn's plan was formed against a tree that no longer exists.

## Detecting Concurrent Sessions

Process-registry lookups are not a reliable emptiness signal — a registry can hold entries whose recorded PIDs no longer match live processes, including the querying session's own. An empty or all-stale registry result therefore does NOT establish that no other session is active, and MUST NOT be reported as such.

Treat concurrency as the default assumption. The load-bearing check is the staleness rule above: compare `HEAD` before and after, and let a moved `HEAD` be the evidence.

## Verification

```bash
# Confirm the intended tree before writing to it
git -C <worktree-path> rev-parse --show-toplevel
git -C <worktree-path> branch --show-current

# Confirm the push shipped exactly what was intended
git rev-list --count --left-right origin/<branch>...HEAD
```

## Cross-references

- `.claude/rules/moai/workflow/worktree-integration.md` — worktree systems, lifecycle, and the disposal contract
- `.claude/rules/moai/workflow/worktree-state-guard.md` — worktree state validation
- `.claude/rules/moai/core/agent-common-protocol.md` § Pre-Spawn Sync Check — divergence check before spawning a write-capable agent
- `.claude/rules/moai/core/verification-claim-integrity.md` — why an unobserved "no concurrent session" claim is a defect claim

---

Version: 1.0.0
Classification: Evolvable operational rule — branch-state isolation; changes no gate semantics.
