# moai-integration-mcp Reference

## API Reference

### Universal MCP Server API

Server Initialization:
- `UniversalMCPServer(name: str)`: Create MCP server instance
- `setup_connectors(config: dict)`: Configure service connectors
- `register_workflows()`: Register orchestration tools
- `start(port: int)`: Start server on specified port

Tool Invocation:
- `invoke_tool(tool_name: str, params: dict)`: Execute MCP tool
- `list_tools()`: Get available tools
- `get_tool_schema(tool_name: str)`: Get tool parameter schema

### Figma Integration API

Design Extraction:
- `extract_figma_components(file_id, include_tokens)`: Extract components
- `sync_figma_tokens(file_id, output_format, include_variants)`: Sync design tokens
- `get_figma_file(file_id)`: Get complete file data
- `get_figma_node(file_id, node_id)`: Get specific node

Output Formats:
- `typescript`: TypeScript type definitions
- `css`: CSS custom properties
- `json`: Raw JSON structure
- `scss`: SCSS variables

### Notion Integration API

Database Operations:
- `query_notion_database(database_id, query)`: Query database
- `create_notion_page(database_id, properties, content)`: Create page
- `update_notion_page(page_id, properties)`: Update page
- `get_notion_page(page_id)`: Get page content

Query Filter Structure:
```json
{
  "filter": {
    "property": "PropertyName",
    "select": {"equals": "Value"}
  },
  "sorts": [
    {"property": "Date", "direction": "descending"}
  ]
}
```

### Nano-Banana AI API

Content Generation:
- `generate_ai_content(prompt, model, max_tokens, temperature)`: Generate content
- `analyze_with_ai(content, analysis_type, include_key_points)`: Analyze content
- `summarize_content(content, max_length)`: Create summary

Analysis Types:
- `component_specification`: UI component analysis
- `summary`: Content summarization
- `best_practices`: Extract best practices
- `patterns`: Identify design patterns
- `action_items`: Extract actionable items

### Custom Connector API

Connector Interface:
- `initialize()`: Set up service client
- `register_tools(server)`: Register connector tools
- `execute_operation(operation_type, parameters)`: Run operation

Registration:
- `register_connector(name, connector_instance)`: Add to server

---

## Configuration Options

### Server Configuration

Environment Variables:
- `FIGMA_TOKEN`: Figma API access token
- `NOTION_TOKEN`: Notion API integration token
- `NANO_BANANA_TOKEN`: Nano-Banana API key
- `MCP_SERVER_PORT`: Server port (default: 3000)

Server Settings:
- `port`: HTTP server port
- `host`: Server host address
- `cors_enabled`: Enable CORS
- `rate_limit`: Requests per minute

### Connector Configuration

Figma Connector:
- `api_key`: Figma personal access token
- `team_id`: Optional team ID filter
- `cache_duration`: File cache duration

Notion Connector:
- `api_key`: Notion integration token
- `version`: API version (default: "2022-06-28")
- `timeout`: Request timeout

Nano-Banana Connector:
- `api_key`: Service API key
- `model`: Default model selection
- `max_tokens`: Default token limit

### Security Configuration

OAuth 2.0 Settings:
- `client_id`: OAuth client identifier
- `client_secret`: OAuth client secret
- `redirect_uri`: OAuth callback URL
- `scopes`: Required permission scopes

Token Management:
- `token_storage`: Storage location for tokens
- `auto_refresh`: Enable automatic token refresh
- `refresh_threshold`: Seconds before expiry to refresh

---

## Integration Patterns

### Design-to-Code Pipeline

Workflow Steps:
1. Extract design data from Figma
2. Process components with AI analysis
3. Generate React/Vue components
4. Create component documentation
5. Output design tokens

Code Example:
```python
async def design_to_code(figma_file_id):
    # Extract design
    design = await mcp_server.invoke_tool("extract_figma_components", {
        "file_id": figma_file_id,
        "include_tokens": True
    })

    # Analyze with AI
    specs = await mcp_server.invoke_tool("analyze_with_ai", {
        "content": json.dumps(design),
        "analysis_type": "component_specification"
    })

    # Generate code
    code = await mcp_server.invoke_tool("generate_ai_content", {
        "prompt": f"Generate React component: {specs}",
        "max_tokens": 3000
    })

    return {"code": code, "tokens": design["design_tokens"]}
```

### Knowledge Base Automation

Workflow Steps:
1. Query Notion database for content
2. Analyze content for specific goals
3. Structure into knowledge base format
4. Create or update documentation pages

Integration Points:
- Source: Notion databases and pages
- Processing: AI analysis and summarization
- Output: Structured documentation

### Multi-Service Orchestration

Coordination Pattern:
1. Initialize all required connectors
2. Define workflow sequence
3. Execute steps with error handling
4. Aggregate results
5. Handle partial failures gracefully

Error Recovery:
- Circuit breaker for failed services
- Retry with exponential backoff
- Fallback to cached data
- Graceful degradation

---

## Troubleshooting

### Connection Failures

Symptoms: Cannot connect to service, timeout errors

Solutions:
1. Verify API tokens are valid and not expired
2. Check network connectivity
3. Confirm service is available (status page)
4. Review rate limit status

### Authentication Errors

Symptoms: 401/403 errors, token rejected

Solutions:
1. Regenerate API token
2. Verify token permissions/scopes
3. Check token expiration
4. Confirm integration is properly configured

### Rate Limiting

Symptoms: 429 errors, requests rejected

Solutions:
1. Implement request throttling
2. Use caching to reduce calls
3. Batch operations where possible
4. Upgrade API tier if needed

### Data Extraction Issues

Symptoms: Missing data, incomplete extraction

Figma Solutions:
- Verify file access permissions
- Check node IDs are correct
- Confirm components are published

Notion Solutions:
- Verify database sharing settings
- Check property names match exactly
- Confirm integration has page access

### Orchestration Failures

Symptoms: Workflow stops midway, inconsistent state

Solutions:
1. Review error logs for failed step
2. Check circuit breaker status
3. Verify all services are responsive
4. Implement checkpoint/resume logic

---

## External Resources

### Official Documentation

- [Figma API Documentation](https://www.figma.com/developers/api)
- [Notion API Reference](https://developers.notion.com/reference)
- [FastMCP Framework](https://github.com/modelcontextprotocol/fastmcp)
- [OAuth 2.0 Specification](https://oauth.net/2/)

### MCP Resources

- Model Context Protocol Specification
- FastMCP Python SDK
- MCP Server Examples
- MCP Tool Development Guide

### Related Skills

- `moai-domain-frontend`: Frontend component generation
- `moai-domain-backend`: Backend API integration
- `moai-docs-generation`: Documentation automation
- `moai-foundation-claude`: Claude Code integration patterns

### Technology Stack

Core Framework:
- FastMCP (Python MCP server)
- AsyncIO (concurrent operations)
- Pydantic (data validation)
- HTTPX (HTTP client)

Security:
- OAuth 2.0 implementation
- Cryptography (encryption)
- JWT token management

Reliability:
- Circuit breaker patterns
- Retry mechanisms
- Monitoring and observability

---

Version: 1.0.0
Last Updated: 2025-12-06
