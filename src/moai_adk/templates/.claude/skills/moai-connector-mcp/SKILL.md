---
name: moai-connector-mcp
aliases: [moai-mcp-development]
category: connector
description: MCP 1.0+ Custom Server Development with FastMCP Framework
version: 3.0.0
modularized: true
tags:
  - connector
  - mcp-development
  - fastmcp
  - server-development
  - protocol
updated: 2025-11-27
status: active
---

## Quick Reference

**MCP Server Development Framework**

**What it does**: Comprehensive guide to building, testing, and deploying custom MCP (Model Context Protocol) servers using FastMCP framework for exposing tools, resources, and prompts to Claude and other AI models.

**Core Capabilities**:
- ✅ FastMCP server development with type-safe decorators
- ✅ Tool/Resource/Prompt architecture patterns
- ✅ Pydantic validation and error handling
- ✅ OAuth2 & API Key authentication patterns
- ✅ Performance monitoring and health checks
- ✅ Docker & Kubernetes deployment
- ✅ Testing strategies and validation
- ✅ Production-grade patterns (caching, circuit breaker, rate limiting)

**When to Use**:
- Building custom MCP servers for internal tools
- Exposing existing services via MCP protocol
- Implementing enterprise authentication patterns
- Deploying MCP servers in production environments
- Optimizing server performance and reliability
- Testing MCP server implementations

---

## Implementation Guide

### Getting Started with FastMCP

**Installation**:
```bash
pip install fastmcp
```

**Minimum Server**:
```python
from fastmcp import FastMCP

server = FastMCP("my-server")

@server.tool()
def hello_world(name: str) -> str:
    """Greet someone."""
    return f"Hello, {name}!"

if __name__ == "__main__":
    server.run()
```

**Core Concepts**:
1. **Tools**: Functions Claude can invoke with validated parameters
2. **Resources**: URI-based data endpoints for exposing information
3. **Prompts**: Reusable conversation templates and system prompts

Detailed guide: [getting-started.md](modules/development/getting-started.md)

---

### MCP Server Architecture

**Three-Component Pattern**:

```
┌─────────────────────────────────────────┐
│         MCP Server (FastMCP)            │
├─────────────────────────────────────────┤
│  Tools (Functions)                      │
│  • @server.tool() decorator             │
│  • Pydantic validation                  │
│  • Workflow-optimized naming            │
│                                         │
│  Resources (Data Endpoints)             │
│  • @server.resource("uri://...") decor  │
│  • Streaming support                    │
│  • Permission-based access              │
│                                         │
│  Prompts (Templates)                    │
│  • @server.prompt("name") decorator     │
│  • Parameter injection                  │
│  • Multi-turn workflows                 │
└─────────────────────────────────────────┘
        ↓
    MCP Protocol (JSON-RPC 2.0)
        ↓
┌─────────────────────────────────────────┐
│    Claude / LLM Client                  │
└─────────────────────────────────────────┘
```

**Design Patterns**: [server-design.md](modules/development/server-design.md)

---

### Production-Ready Server Example

```python
from fastmcp import FastMCP
from pydantic import Field
from typing import Literal, Optional

server = FastMCP("enterprise-database-server")

@server.tool()
def search_records(
    query: str,
    table: Literal["users", "products", "orders"],
    limit: int = Field(default=10, ge=1, le=100),
    filters: Optional[dict] = None
) -> dict:
    """
    Search database records with pagination.

    Args:
        query: Search query string
        table: Table to search
        limit: Max results (1-100)
        filters: Optional filter criteria

    Returns:
        Dict with results and metadata
    """
    if not query or not query.strip():
        raise ValueError("Query cannot be empty")

    results = execute_search(query, table, limit, filters)
    return {
        "status": "success",
        "count": len(results),
        "results": results,
        "total_available": get_total_count(query)
    }

@server.resource("db://{table}/{id}")
def get_record(table: str, id: str) -> dict:
    """Fetch record by ID."""
    record = fetch_record(table, id)
    if not record:
        raise ValueError(f"Record not found: {table}/{id}")
    return record

if __name__ == "__main__":
    server.run()
```

**Implementation Guide**: [implementation.md](modules/development/implementation.md)

---

### Authentication Patterns

**OAuth2 (User-Authenticated)**:
```python
from fastmcp.auth import OAuth2Provider

oauth = OAuth2Provider(
    authorize_url="https://auth.company.com/authorize",
    token_url="https://auth.company.com/token",
    scopes=["read:data", "write:data"]
)

@server.auth(oauth)
@server.tool()
def protected_action(user_id: str) -> dict:
    """Requires OAuth token."""
    return execute_action(user_id)
```

**API Key (Service-to-Service)**:
```python
from fastmcp.auth import APIKeyAuth

api_auth = APIKeyAuth(header="X-API-Key")

@server.auth(api_auth)
@server.resource("secure://{resource_id}")
def secure_resource(resource_id: str) -> str:
    """Requires API key."""
    return fetch_data(resource_id)
```

**Detailed Patterns**: [auth-patterns.md](modules/development/patterns/auth-patterns.md)

---

### Testing MCP Servers

**Unit Testing**:
```python
import pytest
from fastmcp import FastMCP

@pytest.fixture
def server():
    s = FastMCP("test-server")

    @s.tool()
    def add(a: int, b: int) -> int:
        return a + b

    return s

def test_add_tool(server):
    result = server.invoke_tool("add", {"a": 2, "b": 3})
    assert result == 5

def test_invalid_params(server):
    with pytest.raises(ValueError):
        server.invoke_tool("add", {"a": "not-a-number", "b": 3})
```

