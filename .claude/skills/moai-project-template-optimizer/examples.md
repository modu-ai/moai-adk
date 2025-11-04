# Template Optimizer Examples

## Basic Usage Examples

### Standard Template Optimization
```python
# Complete optimization workflow
Skill("moai-project-template-optimizer")

# This will:
# 1. Scan for .moai-backups/ directory
# 2. Analyze available backups
# 3. Compare current templates with backup
# 4. Perform smart merge preserving user customizations
# 5. Update configuration with optimization flags
```

### Backup Analysis Only
```python
# Analyze existing backups without making changes
Skill("moai-project-template-optimizer", mode="analyze_only")

# Output example:
# Backup Analysis Report:
# - Backups found: 3
# - Most recent: backup-2025-10-15-v0.16.0
# - Integrity: 100%
# - Recommended action: Full optimization
```

### Specific Backup Restoration
```python
# Restore from specific backup
Skill("moai-project-template-optimizer", 
       mode="restore", 
       backup="backup-2025-10-15-v0.16.0")

# This will:
# 1. Validate specified backup exists
# 2. Extract user customizations from backup
# 3. Merge with current template structure
# 4. Preserve backup settings while updating structure
```

### Rollback Operation
```python
# Rollback to previous state
Skill("moai-project-template-optimizer", mode="rollback")

# This will:
# 1. Check for rollback points
# 2. Restore previous configuration
# 3. Verify state consistency
# 4. Report rollback success/failure
```

## Advanced Usage Examples

### Template Comparison Analysis
```python
# Get detailed comparison report
Skill("moai-project-template-optimizer", 
       mode="compare",
       detailed_report=True)

# Sample output:
# Template Comparison Report:
# Files compared: 47
# Differences found: 12
# Conflicts: 0
# 
# File-by-file analysis:
# - .claude/settings.json: 3 structural changes
# - CLAUDE.md: 2 section updates
# - hooks/: 7 function updates
```

### Custom Merge Strategy
```python
# Use custom merge preferences
Skill("moai-project-template-optimizer",
       mode="optimize",
       merge_strategy="preserve_user",
       conflict_resolution="template_priority")

# Merge strategies:
# - preserve_user: Always favor user customizations
# - template_priority: Always favor latest template
# - smart_merge: Intelligent three-way merge (default)
```

### Version-Specific Optimization
```python
# Optimize for specific template version
Skill("moai-project-template-optimizer",
       target_version="0.17.0",
       compatibility_check=True)

# This will:
# 1. Check compatibility with current version
# 2. Apply version-specific optimizations
# 3. Handle deprecated features
# 4. Update version tracking
```

## Integration Examples

### With moai-adk Update Workflow
```python
# After moai-adk package update
# Automatically triggered by update detection
Skill("moai-project-template-optimizer")

# Typical update scenario:
# 1. User runs: uv update moai-adk
# 2. Update detects template changes
# 3. Auto-triggers template optimizer
# 4. Preserves user settings during update
```

### With Project Initialization
```python
# As part of /alfred:0-project command
Skill("moai-project-language-initializer")  # First
Skill("moai-project-template-optimizer")   # Then

# Sequence ensures:
# 1. User preferences are set first
# 2. Templates are optimized with user context
# 3. All configurations are consistent
```

### With Sync Operations
```python
# During /alfred:3-sync command
if template_changes_detected:
    Skill("moai-project-template-optimizer", mode="sync_update")

# This ensures:
# 1. Template sync includes user customizations
# 2. History sections are updated
# 3. Version tracking is maintained
```

## Error Handling Examples

### Backup Missing Error
```python
# When no backups are found
Skill("moai-project-template-optimizer")

# Error response:
# "No backups found in .moai-backups/ directory."
# "Options:"
# "1. Create backup before optimization"
# "2. Proceed with template-only optimization"
# "3. Cancel operation"
```

### Conflict Resolution
```python
# When conflicts are detected
Skill("moai-project-template-optimizer", 
       mode="resolve_conflicts")

# Conflict resolution options:
# 1. Use template version (recommended)
# 2. Preserve user customization
# 3. Manual merge required
# 4. Skip conflicting files
```

### Version Incompatibility
```python
# When versions are incompatible
Skill("moai-project-template-optimizer",
       compatibility_check=True)

# Response:
# "Version compatibility warning detected"
# "Current: 0.15.0, Target: 0.17.0"
# "Migration path: v0.15.0 → v0.16.0 → v0.17.0"
# "Recommendation: Stepwise migration"
```

## Performance Examples

### Large Project Optimization
```python
# For projects with many files
Skill("moai-project-template-optimizer",
       mode="optimize",
       performance_mode="fast")

# Optimizations applied:
# - Parallel file processing
# - Hash-based comparison
# - Cached template patterns
# - Batch merge operations
```

### Incremental Updates
```python
# For small template changes
Skill("moai-project-template-optimizer",
       mode="incremental",
       changed_files_only=True)

# Benefits:
# - Faster processing
# - Lower resource usage
# - Reduced backup size
# - Quick rollback capability
```

## Reporting Examples

### Detailed Optimization Report
```python
Skill("moai-project-template-optimizer",
       mode="analyze",
       report_format="detailed")

# Sample report structure:
# {
#   "backup_analysis": {...},
#   "template_comparison": {...},
#   "merge_results": {...},
#   "configuration_updates": {...},
#   "performance_metrics": {...},
#   "recommendations": [...]
# }
```

### Summary Report
```python
Skill("moai-project-template-optimizer",
       report_format="summary")

# Sample summary:
# Template Optimization Summary:
# ✓ Backup analysis completed
# ✓ 12 template updates applied
# ✓ 8 user customizations preserved
# ✓ 0 conflicts detected
# ✓ Optimization successful
```

## Troubleshooting Examples

### Manual Backup Creation
```python
# If automatic backup fails
Skill("moai-project-template-optimizer",
       mode="create_backup",
       backup_name="manual-backup-2025-11-05")

# Manual backup steps:
# 1. Create .moai-backups/ directory
# 2. Copy current configuration files
# 3. Add backup metadata
# 4. Verify backup integrity
```

### Configuration Repair
```python
# If configuration is corrupted
Skill("moai-project-template-optimizer",
       mode="repair_config")

# Repair operations:
# 1. Validate JSON structure
# 2. Fix missing required fields
# 3. Restore default values where needed
# 4. Update version tracking
```

### Recovery from Failed Optimization
```python
# If optimization fails mid-process
Skill("moai-project-template-optimizer",
       mode="recover",
       recovery_point="last_known_good")

# Recovery steps:
# 1. Identify failure point
# 2. Restore from recovery point
# 3. Complete remaining operations
# 4. Update status accordingly
```
