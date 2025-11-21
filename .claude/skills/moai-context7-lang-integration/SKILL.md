---
name: moai-context7-lang-integration
description: Enterprise-grade Context7 MCP integration patterns for language-specific documentation access
---

## Quick Reference

Enterprise-grade Context7 MCP integration patterns for accessing the latest documentation across all programming languages and frameworks with centralized common patterns.

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
- Implementing documentation access in specialized agents

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

**Step 1: Library Resolution Pattern**:
```
Input: User-friendly name (e.g., "fastapi")
Process:
  1. Validate input (non-empty, <50 chars)
  2. Query Context7: resolve-library-id MCP
  3. Return canonical ID: /org/project/version
  4. Cache for 30 days
Output: Context7-compatible library ID
```

**Implementation Notes**:
- Always validate library name before query
- Implement retry logic (3 attempts max)
- Cache results with TTL (30 days stable, 7 days beta)
- Log failed resolutions for optimization
- Return structured error for missing libraries

**Step 2: Documentation Fetching Pattern**:
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

### Error Handling Pattern

**Case 1: Library Not Found**:
```
Error: resolve-library-id fails (LibraryNotFoundError)
Recovery:
  1. Try alternative names: "FastAPI" → "fast-api"
  2. Search library registry manually
  3. Fallback: Provide manual documentation link
  4. Log for optimization
```

**Case 2: Documentation Unavailable**:
```
Error: get-library-docs returns empty (DocumentationNotFoundError)
Recovery:
  1. Retry with broader topic ("routing" → "api")
  2. Fetch summary instead (1000 tokens)
  3. Use cached version from previous fetch
  4. Provide manual link
```

**Case 3: Token Limit Exceeded**:
```
Error: Requested tokens exceed available (TokenLimitExceededError)
Recovery:
  1. Reduce token count automatically
  2. Retry with minimum viable tokens (1000)
  3. Split request into multiple focused queries
  4. Implement budget management across requests
```

### Language Skills Integration Pattern

**Integration Points**:
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

### Token Budget Management Pattern

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

### Caching Implementation Pattern

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
