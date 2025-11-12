---
name: moai-context7-lang-integration
version: "4.0.0"
status: stable
description: "Enterprise-grade Context7 MCP integration patterns for language-specific documentation access with real-time library resolution and intelligent caching"
keywords: [context7, mcp, documentation, integration, language-specific, library-resolution, caching, async-patterns, error-handling]
allowed-tools:
  - Read
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
tier: foundation
tags:
---

# Context7 Language Integration

## Quick Reference

### What It Does

This Skill provides enterprise-grade Context7 MCP integration patterns for accessing the latest documentation across all programming languages and frameworks. It centralizes common patterns that Language Skills reuse to reduce code duplication and token overhead.

**Core Capabilities**:
- Resolve library names to Context7-compatible IDs in real-time
- Fetch latest documentation (automatically updated)
- Intelligent caching with TTL-based invalidation
- Progressive token disclosure (1000-10000 tokens)
- Graceful error handling and fallback strategies

### When to Use

Use this Skill when:
- Building language-specific Skills that need external documentation
- Integrating latest framework/library documentation into guides
- Creating up-to-date API reference materials
- Reducing token consumption through centralized caching
- Implementing documentation access in specialized agents

### Core Concepts

**1. Library ID Format**
- Format: `/org/project` or `/org/project/version`
- Examples: `/tiangolo/fastapi`, `/tiangolo/fastapi/v0.115.0`
- Returned by resolve-library-id, used by get-library-docs

**2. Two-Step Integration Pattern**
- Step 1: resolve-library-id("library-name") → get canonical ID
- Step 2: get-library-docs(canonical-id) → fetch documentation
- Separation allows caching at both layers

**3. Progressive Disclosure**
- Token budget: 1000-10000 tokens per request
- Default: 5000 tokens (balanced)
- Specific topic focus reduces token consumption
- Multiple requests accumulate toward session limit

**4. Error Handling Strategy**
- Try primary library name
- Fallback to alternative names (case variations)
- Return manual documentation link as last resort
- Log failures for debugging and optimization

**5. Caching Strategy**
- Cache key: `{library_id}:{topic}:{tokens}`
- TTL: 30 days (stable versions)
- TTL: 7 days (beta/latest versions)
- Storage: `.moai/cache/context7/`
- Monitoring: Track hit/miss rates

### Quick Start Example

```python
# Step 1: Resolve library name to Context7 ID
from mcp__context7__resolve_library_id import resolve_library_id
from mcp__context7__get_library_docs import get_library_docs

library_name = "fastapi"
library_id = resolve_library_id(library_name)
# Returns: /tiangolo/fastapi or /tiangolo/fastapi/v0.115.0

# Step 2: Get documentation with topic focus
docs = get_library_docs(
    context7_compatible_library_id=library_id,
    tokens=3000,
    topic="routing"
)

# Step 3: Use in workflow
print(f"Documentation for {library_name}:\n{docs[:500]}")
```

---

## Implementation

### Step 1: Library Resolution Pattern

**Purpose**: Convert user-friendly library names to canonical Context7 IDs

**Pattern Design**:
```
Input: User-friendly name (e.g., "fastapi")
Process:
  1. Validate input (non-empty, reasonable length)
  2. Query Context7: resolve-library-id MCP
  3. Return canonical ID: /org/project/version
  4. Cache for 30 days
Output: Context7-compatible library ID
```

**Implementation Notes**:
- Always validate library name before query (non-empty, <50 chars)
- Implement retry logic (3 attempts max)
- Cache results with TTL (30 days stable, 7 days beta)
- Log failed resolutions for optimization
- Return structured error for missing libraries

**Error Cases**:
- Library not found → Try alternative names
- Network timeout → Use cached version
- Malformed name → Return clear error message

### Step 2: Documentation Fetching Pattern

**Purpose**: Get specific documentation topics to optimize token usage

