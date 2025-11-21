---
name: moai-core-agent-guide
description: MoAI agent architecture, delegation patterns, and 19-agent orchestration guide
allowed-tools: [Read]
---

# MoAI Agent Architecture Guide

## Quick Reference

MoAI-ADK orchestrates **19 specialized agents** across 4 layers: Commands → Sub-agents → Skills (135+) → Hooks. This skill documents agent roles, delegation patterns, and model selection strategies for optimal workflow orchestration.

**Agent Roster** (November 2025):
- **Alfred**: Super orchestrator (Sonnet)
- **Core 10**: project-manager, spec-builder, code-builder pipeline, doc-syncer, tag-agent, git-manager, debug-helper, trust-checker, quality-gate, cc-manager
- **Expert 4**: backend-expert, frontend-expert, devops-expert, ui-ux-expert (keyword-triggered)
- **Zero-project 6**: language-detector, backup-merger, project-interviewer, document-generator, feature-selector, template-optimizer
- **Claude Built-in 2**: Explore (codebase analysis), general-purpose

**Model Strategy**:
- **Haiku**: Doc sync, TAG inventory, Git automation, pattern matching (fast execution)
- **Sonnet**: Planning, implementation, debugging, complex reasoning (deep analysis)

---

## Implementation Guide

### Core Agent Delegation

**Workflow Phases**:
```
/moai:0-project → project-manager (initialization)
/moai:1-plan → spec-builder (SPEC generation)
/moai:2-run → implementation-planner → tdd-implementer (RED→GREEN→REFACTOR)
/moai:3-sync → doc-syncer + tag-agent (documentation + validation)
```

**Expert Agent Auto-Trigger** (by SPEC keywords):
```
Keywords in SPEC → Auto-delegation:
- 'api', 'backend' → backend-expert
- 'ui', 'frontend', 'component' → frontend-expert
- 'docker', 'kubernetes', 'ci/cd' → devops-expert
- 'design', 'accessibility', 'figma' → ui-ux-expert
```

### Decision Tree

| Situation | Agent | Reason |
|-----------|-------|---------|
| Understand codebase | Explore | Fast analysis with Glob+Grep |
| Write SPEC | spec-builder | EARS syntax expert |
| Debug failure | debug-helper | Stack trace analysis |
| Implement feature | code-builder pipeline | TDD automation |
| Sync docs | doc-syncer | Living Documents |
| Git/PR management | git-manager | GitFlow automation |
| Verify quality | trust-checker | TRUST 5 principles |
| Release gate | quality-gate | Coverage + security |

---

## Advanced Patterns

### Agent Collaboration Principles

**1. Command Precedence**: Commands override agent guidelines
**2. Single Responsibility**: Each agent handles only its specialty
**3. Zero Overlap**: Hand off to most direct expertise
**4. Confidence Reporting**: Share confidence levels + risks
**5. Escalation Path**: Escalate to Alfred when blocked

### Model Selection Guide

```
Haiku Use Cases:
- Documentation generation (fast string processing)
- TAG inventory scanning (pattern matching)
- Git commit automation (deterministic)
- Coverage reports (rule-based)

Sonnet Use Cases:
- Architecture planning (creative synthesis)
- Bug root cause analysis (multi-step reasoning)
- SPEC authoring (EARS pattern application)
- Complex refactoring (design decisions)
```

---

## Best Practices

### ✅ DO
- Use Explore for large codebase analysis (thoroughness: quick|medium|very thorough)
- Delegate to expert agents for domain-specific tasks
- Let Alfred orchestrate multi-agent workflows
- Document model selection rationale in task notes
- Escalate to Alfred when agents are blocked

### ❌ DON'T
- Bypass Alfred for complex multi-step workflows
- Use Sonnet for simple pattern matching (cost inefficient)
- Skip expert agent auto-triggers (missing domain expertise)
- Ignore agent escalation signals (blocks progress)

---

## Works Well With

- `moai-alfred-orchestration` (Alfred's orchestration patterns)
- `moai-cc-subagent-lifecycle` (Subagent delegation)
- `moai-foundation-trust` (TRUST 5 validation)
- `moai-foundation-tags` (TAG chain management)

---

**Version**: 1.0.0  
**Last Updated**: 2025-11-21  
**Status**: Production Ready  
**Reference Source**: /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-core-agent-guide/reference.md
