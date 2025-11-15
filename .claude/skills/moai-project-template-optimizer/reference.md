# Reference

## Template Optimization Process

### Phase 1: Backup Discovery
```bash
# Find backup files
find .moai-backups/ -name "backup-*.tar.gz" -type f

# Analyze backup contents
tar -tzf backup-20251105-143022.tar.gz | head -20
```

### Phase 2: Comparison Analysis
```python
def compare_templates(backup_path, current_path, package_path):
    """Compare template versions"""
    
    comparison = {
        "backup_only": [],      # Files only in backup
        "current_only": [],     # Files only in current  
        "package_only": [],     # Files only in package
        "all_three": [],        # Files in all versions
        "differences": []       # Files with content differences
    }
    
    # Perform detailed comparison
    # ... comparison logic
    
    return comparison
```

### Phase 3: Smart Merge Strategy
```python
def smart_merge_strategy(comparison_result):
    """Determine optimal merge approach"""
    
    strategy = {
        "preserve_user": [],    # Keep user customizations
        "update_template": [],  # Apply template updates
        "resolve_conflict": []  # Need user input
    }
    
    for file_info in comparison_result.differences:
        if file_info.has_user_customizations:
            if file_info.template_has_updates:
                strategy["resolve_conflict"].append(file_info)
            else:
                strategy["preserve_user"].append(file_info)
        else:
            strategy["update_template"].append(file_info)
    
    return strategy
```

## Integration Points

### With Alfred Commands
- `/alfred:0-project` - Initial project setup optimization
- Template updates - Automatic detection and optimization
- Version upgrades - Template migration workflows

### With Template System
- Package templates - Source of truth for latest versions
- Local templates - User customizations and modifications
- Backup system - Rollback and recovery capabilities

## File Locations

### Template Paths
- **Package Templates**: `src/moai_adk/templates/.claude/`
- **Local Templates**: `.claude/`
- **Backup Storage**: `.moai-backups/`

### Critical Files to Monitor
- `.claude/settings.json` - Claude Code configuration
- `.claude/hooks/` - Hook scripts
- `.moai/config/config.json` - MoAI project configuration
- Template directories and files

## Conflict Resolution Patterns

### User Customizations vs Template Updates
```python
def resolve_template_conflict(file_path, backup_content, current_content, template_content):
    """Resolve conflicts between user customizations and template updates"""
    
    # Identify user customizations
    user_changes = diff_files(backup_content, current_content)
    
    # Identify template updates  
    template_changes = diff_files(backup_content, template_content)
    
    # Check for conflicts
    conflicts = find_conflicts(user_changes, template_changes)
    
    if not conflicts:
        # No conflicts - merge both
        return merge_changes(current_content, template_changes)
    else:
        # Conflicts exist - need user input
        return request_user_resolution(file_path, conflicts)
```

### Configuration File Merging
```python
def merge_config_files(backup_config, current_config, template_config):
    """Smart merge for configuration files"""
    
    merged = {}
    
    # Preserve user settings
    for key, value in current_config.items():
        if key not in template_config or current_config[key] != template_config[key]:
            merged[key] = value
    
    # Add new template settings
    for key, value in template_config.items():
        if key not in merged:
            merged[key] = value
    
    return merged
```
