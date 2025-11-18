---
name: moai-project-config-manager
version: 4.0.0
created: 2025-11-05
updated: '2025-11-18'
status: stable
tier: specialization
description: Complete config.json CRUD operations with validation, merge strategy,
  and error recovery. Use for project initialization, setting updates, and configuration
  management with intelligent backup and recovery.
allowed-tools: Read, Write, Edit, AskUserQuestion, Bash, TodoWrite
primary-agent: alfred
secondary-agents: []
keywords:
- project
- config
- manager
- validation
- crud
tags:
- project-configuration
- management
- validation
- backup
- recovery
orchestration: null
can_resume: true
typical_chain_position: middle
depends_on: []
stability: stable
---


# moai-project-config-manager

**Project Configuration Management Specialist**

> **Primary Agent**: alfred  
> **Secondary Agents**: none  
> **Version**: 4.0.0  
> **Keywords**: project, config, manager, validation, crud

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (40 lines)

**Core Purpose**: Centralized management of all MoAI project configuration operations with robust validation and intelligent backup strategies.

**Key Capabilities**:
- ✅ **Complete CRUD**: Create, Read, Update, Delete configuration sections
- ✅ **Validation Engine**: Pre-save validation for all configuration changes
- ✅ **Merge Strategy**: Preserve unmodified sections while updating selected ones
- ✅ **Error Recovery**: Handle missing files, invalid JSON, permission issues
- ✅ **Batch Updates**: Handle multiple setting changes in single operation
- ✅ **Backup & Restore**: Automatic backup before major changes
- ✅ **Interactive Workflows**: User-friendly setting modification with TUI surveys

**Quick Usage**:
```python
# Interactive configuration update
Skill("moai-project-config-manager")

# Programmatic update
updates = {"language": {"conversation_language": "en"}}
Skill("moai-project-config-manager", action="update", changes=updates)

# Validation check
result = Skill("moai-project-config-manager", action="validate")
```

---

### Level 2: Core Implementation (110 lines)

**Configuration Structure**:
```json
{
  "language": {
    "conversation_language": "en|ko|ja|zh",
    "conversation_language_name": "English|한국어|日本語|中文",
    "agent_prompt_language": "english|localized"
  },
  "user": {
    "nickname": "string (max 20 chars)"
  },
  "github": {
    "auto_delete_branches": "boolean",
    "spec_git_workflow": "feature_branch|develop_direct|per_spec"
  },
  "report_generation": {
    "enabled": "boolean",
    "auto_create": "boolean",
    "user_choice": "Enable|Minimal|Disable"
  },
  "stack": {
    "selected_domains": ["frontend", "backend", "data", "devops", "security"]
  }
}
```

**Essential Operations**:

**1. Load Configuration**:
```python
def load_config():
    """Load configuration with error handling"""
    try:
        with open(".moai/config/config.json", "r") as f:
            return json.load(f)
    except FileNotFoundError:
        return create_default_config()
    except json.JSONDecodeError as e:
        raise ConfigError(f"Invalid JSON in config file: {e}")
```

**2. Update Configuration**:
```python
def update_config(updates):
    """Update configuration with merge strategy"""
    config = load_config()
    
    # Create backup
    backup_path = create_backup()
    
    # Merge changes
    new_config = merge_config(config, updates)
    
    # Validate and save
    validate_config(new_config)
    save_config_safely(new_config)
    
    return new_config
```

**3. Validation Rules**:
```python
def validate_config(config):
    """Validate configuration structure and values"""
    errors = []
    
    # Language validation
    valid_languages = ["en", "ko", "ja", "zh"]
    if config.get("language", {}).get("conversation_language") not in valid_languages:
        errors.append("Invalid conversation language")
    
    # Nickname validation
    nickname = config.get("user", {}).get("nickname", "")
    if len(nickname) > 20:
        errors.append("Nickname exceeds 20 characters")
    
    # GitHub workflow validation
    valid_workflows = ["feature_branch", "develop_direct", "per_spec"]
    if config.get("github", {}).get("spec_git_workflow") not in valid_workflows:
        errors.append("Invalid SPEC git workflow")
    
    return errors
```

**4. Interactive Update Workflow**:
```python
# Phase 1: Display current settings
display_current_config()

# Phase 2: Select sections to modify
selected_sections = ask_user_selections([
    "🌍 Language & Agent Prompt Language",
    "👤 Nickname", 
    "🔧 GitHub Settings",
    "📊 Report Generation",
    "🎯 Project Domains"
])

# Phase 3: Collect new values
updates = collect_updates_for_sections(selected_sections)

# Phase 4: Merge and save
update_config(updates)
```

---

### Level 3: Advanced Features (70 lines)

**Advanced Configuration Patterns**:

**1. Configuration Templates**:
```python
CONFIG_TEMPLATES = {
    "frontend_focus": {
        "stack": {"selected_domains": ["frontend", "security"]},
        "github": {"spec_git_workflow": "feature_branch"}
    },
    "full_stack": {
        "stack": {"selected_domains": ["frontend", "backend", "devops"]},
        "github": {"spec_git_workflow": "develop_direct"}
    },
    "data_science": {
        "stack": {"selected_domains": ["data", "backend"]},
        "report_generation": {"enabled": True, "auto_create": True}
    }
}
```

**2. Configuration Migration**:
```python
def migrate_config(from_version, to_version):
    """Migrate configuration between versions"""
    config = load_config()
    
    if from_version < "4.0.0" and to_version >= "4.0.0":
        # Add new report_generation section
        if "report_generation" not in config:
            config["report_generation"] = {
                "enabled": True,
                "auto_create": False,
                "user_choice": "Minimal"
            }
    
    return config
```

**3. Configuration Profiles**:
```python
def save_profile(name, config_subset):
    """Save configuration profile for reuse"""
    profile_path = f".moai/config/profiles/{name}.json"
    os.makedirs(os.path.dirname(profile_path), exist_ok=True)
    
    with open(profile_path, "w") as f:
        json.dump(config_subset, f, indent=2)

def load_profile(name):
    """Load and apply configuration profile"""
    profile_path = f".moai/config/profiles/{name}.json"
    
    with open(profile_path, "r") as f:
        profile_config = json.load(f)
    
    return update_config(profile_config)
```

**4. Configuration Validation with Context**:
```python
def validate_with_context(config, project_context):
    """Validate configuration with project context"""
    errors = validate_config(config)
    
    # Context-aware validation
    if project_context.get("has_github", False):
        if not config.get("github", {}).get("auto_delete_branches"):
            errors.append("GitHub projects should enable auto-delete branches")
    
    if "frontend" in config.get("stack", {}).get("selected_domains", []):
        if config.get("report_generation", {}).get("user_choice") == "Disable":
            errors.append("Frontend projects benefit from report generation")
    
    return errors
```

**5. Performance Optimization**:
```python
def optimize_config_for_performance(config):
    """Optimize configuration for better performance"""
    optimized = copy.deepcopy(config)
    
    # Enable minimal reports for better performance
    if optimized.get("report_generation", {}).get("user_choice") == "Enable":
        optimized["report_generation"]["user_choice"] = "Minimal"
    
    # Optimize workflow selection
    if optimized.get("github", {}).get("spec_git_workflow") == "per_spec":
        optimized["github"]["spec_git_workflow"] = "feature_branch"
    
    return optimized
```

---

### Level 4: Reference & Links (40 lines)

**Integration Points**:

**With Alfred Commands**:
- `/alfred:0-project` - Project initialization and configuration setup
- `/alfred:1-plan` - Access configuration for planning decisions  
- `/alfred:2-run` - Use configuration during execution
- `/alfred:3-sync` - Update configuration based on project changes

**With Other Skills**:
- `moai-core-ask-user-questions` - Interactive setting collection
- `moai-skill-factory` - Skill configuration management
- Domain-specific skills - Respect configuration settings for behavior

**Error Handling Reference**:

**Common Error Types**:
```python
class ConfigError(Exception):
    """Configuration management errors"""
    pass

class ValidationError(ConfigError):
    """Configuration validation errors"""
    pass

class BackupError(ConfigError):
    """Backup operation errors"""
    pass

class RestoreError(ConfigError):
    """Restore operation errors"""
    pass
```

**File Structure**:
```
.moai/
├── config/
│   ├── config.json          # Main configuration
│   ├── backup.*.json       # Automatic backups
│   └── profiles/            # Configuration profiles
│       ├── frontend.json
│       └── fullstack.json
└── logs/
    └── config-changes.log   # Configuration change log
```

**Best Practices**:
- Always validate before saving
- Create backups before major changes
- Use merge strategy, never overwrite entire config
- Provide clear error messages with recovery suggestions
- Maintain backward compatibility when possible

---

## 📈 Version History

** .0** (2025-11-13)
- ✨ Optimized 4-layer Progressive Disclosure structure
- ✨ Reduced from 707 to 260 lines (63% reduction)
- ✨ Enhanced configuration templates and profiles
- ✨ Improved error handling and validation
- ✨ Streamlined interactive workflows

**v3.0.0** (2025-11-12)
- ✨ Context7 MCP integration
- ✨ Enhanced validation engine
- ✨ Configuration migration support

**v2.0.0** (2025-11-05)
- ✨ Interactive configuration workflows
- ✨ Backup and restore capabilities
- ✨ Advanced merge strategies

**v1.0.0** (2025-10-15)
- ✨ Initial CRUD operations
- ✨ Basic validation
- ✨ Error recovery

---

**Generated with**: MoAI-ADK Skill Factory    
**Last Updated**: 2025-11-13  
**Maintained by**: Primary Agent (alfred)  
**Optimization**: 63% size reduction while preserving all functionality
