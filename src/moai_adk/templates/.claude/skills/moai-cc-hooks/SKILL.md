---
name: moai-cc-hooks
version: 2.0.0
created: 2025-10-22
updated: 2025-11-11
status: active
description: Claude Code Hooks system, event-driven automation, and workflow orchestration. Use when implementing custom hooks, automating workflows, or creating event-based triggers.
keywords: ['hooks', 'events', 'automation', 'workflow', 'triggers']
allowed-tools:
  - Read
  - Bash
  - Glob
---

# Claude Code Hooks System

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-cc-hooks |
| **Version** | 2.0.0 (2025-11-11) |
| **Allowed tools** | Read, Bash, Glob |
| **Auto-load** | On demand when hook patterns detected |
| **Tier** | Claude Code (Core) |

---

## What It Does

Claude Code Hooks system, event-driven automation, and workflow orchestration.

**Key capabilities**:
- ✅ Event-driven automation
- ✅ Custom hook creation
- ✅ Workflow orchestration
- ✅ Context seeding
- ✅ Just-in-time validation

---

## When to Use

- ✅ Implementing custom hooks
- ✅ Automating repetitive workflows
- ✅ Creating event-based triggers
- ✅ Enhancing session management

---

## Core Hook Patterns

### Hook Architecture
1. **Session Events**: Session start/end hooks
2. **Tool Use Events**: Pre/post tool execution
3. **File Events**: File creation/modification triggers
4. **Error Events**: Error handling and recovery
5. **Custom Events**: Application-specific triggers

### Hook Types
- **Guardrail Hooks**: Prevent destructive actions
- **Enhancement Hooks**: Add context and suggestions
- **Automation Hooks**: Trigger automated workflows
- **Validation Hooks**: Quality and compliance checks
- **Notification Hooks**: Alert and report generation

---

## Dependencies

- Claude Code hooks system
- Event handling framework
- Workflow automation tools
- Context management system

---

## Works Well With

- `moai-cc-configuration` (Hook configuration)
- `moai-cc-automation` (Workflow automation)
- `moai-foundation-tags` (Tag validation)

---

## Changelog

- **v2.0.0** (2025-11-11): Added complete metadata, hook architecture patterns
- **v1.0.0** (2025-10-22): Initial hooks system

---

**End of Skill** | Updated 2025-11-11
