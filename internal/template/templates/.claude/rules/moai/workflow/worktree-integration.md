---
paths: "**/.claude/agents/**,**/.claude/worktrees/**,**/.claude/teams/**"
---

# Worktree Integration Guide

Integration guide for MoAI Worktree and Claude Code Native Worktree systems.

## Overview

MoAI-ADK supports two complementary worktree systems for isolated development:

**Claude Code Native Worktree** (`.claude/worktrees/`):
- Ephemeral, session-scoped isolation
- Automatic cleanup when session ends
- Used for subagent isolation via `isolation: worktree` in agent definitions (v2.1.49+)
- CLI access: `claude --worktree` or `claude -w` (user-level flag)

**MoAI Worktree** (`~/.moai/worktrees/{ProjectName}/`):
- Persistent, SPEC-scoped workspaces in global home directory
- Managed via `moai worktree` CLI commands
- Used for multi-session SPEC development and team collaboration

## Comparison Table

| Feature | Claude Native | MoAI |
|---------|--------------|------|
| **Path** | `.claude/worktrees/<name>/` | `~/.moai/worktrees/{Project}/{SPEC}/` |
| **Lifetime** | Ephemeral (session-scoped) | Persistent |
| **Purpose** | Session isolation for subagents | SPEC development, PR creation |
| **CLI** | `claude -w` (user) or `isolation: worktree` (agent) | `moai cc -w <name>` to enter; `moai worktree clean/done/remove` to dispose |
| **Cleanup** | Automatic on session end | Manual via `moai worktree remove` |
| **Branch Strategy** | Temporary branches | Feature branches linked to SPEC |
| **Team Use** | Single agent isolation | Multi-developer collaboration |
| **State Persistence** | None | SPEC state, progress tracking |
| **Hook Support** | WorktreeCreate/WorktreeRemove hooks | WorktreeCreate/WorktreeRemove hooks |

## Terminology Glossary

This glossary is the canonical definition surface for the L1 / L2 worktree-layer terms used across the MoAI rule set. (The former L3 "launch action" tier is retired with `/moai plan --worktree`; a worktree is now entered, not provisioned by a workflow step.) Other rules (`spec-workflow.md`, `worktree-state-guard.md`, `session-handoff.md`, and `CLAUDE.md` §14) cross-reference `§ Terminology Glossary` for these definitions.

| Layer | Name | What it is | Path / Trigger | Lifetime | Owner |
|-------|------|-----------|----------------|----------|-------|
| **L1** | Session worktree | Session-scoped isolation created by `moai worktree new <name>`, a Claude launcher/tool, or an isolated subagent. Entered by short name through `moai cc -w <name>`, `moai codex -w <name>`, `claude -w <name>`, or `EnterWorktree(<name>)`. | `.claude/worktrees/<name>/`; harness-created branch names vary, while kanban/team card worktrees rename to `WT-<slug>` per the rule below | Session-scoped; disposed via the session-end keep/remove prompt, or `git worktree unlock` + `git worktree remove` once the session is done | Creating harness or MoAI shared materializer. L1 trees are not in the L2 lifecycle registry, so `done` / `clean` / `recover` do not close them |
| **L2** | MoAI persistent SPEC worktree | A persistent, SPEC-scoped working directory entered **by absolute path** — `moai cc -w ~/.moai/worktrees/<project>/<SPEC>`. Used for multi-session SPEC development (run + sync phases reuse the same L2 worktree). | `~/.moai/worktrees/<project>/<SPEC>/` | Persistent — lifecycle owned by the `moai worktree` verbs (`sync`, `remove`, `clean`, `recover`, `done`, plus the guard trio `snapshot` / `verify` / `restore`); disposed only via `moai worktree done SPEC-XXX` after both run + sync PRs merge | MoAI (user-managed via `moai worktree` CLI) |

Relationships:
- `moai worktree new <name>` creates an **L1** tree through the current shared materializer and returns its absolute path; it never enters the tree and carries none of the retired `--base`, `--from-current`, BODP, tmux, or team behavior. A **short name** passed to a launcher `-w` resolves that L1 tree; an **L2** persistent worktree is entered by absolute path. The separate `moai cc -w <name> --branch <existing>` form remains the existing-branch path used for the gitflow integration worktree.
- An **L1** ephemeral worktree is materialized autonomously by the Claude Code runtime for an isolated subagent; it is independent of L2 and may occur inside either the main checkout or an L2 worktree.
- When work happens inside an L2 worktree, the paste-ready resume MUST anchor the next session there (Block 0) per `session-handoff.md` § Worktree-Anchored Resume Pattern.

[HARD] **`moai worktree new` is the sole L1 exception; `done` refuses L1 by code.** An L1 tree under `.claude/worktrees/` is never registered in the L2 lifecycle registry. `done` refuses L1 targets in code (tier guard) — with or without `--force` — and `recover` is unchanged by that guard. `clean --merged-only`'s WT- sweep (below) is the documented exception to the L2-only lifecycle. L1 disposal is the session-end keep/remove prompt, or `git worktree unlock` + `git worktree remove` after the session releases its lock.

[HARD] **An unpushed worktree branch is the work's only instance.** Create an L1 tree with `moai worktree new <name>` or a supported native launcher/tool, then enter it through a launcher — never use bare `git worktree add`. Until its branch has been integrated and the remote merge has landed, dispose of no worktree.

[HARD] **Kanban/team card worktree branches carry the `WT-` prefix followed by a descriptive slug.** `EnterWorktree(<name>)` auto-names its branch `worktree-<name>`; for card worktrees, rename immediately after creation with `git branch -m WT-<slug>` (renaming the checked-out branch inside a worktree is safe — the tree, its lock, and the session anchoring are unaffected — and `moai cc -w <name>` re-entry resolves by tree name, not branch name). `WT-` is the session-worktree branch convention (`SessionWorktreeBranchPrefix`, `internal/cli/session_worktree.go`).

[HARD] **The slug describes the change; the card id stays out of the branch name.** At most 3 hyphen-separated tokens, at most 24 characters, lowercase `a-z0-9-` — `WT-branch-naming`, not `WT-t0`. The **worktree directory** still carries the card id (`.claude/worktrees/<card-id>`), which is what the disposal tooling and the evidence path key on, so the id is never lost — it simply stops living in the branch name. Traceability moves onto the dispatch `card:` field, the commit messages, and the evidence path; the full contract is `kanban-dispatch.md` § Isolation is provisioned by MoAI, then entered through a launcher.

Nothing reads a card id back out of a branch name: `internal/cli/session_worktree_prmerge.go` matches the `WT-` prefix only (`strings.HasPrefix`), never the remainder. The prefix is load-bearing; the suffix is for humans.

The rename is also a disposal-path switch, and that is deliberate:

