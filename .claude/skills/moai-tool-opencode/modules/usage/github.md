---
source: https://opencode.ai/docs/github/
fetched: 2026-01-08
title: GitHub Integration
---

# OpenCode GitHub Integration

## Overview

OpenCode enables AI-powered automation within GitHub workflows. The platform integrates with issues, pull requests, and scheduled tasks through GitHub Actions.

## Core Capabilities

### Triggering OpenCode

Mention `/opencode` or `/oc` in your comment, and OpenCode will execute tasks within your GitHub Actions runner.

### Common Use Cases

- Request issue explanations
- Implement fixes on new branches with PR submissions
- Make targeted code changes
- Automate code reviews
- Schedule repository maintenance

## Installation Methods

### Automated Setup

```bash
opencode github install
```

### Manual Configuration

1. Install the GitHub app at `github.com/apps/opencode-agent`
2. Add a workflow YAML file to `.github/workflows/opencode.yml`
3. Store API keys in GitHub Secrets

### Workflow YAML Example

```yaml
name: OpenCode Agent

on:
  issue_comment:
    types: [created]
  pull_request_review_comment:
    types: [created]
  issues:
    types: [opened, edited]
  pull_request:
    types: [opened, synchronize]
  schedule:
    - cron: '0 0 * * *'
  workflow_dispatch:

jobs:
  opencode:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: opencode-ai/opencode-action@v1
        with:
          model: anthropic/claude-sonnet-4
        env:
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
```

## Configuration Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `model` | Yes | Specifies the AI model in `provider/model` format |
| `agent` | No | Selects the processing agent; defaults to `default_agent` from config |
| `share` | No | Controls session sharing (defaults to true for public repos) |
| `prompt` | No | Customizes behavior through instructions |
| `token` | No | Provides GitHub authentication; uses installation token by default |

## Supported Event Types

### Issue Comments

Responds to comments on issues with full thread context.

```yaml
on:
  issue_comment:
    types: [created]
```

**Usage:**
```
/opencode Explain this issue and suggest a fix
```

### Pull Request Review Comments

Responds to code review comments with line numbers and diffs.

```yaml
on:
  pull_request_review_comment:
    types: [created]
```

**Usage:**
```
/oc Can you improve this function?
```

### Issues Creation/Editing

Automatically processes new or updated issues.

```yaml
on:
  issues:
    types: [opened, edited]
```

### Pull Request Events

Responds to PR opens and updates.

```yaml
on:
  pull_request:
    types: [opened, synchronize]
```

### Scheduled Tasks (Cron)

Runs on a schedule for maintenance tasks.

```yaml
on:
  schedule:
    - cron: '0 0 * * *'  # Daily at midnight
```

### Manual Dispatch

Trigger workflows manually from GitHub UI.

```yaml
on:
  workflow_dispatch:
```

## Common Applications

### Code Review Automation

```
/opencode Review this PR for security issues and best practices
```

### Issue Triage

```
/oc Categorize this issue and suggest priority
```

### Feature Implementation

```
/opencode Implement the feature described in this issue and create a PR
```

### Repository Maintenance

Schedule daily checks for outdated dependencies, security vulnerabilities, or code quality issues.

## Best Practices

1. **Store secrets securely:** Use GitHub Secrets for API keys
2. **Limit permissions:** Use minimal required permissions
3. **Review generated code:** Always review AI-generated changes before merging
4. **Set appropriate triggers:** Choose triggers that match your workflow
