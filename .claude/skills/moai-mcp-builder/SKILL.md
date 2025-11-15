---
name: "moai-mcp-builder"
version: "4.0.0"
description: Enterprise MCP (Model Context Protocol) server development using FastMCP 2.0 with production-grade tools, resources, prompts, and intelligent agent-first design. Use when building MCP servers, integrating with LLMs, creating agent tools, implementing RAG systems, or developing protocol-based AI integration solutions.
allowed-tools: 
  - Read
  - Bash
  - WebSearch
  - WebFetch
status: stable
---

# Enterprise MCP Server Builder & AI Integration Platform

## MCP Protocol Capabilities

**Model Context Protocol (MCP)**:
- Standard protocol for LLM context provisioning
- Tool provisioning (agent-callable functions)
- Resource system (data/documents exposure)
- Prompt templates (multi-turn patterns)
- Transport abstraction (stdio, SSE, HTTP)

**FastMCP 2.0 Framework**:
- Python-first MCP implementation
- Type-safe decorators (@mcp.tool, @mcp.resource)
- Automatic OpenAPI generation
- Enterprise auth (OAuth, SAML)
- Proxy and composition patterns
- Production-ready deployments

## Skill Metadata
| Field | Value |
| ----- | ----- |
| **Version** | **4.0.0 Enterprise** |
| **Created** | 2025-11-12 |
| **Framework** | FastMCP 2.0, Python SDK |
| **Protocol** | Model Context Protocol (MCP) |
| **Features** | Tools, Resources, Prompts |
| **Transport** | Stdio, SSE, HTTP/WebSocket |
| **Auth** | OAuth 2.0, SAML, API Keys |
| **Tier** | **4 (Enterprise)** |

## MCP Architecture Overview

### Three Core Components

**1. Tools** (Agent-Callable Functions):
```
@mcp.tool
def search_documents(query: str) -> list[dict]:
    """Search documents by query"""
    # Implementation
    return results

# Tools are exposed as callable actions
# Agents decide when/how to invoke
# Return typed structured data
```

**2. Resources** (Data/Document Exposure):
```
@mcp.resource
def get_document(doc_id: str) -> str:
    """Fetch document content"""
    # Read from database
    return content

# Resources provide context without execution
# Large documents efficiently streamed
# Agents request as needed
```

**3. Prompts** (Multi-Turn Patterns):
```
@mcp.prompt("analyze-code")
def code_analysis(language: str, code_snippet: str) -> str:
    """System prompt for code analysis"""
    return f"""You are a {language} expert...
    Analyze this code:
    {code_snippet}"""

# Pre-built system prompts
# Contextual parameters
# Multi-turn conversation patterns
```

## Tool Design Principles

### Agent-Centric Design

**Principle 1: Build for Workflows, Not APIs**
```
BAD:
@mcp.tool
def create_event(title: str) -> dict: ...
@mcp.tool
def check_availability(date: str) -> bool: ...

GOOD:
@mcp.tool
def schedule_event(
    title: str,
    date: str,
    check_conflicts: bool = True
) -> dict:
    """Single tool combining availability check + creation"""
    if check_conflicts and has_conflict(date):
        raise ValueError(f"Conflict on {date}")
    return create_event(title, date)
```

**Principle 2: Optimize for Limited Context**
```
BAD:
def list_all_users() -> list[dict]:
    """Returns 10,000 users with all fields"""
    return all_users_with_full_data()

GOOD:
def search_users(
    query: str,
    limit: int = 10,
    fields: list[str] = ["id", "name", "email"]
) -> list[dict]:
    """Search with pagination and field filtering"""
    return paginated_filtered_search(query, limit, fields)
```

**Principle 3: Design Actionable Error Messages**
```
BAD:
raise ValueError("Invalid date")

GOOD:
raise ValueError(
    f"Date must be in future. "
    f"Current date: {today}. "
    f"Try: {(today + timedelta(days=1)).isoformat()}"
)
```

**Principle 4: Follow Natural Task Subdivisions**
```
Tool Naming Conventions:
- create_*: Create new resources
- update_*: Modify existing resources
- delete_*: Remove resources
- list_*: Enumerate resources
- get_*: Fetch specific resource
- search_*: Find by criteria
- analyze_*: Generate insights
```

### Type Safety & Documentation

