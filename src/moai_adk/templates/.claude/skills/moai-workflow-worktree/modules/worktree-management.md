# Worktree Management Module

Purpose: Core architecture and management patterns for Git worktree operations including registry management, lifecycle control, and resource optimization.

Version: 1.0.0
Last Updated: 2025-11-29

---

## Quick Reference (30 seconds)

Core Components:
- Registry - Central worktree database and metadata tracking
- Manager - Core operations (create, sync, remove, cleanup)
- Models - Data structures for worktree metadata
- Lifecycle - Worktree creation, maintenance, and removal patterns

Key Patterns:
- Automatic registration and tracking
- Atomic operations with rollback
- Resource optimization and cleanup
- Conflict detection and resolution

---

## Registry Architecture

### Worktree Registry Structure

The registry is the central database that tracks all worktrees and their metadata.

Registry File: `~/.worktrees/{PROJECT_NAME}/.moai-workflow-worktree-registry.json`

Complete Registry Schema:
```json
{
 "version": "1.0.0",
 "created_at": "2025-11-29T20:00:00Z",
 "last_updated": "2025-11-29T22:00:00Z",
 "config": {
 "worktree_root": "/Users/goos/worktrees/MoAI-ADK",
 "auto_sync": true,
 "cleanup_merged": true,
 "default_base": "main",
 "sync_strategy": "merge",
 "registry_type": "local",
 "max_worktrees": 50
 },
 "worktrees": {
 "SPEC-001": {
 "id": "SPEC-001",
 "description": "User Authentication System",
 "path": "/Users/goos/worktrees/MoAI-ADK/SPEC-001",
 "branch": "feature/SPEC-001-user-auth",
 "base_branch": "main",
 "status": "active",
 "created_at": "2025-11-29T20:00:00Z",
 "last_accessed": "2025-11-29T22:00:00Z",
 "last_sync": "2025-11-29T22:30:00Z",
 "git_info": {
 "commits_ahead": 5,
 "commits_behind": 0,
 "uncommitted_changes": true,
 "branch_status": "ahead",
 "merge_conflicts": false
 },
 "metadata": {
 "template": "backend",
 "developer": "alice",
 "priority": "high",
 "estimated_size": "125MB",
 "tags": ["authentication", "security", "api"]
 },
 "operations": {
 "total_syncs": 15,
 "total_conflicts": 2,
 "last_operation": "sync",
 "last_operation_status": "success"
 }
 }
 },
 "statistics": {
 "total_worktrees": 3,
 "active_worktrees": 2,
 "merged_worktrees": 1,
 "total_disk_usage": "500MB",
 "last_cleanup": "2025-11-28T10:00:00Z"
 }
}
```

### Registry Management Patterns

Atomic Registry Operations:
```python
class RegistryManager:
 def update_worktree(self, spec_id: str, updates: dict) -> bool:
 """Atomic registry update with backup and rollback."""

 # 1. Create backup
 backup_path = self.registry_path.with_suffix('.backup.json')
 shutil.copy2(self.registry_path, backup_path)

 try:
 # 2. Load current registry
 registry = self.load_registry()

 # 3. Validate updates
 self.validate_worktree_updates(registry, spec_id, updates)

 # 4. Apply updates atomically
 registry['worktrees'][spec_id].update(updates)
 registry['last_updated'] = datetime.utcnow().isoformat()

 # 5. Write to temporary file first
 temp_path = self.registry_path.with_suffix('.tmp.json')
 self.save_registry(registry, temp_path)

 # 6. Atomic rename
 temp_path.replace(self.registry_path)

 # 7. Cleanup backup
 backup_path.unlink()

 return True

 except Exception as e:
 # Rollback on failure
 if backup_path.exists():
 backup_path.replace(self.registry_path)
 raise RegistryError(f"Failed to update registry: {e}")
```

