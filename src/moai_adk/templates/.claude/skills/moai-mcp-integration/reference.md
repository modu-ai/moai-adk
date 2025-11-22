# MCP 1.0+ Reference Documentation

Complete MCP specification, official servers, and API reference.

---

## MCP Protocol Specification (v1.0+)

### Protocol Overview

MCP (Model Context Protocol) is a standardized protocol for exposing tools, resources, and prompts to Claude and other AI models.

**Protocol Version**: 1.0+
**Status**: Production Ready
**Transport**: stdin/stdout or HTTP(S)

### Core Architecture

```
Client (Claude/LLM)
      ↓
      MCP Protocol (JSON-RPC 2.0 based)
      ↓
MCP Server (Tool/Resource/Prompt Providers)
```

### Three Core Components

#### 1. Tools (Function Calls)
Functions that Claude can invoke to perform tasks.

**Specification**:
```json
{
  "name": "search_database",
  "description": "Search database with pagination",
  "inputSchema": {
    "type": "object",
    "properties": {
      "query": {
        "type": "string",
        "description": "Search query"
      },
      "limit": {
        "type": "integer",
        "description": "Max results",
        "minimum": 1,
        "maximum": 100
      }
    },
    "required": ["query"]
  }
}
```

#### 2. Resources (Data Exposure)
URI-based data endpoints for exposing information.

**Specification**:
```json
{
  "uri": "database://{table}/{id}",
  "name": "Database Record",
  "description": "Fetch record by ID",
  "mimeType": "application/json"
}
```

#### 3. Prompts (Templates)
Reusable conversation templates and system prompts.

**Specification**:
```json
{
  "name": "data-analyst",
  "description": "System prompt for data analysis",
  "arguments": [
    {
      "name": "domain",
      "description": "Analysis domain",
      "required": true
    }
  ]
}
```

---

## Official MCP Servers (2025 Edition)

### Server Catalog

| Server | Purpose | Status | Official | Setup |
|--------|---------|--------|----------|-------|
| **Playwright** | Web automation & testing | Stable | ✅ | `npx @microsoft/mcp-server-playwright` |
| **Figma** | Design system access | Stable | ✅ | `npx @figma/mcp-server-figma` |
| **GitHub** | Repository management | Stable | ✅ | `npx @anthropic-ai/mcp-server-github` |
| **Notion** | Knowledge base integration | Stable | ✅ | `npx @notionhq/mcp-server-notion` |
| **Firebase** | Backend-as-a-Service | Stable | ✅ | `npx @firebase/mcp-server-firebase` |
| **Supabase** | Open-source backend | Stable | ✅ | `npx @supabase/mcp-server-supabase` |
| **PlanetScale** | MySQL database hosting | Stable | ✅ | `npx @planetscale/mcp-server-planetscale` |
| **Stripe** | Payment processing | Stable | ✅ | `npx @stripe/mcp-server-stripe` |
| **Slack** | Messaging & collaboration | Stable | ✅ | `npx @slack/mcp-server-slack` |
| **Google Calendar** | Schedule management | Stable | ✅ | `npx @google/mcp-server-calendar` |

### Detailed Server Specifications

#### Playwright MCP

**Purpose**: Web automation, E2E testing, web scraping
**Package**: `@microsoft/mcp-server-playwright`

**Available Tools**:
- `navigate_to(url: string)` - Navigate to URL
- `click(selector: string)` - Click element
- `fill(selector: string, value: string)` - Fill input
- `get_text(selector: string)` - Get element text
- `get_attribute(selector: string, attribute: string)` - Get attribute
- `execute_script(script: string)` - Execute JavaScript
- `take_screenshot(path?: string)` - Capture screenshot
- `get_page_content()` - Get page HTML
- `wait_for_element(selector: string, timeout?: number)` - Wait for element
- `reload_page()` - Reload page

**Authentication**: None (local execution)

