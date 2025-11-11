---
name: moai-alfred-agent-guide
version: 4.0.0
created: 2025-10-01
updated: 2025-11-12
status: active
tier: specialization
description: "19-agent team structure, decision trees for agent selection, Haiku vs Sonnet model selection, and agent collaboration principles. Enhanced with research capabilities for agent performance analysis and optimization. Use when deciding which sub-agent to invoke, understanding team responsibilities, or learning multi-agent orchestration.. Enhanced with Context7 MCP for up-to-date documentation."
allowed-tools: "Read, Glob, Grep, TodoWrite, WebSearch, WebFetch, mcp__context7__resolve-library-id, mcp__context7__get-library-docs"
primary-agent: "alfred"
secondary-agents: [session-manager, plan-agent]
keywords: [alfred, agent, guide, git, frontend]
tags: [agent, coordination, decision-tree, research, analysis, optimization, team-management,
  performance]
orchestration:
  can_resume: true
  typical_chain_position: "initial"
  depends_on: []
---

# moai-alfred-agent-guide

**Alfred Agent Guide**

> **Primary Agent**: alfred  
> **Secondary Agents**: session-manager, plan-agent  
> **Version**: 4.0.0  
> **Keywords**: alfred, agent, guide, git, frontend

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (Core Concepts)

## What It Does

MoAI-ADK의 19개 Sub-agent 아키텍처, 어떤 agent를 선택할지 결정하는 트리, Haiku/Sonnet 모델 선택 기준을 정의합니다.

---

### Level 2: Practical Implementation (Common Patterns)

Agent Delegation Patterns (v5.0.0)

### Commands → Agents → Skills Architecture

**CRITICAL RULES**:
1. **Commands NEVER execute directly** - Always orchestrate via agents
2. **Agents own domain expertise** - Handle complex reasoning and decisions
3. **Skills provide reusable knowledge** - Called by agents when needed

### Proper Delegation Templates

#### For Commands (Orchestration Only):
```bash
# ❌ WRONG: Direct execution
"Implement SPEC-001"

# ✅ CORRECT: Agent delegation
Task(
  subagent_type="tdd-implementer",
  description="Execute TDD implementation for SPEC-001",
  prompt="You are the tdd-implementer agent. Execute SPEC-001 using TDD cycle."
)
```

#### For Agents (Domain Execution):
```bash
# ❌ WRONG: Direct skill execution without context
Skill("moai-domain-backend")

# ✅ CORRECT: Skill loading with proper context
Skill("moai-domain-backend")  # Load domain knowledge
# Then apply to specific task with context
```

#### For Specialist Agent Activation:
```bash
# ❌ WRONG: Manual domain work
"Design backend API for user authentication"

# ✅ CORRECT: Delegate to domain expert
Task(
  subagent_type="backend-expert",
  description="Design and implement backend authentication system",
  prompt="You are the backend-expert agent. Design comprehensive authentication API."
)
```

### Agent Collaboration Protocols

#### Cross-Agent Coordination:
```bash
# Backend expert coordinates with frontend expert
Task(
  subagent_type="backend-expert",
  description="Create API contract for frontend integration",
  prompt="Coordinate with frontend-expert for API contract. Design endpoints that frontend can consume."
)
```

#### Sequential Agent Handoffs:
```bash
# 1. Plan agent creates strategy
Task(subagent_type="Plan", ...)
# 2. Implementation agent executes
Task(subagent_type="tdd-implementer", ...)
# 3. Quality agent validates
Task(subagent_type="quality-gate", ...)
```

### Anti-Patterns to Avoid

❌ **Direct Command Execution**: Commands implementing features directly
❌ **Agent Bypassing**: Using skills without proper agent context
❌ **Mixed Responsibilities**: Commands doing both orchestration AND implementation
❌ **Unclear Delegation**: Ambiguous handoffs between agents

### Best Practices

✅ **Clear Ownership**: Each task has one responsible agent
✅ **Proper Handoffs**: Explicit agent-to-agent communication
✅ **Skill Context**: Skills loaded within agent domain context
✅ **Traceable Work**: Every action traceable to responsible agent

---

Learn more in `reference.md` for complete agent responsibilities, collaboration patterns, and advanced orchestration strategies.

**Related Skills**: moai-alfred-rules, moai-alfred-practices

---

### Level 3: Advanced Patterns (Expert Reference)

> **Note**: Advanced patterns for complex scenarios.

**Coming soon**: Deep dive into expert-level usage.


---

## 🎯 Best Practices Checklist

**Must-Have:**
- ✅ [Critical practice 1]
- ✅ [Critical practice 2]