- Left as `worktree-<name>`, the tree is **invisible to the PR-merge auto-cleanup sweep** — that sweep enumerates `git worktree list` and considers only `WT-` branches — so disposal stays manual: the session-end keep/remove prompt, or `git worktree unlock` + `git worktree remove`.
- Renamed to `WT-<slug>`, the tree becomes a **sweep candidate**: where `Workflow.Worktree.AutoCleanup` is enabled (distributed default: off), the sweep removes a `WT-` worktree once its branch reads merged (gh `MERGED` state, or the `git branch --merged origin/main` fallback — squash-merge blind) and the tree is clean, re-checking dirtiness immediately before removal, and never while a live session is anchored in the tree.

Either way the unpushed-branch rule above still governs timing — the sweep's merged-branch condition is the same "after the remote merge" boundary. The lane-side procedure that consumes `WT-` branches lives in `kanban-dispatch.md` § Integration into the release branch is self-served.

## Disposing a Worktree the Automatic Sweep Does Not Reach

Automatic disposal covers one shape only: the PR-merge auto-cleanup sweep enumerates `git worktree list` and treats a tree as a candidate solely when its branch carries the launcher's `WT-` prefix. Every other registered worktree — one named after the change it makes, one entered by hand, one whose branch was renamed — falls outside that sweep and stays on disk until someone disposes of it. That is the safe direction (nothing is removed unasked), but it is not a disposal plan.

**Worktree-ness is a property of the checkout, not of the branch name.** A branch-name glob finds only the trees named a particular way; `git worktree list` finds all of them. Any inventory of what is actually on disk therefore starts from the listing, never from a name pattern.

The shipped inventory is the `--stale` sweep's own evaluation, rendered as data:

```bash
moai worktree clean --stale --json
```

It emits one object per non-protected registered worktree, carrying the path, the branch, the keep-reason, and the four predicates behind that reason — dirty state, merge state, anchor state, and ignored-content state. It removes nothing: `--json` is a report, and it overrides `--yes` rather than combining with it. A predicate the sweep short-circuited before asking reads `not-checked`, which is deliberately distinct from `undetermined` — the latter means it asked and could not tell. Neither is a negative.

Read the report, then dispose of what it shows as removable:

```bash
moai worktree clean --stale        # preview: names the trees it would remove
moai worktree clean --stale --yes  # perform the removals
```

Both paths honour the same guards: a dirty tree, an unmerged branch, a tree anchoring a live session, a tree holding gitignored content that nothing regenerates, and a tree whose state could not be read are each kept and reported with the reason. The ignored-content guard matters because `git status --porcelain` and a non-forced `git worktree remove` both disregard gitignored files: without it a tree whose only remaining content is agent memory reads as clean and is destroyed silently. It shares one allowlist with the automatic sweep, so both agree on what is regenerable — runtime state, runtime-managed config, build output, test residue — and anything unclassified keeps the tree. Branches are never deleted — the commits stay reachable by branch name after the directory is gone. The merge comparison is against `origin/main` by default, the same ref the automatic sweep uses, so the two cannot reach opposite conclusions about the same tree; `--base` overrides it.

For a tree outside the sweep entirely, the manual path is unchanged: `git worktree unlock <path>` when a dead session's lock is still on it, then `git worktree remove <path>`. The unpushed-branch rule above governs the timing in every case.

## Claude Code 2.1.50+ Worktree Features

### `claude --worktree` (`-w`) Flag

For users starting isolated sessions:

```bash
# Start new isolated session in worktree
claude --worktree

# With custom name
claude --worktree my-feature

# With tmux for split-pane display (tmux or iTerm2 required)
claude --worktree --tmux
```

Behavior:
- Creates `.claude/worktrees/<name>/` automatically
- Branches from default remote branch
- On session end: prompts to keep (with commits) or auto-deletes (no changes)

tmux flag notes:
- Requires tmux or iTerm2
- NOT supported in VS Code integrated terminal, Windows Terminal, or Ghostty
- Useful for parallel team mode where viewing multiple teammates' output is beneficial

### `isolation: worktree` in Agent Frontmatter

For agents that need isolated execution (v2.1.49+):

```yaml
---
name: my-implementer
isolation: worktree   # Agent runs in its own isolated worktree
background: true      # Agent runs without blocking main conversation
---
```

When to use `isolation: worktree`:
- Implementation teammates that write files (write-capable implementation roles: implementer / tester / designer)
- Prevents file conflicts between parallel teammates
- Each agent gets its own clean worktree at `.claude/worktrees/<auto-name>/`

When NOT to use `isolation: worktree`:
- Read-only teammates (read-only research/review roles: researcher / analyst / reviewer)
- `permissionMode: plan` already prevents writes; adding isolation adds overhead without benefit

#### L1 ephemeral vs L2 persistent — `isolation: worktree` is NOT a re-entry mechanism

`Agent(isolation: "worktree")` creates a NEW **L1 ephemeral** worktree scoped to a single subagent invocation under `.claude/worktrees/<auto-name>/`. It is categorically distinct from an **L2 persistent** SPEC worktree entered with `moai cc -w <name>`. Conflating the two produces the worktree-masked flaky failure mode documented in `worktree-state-guard.md` — an L1 ephemeral worktree diverges from the L2 base and silently breaks parallel-session coordination.

`Agent(isolation: "worktree")` is NOT a re-entry mechanism for existing L2 persistent worktrees. To re-enter an existing worktree:
- **Current-session re-entry** (no `/clear`, same session continuing): use the Claude Code runtime tool `EnterWorktree(<path>)` — see `EnterWorktree` / `ExitWorktree` Tools below.
- **New-session launch** (post-`/clear` or new terminal): use the launcher flag `moai cc -w <name-or-abs-path>` (see the `-w` L2 absolute-path extension below).

### `background: true` in Agent Frontmatter

Run agent without blocking the main conversation (v2.1.46+):

```yaml
---
name: team-coder
background: true   # Returns immediately; results delivered on next turn
---
```

Use with `isolation: worktree` for optimal parallel execution in team mode.

Background-execution policy for write-capable agents is owned by `.claude/rules/moai/core/agent-common-protocol.md` § Background Agent Execution. As of Claude Code v2.1.198 subagents run in the background by default, and a background write surfaces a permission prompt in the main session naming the asking subagent; MoAI aligns with that runtime default rather than forcing foreground. The retained safeguard is concurrency, not backgrounding, and it is scoped to the working tree: one writer per tree — write-capable agents run in parallel only when each writes an independent worktree, with shared-path writes and integration serialized. Use `background: true` for:
- Read-only research and analysis agents
- Agents whose write paths are pre-approved in settings.json `permissions.allow`

Kill background agent: Press `Ctrl+X Ctrl+K` in Claude Code interface (v2.1.83+).

### Worktree Base Branch (`worktree.baseRef`)

Native worktrees (`--worktree` and subagent `isolation: worktree`) branch from the repository's default branch (`origin/HEAD`) by default, so they start from a clean tree matching the remote. If no remote is configured or the fetch fails, the worktree falls back to the current local `HEAD`. To always branch from local `HEAD` instead (carrying unpushed commits and feature-branch state), set `worktree.baseRef` to `"head"` in settings (accepts only `"fresh"` or `"head"`, not arbitrary refs):

```json
{
  "worktree": {
    "baseRef": "head"
  }
}
```

