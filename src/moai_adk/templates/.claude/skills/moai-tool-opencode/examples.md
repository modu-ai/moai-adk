# OpenCode.ai Examples

Practical code examples for common OpenCode configurations and integrations.

---

## Configuration Examples

### Complete opencode.json Configuration

```json
{
  "$schema": "https://opencode.ai/config.json",
  "model": "anthropic/claude-sonnet-4-5",
  "small_model": "anthropic/claude-haiku-3-5",
  "theme": "tokyonight",
  "autoupdate": "notify",

  "tui": {
    "scroll_speed": 3,
    "scroll_acceleration": { "enabled": true },
    "diff_style": "auto"
  },

  "server": {
    "port": 4096,
    "hostname": "127.0.0.1",
    "mdns": false,
    "cors": ["http://localhost:3000"]
  },

  "provider": {
    "anthropic": {
      "options": {
        "timeout": 600000,
        "setCacheKey": true
      }
    }
  },

  "agent": {
    "reviewer": {
      "description": "Code review specialist",
      "model": "anthropic/claude-sonnet-4-5",
      "tools": { "write": false, "bash": false },
      "permission": { "edit": "deny" },
      "mode": "primary"
    }
  },

  "tools": {
    "bash": true,
    "write": true,
    "edit": true
  },

  "permission": {
    "*": "ask",
    "read": "allow",
    "glob": "allow",
    "grep": "allow",
    "bash": {
      "*": "ask",
      "git *": "allow",
      "npm *": "allow",
      "pnpm *": "allow",
      "bun *": "allow"
    }
  },

  "formatter": {
    "prettier": {
      "command": ["npx", "prettier", "--write", "$FILE"],
      "extensions": [".js", ".ts", ".jsx", ".tsx", ".json", ".md"]
    }
  },

  "lsp": {
    "typescript": { "disabled": false }
  },

  "mcp": {
    "context7": {
      "type": "local",
      "command": ["npx", "-y", "@context7/mcp"],
      "enabled": true
    }
  },

  "instructions": ["AGENTS.md", "CONTRIBUTING.md"],

  "compaction": {
    "auto": true,
    "prune": true
  },

  "watcher": {
    "ignore": ["node_modules/**", "dist/**", ".git/**"]
  },

  "share": "manual"
}
```

### Minimal Project Configuration

```json
{
  "$schema": "https://opencode.ai/config.json",
  "model": "anthropic/claude-sonnet-4-5",
  "instructions": ["AGENTS.md"]
}
```

### Enterprise Configuration

```json
{
  "$schema": "https://opencode.ai/config.json",
  "model": "{env:OPENCODE_MODEL}",
  "provider": {
    "anthropic": {
      "options": {
        "apiKey": "{file:~/.secrets/anthropic-key}",
        "baseURL": "https://internal-gateway.company.com/v1"
      }
    }
  },
  "enabled_providers": ["anthropic"],
  "share": "disabled",
  "permission": {
    "*": "ask",
    "external_directory": "deny"
  }
}
```

---

## Agent Examples

### Markdown Agent Definition

`.opencode/agent/security-reviewer.md`:

```markdown
---
description: Security-focused code reviewer that identifies vulnerabilities
model: anthropic/claude-sonnet-4-5
mode: primary
temperature: 0.2
max_steps: 30
tools:
  write: false
  bash: false
  edit: false
permission:
  edit: deny
  bash: deny
---

You are a security-focused code reviewer. Your role is to:

1. Identify potential security vulnerabilities
2. Check for common attack vectors (XSS, SQL injection, CSRF, etc.)
3. Review authentication and authorization patterns
4. Validate input sanitization
5. Check for sensitive data exposure

When reviewing code:
- Focus on security implications
- Reference OWASP guidelines when relevant
- Provide actionable recommendations
- Rate severity of findings (Critical, High, Medium, Low)

Do NOT:
- Make any code changes
- Execute any commands
- Access external systems
```

### JSON Agent Definition

```json
{
  "agent": {
    "test-writer": {
      "description": "Writes comprehensive test suites",
      "model": "anthropic/claude-sonnet-4-5",
      "temperature": 0.3,
      "max_steps": 50,
      "tools": {
        "bash": false
      },
      "permission": {
        "bash": "deny"
      },
      "prompt": "You are a test writing specialist. Focus on writing comprehensive, maintainable tests with good coverage. Use appropriate testing frameworks for the project."
    },
    "documenter": {
      "description": "Generates and updates documentation",
      "model": "anthropic/claude-haiku-3-5",
      "temperature": 0.5,
      "tools": {
        "bash": false
      },
      "prompt": "You are a documentation specialist. Write clear, concise documentation including README files, API docs, and inline comments."
    }
  }
}
```

---

## Custom Command Examples

### File-Based Command

`.opencode/command/test.md`:

```markdown
---
description: Run tests with coverage report
agent: build
---

Run the test suite for this project with coverage reporting.

Focus on: $ARGUMENTS

1. Identify the testing framework used
2. Run tests with coverage enabled
3. Report any failing tests
4. Summarize coverage metrics
```

