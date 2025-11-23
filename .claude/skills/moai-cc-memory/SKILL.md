---
name: moai-cc-memory
description: Claude Code memory management, context persistence, knowledge retention optimization with Progressive Disclosure and modular architecture.
---

## 📊 Skill Metadata

**version**: 1.1.0
**modularized**: true
**last_updated**: 2025-11-23
**compliance_score**: 92%
**auto_trigger_keywords**: memory, context, persistence, token, session

---

## Quick Reference (30 seconds)

Claude Code **Memory Management** optimizes context retention and knowledge persistence through strategic three-layer architecture: working memory, long-term storage, and intelligent caching.

**Core Capabilities**:
- Three-layer memory model (working/long-term/cache)
- Token budget allocation and management
- Memory compression and intelligent caching
- Context seeding and progressive loading
- Retention policies and cleanup

**When to Use**:
- Multi-session state persistence
- Large codebases exceeding context limits
- Token budget optimization (40-60% reduction)
- Long-running agent workflows

**Key Concept**: Layered memory architecture reduces token consumption through strategic compression, selective loading, and intelligent cache invalidation.

---

## 5 Core Patterns (5-10 minutes each)

### Pattern 1: Three-Layer Memory Architecture

**Concept**: Memory organized across three layers with different lifetimes and purposes.

```
Working Memory (Active Context) - 50K tokens, session-only
├─ Current session state
├─ Recently accessed files
├─ Active agent contexts
└─ Lifecycle: Current session

Long-Term Memory (Persistent Storage) - Unlimited (compressed)
├─ Skill knowledge capsules
├─ Agent configurations
├─ Historical context
└─ Lifecycle: Persistent across sessions

Context Cache (Intelligent Buffer) - 20K tokens, TTL-based
├─ Frequently accessed patterns
├─ Compiled knowledge
├─ Precomputed embeddings
└─ Lifecycle: Session + TTL
```

**Use Case**: Structure session memory efficiently for multi-session workflows.

---

### Pattern 2: Context Seeding and Progressive Loading

**Concept**: Initialize sessions with critical context, then load additional data on-demand.

```python
# Seed context with essentials
essential_files = identify_essential_files(context)
agent_configs = load_active_agent_configs()
recent_history = load_recent_history(limit=10)

# Load additional modules as needed
module_context = load_on_demand(module_name)  # Lazy loading
```

**Use Case**: Optimize session startup time for large projects.

---

### Pattern 3: Token Budget Management

**Concept**: Allocate token budgets across components with automatic cleanup triggers.

```python
budget_allocations = {
    "system_context": 0.10,      # 10% for system prompts
    "working_memory": 0.30,      # 30% for active context
    "knowledge_base": 0.20,      # 20% for skills/docs
    "agent_context": 0.15,       # 15% for agent states
    "interaction_buffer": 0.25   # 25% for conversation
}

allocated = allocate_tokens("working_memory", requested=5000)
if insufficient:
    freed_tokens = cleanup_component("interaction_buffer")
```

**Use Case**: Prevent token budget overruns in long-running sessions.

---

### Pattern 4: Memory Compression and Consolidation

**Concept**: Reduce memory footprint through semantic compression and hierarchical summarization.

```python
# Semantic compression: Extract key sentences preserving meaning
compressed = semantic_compress(content, target_ratio=0.3)

# Hierarchical summarization: Multi-level document summaries
master_summary = hierarchical_summarize([doc1, doc2, ...])

# Consolidate memory: Extract patterns and compress history
consolidated = consolidate_memory(session)
```

**Use Case**: Maintain long-term context without exceeding storage limits.

---

### Pattern 5: Intelligent Cache with TTL and Retention Policies

**Concept**: Smart caching with access-pattern-aware TTL and selective memory retention.

```python
# Cache with TTL
cache.set("module_config", value, ttl=timedelta(hours=24))
cached = cache.get("module_config")  # Returns None if expired

# Retention policies
should_retain = should_retain_item(item)
if not should_retain:
    cleanup_memory()  # Remove non-critical items

# LRU eviction for overflow
cache.evict_lru(count=5)  # Remove 5 least-recently-used entries
```

**Use Case**: Optimize cache effectiveness and reduce stale data retention.

---

## Advanced Documentation

Detailed patterns and implementation strategies:

- **[Session Memory Architecture](./modules/session-memory-architecture.md)** - Memory layers, session initialization, state management
- **[Context Loading Strategies](./modules/context-loading-strategies.md)** - Context seeding, progressive loading, consolidation techniques
- **[Retention and Cleanup Policies](./modules/retention-cleanup-policies.md)** - Memory retention rules, forgetting policies, cleanup automation
- **[Token Budget Management](./modules/token-budget-management.md)** - Budget allocation, component cleanup, optimization strategies
- **[Compression and Caching](./modules/compression-caching-optimization.md)** - Semantic compression, hierarchical summarization, intelligent cache invalidation
- **[Memory Profiling and Reference](./modules/memory-profiling-reference.md)** - Performance monitoring, best practices, troubleshooting, integration examples

---

## Works Well With

- `moai-cc-skills` - Knowledge capsule storage and retrieval
- `moai-cc-sessions` - Session lifecycle management
- `moai-cc-configuration` - Environment and secret configuration
- `moai-context7-integration` - Latest memory optimization patterns

---

## Workflow Integration

**Session Initialization**:
```
1. Seed context with essential data (core pattern 2)
2. Allocate token budgets (core pattern 3)
3. Initialize three-layer memory (core pattern 1)
4. Setup cache with retention policies (core pattern 5)
```

**Long-Running Session**:
```
1. Monitor token usage against budgets
2. Compress memory as needed (core pattern 4)
3. Load additional context on-demand (core pattern 2)
4. Clean up expired cache entries (core pattern 5)
```

---

## Changelog

- **v1.1.0** (2025-11-23): Progressive Disclosure refactoring, modularized structure, 6-module architecture
- **v1.0.0** (2025-10-22): Initial memory management with Context7 integration

---

**End of Skill** | Modularized 2025-11-23 | [View Modules](./modules/)
