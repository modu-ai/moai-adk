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

## Mechanical Enforcement (v1.2.0)

The doctrine above is mechanically enforced by a PreToolUse hook. The
enforcer is NOT a static `settings.json` deny entry — a static deny cannot
scope to the primary checkout only and would lock out legitimate worktree
flows. The hook applies the doctrine conditionally.

- **Handler**: `internal/hook/pre_tool.go` `preToolHandler.Handle` calls
  `checkBranchState` (in `internal/hook/branch_guard.go`) after the existing
  dangerous-pattern check and before the default-allow fall-through.
- **Opt-in gate (v1.2.0)**: the call to `checkBranchState` is gated at the
  call site by the `Workflow.BranchGuard.Enabled` config flag (read via the
  handler's ConfigProvider). The default is **false** — the guard ships INERT
  to all users, because the shared-checkout hazard it addresses does not apply
  to single-developer repos. Maintainers of shared multi-session checkouts opt
  in via local config. When disabled, NO `git rev-parse` subprocess runs (the
  primary-checkout discriminant cost is avoided entirely) and the guard
  returns the allow fall-through without evaluating patterns, checkout state,
  or exemption logic. The exemption logic (`MOAI_BRANCH_GUARD_EXEMPT` +
  `manager-git` identity) is unchanged and is consulted only on the enabled
  path.
- **Pattern refinement (v1.2.0)**: the regex set no longer matches the
  read-only forms `git stash list`, `git stash show`, and `git merge-base`.
  `git merge` anchors on trailing whitespace so `merge-base` is excluded;
  `git stash` requires either bare end-of-input or one of the mutating
  subcommands (push/pop/apply/drop). The genuinely dangerous forms
  (switch/checkout/branch/reset --hard/rebase/bare+mutating stash/actual
  merge) remain matched.
- **Deny reason sentinel**: every deny emitted by this path carries the
  prefix `BRANCH_GUARD_VIOLATION:` so the orchestrator can pattern-match the
  source without parsing the full reason string.
- **Discriminant (primary vs worktree)**: the hook compares the absolute
  `git rev-parse --git-dir` against the absolute `git rev-parse
  --git-common-dir` **at the command's actual cwd**; equal paths classify as
  the primary checkout, differing paths as a worktree. The cwd is resolved
  from `input.CWD` (via `resolveProjectRootFromInputOrEnv`: `input.CWD`, then
  `$CLAUDE_PROJECT_DIR`, then `os.Getwd()`), NOT from `$CLAUDE_PROJECT_DIR`
  alone — querying the primary checkout about itself always answered
  "primary", which misclassified a worktree-resident agent's branch-state
  command as a primary-checkout violation (discriminant directory correction,
  v1.2.0). The audit-log project directory stays pinned to
  `$CLAUDE_PROJECT_DIR` → `os.Getwd()` for central logging, independent of
  which directory the git-context discriminant queries. Primary path uses
  `--path-format=absolute` (git 2.31+, March 2021); older git or Apple Git
  that rejects the flag falls back to `git rev-parse --absolute-git-dir` +
  cwd-normalized `--git-common-dir`.
- **Exemption mechanism**: the deny is suppressed when EITHER the invoking
  agent identity is the trusted git agent (`HookInput.AgentType ==
  "manager-git"`) OR the sentinel environment variable
  `MOAI_BRANCH_GUARD_EXEMPT=1` is present. Both axes are implemented and each
  fires on its own — but each is read from a different place, and **neither is
  reachable from inside a tool-spawned subagent**:
  - `AgentType` arrives in the hook payload, and Claude Code populates
    `agent_type` for a main-thread `claude --agent manager-git` launch. A
    subagent spawned through the Agent tool sends no `agent_type` on
    PreToolUse, so the identity axis cannot fire for it.
  - The sentinel is read from the hook process's own environment. The hook
    runs as a separate process spawned **before** the guarded command executes,
    so an `export MOAI_BRANCH_GUARD_EXEMPT=1` inside that command never reaches
    it. The variable must be present in the environment Claude Code itself was
    launched with.

  Exporting the sentinel inside the command being guarded is therefore a no-op.
  A `manager-git` subagent that needs to mutate branch state has two working
  routes: do the work in a worktree (`git -C <worktree>`, which the discriminant
  correctly classifies as non-primary), or have the operator launch the session
  with the sentinel already in its environment. Reading a `BRANCH_GUARD_VIOLATION`
  as "the exemption is broken" is a misdiagnosis — the axes work; the values were
  never delivered.

- **Scan scope**: the pattern set is matched against the command with quoted
  spans collapsed to a placeholder word, so a match reflects the command being
  invoked rather than text carried as data. `moai todo add "… git switch …"` is
  allowed because the command being run is `moai todo add`; `git switch main`
  and `git checkout -b "feat/x"` both still deny, the latter because the
  placeholder preserves the operand after `-b`. A git invocation hidden inside a
  shell wrapper (`bash -c "git switch main"`) is not matched — under-matching an
  obfuscated form is the correct direction to err for a fail-open guard.
- **Fail-open norm**: the deny fires ONLY on positive evidence (primary
  checkout confirmed AND a branch-state pattern matched AND the agent is not
  exempt). Any uncertainty — not a git repo, missing git binary, `git
  rev-parse` exiting non-zero, or indeterminate agent identity — falls
  through to allow, writes an advisory to stderr, and appends a structured
  entry to `.moai/logs/branch-guard-audit.log`. Aligns with the Bash
  Risk-Amplifier Doctrine (WARN-ONLY, FAIL-OPEN). This norm is unchanged by
  the opt-in gate: when disabled the guard returns allow BEFORE reaching any
  uncertainty path, so fail-open is trivially preserved.

Origin: the run-phase SPEC that landed the v1.1.0 mechanical enforcer. The
v1.2.0 opt-in gate and pattern refinement landed in a follow-up behavior-tuning
SPEC; the v1.2.0 discriminant directory correction (query the command cwd, not
$CLAUDE_PROJECT_DIR) landed in a second follow-up SPEC.

## Cross-references

- `.claude/rules/moai/workflow/worktree-integration.md` — worktree systems, lifecycle, and the disposal contract
- `.claude/rules/moai/workflow/worktree-state-guard.md` — worktree state validation
- `.claude/rules/moai/core/agent-common-protocol.md` § Pre-Spawn Sync Check — divergence check before spawning a write-capable agent
- `.claude/rules/moai/core/verification-claim-integrity.md` — why an unobserved "no concurrent session" claim is a defect claim

---

Version: 1.3.0
Classification: Evolvable operational rule — branch-state isolation; changes no gate semantics.
