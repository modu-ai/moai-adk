---
name: moai-foundation-memory
aliases: [moai-foundation-memory, memory-mcp]
description: Persistent memory across sessions using MCP Memory Server for user preferences, project context, and learned patterns
version: 1.0.0
modularized: false
user-invocable: false
category: foundation
tags:
  [
    "foundation",
    "memory",
    "persistence",
    "mcp",
    "user-preferences",
    "project-context",
    "session-state",
  ]
updated: 2026-01-26
status: "active"
allowed-tools:
  - Read
  - Grep
  - Glob
  - mcp__memory__*

# Progressive Disclosure Configuration
progressive_disclosure:
  enabled: true
  level1_tokens: ~100
  level2_tokens: ~3000

# Trigger Conditions for Level 2 Loading
triggers:
  keywords:
    - "memory"
    - "remember"
    - "preference"
    - "persist"
    - "save"
    - "recall"
    - "learned"
    - "user settings"
    - "project context"
    - "session history"
  agents:
    - "manager-spec"
    - "manager-project"
    - "manager-strategy"
  phases:
    - "plan"
    - "run"
---

## Quick Reference

Persistent Memory Management - MCP Memory Server integration for maintaining context across Claude Code sessions, storing user preferences, project-specific knowledge, and learned patterns.

Core Capabilities:

- Persistent key-value storage across sessions
- User preference management
- Project context preservation
- Learned pattern storage
- Session history tracking

When to Use:

- Store user preferences (language, coding style, naming conventions)
- Preserve project-specific decisions and rationale
- Remember frequently used commands and patterns
- Track project milestones and progress
- Store learned code patterns for reuse

Key Operations:

- `mcp__memory__store`: Store a key-value pair
- `mcp__memory__retrieve`: Retrieve a stored value
- `mcp__memory__list`: List all stored keys
- `mcp__memory__delete`: Delete a stored key

---

## Implementation Guide

### MCP Memory Server Setup

The memory server is configured in `.mcp.json`:

```json
{
  "memory": {
    "command": "${SHELL:-/bin/bash}",
    "args": ["-l", "-c", "exec npx -y @modelcontextprotocol/server-memory"]
  }
}
```

### Memory Categories

Organize stored data by category prefixes:

**User Preferences** (prefix: `user_`):
- `user_language`: Conversation language preference
- `user_coding_style`: Preferred coding conventions
- `user_naming_convention`: Variable/function naming style
- `user_timezone`: User's timezone for scheduling

**Project Context** (prefix: `project_`):
- `project_tech_stack`: Technologies used in project
- `project_architecture`: Architecture decisions
- `project_conventions`: Project-specific conventions
- `project_dependencies`: Key dependencies and versions

**Learned Patterns** (prefix: `pattern_`):
- `pattern_error_fixes`: Common error resolution patterns
- `pattern_code_templates`: Frequently used code templates
- `pattern_workflow`: User's preferred workflow

**Session State** (prefix: `session_`):
- `session_last_spec`: Last worked SPEC ID
- `session_active_branch`: Current git branch
- `session_pending_tasks`: Incomplete tasks

### Usage Patterns

**Pattern 1: Store User Preference**

When user explicitly states a preference:

```
User: "I prefer Korean responses"
Action: Store using mcp__memory__store
Key: "user_language"
Value: "ko"
```

**Pattern 2: Retrieve Context on Session Start**

At session initialization:

1. Retrieve `user_language` for response language
2. Retrieve `project_tech_stack` for context
3. Retrieve `session_last_spec` for continuity

**Pattern 3: Learn from User Behavior**

When user corrects or adjusts output:

```
User: "Use camelCase not snake_case"
Action: Store pattern
Key: "user_naming_convention"
Value: "camelCase"
```

**Pattern 4: Project Knowledge Base**

Store important project decisions:

```
Key: "project_auth_decision"
Value: "JWT with refresh tokens, stored in httpOnly cookies"
```

### Best Practices

**Storage Guidelines:**

- Use descriptive, categorized key names
- Keep values concise (under 1000 characters)
- Store JSON for complex data structures
- Include timestamps for time-sensitive data

**Retrieval Guidelines:**

- Check memory on session start
- Retrieve relevant context before tasks
- Use memory to avoid repeated questions

**Privacy Considerations:**

- Never store sensitive credentials
- Avoid storing personal identifiable information
- Store preferences, not personal data

### Integration with Alfred

Alfred should proactively use memory:

**On Session Start:**
1. Retrieve user preferences
2. Apply language and style settings
3. Load project context

**During Interaction:**
1. Store explicit user preferences
2. Learn from corrections
3. Update project context as needed

**On Task Completion:**
1. Store successful patterns
2. Update session state
3. Record milestones

---

## Memory Key Reference

### User Preferences

| Key | Type | Description |
|-----|------|-------------|
| `user_language` | string | Response language (ko, en, ja, etc.) |
| `user_coding_style` | string | Preferred style (descriptive, concise) |
| `user_naming_convention` | string | Naming style (camelCase, snake_case) |
| `user_comment_language` | string | Code comment language |
| `user_timezone` | string | User timezone |
| `user_expertise_level` | string | junior, mid, senior |

### Project Context

