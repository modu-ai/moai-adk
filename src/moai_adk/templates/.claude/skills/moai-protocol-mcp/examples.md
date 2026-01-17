# MCP Implementation Examples

## TypeScript Server Examples

### Basic Resource Server

This example demonstrates a simple MCP server exposing file resources.

```typescript
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import {
  ListResourcesRequestSchema,
  ReadResourceRequestSchema,
} from "@modelcontextprotocol/sdk/types.js";
import * as fs from "fs/promises";
import * as path from "path";

const server = new Server(
  {
    name: "file-resource-server",
    version: "1.0.0",
  },
  {
    capabilities: {
      resources: {},
    },
  },
);

// List available resources
server.setRequestHandler(ListResourcesRequestSchema, async () => {
  const files = await fs.readdir("./data");
  return {
    resources: files.map((file) => ({
      uri: `file:///${file}`,
      name: file,
      mimeType: "text/plain",
    })),
  };
});

// Read resource content
server.setRequestHandler(ReadResourceRequestSchema, async (request) => {
  const uri = request.params.uri;
  const filename = uri.replace("file:///", "");
  const content = await fs.readFile(path.join("./data", filename), "utf-8");

  return {
    contents: [
      {
        uri,
        mimeType: "text/plain",
        text: content,
      },
    ],
  };
});

// Start server
const transport = new StdioServerTransport();
await server.connect(transport);
```

### Tool Server with Database Access

This example shows a server exposing database query tools.

```typescript
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import {
  ListToolsRequestSchema,
  CallToolRequestSchema,
} from "@modelcontextprotocol/sdk/types.js";
import Database from "better-sqlite3";

const db = new Database("./app.db");

const server = new Server(
  {
    name: "database-tool-server",
    version: "1.0.0",
  },
  {
    capabilities: {
      tools: {},
    },
  },
);

// List available tools
server.setRequestHandler(ListToolsRequestSchema, async () => {
  return {
    tools: [
      {
        name: "query_users",
        description: "Query users from the database with optional filters",
        inputSchema: {
          type: "object",
          properties: {
            status: {
              type: "string",
              description: "Filter by user status",
              enum: ["active", "inactive", "pending"],
            },
            limit: {
              type: "number",
              description: "Maximum number of results",
              default: 10,
            },
          },
        },
      },
      {
        name: "get_user_by_id",
        description: "Get a specific user by their ID",
        inputSchema: {
          type: "object",
          properties: {
            id: {
              type: "number",
              description: "The user ID",
            },
          },
          required: ["id"],
        },
      },
    ],
  };
});

// Handle tool calls
server.setRequestHandler(CallToolRequestSchema, async (request) => {
  const { name, arguments: args } = request.params;

  switch (name) {
    case "query_users": {
      const status = args?.status;
      const limit = args?.limit ?? 10;

      let query = "SELECT * FROM users";
      const params: unknown[] = [];

      if (status) {
        query += " WHERE status = ?";
        params.push(status);
      }

      query += " LIMIT ?";
      params.push(limit);

      const users = db.prepare(query).all(...params);

      return {
        content: [
          {
            type: "text",
            text: JSON.stringify(users, null, 2),
          },
        ],
      };
    }

    case "get_user_by_id": {
      const user = db.prepare("SELECT * FROM users WHERE id = ?").get(args.id);

      if (!user) {
        return {
          content: [
            {
              type: "text",
              text: `User with ID ${args.id} not found`,
            },
          ],
          isError: true,
        };
      }

      return {
        content: [
          {
            type: "text",
            text: JSON.stringify(user, null, 2),
          },
        ],
      };
    }

    default:
      throw new Error(`Unknown tool: ${name}`);
  }
});

const transport = new StdioServerTransport();
await server.connect(transport);
```

### Prompt Server

This example demonstrates a server providing reusable prompts.

```typescript
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import {
  ListPromptsRequestSchema,
  GetPromptRequestSchema,
} from "@modelcontextprotocol/sdk/types.js";

const server = new Server(
  {
    name: "prompt-server",
    version: "1.0.0",
  },
  {
    capabilities: {
      prompts: {},
    },
  },
);

