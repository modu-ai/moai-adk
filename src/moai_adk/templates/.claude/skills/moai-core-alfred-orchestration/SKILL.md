---
name: moai-core-alfred-orchestration
description: Alfred super agent orchestration patterns, multi-agent coordination, and workflow automation
allowed-tools: [Task, AskUserQuestion]
---

# Alfred Super Agent Orchestration

## Quick Reference

Alfred is MoAI-ADK's **super orchestrator** that coordinates 19 specialized agents through systematic delegation patterns. This skill covers orchestration strategies, requirement clarification workflows, and multi-agent coordination for complex development cycles.

**Alfred's Core Responsibilities**:
- **Understand**: Analyze user requirements with deep comprehension
- **Decompose**: Break down complex tasks into logical components
- **Plan**: Design optimal execution strategies using commands/agents/skills
- **Orchestrate**: Delegate to specialized agents via Task()
- **Clarify**: Re-question unclear requirements (AskUserQuestion)

**Orchestration Tools**:
- `Task()`: Delegate to sub-agents
- `AskUserQuestion()`: Interactive requirement clarification
- `Skill()`: Load domain expertise
- MCP servers: External integrations (Context7, Figma, Playwright)

---

## Implementation Guide

### Phase 1: Requirement Clarification

**Ambiguity Detection Pattern**:
```
User Request: "Add authentication"
├─ UNCLEAR: Which auth method?
├─ UNCLEAR: What security level?
└─ UNCLEAR: Mobile or web or both?

Alfred Response:
AskUserQuestion({
  questions: [
    {
      question: "Which authentication method?",
      header: "Auth Method",
      options: [
        {label: "JWT Tokens", description: "Stateless, scalable"},
        {label: "Session-based", description: "Traditional, server-side"},
        {label: "OAuth 2.0", description: "Third-party (Google, GitHub)"}
      ]
    },
    {
      question: "Which platforms?",
      header: "Platform",
      multiSelect: true,
      options: [
        {label: "Web", description: "Browser-based"},
        {label: "Mobile", description: "iOS/Android apps"},
        {label: "API", description: "Programmatic access"}
      ]
    }
  ]
})
```

### Phase 2: Agent Selection & Delegation

**Decision Matrix**:
```
Requirement Analysis:
├─ Feature Type: Authentication (backend + frontend)
├─ Complexity: Medium-high (security implications)
├─ Agents Needed: backend-expert, ui-ux-expert, security-expert
└─ Coordination: Sequential (design → implement → validate)

Delegation Sequence:
1. Task(agent="backend-expert", task="Design JWT auth architecture")
2. Task(agent="ui-ux-expert", task="Design login UI (WCAG 2.1 AA)")
3. Task(agent="security-expert", task="OWASP Top 10 validation")
4. Task(agent="tdd-implementer", task="Implement with tests")
```

### Phase 3: Result Integration

**Synthesis Pattern**:
```
Collect Results:
├─ backend-expert: JWT strategy, rate limiting, audit logging
├─ ui-ux-expert: Accessible login form, error messaging
├─ security-expert: No critical vulnerabilities
└─ tdd-implementer: 96% test coverage

Alfred Synthesis:
├─ Combine all insights
├─ Verify consistency across agents
├─ Generate unified SPEC
└─ Report to user with next steps
```

---

## Advanced Patterns

### Multi-Agent Coordination

**Parallel Execution** (independent tasks):
```
Task(agent="backend-expert", task="API performance audit") & 
Task(agent="frontend-expert", task="Component accessibility audit") &
Task(agent="security-expert", task="Vulnerability scan")

# Wait for all 3 to complete
# Synthesize results
```

**Sequential Coordination** (dependent tasks):
```
Phase 1: Design
├─ Task(agent="spec-builder", task="Generate SPEC-001")
└─ Wait for completion...

Phase 2: Implementation (uses Phase 1 results)
├─ Task(agent="implementation-planner", task="Plan based on SPEC-001")
└─ Wait for completion...

Phase 3: Execution (uses Phase 2 results)
└─ Task(agent="tdd-implementer", task="Implement plan")
```

### Error Recovery

**Escalation Handling**:
```
If agent reports failure:
1. Analyze failure context
2. Determine recovery strategy:
   ├─ Retry with adjusted parameters
   ├─ Delegate to different agent
   ├─ Ask user for clarification
   └─ Abort and report impossible

3. Document failure + recovery in logs
```

---

## Best Practices

### ✅ DO
- Clarify ambiguous requirements before delegation
- Select appropriate agents based on task type
- Coordinate multi-agent workflows systematically
- Synthesize results for consistency
- Document delegation decisions
- Use AskUserQuestion for unclear requirements

### ❌ DON'T
- Delegate without understanding requirements
- Skip requirement clarification (leads to rework)
- Micromanage agents (trust specialization)
- Ignore agent failure signals
- Bypass orchestration for complex tasks

---

## Works Well With

- `moai-core-agent-guide` (Agent capabilities reference)
- `moai-core-ask-user-questions` (Interactive clarification)
- `moai-cc-subagent-lifecycle` (Subagent delegation)

---

**Version**: 1.0.0  
**Last Updated**: 2025-11-21  
**Status**: Production Ready
