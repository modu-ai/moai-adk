---
name: moai-project-template-optimizer
description: Handle template comparison and optimization workflows including backup detection, smart merging, and update completion reporting
version: 1.0.0
freedom: medium
type: project
tags: [project, template, optimization, backup, merge, comparison]
tier: alfred
---

# MoAI Project Template Optimizer

This skill manages intelligent template comparison and optimization workflows, detecting existing backups, comparing template versions, and performing smart merges that preserve user customizations while maintaining the latest template structure.

## Core Responsibility

Handle comprehensive template optimization workflows including:
- Backup discovery and analysis in `.moai-backups/` directory
- Template comparison between current files and backup versions
- Smart merge procedures that preserve user customizations
- Template default detection and user content extraction
- Version management and HISTORY section updates
- Configuration updates with optimization flags
- Rollback capability and restore operations

## Workflow Phases

### Phase 1: Backup Discovery & Analysis

**Purpose**: Automatically discover and analyze existing project backups
**Trigger**: Skill invocation or moai-adk version update detection

**Process**:
1. **Backup Directory Scan**
   ```bash
   # Check for .moai-backups/ directory
   if [ -d ".moai-backups/" ]; then
       # List available backups with timestamps
       ls -la .moai-backups/ | grep "backup-"
   fi
   ```

2. **Backup Content Analysis**
   - Read backup metadata and configuration files
   - Identify template files vs user customizations
   - Analyze backup creation timestamps and versions
   - Detect backup completeness and integrity

3. **Backup Selection Strategy**
   - Most recent backup (default)
   - User-specified backup version
   - Backup with specific features or configurations

### Phase 2: Template Comparison Engine

**Purpose**: Detect differences between current templates and backup versions
**Method**: Intelligent file comparison with template-aware parsing

**Comparison Matrix**:
```
File Type          | Comparison Strategy     | Merge Priority
------------------ | ----------------------- | --------------
.claude/settings.json | Structured JSON diff   | Template defaults
.moai/config.json   | User config preservation | User customizations
CLAUDE.md          | Section-based diff      | User content
Hook files         | Code-aware diff         | Template updates
Skills             | Content analysis        | Template defaults
```

**Difference Detection**:
1. **Template Default Detection**: Identify original template content
2. **User Customization Extraction**: Isolate user modifications
3. **Version Compatibility Check**: Ensure merge compatibility
4. **Conflict Identification**: Flag unresolvable differences

### Phase 3: Smart Merge Algorithm

**Purpose**: Preserve user customizations while updating template structure
**Strategy**: Three-way merge with intelligence

**Merge Process**:
1. **Extraction Phase**
   ```python
   # Extract user customizations from backup
   user_customizations = extract_user_content(backup_files)
   
   # Get latest template defaults
   template_defaults = get_current_templates()
   
   # Analyze current project state
   current_state = analyze_current_project()
   ```

2. **Merge Strategy**
   - **Template defaults**: Always use latest template version
   - **User configurations**: Preserve from backup with compatibility updates
   - **Custom content**: Extract and merge with new structure
   - **Deprecated features**: Flag and offer migration options

3. **Conflict Resolution**
   - Template conflicts: Prefer latest template
   - User config conflicts: Preserve user choice with warnings
   - Structural conflicts: Create merge report with resolution options

### Phase 4: Template Default Detection

**Purpose**: Identify and preserve only non-template content during updates
**Method**: Pattern recognition and content analysis

**Detection Patterns**:
```
Template Indicators          | User Content Indicators
---------------------------- | -------------------------
"{{PROJECT_NAME}}"          | Specific project names
"{{CONVERSATION_LANGUAGE}}" | Actual language codes
"src/moai_adk/templates/"  | Custom file paths
Template placeholders       | Real configuration values
Default examples           | Custom implementations
```

**Content Classification**:
1. **Template Content**: Default placeholders, examples, standard structure
2. **User Content**: Custom configurations, specific implementations, personal settings
3. **Mixed Content**: Template files with user modifications (requires careful extraction)

### Phase 5: Version Management & History

**Purpose**: Maintain version tracking and update history sections
**Output**: Comprehensive update reports and history tracking

**Version Tracking**:
```json
{
  "template_version": "0.17.0",
  "previous_version": "0.16.0",
  "update_timestamp": "2025-11-05T12:00:00Z",
  "backup_used": "backup-2025-10-15-v0.16.0",
  "optimization_applied": true
}
```

