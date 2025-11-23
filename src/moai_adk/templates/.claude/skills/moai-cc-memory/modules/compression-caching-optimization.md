# Compression and Caching Optimization

Reduce memory footprint and improve access performance through smart compression and caching.

## Semantic Compression

### Compression Technique

```python
def semantic_compress(content: str, target_ratio: float = 0.3) -> str:
    '''Compress content while preserving semantic meaning.'''
    
    # Extract sentences
    sentences = split_sentences(content)
    
    # Score each sentence by importance
    scores = {}
    for sentence in sentences:
        scores[sentence] = calculate_importance(sentence, content)
    
    # Select top sentences
    target_count = int(len(sentences) * target_ratio)
    selected = sorted(sentences, key=lambda s: scores[s], reverse=True)[:target_count]
    
    # Maintain original order
    compressed = " ".join(sorted(selected, key=lambda s: sentences.index(s)))
    
    return compressed
```

### Compression Ratios

Typical compression achievements:

- Code documentation: 40-50% retention (50-60% reduction)
- API documentation: 30-40% retention (60-70% reduction)
- Interaction history: 20-30% retention (70-80% reduction)
- Configuration files: 60-70% retention (30-40% reduction)

---

## Hierarchical Summarization

### Multi-Level Summary Creation

```python
def hierarchical_summarize(documents: list[str], max_docs=50) -> dict:
    '''Create multi-level summary of large document sets.'''
    
    # Level 1: Individual summaries
    individual_summaries = [summarize(doc, max_length=200) for doc in documents]
    
    # Level 2: Group summaries (5 docs per group)
    group_summaries = []
    for i in range(0, len(individual_summaries), 5):
        group = individual_summaries[i:i+5]
        summary = summarize(" ".join(group), max_length=150)
        group_summaries.append(summary)
    
    # Level 3: Master summary
    master = summarize(" ".join(group_summaries), max_length=300)
    
    return {
        "master_summary": master,
        "group_summaries": group_summaries,
        "individual_summaries": individual_summaries,
        "compression_ratio": len(master) / sum(len(d) for d in documents)
    }
```

---

## Intelligent Cache Management

### TTL and Access Pattern Awareness

```python
class IntelligentCache:
    def __init__(self):
        self.cache = {}
        self.access_patterns = {}
        self.ttl_default = timedelta(hours=24)
    
    def get(self, key: str):
        '''Retrieve with TTL and pattern tracking.'''
        
        if key not in self.cache:
            return None
        
        entry = self.cache[key]
        
        # Check expiration
        if datetime.now() - entry.timestamp > entry.ttl:
            del self.cache[key]
            return None
        
        # Track access
        self.access_patterns[key] = self.access_patterns.get(key, 0) + 1
        entry.last_accessed = datetime.now()
        
        return entry.value
    
    def set(self, key: str, value, ttl: timedelta = None):
        '''Store with smart TTL calculation.'''
        
        # Frequent access = longer TTL
        if self.access_patterns.get(key, 0) > 10:
            calculated_ttl = timedelta(hours=72)
        else:
            calculated_ttl = ttl or self.ttl_default
        
        self.cache[key] = CacheEntry(
            value=value,
            timestamp=datetime.now(),
            ttl=calculated_ttl,
            last_accessed=datetime.now()
        )
    
    def evict_lru(self, count: int = 5) -> int:
        '''Remove least recently used entries.'''
        
        if len(self.cache) <= count:
            return 0
        
        # Sort by access time
        sorted_keys = sorted(
            self.cache.keys(),
            key=lambda k: self.cache[k].last_accessed
        )
        
        # Remove oldest
        for key in sorted_keys[:count]:
            del self.cache[key]
        
        return count
```

### Cache Invalidation

```python
cache_invalidation_rules = {
    "configuration": timedelta(hours=24),      # Config changes 1/day max
    "agent_state": timedelta(minutes=30),      # Agent state: 30min
    "knowledge": timedelta(hours=48),          # Knowledge: 2 days
    "interaction": timedelta(hours=1),         # Interactions: 1 hour
    "computed_results": timedelta(hours=12)    # Results: 12 hours
}
```

---

## Performance Metrics

Monitor compression effectiveness:

- Cache hit rate: Target 70-80%
- Compression ratio: 40-60% of original
- Average retrieval time: < 100ms
- Cache memory overhead: < 5% of total
