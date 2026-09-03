# Main-Checkout Branch Guard

Branch-state isolation rules for the primary project checkout. The checkout is **shared**: several Claude Code sessions, teammates, hooks, and background tools can operate on the same working tree at once. Branch state there is global — a `git switch` in one session changes what every other session sees, mid-operation, with no signal to either side.

> **Loading scope**: Intentionally always-loaded — the guard binds any turn that performs git work, which is not predictable from file paths.

## Why This Matters

`HEAD` is shared mutable state and a read of it goes stale immediately, so a branch switch, reset,
or stash in the primary checkout reaches every concurrent reader mid-operation. Neither resulting
failure raises an error; both surface later as "commits I did not make" or "my changes are on the
wrong branch". The full mechanism: `main-checkout-branch-guard-detail.md` § Why the race is quiet.

## Rules

[ZONE:Evolvable] [HARD] The orchestrator MUST NOT change branch state in the primary project checkout. Specifically forbidden there:

| Forbidden | Why |
|-----------|-----|
| `git checkout <branch>` / `git switch` | relocates every concurrent session's tree |
| `git checkout -b` / `git switch -c` / `git branch <name>` / mutating `git branch` flags: `-d|-D|-m|-M|-c|-C|-f|-u|-t` and their long forms (`--force`/`--delete`/`--move`/`--copy`/`--set-upstream-to`/`--unset-upstream`/`--track`/`--edit-description`) and combined clusters (`-df`, `-vD`, `-vux`) | same, plus leaves a branch other sessions did not expect |
| `git reset --hard` / `git checkout -- <path>` | discards work the orchestrator cannot see the provenance of |
| `git stash` | the stash is repository-global; it silently absorbs other sessions' uncommitted changes |
| `git rebase` / `git merge` onto the checked-out branch | rewrites or advances shared history mid-operation |

[ZONE:Evolvable] Permitted in the primary checkout:

- Read-only inspection: `git status`, `git log`, `git diff`, `git rev-parse`, `git show`; `git branch` queries (bare list, `--list`, `-v`/`-vv`, `--show-current`, `--contains`/`--merged`/`--points-at`)
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

## Mechanical Enforcement

A PreToolUse hook applies this doctrine conditionally — a static `settings.json` deny cannot scope
to the primary checkout and would lock out legitimate worktree flows.

- **Opt-in, inert by default.** `Workflow.BranchGuard.Enabled` gates the call; the distributed
  default is `false`, because the shared-checkout hazard does not apply to single-developer repos.
  Disabled, no `git rev-parse` subprocess runs at all.
- **Deny sentinel.** Every deny on this path is prefixed `BRANCH_GUARD_VIOLATION:`, so the
  orchestrator can match the source without parsing the reason string.
- **Query-vs-mutate discrimination.** The `git branch` matcher denies every mutating form: a
  mutating flag anywhere — `-f`/`--force`, `-d`/`--delete`, `-m`/`--move`, `-c`/`--copy` (and
  their uppercase forms), `-u`/`--set-upstream-to`/`--unset-upstream`, `-t`/`--track`,
  `--edit-description`, or a short-flag cluster containing any of `d/D/m/M/c/C/f/t/u` (`-df`,
  `-vD`, `-vux`) — or a positional branch-name operand with no list action selected (bare and
  option-prefixed creation alike: `git branch <name>`, `git branch -q <name>`, `git branch
  --no-force <name>` — a query flag plus a name operand creates a branch). Read-only queries
  pass: bare `git branch`, `--list`/`-l` (mid-cluster too, e.g. `-al`), operand-free `-v`/`-vv`/
  `-a`/`-r`, `--show-current`, and the filter/format/sort flags with their operands
  (`--contains HEAD main` is a filter pattern, not a creation). Unclassifiable forms — git
  prefix-abbreviations like `--dele` — under-match and pass; under-matching an unclassifiable
  form is the accepted fail-open direction.
- **Fail-open.** The deny fires only on positive evidence — primary checkout confirmed, a
  branch-state pattern matched, agent not exempt. Any uncertainty falls through to allow and
  appends to `.moai/logs/branch-guard-audit.log`.
- **Exemptions are unreachable from a tool-spawned subagent.** Both axes work, but neither value
  reaches one: `AgentType` is populated only for a main-thread `claude --agent manager-git` launch,
  and `MOAI_BRANCH_GUARD_EXEMPT=1` is read from the hook process's own environment, which is spawned
  before the guarded command runs. Exporting it inside that command is a no-op. Reading a
  `BRANCH_GUARD_VIOLATION` as "the exemption is broken" is a misdiagnosis — use a worktree instead.

Pattern set, the primary-vs-worktree discriminant, quoted-span scan scope, and the originating SPEC
IDs: `main-checkout-branch-guard-detail.md` § Mechanical enforcement.

## Cross-references

- `.claude/rules/moai/workflow/worktree-integration.md` — worktree systems, lifecycle, and the disposal contract
- `.claude/rules/moai/workflow/worktree-state-guard.md` — worktree state validation
- `.claude/rules/moai/core/agent-common-protocol.md` § Pre-Spawn Sync Check — divergence check before spawning a write-capable agent
- `.claude/rules/moai/core/verification-claim-integrity.md` — why an unobserved "no concurrent session" claim is a defect claim
- SPEC-WORKTREE-BRANCH-GUARD-001 — the run-phase SPEC that landed the v1.1.0 mechanical enforcer
- SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001 — the discriminant directory correction (query input.CWD, not $CLAUDE_PROJECT_DIR)

---

Version: 1.3.3
Classification: Evolvable operational rule — branch-state isolation; changes no gate semantics.
