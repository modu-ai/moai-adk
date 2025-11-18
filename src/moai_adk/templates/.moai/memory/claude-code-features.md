# Claude Code+ Features & Architecture

**Complete guide for Claude Code integration with MoAI-ADK.**

> **See also**: CLAUDE.md → "Claude Code Features" for quick overview

---

## 4-Layer Modern Architecture

**Commands (Orchestration) → Agents (Expertise) → Skills (Knowledge) → Hooks (Guardrails)**

### 1. Commands (Workflow Orchestration)

Commands are user-facing entry points that trigger workflows:

- `/moai:0-project` - Project initialization and auto-detection
- `/moai:1-plan` - SPEC generation with planning mode
- `/moai:2-run` - TDD implementation with agent delegation
- `/moai:3-sync` - Auto-documentation and synchronization

**Features**:
- Enhanced with Plan Mode for complex tasks
- Interactive Questions for clarification
- Automatic context optimization
- Parallel execution support

---

### 2. Sub-agents (Domain Expertise)

Specialized agents handle specific domains:

**Priority 1: MoAI-ADK Agents**
- `spec-builder` - SPEC-First requirements (EARS format)
- `tdd-implementer` - TDD Red-Green-Refactor cycle
- `backend-expert` - API design, microservices
- `frontend-expert` - React/Vue/Angular components
- `database-expert` - Schema design, optimization
- `security-expert` - OWASP, encryption, auth
- `docs-manager` - Auto-documentation
- `performance-engineer` - Load testing, profiling
- `monitoring-expert` - Observability, logging, alerting

**Priority 2: MoAI-ADK Skills**
- `moai-lang-python` - FastAPI, Pydantic, SQLAlchemy
- `moai-lang-typescript` - Next.js, TypeScript, Zod
- `moai-lang-go` - Fiber, gRPC, concurrency
- `moai-domain-backend` - Server architecture patterns
- `moai-domain-frontend` - UI component design
- `moai-domain-security` - OWASP Top 10, threat modeling

**Priority 3: Claude Code Native Agents**
- `Explore` - Fast codebase pattern discovery
- `Plan` - Task decomposition
- `debug-helper` - Runtime error analysis

**Features**:
- Model selection optimization (Sonnet/Haiku)
- MCP server integration
- Parallel execution support
- Context window management

---

### 3. Skills (Knowledge Progressive Disclosure)

Skills contain reusable knowledge patterns:

- Lazy loading for performance
- Cross-skill references
- Version-controlled knowledge
- Code snippets and examples
- Best practices library

---

### 4. Hooks (Context & Guardrails)

Hooks provide automated checks and context seeding:

- **PreToolUse**: Command validation (sandbox mode)
- **PostToolUse**: Quality checks
- **SessionStart**: Context initialization

---

## Claude Code Key Features

### Plan Mode

**What**: Automatically breaks down complex tasks into phases

**When to use**: Complex multi-domain tasks (3+ domains, 2+ hours)

**Example**:
```bash
/moai:1-plan "Build full-stack payment system"
# → Automatically decomposes into:
# - Stripe API integration (backend-expert)
# - Payment UI (frontend-expert)
# - Database schema (database-expert)
# - Security audit (security-expert)
# - Monitoring setup (monitoring-expert)
```

**Benefits**:
- Parallel execution (3-5x faster)
- Specialized expertise per task
- Automatic sequencing and dependencies
- Context window optimization

---

### Explore Subagent

**What**: Fast codebase pattern discovery using Haiku 4.5

**When to use**: Rapid code search, pattern discovery

**Example**:
```bash
"Where are error handling patterns implemented?"
# → Explore agent finds and summarizes patterns
# → Efficient context usage
```

**Benefits**:
- 70% cheaper than Sonnet
- Fast pattern matching
- Efficient context summarization

---

### MCP Integration

**What**: Model Context Protocol for external service connection

**Supported Services**:
- `@github` - List issues, PRs, repos
- `@filesystem` - File search and navigation
- `@context7` - Documentation lookup
- `@notion` - Workspace management

**Example**:
```bash
@github list issues --label "bug"
mcp__context7__get-library-docs("/facebook/react")
```

**Benefits**:
- External data access without API keys
- Real-time information integration
- Seamless tool orchestration

---

### Context Management

**Commands**:
- `/context` - Check current usage
- `/compact` - Compress conversation
- `/clear` - Clear session
- `/add-dir src/` - Add directory
- `/memory` - Manage persistent memory
- `/cost` - View API costs

**Best Practices**:
- Use `/clear` after SPEC generation
- Monitor `/context` regularly
- Use `/compact` before limits
- Keep memory file concise

---

### Interactive Questions

**Pattern**: Structured decision making with user collaboration

**Usage**:
```json
{
  "questions": [{
    "question": "Implementation approach preference?",
    "header": "Architecture",
    "multiSelect": false,
    "options": [
      {"label": "Standard", "description": "Proven pattern"},
      {"label": "Optimized", "description": "Performance-focused"}
    ]
  }]
}
```

**Benefits**:
- Explicit user intent gathering
- Better decision outcomes
- Collaborative development

---

### Thinking Mode

**What**: Transparent multi-step reasoning (Tab key toggle)

**How**: Enable with Tab key in Claude Code

**Benefits**:
- See reasoning steps
- Verify logic before execution
- Better understanding of decisions

---

## Model Selection Strategy

### Sonnet 4.5 (Expensive but Smart)

**Cost**: $0.003 per 1K tokens

**Use for**:
- SPEC creation (complex reasoning required)
- Architecture decisions
- Security reviews
- Complex refactoring

**Typical usage**: 5-10 tasks per project

---

### Haiku 4.5 (Fast & Cheap)

**Cost**: $0.0008 per 1K tokens (70% cheaper!)

**Use for**:
- Codebase exploration
- Simple implementations
- Code formatting
- Test execution
- Bug fixes

**Typical usage**: 30+ tasks per project

---

## Token Efficiency Strategies

### 1. Agent Delegation

**Before**: Full codebase loaded (130,000+ tokens)
**After**: Focused agents (20-30,000 tokens each)
**Savings**: 80-85% reduction

### 2. Context Pruning

- Load only relevant files per agent
- Exclude large directories (node_modules, dist)
- Use Progressive Disclosure for docs

### 3. Session Management

Use `/clear` after major phases:
```
Phase 1: SPEC creation (50K tokens) → /clear
Phase 2: Implementation (60K tokens) → /clear
Phase 3: Testing (30K tokens)
Total: 140K vs 300K+ (53% savings!)
```

### 4. Cache Utilization

- Reuse SPEC documents
- Cache large codebases
- Share agent sessions

---

## Configuration Example

**.claude/mcp.json**:
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
      "args": ["-y", "@anthropic-ai/mcp-server-github"]
    }
  }
}
```

---

## Best Practices

### ✅ Do's

- ✅ Use Plan Mode for complex tasks (3+ domains)
- ✅ Leverage Explore for fast discovery
- ✅ Use Haiku for simple tasks (70% cheaper)
- ✅ Monitor context with `/context`
- ✅ Use `/clear` between major phases
- ✅ Utilize MCP services for real-time data

### ❌ Don'ts

- ❌ Use Sonnet for simple tasks (expensive)
- ❌ Load full codebase in every agent (inefficient)
- ❌ Skip context optimization (token waste)
- ❌ Use Plan Mode for simple bugs (overkill)
- ❌ Ignore `/context` warnings (quota overrun)

---

**Last Updated**: 2025-11-18
**Version**: v0.26.0
**Format**: Markdown | **Language**: English
