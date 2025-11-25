# Compression and Caching Optimization

Reduce memory footprint and improve access performance through smart compression and caching strategies.

## Core Optimization Strategies

### Strategy 1: Semantic Compression

Compress content while preserving semantic meaning and critical information.

**Compression Technique**:

```python
def semantic_compress(content: str, target_ratio: float = 0.3) -> str:
    """Compress content while preserving semantic meaning.

    Args:
        content: Original text content to compress
        target_ratio: Target compression ratio (0.3 = keep 30% of original)

    Returns:
        Semantically compressed content

    Compression Algorithm:
        1. Sentence segmentation and scoring
        2. Importance ranking (TF-IDF, PageRank)
        3. Top-N sentence selection
        4. Chronological reordering
        5. Coherence validation
    """
    # Phase 1: Extract and score sentences
    sentences = split_sentences(content)
    scores = {}

    for sentence in sentences:
        scores[sentence] = calculate_importance(sentence, content)

    # Phase 2: Select top sentences by score
    target_count = int(len(sentences) * target_ratio)
    selected = sorted(sentences, key=lambda s: scores[s], reverse=True)[:target_count]

    # Phase 3: Maintain chronological order (preserve flow)
    compressed = " ".join(sorted(selected, key=lambda s: sentences.index(s)))

    # Phase 4: Validate coherence
    coherence_score = validate_coherence(compressed, content)
    if coherence_score < 0.6:
        # Adjust selection to improve coherence
        selected = adjust_for_coherence(selected, sentences, scores)
        compressed = " ".join(sorted(selected, key=lambda s: sentences.index(s)))

    return compressed

def calculate_importance(sentence: str, full_content: str) -> float:
    """Calculate sentence importance score.

    Scoring Factors:
        - TF-IDF relevance (0.4 weight)
        - Position in document (0.2 weight)
        - Sentence length (0.1 weight)
        - Keyword density (0.3 weight)
    """
    tfidf_score = calculate_tfidf(sentence, full_content) * 0.4
    position_score = calculate_position_score(sentence, full_content) * 0.2
    length_score = calculate_length_score(sentence) * 0.1
    keyword_score = calculate_keyword_density(sentence) * 0.3

    return tfidf_score + position_score + length_score + keyword_score

def validate_coherence(compressed: str, original: str) -> float:
    """Validate semantic coherence of compressed content.

    Returns:
        Coherence score (0.0-1.0)
    """
    # Compare semantic similarity
    similarity = calculate_semantic_similarity(compressed, original)

    # Check logical flow
    flow_score = validate_logical_flow(compressed)

    # Combine scores
    return (similarity * 0.7 + flow_score * 0.3)
```

**Compression Ratios by Content Type**:

| Content Type | Target Retention | Typical Reduction | Quality Loss |
|--------------|------------------|-------------------|--------------|
| Code documentation | 40-50% | 50-60% | Minimal |
| API documentation | 30-40% | 60-70% | Low |
| Interaction history | 20-30% | 70-80% | Acceptable |
| Configuration files | 60-70% | 30-40% | None |
| Test logs | 15-25% | 75-85% | Acceptable |

---

### Strategy 2: Hierarchical Summarization

Create multi-level summaries for efficient navigation and compression.

**Multi-Level Summary Creation**:

```python
def hierarchical_summarize(documents: List[str], max_docs: int = 50) -> HierarchicalSummary:
    """Create multi-level summary of large document sets.

    Args:
        documents: List of document texts
        max_docs: Maximum documents to process

    Returns:
        Hierarchical summary with 3 levels

    Summary Levels:
        - Level 1 (Individual): 200 chars per document
        - Level 2 (Group): 150 chars per 5 documents
        - Level 3 (Master): 300 chars total
    """
    # Level 1: Individual document summaries
    individual_summaries = [
        summarize(doc, max_length=200) for doc in documents[:max_docs]
    ]

    # Level 2: Group summaries (5 docs per group)
    group_summaries = []
    for i in range(0, len(individual_summaries), 5):
        group = individual_summaries[i:i+5]
        summary = summarize(" ".join(group), max_length=150)
        group_summaries.append({
            "summary": summary,
            "doc_range": f"{i}-{min(i+4, len(individual_summaries)-1)}",
            "token_count": count_tokens(summary)
        })

    # Level 3: Master summary
    master = summarize(" ".join(group_summaries), max_length=300)

    # Calculate compression metrics
    original_size = sum(len(d) for d in documents)
    compressed_size = len(master) + sum(len(g["summary"]) for g in group_summaries)
    compression_ratio = compressed_size / original_size

    return HierarchicalSummary(
        master_summary=master,
        group_summaries=group_summaries,
        individual_summaries=individual_summaries,
        compression_ratio=compression_ratio,
        original_token_count=count_tokens(" ".join(documents)),
        compressed_token_count=count_tokens(master) + sum(count_tokens(g["summary"]) for g in group_summaries)
    )

class HierarchicalSummary:
    """Container for hierarchical summary data."""

    def __init__(self, master_summary: str, group_summaries: List[dict],
                 individual_summaries: List[str], compression_ratio: float,
                 original_token_count: int, compressed_token_count: int):
        self.master_summary = master_summary
        self.group_summaries = group_summaries
        self.individual_summaries = individual_summaries
        self.compression_ratio = compression_ratio
        self.original_token_count = original_token_count
        self.compressed_token_count = compressed_token_count

    def get_level(self, level: int) -> Union[str, List[str], List[dict]]:
        """Retrieve summary at specific level."""
        if level == 1:
            return self.individual_summaries
        elif level == 2:
            return self.group_summaries
        elif level == 3:
            return self.master_summary
        else:
            raise ValueError(f"Invalid level: {level}. Must be 1, 2, or 3.")

    def tokens_saved(self) -> int:
        """Calculate tokens saved through compression."""
        return self.original_token_count - self.compressed_token_count
```

---

### Strategy 3: Intelligent Cache Management

Implement smart caching with TTL, access patterns, and LRU eviction.

**TTL and Access Pattern Awareness**:

```python
class IntelligentCache:
    """Smart cache with TTL and access pattern analysis."""

    def __init__(self, max_size: int = 1000, default_ttl_hours: int = 24):
        self.cache = {}
        self.access_patterns = {}
        self.ttl_default = timedelta(hours=default_ttl_hours)
        self.max_size = max_size

    def get(self, key: str) -> Optional[Any]:
        """Retrieve with TTL and pattern tracking.

        Args:
            key: Cache key

        Returns:
            Cached value or None if expired/missing

        Side Effects:
            - Updates access count
            - Updates last_accessed timestamp
            - Removes expired entries
        """
        if key not in self.cache:
            return None

        entry = self.cache[key]

        # Check expiration
        if datetime.now() - entry.timestamp > entry.ttl:
            del self.cache[key]
            return None

        # Track access pattern
        self.access_patterns[key] = self.access_patterns.get(key, 0) + 1
        entry.last_accessed = datetime.now()
        entry.hit_count += 1

        return entry.value

    def set(self, key: str, value: Any, ttl: Optional[timedelta] = None):
        """Store with smart TTL calculation.

        Args:
            key: Cache key
            value: Value to cache
            ttl: Optional TTL override

        TTL Calculation Logic:
            - Frequent access (>10 hits): 72 hours
            - Medium access (5-10 hits): 48 hours
            - Low access (<5 hits): 24 hours (default)
        """
        # Calculate smart TTL based on access patterns
        access_count = self.access_patterns.get(key, 0)

        if access_count > 10:
            calculated_ttl = timedelta(hours=72)  # Frequently accessed
        elif access_count > 5:
            calculated_ttl = timedelta(hours=48)  # Medium access
        else:
            calculated_ttl = ttl or self.ttl_default

        # Evict if cache is full
        if len(self.cache) >= self.max_size:
            self.evict_lru(count=int(self.max_size * 0.1))  # Evict 10%

        self.cache[key] = CacheEntry(
            value=value,
            timestamp=datetime.now(),
            ttl=calculated_ttl,
            last_accessed=datetime.now(),
            hit_count=0
        )

    def evict_lru(self, count: int = 5) -> int:
        """Remove least recently used entries.

        Args:
            count: Number of entries to evict

        Returns:
            Number of entries actually evicted
        """
        if len(self.cache) <= count:
            return 0

        # Sort by last access time (LRU)
        sorted_keys = sorted(
            self.cache.keys(),
            key=lambda k: self.cache[k].last_accessed
        )

        # Remove oldest entries
        evicted = 0
        for key in sorted_keys[:count]:
            del self.cache[key]
            evicted += 1

        return evicted

    def get_stats(self) -> Dict[str, Any]:
        """Get cache performance statistics."""
        total_hits = sum(self.cache[k].hit_count for k in self.cache)
        total_entries = len(self.cache)
        avg_hits_per_entry = total_hits / total_entries if total_entries > 0 else 0

        return {
            "total_entries": total_entries,
            "total_hits": total_hits,
            "avg_hits_per_entry": avg_hits_per_entry,
            "utilization": total_entries / self.max_size,
            "access_patterns": self.access_patterns
        }

class CacheEntry:
    """Cache entry with metadata."""

    def __init__(self, value: Any, timestamp: datetime, ttl: timedelta,
                 last_accessed: datetime, hit_count: int = 0):
        self.value = value
        self.timestamp = timestamp
        self.ttl = ttl
        self.last_accessed = last_accessed
        self.hit_count = hit_count

    def is_expired(self) -> bool:
        """Check if entry has expired."""
        return datetime.now() - self.timestamp > self.ttl
```

