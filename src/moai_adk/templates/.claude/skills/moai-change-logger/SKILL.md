---
name: moai-change-logger
version: 1.0.0
created: 2025-11-05
updated: 2025-11-05
status: active
description: Comprehensive change tracking and audit logging system that monitors file modifications, code changes, and project evolution. Use when tracking project history, maintaining audit trails, analyzing development patterns, or when detailed change documentation is required for compliance and team collaboration.
keywords: [change-tracking, audit-logging, file-monitoring, git-history, project-evolution, compliance]
allowed-tools:
  - Read
  - Glob
  - Bash
  - Write
---

# Comprehensive Change Logger

## Skill Metadata

| Field | Value |
| ----- | ----- |
| Version | 1.0.0 |
| Tier | Alfred (Project Management) |
| Auto-load | During file operations and project changes |
| Purpose | Track and log all project changes for audit and analysis |

---

## What It Does

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

## When to Use

- ✅ When tracking project history and evolution
- ✅ During code reviews and change analysis
- ✅ For compliance audit and documentation requirements
- ✅ When analyzing team productivity and patterns
- ✅ Before major releases or deployments
- ✅ When investigating issues or regressions
- ✅ For project retrospectives and improvement planning

---

## Change Monitoring Systems

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

## Change Categorization

### 1. Significance Assessment
```python
def assess_change_significance(file_path, change_details):
    """Assess how significant a change is"""
    score = 0.0
    factors = []

    # File type significance
    if is_critical_file(file_path):
        score += 0.4
        factors.append("critical_file")
    elif is_config_file(file_path):
        score += 0.3
        factors.append("config_file")
    elif is_documentation_file(file_path):
        score += 0.1
        factors.append("documentation")

    # Change magnitude
    lines_changed = change_details.lines_added + change_details.lines_removed
    if lines_changed > 100:
        score += 0.3
        factors.append("large_change")
    elif lines_changed > 20:
        score += 0.2
        factors.append("medium_change")
    elif lines_changed > 0:
        score += 0.1
        factors.append("small_change")

    # Impact assessment
    if affects_api(file_path):
        score += 0.2
        factors.append("api_impact")
    if affects_database(file_path):
        score += 0.2
        factors.append("database_impact")
    if affects_security(file_path):
        score += 0.3
        factors.append("security_impact")

    return min(score, 1.0), factors
```

### 2. Change Categories
```python
def categorize_changes(changes):
    """Categorize changes by type and purpose"""
    categories = {
        "feature_development": [],
        "bug_fixes": [],
        "refactoring": [],
        "documentation": [],
        "configuration": [],
        "testing": [],
        "performance": [],
        "security": []
    }

    for change in changes:
        category = determine_change_category(change)
        categories[category].append(change)

    return categories

def determine_change_category(change):
    """Determine the category of a specific change"""
    file_path = change.path.lower()

    if "test" in file_path or "spec" in file_path:
        return "testing"
    elif "readme" in file_path or "doc" in file_path:
        return "documentation"
    elif "config" in file_path or "setting" in file_path:
        return "configuration"
    elif "fix" in change.commit_message or "bug" in change.commit_message:
        return "bug_fixes"
    elif "refactor" in change.commit_message or "clean" in change.commit_message:
        return "refactoring"
    elif "feat" in change.commit_message or "add" in change.commit_message:
        return "feature_development"
    elif "perf" in change.commit_message or "optimize" in change.commit_message:
        return "performance"
    elif "security" in change.commit_message or "auth" in file_path:
        return "security"
    else:
        return "feature_development"  # Default category
```

---

## Audit Trail Generation

### 1. Comprehensive Audit Log
```python
def generate_audit_log(time_period="daily"):
    """Generate comprehensive audit trail"""
    audit_log = {
        "period": time_period,
        "generated_at": datetime.now().isoformat(),
        "summary": generate_period_summary(),
        "detailed_changes": get_detailed_changes(time_period),
        "compliance_metrics": calculate_compliance_metrics(),
        "risk_assessment": assess_change_risks(),
        "recommendations": generate_recommendations()
    }

    return audit_log
```

### 2. Compliance Reporting
```python
def generate_compliance_report():
    """Generate compliance-focused report"""
    compliance_data = {
        "change_authorization": verify_change_authorization(),
        "code_review_coverage": calculate_review_coverage(),
        "test_coverage_changes": assess_test_coverage_impact(),
        "security_impact": assess_security_changes(),
        "documentation_completeness": check_documentation_completeness(),
        "approval_workflow": verify_approval_workflow()
    }

    return compliance_data
```

### 3. Change Impact Analysis
```python
def analyze_change_impact(changes):
    """Analyze the impact of changes on the project"""
    impact_analysis = {
        "affected_modules": identify_affected_modules(changes),
        "dependency_impact": assess_dependency_changes(changes),
        "api_changes": identify_api_changes(changes),
        "database_changes": identify_database_changes(changes),
        "test_impact": assess_test_impact(changes),
        "deployment_impact": assess_deployment_readiness(changes),
        "rollback_complexity": assess_rollback_complexity(changes)
    }

    return impact_analysis
```

---

## Change Analytics

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

## Integration Examples

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

## Report Generation

### 1. Daily Change Report
```python
def generate_daily_report():
    """Generate daily change activity report"""
    Skill("moai-change-logger")

    daily_changes = get_changes_for_period("daily")

    report = {
        "date": datetime.now().strftime("%Y-%m-%d"),
        "summary": {
            "total_changes": len(daily_changes),
            "files_modified": count_modified_files(daily_changes),
            "lines_added": sum(c.lines_added for c in daily_changes),
            "lines_removed": sum(c.lines_removed for c in daily_changes)
        },
        "categories": categorize_changes(daily_changes),
        "significant_changes": [c for c in daily_changes if c.significance > 0.5],
        "productivity_metrics": calculate_daily_productivity(daily_changes)
    }

    save_daily_report(report)
    return report
```

### 2. Weekly Project Analysis
```python
def generate_weekly_analysis():
    """Generate comprehensive weekly project analysis"""
    Skill("moai-change-logger")

    weekly_changes = get_changes_for_period("weekly")

    analysis = {
        "week_start": get_week_start(),
        "week_end": get_week_end(),
        "development_velocity": calculate_weekly_velocity(weekly_changes),
        "quality_trends": analyze_weekly_quality_trends(weekly_changes),
        "team_activity": analyze_team_activity(weekly_changes),
        "risk_indicators": identify_weekly_risks(weekly_changes),
        "achievements": highlight_weekly_achievements(weekly_changes),
        "improvement_areas": identify_improvement_opportunities(weekly_changes)
    }

    generate_weekly_dashboard(analysis)
    return analysis
```

---

## Storage and Retrieval

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

## Usage Examples

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