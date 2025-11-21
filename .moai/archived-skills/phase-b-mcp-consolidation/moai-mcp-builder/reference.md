# MCP Builder Reference

## Official Resources

### Model Context Protocol
- MCP Official Spec: https://modelcontextprotocol.io
- GitHub Repository: https://github.com/modelcontextprotocol
- Python SDK: https://github.com/modelcontextprotocol/python-sdk

### FastMCP Framework
- FastMCP GitHub: https://github.com/jlowin/fastmcp
- FastMCP Docs: https://gofastmcp.com
- FastMCP 2.0 Announcement: https://gofastmcp.com/getting-started/welcome
- DataCamp Tutorial: https://www.datacamp.com/tutorial/building-mcp-server-client-fastmcp

### MCP Registry & Servers
- MCP Registry: https://registry.modelcontextprotocol.io
- Official Servers: https://github.com/modelcontextprotocol
- Community Servers: https://github.com/modelcontextprotocol/servers

## Core Concepts

### MCP Architecture

**Server** (Provides Tools/Resources):
- Implements tools agents can call
- Exposes resources/documents
- Defines prompt templates
- Listens on transport (stdio, HTTP, SSE)

**Client** (LLM or Agent):
- Discovers available tools
- Invokes tools when needed
- Requests resources
- Uses prompt templates

**Transport**:
- Stdio: Simple, local communication
- SSE (Server-Sent Events): HTTP long-polling
- WebSocket: Full-duplex communication
- HTTP: REST-style invocation

### Tool Lifecycle

```
1. Server defines tool with @mcp.tool
2. Tool registered with metadata (name, description, parameters)
3. Client discovers tool via introspection
4. Client invokes: tool_name(param1, param2, ...)
5. Server executes business logic
6. Tool returns structured result (dict, list, str)
7. Client receives typed result, processes/presents to agent
```

### Resource System

**Resources** provide data exposure without execution:

```
- Resource URI: scheme://identifier
- Example: "notion://page-id"
- Client requests: get_resource("notion://page-id")
- Server returns: document content
- Efficient for large content
- Streaming support for big files
```

### Prompts

**Prompt templates** are pre-built system prompts:

```
@server.prompt("code-review")
def code_review_prompt(language: str) -> str:
    return f"You are a {language} code reviewer..."

# Used in multi-turn conversations
# Parameterized based on context
# Consistent behavior across agents
```

## Transport Protocols

### Stdio (Standard In/Out)

```bash
# Simple pipe-based communication
# Use for local development, testing
# Example: agent | mcp-server | agent

fastmcp run server.py  # Starts stdio server
```

### SSE (Server-Sent Events)

```bash
# HTTP long-polling, suitable for browsers
# Lightweight, unidirectional (server→client)
# Good for read-heavy scenarios

fastmcp run server.py --sse --port 8000
# Client connects: http://localhost:8000/events
```

### WebSocket

```bash
# Full-duplex, low-latency communication
# Best for interactive applications
# Supports real-time bidirectional flow

fastmcp run server.py --websocket --port 8000
# ws://localhost:8000
```

## Authentication Patterns

### API Key Authentication

```python
from fastmcp.auth import APIKeyAuth

api_key_auth = APIKeyAuth(header="Authorization", prefix="Bearer")

@server.auth(api_key_auth)
@server.tool()
def protected_tool() -> dict:
    return {"secure": "data"}
```

### OAuth 2.0

```python
from fastmcp.auth import OAuth2Provider

oauth = OAuth2Provider(
    client_id="your-app-id",
    authorize_url="https://provider.com/oauth/authorize",
    token_url="https://provider.com/oauth/token",
    scopes=["read:data", "write:data"]
)

@server.auth(oauth)
@server.resource("data://{id}")
def oauth_protected_resource(id: str) -> str:
    return get_protected_data(id)
```

### Custom Authentication

```python
from fastmcp.auth import AuthProvider

class CustomAuth(AuthProvider):
    async def authenticate(self, request) -> dict:
        token = request.headers.get("X-Custom-Token")
        if not validate_token(token):
            raise AuthenticationError("Invalid token")
        return {"user_id": extract_user(token)}

@server.auth(CustomAuth())
@server.tool()
def custom_auth_tool() -> dict:
    return {"user": "authenticated"}
```

