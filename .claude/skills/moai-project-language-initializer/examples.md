# Examples for moai-project-language-initializer

## Basic Usage Examples

### 1. Complete Project Initialization
```python
# First-time setup with all workflows
Skill("moai-project-language-initializer")

# This will execute:
# 1. Basic batch (language, prompt language, nickname)
# 2. Team mode batch (if mode="team" detected)
# 3. Report generation configuration
# 4. Domain selection for expert agents
```

### 2. Settings Update Only
```python
# Update existing configuration without full initialization
Skill("moai-project-language-initializer", mode="settings")

# Use when changing:
# - Language preferences
# - User nickname
# - Report generation settings
# - Domain selections
```

### 3. Team Mode Configuration
```python
# Configure or update team-specific settings
Skill("moai-project-language-initializer", mode="team_setup")

# Handles:
# - GitHub auto-delete branches
# - Git workflow selection (feature_branch/develop_direct/per_spec)
# - Team collaboration settings
```

## Configuration Examples

### Language Selection Result
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
  }
}
```

### Team Mode Configuration Result
```json
{
  "github": {
    "auto_delete_branches": true,
    "spec_git_workflow": "feature_branch",
    "auto_delete_branches_rationale": "PR 병합 후 원격 브랜치 자동 정리",
    "spec_git_workflow_rationale": "SPEC마다 feature 브랜치 생성으로 팀 리뷰 가능"
  },
  "mode": "team"
}
```

### Report Generation Configuration Result
```json
{
  "report_generation": {
    "enabled": true,
    "auto_create": false,
    "user_choice": "Minimal",
    "warn_user": true,
    "configured_at": "2025-11-05T12:30:00Z"
  }
}
```

### Domain Selection Result
```json
{
  "stack": {
    "selected_domains": ["frontend", "backend", "devops"],
    "domain_selection_date": "2025-11-05T12:30:00Z"
  }
}
```

## Workflow Examples

### English-speaking User Setup
```python
# User selects English and team mode
Skill("moai-project-language-initializer")

# Resulting config.json:
{
  "language": {
    "conversation_language": "en",
    "conversation_language_name": "English",
    "agent_prompt_language": "english"
  },
  "user": {
    "nickname": "Alex",
    "selected_at": "2025-11-05T12:30:00Z"
  },
  "github": {
    "auto_delete_branches": true,
    "spec_git_workflow": "feature_branch"
  },
  "mode": "team",
  "report_generation": {
    "enabled": true,
    "auto_create": false,
    "user_choice": "Minimal",
    "warn_user": true
  },
  "stack": {
    "selected_domains": ["frontend", "backend"]
  }
}
```

### Korean-speaking User Setup (Cost-conscious)
```python
# Korean user who wants to minimize token costs
Skill("moai-project-language-initializer")

# User selects:
# - Korean for conversation
# - English for agent prompts (cost saving)
# - Minimal report generation
# - Individual mode (not team)

# Resulting config.json:
{
  "language": {
    "conversation_language": "ko",
    "conversation_language_name": "한국어",
    "agent_prompt_language": "english"
  },
  "user": {
    "nickname": "김개발",
    "selected_at": "2025-11-05T12:30:00Z"
  },
  "mode": "individual",
  "report_generation": {
    "enabled": true,
    "auto_create": true,
    "user_choice": "Minimal",
    "warn_user": false
  },
  "stack": {
    "selected_domains": ["backend", "database"]
  }
}
```

### Multi-language Team Setup
```python
# Japanese user in an international team
Skill("moai-project-language-initializer")

# Resulting config.json:
{
  "language": {
    "conversation_language": "ja",
    "conversation_language_name": "日本語",
    "agent_prompt_language": "english"
  },
  "user": {
    "nickname": "田中",
    "selected_at": "2025-11-05T12:30:00Z"
  },
  "github": {
    "auto_delete_branches": true,
    "spec_git_workflow": "per_spec"
  },
  "mode": "team",
  "report_generation": {
    "enabled": true,
    "auto_create": false,
    "user_choice": "Enable",
    "warn_user": true
  },
  "stack": {
    "selected_domains": ["frontend", "backend", "devops", "security"]
  }
}
```

## Error Handling Examples

### Invalid Nickname Recovery
```python
# User enters nickname longer than 20 characters
Skill("moai-project-language-initializer")

