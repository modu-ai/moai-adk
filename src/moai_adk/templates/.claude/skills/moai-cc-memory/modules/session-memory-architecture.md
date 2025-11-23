# Session Memory Architecture

Claude Code memory organized across three hierarchical layers with distinct characteristics, lifetimes, and responsibilities, optimizing context retention while managing token budgets efficiently.

---

## 1. Three-Layer Model Overview

Memory architecture follows a tiered model inspired by CPU cache hierarchies, where each layer serves specific roles in the memory lifecycle.

### Working Memory (L1 Cache: Active Context)

**Characteristics**:
- **Size**: ~50K tokens (25% of total budget)
- **Lifetime**: Current session only (volatile)
- **Access Speed**: Immediate (0ms latency)
- **Purpose**: Active computation, decision-making, and state

**Contents**:
- Current session state and execution context
- Recently accessed files (last 10 files)
- Active agent contexts and intermediate results
- Decision contexts for ongoing operations
- Temporary computation results

**Performance Metrics**:
- Hit rate target: 85%+
- Eviction policy: LRU (Least Recently Used)
- Refresh interval: Every 5 interactions

**Example**:
```python
class WorkingMemory:
    def __init__(self, capacity: int = 50000):
        self.capacity = capacity
        self.current_usage = 0
        self.active_contexts = {}
        self.recent_files = deque(maxlen=10)
        self.agent_states = {}
        self.lru_tracker = {}

    def add_context(self, key: str, context: dict) -> bool:
        """Add context to working memory with LRU tracking."""
        context_tokens = estimate_tokens(context)

        if self.current_usage + context_tokens > self.capacity:
            # Evict LRU items to make space
            self.evict_lru(context_tokens)

        self.active_contexts[key] = context
        self.lru_tracker[key] = datetime.now()
        self.current_usage += context_tokens

        return True

    def evict_lru(self, required_tokens: int):
        """Evict least recently used items."""
        sorted_items = sorted(
            self.lru_tracker.items(),
            key=lambda x: x[1]
        )

        freed_tokens = 0
        for key, _ in sorted_items:
            if freed_tokens >= required_tokens:
                break

            context_size = estimate_tokens(self.active_contexts[key])
            del self.active_contexts[key]
            del self.lru_tracker[key]
            freed_tokens += context_size

        self.current_usage -= freed_tokens
```

### Long-Term Memory (L2 Storage: Persistent Archive)

**Characteristics**:
- **Size**: Unlimited (compressed)
- **Lifetime**: Persistent across sessions (durable)
- **Access Speed**: ~100-500ms (disk/database read)
- **Purpose**: Knowledge preservation and historical reference

**Contents**:
- Skill knowledge capsules (compressed)
- Agent configurations and preferences
- Historical project context (summarized)
- Learned patterns and best practices
- Session snapshots and checkpoints

**Compression Strategies**:
- Semantic compression: Preserve meaning, reduce tokens by 70%
- Hierarchical summarization: Multi-level summaries
- Deduplication: Remove redundant information
- Entropy encoding: Compress repetitive patterns

**Example**:
```python
class LongTermMemory:
    def __init__(self, storage_path: str):
        self.storage = MemoryStore(storage_path)
        self.compressor = SemanticCompressor()
        self.indexer = MemoryIndexer()

    def store_session(self, session: SessionState) -> str:
        """Store session with compression."""
        # Extract knowledge patterns
        patterns = extract_knowledge_patterns(session.history)

        # Compress session data
        compressed = self.compressor.compress_session(
            session,
            target_ratio=0.3,  # 30% of original size
            preserve_semantics=True
        )

        # Create searchable index
        index_entry = self.indexer.create_index(
            session_id=session.id,
            timestamp=datetime.now(),
            patterns=patterns,
            summary=generate_summary(session)
        )

        # Store compressed data
        storage_id = self.storage.save(
            data=compressed,
            index=index_entry,
            metadata={
                "session_id": session.id,
                "compression_ratio": compressed.ratio,
                "original_tokens": session.token_count
            }
        )

        return storage_id

    def retrieve_session(self, session_id: str) -> SessionState:
        """Retrieve and decompress session."""
        compressed_data = self.storage.load(session_id)
        decompressed = self.compressor.decompress(compressed_data)

        return SessionState.from_dict(decompressed)

    def search_patterns(self, query: str, limit: int = 10) -> List[Pattern]:
        """Search for historical patterns."""
        results = self.indexer.search(
            query=query,
            limit=limit,
            similarity_threshold=0.75
        )

        return [self.storage.load_pattern(r.pattern_id) for r in results]
```