**Cache Invalidation Rules**:

```python
# Component-specific TTL configuration
cache_invalidation_rules = {
    "configuration": timedelta(hours=24),      # Config changes 1/day max
    "agent_state": timedelta(minutes=30),      # Agent state: 30min
    "knowledge": timedelta(hours=48),          # Knowledge: 2 days
    "interaction": timedelta(hours=1),         # Interactions: 1 hour
    "computed_results": timedelta(hours=12),   # Computational results: 12 hours
    "documentation": timedelta(hours=72),      # Documentation: 3 days
    "static_assets": timedelta(days=7)         # Static content: 7 days
}

def get_ttl_for_key(key: str) -> timedelta:
    """Determine appropriate TTL for cache key."""

    # Extract component from key prefix
    if ":" in key:
        component = key.split(":")[0]
        return cache_invalidation_rules.get(component, timedelta(hours=24))

    return timedelta(hours=24)  # Default TTL
```

---

### Strategy 4: Adaptive Compression

Dynamically adjust compression based on content type and memory pressure.

**Adaptive Compression Algorithm**:

```python
class AdaptiveCompressor:
    """Dynamically adjust compression based on context."""

    def __init__(self, memory_threshold: float = 0.8):
        self.memory_threshold = memory_threshold
        self.compression_history = []

    def compress_adaptive(self, content: str, content_type: str,
                         current_memory_usage: float) -> str:
        """Apply adaptive compression based on memory pressure.

        Args:
            content: Content to compress
            content_type: Type of content (code, docs, logs, etc.)
            current_memory_usage: Current memory utilization (0.0-1.0)

        Returns:
            Compressed content with optimal ratio
        """
        # Determine compression ratio based on memory pressure
        if current_memory_usage > 0.9:
            # Critical: Aggressive compression
            target_ratio = 0.2
        elif current_memory_usage > self.memory_threshold:
            # High: Moderate compression
            target_ratio = 0.4
        elif current_memory_usage > 0.6:
            # Medium: Light compression
            target_ratio = 0.6
        else:
            # Low: Minimal compression
            target_ratio = 0.8

        # Adjust ratio based on content type
        content_type_ratios = {
            "code": 0.7,        # Keep more code
            "documentation": 0.4,  # Docs are compressible
            "logs": 0.2,        # Logs are highly compressible
            "config": 0.9,      # Keep configs intact
            "interaction": 0.3   # Conversations are compressible
        }

        final_ratio = target_ratio * content_type_ratios.get(content_type, 0.5)

        # Perform compression
        compressed = semantic_compress(content, target_ratio=final_ratio)

        # Log compression metrics
        self.compression_history.append({
            "timestamp": datetime.now(),
            "content_type": content_type,
            "original_size": len(content),
            "compressed_size": len(compressed),
            "ratio": len(compressed) / len(content),
            "memory_pressure": current_memory_usage
        })

        return compressed

    def get_compression_stats(self) -> Dict[str, Any]:
        """Get compression statistics."""
        if not self.compression_history:
            return {}

        total_original = sum(h["original_size"] for h in self.compression_history)
        total_compressed = sum(h["compressed_size"] for h in self.compression_history)

        return {
            "total_compressions": len(self.compression_history),
            "total_original_size": total_original,
            "total_compressed_size": total_compressed,
            "avg_compression_ratio": total_compressed / total_original,
            "space_saved": total_original - total_compressed
        }
```

