---
source: https://opencode.ai/docs/share/
fetched: 2026-01-08
title: Share Feature
---

# OpenCode Share Feature Documentation

## Overview

OpenCode enables users to create public links for conversations, facilitating collaboration.

**Important:** "Shared conversations are publicly accessible to anyone with the link."

## How It Works

The sharing mechanism operates in three steps:

1. Generates a unique public URL for the session
2. Synchronizes conversation history to OpenCode servers
3. Makes content accessible via the format `opncd.ai/s/<share-id>`

## Sharing Modes

### Manual (Default)

Users must explicitly invoke `/share` to generate a shareable URL copied to clipboard.

```json
{
  "$schema": "https://opencode.ai/config.json",
  "share": "manual"
}
```

### Auto-Share

Enabling automatic sharing shares all new conversations automatically:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "share": "auto"
}
```

### Disabled

Completely prevents sharing functionality:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "share": "disabled"
}
```

## Un-sharing

The `/unshare` command removes public access and deletes associated conversation data.

```bash
/unshare
```

## Privacy Considerations

### Data Retention

Shared conversations persist until explicitly unshared, including:
- Full conversation history
- Metadata

### Key Recommendations

1. **Review before sharing:** Check content for sensitive information
2. **Avoid sensitive data:** Do not share conversations containing:
   - Proprietary code
   - Confidential data
   - API keys or credentials
3. **Unshare after collaboration:** Remove access when no longer needed
4. **Disable for sensitive projects:** Use `"share": "disabled"` for confidential work

## Enterprise Options

For organizational deployments, sharing can be:

| Option | Description |
|--------|-------------|
| Disabled | Completely prevent sharing for compliance |
| SSO-restricted | Allow only authenticated users |
| Self-hosted | Host share pages on internal infrastructure |

## Commands Reference

| Command | Description |
|---------|-------------|
| `/share` | Generate shareable link (copies to clipboard) |
| `/unshare` | Remove public access and delete data |

## URL Format

Shared conversations use the format:
```
https://opncd.ai/s/<share-id>
```