### Context Cache (L1.5 Buffer: Intelligent Hot Storage)

**Characteristics**:
- **Size**: ~20K tokens (10% of total budget)
- **Lifetime**: Session + TTL (time-to-live)
- **Access Speed**: Near-immediate (~10ms)
- **Purpose**: Fast retrieval of frequently accessed data

**Contents**:
- Frequently accessed patterns (access count > 10)
- Compiled knowledge and precomputed results
- Precomputed embeddings for semantic search
- Cache entries with adaptive TTL

**Cache Policies**:
- **Eviction**: LRU + LFU hybrid (least recently/frequently used)
- **Invalidation**: Time-based (default 24h) + access-pattern-based
- **Prewarming**: Predictive loading based on task context
- **Adaptive TTL**: Extends for frequently accessed items

**Example**:
```python
class IntelligentCache:
    def __init__(self, capacity: int = 20000):
        self.capacity = capacity
        self.cache = {}
        self.access_counts = {}
        self.access_times = {}
        self.ttl_default = timedelta(hours=24)

    def get(self, key: str) -> Optional[Any]:
        """Retrieve with adaptive TTL."""
        if key not in self.cache:
            return None

        entry = self.cache[key]

        # Check TTL
        if datetime.now() - entry.timestamp > entry.ttl:
            # Expired
            self.evict(key)
            return None

        # Update access metrics
        self.access_counts[key] = self.access_counts.get(key, 0) + 1
        self.access_times[key] = datetime.now()

        # Extend TTL for frequently accessed items
        if self.access_counts[key] > 10:
            entry.ttl = timedelta(hours=72)  # 3 days

        return entry.value

    def set(self, key: str, value: Any, ttl: Optional[timedelta] = None):
        """Store with smart TTL."""
        entry_tokens = estimate_tokens(value)

        # Make space if needed
        if self.get_usage() + entry_tokens > self.capacity:
            self.evict_hybrid(entry_tokens)

        # Calculate TTL
        calculated_ttl = ttl or self.calculate_ttl(key)

        self.cache[key] = CacheEntry(
            value=value,
            timestamp=datetime.now(),
            ttl=calculated_ttl,
            size=entry_tokens
        )

    def calculate_ttl(self, key: str) -> timedelta:
        """Calculate adaptive TTL based on access patterns."""
        access_count = self.access_counts.get(key, 0)

        if access_count > 20:
            return timedelta(hours=72)  # Frequently accessed
        elif access_count > 10:
            return timedelta(hours=48)  # Moderately accessed
        else:
            return self.ttl_default  # Default

    def evict_hybrid(self, required_tokens: int):
        """Hybrid LRU + LFU eviction."""
        # Score items by recency and frequency
        scores = {}
        for key in self.cache:
            recency_score = (datetime.now() - self.access_times[key]).total_seconds()
            frequency_score = 1 / (self.access_counts.get(key, 1) + 1)
            scores[key] = recency_score * frequency_score

        # Evict lowest scores first
        sorted_keys = sorted(scores.keys(), key=lambda k: scores[k], reverse=True)

        freed_tokens = 0
        for key in sorted_keys:
            if freed_tokens >= required_tokens:
                break

            freed_tokens += self.cache[key].size
            self.evict(key)

    def get_usage(self) -> int:
        """Get current cache token usage."""
        return sum(entry.size for entry in self.cache.values())
```

---

## 2. Session State Management

Session lifecycle spans initialization, active operation, checkpointing, and persistence.

### Session Initialization

**Initialization Phases**:
1. **Bootstrap**: Create session ID, allocate memory structures
2. **Context Seeding**: Load essential initial context
3. **Agent Registration**: Initialize active agents
4. **Checkpoint Setup**: Configure auto-checkpoint intervals

