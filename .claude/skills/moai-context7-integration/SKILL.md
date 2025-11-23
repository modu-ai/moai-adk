---
name: moai-context7-integration
description: Enterprise Context7 MCP integration for accessing latest documentation across 50+ programming languages and frameworks with real-time library resolution
allowed-tools: [mcp__context7__resolve-library-id, mcp__context7__get-library-docs, Read, Bash, WebFetch]
---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-22  
**compliance_score**: 75%  
**auto_trigger_keywords**: moai, context7, integration  


## Quick Reference (30 seconds)

# Context7 Library Documentation MCP Integration

**What it does**: Unified Context7 MCP integration hub providing real-time documentation access for 50+ programming languages, 200+ frameworks, and libraries through standardized two-step patterns with intelligent caching and progressive disclosure.

**Core Capabilities**:
- ✅ Real-time documentation via Context7 MCP (automatically updated)
- ✅ 50+ languages: Python, JavaScript, TypeScript, Go, Rust, PHP, Ruby, Java, C++, C#, Swift, Kotlin, Scala, R, Elixir, Dart, Haskell, and more
- ✅ 200+ frameworks: FastAPI, Django, React, Next.js, Vue, Angular, Gin, Echo, Rails, Spring Boot, Laravel, and more
- ✅ Intelligent caching with TTL-based invalidation (30 days stable, 7 days beta)
- ✅ Progressive token disclosure (1K-10K tokens)
- ✅ Graceful error handling and fallback strategies

**When to Use**:
- Building language-specific skills needing external documentation
- Integrating latest framework/library docs into guides
- Creating up-to-date API reference materials
- Reducing token consumption through centralized caching
- Implementing documentation access in specialized agents

**Quick Start Example**:
```python
# Step 1: Resolve library name to Context7 ID
library_id = await mcp__context7__resolve_library_id("fastapi")
# Returns: /tiangolo/fastapi

# Step 2: Get documentation with topic focus
docs = await mcp__context7__get_library_docs(
    context7CompatibleLibraryID=library_id,
    topic="routing dependency-injection",
    page=1
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

**Implementation**:
```python
async def resolve_library_id(library_name: str) -> str:
    """
    Resolve user-friendly library name to Context7 ID.
    
    Args:
        library_name: User-friendly name (e.g., "fastapi")
    
    Returns:
        Context7-compatible library ID (e.g., "/tiangolo/fastapi")
    
    Raises:
        LibraryNotFoundError: If library not found
    """
    # Validate input
    if not library_name or len(library_name) > 50:
        raise ValueError("Invalid library name")
    
    # Check cache first (30-day TTL)
    cached_id = get_cached_library_id(library_name)
    if cached_id:
        return cached_id
    
    # Query Context7 MCP
    try:
        library_id = await mcp__context7__resolve_library_id(library_name)
        
        # Cache result
        cache_library_id(library_name, library_id, ttl_days=30)
        
        return library_id
    except Exception as e:
        # Try alternative names
        alternative = try_alternative_names(library_name)
        if alternative:
            return alternative
        
        raise LibraryNotFoundError(f"Library '{library_name}' not found")
```

**Step 2: Documentation Fetching Pattern**:
```
Input: Library ID + topic + token limit
Process:
  1. Check cache for {library_id}:{topic}:{page}
  2. If cache hit and valid (TTL), return cached
  3. If cache miss, query Context7 API
  4. Store result in cache with appropriate TTL
Output: Markdown-formatted documentation
```

**Implementation**:
```python
async def get_library_docs(
    context7_compatible_library_id: str,
    topic: str = "",
    page: int = 1
) -> str:
    """
    Fetch documentation from Context7 MCP.
    
    Args:
        context7_compatible_library_id: Context7 ID (e.g., "/tiangolo/fastapi")
        topic: Focus topic (e.g., "routing dependency-injection")
        page: Page number for pagination (default: 1)
    
    Returns:
        Markdown-formatted documentation
    
    Raises:
        DocumentationNotFoundError: If documentation unavailable
    """
    # Check cache
    cache_key = f"{context7_compatible_library_id}:{topic}:{page}"
    cached_docs = get_cached_docs(cache_key)
    if cached_docs:
        return cached_docs
    
    # Fetch from Context7 MCP
    try:
        docs = await mcp__context7__get_library_docs(
            context7CompatibleLibraryID=context7_compatible_library_id,
            topic=topic,
            page=page
        )
        
        # Cache result (7-day TTL for latest, 30-day for stable)
        ttl_days = 7 if "latest" in context7_compatible_library_id else 30
        cache_docs(cache_key, docs, ttl_days=ttl_days)
        
        return docs
    except Exception as e:
        return handle_documentation_error(e, context7_compatible_library_id, topic)
```

### Token Optimization Strategy

**Progressive Disclosure Levels**:
```python
TOKEN_LEVELS = {
    "summary": 1000,        # Quick reference only
    "standard": 3000,       # API + basic examples
    "detailed": 5000,       # Full documentation
    "comprehensive": 10000  # Complete reference
}

