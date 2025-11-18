# MCP Integration Guide

**Complete guide for Model Context Protocol integration with MoAI-ADK.**

> **See also**: CLAUDE.md → "MCP Integration & External Services" for quick reference

---

## What is MCP?

**MCP (Model Context Protocol)**: Standard for connecting AI models to external services and data sources.

**Key Features**:
- Seamless external service integration
- Real-time data access
- Standardized resource management
- Automatic tool discovery

---

## Setup Configuration

### .claude/mcp.json Structure

```json
{
  "mcpServers": {
    "context7": {
      "command": "npx",
      "args": ["-y", "@upstash/context7-mcp@latest"],
      "env": {
        "CONTEXT7_SESSION_STORAGE": ".moai/sessions/",
        "CONTEXT7_CACHE_SIZE": "1GB",
        "CONTEXT7_SESSION_TTL": "30d"
      }
    },
    "github": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-server-github"],
      "oauth": {
        "clientId": "your-client-id",
        "clientSecret": "your-client-secret",
        "scopes": ["repo", "issues"]
      }
    },
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/root"]
    }
  }
}
```

---

## Context7 MCP (Documentation Lookup)

**Purpose**: Access real-time documentation for libraries and frameworks

**Install**:
```bash
npx @upstash/context7-mcp@latest
```

### Usage Pattern 1: Resolve Library ID

```python
# Get library ID from name
library_id = mcp__context7__resolve-library-id("React")
# Returns: /facebook/react
```

### Usage Pattern 2: Get Latest Documentation

```python
# Fetch latest React documentation
docs = mcp__context7__get-library-docs("/facebook/react")
# Returns: Current API reference, examples, patterns
```

### Usage Pattern 3: Version-Specific Docs

```python
# Get React 18.2.0 documentation
docs = mcp__context7__get-library-docs("/facebook/react/18.2.0")
# Returns: Version-specific docs
```

### Session Management

```json
{
  "env": {
    "CONTEXT7_SESSION_STORAGE": ".moai/sessions/",
    "CONTEXT7_CACHE_SIZE": "1GB",
    "CONTEXT7_SESSION_TTL": "30d"
  }
}
```

**Benefits**:
- Persistent session caching
- 15-minute query cache
- 30-day session retention
- 1GB cache limit

---

## GitHub MCP Server

**Purpose**: Access GitHub issues, PRs, repos, workflows

**Configuration**:
```json
{
  "github": {
    "command": "npx",
    "args": ["-y", "@anthropic-ai/mcp-server-github"],
    "oauth": {
      "clientId": "Ov23liXXXXXXXXXXXXXX",
      "clientSecret": "your-secret",
      "scopes": ["repo", "issues", "pull_requests"]
    }
  }
}
```

### Usage Examples

**List Issues**:
```bash
@github list issues --repo moai-adk --label "bug"
```

**Get PR Details**:
```bash
@github get pullrequest --repo moai-adk --number 123
```

**Create Issue**:
```bash
@github create issue --repo moai-adk --title "Fix login"
```

**Get Repository Info**:
```bash
@github get repository --repo moai-adk
```

---

## Filesystem MCP Server

**Purpose**: Navigate and search project files

**Configuration**:
```json
{
  "filesystem": {
    "command": "npx",
    "args": ["-y", "@modelcontextprotocol/server-filesystem", "/Users/goos/MoAI/MoAI-ADK"]
  }
}
```

### Usage Examples

**List Files**:
```bash
@filesystem list /Users/goos/MoAI/MoAI-ADK/src
```

**Search Content**:
```bash
@filesystem search "def authenticate" --type .py
```

**Read File**:
```bash
@filesystem read /Users/goos/MoAI/MoAI-ADK/README.md
```

---

## Notion MCP Server

**Purpose**: Access Notion workspace and databases

**Configuration**:
```json
{
  "notion": {
    "command": "npx",
    "args": ["-y", "notion-mcp"],
    "env": {
      "NOTION_API_TOKEN": "your-api-token"
    }
  }
}
```

### Usage Examples

**Create Database Page**:
```bash
@notion create page --database-id "xxx" --title "SPEC-001"
```

**List Pages**:
```bash
@notion list pages --database-id "xxx"
```

**Update Properties**:
```bash
@notion update page --page-id "xxx" --status "completed"
```

---

## Best MCP Practices

### Discovery Pattern

