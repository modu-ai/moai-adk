---
name: memory-mcp-optimization
parent: moai-core-context-budget
description: Memory file management and MCP configuration
---

# Module 2: Memory & MCP Optimization

## Memory File Structure

```
.moai/memory/
├── session-summary.md     # < 300 lines
├── architectural-decisions.md   # < 400 lines
├── api-contracts.md      # < 200 lines
├── known-issues.md       # < 150 lines
└── team-conventions.md   # < 200 lines

Total: < 1,250 lines (~25K tokens)
```

## Memory File Template

```markdown
# Session Summary
**Last Updated**: 2025-01-12
**Current Sprint**: Feature/Auth-Refactor

## Completed This Session
1. JWT authentication (commit: abc123)
2. Password hashing (commit: def456)

## In Progress
1. OAuth2 integration (70%)
   - Files: src/auth/oauth.ts

## Key Decisions
**Auth**: JWT in httpOnly cookies (XSS prevention)

## Next Actions
1. Complete OAuth callback
2. Add tests
```

## Memory File Rotation

```bash
# Rotate when exceeding 500 lines
rotate_memory_file() {
    local file="$1"
    if [[ $(wc -l < "$file") -gt 500 ]]; then
        # Archive old content
        mkdir -p .moai/memory/archive
        # Keep last 300 lines
        tail -n 300 "$file" > "${file}.tmp"
        mv "$file" .moai/memory/archive/
        mv "${file}.tmp" "$file"
    fi
}
```

## MCP Context Impact

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

**MCP Overhead**: ~847 tokens per server

## MCP Usage Strategy

- **Documentation phase**: Enable Context7
- **Testing phase**: Enable Playwright
- **Code review**: Enable GitHub
- **Minimal**: Disable all for maximum context

---

**Reference**: [MCP Configuration](https://modelcontextprotocol.io)
