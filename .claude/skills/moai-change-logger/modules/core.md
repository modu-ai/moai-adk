
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


**End of Skill** | Comprehensive change tracking for audit, analysis, and project management


### Level 3: Advanced Patterns (Expert Reference)

> **Note**: Advanced patterns for complex scenarios.

**Coming soon**: Deep dive into expert-level usage.



## 📚 Official References

**Primary Documentation:**
- [Official Docs](https://...) – Complete reference

**Best Practices:**
- [Best Practices Guide](https://...) – Official recommendations



## 📈 Version History

** .0** (2025-11-12)
- ✨ Context7 MCP integration
- ✨ Progressive Disclosure structure
- ✨ 10+ code examples
- ✨ Primary/secondary agents defined
- ✨ Best practices checklist
- ✨ Decision tree
- ✨ Official references




**Generated with**: MoAI-ADK Skill Factory    
**Last Updated**: 2025-11-12  
**Maintained by**: Primary Agent (alfred)


## Implementation Guide

## 🎯 Best Practices Checklist

**Must-Have:**
- ✅ [Critical practice 1]
- ✅ [Critical practice 2]

**Recommended:**
- ✅ [Recommended practice 1]
- ✅ [Recommended practice 2]

**Security:**
- 🔒 [Security practice 1]



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



## 🔄 Integration with Other Skills

**Prerequisite Skills:**
- Skill("prerequisite-1") – [Why needed]

**Complementary Skills:**
- Skill("complementary-1") – [How they work together]

**Next Steps:**
- Skill("next-step-1") – [When to use after this]




## Advanced Patterns






## Context7 Integration

### Related Libraries & Tools
- [Conventional Commits](/conventional-commits): Commit message convention
- [semantic-release](/semantic-release/semantic-release): Automated versioning

### Official Documentation
- [Documentation](https://www.conventionalcommits.org/)
- [API Reference](https://semantic-release.gitbook.io/)

### Version-Specific Guides
Latest stable version: Latest
- [Release Notes](https://github.com/semantic-release/semantic-release/releases)
- [Migration Guide](https://semantic-release.gitbook.io/semantic-release/support/migration-guides)
