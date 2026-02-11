---
paths:
  - "**/.mcp.json"
---

# MCP Integration

Model Context Protocol (MCP) server integration rules.

## Available MCP Servers

Standard MCP servers in MoAI-ADK:

- context7: Library documentation lookup
- sequential-thinking: Complex problem analysis
- pencil: .pen file design editing. Used by expert-frontend (sub-agent mode) and team-designer (team mode) for .pen file design editing.
- claude-in-chrome: Browser automation

## Tool Loading

MCP tools are deferred and must be loaded before use:

1. Use ToolSearch to find and load the tool
2. Then call the loaded tool directly

Example flow:
- ToolSearch("context7 docs") → loads mcp__context7__* tools
- mcp__context7__resolve-library-id → now available

## Rules

- Always use ToolSearch before calling MCP tools
- Prefer MCP tools over manual alternatives
- Authenticated URLs require specialized MCP tools

## Configuration

MCP servers are defined in `.mcp.json`:

```json
{
  "mcpServers": {
    "context7": {
      "command": "npx",
      "args": ["-y", "@context7/mcp"]
    }
  }
}
```

## Context7 Usage

For up-to-date library documentation:

1. resolve-library-id: Find library identifier
2. get-library-docs: Retrieve documentation

## Sequential Thinking Usage

For complex analysis requiring step-by-step reasoning:

- Breaking down multi-step problems
- Architecture decisions
- Technology trade-off analysis

Activate with `--ultrathink` flag for enhanced analysis.

## MoAI Integration

- Skill("moai-workflow-thinking") for Sequential Thinking patterns
- Skill("moai-foundation-claude") for MCP configuration