async def get_docs_with_level(
    library_id: str,
    topic: str,
    level: str = "standard"
) -> str:
    """Fetch documentation with predefined token level."""
    tokens = TOKEN_LEVELS.get(level, 5000)
    return await get_library_docs(library_id, topic, tokens)
```

### Error Handling & Fallback Strategies

**Comprehensive Error Recovery**:
```python
def handle_documentation_error(
    error: Exception,
    library_id: str,
    topic: str
) -> str:
    """Handle Context7 errors with graceful fallback."""
    
    if isinstance(error, LibraryNotFoundError):
        # Try alternative names
        alternatives = suggest_similar_libraries(library_id)
        return f"Library not found. Try: {', '.join(alternatives)}"
    
    elif isinstance(error, DocumentationNotFoundError):
        # Try broader topic
        if topic:
            broader_topic = generalize_topic(topic)
            return get_library_docs(library_id, broader_topic, page=1)
        
        # Return cached version if available
        cached = get_any_cached_docs(library_id)
        if cached:
            return f"[CACHED] {cached}"
        
        # Provide manual link
        return f"Documentation unavailable. Visit: https://{extract_repo(library_id)}"
    
    elif isinstance(error, TokenLimitExceededError):
        # Retry with reduced tokens
        return get_library_docs(library_id, topic, page=1)
    
    else:
        return f"Error fetching documentation: {str(error)}"
```

### Caching Implementation

**Cache Architecture**:
```python
CACHE_CONFIG = {
    "storage": ".moai/cache/context7/",
    "index": ".moai/cache/context7/index.json",
    "ttl_stable": 30,  # 30 days for stable versions
    "ttl_beta": 7,     # 7 days for beta/latest
    "ttl_dev": 1,      # 1 day for development
}

def cache_docs(key: str, docs: str, ttl_days: int = 30):
    """Store documentation in cache with TTL."""
    cache_entry = {
        "content": docs,
        "timestamp": datetime.now().isoformat(),
        "ttl_days": ttl_days,
        "expires_at": (datetime.now() + timedelta(days=ttl_days)).isoformat()
    }
    
    # Write to cache file
    cache_path = Path(CACHE_CONFIG["storage"]) / f"{hash(key)}.json"
    cache_path.parent.mkdir(parents=True, exist_ok=True)
    cache_path.write_text(json.dumps(cache_entry))
    
    # Update index
    update_cache_index(key, cache_path)

def get_cached_docs(key: str) -> Optional[str]:
    """Retrieve documentation from cache if valid."""
    cache_path = get_cache_path(key)
    if not cache_path.exists():
        return None
    
    entry = json.loads(cache_path.read_text())
    
    # Check expiration
    expires_at = datetime.fromisoformat(entry["expires_at"])
    if datetime.now() > expires_at:
        cache_path.unlink()  # Remove expired cache
        return None
    
    return entry["content"]
```

### Language Quick Reference

**Core Languages** (for complete mappings see reference.md):
- **Python**: FastAPI, Django, Flask, Pydantic, SQLAlchemy, Pytest
- **JavaScript/TypeScript**: React, Next.js, Vue, Angular, Express, NestJS
- **Go**: Gin, Echo, Fiber, GORM, Testify
- **Rust**: Actix-Web, Axum, Rocket, Tokio, Serde
- **PHP**: Laravel, Symfony, Doctrine, PHPUnit
- **Ruby**: Rails, Sinatra, RSpec
- **Java/Kotlin**: Spring Boot, Quarkus, Ktor, Hibernate, JUnit
- **C++**: Boost, Qt, GoogleTest
- **C# (.NET)**: ASP.NET Core, Entity Framework, xUnit
- **Swift**: Vapor, SwiftUI

For detailed library mappings of all 50+ languages, see: [reference.md](reference.md)

---

## Advanced Patterns

### Multi-Library Tech Stack Integration

**Modern Python Web Stack**:
```python
async def get_python_web_stack_docs() -> dict[str, str]:
    """Fetch documentation for complete Python web stack."""
    
    stack_libraries = {
        "web_framework": "/tiangolo/fastapi",
        "validation": "/pydantic/pydantic",
        "orm": "/sqlalchemy/sqlalchemy",
        "testing": "/pytest-dev/pytest",
        "async": "/aio-libs/aiohttp"
    }
    
    docs = {}
    for component, library_id in stack_libraries.items():
        docs[component] = await get_library_docs(
            library_id,
            topic="best-practices getting-started",
            page=1
        )
    
    return docs
```

**Modern JavaScript Frontend Stack**:
```python
async def get_javascript_frontend_stack() -> dict[str, str]:
    """Fetch documentation for modern JavaScript frontend stack."""
    
    stack = {
        "framework": "/vercel/next.js",
        "ui_library": "/facebook/react",
        "type_system": "/microsoft/TypeScript",
        "styling": "/tailwindlabs/tailwindcss",
        "testing": "/vitest-dev/vitest",
    }
    
    docs = {}
    for component, library_id in stack.items():
        docs[component] = await get_library_docs(library_id, page=1)
    
    return docs
