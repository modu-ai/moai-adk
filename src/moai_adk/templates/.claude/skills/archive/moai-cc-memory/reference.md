# Claude Code Memory - Reference

## Memory Layer Architecture

```
┌─────────────────────────────────────────┐
│   Application Layer                     │
│   (Session, State Management)           │
├─────────────────────────────────────────┤
│   Memory API Layer                      │
│   (MemoryStore, MemoryCache)            │
├─────────────────────────────────────────┤
│   Backend Storage                       │
│   (Redis, Memory, Disk)                 │
└─────────────────────────────────────────┘
```

## API Reference

### MemoryStore Interface

| Method | Description | Return Value |
|--------|-------------|--------------|
| `get(key)` | Retrieve value | `T \| None` |
| `set(key, value, ttl)` | Store value | `None` |
| `delete(key)` | Delete value | `bool` |
| `exists(key)` | Check existence | `bool` |

### TokenBudgetTracker

```python
tracker = TokenBudgetTracker(max_tokens=200000)
remaining = await tracker.get_remaining_budget()
await tracker.record_usage(request_id, tokens)
```

## Memory Key Convention

```python
# Key format: scope:user_id:component:detail
MemoryKey(
    scope="session",     # Scope
    user_id="user123",   # User
    component="context", # Component
    detail="state"       # Detail
)
```

## Use Cases

| Use Case | Memory Type | TTL |
|----------|-------------|-----|
| Session data | MemoryStore | 1-24 hours |
| Token tracking | TokenBudgetTracker | Session duration |
| Cache | MemoryCache | 5 min-1 hour |
| State | StatefulMemory | Workflow duration |

---

**Last Updated**: 2025-11-22
