# Worktree Commands Module

Purpose: Complete CLI command reference for Git worktree management with detailed usage examples and advanced options.

Version: 1.0.0
Last Updated: 2025-11-29

---

## Quick Reference (30 seconds)

Command Categories:
- Creation: `new` - Create isolated worktree
- Navigation: `list`, `switch`, `go` - Browse and navigate
- Management: `sync`, `remove`, `clean` - Maintain worktrees
- Status: `status` - Check worktree state
- Configuration: `config` - Manage settings

Quick Start:
```bash
moai-worktree new SPEC-001 "User Authentication"
moai-worktree switch SPEC-001
# or: eval $(moai-worktree go SPEC-001)
```

---

## Core Commands

### `moai-worktree new` - Create Worktree

Create a new isolated Git worktree for SPEC development.

Syntax:
```bash
moai-worktree new <spec-id> [description] [options]
```

Arguments:
- `<spec-id>`: SPEC identifier (e.g., SPEC-001, SPEC-AUTH-001)
- `[description]`: Optional description for the worktree

Options:
- `--branch <branch-name>`: Create specific branch instead of auto-generated
- `--base <branch>`: Base branch for new worktree (default: main)
- `--template <template>`: Use predefined template
- `--shallow`: Create shallow clone for faster setup
- `--depth <number>`: Clone depth for shallow clone
- `--force`: Force creation even if worktree exists

Examples:
```bash
# Basic worktree creation
moai-worktree new SPEC-001 "User Authentication System"

# Custom branch name
moai-worktree new SPEC-002 "Payment Integration" --branch feature/payment-gateway

# From develop branch
moai-worktree new SPEC-003 "API Refactoring" --base develop

# Using template
moai-worktree new SPEC-004 "Frontend Overhaul" --template frontend

# Fast creation with shallow clone
moai-worktree new SPEC-005 "Bug Fixes" --shallow --depth 1
```

Auto-Generated Branch Pattern:
- Default: `feature/SPEC-{ID}-{description-kebab-case}`
- Example: `SPEC-001` → `feature/SPEC-001-user-authentication`

Output:
```bash
 Created worktree: SPEC-001
 Branch: feature/SPEC-001-user-authentication
 Path: /Users/goos/worktrees/MoAI-ADK/SPEC-001
 Registered in: /Users/goos/worktrees/MoAI-ADK/.moai-worktree-registry.json

Next steps:
1. Switch to worktree: moai-worktree switch SPEC-001
2. Or use shell eval: eval $(moai-worktree go SPEC-001)
```

---

### `moai-worktree list` - List Worktrees

Display all registered worktrees with their status and metadata.

Syntax:
```bash
moai-worktree list [options]
```

Options:
- `--format <format>`: Output format (table, json, csv)
- `--status <status>`: Filter by status (active, merged, stale)
- `--sort <field>`: Sort by field (name, created, modified, status)
- `--reverse`: Reverse sort order
- `--verbose`: Show detailed information

Examples:
```bash
# Table format (default)
moai-worktree list

# JSON output for scripting
moai-worktree list --format json

# Show only active worktrees
moai-worktree list --status active

# Sort by creation date
moai-worktree list --sort created

# Detailed view
moai-worktree list --verbose
```

Table Output:
```bash

 ID Description Path Status Last Sync 

 SPEC-001 User Authentication System /worktrees/MoAI-ADK/ active 2h ago 
 SPEC-002 Payment Integration /worktrees/MoAI-ADK/ active 1d ago 
 SPEC-003 API Refactoring /worktrees/MoAI-ADK/ merged 3d ago 

```

JSON Output:
```json
{
 "worktrees": [
 {
 "id": "SPEC-001",
 "description": "User Authentication System",
 "path": "/Users/goos/worktrees/MoAI-ADK/SPEC-001",
 "branch": "feature/SPEC-001-user-auth",
 "status": "active",
 "created_at": "2025-11-29T20:00:00Z",
 "last_sync": "2025-11-29T22:00:00Z",
 "base_branch": "main",
 "commits_ahead": 5,
 "commits_behind": 0
 }
 ],
 "total_count": 1,
 "active_count": 1
}
```

