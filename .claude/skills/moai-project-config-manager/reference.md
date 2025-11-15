# Project Configuration Manager - Reference Documentation

## Core API Reference

### Skill Invocation

```python
Skill("moai-project-config-manager")
```

**Parameters**:
- `action` (optional): Operation type
  - `"interactive"` (default): Full interactive configuration workflow
  - `"update"`: Programmatic configuration update
  - `"validate"`: Validate existing configuration
  - `"backup"`: Create configuration backup
  - `"restore"`: Restore from backup
  - `"diff"`: Show changes before applying

- `changes` (optional): Dictionary of configuration changes for non-interactive updates
- `backup_path` (optional): Specific backup file to restore from
- `debug` (optional): Enable detailed logging (boolean)

### Configuration File Path

```
.moai/config.json
```

### Backup File Pattern

```
.moai/config.backup.YYYYMMDD_HHMMSS.json
```

---

## Configuration Schema

### Complete Schema Definition

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "language": {
      "type": "object",
      "description": "Language settings for conversation and agent prompts",
      "properties": {
        "conversation_language": {
          "type": "string",
          "enum": ["en", "ko", "ja", "zh"],
          "description": "ISO language code for user conversations"
        },
        "conversation_language_name": {
          "type": "string", 
          "enum": ["English", "한국어", "日本語", "中文"],
          "description": "Display name for conversation language"
        },
        "agent_prompt_language": {
          "type": "string",
          "enum": ["english", "localized"],
          "description": "Language for Alfred agent prompts"
        }
      },
      "required": ["conversation_language", "conversation_language_name", "agent_prompt_language"]
    },
    
    "user": {
      "type": "object",
      "description": "User-specific settings",
      "properties": {
        "nickname": {
          "type": "string",
          "maxLength": 20,
          "pattern": "^[a-zA-Z0-9가-힣ぁ-ゔ一-龯\\s]+$",
          "description": "User nickname for display in reports and interactions"
        }
      }
    },
    
    "github": {
      "type": "object", 
      "description": "GitHub integration settings",
      "properties": {
        "auto_delete_branches": {
          "type": "boolean",
          "description": "Automatically delete feature branches after PR merge"
        },
        "spec_git_workflow": {
          "type": "string",
          "enum": ["feature_branch", "develop_direct", "per_spec"],
          "description": "Git workflow strategy for SPEC development"
        }
      }
    },
    
    "report_generation": {
      "type": "object",
      "description": "Automatic report generation settings",
      "properties": {
        "enabled": {
          "type": "boolean",
          "description": "Enable/disable automatic report generation"
        },
        "auto_create": {
          "type": "boolean", 
          "description": "Create full (true) or minimal (false) reports"
        },
        "user_choice": {
          "type": "string",
          "enum": ["Enable", "Minimal", "Disable"],
          "description": "User's preferred report generation level"
        },
        "updated_at": {
          "type": "string",
          "format": "date-time",
          "description": "Timestamp of last report setting update"
        }
      }
    },
    
    "stack": {
      "type": "object",
      "description": "Technology stack and domain settings",
      "properties": {
        "selected_domains": {
          "type": "array",
          "items": {
            "type": "string",
            "enum": ["frontend", "backend", "data", "devops", "security"]
          },
          "description": "Selected project domains for specialization"
        },
        "domain_selection_date": {
          "type": "string",
          "format": "date-time", 
          "description": "Timestamp of last domain selection update"
        }
      }
    }
  },
  "required": ["language", "github", "report_generation"]
}
```

### Field Value Mappings

#### Language Code Mapping
| Display Name | Code | Description |
|--------------|------|-------------|
| English | en | Global standard language |
| 한국어 | ko | Korean language |
| 日本語 | ja | Japanese language |
| 中文 | zh | Chinese language |

#### Agent Prompt Language Mapping
| Display Name | Code | Description |
|--------------|------|-------------|
| English (Global Standard) | english | All prompts in English for stability |
| Selected Language (Localized) | localized | Prompts in user's conversation language |

#### Git Workflow Mapping
| Display Name | Code | Description |
|--------------|------|-------------|
| Feature Branch + PR | feature_branch | Create feature branch → PR → develop merge |
| Direct Commit to Develop | develop_direct | Commit directly to develop branch |
| Decide per SPEC | per_spec | Choose workflow per SPEC |

#### Report Generation Mapping
| Display Name | Settings | Description |
|--------------|----------|-------------|
| 📊 Enable | enabled=true, auto_create=true | Full reports with 50-60 tokens |
| ⚡ Minimal | enabled=true, auto_create=false | Essential reports with 20-30 tokens |
| 🚫 Disable | enabled=false, auto_create=false | No automatic reports, 0 tokens |

---

## Error Codes and Handling

### Error Types

| Error Code | Description | Recovery Action |
|------------|-------------|-----------------|
| CONFIG_NOT_FOUND | `.moai/config/config.json` doesn't exist | Create default config |
| INVALID_JSON | Configuration file has invalid JSON syntax | Attempt JSON repair or recreate |
| VALIDATION_FAILED | Configuration fails schema validation | Fix validation errors |
| PERMISSION_DENIED | Cannot write to configuration file | Fix file permissions |
| BACKUP_FAILED | Cannot create backup file | Check disk space and permissions |
| RESTORE_FAILED | Cannot restore from backup | Verify backup file integrity |

### Validation Error Details

#### Language Validation Errors
- `INVALID_LANGUAGE_CODE`: conversation_language not in allowed values
- `MISMATCHED_LANGUAGE_NAME`: conversation_language_name doesn't match code
- `INVALID_AGENT_LANGUAGE`: agent_prompt_language not in allowed values

#### User Validation Errors  
- `NICKNAME_TOO_LONG`: nickname exceeds 20 characters
- `INVALID_NICKNAME_CHARS`: nickname contains invalid characters

#### GitHub Validation Errors
- `INVALID_WORKFLOW`: spec_git_workflow not in allowed values
- `INVALID_BOOLEAN_SETTING`: auto_delete_branches not boolean

#### Report Validation Errors
- `INVALID_REPORT_CHOICE`: user_choice not in allowed values
- `INCONSISTENT_REPORT_SETTINGS`: enabled/auto_create don't match user_choice

#### Domain Validation Errors
- `INVALID_DOMAIN`: selected_domains contains invalid domain name
- `INVALID_DOMAIN_TYPE`: selected_domains is not an array

---

## Operation Details

### Interactive Mode Workflow

1. **Configuration Loading**
   - Load existing `.moai/config/config.json`
   - Validate JSON structure and schema
   - Display current settings

2. **Setting Selection**
   - Present 5 setting groups for modification
   - User selects one or more groups (multi-select)
   - Proceed only if at least one group selected

3. **Value Collection**
   - Ask questions only for selected groups
   - Batch related questions together
   - Map user responses to configuration values

4. **Merge and Validation**
   - Merge new values with existing configuration
   - Validate updated configuration
   - Create automatic backup

5. **Save Operation**
   - Atomic write with temporary file
   - Replace original file on success
   - Restore from backup on failure

### Programmatic Update Mode

```python
# Example: Update language settings
updates = {
    "language": {
        "conversation_language": "en",
        "conversation_language_name": "English", 
        "agent_prompt_language": "english"
    }
}

