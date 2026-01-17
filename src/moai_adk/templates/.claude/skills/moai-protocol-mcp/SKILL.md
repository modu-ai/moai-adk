---
name: "moai-protocol-mcp"
description: "Model Context Protocol specialist covering MCP server development, client integration, resources, tools, prompts, and transport configuration. Use when building MCP servers, configuring Claude Code MCP integration, or implementing AI application connectors."
version: 1.0.0
category: "protocol"
modularized: true
user-invocable: false
tags:
  [
    "mcp",
    "model-context-protocol",
    "server",
    "client",
    "resources",
    "tools",
    "prompts",
  ]
context7-libraries:
  [
    "/modelcontextprotocol/specification",
    "/websites/modelcontextprotocol_io_specification_2025-11-25",
  ]
related-skills:
  ["moai-foundation-claude", "moai-lang-typescript", "moai-lang-python"]
updated: 2026-01-17
status: "active"
allowed-tools:
  - Read
  - Grep
  - Glob
  - Bash
  - mcp__context7__resolve-library-id
progressive_disclosure:
  enabled: true
  level1_tokens: ~100
  level2_tokens: ~5000
triggers:
  keywords:
    - "mcp"
    - "model context protocol"
    - "mcp server"
    - "mcp client"
    - "resources"
    - "mcp tools"
    - "prompts"
    - "stdio transport"
    - "sse transport"
    - "mcp configuration"
  agents:
    - "manager-claude-code"
    - "expert-backend"
    - "builder-plugin"
  phases:
    - "plan"
    - "run"
---

# MoAI Protocol MCP: Model Context Protocol Specialist

## Quick Reference

Model Context Protocol is an open-source standard for connecting AI applications to external systems, functioning as a universal connector for AI applications similar to how USB-C standardizes device connections.

### What is MCP?

MCP provides a standardized way to connect AI systems to data sources such as local files and databases, tools such as search engines and calculators, and workflows such as specialized prompts and templates. The protocol enables composable, secure, and extensible AI integrations.

### Core Architecture

MCP defines two participant types that communicate through a standardized protocol:

MCP Servers expose data and tools, responding to client requests. They provide resources for data access, tools for AI-callable functions, and prompts for reusable instruction templates.

MCP Clients (Hosts) are AI applications that initiate requests and consume resources. Claude Code functions as an MCP client, connecting to servers for enhanced capabilities.

### When to Use MCP

Use MCP when connecting Claude Code to external data sources, building custom tool integrations for AI workflows, creating reusable prompt templates across projects, exposing internal APIs to AI applications, or implementing search, database, or file system access for AI.

### Protocol Version

Current specification version is 2025-11-25. All implementations should target this version for compatibility.

### Context7 Documentation Access

For latest MCP documentation, use the Context7 MCP tools:

Step 1 - Resolve library ID: Use mcp**context7**resolve-library-id with query "model context protocol" to get the Context7-compatible library ID.

Step 2 - Fetch documentation: Use the resolved library ID with topic specification such as "mcp server resources tools" and appropriate token allocation.

---

## Implementation Guide

### Server Capabilities

MCP servers expose three primary features to clients:

Resources provide data for AI consumption. They can be static such as file contents, dynamic such as database queries, or computed such as search results. Resources use URI-based identification and support text and binary content types.

Tools are functions that AI can invoke with parameters. Each tool has a name, description, input schema, and handler function. Tools enable AI to perform actions such as calculations, searches, or API calls.

Prompts are reusable instruction templates with parameter support. They enable standardized interaction patterns that can be shared across AI applications.

Additional server features include logging for sending diagnostic messages to clients and completion for providing intelligent auto-completion suggestions.

### Client Capabilities

MCP clients provide three features to servers:

Roots define accessible file system locations. They establish security boundaries for what files and directories a server can access.

Sampling allows servers to request AI model inference through the client. This enables servers to leverage the client's AI capabilities for processing.

Elicitation enables servers to request information from clients interactively, supporting user input collection through the AI application.

### Transport Mechanisms

MCP supports multiple transport protocols:

Stdio transport uses standard input and output for local process communication. This is the most common transport for local MCP servers and is used by Claude Code for local integrations.

HTTP with SSE transport uses Server-Sent Events over HTTP for remote communication. This enables web-based MCP servers accessible over networks.

WebSocket transport provides bidirectional communication for real-time applications. Experimental support is available in the specification.

### Claude Code MCP Configuration

Configure MCP servers in Claude Code using the .mcp.json file in your project root or the global configuration at ~/.claude/.mcp.json.

Basic configuration structure includes an mcpServers object containing named server definitions. Each server definition includes command for the executable path and args array for command-line arguments.

For example, to configure a local TypeScript MCP server, create an entry with the server name as the key, set command to "npx" and args to ["-y", "@your-org/your-mcp-server"].

For Python-based servers, set command to "python" and args to ["-m", "your_mcp_server"].

Environment variables can be passed using the env object within the server configuration. Common variables include API_KEY for authentication and DATA_PATH for resource locations.

---

## Server Development

### TypeScript Server Template

To create an MCP server in TypeScript, use the @modelcontextprotocol/sdk package.