---

### `moai-worktree switch` - Switch to Worktree

Change current working directory to the specified worktree.

Syntax:
```bash
moai-worktree switch <spec-id> [options]
```

Arguments:
- `<spec-id>`: Target worktree identifier

Options:
- `--auto-sync`: Automatically sync before switching
- `--force`: Force switch even with uncommitted changes
- `--new-terminal`: Open in new terminal window

Examples:
```bash
# Basic switch
moai-worktree switch SPEC-001

# Switch with auto-sync
moai-worktree switch SPEC-002 --auto-sync

# Force switch (with warning)
moai-worktree switch SPEC-003 --force
```

Output:
```bash
 Switched to worktree: SPEC-001
 Current directory: /Users/goos/worktrees/MoAI-ADK/SPEC-001
 Branch: feature/SPEC-001-user-auth
 Status: 5 commits ahead of main
```

---

### `moai-worktree go` - Get Worktree Path

Output the `cd` command for shell integration.

Syntax:
```bash
moai-worktree go <spec-id> [options]
```

Arguments:
- `<spec-id>`: Target worktree identifier

Options:
- `--absolute`: Show absolute path
- `--relative`: Show relative path from current directory
- `--export`: Export as environment variable

Examples:
```bash
# Standard usage (shell eval)
eval $(moai-worktree go SPEC-001)

# Absolute path output
moai-worktree go SPEC-001 --absolute

# Relative path
moai-worktree go SPEC-001 --relative

# Export as variable
moai-worktree go SPEC-001 --export
```

Shell Integration:
```bash
# Method 1: eval (recommended)
eval $(moai-worktree go SPEC-001)

# Method 2: source
moai-worktree go SPEC-001 | source

# Method 3: manual
cd $(moai-worktree go SPEC-001 --absolute)
```

Output:
```bash
# Standard output
cd /Users/goos/worktrees/MoAI-ADK/SPEC-001

# With --export
export WORKTREE_PATH="/Users/goos/worktrees/MoAI-ADK/SPEC-001"
cd "$WORKTREE_PATH"
```

---

## Management Commands

### `moai-worktree sync` - Synchronize Worktree

Synchronize worktree with its base branch.

Syntax:
```bash
moai-worktree sync <spec-id> [options]
```

Arguments:
- `<spec-id>`: Worktree identifier (or `--all` for all worktrees)

Options:
- `--auto-resolve`: Automatically resolve simple conflicts
- `--interactive`: Interactive conflict resolution
- `--dry-run`: Show what would be synced without doing it
- `--force`: Force sync even with uncommitted changes
- `--include <pattern>`: Include only specific files
- `--exclude <pattern>`: Exclude specific files

Examples:
```bash
# Sync specific worktree
moai-worktree sync SPEC-001

# Sync all worktrees
moai-worktree sync --all

# Interactive conflict resolution
moai-worktree sync SPEC-001 --interactive

# Dry run to preview changes
moai-worktree sync SPEC-001 --dry-run

# Include only source files
moai-worktree sync SPEC-001 --include "src/"

# Exclude build artifacts
moai-worktree sync SPEC-001 --exclude "node_modules/" --exclude "dist/"
```

Conflict Resolution Options:

When conflicts are detected:
```bash
Conflict detected in src/auth.py
Choose resolution:
1) Keep worktree version (current)
2) Accept base branch version
3) Open merge tool
4) Skip this file
5) Abort sync

Choice [1-5]:
```

Output:
```bash
 Syncing SPEC-001 with main branch
 Fetching latest changes...
 5 new commits in main branch
 Merging changes into feature/SPEC-001-user-auth
 Sync completed successfully
 Worktree is now up-to-date
```

---

### `moai-worktree remove` - Remove Worktree

