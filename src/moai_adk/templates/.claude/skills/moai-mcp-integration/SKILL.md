---
name: moai-mcp-integration
description: MCP 1.0+ (Model Context Protocol) Enterprise Integration with 10+ Official Servers (2025 Edition)
allowed-tools: [Read, Bash, WebFetch]
---

## Quick Reference (30 seconds)

# MCP 1.0+ Enterprise Integration Hub (2025 Edition)

**What it does**: Unified MCP (Model Context Protocol) integration hub providing access to 10+ official MCP servers, custom MCP server development patterns, and enterprise orchestration strategies.

**Core Capabilities**:
- ✅ MCP 1.0+ protocol specification compliance
- ✅ 10+ official MCP servers: Playwright, Figma, GitHub, Notion, Firebase, Supabase, etc.
- ✅ Custom MCP server development with FastMCP
- ✅ Tool/Resource/Prompt architecture patterns
- ✅ Enterprise deployment strategies (Docker, Kubernetes, serverless)
- ✅ Authentication & authorization patterns
- ✅ Performance monitoring and observability

**When to Use**:
- Integrating external services via MCP protocol
- Building custom MCP servers for internal tools
- Orchestrating multiple MCP servers
- Deploying production-grade MCP infrastructure
- Implementing MCP authentication & security

**Quick Start Example**:
```python
from fastmcp import FastMCP

server = FastMCP("enterprise-mcp-server")

@server.tool()
def search_database(query: str, limit: int = 10) -> list[dict]:
    """Search database with pagination."""
    return execute_query(query, limit)

if __name__ == "__main__":
    server.run()
```

---

## Implementation Guide

### MCP 1.0+ Protocol Overview (2025 Edition)

**MCP Architecture Components**:
```
MCP Server Structure (Spec 1.0+):
  ├─ Tools (Agent-Callable Functions)
  │   └─ @server.tool() decorator
  │   └─ Type-safe parameters (Pydantic)
  │   └─ Structured return values
  │
  ├─ Resources (Data/Document Exposure)
  │   └─ @server.resource() decorator
  │   └─ URI-based access patterns
  │   └─ Streaming support for large data
  │
  └─ Prompts (Conversation Templates)
      └─ @server.prompt() decorator
      └─ Multi-turn conversation patterns
      └─ Contextual parameter injection
```

**Tool Design Best Practices (MCP 1.0+)**:
```
Principle 1: Build for Workflows, Not APIs
  BAD:  create_event(), check_availability() (two separate calls)
  GOOD: schedule_event(check_conflicts=True) (single workflow)

Principle 2: Optimize for Limited Context
  BAD:  list_all_users() → Returns 10,000 records
  GOOD: search_users(query, limit=10, fields=["id", "name"])

Principle 3: Actionable Error Messages
  BAD:  "Invalid date"
  GOOD: f"Date must be future. Current: {today}. Try: {tomorrow}"

Principle 4: Natural Task Subdivisions
  Naming: create_*, update_*, delete_*, list_*, get_*, search_*, analyze_*
```

### Official MCP Servers (10+ Servers, 2025 Edition)

**1. Playwright MCP** - Web Automation & Testing
```json
{
  "mcpServers": {
    "playwright": {
      "command": "npx",
      "args": ["-y", "@microsoft/mcp-server-playwright"],
      "capabilities": ["browser-automation", "web-testing", "screenshots"],
      "tools": [
        "navigate_to(url)",
        "click(selector)",
        "fill_input(selector, value)",
        "take_screenshot(path)",
        "get_page_content()",
        "execute_script(js_code)"
      ],
      "use_cases": ["E2E testing", "Web scraping", "UI automation"]
    }
  }
}
```

**2. Figma MCP** - Design System Access
```json
{
  "mcpServers": {
    "figma": {
      "command": "npx",
      "args": ["-y", "@figma/mcp-server-figma"],
      "oauth": {
        "clientId": "${FIGMA_CLIENT_ID}",
        "clientSecret": "${FIGMA_CLIENT_SECRET}",
        "scopes": ["file:read"]
      },
      "capabilities": ["design-access", "component-library", "dev-mode"],
      "tools": [
        "get_file(file_key)",
        "get_components(file_key)",
        "export_assets(file_key, node_ids)",
        "get_styles(file_key)"
      ],
      "use_cases": ["Design tokens", "Component library", "Design-to-code"]
    }
  }
}
```

