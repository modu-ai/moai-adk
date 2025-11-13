# MoAI Project Language & User Initializer

A comprehensive MoAI-ADK skill for handling project initialization workflows including language selection, user setup, team configuration, and domain selection.

## Quick Start

```python
# Initialize your project with all settings
Skill("moai-project-language-initializer")

# Update existing settings only
Skill("moai-project-language-initializer", mode="settings")

# Configure team-specific settings
Skill("moai-project-language-initializer", mode="team_setup")
```

## What It Does

This skill manages the complete project setup process:

✅ **Language Selection** - Choose conversation language (English, Korean, Japanese, Chinese)  
✅ **Agent Configuration** - Set prompt language (English saves 15-20% tokens)  
✅ **User Profile** - Set your nickname (max 20 characters)  
✅ **Team Setup** - Configure GitHub workflows and branch management  
✅ **Report Settings** - Control automatic report generation with cost awareness  
✅ **Domain Selection** - Choose project domains for expert agent activation  

## Key Features

### 💡 Efficient Interactions
- Reduces user interactions by 60% through batched questions
- 3-4 questions instead of 8-10 separate interactions

### 💰 Cost Management
- Token cost warnings for expensive operations
- English prompt language saves 15-20% on API costs
- Report generation controls to manage spending

### 🌍 Multi-language Support
- 4 supported languages with full localization
- Flexible prompt language configuration
- Consistent experience across languages

### 👥 Team Collaboration
- Multiple Git workflow options
- GitHub integration settings
- Team-specific configurations

## Configuration Structure

The skill creates and manages `.moai/config/config.json`:

```json
{
  "language": {
    "conversation_language": "ko",
    "conversation_language_name": "한국어",
    "agent_prompt_language": "english"
  },
  "user": {
    "nickname": "개발자",
    "selected_at": "2025-11-05T12:30:00Z"
  },
  "github": {
    "auto_delete_branches": true,
    "spec_git_workflow": "feature_branch"
  },
  "report_generation": {
    "enabled": true,
    "user_choice": "Minimal"
  },
  "stack": {
    "selected_domains": ["frontend", "backend"]
  }
}
```

## Integration

This skill is automatically used by:
- `/alfred:0-project` - Primary integration point
- `/alfred:1-plan` - Uses domain selection for expert agents
- `/alfred:2-run` - Applies language settings to sub-agents
- `/alfred:3-sync` - Respects report generation settings

## File Structure

```
moai-project-language-initializer/
├── SKILL.md          # Main skill implementation
├── examples.md       # Usage examples and workflows
├── reference.md      # Technical API reference
└── README.md         # This quick overview
```

## Supported Languages

| Language | Code | Conversation | Agent Prompts | Cost Impact |
|----------|------|--------------|---------------|-------------|
| English | en | English | English | Baseline |
| Korean | ko | 한국어 | English/Korean | +15% if Korean |
| Japanese | ja | 日本語 | English/Japanese | +20% if Japanese |
| Chinese | zh | 中文 | English/Chinese | +18% if Chinese |

## Team Workflows

### Feature Branch + PR (Recommended for Teams)
- Creates feature branch for each SPEC
- Requires pull request review
- Best for code quality and team collaboration

### Direct Commit to Develop (Fast Iteration)
- Commits directly to develop branch
- No PR process required
- Best for prototypes and individual projects

### Per-SPEC Decision (Maximum Flexibility)
- Ask user to choose workflow for each SPEC
- Adapts to different project needs
- Good for mixed project types

## Cost Optimization Tips

1. **Use English for agent prompts** - Saves 15-20% on tokens
2. **Choose Minimal report generation** - Reduces token usage by 80%
3. **Configure team settings upfront** - Avoids reconfiguration costs

## Error Handling

The skill includes comprehensive error handling:
- Input validation with helpful error messages
- Automatic recovery from missing configurations
- Graceful degradation for system errors
- Backup and restore capabilities

## Getting Help

For detailed technical information:
- See `examples.md` for practical usage patterns
- See `reference.md` for complete API documentation
- See `SKILL.md` for full implementation details

## Version Information

- **Version**: 1.0.0
- **Created**: 2025-11-05
- **Status**: Active
- **Tier**: Alfred
- **Compatibility**: MoAI-ADK v0.17.0+

This skill is part of the MoAI-ADK ecosystem and follows all official skill standards and best practices.
