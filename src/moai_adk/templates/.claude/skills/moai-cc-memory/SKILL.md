---
name: moai-cc-memory
description: Claude Code memory management, context persistence, knowledge retention optimization, and token budget management with Context7 integration for latest memory patterns.
---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-22  
**compliance_score**: 75%  
**auto_trigger_keywords**: memory, cc, moai  


## Quick Reference (30 seconds)

Claude Code **Memory Management** optimizes context retention and knowledge persistence across sessions through strategic caching, compression, and retrieval patterns.

**Core Capabilities**:
- Session memory management (working memory optimization)
- Context persistence strategies (long-term storage)
- Knowledge retention optimization (selective caching)
- Memory cleanup and compression
- Token budget management

**When to Use**:
- Long-running sessions requiring state persistence
- Large codebases exceeding context limits
- Optimizing token usage across agent workflows
- Managing knowledge across multiple sessions

**Key Concept**: Effective memory management reduces token consumption by 40-60% through smart caching and selective loading.

---

## Implementation Guide

### Memory Architecture Layers

**Three-Layer Memory Model**:

```
Working Memory (Active Context):
  └─ Current session state
  └─ Recently accessed files/data
  └─ Active agent contexts
  └─ Size: ~50K tokens
  └─ Lifetime: Current session only

Long-Term Memory (Persistent Storage):
  └─ Skill knowledge capsules
  └─ Agent configurations
  └─ Historical project context
  └─ Size: Unlimited (compressed)
  └─ Lifetime: Persistent across sessions

Context Cache (Intelligent Buffer):
  └─ Frequently accessed patterns
  └─ Compiled knowledge
  └─ Precomputed embeddings
  └─ Size: ~20K tokens
  └─ Lifetime: Session + TTL
```

### Memory Management Strategies

**Strategy 1: Context Seeding**

```python
def seed_context(session_id: str, initial_context: dict) -> SessionState:
    """
    Initialize session with strategic context.
    
    Args:
        session_id: Unique session identifier
        initial_context: Critical context for session start
    
    Returns:
        Initialized session state
    """
    # Load only essential context
    essential_files = identify_essential_files(initial_context)
    agent_configs = load_active_agent_configs()
    recent_history = load_recent_history(limit=10)  # Last 10 interactions
    
    session = SessionState(
        id=session_id,
        files=essential_files,
        agents=agent_configs,
        history=recent_history,
        created_at=datetime.now()
    )
    
    # Compress and cache
    session.compress()
    cache_session(session)
    
    return session
```

**Strategy 2: Progressive Loading**

```python
class ProgressiveContextLoader:
    """Load context progressively based on demand."""
    
    def __init__(self, session: SessionState):
        self.session = session
        self.loaded_modules = set()
    
    def load_on_demand(self, module_name: str) -> ModuleContext:
        """
        Load module context only when requested.
        
        Args:
            module_name: Module to load
        
        Returns:
            Module context with essential data
        """
        if module_name in self.loaded_modules:
            # Return cached version
            return self.get_cached_module(module_name)
        
        # Load incrementally
        module_context = self.load_module_minimal(module_name)
        
        # Cache for future use
        self.cache_module(module_name, module_context)
        self.loaded_modules.add(module_name)
        
        return module_context
    
    def load_module_minimal(self, module_name: str) -> ModuleContext:
        """Load only essential module information."""
        
        return ModuleContext(
            name=module_name,
            exports=get_module_exports(module_name),
            dependencies=get_direct_dependencies(module_name),
            summary=get_module_summary(module_name),  # Pre-generated summary
            full_code=None  # Don't load full code until explicitly needed
        )
```

**Strategy 3: Memory Consolidation**

```python
def consolidate_memory(session: SessionState) -> CompressedMemory:
    """
    Consolidate session memory for efficient storage.
    
    Args:
        session: Current session state
    
    Returns:
        Consolidated and compressed memory
    """
    # Extract knowledge patterns
    knowledge_patterns = extract_patterns(session.history)
    
    # Compress interaction history
    compressed_history = compress_history(
        session.history,
        keep_recent=10,  # Keep last 10 interactions in full
        summarize_older=True  # Summarize older interactions
    )
    
    # Create memory snapshot
    consolidated = CompressedMemory(
        session_id=session.id,
        knowledge_patterns=knowledge_patterns,
        compressed_history=compressed_history,
        active_contexts=session.get_active_contexts(),
        timestamp=datetime.now()
    )
    
    return consolidated
```

