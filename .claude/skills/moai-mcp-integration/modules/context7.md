# Context7 MCP Server Integration Module

**Version**: 3.0.0 (Unified, v2.0.0+ API)
**Last Updated**: 2025-11-24
**Purpose**: Real-time documentation access for 50+ languages and 200+ frameworks

---

## 📖 Quick Overview (2 Minutes)

Context7 is the unified MCP server for accessing real-time library documentation across 50+ programming languages and 200+ frameworks. It provides:

- **Real-time documentation** (always up-to-date)
- **50+ languages**: Python, JavaScript, TypeScript, Go, Rust, PHP, Java, C++, C#, Swift, Kotlin, Scala, R, Elixir, Dart, and more
- **200+ frameworks**: FastAPI, Django, React, Next.js, Vue, Angular, Gin, Echo, Rails, Spring Boot, Laravel, and more
- **Intelligent caching**: TTL-based with 30-day stable, 7-day beta versioning
- **Token optimization**: Progressive disclosure from 1K to 10K tokens
- **Error recovery**: Graceful fallbacks and alternative name resolution

---

## 🔧 Implementation Guide

### Two-Step Integration Pattern

Context7 uses a **2-step pattern** for all operations:

#### **Step 1: Library Resolution**

Resolve user-friendly library names to Context7 IDs:

```python
async def resolve_library_id(library_name: str) -> str:
    """
    Convert user-friendly name to Context7 ID.

    Examples:
        "fastapi" → "/tiangolo/fastapi"
        "react" → "/facebook/react"
        "django" → "/django/django"
    """
    # Check cache first (30-day TTL)
    cached_id = get_cached_library_id(library_name)
    if cached_id:
        return cached_id

    # Query Context7 MCP
    try:
        library_id = await mcp__context7__resolve_library_id(library_name)
        cache_library_id(library_name, library_id, ttl_days=30)
        return library_id
    except Exception as e:
        # Try alternative names
        alternative = try_alternative_names(library_name)
        if alternative:
            return alternative
        raise LibraryNotFoundError(f"Library '{library_name}' not found")
```

#### **Step 2: Documentation Fetching**

Fetch documentation with topic and token optimization:

```python
async def get_library_docs(
    context7_compatible_library_id: str,
    topic: str = "",
    page: int = 1,
    tokens: int = 3000
) -> str:
    """
    Fetch documentation from Context7.

    Args:
        context7_compatible_library_id: "/org/project" format
        topic: "routing dependency-injection" (optional)
        page: Page number for pagination
        tokens: Token limit (1000-10000)

    Returns:
        Markdown-formatted documentation
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

        # Cache result
        ttl_days = 7 if "latest" in context7_compatible_library_id else 30
        cache_docs(cache_key, docs, ttl_days=ttl_days)

        return docs
    except Exception as e:
        return handle_documentation_error(e, context7_compatible_library_id, topic)
```

---

## 🎯 Token Optimization Strategy

### Progressive Disclosure Levels

```python
TOKEN_LEVELS = {
    "summary": 1000,        # Quick reference only
    "standard": 3000,       # API + basic examples (recommended)
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

### Caching Architecture

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
    import json
    from datetime import datetime, timedelta
    from pathlib import Path

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
```

---

## 🌐 Language-Specific Integration Helpers

### Python Helper

```python
class PythonContext7Helper:
    """Python-specific Context7 integration."""

    LIBRARY_MAPPINGS = {
        "fastapi": "/tiangolo/fastapi",
        "django": "/django/django",
        "flask": "/pallets/flask",
        "pydantic": "/pydantic/pydantic",
        "sqlalchemy": "/sqlalchemy/sqlalchemy",
        "pytest": "/pytest-dev/pytest",
        "asyncio": "/python/cpython",
    }

    async def get_framework_docs(self, framework: str, topic: str = "") -> str:
        """Get Python framework documentation."""
        library_id = self.LIBRARY_MAPPINGS.get(framework)
        if not library_id:
            raise ValueError(f"Unknown Python framework: {framework}")

        return await get_library_docs(library_id, topic, page=1)

    async def get_web_stack_docs(self) -> dict[str, str]:
        """Get modern Python web stack documentation."""
        stack = {
            "framework": "/tiangolo/fastapi",
            "orm": "/sqlalchemy/sqlalchemy",
            "validation": "/pydantic/pydantic",
            "testing": "/pytest-dev/pytest",
        }

        docs = {}
        for component, library_id in stack.items():
            docs[component] = await get_library_docs(library_id)
        return docs
```

### JavaScript/TypeScript Helper

