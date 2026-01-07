---
source: https://opencode.ai/docs/custom-tools/
fetched: 2026-01-08
title: Custom Tools
---

# Custom Tools in OpenCode

## Overview

OpenCode allows developers to create custom tools that language models can invoke during conversations, complementing built-in tools like `read`, `write`, and `bash`.

## Key Implementation Details

### Tool Definition Language

Tools are defined using TypeScript or JavaScript, though they can execute scripts in any programming language.

### File Locations

| Location | Scope |
|----------|-------|
| `.opencode/tool/` | Project |
| `~/.config/opencode/tool/` | Global |

## Basic Structure

The `tool()` helper from `@opencode-ai/plugin` provides type safety:

```typescript
import { tool } from "@opencode-ai/plugin"

export default tool({
  description: "Query the project database",
  args: {
    query: tool.schema.string().describe("SQL query to execute"),
  },
  async execute(args) {
    return `Executed query: ${args.query}`
  },
})
```

The filename becomes the tool name. This example creates a `database` tool if saved as `database.ts`.

## Multiple Tools Per File

Export multiple named tools to create `<filename>_<exportname>` tools:

```typescript
import { tool } from "@opencode-ai/plugin"

export const add = tool({
  description: "Add two numbers",
  args: {
    a: tool.schema.number().describe("First number"),
    b: tool.schema.number().describe("Second number"),
  },
  async execute(args) {
    return String(args.a + args.b)
  },
})

export const multiply = tool({
  description: "Multiply two numbers",
  args: {
    a: tool.schema.number().describe("First number"),
    b: tool.schema.number().describe("Second number"),
  },
  async execute(args) {
    return String(args.a * args.b)
  },
})
```

If saved as `math.ts`, this creates `math_add` and `math_multiply` tools.

## Arguments & Validation

Use `tool.schema` (powered by Zod) for argument definitions:

```typescript
import { tool } from "@opencode-ai/plugin"

export default tool({
  description: "Create a new user",
  args: {
    name: tool.schema.string().min(1).describe("User's name"),
    email: tool.schema.string().email().describe("User's email"),
    age: tool.schema.number().optional().describe("User's age"),
    role: tool.schema.enum(["admin", "user", "guest"]).describe("User role"),
  },
  async execute(args) {
    // Validation happens automatically
    return `Created user: ${args.name}`
  },
})
```

### Schema Types

| Type | Method |
|------|--------|
| String | `tool.schema.string()` |
| Number | `tool.schema.number()` |
| Boolean | `tool.schema.boolean()` |
| Enum | `tool.schema.enum([...])` |
| Optional | `.optional()` |
| Array | `tool.schema.array(...)` |
| Object | `tool.schema.object({...})` |

## Context Access

Tools receive session context through the `context` parameter:

```typescript
import { tool } from "@opencode-ai/plugin"

export default tool({
  description: "Get session info",
  args: {},
  async execute(args, context) {
    return JSON.stringify({
      agent: context.agent,
      sessionId: context.sessionId,
      messageId: context.messageId,
    })
  },
})
```

## Cross-Language Example

Tools can invoke scripts in other languages:

### TypeScript calling Python

```typescript
import { tool } from "@opencode-ai/plugin"

export default tool({
  description: "Run Python calculation",
  args: {
    a: tool.schema.number().describe("First number"),
    b: tool.schema.number().describe("Second number"),
  },
  async execute(args) {
    const result = await Bun.$`python3 .opencode/tool/add.py ${args.a} ${args.b}`.text()
    return result.trim()
  },
})
```

### Python script (add.py)

```python
import sys

a = float(sys.argv[1])
b = float(sys.argv[2])
print(a + b)
```

## Full Example: API Client

```typescript
import { tool } from "@opencode-ai/plugin"

export default tool({
  description: "Make HTTP requests to external APIs",
  args: {
    url: tool.schema.string().url().describe("URL to fetch"),
    method: tool.schema.enum(["GET", "POST", "PUT", "DELETE"]).default("GET"),
    body: tool.schema.string().optional().describe("Request body"),
  },
  async execute(args) {
    const response = await fetch(args.url, {
      method: args.method,
      body: args.body,
      headers: {
        "Content-Type": "application/json",
      },
    })

    const data = await response.text()
    return data
  },
})
```

## Best Practices

1. **Clear descriptions**: Help the LLM understand when to use the tool
2. **Validate inputs**: Use schema validation for safety
3. **Handle errors**: Return helpful error messages
4. **Keep focused**: Each tool should do one thing well
5. **Document args**: Use `.describe()` for all arguments