result = Skill(
    "moai-project-config-manager",
    action="update",
    changes=updates
)

# Returns:
# {
#     "success": True,
#     "changes_applied": ["language.conversation_language", ...],
#     "backup_path": ".moai/config.backup.20251105_123456.json",
#     "validation_errors": []
# }
```

### Validation Mode

```python
result = Skill("moai-project-config-manager", action="validate")

# Returns:
# {
#     "valid": False,
#     "errors": [
#         {
#             "field": "language.conversation_language",
#             "message": "Invalid language code: 'fr'",
#             "allowed_values": ["en", "ko", "ja", "zh"]
#         }
#     ],
#     "warnings": [
#         {
#             "field": "user.nickname",
#             "message": "Nickname is empty"
#         }
#     ]
# }
```

### Backup and Restore

```python
# Create backup
backup_result = Skill("moai-project-config-manager", action="backup")
# Returns: {"backup_path": ".moai/config.backup.20251105_123456.json"}

# List available backups
backups = Skill("moai-project-config-manager", action="list_backups")
# Returns: {".moai/config.backup.20251105_123456.json": "2025-11-05 12:34:56", ...}

# Restore from backup
restore_result = Skill(
    "moai-project-config-manager",
    action="restore", 
    backup_path=".moai/config.backup.20251105_123456.json"
)
# Returns: {"success": True, "restored_from": "..."}
```

### Diff Mode

```python
diff_result = Skill(
    "moai-project-config-manager",
    action="diff",
    changes={"language": {"conversation_language": "en"}}
)