| Key | Type | Description |
|-----|------|-------------|
| `project_name` | string | Project name |
| `project_tech_stack` | JSON | Technologies and frameworks |
| `project_architecture` | string | Architecture pattern (monolith, microservices) |
| `project_test_framework` | string | Testing framework (pytest, jest) |
| `project_conventions` | JSON | Project-specific conventions |

### Learned Patterns

| Key | Type | Description |
|-----|------|-------------|
| `pattern_preferred_libraries` | JSON | User's preferred libraries |
| `pattern_error_resolutions` | JSON | Common error fixes |
| `pattern_code_templates` | JSON | Frequently used templates |

### Session State

| Key | Type | Description |
|-----|------|-------------|
| `session_last_spec` | string | Last worked SPEC ID |
| `session_active_branch` | string | Current git branch |
| `session_pending_tasks` | JSON | Incomplete tasks |
| `session_last_activity` | string | Timestamp of last activity |

---

## Agent-to-Agent Context Sharing

### Overview

Memory MCP enables agents to share context during workflow execution. This reduces token overhead and ensures consistency across the Plan-Run-Sync cycle.

### Handoff Key Schema

**Handoff Data** (prefix: `handoff_`):
```
handoff_{from_agent}_{to_agent}_{spec_id}
```

Example: `handoff_manager-spec_manager-ddd_SPEC-001`

**Shared Context** (prefix: `context_`):
```
context_{spec_id}_{category}
```

Categories: `requirements`, `architecture`, `api`, `database`, `decisions`

### Workflow Integration

**Plan Phase (manager-spec):**

At SPEC completion, store:
```
Key: context_SPEC-001_requirements
Value: {
  "summary": "User authentication with JWT",
  "acceptance_criteria": ["AC1", "AC2", "AC3"],
  "tech_decisions": ["JWT", "Redis sessions"],
  "constraints": ["No external auth providers"]
}
```

**Run Phase (manager-ddd, expert-backend, expert-frontend):**

On task start, retrieve:
```
Key: context_SPEC-001_requirements
Action: Load requirements summary
```

On architecture decision, store:
```
Key: context_SPEC-001_architecture
Value: {
  "pattern": "Clean Architecture",
  "layers": ["domain", "application", "infrastructure"],
  "api_style": "REST with OpenAPI 3.0"
}
```

**Sync Phase (manager-docs):**

Retrieve all context for documentation:
```
Keys: context_SPEC-001_*
Action: Generate comprehensive documentation
```

### Handoff Protocol

**Step 1: Store handoff before agent completion**
```
Key: handoff_manager-spec_manager-ddd_SPEC-001
Value: {
  "spec_id": "SPEC-001",
  "status": "approved",
  "key_requirements": [...],
  "tech_stack": [...],
  "priority_order": [...],
  "estimated_complexity": "medium"
}
```

**Step 2: Retrieve handoff on agent start**
```
Key: handoff_manager-spec_manager-ddd_SPEC-001
Action: Load context and continue workflow
```

**Step 3: Update progress**
```
Key: context_SPEC-001_progress
Value: {
  "completed_tasks": ["API design", "Database schema"],
  "in_progress": ["Authentication implementation"],
  "blocked": [],
  "completion_percentage": 60
}
```

### Context Categories

| Category | Purpose | Stored By | Used By |
|----------|---------|-----------|---------|
| `requirements` | SPEC requirements | manager-spec | All agents |
| `architecture` | Architecture decisions | manager-strategy | expert-* |
| `api` | API contracts | expert-backend | expert-frontend |
| `database` | Schema decisions | expert-backend | All agents |
| `decisions` | Key decisions log | All agents | manager-docs |
| `progress` | Workflow progress | All agents | Alfred |

### Best Practices for Agent Sharing

**Store Strategically:**
- Store at workflow boundaries (phase completion)
- Store when making important decisions
- Store when context exceeds prompt capacity

**Retrieve Efficiently:**
- Retrieve at agent start
- Retrieve when context is needed
- Cache retrieved values in prompt context

**Keep Values Structured:**
- Use JSON for complex data
- Include timestamps for tracking
- Keep values under 2000 characters

### Example: Full Workflow

```
1. manager-spec completes SPEC-001
   └─ Store: context_SPEC-001_requirements
   └─ Store: handoff_manager-spec_manager-ddd_SPEC-001

2. manager-ddd starts
   └─ Retrieve: handoff_manager-spec_manager-ddd_SPEC-001
   └─ Retrieve: context_SPEC-001_requirements

3. expert-backend implements API
   └─ Retrieve: context_SPEC-001_requirements
   └─ Store: context_SPEC-001_api
   └─ Store: context_SPEC-001_database

4. expert-frontend implements UI
   └─ Retrieve: context_SPEC-001_api
   └─ Store: context_SPEC-001_frontend

5. manager-docs generates documentation
   └─ Retrieve: context_SPEC-001_* (all)
   └─ Generate comprehensive docs
```

---

## Works Well With

- moai-foundation-context - Token budget and session management
- moai-foundation-core - SPEC-First workflow integration
- moai-workflow-project - Project configuration persistence
- moai-foundation-claude - Claude Code patterns

---

## Success Metrics

- Preference Persistence: User preferences maintained across sessions
- Context Continuity: Project context available without re-explanation
- Learning Efficiency: Reduced repetitive questions over time
- Session Recovery: Quick resumption with session state

---

Status: Production Ready
MCP Integration: @modelcontextprotocol/server-memory
Generated with: MoAI-ADK Skill Factory
