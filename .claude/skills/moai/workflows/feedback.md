# Workflow: feedback - GitHub Issue Creation

Purpose: Collect user feedback, bug reports, or feature suggestions and create GitHub issues automatically via the manager-quality agent.

---

## Phase 1: Feedback Collection

### Step 1: Determine Feedback Type

[HARD] Resolve feedback type from $ARGUMENTS if provided (issue, suggestion, question).

If $ARGUMENTS is empty, use AskUserQuestion:

Question: What type of feedback would you like to submit?

Options:

- Bug Report: Technical issues or errors encountered
- Feature Request: Suggestions for improvements or new features
- Question: Clarifications or help needed

### Step 2: Collect Details

[HARD] Solicit feedback title from user via AskUserQuestion (free text input).

[HARD] Solicit detailed description from user via AskUserQuestion (free text input).

[SOFT] Solicit priority level from user:

- Low: Minor issue, workaround available
- Medium: Moderate impact, no urgent workaround needed
- High: Significant impact, blocks workflow

---

## Phase 2: GitHub Issue Creation

[HARD] Delegate to manager-quality subagent with collected feedback details.

Pass to manager-quality:

- Feedback Type: Bug Report, Feature Request, or Question
- Title: User-provided title
- Description: User-provided description
- Priority: Selected priority level
- Conversation Language: From config

### GitHub Issue Labels

- Bug Report: labels "bug"
- Feature Request: labels "enhancement"
- Question: labels "question"

### Issue Creation Command

The manager-quality agent executes: gh issue create --repo modu-ai/moai-adk

Issue body uses a consistent template including:

- Feedback type header
- Description content
- Priority level
- Environment information (MoAI version, OS)

### Result Reporting

[HARD] Provide user with the created issue URL.
[HARD] Confirm successful feedback submission to user.

Display in user's conversation_language:

- Issue number and title
- Direct URL to the created issue
- Applied labels

---

## Post-Submission Options

Use AskUserQuestion after successful submission:

- Continue Development: Return to current development workflow
- Submit Additional Feedback: Report another issue or suggestion
- View Issue: Open created GitHub issue in browser

---

## Execution Pattern

This workflow uses simple sequential execution (no parallelism needed):

- Phase 1 collects all user input at MoAI orchestrator level
- Phase 2 delegates to manager-quality with complete context
- Single agent handles the entire submission process
- Typical execution completes in under 30 seconds

Resume support: Not applicable (atomic operation).

---

## Agent Chain Summary

- Phase 1: MoAI orchestrator (AskUserQuestion for feedback collection)
- Phase 2: manager-quality subagent (GitHub issue creation via gh CLI)

---

Version: 1.0.0
Last Updated: 2026-01-28
