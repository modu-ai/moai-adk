# MCP Client Integration

Guide for configuring MCP servers in Claude Code and building custom MCP clients.

---

## Claude Code Configuration

### Configuration Files

Claude Code reads MCP configuration from two locations:

Project-level configuration at .mcp.json in the project root. This configuration applies only to the current project and is typically committed to version control.

User-level configuration at ~/.claude/.mcp.json for global servers. This configuration applies to all projects and is user-specific.

Configuration priority: Project-level settings override user-level settings for servers with the same name.

### Configuration Schema

The configuration file uses JSON format with an mcpServers object containing named server entries:

Each server entry requires a command field specifying the executable path.

Optional fields include args as an array of command-line arguments, env as an object of environment variables, and disabled as a boolean to temporarily disable the server.

### Environment Variable Expansion

Environment variables in configuration support expansion:

Use ${VAR_NAME} syntax to reference environment variables from the shell environment.

This enables secure credential management without hardcoding secrets in configuration files.

Example: Set env.API_KEY to "${MY_API_KEY}" to use the MY_API_KEY environment variable.

---

## Server Configuration Patterns

### NPX-Based Servers

For TypeScript/Node.js servers distributed via npm:

Set command to "npx" and args to ["-y", "@org/package-name", ...additional-args].

The -y flag auto-confirms npx prompts for package installation.

Additional arguments after the package name are passed to the server.

### Python Module Servers

For Python servers distributed as packages:

Set command to "python" and args to ["-m", "module_name", ...additional-args].

Ensure the Python environment has the required package installed.

Use virtual environments for isolated dependencies.

### Local Development Servers

For servers in development:

Use absolute paths or paths relative to the project root.

TypeScript: Set command to "npx" and args to ["tsx", "./src/server/index.ts"].

JavaScript: Set command to "node" and args to ["./dist/server.js"].

Python: Set command to "python" and args to ["./src/server.py"].

### Docker-Based Servers

For containerized servers:

Set command to "docker" and args to ["run", "-i", "--rm", "image-name"].

The -i flag enables interactive mode for stdio communication.

The --rm flag removes the container after exit.

Mount volumes for persistent data access if needed.

---

## Multi-Server Setup

### Server Isolation

Each MCP server operates independently with its own namespace. Tool and resource names are automatically prefixed with the server name in Claude Code.

Servers cannot directly communicate with each other. Orchestration happens at the client level.

### Naming Conventions

Use descriptive server names that indicate functionality: "github" for GitHub integration, "database" for database access, "search" for search functionality.

Avoid generic names that may conflict with future built-in features.

### Resource Management

Consider memory and CPU usage when running multiple servers. Each server process consumes system resources.

Disable unused servers in configuration to reduce resource consumption.

Use disabled: true to temporarily disable servers without removing configuration.

---

## Capability Negotiation

### Initialization Handshake

When Claude Code starts, it initializes each configured server:

Client sends initialize request with protocolVersion and client capabilities.

Server responds with its capabilities including resources, tools, prompts, and logging.

Client sends initialized notification to confirm successful initialization.

### Version Compatibility

Servers should check the client's protocolVersion for compatibility.

Current specification version is 2025-11-25. Servers may support older versions for backward compatibility.

Reject incompatible versions with appropriate error response.

### Capability Discovery

After initialization, Claude Code queries available resources, tools, and prompts.

Dynamic servers may add or remove capabilities at runtime using notification messages.

Clients re-query capabilities when receiving listChanged notifications.

---

## Connection Lifecycle

### Startup

Claude Code launches configured servers as child processes.

Servers initialize and wait for the initialize request.

Successful initialization enables tool and resource access.

Failed initialization disables the server with an error message.

### Active Session

Normal operation with request/response exchanges.

Servers may send notifications for logging and capability changes.

Claude Code manages message routing to appropriate servers.

### Shutdown

Claude Code sends shutdown request before terminating.

Servers complete pending operations and release resources.

Transport connection closes after shutdown acknowledgment.

---

## Debugging Integration

### Enable Verbose Logging

Configure servers to output debug information to stderr.

Check Claude Code logs for MCP communication details.

Use environment variables to control log levels.

### Common Issues

Connection timeout: Server takes too long to initialize. Check for blocking operations during startup.

Command not found: Verify command path is correct. Check PATH environment variable.

Permission denied: Ensure executable permissions are set. Check file access permissions.

Parse error: Invalid JSON in configuration file. Validate JSON syntax.

### Testing Configuration

Test server command manually in terminal before adding to configuration.

Verify environment variables are set correctly.

Check that all required dependencies are installed.

---

## Building Custom Clients

### Client Implementation

Use the MCP SDK to build custom clients:

Create a Client instance with client name and version.

Implement transport for server communication.

Handle capability negotiation during initialization.

### Transport Selection

StdioClientTransport for local process communication.

SSEClientTransport for HTTP with Server-Sent Events.

Custom transports can implement the Transport interface.

### Resource Access

Use client.listResources to discover available resources.

Use client.readResource with URI to fetch resource content.

Subscribe to resource changes using client.subscribeResource.

### Tool Invocation

Use client.listTools to discover available tools.

Use client.callTool with name and arguments to execute tools.

Handle tool results and errors appropriately.

### Prompt Retrieval

Use client.listPrompts to discover available prompts.

Use client.getPrompt with name and arguments to get prompt messages.

Use prompt messages in AI model interactions.
