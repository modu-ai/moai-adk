---
name: moai-project-language-initializer
version: 4.0.0
created: 2025-11-05
updated: '2025-11-18'
status: stable
description: Handle comprehensive project language and user setup workflows including
  language selection, agent prompt configuration, user profiles, team settings, and
  domain selection
keywords:
- project
- initialization
- language
- configuration
- batch-questions
- team-setup
allowed-tools:
- Read
- Write
- Edit
- MultiEdit
- Glob
- TodoWrite
- AskUserQuestion
stability: stable
---


# MoAI Project Language & User Initializer

This skill manages the comprehensive project initialization workflow that was previously handled in the 0-project.md command. It extracts the complex batched question patterns into a reusable, efficient skill that reduces user interactions while maintaining full functionality.

## Core Responsibility

Handle all project setup workflows including:
- Language selection (Korean, English, Japanese, Chinese)
- Agent prompt language configuration (English vs Localized)
- User nickname collection (max 20 chars)
- Team mode configuration (GitHub settings, Git workflows)
- Domain selection processes
- Report generation settings with token cost warnings
- MCP server configuration (Figma Access Token setup)

## Usage Patterns

### First-Time Project Initialization

```python
# Complete setup workflow
Skill("moai-project-language-initializer")

# Executes: Basic Batch → Team Mode Batch (if applicable) → Report Generation → Domain Selection → MCP Configuration (if applicable)
```

### Settings Modification

```python
# Update specific settings
Skill("moai-project-language-initializer", mode="settings")
```

### Team Mode Configuration

```python
# Configure team-specific settings
Skill("moai-project-language-initializer", mode="team_setup")
```

## Language Support Matrix

| Language | Code | Conversation Language | Agent Prompt Language | Documentation Language |
|----------|------|---------------------|----------------------|----------------------|
| English | en | English | English (recommended) | English |
| Korean | ko | 한국어 | English/Locale choice | 한국어 |
| Japanese | ja | 日本語 | English/Locale choice | 日本語 |
| Chinese | zh | 中文 | English/Locale choice | 中文 |

## Team Mode Workflows

### Feature Branch + PR Workflow
- **Best for**: Team collaboration, code reviews, audit trails
- **Process**: feature/SPEC-{ID} branch → PR review → develop merge
- **Settings**: `spec_git_workflow: "feature_branch"`

### Direct Commit to Develop Workflow  
- **Best for**: Prototypes, individual projects, rapid iteration
- **Process**: Direct develop commits (no branches)
- **Settings**: `spec_git_workflow: "develop_direct"`

### Per-SPEC Decision Workflow
- **Best for**: Flexible teams, mixed project types
- **Process**: Ask user for each SPEC
- **Settings**: `spec_git_workflow: "per_spec"`

## Token Cost Management

### Report Generation Costs
| Setting | Tokens/Report | Reports/Command | Total Session Tokens | Cost Impact |
|---------|---------------|----------------|-------------------|-------------|
| Enable | 50-60 | 3-5 | 150-300 | Full cost |
| Minimal | 20-30 | 1-2 | 20-60 | 80% reduction |
| Disable | 0 | 0 | 0 | Zero cost |

### Agent Prompt Language Costs
| Setting | Language | Token Efficiency | Cost Impact |
|---------|----------|------------------|-------------|
| English | English | Baseline | Standard |
| Localized | Korean/Japanese/Chinese | 15-20% more tokens | Higher cost |

## Configuration Management

The skill automatically manages `.moai/config/config.json` persistence:

### Basic Configuration Structure
```json
{
  "language": {
    "conversation_language": "ko",
    "conversation_language_name": "한국어", 
    "agent_prompt_language": "localized"
  },
  "user": {
    "nickname": "GOOS",
    "selected_at": "2025-11-05T12:00:00Z"
  }
}
```

### Team Mode Additional Configuration
```json
{
  "github": {
    "auto_delete_branches": true,
    "spec_git_workflow": "feature_branch",
    "auto_delete_branches_rationale": "PR 병합 후 원격 브랜치 자동 정리",
    "spec_git_workflow_rationale": "SPEC마다 feature 브랜치 생성으로 팀 리뷰 가능"
  }
}
```

### Report Generation Configuration
```json
{
  "report_generation": {
    "enabled": true,
    "auto_create": false,
    "user_choice": "Minimal",
    "warn_user": true,
    "configured_at": "2025-11-05T12:00:00Z"
  }
}
```

