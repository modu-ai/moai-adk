---
name: moai-project-config-manager
version: 1.0.0
created: 2025-11-05
updated: 2025-11-05
status: active
description: Complete config.json CRUD operations with validation, merge strategy, and error recovery
freedom: medium
type: project
tags: [project, configuration, management, validation, crud]
allowed-tools:
  - Read
  - Write
  - Edit
  - AskUserQuestion
  - Bash
---

# Project Configuration Manager - Skill Guide

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-project-config-manager |
| **Version** | 1.0.0 (2025-11-05) |
| **Core Function** | Configuration file lifecycle management |
| **Scope** | All `.moai/config.json` operations |
| **Freedom** | Medium (guided automation) |

---

## What It Does

**Purpose**: Centralized management of all MoAI project configuration operations with robust validation, error handling, and intelligent merge strategies.

**Key capabilities**:
- ✅ **Complete CRUD**: Create, Read, Update, Delete configuration sections
- ✅ **Validation Engine**: Pre-save validation for all configuration changes
- ✅ **Merge Strategy**: Preserve unmodified sections while updating selected ones
- ✅ **Error Recovery**: Handle missing files, invalid JSON, permission issues
- ✅ **Batch Updates**: Handle multiple setting changes in single operation
- ✅ **Backup & Restore**: Automatic backup before major changes
- ✅ **Interactive Workflows**: User-friendly setting modification with TUI surveys

---

## Core Configuration Structure

The skill manages these main configuration sections:

### 1. Language Settings
```json
"language": {
  "conversation_language": "ko|en|ja|zh",
  "conversation_language_name": "한국어|English|日本語|中文",
  "agent_prompt_language": "english|localized"
}
```

### 2. User Settings
```json
"user": {
  "nickname": "string (max 20 chars)"
}
```

### 3. GitHub Settings
```json
"github": {
  "auto_delete_branches": true|false,
  "spec_git_workflow": "feature_branch|develop_direct|per_spec"
}
```

### 4. Report Generation
```json
"report_generation": {
  "enabled": true|false,
  "auto_create": true|false,
  "user_choice": "Enable|Minimal|Disable"
}
```

### 5. Project Domains
```json
"stack": {
  "selected_domains": ["frontend", "backend", "data", "devops", "security"]
}
```

---

## Workflow: Configuration Management Operations

### Phase 1: Load & Validate

**Always validate before operations**:

```python
# 1. Check file existence
if not Path(".moai/config.json").exists():
    raise ConfigError("Configuration file not found")

# 2. Validate JSON structure
try:
    config = json.loads(Path(".moai/config.json").read_text())
except json.JSONDecodeError as e:
    raise ConfigError(f"Invalid JSON in config.json: {e}")

# 3. Validate required sections
required_sections = ["language", "github", "report_generation"]
for section in required_sections:
    if section not in config:
        raise ConfigError(f"Missing required section: {section}")
```

### Phase 2: Interactive Setting Selection

**Present setting modification options**:

```
## Current Project Settings

✅ **Language**: [current value]
✅ **Nickname**: [current value] 
✅ **Agent Prompt Language**: [current value]
✅ **GitHub Auto-delete Branches**: [current value]
✅ **SPEC Git Workflow**: [current value]
✅ **Report Generation**: [current value]
✅ **Selected Domains**: [current values]

Which settings would you like to modify?
```

**Multi-select options** (AskUserQuestion):
1. "🌍 Language & Agent Prompt Language"
2. "👤 Nickname"
3. "🔧 GitHub Settings" 
4. "📊 Report Generation"
5. "🎯 Project Domains"

### Phase 3: Collect New Values (Batched Questions)

**Language Section** (if selected):
```python
Question 1: "Which conversation language do you prefer?"
Options: ["English", "한국어", "日本語", "中文"]

Question 2: "Which agent prompt language should Alfred use?"
Options: ["English (Global Standard)", "Selected Language (Localized)"]
```

**Nickname Section** (if selected):
```python
Question: "What would you like your nickname to be?"
Type: text_input (max 20 characters)
```

**GitHub Section** (if selected):
```python
Question 1: "Auto-delete branches after merge?"
Options: ["Yes, enable", "No, disable", "Keep current"]

Question 2: "SPEC git workflow preference?"
Options: ["Feature Branch + PR", "Direct Commit to Develop", "Decide per SPEC", "Keep current"]
```

**Report Generation Section** (if selected):
```python
Question: "Report generation preference?"
Options: [
  "📊 Enable (Full reports with 50-60 tokens)",
  "⚡ Minimal (Essential reports with 20-30 tokens)", 
  "🚫 Disable (No automatic reports, 0 tokens)"
]
```

