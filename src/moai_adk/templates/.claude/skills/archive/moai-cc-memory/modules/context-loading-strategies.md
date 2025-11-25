# Context Loading Strategies

Load context strategically to optimize startup time and minimize memory usage.

## Core Loading Strategies

### Strategy 1: Context Seeding

Initialize sessions with critical context only, deferring non-essential data.

**Essential Context Identification**:

```python
def identify_essential_files(context: dict) -> List[str]:
    """Identify which files are essential for session start.

    Args:
        context: Initial session context containing project info

    Returns:
        List of essential file paths to load immediately

    Priority Classification:
        - P0 (Critical): Main entry point, config files
        - P1 (High): Current task files, active modules
        - P2 (Medium): Recent history, common utilities
        - P3 (Low): Documentation, tests (load on demand)
    """
    essential = []

    # P0: Always essential (critical for startup)
    essential.extend([
        context.get("main_entry_point"),      # e.g., src/main.py
        context.get("config_file"),           # e.g., config.json
        context.get("documentation_index")    # e.g., README.md
    ])

    # P1: Project-specific essential
    if context.get("project_type") == "monorepo":
        essential.append(context.get("workspace_config"))

    if context.get("project_type") == "microservices":
        essential.extend([
            context.get("service_registry"),
            context.get("api_gateway_config")
        ])

    # P1: Task-specific essential
    if context.get("current_task"):
        essential.extend(
            identify_task_related_files(context["current_task"])
        )

    # P2: Agent-specific essential
    if context.get("active_agents"):
        for agent in context["active_agents"]:
            essential.extend(agent.get_required_files())

    # Filter None values and duplicates
    return list(set([f for f in essential if f]))

def identify_task_related_files(task: dict) -> List[str]:
    """Identify files related to current task.

    Args:
        task: Task definition with type and scope

    Returns:
        List of task-relevant file paths
    """
    files = []

    # Task type determines file scope
    task_type = task.get("type")

    if task_type == "bug_fix":
        files.extend([
            task.get("buggy_file"),
            task.get("test_file"),
            task.get("error_log")
        ])

    elif task_type == "feature_implementation":
        files.extend([
            task.get("spec_file"),
            task.get("target_module"),
            task.get("related_tests")
        ])

    elif task_type == "refactoring":
        files.extend([
            task.get("target_files"),
            task.get("dependency_files"),
            task.get("test_suite")
        ])

    return [f for f in files if f]
```

**Session Seeding Pattern**:

```python
def seed_context(session_id: str, initial_context: dict) -> SessionState:
    """Initialize session with strategic context.

    Args:
        session_id: Unique session identifier
        initial_context: Critical context for session start

    Returns:
        Initialized session state with seeded context

    Performance Target:
        - Seeding time: < 2 seconds
        - Memory footprint: < 50K tokens
        - Ready-to-use: Immediate
    """
    # Phase 1: Essential Files (P0)
    essential_files = identify_essential_files(initial_context)
    session = SessionState(session_id)

    # Load in priority order
    session.load_files(essential_files, priority="high")

    # Phase 2: Agent Configurations (P1)
    agent_configs = load_active_agent_configs()
    session.set_agent_configs(agent_configs)

    # Phase 3: Recent History (P2)
    recent_history = load_recent_history(limit=10)  # Last 10 interactions
    session.set_history(recent_history)

    # Phase 4: Compress and Cache
    session.compress()
    cache_session(session)

    # Phase 5: Log seeding metrics
    log_seeding_metrics(session, {
        "files_loaded": len(essential_files),
        "tokens_used": session.token_count,
        "duration_ms": session.seeding_duration
    })

    return session

class SessionState:
    """Session state with lazy loading capabilities."""

    def __init__(self, session_id: str):
        self.id = session_id
        self.files = {}
        self.agent_configs = {}
        self.history = []
        self.token_count = 0
        self.seeding_duration = 0
        self.created_at = datetime.now()

    def load_files(self, file_paths: List[str], priority: str = "high"):
        """Load files with priority handling."""
        start = time.time()

        for path in file_paths:
            content = read_file(path)
            self.files[path] = {
                "content": content,
                "priority": priority,
                "loaded_at": datetime.now(),
                "token_count": count_tokens(content)
            }
            self.token_count += self.files[path]["token_count"]

        self.seeding_duration = (time.time() - start) * 1000  # ms

    def compress(self):
        """Compress session data for efficient storage."""
        # Compress older interactions
        if len(self.history) > 10:
            self.history = (
                self.history[-10:] +  # Keep last 10 in full
                [summarize(h) for h in self.history[:-10]]  # Summarize older
            )

    def get_active_contexts(self) -> List[str]:
        """Get list of active context keys."""
        return [k for k, v in self.files.items() if v["priority"] in ["high", "medium"]]
```

