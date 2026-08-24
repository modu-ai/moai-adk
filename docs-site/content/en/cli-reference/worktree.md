---
title: moai worktree Worktrees
weight: 25
draft: false
---

`moai worktree` (alias `moai wt`) manages the Git worktrees used for parallel SPEC development. It offers eight subcommands: syncing, finishing, removing, cleaning, registry recovery, and the state guards that wrap isolated agent runs.

## Entering and listing worktrees is not this command's job

`moai worktree` **manages** worktrees — it does not take you into one and it does not list them.

| What you want to do | Command to use |
|-----------|-------------|
| Start working inside a worktree | `moai cc -w <name>` (or `moai glm -w` / `moai cg -w`) |
| Open one in a new tmux window while keeping the current session | `moai cc -w <name> --spawn` |
| List worktrees | `git worktree list` |
| Create a new worktree | `moai cc -w <name>` (creates `.claude/worktrees/<name>/` automatically) or `git worktree add` |

A short name passed to `-w` resolves under `.claude/worktrees/<name>/`, and is created if it does not exist yet. An absolute path re-enters an existing worktree under `~/.moai/worktrees/` or `<project>/.claude/worktrees/`. Any other absolute path is rejected.

## Subcommands

| Command | Description |
|--------|------|
| `moai worktree sync [branch-name]` | Bring base-branch changes into the worktree |
| `moai worktree done <branch-name>` | Remove the worktree attached to a branch, optionally deleting the branch too |
| `moai worktree remove <path>` | Remove the worktree at the given path |
| `moai worktree clean` | Prune stale references, clean up merged or abandoned worktrees |
| `moai worktree recover` | Repair the worktree registry |
| `moai worktree snapshot` | Capture the working-tree state as a snapshot |
| `moai worktree verify` | Compare the current working tree against a snapshot |
| `moai worktree restore` | Roll the working tree back to the snapshot HEAD state |

## moai worktree sync

```bash
moai worktree sync [branch-name]
```

Given a branch name, syncs that branch's worktree; omit it and the worktree of the current directory is synced.

| Flag | Description |
|--------|------|
| `--base <branch>` | Base branch (default: `main`) |
| `--strategy <mode>` | `merge` (default) or `rebase` |

## moai worktree done

```bash
moai worktree done <branch-name>
```

The branch name is required. It finds the worktree using that branch, removes it, and deletes the branch if you ask for it. **It does not merge** — finish the base-branch merge separately with `git merge` or a PR.

| Flag | Description |
|--------|------|
| `--force` | Force removal even with uncommitted changes |
| `--delete-branch` | Delete the branch after removing the worktree |
| `--auto` | Quiet mode for automation (e.g. cleanup after a PR merge). It does not fail when no worktree is found |

## moai worktree remove

```bash
moai worktree remove <path>
```

The argument is a **file-system path**, not a branch name.

| Flag | Description |
|--------|------|
| `--force` | Force removal even with uncommitted changes |

## moai worktree clean

```bash
moai worktree clean [--merged-only | --stale] [--yes] [--json] [--base <branch>]
```

Run without flags, it only prunes stale worktree references.

| Flag | Description |
|--------|------|
| `--merged-only` | Remove only worktrees whose branch is merged into base |
| `--stale` | Sweep up abandoned worktrees with nothing to lose (preview by default) |
| `--yes` | Actually remove instead of previewing with `--stale` |
| `--json` | With `--stale`: report every non-protected worktree with its keep reason, dirty, merge, and anchor state as JSON. Removes nothing, and overrides `--yes` |
| `--base <branch>` | Base branch used to judge `--merged-only` and `--stale` (default: `origin/main`) |

`--stale` and `--merged-only` cannot be combined.

### --stale safety rules

A worktree becomes a removal candidate only when it satisfies **both** of these conditions.

1. The working tree is clean — no uncommitted changes and no untracked files
2. The branch carries no commits of its own beyond base

If either fails, the worktree is kept and the reason for keeping it is printed alongside. **Branches are never deleted**, so even when the worktree directory disappears the commits remain reachable by branch name. The main checkout and the worktree the command is running in are always protected.

`--stale` previews by default. Add `--yes` to actually delete.

## moai worktree recover

```bash
moai worktree recover
```

Repairs the worktree administrative files with `git worktree repair`, prunes stale references, and finally prints the list of recognized worktrees. It takes no flags.

## moai worktree snapshot

```bash
moai worktree snapshot
```

Captures HEAD, the branch, the porcelain status, and untracked files under `.moai/specs/`, recording them as JSON in `.moai/state/`. Its purpose is to take a reading right before invoking an isolated agent.

| Flag | Description |
|--------|------|
| `--out <path>` | Snapshot output path (default: `.moai/state/worktree-snapshot-<id>.json`) |
| `--agent-name <name>` | Record the agent name (referenced later during verify) |

## moai worktree verify

```bash
moai worktree verify --snapshot <path>
```

Compares the current working tree against a snapshot. `--snapshot` is **required**.

| Flag | Description |
|--------|------|
| `--snapshot <path>` | Path to the pre-run snapshot JSON (required) |
| `--agent-response <path>` | Agent response JSON — used to detect an empty `worktreePath` |
| `--agent-name <name>` | Agent name to record in the divergence and suspect logs |

| Exit code | Meaning |
|-----------|------|
| `0` | clean |
| `1` | divergence detected |
| `2` | suspect (empty `worktreePath`) |
| `3` | both |

## moai worktree restore

```bash
moai worktree restore --snapshot <path>
```

Runs `git restore --source=<snapshot HEAD> --staged --worktree :/` to roll tracked files back to the snapshot HEAD state. **Untracked files cannot be brought back by git**, so only their paths are listed and you have to recreate them yourself.

| Flag | Description |
|--------|------|
| `--snapshot <path>` | Path to the snapshot JSON (required) |
| `--dry-run` | Print the git commands that would run, without running them |

## Examples

```bash
# Create a worktree and enter it right away (.claude/worktrees/feat-auth/)
moai cc -w feat-auth

# Spawn a GLM teammate in a new tmux window while keeping the current session
moai cg -w feat-auth --spawn

# List worktrees
git worktree list

# Sync the current worktree with main (merge)
moai worktree sync

# Sync a specific worktree using rebase
moai worktree sync feature/SPEC-AUTH-001 --strategy rebase

# Preview abandoned worktrees first, then remove them once confirmed
moai worktree clean --stale
moai worktree clean --stale --yes

# Clean up the worktree after merging, and delete the branch
moai worktree done feature/SPEC-AUTH-001 --delete-branch
```

## Related documents

- [Git Worktree Overview](/en/worktree/) — concepts and workflows
- [Complete Guide](/en/worktree/guide) — detailed usage per command
- [CG Mode](/en/multi-llm/cg-mode) — Claude leader + GLM teammate hybrid
- [CLI Overview](/en/getting-started/cli)