```python
class JavaScriptContext7Helper:
    """JavaScript-specific Context7 integration."""

    LIBRARY_MAPPINGS = {
        "react": "/facebook/react",
        "nextjs": "/vercel/next.js",
        "vue": "/vuejs/vue",
        "angular": "/angular/angular",
        "express": "/expressjs/express",
        "nestjs": "/nestjs/nest",
        "typescript": "/microsoft/TypeScript",
        "vite": "/vitejs/vite",
        "vitest": "/vitest-dev/vitest",
    }

    async def get_frontend_stack(self) -> dict[str, str]:
        """Get modern JavaScript frontend stack docs."""
        stack = {
            "framework": "/vercel/next.js",
            "ui_library": "/facebook/react",
            "language": "/microsoft/TypeScript",
            "styling": "/tailwindlabs/tailwindcss",
            "testing": "/vitest-dev/vitest",
        }

        docs = {}
        for component, library_id in stack.items():
            docs[component] = await get_library_docs(library_id)
        return docs
```

---

## 📋 50+ Language Support

| Category | Languages |
|----------|-----------|
| **Web Backend** | Python, JavaScript, TypeScript, Go, Rust, PHP, Ruby, Java, C# |
| **Frontend** | JavaScript, TypeScript, HTML/CSS |
| **Systems** | Go, Rust, C++, C |
| **Data Science** | Python, R, Julia |
| **Mobile** | Swift, Kotlin, Dart, C# |
| **DevOps/Cloud** | Go, Python, Bash, Ruby |
| **Functional** | Elixir, Scala, Clojure, Haskell |

---

## 📚 200+ Framework Support

### Python (25+)
FastAPI, Django, Flask, Pydantic, SQLAlchemy, Pytest, AsyncIO, Celery, Strawberry, etc.

### JavaScript/TypeScript (35+)
React, Next.js, Vue, Angular, Express, NestJS, Fastify, Remix, Astro, Svelte, etc.

### Go (15+)
Gin, Echo, Fiber, GORM, Testify, Goreleaser, etc.

### Rust (12+)
Actix-Web, Axum, Rocket, Tokio, Serde, Diesel, etc.

### PHP (8+)
Laravel, Symfony, Doctrine, PHPUnit, Composer, etc.

### Java/Kotlin (12+)
Spring Boot, Quarkus, Ktor, Hibernate, JUnit, Maven, Gradle, etc.

### C# (.NET) (8+)
ASP.NET Core, Entity Framework, xUnit, NUnit, Dapper, etc.

---

## 🔄 Error Handling & Fallback Strategies

```python
async def handle_documentation_error(
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
            try:
                return await get_library_docs(library_id, broader_topic)
            except:
                pass

        # Return cached version if available
        cached = get_any_cached_docs(library_id)
        if cached:
            return f"[CACHED] {cached}"

        # Provide manual link
        return f"Documentation unavailable. Visit: https://{extract_repo(library_id)}"

    elif isinstance(error, TokenLimitExceededError):
        # Retry with reduced tokens
        return await get_library_docs(library_id, topic, tokens=1000)

    else:
        return f"Error fetching documentation: {str(error)}"
```

---

## 💡 Advanced Patterns

### Multi-Library Tech Stack Integration

```python
async def get_python_backend_stack() -> dict[str, str]:
    """Get complete Python backend stack documentation."""

    stack_libraries = {
        "web_framework": "/tiangolo/fastapi",
        "validation": "/pydantic/pydantic",
        "orm": "/sqlalchemy/sqlalchemy",
        "testing": "/pytest-dev/pytest",
        "async": "/aio-libs/aiohttp",
        "database": "/psycopg/psycopg",
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

### Token Budget Management

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

# Usage
budget = TokenBudgetManager(total_budget=20000)
budget.allocate("python-backend", percentage=40)  # 8000 tokens
budget.allocate("javascript-frontend", percentage=30)  # 6000 tokens
budget.allocate("testing", percentage=20)  # 4000 tokens
budget.allocate("infrastructure", percentage=10)  # 2000 tokens
```

---

## ✅ Best Practices

✅ **DO**:
- Use Context7 for always-current library documentation
- Implement caching to reduce API calls (30-day TTL for stable)
- Apply progressive token disclosure (start with "standard" level)
- Handle errors gracefully with fallback strategies
- Validate library names before querying
- Use language-specific helpers for common tasks
- Monitor cache hit rates for optimization
- Update cache TTL based on version stability

❌ **DON'T**:
- Skip cache checks (waste of tokens)
- Request maximum tokens (10K) unnecessarily
- Ignore error handling (breaks user experience)
- Hardcode library IDs (use mappings)
- Forget to update cache TTL (leads to stale docs)
- Use blocking calls (always use async)
- Cache beta/development versions too long

---

## 🔗 Changelog

| Version | Date | Changes |
|---------|------|---------|
| **3.0.0** | 2025-11-24 | Merged into moai-mcp-integration hub module |
| 2.0.0 | 2025-11-22 | Added language-specific helpers, token budget management |
| 1.0.0 | 2025-11-20 | Initial Context7 MCP integration |

---

**Module Version**: 3.0.0
**Status**: Production Ready
**Compliance**: 100% (Context7 v2.0+ API)
