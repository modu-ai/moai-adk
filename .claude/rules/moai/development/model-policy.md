---
paths: "**/.claude/agents/**"
---

# Model Policy

Rules for agent model field values and multi-model architecture.

## Valid Model Field Values

Agent definition `model` field accepts only these values:
- inherit: Uses parent session's model (default)
- opus: Claude Opus (highest capability)
- sonnet: Claude Sonnet (balanced)
- haiku: Claude Haiku (fastest, lowest cost)

Invalid values (NEVER use):
- glm: Not a model field value (GLM is configured via environment variables)
- high/medium/low: These are CLI policy flags, not model field values

## Model Policy Tiers

Model policy is set via `moai init --model-policy <tier>`:

| Tier | Description | Opus Agents | Sonnet Agents | Haiku Agents |
|------|-------------|-------------|---------------|--------------|
| high | Maximum quality | spec, strategy, security | backend, frontend, ddd, tdd | quality, git, researcher |
| medium | Balanced (default) | spec, strategy, security | backend, frontend, ddd, tdd | quality, git, researcher |
| low | Cost optimized | None | spec, strategy, security | All others |

## CG Mode

CG Mode (Claude + GLM) uses environment variable overrides, not model field changes:
- Leader session: Uses Claude models (no GLM env)
- Teammate sessions: Inherit GLM env from tmux session
- Activation: `moai cg` (requires tmux)

## Rules

- Agent `model` field must be one of: inherit, opus, sonnet, haiku
- GLM is configured via env vars in settings.json, never via model field
- Model policy tier is a CLI concern, not an agent definition concern
- CG Mode uses tmux session-level env isolation for model routing
