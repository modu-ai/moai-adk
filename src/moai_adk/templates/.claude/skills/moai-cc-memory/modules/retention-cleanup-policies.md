# Retention and Cleanup Policies

Determine what to retain and what to forget for effective memory management.

## Forgetting Policies

### Retention Decision Logic

```python
class MemoryRetentionPolicy:
    def should_retain(self, item: MemoryItem) -> bool:
        '''Determine if item should be retained.'''
        
        # Critical information always retained
        if item.is_critical():
            return True
        
        # Recently accessed (last 24 hours)
        if item.last_accessed > datetime.now() - timedelta(hours=24):
            return True
        
        # Frequently accessed (> min threshold)
        if item.access_count > self.min_access_count:
            return True
        
        # Related to active tasks
        if item.relates_to_active_task(self.active_tasks):
            return True
        
        # Everything else: safe to forget
        return False
```

### Critical Items

Items always retained:

- Configuration parameters
- Security credentials
- Task specifications
- User preferences
- Error recovery information

### Access Count Thresholds

```python
retention_rules = {
    "min_access_count": 5,           # Retain if accessed 5+ times
    "recent_access_hours": 24,       # Recent = last 24 hours
    "critical_ttl_days": 30,         # Keep critical for 30 days
    "standard_ttl_days": 7           # Standard items: 7 days
}
```

---

## Cleanup Automation

### Cleanup Trigger Points

```python
def trigger_cleanup(session: SessionState):
    '''Trigger memory cleanup when thresholds met.'''
    
    # Trigger 1: Working memory at 80% capacity
    if session.working_memory.utilization() > 0.80:
        cleanup_working_memory(session)
    
    # Trigger 2: Cache hit rate below threshold
    if session.cache.hit_rate() < 0.65:
        cleanup_cache(session)
    
    # Trigger 3: Token budget approaching limit
    if session.token_usage / session.token_budget > 0.90:
        aggressive_cleanup(session)
    
    # Trigger 4: Time-based cleanup (hourly)
    if should_perform_periodic_cleanup(session):
        periodic_cleanup(session)
```

### Cleanup Operations

```python
def cleanup_working_memory(session):
    '''Remove non-essential items from working memory.'''
    
    policy = MemoryRetentionPolicy(session)
    removed = 0
    
    for item in session.working_memory.items:
        if not policy.should_retain(item):
            session.working_memory.remove(item)
            removed += 1
    
    return removed
```

---

## Lifecycle Management

### Session Lifecycle Phases

1. **Initialization**: Seed with essentials
2. **Growth**: Load context on-demand
3. **Maintenance**: Periodic cleanup
4. **Consolidation**: Archive at checkpoints
5. **Archival**: Long-term storage

### Archival Process

When session ends:

```python
def archive_session(session: SessionState):
    '''Archive session to long-term storage.'''
    
    # Consolidate memory
    consolidated = consolidate_memory(session)
    
    # Compress
    compressed = compress_consolidation(consolidated)
    
    # Store with metadata
    store_archive({
        "session_id": session.id,
        "data": compressed,
        "archived_at": datetime.now(),
        "retention_policy": "standard"
    })
```

---

## Best Practices

- Review retention policies quarterly
- Monitor cleanup effectiveness
- Archive inactive sessions after 7 days
- Maintain 60-70% cache hit rate as target