---

## Performance Metrics and Monitoring

### Key Performance Indicators

**Cache Performance**:

| Metric | Target | Measurement Method |
|--------|--------|--------------------|
| Cache hit rate | 70-80% | hits / (hits + misses) |
| Average retrieval time | < 100ms | Time from get() call to return |
| Cache memory overhead | < 5% | Cache size / total memory |
| Eviction rate | < 10%/hour | Evictions per hour |

**Compression Effectiveness**:

| Metric | Target | Measurement Method |
|--------|--------|--------------------|
| Average compression ratio | 40-60% | compressed_size / original_size |
| Compression time | < 500ms | Time to compress |
| Decompression time | < 200ms | Time to decompress |
| Semantic preservation | > 85% | Coherence score |

**Memory Optimization**:

| Metric | Target | Measurement Method |
|--------|--------|--------------------|
| Memory footprint reduction | 50-70% | Optimized / baseline memory |
| Token savings | > 100K/session | Baseline tokens - optimized tokens |
| Access latency | < 50ms | Time to retrieve compressed content |

### Monitoring Dashboard

```python
def generate_optimization_report(session: SessionState) -> OptimizationReport:
    """Generate comprehensive optimization report.

    Returns:
        Report with cache stats, compression metrics, and recommendations
    """
    # Cache statistics
    cache_stats = session.cache.get_stats()

    # Compression statistics
    compression_stats = session.compressor.get_compression_stats()

    # Memory statistics
    memory_stats = profile_memory_usage(session)

    # Generate recommendations
    recommendations = []

    if cache_stats["utilization"] < 0.5:
        recommendations.append("Cache underutilized. Consider reducing max_size.")

    if cache_stats["avg_hits_per_entry"] < 2:
        recommendations.append("Low cache hit rate. Review TTL and eviction policy.")

    if compression_stats["avg_compression_ratio"] > 0.7:
        recommendations.append("Compression not aggressive enough. Increase target ratio.")

    return OptimizationReport(
        cache_stats=cache_stats,
        compression_stats=compression_stats,
        memory_stats=memory_stats,
        recommendations=recommendations,
        timestamp=datetime.now()
    )
```

---

## Best Practices

**DO**:
- ✅ Monitor cache hit rates continuously
- ✅ Adjust TTL based on access patterns
- ✅ Compress proactively before memory pressure
- ✅ Use hierarchical summaries for large documents
- ✅ Profile compression effectiveness regularly
- ✅ Implement adaptive compression for dynamic workloads

**DON'T**:
- ❌ Over-compress critical configuration
- ❌ Ignore cache eviction metrics
- ❌ Use fixed compression ratios for all content types
- ❌ Skip semantic validation after compression
- ❌ Cache volatile data with long TTL
- ❌ Compress already compressed data

---

**Version**: 3.0.0
**Last Updated**: 2025-11-23
**Status**: Production Ready
