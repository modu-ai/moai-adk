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

---

## Implementation Modules

For detailed implementation:
- **Core Patterns**: modules/core.md
- **Advanced Patterns**: modules/advanced.md

---

    def __init__(self, config_path: str):
        self.config = load_config(config_path)
        self.plugins = {}
        self.router = PluginRouter()
    
    def discover_plugins(self) -> dict[str, PluginInfo]:
        """
        Automatically discover available MCP plugins.
        
        Returns:

**End of Skill** | Updated 2025-11-21