Use `"head"` when isolating subagents that must operate on in-progress work. To branch a native worktree from a specific pull request, pass the PR number prefixed with `#` (e.g. `claude --worktree "#1234"`); Claude Code fetches `pull/<number>/head` and creates the worktree at `.claude/worktrees/pr-<number>`.

This setting governs **Claude-native** worktrees only, which is now every worktree the launcher creates — `moai cc -w <name>` passes `-w` straight through to `claude`.

**Creating a card worktree on a release branch.** Because the default is `origin/HEAD`, a worktree created while a release branch is the intended base starts behind it — measured on one such branch, 34 commits behind, with the reflog reading `branch: Created from origin/main`. Two ways out, and the difference between them is not convenience:

- **Set the base after creation**, while the new tree still has no commits of its own: `git -C <tree> reset --hard <ref>`, then `git -C <tree> merge-base --is-ancestor <ref> HEAD` and read the exit code. Only for a tree that is empty and clean; once it carries a commit, this discards it.
- **Set `baseRef` to `"head"`**, which makes creation inherit rather than default.

`"head"` carries a trap worth stating plainly, because the name invites the wrong reading: it is **the HEAD of the tree the launcher ran in**, not the branch you had in mind. Run `moai cc -w <name>` from the primary checkout while intending a release branch and the new worktree inherits whatever the primary happens to be sitting on — wrong in a quieter direction than the default was, because the reflog now names a plausible commit instead of an obviously-unrelated one. Using it correctly therefore adds a procedural requirement of its own: the launcher must be run from inside a tree already on the intended base.

That asymmetry is the reason to prefer the reset path for card work. `baseRef` is silent whether it lands right or wrong; the reset path ends in an exit code somebody read.

**The stored setting: `git_strategy.worktree_base_branch`.** `baseRef` accepts only `"fresh"` or `"head"`, so it cannot name a branch — which leaves the branch choice resting on `refs/remotes/origin/HEAD`, local repository metadata that does not survive a fresh clone. The moai setting `git_strategy.worktree_base_branch` (in `.moai/config/sections/git-strategy.yaml`) is the reproducible handle on that choice, and it has two consumers:

- **At session start**, from the primary checkout only, moai points `refs/remotes/origin/HEAD` at the configured branch and prints one line saying it did. Native worktrees created afterwards read the corrected symref. Inside a linked worktree the step does nothing at all — the symref is repository-global while the config file is tracked and follows each worktree's own branch, so one writer is both sufficient and the only way two lanes do not reverse each other's writes forever.
- **When MoAI creates the worktree itself** (`moai worktree new <name>` or the creating `moai cc -w <name>` form), the configured branch is passed to the shared `git worktree add` plumbing as the base operand, so the new tree is cut from it rather than from the invoking tree's HEAD. This half honours the setting from any working tree.

The empty value — the shipped default — means take no action on both paths, reproducing the pre-setting behaviour exactly. A value naming a branch that has no remote-tracking counterpart is refused before either write: the session-start step prints one diagnostic line and leaves the symref alone, and worktree creation falls back to the no-operand form. Pointing `refs/remotes/origin/HEAD` at a ref that does not exist would be worse than the mismatch it was meant to fix.

`moai doctor --check 'Worktree Base Branch'` reports the current comparison without writing anything, and distinguishes a plain mismatch (repaired by running the alignment) from an unresolvable value (repaired by correcting the setting). It reports metadata state only — worktrees already cut from the wrong base are not re-created by it.


### `.worktreeinclude` (Copy Gitignored Files into Native Worktrees)

A native worktree is a fresh checkout, so untracked files (`.env`, `.env.local`, local config) are not present. Add a `.worktreeinclude` file at the project root to copy them automatically when Claude creates a worktree. It uses `.gitignore` syntax; only files that match a pattern AND are gitignored are copied (tracked files are never duplicated):

```text
.env
.env.local
.moai/config/sections/*.local.yaml
```

Applies to `claude --worktree`, subagent `isolation: worktree` worktrees, and desktop parallel sessions. NOT processed when a custom `WorktreeCreate` hook replaces the default git behavior — copy local files inside the hook script instead.

### `EnterWorktree` / `ExitWorktree` Tools

`EnterWorktree(<path>)` is the canonical mechanism for entering an existing worktree in the current session. The orchestrator's emitted guidance (paste-ready resume messages, Block 0 of the Worktree-Anchored Resume Pattern, in-session instructions) SHALL use `EnterWorktree(<path>)` for current-session worktree re-entry, replacing the shell-`cd`, `git -C <path>`, and subshell-`cd` patterns. A bare `cd` instruction SHALL NOT appear in orchestrator-emitted current-session-entry guidance; it remains valid only for human-typed, manual-shell contexts.

Claude can move the session into a worktree mid-session via the `EnterWorktree` tool (e.g. when the user says "work in a worktree"), creating one under `.claude/worktrees/`. Once inside, Claude can switch directly to another worktree by calling `EnterWorktree` with a target path; the previous worktree stays on disk untouched. `ExitWorktree` returns to the originating checkout. These are Claude Code runtime tools — MoAI does not mandate their use; they are the interactive counterpart to the launcher `-w` flag and `isolation: worktree` frontmatter. On a harness without these runtime tools, the harness-neutral equivalents are the launcher `-w` forms for new-session entry (`moai cc -w <name>`, `moai glm -w <name>`, `moai codex -w <name>`; after `moai worktree new <name>` for a fresh L1 tree) and, for current-session entry where no runtime tool exists, `git -C <path>` per the entry-guidance scoping below.

`EnterWorktree` is complementary to, not replaced by, the launcher flag `moai cc -w <name-or-abs-path>`:

- **`EnterWorktree(<path>)`** — current-session re-entry (no `/clear`, same session continuing). Use this when the orchestrator is mid-turn and needs to move the active session into an existing worktree.
- **`moai cc -w <name>` (or `moai glm -w`)** — new-session launch (post-`/clear` or new terminal). Use this as the Block 0 new-terminal launcher of the paste-ready resume. The `-w` flag accepts BOTH short names (resolved against `.claude/worktrees/<name>/`) AND absolute paths under `~/.moai/worktrees/<project>/...` (L2 persistent worktrees — see the `claude --worktree` (`-w`) Flag section above for the L2 absolute-path extension).

