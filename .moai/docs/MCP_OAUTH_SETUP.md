# MCP OAuth Setup Guide

Guide for configuring OAuth authentication with MCP servers in Claude Code.

## Overview

Claude Code v2.1.30+ supports OAuth authentication for MCP servers through the `.mcp.json` configuration file. OAuth credentials are securely stored in the system keychain.

## Configuration Format

Add OAuth configuration to your `.mcp.json`:

```json
{
  "mcpServers": {
    "server-name": {
      "command": "path-to-server",
      "args": ["--arg1", "value1"],
      "oauth": {
        "clientId": "your-client-id",
        "callbackPort": 8080
      }
    }
  }
}
```

## OAuth Fields

| Field | Required | Description |
|-------|----------|-------------|
| `clientId` | Yes | OAuth client ID from the service provider |
| `callbackPort` | No | Local port for OAuth callback (default: auto-assigned) |

## Client Secret Storage

Client secrets are **never** stored in `.mcp.json`. Instead:

1. First run: Claude Code prompts for the client secret
2. Secret is stored securely in the system keychain
3. Subsequent runs: Secret is retrieved automatically

**Keychain locations by platform:**
- macOS: Keychain Access
- Linux: Secret Service API (GNOME Keyring, KWallet)
- Windows: Windows Credential Manager

## Example Configurations

### Slack MCP Server

```json
{
  "mcpServers": {
    "slack": {
      "command": "mcp-slack-server",
      "args": [],
      "oauth": {
        "clientId": "1234567890.1234567890123",
        "callbackPort": 3000
      }
    }
  }
}
```

**Setup Steps:**
1. Create a Slack App at api.slack.com/apps
2. Add OAuth redirect URL: `http://localhost:3000/callback`
3. Request scopes: `chat:write`, `channels:read`, `users:read`
4. Copy the Client ID to `.mcp.json`
5. Run Claude Code - it will prompt for Client Secret

### GitHub MCP Server

```json
{
  "mcpServers": {
    "github": {
      "command": "mcp-github-server",
      "args": [],
      "oauth": {
        "clientId": "Iv1.a1b2c3d4e5f6g7h8",
        "callbackPort": 8080
      }
    }
  }
}
```

**Setup Steps:**
1. Create a GitHub OAuth App at github.com/settings/developers
2. Set Authorization callback URL: `http://localhost:8080/callback`
3. Request scopes: `repo`, `read:org`, `write:org`
4. Copy the Client ID to `.mcp.json`
5. Generate and save Client Secret when prompted

### Sentry MCP Server

```json
{
  "mcpServers": {
    "sentry": {
      "command": "mcp-sentry-server",
      "args": [],
      "oauth": {
        "clientId": "your-sentry-client-id",
        "callbackPort": 9000
      }
    }
  }
}
```

**Setup Steps:**
1. Create an integration in Sentry Settings
2. Add Redirect URL: `http://localhost:9000/callback`
3. Request scopes: `project:read`, `event:read`, `issue:write`
4. Copy the Client ID to `.mcp.json`

## Security Best Practices

1. **Never commit `.mcp.json` with real credentials** - Use `.gitignore`
2. **Use environment variables for Client IDs** in shared configs:
   ```json
   {
     "clientId": "${SLACK_CLIENT_ID}"
   }
   ```
3. **Rotate secrets periodically** - Remove from keychain and re-authenticate
4. **Use minimal scopes** - Only request permissions your MCP server needs

## Troubleshooting

### OAuth callback fails

- Ensure the callback port is not blocked by firewall
- Check that the redirect URL matches exactly in your OAuth app settings
- Verify no other process is using the callback port

### Client secret not found

1. Clear stored credentials:
   - macOS: Open Keychain Access, search for "Claude Code"
   - Linux: Use secret-tool or your keyring manager
   - Windows: Credential Manager > Windows Credentials
2. Restart Claude Code - it will prompt for the secret again

### Token expired

Most OAuth tokens expire after a period. Claude Code will automatically refresh tokens when possible. If refresh fails:
1. Clear the stored secret
2. Re-authenticate when prompted

## MoAI-ADK Integration

MoAI-ADK templates include `.mcp.json` handling during `moai init`:

- Template `.mcp.json` is copied if it exists
- Environment variable placeholders are preserved
- OAuth configurations are validated for required fields

For MoAI-specific MCP server configurations, see `.moai/config/mcp-servers.yaml`.

---

Version: 1.0.0
Last Updated: 2026-02-20
Compatibility: Claude Code v2.1.30+
