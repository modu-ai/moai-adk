---
source: https://opencode.ai/docs/models/
fetched: 2026-01-08
title: Models Configuration
---

# OpenCode Models Documentation

## Overview

OpenCode explains how to configure LLM providers and models, supporting 75+ providers through the AI SDK and Models.dev integration.

## Providers

OpenCode preloads popular providers by default. Users can add credentials via the `/connect` command to enable additional providers.

```bash
/connect
```

## Model Selection

Access model selection through the `/models` command after configuring your provider:

```bash
/models
```

## Recommended Models

The following models are recommended for code generation and tool calling:

| Model | Provider |
|-------|----------|
| GPT 5.2 | OpenAI |
| GPT 5.1 Codex | OpenAI |
| Claude Opus 4.5 | Anthropic |
| Claude Sonnet 4.5 | Anthropic |
| Minimax M2.1 | Minimax |
| Gemini 3 Pro | Google |

## Default Configuration

Set default models via the config file using the format `provider_id/model_id`:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "model": "anthropic/claude-sonnet-4",
  "small_model": "anthropic/claude-haiku-3-5"
}
```

For local models:

```json
{
  "model": "lmstudio/google/gemma-3n-e4b"
}
```

## Model Options Configuration

### Global Configuration

```json
{
  "$schema": "https://opencode.ai/config.json",
  "model": "openai/gpt-5",
  "provider": {
    "openai": {
      "models": {
        "gpt-5": {
          "options": {
            "reasoningEffort": "high"
          }
        }
      }
    }
  }
}
```

### Agent-Specific Configuration

```json
{
  "agent": {
    "deep-thinker": {
      "model": "openai/gpt-5",
      "modelOptions": {
        "reasoningEffort": "xhigh"
      }
    }
  }
}
```

## Variants System

Models support multiple configuration variants for different use cases.

### Built-in Variants

**Anthropic:**
- `high` (default): Standard thinking budget
- `max`: Maximum thinking budget

**OpenAI:**
- `none`: No reasoning
- `low`: Low reasoning effort
- `medium`: Medium reasoning effort
- `high`: High reasoning effort (default)
- `xhigh`: Extra high reasoning effort

**Google:**
- `low`: Low effort level
- `high`: High effort level

### Custom Variants

Override existing variants or create custom ones:

```json
{
  "provider": {
    "anthropic": {
      "models": {
        "claude-sonnet-4": {
          "variants": {
            "quick": {
              "options": {
                "thinkingBudget": 1000
              }
            },
            "deep": {
              "options": {
                "thinkingBudget": 50000
              }
            }
          }
        }
      }
    }
  }
}
```

### Switching Variants

Use the `variant_cycle` keybind to switch between variants during a session.

## Model Loading Priority

Models are selected in the following order:

1. **Command-line flag**: `--model` or `-m`
2. **Config file**: `model` field in `opencode.json`
3. **Last used model**: Persisted from previous session
4. **Internal default**: Built-in priority list

## Examples

### Using Multiple Models

```json
{
  "$schema": "https://opencode.ai/config.json",
  "model": "anthropic/claude-sonnet-4",
  "small_model": "anthropic/claude-haiku-3-5",
  "agent": {
    "complex-tasks": {
      "model": "anthropic/claude-opus-4"
    },
    "quick-tasks": {
      "model": "anthropic/claude-haiku-3-5"
    }
  }
}
```

### Local Model Configuration

```json
{
  "$schema": "https://opencode.ai/config.json",
  "model": "ollama/llama2",
  "provider": {
    "ollama": {
      "npm": "@ai-sdk/openai-compatible",
      "options": {
        "baseURL": "http://localhost:11434/v1"
      }
    }
  }
}
```
