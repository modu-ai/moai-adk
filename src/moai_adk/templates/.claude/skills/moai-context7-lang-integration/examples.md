# Context7 Integration Examples

## Quick Start

### Minimal Example: Fetch FastAPI Docs

```python
from moai_context7 import resolve_library_id, get_library_docs

# Step 1: Resolve library name
library_id = resolve_library_id("fastapi")

# Step 2: Fetch documentation
docs = get_library_docs(
    context7_compatible_library_id=library_id,
    tokens=3000
)

# Step 3: Use the documentation
print(docs)
```

**Expected Output**:
- Context7-compatible library ID: `/tiangolo/fastapi` or `/tiangolo/fastapi/v0.115.0`
- Documentation: Up-to-date FastAPI documentation (Markdown format)

---

## Basic Usage

### Example 1: Python Library Documentation

**Goal**: Get latest FastAPI routing documentation

```python
def get_fastapi_docs(topic: str = None, tokens: int = 5000) -> str:
    """Fetch FastAPI documentation with optional topic filtering"""

    from mcp__context7__resolve_library_id import resolve_library_id
    from mcp__context7__get_library_docs import get_library_docs

    # Step 1: Resolve library
    library_id = resolve_library_id("fastapi")

    # Step 2: Fetch documentation
    docs = get_library_docs(
        context7_compatible_library_id=library_id,
        topic=topic,  # e.g., "routing", "dependency-injection"
        tokens=tokens  # Token limit
    )

    return docs

# Usage
fastapi_routing = get_fastapi_docs("routing")
print(fastapi_routing[:500])

# Get dependency injection docs
fastapi_di = get_fastapi_docs("dependency-injection")
print(fastapi_di[:500])
```

### Example 2: JavaScript/TypeScript Library Documentation

**Goal**: Get latest Next.js API routes documentation

```typescript
import { resolveLibraryId, getLibraryDocs } from 'moai-context7-integration';

async function getNextjsDocs(topic?: string): Promise<string> {
    // Step 1: Resolve library
    const libraryId = await resolveLibraryId("next.js");

    // Step 2: Fetch documentation
    const docs = await getLibraryDocs({
        context7CompatibleLibraryID: libraryId,
        topic: topic || "api-routes",
        tokens: 3000
    });

    return docs;
}

// Usage
const apiRoutesDoc = await getNextjsDocs("api-routes");
console.log(apiRoutesDoc);

// Get middleware docs
const middlewareDoc = await getNextjsDocs("middleware");
console.log(middlewareDoc);
```

### Example 3: Go Library Documentation

**Goal**: Get latest Gin web framework documentation

```go
package main

import (
    "fmt"
    "github.com/moai-adk/context7-client"
)

func getGinDocs(topic string) (string, error) {
    // Step 1: Resolve library
    libraryID, err := client.ResolveLibraryID("gin")
    if err != nil {
        return "", err
    }

    // Step 2: Fetch documentation
    docs, err := client.GetLibraryDocs(client.GetLibraryDocsRequest{
        Context7CompatibleLibraryID: libraryID,
        Topic: topic,
        Tokens: 4000,
    })

    return docs, err
}

func main() {
    // Get Gin middleware documentation
    docs, err := getGinDocs("middleware")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    fmt.Println(docs)
}
```

---

## Intermediate Patterns

### Example 4: Error Handling with Fallback

**Goal**: Implement robust error handling with graceful fallbacks