**Pattern Design**:
```
Input: Library ID + topic + token limit
Process:
  1. Check cache for {library_id}:{topic}:{tokens}
  2. If cache hit and valid (TTL), return cached
  3. If cache miss, query Context7 API
  4. Store result in cache
Output: Markdown-formatted documentation
```

**Token Optimization Strategy**:
- Default: 5000 tokens (balanced documentation)
- Specific topic: 3000 tokens (focused API docs)
- Summary only: 1000 tokens (quick reference)
- Comprehensive: 10000 tokens (full reference)

**Progressive Disclosure Levels**:
- Level 1 (1000 tokens): Core API only
- Level 2 (3000 tokens): API + examples
- Level 3 (5000 tokens): Full documentation
- Level 4 (10000 tokens): Full docs + advanced

### Step 3: Error Handling Pattern

**Case 1: Library Not Found**
```
Error: resolve-library-id fails (LibraryNotFoundError)
Recovery:
  1. Try alternative names: "FastAPI" → "fast-api"
  2. Search library registry manually
  3. Fallback: Provide manual documentation link
  4. Log for optimization
```

**Case 2: Documentation Unavailable**
```
Error: get-library-docs returns empty (DocumentationNotFoundError)
Recovery:
  1. Retry with broader topic ("routing" → "api")
  2. Fetch summary instead (1000 tokens)
  3. Use cached version from previous fetch
  4. Provide manual link
```

**Case 3: Token Limit Exceeded**
```
Error: Requested tokens exceed available (TokenLimitExceededError)
Recovery:
  1. Reduce token count automatically
  2. Retry with minimum viable tokens (1000)
  3. Split request into multiple focused queries
  4. Implement budget management across requests
```

### Step 4: Language Skills Integration Pattern

**Integration Points**:
Each Language Skill should include:

```markdown
## External Documentation Access

For up-to-date library documentation:
Skill("moai-context7-lang-integration")

**Available libraries**:
- FastAPI: `/tiangolo/fastapi`
- Django: `/django/django`
- Pydantic: `/pydantic/pydantic`
- SQLAlchemy: `/sqlalchemy/sqlalchemy`
```

**Usage in Implementation**:
```python
def get_fastapi_routing_guide(tokens: int = 3000) -> str:
    # Use Context7 integration
    library_id = "/tiangolo/fastapi"
    docs = get_library_docs(
        context7_compatible_library_id=library_id,
        topic="routing",
        tokens=tokens
    )
    return docs
```

### Step 5: Caching Implementation Pattern

**Cache Architecture**:
```
Cache Structure:
  Key: {library_id}:{topic}:{tokens}
  Value: {content, timestamp, version}
  Storage: .moai/cache/context7/
  Index: .moai/cache/context7/index.json
```

**Caching Rules**:
- TTL: 30 days for stable versions (v1.0, v2.0)
- TTL: 7 days for latest/beta (main, beta, rc)
- TTL: 1 day for development branches
- Hit rate tracking (monitor for optimization)

**Cache Invalidation Strategy**:
- Auto-invalidate on version mismatch
- Manual invalidation option (--clear-cache)
- Smart invalidation (version changed)

**Implementation Checklist**:
- Create cache directory if missing
- Implement TTL validation on read
- Log cache hits/misses for metrics
- Implement cache cleanup (retention policy)
- Monitor for stale entries

---

## Advanced

### Multi-Library Integration Pattern

**Use Case**: Build comprehensive tech stack documentation

```python
def get_tech_stack_docs(stack_name: str) -> dict:
    """Fetch documentation for entire tech stack"""

    stacks = {
        "modern-python": {
            "fastapi": "/tiangolo/fastapi",
            "pydantic": "/pydantic/pydantic",
            "sqlalchemy": "/sqlalchemy/sqlalchemy"
        },
        "modern-js": {
            "nextjs": "/vercel/next.js",
            "react": "/facebook/react",
            "typescript": "/microsoft/TypeScript"
        }
    }

    libraries = stacks.get(stack_name, {})
    results = {}

    for lib_name, lib_id in libraries.items():
        docs = get_library_docs(
            context7_compatible_library_id=lib_id,
            tokens=2000  # Shared token budget
        )
        results[lib_name] = docs

    return results
```

