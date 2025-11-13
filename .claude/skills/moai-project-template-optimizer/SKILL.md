---
name: "moai-project-template-optimizer"
version: "4.0.0"
created: 2025-11-12
updated: 2025-11-12
status: stable
tier: specialization
description: "Handle template comparison and optimization workflows including backup detection, smart merging, and update completion reporting. Enhanced with Context7 MCP for up-to-date documentation."
allowed-tools: "Read, Glob, Grep, WebSearch, WebFetch, mcp__context7__resolve-library-id, mcp__context7__get-library-docs"
primary-agent: "alfred"
secondary-agents: []
keywords: [project, template, optimizer, test]
tags: [project, template, optimization, backup, merge, comparison]
orchestration: 
can_resume: true
typical_chain_position: "middle"
depends_on: []
---

# moai-project-template-optimizer

**Project Template Optimizer**

> **Primary Agent**: alfred  
> **Secondary Agents**: none  
> **Version**: 4.0.0  
> **Keywords**: project, template, optimizer, test

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (Core Concepts)

**Purpose**: Handle template comparison and optimization workflows including backup detection, smart merging, and update completion reporting. Enhanced with Context7 MCP for up-to-date documentation.

**When to Use:**
- ✅ [Use case 1]
- ✅ [Use case 2]
- ✅ [Use case 3]

**Quick Start Pattern:**

```python
# Basic example
# TODO: Add practical example
```


---

### Level 2: Practical Implementation (Common Patterns)

Workflow Phases

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
**Target**: `.moai/config/config.json` and related configuration files

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

---

Usage Patterns

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

---

Performance Optimizations

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

---

Reporting & Analytics

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

---

Implementation Notes

This skill consolidates complex template optimization workflows into a focused, intelligent system that:

- **Reduces complexity**: From 700+ lines to optimized skill implementation
- **Improves reliability**: Automated backup and merge procedures
- **Enhances user experience**: Smart customization preservation
- **Maintains traceability**: Comprehensive version tracking and reporting
- **Provides flexibility**: Multiple operation modes and rollback options

The skill serves as the foundation for template lifecycle management and can be extended with additional optimization patterns as template structures evolve.

---

### Level 3: Advanced Patterns (Expert Reference)

> **Note**: Advanced patterns for complex scenarios.

**Coming soon**: Deep dive into expert-level usage.


---

## 🎯 Best Practices Checklist

**Must-Have:**
- ✅ [Critical practice 1]
- ✅ [Critical practice 2]

**Recommended:**
- ✅ [Recommended practice 1]
- ✅ [Recommended practice 2]

**Security:**
- 🔒 [Security practice 1]


---

## 🔗 Context7 MCP Integration

**When to Use Context7 for This Skill:**

This skill benefits from Context7 when:
- Working with [project]
- Need latest documentation
- Verifying technical details

**Example Usage:**

```python
# Fetch latest documentation
from moai_adk.integrations import Context7Helper

helper = Context7Helper()
docs = await helper.get_docs(
    library_id="/org/library",
    topic="project",
    tokens=5000
)
```

**Relevant Libraries:**

| Library | Context7 ID | Use Case |
|---------|-------------|----------|
| [Library 1] | `/org/lib1` | [When to use] |


---

## 📊 Decision Tree

**When to use moai-project-template-optimizer:**

```
Start
  ├─ Need project?
  │   ├─ YES → Use this skill
  │   └─ NO → Consider alternatives
  └─ Complex scenario?
      ├─ YES → See Level 3
      └─ NO → Start with Level 1
```


---

## 🔄 Integration with Other Skills

**Prerequisite Skills:**
- Skill("prerequisite-1") – [Why needed]

**Complementary Skills:**
- Skill("complementary-1") – [How they work together]

**Next Steps:**
- Skill("next-step-1") – [When to use after this]


---

## 📚 Official References

**Primary Documentation:**
- [Official Docs](https://...) – Complete reference

**Best Practices:**
- [Best Practices Guide](https://...) – Official recommendations


---

## 📈 Version History

**v4.0.0** (2025-11-12)
- ✨ Context7 MCP integration
- ✨ Progressive Disclosure structure
- ✨ 10+ code examples
- ✨ Primary/secondary agents defined
- ✨ Best practices checklist
- ✨ Decision tree
- ✨ Official references



---

**Generated with**: MoAI-ADK Skill Factory v4.0  
**Last Updated**: 2025-11-12  
**Maintained by**: Primary Agent (alfred)
