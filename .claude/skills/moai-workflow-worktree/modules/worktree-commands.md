# Worktree Commands Module

Purpose: Accurate reference for the `moai worktree` command group (alias `wt`) — its eight subcommands and their flags — plus where the tasks it does not cover are handled instead.

Version: 3.0.0

---

## Quick Reference (30 seconds)

`moai worktree` manages worktrees that already exist. It does not create or enter one.

| Task | Command |
|---|---|
| Create and work inside a worktree | `moai cc -w <name>` (launcher) |
| Open a worktree in a new tmux window, keep this session | `moai cc -w <name> --spawn` |
| List worktrees | `git worktree list` |
| Sync a worktree with its base branch | `moai worktree sync [branch-name]` |
| Remove a worktree | `moai worktree remove [path]` |
| Clean stale or merged worktrees | `moai worktree clean` |
| Repair the worktree registry | `moai worktree recover` |
| Finish a worktree after its branch merged | `moai worktree done [branch-name]` |
| Snapshot / verify / restore working-tree state | `moai worktree snapshot` · `verify` · `restore` |

Run `moai worktree <subcommand> --help` for the authoritative flag list of the installed binary.

---

## Management Commands

### moai worktree sync [branch-name]

Sync a worktree with its base branch. With a branch name, syncs the worktree on that branch; with no argument, syncs the worktree at the current directory.

Flags:
- `--base <branch>`: Base branch to sync from (default `main`)
- `--strategy <merge|rebase>`: Sync strategy (default `merge`)

### moai worktree remove [path]

Remove the worktree at the given path. The path is required. Refuses while a live session is anchored in the worktree unless `--force` is given.

Flags:
- `--force`: Remove even with uncommitted changes

### moai worktree clean

Clean stale worktree references.

Flags:
- `--merged-only`: Only remove worktrees whose branches are merged into the base
- `--stale`: Remove abandoned worktrees that are clean and hold no unique commits (preview unless `--yes`)
- `--yes`: Actually perform the `--stale` removals instead of previewing them
- `--json`: Report every non-protected worktree and its state as JSON; removes nothing
- `--base <branch>`: Base branch for the `--merged-only` and `--stale` checks (default `origin/main`)

### moai worktree recover

Repair the worktree registry. No flags.

### moai worktree done [branch-name]

Complete a worktree and clean up after its work has merged. The branch name is required and names the worktree's branch, not its directory. Merging into the base is done separately (git merge or a PR).

Flags:
- `--force`: Remove even with uncommitted changes
- `--delete-branch`: Delete the branch after removing the worktree
- `--auto`: No success output, for automation (for example after a PR merge); failures still exit non-zero

---

## State Guard Commands

### moai worktree snapshot

Capture a working-tree state snapshot for guard verification.

Flags:
- `--out <path>`: Output snapshot path (default `.moai/state/worktree-snapshot-<id>.json`)
- `--agent-name <name>`: Agent name, recorded for a later `verify`

### moai worktree verify

Verify the working-tree state against a snapshot and check an agent response.

Flags:
- `--snapshot <path>`: Pre-state snapshot JSON (required)
- `--agent-response <path>`: Agent response JSON (optional, for suspect detection)
- `--agent-name <name>`: Agent name to record in divergence and suspect logs

### moai worktree restore

Restore the working tree to a snapshot's HEAD state.

Flags:
- `--snapshot <path>`: Snapshot JSON (required)
- `--dry-run`: Print the git command without executing it

---

## Not Provided by This Command

Earlier documentation described a separate `moai-worktree` program with more subcommands and flags. None of them exist; use the replacements below.

| Not available | Use instead |
|---|---|
| `new`, `switch`, `go` (create, enter, print a path) | `moai cc -w <name>` to create and enter; `EnterWorktree(<path>)` to re-enter from a session |
| `list`, `status` | `git worktree list`; `git -C <path> status` |
| `config` | Worktree behavior is configured in the project's `.moai/config/` sections |
| sync `--include`, `--exclude`, `--auto-resolve`, `--interactive`, `--all` | Resolve conflicts with ordinary git in the worktree; run `sync` once per worktree |
| remove `--keep-branch`, `--backup`, `--dry-run` | `remove` keeps the branch; commit or push before removing |
| clean `--days`, `--interactive` | `clean --stale` previews by default; add `--yes` to act |
| creation `--template`, `--shallow`, `--depth` | Prepare the worktree after entering it |

---

Version: 3.0.0
Module: Command reference for `moai worktree`, matched to the CLI source
