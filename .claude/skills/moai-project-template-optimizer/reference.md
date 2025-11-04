# Template Optimizer Reference

## Skill API Reference

### Core Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| mode | string | No | "optimize" | Operation mode: optimize, analyze_only, restore, rollback |
| backup | string | No | latest | Specific backup name to use for restore operations |
| target_version | string | No | latest | Template version to optimize towards |
| merge_strategy | string | No | "smart_merge" | Merge strategy: preserve_user, template_priority, smart_merge |
| conflict_resolution | string | No | "template_priority" | Conflict resolution: template_priority, preserve_user, manual_merge |
| performance_mode | string | No | "standard" | Performance mode: fast, standard, thorough |
| report_format | string | No | "summary" | Report format: summary, detailed, json |
| compatibility_check | boolean | No | true | Enable version compatibility checking |
| changed_files_only | boolean | No | false | Process only changed files for incremental updates |

### Operation Modes

#### optimize
Default mode for complete template optimization workflow.

**Process**:
1. Backup Discovery & Analysis
2. Template Comparison
3. Smart Merge
4. Configuration Updates
5. Version Management

**Parameters**:
- `merge_strategy`: How to handle differences between versions
- `conflict_resolution`: How to resolve merge conflicts
- `target_version`: Specific template version to target

**Returns**: Optimization report with success/failure status

#### analyze_only
Analyze backups and template differences without making changes.

**Process**:
1. Backup Discovery
2. Content Analysis
3. Comparison Report Generation
4. Recommendations

**Parameters**:
- `report_format`: Detail level of analysis report
- `performance_mode`: Analysis speed vs thoroughness

**Returns**: Detailed analysis report with recommendations

#### restore
Restore from a specific backup while preserving template structure.

**Process**:
1. Backup Validation
2. User Customization Extraction
3. Template Structure Update
4. Merge and Apply

**Parameters**:
- `backup`: Required - specific backup name
- `merge_strategy`: How to merge backup with current templates

**Returns**: Restoration report with preserved customizations

#### rollback
Rollback to a previous optimization state.

**Process**:
1. Recovery Point Identification
2. State Restoration
3. Configuration Consistency Check
4. Rollback Verification

**Parameters**:
- `recovery_point`: Specific point to rollback to
- `verify_only`: Check rollback possibility without executing

**Returns**: Rollback report with verification status

## File Structure Reference

### Backup Directory Structure
```
.moai-backups/
├── backup-YYYY-MM-DD-v{version}/
│   ├── .claude/
│   │   ├── settings.json
│   │   ├── skills/
│   │   └── hooks/
│   ├── .moai/
│   │   ├── config.json
│   │   └── scripts/
│   ├── CLAUDE.md
│   ├── backup-metadata.json
│   └── integrity-check.json
└── backup-YYYY-MM-DD-v{version}/
    └── ...
```

### Configuration Structure

#### .moai/config.json Template Optimization Section
```json
{
  "template_optimization": {
    "last_optimized": "2025-11-05T12:00:00Z",
    "backup_version": "backup-2025-10-15-v0.16.0",
    "template_version": "0.17.0",
    "previous_version": "0.16.0",
    "customizations_preserved": [
      "language.settings",
      "github.workflow",
      "report_generation.preferences"
    ],
    "optimization_flags": {
      "merge_applied": true,
      "conflicts_resolved": 0,
      "user_content_extracted": true,
      "template_updates_applied": 12,
      "files_modified": 15
    },
    "recovery_points": [
      {
        "timestamp": "2025-11-05T12:00:00Z",
        "operation": "optimize",
        "backup_created": "backup-2025-11-05-pre-optimize"
      }
    ]
  }
}
```

#### Backup Metadata Structure
```json
{
  "backup_metadata": {
    "created_at": "2025-10-15T10:30:00Z",
    "version": "0.16.0",
    "moai_adk_version": "0.16.0",
    "project_name": "my-project",
    "user_customizations": {
      "language": "ko",
      "team_mode": true,
      "domains": ["frontend", "backend"]
    },
    "file_integrity": {
      "total_files": 47,
      "checksum_verified": true,
      "corrupted_files": 0
    }
  }
}
```

## Template Comparison Matrix

### File Type Handling

| File Type | Comparison Strategy | Merge Priority | User Content Detection |
|-----------|---------------------|----------------|------------------------|
| `.claude/settings.json` | Structured JSON diff | Template defaults | User permission settings |
| `.moai/config.json` | User config preservation | User customizations | All user settings |
| `CLAUDE.md` | Section-based diff | User content | Custom sections |
| Hook files | Code-aware diff | Template updates | Custom implementations |
| Skills | Content analysis | Template defaults | Custom skill content |
| Templates | Pattern matching | Template defaults | Custom templates |

### Difference Detection Patterns

#### Template Indicators
- `{{PROJECT_NAME}}` - Project name placeholder
- `{{CONVERSATION_LANGUAGE}}` - Language placeholder
- `{{USER_NICKNAME}}` - User nickname placeholder
- `src/moai_adk/templates/` - Template source path
- Default example content
- Standard structure comments

#### User Content Indicators
- Specific project names (not placeholders)
- Actual language codes (en, ko, ja, zh)
- Real configuration values
- Custom file paths
- Personal implementations
- Custom sections in documentation