**Strategy 4: Forgetting Policies**

```python
class MemoryRetentionPolicy:
    """Define what to remember and what to forget."""
    
    def __init__(self, session: SessionState):
        self.session = session
        self.retention_rules = self.load_retention_rules()
    
    def should_retain(self, item: MemoryItem) -> bool:
        """
        Determine if memory item should be retained.
        
        Args:
            item: Memory item to evaluate
        
        Returns:
            True if should retain, False if should forget
        """
        # Critical information always retained
        if item.is_critical():
            return True
        
        # Recently accessed information retained
        if item.last_accessed > datetime.now() - timedelta(hours=24):
            return True
        
        # Frequently accessed information retained
        if item.access_count > self.retention_rules["min_access_count"]:
            return True
        
        # Information related to active tasks retained
        if item.relates_to_active_task(self.session.active_tasks):
            return True
        
        # Otherwise, safe to forget
        return False
    
    def cleanup_memory(self) -> int:
        """Remove items that don't need retention."""
        
        removed_count = 0
        for item in self.session.memory_items:
            if not self.should_retain(item):
                self.session.remove_memory(item)
                removed_count += 1
        
        return removed_count
```

### Token Budget Management

**Budget Allocation Strategy**:

```python
class TokenBudgetManager:
    """Manage token allocation across session components."""
    
    def __init__(self, total_budget: int = 200000):
        self.total_budget = total_budget
        self.allocations = {
            "system_context": 0.10,     # 10% for system prompts
            "working_memory": 0.30,     # 30% for active context
            "knowledge_base": 0.20,     # 20% for skills/docs
            "agent_context": 0.15,      # 15% for agent states
            "interaction_buffer": 0.25  # 25% for conversation
        }
        self.current_usage = {k: 0 for k in self.allocations}
    
    def allocate_tokens(self, component: str, requested: int) -> int:
        """
        Allocate tokens to component within budget.
        
        Args:
            component: Component requesting tokens
            requested: Number of tokens requested
        
        Returns:
            Number of tokens allocated
        """
        max_allowed = int(self.total_budget * self.allocations[component])
        current = self.current_usage[component]
        available = max_allowed - current
        
        if requested <= available:
            # Grant full request
            self.current_usage[component] += requested
            return requested
        else:
            # Partial allocation or trigger cleanup
            if available > 0:
                self.current_usage[component] += available
                return available
            else:
                # Budget exceeded, trigger cleanup
                self.cleanup_component(component)
                return self.allocate_tokens(component, requested)
    
    def cleanup_component(self, component: str) -> int:
        """Free up tokens in component."""
        
        freed_tokens = 0
        
        if component == "working_memory":
            # Compress or evict least recently used
            freed_tokens = compress_working_memory()
        elif component == "knowledge_base":
            # Unload cached knowledge
            freed_tokens = unload_cached_knowledge()
        elif component == "interaction_buffer":
            # Summarize old interactions
            freed_tokens = summarize_old_interactions()
        
        self.current_usage[component] -= freed_tokens
        return freed_tokens
```

### Context7 Integration for Memory Patterns

**Fetch Latest Memory Optimization Patterns**:

```python
async def get_memory_optimization_patterns() -> MemoryPatterns:
    """Fetch latest memory optimization patterns from Context7."""
    
    # Get Claude Code memory management patterns
    patterns = await context7.get_library_docs(
        context7_library_id="/anthropic/claude-code",
        topic="memory management token optimization context persistence 2025",
        tokens=3000
    )
    
    return MemoryPatterns(
        compression_strategies=patterns["compression"],
        caching_patterns=patterns["caching"],
        best_practices=patterns["best_practices"]
    )
```

---

## Advanced Patterns

### Memory Compression Techniques

**Technique 1: Semantic Compression**

```python
def semantic_compress(content: str, target_ratio: float = 0.3) -> str:
    """
    Compress content while preserving semantic meaning.
    
    Args:
        content: Original content
        target_ratio: Target compression ratio (0.3 = 30% of original)
    
    Returns:
        Compressed content with preserved semantics
    """
    # Extract key sentences
    sentences = split_sentences(content)
    sentence_scores = score_sentences(sentences)
    
    # Select top sentences by score
    target_count = int(len(sentences) * target_ratio)
    selected_sentences = sorted(
        sentences,
        key=lambda s: sentence_scores[s],
        reverse=True
    )[:target_count]
    
    # Reorder chronologically
    compressed = " ".join(sorted(
        selected_sentences,
        key=lambda s: sentences.index(s)
    ))
    
    return compressed
```