Import Server and StdioServerTransport from the SDK. Create a server instance with name, version, and capabilities object specifying resources, tools, and prompts as empty objects.

Implement resource handlers by calling server.setRequestHandler with ListResourcesRequestSchema to return available resources. Each resource needs uri, name, and mimeType.

Implement tool handlers by setting handlers for ListToolsRequestSchema and CallToolRequestSchema. Tools require name, description, and inputSchema using JSON Schema format.

Connect the server to transport by creating a StdioServerTransport instance and calling server.connect with the transport.

### Python Server Template

For Python servers, use the mcp package available via pip install mcp.

Import Server and stdio_server from mcp.server. Import types for resources and tools from mcp.types.

Create a Server instance with the server name. Define resource handlers using the @server.list_resources decorator returning a list of Resource objects. Implement read handlers with @server.read_resource decorator.

Define tool handlers using @server.list_tools returning Tool objects with name, description, and inputSchema. Implement call handlers with @server.call_tool decorator.

Run the server using asyncio with stdio_server context manager, reading from the server's incoming messages stream.

### Best Practices

Security: Validate all input parameters, implement rate limiting, use least-privilege access patterns, and never expose sensitive credentials in tool responses.

Performance: Cache frequently accessed resources, implement pagination for large datasets, use streaming for large responses, and optimize database queries.

Error Handling: Return descriptive error messages, use appropriate error codes, implement retry logic for transient failures, and log errors for debugging.

---

## Advanced Patterns

### Resource Implementation Patterns

Static resources serve fixed content such as configuration files or documentation. Dynamic resources query external systems such as databases or APIs. Computed resources generate content on demand such as search results or aggregations.

Resource templates enable parameterized URIs for flexible data access. Use URI templates following RFC 6570 for pattern-based resource identification.

### Tool Design Patterns

Atomic tools perform single operations with clear inputs and outputs. Composite tools orchestrate multiple operations for complex workflows. Streaming tools return incremental results for long-running operations.

Tool input schemas should use JSON Schema with clear property descriptions. Mark required parameters explicitly and provide sensible defaults for optional parameters.

### Prompt Engineering for MCP

Design prompts with clear parameter placeholders using double curly brace syntax. Include context about expected inputs and outputs. Provide examples of proper usage within prompt descriptions.

Parameterized prompts enable customization while maintaining consistent structure. Use argument definitions with name, description, and required properties.

### Multi-Server Orchestration

Claude Code can connect to multiple MCP servers simultaneously. Each server operates independently with its own capabilities namespace.

Configure server priorities and failover behavior for resilient integrations. Use naming conventions to avoid tool and resource conflicts across servers.

---

## Module Index

This skill uses progressive disclosure with specialized modules for detailed implementation patterns.

### Core Modules

server-development covers complete MCP server implementation in TypeScript and Python. Topics include SDK setup, resource handlers, tool implementations, and prompt definitions.

client-integration covers MCP client development and Claude Code configuration. Topics include transport configuration, capability negotiation, and connection management.

transport-protocols covers stdio, HTTP with SSE, and WebSocket transport implementations. Topics include transport selection, security considerations, and performance optimization.

security-authorization covers MCP security patterns and authorization flows. Topics include authentication, access control, and secure credential handling.

---

## Works Well With

Related Skills:

moai-foundation-claude provides Claude Code authoring patterns for MCP integration in skills, agents, and plugins.

moai-lang-typescript provides TypeScript patterns for building MCP servers with the official SDK.

moai-lang-python provides Python patterns for building MCP servers using the Python SDK.

moai-domain-backend provides backend architecture patterns for integrating MCP with APIs and databases.

Agents:

manager-claude-code uses this skill for MCP configuration and integration guidance.

expert-backend uses this skill for server-side MCP implementations.

builder-plugin uses this skill for plugin development with MCP server integration.

Commands:

Project initialization commands configure MCP server connections. Plugin commands can install MCP-enabled plugins.

---

## Module References

For detailed implementation patterns, see the modules directory:

modules/server-development.md covers complete server implementation with SDK usage, resource handlers, tool implementations, and testing strategies.

modules/client-integration.md covers Claude Code configuration, multi-server setup, and client-side patterns.

modules/transport-protocols.md covers transport selection, configuration, and optimization for different deployment scenarios.

modules/security-authorization.md covers authentication, authorization, and security best practices.

For API reference summary, see reference.md. For working code templates, see examples.md.

---

## Resources

Official Documentation:

- Specification: https://modelcontextprotocol.io/specification/2025-11-25
- Build Servers: https://modelcontextprotocol.io/docs/develop/build-server
- Build Clients: https://modelcontextprotocol.io/docs/develop/build-client
- GitHub: https://github.com/modelcontextprotocol

SDK Packages:

- TypeScript SDK: @modelcontextprotocol/sdk on npm
- Python SDK: mcp on PyPI

---

Status: Production Ready
Generated with: MoAI-ADK Skill Factory v2.0
Last Updated: 2026-01-17
Version: 1.0.0
Coverage: MCP Specification 2025-11-25, Server Development, Client Integration, Transports
