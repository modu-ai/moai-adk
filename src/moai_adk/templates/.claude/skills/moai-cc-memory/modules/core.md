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


## Works Well With

- `moai-cc-skills` - Knowledge capsule storage
- `moai-context7-integration` - Latest memory optimization patterns
- `moai-cc-agents` - Agent state persistence
- `moai-cc-sessions` - Session lifecycle management


## Changelog

- **v3.0.0** (2025-11-21): Enterprise 4-level progressive disclosure, Context7 integration, token budget management
- **v2.0.0** (2025-11-11): Added complete metadata, memory management patterns
- **v1.0.0** (2025-10-22): Initial memory management


**End of Skill** | Updated 2025-11-21
