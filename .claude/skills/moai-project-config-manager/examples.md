# Project Configuration Manager - Usage Examples

## Quick Start Examples

### Basic Interactive Configuration Update

```python
# Start interactive configuration management
Skill("moai-project-config-manager")

# Expected workflow:
# 1. Displays current settings
# 2. User selects which settings to modify
# 3. Collects new values through TUI surveys
# 4. Validates and saves changes with backup
```

### Programmatic Language Change

```python
# Change conversation language to English
Skill(
    "moai-project-config-manager",
    action="update",
    changes={
        "language": {
            "conversation_language": "en",
            "conversation_language_name": "English",
            "agent_prompt_language": "english"
        }
    }
)

# Result: Configuration updated with automatic backup
```

### GitHub Settings Update

```python
# Enable auto-delete branches and set workflow
Skill(
    "moai-project-config-manager", 
    action="update",
    changes={
        "github": {
            "auto_delete_branches": True,
            "spec_git_workflow": "feature_branch"
        }
    }
)
```

---

## Real-World Scenarios

### Scenario 1: New Project Setup

**Context**: Setting up a new MoAI project for a Korean development team

```python
# Initialize configuration for Korean team
Skill(
    "moai-project-config-manager",
    action="update",
    changes={
        "language": {
            "conversation_language": "ko",
            "conversation_language_name": "한국어",
            "agent_prompt_language": "localized"
        },
        "user": {
            "nickname": "개발팀"
        },
        "github": {
            "auto_delete_branches": True,
            "spec_git_workflow": "feature_branch"
        },
        "report_generation": {
            "enabled": True,
            "auto_create": False,
            "user_choice": "Minimal"
        },
        "stack": {
            "selected_domains": ["frontend", "backend", "devops"]
        }
    }
)

# Validation to ensure setup is correct
validation = Skill("moai-project-config-manager", action="validate")
if validation["valid"]:
    print("✅ Project configuration ready for Korean team")
else:
    print(f"❌ Configuration errors: {validation['errors']}")
```

### Scenario 2: Team Workflow Migration

**Context**: Migrating from direct commits to PR-based workflow

```python
# First, backup current configuration
backup_result = Skill("moai-project-config-manager", action="backup")
print(f"📁 Configuration backed up to: {backup_result['backup_path']}")

# Update GitHub workflow for team collaboration
Skill(
    "moai-project-config-manager",
    action="update", 
    changes={
        "github": {
            "auto_delete_branches": True,
            "spec_git_workflow": "feature_branch"
        }
    }
)

# Verify the change
current_config = Skill("moai-project-config-manager", action="get_current")
workflow = current_config["github"]["spec_git_workflow"]
print(f"🔄 Git workflow updated to: {workflow}")
```

### Scenario 3: Report Generation Optimization

**Context**: Reducing token usage by switching to minimal reports

```python
# Check current report settings
config = Skill("moai-project-config-manager", action="get_current")
current_reports = config["report_generation"]
print(f"Current report setting: {current_reports['user_choice']}")

# Update to minimal reports for cost savings
Skill(
    "moai-project-config-manager",
    action="update",
    changes={
        "report_generation": {
            "enabled": True,
            "auto_create": False,
            "user_choice": "Minimal"
        }
    }
)

print("✅ Updated to minimal reports (saves ~60% tokens)")
```

### Scenario 4: Multi-Language Project Support

**Context**: Adding Japanese language support to existing project

```python
# Show what will change before applying
diff_result = Skill(
    "moai-project-config-manager",
    action="diff",
    changes={
        "language": {
            "conversation_language": "ja",
            "conversation_language_name": "日本語"
        }
    }
)

print("Planned changes:")
for change in diff_result["changes"]:
    print(f"  {change['path']}: {change['old']} → {change['new']}")

# Apply the change
Skill(
    "moai-project-config-manager",
    action="update",
    changes={
        "language": {
            "conversation_language": "ja", 
            "conversation_language_name": "日本語"
        }
    }
)
```

---

## Integration Examples

### Integration with Alfred Commands

#### In /alfred:0-project Command

