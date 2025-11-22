# Context7 Language Integration Examples (2025 Edition)

## Complete Integration Examples

### Example 1: Python FastAPI Integration

**Scenario**: Get FastAPI routing and dependency injection documentation

```python
# Step 1: Resolve library name
library_id = await mcp__context7__resolve_library_id("fastapi")
# Returns: /tiangolo/fastapi

# Step 2: Get routing documentation
routing_docs = await mcp__context7__get_library_docs(
    context7CompatibleLibraryID=library_id,
    topic="routing dependency-injection path-parameters",
    page=1
)

# Result: Markdown documentation
print(routing_docs)
# # FastAPI Routing
# 
# ## Path Parameters
# FastAPI supports path parameters with type validation:
# ```python
# @app.get("/items/{item_id}")
# async def read_item(item_id: int):
#     return {"item_id": item_id}
# ```
# 
# ## Dependency Injection
# Use `Depends()` for dependency injection:
# ...
```

### Example 2: React Hooks Documentation

**Scenario**: Get React hooks and performance patterns

```python
# Resolve React library
library_id = await mcp__context7__resolve_library_id("react")
# Returns: /facebook/react

# Get hooks documentation
hooks_docs = await mcp__context7__get_library_docs(
    context7CompatibleLibraryID=library_id,
    topic="hooks useState useEffect useContext performance",
    page=1
)

# Result: Comprehensive React hooks guide
print(hooks_docs)
# # React Hooks
# 
# ## useState
# Manage component state:
# ```javascript
# const [count, setCount] = useState(0);
# ```
# 
# ## useEffect
# Handle side effects:
# ```javascript
# useEffect(() => {
#   document.title = `Count: ${count}`;
# }, [count]);
# ```
# ...
```

### Example 3: Multi-Library Tech Stack

**Scenario**: Fetch documentation for complete Python web stack

```python
async def get_python_web_stack_docs() -> dict[str, str]:
    """Get documentation for modern Python web stack."""
    
    # Define stack components
    stack = {
        "web_framework": ("fastapi", "routing dependency-injection"),
        "validation": ("pydantic", "models validation"),
        "orm": ("sqlalchemy", "relationships queries"),
        "testing": ("pytest", "fixtures parametrize"),
        "async": ("aiohttp", "client server"),
    }
    
    docs = {}
    for component, (library_name, topic) in stack.items():
        # Resolve library
        library_id = await mcp__context7__resolve_library_id(library_name)
        
        # Fetch documentation
        docs[component] = await mcp__context7__get_library_docs(
            context7CompatibleLibraryID=library_id,
            topic=topic,
            page=1
        )
    
    return docs

# Usage
stack_docs = await get_python_web_stack_docs()
print(stack_docs["web_framework"])  # FastAPI routing docs
print(stack_docs["validation"])     # Pydantic models docs
print(stack_docs["orm"])            # SQLAlchemy relationships docs
```

### Example 4: JavaScript Frontend Stack

**Scenario**: Fetch documentation for React + Next.js + TypeScript stack

```python
async def get_javascript_frontend_stack() -> dict[str, str]:
    """Get documentation for modern JavaScript frontend stack."""
    
    stack = {
        "framework": ("nextjs", "app-router server-components"),
        "ui_library": ("react", "hooks components"),
        "type_system": ("typescript", "types interfaces"),
        "styling": ("tailwindcss", "utilities responsive"),
        "testing": ("vitest", "unit-testing mocking"),
    }
    
    docs = {}
    for component, (library_name, topic) in stack.items():
        library_id = await mcp__context7__resolve_library_id(library_name)
        docs[component] = await mcp__context7__get_library_docs(
            library_id,
            topic=topic,
            page=1
        )
    
    return docs

# Usage
frontend_docs = await get_javascript_frontend_stack()
print(frontend_docs["framework"])   # Next.js App Router docs
print(frontend_docs["ui_library"])  # React hooks docs
print(frontend_docs["styling"])     # Tailwind CSS utilities docs
```

## Error Handling Examples

### Example 5: Library Not Found Recovery

**Scenario**: Handle library not found with alternatives

```python
async def get_docs_with_fallback(library_name: str, topic: str) -> str:
    """Get documentation with fallback strategies."""
    
    try:
        # Try primary library name
        library_id = await mcp__context7__resolve_library_id(library_name)
        return await mcp__context7__get_library_docs(library_id, topic)
    
    except LibraryNotFoundError:
        # Try alternative names
        alternatives = {
            "FastAPI": "fastapi",
            "fast-api": "fastapi",
            "Next.js": "nextjs",
            "next": "nextjs",
        }
        
        alt_name = alternatives.get(library_name)
        if alt_name:
            try:
                library_id = await mcp__context7__resolve_library_id(alt_name)
                return await mcp__context7__get_library_docs(library_id, topic)
            except:
                pass
        
        # Return guidance
        return f"Library '{library_name}' not found. Try: {', '.join(alternatives.values())}"

