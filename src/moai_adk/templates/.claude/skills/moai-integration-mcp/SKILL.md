---
name: moai-integration-mcp
description: Universal MCP integration specialist combining Figma, Notion, Nano-Banana, and custom connectors with comprehensive orchestration patterns
version: 1.0.0
category: integration
tags:
 - mcp
 - figma
 - notion
 - nano-banana
 - connectors
 - orchestration
updated: 2025-11-30
status: active
author: MoAI-ADK Team
---

# MCP Integration Specialist

## Quick Reference (30 seconds)

Universal MCP Integration - Comprehensive MCP (Model Context Protocol) specialist combining Figma design integration, Notion knowledge management, Nano-Banana AI services, and custom connector frameworks with advanced orchestration capabilities.

Core Capabilities:
- Figma Integration: Design system extraction, component generation, token synchronization
- Notion Integration: Database queries, page management, knowledge extraction
- Nano-Banana AI: Content generation, analysis, AI-powered workflows
- Universal Connectors: Extensible framework for custom service integrations
- Multi-Service Orchestration: Complex workflows across multiple platforms
- Enterprise Security: OAuth, credential management, secure authentication

When to Use:
- Integrating multiple external services via MCP
- Building automated design-to-code workflows
- Creating AI-powered content pipelines
- Implementing cross-platform data synchronization
- Developing custom MCP connectors

---

## Implementation Guide (5 minutes)

### Quick Start Workflow

Universal MCP Server Setup:
```python
from moai_integration_mcp import UniversalMCPServer, ServiceOrchestrator

# Initialize universal MCP server
mcp_server = UniversalMCPServer("integration-server")

# Configure connectors
mcp_server.setup_connectors({
 'figma': {'api_key': os.getenv('FIGMA_TOKEN')},
 'notion': {'api_key': os.getenv('NOTION_TOKEN')},
 'nano_banana': {'api_key': os.getenv('NANO_BANANA_TOKEN')}
})

# Register orchestration tools
orchestrator = ServiceOrchestrator(mcp_server)
orchestrator.register_workflows()

# Start server
mcp_server.start(port=3000)
```

Multi-Service Workflow:
```bash
# Design system automation
mcp-tools design_to_code --figma-file "abc123" --output ./src/components

# Knowledge extraction workflow
mcp-tools knowledge_extraction --notion-db "xyz789" --analyze "best_practices"

# AI-powered content generation
mcp-tools ai_workflow --input "./docs/" --output "./generated/" --model "claude-3-5-sonnet"
```

### Core Components

1. Server Architecture (`modules/server-architecture.md`)
- Universal MCP server framework
- Multi-connector management
- Dynamic tool registration
- Configuration and initialization

2. Integration Patterns (`modules/integration-patterns.md`)
- Multi-service orchestration
- Workflow engine and templates
- Data transformation pipelines
- Advanced integration patterns

3. Security & Authentication (`modules/security-authentication.md`)
- OAuth 2.0 flows for all services
- Secure credential storage
- Token management and refresh
- Access control and permissions

4. Error Handling (`modules/error-handling.md`)
- Circuit breaker patterns
- Retry logic with backoff
- Fault tolerance mechanisms
- Monitoring and observability

---

## Advanced Patterns (10+ minutes)

### Multi-Service Orchestration

Design-to-Code Pipeline:
```python
async def complete_design_workflow(figma_file_id: str, target_library: str = "shadcn"):
 """Complete design system to production code workflow."""

 # Phase 1: Extract design data
 design_data = await mcp_server.invoke_tool("extract_figma_components", {
 "file_id": figma_file_id,
 "include_tokens": True
 })

 # Phase 2: Process with AI
 component_specs = []
 for component in design_data["components"]:
 spec = await mcp_server.invoke_tool("analyze_with_ai", {
 "content": json.dumps(component),
 "analysis_type": "component_specification"
 })
 component_specs.append(spec)

 # Phase 3: Generate code
 generated_components = []
 for spec in component_specs:
 code = await mcp_server.invoke_tool("generate_ai_content", {
 "prompt": f"Generate React component for: {spec['analysis']}",
 "max_tokens": 3000
 })
 generated_components.append(code)

 # Phase 4: Create documentation
 documentation = await mcp_server.invoke_tool("generate_ai_content", {
 "prompt": f"Create documentation for components: {json.dumps(component_specs)}",
 "max_tokens": 4000
 })

 return {
 "components": generated_components,
 "documentation": documentation,
 "design_tokens": design_data["design_tokens"],
 "workflow_status": "completed"
 }
```

