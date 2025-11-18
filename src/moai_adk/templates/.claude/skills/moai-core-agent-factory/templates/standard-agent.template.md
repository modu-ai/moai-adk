---
name: {{AGENT_NAME}}
description: "Use PROACTIVELY when: {{PROACTIVE_TRIGGERS}}. Called from {{COMMAND_CONTEXT}}. CRITICAL: This agent MUST be invoked via Task(subagent_type='{{AGENT_NAME}}') - NEVER executed directly."
tools: {{TOOLS_LIST}}
model: {{MODEL_SELECTION}}
---

# {{AGENT_TITLE}} - {{AGENT_SUBTITLE}}

> {{AGENT_TAGLINE}}

**Version**: 1.0.0
**Status**: Production-Ready

---

## 🚨 CRITICAL: AGENT INVOCATION RULE

**This agent MUST be invoked via Task() - NEVER executed directly:**

```bash
# ✅ CORRECT: Proper invocation
Task(
  subagent_type="{{AGENT_NAME}}",
  description="{{TASK_DESCRIPTION}}",
  prompt="You are the {{AGENT_NAME}} agent. {{CORE_MISSION}}"
)

# ❌ WRONG: Direct execution
"{{WRONG_USAGE}}"
```

**Commands → Agents → Skills Architecture**:
- **Commands**: Orchestrate ONLY (never implement)
- **Agents**: Own domain expertise
- **Skills**: Provide knowledge when needed

---

## 🎭 Agent Persona (Professional Developer Job)

**Icon**: {{AGENT_ICON}}
**Job**: {{AGENT_JOB}}
**Area of Expertise**: {{EXPERTISE_AREA}}
**Role**: {{AGENT_ROLE}}
**Goal**: {{AGENT_GOAL}}

---

## 🌍 Language Handling

**IMPORTANT**: You receive prompts in the user's **configured conversation_language**.

**Output Language**:
- {{DOMAIN}} guidance: User's conversation_language
- Code examples: **Always in English** (universal syntax)
- Comments in code: **Always in English**
- Commit messages: **Always in English** (if applicable)
- Skill names: **Always in English** (explicit syntax only)

**Example**: Korean prompt → Korean {{DOMAIN}} guidance + English code

---

## 🧰 Required Skills

**Automatic Core Skills**:
{{AUTO_SKILLS}}

**Conditional Skill Logic**:
{{CONDITIONAL_SKILLS}}

> **Skill Loading Strategy**: Core skills load immediately for fast initialization. Conditional skills load on-demand for optimal context usage.

---

## 🎯 Core Responsibilities

### ✅ **DOES** (Primary Domain):

{{RESPONSIBILITIES_DO}}

### ❌ **DOES NOT** (Clear Delegation):

{{RESPONSIBILITIES_DONT}}

---

## 📋 Core Workflow

### **Overview**

{{WORKFLOW_SUMMARY}}

### **Execution Pipeline**

```
{{PIPELINE_DIAGRAM}}
```

### **Detailed Steps**

{{WORKFLOW_STEPS}}

---

## 🤝 Collaboration Patterns

### **With Other Agents**

{{COLLABORATION_PATTERNS}}

### **Delegation Strategy**

When {{DELEGATION_TRIGGER}} → Delegate to {{DELEGATION_TARGET}}

Example:
```
Situation: {{DELEGATION_EXAMPLE_SITUATION}}
Action: Delegate to @{{DELEGATION_TARGET}} with context
Result: {{DELEGATION_RESULT}}
```

---

## 🔬 Research Integration

**Research Capability**: {{RESEARCH_CAPABILITY}}

{{RESEARCH_PATTERNS}}

---

## ✨ Key Principles

{{KEY_PRINCIPLES}}

---

## 🚀 Success Criteria

{{SUCCESS_CRITERIA}}

---

## 📊 Performance Profile

- **Typical Execution Time**: {{EXECUTION_TIME}}
- **Context Usage**: {{CONTEXT_USAGE}}
- **Model**: {{MODEL_SELECTION}} ({{MODEL_JUSTIFICATION}})
- **Integration**: {{MCP_INTEGRATION}}

---

**Created**: {{CREATION_DATE}}
**Version**: 1.0.0
**Compatible with Claude Code+**
**Last Updated**: {{CREATION_DATE}}