Concurrent Access Protection:
```python
import fcntl
import time

class ConcurrentRegistryManager:
 def with_registry_lock(self, operation_func):
 """Execute operation with file-based locking."""

 lock_path = self.registry_path.with_suffix('.lock')

 with open(lock_path, 'w') as lock_file:
 # Acquire exclusive lock (with timeout)
 try:
 fcntl.flock(lock_file.fileno(), fcntl.LOCK_EX | fcntl.LOCK_NB)
 except IOError:
 # Lock held by another process, wait with timeout
 for _ in range(30): # 30 second timeout
 time.sleep(1)
 try:
 fcntl.flock(lock_file.fileno(), fcntl.LOCK_EX | fcntl.LOCK_NB)
 break
 except IOError:
 continue
 else:
 raise RegistryError("Registry lock timeout")

 try:
 # Execute operation with lock held
 return operation_func()
 finally:
 # Release lock
 fcntl.flock(lock_file.fileno(), fcntl.LOCK_UN)
 lock_path.unlink()
```

### Registry Validation and Integrity

Schema Validation:
```python
import jsonschema
from typing import Dict, Any

class RegistryValidator:
 REGISTRY_SCHEMA = {
 "type": "object",
 "required": ["version", "created_at", "last_updated", "config", "worktrees"],
 "properties": {
 "version": {"type": "string"},
 "config": {
 "type": "object",
 "required": ["worktree_root", "default_base"],
 "properties": {
 "worktree_root": {"type": "string"},
 "auto_sync": {"type": "boolean"},
 "cleanup_merged": {"type": "boolean"},
 "default_base": {"type": "string"}
 }
 },
 "worktrees": {
 "type": "object",
 "patternProperties": {
 "^SPEC-[0-9]+$": {
 "type": "object",
 "required": ["id", "path", "branch", "status", "created_at"],
 "properties": {
 "id": {"type": "string"},
 "path": {"type": "string"},
 "branch": {"type": "string"},
 "status": {"enum": ["active", "merged", "stale", "error"]},
 "created_at": {"type": "string", "format": "date-time"}
 }
 }
 }
 }
 }
 }

 def validate_registry(self, registry: Dict[str, Any]) -> bool:
 """Validate registry schema and integrity."""

 try:
 jsonschema.validate(registry, self.REGISTRY_SCHEMA)
 except jsonschema.ValidationError as e:
 raise RegistryError(f"Registry schema validation failed: {e}")

 # Additional integrity checks
 self._validate_worktree_paths(registry)
 self._validate_unique_paths(registry)
 self._validate_git_repositories(registry)

 return True

 def _validate_worktree_paths(self, registry: Dict[str, Any]) -> None:
 """Validate that all worktree paths exist and are Git repositories."""

 for spec_id, worktree in registry['worktrees'].items():
 path = Path(worktree['path'])

 if not path.exists():
 raise RegistryError(f"Worktree path does not exist: {path}")

 if not (path / '.git').exists():
 raise RegistryError(f"Worktree is not a Git repository: {path}")

 def _validate_unique_paths(self, registry: Dict[str, Any]) -> None:
 """Ensure no duplicate worktree paths."""

 paths = [wt['path'] for wt in registry['worktrees'].values()]
 duplicates = set([p for p in paths if paths.count(p) > 1])

 if duplicates:
 raise RegistryError(f"Duplicate worktree paths detected: {duplicates}")
```

---

## Manager Architecture

### WorktreeManager Core Operations

Manager Class Structure:
```python
class WorktreeManager:
 def __init__(self, repo_path: Path, worktree_root: Path):
 self.repo_path = Path(repo_path).resolve()
 self.worktree_root = Path(worktree_root).resolve()
 self.registry = RegistryManager(self.worktree_root)
 self.git_manager = GitManager(self.repo_path)

 # Core Operations
 def create_worktree(self, spec_id: str, description: str = "", options) -> WorktreeInfo
 def sync_worktree(self, spec_id: str, options) -> SyncResult
 def remove_worktree(self, spec_id: str, options) -> RemoveResult
 def switch_worktree(self, spec_id: str, options) -> SwitchResult
 def cleanup_worktrees(self, options) -> CleanupResult

 # Query Operations
 def list_worktrees(self, filters) -> List[WorktreeInfo]
 def get_worktree(self, spec_id: str) -> Optional[WorktreeInfo]
 def get_current_worktree(self) -> Optional[WorktreeInfo]
 def get_worktree_status(self, spec_id: str) -> WorktreeStatus
```

