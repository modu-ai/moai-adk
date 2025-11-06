# Technical Reference for moai-project-language-initializer

## API Reference

### Skill Invocation

```python
Skill("moai-project-language-initializer", mode="full")
Skill("moai-project-language-initializer", mode="settings")
Skill("moai-project-language-initializer", mode="team_setup")
```

#### Parameters

- `mode` (optional): Execution mode
  - `"full"`: Complete initialization (default)
  - `"settings"`: Update existing settings only
  - `"team_setup"`: Configure team-specific settings

### Return Values

The skill returns a dictionary with the configuration results:

```python
{
  "success": True,
  "configuration": {
    "language": {...},
    "user": {...},
    "github": {...},
    "report_generation": {...},
    "stack": {...}
  },
  "changes_made": ["list", "of", "updated", "sections"],
  "warnings": ["any", "configuration", "warnings"]
}
```

## Configuration Schema

### Language Configuration
```json
{
  "language": {
    "conversation_language": "string",  // "en", "ko", "ja", "zh"
    "conversation_language_name": "string",  // "English", "한국어", "日本語", "中文"
    "agent_prompt_language": "string"  // "english", "localized"
  }
}
```

### User Configuration
```json
{
  "user": {
    "nickname": "string",  // Max 20 characters
    "selected_at": "string"  // ISO 8601 timestamp
  }
}
```

### GitHub Configuration
```json
{
  "github": {
    "auto_delete_branches": "boolean",
    "spec_git_workflow": "string",  // "feature_branch", "develop_direct", "per_spec"
    "auto_delete_branches_rationale": "string",
    "spec_git_workflow_rationale": "string"
  }
}
```

### Report Generation Configuration
```json
{
  "report_generation": {
    "enabled": "boolean",
    "auto_create": "boolean",
    "user_choice": "string",  // "Enable", "Minimal", "Disable"
    "warn_user": "boolean",
    "configured_at": "string"  // ISO 8601 timestamp
  }
}
```

### Stack Configuration
```json
{
  "stack": {
    "selected_domains": ["string"],  // Array of domain names
    "domain_selection_date": "string"  // ISO 8601 timestamp
  }
}
```

## Supported Languages

| Code | Language | Conversation | Agent Prompts | Documentation |
|------|----------|--------------|---------------|---------------|
| en | English | English | English | English |
| ko | Korean | 한국어 | English/한국어 | 한국어 |
| ja | Japanese | 日本語 | English/日本語 | 日本語 |
| zh | Chinese | 中文 | English/中文 | 中文 |

## Domain Options

The skill supports the following domains for expert agent activation:

- **frontend**: UI/UX development, React, Vue, CSS, etc.
- **backend**: API development, server architecture, etc.
- **database**: Database design, SQL, migrations, etc.
- **devops**: Infrastructure, CI/CD, deployment, etc.
- **security**: Security analysis, authentication, etc.
- **testing**: Test automation, QA, etc.
- **ml**: Machine learning, data science, etc.
- **mobile**: iOS, Android, React Native, etc.

## Git Workflow Options

### Feature Branch Workflow (`feature_branch`)
- Creates feature branch for each SPEC
- Requires PR review before merging
- Best for team collaboration
- Branch naming: `feature/SPEC-{ID}`

### Direct Commit Workflow (`develop_direct`)
- Commits directly to develop branch
- No PR process
- Best for individual projects
- Faster iteration cycle

### Per-SPEC Decision Workflow (`per_spec`)
- Asks user to choose workflow for each SPEC
- Maximum flexibility
- Good for mixed project types
- Decision recorded per SPEC

## Error Codes and Messages

### Validation Errors

| Error Code | Message | Cause | Resolution |
|------------|---------|-------|------------|
| `INV_NICKNAME_LENGTH` | "Nickname exceeds 20 character limit" | User input longer than 20 chars | Re-prompt with character limit |
| `INV_LANGUAGE_CODE` | "Unsupported language code" | Invalid language selection | Re-prompt with valid options |
| `INV_DOMAIN_SELECTION` | "Invalid domain selection" | User selected unsupported domain | Re-prompt with domain list |
| `INV_WORKFLOW_TYPE` | "Invalid Git workflow type" | Invalid workflow selection | Re-prompt with workflow options |

### Configuration Errors

| Error Code | Message | Cause | Resolution |
|------------|---------|-------|------------|
| `CONFIG_NOT_FOUND` | "Configuration file not found" | .moai/config.json missing | Create new configuration |
| `CONFIG_CORRUPTED` | "Configuration file corrupted" | Invalid JSON format | Restore from backup or recreate |
| `MISSING_REQUIRED_FIELD` | "Required configuration field missing" | Incomplete configuration | Add missing field with default |
| `INCONSISTENT_CONFIG` | "Inconsistent configuration detected" | Conflicting settings | Prompt for resolution |

### System Errors

| Error Code | Message | Cause | Resolution |
|------------|---------|-------|------------|
| `PERMISSION_DENIED` | "Cannot write configuration file" | File system permissions | Check directory permissions |
| `DISK_SPACE_EXHAUSTED` | "Insufficient disk space" | Disk full | Free up disk space |
| `TOOL_UNAVAILABLE` | "Required tool not available" | AskUserQuestion unavailable | Use fallback interaction method |

## Batch Question Patterns

