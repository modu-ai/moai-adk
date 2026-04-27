---
title: "Multi-LLM CI Guide"
description: "Automated code reviews with multiple AI models in GitHub Actions"
date: 2026-04-27
draft: false
weight: 10
---

# Multi-LLM CI Guide

Learn how to set up automated code reviews with multiple LLMs in GitHub Actions using MoAI-ADK's Multi-LLM CI feature.

## Overview

### What is Multi-LLM CI?

MoAI-ADK's Multi-LLM CI feature provides an integrated CI/CD pipeline that performs code reviews with multiple AI models simultaneously in GitHub Actions.

### Supported LLMs

| LLM | Provider | Trigger | Features |
|-----|----------|---------|----------|
| **Claude** | Anthropic | `/claude` comment | Issue/PR review, OAuth auth |
| **Codex** | OpenAI | Auto on PR open | ⚠️ Private repos only |
| **Gemini** | Google | Auto on PR open | API Key auth |
| **GLM** | Zhipu AI | Auto on PR open | Token auth |

### User Benefits

- **Simultaneous multi-LL reviews**: Get feedback from multiple LLMs in a single PR
- **Unified management**: Consistent setup via `moai github` CLI
- **Secure authentication**: Dedicated auth handling for each LLM
- **Language detection**: Auto-detects project language and assigns appropriate LLMs

## Getting Started

### Prerequisites

- macOS (arm64) - v1.0 baseline
- Go 1.23+
- GitHub repository
- LLM accounts and API tokens

### Initial Setup

```bash
moai github init
```

This command:
- Creates `.github/workflows/` directory
- Deploys workflow templates
- Deploys composite actions
- Guides GitHub Secrets setup

### LLM Authentication

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

### GitHub Secrets Setup

Required secrets for each LLM:
- `CLAUDE_CODE_OAUTH_TOKEN` - Claude OAuth token
- `CODEX_AUTH_JSON` - Codex auth JSON (base64 encoded)
- `GEMINI_API_KEY` - Gemini API Key
- `GLM_API_KEY` - GLM API Token

### Testing Your First PR

When you create a PR, an LLM Panel comment is automatically added:

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

## LLM Authentication

### Claude Setup

#### OAuth Token Issuance

1. Install [Claude Code](https://claude.ai/download)
2. Login and issue OAuth token
3. Automatically saved to `.claude/settings.local.json`

#### moai github auth claude

```bash
moai github auth claude
```

**Interactive setup process:**
```
Claude OAuth token not found.
Would you like to install Claude Code and login? (y/n): y

[Confirmed] OAuth token saved to settings.local.json.
Set GitHub Secret: CLAUDE_CODE_OAUTH_TOKEN to:
<token-value>
```

### Codex Setup (Private Repos Only)

#### Auth JSON Creation

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
Reading file to generate GitHub Secret...
⚠️ Codex is restricted to private repositories (REQ-SEC-001)

Generated Secret:
CODEX_AUTH_JSON=eyJ0...
```

### Gemini Setup

```bash
moai github auth gemini
```

Enter API key and follow GitHub Secret setup guide.

### GLM Setup

```bash
moai github auth glm
```

Automatically reads from GLM token path (`~/.moai/.env.glm`).

## Workflow Templates

### llm-panel.yml

**Trigger:** PR opened

**Purpose:** Automatically create a panel comment displaying status of each LLM

**Note:** Individual reviews triggered via `/claude`, `/codex`, `/gemini`, `/glm` comments

### claude.yml / claude-code-review.yml

- **claude.yml**: Issue trigger (initial review)
- **claude-code-review.yml**: PR trigger (change review)

**Feature:** Triggered by `/claude` comment only

### codex-review.yml

**Security Constraint:**
- Only runs on `private` repos (REQ-SEC-001)
- Visibility check blocks public repos

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

- Auto language detection (detect-language action)
- Auto-triggered on PR synchronize

### glm-review.yml

- GLM-specific environment setup (setup-glm-env action)
- Automatic environment variable injection

### Composite Actions

#### detect-language

**Input:** repository root path
**Output:** language environment variable (`detected_language`)

**Supported Languages:** Go, Python, TypeScript, JavaScript, Rust, Java, Kotlin, C#, Ruby, PHP, Elixir, C++, Scala, R, Flutter, Swift (16 languages)

#### setup-glm-env

Sets up environment variables for GLM team mode:
- `ANTHROPIC_AUTH_TOKEN` (GLM endpoint)
- `ANTHROPIC_BASE_URL` (https://glm.modu-ai.kr)

## Advanced Configuration

### github-actions.yaml Customization

#### Basic Structure

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

#### Per-Language LLM Assignment

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

### Runner Version Management

#### Automatic Update Check

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

#### Doctor Integration

```bash
moai doctor
```

Runner version check integrated into system diagnostics (T-27).

## Troubleshooting

### PR Comment Triggers Not Working

#### Checklist

1. ✅ GitHub Actions workflow enabled?
   - Repository → Actions → workflows

2. ✅ GitHub Secrets configured?
   - Settings → Secrets and variables → Actions

3. ✅ Workflow permissions correct?
   - Requires `contents: read`, `pull-requests: write`

### LLM-Specific Errors

#### Claude

**Error:** `CLAUDE_CODE_OAUTH_TOKEN expired`
**Fix:** Re-run `moai github auth claude`

#### Codex

**Error:** `repository visibility check failed`
**Cause:** Attempting to use Codex on public repo
**Fix:** Make repository private

#### Gemini

**Error:** `GEMINI_API_KEY quota exceeded`
**Fix:** Increase quota in Google Cloud Console

#### GLM

**Error:** `GLM_API_KEY authentication failed`
**Fix:** Verify token in `~/.moai/.env.glm`

## Next Steps

- [CLI Reference](/docs/commands/)
- [Workflow Configuration](/docs/configuration/)
- [Security Policy](/docs/security/)
