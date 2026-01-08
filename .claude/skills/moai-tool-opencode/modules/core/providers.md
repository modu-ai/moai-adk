---
source: https://opencode.ai/docs/providers/
fetched: 2026-01-08
title: Providers
---

# OpenCode Providers Documentation

## Overview

OpenCode supports **75+ LLM providers** through the AI SDK and Models.dev, including local model execution. Setting up a provider requires two steps:

1. Add API keys via the `/connect` command (stored in `~/.local/share/opencode/auth.json`)
2. Configure the provider in your OpenCode config file

## Base URL Customization

Users can override default endpoints:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {
    "anthropic": {
      "options": {
        "baseURL": "https://api.anthropic.com/v1"
      }
    }
  }
}
```

## OpenCode Zen

OpenCode Zen provides curated, tested models from the OpenCode team. Setup requires:

1. Run `/connect`, select opencode, visit opencode.ai/auth
2. Sign in and copy your API key
3. Use `/models` to view recommendations

## Provider Directory

### Amazon Bedrock

- Requires Model Catalog access
- Authentication via AWS credentials, profiles, or bearer tokens
- Supports VPC endpoints with custom endpoint configuration
- Priority: Bearer Token > AWS Credential Chain

```json
{
  "provider": {
    "amazon-bedrock": {
      "region": "us-west-2",
      "profile": "production",
      "endpoint": "https://bedrock-runtime.us-west-2.amazonaws.com"
    }
  }
}
```

### Anthropic

- Recommends Claude Pro/Max subscription
- Supports OAuth flow or manual API key entry
- Available via `/connect` command

```json
{
  "provider": {
    "anthropic": {
      "options": {
        "apiKey": "{env:ANTHROPIC_API_KEY}"
      }
    }
  }
}
```

### Azure OpenAI

- Requires resource creation and model deployment
- Deployment names must match model names
- Set `AZURE_RESOURCE_NAME` environment variable

**Note:** "I'm sorry" errors may indicate content filter issues—switch from DefaultV2 to Default.

```json
{
  "provider": {
    "azure-openai": {
      "options": {
        "resourceName": "{env:AZURE_RESOURCE_NAME}",
        "apiKey": "{env:AZURE_OPENAI_API_KEY}"
      }
    }
  }
}
```

### Google Vertex AI

- Requires Google Cloud project with Vertex AI API enabled
- Key environment variables: `GOOGLE_CLOUD_PROJECT`, `VERTEX_LOCATION`
- Supports service account authentication
- "Global" region recommended for availability

```json
{
  "provider": {
    "google-vertex": {
      "options": {
        "project": "{env:GOOGLE_CLOUD_PROJECT}",
        "location": "us-central1"
      }
    }
  }
}
```

### Local Models (Ollama, LM Studio, llama.cpp)

Custom provider configuration example:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {
    "ollama": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "Ollama (local)",
      "options": {
        "baseURL": "http://localhost:11434/v1"
      },
      "models": {
        "llama2": {
          "name": "Llama 2"
        }
      }
    }
  }
}
```

### Helicone

Observability platform supporting 17+ providers with automatic routing. Optional custom configuration:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {
    "helicone": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "Helicone",
      "options": {
        "baseURL": "https://ai-gateway.helicone.ai",
        "headers": {
          "Helicone-Cache-Enabled": "true",
          "Helicone-User-Id": "opencode"
        }
      }
    }
  }
}
```

Common headers include:
- `Helicone-Cache-Enabled`
- `Helicone-User-Id`
- `Helicone-Property-[Name]`
- `Helicone-Prompt-Id`
- `Helicone-Session-Id`

### OpenRouter & ZenMux

Support custom model additions via config:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {
    "openrouter": {
      "models": {
        "somecoolnewmodel": {}
      }
    }
  }
}
```

OpenRouter supports provider routing:

```json
{
  "provider": {
    "openrouter": {
      "models": {
        "anthropic/claude-sonnet-4": {
          "options": {
            "order": ["baseten"],
            "allow_fallbacks": false
          }
        }
      }
    }
  }
}
```

### Vercel AI Gateway

```json
{
  "provider": {
    "vercel": {
      "models": {
        "anthropic/claude-sonnet-4": {
          "options": {
            "order": ["anthropic", "vertex"]
          }
        }
      }
    }
  }
}
```

Routing options: `order`, `only`, `zeroDataRetention`.

## Custom Provider (OpenAI-Compatible)

1. Run `/connect` and select "Other"
2. Enter unique provider ID
3. Enter API key
4. Configure in `opencode.json`:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {
    "myprovider": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "Display Name",
      "options": {
        "baseURL": "https://api.myprovider.com/v1",
        "apiKey": "{env:CUSTOM_API_KEY}",
        "headers": {
          "Authorization": "Bearer token"
        }
      },
      "models": {
        "model-id": {
          "name": "Display Name",
          "limit": {
            "context": 200000,
            "output": 65536
          }
        }
      }
    }
  }
}
```

Key options: `baseURL`, `apiKey`, `headers`, and model `limit` (context/output tokens).

## Additional Providers Listed

- Amazon Bedrock
- Anthropic
- Azure OpenAI
- Azure Cognitive Services
- Baseten
- Cerebras
- Cloudflare AI Gateway
- Cortecs
- DeepSeek
- Deep Infra
- Fireworks AI
- GitHub Copilot
- Google Vertex AI
- Groq
- Hugging Face
- Helicone
- llama.cpp
- IO.NET
- LM Studio
- Moonshot AI
- MiniMax
- Nebius Token Factory
- Ollama
- Ollama Cloud
- OpenAI
- OpenRouter
- SAP AI Core
- OVHcloud AI Endpoints
- Together AI
- Venice AI
- Vercel AI Gateway
- xAI
- Z.AI
- ZenMux

## Troubleshooting

- Verify credentials: `opencode auth list`
- Confirm provider ID matches across `/connect` and config
- Validate npm package (use provider-specific or `@ai-sdk/openai-compatible`)
- Check `baseURL` correctness