```python
def initialize_or_update_project():
    """Handle project initialization or setting updates"""
    
    # Check if configuration exists
    if not Path(".moai/config/config.json").exists():
        print("🆕 Creating new project configuration...")
        # Use skill to create default configuration
        Skill("moai-project-config-manager", action="create_default")
    else:
        print("🔧 Existing project found. Select operation:")
        # Interactive choice for project setup vs setting modification
        response = AskUserQuestion(
            question="What would you like to do?",
            options=[
                "🆕 Initialize project documentation",
                "🔧 Modify project settings", 
                "📊 View current configuration"
            ]
        )
        
        if response == "🔧 Modify project settings":
            # Delegate to configuration manager
            Skill("moai-project-config-manager")
        elif response == "📊 View current configuration":
            show_current_configuration()
```

#### In /alfred:1-plan Command

```python
def plan_with_configuration():
    """Use configuration settings in planning phase"""
    
    # Load and validate configuration
    validation = Skill("moai-project-config-manager", action="validate")
    if not validation["valid"]:
        print(f"❌ Configuration errors: {validation['errors']}")
        return
    
    # Get configuration for planning
    config = Skill("moai-project-config-manager", action="get_current")
    
    # Use language setting for plan output
    language = config["language"]["conversation_language"]
    if language == "ko":
        plan_prompt = "한국어로 개발 계획을 작성해주세요"
    else:
        plan_prompt = "Please write the development plan in English"
    
    # Use domain settings for specialized planning
    domains = config.get("stack", {}).get("selected_domains", [])
    if "frontend" in domains:
        plan_prompt += "\nInclude frontend architecture considerations"
    if "security" in domains:
        plan_prompt += "\nInclude security planning and best practices"
    
    # Generate plan with configuration context
    generate_development_plan(plan_prompt)
```

#### In /alfred:2-run Command

```python
def execute_with_configuration():
    """Execute tasks using configuration settings"""
    
    # Load configuration
    config = Skill("moai-project-config-manager", action="get_current")
    
    # Configure execution based on settings
    execution_config = {
        "language": config["language"]["conversation_language"],
        "workflow": config["github"]["spec_git_workflow"],
        "reports": config["report_generation"]["user_choice"],
        "domains": config.get("stack", {}).get("selected_domains", [])
    }
    
    # Adjust execution based on configuration
    if execution_config["workflow"] == "feature_branch":
        create_feature_branch()
        execute_in_feature_branch()
    else:
        execute_directly()
    
    # Generate reports based on setting
    if execution_config["reports"] == "Enable":
        generate_full_report()
    elif execution_config["reports"] == "Minimal":
        generate_minimal_report()
```

### Integration with Custom Scripts

#### Configuration Migration Script

```python
#!/usr/bin/env python3
"""
Migrate configuration from old format to new format
"""

def migrate_configuration():
    """Migrate old configuration to new schema"""
    
    print("🔄 Starting configuration migration...")
    
    # Backup current configuration
    backup_result = Skill("moai-project-config-manager", action="backup")
    print(f"📁 Backup created: {backup_result['backup_path']}")
    
    # Load current configuration
    config = Skill("moai-project-config-manager", action="get_current")
    
    # Perform migration transformations
    updates = {}
    
    # Example: Add new report_generation settings if missing
    if "report_generation" not in config:
        updates["report_generation"] = {
            "enabled": True,
            "auto_create": False,
            "user_choice": "Minimal",
            "updated_at": datetime.now().isoformat()
        }
    
    # Example: Migrate old domain format
    if "project_domains" in config and "stack" not in config:
        updates["stack"] = {
            "selected_domains": config["project_domains"],
            "domain_selection_date": datetime.now().isoformat()
        }
    
    # Apply updates if any
    if updates:
        Skill("moai-project-config-manager", action="update", changes=updates)
        print(f"✅ Applied {len(updates)} migration updates")
    else:
        print("✅ Configuration already up to date")
    
    # Validate migrated configuration
    validation = Skill("moai-project-config-manager", action="validate")
    if validation["valid"]:
        print("✅ Migration completed successfully")
    else:
        print(f"❌ Migration validation failed: {validation['errors']}")

if __name__ == "__main__":
    migrate_configuration()
```

#### Configuration Backup Manager