### Worktree Creation Patterns

Complete Creation Workflow:
```python
def create_worktree(self, spec_id: str, description: str = "", options) -> WorktreeInfo:
 """Create a new worktree with comprehensive validation and setup."""

 # 1. Validate input
 self._validate_spec_id(spec_id)
 self._validate_worktree_doesnt_exist(spec_id)

 # 2. Determine configuration
 branch = options.get('branch') or self._generate_branch_name(spec_id, description)
 base_branch = options.get('base') or self.config.default_base
 template = options.get('template')

 # 3. Create worktree path
 worktree_path = self.worktree_root / spec_id

 try:
 # 4. Create Git worktree
 self.git_manager.create_worktree(
 worktree_path=worktree_path,
 branch=branch,
 base_branch=base_branch,
 force=options.get('force', False)
 )

 # 5. Apply template if specified
 if template:
 self._apply_template(worktree_path, template)

 # 6. Register worktree
 worktree_info = WorktreeInfo(
 id=spec_id,
 description=description,
 path=str(worktree_path),
 branch=branch,
 base_branch=base_branch,
 status=WorktreeStatus.ACTIVE,
 created_at=datetime.utcnow(),
 metadata=options.get('metadata', {})
 )

 self.registry.register_worktree(worktree_info)

 # 7. Post-creation hooks
 self._run_creation_hooks(worktree_info)

 return worktree_info

 except Exception as e:
 # Cleanup on failure
 if worktree_path.exists():
 shutil.rmtree(worktree_path, ignore_errors=True)
 raise WorktreeCreationError(f"Failed to create worktree {spec_id}: {e}")

def _apply_template(self, worktree_path: Path, template_name: str) -> None:
 """Apply worktree template for initial setup."""
 import shlex

 template_config = self._load_template_config(template_name)

 # Execute setup commands safely (shell=False to prevent command injection CWE-78)
 for command in template_config.get('setup_commands', []):
 # Parse command string into list for safe execution
 cmd_list = shlex.split(command) if isinstance(command, str) else command
 result = subprocess.run(
 cmd_list,
 cwd=worktree_path,
 capture_output=True,
 text=True
 )

 if result.returncode != 0:
 raise TemplateError(f"Template command failed: {command}\n{result.stderr}")

 # Create template files
 for file_path, content in template_config.get('files', {}).items():
 full_path = worktree_path / file_path
 full_path.parent.mkdir(parents=True, exist_ok=True)
 full_path.write_text(content)

 # Set environment variables
 for key, value in template_config.get('env_vars', {}).items():
 env_file = worktree_path / '.env.local'
 with open(env_file, 'a') as f:
 f.write(f"{key}={value}\n")
```

### Synchronization Patterns

