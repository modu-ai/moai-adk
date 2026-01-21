# Worktree Management Module

Purpose: Core architecture and management patterns for Git worktree operations including registry management, lifecycle control, and resource optimization.

Version: 2.0.0
Last Updated: 2026-01-06

---

## Quick Reference (30 seconds)

Core Components:
- Registry: Central worktree database and metadata tracking
- Manager: Core operations including create, sync, remove, and cleanup
- Models: Data structures for worktree metadata
- Lifecycle: Worktree creation, maintenance, and removal patterns

Key Patterns:
- Automatic registration and tracking
- Atomic operations with rollback
- Resource optimization and cleanup
- Conflict detection and resolution

---

## Module Architecture

This worktree management module is organized into focused sub-modules:

Registry Architecture: Refer to registry-architecture.md
- Registry file structure and schema
- Atomic operations with backup and rollback
- Concurrent access protection with locking
- Validation and integrity checks

Resource Optimization: Refer to resource-optimization.md
- Disk space management and analysis
- Memory-efficient operations
- Performance optimization patterns
- Error handling and recovery

---

## Core Architecture

### Component Overview

WorktreeManager: Central coordinator for all worktree operations.
- Accepts repository path and worktree root directory
- Maintains references to RegistryManager and GitManager
- Provides unified interface for all operations

RegistryManager: Handles persistent storage of worktree metadata.
- JSON-based registry file storage
- Atomic updates with backup/rollback
- Concurrent access protection

GitManager: Interfaces with Git for worktree operations.
- Creates and removes Git worktrees
- Manages branch operations
- Handles sync and merge operations

### Manager Operations

Core Operations:
- create_worktree: Create new isolated worktree with optional template
- sync_worktree: Synchronize worktree with base branch
- remove_worktree: Remove worktree and clean up registration
- switch_worktree: Change to specified worktree directory
- cleanup_worktrees: Batch cleanup based on criteria

Query Operations:
- list_worktrees: Return all worktrees matching optional filters
- get_worktree: Return specific worktree by ID
- get_current_worktree: Detect and return current worktree context
- get_worktree_status: Return detailed status for worktree

Detailed Reference: Refer to registry-architecture.md for registry patterns

---

## Lifecycle Management

### Creation Workflow

Complete worktree creation follows these steps:

1. Validate Input: Check SPEC ID format and uniqueness
2. Determine Configuration: Branch name, base branch, template selection
3. Create Worktree Path: Establish directory structure
4. Create Git Worktree: Use git worktree add command
5. Apply Template: Run setup commands and create files if template specified
6. Register Worktree: Add entry to central registry
7. Post-Creation Hooks: Execute any configured post-creation actions

On Failure:
- Partial worktree directory is removed
- Registry entry is not created or removed if partially created
- Error details captured for debugging

### Synchronization Workflow

Worktree sync with base branch follows these steps:

1. Check Prerequisites: Verify no uncommitted changes (unless forced)
2. Fetch Updates: Get latest changes from remote
3. Analyze Needs: Determine commits ahead/behind
4. Execute Strategy: Merge, rebase, or squash based on configuration
5. Resolve Conflicts: Auto-resolve or interactive based on options
6. Update Registry: Record sync timestamp and result
7. Post-Sync Hooks: Execute any configured post-sync actions

Sync Strategies:
- merge: Preserve history with merge commit
- rebase: Linear history by replaying commits
- squash: Combine all changes into single commit

### Cleanup Workflow

Worktree cleanup based on various criteria:

Finding Cleanup Candidates:
- Merged worktrees: Branch merged to base branch
- Stale worktrees: Not accessed within threshold days
- Large worktrees: Exceeding size threshold

Cleanup Process:
1. Identify candidates based on criteria
2. Sort by priority (merged first, then stale, then large)
3. Interactive selection if requested
4. Remove each worktree with optional backup
5. Update registry and statistics

Detailed Reference: Refer to resource-optimization.md for optimization patterns

---

## Registry Overview

### Registry Purpose

The registry is the central database tracking all worktrees and their metadata.

Storage Location: ~/.worktrees/{PROJECT_NAME}/.moai-worktree-registry.json

Registry Contents:
- version: Schema version for compatibility
- created_at: Registry creation timestamp
- last_updated: Most recent modification timestamp
- config: Global worktree configuration
- worktrees: Map of SPEC ID to worktree metadata
- statistics: Aggregated usage statistics

### Worktree Entry Structure

Each worktree entry contains:

Identity:
- id: SPEC identifier
- description: Human-readable description
- path: Absolute filesystem path

Git State:
- branch: Feature branch name
- base_branch: Branch created from
- commits_ahead: Commits ahead of base
- commits_behind: Commits behind base

Status:
- status: active, merged, stale, or error
- created_at: Creation timestamp
- last_accessed: Most recent access
- last_sync: Most recent synchronization

Metadata:
- template: Template used for creation
- developer: Creator identifier
- priority: Development priority
- tags: Categorization tags

Detailed Reference: Refer to registry-architecture.md for complete schema

---

## Quick Decision Guide

For creating a new worktree, use create_worktree with SPEC ID, description, and optional template. The system handles branch creation, registration, and setup automatically.

For synchronizing a worktree, use sync_worktree with SPEC ID. Configure strategy (merge, rebase, squash) and conflict resolution (auto-resolve, interactive) as needed.

For cleaning up worktrees, use cleanup_worktrees with appropriate filters. Target merged branches first, then stale worktrees, then large worktrees.

For registry maintenance, the system handles most operations automatically. Manual intervention only needed for corruption recovery or registry migration.

---

## Sub-Module References

Registry Architecture (registry-architecture.md):
- Complete registry JSON schema
- Atomic update patterns with rollback
- Concurrent access protection with file locking
- Schema validation and integrity checks
- Data migration patterns

Resource Optimization (resource-optimization.md):
- Disk usage analysis and cleanup
- Memory-efficient streaming operations
- Performance optimization techniques
- Comprehensive error handling
- Recovery patterns for common failures

---

Version: 2.0.0
Last Updated: 2026-01-06
Module: Worktree management overview with progressive disclosure to sub-modules