---

## Strategy 2: Progressive Loading

Load context incrementally as needed, not all at startup.

**On-Demand Loading Implementation**:

```python
class ProgressiveContextLoader:
    """Progressive context loader with caching and priority."""

    def __init__(self, session: SessionState):
        self.session = session
        self.loaded_modules = set()
        self.cache = {}
        self.load_stats = {
            "hits": 0,
            "misses": 0,
            "avg_load_time_ms": 0
        }

    def load_on_demand(self, module_name: str) -> ModuleContext:
        """Load module context only when explicitly requested.

        Args:
            module_name: Module to load

        Returns:
            Module context with essential data

        Performance:
            - Cache hit: < 5ms
            - Cache miss: < 500ms
            - Memory efficient: Load minimal initially
        """
        start = time.time()

        # Check cache first
        if module_name in self.loaded_modules:
            self.load_stats["hits"] += 1
            return self.get_cached_module(module_name)

        # Cache miss: Load incrementally
        self.load_stats["misses"] += 1
        module = self.load_module_minimal(module_name)

        # Cache for future access
        self.cache_module(module_name, module)
        self.loaded_modules.add(module_name)

        # Update load time stats
        load_time = (time.time() - start) * 1000
        self._update_load_stats(load_time)

        return module

    def load_module_minimal(self, name: str) -> ModuleContext:
        """Load only essential module information.

        Minimal Context Includes:
            - Module name and path
            - Exported functions/classes (signatures only)
            - Direct dependencies (no transitive)
            - Auto-generated summary

        Full Context NOT Loaded:
            - Full source code (defer until needed)
            - Transitive dependencies (defer)
            - Test files (defer)
            - Documentation (defer)
        """
        return ModuleContext(
            name=name,
            path=get_module_path(name),
            exports=get_module_exports(name),  # Function signatures only
            dependencies=get_direct_dependencies(name),  # Immediate deps
            summary=get_module_summary(name),  # Pre-generated summary
            full_code=None,  # Deferred loading
            tests=None,      # Deferred loading
            docs=None        # Deferred loading
        )

    def load_module_full(self, name: str) -> ModuleContext:
        """Load complete module context (deferred until needed)."""

        # Get minimal context first
        minimal = self.load_on_demand(name)

        # Enhance with full data
        minimal.full_code = read_file(minimal.path)
        minimal.tests = find_related_tests(name)
        minimal.docs = find_related_docs(name)

        # Update cache
        self.cache_module(name, minimal)

        return minimal

    def get_cached_module(self, name: str) -> ModuleContext:
        """Retrieve module from cache."""
        return self.cache[name]

    def cache_module(self, name: str, module: ModuleContext):
        """Store module in cache with timestamp."""
        self.cache[name] = {
            "module": module,
            "cached_at": datetime.now(),
            "access_count": self.cache.get(name, {}).get("access_count", 0) + 1
        }

    def _update_load_stats(self, load_time_ms: float):
        """Update load time statistics."""
        total_loads = self.load_stats["hits"] + self.load_stats["misses"]
        current_avg = self.load_stats["avg_load_time_ms"]

        # Incremental average
        self.load_stats["avg_load_time_ms"] = (
            (current_avg * (total_loads - 1) + load_time_ms) / total_loads
        )

class ModuleContext:
    """Container for module context data."""

    def __init__(self, name: str, path: str, exports: List[str],
                 dependencies: List[str], summary: str,
                 full_code: Optional[str] = None,
                 tests: Optional[List[str]] = None,
                 docs: Optional[str] = None):
        self.name = name
        self.path = path
        self.exports = exports
        self.dependencies = dependencies
        self.summary = summary
        self.full_code = full_code
        self.tests = tests
        self.docs = docs

    def is_fully_loaded(self) -> bool:
        """Check if full context is loaded."""
        return self.full_code is not None
```

---

## Strategy 3: Memory Consolidation

Consolidate session memory for efficient storage and retrieval.

**Consolidation Process**:

```python
def consolidate_memory(session: SessionState) -> CompressedMemory:
    """Compress session memory for efficient storage.

    Args:
        session: Current session state with full history

    Returns:
        Consolidated and compressed memory snapshot

    Compression Strategy:
        - Recent interactions (last 10): Full text
        - Medium history (11-50): Summaries
        - Old history (51+): Knowledge patterns only
        - Active contexts: Compressed representation
    """
    # Phase 1: Extract knowledge patterns
    knowledge_patterns = extract_knowledge_patterns(session.history)

    # Phase 2: Compress interaction history
    compressed_history = compress_history(
        session.history,
        keep_recent=10,        # Full text for recent
        summarize_medium=40,   # Summaries for medium
        extract_patterns_old=True  # Patterns only for old
    )

    # Phase 3: Compress active contexts
    compressed_contexts = compress_active_contexts(
        session.get_active_contexts()
    )

    # Phase 4: Create consolidated snapshot
    consolidated = CompressedMemory(
        session_id=session.id,
        knowledge_patterns=knowledge_patterns,
        compressed_history=compressed_history,
        active_contexts=compressed_contexts,
        compression_ratio=calculate_compression_ratio(session, compressed_history),
        timestamp=datetime.now()
    )

    return consolidated

def extract_knowledge_patterns(history: List[Interaction]) -> List[Pattern]:
    """Extract reusable knowledge patterns from history.

    Pattern Types:
        - Frequent queries (e.g., "How to authenticate?")
        - Common solutions (e.g., JWT token validation)
        - Error patterns (e.g., Missing environment variable)
        - Code patterns (e.g., Database connection setup)
    """
    patterns = []

    # Extract frequent query patterns
    queries = [h.query for h in history]
    frequent_queries = find_frequent_patterns(queries, min_frequency=3)
    patterns.extend([Pattern("query", q) for q in frequent_queries])

    # Extract common solution patterns
    solutions = [h.solution for h in history if h.solution]
    common_solutions = find_common_patterns(solutions)
    patterns.extend([Pattern("solution", s) for s in common_solutions])

    # Extract error handling patterns
    errors = [h.error for h in history if h.error]
    error_patterns = find_error_patterns(errors)
    patterns.extend([Pattern("error", e) for e in error_patterns])

    return patterns

def compress_history(history: List[Interaction],
                    keep_recent: int = 10,
                    summarize_medium: int = 40,
                    extract_patterns_old: bool = True) -> CompressedHistory:
    """Compress interaction history with tiered approach."""

    compressed = CompressedHistory()

    # Tier 1: Recent (full text)
    compressed.recent = history[-keep_recent:] if len(history) >= keep_recent else history

    # Tier 2: Medium (summaries)
    medium_start = max(0, len(history) - keep_recent - summarize_medium)
    medium_end = len(history) - keep_recent
    medium_interactions = history[medium_start:medium_end]
    compressed.medium = [
        summarize_interaction(i) for i in medium_interactions
    ]

    # Tier 3: Old (patterns only)
    if extract_patterns_old:
        old_interactions = history[:medium_start]
        compressed.old_patterns = extract_knowledge_patterns(old_interactions)

    return compressed

class CompressedMemory:
    """Compressed memory snapshot for efficient storage."""

    def __init__(self, session_id: str, knowledge_patterns: List[Pattern],
                 compressed_history: CompressedHistory,
                 active_contexts: List[str],
                 compression_ratio: float,
                 timestamp: datetime):
        self.session_id = session_id
        self.knowledge_patterns = knowledge_patterns
        self.compressed_history = compressed_history
        self.active_contexts = active_contexts
        self.compression_ratio = compression_ratio
        self.timestamp = timestamp

    def restore_session(self) -> SessionState:
        """Restore full session from compressed memory."""
        session = SessionState(self.session_id)

        # Restore recent history (full)
        session.history = self.compressed_history.recent

        # Restore active contexts
        for context_key in self.active_contexts:
            session.load_context(context_key)

        return session

    def get_storage_size(self) -> int:
        """Calculate storage size in tokens."""
        return (
            sum(count_tokens(str(p)) for p in self.knowledge_patterns) +
            count_tokens(str(self.compressed_history)) +
            count_tokens(str(self.active_contexts))
        )
```

---

## Strategy 4: Lazy Loading Pattern

**Defer loading until access required**:

```python
class LazyLoadingManager:
    """Lazy loading with access tracking."""

    def __init__(self):
        self.pending_loads = {}
        self.access_tracking = {}

    def register_lazy_load(self, key: str, loader_fn: Callable):
        """Register lazy load function."""
        self.pending_loads[key] = loader_fn
        self.access_tracking[key] = 0

    def get(self, key: str) -> Any:
        """Get item, loading if necessary."""
        if key in self.pending_loads:
            # Load on first access
            value = self.pending_loads[key]()
            del self.pending_loads[key]
            return value
        return None
```