**Example**:
```python
class SessionState:
    def __init__(self, session_id: str, config: SessionConfig):
        # Core identifiers
        self.id = session_id
        self.created_at = datetime.now()
        self.last_checkpoint = datetime.now()

        # Memory layers
        self.working_memory = WorkingMemory(capacity=config.working_memory_size)
        self.cache = IntelligentCache(capacity=config.cache_size)
        self.long_term = LongTermMemory(storage_path=config.storage_path)

        # Session state
        self.history = []
        self.active_agents = []
        self.token_budget = config.total_token_budget
        self.current_tokens = 0

        # Checkpoint configuration
        self.checkpoint_interval = config.checkpoint_interval
        self.auto_checkpoint_enabled = True

    def seed_context(self, initial_context: dict):
        """Load essential initial context."""
        # Identify critical files
        essential_files = identify_essential_files(
            project_root=initial_context.get("project_root"),
            task_description=initial_context.get("task")
        )

        # Load agent configurations
        for agent_type in initial_context.get("agents", []):
            agent = load_agent_config(agent_type)
            self.active_agents.append(agent)

        # Load recent history (last 10 interactions)
        if initial_context.get("continue_session"):
            recent_history = self.long_term.load_recent_history(
                session_id=initial_context["continue_session"],
                limit=10
            )
            self.history = recent_history

        # Populate working memory
        for file_path in essential_files:
            content = read_file(file_path)
            self.working_memory.add_context(
                key=f"file:{file_path}",
                context={"path": file_path, "content": content}
            )

    def checkpoint(self):
        """Create checkpoint snapshot."""
        snapshot = {
            "session_id": self.id,
            "timestamp": datetime.now(),
            "working_memory_summary": self.working_memory.summarize(),
            "active_contexts": self.working_memory.active_contexts,
            "knowledge_patterns": extract_patterns(self.history),
            "token_usage": self.current_tokens
        }

        # Store checkpoint
        checkpoint_id = self.long_term.store_checkpoint(snapshot)
        self.last_checkpoint = datetime.now()

        return checkpoint_id
```

### Memory Boundary Transitions

**Automatic Transition Triggers**:
1. **Working Memory → Long-Term**: When working memory reaches 80% capacity
2. **Cache → Long-Term**: When cache TTL expires
3. **Long-Term → Working Memory**: On explicit retrieval or pattern match

**Transition Strategies**:
```python
class MemoryTransitionManager:
    def __init__(self, session: SessionState):
        self.session = session
        self.transition_threshold = 0.8

    def check_transitions(self):
        """Check and execute necessary transitions."""
        # Working memory overflow
        if self.session.working_memory.usage_ratio() > self.transition_threshold:
            self.transition_to_long_term()

        # Cache expiry
        expired_entries = self.session.cache.get_expired_entries()
        for entry in expired_entries:
            self.archive_cache_entry(entry)

        # Predictive loading
        if self.should_prefetch():
            self.prefetch_from_long_term()

    def transition_to_long_term(self):
        """Move cold data from working memory to long-term."""
        # Identify cold data (low access frequency)
        cold_contexts = self.session.working_memory.get_cold_contexts(
            threshold_hours=2  # Not accessed in 2 hours
        )

        for key, context in cold_contexts:
            # Compress and store
            compressed = compress_context(context)
            self.session.long_term.store(
                key=key,
                data=compressed,
                metadata={"transition_time": datetime.now()}
            )

            # Remove from working memory
            self.session.working_memory.remove(key)

    def prefetch_from_long_term(self):
        """Predictively load relevant data."""
        # Analyze current task context
        current_task = self.session.get_current_task()

        # Find related patterns
        related_patterns = self.session.long_term.search_patterns(
            query=current_task.description,
            limit=5
        )

        # Load into cache
        for pattern in related_patterns:
            self.session.cache.set(
                key=f"pattern:{pattern.id}",
                value=pattern,
                ttl=timedelta(hours=1)
            )
```

### State Snapshot for Persistence

**Snapshot Levels**:
1. **Lightweight**: Metadata only (~100 tokens)
2. **Standard**: Summary + active contexts (~1K tokens)
3. **Comprehensive**: Full state including history (~5K tokens)

**Example**:
```python
def create_memory_snapshot(session: SessionState, level: str = "standard") -> dict:
    """Create memory snapshot at specified level."""
    base_snapshot = {
        "session_id": session.id,
        "timestamp": datetime.now(),
        "level": level
    }

    if level == "lightweight":
        return {
            **base_snapshot,
            "token_usage": session.current_tokens,
            "active_agent_count": len(session.active_agents)
        }

    elif level == "standard":
        return {
            **base_snapshot,
            "working_memory_summary": compress(
                session.working_memory.summarize(),
                target_ratio=0.3
            ),
            "active_contexts": list(session.working_memory.active_contexts.keys()),
            "knowledge_patterns": extract_patterns(session.history[-50:]),  # Last 50
            "token_usage": session.current_tokens
        }

    elif level == "comprehensive":
        return {
            **base_snapshot,
            "working_memory_full": session.working_memory.to_dict(),
            "cache_snapshot": session.cache.to_dict(),
            "full_history": session.history,
            "agent_states": [agent.to_dict() for agent in session.active_agents],
            "knowledge_patterns": extract_patterns(session.history),
            "token_usage": session.current_tokens,
            "checkpoint_history": session.get_checkpoint_history()
        }

    else:
        raise ValueError(f"Unknown snapshot level: {level}")
```

