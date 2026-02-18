# MCP OAuth Setup Guide

For MCP servers that require OAuth authentication (e.g., Slack, GitHub, Sentry).

## Configuration

Edit `.mcp.json` in your project root:

```json
{
  "mcpServers": {
    "slack": {
      "type": "http",
      "url": "https://mcp.slack.com/mcp",
      "oauth": {
        "clientId": "your-client-id",
        "callbackPort": 8080
      }
    }
  }
}
```

## Adding OAuth Credentials

```bash
claude mcp add --transport http \
  --client-id YOUR_CLIENT_ID \
  --client-secret \
  --callback-port 8080 \
  slack https://mcp.slack.com/mcp
```

The `--client-secret` flag prompts for secure input. Secrets are stored in the system keychain, not in config files.

## Common OAuth MCP Servers

| Server | URL | Notes |
|--------|-----|-------|
| Slack | https://mcp.slack.com/mcp | Requires Slack app registration |
| GitHub | https://api.githubcopilot.com/mcp | Requires GitHub OAuth app |
| Sentry | https://mcp.sentry.io/mcp | Requires Sentry integration |

## Troubleshooting

- **"Dynamic Client Registration not supported"**: Add `--client-id` and `--client-secret` flags
- **Callback port conflict**: Change `callbackPort` to an unused port (e.g., 8081, 8082)
- **Secret not found**: Re-run `claude mcp add` with `--client-secret` to re-enter credentials
