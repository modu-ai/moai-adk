        
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


## Works Well With

- `moai-cc-configuration` - MCP server configuration management
- `moai-essentials-debug` - MCP server debugging
- `moai-foundation-trust` - Security and compliance validation
- `moai-context7-integration` - Latest MCP standards and patterns
- Context7 MCP - Official MCP protocol documentation


## Changelog

- **v1.0.0** (2025-11-21): Initial consolidated MCP integration skill combining moai-cc-mcp-builder, moai-cc-mcp-plugins, and moai-mcp-builder with Context7 integration


**End of Skill** | Updated 2025-11-21
