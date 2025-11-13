---
name: "moai-alfred-context-budget"
version: "4.0.0"
tier: Alfred
description: "Enterprise Claude Code context window optimization with 2025 best practices: aggressive clearing, memory file management, MCP optimization, strategic chunking, and quality-over-quantity principles for 200K token context windows."
primary-agent: "alfred"
secondary-agents: ["session-manager", "plan-agent"]
keywords: ["context", "optimization", "memory", "tokens", "performance", "mcp", "claude-code"]
status: stable
---

# moai-alfred-context-budget

**Enterprise Context Window Optimization for Claude Code**

## Overview

Enterprise-grade context window management for Claude Code covering 200K token optimization, aggressive clearing strategies, memory file management, MCP server optimization, and 2025 best practices for maintaining high-quality AI-assisted development sessions.

**Core Capabilities**:
- ✅ Context budget allocation (200K tokens)
- ✅ Aggressive context clearing patterns
- ✅ Memory file optimization (<500 lines each)
- ✅ MCP server efficiency monitoring
- ✅ Strategic chunking for long-running tasks
- ✅ Quality-over-quantity principles

---

## Quick Reference

### When to Use This Skill

**Automatic Activation**:
- Context window approaching 80% usage
- Performance degradation detected
- Session handoff preparation
- Large project context management

**Manual Invocation**:
```
Skill("moai-alfred-context-budget")
```

### Key Principles (2025)

1. **Avoid Last 20%** - Performance degrades in final fifth of context
2. **Aggressive Clearing** - `/clear` every 1-3 messages for quality
3. **Lean Memory Files** - Keep each file < 500 lines
4. **Disable Unused MCPs** - Each server adds tool definitions
5. **Quality > Quantity** - 10% with relevant info beats 90% with noise

---

## Pattern 1: Context Budget Allocation

### Overview

Claude Code provides 200K token context window with strategic allocation across system, tools, history, and working context.

### Context Budget Breakdown

```yaml
# Claude Code Context Budget (200K tokens)
Total Context Window: 200,000 tokens

Allocation:
  System Prompt: 2,000 tokens (1%)
    - Core instructions
    - CLAUDE.md project guidelines
    - Agent directives
  
  Tool Definitions: 5,000 tokens (2.5%)
    - Read, Write, Edit, Bash, etc.
    - MCP server tools (Context7, Playwright, etc.)
    - Skill() invocation metadata
  
  Session History: 30,000 tokens (15%)
    - Previous messages
    - Tool call results
    - User interactions
  
  Project Context: 40,000 tokens (20%)
    - Memory files (.moai/memory/)
    - Key source files
    - Documentation snippets
  
  Available for Response: 123,000 tokens (61.5%)
    - Current task processing
    - Code generation
    - Analysis output
```

### Monitoring Context Usage

```bash
# Check current context usage
/context

# Example output interpretation:
# Context Usage: 156,234 / 200,000 tokens (78%)
# ⚠️ WARNING: Approaching 80% threshold
# Action: Consider /clear or archive old discussions
```

### Context Budget Anti-Patterns

```yaml
# BAD: Unoptimized Context
Session History: 80,000 tokens (40%)  # Too much history
  - 50 messages of exploratory debugging
  - Stale error logs from 2 hours ago
  - Repeated "try this" iterations

Project Context: 90,000 tokens (45%)  # Too much loaded
  - Entire src/ directory (unnecessary)
  - node_modules types (never needed)
  - 10 documentation files (only need 2)

Available for Response: 23,000 tokens (11.5%)  # TOO LOW!
  - Can't generate quality code
  - Forced to truncate responses
  - Poor reasoning quality

# GOOD: Optimized Context
Session History: 15,000 tokens (7.5%)  # Cleared regularly
  - Only last 5-7 relevant messages
  - Current task discussion
  - Key decisions documented

Project Context: 25,000 tokens (12.5%)  # Targeted loading
  - 3-4 files for current task
  - CLAUDE.md (always)
  - Specific memory files (on-demand)

Available for Response: 155,000 tokens (77.5%)  # OPTIMAL!
  - High-quality code generation
  - Deep reasoning capacity
  - Complex refactoring support
```

---

## Pattern 2: Aggressive Context Clearing

### Overview