**Parameter Types**:
```python
from typing import Literal, Optional, Union
from pydantic import Field, validator

@mcp.tool
def process_data(
    data: dict,
    mode: Literal["fast", "thorough"] = "fast",
    timeout: Optional[int] = None,
    callbacks: list[str] = Field(default_factory=list),
) -> dict:
    """
    Process data with specified mode.
    
    Args:
        data: Input dictionary
        mode: Processing speed preference
        timeout: Optional timeout in seconds
        callbacks: List of webhook URLs
        
    Returns:
        Processed result dictionary
    """
    # Implementation with validation
```

## FastMCP Server Implementation

### Basic Server Structure

```python
from fastmcp import FastMCP

server = FastMCP("my-server")

@server.tool()
def get_weather(city: str) -> str:
    """Get current weather for a city"""
    # Fetch from API
    return f"Weather in {city}: 72°F, sunny"

@server.resource("weather://{city}")
def weather_resource(city: str) -> str:
    """Get detailed weather data"""
    return f"""Weather Report for {city}
    Temperature: 72°F
    Humidity: 65%
    Wind: 10mph"""

@server.prompt("weather-analyst")
def weather_prompt() -> str:
    """System prompt for weather analysis"""
    return """You are a weather expert. Analyze the provided 
    weather data and give actionable recommendations."""

if __name__ == "__main__":
    server.run()
```

### Authentication Layer

```python
from fastmcp.auth import OAuth2Provider, APIKeyAuth

# OAuth2 Configuration
oauth = OAuth2Provider(
    authorize_url="https://auth.example.com/authorize",
    token_url="https://auth.example.com/token",
    scopes=["read:data", "write:data"]
)

@server.auth(oauth)
@server.tool()
def protected_tool(user_id: str) -> dict:
    """Tool requiring OAuth authentication"""
    return get_user_data(user_id)

# API Key Authentication
api_auth = APIKeyAuth(header="X-API-Key")

@server.auth(api_auth)
@server.resource("secure://{resource_id}")
def secure_resource(resource_id: str) -> str:
    """Resource with API key protection"""
    return fetch_secure_data(resource_id)
```

## Real-World Patterns

### Pattern 1: Database Query Tool

```python
from sqlalchemy import create_engine, select
from sqlalchemy.orm import Session

engine = create_engine("postgresql://...")

@server.tool()
def query_database(
    table: Literal["users", "products", "orders"],
    filters: dict = None,
    limit: int = 10
) -> list[dict]:
    """
    Query database with filters and pagination.
    
    Examples:
    - table="users", filters={"status": "active"}
    - table="products", filters={"price_gt": 100}
    """
    with Session(engine) as session:
        query = get_query(table)
        if filters:
            query = apply_filters(query, filters)
        return [dict(row) for row in query.limit(limit)]
```

### Pattern 2: API Integration Tool

```python
import httpx
from datetime import datetime, timedelta

@server.tool()
def search_articles(
    query: str,
    source: Literal["hacker-news", "arxiv", "medium"] = "hacker-news",
    hours_back: int = 24
) -> list[dict]:
    """
    Search articles from multiple sources.
    
    Args:
        query: Search terms
        source: Publication source
        hours_back: Search time window
    """
    cutoff = datetime.now() - timedelta(hours=hours_back)
    
    clients = {
        "hacker-news": HNClient(),
        "arxiv": ArxivClient(),
        "medium": MediumClient(),
    }
    
    results = clients[source].search(query, since=cutoff)
    return sorted(results, key=lambda x: x["relevance"], reverse=True)[:10]
```

### Pattern 3: File Processing Tool

```python
from pathlib import Path
import json, yaml

@server.tool()
def process_config_file(
    file_path: str,
    format: Literal["json", "yaml", "toml"] = "json",
    validate: bool = True
) -> dict:
    """
    Load and validate configuration file.
    
    Args:
        file_path: Path to config file
        format: File format auto-detection
        validate: Validate against schema
    """
    path = Path(file_path)
    if not path.exists():
        raise FileNotFoundError(f"Config file not found: {file_path}")
    
    loaders = {
        "json": json.loads,
        "yaml": yaml.safe_load,
        "toml": tomli.loads,
    }
    
    content = path.read_text()
    config = loaders[format](content)
    
    if validate:
        validate_config_schema(config)
    
    return config
```

### Pattern 4: LLM Integration Tool

