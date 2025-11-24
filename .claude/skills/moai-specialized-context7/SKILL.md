---
name: moai-specialized-context7
description: Context7 MCP library documentation integration for real-time API documentation
version: 1.0.0
modularized: false
last_updated: 2025-11-24
allowed_tools: [Context7, Skill, Task]
compliance_score: 100%
category_tier: 6
auto_trigger_keywords: [context7, mcp, library, documentation, api, patterns, integration, real-time]
agent_coverage: [tdd-implementer, quality-gate, docs-manager]
context7_references: [/modelcontextprotocol/servers-context7, /upstash/context7-mcp]
invocation_api_version: 1.0
dependencies: [moai-foundation-trust, moai-context7-integration]
deprecated: false
modules: null
successor: null
---

# moai-specialized-context7: Real-Time API Documentation

## Quick Reference (Level 1)

**Context7 MCP Integration**: Fetch live library documentation from Context7 MCP servers for accurate, up-to-date API references without hallucinations.

**Key Capabilities**:
- Real-time library ID resolution via Context7
- Context-aware documentation retrieval
- Automatic fallback to static documentation
- Version compatibility checking
- Multi-library orchestration

**When to Use**: Every code generation task involving external libraries. Always invoke Context7 before writing code to ensure API accuracy.

---

## Implementation Guide (Level 2)

### Library ID Resolution

```python
# Resolve library names to Context7-compatible IDs
async def resolve_library_id(library_name: str) -> str:
    resolver = mcp__context7__resolve_library_id(library_name)
    return resolver['library_id']  # Returns: /org/project/version
```

### Documentation Retrieval

```python
# Get latest documentation with fallback
async def get_docs(lib_id: str, topic: str, tokens: int = 5000) -> dict:
    try:
        docs = await mcp__context7__get_library_docs(
            context7CompatibleLibraryID=lib_id,
            topic=topic,
            page=1
        )
        return docs
    except Exception:
        # Fallback to cached/static docs
        return load_static_docs(lib_id)
```

### Common Patterns

**Next.js Documentation**:
```python
lib_id = "/vercel/next.js"  # Auto-resolves to latest version
docs = await get_docs(lib_id, "app-router server-components")
```

**FastAPI Documentation**:
```python
lib_id = "/tiangolo/fastapi"  # Latest FastAPI patterns
docs = await get_docs(lib_id, "dependency-injection async-support")
```

**Database Libraries**:
```python
lib_id = "/sqlalchemy/sqlalchemy"  # SQLAlchemy 2.0 patterns
docs = await get_docs(lib_id, "orm async connection-pooling")
```

---

## Advanced Patterns (Level 3)

### Multi-Library Orchestration

Coordinate multiple library contexts for full-stack development:

```python
async def full_stack_context():
    frontend = await get_docs("/vercel/next.js", "react server-components")
    backend = await get_docs("/tiangolo/fastapi", "async api development")
    database = await get_docs("/sqlalchemy/sqlalchemy", "async orm patterns")
    return {frontend, backend, database}
```

### Caching Strategy

Implement intelligent caching to reduce redundant Context7 calls:

```python
cache = {
    'ttl': 3600,  # 1 hour
    'max_size': 100,  # MB
    'eviction': 'lru'
}
```

### Error Handling

Graceful degradation when Context7 is unavailable:

```python
def get_docs_with_fallback(lib_id: str, topic: str):
    try:
        return context7_get_docs(lib_id, topic)
    except TimeoutError:
        return static_docs[lib_id]
    except NotFoundError:
        return best_guess_docs(lib_id, topic)
```

---

## Usage Rules

1. **Always invoke Context7** before implementing code with external libraries
2. **Pass topic keywords** for accuracy (e.g., "async support", "dependency injection")
3. **Handle timeouts gracefully** with fallback mechanisms
4. **Cache results** to minimize API calls
5. **Update docs monthly** for latest patterns

---

**Status**: Production Ready
**Enhanced with**: Context7 MCP real-time documentation
**Auto-triggers**: Library usage detection, API implementation tasks