The `/clear` command should become muscle memory, executed every 1-3 messages to maintain output quality.

### When to Clear Context

```typescript
// Decision Tree for /clear Usage

interface ContextClearingStrategy {
  trigger: string;
  frequency: string;
  action: string;
}

const clearingStrategies: ContextClearingStrategy[] = [
  {
    trigger: "Task completed",
    frequency: "Every task",
    action: "/clear immediately after success"
  },
  {
    trigger: "Context > 80%",
    frequency: "Automatic",
    action: "/clear + document key decisions in memory file"
  },
  {
    trigger: "Debugging session",
    frequency: "Every 3 attempts",
    action: "/clear stale error logs, keep only current"
  },
  {
    trigger: "Switching tasks",
    frequency: "Every switch",
    action: "/clear + update session-summary.md"
  },
  {
    trigger: "Poor output quality",
    frequency: "Immediate",
    action: "/clear + re-state requirements concisely"
  }
];
```

### Clearing Workflow Pattern

```bash
#!/bin/bash
# Example: Task completion workflow with clearing

# Step 1: Complete current task
implement_feature() {
    echo "Implementing authentication..."
    # ... work done ...
    echo "✓ Authentication implemented"
}

# Step 2: Document key decisions BEFORE clearing
document_decision() {
    cat >> .moai/memory/auth-decisions.md <<EOF
## Authentication Implementation ($(date +%Y-%m-%d))

**Approach**: JWT with httpOnly cookies
**Rationale**: Prevents XSS attacks, CSRF protection via SameSite
**Key Files**: src/auth/jwt.ts, src/middleware/auth-check.ts
EOF
}

# Step 3: Clear context
# /clear

# Step 4: Start fresh with next task
start_next_task() {
    echo "Context cleared. Starting API rate limiting..."
}

# Workflow
implement_feature
document_decision
# User manually executes: /clear
# start_next_task
```

### What to Clear vs Keep

```yaml
# ✅ ALWAYS CLEAR (After Documenting)
Clear:
  - Exploratory debugging sessions
  - "Try this" iteration history
  - Stale error logs
  - Completed task discussions
  - Old file diffs
  - Abandoned approaches

# 📝 DOCUMENT BEFORE CLEARING
Document First:
  - Key architectural decisions
  - Non-obvious implementation choices
  - Failed approaches (why they failed)
  - Performance insights
  - Security considerations

# ❌ NEVER CLEAR (Part of Project Context)
Keep:
  - CLAUDE.md (project guidelines)
  - Active memory files
  - Current task requirements
  - Ongoing conversation
  - Recent (< 3 messages) exchanges
```

---

## Pattern 3: Memory File Management

### Overview

Memory files are read at session start, consuming context tokens. Keep them lean and focused.

### Memory File Structure (Best Practices)

```
.moai/memory/
├── session-summary.md          # < 300 lines (current session state)
├── architectural-decisions.md   # < 400 lines (ADRs)
├── api-contracts.md            # < 200 lines (interface specs)
├── known-issues.md             # < 150 lines (blockers, workarounds)
└── team-conventions.md         # < 200 lines (code style, patterns)

Total Memory Budget: < 1,250 lines (~25K tokens)
```

### Memory File Template

```markdown
<!-- .moai/memory/session-summary.md -->
# Session Summary

**Last Updated**: 2025-01-12 14:30
**Current Sprint**: Feature/Auth-Refactor
**Active Tasks**: 2 in progress, 3 pending

## Current State

### ✅ Completed This Session
1. JWT authentication implementation (commit: abc123)
2. Password hashing with bcrypt (commit: def456)

### 🔄 In Progress
1. OAuth2 integration (70% complete)
   - Provider setup done
   - Callback handler in progress
   - Files: src/auth/oauth.ts

### 📋 Pending
1. Rate limiting middleware
2. Session management
3. CSRF protection

## Key Decisions

**Auth Strategy**: JWT in httpOnly cookies (XSS prevention)
**Password Min Length**: 12 chars (OWASP 2025 recommendation)

## Blockers

None currently.

## Next Actions

1. Complete OAuth callback handler
2. Add tests for OAuth flow
3. Document OAuth setup in README
```

### Memory File Anti-Patterns