### Parallel Request Pattern

**Benefits**: Reduce total fetch time for multiple libraries

```python
import asyncio

async def fetch_library_async(lib_id: str, topic: str) -> str:
    # Implementation depends on async Context7 client
    return get_library_docs(
        context7_compatible_library_id=lib_id,
        topic=topic,
        tokens=3000
    )

async def fetch_multiple(libraries: list) -> dict:
    tasks = [
        fetch_library_async(lib['id'], lib['topic'])
        for lib in libraries
    ]
    results = await asyncio.gather(*tasks, return_exceptions=True)
    return dict(zip([l['name'] for l in libraries], results))
```

### Token Budget Management Pattern

**Strategy**: Allocate and track tokens across multiple requests

```python
class TokenBudgetManager:
    def __init__(self, total_budget: int = 20000):
        self.total_budget = total_budget
        self.allocated = {}
        self.used = {}

    def allocate(self, skill_name: str, percentage: int) -> int:
        tokens = int(self.total_budget * percentage / 100)
        self.allocated[skill_name] = tokens
        self.used[skill_name] = 0
        return tokens

    def get_available(self, skill_name: str) -> int:
        return self.allocated.get(skill_name, 0) - self.used.get(skill_name, 0)

    def use_tokens(self, skill_name: str, amount: int):
        available = self.get_available(skill_name)
        if amount > available:
            raise TokenBudgetExceededError(
                f"{skill_name} budget exceeded: {amount} > {available}"
            )
        self.used[skill_name] = self.used.get(skill_name, 0) + amount
```

### Performance Optimization Pattern

**Techniques**:

1. **Lazy Loading**: Fetch only when explicitly needed
   ```python
   docs = None  # Don't fetch until requested
   if user_needs_full_docs():
       docs = get_library_docs(...)
   ```

2. **Partial Content**: Request specific topic instead of full docs
   ```python
   # Good: Fetch only routing documentation
   docs = get_library_docs(topic="routing", tokens=1000)
   # Bad: Fetch entire documentation
   docs = get_library_docs(tokens=10000)
   ```

3. **Compression**: Gzip cached documentation
   ```python
   import gzip
   compressed = gzip.compress(docs.encode())  # ~60% reduction
   ```

4. **Indexing**: Create searchable index of cached docs
   ```python
   index = {
       "routing": {offset: 0, length: 5000},
       "async": {offset: 5000, length: 3000},
   }
   ```

5. **Version Pinning**: Cache specific versions longer
   ```python
   library_id = "/tiangolo/fastapi/v0.115.0"  # Pinned version
   # Cached for 90 days instead of 30
   ```

### Custom Library Mapping Pattern

**Purpose**: Map language-specific names to canonical Context7 IDs

```python
LIBRARY_MAPPINGS = {
    "python": {
        "fastapi": "/tiangolo/fastapi",
        "FastAPI": "/tiangolo/fastapi",
        "fast-api": "/tiangolo/fastapi",
        "django": "/django/django",
        "Django": "/django/django",
        "pydantic": "/pydantic/pydantic",
    },
    "javascript": {
        "next.js": "/vercel/next.js",
        "nextjs": "/vercel/next.js",
        "NextJS": "/vercel/next.js",
        "react": "/facebook/react",
        "React": "/facebook/react",
    },
    "go": {
        "gin": "/gin-gonic/gin",
        "Gin": "/gin-gonic/gin",
        "gorm": "/go-gorm/gorm",
        "GORM": "/go-gorm/gorm",
    }
}

def resolve_with_mapping(language: str, library_name: str) -> str:
    mapping = LIBRARY_MAPPINGS.get(language, {})
    canonical_id = mapping.get(library_name)
    if canonical_id:
        return canonical_id
    # Fallback to MCP resolver
    return resolve_library_id(library_name)
```