### JSON Commands

```json
{
  "command": {
    "review": {
      "template": "Review the code in @$1 for best practices, potential bugs, and improvements.",
      "description": "Code review for a specific file",
      "agent": "reviewer"
    },
    "explain": {
      "template": "Explain how $ARGUMENTS works in this codebase. Include relevant code snippets.",
      "description": "Explain a concept or feature",
      "agent": "plan"
    },
    "refactor": {
      "template": "Refactor @$1 to improve: $2. Maintain existing functionality.",
      "description": "Refactor a file with specific goals",
      "agent": "build"
    },
    "deps": {
      "template": "Analyze dependencies:\n\n`!cat package.json`\n\nCheck for:\n1. Outdated packages\n2. Security vulnerabilities\n3. Unused dependencies",
      "description": "Analyze project dependencies",
      "agent": "plan"
    }
  }
}
```

---

## Custom Tool Examples

### Simple Tool

`.opencode/tool/greet.ts`:

```typescript
import { tool } from "@opencode-ai/plugin"

export default tool({
  description: "Greet a user by name",
  args: {
    name: tool.schema.string().describe("The name to greet")
  },
  async execute(args) {
    return `Hello, ${args.name}! Welcome to OpenCode.`
  }
})
```

### Tool with External Script

`.opencode/tool/analyze.ts`:

```typescript
import { tool } from "@opencode-ai/plugin"
import { $ } from "bun"

export default tool({
  description: "Run static analysis on a file",
  args: {
    path: tool.schema.string().describe("Path to the file to analyze")
  },
  async execute(args, context) {
    const result = await $`python scripts/analyze.py ${args.path}`.text()
    return result
  }
})
```

### Tool with Multiple Exports

`.opencode/tool/git.ts`:

```typescript
import { tool } from "@opencode-ai/plugin"
import { $ } from "bun"

export const status = tool({
  description: "Get git status",
  args: {},
  async execute() {
    return await $`git status --short`.text()
  }
})

export const diff = tool({
  description: "Get git diff for a file",
  args: {
    path: tool.schema.string().optional().describe("File path (optional)")
  },
  async execute(args) {
    if (args.path) {
      return await $`git diff ${args.path}`.text()
    }
    return await $`git diff`.text()
  }
})

export const log = tool({
  description: "Get recent git commits",
  args: {
    count: tool.schema.number().default(5).describe("Number of commits")
  },
  async execute(args) {
    return await $`git log --oneline -n ${args.count}`.text()
  }
})
```

---

## Plugin Examples

### Basic Plugin

`.opencode/plugin/logger.ts`:

```typescript
export const LoggerPlugin = async ({ project, client }) => {
  console.log(`Plugin loaded for project: ${project.name}`)

  return {
    "session.created": async (event) => {
      console.log(`Session created: ${event.sessionID}`)
    },
    "session.deleted": async (event) => {
      console.log(`Session deleted: ${event.sessionID}`)
    },
    "tool.execute.before": async (event) => {
      console.log(`Tool ${event.tool} called with:`, event.input)
    },
    "tool.execute.after": async (event) => {
      console.log(`Tool ${event.tool} completed`)
    }
  }
}
```

### Security Plugin

`.opencode/plugin/security.ts`:

```typescript
export const SecurityPlugin = async ({ project }) => {
  const sensitivePatterns = [
    /\.env$/,
    /\.env\.\w+$/,
    /credentials\.json$/,
    /secrets\.yaml$/,
    /\.pem$/,
    /\.key$/
  ]

  return {
    "tool.execute.before": async (event) => {
      if (event.tool === "read" || event.tool === "edit") {
        const path = event.input.path || event.input.file
        if (path && sensitivePatterns.some(p => p.test(path))) {
          throw new Error(`Access to sensitive file denied: ${path}`)
        }
      }

      if (event.tool === "bash") {
        const command = event.input.command
        const dangerousCommands = ["rm -rf", "sudo", "chmod 777"]
        if (dangerousCommands.some(dc => command.includes(dc))) {
          throw new Error(`Dangerous command blocked: ${command}`)
        }
      }
    }
  }
}
```

### Notification Plugin

`.opencode/plugin/notify.ts`:

```typescript
import { $ } from "bun"

export const NotifyPlugin = async ({ project }) => {
  const notify = async (title: string, message: string) => {
    // macOS notification
    await $`osascript -e 'display notification "${message}" with title "${title}"'`
  }

  return {
    "session.idle": async (event) => {
      await notify("OpenCode", "Session is idle - ready for next task")
    },
    "tool.execute.after": async (event) => {
      if (event.tool === "bash" && event.duration > 30000) {
        await notify("OpenCode", `Long-running command completed: ${event.tool}`)
      }
    }
  }
}
```

---

## SDK Examples

### Basic Client Usage