Remove a worktree and clean up its registration.

Syntax:
```bash
moai-worktree remove <spec-id> [options]
```

Arguments:
- `<spec-id>`: Worktree identifier to remove

Options:
- `--force`: Force removal without confirmation
- `--keep-branch`: Keep the branch after removing worktree
- `--backup`: Create backup before removal
- `--dry-run`: Show what would be removed without doing it

Examples:
```bash
# Interactive removal
moai-worktree remove SPEC-001

# Force removal
moai-worktree remove SPEC-001 --force

# Keep branch for future use
moai-worktree remove SPEC-001 --keep-branch

# Create backup
moai-worktree remove SPEC-001 --backup

# Preview removal
moai-worktree remove SPEC-001 --dry-run
```

Interactive Confirmation:
```bash
Are you sure you want to remove worktree SPEC-001?
Path: /Users/goos/worktrees/MoAI-ADK/SPEC-001
Branch: feature/SPEC-001-user-auth (5 commits ahead)

Options:
1) Remove worktree and branch
2) Remove worktree, keep branch
3) Create backup then remove
4) Cancel

Choice [1-4]:
```

Output:
```bash
 Removing worktree: SPEC-001
 Path: /Users/goos/worktrees/MoAI-ADK/SPEC-001
 Branch: feature/SPEC-001-user-auth (merged)
 Backup created: /Users/goos/worktrees/MoAI-ADK/.backups/SPEC-001-20251129.tar.gz
 Registration removed
 Worktree removed successfully
```

---

### `moai-worktree clean` - Clean Up Worktrees

Remove worktrees for merged branches or stale worktrees.

Syntax:
```bash
moai-worktree clean [options]
```

Options:
- `--merged-only`: Only remove worktrees with merged branches
- `--stale`: Remove worktrees not updated in specified days
- `--days <number>`: Stale threshold in days (default: 30)
- `--interactive`: Interactive selection of worktrees to remove
- `--dry-run`: Show what would be cleaned without doing it
- `--force`: Skip confirmation prompts

Examples:
```bash
# Clean merged worktrees
moai-worktree clean --merged-only

# Clean stale worktrees (not updated in 30 days)
moai-worktree clean --stale

# Custom stale threshold (14 days)
moai-worktree clean --stale --days 14

# Interactive cleaning
moai-worktree clean --interactive

# Preview what would be cleaned
moai-worktree clean --dry-run

# Force clean without prompts
moai-worktree clean --force
```

Interactive Selection:
```bash
Found 3 worktrees eligible for cleanup:

1. SPEC-003 (merged) - API Refactoring
 Path: /worktrees/MoAI-ADK/SPEC-003
 Last updated: 2025-11-15

2. SPEC-005 (stale) - Bug Fixes
 Path: /worktrees/MoAI-ADK/SPEC-005
 Last updated: 2025-10-20 (40 days ago)

3. SPEC-007 (merged) - Performance Updates
 Path: /worktrees/MoAI-ADK/SPEC-007
 Last updated: 2025-11-10

Select worktrees to remove (space-separated numbers): 1 3
```

Output:
```bash
 Scanning for cleanup candidates...
 Found 3 worktrees to clean
 Removing SPEC-003 (merged)
 Removing SPEC-007 (merged)
 2 worktrees removed
 1 worktree kept (SPEC-005 - stale but protected)
 Cleanup completed
```

---

## Status and Configuration Commands

### `moai-worktree status` - Show Worktree Status

Display detailed status information about worktrees.

Syntax:
```bash
moai-worktree status [spec-id] [options]
```

Arguments:
- `[spec-id]`: Specific worktree (optional, shows current if not specified)

Options:
- `--all`: Show status of all worktrees
- `--sync-check`: Check if worktrees need sync
- `--detailed`: Show detailed Git status
- `--format <format>`: Output format (table, json)