## Error Handling

### Standard Exceptions

```python
from fastmcp.exceptions import (
    MCPError,
    InvalidRequest,
    ResourceNotFound,
    ToolExecutionError,
    RateLimitError
)

@server.tool()
def error_handling_tool() -> dict:
    # Not found
    if not resource_exists():
        raise ResourceNotFound(f"Resource not found")
    
    # Invalid input
    if not validate_input(data):
        raise InvalidRequest("Invalid data format")
    
    # Execution error
    try:
        result = execute_operation()
    except Exception as e:
        raise ToolExecutionError(f"Operation failed: {str(e)}")
    
    # Rate limiting
    if rate_limited():
        raise RateLimitError("Too many requests", retry_after=60)
    
    return result
```

### Retry Logic

```python
from fastmcp.retry import exponential_backoff

@server.tool()
@exponential_backoff(max_retries=3, base_delay=1)
def resilient_tool() -> dict:
    # Automatically retries on transient errors
    return make_flaky_api_call()
```

## Deployment Considerations

### Performance Optimization

**Connection Pooling**:
```python
from sqlalchemy.pool import QueuePool

engine = create_engine(
    "postgresql://...",
    poolclass=QueuePool,
    pool_size=10,
    max_overflow=20
)
```

**Caching**:
```python
from functools import lru_cache

@lru_cache(maxsize=1000, ttl=3600)
def get_user(user_id: str) -> dict:
    return fetch_from_db(user_id)
```

**Async Operations**:
```python
import asyncio

async def batch_tool(items: list[str]) -> list[dict]:
    tasks = [async_process(item) for item in items]
    return await asyncio.gather(*tasks)
```

### Monitoring Setup

**Prometheus Metrics**:
```python
from prometheus_client import Counter, Histogram

tool_invocations = Counter(
    'mcp_tool_invocations_total',
    'Total tool invocations',
    ['tool_name']
)

tool_duration = Histogram(
    'mcp_tool_duration_seconds',
    'Tool execution duration',
    ['tool_name']
)

@server.tool()
def monitored_tool() -> dict:
    with tool_duration.labels(tool_name='monitored_tool').time():
        tool_invocations.labels(tool_name='monitored_tool').inc()
        return execute()
```

### Security Best Practices

- Validate all inputs (type checking + content)
- Implement rate limiting per client/user
- Use secrets management (AWS Secrets, Vault)
- Enable request signing/verification
- Audit logging for compliance
- HTTPS/TLS for all connections
- Sanitize error messages (no internal details)
- Keep dependencies updated

## Common Integration Patterns

### Database-Backed MCP

```python
from sqlalchemy import create_engine
from sqlalchemy.orm import Session

engine = create_engine(os.getenv("DATABASE_URL"))

@server.tool()
def query_users(status: str = "active") -> list[dict]:
    with Session(engine) as session:
        users = session.query(User).filter_by(status=status).all()
        return [user.to_dict() for user in users]
```

### API-Backed MCP

```python
import httpx

@server.tool()
async def fetch_external_data(endpoint: str) -> dict:
    async with httpx.AsyncClient() as client:
        response = await client.get(f"https://api.example.com/{endpoint}")
        response.raise_for_status()
        return response.json()
```

### File System MCP

```python
from pathlib import Path

@server.resource("file://{path}")
def read_file(path: str) -> str:
    file_path = Path(path)
    if not file_path.exists():
        raise FileNotFoundError(f"File not found: {path}")
    return file_path.read_text()

@server.tool()
def list_files(directory: str) -> list[str]:
    return [f.name for f in Path(directory).glob("*")]
```

### LLM-Powered MCP

```python
from anthropic import Anthropic

client = Anthropic()

@server.tool()
def generate_content(prompt: str) -> str:
    response = client.messages.create(
        model="claude-opus-4-1",
        max_tokens=2000,
        messages=[{"role": "user", "content": prompt}]
    )
    return response.content[0].text
```

---

**Last Updated**: November 2025
**FastMCP Version**: 2.0+
**Python Version**: 3.10+
**Protocol Version**: 1.0
