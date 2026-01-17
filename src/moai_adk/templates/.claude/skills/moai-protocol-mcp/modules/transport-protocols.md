# MCP Transport Protocols

Detailed guide for MCP transport mechanisms including stdio, HTTP with SSE, and WebSocket.

---

## Transport Overview

### Purpose of Transports

Transports handle the low-level communication between MCP clients and servers. They are responsible for message serialization, delivery, and connection management.

The transport layer abstracts communication details, allowing the same server logic to work with different transport mechanisms.

### Message Format

All transports use JSON-RPC 2.0 message format:

Request messages contain jsonrpc, id, method, and optional params.

Response messages contain jsonrpc, id, and either result or error.

Notification messages contain jsonrpc, method, and optional params without id.

Messages are serialized as JSON strings with UTF-8 encoding.

---

## Stdio Transport

### Overview

Stdio transport uses standard input and output streams for communication. It is the most common transport for local MCP servers integrated with Claude Code.

### Communication Flow

Client writes JSON-RPC messages to the server's stdin.

Server reads messages from stdin, processes them, and writes responses to stdout.

Error messages and debug output go to stderr to avoid interfering with protocol communication.

### Message Framing

Messages are delimited by newlines. Each JSON-RPC message occupies a single line.

Servers must handle partial reads and buffer input until a complete message is received.

Large messages may require chunked reading and buffering.

### Implementation

TypeScript uses StdioServerTransport from the SDK which handles stdin/stdout streams automatically.

Python uses the stdio_server context manager which provides async read and write streams.

Both implementations handle message parsing, serialization, and buffering.

### Process Management

Claude Code launches server processes and manages their lifecycle.

Servers should handle SIGTERM and SIGINT for graceful shutdown.

Exit codes indicate success (0) or failure (non-zero) status.

### Advantages

Simple to implement and debug.

No network configuration required.

Secure by default since communication is process-local.

Low latency for local operations.

### Limitations

Only works for local processes.

No built-in authentication.

Single client per server instance.

---

## HTTP with SSE Transport

### Overview

HTTP with Server-Sent Events (SSE) enables MCP communication over HTTP. This transport supports remote servers accessible via network.

### Architecture

Client sends requests via HTTP POST to the server endpoint.

Server sends responses and notifications via SSE stream.

Connection remains open for continuous communication.

### Request Handling

Client sends JSON-RPC request as POST body with Content-Type: application/json.

Server processes the request and sends response via SSE.

Multiple requests can be in flight simultaneously.

### SSE Message Format

SSE messages use the standard event stream format:

Each message starts with "data: " followed by the JSON-RPC message.

Messages end with double newline to indicate message boundary.

Event types can distinguish between responses and notifications.

### Connection Management

Clients maintain persistent SSE connection for receiving messages.

Implement reconnection logic for connection failures.

Use heartbeat messages to detect connection health.

### Authentication

Use Authorization header for authentication.

Support Bearer tokens, API keys, or other auth mechanisms.

Validate authentication on every request.

### CORS Configuration

Configure CORS headers for browser-based clients.

Set Access-Control-Allow-Origin appropriately.

Allow required headers and methods.

### Implementation Considerations

Handle connection timeouts and reconnection.

Implement request timeout for long-running operations.

Buffer messages during reconnection.

Use HTTPS for secure communication.

### Advantages

Works across network boundaries.

Supports multiple clients.

Built-in authentication options.

Compatible with standard web infrastructure.

### Limitations

Higher latency than stdio.

Requires network configuration.

More complex to implement and debug.

---

## WebSocket Transport

### Overview

WebSocket transport provides bidirectional communication over a single connection. This transport is experimental in the MCP specification.

### Communication Flow

Client establishes WebSocket connection to server.

Both client and server can send messages at any time.

Messages use JSON-RPC format over WebSocket frames.

### Message Handling

Each WebSocket message contains a complete JSON-RPC message.

No additional framing needed.

Binary frames not typically used for MCP.

### Connection Lifecycle

Client initiates connection with WebSocket handshake.

Connection remains open for duration of session.

Either party can close the connection.

### Implementation Notes

WebSocket transport is not yet standardized in MCP.

Implementation details may vary between SDKs.

Check SDK documentation for current support status.

### Advantages

True bidirectional communication.

Lower overhead than HTTP polling.

Good for real-time applications.

### Limitations

Experimental status in specification.

More complex connection management.

May have firewall compatibility issues.

---

## Transport Selection Guide

### Local Development

Use stdio transport for local development servers.

Simplest to set up and debug.

No network configuration required.

### Production Local Servers

Stdio transport remains appropriate.

Consider security implications of server access.

Implement proper error handling and logging.

### Remote Servers

Use HTTP with SSE for network-accessible servers.

Implement authentication and authorization.

Use HTTPS for secure communication.

### High-Throughput Applications

Evaluate based on message volume and latency requirements.

Stdio has lowest latency for local.

HTTP with SSE may have connection overhead.

### Browser-Based Clients

HTTP with SSE or WebSocket required.

Configure CORS appropriately.

Handle connection management in client.

---

## Performance Optimization

### Message Batching

Batch multiple operations when possible.

Reduces round-trip overhead.

Consider trade-offs with latency.

### Connection Pooling

Reuse connections for multiple requests.

Avoid connection establishment overhead.

Implement connection health monitoring.

### Compression

Consider message compression for large payloads.

gzip compression supported by HTTP transport.

Evaluate CPU vs bandwidth trade-offs.

### Streaming Responses

Use streaming for large data transfers.

Avoid loading entire response into memory.

Implement backpressure handling.

---

## Error Handling

### Connection Errors

Implement retry logic with exponential backoff.

Handle network failures gracefully.

Provide meaningful error messages.

### Timeout Handling

Set appropriate timeouts for requests.

Handle timeout errors at application level.

Consider long-running operation patterns.

### Protocol Errors

Handle JSON parsing errors.

Validate message structure.

Return appropriate error codes.