**3. GitHub MCP** - Repository Management
```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-server-github"],
      "oauth": {
        "clientId": "${GITHUB_CLIENT_ID}",
        "clientSecret": "${GITHUB_CLIENT_SECRET}",
        "scopes": ["repo", "issues", "pull_requests"]
      },
      "capabilities": ["repository", "issues", "pull-requests", "workflows"],
      "tools": [
        "create_issue(repo, title, body)",
        "create_pr(repo, title, base, head)",
        "list_issues(repo, state)",
        "merge_pr(repo, pr_number)"
      ]
    }
  }
}
```

**4. Notion MCP** - Knowledge Base Integration
```json
{
  "mcpServers": {
    "notion": {
      "command": "npx",
      "args": ["-y", "@notionhq/mcp-server-notion"],
      "auth": {
        "type": "bearer",
        "token": "${NOTION_API_KEY}"
      },
      "capabilities": ["database", "pages", "blocks"],
      "tools": [
        "query_database(database_id, filter)",
        "create_page(parent_id, properties)",
        "update_page(page_id, properties)",
        "append_blocks(page_id, blocks)"
      ]
    }
  }
}
```

**5. Firebase MCP** - Backend-as-a-Service (2025 New)
```json
{
  "mcpServers": {
    "firebase": {
      "command": "npx",
      "args": ["-y", "@firebase/mcp-server-firebase"],
      "auth": {
        "type": "service-account",
        "credential_path": "${FIREBASE_SERVICE_ACCOUNT_PATH}"
      },
      "capabilities": ["firestore", "auth", "storage", "functions"],
      "tools": [
        "firestore_query(collection, filters)",
        "firestore_create(collection, data)",
        "auth_create_user(email, password)",
        "storage_upload(path, file)"
      ]
    }
  }
}
```

**6. Supabase MCP** - Open-Source Backend (2025 New)
```json
{
  "mcpServers": {
    "supabase": {
      "command": "npx",
      "args": ["-y", "@supabase/mcp-server-supabase"],
      "auth": {
        "type": "api-key",
        "key": "${SUPABASE_API_KEY}",
        "url": "${SUPABASE_URL}"
      },
      "capabilities": ["database", "auth", "storage", "realtime"],
      "tools": [
        "query_table(table, filters)",
        "insert_row(table, data)",
        "update_row(table, id, data)",
        "subscribe_changes(table, callback)"
      ]
    }
  }
}
```

**7-10. Additional Official Servers**:
- **PlanetScale MCP**: MySQL database hosting
- **Stripe MCP**: Payment processing
- **Slack MCP**: Messaging & collaboration
- **Google Calendar MCP**: Schedule management

### Custom MCP Server Development

**FastMCP Server Template (MCP 1.0+)**:
```python
from fastmcp import FastMCP
from typing import Literal, Optional
from pydantic import Field

server = FastMCP("enterprise-database-server")

# Tool: Database Query
@server.tool()
def search_database(
    query: str,
    table: Literal["users", "products", "orders"],
    limit: int = Field(default=10, ge=1, le=100),
    filters: Optional[dict] = None
) -> list[dict]:
    """
    Search database with filters and pagination.
    
    Args:
        query: Search query string
        table: Database table to search
        limit: Maximum results (1-100)
        filters: Optional filter conditions
    
    Returns:
        List of matching records
    
    Examples:
        >>> search_database("laptop", "products", limit=5)
        [{"id": 1, "name": "Laptop Pro", "price": 1200}, ...]
    """
    # Validate inputs
    if not query:
        raise ValueError("Query cannot be empty")
    
    # Execute query with type safety
    results = execute_query(query, table, filters)
    return results[:limit]

# Resource: Data Exposure
@server.resource("database://{table}/{id}")
def get_record(table: str, id: str) -> dict:
    """
    Fetch specific record by ID.
    
    Args:
        table: Table name
        id: Record identifier
    
    Returns:
        Record data
    """
    return fetch_record(table, id)

# Prompt: Conversation Template
@server.prompt("data-analyst")
def analyst_prompt(domain: str) -> str:
    """
    System prompt for data analysis tasks.
    
    Args:
        domain: Analysis domain (sales, marketing, operations)
    
    Returns:
        Formatted system prompt
    """
    return f"""You are a {domain} data analyst expert.
    Analyze the provided data and generate actionable insights.
    Focus on trends, anomalies, and recommendations."""

if __name__ == "__main__":
    server.run()
```

### Authentication & Authorization

**OAuth2 Pattern**:
```python
from fastmcp.auth import OAuth2Provider

oauth = OAuth2Provider(
    authorize_url="https://auth.example.com/authorize",
    token_url="https://auth.example.com/token",
    scopes=["read:data", "write:data"]
)

@server.auth(oauth)
@server.tool()
def protected_operation(user_id: str) -> dict:
    """Tool requiring OAuth authentication."""
    return execute_protected_action(user_id)
```

