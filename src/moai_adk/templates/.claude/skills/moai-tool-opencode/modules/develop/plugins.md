---
source: https://opencode.ai/docs/plugins/
fetched: 2026-01-08
title: Plugins
---

# OpenCode Plugins Documentation

## Overview

OpenCode enables extensibility through plugins - JavaScript/TypeScript modules that hook into various system events. Developers can create custom functionality by implementing event handlers and returning a hooks object from their plugin function.

## Loading Plugins

### Local Plugins

Reside in project-level or global directories and load automatically:

| Location | Scope |
|----------|-------|
| `.opencode/plugin/` | Project |
| `~/.config/opencode/plugin/` | Global |

### NPM Plugins

Specify in configuration:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "plugin": [
    "opencode-helicone-session",
    "@my-org/custom-plugin"
  ]
}
```

Both scoped and unscoped packages work. The system installs dependencies via Bun and caches them in `~/.cache/opencode/node_modules/`.

## Plugin Structure

A basic plugin exports an async function receiving a context object:

```javascript
export const MyPlugin = async ({ project, client, $, directory, worktree }) => {
  // Setup code here

  return {
    // Hook implementations
  }
}
```

### Context Object

| Property | Description |
|----------|-------------|
| `project` | Project metadata |
| `client` | SDK client instance |
| `$` | Bun's shell API |
| `directory` | Plugin directory path |
| `worktree` | Git worktree path |

## Custom Tools

Extend OpenCode's functionality using the `@opencode-ai/plugin` package:

```typescript
import { tool } from "@opencode-ai/plugin"

export default tool({
  description: "Custom tool description",
  args: {
    input: tool.schema.string().describe("Input parameter"),
  },
  async execute(args) {
    return `Result: ${args.input}`
  },
})
```

## Event Subscriptions

React to various system events:

```javascript
export const MyPlugin = async (ctx) => {
  return {
    "session.created": async (event) => {
      console.log("Session created:", event.session.id)
    },
    "file.written": async (event) => {
      console.log("File written:", event.path)
    },
    "tool.executing": async (event) => {
      console.log("Tool executing:", event.tool)
    }
  }
}
```

### Available Events

| Event | Description |
|-------|-------------|
| `session.created` | Session created |
| `session.deleted` | Session deleted |
| `file.written` | File written |
| `file.deleted` | File deleted |
| `lsp.diagnostic` | LSP diagnostic received |
| `permission.requested` | Permission requested |
| `tool.executing` | Tool about to execute |
| `tool.executed` | Tool finished executing |
| `experimental.session.compacting` | Session compaction |

## Structured Logging

Use `client.app.log()` with severity levels:

```javascript
export const MyPlugin = async ({ client }) => {
  await client.app.log({
    level: "info",
    message: "Plugin initialized"
  })

  await client.app.log({
    level: "warn",
    message: "Something might be wrong"
  })

  await client.app.log({
    level: "error",
    message: "Something went wrong"
  })

  return {}
}
```

**Important:** Use this instead of `console.log()` for proper log handling.

## Session Compaction Hooks

Inject domain-specific context before LLM-generated continuation summaries:

```javascript
export const MyPlugin = async (ctx) => {
  return {
    "experimental.session.compacting": async (event) => {
      return {
        context: "Important context to preserve during compaction"
      }
    }
  }
}
```

## Dependencies

Local plugins requiring external packages need a `package.json` in the config directory:

```json
{
  "dependencies": {
    "axios": "^1.0.0",
    "lodash": "^4.0.0"
  }
}
```

OpenCode processes this automatically at startup.

## Full Example

```javascript
import { tool } from "@opencode-ai/plugin"

// Custom tool
const myTool = tool({
  description: "Perform custom operation",
  args: {
    data: tool.schema.string().describe("Data to process"),
  },
  async execute(args) {
    return `Processed: ${args.data}`
  },
})

// Plugin with hooks
export const MyPlugin = async ({ client, project }) => {
  // Initialize
  await client.app.log({
    level: "info",
    message: `Plugin loaded for ${project.name}`
  })

  return {
    "session.created": async (event) => {
      await client.app.log({
        level: "debug",
        message: `New session: ${event.session.id}`
      })
    },

    "tool.executed": async (event) => {
      if (event.tool === "bash") {
        await client.app.log({
          level: "info",
          message: `Bash command executed: ${event.input}`
        })
      }
    }
  }
}

export { myTool }
```