```python
from anthropic import Anthropic

client = Anthropic()

@server.tool()
def analyze_text(
    text: str,
    analysis_type: Literal["summary", "sentiment", "entities", "topics"],
    max_tokens: int = 1000
) -> str:
    """
    Analyze text using Claude API.
    
    Args:
        text: Input text to analyze
        analysis_type: Type of analysis
        max_tokens: Max response length
    """
    prompts = {
        "summary": f"Summarize: {text}",
        "sentiment": f"Analyze sentiment: {text}",
        "entities": f"Extract entities: {text}",
        "topics": f"Identify topics: {text}",
    }
    
    response = client.messages.create(
        model="claude-opus-4-1",
        max_tokens=max_tokens,
        messages=[{"role": "user", "content": prompts[analysis_type]}]
    )
    
    return response.content[0].text
```

### Pattern 5: Proxy Server (Composing Multiple MCPs)

```python
from fastmcp.proxy import ProxyServer

proxy = ProxyServer("multi-protocol-proxy")

# Mount multiple MCP servers
proxy.mount("/weather", weather_server)
proxy.mount("/database", database_server)
proxy.mount("/api", api_server)

@proxy.tool()
def unified_search(
    query: str,
    search_in: Literal["weather", "database", "api"] = "all"
) -> dict:
    """
    Search across mounted servers.
    """
    results = {}
    if search_in in ["all", "weather"]:
        results["weather"] = proxy.invoke("/weather/search", query)
    if search_in in ["all", "database"]:
        results["db"] = proxy.invoke("/database/search", query)
    if search_in in ["all", "api"]:
        results["api"] = proxy.invoke("/api/search", query)
    return results

if __name__ == "__main__":
    proxy.run()
```

## Deployment Strategies

### Local Development

```bash
# Install dependencies
pip install fastmcp[dev]

# Run in development mode
fastmcp dev my_server.py

# Test with cURL
curl -X POST http://localhost:8000/tool/search_documents \
  -H "Content-Type: application/json" \
  -d '{"query": "python"}'
```

### Docker Deployment

```dockerfile
FROM python:3.11-slim

WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt

COPY . .

EXPOSE 8000
CMD ["fastmcp", "run", "server.py", "--host", "0.0.0.0"]
```

### Cloud Deployment

```yaml
# deploy.yaml (Kubernetes)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mcp-server
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: mcp-server
        image: my-mcp-server:latest
        ports:
        - containerPort: 8000
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: url
```

## Best Practices

### Error Handling

```python
from fastmcp.exceptions import MCPError

@server.tool()
def resilient_tool(resource_id: str) -> dict:
    """Tool with comprehensive error handling"""
    try:
        resource = fetch_resource(resource_id)
        if not resource:
            raise MCPError(
                f"Resource not found: {resource_id}. "
                f"Try: list_available_resources()"
            )
        return process_resource(resource)
    except DatabaseError as e:
        raise MCPError(f"Database error: {str(e)}", retry=True)
    except ValidationError as e:
        raise MCPError(f"Invalid input: {e.details}", retry=False)
```

### Performance Optimization

```python
from functools import lru_cache
import asyncio

@lru_cache(maxsize=128)
def expensive_computation(key: str) -> dict:
    """Cache results of expensive operations"""
    return fetch_and_process(key)

@server.tool()
async def batch_process(items: list[str]) -> list[dict]:
    """Use async for concurrent operations"""
    tasks = [process_item_async(item) for item in items]
    return await asyncio.gather(*tasks)
```

### Monitoring & Logging

```python
import logging
from fastmcp.monitoring import instrument

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

@server.tool()
@instrument(track_duration=True, track_errors=True)
def monitored_tool(data: dict) -> dict:
    """Tool with automatic monitoring"""
    logger.info(f"Processing: {data}")
    result = process(data)
    logger.info(f"Result: {result}")
    return result
```

---

**MCP Production Readiness Checklist**:
- [ ] All tools have type hints and docstrings
- [ ] Error messages are actionable
- [ ] Authentication configured (OAuth/API keys)
- [ ] Logging and monitoring enabled
- [ ] Rate limiting implemented
- [ ] Unit tests >85% coverage
- [ ] Integration tests for tool chains
- [ ] Performance benchmarks <500ms p99
- [ ] Documentation with examples
- [ ] Deployment strategy defined