```python
#!/usr/bin/env python3
"""
Manage configuration backups with cleanup and restore capabilities
"""

def backup_manager():
    """Manage configuration backups"""
    
    while True:
        print("\n🗂️ Configuration Backup Manager")
        print("1. Create backup")
        print("2. List backups") 
        print("3. Restore backup")
        print("4. Clean old backups")
        print("5. Exit")
        
        choice = input("Select option (1-5): ")
        
        if choice == "1":
            # Create backup
            result = Skill("moai-project-config-manager", action="backup")
            print(f"✅ Backup created: {result['backup_path']}")
        
        elif choice == "2":
            # List backups
            backups = Skill("moai-project-config-manager", action="list_backups")
            if backups:
                print("\n📁 Available backups:")
                for path, timestamp in backups.items():
                    print(f"  {path} - {timestamp}")
            else:
                print("📭 No backups found")
        
        elif choice == "3":
            # Restore backup
            backups = Skill("moai-project-config-manager", action="list_backups")
            if not backups:
                print("📭 No backups available to restore")
                continue
            
            print("\nSelect backup to restore:")
            for i, (path, timestamp) in enumerate(backups.items(), 1):
                print(f"{i}. {path} - {timestamp}")
            
            try:
                selection = int(input("Enter backup number: ")) - 1
                backup_path = list(backups.keys())[selection]
                
                result = Skill(
                    "moai-project-config-manager",
                    action="restore",
                    backup_path=backup_path
                )
                
                if result["success"]:
                    print(f"✅ Restored from: {backup_path}")
                else:
                    print(f"❌ Restore failed: {result.get('error', 'Unknown error')}")
            
            except (ValueError, IndexError):
                print("❌ Invalid selection")
        
        elif choice == "4":
            # Clean old backups
            days = int(input("Delete backups older than how many days? "))
            result = Skill(
                "moai-project-config-manager", 
                action="cleanup_backups",
                older_than_days=days
            )
            print(f"🗑️ Deleted {result['deleted_count']} old backups")
        
        elif choice == "5":
            break
        
        else:
            print("❌ Invalid option")

if __name__ == "__main__":
    backup_manager()
```

---

## Error Handling Examples

### Handling Configuration Errors

```python
def safe_configuration_update(changes):
    """Safely update configuration with error handling"""
    
    try:
        # Validate current configuration first
        validation = Skill("moai-project-config-manager", action="validate")
        if not validation["valid"]:
            print(f"⚠️ Current configuration has errors: {validation['errors']}")
        
        # Show what will change
        diff_result = Skill(
            "moai-project-config-manager",
            action="diff",
            changes=changes
        )
        
        print("Planned changes:")
        for change in diff_result["changes"]:
            print(f"  {change['path']}: {change['old']} → {change['new']}")
        
        # Ask for confirmation
        confirm = AskUserQuestion(
            question="Apply these changes?",
            options=["Yes", "No"]
        )
        
        if confirm == "Yes":
            # Apply changes
            result = Skill(
                "moai-project-config-manager",
                action="update",
                changes=changes
            )
            
            if result.get("success", True):
                print("✅ Configuration updated successfully")
                if "backup_path" in result:
                    print(f"📁 Backup created: {result['backup_path']}")
            else:
                print(f"❌ Update failed: {result.get('error', 'Unknown error')}")
        
        else:
            print("❌ Changes cancelled")
    
    except Exception as e:
        print(f"❌ Unexpected error: {e}")
        
        # Try to restore from backup if available
        if "backup_path" in locals():
            restore_result = Skill(
                "moai-project-config-manager",
                action="restore",
                backup_path=backup_path
            )
            if restore_result["success"]:
                print("✅ Restored configuration from backup")
```

### Configuration Recovery

