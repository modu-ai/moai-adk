---
title: "Multi-LLM CI Guide"
description: "Automated code reviews with multiple AI models in GitHub Actions"
draft: false
weight: 10
---

This guide walks through setting up multi-LLM code review in GitHub Actions
with MoAI-ADK's Multi-LLM CI feature. There is no reason to lock reviewers to
a single model either — every model has different strengths and unit costs, so
the Tokenomics perspective of multi-LLM allocation applies to PR reviews just
the same.

## Overview

### What is Multi-LLM CI?

MoAI-ADK's Multi-LLM CI feature provides an integrated CI/CD pipeline that
runs code reviews with multiple AI models simultaneously in GitHub Actions.

### Supported LLMs

| LLM | Provider | Trigger | Notes |
|-----|--------|-------------|------|
| **Claude** | Anthropic | `/claude` comment | Issue/PR review, OAuth authentication |
| **Codex** | OpenAI | Auto on PR open | Private repositories only |
| **Gemini** | Google | Auto on PR open | API Key authentication |
| **GLM** | Zhipu AI | Auto on PR open | Token authentication |

## Getting started

### Prerequisites

- MoAI-ADK installed (macOS · Linux · Windows)
- A GitHub repository
- An account and API token for each LLM

### Initial setup

```bash
moai github init
```

What this command does:
- Creates the `.github/workflows/` directory
- Deploys the workflow templates
- Deploys the composite actions
- Guides you through the GitHub Secrets setup

### LLM authentication setup

```bash
# Claude (OAuth)
moai github auth claude

# Codex (private repos)
moai github auth codex

# Gemini
moai github auth gemini

# GLM
moai github auth glm
```

### GitHub Secrets setup

Secrets required per LLM:
- `CLAUDE_CODE_OAUTH_TOKEN` - Claude OAuth token
- `CODEX_AUTH_JSON` - Codex auth JSON (base64 encoded)
- `GEMINI_API_KEY` - Gemini API Key
- `GLM_API_KEY` - GLM API Token

### Testing your first PR

Opening a PR automatically adds an LLM Panel comment:

```markdown
## LLM Code Review Status

| LLM | Status |
|-----|--------|
| Claude | Pending (add `/claude` comment) |
| Codex | ✓ Ready |
| Gemini | ⚠️ Token missing |
| GLM | ✓ Ready |

Trigger individual reviews:
- Add `/claude` comment to trigger Claude
- Add `/codex` comment to trigger Codex
- Add `/gemini` comment to trigger Gemini
- Add `/glm` comment to trigger GLM
```

## LLM authentication details

### Claude setup

#### Issuing an OAuth token

1. Install [Claude Code](https://claude.ai/download)
2. Log in and issue an OAuth token
3. Automatically saved to `.claude/settings.local.json`

#### moai github auth claude

```bash
moai github auth claude
```

**Interactive setup process:**
```
Claude OAuth token not found.
Would you like to install Claude Code and log in? (y/n): y

[Confirmed] The OAuth token has been saved to settings.local.json.
Set the following value in the GitHub Secret CLAUDE_CODE_OAUTH_TOKEN:
<token-value>
```

### Codex setup (private repos only)

#### Creating the auth JSON

```json
{
  "token": "sk-...",
  "base_url": "https://api.openai.com/v1"
}
```

#### moai github auth codex

```bash
moai github auth codex
```

**Interactive setup:**
```
OpenAI auth.json file path: ~/.codex/auth.json
Reading the file and creating the GitHub Secret...
Note: Codex is only usable in private repositories (REQ-SEC-001)

Generated Secret:
CODEX_AUTH_JSON=eyJ0...
```

### Gemini setup

```bash
moai github auth gemini
```

After you enter the API Key, the GitHub Secret setup guide is provided
automatically.

### GLM setup

```bash
moai github auth glm
```

Reads automatically from the GLM token path (`~/.moai/.env.glm`).

## Understanding the workflow templates

### llm-panel.yml

**Trigger:** PR opened

**Role:** automatically creates the panel comment that visually shows each LLM's status

**Note:** trigger individual reviews with `/claude`, `/codex`, `/gemini`, `/glm` comments

### claude.yml / claude-code-review.yml

- **claude.yml**: Issue trigger (draft review)
- **claude-code-review.yml**: PR trigger (change review)

**Characteristic:** triggered only by the `/claude` comment

### codex-review.yml

**Security constraints:**
- Only runs on `private` repos (REQ-SEC-001)
- Blocks public repos via a `visibility` check

**workflow:**
```yaml
private-guard:
  runs-on: ubuntu-latest
  steps:
    - name: Check Repository Visibility
      run: |
        if [[ "${{ github.repository_visibility }}" == "public" ]]; then
          echo "::error::Codex review is restricted to private repositories"
          exit 1
        fi
```

### gemini-review.yml

- Automatic language detection (detect-language action)
- Auto-triggered on PR synchronized

### glm-review.yml

- GLM-specific environment setup (setup-glm-env action)
- Automatic environment variable injection

### Composite Actions

#### detect-language

**Input:** repository root path
**Output:** language environment variable (`detected_language`)

**Supported languages:** Go, Python, TypeScript, JavaScript, Rust, Java, Kotlin, C#, Ruby, PHP, Elixir, C++, Scala, R, Flutter, Swift (16)

#### setup-glm-env

Sets the environment variables required by GLM team mode:
- `ANTHROPIC_AUTH_TOKEN` (GLM endpoint)
- `ANTHROPIC_BASE_URL` (https://glm.modu-ai.kr)

## Advanced configuration

### Customizing github-actions.yaml

#### Basic structure

```yaml
# .moai/config/sections/github-actions.yaml
llm_review:
  enabled: true
  runners:
    claude: true
    codex: true
    gemini: true
    glm: true
  triggers:
    on_pr_open: true
    on_comment:
      claude: "/claude"
      codex: "/codex"
      gemini: "/gemini"
      glm: "/glm"
```

#### Per-language LLM assignment

```yaml
language_rules:
  go:
    - gemini
    - claude
  python:
    - claude
    - glm
  typescript:
    - codex
    - claude
```

### Runner version management

#### Checking for updates automatically

```bash
moai github status
```

**Sample output:**
```
✓ GitHub Actions Runner
  Version: 2.700.1 (10 days old)
  Status: OK

⚠️ Update available: 2.701.0
Run: moai doctor --fix
```

#### Doctor integration

```bash
moai doctor
```

The runner version check is integrated into the system diagnostics.

## Troubleshooting

### PR comment triggers not working

#### Checklist

1. Is the GitHub Actions workflow enabled?
   - Check Repository → Actions → workflows

2. Are the GitHub Secrets configured?
   - Settings → Secrets and variables → Actions

3. Are the workflow permissions correct?
   - `contents: read` and `pull-requests: write` are required

### Per-LLM error handling

#### Claude

**Error:** `CLAUDE_CODE_OAUTH_TOKEN expired`
**Fix:** re-run `moai github auth claude`

#### Codex

**Error:** `repository visibility check failed`
**Cause:** attempting to use Codex on a public repo
**Fix:** switch the repo to private

#### Gemini

**Error:** `GEMINI_API_KEY quota exceeded`
**Fix:** increase the quota in the Google Cloud Console

#### GLM

**Error:** `GLM_API_KEY authentication failed`
**Fix:** check the token in `~/.moai/.env.glm`

## Next steps

- [CLI Reference](/en/workflow-commands/)
- [Workflow Settings Reference](/en/advanced/settings-json/)
- [Security Policy](/en/advanced/security-notes/)
