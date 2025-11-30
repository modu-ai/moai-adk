---
name: moai:9-feedback
description: "Submit feedback or report issues"
argument-hint: "[issue|suggestion|question]"
allowed-tools: Read, Write, Edit, Grep, Glob, WebFetch, WebSearch, Bash, TodoWrite, AskUserQuestion, Task, Skill
model: haiku
---

##  Pre-execution Context

!git status --porcelain
!git branch --show-current
!git log --oneline -1

##  Essential Files

@.moai/config/config.json

---

#  MoAI-ADK Step 9: Feedback Loop

> Architecture: Commands → Agents → Skills. This command orchestrates ONLY through `Task()` tool.
> Delegation Model: Feedback collection delegated to `manager-quality` agent.

Workflow Integration: This command implements the feedback loop of the MoAI workflow, allowing users to report issues or suggestions directly from the CLI.

---

##  Command Purpose

Collect user feedback, bug reports, or feature suggestions and create GitHub issues automatically.

Run on: `$ARGUMENTS` (Feedback type)

---

##  Execution Philosophy

`/moai:9-feedback` performs feedback collection through agent delegation:

```
User Command: /moai:9-feedback [type]
    ↓
Phase 1: Task(subagent_type="manager-quality")
    → Analyze feedback type
    → Collect details via AskUserQuestion
    → Create GitHub Issue via Skill
    ↓
Output: Issue created with link
```

### Key Principle: Zero Direct Tool Usage

This command uses ONLY these tools:

-  Task() for agent delegation
-  AskUserQuestion() for user interaction (delegated to agent)
-  No Bash (delegated to agent)

---

##  Associated Agents & Skills

| Agent/Skill     | Purpose                                       |
| --------------- | --------------------------------------------- |
| manager-quality | Feedback collection and GitHub issue creation |

---

##  Execution Process

### Step 1: Delegate to Quality Gate Agent

Use the manager-quality subagent to collect and submit user feedback:

Task: Collect user feedback and create a GitHub issue.

Context:

- Feedback Type: $ARGUMENTS (default to 'issue' if empty)
- Conversation Language: {{CONVERSATION_LANGUAGE}}

Instructions:

1. Determine Feedback Type:

   - If $ARGUMENTS is provided, use it.
   - If not, ask user to select type:
     - Bug Report
     - Feature Request
     - Question/Other

2. Collect Details:

   - Ask for 'Title' (short summary)
   - Ask for 'Description' (detailed explanation)
   - Ask for 'Priority' (Low/Medium/High)

3. Create GitHub Issue:

   - Use appropriate tools with GitHub CLI (`gh issue create`) to submit.
   - Add appropriate labels (bug, enhancement, question) via `--label` flag.
   - Format the body with standard templates.

4. Report Result:
   - Show the created issue URL.
   - Confirm success to the user.

Important:

- Use conversation_language for all user interactions.
- NO EMOJIS in AskUserQuestion options.

---

##  Summary: Your Execution Checklist

Before you consider this command complete, verify:

- [ ] Agent Called: `manager-quality` agent was invoked.
- [ ] Feedback Collected: User was asked for details.
- [ ] Issue Created: GitHub issue was successfully created.
- [ ] Link Provided: User received the issue URL.

---

##  Quick Reference

| Scenario         | Entry Point                   | Expected Outcome                            |
| ---------------- | ----------------------------- | ------------------------------------------- |
| Report bug       | `/moai:9-feedback issue`      | GitHub issue created with bug label         |
| Request feature  | `/moai:9-feedback suggestion` | GitHub issue created with enhancement label |
| Ask question     | `/moai:9-feedback question`   | GitHub issue created with question label    |
| General feedback | `/moai:9-feedback`            | Interactive feedback collection             |

Associated Agent:

- `manager-quality` - Feedback manager and GitHub issue creator

Feedback Types:

- Bug Report: Technical issues or errors
- Feature Request: Suggestions for improvements
- Question: Clarifications or help needed
- Other: General feedback

Version: 1.0.0 (Agent-Delegated Pattern)
Last Updated: 2025-11-25
Architecture: Commands → Agents → Skills (Complete delegation)

---

## Final Step: Next Action Selection

After feedback submission completes, use AskUserQuestion tool to guide user to next action:

```python
AskUserQuestion({
    "questions": [{
        "question": "Feedback submitted successfully. What would you like to do next?",
        "header": "Next Steps",
        "multiSelect": false,
        "options": [
            {
                "label": "Continue Development",
                "description": "Return to current workflow"
            },
            {
                "label": "Submit Another Feedback",
                "description": "Report another issue or suggestion"
            },
            {
                "label": "View Issue",
                "description": "Open created GitHub issue in browser"
            }
        ]
    }]
})
```

Important:

- Use conversation language from config
- No emojis in any AskUserQuestion fields
- Always provide clear next step options

---

##  EXECUTION DIRECTIVE

You must NOW execute the command following the "Execution Process" described above.

1. Use the manager-quality subagent.
2. Do NOT just describe what you will do. DO IT.