```python
from typing import Optional
from enum import Enum

class LibraryNotFoundError(Exception):
    pass

class DocumentationUnavailableError(Exception):
    pass

def safe_get_docs(
    library_name: str,
    topic: Optional[str] = None,
    max_retries: int = 3
) -> Optional[str]:
    """Get documentation with comprehensive error handling and fallback"""

    # Alternative names for common libraries
    alternative_names = {
        "fastapi": ["FastAPI", "fast-api"],
        "django": ["Django"],
        "pydantic": ["Pydantic"],
        "sqlalchemy": ["SQLAlchemy", "sqlalchemy-orm"]
    }

    # Step 1: Try to resolve library
    names_to_try = [library_name] + alternative_names.get(library_name, [])

    library_id = None
    for alt_name in names_to_try:
        try:
            library_id = resolve_library_id(alt_name)
            print(f"Resolved: {alt_name} -> {library_id}")
            break
        except LibraryNotFoundError:
            continue

    if not library_id:
        # Fallback 2: Return manual link
        return f"Library not found in Context7. See: https://docs.example.com/{library_name}"

    # Step 2: Try to fetch documentation with retries
    for attempt in range(max_retries):
        try:
            docs = get_library_docs(
                context7_compatible_library_id=library_id,
                topic=topic,
                tokens=5000
            )
            if docs:
                return docs
        except DocumentationUnavailableError:
            if attempt < max_retries - 1:
                # Retry with broader topic
                print(f"Attempt {attempt + 1} failed, retrying with broader topic...")
                topic = "api"  # Broader fallback topic
                continue
            else:
                # Last attempt: fetch summary
                try:
                    return get_library_docs(
                        context7_compatible_library_id=library_id,
                        topic="overview",
                        tokens=1000
                    )
                except:
                    break

    # Final fallback: manual link
    return f"Documentation unavailable for {library_name}. See: https://docs.example.com/{library_name}"
```

### Example 5: Caching with TTL Management

**Goal**: Implement intelligent caching with time-based invalidation

```python
from datetime import datetime, timedelta
from typing import Optional, Dict, Any
import json
from pathlib import Path

class DocumentCache:
    def __init__(self, cache_dir: str = ".moai/cache/context7", ttl_days: int = 30):
        self.cache_dir = Path(cache_dir)
        self.cache_dir.mkdir(parents=True, exist_ok=True)
        self.ttl_days = ttl_days
        self.index_file = self.cache_dir / "index.json"

    def _get_cache_key(self, library_id: str, topic: Optional[str], tokens: int) -> str:
        """Generate cache key from parameters"""
        return f"{library_id.replace('/', '_')}_{topic or 'all'}_{tokens}"

    def _is_valid(self, metadata: Dict[str, Any]) -> bool:
        """Check if cached entry is still valid (TTL not expired)"""
        created_at = datetime.fromisoformat(metadata['created_at'])
        expires_at = created_at + timedelta(days=self.ttl_days)
        return datetime.now() < expires_at

    def get(self, library_id: str, topic: Optional[str] = None, tokens: int = 5000) -> Optional[str]:
        """Retrieve documentation from cache if valid"""
        cache_key = self._get_cache_key(library_id, topic, tokens)
        cache_file = self.cache_dir / f"{cache_key}.md"

        if not cache_file.exists():
            return None  # Cache miss

        # Load metadata
        metadata_file = self.cache_dir / f"{cache_key}.json"
        if metadata_file.exists():
            with open(metadata_file) as f:
                metadata = json.load(f)
                if not self._is_valid(metadata):
                    # Cache expired
                    cache_file.unlink()  # Delete expired cache
                    return None

        # Return cached content
        with open(cache_file) as f:
            return f.read()

    def set(self, library_id: str, docs: str, topic: Optional[str] = None, tokens: int = 5000):
        """Store documentation in cache with metadata"""
        cache_key = self._get_cache_key(library_id, topic, tokens)
        cache_file = self.cache_dir / f"{cache_key}.md"
        metadata_file = self.cache_dir / f"{cache_key}.json"

        # Save documentation
        with open(cache_file, 'w') as f:
            f.write(docs)

        # Save metadata
        metadata = {
            'library_id': library_id,
            'topic': topic,
            'tokens': tokens,
            'created_at': datetime.now().isoformat(),
            'version': '1.0'
        }
        with open(metadata_file, 'w') as f:
            json.dump(metadata, f)

    def get_cached_docs(self, library_id: str, topic: Optional[str] = None, tokens: int = 5000) -> str:
        """Get documentation with automatic cache management"""
        # Try cache first
        cached_docs = self.get(library_id, topic, tokens)
        if cached_docs:
            print(f"Cache HIT: {library_id}:{topic}")
            return cached_docs

        # Cache miss: fetch from Context7
        print(f"Cache MISS: {library_id}:{topic}, fetching...")
        docs = get_library_docs(
            context7_compatible_library_id=library_id,
            topic=topic,
            tokens=tokens
        )

        # Store in cache
        self.set(library_id, docs, topic, tokens)
        return docs

# Usage
cache = DocumentCache(ttl_days=30)

# First call: fetches from Context7 and caches
docs1 = cache.get_cached_docs("/tiangolo/fastapi", topic="routing")

# Second call: returns from cache (instant)
docs2 = cache.get_cached_docs("/tiangolo/fastapi", topic="routing")
```