```python
def recover_configuration():
    """Recover from configuration corruption"""
    
    print("🔧 Configuration Recovery Mode")
    
    # Check if configuration file exists
    if not Path(".moai/config/config.json").exists():
        print("📭 Configuration file not found")
        create_new = AskUserQuestion(
            question="Create new default configuration?",
            options=["Yes", "No"]
        )
        
        if create_new == "Yes":
            Skill("moai-project-config-manager", action="create_default")
            print("✅ New configuration created")
        return
    
    # Try to validate existing configuration
    validation = Skill("moai-project-config-manager", action="validate")
    
    if validation["valid"]:
        print("✅ Configuration is valid")
        return
    
    print(f"❌ Configuration errors found: {validation['errors']}")
    
    # Show recovery options
    recovery_option = AskUserQuestion(
        question="How would you like to recover?",
        options=[
            "🔧 Attempt automatic repair",
            "📁 Restore from backup",
            "🆕 Create new configuration",
            "❌ Exit"
        ]
    )
    
    if recovery_option == "🔧 Attempt automatic repair":
        # Try to repair common issues
        print("🔧 Attempting automatic repair...")
        result = Skill("moai-project-config-manager", action="auto_repair")
        
        if result["success"]:
            print("✅ Configuration repaired successfully")
        else:
            print(f"❌ Automatic repair failed: {result['error']}")
    
    elif recovery_option == "📁 Restore from backup":
        # List available backups
        backups = Skill("moai-project-config-manager", action="list_backups")
        
        if backups:
            print("📁 Available backups:")
            for i, (path, timestamp) in enumerate(backups.items(), 1):
                print(f"{i}. {path} - {timestamp}")
            
            try:
                selection = int(input("Select backup to restore: ")) - 1
                backup_path = list(backups.keys())[selection]
                
                result = Skill(
                    "moai-project-config-manager",
                    action="restore",
                    backup_path=backup_path
                )
                
                if result["success"]:
                    print("✅ Configuration restored from backup")
                else:
                    print(f"❌ Restore failed: {result['error']}")
            
            except (ValueError, IndexError):
                print("❌ Invalid backup selection")
        else:
            print("📭 No backups available")
    
    elif recovery_option == "🆕 Create new configuration":
        # Create fresh configuration
        Skill("moai-project-config-manager", action="create_default")
        print("✅ New default configuration created")
```

---

## Testing and Validation Examples

### Unit Test Examples

```python
import pytest
from unittest.mock import patch, MagicMock

def test_language_validation():
    """Test language configuration validation"""
    
    # Valid language configuration
    valid_config = {
        "language": {
            "conversation_language": "ko",
            "conversation_language_name": "한국어",
            "agent_prompt_language": "localized"
        }
    }
    
    result = Skill("moai-project-config-manager", action="validate", config=valid_config)
    assert result["valid"] == True
    
    # Invalid language configuration
    invalid_config = {
        "language": {
            "conversation_language": "fr",  # Invalid language code
            "conversation_language_name": "Français",
            "agent_prompt_language": "localized"
        }
    }
    
    result = Skill("moai-project-config-manager", action="validate", config=invalid_config)
    assert result["valid"] == False
    assert any("Invalid language code" in error["message"] for error in result["errors"])

def test_configuration_merge():
    """Test configuration merging logic"""
    
    original_config = {
        "language": {
            "conversation_language": "ko",
            "conversation_language_name": "한국어",
            "agent_prompt_language": "localized"
        },
        "github": {
            "auto_delete_branches": False,
            "spec_git_workflow": "develop_direct"
        }
    }
    
    # Update only language settings
    changes = {
        "language": {
            "conversation_language": "en",
            "conversation_language_name": "English"
        }
    }
    
    result = Skill(
        "moai-project-config-manager",
        action="merge_test",
        original=original_config,
        changes=changes
    )
    
    # Language should be updated
    assert result["merged"]["language"]["conversation_language"] == "en"
    # GitHub settings should be preserved
    assert result["merged"]["github"]["auto_delete_branches"] == False

@patch('pathlib.Path.exists')
@patch('pathlib.Path.read_text')
def test_invalid_json_handling(mock_read, mock_exists):
    """Test handling of invalid JSON configuration"""
    
    mock_exists.return_value = True
    mock_read.return_value = '{"invalid": json content}'
    
    # Should attempt repair or raise appropriate error
    with pytest.raises(ConfigError) as exc_info:
        Skill("moai-project-config-manager", action="load")
    
    assert "Invalid JSON" in str(exc_info.value)
```

### Integration Test Examples

