---
name: moai-change-logger
version: 4.0.0
created: 2025-11-05
updated: '2025-11-18'
status: stable
tier: specialization
description: Comprehensive change tracking and audit logging system that monitors
  file modifications, code changes, and project evolution. Use when tracking project
  history, maintaining audit trails, analyzing development patterns, or when detailed
  change documentation is required for compliance and team collaboration.. Enhanced
  with Context7 MCP for up-to-date documentation.
allowed-tools: Read, Glob, Bash, Write, WebSearch, WebFetch, mcp__context7__resolve-library-id,
  mcp__context7__get-library-docs
primary-agent: alfred
secondary-agents: []
keywords:
- change
- logger
- auth
- test
- git
tags: []
orchestration: null
can_resume: true
typical_chain_position: middle
depends_on: []
stability: stable
---


# moai-change-logger

**Change Logger**

> **Primary Agent**: alfred  
> **Secondary Agents**: none  
> **Version**: 4.0.0  
> **Keywords**: change, logger, auth, test, git

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (Core Concepts)

What It Does

Advanced change tracking and audit logging system that monitors all file modifications, code changes, and project evolution. Provides comprehensive audit trails, change analytics, and project insights for team collaboration and compliance requirements.

**Core capabilities**:
- ✅ Real-time file change monitoring and logging
- ✅ Git integration for commit and branch tracking
- ✅ Change significance assessment and categorization
- ✅ Developer activity tracking and productivity metrics
- ✅ Project evolution analysis and trend identification
- ✅ Audit trail generation for compliance requirements
- ✅ Change impact analysis and dependency tracking
- ✅ Automated change reports and summaries

---

---

### Level 2: Practical Implementation (Common Patterns)

When to Use

- ✅ When tracking project history and evolution
- ✅ During code reviews and change analysis
- ✅ For compliance audit and documentation requirements
- ✅ When analyzing team productivity and patterns
- ✅ Before major releases or deployments
- ✅ When investigating issues or regressions
- ✅ For project retrospectives and improvement planning

---

---

Change Monitoring Systems

### 1. File-Level Monitoring
```python
def monitor_file_changes():
    """Track individual file modifications"""
    changes = {
        "timestamp": datetime.now().isoformat(),
        "files_modified": [],
        "files_created": [],
        "files_deleted": [],
        "significance_score": 0.0,
        "categories": []
    }

    # Monitor file system changes
    for file_path in get_modified_files():
        change = {
            "path": file_path,
            "type": "modified",
            "size_delta": calculate_size_change(file_path),
            "lines_added": count_added_lines(file_path),
            "lines_removed": count_removed_lines(file_path),
            "significance": assess_file_significance(file_path)
        }
        changes["files_modified"].append(change)

    return changes
```

### 2. Git Integration
```python
def track_git_activity():
    """Monitor Git repository activity"""
    git_activity = {
        "current_branch": get_current_branch(),
        "last_commit": get_last_commit_info(),
        "uncommitted_changes": get_git_status(),
        "branch_history": get_recent_commits(limit=10),
        "merge_activity": track_merge_activity(),
        "tag_activity": track_tag_releases()
    }

    return git_activity
```

### 3. Developer Activity Tracking
```python
def track_developer_activity():
    """Monitor developer productivity patterns"""
    activity = {
        "session_duration": calculate_session_duration(),
        "files_touched": count_files_touched(),
        "lines_written": count_lines_written(),
        "tools_used": track_tool_usage(),
        "commands_executed": track_command_history(),
        "productivity_score": calculate_productivity_score()
    }

    return activity
```

---

---

Change Analytics

### 1. Development Patterns
```python
def analyze_development_patterns():
    """Analyze team development patterns"""
    patterns = {
        "peak_development_hours": find_peak_activity_hours(),
        "most_active_days": find_most_active_days(),
        "change_frequency": calculate_change_frequency(),
        "batch_size_distribution": analyze_commit_sizes(),
        "file_type_distribution": analyze_file_type_changes(),
        "collaboration_patterns": analyze_collaboration_patterns()
    }

    return patterns
```

### 2. Productivity Metrics
```python
def calculate_productivity_metrics():
    """Calculate team and individual productivity metrics"""
    metrics = {
        "code_velocity": calculate_code_velocity(),
        "change_success_rate": calculate_change_success_rate(),
        "rework_percentage": calculate_rework_percentage(),
        "test_to_code_ratio": calculate_test_coverage_ratio(),
        "documentation_to_code_ratio": calculate_documentation_ratio(),
        "bug_fix_time": calculate_bug_fix_time()
    }

    return metrics
```

### 3. Quality Indicators
```python
def assess_code_quality_trends():
    """Assess code quality trends over time"""
    quality_trends = {
        "code_complexity": track_complexity_changes(),
        "technical_debt": track_technical_debt(),
        "test_coverage": track_test_coverage_changes(),
        "code_duplication": track_duplication_changes(),
        "security_vulnerabilities": track_security_issues(),
        "performance_trends": track_performance_changes()
    }

    return quality_trends
```

---

---

Integration Examples

### Example 1: Real-Time Change Monitoring
```python
def monitor_project_changes():
    """Real-time monitoring of project changes"""
    while monitoring_active:
        changes = detect_recent_changes()

        if changes:
            Skill("moai-change-logger")

            for change in changes:
                significance, factors = assess_change_significance(change.path, change)
                category = determine_change_category(change)

                log_change({
                    "timestamp": change.timestamp,
                    "file": change.path,
                    "type": change.type,
                    "significance": significance,
                    "category": category,
                    "factors": factors
                })

                # Notify for significant changes
                if significance > 0.7:
                    notify_significant_change(change, category)

        sleep(60)  # Check every minute
```

