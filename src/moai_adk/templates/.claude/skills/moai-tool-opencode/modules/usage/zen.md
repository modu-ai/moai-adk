---
source: https://opencode.ai/docs/zen/
fetched: 2026-01-08
title: OpenCode Zen
---

# OpenCode Zen Documentation

## Overview

OpenCode Zen is "a list of tested and verified models provided by the OpenCode team." It operates as an optional AI gateway offering curated models vetted for coding agent performance.

## Getting Started

### Authentication

1. Visit OpenCode Zen and authenticate
2. Add billing details
3. Retrieve your API key
4. Connect via the `/connect` command in the TUI

```bash
opencode
/connect
# Select "opencode"
# Enter your API key
```

## Available Models

The platform provides access to multiple model families:

### Premium Models
- GPT 5 series
- Claude variants (Sonnet, Haiku, Opus)
- Gemini options

### Open Source Models
- Qwen3
- Kimi

### Free Models (Beta)
- Grok Code
- GLM 4.7
- MiniMax M2.1
- Big Pickle

**Note:** Free models are available during beta periods.

## Pricing Structure

Pay-as-you-go model charging per 1M tokens:

| Model Tier | Input Token Price |
|------------|-------------------|
| Economy | $0.40 per 1M tokens |
| Standard | $3.00 per 1M tokens |
| Premium | $15.00 per 1M tokens |

Exact pricing varies by model.

## Team Features

### Workspaces

Workspaces support role-based access:

| Role | Capabilities |
|------|--------------|
| Admin | Manage models, members, and billing |
| Member | Control only personal API keys |

### Admin Controls

Admins can:
- Disable specific models
- Set spending limits per team member
- Manage workspace billing

### Bring Your Own Key (BYOK)

Users can integrate personal API keys while accessing other Zen models:

1. Add your OpenAI or Anthropic API key
2. Direct billing goes to your provider account
3. Access other Zen models with Zen billing

## Privacy & Data Handling

### Data Retention Policies

| Provider | Retention |
|----------|-----------|
| Most models | Zero retention |
| OpenAI | 30-day retention |
| Free-tier models (beta) | May use data for improvement |

### Hosting Location

All models are hosted in the US.

## Configuration

To use Zen models in your config:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "model": "opencode/claude-sonnet-4"
}
```

## Best Practices

1. Start with recommended models from `/models`
2. Monitor usage through the Zen dashboard
3. Set spending limits for team members
4. Review data retention policies for sensitive projects
