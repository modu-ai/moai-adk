# MCP Server Development

Comprehensive guide for building MCP servers in TypeScript and Python.

---

## Server Architecture

### Core Components

MCP servers consist of three main components:

Server Instance manages the protocol state, capability negotiation, and message routing. It maintains the connection lifecycle and handles initialization handshakes.

Transport Layer handles message serialization and communication. Stdio transport is most common for local servers, while HTTP with SSE supports remote deployments.

Request Handlers implement the actual functionality for resources, tools, and prompts. Each handler receives validated requests and returns structured responses.

### Capability Declaration

Servers declare capabilities during initialization:

Resources capability indicates the server provides data access. The subscribe flag enables clients to receive change notifications. The listChanged flag indicates the resource list may change dynamically.

Tools capability indicates the server provides callable functions. The listChanged flag indicates tools may be added or removed at runtime.

Prompts capability indicates the server provides instruction templates. The listChanged flag indicates prompts may change.

Logging capability indicates the server can send log messages to the client.

---

## TypeScript Server Development

### Project Setup

Initialize a TypeScript MCP server project:

Create package.json with type set to "module" for ES modules support. Add dependencies for @modelcontextprotocol/sdk and typescript. Configure build scripts for tsc compilation.

Create tsconfig.json with target set to "ES2022", module set to "NodeNext", and moduleResolution set to "NodeNext". Enable strict mode and set outDir to "./dist".

### Server Implementation Pattern

Import the Server class and transport from the SDK. Create the server instance with name, version, and capabilities object.

Register handlers for each capability using setRequestHandler with the appropriate request schema. Each handler receives a request object and returns a response object.

Connect the server to the transport to begin processing messages. Handle shutdown signals to cleanly disconnect.

### Resource Implementation

Resources require two handlers:

ListResourcesRequestSchema handler returns an array of resource metadata including uri, name, description, and mimeType.

ReadResourceRequestSchema handler receives the resource URI and returns content as text or blob with appropriate MIME type.

Resource URIs should follow consistent naming conventions. Use URI schemes to namespace different resource types such as file://, config://, or db://.

### Tool Implementation

Tools require two handlers:

ListToolsRequestSchema handler returns an array of tool definitions including name, description, and inputSchema.

CallToolRequestSchema handler receives tool name and arguments, executes the operation, and returns content array with results.

Input schemas use JSON Schema format. Define all parameters with types and descriptions. Mark required parameters explicitly.

Tool responses include content array with text or image content types. Set isError to true for error responses that should be displayed to the user.

### Prompt Implementation

Prompts require two handlers:

ListPromptsRequestSchema handler returns an array of prompt definitions including name, description, and arguments array.

GetPromptRequestSchema handler receives prompt name and arguments, returning messages array with role and content.

Prompt arguments have name, description, and required properties. Use argument values to customize the returned messages.

---

## Python Server Development

### Project Setup

Create a Python package with pyproject.toml or setup.py. Add mcp as a dependency. The package provides Server class, transport utilities, and type definitions.

Structure the project with a main module containing the server implementation. Use async/await patterns throughout as MCP operations are asynchronous.

### Server Implementation Pattern

Create a Server instance with the server name. Use decorator-based handler registration for cleaner code.

The @server.list_resources decorator registers the resource list handler. The @server.read_resource decorator registers the resource read handler.

Similarly, @server.list_tools and @server.call_tool decorators handle tool operations. @server.list_prompts and @server.get_prompt handle prompt operations.

Run the server using asyncio with the stdio_server context manager providing read and write streams.

### Async Best Practices

Use async/await consistently for all I/O operations. Avoid blocking calls that can delay message processing.

Use asyncio.gather for parallel operations when handling multiple independent tasks. Implement proper timeout handling for external API calls.

Handle cancellation gracefully by checking for CancelledError and cleaning up resources.

---

## Error Handling

### Protocol Errors

Use standard JSON-RPC error codes for protocol-level errors:

-32600 Invalid Request for malformed request structure.
-32601 Method not found for unknown methods.
-32602 Invalid params for parameter validation failures.
-32603 Internal error for unexpected server errors.

Throw McpError with appropriate code and message for protocol errors.

### Application Errors

For tool execution errors, return a response with isError set to true and descriptive error message in content.

Distinguish between user-facing errors that should be displayed and internal errors that require protocol-level handling.

Log errors appropriately for debugging while returning safe error messages to clients.

---

## Testing Strategies

### Unit Testing

Test handlers in isolation by mocking dependencies. Verify correct response structure for various inputs. Test error handling paths.

Use the SDK's InMemoryTransport for integration testing without process communication.

### Integration Testing

Test full server lifecycle including initialization and shutdown. Verify capability negotiation produces expected results.

Test with actual Claude Code configuration to ensure compatibility.

### Manual Testing

Use command-line JSON-RPC messages piped to the server for quick testing. Validate responses match expected format.

Enable debug logging to trace message flow and identify issues.

---

## Deployment Considerations

### Local Servers

Package as npm module or Python package for easy installation. Use npx for TypeScript servers or python -m for Python servers.

Document required environment variables and configuration options.

### Remote Servers

Implement HTTP with SSE transport for network-accessible servers. Add authentication using Authorization headers.

Configure CORS headers for browser-based clients. Implement rate limiting to prevent abuse.

### Security

Validate all inputs before processing. Sanitize file paths to prevent directory traversal.

Use least-privilege access patterns. Avoid exposing sensitive data in error messages.

Implement audit logging for sensitive operations.