```markdown
<!-- ❌ BAD: Bloated Memory File (1,200 lines) -->
# Session Summary

## Completed Tasks (Last 3 Weeks)
<!-- 800 lines of old task history -->
<!-- This is what git commit history is for! -->

## All Code Snippets Ever Written
```javascript
// 400 lines of full code snippets
// Should be in git, not memory files
```

<!-- ✅ GOOD: Lean Memory File (180 lines) -->
# Session Summary

**Last Updated**: 2025-01-12 14:30

## Active Work (This Session)
- OAuth integration: 70% (src/auth/oauth.ts)
- Blocker: None

## Key Decisions (Last 7 Days)
1. Auth: JWT in httpOnly cookies (XSS prevention)
2. Hashing: bcrypt, cost factor 12

## Next Actions
1. Complete OAuth callback
2. Add OAuth tests
3. Update README

<!-- Archive older content to .moai/memory/archive/ -->
```

### Memory File Rotation Strategy

```bash
#!/bin/bash
# Rotate memory files when they exceed limits

rotate_memory_file() {
    local file="$1"
    local max_lines=500
    local current_lines=$(wc -l < "$file")
    
    if [[ $current_lines -gt $max_lines ]]; then
        echo "Rotating $file ($current_lines lines > $max_lines limit)"
        
        # Archive old content
        local timestamp=$(date +%Y%m%d)
        local archive_dir=".moai/memory/archive"
        mkdir -p "$archive_dir"
        
        # Keep only recent content (last 300 lines)
        tail -n 300 "$file" > "${file}.tmp"
        
        # Archive full file
        mv "$file" "${archive_dir}/$(basename "$file" .md)-${timestamp}.md"
        
        # Replace with trimmed version
        mv "${file}.tmp" "$file"
        
        echo "✓ Archived to ${archive_dir}/"
    fi
}

# Check all memory files
for file in .moai/memory/*.md; do
    rotate_memory_file "$file"
done
```

---

## Pattern 4: MCP Server Optimization

### Overview

Each enabled MCP server adds tool definitions to system prompt, consuming context tokens. Disable unused servers.

### MCP Context Impact

```json
// .claude/mcp.json - Context-optimized configuration

{
  "mcpServers": {
    // ✅ ENABLED: Active development tools
    "context7": {
      "command": "npx",
      "args": ["-y", "@context7/mcp"],
      "env": {
        "CONTEXT7_API_KEY": "your-key"
      }
    },
    
    // ❌ DISABLED: Not needed for current project
    // "playwright": {
    //   "command": "npx",
    //   "args": ["-y", "@playwright/mcp"]
    // },
    
    // ✅ ENABLED: Documentation research
    "sequential-thinking": {
      "command": "npx",
      "args": ["-y", "@sequential-thinking/mcp"]
    }
    
    // ❌ DISABLED: Slackbot not in use
    // "slack": {
    //   "command": "npx",
    //   "args": ["-y", "@slack/mcp"]
    // }
  }
}
```

### Measuring MCP Overhead

```bash
# Monitor MCP context usage
/context

# Example output:
# MCP Servers (3 enabled):
# - context7: 847 tokens (tool definitions)
# - sequential-thinking: 412 tokens
# - playwright: 1,234 tokens (DISABLED to save tokens)
#
# Total MCP Overhead: 1,259 tokens
```

### MCP Usage Strategy

```python
class MCPUsageStrategy:
    """Strategic MCP server management for context optimization"""
    
    STRATEGIES = {
        "documentation_heavy": {
            "enable": ["context7"],
            "disable": ["playwright", "slack", "github"],
            "rationale": "Research phase, need API docs access"
        },
        "testing_phase": {
            "enable": ["playwright", "sequential-thinking"],
            "disable": ["context7", "slack"],
            "rationale": "E2E testing, browser automation needed"
        },
        "code_review": {
            "enable": ["github", "sequential-thinking"],
            "disable": ["context7", "playwright", "slack"],
            "rationale": "PR review, need GitHub API access"
        },
        "minimal": {
            "enable": [],
            "disable": ["*"],
            "rationale": "Maximum context availability, no external tools"
        }
    }
    
    @staticmethod
    def optimize_for_phase(phase: str):
        """
        Reconfigure .claude/mcp.json for current development phase
        """
        strategy = MCPUsageStrategy.STRATEGIES.get(phase, "minimal")
        print(f"Optimizing MCP servers for: {phase}")
        print(f"Enable: {strategy['enable']}")
        print(f"Disable: {strategy['disable']}")
        print(f"Rationale: {strategy['rationale']}")
        # Update .claude/mcp.json accordingly
```

