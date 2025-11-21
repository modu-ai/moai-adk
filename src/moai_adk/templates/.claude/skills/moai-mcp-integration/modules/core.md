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
