---
name: moai-worktree
description: Git worktree management for parallel SPEC development with isolated workspaces, automatic registration, and seamless MoAI-ADK integration
version: 1.0.0
modularized: true
category: workflow
updated: 2025-11-29
status: active
tags:
 - git
 - worktree
 - parallel
 - development
 - spec
 - isolation
allowed-tools: Read, Grep, Glob
---

# MoAI Worktree Management

Git worktree management system for parallel SPEC development with isolated workspaces, automatic registration, and seamless MoAI-ADK integration.

Core Philosophy: Each SPEC deserves its own isolated workspace to enable true parallel development without context switching overhead.

## Quick Reference (30 seconds)

What is MoAI Worktree Management?
A specialized Git worktree system that creates isolated development environments for each SPEC, enabling parallel development without conflicts.

Key Features:
- Isolated Workspaces: Each SPEC gets its own worktree with independent Git state
- Automatic Registration: Worktree registry tracks all active workspaces
- Parallel Development: Multiple SPECs can be developed simultaneously
- Seamless Integration: Works with `/moai:1-plan`, `/moai:2-run`, `/moai:3-sync` workflow
- Smart Synchronization: Automatic sync with base branch when needed
- Cleanup Automation: Automatic cleanup of merged worktrees

Quick Access:
- CLI commands → [Worktree Commands Module](modules/worktree-commands.md)
- Management patterns → [Worktree Management Module](modules/worktree-management.md)
- Parallel workflow → [Parallel Development Module](modules/parallel-development.md)
- Integration guide → [Integration Patterns Module](modules/integration-patterns.md)
- Troubleshooting → [Troubleshooting Module](modules/troubleshooting.md)

Use Cases:
- Multiple SPECs development in parallel
- Isolated testing environments
- Feature branch isolation
- Code review workflows
- Experimental feature development

---

## Implementation Guide (5 minutes)

### 1. Core Architecture - Worktree Management System

Purpose: Create isolated Git worktrees for parallel SPEC development.

Key Components:

1. Worktree Registry - Central registry tracking all worktrees
2. Manager Layer - Core worktree operations (create, switch, remove, sync)
3. CLI Interface - User-friendly command interface
4. Models - Data structures for worktree metadata
5. Integration Layer - MoAI-ADK workflow integration

Registry Structure:
```json
{
 "worktrees": {
 "SPEC-001": {
 "id": "SPEC-001",
 "path": "/Users/goos/worktrees/MoAI-ADK/SPEC-001",
 "branch": "feature/SPEC-001-user-auth",
 "created_at": "2025-11-29T22:00:00Z",
 "last_sync": "2025-11-29T22:30:00Z",
 "status": "active",
 "base_branch": "main"
 }
 },
 "config": {
 "worktree_root": "/Users/goos/worktrees/MoAI-ADK",
 "auto_sync": true,
 "cleanup_merged": true
 }
}
```

File Structure:
```
~/worktrees/MoAI-ADK/
 .moai-worktree-registry.json # Central registry
 SPEC-001/ # Worktree directory
 .git # Git worktree metadata
 (project files) # Complete project copy
 SPEC-002/ # Another worktree
 .git
 (project files)
```

Detailed Reference: [Worktree Management Module](modules/worktree-management.md)

---

### 2. CLI Commands - Complete Command Interface

Purpose: Provide intuitive CLI commands for worktree management.

Core Commands:

```bash
# Create new worktree for SPEC
moai-worktree new SPEC-001 "User Authentication System"

# List all worktrees
moai-worktree list

# Switch to specific worktree
moai-worktree switch SPEC-001

# Get worktree path for shell integration
eval $(moai-worktree go SPEC-001)

# Sync worktree with base branch
moai-worktree sync SPEC-001

# Remove worktree
moai-worktree remove SPEC-001

# Clean up merged worktrees
moai-worktree clean

# Show worktree status
moai-worktree status

# Configuration management
moai-worktree config get
moai-worktree config set worktree_root ~/my-worktrees
```

Command Categories:

1. Creation: `new` - Create isolated worktree
2. Navigation: `list`, `switch`, `go` - Browse and navigate
3. Management: `sync`, `remove`, `clean` - Maintain worktrees
4. Status: `status` - Check worktree state
5. Configuration: `config` - Manage settings

Shell Integration:
```bash
# Direct switching
moai-worktree switch SPEC-001

# Shell eval pattern (recommended)
eval $(moai-worktree go SPEC-001)

# Output example:
# cd /Users/goos/worktrees/MoAI-ADK/SPEC-001
```

Detailed Reference: [Worktree Commands Module](modules/worktree-commands.md)

---

### 3. Parallel Development Workflow - Isolated SPEC Development

Purpose: Enable true parallel development without context switching.

Workflow Integration:

```
1. Plan Phase (/moai:1-plan):
 SPEC creation
 moai-worktree new SPEC-XXX
 Automatic worktree setup

2. Development Phase:
 Isolated worktree environment
 Independent Git state
 Zero context switching

3. Sync Phase (/moai:3-sync):
 moai-worktree sync SPEC-XXX
 Clean integration
 Conflict resolution

4. Cleanup Phase:
 moai-worktree clean
 Automatic cleanup
 Registry maintenance
```

Parallel Development Benefits:

1. Context Isolation: Each SPEC has its own Git state, files, and environment
2. Zero Switching Cost: Instant switching between worktrees
3. Independent Development: Work on multiple SPECs simultaneously
4. Safe Experimentation: Isolated environment for experimental features
5. Clean Integration: Automatic sync and conflict resolution