**Configuration**:
```json
{
  "mcpServers": {
    "playwright": {
      "command": "npx",
      "args": ["-y", "@microsoft/mcp-server-playwright"],
      "env": {
        "HEADLESS": "true",
        "BROWSER": "chromium"
      }
    }
  }
}
```

#### GitHub MCP

**Purpose**: Repository management, PR/issue automation
**Package**: `@anthropic-ai/mcp-server-github`

**Available Tools**:
- `create_issue(repo, title, body, labels)` - Create issue
- `create_pr(repo, title, body, base, head)` - Create pull request
- `list_issues(repo, state, labels)` - List issues
- `list_prs(repo, state)` - List pull requests
- `merge_pr(repo, pr_number)` - Merge pull request
- `update_issue(repo, issue_number, state, title, body)` - Update issue
- `add_comment(repo, issue_number, body)` - Comment on issue

**Authentication**: OAuth2 (GitHub Personal Access Token)

**Required Scopes**:
- `repo` - Repository access
- `issues` - Issue management
- `pull_requests` - PR management
- `workflows` - Actions workflows

**Configuration**:
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
      }
    }
  }
}
```

#### Notion MCP

**Purpose**: Knowledge base, documentation, database management
**Package**: `@notionhq/mcp-server-notion`

**Available Tools**:
- `query_database(database_id, filter, sort)` - Query database
- `create_page(parent_id, properties, children)` - Create page
- `update_page(page_id, properties)` - Update page
- `append_blocks(page_id, blocks)` - Add blocks to page
- `search(query)` - Search workspace

**Authentication**: Bearer token (Notion API key)

**Configuration**:
```json
{
  "mcpServers": {
    "notion": {
      "command": "npx",
      "args": ["-y", "@notionhq/mcp-server-notion"],
      "auth": {
        "type": "bearer",
        "token": "${NOTION_API_KEY}"
      }
    }
  }
}
```

#### Firebase MCP

**Purpose**: Real-time database, authentication, storage
**Package**: `@firebase/mcp-server-firebase`

**Available Tools**:
- `firestore_query(collection, filters)` - Query Firestore
- `firestore_create(collection, data)` - Create document
- `firestore_update(collection, doc_id, data)` - Update document
- `firestore_delete(collection, doc_id)` - Delete document
- `auth_create_user(email, password)` - Create user
- `auth_delete_user(uid)` - Delete user
- `storage_upload(path, file)` - Upload to storage
- `storage_download(path)` - Download from storage

**Authentication**: Service Account JSON

**Configuration**:
```json
{
  "mcpServers": {
    "firebase": {
      "command": "npx",
      "args": ["-y", "@firebase/mcp-server-firebase"],
      "auth": {
        "type": "service-account",
        "credential_path": "${FIREBASE_SERVICE_ACCOUNT_PATH}"
      }
    }
  }
}
```

---

## FastMCP API Reference

### Server Class

```python
from fastmcp import FastMCP

server = FastMCP(
    name: str,
    version: str = "1.0.0",
    description: str = ""
)
```

### Decorators

#### @server.tool()
Register a tool (callable function).

```python
@server.tool()
def my_tool(arg1: str, arg2: int = 10) -> dict:
    """Tool description."""
    return {"result": "value"}
```

**Parameters**:
- `arg: type` - Parameter with type annotation (required)
- `arg: type = default` - Optional parameter with default
- `arg: Type[Enum]` - Constrained choices
- `Field(...)` - Pydantic field with constraints

**Return Types**:
- `str`, `int`, `float`, `bool`
- `list[T]`, `dict[str, T]`
- `Literal["choice1", "choice2"]` - Constrained values
- `Optional[T]` - Nullable types

#### @server.resource(uri: str)
Register a resource endpoint.

```python
@server.resource("db://{table}/{id}")
def get_record(table: str, id: str) -> dict:
    """Fetch record resource."""
    return fetch_data(table, id)
```

**URI Pattern Syntax**:
- `{name}` - String parameter
- `{id:int}` - Typed parameter
- Static/dynamic mixing: `data://users/{user_id}/profile`

#### @server.prompt(name: str)
Register a prompt template.

