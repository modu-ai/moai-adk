---
name: moai-core-agent-guide
version: 4.0.0
created: 2025-11-18
updated: '2025-11-18'
status: stable
description: Complete guide to MoAI-ADK agents, delegation patterns, and agent orchestration
  for intelligent task distribution
allowed-tools:
- Read
- Bash
- WebFetch
stability: stable
---


# MoAI Agent Architecture & Delegation Guide - 

**Master guide for agent selection, delegation patterns, and orchestration strategies**

> **Scope**: MoAI-ADK Agent System
> **Framework**: Task delegation, parallel execution, context management
> **Keywords**: agents, delegation, orchestration, task-distribution, specialist-expertise

## Level 1: Quick Reference

### Agent Hierarchy

**Tier 1: Specialized Agents** (domain expertise)
- spec-builder: SPEC-First requirements
- tdd-implementer: TDD implementation
- backend-expert: API & services
- frontend-expert: UI components
- security-expert: Security patterns

**Tier 2: Support Agents** (task-specific)
- docs-manager: Documentation
- quality-gate: Quality validation
- performance-engineer: Optimization

**Tier 3: Skill-based Agents** (knowledge capsules)
- Delegated to moai-* Skills via Skill() calls

---

## Level 2: Agent Delegation Patterns

### When to Use Agent Delegation

Use delegation when:
- Task requires specialized expertise
- Multiple domains involved
- Token budget optimization needed
- Parallel execution beneficial

### Sequential vs Parallel

**Sequential**: Steps depend on previous results
**Parallel**: Independent tasks run simultaneously (3-5x faster)

---

## Level 3: Advanced Orchestration

### Multi-Agent Workflows

- Agent chaining and composition
- Context passing between agents
- Session management and persistence
- Error handling and fallback strategies

---

## References

- **SPEC-First Methodology**: See CLAUDE.md → Agent Delegation
- **Token Efficiency**: See .moai/memory/token-efficiency.md
- **Advanced Patterns**: See .moai/memory/agent-delegation.md

---

**Last Updated**: 2025-11-18
**Format**: Markdown | **Language**: English
**Status**: Stable
**Version**: 4.0.0