```python
def test_full_configuration_workflow():
    """Test complete configuration management workflow"""
    
    # Start with clean state
    if Path(".moai/config/config.json").exists():
        Path(".moai/config/config.json").unlink()
    
    # 1. Create initial configuration
    Skill("moai-project-config-manager", action="create_default")
    
    # Verify creation
    assert Path(".moai/config/config.json").exists()
    
    validation = Skill("moai-project-config-manager", action="validate")
    assert validation["valid"] == True
    
    # 2. Update configuration
    changes = {
        "language": {
            "conversation_language": "en",
            "conversation_language_name": "English"
        },
        "github": {
            "auto_delete_branches": True
        }
    }
    
    Skill("moai-project-config-manager", action="update", changes=changes)
    
    # Verify updates
    config = Skill("moai-project-config-manager", action="get_current")
    assert config["language"]["conversation_language"] == "en"
    assert config["github"]["auto_delete_branches"] == True
    
    # 3. Create backup and test restore
    backup_result = Skill("moai-project-config-manager", action="backup")
    
    # Make additional changes
    Skill("moai-project-config-manager", action="update", changes={
        "user": {"nickname": "TestUser"}
    })
    
    # Restore from backup
    Skill("moai-project-config-manager", action="restore", 
          backup_path=backup_result["backup_path"])
    
    # Verify restore
    config = Skill("moai-project-config-manager", action="get_current")
    assert "user" not in config  # Should be gone after restore
    assert config["language"]["conversation_language"] == "en"  # Should remain

def test_error_recovery():
    """Test error recovery mechanisms"""
    
    # Create valid configuration
    Skill("moai-project-config-manager", action="create_default")
    
    # Simulate corruption by writing invalid JSON
    with open(".moai/config/config.json", "w") as f:
        f.write('{"invalid": json}')
    
    # Should handle gracefully
    try:
        validation = Skill("moai-project-config-manager", action="validate")
        assert validation["valid"] == False
        assert len(validation["errors"]) > 0
    except Exception as e:
        # Should raise appropriate error, not crash
        assert "Invalid JSON" in str(e)
    
    # Recovery should work
    Skill("moai-project-config-manager", action="auto_repair")
    
    # Should be valid after repair
    validation = Skill("moai-project-config-manager", action="validate")
    assert validation["valid"] == True
```

---

## Performance Monitoring Examples

### Configuration Performance Metrics

