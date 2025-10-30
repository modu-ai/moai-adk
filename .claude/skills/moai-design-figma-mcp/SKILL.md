---
name: moai-design-figma-mcp
description: Integrating Figma with MCP servers for real-time design collaboration. Connect Figma files, manage webhooks, and sync design changes. Use when setting up Figma MCP integration, automating design workflows, or building design-to-code pipelines.
allowed-tools: Read, Bash, WebFetch
version: 1.0.0
tier: design
created: 2025-10-31
---

# Design: Figma MCP Integration

## What it does

Establishes Model Context Protocol (MCP) integration with Figma to enable real-time design file access, webhook management, and automated synchronization between design and development workflows.

## When to use

- Setting up Figma MCP server connections
- Configuring webhooks for design change notifications
- Building automated design-to-code pipelines
- Syncing Figma files with local development environments
- Monitoring design system changes in real-time

## Key Patterns

### 1. MCP Server Configuration

**Pattern**: Configure Figma MCP server in Claude Code settings

\`\`\`json
{
  "mcpServers": {
    "figma": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-figma"],
      "env": {
        "FIGMA_PERSONAL_ACCESS_TOKEN": "\${FIGMA_TOKEN}"
      }
    }
  }
}
\`\`\`

### 2. Webhook Setup

**Pattern**: Register webhooks for file update notifications

\`\`\`typescript
// Register webhook for file updates
const webhookUrl = 'https://your-server.com/figma-webhook';
const fileKey = 'abc123def456';

await fetch(\`https://api.figma.com/v2/webhooks\`, {
  method: 'POST',
  headers: {
    'X-Figma-Token': process.env.FIGMA_TOKEN,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    event_type: 'FILE_UPDATE',
    team_id: 'your-team-id',
    endpoint: webhookUrl,
    passcode: 'webhook-secret'
  })
});
\`\`\`

### 3. Real-Time Sync

**Pattern**: Listen for webhook events and trigger updates

\`\`\`typescript
// Webhook handler
app.post('/figma-webhook', async (req, res) => {
  const { event_type, file_key, timestamp } = req.body;
  
  if (event_type === 'FILE_UPDATE') {
    // Trigger design token extraction
    await extractDesignTokens(file_key);
    
    // Regenerate components
    await generateComponents(file_key);
    
    // Commit changes
    await commitToGit(\`Updated from Figma: \${timestamp}\`);
  }
  
  res.status(200).send('OK');
});
\`\`\`

## Best Practices

- Store Figma tokens securely in environment variables (never commit to git)
- Implement webhook signature verification for security
- Use rate limiting to avoid API throttling
- Cache Figma file data to reduce API calls
- Implement retry logic for failed requests
- Log all webhook events for debugging
- Version control generated code separately from manual code

## Resources

- Figma MCP Server: https://github.com/modelcontextprotocol/servers/tree/main/src/figma
- Figma Webhooks Documentation: https://www.figma.com/developers/api#webhooks
- MCP Specification: https://modelcontextprotocol.io/

## Examples

**Example: Initialize Figma MCP Connection**

\`\`\`bash
# Install Figma MCP server
npm install -g @modelcontextprotocol/server-figma

# Configure environment
export FIGMA_PERSONAL_ACCESS_TOKEN="figd_xxxx"

# Test connection
curl -H "X-Figma-Token: \$FIGMA_PERSONAL_ACCESS_TOKEN" \
  https://api.figma.com/v1/me
\`\`\`

**Example: Query File via MCP**

\`\`\`typescript
// Using MCP to fetch Figma file
const mcp = await connectMCP('figma');
const file = await mcp.call('getFile', { fileKey: 'abc123' });
console.log(file.name, file.lastModified);
\`\`\`

## Changelog
- 2025-10-31: v1.0.0 - Initial release with MCP integration, webhook support

## Works well with
- \`moai-design-figma-to-code\` (Code generation from Figma)
- \`moai-cc-mcp-plugins\` (MCP plugin management)
