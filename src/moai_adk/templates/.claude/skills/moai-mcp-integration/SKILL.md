---
name: moai-mcp-integration
description: Comprehensive MCP (Model Context Protocol) development with AI-powered creation, plugin management, enterprise deployment patterns, and Context7 integration for latest MCP standards.
---

## Quick Reference (30 seconds)

**Model Context Protocol (MCP)** is the standard interface for LLMs to access external tools, data, and services. This skill consolidates MCP server development (FastMCP), plugin management, and enterprise orchestration.

**Core Capabilities**:
- **FastMCP Server Development**: Enterprise MCP server creation with type safety
- **Plugin Orchestration**: Dynamic plugin discovery, loading, and coordination
- **Enterprise Integration**: CI/CD pipelines, authentication, monitoring
- **AI-Powered Optimization**: Context7-driven best practices and performance tuning

**When to Use**:
- Building new MCP servers for external services
- Managing multiple MCP plugins across projects
- Orchestrating MCP ecosystems with intelligent routing
- Deploying production-grade MCP infrastructure

**Key Concept**: MCP servers expose **Tools** (agent-callable functions), **Resources** (data/documents), and **Prompts** (conversation templates) through a standard protocol.

---

## Implementation Guide

### Part 1: FastMCP-Based MCP Server Development

**MCP Architecture Components**:

```
MCP Server Structure:
  ├─ Tools (Agent-Callable Functions)
  │   └─ @server.tool() decorator
  │   └─ Type-safe parameters with Pydantic
  │   └─ Return structured data
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

**FastMCP Server Template**:

```python
from fastmcp import FastMCP
from typing import Literal, Optional
from pydantic import Field

server = FastMCP("enterprise-mcp-server")

# Tool: Agent-callable function
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
    # Implementation with type validation
    results = execute_query(query, table, filters)
    return results[:limit]

# Resource: Data exposure
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

# Prompt: Conversation template
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

**Agent-Centric Tool Design Principles**:

```
Principle 1: Build for Workflows, Not APIs
  BAD:  @server.tool() create_event()
        @server.tool() check_availability()
  GOOD: @server.tool() schedule_event(check_conflicts=True)

Principle 2: Optimize for Limited Context
  BAD:  list_all_users() → Returns 10,000 records
  GOOD: search_users(query, limit=10, fields=["id", "name"])

Principle 3: Design Actionable Error Messages
  BAD:  raise ValueError("Invalid date")
  GOOD: raise ValueError(f"Date must be future. Current: {today}. Try: {tomorrow}")

Principle 4: Follow Natural Task Subdivisions
  Naming Conventions:
    - create_*: Create new resources
    - update_*: Modify existing
    - delete_*: Remove resources
    - list_*: Enumerate resources
    - get_*: Fetch specific item
    - search_*: Find by criteria
    - analyze_*: Generate insights
```

**Authentication Patterns**:

```python
from fastmcp.auth import OAuth2Provider, APIKeyAuth

# OAuth2 for user-specific operations
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

# API Key for service-to-service
api_auth = APIKeyAuth(header="X-API-Key")

@server.auth(api_auth)
@server.resource("secure://{resource_id}")
def secure_resource(resource_id: str) -> str:
    """Resource with API key protection."""
    return fetch_secure_data(resource_id)
```

### Part 2: Plugin Orchestration Management

**Dynamic Plugin Discovery**:

