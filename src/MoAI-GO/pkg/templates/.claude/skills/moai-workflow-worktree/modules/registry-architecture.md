# Registry Architecture Module

Purpose: Detailed registry structure, atomic operations, concurrent access protection, and validation patterns for worktree metadata management.

Version: 1.0.0
Last Updated: 2026-01-06

---

## Quick Reference (30 seconds)

Registry Components:
- File Location: {repo}/.moai/worktrees/.moai-worktree-registry.json
- Format: JSON with versioned schema
- Operations: Atomic updates with backup and rollback
- Concurrency: File-based locking for multi-process safety

---

## Registry Schema

### Complete Registry Structure

The registry file contains these top-level sections:

Version and Timestamps:
- version: Schema version string (e.g., "1.0.0")
- created_at: ISO 8601 timestamp of registry creation
- last_updated: ISO 8601 timestamp of most recent modification

Configuration Section:
- worktree_root: Absolute path to worktree base directory
- auto_sync: Boolean flag for automatic synchronization
- cleanup_merged: Boolean flag for automatic cleanup
- default_base: Default base branch name
- sync_strategy: Preferred sync method (merge, rebase, squash)
- registry_type: local or team mode
- max_worktrees: Maximum allowed worktrees

Worktrees Map:
- Key: SPEC ID (e.g., "SPEC-001")
- Value: Complete worktree metadata object

Statistics Section:
- total_worktrees: Count of all registered worktrees
- active_worktrees: Count of active worktrees
- merged_worktrees: Count of merged worktrees
- total_disk_usage: Aggregate storage size
- last_cleanup: Timestamp of most recent cleanup

### Worktree Entry Schema

Each worktree entry contains nested sections:

Identity Fields:
- id: SPEC identifier (required)
- description: Human-readable description (optional)
- path: Absolute filesystem path (required)

Branch Information:
- branch: Feature branch name (required)
- base_branch: Parent branch name (required)
- status: Current state - active, merged, stale, or error (required)
- created_at: ISO 8601 creation timestamp (required)
- last_accessed: ISO 8601 last access timestamp
- last_sync: ISO 8601 last sync timestamp

Git State (git_info section):
- commits_ahead: Integer count of commits ahead of base
- commits_behind: Integer count of commits behind base
- uncommitted_changes: Boolean indicating uncommitted changes
- branch_status: ahead, behind, diverged, or up-to-date
- merge_conflicts: Boolean indicating unresolved conflicts

Metadata (metadata section):
- template: Template used for creation
- developer: Developer identifier
- priority: high, medium, or low
- estimated_size: Estimated disk usage
- tags: Array of categorization tags

Operations (operations section):
- total_syncs: Count of sync operations
- total_conflicts: Count of conflicts encountered
- last_operation: Name of most recent operation
- last_operation_status: success or error

---

## Atomic Operations

### Update Pattern with Backup

All registry updates follow an atomic pattern to prevent corruption.

Update Workflow:
1. Create backup of current registry with .backup.json suffix
2. Load current registry into memory
3. Validate proposed updates against schema
4. Apply updates to in-memory copy
5. Write updated registry to temporary file with .tmp.json suffix
6. Atomic rename of temporary file to registry path
7. Remove backup file on success

Rollback on Failure:
- If any step fails, restore from backup file
- Temporary file is removed on failure
- Error details are captured and raised

### Validation Before Update

Before applying updates, validate:
- SPEC ID format matches pattern ^SPEC-[0-9]+$
- Path uniqueness across all worktrees
- Required fields are present
- Timestamp formats are valid ISO 8601
- Status values are from allowed enum

---

## Concurrent Access Protection

### File Locking

Multiple processes may access the registry simultaneously. Protection uses file-based locking.

Lock Acquisition:
- Create lock file with .lock suffix
- Acquire exclusive lock using fcntl
- Non-blocking attempt first
- If blocked, retry with timeout (30 seconds default)
- Raise error if lock cannot be acquired

Lock Release:
- Execute protected operation within lock context
- Release lock in finally block
- Remove lock file after release

### Timeout Handling

If lock cannot be acquired within timeout:
- Log warning about long-running operation
- Raise RegistryError with descriptive message
- Caller can retry or abort

---

## Registry Validation

### Schema Validation

Validate registry structure against JSON Schema.

Top-Level Requirements:
- Required fields: version, created_at, last_updated, config, worktrees
- Config must include: worktree_root, default_base
- Worktrees must be object with valid pattern keys

Worktree Entry Requirements:
- Required fields: id, path, branch, status, created_at
- Status must be one of: active, merged, stale, error
- Timestamps must be valid ISO 8601 format

### Integrity Checks

Beyond schema validation, verify:

Path Existence:
- Each worktree path must exist on filesystem
- Each path must contain .git file (worktree marker)

Path Uniqueness:
- No duplicate paths across worktrees
- Paths must be canonical (no symlink variations)

Git Repository Validation:
- Verify worktree is valid Git repository
- Check branch matches expected value
- Verify remote tracking configuration

---

## Registry Management Patterns

### Worktree Registration

When registering a new worktree:

Create Entry:
- Generate worktree entry with all required fields
- Set initial status to active
- Set created_at and last_accessed to current timestamp
- Initialize git_info with current state
- Initialize operations counters to zero

Update Registry:
- Add entry to worktrees map
- Update statistics (increment total_worktrees, active_worktrees)
- Set last_updated timestamp

### Worktree Update

When updating existing worktree:

Common Updates:
- last_accessed: Updated on each access
- last_sync: Updated after sync operation
- git_info: Refreshed with current Git state
- operations: Increment counters, update last operation

Status Transitions:
- active -> merged: When branch merged to base
- active -> stale: When last_accessed exceeds threshold
- active -> error: When operation fails
- any -> active: When worktree recovered

### Worktree Removal

When removing worktree:

Registry Update:
- Remove entry from worktrees map
- Update statistics (decrement counters)
- Optionally record in completed_specs for audit

Cleanup:
- Remove filesystem path
- Remove Git worktree reference
- Delete associated branch if requested

---

## Error Recovery

### Corruption Detection

Detect registry corruption by:
- JSON parse failure
- Schema validation failure
- Missing required sections
- Invalid data types

### Recovery Options

On Corruption:
1. Check for backup file (.backup.json)
2. If backup exists and valid, restore from backup
3. If no valid backup, scan filesystem for worktrees
4. Rebuild registry from discovered worktrees
5. Log recovery actions for audit

### Migration Patterns

When schema version changes:

Version Check:
- Compare registry version with expected version
- If older, apply migrations in sequence

Migration Steps:
- Backup current registry
- Apply transformation for each version increment
- Update version field
- Validate migrated registry
- Save updated registry

---

Version: 1.0.0
Last Updated: 2026-01-06
Module: Registry architecture with schema, operations, and validation patterns
