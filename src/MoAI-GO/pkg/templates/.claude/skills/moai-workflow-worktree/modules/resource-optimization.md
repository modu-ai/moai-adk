# Resource Optimization Module

Purpose: Disk space management, memory-efficient operations, performance optimization, and error handling patterns for worktree management.

Version: 1.0.0
Last Updated: 2026-01-06

---

## Quick Reference (30 seconds)

Optimization Areas:
- Disk Space: Analysis, cleanup, and compression strategies
- Memory: Streaming operations and caching patterns
- Performance: Parallel operations and efficient queries
- Error Handling: Recovery patterns and resilience

---

## Disk Space Management

### Usage Analysis

Analyze disk usage across all worktrees.

Analysis Components:
- Per-worktree size calculation (recursive file size sum)
- Identification of large worktrees exceeding thresholds
- Detection of excessive build artifacts
- Comparison against configured limits

Report Contents:
- total_size: Aggregate size of all worktrees
- worktree_sizes: Map of SPEC ID to size in bytes
- large_worktrees: List of worktrees exceeding threshold
- optimization_suggestions: Recommended cleanup actions

Size Thresholds:
- Warning threshold: 500MB per worktree
- Critical threshold: 1GB per worktree
- Configurable via max_worktree_size setting

### Optimization Strategies

Reduce disk usage through multiple strategies:

Build Artifact Removal:
- node_modules (recreatable via npm install)
- dist/build directories (recreatable via build command)
- .cache directories
- Compiled files (.pyc, .class)

Cache Cleanup:
- pytest_cache and __pycache__ directories
- .npm and .yarn cache
- IDE index files

Git Repository Optimization:
- git gc --aggressive --prune=now
- Remove orphaned objects
- Compress pack files

Worktree Compression:
- For inactive worktrees (no access in 90+ days)
- Create compressed archive
- Remove original directory
- Maintain registry entry with archived flag

### Automatic Cleanup

Configure automatic cleanup based on triggers:

Cleanup Triggers:
- Disk usage exceeds threshold percentage
- Total worktrees exceed maximum count
- Individual worktree exceeds size limit
- Scheduled cleanup interval reached

Cleanup Priority:
1. Merged worktrees (highest priority)
2. Stale worktrees (not accessed in N days)
3. Error state worktrees
4. Large worktrees with available archives

---

## Memory Efficiency

### Streaming Operations

For large registries, use streaming JSON parsing to avoid loading entire file.

Streaming Pattern:
- Use incremental JSON parser (ijson)
- Process worktrees one at a time
- Apply filters during iteration
- Yield matching worktrees

Benefits:
- Constant memory usage regardless of registry size
- Faster time-to-first-result for queries
- Reduced memory pressure on constrained systems

### Registry Caching

Implement caching to avoid repeated disk reads.

Cache Configuration:
- Cache TTL: 5 minutes default
- Cache invalidation on write operations
- Lazy loading on first access

Cache Strategy:
- Store parsed registry in memory
- Track cache timestamp
- Refresh when TTL expired or after write
- Clear cache on explicit invalidation

### Lazy Loading

Defer expensive operations until needed:
- Git status queries only when status requested
- Disk usage calculation only when size analysis needed
- Remote fetch only when sync check requested

---

## Performance Optimization

### Parallel Operations

Execute operations on multiple worktrees concurrently.

Parallel Sync:
- Identify worktrees needing sync
- Limit concurrent operations to CPU count or 4 (whichever lower)
- Execute sync operations in thread pool
- Aggregate results and report

Parallel Analysis:
- Disk usage calculation across all worktrees
- Git status queries for all worktrees
- Stale detection for all worktrees

Resource Limits:
- max_concurrent_worktrees based on system resources
- Memory monitoring during parallel operations
- Graceful degradation to sequential on resource pressure

### Query Optimization

Optimize common queries:

List Worktrees:
- Cache full list, apply filters in memory
- Index by status for fast status queries
- Sort results in memory after filter

Find by ID:
- Direct dictionary lookup (O(1))
- No iteration required

Find Stale:
- Pre-compute stale threshold date
- Compare last_accessed timestamps

### Batch Operations

Combine multiple operations for efficiency:

Batch Sync:
- Fetch all remotes once
- Apply sync to each worktree
- Single registry update at end

Batch Cleanup:
- Identify all candidates first
- Remove worktrees in parallel
- Single registry update at end

---

## Error Handling

### Error Hierarchy

Structured error classes for different failure types:

WorktreeError: Base class for all worktree errors
- message: Human-readable error description
- context: Additional debugging information

WorktreeCreationError: Errors during worktree creation
- partial_worktree: Path to incomplete worktree for cleanup
- creation_step: Step at which creation failed

SynchronizationError: Errors during sync operations
- sync_state: State at time of failure
- backup_ref: Git ref for recovery

RegistryError: Errors related to registry operations
- registry_path: Path to problematic registry
- operation: Operation that failed

### Recovery Patterns

Recover from common failure scenarios:

Creation Failure Recovery:
1. Remove partial worktree directory if exists
2. Remove registry entry if created
3. Clean up temporary files
4. Report detailed error for debugging

Sync Failure Recovery:
1. Check for backup ref in sync state
2. Reset worktree to backup ref if available
3. Mark worktree status as error
4. Set manual intervention flag
5. Log recovery actions

Registry Corruption Recovery:
1. Check for backup file
2. Validate backup file integrity
3. Restore from backup if valid
4. Scan filesystem for worktrees if no backup
5. Rebuild registry from filesystem state

### Resilience Patterns

Build resilience into operations:

Retry Logic:
- Transient failures (network, lock contention) retry 3 times
- Exponential backoff between retries
- Final failure logged and reported

Graceful Degradation:
- Operations continue despite individual failures
- Aggregate failures reported at end
- Partial success is acceptable

Consistency Checks:
- Periodic registry validation
- Filesystem consistency checks
- Automatic repair for common issues

---

## Analytics and Monitoring

### Usage Analytics

Track worktree usage patterns:

Metrics Collected:
- Worktree creation rate
- Average worktree lifetime
- Sync frequency per worktree
- Conflict rate during sync
- Disk usage trends

Reporting Intervals:
- Daily summary
- Weekly aggregate
- Monthly trends

### Performance Monitoring

Track operation performance:

Timing Metrics:
- Creation time per worktree
- Sync duration per worktree
- Cleanup duration for batch operations

Resource Metrics:
- Peak memory during operations
- Disk I/O during sync
- Network usage during fetch

### Recommendations Engine

Generate optimization recommendations:

Based on Usage Patterns:
- Suggest cleanup for low-activity worktrees
- Recommend sync frequency adjustments
- Identify unnecessary worktrees

Based on Performance:
- Suggest disk optimization for slow operations
- Recommend parallel operation adjustments
- Identify resource bottlenecks

---

Version: 1.0.0
Last Updated: 2026-01-06
Module: Resource optimization patterns for disk, memory, performance, and error handling