### Domain Selection Configuration
```json
{
  "stack": {
    "selected_domains": ["frontend", "backend"],
    "domain_selection_date": "2025-11-05T12:00:00Z"
  }
}
```

## Error Handling & Validation

### Input Validation
- Nickname: Max 20 characters, special characters allowed
- Language selection: Must be from supported languages list
- Domain selection: Multi-select with skip option
- Team settings: Boolean and enum validation

### Configuration Validation
- JSON schema validation for config.json
- Backward compatibility checks
- Mode detection validation
- Required field presence checks

### Error Recovery
- Graceful degradation for missing config sections
- Default value application for invalid inputs
- Retry mechanisms for failed batch calls
- Rollback capability for partial configurations

## Integration Points

### With Alfred Commands
- `/alfred:0-project`: Primary integration point
- `/alfred:1-plan`: Uses domain selection for expert activation
- `/alfred:2-run`: Applies language settings to sub-agent prompts
- `/alfred:3-sync`: Respects report generation settings

### With Other Skills
- `moai-core-ask-user-questions`: Uses TUI survey patterns
- `moai-skill-factory`: Can be invoked for skill template application
- `moai-core-agent-guide`: Provides agent lineup based on domains

### Configuration Dependencies
- `.moai/config/config.json`: Primary configuration store
- `mode`: Determines team vs personal workflow
- `github`: Team-specific settings
- `language`: Conversation and prompt language settings

## Best Practices

### For Users
- Choose English for agent prompts to reduce token costs (15-20% savings)
- Enable Minimal report generation for cost-effective operation
- Configure team settings upfront for consistent workflow
- Select relevant domains for expert agent activation

### For Developers
- Use batch patterns to minimize user interactions
- Provide clear token cost warnings before expensive operations
- Validate all inputs before persisting configuration
- Maintain backward compatibility with existing config files

### For Team Collaboration
- Use Feature Branch + PR workflow for code review
- Enable auto-delete branches for repository hygiene
- Select appropriate domains for expert agent routing
- Configure consistent language settings across team
- Set up MCP servers with proper authentication (Figma tokens)

## MCP Server Configuration

### Figma Access Token Setup

When Figma MCP is detected in `.claude/settings.json`, guide users through token configuration:

#### Detection Logic
```python
# Check if Figma MCP is configured
settings_path = Path(".claude/settings.json")
if settings_path.exists():
    settings = json.loads(settings_path.read_text())
    figma_configured = "mcpServers" in settings and "figma" in settings["mcpServers"]
```

#### Token Setup Workflow
1. **Verify Figma MCP Installation**: Check for Figma in mcpServers configuration
2. **Guide Token Creation**: Direct user to Figma developer portal
3. **Secure Token Storage**: Configure environment variable or `.env` file
4. **Validation**: Test Figma MCP connectivity

#### User Guidance Messages
```
🔐 Figma Access Token Setup Required

Your project has Figma MCP configured, but needs an access token:

Steps:
1. Visit: https://www.figma.com/developers/api#access-tokens
2. Create a new access token
3. Choose storage method:
   - Environment variable (recommended): export FIGMA_ACCESS_TOKEN=your_token
   - .env file: Add FIGMA_ACCESS_TOKEN=your_token to .env
   - Shell profile: Add to ~/.zshrc or ~/.bashrc

4. Restart Claude Code to activate token
```

#### Token Validation
```python
def validate_figma_token():
    """Test Figma MCP connectivity with token"""
    # Try to access Figma files via MCP
    # Return success/failure with guidance
```

### MCP Server Status Checking

Provide users with current MCP server status:

```python
def check_mcp_status():
    """Check all configured MCP servers"""
    servers = {
        "context7": check_context7_mcp(),
        "figma": check_figma_mcp(),
        "playwright": check_playwright_mcp()
    }
    return servers
```

## Implementation Notes

This skill extracts and consolidates the complex initialization logic from the original 0-project.md command (~800 lines) into a focused, reusable skill (~500 lines) while maintaining:

- **Full functionality**: All original features preserved
- **UX improvements**: Batch calling patterns maintained
- **Error handling**: Comprehensive validation and recovery
- **Integration**: Seamless compatibility with existing workflows
- **Performance**: Optimized configuration management

The skill serves as a foundation for project initialization and can be extended with additional configuration patterns as needed.