Smart Synchronization Strategy:
```python
def sync_worktree(self, spec_id: str, options) -> SyncResult:
 """Synchronize worktree with base branch using intelligent strategies."""

 worktree = self.get_worktree(spec_id)
 if not worktree:
 raise WorktreeNotFoundError(f"Worktree not found: {spec_id}")

 sync_strategy = options.get('strategy', self.config.sync_strategy)

 try:
 # 1. Check for uncommitted changes
 if self._has_uncommitted_changes(worktree.path) and not options.get('force'):
 raise UncommittedChangesError("Worktree has uncommitted changes")

 # 2. Fetch latest changes
 self.git_manager.fetch_updates(worktree.path)

 # 3. Analyze synchronization needs
 sync_analysis = self._analyze_sync_needs(worktree)

 if not sync_analysis.needs_sync:
 return SyncResult(skipped=True, reason="Already up to date")

 # 4. Execute synchronization based on strategy
 if sync_strategy == "merge":
 result = self._merge_sync(worktree, sync_analysis, options)
 elif sync_strategy == "rebase":
 result = self._rebase_sync(worktree, sync_analysis, options)
 elif sync_strategy == "squash":
 result = self._squash_sync(worktree, sync_analysis, options)
 else:
 raise ValueError(f"Unknown sync strategy: {sync_strategy}")

 # 5. Update registry
 self._update_sync_timestamp(spec_id)

 # 6. Post-sync hooks
 self._run_sync_hooks(worktree, result)

 return result

 except Exception as e:
 self._update_sync_timestamp(spec_id, status="error")
 raise SynchronizationError(f"Sync failed for {spec_id}: {e}")

def _merge_sync(self, worktree: WorktreeInfo, analysis: SyncAnalysis, options: dict) -> SyncResult:
 """Perform merge-based synchronization."""

 worktree_path = Path(worktree.path)

 # Check for conflicts
 if analysis.has_conflicts:
 if options.get('auto_resolve'):
 self._auto_resolve_conflicts(worktree_path, analysis.conflicts)
 elif options.get('interactive'):
 self._interactive_resolve_conflicts(worktree_path, analysis.conflicts)
 else:
 raise MergeConflictError(f"Merge conflicts detected in {worktree.id}")

 # Perform merge
 merge_result = self.git_manager.merge_branch(
 worktree_path=worktree_path,
 source_branch=worktree.base_branch,
 target_branch=worktree.branch,
 strategy=options.get('merge_strategy', 'recursive')
 )

 return SyncResult(
 strategy="merge",
 files_changed=merge_result.files_changed,
 conflicts_resolved=len(analysis.conflicts) if analysis.has_conflicts else 0,
 commits_merged=analysis.commits_behind
 )
```

### Cleanup and Optimization Patterns

Intelligent Cleanup Strategy:
```python
def cleanup_worktrees(self, options) -> CleanupResult:
 """Perform intelligent cleanup of worktrees based on various criteria."""

 cleanup_candidates = []

 # 1. Find merged worktrees
 if options.get('merged_only', False):
 merged_worktrees = self._find_merged_worktrees()
 cleanup_candidates.extend(merged_worktrees)

 # 2. Find stale worktrees
 if options.get('stale', False):
 stale_threshold = options.get('days', 30)
 stale_worktrees = self._find_stale_worktrees(stale_threshold)
 cleanup_candidates.extend(stale_worktrees)

 # 3. Find large worktrees
 if options.get('large_only', False):
 size_threshold = options.get('size_threshold', '1GB')
 large_worktrees = self._find_large_worktrees(size_threshold)
 cleanup_candidates.extend(large_worktrees)

 # 4. Remove duplicates and sort by priority
 cleanup_candidates = list(set(cleanup_candidates))
 cleanup_candidates.sort(key=lambda w: self._cleanup_priority(w), reverse=True)

 # 5. Interactive selection if requested
 if options.get('interactive'):
 cleanup_candidates = self._interactive_cleanup_selection(cleanup_candidates)

 # 6. Perform cleanup
 removed_count = 0
 total_size_freed = 0

 for worktree in cleanup_candidates:
 try:
 result = self.remove_worktree(
 worktree.id,
 backup=options.get('backup', False),
 force=options.get('force', False)
 )

 removed_count += 1
 total_size_freed += result.size_freed

 except Exception as e:
 logger.warning(f"Failed to remove worktree {worktree.id}: {e}")

 return CleanupResult(
 worktrees_removed=removed_count,
 size_freed=total_size_freed,
 candidates=len(cleanup_candidates)
 )

def _find_merged_worktrees(self) -> List[WorktreeInfo]:
 """Find worktrees whose branches have been merged to base."""

 merged_worktrees = []

 for worktree in self.list_worktrees(status='active'):
 if self.git_manager.is_branch_merged(
 branch=worktree.branch,
 into=worktree.base_branch,
 repo_path=worktree.path
 ):
 merged_worktrees.append(worktree)

 return merged_worktrees

def _find_stale_worktrees(self, days_threshold: int) -> List[WorktreeInfo]:
 """Find worktrees not accessed in specified days."""

 threshold_date = datetime.utcnow() - timedelta(days=days_threshold)
 stale_worktrees = []

 for worktree in self.list_worktrees():
 last_accessed = datetime.fromisoformat(worktree.last_accessed)
 if last_accessed < threshold_date:
 stale_worktrees.append(worktree)

 return stale_worktrees
```