**Project Domains Section** (if selected):
```python
Question: "Select project domains?"
Options: ["frontend", "backend", "data", "devops", "security", "Clear all", "Keep current"]
Multi-select: true
```

### Phase 4: Intelligent Merge & Update

**Update strategy - preserve unmodified sections**:

```python
def merge_config_updates(config, updates):
    """Merge user updates into existing config"""
    new_config = copy.deepcopy(config)
    
    for section, changes in updates.items():
        if section in new_config:
            new_config[section].update(changes)
        else:
            new_config[section] = changes
    
    return new_config
```

**Update mappings**:

**Language Updates**:
```python
language_code_map = {
    "English": "en", "한국어": "ko", "日本語": "ja", "中文": "zh"
}
agent_language_map = {
    "English (Global Standard)": "english",
    "Selected Language (Localized)": "localized"
}
```

**GitHub Updates**:
```python
workflow_map = {
    "Feature Branch + PR": "feature_branch",
    "Direct Commit to Develop": "develop_direct", 
    "Decide per SPEC": "per_spec"
}
```

**Report Updates**:
```python
report_map = {
    "📊 Enable": {"enabled": True, "auto_create": True, "user_choice": "Enable"},
    "⚡ Minimal": {"enabled": True, "auto_create": False, "user_choice": "Minimal"},
    "🚫 Disable": {"enabled": False, "auto_create": False, "user_choice": "Disable"}
}
```

### Phase 5: Validation & Save

**Pre-save validation**:
```python
def validate_config(config):
    """Validate configuration before saving"""
    errors = []
    
    # Validate language codes
    valid_languages = ["en", "ko", "ja", "zh"]
    if config.get("language", {}).get("conversation_language") not in valid_languages:
        errors.append("Invalid conversation language")
    
    # Validate nickname length
    nickname = config.get("user", {}).get("nickname", "")
    if len(nickname) > 20:
        errors.append("Nickname exceeds 20 characters")
    
    # Validate GitHub workflow
    valid_workflows = ["feature_branch", "develop_direct", "per_spec"]
    if config.get("github", {}).get("spec_git_workflow") not in valid_workflows:
        errors.append("Invalid SPEC git workflow")
    
    return errors
```

**Atomic save with backup**:
```python
def save_config_safely(config):
    """Save configuration with automatic backup"""
    # Create backup
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    backup_path = f".moai/config.backup.{timestamp}.json"
    shutil.copy2(".moai/config.json", backup_path)
    
    try:
        # Validate before save
        errors = validate_config(config)
        if errors:
            raise ConfigError(f"Validation failed: {errors}")
        
        # Atomic write
        with open(".moai/config.json.tmp", "w") as f:
            json.dump(config, f, indent=2, ensure_ascii=False)
        
        os.replace(".moai/config.json.tmp", ".moai/config.json")
        
    except Exception as e:
        # Restore from backup on error
        shutil.copy2(backup_path, ".moai/config.json")
        raise ConfigError(f"Save failed, restored backup: {e}")
```

---

## Error Handling & Recovery

### Common Error Scenarios

**1. Missing Configuration File**
```python
if not Path(".moai/config.json").exists():
    # Create default configuration
    default_config = create_default_config()
    save_config_safely(default_config)
    return "Created new configuration file with defaults"
```

**2. Invalid JSON**
```python
try:
    config = json.loads(config_content)
except json.JSONDecodeError:
    # Try to repair common JSON issues
    repaired = attempt_json_repair(config_content)
    if repaired:
        config = json.loads(repaired)
        save_config_safely(config)
    else:
        raise ConfigError("Cannot repair invalid JSON")
```

**3. Permission Issues**
```python
try:
    save_config_safely(config)
except PermissionError:
    # Try to fix permissions
    os.chmod(".moai", 0o755)
    os.chmod(".moai/config.json", 0o644)
    save_config_safely(config)
except Exception as e:
    raise ConfigError(f"Permission denied: {e}")
```

### Validation Rules

**Language Settings**:
- `conversation_language`: Must be one of ["en", "ko", "ja", "zh"]
- `conversation_language_name`: Must match language code
- `agent_prompt_language`: Must be "english" or "localized"

**User Settings**:
- `nickname`: Max 20 characters, no special chars

**GitHub Settings**:
- `auto_delete_branches`: Must be boolean
- `spec_git_workflow`: Must be one of ["feature_branch", "develop_direct", "per_spec"]

**Report Settings**:
- `enabled`: Must be boolean
- `auto_create`: Must be boolean
- `user_choice`: Must be one of ["Enable", "Minimal", "Disable"]