```python
class MCPPluginOrchestrator:
    """Manage multiple MCP plugins with intelligent routing."""
    
    def __init__(self, config_path: str):
        self.config = load_config(config_path)
        self.plugins = {}
        self.router = PluginRouter()
    
    def discover_plugins(self) -> dict[str, PluginInfo]:
        """
        Automatically discover available MCP plugins.
        
        Returns:
            Dictionary of plugin name to plugin information
        """
        discovered = {}
        
        # Scan configuration for declared plugins
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
    
    def detect_capabilities(self, plugin_config: dict) -> list[str]:
        """Detect plugin capabilities from configuration."""
        
        capabilities = []
        
        # Infer from command/args
        if "github" in plugin_config["command"]:
            capabilities.extend(["repository", "issues", "pull-requests"])
        elif "filesystem" in plugin_config["command"]:
            capabilities.extend(["file-read", "file-write", "directory-list"])
        elif "sqlite" in plugin_config["command"]:
            capabilities.extend(["database", "query", "transactions"])
        
        # Read from metadata if available
        if "capabilities" in plugin_config:
            capabilities.extend(plugin_config["capabilities"])
        
        return capabilities
    
    def route_request(self, request: PluginRequest) -> PluginResponse:
        """
        Route request to appropriate plugin.
        
        Args:
            request: Plugin request with capability requirements
        
        Returns:
            Plugin response
        """
        # Find plugins matching required capabilities
        matching_plugins = [
            name for name, info in self.plugins.items()
            if all(cap in info.capabilities for cap in request.required_capabilities)
        ]
        
        if not matching_plugins:
            raise NoPluginAvailable(f"No plugin supports: {request.required_capabilities}")
        
        # Select best plugin (load balancing, health, performance)
        selected_plugin = self.select_best_plugin(matching_plugins)
        
        # Execute request
        return self.execute_plugin_request(selected_plugin, request)
```

**Plugin Configuration Management**:

```json
{
  "mcpServers": {
    "github-integration": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-server-github"],
      "oauth": {
        "clientId": "${GITHUB_CLIENT_ID}",
        "clientSecret": "${GITHUB_CLIENT_SECRET}",
        "scopes": ["repo", "issues", "pull_requests"]
      },
      "capabilities": ["repository", "issues", "pull-requests", "workflows"],
      "performance": {
        "cache_ttl": 300,
        "rate_limit": 5000,
        "timeout": 30
      }
    },
    
    "filesystem-access": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-filesystem",
        "/allowed/path1",
        "/allowed/path2"
      ],
      "capabilities": ["file-read", "file-write", "directory-list"],
      "security": {
        "allowed_paths": ["/allowed/path1", "/allowed/path2"],
        "denied_patterns": ["*.env", "*.key", ".git/*"],
        "max_file_size": 10485760
      }
    },
    
    "database-connector": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-sqlite", "/data/app.db"],
      "capabilities": ["database", "query", "transactions"],
      "performance": {
        "connection_pool": 10,
        "query_timeout": 5,
        "cache_enabled": true
      }
    }
  }
}
```

**Performance Monitoring**:

```python
class PluginPerformanceMonitor:
    """Monitor plugin performance and health."""
    
    def __init__(self):
        self.metrics = {}
    
    def track_request(self, plugin_name: str, duration_ms: float, success: bool):
        """Track individual request metrics."""
        
        if plugin_name not in self.metrics:
            self.metrics[plugin_name] = PluginMetrics()
        
        metrics = self.metrics[plugin_name]
        metrics.total_requests += 1
        metrics.total_duration_ms += duration_ms
        
        if success:
            metrics.successful_requests += 1
        else:
            metrics.failed_requests += 1
        
        # Update percentiles
        metrics.update_percentiles(duration_ms)
    
    def get_health_report(self, plugin_name: str) -> HealthReport:
        """Generate plugin health report."""
        
        metrics = self.metrics.get(plugin_name)
        if not metrics:
            return HealthReport(status="unknown", reason="No metrics available")
        
        # Calculate health indicators
        success_rate = metrics.successful_requests / metrics.total_requests
        avg_latency = metrics.total_duration_ms / metrics.total_requests
        p99_latency = metrics.get_percentile(99)
        
        # Determine health status
        if success_rate < 0.95:
            status = "unhealthy"
            reason = f"Success rate: {success_rate:.1%} < 95%"
        elif p99_latency > 5000:  # 5 seconds
            status = "degraded"
            reason = f"P99 latency: {p99_latency:.0f}ms > 5000ms"
        else:
            status = "healthy"
            reason = "All metrics within thresholds"
        
        return HealthReport(
            status=status,
            reason=reason,
            success_rate=success_rate,
            avg_latency_ms=avg_latency,
            p99_latency_ms=p99_latency
        )
```

