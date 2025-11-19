---
name: moai:9-feedback
description: "Submit feedback or report issues"
argument-hint: "[issue|suggestion|question]"
allowed-tools:
  - Task
  - AskUserQuestion
---

# 🗣️ MoAI-ADK Step 9: Feedback Loop

> **Architecture**: Commands → Agents → Skills. This command orchestrates ONLY through `Task()` tool.
> **Delegation Model**: Feedback collection delegated to `quality-gate` agent.

**Workflow Integration**: This command implements the feedback loop of the MoAI workflow, allowing users to report issues or suggestions directly from the CLI.

---

## 🎯 Command Purpose

Collect user feedback, bug reports, or feature suggestions and create GitHub issues automatically.

**Run on**: `$ARGUMENTS` (Feedback type)

---

## 💡 Execution Philosophy

```text
/moai:9-feedback
```

performs feedback collection through agent delegation:

User Command: /moai:9-feedback [type]
↓
Phase 1: Task(subagent_type="quality-gate")
→ Analyze feedback type
→ Collect details via AskUserQuestion
→ Create GitHub Issue via Skill
↓
Output: Issue created with link

````

### Key Principle: Zero Direct Tool Usage

**This command uses ONLY these tools:**

- ✅ **Task()** for agent delegation
- ✅ **AskUserQuestion()** for user interaction (delegated to agent)
- ❌ No Bash (delegated to agent)

---

## 🚀 Execution Process

### Step 1: Delegate to Quality Gate Agent

Use Task tool to call the `quality-gate` agent (which has access to issue creation skills):

```yaml
Tool: Task
Parameters:
- subagent_type: "quality-gate"
- description: "Collect and submit user feedback"
- prompt: """You are the quality-gate agent acting as the feedback manager.

**Task**: Collect user feedback and create a GitHub issue.

**Context**:
- Feedback Type: $ARGUMENTS (default to 'issue' if empty)
- Conversation Language: {{CONVERSATION_LANGUAGE}}

**Instructions**:

1. **Determine Feedback Type**:
   - If $ARGUMENTS is provided, use it.
   - If not, ask user to select type:
     - Bug Report
     - Feature Request
     - Improvement
     - Refactoring
     - Documentation
     - Question/Other

2. **Load Feedback Template**:
   - Use `Skill("moai-core-feedback-templates")` to retrieve the appropriate template for the selected type.
   - Read the template structure (Description, Scenario/Reproduction, Expected vs Actual, etc.).

3. **Collect Details**:
   - Ask for 'Title' (short summary).
   - Ask for specific details required by the template (e.g., "Reproduction Steps" for bugs, "Use Case" for features).
   - Ask for 'Priority' (Low/Medium/High).

4. **Create GitHub Issue**:
   - Use `Skill("moai-core-issue-labels")` or `Bash` (gh issue create) to submit.
   - Add appropriate labels based on type (bug, enhancement, documentation, etc.).
   - **CRITICAL**: Format the issue body EXACTLY according to the loaded template.

5. **Report Result**:
   - Show the created issue URL.
   - Confirm success to the user.

**Important**:
- Use conversation_language for all user interactions.
- NO EMOJIS in AskUserQuestion options.
- Ensure the issue body follows the structured template from the skill.
"""
````

---

## 🎯 Summary: Your Execution Checklist

Before you consider this command complete, verify:

- [ ] **Agent Called**: `quality-gate` agent was invoked.
- [ ] **Feedback Collected**: User was asked for details.
- [ ] **Issue Created**: GitHub issue was successfully created.
- [ ] **Link Provided**: User received the issue URL.

---

## ⚡️ EXECUTION DIRECTIVE

**You must NOW execute the command following the "Execution Process" described above.**

1. Call the `Task` tool with `subagent_type="quality-gate"`.
2. Do NOT just describe what you will do. DO IT.
