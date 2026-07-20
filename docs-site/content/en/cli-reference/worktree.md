---
title: moai worktree Worktrees
weight: 25
draft: false
---

`moai worktree` (alias `moai wt`) manages Git worktrees for parallel SPEC development. It provides subcommands to create, list, switch, sync, remove, and clean worktrees.

## Subcommands

| Command | Description |
|--------|------|
| `moai worktree new [branch]` | Create a new worktree |
| `moai worktree list` | List active worktrees |
| `moai worktree status` | Show worktree status |
| `moai worktree switch [branch]` | Switch to a worktree |
| `moai worktree go [branch]` | Print the worktree path for shell navigation |
| `moai worktree sync [branch]` | Sync a worktree with its base branch |
| `moai worktree done [branch]` | Finish and clean up a worktree |
| `moai worktree remove [path]` | Remove a worktree |
| `moai worktree clean` | Clean up stale worktree references |
| `moai worktree recover` | Recover the worktree registry |

## moai worktree new

```bash
moai worktree new [branch-name]
```

| Flag | Description |
|--------|------|
| `--path <dir>` | Specify the worktree path (default: `.moai/worktrees/<SPEC-ID>` for a SPEC ID, otherwise `../<branch-name>`) |
| `--base <branch>` | Base branch (default: `origin/main`, auto-fetched). `--base main` is for local-only commits |
| `--from-current` | Use the current HEAD as the worktree base (skips `git fetch origin main`) |
| `--tmux` | Create a tmux session after creating the worktree |
| `--team` | Spawn a Claude/GLM session in the new worktree (tmux+CG → `moai glm` window, tmux+CC → `moai cc` window, no-tmux → in-process, no-flag → handoff guidance) |

## moai worktree done

```bash
moai worktree done [branch-name]
```

| Flag | Description |
|--------|------|
| `--force` | Force removal even with uncommitted changes |
| `--delete-branch` | Delete the branch after removing the worktree |
| `--auto` | Auto mode — silent execution for automation (e.g. after a PR merge) |

## Examples

```bash
# Create a worktree for a SPEC (origin/main base)
moai worktree new SPEC-AUTH-001

# Local worktree based on the current HEAD
moai worktree new feature-x --from-current

# List active worktrees
moai worktree list

# Move to a worktree from the shell
cd "$(moai worktree go SPEC-AUTH-001)"

# Clean up + delete branch when done
moai worktree done SPEC-AUTH-001 --delete-branch
```

## Related documents

- [Worktree Workflow](/en/advanced/autonomous-loops) — parallel development patterns
- [CG Mode](/en/multi-llm/cg-mode) — `--team` hybrid execution
- [CLI Overview](/en/getting-started/cli)