### Part 3: Claude Code Integration

**Automatic MCP Server Generation**:

```python
async def generate_mcp_server(service_spec: ServiceSpec) -> GeneratedMCPServer:
    """
    Generate MCP server from service specification.
    
    Args:
        service_spec: Service requirements and API definition
    
    Returns:
        Generated MCP server code and configuration
    """
    # Get latest MCP patterns from Context7
    mcp_patterns = await context7.get_library_docs(
        context7_library_id="/modelcontextprotocol/servers",
        topic="MCP server development best practices FastMCP 2025",
        tokens=5000
    )
    
    # Analyze service specification
    tool_designs = design_tools_from_spec(service_spec)
    resource_designs = design_resources_from_spec(service_spec)
    
    # Generate code using templates
    server_code = generate_fastmcp_server(
        tools=tool_designs,
        resources=resource_designs,
        patterns=mcp_patterns
    )
    
    # Generate configuration
    config = generate_server_config(service_spec)
    
    # Generate tests
    tests = generate_server_tests(tool_designs, resource_designs)
    
    return GeneratedMCPServer(
        server_code=server_code,
        config=config,
        tests=tests,
        deployment_guide=generate_deployment_docs(service_spec)
    )
```

**Smart Plugin Discovery**:

```python
class SmartPluginDiscovery:
    """ML-based plugin recommendation system."""
    
    async def recommend_plugins(self, task_description: str) -> list[PluginRecommendation]:
        """
        Recommend plugins based on task requirements.
        
        Args:
            task_description: Natural language task description
        
        Returns:
            Ranked list of plugin recommendations
        """
        # Analyze task requirements
        requirements = extract_requirements(task_description)
        
        # Get plugin catalog with capabilities
        available_plugins = await self.get_plugin_catalog()
        
        # Match plugins to requirements
        candidates = []
        for plugin in available_plugins:
            match_score = calculate_match_score(plugin.capabilities, requirements)
            if match_score > 0.5:
                candidates.append(PluginRecommendation(
                    plugin=plugin,
                    match_score=match_score,
                    reasoning=explain_match(plugin, requirements)
                ))
        
        # Rank by score
        candidates.sort(key=lambda x: x.match_score, reverse=True)
        
        return candidates[:5]  # Top 5 recommendations
```

### Context7 Integration for MCP Patterns

**Fetch Latest MCP Standards**:

```python
async def get_mcp_standards() -> MCPStandards:
    """Fetch latest MCP development standards from Context7."""
    
    # Get official MCP documentation
    mcp_docs = await context7.get_library_docs(
        context7_library_id="/modelcontextprotocol/servers",
        topic="MCP protocol specification tool design best practices 2025",
        tokens=5000
    )
    
    # Get FastMCP framework patterns
    fastmcp_docs = await context7.get_library_docs(
        context7_library_id="/fastmcp",
        topic="FastMCP server development authentication deployment 2025",
        tokens=3000
    )
    
    return MCPStandards(
        protocol_spec=mcp_docs["protocol"],
        tool_design_patterns=mcp_docs["tool_design"],
        fastmcp_best_practices=fastmcp_docs["best_practices"],
        authentication_patterns=fastmcp_docs["authentication"],
        deployment_strategies=fastmcp_docs["deployment"]
    )
```

---

## Advanced Patterns

### Multi-Protocol Support

**Protocol Abstraction Layer**:

```python
class ProtocolAdapter:
    """Support multiple transport protocols."""
    
    def __init__(self, server: FastMCP):
        self.server = server
        self.transports = {}
    
    def add_transport(self, name: str, transport: Transport):
        """Register transport protocol."""
        self.transports[name] = transport
    
    async def serve(self, protocol: str = "stdio"):
        """Start server with specified protocol."""
        
        if protocol == "stdio":
            await self.serve_stdio()
        elif protocol == "http":
            await self.serve_http()
        elif protocol == "websocket":
            await self.serve_websocket()
        else:
            raise UnsupportedProtocol(f"Protocol {protocol} not supported")
    
    async def serve_http(self, host: str = "0.0.0.0", port: int = 8000):
        """Serve over HTTP with SSE support."""
        from fastapi import FastAPI
        
        app = FastAPI()
        
        @app.post("/tools/{tool_name}")
        async def invoke_tool(tool_name: str, params: dict):
            return await self.server.invoke_tool(tool_name, params)
        
        @app.get("/resources/{resource_uri:path}")
        async def get_resource(resource_uri: str):
            return await self.server.get_resource(resource_uri)
        
        import uvicorn
        await uvicorn.run(app, host=host, port=port)
```

### Security & Compliance

**Security Hardening**:

```python
class MCPSecurityManager:
    """Enforce security policies for MCP servers."""
    
    def __init__(self):
        self.policies = load_security_policies()
    
    def validate_tool_invocation(self, tool_name: str, params: dict, context: RequestContext):
        """Validate tool invocation against security policies."""
        
        # Check authentication
        if not context.is_authenticated:
            raise Unauthorized("Authentication required")
        
        # Check authorization
        if not self.has_permission(context.user, tool_name):
            raise Forbidden(f"No permission for tool: {tool_name}")
        
        # Validate parameters
        for key, value in params.items():
            if self.is_dangerous_parameter(key, value):
                raise SecurityViolation(f"Dangerous parameter: {key}")
        
        # Check rate limits
        if self.is_rate_limited(context.user, tool_name):
            raise RateLimitExceeded("Too many requests")
        
        return True
    
    def is_dangerous_parameter(self, key: str, value: Any) -> bool:
        """Detect potentially dangerous parameters."""
        
        # Check for SQL injection patterns
        if isinstance(value, str) and ("DROP TABLE" in value.upper() or "--" in value):
            return True
        
        # Check for path traversal
        if isinstance(value, str) and (".." in value or value.startswith("/")):
            return True
        
        # Check for command injection
        dangerous_chars = [";", "|", "&", "`", "$"]
        if isinstance(value, str) and any(char in value for char in dangerous_chars):
            return True
        
        return False
```

### Monitoring & Observability

**Comprehensive Monitoring**:

```python
class MCPMonitoring:
    """Enterprise monitoring for MCP infrastructure."""
    
    def __init__(self):
        self.metrics_collector = MetricsCollector()
        self.alerting = AlertingSystem()
    
    def track_operation(self, operation_type: str, plugin_name: str, duration_ms: float, success: bool):
        """Track individual operations."""
        
        # Collect metrics
        self.metrics_collector.record({
            "operation": operation_type,
            "plugin": plugin_name,
            "duration_ms": duration_ms,
            "success": success,
            "timestamp": datetime.now()
        })
        
        # Check thresholds and alert
        if duration_ms > 5000:
            self.alerting.send_alert(
                severity="warning",
                message=f"{plugin_name} operation slow: {duration_ms:.0f}ms"
            )
        
        if not success:
            self.alerting.send_alert(
                severity="error",
                message=f"{plugin_name} operation failed"
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

## Works Well With

- `moai-cc-configuration` - MCP server configuration management
- `moai-essentials-debug` - MCP server debugging
- `moai-foundation-trust` - Security and compliance validation
- `moai-context7-integration` - Latest MCP standards and patterns
- Context7 MCP - Official MCP protocol documentation

---

## Changelog

- **v1.0.0** (2025-11-21): Initial consolidated MCP integration skill combining moai-cc-mcp-builder, moai-cc-mcp-plugins, and moai-mcp-builder with Context7 integration

---

**End of Skill** | Updated 2025-11-21