**Recommended:**
- ✅ [Recommended practice 1]
- ✅ [Recommended practice 2]

**Security:**
- 🔒 [Security practice 1]


---

## 🔗 Context7 MCP Integration

**When to Use Context7 for This Skill:**

This skill benefits from Context7 when:
- Working with [alfred]
- Need latest documentation
- Verifying technical details

**Example Usage:**

```python
# Fetch latest documentation
from moai_adk.integrations import Context7Helper

helper = Context7Helper()
docs = await helper.get_docs(
    library_id="/org/library",
    topic="alfred",
    tokens=5000
)
```

**Relevant Libraries:**

| Library | Context7 ID | Use Case |
|---------|-------------|----------|
| [Library 1] | `/org/lib1` | [When to use] |


---

## 📊 Decision Tree

**When to use moai-alfred-agent-guide:**

```
Start
  ├─ Need alfred?
  │   ├─ YES → Use this skill
  │   └─ NO → Consider alternatives
  └─ Complex scenario?
      ├─ YES → See Level 3
      └─ NO → Start with Level 1
```


---

## 🔄 Integration with Other Skills

**Prerequisite Skills:**
- Skill("prerequisite-1") – [Why needed]

**Complementary Skills:**
- Skill("complementary-1") – [How they work together]

**Next Steps:**
- Skill("next-step-1") – [When to use after this]


---

## 📚 Official References

Agent Delegation Patterns (v5.0.0)

### Commands → Agents → Skills Architecture

**CRITICAL RULES**:
1. **Commands NEVER execute directly** - Always orchestrate via agents
2. **Agents own domain expertise** - Handle complex reasoning and decisions
3. **Skills provide reusable knowledge** - Called by agents when needed

### Proper Delegation Templates

#### For Commands (Orchestration Only):
```bash
# ❌ WRONG: Direct execution
"Implement SPEC-001"

# ✅ CORRECT: Agent delegation
Task(
  subagent_type="tdd-implementer",
  description="Execute TDD implementation for SPEC-001",
  prompt="You are the tdd-implementer agent. Execute SPEC-001 using TDD cycle."
)
```

#### For Agents (Domain Execution):
```bash
# ❌ WRONG: Direct skill execution without context
Skill("moai-domain-backend")

# ✅ CORRECT: Skill loading with proper context
Skill("moai-domain-backend")  # Load domain knowledge
# Then apply to specific task with context
```

#### For Specialist Agent Activation:
```bash
# ❌ WRONG: Manual domain work
"Design backend API for user authentication"

# ✅ CORRECT: Delegate to domain expert
Task(
  subagent_type="backend-expert",
  description="Design and implement backend authentication system",
  prompt="You are the backend-expert agent. Design comprehensive authentication API."
)
```

### Agent Collaboration Protocols

#### Cross-Agent Coordination:
```bash
# Backend expert coordinates with frontend expert
Task(
  subagent_type="backend-expert",
  description="Create API contract for frontend integration",
  prompt="Coordinate with frontend-expert for API contract. Design endpoints that frontend can consume."
)
```

#### Sequential Agent Handoffs:
```bash
# 1. Plan agent creates strategy
Task(subagent_type="Plan", ...)
# 2. Implementation agent executes
Task(subagent_type="tdd-implementer", ...)
# 3. Quality agent validates
Task(subagent_type="quality-gate", ...)
```

### Anti-Patterns to Avoid

❌ **Direct Command Execution**: Commands implementing features directly
❌ **Agent Bypassing**: Using skills without proper agent context
❌ **Mixed Responsibilities**: Commands doing both orchestration AND implementation
❌ **Unclear Delegation**: Ambiguous handoffs between agents

### Best Practices

✅ **Clear Ownership**: Each task has one responsible agent
✅ **Proper Handoffs**: Explicit agent-to-agent communication
✅ **Skill Context**: Skills loaded within agent domain context
✅ **Traceable Work**: Every action traceable to responsible agent

---

Learn more in `reference.md` for complete agent responsibilities, collaboration patterns, and advanced orchestration strategies.

**Related Skills**: moai-alfred-rules, moai-alfred-practices

---

## 📈 Version History

**v4.0.0** (2025-11-12)
- ✨ Context7 MCP integration
- ✨ Progressive Disclosure structure
- ✨ 10+ code examples
- ✨ Primary/secondary agents defined
- ✨ Best practices checklist
- ✨ Decision tree
- ✨ Official references



---

**Generated with**: MoAI-ADK Skill Factory v4.0  
**Last Updated**: 2025-11-12  
**Maintained by**: Primary Agent (alfred)
