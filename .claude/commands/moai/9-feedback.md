---
description: "Submit feedback or report issues"
argument-hint: "[issue|suggestion|question]"
type: local
allowed-tools: Read, Write, Edit, Bash, Grep, Glob, TodoWrite, AskUserQuestion
model: sonnet
---

## Pre-execution Context

!git status --porcelain
!git branch --show-current
!git log --oneline -1

## Essential Files

@.moai/config/config.yaml

---

# MoAI-ADK Step 9: Feedback Loop

> Architecture: Commands → Agents → Skills. This command orchestrates ONLY through `Task()` tool.
> Delegation Model: Feedback collection delegated to `manager-quality` agent.

Workflow Integration: This command implements the feedback loop of the MoAI workflow, allowing users to report issues or suggestions directly from the CLI.

---

## Command Purpose

Collect user feedback, bug reports, or feature suggestions and create GitHub issues automatically.

Run on: `$ARGUMENTS` (Feedback type)

---

## Execution Philosophy

/moai:9-feedback performs feedback collection through agent delegation:

Execution Flow:

- User Command: /moai:9-feedback [type]
- Phase 1: Task with subagent_type "manager-quality"
  - Analyze feedback type
  - Collect details via AskUserQuestion
  - Create GitHub Issue via Skill
- Output: Issue created with link

### Key Principle: Full Delegation Pattern

This command exclusively uses these tools:

- [HARD] Task() for agent delegation
  WHY: Task orchestration ensures feedback collection responsibility lies with specialized agents
  IMPACT: Direct tool usage bypasses quality assurance and agent expertise

- [HARD] AskUserQuestion() for user interaction AT COMMAND LEVEL ONLY
  WHY: Subagents via Task() are stateless and cannot interact with users directly
  IMPACT: Expecting agents to use AskUserQuestion causes workflow failures
  CORRECT: Command collects feedback type and details, passes to agent as parameters

- [HARD] Delegate all Bash operations to manager-quality agent
  WHY: Agent context ensures proper error handling and feedback capture
  IMPACT: Direct Bash execution loses feedback traceability and error context

---

## Associated Agents and Skills

Associated Agents for Feedback Collection:

- manager-quality: Feedback collection and GitHub issue creation

---

## Agent Invocation Patterns (CLAUDE.md Compliance)

This command uses agent execution patterns defined in CLAUDE.md (lines 96-120).

### Sequential Phase-Based Chaining PASS

Command implements simple sequential execution through 2 phases:

Phase Flow:

- Phase 1: Feedback Collection (manager-quality analyzes type and collects details)
- Phase 2: GitHub Issue Creation (manager-quality creates issue with collected information)

Each phase receives outputs from previous phase as context.

WHY: Sequential execution ensures complete feedback capture before submission

- Phase 2 requires validated feedback details from Phase 1
- Issue creation requires all user input to be collected

IMPACT: Skipping Phase 1 would create incomplete GitHub issues

### Parallel Execution FAIL

Not applicable - simple linear workflow

WHY: Feedback workflow has minimal complexity

- Only one agent (manager-quality) handles entire process
- Single feedback submission at a time
- No independent operations to parallelize

IMPACT: Parallel execution unnecessary for single-agent linear workflow

### Resumable Agent Support FAIL

Not applicable - command completes in single execution

WHY: Feedback submission is fast atomic operation

- Typical execution completes in under 30 seconds
- GitHub API calls are atomic and fast
- No long-running processes requiring checkpoints

IMPACT: Resume pattern unnecessary for simple feedback workflows

---

Refer to CLAUDE.md "Agent Chaining Patterns" (lines 96-120) for complete pattern architecture.

---

## Execution Process

### Step 1: Delegate to Quality Gate Agent

Use the manager-quality subagent to collect and submit user feedback:

Task: Collect user feedback and create a GitHub issue.

Context:

- Feedback Type: $ARGUMENTS (default to 'issue' if empty)
- Conversation Language: {{CONVERSATION_LANGUAGE}}

Instructions:

1. Determine Feedback Type:

   [HARD] Resolve feedback type from $ARGUMENTS if provided
   WHY: Pre-specified type accelerates feedback collection
   IMPACT: Skipping argument check forces unnecessary user interaction

   [HARD] Prompt user to select type when $ARGUMENTS is empty
   WHY: Interactive selection ensures proper categorization
   IMPACT: Assuming default type may misclassify feedback

   Supported Feedback Types:
   - Bug Report: Technical issues or errors
   - Feature Request: Suggestions for improvements
   - Question/Other: Clarifications or general feedback