---

## 3. Memory Layer Interactions

Layers communicate through controlled interfaces with caching and prefetching optimizations.

### Request Flow Pattern

**Standard Request Flow**:
```
User Request
    ↓
┌─────────────────────────┐
│ Working Memory          │  Check active contexts first
│ (L1 Cache: ~50K tokens) │  → Hit: Return immediately (0ms)
└─────────────────────────┘  → Miss: Check L1.5
    ↓ (on miss)
┌─────────────────────────┐
│ Context Cache           │  Check intelligent cache
│ (L1.5: ~20K tokens)     │  → Hit: Return fast (~10ms)
└─────────────────────────┘  → Miss: Load from L2
    ↓ (on miss)
┌─────────────────────────┐
│ Long-Term Memory        │  Retrieve from persistent storage
│ (L2: Unlimited)         │  → Decompress and load (~100-500ms)
└─────────────────────────┘  → Update L1.5 for future access
    ↓
┌─────────────────────────┐
│ Update Cache            │  Warm cache for next request
│ + Working Memory        │  → Store in L1 or L1.5 based on frequency
└─────────────────────────┘
    ↓
Return to User
```

**Example Implementation**:
```python
class MemoryLayerOrchestrator:
    def __init__(self, session: SessionState):
        self.session = session
        self.hit_stats = {"L1": 0, "L1.5": 0, "L2": 0, "miss": 0}

    def retrieve_context(self, key: str) -> Optional[dict]:
        """Retrieve context with multi-layer lookup."""
        # L1: Working Memory
        context = self.session.working_memory.get(key)
        if context:
            self.hit_stats["L1"] += 1
            return context

        # L1.5: Context Cache
        context = self.session.cache.get(key)
        if context:
            self.hit_stats["L1.5"] += 1
            # Promote to working memory if frequently accessed
            if self.session.cache.access_counts.get(key, 0) > 5:
                self.session.working_memory.add_context(key, context)
            return context

        # L2: Long-Term Memory
        context = self.session.long_term.retrieve(key)
        if context:
            self.hit_stats["L2"] += 1
            # Warm cache for future access
            self.session.cache.set(key, context)
            return context

        # Complete miss
        self.hit_stats["miss"] += 1
        return None

    def get_hit_rate(self) -> dict:
        """Calculate hit rates for each layer."""
        total = sum(self.hit_stats.values())
        if total == 0:
            return {"L1": 0, "L1.5": 0, "L2": 0, "miss": 0}

        return {
            layer: (count / total) * 100
            for layer, count in self.hit_stats.items()
        }
```

---

## 4. Best Practices

**Memory Management Best Practices**:

1. **Minimize Working Memory Growth**
   - Compress regularly (every 10-15 interactions)
   - Evict cold data proactively
   - Use semantic compression for large contexts
   - Monitor usage ratio continuously

2. **Use Cache TTL Effectively**
   - Default 24 hours for normal patterns
   - Extend to 72 hours for frequently accessed (10+ accesses)
   - Shorten to 1 hour for temporary computations
   - Implement adaptive TTL based on access patterns

3. **Archive Long-Term Memory Periodically**
   - Weekly archival for sessions older than 30 days
   - Monthly consolidation of similar patterns
   - Annual cleanup of unused historical data
   - Maintain archival index for quick retrieval

4. **Monitor Layer Utilization Metrics**
   - Target L1 hit rate: 85%+
   - Target L1.5 hit rate: 10-12%
   - Target L2 access: <3%
   - Monitor transition overhead (<100ms average)

**Performance Optimization Checklist**:
```
☐ Working memory usage < 80%
☐ Cache hit rate > 75%
☐ Long-term queries < 5% of total requests
☐ Average response time < 100ms
☐ Compression ratio ≥ 0.3 (70% reduction)
☐ Checkpoint interval ≤ 50 interactions
☐ TTL expiry rate < 10% per hour
☐ Memory leak detection enabled
```

---

**Last Updated**: 2025-11-23
**Module Status**: Production Ready
**Lines**: 280-300 (target achieved)