Examples:
```bash
# Current worktree status
moai-worktree status

# Specific worktree status
moai-worktree status SPEC-001

# All worktrees with sync check
moai-worktree status --all --sync-check

# Detailed Git status
moai-worktree status SPEC-001 --detailed

# JSON output
moai-worktree status --all --format json
```

Current Worktree Status:
```bash
Worktree: SPEC-001 (current)
Path: /Users/goos/worktrees/MoAI-ADK/SPEC-001
Branch: feature/SPEC-001-user-auth
Base: main

Git Status:
- 5 commits ahead of main
- 0 commits behind main
- 3 modified files
- 2 untracked files

Sync Status: Up to date
Last Sync: 2 hours ago
```

All Worktrees Status:
```bash

 ID Branch Status Ahead Behind Sync Need 

 SPEC-001 *user-auth active 5 0 No 
 SPEC-002 payment active 2 1 Yes 
 SPEC-003 refactor merged 0 0 No 

```

---

### `moai-worktree config` - Configuration Management

Manage moai-worktree configuration settings.

Syntax:
```bash
moai-worktree config <action> [key] [value]
```

Actions:
- `get [key]`: Get configuration value
- `set <key> <value>`: Set configuration value
- `list`: List all configuration
- `reset [key]`: Reset to default value
- `edit`: Open configuration in editor

Configuration Keys:
- `worktree_root`: Root directory for worktrees
- `auto_sync`: Enable automatic sync (true/false)
- `cleanup_merged`: Auto-cleanup merged worktrees (true/false)
- `default_base`: Default base branch (main/develop)
- `template_dir`: Directory for worktree templates
- `sync_strategy`: Sync strategy (merge, rebase, squash)

Examples:
```bash
# Show all configuration
moai-worktree config list

# Get specific value
moai-worktree config get worktree_root

# Set configuration
moai-worktree config set worktree_root ~/my-worktrees
moai-worktree config set auto_sync true
moai-worktree config set default_base develop

# Reset to default
moai-worktree config reset worktree_root

# Edit configuration
moai-worktree config edit
```

Configuration Output:
```bash
Current Configuration:
worktree_root: /Users/goos/worktrees/MoAI-ADK
auto_sync: true
cleanup_merged: true
default_base: main
template_dir: ~/.moai-worktree/templates
sync_strategy: merge
registry_type: local
```

---

## Advanced Usage Patterns

### Batch Operations

```bash
# Sync all active worktrees
for spec in $(moai-worktree list --status active --format json | jq -r '.worktrees[].id'); do
 moai-worktree sync "$spec"
done

# Clean all merged worktrees
moai-worktree clean --merged-only --force

# Create worktrees from SPEC list
cat specs.txt | xargs -I {} moai-worktree new {} "Auto-generated worktree"
```

### Shell Aliases and Functions

```bash
# Add to ~/.bashrc or ~/.zshrc
alias mw='moai-worktree'
alias mwl='moai-worktree list'
alias mws='moai-worktree switch'
alias mwg='eval $(moai-worktree go'

# Function for quick SPEC worktree creation
mwnew() {
 local spec_id="$1"
 local description="$2"
 moai-worktree new "$spec_id" "$description"
 moai-worktree switch "$spec_id"
}

# Function for worktree status overview
mwstatus() {
 echo "=== Worktree Overview ==="
 moai-worktree status --all
 echo ""
 echo "=== Current Worktree ==="
 moai-worktree status
}
```

### Integration with Git Hooks

```bash
# .git/hooks/post-checkout
#!/bin/bash
if [ -f "../.moai-worktree-registry.json" ]; then
 SPEC_ID=$(basename $(pwd))
 moai-worktree status "$SPEC_ID" --sync-check
fi

# .git/hooks/pre-push
#!/bin/bash
if [ -f "../.moai-worktree-registry.json" ]; then
 SPEC_ID=$(basename $(pwd))
 echo "Pushing from worktree: $SPEC_ID"
fi
```

---

Version: 1.0.0
Last Updated: 2025-11-29
Module: Complete CLI command reference with advanced usage patterns
