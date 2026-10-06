# Configuration Schema and Language Fields

Full field reference for project configuration and supported language metadata.

## Project Configuration Fields

| Field | Content |
|-------|---------|
| Project name | Display name for the project |
| Project type | Classification (web_application, mobile_application, cli_tool, etc.) |
| Initialization timestamp | When the project was initialized |
| Language configuration | Conversation, agent prompt, and documentation language settings |
| System version | The project management system version |
| Initialization status | Whether all modules are fully initialized |

## Language Configuration Fields (per supported language)

| Field | Content |
|-------|---------|
| Name | Display name in English |
| Native name | Display name in the native language |
| Code | Language code (e.g., en, ko, ja, zh) |
| Locale | System locale string |
| Agent prompt language | Whether to use English or localized prompts |
| Token cost impact | Percentage increase in token usage for non-English prompts |

## Integrated Configuration Status

The project status includes:

- Project metadata and type classification
- Language configuration and associated costs
- Documentation completion status
- Template optimization results
- Module initialization states

## Language Settings Update Process

When updating language settings, configure:

- conversation_language: language for user-facing responses
- agent_prompt_language: language for internal agent instructions (often kept as English for cost optimization)
- documentation language: language for generated documentation

Updates trigger:

- Configuration file modifications
- Documentation structure updates
- Template localization adjustments

See `.moai/config/sections/language.yaml` for the canonical language configuration file in MoAI projects.