Knowledge Base Automation:
```python
async def knowledge_base_workflow(notion_database: str, analysis_goals: list):
 """Automated knowledge extraction and organization workflow."""

 # Extract content from Notion
 content = await mcp_server.invoke_tool("query_notion_database", {
 "database_id": notion_database,
 "query": {"filter": {"property": "Status", "select": {"equals": "Published"}}}
 })

 # Analyze with AI for each goal
 analyses = {}
 for goal in analysis_goals:
 analysis = await mcp_server.invoke_tool("analyze_with_ai", {
 "content": json.dumps(content["results"]),
 "analysis_type": goal
 })
 analyses[goal] = analysis

 # Structure knowledge base
 structured_kb = await mcp_server.invoke_tool("generate_ai_content", {
 "prompt": f"Create structured knowledge base from analyses: {json.dumps(analyses)}",
 "max_tokens": 5000
 })

 return {
 "raw_content": content,
 "analyses": analyses,
 "structured_knowledge": structured_kb,
 "source_count": len(content["results"])
 }
```

### Custom Connector Development

Extensible Connector Framework:
```python
class CustomConnector:
 def __init__(self, service_config: dict):
 self.config = service_config
 self.client = None

 async def initialize(self):
 """Initialize custom service client."""
 self.client = CustomServiceClient(self.config)

 def register_tools(self, server):
 """Register connector-specific tools."""

 @server.tool()
 async def custom_service_operation(
 operation_type: str,
 parameters: dict = {}
 ) -> dict:
 """Execute operation on custom service."""
 try:
 result = await self.client.execute_operation(
 operation_type,
 parameters
 )

 return {
 "status": "success",
 "result": result,
 "operation": operation_type
 }

 except Exception as e:
 return {
 "status": "error",
 "error": str(e),
 "operation": operation_type
 }

# Register custom connector
mcp_server.register_connector('custom_service', CustomConnector(config))
```

---

## Works Well With

Complementary Skills:
- `moai-domain-frontend` - Frontend component generation and integration
- `moai-domain-backend` - Backend API integration patterns
- `moai-docs-generation` - Automated documentation workflows
- `moai-foundation-claude` - Claude Code integration patterns

External Services:
- Figma (design systems, component extraction)
- Notion (knowledge management, documentation)
- Nano-Banana (AI content generation)
- Custom APIs and web services
- Database systems and storage

Integration Platforms:
- FastMCP server framework
- OAuth 2.0 providers
- REST APIs and GraphQL
- Message queues and event systems
- Cloud storage services

---

## Usage Examples

### Design System Integration
```python
# Extract and sync design tokens
tokens = await mcp_server.invoke_tool("sync_figma_tokens", {
 "file_id": "design-system-file",
 "output_format": "typescript",
 "include_variants": True
})

# Generate component library
components = await mcp_server.invoke_tool("extract_figma_components", {
 "file_id": "component-library",
 "target_framework": "react",
 "include_stories": True
})
```

### Knowledge Base Management
```python
# Extract and analyze knowledge
analysis = await mcp_server.invoke_tool("knowledge_extraction_workflow", {
 "notion_database_id": "knowledge-base",
 "analysis_goals": ["best_practices", "patterns", "action_items"],
 "output_format": "structured_json"
})

# Create new documentation
doc_page = await mcp_server.invoke_tool("create_notion_page", {
 "database_id": "documentation-db",
 "properties": {
 "Title": {"title": [{"text": {"content": "Best Practices Guide"}}]},
 "Category": {"select": {"name": "Guidelines"}}
 },
 "content": analysis["structured_knowledge"]
})
```

### AI-Powered Workflows
```python
# Generate content with AI
ai_content = await mcp_server.invoke_tool("generate_ai_content", {
 "prompt": "Create comprehensive API documentation",
 "model": "claude-3-5-sonnet",
 "max_tokens": 4000,
 "temperature": 0.7
})

# Analyze and summarize
summary = await mcp_server.invoke_tool("analyze_with_ai", {
 "content": ai_content["content"],
 "analysis_type": "summary",
 "include_key_points": True
})
```

---

## Technology Stack

Core Framework:
- FastMCP (Python MCP server framework)
- AsyncIO for concurrent operations
- Pydantic for data validation
- HTTPX for HTTP client operations

Service Integrations:
- Figma API (design systems)
- Notion API (knowledge management)
- Nano-Banana API (AI services)
- Custom REST/GraphQL APIs

Security & Authentication:
- OAuth 2.0 implementation
- Cryptography for encryption
- JWT token management
- Secure credential storage

Error Handling & Reliability:
- Circuit breaker patterns
- Retry mechanisms with backoff
- Comprehensive error classification
- Monitoring and observability

Development Tools:
- Type hints and validation
- Comprehensive logging
- Performance monitoring
- Debugging and profiling tools

---

*For detailed implementation patterns, connector development, and advanced workflows, see the `modules/` directory.*