```python
@server.prompt("analyst")
def analyst_prompt(domain: str) -> str:
    """Get analyst system prompt."""
    return f"You are a {domain} analyst..."
```

### Running the Server

```python
# Stdio mode (default)
if __name__ == "__main__":
    server.run()

# HTTP mode
if __name__ == "__main__":
    server.run_http(host="0.0.0.0", port=8000)
```

### Error Handling

```python
from fastmcp import MCPError

@server.tool()
def risky_operation():
    try:
        result = execute_risky_task()
    except ValueError as e:
        raise MCPError(f"Invalid input: {e}")
    except Exception as e:
        raise MCPError(f"Operation failed: {e}")
    return result
```

---

## Configuration Schema

### Full Configuration Structure

```json
{
  "mcpServers": {
    "server-name": {
      "command": "python|npx|node|...",
      "args": ["arg1", "arg2"],
      "env": {
        "ENV_VAR": "value"
      },
      "cwd": "/path/to/directory",
      "capabilities": ["tools", "resources", "prompts"],
      "auth": {
        "type": "oauth2|api-key|bearer|service-account",
        "clientId": "${ENV_VAR}",
        "clientSecret": "${ENV_VAR}"
      },
      "timeouts": {
        "initialization_ms": 10000,
        "method_call_ms": 30000
      }
    }
  }
}
```

### Environment Variables

**Substitution Syntax**: `${VAR_NAME}`

**Common Variables**:
- `${GITHUB_TOKEN}` - GitHub PAT
- `${NOTION_API_KEY}` - Notion API key
- `${FIREBASE_SERVICE_ACCOUNT_PATH}` - Firebase credentials
- `${FIGMA_CLIENT_ID}` / `${FIGMA_CLIENT_SECRET}` - Figma OAuth

---

## Troubleshooting Guide

### Server Won't Start

**Issue**: "Command not found"
**Solution**: Ensure command is in PATH or use full path

```json
{
  "command": "/usr/local/bin/python",
  "args": ["server.py"]
}
```

### Tool Not Appearing in Claude

**Issue**: Tools registered but Claude can't see them
**Solutions**:
1. Check tool has proper docstring
2. Verify return type annotation
3. Ensure no syntax errors in schema
4. Restart server

### Authentication Failing

**Issue**: "Invalid credentials"
**Solutions**:
1. Verify environment variables set correctly
2. Check credential permissions/scopes
3. Ensure token not expired
4. Check authentication type matches server

### Performance Issues

**Issue**: Tools are slow
**Solutions**:
1. Add pagination for large results
2. Use connection pooling
3. Implement caching where appropriate
4. Monitor with metrics collection

### Memory Leaks

**Issue**: Server memory grows over time
**Solutions**:
1. Close resources in finally blocks
2. Limit cache sizes
3. Use context managers
4. Monitor with tools like `memory_profiler`

---

## Best Practices Summary

### Tool Design
- ✅ One tool = one meaningful task
- ✅ Include pagination limits
- ✅ Validate all inputs
- ✅ Return consistent structures
- ✅ Provide clear error messages

### Authentication
- ✅ Never hardcode credentials
- ✅ Use environment variables
- ✅ Rotate keys regularly
- ✅ Implement per-tool auth where needed
- ✅ Log failed auth attempts

### Performance
- ✅ Set appropriate timeouts
- ✅ Stream large results
- ✅ Cache when appropriate
- ✅ Monitor latency metrics
- ✅ Implement circuit breakers

### Deployment
- ✅ Use container orchestration (K8s)
- ✅ Implement health checks
- ✅ Set resource limits
- ✅ Enable logging and metrics
- ✅ Plan for graceful shutdown

---

## Resources

- **Official Spec**: https://modelcontextprotocol.io/spec
- **FastMCP Docs**: https://fastmcp.dev/
- **Official Servers**: https://github.com/modelcontextprotocol/
- **Community**: https://discord.gg/anthropic

---

**Last Updated**: 2025-11-22 | MCP Protocol v1.0+