**Testing Guide**: [testing.md](modules/development/testing.md)

---

### Deployment

**Docker**:
```dockerfile
FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY server.py .
EXPOSE 8000
CMD ["python", "server.py"]
```

**Kubernetes**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mcp-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: mcp-server
  template:
    metadata:
      labels:
        app: mcp-server
    spec:
      containers:
      - name: mcp-server
        image: mcp-server:latest
        ports:
        - containerPort: 8000
        livenessProbe:
          httpGet:
            path: /health
            port: 8000
          initialDelaySeconds: 10
          periodSeconds: 10
```

**Deployment Guide**: [deployment.md](modules/development/deployment.md)

---

## Advanced Patterns

### Tool Design Best Practices

✅ **DO**:
- Design tools for single meaningful workflow steps
- Validate all inputs with Pydantic Field constraints
- Provide clear, actionable error messages
- Use pagination for large result sets
- Include comprehensive docstrings

❌ **DON'T**:
- Mix multiple responsibilities in one tool
- Return unlimited result sets
- Skip input validation
- Expose sensitive data
- Ignore error handling

**Tool Design Guide**: [tool-design.md](modules/development/patterns/tool-design.md)

---

### Performance Optimization

**Caching Strategy**:
```python
from functools import wraps
from datetime import datetime, timedelta

class MCPCache:
    def __init__(self, ttl_seconds=300):
        self.cache = {}
        self.ttl = ttl_seconds

    def cached(self, func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            key = str((func.__name__, args, kwargs))
            if key in self.cache:
                value, timestamp = self.cache[key]
                if (datetime.now() - timestamp).total_seconds() < self.ttl:
                    return value

            result = func(*args, **kwargs)
            self.cache[key] = (result, datetime.now())
            return result
        return wrapper

cache = MCPCache(ttl_seconds=600)

@cache.cached
def expensive_operation(param: str) -> dict:
    return fetch_and_process(param)
```

**Advanced Patterns**:
- [Caching](modules/development/patterns/caching.md)
- [Circuit Breaker](modules/development/patterns/circuit-breaker.md)
- [Rate Limiting](modules/development/patterns/rate-limiting.md)

---

### Monitoring & Observability

**Health Checks**:
```python
@server.resource("health://status")
def health_check() -> dict:
    """Server health status."""
    return {
        "status": "healthy",
        "version": "1.0.0",
        "uptime_seconds": get_uptime(),
        "active_connections": get_connection_count()
    }
```

**Metrics & Logging**:
```python
import logging
import time

logger = logging.getLogger(__name__)

@server.tool()
def monitored_operation(params: dict) -> dict:
    start = time.time()
    try:
        result = execute_operation(params)
        duration = time.time() - start
        logger.info(f"Operation completed in {duration:.2f}s")
        return result
    except Exception as e:
        logger.error(f"Operation failed: {str(e)}")
        raise
```

**Monitoring Guide**: [monitoring.md](modules/development/patterns/monitoring.md)

---

## Works Well With

- `moai-context7-integration` - Documentation access for API patterns
- `moai-cc-configuration` - MCP server configuration management
- `moai-essentials-debug` - Server debugging and troubleshooting
- `moai-domain-backend` - Backend service architecture
- `moai-domain-cloud` - Cloud deployment patterns
- `moai-quality-security` - Security validation and OWASP compliance

---

## Core Concepts

1. **FastMCP Framework**: Python library for rapid MCP server development
2. **Type Safety**: Pydantic models ensure Claude understands parameter constraints
3. **Workflow Design**: Tools for single meaningful tasks, not granular APIs
4. **Authentication Strategy**: OAuth2 for user apps, API keys for service-to-service
5. **Production Readiness**: Monitoring, health checks, error handling, caching
6. **Testing**: Comprehensive test coverage before deployment

---

## Module Navigation

**Getting Started**:
- [getting-started.md](modules/development/getting-started.md) - FastMCP basics

**Core Development**:
- [server-design.md](modules/development/server-design.md) - Architecture patterns
- [implementation.md](modules/development/implementation.md) - Real examples
- [testing.md](modules/development/testing.md) - Test strategies

**Deployment & Operations**:
- [deployment.md](modules/development/deployment.md) - Docker/Kubernetes

**Advanced Patterns** (modules/development/patterns/):
- [tool-design.md](modules/development/patterns/tool-design.md) - Tool best practices
- [resource-design.md](modules/development/patterns/resource-design.md) - Resource patterns
- [auth-patterns.md](modules/development/patterns/auth-patterns.md) - Authentication
- [monitoring.md](modules/development/patterns/monitoring.md) - Observability
- [caching.md](modules/development/patterns/caching.md) - Cache strategies
- [circuit-breaker.md](modules/development/patterns/circuit-breaker.md) - Resilience
- [rate-limiting.md](modules/development/patterns/rate-limiting.md) - Rate control

---

## Changelog

- **v3.0.0** (2025-11-27): Complete restructure from server integration to server development focus, modularized with development patterns
- **v2.1.0** (2025-11-22): Modularized structure - SKILL.md refactored, reference.md and examples.md added
- **v2.0.0** (2025-11-22): MCP 1.0+ protocol complete spec update
- **v1.0.0** (2025-11-21): Initial MCP integration skill

---

**Status**: Production Ready | See `modules/development/` for detailed patterns | Last Updated: 2025-11-27