---

## Strategy 5: Eager Loading Strategy

**Load frequently accessed context eagerly**:

```python
def eager_load_frequent_contexts(session: SessionState):
    """Load frequently accessed contexts at startup."""

    # Analyze access patterns from history
    access_patterns = analyze_access_patterns(session.history)

    # Load top 10 most accessed contexts
    frequent_contexts = sorted(
        access_patterns.items(),
        key=lambda x: x[1],
        reverse=True
    )[:10]

    for context_key, access_count in frequent_contexts:
        if access_count > 5:  # Threshold for eager loading
            session.load_context(context_key)
```

---

## Strategy 6: Adaptive Loading Optimization

**Dynamically adjust loading strategy based on usage**:

```python
class AdaptiveLoader:
    """Adaptive loading strategy based on runtime behavior."""

    def __init__(self, session: SessionState):
        self.session = session
        self.access_patterns = {}
        self.loading_strategy = "progressive"  # Start with progressive

    def analyze_and_adapt(self):
        """Analyze access patterns and adapt loading strategy."""

        # Calculate access frequency
        total_accesses = sum(self.access_patterns.values())
        cache_hit_rate = self.session.cache_hit_rate

        # Adapt strategy based on metrics
        if cache_hit_rate < 0.5:
            # Low cache hit rate: Switch to eager loading
            self.loading_strategy = "eager"
            self.eager_load_frequently_accessed()
        elif cache_hit_rate > 0.9:
            # High cache hit rate: Optimize with lazy loading
            self.loading_strategy = "lazy"
            self.unload_infrequent_contexts()
        else:
            # Moderate hit rate: Keep progressive
            self.loading_strategy = "progressive"

    def eager_load_frequently_accessed(self):
        """Switch to eager loading for frequent contexts."""
        frequent = [k for k, v in self.access_patterns.items() if v > 5]
        for key in frequent:
            self.session.load_context(key)

    def unload_infrequent_contexts(self):
        """Unload rarely accessed contexts to save memory."""
        infrequent = [k for k, v in self.access_patterns.items() if v < 2]
        for key in infrequent:
            self.session.unload_context(key)
```

---

## Performance Optimization

### Loading Performance Targets

| Operation | Target | Measurement |
|-----------|--------|-------------|
| Context seeding | < 2 seconds | Startup time |
| On-demand loading (cached) | < 5ms | Cache hit latency |
| On-demand loading (uncached) | < 500ms | Cache miss latency |
| Memory consolidation | < 5 seconds | Full session compress |
| Context restoration | < 1 second | Reload from compressed |

### Memory Efficiency Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Initial load tokens | < 50K | Seeding phase |
| Full session tokens | < 150K | Active session |
| Compressed snapshot tokens | < 30K | Consolidated storage |
| Compression ratio | > 70% | Size reduction |
| Cache hit rate | > 80% | Progressive loading |

### Loading Strategy Comparison

| Strategy | Startup Time | Memory Usage | Cache Hit Rate | Best Use Case |
|----------|--------------|--------------|----------------|---------------|
| **Lazy Loading** | Fast (<1s) | Low (30K) | Moderate (60%) | Large codebases, infrequent access |
| **Eager Loading** | Slow (5-10s) | High (150K) | High (90%) | Small codebases, frequent access |
| **Progressive** | Medium (2-3s) | Medium (80K) | High (85%) | General purpose, balanced |
| **Adaptive** | Medium (3s) | Dynamic | Highest (95%) | Production systems, optimization |

### Best Practices

**DO**:
- ✅ Load essential context first (P0, P1)
- ✅ Defer non-critical context until needed (P2, P3)
- ✅ Cache frequently accessed modules
- ✅ Compress old interactions into patterns
- ✅ Monitor loading performance metrics
- ✅ Use progressive disclosure for large files
- ✅ Apply adaptive loading for production systems

**DON'T**:
- ❌ Load entire codebase at startup
- ❌ Keep full history without compression
- ❌ Ignore cache hit rates
- ❌ Load transitive dependencies eagerly
- ❌ Skip consolidation for long sessions
- ❌ Use eager loading for large codebases

---

**Version**: 3.0.0
**Last Updated**: 2025-11-23
**Status**: Production Ready
