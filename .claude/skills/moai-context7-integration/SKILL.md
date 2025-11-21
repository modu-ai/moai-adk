---
name: moai-context7-integration
description: Enhanced context7 integration with AI-powered features
---

## Quick Reference

Context7 integration patterns for accessing latest documentation across all programming languages and frameworks with real-time library resolution and intelligent caching.

**Core Capabilities**:
- Resolve library names to Context7-compatible IDs in real-time
- Fetch latest documentation (automatically updated)
- Intelligent caching with TTL-based invalidation
- Progressive token disclosure (1000-10000 tokens)
- Graceful error handling and fallback strategies

**When to Use**:
- Building language-specific Skills that need external documentation
- Integrating latest framework/library documentation into guides
- Creating up-to-date API reference materials
- Reducing token consumption through centralized caching

**Quick Start Example**:
```python
# Step 1: Resolve library name to Context7 ID
library_id = resolve_library_id("fastapi")
# Returns: /tiangolo/fastapi

# Step 2: Get documentation with topic focus
docs = get_library_docs(
    context7_compatible_library_id=library_id,
    tokens=3000,
    topic="routing"
)
```

---

## Implementation Guide

### Two-Step Integration Pattern

**Step 1: Library Resolution**:
```
Input: User-friendly name (e.g., "fastapi")
Process:
  1. Validate input (non-empty, reasonable length)
  2. Query Context7: resolve-library-id MCP
  3. Return canonical ID: /org/project/version
  4. Cache for 30 days
Output: Context7-compatible library ID
```

**Step 2: Documentation Fetching**:
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

**Error Handling Pattern**:
```
Case 1: Library Not Found
  → Try alternative names: "FastAPI" → "fast-api"
  → Search library registry manually
  → Fallback: Provide manual documentation link

Case 2: Documentation Unavailable
  → Retry with broader topic ("routing" → "api")
  → Fetch summary instead (1000 tokens)
  → Use cached version from previous fetch

Case 3: Token Limit Exceeded
  → Reduce token count automatically
  → Retry with minimum viable tokens (1000)
  → Split request into multiple focused queries
```

---

## Advanced Patterns

### Multi-Library Integration Pattern

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
            tokens=2000
        )
        results[lib_name] = docs

    return results
```

### Custom Library Mapping Pattern

```python
LIBRARY_MAPPINGS = {
    "python": {
        "fastapi": "/tiangolo/fastapi",
        "django": "/django/django",
        "pydantic": "/pydantic/pydantic",
    },
    "javascript": {
        "nextjs": "/vercel/next.js",
        "react": "/facebook/react",
    },
    "go": {
        "gin": "/gin-gonic/gin",
        "gorm": "/go-gorm/gorm",
    }
}

def resolve_with_mapping(language: str, library_name: str) -> str:
    mapping = LIBRARY_MAPPINGS.get(language, {})
    canonical_id = mapping.get(library_name)
    if canonical_id:
        return canonical_id
    return resolve_library_id(library_name)
```

**Caching Strategy**:
- Cache key: `{library_id}:{topic}:{tokens}`
- TTL: 30 days (stable versions)
- TTL: 7 days (beta/latest versions)
- Storage: `.moai/cache/context7/`
- Monitoring: Track hit/miss rates