# Returns:
# {
#     "changes": [
#         {
#             "path": "language.conversation_language",
#             "old": "ko",
#             "new": "en",
#             "type": "value_change"
#         }
#     ],
#     "summary": "1 field will be updated"
# }
```

---

## Integration Examples

### With Alfred Commands

**Project Initialization**:
```python
# In /alfred:0-project command
if not Path(".moai/config/config.json").exists():
    Skill("moai-project-config-manager", action="create_default")
```

**Setting Modification**:
```python
# In /alfred:0-project setting subcommand
Skill("moai-project-config-manager")  # Interactive mode
```

**Configuration Validation**:
```python
# In any Alfred command before using config
validation = Skill("moai-project-config-manager", action="validate")
if not validation["valid"]:
    print(f"Configuration errors: {validation['errors']}")
    return
```

### With Other Skills

**Language-Aware Skills**:
```python
config = load_config(".moai/config/config.json")
language = config["language"]["conversation_language"]

if language == "ko":
    # Use Korean responses
    response = generate_korean_response(prompt)
else:
    # Use English responses  
    response = generate_english_response(prompt)
```

**Report Generation Skills**:
```python
config = load_config(".moai/config/config.json")
report_settings = config["report_generation"]

if report_settings["enabled"]:
    if report_settings["auto_create"]:
        generate_full_report()
    else:
        generate_minimal_report()
```

---

## Debug and Troubleshooting

### Debug Mode Usage

```python
# Enable debug logging
result = Skill("moai-project-config-manager", debug=True)

# Debug output includes:
# - Configuration loading steps
# - Validation process details  
# - Merge operation results
# - File operation details
# - Error stack traces
```

### Common Debug Scenarios

**Configuration Loading Issues**:
```debug
DEBUG: Loading configuration from .moai/config.json
DEBUG: File size: 2048 bytes
DEBUG: JSON parsing successful
DEBUG: Schema validation passed
```

**Validation Failures**:
```debug
DEBUG: Validating configuration
ERROR: Field 'language.conversation_language' has invalid value 'fr'
ERROR: Allowed values: ['en', 'ko', 'ja', 'zh']
DEBUG: Validation failed with 1 errors
```

**File Operation Issues**:
```debug
DEBUG: Creating backup .moai/config.backup.20251105_123456.json
DEBUG: Writing temporary file .moai/config.json.tmp
DEBUG: Atomic replace operation
ERROR: Permission denied when writing .moai/config.json
DEBUG: Attempting to restore from backup
```

### Manual Configuration Repair

If the skill cannot automatically repair issues:

1. **JSON Syntax Errors**:
   ```bash
   # Validate JSON syntax
   python3 -m json.tool .moai/config.json
   
   # If valid, backup and manually edit
   cp .moai/config.json .moai/config.backup.manual.json
   # Edit .moai/config.json with proper JSON
   ```

2. **Schema Validation**:
   ```bash
   # Use online JSON schema validator
   # Upload .moai/config.json and the schema from reference.md
   ```

3. **Permission Issues**:
   ```bash
   # Fix directory permissions
   chmod 755 .moai/
   chmod 644 .moai/config.json
   ```

---

## Performance Considerations

### Configuration Loading
- Average time: < 10ms for typical configuration files
- Memory usage: < 1MB for configuration in memory
- File I/O: Optimized with single read/write operations

### Validation Performance
- Schema validation: < 5ms for typical configurations
- Custom validation rules: < 2ms per rule
- Batch validation: Validated all sections in parallel

### Backup Operations
- Backup creation: < 50ms for typical configurations
- Backup storage: Uses copy-on-write for efficiency
- Cleanup: Automatic cleanup of backups older than 30 days

### Recommended Limits
- Configuration file size: < 100KB recommended
- Number of backup files: < 50 recommended
- Concurrent operations: 1 at a time (file locking)

---

## Security Considerations

### File Access Security
- Configuration files should have restrictive permissions (644)
- Backup files inherit same permissions as original
- Temporary files are created with secure permissions

### Input Validation
- All user inputs are validated against schema
- Special characters are sanitized in text inputs
- File paths are validated to prevent directory traversal

### Backup Security
- Backup files contain sensitive configuration data
- Backups should be included in .gitignore
- Consider encryption for highly sensitive configurations

### Audit Trail
- All configuration changes are logged with timestamps
- Backup files provide change history
- Operation logs include user context when available
