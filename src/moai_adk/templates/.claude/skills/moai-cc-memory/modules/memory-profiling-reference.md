# Memory Profiling and Reference

Monitor memory usage and profile session performance.

## Memory Profiling

### Profile Session Memory Usage

```python
def profile_memory_usage(session: SessionState) -> dict:
    '''Profile memory consumption across all layers.'''
    
    profile = {
        "working_memory": measure_tokens(session.active_contexts),
        "long_term_memory": measure_tokens(session.archived_contexts),
        "cache": measure_tokens(session.cache.items),
        "agent_states": sum(measure_tokens(a.state) for a in session.agents),
        "conversation_history": measure_tokens(session.history)
    }
    
    profile["total_tokens"] = sum(profile.values())
    profile["utilization_rate"] = profile["total_tokens"] / session.token_budget
    
    return profile
```

### Monitoring Points

Track these metrics continuously:

- **Working Memory Utilization**: 0-100% (alert > 80%)
- **Cache Hit Rate**: 0-100% (target 70-80%)
- **Token Budget Usage**: 0-100% (alert > 90%)
- **Compression Effectiveness**: Ratio of compressed/original
- **Cleanup Frequency**: Times triggered per session

---

## Best Practices

### Session Management

1. **Initialize**: Seed with essentials only
2. **Monitor**: Track utilization continuously
3. **Optimize**: Compress and cache proactively
4. **Cleanup**: Trigger before hitting limits
5. **Archive**: Save long-term after session

### Memory Strategies

- Keep working memory under 70% capacity
- Maintain 70-80% cache hit rate
- Compress automatically at 80% utilization
- Archive sessions after 24 hours inactive
- Review retention policies quarterly

### Token Budget Best Practices

- Monitor component allocations
- Adjust based on usage patterns
- Trigger cleanup before exhaustion
- Document budget changes
- Archive metrics for analysis

---

## Troubleshooting

### High Working Memory Usage

**Symptoms**: Working memory > 80%

**Solutions**:
1. Compress inactive contexts
2. Archive completed tasks
3. Increase consolidation frequency
4. Review loaded modules (remove unused)

### Low Cache Hit Rate

**Symptoms**: Cache hit rate < 60%

**Solutions**:
1. Increase cache TTL for frequent items
2. Preload expected items on demand
3. Profile access patterns
4. Adjust cache eviction policy

### Token Budget Exhaustion

**Symptoms**: Total usage > 90% of budget

**Solutions**:
1. Trigger component cleanup
2. Compress old interactions
3. Reduce history retention
4. Archive non-critical knowledge

---

## Integration Examples

### With Agent Workflows

```python
# Before agent execution
memory_profile = profile_memory_usage(session)
if memory_profile["utilization_rate"] > 0.85:
    trigger_cleanup(session)

# Execute agent
result = execute_agent(session)

# After completion
consolidate_memory(session)
update_cache(session)
```

### With Long-Running Tasks

```python
# Checkpoint-based memory management
for checkpoint in task.checkpoints:
    # Save state
    snapshot = create_memory_snapshot(session)
    
    # Cleanup
    cleanup_expired_items(session)
    compress_memory(session)
    
    # Continue
    continue_task()
```

---

## References

- Claude Code Context Management
- Token Optimization Strategy
- Memory Architecture Patterns
- Cache Invalidation Best Practices