Example Workflow:
```bash
# Create SPEC-001 worktree
moai-worktree new SPEC-001 "User Authentication"
cd $(moai-worktree go SPEC-001)

# Develop in isolation
/moai:2-run SPEC-001

# Create another worktree
cd /path/to/main/repo
moai-worktree new SPEC-002 "Payment Integration"
cd $(moai-worktree go SPEC-002)

# Parallel development on SPEC-002
/moai:2-run SPEC-002

# Switch back and continue
moai-worktree switch SPEC-001
# Continue development...

# Sync and integrate
moai-worktree sync SPEC-001
moai-worktree sync SPEC-002
```

Detailed Reference: [Parallel Development Module](modules/parallel-development.md)

---

### 4. Integration Patterns - MoAI-ADK Workflow Integration

Purpose: Seamless integration with MoAI-ADK Plan-Run-Sync workflow.

Integration Points:

1. Plan Phase Integration (`/moai:1-plan`):
 ```bash
 # After SPEC creation
 moai-worktree new SPEC-{SPEC_ID}

 # Output guidance
 echo "1. Switch to worktree: moai-worktree switch SPEC-{SPEC_ID}"
 echo "2. Or use shell eval: eval $(moai-worktree go SPEC-{SPEC_ID})"
 ```

2. Development Phase (`/moai:2-run`):
 - Worktree isolation provides clean development environment
 - Independent Git state prevents conflicts
 - Automatic registry tracking

3. Sync Phase (`/moai:3-sync`):
 ```bash
 # Before PR creation
 moai-worktree sync SPEC-{SPEC_ID}

 # After PR merge
 moai-worktree clean --merged-only
 ```

Auto-Detection Patterns:

```bash
# Check if worktree environment
if [ -f "../.moai-worktree-registry.json" ]; then
 SPEC_ID=$(basename $(pwd))
 echo "Detected worktree environment: $SPEC_ID"
fi

# Auto-sync on status check
moai-worktree status
# Automatically syncs if out of date
```

Configuration Integration:
```json
{
 "moai": {
 "worktree": {
 "auto_create": true,
 "auto_sync": true,
 "cleanup_merged": true,
 "worktree_root": "~/worktrees/{PROJECT_NAME}"
 }
 }
}
```

Detailed Reference: [Integration Patterns Module](modules/integration-patterns.md)

---

## Advanced Implementation (10+ minutes)

### Multi-Developer Worktree Coordination

Shared Worktree Registry:
```bash
# Team worktree configuration
moai-worktree config set registry_type team
moai-worktree config set shared_registry ~/team-worktrees/MoAI-ADK/

# Developer-specific worktrees
moai-worktree new SPEC-001 --developer alice
moai-worktree new SPEC-001 --developer bob

# Team coordination
moai-worktree list --all-developers
moai-worktree status --team-overview
```

### Advanced Synchronization Strategies

Selective Sync Patterns:
```bash
# Sync only specific files
moai-worktree sync SPEC-001 --include "src/"

# Exclude specific patterns
moai-worktree sync SPEC-001 --exclude "node_modules/"

# Auto-resolve conflicts
moai-worktree sync SPEC-001 --auto-resolve

# Interactive conflict resolution
moai-worktree sync SPEC-001 --resolve-interactive
```

### Worktree Templates and Presets

Custom Worktree Templates:
```bash
# Create worktree with specific setup
moai-worktree new SPEC-001 --template frontend
# Includes: npm install, eslint setup, pre-commit hooks

moai-worktree new SPEC-002 --template backend
# Includes: virtual environment, test setup, database config

# Custom template configuration
moai-worktree config set templates.frontend.setup "npm install && npm run setup"
moai-worktree config set templates.backend.setup "python -m venv venv && source venv/bin/activate && pip install -r requirements.txt"
```

### Performance Optimization

Optimized Worktree Operations:
```bash
# Fast worktree creation (shallow clone)
moai-worktree new SPEC-001 --shallow --depth 1

# Background sync
moai-worktree sync SPEC-001 --background

# Parallel operations
moai-worktree sync --all --parallel

# Caching for faster operations
moai-worktree config set enable_cache true
moai-worktree config set cache_ttl 3600
```

---

## Works Well With

Commands:
- moai:1-plan - SPEC creation with automatic worktree setup
- moai:2-run - Development in isolated worktree environment
- moai:3-sync - Integration with automatic worktree sync
- moai:9-feedback - Worktree workflow improvements

Skills:
- moai-foundation-core - Parallel development patterns
- moai-workflow-project - Project management integration
- moai-workflow-spec - SPEC-driven development
- moai-git-strategy - Git workflow optimization

Tools:
- Git worktree - Native Git worktree functionality
- Rich CLI - Formatted terminal output
- Click framework - Command-line interface framework

---

## Quick Decision Matrix

| Scenario | Primary Pattern | Supporting |
|----------|------------------|------------|
| New SPEC development | Worktree isolation + Auto-setup | Integration with /moai:1-plan |
| Parallel development | Multiple worktrees + Shell integration | Fast switching patterns |
| Team coordination | Shared registry + Developer prefixes | Conflict resolution |
| Code review | Isolated review worktrees | Clean sync patterns |
| Experimental features | Temporary worktrees + Auto-cleanup | Safe experimentation |

Module Deep Dives:
- [Worktree Commands](modules/worktree-commands.md) - Complete CLI reference
- [Worktree Management](modules/worktree-management.md) - Core architecture
- [Parallel Development](modules/parallel-development.md) - Workflow patterns
- [Integration Patterns](modules/integration-patterns.md) - MoAI-ADK integration
- [Troubleshooting](modules/troubleshooting.md) - Problem resolution

Full Examples: [examples.md](examples.md)
External Resources: [reference.md](reference.md)

---

Version: 1.0.0
Last Updated: 2025-11-29
Status: Active (Complete modular architecture)