**Technique 2: Hierarchical Summarization**

```python
def hierarchical_summarize(documents: list[str]) -> str:
    """
    Create multi-level summary of documents.
    
    Args:
        documents: List of document texts
    
    Returns:
        Hierarchical summary
    """
    # Level 1: Individual document summaries
    doc_summaries = [summarize(doc, max_length=200) for doc in documents]
    
    # Level 2: Group summaries
    group_summaries = []
    for i in range(0, len(doc_summaries), 5):
        group = doc_summaries[i:i+5]
        group_summary = summarize(" ".join(group), max_length=150)
        group_summaries.append(group_summary)
    
    # Level 3: Master summary
    master_summary = summarize(" ".join(group_summaries), max_length=300)
    
    return HierarchicalSummary(
        master=master_summary,
        groups=group_summaries,
        individuals=doc_summaries
    )
```

### Intelligent Caching

**Cache Invalidation Strategy**:

```python
class IntelligentCache:
    """Smart cache with TTL and access pattern analysis."""
    
    def __init__(self):
        self.cache = {}
        self.access_patterns = {}
        self.ttl_default = timedelta(hours=24)
    
    def get(self, key: str) -> Optional[Any]:
        """Retrieve from cache with access pattern tracking."""
        
        if key in self.cache:
            entry = self.cache[key]
            
            # Check TTL
            if datetime.now() - entry.timestamp > entry.ttl:
                # Expired, remove
                del self.cache[key]
                return None
            
            # Update access pattern
            self.access_patterns[key] = self.access_patterns.get(key, 0) + 1
            entry.last_accessed = datetime.now()
            
            return entry.value
        
        return None
    
    def set(self, key: str, value: Any, ttl: Optional[timedelta] = None) -> None:
        """Store in cache with smart TTL."""
        
        # Determine TTL based on access pattern
        if key in self.access_patterns and self.access_patterns[key] > 10:
            # Frequently accessed: longer TTL
            calculated_ttl = timedelta(hours=72)
        else:
            calculated_ttl = ttl or self.ttl_default
        
        self.cache[key] = CacheEntry(
            value=value,
            timestamp=datetime.now(),
            ttl=calculated_ttl,
            last_accessed=datetime.now()
        )
    
    def evict_lru(self, count: int = 1) -> int:
        """Evict least recently used entries."""
        
        if len(self.cache) <= count:
            return 0
        
        # Sort by last access time
        sorted_keys = sorted(
            self.cache.keys(),
            key=lambda k: self.cache[k].last_accessed
        )
        
        # Remove oldest
        for key in sorted_keys[:count]:
            del self.cache[key]
        
        return count
```

### Performance Optimization

**Memory Profiling**:

```python
def profile_memory_usage(session: SessionState) -> MemoryProfile:
    """
    Profile memory usage across session components.
    
    Args:
        session: Current session state
    
    Returns:
        Detailed memory usage breakdown
    """
    profile = MemoryProfile()
    
    # Measure working memory
    profile.working_memory = measure_tokens(session.active_contexts)
    
    # Measure knowledge base
    profile.knowledge_base = measure_tokens(session.loaded_skills)
    
    # Measure interaction buffer
    profile.interaction_buffer = measure_tokens(session.conversation_history)
    
    # Measure agent contexts
    profile.agent_contexts = sum(
        measure_tokens(agent.state) for agent in session.active_agents
    )
    
    # Calculate totals
    profile.total_tokens = sum([
        profile.working_memory,
        profile.knowledge_base,
        profile.interaction_buffer,
        profile.agent_contexts
    ])
    
    profile.utilization_rate = profile.total_tokens / session.token_budget
    
    return profile
```

---

## Works Well With

- `moai-cc-skills` - Knowledge capsule storage
- `moai-context7-integration` - Latest memory optimization patterns
- `moai-cc-agents` - Agent state persistence
- `moai-cc-sessions` - Session lifecycle management

---

## Changelog

- **v3.0.0** (2025-11-21): Enterprise 4-level progressive disclosure, Context7 integration, token budget management
- **v2.0.0** (2025-11-11): Added complete metadata, memory management patterns
- **v1.0.0** (2025-10-22): Initial memory management

---

**End of Skill** | Updated 2025-11-21