### Basic Batch Question
```json
{
  "type": "batch",
  "questions": [
    {
      "id": "language",
      "question": "Select your preferred conversation language",
      "type": "single_choice",
      "options": ["English", "한국어", "日本어", "中文"]
    },
    {
      "id": "prompt_language",
      "question": "Choose agent prompt language (English saves 15-20% tokens)",
      "type": "single_choice",
      "options": ["English (Recommended)", "Same as conversation language"]
    },
    {
      "id": "nickname",
      "question": "Enter your nickname (max 20 characters)",
      "type": "text_input",
      "validation": {
        "max_length": 20,
        "required": true
      }
    }
  ]
}
```

### Team Mode Batch Question
```json
{
  "type": "batch",
  "trigger": "mode === 'team'",
  "questions": [
    {
      "id": "auto_delete_branches",
      "question": "Automatically delete branches after PR merge?",
      "type": "boolean",
      "default": true
    },
    {
      "id": "git_workflow",
      "question": "Select Git workflow for SPEC development",
      "type": "single_choice",
      "options": [
        "Feature Branch + PR (Recommended for teams)",
        "Direct Commit to Develop (Fast iteration)",
        "Ask me for each SPEC (Flexible)"
      ]
    }
  ]
}
```

### Report Generation Batch Question
```json
{
  "type": "batch",
  "questions": [
    {
      "id": "report_generation",
      "question": "Configure report generation (affects token costs)",
      "type": "single_choice",
      "options": [
        "Enable (50-60 tokens/report, Full analysis)",
        "Minimal (20-30 tokens/report, Essential only)",
        "Disable (0 tokens, No reports)"
      ],
      "warning": "Report generation affects Claude API costs"
    }
  ]
}
```

## Token Cost Analysis

### Report Generation Costs

| Setting | Base Cost | Multiplier | Examples per Session | Total Tokens |
|---------|-----------|------------|---------------------|--------------|
| Enable | 50-60 | 3-5x | product.md, structure.md, tech.md | 150-300 |
| Minimal | 20-30 | 1-2x | Essential reports only | 20-60 |
| Disable | 0 | 0x | No reports | 0 |

### Agent Prompt Language Costs

| Language | Base Multiplier | Additional Cost | Example Session | Total Impact |
|----------|-----------------|-----------------|-----------------|--------------|
| English | 1.0x | 0% | 1000 tokens | Baseline |
| Korean | 1.15x | +15% | 1150 tokens | +150 tokens |
| Japanese | 1.20x | +20% | 1200 tokens | +200 tokens |
| Chinese | 1.18x | +18% | 1180 tokens | +180 tokens |

## Performance Metrics

### Interaction Efficiency
- **Traditional approach**: 6-8 separate questions
- **Batch approach**: 3-4 batched questions
- **Efficiency gain**: 50-60% reduction in user interactions

### Configuration Persistence
- **Write operations**: 1 atomic write per skill execution
- **Read operations**: 1 read on skill load, 1 read on save
- **Backup strategy**: Automatic backup before major changes

### Error Recovery Time
- **Validation errors**: < 1 second recovery
- **Configuration errors**: < 2 seconds recovery
- **System errors**: Fallback to manual configuration

## Integration Hooks

### Pre-Configuration Hook
```python
def pre_config_hook(existing_config):
    # Called before loading configuration
    # Can modify or validate existing config
    return modified_config
```

### Post-Configuration Hook
```python
def post_config_hook(new_config, changes):
    # Called after configuration is saved
    # Can trigger additional setup steps
    trigger_additional_setup(changes)
```

### Validation Hook
```python
def validation_hook(config_section, value):
    # Called during configuration validation
    # Can implement custom validation rules
    return is_valid, error_message
```

## Testing Framework

### Unit Tests
```python
def test_language_validation():
    assert validate_language("en") == True
    assert validate_language("invalid") == False

def test_nickname_validation():
    assert validate_nickname("ValidName") == True
    assert validate_nickname("VeryLongNameThatExceedsLimit") == False
```

### Integration Tests
```python
def test_full_initialization():
    config = run_skill_initialization()
    assert config["language"]["conversation_language"] in SUPPORTED_LANGUAGES
    assert config["user"]["nickname"] is not None
    assert len(config["user"]["nickname"]) <= 20
```

### Performance Tests
```python
def test_interaction_efficiency():
    interactions = count_interactions_for_full_setup()
    assert interactions <= 4  # Should not exceed 4 interactions
```

## Security Considerations

### Input Sanitization
- All user inputs sanitized before storage
- JSON injection prevention
- Cross-site scripting (XSS) prevention in web contexts

### File Permissions
- Configuration files created with secure permissions (600)
- Backup files limited to owner access
- No sensitive information in logs

### Data Privacy
- Nicknames are the only personal data stored
- No API keys or credentials stored in configuration
- Configuration can be easily reviewed and modified

## Migration Guide

### From Previous Version
```python
# Old configuration format
{
  "language": "ko",
  "nickname": "User"
}

# New configuration format (auto-migrated)
{
  "language": {
    "conversation_language": "ko",
    "conversation_language_name": "한국어",
    "agent_prompt_language": "english"
  },
  "user": {
    "nickname": "User",
    "selected_at": "2025-11-05T12:30:00Z"
  }
}
```

### Backup and Restore
```bash
# Create backup
cp .moai/config.json .moai/config.json.backup

# Restore from backup
cp .moai/config.json.backup .moai/config.json
```

This reference provides comprehensive technical details for developers working with the moai-project-language-initializer skill.