2. Collect Details:

   [HARD] Solicit feedback title from user
   WHY: Title establishes feedback summary for issue search and triage
   IMPACT: Missing title reduces issue discoverability

   [HARD] Solicit detailed description from user
   WHY: Description provides context for developers to understand and respond
   IMPACT: Vague descriptions create follow-up communication overhead

   [SOFT] Solicit priority level from user (Low/Medium/High)
   WHY: Priority directs team resource allocation
   IMPACT: Missing priority defaults to normal urgency

3. Create GitHub Issue:

   [HARD] Execute GitHub CLI (gh issue create) command with collected feedback
   WHY: GitHub integration ensures feedback enters official issue tracking system
   IMPACT: Untracked feedback is lost to follow-up and implementation

   [HARD] Apply appropriate labels based on feedback type
   WHY: Labels enable automated triage and dashboard organization
   IMPACT: Untagged issues are invisible to responsible team members

   [SOFT] Format issue body using consistent template
   WHY: Standardized templates improve issue clarity and consistency
   IMPACT: Inconsistent formatting wastes developer time parsing issues

4. Report Result:
   [HARD] Provide user with created issue URL
   WHY: Direct issue link enables immediate user access and tracking
   IMPACT: Missing link requires users to manually search for their feedback

   [HARD] Confirm successful feedback submission to user
   WHY: Confirmation provides closure and acknowledgment of user contribution
   IMPACT: Silent completion leaves user uncertain about feedback status

Language and Accessibility:

- [HARD] Use conversation_language for all user-facing interactions
  WHY: User language ensures comprehension and accessibility
  IMPACT: Wrong language creates friction and reduces usability

- [HARD] Provide text-based options in AskUserQuestion (exclude emojis)
  WHY: Text options ensure consistency across all platforms and locales
  IMPACT: Emoji options are platform-dependent and may not display correctly

---

## Summary: Execution Verification Checklist

Before considering command execution complete, verify all requirements:

- [HARD] Agent Invoked: manager-quality agent executed with feedback details
  Verification: Confirm agent Task() call with feedback context
  Acceptance Criteria: Agent response confirms feedback reception

- [HARD] Feedback Collected: User provided title, description, and type
  Verification: Confirm all required fields captured
  Acceptance Criteria: No empty or null feedback fields

- [HARD] Issue Created: GitHub issue successfully submitted to repository
  Verification: Confirm gh issue create command executed successfully
  Acceptance Criteria: GitHub API response contains issue ID and URL

- [HARD] Link Provided: User received direct URL to created issue
  Verification: Confirm issue URL displayed in user response
  Acceptance Criteria: User can click URL and access their feedback immediately

---

## Quick Reference

Scenario: Report bug

- Entry Point: /moai:9-feedback issue
- Expected Outcome: GitHub issue created with bug label

Scenario: Request feature

- Entry Point: /moai:9-feedback suggestion
- Expected Outcome: GitHub issue created with enhancement label

Scenario: Ask question

- Entry Point: /moai:9-feedback question
- Expected Outcome: GitHub issue created with question label

Scenario: General feedback

- Entry Point: /moai:9-feedback
- Expected Outcome: Interactive feedback collection

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

## Output Format

Structure agent responses with semantic sections for clarity and consistency:

Response Structure:

The agent response must include these sections:

Feedback Summary:
Concise recap of feedback type, title, and key points submitted

GitHub Integration:
Confirmation that feedback was submitted to GitHub with issue URL and ID

User Language:
Response provided in user's conversation_language configuration

Next Actions:
Clear options for user to continue workflow or submit additional feedback

WHY: Semantic structure ensures consistent response quality and user experience
IMPACT: Unstructured responses create ambiguity and require user clarification

---

## Post-Submission User Direction

After successful feedback submission, present user with next action options:

[HARD] Display post-submission summary to user
WHY: Summary confirms successful completion and provides issue reference
IMPACT: Missing summary leaves user uncertain about submission status

[SOFT] Offer user next step choices via AskUserQuestion
WHY: Guided choices help users continue productive workflows
IMPACT: Abrupt completion requires user to determine next actions

Next Step Options:

- Continue Development: Return to current development workflow
- Submit Additional Feedback: Report another issue or suggestion
- View Issue: Open created GitHub issue in browser

Requirements:

- [HARD] Use conversation_language for all post-submission messaging
  WHY: Consistent language maintains user context
  IMPACT: Language switching disorients users

- [HARD] Express all options as text labels without emoji characters
  WHY: Text-only labels ensure universal platform compatibility
  IMPACT: Emojis may fail to display across different systems

---

## EXECUTION DIRECTIVE

You must NOW execute the command following the "Execution Process" described above.

1. Use the manager-quality subagent.
2. Do NOT just describe what you will do. DO IT.
