# Main-Checkout Branch Guard

Branch-state isolation rules for the primary project checkout. The checkout is **shared**: several Claude Code sessions, teammates, hooks, and background tools can operate on the same working tree at once. Branch state there is global — a `git switch` in one session changes what every other session sees, mid-operation, with no signal to either side.

> **Loading scope**: Intentionally always-loaded — the guard binds any turn that performs git work, which is not predictable from file paths.

## Rules

[ZONE:Evolvable] [HARD] The orchestrator MUST NOT change branch state in the primary project checkout. Specifically forbidden there:

| Forbidden | Why |
|-----------|-----|
| `git checkout <branch>` / `git switch` | relocates every concurrent session's tree |
| `git checkout -b` / `git switch -c` / `git branch <name>` / any mutating `git branch` form (flag classification: § Mechanical Enforcement below) | same, plus leaves a branch other sessions did not expect |
| `git reset --hard` / `git checkout -- <path>` | discards work the orchestrator cannot see the provenance of |
| `git stash` | the stash is repository-global; it silently absorbs other sessions' uncommitted changes |
| `git rebase` / `git merge` onto the checked-out branch | rewrites or advances shared history mid-operation |

[ZONE:Evolvable] Permitted in the primary checkout:

- Read-only inspection: `git status`, `git log`, `git diff`, `git rev-parse`, `git show`; `git branch` queries (bare list, `--list`, `-v`/`-vv`, `--show-current`, `--contains`/`--merged`/`--points-at`)
- `git fetch` (updates remote-tracking refs only; never touches the working tree)
- Commits **to the branch already checked out**, staged by explicit pathspec rather than `git add -A`
- `git push` of the already-checked-out branch

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

## Mechanical Enforcement

A PreToolUse hook applies this doctrine conditionally — a static `settings.json` deny cannot scope
to the primary checkout and would lock out legitimate worktree flows.

- **Opt-in, inert by default.** `Workflow.BranchGuard.Enabled` gates the call; the distributed
  default is `false`, because the shared-checkout hazard does not apply to single-developer repos.
  Disabled, no `git rev-parse` subprocess runs at all.
- **Deny sentinel.** Every deny on this path is prefixed `BRANCH_GUARD_VIOLATION:`, so the
  orchestrator can match the source without parsing the reason string.
- **Query-vs-mutate discrimination.** The `git branch` matcher denies every mutating form and
  passes read-only queries; an unclassifiable form under-matches and passes, under-match being the
  accepted fail-open direction.
- **Fail-open.** The deny fires only on positive evidence — primary checkout confirmed, a
  branch-state pattern matched, agent not exempt. Any uncertainty falls through to allow and
  appends to `.moai/logs/branch-guard-audit.log`.
- **Exemptions are unreachable from a tool-spawned subagent.** Both axes work, but neither value
  reaches one, so exporting `MOAI_BRANCH_GUARD_EXEMPT=1` inside the guarded command is a no-op.
  Reading a `BRANCH_GUARD_VIOLATION` as "the exemption is broken" is a misdiagnosis — use a
  worktree instead.

Flag classification, the pattern set, the primary-vs-worktree discriminant, and quoted-span scan
scope: `main-checkout-branch-guard-detail.md` § Mechanical enforcement.

## Cross-references

- `.claude/rules/moai/workflow/worktree-integration.md` — worktree systems, lifecycle, and the disposal contract
- `.claude/rules/moai/core/agent-common-protocol.md` § Pre-Spawn Sync Check — divergence check before spawning a write-capable agent
- `main-checkout-branch-guard-detail.md` — the lazy companion. Load it for § Why the race is quiet (relocated § Why This Matters) · § Mechanical enforcement · § Procedure — Isolate With a Worktree · § Verification

---

Version: 1.3.3
Classification: Evolvable operational rule — branch-state isolation; changes no gate semantics.