### Example 2: Pre-Commit Change Analysis
```python
def analyze_commit_changes():
    """Analyze changes before commit"""
    staged_changes = get_staged_changes()

    Skill("moai-change-logger")

    analysis = {
        "total_files": len(staged_changes),
        "categories": categorize_changes(staged_changes),
        "impact": analyze_change_impact(staged_changes),
        "quality_checks": run_quality_checks(staged_changes),
        "recommendations": generate_commit_recommendations(staged_changes)
    }

    display_commit_analysis(analysis)

    return analysis
```

### Example 3: Release Change Summary
```python
def generate_release_summary():
    """Generate comprehensive release change summary"""
    Skill("moai-change-logger")

    # Get changes since last release
    changes = get_changes_since_last_release()

    summary = {
        "release_version": get_next_version(),
        "changes_count": len(changes),
        "categories": categorize_changes(changes),
        "significance_distribution": analyze_significance_distribution(changes),
        "contributors": identify_contributors(changes),
        "risk_assessment": assess_release_risks(changes),
        "testing_coverage": assess_testing_coverage(changes),
        "documentation_updates": check_documentation_updates(changes)
    }

    generate_release_notes(summary)
    save_change_archive(summary)

    return summary
```

---

---

Storage and Retrieval

### 1. Change Storage Format
```python
def store_change_record(change_record):
    """Store change record in structured format"""
    change_data = {
        "id": generate_change_id(),
        "timestamp": change_record.timestamp,
        "file_path": change_record.file_path,
        "change_type": change_record.type,
        "significance": change_record.significance,
        "category": change_record.category,
        "author": change_record.author,
        "commit_hash": change_record.commit_hash,
        "metadata": {
            "lines_added": change_record.lines_added,
            "lines_removed": change_record.lines_removed,
            "file_size_delta": change_record.size_delta,
            "tags": change_record.tags
        }
    }

    # Store in change log database
    save_to_change_log(change_data)

    # Update indexes
    update_search_indexes(change_data)
    update_category_indexes(change_data)
```

### 2. Change Retrieval
```python
def query_changes(query_params):
    """Query changes based on various criteria"""
    filters = []

    if query_params.get("date_range"):
        filters.append(date_range_filter(query_params["date_range"]))

    if query_params.get("categories"):
        filters.append(category_filter(query_params["categories"]))

    if query_params.get("significance_threshold"):
        filters.append(significance_filter(query_params["significance_threshold"]))

    if query_params.get("file_pattern"):
        filters.append(file_pattern_filter(query_params["file_pattern"]))

    if query_params.get("author"):
        filters.append(author_filter(query_params["author"]))

    results = apply_filters(filters)
    return sort_results(results, query_params.get("sort_by", "timestamp"))
```

---

---

Usage Examples

### Example 1: Track Specific File Changes
```python
# User wants to track changes to authentication system
Skill("moai-change-logger")

auth_changes = query_changes({
    "file_pattern": "*auth*",
    "date_range": "last_7_days",
    "categories": ["feature_development", "security", "bug_fixes"]
})

# Display authentication system evolution
display_change_timeline(auth_changes)
```

### Example 2: Compliance Audit
```python
# Generate compliance audit for last month
Skill("moai-change-logger")

compliance_report = generate_compliance_report()
audit_trail = generate_audit_log("monthly")

# Check for compliance issues
issues = identify_compliance_issues(compliance_report, audit_trail)

if issues:
    display_compliance_alerts(issues)
else:
    display_compliance_status("compliant")
```

### Example 3: Project Health Analysis
```python
# Analyze project health based on recent changes
Skill("moai-change-logger")

recent_changes = get_changes_for_period("last_30_days")
health_analysis = {
    "code_quality": assess_code_quality_trends(recent_changes),
    "development_velocity": calculate_velocity_trends(recent_changes),
    "risk_indicators": identify_risk_patterns(recent_changes),
    "team_productivity": assess_productivity_patterns(recent_changes)
}

display_project_health_dashboard(health_analysis)
```

---

**End of Skill** | Comprehensive change tracking for audit, analysis, and project management

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
- Working with [change]
- Need latest documentation
- Verifying technical details

**Example Usage:**

```python
# Fetch latest documentation
from moai_adk.integrations import Context7Helper

helper = Context7Helper()
docs = await helper.get_docs(
    library_id="/org/library",
    topic="change",
    tokens=5000
)
```

**Relevant Libraries:**

| Library | Context7 ID | Use Case |
|---------|-------------|----------|
| [Library 1] | `/org/lib1` | [When to use] |


---

## 📊 Decision Tree

**When to use moai-change-logger:**

```
Start
  ├─ Need change?
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

** .0** (2025-11-12)
- ✨ Context7 MCP integration
- ✨ Progressive Disclosure structure
- ✨ 10+ code examples
- ✨ Primary/secondary agents defined
- ✨ Best practices checklist
- ✨ Decision tree
- ✨ Official references



---

**Generated with**: MoAI-ADK Skill Factory    
**Last Updated**: 2025-11-12  
**Maintained by**: Primary Agent (alfred)