```python
def monitor_configuration_performance():
    """Monitor configuration operation performance"""
    
    import time
    
    operations = []
    
    # Test configuration loading
    start = time.time()
    config = Skill("moai-project-config-manager", action="get_current")
    load_time = time.time() - start
    operations.append(("load", load_time))
    
    # Test validation
    start = time.time()
    validation = Skill("moai-project-config-manager", action="validate")
    validation_time = time.time() - start
    operations.append(("validate", validation_time))
    
    # Test backup creation
    start = time.time()
    backup = Skill("moai-project-config-manager", action="backup")
    backup_time = time.time() - start
    operations.append(("backup", backup_time))
    
    # Test update operation
    start = time.time()
    Skill("moai-project-config-manager", action="update", changes={
        "report_generation": {"updated_at": datetime.now().isoformat()}
    })
    update_time = time.time() - start
    operations.append(("update", update_time))
    
    # Report performance
    print("📊 Configuration Performance Metrics:")
    for operation, duration in operations:
        status = "✅" if duration < 0.1 else "⚠️" if duration < 0.5 else "❌"
        print(f"  {status} {operation}: {duration:.3f}s")
    
    # Performance recommendations
    avg_time = sum(duration for _, duration in operations) / len(operations)
    if avg_time > 0.2:
        print("\n⚠️ Performance Recommendations:")
        print("  - Consider optimizing JSON parsing")
        print("  - Check disk I/O performance")
        print("  - Review validation rules complexity")
    else:
        print("\n✅ Performance is within acceptable ranges")

def analyze_configuration_usage():
    """Analyze configuration usage patterns"""
    
    # Load configuration
    config = Skill("moai-project-config-manager", action="get_current")
    
    # Analyze settings usage
    analysis = {
        "language_efficiency": analyze_language_settings(config),
        "github_workflow": analyze_github_settings(config),
        "report_optimization": analyze_report_settings(config),
        "domain_coverage": analyze_domain_settings(config)
    }
    
    print("📈 Configuration Usage Analysis:")
    
    # Language analysis
    lang_analysis = analysis["language_efficiency"]
    print(f"\n🌍 Language Settings:")
    print(f"  Conversation Language: {lang_analysis['conversation_language']}")
    print(f"  Agent Language: {lang_analysis['agent_prompt_language']}")
    print(f"  Localization Impact: {lang_analysis['localization_impact']}")
    
    # GitHub analysis
    github_analysis = analysis["github_workflow"]
    print(f"\n🔧 GitHub Workflow:")
    print(f"  Auto-delete Branches: {github_analysis['auto_delete']}")
    print(f"  Workflow Type: {github_analysis['workflow_type']}")
    print(f"  Team Collaboration: {github_analysis['collaboration_score']}/10")
    
    # Report analysis
    report_analysis = analysis["report_optimization"]
    print(f"\n📊 Report Generation:")
    print(f"  Setting: {report_analysis['current_setting']}")
    print(f"  Token Efficiency: {report_analysis['token_efficiency']}")
    print(f"  Cost Optimization: {report_analysis['cost_impact']}")
    
    # Domain analysis
    domain_analysis = analysis["domain_coverage"]
    print(f"\n🎯 Domain Coverage:")
    print(f"  Selected Domains: {', '.join(domain_analysis['domains'])}")
    print(f"  Coverage Score: {domain_analysis['coverage_score']}/5")
    print(f"  Recommendations: {domain_analysis['recommendations']}")

def analyze_language_settings(config):
    """Analyze language configuration efficiency"""
    
    lang = config.get("language", {})
    
    conversation_lang = lang.get("conversation_language", "en")
    agent_lang = lang.get("agent_prompt_language", "english")
    
    # Calculate localization impact
    localization_impact = "low" if agent_lang == "english" else "high"
    
    return {
        "conversation_language": conversation_lang,
        "agent_prompt_language": agent_lang,
        "localization_impact": localization_impact
    }

def analyze_github_settings(config):
    """Analyze GitHub workflow settings"""
    
    github = config.get("github", {})
    
    auto_delete = github.get("auto_delete_branches", False)
    workflow = github.get("spec_git_workflow", "develop_direct")
    
    # Calculate collaboration score
    collaboration_scores = {
        "feature_branch": 9,
        "develop_direct": 5,
        "per_spec": 7
    }
    
    return {
        "auto_delete": auto_delete,
        "workflow_type": workflow,
        "collaboration_score": collaboration_scores.get(workflow, 5)
    }

def analyze_report_settings(config):
    """Analyze report generation optimization"""
    
    reports = config.get("report_generation", {})
    
    setting = reports.get("user_choice", "Minimal")
    enabled = reports.get("enabled", True)
    
    # Calculate efficiency metrics
    efficiency_map = {
        "Enable": {"tokens": "50-60", "cost": "high"},
        "Minimal": {"tokens": "20-30", "cost": "medium"},
        "Disable": {"tokens": "0", "cost": "low"}
    }
    
    efficiency = efficiency_map.get(setting, efficiency_map["Minimal"])
    
    return {
        "current_setting": setting,
        "token_efficiency": f"{efficiency['tokens']} tokens/report",
        "cost_impact": efficiency["cost"]
    }

def analyze_domain_settings(config):
    """Analyze domain selection coverage"""
    
    stack = config.get("stack", {})
    domains = stack.get("selected_domains", [])
    
    # Calculate coverage score
    available_domains = ["frontend", "backend", "data", "devops", "security"]
    coverage_score = len(domains) / len(available_domains) * 5
    
    # Generate recommendations
    recommendations = []
    if "frontend" not in domains:
        recommendations.append("Consider frontend domain for UI projects")
    if "security" not in domains:
        recommendations.append("Add security domain for production projects")
    if len(domains) < 2:
        recommendations.append("Select more domains for comprehensive coverage")
    
    return {
        "domains": domains,
        "coverage_score": round(coverage_score, 1),
        "recommendations": recommendations
    }

if __name__ == "__main__":
    monitor_configuration_performance()
    analyze_configuration_usage()
```

---

## Advanced Usage Patterns

### Configuration Templates