---

## Resource Optimization

### Disk Space Management

Disk Usage Analysis:
```python
def analyze_disk_usage(self) -> DiskUsageReport:
 """Analyze disk usage of all worktrees."""

 total_size = 0
 worktree_sizes = {}

 for worktree in self.list_worktrees():
 size = self._calculate_worktree_size(worktree.path)
 worktree_sizes[worktree.id] = size
 total_size += size

 # Find size anomalies
 large_worktrees = [
 (wid, size) for wid, size in worktree_sizes.items()
 if size > self._get_size_threshold(wid)
 ]

 return DiskUsageReport(
 total_size=total_size,
 worktree_sizes=worktree_sizes,
 large_worktrees=large_worktrees,
 optimization_suggestions=self._generate_optimization_suggestions(worktree_sizes)
 )

def _calculate_worktree_size(self, worktree_path: str) -> int:
 """Calculate total size of worktree in bytes."""

 total_size = 0
 worktree_path = Path(worktree_path)

 for file_path in worktree_path.rglob('*'):
 if file_path.is_file():
 total_size += file_path.stat().st_size

 return total_size

def optimize_disk_usage(self, aggressive: bool = False) -> OptimizationResult:
 """Optimize disk usage across worktrees."""

 optimizations = []
 total_freed = 0

 # 1. Remove unnecessary files
 for worktree in self.list_worktrees():
 worktree_path = Path(worktree.path)

 # Remove build artifacts
 build_artifacts = self._find_build_artifacts(worktree_path)
 if build_artifacts:
 freed = self._remove_files(build_artifacts)
 optimizations.append(f"Removed {len(build_artifacts)} build artifacts from {worktree.id}")
 total_freed += freed

 # Clean Git garbage
 if aggressive:
 git_freed = self._git_gc(worktree_path)
 optimizations.append(f"Git GC freed {git_freed} from {worktree.id}")
 total_freed += git_freed

 # 2. Compress old worktrees
 if aggressive:
 old_worktrees = self._find_old_worktrees(days=90)
 for worktree in old_worktrees:
 compressed = self._compress_worktree(worktree)
 if compressed:
 optimizations.append(f"Compressed {worktree.id}")
 total_freed += compressed

 return OptimizationResult(
 optimizations=optimizations,
 total_freed=total_freed
 )
```

### Memory and Performance Optimization

Memory-Efficient Operations:
```python
class MemoryOptimizedManager:
 def __init__(self, repo_path: Path, worktree_root: Path):
 self.repo_path = repo_path
 self.worktree_root = worktree_root
 self._registry_cache = None
 self._cache_timestamp = None
 self._cache_ttl = timedelta(minutes=5)

 @property
 def registry(self) -> Dict[str, Any]:
 """Lazy-loaded registry with caching."""

 now = datetime.utcnow()

 if (self._registry_cache is None or
 self._cache_timestamp is None or
 now - self._cache_timestamp > self._cache_ttl):

 self._registry_cache = self._load_registry()
 self._cache_timestamp = now

 return self._registry_cache

 def list_worktrees_streaming(self, filters: Dict[str, Any] = None) -> Iterator[WorktreeInfo]:
 """Stream worktrees without loading entire registry."""

 registry_path = self.worktree_root / ".moai-workflow-worktree-registry.json"

 if not registry_path.exists():
 return

 # Stream JSON parsing for large registries
 import ijson

 with open(registry_path, 'rb') as f:
 worktrees = ijson.items(f, 'worktrees.item')

 for worktree_data in worktrees:
 worktree = WorktreeInfo.from_dict(worktree_data)

 if self._matches_filters(worktree, filters):
 yield worktree
```