The shell-`cd` form (`cd <path> && <launcher>`), the `git -C <path>` form, and the subshell-`cd` form (`(cd <path> && ...)`) are DEPRECATED for orchestrator-emitted current-session worktree entry guidance on harnesses that carry a native current-session entry tool (Claude Code: `EnterWorktree`). They break `Agent(isolation: "worktree")` CWD isolation (the agent's CWD is the worktree root; a `cd /absolute/path` bypasses it) and were the root cause of prior incidents where a sub-agent used `git -C` instead of `EnterWorktree` and was corrected mid-run. On a harness without a native current-session entry tool (Codex), `git -C <path>` remains the documented means of driving a worktree from the current session — the root `AGENTS.md` worktrees contract mandates exactly that form — and is not deprecated there.

## Worktree Selection Rules [ZONE:Evolvable] [HARD]

### Decision Tree

```
Is this a parallel write workers within a hierarchical team (e.g., manager-lead fan-out)?
  YES → Use Agent(isolation: "worktree") for write agents
        Do NOT use isolation for read-only agents
  NO ↓

Is this a multi-session SPEC development?
  YES → Enter a worktree: moai cc -w <name>
  NO ↓

Is this a user-initiated parallel session?
  YES → Use claude --worktree (-w)
  NO ↓

Is this a one-shot sub-agent task?
  YES → Use Agent(isolation: "worktree") if agent writes files
        Use Agent() without isolation if agent is read-only
  NO → No worktree needed
```

### HARD Rules

- [ZONE:Evolvable] [HARD] Implementation leaf workers spawned in parallel by `manager-lead` (or any parallel-write fan-out shape) MUST use `isolation: "worktree"` when spawned via Agent()
- [ZONE:Evolvable] [HARD] Read-only teammates (read-only research/review roles: researcher / analyst / reviewer) MUST NOT use `isolation: "worktree"` — read-only enforcement rests on tool restriction, and a teammate is read-only only when **no tool in its list can write**. Omitting `Write`/`Edit` is necessary but NOT sufficient: `Bash`, a write-capable MCP tool, and `Agent` each reach the working tree on their own, and the built-in `Explore` omits `Write`/`Edit` yet carries `Bash`. The spawn-time `mode` parameter is deprecated and ignored since Claude Code v2.1.213, so tool restriction is the only channel that carries the guarantee — audit the list against all three write paths before calling a teammate read-only. Per-path detail: `.claude/rules/moai/development/agent-authoring.md` § Tool Permissions
- [ZONE:Evolvable] [HARD] One-shot sub-agents that write files across 3 or more paths per invocation MUST use `isolation: "worktree"`. This includes write-heavy retained agents (manager-develop), per-spawn `Agent(general-purpose)` specialists with a write-heavy domain whitelist (e.g. backend / frontend / devops / refactoring), and team-mode role profiles (implementer, tester, designer).
<!-- @MX:ANCHOR: WorktreeMUSTRule — invariant contract; all write-heavy agents MUST declare isolation:worktree; enforced by LR-05 lint rule -->
<!-- @MX:REASON: MUST level required to eliminate silent file-write conflict failure mode in parallel Agent() execution. -->
- [ZONE:Evolvable] [HARD] GitHub workflow agents (fixer agents in /moai github issues) MUST use `isolation: "worktree"` for branch isolation

### Parallel-Session Branch Conflict Auto-Isolation

[ZONE:Evolvable] [HARD] **When** the orchestrator detects (via the Pre-Spawn Sync Check active-sessions registry OR the Pre-Edit Sync Check, `.moai/state/active-sessions.json`) that ≥1 foreign active session is on the same checkout during **any write work** — whether worktree entry was chosen OR the orchestrator is editing the shared tree directly (direct main-session Edit/Write/Bash, which bypasses the spawn gate; see `.claude/rules/moai/core/agent-common-protocol.md` § Pre-Edit Sync Check) — the orchestrator SHALL auto-create one worktree per foreign registry entry (or isolate the direct-edit work into a worktree) to prevent cross-session branch-state interference. This auto-isolation procedure resolves the parallel-session branch conflict mechanically rather than surfacing it as a manual race. The "worktree entry is chosen" conjunct is no longer required: a foreign active session during direct-edit work triggers isolation too, because direct edits share the same branch-state mutable surface as spawned-agent writes.

**Conservative predicate** — ANY foreign active-session registry entry triggers auto-isolation. False positives are cheap (an extra worktree is inexpensive and user-deletable); false negatives corrupt the working tree (a genuine conflict goes unresolved and produces cross-session branch-state interference). Stale-registry false positives MAY produce a worktree that the user later deletes.

**Naming scheme** — each auto-created worktree is named `auto-<session-short>-<spec-id>` where `<session-short>` is the first 8 characters of THAT foreign session entry's UUID and `<spec-id>` is the active SPEC identifier. The naming is deterministic so the auto-created worktree is greppable and traceable to the originating foreign session. No "or equivalent" clause — the scheme is fixed.

**Landing paths** — each auto-created worktree SHALL land under `.claude/worktrees/auto-<session-short>-<spec-id>/` (L1 Claude-native) OR `~/.moai/worktrees/<project>/auto-<session-short>-<spec-id>/` (L2 persistent), so the two sessions do NOT share a branch-state mutable surface. The primary-checkout branch guard exempts worktree paths from its deny, so the auto-isolation procedure does NOT trip the branch guard.

**Surface** — the orchestrator surfaces the auto-isolation as an info log (NOT an `AskUserQuestion` round — the procedure auto-resolves the race, it does not ask the user to resolve it). The info log notes the registry entry's age so a stale-registry false positive is visible.

**Multiple foreign sessions (≥2)** — the procedure auto-creates N worktrees, one per foreign registry entry (Edge-3: each session gets its own isolated branch-state surface).

## Sentinel Key Glossary

Structured error codes emitted by `moai agent lint` and `moai workflow lint` for programmatic detection:

| Sentinel Key | Source | Meaning |
|---|---|---|
| `ORC_WORKTREE_MISSING` | LR-05 (agent lint) | Write-heavy agent lacks `isolation: worktree` in frontmatter |
| `ORC_WORKTREE_ON_READONLY` | LR-09 (agent lint) | Read-only agent (`permissionMode: plan`) has `isolation: worktree` — prohibited overhead |
| `ORC_WORKTREE_REQUIRED` | `moai workflow lint` | Legacy sentinel from the retired Agent Teams `role_profiles` isolation check — inert since the `team:` config block was removed from workflow.yaml |

### When to Use Which

### Use `claude --worktree` (`-w`) for:

- **User-initiated isolation**: Starting a fresh session for exploratory work
- **Parallel sessions**: Running multiple independent Claude sessions on same repo
- **Quick experiments**: Testing code changes without affecting main workspace

### Use `Agent(isolation: "worktree")` for:

- **Parallel team agents**: Multiple implementation teammates working simultaneously
- **File conflict prevention**: Agents that write to different file patterns
- **One-shot sub-agents**: Sub-agents making cross-file modifications
- **GitHub issue fixing**: Each issue gets isolated worktree for branch safety

### Use MoAI Worktree (`moai worktree`) for:

- **SPEC implementation**: Multi-session development of a feature
- **PR development**: Complete feature branches with commits
- **Persistent workspaces**: Work that spans multiple Claude sessions

## Integration Pattern (Hybrid Approach)

The recommended workflow combines both worktree systems:

```
PLAN PHASE
  Claude Native (-w): Quick exploration, ephemeral, no persistence
  Team researchers: No worktree (read-only, permissionMode: plan)

RUN PHASE
  MoAI Worktree: SPEC implementation, persistent state
  Team write agents: Agent(isolation: "worktree") for parallel execution
  Team read agents: No worktree (quality validation, analysis)

SYNC PHASE
  MoAI Worktree: PR creation from persistent workspace
```

## Agent Configuration by Role

### Implementation Agents (isolation: worktree + background: true)

```yaml
# Implementation teammates (write-capable implementation roles: implementer / tester / designer)
# Spawned via: Agent(subagent_type: "general-purpose", mode: "acceptEdits", isolation: "worktree")
isolation: worktree   # Isolated worktree per agent
background: true      # Non-blocking parallel execution
permissionMode: acceptEdits
```

### Research/Analysis Agents (no isolation needed)

```yaml
# Read-only teammates (read-only research/review roles: researcher / analyst / reviewer)
# Spawned via: Agent(subagent_type: "general-purpose") with a read-only tools list
# No isolation: worktree (read-only via tool restriction — the spawn-time mode
# parameter is deprecated/ignored since v2.1.213; a parent bypassPermissions/
# acceptEdits mode takes precedence over child permission settings)
permissionMode: plan  # advisory frontmatter; the tool restriction is the guarantee
```

## WorktreeCreate and WorktreeRemove Hooks (Not Registered by Default)

Claude Code v2.1.49+ defines `WorktreeCreate` / `WorktreeRemove` hooks that **replace** Claude Code's default git worktree behavior — not extend it. Per the official contract (https://code.claude.com/docs/en/hooks):

| Hook | Role | stdout contract | Failure mode |
|---|---|---|---|
| WorktreeCreate | Active creator — MUST actually create the worktree directory and echo its absolute path to stdout (plain text only, no JSON; HTTP hooks use `{"hookSpecificOutput": {"worktreePath": "..."}}`). | Single line: `/absolute/path/to/worktree` | Empty stdout OR any non-zero exit aborts creation |
| WorktreeRemove | Observer — runs during/after removal for cleanup. | No output required | Failures logged in debug mode only |

The stdin JSON for both events includes `worktree_path` (Claude Code's proposed path), `name`, `cwd`, `session_id`, `transcript_path`, `hook_event_name`.

**MoAI-ADK does NOT register these hooks by default.** Claude Code's default git worktree handling is sufficient for our agent isolation use case — write-heavy work is declared `isolation: worktree` by the retained `manager-develop` agent, by per-spawn `Agent(general-purpose)` specialists with a write-heavy domain whitelist, and by team-mode role profiles (implementer, tester, designer) per the Worktree Selection Rules above. Registering observer-only hooks here would replace the default behavior with non-functional stubs and produce `"WorktreeCreate hook returned a path that is not a directory: {}"` because an empty JSON object cannot be parsed as a path.

If a future use case requires custom worktree creation (e.g., non-git VCS, shared-file symlinks, per-worktree database setup), implement an active creator hook that:

1. Reads stdin JSON (fields: `worktree_path`, `name`, `cwd`, `session_id`).
2. Performs `git worktree add` (or equivalent for the VCS), redirecting its stdout to `/dev/null` so it does not pollute the hook stdout.
3. Prints **only** the absolute worktree path to stdout. All progress/diagnostic output goes to stderr.
4. Exits 0 on success; any non-zero exit aborts creation.

Handler files at `internal/hook/worktree_{create,remove}.go` and `internal/cli/hook.go` `worktree-create` / `worktree-remove` subcommands are preserved as opt-in infrastructure for future active-creator implementations. They are not registered in `.claude/settings.json` until such an implementation lands. Likewise, `.claude/hooks/moai/handle-worktree-{create,remove}.sh` wrapper scripts exist but are not invoked by any settings.json entry.

## Prompt Path Rules for Worktree-Isolated Agents

When the orchestrator generates prompts for agents spawned with `isolation: "worktree"`, paths in the prompt determine where the agent operates. Incorrect paths bypass worktree isolation entirely.

### HARD Rules

- [ZONE:Frozen] [HARD] Do NOT include absolute paths to the main project directory in agent prompts for write-target files
- [ZONE:Frozen] [HARD] Do NOT include `cd /absolute/project/path &&` in Bash commands within agent prompts
- [ZONE:Frozen] [HARD] Reference write-target files by project-root-relative paths (e.g., `src/domains/auth/handler.go`) and let the agent resolve from its own CWD
- [ZONE:Frozen] [HARD] `$CLAUDE_PROJECT_DIR` in hook commands is acceptable — Claude Code resolves this to the correct directory for the agent's context

### Path Categories

| Category | Example | Absolute Path OK? | Reason |
|----------|---------|-------------------|--------|
| Write-target files | Source code, tests | NO — use relative | Agent CWD is worktree root; relative paths resolve correctly |
| Read-only references | Skills, configs via `.claude/skills/<name>/<file>` | NO — use project-root-relative | Worktree root is the project root, so the root-relative form resolves there; content is identical in the main repo |
| SPEC documents | `.moai/specs/SPEC-XXX/spec.md` | Relative preferred | SPEC files are copied to worktree during checkout |
| Bash commands | `go test ./...` | NO `cd` prefix | Agent CWD is already set to worktree root |

### How It Works

When `isolation: "worktree"` is set, Claude Code:
1. Creates a temporary worktree from the current branch
2. Sets the agent's CWD to the worktree root
3. The agent constructs absolute paths from its own CWD

```
Main repo:  $HOME/project/src/auth/handler.go
Worktree:   $HOME/project/.claude/worktrees/abc123/src/auth/handler.go
```

Both share the same project structure. `src/auth/handler.go` resolves correctly in either context.

### Anti-Pattern Examples

```
# WRONG: Absolute path in prompt bypasses worktree
"Read $HOME/project/src/auth/handler.go and fix the bug"

# WRONG: cd to main project in Bash command
"Run: cd $HOME/project && go test ./..."

# CORRECT: Relative path — agent resolves from its own CWD
"The bug is in src/auth/handler.go. Read the file and fix it."

# CORRECT: No cd prefix — agent CWD is already worktree root
"Run: go test ./..."
```

## Teammate Session Launch (`--spawn`)

`moai cc -w <name> --spawn` (likewise `moai glm`) opens a session in the named worktree in a NEW tmux window and returns, so the caller keeps its own session. Without `--spawn` the same command enters the worktree in place by replacing the current process. See `.claude/skills/moai-workflow-worktree/SKILL.md` § `--spawn` for requirements, error messages, and example invocations.

> **Keep policy and display fields separate.** `llm.team_mode` and `llm.gateway.teammate_mode` / `teammate_provider` in `llm.yaml` describe launcher and teammate-role policy. Claude Code `teammateMode` controls display. Legacy `team_mode: cg` cannot select a launcher; run `moai migrate cg` for explicit migration. A tmux display setting alone does not verify mixed-provider teammate routing.

### HARD Rules

[ZONE:Frozen] [HARD] CLI launch decisions MUST NOT invoke `AskUserQuestion`. Every launch outcome is decided from observable state (tmux session presence, `teammateMode`, GLM env vars) and reported through exit codes and stderr. This satisfies the Branch Origin Decision Protocol (see `.claude/rules/moai/development/branch-origin-protocol.md` § HARD Rules).

Static guard: `internal/cli/worktree/new_test.go` `TestNew_NoAskUserQuestion` scans the worktree-creation source for `AskUserQuestion` / `mcp__askuser` references.

[ZONE:Evolvable] [HARD] `--spawn` refuses rather than degrades. Outside tmux, or without the `tmux` / `moai` binaries, it returns a non-zero exit instead of falling back to an in-place launch — a silent fallback would replace the caller's session, the outcome the flag exists to avoid. Refusal happens before any settings mutation.

### Retired: `moai worktree new --team` and the swarm registry

The `--team` flag and its four launch patterns are retired. Entering a worktree is `-w`; spawning a teammate window is `--spawn`. The write-only `.moai/state/swarm/<SPEC-ID>.json` registry was retired with it — no code ever read it, and the `moai swarm status / done / kill-all` commands it was a baseline for were never built.

### Cross-references

- `.claude/skills/moai-workflow-worktree/SKILL.md` § `--spawn` (requirements + examples)
- `internal/cli/spawn.go`, `internal/cli/spawn_test.go`
- Branch Origin Decision Protocol (BODP)

## Minimum Version Requirements

| Feature | Minimum Version | Notes |
|---------|----------------|-------|
| `isolation: worktree` in Agent frontmatter | 2.1.49 | Basic worktree isolation |
| `background: true` in Agent frontmatter | 2.1.46 | Non-blocking agent execution |
| `claude --worktree` user flag | 2.1.50 | User-initiated worktree sessions |
| `Ctrl+X Ctrl+K` to kill background agent | 2.1.83 | Kill stuck background agents |
| Worktree CWD isolation fix | **2.1.97** | Prior versions leaked agent CWD back to parent session |
| Stop/SubagentStop hook stability | **2.1.97** | Prior versions failed on long-running sessions |
| `moai doctor` MCP scope duplicate detection | **2.1.110** | Warns on MCP server duplication across `.mcp.json` + settings.json |
| Bash tool timeout ceiling enforcement | **2.1.110** | Maximum 600,000ms (10 min) enforced by runtime |
| `effortLevel` setting for Opus 4.7 | **2.1.110** | Supports `low`/`medium`/`high`/`xhigh`/`max` effort levels |
| `CLAUDE_ENV_FILE` on Windows | **2.1.111** | Prior versions: no-op on Windows; fixed to inject env as on macOS/Linux |
| `disableBypassPermissionsMode` policy | **2.1.111** | Prevents agents from requesting `bypassPermissions` when `true` |

**Recommended**: Claude Code **2.1.186 or later** for current background-agent permission-prompt semantics, Opus 4.7+ / 4.8 / Opus 5 support, MCP doctor warnings, and Windows CLAUDE_ENV_FILE parity. Minimum baseline: **2.1.97** for worktree isolation.

## Troubleshooting

| Issue | Cause | Solution |
|-------|-------|----------|
| Worktree not found | Removed manually | Run `git worktree list` to verify |
| Agent worktree conflicts | Multiple agents same file | Check file ownership in team config |
| Stale worktree branches | Incomplete cleanup | Run `git worktree prune` |
| Hooks not firing | Missing wrapper script | Check `.claude/hooks/moai/` directory |
| `--tmux` not working | Unsupported terminal | Use tmux or iTerm2 (not VS Code, Ghostty) |

## Refused Commands in a Worktree-Isolated Session

### Which guard refused — read the message, it is the only discriminator

Two different guards refuse commands in the same worktree-isolated session, and nothing but the wording of the refusal tells them apart. Read the message first; everything else follows from it.

| The refusal says | Whose guard | Can it be changed here? |
|---|---|---|
| `Dangerous command blocked:` … | **This project's own** pre-tool hook, in `internal/hook/pre_tool.go` | **Yes** — it is this repository's code |
| … `too complex to verify that it stays inside the worktree` | **The Claude Code binary** | **No** — there is no source here that implements or configures it |

A reader who confuses the two searches the wrong place. Hunting through this repository's source for the worktree guard finds only prose quoting the message and never the guard; treating the dangerous-command refusal as an untouchable upstream behaviour leaves a local hook unexamined when it is in fact the thing refusing.

A third, separate guard also lives here: `internal/hook/branch_guard.go` defends branch-state mutation in the primary checkout. It neither implements nor configures the worktree guard, and editing it has no effect on a `too complex to verify` refusal.

### What has been observed to trip the worktree guard

The full refusal reads:

> This session is isolated in the worktree `<worktree-path>`, but this command is too complex to verify that it stays inside the worktree.

**This table is a record of what has been seen, not a specification.** It is explicitly non-exhaustive, and a handful of observations do not establish a general rule — the trigger condition was never narrowed, so do not infer from it that "complex commands are refused", and do not infer a tokenization mechanism from the shape of these cases.

| Observed trigger | Provenance | Notes |
|---|---|---|
| A quoted-delimiter heredoc (`<<'EOF'`) whose **body contains braces wrapping quoted key/value pairs** — a JSON line, for example | **First-hand** — measured by paired probes in a worktree-isolated session | Body size, pipes, backticks, and command substitution are **not** the trigger: a body of roughly 6.7 KB of prose was accepted, a ten-character JSON line in the same position was refused, and a command substitution in the body was folded correctly and passed |
| **Several mutation steps bundled into one compound command** | **Second-hand** — reported by another working lane, not measured here | Recorded because it carried the same refusal sentence; it has not been reproduced by the session that wrote this section |
| A heredoc whose **body names a git subcommand** — prose such as `run git merge --no-ff <sha>` or `git checkout -b <branch>` fed to a non-git command | **First-hand** — measured by paired probes in a worktree-isolated session at 2.1.251 | Refused with the variant sentence `… this command names git in a form too complex to verify …`. The same heredoc without git text passed, and the same git text passed when carried as a double-quoted argument or read from a file (`--stdin < <file>`). **This row was contradicted at 2.1.275** — see § Two observations disagree on the heredoc-carrying-git-text shape below; do not cite it as current behaviour |

**Gap**: the boundary between an accepted and a refused command is unmeasured in every row. Anything outside these observations is unknown, not permitted.

### Why the two heredoc delimiter forms differ

The distinction the observation above turns on is the delimiter quoting, and it
is bash semantics, not guard behavior:

- **A quoted delimiter (`<<'EOF'`) makes the body inert text.** Bash performs no
  expansion of any kind inside such a body — no parameter expansion, no command
  substitution, no arithmetic expansion, and no brace expansion. A brace there
  is a literal character and cannot be brace expansion. This is the same fact a
  guard relies on when it folds a command substitution appearing in such a
  body, as observed above.
- **An unquoted delimiter (`<<EOF`) makes the body live text.** Parameter
  expansion, command substitution, arithmetic expansion, and brace expansion
  all apply. No guard — and no reader of this file — may treat an
  unquoted-delimiter body as inert, and a refusal that errs on the side of
  caution there is correct behavior, not a defect.

The observed asymmetry is therefore narrow: in a position provably free of
expansion (the quoted-delimiter body), braces alone are treated as live, while
command substitutions in the same position are already folded.

The refusal half of this asymmetry was re-measured by paired probes in a
worktree-isolated session on this repository: the brace form was refused with
the `too complex to verify` sentence above before the command executed (the
target file was confirmed absent afterward), and the paired probe — the same
shape with a command substitution in the body — executed and passed with the
body preserved as a literal. This remains a record of observations, not a
specification; the boundary of the guard's analyzer is still unmeasured.

### The refusal's message shapes

The refusal is not one sentence. Seven distinct clauses have been seen — two of them first met while writing this section — and the difference matters because a reader who greps for the wording they remember concludes the guard did not fire.

Four shapes are pinned in this repository as quoted fixtures (`internal/hook/worktree_guard_refusal_test.go`); three more have been observed since and are not pinned anywhere:

| Shape | The clause after `…but this command` | Provenance | Pinned |
|---|---|---|---|
| Cross-tree redirect via `-C` | `redirects git to the shared checkout via -C` | First-hand, main session (card t529); re-measured first-hand at 2.1.275 (card t852) — byte-identical but for the path | `sampleGuardRefusalDashC` |
| Cross-tree redirect via `--git-dir` | `redirects git to the shared checkout via --git-dir` | First-hand, **background subagent** (card t529) | `sampleGuardRefusalGitDir` |
| Unverifiable command | `is too complex to verify that it stays inside the worktree` | First-hand, main session (card t529); re-measured first-hand at 2.1.275 (card t852) | `sampleGuardRefusalComplex` |
| Working-directory resolution | `'s working directory resolved to the shared checkout (<path>)` | **Card-quoted, never measured.** The quote is truncated and the continuation is unobserved — the fixture reproduces that truncation deliberately | `sampleGuardRefusalCwdCardQuoted` |
| Runtime-computed target | `points git at a directory computed at runtime (-C <path>)` | Second-hand — another lane, 2.1.275 (card t880); not reproduced here | — |
| Unverifiable git form | `names git in a form too complex to verify` | First-hand at 2.1.251 (recorded in the trigger table above); second-hand at 2.1.275 (card t880) | — |
| Unverifiable non-git command | `runs <command> with <argument> in a plain command, so what it runs cannot be shown not to be git` | First-hand at 2.1.275 (card t852) — see the reproduction note below | — |

**A seventh shape exists and resisted narrowing.** The row above was met while writing this very section: a compound command assigning a shell variable and then running `printf` with a long multi-line argument was refused, with the guard naming `printf` and quoting its whole argument. Four paired probes in the same session failed to reproduce it — `printf 'gitignore'` alone passed, an argument carrying backticks passed, the two combined passed, and the same `printf` redirected into a runtime-computed `"$VAR/path"` passed. So neither the `git` substring, nor backticks, nor a computed redirect target is the trigger on its own.

This is worth more than the row itself: it is the clearest available demonstration that **the guard's analyzer refuses on a property none of the observations here has isolated**, and that a refusal can name a command that has nothing to do with git. Treat an unfamiliar refusal clause as a seventh, eighth, or ninth shape rather than as a misfire, and record its verbatim wording — the catalogue above grew twice while this section was being written.

**The classifier does not key on any of these clauses.** `internal/hook/post_tool_failure.go` matches the anchor `isolated in the worktree` alone, which is why the three unpinned shapes still classify as `WorktreeGuardRefusal` rather than falling into the catch-all. The anchor is an observed dependency on upstream wording, not a contract — if the runtime rewrites that opening clause, detection goes silently to zero.

### The axis is git — with one measured exception

Card t741 established that a cross-tree **non-git** argument is not refused: `ls -la <another worktree>/go.mod` passed from a session anchored elsewhere, while `git -C <another worktree> log` in the same session was refused and `git -C <own worktree> log` passed. The discriminating axis there is git, not the crossing of trees.

Card t880 measured the same axis from the other side at 2.1.275 and reported the refusal target as **a git call that cannot be statically bound to the worktree** — not nested shell expansion as such. Backticks, `$( )`, `for` loops, and heredocs were all reported to pass when no git call was involved.

**That generalization does not hold universally, and the counter-example was measured here.** At 2.1.275, in a worktree-isolated session with the guard demonstrably live (a `git -C` control refused in the same session), a quoted-delimiter heredoc whose body was the ten characters of a JSON object — no git anywhere in the command — was refused with the `too complex to verify` clause. So "git-free commands pass" is false as stated; what survives is the narrower claim that a statically unbindable git call is *sufficient* for refusal, not that it is *necessary*.

**Gap**: no measurement here narrows what else is sufficient. The brace observation in the trigger table above and this one are the same shape, and its boundary is still unmeasured.

### Two observations disagree on the heredoc-carrying-git-text shape

The trigger table above records, first-hand at **2.1.251**, that a quoted-delimiter heredoc whose body names a git subcommand was refused. Two later observations disagree with it and with each other:

| Session kind | Version | Result |
|---|---|---|
| Worktree-isolated | 2.1.251 | **Refused** — `names git in a form too complex to verify` (trigger table above) |
| Worktree-isolated | 2.1.275 | **Passed** — measured first-hand for card t852, with a `git -C` control refused in the same session, so the pass is a pass and not a dead guard |
| Primary checkout | 2.1.275 | **Refused** — reported by another lane (card t852 dispatch) |

Two explanations fit, and **neither is established**:

- **A version change.** The behaviour flipped between 2.1.251 and 2.1.275 for this shape. The same session that saw the flip also re-measured the JSON-brace shape as still refused at 2.1.275, so any such change was selective rather than a general relaxation.
- **A layer difference.** The worktree-isolated session meets the Claude Code runtime guard, while the primary checkout meets this repository's own branch guard (`internal/hook/branch_guard.go`) — two different refusers with overlapping-sounding wording, per the discriminator table at the top of this section.

**To settle it**, run the identical heredoc in both session kinds at one pinned version and compare the refusal wording as well as the outcome; the wording is what separates the two guards. Until that is done, do not cite either observation as the behaviour.

### A background subagent carries no anchor of its own

A background subagent is **not** pinned to the tree it was spawned in. Its working directory is re-resolved against the session's current anchor at each call, including while the session is moving between trees.

Card t741 measured this directly: twelve path-free calls from one background subagent, no `cd` anywhere in them, and the working directory changed twice across the run as the parent session moved — twelve passes, zero refusals.

Two consequences, and the second is the dangerous one:

- **A refused subagent command is not silent.** The `--git-dir` fixture above was captured from a background subagent, so a refusal reaches `PostToolUseFailure` from a subagent exactly as it does from a main session.
- **A re-anchored subagent is entirely silent.** A path-free command keeps succeeding while the tree underneath it changes. Nothing refuses, nothing is logged, and the subagent cannot tell that its later work landed in a different tree from its earlier work. This is the audit-degradation path that refusal records do not catch.

Operationally: do not move a session's worktree while a background auditor is live. Where that is unavoidable, the resulting verdict's Gaps section must carry the refusal record — `verification-claim-integrity.md` §3.1 (a refusal is a Gap, never a silent substitution) governs, and a lead reading the verdict is the only thing enforcing it.

**Not recommended**: per-agent anchor registration. It would require changing the Claude Code binary rather than this repository, and it inverts the guard's purpose — a subagent pinned to its spawn tree keeps writing to a tree the operator believes was disposed of.

### Workarounds — two situations, not two competing options

Which one applies is decided by what you were trying to do, so identify the situation before reaching for a form.

**Writing file content → use the `Write` tool.** This is the repository's convention. It reaches the filesystem without issuing a shell command, so the guard never parses the content and the body's shape stops mattering. Reach for it first rather than reshaping the heredoc.

Ad-hoc detours happen to work — routing the content through an interpreter's own file-write call, or splitting the file into brace-free pieces — but they are **not** the convention and are not equal options. Two paths going their own way is not a convention: a reader who meets one detour in one commit and a different one elsewhere learns nothing reusable.

**Running a compound command → split it into separate plain commands.** Issue the steps one at a time rather than chaining them. Note that this trades away one property worth keeping in mind: an environment scrub written as `unset … && <command>` is load-bearing as a single invocation, because each command runs in a fresh process, so that particular pairing is not one to split apart.

### Acceptance-criteria commands — the measured boundary and the authoring rule

A corpus-wide, block-level census of acceptance-criteria verification commands (111 files
carrying complex git forms across `**/acceptance.md`) probed each file with a read-only replica
of its block composition in a worktree-isolated session at Claude Code **2.1.278** (every
disposition traces to a recorded probe). Like every table in this section it is a record of
observations, not a specification of the parser.

Refused — the `names git in a form too complex to verify` refusal fires and nothing executes:

| Refused form | Notes |
|---|---|
| git inside `$()` whose result a later statement expands — `BASE=$(git merge-base A B)` then `git diff "$BASE"..HEAD` | refuses for `;`, newline, and `&&` separation alike; refuses with `head`, `tail`, and `sort` pipeline terminators too |
| git nested inside another git command's arguments — `git diff "$(git merge-base A B)"..HEAD` | quoted or unquoted |
| a tree-write bundle: variable assignment (`mktemp`, `mkdir`) + git + an `rm -rf` tail in one invocation | refuses even when every git verb in it is read-only — the bundle is what refuses, not the mutation |
| git piped into `while`/`for`, or `for f in $(git …)` | refuses with or without git inside the loop body |
| a `( … )` subshell compound containing git | |
| `test "$(<git $()>)"` | refuses even with a `wc -l` pipeline terminator |
| a non-git command (`printf`, `rg`) whose argument carries a git `$()`; an env-prefixed `VAR="$(git …)" go test …` | `echo` is the measured exception, below |

Measured executable: plain separately-invocable git verbs; `;` / `&&` / `||` compounds of them;
`$?`-capture lines; a `$()` assignment with no later expansion; `$()` embedded in the same
statement's `echo "…$(…)…"` (bare, `|wc`, `|head`, `|awk` terminators all pass); a `$()`
assignment expanded later when the `$()` pipeline terminates in a counter stage (`wc -l`,
`grep -c`, `awk '{print $1}'`); redirect-to-file capture (`git … > /tmp/f`); git text inside a
quoted grep pattern or inside a comment.

**Gap**: the counter-terminator exception and the echo-embedding exception are observed
boundaries, not a mechanism — the territory between the rows is unmeasured, and per the section
gap note above, unknown is not permitted.

**Authoring rule for AC verification commands** (normative — this is the repo-side fix; the
guard itself is the binary's):

1. One plain git verb per line, run from the worktree root.
2. Capture exit codes in separate lines — `git diff --quiet …; echo "exit=$?"` — so the exit
   code is the verification's own field (`verification-completeness.md` §2.1), never a
   `$()`-captured variable.
3. Never git inside `$()`. Derive ranges with the three-dot form (`git diff --name-only
   develop...HEAD`) and record the merge-base on its own line (`git merge-base develop HEAD`)
   when the base value itself is evidence.
4. No write/cleanup composition tail in a verification command. Scratch writes live under
   `/tmp`; cleanup is a separate plain step, never an `rm -rf` bundled into the same invocation.
5. Never relocate a refused command into a script file — the guard cannot read inside a script,
   so the relocation hides the risk instead of removing it. Reduce the verification instead.
6. Pin the tree SHA the measurement was taken on (`verification-completeness.md` §4).

Working example — a refused form and its executable restatement:

```bash
# REFUSED (assignment + later expansion of a git-bearing substitution):
B=$(git merge-base develop HEAD)
git diff --name-only "$B"..HEAD -- internal/pkg/ | wc -l

# EXECUTABLE (plain verbs; base recorded on its own line; three-dot range):
git merge-base develop HEAD                          # record the base value as evidence
git diff --name-only develop...HEAD -- internal/pkg/ | wc -l
git diff --quiet develop...HEAD -- internal/pkg/; echo "diff_exit=$?"
```

**Versions measured**: the trigger table and the delimiter asymmetry were measured at Claude Code **2.1.251**. The message-shape catalogue, the git-axis counter-example, and the heredoc disagreement were measured at **2.1.275** (card t852; `claude --version` read in the measuring session). The subagent-anchor observations were measured at the version current when card t741 was measured, which was not recorded there.

Guard behaviour is version-dependent — one shape has already been observed to flip between these two versions — so **state the version whenever you add a row here, and read the version before citing one.** No behaviour above is known to hold at any version other than the one its row names.

## SPEC-to-Worktree Mapping

[ZONE:Frozen] [HARD] Per-step worktree applicability is governed by `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline (canonical source). This table summarizes the mapping for quick reference; on conflict, spec-workflow.md wins.

| Step | Phase   | Worktree?                | Location                              | Lifecycle event              |
|------|---------|--------------------------|---------------------------------------|------------------------------|
| 1    | Plan    | **NO** (main checkout)   | n/a — `plan/SPEC-XXX` branch on main  | plan PR merged               |
| 2    | Run     | **opt-in (`moai cc -w <name>`)** | the entered worktree                  | run PR merged                |
| 3    | Sync    | **opt-in** — same as Step 2            | same path as Step 2 (do NOT recreate) | sync PR merged               |
| 4    | Cleanup | n/a                      | host checkout                         | `moai worktree done SPEC-XXX` |

Worktree usage is user opt-in; the default flow runs all phases on a `feat/SPEC-XXX` branch in the main checkout. To use one, enter it with `moai cc -w <name>` before invoking the phase.

[ZONE:Frozen] [HARD] Disposal contract: `moai worktree done SPEC-XXX` MUST run only after BOTH run PR AND sync PR are merged. Premature disposal between Step 2 merge and Step 3 merge breaks Sync.

---

Version: 4.5.0 (guard-refusal message-shape catalogue — four pinned fixtures plus three observed since, one of them met while writing the section and unreproduced by four narrowing probes; the git axis and its measured counter-example; the heredoc disagreement recorded unresolved across two session kinds; background subagents carry no anchor of their own; per-row version attribution)