---

## Pattern 5: Strategic Chunking

### Overview

Break large tasks into smaller pieces completable within optimal context bounds (< 80% usage).

### Task Chunking Strategy

```typescript
// Task size estimation and chunking

interface Task {
  name: string;
  estimatedTokens: number;
  dependencies: string[];
}

const chunkTask = (largeTask: Task): Task[] => {
  const MAX_CHUNK_TOKENS = 120_000; // 60% of 200K context
  
  if (largeTask.estimatedTokens <= MAX_CHUNK_TOKENS) {
    return [largeTask]; // No chunking needed
  }
  
  // Example: Authentication system (estimated 250K tokens)
  const chunks: Task[] = [
    {
      name: "Chunk 1: User model & password hashing",
      estimatedTokens: 80_000,
      dependencies: []
    },
    {
      name: "Chunk 2: JWT generation & validation",
      estimatedTokens: 70_000,
      dependencies: ["Chunk 1"]
    },
    {
      name: "Chunk 3: Login/logout endpoints",
      estimatedTokens: 60_000,
      dependencies: ["Chunk 2"]
    },
    {
      name: "Chunk 4: Session middleware & guards",
      estimatedTokens: 40_000,
      dependencies: ["Chunk 3"]
    }
  ];
  
  return chunks;
};

// Workflow:
// 1. Complete Chunk 1
// 2. /clear
// 3. Document Chunk 1 results in memory file
// 4. Start Chunk 2 with minimal context
```

### Chunking Anti-Patterns

```yaml
# ❌ BAD: Mixing Unrelated Tasks
Chunk 1 (200K tokens - OVERLOADED):
  - User authentication
  - Payment processing
  - Email notifications
  - Admin dashboard
  - Analytics integration
# Result: Poor quality on ALL tasks, context overflow

# ✅ GOOD: Focused Chunks
Chunk 1 (60K tokens):
  - User authentication only
  - Complete, test, document
  
Chunk 2 (70K tokens):
  - Payment processing only
  - Builds on auth from Chunk 1
  
Chunk 3 (50K tokens):
  - Email notifications
  - Uses auth + payment data
```

---

## Pattern 6: Quality Over Quantity Context

### Overview

10% context with highly relevant information produces better results than 90% filled with noise.

### Context Quality Checklist

```markdown
## Before Adding to Context

Ask yourself:

1. **Relevance**: Does this directly support current task?
   - ✅ YES: Load file
   - ❌ NO: Skip or summarize

2. **Freshness**: Is this information current?
   - ✅ Current: Keep in context
   - ❌ Stale (>1 hour): Archive or delete

3. **Actionability**: Will Claude use this to generate code?
   - ✅ Actionable: Include
   - ❌ FYI only: Document in memory file, remove from context

4. **Uniqueness**: Is this duplicated elsewhere?
   - ✅ Unique: Keep
   - ❌ Duplicate: Remove duplicates, keep one canonical source

## High-Quality Context Example (30K tokens, 15%)

Context Contents:
1. CLAUDE.md (2K tokens) - Always loaded
2. src/auth/jwt.ts (5K tokens) - Current file being edited
3. src/types/auth.ts (3K tokens) - Type definitions needed
4. .moai/memory/session-summary.md (4K tokens) - Current session state
5. tests/auth.test.ts (8K tokens) - Test file for reference
6. Last 5 messages (8K tokens) - Recent discussion

Total: 30K tokens
Quality: HIGH - Every token is directly relevant to current task

## Low-Quality Context Example (170K tokens, 85%)

Context Contents:
1. CLAUDE.md (2K tokens)
2. Entire src/ directory (80K tokens) - ❌ 90% irrelevant
3. node_modules/ types (40K tokens) - ❌ Never needed
4. 50 previous messages (30K tokens) - ❌ Stale debugging sessions
5. 10 documentation files (18K tokens) - ❌ Only need 1-2

Total: 170K tokens
Quality: LOW - <10% of tokens are actually useful
Result: Poor code generation, missed context, truncated responses
```