```typescript
import { createOpencode, createOpencodeClient } from "@opencode-ai/sdk"

// Create full instance
const { client, server } = await createOpencode({
  port: 4096,
  hostname: "127.0.0.1"
})

// Or connect to existing server
const client = createOpencodeClient({
  baseUrl: "http://localhost:4096"
})

// Health check
const health = await client.global.health()
console.log("Server status:", health.status)

// List sessions
const sessions = await client.session.list()
console.log("Sessions:", sessions)
```

### Session Management

```typescript
import { createOpencodeClient } from "@opencode-ai/sdk"

const client = createOpencodeClient({ baseUrl: "http://localhost:4096" })

// Create session
const session = await client.session.create({
  projectID: "my-project"
})

// Send message and wait for response
const response = await client.session.message(session.id, {
  content: "Explain the authentication flow in this codebase"
})

console.log("Response:", response.content)

// Inject context without AI response
await client.session.message(session.id, {
  content: "Remember: Always use TypeScript strict mode",
  noReply: true
})

// Export session
const markdown = await client.session.export(session.id)
```

### Event Streaming

```typescript
import { createOpencodeClient } from "@opencode-ai/sdk"

const client = createOpencodeClient({ baseUrl: "http://localhost:4096" })

// Subscribe to events
const unsubscribe = client.event.subscribe((event) => {
  switch (event.type) {
    case "message.created":
      console.log("New message:", event.data.content)
      break
    case "tool.started":
      console.log("Tool started:", event.data.tool)
      break
    case "tool.completed":
      console.log("Tool completed:", event.data.tool)
      break
    case "session.idle":
      console.log("Session idle")
      break
  }
})

// Later: unsubscribe()
```

### File Operations

```typescript
import { createOpencodeClient } from "@opencode-ai/sdk"

const client = createOpencodeClient({ baseUrl: "http://localhost:4096" })

// Search files
const files = await client.file.find({ pattern: "*.ts" })
console.log("TypeScript files:", files)

// Read file content
const content = await client.file.read({ path: "src/index.ts" })
console.log("Content:", content)

// Text search
const matches = await client.file.grep({ pattern: "TODO", path: "src/" })
console.log("TODOs found:", matches)
```

---

## AGENTS.md Examples

### Web Application Project

```markdown
# Project Guidelines

## Overview
This is a Next.js 14 application with TypeScript, Tailwind CSS, and Prisma.

## Code Style
- Use TypeScript strict mode
- Follow React Server Components patterns
- Use `use client` directive only when necessary
- Prefer Tailwind CSS over inline styles

## Testing
- Write tests with Vitest
- Use React Testing Library for component tests
- Aim for 80%+ code coverage

## File Structure
- `app/` - Next.js App Router pages
- `components/` - React components
- `lib/` - Utility functions
- `prisma/` - Database schema

## Common Commands
- `pnpm dev` - Start development server
- `pnpm test` - Run tests
- `pnpm lint` - Run ESLint
- `pnpm db:push` - Push schema to database

## Important Patterns
- Use Server Actions for mutations
- Validate inputs with Zod
- Handle errors with Error Boundaries
- Use Suspense for loading states
```

### Python API Project

```markdown
# Python API Guidelines

## Stack
- Python 3.12+
- FastAPI
- SQLAlchemy 2.0
- Pydantic v2

## Code Style
- Follow PEP 8
- Use type hints everywhere
- Document with docstrings (Google style)
- Maximum line length: 88 (Black formatter)

## Testing
- Use pytest
- Fixtures in conftest.py
- Aim for 90%+ coverage on core logic

## Project Structure
```
src/
  api/         # FastAPI routes
  models/      # SQLAlchemy models
  schemas/     # Pydantic schemas
  services/    # Business logic
  utils/       # Helpers
tests/
  unit/
  integration/
```

## Commands
- `uv run pytest` - Run tests
- `uv run ruff check .` - Lint
- `uv run ruff format .` - Format

## Patterns
- Dependency injection with FastAPI Depends
- Repository pattern for data access
- Service layer for business logic
```

---

## GitHub Actions Workflow

`.github/workflows/opencode.yml`:

```yaml
name: OpenCode Agent

on:
  issue_comment:
    types: [created]
  pull_request_review_comment:
    types: [created]

jobs:
  opencode:
    if: contains(github.event.comment.body, '/oc') || contains(github.event.comment.body, '/opencode')
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: opencode-agent/opencode-action@v1
        with:
          model: anthropic/claude-sonnet-4-20250514
          agent: build
          share: ${{ github.event.repository.private == false }}
        env:
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
```

---

## MCP Server Examples

### Multiple MCP Servers

```json
{
  "mcp": {
    "context7": {
      "type": "local",
      "command": ["npx", "-y", "@context7/mcp"],
      "enabled": true
    },
    "sentry": {
      "type": "remote",
      "url": "https://mcp.sentry.io",
      "oauth": true
    },
    "custom-db": {
      "type": "local",
      "command": ["node", "mcp-servers/db-server.js"],
      "env": {
        "DATABASE_URL": "{env:DATABASE_URL}"
      }
    }
  },
  "tools": {
    "context7_*": true,
    "sentry_*": true,
    "custom-db_*": "ask"
  }
}
```