### Example 6: Multiple Libraries Integration

**Goal**: Fetch documentation for entire tech stack

```python
def get_full_stack_docs(stack: str) -> dict:
    """Get documentation for entire tech stack"""

    stacks = {
        "modern-python": [
            {"name": "fastapi", "id": "/tiangolo/fastapi", "topic": "overview"},
            {"name": "pydantic", "id": "/pydantic/pydantic", "topic": "validation"},
            {"name": "sqlalchemy", "id": "/sqlalchemy/sqlalchemy", "topic": "orm"}
        ],
        "modern-js": [
            {"name": "next.js", "id": "/vercel/next.js", "topic": "routing"},
            {"name": "react", "id": "/facebook/react", "topic": "hooks"},
            {"name": "typescript", "id": "/microsoft/TypeScript", "topic": "types"}
        ],
        "modern-go": [
            {"name": "gin", "id": "/gin-gonic/gin", "topic": "middleware"},
            {"name": "gorm", "id": "/go-gorm/gorm", "topic": "models"},
            {"name": "cobra", "id": "/spf13/cobra", "topic": "commands"}
        ]
    }

    libraries = stacks.get(stack, [])
    results = {}
    total_tokens_used = 0

    for lib in libraries:
        print(f"Fetching {lib['name']}...")
        try:
            docs = get_library_docs(
                context7_compatible_library_id=lib['id'],
                topic=lib['topic'],
                tokens=2000  # Shared token budget
            )
            results[lib['name']] = {
                'status': 'success',
                'docs': docs,
                'tokens_used': 2000
            }
            total_tokens_used += 2000
        except Exception as e:
            results[lib['name']] = {
                'status': 'error',
                'error': str(e),
                'tokens_used': 0
            }

    print(f"Total tokens used: {total_tokens_used}")
    return results

# Usage
stack_docs = get_full_stack_docs("modern-python")
for lib_name, result in stack_docs.items():
    if result['status'] == 'success':
        print(f"{lib_name}: {len(result['docs'])} chars")
    else:
        print(f"{lib_name}: ERROR - {result['error']}")
```

---

## Advanced Patterns

### Example 7: Async Parallel Fetching

**Goal**: Fetch multiple libraries concurrently to reduce latency

```python
import asyncio
from typing import List, Dict

async def fetch_library_docs_async(library_id: str, topic: Optional[str] = None) -> str:
    """Async fetch with timeout protection"""
    try:
        # In production, use async Context7 client
        docs = await asyncio.to_thread(
            get_library_docs,
            library_id,
            None,  # topic
            4000   # tokens
        )
        return docs
    except asyncio.TimeoutError:
        return ""
    except Exception as e:
        print(f"Error fetching {library_id}: {e}")
        return ""

async def fetch_multiple_libraries(libraries: List[Dict]) -> Dict[str, str]:
    """Fetch multiple libraries concurrently"""

    tasks = [
        fetch_library_docs_async(
            lib['id'],
            lib.get('topic')
        )
        for lib in libraries
    ]

    # Execute all tasks concurrently with timeout
    try:
        results = await asyncio.wait_for(
            asyncio.gather(*tasks, return_exceptions=True),
            timeout=30.0  # 30 second timeout for all requests
        )
    except asyncio.TimeoutError:
        print("Parallel fetch timed out after 30 seconds")
        results = ["" for _ in libraries]

    return {
        lib['name']: result
        for lib, result in zip(libraries, results)
        if isinstance(result, str)
    }

# Usage
libraries = [
    {'name': 'fastapi', 'id': '/tiangolo/fastapi', 'topic': 'routing'},
    {'name': 'pydantic', 'id': '/pydantic/pydantic', 'topic': 'validation'},
    {'name': 'sqlalchemy', 'id': '/sqlalchemy/sqlalchemy', 'topic': 'orm'}
]

# Sequential: ~9 seconds (3 libs * 3 seconds each)
# Parallel: ~3 seconds (concurrent)
results = asyncio.run(fetch_multiple_libraries(libraries))

for lib_name, docs in results.items():
    print(f"{lib_name}: {len(docs)} characters fetched")
```

