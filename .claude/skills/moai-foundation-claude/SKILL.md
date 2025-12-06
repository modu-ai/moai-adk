---
name: moai-foundation-claude
aliases: [moai-foundation-claude]
category: foundation
description: Canonical Claude Code skill authoring kit covering agent Skills, sub-agent templates, custom slash commands, orchestration patterns, hooks, memory, settings, and IAM permission rules aligned with official documentation.
version: 2.0.0
modularized: false
allowed-tools: Read, Write, Edit, Grep, Glob
tags:
 - foundation
 - claude-code
 - skills
 - sub-agents
 - slash-commands
 - orchestration
 - hooks
 - memory
 - settings
 - iam
 - best-practices
# Claude Code Authoring Kit

Comprehensive reference for Claude Code Skills, sub-agents, custom slash commands, hooks, memory, settings, and IAM permissions with official standards compliance.

Official Documentation References:

- [Skills Guide](reference/claude-code-skills-official.md) - Agent Skills creation and management
- [Sub-agents Guide](reference/claude-code-sub-agents-official.md) - Sub-agent development and delegation
- [Custom Slash Commands](reference/claude-code-custom-slash-commands-official.md) - Command creation and orchestration
- [Hooks System](reference/claude-code-hooks-official.md) - Event-driven automation
- [Memory Management](reference/claude-code-memory-official.md) - Context and knowledge persistence
- [Settings Configuration](reference/claude-code-settings-official.md) - Configuration hierarchy and management
- [IAM & Permissions](reference/claude-code-iam-official.md) - Access control and security
- [Complete Configuration](reference/complete-configuration-guide.md) - Comprehensive setup guide

## Quick Reference (30 seconds)

Skills: `~/.claude/skills/` (personal) or `.claude/skills/` (project), ≤500 lines, progressive disclosure
Sub-agents: `Task(subagent_type="...")` delegation only, no nested spawning
Commands: `$ARGUMENTS`/`$1`/`$2` parameters, `@file` references, stored in `.claude/commands/`
Hooks: Event-driven automation in `settings.json`, PreToolUse/PostToolUse events
Memory: Enterprise → Project → User hierarchy, `@import.md` syntax
Settings: 4-tier hierarchy (Enterprise → User → Project → Local)
IAM: Tiered permissions (Read→Bash→Write→Admin), tool-specific rules

## Implementation Guide (5 minutes)

### Features

- Comprehensive Claude Code authoring reference
- Skills, sub-agents, commands, hooks, and settings
- IAM permissions and security best practices
- Progressive disclosure documentation architecture
- MCP integration patterns and examples

### When to Use

- Creating new Claude Code skills following official standards
- Developing custom sub-agents with proper delegation patterns
- Implementing custom slash commands with parameter handling
- Setting up hooks for event-driven automation
- Configuring IAM permissions and access control

### Core Patterns

Pattern 1: Skill Creation (≤500 lines)
```yaml
---
name: skill-name
description: One-line description (max 200 chars)
version: 1.0.0
updated: 2025-11-26
status: active
---

## Quick Reference (30 seconds)
## Implementation Guide (5 minutes)
## Advanced Implementation (10+ minutes)
```

Pattern 2: Sub-agent Delegation
```python
# Sequential delegation
result1 = Task(subagent_type="workflow-spec", prompt="Analyze")
result2 = Task(subagent_type="code-backend", prompt="Implement", context=result1)

# Parallel delegation
results = await Promise.all([
 Task(subagent_type="code-backend", prompt="Backend"),
 Task(subagent_type="code-frontend", prompt="Frontend")
])
```

Pattern 3: Custom Command with Hooks
```json
{
 "hooks": {
 "PreToolUse": [{
 "matcher": "Write",
 "hooks": [{"type": "command", "command": "validate-write"}]
 }]
 }
}
```

## MoAI-ADK Skills & Sub-agents Directory

### Quick Access Patterns