---

## Error Handling and Recovery

### Comprehensive Error Handling

Error Hierarchy and Recovery:
```python
class WorktreeError(Exception):
 """Base class for worktree-related errors."""
 pass

class WorktreeCreationError(WorktreeError):
 """Errors during worktree creation."""
 def __init__(self, message: str, partial_worktree: Optional[Path] = None):
 super().__init__(message)
 self.partial_worktree = partial_worktree

class SynchronizationError(WorktreeError):
 """Errors during synchronization."""
 def __init__(self, message: str, sync_state: Optional[Dict] = None):
 super().__init__(message)
 self.sync_state = sync_state

class RecoveryManager:
 def recover_from_creation_failure(self, error: WorktreeCreationError) -> RecoveryResult:
 """Recover from worktree creation failure."""

 if error.partial_worktree:
 # Clean up partial worktree
 if error.partial_worktree.exists():
 shutil.rmtree(error.partial_worktree, ignore_errors=True)

 # Remove any registry entries
 self.registry.remove_worktree(error.spec_id, ignore_missing=True)

 return RecoveryResult(success=True, message="Cleanup completed")

 def recover_from_sync_failure(self, error: SynchronizationError) -> RecoveryResult:
 """Recover from synchronization failure."""

 worktree_id = error.worktree_id

 # Reset worktree to last known good state
 if error.sync_state and 'backup_ref' in error.sync_state:
 self.git_manager.reset_to_ref(
 worktree_id=worktree_id,
 ref=error.sync_state['backup_ref']
 )

 # Mark worktree as needing manual intervention
 self.registry.update_worktree(worktree_id, {
 'status': 'error',
 'last_error': str(error),
 'needs_manual_intervention': True
 })

 return RecoveryResult(
 success=True,
 message=f"Worktree {worktree_id} reset to safe state"
 )
```

---

## Monitoring and Analytics

### Worktree Analytics

Usage Analytics:
```python
class WorktreeAnalytics:
 def generate_usage_report(self, period: timedelta = timedelta(days=30)) -> UsageReport:
 """Generate comprehensive usage report."""

 end_date = datetime.utcnow()
 start_date = end_date - period

 worktrees = self.list_worktrees()

 # Calculate metrics
 metrics = {
 'total_worktrees_created': len([
 w for w in worktrees
 if datetime.fromisoformat(w.created_at) >= start_date
 ]),
 'average_sync_frequency': self._calculate_average_sync_frequency(worktrees, period),
 'most_active_worktree': self._find_most_active_worktree(worktrees, period),
 'conflict_rate': self._calculate_conflict_rate(worktrees, period),
 'storage_growth': self._calculate_storage_growth(worktrees, period)
 }

 # Generate recommendations
 recommendations = self._generate_recommendations(metrics, worktrees)

 return UsageReport(
 period=period,
 metrics=metrics,
 recommendations=recommendations,
 generated_at=end_date
 )

 def _calculate_average_sync_frequency(self, worktrees: List[WorktreeInfo], period: timedelta) -> float:
 """Calculate average sync frequency per worktree."""

 if not worktrees:
 return 0.0

 total_syncs = sum(w.operations.total_syncs for w in worktrees)
 days_in_period = period.days

 return total_syncs / (len(worktrees) * days_in_period)
```

---

Version: 1.0.0
Last Updated: 2025-11-29
Module: Core worktree management architecture with registry, lifecycle, and optimization patterns
