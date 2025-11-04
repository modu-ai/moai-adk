# Project Template Optimizer

Handle template comparison and optimization workflows including backup detection, smart merging, and update completion reporting.

## Purpose

Manages intelligent template comparison and optimization workflows, detecting existing backups, comparing template versions, and performing smart merges that preserve user customizations while maintaining the latest template structure.

## Core Workflow

1. **Backup Discovery** - Find and analyze existing project backups
2. **Template Comparison** - Compare current files with backup and package templates  
3. **Smart Merge** - Preserve user customizations while applying template updates
4. **Update Completion** - Report changes and maintain synchronization

## When to Use

- After moai-adk version updates
- When template conflicts are detected
- During project optimization workflows
- When synchronizing local and package templates

## Quick Start

```bash
# Run template optimization
Skill("moai-project-template-optimizer")

# Alfred will:
# 1. Scan for backups in .moai-backups/
# 2. Compare template versions
# 3. Propose smart merge strategy
# 4. Execute with user approval
```

## Features

- **Automatic Backup Detection** - Finds existing project backups
- **Intelligent Comparison** - Analyzes differences between versions
- **Smart Merging** - Preserves user customizations
- **Conflict Resolution** - Handles template vs user content conflicts
- **Update Reporting** - Comprehensive change documentation

## Files Structure

- `SKILL.md` - Complete workflow documentation
- `examples.md` - Practical usage examples
- `reference.md` - Technical implementation details
- `README.md` - This overview
