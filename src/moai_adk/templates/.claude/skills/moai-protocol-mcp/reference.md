# MCP Reference Documentation

## Official Resources

### Specification

- Main Specification: https://modelcontextprotocol.io/specification/2025-11-25
- Architecture Overview: https://modelcontextprotocol.io/specification/2025-11-25/architecture
- Base Protocol: https://modelcontextprotocol.io/specification/2025-11-25/protocol

### Server Development

- Build Server Guide: https://modelcontextprotocol.io/docs/develop/build-server
- TypeScript SDK: https://github.com/modelcontextprotocol/typescript-sdk
- Python SDK: https://github.com/modelcontextprotocol/python-sdk

### Client Development

- Build Client Guide: https://modelcontextprotocol.io/docs/develop/build-client
- Claude Desktop Integration: https://modelcontextprotocol.io/docs/quickstart/claude-desktop

### GitHub Repositories

- Main Specification: https://github.com/modelcontextprotocol/specification
- TypeScript SDK: https://github.com/modelcontextprotocol/typescript-sdk
- Python SDK: https://github.com/modelcontextprotocol/python-sdk
- Example Servers: https://github.com/modelcontextprotocol/servers

---

## Protocol Schema Reference

### Message Types

All MCP messages follow JSON-RPC 2.0 format with these base types:

Request messages contain jsonrpc set to "2.0", id as string or number, method as string, and optional params object.

Response messages contain jsonrpc set to "2.0", id matching the request, and either result object or error object with code, message, and optional data.

Notification messages contain jsonrpc set to "2.0", method as string, and optional params object. Notifications have no id and expect no response.

### Capability Negotiation

During initialization, clients and servers exchange capability information:

Client capabilities include roots for file system access, sampling for AI model access, and experimental features.

Server capabilities include resources with subscribe and listChanged flags, tools with listChanged flag, prompts with listChanged flag, and logging support.

### Resource URIs

Resources use URI format with these common schemes:

File scheme (file://) for local file system resources.

Custom schemes for application-specific resources such as database://, api://, or search://.

URI templates follow RFC 6570 for parameterized resource access.

### Tool Input Schema

Tools use JSON Schema for input validation:

Required fields include type set to "object", properties object defining each parameter, and required array listing mandatory parameters.

Each property needs type, description, and optionally enum for constrained values, default for optional parameters, or format for string validation.

### Error Codes

Standard JSON-RPC error codes:

-32700 for Parse error indicating invalid JSON.

-32600 for Invalid Request indicating malformed request structure.

-32601 for Method not found indicating unknown method.

-32602 for Invalid params indicating parameter validation failure.

-32603 for Internal error indicating server-side error.

MCP-specific error codes start at -32000 and are defined per implementation.

---

## Transport Configuration

### Stdio Transport

Configuration for local process communication:

Command specifies the executable path. Args provides command-line arguments array. Env sets environment variables object.

Security: Stdio transport runs with the permissions of the parent process. Ensure proper sandboxing when running untrusted servers.

### HTTP with SSE Transport

Configuration for remote server communication:

URL specifies the server endpoint. Headers provides authentication headers. Timeout sets connection timeout in milliseconds.

Security: Always use HTTPS for remote servers. Validate SSL certificates. Implement authentication using Authorization headers.

### Connection Lifecycle

Initialization: Client sends initialize request with protocol version and capabilities. Server responds with its capabilities.

Operation: Normal message exchange using requests, responses, and notifications.

Shutdown: Client sends shutdown request. Server completes pending operations. Transport connection closes.

---

## Claude Code Integration

### Configuration File Locations

Project-level configuration at .mcp.json in project root.

User-level configuration at ~/.claude/.mcp.json for global servers.

Priority: Project configuration takes precedence over user configuration.

### Server Configuration Schema

The mcpServers object contains named server entries:

Each entry requires command as the executable string.

Optional fields include args as string array, env as object of environment variables, and disabled as boolean.

### Example Configurations

TypeScript server via npx: Set command to "npx" and args to ["-y", "@org/server-name"].

Python server: Set command to "python" and args to ["-m", "server_module"].

Local development server: Set command to "node" and args to ["./dist/server.js"].

Docker container: Set command to "docker" and args to ["run", "-i", "server-image"].

### Debugging MCP Servers

Enable verbose logging in Claude Code settings.

Check server process output for error messages.

Validate configuration JSON syntax.

Test server independently using command-line invocation.

---

## SDK API Reference

### TypeScript SDK

Server class provides the main server implementation.

StdioServerTransport handles stdio communication.

setRequestHandler registers handlers for specific request schemas.

connect binds server to transport and starts message processing.

### Python SDK

Server class provides the main server implementation.

stdio_server provides async context manager for stdio transport.

Decorators @server.list_resources, @server.read_resource, @server.list_tools, @server.call_tool register handlers.

run_server starts the server with specified transport.

---

## Version History

2025-11-25 (Latest): Current stable specification with full feature set.

2024-11-05: Initial public specification release.

Always check modelcontextprotocol.io for the latest specification updates.