**Domain Settings**:
- `selected_domains`: Must be array of valid domain strings

---

## Usage Examples

### Basic Configuration Update
```python
# Load current configuration
Skill("moai-project-config-manager")

# Interactive workflow:
# 1. Display current settings
# 2. User selects sections to modify
# 3. Collect new values
# 4. Validate and merge changes
# 5. Save with backup
```

### Programmatic Updates
```python
# Direct configuration update (non-interactive)
updates = {
    "language": {
        "conversation_language": "en",
        "conversation_language_name": "English"
    }
}

Skill("moai-project-config-manager", action="update", changes=updates)
```

### Configuration Validation
```python
# Validate existing configuration
result = Skill("moai-project-config-manager", action="validate")
if result.errors:
    print(f"Configuration errors: {result.errors}")
else:
    print("Configuration is valid")
```

### Configuration Backup
```python
# Create manual backup
backup_path = Skill("moai-project-config-manager", action="backup")
print(f"Configuration backed up to: {backup_path}")
```

---

## Best Practices

### 1. Always Validate
- Validate before every save operation
- Check JSON structure, required fields, and value constraints
- Provide clear error messages for validation failures

### 2. Preserve Data
- Always use merge strategy, never overwrite entire config
- Create backups before major changes
- Maintain backward compatibility

### 3. User Experience
- Show current values before asking for changes
- Group related settings together
- Use clear, descriptive option labels
- Provide progress feedback during operations

### 4. Error Recovery
- Automatic backup creation
- Graceful degradation on errors
- Clear error messages with recovery suggestions
- Rollback capability for failed operations

---

## Integration Points

### With Alfred Commands
- **`/alfred:0-project`**: Use for project initialization and setting updates
- **`/alfred:1-plan`**: Access configuration for planning decisions
- **`/alfred:2-run`**: Use configuration during execution
- **`/alfred:3-sync`**: Update configuration based on project changes

### With Other Skills
- **`moai-alfred-ask-user-questions`**: Use for interactive setting collection
- **`moai-skill-factory`**: Integrate with skill configuration management
- **Domain-specific skills**: Respect configuration settings for behavior

---

## Configuration Schema Reference

Complete configuration schema with validation rules:

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "language": {
      "type": "object",
      "properties": {
        "conversation_language": {
          "type": "string",
          "enum": ["en", "ko", "ja", "zh"]
        },
        "conversation_language_name": {
          "type": "string",
          "enum": ["English", "한국어", "日本語", "中文"]
        },
        "agent_prompt_language": {
          "type": "string", 
          "enum": ["english", "localized"]
        }
      },
      "required": ["conversation_language", "conversation_language_name", "agent_prompt_language"]
    },
    "user": {
      "type": "object",
      "properties": {
        "nickname": {
          "type": "string",
          "maxLength": 20,
          "pattern": "^[a-zA-Z0-9가-힣ぁ-ゔ一-龯\\s]+$"
        }
      }
    },
    "github": {
      "type": "object",
      "properties": {
        "auto_delete_branches": {"type": "boolean"},
        "spec_git_workflow": {
          "type": "string",
          "enum": ["feature_branch", "develop_direct", "per_spec"]
        }
      }
    },
    "report_generation": {
      "type": "object",
      "properties": {
        "enabled": {"type": "boolean"},
        "auto_create": {"type": "boolean"},
        "user_choice": {
          "type": "string",
          "enum": ["Enable", "Minimal", "Disable"]
        }
      }
    },
    "stack": {
      "type": "object",
      "properties": {
        "selected_domains": {
          "type": "array",
          "items": {
            "type": "string",
            "enum": ["frontend", "backend", "data", "devops", "security"]
          }
        }
      }
    }
  },
  "required": ["language", "github", "report_generation"]
}
```

---

## Troubleshooting

### Common Issues

**Configuration not saving**:
1. Check file permissions on `.moai/` directory
2. Verify JSON syntax with online validator
3. Check for disk space availability
4. Look for backup files in `.moai/`

**Validation errors**:
1. Review error messages for specific constraint violations
2. Check configuration schema reference
3. Verify all required fields are present
4. Ensure data types match schema

**Merge conflicts**:
1. Recent changes may overwrite manual edits
2. Check backup files for previous versions
3. Use `action="diff"` to see changes before applying
4. Manual edit may be required for complex conflicts

### Debug Mode

Enable debug logging:
```python
Skill("moai-project-config-manager", debug=True)
```

This provides detailed output for:
- Configuration loading steps
- Validation process details
- Merge operation results
- Save operation status
