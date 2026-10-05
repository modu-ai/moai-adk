# MoAI Worktree Examples

Purpose: Usage examples for worktree-based development with MoAI-ADK, written against the commands that exist: the `moai cc -w` launcher for creating and entering a worktree, `moai worktree` for managing one, and plain git for inspection.

Version: 2.0.0

---

## Command Integration Examples

### Example 1: SPEC Development in a Worktree

Scenario: Plan, implement, and close one SPEC in an isolated worktree.

```bash
# Create the worktree and start a session inside it
moai cc -w user-auth

# Inside that session
/moai plan "User Authentication System"
/moai run SPEC-AUTH-001

# Bring the worktree up to date with its base (run inside the worktree:
# with no argument, sync acts on the worktree at the current directory)
moai worktree sync --base main
/moai sync SPEC-AUTH-001

# After the branch has merged, remove the worktree
moai worktree remove .claude/worktrees/user-auth
```

### Example 2: Parallel SPEC Development

Scenario: Work on several SPECs at once, one worktree and one session each.

```bash
# One worktree per SPEC, each opened in its own tmux window
moai cc -w auth --spawn
moai cc -w payment --spawn
moai cc -w dashboard --spawn

# Inspect every worktree from any shell
git worktree list

# Remove worktrees whose branches have merged into the base
moai worktree clean --merged-only
```

To return to a worktree from a running session, use `EnterWorktree(<path>)`; from a new terminal, run `moai cc -w <name>` again — an existing worktree of that name is reused, not recreated.

### Example 3: Checking a Worktree During Development

```bash
# Branch, ahead/behind, and changed files for one worktree
git -C .claude/worktrees/auth status
git -C .claude/worktrees/auth log --oneline -5

# Rebase a worktree onto the base instead of merging, naming it by its branch
moai worktree sync <branch-name> --strategy rebase --base main
```

---

## Workflow Integration Examples

### Example 4: End-to-End Script

```bash
#!/bin/bash
# spec_development_workflow.sh — one SPEC, start to finish
NAME="$1"

# Create and enter the worktree; run /moai plan, /moai run, /moai sync inside it
moai cc -w "$NAME"

# After the session ends and the branch has merged, close it by branch name:
moai worktree done <branch-name> --delete-branch
```

`moai worktree done` takes the worktree's **branch** name, not its directory name. For a worktree entered by short name under `.claude/worktrees/`, follow the project's worktree rules for disposal (the session-end keep/remove prompt, or `git worktree remove <path>` once its branch has merged).

### Example 5: Stale-Worktree Cleanup

```bash
# Preview abandoned worktrees (clean, no unique commits)
moai worktree clean --stale

# Remove them
moai worktree clean --stale --yes

# Machine-readable report of every non-protected worktree; removes nothing
moai worktree clean --json
```

---

## Troubleshooting Examples

### Example 6: Common Issues and Solutions

```bash
# Problem 1: Worktree creation fails
df -h .
git status
git remote -v

# Problem 2: Sync conflicts
git -C .claude/worktrees/auth status          # see the conflicted files
# resolve them with ordinary git inside the worktree, then:
moai worktree sync --base main

# Problem 3: Registry out of step with the worktrees on disk
moai worktree recover
git worktree prune

# Problem 4: Permission issues
ls -la .claude/worktrees/
```

---

Version: 2.0.0
Examples: Usage patterns for worktree development with MoAI-ADK