```bash
# 1. Find available resources
/mcp list resources

# 2. Check server status
/mcp status

# 3. Explore capabilities
claude /doctor
```

### Usage Pattern

```python
# Direct MCP tool calls (80% of cases)
result = mcp__context7__get-library-docs("/facebook/react")

# MCP Agent integration (20% complex cases)
Task(subagent_type="mcp-context7-integrator", prompt="...")
```

### Error Handling

```bash
# Connection failed?
/mcp restart

# Invalid library ID?
lib_id = mcp__context7__resolve-library-id("react")
# → Returns correct format

# Rate limited?
# Wait 1 minute and retry
```

---

## Integration with MoAI-ADK Workflow

### Phase 1: Planning with Context7

```bash
/moai:1-plan "Implement React hooks"
# MCP Context7 fetches latest React documentation
# → Includes current best practices
```

### Phase 2: Implementation

```bash
/moai:2-run SPEC-REACT-001
# MCP GitHub checks for similar issues
# → Learns from existing solutions
```

### Phase 3: Documentation

```bash
/moai:3-sync auto SPEC-REACT-001
# MCP Notion updates project tracking
# → Records completed feature
```

---

## Performance Optimization

### Caching Strategy

```json
{
  "env": {
    "CONTEXT7_CACHE_SIZE": "1GB",
    "CONTEXT7_SESSION_TTL": "30d"
  }
}
```

**Benefits**:
- 15-minute query cache (automatic)
- Persistent session storage
- Avoid rate limits
- Faster documentation lookup

### Resource Limits

```json
{
  "env": {
    "MCP_TIMEOUT": "30s",
    "MCP_MAX_RETRIES": "3",
    "MCP_CACHE_TTL": "1h"
  }
}
```

---

## Security Considerations

### Credential Management

```json
{
  "oauth": {
    "clientId": "Ov23liXXXXXXXXXXXXXX",
    "clientSecret": "ghp_xxxxxxxxxxxxxxx"
  }
}
```

**Best Practices**:
- Never commit credentials to git
- Use environment variables
- Rotate tokens regularly
- Audit access logs

### Rate Limiting

**GitHub API**: 60 requests per hour (unauthenticated), 5000 (authenticated)

**Context7**: 1000 requests per day

**Notion**: 3 requests per second

---

## Troubleshooting

### MCP Server Not Found

```bash
# Install missing server
npx @upstash/context7-mcp@latest

# Verify installation
which context7-mcp
```

### Connection Error

```bash
# Restart MCP servers
/mcp restart

# Check connectivity
curl -I https://api.context7.io
```

### Authentication Failed

```bash
# Verify credentials
echo $GITHUB_TOKEN  # Should not be empty
echo $NOTION_API_TOKEN  # Should not be empty

# Update .claude/mcp.json with correct credentials
```

### Rate Limited

```bash
# Wait before retrying (backoff strategy)
# Context7: Wait 1 minute
# GitHub: Wait 1 hour
```

---

## Advanced Features

### Session Sharing Across Agents

```python
# Agent 1: Research with Context7
research = Task(
    subagent_type="mcp-context7-integrator",
    prompt="Research React patterns"
)

# Agent 2: Uses research findings
implementation = Task(
    subagent_type="frontend-expert",
    prompt=f"Implement based on: {research}",
    shared_session=research.session_id
)
```

### Automated Workflows

```bash
# Day 1: Research documentation
/moai:0-project
# MCP Context7 auto-loads current docs

# Day 2: Create issue on GitHub
/moai:1-plan "Fix authentication bug"
# MCP GitHub checks existing issues

# Day 3: Record completion
/moai:3-sync auto
# MCP Notion marks task done
```

---

## Checklist

- [ ] .claude/mcp.json created
- [ ] Context7 installed and configured
- [ ] GitHub credentials set
- [ ] Filesystem access configured
- [ ] MCP servers tested (`/mcp status`)
- [ ] Environment variables set
- [ ] Credentials secured (not in git)
- [ ] Rate limits understood
- [ ] Error handling tested
- [ ] Performance baseline established

---

**Last Updated**: 2025-11-18
**Version**: v0.26.0
**Format**: Markdown | **Language**: English
**Context7 MCP**: Latest (@upstash/context7-mcp@latest)
**GitHub MCP**: Latest (@anthropic-ai/mcp-server-github)
**Filesystem MCP**: Latest (@modelcontextprotocol/server-filesystem)