### Example 8: Token Budget Management

**Goal**: Allocate and track tokens across multiple concurrent operations

```python
from dataclasses import dataclass, field
from typing import Dict

@dataclass
class SkillTokenAllocation:
    skill_name: str
    allocated_tokens: int
    used_tokens: int = 0

    @property
    def remaining_tokens(self) -> int:
        return self.allocated_tokens - self.used_tokens

    @property
    def usage_percentage(self) -> float:
        return (self.used_tokens / self.allocated_tokens * 100) if self.allocated_tokens > 0 else 0.0

class TokenBudgetManager:
    def __init__(self, total_budget: int = 20000):
        self.total_budget = total_budget
        self.allocations: Dict[str, SkillTokenAllocation] = {}

    def allocate(self, skill_name: str, percentage: int) -> int:
        """Allocate percentage of total budget to skill"""
        if sum(a.allocated_tokens for a in self.allocations.values()) + \
           int(self.total_budget * percentage / 100) > self.total_budget:
            raise ValueError(f"Budget overflow: {percentage}% exceeds available")

        tokens = int(self.total_budget * percentage / 100)
        self.allocations[skill_name] = SkillTokenAllocation(skill_name, tokens)
        return tokens

    def get_available(self, skill_name: str) -> int:
        """Get remaining tokens for skill"""
        if skill_name not in self.allocations:
            return 0
        return self.allocations[skill_name].remaining_tokens

    def use_tokens(self, skill_name: str, amount: int):
        """Deduct tokens from allocation"""
        if skill_name not in self.allocations:
            raise KeyError(f"Skill {skill_name} not allocated")

        allocation = self.allocations[skill_name]
        if amount > allocation.remaining_tokens:
            raise ValueError(
                f"Budget exceeded for {skill_name}: "
                f"requested {amount}, available {allocation.remaining_tokens}"
            )

        allocation.used_tokens += amount

    def get_stats(self) -> Dict:
        """Get budget usage statistics"""
        return {
            'total_budget': self.total_budget,
            'allocated': sum(a.allocated_tokens for a in self.allocations.values()),
            'used': sum(a.used_tokens for a in self.allocations.values()),
            'skills': {
                name: {
                    'allocated': a.allocated_tokens,
                    'used': a.used_tokens,
                    'remaining': a.remaining_tokens,
                    'usage_pct': a.usage_percentage
                }
                for name, a in self.allocations.items()
            }
        }

# Usage
budget = TokenBudgetManager(total_budget=20000)
budget.allocate("moai-lang-python", 40)      # 8000 tokens
budget.allocate("moai-lang-typescript", 35)  # 7000 tokens
budget.allocate("moai-lang-go", 25)          # 5000 tokens

# When fetching docs
available = budget.get_available("moai-lang-python")
tokens_to_use = min(5000, available)  # Don't exceed allocated
docs = get_library_docs(..., tokens=tokens_to_use)
budget.use_tokens("moai-lang-python", tokens_to_use)

# Check stats
stats = budget.get_stats()
print(f"Total used: {stats['used']}/{stats['total_budget']} tokens")
```

### Example 9: Production Monitoring

**Goal**: Monitor and optimize documentation fetch performance

