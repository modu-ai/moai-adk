---
name: moai-mcp-integration
description: MCP 1.0+ (Model Context Protocol) Enterprise Integration with 10+ Official Servers
allowed-tools: [Read, Bash, WebFetch]
---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-22  
**compliance_score**: 75%  
**auto_trigger_keywords**: moai, mcp, integration  


## Quick Reference

**MCP 1.0+ Enterprise Integration Hub**

**What it does**: Unified MCP (Model Context Protocol) integration framework for connecting to 10+ official servers, building custom MCP servers, and orchestrating enterprise workflows.

**Core Capabilities**:
- ✅ MCP 1.0+ protocol compliance (Tool/Resource/Prompt architecture)
- ✅ 10+ official servers: Playwright, Figma, GitHub, Notion, Firebase, Supabase, etc.
- ✅ Custom FastMCP server development with Pydantic validation
- ✅ OAuth2 & API Key authentication patterns
- ✅ Docker & Kubernetes enterprise deployment
- ✅ Multi-server orchestration and plugin routing
- ✅ Performance monitoring and observability

**When to Use**:
- Integrating external services via MCP protocol
- Building custom MCP servers for internal tools
- Orchestrating multiple MCP servers in production
- Implementing enterprise authentication & security
- Setting up monitoring and health dashboards

---

## Implementation Guide

### MCP Architecture Overview

**MCP Server Components** (v1.0+):
```
MCP Server (Tool/Resource/Prompt Pattern):
├── Tools: Agent-callable functions with Pydantic validation
│   └─ @server.tool() decorator
│   └─ Type-safe parameters and return values
│   └─ Workflow-optimized naming (search_*, create_*, update_*)
│
├── Resources: Data/document exposure via URI patterns
│   └─ @server.resource("uri://{id}") decorator
│   └─ Streaming support for large datasets
│   └─ Permission-based access control
│
└── Prompts: Conversation templates for multi-turn workflows
    └─ @server.prompt("name") decorator
    └─ Contextual parameter injection
    └─ System prompt customization
```

### FastMCP Server Template

**Production-Ready Server Example**:
```python
from fastmcp import FastMCP
from pydantic import Field
from typing import Literal, Optional

server = FastMCP("enterprise-server")

@server.tool()
def search_database(
    query: str,
    table: Literal["users", "products", "orders"],
    limit: int = Field(default=10, ge=1, le=100),
    filters: Optional[dict] = None
) -> list[dict]:
    """Search with pagination and validation."""
    if not query:
        raise ValueError("Query cannot be empty")
    return execute_query(query, table, filters)[:limit]

@server.resource("db://{table}/{id}")
def get_record(table: str, id: str) -> dict:
    """Fetch record by ID."""
    return fetch_record(table, id)

if __name__ == "__main__":
    server.run()
```

### Authentication Patterns

**OAuth2** (for user-authenticated services):
```python
from fastmcp.auth import OAuth2Provider

oauth = OAuth2Provider(
    authorize_url="https://auth.service.com/authorize",
    token_url="https://auth.service.com/token",
    scopes=["read:data", "write:data"]
)

@server.auth(oauth)
@server.tool()
def protected_action(user_id: str) -> dict:
    """Requires OAuth token."""
    return execute_action(user_id)
```

**API Key** (for service-to-service):
```python
from fastmcp.auth import APIKeyAuth

api_auth = APIKeyAuth(header="X-API-Key")

@server.auth(api_auth)
@server.resource("secure://{resource_id}")
def secure_resource(resource_id: str) -> str:
    """Requires API key."""
    return fetch_data(resource_id)
```

### Enterprise Deployment

**Docker** (single container):
```dockerfile
FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt . && pip install -r requirements.txt
COPY server.py .
EXPOSE 8000
CMD ["python", "server.py"]
```

**Kubernetes** (production scale):
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
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: mcp-secrets
              key: database-url
```

---

## Best Practices

✅ **DO**:
- Design tools for workflows (single request solving one task)
- Validate all inputs with Pydantic models
- Provide actionable error messages
- Use pagination for large datasets
- Implement authentication for sensitive operations
- Monitor performance and availability
- Follow MCP 1.0+ protocol specifications

❌ **DON'T**:
- Expose sensitive data without authentication
- Return unlimited result sets
- Mix responsibilities (one tool = one task)
- Skip input validation
- Deploy without monitoring
- Use deprecated MCP 0.x patterns
- Ignore error handling

---

## Works Well With

- `moai-context7-integration` - Documentation access for 50+ languages
- `moai-cc-configuration` - MCP server configuration management
- `moai-essentials-debug` - MCP server debugging and troubleshooting
- `moai-domain-backend` - Backend service architecture
- `moai-domain-cloud` - Cloud deployment patterns

---

## Core Concepts

1. **Tool/Resource/Prompt Architecture**: MCP's three-pillar pattern for exposing server capabilities
2. **Type Safety**: Pydantic validation ensures Claude understands parameter constraints
3. **Workflow Optimization**: Design tools for single meaningful tasks, not granular APIs
4. **Authentication Strategy**: OAuth2 for users, API keys for services, mutual TLS for infrastructure
5. **Observability**: Monitor latency, success rate, error patterns for production health

---

## Changelog

- **v2.1.0** (2025-11-22): Modularized structure - SKILL.md refactored, reference.md and examples.md added
- **v2.0.0** (2025-11-22): MCP 1.0+ protocol complete spec update
- **v1.0.0** (2025-11-21): Initial MCP integration skill

---

**End of Core Skill** | See `modules/`, `examples.md`, and `reference.md` for detailed patterns | Status: Production Ready

