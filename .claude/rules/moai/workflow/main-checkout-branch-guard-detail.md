---
description: "Detail companion for main-checkout-branch-guard.md — the shared-HEAD race rationale and the PreToolUse enforcer's implementation detail (opt-in gate, pattern set, primary-vs-worktree discriminant, exemption reachability, fail-open norm)"
paths: "**/main-checkout-branch-guard*.md,**/internal/hook/branch_guard.go,**/internal/hook/pre_tool.go"
---

# Main-Checkout Branch Guard — Detail Companion

> Detail companion of `main-checkout-branch-guard.md` (the always-loaded stub). The stub owns the
> prohibition table, the permitted-operations list, the worktree procedure, the staleness rule, and
> the verification commands. This file owns the rationale for why the race is quiet and the full
> implementation detail of the PreToolUse enforcer. Load it when diagnosing a
> `BRANCH_GUARD_VIOLATION` deny, or when changing the enforcer itself.

## Why the race is quiet

Two properties combine badly:

1. **`HEAD` is shared mutable state.** A branch switch, reset, or stash in the primary checkout applies to every concurrent reader of that tree.
2. **A read of `HEAD` goes stale immediately.** The branch and commit observed at the start of a turn describe the tree *at that instant*, not at the moment of the next tool call.

The resulting race is quiet. A commit landing on the current branch from another session moves `HEAD` under a turn that already read it — a later `push` then ships more than the turn believed it was shipping. The same race with a *switch* instead of a commit is worse: another session's working tree changes shape mid-edit, and uncommitted work can end up attributed to the wrong branch.

Neither failure raises an error. Both surface later as "commits I did not make" or "my changes are on the wrong branch".


## Mechanical enforcement — implementation detail (v1.2.0)

The doctrine above is mechanically enforced by a PreToolUse hook. The
enforcer is NOT a static `settings.json` deny entry — a static deny cannot
scope to the primary checkout only and would lock out legitimate worktree
flows. The hook applies the doctrine conditionally.

- **Handler**: `internal/hook/pre_tool.go` `preToolHandler.Handle` calls
  `checkBranchState` (in `internal/hook/branch_guard.go`) after the existing
  dangerous-pattern check and before the default-allow fall-through.
- **Opt-in gate (v1.2.0, SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001 REQ-1/REQ-3)**:
  the call to `checkBranchState` is gated at the call site by
  `Workflow.BranchGuard.Enabled` (read via the handler's ConfigProvider). The
  distributed default is **false** — the guard ships INERT to all users,
  because the shared-checkout hazard it addresses does not apply to
  single-developer repos. Maintainers of shared multi-session checkouts opt in
  via local config. When disabled, NO `git rev-parse` subprocess runs (the
  primary-checkout discriminant cost is avoided entirely) and the guard
  returns the allow fall-through without evaluating patterns, checkout state,
  or exemption logic. The exemption logic (`MOAI_BRANCH_GUARD_EXEMPT` +
  `manager-git` identity) is unchanged and is consulted only on the enabled
  path (REQ-6 backward compat).
- **Pattern refinement (v1.2.0, REQ-2)**: the regex set no longer matches the
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

  The deny reason's remediation text aligns with this reachability caveat
  (v1.3.1): it directs the caller to a worktree and states that the manager-git
  identity and sentinel exemptions fire only for main-thread launches — it must
  not suggest delegating to a `manager-git` subagent, which receives the same
  deny again.

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

Origin: SPEC-WORKTREE-BRANCH-GUARD-001 (REQ-WBG-001 through REQ-WBG-013).
Opt-in gate + pattern refinement: SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001
(REQ-1 through REQ-6).
Discriminant directory correction: SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001
(REQ-WBG-D-001 through REQ-WBG-D-008).


---

Classification: Lazy companion — rationale and implementation detail only. Every prohibition and
every permitted-operation clause stays in `main-checkout-branch-guard.md`.
