# Troubleshooting Module

Purpose: Comprehensive troubleshooting guide for Git worktree management issues, common errors, and resolution strategies.

Version: 1.0.0

---

## Quick Reference (30 seconds)

Common Issue Categories:
- Creation Failures: Worktree creation errors and directory conflicts
- Registry Issues: Registry corruption, sync failures, orphaned entries
- Git State Problems: Detached HEAD, branch conflicts, merge issues
- Path and Permission: Directory access, path resolution, file permission errors
- Integration Issues: MoAI workflow integration and command coordination

Quick Diagnostics:
- List Git worktrees: git worktree list
- Report every worktree and its state: moai worktree clean --json (removes nothing)
- Repair the worktree registry: moai worktree recover
- Check branch state: git status

---

## Creation Failures

### Worktree Already Exists Error

Symptoms:
- Error message indicating worktree or directory already exists
- Creation command fails with exit code 128
- Partial worktree state from previous failed creation

Root Causes:
- Previous worktree creation was interrupted
- Directory exists but registry entry is missing
- Git worktree metadata is corrupted

Resolution Steps:
1. Check if the worktree directory exists at the expected path
2. If directory exists, check if it contains valid Git worktree metadata by looking for .git file
3. If registry entry is missing, manually add the worktree to registry or remove the directory
4. If Git metadata is corrupted, remove the directory and prune Git worktree metadata using git worktree prune
5. Retry the worktree creation command

Prevention:
- Always use moai worktree remove (or git worktree remove) instead of manual directory deletion
- Run git worktree prune after any failed Git operations
- Run moai worktree recover when the registry and the directories on disk disagree

### Branch Already Checked Out Error

Symptoms:
- Error stating branch is already checked out in another worktree
- Cannot create worktree with desired branch name
- Git refuses to create worktree

Root Causes:
- Attempting to use same branch in multiple worktrees
- Previous worktree using this branch was not properly cleaned up
- Branch lock file exists from crashed session

Resolution Steps:
1. List all existing worktrees to find where the branch is checked out
2. If the branch is in an orphaned worktree, run git worktree prune to clean up
3. If the branch is legitimately in use, specify a different branch name with the branch option
4. If you need the same code, consider creating a new branch based on the existing one

### Directory Permission Denied

Symptoms:
- Permission denied error when creating worktree directory
- Unable to write to worktree root location
- Creation succeeds partially but fails on file operations

Root Causes:
- Insufficient permissions on worktree root directory
- Parent directory does not exist
- File system is read-only or mounted incorrectly

Resolution Steps:
1. Verify you have write permissions on the worktree root directory
2. Ensure all parent directories exist and are accessible
3. Check file system mount options and permissions
4. If using custom worktree root, verify the path is writable

---

## Registry Issues

### Registry File Corruption

Symptoms:
- moai worktree commands fail while reading worktree administrative files
- Worktrees that exist on disk are missing from git worktree list
- Commands report a worktree not found despite its directory existing

Root Causes:
- Concurrent write operations corrupted the file
- Disk space ran out during write operation
- Manual editing introduced syntax errors
- Process was killed during registry update

Resolution Steps:
1. Run moai worktree recover — it scans the worktree root, runs git worktree repair, prunes stale references, and lists what it recovered
2. Confirm the result with git worktree list
3. If a directory is still missing from the list, check that it contains a valid .git file before retrying

### Orphaned Registry Entries

Symptoms:
- Registry lists worktrees that no longer exist on disk
- git worktree list shows entries marked prunable
- Commands fail when trying to operate on listed worktrees

Root Causes:
- Worktree directory was manually deleted
- External tool removed the directory
- File system error caused data loss
- Registry was not updated after worktree removal

Resolution Steps:
1. Run git worktree list to identify orphaned entries
2. For each orphaned entry, confirm the directory truly does not exist
3. Remove orphaned entries with git worktree prune, or run moai worktree recover to repair and prune together
4. Alternatively, recreate the worktree if the branch still exists

### Registry Sync Conflicts

Symptoms:
- Multiple developers see different worktree states
- Shared registry shows conflicting entries
- Team coordination failures

Root Causes:
- Registry is not version controlled
- Concurrent updates from different machines
- Network file system delays causing stale reads

Resolution Steps:
1. Identify which registry state is authoritative
2. Backup all registry versions before merging
3. Merge registry entries manually, keeping valid worktrees
4. Consider using developer-prefixed worktree IDs to avoid conflicts
5. Implement proper file locking for shared registry access

---

## Git State Problems

### Detached HEAD State

Symptoms:
- Git status shows detached HEAD state in worktree
- Commits are not attached to any branch
- Warning messages about commits being lost

Root Causes:
- Checkout of specific commit instead of branch
- Interrupted rebase or merge operation
- Manual Git operations in worktree

Resolution Steps:
1. Check if there are uncommitted changes that need saving
2. If commits exist that should be preserved, create a new branch from current HEAD
3. Checkout the intended branch for this worktree
4. If the worktree should track a specific branch, use git checkout branch-name

Prevention:
- Avoid manual git checkout of commit hashes in worktrees
- Complete rebase and merge operations before switching worktrees
- Use moai worktree commands for standard operations

### Merge Conflicts During Sync

Symptoms:
- Sync operation stops with conflict markers
- Files contain conflict markers that need resolution
- Unable to proceed with development until conflicts resolved

Root Causes:
- Base branch has changes that conflict with worktree changes
- Long-running worktree has diverged significantly from base
- Multiple developers modified same files