# Skill will:
# 1. Detect invalid nickname length
# 2. Show validation error
# 3. Re-prompt for nickname with character limit
# 4. Continue with valid nickname
```

### Missing Configuration Recovery
```python
# Config.json is missing or corrupted
Skill("moai-project-language-initializer")

# Skill will:
# 1. Detect missing configuration
# 2. Initialize with default values
# 3. Ask user to confirm or modify defaults
# 4. Create new configuration file
```

### Incompatible Settings Recovery
```python
# User selects conflicting settings
Skill("moai-project-language-initializer")

# Example conflict: Team mode without GitHub settings
# Skill will:
# 1. Detect inconsistency
# 2. Explain the conflict
# 3. Offer resolution options
# 4. Apply consistent configuration
```

## Integration Examples

### With Alfred Commands
```bash
# In /alfred:0-project command
/alfred:0-project
# -> Calls Skill("moai-project-language-initializer") internally
# -> Results stored in .moai/config.json
# -> Continues with project setup based on configuration

# In /alfred:1-plan command
/alfred:1-plan "new feature"
# -> Reads domain selection from config.json
# -> Activates relevant expert agents based on selected domains
```

### With Other Skills
```python
# After language initialization, use domain-specific skills
Skill("moai-project-language-initializer")  # Setup language and domains
# -> User selects "frontend" and "backend" domains

# Later commands automatically use expert agents:
Skill("ui-ux-expert")  # For frontend tasks
Skill("backend-expert")  # For backend tasks
```

## Performance Examples

### Token Cost Comparison
```python
# Expensive setup (Localized prompts + Full reports)
Skill("moai-project-language-initializer")
# User selects: Korean + Localized prompts + Enable reports
# Session tokens: ~300-350
# Cost impact: High

# Cost-effective setup (English prompts + Minimal reports)
Skill("moai-project-language-initializer") 
# User selects: Korean + English prompts + Minimal reports
# Session tokens: ~80-120
# Cost impact: Low (65% savings)
```

### Interaction Efficiency
```python
# Traditional approach (separate questions)
AskUserQuestion(language_question)  # Turn 1
AskUserQuestion(prompt_language_question)  # Turn 2
AskUserQuestion(nickname_question)  # Turn 3
# Total: 3 user interactions

# Batch approach (this skill)
Skill("moai-project-language-initializer")
# Combines 3 questions into 1 interaction
# Total: 1 user interaction (67% reduction)
```

## Testing Examples

### Mock Configuration for Testing
```json
{
  "language": {
    "conversation_language": "ko",
    "conversation_language_name": "한국어",
    "agent_prompt_language": "english"
  },
  "user": {
    "nickname": "테스트유저",
    "selected_at": "2025-11-05T12:30:00Z"
  },
  "mode": "team",
  "github": {
    "auto_delete_branches": true,
    "spec_git_workflow": "feature_branch"
  },
  "report_generation": {
    "enabled": true,
    "auto_create": true,
    "user_choice": "Minimal",
    "warn_user": true
  },
  "stack": {
    "selected_domains": ["frontend", "backend", "devops"]
  }
}
```

### Validation Test Cases
```python
# Test edge cases:
Skill("moai-project-language-initializer", test_case="empty_nickname")
Skill("moai-project-language-initializer", test_case="invalid_language")
Skill("moai-project-language-initializer", test_case="missing_config")
Skill("moai-project-language-initializer", test_case="corrupted_json")
```

These examples demonstrate the skill's flexibility, error handling, cost management features, and integration capabilities with the broader MoAI-ADK ecosystem.
