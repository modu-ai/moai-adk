# Connect Claude Code to tools via MCP

Claude Code can connect to hundreds of external tools and data sources through the Model Context Protocol (MCP), an open-source standard for AI-tool integrations.

## Overview

MCP servers give Claude Code access to your tools, databases, and APIs. With MCP servers connected, Claude Code can:

- **Implement features from issue trackers**: "Add the feature described in JIRA issue ENG-4521 and create a PR on GitHub."
- **Analyze monitoring data**: "Check Sentry and Statsig to check the usage of the feature described in ENG-4521."
- **Query databases**: "Find emails of 10 random users who used feature ENG-4521, based on our Postgres database."
- **Integrate designs**: "Update our standard email template based on the new Figma designs that were posted in Slack"
- **Automate workflows**: "Create Gmail drafts inviting these 10 users to a feedback session about the new feature."

## Server Configuration Types

### 1. Local Stdio Servers

```bash
claude mcp add airtable --env AIRTABLE_API_KEY=YOUR_KEY -- npx -y airtable-mcp-server
```

### 2. Remote SSE Servers

Provide real-time streaming connections:

```bash
claude mcp add --transport sse linear https://mcp.linear.app/sse
```

### 3. Remote HTTP Servers

Use standard request/response patterns:

```bash
claude mcp add --transport http notion https://mcp.notion.com/mcp
```

## Token Management (2025)

When MCP tools produce large outputs, Claude Code helps manage the token usage to prevent overwhelming your conversation context:

- **Output warning threshold**: Claude Code displays a warning when any MCP tool output exceeds 10,000 tokens
- **Configurable limit**: You can adjust the maximum allowed MCP output tokens using the `MAX_MCP_OUTPUT_TOKENS` environment variable
- **Default limit**: The default maximum is 25,000 tokens

## Server Management Commands

- **List servers**: `claude mcp list`
- **Get server details**: `claude mcp get github`
- **Remove server**: `claude mcp remove github`
- **Test server**: `claude mcp test github`

## Installation Scopes

MCP servers can be configured at three different scope levels:

1. **Local scope**: Personal, project-specific configurations stored in your project-specific user settings
2. **Project scope**: Team-shared configurations in `.mcp.json`
3. **User scope**: Cross-project configurations available across all projects

## API Integration Features (2025)

Claude's Model Context Protocol (MCP) connector feature enables you to connect to remote MCP servers directly from the Messages API without a separate MCP client. This feature requires the beta header: `"anthropic-beta": "mcp-client-2025-04-04"`

The API integration supports:
- Direct API integration without implementing an MCP client
- Tool calling support through the Messages API
- OAuth authentication for authenticated servers
- Multiple servers in a single request

## Authentication

Many cloud-based MCP servers require OAuth 2.0 authentication. Use `/mcp` command to authenticate through the browser-based OAuth flow.

## Popular MCP Servers

Includes integrations with:

- **Development Tools**: Sentry, Socket, GitHub
- **Project Management**: Asana, Jira, Linear
- **Databases**: PostgreSQL, MySQL, SQLite
- **Design Platforms**: Figma, Sketch
- **Infrastructure**: AWS, Google Cloud, Docker
- **Communication**: Slack, Discord, Gmail
- **Storage**: Google Drive, Dropbox, S3
- **And many more**

## Security Warning

> Use third party MCP servers at your own risk - Anthropic has not verified the correctness or security of all these servers.

## Example Use Cases

### Database Integration

```bash
# Add PostgreSQL MCP server
claude mcp add postgres --env DATABASE_URL=postgresql://... -- npx -y postgres-mcp-server
```

### API Integration

```bash
# Add Stripe MCP server
claude mcp add stripe --env STRIPE_API_KEY=sk_test_... -- npx -y stripe-mcp-server
```

### Monitoring Integration

```bash
# Add Sentry MCP server
claude mcp add --transport http sentry https://mcp.sentry.io/v1
```

## Configuration File (.mcp.json)

Project-level MCP configuration:

```json
{
  "servers": {
    "github": {
      "transport": "stdio",
      "command": "npx",
      "args": ["-y", "github-mcp-server"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    }
  }
}
```

## Environment Variable Expansion

MCP supports environment variable expansion in configuration:

- Use `${VAR_NAME}` syntax
- Variables are resolved at runtime
- Supports default values: `${VAR_NAME:-default_value}`

## Best Practices

1. **Security**: Never commit API keys to version control
2. **Scoping**: Use project-level servers for team tools
3. **Testing**: Test servers in isolation first
4. **Documentation**: Document server requirements for team
5. **Updates**: Keep MCP servers updated for latest features
6. **Token Management**: Monitor MCP output token usage to avoid context overflow
7. **Error Handling**: Implement proper error handling for server connections

## Platform Integration

Claude Code maintains awareness of your entire project structure, can find up-to-date information from the web, and with MCP can pull from external datasources like Google Drive, Figma, and Slack. MCP lets Claude read your design docs in Google Drive, update your tickets in Jira, or use your custom developer tooling.

## SDK Integration

The Claude Code SDK provides all the building blocks you need to build production-ready agents. The Model Context Protocol (MCP) lets you give your agents custom tools and capabilities for seamless integration with external services.

**Note**: MCP capabilities have been significantly enhanced in 2025 with improved token management, API integration, and expanded server ecosystem.
