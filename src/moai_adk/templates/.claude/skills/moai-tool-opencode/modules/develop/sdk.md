---
source: https://opencode.ai/docs/sdk/
fetched: 2026-01-08
title: SDK
---

# OpenCode SDK Documentation

## Overview

The OpenCode SDK is a type-safe JavaScript/TypeScript client for interacting with the OpenCode server. It enables programmatic control and integration capabilities.

## Installation

```bash
npm install @opencode-ai/sdk
```

## Setup

### Full Client (Starts Server)

Create a client instance that starts both server and client:

```javascript
import { createOpencode } from "@opencode-ai/sdk"

const { client } = await createOpencode()
```

### Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `hostname` | `127.0.0.1` | Server hostname |
| `port` | `4096` | Server port |
| `abortSignal` | - | Abort signal for cancellation |
| `timeout` | `5000` | Request timeout in ms |
| `config` | - | Configuration object |

### Client-Only Mode

For existing server instances:

```javascript
import { createOpencodeClient } from "@opencode-ai/sdk"

const client = createOpencodeClient({
  baseUrl: "http://localhost:4096"
})
```

## Core APIs

### Global

Health checks returning version information:

```javascript
const health = await client.global.health()
console.log(health.version)
```

### App

Logging entries and listing available agents:

```javascript
// Log an entry
await client.app.log({ level: "info", message: "Hello" })

// List agents
const agents = await client.app.agents()
```

### Project

List and retrieve current project information:

```javascript
const project = await client.project.get()
console.log(project.name, project.path)
```

### Sessions

Comprehensive session management:

```javascript
// Create session
const session = await client.session.create({
  title: "My Session"
})

// List sessions
const sessions = await client.session.list()

// Get session
const existing = await client.session.get(sessionId)

// Delete session
await client.session.delete(sessionId)

// Share session
const shareUrl = await client.session.share(sessionId)
```

### Files

Search capabilities and file reading:

```javascript
// Text search
const results = await client.file.search({
  query: "function",
  type: "text"
})

// Read file
const content = await client.file.read(filePath)

// File status
const status = await client.file.status(filePath)
```

### TUI

Terminal UI control:

```javascript
// Append prompt
await client.tui.prompt({ text: "Hello" })

// Show toast
await client.tui.toast({ message: "Success!" })

// Execute command
await client.tui.command({ name: "init" })

// Open dialog
await client.tui.dialog({ type: "confirm", message: "Proceed?" })
```

### Auth

Set authentication credentials for providers:

```javascript
await client.auth.set({
  provider: "anthropic",
  apiKey: "sk-..."
})
```

### Events

Server-sent events stream for real-time updates:

```javascript
const events = client.events.subscribe()

for await (const event of events) {
  console.log(event.type, event.data)
}
```

## Session Operations

### Send Message

```javascript
const response = await client.session.prompt(sessionId, {
  content: "Explain this code",
  context: "@src/index.ts"
})
```

### Execute Command

```javascript
await client.session.command(sessionId, {
  name: "init"
})
```

### Execute Shell

```javascript
const output = await client.session.shell(sessionId, {
  command: "npm test"
})
```

### Share Session

```javascript
const url = await client.session.share(sessionId)
```

### Context Injection (No Reply)

Inject context without triggering AI response:

```javascript
await client.session.prompt(sessionId, {
  content: "Context information",
  noReply: true
})
```

## Types & Error Handling

### TypeScript Types

```typescript
import type {
  Session,
  Message,
  Part,
  Project,
  Agent
} from "@opencode-ai/sdk"
```

### Error Handling

```javascript
try {
  const session = await client.session.get(sessionId)
} catch (error) {
  if (error.status === 404) {
    console.log("Session not found")
  }
}
```

## Full Example

```javascript
import { createOpencode } from "@opencode-ai/sdk"

async function main() {
  // Start server and client
  const { client } = await createOpencode({
    port: 4096
  })

  // Create a session
  const session = await client.session.create({
    title: "Code Review"
  })

  // Send a prompt
  const response = await client.session.prompt(session.id, {
    content: "Review the code in @src/main.ts"
  })

  console.log(response.content)

  // Share the session
  const shareUrl = await client.session.share(session.id)
  console.log("Shared at:", shareUrl)
}

main()
```