Core Skills: `Skill("moai-foundation-claude")` - This comprehensive authoring kit
Agent Creation: `Task(subagent_type="agent-factory")` - Create standardized sub-agents
Skill Creation: `Task(subagent_type="skill-factory")` - Create compliant skills
Quality Gates: `Task(subagent_type="quality-gate")` - TRUST 5 validation
Documentation: `Task(subagent_type="docs-manager")` - Generate technical docs

Key Specialized Skills:
- `moai-lang-python` - Python 3.13+ with FastAPI, async patterns
- `moai-lang-typescript` - TypeScript 5.9+ with React 19, Next.js 16
- `moai-domain-backend` - Modern backend architecture patterns
- `moai-domain-frontend` - React 19, Next.js 15, Vue 3.5 frameworks
- `moai-quality-security` - OWASP Top 10, threat modeling
- `moai-essentials-debug` - AI-powered debugging with Context7

Essential Sub-agents:
- `spec-builder` - EARS format specification generation
- `tdd-implementer` - RED-GREEN-REFACTOR TDD execution
- `security-expert` - Security analysis and validation
- `backend-expert` - Backend architecture and API development
- `frontend-expert` - Frontend UI implementation
- `performance-engineer` - Performance optimization and analysis

## Essential Implementation Patterns

### Command→Agent→Skill Orchestration

Sequential Workflow:
```python
# Phase 1: Analysis → spec-builder → analysis
analysis = Task(subagent_type="spec-builder", prompt="Analyze: $ARGUMENTS")
# Phase 2: Implementation → tdd-implementer → code + tests
implementation = Task(subagent_type="tdd-implementer", prompt="Implement: $analysis.spec_id")
# Phase 3: Validation → quality-gate → approval
validation = Task(subagent_type="quality-gate", prompt="Validate: $implementation")
```

Parallel Development:
```python
# Execute independent tasks simultaneously
results = await Promise.all([
 Task(subagent_type="backend-expert", prompt="Backend: $1"),
 Task(subagent_type="frontend-expert", prompt="Frontend: $1"),
 Task(subagent_type="docs-manager", prompt="Docs: $1")
])
# Integrate all results
Task(subagent_type="quality-gate", prompt="Integrate", context={"results": results})
```

### Token Session Management

Independent 200K Sessions: Each `Task()` creates a new 200K token context
```python
Task(subagent_type="backend-expert", prompt="Complex task") # 200K session
Task(subagent_type="frontend-expert", prompt="UI task") # New 200K session
```

### File Reference Patterns

Parameter Handling:
- Positional: `$1`, `$2`, `$3`
- All arguments: `$ARGUMENTS`
- File references: `@config.yaml`, `@path/to/file.md`

### Hook Integration

Pre/Post Tool Execution (see [Hooks Guide](reference/claude-code-hooks-official.md)):
```json
{
 "hooks": {
 "PreToolUse": [{"matcher": "Bash", "hooks": [{"type": "command", "command": "validate-command"}]}],
 "PostToolUse": [{"matcher": "Write", "hooks": [{"type": "command", "command": "backup-file"}]}]
 }
}
```

## Key Reference Guides

Comprehensive Documentation:
- [Commands Guide](reference/claude-code-custom-slash-commands-official.md) - Complete command creation
- [Hooks System](reference/claude-code-hooks-official.md) - Event-driven automation
- [Memory Management](reference/claude-code-memory-official.md) - Context persistence
- [Settings Config](reference/claude-code-settings-official.md) - Configuration hierarchy
- [IAM Permissions](reference/claude-code-iam-official.md) - Access control
- [Complete Setup](reference/complete-configuration-guide.md) - Full configuration

## Works Well With

- moai-core-agent-factory - For creating new sub-agents with proper standards
- moai-cc-commands - For creating custom slash commands
- moai-cc-hooks - For implementing Claude Code hooks
- moai-cc-configuration - For managing Claude Code settings
- moai-quality-gate - For validating code quality standards
- moai-essentials-debug - For debugging skill loading issues