```

### Token Budget Management

**Budget Allocation Pattern**:
```python
class TokenBudgetManager:
    """Manage Context7 token allocation across skills."""
    
    def __init__(self, total_budget: int = 20000):
        self.total_budget = total_budget
        self.allocated = {}
        self.used = {}
    
    def allocate(self, skill_name: str, percentage: int) -> int:
        """Allocate token budget to skill."""
        tokens = int(self.total_budget * percentage / 100)
        self.allocated[skill_name] = tokens
        self.used[skill_name] = 0
        return tokens
    
    def get_available(self, skill_name: str) -> int:
        """Get remaining tokens for skill."""
        return self.allocated.get(skill_name, 0) - self.used.get(skill_name, 0)
    
    def use_tokens(self, skill_name: str, amount: int):
        """Record token usage."""
        available = self.get_available(skill_name)
        if amount > available:
            raise TokenBudgetExceededError(
                f"{skill_name} budget exceeded: {amount} > {available}"
            )
        self.used[skill_name] = self.used.get(skill_name, 0) + amount

# Usage example
budget = TokenBudgetManager(total_budget=20000)
budget.allocate("python-backend", percentage=40)  # 8000 tokens
budget.allocate("javascript-frontend", percentage=30)  # 6000 tokens
budget.allocate("testing", percentage=20)  # 4000 tokens
budget.allocate("infrastructure", percentage=10)  # 2000 tokens
```

### Language-Specific Integration Helpers

**Python Helper**:
```python
class PythonContext7Helper:
    """Python-specific Context7 integration."""
    
    LIBRARY_MAPPINGS = {
        "fastapi": "/tiangolo/fastapi",
        "django": "/django/django",
        "pydantic": "/pydantic/pydantic",
        "sqlalchemy": "/sqlalchemy/sqlalchemy",
        "pytest": "/pytest-dev/pytest",
    }
    
    async def get_framework_docs(self, framework: str, topic: str = "") -> str:
        """Get Python framework documentation."""
        library_id = self.LIBRARY_MAPPINGS.get(framework)
        if not library_id:
            raise ValueError(f"Unknown Python framework: {framework}")
        
        return await get_library_docs(library_id, topic, page=1)
    
    async def get_testing_patterns(self) -> str:
        """Get pytest testing patterns."""
        return await get_library_docs(
            "/pytest-dev/pytest",
            topic="fixtures parametrize mocking best-practices",
            page=1
        )
```

**JavaScript Helper**:
```python
class JavaScriptContext7Helper:
    """JavaScript-specific Context7 integration."""
    
    LIBRARY_MAPPINGS = {
        "react": "/facebook/react",
        "nextjs": "/vercel/next.js",
        "express": "/expressjs/express",
        "typescript": "/microsoft/TypeScript",
        "vitest": "/vitest-dev/vitest",
    }
    
    async def get_framework_docs(self, framework: str, topic: str = "") -> str:
        """Get JavaScript framework documentation."""
        library_id = self.LIBRARY_MAPPINGS.get(framework)
        if not library_id:
            raise ValueError(f"Unknown JavaScript framework: {framework}")
        
        return await get_library_docs(library_id, topic, page=1)
    
    async def get_react_patterns(self) -> str:
        """Get React patterns and best practices."""
        return await get_library_docs(
            "/facebook/react",
            topic="hooks patterns performance best-practices",
            page=1
        )
```

---

## Best Practices

✅ **DO**:
- Use Context7 for latest library documentation
- Implement caching to reduce API calls
- Apply progressive token disclosure (start small, expand as needed)
- Handle errors gracefully with fallback strategies
- Validate library names before querying
- Use language-specific helpers for common tasks
- Monitor cache hit rates for optimization
- Update cache TTL based on version stability

❌ **DON'T**:
- Skip cache checks (waste of tokens)
- Request maximum tokens (10K) unnecessarily
- Ignore error handling (breaks user experience)
- Hardcode library IDs (use mappings or resolution)
- Forget to update cache TTL (leads to stale docs)
- Use blocking calls (always use async)

---

## Related Skills

- `moai-lang-*` - Language-specific skills (Python, TypeScript, Go, etc.)
- `moai-domain-*` - Domain-specific skills (backend, frontend, database)
- `moai-cc-memory` - Token budget optimization
- `moai-foundation-trust` - Quality validation
- `moai-cc-skill-factory` - Skill creation patterns

---

## Changelog

- **v3.0.0** (2025-11-22): Merged moai-context7-integration + moai-context7-lang-integration into unified skill with 50+ languages, 200+ library mappings, comprehensive reference and examples
- **v2.0.0** (2025-11-21): Added language-specific integration patterns
- **v1.0.0** (2025-11-20): Initial Context7 MCP integration

---

**End of Skill** | Updated 2025-11-22 | Status: Production Ready (v3.0.0)

---

## Works Well With

- `moai-context7-lang-integration` - Language-specific Context7 patterns and library mappings
- `moai-cc-skill-factory` - Skill generation with Context7 best practices
- `moai-domain-backend` - Backend Context7 documentation access
- `moai-domain-frontend` - Frontend Context7 documentation access
- `moai-essentials-perf` - Performance Context7 optimization patterns
- `moai-domain-security` - Security Context7 vulnerability patterns

---