**HISTORY Section Updates**:
- Document template version changes
- Track optimization procedures applied
- Record user customizations preserved
- Note any manual interventions required

### Phase 6: Configuration Updates

**Purpose**: Update project configuration with optimization flags and metadata
**Target**: `.moai/config.json` and related configuration files

**Optimization Configuration**:
```json
{
  "template_optimization": {
    "last_optimized": "2025-11-05T12:00:00Z",
    "backup_version": "backup-2025-10-15-v0.16.0",
    "template_version": "0.17.0",
    "customizations_preserved": ["language", "team_settings", "domains"],
    "optimization_flags": {
      "merge_applied": true,
      "conflicts_resolved": 0,
      "user_content_extracted": true
    }
  }
}
```

## Usage Patterns

### Standard Template Optimization
```python
# Complete optimization workflow
Skill("moai-project-template-optimizer")

# Executes: Backup Discovery → Template Comparison → Smart Merge → Configuration Update
```

### Backup Analysis Only
```python
# Analyze existing backups without making changes
Skill("moai-project-template-optimizer", mode="analyze_only")
```

### Specific Backup Restoration
```python
# Restore from specific backup
Skill("moai-project-template-optimizer", 
       mode="restore", 
       backup="backup-2025-10-15-v0.16.0")
```

### Rollback Operation
```python
# Rollback to previous state
Skill("moai-project-template-optimizer", mode="rollback")
```

## Error Handling & Validation

### Backup Validation
- Backup directory existence and accessibility
- Backup file integrity and completeness
- Version compatibility checks
- Required file presence validation

### Merge Validation
- Template version compatibility matrix
- User configuration validation
- File permission and accessibility checks
- Dependency verification for updated templates

### Rollback Validation
- Rollback point availability
- State restoration verification
- Configuration consistency checks
- Error recovery mechanisms

## Integration Points

### With Alfred Commands
- `/alfred:0-project`: Post-initialization template optimization
- `/alfred:3-sync`: Template synchronization after updates
- moai-adk update workflows: Automatic optimization triggers

### With Other Skills
- `moai-alfred-autofixes`: Uses auto-fix safety protocols
- `moai-project-language-initializer`: Preserves language configurations
- `moai-alfred-ask-user-questions`: User confirmation workflows

### With Package Templates
- Template source: `src/moai_adk/templates/`
- Synchronization: Local project paths
- Version tracking: Template version management

## Best Practices

### For Template Updates
- Always create backup before optimization
- Verify user customizations preservation
- Test merge results before finalizing
- Document all changes for audit trail

### For Backup Management
- Regular backup creation schedules
- Automated cleanup of old backups
- Backup integrity verification
- Version labeling consistency

### For User Experience
- Clear progress reporting during optimization
- Detailed before/after analysis reports
- User confirmation for major changes
- Rollback options always available

## Performance Optimizations

### Efficient Comparison
- Hash-based file comparison for speed
- Incremental backup analysis
- Parallel file processing for large projects
- Caching of template default patterns

### Smart Merge Optimization
- Conflict prediction and prevention
- User customization pattern caching
- Template version compatibility matrix
- Merge operation batching

## Reporting & Analytics

### Optimization Reports
```
Template Optimization Report
Generated: 2025-11-05T12:00:00Z

Backup Analysis:
- Backups found: 3
- Selected backup: backup-2025-10-15-v0.16.0
- Backup integrity: 100%

Template Comparison:
- Files compared: 47
- Differences found: 12
- Conflicts detected: 0

Smart Merge Results:
- User customizations preserved: 8
- Template updates applied: 12
- Files modified: 15

Configuration Updates:
- Optimization flags set: true
- Version tracking updated: true
- History section updated: true

Performance:
- Total time: 2.3 seconds
- Success rate: 100%
```

### Analytics Tracking
- Optimization frequency and success rates
- Common customization patterns
- Template update adoption metrics
- User satisfaction indicators

## Implementation Notes

This skill consolidates complex template optimization workflows into a focused, intelligent system that:

- **Reduces complexity**: From 700+ lines to optimized skill implementation
- **Improves reliability**: Automated backup and merge procedures
- **Enhances user experience**: Smart customization preservation
- **Maintains traceability**: Comprehensive version tracking and reporting
- **Provides flexibility**: Multiple operation modes and rollback options

The skill serves as the foundation for template lifecycle management and can be extended with additional optimization patterns as template structures evolve.