# Usage
docs = await get_docs_with_fallback("FastAPI", "routing")  # Works with capitalization
docs = await get_docs_with_fallback("fast-api", "routing")  # Works with hyphen
```

### Example 6: Token Budget Management

**Scenario**: Manage token allocation across multiple skills

```python
class TokenBudgetManager:
    """Manage Context7 token budget."""
    
    def __init__(self, total_budget: int = 20000):
        self.total_budget = total_budget
        self.allocated = {}
        self.used = {}
    
    def allocate(self, skill_name: str, percentage: int) -> int:
        """Allocate percentage of budget to skill."""
        tokens = int(self.total_budget * percentage / 100)
        self.allocated[skill_name] = tokens
        self.used[skill_name] = 0
        return tokens
    
    def use_tokens(self, skill_name: str, amount: int):
        """Record token usage."""
        available = self.allocated[skill_name] - self.used[skill_name]
        if amount > available:
            raise TokenBudgetExceededError(
                f"{skill_name} exceeded budget: {amount} > {available}"
            )
        self.used[skill_name] += amount
    
    def get_report(self) -> dict:
        """Generate usage report."""
        return {
            skill: {
                "allocated": self.allocated[skill],
                "used": self.used[skill],
                "remaining": self.allocated[skill] - self.used[skill],
                "utilization": self.used[skill] / self.allocated[skill]
            }
            for skill in self.allocated
        }

# Usage
budget = TokenBudgetManager(total_budget=20000)
budget.allocate("python-backend", 40)   # 8000 tokens
budget.allocate("javascript-frontend", 30)  # 6000 tokens
budget.allocate("testing", 20)         # 4000 tokens
budget.allocate("infrastructure", 10)  # 2000 tokens

# Use tokens
budget.use_tokens("python-backend", 3000)
budget.use_tokens("javascript-frontend", 2000)

# Get report
report = budget.get_report()
print(report["python-backend"])
# {
#   "allocated": 8000,
#   "used": 3000,
#   "remaining": 5000,
#   "utilization": 0.375
# }
```

## Caching Examples

### Example 7: Intelligent Caching

**Scenario**: Cache documentation with TTL-based invalidation

```python
import hashlib
import json
from datetime import datetime, timedelta
from pathlib import Path

class Context7Cache:
    """Intelligent caching for Context7 documentation."""
    
    def __init__(self, cache_dir: str = ".moai/cache/context7/"):
        self.cache_dir = Path(cache_dir)
        self.cache_dir.mkdir(parents=True, exist_ok=True)
        self.index_path = self.cache_dir / "index.json"
    
    def generate_key(self, library_id: str, topic: str, page: int) -> str:
        """Generate cache key."""
        key_string = f"{library_id}:{topic}:{page}"
        return hashlib.sha256(key_string.encode()).hexdigest()[:16]
    
    def get(self, library_id: str, topic: str, page: int) -> Optional[str]:
        """Retrieve from cache if valid."""
        key = self.generate_key(library_id, topic, page)
        cache_path = self.cache_dir / f"{key}.json"
        
        if not cache_path.exists():
            return None
        
        # Load cache entry
        entry = json.loads(cache_path.read_text())
        
        # Check expiration
        expires_at = datetime.fromisoformat(entry["expires_at"])
        if datetime.now() > expires_at:
            cache_path.unlink()  # Remove expired
            return None
        
        return entry["content"]
    
    def set(self, library_id: str, topic: str, page: int, content: str, ttl_days: int = 30):
        """Store in cache with TTL."""
        key = self.generate_key(library_id, topic, page)
        cache_path = self.cache_dir / f"{key}.json"
        
        entry = {
            "library_id": library_id,
            "topic": topic,
            "page": page,
            "content": content,
            "cached_at": datetime.now().isoformat(),
            "expires_at": (datetime.now() + timedelta(days=ttl_days)).isoformat(),
            "ttl_days": ttl_days
        }
        
        cache_path.write_text(json.dumps(entry))

