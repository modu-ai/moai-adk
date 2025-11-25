# Retention and Cleanup Policies

Determine what to retain and what to forget for effective memory management.

## Core Retention Strategies

### Strategy 1: Forgetting Policies

Systematically determine which memory items should be retained or discarded.

**Retention Decision Logic**:

```python
class MemoryRetentionPolicy:
    """Determine memory retention based on multiple criteria."""

    def __init__(self, session: SessionState):
        self.session = session
        self.retention_rules = self.load_retention_rules()
        self.min_access_count = 5
        self.recent_hours = 24

    def should_retain(self, item: MemoryItem) -> bool:
        """Determine if memory item should be retained.

        Args:
            item: Memory item to evaluate

        Returns:
            True if should retain, False if safe to discard

        Retention Criteria (any TRUE means retain):
            1. Critical information (always retained)
            2. Recently accessed (last 24 hours)
            3. Frequently accessed (≥5 times)
            4. Related to active tasks
            5. User-marked important
        """
        # Criterion 1: Critical information always retained
        if item.is_critical():
            return True

        # Criterion 2: Recently accessed (last 24 hours)
        if item.last_accessed > datetime.now() - timedelta(hours=self.recent_hours):
            return True

        # Criterion 3: Frequently accessed (≥ min threshold)
        if item.access_count >= self.min_access_count:
            return True

        # Criterion 4: Related to active tasks
        if item.relates_to_active_task(self.session.active_tasks):
            return True

        # Criterion 5: User-marked important
        if item.user_marked_important:
            return True

        # Default: safe to forget
        return False

    def cleanup_memory(self) -> CleanupReport:
        """Execute memory cleanup based on retention policy.

        Returns:
            Report with cleanup statistics
        """
        removed_items = []
        retained_items = []

        for item in self.session.memory_items:
            if not self.should_retain(item):
                self.session.remove_memory(item)
                removed_items.append(item)
            else:
                retained_items.append(item)

        return CleanupReport(
            removed_count=len(removed_items),
            retained_count=len(retained_items),
            tokens_freed=sum(item.token_count for item in removed_items),
            timestamp=datetime.now()
        )

class MemoryItem:
    """Memory item with retention metadata."""

    def __init__(self, content: Any, item_type: str, created_at: datetime):
        self.content = content
        self.item_type = item_type
        self.created_at = created_at
        self.last_accessed = created_at
        self.access_count = 0
        self.token_count = count_tokens(str(content))
        self.user_marked_important = False

    def is_critical(self) -> bool:
        """Check if item is critical."""
        critical_types = [
            "configuration",
            "security_credentials",
            "task_specification",
            "user_preferences",
            "error_recovery"
        ]
        return self.item_type in critical_types

    def relates_to_active_task(self, active_tasks: List[Task]) -> bool:
        """Check if item relates to any active task."""
        for task in active_tasks:
            if task.id in str(self.content) or task.relates_to(self):
                return True
        return False
```

**Critical Items Always Retained**:

| Item Type | Reason | TTL |
|-----------|--------|-----|
| Configuration parameters | System functionality | 30 days |
| Security credentials | Access control | Until rotation |
| Task specifications | Active work | Until completion |
| User preferences | UX personalization | Permanent |
| Error recovery info | Failure handling | 14 days |

**Access Count Thresholds**:

```python
retention_rules = {
    "min_access_count": 5,           # Retain if accessed 5+ times
    "recent_access_hours": 24,       # Recent = last 24 hours
    "critical_ttl_days": 30,         # Keep critical for 30 days
    "standard_ttl_days": 7,          # Standard items: 7 days
    "ephemeral_ttl_hours": 2,        # Temporary data: 2 hours
    "cache_ttl_hours": 24,           # Cached data: 24 hours
    "log_ttl_days": 3                # Log data: 3 days
}
```

---

### Strategy 2: Cleanup Automation

Automatically trigger cleanup at strategic points to maintain optimal memory.

**Cleanup Trigger Points**:

```python
def trigger_cleanup(session: SessionState) -> CleanupTriggerResult:
    """Trigger memory cleanup when thresholds met.

    Args:
        session: Current session state

    Returns:
        Result of cleanup triggers

    Trigger Conditions (OR logic):
        1. Working memory ≥ 80% capacity
        2. Cache hit rate < 65%
        3. Token budget ≥ 90% utilized
        4. Periodic cleanup interval (hourly)
        5. User-requested cleanup
    """
    cleanup_results = []

    # Trigger 1: Working memory at 80% capacity
    if session.working_memory.utilization() >= 0.80:
        result = cleanup_working_memory(session)
        cleanup_results.append(("working_memory", result))

    # Trigger 2: Cache hit rate below threshold
    if session.cache.hit_rate() < 0.65:
        result = cleanup_cache(session)
        cleanup_results.append(("cache", result))

    # Trigger 3: Token budget approaching limit
    if session.token_usage / session.token_budget >= 0.90:
        result = aggressive_cleanup(session)
        cleanup_results.append(("aggressive", result))

    # Trigger 4: Time-based cleanup (hourly)
    if should_perform_periodic_cleanup(session):
        result = periodic_cleanup(session)
        cleanup_results.append(("periodic", result))

    # Trigger 5: User-requested cleanup
    if session.user_cleanup_requested:
        result = user_requested_cleanup(session)
        cleanup_results.append(("user_requested", result))
        session.user_cleanup_requested = False

    return CleanupTriggerResult(
        triggers_activated=len(cleanup_results),
        cleanup_results=cleanup_results,
        total_tokens_freed=sum(r[1].tokens_freed for r in cleanup_results),
        timestamp=datetime.now()
    )

def should_perform_periodic_cleanup(session: SessionState) -> bool:
    """Check if periodic cleanup interval has elapsed."""
    last_cleanup = session.last_periodic_cleanup
    cleanup_interval = timedelta(hours=1)

    return datetime.now() - last_cleanup >= cleanup_interval
```

**Cleanup Operations**:

```python
def cleanup_working_memory(session: SessionState) -> CleanupResult:
    """Remove non-essential items from working memory.

    Cleanup Strategy:
        1. Apply retention policy
        2. Remove low-priority items
        3. Compress remaining items
        4. Archive removed items
    """
    policy = MemoryRetentionPolicy(session)
    removed_count = 0
    tokens_freed = 0

    # Phase 1: Apply retention policy
    for item in session.working_memory.items:
        if not policy.should_retain(item):
            tokens_freed += item.token_count
            session.working_memory.remove(item)
            removed_count += 1

    # Phase 2: Compress remaining items
    for item in session.working_memory.items:
        if item.is_compressible():
            original_size = item.token_count
            item.compress()
            tokens_freed += (original_size - item.token_count)

    return CleanupResult(
        removed_count=removed_count,
        tokens_freed=tokens_freed,
        duration_ms=(datetime.now() - start_time).total_seconds() * 1000
    )

def cleanup_cache(session: SessionState) -> CleanupResult:
    """Clean up cache to improve hit rate."""
    # Remove expired entries
    expired_removed = session.cache.remove_expired()

    # Evict LRU entries
    lru_removed = session.cache.evict_lru(count=int(session.cache.max_size * 0.1))

    # Remove low-value entries (accessed once)
    low_value_removed = session.cache.remove_low_value(min_hits=2)

    return CleanupResult(
        removed_count=expired_removed + lru_removed + low_value_removed,
        tokens_freed=calculate_cache_tokens_freed(session),
        duration_ms=measure_cleanup_duration()
    )

def aggressive_cleanup(session: SessionState) -> CleanupResult:
    """Perform aggressive cleanup when near token budget limit.

    Aggressive Strategy:
        - Remove all non-critical items
        - Compress aggressively (30% retention)
        - Archive everything possible
        - Clear all caches except hot data
    """
    # Phase 1: Remove all non-critical items
    removed_count = session.remove_all_non_critical()

    # Phase 2: Aggressive compression
    tokens_freed = session.compress_all(target_ratio=0.3)

    # Phase 3: Archive to long-term storage
    session.archive_all_archivable()

    # Phase 4: Clear cold caches
    session.cache.clear_cold_entries()

    return CleanupResult(
        removed_count=removed_count,
        tokens_freed=tokens_freed,
        aggressive=True,
        duration_ms=measure_cleanup_duration()
    )
```

---

### Strategy 3: Lifecycle Management

Manage memory through distinct session lifecycle phases.

**Session Lifecycle Phases**:

```python
class SessionLifecycleManager:
    """Manage session through distinct lifecycle phases."""

    def __init__(self, session: SessionState):
        self.session = session

    def phase_initialization(self):
        """Phase 1: Initialization (Seed with essentials)."""
        # Load critical context only
        essentials = identify_essential_files(self.session.context)
        self.session.load_files(essentials)

        # Set initial state
        self.session.phase = "initialization"
        self.session.token_budget = 200000
        self.session.memory_threshold = 0.8

    def phase_growth(self):
        """Phase 2: Growth (Load context on-demand)."""
        # Progressive loading as needed
        self.session.phase = "growth"

        # Monitor utilization
        while self.session.utilization() < 0.7:
            # Load additional context on demand
            if self.session.has_pending_loads():
                self.session.load_next_context()

    def phase_maintenance(self):
        """Phase 3: Maintenance (Periodic cleanup)."""
        self.session.phase = "maintenance"

        # Trigger periodic cleanup
        cleanup_result = periodic_cleanup(self.session)

        # Compress inactive contexts
        self.session.compress_inactive()

        # Archive old interactions
        self.session.archive_old_interactions(older_than=timedelta(hours=2))

    def phase_consolidation(self):
        """Phase 4: Consolidation (Archive at checkpoints)."""
        self.session.phase = "consolidation"

        # Create memory snapshot
        consolidated = consolidate_memory(self.session)

        # Compress and store
        compressed = compress_consolidation(consolidated)
        self.session.snapshots.append(compressed)

        # Clear non-essential memory
        self.session.clear_non_essential()

    def phase_archival(self):
        """Phase 5: Archival (Long-term storage)."""
        self.session.phase = "archival"

        # Full session archive
        archive_data = archive_session(self.session)

        # Store with metadata
        store_archive({
            "session_id": self.session.id,
            "data": archive_data,
            "archived_at": datetime.now(),
            "retention_policy": "standard",
            "token_count": archive_data.token_count
        })

        # Clear session memory
        self.session.clear_all()
```

**Archival Process**:

```python
def archive_session(session: SessionState) -> ArchivedSession:
    """Archive session to long-term storage.

    Archival Strategy:
        1. Consolidate memory
        2. Compress consolidated data
        3. Generate metadata
        4. Store with retention policy
        5. Clear active memory

    Returns:
        Archived session data
    """
    # Step 1: Consolidate memory
    consolidated = consolidate_memory(session)

    # Step 2: Compress consolidated data
    compressed = compress_consolidation(
        consolidated,
        target_ratio=0.25  # Aggressive for archival
    )

    # Step 3: Generate metadata
    metadata = {
        "session_id": session.id,
        "user_id": session.user_id,
        "start_time": session.created_at,
        "end_time": datetime.now(),
        "duration": (datetime.now() - session.created_at).total_seconds(),
        "total_interactions": len(session.history),
        "token_count": compressed.token_count,
        "compression_ratio": compressed.compression_ratio
    }

    # Step 4: Store with retention policy
    archived = ArchivedSession(
        session_id=session.id,
        data=compressed,
        metadata=metadata,
        retention_policy="standard",
        archived_at=datetime.now()
    )

    return archived

class ArchivedSession:
    """Container for archived session data."""

    def __init__(self, session_id: str, data: Any, metadata: dict,
                 retention_policy: str, archived_at: datetime):
        self.session_id = session_id
        self.data = data
        self.metadata = metadata
        self.retention_policy = retention_policy
        self.archived_at = archived_at

    def restore(self) -> SessionState:
        """Restore session from archive."""
        session = SessionState(self.session_id)

        # Decompress data
        decompressed = decompress_data(self.data)

        # Restore context
        session.load_from_consolidated(decompressed)

        # Restore metadata
        session.metadata = self.metadata

        return session

    def get_storage_size(self) -> int:
        """Calculate storage size in tokens."""
        return self.data.token_count
```

---

## Performance Optimization

### Cleanup Performance Targets

| Operation | Target | Measurement |
|-----------|--------|-------------|
| Retention policy evaluation | < 100ms | Per 1000 items |
| Working memory cleanup | < 500ms | Full cleanup |
| Cache cleanup | < 200ms | Full cleanup |
| Aggressive cleanup | < 2 seconds | Emergency cleanup |
| Session archival | < 5 seconds | Full archive |

### Memory Efficiency Metrics

| Metric | Target | Description |
|--------|--------|-------------|
| Retention rate | 20-30% | Items retained after cleanup |
| Tokens freed per cleanup | > 50K | Cleanup effectiveness |
| Cleanup frequency | < 1/hour | Under normal load |
| Archive compression ratio | > 75% | Archival efficiency |

---

## Best Practices

**DO**:
- ✅ Review retention policies quarterly
- ✅ Monitor cleanup effectiveness continuously
- ✅ Archive inactive sessions after 7 days
- ✅ Maintain 60-70% cache hit rate as target
- ✅ Trigger cleanup proactively before limits
- ✅ Use lifecycle phases for systematic management

**DON'T**:
- ❌ Delete critical items during cleanup
- ❌ Skip archival for long-running sessions
- ❌ Ignore cleanup trigger thresholds
- ❌ Use fixed retention rules for all content types
- ❌ Archive without compression
- ❌ Clean up recently accessed items

---

**Version**: 3.0.0
**Last Updated**: 2025-11-23
**Status**: Production Ready