Resolution Steps:
1. Identify all files with conflict markers using git status
2. For each conflicted file, choose one resolution strategy:
   - Keep worktree version if your changes are correct
   - Accept base branch version if upstream changes take precedence
   - Manually merge by editing the file to combine both changes
3. After resolving each file, stage it with git add
4. Complete the merge with git commit
5. Verify resolution did not break functionality

Conflict Prevention Strategies:
- Sync frequently to minimize divergence
- Coordinate with team members on file ownership
- Break large changes into smaller, focused commits
- Use feature flags to allow parallel development

### Branch Not Found Error

Symptoms:
- Worktree references branch that does not exist
- Git operations fail with branch not found error
- Registry shows branch that was deleted remotely

Root Causes:
- Branch was deleted on remote after worktree creation
- Branch name was changed or renamed
- Local branch tracking was lost

Resolution Steps:
1. Check if branch exists locally using git branch --list
2. Check if branch exists on remote using git branch --remote
3. If branch was renamed, update worktree to track new branch name
4. If branch was deleted intentionally, consider removing the worktree
5. If branch was deleted accidentally, check reflog for recovery options

---

## Path and Permission Issues

### Path Resolution Failures

Symptoms:
- Commands cannot find worktree at expected path
- Relative paths not resolving correctly
- Path contains special characters causing errors

Root Causes:
- Working directory changed since worktree creation
- Path contains spaces or special characters not properly escaped
- Symlinks in path causing resolution issues
- Environment variables in path not expanded

Resolution Steps:
1. Use absolute paths instead of relative paths for worktree operations
2. Ensure paths with spaces are properly quoted in shell commands
3. Resolve symlinks to actual paths if causing issues
4. Verify environment variables are set and contain valid paths

### File Permission Errors

Symptoms:
- Cannot write to files within worktree
- Git operations fail with permission denied
- File creation succeeds but modification fails

Root Causes:
- File permissions changed after worktree creation
- Different user or group owns worktree files
- File system ACLs blocking access
- Read-only file system or mount

Resolution Steps:
1. Check file ownership matches current user
2. Verify file permissions allow write access
3. If permissions changed, reset with appropriate chmod command
4. Check for extended ACLs that may be restricting access
5. Verify file system is mounted with write permissions

---

## Integration Issues

### MoAI Command Coordination Failures

Symptoms:
- /moai plan does not create expected worktree
- /moai run cannot find worktree for SPEC
- Workflow commands operate on wrong worktree

Root Causes:
- SPEC ID not matching worktree ID format
- Worktree was created manually without MoAI integration
- Configuration mismatch between MoAI and worktree settings
- Working directory context not in expected location

Resolution Steps:
1. Verify SPEC ID format matches expected worktree naming convention
2. Ensure worktree was created through MoAI workflow or properly registered
3. Check moai configuration for worktree integration settings
4. Confirm working directory context before running workflow commands

### Auto-Detection Not Working

Symptoms:
- Worktree environment not detected when in worktree directory
- SPEC ID not automatically extracted from path
- Integration hooks not firing

Root Causes:
- Registry file not in expected parent directory location
- Worktree directory structure differs from expected pattern
- Environment detection script has errors
- Shell configuration not loading worktree functions

Resolution Steps:
1. Verify registry file exists in worktree parent directory
2. Check worktree directory follows expected naming pattern
3. Ensure shell profile loads worktree integration functions
4. Test detection manually by checking for registry file presence

---

## Diagnostic Commands

### Worktree State Verification

Status Commands:
- git worktree list: Native Git worktree listing
- git -C <path> status: Detailed status for a specific worktree
- moai worktree clean --json: Every non-protected worktree and its state, as JSON (removes nothing)

Registry Commands:
- moai worktree recover: Repair administrative files, prune stale references, list the result
- git worktree prune: Remove stale Git worktree metadata

Git State Commands:
- git status: Current worktree Git state
- git log --oneline -5: Recent commits in worktree
- git branch -vv: Branch tracking information

### Cleanup and Recovery

Cleanup Commands:
- moai worktree clean --stale: Preview abandoned worktrees (add --yes to remove them)
- moai worktree clean --merged-only: Clean only merged worktrees
- git worktree prune: Remove stale Git worktree metadata

Recovery Commands:
- moai worktree recover: Reconstruct administrative files from the directories on disk
- moai worktree remove <path>: Remove a worktree; its branch is kept

---

## Best Practices for Prevention

### Regular Maintenance

Weekly Tasks:
- Run git worktree list and moai worktree clean --json to check for issues
- Sync active worktrees with base branch to minimize conflicts
- Clean up merged worktrees to reduce clutter

Monthly Tasks:
- Prune stale Git worktree metadata
- Run moai worktree recover if the registry and the disk disagree
- Review and archive old worktrees

### Safe Operation Patterns

Creation:
- Always create worktrees with the moai cc -w <name> launcher instead of manual Git commands
- Verify branch name is unique before creation
- Use descriptive SPEC IDs for easy identification

Modification:
- Commit changes before switching worktrees
- Use sync command before making significant changes
- Resolve conflicts immediately rather than deferring

Removal:
- Use moai worktree remove instead of manual deletion
- remove keeps the branch; use moai worktree done <branch-name> --delete-branch to remove both
- Commit or push work before removing a worktree

### Team Coordination

Naming Conventions:
- Use developer prefix for worktree IDs in shared environments
- Follow consistent branch naming patterns
- Document worktree purposes in descriptions

Communication:
- Notify team before removing shared worktrees
- Coordinate on long-running feature branches
- Share worktree status updates for blocking issues

---

Version: 1.0.0
Module: Comprehensive troubleshooting and problem resolution