## Error Reference

### Backup Errors

#### Error: Backup Directory Not Found
```
Message: "No .moai-backups/ directory found"
Cause: Project has never created backups
Solution: Create initial backup before optimization
Recovery: Proceed with template-only optimization
```

#### Error: Backup Integrity Check Failed
```
Message: "Backup {backup_name} integrity check failed"
Cause: Backup files are corrupted or incomplete
Solution: Use different backup or create new backup
Recovery: Restore from known good backup
```

#### Error: Version Incompatibility
```
Message: "Backup version {backup_version} incompatible with current templates"
Cause: Template structure has breaking changes
Solution: Use migration path or manual merge
Recovery: Stepwise version migration
```

### Merge Errors

#### Error: Unresolvable Conflicts
```
Message: "Found {count} unresolvable merge conflicts"
Cause: Template and user changes conflict fundamentally
Solution: Manual conflict resolution required
Recovery: Preserve user version, apply template updates manually
```

#### Error: Template Structure Validation Failed
```
Message: "Updated template structure validation failed"
Cause: Merge resulted in invalid template structure
Solution: Restore from backup and retry with different strategy
Recovery: Use template_priority merge strategy
```

### Configuration Errors

#### Error: Configuration Update Failed
```
Message: "Failed to update .moai/config.json"
Cause: JSON structure invalid or permissions issue
Solution: Check file permissions and JSON syntax
Recovery: Manual configuration update
```

#### Error: Rollback Point Not Found
```
Message: "No rollback point found for {timestamp}"
Cause: Recovery point corrupted or missing
Solution: Use available rollback point or full restore
Recovery: Full backup restoration
```

## Performance Reference

### Comparison Algorithms

#### Hash-Based Comparison
- **Use Case**: Large projects with many files
- **Speed**: Very fast
- **Accuracy**: High
- **Memory Usage**: Low
- **Implementation**: SHA-256 file hashing

#### Line-by-Line Comparison
- **Use Case**: Small configuration files
- **Speed**: Medium
- **Accuracy**: Very high
- **Memory Usage**: Medium
- **Implementation**: Unified diff algorithm

#### Structure-Aware Comparison
- **Use Case**: JSON/YAML configuration files
- **Speed**: Medium
- **Accuracy**: Very high
- **Memory Usage**: High
- **Implementation**: JSON diff with semantic analysis

### Merge Optimization Strategies

#### Parallel Processing
- **Files**: Process multiple files simultaneously
- **Threads**: Auto-detected based on CPU cores
- **Memory**: Proportional to file count
- **Speed Improvement**: 2-4x faster

#### Incremental Updates
- **Scope**: Only changed files since last optimization
- **Comparison**: Use cached file hashes
- **Memory**: Very low
- **Speed Improvement**: 5-10x faster

#### Batch Operations
- **Merge**: Group similar operations together
- **I/O**: Reduce file system calls
- **Memory**: Medium
- **Speed Improvement**: 1.5-2x faster

## Integration Reference

### Alfred Command Integration

#### /alfred:0-project Integration
```python
# Post-initialization template optimization
if project_initialized:
    Skill("moai-project-template-optimizer")
```

#### /alfred:3-sync Integration
```python
# Template synchronization after updates
if template_changes_detected:
    Skill("moai-project-template-optimizer", mode="sync_update")
```

### moai-adk Update Integration
```python
# Automatic optimization after package updates
if moai_adk_updated:
    Skill("moai-project-template-optimizer", 
           mode="post_update",
           preserve_all_settings=True)
```

### Skill Integration Points

#### moai-project-language-initializer
```python
# Preserve language configurations during optimization
language_settings = Skill("moai-project-language-initializer", mode="get_settings")
Skill("moai-project-template-optimizer", preserve_language=True)
```

#### moai-alfred-autofixes
```python
# Use auto-fix safety protocols for template fixes
Skill("moai-project-template-optimizer", 
       auto_fix_enabled=True,
       safety_checks=True)
```

## Troubleshooting Reference

### Common Issues and Solutions

#### Slow Performance
**Symptoms**: Optimization takes >30 seconds
**Causes**: Large project, slow I/O, insufficient memory
**Solutions**:
- Use `performance_mode="fast"`
- Enable `changed_files_only=True`
- Close other applications

#### Memory Issues
**Symptoms**: Out of memory errors during optimization
**Causes**: Very large projects, parallel processing
**Solutions**:
- Use `performance_mode="standard"`
- Disable parallel processing
- Process files in batches

#### Permission Errors
**Symptoms**: Cannot read/write configuration files
**Causes**: File permissions, readonly filesystem
**Solutions**:
- Check file permissions with `ls -la`
- Use `sudo` if necessary (not recommended)
- Copy project to writable location

#### Backup Corruption
**Symptoms**: Backup integrity check fails
**Causes**: Incomplete backup, disk errors
**Solutions**:
- Create new backup before optimization
- Use different backup if available
- Check disk health

### Debug Mode
```python
# Enable detailed logging for troubleshooting
Skill("moai-project-template-optimizer",
       debug_mode=True,
       log_level="debug",
       save_logs=True)
```

Debug information includes:
- Detailed step-by-step progress
- File-by-file comparison results
- Merge decision reasoning
- Performance metrics
- Error stack traces
