# Session Memory Architecture

Memory organized across three layers with distinct characteristics, lifetimes, and responsibilities.

## Three-Layer Model Overview

**Working Memory (Active Context)**
- Size: ~50K tokens
- Lifetime: Current session only
- Purpose: Active computation and state
- Contents:
  - Current session state
  - Recently accessed files
  - Active agent contexts
  - Decision contexts for ongoing operations

**Long-Term Memory (Persistent Storage)**
- Size: Unlimited (compressed)
- Lifetime: Persistent across sessions
- Purpose: Knowledge preservation
- Contents:
  - Skill knowledge capsules
  - Agent configurations
  - Historical project context
  - Learned patterns and best practices

**Context Cache (Intelligent Buffer)**
- Size: ~20K tokens
- Lifetime: Session + TTL
- Purpose: Fast retrieval of frequently needed data
- Contents:
  - Frequently accessed patterns
  - Compiled knowledge
  - Precomputed embeddings
  - Cache entries with TTL

---

## Session State Management

### Session Initialization

```python
class SessionState:
    def __init__(self, session_id: str):
        self.id = session_id
        self.created_at = datetime.now()
        self.working_memory = WorkingMemory(capacity=50000)
        self.cache = IntelligentCache()
        self.history = []
        self.active_agents = []
        self.token_budget = 200000
```

### Memory Boundaries

Working memory vs. long-term memory transition happens automatically:

1. **Automatic**: When working memory reaches 80% capacity
2. **Manual**: Explicit consolidation call
3. **Periodic**: On session checkpoints

### State Snapshot for Persistence

```python
def create_memory_snapshot(session):
    return {
        "session_id": session.id,
        "timestamp": datetime.now(),
        "working_memory_summary": compress(session.working_memory),
        "active_contexts": session.get_active_contexts(),
        "knowledge_patterns": extract_patterns(session.history)
    }
```

---

## Memory Layer Interactions

Layers interact through controlled interfaces:

```
User Request
    ↓
Working Memory (check cache first)
    ↓
If miss: Load from Long-Term Memory
    ↓
Update Context Cache for future access
    ↓
Return to User
```

---

## Best Practices

- Minimize working memory growth (compress regularly)
- Use cache TTL effectively (default 24 hours)
- Archive long-term memory periodically
- Monitor layer utilization metrics