```python
def apply_configuration_template(template_name):
    """Apply predefined configuration templates"""
    
    templates = {
        "startup": {
            "description": "Fast-moving startup with direct workflow",
            "settings": {
                "language": {
                    "conversation_language": "en",
                    "conversation_language_name": "English",
                    "agent_prompt_language": "english"
                },
                "github": {
                    "auto_delete_branches": True,
                    "spec_git_workflow": "develop_direct"
                },
                "report_generation": {
                    "enabled": True,
                    "auto_create": False,
                    "user_choice": "Minimal"
                },
                "stack": {
                    "selected_domains": ["frontend", "backend"]
                }
            }
        },
        "enterprise": {
            "description": "Enterprise team with PR workflow",
            "settings": {
                "language": {
                    "conversation_language": "en",
                    "conversation_language_name": "English", 
                    "agent_prompt_language": "english"
                },
                "github": {
                    "auto_delete_branches": True,
                    "spec_git_workflow": "feature_branch"
                },
                "report_generation": {
                    "enabled": True,
                    "auto_create": True,
                    "user_choice": "Enable"
                },
                "stack": {
                    "selected_domains": ["frontend", "backend", "devops", "security"]
                }
            }
        },
        "research": {
            "description": "Research project with minimal workflow",
            "settings": {
                "language": {
                    "conversation_language": "en",
                    "conversation_language_name": "English",
                    "agent_prompt_language": "english"
                },
                "github": {
                    "auto_delete_branches": False,
                    "spec_git_workflow": "per_spec"
                },
                "report_generation": {
                    "enabled": True,
                    "auto_create": True,
                    "user_choice": "Enable"
                },
                "stack": {
                    "selected_domains": ["data", "frontend"]
                }
            }
        }
    }
    
    if template_name not in templates:
        print(f"❌ Unknown template: {template_name}")
        print(f"Available templates: {', '.join(templates.keys())}")
        return
    
    template = templates[template_name]
    print(f"📋 Applying template: {template['description']}")
    
    # Apply template settings
    Skill("moai-project-config-manager", action="update", changes=template["settings"])
    
    print(f"✅ Template '{template_name}' applied successfully")

# Usage examples:
# apply_configuration_template("startup")
# apply_configuration_template("enterprise") 
# apply_configuration_template("research")
```

### Configuration Synchronization

```python
def synchronize_configurations(source_project, target_project):
    """Synchronize configuration between projects"""
    
    print(f"🔄 Synchronizing configuration from {source_project} to {target_project}")
    
    # Load source configuration
    source_config = Skill(
        "moai-project-config-manager",
        action="load_from_path",
        project_path=source_project
    )
    
    # Backup target configuration
    target_backup = Skill(
        "moai-project-config-manager",
        action="backup",
        project_path=target_project
    )
    
    print(f"📁 Target configuration backed up: {target_backup['backup_path']}")
    
    # Determine which sections to synchronize
    sync_options = [
        "🌍 Language settings",
        "🔧 GitHub settings", 
        "📊 Report settings",
        "🎯 Domain settings",
        "👤 User settings",
        "📋 All settings"
    ]
    
    selection = AskUserQuestion(
        question="Which settings to synchronize?",
        options=sync_options
    )
    
    # Prepare changes based on selection
    changes = {}
    
    if selection in ["🌍 Language settings", "📋 All settings"]:
        changes["language"] = source_config.get("language", {})
    
    if selection in ["🔧 GitHub settings", "📋 All settings"]:
        changes["github"] = source_config.get("github", {})
    
    if selection in ["📊 Report settings", "📋 All settings"]:
        changes["report_generation"] = source_config.get("report_generation", {})
    
    if selection in ["🎯 Domain settings", "📋 All settings"]:
        changes["stack"] = source_config.get("stack", {})
    
    if selection in ["👤 User settings", "📋 All settings"]:
        changes["user"] = source_config.get("user", {})
    
    # Apply synchronization
    if changes:
        Skill(
            "moai-project-config-manager",
            action="update",
            changes=changes,
            project_path=target_project
        )
        
        print(f"✅ Synchronized {len(changes)} configuration sections to {target_project}")
    else:
        print("❌ No settings selected for synchronization")
```

These examples demonstrate comprehensive usage patterns for the moai-project-config-manager skill, from basic operations to advanced integration scenarios, error handling, testing, and performance monitoring.
