---
description: >
  Collects user feedback, bug reports, or feature suggestions and creates
  GitHub issues automatically via the orchestrator-direct gh CLI. Supports bug
  reports, feature requests, and questions with priority classification.
  Use when submitting feedback, reporting bugs, or requesting features.
user-invocable: false
metadata:
  version: "2.5.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-21"
  tags: "feedback, bug-report, feature-request, github-issues, quality"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["feedback", "bug", "issue", "suggestion", "report", "feature request"]
  agents: []
  phases: ["feedback"]
---

# Workflow: feedback - GitHub Issue Creation

Purpose: Collect user feedback, bug reports, or feature suggestions and create GitHub issues automatically via the orchestrator-direct `gh` CLI (the archived-agent-rejection.md §C migration table provides no retained-agent owner for issue creation, so issue creation is performed orchestrator-direct).

Prerequisite: The `gh` CLI must be installed and authenticated (`gh auth status`). If not available, guide user to install via https://cli.github.com/.

### gh Availability and Failure Fallback

[HARD] Before issue creation, run `gh auth status`. When `gh` is unauthenticated OR the GitHub API is rate-limited (detected via a non-zero `gh auth status` exit, or a rate-limit signal such as an HTTP 403 / "rate limit exceeded" message on a `gh` call), the workflow MUST follow a graceful fallback path instead of failing silently:

1. Report the detected condition (unauthenticated vs rate-limited) to the user in `conversation_language`.
2. Guide the user to resolve it — run `gh auth login` for the unauthenticated case, or wait for the rate-limit window to reset for the rate-limited case.
3. Offer to save the drafted issue body locally (e.g. under `.moai/state/feedback-draft-<timestamp>.md`) so no drafted content is lost. On acceptance, write the full drafted title + body to that path and report the saved path to the user.

No drafted feedback is discarded on a `gh` failure; the local draft is the recovery artifact.

---

## Phase 1: Feedback Collection

### Step 1: Determine Feedback Type

[HARD] Resolve feedback type from $ARGUMENTS if provided (issue, suggestion, question).

If $ARGUMENTS is empty, use AskUserQuestion:

Question: What type of feedback would you like to submit?

Options:

- Bug Report (Recommended): Technical issues or errors encountered
- Feature Request: Suggestions for improvements or new features
- Question: Clarifications or help needed

### Step 2: Collect Details

[HARD] Solicit feedback title from user via AskUserQuestion (free text input).

[HARD] Solicit detailed description from user via AskUserQuestion (free text input).

[SOFT] Solicit priority level from user:

- Low: Minor issue, workaround available
- Medium: Moderate impact, no urgent workaround needed
- High: Significant impact, blocks workflow

### Step 3: Duplicate Detection

[HARD] After the title is drafted and before issue creation, run a duplicate-issue search against the resolved feedback target repository:

`gh issue list --repo <resolved-target> --search "<title keywords>" --state open`

- `<resolved-target>` is the resolved feedback target repository (config `feedback.repository`, default `modu-ai/moai-adk`).
- `<title keywords>` are the salient keywords extracted from the drafted title.

[HARD] The duplicate-detection step emits a structured "possible duplicates" candidate-report — the list of likely-duplicate issues (issue number, title, URL, state) — for the orchestrator to present. The step itself MUST NOT prompt the user inline; it only produces the candidate-report. The orchestrator decides at its own level whether to proceed with a new issue, link to an existing one, or continue.

---

## Phase 2: GitHub Issue Creation

[HARD] Create the GitHub issue orchestrator-direct via the `gh` CLI with collected feedback details (no subagent spawn — issue creation has no retained-agent owner per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C).

Inputs for the `gh issue create` invocation:

- Feedback Type: Bug Report, Feature Request, or Question
- Title: User-provided title
- Description: User-provided description
- Priority: Selected priority level
- Conversation Language: From config

### GitHub Issue Labels

- Bug Report: labels "bug"
- Feature Request: labels "enhancement"
- Question: labels "question"

### Issue Language Policy

[HARD] GitHub Issue title and template headers MUST be written in the user's `conversation_language`:

- **Title**: Written in conversation_language
- **Body template headers**: Section names (e.g., "Description", "Priority", "Environment") translated to conversation_language
- **Body content**: User-provided text preserved verbatim (not translated, even if language differs from conversation_language)
- **Labels**: English only (GitHub standard: "bug", "enhancement", "question")

Language examples:

| conversation_language | Title example | Section header example |
|----------------------|---------------|----------------------|
| en | "Login fails with 500 error" | "## Description", "## Priority", "## Environment" |
| ko | "로그인 시 500 에러 발생" | "## 설명", "## 우선순위", "## 환경 정보" |
| ja | "ログイン時に500エラー発生" | "## 説明", "## 優先度", "## 環境情報" |
| zh | "登录时出现500错误" | "## 描述", "## 优先级", "## 环境信息" |

### Issue Creation Command

The orchestrator executes directly: `gh issue create --repo <resolved-target>`, where `<resolved-target>` is the resolved feedback target repository (config `feedback.repository`, default `modu-ai/moai-adk`).

Issue body uses a consistent template in the user's conversation_language, including:

- Feedback type header (translated)
- Description content (user's original text)
- Priority level (translated)
- Tool-diagnostic information (see Diagnostic Attachment below)

### Diagnostic Attachment

[HARD] The workflow auto-collects and appends tool-diagnostic information to the issue body. Two items are GUARANTEED (always collected):

- MoAI version — `moai version`
- Operating system — `uname` (OS name / release)

Additionally, the workflow attempts the following on a best-effort basis. Their absence is NOT a failure — most users run a prebuilt `moai` binary with no Go toolchain installed, and the orchestrator may not carry error context:

- Go toolchain version — `go version` (tool build-provenance; describes the tool binary's provenance, not the user's project language — absent when no Go toolchain is installed)
- Last-failed-command / error context — appended additively ONLY when the orchestrator passes it into the feedback invocation. The workflow never reads session error history itself.

[HARD] Diagnostic attachment is restricted to tool-diagnostic information only. The workflow MUST NOT attach arbitrary user file contents to the issue body — no reading or embedding of source files, configuration files, or any user-supplied file contents beyond the tool-diagnostic set above.

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

- Continue Development (Recommended): Return to current development workflow
- Submit Additional Feedback: Report another issue or suggestion
- View Issue: Open created GitHub issue in browser

---

## Execution Pattern

This workflow uses simple sequential execution (no parallelism needed):

- Phase 1 collects all user input at MoAI orchestrator level
- Phase 2 creates the issue orchestrator-direct via the `gh` CLI with complete context
- The orchestrator handles the entire submission process (no subagent spawn)

Resume support: Not applicable (atomic operation).

---

## Agent Chain Summary

- Phase 1: MoAI orchestrator (AskUserQuestion for feedback collection)
- Phase 2: MoAI orchestrator-direct (GitHub issue creation via gh CLI)

---

Version: 2.0.0
Last Updated: 2026-02-07