const prompts = {
  "code-review": {
    name: "code-review",
    description: "Perform a code review with specific focus areas",
    arguments: [
      {
        name: "language",
        description: "Programming language of the code",
        required: true,
      },
      {
        name: "focus",
        description: "Focus area: security, performance, or readability",
        required: false,
      },
    ],
  },
  "explain-code": {
    name: "explain-code",
    description: "Explain code in simple terms",
    arguments: [
      {
        name: "audience",
        description: "Target audience: beginner, intermediate, or expert",
        required: false,
      },
    ],
  },
};

server.setRequestHandler(ListPromptsRequestSchema, async () => {
  return {
    prompts: Object.values(prompts),
  };
});

server.setRequestHandler(GetPromptRequestSchema, async (request) => {
  const { name, arguments: args } = request.params;

  if (name === "code-review") {
    const language = args?.language ?? "unknown";
    const focus = args?.focus ?? "general";

    return {
      messages: [
        {
          role: "user",
          content: {
            type: "text",
            text: `Please review the following ${language} code with a focus on ${focus}.

Provide:
1. Summary of what the code does
2. Issues found (if any)
3. Suggestions for improvement
4. Security considerations`,
          },
        },
      ],
    };
  }

  if (name === "explain-code") {
    const audience = args?.audience ?? "intermediate";

    return {
      messages: [
        {
          role: "user",
          content: {
            type: "text",
            text: `Please explain the following code for a ${audience} developer.

Include:
- What the code does
- Key concepts used
- How the parts work together`,
          },
        },
      ],
    };
  }

  throw new Error(`Unknown prompt: ${name}`);
});

const transport = new StdioServerTransport();
await server.connect(transport);
```

---

## Python Server Examples

### Basic Python MCP Server

```python
import asyncio
from mcp.server import Server
from mcp.server.stdio import stdio_server
from mcp.types import Resource, Tool, TextContent

server = Server("python-mcp-server")

# Resource handlers
@server.list_resources()
async def list_resources() -> list[Resource]:
    return [
        Resource(
            uri="config://app",
            name="Application Configuration",
            mimeType="application/json"
        ),
        Resource(
            uri="status://health",
            name="Health Status",
            mimeType="application/json"
        )
    ]

@server.read_resource()
async def read_resource(uri: str) -> str:
    if uri == "config://app":
        return '{"app_name": "MyApp", "version": "1.0.0"}'
    elif uri == "status://health":
        return '{"status": "healthy", "uptime": 3600}'
    raise ValueError(f"Unknown resource: {uri}")

# Tool handlers
@server.list_tools()
async def list_tools() -> list[Tool]:
    return [
        Tool(
            name="calculate",
            description="Perform mathematical calculations",
            inputSchema={
                "type": "object",
                "properties": {
                    "expression": {
                        "type": "string",
                        "description": "Mathematical expression to evaluate"
                    }
                },
                "required": ["expression"]
            }
        )
    ]

@server.call_tool()
async def call_tool(name: str, arguments: dict) -> list[TextContent]:
    if name == "calculate":
        expression = arguments["expression"]
        # Safe evaluation for basic math
        allowed_chars = set("0123456789+-*/.()")
        if not all(c in allowed_chars or c.isspace() for c in expression):
            return [TextContent(
                type="text",
                text=f"Error: Invalid characters in expression"
            )]

        try:
            result = eval(expression)
            return [TextContent(type="text", text=str(result))]
        except Exception as e:
            return [TextContent(type="text", text=f"Error: {str(e)}")]

    raise ValueError(f"Unknown tool: {name}")

async def main():
    async with stdio_server() as (read_stream, write_stream):
        await server.run(read_stream, write_stream)

if __name__ == "__main__":
    asyncio.run(main())
```

### Python Server with API Integration

```python
import asyncio
import httpx
from mcp.server import Server
from mcp.server.stdio import stdio_server
from mcp.types import Tool, TextContent

server = Server("api-integration-server")

@server.list_tools()
async def list_tools() -> list[Tool]:
    return [
        Tool(
            name="search_web",
            description="Search the web using a search API",
            inputSchema={
                "type": "object",
                "properties": {
                    "query": {
                        "type": "string",
                        "description": "Search query"
                    },
                    "max_results": {
                        "type": "number",
                        "description": "Maximum number of results",
                        "default": 5
                    }
                },
                "required": ["query"]
            }
        ),
        Tool(
            name="get_weather",
            description="Get current weather for a location",
            inputSchema={
                "type": "object",
                "properties": {
                    "location": {
                        "type": "string",
                        "description": "City name or coordinates"
                    }
                },
                "required": ["location"]
            }
        )
    ]

