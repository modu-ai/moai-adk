# moai-context7-integration - Reference Documentation

## Table of Contents

- [MCP Tool Reference](#mcp-tool-reference)
- [API Documentation](#api-documentation)
- [Library ID Mapping](#library-id-mapping)
- [Configuration Options](#configuration-options)
- [Error Handling](#error-handling)
- [Performance Optimization](#performance-optimization)

---

## MCP Tool Reference

### mcp__context7__resolve-library-id

**Purpose**: Resolve library/product name to Context7-compatible library ID.

**Parameters**:
- `libraryName` (string, required): The name of the library to resolve

**Returns**:
```json
{
  "library_id": "/facebook/react",
  "name": "React",
  "description": "A JavaScript library for building user interfaces",
  "version": "18.2.0",
  "confidence": 0.95,
  "alternatives": [
    "/reactnative/react",
    "/preactjs/preact"
  ]
}
```

**Example Usage**:
```python
# Resolve React library
result = mcp__context7__resolve-library_id(libraryName="React")
library_id = result["library_id"]  # "/facebook/react"

# Resolve with partial name
result = mcp__context7__resolve-library-id(libraryName="fastapi")
library_id = result["library_id"]  # "/tiangolo/fastapi"

# Resolve with alternative
result = mcp__context7__resolve-library-id(libraryName="vue")
library_id = result["library_id"]  # "/vuejs/vue"
```

### mcp__context7__get-library-docs

**Purpose**: Fetch up-to-date documentation for a library.

**Parameters**:
- `context7CompatibleLibraryID` (string, required): Exact Context7-compatible library ID
- `topic` (string, optional): Focus documentation on specific topic
- `tokens` (number, optional): Maximum tokens of documentation to retrieve (default: 5000)

**Returns**:
```json
{
  "library_id": "/facebook/react",
  "topic": "hooks",
  "content": {
    "overview": "React Hooks are functions that let you use state...",
    "examples": [
      {
        "title": "useState Hook",
        "code": "import { useState } from 'react';\nfunction Counter() { const [count, setCount] = useState(0); return ...; }",
        "description": "Basic state management in functional components",
        "language": "javascript"
      }
    ],
    "best_practices": [...],
    "common_patterns": [...],
    "api_reference": {
      "functions": [...],
      "types": [...],
      "constants": [...]
    },
    "related_topics": ["useEffect", "useContext", "custom-hooks"],
    "troubleshooting": [...]
  },
  "metadata": {
    "version": "18.2.0",
    "last_updated": "2025-11-12",
    "source": "official_docs",
    "tokens_used": 4832
  }
}
```

**Example Usage**:
```python
# Get React hooks documentation
docs = mcp__context7__get-library-docs(
    context7CompatibleLibraryID="/facebook/react",
    topic="hooks"
)

# Get Django ORM documentation
docs = mcp__context7__get-library-docs(
    context7CompatibleLibraryID="/django/django",
    topic="orm",
    tokens=3000
)

# Get FastAPI getting started
docs = mcp__context7__get-library-docs(
    context7CompatibleLibraryID="/tiangolo/fastapi",
    topic="getting-started"
)
```

---

## API Documentation

### Context7Client Class

```python
from context7 import Context7Client

class Context7Client:
    """Client for interacting with Context7 MCP service."""
    
    def __init__(self, api_key: str = None, base_url: str = None):
        """
        Initialize Context7 client.
        
        Args:
            api_key: Context7 API key (optional, uses environment variable)
            base_url: Custom base URL (optional)
        """
        pass
    
    def resolve_library_id(self, library_name: str) -> str:
        """
        Resolve library name to Context7 ID.
        
        Args:
            library_name: Name of the library to resolve
            
        Returns:
            Context7-compatible library ID
            
        Raises:
            LibraryNotFoundError: If library not found
            APIError: If API request fails
        """
        pass
    
    def get_library_docs(self, library_id: str, topic: str = None, tokens: int = 5000) -> dict:
        """
        Get library documentation.
        
        Args:
            library_id: Context7-compatible library ID
            topic: Specific topic to focus on
            tokens: Maximum tokens to retrieve
            
        Returns:
            Documentation content dictionary
            
        Raises:
            InvalidLibraryIDError: If library ID is invalid
            DocumentationNotFoundError: If documentation not found
            APIError: If API request fails
        """
        pass
    
    def search_libraries(self, query: str, language: str = None) -> list:
        """
        Search for libraries by query.
        
        Args:
            query: Search query
            language: Programming language filter
            
        Returns:
            List of matching libraries
        """
        pass
```

### Error Classes

```python
class Context7Error(Exception):
    """Base class for Context7 errors."""
    pass

class LibraryNotFoundError(Context7Error):
    """Raised when library is not found."""
    pass

class InvalidLibraryIDError(Context7Error):
    """Raised when library ID is invalid."""
    pass

class DocumentationNotFoundError(Context7Error):
    """Raised when documentation is not found."""
    pass

class APIError(Context7Error):
    """Raised when API request fails."""
    pass

class RateLimitError(Context7Error):
    """Raised when rate limit is exceeded."""
    pass
```

---

## Library ID Mapping

### Popular Library IDs

#### Frontend Frameworks

| Library | Context7 ID | Description |
|---------|-------------|-------------|
| React | `/facebook/react` | A JavaScript library for building user interfaces |
| Vue.js | `/vuejs/vue` | Progressive JavaScript framework |
| Angular | `/angular/angular` | Platform for building mobile and desktop apps |
| Svelte | `/sveltejs/svelte` | Cybernetically enhanced web apps |
| Next.js | `/vercel/next.js` | React framework for production |
| Nuxt.js | `/nuxt/nuxt` | Vue.js framework for production |

#### Backend Frameworks

| Library | Context7 ID | Description |
|---------|-------------|-------------|
| Django | `/django/django` | Web framework for perfectionists |
| FastAPI | `/tiangolo/fastapi` | Modern, fast web framework for building APIs |
| Flask | `/pallets/flask` | Lightweight WSGI web application framework |
| Express.js | `/expressjs/express` | Fast, unopinionated, minimalist web framework |
| NestJS | `/nestjs/nest` | Progressive Node.js framework |

#### Database & ORM

| Library | Context7 ID | Description |
|---------|-------------|-------------|
| SQLAlchemy | `/sqlalchemy/sqlalchemy` | Python SQL toolkit and ORM |
| Prisma | `/prisma/prisma` | Next-generation ORM |
| TypeORM | `/typeorm/typeorm` | ORM for TypeScript and JavaScript |
| Mongoose | `/automattic/mongoose` | MongoDB object modeling tool |

#### Testing

| Library | Context7 ID | Description |
|---------|-------------|-------------|
| Jest | `/facebook/jest` | Delightful JavaScript Testing |
| Pytest | `/pytest-dev/pytest` | Python testing framework |
| Cypress | `/cypress-io/cypress` | Fast, easy and reliable testing |
| Playwright | `/microsoft/playwright` | Node.js library to automate Chromium |

### Language-Specific Patterns

#### Python Libraries

```python
# Python libraries follow pattern: /organization/repository
"/django/django"          # Django
"/tiangolo/fastapi"      # FastAPI
"/pallets/flask"         # Flask
"/pytest-dev/pytest"     # Pytest
"/sqlalchemy/sqlalchemy" # SQLAlchemy
```

#### JavaScript/TypeScript Libraries

```python
# JavaScript libraries follow pattern: /organization/repository
"/facebook/react"        # React
"/facebook/jest"         # Jest
"/vuejs/vue"            # Vue.js
"/sveltejs/svelte"       # Svelte
"/microsoft/playwright"  # Playwright
```

#### Go Libraries

```python
# Go libraries follow pattern: /organization/repository
"/golang/go"             # Go language
"/gin-gonic/gin"         # Gin web framework
"/grpc/grpc-go"          # gRPC for Go
"/go-swagger/go-swagger" # Swagger for Go
```

---

## Configuration Options

### Environment Variables

```bash
# Context7 API Configuration
CONTEXT7_API_KEY="your-api-key"
CONTEXT7_BASE_URL="https://api.context7.io"
CONTEXT7_TIMEOUT=30
CONTEXT7_CACHE_TTL=3600

# Cache Configuration
CONTEXT7_CACHE_DIR=".context7_cache"
CONTEXT7_CACHE_ENABLED=true

# Rate Limiting
CONTEXT7_RATE_LIMIT=100
CONTEXT7_RATE_WINDOW=60
```

### Client Configuration

```python
# Initialize client with custom configuration
client = Context7Client(
    api_key="your-api-key",
    base_url="https://custom-context7.example.com",
    timeout=30,
    cache_ttl=3600,
    rate_limit=100
)

# Configure caching
client.configure_cache(
    enabled=True,
    directory=".context7_cache",
    ttl=3600
)

# Configure rate limiting
client.configure_rate_limiting(
    max_requests=100,
    window_seconds=60
)
```

### Alfred Integration Configuration

```json
{
  "context7": {
    "enabled": true,
    "api_key": "${CONTEXT7_API_KEY}",
    "cache_enabled": true,
    "cache_ttl": 3600,
    "default_tokens": 5000,
    "max_tokens": 10000,
    "rate_limit": {
      "requests_per_minute": 100,
      "burst_size": 20
    }
  }
}
```

---

## Error Handling

### Common Error Scenarios

#### Library Not Found

```python
try:
    library_id = client.resolve_library_id("nonexistent-library")
except LibraryNotFoundError as e:
    print(f"Library not found: {e}")
    # Handle error - suggest alternatives or search
    alternatives = client.search_libraries("nonexistent-library")
    if alternatives:
        print(f"Did you mean: {alternatives[0]['name']}?")
```

#### Invalid Library ID

```python
try:
    docs = client.get_library_docs("/invalid/library-id")
except InvalidLibraryIDError as e:
    print(f"Invalid library ID: {e}")
    # Resolve correct library ID
    library_id = client.resolve_library_id("library-name")
    docs = client.get_library_docs(library_id)
```

#### Documentation Not Found

```python
try:
    docs = client.get_library_docs("/facebook/react", topic="nonexistent-topic")
except DocumentationNotFoundError as e:
    print(f"Documentation not found: {e}")
    # Get general documentation
    docs = client.get_library_docs("/facebook/react")
```

#### Rate Limiting

```python
try:
    docs = client.get_library_docs("/facebook/react")
except RateLimitError as e:
    print(f"Rate limit exceeded: {e}")
    # Wait and retry
    import time
    time.sleep(e.retry_after)
    docs = client.get_library_docs("/facebook/react")
```

#### API Errors

```python
try:
    docs = client.get_library_docs("/facebook/react")
except APIError as e:
    print(f"API error: {e}")
    # Handle based on error type
    if e.status_code == 401:
        print("Invalid API key")
    elif e.status_code == 500:
        print("Server error - try again later")
    else:
        print(f"Unexpected error: {e}")
```

### Error Recovery Strategies

#### Retry with Exponential Backoff

```python
import time
from context7 import APIError

def get_docs_with_retry(client, library_id, topic=None, max_retries=3):
    """Get documentation with retry logic."""
    for attempt in range(max_retries):
        try:
            return client.get_library_docs(library_id, topic)
        except APIError as e:
            if attempt == max_retries - 1:
                raise
            
            # Exponential backoff
            wait_time = 2 ** attempt
            time.sleep(wait_time)
            print(f"Retry {attempt + 1}/{max_retries} after {wait_time}s")
```

#### Fallback to Alternatives

```python
def get_docs_with_fallback(client, library_name, topic=None):
    """Get documentation with fallback to alternatives."""
    try:
        # Try exact match first
        library_id = client.resolve_library_id(library_name)
        return client.get_library_docs(library_id, topic)
    except LibraryNotFoundError:
        # Search for alternatives
        alternatives = client.search_libraries(library_name)
        if alternatives:
            print(f"Library '{library_name}' not found. Trying alternative: {alternatives[0]['name']}")
            library_id = alternatives[0]['id']
            return client.get_library_docs(library_id, topic)
        else:
            raise LibraryNotFoundError(f"No alternatives found for '{library_name}'")
```

---

## Performance Optimization

### Caching Strategies

#### Memory Caching

```python
from functools import lru_cache
from context7 import Context7Client

class CachedContext7Client:
    def __init__(self, cache_size=128):
        self.client = Context7Client()
        self.cache_size = cache_size
    
    @lru_cache(maxsize=128)
    def resolve_library_id_cached(self, library_name: str) -> str:
        """Cached library ID resolution."""
        return self.client.resolve_library_id(library_name)
    
    @lru_cache(maxsize=64)
    def get_library_docs_cached(self, library_id: str, topic: str = None) -> dict:
        """Cached documentation retrieval."""
        cache_key = f"{library_id}:{topic or 'general'}"
        return self.client.get_library_docs(library_id, topic)
```

#### Disk Caching

```python
import json
import hashlib
import os
from pathlib import Path
from context7 import Context7Client

class DiskCachedContext7Client:
    def __init__(self, cache_dir=".context7_cache", ttl=3600):
        self.client = Context7Client()
        self.cache_dir = Path(cache_dir)
        self.cache_dir.mkdir(exist_ok=True)
        self.ttl = ttl
    
    def _get_cache_path(self, key: str) -> Path:
        """Get cache file path for key."""
        hash_key = hashlib.md5(key.encode()).hexdigest()
        return self.cache_dir / f"{hash_key}.json"
    
    def _is_cache_valid(self, cache_path: Path) -> bool:
        """Check if cache file is still valid."""
        if not cache_path.exists():
            return False
        
        import time
        file_age = time.time() - cache_path.stat().st_mtime
        return file_age < self.ttl
    
    def get_library_docs(self, library_id: str, topic: str = None) -> dict:
        """Get documentation with disk caching."""
        cache_key = f"{library_id}:{topic or 'general'}"
        cache_path = self._get_cache_path(cache_key)
        
        # Check cache
        if self._is_cache_valid(cache_path):
            with open(cache_path, 'r') as f:
                return json.load(f)
        
        # Fetch from API
        docs = self.client.get_library_docs(library_id, topic)
        
        # Cache result
        with open(cache_path, 'w') as f:
            json.dump(docs, f)
        
        return docs
```

### Batch Operations

#### Batch Library Resolution

```python
from concurrent.futures import ThreadPoolExecutor, as_completed

class BatchContext7Client:
    def __init__(self, max_workers=5):
        self.client = Context7Client()
        self.max_workers = max_workers
    
    def resolve_multiple_libraries(self, library_names: list) -> dict:
        """Resolve multiple library names in parallel."""
        results = {}
        
        with ThreadPoolExecutor(max_workers=self.max_workers) as executor:
            # Submit all tasks
            future_to_name = {
                executor.submit(self.client.resolve_library_id, name): name
                for name in library_names
            }
            
            # Collect results
            for future in as_completed(future_to_name):
                name = future_to_name[future]
                try:
                    results[name] = future.result()
                except Exception as e:
                    results[name] = {"error": str(e)}
        
        return results
    
    def get_multiple_docs(self, library_configs: list) -> dict:
        """Get documentation for multiple libraries in parallel."""
        results = {}
        
        with ThreadPoolExecutor(max_workers=self.max_workers) as executor:
            future_to_config = {
                executor.submit(
                    self.client.get_library_docs,
                    config["library_id"],
                    config.get("topic"),
                    config.get("tokens", 5000)
                ): config
                for config in library_configs
            }
            
            for future in as_completed(future_to_config):
                config = future_to_config[future]
                key = f"{config['library_id']}:{config.get('topic', 'general')}"
                try:
                    results[key] = future.result()
                except Exception as e:
                    results[key] = {"error": str(e)}
        
        return results
```

### Token Optimization

#### Intelligent Token Allocation

```python
class TokenOptimizedClient:
    def __init__(self, default_tokens=5000, max_tokens=10000):
        self.client = Context7Client()
        self.default_tokens = default_tokens
        self.max_tokens = max_tokens
    
    def get_optimized_docs(self, library_id: str, topic: str = None, importance="normal") -> dict:
        """Get documentation with optimized token allocation."""
        # Allocate tokens based on importance
        if importance == "high":
            tokens = self.max_tokens
        elif importance == "low":
            tokens = self.default_tokens // 2
        else:
            tokens = self.default_tokens
        
        return self.client.get_library_docs(library_id, topic, tokens)
    
    def get_summary_docs(self, library_id: str, topic: str = None) -> dict:
        """Get summary documentation with reduced tokens."""
        return self.client.get_library_docs(library_id, topic, tokens=2000)
    
    def get_comprehensive_docs(self, library_id: str, topic: str = None) -> dict:
        """Get comprehensive documentation with maximum tokens."""
        return self.client.get_library_docs(library_id, topic, tokens=self.max_tokens)
```

---

## Monitoring and Analytics

### Usage Tracking

```python
import time
from dataclasses import dataclass
from typing import Dict, List

@dataclass
class UsageMetrics:
    total_requests: int = 0
    successful_requests: int = 0
    failed_requests: int = 0
    total_tokens_used: int = 0
    average_response_time: float = 0.0
    cache_hits: int = 0
    cache_misses: int = 0

class MonitoredContext7Client:
    def __init__(self):
        self.client = Context7Client()
        self.metrics = UsageMetrics()
        self.request_times: List[float] = []
    
    def get_library_docs(self, library_id: str, topic: str = None, tokens: int = 5000) -> dict:
        """Get documentation with monitoring."""
        start_time = time.time()
        
        try:
            docs = self.client.get_library_docs(library_id, topic, tokens)
            
            # Update metrics
            self.metrics.total_requests += 1
            self.metrics.successful_requests += 1
            self.metrics.total_tokens_used += docs.get("metadata", {}).get("tokens_used", 0)
            
            # Track response time
            response_time = time.time() - start_time
            self.request_times.append(response_time)
            self.metrics.average_response_time = sum(self.request_times) / len(self.request_times)
            
            return docs
            
        except Exception as e:
            self.metrics.total_requests += 1
            self.metrics.failed_requests += 1
            raise
    
    def get_metrics(self) -> UsageMetrics:
        """Get current usage metrics."""
        return self.metrics
    
    def reset_metrics(self) -> None:
        """Reset usage metrics."""
        self.metrics = UsageMetrics()
        self.request_times.clear()
```

### Performance Alerts

```python
class AlertingContext7Client:
    def __init__(self, slow_request_threshold=5.0, error_rate_threshold=0.1):
        self.client = Context7Client()
        self.slow_request_threshold = slow_request_threshold
        self.error_rate_threshold = error_rate_threshold
        self.metrics = UsageMetrics()
    
    def get_library_docs(self, library_id: str, topic: str = None, tokens: int = 5000) -> dict:
        """Get documentation with performance alerting."""
        start_time = time.time()
        
        try:
            docs = self.client.get_library_docs(library_id, topic, tokens)
            response_time = time.time() - start_time
            
            # Check for slow requests
            if response_time > self.slow_request_threshold:
                self._alert_slow_request(library_id, topic, response_time)
            
            self.metrics.total_requests += 1
            self.metrics.successful_requests += 1
            
            return docs
            
        except Exception as e:
            self.metrics.total_requests += 1
            self.metrics.failed_requests += 1
            
            # Check error rate
            error_rate = self.metrics.failed_requests / self.metrics.total_requests
            if error_rate > self.error_rate_threshold:
                self._alert_high_error_rate(error_rate)
            
            raise
    
    def _alert_slow_request(self, library_id: str, topic: str, response_time: float):
        """Alert on slow request."""
        print(f"ALERT: Slow request detected - Library: {library_id}, Topic: {topic}, Time: {response_time:.2f}s")
    
    def _alert_high_error_rate(self, error_rate: float):
        """Alert on high error rate."""
        print(f"ALERT: High error rate detected - Rate: {error_rate:.2%}")
```

---

**Last Updated**: 2025-11-12
**Version**: 4.0.0
**Maintained By**: Alfred SuperAgent (MoAI-ADK)