```python
from dataclasses import dataclass
from datetime import datetime
from typing import List
import logging

@dataclass
class DocumentFetchMetric:
    library_name: str
    library_id: str
    topic: Optional[str]
    fetch_time_ms: float
    tokens_used: int
    cache_hit: bool
    success: bool
    error_message: Optional[str] = None
    timestamp: datetime = field(default_factory=datetime.now)

class DocumentMonitor:
    def __init__(self, logger_name: str = "document-monitor"):
        self.metrics: List[DocumentFetchMetric] = []
        self.logger = logging.getLogger(logger_name)

    def log_fetch(self, metric: DocumentFetchMetric):
        """Log fetch operation with metrics"""
        self.metrics.append(metric)

        status = "HIT" if metric.cache_hit else "MISS"
        result = "OK" if metric.success else f"ERROR: {metric.error_message}"

        self.logger.info(
            f"[{metric.library_name}] "
            f"time={metric.fetch_time_ms:.0f}ms, "
            f"tokens={metric.tokens_used}, "
            f"cache={status}, "
            f"result={result}"
        )

    def get_stats(self) -> Dict:
        """Calculate performance statistics"""
        if not self.metrics:
            return {}

        total = len(self.metrics)
        cache_hits = sum(1 for m in self.metrics if m.cache_hit)
        successful = sum(1 for m in self.metrics if m.success)

        fetch_times = [m.fetch_time_ms for m in self.metrics if m.success]
        avg_time = sum(fetch_times) / len(fetch_times) if fetch_times else 0

        total_tokens = sum(m.tokens_used for m in self.metrics)

        return {
            'total_fetches': total,
            'successful': successful,
            'failed': total - successful,
            'success_rate': successful / total,
            'cache_hit_rate': cache_hits / total,
            'avg_fetch_time_ms': avg_time,
            'total_tokens_used': total_tokens,
            'avg_tokens_per_fetch': total_tokens / total if total > 0 else 0
        }

    def get_optimization_recommendations(self) -> List[str]:
        """Provide optimization recommendations based on metrics"""
        stats = self.get_stats()
        recommendations = []

        if stats.get('cache_hit_rate', 0) < 0.5:
            recommendations.append("Cache hit rate < 50%: Consider caching more aggressively")

        if stats.get('avg_fetch_time_ms', 0) > 2000:
            recommendations.append("Average fetch time > 2s: Consider parallel requests")

        if stats.get('avg_tokens_per_fetch', 0) > 5000:
            recommendations.append("High token usage per fetch: Consider topic-specific queries")

        if stats.get('success_rate', 1.0) < 0.95:
            recommendations.append("Success rate < 95%: Review error handling")

        return recommendations

# Usage
monitor = DocumentMonitor()

# Track fetch with cache hit
import time
start = time.time()
docs = cache.get_cached_docs("/tiangolo/fastapi")  # Returns from cache
elapsed_ms = (time.time() - start) * 1000

metric = DocumentFetchMetric(
    library_name="fastapi",
    library_id="/tiangolo/fastapi",
    topic=None,
    fetch_time_ms=elapsed_ms,
    tokens_used=0,  # No tokens used (cache hit)
    cache_hit=True,
    success=True
)
monitor.log_fetch(metric)

# Get recommendations
stats = monitor.get_stats()
print(f"Cache hit rate: {stats['cache_hit_rate']:.1%}")

recommendations = monitor.get_optimization_recommendations()
for rec in recommendations:
    print(f"  - {rec}")
```

---

## Migration Guides

### From Manual Documentation Links to Context7

**Before** (Old Pattern):
```markdown
## References
- [FastAPI Docs](https://fastapi.tiangolo.com)
- [Pydantic Docs](https://docs.pydantic.dev)
- [SQLAlchemy Docs](https://docs.sqlalchemy.org)

Status: Manual links, outdated, may break, no version sync
```

**After** (Context7 Pattern):
```markdown
## References

For up-to-date documentation:
Skill("moai-context7-lang-integration")

**Get latest docs**:
- FastAPI: `get_library_docs("/tiangolo/fastapi")`
- Pydantic: `get_library_docs("/pydantic/pydantic")`
- SQLAlchemy: `get_library_docs("/sqlalchemy/sqlalchemy")`

Status: Automatic updates, always latest version, no broken links
```

### Version-Specific Documentation

**Pattern**: Pin to specific versions when needed

```python
# General usage: Fetch latest version
library_id = resolve_library_id("fastapi")
# Returns: /tiangolo/fastapi/v0.115.0 (latest at fetch time)

# Production usage: Pin to specific version
library_id = "/tiangolo/fastapi/v0.100.0"  # Locked version
docs = get_library_docs(context7_compatible_library_id=library_id)

# Benefits of version pinning:
# - Ensures consistent documentation across team
# - Avoids breaking changes from version updates
# - Supports legacy framework documentation
# - Simplifies version compatibility testing
```