---

## Best Practices Checklist

**Context Allocation**:
- [ ] Context usage maintained below 80%
- [ ] System prompt < 2K tokens
- [ ] MCP tools < 5K tokens total
- [ ] Session history < 30K tokens
- [ ] Project context < 40K tokens
- [ ] Available response capacity > 100K tokens

**Aggressive Clearing**:
- [ ] `/clear` executed every 1-3 messages
- [ ] Context cleared after each task completion
- [ ] Key decisions documented before clearing
- [ ] Stale error logs removed immediately
- [ ] Exploratory sessions cleared regularly

**Memory File Management**:
- [ ] Each memory file < 500 lines
- [ ] Total memory files < 1,250 lines
- [ ] session-summary.md updated before task switches
- [ ] Old content archived to .moai/memory/archive/
- [ ] No raw code stored in memory (summarize instead)

**MCP Optimization**:
- [ ] Unused MCP servers disabled
- [ ] `/context` checked regularly
- [ ] MCP overhead < 5K tokens
- [ ] Servers enabled/disabled per development phase

**Strategic Chunking**:
- [ ] Large tasks split into < 120K token chunks
- [ ] Related work grouped in same chunk
- [ ] Chunk dependencies documented
- [ ] `/clear` between chunks
- [ ] Previous chunk results in memory file

**Quality Over Quantity**:
- [ ] Only load files needed for current task
- [ ] Remove stale information (>1 hour old)
- [ ] Eliminate duplicate context
- [ ] Summarize instead of including full files
- [ ] Verify every loaded item is actionable

---

## Common Pitfalls to Avoid

### Pitfall 1: Loading Entire Codebase

```bash
# ❌ BAD
# User: "Help me understand this project"
# Claude loads all 200 files in src/

# ✅ GOOD
# User: "Help me understand the authentication flow"
# Claude loads only:
# - src/auth/jwt.ts
# - src/middleware/auth-check.ts
# - tests/auth.test.ts
```

### Pitfall 2: Never Clearing Context

```yaml
# ❌ BAD: 3-Hour Session Without Clearing
Context: 195K / 200K tokens (97.5%)
  - 80 messages of trial-and-error debugging
  - 15 failed approaches still in context
  - Stale error logs from 2 hours ago
Result: "I need to truncate my response..."

# ✅ GOOD: Clearing Every 5-10 Minutes
Context: 45K / 200K tokens (22.5%)
  - Only last 5 relevant messages
  - Current task files
  - Fresh, high-quality context
Result: Complete, high-quality responses
```

### Pitfall 3: Bloated Memory Files

```markdown
<!-- ❌ BAD: 2,000-line session-summary.md -->
- Takes 40K tokens just to load
- 90% is outdated information
- Prevents loading actual source files

<!-- ✅ GOOD: 250-line session-summary.md -->
- Takes 5K tokens to load
- 100% current and relevant
- Leaves room for source files
```

---

## Tool Versions (2025)

| Tool | Version | Purpose |
|------|---------|---------|
| Claude Code | 1.5.0+ | CLI interface |
| Claude Sonnet | 4.5+ | Model (200K context) |
| Context7 MCP | Latest | Documentation research |
| Sequential Thinking MCP | Latest | Problem solving |

---

## References

- [Claude Code Context Management](https://docs.claude.com/en/docs/build-with-claude/context-windows) - Official documentation
- [Claude Code Best Practices](https://www.shuttle.dev/blog/2025/10/16/claude-code-best-practices) - Community guide
- [Context Window Optimization](https://sparkco.ai/blog/mastering-claudes-context-window-a-2025-deep-dive) - 2025 deep dive
- [Memory Management Strategies](https://cuong.io/blog/2025/06/15-claude-code-best-practices-memory-management) - Advanced patterns

---

## Changelog

- **v4.0.0** (2025-01-12): Enterprise upgrade with 2025 best practices, aggressive clearing patterns, MCP optimization, strategic chunking
- **v1.0.0** (2025-03-29): Initial release

---

## Works Well With

- `moai-alfred-practices` - Development best practices
- `moai-alfred-session-state` - Session management
- `moai-cc-memory` - Memory file patterns
- `moai-alfred-workflow` - 4-step workflow optimization