# Usage
cache = Context7Cache()

# Try cache first
docs = cache.get("/tiangolo/fastapi", "routing", 1)
if not docs:
    # Cache miss, fetch from Context7
    docs = await mcp__context7__get_library_docs("/tiangolo/fastapi", "routing", 1)
    cache.set("/tiangolo/fastapi", "routing", 1, docs, ttl_days=30)
```

### Example 8: Progressive Token Disclosure

**Scenario**: Start with minimal tokens, expand as needed

```python
async def get_docs_progressive(library_id: str, topic: str) -> str:
    """Get documentation with progressive token disclosure."""
    
    # Level 1: Summary (1000 tokens)
    summary = await mcp__context7__get_library_docs(
        library_id,
        topic=topic,
        page=1
    )
    
    # Check if summary sufficient
    if is_sufficient(summary):
        return summary
    
    # Level 2: Standard (3000 tokens)
    standard = await mcp__context7__get_library_docs(
        library_id,
        topic=topic,
        page=2
    )
    
    if is_sufficient(summary + standard):
        return summary + "\n\n" + standard
    
    # Level 3: Detailed (5000 tokens)
    detailed = await mcp__context7__get_library_docs(
        library_id,
        topic=topic,
        page=3
    )
    
    return summary + "\n\n" + standard + "\n\n" + detailed

def is_sufficient(content: str) -> bool:
    """Check if documentation is sufficient."""
    # Simple heuristic: check length and completeness
    return len(content) > 500 and "example" in content.lower()

# Usage
docs = await get_docs_progressive("/tiangolo/fastapi", "routing")
# Automatically fetches only necessary pages
```

## Language-Specific Helper Examples

### Example 9: Python Helper

**Scenario**: Simplified Python library documentation access

```python
class PythonContext7Helper:
    """Python-specific Context7 helper."""
    
    LIBRARIES = {
        "fastapi": "/tiangolo/fastapi",
        "django": "/django/django",
        "pydantic": "/pydantic/pydantic",
        "sqlalchemy": "/sqlalchemy/sqlalchemy",
        "pytest": "/pytest-dev/pytest",
    }
    
    async def get_framework_docs(self, framework: str, topic: str = "") -> str:
        """Get Python framework documentation."""
        library_id = self.LIBRARIES.get(framework)
        if not library_id:
            raise ValueError(f"Unknown framework: {framework}")
        
        return await mcp__context7__get_library_docs(library_id, topic)
    
    async def get_testing_guide(self) -> str:
        """Get pytest testing guide."""
        return await self.get_framework_docs("pytest", "fixtures parametrize mocking")
    
    async def get_async_patterns(self) -> str:
        """Get async patterns."""
        return await self.get_framework_docs("fastapi", "async background-tasks")

# Usage
python_helper = PythonContext7Helper()
fastapi_docs = await python_helper.get_framework_docs("fastapi", "routing")
testing_docs = await python_helper.get_testing_guide()
```

### Example 10: JavaScript Helper

**Scenario**: Simplified JavaScript library documentation access

```python
class JavaScriptContext7Helper:
    """JavaScript-specific Context7 helper."""
    
    LIBRARIES = {
        "react": "/facebook/react",
        "nextjs": "/vercel/next.js",
        "express": "/expressjs/express",
        "typescript": "/microsoft/TypeScript",
        "vitest": "/vitest-dev/vitest",
    }
    
    async def get_framework_docs(self, framework: str, topic: str = "") -> str:
        """Get JavaScript framework documentation."""
        library_id = self.LIBRARIES.get(framework)
        if not library_id:
            raise ValueError(f"Unknown framework: {framework}")
        
        return await mcp__context7__get_library_docs(library_id, topic)
    
    async def get_react_patterns(self) -> str:
        """Get React patterns."""
        return await self.get_framework_docs("react", "hooks patterns performance")
    
    async def get_nextjs_guide(self) -> str:
        """Get Next.js App Router guide."""
        return await self.get_framework_docs("nextjs", "app-router server-components")

# Usage
js_helper = JavaScriptContext7Helper()
react_docs = await js_helper.get_react_patterns()
nextjs_docs = await js_helper.get_nextjs_guide()
```

---

**Last Updated**: 2025-11-22  
**Maintained by**: MoAI-ADK Team
