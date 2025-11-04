# Examples

## Template Optimization Workflow

### Basic Template Comparison
```bash
# Detect available backups
ls -la .moai-backups/
# backup-20251105-143022.tar.gz

# Start optimization workflow
Skill("moai-project-template-optimizer")

# Alfred will:
# 1. Detect backup files
# 2. Compare with current templates
# 3. Identify differences
# 4. Propose smart merge strategy
```

### Smart Merge Example
```python
# Template optimization decision process
def optimize_templates():
    backup = find_latest_backup()
    current = read_current_templates()
    package = read_package_templates()
    
    # Analyze differences
    differences = compare_all(backup, current, package)
    
    # Smart merge strategy
    for diff in differences:
        if diff.type == "user_customization":
            # Preserve user changes
            preserve_user_content(diff)
        elif diff.type == "template_update":
            # Apply package template updates
            apply_template_update(diff)
        elif diff.type == "conflict":
            # Ask user for resolution
            resolve_with_user_input(diff)
```

### Backup Analysis Example
```
Detected backup: backup-20251105-143022.tar.gz
Created: 2025-11-05 14:30:22
Size: 2.4MB
Contents: 15 files, 3 directories

Analysis:
- 5 files have user customizations
- 3 files need template updates
- 2 files have conflicts requiring resolution

Recommended action: Smart merge with user confirmation
```