**API Key Pattern**:
```python
from fastmcp.auth import APIKeyAuth

api_auth = APIKeyAuth(header="X-API-Key")

@server.auth(api_auth)
@server.resource("secure://{resource_id}")
def secure_resource(resource_id: str) -> str:
    """Resource with API key protection."""
    return fetch_secure_data(resource_id)
```

### Enterprise Deployment

**Docker Deployment**:
```dockerfile
FROM python:3.11-slim

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install -r requirements.txt

# Copy MCP server
COPY server.py .

# Expose MCP port
EXPOSE 8000

# Run server
CMD ["python", "server.py"]
```

**Kubernetes Deployment**:
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
        - name: API_KEY
          valueFrom:
            secretKeyRef:
              name: mcp-secrets
              key: api-key
---
apiVersion: v1
kind: Service
metadata:
  name: mcp-server-service
spec:
  selector:
    app: mcp-server
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8000
  type: LoadBalancer
```

---

## Advanced Patterns

### Multi-Server Orchestration

**Plugin Orchestrator**:
```python
class MCPPluginOrchestrator:
    """Orchestrate multiple MCP servers."""
    
    def __init__(self, config_path: str):
        self.config = load_config(config_path)
        self.plugins = {}
        self.router = PluginRouter()
    
    def discover_plugins(self) -> dict[str, PluginInfo]:
        """Discover available MCP plugins."""
        discovered = {}
        
        for plugin_name, plugin_config in self.config["mcpServers"].items():
            plugin_info = PluginInfo(
                name=plugin_name,
                command=plugin_config["command"],
                args=plugin_config.get("args", []),
                capabilities=self.detect_capabilities(plugin_config),
                health_status=self.check_health(plugin_name)
            )
            discovered[plugin_name] = plugin_info
        
        return discovered
    
    def route_request(self, request: PluginRequest) -> PluginResponse:
        """Route request to appropriate plugin."""
        matching_plugins = [
            name for name, info in self.plugins.items()
            if all(cap in info.capabilities for cap in request.required_capabilities)
        ]
        
        if not matching_plugins:
            raise NoPluginAvailable(f"No plugin supports: {request.required_capabilities}")
        
        selected_plugin = self.select_best_plugin(matching_plugins)
        return self.execute_plugin_request(selected_plugin, request)
```

### Performance Monitoring

**Monitoring Dashboard**:
```python
class MCPMonitoring:
    """Enterprise monitoring for MCP infrastructure."""
    
    def __init__(self):
        self.metrics_collector = MetricsCollector()
        self.alerting = AlertingSystem()
    
    def track_operation(self, operation_type: str, plugin_name: str, duration_ms: float, success: bool):
        """Track individual operations."""
        self.metrics_collector.record({
            "operation": operation_type,
            "plugin": plugin_name,
            "duration_ms": duration_ms,
            "success": success,
            "timestamp": datetime.now()
        })
        
        if duration_ms > 5000:
            self.alerting.send_alert(
                severity="warning",
                message=f"{plugin_name} operation slow: {duration_ms:.0f}ms"
            )
    
    def generate_health_dashboard(self) -> HealthDashboard:
        """Generate real-time health dashboard."""
        metrics = self.metrics_collector.get_recent_metrics(hours=24)
        
        return HealthDashboard(
            total_requests=len(metrics),
            success_rate=calculate_success_rate(metrics),
            avg_latency=calculate_avg_latency(metrics),
            p99_latency=calculate_percentile(metrics, 99),
            plugin_health=self.get_plugin_health_summary(),
            alerts=self.alerting.get_active_alerts()
        )
```

---

## Best Practices

✅ **DO**:
- Follow MCP 1.0+ protocol specification
- Use type-safe parameters (Pydantic)
- Implement authentication for sensitive tools
- Provide clear error messages
- Monitor performance metrics
- Use tool/resource/prompt architecture

❌ **DON'T**:
- Expose sensitive data without authentication
- Return large datasets (use pagination)
- Skip input validation
- Ignore error handling
- Deploy without monitoring
- Use deprecated MCP 0.x patterns

---

## Related Skills

- `moai-cc-configuration` - MCP server configuration
- `moai-essentials-debug` - MCP server debugging
- `moai-foundation-trust` - Security validation
- `moai-context7-lang-integration` - Documentation access

---

## Changelog

- **v2.0.0** (2025-11-22): Complete update with MCP 1.0+ protocol, 10+ official servers, reference.md and examples.md added
- **v1.0.0** (2025-11-21): Initial MCP integration skill

---

**End of Skill** | Updated 2025-11-22 | Status: Production Ready