@server.call_tool()
async def call_tool(name: str, arguments: dict) -> list[TextContent]:
    async with httpx.AsyncClient() as client:
        if name == "search_web":
            query = arguments["query"]
            max_results = arguments.get("max_results", 5)

            # Example API call (replace with actual search API)
            response = await client.get(
                "https://api.example.com/search",
                params={"q": query, "limit": max_results}
            )

            return [TextContent(
                type="text",
                text=response.text
            )]

        if name == "get_weather":
            location = arguments["location"]

            # Example API call (replace with actual weather API)
            response = await client.get(
                "https://api.example.com/weather",
                params={"location": location}
            )

            return [TextContent(
                type="text",
                text=response.text
            )]

    raise ValueError(f"Unknown tool: {name}")

async def main():
    async with stdio_server() as (read_stream, write_stream):
        await server.run(read_stream, write_stream)

if __name__ == "__main__":
    asyncio.run(main())
```

---

## Claude Code Configuration Examples

### Basic .mcp.json Configuration

```json
{
  "mcpServers": {
    "file-server": {
      "command": "node",
      "args": ["./mcp-servers/file-server/dist/index.js"]
    },
    "database-server": {
      "command": "python",
      "args": ["-m", "mcp_servers.database"],
      "env": {
        "DATABASE_URL": "postgresql://localhost/mydb"
      }
    }
  }
}
```

### Multi-Server Configuration with NPX

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "./data"]
    },
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    },
    "postgres": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-postgres"],
      "env": {
        "POSTGRES_URL": "postgresql://user:pass@localhost/db"
      }
    }
  }
}
```

### Development Server Configuration

```json
{
  "mcpServers": {
    "local-dev": {
      "command": "npx",
      "args": ["tsx", "./src/mcp-server/index.ts"],
      "env": {
        "NODE_ENV": "development",
        "DEBUG": "mcp:*"
      }
    }
  }
}
```

---

## Testing MCP Servers

### Manual Testing via Command Line

Test stdio server by piping JSON-RPC messages:

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{}}}' | node ./dist/server.js
```

### Automated Testing with Jest

```typescript
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { InMemoryTransport } from "@modelcontextprotocol/sdk/inMemory.js";

describe("MCP Server", () => {
  let server: Server;
  let transport: InMemoryTransport;

  beforeEach(async () => {
    server = createMyServer();
    transport = new InMemoryTransport();
    await server.connect(transport);
  });

  it("should list tools", async () => {
    const response = await transport.send({
      jsonrpc: "2.0",
      id: 1,
      method: "tools/list",
      params: {},
    });

    expect(response.result.tools).toHaveLength(2);
    expect(response.result.tools[0].name).toBe("my_tool");
  });

  it("should call tool successfully", async () => {
    const response = await transport.send({
      jsonrpc: "2.0",
      id: 2,
      method: "tools/call",
      params: {
        name: "my_tool",
        arguments: { input: "test" },
      },
    });

    expect(response.result.content[0].text).toContain("result");
  });
});
```

---

## Error Handling Patterns

### TypeScript Error Handling

```typescript
server.setRequestHandler(CallToolRequestSchema, async (request) => {
  try {
    const result = await performOperation(request.params);
    return { content: [{ type: "text", text: result }] };
  } catch (error) {
    if (error instanceof ValidationError) {
      return {
        content: [
          { type: "text", text: `Validation failed: ${error.message}` },
        ],
        isError: true,
      };
    }

    // Re-throw for protocol-level error handling
    throw new McpError(
      ErrorCode.InternalError,
      `Operation failed: ${error.message}`,
    );
  }
});
```

### Python Error Handling

```python
@server.call_tool()
async def call_tool(name: str, arguments: dict) -> list[TextContent]:
    try:
        result = await perform_operation(name, arguments)
        return [TextContent(type="text", text=result)]
    except ValidationError as e:
        return [TextContent(type="text", text=f"Validation failed: {e}")]
    except OperationError as e:
        raise McpError(ErrorCode.INTERNAL_ERROR, str(e))
```
