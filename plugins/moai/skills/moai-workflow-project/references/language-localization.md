# Language and Localization Detail

Automatic detection, multilingual documentation structure, and agent prompt localization.

## Automatic Language Detection Process

The system analyzes the project for language indicators:

- File content analysis examining comments and strings
- Configuration file examination for locale settings
- System locale detection
- Directory structure patterns

## Multilingual Documentation Structure

For multiple languages, the system generates:

- Language-specific documentation directories (e.g., `docs/ko` for Korean, `docs/en` for English)
- Language negotiation configuration
- Automatic redirection setup between language versions

## Agent Prompt Localization

The localization system provides:

- Language-specific instructions for agents
- Cultural context adaptations
- Token cost optimization recommendations for multilingual prompts

## Token Cost Impact

Non-English agent prompt languages increase token usage. Typical impact:

| Language | Approximate Token Cost Increase |
|----------|--------------------------------|
| English | baseline (1.0x) |
| Korean | ~1.5x |
| Japanese | ~1.5x |
| Chinese | ~1.3x |

Recommendation: keep `agent_prompt_language: en` for cost optimization while setting `conversation_language` to the user's preferred language. Code comments and documentation language are configurable separately.
